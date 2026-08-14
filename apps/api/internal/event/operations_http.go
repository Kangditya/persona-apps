package event

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
    "net/netip"
    "strconv"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/audit"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/outbox"
    "github.com/gin-gonic/gin"
)

const (
    commandBodyLimit = 64 << 10
    replayRetention  = 24 * time.Hour
    operationsSource = "operations-web"
)

type OperationsHandler struct {
    db     *sql.DB
    cipher idempotency.Cipher
    logger *slog.Logger
    now    func() time.Time
}

type operationsEvent struct {
    ID                   string     `json:"id"`
    EventYear            int        `json:"event_year"`
    Name                 string     `json:"name"`
    Status               Status     `json:"status"`
    RegistrationOpensAt  *time.Time `json:"registration_opens_at,omitempty"`
    RegistrationClosesAt *time.Time `json:"registration_closes_at,omitempty"`
    ParticipantQuota     *int64     `json:"participant_quota,omitempty"`
    Version              int64      `json:"version"`
    CreatedAt            time.Time  `json:"created_at"`
    UpdatedAt            time.Time  `json:"updated_at"`
}

type operationsEventResponse struct {
    Data operationsEvent `json:"data"`
}

type operationsEventListResponse struct {
    Data []operationsEvent `json:"data"`
    Page operationsPage    `json:"page"`
}

type operationsPage struct {
    Limit      int    `json:"limit"`
    NextCursor string `json:"next_cursor,omitempty"`
}

type createEventRequest struct {
    EventYear            int        `json:"event_year"`
    Name                 string     `json:"name"`
    RegistrationOpensAt  *time.Time `json:"registration_opens_at"`
    RegistrationClosesAt *time.Time `json:"registration_closes_at"`
    ParticipantQuota     *int64     `json:"participant_quota"`
}

type patchEventRequest struct {
    ExpectedVersion      int64           `json:"expected_version"`
    Name                 json.RawMessage `json:"name"`
    RegistrationOpensAt  json.RawMessage `json:"registration_opens_at"`
    RegistrationClosesAt json.RawMessage `json:"registration_closes_at"`
    ParticipantQuota     json.RawMessage `json:"participant_quota"`
}

type eventTransitionRequest struct {
    ExpectedVersion int64 `json:"expected_version"`
}

func NewOperationsHandler(db *sql.DB, cipher idempotency.Cipher, logger *slog.Logger) *OperationsHandler {
    return &OperationsHandler{db: db, cipher: cipher, logger: logger, now: time.Now}
}

func (handler *OperationsHandler) RegisterRoutes(group *gin.RouterGroup, authentication *auth.Service) {
    group.GET("/events", authentication.Require("event.read"), handler.list)
    group.GET("/events/:event_id", authentication.Require("event.read"), handler.get)
    group.POST("/events", authentication.Require("event.manage"), handler.create)
    group.PATCH("/events/:event_id", authentication.Require("event.manage"), handler.patch)
    group.POST("/events/:event_id/publish", authentication.Require("event.manage"), handler.transition(ActionPublish, Publish))
    group.POST("/events/:event_id/activate", authentication.Require("event.manage"), handler.transition(ActionActivate, Activate))
    group.POST("/events/:event_id/suspend", authentication.Require("event.manage"), handler.transition(ActionSuspend, Suspend))
    group.POST("/events/:event_id/close", authentication.Require("event.manage"), handler.transition(ActionClose, Close))
    group.POST("/events/:event_id/archive", authentication.Require("event.manage"), handler.transition(ActionArchive, Archive))
}

func (handler *OperationsHandler) list(c *gin.Context) {
    input, err := operationsListInput(c)
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    result, err := NewRepository(handler.db).List(c.Request.Context(), input)
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

func (handler *OperationsHandler) get(c *gin.Context) {
    id, ok := operationsEventID(c)
    if !ok {
        return
    }
    value, err := NewRepository(handler.db).Get(c.Request.Context(), id)
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsEventResponse{Data: toOperationsEvent(value)})
}

func (handler *OperationsHandler) create(c *gin.Context) {
    var request createEventRequest
    if !decodeEventCommand(c, &request) {
        return
    }
    mutation, err := Create(CreateInput{
        EventYear: request.EventYear, Name: request.Name,
        RegistrationOpensAt: request.RegistrationOpensAt, RegistrationClosesAt: request.RegistrationClosesAt,
        ParticipantQuota: request.ParticipantQuota,
    })
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
    response, _, err := idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, idempotency.Command{Namespace: namespace, Key: key, RequestHash: hash, Retention: replayRetention}, func(tx *sql.Tx) (idempotency.Response, error) {
        created, createErr := NewRepository(tx).Create(c.Request.Context(), mutation.Event)
        if createErr != nil {
            return idempotency.Response{}, createErr
        }
        if auditErr := audit.Write(c.Request.Context(), tx, handler.auditEntry(c, principal, mutation.Action, "Event", created.ID, nil, created.Snapshot())); auditErr != nil {
            return idempotency.Response{}, auditErr
        }
        return eventReplayResponse(http.StatusCreated, created)
    })
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    writeReplayResponse(c, response)
}

