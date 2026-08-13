package offering

import (
    "errors"
    "fmt"
    "strings"
    "time"
    "unicode/utf8"

    "github.com/Kangditya/persona-apps/apps/api/internal/event"
)

const (
    MaxCodeLength          = 64
    MaxNameLength          = 255
    MaxKindLength          = 64
    MaxDescriptionLength   = 10_000
    MaxSafeInteger         = event.MaxSafeInteger
    MaxParticipantCapacity = int64(2_147_483_647)
)

var (
    ErrNotFound            = errors.New("offering not found")
    ErrInvalidInput        = errors.New("invalid offering input")
    ErrInvalidTransition   = errors.New("invalid offering transition")
    ErrInvalidParentState  = errors.New("invalid parent event state")
    ErrStaleVersion        = errors.New("stale offering version")
    ErrDuplicateCode       = errors.New("duplicate offering code")
    ErrInvalidCursor       = errors.New("invalid offering cursor")
    ErrConfigurationFrozen = errors.New("offering configuration is frozen")
    ErrImmutable           = errors.New("offering is immutable")
    ErrNoChanges           = errors.New("offering update has no changes")
    ErrInvalidAvailability = errors.New("invalid offering availability")
)

type Status string

const (
    StatusDraft       Status = "DRAFT"
    StatusPublished   Status = "PUBLISHED"
    StatusUnavailable Status = "UNAVAILABLE"
    StatusArchived    Status = "ARCHIVED"
)

type Offering struct {
    ID                  string
    EventID             string
    Code                string
    Name                string
    Kind                string
    Description         *string
    PriceMinor          int64
    CurrencyCode        string
    ParticipantCapacity int32
    ParticipantQuota    *int64
    Status              Status
    PublishedAt         *time.Time
    Version             int64
    CreatedAt           time.Time
    UpdatedAt           time.Time
}

type CreateInput struct {
    EventID             string
    Code                string
    Name                string
    Kind                string
    Description         *string
    PriceMinor          int64
    CurrencyCode        string
    ParticipantCapacity int64
    ParticipantQuota    *int64
}

type OptionalString struct {
    Set   bool
    Value *string
}

type OptionalInt64 struct {
    Set   bool
    Value *int64
}

type UpdateInput struct {
    ExpectedVersion  int64
    Name             *string
    Description      OptionalString
    PriceMinor       *int64
    ParticipantQuota OptionalInt64
}

type PublishInput struct {
    ExpectedVersion int64
    OccurredAt      time.Time
}

type TransitionInput struct {
    ExpectedVersion int64
}

type Snapshot struct {
    EventID             string     `json:"event_id"`
    Code                string     `json:"code"`
    Name                string     `json:"name"`
    Kind                string     `json:"offering_kind"`
    Description         *string    `json:"description"`
    PriceMinor          int64      `json:"price_minor"`
    CurrencyCode        string     `json:"currency_code"`
    ParticipantCapacity int32      `json:"participant_capacity"`
    ParticipantQuota    *int64     `json:"participant_quota"`
    Status              Status     `json:"status"`
    PublishedAt         *time.Time `json:"published_at"`
    Version             int64      `json:"version"`
}

func (offering Offering) Snapshot() Snapshot {
    return Snapshot{
        EventID:             offering.EventID,
        Code:                offering.Code,
        Name:                offering.Name,
        Kind:                offering.Kind,
        Description:         copyString(offering.Description),
        PriceMinor:          offering.PriceMinor,
        CurrencyCode:        offering.CurrencyCode,
        ParticipantCapacity: offering.ParticipantCapacity,
        ParticipantQuota:    copyInt64(offering.ParticipantQuota),
        Status:              offering.Status,
        PublishedAt:         copyTime(offering.PublishedAt),
        Version:             offering.Version,
    }
}

func validStatus(status Status) bool {
    switch status {
    case StatusDraft, StatusPublished, StatusUnavailable, StatusArchived:
        return true
    default:
        return false
    }
}

