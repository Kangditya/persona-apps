package auth

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "crypto/subtle"
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "sort"
    "strings"
    "time"

    "github.com/Kangditya/persona-apps/apps/api/internal/config"
    "github.com/Kangditya/persona-apps/apps/api/internal/platform/httpx"
    "github.com/coreos/go-oidc/v3/oidc"
    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
)

const (
    loginCookieName   = "__Host-operations_login"
    sessionCookieName = "__Host-operations_session"
)

var permissionAllowlist = map[string]struct{}{
    "event.read": {}, "event.manage": {}, "offering.read": {}, "offering.manage": {},
    "purchase.read": {}, "payment.read": {}, "payment.verify": {}, "participant.read": {},
    "dashboard.read": {}, "audit.read": {}, "admin.manage": {},
}

type Principal struct {
    OperatorID  string
    DisplayName string
    Permissions map[string]struct{}
    ExpiresAt   time.Time
}

type principalContextKey struct{}

func FromContext(ctx context.Context) (Principal, bool) {
    value, ok := ctx.Value(principalContextKey{}).(Principal)
    return value, ok
}

func (s *Service) AllowedOrigins() map[string]struct{} {
    return s.config.AllowedOrigins
}

type Service struct {
    db       *sql.DB
    config   config.AuthConfig
    oauth    oauth2.Config
    verifier *oidc.IDTokenVerifier
}

type loginState struct {
    State     string
    Nonce     string
    Verifier  string
    ReturnTo  string
    ExpiresAt time.Time
}

type operatorUserRow struct {
    ID          string
    DisplayName string
    Status      string
}

type operatorSessionRow struct {
    OperatorID         string
    DisplayName        string
    OperatorStatus     string
    PermissionSnapshot string
    ExpiresAt          time.Time
}

func (row *operatorUserRow) destinations() []any {
    return []any{&row.ID, &row.DisplayName, &row.Status}
}

func (row *operatorSessionRow) destinations() []any {
    return []any{&row.OperatorID, &row.DisplayName, &row.OperatorStatus, &row.PermissionSnapshot, &row.ExpiresAt}
}

func (row operatorSessionRow) principal(now time.Time) (Principal, bool) {
    if row.OperatorStatus != "ACTIVE" || !now.Before(row.ExpiresAt) {
        return Principal{}, false
    }
    var permissions []string
    if json.Unmarshal([]byte(row.PermissionSnapshot), &permissions) != nil {
        return Principal{}, false
    }
    result := Principal{OperatorID: row.OperatorID, DisplayName: row.DisplayName, Permissions: map[string]struct{}{}, ExpiresAt: row.ExpiresAt.UTC()}
    for _, permission := range permissions {
        if _, allowed := permissionAllowlist[permission]; allowed {
            result.Permissions[permission] = struct{}{}
        }
    }
    return result, true
}

func New(ctx context.Context, db *sql.DB, configuration config.AuthConfig) (*Service, error) {
    provider, err := oidc.NewProvider(ctx, configuration.IssuerURL)
    if err != nil {
        return nil, fmt.Errorf("discover OIDC provider: %w", err)
    }
    return &Service{
        db: db, config: configuration,
        oauth:    oauth2.Config{ClientID: configuration.ClientID, ClientSecret: configuration.ClientSecret, RedirectURL: configuration.RedirectURL, Endpoint: provider.Endpoint(), Scopes: []string{oidc.ScopeOpenID}},
        verifier: provider.Verifier(&oidc.Config{ClientID: configuration.ClientID}),
    }, nil
}

func (s *Service) Require(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        r := c.Request
        principal, ok := s.authenticate(r)
        if !ok {
            httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "authentication required", httpx.RequestID(r.Context()), nil)
            c.Abort()
            return
        }
        if isUnsafe(r.Method) && !s.validCSRF(r, principal.OperatorID) {
            httpx.WriteError(c, http.StatusForbidden, "forbidden", "request is not authorized", httpx.RequestID(r.Context()), nil)
            c.Abort()
            return
        }
        if _, allowed := principal.Permissions[permission]; !allowed {
            httpx.WriteError(c, http.StatusForbidden, "forbidden", "request is not authorized", httpx.RequestID(r.Context()), nil)
            c.Abort()
            return
        }
        c.Request = r.WithContext(context.WithValue(r.Context(), principalContextKey{}, principal))
        c.Next()
    }
}

