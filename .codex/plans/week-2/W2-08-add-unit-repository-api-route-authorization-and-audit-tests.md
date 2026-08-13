# Task: W2-08 Add Unit, Repository, API, Route, Authorization, and Audit Tests

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
- Tracker objective: Add unit, repository, API, route, authorization, and audit
  tests.
- Dependencies: W2-01 through W2-07.
- Week exit criterion: An authorized operator can publish an Event and
  Offering; the public sees only active/published data; historical
  pricing/configuration is preserved.

## Objective

Close Week 2 with an evidence-backed, cross-layer regression matrix proving
the Event/Offering lifecycle, persistence, public exposure, Operations
authorization, audit/replay atomicity, frontend route behavior, and known
security boundaries against the real stack.

This task fills verified coverage gaps. It does not postpone essential tests
from earlier tasks or create a second test framework.

## Context

The repository already runs frontend Vitest, Go tests/vet/build, migration
lifecycle checks on PostgreSQL 18, Compose validation, and container build in
CI. Database integration tests use `TEST_DATABASE_URL` and skip locally when it
is absent; CI reruns Go tests after migrations with that variable set.

Week 2 introduces meaningful lifecycle, concurrency, authentication, CORS,
rate-limit, audit, outbox, idempotency, and public-exposure behavior. Passing
unit tests alone cannot prove the database constraints or complete the tracker
exit criterion.

## Source of Truth

- all Week 2 plans and their executed task/final-review evidence;
- `docs/PRD.md`, `docs/PRODUCT_MAP.md`, `docs/ARCHITECTURE.md`,
  `docs/DECISIONS.md`, and `docs/CONVENTIONS.md`;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- `docs/security/PERMISSIONS.md` and `docs/security/AUTHENTICATION.md`;
- `docs/database/ERD.md`, `MIGRATION_PLAN.md`, and `REVIEW.md`;
- both reviewed OpenAPI contracts;
- current source, migrations, Makefile, Compose file, and
  `.github/workflows/validate.yml`;
- the live tracker rows for W2-01 through W2-08 and Week 2 exit criterion.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate only after W2-01 through W2-07 are implemented or explicitly identify
which dependency is incomplete. Start with a test inventory and trace each
acceptance criterion to existing evidence before adding tests. Do not commit or
push unless asked.

## Scope

### In Scope

- Audit existing Week 2 tests and add only missing domain, application,
  repository, migration, HTTP, route, auth, audit/outbox/idempotency, frontend,
  and security-boundary coverage.
- Run repository tests against disposable PostgreSQL 18 with all migrations
  applied and prove rollback/reapply of the Week 2 migration.
- Exercise real Gin route registration/middleware ordering, not handlers called
  directly as the only API evidence.
- Prove Storefront field minimization and active/published visibility.
- Prove Operations session/permission/Origin/CSRF/version/idempotency behavior.
- Prove audit, outbox, replay, and business state transaction atomicity.
- Prove frontend route builders, request construction, state rendering, and
  basic real-browser flows without adding a full E2E harness.
- Run format/lint/typecheck/test/build/Compose/container validation plus
  focused race, contract, secret, and dependency-vulnerability checks.
- Record failures, unavailable checks, assumptions, and deferred risk
  precisely; do not convert a skipped test into a pass.

### Out of Scope

- Arbitrary line-coverage percentages, snapshot-test volume, or tests that
  merely duplicate type checking.
- A new assertion framework, mock server, container orchestration library,
  browser E2E framework, test database manager, or CI service unless an actual
  untestable gap is demonstrated and separately approved.
- Load/soak tests, full WCAG audit, cross-browser matrix, penetration test,
  distributed-rate-limit proof, disaster recovery, or production monitoring;
  later roadmap work owns those broader validations.
- Fixing unrelated pre-existing failures or expanding the product feature
  surface. Newly discovered Week 2 defects are fixed in their owning task/layer
  and regression-tested, not hidden in test expectations.
- Claiming that automated checks guarantee the absence of every vulnerability.

## Existing State

- `make validate` runs frontend format/lint/typecheck/test/build, Go
  format/vet/test/build, and Compose config validation.
- CI provisions PostgreSQL 18, validates/applies migrations, runs reference
  seeds twice, reruns Go tests with `TEST_DATABASE_URL`, rolls back/reapplies one
  migration, builds the API image, and uses read-only repository permissions.
- The idempotency executor has unit and PostgreSQL integration tests; platform
  auth, transaction, request ID, config, migrations, and routes have focused
  tests.
