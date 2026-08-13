# Task: W2-04 Add Public Event and Offering Queries

## Executed

## Status

Executed — 2026-08-13. The user explicitly authorized execution.

## Tracker

- Week: Week 2
- Epic: Event & Offering
- Week goal: Allow Operations to configure commerce and Storefront to discover
  it.
- Milestone: Public Event and Offering discovery.
- Weekly allocation: 42 hours shared across W2-01 through W2-08; the tracker
  does not assign per-task estimates, so this plan does not invent them.
- Tracker objective: Add public Event and Offering queries.
- Dependencies: W2-01, W2-02, and W2-03.

## Objective

Implement the three contracted guest-readable public queries with strict
active/published filtering, minimal public DTOs, advisory availability, safe
errors, request correlation, explicit rate limiting, and no leakage of
Operations-only state.

## Context

`contracts/openapi/storefront.yaml` already defines:

```text
GET /api/public/v1/events/active
GET /api/public/v1/events/{event_id}/offerings
GET /api/public/v1/offerings/{offering_id}
```

The public Gin group exists but has no handlers. The contract includes `429`
responses, but no rate limiter is implemented. The current Gin engine also has
no explicit trusted-proxy configuration. The local Vite `/api` proxy currently
rewrites the prefix and has not been proven against the canonical business
routes.

## Source of Truth

- `docs/PRD.md` Storefront scope, security, performance, and privacy sections;
- `docs/PRODUCT_MAP.md` Storefront and Offering Catalogue ownership;
- `docs/ARCHITECTURE.md` sections 5.1, 10, 13, 16 through 19, and 28;
- `docs/DECISIONS.md` ADR-002, ADR-010, ADR-018, ADR-032, ADR-039,
  ADR-042, ADR-043, and ADR-046;
- `docs/security/PERMISSIONS.md` Storefront boundary;
- `docs/CONVENTIONS.md` HTTP, error, pagination, frontend, and logging rules;
- `contracts/openapi/storefront.yaml`;
- current Gin router, `httpx`, Storefront API client, and Vite proxy source;
- W2-03 repository behavior.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Before activation, re-read the exact public contract and W2-03 repository
outputs. Contract corrections must land before or with runtime behavior. Do
not implement client screens here; W2-07 owns them.

## Scope

### In Scope

- Implement the three existing public read endpoints.
- Register an Event/Offering public HTTP adapter on the existing
  `/api/public/v1` Gin group.
- Return only the active Event and only published Offerings belonging to an
  active Event.
- Preserve the contract's active-event absence behavior: `200` with
  `{"data":null}` rather than `404`.
- Return public `404` for an unavailable event-specific list/detail resource
  without revealing whether an internal draft or archived record exists.
- Return advisory availability using W2-02/W2-03 semantics.
- Correct `available_participant_units` to nullable for fully unbounded quota.
- Align money/quota schema maximums with JavaScript-safe integer exactness.
- Add bounded public read rate limiting with safe client-IP derivation.
- Fix and test API base-path/proxy composition so canonical `/api/public/v1`
  paths reach the API unchanged.
- Preserve structured API error code and request ID in the shared client so the
  Storefront can distinguish `not_found`, `rate_limited`, and dependency errors
  without parsing messages.
- Preserve error `details` only as untrusted structured data; consumers validate
  recognized field-error shapes and never render or log the object wholesale.
- Preserve a valid bounded `Retry-After` delta from `429` responses for UI
  guidance; ignore malformed values rather than scheduling an unsafe delay.
- Preserve request ID, safe error envelope, cancellation, timeout, and
  structured logging behavior.

### Out of Scope

- Checkout, Purchase tokens, payment evidence, CORS credential handling for
  Operations, or any public mutation.
- Storefront screens, SEO/SSR, search/filter, total counts, or generic public
  pagination.
- Redis/distributed limiting, user accounts, CAPTCHAs, WAF/CDN provisioning, or
  analytics.
- Exposing draft, suspended, unavailable, closed, archived, audit, version,
  internal quota composition, reservation rows, operator data, or storage
  references.

## Existing State

- Gin uses `gin.New()` and separate public/Operations route groups.
- Request ID middleware and safe JSON error writing exist.
- `registerPublicRoutes` currently creates an unused group.
- No recovery middleware is registered; an unhandled panic would bypass the
  intended JSON error boundary.
- Gin proxy trust is not configured explicitly.
- The shared frontend client collapses `404` into an unknown error and `429`
  into unavailable, and discards the stable error code/request ID needed for
  intentional public states.
- The public OpenAPI DTO excludes Operations timestamps and internal fields,
  but availability is currently typed only as a non-negative integer.
