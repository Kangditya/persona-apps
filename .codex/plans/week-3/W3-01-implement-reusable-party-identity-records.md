# Task: W3-01 Implement Reusable Party Identity Records

## Status

Completed and verified on 2026-08-21. The executed task is archived at
`.codex/archive/2026-08-21-TASK-w3-01-implement-reusable-party-identity-records.md`;
no commit or push was requested.

## Tracker

- Week: Week 3
- Epic: Common Purchase
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Implement reusable Party identity records.
- Dependencies: W1-03 and W1-06.
- Acceptance summary: a Party can be created once and referenced by role-bearing
  commerce records without copying identity data.

## Objective

Implement the smallest reusable Party domain and PostgreSQL persistence slice
needed by Common Purchase. A Party is the stable identity record; purchaser,
payer, intended participant, sponsor, applicant, recipient, and Sohibul Qurban
remain roles that reference a Party rather than identity tables of their own.

## Source of Truth

- `docs/PRD.md`, especially Party Management and the actor distinctions;
- `docs/PRODUCT_MAP.md`, Identity ownership, and the Week 3 roadmap;
- `docs/ARCHITECTURE.md`, Identity bounded context and modular-monolith rules;
- `docs/DECISIONS.md`, especially ADR-013, ADR-014, ADR-038, and ADR-042;
- `docs/CONVENTIONS.md`;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- `docs/database/ERD.md` and `docs/database/MIGRATION_PLAN.md`;
- migrations `0001_qurban_foundation` and `0005_mvp_commerce_safety`;
- `.codex/CURRENT_STATE.md` and the current Go source.

The tracker `Scope` tab explicitly defers automatic identity deduplication.
That constraint is authoritative for this plan.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Before BUILD, create a new active `.codex/TASK.md` from this approved plan.
Do not replace or archive the current unrelated active task while its own
verification remains incomplete. Do not commit or push unless requested.

## Existing State

- Migration `0001` already creates `parties` with UUID identity, `PERSON` or
  `ORGANIZATION` type, display name, optional email and phone, and timestamps.
- Email and phone indexes are intentionally non-unique. The database does not
  establish contact data as a global identity key.
- `purchases`, `purchase_participants`, and later commerce records already
  reference `parties.id`; no Party-role join table is required for Week 3.
- The Go API has no Identity/Party domain or repository yet.
- The Storefront OpenAPI checkout request currently carries purchaser and payer
  contact values rather than Party IDs. W3-02 owns the explicit role-linking
  semantics; W3-01 must not guess them.

## Scope

### In Scope

- Add a focused `identity` package containing the Party domain type,
  validation, and PostgreSQL persistence needed by purchasing.
- Support the two already-accepted Party types without introducing more types.
- Create a Party with a server-generated UUID and read an existing Party by
  UUID within a caller-owned transaction.
- Allow commerce code to reference the same Party UUID from multiple role
  columns when the caller has an explicit approved relationship.
- Translate PostgreSQL failures to stable domain/application errors without
  exposing SQL or driver details.
- Add focused domain and repository tests, including disposable PostgreSQL
  coverage for persisted values and transaction rollback.

### Out of Scope

- Automatic, fuzzy, or global deduplication; merge workflows; contact matching;
  household/group identity; identity confidence scores; or a master-data UI.
- Treating equal display names, emails, or phone numbers as proof that two
  role inputs are the same person.
- Unique email/phone constraints or new normalized-contact columns.
- Public or Operations Party CRUD endpoints.
- Authentication identities, OIDC subjects, operator administration, or
  customer login accounts.
- Generic repositories, factories, registries, event buses, or interfaces with
  only one implementation.
- Party deletion. Referenced identity history must not be made disappear.

## Target State and Invariants

- A Party row owns identity data once; commerce rows store only the accepted
  foreign key plus their required historical snapshots.
- A Party UUID is stable and may be referenced by more than one role.
- The repository never searches for or merges a Party by display name, email,
  or phone as a side effect of creation.
- Optional contact fields remain optional and are persisted as actual nulls
  when absent; empty strings are not a second representation of absence.
- Domain validation enforces only artifact-backed database/contract bounds.
  It does not invent locale-specific phone or email semantics.
- Repository methods accept the existing transaction-compatible DB handle so
  W3-06 can create/reuse Parties atomically with a Purchase.

## Implementation Plan

1. Reconfirm the current migration constraints and all Party foreign-key users
   immediately before BUILD; do not add a migration unless the verified schema
   differs from the reviewed artifacts.
2. Add a small Party domain model and constructor/validation path for UUID,
   accepted type, display name, optional contact values, and timestamps.
3. Add a PostgreSQL repository with only the operations actually required by
   W3-02/W3-06: create and get-by-ID. Add another query only when an approved
   downstream plan proves it is necessary.
4. Keep transaction ownership with the purchasing command; do not begin or
   commit transactions inside the Party repository.
5. Add focused tests for accepted Party types, invalid values, null contact
   persistence, exact round trips, missing IDs, rollback, and reuse of one ID
   across role-bearing fixture records.
6. Run the applicable Go and repository checks and finish with a review that
   separates implemented, verified, assumed, and deferred behavior.

## Planned File Changes

Expected minimum:

- `apps/api/internal/identity/domain.go`;
- `apps/api/internal/identity/postgres.go`;
- focused `_test.go` files in the same package.

No OpenAPI, router, frontend, or migration file is planned for W3-01. If BUILD
finds such a change necessary, return to PLAN and document the evidence first.

## API and Database Impact

- API: none in this task.
- Database: reuse `parties`; no schema change is currently justified.
- Data migration/backfill: none.
- Audit/outbox/idempotency: none for this internal persistence slice; the
  owning W3-06 Purchase command composes those applicable effects.

## Acceptance Criteria

- A valid `PERSON` or `ORGANIZATION` Party can be created and read by UUID.
- Optional contact fields round-trip without empty/null ambiguity.
- One Party ID can be referenced by multiple commerce-role columns without
  duplicating the Party row or identity fields.
- Creation never performs implicit contact/name matching or merging.
- No new Party table, role table, framework, dependency, or migration is added
  without new reviewed evidence.
- Focused tests prove validation, persistence, missing-row behavior, rollback,
  and reference reuse.

## Verification

Run from `apps/api`:

```bash
go test ./internal/identity/...
go vet ./...
go test ./...
go build ./...
```

Then run from the repository root:

```bash
make validate
docker compose -f infrastructure/compose.yaml config
```

Use the repository's disposable PostgreSQL workflow for repository tests.
Report unrelated pre-existing validation failures separately; do not claim
they were caused or fixed by this task.

## Decisions, Assumptions, and Deferred Work

- Accepted: Party is the reusable identity; commerce concepts are roles.
- Accepted: automatic identity deduplication is deferred.
- Assumption constrained by current schema: UUID is the only authoritative
  Party reference in Week 3.
- Deferred to W3-02: how a checkout request explicitly states that two roles
  refer to the same Party.
- Deferred to later approved work: merge/dedup operations and customer-facing
  Party management.

## Final Review Requirements

The final report must list the exact files changed, schema impact, focused and
full checks run, failures or skipped checks, and remaining identity risks. Mark
the behavior as complete only when the acceptance criteria are verified; then
archive the active task using the repository's canonical execution workflow.