- Frontend tests use Vitest, injected `fetch`, pure functions, and structural
  React checks; no DOM/E2E package is directly installed.
- The repository has no numeric coverage gate and no committed dependency
  vulnerability scanner workflow.

## Target State

- Every Week 2 acceptance criterion maps to at least one appropriately placed
  test or a documented manual/browser/database verification step.
- The Week 2 historical-pricing/configuration exit criterion is evidenced by
  append-only Event/Offering audit before/after data. Purchase-specific applied
  price/configuration snapshots remain W3-04 and are not falsely claimed here.
- Domain tests prove complete lifecycle matrices and validation boundaries;
  repository tests prove constraints, version predicates, deterministic
  queries, public visibility, availability, and concurrency on PostgreSQL.
- Router-level tests prove exact methods/paths/middleware order and stable
  envelopes for both APIs.
- Security tests prove denial paths have no business/audit/outbox/replay side
  effects and no sensitive response/log content.
- Frontend tests prove exact endpoint/headers/path builders and presentation
  boundaries; real-browser smoke proves the thin workflows against the real
  API.
- CI continues to run all committed tests without hidden network dependencies
  or production credentials.
- Week 2 closes only if the full validation set passes and no known unresolved
  critical/high dependency or implementation vulnerability remains. Findings
  are fixed or explicitly block completion; they are never silently waived.

## Constraints

- Put tests next to the owning production code. Do not create a monolithic
  `week2_test` package or duplicate helpers across modules without evidence.
- Prefer Go standard `testing`/`httptest`, current Vitest, injected `fetch`,
  existing PostgreSQL/Compose, and current browser smoke capabilities.
- Test observable contracts and invariants, not private function arrangement
  or SQL text formatting.
- Use fixed clocks/IDs only through existing injectable boundaries or the
  smallest function parameter needed for deterministic behavior. Do not add a
  general dependency-injection framework for tests.
- Integration fixtures use unique owned identifiers and targeted cleanup.
  Never globally `TRUNCATE` shared tables or assume package execution order.
- Use transactions/rollback for single-connection cases. Multi-connection
  concurrency fixtures may commit setup but must delete only their own rows in
  dependency-safe order.
- Secret/vulnerability scans are evidence, not proof of total security. Record
  exact tool/version/database date and distinguish findings from false
  positives or unavailable network checks.

## Coverage Matrix

| Boundary | Required proof |
| --- | --- |
| Event domain/application | All accepted/rejected transitions, configuration freeze policy, validation, version conflict, audit/outbox/replay intent. |
| Offering domain/application | Lifecycle/edit policy, parent Event guards, exact price/capacity/quota boundaries, publication time, advisory availability semantics. |
| Migration/repository | Empty apply, down/up, version defaults, constraints, deterministic cursors, visibility SQL, availability aggregation, concurrent activation/update. |
| Public API | Three exact routes, active/published filtering, `data: null`, indistinguishable `404`, nullable availability, rate limit, recovery, safe errors/fields. |
| Operations API | Exact CRUD/lifecycle routes, strict JSON, pagination, expected version, null-vs-omit patch, conflict/status mapping. |
| Authorization/security | `401`/`403`, exact permissions, Origin/CSRF, auth-before-replay, trusted proxy, exact CORS, request bounds, no sensitive leakage. |
| Audit/outbox/idempotency | Exactly-once accepted effects, minimized fields, actor/source/request ID, same-request replay, different-request conflict, rollback on required-write failure. |
| Operations web | Central paths, session/CSRF memory model, exact headers/keys/versions, all data states, no optimistic contested state, accessibility smoke. |
| Storefront web | Central paths, exact public requests, registration/price/availability presentation, no Operations fields/fake CTA, PWA/API exclusion, accessibility smoke. |

## Implementation Requirements

### Domain and application tests

- Table-test every Event lifecycle source/target pair, including activation
  from published/suspended, terminal states, and configuration edit guards.
- Table-test every Offering lifecycle pair, parent Event compatibility,
  published-before-edit guard, first-publication timestamp preservation, and
  archive immutability.
- Cover trimmed/empty/oversized strings, year/window ordering, exact safe
  integer boundaries, price/currency/capacity/quota validation, null clearing,
  and unavailable/finite/unbounded availability.
- Prove application services return stable typed errors and do not audit/emit
  after rejected validation/transition/version checks.

### Migration and PostgreSQL repository tests

