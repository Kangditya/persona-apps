# Task: W4-09 Add Duplicate-Submission, Verification, Authorization, and Activation Tests

## Status

Ready for plan review on 2026-09-03. Planning-only; live tracker remains
`Backlog`. Activate only after W4-01 through W4-08 are implemented and their
focused suites pass.

## Tracker

- Week: Week 4
- Epic: Payment & Activation
- Application: Cross-cutting
- Category: Test
- Priority: P0
- Estimate: 8 hours
- Tracker objective: Add duplicate-submission, verification, authorization, and
  activation tests.
- Dependencies: W4-03, W4-05, W4-06, W4-07, and W4-08.
- Acceptance summary: tests prove duplicate submission, permission,
  verification, quota, rollback, and exactly-once activation behavior.

## Objective

Audit the final Week 4 vertical slice and add only the smallest missing
regressions needed to prove this real flow end to end:

```text
ephemeral Purchase token
-> private bounded evidence submission
-> submitted verification queue/detail
-> authorized verify or reject
-> Payment/Purchase/quota/ledger/audit/outbox transaction
-> exactly-once Sohibul Qurban activation
-> safe Storefront current status
```

This is a risk-based gap-closing task, not a second copy of every W4 unit test.

## Context

W3-09 already established composed PostgreSQL, authorization-ordering,
semantic-idempotency, quota-race, response-exposure, PWA, and real-browser
patterns. W4 adds sensitive file storage, Purchase Bearer authentication,
append-oriented financial attempts, privileged decisions, ledger effects,
multi-aggregate status changes, participant activation, and two new UIs. Those
risks require final cross-layer evidence after implementation—not before.

## Source of Truth

- approved and executed W4-01 through W4-08 plans/final reports/diffs;
- all canonical product, architecture, decision, convention, lifecycle,
  permission, authentication, database, deployment, and OpenAPI sources cited
  by those tasks;
- current Go/frontend/storage/PWA source and complete focused test inventory;
- W3-09 coverage methodology and disposable-PostgreSQL/browser evidence;
- `.codex/CURRENT_STATE.md`, roadmap, and live tracker W4-09 metadata.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

At activation:

1. Re-read every W4 final report and inspect the complete current diff/source.
2. Map each W4 acceptance criterion to a named automated or recorded manual
   check.
3. Mark uncovered risks before editing tests.
4. Reuse owning tests; add only missing cross-boundary proof.
5. Record a PLAN VARIANCE before fixing any production/contract defect revealed
   by a new failing test.

## Scope

### In Scope

- A concise Week 4 acceptance-to-evidence matrix naming reused tests and gaps.
- Public upload boundary: Bearer ordering, multipart shape, content sniffing,
  filename/path safety, 10 MiB boundary, digest, amount/currency, safe errors,
  no-store/CORS/rate limit, and zero side effects.
- Storage contract: immutable create-only object, restart retrieval, private
  permissions/ACL, rollback cleanup, crash-orphan policy evidence, missing
  object, and no secret/reference leakage.
- Submission duplicates: same semantic intent replay, changed intent conflict,
  concurrent same key, concurrent different keys, rejected resubmission, and
  one-current-submission invariant if approved.
- Quota: pause expiry, released/expired reacquisition, numbered attempts,
  unavailable capacity, and repeated contention under Event/Offering locks.
- Operations decisions: session/Origin/CSRF/permission before replay; verify,
  reject reason, same/different/new-key behavior, and verify-vs-reject race.
- Atomic verify effects: exact Payment/Purchase histories, final statuses,
  reservation, one ledger credit, N participant outcomes/histories, audit,
  outbox, and replay row.
- Failure injection at Payment, history, Purchase, quota, ledger, each activation
  position, audit, outbox, replay encryption/update, and storage boundary with
  exact zero/unchanged effects.
- Natural duplicate guards for ledger entry and `(purchase_id, sequence_no)`;
  identical replay creates no histories/audit/outbox duplicates.
- Exact public/Operations response and log/audit/outbox exposure allowlists.
- Operations and Storefront API/query/component/path/PWA/accessibility tests plus
  one real cross-app browser journey against real Go/PostgreSQL/storage/OIDC.
- Focused race runs, migration lifecycle when W4 adds a migration, full
  repository validation, and final risk/deferred review.

### Out of Scope

- Arbitrary coverage percentage, redundant snapshots, fuzz/load/soak/penetration
  claims, or a new test/browser/storage/mock framework.
- Provider webhook, partial/overpayment, refund, Saving, Giveaway, cancellation,
  durable tracking/token recovery, W5/W6 completion, polling/SSE, or audit UI.
- Real production/staging bucket/database/OIDC access, production evidence, or
  committed credentials/fixtures.
- Fixing unrelated code merely to make preferred test structure possible.

## Existing Evidence to Reuse

- W3 checkout tests already own Purchase token generation/hash, durable encrypted
  replay, public CORS/rate limit, transaction rollback, quota contention, and
  raw-token non-retention before W4's narrow handoff.
- Platform auth tests own session/Origin/CSRF/permission ordering; W4 needs a
  composed Payment proof, not duplicate generic cases.
- Platform idempotency, audit, outbox, transaction, HTTP error, and PWA tests own
  their generic invariants.
- W4-01 through W4-08 focused tests must own local domain/repository/adapter/UI
  branches before this task begins.

## Coverage Matrix Requirements

For every acceptance criterion record:

- owning task and risk;
- named current test/manual evidence;
- layer: domain, repository, adapter, HTTP, composed API, frontend, browser;
- database/storage dependency and whether it actually ran;
- missing assertion or explicit reuse decision;
- final result and any plan variance.

