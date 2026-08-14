package idempotency

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

var (
    ErrConflict          = errors.New("idempotency key was reused with a different request")
    ErrUnavailableReplay = errors.New("idempotency replay cannot be decrypted")
)

type Key struct {
    ID    string
    Value []byte
}

type Cipher struct {
    primary Key
    keys    map[string][]byte
}

type envelope struct {
    KeyID      string
    Nonce      string
    Ciphertext string
}

func NewCipher(keys []Key) (Cipher, error) {
    if len(keys) == 0 {
        return Cipher{}, errors.New("at least one idempotency response key is required")
    }
    result := Cipher{primary: keys[0], keys: make(map[string][]byte, len(keys))}
    for _, key := range keys {
        if key.ID == "" || len(key.Value) != 32 {
            return Cipher{}, errors.New("idempotency response keys require an ID and 32 bytes")
        }
        if _, exists := result.keys[key.ID]; exists {
            return Cipher{}, fmt.Errorf("duplicate idempotency response key ID %q", key.ID)
        }
        result.keys[key.ID] = key.Value
    }
    return result, nil
}

func (c Cipher) Seal(aad, body []byte) (json.RawMessage, error) {
    block, err := aes.NewCipher(c.primary.Value)
    if err != nil {
        return nil, fmt.Errorf("create replay cipher: %w", err)
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("create replay GCM: %w", err)
    }
    nonce := make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, fmt.Errorf("generate replay nonce: %w", err)
    }
    value, err := json.Marshal(envelope{KeyID: c.primary.ID, Nonce: base64.RawStdEncoding.EncodeToString(nonce), Ciphertext: base64.RawStdEncoding.EncodeToString(gcm.Seal(nil, nonce, body, aad))})
    if err != nil {
        return nil, fmt.Errorf("marshal replay envelope: %w", err)
    }
    return value, nil
}

func (c Cipher) Open(aad, stored []byte) (json.RawMessage, error) {
    var value envelope
    if err := json.Unmarshal(stored, &value); err != nil {
        return nil, ErrUnavailableReplay
    }
    key, found := c.keys[value.KeyID]
    if !found {
        return nil, ErrUnavailableReplay
    }
    nonce, err := base64.RawStdEncoding.DecodeString(value.Nonce)
    if err != nil {
        return nil, ErrUnavailableReplay
    }
    ciphertext, err := base64.RawStdEncoding.DecodeString(value.Ciphertext)
    if err != nil {
        return nil, ErrUnavailableReplay
    }
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, ErrUnavailableReplay
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil || len(nonce) != gcm.NonceSize() {
        return nil, ErrUnavailableReplay
    }
    body, err := gcm.Open(nil, nonce, ciphertext, aad)
    if err != nil || !json.Valid(body) {
        return nil, ErrUnavailableReplay
    }
    return body, nil
}

type Command struct {
    Namespace   string
    Key         string
    RequestHash string
    Retention   time.Duration
}

type Response struct {
    Status int
    Body   json.RawMessage
}

func Namespace(surface, command, callerScope string) (string, error) {
    if surface == "" || command == "" || callerScope == "" {
        return "", errors.New("idempotency namespace requires surface, command, and caller scope")
    }
    return strings.Join([]string{surface, command, callerScope}, "."), nil
}

func RequestHash(value any) (string, error) {
    body, err := json.Marshal(value)
    if err != nil {
        return "", fmt.Errorf("marshal idempotency request: %w", err)
    }
    sum := sha256.Sum256(body)
    return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

func Execute(ctx context.Context, db *sql.DB, crypt Cipher, command Command, work func(*sql.Tx) (Response, error)) (response Response, replayed bool, err error) {
    if command.Namespace == "" || command.Key == "" || command.RequestHash == "" || command.Retention <= 0 {
        return Response{}, false, errors.New("idempotency command is incomplete")
    }
    err = database.Within(ctx, db, func(tx *sql.Tx) error {
        if _, deleteErr := tx.ExecContext(ctx, "DELETE FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2 AND expires_at <= now()", command.Namespace, command.Key); deleteErr != nil {
            return fmt.Errorf("remove expired idempotency record: %w", deleteErr)
        }
        var claimed string
        claimErr := tx.QueryRowContext(ctx, "INSERT INTO idempotency_records (namespace, idempotency_key, request_hash, expires_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING namespace", command.Namespace, command.Key, command.RequestHash, time.Now().Add(command.Retention)).Scan(&claimed)
        switch {
        case claimErr == nil:
            result, workErr := work(tx)
            if workErr != nil {
                return workErr
            }
            if result.Status < 200 || result.Status >= 300 || !json.Valid(result.Body) {
                return errors.New("only successful JSON responses are replayable")
            }
            sealed, sealErr := crypt.Seal(aad(command, result.Status), result.Body)
            if sealErr != nil {
                return sealErr
            }
            if _, updateErr := tx.ExecContext(ctx, "UPDATE idempotency_records SET response_status = $1, response_body = $2::jsonb WHERE namespace = $3 AND idempotency_key = $4", result.Status, string(sealed), command.Namespace, command.Key); updateErr != nil {
                return fmt.Errorf("store idempotency response: %w", updateErr)
            }
            response = result
            return nil
        case !errors.Is(claimErr, sql.ErrNoRows):
            return fmt.Errorf("claim idempotency key: %w", claimErr)
        }

        var hash string
        var status sql.NullInt64
        var stored []byte
        if readErr := tx.QueryRowContext(ctx, "SELECT request_hash, response_status, response_body FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2", command.Namespace, command.Key).Scan(&hash, &status, &stored); readErr != nil {
            return fmt.Errorf("read idempotency replay: %w", readErr)
        }
        if hash != command.RequestHash {
            return ErrConflict
        }
        if !status.Valid || len(stored) == 0 {
            return errors.New("idempotency replay is incomplete")
        }
        body, openErr := crypt.Open(aad(command, int(status.Int64)), stored)
        if openErr != nil {
            return openErr
        }
        response, replayed = Response{Status: int(status.Int64), Body: body}, true
        return nil
    })
    return response, replayed, err
}

func aad(command Command, status int) []byte {
    return []byte(strings.Join([]string{command.Namespace, command.Key, command.RequestHash, fmt.Sprintf("%d", status)}, "\x00"))
}
