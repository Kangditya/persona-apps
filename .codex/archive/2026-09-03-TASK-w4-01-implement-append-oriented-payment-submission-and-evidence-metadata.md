# Task: W4-01 Implement Append-Oriented Payment Submission and Evidence Metadata

## Executed

## Status

Completed and verified on 2026-09-03 after explicit execution approval. The
live tracker remains `Backlog`; no tracker write, commit, or push was requested.

## Tracker

- Week: Week 4
- Epic: Payment & Activation
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 6 hours
- Tracker objective: Implement append-oriented Payment submission and evidence
  metadata.
- Dependencies: W3-03 and W3-06.
- Acceptance summary: a Purchase can receive an append-only Payment submission
  with bounded evidence metadata.

## Execution Reconciliation — 2026-09-03

Current source, contracts, migrations, canonical commerce/security documents,
and the W3 execution baseline were re-read before BUILD. W4-01 remains bounded
to a Payment submission module behind a storage port; W4-02 still owns the
concrete durable evidence adapter and production route wiring.

The following conservative W4-01 decisions are applied so domain and
persistence work can proceed without inventing W4-02 infrastructure:

- manual-transfer submissions persist method `MANUAL_TRANSFER`, directly from
  the accepted Phase 1 manual-transfer evidence rule;
- Payment references use `PAY-` plus 128 random bits encoded as uppercase,
  unpadded Base32, avoiding sequential volume disclosure;
- idempotency replay is durable for a submission intent; a genuinely new
  post-rejection attempt requires a new key;
- at most one `SUBMITTED` Payment may exist for one Purchase, enforced by an
  additive partial unique index so different keys cannot create parallel review
  attempts;
- a zero-total Purchase cannot submit a positive Payment record under the
  current schema and returns a state conflict; zero-price auto-eligibility is
  explicitly deferred rather than invented;
- the W4-01 public handler is exercised with a storage fake but is not registered
  in the production router until W4-02 supplies an approved concrete adapter.

### PLAN VARIANCE — additive one-submitted-Payment constraint

- Planned work: add a migration only if the duplicate-submission invariant is
  approved by reconciliation.
- Unexpected requirement: command idempotency cannot prevent two different
  keys from creating concurrent `SUBMITTED` records for one Purchase.
- Necessity: the Operations review flow and append-oriented resubmission rule
  require one current attempt while preserving rejected attempts.
- Impact: add migration 0007 with a partial unique index on `purchase_id` where
  status is `SUBMITTED`, plus matching ERD/migration documentation and tests.
- Scope: the constraint is Payment-submission-owned and does not change rejected
  history, provider callbacks, Saving, or Giveaway.

## Objective

Implement the framework-neutral Payment domain and PostgreSQL persistence used
by Purchase-token-scoped manual-transfer evidence submission. A successful
submission creates one immutable evidence-bearing `SUBMITTED` Payment and its
initial status history, pauses or reacquires Purchase quota as required, writes
a minimized outbox event, and participates in the existing command-scoped
idempotency transaction.

This task owns Payment submission behavior and metadata. W4-02 owns the concrete
private evidence store and production route wiring. Evidence bytes never enter
PostgreSQL.

## Context

The Storefront OpenAPI already drafts
`POST /api/public/v1/purchases/{purchase_id}/payment-evidence` as a multipart,
Purchase-Bearer, idempotent command. Migrations 0002 and 0005 already provide
`payment_records`, `payment_status_history`, evidence metadata, quota attempts,
Purchase token hashes, outbox, and the generic replay ledger. There is no
runtime `payment` module, Purchase-token validator, Payment repository, evidence
submission handler, reservation pause/reacquisition command, or storage adapter.

The accepted Phase 1 flow is manual transfer evidence. Each rejected attempt
remains immutable and a later submission creates a new Payment. Submission must
declare the exact Purchase total/currency and either pause the current reserved
quota expiry or atomically reacquire quota after release.

## Source of Truth

- `docs/PRD.md`, especially Phase 1 Common-Purchase Rules, Payment and Ledger,
  security, privacy, and data integrity;
- `docs/PRODUCT_MAP.md`, Payment & Funding and the first vertical slice;
- `docs/ARCHITECTURE.md`, Payment ownership, transactions, API boundaries,
  concurrency, external adapters, and verification strategy;
- `docs/DECISIONS.md`, especially ADR-013, ADR-014, ADR-017 through ADR-020,
  ADR-024, ADR-026, ADR-038, and ADR-041 through ADR-050;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- `docs/security/AUTHENTICATION.md` and `docs/security/PERMISSIONS.md`;
- `docs/database/ERD.md`, `docs/database/MIGRATION_PLAN.md`, and migrations
  0002, 0003, 0005, and 0006;
