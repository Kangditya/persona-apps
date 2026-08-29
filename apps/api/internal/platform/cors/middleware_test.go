package cors

import (
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

func TestMiddlewareAllowsOnlyExactConfiguredPreflights(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.Use(httpx.RequestIDMiddleware(slog.New(slog.NewTextHandler(io.Discard, nil))))
    group := router.Group("/api/operations/v1")
    group.Use(Middleware(Policy{
        Origins:     map[string]struct{}{"https://operations.example.invalid": {}},
        Methods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch},
        Headers:     []string{"Content-Type", "X-CSRF-Token", "Idempotency-Key", "X-Request-ID"},
        Credentials: true,
    }))
    group.OPTIONS("/*path", func(c *gin.Context) { c.Status(http.StatusNoContent) })

    tests := []struct {
        name       string
        origin     string
        method     string
        headers    string
        wantStatus int
        wantOrigin string
    }{
        {name: "allowed", origin: "https://operations.example.invalid", method: "POST", headers: "content-type, x-csrf-token", wantStatus: 204, wantOrigin: "https://operations.example.invalid"},
        {name: "arbitrary origin", origin: "https://attacker.invalid", method: "POST", wantStatus: 403},
        {name: "unsupported method", origin: "https://operations.example.invalid", method: "DELETE", wantStatus: 403, wantOrigin: "https://operations.example.invalid"},
        {name: "unsupported header", origin: "https://operations.example.invalid", method: "POST", headers: "authorization", wantStatus: 403, wantOrigin: "https://operations.example.invalid"},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            request := httptest.NewRequest(http.MethodOptions, "/api/operations/v1/events", nil)
            request.Header.Set("Origin", test.origin)
            request.Header.Set("Access-Control-Request-Method", test.method)
            request.Header.Set("Access-Control-Request-Headers", test.headers)
            response := httptest.NewRecorder()
            router.ServeHTTP(response, request)
            if response.Code != test.wantStatus || response.Header().Get("Access-Control-Allow-Origin") != test.wantOrigin {
                t.Fatalf("preflight status/origin = %d/%q", response.Code, response.Header().Get("Access-Control-Allow-Origin"))
            }
            if test.wantStatus == http.StatusNoContent && response.Header().Get("Access-Control-Allow-Credentials") != "true" {
                t.Fatal("credentialed preflight omitted credentials")
            }
        })
    }
}
