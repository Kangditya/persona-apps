package config

import (
    "encoding/base64"
    "errors"
    "fmt"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/platform/idempotency"
)

type Environment string

const (
    Development Environment = "development"
    Test        Environment = "test"
    Staging     Environment = "staging"
    Production  Environment = "production"
)

type Config struct {
    HTTPAddress              string
    DatabaseURL              string
    Environment              Environment
    AllowDestructiveCommands bool
}

type APIConfig struct {
    Config
    Auth *AuthConfig
}

type AuthConfig struct {
    IssuerURL              string
    ClientID               string
    ClientSecret           string
    RedirectURL            string
    PermissionClaim        string
    OperationsWebOrigin    string
    AllowedOrigins         map[string]struct{}
    CookieEncryptionKey    []byte
    ResponseEncryptionKeys []idempotency.Key
    MaxSessionLifetime     time.Duration
}

func Load() (Config, error) {
    address := os.Getenv("HTTP_ADDRESS")
    if address == "" {
        address = ":8080"
    }
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        return Config{}, errors.New("DATABASE_URL is required")
    }
    runtimeEnvironment := Environment(os.Getenv("APP_ENV"))
    if runtimeEnvironment == "" {
        return Config{}, errors.New("APP_ENV is required (development, test, staging, or production)")
    }
    if !isValidEnvironment(runtimeEnvironment) {
        return Config{}, errors.New("APP_ENV must be development, test, staging, or production")
    }
    allowDestructiveCommands := false
    if raw := os.Getenv("ALLOW_DESTRUCTIVE_DB_COMMANDS"); raw != "" {
        var err error
        allowDestructiveCommands, err = strconv.ParseBool(raw)
        if err != nil {
            return Config{}, errors.New("ALLOW_DESTRUCTIVE_DB_COMMANDS must be a boolean")
        }
    }
    return Config{HTTPAddress: address, DatabaseURL: databaseURL, Environment: runtimeEnvironment, AllowDestructiveCommands: allowDestructiveCommands}, nil
}

func LoadAPI() (APIConfig, error) {
    base, err := Load()
    if err != nil {
        return APIConfig{}, err
    }
    if (base.Environment == Staging || base.Environment == Production) && base.AllowDestructiveCommands {
        return APIConfig{}, errors.New("ALLOW_DESTRUCTIVE_DB_COMMANDS must be false for the long-running staging or production API")
    }
    auth, configured, err := loadAuth(base.Environment)
    if err != nil {
        return APIConfig{}, err
    }
    if !configured && (base.Environment == Staging || base.Environment == Production) {
        return APIConfig{}, errors.New("OIDC and response-encryption configuration is required outside development and test")
    }
    return APIConfig{Config: base, Auth: auth}, nil
}

