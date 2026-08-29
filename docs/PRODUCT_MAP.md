# Product Capability Map

## Qurban Commerce and Operations Platform

**Status:** Initial canonical map  
**Purpose:** Product decomposition, epic planning, roadmap sequencing, and application ownership

---

## 1. Product Map

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

This map represents business capabilities. It is not a sitemap, database schema, or mandatory one-folder-per-node architecture.

---

## 2. Storefront

```text
Storefront
├── Event Landing
├── Qurban Offering Catalogue
│   ├── Offering Listings
│   ├── Search & Filter
│   └── Offering Details
├── Common Purchase Journey
├── Saving Journey
├── Giveaway Journey
├── Purchase Tracking
└── Participant Documents
```

### Application Ownership

```text
apps/storefront-web
```

### Notes

- Use `Offering Catalogue`, not only `Animal Catalogue`.
- Phase 1 Offerings are event-scoped sellable packages, shares, or categories,
  not physical Livestock records.
- Phase 1 uses direct checkout with one Offering per Purchase; there is no
  Shopping Cart or purchase-item aggregate.
- Other Offering forms and multi-offering checkout require later requirements.

---

## 3. Purchasing

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

### Backend Ownership

```text
apps/api/internal/modules/purchasing
apps/api/internal/modules/saving
apps/api/internal/modules/giveaway
```

### Core Rule

All eligible channels converge into one canonical Purchase lifecycle.

For Phase 1 common purchasing, checkout reserves Event and Offering quota in
participant units for 24 hours. Submitted evidence pauses expiry; activation
consumes quota; expiry, cancellation, or rejection releases it.

---

## 4. Party & Participant

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

### Backend Ownership

```text
apps/api/internal/modules/identity
apps/api/internal/modules/participant
```

### Core Rule

The same person may fulfill multiple roles, but role relationships remain explicit.

---

