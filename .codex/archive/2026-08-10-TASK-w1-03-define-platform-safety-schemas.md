# Task: W1-03 Define Platform Safety Schemas
## Executed

## Status

Completed and verified on 2026-08-10. W1-01 was executed and the current
source was revalidated before direct execution.

## Revalidation

- Migrations 0001 through 0004, ADR-041, ADR-042, and ADR-043 match this
  task prerequisite and constraints.
- audit_log.reason already exists, so migration 0005 preserves it and adds
  only the missing source and permission audit context.
- Plan variance: the migration CLI test asserts the discovered pair count, so
  its expected value changes from four to five with migration 0005.

## Tracker

- Workstream: Platform Safety
- Estimate: 6 hours
- Dependencies: W1-01
- Tracker objective: Define Event, Offering, Party, Purchase, Payment, Sohibul
  Qurban, Audit, and Idempotency schemas.

## Objective

Translate W1-01 rules and ADR-041 into one additive PostgreSQL migration while
preserving all historical migration files and the existing domain boundaries.

## Existing Foundation to Reuse

- Migrations `0001` through `0004` already define the 30-table proposed ERD,
  audit/outbox records, and command idempotency ledger.
- Purchases already snapshot Offering name, kind, unit price, capacity, amount,
  and currency.
- `financial_ledger_entries.idempotency_key` and
  `idempotency_records(namespace, idempotency_key)` already have correct
  ownership under ADR-041.
- `operator_users.external_subject` already maps external operator identities.

## Required Schema Delta

Create `0005_mvp_commerce_safety.up.sql` and matching down migration. Do not
edit `0001` through `0004`.

### Event and Offering

- Extend Event status checks with `SUSPENDED`.
- Add a partial unique index that permits at most one `ACTIVE` Event.
- Add nullable non-negative `offerings.participant_quota`; null means no
  Offering-specific cap beyond Event quota.
- Keep `offerings.participant_capacity` as the maximum participant units in one
  Purchase; do not reinterpret it as total Offering quota.

### Intended participants and Purchase access

- Add `purchase_participants` with UUID identity, `event_id`, `purchase_id`,
  nullable `party_id`, positive `sequence_no`, immutable display-name snapshot,
  and timestamps.
- Enforce event-aware Purchase reference and unique
  `(purchase_id, sequence_no)`.
- Add nullable unique `purchases.access_token_hash bytea`; common guest
  checkout must populate it, while future non-public channel creation need not.
  Store only SHA-256 output, never the raw guest token.
- Keep `sohibul_qurban` as the post-eligibility outcome. Activation copies each
  intended participant exactly once using the same sequence number.

### Quota reservations

- Add `quota_reservations` with UUID identity, event-aware Purchase and
  Offering references, positive `attempt_no`, positive `participant_units`,
  status, nullable `expires_at`, `consumed_at`, `released_at`, `release_reason`,
  positive version, and timestamps.
- Status values are `RESERVED`, `CONSUMED`, `RELEASED`, and `EXPIRED`.
- Enforce unique `(purchase_id, attempt_no)`, at most one `RESERVED` row per
  Purchase, and event-aware Offering/Purchase consistency. Reacquisition after
  release creates the next attempt and preserves prior attempts.
- `RESERVED` normally has `expires_at`; evidence submission may set it null
  while review is pending. Terminal timestamps must match terminal status.
- Index active reservations by Event and Offering for locked quota checks and
  index expiring reservations for the expiry command.

### Payment evidence

- Keep `payment_records.evidence_reference` as the opaque object key.
- Add `evidence_filename`, `evidence_media_type`, `evidence_size_bytes`, and
  `evidence_sha256` with all-or-none checks when evidence exists.
- Size must be positive and no greater than 10 MiB. Media type is restricted to
  JPEG, PNG, or PDF. SHA-256 is exactly 32 bytes.
- Do not add object bytes, public URLs, provider SDK fields, or mutable evidence
  replacement fields.

