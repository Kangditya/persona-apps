# Product Requirements Document

## Qurban Commerce and Operations Platform

**Status:** Initial product baseline  
**Document owner:** Product and Engineering  
**Repository:** `Kangditya/persona-apps`  
**Primary applications:** `storefront-web`, `operations-web`, `api`

---

## 1. Document Purpose

This document defines the initial product boundaries for an annual qurban commerce and operations platform.

The platform is intended to support:

- public purchasing journeys;
- Sohibul Qurban registration as an outcome of purchasing;
- annual event preparation;
- livestock allocation and operational tracking;
- slaughter-day execution;
- distribution and post-event reporting.

This PRD is intentionally flexible. Detailed workflows, validations, user interfaces, integrations, and operational rules may be refined through discovery, Figma design, field observations, and stakeholder feedback without changing the core product direction.

---

## 2. Product Summary

The platform coordinates qurban purchasing and annual event operations for approximately:

- up to **1,000 Sohibul Qurban** per event;
- potentially **thousands of livestock units, allocations, operational records, and distribution records**;
- multiple purchasing and funding models;
- public users, internal operators, field teams, finance teams, and event administrators.

The platform consists of two principal user-facing applications:

1. **Storefront Web**  
   Public-facing application for exploring qurban offerings, purchasing, saving plans, giveaway participation, payment, and purchase tracking.

2. **Operations Web**  
   Internal application for managing purchasing, participants, payments, livestock, event preparation, slaughter execution, distribution, and reporting.

Both applications are supported by a shared backend domain model. The initial implementation may use a single modular Go API while preserving boundaries that allow selected modules or APIs to be separated later.

---

## 3. Product Problem

Annual qurban programs commonly depend on disconnected spreadsheets, messaging applications, manual payment verification, paper records, and operator knowledge.

At larger scale, this creates several risks:

- purchasing records are inconsistent across channels;
- the purchaser, payer, sponsor, account holder, and Sohibul Qurban may be incorrectly treated as the same person;
- installment balances and purchase conversion are difficult to reconcile;
- livestock allocation is not traceable;
- field teams lack a single operational status;
- slaughter queues and execution updates are delayed;
- distribution records are incomplete;
- event reporting requires manual reconstruction;
- participants cannot reliably view their latest purchase and event status.

The product must provide one source of operational truth while allowing different purchasing journeys.

---

## 4. Product Goals

### 4.1 Primary Goals

- Support three purchasing channels:
  - common purchasing;
  - saving purchasing;
  - giveaway purchasing.
- Produce a valid qurban purchase and Sohibul Qurban record from each eligible channel.
- Provide internal teams with near-real-time operational visibility.
- Track the relationship between purchases, participants, livestock, allocations, payments, slaughter execution, and distribution.
- Preserve an auditable history for financial and operational changes.
- Support recurring annual events without rebuilding the system each year.

### 4.2 Secondary Goals

- Reduce manual reconciliation.
- Reduce duplicate participant and purchase records.
- Improve payment and saving-plan transparency.
- Improve field coordination during the slaughter event.
- Provide data suitable for management reporting.
- Establish stable domain and API contracts for future mobile applications, kiosks, integrations, or partner channels.

### 4.3 Non-Goals for the Initial Release

Unless later prioritized, the initial release does not require:

- a public marketplace with multiple independent vendors;
- generalized livestock trading;
- full accounting or enterprise resource planning;
- payroll and volunteer management;
- route optimization for meat delivery;
- advanced customer relationship management;
- fully automated bank settlement;
- multi-country taxation;
- microservices from day one.

---

## 5. Users and Roles

