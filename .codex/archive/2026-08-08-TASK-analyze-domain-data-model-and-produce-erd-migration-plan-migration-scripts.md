# Task: Analyze Domain Data Model and Produce ERD + Migration Plan + Migration Scripts

## Executed

## Status

```text
Pending
```

## Objective

Analyze the repository and derive the database model required by the current Qurban product scope.

The output of this task is **database design artifacts**, not application feature implementation:

1. a repository-grounded ERD;
2. a dependency-ordered migration plan;
3. migration scripts following the repository's existing database dialect and migration conventions;
4. a review report explaining assumptions, unresolved gaps, and risks.

Do not apply migrations to any database as part of this task.

---

## Required Workflow

```text
DISCOVER → MODEL → PLAN → SCRIPT → VERIFY → REVIEW
```

The analyzer must complete each phase in order.

### 1. DISCOVER

Inspect the repository before designing any schema.

At minimum inspect:

```text
docs/PRD.md
docs/PRODUCT_MAP.md
docs/ARCHITECTURE.md
docs/DECISIONS.md
.codex/CURRENT_STATE.md
.codex/AGENTS.md
apps/api
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
```

Also locate and inspect, where present:

* existing database migrations;
* database bootstrap/schema files;
* ORM/query models;
* repositories and persistence adapters;
* domain entities/value objects;
* API request/response contracts that imply persisted state;
* seeds/fixtures;
* indexes, foreign keys, enums, lookup tables, and soft-delete conventions;
* timestamp, UUID, monetary, status, and audit-field conventions.

Determine from repository evidence:

```text
database engine / SQL dialect
migration tool and filename convention
schema naming convention
primary-key strategy
UUID strategy
foreign-key convention
created_at / updated_at / deleted_at convention
status / enum representation
money and quantity representation
transaction boundaries
existing tables that must be preserved
```

Do not assume PostgreSQL, MySQL, SQLite, or another engine before verifying it from the repository.

---

## Source-of-Truth Rules

Use repository evidence in this priority order:

1. current product requirements and decisions;
2. current architecture documentation;
3. existing backend/domain code;
4. existing database schema and migrations;
5. current API contracts;
6. explicit development fixtures or mocks.

When sources disagree:

* record the conflict;
* identify which source appears authoritative;
* do not silently choose a business rule;
* mark unresolved decisions as `TBD` when repository evidence is insufficient.

Do not introduce speculative multitenancy, microservices, accounting, inventory, payment-provider, or reporting structures unless they are supported by current product scope.

---

## Domain Analysis Scope

Derive persistence requirements from the product currently represented by the repository, including only domains supported by evidence.

Expected areas to investigate include, where applicable:

```text
qurban events
purchase channels
common purchasing
saving / installment purchasing
giveaway purchasing
purchasers
Sohibul Qurban / participants
offerings / packages
livestock
livestock allocation
payments / payment verification
purchase tracking
slaughter / operational processing
distribution
participant documents
operational users / authorization references
```

These names are investigation targets, not mandatory tables.

Do not assume that:

* purchaser and Sohibul Qurban are always the same person;
* one purchase always represents one participant;
* one animal always belongs to exactly one participant;
* installment, common, and giveaway purchasing have identical lifecycle rules;
* payment state is equivalent to purchase state;
* frontend display models are canonical database models.

Derive cardinality from product requirements and backend behavior rather than from UI layout.

---

## ERD Requirements

Create a database ERD artifact at:

```text
docs/database/ERD.md
```

If the repository already has an established database-documentation location, use that location instead and record the variance.

The ERD document must contain:

### A. Model summary

For every proposed or existing persisted entity, document:

```text
table name
purpose
primary key
business identifier / UUID if applicable
important columns
foreign keys
unique constraints
important indexes
nullable relationships
lifecycle/status fields
audit fields
soft-delete behavior
```

### B. Relationship model

Represent the ERD using Mermaid `erDiagram` unless the repository already uses another textual ERD format.

The diagram must show:

* one-to-one relationships;
* one-to-many relationships;
* many-to-many relationships through explicit junction tables;
* optional relationships;
* ownership boundaries;
* important lookup/reference tables.

### C. Current vs target state

Clearly classify tables as:

```text
EXISTING
CHANGE
NEW
DEFERRED
```

Do not present a proposed table as existing.

### D. Design rationale

For non-trivial modeling decisions, explain why the model was chosen, especially for:

* purchaser vs participant separation;
* purchase channel modeling;
* installment/payment modeling;
* livestock ownership/allocation;
* event-scoped data;
* lifecycle/status representation;
* mutable operational state vs immutable history;
* document references;
* auditability.

---

## Migration Plan Requirements

Create:

```text
docs/database/MIGRATION_PLAN.md
```

The plan must be derived from the ERD and existing schema.

For each migration step document:

```text
sequence
migration name
purpose
tables affected
columns added/changed
constraints
indexes
foreign keys
data backfill requirement
compatibility impact
rollback strategy
risk level
```

Order migrations by dependency so that they can be applied safely.

The plan must distinguish:

```text
schema-only migration
data migration / backfill
constraint-hardening migration
index migration
cleanup migration
```

