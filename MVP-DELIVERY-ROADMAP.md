# Qurban Full Event-Day MVP Delivery Roadmap

## Status

Approved planning baseline as of 2026-08-24. This roadmap replaces the former
eight-week commerce-only scope in `TWO-MONTH-MVP.md`.

It is a planning artifact, not implementation evidence. A task becomes active
only after its header-only draft is revalidated, reviewed, and copied to
`.codex/TASK.md`.

## Objective

Deliver one traceable Qurban journey that can operate through a complete
3–4-day Eid al-Adha event:

```text
Event and Offering
→ Common Purchase and Payment
→ Sohibul Qurban Activation
→ Livestock Intake and Pen Assignment
→ Allocation
→ Slaughter Schedule, Queue, and Completion
→ Distribution
→ Customer Status, Evidence, and Documents
→ Realtime Multi-Team Operations on Mobile
```

## Planning Boundaries

- Preserve the completed W1-01 through W3-06 work.
- `COMMON` is the only implemented purchasing channel in this MVP roadmap;
  Saving and Giveaway remain deferred.
- The Go modular monolith and PostgreSQL remain authoritative.
- Storefront and Operations remain separate Next.js applications.
- Use polling first and SSE for high-value one-way updates. Do not add
  WebSocket, Redis, a broker, microservices, or a native mobile app without
  measured need and an accepted decision.
- Mobile delivery uses the responsive PWAs. Offline work is limited to
  explicitly allowlisted, non-financial field milestones with idempotent replay
  and visible conflict handling.
- Estimates are planning values, not Actual Hours.

## Requirement Coverage

| Requirement                                         | Delivery weeks | Release evidence                                                                  |
| --------------------------------------------------- | -------------- | --------------------------------------------------------------------------------- |
| 3–4-day Event execution                             | 7, 10, 15, 16  | Configured local execution days; 72–96-hour soak; complete rehearsal              |
| Slaughter-event support                             | 7, 10, 13–16   | Sessions, stations, queues, milestones, incidents, dashboard, UAT                 |
| Customer event-day journey                          | 5, 11, 12, 16  | Purchase-token timeline, schedule, status, distribution, notifications, documents |
| Livestock/Sohibul journey from pens to distribution | 8–12, 15–16    | Traceable livestock, allocation, slaughter, and distribution histories            |
| Several field teams and technical support           | 7, 10, 14–16   | Team/shift/station assignments, handovers, diagnostics, escalation, UAT           |
| Realtime mobile and degraded connectivity           | 13–16          | Projection, polling, SSE, mobile PWA, bounded offline replay, device matrix       |

## Delivery Plan

### Week 1 — Platform Safety — 32 hours — Completed

Freeze Phase 1 commerce rules, architecture, contracts, migrations, lifecycles,
permissions, and platform safety foundations.

### Week 2 — Event and Offering — 42 hours — Completed

Deliver public Event/Offering discovery and authenticated Operations
administration with persistence, audit, authorization, and tests.

### Week 3 — Party and Common Purchase — 50 hours

W3-01 through W3-06 are completed. Remaining work:

- W3-07 Build Storefront Direct Checkout Form and Confirmation.
- W3-08 Build Operations Purchase List and Detail.
- W3-09 Add Validation, Authorization, Idempotency, and Quota Contention Tests.

Exit: a public user can create one pending Common Purchase, and authorized
operators can inspect it without database access.

### Week 4 — Payment and Activation — 48 hours

- W4-01 Implement Append-Oriented Payment Submission and Evidence Metadata.
- W4-02 Store Evidence Through an Approved Storage Adapter or Safe MVP Mechanism.
- W4-03 Implement Verification and Rejection Commands.
- W4-04 Implement Purchase Eligibility Transition.
- W4-05 Implement Idempotent Sohibul Qurban Activation.
- W4-06 Add Operations Payment Verification Queue and Detail.
- W4-07 Add Storefront Payment Submission and Status Views.
- W4-08 Record Privileged Actions in Append-Only Audit History.
- W4-09 Add Duplicate Submission, Verification, Authorization, and Activation Tests.

