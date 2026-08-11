package cli

import (
    "context"
    "database/sql"
    "errors"
    "flag"
    "fmt"
    "io"
    "strings"

    "github.com/golang-migrate/migrate/v4"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    "github.com/Kangditya/persona-apps/apps/api/internal/database/migration"
    "github.com/Kangditya/persona-apps/apps/api/internal/database/seeder"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database"
)

func Run(ctx context.Context, args []string, out, errOut io.Writer) error {
    if out == nil {
        out = io.Discard
    }
    if errOut == nil {
        errOut = io.Discard
    }
    if len(args) == 0 {
        return errors.New(usage())
    }

    switch args[0] {
    case "migrate":
        return runMigrate(ctx, args[1:], out, errOut)
    case "seed":
        return runSeed(ctx, args[1:], out, errOut)
    case "setup":
        if len(args) != 1 {
            return errors.New("usage: db setup")
        }
        return runSetup(ctx, out)
    case "help", "--help", "-h":
        _, _ = fmt.Fprintln(out, usage())
        return nil
    default:
        return fmt.Errorf("unknown command %q\n%s", args[0], usage())
    }
}

func runMigrate(ctx context.Context, args []string, out, errOut io.Writer) error {
    if len(args) == 0 {
        return errors.New("usage: db migrate {validate|status|version|up|down|create}")
    }

    switch args[0] {
    case "validate":
        if err := rejectExtraFlags("migrate validate", args[1:], errOut); err != nil {
            return err
        }
        files, err := migration.Discover(migration.ResolveDirectory())
        if err != nil {
            return fmt.Errorf("validate migrations: %w", err)
        }
        _, _ = fmt.Fprintf(out, "valid: %d migrations in %s\n", len(files), migration.ResolveDirectory())
        return nil
    case "status":
        if err := rejectExtraFlags("migrate status", args[1:], errOut); err != nil {
            return err
        }
        return runMigrationStatus(ctx, out)
    case "version":
        if err := rejectExtraFlags("migrate version", args[1:], errOut); err != nil {
            return err
        }
        return runMigrationVersion(ctx, out)
    case "up":
        steps, err := parseSteps("migrate up", args[1:], -1, errOut)
        if err != nil {
            return err
        }
        return runMigrationUp(ctx, steps, out)
    case "down":
        steps, err := parseSteps("migrate down", args[1:], 0, errOut)
        if err != nil {
            return err
        }
        return runMigrationDown(ctx, steps, out)
    case "create":
        if len(args) != 2 {
            return errors.New("usage: db migrate create NAME")
        }
        created, err := migration.Create(migration.ResolveDirectory(), args[1])
        if err != nil {
            return fmt.Errorf("create migration: %w", err)
        }
        _, _ = fmt.Fprintf(out, "created %s and %s\n", created.UpPath, created.DownPath)
        return nil
    default:
        return fmt.Errorf("unknown migration command %q", args[0])
    }
}

func runMigrationStatus(ctx context.Context, out io.Writer) error {
    files, migrator, db, err := openMigrator(ctx)
    if err != nil {
        return err
    }
    defer db.Close()
    defer migrator.Close()

    state, err := migration.Current(migrator)
    if err != nil {
        return err
    }
    _, _ = fmt.Fprintln(out, "version\tname\tstate")
    for _, file := range files {
        status := "pending"
        if state.HasVersion && file.Version < state.Version {
            status = "applied"
        }
        if state.HasVersion && file.Version == state.Version {
            status = "applied"
            if state.Dirty {
                status = "dirty"
            }
        }
        _, _ = fmt.Fprintf(out, "%04d\t%s\t%s\n", file.Version, file.Name, status)
    }
    return nil
}

func runMigrationVersion(ctx context.Context, out io.Writer) error {
    _, migrator, db, err := openMigrator(ctx)
    if err != nil {
        return err
    }
    defer db.Close()
    defer migrator.Close()

    state, err := migration.Current(migrator)
    if err != nil {
        return err
    }
    if !state.HasVersion {
        _, _ = fmt.Fprintln(out, "version: none\ndirty: false")
        return nil
    }
    _, _ = fmt.Fprintf(out, "version: %d\ndirty: %t\n", state.Version, state.Dirty)
    return nil
}

func runMigrationUp(ctx context.Context, steps int, out io.Writer) error {
    _, migrator, db, err := openMigrator(ctx)
    if err != nil {
        return err
    }
    defer db.Close()
    defer migrator.Close()

    if steps > 0 {
        err = migrator.Steps(steps)
    } else {
        err = migrator.Up()
    }
    if errors.Is(err, migrate.ErrNoChange) {
        _, _ = fmt.Fprintln(out, "no pending migrations")
        return nil
    }
    if err != nil {
        return fmt.Errorf("apply migrations: %w", err)
    }
    if steps > 0 {
        _, _ = fmt.Fprintf(out, "applied up to %d migration step(s)\n", steps)
    } else {
        _, _ = fmt.Fprintln(out, "migrations applied")
    }
    return nil
}