When a destructive change is required, prefer an expand-and-contract sequence unless repository constraints make that impossible.

Do not delete or rename a persisted field without documenting compatibility impact and rollback behavior.

---

## Migration Script Requirements

Generate actual migration scripts using the migration framework and SQL dialect already established by the repository.

### Location

Use the repository's existing migration directory and naming convention.

If no migration framework exists:

* do not silently introduce one;
* document that fact in the migration plan;
* create standalone SQL scripts only if that is compatible with the existing architecture;
* record the proposed migration-tool decision as `TBD` if a tool choice requires product/engineering approval.

### Script rules

Migration scripts must:

* match the verified SQL dialect;
* create foreign keys only after their referenced structures exist;
* include required unique constraints;
* include indexes justified by known lookup/query patterns;
* preserve existing data;
* avoid irreversible destructive changes where possible;
* avoid production-specific IDs or credentials;
* avoid fake seed data unless seeds are already part of the repository convention;
* use deterministic names for constraints and indexes when the repository does so;
* provide down/rollback migrations when supported by the established migration framework.

Do not run the migrations.

---

## Modeling Standards

Follow existing repository conventions first. Where no convention exists, document the proposed choice rather than silently standardizing the codebase.

Pay particular attention to:

### Identity

Determine whether tables use:

```text
BIGINT / sequence IDs
UUID primary keys
internal ID + public UUID
another repository-specific strategy
```

Do not create competing identity strategies without justification.

### Money

Do not use floating-point storage for monetary values.

Use the repository's established exact numeric representation and record currency semantics where applicable.

### Time

Identify the repository convention for:

```text
timestamps
time zones
event dates
business dates
created_at / updated_at
soft deletion
```

### Status and history

Distinguish between:

* current status stored on the aggregate;
* append-only status/history records;
* payment lifecycle;
* operational lifecycle.

Do not overload one status field to represent unrelated lifecycles.

### Referential integrity

Prefer explicit foreign keys when consistent with the repository.

Document any relationship intentionally enforced only at application level and explain why.

---

## Analysis Questions the Task Must Resolve

The final artifacts must answer, from repository evidence where possible:

1. What are the core aggregates and persisted entities?
2. Which entities already exist in the database or backend?
3. Which tables are missing for the current product scope?
4. How are Event, Purchase, Purchaser, Participant/Sohibul Qurban, Payment, Livestock, Allocation, and Distribution related?
5. How are the three purchasing channels represented without duplicating shared purchase data?
6. What data must be event-scoped?
7. Which relationships are optional versus mandatory?
8. Which records require immutable/history tracking?
9. Which fields require unique constraints?
10. Which query patterns justify indexes?
11. Which migrations require data backfill?
12. Which decisions remain unresolved and therefore must not be encoded yet?

---

## Verification

Verification is static and repository-based. Do not connect to or mutate an external database.

Run the applicable checks available in the repository, for example:

```bash
make validate
make test
make lint
go test ./...
pnpm run lint
pnpm run typecheck
```

Only run commands that are relevant to files touched by this task.

For migrations, use the migration tool's dry-run, validation, parse, or status command when supported and when it does not apply changes.

Also verify manually:

1. every ERD table is backed by repository evidence or explicitly labeled proposed;
2. every foreign key has a valid referenced table/key;
3. cardinalities in the Mermaid ERD match the described constraints;
4. migration ordering respects FK dependencies;
5. indexes have a documented query/access justification;
6. migration scripts match the discovered SQL dialect;
7. no migration is executed;
8. no secrets or environment-specific values are introduced;
9. public/storefront and operations concerns do not force duplicated persistence models without a domain reason;
10. deferred product decisions remain deferred instead of being guessed.

If a verification command cannot run, record the exact command and blocker.

---

## Deliverables

The task is complete only when it produces:

1. repository/domain analysis summary;
2. `docs/database/ERD.md` or the repository-equivalent path;
3. `docs/database/MIGRATION_PLAN.md` or the repository-equivalent path;
4. migration scripts in the established migration directory;
5. a migration/script inventory showing execution order;
6. a list of schema assumptions and `TBD` decisions;
7. a final verification and review report.

The final report must contain:

```text
Database engine / dialect discovered
Migration framework discovered
Existing schema reused
New tables proposed
Existing tables changed
Deferred tables / relationships
Migration files created
Backfills required
Indexes added and rationale
Destructive operations, if any
Rollback coverage
Verification commands executed
Verification failures / blockers
Plan variances
Remaining risks
Recommended next task
```

---

## Constraints

* Analyze first; do not model directly from UI screenshots.
* Keep the schema aligned with current product scope.
* Do not invent unsupported business rules.
* Do not implement unrelated Go or frontend features.
* Do not run migrations against development, staging, or production databases.
* Do not commit or push changes.
* Do not replace existing migrations that may already have been applied.
* Prefer additive migrations over editing historical migration files.
* Preserve public Storefront and internal Operations API boundaries at the application layer while keeping canonical domain data normalized in persistence.
* Record uncertainty instead of encoding guesses into schema.