- Validate all migration pairs, apply from empty, verify schema/version
  columns/constraints, roll down exactly the Week 2 migration, and reapply.
- Event repository: create/get/list/update, duplicate year, stale version,
  deterministic cursor pages, malformed cursor, inserted-between-pages
  behavior, no unbounded list, and one-active constraint.
- Offering repository: event-scoped uniqueness, create/get/list/update, stale
  version, deterministic cursor, public visibility matrix, and availability
  aggregation for every reservation state and nullable quota combination.
- Run two real database connections for competing activation and same-version
  update. Assert final rows and translated conflict, not only returned errors.
- Prove transaction rollback when audit/outbox/replay fails after a repository
  write.

### Public API and route tests

- Exercise the composed Gin engine for all three public methods/paths, malformed
  UUID, unexpected method, no active Event, empty catalogue, hidden statuses,
  unavailable detail, dependency failure, panic recovery, and request IDs.
- Assert an allowlist of public JSON keys and explicit absence of version,
  internal quota, reservation, audit, operator/session, and SQL fields.
- Verify direct IP/trusted-proxy handling, spoofed forwarded headers, limiter
  allow/exhaust/recover behavior, bounded entry cleanup, `Retry-After`, and
  standard `429` envelope.
- Verify no public read writes audit, outbox, or idempotency state.

### Operations API and route tests

- Exercise the composed router for every Event/Offering list/get/create/patch
  and lifecycle operation with exact permission fixtures.
- Cover body-size limits, media type, malformed/empty/trailing JSON, unknown
  fields, boundary numbers, nullable clear vs omission, invalid UUID/cursor,
  and syntactic `400` vs semantic `422` behavior.
- Cover unauthenticated, expired/revoked session, forbidden permission,
  disallowed/missing Origin, missing/mismatched CSRF, allowed request, and safe
  read behavior.
- Assert denial happens before repositories, audit, outbox, or replay lookup.
- Cover stale version, duplicate year/code, active/state conflict, same-key
  replay, different-payload conflict, namespace/actor isolation, expired key,
  and failed-command retry.
- Verify exact-origin CORS preflights, credentials, allowed methods/headers,
  `Vary: Origin`, and rejection of arbitrary/wildcard credentialed origins.

### Audit, outbox, and replay assertions

- For every privileged command, assert exactly one audit row with trusted
  actor, permission, source, request ID, target, action, and minimized
  before/after values.
- Assert raw cookies, tokens, CSRF, encryption keys, complete claims, and
  unrelated personal data are absent from audit, outbox, replay bodies,
  responses, and captured logs. The idempotency key appears only in the
  canonical `idempotency_records.idempotency_key` column, never duplicated into
  those payloads or logs.
- Assert required Event/Offering lifecycle outbox type/payload and transaction
  visibility.
- Assert successful replay returns the stored status/body and creates no second
  business, audit, or outbox effect.
- Force each required writer to fail and prove the whole transaction rolls
  back. External post-commit consumers are not simulated in Week 2.

### Frontend tests and smoke

- Operations: exact request URLs/credentials/CSRF/idempotency/version payloads,
  stable query keys/invalidation, permission affordances, conflict preservation,
  integer/time conversion, central routes, and no secret persistence.
- Storefront: exact public URLs, active-null/empty/404/429/503 states, route
  builders, registration boundaries, exact minor-unit text, availability
  labels, public field allowlist, and PWA/API exclusion.
- Reuse pure/structural Vitest coverage and injected fetch. Do not add a DOM
  framework merely to shallow-render components.
- Run real-browser smoke for both applications at desktop/mobile width,
  keyboard-only navigation, direct route load, mutation pending/conflict, and
  public empty/populated/error states. Record test data and permissions used.

### Contract and security verification

- Format and parse both OpenAPI 3.1 documents with the same pinned Redocly
  workflow used in W1-05; reconcile all new warnings/errors rather than treating
  YAML parsing as contract validation.
- Trace every implemented operation ID to one router-level test and every
  runtime response field to the correct public/Operations schema.
- Run focused Go race tests for Event/Offering/platform code and real database
  concurrency tests.
- Run a scoped secret-pattern scan over changed/source files and inspect hits;
  do not print real environment values into evidence.
- Run current Go and production pnpm dependency vulnerability checks with exact
  tool versions recorded. Network/database unavailability is reported, not
  described as a clean result.