func (s *Service) login(c *gin.Context) {
    r, w := c.Request, c.Writer
    returnTo := r.URL.Query().Get("return_to")
    if returnTo == "" {
        returnTo = "/"
    }
    if !validReturnTo(returnTo) {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "invalid return_to", httpx.RequestID(r.Context()), nil)
        return
    }
    state := loginState{State: randomToken(32), Nonce: randomToken(32), Verifier: oauth2.GenerateVerifier(), ReturnTo: returnTo, ExpiresAt: time.Now().Add(10 * time.Minute)}
    value, err := s.sealLoginState(state)
    if err != nil {
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", httpx.RequestID(r.Context()), nil)
        return
    }
    http.SetCookie(w, &http.Cookie{Name: loginCookieName, Value: value, Path: "/", MaxAge: 600, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
    http.Redirect(w, r, s.oauth.AuthCodeURL(state.State, oidc.Nonce(state.Nonce), oauth2.S256ChallengeOption(state.Verifier)), http.StatusFound)
}

func (s *Service) callback(c *gin.Context) {
    r, w := c.Request, c.Writer
    defer s.clearLoginCookie(w)
    if r.URL.Query().Get("error") != "" {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "identity provider rejected login", httpx.RequestID(r.Context()), nil)
        return
    }
    state, err := s.openLoginState(r)
    if err != nil || time.Now().After(state.ExpiresAt) || subtle.ConstantTimeCompare([]byte(state.State), []byte(r.URL.Query().Get("state"))) != 1 {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "invalid login callback", httpx.RequestID(r.Context()), nil)
        return
    }
    code := r.URL.Query().Get("code")
    if code == "" {
        httpx.WriteError(c, http.StatusBadRequest, "invalid_request", "missing authorization code", httpx.RequestID(r.Context()), nil)
        return
    }
    token, err := s.oauth.Exchange(r.Context(), code, oauth2.VerifierOption(state.Verifier))
    if err != nil {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "invalid login callback", httpx.RequestID(r.Context()), nil)
        return
    }
    rawIDToken, ok := token.Extra("id_token").(string)
    if !ok || rawIDToken == "" {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "identity token is required", httpx.RequestID(r.Context()), nil)
        return
    }
    idToken, err := s.verifier.Verify(r.Context(), rawIDToken)
    if err != nil {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "invalid identity token", httpx.RequestID(r.Context()), nil)
        return
    }
    var claims map[string]json.RawMessage
    if err := idToken.Claims(&claims); err != nil {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "invalid identity token", httpx.RequestID(r.Context()), nil)
        return
    }
    var subject, nonce string
    if json.Unmarshal(claims["sub"], &subject) != nil || json.Unmarshal(claims["nonce"], &nonce) != nil || subject == "" || subtle.ConstantTimeCompare([]byte(nonce), []byte(state.Nonce)) != 1 {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "invalid identity token", httpx.RequestID(r.Context()), nil)
        return
    }
    permissions, err := claimedPermissions(claims[s.config.PermissionClaim])
    if err != nil {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "invalid permission claim", httpx.RequestID(r.Context()), nil)
        return
    }
    operatorID, _, err := s.activeOperator(r.Context(), subject)
    if err != nil {
        httpx.WriteError(c, http.StatusForbidden, "forbidden", "operator is not active", httpx.RequestID(r.Context()), nil)
        return
    }
    expiry := time.Now().Add(s.config.MaxSessionLifetime)
    if idToken.Expiry.Before(expiry) {
        expiry = idToken.Expiry
    }
    sessionToken, csrfToken := randomToken(32), randomToken(32)
    permissionJSON, _ := json.Marshal(permissions)
    if _, err := s.db.ExecContext(r.Context(), "INSERT INTO operator_sessions (operator_user_id, session_token_hash, csrf_token_hash, permission_snapshot, expires_at) VALUES ($1, $2, $3, $4::jsonb, $5)", operatorID, digest(sessionToken), digest(csrfToken), string(permissionJSON), expiry); err != nil {
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", httpx.RequestID(r.Context()), nil)
        return
    }
    http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: sessionToken, Path: "/", Expires: expiry, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
    http.Redirect(w, r, s.config.OperationsWebOrigin+state.ReturnTo, http.StatusFound)
}

func (s *Service) session(c *gin.Context) {
    r := c.Request
    principal, ok := s.authenticate(r)
    if !ok {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "authentication required", httpx.RequestID(r.Context()), nil)
        return
    }
    csrfToken := randomToken(32)
    if _, err := s.db.ExecContext(r.Context(), "UPDATE operator_sessions SET csrf_token_hash = $1, last_seen_at = now() WHERE session_token_hash = $2", digest(csrfToken), digest(s.sessionToken(r))); err != nil {
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", httpx.RequestID(r.Context()), nil)
        return
    }
    permissions := make([]string, 0, len(principal.Permissions))
    for permission := range principal.Permissions {
        permissions = append(permissions, permission)
    }
    sort.Strings(permissions)
    c.Header("Cache-Control", "no-store")
    c.JSON(http.StatusOK, map[string]any{"data": map[string]any{"operator": map[string]string{"id": principal.OperatorID, "display_name": principal.DisplayName}, "permissions": permissions, "expires_at": principal.ExpiresAt.UTC(), "csrf_token": csrfToken}})
}

