# Full Event-Day Operations Lifecycles

## Authority

This document is the approved planning baseline for the Full Event-Day MVP. It
applies ADR-051 through ADR-055. It defines required lifecycle behavior but does
not claim that runtime handlers or migrations exist.

Every accepted field command authenticates and authorizes first, verifies Event
and assignment scope, locks or conditionally updates authoritative rows,
validates lifecycle/version/idempotency, writes state/history/audit/outbox/replay
in one transaction, and commits once.

## Event Execution Day

```text
PLANNED → READY → OPEN → IN_PROGRESS → COMPLETED
                  └───────────────→ SUSPENDED → IN_PROGRESS
PLANNED | READY → CANCELLED
```

- One Event has exactly three or four ordered local execution dates in its IANA
  timezone.
- Opening requires configured operating windows, responsible teams, and
  readiness checks.
- Completion requires no unresolved blocking work or an authorized, reasoned
  exception/handover.
- Shift/day completion does not close the Event aggregate.

## Field Team, Shift, and Assignment

```text
Team:       ACTIVE ↔ SUSPENDED → CLOSED
Shift:      PLANNED → OPEN → HANDED_OVER | COMPLETED | CANCELLED
Assignment: PLANNED → ACTIVE → RELEASED | REASSIGNED
```

- Membership and assignments are Event-scoped.
- Opening a shift snapshots its responsible members and assignments.
- Handover records outgoing/incoming owner, unfinished work, incidents, and
  acknowledgement.
- Permission plus active assignment is required for assigned field commands.

## Readiness and Check-In

```text
PENDING → READY | NOT_READY | WAIVED
NOT_READY → READY
Check-in: EXPECTED → PRESENT | PROXY_PRESENT | ABSENT | CANCELLED
```

- A waiver requires a privileged permission, reason, and audit record.
- Sohibul attendance mode (`SELF`, `PROXY`, `NONE`) is separate from payment
  eligibility and activation.

## Livestock

```text
REGISTERED → INSPECTED → READY → ALLOCATED → QUEUED → SLAUGHTERED
      └──────────────→ HELD ↔ INSPECTED | READY
REGISTERED | INSPECTED | READY | HELD → CANCELLED
```

- Inspection, readiness, weight, pen/location, and status histories are
  append-oriented.
- Only `READY` livestock may receive a new confirmed Allocation.
- Location changes release the prior current assignment and create the next
  history record atomically.

## Allocation

```text
PROVISIONAL → CONFIRMED → RELEASED
      └───────────────→ REASSIGNED → CONFIRMED | RELEASED
```

- Only eligible Purchases/Sohibul Qurban may be allocated.
- Livestock participant capacity is never exceeded, including under concurrent
  allocation or reassignment.
- Release/reassignment requires reason, history, version protection, and audit.

## Slaughter Session and Record

```text
Session: PLANNED → OPEN → IN_PROGRESS → COMPLETED
         PLANNED | OPEN → CANCELLED

Record:  QUEUED → CALLED → IN_PROGRESS → COMPLETED
         QUEUED | CALLED | IN_PROGRESS → HELD → QUEUED | CALLED | CANCELLED
```

- Queue admission requires Event-day, livestock readiness, Allocation, and
  applicable attendance/check-in guards.
- Queue position, station, and team/shift assignment changes are versioned and
  conflict-protected.
- Completion records responsible operator/team, timestamps, applicable
  attendance/proxy outcome, history, audit, and outbox effect.

## Distribution

```text
PENDING → PREPARED → READY → COLLECTED | DELIVERED
                └────────→ FAILED → READY | CANCELLED
```

- A record explicitly targets a Sohibul entitlement or beneficiary portion and
  declares `PICKUP` or `DELIVERY`.
- Readiness requires applicable slaughter/preparation guards.
- Collection/delivery records required proof and responsible operator/team.
- Failure, reassignment, cancellation, and completion require history and
  auditable reason/evidence metadata.

## Operational Incident

```text
OPEN → ACKNOWLEDGED → IN_PROGRESS → RESOLVED → CLOSED
                        └────────→ ESCALATED → IN_PROGRESS
OPEN | ACKNOWLEDGED | IN_PROGRESS | ESCALATED → CANCELLED
```

- Severity, affected Event/day/team/work, owner, escalation, resolution, and
  handover are explicit.
- Closing a parent shift/day requires blocking incidents to be resolved,
  cancelled, or explicitly handed over.

## Projection and Delivery

```text
Outbox:      PENDING → PUBLISHED
                    └→ RETRYING → PUBLISHED | DEAD_LETTER
Projection:  CURRENT ↔ LAGGING → REBUILDING → CURRENT
SSE client:  CONNECTING → LIVE → RECONNECTING → LIVE | POLLING_FALLBACK
```

- Projection state is not command truth.
- Workers use durable checkpoints and idempotent event application.
- Dashboards expose source timestamp, lag, and errors.
- SSE reconnects from `Last-Event-ID` or an equivalent durable cursor; polling
  remains the fallback.

## Degraded-Connectivity Field Replay

Allowlisted offline milestones are limited to non-financial field operations
whose payload contains no unnecessary personal data or credential material.

```text
LOCAL_PENDING → REPLAYING → APPLIED
                     └───→ CONFLICT | REJECTED | EXPIRED
CONFLICT → OPERATOR_CONFIRMED_RETRY | DISCARDED
```

Replay re-runs authentication, permission, Event/team/shift assignment,
lifecycle, version, capacity, and idempotency checks. Local queue state never
grants authority. Payment, evidence, identity, authorization, Event/team
configuration, allocation-capacity overrides, and other sensitive commands are
online-only.

## Error and Retry Rules

- Invalid lifecycle or stale version returns a stable conflict error.
- Capacity failure is distinct from validation, authorization, and dependency
  failure.
- Same namespace/key/semantic request replays the stored outcome; another
  request conflicts.
- Audit/outbox/history/replay failure aborts the owning command transaction.
- No response, log, audit, outbox, projection, support diagnostic, or offline
  queue contains raw session/CSRF/Purchase tokens, credentials, evidence bytes,
  or unnecessary participant/beneficiary data.
