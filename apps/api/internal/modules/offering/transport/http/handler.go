package http

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "net/netip"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    offeringapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/application"
    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
    platformauth "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/gin-gonic/gin"
)

const offeringReplayRetention = 24 * time.Hour

type Handler struct {
    db      *sql.DB
    service *offeringapplication.Service
    cipher  idempotency.Cipher
    logger  *slog.Logger
}

func NewHandler(db *sql.DB, service *offeringapplication.Service, cipher idempotency.Cipher, logger *slog.Logger) *Handler {
    return &Handler{db: db, service: service, cipher: cipher, logger: logger}
}

func (handler *Handler) list(c *gin.Context) {
    eventID, ok := operationsUUID(c, "event_id")
    if !ok {
        return
    }
    input, err := offeringListInput(c)
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    result, err := handler.service.List(c.Request.Context(), eventID, input)
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    data := make([]operationsOffering, 0, len(result.Offerings))
    for _, value := range result.Offerings {
        data = append(data, toOperationsOffering(value))
    }
    c.JSON(http.StatusOK, operationsOfferingListResponse{Data: data, Page: operationsPage{Limit: result.Limit, NextCursor: result.NextCursor}})
}

func (handler *Handler) get(c *gin.Context) {
    id, ok := operationsUUID(c, "offering_id")
    if !ok {
        return
    }
    value, err := handler.service.Get(c.Request.Context(), id)
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsOfferingResponse{Data: toOperationsOffering(value)})
}

func (handler *Handler) create(c *gin.Context) {
    eventID, ok := operationsUUID(c, "event_id")
    if !ok {
        return
    }
    var request createOfferingRequest
    if !decodeOfferingCommand(c, &request) {
        return
    }
    principal, ok := offeringPrincipal(c)
    if !ok {
        return
    }
    key, ok := offeringIdempotencyKey(c)
    if !ok {
        return
    }
    input := request.input(eventID)
    validated, err := handler.service.ValidateCreate(c.Request.Context(), eventID, input)
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    hash, err := idempotency.RequestHash(validated.After)
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    namespace, _ := idempotency.Namespace("operations", offeringdomain.ActionCreate, principal.OperatorID)
    actor := offeringActor(c, principal)
    response, _, err := idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, idempotency.Command{Namespace: namespace, Key: key, RequestHash: hash, Retention: offeringReplayRetention}, func(tx *sql.Tx) (idempotency.Response, error) {
        view, createErr := handler.service.CreateInTransaction(c.Request.Context(), tx, eventID, input, actor)
        if createErr != nil {
            return idempotency.Response{}, createErr
        }
        return offeringReplayResponse(http.StatusCreated, view)
    })
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    writeOfferingReplayResponse(c, response)
}

func (handler *Handler) patch(c *gin.Context) {
    id, ok := operationsUUID(c, "offering_id")
    if !ok {
        return
    }
    var request patchOfferingRequest
    if !decodeOfferingCommand(c, &request) {
        return
    }
    input, err := request.input()
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    principal, ok := offeringPrincipal(c)
    if !ok {
        return
    }
    result, err := handler.service.Update(c.Request.Context(), id, input, offeringActor(c, principal))
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsOfferingResponse{Data: toOperationsOffering(result)})
}

func (handler *Handler) transition(action string, command offeringapplication.Transition) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, ok := operationsUUID(c, "offering_id")
        if !ok {
            return
        }
        var request offeringTransitionRequest
        if !decodeOfferingCommand(c, &request) {
            return
        }
        if request.ExpectedVersion <= 0 || request.ExpectedVersion > offeringdomain.MaxSafeInteger {
            writeOfferingOperationsError(c, handler.logger, offeringdomain.ErrInvalidInput)
            return
        }
        principal, ok := offeringPrincipal(c)
        if !ok {
            return
        }
        key, ok := offeringIdempotencyKey(c)
        if !ok {
            return
        }
        namespace, _ := idempotency.Namespace("operations", action, principal.OperatorID)
        hash, err := idempotency.RequestHash(struct {
            ID              string `json:"id"`
            ExpectedVersion int64  `json:"expected_version"`
        }{ID: id, ExpectedVersion: request.ExpectedVersion})
        if err != nil {
            writeOfferingOperationsError(c, handler.logger, err)
            return
        }
        actor := offeringActor(c, principal)
        response, _, err := idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, idempotency.Command{Namespace: namespace, Key: key, RequestHash: hash, Retention: offeringReplayRetention}, func(tx *sql.Tx) (idempotency.Response, error) {
            view, transitionErr := handler.service.TransitionInTransaction(c.Request.Context(), tx, id, offeringdomain.TransitionInput{ExpectedVersion: request.ExpectedVersion}, command, actor)
            if transitionErr != nil {
                return idempotency.Response{}, transitionErr
            }
            return offeringReplayResponse(http.StatusOK, view)
        })
        if err != nil {
            writeOfferingOperationsError(c, handler.logger, err)
            return
        }
        writeOfferingReplayResponse(c, response)
    }
}

