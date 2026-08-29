# Task: W1-06 Establish API Platform Foundations

## Executed

## Status

Inactive draft. Activate only after W1-02, W1-03, and W1-04 are executed and
current source is revalidated.

## Tracker

- Workstream: Platform Safety
- Estimate: 7 hours
- Dependencies: W1-02; W1-03; W1-04
- Tracker objective: Establish migration runner, transaction boundary, module
  registration, request IDs, audit writer, and idempotency foundation.

## Objective

Implement the smallest executable Go platform layer needed by Phase 1 commands
while reusing the existing server, database adapter, migration CLI, and schema.

## Approved execution amendment

W1-05 added ADR-045. This task implements its AES-256-GCM replay envelope and
key ring alongside the reviewed platform scope. No domain handler, object
storage adapter, cleanup worker, ORM, router, or broker is introduced.

## Existing Foundation to Reuse

- Keep `http.ServeMux`, current server timeouts, `/health`, and `/ready`.
- Keep the `database/sql` plus pgx adapter and existing `cmd/db` runner.
- Keep ADR-041's `idempotency_records` and direct financial ledger guard.
- Use the W1-03 session/audit schema and W1-04 permission vocabulary.

## Required Implementation

### Composition and transactions

- Keep one application composition root that constructs concrete platform
  dependencies and explicitly registers routes on one `ServeMux`.
- Add a small transaction helper that begins, invokes a callback, rolls back on
  error/panic, and commits only on success. Do not add an ORM or unit-of-work
  framework.
- Do not create empty Event, Offering, Purchase, or Payment module shells.

### Request IDs and errors

- Add middleware that validates a bounded incoming `X-Request-ID` or generates
  a UUID, stores it in context, returns it in the response, and attaches it to
  structured logs.
- Add the common
  `{"error":{"code":"...","message":"...","request_id":"...","details":{}}}`
  JSON error writer; W1-05 must reuse the same shape. Never expose internal
  error text for `500` responses.

### Operations authentication and authorization

- Add W1-02 OIDC Authorization Code plus PKCE login/callback handling using
  `github.com/coreos/go-oidc/v3/oidc` and `golang.org/x/oauth2`.
- Persist and revoke hashed opaque sessions, enforce expiry/inactive operators,
  validate Origin/CSRF on unsafe methods, and attach actor plus permissions to
  request context.
- Add a narrow permission check helper used by operations handlers. Authenticate
  and authorize before idempotency lookup or replay.

### Audit writer

- Add a concrete audit writer that accepts the caller's `*sql.Tx` and inserts
  actor, source, permission, request ID, action, target, reason, and minimized
  before/after JSON.
- Audit failure aborts the owning privileged command; there is no out-of-band
  audit fallback.

### Idempotency executor

- Add a concrete command executor receiving namespace, caller scope, key,
  canonical request hash, retention, and a transaction callback.
- Build namespace from API surface, command, and stable caller scope. Hash
  method, route identity, and canonical validated payload; multipart hashes use
  evidence digest/metadata, not random MIME boundaries.
- Inside one transaction, use `INSERT ... ON CONFLICT DO NOTHING` to claim the
  replay key. A concurrent duplicate waits for the winner, then reads its
  completed response. A winner rollback leaves no completed record.
- Same key/hash returns stored status/body. Same key with another hash returns
  `409 idempotency_conflict`. Transient infrastructure failures are not stored.
- Domain mutation, histories, audit, outbox, and replay response commit
  together. No authorization result is cached.
- Retention is supplied per command; no cleanup worker is added before a real
  command defines its window.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `apps/api/internal/app/server.go` | Modify | Compose middleware, platform routes, and explicit registration. |
