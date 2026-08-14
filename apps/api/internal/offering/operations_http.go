package offering

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

    "github.com/Kangditya/persona-apps/apps/api/internal/event"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/audit"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/auth"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/outbox"
    "github.com/gin-gonic/gin"
)

const (
    offeringCommandBodyLimit = 64 << 10
    offeringReplayRetention  = 24 * time.Hour
)

type OperationsHandler struct {
    db     *sql.DB
    cipher idempotency.Cipher
    logger *slog.Logger
    now    func() time.Time
}

type operationsOffering struct {
    ID                        string     `json:"id"`
    EventID                   string     `json:"event_id"`
    Code                      string     `json:"code"`
    Name                      string     `json:"name"`
    Kind                      string     `json:"offering_kind"`
    Description               *string    `json:"description,omitempty"`
    PriceMinor                int64      `json:"price_minor"`
    CurrencyCode              string     `json:"currency_code"`
    ParticipantCapacity       int32      `json:"participant_capacity"`
    ParticipantQuota          *int64     `json:"participant_quota,omitempty"`
    AvailableParticipantUnits *int64     `json:"available_participant_units"`
    Status                    Status     `json:"status"`
    PublishedAt               *time.Time `json:"published_at,omitempty"`
    Version                   int64      `json:"version"`
    CreatedAt                 time.Time  `json:"created_at"`
    UpdatedAt                 time.Time  `json:"updated_at"`
}

type operationsOfferingResponse struct {
    Data operationsOffering `json:"data"`
}

type operationsOfferingListResponse struct {
    Data []operationsOffering `json:"data"`
    Page operationsPage       `json:"page"`
}

type operationsPage struct {
    Limit      int    `json:"limit"`
    NextCursor string `json:"next_cursor,omitempty"`
}

type createOfferingRequest struct {
    Code                string  `json:"code"`
    Name                string  `json:"name"`
    Kind                string  `json:"offering_kind"`
    Description         *string `json:"description"`
    PriceMinor          int64   `json:"price_minor"`
    CurrencyCode        string  `json:"currency_code"`
    ParticipantCapacity int64   `json:"participant_capacity"`
    ParticipantQuota    *int64  `json:"participant_quota"`
}

type patchOfferingRequest struct {
    ExpectedVersion  int64           `json:"expected_version"`
    Name             json.RawMessage `json:"name"`
    Description      json.RawMessage `json:"description"`
    PriceMinor       json.RawMessage `json:"price_minor"`
    ParticipantQuota json.RawMessage `json:"participant_quota"`
}

type offeringTransitionRequest struct {
    ExpectedVersion int64 `json:"expected_version"`
}

func NewOperationsHandler(db *sql.DB, cipher idempotency.Cipher, logger *slog.Logger) *OperationsHandler {
    return &OperationsHandler{db: db, cipher: cipher, logger: logger, now: time.Now}
}

func (handler *OperationsHandler) RegisterRoutes(group *gin.RouterGroup, authentication *auth.Service) {
    group.GET("/events/:event_id/offerings", authentication.Require("offering.read"), handler.list)
    group.GET("/offerings/:offering_id", authentication.Require("offering.read"), handler.get)
    group.POST("/events/:event_id/offerings", authentication.Require("offering.manage"), handler.create)
    group.PATCH("/offerings/:offering_id", authentication.Require("offering.manage"), handler.patch)
    group.POST("/offerings/:offering_id/publish", authentication.Require("offering.manage"), handler.transition(ActionPublish, func(parent event.Event, current Offering, input TransitionInput, now time.Time) (Mutation, error) {
        return Publish(parent, current, PublishInput{ExpectedVersion: input.ExpectedVersion, OccurredAt: now})
    }))
    group.POST("/offerings/:offering_id/unavailable", authentication.Require("offering.manage"), handler.transition(ActionUnavailable, func(_ event.Event, current Offering, input TransitionInput, _ time.Time) (Mutation, error) {
        return MarkUnavailable(current, input)
    }))
    group.POST("/offerings/:offering_id/archive", authentication.Require("offering.manage"), handler.transition(ActionArchive, func(_ event.Event, current Offering, input TransitionInput, _ time.Time) (Mutation, error) {
        return Archive(current, input)
    }))
}

