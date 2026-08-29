package http

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "log/slog"
    "net/http"

    identitydomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/domain"
    purchasingapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/application"
    purchasingdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/domain"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/gin-gonic/gin"
)

const publicPurchaseReferenceTries = 5

type Handler struct {
    db      *sql.DB
    service *purchasingapplication.Service
    cipher  idempotency.Cipher
    logger  *slog.Logger
}

func NewHandler(db *sql.DB, service *purchasingapplication.Service, cipher idempotency.Cipher, logger *slog.Logger) *Handler {
    return &Handler{db: db, service: service, cipher: cipher, logger: logger}
}

func (handler *Handler) create(c *gin.Context) {
    var request publicCreatePurchaseRequest
    if !decodePublicPurchase(c, &request) {
        return
    }
    key, err := httpx.IdempotencyKey(c.GetHeader("Idempotency-Key"))
    if err != nil {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "Idempotency-Key is required and invalid", httpx.RequestID(c.Request.Context()), nil)
        return
    }
    offeringID, err := httpx.CanonicalUUID(request.OfferingID)
    if err != nil {
        writePublicPurchaseError(c, handler.logger, purchasingdomain.ErrInvalidInput)
        return
    }
    relationships, err := request.relationships()
    if err != nil {
        writePublicPurchaseError(c, handler.logger, err)
        return
    }
    hash, err := idempotency.RequestHash(struct {
        OfferingID    string                         `json:"offering_id"`
        Relationships purchasingdomain.Relationships `json:"relationships"`
    }{OfferingID: offeringID, Relationships: relationships})
    if err != nil {
        writePublicPurchaseError(c, handler.logger, err)
        return
    }
    if handler.db == nil {
        writePublicPurchaseError(c, handler.logger, httpx.ErrDependencyUnavailable)
        return
    }
    command := idempotency.Command{Namespace: "storefront.purchase.create.guest", Key: key, RequestHash: hash}
    var response idempotency.Response
    for attempt := 0; attempt < publicPurchaseReferenceTries; attempt++ {
        response, _, err = idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, command, func(tx *sql.Tx) (idempotency.Response, error) {
            purchase, reservation, accessToken, createErr := handler.service.CreateCommonPurchaseInTransaction(c.Request.Context(), tx, offeringID, relationships)
            if createErr != nil {
                return idempotency.Response{}, createErr
            }
            body, marshalErr := json.Marshal(publicPurchaseResponse(purchase, reservation, accessToken))
            if marshalErr != nil {
                return idempotency.Response{}, fmt.Errorf("marshal public purchase response: %w", marshalErr)
            }
            return idempotency.Response{Status: http.StatusCreated, Body: body}, nil
        })
        if !errors.Is(err, purchasingdomain.ErrDuplicateReference) {
            break
        }
    }
    if errors.Is(err, purchasingdomain.ErrDuplicateReference) {
        err = errors.New("purchase reference generation exhausted")
    }
    if err != nil {
        writePublicPurchaseError(c, handler.logger, err)
        return
    }
    c.Data(response.Status, "application/json", response.Body)
}

func (handler *Handler) list(c *gin.Context) {
    input, err := operationsListInput(c)
    if err != nil {
        writeOperationsError(c, handler.logger, err)
        return
    }
    result, err := handler.service.List(c.Request.Context(), input)
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

func (handler *Handler) get(c *gin.Context) {
    id, err := httpx.CanonicalUUID(c.Param("purchase_id"))
    if err != nil {
        writeOperationsError(c, handler.logger, purchasingdomain.ErrInvalidInput)
        return
    }
    detail, err := handler.service.Get(c.Request.Context(), id)
    if err != nil {
        writeOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsPurchaseResponse{Data: toOperationsPurchase(detail, true)})
}

func decodePublicPurchase(c *gin.Context, destination any) bool {
    err := httpx.DecodeJSON(c.Writer, c.Request, destination, publicPurchaseBodyLimit)
    switch {
    case err == nil:
        return true
    case errors.Is(err, httpx.ErrRequestTooLarge):
        httpx.WriteError(c, http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large", httpx.RequestID(c.Request.Context()), nil)
    case errors.Is(err, httpx.ErrUnsupportedMediaType):
        httpx.WriteError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json", httpx.RequestID(c.Request.Context()), nil)
    default:
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid JSON request", httpx.RequestID(c.Request.Context()), nil)
    }
    return false
}

func writePublicPurchaseError(c *gin.Context, logger *slog.Logger, err error) {
    requestID := httpx.RequestID(c.Request.Context())
    switch {
    case errors.Is(err, purchasingdomain.ErrInvalidInput), errors.Is(err, purchasingdomain.ErrInvalidRelationships), errors.Is(err, identitydomain.ErrInvalidInput), errors.Is(err, purchasingdomain.ErrNotFound):
        httpx.WriteError(c, http.StatusUnprocessableEntity, "validation_failed", "purchase validation failed", requestID, nil)
    case errors.Is(err, purchasingdomain.ErrStateConflict):
        httpx.WriteError(c, http.StatusConflict, "state_conflict", "purchase state conflicts with the request", requestID, nil)
    case errors.Is(err, purchasingdomain.ErrQuotaUnavailable):
        httpx.WriteError(c, http.StatusConflict, "quota_unavailable", "participant quota is unavailable", requestID, nil)
    case errors.Is(err, idempotency.ErrConflict), errors.Is(err, idempotency.ErrUnavailableReplay):
        httpx.WriteError(c, http.StatusConflict, "idempotency_conflict", "idempotency key conflicts with the request", requestID, nil)
    case httpx.IsDependencyUnavailable(err):
        logPublicPurchaseError(logger, slog.LevelWarn, requestID, err)
        httpx.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "service temporarily unavailable", requestID, nil)
    default:
        logPublicPurchaseError(logger, slog.LevelError, requestID, err)
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "internal server error", requestID, nil)
    }
}

func writeOperationsError(c *gin.Context, logger *slog.Logger, err error) {
    switch {
    case errors.Is(err, purchasingdomain.ErrNotFound):
        httpx.WriteError(c, http.StatusNotFound, "not_found", "purchase not found", httpx.RequestID(c.Request.Context()), nil)
    case errors.Is(err, purchasingdomain.ErrInvalidInput), errors.Is(err, purchasingdomain.ErrInvalidCursor):
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

func logPublicPurchaseError(logger *slog.Logger, level slog.Level, requestID string, err error) {
    if logger != nil {
        logger.Log(context.Background(), level, "public purchase checkout failed", "request_id", requestID, "error_type", fmt.Sprintf("%T", err))
    }
}
