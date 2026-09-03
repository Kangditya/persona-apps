# Task: W4-03 Implement Verification and Rejection Commands

## Status

Ready for plan review on 2026-09-03. Planning-only; tracker status remains
`Backlog`. BUILD is blocked until W4-01/W4-02 decisions are resolved and the
Week 4 execution sequence is explicitly rebaselined.

Canonical atomicity prevents a partially implemented verify route. The
recommended execution order prepares W4-04 eligibility/quota and W4-05
activation primitives before this task publishes the composed command.

## Tracker

- Week: Week 4
- Epic: Payment & Activation
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 6 hours
- Tracker objective: Implement verification and rejection commands.
- Tracker dependencies: W4-01, W4-02, and W1-04.
- Reconciled execution dependencies: W4-01, W4-02, W4-04, and W4-05, plus the
  audit requirements planned for later verification in W4-08.
- Acceptance summary: an authorized operator can verify or reject submitted
  Payments with reason, audit, and replay safety.

## Objective

Compose the permission-gated Operations Payment decision commands over the
Payment, Purchase, quota, ledger, Sohibul Qurban, audit, outbox, and idempotency
components. Verification must commit the entire accepted lifecycle once;
rejection must record the reason and release quota once while leaving the
Purchase pending.

## Context

The Operations contract already drafts `POST .../payments/{id}/verify` and
`POST .../payments/{id}/reject` with `payment.verify`, CSRF, idempotency, and a
combined action response. Current Operations routing has no Payment handler.
The existing auth middleware enforces session, Origin, CSRF, then permission;
the replay executor, audit writer, outbox writer, and transaction helper are
implemented and must be reused.

The canonical verify command is indivisible:

```text
Payment SUBMITTED -> VERIFIED
Purchase PENDING_PAYMENT -> PAID -> ELIGIBLE
quota RESERVED -> CONSUMED
one ledger credit
one ACTIVE Sohibul Qurban per resolved intended participant
one minimized privileged audit record
outbox events and replay response
```

## Source of Truth

- all canonical and current sources listed by W4-01/W4-02;
- `docs/domain/COMMERCE_LIFECYCLES.md`, Purchase/Payment and Sohibul tables;
- `docs/security/PERMISSIONS.md`, `payment.verify`, route mapping, and audit
  ordering;
- `contracts/openapi/operations.yaml`, Payment action paths and schemas;
- migrations 0002, 0003, 0005, and 0006;
- existing Event/Offering Operations command handlers as the nearest runtime
  pattern, without copying their simpler transaction semantics blindly;
- W4-04, W4-05, and W4-08 approved outputs.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Do not register the verify route until every canonical effect is present in the
same transaction. A route that only marks Payment verified is a data-integrity
bug, not an incremental delivery.

## Scope

### In Scope

- Add Operations Payment verify/reject handlers under the real Payment module.
- Require the existing Operations session, exact Origin, CSRF token,
  `payment.verify`, and `Idempotency-Key` before replay lookup.
- Strictly validate Payment UUID; verify has no semantic body, while reject
  requires a trimmed bounded non-empty reason.
- Lock and re-read authoritative rows in the approved global order; validate
  Payment/Purchase/status/amount/currency/reservation/participant relationships.
- Verification calls W4-04 and W4-05 primitives, writes one unique ledger
  credit, audit, outbox, response, and replay in the same transaction.
- Rejection sets Payment `REJECTED`, appends Payment history, releases the
  current reservation with reason, keeps Purchase `PENDING_PAYMENT`, writes
  audit/outbox/replay, and creates no ledger/activation effect.
- Use a stable effect-level ledger idempotency key derived from the command
  identity without copying the raw HTTP key into domain/audit/outbox data.
- Return the contracted minimized Payment, Purchase, quota, and activation
  summary; do not expose evidence references/digests, tokens, SQL, or audit
  payloads.
- Add unit, handler, PostgreSQL, composed-route, authorization-ordering,
  idempotency, concurrency, rollback, and exposure tests.

### Out of Scope

- Partial/overpayment, refunds, transfers, adjustments, provider callbacks,
  Saving/ Giveaway eligibility, manual Purchase overrides, or evidence edits.
- Participant resolution/replacement, zero-price policy, storage selection, or
  audit retention; these must be resolved before this command.
- A workflow engine, generic command bus, distributed transaction, queue, or
  new authorization/idempotency framework.
- Operations UI; W4-06 consumes these routes.

## Existing State

- Payment, history, ledger, Purchase history, quota, participant, audit, outbox,
  and replay tables exist.
- No Payment Go repository/handler or eligibility/activation mutation exists.
- Existing auth `Require` already enforces the required security ordering.
- Operations contract currently omits a verify request body and requires a
  rejection `ReasonRequest`.
- `financial_ledger_entries.idempotency_key` is the final append-only effect
  guard; histories/audit/outbox must not receive independent keys.

## Target State and Command Order

1. Middleware authenticates session, validates Origin/CSRF, and authorizes
   `payment.verify`.
2. Handler validates UUID/body/idempotency key and hashes normalized intent.
3. Replay executor claims `(namespace, key)` for operator scope.
4. Work transaction discovers context, takes approved locks, and revalidates
   every state and relationship.
