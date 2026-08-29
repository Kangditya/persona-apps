# Task: W1-04 Define Lifecycles and Permissions

## Status

Inactive draft. Activate only after W1-01 is executed and current source is
revalidated.

## Tracker

- Workstream: Platform Safety
- Estimate: 4 hours
- Dependencies: W1-01
- Tracker objective: Define status transitions and the permission matrix.

## Objective

Provide one authoritative Phase 1 lifecycle specification and least-privilege
permission matrix for OpenAPI and runtime implementation.

## Existing Foundation to Reuse

- The PRD requires aggregate-owned lifecycles rather than one global status.
- Existing SQL already defines statuses for Event, Offering, Purchase, Payment,
  and Sohibul Qurban.
- ADR-019 requires permission-based backend authorization.
- ADR-024 and ADR-041 define audit and idempotency behavior around commands.

## Required Lifecycle Specification

- Event:
  `DRAFT -> PUBLISHED -> ACTIVE <-> SUSPENDED -> CLOSED -> ARCHIVED`.
- Offering:
  `DRAFT -> PUBLISHED -> UNAVAILABLE -> PUBLISHED` and any non-archived state
  may become `ARCHIVED`; only active-event published offerings are public.
- Quota Reservation: `RESERVED -> CONSUMED | RELEASED | EXPIRED`; terminal
  states cannot reopen. Reacquisition creates the Purchase's next numbered
  reservation attempt while prior attempts remain immutable.
- Purchase: `DRAFT -> PENDING_PAYMENT -> PAID -> ELIGIBLE -> ALLOCATED ->
  COMPLETED`; `DRAFT` or `PENDING_PAYMENT` may become `CANCELLED`. Payment
  rejection does not itself cancel the Purchase.
- Payment: `SUBMITTED -> VERIFIED | REJECTED`; later refund behavior is outside
  the common-purchase MVP even though the baseline schema contains `REFUNDED`.
- Sohibul Qurban: activation creates `ACTIVE`; `ACTIVE -> REPLACED | CANCELLED`.
  The baseline `PENDING` state is not used by automatic Phase 1 activation.
- Every transition specification names actor, permission, preconditions,
  transaction effects, history, audit, outbox, idempotency, and conflict cases.

## Permission Vocabulary

Use these stable permission strings:

```text
event.read
event.manage
offering.read
offering.manage
purchase.read
payment.read
payment.verify
participant.read
dashboard.read
audit.read
admin.manage
```

## Role Matrix

| Role | Permissions |
| --- | --- |
| Event Operator | `event.read`, `event.manage`, `offering.read`, `offering.manage`, `purchase.read`, `participant.read`, `dashboard.read` |
| Finance Operator | `event.read`, `offering.read`, `purchase.read`, `payment.read`, `payment.verify`, `participant.read`, `dashboard.read` |
| Operations Manager | All except `admin.manage` |
| Customer Service Operator | `event.read`, `offering.read`, `purchase.read`, `payment.read`, `participant.read` |
| System Administrator | All listed permissions |

Additional rules:

- Roles are provisioning conveniences; permissions are authoritative.
- Storefront purchase tokens are Purchase-scoped credentials, not roles.
- Public users cannot call operations endpoints or view audit, operator,
  permission, internal evidence-reference, or storage fields.
- Backend application policies enforce permissions; frontend guards are only
  navigation and presentation controls.
- Payment verification/rejection requires `payment.verify` and an audit reason
  for rejection or override.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `docs/domain/COMMERCE_LIFECYCLES.md` | Create | Define transitions, guards, effects, and failure cases. |
| `docs/security/PERMISSIONS.md` | Create | Define permission vocabulary and role mapping. |
| `docs/PRD.md` | Modify | Replace illustrative Phase 1 statuses with resolved rules. |
| `docs/ARCHITECTURE.md` | Modify | Link authoritative lifecycle and authorization specifications. |
| `docs/DECISIONS.md` | Modify | Accept lifecycle and permission decisions where required. |
| `.codex/CURRENT_STATE.md` | Modify | Record specifications as documented, not implemented. |
| `.codex/TASK.md` | Create, execute, archive | Preserve the activated task and final review. |

## Acceptance Criteria

1. Each Phase 1 aggregate has an explicit transition table with command actor,
   permission, guard, effects, and errors.
2. No transition bypasses history, audit, quota, or idempotency requirements.
3. Every operations endpoint planned in W1-05 maps to a listed permission.
4. The role matrix grants Finance payment verification while keeping public
   and Customer Service access non-privileged.
5. Storefront token scope and backend authority are explicit.
6. No auth runtime, role-management UI, or domain-module shell is created.

## Verification

- Cross-check status names against migration SQL and W1-03's additive delta.
- Trace payment verification and rejection across Payment, Purchase, quota,
  participant, audit, and outbox transitions.
- Cross-check every W1-05 operations endpoint against one permission.
- Run `make format-check` and `make validate`.

## Risks and Deferred Work

- Operator provisioning, custom roles, event-scoped permissions, and permission
  administration UI remain deferred.
- Refund, replacement, allocation, saving, and giveaway transition details are
  outside Phase 1 unless required by current schema integrity.

## Final Review Requirements

Distinguish documented lifecycle policy, verified matrix coverage, assumptions,
deferred permissions, remaining security risks, and the next ready task.
