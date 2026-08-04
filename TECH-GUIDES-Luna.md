# Technical Guide: Qurban Commerce and Operations Platform

## 1. Purpose

This document records the technical discovery of the repository and explains how the intended technology stack supports the Qurban business process described in `README.md`, `docs/PRD.md`, `docs/PRODUCT_MAP.md`, `docs/ARCHITECTURE.md`, and `docs/DECISIONS.md`.

The platform is intended to manage an annual livestock-purchasing program and the related operational event: purchasing, payment, participant activation, livestock preparation, allocation, slaughter execution, distribution, and reporting.

## 2. Important Current-State Qualification

The repository currently contains the platform foundation, not the complete Qurban product.

Implemented foundation:

- React application shells;
- Go API shell;
- PostgreSQL connectivity and local Docker Compose infrastructure;
- shared frontend packages;
- placeholder public and operations OpenAPI contracts;
- PWA shell configuration;
- API health and readiness endpoints;
- repository formatting, linting, testing, build, and validation workflows.

The following business capabilities are currently deferred:

- Qurban events and event configuration;
- offering catalogue;
- Common, Saving, and Giveaway purchasing;
- payment verification and funding;
- participant and Sohibul Qurban activation;
- livestock management;
- allocation;
- slaughter operations;
- distribution;
- authentication and authorization;
- reporting projections and operational dashboards;
- background processing and real-time updates;
- production deployment.

Therefore, the architecture described below is the intended target architecture. It should not be interpreted as functionality already implemented in the current codebase.

## 3. System Overview

```text
Public users
    |
    v
Storefront Web
    | HTTPS / JSON
    v
Go API - system of record
    |
    +-- Event
    +-- Identity
    +-- Offering
    +-- Purchasing
    +-- Payment
    +-- Saving
    +-- Giveaway
    +-- Participant
    +-- Livestock
    +-- Allocation
    +-- Slaughter
    +-- Distribution
    +-- Notification
    +-- Reporting
    |
    +-- PostgreSQL
    +-- Background worker
    +-- External provider adapters

Internal operators
    |
    v
Operations Web
```

The initial backend architecture is a Go modular monolith rather than a microservice system. Domain modules are separated inside one deployable backend so that purchasing, payment eligibility, quota, participant activation, and allocation can remain transactionally close.

Service extraction is intentionally deferred until there is measurable evidence such as independent scaling requirements, different deployment cadence, compliance isolation, dedicated ownership, or availability boundaries.

## 4. Monorepo Structure

```text
apps/
├── storefront-web/       # Public React application
├── operations-web/       # Internal React application
└── api/                  # Go modular-monolith backend

packages/
├── api-client/           # Shared transport behavior
├── ui/                   # Shared accessible UI primitives
├── typescript-config/    # Shared TypeScript configuration
└── validation/           # Planned/shared validation utilities

contracts/
└── openapi/
    ├── storefront.yaml   # Public API contract
    └── operations.yaml   # Internal API contract

docs/
├── PRD.md
├── PRODUCT_MAP.md
├── ARCHITECTURE.md
└── DECISIONS.md

infrastructure/
└── compose.yaml          # Local PostgreSQL infrastructure
```

The monorepo keeps frontend applications, backend code, contracts, infrastructure, and documentation aligned. Application-specific business behavior must remain in the owning application or backend domain module rather than being moved into shared packages for convenience.

## 5. Frontend Stack

Both web applications use:

- React 19;
- TypeScript;
- Vite;
- React Router;
- Remix-style route organization while retaining the Vite SPA runtime;
- TanStack Query for remote/server state;
- Tailwind CSS v4;
- Oxlint;
- Vitest;
- Vite PWA and Workbox tooling.

React Router owns navigation and route composition. Each application should maintain centralized route registries under `src/routes/paths.ts` and `src/routes/routes.tsx`.

TanStack Query is intended to own API request state, including loading, error, cache, refetch, and mutation invalidation. Its cache is only a client-side view. It must never be treated as authoritative for payment state, quota, saving balances, allocation capacity, or queue position.

### Storefront Web

Location: `apps/storefront-web`

Default development URL: `http://127.0.0.1:5173`

