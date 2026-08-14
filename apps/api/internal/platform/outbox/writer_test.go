package outbox

import (
    "context"
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
)

func TestWriteUsesTheCallerTransaction(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    mock.ExpectBegin()
    mock.ExpectExec(regexp.QuoteMeta("INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload, occurred_at) VALUES ($1, $2::uuid, $3, $4::jsonb, $5)")).
        WithArgs("Event", "11111111-1111-1111-1111-111111111111", "EventPublished", `{"status":"PUBLISHED"}`, sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectRollback()

    tx, err := db.BeginTx(context.Background(), nil)
    if err != nil {
        t.Fatal(err)
    }
    if err := Write(context.Background(), tx, Event{AggregateType: "Event", AggregateID: "11111111-1111-1111-1111-111111111111", EventType: "EventPublished", Payload: map[string]string{"status": "PUBLISHED"}, OccurredAt: time.Now()}); err != nil {
        t.Fatal(err)
    }
    _ = tx.Rollback()
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}
