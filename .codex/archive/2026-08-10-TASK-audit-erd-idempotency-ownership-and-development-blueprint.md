# Task: Audit ERD Idempotency Ownership and Development Blueprint

## Executed

## Status

Completed and verified on 2026-08-10.

## Objective

Deep-audit the proposed database ERD and its matching PostgreSQL migrations to
define which entities and processes need idempotency, which mechanism owns it,
and how future API and worker implementations must preserve retry safety.

## Constraints

- Treat the canonical product, architecture, decisions, current state, ERD,
  and migration SQL as source material.
- Do not add an `idempotency_key` column to every entity mechanically.
- Distinguish request replay, natural uniqueness, concurrency control,
  append-only effects, and asynchronous consumer deduplication.
- Keep this task documentation-only: do not change migrations, API contracts,
  runtime code, generated clients, dependencies, or infrastructure.
- Do not invent unresolved provider, retention, authentication, notification,
  or distribution policies.
- Do not commit or push.

## Planned Changes

| File | Action | Purpose |
| --- | --- | --- |
| `docs/DECISIONS.md` | Modify | Record command-scoped idempotency as the accepted cross-module direction. |
| `docs/database/ERD.md` | Modify | Add the complete entity and process idempotency audit for development. |
| `.codex/CURRENT_STATE.md` | Modify | Reconcile stale database-schema and migration-tooling statements. |
| `.codex/TASK.md` | Create, execute, archive | Preserve the approved workflow and final review. |

## Acceptance Criteria

1. Every business table in the ERD is classified by its idempotency ownership.
2. Retry-sensitive business processes identify their replay key, database
   guard, transaction boundary, concurrency control, audit behavior, and
   asynchronous behavior.
3. The artifacts clearly state that generic command idempotency belongs in
   `idempotency_records`, while only justified effect records keep direct
   deduplication keys.
4. Same-key replay, payload mismatch, concurrent duplicate, authorization,
   response caching, error handling, and retention behavior are documented.
5. Provider webhook and outbox-consumer gaps are marked deferred rather than
   silently invented.
6. No migration, runtime, contract, dependency, or infrastructure file changes.
7. Current-state documentation matches the implemented migration and database
   lifecycle artifacts.

## Verification

- Format the changed Markdown files.
- Run `make format-check`.
- Cross-check every ERD table against the entity audit.
- Confirm `apps/api/migrations` is unchanged.
- Run `make validate` and report any unavailable environment checks exactly.
- Review the final diff for product, architecture, persistence, API, audit,
  authorization, concurrency, and idempotency consistency.

## Final Review Requirements

Distinguish implemented documentation, verified consistency, assumed future
runtime behavior, deferred decisions, remaining risks, and the recommended
next implementation task.

## Final Review

### Implemented

- Accepted ADR-041 for command-scoped idempotency with domain-specific
  duplicate guards.
- Added an idempotency classification for all 30 business tables in the ERD.
- Added a process ownership matrix, replay lifecycle, layer responsibilities,
  verification scenarios, and deferred schema decisions.
- Corrected current-state documentation to reflect the existing migration SQL
  and database lifecycle tooling without implying runtime business modules.

### Verified

- The entity audit covers all 30 business tables in the ERD.
- `apps/api/migrations` is unchanged.
- Markdown formatting and whitespace checks pass.
- `make validate` passes, including frontend lint, type checking, tests and
  builds; Go vet, tests and build; and Docker Compose configuration validation.

### Assumed

- Future command handlers and workers will implement the accepted transaction,
  authorization, replay, and response-storage behavior. No business API runtime
  currently exists to verify these behaviors end to end.

### Deferred

- Provider webhook inbox identity and uniqueness.
- Outbox consumer receipt storage.
- Notification delivery-attempt persistence.
- Command-specific idempotency retention periods.
- Guest caller scope for public idempotency keys.
- Whether distribution commands need an explicit entity version.

### Plan Variance and Dependencies

- No variance from the approved documentation-only plan.
- No migration, API contract, runtime, generated client, dependency, or
  infrastructure changes were made.

### Remaining Risks

- The blueprint is not runtime enforcement; implementation can still diverge
  until the first vertical slice adds integration tests.
- Provider identity, retention, guest scoping, and asynchronous consumer
  policies must be decided before their related features are implemented.

### Recommended Next Task

Implement the shared command-idempotency framework in the first retry-sensitive
vertical slice, expose its contract in the appropriate OpenAPI document, and
verify replay, payload mismatch, authorization, concurrency, rollback, audit,
and outbox behavior with database-backed integration tests.