| User                          | Primary Needs                                                              |
| ----------------------------- | -------------------------------------------------------------------------- |
| Visitor                       | View qurban programs, packages, prices, schedules, and general information |
| Purchaser                     | Create and pay for a common purchase                                       |
| Saving Account Holder         | Create a saving plan, make installments, and monitor progress              |
| Giveaway Applicant or Nominee | Submit or receive giveaway participation                                   |
| Sponsor                       | Fund giveaway purchases or programs                                        |
| Sohibul Qurban                | Be registered against an eligible purchase and view event information      |
| Customer Service Operator     | Assist users, correct permitted data, and resolve purchasing issues        |
| Finance Operator              | Verify payments, reconcile installments, and review financial status       |
| Livestock Operator            | Register, inspect, group, and allocate livestock                           |
| Event Operator                | Manage event readiness, check-in, queues, and field status                 |
| Slaughter Team                | Receive assigned livestock and record execution milestones                 |
| Distribution Team             | Record portions, beneficiaries, pickup, or delivery completion             |
| Operations Manager            | Monitor capacity, exceptions, throughput, and completion                   |
| System Administrator          | Manage users, roles, configuration, and audit access                       |

A person may hold more than one role. A purchaser, payer, sponsor, saving account holder, and Sohibul Qurban must remain separate concepts even when they refer to the same person.

---

## 6. Core Domain Concepts

### 6.1 Qurban Event

Represents one annual operational cycle.

Typical information:

- event name and year;
- registration and purchasing period;
- event location;
- slaughter dates;
- operational capacity;
- participant quota;
- status;
- policy and configuration snapshot.

### 6.2 Purchasing Channel

The method through which a qurban purchase is initiated or funded.

Initial values:

- `COMMON`
- `SAVING`
- `GIVEAWAY`

### 6.3 Purchase

The commercial or funding record that produces an entitlement for one or more Sohibul Qurban.

A purchase may include:

- purchaser;
- payer;
- channel;
- qurban event;
- package or offering;
- amount;
- payment state;
- participant records;
- livestock allocation;
- lifecycle status.

### 6.4 Sohibul Qurban

A person whose qurban intention is represented by an eligible purchase or allocation.

Sohibul Qurban registration is not itself a purchasing channel. It is an outcome or participant assignment created after the applicable purchasing requirements are met.

### 6.5 Saving Account

A funding plan that accumulates installments toward a target qurban offering or amount.

A saving account becomes a qurban purchase only after conversion criteria are satisfied.

### 6.6 Giveaway Program

A funded program that identifies, selects, or assigns a recipient as Sohibul Qurban.

The sponsor or payer may be different from the recipient.

### 6.7 Livestock

A physical animal tracked through acquisition or intake, health and quality checks, grouping, location, allocation, slaughter, and completion.

### 6.8 Allocation

The relationship between eligible purchases or Sohibul Qurban and livestock.

The model must support:

- one participant allocated to one eligible livestock unit;
- multiple participants sharing one eligible livestock unit;
- package rules that determine allocation capacity;
- reassignment with audit history.

---

## 7. Purchasing Channels

## 7.1 Common Purchasing

### Objective

Allow a purchaser to select an available qurban offering, submit participant details, and complete payment.

### Baseline Flow

1. User selects an event and offering.
2. User provides purchaser and participant information.
3. System validates availability and quota.
4. System creates an order or pending purchase.
5. Payment is initiated or evidence is submitted.
6. Payment is verified or confirmed.
7. Purchase becomes eligible.
8. Sohibul Qurban records are activated.
9. Livestock allocation occurs immediately or later.

### Required Capabilities

- offering catalogue;
- price and availability display;
- direct checkout;
- participant entry;
- payment instruction;
- payment status;
- purchase status tracking;
- receipt or proof;
- cancellation and expiry rules;
- operator-assisted purchase.

### Phase 1 Common-Purchase Rules

- The Storefront exposes only the active event. Event lifecycle is
  `DRAFT -> PUBLISHED -> ACTIVE <-> SUSPENDED -> CLOSED -> ARCHIVED`, with at
  most one active event.
- An MVP Offering is an event-scoped sellable package, share, or category, not
  a physical Livestock record.
- One Purchase selects exactly one Offering. Phase 1 has no Shopping Cart or
  purchase-item aggregate.
- Checkout snapshots the Offering, price, currency, participant capacity, and
  intended participant names.
- Quota uses participant units against both Event and Offering limits. Checkout
  atomically reserves units for 24 hours.
- Submitted payment evidence pauses reservation expiry until review.
  Activation consumes quota; expiry, cancellation, or rejection releases it.
  Evidence resubmission after release must reacquire quota atomically.
