# Task: W3-05 Implement Quota Reservation and Consumption Policy

## Executed

## Status

Completed and verified on 2026-08-21. No commit or push was requested.

## Tracker

- Week: Week 3
- Epic: Common Purchase
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 6 hours
- Tracker objective: Implement quota reservation and consumption policy.
- Dependencies: W2-02 and W3-03.
- Acceptance summary: concurrent checkout cannot reserve participant units
  beyond configured Event or Offering participant quota; per-Purchase Offering
  capacity remains separately enforced.

## Objective

Implement the transactional quota-reservation invariant required by Common
Purchase checkout and encode the accepted lifecycle meaning of reserved and
consumed units. Use the existing reservation table and row locks; do not add a
counter cache, queue, or distributed lock.

## Source of Truth

- `docs/PRD.md`, Phase 1 quota, reservation, and 24-hour hold rules;
- `docs/ARCHITECTURE.md`, PostgreSQL transaction and concurrency policy;
- `docs/DECISIONS.md`, especially ADR-009, ADR-011, ADR-020, ADR-038,
  ADR-041, ADR-042, and ADR-044;
- `docs/domain/COMMERCE_LIFECYCLES.md`, authoritative reservation transitions;
- `docs/database/ERD.md`, `docs/database/OPERATIONS.md`, and migrations `0005`
  and `0006`;
- W2 Event/Offering repository behavior and W3-03/W3-04 outputs.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate only after dependencies are implemented and verified. Do not commit
or push unless requested.

## Existing State

- `quota_reservations` already stores Event, Offering, Purchase, attempt number,
  participant units, status, expiry/consumed/released timestamps, and release
  reason.
- Accepted statuses are `RESERVED`, `CONSUMED`, `RELEASED`, and `EXPIRED`.
  Terminal attempts do not reopen, and a partial unique index permits at most
  one `RESERVED` attempt for a Purchase.
- Existing indexes support Event/Offering usage aggregation and expiry scans.
- Offering catalogue availability already calculates advisory values from
  `RESERVED` plus `CONSUMED`; checkout must enforce the same meaning under
  locks rather than trusting the public value.
- Event has an optional participant quota. Offering has a required
  per-Purchase participant capacity and an optional aggregate participant
  quota. Storefront availability is explicitly advisory.
- No worker or Purchase command currently creates or transitions reservations.

## Scope

### In Scope

- Lock the selected Event then Offering in one documented order inside the
  checkout transaction.
- Verify the Event is `ACTIVE`, the Offering is `PUBLISHED`, the Offering
  belongs to that Event, and current time is inside configured registration
  bounds when those bounds exist.
- Aggregate participant units in `RESERVED` and `CONSUMED` states for the Event
  and Offering while authoritative rows are locked.
- Reject a reservation that exceeds the optional Offering participant quota or
  optional Event participant quota with stable `quota_unavailable` behavior.
  The existing required Offering participant capacity remains the per-Purchase
  limit already validated by the Purchase aggregate; it is not an aggregate
  quota.
- Insert the first numbered reservation attempt as `RESERVED` with the exact
  Purchase participant count and an expiry 24 hours after the command's
  captured timestamp.
- Define and test the accepted counting/transition policy so later payment,
  cancellation, rejection, expiry, and reacquisition tasks have one invariant
  to follow.
- Add real concurrent PostgreSQL tests proving the capacity boundary.

### Out of Scope

- A background expiry worker, evidence-submission pause, cancellation/release,
  payment verification/consumption command, or quota reacquisition endpoint.
  Those require their owning later vertical slices.
- Treating `expires_at <= now` as released without the accepted `EXPIRED`
  transition. Status remains authoritative; overdue `RESERVED` rows count until
  a worker transitions them, which is conservative and race-safe.
- Cached counters, triggers, advisory locks, Redis, queues, distributed locks,
  or a generic inventory service.
- Updating the advisory public catalogue as if it were the command guard.
- Event/Offering lifecycle changes or new schema columns.

## Quota Policy

- `RESERVED` and `CONSUMED` participant units count against configured Event
  and Offering participant quotas.
- `RELEASED` and `EXPIRED` units do not count.
- Offering aggregate availability is unbounded when `participant_quota IS
  NULL`; otherwise it is `participant_quota - used_for_offering`.
- Event availability is unbounded when `participant_quota IS NULL`; otherwise
  it is `participant_quota - used_for_event`.
- The existing `participant_capacity` is a positive per-Purchase maximum; a
  request succeeds only when its count fits that snapshot and both configured
  aggregate quotas.
- `RESERVED -> CONSUMED` occurs only in the later payment-verification command
  after its accepted amount/currency/state checks.
- `RESERVED -> RELEASED` and `RESERVED -> EXPIRED` belong to later cancellation,
  rejection, and worker commands. Terminal attempts never reopen.
- Reacquisition creates the next attempt number; it never edits a terminal
  attempt back to `RESERVED`.

W3-05 implements reservation and the shared invariant. It does not scaffold
unused later handlers merely because their lifecycle is documented.

## Lock and Transaction Plan

1. Begin the caller-owned W3-06 transaction.
2. Lock the Event row `FOR UPDATE`.
3. Lock the selected Offering row `FOR UPDATE` and verify its Event ID.
4. Validate Event/Offering status and registration window.
5. Sum `RESERVED` plus `CONSUMED` units for the Event and Offering.
6. Compare the requested participant count to both configured aggregate quotas
   using checked integer arithmetic.
7. Persist Purchase/snapshots/participants and reservation attempt as one
   transaction; W3-06 also writes outbox and replay before one commit.

The Event lock intentionally serializes checkout within one Event for the MVP,
which is simple and safe at the documented event scale. If implemented, add a
short `ponytail:` comment naming this per-Event contention ceiling and the
upgrade trigger: only replace it after measured contention justifies a more
granular atomic guard. Every future quota mutation must follow the same lock
order unless an accepted design replaces it.

