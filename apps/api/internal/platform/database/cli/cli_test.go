package cli

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/database/migration"
)

func TestRunValidatesMigrationsAndListsSeeds(t *testing.T) {
    directory, err := filepath.Abs("../../../../migrations")
    if err != nil {
        t.Fatal(err)
    }
    t.Setenv("MIGRATIONS_DIR", directory)

    migrations, err := migration.Discover(directory)
    if err != nil {
        t.Fatal(err)
    }

    var output bytes.Buffer
    if err := Run(context.Background(), []string{"migrate", "validate"}, &output, &output); err != nil {
        t.Fatalf("validate error = %v", err)
    }
    if !strings.Contains(output.String(), fmt.Sprintf("valid: %d migrations", len(migrations))) {
        t.Fatalf("validate output = %q", output.String())
    }

    output.Reset()
    if err := Run(context.Background(), []string{"seed", "list"}, &output, &output); err != nil {
        t.Fatalf("seed list error = %v", err)
    }
    if !strings.Contains(output.String(), "development.sample-event") {
        t.Fatalf("seed list output = %q", output.String())
    }
}

func TestRunRejectsInvalidStepsAndUnsafeRollback(t *testing.T) {
    var output bytes.Buffer
    if err := Run(context.Background(), []string{"migrate", "up", "--steps", "0"}, &output, &output); err == nil {
        t.Fatal("up accepted zero steps")
    }

    t.Setenv("DATABASE_URL", "postgres://example.invalid/db")
    t.Setenv("APP_ENV", "production")
    t.Setenv("ALLOW_DESTRUCTIVE_DB_COMMANDS", "false")
    if err := Run(context.Background(), []string{"migrate", "down", "--steps", "1"}, &output, &output); err == nil || !strings.Contains(err.Error(), "rollback is blocked") {
        t.Fatalf("unsafe rollback error = %v", err)
    }
}

func TestRunMissingArgumentsReturnsUsage(t *testing.T) {
    if err := Run(context.Background(), nil, os.Stdout, os.Stderr); err == nil || !strings.Contains(err.Error(), "usage: db") {
        t.Fatalf("missing argument error = %v", err)
    }
}
