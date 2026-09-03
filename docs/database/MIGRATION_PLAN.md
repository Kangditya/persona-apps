# Qurban Database Migration Plan

## Repository discovery

The repository establishes PostgreSQL as the transactional database through
`infrastructure/compose.yaml`, `apps/api/go.mod`, and the PostgreSQL adapter.
The Go runner under `apps/api/cmd/db` consumes the existing migration pairs
through `golang-migrate/migrate/v4`. The runner owns migration history through
the library's `schema_migrations` table and uses PostgreSQL advisory locking.
Seed history is separate in `schema_seeds`. No ORM or query generator is used.

The baseline uses standalone PostgreSQL SQL pairs with the filename convention:

```text
NNNN_<scope>.up.sql
NNNN_<scope>.down.sql
```

The files are not executed during API startup. They are executed explicitly by
the database CLI and Make targets.

## Schema baseline

- PostgreSQL `public` schema.
- UUID primary keys with `gen_random_uuid()` defaults.
- Lowercase plural `snake_case` names.
- `timestamptz` timestamps; the application writes UTC values.
- `bigint` minor units plus uppercase three-letter currency code for money.
- `numeric(10,3)` kilograms for livestock weight.
- Table-local text status checks rather than PostgreSQL enum types.
- No `deleted_at`; status/history preserves business history.
- Explicit foreign keys, with event-aware composite foreign keys for critical
  event-scoped references.
- No backfill because no existing business tables or data were discovered.

## Dependency order

```text
0001 foundation
  ├── operator_users
  ├── parties
  ├── qurban_events
  ├── event_locations
  └── offerings
        │
0002 commerce and funding
  ├── saving_accounts
  ├── giveaway_programs
  ├── giveaway_applications
  ├── giveaway_assignments
  ├── purchases
  ├── purchase and participant histories
  ├── payment_records and history
  └── financial_ledger_entries
        │
0003 operations and platform
  ├── livestock and histories
  ├── allocations and history
  ├── slaughter sessions/stations/records
  ├── distribution and history
  ├── audit_log
  ├── outbox_events
  └── idempotency_records
        │
0004 metadata
  └── schema_seeds
        │
0005 MVP commerce safety
  ├── purchase_participants
  ├── quota_reservations
  ├── operator_sessions
  └── safety constraints and indexes
        │
0006 Event and Offering versions and bounds
  ├── qurban_events
  ├── offerings
  └── used-reservation aggregation index
        │
0007 submitted Payment guard
  └── one SUBMITTED Payment per Purchase
```

## Migration inventory

| Sequence | Migration                                            | Type               | Tables affected                                                                                                                                                                                      | Purpose                                                                                                                                        | Backfill | Risk                                                                                               |
| -------: | ---------------------------------------------------- | ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | -------- | -------------------------------------------------------------------------------------------------- |
|     0001 | `0001_qurban_foundation.up.sql`                      | Schema-only        | `operator_users`, `parties`, `qurban_events`, `event_locations`, `offerings`                                                                                                                         | Establish identity references, annual event boundary, locations, and commercial offerings.                                                     | None.    | Low: all tables are new and additive.                                                              |
|     0002 | `0002_commerce_and_funding.up.sql`                   | Schema-only        | `saving_accounts`, `giveaway_programs`, `giveaway_applications`, `giveaway_assignments`, `purchases`, purchase/participant histories, `payment_records`, payment history, `financial_ledger_entries` | Model three pre-purchase/channel sources, one canonical purchase, participant activation, payment verification, and append-only money effects. | None.    | Medium: channel/source checks and ledger semantics are core invariants.                            |
|     0003 | `0003_operations_and_platform.up.sql`                | Schema-only        | livestock, inspection/location/status histories, allocations, allocation history, slaughter, distribution, audit, outbox, idempotency                                                                | Add operational lifecycle, capacity claims, event-day records, minimal distribution status, and platform integrity records.                    | None.    | Medium: allocation capacity still needs transactional application logic.                           |
|     0004 | `0004_schema_seeds.up.sql`                           | Metadata           | `schema_seeds`                                                                                                                                                                                       | Track deterministic seed execution separately from migration version state.                                                                    | None.    | Low: independent metadata table.                                                                   |
|     0005 | `0005_mvp_commerce_safety.up.sql`                    | Additive safety    | qurban event, offering, purchase, payment, audit, intended participant, quota reservation, and session storage                                                                                       | Enforce Phase 1 lifecycle, quota, evidence, Purchase-token, and session safety without rewriting history.                                      | None.    | Medium: constraint changes require disposable-DB verification.                                     |
|     0006 | `0006_add_event_offering_versions_and_bounds.up.sql` | Additive hardening | `qurban_events`, `offerings`, `quota_reservations`                                                                                                                                                   | Add optimistic versions, exact JSON-safe integer bounds, and the query-backed used-reservation aggregation index.                              | None.    | Medium: additive checks must be verified against existing rows and rollback drops version columns. |
|     0007 | `0007_add_submitted_payment_guard.up.sql`            | Additive safety    | `payment_records`                                                                                                                                                                                    | Prevent concurrent different intents from creating more than one Payment awaiting review for a Purchase.                                       | None.    | Medium: pre-existing duplicate submitted rows must be resolved before the index can be applied.    |

