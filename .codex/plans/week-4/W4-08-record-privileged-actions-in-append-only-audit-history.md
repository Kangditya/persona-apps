# Task: W4-08 Record Privileged Actions in Append-Only Audit History

## Status

Ready for plan review on 2026-09-03. Planning-only; live tracker remains
`Backlog`. This task hardens and verifies the audit effects that W4-03 must
already write atomically; it must not add duplicate audit rows after the fact.

## Tracker

- Week: Week 4
- Epic: Payment & Activation
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 4 hours
- Tracker objective: Record privileged actions in append-only audit history.
- Dependencies: W1-06, W4-03, and W4-05.
- Acceptance summary: Payment decisions and activation actions commit immutable
  minimized audit records with their authoritative changes.

## Objective

Define, implement where missing, and prove the minimized append-only audit
contract for Payment verification/rejection and their coupled eligibility/
activation effects. One parent decision audit entry must capture authority and
safe before/after summaries in the same transaction; audit failure must abort
the entire decision.

## Context

Migration 0003 provides `audit_log`; migration 0005 adds source and permission.
The existing `platform/audit.Write` inserts caller-supplied before/after JSON in
the caller transaction. Event/Offering commands already use it. There is no
Payment command yet, no Payment-specific payload policy, and no database rule
preventing direct UPDATE/DELETE beyond application/database-role discipline.

Canonical lifecycle calls for a parent Payment audit record, not one row for
every derived status/history/outbox/participant effect. Rejection requires a
reason. Audit must not contain evidence object identity, digest, filename,
contacts, credentials, or raw command keys.

## Source of Truth

- W4-03 through W4-05 plans and verified outputs;
- PRD reporting/audit, security, privacy, and traceability rules;
- Architecture Audit section and ADR-017, ADR-019, ADR-020, ADR-024, ADR-038,
  ADR-041, ADR-042, and ADR-044;
- `docs/domain/COMMERCE_LIFECYCLES.md` and `docs/security/PERMISSIONS.md`;
- ERD audit ownership, migrations 0003/0005, existing audit writer/tests, and
  Event/Offering audit entries;
- Operations Payment contract only for target/action correlation, not audit
  payload exposure.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

At activation, map every W4 privileged command to the exact existing audit row
written in its transaction. Add code only for a demonstrated missing/unsafe
field or invariant; do not create a second logging system.

## Scope

### In Scope

- Define exact action names, target type/ID, source, permission, actor, request
  ID, reason, and minimized before/after shapes for verify and reject.
- Include safe status/amount/currency/quota/activation counts necessary to
  understand the decision, without copying full aggregates.
- Require `actor_operator_id`, stable actor reference, `operations-web` source,
  `payment.verify` permission, Payment target, and request ID according to the
  existing writer/schema.
- Require rejection reason in Payment/history/quota/audit; verification audit
  uses no fabricated reason unless an approved operator reason is added to the
  contract.
- Keep one parent audit entry for the Payment decision. Purchase/Sohibul histories
  and outbox events carry their own lifecycle trace without duplicate audit
  rows.
- Validate required audit entry fields before SQL if the current writer permits
  incomplete privileged records.
- Prove normal application code only inserts audit rows and has no audit update/
  delete command or route.
- Add focused writer/payload/transaction/PostgreSQL/authorization/exposure tests.

### Out of Scope

- Audit list/UI/export, retention duration, legal hold, signing/hash chaining,
  WORM storage, SIEM forwarding, anomaly detection, or organization tenancy.
- Database role provisioning or a trigger that blocks migrations/admin recovery
  unless separately approved.
- Per-history/per-participant duplicate audit rows, guest submission audit,
  evidence metadata, or raw before/after database rows.
- Rewriting Event/Offering audit records or building generic reflection-based
  redaction.

## Existing State

- `audit_log` is append-only by application convention and has no `updated_at`.
- `audit.Write` serializes arbitrary caller payloads and writes within `*sql.Tx`;
  it validates client IP but not all required privileged fields.
- Operations auth context exposes operator ID/display name and the handler knows
  permission/source/request ID/client IP.
- W4-03 must write audit before its replay transaction commits. W4-08 verifies
  and tightens that behavior; it cannot repair already-committed commands.

## Target Audit Shape

- Action names are stable, explicit Payment decision verbs accepted with W4-03.
- Target is the Payment UUID. Before includes Payment/Purchase/reservation status
  and safe exact amount/currency. After includes terminal Payment status,
  resulting Purchase/quota status, ledger-created boolean, and activated count.
- Rejection reason is present and bounded; verification reason is absent unless
  the contract explicitly collects one.
- Payload excludes Party display/contact, evidence filename/reference/digest/
  bytes, tokens/hashes, session/CSRF, idempotency key/hash/namespace, storage
  details, SQL, and full outbox/ledger records.

## Planned File Changes

- W4-03 Payment command audit-entry construction and tests;
- `apps/api/internal/platform/audit/writer.go` and tests only for proven required
  validation/sanitization shared by privileged commands;
- W4 composed PostgreSQL integration tests for exact audit rows and rollback;
- canonical docs/current state only if accepted audit naming/retention policy is
  clarified;
- no frontend, route, schema, or dependency change is currently planned.

## Acceptance Criteria

1. Each successful verify/reject intent commits exactly one parent audit row in
   the same transaction as authoritative state and replay.
2. Audit records the authenticated operator, source, `payment.verify`, request
   ID, action, Payment target, safe before/after summaries, and rejection reason.
3. No derived participant/history/outbox effect creates redundant privileged
   audit rows.
4. Missing audit-required context or injected audit failure aborts Payment,
   Purchase, quota, ledger, activation, outbox, and replay effects.
5. Replay creates no new audit row; idempotency conflict/authorization/CSRF/
   validation/state conflict creates none.
6. Evidence, storage, contacts, credentials, command keys/hashes, and internal
   errors are absent from audit JSON and logs.
7. No normal API route or repository method mutates/deletes audit history.
8. Exact PostgreSQL readback proves values, nullability, transaction ownership,
   and append-only behavior for the W4 slice.

## Testing and Verification

- Writer tests for required fields, safe client IP, JSON failures, SQL errors,
  and preserved causes.
- Payment payload tests assert exact allowlists and forbidden-string absence.
- PostgreSQL composed tests query audit rows after verify/reject/replay/conflict/
  rollback and assert counts plus exact context.
- Authorization-order tests prove forbidden callers never write or read replay.
- Search production source for UPDATE/DELETE against `audit_log` and review
  database application-role privileges if available.

```bash
cd apps/api
go test ./internal/platform/audit/... ./internal/payment/... ./internal/app/... -count=1
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
```

## Risks and Deferred Work

- Append-only is currently enforced by application paths, schema shape, and
  deployment role discipline, not a database immutability trigger. Add stronger
  enforcement only with an approved operational/admin recovery design.
- Audit retention remains intentionally open. Do not delete, expire, or promise
  indefinite retention in this task.
- One parent audit row is the minimum coherent record; histories and outbox
  already trace derived entities.
- W4-08 completion depends on W4-03/W4-05 composed behavior and cannot be
  claimed from audit-writer unit tests alone.

## Deliverables and Final Report

Deliver exact Payment decision audit policy, minimal writer/command changes,
database readback, rollback/replay/security evidence, and no audit duplication.
Report action/target/payload allowlists, actor/permission/source/reason, row
counts, enforcement ceiling, validation, and deferred retention/export work.
