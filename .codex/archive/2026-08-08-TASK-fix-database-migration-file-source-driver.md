# Task: Fix Database Migration File Source Driver
## Executed

## Objective

Make `go run ./cmd/db migrate up` open the repository's filesystem migration
source successfully.

## Constraints

- Preserve the existing `golang-migrate/migrate/v4` runner and migration pairs.
- Do not add dependencies, schema changes, API changes, or unrelated refactors.
- Do not apply migrations to a database as part of verification.

## Acceptance Criteria

- The `file` source driver is registered by the database migration package.
- A regression test fails if the `file://` source driver registration is removed.
- Existing migration discovery and Go checks continue to pass.

## Planned Changes

| File | Action | Purpose |
|---|---|---|
| `.codex/TASK.md` | create | Record the approved active task and constraints. |
| `apps/api/internal/database/migration/migration.go` | modify | Register `golang-migrate`'s filesystem source driver. |
| `apps/api/internal/database/migration/migration_test.go` | modify | Add a focused registration regression test. |

## Verification Requirements

- Focused migration tests.
- `go vet ./...`.
- `go test ./...`.
- `go build ./...`.
- `make validate`.
- `docker compose -f infrastructure/compose.yaml config`.

## Status

Completed.

## Final Review

- Implemented: registered `golang-migrate`'s filesystem source driver and
  added a regression test for `file://` source opening.
- Verified: focused migration/CLI tests, `go vet ./...`, `go test ./...`,
  `go build ./...`, `make validate`, Compose config validation, and the
  read-only migration validation command all passed.
- Assumed: the caller supplies valid database configuration and a reachable
  PostgreSQL instance when applying migrations.
- Deferred: running `migrate up` against a database because it changes schema
  state; use a disposable PostgreSQL instance for that integration check.
- Next recommended task: run the migration up/status/seed/rollback workflow
  against disposable PostgreSQL.
