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

    eventapplication "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/application"
    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    platformauth "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/gin-gonic/gin"
)

const (
    commandBodyLimit = 64 << 10
    replayRetention  = 24 * time.Hour
)

type Handler struct {
    db      *sql.DB
    service *eventapplication.Service
    cipher  idempotency.Cipher
    logger  *slog.Logger
}

func NewHandler(db *sql.DB, service *eventapplication.Service, cipher idempotency.Cipher, logger *slog.Logger) *Handler {
    return &Handler{db: db, service: service, cipher: cipher, logger: logger}
}

func (handler *Handler) list(c *gin.Context) {
    input, err := operationsListInput(c)
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    result, err := handler.service.List(c.Request.Context(), input)
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    data := make([]operationsEvent, 0, len(result.Events))
    for _, value := range result.Events {
        data = append(data, toOperationsEvent(value))
    }
    c.JSON(http.StatusOK, operationsEventListResponse{Data: data, Page: operationsPage{Limit: result.Limit, NextCursor: result.NextCursor}})
}

func (handler *Handler) get(c *gin.Context) {
    id, ok := operationsEventID(c)
    if !ok {
        return
    }
    value, err := handler.service.Get(c.Request.Context(), id)
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsEventResponse{Data: toOperationsEvent(value)})
}

func (handler *Handler) create(c *gin.Context) {
    var request createEventRequest
    if !decodeEventCommand(c, &request) {
        return
    }
    input := request.input()
    mutation, err := eventdomain.Create(input)
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    principal, ok := operationsPrincipal(c)
    if !ok {
        return
    }
    key, ok := operationsIdempotencyKey(c)
    if !ok {
        return
    }
    namespace, _ := idempotency.Namespace("operations", mutation.Action, principal.OperatorID)
    hash, err := idempotency.RequestHash(mutation.After)
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    actor := eventActor(c, principal)
    response, _, err := idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, idempotency.Command{Namespace: namespace, Key: key, RequestHash: hash, Retention: replayRetention}, func(tx *sql.Tx) (idempotency.Response, error) {
        created, createErr := handler.service.CreateInTransaction(c.Request.Context(), tx, input, actor)
        if createErr != nil {
            return idempotency.Response{}, createErr
        }
        return eventReplayResponse(http.StatusCreated, created)
    })
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    writeReplayResponse(c, response)
}

func (handler *Handler) patch(c *gin.Context) {
    id, ok := operationsEventID(c)
    if !ok {
        return
    }
    var request patchEventRequest
    if !decodeEventCommand(c, &request) {
        return
    }
    input, err := request.input()
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    principal, ok := operationsPrincipal(c)
    if !ok {
        return
    }
    result, err := handler.service.Update(c.Request.Context(), id, input, eventActor(c, principal))
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsEventResponse{Data: toOperationsEvent(result)})
}

func (handler *Handler) transition(action string, command eventapplication.Transition) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, ok := operationsEventID(c)
        if !ok {
            return
        }
        var request eventTransitionRequest
        if !decodeEventCommand(c, &request) {
            return
        }
        if request.ExpectedVersion <= 0 || request.ExpectedVersion > eventdomain.MaxSafeInteger {
            writeEventOperationsError(c, handler.logger, eventdomain.ErrInvalidInput)
            return
        }
        principal, ok := operationsPrincipal(c)
        if !ok {
            return
        }
        key, ok := operationsIdempotencyKey(c)
        if !ok {
            return
        }
        namespace, _ := idempotency.Namespace("operations", action, principal.OperatorID)
        hash, err := idempotency.RequestHash(struct {
            ID              string `json:"id"`
            ExpectedVersion int64  `json:"expected_version"`
        }{ID: id, ExpectedVersion: request.ExpectedVersion})
        if err != nil {
            writeEventOperationsError(c, handler.logger, err)
            return
        }
        actor := eventActor(c, principal)
        response, _, err := idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, idempotency.Command{Namespace: namespace, Key: key, RequestHash: hash, Retention: replayRetention}, func(tx *sql.Tx) (idempotency.Response, error) {
            saved, transitionErr := handler.service.TransitionInTransaction(c.Request.Context(), tx, id, eventdomain.TransitionInput{ExpectedVersion: request.ExpectedVersion}, command, actor)
            if transitionErr != nil {
                return idempotency.Response{}, transitionErr
            }
            return eventReplayResponse(http.StatusOK, saved)
        })
        if err != nil {
            writeEventOperationsError(c, handler.logger, err)
            return
        }
        writeReplayResponse(c, response)
    }
}