Exit: payment decisions, eligibility, quota consumption/release, and Sohibul
Qurban activation commit exactly once with history, audit, and replay safety.

### Week 5 — Storefront Commerce Journey — 36 hours

- W5-01 Complete Mobile-First Event and Offering Pages.
- W5-02 Complete Direct Checkout and Participant Forms.
- W5-03 Add Explicit Loading, Validation, Empty, Error, Stale, and Conflict States.
- W5-04 Complete Payment Instructions and Evidence Submission.
- W5-05 Add Purchase Tracking by Safe Reference and Authentication Policy.
- W5-06 Add Accessibility Checks and Responsive Behavior.
- W5-07 Add Route, Component, API Client, and Critical Flow Tests.

Exit: a public user completes the commerce journey on representative mobile
sizes without exposing or caching sensitive data.

### Week 6 — Operations Commerce Control Plane — 44 hours

- W6-01 Complete Operator Login and Route Access Behavior.
- W6-02 Complete Event and Offering Administration.
- W6-03 Complete Purchase Search, Filters, Detail, and Exception Display.
- W6-04 Complete Payment Verification Queue and Decisions.
- W6-05 Add Sohibul Qurban List and Detail.
- W6-06 Add Dashboard Metrics for Purchases, Quota, Payment Backlog, Eligibility, and Activation.
- W6-07 Use Polling for Dashboard Freshness Where Needed.
- W6-08 Display Projection or Query Freshness and Errors Explicitly.
- W6-09 Add Operations Workflow and Permission Tests.

Exit: Operations can run the authoritative Common Purchase lifecycle without
direct database access.

### Week 7 — Event-Day and Team Foundations — 42 hours

- W7-01 Implement the 3–4-Day Event Execution Calendar and Timezone Rules.
- W7-02 Implement Event-Scoped Field Teams and Membership.
- W7-03 Implement Shifts, Station Assignments, and Handover Records.
- W7-04 Implement Event Readiness and Participant or Livestock Check-In.
- W7-05 Implement Operational Incidents and Support Escalation.
- W7-06 Add Event-Day Operations APIs, Permissions, Audit, and Idempotency.
- W7-07 Build Operations Event-Day Setup, Readiness, Team, and Incident Screens with Tests.

Exit: one Event can configure exactly three or four local execution days and
coordinate authorized teams, shifts, readiness, check-in, handover, and
incidents.

### Week 8 — Livestock and Pen Lifecycle — 41 hours

- W8-01 Reconcile the Existing Operational Schema and Plan Only Additive Livestock Changes.
- W8-02 Implement Livestock Registry, Identity, Classification, and Source Tracking.
- W8-03 Implement Inspection, Health, Weight, and Readiness Transitions.
- W8-04 Implement Pen or Location Assignment and History.
- W8-05 Add Livestock Operations APIs, Permissions, Audit, and Safe Batch Queries.
- W8-06 Build Mobile-First Livestock Intake, Inspection, Pen, and Scan-Fallback Screens.
- W8-07 Add Livestock Validation, Concurrency, Authorization, Audit, and Mobile Flow Tests.

Exit: every livestock unit is traceable from intake through readiness and its
current/historical pen or location.

### Week 9 — Allocation and Manifests — 50 hours

- W9-01 Implement the Allocation Lifecycle and Capacity Policy.
- W9-02 Allocate Eligible Purchases and Sohibul Qurban to Livestock.
- W9-03 Enforce Shared Livestock Capacity Under Concurrency.
- W9-04 Implement Allocation Release, Reassignment, Versioning, and Audit.
- W9-05 Build Allocation and Team or Station Manifests.
- W9-06 Add Allocation Operations APIs and Permission Boundaries.
- W9-07 Build Mobile-First Allocation, Reassignment, and Manifest Screens.
- W9-08 Add Allocation Capacity, Contention, Reassignment, Authorization, and Audit Tests.