- The shared client supports timeouts, cancellation, JSON parsing, and error
  normalization.

## Target State

- Public handlers are thin Gin adapters: bind/validate UUIDs, call
  framework-neutral queries with `c.Request.Context()`, map results to public
  DTOs, and render the contract envelope.
- Public query SQL itself enforces active/published visibility; filtering is
  not left to handlers or the frontend.
- Event-specific list and detail endpoints use indistinguishable `not_found`
  behavior for nonexistent and non-public records.
- Public errors expose only stable codes/messages/request IDs; internal SQL or
  topology is logged safely, not returned.
- A process-local token bucket limits guest reads by verified client IP, with a
  default of 60 requests/minute and burst 20, bounded entry memory, periodic
  stale-entry cleanup, and documented per-replica scope.
- The initial limiter holds at most 10,000 client entries and expires entries
  idle for 10 minutes. If cleanup cannot make space, an unseen key is rejected
  with `429` rather than growing memory without bound; ingress remains the
  mitigation for distributed key-filling attacks.
- Perform stale-entry cleanup opportunistically at a bounded interval under the
  limiter lock; do not add a background goroutine solely for this map.
- Forwarded client IP is ignored unless exact trusted proxy CIDRs are provided
  through validated deployment configuration. Direct peer IP is the default.
- Ingress rate limiting remains defense in depth for multi-replica deployment.
- Gin recovery converts unexpected panics to the standard `500 internal_error`
  envelope without leaking stack details to callers.
- Unknown paths and unsupported methods under API prefixes use the same JSON
  error/request-ID boundary rather than Gin's current plain-text fallback.

## Constraints

- The current Go module does not declare `golang.org/x/time`. Add the official
  `golang.org/x/time/rate` package as one direct, pinned,
  vulnerability-checked dependency; it is safer and smaller than inventing a
  token bucket. Use normal Go tooling and never hand-edit `go.sum`.
- The current Go 1.26.5 standard library and module declare no UUID parser. Add
  `github.com/google/uuid` as one direct, pinned, vulnerability-checked HTTP
  adapter dependency for canonical path validation; keep it out of generic
  domain abstractions.
- Bound limiter keys and cleanup; an unbounded map keyed by attacker-controlled
  IP/header data is itself a denial-of-service risk.
- Do not trust `X-Forwarded-For` from arbitrary peers. Never use a public
  header value directly as a security key.
- Public reads are guest-accessible but remain subject to input validation,
  response minimization, rate limiting, and server timeouts.
- Return `Cache-Control: no-store` for Week 2 catalogue reads because Event
  status and advisory quota can change; introduce cache semantics only through
  a later accepted requirement.
- Do not log full request bodies, query values containing personal data, or
  internal database errors at the response boundary.

## Implementation Requirements

### Contract corrections

- Update public availability to `integer | null` with clear unbounded meaning.
- Add maximums for Event/Offering quotas and `price_minor` consistent with the
  exact JSON number boundary.
- Add the PostgreSQL `integer` maximum to public `participant_capacity`; do not
  apply the larger bigint/JSON bound to an `int4` column.
- Update the contract description so implemented discovery routes are no
  longer described as wholly unimplemented; do not claim later Purchase routes.
- Add a reusable method-not-allowed response where router behavior can emit
  `405` and add `method_not_allowed` to the Storefront error vocabulary, while
  preserving each operation's declared success/error set.
- Preserve the three existing route paths and operation IDs.
- Document that active Event discovery is independent of whether its
  registration window is currently open; return window instants for the UI.
- Apply the approved endpoint-specific opaque cursor pagination to the public
  Offering list: default `limit` 50, maximum 100, no total count. Return a
  `next_cursor` only when a further page exists; never silently truncate or
  issue an unbounded catalogue query.

### HTTP adapter

- Add a public registrar owned by Event/Offering rather than putting business
  handlers in `internal/app`.
- Validate UUID path parameters before repository calls.
- Map dependency outages to `503`, unexpected failures to sanitized `500`,
  malformed IDs/request IDs to `400`, and unavailable public resources to
  `404`.
- Map API-prefix no-route/no-method outcomes to structured `404`/`405` without
  changing the simple health/readiness response contract.
- Return deterministic Offering order from W2-03.
- Use the existing response/error envelope patterns; add the smallest success
  writer only if repeated code justifies it.

### Rate limit and network boundary

- Add config names for public limiter rate/burst and trusted proxy CIDRs with
  safe defaults and validation. No secret values are committed.