- `contracts/openapi/storefront.yaml`;
- current purchasing, identity, audit, outbox, idempotency, HTTP, CORS, rate
  limit, route-composition, and PostgreSQL integration source;
- completed W3-01 through W3-09 execution evidence;
- `.codex/CURRENT_STATE.md`, `MVP-DELIVERY-ROADMAP.md`, and live tracker rows
  26 through 34.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

At activation, re-read the resolved ADR and W4-02 plan, then activate this file
as the sole `.codex/TASK.md`. Do not register a public upload route against a
fake, in-memory, or unavailable evidence store.

## Scope

### In Scope

- Add a real `apps/api/internal/payment` vertical slice with Payment types,
  validation, persistence, submission orchestration, public DTO mapping, and
  focused tests.
- Authenticate the opaque Purchase Bearer token against the identified Purchase
  using SHA-256 plus constant-time comparison before idempotency replay.
- Bound multipart parsing and stream one JPEG, PNG, or PDF evidence file through
  a narrow evidence-store port without buffering more than the accepted limit.
- Derive trusted media type from file content, calculate SHA-256 while streaming,
  retain a normalized bounded original filename, and persist only the opaque
  reference plus metadata.
- Require exact Purchase amount and currency; reject underpayment, overpayment,
  malformed money, unsupported content, empty evidence, and evidence over
  10 MiB.
- Create a new Payment and initial `NULL -> SUBMITTED` history row; never edit a
  rejected Payment into a new attempt.
- Pause a current `RESERVED` quota attempt by setting `expires_at` to null while
  review is pending.
- If the prior attempt is `RELEASED` or `EXPIRED`, lock Event then Offering,
  recheck capacity, and create the next numbered `RESERVED` attempt before the
  Payment commits.
- Write one minimized `PaymentEvidenceSubmitted` outbox event with no filename,
  digest, object reference, Purchase token, idempotency key, contact, or bytes.
- Use semantic request hashing over Purchase ID, declared amount/currency, and
  trusted evidence digest/metadata, never raw multipart boundaries.
- Map errors through the existing safe envelope and retain public no-store,
  exact-origin, rate-limit, and request-ID behavior.
- Add an additive migration only if the approved duplicate-submission invariant
  requires a partial unique guard for one in-review Payment per Purchase.

### Out of Scope

- Payment verification/rejection, Purchase eligibility, ledger credit,
  Sohibul Qurban activation, or privileged audit; these belong to W4-03 through
  W4-05 and W4-08.
- A concrete filesystem, S3, Supabase Storage, or other provider implementation;
  W4-02 owns the selected adapter.
- Public Purchase tracking/resumption, cancellation, token recovery/rotation,
  payment instructions, refunds, partial payments, balance calculation,
  provider callbacks, Saving, or Giveaway.
- Public object URLs, presigned URLs, evidence bytes in PostgreSQL, mutable
  evidence replacement, or exposing internal references/digests.
- A generic repository framework, command bus, upload framework, or new
  dependency when the Go standard library and existing platform code suffice.

## Existing State

- `payment_records` requires positive amount, currency, method, status, and one
  target; evidence metadata is all-or-none and limited to JPEG/PNG/PDF and
  10 MiB.
- `payment_status_history` and `financial_ledger_entries` exist but have no Go
  owner. Verification, not submission, creates the ledger credit.
- `purchases.access_token_hash` stores the only Purchase credential material;
  no runtime validator currently exists.
- `quota_reservations` permits one `RESERVED` attempt and supports nullable
  expiry, terminal attempts, and numbered reacquisition, but current Go code
  implements checkout attempt one only and models expiry as non-null.
- The shared idempotency executor provides encrypted replay, transactional key
  claiming, semantic request hashes, and command-specific retention.
- The Storefront contract contains the upload shape and response but runtime
  registration currently exposes only Purchase creation.

## Target State

- A Purchase-token-authenticated submission reaches Payment application logic
  only after strict multipart, token, and evidence validation.
- Exactly one accepted intent produces one new `SUBMITTED` Payment, one initial
  history row, a safely stored evidence reference/metadata set, the applicable
  quota pause/reacquisition, one outbox row, and one replay response.
- Replaying the same namespace/key/semantic evidence returns the original 201
  response and creates no file, Payment, history, reservation, or outbox
  duplicate.
- A new intent after rejection creates a new Payment and preserves the rejected
  attempt unchanged.

## Implementation Requirements

### Domain and Persistence

- Keep Payment statuses limited to the schema values; W4-01 constructs only
  `SUBMITTED`.
- Validate UUIDs, exact integer bounds, uppercase currency, trusted evidence
  metadata, target Purchase/Payer/Event relationships, and UTC timestamps in
  the owning layer.
