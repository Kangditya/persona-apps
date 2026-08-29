# Database Design Review

> Historical review snapshot. This file records the database-design review at
> the time migrations 0001–0006 were introduced. It is not current product,
> implementation, or delivery status. Use `ERD.md`, `MIGRATION_PLAN.md`,
> `.codex/CURRENT_STATE.md`, ADR-051 through ADR-055, and
> `MVP-DELIVERY-ROADMAP.md` for current boundaries.

## Outcome

The repository-grounded database design is implemented as documentation,
standalone PostgreSQL migration scripts, and an explicit Go lifecycle command.
No migration was applied to staging or production.

## Required report

| Item                                           | Result                                                                                                                                                                                                                                                                                                 |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Database engine / dialect discovered           | PostgreSQL; Compose uses `postgres:18-alpine`, and Go uses `pgx/v5` with `database/sql`.                                                                                                                                                                                                               |
| Migration framework discovered                 | `golang-migrate/migrate/v4`, selected because it consumes paired SQL files and supports PostgreSQL locking/version state.                                                                                                                                                                              |
| Existing schema reused                         | None. No business tables, SQL schema, ORM, query models, seeds, or migration history exist.                                                                                                                                                                                                            |
| New tables proposed                            | 33 tables across foundation, commerce/funding, operations, platform integrity, and Phase 1 safety. See `ERD.md`.                                                                                                                                                                                       |
| Existing tables changed                        | qurban_events, offerings, purchases, payment_records, and audit_log receive additive 0005 changes; 0006 adds Event/Offering versions, exact-integer bounds, and a used-reservation aggregation index.                                                                                                  |
| Deferred tables / relationships at review time | Contacts, roles/permissions administration, offering composition/cart, saving policy, giveaway selection, attendance, distribution entitlement/beneficiaries, notifications, projections, and tenancy. ADR-051–ADR-055 now fix the Event-day targets, but additive implementation remains future work. |
| Migration files created                        | Twelve files: six `.up.sql` and six `.down.sql` files under `apps/api/migrations/`; 0006 adds Event/Offering concurrency and integer-safety hardening.                                                                                                                                                 |
| Backfills required                             | None; no existing schema/data.                                                                                                                                                                                                                                                                         |
| Indexes added                                  | Event/status/channel queues, public references, role/source lookups, history chronology, financial reconciliation, quota/session queues, used-reservation aggregation, livestock/allocation/queue operations, audit targets, unpublished outbox, and idempotency expiry.                               |
| Destructive operations                         | Only rollback scripts use `DROP TABLE`; up migrations are additive. No `CASCADE` is used.                                                                                                                                                                                                              |
| Rollback coverage                              | Complete per migration unit; apply down scripts in reverse order.                                                                                                                                                                                                                                      |
| API contract impact                            | None. Both OpenAPI contracts remain placeholders.                                                                                                                                                                                                                                                      |
| Dependency impact                              | Added `golang-migrate/migrate/v4` and test-only `go-sqlmock`; frontend dependencies unchanged.                                                                                                                                                                                                         |

## Implemented

- `docs/database/ERD.md` documents the proposed normalized model, Mermaid
  relationships, current/target/deferred classifications, conventions, and
  design rationale.
- `docs/database/MIGRATION_PLAN.md` documents dependency order, migration
  inventory, rollback, compatibility, indexes, risks, and unresolved decisions.
- `0001_qurban_foundation` defines parties, operator references, annual events,
  event locations, and offerings.
- `0002_commerce_and_funding` defines all three channel source workflows, one
  canonical purchase, participant activation, payment verification, and an
  append-only financial ledger.
- `0003_operations_and_platform` defines livestock lifecycle/history,
  allocation, slaughter scheduling/execution, minimal distribution status,
  audit, outbox, and idempotency records.

* `0004_schema_seeds` creates independent seed execution history.
* `0005_mvp_commerce_safety` adds the accepted Event, Offering, intended
* participant, quota, evidence, Purchase-token, session, and audit safeguards.
* `0006_add_event_offering_versions_and_bounds` adds Event/Offering optimistic
  versions, exact JSON-safe integer bounds, and the partial aggregation index
  for reservations in `RESERVED` or `CONSUMED` state.

- `apps/api/cmd/db` provides migration validation/status/version/up/bounded
  down/create, seed list/status/run, and setup commands.
- `internal/platform/database/seeder` provides ordered reference/development groups,
  per-seed transactions, idempotent history, and environment guards.

## Verified

- Repository discovery confirmed PostgreSQL and the absence of a prior
  migration framework or existing business schema.
- Focused Go tests passed in `apps/api` (17 tests across 9 packages, including
  migration validation, CLI parsing, seed ordering, commit/rollback, history,
  and environment guards).

