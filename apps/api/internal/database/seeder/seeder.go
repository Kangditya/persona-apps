package seeder

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Kangditya/persona-apps/apps/api/internal/config"
)

type Group string

const (
	Reference   Group = "reference"
	Development Group = "development"
)

type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Seeder interface {
	Name() string
	Group() Group
	Order() int
	Run(context.Context, DBTX) error
}

type Registry struct {
	seeds map[string]Seeder
}

func NewRegistry(seeds ...Seeder) (*Registry, error) {
	registry := &Registry{seeds: make(map[string]Seeder)}
	for _, seed := range seeds {
		if err := registry.Register(seed); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *Registry) Register(seed Seeder) error {
	if seed == nil {
		return errors.New("cannot register a nil seed")
	}
	if seed.Name() == "" {
		return errors.New("seed name is required")
	}
	if seed.Group() != Reference && seed.Group() != Development {
		return fmt.Errorf("seed %q has unsupported group %q", seed.Name(), seed.Group())
	}
	if _, exists := r.seeds[seed.Name()]; exists {
		return fmt.Errorf("duplicate seed name %q", seed.Name())
	}
	r.seeds[seed.Name()] = seed
	return nil
}

func (r *Registry) All() []Seeder {
	seeds := make([]Seeder, 0, len(r.seeds))
	for _, seed := range r.seeds {
		seeds = append(seeds, seed)
	}
	sortSeeds(seeds)
	return seeds
}

func (r *Registry) Find(name string) (Seeder, error) {
	seed, ok := r.seeds[name]
	if !ok {
		return nil, fmt.Errorf("unknown seed %q", name)
	}
	return seed, nil
}

func (r *Registry) Group(group Group) []Seeder {
	seeds := make([]Seeder, 0)
	for _, seed := range r.seeds {
		if seed.Group() == group {
			seeds = append(seeds, seed)
		}
	}
	sortSeeds(seeds)
	return seeds
}

func sortSeeds(seeds []Seeder) {
	sort.Slice(seeds, func(i, j int) bool {
		if seeds[i].Order() != seeds[j].Order() {
			return seeds[i].Order() < seeds[j].Order()
		}
		return seeds[i].Name() < seeds[j].Name()
	})
}

type Result struct {
	Name    string
	Applied bool
}

type Status struct {
	Seeder      Seeder
	Applied     bool
	ExecutedAt  time.Time
	HasExecuted bool
}

type Runner struct {
	DB       *sql.DB
	Registry *Registry
}

const seedLockKey int64 = 781245913

func (r Runner) Run(ctx context.Context, seeds []Seeder, environment config.Environment) ([]Result, error) {
	for _, seed := range seeds {
		if seed.Group() == Development && !environment.AllowsDevelopmentSeeds() {
			return nil, fmt.Errorf("development seed %q is only allowed in development or test", seed.Name())
		}
	}
	if r.DB == nil {
		return nil, errors.New("database is required")
	}

	results := make([]Result, 0, len(seeds))
	for _, seed := range seeds {
		result, err := r.runOne(ctx, seed)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (r Runner) runOne(ctx context.Context, seed Seeder) (Result, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, fmt.Errorf("begin seed %q: %w", seed.Name(), err)
	}
	rollback := func() { _ = tx.Rollback() }

	// ponytail: one advisory lock serializes seed runs; use per-seed locks if throughput matters.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, seedLockKey); err != nil {
		rollback()
		return Result{}, fmt.Errorf("lock seed execution: %w", err)
	}

	var executedAt time.Time
	err = tx.QueryRowContext(ctx, `SELECT executed_at FROM schema_seeds WHERE name = $1`, seed.Name()).Scan(&executedAt)
	switch {
	case err == nil:
		if err := tx.Commit(); err != nil {
			return Result{}, fmt.Errorf("commit skipped seed %q: %w", seed.Name(), err)
		}
		return Result{Name: seed.Name()}, nil
	case !errors.Is(err, sql.ErrNoRows):
		rollback()
		return Result{}, fmt.Errorf("read seed history for %q: %w", seed.Name(), err)
	}

	if err := seed.Run(ctx, tx); err != nil {
		rollback()
		return Result{}, fmt.Errorf("run seed %q: %w", seed.Name(), err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_seeds (name) VALUES ($1)`, seed.Name()); err != nil {
		rollback()
		return Result{}, fmt.Errorf("record seed %q: %w", seed.Name(), err)
	}
	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit seed %q: %w", seed.Name(), err)
	}
	return Result{Name: seed.Name(), Applied: true}, nil
}

func (r Runner) Status(ctx context.Context) ([]Status, error) {
	if r.DB == nil {
		return nil, errors.New("database is required")
	}
	if r.Registry == nil {
		return nil, errors.New("seed registry is required")
	}

	rows, err := r.DB.QueryContext(ctx, `SELECT name, executed_at FROM schema_seeds ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("read seed history: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]time.Time)
	for rows.Next() {
		var name string
		var executedAt time.Time
		if err := rows.Scan(&name, &executedAt); err != nil {
			return nil, fmt.Errorf("scan seed history: %w", err)
		}
		applied[name] = executedAt
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read seed history: %w", err)
	}

	status := make([]Status, 0, len(r.Registry.All()))
	for _, seed := range r.Registry.All() {
		executedAt, ok := applied[seed.Name()]
		status = append(status, Status{Seeder: seed, Applied: ok, ExecutedAt: executedAt, HasExecuted: ok})
	}
	return status, nil
}

func DefaultRegistry() (*Registry, error) {
	return NewRegistry(sampleEvent{})
}