The matching `.down.sql` files are rollback scripts for each unit. Rollback is
destructive for the unit being reverted and must only be used when the owning
deployment has confirmed that its data is disposable or separately backed up.

## Step details

### 0001 — Qurban foundation

- Tables: `operator_users`, `parties`, `qurban_events`, `event_locations`, `offerings`.
- Columns: UUID identity, operator status, party kind/contact basics, event windows/quota, event locations, offering price/currency/capacity/status.
- Constraints: unique operator subject, event year, event location code, offering code; event status/type checks; non-negative quota/price and positive capacity; registration window ordering.
- Foreign keys: locations and offerings reference events.
- Indexes: event status, location type, offering event/status, party contact lookup.
- Data migration: none.
- Compatibility impact: no existing schema/API is changed; this enables future event/offering APIs.
- Rollback: `0001_qurban_foundation.down.sql` drops tables in reverse dependency order. It does not remove PostgreSQL UUID functionality.

### 0002 — Commerce and funding

- Tables: all channel source tables, `purchases`, status histories, participant records, payment records, and ledger entries.
- Columns: event-aware source references, role-specific party references, purchase pricing snapshots, lifecycle statuses, positive minor-unit amounts, ledger direction/type, versions, idempotency/provider references.
- Constraints: purchase channel/source exclusivity; exactly one payment/ledger context; exactly one giveaway applicant/nominee; one assignment per application and recipient per program; event-aware foreign keys; unique public references; positive amounts.
- Foreign keys: source aggregates reference foundation tables; canonical purchase references source aggregates; histories reference aggregates; ledger references financial context.
- Indexes: event/status/channel queues, party lookups, source partial uniqueness, history chronology, payment/provider lookup, ledger reconciliation paths.
- Data migration: none.
- Compatibility impact: none; no prior tables or fields exist.
- Rollback: `0002_commerce_and_funding.down.sql` removes children before sources. It is destructive and must follow 0003 rollback first.

### 0003 — Operations and platform

- Tables: livestock, inspections, status/location histories, allocations, slaughter, minimal distribution, audit, outbox, idempotency.
- Columns: current operational state plus immutable histories, version columns for contested aggregates, event-day queue records, audit JSON snapshots, outbox publication state, command replay records.
- Constraints: event-aware references, unique livestock codes/tags, one current location, one active participant allocation, one slaughter record per livestock/event, target requirements for allocation/distribution, status checks.
- Foreign keys: livestock and operations reference events/locations/purchases/participants; audit actor references point to known local operators/parties when available.
- Indexes: readiness/allocation/queue/distribution operational queries, audit target/time, unpublished outbox, idempotency expiry.
- Data migration: none.
- Compatibility impact: none; no prior tables or fields exist.
- Rollback: `0003_operations_and_platform.down.sql` drops platform/operations tables in reverse dependency order. It must run before 0002 rollback.

### 0004 — Seed metadata

- Table: `schema_seeds`.
- Columns: immutable seed name and UTC execution timestamp.
- Constraints: primary key on seed name; no business foreign keys.
- Data migration: none.
- Compatibility impact: adds independent lifecycle metadata used by the Go
  seeder runner.