- Review trusted-proxy, CORS, CSRF, rate-limit, response-minimization, SQL
  parameterization, strict JSON, PWA cache, and log/audit redaction manually
  against the implemented source.

## Planned File Changes

Prefer extending focused tests created in W2-01 through W2-07. The likely
ownership map is:

| File | Action | Purpose |
| --- | --- | --- |
| `apps/api/internal/event/domain_test.go` and `service_test.go` | Modify | Complete lifecycle, validation, version, and side-effect matrix. |
| `apps/api/internal/offering/domain_test.go` and `service_test.go` | Modify | Complete lifecycle, parent-state, price/quota, and availability matrix. |
| `apps/api/internal/event/postgres_integration_test.go` | Modify | Event constraints, cursors, versions, concurrency, rollback. |
| `apps/api/internal/offering/postgres_integration_test.go` | Modify | Offering visibility, availability, cursors, versions, concurrency. |
| `apps/api/internal/event/public_http_test.go` and `operations_http_test.go` | Create/modify | Public/Operations Event route and exposure matrix. |
| `apps/api/internal/offering/public_http_test.go` and `operations_http_test.go` | Create/modify | Public/Operations Offering route and exposure matrix. |
| `apps/api/internal/app/server_test.go` | Modify | Prove composed registration, middleware order, methods, recovery, and health isolation. |
| `apps/api/internal/platform/auth/*_test.go` | Modify | Session/permission/Origin/CSRF and session response conformance. |
| `apps/api/internal/platform/audit/writer_test.go` | Create/modify if W2-05 has uncovered writer behavior | Atomic insert, verified client IP, minimization, and failure behavior. |
| `apps/api/internal/platform/outbox/writer_test.go` | Modify | Transaction/event payload/failure behavior. |
| `apps/api/internal/platform/idempotency/*_test.go` | Modify only for uncovered integration behavior | Auth ordering stays at route tests; executor retains replay atomicity tests. |
| `apps/api/internal/platform/cors/*_test.go`, `ratelimit/*_test.go`, and `httpx/*_test.go` | Modify | Exact origins, bounded limiter, strict JSON, and safe recovery. |
| `apps/operations-web/src/api/*.test.ts`, route tests, and focused page/helper tests | Modify/create | Operations requests, paths, forms, query states, and security invariants. |
| `apps/storefront-web/src/api/*.test.ts`, route tests, and focused page/helper tests | Modify/create | Public requests, paths, presentation, exposure, and PWA boundary. |
| `.github/workflows/validate.yml` | Modify only if committed tests are not already exercised | Preserve least privilege and PostgreSQL lifecycle; add no redundant job. |
| `.codex/CURRENT_STATE.md` and `.codex/TASK.md` | Modify/create/archive when executed | Record final Week 2 evidence and remaining risk. |

No permanent coverage dashboard, fixture framework, or new test-only production
route is planned.

## Dependencies and Sequencing

- Earlier tasks must land their own critical tests; begin W2-08 by identifying
  gaps rather than rewriting them.
- Run fast unit/contract tests before database concurrency and browser smoke.
- Fix any discovered defect at the owning domain/repository/API/UI boundary,
  then add the smallest regression that would have caught it.
- Run the full repository/CI-equivalent verification only after focused tests
  pass.

## API and Database Impact

- This is primarily a verification task. It should not change API or schema
  behavior merely to make tests easier.
- Any legitimate contract/schema correction discovered must amend its canonical
  document/migration plan and be called out as plan variance before completion.
- Tests operate only on disposable/local data and clean their owned fixtures.

## Security and Evidence Policy

- A green test suite is not proof of zero vulnerabilities. The completion claim
  is limited to the reviewed threat boundaries and no known unresolved
  critical/high findings from the recorded checks.
- Do not weaken auth, TLS/cookie defaults, CORS, CSRF, rate limits, input bounds,
  database constraints, or audit requirements to simplify a test.
- Do not commit scanner caches, reports containing environment details,
  coverage output, browser session data, or test credentials.
- If a security check cannot run, record the exact reason and keep the relevant
  acceptance criterion unverified rather than silently accepting the gap.

## Acceptance Criteria

1. The coverage matrix maps every W2-01 through W2-07 acceptance criterion and
   the Week 2 tracker exit criterion to passing automated or recorded manual
   evidence.
2. Event/Offering configuration and price changes leave append-only minimized
   audit history; the final report clearly separates this from W3-04 Purchase
   snapshots and deferred audit-retention policy.
