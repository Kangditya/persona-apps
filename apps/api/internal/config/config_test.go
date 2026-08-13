package config

import (
    "reflect"
    "testing"
)

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

func TestLoadAPIUsesSafePublicDefaults(t *testing.T) {
    t.Setenv("DATABASE_URL", "postgres://example.invalid/db")
    t.Setenv("APP_ENV", "development")
    t.Setenv("PUBLIC_RATE_LIMIT_PER_MINUTE", "")
    t.Setenv("PUBLIC_RATE_LIMIT_BURST", "")
    t.Setenv("TRUSTED_PROXY_CIDRS", "")

    configuration, err := LoadAPI()
    if err != nil {
        t.Fatal(err)
    }
    if configuration.Public.RateLimitPerMinute != 60 || configuration.Public.RateLimitBurst != 20 {
        t.Fatalf("public defaults = %#v", configuration.Public)
    }
    if configuration.Public.TrustedProxyCIDRs != nil {
        t.Fatalf("trusted proxies = %#v, want disabled", configuration.Public.TrustedProxyCIDRs)
    }
}

func TestLoadAPIValidatesPublicNetworkConfiguration(t *testing.T) {
    tests := []struct {
        name  string
        name2 string
        value string
    }{
        {name: "zero rate", name2: "PUBLIC_RATE_LIMIT_PER_MINUTE", value: "0"},
        {name: "invalid burst", name2: "PUBLIC_RATE_LIMIT_BURST", value: "many"},
        {name: "raw address is not CIDR", name2: "TRUSTED_PROXY_CIDRS", value: "192.0.2.1"},
        {name: "global CIDR is forbidden", name2: "TRUSTED_PROXY_CIDRS", value: "0.0.0.0/0"},
        {name: "host bits must be zero", name2: "TRUSTED_PROXY_CIDRS", value: "192.0.2.1/24"},
    }
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            t.Setenv("DATABASE_URL", "postgres://example.invalid/db")
            t.Setenv("APP_ENV", "development")
            t.Setenv("PUBLIC_RATE_LIMIT_PER_MINUTE", "")
            t.Setenv("PUBLIC_RATE_LIMIT_BURST", "")
            t.Setenv("TRUSTED_PROXY_CIDRS", "")
            t.Setenv(test.name2, test.value)

            if _, err := LoadAPI(); err == nil {
                t.Fatal("LoadAPI accepted unsafe public configuration")
            }
        })
    }
}

func TestLoadAPINormalizesTrustedProxyCIDRs(t *testing.T) {
    t.Setenv("DATABASE_URL", "postgres://example.invalid/db")
    t.Setenv("APP_ENV", "development")
    t.Setenv("TRUSTED_PROXY_CIDRS", "192.0.2.0/24, 2001:db8::/32, 192.0.2.0/24")

    configuration, err := LoadAPI()
    if err != nil {
        t.Fatal(err)
    }
    want := []string{"192.0.2.0/24", "2001:db8::/32"}
    if !reflect.DeepEqual(configuration.Public.TrustedProxyCIDRs, want) {
        t.Fatalf("trusted proxies = %#v, want %#v", configuration.Public.TrustedProxyCIDRs, want)
    }
}
