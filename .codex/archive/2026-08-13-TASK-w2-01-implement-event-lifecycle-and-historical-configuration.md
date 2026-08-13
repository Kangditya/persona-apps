# Task: W2-01 Implement Event Lifecycle and Historical Configuration

## Executed

## Status

Executed and verified for the framework-neutral Event domain boundary on
2026-08-13.

## Tracker

- Week: Week 2
- Epic: Event & Offering
- Week goal: Allow Operations to configure commerce and Storefront to discover
  it.
- Milestone: Public Event and Offering discovery.
- Weekly allocation: 42 hours shared across W2-01 through W2-08; the tracker
  does not assign per-task estimates, so this plan does not invent them.
- Tracker objective: Implement Event lifecycle and historical configuration.
- Dependency: Verified Week 1 platform, schema, lifecycle, permission, and
  OpenAPI foundations.

## Objective

Implement the Event domain and application behavior that owns annual event
configuration and enforces the accepted Phase 1 lifecycle without allowing
future edits to rewrite historical truth.

## Context

The lifecycle is already authoritative in
`docs/domain/COMMERCE_LIFECYCLES.md`:

```text
DRAFT -> PUBLISHED -> ACTIVE <-> SUSPENDED -> CLOSED -> ARCHIVED
```

Migration 0001 already creates `qurban_events`; migration 0005 already adds
`SUSPENDED` and the partial unique index that permits at most one `ACTIVE`
Event. No runtime Event module or command currently exists.

## Source of Truth

- `docs/PRD.md` sections 6.1, 7.1, 10, 11.1, 12, and 13;
- `docs/PRODUCT_MAP.md` sections 8, 11, 13, and 14;
- `docs/ARCHITECTURE.md` sections 7.1, 9, 10, 13, 14, 18, and 28;
- `docs/DECISIONS.md` ADR-009, ADR-011, ADR-019, ADR-020, ADR-024,
  ADR-038, ADR-041, ADR-042, ADR-044, and ADR-046;
- `docs/domain/COMMERCE_LIFECYCLES.md` Event policy;
- `docs/security/PERMISSIONS.md` Event route mapping;
- `apps/api/migrations/0001_qurban_foundation.up.sql` and
  `0005_mvp_commerce_safety.up.sql`;
- `contracts/openapi/operations.yaml` Event schemas and operations;
- `.codex/CURRENT_STATE.md` for the implementation gap.

If these sources conflict during activation, stop and amend the canonical
owner before BUILD.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate this draft by copying its reviewed content into a new
`.codex/TASK.md`. Revalidate the listed sources and the worktree first. Archive
only after its acceptance criteria pass. Do not commit or push unless asked.

## Scope

### In Scope

- Event value types, validation, lifecycle guards, and framework-neutral
  application inputs/outputs.
- Create, update, publish, activate/reactivate, suspend, close, and archive
  command behavior.
- The confirmed configuration-edit policy:
  - configuration may change in `DRAFT`, `PUBLISHED`, or `SUSPENDED`;
  - an `ACTIVE` Event must be suspended before configuration changes;
  - `CLOSED` and `ARCHIVED` Events are immutable.
- At-most-one-active behavior, backed by the existing partial unique index and
  translated to a stable state conflict rather than a raw SQL error.
- Historical configuration through append-only audit rows with minimized
  before/after configuration and lifecycle status.
- Application orchestration for same-transaction Event state, audit, outbox,
  and replay effects; W2-03/W2-05 provide and prove the concrete PostgreSQL/HTTP
  composition.
- Version-aware command inputs, provided by W2-03, so stale Operations edits
  cannot silently overwrite current state.

### Out of Scope

- Dedicated Event version-history tables or historical-administration UI.
- Event locations, slaughter dates, allocation, event-day readiness, or
  distribution.
- Per-event timezone configuration; timestamps remain UTC instants and clients
  display an explicit local timezone.
- Offering behavior, HTTP handlers, frontend pages, dashboard projections, and
  Purchase snapshot implementation.
- Background expiration, notification, or event-worker implementation.

## Existing State

- SQL already stores UUID, unique event year, name, status, optional
  registration window, optional participant quota, and timestamps.
- SQL validates window ordering and the accepted status vocabulary.
- SQL already prevents two rows from being `ACTIVE` concurrently.
- Event and Offering lifecycle changes use `audit_log` for transition history;
  there is intentionally no Event status-history table.
- The platform already provides Gin composition, PostgreSQL access,
  transaction handling, audit writing, outbox storage, structured errors,
  authorization, and encrypted idempotency replay.
- None of that platform code currently executes an Event business command.

## Target State

