package identity

import (
    "errors"
    "strings"
    "testing"
)

func TestNewParty(t *testing.T) {
    email := "  purchaser@example.test  "
    phone := "  +628100000000  "
    party, err := NewParty(CreateInput{
        Type:        PartyTypePerson,
        DisplayName: "  Purchaser  ",
        Email:       &email,
        Phone:       &phone,
    })
    if err != nil {
        t.Fatal(err)
    }
    if party.Type != PartyTypePerson || party.DisplayName != "Purchaser" || party.Email == nil || *party.Email != "purchaser@example.test" || party.Phone == nil || *party.Phone != "+628100000000" {
        t.Fatalf("Party = %#v", party)
    }
}

func TestNewPartyRejectsInvalidInput(t *testing.T) {
    blank := "   "
    tests := []CreateInput{
        {Type: "UNKNOWN", DisplayName: "Party"},
        {Type: PartyTypePerson, DisplayName: "   "},
        {Type: PartyTypeOrganization, DisplayName: strings.Repeat("x", MaxDisplayNameLength+1)},
        {Type: PartyTypePerson, DisplayName: "Party", Email: &blank},
        {Type: PartyTypePerson, DisplayName: "Party", Phone: stringPointer(strings.Repeat("x", MaxPhoneLength+1))},
    }
    for _, input := range tests {
        if _, err := NewParty(input); !errors.Is(err, ErrInvalidInput) {
            t.Fatalf("NewParty(%#v) error = %v", input, err)
        }
    }
}

func stringPointer(value string) *string {
    return &value
}
