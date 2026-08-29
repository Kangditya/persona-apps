package domain

import (
    "errors"
    "testing"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
)

func TestNewSnapshotPurchaseCapturesLockedOffering(t *testing.T) {
    now := time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)
    source := CheckoutSource{
        Event: eventdomain.Event{ID: "event-id", Status: eventdomain.StatusActive},
        Offering: offeringdomain.Offering{
            ID: "offering-id", EventID: "event-id", Name: " Two shares ", Kind: " SHARE ", PriceMinor: 100,
            CurrencyCode: "IDR", ParticipantCapacity: 2, Status: offeringdomain.StatusPublished,
        },
        OccurredAt: now,
    }
    purchase, err := NewSnapshotPurchase(SnapshotPurchaseInput{
        Source: source, PurchaseRef: "purchase-ref", PurchaserPartyID: "purchaser-id", PayerPartyID: "payer-id",
        Participants: []PurchaseParticipantInput{{DisplayName: "Siti"}, {DisplayName: "Ahmad"}}, AccessTokenHash: make([]byte, 32),
    })
    if err != nil {
        t.Fatal(err)
    }
    if purchase.EventID != "event-id" || purchase.OfferingID != "offering-id" || purchase.OfferingNameSnapshot != "Two shares" || purchase.OfferingKindSnapshot != "SHARE" || purchase.OfferingUnitPriceMinor != 100 || purchase.ParticipantCapacitySnapshot != 2 || purchase.ParticipantCount != 2 || purchase.TotalAmountMinor != 200 || purchase.CurrencyCode != "IDR" {
        t.Fatalf("snapshot Purchase = %#v", purchase)
    }

    source.Offering.EventID = "other-event"
    if _, err := NewSnapshotPurchase(SnapshotPurchaseInput{Source: source, PurchaseRef: "purchase-ref", PurchaserPartyID: "purchaser-id", PayerPartyID: "payer-id", Participants: []PurchaseParticipantInput{{DisplayName: "Siti"}}, AccessTokenHash: make([]byte, 32)}); !errors.Is(err, ErrInvalidInput) {
        t.Fatalf("mismatched source error = %v, want invalid input", err)
    }
}
