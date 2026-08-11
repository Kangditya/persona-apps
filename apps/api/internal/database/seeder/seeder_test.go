package seeder

import (
    "context"
    "database/sql"
    "errors"
    "regexp"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
)

type testSeed struct {
    name   string
    group  Group
    order  int
    err    error
    called int
}

func (s *testSeed) Name() string { return s.name }
func (s *testSeed) Group() Group { return s.group }
func (s *testSeed) Order() int   { return s.order }
func (s *testSeed) Run(context.Context, DBTX) error {
    s.called++
    return s.err
}

func TestRegistryOrdersAndRejectsDuplicates(t *testing.T) {
    first := &testSeed{name: "second", group: Reference, order: 20}
    second := &testSeed{name: "first", group: Reference, order: 10}
    registry, err := NewRegistry(first, second)
    if err != nil {
        t.Fatalf("NewRegistry() error = %v", err)
    }
    all := registry.All()
    if all[0].Name() != "first" || all[1].Name() != "second" {
        t.Fatalf("ordered seeds = %q, %q", all[0].Name(), all[1].Name())
    }
    if err := registry.Register(&testSeed{name: "first", group: Reference}); err == nil {
        t.Fatal("Register() accepted a duplicate seed")
    }
    if _, err := registry.Find("missing"); err == nil {
        t.Fatal("Find() accepted an unknown seed")
    }
}

func TestRunnerAppliesAndTracksSeed(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    seed := &testSeed{name: "reference.example", group: Reference}
    mock.ExpectBegin()
    mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock($1)`)).WithArgs(seedLockKey).WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT executed_at FROM schema_seeds WHERE name = $1`)).WithArgs(seed.Name()).WillReturnError(sql.ErrNoRows)
    mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO schema_seeds (name) VALUES ($1)`)).WithArgs(seed.Name()).WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectCommit()

    results, err := (Runner{DB: db}).Run(context.Background(), []Seeder{seed}, config.Production)
    if err != nil {
        t.Fatalf("Run() error = %v", err)
    }
    if len(results) != 1 || !results[0].Applied || seed.called != 1 {
        t.Fatalf("results = %#v, called = %d", results, seed.called)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func TestRunnerSkipsAppliedSeed(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    seed := &testSeed{name: "reference.example", group: Reference}
    mock.ExpectBegin()
    mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock($1)`)).WithArgs(seedLockKey).WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT executed_at FROM schema_seeds WHERE name = $1`)).WithArgs(seed.Name()).WillReturnRows(sqlmock.NewRows([]string{"executed_at"}).AddRow(time.Now()))
    mock.ExpectCommit()

    results, err := (Runner{DB: db}).Run(context.Background(), []Seeder{seed}, config.Production)
    if err != nil {
        t.Fatalf("Run() error = %v", err)
    }
    if len(results) != 1 || results[0].Applied || seed.called != 0 {
        t.Fatalf("results = %#v, called = %d", results, seed.called)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func TestRunnerRollsBackFailedSeed(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    seed := &testSeed{name: "reference.example", group: Reference, err: errors.New("boom")}
    mock.ExpectBegin()
    mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock($1)`)).WithArgs(seedLockKey).WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectQuery(regexp.QuoteMeta(`SELECT executed_at FROM schema_seeds WHERE name = $1`)).WithArgs(seed.Name()).WillReturnError(sql.ErrNoRows)
    mock.ExpectRollback()

    if _, err := (Runner{DB: db}).Run(context.Background(), []Seeder{seed}, config.Production); err == nil {
        t.Fatal("Run() accepted a failed seed")
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func TestRunnerBlocksDevelopmentSeedsOutsideSafeEnvironments(t *testing.T) {
    seed := &testSeed{name: "development.example", group: Development}
    if _, err := (Runner{}).Run(context.Background(), []Seeder{seed}, config.Production); err == nil {
        t.Fatal("Run() allowed a development seed in production")
    }
}
