# Authentication and Access Contract

## Status and Scope

This document defines the accepted Phase 1 authentication boundary. It is a
runtime contract for W1-03, W1-05, W1-06, and W1-07; none of the behavior is
implemented yet.

- Storefront browsing and checkout use guest access.
- Purchase tracking and commands use a Purchase Bearer token.
- The Operations API uses an OIDC-backed server session.

Public and operations credentials are separate and never grant access across
the boundary.

## Operations OIDC Flow

Use one configured OpenID Connect issuer and Authorization Code flow with PKCE.
The Go API is the relying party and owns these steps:

1. Generate cryptographically random state, nonce, and PKCE verifier.
2. Put state, nonce, verifier, return path, and a 10-minute expiry in an
   authenticated-encrypted, `Secure`, `HttpOnly`, `SameSite=Lax`, host-only
   login cookie.
3. Redirect with `openid` scope, the nonce, and an S256 PKCE challenge.
4. On callback, reject provider errors, missing values, expired login state, or
   state mismatch before exchanging the code.
5. Exchange the code with the original PKCE verifier.
6. Require an ID token. Discover keys from the configured issuer and verify its
   issuer, audience/client ID, signature, and expiry.
7. Compare the ID token nonce with the stored nonce. The OIDC library exposes
   nonce but the application owns this comparison.
8. Extract the subject and configured permission claim only after successful
   verification. Clear the login cookie after success or terminal failure.

Use `github.com/coreos/go-oidc/v3/oidc` for discovery and ID-token
verification. Use `golang.org/x/oauth2` for Authorization Code and PKCE with
`GenerateVerifier`, `S256ChallengeOption`, and `VerifierOption`.

The API requests no offline access and stores no provider access token or
refresh token because Phase 1 does not call provider APIs after login.

## Operator Mapping and Permissions

- Map the verified `sub` claim to `operator_users.external_subject`.
- Exactly one issuer is configured, so subject uniqueness is scoped by that
  deployment. Multiple issuers require a new schema and decision.
- Reject missing, unknown, or inactive operators.
- Read permissions only from the configured claim and intersect them with the
  application's allowlist. Ignore unknown values.
- Store the resulting permission snapshot in the server session.
- Backend policies check the required permission on every operations request.
  Frontend route guards are not authorization controls.
- Permission changes take effect after session revocation or the next login;
  administrators must be able to revoke existing sessions.

## Operations Session

- Generate at least 32 random bytes and encode them as an opaque session token.
- Store only its SHA-256 hash. Never log or persist the raw token.
- Send it in a `__Host-operations_session` cookie with `Secure`, `HttpOnly`,
  `SameSite=Lax`, `Path=/`, and no `Domain`.
- Set expiry to the earlier of the validated identity expiry and the configured
  maximum session lifetime.
- Reject expired, revoked, missing, malformed, unknown, or inactive-operator
  sessions.
- Logout is an authenticated unsafe command: validate CSRF, mark the session
  revoked, and expire the cookie. Repeated logout remains harmless.

## CSRF and Origin Checks

- Configure an exact allowlist of Operations Web origins. Wildcard origins are
  forbidden when credentials are enabled.
- For `POST`, `PUT`, `PATCH`, and `DELETE`, require both an allowed
  `Origin` and `X-CSRF-Token`.
- Generate a random CSRF token per session and store only its SHA-256 hash.
- An authenticated `GET /api/operations/v1/auth/session` rotates the token,
  stores its hash, and returns the raw token in the response body. The
  Operations Web keeps it in memory and sends it only in the CSRF header.
- The OIDC callback is protected by state, nonce, and PKCE rather than the
  operations CSRF header.

## Storefront and Purchase Access

- Event and Offering reads plus Purchase creation are guest-accessible.
- Purchase creation generates at least 32 random bytes, returns the opaque
  access token once, and stores only its SHA-256 hash.
- Tracking, cancellation, and payment-evidence submission send
  `Authorization: Bearer <purchase-token>`.
- Validate the token against the identified Purchase using constant-time hash
  comparison. It grants access only to the public fields and commands of that
  Purchase.
- A Purchase token is not a user account, operator session, OIDC token,
  permission set, or payment credential.
- Never put the raw token in URLs, logs, analytics, audit payloads, or error
  messages.

Public account registration, token recovery/rotation, and participant-wide
identity remain deferred. Rate limiting and normal input validation still
apply to public endpoints.

## Configuration Contract

W1-06 and W1-07 must validate names equivalent to:

- `OIDC_ISSUER_URL`
- `OIDC_CLIENT_ID`
- `OIDC_CLIENT_SECRET`
- `OIDC_REDIRECT_URL`
- `OIDC_PERMISSION_CLAIM`
- `OPERATIONS_ALLOWED_ORIGINS`
- `AUTH_COOKIE_ENCRYPTION_KEY`

Values are environment secrets or deployment configuration and must not be
committed. Redirect URLs and allowed origins must use HTTPS outside local
development.

## Deferred

- Identity-provider selection and administration.
- Local passwords, password recovery, and API-owned MFA.
- Operator provisioning and permission-management UI.
- Event/location-scoped permission evaluation.
- Multiple OIDC issuers.
- Public accounts and Purchase-token recovery or rotation.
