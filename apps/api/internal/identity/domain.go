package identity

import (
    "errors"
    "fmt"
    "strings"
    "time"
    "unicode/utf8"
)

const (
    MaxDisplayNameLength = 255
    MaxEmailLength       = 320
    MaxPhoneLength       = 64
)

var (
    ErrInvalidInput = errors.New("invalid party input")
    ErrNotFound     = errors.New("party not found")
)

type PartyType string

const (
    PartyTypePerson       PartyType = "PERSON"
    PartyTypeOrganization PartyType = "ORGANIZATION"
)

type Party struct {
    ID          string
    Type        PartyType
    DisplayName string
    Email       *string
    Phone       *string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type CreateInput struct {
    Type        PartyType
    DisplayName string
    Email       *string
    Phone       *string
}

func NewParty(input CreateInput) (Party, error) {
    party := Party{
        Type:        input.Type,
        DisplayName: strings.TrimSpace(input.DisplayName),
        Email:       copyString(input.Email),
        Phone:       copyString(input.Phone),
    }
    if party.Type != PartyTypePerson && party.Type != PartyTypeOrganization {
        return Party{}, fmt.Errorf("%w: party type must be PERSON or ORGANIZATION", ErrInvalidInput)
    }
    if party.DisplayName == "" || utf8.RuneCountInString(party.DisplayName) > MaxDisplayNameLength {
        return Party{}, fmt.Errorf("%w: display name must contain 1 to %d characters", ErrInvalidInput, MaxDisplayNameLength)
    }

    var err error
    if party.Email, err = normalizeContact(party.Email, MaxEmailLength, "email"); err != nil {
        return Party{}, err
    }
    if party.Phone, err = normalizeContact(party.Phone, MaxPhoneLength, "phone"); err != nil {
        return Party{}, err
    }
    return party, nil
}

func normalizeContact(value *string, maximum int, name string) (*string, error) {
    if value == nil {
        return nil, nil
    }
    normalized := strings.TrimSpace(*value)
    if normalized == "" || utf8.RuneCountInString(normalized) > maximum {
        return nil, fmt.Errorf("%w: %s must contain 1 to %d characters when present", ErrInvalidInput, name, maximum)
    }
    return &normalized, nil
}

func copyString(value *string) *string {
    if value == nil {
        return nil
    }
    copied := *value
    return &copied
}
