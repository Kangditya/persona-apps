# Task: W3-09 Add Validation, Authorization, Idempotency, and Quota-Contention Tests

## Executed

## Status

Verified and executed on 2026-09-01. The user explicitly approved execution on
2026-08-31 and this plan was copied to a new `.codex/TASK.md`. W3-07 and W3-08
are implemented and verified. No commit or push was requested.

## Tracker

- Week: Week 3
- Epic: Party & Purchase
- Application: Cross-cutting
- Category: Test
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Add validation, authorization, idempotency, and
  quota-contention tests.
- Dependencies: W3-03, W3-05, W3-06, W3-07, and W3-08.
- Acceptance summary: tests prove invalid Purchase requests fail safely,
  Operations reads enforce authorization, retries do not duplicate effects,
  and concurrent checkout cannot exceed quota across the API and UI
  boundaries.

## Objective

Audit the evidence produced by W3-01 through W3-08 and close only the remaining
high-value Common Purchase regression gaps at their owning layers. Finish Week
3 with one traceable real flow:

```text
published Offering
-> explicit-role Storefront checkout
-> atomic pending Purchase and reservation
-> safe confirmation
-> authorized Operations list/detail
```

This is a gap-closing verification task, not a test dump. Existing focused
tests remain the primary evidence; add the smallest regression that proves each
uncovered boundary and fix production code only when a new test demonstrates a
real defect.

## Context

The backend already has domain, repository, PostgreSQL, handler, and composed
route tests for Party roles, Purchase construction, historical snapshots,
reservation semantics, last-unit contention, unique references, one-time token
hashing, public checkout replay/conflict, outbox minimization, and Operations
Purchase reads. W3-07 and W3-08 are expected to add focused API/form/path/query
tests and real-browser smoke for the two frontends.

Remaining Week 3 proof must be based on the final implemented surfaces. It must
not rewrite already-passing tests, declare skipped database tests green, use a
test-only authorization bypass, or substitute mocked UI success for the
composed PostgreSQL-backed flow.

## Source of Truth

- `docs/PRD.md`, especially Common Purchasing, Party distinctions, Storefront
  and Operations scope, security, privacy, reliability, and data integrity;
- `docs/PRODUCT_MAP.md` and the Common Purchase vertical slice;
- `docs/ARCHITECTURE.md`, API/frontend boundaries, transactions,
  authorization, caching, concurrency, and verification strategy;
- `docs/DECISIONS.md`, especially ADR-002, ADR-010 through ADR-014, ADR-018
  through ADR-020, ADR-024, ADR-033, ADR-038, ADR-039, ADR-041 through ADR-050;
- `docs/CONVENTIONS.md`;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- `docs/security/AUTHENTICATION.md` and
  `docs/security/PERMISSIONS.md`;
- `docs/database/ERD.md`, `docs/database/MIGRATION_PLAN.md`, and migrations
  0001 through 0006;
- `contracts/openapi/storefront.yaml` and
  `contracts/openapi/operations.yaml`;
- completed W3-01 through W3-06 execution evidence and the implemented W3-07
  and W3-08 outputs;