| `apps/api/internal/platform/database/transaction.go` | Create | Provide the minimal safe transaction callback. |
| `apps/api/internal/platform/httpx/request_id.go` | Create | Implement request correlation middleware. |
| `apps/api/internal/platform/httpx/errors.go` | Create | Write the shared error envelope. |
| `apps/api/internal/platform/auth/*` | Create | Implement OIDC, session, CSRF, and permission behavior. |
| `apps/api/internal/platform/audit/writer.go` | Create | Insert append-only audit rows in caller transactions. |
| `apps/api/internal/platform/idempotency/executor.go` | Create | Implement transactional command replay. |
| `apps/api/internal/config/config.go` | Modify | Validate required OIDC/session/origin configuration. |
| `apps/api/go.mod` and `apps/api/go.sum` | Modify | Add only go-oidc and oauth2 dependencies. |
| `.env.example` | Modify | List configuration names without secret values. |
| Matching `*_test.go` files | Create/modify | Cover request ID, auth, transaction, audit, and idempotency behavior. |
| `.codex/CURRENT_STATE.md` | Modify | Record implemented platform behavior and remaining vertical-slice gaps. |
| `.codex/TASK.md` | Create, execute, archive | Preserve the activated task and final review. |

## Acceptance Criteria

1. Existing health/readiness behavior remains passing through the composed
   middleware stack.
2. Transaction tests prove commit, callback error rollback, commit error, and
   panic rollback behavior.
3. OIDC/session tests cover state, nonce, PKCE, token validation, unknown or
   inactive operator, expiry, revocation, CSRF, and permission denial.
4. Audit rows use the caller transaction and audit failure rolls back the
   command.
5. Database-backed idempotency tests cover first execution, replay, mismatch,
   concurrent duplicate, rollback retry, authorization-before-replay, and
   atomic audit/outbox/response behavior.
6. No generic repository framework, third-party router, ORM, broker, cleanup
   worker, or empty domain module is added.
7. Migration lifecycle remains explicit and is exercised, not reimplemented.

## Verification

- Run `gofmt`, `go vet ./...`, `go test ./...`, and `go build ./...` in
  `apps/api`.
- Run database-backed tests against disposable PostgreSQL 18 with W1-03
  migrations applied.
- Run migration validate/status/version and `make validate`.
- Review logs and test failures for token, evidence, SQL, or personal-data
  leakage.

## Risks and Deferred Work

- No business endpoint uses the foundation until the first vertical slice.
- Command-specific retention and cleanup are deferred until real commands set
  retention windows.
- OIDC provider administration, evidence storage adapter, public rate limiting,
  and role-management UI are deferred.

## Final Review Requirements

Distinguish implemented platform behavior, database-backed verification,
assumptions, deferred adapters, dependency changes, plan variance, remaining
security/data-integrity risks, and the next task.

## Final Review

### Implemented

- Request ID middleware, error envelopes, transaction helper, caller-owned
  audit writer, encrypted replay executor, and the reviewed OIDC/OAuth2
  dependencies.
- OIDC PKCE login/callback, encrypted ten-minute login state, hashed revocable
  sessions, exact Origin and CSRF checks, permission guard, and fail-closed
  configuration.
- ADR-045 AES-256-GCM replay envelopes with key identifiers and authenticated
  binding to namespace, key, request hash, and status.

### Verified

- API go test, race test, vet, and build passed; the pinned vulnerability scan
  reported no output or failure.
- The local signed-JWT provider test covers discovery, PKCE exchange, state,
  nonce, active operator mapping, and session insertion.
- PostgreSQL 18 migrations reached version 5 cleanly, and the real replay
  integration test passed after migration. make validate passed.

### Assumed

- The configured OIDC provider returns an array permission claim and honors
  standard Authorization Code with PKCE behavior.
- Retained replay keys remain available until all records encrypted with them
  expire.

### Deferred

- Domain handlers, Purchase token runtime, evidence storage, rate limiting,
  session-revocation administration, provider administration, cleanup worker,
  and object-storage adapter.

### Next Task

Implement W1-07 staging requirements and CI PostgreSQL migration/seed/test
verification.
