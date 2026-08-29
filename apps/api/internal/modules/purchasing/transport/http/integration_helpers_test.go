package http

import (
    "context"
    "crypto/sha256"
    "database/sql"
    "os"
    "testing"
    "time"

    purchasingpersistence "github.com/Kangditya/persona-apps/apps/api/internal/modules/purchasing/infrastructure/persistence"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

func integrationDatabase(t *testing.T) *sql.DB {
    t.Helper()
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration coverage")
    }
    database, err := platformdb.Open(dsn)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })
    return database
}

func availableIntegrationEventYear(t *testing.T, ctx context.Context, database *sql.DB) int {
    t.Helper()
    var year int
    if err := database.QueryRowContext(ctx, `SELECT candidate FROM generate_series(6000, 7999) AS candidate WHERE NOT EXISTS (SELECT 1 FROM qurban_events WHERE event_year = candidate) ORDER BY candidate LIMIT 1`).Scan(&year); err != nil {
        t.Fatal(err)
    }
    return year
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

func reserveCheckout(ctx context.Context, database *sql.DB, eventID, offeringID, partyID, reference string, occurredAt time.Time) (Purchase, Reservation, error) {
    var created Purchase
    var reservation Reservation
    err := platformdb.Within(ctx, database, func(transaction *sql.Tx) error {
        source, err := purchasingpersistence.LockCheckout(ctx, transaction, eventID, offeringID, occurredAt)
        if err != nil {
            return err
        }
        hash := sha256.Sum256([]byte(reference))
        purchase, err := NewPurchase(CreateInput{
            EventID: eventID, PurchaseRef: reference, PurchaserPartyID: partyID, PayerPartyID: partyID,
            OfferingID: offeringID, OfferingNameSnapshot: "One share", OfferingKindSnapshot: "SHARE",
            OfferingUnitPriceMinor: 100, ParticipantCapacitySnapshot: int64(source.Offering.ParticipantCapacity),
            TotalAmountMinor: 100, CurrencyCode: "IDR",
            Participants: []PurchaseParticipantInput{{PartyID: &partyID, DisplayName: "Participant"}}, AccessTokenHash: hash[:],
        })
        if err != nil {
            return err
        }
        created, err = purchasingpersistence.NewRepository(transaction).Create(ctx, purchase)
        if err != nil {
            return err
        }
        reservation, err = purchasingpersistence.Reserve(ctx, transaction, source, created)
        return err
    })
    return created, reservation, err
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