- Use `PUBLIC_RATE_LIMIT_PER_MINUTE` (default `60`),
  `PUBLIC_RATE_LIMIT_BURST` (default `20`), and comma-separated
  `TRUSTED_PROXY_CIDRS` (default empty/disabled). Invalid numbers, addresses,
  or CIDRs fail startup; raw forwarded-header values are never configuration.
- Call Gin's trusted-proxy configuration explicitly; `nil`/disabled is the
  default until CIDRs are configured.
- Apply the limiter only to public business routes, not health/readiness or
  OIDC callbacks.
- Return `429 rate_limited` and a reasonable `Retry-After` value without
  exposing the limiter key.
- Add `Retry-After` to the contracted `429` response headers.
- Clean idle limiter entries under a fixed bound/interval and test the bound.

### Base-path correction

- Use one convention: frontend base URL defaults to the same origin (empty
  prefix) and endpoint modules pass canonical `/api/public/v1/...` paths;
  deployment may replace the base with one absolute API origin.
- Remove the Vite rewrite that strips `/api`, or make the endpoint/base split
  equivalent and prove it with a proxy-level test. Prefer removing the rewrite
  because the API itself owns the `/api` prefix.
- Proxy root `/health` and `/ready` explicitly in local development if
  diagnostics still use those infrastructure paths; do not force them under
  the business `/api` prefix.
- Keep PWA navigation fallback denying `/api/`; never cache these responses in
  the service worker.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `contracts/openapi/storefront.yaml` | Modify | Correct nullable availability and exact integer bounds. |
| `apps/api/internal/event/public_http.go` | Create | Register and render public Event query. |
| `apps/api/internal/offering/public_http.go` | Create | Register and render public Offering queries. |
| `apps/api/internal/platform/ratelimit/public.go` | Create | Minimal bounded per-process public limiter. |
| `apps/api/internal/platform/ratelimit/public_test.go` | Create | Verify limits, cleanup/bounds, and key safety. |
| `apps/api/internal/platform/httpx/recovery.go` | Create or fold into router | Convert panics to safe request-ID errors. |
| `apps/api/internal/config/config.go` | Modify | Validate limiter and trusted-proxy settings. |
| `apps/api/internal/config/config_test.go` | Modify | Cover secure defaults and invalid configuration. |
| `apps/api/internal/app/router.go` | Modify | Compose trusted proxies, recovery, limiter, and public registrar. |
| `apps/api/internal/app/routes_public.go` | Modify | Accept/register real public module routes. |
| `apps/api/cmd/api/main.go` and `internal/app/server.go` | Modify | Inject concrete public query dependencies/configuration. |
| `apps/api/go.mod` and `apps/api/go.sum` | Modify through Go tooling | Add reviewed direct rate/UUID adapter dependencies; neither is currently present. |
| `packages/api-client/src/index.ts` and tests | Modify | Preserve safe error code/request ID/retry delay and distinguish public states. |
| `apps/storefront-web/vite.config.ts` | Modify | Preserve the canonical `/api` prefix through local proxying. |
| `apps/storefront-web/src/api/client.ts` | Modify | Establish the tested business-route base-path convention. |
| Matching Go/TypeScript tests | Create/modify | Cover routes, exposure, rate limits, recovery, and path composition. |
| `.env.example` and deployment docs | Modify if names are added | Document safe configuration names/defaults only. |
| `.codex/CURRENT_STATE.md` and `.codex/TASK.md` | Modify/create/archive when executed | Record verified runtime state. |

Consolidate files if clarity improves; do not create a generic HTTP framework
inside the repository.

## Dependencies and Sequencing

- Requires W2-03 public repository queries.
- Can run in parallel with W2-05 after W2-03 if shared app composition changes
  are coordinated.
- W2-07 consumes these endpoints and helps prove the base-path convention.
- W2-08 adds exhaustive API/exposure/authorization regression coverage, but
  this task retains focused rate-limit and route tests.

## API and Database Impact

- Changes public schema clarity/bounds and adds the approved Offering-list
  `limit`, `cursor`, and continuation metadata; no route or operation ID
  changes.
- No migration is expected beyond W2-03.
- Public responses never include Operations version fields.

## Authorization, Audit, Idempotency, and Concurrency

- These reads are guest-accessible and have no business permission or CSRF
  requirement.
- Reads do not write audit, outbox, or idempotency rows.
- Rate limiting is abuse control, not authentication or authorization.
- Availability remains advisory and is not protected by frontend cache or read
  locks; later checkout is authoritative.

## Acceptance Criteria

1. `GET /events/active` returns one active Event or `data: null`, never an
   internal Event status.
2. Public list/detail queries can return only published Offerings under the
   active Event, with indistinguishable `404` behavior for hidden/nonexistent
   records.
