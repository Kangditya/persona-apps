package idempotency

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

func TestExecutePostgreSQL(t *testing.T) {
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration coverage")
    }
    db, err := database.Open(dsn)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = db.Close() })
    crypt, err := NewCipher([]Key{{ID: "test", Value: make([]byte, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    command := Command{Namespace: "test.replay." + t.Name(), Key: fmt.Sprintf("key-%d", time.Now().UnixNano()), RequestHash: "hash", Retention: time.Hour}
    failedCommand := Command{Namespace: command.Namespace, Key: command.Key + "-failed", RequestHash: "failed-hash", Retention: time.Hour}
    t.Cleanup(func() {
        _, _ = db.ExecContext(context.Background(), `DELETE FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2`, command.Namespace, command.Key)
        _, _ = db.ExecContext(context.Background(), `DELETE FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2`, failedCommand.Namespace, failedCommand.Key)
    })
    created := 0
    work := func(*sql.Tx) (Response, error) {
        created++
        return Response{Status: 201, Body: json.RawMessage("{\"ok\":true}")}, nil
    }
    first, replayed, err := Execute(context.Background(), db, crypt, command, work)
    if err != nil || replayed || first.Status != 201 {
        t.Fatalf("first Execute() = %#v, %t, %v", first, replayed, err)
    }
    second, replayed, err := Execute(context.Background(), db, crypt, command, work)
    if err != nil || !replayed || string(second.Body) != "{\"ok\":true}" || created != 1 {
        t.Fatalf("replay Execute() = %#v, %t, created=%d, err=%v", second, replayed, created, err)
    }
    command.RequestHash = "different"
    if _, _, err := Execute(context.Background(), db, crypt, command, work); err != ErrConflict {
        t.Fatalf("mismatch error = %v", err)
    }
    if _, err := db.ExecContext(context.Background(), `UPDATE idempotency_records SET expires_at = now() - interval '1 second' WHERE namespace = $1 AND idempotency_key = $2`, command.Namespace, command.Key); err != nil {
        t.Fatal(err)
    }
    fresh, replayed, err := Execute(context.Background(), db, crypt, command, work)
    if err != nil || replayed || fresh.Status != 201 || created != 2 {
        t.Fatalf("expired Execute() = %#v, %t, created=%d, err=%v", fresh, replayed, created, err)
    }

    if _, _, err := Execute(context.Background(), db, crypt, failedCommand, func(*sql.Tx) (Response, error) {
        return Response{}, errors.New("required audit write failed")
    }); err == nil {
        t.Fatal("failed command unexpectedly committed")
    }
    var failedRecords int
    if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2`, failedCommand.Namespace, failedCommand.Key).Scan(&failedRecords); err != nil {
        t.Fatal(err)
    }
    if failedRecords != 0 {
        t.Fatalf("failed command left %d replay records", failedRecords)
    }
    retried, replayed, err := Execute(context.Background(), db, crypt, failedCommand, func(*sql.Tx) (Response, error) {
        return Response{Status: 201, Body: json.RawMessage(`{"retried":true}`)}, nil
    })
    if err != nil || replayed || retried.Status != 201 {
        t.Fatalf("failed-command retry = %#v, %t, %v", retried, replayed, err)
    }

    durable := Command{Namespace: command.Namespace, Key: command.Key + "-durable", RequestHash: "durable-hash"}
    t.Cleanup(func() {
        _, _ = db.ExecContext(context.Background(), `DELETE FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2`, durable.Namespace, durable.Key)
    })
    if _, replayed, err := Execute(context.Background(), db, crypt, durable, work); err != nil || replayed {
        t.Fatalf("durable first execution replayed=%t err=%v", replayed, err)
    }
    var expiresAt sql.NullTime
    if err := db.QueryRowContext(context.Background(), `SELECT expires_at FROM idempotency_records WHERE namespace = $1 AND idempotency_key = $2`, durable.Namespace, durable.Key).Scan(&expiresAt); err != nil {
        t.Fatal(err)
    }
    if expiresAt.Valid {
        t.Fatalf("durable replay expires at %s", expiresAt.Time)
    }
    if _, replayed, err := Execute(context.Background(), db, crypt, durable, work); err != nil || !replayed {
        t.Fatalf("durable replay replayed=%t err=%v", replayed, err)
    }
}
