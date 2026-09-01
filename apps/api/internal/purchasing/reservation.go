package purchasing

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/event"
    "github.com/Kangditya/persona-apps/apps/api/internal/offering"
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
    Event      event.Event
    Offering   offering.Offering
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

// LockCheckoutForOffering discovers the immutable parent reference before
// taking the authoritative Event then Offering locks used by checkout.
func LockCheckoutForOffering(ctx context.Context, transaction *sql.Tx, offeringID string, occurredAt time.Time) (CheckoutSource, error) {
    offeringID = strings.TrimSpace(offeringID)
    if transaction == nil || offeringID == "" || occurredAt.IsZero() {
        return CheckoutSource{}, ErrInvalidInput
    }
    var eventID string
    if err := transaction.QueryRowContext(ctx, `SELECT event_id FROM offerings WHERE id = $1`, offeringID).Scan(&eventID); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return CheckoutSource{}, ErrNotFound
        }
        return CheckoutSource{}, fmt.Errorf("lookup offering checkout parent: %w", translateError(err))
    }
    return LockCheckout(ctx, transaction, eventID, offeringID, occurredAt)
}

// LockCheckout obtains the authoritative Event and Offering rows in the only
// allowed order for a checkout transaction.
func LockCheckout(ctx context.Context, transaction *sql.Tx, eventID, offeringID string, occurredAt time.Time) (CheckoutSource, error) {
    eventID = strings.TrimSpace(eventID)
    offeringID = strings.TrimSpace(offeringID)
    if transaction == nil || eventID == "" || offeringID == "" || occurredAt.IsZero() {
        return CheckoutSource{}, ErrInvalidInput
    }

    // ponytail: per-Event lock serializes checkout; use a narrower atomic guard only after measured contention.
    lockedEvent, err := event.NewRepository(transaction).GetForUpdate(ctx, eventID)
    if err != nil {
        return CheckoutSource{}, checkoutParentError(err)
    }
    lockedOffering, err := offering.NewRepository(transaction).GetForUpdate(ctx, offeringID)
    if err != nil {
        return CheckoutSource{}, checkoutParentError(err)
    }
    source := CheckoutSource{Event: lockedEvent, Offering: lockedOffering, OccurredAt: occurredAt.UTC()}
    if err := validateCheckoutSource(source); err != nil {
        return CheckoutSource{}, err
    }
    return source, nil
}

// Reserve creates attempt one after LockCheckout has established the source
// state and retained its row locks in the caller-owned transaction.
func Reserve(ctx context.Context, transaction *sql.Tx, source CheckoutSource, purchase Purchase) (Reservation, error) {
    if transaction == nil {
        return Reservation{}, ErrInvalidInput
    }
    if err := validateCheckoutSource(source); err != nil {
        return Reservation{}, err
    }
    if err := validateReservationPurchase(source, purchase); err != nil {
        return Reservation{}, err
    }

    eventUsed, offeringUsed, err := reservationUsage(ctx, transaction, source.Event.ID, source.Offering.ID)
    if err != nil {
        return Reservation{}, err
    }
    available, err := offering.AvailableParticipantUnits(offering.AvailabilityInput{
        EventQuota:    source.Event.ParticipantQuota,
        OfferingQuota: source.Offering.ParticipantQuota,
        EventUsage:    offering.ReservationUsage{Reserved: eventUsed},
        OfferingUsage: offering.ReservationUsage{Reserved: offeringUsed},
    })
    if err != nil {
        return Reservation{}, fmt.Errorf("calculate checkout quota availability: %w", err)
    }
    if available != nil && int64(purchase.ParticipantCount) > *available {
        return Reservation{}, ErrQuotaUnavailable
    }

    expiresAt := source.OccurredAt.UTC().Add(ReservationHold)
    reservation, err := scanReservation(transaction.QueryRowContext(ctx, `
INSERT INTO quota_reservations (event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at)
VALUES ($1, $2, $3, 1, $4, $5, $6)
RETURNING id, event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at, version, created_at, updated_at`,
        source.Event.ID, purchase.ID, source.Offering.ID, purchase.ParticipantCount, ReservationStatusReserved, expiresAt))
    if err != nil {
        return Reservation{}, fmt.Errorf("insert quota reservation: %w", translateError(err))
    }
    return reservation, nil
}

func validateCheckoutSource(source CheckoutSource) error {
    if source.Event.ID == "" || source.Offering.ID == "" || source.Offering.EventID != source.Event.ID || source.OccurredAt.IsZero() {
        return ErrInvalidInput
    }
    if source.Event.Status != event.StatusActive || source.Offering.Status != offering.StatusPublished {
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

func validateReservationPurchase(source CheckoutSource, purchase Purchase) error {
    if purchase.ID == "" || purchase.EventID != source.Event.ID || purchase.OfferingID != source.Offering.ID || purchase.ParticipantCount < 1 || int64(purchase.ParticipantCount) > int64(source.Offering.ParticipantCapacity) {
        return ErrInvalidInput
    }
    return nil
}

func reservationUsage(ctx context.Context, transaction *sql.Tx, eventID, offeringID string) (int64, int64, error) {
    var eventUsed, offeringUsed int64
    if err := transaction.QueryRowContext(ctx, `
SELECT COALESCE(SUM(participant_units), 0), COALESCE(SUM(participant_units) FILTER (WHERE offering_id = $2), 0)
FROM quota_reservations
WHERE event_id = $1 AND status IN ($3, $4)`, eventID, offeringID, ReservationStatusReserved, ReservationStatusConsumed).Scan(&eventUsed, &offeringUsed); err != nil {
        return 0, 0, fmt.Errorf("get checkout reservation usage: %w", translateError(err))
    }
    return eventUsed, offeringUsed, nil
}

func scanReservation(value scanner) (Reservation, error) {
    var reservation Reservation
    var status string
    if err := value.Scan(
        &reservation.ID, &reservation.EventID, &reservation.PurchaseID, &reservation.OfferingID,
        &reservation.AttemptNo, &reservation.ParticipantUnits, &status, &reservation.ExpiresAt,
        &reservation.Version, &reservation.CreatedAt, &reservation.UpdatedAt,
    ); err != nil {
        return Reservation{}, err
    }
    reservation.Status = ReservationStatus(status)
    reservation.ExpiresAt = reservation.ExpiresAt.UTC()
    reservation.CreatedAt = reservation.CreatedAt.UTC()
    reservation.UpdatedAt = reservation.UpdatedAt.UTC()
    return reservation, nil
}

func checkoutParentError(err error) error {
    switch {
    case errors.Is(err, event.ErrNotFound), errors.Is(err, offering.ErrNotFound):
        return ErrNotFound
    case errors.Is(err, event.ErrInvalidInput), errors.Is(err, offering.ErrInvalidInput):
        return ErrInvalidInput
    default:
        return err
    }
}
