package idempotency

import (
	"context"
	"database/sql"
	"encoding/json"
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
	defer db.Close()
	crypt, err := NewCipher([]Key{{ID: "test", Value: make([]byte, 32)}})
	if err != nil {
		t.Fatal(err)
	}
	command := Command{Namespace: "test.replay." + t.Name(), Key: fmt.Sprintf("key-%d", time.Now().UnixNano()), RequestHash: "hash", Retention: time.Hour}
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
}
