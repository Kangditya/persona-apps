package offering

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "os"
    "sync"
    "testing"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/event"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

func TestRepositoryPostgreSQL(t *testing.T) {
    db := integrationDatabase(t)
    ctx := context.Background()
    acquireIntegrationLock(t, ctx, db)
    prefix := fmt.Sprintf("offering-repository-%d", time.Now().UnixNano())
    baseYear := 3000 + int(time.Now().UnixNano()%5000)
    eventRepository := event.NewRepository(db)
    parent := createActiveEvent(t, ctx, eventRepository, baseYear, prefix)
    repository := NewRepository(db)
    partyID := createParty(t, ctx, db, prefix)
    t.Cleanup(func() {
        cleanupOfferingFixture(t, ctx, db, parent.ID, partyID)
    })

    create := func(code string, quota *int64, publish bool) Offering {
        t.Helper()
        mutation, err := Create(parent, CreateInput{
            EventID:             parent.ID,
            Code:                code,
            Name:                prefix + " " + code,
            Kind:                "SHARE",
            PriceMinor:          1_500_000,
            CurrencyCode:        "IDR",
            ParticipantCapacity: 7,
            ParticipantQuota:    quota,
        })
        if err != nil {
            t.Fatal(err)
        }
        value, err := repository.Create(ctx, mutation.Offering)
        if err != nil {
            t.Fatal(err)
        }
        if !publish {
            return value
        }
        published, err := Publish(parent, value, PublishInput{ExpectedVersion: value.Version, OccurredAt: time.Now()})
        if err != nil {
            t.Fatal(err)
        }
        value, err = repository.Save(ctx, published.Offering, value.Version)
        if err != nil {
            t.Fatal(err)
        }
        return value
    }

    offeringA := create("A-SHARE", int64Pointer(8), true)
    offeringB := create("B-SHARE", int64Pointer(8), true)
    offeringC := create("C-DRAFT", nil, false)
    if _, err := repository.Create(ctx, Offering{EventID: parent.ID, Code: offeringA.Code, Name: prefix + " duplicate", Kind: "SHARE", PriceMinor: 1, CurrencyCode: "IDR", ParticipantCapacity: 1, Status: StatusDraft}); !errors.Is(err, ErrDuplicateCode) {
        t.Fatalf("duplicate offering error = %v, want duplicate code", err)
    }

    page, err := repository.List(ctx, parent.ID, ListInput{Limit: 2})
    if err != nil {
        t.Fatal(err)
    }
    if len(page.Offerings) != 2 || page.Offerings[0].ID != offeringA.ID || page.Offerings[1].ID != offeringB.ID || page.NextCursor == "" {
        t.Fatalf("offering page = %#v", page)
    }
    next, err := repository.List(ctx, parent.ID, ListInput{Limit: 2, Cursor: page.NextCursor})
    if err != nil || len(next.Offerings) != 1 || next.Offerings[0].ID != offeringC.ID {
        t.Fatalf("offering next page = %#v, %v", next, err)
    }
    if _, err := repository.List(ctx, parent.ID, ListInput{Limit: 2, Cursor: "not-a-cursor"}); !errors.Is(err, ErrInvalidCursor) {
        t.Fatalf("invalid cursor error = %v", err)
    }

    insertReservation(t, ctx, db, parent, offeringA, partyID, prefix, 1, "RESERVED", 3)
    insertReservation(t, ctx, db, parent, offeringB, partyID, prefix, 2, "CONSUMED", 4)
    insertReservation(t, ctx, db, parent, offeringA, partyID, prefix, 3, "RELEASED", 99)
    insertReservation(t, ctx, db, parent, offeringA, partyID, prefix, 4, "EXPIRED", 99)

    publicPage, err := repository.ListPublic(ctx, parent.ID, ListInput{Limit: 1})
    if err != nil {
        t.Fatal(err)
    }
    if len(publicPage.Offerings) != 1 || publicPage.Offerings[0].Offering.ID != offeringA.ID || !equalInt64(publicPage.Offerings[0].AvailableParticipantUnits, int64Pointer(3)) || publicPage.NextCursor == "" {
        t.Fatalf("public offering page = %#v", publicPage)
    }
    publicNext, err := repository.ListPublic(ctx, parent.ID, ListInput{Limit: 1, Cursor: publicPage.NextCursor})
    if err != nil || len(publicNext.Offerings) != 1 || publicNext.Offerings[0].Offering.ID != offeringB.ID || !equalInt64(publicNext.Offerings[0].AvailableParticipantUnits, int64Pointer(3)) {
        t.Fatalf("public offering next page = %#v, %v", publicNext, err)
    }
    if _, err := repository.GetPublic(ctx, offeringC.ID); !errors.Is(err, ErrNotFound) {
        t.Fatalf("draft public lookup error = %v", err)
    }

    updatedMutation, err := Update(parent, offeringC, UpdateInput{ExpectedVersion: offeringC.Version, Name: stringPointer(prefix + " updated")})
    if err != nil {
        t.Fatal(err)
    }
    updated, err := repository.Save(ctx, updatedMutation.Offering, offeringC.Version)
    if err != nil || updated.Version != offeringC.Version+1 {
        t.Fatalf("save offering = %#v, %v", updated, err)
    }
    if _, err := repository.Save(ctx, updatedMutation.Offering, offeringC.Version); !errors.Is(err, ErrStaleVersion) {
        t.Fatalf("stale offering error = %v", err)
    }

    concurrent := create("D-CONCURRENT", nil, false)
    if err := assertConcurrentUpdate(ctx, db, repository, parent, concurrent, prefix); err != nil {
        t.Fatal(err)
    }

    unavailable, err := MarkUnavailable(offeringA, TransitionInput{ExpectedVersion: offeringA.Version})
    if err != nil {
        t.Fatal(err)
    }
    if _, err := repository.Save(ctx, unavailable.Offering, offeringA.Version); err != nil {
        t.Fatal(err)
    }
    if _, err := repository.GetPublic(ctx, offeringA.ID); !errors.Is(err, ErrNotFound) {
        t.Fatalf("unavailable public lookup error = %v", err)
    }

    suspended, err := event.Suspend(parent, event.TransitionInput{ExpectedVersion: parent.Version})
    if err != nil {
        t.Fatal(err)
    }
    if _, err := eventRepository.Save(ctx, suspended.Event, parent.Version); err != nil {
        t.Fatal(err)
    }
    if _, err := repository.ListPublic(ctx, parent.ID, ListInput{Limit: 1}); !errors.Is(err, ErrNotFound) {
        t.Fatalf("suspended event public list error = %v", err)
    }

    rolledBackID := ""
    rollbackErr := database.Within(ctx, db, func(transaction *sql.Tx) error {
        mutation, err := Create(parent, CreateInput{EventID: parent.ID, Code: "ROLLBACK", Name: prefix + " rollback", Kind: "SHARE", PriceMinor: 1, CurrencyCode: "IDR", ParticipantCapacity: 1})
        if err != nil {
            return err
        }
        created, err := NewRepository(transaction).Create(ctx, mutation.Offering)
        if err != nil {
            return err
        }
        rolledBackID = created.ID
        return errors.New("simulate outbox write failure")
    })
    if rollbackErr == nil {
        t.Fatal("transaction unexpectedly committed")
    }
    if _, err := repository.Get(ctx, rolledBackID); !errors.Is(err, ErrNotFound) {
        t.Fatalf("rolled back offering lookup error = %v", err)
    }

    if _, err := db.ExecContext(ctx, `UPDATE offerings SET price_minor = $1 WHERE id = $2`, MaxSafeInteger+1, offeringB.ID); err == nil {
        t.Fatal("database accepted an unsafe Offering price")
    }
}

