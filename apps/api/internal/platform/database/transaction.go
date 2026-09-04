package database

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
)

// ErrCommitUncertain means PostgreSQL may have committed after the connection failed.
var ErrCommitUncertain = errors.New("transaction commit outcome is uncertain")

// Within runs work in one transaction. A panic is rolled back and propagated.
func Within(ctx context.Context, db *sql.DB, work func(*sql.Tx) error) (err error) {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }

    defer func() {
        if recovered := recover(); recovered != nil {
            _ = tx.Rollback()
            panic(recovered)
        }
        if err != nil {
            _ = tx.Rollback()
        }
    }()

    if err := work(tx); err != nil {
        return err
    }
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w: %w", ErrCommitUncertain, err)
    }
    return nil
}
