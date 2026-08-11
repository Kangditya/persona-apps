# Task: W1-07 Configure Staging and CI Migration Checks

## Executed

## Status

Inactive draft. Activate only after W1-02 and W1-06 are executed and current
source is revalidated.

## Tracker

- Workstream: Platform Safety
- Estimate: 3 hours
- Dependencies: W1-02; W1-06
- Tracker objective: Configure staging environment and CI migration checks.

## Objective

Define a provider-neutral staging contract and make GitHub Actions prove the
real PostgreSQL migration/seed lifecycle on every validation run.

## Approved execution amendment

W1-06 configuration now validates OIDC, exact origins, cookie encryption, and
encrypted idempotency response keys. This task additionally rejects destructive
long-running staging and production API configuration; it does not add cloud
infrastructure.

## Existing Foundation to Reuse

- `APP_ENV=staging` and destructive-command guards already exist.
- `infrastructure/compose.yaml` already pins PostgreSQL 18 for local use.
- `.github/workflows/validate.yml` already runs frontend, Go, Compose, and API
  image checks.
- `cmd/db` already supports validate, status, version, up, bounded down, seed,
  and setup commands.

## Staging Contract

- Document required secret/config names for database, OIDC issuer/client,
  redirect URL, allowed Operations origins, secure cookie settings, and private
  evidence object storage. Commit names and validation rules, never values.
- Require `APP_ENV=staging`; keep
  `ALLOW_DESTRUCTIVE_DB_COMMANDS=false`/unset in the long-running API.
- Require TLS at the public edge, secure cookies, explicit CORS origins, private
  evidence objects, least-privilege database/object credentials, and secret
  injection outside source control.
- Deployment order is: build immutable artifact, run migration job, run
  reference seeds, deploy API, then verify `/health` and `/ready`.
- Migration failure stops deployment. API startup never migrates automatically.
- Document rollback as an explicit operator action after backup/compatibility
  review; do not automate production-style rollback in the staging contract.
- Stay provider-neutral: no cloud account, Terraform stack, Kubernetes
  manifests, DNS, certificates, or secret values in this task.

## CI Migration Job

Extend the existing workflow with a PostgreSQL `18-alpine` service and health
check. With `APP_ENV=test` and a disposable `DATABASE_URL`, run in order:

```text
go run ./cmd/db migrate validate
go run ./cmd/db migrate up
go run ./cmd/db migrate version
go run ./cmd/db migrate status
go run ./cmd/db seed run --group reference
go run ./cmd/db seed run --group reference
go test ./...
go run ./cmd/db migrate down --steps 1
go run ./cmd/db migrate up --steps 1
go run ./cmd/db migrate version
```

- The second reference-seed run must succeed without duplicate effects.
- The bounded rollback proves only the latest migration pair and runs only on
  the disposable CI database.
- Keep all existing validation and build steps; do not create another workflow
  when one job can cover the check clearly.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `docs/deployment/STAGING.md` | Create | Define provider-neutral config, secrets, deploy order, probes, and rollback policy. |
| `.github/workflows/validate.yml` | Modify | Add PostgreSQL 18 migration, seed-idempotency, integration-test, rollback, and reapply checks. |
| `apps/api/internal/config/config.go` | Modify | Enforce staging-required auth/origin/cookie configuration from W1-06. |
| `apps/api/internal/config/config_test.go` | Modify | Verify staging validation and destructive defaults. |
| `.codex/CURRENT_STATE.md` | Modify | Record CI coverage and the remaining lack of a provider deployment. |
| `.codex/TASK.md` | Create, execute, archive | Preserve the activated task and final review. |

## Acceptance Criteria

1. Staging requirements are provider-neutral, complete, and contain no secret
   values or committed `.env` file.
2. CI starts PostgreSQL 18 and applies every migration from an empty database.
3. CI reports migration version/status, runs reference seeds twice, executes
   database-backed Go tests, rolls back one step, and reapplies it.
4. Existing frontend, Go, Compose, and Docker build validation remains present.
5. Staging startup cannot silently enable destructive database commands.
6. No cloud-specific infrastructure or production deployment claim is added.

## Verification

- Validate workflow YAML and `docker compose -f infrastructure/compose.yaml
  config`.
- Reproduce the CI database command sequence locally against disposable
  PostgreSQL 18.
- Run `make validate`.
- Search tracked changes for credentials, tokens, `.env`, storage URLs, and
  provider-specific resource identifiers.

## Risks and Deferred Work

- This creates a deployable contract, not an actual cloud environment.
- Backup/restore drills, monitoring, domains, TLS provisioning, evidence
  retention, and production rollout remain separate tasks.
- If CI duration becomes material, split migration checks only after measured
  evidence; one workflow is the MVP.

## Final Review Requirements

Distinguish documented staging requirements, executed CI/local verification,
assumptions, deferred provider work, plan variance, remaining operational
risks, and the next recommended vertical-slice task.

## Final Review

### Implemented

- Provider-neutral staging configuration, security controls, deploy order,
  probes, and explicit rollback policy.
- Staging and production API startup rejection when destructive database
  commands are enabled.
- One existing Validate workflow job now starts PostgreSQL 18 and verifies the
  full migration, reference-seed, database-test, bounded rollback, and reapply
  lifecycle.

### Verified

- actionlint and Docker Compose configuration validation passed.
- Local PostgreSQL 18 reproduced the direct CI command sequence: empty
  database migration to clean version 5, status, two reference-seed runs,
  database-backed tests, one-step rollback, reapply, and final clean version.
- make validate and the scoped secret-pattern scan passed.

### Assumed

- A deployment environment injects the documented secrets and provides a TLS
  edge plus private object storage before staging deployment.

### Deferred

- Cloud provisioning, DNS, certificate automation, secret-manager setup,
  monitoring, backups, restore drills, evidence adapter, and production
  rollout.

### Next Task

Build the first domain vertical slice: common Purchase creation through payment
verification and Sohibul Qurban activation using the verified platform.