3. Complete Event/Offering lifecycle, validation, edit-freeze, version, and
   availability matrices pass at the owning domain/application layer.
4. Migration apply/down/up, constraints, deterministic queries, public
   visibility, availability, transaction rollback, and real concurrency pass
   on disposable PostgreSQL 18.
5. Public router tests prove only active/published data is exposed with correct
   null/404/rate-limit/recovery behavior and a strict public field allowlist.
6. Operations router tests prove exact routes/permissions, auth/Origin/CSRF
   ordering, strict input, versions, idempotency, atomic audit/outbox, and safe
   conflict/error mapping.
7. Denied/failed/replayed commands have the exact expected zero/one side-effect
   counts and no secret/internal data leakage.
8. Operations and Storefront tests plus real-browser smoke prove centralized
   routes, exact API requests, required states, basic accessibility/responsive
   behavior, and no API service-worker caching.
9. Both OpenAPI contracts parse/validate with the pinned workflow and match
   runtime paths/fields/statuses.
10. Focused race, secret, and dependency-vulnerability checks complete with no
   unresolved known critical/high finding; unavailable checks remain explicit
   blockers or documented unverified risk, not false passes.
11. `make validate`, full Go vet/test/build, Compose config, migration/seed
    lifecycle, and API container build pass without skipped CI database tests.
12. No arbitrary coverage quota, redundant framework, global database cleanup,
    test-only auth bypass, or unrelated product feature is added.

## Verification

Run against an explicitly disposable PostgreSQL 18 database and record command
versions/results:

```bash
make validate
cd apps/api && go vet ./... && go test ./... && go build ./...
cd apps/api && go test -race ./internal/event/... ./internal/offering/... ./internal/platform/...
cd apps/api && go run ./cmd/db migrate validate
cd apps/api && go run ./cmd/db migrate up
cd apps/api && TEST_DATABASE_URL="$DATABASE_URL" go test ./...
cd apps/api && go run ./cmd/db migrate down --steps 1
cd apps/api && go run ./cmd/db migrate up --steps 1
docker compose -f infrastructure/compose.yaml config
docker build -f apps/api/Dockerfile apps/api
```

Also run the pinned OpenAPI validation, scoped secret inspection, and current
dependency vulnerability checks described above. Use the existing CI workflow
as the reproducible reference and confirm integration tests did not skip in
the database-enabled run.

## Deliverables

- Completed focused regression matrix in owning test files.
- PostgreSQL, router/auth/audit/replay, contract, frontend/browser, race, and
  security-check evidence.
- Defect fixes only where a Week 2 test proves the owning behavior was wrong.
- Final current-state update and archived executed task after all required
  evidence passes.

## Risks and Clarifications to Review

- Confirmed default: earlier tasks retain their critical focused tests; W2-08
  audits and closes gaps instead of becoming a test dump.
- Confirmed default: no new E2E/DOM/coverage framework unless a concrete
  untestable interaction is demonstrated and approved.
- Confirmed default: process-local rate-limit behavior is tested per replica;
  uniform multi-replica ingress enforcement remains deployment work.
- Confirmed default: “no vulnerability gap” is reported honestly as no known
  unresolved critical/high finding within the reviewed/scanned scope, never as
  an absolute guarantee.
- Append-only audit behavior is verified, but long-term retention/export and
  backup policy remains a separately deferred compliance/operations decision.
- Frontend CSP, HSTS, frame protection, and other edge response headers require
  the eventual hosting/ingress owner; no production web-host configuration is
  present in this repository. Record that deployment risk rather than claiming
  browser headers were verified from Vite source.
- Week 2 proves transactional outbox writes, not publication/cleanup; monitor
  backlog size in any environment that executes commands before a later outbox
  consumer is delivered.
- Execution needs a disposable PostgreSQL target and a test operator/session
  fixture with exact permissions; browser login smoke also needs a configured
  test OIDC identity/provider. It must not use shared or production data or add
  a production auth bypass.
- The tracker provides one 42-hour Week 2 allocation and no per-task estimates.
  Confirm whether that is one engineer's hard cap or team capacity before
  scheduling; do not silently drop security/verification controls to force a
  guessed allocation.

## Final Report Requirements

Distinguish implemented regression coverage, exact passing evidence, skipped or
unavailable checks, assumptions/test fixtures, defects fixed, deferred Week
5/7 validation, known residual risks, plan variance, Week 2 exit-criterion
status, and the next dependency-ready Week 3 task.
