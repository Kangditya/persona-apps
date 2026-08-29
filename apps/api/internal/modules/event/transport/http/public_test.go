package http

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

type activeReaderFunc func(context.Context) (*Event, error)

func (function activeReaderFunc) Active(ctx context.Context) (*Event, error) {
    return function(ctx)
}

func TestPublicActiveEvent(t *testing.T) {
    opens := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
    closes := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
    quota := int64(500)
    router := publicEventRouter(activeReaderFunc(func(context.Context) (*Event, error) {
        return &Event{
            ID:                   "11111111-1111-1111-1111-111111111111",
            EventYear:            2026,
            Name:                 "Qurban 2026",
            Status:               StatusActive,
            RegistrationOpensAt:  &opens,
            RegistrationClosesAt: &closes,
            ParticipantQuota:     &quota,
            Version:              3,
            CreatedAt:            opens,
            UpdatedAt:            closes,
        }, nil
    }))

    response := servePublicEvent(router, "/api/public/v1/events/active")
    if response.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
    }
    if response.Header().Get(httpx.RequestIDHeader) == "" {
        t.Fatal("response is missing request ID")
    }
    var payload struct {
        Data map[string]any `json:"data"`
    }
    if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
        t.Fatal(err)
    }
    if payload.Data["status"] != string(StatusActive) || payload.Data["name"] != "Qurban 2026" {
        t.Fatalf("public data = %#v", payload.Data)
    }
    for _, internalField := range []string{"participant_quota", "version", "created_at", "updated_at"} {
        if _, found := payload.Data[internalField]; found {
            t.Fatalf("public response exposes %q: %#v", internalField, payload.Data)
        }
    }
}

func TestPublicActiveEventReturnsNullOrSafeErrors(t *testing.T) {
    tests := []struct {
        name       string
        reader     ActiveReader
        wantStatus int
        wantCode   string
        wantBody   string
    }{
        {
            name: "no active event",
            reader: activeReaderFunc(func(context.Context) (*Event, error) {
                return nil, nil
            }),
            wantStatus: http.StatusOK,
            wantBody:   `{"data":null}`,
        },
        {
            name: "dependency unavailable",
            reader: activeReaderFunc(func(context.Context) (*Event, error) {
                return nil, httpx.ErrDependencyUnavailable
            }),
            wantStatus: http.StatusServiceUnavailable,
            wantCode:   "service_unavailable",
        },
        {
            name: "unexpected error is sanitized",
            reader: activeReaderFunc(func(context.Context) (*Event, error) {
                return nil, errors.New("database topology must not reach the caller")
            }),
            wantStatus: http.StatusInternalServerError,
            wantCode:   "internal_error",
        },
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            response := servePublicEvent(publicEventRouter(test.reader), "/api/public/v1/events/active")
            if response.Code != test.wantStatus {
                t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
            }
            if test.wantBody != "" && strings.TrimSpace(response.Body.String()) != test.wantBody {
                t.Fatalf("body = %q, want %q", response.Body.String(), test.wantBody)
            }
            if test.wantCode != "" {
                var payload struct {
                    Error struct {
                        Code string `json:"code"`
                    } `json:"error"`
                }
                if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
                    t.Fatal(err)
                }
                if payload.Error.Code != test.wantCode {
                    t.Fatalf("error code = %q, want %q", payload.Error.Code, test.wantCode)
                }
            }
            if strings.Contains(response.Body.String(), "database topology") {
                t.Fatalf("unsafe error body = %q", response.Body.String())
            }
        })
    }
}

func publicEventRouter(reader ActiveReader) *gin.Engine {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    router.Use(httpx.RequestIDMiddleware(logger))
    RegisterPublicRoutes(router.Group("/api/public/v1"), reader, logger)
    return router
}

func servePublicEvent(router *gin.Engine, path string) *httptest.ResponseRecorder {
    response := httptest.NewRecorder()
    router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
    return response
}