- Common-purchase payment evidence is an append-oriented object-storage
  reference. JPEG, PNG, and PDF are accepted up to 10 MiB; PostgreSQL stores
  metadata and SHA-256, not file bytes.
- Evidence must declare the exact outstanding amount. Underpayment and
  overpayment submissions are rejected; balance policy remains deferred.
- Authorized Finance or Operations Managers verify or reject evidence.
  Successful verification atomically marks Payment verified, records Purchase
  `PAID` then `ELIGIBLE`, consumes quota, activates each Sohibul Qurban once,
  and writes audit and outbox effects.

---

## 7.2 Saving Purchasing

### Objective

Allow a user to accumulate funds toward a qurban target and convert the completed balance into a purchase.

### Baseline Flow

1. User selects or defines an eligible target.
2. System creates a saving account.
3. User makes one or more installments.
4. Finance verifies or reconciles installments.
5. System calculates remaining balance.
6. Account reaches conversion eligibility.
7. User or operator confirms the target offering when required.
8. Saving account is converted into a purchase.
9. Sohibul Qurban records are activated.

### Baseline Statuses

- `DRAFT`
- `ACTIVE`
- `PARTIALLY_FUNDED`
- `FULLY_FUNDED`
- `CONVERSION_PENDING`
- `CONVERTED`
- `CANCELLED`
- `EXPIRED`

### Required Capabilities

- saving target;
- installment schedule or flexible deposit mode;
- installment ledger;
- balance calculation;
- target price change policy;
- overpayment and underpayment handling;
- conversion workflow;
- refund or transfer policy;
- reminders;
- operator correction with audit trail.

---

## 7.3 Giveaway Purchasing

### Objective

Allow sponsors or program owners to fund qurban purchases for selected recipients.

### Baseline Flow

1. Operator creates a giveaway program.
2. Sponsor funding or budget is recorded.
3. Applicants, nominees, or recipients are registered.
4. Eligibility or selection is processed.
5. Recipient is approved.
6. System creates or assigns a funded purchase.
7. Recipient becomes Sohibul Qurban.
8. Livestock allocation occurs.
9. Sponsor and recipient receive the appropriate reporting.

### Required Capabilities

- giveaway program definition;
- sponsor information;
- funding ceiling;
- recipient intake;
- eligibility criteria;
- selection and approval;
- duplicate recipient prevention;
- purchase issuance;
- sponsor reporting;
- privacy controls.

---

## 8. Storefront Web Scope

The Storefront Web is the public interaction channel.

### 8.1 Initial Modules

- event landing page;
- qurban offering catalogue;
- common purchasing;
- saving account registration;
- installment submission or payment;
- giveaway application or recipient flow;
- purchaser and participant profile;
- payment status;
- purchase tracking;
- event schedule and instructions;
- notifications;
- receipt, certificate, or post-event documentation.

### 8.2 Storefront Principles

- mobile-first;
- accessible without operator assistance for normal journeys;
- clear distinction between purchaser and Sohibul Qurban;
- resumable multi-step flows;
- explicit payment and eligibility statuses;
- no disclosure of internal operational data;
- graceful handling of unavailable quota or offerings.

---

## 9. Operations Web Scope

The Operations Web is the internal operational control plane.

### 9.1 Initial Modules

- real-time overview dashboard;
- event configuration;
- purchasing management;
- saving-account management;
- giveaway-program management;
- participant and Sohibul Qurban management;
- payment verification and reconciliation;
- offering and quota management;
- livestock registry;
- inspection and readiness;
- pen or location assignment;
- participant-to-livestock allocation;
- event check-in;
- slaughter queue and execution;
- distribution;
- incidents and exceptions;
- reporting;
- user, role, and audit administration.

### 9.2 Dashboard Baseline

The dashboard should support near-real-time visibility into:

- purchasing volume by channel;
- payment and verification backlog;
- active and fully funded saving accounts;
- giveaway funding and assignment;
- Sohibul Qurban quota usage;
- livestock availability and allocation;
- event readiness;
- slaughter queue and throughput;
- operational exceptions;
- distribution completion.