5. Verify or reject effects, history, ledger where applicable, audit, outbox,
   and response are written.
6. Encrypted replay response and all authoritative state commit once.

Same key plus same intent replays. Same key plus another Payment/action/reason
conflicts. A new key against a terminal Payment returns `state_conflict` and
does not append another effect.

## Verification Requirements

### Verify

- Require `SUBMITTED` Payment linked to one `COMMON` Purchase still
  `PENDING_PAYMENT` with an active reservation and exact amount/currency.
- Revalidate configured quota under the accepted locks before consumption.
- Append `SUBMITTED -> VERIFIED` Payment history and the two ordered Purchase
  histories.
- Insert exactly one `PAYMENT`/`CREDIT` ledger entry tied to Payment/Purchase.
- Consume quota and activate resolved participants through W4-04/W4-05.
- Write one parent payment-decision audit entry and minimized lifecycle outbox
  events at one UTC occurrence time.

### Reject

- Require `SUBMITTED`; normalize and retain the reason in Payment/history,
  quota release, and audit according to their fields.
- Append `SUBMITTED -> REJECTED`, release only `RESERVED`, keep Purchase
  pending, and report activated count zero.
- Do not delete evidence, create ledger credit, change participant outcomes, or
  reopen terminal quota.

## Planned File Changes

- W4-01 `apps/api/internal/payment` domain/repository files and tests;
- `apps/api/internal/payment/operations_http.go` and focused tests;
- W4-04 purchasing transition/quota files and W4-05 participant activation;
- existing audit/outbox/idempotency utilities only for proven missing behavior;
- `apps/api/internal/app/server.go`, `routes_operations.go`, bootstrap wiring,
  and composed integration tests;
- `contracts/openapi/operations.yaml` for any verified response/error/reason or
  evidence-access clarification;
- no migration unless a focused PostgreSQL test proves an approved invariant
  is absent.

## Acceptance Criteria

1. Only an authenticated, CSRF-valid operator with `payment.verify` reaches the
   decision or replay lookup.
2. Verify commits all canonical Payment, Purchase, quota, ledger, activation,
   audit, outbox, and replay effects together or none.
3. Reject requires a reason, preserves evidence, releases reserved quota,
   leaves Purchase pending, and creates no ledger/activation.
4. Exact same-intent retry returns the original response with zero additional
   rows; changed intent conflicts; terminal state with a new key conflicts.
5. Ledger effect uniqueness remains the final duplicate guard even if replay
   handling is bypassed by a defect.
6. Any injected persistence/audit/outbox/replay failure rolls back all database
   effects and leaves the command safely retryable.
7. Amount, currency, status, quota, Event/Offering relationship, and participant
   Party requirements are revalidated from locked records.
8. Responses/logs/audit/outbox exclude evidence reference/digest/bytes,
   credentials, raw idempotency key, contacts, and internal errors.
9. Runtime routes, permissions, CSRF, idempotency, contract, and action response
   agree exactly.

## Testing

- Domain/repository tests for allowed states, amount/currency, ordered history,
  ledger derivation, reject reason, and error translation.
- Handler tests for security ordering, strict input, request hash, stable errors,
  safe response fields, and no-store.
- PostgreSQL integration for same/different-key retries, new-key terminal
  conflicts, all rollback injection points, and ledger/activation uniqueness.
- Repeated concurrent verify/verify and verify/reject races against one Payment;
  exactly one valid terminal outcome may commit.
- Authorization tests prove missing/expired/revoked/forbidden/CSRF-invalid
  callers cannot observe a replay or touch repositories.

## Verification

```bash
cd apps/api
go test ./internal/payment/... ./internal/purchasing/... ./internal/participant/... ./internal/platform/... ./internal/app/... -count=1
go test -race ./internal/payment/... ./internal/purchasing/... ./internal/participant/...
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
docker compose -f infrastructure/compose.yaml config
```

Validate Operations OpenAPI and run composed tests on a disposable PostgreSQL
18 database with exact row-count/readback and secret-exposure inspection.

## Risks, Decisions, and Sequencing

- Tracker order W4-03 -> W4-04 -> W4-05 conflicts with the canonical rule that
  verify may not ship before eligibility and activation exist. Recommended:
  execute W4-04 and W4-05 internal primitives first, then activate W4-03 to
  compose/publish the route. Approve and synchronize that rebaseline separately.
- The zero-price and unresolved-participant gates from W4-01/W4-05 can block a
  valid Purchase from verification; they require canonical answers.
- Define one global lock order for evidence reacquisition and verification to
  avoid deadlocks with checkout. Reuse Event-then-Offering whenever quota totals
  are contested, then lock Purchase/Payment/reservation/positions consistently.
- A single parent audit row is recommended; per-participant audit rows duplicate
  activation history/outbox without adding authority.
- Replay retention must satisfy ADR-041 and be explicit before implementation.

## Deliverables and Final Report

Deliver the two composed Operations commands, contract-aligned responses,
atomic database effects, focused/concurrent integration proof, and route wiring.
The final report must enumerate before/after states and exact row counts, audit/
outbox payload fields, authorization ordering, replay behavior, failures,
variances, and deferred policy.
