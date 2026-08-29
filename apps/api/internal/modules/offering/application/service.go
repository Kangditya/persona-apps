package application

import (
    "context"
    "database/sql"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
    platformaudit "github.com/Kangditya/persona-apps/apps/api/internal/platform/audit"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/outbox"
)

type Repository interface {
    Create(context.Context, offeringdomain.Offering) (offeringdomain.Offering, error)
    Get(context.Context, string) (offeringdomain.Offering, error)
    GetForUpdate(context.Context, string) (offeringdomain.Offering, error)
    List(context.Context, string, offeringdomain.ListInput) (offeringdomain.ListResult, error)
    ListWithAvailability(context.Context, string, offeringdomain.ListInput) (offeringdomain.CatalogueResult, error)
    GetWithAvailability(context.Context, string) (offeringdomain.CatalogueOffering, error)
    Save(context.Context, offeringdomain.Offering, int64) (offeringdomain.Offering, error)
    ListPublic(context.Context, string, offeringdomain.ListInput) (offeringdomain.CatalogueResult, error)
    GetPublic(context.Context, string) (offeringdomain.CatalogueOffering, error)
}

type RepositoryFactory func(platformdb.DBTX) Repository

type EventReader interface {
    Get(context.Context, string) (eventdomain.Event, error)
    GetForUpdate(context.Context, *sql.Tx, string) (eventdomain.Event, error)
}

type Actor struct {
    OperatorID string
    RequestID  string
    ClientIP   string
}

type Transition func(eventdomain.Event, offeringdomain.Offering, offeringdomain.TransitionInput, time.Time) (offeringdomain.Mutation, error)

type Service struct {
    db           *sql.DB
    repositories RepositoryFactory
    events       EventReader
    now          func() time.Time
}

func NewService(db *sql.DB, repositories RepositoryFactory, events EventReader) *Service {
    return &Service{db: db, repositories: repositories, events: events, now: time.Now}
}

func (service *Service) List(ctx context.Context, eventID string, input offeringdomain.ListInput) (offeringdomain.CatalogueResult, error) {
    if _, err := service.events.Get(ctx, eventID); err != nil {
        return offeringdomain.CatalogueResult{}, err
    }
    return service.repositories(service.db).ListWithAvailability(ctx, eventID, input)
}

func (service *Service) Get(ctx context.Context, id string) (offeringdomain.CatalogueOffering, error) {
    return service.repositories(service.db).GetWithAvailability(ctx, id)
}

func (service *Service) ListPublic(ctx context.Context, eventID string, input offeringdomain.ListInput) (offeringdomain.CatalogueResult, error) {
    return service.repositories(service.db).ListPublic(ctx, eventID, input)
}

func (service *Service) GetPublic(ctx context.Context, id string) (offeringdomain.CatalogueOffering, error) {
    return service.repositories(service.db).GetPublic(ctx, id)
}

func (service *Service) ValidateCreate(ctx context.Context, parentID string, input offeringdomain.CreateInput) (offeringdomain.Mutation, error) {
    parent, err := service.events.Get(ctx, parentID)
    if err != nil {
        return offeringdomain.Mutation{}, err
    }
    return offeringdomain.Create(parent, input)
}

func (service *Service) CreateInTransaction(ctx context.Context, tx *sql.Tx, parentID string, input offeringdomain.CreateInput, actor Actor) (offeringdomain.CatalogueOffering, error) {
    parent, err := service.events.GetForUpdate(ctx, tx, parentID)
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    mutation, err := offeringdomain.Create(parent, input)
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    repository := service.repositories(tx)
    created, err := repository.Create(ctx, mutation.Offering)
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    if err := writeAudit(ctx, tx, actor, mutation.Action, created.ID, nil, created.Snapshot()); err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    return repository.GetWithAvailability(ctx, created.ID)
}

func (service *Service) Update(ctx context.Context, id string, input offeringdomain.UpdateInput, actor Actor) (offeringdomain.CatalogueOffering, error) {
    var result offeringdomain.CatalogueOffering
    err := platformdb.Within(ctx, service.db, func(tx *sql.Tx) error {
        var err error
        result, err = service.UpdateInTransaction(ctx, tx, id, input, actor)
        return err
    })
    return result, err
}

func (service *Service) UpdateInTransaction(ctx context.Context, tx *sql.Tx, id string, input offeringdomain.UpdateInput, actor Actor) (offeringdomain.CatalogueOffering, error) {
    repository := service.repositories(tx)
    current, err := repository.GetForUpdate(ctx, id)
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    parent, err := service.events.GetForUpdate(ctx, tx, current.EventID)
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    mutation, err := offeringdomain.Update(parent, current, input)
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    if mutation.Action != "" {
        saved, saveErr := repository.Save(ctx, mutation.Offering, input.ExpectedVersion)
        if saveErr != nil {
            return offeringdomain.CatalogueOffering{}, saveErr
        }
        if err := writeAudit(ctx, tx, actor, mutation.Action, saved.ID, mutation.Before, saved.Snapshot()); err != nil {
            return offeringdomain.CatalogueOffering{}, err
        }
    }
    return repository.GetWithAvailability(ctx, id)
}

func (service *Service) TransitionInTransaction(ctx context.Context, tx *sql.Tx, id string, input offeringdomain.TransitionInput, command Transition, actor Actor) (offeringdomain.CatalogueOffering, error) {
    repository := service.repositories(tx)
    current, err := repository.GetForUpdate(ctx, id)
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    parent, err := service.events.GetForUpdate(ctx, tx, current.EventID)
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    mutation, err := command(parent, current, input, service.clock())
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    saved, err := repository.Save(ctx, mutation.Offering, input.ExpectedVersion)
    if err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    if err := writeAudit(ctx, tx, actor, mutation.Action, saved.ID, mutation.Before, saved.Snapshot()); err != nil {
        return offeringdomain.CatalogueOffering{}, err
    }
    if mutation.OutboxEventType != "" {
        if err := outbox.Write(ctx, tx, outbox.Event{
            AggregateType: "Offering", AggregateID: saved.ID, EventType: mutation.OutboxEventType,
            Payload: map[string]any{"offering_id": saved.ID, "event_id": saved.EventID, "status": saved.Status, "version": saved.Version}, OccurredAt: service.clock(),
        }); err != nil {
            return offeringdomain.CatalogueOffering{}, err
        }
    }
    return repository.GetWithAvailability(ctx, saved.ID)
}

func (service *Service) clock() time.Time {
    if service.now == nil {
        return time.Now().UTC()
    }
    return service.now().UTC()
}

func writeAudit(ctx context.Context, tx *sql.Tx, actor Actor, action, targetID string, before, after any) error {
    return platformaudit.Write(ctx, tx, platformaudit.Entry{
        ActorOperatorID: actor.OperatorID, ActorReference: actor.OperatorID,
        Action: action, TargetType: "Offering", TargetID: targetID,
        RequestID: actor.RequestID, Source: "operations-web", Permission: "offering.manage",
        ClientIP: actor.ClientIP, Before: before, After: after,
    })
}
