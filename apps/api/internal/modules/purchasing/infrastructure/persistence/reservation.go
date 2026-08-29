package persistence

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"
    "time"

    eventdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/domain"
    eventpersistence "github.com/Kangditya/persona-apps/apps/api/internal/modules/event/infrastructure/persistence"
    offeringdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/domain"
    offeringpersistence "github.com/Kangditya/persona-apps/apps/api/internal/modules/offering/infrastructure/persistence"
    purchasingdomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/domain"
)

func LockCheckoutForOffering(ctx context.Context, transaction *sql.Tx, offeringID string, occurredAt time.Time) (purchasingdomain.CheckoutSource, error) {
    offeringID = strings.TrimSpace(offeringID)
    if transaction == nil || offeringID == "" || occurredAt.IsZero() {
        return purchasingdomain.CheckoutSource{}, purchasingdomain.ErrInvalidInput
    }
    var eventID string
    if err := transaction.QueryRowContext(ctx, `SELECT event_id FROM offerings WHERE id = $1`, offeringID).Scan(&eventID); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return purchasingdomain.CheckoutSource{}, purchasingdomain.ErrNotFound
        }
        return purchasingdomain.CheckoutSource{}, fmt.Errorf("lookup offering checkout parent: %w", translateError(err))
    }
    return LockCheckout(ctx, transaction, eventID, offeringID, occurredAt)
}

func LockCheckout(ctx context.Context, transaction *sql.Tx, eventID, offeringID string, occurredAt time.Time) (purchasingdomain.CheckoutSource, error) {
    eventID = strings.TrimSpace(eventID)
    offeringID = strings.TrimSpace(offeringID)
    if transaction == nil || eventID == "" || offeringID == "" || occurredAt.IsZero() {
        return purchasingdomain.CheckoutSource{}, purchasingdomain.ErrInvalidInput
    }

    // ponytail: per-Event lock serializes checkout; use a narrower atomic guard only after measured contention.
    lockedEvent, err := eventpersistence.NewRepository(transaction).GetForUpdate(ctx, eventID)
    if err != nil {
        return purchasingdomain.CheckoutSource{}, checkoutParentError(err)
    }
    lockedOffering, err := offeringpersistence.NewRepository(transaction).GetForUpdate(ctx, offeringID)
    if err != nil {
        return purchasingdomain.CheckoutSource{}, checkoutParentError(err)
    }
    source := purchasingdomain.CheckoutSource{Event: lockedEvent, Offering: lockedOffering, OccurredAt: occurredAt.UTC()}
    if err := purchasingdomain.ValidateCheckoutSource(source); err != nil {
        return purchasingdomain.CheckoutSource{}, err
    }
    return source, nil
}

func Reserve(ctx context.Context, transaction *sql.Tx, source purchasingdomain.CheckoutSource, purchase purchasingdomain.Purchase) (purchasingdomain.Reservation, error) {
    if transaction == nil {
        return purchasingdomain.Reservation{}, purchasingdomain.ErrInvalidInput
    }
    if err := purchasingdomain.ValidateCheckoutSource(source); err != nil {
        return purchasingdomain.Reservation{}, err
    }
    if err := purchasingdomain.ValidateReservationPurchase(source, purchase); err != nil {
        return purchasingdomain.Reservation{}, err
    }

    eventUsed, offeringUsed, err := reservationUsage(ctx, transaction, source.Event.ID, source.Offering.ID)
    if err != nil {
        return purchasingdomain.Reservation{}, err
    }
    available, err := offeringdomain.AvailableParticipantUnits(offeringdomain.AvailabilityInput{
        EventQuota:    source.Event.ParticipantQuota,
        OfferingQuota: source.Offering.ParticipantQuota,
        EventUsage:    offeringdomain.ReservationUsage{Reserved: eventUsed},
        OfferingUsage: offeringdomain.ReservationUsage{Reserved: offeringUsed},
    })
    if err != nil {
        return purchasingdomain.Reservation{}, fmt.Errorf("calculate checkout quota availability: %w", err)
    }
    if available != nil && int64(purchase.ParticipantCount) > *available {
        return purchasingdomain.Reservation{}, purchasingdomain.ErrQuotaUnavailable
    }

    expiresAt := source.OccurredAt.UTC().Add(purchasingdomain.ReservationHold)
    reservation, err := scanReservation(transaction.QueryRowContext(ctx, `
INSERT INTO quota_reservations (event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at)
VALUES ($1, $2, $3, 1, $4, $5, $6)
RETURNING id, event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at, version, created_at, updated_at`,
        source.Event.ID, purchase.ID, source.Offering.ID, purchase.ParticipantCount, purchasingdomain.ReservationStatusReserved, expiresAt))
    if err != nil {
        return purchasingdomain.Reservation{}, fmt.Errorf("insert quota reservation: %w", translateError(err))
    }
    return reservation, nil
}

