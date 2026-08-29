package domain

import (
    "bytes"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "time"
)

const (
    maxCursorLength  = 512
    maxCursorPayload = 256
)

type Cursor struct {
    Version   int       `json:"v"`
    CreatedAt time.Time `json:"t"`
    ID        string    `json:"i"`
}

func NormalizeListInput(input ListInput) (int, *Cursor, error) {
    limit := input.Limit
    if limit == 0 {
        limit = DefaultListLimit
    }
    if limit < 1 || limit > MaxListLimit {
        return 0, nil, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidInput, MaxListLimit)
    }
    if input.Status != "" && !validListStatus(input.Status) {
        return 0, nil, fmt.Errorf("%w: unsupported purchase status filter", ErrInvalidInput)
    }
    if input.Cursor == "" {
        return limit, nil, nil
    }
    cursor, err := DecodeCursor(input.Cursor)
    if err != nil {
        return 0, nil, err
    }
    return limit, &cursor, nil
}

func validListStatus(status Status) bool {
    switch status {
    case StatusPendingPayment, StatusPaid, StatusEligible, StatusCancelled:
        return true
    default:
        return false
    }
}

func EncodeCursor(cursor Cursor) (string, error) {
    payload, err := json.Marshal(cursor)
    if err != nil {
        return "", fmt.Errorf("encode purchase cursor: %w", err)
    }
    return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeCursor(value string) (Cursor, error) {
    if len(value) == 0 || len(value) > maxCursorLength {
        return Cursor{}, ErrInvalidCursor
    }
    payload, err := base64.RawURLEncoding.DecodeString(value)
    if err != nil || len(payload) == 0 || len(payload) > maxCursorPayload {
        return Cursor{}, ErrInvalidCursor
    }
    decoder := json.NewDecoder(bytes.NewReader(payload))
    decoder.DisallowUnknownFields()
    var cursor Cursor
    if err := decoder.Decode(&cursor); err != nil {
        return Cursor{}, ErrInvalidCursor
    }
    if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
        return Cursor{}, ErrInvalidCursor
    }
    if cursor.Version != 1 || cursor.CreatedAt.IsZero() || len(cursor.ID) != 36 {
        return Cursor{}, ErrInvalidCursor
    }
    cursor.CreatedAt = cursor.CreatedAt.UTC()
    return cursor, nil
}
