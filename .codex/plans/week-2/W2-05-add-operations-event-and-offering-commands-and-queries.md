# Task: W2-05 Add Operations Event and Offering Commands and Queries

## Status

Draft — awaiting review. The discovery defaults were confirmed on 2026-08-13,
but this task is not Plan Approved and must not enter BUILD until this file is
reviewed and explicitly approved.

## Tracker

- Week: Week 2
- Epic: Event & Offering
- Week goal: Allow Operations to configure commerce and Storefront to discover
  it.
- Milestone: Public Event and Offering discovery.
- Weekly allocation: 42 hours shared across W2-01 through W2-08; the tracker
  does not assign per-task estimates, so this plan does not invent them.
- Tracker objective: Add Operations Event and Offering commands and queries.
- Dependencies: W2-01, W2-02, and W2-03; coordinate router/configuration changes
  with W2-04.

## Objective

Expose the approved Event and Offering behavior through the existing
Operations OpenAPI surface with authoritative permission checks, CSRF/origin
protection, optimistic concurrency, transactionally atomic audit/outbox work,
and replay-safe retry-sensitive commands.

## Context

`contracts/openapi/operations.yaml` already drafts Event and Offering CRUD,
list, and lifecycle routes. The Gin Operations group already supports session
authentication and permission middleware, but no Event or Offering routes are
registered. The current contract also lacks mutation versions, cannot express
clearing nullable patch fields, and does not consistently declare idempotency
for retry-sensitive creates.

The existing auth middleware must remain the policy boundary: unsafe requests
are authenticated and origin/CSRF-checked before any command replay lookup.

## Source of Truth

- `docs/PRD.md` Operations, audit, security, and Event/Offering requirements;
- `docs/PRODUCT_MAP.md` Operations ownership and Week 2 roadmap;
- `docs/ARCHITECTURE.md` API boundaries, transaction policy, auth, errors,
  idempotency, and deployment topology;
- `docs/DECISIONS.md` ADR-002, ADR-009, ADR-011, ADR-018, ADR-019, ADR-020,
  ADR-024, ADR-038, ADR-039, ADR-041, ADR-042, ADR-044, and ADR-046;
- `docs/CONVENTIONS.md` HTTP, validation, pagination, money, logging, and Go
  rules;
- `docs/security/PERMISSIONS.md` exact Event and Offering permission mapping;
- `docs/security/AUTHENTICATION.md` session, Origin, and CSRF model;
- `contracts/openapi/operations.yaml`;
- current `auth`, `audit`, `idempotency`, `database`, `httpx`, router, and
  configuration source;
- W2-01 through W2-03 domain/application/repository outputs.

If an existing contract statement conflicts with the accepted lifecycle or
security policy, correct the contract and its canonical owner before or with
the runtime change; do not hide divergence in a handler.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate this task only after its dependencies are implemented and this draft
is explicitly approved. Recheck the active task boundary and worktree before
BUILD. Do not commit or push unless asked.

## Scope

### In Scope

- Register the contracted Operations Event and Offering list, get, create,
  patch, and lifecycle endpoints.
- Keep Gin handlers thin and delegate rules to W2-01/W2-02 application
  services and persistence to W2-03 repositories.
- Enforce `event.read`, `event.manage`, `offering.read`, and `offering.manage`
  with the existing backend permission middleware.
- Require the authenticated session's CSRF token and an allowed exact Origin
  for unsafe cookie-authenticated requests.
- Add Event and Offering `version` to Operations DTOs; require
  `expected_version` for patch and lifecycle commands.
- Preserve omitted patch values while supporting explicit `null` to clear
  nullable registration windows, Event/Offering quota, and Offering
  description.
- Require `Idempotency-Key` for Event/Offering create and lifecycle commands;
  make patch version-safe without storing replay responses.
- Write state, minimized audit, required outbox, and idempotency effects in one
  database transaction.
- Add exact-origin CORS only for explicitly configured browser origins when
  deployment is cross-origin; keep same-origin proxying the preferred default.
- Correct existing session response conformance needed by these screens,
  including `expires_at` and deterministic permission ordering.
- Return stable envelopes, pagination cursors, request IDs, and contract errors
  without leaking SQL, claims, CSRF values, or internal topology.
- Correct the Operations local proxy/API base-path composition so canonical
  `/api/operations/v1/...` paths reach Gin unchanged.

### Out of Scope