Dashboard metrics must be derived from authoritative transactional records, not manually maintained counters.

---

## 10. Cross-Channel Business Rules

1. Every purchase belongs to exactly one purchasing channel.
2. Sohibul Qurban may only become active when the relevant channel's eligibility conditions are met.
3. Purchaser, payer, sponsor, saving account holder, applicant, recipient, and Sohibul Qurban must be modeled independently.
4. A purchase may contain one or more participant records, subject to offering rules.
5. Quota reservation and quota consumption must be explicitly defined.
6. A saving account is not automatically a completed purchase.
7. A giveaway application is not automatically an approved purchase.
8. Payment records must be append-only where possible; corrections require traceability.
9. Livestock allocation must not exceed configured livestock capacity.
10. Reallocation, cancellation, refund, and manual override actions require authorization and audit events.
11. Event-specific rules must be configurable or versioned so future events do not retroactively modify historical records.
12. Financial amounts must use integer minor units or exact decimal types; floating-point arithmetic is prohibited.

---

## 11. Functional Requirements

### 11.1 Event Management

- Create and manage annual qurban events.
- Configure registration, payment, allocation, slaughter, and distribution windows.
- Configure participant quota and channel-specific capacity.
- Activate, suspend, close, and archive events.
- Preserve historical event configuration.

### 11.2 Offering and Pricing

- Define qurban offerings and package rules.
- Define species, category, sharing capacity, and pricing.
- Associate offerings with an event.
- Control publication and availability.
- Preserve the price applied to an existing purchase.

### 11.3 Party and Participant Management

- Maintain person and organization identities.
- Allow channel-specific roles without duplicating identities.
- Detect likely duplicates.
- Record participant names for ceremonial and reporting usage.
- Apply privacy and access controls.

### 11.4 Purchase Management

- Create purchases from all supported channels.
- Validate event, offering, quota, and eligibility.
- Track purchase lifecycle.
- Support operator-assisted corrections.
- Generate human-readable purchase references.
- Maintain complete status history.

### 11.5 Payment and Ledger

- Record expected and received amounts.
- Support payment references and evidence.
- Support verification workflow.
- Support installment ledger.
- Detect balance, overpayment, and underpayment.
- Record refunds, transfers, and adjustments with audit history.
- Permit future payment-provider integration.

### 11.6 Livestock Management

- Register livestock with unique operational identity.
- Record source, species, category, weight, condition, and readiness.
- Assign physical location.
- Track lifecycle status.
- Prevent invalid or duplicate allocation.
- Support batch operations where safe.

### 11.7 Allocation

- Allocate eligible participants or purchases to livestock.
- Validate sharing capacity.
- Support provisional and confirmed allocations.
- Record reassignment reasons.
- Provide allocation manifests.

### 11.8 Slaughter Operations

- Define slaughter sessions, stations, or queues.
- Assign livestock to operational sequence.
- Record check-in and readiness.
- Track milestones such as queued, called, in process, completed, held, or cancelled.
- Record exceptions and responsible operators.
- Surface live status to authorized users.

### 11.9 Distribution

- Define distribution method and entitlement rules.
- Record packaging or portion information where required.
- Assign beneficiary, pickup, or delivery records.
- Track completion and exceptions.
- Support aggregate operational reporting.

### 11.10 Notification

- Issue transactional notifications for meaningful status changes.
- Support templates by channel and audience.
- Prevent duplicate delivery.
- Record notification attempts.
- Keep delivery provider replaceable.

### 11.11 Reporting and Audit

- Report purchasing by channel.
- Report payments and outstanding balances.
- Report participant quota and allocation.
- Report livestock readiness and utilization.
- Report slaughter throughput.
- Report distribution completion.
- Record actor, timestamp, action, target, and relevant before/after data for sensitive changes.

---

## 12. Status Model Guidelines

Each aggregate owns its own lifecycle. Avoid one global status enum.

Illustrative status groups:

| Aggregate            | Example Statuses                                                             |
| -------------------- | ---------------------------------------------------------------------------- |
| Event                | Draft, Published, Active, Suspended, Closed, Archived                        |
| Quota Reservation    | Reserved, Consumed, Released, Expired                                        |
| Purchase             | Draft, Pending Payment, Paid, Eligible, Allocated, Completed, Cancelled      |
| Payment              | Pending, Submitted, Verified, Rejected, Refunded                             |
| Saving Account       | Draft, Active, Partially Funded, Fully Funded, Converted, Cancelled, Expired |
| Giveaway Application | Submitted, Under Review, Approved, Rejected, Assigned                        |
| Livestock            | Registered, Inspected, Ready, Allocated, Queued, Slaughtered, Held           |
| Allocation           | Provisional, Confirmed, Released, Reassigned                                 |
| Distribution         | Pending, Prepared, Ready, Collected, Delivered, Failed                       |

The Event lifecycle above is fixed for Phase 1. Final transitions for the
remaining aggregates must be completed during domain modeling.

---

## 13. Non-Functional Requirements

### 13.1 Capacity

Initial design target:

- 1,000 active Sohibul Qurban per annual event;
- thousands of purchases, payments, installment entries, livestock records, and audit events;
- concurrent internal operators during peak event hours;
- bursty storefront traffic near deadlines.

The architecture should support higher scale through measured optimization rather than premature distribution.

### 13.2 Availability

- Critical event-day functions should remain operational during partial subsystem failures.
- Long-running or external operations should not block transactional requests.
- The system should provide clear degraded-mode behavior for notifications and analytics.

### 13.3 Performance

Initial targets, subject to measurement:

- standard API reads: p95 under 500 ms;
- standard transactional writes: p95 under 1 second, excluding external providers;
- dashboard freshness: generally under 10 seconds for operational metrics;
- paginated list endpoints for all potentially large datasets.

### 13.4 Security

- role-based access control;
- least-privilege administration;
- secure session or token handling;
- encryption in transit;
- protected sensitive personal data;
- audit logging for privileged actions;
- rate limiting for public endpoints;
- input validation and output encoding;
- secret management outside source control.

### 13.5 Reliability and Data Integrity

- transactional consistency for purchase, payment, quota, and allocation changes;
- idempotency for payment callbacks and retried commands;
- optimistic or explicit concurrency controls for contested resources;
- database constraints for critical invariants;
- recoverable background processing;
- backup and restore procedures.

### 13.6 Observability

- structured logs with correlation identifiers;
- metrics for API, database, queue, and business operations;
- traces where cross-module or external calls justify them;
- alerting for payment failures, queue backlog, allocation conflicts, and event-day degradation.

### 13.7 Maintainability

- modular domain boundaries;
- generated or shared API contracts where practical;
- automated validation through repository-level commands;
- no duplicated business rules between web applications;
- explicit architecture decision records for major deviations.

---

## 14. Data and Privacy

The system may process:

- names;
- contact information;
- addresses;
- identity or eligibility evidence;
- payment evidence;
- participant and sponsor relationships;
- event attendance;
- distribution information.

Requirements:

- collect only necessary data;
- classify sensitive fields;
- restrict field-level access where required;
- define retention rules;
- avoid exposing giveaway or beneficiary data publicly;
- support correction and deletion policies where legally and operationally applicable;
- preserve financial and audit records according to applicable retention rules.

---

## 15. Integrations

Potential integrations include:

- payment gateways;
- bank transfer reconciliation;
- messaging providers;
- email providers;
- object storage;
- identity providers;
- accounting exports;
- weighing or livestock devices;
- QR or barcode scanners.

All external integrations should be isolated behind adapters. The initial release may use manual workflows while retaining stable internal interfaces.

---

## 16. Product Analytics

Track at minimum:

- channel conversion rate;
- abandoned checkout rate;
- payment completion time;
- saving-account funding progress;
- saving-to-purchase conversion rate;
- giveaway approval and assignment rate;
- quota utilization;
- livestock allocation completion;
- event throughput;
- operational exception rate;
- distribution completion;
- notification delivery rate.

Analytics must not become the source of truth for transactional status.

---

## 17. Release Strategy

### Phase 0 — Foundation

- monorepo tooling;
- application shells;
- database and local infrastructure;
- authentication baseline;
- event and user foundations;
- API conventions;
- observability baseline.

