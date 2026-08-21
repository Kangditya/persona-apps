package purchasing

import (
    "errors"
    "testing"

    "github.com/Kangditya/persona-apps/apps/api/internal/identity"
)

func TestNewRelationshipsReusesExplicitPartyReferences(t *testing.T) {
    relationships, err := NewRelationships(RelationshipsInput{
        Purchaser: declaration("purchaser", "  Siti Aminah  "),
        Payer:     PayerInput{Ref: "purchaser"},
        Participants: []ParticipantInput{
            {PartyRef: "purchaser"},
            {DisplayName: "  Ahmad  "},
        },
    })
    if err != nil {
        t.Fatal(err)
    }
    if len(relationships.DeclaredParties) != 1 || relationships.PurchaserRef != "purchaser" || relationships.PayerRef != "purchaser" {
        t.Fatalf("relationships = %#v", relationships)
    }
    if relationships.Participants[0].PartyRef != "purchaser" || relationships.Participants[0].DisplayName != "Siti Aminah" {
        t.Fatalf("resolved participant = %#v", relationships.Participants[0])
    }
    if relationships.Participants[1].PartyRef != "" || relationships.Participants[1].DisplayName != "Ahmad" {
        t.Fatalf("unresolved participant = %#v", relationships.Participants[1])
    }
}

func TestNewRelationshipsKeepsDistinctDeclarationsAndOrder(t *testing.T) {
    payer := declaration("payer", "Budi Santoso")
    relationships, err := NewRelationships(RelationshipsInput{
        Purchaser: declaration("purchaser", "Siti Aminah"),
        Payer:     PayerInput{Declaration: &payer},
        Participants: []ParticipantInput{
            {PartyRef: "payer"},
            {PartyRef: "purchaser"},
            {DisplayName: "Ahmad"},
        },
    })
    if err != nil {
        t.Fatal(err)
    }
    if len(relationships.DeclaredParties) != 2 || relationships.PayerRef != "payer" {
        t.Fatalf("relationships = %#v", relationships)
    }
    want := []IntendedParticipant{
        {PartyRef: "payer", DisplayName: "Budi Santoso"},
        {PartyRef: "purchaser", DisplayName: "Siti Aminah"},
        {DisplayName: "Ahmad"},
    }
    for index := range want {
        if relationships.Participants[index] != want[index] {
            t.Fatalf("participant %d = %#v, want %#v", index, relationships.Participants[index], want[index])
        }
    }
}

func TestNewRelationshipsRejectsAmbiguousOrUnknownInputs(t *testing.T) {
    payer := declaration("purchaser", "Budi Santoso")
    tests := []RelationshipsInput{
        {Purchaser: declaration("purchaser", "Siti Aminah"), Payer: PayerInput{}, Participants: []ParticipantInput{{DisplayName: "Ahmad"}}},
        {Purchaser: declaration("purchaser", "Siti Aminah"), Payer: PayerInput{Ref: "unknown"}, Participants: []ParticipantInput{{DisplayName: "Ahmad"}}},
        {Purchaser: declaration("purchaser", "Siti Aminah"), Payer: PayerInput{Declaration: &payer}, Participants: []ParticipantInput{{DisplayName: "Ahmad"}}},
        {Purchaser: declaration("purchaser", "Siti Aminah"), Payer: PayerInput{Ref: "purchaser"}, Participants: []ParticipantInput{{PartyRef: "purchaser", DisplayName: "Ahmad"}}},
        {Purchaser: declaration("purchaser", "Siti Aminah"), Payer: PayerInput{Ref: "purchaser"}, Participants: []ParticipantInput{{PartyRef: "unknown"}}},
        {Purchaser: declaration("invalid ref", "Siti Aminah"), Payer: PayerInput{Ref: "purchaser"}, Participants: []ParticipantInput{{DisplayName: "Ahmad"}}},
    }
    for _, input := range tests {
        if _, err := NewRelationships(input); !errors.Is(err, ErrInvalidRelationships) {
            t.Fatalf("NewRelationships(%#v) error = %v", input, err)
        }
    }
}

func declaration(ref, displayName string) PartyDeclarationInput {
    return PartyDeclarationInput{
        Ref: ref,
        Party: identity.CreateInput{
            Type:        identity.PartyTypePerson,
            DisplayName: displayName,
        },
    }
}
