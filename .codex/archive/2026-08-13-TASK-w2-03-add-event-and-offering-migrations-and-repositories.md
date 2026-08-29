# Task: W2-03 Add Event and Offering Migrations and Repositories

## Executed

## Status

Executed — 2026-08-13.

## Tracker

- Week: Week 2
- Epic: Event & Offering
- Week goal: Allow Operations to configure commerce and Storefront to discover
  it.
- Milestone: Public Event and Offering discovery.
- Weekly allocation: 42 hours shared across W2-01 through W2-08; the tracker
  does not assign per-task estimates, so this plan does not invent them.
- Tracker objective: Add Event and Offering migrations and repositories.
- Dependencies: W2-01 and W2-02 domain/application contracts.

## Objective

Implement the minimum PostgreSQL delta and concrete persistence needed by the
reviewed Event and Offering behavior, without recreating existing tables,
editing shared migrations, or introducing a generic repository framework.

## Context

The tracker title predates current source. Event and Offering schema already
exists in migrations 0001 and 0005. The real remaining migration need is
optimistic concurrency and query/index support discovered from the Week 2
runtime contracts. No Go business repository currently exists.

## Source of Truth

- `docs/ARCHITECTURE.md` sections 7, 9, 10, and 18;
- `docs/DECISIONS.md` ADR-009, ADR-011, ADR-024, ADR-040, ADR-041,
  ADR-042, and ADR-044;
- `docs/CONVENTIONS.md` sections 4.1 through 4.4 and 6;
- `docs/database/ERD.md`, `MIGRATION_PLAN.md`, `OPERATIONS.md`, and
  `REVIEW.md`;
- migrations 0001 through 0005 and their down pairs;
- `apps/api/internal/database/migration` and `cmd/db` lifecycle;
- W2-01 and W2-02 reviewed domain/application interfaces;
- both OpenAPI contracts' deterministic list orders.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Before activation, inspect the latest migration number, actual SQL, repository
interfaces emerging from W2-01/W2-02, and any user changes. Generate the next
migration using the existing database CLI. Never edit migrations 0001-0005.

## Scope

### In Scope

- One additive migration pair adding `version bigint NOT NULL DEFAULT 1 CHECK
  (version > 0)` to `qurban_events` and `offerings`.
- Database upper bounds for API-visible Event/Offering bigint values so the
  exact JSON safe-integer rule is enforced below the application layer.
- The narrow indexes required by verified keyset list and public visibility
  queries, only after confirming existing indexes are insufficient.
- Event repository behavior for create, get, deterministic list, patch, and
  lifecycle transitions.
- Offering repository behavior for create, get, event-scoped deterministic
  list, patch, lifecycle transitions, and public catalogue reads.
- Availability aggregation from Event/Offering quota and reservation usage.
- Transaction-aware repository methods used by application commands so state,
  audit, outbox, and idempotency commit once.
- Stable translation of no-row, unique, check, foreign-key, and stale-version
  outcomes to application errors.
- PostgreSQL integration tests for real constraints, rollback, ordering,
  availability, and concurrent conflicts.

### Out of Scope

- Recreating or renaming existing Event/Offering tables.
- Editing historical migration pairs or adding seed rows for business data.
- An ORM, SQL generator, generic base repository, unit-of-work abstraction,
  cache, Redis, or stored availability counter.
- Purchase creation, quota reservation mutation, payment, participant, or
  worker repository behavior.
- Full-text search, arbitrary sorting, total counts, or offset pagination.

## Existing State

- Migrations 0001-0005 are verified by the explicit `golang-migrate` runner
  and CI PostgreSQL 18 lifecycle.
- `qurban_events` has unique year, window/quota checks, status index, and a
  partial unique active index.
- `offerings` has event-scoped unique code, price/currency/capacity/status
  checks, optional quota, and `(event_id, status)` index.
- `quota_reservations` has active and expiring partial indexes.
- Neither Event nor Offering has a `version` column.
- The platform uses `database/sql` with pgx; no query builder or generator is
  installed.