func (handler *OperationsHandler) list(c *gin.Context) {
    eventID, ok := operationsUUID(c, "event_id")
    if !ok {
        return
    }
    input, err := offeringListInput(c)
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    if _, err := event.NewRepository(handler.db).Get(c.Request.Context(), eventID); err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    result, err := NewRepository(handler.db).ListWithAvailability(c.Request.Context(), eventID, input)
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

func (handler *OperationsHandler) get(c *gin.Context) {
    id, ok := operationsUUID(c, "offering_id")
    if !ok {
        return
    }
    value, err := NewRepository(handler.db).GetWithAvailability(c.Request.Context(), id)
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsOfferingResponse{Data: toOperationsOffering(value)})
}

func (handler *OperationsHandler) create(c *gin.Context) {
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
    parent, err := event.NewRepository(handler.db).Get(c.Request.Context(), eventID)
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    validated, err := Create(parent, CreateInput{
        EventID: eventID, Code: request.Code, Name: request.Name, Kind: request.Kind,
        Description: request.Description, PriceMinor: request.PriceMinor, CurrencyCode: request.CurrencyCode,
        ParticipantCapacity: request.ParticipantCapacity, ParticipantQuota: request.ParticipantQuota,
    })
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    hash, err := idempotency.RequestHash(validated.After)
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    namespace, _ := idempotency.Namespace("operations", ActionCreate, principal.OperatorID)
    response, _, err := idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, idempotency.Command{Namespace: namespace, Key: key, RequestHash: hash, Retention: offeringReplayRetention}, func(tx *sql.Tx) (idempotency.Response, error) {
        parent, parentErr := event.NewRepository(tx).GetForUpdate(c.Request.Context(), eventID)
        if parentErr != nil {
            return idempotency.Response{}, parentErr
        }
        mutation, parentErr := Create(parent, CreateInput{
            EventID: eventID, Code: request.Code, Name: request.Name, Kind: request.Kind,
            Description: request.Description, PriceMinor: request.PriceMinor, CurrencyCode: request.CurrencyCode,
            ParticipantCapacity: request.ParticipantCapacity, ParticipantQuota: request.ParticipantQuota,
        })
        if parentErr != nil {
            return idempotency.Response{}, parentErr
        }
        repository := NewRepository(tx)
        created, createErr := repository.Create(c.Request.Context(), mutation.Offering)
        if createErr != nil {
            return idempotency.Response{}, createErr
        }
        if auditErr := audit.Write(c.Request.Context(), tx, handler.auditEntry(c, principal, mutation.Action, created.ID, nil, created.Snapshot())); auditErr != nil {
            return idempotency.Response{}, auditErr
        }
        view, viewErr := repository.GetWithAvailability(c.Request.Context(), created.ID)
        if viewErr != nil {
            return idempotency.Response{}, viewErr
        }
        return offeringReplayResponse(http.StatusCreated, view)
    })
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    writeOfferingReplayResponse(c, response)
}

func (handler *OperationsHandler) patch(c *gin.Context) {
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
    var result CatalogueOffering
    err = database.Within(c.Request.Context(), handler.db, func(tx *sql.Tx) error {
        repository := NewRepository(tx)
        current, getErr := repository.GetForUpdate(c.Request.Context(), id)
        if getErr != nil {
            return getErr
        }
        parent, parentErr := event.NewRepository(tx).GetForUpdate(c.Request.Context(), current.EventID)
        if parentErr != nil {
            return parentErr
        }
        mutation, updateErr := Update(parent, current, input)
        if updateErr != nil {
            return updateErr
        }
        if mutation.Action != "" {
            saved, saveErr := repository.Save(c.Request.Context(), mutation.Offering, input.ExpectedVersion)
            if saveErr != nil {
                return saveErr
            }
            if auditErr := audit.Write(c.Request.Context(), tx, handler.auditEntry(c, principal, mutation.Action, saved.ID, mutation.Before, saved.Snapshot())); auditErr != nil {
                return auditErr
            }
        }
        view, viewErr := repository.GetWithAvailability(c.Request.Context(), id)
        if viewErr != nil {
            return viewErr
        }
        result = view
        return nil
    })
    if err != nil {
        writeOfferingOperationsError(c, handler.logger, err)
        return
    }
    c.JSON(http.StatusOK, operationsOfferingResponse{Data: toOperationsOffering(result)})
}