func (s *Service) logout(c *gin.Context) {
    r, w := c.Request, c.Writer
    principal, ok := s.authenticate(r)
    if !ok {
        httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated", "authentication required", httpx.RequestID(r.Context()), nil)
        return
    }
    if !s.validCSRF(r, principal.OperatorID) {
        httpx.WriteError(c, http.StatusForbidden, "forbidden", "request is not authorized", httpx.RequestID(r.Context()), nil)
        return
    }
    if _, err := s.db.ExecContext(r.Context(), "UPDATE operator_sessions SET revoked_at = now() WHERE session_token_hash = $1 AND revoked_at IS NULL", digest(s.sessionToken(r))); err != nil {
        httpx.WriteError(c, http.StatusInternalServerError, "internal_error", "", httpx.RequestID(r.Context()), nil)
        return
    }
    http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
    w.WriteHeader(http.StatusNoContent)
}

func (s *Service) authenticate(r *http.Request) (Principal, bool) {
    token := s.sessionToken(r)
    if token == "" {
        return Principal{}, false
    }
    var row operatorSessionRow
    err := s.db.QueryRowContext(r.Context(), "SELECT u.id, u.display_name, u.status, s.permission_snapshot, s.expires_at FROM operator_sessions s JOIN operator_users u ON u.id = s.operator_user_id WHERE s.session_token_hash = $1 AND s.revoked_at IS NULL", digest(token)).Scan(row.destinations()...)
    if err != nil {
        return Principal{}, false
    }
    return row.principal(time.Now())
}

func (s *Service) validCSRF(r *http.Request, operatorID string) bool {
    if !isUnsafe(r.Method) {
        return true
    }
    if _, allowed := s.config.AllowedOrigins[r.Header.Get("Origin")]; !allowed {
        return false
    }
    token := r.Header.Get("X-CSRF-Token")
    if token == "" {
        return false
    }
    var stored []byte
    if err := s.db.QueryRowContext(r.Context(), "SELECT csrf_token_hash FROM operator_sessions WHERE session_token_hash = $1 AND operator_user_id = $2 AND revoked_at IS NULL", digest(s.sessionToken(r)), operatorID).Scan(&stored); err != nil {
        return false
    }
    return subtle.ConstantTimeCompare(stored, digest(token)) == 1
}

func (s *Service) activeOperator(ctx context.Context, subject string) (string, string, error) {
    var row operatorUserRow
    if err := s.db.QueryRowContext(ctx, "SELECT id, display_name, status FROM operator_users WHERE external_subject = $1", subject).Scan(row.destinations()...); err != nil || row.Status != "ACTIVE" {
        return "", "", errors.New("operator unavailable")
    }
    return row.ID, row.DisplayName, nil
}

func (s *Service) sessionToken(r *http.Request) string {
    cookie, err := r.Cookie(sessionCookieName)
    if err != nil || len(cookie.Value) < 43 || len(cookie.Value) > 255 {
        return ""
    }
    return cookie.Value
}

func (s *Service) sealLoginState(state loginState) (string, error) {
    body, err := json.Marshal(state)
    if err != nil {
        return "", err
    }
    block, err := aes.NewCipher(s.config.CookieEncryptionKey)
    if err != nil {
        return "", err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    nonce := make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(append(nonce, gcm.Seal(nil, nonce, body, []byte(loginCookieName))...)), nil
}

func (s *Service) openLoginState(r *http.Request) (loginState, error) {
    cookie, err := r.Cookie(loginCookieName)
    if err != nil {
        return loginState{}, err
    }
    value, err := base64.RawURLEncoding.DecodeString(cookie.Value)
    if err != nil {
        return loginState{}, err
    }
    block, err := aes.NewCipher(s.config.CookieEncryptionKey)
    if err != nil {
        return loginState{}, err
    }
    gcm, err := cipher.NewGCM(block)
    if err != nil || len(value) < gcm.NonceSize() {
        return loginState{}, errors.New("invalid login cookie")
    }
    body, err := gcm.Open(nil, value[:gcm.NonceSize()], value[gcm.NonceSize():], []byte(loginCookieName))
    if err != nil {
        return loginState{}, err
    }
    var state loginState
    return state, json.Unmarshal(body, &state)
}

func (s *Service) clearLoginCookie(w http.ResponseWriter) {
    http.SetCookie(w, &http.Cookie{Name: loginCookieName, Value: "", Path: "/", MaxAge: -1, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func claimedPermissions(raw json.RawMessage) ([]string, error) {
    var values []string
    if len(raw) == 0 || json.Unmarshal(raw, &values) != nil {
        return nil, errors.New("permission claim must be an array")
    }
    result := make([]string, 0, len(values))
    seen := map[string]struct{}{}
    for _, value := range values {
        if _, allowed := permissionAllowlist[value]; allowed {
            if _, duplicate := seen[value]; duplicate {
                continue
            }
            seen[value] = struct{}{}
            result = append(result, value)
        }
    }
    sort.Strings(result)
    return result, nil
}

func digest(value string) []byte { result := sha256.Sum256([]byte(value)); return result[:] }
func randomToken(size int) string {
    bytes := make([]byte, size)
    if _, err := rand.Read(bytes); err != nil {
        panic("crypto/rand unavailable")
    }
    return base64.RawURLEncoding.EncodeToString(bytes)
}
func isUnsafe(method string) bool {
    return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}
func validReturnTo(value string) bool {
    return strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") && !strings.Contains(value, "\\")
}
