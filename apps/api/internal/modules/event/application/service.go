package application

import (
    "context"
    "database/sql"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    platformaudit "github.com/Kangditya/persona-apps/apps/api/internal/platform/audit"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/outbox"
)

type Repository interface {
    Create(context.Context, eventdomain.Event) (eventdomain.Event, error)
    Get(context.Context, string) (eventdomain.Event, error)
    GetForUpdate(context.Context, string) (eventdomain.Event, error)
    Active(context.Context) (*eventdomain.Event, error)
    List(context.Context, eventdomain.ListInput) (eventdomain.ListResult, error)
    Save(context.Context, eventdomain.Event, int64) (eventdomain.Event, error)
}

type RepositoryFactory func(platformdb.DBTX) Repository

type Actor struct {
    OperatorID string
    RequestID  string
    ClientIP   string
}

type Transition func(eventdomain.Event, eventdomain.TransitionInput) (eventdomain.Mutation, error)

type Service struct {
    db           *sql.DB
    repositories RepositoryFactory
    now          func() time.Time
}

func NewService(db *sql.DB, repositories RepositoryFactory) *Service {
    return &Service{db: db, repositories: repositories, now: time.Now}
}

func (service *Service) Active(ctx context.Context) (*eventdomain.Event, error) {
    return service.repositories(service.db).Active(ctx)
}

func (service *Service) List(ctx context.Context, input eventdomain.ListInput) (eventdomain.ListResult, error) {
    return service.repositories(service.db).List(ctx, input)
}

func (service *Service) Get(ctx context.Context, id string) (eventdomain.Event, error) {
    return service.repositories(service.db).Get(ctx, id)
}

func (service *Service) GetForUpdate(ctx context.Context, tx *sql.Tx, id string) (eventdomain.Event, error) {
    return service.repositories(tx).GetForUpdate(ctx, id)
}

func (service *Service) CreateInTransaction(ctx context.Context, tx *sql.Tx, input eventdomain.CreateInput, actor Actor) (eventdomain.Event, error) {
    mutation, err := eventdomain.Create(input)
    if err != nil {
        return eventdomain.Event{}, err
    }
    created, err := service.repositories(tx).Create(ctx, mutation.Event)
    if err != nil {
        return eventdomain.Event{}, err
    }
    if err := writeAudit(ctx, tx, actor, mutation.Action, "Event", created.ID, nil, created.Snapshot()); err != nil {
        return eventdomain.Event{}, err
    }
    return created, nil
}

func (service *Service) Update(ctx context.Context, id string, input eventdomain.UpdateInput, actor Actor) (eventdomain.Event, error) {
    var result eventdomain.Event
    err := platformdb.Within(ctx, service.db, func(tx *sql.Tx) error {
        var err error
        result, err = service.UpdateInTransaction(ctx, tx, id, input, actor)
        return err
    })
    return result, err
}

func (service *Service) UpdateInTransaction(ctx context.Context, tx *sql.Tx, id string, input eventdomain.UpdateInput, actor Actor) (eventdomain.Event, error) {
    repository := service.repositories(tx)
    current, err := repository.GetForUpdate(ctx, id)
    if err != nil {
        return eventdomain.Event{}, err
    }
    mutation, err := eventdomain.Update(current, input)
    if err != nil {
        return eventdomain.Event{}, err
    }
    if mutation.Action == "" {
        return mutation.Event, nil
    }
    saved, err := repository.Save(ctx, mutation.Event, input.ExpectedVersion)
    if err != nil {
        return eventdomain.Event{}, err
    }
    if err := writeAudit(ctx, tx, actor, mutation.Action, "Event", saved.ID, mutation.Before, saved.Snapshot()); err != nil {
        return eventdomain.Event{}, err
    }
    return saved, nil
}

func (service *Service) TransitionInTransaction(ctx context.Context, tx *sql.Tx, id string, input eventdomain.TransitionInput, command Transition, actor Actor) (eventdomain.Event, error) {
    repository := service.repositories(tx)
    current, err := repository.GetForUpdate(ctx, id)
    if err != nil {
        return eventdomain.Event{}, err
    }
    mutation, err := command(current, input)
    if err != nil {
        return eventdomain.Event{}, err
    }
    saved, err := repository.Save(ctx, mutation.Event, input.ExpectedVersion)
    if err != nil {
        return eventdomain.Event{}, err
    }
    if err := writeAudit(ctx, tx, actor, mutation.Action, "Event", saved.ID, mutation.Before, saved.Snapshot()); err != nil {
        return eventdomain.Event{}, err
    }
    if mutation.OutboxEventType != "" {
        if err := outbox.Write(ctx, tx, outbox.Event{
            AggregateType: "Event", AggregateID: saved.ID, EventType: mutation.OutboxEventType,
            Payload: map[string]any{"event_id": saved.ID, "status": saved.Status, "version": saved.Version}, OccurredAt: service.clock(),
        }); err != nil {
            return eventdomain.Event{}, err
        }
    }
    return saved, nil
}

func (service *Service) clock() time.Time {
    if service.now == nil {
        return time.Now().UTC()
    }
    return service.now().UTC()
}

func writeAudit(ctx context.Context, tx *sql.Tx, actor Actor, action, targetType, targetID string, before, after any) error {
    return platformaudit.Write(ctx, tx, platformaudit.Entry{
        ActorOperatorID: actor.OperatorID, ActorReference: actor.OperatorID,
        Action: action, TargetType: targetType, TargetID: targetID,
        RequestID: actor.RequestID, Source: "operations-web", Permission: "event.manage",
        ClientIP: actor.ClientIP, Before: before, After: after,
    })
}
