# Task: W3-03 Implement Canonical Purchase with COMMON Channel

## Executed

## Status

Completed and verified on 2026-08-21. W3-01 and W3-02 were verified before
BUILD; no commit or push was requested.

## Tracker

- Week: Week 3
- Epic: Common Purchase
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 7 hours
- Tracker objective: Implement canonical Purchase with `COMMON` channel.
- Dependencies: W2-01, W2-02, and W3-02.
- Acceptance summary: a valid direct checkout produces one canonical Purchase
  in `PENDING_PAYMENT` status.

## Objective

Add the backend Purchasing vertical slice for the accepted canonical Purchase
model. The slice persists and reads `COMMON` Purchases and their actor and
participant relationships, starts each valid Purchase in `PENDING_PAYMENT`,
and exposes the already-contracted authorized Operations reads. Public command
composition remains W3-06 so no half-safe checkout endpoint is published here.

## Source of Truth

- `docs/PRD.md`, Common Purchase, Purchase, status, and actor rules;
- `docs/PRODUCT_MAP.md` Purchasing ownership and Week 3 roadmap;
- `docs/ARCHITECTURE.md`, Purchasing bounded context, transaction boundaries,
  API separation, errors, pagination, and deployment model;
- `docs/DECISIONS.md`, especially ADR-009, ADR-011 through ADR-014, ADR-018,
  ADR-019, ADR-020, ADR-024, ADR-033, ADR-038, ADR-041 through ADR-045;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- `docs/security/PERMISSIONS.md`;
- `docs/database/ERD.md` and migrations `0002` and `0005`;
- both OpenAPI contracts;
- W2 Event/Offering outputs and approved W3-01/W3-02 outputs;
- `.codex/CURRENT_STATE.md` and current source.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate only after dependencies and the role-linking contract are approved
and verified. Do not commit or push unless requested.

## Existing State

- PostgreSQL already contains `purchases`, `purchase_status_history`, and
  `purchase_participants` with the required Party and Offering/Event foreign
  keys, channel/status checks, snapshots, version, and uniqueness constraints.
- `quota_reservations` and access-token hashing columns already exist but are
  owned by W3-05/W3-06 command composition.
- No Go `purchasing` module, Purchase repository, handler, or registered route
  exists.
- Storefront and Operations OpenAPI Purchase routes are contract-only.
- Existing Event/Offering modules demonstrate the repository, explicit DTO,
  keyset pagination, transaction, and Gin route patterns to reuse.

## Scope

### In Scope

- Add a focused `purchasing` domain/application/repository slice for canonical
  Purchases with channel `COMMON`.
- Validate the initial aggregate shape: one Event, one Offering, explicit
  purchaser and payer relationships, one or more ordered intended
  participants, supported currency, and initial `PENDING_PAYMENT` state.
- Persist a Purchase, its participant rows, and its initial status-history row
  within a caller-owned transaction.
- Reject non-`COMMON` creation in this Week 3 command path without narrowing the
  shared database model needed by later channels.
- Read a Purchase by UUID and list Purchases with the Operations contract's
  bounded filters, deterministic ordering, and opaque keyset cursor.
- Register Operations list/detail routes behind `purchase.read` and return
  explicit Operations DTOs from stored snapshots/Party references.
- Add focused domain, repository, handler, authorization, pagination, and
  disposable PostgreSQL tests.

### Out of Scope

- Publishing `POST /api/public/v1/purchases`; W3-06 owns the atomic public
  command, token, reference, idempotency, reservation, outbox, and replay.
- Purchase cancellation, evidence upload, verification, payment events,
  activation, expiry, completion, or refund transitions.
- Creating `sohibul_qurban`, ledger entries, payment attempts, or payment
  evidence.
- `SAVING` or `GIVEAWAY` commands and their source records.
- Editable Purchases, carts, multi-offering Purchases, bulk commands, exports,
  or a generic lifecycle/repository/command framework.
- A dashboard projection or frontend implementation.