- Operator/role administration, login-flow redesign, refresh-token storage,
  or frontend authorization as a security boundary.
- Bulk edit, import/export, delete, restore, duplicated Events/Offerings, or a
  generic command bus.
- Generated API clients, schema-to-Go generation, reflection registries, or a
  generic repository/service/controller framework.
- Cross-origin credential support for wildcard origins, origin patterns, or
  caller-supplied callback URLs.
- Public queries and public rate limiting, which belong to W2-04.
- Operations screens, which belong to W2-06.

## Existing State

- Operations routes live under `/api/operations/v1` and auth routes are
  registered only when the auth service is configured.
- `auth.Service.Require(permission)` authenticates a session and validates
  unsafe Origin/CSRF requirements before invoking a protected handler.
- The authenticated session response contract requires `expires_at`, while
  the current runtime response omits it; permission projection order is not
  deterministic.
- `platform/audit`, `database.Within`, and the encrypted idempotency executor
  exist and have focused tests.
- Response-encryption key configuration is validated, but domain command
  composition must still construct and inject the replay cipher/executor.
- `outbox_events` exists, but the first Event/Offering user must add only the
  minimal transactional writer needed for real events.
- Existing JSON binding patterns do not yet prove unknown-field rejection,
  trailing-document rejection, or the nullable-vs-omitted patch distinction.

## Target State

- Contract and runtime agree on every path, method, request, response, status,
  permission, version, and idempotency requirement.
- Reads require their exact read permission; mutations require their exact
  manage permission. A valid session with the wrong permission receives `403`;
  missing/expired authentication receives `401`.
- Unsafe cookie-authenticated commands fail before application or replay work
  if Origin or CSRF validation fails.
- A repeated command with the same identity, route namespace, idempotency key,
  and canonical validated request replays the stored response; key reuse with
  different input returns `409 idempotency_conflict`.
- A stale `expected_version` returns `409 stale_version`; successful mutations
  increment and return the new version.
- Publicly observable failures use stable error codes. Constraint and driver
  details are translated once at the repository/application boundary.
- Lists are bounded and cursor-based with deterministic sort orders established
  by W2-03; malformed cursors fail validation rather than becoming SQL input.
- CORS preflight is handled before auth only for exact configured origins,
  methods, and headers. Credentialed responses never use `*`.
- Auth/session and Operations business responses use `Cache-Control: no-store`;
  private session, permission, CSRF, and mutable administration data never enter
  browser/shared HTTP caches.

## Constraints

- Reuse the existing auth, audit, idempotency, transaction, error, and config
  code. Extend shared code only where both Event and Offering genuinely need
  the same trust-boundary behavior.
- Use standard library JSON decoding with a small request-size limit,
  `DisallowUnknownFields`, and exactly one JSON value. Do not add a validator
  or binding dependency.
- Limit Event/Offering JSON command bodies to 64 KiB and require
  `Content-Type: application/json`; return contract-defined `413`/`415`
  responses rather than reading unbounded or ambiguous input.
- Parse the media type with the standard library so valid charset parameters
  are handled consistently; reject empty/null bodies for required command
  objects and close request bodies through the normal server lifecycle.
- Reuse the reviewed direct `github.com/google/uuid` HTTP-adapter dependency
  introduced by W2-04. If task order changes, add it once through normal Go
  tooling; do not leak it into generic domain abstractions or hand-edit
  `go.sum`.
- Keep command-specific types in their owning modules. Do not introduce a
  generic patch, command, lifecycle, pagination, or response framework.
- Treat all client values as untrusted even after TypeScript/OpenAPI
  validation. The Go boundary and PostgreSQL constraints remain authoritative.
- Never persist or log raw session cookies, access/refresh tokens, CSRF tokens,
  full authorization claims, or response-encryption keys. The caller's
  idempotency key is stored only in the canonical
  `idempotency_records.idempotency_key` ownership column; it must not be copied
  into audit, outbox, replay bodies, responses, or logs.

## Implementation Requirements

### Contract alignment

- Add required positive `version` fields to Operations Event and Offering
  responses only; public DTOs remain version-free.
- Add required nullable advisory `available_participant_units` to Operations
  Offering responses so the real screens do not infer stock or call public
  endpoints; apply the same finite/unbounded meaning and integer bound as the
  Storefront DTO.
- Add required positive `expected_version` to patch and lifecycle request
  bodies. Transition routes must no longer be empty-body commands.
