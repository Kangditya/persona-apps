package config

import "testing"

func TestLoadRequiresAndValidatesEnvironment(t *testing.T) {
    t.Setenv("DATABASE_URL", "postgres://example.invalid/db")
    t.Setenv("APP_ENV", "development")
    t.Setenv("ALLOW_DESTRUCTIVE_DB_COMMANDS", "false")

    configuration, err := Load()
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }
    if configuration.Environment != Development {
        t.Fatalf("environment = %q, want %q", configuration.Environment, Development)
    }

    t.Setenv("APP_ENV", "productionish")
    if _, err := Load(); err == nil {
        t.Fatal("Load() accepted an unsupported environment")
    }
}

func TestEnvironmentSafety(t *testing.T) {
    if !Development.AllowsDevelopmentSeeds() || !Test.AllowsDevelopmentSeeds() {
        t.Fatal("development and test should allow development seeds")
    }
    if Production.AllowsDevelopmentSeeds() || Staging.AllowsDevelopmentSeeds() {
        t.Fatal("staging and production should reject development seeds")
    }
    if Production.AllowsRollback(false) || !Production.AllowsRollback(true) {
        t.Fatal("production rollback should require explicit opt-in")
    }
}

func TestLoadAPIFailsClosedForPartialAuthenticationConfiguration(t *testing.T) {
    t.Setenv("DATABASE_URL", "postgres://example.invalid/db")
    t.Setenv("APP_ENV", "development")
    t.Setenv("OIDC_ISSUER_URL", "https://issuer.example.invalid")
    t.Setenv("OIDC_CLIENT_ID", "")
    if _, err := LoadAPI(); err == nil {
        t.Fatal("LoadAPI() accepted partial authentication configuration")
    }
}

func TestLoadAPIRejectsDestructiveStagingProcess(t *testing.T) {
    t.Setenv("DATABASE_URL", "postgres://example.invalid/db")
    t.Setenv("APP_ENV", "staging")
    t.Setenv("ALLOW_DESTRUCTIVE_DB_COMMANDS", "true")
    if _, err := LoadAPI(); err == nil {
        t.Fatal("LoadAPI() accepted destructive staging configuration")
    }
}