func reservationUsage(ctx context.Context, transaction *sql.Tx, eventID, offeringID string) (int64, int64, error) {
    var eventUsed, offeringUsed int64
    if err := transaction.QueryRowContext(ctx, `
SELECT COALESCE(SUM(participant_units), 0), COALESCE(SUM(participant_units) FILTER (WHERE offering_id = $2), 0)
FROM quota_reservations
WHERE event_id = $1 AND status IN ($3, $4)`, eventID, offeringID, purchasingdomain.ReservationStatusReserved, purchasingdomain.ReservationStatusConsumed).Scan(&eventUsed, &offeringUsed); err != nil {
        return 0, 0, fmt.Errorf("get checkout reservation usage: %w", translateError(err))
    }
    return eventUsed, offeringUsed, nil
}

type reservationScanner interface {
    Scan(...any) error
}

type reservationRow struct {
    ID               string
    EventID          string
    PurchaseID       string
    OfferingID       string
    AttemptNo        int
    ParticipantUnits int
    Status           string
    ExpiresAt        sql.NullTime
    Version          int64
    CreatedAt        time.Time
    UpdatedAt        time.Time
}

func scanReservation(value reservationScanner) (purchasingdomain.Reservation, error) {
    var row reservationRow
    if err := value.Scan(row.destinations()...); err != nil {
        return purchasingdomain.Reservation{}, err
    }
    return row.reservation()
}

func (row *reservationRow) destinations() []any {
    return []any{
        &row.ID,
        &row.EventID,
        &row.PurchaseID,
        &row.OfferingID,
        &row.AttemptNo,
        &row.ParticipantUnits,
        &row.Status,
        &row.ExpiresAt,
        &row.Version,
        &row.CreatedAt,
        &row.UpdatedAt,
    }
}

func (row reservationRow) reservation() (purchasingdomain.Reservation, error) {
    if !row.ExpiresAt.Valid {
        return purchasingdomain.Reservation{}, errors.New("quota reservation expiry is required")
    }
    return purchasingdomain.Reservation{
        ID:               row.ID,
        EventID:          row.EventID,
        PurchaseID:       row.PurchaseID,
        OfferingID:       row.OfferingID,
        AttemptNo:        row.AttemptNo,
        ParticipantUnits: row.ParticipantUnits,
        Status:           purchasingdomain.ReservationStatus(row.Status),
        ExpiresAt:        row.ExpiresAt.Time.UTC(),
        Version:          row.Version,
        CreatedAt:        row.CreatedAt.UTC(),
        UpdatedAt:        row.UpdatedAt.UTC(),
    }, nil
}

func checkoutParentError(err error) error {
    switch {
    case errors.Is(err, eventdomain.ErrNotFound), errors.Is(err, offeringdomain.ErrNotFound):
        return purchasingdomain.ErrNotFound
    case errors.Is(err, eventdomain.ErrInvalidInput), errors.Is(err, offeringdomain.ErrInvalidInput):
        return purchasingdomain.ErrInvalidInput
    default:
        return err
    }
}
