# Task: W1-02 Select HTTP, Migration, and Authentication Foundations

## Status

Inactive draft. Activate only after W1-01 is executed and current source is
revalidated.

## Tracker

- Workstream: Platform Safety
- Estimate: 4 hours
- Dependencies: W1-01
- Tracker objective: Select and document the Go HTTP router, migration tooling,
  and authentication approach.

## Objective

Lock the smallest secure platform choices needed by the first vertical slice,
reusing working repository foundations instead of introducing parallel stacks.

## Existing Foundation to Reuse

- `apps/api/internal/app/server.go` already uses Go 1.22+ method-aware
  `http.ServeMux` routes and server timeouts.
- ADR-040 and `apps/api/cmd/db` already use `golang-migrate/migrate/v4`.
- `operator_users.external_subject` already provides an identity-provider
  subject mapping.
- Public and operations API surfaces are already separated architecturally.

## Locked Decisions

### HTTP and migrations

- Keep `net/http` and `http.ServeMux`; add no third-party router.
- Register public and operations routes explicitly from the application
  composition root. Do not create empty domain modules.
- Keep `golang-migrate/migrate/v4`, numbered SQL pairs, and the explicit
  `cmd/db` lifecycle. API startup never runs migrations.
- Preserve migrations `0001` through `0004`; future changes use new pairs.

### Operations authentication

- Use provider-neutral OpenID Connect Authorization Code flow with PKCE.
- The Go API owns login, callback, token validation, and the browser session.
- Use `github.com/coreos/go-oidc/v3/oidc` to discover the provider and verify ID
  token issuer, audience, signature, expiry, and nonce. Use
  `golang.org/x/oauth2` for Authorization Code and PKCE exchange.
- Match the verified OIDC `sub` claim to `operator_users.external_subject`.
  Unknown or inactive operators are denied.
- Create a cryptographically random opaque session token, return it only in a
  `Secure`, `HttpOnly`, `SameSite=Lax` cookie, and store only its SHA-256 hash.
- Sessions are server-side, revocable, and expire no later than the validated
  identity session. Logout revokes the row and clears the cookie.
- Copy only an allowlisted permission claim into the session snapshot. Backend
  authorization checks the session permissions on every operations request.
- Bind login state, nonce, and PKCE verifier to a short-lived signed or
  server-side transaction. Never accept callback parameters without the match.
- Unsafe cookie-authenticated requests require an allowed `Origin` and a CSRF
  token supplied in `X-CSRF-Token`; only its hash is stored server-side.
- Do not implement local passwords, password reset, account recovery, or MFA.
  Those remain identity-provider responsibilities.

### Storefront access

- Storefront browsing and checkout remain guest-accessible for the MVP.
- Purchase creation returns a separate random opaque purchase-access token
  once; PostgreSQL stores only its SHA-256 hash.
- Purchase tracking, cancellation, and evidence submission require the token
  as a Bearer credential and scope access to that Purchase only.
- A purchase token is not an operator session and carries no operations
  permissions.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `docs/DECISIONS.md` | Modify | Accept the router, migration, OIDC session, and guest-access choices. |
| `docs/ARCHITECTURE.md` | Modify | Document public and operations authentication boundaries. |
| `docs/security/AUTHENTICATION.md` | Create | Specify OIDC, session, CSRF, logout, and purchase-token contracts. |
| `.codex/CURRENT_STATE.md` | Modify | Record selected but unimplemented platform choices. |
| `.codex/TASK.md` | Create, execute, archive | Preserve the activated task and final review. |

## Acceptance Criteria

1. The repository has one accepted HTTP router and one migration tool.
2. Operations login, callback validation, operator mapping, session storage,
   cookie settings, permission snapshot, CSRF, expiry, and logout are explicit.
3. Storefront guest access and Purchase-scoped token behavior are explicit.
4. The selected Go libraries and their responsibilities are documented without
   adding them to `go.mod` before runtime implementation.
5. No provider-specific tenant, cloud resource, local password, or speculative
   frontend auth framework is introduced.
6. No runtime, migration, OpenAPI, CI, or infrastructure changes occur.

## Verification

- Cross-check decisions against current server, config, database CLI, schema,
  and ADR-040.
- Confirm the OIDC verifier contract from current upstream documentation.
- Run `make format-check` and `make validate`.
- Review the final diff for token leakage, CSRF, authorization, and session
  revocation gaps.

## Risks and Deferred Work

- Provider issuer, client credentials, claim name, and redirect URLs are
  deployment configuration, not source defaults.
- Operator provisioning and permission administration UI are deferred.
- Runtime implementation belongs to W1-06; staging values belong to W1-07.

## Final Review Requirements

Distinguish selected decisions, verified compatibility, assumptions, deferred
provider configuration, remaining security risks, and the next ready task.