- current backend, frontend, shared-client, PWA, CI, and test source;
- `.codex/CURRENT_STATE.md` and `MVP-DELIVERY-ROADMAP.md`.
- the user-supplied
  [QurbanPlus Penpot file](https://design.penpot.app/#/workspace?team-id=81f57451-85cc-819d-8008-7c2bac979fc9&file-id=81f57451-85cc-819d-8008-7c2c1dbf6c2a&page-id=81f57451-85cc-819d-8008-7c2c1dbf6c2b),
  specifically pages
  `13 · Storefront · 04 Checkout Participant`,
  `16 · Storefront · 07 Purchase Confirmation`, and
  `24 · Operations · 05 Purchasing Queue` as visual verification references,
  with canonical product/API/security artifacts taking precedence.

## Penpot Design Evidence and Test Boundary

- Checkout and confirmation each provide one `862 × 1650` single-column board
  with buyer/intended-participant/package sections, minimum-data/help content,
  and safe confirmation reference/status/summary hierarchy.
- The confirmation mockup includes payment instructions and automatic payment
  verification copy that W3-07 must omit because those capabilities are not
  implemented.
- Operations provides one `1440 × 960` skeletal desktop board with shell,
  summary-card, filter, table, and lower-panel regions, but no complete business
  copy, mobile board, or detail design.

W3-09 verifies artifact-safe conformance, not pixel identity: correct section
order, semantics, responsive behavior, accessibility, token/privacy boundaries,
and explicit absence of unsupported payment/KPI/mutation claims. It does not
add a screenshot-diff framework or treat blank design placeholders as product
requirements.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

At activation:

1. Re-read W3-07 and W3-08 final reports and current diffs.
2. Map every Week 3 acceptance criterion to existing automated/manual evidence.
3. Mark gaps before changing a test.
4. Add or strengthen only the owning tests needed for those gaps.
5. Record a PLAN VARIANCE before fixing any production defect or contract
   mismatch discovered by the new evidence.

## Execution Reconciliation — 2026-08-31

The approved plan was reconciled against the canonical documents, migrations
0001 through 0006, both OpenAPI contracts, current source, W3-01 through W3-08
archives, and the final W3-07/W3-08 diffs. The baseline focused suites passed
against the disposable PostgreSQL database `persona_apps_w309_01a056ca`:

- `go test ./internal/purchasing/... ./internal/app/... -count=1`;
- Storefront: 10 files / 31 tests;
- Operations: 9 files / 21 tests.

| Week 3 owner | Existing primary evidence | W3-09 gap or reuse decision |
| --- | --- | --- |
| W3-01 Party | `identity` domain/repository/PostgreSQL tests prove validation, nullable contacts, rollback, and explicit UUID reuse. | Reuse; composed invalid checkout will additionally prove no Party side effect. |
| W3-02 roles | `relationships_test.go`, Purchase persistence tests, Storefront request examples, and checkout mapping tests prove shared/distinct/name-only roles and ordering. | Reuse; include relationship-invalid zero-effect coverage. |
| W3-03 Purchase/read | Purchase domain/repository/handler/PostgreSQL tests and W3-08 browser evidence prove canonical persistence, filters/cursors, list/detail, and `purchase.read`. | Add composed denial-before-Purchase-dependency and exact exposure assertions. |
| W3-04 snapshots/totals | Snapshot unit/PostgreSQL tests and frontend response guards prove historical values and checked per-participant totals. | Reuse unchanged. |
| W3-05 reservation | Reservation tests prove Event-then-Offering locking, counted/terminal states, rollback, unbounded quotas, and a three-run repository-level last-unit race. | Add a bounded three-run distinct-intent race at the composed public route. |
| W3-06 checkout | Composed PostgreSQL, reference-collision, idempotency, CORS, token/hash, outbox, and encrypted durable-replay tests prove most command behavior. | Add zero-effect trust-boundary cases, semantic JSON canonicalization, exact allowlists, and strengthen failed-attempt effects. |
| W3-07 Storefront | API/role/retry/PWA tests and recorded desktop/360px browser evidence prove exact public transport, form states, hierarchy, and network-only caching. | Add edited-intent/error cases and prevent the raw token from being retained in TanStack Mutation state. |
| W3-08 Operations | API/query/path/PWA tests and recorded OIDC/browser/PostgreSQL evidence prove private reads, filters/cursors, roles/snapshots, logout cleanup, hierarchy, and responsive layout. | Strengthen exact GET header/detail exposure assertions and repeat the combined browser journey. |

### PLAN VARIANCE — discard the raw Purchase token at the Storefront API boundary

- Planned work: prove that the raw checkout token is absent from DOM, URL,
  storage, logs, PWA caches, and frontend query state.
- Unexpected evidence: `createPurchasesApi().create` returns `accessToken`, so
  `useMutation` retains the raw token in its mutation result even though no
  implemented feature consumes it.
- Necessity: retaining the credential conflicts with the approved W3-09 token
  boundary and W3-07's safe-confirmation claim.
- Impact: parse and validate the contracted token, then discard it and return
  only the safe `CreatedPurchase` view to the caller. Update the existing API
  test and Checkout screen; there is no server/OpenAPI/database behavior change.
- Scope: this is the smallest owning fix and remains within W3-09 privacy and
  exposure acceptance criteria.

### PLAN VARIANCE — normalize the existing Week 3 Go slice

- Planned work: finish with a passing `make validate` and formatting check.
- Unexpected evidence: the baseline formatter check reports exactly 18 files,
  all in the existing W3 Purchase test/implementation slice.
- Necessity: leaving them unchanged makes the task's required repository
  validation impossible to pass.
- Impact: after functional edits, run the repository formatter only on those
  18 reported files and review non-target files as formatting-only changes.
  No behavior, contract, migration, dependency, or generated output changes.
- Scope: the variance is bounded to the formatter's existing W3 list and the
  task's mandatory verification gate.

## Scope

### In Scope

- A concise Week 3 coverage matrix that references existing tests and identifies
  the exact missing proof.
- Public checkout trust-boundary tests for the remaining request, relationship,
  response, error, CORS, and secret-exposure cases.
- Operations Purchase read tests for authentication, `purchase.read`,
  filters/cursors, detail exposure, and denial before repository access.
- Durable idempotency tests at the composed checkout boundary: exact semantic
  replay, payload conflict, concurrent duplicate, failed-work retry, and
  zero/one side-effect counts.
- Real PostgreSQL contention at the composed public route using distinct
  intents for the same final quota units, in addition to the existing
  reservation-layer lock test.
- Storefront checkout tests for exact role payloads, frozen retry intent, safe
  errors, confirmation fields, and raw-token non-exposure/non-persistence.
- Operations tests for exact credentialed reads, query isolation, permission
  states, snapshots, Party roles, and no CSRF/idempotency headers on GET.
- Penpot-informed structural and browser checks for the accepted checkout,
  confirmation, and Operations information hierarchy plus all documented
  design-to-contract omissions.
- Contract/runtime exposure checks, focused race coverage, real-browser
  end-to-end smoke, and complete repository validation.
- Narrow production defect fixes only when the new test fails for behavior
  already required by these artifacts.

### Out of Scope

- An arbitrary line/branch coverage percentage, snapshot volume, or tests that
  only duplicate type checking.
- A new assertion, mock-server, fixture, database-manager, DOM, browser-E2E,
  load, or CI framework without a demonstrated untestable gap and separate
  approval.
- Payment, evidence, cancellation, Purchase tracking GET, token recovery,
  activation, Saving, Giveaway, W6 search, or later lifecycle tests.
- Full WCAG/cross-browser matrix, penetration test, uniform multi-replica rate
  limiting, load/soak, backup/restore, production deployment, or generalized
  vulnerability claims.
- Test-only routes, weakened middleware, fake auth bypasses, disabled
  constraints, global database truncation, production/shared data, or
  committed credentials.
- Refactoring working production code solely to make a preferred test style
  possible.

## Existing Evidence to Reuse

- `purchase_test.go` and `relationships_test.go` cover Purchase invariants,
  checked totals, explicit shared/distinct Party roles, participant order, and
  invalid relationship forms.
- `postgres_test.go` and `postgres_integration_test.go` cover Purchase/history/
  participant atomic persistence, snapshots, Party summaries, filters, cursors,
  and rollback.
- `snapshot_*_test.go` proves stored commercial/participant history remains
  stable after source changes.
- `reservation_*_test.go` covers state/window/quota semantics and a real
  PostgreSQL concurrent last-unit race.
- `public_http_*_test.go` covers reference format/collision, role mapping,
  token shape/hash, and parts of public command composition.
- `apps/api/internal/app/public_purchase_integration_test.go` covers concurrent
  same-key replay, changed-payload conflict, sequential quota rejection, CORS,
  response/stored effects, durable encrypted replay, and secret-minimized
  outbox.
- `apps/api/internal/app/purchase_integration_test.go` covers permitted
  Operations list/detail and forbidden reads on PostgreSQL.
- Platform auth, idempotency, transaction, HTTP error, CORS, rate-limit, and PWA
  tests already own their generic behavior.
- W3-07/W3-08 must contribute focused frontend tests before this task begins.

Do not reproduce these cases merely to increase test count. Strengthen them
only when the final vertical-slice matrix identifies a missing assertion.

## Target State

- Every W3-01 through W3-08 acceptance criterion maps to a named passing
  automated test or a recorded real-browser/database check.
- Invalid input is rejected at the correct boundary with no Party, Purchase,
  history, participant, reservation, outbox, or replay effect.
- Unauthenticated and forbidden Operations reads are denied before Purchase
  repository access; a permitted read returns only the Operations contract.
- One semantic checkout intent commits once. Same-intent retries return the
  original response/token; changed intent conflicts; failed work remains
  safely retryable.
- Two different checkout intents racing for the same last participant unit
  produce exactly one committed Purchase/reservation and one stable
  `quota_unavailable` result.
- Storefront and Operations UI tests prove their respective public/private
  transport boundaries, explicit states, role presentation, and cache/security
  constraints.
- The real browser demonstrates checkout to confirmation and Operations
  inspection against one disposable PostgreSQL-backed API without exposing the
  token or using database access in the frontends.
- All required commands pass without hidden integration-test skips. Any
  unavailable check remains explicitly unverified.

## Constraints

- Put each test beside the production behavior it protects.
- Prefer Go `testing`/`httptest`, Vitest, injected `fetch`, existing query
  helpers, current browser control, PostgreSQL 18, and current Compose/CI
  workflows.
- Test observable contracts and invariants, not private function layout or SQL
  string formatting.
- Use unique fixture prefixes and targeted dependency-safe cleanup. Single-
  connection cases may use rollback; multi-connection races clean only owned
  committed rows.
- Assert final database and response effects, not only returned status codes.
- Use deterministic clocks/entropy only through existing narrow seams. Do not
  add a dependency-injection container.
- Do not print environment secrets, raw sessions, CSRF, Purchase tokens,
  idempotency keys, full contacts, or replay encryption keys in test failure
  output.
- A green suite is evidence for the reviewed boundaries, not a claim that no
  vulnerability can exist.

## Coverage Matrix

| Boundary | Required proof |
| --- | --- |
| Public request validation | Required media type/key/body, 64 KiB bound, one JSON object, unknown/trailing fields, canonical UUID, Party declaration/reference exclusivity, participant order/count/string bounds, and safe `400/413/415/422`. |
| Checkout response/exposure | Strict public field allowlist, stored snapshots/total/expiry, one-time 43-character token only on initial/replayed POST, no hash/internal topology/contact leakage, `no-store`, request ID, and exact CORS. |
| Idempotency | JSON formatting/key order produces the same semantic hash; same key/request replays exact bytes; different semantic input conflicts; concurrent same key has one effect; failed required work leaves no completed replay and retries safely. |
| Quota contention | `RESERVED`/`CONSUMED` count, terminal states do not, Event then Offering lock order, per-Purchase capacity, exact last unit, two distinct public intents race, one success, no overflow/partial effects. |
| Operations authorization | Missing/expired/revoked session `401`, missing `purchase.read` `403`, permission succeeds, denial precedes repository/dependency access, no CSRF/idempotency on GET, no public-token crossover. |
| Operations queries | Exact Event/status filters, limit/cursor bounds, deterministic pagination, invalid UUID/cursor `400`, detail `404`, summary/detail field difference, explicit purchaser/payer/participant snapshots. |
| Storefront UI | Exact role-to-payload mapping, no automatic mutation retry, frozen same-intent key, edited-intent new key, all request states, safe confirmation, token absent from DOM/URL/storage/log/PWA/query cache, keyboard/mobile behavior. |
| Operations UI | Session/access states, exact credentialed list/detail calls, query-key isolation, filters/cursor reset, safe snapshots/roles, logout private-cache removal, no mutation affordance, keyboard/mobile behavior. |
| Contracts and integration | Both OpenAPI documents validate; runtime paths/status/fields match; real Next.js clients use the real Go API/PostgreSQL; PWA remains network-only for API/auth/business data. |
| Penpot reconciliation | Storefront section/status/summary/help hierarchy and Operations filter/results hierarchy render at 862px/1440px and adapt at 360px; unsupported payment CTA/copy, KPI cards, totals, and mutations remain absent. |

## Implementation Requirements

### Validation and side effects

- Add table-driven handler cases only for validation branches not already
  proven by lower layers or the composed route.
- For every rejected pre-transaction request, assert no business/replay rows.
- For injected transactional failures, assert all owned Party/Purchase/history/
  participant/reservation/outbox/replay rows are absent and a corrected retry
  can succeed.
- Keep syntactic `400`, size `413`, media `415`, semantic `422`, state/quota/
  idempotency `409`, rate `429`, dependency `503`, and unexpected `500`
  outcomes distinct.

### Idempotency and quota

- Compare full first/replayed status, safe headers, and response bytes, including
  the same one-time token.
- Prove semantic canonicalization with equivalent JSON formatting/key order and
  conflict with one meaningful field change.
- Add one composed public-route race with separate high-entropy keys competing
  for the last unit. Assert one `201`, one `409 quota_unavailable`, exactly one
  Purchase/reservation/history/replay set, and configured quota not exceeded.
- Repeat the race enough times to detect lock-order regressions while keeping
  the check bounded and reliable. Record the repetition count.
- Preserve the existing Event-then-Offering lock order and real PostgreSQL
  requirement; mocks cannot prove contention.

### Authorization and exposure

- Use the real composed Operations route/middleware with owned session fixtures.
- Prove unauthenticated/forbidden requests return before a missing or failing
  Purchase dependency would surface.
- Assert list/detail allowlists and explicit absence of access-token hash/raw
  token, idempotency, reservation, outbox, audit, session, CSRF, SQL, and
  unrelated Party fields.
- Scan captured logs/responses/database JSON for owned secret sentinel values
  without printing those values on failure.

### Frontend and browser

- Extend W3-07/W3-08 focused tests rather than adding a parallel Week 3 suite.
- Verify the Storefront POST is non-credentialed and Operations GETs are
  credentialed; neither surface receives the other's secret/header model.
- Use pure/structural tests for role mapping, form/filters, query keys, safe
  presentation, and path builders; use injected fetch for transport.
- Run one real-browser journey with disposable data: Offering detail ->
  checkout -> safe confirmation -> authenticated Operations list/detail ->
  logout. Include validation, quota conflict, forbidden operator, direct route,
  keyboard, and 360px checks.
- At `862px` and `1440px`, compare the implemented information hierarchy with
  the corresponding Penpot boards. At `360px`, verify the repository-defined
  accessible stacking behavior because Penpot supplies no mobile board.
- Inspect URL, rendered text, browser storage, logs, network requests, and
  generated service-worker policy for Purchase-token leakage.
- Assert the Storefront does not render the Penpot-only payment-instructions
  action/verification promise and Operations does not render placeholder KPI,
  total, or mutation controls.

## Planned File Changes

The final gap matrix decides the exact subset. Expected ownership is:

| File | Action | Purpose |
| --- | --- | --- |
| `apps/api/internal/app/public_purchase_integration_test.go` | Modify | Close composed validation, semantic replay, distinct-intent contention, rollback, and exposure gaps. |
| `apps/api/internal/app/purchase_integration_test.go` | Modify | Close Operations authentication/permission ordering, filter/cursor, detail, and exposure gaps. |
| `apps/api/internal/purchasing/*_test.go` | Modify only for identified gaps | Strengthen owning domain/repository/reservation assertions without duplication. |
| `apps/storefront-web/src/api/purchases.test.ts` | Modify | Complete exact public transport, guarded response, and error coverage. |
| `apps/storefront-web/src/features/purchases/checkout.test.ts` | Modify | Complete role, validation, frozen-intent, and token-boundary coverage. |
| `apps/storefront-web/src/routes/paths.test.ts` and focused screen/structure tests | Modify/create only if needed | Prove direct checkout and safe confirmation UI boundaries with the existing stack. |
| `apps/operations-web/src/api/purchases.test.ts` | Modify | Complete exact private read, filters/cursors, response guard, and error coverage. |
| `apps/operations-web/src/features/purchases/queries.ts` and focused tests | Modify/create only if needed | Prove query-key/filter isolation and private-root cleanup compatibility. |
| `apps/operations-web/src/routes/paths.test.ts` and focused screen/structure tests | Modify/create only if needed | Prove direct detail, role/snapshot states, and no mutation/token content. |
| Production files named above | Modify only after a failing required test and recorded PLAN VARIANCE | Fix the owning defect with the smallest coherent change. |
| `.codex/CURRENT_STATE.md` | Modify after verified execution | Record the complete Week 3 vertical-slice evidence and remaining lifecycle gaps. |

No new migration, API operation, dependency, permanent test harness, coverage
service, or CI job is planned.

## API, Database, Authorization, and Dependency Impact

- API/OpenAPI: no planned behavior change. Contract/runtime corrections require
  a recorded variance and owning-document update.
- Database/migration: no schema change; tests use existing tables and an
  explicitly disposable PostgreSQL 18 database.
- Authorization: verify existing `purchase.read` and public guest boundaries;
  add no permission or bypass.
- Audit: Operations Purchase reads and guest checkout add no privileged audit
  row. Tests must not invent one.
- Idempotency/outbox: verify existing durable encrypted replay and exactly two
  minimized Purchase outbox events; add no framework.
- Dependency/lockfile: none.

## Acceptance Criteria

1. A completed matrix maps every W3-01 through W3-08 acceptance criterion to
   passing automated or recorded evidence and names any remaining gap.
2. Public syntactic/semantic validation and response guards return stable
   errors with zero partial effects and no internal/secret leakage.
3. Same semantic intent replays the exact original `201` response/token once;
   changed input conflicts; concurrent duplicates create one business effect;
   failed work remains retryable.
4. A repeated real PostgreSQL public-route race with distinct keys proves one
   winner for the last unit and no Event/Offering quota overflow.
5. Operations missing/invalid auth and missing `purchase.read` are denied
   before repository access; permitted list/detail match the Operations
   contract and filters/cursors.
6. Storefront tests prove exact explicit-role payloads, stable manual retry,
   safe states/confirmation, and raw-token absence from every prohibited
   client surface.
7. Operations tests prove exact credentialed reads, query isolation, explicit
   roles/snapshots, permission states, and logout cleanup with no write headers
   or controls.
8. Public and Operations exposure allowlists remain separate; frontend service
   workers cache no API/auth/participant/financial/operational data.
9. Both OpenAPI contracts validate and the real browser completes the Week 3
   journey against the real API/disposable database at desktop and 360px.
10. Penpot-backed structural checks pass at the provided 862px/1440px boards
    and the 360px adaptation, while all design elements outside implemented
    contracts remain absent.
11. `make validate`, full Go vet/test/build, database-enabled tests, focused
    race tests, Compose validation, and applicable migration lifecycle checks
    pass without hidden skips.
12. Every test fixture is uniquely owned and cleaned; no secret, database dump,
    browser credential, generated output, or scanner cache is committed.
13. No arbitrary coverage gate, redundant framework, test-only bypass,
    unrequested product behavior, or unrelated refactor is added.

## Verification

Use an explicitly disposable PostgreSQL 18 database and record exact command
versions/results:

```bash
make validate
cd apps/api && go vet ./... && go test ./... && go build ./...
cd apps/api && TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/purchasing/... ./internal/app/...
cd apps/api && TEST_DATABASE_URL="$DATABASE_URL" go test -race ./internal/purchasing/... ./internal/app/...
cd apps/api && go run ./cmd/db migrate validate
cd apps/api && go run ./cmd/db migrate up
cd apps/api && go run ./cmd/db migrate down --steps 1
cd apps/api && go run ./cmd/db migrate up --steps 1
docker compose -f infrastructure/compose.yaml config
```

Run the Storefront and Operations focused test/typecheck/lint/build commands
from W3-07 and W3-08. Validate both OpenAPI 3.1 documents with the repository's
recorded pinned Redocly 2.46.1 workflow, and record unchanged baseline warnings
separately from errors. Run the real browser journey and scoped secret/
persistence inspection described above.

Do not mark database or browser criteria passed when `TEST_DATABASE_URL` or the
test OIDC identity/provider is unavailable. Record the exact blocker and
remaining risk.

## Deliverables

- Evidence-linked Week 3 coverage matrix.
- Minimal owning-layer regression additions and any test-proven narrow defect
  fixes.
- PostgreSQL contention/idempotency/rollback and Operations authorization
  evidence.
- Storefront/Operations transport, UI security, accessibility, and real-browser
  evidence.
- Contract, repository, and current-state review that closes Week 3 only when
  every required check passes.

## Risks, Decisions, and Deferred Work

- W3-09 starts only after W3-07 and W3-08 are implemented; planning detail does
  not make it execution-ready.
- Existing backend coverage is substantial. The expected work is to close
  composed/frontend gaps, especially distinct-intent public quota contention,
  auth-before-dependency proof, semantic replay, and token non-persistence.
- A configured local test OIDC identity is required for the authenticated
  browser path. Do not replace it with a production auth bypass.
- Process-local public rate limiting and local web security headers do not prove
  uniform ingress behavior; deployment/security hardening remains later work.
- Penpot has no mobile board for these screens and its Operations queue is
  skeletal. W3-09 verifies the supplied desktop composition plus the
  repository's conservative responsive fallback; it does not invent missing
  mobile or KPI behavior.
- Deferred: Payment/activation (Week 4), complete Storefront/Operations
  journeys (Weeks 5–6), and release-wide load, device, resilience, and security
  evidence (Weeks 14–16).

## Final Report Requirements

The final report must name every test added/strengthened, test-proven defect
fix, database fixture and repetition count, exact verification results,
OpenAPI warnings/errors, browser/security observations, plan variance, and
remaining gaps. It must distinguish implemented, verified, assumed, and
deferred behavior and must not claim absolute vulnerability absence or later
Purchase lifecycle completion.

## Execution Report

### Implemented

- Strengthened `TestPublicCommonPurchaseCheckoutPostgreSQL` with composed
  `400`, `413`, `415`, and `422` trust-boundary cases. Each case asserts zero
  Party, Purchase, history, participant, reservation, outbox, and replay rows.
- Replayed the successful public command with reordered and reformatted JSON
  to prove semantic canonicalization and byte-identical status/body/token,
  while a meaningful participant-name change still returns
  `idempotency_conflict`.
- Added exact public response-object allowlists and owned-sentinel scans for
  response, API logs, outbox payloads, encrypted durable replay, and stored
  SHA-256 token hashes without printing secret values on failure.
- Added three bounded, real-PostgreSQL composed-route races. Each iteration
  uses two distinct checkout intents for one final participant unit and proves
  one `201`, one stable `409 quota_unavailable`, one Purchase/history/
  participant/reservation/replay set, two outbox events, one Party, and no
  quota overflow or partial loser effects.
- Strengthened
  `TestPublicPurchaseCheckoutRetriesDuplicateReferencePostgreSQL` so the
  failed colliding attempt plus corrected retry proves only one final Party,
  history, participant, reservation, replay, and two outbox events.
- Strengthened `TestOperationsPurchaseReadsPostgreSQL` with missing,
  forbidden, expired, and revoked session cases against a deliberately nil
  Purchase repository. The stable `401`/`403` responses prove authorization
  stops before repository access. Permitted list/detail responses now assert
  exact field allowlists and token-sentinel absence.
- Added `assertExactJSONKeys` by reusing the existing integration allowlist
  helper; no new assertion, fixture, database, browser, or CI framework was
  introduced.
- Changed the Storefront Purchase API boundary to validate the contracted raw
  access token and immediately discard it, returning only `CreatedPurchase`.
  `CheckoutPage` therefore stores only safe confirmation data in TanStack
  Mutation state.
- Strengthened Storefront tests for token non-retention, edited-intent/new-key
  behavior, frozen original input, and safe idempotency/size error states.
- Strengthened the Operations Purchase client test so detail reads prove
  cookie credentials plus the absence of CSRF and idempotency headers.
- Applied the canonical repository formatter to the exact 18 pre-existing
  Week 3 Go files reported by the baseline check. Non-target changes are
  mechanical formatting only; no contract, migration, dependency, generated
  client, or business behavior was added.
- Updated `.codex/CURRENT_STATE.md` with the verified Week 3 closure and W4-01
  as the next recommended task. The live tracker was not changed because the
  user did not request tracker synchronization.

### Test-Proven Defect Fix

The new Storefront privacy assertion initially failed because the existing API
result exposed `accessToken`; TanStack `useMutation` retained that result even
though no implemented Week 3 feature consumed the credential. The owning fix
keeps response/token validation but returns only the safe Purchase view. No
server or OpenAPI behavior changed.

### PostgreSQL Fixtures and Repetition

- Automated integration tests use their existing timestamp-owned fixture
  prefixes and dependency-safe cleanup. The new public contention loop runs
  exactly `3` repetitions.
- Full verification used the explicitly disposable PostgreSQL 18 database
  `persona_apps_w309_01a056ca`.
- The browser fixture used Event
  `93090000-0000-4000-8000-000000000001`, Offering
  `93090000-0000-4000-8000-000000000002`, and a local OIDC operator. Final
  inspection found two browser-created Purchases, two histories, two ordered
  participant rows, two active reservations, four minimized outbox events,
  two encrypted replay rows, valid 32-byte token hashes, and one revoked
  Operations session. The database is removed after evidence capture.

### Verified

- Baseline focused suites passed before editing: Storefront 10 files / 31
  tests, Operations 9 files / 21 tests, and PostgreSQL-backed Purchasing/App.
- The privacy regression was observed failing with the raw token in the
  Storefront result, then passed after the narrow API-boundary fix.
- Final Storefront suite: 10 files / 32 tests. Final Operations suite: 9 files
  / 21 tests. Both typechecks, Oxlint runs, and production Next.js builds
  passed, including the dynamic checkout and Purchase-detail routes.
- `TEST_DATABASE_URL=... go test ./internal/purchasing/... ./internal/platform/idempotency/... ./internal/app/... -count=1`
  passed.
- The same focused Purchasing/App packages passed under `go test -race`.
- Full PostgreSQL-enabled `go test ./... -count=1`, `go vet ./...`, and
  `go build ./...` passed.
- `go run ./cmd/db migrate validate`, bounded migration down one step, and up
  one step passed on the disposable database.
- `make validate` passed completely: Prettier, canonical Go format, frontend
  lint/typecheck/test/build, Go vet/test/build, and Compose configuration.
- Independent `docker compose -f infrastructure/compose.yaml config` passed.
- Redocly CLI 2.46.1 validated both OpenAPI 3.1 documents with zero errors.
  The unchanged baseline remains one Storefront `info.license` warning and five
  Operations warnings: `info.license`, two redirect-only operations without a
  2XX response, and two unused response components.
- `git diff --check` and final formatter checks pass. Ignored `.next` output,
  local database data, and runtime credentials are not repository artifacts.

### Real-Browser and Security Observations

- Production Storefront and Operations Next.js servers used same-origin proxy
  requests to the real Go API, disposable PostgreSQL, and local OIDC issuer.
- Storefront evidence covered empty-form focus, explicit distinct payer and
  name-only participant roles, a preserved `quota_unavailable` state for a
  two-participant request against one unit, edited-intent success, safe inline
  confirmation, one H1, keyboard/focus behavior, and no horizontal overflow at
  862px and 360px.
- Confirmation exposed only the non-secret reference, pending status, captured
  Offering/total, participant snapshot, and reservation expiry. The URL,
  rendered text, main actions, and console contained no raw token; payment
  actions, verification promises, and tracking controls remained absent.
- Operations evidence covered signed-out, forbidden-without-`purchase.read`,
  permitted, invalid-filter, exact Event/status filter, list, direct detail,
  explicit purchaser/payer/name-only snapshots, logout, and protected direct
  return. No mutation/KPI/total control appeared and console warnings/errors
  were empty.
- The Storefront 862px and both 360px layouts had one H1 and no horizontal
  overflow. The in-app browser capped the requested Operations desktop width
  at an actual 1280px; the unchanged W3-08 Penpot-aligned desktop evidence is
  retained for the 1440px reference. Current Operations detail at 360px had
  `scrollWidth == clientWidth`.
- Browser storage was not read directly. Source inspection, the failing-then-
  passing API-result test, TanStack boundary, and deterministic service-worker
  tests prove no application persistence/query/PWA cache path for the raw
  token or Operations business data.
- API logs recorded method, bounded path, request ID, and duration only. Final
  database checks confirmed replay ciphertext and outbox JSON contain none of
  the owned contact/name sentinels.

### Plan Variance

- Implemented the recorded narrow Storefront token-retention fix after its new
  test failed.
- Implemented the recorded bounded formatter normalization so the mandatory
  `make validate` gate could pass.
- No other production defect, contract mismatch, schema change, dependency,
  framework, CI job, or unrelated refactor was found or added.

### Assumed

- The local OIDC issuer intentionally omits `purchase.read`; the disposable
  session was granted that accepted permission only after the browser first
  proved the forbidden state. Production permission provisioning remains a
  deployment responsibility.
- W3-01 through W3-08 archived evidence remains applicable where W3-09 reused
  unchanged owning tests instead of duplicating them.

### Deferred / Remaining Risk

- This evidence closes Week 3 Common Purchase creation and read boundaries; it
  does not prove absolute vulnerability absence or complete later Purchase
  lifecycle behavior.
- Payment/evidence/verification/eligibility/Sohibul activation begin in Week
  4. Purchase tracking, cancellation, token recovery, Saving, Giveaway,
  Livestock, Allocation, Slaughter, Distribution, and event-day operations
  remain deferred to their approved tasks.
- Uniform multi-replica rate limiting, deployment headers/ingress, full WCAG
  and cross-browser/device coverage, penetration testing, load/soak,
  backup/restore, staging, and production verification remain later release
  gates.
- No commit or push was requested or performed.