func createActiveEvent(t *testing.T, ctx context.Context, repository *event.Repository, year int, name string) event.Event {
    t.Helper()
    quota := int64(10)
    createdMutation, err := event.Create(event.CreateInput{EventYear: year, Name: name, ParticipantQuota: &quota})
    if err != nil {
        t.Fatal(err)
    }
    created, err := repository.Create(ctx, createdMutation.Event)
    if err != nil {
        t.Fatal(err)
    }
    publishedMutation, err := event.Publish(created, event.TransitionInput{ExpectedVersion: created.Version})
    if err != nil {
        t.Fatal(err)
    }
    published, err := repository.Save(ctx, publishedMutation.Event, created.Version)
    if err != nil {
        t.Fatal(err)
    }
    activeMutation, err := event.Activate(published, event.TransitionInput{ExpectedVersion: published.Version})
    if err != nil {
        t.Fatal(err)
    }
    active, err := repository.Save(ctx, activeMutation.Event, published.Version)
    if err != nil {
        t.Fatal(err)
    }
    return active
}

func createParty(t *testing.T, ctx context.Context, db *sql.DB, prefix string) string {
    t.Helper()
    var id string
    if err := db.QueryRowContext(ctx, `INSERT INTO parties (party_type, display_name) VALUES ('PERSON', $1) RETURNING id`, prefix+" purchaser").Scan(&id); err != nil {
        t.Fatal(err)
    }
    return id
}