3. Availability correctly exposes finite values or `null` for unbounded and is
   documented as advisory.
4. Canonical browser paths reach `/api/public/v1/...` without proxy prefix
   stripping; service workers do not cache API data.
5. Public reads are rate-limited with safe direct/trusted client-IP handling,
   bounded limiter memory, `Retry-After`, and the contracted `429` envelope.
6. Panics and infrastructure failures produce sanitized request-ID-bearing
   errors with no SQL, stack, secret, audit, or Operations field leakage.
7. The shared client preserves stable safe error code/request ID, allowing
   `404`, `429`, and `503` UI states without message matching.
8. API-prefix `404`/`405` responses are structured and request-correlated;
   health/readiness success shapes remain unchanged.
9. No mutation, authentication system, distributed cache/limiter, generic
   pagination framework, or generated client is introduced.

## Testing

- Handler tests for active/no-active, list/detail, malformed UUID, hidden
  statuses, empty catalogue, repository failure, and request IDs.
- Exposure assertions against every public JSON field.
- Rate-limit allow/exhaust/recover behavior; spoofed forwarded headers;
  trusted CIDR; limiter-entry bound/cleanup; `Retry-After`.
- Recovery test verifying a panic is sanitized.
- Vite/API-client test proving the exact URL called.
- PostgreSQL-backed public visibility and availability tests remain in W2-03.

## Verification

```bash
cd apps/api && go run ./cmd/format -check
cd apps/api && go vet ./...
cd apps/api && go test ./internal/event/... ./internal/offering/... ./internal/platform/ratelimit/... ./internal/app/...
cd apps/api && go test ./...
pnpm --filter @persona-apps/storefront-web test
pnpm --filter @persona-apps/storefront-web typecheck
pnpm --filter @persona-apps/storefront-web lint
make validate
```

Also parse/lint the Storefront OpenAPI document with the repository's pinned
contract workflow used by W1-05. Do not claim contract validation from YAML
parsing alone.

## Deliverables

- Three implemented public queries and registration.
- Corrected public contract semantics.
- Minimal rate-limit/trusted-proxy/recovery boundary with focused tests.
- Verified base-path behavior and current-state/task archive updates.

## Risks and Clarifications to Review

- Confirmed default: 60 public reads/minute per verified client IP, burst 20.
- Confirmed default: the limiter is per process; ingress enforcement is needed
  for uniform multi-replica limits.
- Confirmed default: forwarded IP headers are ignored unless exact proxy CIDRs
  are configured.
- Confirmed default: catalogue discovery remains visible outside registration
  hours and communicates the window state.
- Confirmed execution decision: the public Offering endpoint uses
  endpoint-specific opaque cursor pagination, default `limit` 50 and maximum
  100. No total count, generic pagination abstraction, publication cap, or
  unbounded response is introduced.
- Same-origin reverse proxy is the preferred deployment path; explicit CORS is
  handled with Operations in W2-05 when origins differ.

## Final Report Requirements

Distinguish implemented public behavior, contract and exposure verification,
rate-limit scope, proxy assumptions, deferred checkout/SSR/pagination work,
plan variance, remaining abuse risks, and the next frontend task.

## Plan Variance

- The reviewed draft rejected an unbounded public catalogue and left its
  maximum size unresolved.
- User execution authorization on 2026-08-13 selected the safe
  endpoint-specific alternative: opaque cursor pagination with a default of
  50 and a maximum of 100 Offerings per response.
- The OpenAPI contract and runtime will use W2-03's existing typed cursor;
  this is not a generic pagination framework and does not add a total-count
  query.

## Execution Review

### Implemented

- Added guest `GET /api/public/v1/events/active`,
  `GET /api/public/v1/events/{event_id}/offerings`, and
  `GET /api/public/v1/offerings/{offering_id}` handlers in the owning Event
  and Offering modules. Repository SQL constrains visibility to an active
  Event and published Offerings; hidden and nonexistent resources share the
  public `404` response.
- Returned only public DTO fields. Availability is advisory and either a
  finite non-negative count or `null` when neither Event nor Offering has a
  finite quota. The bounded Offering list uses the approved opaque cursor,
  default limit 50, maximum 100, deterministic code/ID order, no total count,
  and `next_cursor` only when another page exists.
- Added the public HTTP boundary: request IDs, safe JSON API `404`/`405`,
  panic recovery, sanitized dependency/unknown errors, `Cache-Control:
  no-store`, and a process-local 60/minute-per-IP token bucket (burst 20).
  Limiter entries are capped at 10,000, idle entries expire after ten minutes,
  and forwarded IPs are ignored unless validated trusted-proxy CIDRs are
  configured.
