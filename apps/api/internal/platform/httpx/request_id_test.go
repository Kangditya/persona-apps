package httpx

import (
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.Use(RequestIDMiddleware(slog.New(slog.NewTextHandler(io.Discard, nil))))
    router.GET("/", func(c *gin.Context) {
        if RequestID(c.Request.Context()) == "" {
            t.Fatal("request ID missing from context")
        }
        c.Status(http.StatusNoContent)
    })

    t.Run("accepts bounded ID", func(t *testing.T) {
        request := httptest.NewRequest(http.MethodGet, "/", nil)
        request.Header.Set(RequestIDHeader, "request_1")
        response := httptest.NewRecorder()
        router.ServeHTTP(response, request)
        if got := response.Header().Get(RequestIDHeader); got != "request_1" {
            t.Fatalf("header = %q", got)
        }
    })
    t.Run("rejects untrusted ID without reflection", func(t *testing.T) {
        request := httptest.NewRequest(http.MethodGet, "/", nil)
        request.Header.Set(RequestIDHeader, "bad value")
        response := httptest.NewRecorder()
        router.ServeHTTP(response, request)
        if response.Code != http.StatusBadRequest {
            t.Fatalf("status = %d", response.Code)
        }
        if got := response.Header().Get(RequestIDHeader); got == "" || got == "bad value" {
            t.Fatalf("unsafe request ID = %q", got)
        }
    })
}
