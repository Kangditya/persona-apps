package purchasing

import (
	"errors"
	"testing"
	"time"

	"github.com/Kangditya/persona-apps/apps/api/internal/event"
	"github.com/Kangditya/persona-apps/apps/api/internal/offering"
)

func TestValidateCheckoutSource(t *testing.T) {
	now := time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)
	source := CheckoutSource{
		Event:    event.Event{ID: "event-1", Status: event.StatusActive},
		Offering: offering.Offering{ID: "offering-1", EventID: "event-1", Status: offering.StatusPublished, ParticipantCapacity: 1},
		OccurredAt: now,
	}
	if err := validateCheckoutSource(source); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		mutate func(*CheckoutSource)
		want   error
	}{
		{"event mismatch", func(value *CheckoutSource) { value.Offering.EventID = "other" }, ErrInvalidInput},
		{"inactive event", func(value *CheckoutSource) { value.Event.Status = event.StatusSuspended }, ErrStateConflict},
		{"unpublished offering", func(value *CheckoutSource) { value.Offering.Status = offering.StatusDraft }, ErrStateConflict},
		{"before opening", func(value *CheckoutSource) { opening := now.Add(time.Second); value.Event.RegistrationOpensAt = &opening }, ErrStateConflict},
		{"at closing", func(value *CheckoutSource) { closing := now; value.Event.RegistrationClosesAt = &closing }, ErrStateConflict},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			value := source
			testCase.mutate(&value)
			if err := validateCheckoutSource(value); !errors.Is(err, testCase.want) {
				t.Fatalf("validateCheckoutSource() error = %v, want %v", err, testCase.want)
			}
		})
	}
}

func TestValidateReservationPurchase(t *testing.T) {
	source := CheckoutSource{
		Event:    event.Event{ID: "event-1"},
		Offering: offering.Offering{ID: "offering-1", EventID: "event-1", ParticipantCapacity: 2},
	}
	if err := validateReservationPurchase(source, Purchase{ID: "purchase-1", EventID: "event-1", OfferingID: "offering-1", ParticipantCount: 2}); err != nil {
		t.Fatal(err)
	}
	for _, purchase := range []Purchase{
		{ID: "purchase-1", EventID: "other", OfferingID: "offering-1", ParticipantCount: 1},
		{ID: "purchase-1", EventID: "event-1", OfferingID: "other", ParticipantCount: 1},
		{ID: "purchase-1", EventID: "event-1", OfferingID: "offering-1", ParticipantCount: 0},
		{ID: "purchase-1", EventID: "event-1", OfferingID: "offering-1", ParticipantCount: 3},
	} {
		if err := validateReservationPurchase(source, purchase); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("validateReservationPurchase(%#v) error = %v, want invalid input", purchase, err)
		}
	}
}
