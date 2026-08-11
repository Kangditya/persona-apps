package auth

import (
    "bytes"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

func testRouter(service *Service) *gin.Engine {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.Use(httpx.RequestIDMiddleware(slog.New(slog.NewTextHandler(io.Discard, nil))))
    service.RegisterRoutes(router.Group("/api/operations/v1"))
    return router
}

func TestPureAuthenticationGuards(t *testing.T) {
    if !validReturnTo("/operations") || validReturnTo("//attacker.invalid") || validReturnTo("https://attacker.invalid") {
        t.Fatal("return path validation is unsafe")
    }
    permissions, err := claimedPermissions([]byte("[\"event.read\",\"unknown\"]"))
    if err != nil || len(permissions) != 1 || permissions[0] != "event.read" {
        t.Fatalf("claimedPermissions() = %v, %v", permissions, err)
    }
    service := Service{config: config.AuthConfig{CookieEncryptionKey: bytes.Repeat([]byte{2}, 32)}}
    sealed, err := service.sealLoginState(loginState{State: "state"})
    if err != nil || sealed == "" {
        t.Fatalf("sealLoginState() = %q, %v", sealed, err)
    }
}

func TestRequireAbortsUnauthenticatedRequests(t *testing.T) {
    service := &Service{}
    router := testRouter(service)
    reached := false
    router.GET("/api/operations/v1/protected", service.Require("event.read"), func(*gin.Context) {
        reached = true
    })

    request := httptest.NewRequest(http.MethodGet, "/api/operations/v1/protected", nil)
    response := httptest.NewRecorder()
    router.ServeHTTP(response, request)

    if response.Code != http.StatusUnauthorized {
        t.Fatalf("status code = %d, want %d", response.Code, http.StatusUnauthorized)
    }
    if reached {
        t.Fatal("protected handler ran after authentication failure")
    }
}