## 5. Payment & Funding

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
└── Financial Reconciliation
```

### Backend Ownership

```text
apps/api/internal/modules/payment
apps/api/internal/modules/saving
apps/api/internal/modules/giveaway
```

### Notes

Payment gateways and banks are adapters to this capability.

---

## 6. Livestock

```text
Livestock
├── Livestock Registry
├── Livestock Classification
├── Health Inspection
├── Weight Recording
├── Livestock Readiness
├── Pen Assignment
├── Livestock Availability
└── Livestock History
```

### Backend Ownership

```text
apps/api/internal/modules/livestock
```

### Core Lifecycle

```text
REGISTERED
→ INSPECTED
→ READY
→ ALLOCATED
→ QUEUED
→ SLAUGHTERED
```

Final statuses remain subject to domain modeling.

---

## 7. Allocation

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

### Backend Ownership

```text
apps/api/internal/modules/allocation
```

### Core Rule

Allocation capacity and reassignment must be transactionally safe and auditable.

---

## 8. Event Operations

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

### Backend Ownership

```text
apps/api/internal/modules/event
apps/api/internal/modules/slaughter
```

### Application Ownership

```text
apps/operations-web
```

### Notes

The real-time dashboard is a read model over Event Operations and related domains.

The Full Event-Day MVP treats the following as one vertical operational chain:

```text
3–4 Local Execution Days
→ Field Teams, Shifts, Stations, and Readiness
→ Livestock and Participant Check-In
→ Slaughter Queue and Milestones
→ Distribution
→ Customer Status and Completion Evidence
```

---

## 9. Distribution

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

### Backend Ownership

```text
apps/api/internal/modules/distribution
```

### Open Scope

The Full Event-Day MVP covers both explicit Sohibul Qurban entitlement and
beneficiary distribution, with pickup or delivery, proof, exceptions, and
completion. Route optimization and generalized ecommerce delivery remain out
of scope.

---

## 10. Identity & Access

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

### Ownership

```text
apps/api/internal/modules/identity
apps/api/internal/platform/auth
```

Public and operations authentication may use different mechanisms.

---

## 11. Administration & Reporting

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

### Application Ownership

```text
apps/operations-web
```

### Backend Ownership

```text
apps/api/internal/modules/reporting
apps/api/internal/modules/* application queries
```

Administration is a user interface capability over authoritative domains. It must not become a single unrestricted `admin` domain.

---

## 12. Application Map

### Storefront Web

```text
apps/storefront-web
├── Event Landing
├── Offering Catalogue
├── Common Purchasing
├── Saving Purchasing
├── Giveaway Purchasing
├── Payment Interaction
├── Purchase Tracking
└── Participant Profile & Documents
```

### Operations Web

```text
apps/operations-web
├── Event Dashboard
├── Purchasing Operations
├── Payment Verification
├── Saving Operations
├── Giveaway Operations
├── Participant Operations
├── Livestock Operations
├── Allocation Operations
├── Slaughter Operations
├── Distribution Operations
└── Administration & Reporting
```

### Go API

```text
apps/api/internal/modules/
├── event/
├── identity/
├── offering/
├── purchasing/
├── payment/
├── saving/
├── giveaway/
├── participant/
├── livestock/
├── allocation/
├── slaughter/
├── distribution/
├── notification/
└── reporting/
```

---

## 13. Delivery Roadmap

The approved implementation sequence is maintained in
`MVP-DELIVERY-ROADMAP.md`. Its 16-week boundary keeps Alternative Purchasing
deferred while completing the full `COMMON` commerce-to-event-day journey.

### Phase 1 — Commerce Foundation

```text
Event
Offering Catalogue
Common Purchasing
Payment Verification
Sohibul Qurban Activation
Basic Operations Dashboard
```

### Phase 2 — Alternative Purchasing

```text
Saving Purchasing
Installment Ledger
Saving Conversion
Giveaway Program
Giveaway Recipient Assignment
```

### Phase 3 — Livestock and Allocation

```text
Livestock Registry
Inspection
Readiness
Participant Allocation
Shared Livestock Capacity
```

### Phase 4 — Event-Day Operations

```text
Check-in
Slaughter Schedule
Queue
Live Status
Incident Handling
```

### Phase 5 — Distribution and Reporting

```text
Distribution
Pickup or Delivery
Proof
Certificates
Reports
Historical Event Archive
```

---

## 14. Recommended First Vertical Slice

```text
Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Operations Dashboard Projection
```

The slice is complete only when it includes:

- backend domain behavior;
- persistence;
- API contracts;
- Storefront flow;
- Operations verification flow;
- authorization;
- audit;
- tests;
- observability.

---

## 15. Open Requirements

### Resolved for Phase 1

- **Offering model:** event-scoped sellable packages, shares, or categories,
  separate from physical Livestock.
- **Checkout model:** direct checkout with exactly one Offering per Purchase and
  no Shopping Cart.

### Resolved for the Full Event-Day MVP

- **Execution calendar:** exactly three or four inclusive local execution days
  in an Event IANA timezone.
- **Field organization:** event-scoped teams, members, shifts, stations,
  handovers, incidents, and support escalation.
- **Personal slaughter:** Event/participant-configurable self, proxy, or no
  attendance; never inferred from financial eligibility.
- **Distribution:** explicit Sohibul entitlement and beneficiary records,
  pickup or delivery, proof, exceptions, and completion.
- **Realtime:** polling first, then SSE for one-way value; no WebSocket without
  a proven bidirectional need.
- **Mobile/degraded mode:** responsive PWAs plus allowlisted, idempotent,
  non-financial offline field milestones; sensitive commands stay online-only.

### Still Open

1. **Saving price policy**
   Confirm whether a saving plan locks price and offering at creation.

2. **Giveaway selection**
   Confirm whether the recipient is selected by sponsor, committee, manual approval, or random draw.

These remaining questions affect only Saving and Giveaway. The Event-Day MVP
requirements above are accepted baselines and must be revalidated, not
reinvented, when each task becomes active.
