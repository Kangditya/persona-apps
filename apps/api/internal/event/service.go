package event

import "fmt"

const (
    ActionCreate   = "event.create"
    ActionUpdate   = "event.update"
    ActionPublish  = "event.publish"
    ActionActivate = "event.activate"
    ActionSuspend  = "event.suspend"
    ActionClose    = "event.close"
    ActionArchive  = "event.archive"

    OutboxEventPublished = "EventPublished"
    OutboxEventActivated = "EventActivated"
    OutboxEventSuspended = "EventSuspended"
    OutboxEventClosed    = "EventClosed"
    OutboxEventArchived  = "EventArchived"
)

type Mutation struct {
    Event           Event
    Action          string
    RetrySensitive  bool
    Before          *Snapshot
    After           Snapshot
    OutboxEventType string
}

func Create(input CreateInput) (Mutation, error) {
    event, err := validateCreate(input)
    if err != nil {
        return Mutation{}, err
    }
    return Mutation{
        Event:          event,
        Action:         ActionCreate,
        RetrySensitive: true,
        After:          event.Snapshot(),
    }, nil
}

func Update(current Event, input UpdateInput) (Mutation, error) {
    if err := validateExpectedVersion(current, input.ExpectedVersion); err != nil {
        return Mutation{}, err
    }
    switch current.Status {
    case StatusDraft, StatusPublished, StatusSuspended:
    case StatusActive:
        return Mutation{}, ErrConfigurationFrozen
    case StatusClosed, StatusArchived:
        return Mutation{}, ErrImmutable
    default:
        return Mutation{}, fmt.Errorf("%w: unsupported status", ErrInvalidInput)
    }

    next := clone(current)
    changed := false
    if input.Name != nil {
        next.Name = *input.Name
        changed = true
    }
    if input.RegistrationOpensAt.Set {
        next.RegistrationOpensAt = copyTime(input.RegistrationOpensAt.Value)
        changed = true
    }
    if input.RegistrationClosesAt.Set {
        next.RegistrationClosesAt = copyTime(input.RegistrationClosesAt.Value)
        changed = true
    }
    if input.ParticipantQuota.Set {
        next.ParticipantQuota = copyInt64(input.ParticipantQuota.Value)
        changed = true
    }
    if !changed {
        return Mutation{}, ErrNoChanges
    }
    if err := validateConfiguration(&next); err != nil {
        return Mutation{}, err
    }
    if err := incrementVersion(&next); err != nil {
        return Mutation{}, err
    }
    before := current.Snapshot()
    return Mutation{
        Event:  next,
        Action: ActionUpdate,
        Before: &before,
        After:  next.Snapshot(),
    }, nil
}

func Publish(current Event, input TransitionInput) (Mutation, error) {
    return transition(current, input, StatusPublished, ActionPublish, OutboxEventPublished, StatusDraft)
}

func Activate(current Event, input TransitionInput) (Mutation, error) {
    return transition(current, input, StatusActive, ActionActivate, OutboxEventActivated, StatusPublished, StatusSuspended)
}

func Suspend(current Event, input TransitionInput) (Mutation, error) {
    return transition(current, input, StatusSuspended, ActionSuspend, OutboxEventSuspended, StatusActive)
}

func Close(current Event, input TransitionInput) (Mutation, error) {
    return transition(current, input, StatusClosed, ActionClose, OutboxEventClosed, StatusActive, StatusSuspended)
}

func Archive(current Event, input TransitionInput) (Mutation, error) {
    return transition(current, input, StatusArchived, ActionArchive, OutboxEventArchived, StatusClosed)
}

func transition(current Event, input TransitionInput, target Status, action, outboxEventType string, allowed ...Status) (Mutation, error) {
    if err := validateExpectedVersion(current, input.ExpectedVersion); err != nil {
        return Mutation{}, err
    }
    for _, status := range allowed {
        if current.Status == status {
            next := clone(current)
            next.Status = target
            if err := incrementVersion(&next); err != nil {
                return Mutation{}, err
            }
            before := current.Snapshot()
            return Mutation{
                Event:           next,
                Action:          action,
                RetrySensitive:  true,
                Before:          &before,
                After:           next.Snapshot(),
                OutboxEventType: outboxEventType,
            }, nil
        }
    }
    return Mutation{}, fmt.Errorf("%w: %s cannot become %s", ErrInvalidTransition, current.Status, target)
}

func validateExpectedVersion(current Event, expected int64) error {
    if err := validateCurrent(current); err != nil {
        return err
    }
    if expected <= 0 {
        return fmt.Errorf("%w: expected version must be positive", ErrInvalidInput)
    }
    if current.Version != expected {
        return ErrStaleVersion
    }
    return nil
}

func incrementVersion(event *Event) error {
    if event.Version >= MaxSafeInteger {
        return fmt.Errorf("%w: version cannot increase further", ErrInvalidInput)
    }
    event.Version++
    return nil
}

func clone(event Event) Event {
    event.RegistrationOpensAt = copyTime(event.RegistrationOpensAt)
    event.RegistrationClosesAt = copyTime(event.RegistrationClosesAt)
    event.ParticipantQuota = copyInt64(event.ParticipantQuota)
    return event
}