func operationsUUID(c *gin.Context, parameter string) (string, bool) {
    id, err := httpx.CanonicalUUID(c.Param(parameter))
    if err != nil {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid "+parameter, httpx.RequestID(c.Request.Context()), nil)
        return "", false
    }
    return id, true
}

func offeringPrincipal(c *gin.Context) (platformauth.Principal, bool) {
    principal, ok := platformauth.FromContext(c.Request.Context())
    if !ok {
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", httpx.RequestID(c.Request.Context()), nil)
    }
    return principal, ok
}

func offeringActor(c *gin.Context, principal platformauth.Principal) offeringapplication.Actor {
    return offeringapplication.Actor{OperatorID: principal.OperatorID, RequestID: httpx.RequestID(c.Request.Context()), ClientIP: offeringClientIP(c.ClientIP())}
}

func offeringIdempotencyKey(c *gin.Context) (string, bool) {
    key, err := httpx.IdempotencyKey(c.GetHeader("Idempotency-Key"))
    if err != nil {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid Idempotency-Key", httpx.RequestID(c.Request.Context()), nil)
        return "", false
    }
    return key, true
}

func decodeOfferingCommand(c *gin.Context, destination any) bool {
    err := httpx.DecodeJSON(c.Writer, c.Request, destination, offeringCommandBodyLimit)
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

func offeringClientIP(value string) string {
    address, err := netip.ParseAddr(value)
    if err != nil {
        return ""
    }
    return address.Unmap().String()
}

func writeOfferingReplayResponse(c *gin.Context, response idempotency.Response) {
    c.Data(response.Status, "application/json", response.Body)
}

func writeOfferingOperationsError(c *gin.Context, logger *slog.Logger, err error) {
    requestID := httpx.RequestID(c.Request.Context())
    switch {
    case errors.Is(err, httpx.ErrInvalidJSON):
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid request", requestID, nil)
    case errors.Is(err, offeringdomain.ErrInvalidInput), errors.Is(err, offeringdomain.ErrNoChanges):
        httpx.WriteError(c, http.StatusUnprocessableEntity, "validation_failed", "request validation failed", requestID, nil)
    case errors.Is(err, offeringdomain.ErrNotFound), errors.Is(err, eventdomain.ErrNotFound):
        httpx.WriteError(c, http.StatusNotFound, "not_found", "resource not found", requestID, nil)
    case errors.Is(err, offeringdomain.ErrStaleVersion):
        httpx.WriteError(c, http.StatusConflict, "stale_version", "resource version is stale", requestID, nil)
    case errors.Is(err, offeringdomain.ErrInvalidTransition), errors.Is(err, offeringdomain.ErrInvalidParentState), errors.Is(err, offeringdomain.ErrDuplicateCode), errors.Is(err, offeringdomain.ErrConfigurationFrozen), errors.Is(err, offeringdomain.ErrImmutable):
        httpx.WriteError(c, http.StatusConflict, "state_conflict", "resource state conflicts with the request", requestID, nil)
    case errors.Is(err, idempotency.ErrConflict), errors.Is(err, idempotency.ErrUnavailableReplay):
        httpx.WriteError(c, http.StatusConflict, "idempotency_conflict", "idempotency key conflicts with the request", requestID, nil)
    case errors.Is(err, offeringdomain.ErrInvalidCursor):
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid list parameters", requestID, nil)
    case httpx.IsDependencyUnavailable(err):
        logOfferingOperationsError(logger, slog.LevelWarn, requestID, err)
        httpx.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "service temporarily unavailable", requestID, nil)
    default:
        logOfferingOperationsError(logger, slog.LevelError, requestID, err)
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", requestID, nil)
    }
}

func logOfferingOperationsError(logger *slog.Logger, level slog.Level, requestID string, err error) {
    if logger != nil {
        logger.Log(context.Background(), level, "operations offering command failed", "request_id", requestID, "error_type", fmt.Sprintf("%T", err))
    }
}
