package purchasing

import (
    "database/sql"
    "errors"
    "log/slog"
    "net/http"
    "strconv"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/gin-gonic/gin"
)

type OperationsHandler struct {
    db     *sql.DB
    logger *slog.Logger
}

type operationsPartySummary struct {
    ID          string `json:"id"`
    DisplayName string `json:"display_name"`
}

type operationsParticipantSnapshot struct {
    SequenceNo  int    `json:"sequence_no"`
    DisplayName string `json:"display_name"`
}

type operationsPurchase struct {
    ID                   string                          `json:"id"`
    EventID              string                          `json:"event_id"`
    PurchaseRef          string                          `json:"purchase_ref"`
    Channel              Channel                         `json:"channel"`
    Purchaser            operationsPartySummary          `json:"purchaser"`
    Payer                *operationsPartySummary         `json:"payer,omitempty"`
    OfferingID           string                          `json:"offering_id"`
    OfferingNameSnapshot string                          `json:"offering_name_snapshot"`
    ParticipantCount     int                             `json:"participant_count"`
    Participants         []operationsParticipantSnapshot `json:"participants,omitempty"`
    TotalAmountMinor     int64                           `json:"total_amount_minor"`
    CurrencyCode         string                          `json:"currency_code"`
    Status               Status                          `json:"status"`
    CreatedAt            time.Time                       `json:"created_at"`
    UpdatedAt            time.Time                       `json:"updated_at"`
}

type operationsPurchaseResponse struct {
    Data operationsPurchase `json:"data"`
}

type operationsPurchaseListResponse struct {
    Data []operationsPurchase `json:"data"`
    Page operationsPage       `json:"page"`
}

type operationsPage struct {
    Limit      int    `json:"limit"`
    NextCursor string `json:"next_cursor,omitempty"`
}

func NewOperationsHandler(db *sql.DB, logger *slog.Logger) *OperationsHandler {
    return &OperationsHandler{db: db, logger: logger}
}

func (handler *OperationsHandler) RegisterRoutes(group *gin.RouterGroup, authentication *auth.Service) {
    group.GET("/purchases", authentication.Require("purchase.read"), handler.list)
    group.GET("/purchases/:purchase_id", authentication.Require("purchase.read"), handler.get)
}

func (handler *OperationsHandler) list(c *gin.Context) {
    input, err := operationsListInput(c)
    if err != nil {
        writeOperationsError(c, handler.logger, err)
        return
    }
    result, err := NewRepository(handler.db).List(c.Request.Context(), input)
    if err != nil {
        writeOperationsError(c, handler.logger, err)
        return
    }
    purchases := make([]operationsPurchase, 0, len(result.Purchases))
    for _, detail := range result.Purchases {
        purchases = append(purchases, toOperationsPurchase(detail, false))
    }
    c.JSON(http.StatusOK, operationsPurchaseListResponse{Data: purchases, Page: operationsPage{Limit: result.Limit, NextCursor: result.NextCursor}})
}

func (handler *OperationsHandler) get(c *gin.Context) {
    id, err := httpx.CanonicalUUID(c.Param("purchase_id"))
    if err != nil {
        writeOperationsError(c, handler.logger, ErrInvalidInput)
        return
    }
    detail, err := NewRepository(handler.db).Get(c.Request.Context(), id)
    if err != nil {
        writeOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsPurchaseResponse{Data: toOperationsPurchase(detail, true)})
}

func operationsListInput(c *gin.Context) (ListInput, error) {
    values := c.Request.URL.Query()
    for name := range values {
        if name != "cursor" && name != "limit" && name != "event_id" && name != "status" {
            return ListInput{}, ErrInvalidInput
        }
    }
    input := ListInput{}
    if values, found := values["cursor"]; found {
        if len(values) != 1 || values[0] == "" {
            return ListInput{}, ErrInvalidCursor
        }
        input.Cursor = values[0]
    }
    if values, found := values["limit"]; found {
        if len(values) != 1 || values[0] == "" {
            return ListInput{}, ErrInvalidInput
        }
        limit, err := strconv.Atoi(values[0])
        if err != nil {
            return ListInput{}, ErrInvalidInput
        }
        input.Limit = limit
    }
    if values, found := values["event_id"]; found {
        if len(values) != 1 {
            return ListInput{}, ErrInvalidInput
        }
        eventID, err := httpx.CanonicalUUID(values[0])
        if err != nil {
            return ListInput{}, ErrInvalidInput
        }
        input.EventID = eventID
    }
    if values, found := values["status"]; found {
        if len(values) != 1 || values[0] == "" {
            return ListInput{}, ErrInvalidInput
        }
        input.Status = Status(values[0])
    }
    if _, _, err := normalizeListInput(input); err != nil {
        return ListInput{}, err
    }
    return input, nil
}

func toOperationsPurchase(detail Detail, includeParticipants bool) operationsPurchase {
    purchase := detail.Purchase
    result := operationsPurchase{
        ID: purchase.ID, EventID: purchase.EventID, PurchaseRef: purchase.PurchaseRef, Channel: purchase.Channel,
        Purchaser: operationsPartySummary{ID: detail.Purchaser.ID, DisplayName: detail.Purchaser.DisplayName},
        OfferingID: purchase.OfferingID, OfferingNameSnapshot: purchase.OfferingNameSnapshot,
        ParticipantCount: purchase.ParticipantCount, TotalAmountMinor: purchase.TotalAmountMinor,
        CurrencyCode: purchase.CurrencyCode, Status: purchase.Status,
        CreatedAt: purchase.CreatedAt.UTC(), UpdatedAt: purchase.UpdatedAt.UTC(),
    }
    if detail.Payer != nil {
        result.Payer = &operationsPartySummary{ID: detail.Payer.ID, DisplayName: detail.Payer.DisplayName}
    }
    if includeParticipants {
        result.Participants = make([]operationsParticipantSnapshot, 0, len(purchase.Participants))
        for _, participant := range purchase.Participants {
            result.Participants = append(result.Participants, operationsParticipantSnapshot{SequenceNo: participant.SequenceNo, DisplayName: participant.DisplayName})
        }
    }
    return result
}

func writeOperationsError(c *gin.Context, logger *slog.Logger, err error) {
    switch {
    case errors.Is(err, ErrNotFound):
        httpx.WriteError(c, http.StatusNotFound, "not_found", "purchase not found", httpx.RequestID(c.Request.Context()), nil)
    case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidCursor):
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid purchase request", httpx.RequestID(c.Request.Context()), nil)
    case httpx.IsDependencyUnavailable(err):
        httpx.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "service unavailable", httpx.RequestID(c.Request.Context()), nil)
    default:
        if logger != nil {
            logger.Error("purchase operations request failed", "error", err)
        }
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", httpx.RequestID(c.Request.Context()), nil)
    }
}