- Represent clearable patch fields with OpenAPI 3.1 nullable unions while
  preserving omission semantics; generated or hand-written transport types
  must distinguish absent from present-null.
- Require at least one mutable patch field in addition to
  `expected_version`—for example with explicit OpenAPI `anyOf` required-field
  branches. A body containing only the concurrency token is not a successful
  no-op mutation.
- Bound names/codes/kinds/descriptions, years, money, capacities, quotas,
  cursors, and list limits consistently with the schema and database. In
  particular, capacity is at most PostgreSQL `integer` max
  (`2_147_483_647`), while bigint money/quota/version values use the smaller
  JSON-safe bound.
- Declare `Idempotency-Key` on create and lifecycle operations. The proposed
  Week 2 replay retention is 24 hours, but ADR-041 makes retention
  command-specific; approve that duration before BUILD rather than calling it
  an existing platform default. Patch relies on expected version and is not
  silently replay-enabled.
- Add `stale_version` to the Operations error-code vocabulary and add general
  JSON `request_too_large`/`unsupported_media_type` responses to applicable
  commands; do not reuse evidence-upload wording.
- Add `method_not_allowed` and the shared API-prefix `405` response established
  in W2-04 so the Operations contract matches composed-router behavior.
- Update the contract description so Event/Offering routes are not described as
  wholly unimplemented while later Operations groups remain draft.
- Retain the current route names unless canonical source review finds an actual
  mismatch; do not add aliases.

### Read handlers

- Validate path UUIDs, list limits, and opaque cursors before repository calls.
- Map domain/application objects to explicit Operations DTOs; never serialize
  database structs or auth context directly.
- Return archived records to authorized Operations reads when requested by the
  existing contract, but never treat them as mutable.
- Preserve not-found behavior without revealing unrelated internal records or
  database details.

### Command handlers and transactions

- Authenticate/authorize and validate CSRF/Origin first, then parse and
  validate input, then perform idempotency lookup for commands that declare it.
- Canonicalize only validated semantic input for idempotency hashing; irrelevant
  JSON key order/whitespace must not create a different operation, while a
  genuinely different request must conflict.
- Start one transaction for the mutation, audit row, outbox event, and replay
  response. Any required write failure rolls the transaction back.
- Populate audit actor, permission, source, request ID, action, target, and
  minimized before/after fields from trusted server context. Populate
  `client_ip` only from W2-04's direct/trusted-proxy resolution, never a raw
  forwarded header. Never accept actor/audit metadata from the request body.
- Return the resulting resource/version for accepted commands and the original
  stored status/body for a valid replay.
- If a syntactically valid patch supplies mutable fields that normalize to the
  current values, return the current resource/version without a database
  update, audit row, outbox event, or version bump. This is a no-op, not a
  fabricated historical change.

### CORS and session conformance

- Add separate exact allowlists for Storefront and Operations only if their
  deployed origins differ from the API origin. Empty configuration means no
  cross-origin access.
- Reuse the validated `OPERATIONS_ALLOWED_ORIGINS` auth setting for Operations
  CORS and add optional `STOREFRONT_ALLOWED_ORIGINS` for public cross-origin
  reads. Do not introduce two competing Operations-origin settings.
- Permit credentials only for exact trusted Operations origins. Restrict
  methods and request headers to those actually used, including
  `Content-Type`, `X-CSRF-Token`, `Idempotency-Key`, and request correlation.
- Expose only response headers the browsers need, including `X-Request-ID` and
  `Retry-After`; do not expose cookie or internal tracing headers.
- Add `Vary: Origin` where a response varies by Origin and reject disallowed
  preflights without reflecting the caller value.
- Keep cookie attributes and CSRF comparison in the existing auth layer.
- CORS does not weaken the accepted `Secure`, `HttpOnly`, `SameSite=Lax`,
  host-only cookie policy. A genuinely cross-site Operations deployment needs
  a separate reviewed auth decision; same-origin or same-site deployment is the
  Week 2 boundary.
- Make `/auth/session` satisfy its contract by returning `expires_at` and a
  stable sorted permission list, and mark it `no-store`; add no token
  persistence in browser storage.

### Composition and failure behavior

- Compose concrete repositories/services/handlers explicitly in `cmd/api` and
  the existing server/router inputs; do not add a container or service locator.
- Fail closed: do not register protected business routes without working auth,
  audit, transaction, and replay dependencies.
