package database

import (
    "context"
    "database/sql"
)

// DBTX is the database surface shared by repositories and transaction tests.
type DBTX interface {
    ExecContext(context.Context, string, ...any) (sql.Result, error)
    QueryContext(context.Context, string, ...any) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...any) *sql.Row
}
