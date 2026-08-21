package auth

import (
    "bytes"
    "encoding/json"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
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

func TestSessionReturnsExpiryAndSortedPermissionsWithoutCaching(t *testing.T) {
    database, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })

    token := strings.Repeat("s", 43)
    expiry := time.Now().Add(time.Hour).UTC()
    mock.ExpectQuery("SELECT u.id, u.display_name").WithArgs(digest(token)).WillReturnRows(
        sqlmock.NewRows([]string{"id", "display_name", "status", "permission_snapshot", "expires_at"}).
            AddRow("11111111-1111-1111-1111-111111111111", "Operator", "ACTIVE", `["offering.read","event.manage","event.read"]`, expiry),
    )
    mock.ExpectExec("UPDATE operator_sessions SET csrf_token_hash").WithArgs(sqlmock.AnyArg(), digest(token)).WillReturnResult(sqlmock.NewResult(0, 1))

    service := &Service{db: database}
    request := httptest.NewRequest(http.MethodGet, "/api/operations/v1/auth/session", nil)
    request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
    response := httptest.NewRecorder()
    testRouter(service).ServeHTTP(response, request)

    if response.Code != http.StatusOK {
        t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
    }
    if response.Header().Get("Cache-Control") != "no-store" {
        t.Fatalf("Cache-Control = %q", response.Header().Get("Cache-Control"))
    }
    var body struct {
        Data struct {
            Permissions []string  `json:"permissions"`
            ExpiresAt   time.Time `json:"expires_at"`
            CSRFToken   string    `json:"csrf_token"`
        } `json:"data"`
    }
    if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
        t.Fatal(err)
    }
    if strings.Join(body.Data.Permissions, ",") != "event.manage,event.read,offering.read" {
        t.Fatalf("permissions = %#v", body.Data.Permissions)
    }
    if !body.Data.ExpiresAt.Equal(expiry) || len(body.Data.CSRFToken) != 43 {
        t.Fatalf("session data = %#v", body.Data)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func TestRequireEnforcesSessionOriginCSRFAuthorizationOrder(t *testing.T) {
    token := strings.Repeat("s", 43)
    csrf := strings.Repeat("c", 43)
    tests := []struct {
        name        string
        method      string
        permissions string
        origin      string
        csrf        string
        storedCSRF  []byte
        expiry      time.Time
        wantStatus  int
        wantReached bool
    }{
        {name: "expired", method: http.MethodGet, permissions: `["event.read"]`, expiry: time.Now().Add(-time.Minute), wantStatus: http.StatusUnauthorized},
        {name: "wrong permission", method: http.MethodGet, permissions: `["offering.read"]`, expiry: time.Now().Add(time.Hour), wantStatus: http.StatusForbidden},
        {name: "disallowed origin", method: http.MethodPost, permissions: `["event.read"]`, origin: "https://attacker.invalid", csrf: csrf, expiry: time.Now().Add(time.Hour), wantStatus: http.StatusForbidden},
        {name: "missing csrf", method: http.MethodPost, permissions: `["event.read"]`, origin: "https://operations.example.invalid", expiry: time.Now().Add(time.Hour), wantStatus: http.StatusForbidden},
        {name: "mismatched csrf", method: http.MethodPost, permissions: `["event.read"]`, origin: "https://operations.example.invalid", csrf: csrf, storedCSRF: digest("different"), expiry: time.Now().Add(time.Hour), wantStatus: http.StatusForbidden},
        {name: "authorized", method: http.MethodPost, permissions: `["event.read"]`, origin: "https://operations.example.invalid", csrf: csrf, storedCSRF: digest(csrf), expiry: time.Now().Add(time.Hour), wantStatus: http.StatusNoContent, wantReached: true},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            database, mock, err := sqlmock.New()
            if err != nil {
                t.Fatal(err)
            }
            t.Cleanup(func() { _ = database.Close() })
            mock.ExpectQuery("SELECT u.id, u.display_name").WithArgs(digest(token)).WillReturnRows(
                sqlmock.NewRows([]string{"id", "display_name", "status", "permission_snapshot", "expires_at"}).
                    AddRow("11111111-1111-1111-1111-111111111111", "Operator", "ACTIVE", test.permissions, test.expiry),
            )
            if test.storedCSRF != nil {
                mock.ExpectQuery("SELECT csrf_token_hash").WithArgs(digest(token), "11111111-1111-1111-1111-111111111111").WillReturnRows(
                    sqlmock.NewRows([]string{"csrf_token_hash"}).AddRow(test.storedCSRF),
                )
            }
            service := &Service{db: database, config: config.AuthConfig{AllowedOrigins: map[string]struct{}{"https://operations.example.invalid": {}}}}
            reached := false
            router := gin.New()
            router.Handle(test.method, "/protected", service.Require("event.read"), func(c *gin.Context) {
                reached = true
                c.Status(http.StatusNoContent)
            })
            request := httptest.NewRequest(test.method, "/protected", nil)
            request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
            request.Header.Set("Origin", test.origin)
            request.Header.Set("X-CSRF-Token", test.csrf)
            response := httptest.NewRecorder()
            router.ServeHTTP(response, request)

            if response.Code != test.wantStatus || reached != test.wantReached {
                t.Fatalf("status/reached = %d/%t, want %d/%t", response.Code, reached, test.wantStatus, test.wantReached)
            }
            if err := mock.ExpectationsWereMet(); err != nil {
                t.Fatal(err)
            }
        })
    }
}
