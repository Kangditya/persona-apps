package purchasing

import (
    "bytes"
    "errors"
    "strings"
    "testing"
)

func TestPublicPurchaseReferenceUsesApprovedFormat(t *testing.T) {
    reference, err := newPublicPurchaseReference(2026, bytes.NewReader(make([]byte, publicPurchaseReferenceBytes)))
    if err != nil {
        t.Fatal(err)
    }
    if reference != "QRB-2026-AAAAAAAAAAAAAAAA" {
        t.Fatalf("reference = %q", reference)
    }
    if _, err := newPublicPurchaseReference(1899, bytes.NewReader(nil)); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("invalid year error = %v", err)
    }
}

func TestPublicPurchaseRelationshipsKeepExplicitRoles(t *testing.T) {
    request := publicCreatePurchaseRequest{
        Purchaser: publicPartyDeclarationRequest{
            PartyRef: field("purchaser"), DisplayName: field(" Siti Aminah "), Email: field(" siti@example.test "),
        },
        Payer: publicPartyDeclarationRequest{PartyRef: field("purchaser")},
        Participants: []publicParticipantRequest{
            {PartyRef: field("purchaser")},
            {DisplayName: field(" Ahmad ")},
        },
    }
    relationships, err := request.relationships()
    if err != nil {
        t.Fatal(err)
    }
    if len(relationships.DeclaredParties) != 1 || relationships.PurchaserRef != "purchaser" || relationships.PayerRef != "purchaser" || relationships.DeclaredParties[0].Party.DisplayName != "Siti Aminah" || relationships.DeclaredParties[0].Party.Email == nil || *relationships.DeclaredParties[0].Party.Email != "siti@example.test" || relationships.Participants[1].DisplayName != "Ahmad" {
        t.Fatalf("relationships = %#v", relationships)
    }

    request.Participants[0].DisplayName = field("duplicate shape")
    if _, err := request.relationships(); !errors.Is(err, ErrInvalidRelationships) {
        t.Fatalf("mixed participant shape error = %v", err)
    }
}

func TestPublicPurchaseAccessTokenIsOpaqueAndHashed(t *testing.T) {
    token, hash, err := newPublicPurchaseAccessToken(bytes.NewReader(make([]byte, publicPurchaseAccessTokenBytes)))
    if err != nil {
        t.Fatal(err)
    }
    if len(token) != 43 || strings.ContainsAny(token, "+/=") || len(hash) != 32 || bytes.Equal([]byte(token), hash) {
        t.Fatalf("token/hash = %q/%x", token, hash)
    }
}

func field(value string) publicStringField {
    return publicStringField{Value: value, Set: true}
}
