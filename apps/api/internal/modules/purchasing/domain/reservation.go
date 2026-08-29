package domain

import (
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
)

const ReservationHold = 24 * time.Hour

type ReservationStatus string

const (
    ReservationStatusReserved ReservationStatus = "RESERVED"
    ReservationStatusConsumed ReservationStatus = "CONSUMED"
    ReservationStatusReleased ReservationStatus = "RELEASED"
    ReservationStatusExpired  ReservationStatus = "EXPIRED"
)

type CheckoutSource struct {
    Event      eventdomain.Event
    Offering   offeringdomain.Offering
    OccurredAt time.Time
}

type Reservation struct {
    ID               string
    EventID          string
    PurchaseID       string
    OfferingID       string
    AttemptNo        int
    ParticipantUnits int
    Status           ReservationStatus
    ExpiresAt        time.Time
    Version          int64
    CreatedAt        time.Time
    UpdatedAt        time.Time
}

func ValidateCheckoutSource(source CheckoutSource) error {
    if source.Event.ID == "" || source.Offering.ID == "" || source.Offering.EventID != source.Event.ID || source.OccurredAt.IsZero() {
        return ErrInvalidInput
    }
    if source.Event.Status != eventdomain.StatusActive || source.Offering.Status != offeringdomain.StatusPublished {
        return ErrStateConflict
    }
    now := source.OccurredAt.UTC()
    if source.Event.RegistrationOpensAt != nil && now.Before(source.Event.RegistrationOpensAt.UTC()) {
        return ErrStateConflict
    }
    if source.Event.RegistrationClosesAt != nil && !now.Before(source.Event.RegistrationClosesAt.UTC()) {
        return ErrStateConflict
    }
    return nil
}

func validateCheckoutSource(source CheckoutSource) error {
    return ValidateCheckoutSource(source)
}

func ValidateReservationPurchase(source CheckoutSource, purchase Purchase) error {
    if purchase.ID == "" || purchase.EventID != source.Event.ID || purchase.OfferingID != source.Offering.ID || purchase.ParticipantCount < 1 || int64(purchase.ParticipantCount) > int64(source.Offering.ParticipantCapacity) {
        return ErrInvalidInput
    }
    return nil
}

func validateReservationPurchase(source CheckoutSource, purchase Purchase) error {
    return ValidateReservationPurchase(source, purchase)
}
