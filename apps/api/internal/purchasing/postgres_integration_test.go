package purchasing

import (
    "context"
    "crypto/sha256"
    "database/sql"
    "errors"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/identity"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

func TestRepositoryPostgreSQL(t *testing.T) {
    database := integrationDatabase(t)
    ctx := context.Background()
    acquireIntegrationLock(t, ctx, database)
    prefix := fmt.Sprintf("purchase-repository-%d", time.Now().UnixNano())
    eventYear := availableIntegrationEventYear(t, ctx, database)
    var eventID, offeringID, purchaserID, payerID string
    t.Cleanup(func() {
        cleanupIntegrationPurchases(t, ctx, database, eventID, offeringID, purchaserID, payerID)
    })

    partyRepository := identity.NewRepository(database)
    purchaserEmail := prefix + "-purchaser@example.test"
    purchaser, err := identity.NewParty(identity.CreateInput{Type: identity.PartyTypePerson, DisplayName: "Purchaser", Email: &purchaserEmail})
    if err != nil {
        t.Fatal(err)
    }
    purchaser, err = partyRepository.Create(ctx, purchaser)
    if err != nil {
        t.Fatal(err)
    }
    purchaserID = purchaser.ID
    payer, err := identity.NewParty(identity.CreateInput{Type: identity.PartyTypePerson, DisplayName: "Payer"})
    if err != nil {
        t.Fatal(err)
    }
    payer, err = partyRepository.Create(ctx, payer)
    if err != nil {
        t.Fatal(err)
    }
    payerID = payer.ID

    if err := database.QueryRowContext(ctx, `INSERT INTO qurban_events (event_year, name, status) VALUES ($1, $2, 'DRAFT') RETURNING id`, eventYear, prefix).Scan(&eventID); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `INSERT INTO offerings (event_id, code, name, offering_kind, price_minor, currency_code, participant_capacity, status) VALUES ($1, 'SHARE', 'Stored share', 'SHARE', 100, 'IDR', 2, 'DRAFT') RETURNING id`, eventID).Scan(&offeringID); err != nil {
        t.Fatal(err)
    }

    hash := sha256.Sum256([]byte(prefix))
    input := CreateInput{
        EventID: eventID, PurchaseRef: prefix, PurchaserPartyID: purchaserID, PayerPartyID: payerID,
        OfferingID: offeringID, OfferingNameSnapshot: "Stored share", OfferingKindSnapshot: "SHARE",
        OfferingUnitPriceMinor: 100, ParticipantCapacitySnapshot: 2, TotalAmountMinor: 200, CurrencyCode: "IDR",
        Participants:    []PurchaseParticipantInput{{PartyID: &purchaserID, DisplayName: "Purchaser"}, {DisplayName: "Name only"}},
        AccessTokenHash: hash[:],
    }
    purchase, err := NewPurchase(input)
    if err != nil {
        t.Fatal(err)
    }
    var created Purchase
    if err := platformdb.Within(ctx, database, func(transaction *sql.Tx) error {
        var createErr error
        created, createErr = NewRepository(transaction).Create(ctx, purchase)
        return createErr
    }); err != nil {
        t.Fatal(err)
    }
    found, err := NewRepository(database).Get(ctx, created.ID)
    if err != nil {
        t.Fatal(err)
    }
    if found.Purchase.EventID != eventID || found.Purchase.OfferingID != offeringID || found.Purchase.Status != StatusPendingPayment || found.Purchaser.ID != purchaserID || found.Payer == nil || found.Payer.ID != payerID || len(found.Purchase.Participants) != 2 || found.Purchase.Participants[0].PartyID == nil || *found.Purchase.Participants[0].PartyID != purchaserID || found.Purchase.Participants[1].PartyID != nil {
        t.Fatalf("persisted purchase = %#v", found)
    }
    var fromStatus sql.NullString
    var toStatus string
    if err := database.QueryRowContext(ctx, `SELECT from_status, to_status FROM purchase_status_history WHERE purchase_id = $1`, created.ID).Scan(&fromStatus, &toStatus); err != nil {
        t.Fatal(err)
    }
    if !fromStatus.Valid || fromStatus.String != string(StatusDraft) || toStatus != string(StatusPendingPayment) {
        t.Fatalf("initial history = %q/%q", fromStatus.String, toStatus)
    }

    rolledBack := purchase
    rolledBack.PurchaseRef = prefix + "-rollback"
    rollbackHash := sha256.Sum256([]byte(prefix + "-rollback"))
    rolledBack.accessTokenHash = rollbackHash[:]
    var rolledBackID string
    rollbackErr := platformdb.Within(ctx, database, func(transaction *sql.Tx) error {
        inserted, createErr := NewRepository(transaction).Create(ctx, rolledBack)
        if createErr != nil {
            return createErr
        }
        rolledBackID = inserted.ID
        return errors.New("simulate parent command failure")
    })
    if rollbackErr == nil || rolledBackID == "" {
        t.Fatalf("rollback result/id = %v/%q", rollbackErr, rolledBackID)
    }
    if _, err := NewRepository(database).Get(ctx, rolledBackID); !errors.Is(err, ErrNotFound) {
        t.Fatalf("rolled-back purchase lookup error = %v", err)
    }
}

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

func acquireIntegrationLock(t *testing.T, ctx context.Context, database *sql.DB) {
    t.Helper()
    connection, err := database.Conn(ctx)
    if err != nil {
        t.Fatal(err)
    }
    if _, err := connection.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, int64(20260822)); err != nil {
        _ = connection.Close()
        t.Fatal(err)
    }
    t.Cleanup(func() {
        _, _ = connection.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, int64(20260822))
        _ = connection.Close()
    })
}

func availableIntegrationEventYear(t *testing.T, ctx context.Context, database *sql.DB) int {
    t.Helper()
    var year int
    if err := database.QueryRowContext(ctx, `SELECT candidate FROM generate_series(6000, 7999) AS candidate WHERE NOT EXISTS (SELECT 1 FROM qurban_events WHERE event_year = candidate) ORDER BY candidate LIMIT 1`).Scan(&year); err != nil {
        t.Fatal(err)
    }
    return year
}

func cleanupIntegrationPurchases(t *testing.T, ctx context.Context, database *sql.DB, eventID, offeringID, purchaserID, payerID string) {
    t.Helper()
    if eventID != "" {
        _, _ = database.ExecContext(ctx, `DELETE FROM quota_reservations WHERE event_id = $1`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM purchase_status_history WHERE purchase_id IN (SELECT id FROM purchases WHERE event_id = $1)`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM purchase_participants WHERE event_id = $1`, eventID)
        _, _ = database.ExecContext(ctx, `DELETE FROM purchases WHERE event_id = $1`, eventID)
    }
    if offeringID != "" {
        _, _ = database.ExecContext(ctx, `DELETE FROM offerings WHERE id = $1`, offeringID)
    }
    if eventID != "" {
        _, _ = database.ExecContext(ctx, `DELETE FROM qurban_events WHERE id = $1`, eventID)
    }
    for _, partyID := range []string{purchaserID, payerID} {
        if partyID != "" {
            _, _ = database.ExecContext(ctx, `DELETE FROM parties WHERE id = $1`, partyID)
        }
    }
}
