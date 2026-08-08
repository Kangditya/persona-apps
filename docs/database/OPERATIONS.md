# Database Operations

The Go database command owns migration and seed lifecycle operations. Run the
Make targets from the repository root; direct CLI commands run from
`apps/api`.

## Configuration

Load the local values from `.env.example` into an untracked `.env` file or
export them in the command environment:

```text
DATABASE_URL=postgres://...
APP_ENV=development|test|staging|production
ALLOW_DESTRUCTIVE_DB_COMMANDS=false
```

`DATABASE_URL` and `APP_ENV` are required. `ALLOW_DESTRUCTIVE_DB_COMMANDS`
defaults to `false`. Rollbacks in staging or production require an explicit
`true` value. Development seeds are accepted only in `development` or `test`.
The optional `MIGRATIONS_DIR` overrides the default `migrations` directory
when the direct CLI is run from `apps/api`.

The API does not run migrations or seeds during startup.

## Migration commands

```bash
make db-validate
make db-status
make db-version
make db-migrate
make db-migrate-steps steps=1
make db-rollback
make db-rollback-steps steps=1
make db-migration-create name=create_qurban_events
```

Equivalent direct commands:

```bash
cd apps/api
go run ./cmd/db migrate validate
go run ./cmd/db migrate status
go run ./cmd/db migrate version
go run ./cmd/db migrate up
go run ./cmd/db migrate up --steps 1
go run ./cmd/db migrate down --steps 1
go run ./cmd/db migrate create create_qurban_events
```

`db-rollback` is bounded to one migration. A full rollback is never the
default. `goto` and `force` are intentionally not exposed; dirty migration
recovery must be handled as an explicit future operational capability.

The runner consumes the existing `apps/api/migrations/NNNN_name.up.sql` and
`NNNN_name.down.sql` pairs. Validation checks names, pairs, ordering, duplicate
versions, and non-empty files. PostgreSQL remains the SQL syntax validator when
a migration is executed. The runner uses `schema_migrations` and the
PostgreSQL adapter's advisory lock.

## Seed commands

```bash
make db-seed-list
make db-seed-status
make db-seed name=development.sample-event
make db-seed-all
make db-seed-reference
make db-seed-development
```

Equivalent direct commands:

```bash
cd apps/api
go run ./cmd/db seed list
go run ./cmd/db seed status
go run ./cmd/db seed run development.sample-event
go run ./cmd/db seed run --all
go run ./cmd/db seed run --group reference
go run ./cmd/db seed run --group development
```

Seed history is stored independently in `schema_seeds(name, executed_at)`.
Each selected seed runs in its own transaction, takes a PostgreSQL advisory
transaction lock, and records history only after success. A registered seed
with an existing history row is skipped. Seed names are immutable versions;
changed seed behavior requires a new name because no checksum policy is
needed yet.

The reference group is currently empty because the schema has no legitimate
production lookup/bootstrap rows. The development group contains only
`development.sample-event`, which inserts deterministic demo event, party,
location, and offering rows using stable UUIDs and conflict-safe inserts.
`seed run --all` excludes development seeds outside development/test.

## Workflows

First-time local database:

```bash
make db-setup
make db-seed-development   # optional local/demo data
```

`db-setup` validates the migration source, applies pending migrations, and
runs reference seeds only.

Normal developer update:

```bash
git pull
make db-status
make db-migrate
make db-seed-reference
```

Create a migration, then edit and review both generated files before applying:

```bash
make db-migration-create name=create_qurban_events
make db-validate
make db-migrate
```

Rollback the latest migration only after confirming its data is disposable or
backed up:

```bash
make db-rollback
```

## CLI and Make parity

| Go command                        | Make target                          |
| --------------------------------- | ------------------------------------ |
| `db migrate validate`             | `make db-validate`                   |
| `db migrate status`               | `make db-status`                     |
| `db migrate version`              | `make db-version`                    |
| `db migrate up`                   | `make db-migrate`                    |
| `db migrate up --steps N`         | `make db-migrate-steps steps=N`      |
| `db migrate down --steps 1`       | `make db-rollback`                   |
| `db migrate down --steps N`       | `make db-rollback-steps steps=N`     |
| `db migrate create NAME`          | `make db-migration-create name=NAME` |
| `db seed list`                    | `make db-seed-list`                  |
| `db seed status`                  | `make db-seed-status`                |
| `db seed run NAME`                | `make db-seed name=NAME`             |
| `db seed run --all`               | `make db-seed-all`                   |
| `db seed run --group reference`   | `make db-seed-reference`             |
| `db seed run --group development` | `make db-seed-development`           |
| `db setup`                        | `make db-setup`                      |

Make recipes are thin wrappers. They do not contain SQL, credentials, or
environment-specific connection strings, and required `steps`/`name`
variables fail with a usage message.

## Deployment and CI boundary

Run database lifecycle operations as an explicit deployment job before the API
instances start:

```text
build artifact → migrate up → reference seeds when required → start API
```

Do not make every API replica run migrations at startup. CI can run
`make db-validate`, unit tests, and Go vet/build without access to a shared
database. A future disposable PostgreSQL job should run migration up/version,
reference seed twice, status, bounded rollback, and re-apply. This repository
does not connect these commands to staging or production automatically.