- Map validation to `400`, unauthenticated to `401`, forbidden to `403`, absent
  resources to `404`, stale/idempotency/state/unique conflicts to `409`, rate
  limits to `429` where applicable, dependency outages to `503`, and unexpected
  failures to sanitized `500`.

## Planned File Changes

Exact splitting may be reduced during activation; these are ownership targets,
not permission to create empty layers.

| File | Action | Purpose |
| --- | --- | --- |
| `contracts/openapi/operations.yaml` | Modify | Align versions, nullable patches, idempotency headers, bounds, and responses. |
| `apps/api/internal/event/operations_http.go` | Create | Event list/get/create/patch/lifecycle Gin adapter. |
| `apps/api/internal/offering/operations_http.go` | Create | Offering list/get/create/patch/lifecycle Gin adapter. |
| `apps/api/internal/platform/httpx/json.go` | Create only if current helpers cannot be extended cleanly | Bounded strict JSON decoding shared by real command handlers. |
| `apps/api/internal/platform/outbox/writer.go` | Create | Minimal transaction-bound writer for accepted domain events. |
| `apps/api/internal/platform/outbox/writer_test.go` | Create | Verify payload, transaction, and failure behavior. |
| `apps/api/internal/platform/audit/writer.go` and tests | Modify | Add validated client IP to existing transaction-bound audit entries. |
| `apps/api/internal/platform/cors/middleware.go` | Create | Exact-origin CORS/preflight behavior for configured deployments. |
| `apps/api/internal/platform/cors/middleware_test.go` | Create | Verify credentials, allowlists, preflight, and `Vary`. |
| `apps/api/internal/platform/auth/service.go` and matching tests | Modify | Return contract-complete session expiry and deterministic permissions. |
| `apps/api/internal/app/routes_operations.go` | Modify | Register protected Event/Offering routes with exact permissions. |
| `apps/api/internal/app/router.go`, `server.go`, and tests | Modify | Compose CORS and concrete registrars without changing public auth boundaries. |
| `apps/api/cmd/api/main.go` | Modify | Construct repositories, services, replay cipher/executor, audit, and handlers explicitly. |
| `apps/api/internal/config/config.go` and tests | Modify | Validate optional exact CORS origin lists and safe defaults. |
| `apps/api/go.mod` and `apps/api/go.sum` | Coordinate with W2-04 | Reuse one reviewed direct UUID adapter dependency without duplicate module churn. |
| `apps/operations-web/vite.config.ts` | Modify | Preserve `/api` and proxy root diagnostics explicitly for local development. |
| `apps/operations-web/src/api/client.ts` and tests | Modify | Use same-origin base plus canonical endpoints without double-prefixing. |
| `.env.example` and deployment docs | Modify if configuration names are added | Document origin variables without credentials or permissive examples. |
| Focused module/route tests | Create/modify | Verify contract mapping, auth ordering, replay, concurrency, and failures. |
| `.codex/CURRENT_STATE.md` and `.codex/TASK.md` | Modify/create/archive when executed | Record only verified runtime behavior. |

## Dependencies and Sequencing

- W2-01/W2-02 own behavior; W2-03 owns concrete persistence and versioning.
- Coordinate shared router, config, Go module, and API-client assumptions with
  W2-04; merge one coherent change rather than duplicate middleware.
- W2-06 consumes the resulting Operations API and session shape.
- W2-08 expands cross-layer tests but does not defer critical auth, replay,
  transaction, or contract tests from this task.

## API and Database Impact

- Operations contract gains version/expected-version semantics and corrected
  nullable patch shapes; route paths remain stable.
- No migration beyond W2-03 is expected.
- Create and lifecycle replay uses existing `idempotency_records`; lifecycle work
  emits existing `outbox_events` rows. No new generic command tables are added.

## Authorization, Audit, Idempotency, and Concurrency

- Exact permissions are enforced server-side per route; list/detail UI hiding
  is only a convenience.
- Authentication, permission, Origin, and CSRF checks happen before replay
  lookup so one operator cannot probe another operation's stored response.
- Every privileged mutation is audited atomically. Reads are not audited unless
  a canonical requirement is added.
- Create/lifecycle commands use the reviewed command-specific replay window.
  Patches require expected version and return stale conflict rather than
  overwriting.
- Repository version predicates and database constraints remain authoritative;
  idempotency is retry safety, not a lock.

## Acceptance Criteria