## Target State and Invariants

- One canonical Purchase aggregate represents the direct checkout; there is no
  parallel Common-order table.
- The Week 3 constructor accepts only `COMMON` and always starts at
  `PENDING_PAYMENT` with the initial version required by the schema.
- Purchaser, payer, and intended participant roles follow the explicit W3-02
  mapping and never collapse by contact comparison.
- A Purchase contains exactly one Offering and the Offering belongs to the
  persisted Event.
- Participant count equals the number of ordered participant rows.
- Initial Purchase and its `DRAFT -> PENDING_PAYMENT` history entry succeed or
  fail together.
- Operations reads return historical commercial and participant-name values
  from Purchase snapshots. Purchaser/payer summaries may resolve the referenced
  current Party because the accepted schema does not snapshot those identities.
- List order is `created_at DESC, id ASC`; cursors are bounded and opaque.

## Implementation Plan

1. Recheck W3-01/W3-02 outputs and the actual migration columns before BUILD.
2. Add a small Purchase aggregate and constructor with only accepted initial
   state/channel/relationship/count invariants.
3. Add PostgreSQL create/get/list methods using explicit SQL and the current
   transaction-compatible DB pattern. Insert Purchase, participants, and the
   accepted `DRAFT -> PENDING_PAYMENT` history entry in one caller-owned
   transaction.
4. Map duplicate, foreign-key, constraint, not-found, and cursor failures once
   to stable application errors; never expose driver text.
5. Add explicit Operations DTO mapping and list/detail handlers protected by
   `purchase.read`; keep authorization before repository access.
6. Register only the two Operations read routes and their explicit OPTIONS
   paths. Avoid coupling registration to unrelated Event/Offering handlers.
7. Add focused unit, repository, handler, permission, and pagination tests.
8. Run focused/full verification and review the contract/runtime diff.

## Planned File Changes

Expected minimum:

- `apps/api/internal/purchasing/domain.go`;
- `apps/api/internal/purchasing/service.go` only if orchestration is more than a
  direct domain/repository call;
- `apps/api/internal/purchasing/postgres.go`;
- `apps/api/internal/purchasing/operations_http.go`;
- focused `_test.go` files;
- `apps/api/internal/app/server.go`;
- `apps/api/internal/app/router.go` and/or route registration files only as
  required by the existing constructor flow;
- `apps/api/internal/app/routes_operations.go`.

Do not add an interface solely for the PostgreSQL implementation. Introduce a
narrow test seam only where current handler/service tests actually require it.

## Operations API Contract

- `GET /api/operations/v1/purchases` requires `purchase.read`, validates Event
  and status filters, uses the contract limit bound, and returns a next cursor
  only when another page exists.
- `GET /api/operations/v1/purchases/{purchase_id}` requires `purchase.read`,
  validates the UUID before querying, and returns contract-defined not-found
  behavior without leaking unrelated records.
- Responses use explicit purchaser/payer summaries and stored Purchase
  snapshots. They do not serialize database rows or auth/session context.
- Operations responses retain `Cache-Control: no-store`, stable envelopes, and
  request IDs.
- No write route is registered in this task.

## Database and Transaction Impact

- Reuse existing tables and constraints; no migration is currently justified.
- Repository create methods must participate in the caller's transaction and
  must not commit independently.
- W3-03 establishes the minimal Purchase/status/participant write. W3-06 wraps
  it with Party resolution/creation, snapshots, quota reservation, token,
  outbox, and idempotency replay in the final atomic command.
- Read queries must avoid N+1 Party/participant fetches for list/detail while
  staying explicit and comprehensible. Do not introduce a data-loader layer.

## Authorization, Audit, and Security

- Operations reads require backend `purchase.read`; frontend visibility is not
  authorization.
- Missing/expired auth is `401`; authenticated lack of permission is `403`.
- No privileged mutation occurs in this task, so no audit/outbox row is written
  by the reads.
