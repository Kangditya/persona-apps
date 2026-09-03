package purchasing

import (
    "context"
    "crypto/sha256"
    "crypto/subtle"
    "database/sql"
    "encoding/base64"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/event"
    "github.com/Kangditya/persona-apps/apps/api/internal/offering"
)

type PaymentReservation struct {
    ID               string
    EventID          string
    PurchaseID       string
    OfferingID       string
    AttemptNo        int
    ParticipantUnits int
    Status           ReservationStatus
    ExpiresAt        *time.Time
    Version          int64
    CreatedAt        time.Time
    UpdatedAt        time.Time
}

type PaymentSubmissionSource struct {
    Purchase    Purchase
    Reservation PaymentReservation
}

func AuthorizePurchaseToken(ctx context.Context, database DBTX, purchaseID, token string) error {
    purchaseID = strings.TrimSpace(purchaseID)
    decoded, err := base64.RawURLEncoding.DecodeString(token)
    if database == nil || purchaseID == "" || err != nil || len(decoded) != 32 || base64.RawURLEncoding.EncodeToString(decoded) != token {
        return ErrUnauthenticated
    }
    var stored []byte
    if err := database.QueryRowContext(ctx, `SELECT access_token_hash FROM purchases WHERE id = $1`, purchaseID).Scan(&stored); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ErrNotFound
        }
        return fmt.Errorf("read purchase access token: %w", translateError(err))
    }
    digest := sha256.Sum256([]byte(token))
    if len(stored) != len(digest) || subtle.ConstantTimeCompare(stored, digest[:]) != 1 {
        return ErrForbidden
    }
    return nil
}

func PreparePaymentSubmission(ctx context.Context, transaction *sql.Tx, purchaseID string, occurredAt time.Time) (PaymentSubmissionSource, error) {
    purchaseID = strings.TrimSpace(purchaseID)
    if transaction == nil || purchaseID == "" || occurredAt.IsZero() {
        return PaymentSubmissionSource{}, ErrInvalidInput
    }
    occurredAt = occurredAt.UTC()

    var eventID, offeringID string
    if err := transaction.QueryRowContext(ctx, `SELECT event_id, offering_id FROM purchases WHERE id = $1`, purchaseID).Scan(&eventID, &offeringID); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return PaymentSubmissionSource{}, ErrNotFound
        }
        return PaymentSubmissionSource{}, fmt.Errorf("read payment submission context: %w", translateError(err))
    }

    lockedEvent, err := event.NewRepository(transaction).GetForUpdate(ctx, eventID)
    if err != nil {
        return PaymentSubmissionSource{}, checkoutParentError(err)
    }
    lockedOffering, err := offering.NewRepository(transaction).GetForUpdate(ctx, offeringID)
    if err != nil {
        return PaymentSubmissionSource{}, checkoutParentError(err)
    }

    purchase, err := scanPurchase(transaction.QueryRowContext(ctx, `SELECT `+purchaseColumns+` FROM purchases WHERE id = $1 FOR UPDATE`, purchaseID))
    if errors.Is(err, sql.ErrNoRows) {
        return PaymentSubmissionSource{}, ErrNotFound
    }
    if err != nil {
        return PaymentSubmissionSource{}, fmt.Errorf("lock payment purchase: %w", translateError(err))
    }
    if purchase.EventID != lockedEvent.ID || purchase.OfferingID != lockedOffering.ID || purchase.Channel != ChannelCommon || purchase.Status != StatusPendingPayment || purchase.PayerPartyID == nil || purchase.TotalAmountMinor <= 0 {
        return PaymentSubmissionSource{}, ErrStateConflict
    }

    reservation, err := scanPaymentReservation(transaction.QueryRowContext(ctx, `
SELECT id, event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at, version, created_at, updated_at
FROM quota_reservations
WHERE purchase_id = $1
ORDER BY attempt_no DESC
LIMIT 1
FOR UPDATE`, purchaseID))
    if errors.Is(err, sql.ErrNoRows) {
        return PaymentSubmissionSource{}, ErrStateConflict
    }
    if err != nil {
        return PaymentSubmissionSource{}, fmt.Errorf("lock payment reservation: %w", translateError(err))
    }
    if reservation.EventID != purchase.EventID || reservation.OfferingID != purchase.OfferingID || reservation.ParticipantUnits != purchase.ParticipantCount {
        return PaymentSubmissionSource{}, ErrStateConflict
    }

    switch reservation.Status {
    case ReservationStatusReserved:
        switch {
        case reservation.ExpiresAt == nil:
            return PaymentSubmissionSource{}, ErrStateConflict
        case occurredAt.Before(*reservation.ExpiresAt):
            reservation, err = pausePaymentReservation(ctx, transaction, reservation)
        default:
            reservation, err = expireAndReacquirePaymentReservation(ctx, transaction, CheckoutSource{Event: lockedEvent, Offering: lockedOffering, OccurredAt: occurredAt}, purchase, reservation)
        }
    case ReservationStatusReleased, ReservationStatusExpired:
        reservation, err = reacquirePaymentReservation(ctx, transaction, CheckoutSource{Event: lockedEvent, Offering: lockedOffering, OccurredAt: occurredAt}, purchase, reservation.AttemptNo+1)
    default:
        return PaymentSubmissionSource{}, ErrStateConflict
    }
    if err != nil {
        return PaymentSubmissionSource{}, err
    }
    return PaymentSubmissionSource{Purchase: purchase, Reservation: reservation}, nil
}

