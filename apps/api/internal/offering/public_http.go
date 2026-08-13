package offering

import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "strconv"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type PublicCatalogueReader interface {
    ListPublic(context.Context, string, ListInput) (CatalogueResult, error)
    GetPublic(context.Context, string) (CatalogueOffering, error)
}

type publicOffering struct {
    ID                        string  `json:"id"`
    EventID                   string  `json:"event_id"`
    Code                      string  `json:"code"`
    Name                      string  `json:"name"`
    Kind                      string  `json:"offering_kind"`
    Description               *string `json:"description,omitempty"`
    PriceMinor                int64   `json:"price_minor"`
    CurrencyCode              string  `json:"currency_code"`
    ParticipantCapacity       int32   `json:"participant_capacity"`
    AvailableParticipantUnits *int64  `json:"available_participant_units"`
    Status                    Status  `json:"status"`
}

type publicOfferingPage struct {
    Limit      int    `json:"limit"`
    NextCursor string `json:"next_cursor,omitempty"`
}

type publicOfferingResponse struct {
    Data publicOffering `json:"data"`
}

type publicOfferingListResponse struct {
    Data []publicOffering   `json:"data"`
    Page publicOfferingPage `json:"page"`
}

func RegisterPublicRoutes(group *gin.RouterGroup, reader PublicCatalogueReader, logger *slog.Logger) {
    group.GET("/events/:event_id/offerings", func(c *gin.Context) {
        eventID, ok := publicUUID(c, "event_id")
        if !ok {
            return
        }
        input, err := publicListInput(c)
        if err != nil {
            writeInvalidRequest(c, "invalid list parameters")
            return
        }
        if reader == nil {
            writePublicOfferingError(c, logger, httpx.ErrDependencyUnavailable)
            return
        }
        result, err := reader.ListPublic(c.Request.Context(), eventID, input)
        if err != nil {
            writePublicOfferingError(c, logger, err)
            return
        }
        data := make([]publicOffering, 0, len(result.Offerings))
        for _, value := range result.Offerings {
            if value.Offering.Status != StatusPublished {
                writePublicOfferingError(c, logger, fmt.Errorf("unexpected public offering status %q", value.Offering.Status))
                return
            }
            data = append(data, toPublicOffering(value))
        }
        c.JSON(http.StatusOK, publicOfferingListResponse{
            Data: data,
            Page: publicOfferingPage{Limit: result.Limit, NextCursor: result.NextCursor},
        })
    })

    group.GET("/offerings/:offering_id", func(c *gin.Context) {
        offeringID, ok := publicUUID(c, "offering_id")
        if !ok {
            return
        }
        if reader == nil {
            writePublicOfferingError(c, logger, httpx.ErrDependencyUnavailable)
            return
        }
        result, err := reader.GetPublic(c.Request.Context(), offeringID)
        if err != nil {
            writePublicOfferingError(c, logger, err)
            return
        }
        if result.Offering.Status != StatusPublished {
            writePublicOfferingError(c, logger, fmt.Errorf("unexpected public offering status %q", result.Offering.Status))
            return
        }
        c.JSON(http.StatusOK, publicOfferingResponse{Data: toPublicOffering(result)})
    })
}

func publicUUID(c *gin.Context, parameter string) (string, bool) {
    raw := c.Param(parameter)
    parsed, err := uuid.Parse(raw)
    if err != nil || parsed == uuid.Nil || parsed.String() != raw {
        writeInvalidRequest(c, "invalid "+parameter)
        return "", false
    }
    return raw, true
}

func publicListInput(c *gin.Context) (ListInput, error) {
    values := c.Request.URL.Query()
    for name := range values {
        if name != "cursor" && name != "limit" {
            return ListInput{}, ErrInvalidInput
        }
    }
    input := ListInput{}
    if cursors, found := values["cursor"]; found {
        if len(cursors) != 1 || cursors[0] == "" {
            return ListInput{}, ErrInvalidCursor
        }
        input.Cursor = cursors[0]
    }
    if limits, found := values["limit"]; found {
        if len(limits) != 1 || limits[0] == "" {
            return ListInput{}, ErrInvalidInput
        }
        limit, err := strconv.Atoi(limits[0])
        if err != nil || limit < 1 || limit > MaxListLimit {
            return ListInput{}, ErrInvalidInput
        }
        input.Limit = limit
    }
    if _, _, err := normalizeListInput(input); err != nil {
        return ListInput{}, err
    }
    return input, nil
}

func toPublicOffering(value CatalogueOffering) publicOffering {
    offering := value.Offering
    return publicOffering{
        ID:                        offering.ID,
        EventID:                   offering.EventID,
        Code:                      offering.Code,
        Name:                      offering.Name,
        Kind:                      offering.Kind,
        Description:               copyString(offering.Description),
        PriceMinor:                offering.PriceMinor,
        CurrencyCode:              offering.CurrencyCode,
        ParticipantCapacity:       offering.ParticipantCapacity,
        AvailableParticipantUnits: copyInt64(value.AvailableParticipantUnits),
        Status:                    offering.Status,
    }
}

func writeInvalidRequest(c *gin.Context, message string) {
    httpx.WriteError(c, http.StatusBadRequest, "invalid_request", message, httpx.RequestID(c.Request.Context()), nil)
}

func writePublicOfferingError(c *gin.Context, logger *slog.Logger, err error) {
    requestID := httpx.RequestID(c.Request.Context())
    switch {
    case errors.Is(err, ErrNotFound):
        httpx.WriteError(c, http.StatusNotFound, "not_found", "resource not found", requestID, nil)
    case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidCursor):
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid request", requestID, nil)
    case httpx.IsDependencyUnavailable(err):
        logPublicOfferingError(logger, slog.LevelWarn, requestID, err)
        httpx.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "service temporarily unavailable", requestID, nil)
    default:
        logPublicOfferingError(logger, slog.LevelError, requestID, err)
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "internal server error", requestID, nil)
    }
}

func logPublicOfferingError(logger *slog.Logger, level slog.Level, requestID string, err error) {
    if logger == nil {
        return
    }
    logger.Log(context.Background(), level, "public offering query failed", "request_id", requestID, "error_type", fmt.Sprintf("%T", err))
}
