package database

import (
    "context"
    "database/sql"
    "errors"
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
)

func TestWithinCommitsOnSuccess(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    mock.ExpectBegin()
    mock.ExpectCommit()
    if err := Within(context.Background(), db, func(*sql.Tx) error { return nil }); err != nil {
        t.Fatalf("Within() error = %v", err)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func TestWithinRollsBackOnErrorAndPanic(t *testing.T) {
    t.Run("callback error", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatal(err)
        }
        defer db.Close()
        mock.ExpectBegin()
        mock.ExpectRollback()
        want := errors.New("stop")
        if err := Within(context.Background(), db, func(*sql.Tx) error { return want }); !errors.Is(err, want) {
            t.Fatalf("Within() error = %v, want %v", err, want)
        }
        if err := mock.ExpectationsWereMet(); err != nil {
            t.Fatal(err)
        }
    })
    t.Run("panic", func(t *testing.T) {
        db, mock, err := sqlmock.New()
        if err != nil {
            t.Fatal(err)
        }
        defer db.Close()
        mock.ExpectBegin()
        mock.ExpectRollback()
        defer func() {
            if recover() != "panic" {
                t.Fatal("panic was not propagated")
            }
            if err := mock.ExpectationsWereMet(); err != nil {
                t.Fatal(err)
            }
        }()
        _ = Within(context.Background(), db, func(*sql.Tx) error { panic("panic") })
    })
}
