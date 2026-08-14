package httpx

import (
    "errors"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestDecodeJSONEnforcesTheCommandBoundary(t *testing.T) {
    type body struct {
        Name string `json:"name"`
    }
    tests := []struct {
        name        string
        contentType string
        input       string
        limit       int64
        want        error
    }{
        {name: "valid", contentType: "application/json; charset=utf-8", input: `{"name":"Qurban"}`, limit: 64},
        {name: "unsupported media", contentType: "text/plain", input: `{}`, limit: 64, want: ErrUnsupportedMediaType},
        {name: "unknown field", contentType: "application/json", input: `{"name":"Qurban","extra":true}`, limit: 64, want: ErrInvalidJSON},
        {name: "trailing value", contentType: "application/json", input: `{"name":"Qurban"}{}`, limit: 64, want: ErrInvalidJSON},
        {name: "null", contentType: "application/json", input: `null`, limit: 64, want: ErrInvalidJSON},
        {name: "too large", contentType: "application/json", input: `{"name":"Qurban"}`, limit: 4, want: ErrRequestTooLarge},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            request := httptest.NewRequest("POST", "/", strings.NewReader(test.input))
            request.Header.Set("Content-Type", test.contentType)
            var decoded body
            err := DecodeJSON(httptest.NewRecorder(), request, &decoded, test.limit)
            if !errors.Is(err, test.want) {
                t.Fatalf("DecodeJSON() error = %v, want %v", err, test.want)
            }
        })
    }
}

func TestCanonicalUUIDRejectsNonCanonicalValues(t *testing.T) {
    const value = "abcdefab-cdef-4abc-8def-abcdefabcdef"
    if got, err := CanonicalUUID(value); err != nil || got != value {
        t.Fatalf("CanonicalUUID() = %q, %v", got, err)
    }
    if _, err := CanonicalUUID(strings.ToUpper(value)); err == nil {
        t.Fatal("CanonicalUUID() accepted a non-canonical UUID")
    }
}

func TestIdempotencyKeyUsesTheContractBounds(t *testing.T) {
    if _, err := IdempotencyKey("intent-1"); err != nil {
        t.Fatal(err)
    }
    if _, err := IdempotencyKey("short"); err == nil {
        t.Fatal("IdempotencyKey() accepted a short key")
    }
}