func loadAuth(environment Environment) (*AuthConfig, bool, error) {
    required := []string{"OIDC_ISSUER_URL", "OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET", "OIDC_REDIRECT_URL", "OIDC_PERMISSION_CLAIM", "OPERATIONS_WEB_ORIGIN", "OPERATIONS_ALLOWED_ORIGINS", "AUTH_COOKIE_ENCRYPTION_KEY", "IDEMPOTENCY_RESPONSE_KEYS"}
    found := false
    for _, name := range required {
        found = found || os.Getenv(name) != ""
    }
    if !found {
        return nil, false, nil
    }
    for _, name := range required {
        if os.Getenv(name) == "" {
            return nil, false, fmt.Errorf("%s is required when Operations authentication is configured", name)
        }
    }
    issuer := os.Getenv("OIDC_ISSUER_URL")
    redirect := os.Getenv("OIDC_REDIRECT_URL")
    origin := os.Getenv("OPERATIONS_WEB_ORIGIN")
    if err := validateURL("OIDC_ISSUER_URL", issuer, environment); err != nil {
        return nil, false, err
    }
    if err := validateURL("OIDC_REDIRECT_URL", redirect, environment); err != nil {
        return nil, false, err
    }
    if err := validateOrigin("OPERATIONS_WEB_ORIGIN", origin, environment); err != nil {
        return nil, false, err
    }
    cookieKey, err := decode32("AUTH_COOKIE_ENCRYPTION_KEY", os.Getenv("AUTH_COOKIE_ENCRYPTION_KEY"))
    if err != nil {
        return nil, false, err
    }
    keys, err := parseKeyRing(os.Getenv("IDEMPOTENCY_RESPONSE_KEYS"))
    if err != nil {
        return nil, false, err
    }
    origins := map[string]struct{}{}
    for _, value := range strings.Split(os.Getenv("OPERATIONS_ALLOWED_ORIGINS"), ",") {
        value = strings.TrimSpace(value)
        if err := validateOrigin("OPERATIONS_ALLOWED_ORIGINS", value, environment); err != nil {
            return nil, false, err
        }
        origins[value] = struct{}{}
    }
    if _, allowed := origins[origin]; !allowed {
        return nil, false, errors.New("OPERATIONS_WEB_ORIGIN must be included in OPERATIONS_ALLOWED_ORIGINS")
    }
    maxLifetime := 8 * time.Hour
    if raw := os.Getenv("OPERATIONS_SESSION_MAX_LIFETIME"); raw != "" {
        maxLifetime, err = time.ParseDuration(raw)
        if err != nil || maxLifetime <= 0 || maxLifetime > 24*time.Hour {
            return nil, false, errors.New("OPERATIONS_SESSION_MAX_LIFETIME must be between zero and 24h")
        }
    }
    return &AuthConfig{IssuerURL: issuer, ClientID: os.Getenv("OIDC_CLIENT_ID"), ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"), RedirectURL: redirect, PermissionClaim: os.Getenv("OIDC_PERMISSION_CLAIM"), OperationsWebOrigin: origin, AllowedOrigins: origins, CookieEncryptionKey: cookieKey, ResponseEncryptionKeys: keys, MaxSessionLifetime: maxLifetime}, true, nil
}

func decode32(name, value string) ([]byte, error) {
    decoded, err := base64.RawStdEncoding.DecodeString(value)
    if err != nil {
        decoded, err = base64.StdEncoding.DecodeString(value)
    }
    if err != nil || len(decoded) != 32 {
        return nil, fmt.Errorf("%s must be base64 encoded 32 bytes", name)
    }
    return decoded, nil
}

func parseKeyRing(value string) ([]idempotency.Key, error) {
    var keys []idempotency.Key
    for _, part := range strings.Split(value, ",") {
        id, encoded, found := strings.Cut(strings.TrimSpace(part), ":")
        if !found || id == "" || strings.ContainsAny(id, " \t\n") {
            return nil, errors.New("IDEMPOTENCY_RESPONSE_KEYS must contain key-id:base64-32-byte-key values")
        }
        key, err := decode32("IDEMPOTENCY_RESPONSE_KEYS", encoded)
        if err != nil {
            return nil, err
        }
        keys = append(keys, idempotency.Key{ID: id, Value: key})
    }
    if _, err := idempotency.NewCipher(keys); err != nil {
        return nil, err
    }
    return keys, nil
}

func validateURL(name, value string, environment Environment) error {
    parsed, err := url.Parse(value)
    if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
        return fmt.Errorf("%s must be an absolute origin URL", name)
    }
    if parsed.Scheme != "https" && !(environment == Development || environment == Test) {
        return fmt.Errorf("%s must use HTTPS outside development and test", name)
    }
    return nil
}

func validateOrigin(name, value string, environment Environment) error {
    if err := validateURL(name, value, environment); err != nil {
        return err
    }
    parsed, _ := url.Parse(value)
    if parsed.Path != "" && parsed.Path != "/" {
        return fmt.Errorf("%s must not include a path", name)
    }
    return nil
}

func isValidEnvironment(value Environment) bool {
    switch value {
    case Development, Test, Staging, Production:
        return true
    default:
        return false
    }
}

func (e Environment) AllowsDevelopmentSeeds() bool {
    return e == Development || e == Test
}

func (e Environment) AllowsRollback(explicit bool) bool {
    return e == Development || e == Test || explicit
}