Intended responsibilities:

- event and offering discovery;
- common purchasing;
- saving-plan registration and installment interaction;
- giveaway participation;
- purchaser and participant data collection;
- payment initiation or evidence submission;
- purchase tracking;
- participant-facing event information and documents.

The Storefront must not access the database directly, implement authoritative business rules, expose internal operational data, or perform privileged status overrides.

### Operations Web

Location: `apps/operations-web`

Default development URL: `http://127.0.0.1:5174`

Intended responsibilities:

- event configuration;
- purchase administration;
- payment verification;
- saving and giveaway operations;
- participant administration;
- livestock and allocation operations;
- check-in, slaughter, and distribution workflows;
- dashboards, incidents, reporting, and audit access.

The Operations Web is an internal control plane, but backend authorization remains mandatory. Frontend route guards are usability features, not security controls.

### Shared UI and PWA

`packages/ui` provides domain-agnostic accessible UI primitives, semantic tokens, and reusable components. It must not contain Qurban business rules, routes, API access, authentication, or environment reads.

The two applications have independent PWA manifests and service workers. Only immutable shell assets are precached. API responses, participant records, financial data, authentication state, and mutations are not cached or replayed offline. This prevents stale operational data from being presented as authoritative.

## 6. Backend Stack

Location: `apps/api`

The Go API currently provides:

- `GET /health`;
- PostgreSQL-backed `GET /ready`;
- graceful shutdown;
- structured logging;
- local live reload through Air;
- PostgreSQL connectivity.

The intended module layout is:

```text
apps/api/internal/
├── platform/
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

The preferred dependency direction is:

```text
HTTP and integration adapters
            |
            v
Application use cases
            |
            v
Domain model
```

Domain code should not depend on HTTP frameworks, PostgreSQL drivers, or external provider SDKs. External providers should be isolated behind ports and adapters.

## 7. PostgreSQL and Data Integrity

PostgreSQL is the authoritative transactional database. It is appropriate for this domain because the platform requires relational integrity, transactions, financial ledgers, event-scoped records, audit history, concurrency control, and reporting queries.

Database rules include:

- UUIDs for public identifiers;
- exact numeric or integer minor units for money;
- UTC timestamps with explicit timezone handling;
- foreign keys for critical relationships;
- unique and check constraints for stable invariants;
- append-oriented financial ledgers;
- status history for sensitive lifecycle changes;
- event references on event-scoped aggregates.

Critical operations should use database transactions, including quota reservation, payment verification, saving conversion, giveaway assignment, allocation confirmation, and contested queue transitions.

Historical event data must remain stable when future prices, offerings, quota rules, or policies change. Use snapshots or versioned references for historically significant configuration.

## 8. Core Business Model

### Annual Qurban Event

A `QurbanEvent` represents one annual operating cycle. It contains the event year, purchasing windows, slaughter dates, location, quota, capacity, status, and event-specific configuration.

All event-scoped records must reference the event so historical years remain isolated.

### Offering

An Offering is the commercial abstraction shown to users. It may represent an individual livestock unit, category, package, livestock share, saving target, or funded program.

Offering data includes price, species, category, participant capacity, publication status, and availability rules.

An Offering is not the same as physical livestock. Livestock is managed separately by Operations.

### Canonical Purchase

Every successful purchasing channel converges into one canonical Purchase lifecycle:

```text
Common checkout ------------------+
                                  |
Saving conversion ----------------+--> Purchase --> Sohibul Qurban
                                  |