## Target State

- Migration 0006 or the next verified free sequence adds optimistic version
  columns, exact-integer constraints, and only evidence-backed indexes.
- Repositories live with their owning modules and expose only methods required
  by W2-01 through W2-05.
- Command updates use one atomic predicate such as `WHERE id = $id AND version
  = $expected`, increment version on success, and distinguish not-found from
  stale state without overwriting newer data.
- Lifecycle commands lock/revalidate parent and target state when needed.
- Event list order is `event_year DESC, id ASC`; Offering list order is `code
  ASC, id ASC`; opaque cursors encode typed sort values and ID without exposing
  SQL or accepting arbitrary clauses.
- Public queries return only `ACTIVE` Event and `PUBLISHED` Offerings.
- Availability is computed from one authoritative query snapshot and returned
  as nullable when no finite quota exists.
- Offering list/detail reads that expose availability aggregate it in bounded
  SQL rather than issuing one reservation query per Offering.

## Constraints

- Use parameterized SQL only. Cursor values become typed query parameters,
  never SQL fragments.
- Keep SQL in concrete Postgres adapters; domain and application layers do not
  depend on driver errors or row structs.
- Repository methods participating in a command accept the caller's
  transaction or a minimal local query interface justified by both `*sql.DB`
  and `*sql.Tx`; do not add an interface merely for one implementation.
- Append-only audit/outbox/idempotency records are written by their owning
  application transaction, not hidden behind repository auto-magic.
- Use keyset pagination with a bounded limit of 1-100 for Operations lists. The
  public Offering list follows the explicit size/pagination decision required
  by W2-04; do not silently reuse an Operations cursor contract or issue an
  unbounded query.
- Down migration is bounded and explicit; it must not silently discard a
  version value if rollback compatibility is unsafe. Because these are new
  columns with default 1 and no shared data is recorded, the reviewed rollback
  may drop them in disposable/test environments.
- The Event/Offering quota, price, and version maximum is
  `9_007_199_254_740_991`, matching exact JSON numbers. Inspect existing rows
  before adding constraints; do not let a migration surprise a shared database
  with an opaque check violation.

## Implementation Requirements

### Migration

- Create the pair with `go run ./cmd/db migrate create
  add_event_offering_versions_and_bounds` or an equally descriptive verified
  name.
- Add Event and Offering version columns atomically.
- Add named maximum constraints for `qurban_events.participant_quota`,
  `offerings.price_minor`, `offerings.participant_quota`, and both version
  columns. Retain existing non-negative/positive checks; use names that can be
  translated and rolled back exactly.
- Compare intended queries against existing indexes using `EXPLAIN` on
  representative test data before adding composite indexes. Likely candidates
  are Event order/status and Offering event/code/public visibility, but BUILD
  must not add them without query evidence.
- Update ERD/migration documentation only where it describes the changed
  current target schema; do not copy application rules into database docs.

### Repository mapping

- Use dedicated persistence row structs and explicit scanners/mappers where
  they prevent database types from leaking into the domain.
- Normalize nullable timestamps, descriptions, and quotas deliberately.
- Decode cursor values with strict version/type/size validation and return
  `invalid_request` for malformed cursors at the adapter boundary.
- Fetch `limit + 1` rows to derive `next_cursor`; do not run a second total
  count query.

### Mutation safety

- Create maps unique event year or event-scoped Offering code violations to a
  conflict.
- Event activation relies on the existing partial unique index as the final
  concurrent guard.
- Offering commands revalidate the parent Event in the same transaction.
- Patch and lifecycle updates increment `version`; zero affected rows trigger a
  follow-up existence check to distinguish not-found from stale/invalid state.
- Avoid database-error string matching. Use pgx/PostgreSQL error codes and
  named constraints when translation is necessary.

### Availability

- Sum `participant_units` for reservations in `RESERVED` or `CONSUMED` states
  at Event and Offering scope.