type offeringTransition func(event.Event, Offering, TransitionInput, time.Time) (Mutation, error)

func (handler *OperationsHandler) transition(action string, command offeringTransition) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, ok := operationsUUID(c, "offering_id")
        if !ok {
            return
        }
        var request offeringTransitionRequest
        if !decodeOfferingCommand(c, &request) {
            return
        }
        if request.ExpectedVersion <= 0 || request.ExpectedVersion > MaxSafeInteger {
            writeOfferingOperationsError(c, handler.logger, ErrInvalidInput)
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
        response, _, err := idempotency.Execute(c.Request.Context(), handler.db, handler.cipher, idempotency.Command{Namespace: namespace, Key: key, RequestHash: hash, Retention: offeringReplayRetention}, func(tx *sql.Tx) (idempotency.Response, error) {
            repository := NewRepository(tx)
            current, getErr := repository.GetForUpdate(c.Request.Context(), id)
            if getErr != nil {
                return idempotency.Response{}, getErr
            }
            parent, parentErr := event.NewRepository(tx).GetForUpdate(c.Request.Context(), current.EventID)
            if parentErr != nil {
                return idempotency.Response{}, parentErr
            }
            mutation, transitionErr := command(parent, current, TransitionInput{ExpectedVersion: request.ExpectedVersion}, handler.now())
            if transitionErr != nil {
                return idempotency.Response{}, transitionErr
            }
            saved, saveErr := repository.Save(c.Request.Context(), mutation.Offering, request.ExpectedVersion)
            if saveErr != nil {
                return idempotency.Response{}, saveErr
            }
            if auditErr := audit.Write(c.Request.Context(), tx, handler.auditEntry(c, principal, mutation.Action, saved.ID, mutation.Before, saved.Snapshot())); auditErr != nil {
                return idempotency.Response{}, auditErr
            }
            if mutation.OutboxEventType != "" {
                if outboxErr := outbox.Write(c.Request.Context(), tx, outbox.Event{AggregateType: "Offering", AggregateID: saved.ID, EventType: mutation.OutboxEventType, Payload: map[string]any{"offering_id": saved.ID, "event_id": saved.EventID, "status": saved.Status, "version": saved.Version}, OccurredAt: handler.now()}); outboxErr != nil {
                    return idempotency.Response{}, outboxErr
                }
            }
            view, viewErr := repository.GetWithAvailability(c.Request.Context(), saved.ID)
            if viewErr != nil {
                return idempotency.Response{}, viewErr
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

func (request patchOfferingRequest) input() (UpdateInput, error) {
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
    if request.Description != nil {
        fields++
        value, err := requestNullableString(request.Description)
        if err != nil {
            return UpdateInput{}, err
        }
        input.Description = OptionalString{Set: true, Value: value}
    }
    if request.PriceMinor != nil {
        fields++
        var value int64
        if json.Unmarshal(request.PriceMinor, &value) != nil {
            return UpdateInput{}, httpx.ErrInvalidJSON
        }
        input.PriceMinor = &value
    }
    if request.ParticipantQuota != nil {
        fields++
        value, err := requestNullableInt64(request.ParticipantQuota)
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

func requestNullableString(raw json.RawMessage) (*string, error) {
    if string(raw) == "null" {
        return nil, nil
    }
    var value string
    if json.Unmarshal(raw, &value) != nil {
        return nil, httpx.ErrInvalidJSON
    }
    return &value, nil
}

func requestNullableInt64(raw json.RawMessage) (*int64, error) {
    if string(raw) == "null" {
        return nil, nil
    }
    var value int64
    if json.Unmarshal(raw, &value) != nil {
        return nil, httpx.ErrInvalidJSON
    }
    return &value, nil
}

func offeringListInput(c *gin.Context) (ListInput, error) {
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

func operationsUUID(c *gin.Context, parameter string) (string, bool) {
    id, err := httpx.CanonicalUUID(c.Param(parameter))
    if err != nil {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid "+parameter, httpx.RequestID(c.Request.Context()), nil)
        return "", false
    }
    return id, true
}

func offeringPrincipal(c *gin.Context) (auth.Principal, bool) {
    principal, ok := auth.FromContext(c.Request.Context())
    if !ok {
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", httpx.RequestID(c.Request.Context()), nil)
    }
    return principal, ok
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

func (handler *OperationsHandler) auditEntry(c *gin.Context, principal auth.Principal, action, targetID string, before, after any) audit.Entry {
    return audit.Entry{
        ActorOperatorID: principal.OperatorID, ActorReference: principal.OperatorID,
        Action: action, TargetType: "Offering", TargetID: targetID,
        RequestID: httpx.RequestID(c.Request.Context()), Source: "operations-web",
        Permission: "offering.manage", ClientIP: offeringClientIP(c.ClientIP()), Before: before, After: after,
    }
}

func offeringClientIP(value string) string {
    address, err := netip.ParseAddr(value)
    if err != nil {
        return ""
    }
    return address.Unmap().String()
}

func offeringReplayResponse(status int, value CatalogueOffering) (idempotency.Response, error) {
    body, err := json.Marshal(operationsOfferingResponse{Data: toOperationsOffering(value)})
    if err != nil {
        return idempotency.Response{}, err
    }
    return idempotency.Response{Status: status, Body: body}, nil
}

func writeOfferingReplayResponse(c *gin.Context, response idempotency.Response) {
    c.Data(response.Status, "application/json", response.Body)
}

func toOperationsOffering(value CatalogueOffering) operationsOffering {
    offering := value.Offering
    return operationsOffering{
        ID: offering.ID, EventID: offering.EventID, Code: offering.Code, Name: offering.Name, Kind: offering.Kind,
        Description: copyString(offering.Description), PriceMinor: offering.PriceMinor, CurrencyCode: offering.CurrencyCode,
        ParticipantCapacity: offering.ParticipantCapacity, ParticipantQuota: copyInt64(offering.ParticipantQuota),
        AvailableParticipantUnits: copyInt64(value.AvailableParticipantUnits), Status: offering.Status,
        PublishedAt: copyTime(offering.PublishedAt), Version: offering.Version,
        CreatedAt: offering.CreatedAt.UTC(), UpdatedAt: offering.UpdatedAt.UTC(),
    }
}

func writeOfferingOperationsError(c *gin.Context, logger *slog.Logger, err error) {
    requestID := httpx.RequestID(c.Request.Context())
    switch {
    case errors.Is(err, httpx.ErrInvalidJSON):
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid request", requestID, nil)
    case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrNoChanges):
        httpx.WriteError(c, http.StatusUnprocessableEntity, "validation_failed", "request validation failed", requestID, nil)
    case errors.Is(err, ErrNotFound), errors.Is(err, event.ErrNotFound):
        httpx.WriteError(c, http.StatusNotFound, "not_found", "resource not found", requestID, nil)
    case errors.Is(err, ErrStaleVersion):
        httpx.WriteError(c, http.StatusConflict, "stale_version", "resource version is stale", requestID, nil)
    case errors.Is(err, ErrInvalidTransition), errors.Is(err, ErrInvalidParentState), errors.Is(err, ErrDuplicateCode), errors.Is(err, ErrConfigurationFrozen), errors.Is(err, ErrImmutable):
        httpx.WriteError(c, http.StatusConflict, "state_conflict", "resource state conflicts with the request", requestID, nil)
    case errors.Is(err, idempotency.ErrConflict), errors.Is(err, idempotency.ErrUnavailableReplay):
        httpx.WriteError(c, http.StatusConflict, "idempotency_conflict", "idempotency key conflicts with the request", requestID, nil)
    case errors.Is(err, ErrInvalidCursor):
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