Exit: eligible Sohibul Qurban and Purchases are allocated without exceeding
capacity, and all release/reassignment history remains auditable.

### Week 10 — Slaughter Execution — 56 hours

- W10-01 Implement Slaughter Sessions and Schedules Across Event Execution Days.
- W10-02 Implement Slaughter Stations with Team and Shift Assignments.
- W10-03 Implement Participant and Livestock Check-In and Queue Admission.
- W10-04 Implement Atomic Queued, Called, In-Progress, Held, Completed, and Cancelled Transitions.
- W10-05 Implement Configurable Sohibul Qurban Self, Proxy, and No-Attendance Flow.
- W10-06 Integrate Slaughter Incidents, Support Escalation, and Shift Handover.
- W10-07 Add Slaughter Operations APIs, Permissions, Audit, Idempotency, and Conflict Handling.
- W10-08 Build Mobile-First Slaughter Queue, Station, Scan, and Quick-Action Screens.
- W10-09 Add Slaughter Queue, Multi-Team, Attendance, Conflict, Audit, and Mobile Tests.

Exit: multiple authorized teams can safely execute and hand over slaughter work
through all three or four days without duplicate or invalid transitions.

### Week 11 — Distribution — 51 hours

- W11-01 Implement Sohibul Entitlement and Portion Rules.
- W11-02 Implement Beneficiary Assignment and Privacy Boundaries.
- W11-03 Implement Portion Preparation and Distribution Readiness.
- W11-04 Implement Pickup, Delivery, Collection Confirmation, and Proof.
- W11-05 Implement Distribution Failure, Reassignment, Exception, and Completion Transitions.
- W11-06 Add Distribution Operations APIs, Permissions, Audit, and Idempotency.
- W11-07 Build Mobile-First Operations Distribution and Safe Storefront Status Screens.
- W11-08 Add Entitlement, Beneficiary, Pickup, Delivery, Proof, Privacy, and Concurrency Tests.

Exit: each applicable Sohibul entitlement or beneficiary portion is traceable
through preparation and pickup/delivery completion without public data leakage.

### Week 12 — Customer Event-Day Journey — 40 hours

- W12-01 Publish Event-Day Schedule and Instructions to Storefront.
- W12-02 Add Purchase-Token-Scoped Event Timeline and Current Status.
- W12-03 Add Safe Attendance, Queue, Slaughter, and Exception Status.
- W12-04 Add Safe Distribution Entitlement, Pickup, Delivery, and Completion Status.
- W12-05 Implement Deduplicated Event-Day Notifications and Delivery Attempts.
- W12-06 Add Minimal Completion Evidence, Receipt, and Certificate Documents.
- W12-07 Add Mobile, Accessibility, Privacy, Token-Scope, Notification, and Document Tests.

Exit: the customer sees a privacy-safe, current journey from schedule through
slaughter and distribution on mobile without internal operational fields.

### Week 13 — Realtime Operations Projection — 46 hours

- W13-01 Harden the PostgreSQL Outbox Worker, Retry, and Backlog Monitoring.
- W13-02 Implement Rebuildable Event-Day Projections and Checkpoints.
- W13-03 Add Dashboard Read Models for Readiness, Livestock, Allocation, Slaughter, Distribution, and Incidents.
- W13-04 Add Bounded Polling with Explicit Freshness and Error States.
- W13-05 Add Server-Sent Events with Reconnection, Last-Event-ID, and Missed-Event Recovery.
- W13-06 Build Realtime Operations Dashboard and Team-Lane Views.
- W13-07 Add Projection Rebuild, Worker Retry, Polling, SSE Reconnection, Authorization, and Lag Tests.

Exit: dashboards are generally under the approved freshness target, expose lag,
and can rebuild from authoritative records without becoming command truth.

### Week 14 — Mobile and Degraded Field Operation — 45 hours

