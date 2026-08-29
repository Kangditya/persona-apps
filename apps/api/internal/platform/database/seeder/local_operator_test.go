package seeder

import (
    "context"
    "database/sql"
    "testing"
)

type localOperatorDB struct {
    called bool
}

func (db *localOperatorDB) ExecContext(context.Context, string, ...any) (sql.Result, error) {
    db.called = true
    return nil, nil
}

func (*localOperatorDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
    return nil, nil
}

func (*localOperatorDB) QueryRowContext(context.Context, string, ...any) *sql.Row {
    return nil
}

func TestLocalOperatorSeedIsDevelopmentOnly(t *testing.T) {
    seed := localOperator{}
    if seed.Group() != Development {
        t.Fatalf("group = %q", seed.Group())
    }
    db := &localOperatorDB{}
    if err := seed.Run(context.Background(), db); err != nil {
        t.Fatal(err)
    }
    if !db.called {
        t.Fatal("seed did not execute")
    }
}