* The migration source validates six paired files.

- The first `make validate` format check failed only because the new Markdown
  files were unformatted; the files were then formatted and the later full
  validation passed all preceding format, lint, typecheck, test, and build
  stages.
- The scripts are numbered in FK dependency order and their down scripts are
  reverse-ordered.
- Every scripted table appears in the ERD and every ERD scripted table appears
  in a migration inventory.
- Foreign-key target tables are created before dependent tables.
- Money uses positive integer minor units plus explicit direction; no floating
  point money fields are present.
- Event-scoped tables carry `event_id`; critical cross-event references use
  composite foreign keys.
- Unresolved product decisions remain out of the schema.

* Disposable PostgreSQL execution is recorded after the W1-03 verification run.

## Historical verification blockers

- `make validate` did not complete because its final Compose step could not
  start: `/bin/sh: 1: docker: not found`.
- `docker compose -f infrastructure/compose.yaml config` is therefore
  unrun/unverified for the same missing Docker executable.
- `psql --version` could not run because the PostgreSQL client is not installed.
- The runner's filesystem validation and unit tests passed; SQL syntax and
  migration up/down behavior remain blocked on a disposable PostgreSQL
  service.

## Disposable database verification

- Started an isolated PostgreSQL 18 Compose project on port 55433 with a new
  named volume.
- Applied migrations 0001 through 0005; version is 5 and not dirty.
- Ran the empty reference seed group twice without duplicate effects.
- Rolled back only 0005, reapplied it, and confirmed version 5 again.
- Inspected live constraints, foreign keys, partial indexes, evidence checks,
  reservation guards, token hashes, and session hashes against the ERD.
- No migration was applied to staging or production.

### W2-03 follow-up verification (2026-08-13)

- On a separate disposable local PostgreSQL 18 database, migration source
  validation found six pairs and the database reported version `6`, not dirty.
- Migration `0006` was rolled back one bounded step to version `5` and
  reapplied to version `6`, not dirty. No staging or production target was
  touched.
- Focused Event/Offering repository integration tests and the full API test
  suite passed with `TEST_DATABASE_URL`; `go vet ./...` and `go build ./...`
  also passed.
- Representative `EXPLAIN` output retained the existing Event-year and
  Offering event/code unique indexes for deterministic lists. Before 0006,
  used-reservation aggregation filtered `CONSUMED` rows after an existing
  active-reservation index; after 0006 it used
  `idx_quota_reservations_used_event_offering`. That is the only new index.

## Assumed

- `gen_random_uuid()` is available in the verified PostgreSQL 18 runtime.
- The application will generate UTC timestamps and update `updated_at` on
  aggregate writes; no trigger convention exists yet.
- `operator_users.external_subject` can reference a future authentication
  subject without committing to an authentication provider.
- The minimal distribution record was an operational handoff. ADR-054 now
  fixes the Full Event-Day entitlement/beneficiary/pickup/delivery target;
  additive implementation remains future work.
- Allocation capacity totals will be protected by a transaction and row lock or
  equivalent application command; SQL alone cannot sum capacity across rows.

## Deferred

- Disposable PostgreSQL execution of the runner, migrations, and seed history.

* OIDC runtime, role/permission administration, and event/location scopes.

- Offering variants, package/share composition, multi-offering checkout, and
  quota command implementation.
- Saving price locks, installment schedules, transfers/refunds policy, and
  reminders.
- Giveaway eligibility criteria and recipient-selection workflow.
- Participant documents and object-storage references.
- Runtime attendance/proxy/queue behavior and additive storage required by
  ADR-053.
- Runtime Distribution beneficiaries, portions/entitlements, pickup/delivery,
  and proof required by ADR-054; route optimization remains deferred.
- Notification attempts, payment-provider inboxes, and report/dashboard
  projections.
- Generic multitenancy or `organisation_id`.

## Remaining risks

1. Verification used one empty PostgreSQL 18 database. Staging and production
   compatibility remain unverified until deployment configuration exists.
2. Offering and allocation rules are intentionally broad, so capacity and
   price-lock constraints will need hardening after product decisions.
3. Authentication storage is present, but OIDC, session, CSRF, and permission
   runtime behavior remain unimplemented.
4. Distribution remains schema-only/minimal and must not be treated as the
   ADR-054 runtime entitlement model.
5. Rollback remains destructive and should be exercised only against a
   disposable/local database until a deployment backup and recovery policy is
   established.

## Recommended next task

The migration/seed workflow and the Event/Offering plus Week 3 Common Purchase
backend slices were completed after this historical review. The current next
task is:

```text
W3-07 Storefront Direct Checkout Form and Confirmation
```

Revalidate the active tracker row and current source before execution.