func pausePaymentReservation(ctx context.Context, transaction *sql.Tx, reservation PaymentReservation) (PaymentReservation, error) {
    paused, err := scanPaymentReservation(transaction.QueryRowContext(ctx, `
UPDATE quota_reservations
SET expires_at = NULL, version = version + 1, updated_at = now()
WHERE id = $1 AND status = $2 AND version = $3 AND expires_at IS NOT NULL
RETURNING id, event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at, version, created_at, updated_at`, reservation.ID, ReservationStatusReserved, reservation.Version))
    if errors.Is(err, sql.ErrNoRows) {
        return PaymentReservation{}, ErrStateConflict
    }
    if err != nil {
        return PaymentReservation{}, fmt.Errorf("pause payment reservation: %w", translateError(err))
    }
    return paused, nil
}

func expireAndReacquirePaymentReservation(ctx context.Context, transaction *sql.Tx, source CheckoutSource, purchase Purchase, reservation PaymentReservation) (PaymentReservation, error) {
    result, err := transaction.ExecContext(ctx, `
UPDATE quota_reservations
SET status = $1, released_at = $2, release_reason = 'PAYMENT_EVIDENCE_REACQUIRE', version = version + 1, updated_at = now()
WHERE id = $3 AND status = $4 AND version = $5 AND expires_at IS NOT NULL AND expires_at <= $2`, ReservationStatusExpired, source.OccurredAt, reservation.ID, ReservationStatusReserved, reservation.Version)
    if err != nil {
        return PaymentReservation{}, fmt.Errorf("expire payment reservation: %w", translateError(err))
    }
    changed, err := result.RowsAffected()
    if err != nil {
        return PaymentReservation{}, fmt.Errorf("read expired reservation result: %w", err)
    }
    if changed != 1 {
        return PaymentReservation{}, ErrStateConflict
    }
    return reacquirePaymentReservation(ctx, transaction, source, purchase, reservation.AttemptNo+1)
}

func reacquirePaymentReservation(ctx context.Context, transaction *sql.Tx, source CheckoutSource, purchase Purchase, attemptNo int) (PaymentReservation, error) {
    if attemptNo < 1 {
        return PaymentReservation{}, ErrInvalidInput
    }
    if err := validateCheckoutSource(source); err != nil {
        return PaymentReservation{}, err
    }
    if err := validateReservationPurchase(source, purchase); err != nil {
        return PaymentReservation{}, err
    }
    eventUsed, offeringUsed, err := reservationUsage(ctx, transaction, purchase.EventID, purchase.OfferingID)
    if err != nil {
        return PaymentReservation{}, err
    }
    available, err := offering.AvailableParticipantUnits(offering.AvailabilityInput{
        EventQuota: source.Event.ParticipantQuota, OfferingQuota: source.Offering.ParticipantQuota,
        EventUsage: offering.ReservationUsage{Reserved: eventUsed}, OfferingUsage: offering.ReservationUsage{Reserved: offeringUsed},
    })
    if err != nil {
        return PaymentReservation{}, fmt.Errorf("calculate payment quota availability: %w", err)
    }
    if available != nil && int64(purchase.ParticipantCount) > *available {
        return PaymentReservation{}, ErrQuotaUnavailable
    }
    reservation, err := scanPaymentReservation(transaction.QueryRowContext(ctx, `
INSERT INTO quota_reservations (event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, NULL)
RETURNING id, event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at, version, created_at, updated_at`, purchase.EventID, purchase.ID, purchase.OfferingID, attemptNo, purchase.ParticipantCount, ReservationStatusReserved))
    if err != nil {
        return PaymentReservation{}, fmt.Errorf("reacquire payment reservation: %w", translateError(err))
    }
    return reservation, nil
}

func scanPaymentReservation(value scanner) (PaymentReservation, error) {
    var reservation PaymentReservation
    var status string
    var expiresAt sql.NullTime
    if err := value.Scan(
        &reservation.ID, &reservation.EventID, &reservation.PurchaseID, &reservation.OfferingID,
        &reservation.AttemptNo, &reservation.ParticipantUnits, &status, &expiresAt,
        &reservation.Version, &reservation.CreatedAt, &reservation.UpdatedAt,
    ); err != nil {
        return PaymentReservation{}, err
    }
    reservation.Status = ReservationStatus(status)
    if expiresAt.Valid {
        value := expiresAt.Time.UTC()
        reservation.ExpiresAt = &value
    }
    reservation.CreatedAt = reservation.CreatedAt.UTC()
    reservation.UpdatedAt = reservation.UpdatedAt.UTC()
    return reservation, nil
}
