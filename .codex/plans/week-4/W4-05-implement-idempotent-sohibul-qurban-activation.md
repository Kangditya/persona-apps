# Task: W4-05 Implement Idempotent Sohibul Qurban Activation

## Status

Ready for plan review on 2026-09-03, blocked for BUILD by the unresolved
name-only participant and participant-reference decisions. This is planning
only; tracker remains `Backlog`.

## Tracker

- Week: Week 4
- Epic: Payment & Activation
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Implement idempotent Sohibul Qurban activation.
- Dependencies: W4-04 and W3-02.
- Acceptance summary: eligible Purchase activation creates each Sohibul Qurban
  outcome once, even on retry.

## Objective

Implement a real participant vertical slice that creates one Party-backed
`ACTIVE` Sohibul Qurban outcome and initial status history for every resolved
intended Purchase participant. The function is transaction-owned by W4-03 and
uses natural uniqueness plus parent command replay to make activation exact.

## Context

The schema already has `sohibul_qurban`, unique `(purchase_id, sequence_no)`,
unique `participant_ref`, and status history. Current checkout permits either a
Party-linked participant or a name-only participant with null `party_id`.
Canonical lifecycle policy requires a Party for every participant before
activation, while ADR-048 says name-only checkout does not create a Party.
There is no participant-resolution command or accepted automatic identity rule.

## Source of Truth

- `docs/PRD.md`, role separation and Phase 1 activation rules;
- `docs/PRODUCT_MAP.md`, Party & Participant ownership;
- `docs/ARCHITECTURE.md`, Participant domain and eligibility boundary;
- ADR-013, ADR-014, ADR-019, ADR-020, ADR-024, ADR-038, ADR-041, ADR-042,
  ADR-044, ADR-048, and ADR-053;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- migrations 0002/0005, ERD participant notes, W3-02/W3-03 source, and W4-03/
  W4-04 plans;
- Operations/Public participant contract shapes where they affect exposure.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Before BUILD, resolve how a name-only intended participant obtains an approved
Party link. Do not infer identity from matching names/contact, silently create a
durable person from an unverified ceremonial name, or skip that participant.

## Scope

### In Scope

- Add `apps/api/internal/participant` only with concrete activation behavior,
  persistence, DTOs needed by W4-03, and focused tests.
- Require an eligible Common Purchase snapshot and ordered intended participant
  positions with non-null Party IDs.
- Generate the approved unique participant reference and create `ACTIVE`
  outcomes directly; the baseline `PENDING` status is not used.
- Copy Event, Purchase, Party, sequence, and display-name snapshot from locked
  authoritative records; set `verified_at` and timestamps from the parent
  verification occurrence time.
- Insert one `NULL -> ACTIVE` history row per new outcome.
- Treat an existing `(purchase_id, sequence_no)` with the same immutable source
  values as the natural duplicate result; mismatched existing data is a state
  conflict.
- Write minimized activation outbox events as part of the W4-03 transaction;
  use the parent audit record rather than per-row privileged audit duplication.
- Return activated/existing counts for the contracted action response.
- Add unit, repository, PostgreSQL, concurrency, rollback, and exposure tests.

### Out of Scope

- Participant resolution/verification UI or command unless explicitly approved
  to close the name-only gate.
- Automatic Party matching, deduplication, merge, contact inference, account
  creation, replacement, cancellation, attendance, allocation, or documents.
- A generic participant framework or independent activation HTTP route.

## Existing State

- `purchase_participants` preserves ordered display-name snapshots and optional
  Party IDs.
- W3-02 can create Party-linked purchaser/payer participants and name-only
  participants. Equal contact/name never implies identity reuse.
- `sohibul_qurban` requires non-null Party, unique Purchase position, and a
  unique text reference; no exact reference format is accepted.
- The Operations Payment action response exposes only activation count/status,
  not participant identities or references.

## Target State and Invariants

- Activation is called only from the authorized W4-03 verification transaction
  after W4-04 establishes eligibility.
- The number of active outcomes equals the Purchase participant count; partial
  activation is forbidden.
- Sequence/order and historical display names remain unchanged.
- Retry cannot append another outcome/history/outbox event. A conflicting
  natural duplicate aborts the entire verification.
- Attendance remains independent and uninitialized until its later slice.

## Planned File Changes

- `apps/api/internal/participant/activation.go` and tests;
- `apps/api/internal/participant/postgres.go` and PostgreSQL tests;
- W4-03 Payment composition and outbox response mapping;
- a canonical ADR/contract/source update for participant resolution/reference
  only after approval;
- no migration unless the approved resolution policy needs new persistent data.

## Acceptance Criteria

1. An eligible Purchase with N resolved participant positions produces exactly
   N active Sohibul Qurban rows and N initial history rows in one transaction.
2. Each row preserves Event/Purchase/Party/sequence/display snapshot and uses
   the same verification occurrence time.
3. Parent replay and unique `(purchase_id, sequence_no)` prevent duplicate
   outcomes; a direct identical retry returns the existing set without new
   history, while mismatched data conflicts.
4. Any unresolved Party, missing/duplicate sequence, wrong Event/Purchase,
   non-eligible Purchase, or reference collision aborts all activation and the
   parent verification.
5. Activation creates `ACTIVE` directly and does not use `PENDING`, attendance,
   allocation, replacement, or cancellation behavior.
6. Outbox/API/audit exposure is minimized; no contact, token, evidence, or
   unneeded participant detail is copied.
7. Concurrent activation attempts yield one complete set and no partial or
   duplicate history.

## Testing and Verification

- Unit tests for source validation, sequence completeness, direct ACTIVE state,
  snapshot preservation, and conflict detection.
- PostgreSQL tests for N positions, identical duplicate, mismatched duplicate,
  reference collision, null Party, rollback, and concurrent activation.
- W4-03 composed tests prove Payment/Purchase/quota/ledger/audit/outbox/replay
  roll back when any activation row fails.

```bash
cd apps/api
go test ./internal/participant/... ./internal/purchasing/... ./internal/payment/... -count=1
go test -race ./internal/participant/...
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
```

## Decision Gates and Risks

- **Name-only participant:** choose an explicit resolution workflow/permission
  before verification, disallow name-only checkout for the Payment slice, or
  define another accepted policy. Current artifacts cannot activate such rows.
- **Participant reference:** define whether the unique reference is a random
  support identifier, deterministic Purchase-position reference, or UUID-like
  internal value before it becomes persisted behavior.
- An identical direct activation retry is defensive behavior; HTTP idempotency
  remains owned by the parent Payment verify command.
- Do not update tracker dependency/status until the W4-03/04/05 execution
  rebaseline is explicitly approved.

## Deliverables and Final Report

Deliver the resolved participant policy, concrete activation domain/repository,
natural duplicate handling, histories/outbox composition, and focused database
evidence. Report created/existing counts, reference policy, rollback/concurrency,
unresolved-participant behavior, and deferred participant lifecycle work.
