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
```

## Migration inventory

| Sequence | Migration                             | Type        | Tables affected                                                                                                                                                                                      | Purpose                                                                                                                                        | Backfill | Risk                                                                     |
| -------: | ------------------------------------- | ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------------------------------------------------------ |
|     0001 | `0001_qurban_foundation.up.sql`       | Schema-only | `operator_users`, `parties`, `qurban_events`, `event_locations`, `offerings`                                                                                                                         | Establish identity references, annual event boundary, locations, and commercial offerings.                                                     | None.    | Low: all tables are new and additive.                                    |
|     0002 | `0002_commerce_and_funding.up.sql`    | Schema-only | `saving_accounts`, `giveaway_programs`, `giveaway_applications`, `giveaway_assignments`, `purchases`, purchase/participant histories, `payment_records`, payment history, `financial_ledger_entries` | Model three pre-purchase/channel sources, one canonical purchase, participant activation, payment verification, and append-only money effects. | None.    | Medium: channel/source checks and ledger semantics are core invariants.  |
|     0003 | `0003_operations_and_platform.up.sql` | Schema-only | livestock, inspection/location/status histories, allocations, allocation history, slaughter, distribution, audit, outbox, idempotency                                                                | Add operational lifecycle, capacity claims, event-day records, minimal distribution status, and platform integrity records.                    | None.    | Medium: allocation capacity still needs transactional application logic. |
|     0004 | `0004_schema_seeds.up.sql`            | Metadata    | `schema_seeds`                                                                                                                                                                                       | Track deterministic seed execution separately from migration version state.                                                                    | None.    | Low: independent metadata table.                                         |

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

None in this task. Stable invariants are created with their tables. Future
hardening may be required after real command flows establish policies for quota
reservation, price locking, giveaway selection, distribution entitlement, and
authentication scope.

### Index migrations

No separate index migration is needed because the tables are new and indexes
are part of each creation unit. The index names and query justification are
listed in `docs/database/REVIEW.md`.

### Cleanup migrations

None. No existing fields or tables are renamed or deleted.

## Rollback and compatibility

Apply up migrations in ascending order. Roll back in descending order:

```text
0004 down → 0003 down → 0002 down → 0001 down
```

The down scripts intentionally use `DROP TABLE` and are destructive. They do
not use `CASCADE`, which makes an unexpected dependency fail visibly. No
historical migration file is edited because none exists.

`golang-migrate` owns the `schema_migrations` version table and PostgreSQL
advisory lock. The down scripts remain destructive and must be invoked with an
explicit bounded step. Deployment automation runs the CLI as a separate job;
API startup does not run migrations.

## Unresolved decisions and schema impact

| Decision                           | Current treatment                                                                               | Required before                                 |
| ---------------------------------- | ----------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| Offering/package/share composition | `offering_kind` remains descriptive text; no item/variant tables.                               | Offering implementation beyond the first slice. |
| Multiple offerings in one checkout | One purchase has one offering; no cart or purchase-item table.                                  | Multi-offering checkout.                        |
| Saving price/target lock           | Saving stores target amount and optional offering, without lock history.                        | Saving conversion implementation.               |
| Giveaway selection                 | Application/assignment status only; selection policy is not stored.                             | Giveaway workflow implementation.               |
| Personal slaughter/attendance      | Generic slaughter records only; no participant queue/attendance model.                          | Field workflow confirmation.                    |
| Distribution scope                 | Minimal purchase/participant status record; no beneficiaries, portions, proof, or route model.  | Distribution requirements.                      |
| Authentication/authorization       | Minimal operator reference only; roles, permissions, sessions, and scopes deferred.             | Auth ADR and operations access implementation.  |
| Migration tooling                  | `golang-migrate/migrate/v4` consumes the existing SQL pairs; CLI and Make targets are explicit. | Dirty recovery commands and disposable-DB CI.   |

## Plan variance

The approved migration tooling plan is implemented. Migration `0004` adds only
seed metadata; the domain ERD and historical migrations remain unchanged.
Reference seeds remain empty because no production bootstrap data is justified
by the current schema. No Compose file or API startup path was changed.
