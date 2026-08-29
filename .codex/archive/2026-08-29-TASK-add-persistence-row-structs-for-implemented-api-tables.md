# Task: Add Persistence Row Structs for Implemented API Tables
## Executed

## Objective

Add the minimum module-owned persistence row structs and explicit row-to-domain
mappings needed by the currently implemented `apps/api` vertical slices.
Keep PostgreSQL SQL migrations and handwritten SQL as the source of truth.

This task does not adopt GORM. It must not add an ORM, a query generator, or
placeholder model shells for domains that do not have runtime modules yet.

## Required approval gate

Do not begin the BUILD phase until this task plan has been reviewed and
explicitly approved.

## Source of truth

- `docs/PRD.md`
- `docs/PRODUCT_MAP.md`
- `docs/ARCHITECTURE.md`
- `docs/DECISIONS.md`
- `docs/CONVENTIONS.md`
- `.codex/CURRENT_STATE.md`
- `docs/database/ERD.md`
- `docs/database/MIGRATION_PLAN.md`

## In scope

### Active application persistence

Formalize or reuse small row structs inside the owning persistence packages
for the tables currently used by implemented business flows:

- `parties`
- `qurban_events`
- `offerings`
- `purchases`
- `purchase_participants`
- `quota_reservations`

The existing `offeringRow` is reused if it already covers the required mapping;
do not create a duplicate type merely to make filenames symmetrical.

### Active platform persistence

Inspect the existing authentication and platform writers and add a row type
only where it removes repeated table-shaped scanning or mapping without
changing behavior:

- `operator_users`
- `operator_sessions`

Keep the existing typed operation inputs for `audit_log`, `outbox_events`, and
`idempotency_records` unless a concrete read/mapping need is found. Do not add
speculative read models for write-only paths.

### Mapping and safety rules

- Keep persistence rows in infrastructure/platform packages; do not expose
  table rows through HTTP contracts or domain APIs.
- Map rows explicitly to domain structs.
- Represent nullable SQL columns with `sql.Null*` values or equivalent local
  persistence types, then convert them to domain pointers/values at the
  boundary.
- Preserve UTC normalization, binary/hash handling, JSON handling, exact
  integer money values, and existing status conversions.
- Reuse current SQL statements, transactions, row locks, version checks,
  idempotency, audit, outbox, and concurrency behavior.
- Keep the change local to the owning module or platform package and avoid a
  generic base model, reflection mapper, or shared `models` package.

### Tests and verification

- Add the smallest focused mapping tests needed for nullable fields, timestamps,
  status conversion, JSON/binary values, and existing error behavior.
- Preserve and run the existing repository and PostgreSQL integration tests.
- Verify that migrations remain explicit SQL migrations and API startup remains
  migration-free.

## Out of scope

- Adding `gorm.io/gorm`, a GORM PostgreSQL driver, or `AutoMigrate`.
- Rewriting the current repositories from `database/sql` to an ORM.
- Creating model structs for every schema table before its vertical slice exists.
- Implementing deferred Saving, Giveaway, Payment, Ledger, Sohibul Qurban,
  Livestock, Allocation, Slaughter, Distribution, or Event Location modules.
- Adding schema migrations, changing OpenAPI contracts, generating API clients,
  or changing frontend behavior.
- Adding soft-delete fields, speculative tenancy fields, or generic CRUD
  scaffolding.

## Execution plan

### DISCOVER

- Record the pre-existing worktree state and avoid modifying unrelated changes.
- Trace each active repository, authentication query, and platform writer to its
  callers.
- Inventory existing domain structs, row helpers, scan functions, nullable
  columns, and table ownership.
- Confirm no GORM dependency, tags, or migration calls are introduced by the
  current worktree.

### PLAN

- Confirm the final row inventory and the smallest file set before editing.
- Decide per table whether to reuse an existing row type, formalize a local
  row type, or keep an existing typed operation input because no row model is
  needed.
- Record any necessary plan variance before touching an unexpected file.

### BUILD

- Add or refine only the approved local persistence row structs.
- Add explicit row-to-domain and domain-to-row/argument mapping where needed.
- Keep SQL text, transaction ownership, and public/domain types stable.
- Add focused tests for the mapping branches that can regress.

### VERIFY

Run the applicable checks, including:

```text
make validate
cd apps/api && go vet ./... && go test ./... && go build ./...
docker compose -f infrastructure/compose.yaml config
```

Also confirm with repository search that:

- no GORM dependency, import, tag, or `AutoMigrate` call exists;
- no deferred-domain model shells were added;
- no migration or OpenAPI contract changed;
- existing uncommitted work outside this task is unchanged.

### REVIEW

Report separately what is implemented, verified, assumed, and deferred. Note
any remaining risks, especially whether a future vertical slice needs a new
table row or repository boundary. Archive the completed active task only after
acceptance criteria and verification pass.

## Acceptance criteria

- The API remains on `database/sql` plus pgx and explicit SQL migrations.
- Active persistence rows are module-owned, minimal, and explicitly mapped.
- Domain and transport contracts do not expose database rows.
- Existing Event, Offering, Party, Common Purchase, quota, auth, audit,
  outbox, idempotency, and concurrency behavior remains intact.
- Deferred schema-only domains receive no placeholder model layer.
- Focused tests and required repository verification pass.
- The final review distinguishes implemented, verified, assumed, and deferred
  behavior before this task is archived.

## Execution Record

### Implemented

- Added local persistence rows and explicit mappings for Event, Party, Purchase,
  Purchase detail, Purchase participant, Quota Reservation, operator user, and
  operator session data.
- Reused the existing Offering row and its mapping.
- Added focused row-mapping tests for nullable values, UTC timestamps, status
  conversion, and required reservation expiry.
- Kept `database/sql`, pgx, handwritten SQL, explicit migrations, transactions,
  locks, idempotency, audit, outbox, and API contracts unchanged.
- Added no GORM dependency, tag, `AutoMigrate` call, deferred-domain model, or
  schema migration.

### Verified

- `make validate` passed.
- `go vet ./...`, `go test ./...`, and `go build ./...` passed.
- Explicit formatter coverage for all changed Go files passed.
- Full Go tests passed against a clean disposable PostgreSQL database with all
  migrations applied, including repository and application integration tests.
- Compose configuration passed.
- Repository search found no GORM dependency, import, tag, or `AutoMigrate`
  call, and no migration or OpenAPI contract change was made.
- The disposable PostgreSQL database was removed after verification.

### Assumed

- Existing typed operation inputs remain the smallest correct representation
  for write-only audit, outbox, and idempotency paths.
- Future persistence rows will be introduced with their owning vertical slice,
  not as a prebuilt mirror of the complete schema.

### Deferred and remaining risk

- The pre-existing shared `persona_apps` database remains at migration 0005 in
  a dirty state because its existing data violates the one-active-event unique
  index. No repair, reset, or data change was performed; recovery requires an
  operator-reviewed database decision.
- PostgreSQL integration coverage is verified on the clean disposable database,
  not the dirty shared database.
- Saving, Giveaway, Payment, Ledger, Sohibul Qurban, Livestock, Allocation,
  Slaughter, Distribution, and Event Location model layers remain deferred until
  their vertical slices are implemented.
