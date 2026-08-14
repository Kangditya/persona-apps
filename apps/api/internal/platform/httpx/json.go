package httpx

import (
    "bytes"
    "encoding/json"
    "errors"
    "io"
    "mime"
    "net/http"
    "regexp"

    "github.com/google/uuid"
)

var idempotencyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{7,254}$`)

var (
    ErrInvalidJSON          = errors.New("invalid JSON request")
    ErrRequestTooLarge      = errors.New("request body is too large")
    ErrUnsupportedMediaType = errors.New("unsupported media type")
)

func DecodeJSON(w http.ResponseWriter, request *http.Request, destination any, maximumBytes int64) error {
    mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
    if err != nil || mediaType != "application/json" {
        return ErrUnsupportedMediaType
    }

    request.Body = http.MaxBytesReader(w, request.Body, maximumBytes)
    decoder := json.NewDecoder(request.Body)
    var raw json.RawMessage
    if err := decoder.Decode(&raw); err != nil {
        var tooLarge *http.MaxBytesError
        if errors.As(err, &tooLarge) {
            return ErrRequestTooLarge
        }
        return ErrInvalidJSON
    }
    if err := decoder.Decode(&struct{}{}); err != io.EOF || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
        return ErrInvalidJSON
    }

    decoder = json.NewDecoder(bytes.NewReader(raw))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(destination); err != nil {
        return ErrInvalidJSON
    }
    return nil
}

func CanonicalUUID(value string) (string, error) {
    parsed, err := uuid.Parse(value)
    if err != nil || parsed == uuid.Nil || parsed.String() != value {
        return "", errors.New("invalid UUID")
    }
    return value, nil
}

func IdempotencyKey(value string) (string, error) {
    if !idempotencyKeyPattern.MatchString(value) {
        return "", errors.New("invalid idempotency key")
    }
    return value, nil
}
