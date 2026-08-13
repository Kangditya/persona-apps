package event

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "os"
    "sync"
    "testing"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

func TestRepositoryPostgreSQL(t *testing.T) {
    db := integrationDatabase(t)
    ctx := context.Background()
    acquireIntegrationLock(t, ctx, db)
    repository := NewRepository(db)
    prefix := fmt.Sprintf("event-repository-%d", time.Now().UnixNano())
    baseYear := 3000 + int(time.Now().UnixNano()%5000)
    t.Cleanup(func() {
        if _, err := db.ExecContext(ctx, `DELETE FROM qurban_events WHERE name LIKE $1`, prefix+"%"); err != nil {
            t.Errorf("cleanup events: %v", err)
        }
    })

    create := func(year int, suffix string) Event {
        t.Helper()
        mutation, err := Create(CreateInput{EventYear: year, Name: prefix + "-" + suffix})
        if err != nil {
            t.Fatal(err)
        }
        value, err := repository.Create(ctx, mutation.Event)
        if err != nil {
            t.Fatal(err)
        }
        if value.Version != 1 {
            t.Fatalf("created version = %d, want 1", value.Version)
        }
        return value
    }

    first := create(baseYear, "first")
    second := create(baseYear+1, "second")
    third := create(baseYear+2, "third")
    if _, err := repository.Create(ctx, Event{EventYear: first.EventYear, Name: prefix + "-duplicate", Status: StatusDraft}); !errors.Is(err, ErrDuplicateYear) {
        t.Fatalf("duplicate event error = %v, want duplicate year", err)
    }

    page, err := repository.List(ctx, ListInput{Limit: 1})
    if err != nil {
        t.Fatal(err)
    }
    if len(page.Events) != 1 || page.NextCursor == "" {
        t.Fatalf("first page = %#v", page)
    }
    nextPage, err := repository.List(ctx, ListInput{Limit: 1, Cursor: page.NextCursor})
    if err != nil || len(nextPage.Events) != 1 || nextPage.Events[0].ID == page.Events[0].ID {
        t.Fatalf("cursor page = %#v, %v", nextPage, err)
    }
    insertedBetweenPages := create(baseYear+3, "inserted-between-pages")
    thirdCursor, err := encodeCursor(eventCursor{Version: 1, EventYear: third.EventYear, ID: third.ID})
    if err != nil {
        t.Fatal(err)
    }
    next, err := repository.List(ctx, ListInput{Limit: MaxListLimit, Cursor: thirdCursor})
    if err != nil {
        t.Fatal(err)
    }
    if !containsEvent(next.Events, first.ID) || !containsEvent(next.Events, second.ID) || containsEvent(next.Events, insertedBetweenPages.ID) {
        t.Fatalf("next page = %#v", next)
    }
    if _, err := repository.List(ctx, ListInput{Limit: 2, Cursor: "not-a-cursor"}); !errors.Is(err, ErrInvalidCursor) {
        t.Fatalf("invalid cursor error = %v", err)
    }

    updatedMutation, err := Update(first, UpdateInput{ExpectedVersion: first.Version, Name: stringPointer(prefix + "-updated")})
    if err != nil {
        t.Fatal(err)
    }
    updated, err := repository.Save(ctx, updatedMutation.Event, first.Version)
    if err != nil || updated.Version != first.Version+1 {
        t.Fatalf("save event = %#v, %v", updated, err)
    }
    if _, err := repository.Save(ctx, updatedMutation.Event, first.Version); !errors.Is(err, ErrStaleVersion) {
        t.Fatalf("stale event error = %v", err)
    }

    if err := assertConcurrentActivation(ctx, db, repository, second, third); err != nil {
        t.Fatal(err)
    }

    rolledBackID := ""
    rollbackErr := database.Within(ctx, db, func(transaction *sql.Tx) error {
        mutation, err := Create(CreateInput{EventYear: baseYear + 4, Name: prefix + "-rollback"})
        if err != nil {
            return err
        }
        created, err := NewRepository(transaction).Create(ctx, mutation.Event)
        if err != nil {
            return err
        }
        rolledBackID = created.ID
        return errors.New("simulate audit write failure")
    })
    if rollbackErr == nil {
        t.Fatal("transaction unexpectedly committed")
    }
    if _, err := repository.Get(ctx, rolledBackID); !errors.Is(err, ErrNotFound) {
        t.Fatalf("rolled back event lookup error = %v", err)
    }

    if _, err := db.ExecContext(ctx, `INSERT INTO qurban_events (event_year, name, status, participant_quota) VALUES ($1, $2, 'DRAFT', $3)`, baseYear+5, prefix+"-unsafe-quota", MaxSafeInteger+1); err == nil {
        t.Fatal("database accepted an unsafe Event quota")
    }
}

func assertConcurrentActivation(ctx context.Context, db *sql.DB, repository *Repository, first, second Event) error {
    publish := func(value Event) (Event, error) {
        mutation, err := Publish(value, TransitionInput{ExpectedVersion: value.Version})
        if err != nil {
            return Event{}, err
        }
        return repository.Save(ctx, mutation.Event, value.Version)
    }
    first, err := publish(first)
    if err != nil {
        return err
    }
    second, err = publish(second)
    if err != nil {
        return err
    }

    start := make(chan struct{})
    errorsByEvent := make(chan error, 2)
    var group sync.WaitGroup
    for _, value := range []Event{first, second} {
        group.Add(1)
        go func(value Event) {
            defer group.Done()
            <-start
            errorsByEvent <- database.Within(ctx, db, func(transaction *sql.Tx) error {
                locked, err := NewRepository(transaction).GetForUpdate(ctx, value.ID)
                if err != nil {
                    return err
                }
                mutation, err := Activate(locked, TransitionInput{ExpectedVersion: locked.Version})
                if err != nil {
                    return err
                }
                _, err = NewRepository(transaction).Save(ctx, mutation.Event, locked.Version)
                return err
            })
        }(value)
    }
    close(start)
    group.Wait()
    close(errorsByEvent)

    succeeded := 0
    conflicted := 0
    for err := range errorsByEvent {
        switch {
        case err == nil:
            succeeded++
        case errors.Is(err, ErrActiveConflict):
            conflicted++
        default:
            return fmt.Errorf("concurrent activation: %w", err)
        }
    }
    if succeeded != 1 || conflicted != 1 {
        return fmt.Errorf("concurrent activation results: succeeded=%d conflicted=%d", succeeded, conflicted)
    }
    return nil
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

func containsEvent(events []Event, id string) bool {
    for _, value := range events {
        if value.ID == id {
            return true
        }
    }
    return false
}