- W14-01 Define and Verify the Supported Mobile Device and Browser Matrix.
- W14-02 Enforce the Offline Field-Command Allowlist and Online-Only Safety Boundary.
- W14-03 Implement a Bounded Non-Sensitive Device-Local Field Command Queue.
- W14-04 Implement Idempotent Replay, Conflict Resolution, and Operator Confirmation.
- W14-05 Add Connectivity, Pending-Sync, Failure, and Recovery Visibility.
- W14-06 Add Native QR or Barcode Detection with Manual Code Entry Fallback.
- W14-07 Add Field Diagnostics, Support Status, and Escalation Handoff.
- W14-08 Add Low-Bandwidth, Disconnect, Replay, Multi-Device, Scanner-Fallback, and Security Tests.

Exit: representative mobile devices remain usable under weak connectivity;
queued field-safe actions reconcile visibly, while sensitive/financial commands
remain online-only.

### Week 15 — Integration, Security, and Event Resilience — 52 hours

- W15-01 Run Complete Backend, Frontend, Contract, Migration, and Integration Suites.
- W15-02 Add End-to-End Purchase-to-Distribution Scenarios.
- W15-03 Test Concurrent Multi-Team Slaughter Execution and Shift Handover.
- W15-04 Test Quota, Allocation, Queue, Offline Replay, and Stale-Version Contention.
- W15-05 Verify Public and Internal Data Separation, Permissions, and Audit Completeness.
- W15-06 Test Backup, Restore, Projection Rebuild, and Worker Recovery.
- W15-07 Add Operational Metrics and Alerts and Run Load Plus 72–96-Hour Soak Tests.
- W15-08 Fix Critical and High-Severity Defects.

Exit: the full system survives contested operations, reconnects, worker failure,
restore, projection rebuild, peak load, and a continuous event-duration soak
with no unresolved critical security or data-integrity defect.

### Week 16 — Staging, UAT, Rehearsal, and Pilot — 48 hours

- W16-01 Deploy the Full Event-Day Release Candidate to Staging.
- W16-02 Seed Representative Non-Sensitive Commerce and Event-Day Data.
- W16-03 Run Customer, Finance, Livestock, Allocation, Slaughter, Distribution, Manager, and Support UAT.
- W16-04 Rehearse the Complete 3–4-Day Event Scenario and Shift Handovers.
- W16-05 Verify the Mobile Device, Browser, Scanner, and Connectivity Matrix in Staging.
- W16-06 Test Rollback, Restore, Projection Recovery, and Incident Procedures.
- W16-07 Prepare Role-Specific Runbooks, Support Escalation, Release Checklist, and Known Limitations.
- W16-08 Fix Release-Blocking Defects Only.
- W16-09 Conduct a Controlled Pilot and Record Go or No-Go Evidence.

Exit: the complete event-day journey passes UAT and rehearsal, operators have
usable procedures, recovery paths are proven, and a controlled pilot produces a
recorded go/no-go decision.

## Total Planned Effort

| Weeks | Planned hours |
| ----- | ------------: |
| 1–6   |           252 |
| 7–16  |           471 |
| Total |           723 |

Actual Hours remain blank until measured. Reforecast estimates after each week;
never infer actual time from status, commits, or elapsed calendar time.

## Definition of Done

The MVP is complete only when all of the following are evidence-backed:

- the Common Purchase, Payment, and Sohibul Qurban lifecycle is complete;
- one Event configures exactly three or four local execution days;
- field teams, shifts, stations, handovers, incidents, and support escalation
  are event-scoped and permission-controlled;
- livestock is traceable through intake, inspection, readiness, pen/location,
  allocation, slaughter, and distribution;
- customers can safely track their own event-day journey;
- dashboards use rebuildable projections and expose freshness/lag;
- polling and SSE recover without becoming transactional truth;
- representative mobile devices support the critical field workflows;
- bounded offline field actions replay idempotently and conflicts remain visible;
- tests, security, audit, backup/restore, load, 72–96-hour soak, UAT, rehearsal,
  runbooks, and controlled pilot evidence pass;
- no unresolved critical security, privacy, financial, allocation, slaughter,
  distribution, or data-integrity defect remains.
