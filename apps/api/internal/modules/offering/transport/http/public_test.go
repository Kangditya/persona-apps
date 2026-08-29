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

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

type publicCatalogueReaderStub struct {
    list   func(context.Context, string, ListInput) (CatalogueResult, error)
    detail func(context.Context, string) (CatalogueOffering, error)
}

func (stub publicCatalogueReaderStub) ListPublic(ctx context.Context, eventID string, input ListInput) (CatalogueResult, error) {
    return stub.list(ctx, eventID, input)
}

func (stub publicCatalogueReaderStub) GetPublic(ctx context.Context, offeringID string) (CatalogueOffering, error) {
    return stub.detail(ctx, offeringID)
}

func TestPublicOfferingListIsBoundedAndMinimized(t *testing.T) {
    cursor, err := encodeCursor(offeringCursor{Version: 1, Code: "share-a", ID: "22222222-2222-2222-2222-222222222222"})
    if err != nil {
        t.Fatal(err)
    }
    available := int64(7)
    reader := publicCatalogueReaderStub{
        list: func(_ context.Context, eventID string, input ListInput) (CatalogueResult, error) {
            if eventID != "11111111-1111-1111-1111-111111111111" || input.Limit != 1 || input.Cursor != cursor {
                t.Fatalf("list input = %q %#v", eventID, input)
            }
            return CatalogueResult{Offerings: []CatalogueOffering{{
                Offering: Offering{
                    ID:                  "22222222-2222-2222-2222-222222222222",
                    EventID:             eventID,
                    Code:                "share-a",
                    Name:                "One share",
                    Kind:                "SHARE",
                    PriceMinor:          3_500_000,
                    CurrencyCode:        "IDR",
                    ParticipantCapacity: 7,
                    Status:              StatusPublished,
                    Version:             8,
                },
                AvailableParticipantUnits: &available,
            }}, Limit: 1, NextCursor: "next"}, nil
        },
        detail: func(context.Context, string) (CatalogueOffering, error) {
            return CatalogueOffering{}, errors.New("detail must not run")
        },
    }
    response := servePublicOffering(publicOfferingRouter(reader), "/api/public/v1/events/11111111-1111-1111-1111-111111111111/offerings?limit=1&cursor="+cursor)
    if response.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
    }
    var payload struct {
        Data []map[string]any `json:"data"`
        Page struct {
            Limit      int    `json:"limit"`
            NextCursor string `json:"next_cursor"`
        } `json:"page"`
    }
    if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
        t.Fatal(err)
    }
    if len(payload.Data) != 1 || payload.Page.Limit != 1 || payload.Page.NextCursor != "next" {
        t.Fatalf("public page = %#v", payload)
    }
    if payload.Data[0]["available_participant_units"] != float64(7) {
        t.Fatalf("availability = %#v", payload.Data[0]["available_participant_units"])
    }
    for _, internalField := range []string{"participant_quota", "version", "published_at", "created_at", "updated_at"} {
        if _, found := payload.Data[0][internalField]; found {
            t.Fatalf("public response exposes %q: %#v", internalField, payload.Data[0])
        }
    }
}

func TestPublicOfferingValidationAndErrorMapping(t *testing.T) {
    hiddenID := "22222222-2222-2222-2222-222222222222"
    reader := publicCatalogueReaderStub{
        list: func(context.Context, string, ListInput) (CatalogueResult, error) {
            return CatalogueResult{}, ErrNotFound
        },
        detail: func(_ context.Context, offeringID string) (CatalogueOffering, error) {
            switch offeringID {
            case hiddenID:
                return CatalogueOffering{}, ErrNotFound
            case "33333333-3333-3333-3333-333333333333":
                return CatalogueOffering{}, httpx.ErrDependencyUnavailable
            default:
                return CatalogueOffering{}, errors.New("database password must not reach the caller")
            }
        },
    }
    router := publicOfferingRouter(reader)
    tests := []struct {
        name       string
        path       string
        wantStatus int
        wantCode   string
    }{
        {name: "invalid noncanonical UUID", path: "/api/public/v1/offerings/22222222222222222222222222222222", wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
        {name: "invalid query limit", path: "/api/public/v1/events/11111111-1111-1111-1111-111111111111/offerings?limit=0", wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
        {name: "malformed cursor", path: "/api/public/v1/events/11111111-1111-1111-1111-111111111111/offerings?cursor=not-a-cursor", wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
        {name: "unknown query is rejected", path: "/api/public/v1/events/11111111-1111-1111-1111-111111111111/offerings?sort=price", wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
        {name: "hidden detail is not found", path: "/api/public/v1/offerings/" + hiddenID, wantStatus: http.StatusNotFound, wantCode: "not_found"},
        {name: "dependency failure", path: "/api/public/v1/offerings/33333333-3333-3333-3333-333333333333", wantStatus: http.StatusServiceUnavailable, wantCode: "service_unavailable"},
        {name: "unexpected failure is sanitized", path: "/api/public/v1/offerings/44444444-4444-4444-4444-444444444444", wantStatus: http.StatusInternalServerError, wantCode: "internal_error"},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            response := servePublicOffering(router, test.path)
            if response.Code != test.wantStatus {
                t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
            }
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
            if strings.Contains(response.Body.String(), "database password") {
                t.Fatalf("unsafe error body = %q", response.Body.String())
            }
        })
    }
}

func TestPublicOfferingRendersNullForUnboundedAvailability(t *testing.T) {
    reader := publicCatalogueReaderStub{
        list: func(context.Context, string, ListInput) (CatalogueResult, error) {
            return CatalogueResult{Offerings: []CatalogueOffering{{
                Offering: Offering{
                    ID:                  "22222222-2222-2222-2222-222222222222",
                    EventID:             "11111111-1111-1111-1111-111111111111",
                    Code:                "share-a",
                    Name:                "One share",
                    Kind:                "SHARE",
                    PriceMinor:          3_500_000,
                    CurrencyCode:        "IDR",
                    ParticipantCapacity: 7,
                    Status:              StatusPublished,
                },
                AvailableParticipantUnits: nil,
            }}, Limit: DefaultListLimit}, nil
        },
        detail: func(context.Context, string) (CatalogueOffering, error) {
            return CatalogueOffering{}, errors.New("detail must not run")
        },
    }
    response := servePublicOffering(publicOfferingRouter(reader), "/api/public/v1/events/11111111-1111-1111-1111-111111111111/offerings")
    if response.Code != http.StatusOK {
        t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
    }
    var payload struct {
        Data []map[string]any `json:"data"`
    }
    if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
        t.Fatal(err)
    }
    if len(payload.Data) != 1 {
        t.Fatalf("data = %#v", payload.Data)
    }
    value, found := payload.Data[0]["available_participant_units"]
    if !found || value != nil {
        t.Fatalf("availability = %#v, found %t", value, found)
    }
}

func publicOfferingRouter(reader PublicCatalogueReader) *gin.Engine {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))
    router.Use(httpx.RequestIDMiddleware(logger))
    RegisterPublicRoutes(router.Group("/api/public/v1"), reader, logger)
    return router
}

func servePublicOffering(router *gin.Engine, path string) *httptest.ResponseRecorder {
    response := httptest.NewRecorder()
    router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
    return response
}