- Query and lock the identified Purchase and relevant reservation in the caller
  transaction. Never trust a client-supplied payer, Event, method, or storage
  reference.
- Insert Payment plus initial history through the same caller-owned transaction.
- Translate constraint conflicts to stable domain errors without returning SQL.

### Trust Boundary and Idempotency

- Require exactly one `Authorization: Bearer` credential, valid multipart
  content, one evidence part, amount, currency, and `Idempotency-Key`.
- Reject malformed/unknown duplicate fields, multiple evidence parts, empty
  files, oversized filenames, unsupported magic bytes, and trailing form data.
- Authenticate the token before replay lookup. The namespace must include the
  Storefront command and stable Purchase scope, not IP, request ID, contact, or
  token value.
- Freeze one client key per submission intent. A same-key digest/metadata replay
  succeeds; any semantic mismatch returns `idempotency_conflict`.
- Store evidence only after the command claim is won, following the W4-02
  consistency protocol, so concurrent retries do not create duplicate objects.

### Quota and Atomicity

- Existing `RESERVED` quota is paused, not replaced.
- Terminal quota never reopens. Reacquisition creates `attempt_no + 1` after
  Event-then-Offering locks and counts only `RESERVED` plus `CONSUMED` units.
- Payment, history, reservation change/new attempt, outbox, and replay response
  commit together. Storage cleanup behavior follows W4-02 because external
  bytes cannot share the PostgreSQL transaction.

## Planned File Changes

Expected minimum, subject to the resolved gates:

- `apps/api/internal/payment/payment.go` and focused tests;
- `apps/api/internal/payment/postgres.go` and PostgreSQL tests;
- `apps/api/internal/payment/public_http.go` and focused boundary tests;
- `apps/api/internal/purchasing/postgres.go` and/or a focused token/reservation
  file for Purchase-token validation, pause, and reacquisition;
- `contracts/openapi/storefront.yaml` for any approved filename, current-payment,
  replay-retention, or error clarification;
- a new additive migration pair only for an approved missing invariant;
- canonical ADR/current-state files only when the decision is approved;
- app bootstrap/route files only in W4-02 when a real store is available.

No new dependency or generated client is planned.

## Acceptance Criteria

1. A valid Purchase token and evidence intent creates one `SUBMITTED` Payment
   with exact Purchase amount/currency and complete trusted metadata.
2. Evidence bytes are absent from PostgreSQL, logs, errors, audit, outbox, and
   API responses; storage reference and SHA-256 remain internal.
3. JPEG, PNG, and PDF content up to 10 MiB succeeds; empty, mismatched,
   unsupported, or oversized content fails before business state commits.
4. Same-key semantic replay returns the original response and creates no new
   storage or database effect; changed intent conflicts.
5. An active reservation expiry is paused. A terminal reservation remains
   terminal and a new capacity-safe numbered attempt is created.
6. Reacquisition failure returns `quota_unavailable` and commits no Payment,
   history, outbox, replay, or durable evidence reference.
7. A rejected Payment remains unchanged and a new approved intent appends a new
   Payment rather than overwriting it.
8. Missing/wrong Purchase tokens are rejected before replay or storage access.
9. The public contract, handler, persistence constraints, and tests agree on
   amount, currency, filename, media type, size, status, and safe exposure.
10. No verification, eligibility, activation, provider callback, payment
    instruction, or tracking behavior is claimed.

## Testing

- Domain/table tests for money, currency, metadata all-or-none, status, Party,
  and append-oriented construction.
- Repository tests for Payment/history atomicity, current-submission conflict,
  reservation pause, numbered reacquisition, rollback, and safe error mapping.
- Handler tests for token ordering, multipart limits, magic-byte detection,
  duplicate fields, exact response exposure, no-store, rate limit, and errors.
- PostgreSQL tests for same-key replay, different-key contention, quota
  reacquisition races, and zero partial effects.
- Storage calls use a tiny fake port in W4-01; concrete adapter behavior belongs
  to W4-02.

## Verification

```bash
cd apps/api
go test ./internal/payment/... ./internal/purchasing/... ./internal/platform/idempotency/... ./internal/app/... -count=1
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
docker compose -f infrastructure/compose.yaml config
```

Also validate the Storefront OpenAPI, run the focused PostgreSQL suite against
a disposable migrated database, repeat duplicate/reacquisition races, and
inspect database/log/test captures for evidence, token, key, digest, path, and
contact leakage.

## Decision Gates and Risks

- **Storage gate:** remains assigned to W4-02. W4-01 supplies the port and tested
  handler but does not register a production route without a real private store.
- **Zero-price gate:** ADR-056 makes W4-01 return `state_conflict`; zero-price
  auto-eligibility remains a separate unresolved product decision.