- Use `COALESCE` for sums but preserve nullable quota fields.
- Clamp finite remaining results at zero.
- Keep public availability advisory; do not lock rows for a read-only catalogue
  calculation or claim it reserves capacity.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `apps/api/migrations/0006_add_event_offering_versions_and_bounds.up.sql` | Create using next verified number | Add optimistic versions, exact bigint bounds, and evidence-backed indexes. |
| `apps/api/migrations/0006_add_event_offering_versions_and_bounds.down.sql` | Create | Provide explicit bounded rollback for the same unit. |
| `apps/api/internal/event/postgres.go` | Create | Concrete Event persistence and cursor queries. |
| `apps/api/internal/event/postgres_integration_test.go` | Create | Verify Event SQL, constraints, ordering, and conflicts. |
| `apps/api/internal/offering/postgres.go` | Create | Concrete Offering persistence and availability queries. |
| `apps/api/internal/offering/postgres_integration_test.go` | Create | Verify Offering SQL, visibility, ordering, and availability. |
| `docs/database/ERD.md` | Modify | Record version columns and any accepted index delta. |
| `docs/database/MIGRATION_PLAN.md` | Modify | Inventory the new additive migration. |
| `docs/database/REVIEW.md` | Modify | Record constraint/index rationale and rollback risk. |
| `.codex/CURRENT_STATE.md` | Modify when executed | Record applied source artifacts without claiming shared deployment. |
| `.codex/TASK.md` | Create, execute, archive | Preserve approved execution and final review. |

File names may be consolidated if the implementation stays clearer, but Event
and Offering ownership must not be moved into a generic repository package.

## Dependencies and Sequencing

- Activate after W2-01/W2-02 define the exact repository needs.
- W2-04 and W2-05 depend on these concrete reads/mutations.
- W2-06/W2-07 depend on stable endpoint behavior, not directly on SQL.
- W2-08 extends but does not postpone critical repository integration tests;
  the behavior-changing migration must leave focused tests here.

## API and Database Impact

- Adds version columns; W2-05 must expose them in Operations responses and
  require expected versions for mutations.
- Public DTOs need not expose internal concurrency versions.
- No existing identifier, status, route, or business reference changes.
- Index additions must match actual filters and deterministic sort order.

## Authorization, Audit, Idempotency, and Concurrency

- Repositories do not decide caller permissions; application/HTTP policy does.
- Repository mutations run inside the caller's transaction so required audit,
  outbox, and replay records are atomic.
- Version predicates prevent lost updates. The one-active partial index
  prevents concurrent active Events. Later checkout row locks protect quota.
- Repositories never look up or replay idempotency responses themselves.

## Acceptance Criteria

1. Existing migrations are untouched and the new pair validates, applies,
   rolls back one step, and reapplies cleanly on disposable PostgreSQL 18.
2. Event and Offering version columns default/check correctly and stale
   mutations cannot overwrite current rows.
3. Concurrent Event activation cannot produce two active rows and yields a
   stable conflict to the application.
4. Event and Offering list queries are bounded, deterministic, cursor-based,
   and resistant to SQL injection through cursor/filter input; public catalogue
   bounding follows W2-04's reviewed endpoint-specific decision.
5. Public query methods cannot return non-active Events or non-published
   Offerings.
6. Availability aggregation implements the confirmed status and nullable-quota
   semantics.
7. Audit/outbox/replay can share the caller transaction with repository state
   changes.
8. No ORM, generator, generic repository, cache, unbounded query, or business
   seed is added.

## Testing

- Migration validation and lifecycle from an empty database.
- Create/get/list/patch and not-found/duplicate/stale cases for both modules.
- Cursor first page, next page, malformed cursor, tie-break order, and rows
  inserted between pages.
- Concurrent activation and concurrent same-version update.
- Public visibility for every Event/Offering status combination.
- Availability across reservation states and nullable quotas.
- Transaction rollback when audit/outbox/application work fails after a
  repository mutation.

## Verification

Against a disposable PostgreSQL 18 database:

```bash
cd apps/api && go run ./cmd/db migrate validate
cd apps/api && go run ./cmd/db migrate up
cd apps/api && go run ./cmd/db migrate version
cd apps/api && TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/event/... ./internal/offering/...
cd apps/api && go run ./cmd/db migrate down --steps 1
cd apps/api && go run ./cmd/db migrate up --steps 1
cd apps/api && go run ./cmd/db migrate version
cd apps/api && go test ./...
make validate
```

Use an explicitly disposable database; never run rollback against an ambiguous
or shared target.

## Deliverables

- One additive migration pair.
- Concrete Event and Offering Postgres adapters with focused integration tests.
- Updated database/current-state documentation and verified task archive.

## Risks and Clarifications to Review

- The tracker says “add migrations,” but current source already has the base
  schema. This plan deliberately adds only the missing concurrency delta.
- Index candidates remain evidence-gated until actual queries can be explained.
- Public availability is a snapshot and may change immediately after a read.
- Migration rollback drops new version columns; execute it only under the
  repository's existing destructive-command policy.

## Final Report Requirements

Distinguish migration changes, PostgreSQL-backed verification, assumptions,
index evidence, rollback risk, deferred quota mutation work, plan variance,
and the next API task.

## Plan Variance

- Planned work: create additive migration `0006` and validate the established
  migration lifecycle.
- Unexpected requirement: `internal/database/cli/cli_test.go` asserts the
  historical literal `valid: 5 migrations`.
- Necessity: the newly required migration correctly changes that observable
  count, so leaving the assertion unchanged would make the repository's full
  validation fail despite a valid migration pair.
- Impact: update the test to derive the expected discovered count from the
  canonical migration directory; no production behavior, migration logic, or
  public contract changes.

## Execution Review

### Implemented

- Created additive migration `0006_add_event_offering_versions_and_bounds`.
  It adds positive, JSON-safe Event and Offering versions; JSON-safe Event and
  Offering quotas and Offering prices; and the partial
  `idx_quota_reservations_used_event_offering` aggregation index.
- Added concrete Event and Offering PostgreSQL repositories in their owning
  modules. They use parameterized SQL, bounded typed keyset cursors,
  conditional version updates, pgx SQLSTATE/constraint translation, and the
  caller's `*sql.DB` or `*sql.Tx` through the minimal shared query interface.
- Added repository integration coverage for ordering, malformed cursors,
  duplicates, stale writes, concurrent active-event/update conflicts, public
  visibility, availability status semantics, database bounds, and rollback of
  a larger transaction after a repository mutation.
- Updated the database ERD, migration plan, design review, current state, and
  the migration CLI test so the expected migration count derives from the
  canonical directory.

### Verified

- `go run ./cmd/format -check` passed for all changed Go files.
- Migration validation reported six pairs. On a disposable local PostgreSQL
  18 instance, `0006` was at version 6 and clean, rolled back one bounded step
  to version 5, then reapplied to version 6 and clean.
- `TEST_DATABASE_URL=... go test -count=1 ./internal/event/... ./internal/offering/...`
  and `TEST_DATABASE_URL=... go test -count=1 ./...` passed.
- `go vet ./...` and `go build ./...` passed.
- Representative live query plans retained existing Event/Offering list
  indexes and used the new partial reservation index for `RESERVED` plus
  `CONSUMED` aggregation.

### Assumed

- The verified database is explicitly disposable and local; no claim is made
  about staging or production compatibility.
- Public catalogue pagination is intentionally finalized by W2-04, not by this
  repository task.

### Deferred

- HTTP handlers, response contracts, authorization, audit/outbox writes, and
  idempotent command composition are owned by W2-04/W2-05.
- Quota mutation remains a future transactional checkout concern; catalogue
  availability is advisory and does not reserve capacity.

### Remaining Risk and Next Task

- The 0006 down migration drops version columns and must remain limited to a
  disposable or recovered target.
- Next: W2-04 adds the bounded public Event/Offering query surface, its
  contract, error behavior, and HTTP-edge safeguards.