### Phase 1 — Commerce Core

- event and offering catalogue;
- common purchasing;
- participant registration;
- payment recording and verification;
- purchase tracking;
- initial operations dashboard.

### Phase 2 — Alternative Purchasing

- saving accounts and installments;
- saving conversion;
- giveaway programs and recipients;
- cross-channel reconciliation.

### Phase 3 — Livestock and Allocation

- livestock registry;
- inspection and readiness;
- allocation;
- location management;
- operational manifests.

### Phase 4 — Event-Day Operations

- check-in;
- queue and slaughter workflow;
- live operational dashboard;
- incident handling;
- resilient field workflows.

### Phase 5 — Distribution and Reporting

- distribution management;
- certificates or post-event documents;
- sponsor and participant reports;
- management reporting;
- historical event archive.

Priorities may be changed without invalidating this PRD, provided the core domain boundaries remain intact.

---

## 18. Success Criteria

The initial product is successful when:

- all three purchasing channels can create traceable eligible purchases;
- each active Sohibul Qurban can be traced to its channel and purchase basis;
- finance can reconcile common payments and saving installments;
- giveaway funding and recipient assignment are traceable;
- operations can determine quota, livestock, allocation, and event readiness from one system;
- event teams can update and monitor slaughter progress;
- sensitive overrides and corrections are auditable;
- annual events remain historically isolated;
- the system operates under expected peak load with no critical data inconsistency.

---

## 19. Open Product Decisions

The following decisions remain intentionally open:

- whether saving targets lock price;
- saving cancellation and transfer policy;
- giveaway eligibility and selection workflow;
- livestock procurement ownership;
- cattle share and other package allocation rules;
- participant attendance requirements;
- event-day offline or low-connectivity mode;
- distribution entitlement model;
- certificate generation;
- payment gateway selection;
- notification channels;
- data retention periods;
- organization and multi-tenant requirements;
- public self-service identity model.

Phase 1 common purchasing has resolved Offering shape, direct checkout, quota
reservation, payment evidence, and participant activation through ADR-042.
Those decisions do not define later Saving, Giveaway, refund, payment-gateway,
or livestock-allocation policy.

These decisions should be captured through updates to this PRD or Architecture Decision Records.

---

## 20. Change Management

This document is a living product baseline.

A requirement may be refined when:

- user research identifies a better workflow;
- Figma designs expose missing states;
- field operations require additional controls;
- legal, religious, financial, or privacy guidance changes;
- implementation discovery reveals a material constraint.

Changes should preserve:

- separation of purchasing channels;
- separation of participant and payer roles;
- event-based historical integrity;
- traceability of financial and operational actions;
- shared domain rules across applications.

---

## 21. Canonical Product Capability Map

The product is organized by business capability rather than by individual screens.

```text
Qurban Platform
├── Storefront
├── Purchasing
├── Party & Participant
├── Payment & Funding
├── Livestock
├── Allocation
├── Event Operations
├── Distribution
├── Identity & Access
└── Administration & Reporting
```

This capability map is canonical for product decomposition, roadmap planning, and epic creation. UI navigation may differ.

### 21.1 Storefront

```text
Storefront
├── Event Landing
├── Qurban Offering Catalogue
├── Search & Filter
├── Offering Detail
├── Common Purchase Journey
├── Saving Journey
├── Giveaway Journey
├── Purchase Tracking
└── Participant Documents
```

The catalogue is defined as an **Offering Catalogue**, not only an animal catalogue. An offering may represent an individual animal, category, package, livestock share, saving target, or funded program.

### 21.2 Purchasing

```text
Purchasing
├── Common Purchasing
├── Saving Purchasing
├── Giveaway Purchasing
├── Checkout
├── Purchase Validation
├── Purchase Confirmation
├── Cancellation
└── Purchase History
```

Phase 1 uses direct checkout with exactly one Offering per Purchase. Shopping
Cart and multi-offering checkout require a later requirement and decision.

### 21.3 Party & Participant

```text
Party & Participant
├── Purchaser
├── Payer
├── Saving Account Holder
├── Sponsor
├── Giveaway Applicant
├── Giveaway Recipient
├── Sohibul Qurban
└── Participant Verification
```