Do not call skipped database/provider/browser tests passing evidence.

## Critical Composed Scenarios

### Submission

1. Valid token/file commits exactly one object, Payment/history, quota pause or
   reacquisition, outbox, and replay.
2. Same key/content with different multipart boundary returns byte-equivalent
   response and zero extra effects.
3. Same key with changed bytes/amount/currency conflicts; no new object/effect.
4. Two concurrent same-key requests create once. Two different keys obey the
   approved in-review duplicate policy.
5. Wrong/missing token, bad media magic, 0/limit+1 bytes, traversal filename,
   amount mismatch, and quota failure produce zero committed effects.

### Decision

1. Authorized verify commits all canonical effects and N active participants.
2. Same-key retry returns the original action result and no duplicate effect.
3. Different key against terminal Payment conflicts with unchanged row counts.
4. Concurrent verify/verify and verify/reject produce exactly one terminal
   outcome.
5. Reject preserves evidence, releases quota, keeps Purchase pending, records
   reason/audit, and creates no ledger/activation.
6. Every injected failure rolls back the entire database transaction and leaves
   the command safely retryable.

### Cross-Surface

1. Storefront checkout token passes only ephemerally, submits evidence, and
   refreshes safe status; reload loses credential/private state.
2. Operations payment.read-only can inspect metadata but cannot retrieve bytes
   or decide; payment.verify can review and decide.
3. Logout removes all private queries/object URLs; service workers never cache
   API/auth/payment/evidence/participant data.
4. Public and Operations responses contain only their exact allowlisted fields.

## Planned File Changes

- existing W4 owning `*_test.go`, `*.test.ts`, and component tests only where
  the matrix proves a gap;
- one composed Go/PostgreSQL/storage integration test file if existing app tests
  cannot express the complete flow cleanly;
- existing frontend/PWA/path/API tests for missing cross-surface assertions;
- production/contract files only through a recorded variance after a failing
  regression demonstrates a real W4 defect;
- `.codex/TASK.md`, archive, and current state only during later approved
  execution.

No new testing dependency is planned.

## Acceptance Criteria

1. Every W4-01 through W4-08 criterion maps to named passing evidence or an
   explicitly documented deferred blocker.
2. Duplicate submission and decision retries create exactly one intended
   business/storage effect; semantic changes conflict safely.
3. Authorization, Origin, CSRF, permission, and Purchase token checks precede
   replay/storage/repository access at composed routes.
4. Verify and reject concurrency cannot split or duplicate Payment/Purchase/
   quota/ledger/activation/audit/outbox state.
5. Verification creates one ledger credit and exactly N active outcomes/history
   rows; replay creates none additional.
6. Rejection releases quota, preserves append-oriented evidence/history, keeps
   Purchase pending, and supports a new capacity-checked submission.
7. All injected failures leave exact pre-command authoritative state and no
   completed replay; storage cleanup/orphan outcome matches the accepted ADR.
8. Public/Operations UI and APIs preserve token/evidence/internal-field privacy,
   accessible states, network-only PWA behavior, and safe logout/reload.
9. Focused Go/TypeScript, race, PostgreSQL, storage, migration, OpenAPI,
   production builds, Compose, real-browser, and `make validate` evidence pass.
10. No skipped check, mock-only flow, migration existence, or UI render is
    mislabeled as end-to-end proof.

## Testing and Verification

Use existing commands and a uniquely named disposable PostgreSQL 18 database
plus temporary/private evidence store. Run focused tests first, then:

```bash
cd apps/api
go test ./internal/payment/... ./internal/purchasing/... ./internal/participant/... ./internal/app/... -count=1
go test -race ./internal/payment/... ./internal/purchasing/... ./internal/participant/...
go vet ./...
go test ./...
go build ./...
cd ../..
pnpm --filter @persona-apps/storefront-web test
pnpm --filter @persona-apps/operations-web test
pnpm --filter @persona-apps/storefront-web build
pnpm --filter @persona-apps/operations-web build
make validate
docker compose -f infrastructure/compose.yaml config
```

If a W4 migration exists, also validate/apply/version/down-one/reapply it through
the repository database CLI. Validate both OpenAPI contracts and run the real
cross-app browser journey at representative desktop and 360px widths.

## Review and Risks

- W4-09 cannot compensate for unresolved storage, zero-price, participant,
  reference/method, replay-retention, public-status, or evidence-permission
  decisions. Those must be closed or explicitly remain blockers.
- Race tests must use distinct connections/transactions and repeat enough to
  exercise contention, while avoiding a fake claim of load testing.
- Evidence bytes must use synthetic fixtures generated for tests and be removed
  with the disposable store/database. Never use personal or production data.
- The shortest valid test is preferred; reuse stable W3/platform evidence and
  avoid a bespoke end-to-end framework.

## Deliverables

- Week 4 acceptance/evidence matrix.
- Minimal missing regression tests and any evidence-backed defect fixes.
- Exact disposable PostgreSQL/storage, race, contract, frontend, PWA,
  accessibility, browser, and full-validation results.
- Current-state update and canonical task archive only after all acceptance
  criteria are verified.

## Final Report Requirements

Report the matrix, exact commands/results, database/object row counts, race
outcomes, authorization ordering, replay hashes/retention without raw keys,
storage cleanup, audit/outbox allowlists, browser states/sizes, plan variances,
remaining risks, and next task. Distinguish implemented, verified, assumed,
blocked, and deferred behavior.