- **Duplicate gate:** resolved by migration 0007's partial unique index for one
  `SUBMITTED` Payment per Purchase.
- **Reference/method gate:** ADR-056 fixes `PAY-<128-bit Base32>` and
  `MANUAL_TRANSFER`.
- **Replay-retention gate:** ADR-056 makes same-intent submission replay durable;
  a new post-rejection intent uses a new key.
- Storage commit uncertainty, crash-orphan cleanup, private retrieval, and
  evidence retention remain explicit W4-02 work.

## Deliverables

- Approved Payment submission domain/persistence plan and decision record(s).
- Minimal Payment module, Purchase-token boundary, quota pause/reacquisition,
  public command adapter, contract alignment, and focused evidence.
- Current-state update and canonical task archive only after later explicit
  BUILD approval and successful verification.

## Final Report Requirements

Report exact files, resolved decisions, evidence limits, token/auth ordering,
idempotency namespace/hash/retention, quota attempt behavior, storage effects,
database rows, focused/full validation, plan variances, and deferred work.
Distinguish implemented, verified, assumed, and deferred behavior.

## Final Review — 2026-09-03

### Implemented

- Added the Payment domain, storage port, PostgreSQL submission repository,
  strict multipart public handler, safe response/error mapping, and focused
  tests under `apps/api/internal/payment`.
- Added Purchase Bearer validation using canonical base64url shape, SHA-256,
  constant-time comparison, and authentication before replay lookup.
- Added Event-then-Offering-locked reservation preparation: pause an unexpired
  hold, reject an already paused/consumed attempt, expire and reacquire an
  overdue attempt, or reacquire after `RELEASED`/`EXPIRED` with quota checks.
- Added durable semantic command replay scoped to the Purchase. The request hash
  uses normalized amount, currency, filename, trusted media type, byte size,
  and evidence digest, not multipart boundaries or raw bytes.
- Added `MANUAL_TRANSFER`, random 128-bit `PAY-` references, append-only Payment
  plus initial history, and one minimized `PaymentEvidenceSubmitted` outbox row.
- Added migration 0007's partial unique one-`SUBMITTED`-Payment-per-Purchase
  guard, ADR-056, Storefront contract constraints, database documentation, and
  current-state reconciliation.

### Verified

- `TEST_DATABASE_URL=... go test ./internal/payment/... ./internal/purchasing/... -count=1`:
  54 focused tests passed against disposable PostgreSQL 18 during the focused
  run; the final repository run includes the same coverage.
- `TEST_DATABASE_URL=... go test -race ./internal/payment/... -count=1`:
  26 tests passed with the race detector.
- `TEST_DATABASE_URL=... go test ./... -count=1`: 279 tests passed in
  23 packages.
- `go vet ./...` and `go build ./...`: passed.
- Database CLI: seven migrations validated; version 7 clean; down one reached
  version 6 clean; reapply returned to version 7 clean. PostgreSQL readback
  confirmed the exact partial unique index predicate.
- Redocly CLI 2.46.1 validated Storefront OpenAPI with zero errors and the one
  pre-existing `info.license` warning.
- `make validate`, Compose config, explicit Go/Markdown/OpenAPI formatting, and
  `git diff --check`: passed.
- PostgreSQL tests prove valid metadata, exact 10 MiB acceptance, JPEG/PNG/PDF
  content matching, same-intent replay across different multipart boundaries,
  changed-intent conflict, same/different-key concurrency, one current
  submission, rollback cleanup, wrong-token denial before storage, released
  attempt reacquisition, expired-attempt quota rollback, amount/currency
  rejection, and zero-total state conflict.

### Assumptions

- W4-01 uses a synchronous storage port whose `Put` completion means bytes are
  available for the database reference. The concrete durability guarantee is
  not assumed; W4-02 must define it.
- A bounded 10 MiB evidence payload is held in memory after streaming multipart
  validation so its semantic digest is known before the idempotency claim. The
  buffer and digest are cleared when the request completes.

### Deferred

- Concrete private evidence storage, runtime configuration, route/CORS wiring,
  private retrieval, commit-uncertainty/orphan cleanup, and retention are W4-02.
- Payment verification/rejection, ledger credit, Purchase eligibility, Sohibul
  Qurban activation, privileged audit, Storefront/Operations UI, zero-price
  eligibility, refunds, partial payments, provider callbacks, and token recovery
  remain later tasks.
- The active public server still returns no Payment evidence route because
  registering it without an approved durable store would be unsafe.

### Plan Variance

- Migration 0007 and ADR-056 were required to resolve the plan's explicit
  duplicate/reference/method/replay gates. No dependency, generated client,
  payment provider, concrete storage adapter, or unrelated module was added.
