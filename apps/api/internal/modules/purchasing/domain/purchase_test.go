package domain

import (
    "errors"
    "strings"
    "testing"
)

func TestNewPurchaseStartsCommonPendingPaymentWithOrderedParticipants(t *testing.T) {
    purchaser := "purchaser-id"
    purchase, err := NewPurchase(CreateInput{
        EventID: "event-id", PurchaseRef: "purchase-ref", PurchaserPartyID: purchaser, PayerPartyID: purchaser,
        OfferingID: "offering-id", OfferingNameSnapshot: " Share ", OfferingKindSnapshot: " SHARE ",
        OfferingUnitPriceMinor: 100, ParticipantCapacitySnapshot: 2, TotalAmountMinor: 200, CurrencyCode: "IDR",
        Participants:    []PurchaseParticipantInput{{PartyID: &purchaser, DisplayName: " Siti "}, {DisplayName: " Ahmad "}},
        AccessTokenHash: make([]byte, 32),
    })
    if err != nil {
        t.Fatal(err)
    }
    if purchase.Channel != ChannelCommon || purchase.Status != StatusPendingPayment || purchase.Version != 1 || purchase.PayerPartyID == nil || *purchase.PayerPartyID != purchaser {
        t.Fatalf("purchase = %#v", purchase)
    }
    if len(purchase.Participants) != 2 || purchase.Participants[0].SequenceNo != 1 || purchase.Participants[0].DisplayName != "Siti" || purchase.Participants[1].SequenceNo != 2 || purchase.Participants[1].PartyID != nil || purchase.Participants[1].DisplayName != "Ahmad" {
        t.Fatalf("participants = %#v", purchase.Participants)
    }
}

func TestNewPurchaseRejectsInvalidInitialShape(t *testing.T) {
    valid := CreateInput{
        EventID: "event-id", PurchaseRef: "purchase-ref", PurchaserPartyID: "purchaser-id", PayerPartyID: "payer-id",
        OfferingID: "offering-id", OfferingNameSnapshot: "Share", OfferingKindSnapshot: "SHARE",
        OfferingUnitPriceMinor: 100, ParticipantCapacitySnapshot: 1, TotalAmountMinor: 100, CurrencyCode: "IDR",
        Participants: []PurchaseParticipantInput{{DisplayName: "Siti"}}, AccessTokenHash: make([]byte, 32),
    }
    cases := []CreateInput{
        func() CreateInput { value := valid; value.PayerPartyID = ""; return value }(),
        func() CreateInput { value := valid; value.Participants = nil; return value }(),
        func() CreateInput { value := valid; value.ParticipantCapacitySnapshot = 0; return value }(),
        func() CreateInput {
            value := valid
            value.ParticipantCapacitySnapshot = MaxParticipantCapacity + 1
            return value
        }(),
        func() CreateInput { value := valid; value.TotalAmountMinor = 99; return value }(),
        func() CreateInput { value := valid; value.CurrencyCode = "idr"; return value }(),
        func() CreateInput { value := valid; value.AccessTokenHash = nil; return value }(),
        func() CreateInput {
            value := valid
            value.OfferingNameSnapshot = strings.Repeat("x", MaxParticipantDisplayLength+1)
            return value
        }(),
    }
    for _, input := range cases {
        if _, err := NewPurchase(input); !errors.Is(err, ErrInvalidInput) {
            t.Fatalf("NewPurchase(%#v) error = %v", input, err)
        }
    }
}

func TestNewPurchaseRejectsOverflowingParticipantTotal(t *testing.T) {
    input := CreateInput{
        EventID: "event-id", PurchaseRef: "purchase-ref", PurchaserPartyID: "purchaser-id", PayerPartyID: "payer-id",
        OfferingID: "offering-id", OfferingNameSnapshot: "Share", OfferingKindSnapshot: "SHARE",
        OfferingUnitPriceMinor: MaxSafeInteger/2 + 1, ParticipantCapacitySnapshot: 2, TotalAmountMinor: 0, CurrencyCode: "IDR",
        Participants: []PurchaseParticipantInput{{DisplayName: "Siti"}, {DisplayName: "Ahmad"}}, AccessTokenHash: make([]byte, 32),
    }
    if _, err := NewPurchase(input); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("NewPurchase() error = %v, want invalid input", err)
    }
}