1. Every contracted Event/Offering Operations route is registered under the
   existing versioned prefix with its exact read/manage permission.
2. Contract and runtime agree on DTO fields, nullable clearing, versions,
   expected versions, idempotency headers, limits, errors, and status codes.
3. Missing/expired authentication, missing permissions, invalid Origin, and
   invalid CSRF fail with no command, audit, outbox, or replay side effects.
4. Accepted commands commit state, one minimized audit row, required outbox
   row, and replay state atomically; required-write failure rolls everything
   back.
5. Same-key/same-request commands replay exactly; same-key/different-request
   conflicts; authorization always precedes both outcomes.
6. Stale versions and concurrent lifecycle/unique conflicts return stable
   `409` codes without raw PostgreSQL details.
7. Exact-origin CORS never reflects arbitrary origins or combines wildcard
   origins with credentials; same-origin operation remains valid with no CORS
   configuration.
8. Session response includes expiry and stable permissions; secrets and
   internal claims do not enter JSON, audit, replay hashes, or logs.
9. JSON bodies are type/size/media bounded, private responses are `no-store`,
   and canonical local/prod paths do not strip or duplicate `/api`.
10. No generic command bus, generated client, new auth system, bulk feature, or
   speculative abstraction is introduced.

## Testing

- Contract-shape tests for version, expected version, nullable clear/omit,
  idempotency headers, integer bounds, and error responses.
- Route tests for each path/method/permission and `404`/`405` behavior.
- Strict JSON tests: body limit, malformed JSON, unknown fields, trailing value,
  null, omission, boundary numbers, media type, and `413`/`415` mapping.
- Auth matrix: unauthenticated, expired, forbidden, allowed, invalid Origin,
  missing/mismatched CSRF, and safe read behavior.
- Idempotency matrix: first execution, same replay, hash conflict, expiry,
  failed transaction, actor/namespace isolation, and auth-before-replay.
- Audit/outbox assertions for accepted commands and rollback on required-write
  failure, with explicit secret-redaction checks.
- Expected-version and duplicate/active-state conflict mapping.
- No-op patch behavior: current response/version and zero mutation/audit/outbox
  effects.
- CORS tests for same-origin, allowed/disallowed exact origins, credentialed
  preflight, unsupported method/header, and `Vary: Origin`.
- Session conformance test for `expires_at`, deterministic permissions, and
  `Cache-Control: no-store`.
- Operations API-client/proxy tests for canonical business and root diagnostic
  paths without prefix stripping or duplication.

## Verification

```bash
cd apps/api && go run ./cmd/format -check
cd apps/api && go vet ./...
cd apps/api && go test ./internal/event/... ./internal/offering/...
cd apps/api && go test ./internal/platform/...
cd apps/api && go test ./internal/app/...
cd apps/api && go test ./...
cd apps/api && go build ./...
make validate
```

Also run the repository's pinned Operations OpenAPI lint/validation workflow
used by W1-05. Against disposable PostgreSQL, prove atomic rollback and
replay/concurrency behavior; mocks alone are insufficient.

## Deliverables

- Contract-aligned Operations Event/Offering API.
- Minimal shared strict-JSON, CORS, and outbox additions only where required by
  real routes.
- Focused authorization, CSRF/origin, concurrency, audit, replay, and contract
  tests.
- Current-state update and archived executed task only after verification.

## Risks and Clarifications to Review

- Confirmed default: create and lifecycle operations are retry-sensitive and
  require a stable idempotency key; patch uses expected version only.
- Confirmed default: credentialed CORS uses exact Operations origins and empty
  configuration means same-origin only.
- Confirmed default: CORS does not change the accepted `SameSite=Lax` cookie;
  cross-site credential deployment remains outside Week 2.
- Confirmed default: version conflicts return `409` and do not auto-merge.
- Confirmed default: nullable patch fields distinguish omission from explicit
  clearing.
- Remaining retention decision before BUILD: approve the proposed 24-hour
  replay window for Event/Offering create and lifecycle commands, or specify a
  different command-specific duration required by operations.
- Activation must resolve shared router/config edits with W2-04 in one pass to
  avoid order-dependent middleware or duplicated code.

## Final Report Requirements

Distinguish implemented endpoints, contract validation, PostgreSQL-backed
atomicity, auth/replay ordering, assumed deployment origins, deferred frontend
and administration work, plan variance, remaining security/concurrency risks,
and the next screen task.
