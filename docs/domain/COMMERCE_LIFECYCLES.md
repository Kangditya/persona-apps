# Phase 1 Commerce Lifecycles

## Authority

This document is the authoritative Phase 1 lifecycle policy.

Operational Event, team, livestock, allocation, slaughter, distribution,
incident, projection, and degraded-connectivity lifecycles are defined in
`EVENT_OPERATIONS_LIFECYCLES.md`. Commerce eligibility remains independent of
attendance and field execution.

It applies ADR-042 and ADR-044 to migration 0005 for W1-05 OpenAPI and W1-06
runtime work. It adds no runtime handler.

Each accepted command authenticates and authorizes before idempotency lookup.
One transaction locks authoritative rows, verifies state and version, persists
state, history, audit, outbox, and a replay response, then commits once.
Purchase, Payment, and Sohibul Qurban use their status-history tables. Event,
Offering, and quota reservation use append-only audit_log before/after status
as the current transition history because no dedicated history table exists. A
future history table requires an additive migration.

## Event

| Command    | Transition                    | Actor and permission | Guard and effects                                            | Audit, outbox, replay                | Error                                |
| ---------- | ----------------------------- | -------------------- | ------------------------------------------------------------ | ------------------------------------ | ------------------------------------ |
| Create     | none to DRAFT                 | event.manage         | Validate unique year and create DRAFT.                       | Audit, outbox, replay when declared. | validation_failed or state_conflict. |
| Publish    | DRAFT to PUBLISHED            | event.manage         | Lock Event and require DRAFT.                                | Audit, EventPublished, replay.       | state_conflict.                      |
| Activate   | PUBLISHED to ACTIVE           | event.manage         | Lock Event and active-event guard.                           | Audit, EventActivated, replay.       | state_conflict.                      |
| Suspend    | ACTIVE to SUSPENDED           | event.manage         | Lock Event and require ACTIVE; commerce becomes unavailable. | Audit, EventSuspended, replay.       | state_conflict.                      |
| Reactivate | SUSPENDED to ACTIVE           | event.manage         | Lock Event and active-event guard.                           | Audit, EventActivated, replay.       | state_conflict.                      |
| Close      | ACTIVE or SUSPENDED to CLOSED | event.manage         | Lock Event and require current allowed status.               | Audit, EventClosed, replay.          | state_conflict.                      |
| Archive    | CLOSED to ARCHIVED            | event.manage         | Lock Event and require CLOSED.                               | Audit, EventArchived, replay.        | state_conflict.                      |

## Offering

| Command          | Transition                                   | Actor and permission | Guard and effects                                                      | Audit, outbox, replay                | Error                                |
| ---------------- | -------------------------------------------- | -------------------- | ---------------------------------------------------------------------- | ------------------------------------ | ------------------------------------ |
| Create           | none to DRAFT                                | offering.manage      | Validate Event, code, price, capacity, and quota.                      | Audit, outbox, replay when declared. | validation_failed or state_conflict. |
| Publish          | DRAFT or UNAVAILABLE to PUBLISHED            | offering.manage      | Lock Offering and Event; public exposure requires ACTIVE Event.        | Audit, OfferingPublished, replay.    | state_conflict.                      |
| Mark unavailable | PUBLISHED to UNAVAILABLE                     | offering.manage      | Lock Offering and require PUBLISHED. Purchase snapshots do not change. | Audit, OfferingUnavailable, replay.  | state_conflict.                      |
| Archive          | DRAFT, PUBLISHED, or UNAVAILABLE to ARCHIVED | offering.manage      | Lock Offering and require non-ARCHIVED status.                         | Audit, OfferingArchived, replay.     | state_conflict.                      |

## Quota reservation

