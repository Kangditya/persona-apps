package httpx

import (
    "encoding/json"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/gin-gonic/gin"
)

func TestRecoveryMiddlewareSanitizesPanics(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    router.Use(RequestIDMiddleware(logger), RecoveryMiddleware(logger))
    router.GET("/panic", func(*gin.Context) {
        panic("database password must not reach the caller")
    })

    response := httptest.NewRecorder()
    router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))

    if response.Code != http.StatusInternalServerError {
        t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
    }
    if response.Header().Get(RequestIDHeader) == "" {
        t.Fatal("response is missing request ID")
    }
    if body := response.Body.String(); body == "" || strings.Contains(body, "database password") {
        t.Fatalf("unsafe recovery body = %q", body)
    }
    var payload struct {
        Error struct {
            Code      string `json:"code"`
            RequestID string `json:"request_id"`
        } `json:"error"`
    }
    if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
        t.Fatal(err)
    }
    if payload.Error.Code != "internal_error" || payload.Error.RequestID != response.Header().Get(RequestIDHeader) {
        t.Fatalf("recovery payload = %#v", payload)
    }
}