- Rollback: `0004_schema_seeds.down.sql` drops only the seed history table.

### 0005 — MVP commerce safety

- Tables: adds purchase_participants, quota_reservations, and operator_sessions;
  alters qurban_events, offerings, purchases, payment_records, and audit_log.
- Constraints: one active Event; SUSPENDED event status; optional Offering quota;
  common-Purchase 32-byte access-token hashes; intended-participant sequence;
  event-aware Purchase and Offering reservation references; one active
  reservation; terminal reservation timestamps; evidence metadata type, size,
  and digest; hashed revocable sessions with permission snapshots.
- Indexes: active Event, Purchase-token lookup, active and expiring quota
  reservations, and active/expiring operator sessions.
- Data migration: none. The repository has no shared environment migration or
  business runtime, so there is no backfill.
- Compatibility impact: additive columns and tables; the Event status check
  expands. The migration preserves audit_log.reason and source_application,
  adding source and permission context for future writers.
- Rollback: drops only migration 0005 additions, then restores the original
  Event status check. It is destructive for 0005 tables and correctly refuses
  rollback while SUSPENDED Events remain.

### 0006 — Event and Offering versions and bounds

- Tables: alters `qurban_events` and `offerings`; adds an index on
  `quota_reservations` without changing reservation rows.
- Columns: positive `bigint version NOT NULL DEFAULT 1` on Events and
  Offerings.
- Constraints: positive and JSON-safe version values; JSON-safe optional Event
  and Offering participant quotas; JSON-safe Offering `price_minor`.
- Indexes: `idx_quota_reservations_used_event_offering` covers Event/Offering
  aggregation of `RESERVED` and `CONSUMED` participant units. Existing Event
  and Offering unique indexes already support their deterministic list order,
  so no duplicate list index is added.
- Data migration: none. The additive default supplies version `1`; local
  disposable-data preflight found no out-of-range values before the checks.
- Compatibility impact: public and Operations response contracts must treat
  versions as internal until the later Operations command API exposes them.
- Rollback: drops only the 0006 index, named checks, and version columns. It
  is destructive for version values and must only run against a disposable or
  separately recovered deployment.

### 0007 — Submitted Payment guard

- Table: adds one partial unique index to `payment_records`.
- Constraint: one Purchase may have at most one Payment with status
  `SUBMITTED`; rejected and other terminal attempts remain append-oriented.
- Data migration: none. Deployment preflight must detect existing duplicate
  submitted rows before applying the index.
- Compatibility impact: concurrent different idempotency keys now produce one
  current review attempt and one stable state conflict instead of parallel
  submitted evidence.
- Rollback: drops only the partial unique index and preserves every Payment row.

## Constraint and index classification

### Schema-only migrations

The first three migrations are schema-only. They create new tables, constraints,
and indexes atomically. Migration 0004 adds only seed metadata; the SQL does
not insert business data or seed lookup rows.

### Data migrations/backfills

None required. The repository contains no existing persisted business records.
If tables are introduced before this baseline is applied, a new expand/backfill
migration must be added rather than editing these files.

### Constraint-hardening migrations

Migrations 0006 and 0007 harden existing Event/Offering and Payment tables with
exact-integer bounds and one current submitted-Payment guard. Future hardening
may still be required after
real command flows establish policies for quota reservation, price locking,
giveaway selection, Event-day attendance, allocation, Distribution, projection
recovery, and authentication scope. ADR-051 through ADR-055 fix the Full
Event-Day target policies; implementation still requires measured query and
constraint verification.

### Index migrations

Migration 0006 adds one narrow partial index after a representative
availability aggregation plan showed the existing active-reservation index did
not cover `CONSUMED` reservations. Existing Event and Offering list indexes
were retained after their representative plans proved sufficient. The index
names and query justification are listed in `docs/database/REVIEW.md`.

### Cleanup migrations

None. No existing fields or tables are renamed or deleted.

## Rollback and compatibility

Apply up migrations in ascending order. Roll back in descending order:

```text
0006 down → 0005 down → 0004 down → 0003 down → 0002 down → 0001 down
```

