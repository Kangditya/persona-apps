package auth

import (
    "context"
    "crypto"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "encoding/base64"
    "encoding/binary"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "net/url"
    "testing"
    "time"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/Kangditya/persona-apps/apps/api/internal/config"
)

type fakeOIDC struct {
    key    *rsa.PrivateKey
    token  string
    server *httptest.Server
}

func TestCallbackCreatesHashedSessionAfterPKCEAndNonceValidation(t *testing.T) {
    fake := newFakeOIDC(t)
    defer fake.server.Close()
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    service, err := New(context.Background(), db, config.AuthConfig{
        IssuerURL: fake.server.URL, ClientID: "client", ClientSecret: "secret",
        RedirectURL: "http://localhost/callback", PermissionClaim: "permissions",
        OperationsWebOrigin: "https://operations.example.invalid", AllowedOrigins: map[string]struct{}{"https://operations.example.invalid": {}},
        CookieEncryptionKey: make([]byte, 32), MaxSessionLifetime: time.Hour,
    })
    if err != nil {
        t.Fatal(err)
    }
    router := testRouter(service)
    loginRequest := httptest.NewRequest(http.MethodGet, "/api/operations/v1/auth/login?return_to=/welcome", nil)
    loginResponse := httptest.NewRecorder()
    router.ServeHTTP(loginResponse, loginRequest)
    loginCookie := loginResponse.Result().Cookies()[0]
    probe := httptest.NewRequest(http.MethodGet, "/", nil)
    probe.AddCookie(loginCookie)
    loginState, err := service.openLoginState(probe)
    if err != nil {
        t.Fatalf("openLoginState() error = %v", err)
    }
    location, err := url.Parse(loginResponse.Result().Header.Get("Location"))
    if err != nil {
        t.Fatal(err)
    }
    nonce := location.Query().Get("nonce")
    if loginState.State != location.Query().Get("state") {
        t.Fatal("login redirect state did not match login cookie")
    }
    fake.token = signedIDToken(t, fake.key, fake.server.URL, "client", "operator-subject", nonce)
    mock.ExpectQuery("SELECT id, display_name, status FROM operator_users").WithArgs("operator-subject").WillReturnRows(sqlmock.NewRows([]string{"id", "display_name", "status"}).AddRow("00000000-0000-0000-0000-000000000001", "Operator", "ACTIVE"))
    mock.ExpectExec("INSERT INTO operator_sessions").WithArgs("00000000-0000-0000-0000-000000000001", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
    callback := httptest.NewRequest(http.MethodGet, "/api/operations/v1/auth/callback?code=code&state="+url.QueryEscape(location.Query().Get("state")), nil)
    callback.AddCookie(loginCookie)
    response := httptest.NewRecorder()
    router.ServeHTTP(response, callback)
    if response.Code != http.StatusFound {
        t.Fatalf("callback status = %d body=%s", response.Code, response.Body.String())
    }
    if got := response.Result().Header.Get("Location"); got != "https://operations.example.invalid/welcome" {
        t.Fatalf("redirect = %q", got)
    }
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Fatal(err)
    }
}

func newFakeOIDC(t *testing.T) *fakeOIDC {
    t.Helper()
    key, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        t.Fatal(err)
    }
    fake := &fakeOIDC{key: key}
    fake.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/.well-known/openid-configuration":
            _ = json.NewEncoder(w).Encode(map[string]string{"issuer": fake.server.URL, "authorization_endpoint": fake.server.URL + "/authorize", "token_endpoint": fake.server.URL + "/token", "jwks_uri": fake.server.URL + "/keys"})
        case "/keys":
            n := base64.RawURLEncoding.EncodeToString(fake.key.PublicKey.N.Bytes())
            e := make([]byte, 4)
            binary.BigEndian.PutUint32(e, uint32(fake.key.PublicKey.E))
            _ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{"kty": "RSA", "kid": "test", "alg": "RS256", "use": "sig", "n": n, "e": base64.RawURLEncoding.EncodeToString(bytesWithoutLeadingZero(e))}}})
        case "/token":
            w.Header().Set("Content-Type", "application/json")
            _ = json.NewEncoder(w).Encode(map[string]any{"access_token": "ignored", "token_type": "Bearer", "id_token": fake.token})
        default:
            http.NotFound(w, r)
        }
    }))
    return fake
}

func signedIDToken(t *testing.T, key *rsa.PrivateKey, issuer, audience, subject, nonce string) string {
    t.Helper()
    header := base64.RawURLEncoding.EncodeToString([]byte("{\"alg\":\"RS256\",\"kid\":\"test\"}"))
    payload, err := json.Marshal(map[string]any{"iss": issuer, "aud": audience, "sub": subject, "nonce": nonce, "exp": time.Now().Add(time.Hour).Unix(), "permissions": []string{"event.read"}})
    if err != nil {
        t.Fatal(err)
    }
    input := header + "." + base64.RawURLEncoding.EncodeToString(payload)
    sum := sha256.Sum256([]byte(input))
    signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
    if err != nil {
        t.Fatal(err)
    }
    return input + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func bytesWithoutLeadingZero(value []byte) []byte {
    for len(value) > 1 && value[0] == 0 {
        value = value[1:]
    }
    return value
}
