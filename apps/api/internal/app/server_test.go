package app

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/Kangditya/persona-apps/apps/api/internal/config"
)

type pingFunc func(context.Context) error

func (f pingFunc) PingContext(ctx context.Context) error {
    return f(ctx)
}

func TestHealthAndReadiness(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))

    tests := []struct {
        name        string
        path        string
        ping        pingFunc
        wantCode    int
        wantBody    string
        wantContent string
    }{
        {
            name:        "health does not require the database",
            path:        "/health",
            ping:        func(context.Context) error { return errors.New("not called") },
            wantCode:    http.StatusOK,
            wantBody:    "{\"status\":\"ok\"}\n",
            wantContent: "application/json",
        },
        {
            name:        "ready database reachable",
            path:        "/ready",
            ping:        func(context.Context) error { return nil },
            wantCode:    http.StatusOK,
            wantBody:    "{\"status\":\"ready\"}\n",
            wantContent: "application/json",
        },
        {
            name:        "not ready database unreachable",
            path:        "/ready",
            ping:        func(context.Context) error { return errors.New("database unavailable") },
            wantCode:    http.StatusServiceUnavailable,
            wantBody:    "{\"status\":\"unavailable\"}\n",
            wantContent: "application/json",
        },
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            server := newTestServer(t, test.ping, logger, config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 20})
            request := httptest.NewRequest(http.MethodGet, test.path, nil)
            response := httptest.NewRecorder()

            server.Handler.ServeHTTP(response, request)

            if response.Code != test.wantCode {
                t.Fatalf("status code = %d, want %d", response.Code, test.wantCode)
            }
            if response.Body.String() != test.wantBody {
                t.Fatalf("body = %q, want %q", response.Body.String(), test.wantBody)
            }
            if content := response.Header().Get("Content-Type"); content != test.wantContent {
                t.Fatalf("content type = %q, want %q", content, test.wantContent)
            }
        })
    }
}

func TestHealthRejectsUnsupportedMethod(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    server := newTestServer(t, pingFunc(func(context.Context) error { return nil }), logger, config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 20})
    request := httptest.NewRequest(http.MethodPost, "/health", nil)
    response := httptest.NewRecorder()

    server.Handler.ServeHTTP(response, request)

    if response.Code != http.StatusMethodNotAllowed {
        t.Fatalf("status code = %d, want %d", response.Code, http.StatusMethodNotAllowed)
    }
}

func TestRouterUsesAPIErrorBoundary(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    server := newTestServer(t, pingFunc(func(context.Context) error { return nil }), logger, config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 20})

    tests := []struct {
        name       string
        method     string
        path       string
        wantStatus int
        wantCode   string
    }{
        {name: "non API route", method: http.MethodGet, path: "/missing", wantStatus: http.StatusNotFound},
        {name: "unknown API route", method: http.MethodGet, path: "/api/public/v1/missing", wantStatus: http.StatusNotFound, wantCode: "not_found"},
        {name: "unavailable public dependency", method: http.MethodGet, path: "/api/public/v1/events/active", wantStatus: http.StatusServiceUnavailable, wantCode: "service_unavailable"},
        {name: "unsupported public method", method: http.MethodPost, path: "/api/public/v1/events/active", wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed"},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            request := httptest.NewRequest(test.method, test.path, nil)
            response := httptest.NewRecorder()

            server.Handler.ServeHTTP(response, request)

            if response.Code != test.wantStatus {
                t.Fatalf("status code = %d, want %d", response.Code, test.wantStatus)
            }
            if test.wantCode == "" {
                return
            }
            if response.Header().Get("X-Request-ID") == "" {
                t.Fatal("API response is missing request ID")
            }
            if response.Header().Get("Cache-Control") != "no-store" && test.wantCode != "not_found" && test.wantCode != "method_not_allowed" {
                t.Fatalf("public response cache control = %q", response.Header().Get("Cache-Control"))
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
            if payload.Error.Code != test.wantCode || payload.Error.RequestID != response.Header().Get("X-Request-ID") {
                t.Fatalf("API error = %#v", payload.Error)
            }
        })
    }
}

func TestPublicRateLimitUsesSafeClientIP(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    database := pingFunc(func(context.Context) error { return nil })

    t.Run("untrusted forwarded address is ignored", func(t *testing.T) {
        server := newTestServer(t, database, logger, config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 1})
        first := publicRequest("192.0.2.1:1234", "203.0.113.1")
        second := publicRequest("192.0.2.1:1234", "203.0.113.2")
        firstResponse := httptest.NewRecorder()
        secondResponse := httptest.NewRecorder()
        server.Handler.ServeHTTP(firstResponse, first)
        server.Handler.ServeHTTP(secondResponse, second)
        if firstResponse.Code != http.StatusServiceUnavailable || secondResponse.Code != http.StatusTooManyRequests {
            t.Fatalf("statuses = %d, %d", firstResponse.Code, secondResponse.Code)
        }
        if secondResponse.Header().Get("Retry-After") != "1" {
            t.Fatalf("retry after = %q", secondResponse.Header().Get("Retry-After"))
        }
    })

    t.Run("configured proxy may supply forwarded client IP", func(t *testing.T) {
        server := newTestServer(t, database, logger, config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 1, TrustedProxyCIDRs: []string{"127.0.0.1/32"}})
        first := publicRequest("127.0.0.1:1234", "203.0.113.1")
        second := publicRequest("127.0.0.1:1234", "203.0.113.2")
        firstResponse := httptest.NewRecorder()
        secondResponse := httptest.NewRecorder()
        server.Handler.ServeHTTP(firstResponse, first)
        server.Handler.ServeHTTP(secondResponse, second)
        if firstResponse.Code != http.StatusServiceUnavailable || secondResponse.Code != http.StatusServiceUnavailable {
            t.Fatalf("statuses = %d, %d", firstResponse.Code, secondResponse.Code)
        }
    })
}

func TestServerComposesPublicEventRepository(t *testing.T) {
    database, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })
    now := time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC)
    mock.ExpectQuery(regexp.QuoteMeta("SELECT id, event_year, name, status, registration_opens_at, registration_closes_at, participant_quota, version, created_at, updated_at FROM qurban_events WHERE status = $1 ORDER BY event_year DESC, id ASC LIMIT 1")).
        WithArgs("ACTIVE").
        WillReturnRows(sqlmock.NewRows([]string{"id", "event_year", "name", "status", "registration_opens_at", "registration_closes_at", "participant_quota", "version", "created_at", "updated_at"}).
            AddRow("11111111-1111-1111-1111-111111111111", 2026, "Qurban 2026", "ACTIVE", now, now.Add(time.Hour), 100, 1, now, now))

    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    server := newTestServer(t, database, logger, config.PublicConfig{RateLimitPerMinute: 60, RateLimitBurst: 20})
    response := httptest.NewRecorder()
    server.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/public/v1/events/active", nil))

    if response.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
    }
    if response.Header().Get("Cache-Control") != "no-store" {
        t.Fatalf("cache control = %q", response.Header().Get("Cache-Control"))
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func newTestServer(t *testing.T, database readinessChecker, logger *slog.Logger, public config.PublicConfig) *http.Server {
    t.Helper()
    server, err := NewServer(":0", database, logger, public, nil)
    if err != nil {
        t.Fatal(err)
    }
    return server
}

func publicRequest(remoteAddress, forwardedFor string) *http.Request {
    request := httptest.NewRequest(http.MethodGet, "/api/public/v1/events/active", nil)
    request.RemoteAddr = remoteAddress
    request.Header.Set("X-Forwarded-For", forwardedFor)
    return request
}
