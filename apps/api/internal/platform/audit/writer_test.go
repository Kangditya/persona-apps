package audit

import (
    "context"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
)

func TestWriteValidatesAndNormalizesClientIP(t *testing.T) {
    database, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = database.Close() })

    mock.ExpectBegin()
    mock.ExpectRollback()
    transaction, err := database.Begin()
    if err != nil {
        t.Fatal(err)
    }
    if err := Write(context.Background(), transaction, Entry{ClientIP: "forwarded-for-is-not-an-address"}); err == nil {
        t.Fatal("Write accepted an invalid client IP")
    }
    _ = transaction.Rollback()

    mock.ExpectBegin()
    mock.ExpectExec("INSERT INTO audit_log").
        WithArgs("11111111-1111-1111-1111-111111111111", "operator-reference", "event.create", "Event", "22222222-2222-2222-2222-222222222222", "request-1", "", `null`, `{"name":"Event"}`, "operations-web", "event.manage", "192.0.2.1").
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    transaction, err = database.Begin()
    if err != nil {
        t.Fatal(err)
    }
    err = Write(context.Background(), transaction, Entry{
        ActorOperatorID: "11111111-1111-1111-1111-111111111111",
        ActorReference:  "operator-reference",
        Action:          "event.create", TargetType: "Event", TargetID: "22222222-2222-2222-2222-222222222222",
        RequestID: "request-1", Source: "operations-web", Permission: "event.manage",
        ClientIP: "::ffff:192.0.2.1", After: map[string]string{"name": "Event"},
    })
    if err != nil {
        t.Fatal(err)
    }
    if err := transaction.Commit(); err != nil {
        t.Fatal(err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}
