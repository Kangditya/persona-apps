package domain

import (
    "errors"
    "fmt"
    "regexp"

    identitydomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/domain"
)

const maxPartyReferenceLength = 64

var (
    ErrInvalidRelationships = errors.New("invalid purchase relationships")
    partyReferencePattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
)

type PartyDeclarationInput struct {
    Ref   string
    Party identitydomain.CreateInput
}

type PayerInput struct {
    Declaration *PartyDeclarationInput
    Ref         string
}

type ParticipantInput struct {
    PartyRef    string
    DisplayName string
}

type RelationshipsInput struct {
    Purchaser    PartyDeclarationInput
    Payer        PayerInput
    Participants []ParticipantInput
}

type DeclaredParty struct {
    Ref   string
    Party identitydomain.Party
}

type IntendedParticipant struct {
    PartyRef    string
    DisplayName string
}

type Relationships struct {
    DeclaredParties []DeclaredParty
    PurchaserRef    string
    PayerRef        string
    Participants    []IntendedParticipant
}

func NewRelationships(input RelationshipsInput) (Relationships, error) {
    purchaser, err := newDeclaredParty(input.Purchaser)
    if err != nil {
        return Relationships{}, err
    }
    parties := []DeclaredParty{purchaser}
    declared := map[string]identitydomain.Party{purchaser.Ref: purchaser.Party}

    payerRef := input.Payer.Ref
    if input.Payer.Declaration != nil {
        if input.Payer.Ref != "" {
            return Relationships{}, fmt.Errorf("%w: payer cannot be both a declaration and reference", ErrInvalidRelationships)
        }
        payer, payerErr := newDeclaredParty(*input.Payer.Declaration)
        if payerErr != nil {
            return Relationships{}, payerErr
        }
        if _, exists := declared[payer.Ref]; exists {
            return Relationships{}, fmt.Errorf("%w: duplicate party reference", ErrInvalidRelationships)
        }
        parties = append(parties, payer)
        declared[payer.Ref] = payer.Party
        payerRef = payer.Ref
    } else {
        if err := validatePartyReference(payerRef); err != nil {
            return Relationships{}, err
        }
        if _, exists := declared[payerRef]; !exists {
            return Relationships{}, fmt.Errorf("%w: payer references an undeclared party", ErrInvalidRelationships)
        }
    }

    if len(input.Participants) == 0 {
        return Relationships{}, fmt.Errorf("%w: at least one participant is required", ErrInvalidRelationships)
    }
    participants := make([]IntendedParticipant, 0, len(input.Participants))
    for _, inputParticipant := range input.Participants {
        referenceSet := inputParticipant.PartyRef != ""
        nameSet := inputParticipant.DisplayName != ""
        if referenceSet == nameSet {
            return Relationships{}, fmt.Errorf("%w: participant must be exactly one of a party reference or display name", ErrInvalidRelationships)
        }
        if referenceSet {
            if err := validatePartyReference(inputParticipant.PartyRef); err != nil {
                return Relationships{}, err
            }
            party, exists := declared[inputParticipant.PartyRef]
            if !exists {
                return Relationships{}, fmt.Errorf("%w: participant references an undeclared party", ErrInvalidRelationships)
            }
            participants = append(participants, IntendedParticipant{PartyRef: inputParticipant.PartyRef, DisplayName: party.DisplayName})
            continue
        }
        nameOnlyParty, nameErr := identitydomain.NewParty(identitydomain.CreateInput{Type: identitydomain.PartyTypePerson, DisplayName: inputParticipant.DisplayName})
        if nameErr != nil {
            return Relationships{}, fmt.Errorf("%w: participant display name: %w", ErrInvalidRelationships, nameErr)
        }
        participants = append(participants, IntendedParticipant{DisplayName: nameOnlyParty.DisplayName})
    }

    return Relationships{
        DeclaredParties: parties,
        PurchaserRef:    purchaser.Ref,
        PayerRef:        payerRef,
        Participants:    participants,
    }, nil
}

func newDeclaredParty(input PartyDeclarationInput) (DeclaredParty, error) {
    if err := validatePartyReference(input.Ref); err != nil {
        return DeclaredParty{}, err
    }
    party, err := identitydomain.NewParty(input.Party)
    if err != nil {
        return DeclaredParty{}, fmt.Errorf("%w: party declaration: %w", ErrInvalidRelationships, err)
    }
    return DeclaredParty{Ref: input.Ref, Party: party}, nil
}

func validatePartyReference(value string) error {
    if len(value) > maxPartyReferenceLength || !partyReferencePattern.MatchString(value) {
        return fmt.Errorf("%w: party reference must match the approved request pattern", ErrInvalidRelationships)
    }
    return nil
}