func validateCreate(parent event.Event, input CreateInput) (Offering, error) {
    if !allowsCommercialParent(parent.Status) {
        return Offering{}, ErrInvalidParentState
    }
    if input.ParticipantCapacity < 1 || input.ParticipantCapacity > MaxParticipantCapacity {
        return Offering{}, fmt.Errorf("%w: participant capacity must be between 1 and %d", ErrInvalidInput, MaxParticipantCapacity)
    }
    offering := Offering{
        EventID:             strings.TrimSpace(input.EventID),
        Code:                input.Code,
        Name:                input.Name,
        Kind:                input.Kind,
        Description:         copyString(input.Description),
        PriceMinor:          input.PriceMinor,
        CurrencyCode:        input.CurrencyCode,
        ParticipantCapacity: int32(input.ParticipantCapacity),
        ParticipantQuota:    copyInt64(input.ParticipantQuota),
        Status:              StatusDraft,
        Version:             1,
    }
    if err := validateOffering(&offering); err != nil {
        return Offering{}, err
    }
    return offering, nil
}

func validateCurrent(offering Offering) error {
    if !validStatus(offering.Status) {
        return fmt.Errorf("%w: unsupported status", ErrInvalidInput)
    }
    if offering.Version <= 0 || offering.Version > MaxSafeInteger {
        return fmt.Errorf("%w: version must be between 1 and %d", ErrInvalidInput, MaxSafeInteger)
    }
    return validateOffering(&offering)
}

func validateOffering(offering *Offering) error {
    offering.EventID = strings.TrimSpace(offering.EventID)
    if offering.EventID == "" {
        return fmt.Errorf("%w: event ID is required", ErrInvalidInput)
    }
    if err := validateRequiredText("code", &offering.Code, MaxCodeLength); err != nil {
        return err
    }
    if err := validateRequiredText("name", &offering.Name, MaxNameLength); err != nil {
        return err
    }
    if err := validateRequiredText("offering kind", &offering.Kind, MaxKindLength); err != nil {
        return err
    }
    if offering.Description != nil && utf8.RuneCountInString(*offering.Description) > MaxDescriptionLength {
        return fmt.Errorf("%w: description must contain at most %d characters", ErrInvalidInput, MaxDescriptionLength)
    }
    if offering.PriceMinor < 0 || offering.PriceMinor > MaxSafeInteger {
        return fmt.Errorf("%w: price minor must be between 0 and %d", ErrInvalidInput, MaxSafeInteger)
    }
    if !validCurrency(offering.CurrencyCode) {
        return fmt.Errorf("%w: currency code must be three uppercase letters", ErrInvalidInput)
    }
    if offering.ParticipantCapacity < 1 {
        return fmt.Errorf("%w: participant capacity must be between 1 and %d", ErrInvalidInput, MaxParticipantCapacity)
    }
    offering.ParticipantQuota = copyInt64(offering.ParticipantQuota)
    if offering.ParticipantQuota != nil && (*offering.ParticipantQuota < 0 || *offering.ParticipantQuota > MaxSafeInteger) {
        return fmt.Errorf("%w: participant quota must be between 0 and %d", ErrInvalidInput, MaxSafeInteger)
    }
    return nil
}

func validateRequiredText(name string, value *string, maximum int) error {
    *value = strings.TrimSpace(*value)
    if *value == "" {
        return fmt.Errorf("%w: %s is required", ErrInvalidInput, name)
    }
    if utf8.RuneCountInString(*value) > maximum {
        return fmt.Errorf("%w: %s must contain at most %d characters", ErrInvalidInput, name, maximum)
    }
    return nil
}

func validCurrency(value string) bool {
    if len(value) != 3 {
        return false
    }
    for _, character := range value {
        if character < 'A' || character > 'Z' {
            return false
        }
    }
    return true
}

func allowsCommercialParent(status event.Status) bool {
    switch status {
    case event.StatusDraft, event.StatusPublished, event.StatusActive, event.StatusSuspended:
        return true
    default:
        return false
    }
}

func copyString(value *string) *string {
    if value == nil {
        return nil
    }
    copy := *value
    return &copy
}

func copyInt64(value *int64) *int64 {
    if value == nil {
        return nil
    }
    copy := *value
    return &copy
}

func copyTime(value *time.Time) *time.Time {
    if value == nil {
        return nil
    }
    copy := value.UTC()
    return &copy
}