The system must not collapse all roles into a generic `Customer` concept.

### 21.4 Payment & Funding

```text
Payment & Funding
├── Payment Methods
├── Payment Instructions
├── Payment Confirmation
├── Payment Verification
├── Installment Ledger
├── Saving Balance
├── Sponsor Funding
├── Refund
└── Reconciliation
```

A payment gateway is an integration adapter, not the business capability itself.

### 21.5 Livestock

```text
Livestock
├── Livestock Registry
├── Classification
├── Health Inspection
├── Weight Recording
├── Readiness
├── Pen Assignment
├── Availability
└── Livestock History
```

Livestock must be modeled as a lifecycle-managed operational entity, not generic inventory.

### 21.6 Allocation

```text
Allocation
├── Purchase Allocation
├── Sohibul Qurban Allocation
├── Shared Livestock Capacity
├── Provisional Allocation
├── Confirmed Allocation
├── Reallocation
└── Allocation Manifest
```

Allocation is a first-class domain due to capacity, concurrency, reassignment, and audit requirements.

### 21.7 Event Operations

```text
Event Operations
├── Event Configuration
├── Event Readiness
├── Participant Check-in
├── Livestock Check-in
├── Slaughter Schedule
├── Slaughter Queue
├── Slaughter Station
├── Live Status Tracking
├── Operational Incident
└── Event Completion
```

The dashboard is a view over event operations, not the domain itself.

### 21.8 Distribution

```text
Distribution
├── Distribution Planning
├── Portion Preparation
├── Beneficiary Assignment
├── Pickup Management
├── Delivery Management
├── Collection Confirmation
├── Delivery Proof
└── Distribution Completion
```

Final distribution rules remain subject to operational discovery.

### 21.9 Identity & Access

```text
Identity & Access
├── Storefront Registration
├── Storefront Login
├── Operator Login
├── User Profile
├── Role Management
├── Permission Management
├── Event Scope
└── Audit Access
```

Public and internal identities may use different authentication policies.

### 21.10 Administration & Reporting

```text
Administration & Reporting
├── Event Administration
├── Offering Management
├── Purchasing Administration
├── Payment Verification Queue
├── Saving Administration
├── Giveaway Administration
├── Participant Management
├── Livestock Management
├── User & Permission Management
├── Audit Log
├── Operational Dashboard
└── Reports & Export
```

---

## 22. Revised Delivery Phases

### Phase 1 — Commerce Foundation

- Event;
- Offering Catalogue;
- Common Purchasing;
- Payment Verification;
- Sohibul Qurban Activation;
- Basic Operations Dashboard.

### Phase 2 — Alternative Purchasing

- Saving Purchasing;
- Installment Ledger;
- Saving Conversion;
- Giveaway Program;
- Giveaway Recipient Assignment.

### Phase 3 — Livestock and Allocation

- Livestock Registry;
- Inspection;
- Readiness;
- Participant Allocation;
- Shared Livestock Capacity.

### Phase 4 — Event-Day Operations

- Check-in;
- Slaughter Schedule;
- Queue;
- Live Status;
- Incident Handling.

### Phase 5 — Distribution and Reporting

- Distribution;
- Pickup or Delivery;
- Proof;
- Certificates;
- Reports;
- Historical Event Archive.

---

## 23. Capability-Level Open Questions

### Resolved for Phase 1

- Offerings are event-scoped sellable packages, shares, or categories and
  remain separate from physical Livestock.
- One direct checkout selects exactly one Offering; there is no Shopping Cart
  or purchase-item aggregate.

### Still Open

The following requirements remain unresolved and must be verified before their
affected implementation:

1. Whether saving plans lock the offering and price at creation.
2. Whether giveaway recipients are selected by sponsor, committee, manual approval, or random draw.
3. Whether each Sohibul Qurban performs the slaughter personally and therefore requires individual attendance and queue scheduling.
4. Whether distribution includes beneficiary delivery, Sohibul Qurban entitlement, or both.

These decisions should update the PRD, Product Map, and relevant ADRs before their affected phase enters BUILD.