func operationsEventID(c *gin.Context) (string, bool) {
    id, err := httpx.CanonicalUUID(c.Param("event_id"))
    if err != nil {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid event_id", httpx.RequestID(c.Request.Context()), nil)
        return "", false
    }
    return id, true
}

func operationsPrincipal(c *gin.Context) (platformauth.Principal, bool) {
    principal, ok := platformauth.FromContext(c.Request.Context())
    if !ok {
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", httpx.RequestID(c.Request.Context()), nil)
    }
    return principal, ok
}

func eventActor(c *gin.Context, principal platformauth.Principal) eventapplication.Actor {
    return eventapplication.Actor{OperatorID: principal.OperatorID, RequestID: httpx.RequestID(c.Request.Context()), ClientIP: safeClientIP(c.ClientIP())}
}

func operationsIdempotencyKey(c *gin.Context) (string, bool) {
    value, err := httpx.IdempotencyKey(c.GetHeader("Idempotency-Key"))
    if err != nil {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid Idempotency-Key", httpx.RequestID(c.Request.Context()), nil)
        return "", false
    }
    return value, true
}

func decodeEventCommand(c *gin.Context, destination any) bool {
    err := httpx.DecodeJSON(c.Writer, c.Request, destination, commandBodyLimit)
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

func safeClientIP(value string) string {
    address, err := netip.ParseAddr(value)
    if err != nil {
        return ""
    }
    return address.Unmap().String()
}

func writeReplayResponse(c *gin.Context, response idempotency.Response) {
    c.Data(response.Status, "application/json", response.Body)
}

func writeEventOperationsError(c *gin.Context, logger *slog.Logger, err error) {
    requestID := httpx.RequestID(c.Request.Context())
    switch {
    case errors.Is(err, httpx.ErrInvalidJSON):
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid request", requestID, nil)
    case errors.Is(err, eventdomain.ErrInvalidInput), errors.Is(err, eventdomain.ErrNoChanges):
        httpx.WriteError(c, http.StatusUnprocessableEntity, "validation_failed", "request validation failed", requestID, nil)
    case errors.Is(err, eventdomain.ErrNotFound):
        httpx.WriteError(c, http.StatusNotFound, "not_found", "resource not found", requestID, nil)
    case errors.Is(err, eventdomain.ErrStaleVersion):
        httpx.WriteError(c, http.StatusConflict, "stale_version", "resource version is stale", requestID, nil)
    case errors.Is(err, eventdomain.ErrInvalidTransition), errors.Is(err, eventdomain.ErrDuplicateYear), errors.Is(err, eventdomain.ErrActiveConflict), errors.Is(err, eventdomain.ErrConfigurationFrozen), errors.Is(err, eventdomain.ErrImmutable):
        httpx.WriteError(c, http.StatusConflict, "state_conflict", "resource state conflicts with the request", requestID, nil)
    case errors.Is(err, idempotency.ErrConflict), errors.Is(err, idempotency.ErrUnavailableReplay):
        httpx.WriteError(c, http.StatusConflict, "idempotency_conflict", "idempotency key conflicts with the request", requestID, nil)
    case errors.Is(err, eventdomain.ErrInvalidCursor):
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid list parameters", requestID, nil)
    case httpx.IsDependencyUnavailable(err):
        logEventOperationsError(logger, slog.LevelWarn, requestID, err)
        httpx.WriteError(c, http.StatusServiceUnavailable, "service_unavailable", "service temporarily unavailable", requestID, nil)
    default:
        logEventOperationsError(logger, slog.LevelError, requestID, err)
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", requestID, nil)
    }
}

func logEventOperationsError(logger *slog.Logger, level slog.Level, requestID string, err error) {
    if logger != nil {
        logger.Log(context.Background(), level, "operations event command failed", "request_id", requestID, "error_type", fmt.Sprintf("%T", err))
    }
}
