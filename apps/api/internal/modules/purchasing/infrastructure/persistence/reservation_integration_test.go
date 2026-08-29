package persistence

import (
    "context"
    "crypto/sha256"
    "database/sql"
    "errors"
    "fmt"
    "testing"
    "time"

    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

func TestReservePostgreSQL(t *testing.T) {
    database := integrationDatabase(t)
    ctx := context.Background()
    acquireActiveEventIntegrationLock(t, ctx, database)
    prefix := fmt.Sprintf("reservation-%d", time.Now().UnixNano())
    quota := int64(3)
    eventID, offeringID, partyID := createReservationFixture(t, ctx, database, prefix, &quota, &quota)
    t.Cleanup(func() { cleanupReservationFixture(t, ctx, database, eventID, partyID) })
    now := time.Now().UTC().Truncate(time.Microsecond)

    first, reservation, err := reserveCheckout(ctx, database, eventID, offeringID, partyID, prefix+"-first", now)
    if err != nil {
        t.Fatal(err)
    }
    if reservation.EventID != eventID || reservation.PurchaseID != first.ID || reservation.OfferingID != offeringID || reservation.AttemptNo != 1 || reservation.ParticipantUnits != 1 || reservation.Status != ReservationStatusReserved || !reservation.ExpiresAt.Equal(now.Add(ReservationHold)) {
        t.Fatalf("reservation = %#v", reservation)
    }
    rollbackPurchaseID, rollbackErr := reserveCheckoutThenRollback(ctx, database, eventID, offeringID, partyID, prefix+"-rollback", now)
    if rollbackErr == nil || rollbackPurchaseID == "" {
        t.Fatalf("reservation rollback error/id = %v/%q", rollbackErr, rollbackPurchaseID)
    }
    if _, err := NewRepository(database).Get(ctx, rollbackPurchaseID); !errors.Is(err, ErrNotFound) {
        t.Fatalf("rolled-back Purchase lookup error = %v", err)
    }
    var rollbackReservations int
    if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM quota_reservations WHERE purchase_id = $1`, rollbackPurchaseID).Scan(&rollbackReservations); err != nil {
        t.Fatal(err)
    }
    if rollbackReservations != 0 {
        t.Fatalf("rolled-back reservation count = %d, want 0", rollbackReservations)
    }

    _, secondReservation, err := reserveCheckout(ctx, database, eventID, offeringID, partyID, prefix+"-second", now)
    if err != nil {
        t.Fatal(err)
    }
    markReservationTerminal(t, ctx, database, secondReservation.ID, ReservationStatusConsumed)
    _, thirdReservation, err := reserveCheckout(ctx, database, eventID, offeringID, partyID, prefix+"-third", now)
    if err != nil {
        t.Fatal(err)
    }
    markReservationTerminal(t, ctx, database, thirdReservation.ID, ReservationStatusReleased)
    _, fourthReservation, err := reserveCheckout(ctx, database, eventID, offeringID, partyID, prefix+"-fourth", now)
    if err != nil {
        t.Fatal(err)
    }
    markReservationTerminal(t, ctx, database, fourthReservation.ID, ReservationStatusExpired)
    if _, _, err := reserveCheckout(ctx, database, eventID, offeringID, partyID, prefix+"-fifth", now); err != nil {
        t.Fatalf("terminal reservations should not consume quota: %v", err)
    }

    rolledBack, _, err := reserveCheckout(ctx, database, eventID, offeringID, partyID, prefix+"-over-quota", now)
    if !errors.Is(err, ErrQuotaUnavailable) {
        t.Fatalf("over-quota error = %v, want quota unavailable", err)
    }
    if _, err := NewRepository(database).Get(ctx, rolledBack.ID); !errors.Is(err, ErrNotFound) {
        t.Fatalf("over-quota Purchase was committed: %v", err)
    }

    if _, err := database.ExecContext(ctx, `UPDATE qurban_events SET participant_quota = NULL WHERE id = $1`, eventID); err != nil {
        t.Fatal(err)
    }
    var unboundedOfferingID string
    if err := database.QueryRowContext(ctx, `INSERT INTO offerings (event_id, code, name, offering_kind, price_minor, currency_code, participant_capacity, participant_quota, status) VALUES ($1, 'UNBOUNDED', $2, 'SHARE', 100, 'IDR', 1, NULL, 'PUBLISHED') RETURNING id`, eventID, prefix+" unbounded").Scan(&unboundedOfferingID); err != nil {
        t.Fatal(err)
    }
    if _, _, err := reserveCheckout(ctx, database, eventID, unboundedOfferingID, partyID, prefix+"-unbounded-a", now); err != nil {
        t.Fatal(err)
    }
    if _, _, err := reserveCheckout(ctx, database, eventID, unboundedOfferingID, partyID, prefix+"-unbounded-b", now); err != nil {
        t.Fatalf("nil Event and Offering quotas must not treat per-Purchase capacity as a global quota: %v", err)
    }
}

func TestReservePostgreSQLConcurrentLastUnit(t *testing.T) {
    database := integrationDatabase(t)
    ctx := context.Background()
    acquireActiveEventIntegrationLock(t, ctx, database)
    prefix := fmt.Sprintf("reservation-concurrent-%d", time.Now().UnixNano())
    quota := int64(1)
    eventID, offeringID, partyID := createReservationFixture(t, ctx, database, prefix, &quota, &quota)
    t.Cleanup(func() { cleanupReservationFixture(t, ctx, database, eventID, partyID) })

    commandContext, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    start := make(chan struct{})
    results := make(chan error, 2)
    for attempt := 0; attempt < 2; attempt++ {
        attempt := attempt
        go func() {
            <-start
            _, _, err := reserveCheckout(commandContext, database, eventID, offeringID, partyID, fmt.Sprintf("%s-%d", prefix, attempt), time.Now().UTC().Truncate(time.Microsecond))
            results <- err
        }()
    }
    close(start)

    successes := 0
    quotaFailures := 0
    for attempt := 0; attempt < 2; attempt++ {
        err := <-results
        switch {
        case err == nil:
            successes++
        case errors.Is(err, ErrQuotaUnavailable):
            quotaFailures++
        default:
            t.Fatalf("concurrent reservation error = %v", err)
        }
    }
    if successes != 1 || quotaFailures != 1 {
        t.Fatalf("concurrent successes/failures = %d/%d, want 1/1", successes, quotaFailures)
    }
    var reserved int64
    if err := database.QueryRowContext(ctx, `SELECT COALESCE(SUM(participant_units), 0) FROM quota_reservations WHERE event_id = $1 AND offering_id = $2 AND status = 'RESERVED'`, eventID, offeringID).Scan(&reserved); err != nil {
        t.Fatal(err)
    }
    if reserved != 1 {
        t.Fatalf("reserved units = %d, want 1", reserved)
    }
}

func reserveCheckout(ctx context.Context, database *sql.DB, eventID, offeringID, partyID, reference string, occurredAt time.Time) (Purchase, Reservation, error) {
    var created Purchase
    var reservation Reservation
    err := platformdb.Within(ctx, database, func(transaction *sql.Tx) error {
        source, err := LockCheckout(ctx, transaction, eventID, offeringID, occurredAt)
        if err != nil {
            return err
        }
        hash := sha256.Sum256([]byte(reference))
        purchase, err := NewPurchase(CreateInput{
            EventID: eventID, PurchaseRef: reference, PurchaserPartyID: partyID, PayerPartyID: partyID,
            OfferingID: offeringID, OfferingNameSnapshot: "One share", OfferingKindSnapshot: "SHARE",
            OfferingUnitPriceMinor: 100, ParticipantCapacitySnapshot: int64(source.Offering.ParticipantCapacity),
            TotalAmountMinor: 100, CurrencyCode: "IDR",
            Participants:    []PurchaseParticipantInput{{PartyID: &partyID, DisplayName: "Participant"}},
            AccessTokenHash: hash[:],
        })
        if err != nil {
            return err
        }
        created, err = NewRepository(transaction).Create(ctx, purchase)
        if err != nil {
            return err
        }
        reservation, err = Reserve(ctx, transaction, source, created)
        return err
    })
    return created, reservation, err
}

func reserveCheckoutThenRollback(ctx context.Context, database *sql.DB, eventID, offeringID, partyID, reference string, occurredAt time.Time) (string, error) {
    var purchaseID string
    err := platformdb.Within(ctx, database, func(transaction *sql.Tx) error {
        source, err := LockCheckout(ctx, transaction, eventID, offeringID, occurredAt)
        if err != nil {
            return err
        }
        hash := sha256.Sum256([]byte(reference))
        purchase, err := NewPurchase(CreateInput{
            EventID: eventID, PurchaseRef: reference, PurchaserPartyID: partyID, PayerPartyID: partyID,
            OfferingID: offeringID, OfferingNameSnapshot: "One share", OfferingKindSnapshot: "SHARE",
            OfferingUnitPriceMinor: 100, ParticipantCapacitySnapshot: int64(source.Offering.ParticipantCapacity),
            TotalAmountMinor: 100, CurrencyCode: "IDR",
            Participants:    []PurchaseParticipantInput{{PartyID: &partyID, DisplayName: "Participant"}},
            AccessTokenHash: hash[:],
        })
        if err != nil {
            return err
        }
        created, err := NewRepository(transaction).Create(ctx, purchase)
        if err != nil {
            return err
        }
        purchaseID = created.ID
        if _, err := Reserve(ctx, transaction, source, created); err != nil {
            return err
        }
        return errors.New("simulate parent command failure")
    })
    return purchaseID, err
}

func createReservationFixture(t *testing.T, ctx context.Context, database *sql.DB, prefix string, eventQuota, offeringQuota *int64) (string, string, string) {
    t.Helper()
    var eventID, offeringID, partyID string
    if err := database.QueryRowContext(ctx, `INSERT INTO qurban_events (event_year, name, status, participant_quota) VALUES ($1, $2, 'ACTIVE', $3) RETURNING id`, availableIntegrationEventYear(t, ctx, database), prefix, eventQuota).Scan(&eventID); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `INSERT INTO offerings (event_id, code, name, offering_kind, price_minor, currency_code, participant_capacity, participant_quota, status) VALUES ($1, 'SHARE', 'One share', 'SHARE', 100, 'IDR', 1, $2, 'PUBLISHED') RETURNING id`, eventID, offeringQuota).Scan(&offeringID); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `INSERT INTO parties (party_type, display_name) VALUES ('PERSON', $1) RETURNING id`, prefix+" purchaser").Scan(&partyID); err != nil {
        t.Fatal(err)
    }
    return eventID, offeringID, partyID
}

func markReservationTerminal(t *testing.T, ctx context.Context, database *sql.DB, reservationID string, status ReservationStatus) {
    t.Helper()
    now := time.Now().UTC().Add(time.Minute)
    var err error
    switch status {
    case ReservationStatusConsumed:
        _, err = database.ExecContext(ctx, `UPDATE quota_reservations SET status = $1, expires_at = NULL, consumed_at = $2 WHERE id = $3`, status, now, reservationID)
    case ReservationStatusReleased:
        _, err = database.ExecContext(ctx, `UPDATE quota_reservations SET status = $1, expires_at = NULL, released_at = $2 WHERE id = $3`, status, now, reservationID)
    case ReservationStatusExpired:
        _, err = database.ExecContext(ctx, `UPDATE quota_reservations SET status = $1, released_at = $2 WHERE id = $3`, status, now, reservationID)
    default:
        t.Fatalf("unsupported terminal status %q", status)
    }
    if err != nil {
        t.Fatal(err)
    }
}

func acquireActiveEventIntegrationLock(t *testing.T, ctx context.Context, database *sql.DB) {
    t.Helper()
    connection, err := database.Conn(ctx)
    if err != nil {
        t.Fatal(err)
    }
    if _, err := connection.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, int64(20260813)); err != nil {
        _ = connection.Close()
        t.Fatal(err)
    }
    t.Cleanup(func() {
        _, _ = connection.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, int64(20260813))
        _ = connection.Close()
    })
}

func cleanupReservationFixture(t *testing.T, ctx context.Context, database *sql.DB, eventID, partyID string) {
    t.Helper()
    if eventID != "" {
        _, _ = database.ExecContext(ctx, `DELETE FROM quota_reservations WHERE event_id = $1`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM purchase_status_history WHERE purchase_id IN (SELECT id FROM purchases WHERE event_id = $1)`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM purchase_participants WHERE event_id = $1`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM purchases WHERE event_id = $1`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM offerings WHERE event_id = $1`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM qurban_events WHERE id = $1`, eventID)
    }
    if partyID != "" {
        _, _ = database.ExecContext(ctx, `DELETE FROM parties WHERE id = $1`, partyID)
    }
}