func insertReservation(t *testing.T, ctx context.Context, db *sql.DB, parent event.Event, value Offering, partyID, prefix string, number int, status string, units int) {
    t.Helper()
    var purchaseID string
    hash := make([]byte, 32)
    hash[0] = byte(number)
    if err := db.QueryRowContext(ctx, `INSERT INTO purchases (event_id, purchase_ref, channel, purchaser_party_id, payer_party_id, offering_id, participant_count, offering_name_snapshot, offering_kind_snapshot, offering_unit_price_minor, participant_capacity_snapshot, total_amount_minor, currency_code, status, access_token_hash) VALUES ($1, $2, 'COMMON', $3, $3, $4, 1, $5, $6, 1, 1, 1, 'IDR', 'PENDING_PAYMENT', $7) RETURNING id`, parent.ID, fmt.Sprintf("%s-purchase-%d", prefix, number), partyID, value.ID, value.Name, value.Kind, hash).Scan(&purchaseID); err != nil {
        t.Fatal(err)
    }
    query := `INSERT INTO quota_reservations (event_id, purchase_id, offering_id, attempt_no, participant_units, status) VALUES ($1, $2, $3, 1, $4, $5)`
    arguments := []any{parent.ID, purchaseID, value.ID, units, status}
    switch status {
    case "CONSUMED":
        query = `INSERT INTO quota_reservations (event_id, purchase_id, offering_id, attempt_no, participant_units, status, consumed_at) VALUES ($1, $2, $3, 1, $4, $5, now())`
    case "RELEASED":
        query = `INSERT INTO quota_reservations (event_id, purchase_id, offering_id, attempt_no, participant_units, status, released_at) VALUES ($1, $2, $3, 1, $4, $5, now())`
    case "EXPIRED":
        query = `INSERT INTO quota_reservations (event_id, purchase_id, offering_id, attempt_no, participant_units, status, expires_at, released_at) VALUES ($1, $2, $3, 1, $4, $5, now(), now())`
    }
    if _, err := db.ExecContext(ctx, query, arguments...); err != nil {
        t.Fatal(err)
    }
}

func assertConcurrentUpdate(ctx context.Context, db *sql.DB, repository *Repository, parent event.Event, current Offering, prefix string) error {
    first, err := Update(parent, current, UpdateInput{ExpectedVersion: current.Version, Name: stringPointer(prefix + " concurrent one")})
    if err != nil {
        return err
    }
    second, err := Update(parent, current, UpdateInput{ExpectedVersion: current.Version, Name: stringPointer(prefix + " concurrent two")})
    if err != nil {
        return err
    }
    start := make(chan struct{})
    errorsByUpdate := make(chan error, 2)
    var group sync.WaitGroup
    for _, mutation := range []Mutation{first, second} {
        group.Add(1)
        go func(mutation Mutation) {
            defer group.Done()
            <-start
            _, err := repository.Save(ctx, mutation.Offering, current.Version)
            errorsByUpdate <- err
        }(mutation)
    }
    close(start)
    group.Wait()
    close(errorsByUpdate)

    succeeded := 0
    stale := 0
    for err := range errorsByUpdate {
        switch {
        case err == nil:
            succeeded++
        case errors.Is(err, ErrStaleVersion):
            stale++
        default:
            return fmt.Errorf("concurrent offering update: %w", err)
        }
    }
    if succeeded != 1 || stale != 1 {
        return fmt.Errorf("concurrent offering update results: succeeded=%d stale=%d", succeeded, stale)
    }
    return nil
}

func cleanupOfferingFixture(t *testing.T, ctx context.Context, db *sql.DB, eventID, partyID string) {
    t.Helper()
    for _, statement := range []struct {
        query string
        id    string
    }{
        {query: `DELETE FROM quota_reservations WHERE event_id = $1`, id: eventID},
        {query: `DELETE FROM purchases WHERE event_id = $1`, id: eventID},
        {query: `DELETE FROM offerings WHERE event_id = $1`, id: eventID},
        {query: `DELETE FROM qurban_events WHERE id = $1`, id: eventID},
        {query: `DELETE FROM parties WHERE id = $1`, id: partyID},
    } {
        if _, err := db.ExecContext(ctx, statement.query, statement.id); err != nil {
            t.Errorf("cleanup offering fixture: %v", err)
        }
    }
}

func integrationDatabase(t *testing.T) *sql.DB {
    t.Helper()
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration coverage")
    }
    db, err := database.Open(dsn)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = db.Close() })
    return db
}

func acquireIntegrationLock(t *testing.T, ctx context.Context, db *sql.DB) {
    t.Helper()
    connection, err := db.Conn(ctx)
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

func equalInt64(got, want *int64) bool {
    return got == nil && want == nil || got != nil && want != nil && *got == *want
}