Giveaway assignment --------------+
```

Do not create three unrelated order systems with duplicated lifecycle logic.

### Explicit Roles

These roles must remain separate even if the same person fulfills multiple roles:

- purchaser;
- payer;
- saving-account holder;
- sponsor;
- giveaway applicant or nominee;
- giveaway recipient;
- Sohibul Qurban.

Sohibul Qurban is an outcome of an eligible purchase or allocation, not a purchasing channel.

## 9. Purchasing Business Processes

### Common Purchasing

```text
1. Select event and offering
2. Enter purchaser and participant information
3. Validate quota and availability
4. Create pending purchase
5. Initiate payment or submit evidence
6. Verify payment
7. Mark purchase eligible
8. Activate Sohibul Qurban records
9. Allocate livestock immediately or later
```

Quota checks and payment transitions must be performed by the API. The frontend must handle stale state and conflict responses rather than overriding the backend.

### Saving Purchasing

A Saving Account is funding progress, not a completed Purchase.

```text
1. Select a saving target
2. Create Saving Account
3. Record installments
4. Verify and reconcile installments
5. Calculate remaining balance
6. Reach conversion eligibility
7. Confirm conversion
8. Convert into canonical Purchase
9. Activate Sohibul Qurban records
```

Typical saving states include `DRAFT`, `ACTIVE`, `PARTIALLY_FUNDED`, `FULLY_FUNDED`, `CONVERSION_PENDING`, `CONVERTED`, `CANCELLED`, and `EXPIRED`.

The saving module must support installment history, underpayment, overpayment, refunds, transfers, price policy, expiry, operator corrections, and idempotent conversion.

### Giveaway Purchasing

Giveaway funding separates the sponsor from the recipient.

```text
1. Create giveaway program
2. Record sponsor funding or budget
3. Register applicants or nominees
4. Review eligibility
5. Approve recipient
6. Create or assign funded Purchase
7. Activate recipient as Sohibul Qurban
8. Allocate livestock
9. Provide appropriate sponsor and recipient reporting
```

Giveaway workflows require privacy controls, duplicate-recipient prevention, approval audit history, funding ceilings, and explicit relationships between sponsor, applicant, recipient, and participant.

## 10. Livestock, Allocation, and Operations

### Livestock Lifecycle

Livestock is a physical lifecycle-managed entity:

```text
REGISTERED -> INSPECTED -> READY -> ALLOCATED -> QUEUED -> SLAUGHTERED
```

The system should track identity, species, category, source, weight, condition, inspection, readiness, location, allocation, queue position, slaughter, holds, and history.

### Allocation

Allocation connects eligible purchases or participants to livestock. It must support:

- individual and shared capacity;
- provisional allocation;
- confirmed allocation;
- release;
- reassignment;
- allocation manifests.

Allocation is concurrency-sensitive. Use row locks, atomic conditional updates, unique constraints, optimistic versioning, and idempotency where appropriate. Reassignment and override actions require authorization, a reason, and an audit record.

### Slaughter Operations

The event-day workflow includes participant check-in, livestock check-in, schedule and queue management, station/team execution, status milestones, and incident handling.

Typical operational states include queued, called, in progress, completed, held, and cancelled.

### Distribution

Distribution may include portion preparation, beneficiary assignment, pickup, delivery, collection confirmation, proof, failed delivery, and completion. The exact entitlement and delivery model remains an open product decision.

## 11. Real-Time and Background Processing

The requirement is near-real-time operational visibility, not hard real-time guarantees.

The intended flow is:

```text
1. Authoritative transaction commits
2. Domain event is inserted into an outbox in the same transaction
3. Worker processes the outbox
4. Projection or integration task is updated
5. Operations dashboard refreshes or receives an update
```

Example events:

```text
PurchasePaid
SavingAccountFullyFunded
GiveawayRecipientApproved
ParticipantActivated
LivestockReady
AllocationConfirmed
LivestockQueued
SlaughterCompleted
DistributionCompleted
OperationalIncidentRaised
```

The recommended transport progression is:

1. Polling for the first dashboard implementation;
2. Server-Sent Events for valuable one-way live updates;
3. WebSockets only for genuinely bidirectional interactions.

The Operations dashboard should read from projections optimized for dashboard queries. Projections are rebuildable and must not become the source of transactional truth. Projection lag should be observable.

The intended worker process is `apps/api/cmd/worker`. It should handle outbox processing, notifications, payment callbacks, projection updates, report generation, and retryable integration work.

A dedicated message broker is not required initially. PostgreSQL-backed outbox and job processing may be sufficient for the expected scale of approximately 1,000 Sohibul Qurban and thousands of related records per annual event.

## 12. API and Integration Boundaries

The intended API surfaces are:

```text
/api/public/v1/*
/api/operations/v1/*
```

Public and internal API contracts must remain separate:

```text
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
```

Both surfaces may reuse application services, but they should have different DTOs, authorization rules, data exposure policies, pagination behavior, and error handling.

External systems should be isolated behind adapters, including:

- payment gateways;
- bank reconciliation;
- messaging and email providers;
- object storage;
- identity providers;
- document rendering;
- accounting exports;
- livestock devices or scanners.

Webhook handlers must authenticate the sender, store provider identifiers, be idempotent, tolerate duplicates and reordering, acknowledge quickly, and defer retryable work to the background worker.

## 13. Security and Audit

Authentication and authorization are not implemented yet.

The intended Storefront options include phone/email OTP, passwordless accounts, guest purchases, or an external identity provider.

Operations should use stronger authentication, managed accounts, role-based permissions, and possibly multi-factor authentication and event/location scope.

Example permissions:

```text
purchase.read
purchase.correct
payment.verify
saving.adjust
giveaway.approve
livestock.manage
allocation.override
slaughter.update
distribution.complete
audit.read
admin.manage
```

Authorization must be enforced in backend application policies, not only through frontend route guards.

Privileged changes should produce append-only audit records containing actor, permission context, request ID, timestamp, action, target, reason, and relevant before/after values.

## 14. Development and Validation

Requirements:

- Node.js 24 or newer;
- pnpm 10.30.0;
- Go 1.26 or newer;
- Docker Compose;
- Make;
- Air for Go live reload.

Common commands:

```bash
pnpm install
make infra-up
make dev
make validate
```

The complete local stack runs:

```text
Storefront Web: 127.0.0.1:5173
Operations Web: 127.0.0.1:5174
Go API:         127.0.0.1:8080
PostgreSQL:     127.0.0.1:5433 by default
```

`make validate` is intended to cover formatting, linting, type checking, tests, builds, Go checks, and Docker Compose configuration validation.

## 15. Delivery Roadmap

### Phase 1: Commerce Foundation

```text
Event
-> Offering
-> Common Purchase
-> Payment Verification
-> Sohibul Qurban Activation
-> Basic Operations Dashboard
```

### Phase 2: Alternative Purchasing

```text
Saving Accounts
-> Installments
-> Saving Conversion
-> Giveaway Program
-> Recipient Assignment
```

### Phase 3: Livestock and Allocation

```text
Livestock Registry
-> Inspection
-> Readiness
-> Participant Allocation
-> Shared Livestock Capacity
```

### Phase 4: Event-Day Operations

```text
Check-in
-> Slaughter Schedule
-> Queue
-> Live Status
-> Incident Handling
```

### Phase 5: Distribution and Reporting

```text
Distribution
-> Pickup or Delivery
-> Proof
-> Certificates
-> Reports
-> Historical Event Archive
```

## 16. Key Architectural Principles

1. Storefront and Operations are separate applications with separate exposure and security boundaries.
2. The Go API is the system of record and owns business invariants.
3. PostgreSQL owns authoritative transactional data.
4. Common, Saving, and Giveaway are separate channels that converge into one Purchase lifecycle.
5. Saving Accounts and Giveaway Applications are not Purchases until their conversion or approval rules are satisfied.
6. Purchaser, payer, sponsor, recipient, and Sohibul Qurban are separate business concepts.
7. Livestock is a physical lifecycle, not generic inventory.
8. Allocation is a first-class, concurrency-sensitive, auditable domain.
9. Dashboards are projections, not transactional truth.
10. Real-time updates should start with polling, then use SSE when justified.
11. Financial and privileged changes require traceability and idempotency.
12. Microservices and speculative multitenancy are intentionally deferred.
13. Product and architecture documents describe the target direction; current implementation status must be checked before claiming a capability is complete.

## 17. Primary References

- `README.md`
- `docs/PRD.md`
- `docs/PRODUCT_MAP.md`
- `docs/ARCHITECTURE.md`
- `docs/DECISIONS.md`
- `.codex/CURRENT_STATE.md`
- `apps/storefront-web/package.json`
- `apps/operations-web/package.json`
- `apps/api/go.mod`
- `infrastructure/compose.yaml`
