package domain

import (
    "fmt"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
)

const (
    ActionCreate      = "offering.create"
    ActionUpdate      = "offering.update"
    ActionPublish     = "offering.publish"
    ActionUnavailable = "offering.unavailable"
    ActionArchive     = "offering.archive"

    OutboxOfferingPublished   = "OfferingPublished"
    OutboxOfferingUnavailable = "OfferingUnavailable"
    OutboxOfferingArchived    = "OfferingArchived"
)

type Mutation struct {
    Offering        Offering
    Action          string
    RetrySensitive  bool
    Before          *Snapshot
    After           Snapshot
    OutboxEventType string
}

type ReservationUsage struct {
    Reserved int64
    Consumed int64
    Released int64
    Expired  int64
}

type AvailabilityInput struct {
    EventQuota    *int64
    OfferingQuota *int64
    EventUsage    ReservationUsage
    OfferingUsage ReservationUsage
}

func Create(parent eventdomain.Event, input CreateInput) (Mutation, error) {
    offering, err := validateCreate(parent, input)
    if err != nil {
        return Mutation{}, err
    }
    return Mutation{
        Offering:       offering,
        Action:         ActionCreate,
        RetrySensitive: true,
        After:          offering.Snapshot(),
    }, nil
}

func Update(parent eventdomain.Event, current Offering, input UpdateInput) (Mutation, error) {
    if err := validateExpectedVersion(current, input.ExpectedVersion); err != nil {
        return Mutation{}, err
    }
    if !allowsCommercialParent(parent.Status) {
        return Mutation{}, ErrInvalidParentState
    }
    switch current.Status {
    case StatusDraft, StatusUnavailable:
    case StatusPublished:
        return Mutation{}, ErrConfigurationFrozen
    case StatusArchived:
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
    if input.Description.Set {
        next.Description = copyString(input.Description.Value)
        changed = true
    }
    if input.PriceMinor != nil {
        next.PriceMinor = *input.PriceMinor
        changed = true
    }
    if input.ParticipantQuota.Set {
        next.ParticipantQuota = copyInt64(input.ParticipantQuota.Value)
        changed = true
    }
    if !changed {
        return Mutation{}, ErrNoChanges
    }
    if err := validateOffering(&next); err != nil {
        return Mutation{}, err
    }
    if sameConfiguration(current, next) {
        return Mutation{Offering: clone(current), After: current.Snapshot()}, nil
    }
    if err := incrementVersion(&next); err != nil {
        return Mutation{}, err
    }
    before := current.Snapshot()
    return Mutation{
        Offering: next,
        Action:   ActionUpdate,
        Before:   &before,
        After:    next.Snapshot(),
    }, nil
}

func sameConfiguration(left, right Offering) bool {
    return left.Name == right.Name &&
        sameString(left.Description, right.Description) &&
        left.PriceMinor == right.PriceMinor &&
        equalOptionalInt64(left.ParticipantQuota, right.ParticipantQuota)
}

func sameString(left, right *string) bool {
    return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func equalOptionalInt64(left, right *int64) bool {
    return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func Publish(parent eventdomain.Event, current Offering, input PublishInput) (Mutation, error) {
    if err := validateExpectedVersion(current, input.ExpectedVersion); err != nil {
        return Mutation{}, err
    }
    if !allowsPublish(parent.Status) {
        return Mutation{}, ErrInvalidParentState
    }
    if current.Status != StatusDraft && current.Status != StatusUnavailable {
        return Mutation{}, fmt.Errorf("%w: %s cannot become %s", ErrInvalidTransition, current.Status, StatusPublished)
    }
    next := clone(current)
    next.Status = StatusPublished
    if next.PublishedAt == nil {
        if input.OccurredAt.IsZero() {
            return Mutation{}, fmt.Errorf("%w: publication time is required", ErrInvalidInput)
        }
        now := input.OccurredAt.UTC()
        next.PublishedAt = &now
    }
    if err := incrementVersion(&next); err != nil {
        return Mutation{}, err
    }
    before := current.Snapshot()
    return Mutation{
        Offering:        next,
        Action:          ActionPublish,
        RetrySensitive:  true,
        Before:          &before,
        After:           next.Snapshot(),
        OutboxEventType: OutboxOfferingPublished,
    }, nil
}

func MarkUnavailable(current Offering, input TransitionInput) (Mutation, error) {
    return transition(current, input, StatusUnavailable, ActionUnavailable, OutboxOfferingUnavailable, StatusPublished)
}

func Archive(current Offering, input TransitionInput) (Mutation, error) {
    return transition(current, input, StatusArchived, ActionArchive, OutboxOfferingArchived, StatusDraft, StatusPublished, StatusUnavailable)
}

func IsPublic(parent eventdomain.Event, offering Offering) bool {
    return parent.Status == eventdomain.StatusActive && offering.Status == StatusPublished
}

func AvailableParticipantUnits(input AvailabilityInput) (*int64, error) {
    eventRemaining, err := remaining(input.EventQuota, input.EventUsage)
    if err != nil {
        return nil, err
    }
    offeringRemaining, err := remaining(input.OfferingQuota, input.OfferingUsage)
    if err != nil {
        return nil, err
    }
    switch {
    case eventRemaining == nil:
        return offeringRemaining, nil
    case offeringRemaining == nil:
        return eventRemaining, nil
    case *eventRemaining <= *offeringRemaining:
        return eventRemaining, nil
    default:
        return offeringRemaining, nil
    }
}

func transition(current Offering, input TransitionInput, target Status, action, outboxEventType string, allowed ...Status) (Mutation, error) {
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
                Offering:        next,
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

func validateExpectedVersion(current Offering, expected int64) error {
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

func allowsPublish(status eventdomain.Status) bool {
    switch status {
    case eventdomain.StatusPublished, eventdomain.StatusActive, eventdomain.StatusSuspended:
        return true
    default:
        return false
    }
}

func incrementVersion(offering *Offering) error {
    if offering.Version >= MaxSafeInteger {
        return fmt.Errorf("%w: version cannot increase further", ErrInvalidInput)
    }
    offering.Version++
    return nil
}

func remaining(quota *int64, usage ReservationUsage) (*int64, error) {
    if quota != nil && (*quota < 0 || *quota > MaxSafeInteger) {
        return nil, ErrInvalidAvailability
    }
    used, err := usedUnits(usage)
    if err != nil {
        return nil, err
    }
    if quota == nil {
        return nil, nil
    }
    if used >= *quota {
        return int64Pointer(0), nil
    }
    return int64Pointer(*quota - used), nil
}

func usedUnits(usage ReservationUsage) (int64, error) {
    if usage.Reserved < 0 || usage.Consumed < 0 || usage.Released < 0 || usage.Expired < 0 {
        return 0, ErrInvalidAvailability
    }
    if usage.Reserved >= MaxSafeInteger || usage.Consumed >= MaxSafeInteger || usage.Reserved > MaxSafeInteger-usage.Consumed {
        return MaxSafeInteger, nil
    }
    return usage.Reserved + usage.Consumed, nil
}

func clone(offering Offering) Offering {
    offering.Description = copyString(offering.Description)
    offering.ParticipantQuota = copyInt64(offering.ParticipantQuota)
    offering.PublishedAt = copyTime(offering.PublishedAt)
    return offering
}

func int64Pointer(value int64) *int64 {
    return &value
}