| Command      | Transition                          | Actor and permission              | Guard and effects                                                                               | Audit, outbox, replay                                                          | Error                                                          |
| ------------ | ----------------------------------- | --------------------------------- | ----------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ | -------------------------------------------------------------- |
| Reserve      | none to RESERVED                    | Guest checkout                    | Lock Event and Offering totals, validate availability, create next attempt with 24-hour expiry. | Attempt is history; checkout replay and PurchaseReserved. Audit when assisted. | quota_unavailable, validation_failed, or idempotency_conflict. |
| Pause expiry | RESERVED to RESERVED                | Guest Purchase token              | Lock Purchase and reservation; null expiry only while evidence awaits review.                   | Evidence replay and EvidenceSubmitted.                                         | unauthenticated or state_conflict.                             |
| Consume      | RESERVED to CONSUMED                | payment.verify command            | Lock Payment, Purchase, reservation; revalidate amount, currency, status, and quota.            | Audit, QuotaConsumed, parent replay.                                           | quota_unavailable or state_conflict.                           |
| Release      | RESERVED to RELEASED                | rejection or cancellation command | Lock reservation, require RESERVED, set released_at and reason.                                 | Privileged audit, QuotaReleased, parent replay.                                | state_conflict.                                                |
| Expire       | RESERVED to EXPIRED                 | System worker                     | Lock due reservation and require unpaused expiry.                                               | System audit, QuotaExpired, deterministic worker replay or state guard.        | state_conflict.                                                |
| Reacquire    | terminal attempt plus next RESERVED | Guest Purchase token              | Lock totals and create next numbered attempt.                                                   | New history row, evidence replay, outbox.                                      | quota_unavailable or state_conflict.                           |

Terminal reservation states never reopen. SQL permits one RESERVED attempt;
locked totals or an equivalent atomic guard protect Event and Offering quota.

## Purchase and Payment

| Command        | Transition                                                                         | Actor and permission                        | Guard and effects                                                                                                                            | History, audit, outbox, replay                                                                                                    | Error                                                                          |
| -------------- | ---------------------------------------------------------------------------------- | ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| Checkout       | DRAFT to PENDING_PAYMENT                                                           | Guest checkout                              | Create COMMON Purchase, intended participants, token hash, and reservation in one transaction.                                               | Purchase history, assisted audit, PurchasePendingPayment, public replay.                                                          | validation_failed, quota_unavailable, or idempotency_conflict.                 |
| Cancel         | DRAFT or PENDING_PAYMENT to CANCELLED                                              | Guest Purchase token or privileged override | Lock Purchase and reservation; release quota when present.                                                                                   | Purchase history, override audit, PurchaseCancelled, replay.                                                                      | unauthenticated, forbidden, or state_conflict.                                 |
| Verify payment | SUBMITTED to VERIFIED; PENDING_PAYMENT to PAID then ELIGIBLE; RESERVED to CONSUMED | payment.verify                              | Lock Payment, Purchase, reservation, and positions. Require exact amount/currency, valid status/capacity, and a Party for every participant. | Payment history, two Purchase history rows, audit permission, ledger key, outbox, replay, and Sohibul activation commit together. | validation_failed, quota_unavailable, state_conflict, or idempotency_conflict. |
| Reject payment | SUBMITTED to REJECTED; RESERVED to RELEASED; Purchase stays PENDING_PAYMENT        | payment.verify                              | Lock Payment, Purchase, reservation; require rejection reason. No ledger or activation.                                                      | Payment history, audit reason and permission, PaymentRejected and QuotaReleased, replay.                                          | validation_failed, state_conflict, or idempotency_conflict.                    |

Purchase has modeled later transitions ELIGIBLE to ALLOCATED to COMPLETED. No
Phase 1 command or W1-05 endpoint reaches them. Payment REFUNDED, saving,
giveaway, allocation, and refund detail remain deferred.

## Sohibul Qurban

| Command                         | Transition                      | Actor and permission   | Guard and effects                                                                                                                                                                                                                | History, audit, outbox, replay                                                                | Error                                   |
| ------------------------------- | ------------------------------- | ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- | --------------------------------------- |
| Activate from verified Purchase | none to ACTIVE                  | payment.verify command | In the same payment-verification transaction, create one ACTIVE record for each intended participant with a Party. The unique purchase and sequence pair prevents duplicate activation. The baseline PENDING status is not used. | Sohibul Qurban history from null to ACTIVE, parent audit, outbox, and replay commit together. | state_conflict or idempotency_conflict. |
| Replace or cancel               | ACTIVE to REPLACED or CANCELLED | admin.manage           | No Phase 1 endpoint. A future command locks ACTIVE participant and requires reason.                                                                                                                                              | Sohibul Qurban history, audit, outbox, replay.                                                | forbidden or state_conflict.            |

## Error and retry rules

State guards return state_conflict and capacity guards return quota_unavailable.
Missing or inactive credentials return unauthenticated or forbidden. Same
namespace, key, and request hash replays the stored response; another hash
returns idempotency_conflict. Request IDs are correlation only. Audit payloads
exclude raw Purchase tokens, session tokens, CSRF values, evidence bytes, and
storage credentials.