- Added validated public limiter/proxy configuration, documented safe defaults,
  and preserved exact canonical `/api` paths through both Vite development
  proxies. The Operations change is base/proxy compatibility only; it adds no
  Operations business capability.
- Corrected the Storefront contract for nullable availability, safe integer
  bounds, `Retry-After`, no-store responses, structured method errors, and the
  approved list page. Updated the shared client to preserve bounded safe error
  code, request ID, untrusted structured details, and valid `Retry-After`.
- Kept the Event/Offering lifecycle and PostgreSQL work from W2-01 through
  W2-03 intact: additive migration 0006, version/bounds constraints, typed
  repositories, active/published public SQL, and advisory reservation usage.
- Updated direct/indirect dependencies after security scanning: direct UUID
  parsing and official Go token-bucket support; fixed Go `quic-go`, `x/text`,
  `x/net`, `x/crypto`, and `x/sys` releases; and the locked `nanoid` override
  used by the Vite dependency graph.

### Verified

- Scoped changed-Go formatter check passed: 23 files scanned, 0 changed.
- With the disposable local PostgreSQL database,
  `TEST_DATABASE_URL=... go test -count=1 ./...`, `go vet ./...`, `go build
  ./...`, and migration validation (six migration pairs) passed.
- `go test -race -count=1 ./internal/platform/ratelimit/... ./internal/app/...`
  passed.
- Root Storefront/Operations tests (14 tests across 7 files), type checks,
  linting, and production builds passed. The Vite proxy test proves that the
  canonical `/api/public/v1/...` path is not rewritten.
- `pnpm audit --audit-level high` reported no known vulnerabilities. Go
  `govulncheck` found zero reachable-symbol and imported-package findings. Its
  only residual module-graph advisory is the upstream-unfixed,
  uncalled `golang.org/x/crypto/openpgp` package; this API does not import or
  call it, so it is not an application-reachable vulnerability.
- Storefront/Operations OpenAPI YAML and related source passed Prettier
  parsing/format checking. The repository has no pinned semantic OpenAPI
  validator or generated-client workflow, so no uninstalled tool was claimed
  as contract validation; route/DTO/handler tests and source review supplied
  the structural verification.
- `docker compose -f infrastructure/compose.yaml config` passed.
- `make validate` reached the existing Go formatter baseline and stopped only
  on six untouched files: `internal/platform/audit/writer.go`, two database
  transaction files, and three idempotency executor files. Prettier passed;
  the task's changed Go files pass the same formatter. Those pre-existing
  baseline files were deliberately not reformatted as unrelated work.

### Assumed

- The user-approved public list policy is endpoint-specific cursor pagination
  (50 default, 100 maximum), not the plan draft's unresolved/unpaginated
  alternative.
- The verified PostgreSQL target is local and disposable. No staging or
  production deployment compatibility claim is made.
- Trusted proxy CIDRs are configured only for actual reverse-proxy ranges;
  otherwise Gin derives the limiter key from the direct peer.

### Deferred

- Checkout/Purchase mutation, idempotency retention policy, payment/evidence,
  audit/outbox persistence, and Operations Event/Offering commands are W2-05
  and later work.
- Storefront catalogue screens, loading/error UI, SSR/SEO, search/filtering,
  and query-cache policy are W2-07 work.
- The local limiter is intentional initial protection, not distributed abuse
  control. Ingress/CDN limits and any cross-origin Operations CORS decision
  remain deployment/W2-05 concerns.

### Plan Variance

- The approved bounded cursor policy replaces the draft's unresolved catalogue
  size, and is recorded in the contract/runtime rather than creating a generic
  pagination framework.
- No repository-pinned semantic OpenAPI linter exists despite the original
  verification wording. YAML format parsing plus contract/source/tests were
  used instead; no dependency or workflow was invented.
- Because `VITE_API_BASE_URL` is shared, preserving canonical paths required a
  small Operations base/proxy compatibility adjustment alongside Storefront.
- Security checks required narrow locked dependency updates. They are included
  because they remove reachable Go findings, fixed `x/net`/`x/crypto`/`x/sys`
  advisory releases, and the high-severity JavaScript audit finding, not as a
  broad dependency refresh.

### Remaining Risk and Next Task

- Availability remains advisory until a transactional checkout locks and
  revalidates quota. Per-process rate limiting must be paired with ingress/CDN
  enforcement in multi-replica production.
- Next: W2-05 composes authenticated, authorized, audited, idempotent,
  version-aware Operations Event/Offering commands in a transaction.