## Target State and Invariants

- Two concurrent requests cannot both observe and consume the same last unit.
- No successful transaction leaves participant counts above either configured
  aggregate quota.
- The reservation's Event, Offering, Purchase, and participant count match the
  Purchase aggregate exactly.
- Attempt 1 is unique for initial checkout; only one `RESERVED` attempt exists
  per Purchase.
- The 24-hour expiry is derived from one captured command timestamp and is
  returned from the stored row, not recomputed for the response.
- Any eligibility, quota, insert, outbox, or replay failure rolls the entire
  checkout transaction back.

## Implementation Plan

1. Reuse the existing Offering availability SQL meaning but place the command
   check in the owning purchasing transaction under locks; do not call the
   unlocked public catalogue query as a guard.
2. Add the minimal reservation domain values and PostgreSQL reserve operation
   to `purchasing` (or a narrowly named owned file in that module).
3. Reuse W3-03's DB transaction seam and explicit SQL style; do not add a quota
   repository interface with one implementation.
4. Map invalid state/window to `state_conflict`, invalid input to
   `validation_failed`, and exhausted quota to `quota_unavailable`.
5. Add boundary tests for unlimited/limited Event and Offering quota,
   per-Purchase capacity, counted/terminal statuses, exact last unit, wrong Event,
   inactive/unpublished/window-closed state, and rollback.
6. Add a real concurrent test with two transactions competing for the last
   unit and assert exactly one success.

## Planned File Changes

Expected minimum:

- focused quota/reservation code in `apps/api/internal/purchasing/`;
- focused unit and disposable PostgreSQL tests in the same module;
- W3-03 repository transaction code only where composition requires it.

No migration, worker, route, dependency, or infrastructure change is planned.

## API and Database Impact

- API: no new route; W3-06 exposes the reservation through successful checkout
  and `reservation_expires_at`.
- Database: reuse `quota_reservations` and existing indexes/checks.
- Audit: guest reservation is not a privileged mutation and has no standalone
  audit row; assisted reservation must add audit only when an assisted command
  actually exists.
- Outbox: W3-06 emits the accepted `PurchaseReserved` event inside the parent
  transaction.
- Idempotency: reservation participates in W3-06's checkout replay and has no
  separate public idempotency namespace.

## Acceptance Criteria

- A valid initial checkout creates exactly one attempt-1 `RESERVED` row with
  matching references/units and a stored 24-hour expiry.
- `RESERVED` and `CONSUMED` count; `RELEASED` and `EXPIRED` do not.
- Event and Offering participant quota `NULL` values are unbounded; the
  required Offering participant capacity remains enforced per Purchase.
- Event/Offering state, relationship, and registration-window guards are
  authoritative inside the transaction.
- A real contention test proves exactly one of two last-unit reservations can
  commit and final usage never exceeds capacity.
- Failures leave no Purchase, participant, reservation, outbox, or replay row.
- No worker, counter cache, distributed lock, migration, or unused transition
  handler is added.

## Verification

```bash
cd apps/api
go test ./internal/purchasing/...
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
docker compose -f infrastructure/compose.yaml config
```

Run migrations and concurrency tests against disposable PostgreSQL; an in-memory
or mocked test alone cannot verify row-lock behavior. Record test repetition or
stress count used for the contention case.

## Decisions, Assumptions, and Deferred Work

- Accepted: a checkout reservation lasts 24 hours.
- Accepted: locked authoritative totals or an equivalent atomic guard protect
  Event and Offering quota; this plan chooses the existing-row lock solution.
- Conservative interpretation: overdue `RESERVED` rows count until explicitly
  transitioned to `EXPIRED`.
- Deferred: consume/release/expire/reacquire commands and the expiry worker.

## Final Review Requirements

Report lock order, exact counting query/semantics, transaction composition,
contention evidence, error mapping, and later transitions not implemented.
Do not call capacity protected without a real concurrent PostgreSQL check.

## Final Review

### Implemented

- `LockCheckout` locks Event then Offering in the caller-owned PostgreSQL
  transaction, requires an active Event, published Offering, matching Event
  relationship, and an open registration window.
- `Reserve` creates attempt-one `RESERVED` records with the Purchase's exact
  participant count and a stored 24-hour expiry. It counts only `RESERVED` and
  `CONSUMED` rows under the retained Event lock.
- Aggregate guards use the documented optional Event and Offering
  `participant_quota` values. Required Offering `participant_capacity` remains
  the existing per-Purchase bound; it is intentionally not misused as a global
  quota.

### Verified

- Unit coverage validates source status, Event/Offering relationship,
  registration-window, and participant-count guards.
- Disposable PostgreSQL coverage verifies stored expiry, counted versus
  terminal reservation states, rollback after a post-reservation failure,
  bounded and unbounded quota semantics, and a real two-transaction last-unit
  race. The race passed three consecutive repetitions with exactly one commit.
- `go vet ./...`, database-backed `go test ./...`, `go build ./...`,
  `make validate`, `docker compose -f infrastructure/compose.yaml config`, and
  `git diff --check` passed.

### Assumed

- Registration windows are interpreted as `[opens_at, closes_at)`: opening is
  inclusive and the exact closing instant is no longer eligible. This is an
  explicit implementation boundary because canonical artifacts require a
  window but do not state endpoint inclusivity.

### Deferred

- W3-04 remains blocked on the product total formula; W3-06 will compose this
  locked reservation with snapshots, Party creation, reference/token,
  idempotency, outbox, and public checkout.
- Consume, release, expiry, reacquisition, evidence pause, workers, routes,
  counters, and distributed locking were not added.
