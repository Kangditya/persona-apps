package event

import (
    "errors"
    "fmt"
    "strings"
    "time"
    "unicode/utf8"
)

const (
    MinEventYear   = 1900
    MaxEventYear   = 9999
    MaxNameLength  = 255
    MaxSafeInteger = int64(9_007_199_254_740_991)
)

var (
    ErrNotFound            = errors.New("event not found")
    ErrInvalidInput        = errors.New("invalid event input")
    ErrInvalidTransition   = errors.New("invalid event transition")
    ErrStaleVersion        = errors.New("stale event version")
    ErrDuplicateYear       = errors.New("duplicate event year")
    ErrActiveConflict      = errors.New("active event conflict")
    ErrInvalidCursor       = errors.New("invalid event cursor")
    ErrConfigurationFrozen = errors.New("event configuration is frozen")
    ErrImmutable           = errors.New("event is immutable")
    ErrNoChanges           = errors.New("event update has no changes")
)

type Status string

const (
    StatusDraft     Status = "DRAFT"
    StatusPublished Status = "PUBLISHED"
    StatusActive    Status = "ACTIVE"
    StatusSuspended Status = "SUSPENDED"
    StatusClosed    Status = "CLOSED"
    StatusArchived  Status = "ARCHIVED"
)

type Event struct {
    ID                   string
    EventYear            int
    Name                 string
    Status               Status
    RegistrationOpensAt  *time.Time
    RegistrationClosesAt *time.Time
    ParticipantQuota     *int64
    Version              int64
    CreatedAt            time.Time
    UpdatedAt            time.Time
}

type CreateInput struct {
    EventYear            int
    Name                 string
    RegistrationOpensAt  *time.Time
    RegistrationClosesAt *time.Time
    ParticipantQuota     *int64
}

type OptionalTime struct {
    Set   bool
    Value *time.Time
}

type OptionalInt64 struct {
    Set   bool
    Value *int64
}

type UpdateInput struct {
    ExpectedVersion      int64
    Name                 *string
    RegistrationOpensAt  OptionalTime
    RegistrationClosesAt OptionalTime
    ParticipantQuota     OptionalInt64
}

type TransitionInput struct {
    ExpectedVersion int64
}

type Snapshot struct {
    EventYear            int        `json:"event_year"`
    Name                 string     `json:"name"`
    Status               Status     `json:"status"`
    RegistrationOpensAt  *time.Time `json:"registration_opens_at"`
    RegistrationClosesAt *time.Time `json:"registration_closes_at"`
    ParticipantQuota     *int64     `json:"participant_quota"`
    Version              int64      `json:"version"`
}

func (event Event) Snapshot() Snapshot {
    return Snapshot{
        EventYear:            event.EventYear,
        Name:                 event.Name,
        Status:               event.Status,
        RegistrationOpensAt:  copyTime(event.RegistrationOpensAt),
        RegistrationClosesAt: copyTime(event.RegistrationClosesAt),
        ParticipantQuota:     copyInt64(event.ParticipantQuota),
        Version:              event.Version,
    }
}

func validateCreate(input CreateInput) (Event, error) {
    event := Event{
        EventYear:            input.EventYear,
        Name:                 input.Name,
        Status:               StatusDraft,
        RegistrationOpensAt:  copyTime(input.RegistrationOpensAt),
        RegistrationClosesAt: copyTime(input.RegistrationClosesAt),
        ParticipantQuota:     copyInt64(input.ParticipantQuota),
        Version:              1,
    }
    if err := validateConfiguration(&event); err != nil {
        return Event{}, err
    }
    return event, nil
}

func validateCurrent(event Event) error {
    if !validStatus(event.Status) {
        return fmt.Errorf("%w: unsupported status", ErrInvalidInput)
    }
    if event.Version <= 0 || event.Version > MaxSafeInteger {
        return fmt.Errorf("%w: version must be between 1 and %d", ErrInvalidInput, MaxSafeInteger)
    }
    return validateConfiguration(&event)
}

func validateConfiguration(event *Event) error {
    if event.EventYear < MinEventYear || event.EventYear > MaxEventYear {
        return fmt.Errorf("%w: event year must be between %d and %d", ErrInvalidInput, MinEventYear, MaxEventYear)
    }

    event.Name = strings.TrimSpace(event.Name)
    if event.Name == "" || utf8.RuneCountInString(event.Name) > MaxNameLength {
        return fmt.Errorf("%w: name must contain 1 to %d characters", ErrInvalidInput, MaxNameLength)
    }

    event.RegistrationOpensAt = copyTime(event.RegistrationOpensAt)
    event.RegistrationClosesAt = copyTime(event.RegistrationClosesAt)
    if event.RegistrationOpensAt != nil && event.RegistrationClosesAt != nil && !event.RegistrationOpensAt.Before(*event.RegistrationClosesAt) {
        return fmt.Errorf("%w: registration opening must be before closing", ErrInvalidInput)
    }

    event.ParticipantQuota = copyInt64(event.ParticipantQuota)
    if event.ParticipantQuota != nil && (*event.ParticipantQuota < 0 || *event.ParticipantQuota > MaxSafeInteger) {
        return fmt.Errorf("%w: participant quota must be between 0 and %d", ErrInvalidInput, MaxSafeInteger)
    }
    return nil
}

func validStatus(status Status) bool {
    switch status {
    case StatusDraft, StatusPublished, StatusActive, StatusSuspended, StatusClosed, StatusArchived:
        return true
    default:
        return false
    }
}

func copyTime(value *time.Time) *time.Time {
    if value == nil {
        return nil
    }
    copy := value.UTC()
    return &copy
}

func copyInt64(value *int64) *int64 {
    if value == nil {
        return nil
    }
    copy := *value
    return &copy
}
