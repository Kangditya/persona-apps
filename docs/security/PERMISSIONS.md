# Phase 1 Operations Permissions

## Authority

This document is the authoritative Phase 1 permission vocabulary and route
mapping for W1-05 OpenAPI and W1-06 runtime work. It applies ADR-019, ADR-024,
ADR-043, and ADR-044. Roles are provisioning conveniences; backend permission
checks are authoritative.

The Operations session stores an allowlisted permission snapshot. The API
authenticates the session, checks expiry and revocation, validates Origin and
CSRF for unsafe cookie requests, then authorizes the permission before
idempotency lookup or replay. Frontend route guards are presentation only.

## Vocabulary

| Permission       | Grants                                                  | Does not grant                                  |
| ---------------- | ------------------------------------------------------- | ----------------------------------------------- |
| event.read       | Read Event lists and details.                           | Event changes.                                  |
| event.manage     | Create and transition Events.                           | Offering changes or payment review.             |
| offering.read    | Read Offering lists and details.                        | Offering changes.                               |
| offering.manage  | Create and transition Offerings.                        | Event changes or payment review.                |
| purchase.read    | Read Operations Purchase views.                         | Purchase-token access or cancellation override. |
| payment.read     | Read Operations Payment views.                          | Payment verification or rejection.              |
| payment.verify   | Verify or reject submitted Payments.                    | Refund policy or ledger correction.             |
| participant.read | Read participant views.                                 | Participant replacement or cancellation.        |
| dashboard.read   | Read non-authoritative dashboard projections.           | Transactional commands.                         |
| audit.read       | Read minimized privileged audit records.                | Audit mutation.                                 |
| admin.manage     | Perform explicitly documented administration overrides. | A blanket bypass of lifecycle guards.           |

No participant.manage permission exists in Phase 1. A participant replacement or
cancellation command is deferred and requires admin.manage plus a reason until a
narrower permission is accepted.

## Role mapping

| Role                      | Permissions                                                                                                                                         |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| Event Operator            | event.read, event.manage, offering.read, offering.manage, purchase.read, participant.read, dashboard.read                                           |
| Finance Operator          | event.read, offering.read, purchase.read, payment.read, payment.verify, participant.read, dashboard.read                                            |
| Operations Manager        | event.read, event.manage, offering.read, offering.manage, purchase.read, payment.read, payment.verify, participant.read, dashboard.read, audit.read |
| Customer Service Operator | event.read, offering.read, purchase.read, payment.read, participant.read                                                                            |
| System Administrator      | All listed permissions                                                                                                                              |

Finance Operators can verify or reject Payment evidence. Customer Service can
read Purchase and Payment data but cannot verify, reject, administer, or read
audit records.

## W1-05 Operations route mapping

| Planned operation                                                                                        | Required permission or security rule                                                           |
| -------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| GET /api/operations/v1/auth/login                                                                        | None. Starts OIDC login and creates short-lived state, nonce, and PKCE context.                |
| GET /api/operations/v1/auth/callback                                                                     | None. Validates the bound OIDC callback before creating a session.                             |
| GET /api/operations/v1/auth/session                                                                      | Authenticated unrevoked session; no business permission. It returns or rotates the CSRF token. |
| POST /api/operations/v1/auth/logout                                                                      | Authenticated unrevoked session plus Origin and CSRF validation; no business permission.       |
| GET /api/operations/v1/events and GET /api/operations/v1/events/{event_id}                               | event.read                                                                                     |
| POST /api/operations/v1/events and PATCH /api/operations/v1/events/{event_id}                            | event.manage                                                                                   |
| POST /api/operations/v1/events/{event_id}/publish, activate, suspend, close, or archive                  | event.manage                                                                                   |
| GET /api/operations/v1/events/{event_id}/offerings and GET /api/operations/v1/offerings/{offering_id}    | offering.read                                                                                  |
| POST /api/operations/v1/events/{event_id}/offerings and PATCH /api/operations/v1/offerings/{offering_id} | offering.manage                                                                                |
| POST /api/operations/v1/offerings/{offering_id}/publish, unavailable, or archive                         | offering.manage                                                                                |
| GET /api/operations/v1/purchases and GET /api/operations/v1/purchases/{purchase_id}                      | purchase.read                                                                                  |
| GET /api/operations/v1/payments and GET /api/operations/v1/payments/{payment_id}                         | payment.read                                                                                   |
| POST /api/operations/v1/payments/{payment_id}/verify or reject                                           | payment.verify; rejection requires an audit reason.                                            |
| GET /api/operations/v1/participants and GET /api/operations/v1/participants/{participant_id}             | participant.read                                                                               |
| GET /api/operations/v1/dashboard/summary                                                                 | dashboard.read                                                                                 |
| GET /api/operations/v1/audit                                                                             | audit.read                                                                                     |

The four authentication operations are Operations routes but not business
commands. Mapping them to a business permission would prevent an eligible user
from starting or completing login. Their security boundary is the OIDC flow,
session validity, Origin, and CSRF contract in AUTHENTICATION.md.

## Storefront boundary

Storefront Event and Offering reads plus checkout are guest-accessible. Purchase
tracking, cancellation, and payment-evidence submission require the opaque
Purchase Bearer token. That token scopes one Purchase only; it is not an
Operations session, role, permission, or payment credential.

Public callers never access Operations routes, audit records, operator details,
permission snapshots, session/CSRF hashes, Purchase-token hashes, internal
evidence references, or object-storage fields.

## Audit and override rules

Every privileged lifecycle command writes actor, source, required permission,
request ID, action, target, and minimized before/after data in the owning
transaction. Payment rejection and every admin.manage override require a
non-empty reason. Audit failure aborts the privileged command.

Authorization always precedes idempotency replay. A guessed replay key must not
reveal an earlier response to an unauthenticated or unauthorized caller.

## Deferred

OIDC provider configuration, operator provisioning, custom roles, permission
administration UI, event/location scope evaluation, participant.write policy,
and session revocation administration remain future work. No local password,
role table, frontend authorization framework, or runtime handler is created by
this specification.