The down scripts intentionally use `DROP TABLE` and are destructive. They do
not use `CASCADE`, which makes an unexpected dependency fail visibly. No
historical migration file is edited because none exists.

`golang-migrate` owns the `schema_migrations` version table and PostgreSQL
advisory lock. The down scripts remain destructive and must be invoked with an
explicit bounded step. Deployment automation runs the CLI as a separate job;
API startup does not run migrations.

## Unresolved decisions and schema impact

| Decision                           | Current treatment                                                                                                    | Required before                                 |
| ---------------------------------- | -------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| Offering/package/share composition | `offering_kind` remains descriptive text; no item/variant tables.                                                    | Offering implementation beyond the first slice. |
| Multiple offerings in one checkout | One purchase has one offering; no cart or purchase-item table.                                                       | Multi-offering checkout.                        |
| Saving price/target lock           | Saving stores target amount and optional offering, without lock history.                                             | Saving conversion implementation.               |
| Giveaway selection                 | Application/assignment status only; selection policy is not stored.                                                  | Giveaway workflow implementation.               |
| Personal slaughter/attendance      | ADR-053 fixes configurable `SELF`, `PROXY`, or `NONE`; current schema still lacks attendance/history.                | Planned additive Event-day migration.           |
| Distribution scope                 | ADR-054 fixes Sohibul entitlement plus beneficiary pickup/delivery/proof; current table remains minimal.             | Planned additive Distribution migration.        |
| Authentication/authorization       | Session and Purchase-token hashes are stored; runtime OIDC, permissions, scopes, and administration remain deferred. | W1-06 operations access implementation.         |
| Migration tooling                  | `golang-migrate/migrate/v4` consumes the existing SQL pairs; CLI and Make targets are explicit.                      | Dirty recovery commands and disposable-DB CI.   |

## Plan variance

The approved migration tooling plan is implemented. Migration `0004` adds only
seed metadata; the domain ERD and historical migrations remain unchanged.
Reference seeds remain empty because no production bootstrap data is justified
by the current schema. No Compose file or API startup path was changed.

Migration `0006` is an additive follow-up: it does not edit historical pairs,
adds no seed data, and derives its single new index from an observed
availability query plan.

## Planned additive Event-day migration sequence

No SQL file is created by this planning update. Active implementation tasks
must inspect the live schema and choose the next unused versions while
preserving this dependency order:

| Planned unit              | Scope                                                                                                                                              | Depends on                                             | Rollback boundary                                                                               |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ | ----------------------------------------------------------------------------------------------- |
| Event execution and teams | Event timezone, 3–4 execution days, field teams/memberships, shifts, assignments, handovers, readiness, attendance, incidents, histories, indexes. | Current Event, locations, operators, Sohibul Qurban.   | Drop only the new Event-day/team records after proving no dependent Livestock/Slaughter data.   |
| Livestock hardening       | Add only missing execution-day/location/version/history constraints needed by the implemented lifecycle.                                           | Event execution and teams; existing livestock schema.  | Remove only additive hardening; never drop historical livestock tables in a shared environment. |
| Allocation hardening      | Capacity/version/history/index changes proven by transactional allocation queries.                                                                 | Eligible Purchase/Sohibul, Livestock.                  | Block rollback while new Allocation state/history depends on the additions.                     |
| Slaughter execution       | Session/day/team/shift/station linkage, attendance/check-in linkage, queue history, incident linkage, constraints, indexes.                        | Event execution, teams, Livestock, Allocation.         | Block rollback while execution records exist unless a separately verified recovery exists.      |
| Distribution completion   | Entitlements, beneficiaries, portions, pickup/delivery method, proof metadata, failure/reassignment/completion history, constraints, indexes.      | Slaughter completion and existing distribution schema. | Remove only additive records after backup/recovery proof.                                       |
| Projection checkpoints    | Durable projection cursors/checkpoints and worker backlog/dead-letter metadata where existing outbox fields are insufficient.                      | Authoritative domain/outbox events.                    | Projection storage is rebuildable, but rollback must not alter source domain records.           |

Every pair remains additive, bounded, separately reversible, and validated by
fresh-database up/down/up plus compatibility tests. Historical migrations 0001
through 0006 remain unchanged.
