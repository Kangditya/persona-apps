package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	issuer       = "http://localhost:7071"
	clientID     = "persona-operations"
	clientSecret = "local-development-secret"
	redirectURI  = "http://localhost:5174/api/operations/v1/auth/callback"
)

type authorization struct {
	nonce, challenge string
}

var (
	privateKey *rsa.PrivateKey
	codes      = map[string]authorization{}
	codesMu    sync.Mutex
)

func main() {
	var err error
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/.well-known/openid-configuration", discovery)
	http.HandleFunc("/keys", keys)
	http.HandleFunc("/authorize", authorize)
	http.HandleFunc("/token", token)
	log.Printf("local OIDC issuer listening on %s", issuer)
	log.Fatal(http.ListenAndServe(":7071", nil))
}

func discovery(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"issuer": issuer, "authorization_endpoint": issuer + "/authorize",
		"token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys",
		"response_types_supported": []string{"code"}, "subject_types_supported": []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"code_challenge_methods_supported":      []string{"S256"},
	})
}

func keys(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{"keys": []any{map[string]string{
		"kty": "RSA", "kid": "local", "use": "sig", "alg": "RS256",
		"n": base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()), "e": "AQAB",
	}}})
}

func authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("client_id") != clientID || q.Get("redirect_uri") != redirectURI || q.Get("response_type") != "code" || q.Get("code_challenge_method") != "S256" || q.Get("state") == "" || q.Get("nonce") == "" || q.Get("code_challenge") == "" {
		http.Error(w, "invalid authorization request", http.StatusBadRequest)
		return
	}
	code := randomString(32)
	codesMu.Lock()
	codes[code] = authorization{nonce: q.Get("nonce"), challenge: q.Get("code_challenge")}
	codesMu.Unlock()
	callback, _ := url.Parse(redirectURI)
	params := callback.Query()
	params.Set("code", code)
	params.Set("state", q.Get("state"))
	callback.RawQuery = params.Encode()
	http.Redirect(w, r, callback.String(), http.StatusFound)
}

func token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid token request", http.StatusBadRequest)
		return
	}
	id, secret, ok := r.BasicAuth()
	if !ok {
		id, secret = r.Form.Get("client_id"), r.Form.Get("client_secret")
	}
	code := r.Form.Get("code")
	codesMu.Lock()
	auth, found := codes[code]
	delete(codes, code)
	codesMu.Unlock()
	verifierHash := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
	challenge := base64.RawURLEncoding.EncodeToString(verifierHash[:])
	if id != clientID || secret != clientSecret || r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("redirect_uri") != redirectURI || !found || challenge != auth.challenge {
		http.Error(w, "invalid token request", http.StatusBadRequest)
		return
	}
	now := time.Now()
	claims := map[string]any{
		"iss": issuer, "aud": clientID, "sub": "local-operator", "nonce": auth.nonce,
		"iat": now.Unix(), "exp": now.Add(time.Hour).Unix(),
		"permissions": []string{"event.read", "event.manage", "offering.read", "offering.manage"},
	}
	writeJSON(w, map[string]any{"access_token": "unused", "token_type": "Bearer", "expires_in": 3600, "id_token": sign(claims)})
}

func sign(claims map[string]any) string {
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "local", "typ": "JWT"})
	payload, _ := json.Marshal(claims)
	input := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(input))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		panic(err)
	}
	return input + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func randomString(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