func runMigrationDown(ctx context.Context, steps int, out io.Writer) error {
    configuration, err := loadConfig()
    if err != nil {
        return err
    }
    if !configuration.Environment.AllowsRollback(configuration.AllowDestructiveCommands) {
        return errors.New("rollback is blocked outside development/test; set ALLOW_DESTRUCTIVE_DB_COMMANDS=true for an explicit staging/production operation")
    }

    _, migrator, db, err := openMigrator(ctx)
    if err != nil {
        return err
    }
    defer db.Close()
    defer migrator.Close()

    if err := migrator.Steps(-steps); err != nil {
        return fmt.Errorf("rollback %d migration step(s): %w", steps, err)
    }
    _, _ = fmt.Fprintf(out, "rolled back %d migration step(s)\n", steps)
    return nil
}

func runSeed(ctx context.Context, args []string, out, errOut io.Writer) error {
    if len(args) == 0 {
        return errors.New("usage: db seed {list|status|run}")
    }
    registry, err := seeder.DefaultRegistry()
    if err != nil {
        return fmt.Errorf("load seed registry: %w", err)
    }

    switch args[0] {
    case "list":
        if err := rejectExtraFlags("seed list", args[1:], errOut); err != nil {
            return err
        }
        _, _ = fmt.Fprintln(out, "name\tgroup\torder")
        for _, seed := range registry.All() {
            _, _ = fmt.Fprintf(out, "%s\t%s\t%d\n", seed.Name(), seed.Group(), seed.Order())
        }
        return nil
    case "status":
        if err := rejectExtraFlags("seed status", args[1:], errOut); err != nil {
            return err
        }
        return runSeedStatus(ctx, registry, out)
    case "run":
        return runSelectedSeeds(ctx, args[1:], registry, out, errOut)
    default:
        return fmt.Errorf("unknown seed command %q", args[0])
    }
}

func runSeedStatus(ctx context.Context, registry *seeder.Registry, out io.Writer) error {
    configuration, db, err := openDatabase(ctx)
    if err != nil {
        return err
    }
    defer db.Close()

    status, err := (seeder.Runner{DB: db, Registry: registry}).Status(ctx)
    if err != nil {
        return fmt.Errorf("seed status: %w", err)
    }
    _, _ = fmt.Fprintf(out, "environment: %s\n", configuration.Environment)
    _, _ = fmt.Fprintln(out, "name\tgroup\tstate\texecuted_at")
    for _, item := range status {
        state := "pending"
        executedAt := "-"
        if item.Applied {
            state = "applied"
            executedAt = item.ExecutedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
        }
        _, _ = fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", item.Seeder.Name(), item.Seeder.Group(), state, executedAt)
    }
    return nil
}

func runSelectedSeeds(ctx context.Context, args []string, registry *seeder.Registry, out, errOut io.Writer) error {
    flags := newFlagSet("seed run", errOut)
    all := flags.Bool("all", false, "run all eligible seeds")
    group := flags.String("group", "", "run a seed group")
    if err := flags.Parse(args); err != nil {
        return err
    }
    if *all && *group != "" || *all && len(flags.Args()) > 0 || *group != "" && len(flags.Args()) > 0 {
        return errors.New("choose exactly one of a seed name, --all, or --group GROUP")
    }

    configuration, err := loadConfig()
    if err != nil {
        return err
    }
    var seeds []seeder.Seeder
    switch {
    case *all:
        seeds = registry.All()
        if !configuration.Environment.AllowsDevelopmentSeeds() {
            seeds = registry.Group(seeder.Reference)
        }
    case *group != "":
        if *group != string(seeder.Reference) && *group != string(seeder.Development) {
            return errors.New("seed group must be reference or development")
        }
        if *group == string(seeder.Development) && !configuration.Environment.AllowsDevelopmentSeeds() {
            return errors.New("development seeds are blocked outside development/test")
        }
        seeds = registry.Group(seeder.Group(*group))
    default:
        if len(flags.Args()) != 1 {
            return errors.New("usage: db seed run NAME | --all | --group GROUP")
        }
        seed, err := registry.Find(flags.Args()[0])
        if err != nil {
            return err
        }
        if seed.Group() == seeder.Development && !configuration.Environment.AllowsDevelopmentSeeds() {
            return errors.New("development seeds are blocked outside development/test")
        }
        seeds = []seeder.Seeder{seed}
    }
    if len(seeds) == 0 {
        _, _ = fmt.Fprintln(out, "no seeds selected")
        return nil
    }

    db, err := openDatabaseWithConfig(ctx, configuration)
    if err != nil {
        return err
    }
    defer db.Close()
    results, err := (seeder.Runner{DB: db, Registry: registry}).Run(ctx, seeds, configuration.Environment)
    for _, result := range results {
        state := "skipped"
        if result.Applied {
            state = "applied"
        }
        _, _ = fmt.Fprintf(out, "seed %s: %s\n", result.Name, state)
    }
    if err != nil {
        return err
    }
    return nil
}