- Event behavior is implemented in one meaningful module under
  `apps/api/internal/event`, with Gin and SQL absent from its domain layer.
- Event creation produces `DRAFT`; lifecycle commands accept only the exact
  transitions in the authoritative matrix.
- `activate` accepts both `PUBLISHED -> ACTIVE` and
  `SUSPENDED -> ACTIVE`; no extra `reactivate` route is introduced.
- Event activation is not conditioned on having a published Offering because
  no accepted requirement establishes that rule.
- Registration opening/closing is separate from Event lifecycle. Discovery
  may remain visible outside the window; future checkout is responsible for
  rejecting commerce outside it.
- Privileged transitions write actor, permission, source, request ID, target,
  and minimized before/after values atomically with Event state and outbox.
- Archived data remains readable through Operations queries but cannot be
  changed.

## Constraints

- Reuse the current `database/sql`, pgx, transaction, audit, idempotency, and
  auth foundations; add no ORM, event framework, repository framework, or
  dependency.
- Domain code receives `context.Context` through application boundaries, not
  `*gin.Context`, `*sql.Tx`, or transport DTOs.
- Use explicit Event status constants and transition functions. Do not create a
  global status enum or generic state-machine abstraction.
- Validate trimmed names, supported years, window order, and non-negative
  quota at the trust boundary and again where domain integrity requires it.
- API-visible integer values must remain within JavaScript's safe integer
  range; PostgreSQL constraints remain the final integrity boundary.
- Never audit secrets, session/CSRF values, raw idempotency keys, or unrelated
  operator claims.

## Implementation Requirements

### Domain behavior

- Define `Event`, `Status`, configuration inputs, and narrow errors for invalid
  input, invalid transition, stale version, duplicate year, and active-event
  conflict.
- Keep `event_year` immutable after creation, matching the existing Operations
  patch contract and preserving annual identity.
- Keep transition validation pure and table-driven only if the table is simpler
  than explicit branches.
- Preserve current configuration when a patch omits a field; support explicit
  clearing of nullable windows/quota after W2-05 corrects the contract shape.
- Require `expected_version` for update and lifecycle commands once W2-03 adds
  persistence versioning.

### Application commands

- Commands load authoritative state, enforce permission through the Operations
  adapter/application policy, validate expected version, and call the owning
  repository operation.
- Event create and lifecycle commands are retry-sensitive and use the
  command-scoped idempotency contract finalized in W2-05. Authorization
  precedes replay. Patch/update relies on expected version and is not stored as
  a replayable command.
- Emit Event lifecycle outbox types already named by the authoritative policy:
  `EventPublished`, `EventActivated`, `EventSuspended`, `EventClosed`, and
  `EventArchived`. If create/update need events, name and document them during
  activation instead of guessing here.

### History and immutability

- Use `audit_log` for Event creation, configuration edits, and transitions.
- Before/after payloads contain only public configuration, lifecycle status,
  and version; omit credentials and internal session information.
- Do not mutate or delete prior audit rows.
- Purchase-level historical pricing/configuration remains W3-04; Week 2 must
  not claim that uncreated Purchases are already historically snapshotted at
  runtime.

## Planned File Changes

Exact file splitting may be reduced during activation, but ownership remains:

| File | Action | Purpose |
| --- | --- | --- |
| `apps/api/internal/event/domain.go` | Create | Event model, status vocabulary, validation, and transitions. |
| `apps/api/internal/event/service.go` | Create | Framework-neutral create, update, and lifecycle commands. |
| `apps/api/internal/event/errors.go` | Create or fold into the two files above | Stable domain/application errors only if separate placement improves clarity. |
| `apps/api/internal/event/*_test.go` | Create | Focused domain and command tests. |
| `.codex/CURRENT_STATE.md` | Modify when executed | Record implemented Event behavior without claiming later routes/screens. |
| `.codex/TASK.md` | Create, execute, archive | Preserve approved execution and final review. |

Repository persistence belongs to W2-03; Operations HTTP registration belongs
to W2-05. Do not add placeholder files for those layers in this task.

## Dependencies and Sequencing

- W2-01 and W2-02 may define domain behavior independently.
- W2-03 supplies concrete repository implementation and the version migration
  required by command execution.
- W2-05 exposes W2-01 commands through the Operations contract.
- W2-04, W2-06, and W2-07 consume Event reads after repository support exists.
- W2-08 closes the cross-layer verification matrix.

## API and Database Impact

- This task defines application behavior but does not independently modify the
  OpenAPI files or migrations.
- W2-03 must add Event `version` storage; W2-05 must add the corresponding
  Operations DTO and expected-version command fields.
