package idempotency

import (
    "bytes"
    "context"
    "database/sql"
    "errors"
    "net/http"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
)

func TestCipherEncryptsAndAuthenticatesReplay(t *testing.T) {
    crypt, err := NewCipher([]Key{{ID: "primary", Value: bytes.Repeat([]byte{1}, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    body := []byte("{\"access_token\":\"secret\"}")
    sealed, err := crypt.Seal([]byte("bound"), body)
    if err != nil {
        t.Fatal(err)
    }
    if bytes.Contains(sealed, []byte("secret")) {
        t.Fatal("replay envelope retained plaintext")
    }
    opened, err := crypt.Open([]byte("bound"), sealed)
    if err != nil || !bytes.Equal(opened, body) {
        t.Fatalf("Open() = %q, %v", opened, err)
    }
    if _, err := crypt.Open([]byte("other"), sealed); !errors.Is(err, ErrUnavailableReplay) {
        t.Fatalf("AAD mismatch error = %v", err)
    }
}

func TestExecuteStoresAndReplaysEncryptedResponse(t *testing.T) {
    crypt, err := NewCipher([]Key{{ID: "primary", Value: bytes.Repeat([]byte{1}, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    command := Command{Namespace: "storefront.purchase.scope", Key: "retry-key", RequestHash: "hash", Retention: time.Hour}
    t.Run("first execution", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatal(err)
        }
        defer db.Close()
        mock.ExpectBegin()
        mock.ExpectExec("DELETE FROM idempotency_records").WithArgs(command.Namespace, command.Key).WillReturnResult(sqlmock.NewResult(0, 0))
        mock.ExpectQuery("INSERT INTO idempotency_records").WithArgs(command.Namespace, command.Key, command.RequestHash, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"namespace"}).AddRow(command.Namespace))
        mock.ExpectExec("UPDATE idempotency_records").WithArgs(201, sqlmock.AnyArg(), command.Namespace, command.Key).WillReturnResult(sqlmock.NewResult(0, 1))
        mock.ExpectCommit()
        result, replayed, err := Execute(context.Background(), db, crypt, command, func(*sql.Tx) (Response, error) { return Response{Status: 201, Body: []byte("{\"created\":true}")}, nil })
        if err != nil || replayed || result.Status != 201 {
            t.Fatalf("Execute() = %#v, %t, %v", result, replayed, err)
        }
        if err := mock.ExpectationsWereMet(); err != nil {
            t.Fatal(err)
        }
    })
    t.Run("same request replays", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatal(err)
        }
        defer db.Close()
        stored, err := crypt.Seal(aad(command, 201), []byte("{\"created\":true}"))
        if err != nil {
            t.Fatal(err)
        }
        mock.ExpectBegin()
        mock.ExpectExec("DELETE FROM idempotency_records").WithArgs(command.Namespace, command.Key).WillReturnResult(sqlmock.NewResult(0, 0))
        mock.ExpectQuery("INSERT INTO idempotency_records").WithArgs(command.Namespace, command.Key, command.RequestHash, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"namespace"}))
        mock.ExpectQuery("SELECT request_hash").WithArgs(command.Namespace, command.Key).WillReturnRows(sqlmock.NewRows([]string{"request_hash", "response_status", "response_body"}).AddRow(command.RequestHash, 201, stored))
        mock.ExpectCommit()
        result, replayed, err := Execute(context.Background(), db, crypt, command, func(*sql.Tx) (Response, error) { t.Fatal("replay executed work"); return Response{}, nil })
        if err != nil || !replayed || string(result.Body) != "{\"created\":true}" {
            t.Fatalf("Execute() = %#v, %t, %v", result, replayed, err)
        }
        if err := mock.ExpectationsWereMet(); err != nil {
            t.Fatal(err)
        }
    })
}

func TestExecuteStoresDurableReplayWithoutExpiry(t *testing.T) {
    crypt, err := NewCipher([]Key{{ID: "primary", Value: bytes.Repeat([]byte{1}, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    command := Command{Namespace: "storefront.purchase.create.guest", Key: "durable-key", RequestHash: "hash"}
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    mock.ExpectBegin()
    mock.ExpectExec("DELETE FROM idempotency_records").WithArgs(command.Namespace, command.Key).WillReturnResult(sqlmock.NewResult(0, 0))
    mock.ExpectQuery("INSERT INTO idempotency_records").WithArgs(command.Namespace, command.Key, command.RequestHash, nil).WillReturnRows(sqlmock.NewRows([]string{"namespace"}).AddRow(command.Namespace))
    mock.ExpectExec("UPDATE idempotency_records").WithArgs(201, sqlmock.AnyArg(), command.Namespace, command.Key).WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectCommit()
    result, replayed, err := Execute(context.Background(), db, crypt, command, func(*sql.Tx) (Response, error) {
        return Response{Status: 201, Body: []byte(`{"created":true}`)}, nil
    })
    if err != nil || replayed || result.Status != http.StatusCreated {
        t.Fatalf("Execute() = %#v, %t, %v", result, replayed, err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func TestExecuteRejectsNegativeRetention(t *testing.T) {
    crypt, err := NewCipher([]Key{{ID: "primary", Value: bytes.Repeat([]byte{1}, 32)}})
    if err != nil {
        t.Fatal(err)
    }
    if _, _, err := Execute(context.Background(), nil, crypt, Command{Namespace: "scope", Key: "retry-key", RequestHash: "hash", Retention: -time.Second}, func(*sql.Tx) (Response, error) {
        return Response{}, nil
    }); err == nil {
        t.Fatal("negative retention was accepted")
    }
}