func (handler *OperationsHandler) patch(c *gin.Context) {
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
    var result Event
    err = database.Within(c.Request.Context(), handler.db, func(tx *sql.Tx) error {
        repository := NewRepository(tx)
        current, getErr := repository.GetForUpdate(c.Request.Context(), id)
        if getErr != nil {
            return getErr
        }
        mutation, updateErr := Update(current, input)
        if updateErr != nil {
            return updateErr
        }
        if mutation.Action == "" {
            result = mutation.Event
            return nil
        }
        saved, saveErr := repository.Save(c.Request.Context(), mutation.Event, input.ExpectedVersion)
        if saveErr != nil {
            return saveErr
        }
        if auditErr := audit.Write(c.Request.Context(), tx, handler.auditEntry(c, principal, mutation.Action, "Event", saved.ID, mutation.Before, saved.Snapshot())); auditErr != nil {
            return auditErr
        }
        result = saved
        return nil
    })
    if err != nil {
        writeEventOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsEventResponse{Data: toOperationsEvent(result)})
}

type eventTransition func(Event, TransitionInput) (Mutation, error)

func (handler *OperationsHandler) transition(action string, command eventTransition) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, ok := operationsEventID(c)
        if !ok {
            return
        }
        var request eventTransitionRequest
        if !decodeEventCommand(c, &request) {
            return
        }
        if request.ExpectedVersion <= 0 || request.ExpectedVersion > MaxSafeInteger {
            writeEventOperationsError(c, handler.logger, ErrInvalidInput)
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
        response, _, err := idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, idempotency.Command{Namespace: namespace, Key: key, RequestHash: hash, Retention: replayRetention}, func(tx *sql.Tx) (idempotency.Response, error) {
            repository := NewRepository(tx)
            current, getErr := repository.GetForUpdate(c.Request.Context(), id)
            if getErr != nil {
                return idempotency.Response{}, getErr
            }
            mutation, transitionErr := command(current, TransitionInput{ExpectedVersion: request.ExpectedVersion})
            if transitionErr != nil {
                return idempotency.Response{}, transitionErr
            }
            saved, saveErr := repository.Save(c.Request.Context(), mutation.Event, request.ExpectedVersion)
            if saveErr != nil {
                return idempotency.Response{}, saveErr
            }
            if auditErr := audit.Write(c.Request.Context(), tx, handler.auditEntry(c, principal, mutation.Action, "Event", saved.ID, mutation.Before, saved.Snapshot())); auditErr != nil {
                return idempotency.Response{}, auditErr
            }
            if mutation.OutboxEventType != "" {
                if outboxErr := outbox.Write(c.Request.Context(), tx, outbox.Event{AggregateType: "Event", AggregateID: saved.ID, EventType: mutation.OutboxEventType, Payload: map[string]any{"event_id": saved.ID, "status": saved.Status, "version": saved.Version}, OccurredAt: handler.now()}); outboxErr != nil {
                    return idempotency.Response{}, outboxErr
                }
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

func (request patchEventRequest) input() (UpdateInput, error) {
    input := UpdateInput{ExpectedVersion: request.ExpectedVersion}
    fields := 0
    if request.Name != nil {
        fields++
        var value string
        if string(request.Name) == "null" || json.Unmarshal(request.Name, &value) != nil {
            return UpdateInput{}, httpx.ErrInvalidJSON
        }
        input.Name = &value
    }
    if request.RegistrationOpensAt != nil {
        fields++
        value, err := nullableTime(request.RegistrationOpensAt)
        if err != nil {
            return UpdateInput{}, err
        }
        input.RegistrationOpensAt = OptionalTime{Set: true, Value: value}
    }
    if request.RegistrationClosesAt != nil {
        fields++
        value, err := nullableTime(request.RegistrationClosesAt)
        if err != nil {
            return UpdateInput{}, err
        }
        input.RegistrationClosesAt = OptionalTime{Set: true, Value: value}
    }
    if request.ParticipantQuota != nil {
        fields++
        value, err := nullableInt64(request.ParticipantQuota)
        if err != nil {
            return UpdateInput{}, err
        }
        input.ParticipantQuota = OptionalInt64{Set: true, Value: value}
    }
    if fields == 0 {
        return UpdateInput{}, ErrNoChanges
    }
    return input, nil
}

func nullableTime(raw json.RawMessage) (*time.Time, error) {
    if string(raw) == "null" {
        return nil, nil
    }
    var value time.Time
    if json.Unmarshal(raw, &value) != nil {
        return nil, httpx.ErrInvalidJSON
    }
    value = value.UTC()
    return &value, nil
}

func nullableInt64(raw json.RawMessage) (*int64, error) {
    if string(raw) == "null" {
        return nil, nil
    }
    var value int64
    if json.Unmarshal(raw, &value) != nil {
        return nil, httpx.ErrInvalidJSON
    }
    return &value, nil
}

func operationsListInput(c *gin.Context) (ListInput, error) {
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
        if err != nil {
            return ListInput{}, ErrInvalidInput
        }
        input.Limit = limit
    }
    if _, _, err := normalizeListInput(input); err != nil {
        return ListInput{}, err
    }
    return input, nil
}

func operationsEventID(c *gin.Context) (string, bool) {
    id, err := httpx.CanonicalUUID(c.Param("event_id"))
    if err != nil {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid event_id", httpx.RequestID(c.Request.Context()), nil)
        return "", false
    }
    return id, true
}

func operationsPrincipal(c *gin.Context) (auth.Principal, bool) {
    principal, ok := auth.FromContext(c.Request.Context())
    if !ok {
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", httpx.RequestID(c.Request.Context()), nil)
    }
    return principal, ok
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

func (handler *OperationsHandler) auditEntry(c *gin.Context, principal auth.Principal, action, targetType, targetID string, before, after any) audit.Entry {
    return audit.Entry{
        ActorOperatorID: principal.OperatorID, ActorReference: principal.OperatorID,
        Action: action, TargetType: targetType, TargetID: targetID,
        RequestID: httpx.RequestID(c.Request.Context()), Source: operationsSource,
        Permission: "event.manage", ClientIP: safeClientIP(c.ClientIP()), Before: before, After: after,
    }
}

func safeClientIP(value string) string {
    address, err := netip.ParseAddr(value)
    if err != nil {
        return ""
    }
    return address.Unmap().String()
}

func eventReplayResponse(status int, value Event) (idempotency.Response, error) {
    body, err := json.Marshal(operationsEventResponse{Data: toOperationsEvent(value)})
    if err != nil {
        return idempotency.Response{}, err
    }
    return idempotency.Response{Status: status, Body: body}, nil
}

func writeReplayResponse(c *gin.Context, response idempotency.Response) {
    c.Data(response.Status, "application/json", response.Body)
}

func toOperationsEvent(value Event) operationsEvent {
    return operationsEvent{
        ID: value.ID, EventYear: value.EventYear, Name: value.Name, Status: value.Status,
        RegistrationOpensAt: copyTime(value.RegistrationOpensAt), RegistrationClosesAt: copyTime(value.RegistrationClosesAt),
        ParticipantQuota: copyInt64(value.ParticipantQuota), Version: value.Version,
        CreatedAt: value.CreatedAt.UTC(), UpdatedAt: value.UpdatedAt.UTC(),
    }
}

func writeEventOperationsError(c *gin.Context, logger *slog.Logger, err error) {
    requestID := httpx.RequestID(c.Request.Context())
    switch {
    case errors.Is(err, httpx.ErrInvalidJSON):
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid request", requestID, nil)
    case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrNoChanges):
        httpx.WriteError(c, http.StatusUnprocessableEntity, "validation_failed", "request validation failed", requestID, nil)
    case errors.Is(err, ErrNotFound):
        httpx.WriteError(c, http.StatusNotFound, "not_found", "resource not found", requestID, nil)
    case errors.Is(err, ErrStaleVersion):
        httpx.WriteError(c, http.StatusConflict, "stale_version", "resource version is stale", requestID, nil)
    case errors.Is(err, ErrInvalidTransition), errors.Is(err, ErrDuplicateYear), errors.Is(err, ErrActiveConflict), errors.Is(err, ErrConfigurationFrozen), errors.Is(err, ErrImmutable):
        httpx.WriteError(c, http.StatusConflict, "state_conflict", "resource state conflicts with the request", requestID, nil)
    case errors.Is(err, idempotency.ErrConflict), errors.Is(err, idempotency.ErrUnavailableReplay):
        httpx.WriteError(c, http.StatusConflict, "idempotency_conflict", "idempotency key conflicts with the request", requestID, nil)
    case errors.Is(err, ErrInvalidCursor):
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
