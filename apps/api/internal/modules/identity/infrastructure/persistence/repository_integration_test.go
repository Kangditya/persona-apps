package persistence

import (
    "context"
    "crypto/sha256"
    "database/sql"
    "errors"
    "fmt"
    "os"
    "testing"
    "time"

    identitydomain "github.com/Kangditya/persona-apps/apps/api/internal/modules/identity/domain"
    platformdb "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

func TestRepositoryPostgreSQL(t *testing.T) {
    database := integrationDatabase(t)
    ctx := context.Background()
    acquireIntegrationLock(t, ctx, database)
    repository := NewRepository(database)
    prefix := fmt.Sprintf("identity-repository-%d", time.Now().UnixNano())
    eventYear := 5000 + int(time.Now().UnixNano()%4000)
    var eventID, offeringID, purchaseRef, partyID string
    t.Cleanup(func() {
        if purchaseRef != "" {
            _, _ = database.ExecContext(ctx, `DELETE FROM purchases WHERE purchase_ref = $1`, purchaseRef)
        }
        if offeringID != "" {
            _, _ = database.ExecContext(ctx, `DELETE FROM offerings WHERE id = $1`, offeringID)
        }
        if eventID != "" {
            _, _ = database.ExecContext(ctx, `DELETE FROM qurban_events WHERE id = $1`, eventID)
        }
        if partyID != "" {
            _, _ = database.ExecContext(ctx, `DELETE FROM parties WHERE id = $1`, partyID)
        }
    })

    email := prefix + "@example.test"
    party, err := identitydomain.NewParty(identitydomain.CreateInput{Type: identitydomain.PartyTypePerson, DisplayName: "  Purchaser  ", Email: &email})
    if err != nil {
        t.Fatal(err)
    }
    party, err = repository.Create(ctx, party)
    if err != nil {
        t.Fatal(err)
    }
    partyID = party.ID
    roundTripped, err := repository.Get(ctx, party.ID)
    if err != nil || roundTripped.DisplayName != "Purchaser" || roundTripped.Email == nil || *roundTripped.Email != email || roundTripped.Phone != nil || roundTripped.CreatedAt.IsZero() || roundTripped.UpdatedAt.IsZero() {
        t.Fatalf("round trip Party = %#v, %v", roundTripped, err)
    }

    if err := database.QueryRowContext(ctx, `INSERT INTO qurban_events (event_year, name, status) VALUES ($1, $2, 'DRAFT') RETURNING id`, eventYear, prefix).Scan(&eventID); err != nil {
        t.Fatal(err)
    }
    if err := database.QueryRowContext(ctx, `INSERT INTO offerings (event_id, code, name, offering_kind, price_minor, currency_code, participant_capacity, status) VALUES ($1, 'SHARE', $2, 'SHARE', 100, 'IDR', 1, 'DRAFT') RETURNING id`, eventID, prefix).Scan(&offeringID); err != nil {
        t.Fatal(err)
    }
    purchaseRef = prefix
    tokenHash := sha256.Sum256([]byte(prefix))
    if _, err := database.ExecContext(ctx, `
INSERT INTO purchases (event_id, purchase_ref, channel, purchaser_party_id, payer_party_id, offering_id, participant_count, offering_name_snapshot, offering_kind_snapshot, offering_unit_price_minor, participant_capacity_snapshot, total_amount_minor, currency_code, status, access_token_hash)
VALUES ($1, $2, 'COMMON', $3, $3, $4, 1, 'Share', 'SHARE', 100, 1, 100, 'IDR', 'PENDING_PAYMENT', $5)`, eventID, purchaseRef, party.ID, offeringID, tokenHash[:]); err != nil {
        t.Fatal(err)
    }
    var references int
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM purchases WHERE purchase_ref = $1 AND purchaser_party_id = $2 AND payer_party_id = $2`, purchaseRef, party.ID).Scan(&references); err != nil || references != 1 {
        t.Fatalf("Party role reference count = %d, %v", references, err)
    }
    var partyCount int
    if err := database.QueryRowContext(ctx, `SELECT count(*) FROM parties WHERE id = $1`, party.ID).Scan(&partyCount); err != nil || partyCount != 1 {
        t.Fatalf("Party count = %d, %v", partyCount, err)
    }

    rolledBackID := ""
    rollbackErr := platformdb.Within(ctx, database, func(transaction *sql.Tx) error {
        value, createErr := identitydomain.NewParty(identitydomain.CreateInput{Type: identitydomain.PartyTypeOrganization, DisplayName: prefix + " rollback"})
        if createErr != nil {
            return createErr
        }
        created, createErr := NewRepository(transaction).Create(ctx, value)
        if createErr != nil {
            return createErr
        }
        rolledBackID = created.ID
        return errors.New("simulate parent command failure")
    })
    if rollbackErr == nil {
        t.Fatal("transaction unexpectedly committed")
    }
    if _, err := repository.Get(ctx, rolledBackID); !errors.Is(err, identitydomain.ErrNotFound) {
        t.Fatalf("rolled-back Party lookup error = %v", err)
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
    if _, err := connection.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, int64(20260821)); err != nil {
        _ = connection.Close()
        t.Fatal(err)
    }
    t.Cleanup(func() {
        _, _ = connection.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, int64(20260821))
        _ = connection.Close()
    })
}