### Operations sessions and audit

- Add `operator_sessions` with UUID identity, operator reference, unique
  session-token SHA-256 hash, CSRF-token SHA-256 hash, permission snapshot,
  expiry, revocation, created timestamp, and last-seen timestamp.
- Add `audit_log.source`, `audit_log.permission`, and optional
  `audit_log.reason` so privileged commands preserve authorization context.
- Keep audit rows append-only and retain the existing request identifier and
  before/after payloads.

### Idempotency

- Do not add `idempotency_key` to aggregates, histories, audit, or outbox.
- Keep the shared replay ledger and financial effect key unchanged unless an
  integration test proves a concrete missing constraint.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `apps/api/migrations/0005_mvp_commerce_safety.up.sql` | Create | Add the required Phase 1 safety schema. |
| `apps/api/migrations/0005_mvp_commerce_safety.down.sql` | Create | Reverse only migration 0005 in dependency order. |
| `docs/database/ERD.md` | Modify | Add new entities, fields, relationships, and ownership notes. |
| `docs/database/MIGRATION_PLAN.md` | Modify | Inventory 0005, compatibility, rollout, and rollback. |
| `docs/database/REVIEW.md` | Modify | Record constraint and index rationale. |
| `.codex/CURRENT_STATE.md` | Modify | Record schema artifacts without claiming applied environments. |
| `.codex/TASK.md` | Create, execute, archive | Preserve the activated task and final review. |

## Acceptance Criteria

1. Migration 0005 is additive and historical migrations are byte-for-byte
   unchanged.
2. Event, Offering, intended-participant, quota, Purchase-token, evidence,
   session, and audit requirements have explicit constraints and indexes.
3. Purchase participant input remains separate from activated Sohibul Qurban.
4. Down migration removes only 0005 objects and restores altered checks safely.
5. ERD, migration plan, review, and SQL agree.
6. No per-entity idempotency columns or speculative provider inbox are added.

## Verification

- Run migration discovery/validation and all Go tests.
- Apply `0001..0005` to disposable PostgreSQL 18, inspect version/status, roll
  back exactly one step, and reapply it.
- Confirm a second reference-seed run is harmless.
- Compare checks, foreign keys, unique constraints, and indexes with the ERD.
- Run `make validate` and confirm historical migration files are unchanged.

## Risks and Deferred Work

- Provider webhook inbox, evidence retention, and object-storage implementation
  remain deferred until concrete adapters exist.
- Permission administration remains outside this schema task.
- Runtime transaction behavior remains W1-06 work.

## Final Review Requirements

Distinguish created schema, disposable-database verification, assumptions,
deferred integration fields, migration risk, and the next dependency-ready
task.

## Final Review

### Implemented

Migration 0005 adds Event suspension and one-active-event enforcement, Offering
quota, intended participants, quota attempts, evidence metadata, Purchase-token
hashes, Operations sessions, and audit source/permission. It preserves the
separation between participant input and the Party-backed Sohibul Qurban
outcome, plus existing generic idempotency ownership.

### Verified

Migration discovery reports five valid pairs and migrations 0001 through 0004
are unchanged. Isolated PostgreSQL 18 applied 0001 through 0005, reported
version 5 and a clean state, ran reference seeds twice, rolled back 0005, and
reapplied it. Live constraints, foreign keys, partial indexes, and hash checks
agree with the ERD. make validate and the focused migration CLI test pass.

### Assumed

Future checkout and verification commands will lock quota totals and resolve an
intended participant Party before automatic activation.

### Deferred

OIDC, session, CSRF, permission, object-storage, quota-command, provider inbox,
and evidence-retention runtime behavior remain unimplemented.

### Risks

The round trip used one empty local PostgreSQL 18 database; staging and
production deployment compatibility remain unverified.

### Next Task

Execute W1-04 to define the authoritative Phase 1 lifecycle and permission
specifications that W1-05 and W1-06 will consume.