- Do not expose access-token hashes, raw token material, idempotency keys,
  internal errors, or full Party contact data beyond the contract.

## Acceptance Criteria

- The domain accepts a valid `COMMON` aggregate and rejects unsupported
  channels or invalid initial state/relationships/counts.
- Persistence atomically creates one Purchase, its intended participants, and
  its initial `PENDING_PAYMENT` history row.
- A failed child/history insert leaves no partial Purchase.
- Operations list/detail routes match the contract, require `purchase.read`,
  paginate deterministically, and use stored snapshots.
- Unauthorized/forbidden/not-found/invalid cursor cases return stable contract
  errors without repository leakage.
- No public Purchase write route or future lifecycle behavior is exposed.
- No speculative table, migration, dependency, or framework is added.

## Verification

```bash
cd apps/api
go test ./internal/purchasing/... ./internal/app/...
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
docker compose -f infrastructure/compose.yaml config
```

Also validate both OpenAPI contracts and run the disposable PostgreSQL
migration/repository workflow. Distinguish pre-existing failures from task
regressions.

## Dependencies and Sequencing

- W3-01 supplies Party persistence.
- W3-02 supplies approved explicit role mapping.
- W2 Event/Offering provides valid source records.
- W3-04 adds accepted immutable snapshot capture.
- W3-05 adds quota reservation.
- W3-06 composes and publishes the public command.
- W3-08 consumes the Operations reads; W3-09 adds cross-cutting regression and
  contention coverage rather than replacing this task's focused tests.

## Decisions, Assumptions, and Deferred Work

- Accepted: all channels converge on canonical Purchase; only `COMMON` is
  created in Week 3.
- Accepted: initial direct-checkout status is `PENDING_PAYMENT`.
- Assumption: Operations Purchase list/detail are implemented here because the
  contracted W3-08 frontend requires a backend owner and no later backend task
  claims those reads. Reconfirm during plan review.
- Deferred: every later lifecycle command and non-Common source workflow.

## Final Review Requirements

Report domain invariants, routes, permissions, SQL/transaction effects,
contract conformance, checks run, and deferred lifecycle behavior. Do not claim
checkout completion until W3-04 through W3-06 are also implemented and
verified.

## Final Review

### Implemented

- A focused canonical Purchase aggregate that always creates a `COMMON`,
  `PENDING_PAYMENT`, version-one record with explicit purchaser, optional
  payer, ordered participant snapshots, bounded integer amounts, and a hashed
  access-token input.
- PostgreSQL create/get/list persistence. Create writes the Purchase,
  participants, and `DRAFT -> PENDING_PAYMENT` history through the
  caller-owned transaction; reads use stored commercial snapshots and current
  Party display summaries.
- Authorized Operations `GET /purchases` and `GET /purchases/{purchase_id}`
  routes behind `purchase.read`, including bounded filters, opaque keyset
  pagination, explicit DTOs, and participant detail only.

### Verified

- Focused unit, repository, handler, authorization, pagination, and
  disposable-PostgreSQL integration coverage passed, including transaction
  rollback and authenticated Operations reads.
- `go vet ./...`, database-backed `go test ./...`, and `go build ./...`
  passed from `apps/api`.
- `make validate`, `docker compose -f infrastructure/compose.yaml config`,
  `git diff --check`, and Redocly validation of the Operations contract passed.
  Redocly reports only five existing Operations warnings: missing license,
  two redirect-only authentication operations, and two unused response
  components.

### Assumed

- This task establishes only the direct persistence shape. The future public
  command is the sole writer and will supply a validated Event/Offering pair,
  Party resolution, snapshots, reservation, token, outbox, and idempotency
  within its command transaction.

### Deferred

- W3-04 verifies and captures Event/Offering source snapshots and exact total.
- W3-05 reserves quota; W3-06 publishes the atomic public checkout command.
- No payment, evidence, cancellation, eligibility, lifecycle transition,
  non-COMMON channel, or dashboard behavior was added.
