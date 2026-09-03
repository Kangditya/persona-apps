# Task: W4-04 Implement Purchase Eligibility Transition

## Status

Ready for plan review on 2026-09-03. Planning-only; live tracker is `Backlog`.
The tracker dependency on W4-03 is recorded below, but BUILD requires the
sequencing decision described in this plan.

## Tracker

- Week: Week 4
- Epic: Payment & Activation
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Implement Purchase eligibility transition.
- Dependencies: W4-03, W3-05, and W1-04.
- Acceptance summary: verified Payment atomically makes the Purchase eligible
  and consumes or releases quota according to the accepted lifecycle.

## Objective

Implement the framework-neutral Purchasing transaction primitives that move a
valid Common Purchase through `PENDING_PAYMENT -> PAID -> ELIGIBLE`, append both
history entries, consume its active quota reservation, and return a stable
snapshot for W4-03. Also expose the narrow reservation-release primitive used
by rejection.

This task does not publish a partial verification route. W4-03 remains the
owning command and must combine these effects with Payment, ledger, activation,
audit, outbox, and replay.

## Source of Truth

- W4-03 plan and its canonical sources;
- `docs/domain/COMMERCE_LIFECYCLES.md`, quota and Purchase/Payment matrices;
- ADR-013, ADR-017, ADR-023, ADR-038, ADR-041, ADR-042, and ADR-044;
- migrations 0002, 0005, and 0006;
- current purchasing Purchase/repository/reservation code and W3 quota tests;
- Operations `PaymentAction` contract.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate only after the Week 4 sequencing adjustment is approved. The output is
a real transaction component with focused PostgreSQL behavior, not an empty
interface or independently exposed endpoint.

## Scope

### In Scope

- Add locked Purchase lookup and guarded eligibility mutation to the existing
  Purchasing module.
- Require `COMMON`, `PENDING_PAYMENT`, exact Payment amount/currency, matching
  Event/Purchase/Offering, and one current `RESERVED` quota attempt.
- Revalidate contested Event/Offering quota using the established lock order
  before consuming the reservation.
- Persist final Purchase `ELIGIBLE`, `eligible_at`, `updated_at`, and one command
  version increment while appending ordered `PENDING_PAYMENT -> PAID` and
  `PAID -> ELIGIBLE` history rows at the same occurrence time.
- Change reservation `RESERVED -> CONSUMED`, set `consumed_at`, clear no history,
  and enforce conditional expected status/version.
- Add rejection support that changes only `RESERVED -> RELEASED` with timestamp
  and bounded reason; terminal attempts never reopen.
- Return minimized snapshots to W4-03 for audit/outbox/response construction.
- Add domain, repository, PostgreSQL, concurrency, and rollback tests.

### Out of Scope

- Changing Payment status, writing ledger/audit/outbox/replay, activating
  participants, or registering HTTP routes.
- Purchase cancellation/expiry worker, manual eligibility override, allocation,
  refunds, Saving, Giveaway, or zero-price auto-eligibility.
- New quota counters, distributed locks, cache correctness, or broad Purchase
  repository refactoring.

## Existing State

- Purchase types already include Paid/Eligible and the database has
  `eligible_at`, version, and history fields, but Go persistence is create/read
  only.
- Reservation code creates attempt one and calculates locked usage but has no
  nullable-expiry, consume, release, or reacquire mutation.
- The database permits only one `RESERVED` attempt per Purchase and indexes
  `RESERVED` plus `CONSUMED` usage.
- No public/Operations contract directly exposes this primitive; it is an
  internal part of the W4-03 command.

## Target State and Invariants

- Eligibility can be called only inside a caller-owned transaction after the
  owning Payment command has established authoritative context.
- Two ordered Purchase history rows preserve the accepted logical states even
  though only final `ELIGIBLE` becomes visible after commit.
- One command increments the aggregate version once; history, not multiple
  transient updates, records the two logical transitions unless an accepted
  version policy says otherwise.
- Reservation consumption is conditional and cannot double-consume or exceed
  current capacity.
- Release is idempotent only through the parent command replay; a second direct
  release returns state conflict rather than silently rewriting reason/time.

## Planned File Changes

- `apps/api/internal/purchasing/purchase.go` for transition inputs/results and
  domain guards;
- `apps/api/internal/purchasing/postgres.go` for locked read/save/history;
- `apps/api/internal/purchasing/reservation.go` for consume/release and nullable
  expiry representation;
- focused unit, SQL-mock, and PostgreSQL integration tests;
- W4-03 composition tests, without an independent route or dependency.

No schema change is currently planned.

## Acceptance Criteria

1. Only a matching Common/Pending Purchase with exact Payment amount/currency
   and one reserved quota attempt can become eligible.
2. Eligibility produces final `ELIGIBLE`, one `eligible_at`, one version change,
   and exactly two correctly ordered Purchase history rows.
3. The same transaction changes quota to `CONSUMED` once with matching units and
   timestamp; capacity is revalidated under the accepted locks.
4. Rejection support releases only the current reserved attempt with one reason
   and does not change Purchase status.
5. Stale, missing, terminal, cross-Event/Offering, amount/currency, capacity,
   or participant-unit mismatch returns a stable domain conflict and no write.
6. Concurrent consume/release attempts yield one valid terminal state and no
   lost update.
7. Any history/reservation/Purchase failure rolls back every effect.
8. The primitive cannot be invoked through HTTP outside the W4-03 composed
   command and introduces no ledger/activation/audit duplication.

## Testing and Verification

- Unit tests for allowed/invalid transitions and exact amount/currency/unit
  comparisons.
- SQL tests for lock/conditional update/order/history arguments.
- PostgreSQL tests for consume, release, rollback, stale version, and repeated
  consume-vs-release races.
- W4-03 composed integration remains the final proof of atomic eligibility.

```bash
cd apps/api
go test ./internal/purchasing/... -count=1
go test -race ./internal/purchasing/...
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
```

## Risks, Decisions, and Sequencing

- Rebaseline this internal primitive before W4-03 route publication; otherwise
  tracker order would force a non-atomic verify command.
- Resolve whether one command version increment or two status increments is
  authoritative. The minimal recommendation is one version increment with two
  ordered histories because no API accepts Purchase expected-version today.
- Zero-price Purchase eligibility remains blocked by W4-01's product/schema
  decision and is not solved here.
- Global lock order must be shared with W4-01 reacquisition and W4-03 verify.

## Deliverables and Final Report

Deliver guarded Purchase eligibility and quota consume/release primitives with
focused PostgreSQL proof. Report exact lock/update order, histories, timestamps,
versions, row counts, conflict/rollback evidence, assumptions, and deferred
zero-price/cancellation/expiry work.