func runSetup(ctx context.Context, out io.Writer) error {
    directory := migration.ResolveDirectory()
    files, err := migration.Discover(directory)
    if err != nil {
        return fmt.Errorf("validate migrations: %w", err)
    }
    configuration, db, err := openDatabase(ctx)
    if err != nil {
        return err
    }
    defer db.Close()
    migrator, err := migration.New(db, directory)
    if err != nil {
        return fmt.Errorf("open migrations: %w", err)
    }
    defer migrator.Close()
    if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
        return fmt.Errorf("setup migrations (%d discovered): %w", len(files), err)
    }

    registry, err := seeder.DefaultRegistry()
    if err != nil {
        return fmt.Errorf("load seed registry: %w", err)
    }
    seeds := registry.Group(seeder.Reference)
    if len(seeds) == 0 {
        _, _ = fmt.Fprintln(out, "migrations applied; no reference seeds registered")
        return nil
    }
    results, err := (seeder.Runner{DB: db, Registry: registry}).Run(ctx, seeds, configuration.Environment)
    for _, result := range results {
        state := "skipped"
        if result.Applied {
            state = "applied"
        }
        _, _ = fmt.Fprintf(out, "seed %s: %s\n", result.Name, state)
    }
    if err != nil {
        return fmt.Errorf("setup reference seeds: %w", err)
    }
    return nil
}

func openMigrator(ctx context.Context) ([]migration.File, *migrate.Migrate, *sql.DB, error) {
    directory := migration.ResolveDirectory()
    files, err := migration.Discover(directory)
    if err != nil {
        return nil, nil, nil, fmt.Errorf("validate migrations: %w", err)
    }
    _, db, err := openDatabase(ctx)
    if err != nil {
        return nil, nil, nil, err
    }
    migrator, err := migration.New(db, directory)
    if err != nil {
        db.Close()
        return nil, nil, nil, fmt.Errorf("open migrations: %w", err)
    }
    return files, migrator, db, nil
}

func openDatabase(ctx context.Context) (config.Config, *sql.DB, error) {
    configuration, err := loadConfig()
    if err != nil {
        return config.Config{}, nil, err
    }
    db, err := openDatabaseWithConfig(ctx, configuration)
    if err != nil {
        return config.Config{}, nil, err
    }
    return configuration, db, nil
}

func openDatabaseWithConfig(ctx context.Context, configuration config.Config) (*sql.DB, error) {
    db, err := database.Open(configuration.DatabaseURL)
    if err != nil {
        return nil, err
    }
    if err := db.PingContext(ctx); err != nil {
        db.Close()
        return nil, errors.New("connect to database: database is unavailable")
    }
    return db, nil
}

func loadConfig() (config.Config, error) {
    configuration, err := config.Load()
    if err != nil {
        return config.Config{}, fmt.Errorf("load database configuration: %w", err)
    }
    return configuration, nil
}

func parseSteps(command string, args []string, defaultValue int, errOut io.Writer) (int, error) {
    flags := newFlagSet(command, errOut)
    steps := flags.Int("steps", defaultValue, "number of migration steps")
    if err := flags.Parse(args); err != nil {
        return 0, err
    }
    if len(flags.Args()) > 0 {
        return 0, fmt.Errorf("usage: db %s --steps N", command)
    }
    if command == "migrate up" && *steps < -1 || command == "migrate down" && *steps <= 0 {
        return 0, errors.New("migration steps must be a positive integer")
    }
    return *steps, nil
}

func rejectExtraFlags(command string, args []string, errOut io.Writer) error {
    flags := newFlagSet(command, errOut)
    if err := flags.Parse(args); err != nil {
        return err
    }
    if len(flags.Args()) > 0 {
        return fmt.Errorf("usage: db %s", command)
    }
    return nil
}

func newFlagSet(name string, errOut io.Writer) *flag.FlagSet {
    flags := flag.NewFlagSet(name, flag.ContinueOnError)
    flags.SetOutput(errOut)
    return flags
}

func usage() string {
    return strings.TrimSpace(`usage: db <command>

commands:
  migrate validate
  migrate status
  migrate version
  migrate up [--steps N]
  migrate down --steps N
  migrate create NAME
  seed list
  seed status
  seed run NAME|--all|--group GROUP
  setup`)
}
