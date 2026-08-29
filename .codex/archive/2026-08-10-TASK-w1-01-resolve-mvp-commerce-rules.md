# Task: W1-01 Resolve MVP Commerce Rules

## Executed

## Status

Completed and verified on 2026-08-10.

## Tracker

- Workstream: Platform Safety
- Estimate: 4 hours
- Dependencies: None
- Tracker objective: Resolve MVP Event, Offering, direct-checkout, quota,
  payment-evidence, and participant-activation rules.

## Objective

Replace the first commerce slice's open product questions with explicit MVP
rules that W1-02 through W1-07 can implement without inventing behavior.

## Existing Foundation to Reuse

- `docs/PRD.md` and `docs/PRODUCT_MAP.md` already define Event, Offering,
  Purchase, Payment, and Sohibul Qurban boundaries.
- ADR-032 keeps Offering separate from physical Livestock.
- ADR-033 already permits direct checkout without a Shopping Cart.
- ADR-041 already defines command-scoped idempotency.
- The proposed ERD already stores purchase price and participant snapshots.

## Locked MVP Rules

### Event

- Lifecycle:
  `DRAFT -> PUBLISHED -> ACTIVE <-> SUSPENDED -> CLOSED -> ARCHIVED`.
- Only `PUBLISHED` may become `ACTIVE`; only `ACTIVE` may be suspended or
  closed; a suspended event may be reactivated or closed.
- `CLOSED` and `ARCHIVED` are terminal for commerce commands.
- At most one event may be `ACTIVE` at a time.
- Storefront catalogue and checkout are available only for the active event.

### Offering and checkout

- An MVP Offering is an event-scoped sellable package, share, or category. It
  is not a physical Livestock record.
- One Purchase selects exactly one Offering. There is no Shopping Cart or
  purchase-item aggregate in the MVP.
- Checkout records one purchaser, one payer, and one or more intended
  participants. These roles may refer to different parties.
- `participant_count` must be positive and no greater than the Offering's
  per-purchase `participant_capacity`.
- Purchase amount, currency, Offering name/code, unit price, and participant
  names are snapshotted so later catalogue edits do not rewrite history.

### Quota

- Quota is measured in participant units against both the Event quota and the
  selected Offering quota.
- Checkout atomically creates a pending Purchase and reserves its participant
  units for 24 hours. Availability checks and reservation creation occur in
  one database transaction under row locking or an equivalent atomic guard.
- A reservation expires after 24 hours when no payment evidence is awaiting
  review. Evidence submission pauses automatic expiry until Finance decides.
- Activation consumes the reservation. Expiry, purchase cancellation, and
  rejected evidence release it.
- A new evidence submission after release must reacquire quota atomically; it
  fails with conflict when capacity is no longer available.
- Storefront availability is advisory. The checkout command is authoritative.

### Payment evidence

- Common-purchase MVP payment is manual transfer evidence, not a payment
  gateway callback.
- Evidence bytes live in object storage. PostgreSQL stores only an opaque
  object reference plus original filename, media type, byte size, and SHA-256
  digest.
- Accepted media types are JPEG, PNG, and PDF, with a 10 MiB maximum.
- Evidence records are append-oriented. Rejection does not overwrite or delete
  the rejected evidence; resubmission creates another record.
- A submission must declare exactly the Purchase's outstanding amount. The MVP
  rejects underpayment and overpayment instead of creating balance policy.
- Only an authorized Finance or Operations Manager may verify or reject.

### Participant activation

- Payment verification locks the Payment, Purchase, and quota reservation and
  revalidates amount, currency, status, and capacity.
- Successful verification atomically marks Payment `VERIFIED`, records the
  Purchase `PAID` then `ELIGIBLE` transitions, consumes quota, creates one
  active Sohibul Qurban per intended participant, and writes audit/outbox
  effects.
- Activation is performed exactly once. Command idempotency and natural
  uniqueness on `(purchase_id, sequence_no)` prevent duplicates.
- Payment rejection records its reason, releases quota, and leaves the Purchase
  pending for a possible evidence resubmission and quota reacquisition.

## Scope

- Resolve the matching open questions in canonical product documents.
- Add or update accepted ADRs for rules with cross-module or data-integrity
  consequences.
- Keep schema design, runtime implementation, OpenAPI, and deployment in their
  dependent tasks.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `docs/PRD.md` | Modify | Replace the affected open questions with the locked MVP rules. |
| `docs/PRODUCT_MAP.md` | Modify | Mark Offering and checkout questions resolved for Phase 1. |
| `docs/DECISIONS.md` | Modify | Record Event, direct-checkout, quota, evidence, and activation decisions. |
| `.codex/CURRENT_STATE.md` | Modify | Identify W1-01 decisions as documented but not yet implemented. |
| `.codex/TASK.md` | Execute and archive | Preserve the approved task and final review. |

## Acceptance Criteria

1. Canonical documents use the same Event lifecycle and one-active-event rule.
2. Phase 1 explicitly uses one Offering per Purchase and no Shopping Cart.
3. Event and Offering quota reservation, pause, consumption, expiry, release,
   and reacquisition behavior is unambiguous.
4. Evidence storage, validation, append behavior, review authority, and
   rejection behavior are explicit.
5. Participant activation names one atomic transaction and exactly-once guards.
6. Deferred saving, giveaway, gateway, refund, and livestock-allocation policy
   remains unresolved rather than being accidentally decided here.
7. No schema, runtime, OpenAPI, dependency, CI, or infrastructure file changes.

## Verification

- Review all affected open-question lists for stale Phase 1 wording.
- Cross-check the rules against ADR-032, ADR-033, ADR-041, and the current ERD.
- Run `make format-check` and `make validate`.
- Review the diff for product, payment, quota, audit, authorization,
  concurrency, and idempotency consistency.

## Risks and Deferred Work

- Object-storage provider and retention policy remain deferred to deployment
  and compliance work.
- Refunds, partial payments, payment gateways, saving, giveaway, livestock
  allocation, and multi-offering checkout are outside this task.
- W1-03 must translate these rules into additive schema changes; W1-04 must
  define complete transition and permission matrices.

## Final Review Requirements

Distinguish documented rules, verified consistency, assumptions, deferred
behavior, plan variance, remaining risks, and the next dependency-ready task.

## Final Review

### Implemented

- Resolved Phase 1 Event, Offering, one-Offering direct checkout, quota,
  payment-evidence, and participant-activation rules in the PRD and Product
  Map.
- Accepted ADR-042 for the cross-module transaction and data-integrity rules.
- Updated current state without claiming schema or runtime implementation.

### Verified

- The affected Phase 1 open-question wording is removed from the PRD and
  Product Map.
- Product documents and ADR-042 agree on lifecycle, quota timing, evidence
  validation, review authority, and exactly-once activation.
- No runtime, schema, OpenAPI, dependency, CI, or infrastructure file changed.
- `make validate` passes.

### Assumed

- W1-03 and W1-04 will translate the accepted rules into additive schema and
  enforceable lifecycle/permission specifications.

### Deferred

- Saving, Giveaway, refunds, payment providers, evidence retention,
  multi-offering checkout, livestock allocation, and runtime enforcement.

### Plan Variance and Remaining Risk

- No scope variance. Prettier reformatted the changed PRD.
- The rules remain documentation until dependent schema and runtime tasks are
  executed.

### Next Task

Execute W1-02 to accept the HTTP router, migration tooling, operations OIDC
session, and guest Purchase-token approach.
