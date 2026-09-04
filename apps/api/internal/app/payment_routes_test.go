package app

import (
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    "github.com/Kangditya/persona-apps/apps/api/internal/payment"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
)

func TestPaymentRoutesAreCompositionAndCorsGated(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    public := config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 20, StorefrontAllowedOrigins: map[string]struct{}{"https://store.example.test": {}}}
    disabled, err := NewServer(":0", nil, logger, public, nil, nil, nil, nil, nil, nil, nil)
    if err != nil {
        t.Fatal(err)
    }
    missing := httptest.NewRecorder()
    disabled.Handler.ServeHTTP(missing, httptest.NewRequest(http.MethodPost, "/api/public/v1/purchases/11111111-1111-1111-1111-111111111111/payment-evidence", nil))
    if missing.Code != http.StatusNotFound {
        t.Fatalf("disabled = %d", missing.Code)
    }
    missingOperations := httptest.NewRecorder()
    disabled.Handler.ServeHTTP(missingOperations, httptest.NewRequest(http.MethodGet, "/api/operations/v1/payments/11111111-1111-1111-1111-111111111111/evidence", nil))
    if missingOperations.Code != http.StatusNotFound {
        t.Fatalf("disabled operations = %d", missingOperations.Code)
    }
    cipher, err := idempotency.NewCipher([]idempotency.Key{{ID: "test", Value: make([]byte, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    enabled, err := NewServer(":0", nil, logger, public, nil, payment.NewPublicHandler(nil, cipher, logger, nil), nil, nil, nil, nil, nil)
    if err != nil {
        t.Fatal(err)
    }
    request := httptest.NewRequest(http.MethodOptions, "/api/public/v1/purchases/11111111-1111-1111-1111-111111111111/payment-evidence", nil)
    request.Header.Set("Origin", "https://store.example.test")
    request.Header.Set("Access-Control-Request-Method", http.MethodPost)
    request.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type, Idempotency-Key, X-Request-ID")
    response := httptest.NewRecorder()
    enabled.Handler.ServeHTTP(response, request)
    allow := response.Header().Get("Access-Control-Allow-Headers")
    if response.Code != http.StatusNoContent || response.Header().Get("Access-Control-Allow-Credentials") != "" || allow != "X-Request-ID, Content-Type, Idempotency-Key, Authorization" {
        t.Fatalf("cors = %d %#v", response.Code, response.Header())
    }
}