- The existing public Event DTO remains intentionally smaller than the
  Operations Event DTO.

## Authorization, Audit, Idempotency, and Concurrency

- Reads require `event.read`; changes require `event.manage` when transported
  through Operations routes.
- Backend permission enforcement is authoritative; frontend visibility is not.
- Every privileged change writes audit in the owning transaction. Audit failure
  aborts the command.
- Lifecycle commands are replay-safe through the platform executor. Same key
  and canonical request replays; another request conflicts.
- Version checks and the database partial unique index protect contested state;
  idempotency does not replace either mechanism.

## Acceptance Criteria

1. Every accepted Event transition and every rejected transition is explicit
   and matches the authoritative lifecycle.
2. `activate` handles publication and reactivation while at most one Event can
   become active.
3. Active, closed, and archived configuration edits follow the confirmed
   freeze policy.
4. Stale versions, duplicate years, invalid windows, invalid quota, and
   concurrent activation map to stable validation/conflict errors.
5. Application tests prove the command boundary invokes state, audit, outbox,
   and replay effects exactly as declared and performs none after rejected
   validation/transition/version checks; W2-05 owns real-database atomicity.
6. Event history is append-only and historical Purchase snapshot work remains
   explicitly deferred.
7. No HTTP framework, SQL driver, provider type, new dependency, or generic
   state-machine abstraction enters the domain layer.

## Testing

- Table-test all lifecycle edges, including forbidden shortcuts and terminal
  states.
- Test configuration-edit status guards, validation boundaries, and nullable
  clearing semantics at the application input boundary.
- Test stale expected versions and application-side error/side-effect mapping
  with narrow fakes. W2-03 owns real duplicate/concurrent repository proof;
  W2-05 owns PostgreSQL audit rollback and create/lifecycle replay proof.
- Prefer the smallest standard-library tests; no new test framework or fixture
  system.

## Verification

When this task is activated with its dependency-ready implementation:

```bash
cd apps/api && go run ./cmd/format -check
cd apps/api && go vet ./...
cd apps/api && go test ./internal/event/... ./internal/platform/...
cd apps/api && go test ./...
cd apps/api && go build ./...
make validate
```

This task may archive after its scoped domain/application criteria pass. It
must not claim the database one-active-event or transaction invariant; W2-03,
W2-05, and W2-08 prove those against disposable PostgreSQL 18.

## Deliverables

- Reviewed Event domain and application implementation.
- Focused lifecycle and command tests.
- Current-state update and an archived executed task only after verification.

## Risks and Clarifications to Review

- Confirmed default: Event configuration is editable only in `DRAFT`,
  `PUBLISHED`, or `SUSPENDED`; active configuration requires suspension first.
- Confirmed default: dedicated Event history tables are deferred; append-only
  audit is the Week 2 historical record.
- Confirmed default: no per-event timezone or activation-requires-offering rule
  is added.
- Remaining implementation risk: W2-03 and W2-05 must preserve nullable patch
  semantics and expected-version behavior end to end.
- Canonical audit retention/export policy is still deferred. Week 2 proves
  append-only application history, not indefinite backup/compliance retention.

## Final Report Requirements

Distinguish implemented domain behavior, verified database behavior,
assumptions, deferred history/UI work, plan variance, remaining lifecycle and
concurrency risks, and the next dependency-ready task.

## Execution Review

### Implemented

- Added the framework-neutral `internal/event` model, exact lifecycle commands,
  configuration-freeze policy, nullable patch-clearing inputs, version checks,
  and safe-integer/domain validation.
- Each command produces minimized before/after audit state, an action, explicit
  lifecycle outbox type where the lifecycle policy names one, and retry intent;
  it does not write effects outside a transaction.
- Recorded the Event-only implementation boundary in `.codex/CURRENT_STATE.md`.

### Verified

- Focused Event format check, `go vet`, and lifecycle/validation tests pass.
- Full API `go test ./...` and `go build ./...` pass with the new module.

### Assumed

- The confirmed configuration policy permits edits only in `DRAFT`,
  `PUBLISHED`, and `SUSPENDED`; activation has no Offering prerequisite.

### Deferred

- PostgreSQL Event rows, one-active conflict translation, transaction wiring,
  audit/outbox writes, authorization, idempotency replay, and Operations HTTP
  commands remain owned by W2-03 and W2-05.
- Purchase snapshots and Event history UI remain outside this task.

### Remaining Risks and Next Task

- The returned effect metadata must be persisted in the same transaction as a
  version-checked Event update; no database invariant is claimed here.
- Next: W2-02 implements the similarly framework-neutral Offering rules.
