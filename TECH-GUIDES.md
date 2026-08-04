# Qurban Platform Technical Guide

## 1. Purpose

This guide explains how the `persona-apps` monorepo is intended to support the complete Qurban Commerce and Operations lifecycle: annual event setup, public purchasing, funding and payment verification, Sohibul Qurban activation, livestock operations, allocation, slaughter-day execution, distribution, and reporting.

It also distinguishes clearly between:

- the architecture and business capabilities already defined in the documentation;
- the technical foundation currently implemented in code; and
- the business functionality that still needs to be built.

The most important current-state conclusion is:

> The repository has a suitable technical and architectural foundation for the documented platform, but it does not yet implement the Qurban business processes. The README explicitly describes the applications as shells, and the OpenAPI contracts currently contain no business endpoints.

## 2. Product Context and Expected Scale

The platform is designed for one operating organization that runs recurring annual Qurban events.

The initial documented capacity target is approximately:

- up to 1,000 active Sohibul Qurban in one annual event;
- thousands of purchases, payments, installment entries, livestock records, allocations, audit events, and distribution records;
- multiple concurrent internal operators during peak event hours;
- bursty Storefront traffic near registration and payment deadlines;
- an Operations dashboard with data generally less than 10 seconds old.

This is a high-consequence, bursty operational workload rather than an internet-scale workload. Correct payment, quota, livestock-capacity, allocation, and queue transitions are more important than prematurely splitting the system into many services.

## 3. Monorepo Structure

```text
persona-apps/
├── apps/
│   ├── storefront-web/     Public participant and purchaser application
│   ├── operations-web/     Internal operational control plane
│   └── api/                Go API and future background worker
├── packages/
│   ├── api-client/         Shared HTTP transport behavior
│   ├── ui/                 Shared accessible, domain-neutral UI primitives
│   └── typescript-config/  Shared TypeScript configuration
├── contracts/openapi/
│   ├── storefront.yaml     Public API contract
│   └── operations.yaml     Internal API contract
├── infrastructure/
│   └── compose.yaml        Local PostgreSQL infrastructure
├── docs/                   Canonical product and architecture documents
├── Makefile                Repository development and validation commands
├── pnpm-workspace.yaml     JavaScript workspace definition
└── turbo.json              Workspace task orchestration
```

The monorepo keeps frontend applications, backend code, API contracts, shared packages, infrastructure, and product decisions versioned together. It does not mean every application shares all code. Storefront-specific and Operations-specific behavior remains in the owning application.

## 4. Current Technology Stack

### 4.1 Frontend runtime

Both web applications currently use:

| Technology            | Current role                                               |
| --------------------- | ---------------------------------------------------------- |
| React 19              | Component and user-interface runtime                       |
| TypeScript 6          | Static typing and frontend contracts                       |
| Vite 8                | Development server and production SPA build                |
| React Router 8        | Centralized navigation and route composition               |
| TanStack Query 5      | Remote API/server-state lifecycle                          |
| Tailwind CSS 4        | Styling through shared semantic tokens                     |
| React Aria Components | Accessible composite interactions in the shared UI package |
| Vite PWA and Workbox  | Installable shell and immutable application-asset caching  |
| Vitest 4              | Frontend tests                                             |
| Oxlint                | JavaScript and TypeScript linting                          |

The applications are Vite single-page applications. “Remix-style routing” refers to route hierarchy, layouts, boundaries, and declared route-data needs; it does not mean the repository runs a Remix server or server-side rendering.

TanStack Query is responsible for remote data fetching, loading/error/stale states, cache updates, and invalidation. Its cache is never the source of truth for contested business state such as payment status, quota, saving balance, livestock capacity, allocation, or queue position.

### 4.2 Frontend applications

#### Storefront Web

`apps/storefront-web` is the public-facing channel. Its intended responsibilities are:

- event and offering discovery;
- common purchasing;
- saving-plan registration and installment interaction;
- giveaway application or recipient journeys;
- purchaser and participant data collection;
- payment interaction;
- purchase tracking;
- participant instructions and documents.

It runs locally on `127.0.0.1:5173`.

#### Operations Web

`apps/operations-web` is the internal control plane. Its intended responsibilities are:

- event configuration;
- purchasing administration;
- payment verification and reconciliation;
- saving and giveaway operations;
- participant and Sohibul Qurban management;
- livestock registration and readiness;
- allocation;
- check-in and slaughter queues;
- incident handling;
- distribution;
- operational dashboards, audit, and reporting.

It runs locally on `127.0.0.1:5174`.

Keeping these applications separate is important because they have different audiences, authentication strength, permissions, sensitive fields, failure risks, and deployment needs.

### 4.3 Shared frontend packages

`@persona-apps/ui` contains accessible, domain-neutral UI foundations. It must not contain Qurban business rules, routes, authentication, API access, or environment-specific behavior.

`@persona-apps/api-client` contains shared transport concerns such as request serialization, timeout and cancellation handling, response parsing, normalized errors, and a future authentication-header extension point. Application-owned endpoint modules remain separate because Storefront and Operations use different OpenAPI contracts and expose different data.

### 4.4 Backend

The backend is a Go modular monolith under `apps/api`.

Current backend dependencies and infrastructure include:

- Go 1.26;
- `pgx/v5` for PostgreSQL connectivity;
- a standard Go HTTP API process;
- structured logging foundation;
- graceful shutdown;
- Air for development live reload;
- `/health` for process health;
- `/ready` for PostgreSQL-backed readiness.

The API normally runs on `127.0.0.1:8080`. The Make target can select the next free port when the requested development port is occupied.

The modular-monolith choice keeps transactions for purchases, payments, quota, saving conversion, giveaway assignment, participant activation, livestock allocation, and queue transitions close to one authoritative database. Domain modules are intended to have explicit domain, application, and adapter boundaries without requiring separate deployments.

### 4.5 Database

PostgreSQL is the intended system of record. The local Docker Compose environment currently uses PostgreSQL 18 Alpine and exposes configurable host port `5433` by default.

PostgreSQL is appropriate for this platform because it provides:

- ACID transactions for multi-record business changes;
- foreign keys and check constraints for critical invariants;
- unique constraints for business references and idempotency;
- row-level locks or atomic conditional updates for contested resources;
- exact numeric or integer-minor-unit money storage;
- append-oriented financial and audit records;
- event-scoped historical queries;
- reporting and projection support.

Redis is intentionally not part of the baseline. It should be introduced only for a measured need such as distributed rate limiting, session storage, cache coordination, or short-lived locking. Database correctness must never depend on cache correctness.

### 4.6 API contracts

The architecture separates public and internal API surfaces:

```text
/api/public/v1/*
/api/operations/v1/*
```

Their OpenAPI source files are:

```text
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
```

This separation prevents internal override operations and sensitive operational data from leaking into the public Storefront contract. The two surfaces may call the same backend application services while using different authorization policies and transport DTOs.

At present, both OpenAPI files are placeholders with `paths: {}`. No Qurban business API contract is implemented yet.

### 4.7 Workspace and quality tooling

The JavaScript workspace uses pnpm 10 and Turborepo. The root Makefile coordinates frontend, backend, and infrastructure commands.

The main repository validation command is:

```bash
make validate
```

It covers formatting checks, frontend linting, TypeScript checking, frontend tests and builds, Go vet/tests/build, and Docker Compose configuration validation.

## 5. Business Domain Model

The platform is organized around business capabilities rather than screens or generic e-commerce terminology.

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

### 5.1 Annual Qurban Event

`QurbanEvent` is the top-level operational cycle. Event-scoped records must reference an event so that one year’s prices, quotas, dates, and policies cannot silently change historical data from another year.

Historically significant values should use snapshots or versioned references. An archived event should normally be readable and immutable.

### 5.2 Offering versus Livestock

An `Offering` is the commercial item shown or selected during a purchasing journey. It may later represent a package, category, individual animal, cattle share, saving target, or funded program.

`Livestock` is a physical operational entity with identity, intake, inspection, weight, readiness, location, allocation, slaughter progression, and history.

These concepts must not be collapsed into a generic product-and-stock model. A customer can purchase an offering before a specific livestock unit is assigned.

### 5.3 Three purchasing channels, one Purchase lifecycle

The supported channels are:

```text
COMMON
SAVING
GIVEAWAY
```

They converge into one canonical Purchase lifecycle:

```text
Common Checkout ───────────────┐
                               │
Saving Account → Conversion ───┼──> Purchase ──> Sohibul Qurban
                               │
Giveaway Approval ─────────────┘
```

Key rules are:

- every Purchase belongs to exactly one channel;
- Common purchasing can create a pending Purchase directly;
- a Saving Account is funding before conversion, not automatically a Purchase;
- a Giveaway application is not a Purchase until approval and funded assignment;
- downstream eligibility, participant activation, allocation, event execution, and reporting use the canonical Purchase model.

### 5.4 People and roles

The same person can hold several roles, but the data relationships must remain explicit:

- purchaser;
- payer;
- saving-account holder;
- sponsor;
- giveaway applicant or nominee;
- giveaway recipient;
- Sohibul Qurban.

Sohibul Qurban is an outcome of an eligible Purchase, not a fourth purchasing channel.

### 5.5 Financial records

Payments, installments, refunds, transfers, and adjustments should be represented as traceable ledger entries rather than only a mutable balance field.

This supports:

- duplicate callback protection;
- reconciliation;
- overpayment and underpayment handling;
- compensating corrections;
- saving balances;
- sponsor funding;
- complete financial audit history.

Money must use integer minor units or an exact decimal representation. Floating-point arithmetic is prohibited for financial values.

### 5.6 Allocation

Allocation is a first-class domain, not a hidden order-fulfillment field. It connects eligible Purchases or Sohibul Qurban records to physical livestock capacity.

It must support:

- individual and shared livestock arrangements;
- provisional and confirmed assignment;
- capacity checks;
- release and reassignment;
- operator override with a reason;
- auditable history;
- concurrency-safe contested updates.

## 6. End-to-End Business Process Coverage

The following sections describe how the intended stack maps to each documented business process. These are target flows, not claims that they are already executable.

### 6.1 Annual event preparation

```text
Operator creates event
→ configures dates, location, policies, and quotas
→ defines and publishes offerings
→ prepares purchasing and operational windows
→ event becomes active
```

Technical ownership:

- Operations Web provides configuration workflows;
- Event and Offering backend modules enforce lifecycle and publication rules;
- PostgreSQL stores event-scoped configuration and historical snapshots;
- Operations OpenAPI exposes privileged commands and views;
- audit records capture sensitive configuration changes.

### 6.2 Common purchasing

```text
Visitor selects event and offering
→ enters purchaser and participant information
→ API validates event, availability, and quota
→ pending Purchase is created
→ payment is submitted or initiated
→ Finance verifies payment
→ Purchase becomes eligible
→ Sohibul Qurban is activated
→ allocation occurs immediately or later
```

Critical controls:

- authoritative quota checks occur in the Go API;
- payment and Purchase transitions use database transactions;
- retry-sensitive commands use idempotency keys;
- price and offering details are preserved for historical truth;
- public status excludes internal operational fields.

### 6.3 Saving purchasing

```text
User creates Saving Account and target
→ submits one or more installments
→ Finance verifies ledger entries
→ balance reaches fully funded state
→ conversion eligibility is evaluated
→ user or operator confirms the target if required
→ idempotent conversion creates/activates Purchase
→ Sohibul Qurban is activated
```

Critical controls:

- Saving Account and Purchase are separate aggregates;
- `FULLY_FUNDED` does not necessarily mean `CONVERTED`;
- installment corrections remain traceable;
- conversion is transactional and idempotent;
- price-lock, transfer, cancellation, refund, and expiry policies must be decided before implementation.

### 6.4 Giveaway purchasing

```text
Operator creates Giveaway Program
→ sponsor funding or budget is recorded
→ applicants, nominees, or recipients are registered
→ eligibility and selection are reviewed
→ recipient is approved
→ funded assignment creates/activates Purchase
→ recipient becomes Sohibul Qurban
→ livestock is allocated
→ sponsor and recipient receive appropriate reports
```

Critical controls:

- sponsor, applicant, payer, recipient, and Sohibul Qurban remain distinct;
- duplicate-recipient checks are required;
- private beneficiary data is not exposed publicly;
- approval and reassignment require permissions and audit records;
- the exact recipient-selection policy is still an open product decision.

### 6.5 Livestock preparation

```text
Livestock registered
→ classified and weighed
→ health/quality inspection recorded
→ readiness confirmed or held
→ pen/location assigned
→ available capacity exposed for allocation
```

The Livestock module owns physical lifecycle truth. Batch operations may improve event preparation, but every update must remain attributable and safe.

### 6.6 Allocation

```text
Eligible Purchase or participant selected
→ suitable livestock capacity checked
→ provisional allocation created
→ allocation confirmed
→ manifest produced
→ reassignment recorded when necessary
```

Concurrent operators may attempt to use the same capacity. The backend must prevent over-allocation with constraints, row locks, conditional updates, optimistic versions, or a suitable combination. A stale frontend receives a conflict response and refreshes authoritative state instead of forcing its cached choice.

### 6.7 Event-day and slaughter operations

```text
Participant and livestock check-in
→ readiness validated
→ livestock assigned to session/station/queue
→ queued
→ called or moved into process
→ slaughter completed, held, or cancelled
→ exceptions and responsible operators recorded
```

The Operations Web is optimized for fast field updates and exception visibility. Every state-changing command still passes through normal authorization and domain validation. The live dashboard observes resulting projections; it does not directly mutate queue or slaughter truth.

### 6.8 Distribution

```text
Slaughter/preparation prerequisites completed
→ portions or entitlements prepared
→ beneficiary, pickup, or delivery assignment created
→ collection/delivery completed
→ evidence and exceptions recorded
→ aggregate completion reported
```

Distribution is Qurban-specific and must not be reduced to ordinary e-commerce shipping. Final rules for Sohibul Qurban entitlement, beneficiaries, pickup, and delivery remain open.

### 6.9 Reporting and historical archive

Reports are derived from authoritative transactional records and rebuildable projections. Expected reporting includes:

- purchases by channel;
- payment verification and outstanding balances;
- saving progress and conversion;
- giveaway funding and assignments;
- participant quota usage;
- livestock readiness and utilization;
- allocations and exceptions;
- slaughter throughput;
- distribution completion;
- audit history;
- annual event archive.

Analytics and dashboard records are read models. They must never replace transactional data as the source of truth.

## 7. Near-Real-Time Processing for Peak Events

The documented requirement is near-real-time operations, with dashboard freshness generally under 10 seconds. It does not require hard real-time processing.

### 7.1 Recommended event flow

```text
Operator or public command
→ Go API validates permission and business rules
→ PostgreSQL transaction updates authoritative records
→ same transaction writes an outbox event
→ transaction commits
→ Go worker claims the outbox item
→ worker updates a dashboard projection
→ Operations Web refreshes by polling or receives an SSE notification
→ UI fetches the latest projection
```

Example events include:

- `PurchasePaid`;
- `SavingAccountFullyFunded`;
- `GiveawayRecipientApproved`;
- `ParticipantActivated`;
- `LivestockReady`;
- `AllocationConfirmed`;
- `LivestockQueued`;
- `SlaughterCompleted`;
- `DistributionCompleted`;
- `OperationalIncidentRaised`.

### 7.2 Why PostgreSQL outbox first

A PostgreSQL-backed outbox is a practical fit for the documented scale because it:

- writes the business change and event atomically;
- prevents notification or projection failures from rolling back valid business work;
- supports retry and recovery;
- avoids distributed transactions;
- can be operated by a small team;
- delays message-broker complexity until measurements justify it.

Outbox consumers must be idempotent. The system should monitor queue depth, oldest unprocessed event age, retry counts, and dead-letter or terminal failures.

### 7.3 Polling, SSE, and WebSocket

The accepted progression is:

1. use polling for the first dashboard;
2. add Server-Sent Events for valuable one-way live updates;
3. use WebSocket only if a genuinely bidirectional real-time interaction appears.

Polling is simple and can satisfy the sub-10-second freshness goal at the initial scale. SSE is suitable when operators need lower-latency one-way notifications without WebSocket complexity.

Neither polling nor SSE should carry unvalidated authoritative state transitions. They notify clients that data changed; normal API queries and commands remain the consistency boundary.

### 7.4 Contention and consistency

High-risk resources include:

- offering and participant quota;
- payment callbacks;
- saving conversion;
- giveaway assignment;
- livestock sharing capacity;
- allocation confirmation;
- slaughter queue transitions.

The backend should use, as appropriate:

- database transactions;
- unique and check constraints;
- row-level locking;
- atomic conditional updates;
- optimistic version columns;
- idempotency records;
- structured `409 Conflict` responses.

The frontend must treat conflicts as a normal operational state: show the conflict, refresh current data, and let the operator retry an allowed action.

### 7.5 Partial failures and degraded mode

The architecture separates authoritative transactions from asynchronous side effects:

- notification failure does not undo a verified payment;
- report-generation failure does not undo slaughter completion;
- temporary projection lag does not change allocation truth;
- payment-provider retries do not create duplicate ledger entries;
- stale dashboards expose freshness or lag rather than pretending to be current.

The current PWA policy caches only immutable application-shell assets. API, authentication, participant, financial, and operational records remain network-only. Offline mode must not show cached sensitive data as authoritative or queue uncontracted mutations for later replay.

A true low-connectivity event-day workflow is still an unresolved product and architecture decision.

## 8. Security, Authorization, and Audit

### 8.1 Authentication boundaries

Storefront and Operations may use different authentication mechanisms. Operations requires stronger authentication because operators can verify payments, override allocations, alter queue priority, and access private participant information.

The exact authentication mechanisms have not yet been selected.

### 8.2 Permission-based authorization

Authorization is intended to use permissions such as:

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

Frontend route guards improve navigation and usability but are not security controls. The Go application layer must enforce permissions and any event or location scope.

### 8.3 Audit records

Sensitive changes require append-only audit records containing the actor, permission context, request ID, timestamp, action, target, reason where required, relevant before/after values, and source application.

Examples include:

- payment verification override;
- installment adjustment;
- giveaway approval;
- participant replacement;
- livestock reassignment;
- queue priority override;
- distribution reversal.

Sensitive personal or payment evidence should be minimized in logs and audit payloads.

## 9. Observability and Operational Readiness

The production design should expose both technical and business signals.

Technical signals include:

- request rate, errors, and latency;
- PostgreSQL connection and query performance;
- outbox backlog and oldest-event age;
- worker retries and failures;
- integration failures;
- polling/SSE connection health;
- projection lag.

Business signals include:

- purchases by channel;
- pending payment verification;
- saving accounts by funding state;
- approved but unassigned giveaway recipients;
- quota utilization;
- unallocated participants;
- livestock readiness;
- slaughter queue and throughput;
- operational incidents;
- distribution completion.

Structured logs should include correlation or request IDs and avoid secrets or unnecessary personal information. Tracing can be added when asynchronous or extracted-service flows become difficult to correlate; it is not mandatory for the initial monolith.

## 10. Initial Deployment Shape

The intended first production topology is:

```text
Internet
   │
CDN / ingress
   ├── Storefront origin ──> Storefront static SPA
   ├── Operations origin ──> Operations static SPA
   └── API origin ─────────> Go API
                                  │
                          ┌───────┴───────┐
                          ▼               ▼
                     PostgreSQL       Go worker
```

Potential integrations such as payment gateways, banks, messaging providers, email, object storage, identity providers, and document rendering remain behind application ports and provider adapters.

A message broker or microservices are not required initially. A module should be extracted only after measured needs demonstrate independent scaling, security isolation, release cadence, ownership, availability, or integration-throughput requirements.

## 11. Current Implementation Reality

### 11.1 Implemented now

The repository currently contains:

- pnpm workspace and Turborepo task orchestration;
- Make-based repository commands;
- two React application shells;
- centralized frontend route registries and placeholder capability pages;
- TanStack Query providers and typed API transport boundaries;
- shared UI primitives and semantic tokens;
- independent Storefront and Operations PWA shells;
- Go API process foundation;
- PostgreSQL connection and readiness handling;
- `/health` and `/ready` endpoints;
- Docker Compose PostgreSQL service;
- placeholder separate public and Operations OpenAPI contracts;
- tests, linting, type checking, builds, and repository validation;
- canonical PRD, capability map, architecture, and decision register.

### 11.2 Not implemented yet

The following are still architectural intent rather than executable business functionality:

- Qurban Event and Offering modules;
- Common, Saving, and Giveaway purchasing;
- canonical Purchase lifecycle;
- person, role, participant, and Sohibul Qurban records;
- payment ledger, verification, reconciliation, refunds, and gateway integration;
- livestock registry, inspection, readiness, and location;
- allocation and capacity protection;
- check-in, slaughter queue, stations, and incident workflows;
- distribution and evidence;
- authentication and permission enforcement;
- business database schema and migrations;
- audit framework;
- idempotency framework;
- outbox and background worker;
- dashboard projections;
- polling or SSE business updates;
- notifications;
- production observability and deployment.

Therefore, the correct assessment is:

| Question                                                                    | Assessment                                                                                                        |
| --------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Does the documented product model cover the annual Qurban lifecycle?        | Yes, at capability and architecture level.                                                                        |
| Does the selected stack fit the initial scale and consistency requirements? | Yes, provided the documented transaction, outbox, concurrency, audit, and observability controls are implemented. |
| Does the current code execute all README/PRD business processes?            | No. It is still a validated foundation and application shell.                                                     |
| Is real-time event processing implemented?                                  | No. Polling, SSE, projections, outbox, and worker behavior are target architecture only.                          |
| Is the system production-ready?                                             | No. Business modules, security, migrations, resilience, observability, and deployment remain incomplete.          |

## 12. Open Product Decisions

Implementation must not invent the following unresolved rules:

1. Whether Offerings represent individual livestock, categories, packages, cattle shares, or a combination.
2. Whether one checkout can contain multiple Offerings.
3. Whether Saving locks the Offering, price, both, or neither.
4. How Giveaway recipients are selected and approved.
5. Whether every Sohibul Qurban personally participates in slaughter.
6. Whether personal slaughter needs attendance registration, queue numbers, stations, or proxy representation.
7. Whether Distribution covers Sohibul Qurban entitlement, beneficiaries, pickup, delivery, or a combination.
8. How low-connectivity or offline event-day operations should work.
9. Which payment, identity, notification, object-storage, and document providers should be used.

These decisions affect schemas, APIs, state machines, authorization, and UI flows and should be resolved before their vertical slice enters implementation.

## 13. Recommended Delivery Order

The product should be delivered through complete vertical slices instead of creating empty modules for the entire domain.

### Phase 1: Commerce foundation

```text
Qurban Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Basic Operations Dashboard
```

This first slice should include domain rules, migrations, transactions, OpenAPI contracts, Storefront flow, Operations verification flow, authorization, audit, idempotency, tests, and observability.

### Phase 2: Alternative purchasing

- Saving Accounts and installment ledger;
- Saving conversion;
- Giveaway Programs;
- recipient approval and assignment;
- cross-channel reconciliation.

### Phase 3: Livestock and allocation

- livestock registry;
- inspection and readiness;
- locations;
- participant allocation;
- shared livestock capacity;
- manifests and reassignment.

### Phase 4: Event-day operations

- participant and livestock check-in;
- slaughter sessions, queues, and stations;
- live status projections;
- incident handling;
- resilient field workflows.

### Phase 5: Distribution and reporting

- portion and entitlement preparation;
- beneficiary assignment;
- pickup or delivery;
- proof and completion;
- certificates and reports;
- historical event archive.

## 14. Final Technical Assessment

The monorepo structure is a sensible fit for this platform:

- two independent React applications preserve public and internal boundaries;
- a Go modular monolith centralizes authoritative business rules and keeps high-integrity operations transactionally close;
- PostgreSQL provides the consistency, constraints, ledger, concurrency, audit, and reporting capabilities required by the domain;
- separate OpenAPI contracts protect public/internal data boundaries;
- an outbox plus Go worker can support reliable asynchronous work at the documented scale;
- polling first and SSE later can provide near-real-time operational visibility without unnecessary WebSocket complexity;
- pnpm, Turborepo, Make, tests, and static analysis provide a coherent monorepo workflow.

However, the architecture’s suitability should not be confused with implementation completeness. The codebase currently proves the application and tooling foundation only. The business platform becomes real when each vertical slice implements and verifies its domain behavior, persistence, API contracts, security, audit, concurrency controls, asynchronous processing, frontend workflows, and operational monitoring.

## 15. Canonical References

This guide summarizes, but does not replace, the repository sources of truth:

- `README.md` — repository status and local usage;
- `docs/PRD.md` — product goals, actors, requirements, and business rules;
- `docs/PRODUCT_MAP.md` — capability decomposition, application ownership, roadmap, and open requirements;
- `docs/ARCHITECTURE.md` — runtime, module, data, API, real-time, security, and deployment design;
- `docs/DECISIONS.md` — accepted and superseded architecture decisions;
- `.codex/CURRENT_STATE.md` — implemented foundation and known gaps.

## 16. Estimated Delivery Timeline and Development Time Log

### 16.1 Estimate purpose

This section provides an engineering estimate, not a fixed delivery promise. The estimate is based on the scope documented in this repository and the current state in which the monorepo foundation exists but the Qurban business capabilities are not implemented.

The baseline answers this question:

> How much focused engineering time is required for one full-time developer to deliver, verify, pilot, and prepare the documented platform for an initial production event?

### 16.2 Baseline assumptions

The estimate assumes:

- one full-time, experienced full-stack developer;
- 40 logged engineering hours per week;
- sequential delivery through vertical slices;
- the current React, Go, PostgreSQL, OpenAPI, pnpm, and Turborepo stack remains in place;
- no major rewrite or migration to microservices;
- product owners resolve phase-specific open decisions before implementation reaches them;
- existing UI foundations can be reused, while production workflows still require design and usability refinement;
- normal development includes discovery, implementation, code review, tests, documentation, and defect correction;
- staging infrastructure and external provider test environments are available when integration work starts.

The estimate does not include waiting time caused by:

- unresolved stakeholder, religious, legal, finance, or privacy decisions;
- delayed Figma approval or field-workflow validation;
- payment, identity, messaging, email, or object-storage vendor procurement;
- app-store review, because the current applications are web PWAs;
- data cleanup or migration from unknown legacy spreadsheets;
- procurement and installation of event hardware, scanners, printers, or networking;
- production support after the first event enters normal operations.

### 16.3 Summary estimate

| Phase                               | Delivery scope                 | Engineer-weeks | Logged engineering hours |
| ----------------------------------- | ------------------------------ | -------------: | -----------------------: |
| Phase 0                             | Remaining platform foundation  |            6–8 |                  240–320 |
| Phase 1                             | Commerce foundation            |          12–16 |                  480–640 |
| Phase 2                             | Alternative purchasing         |          10–14 |                  400–560 |
| Phase 3                             | Livestock and allocation       |          10–14 |                  400–560 |
| Phase 4                             | Event-day operations           |          12–16 |                  480–640 |
| Phase 5                             | Distribution and reporting     |           8–12 |                  320–480 |
| Phase 6                             | Pilot and production readiness |            6–8 |                  240–320 |
| **Total before planning reserve**   | **Complete initial platform**  |      **64–88** |          **2,560–3,520** |
| **Total with 15% planning reserve** | **Management planning range**  |     **74–101** |          **2,944–4,048** |

The midpoint before reserve is approximately 76 engineer-weeks or 3,040 hours. For one full-time developer, the unreserved estimate is approximately 15–20 calendar months. The safer management range with reserve is approximately 17–23 calendar months.

The reserve covers ordinary uncertainty across a long project, including integration rework, field discoveries, production defects, and requirements clarification. It does not cover an intentional scope expansion.

### 16.4 Phase 0 — Remaining platform foundation

**Estimated duration:** 6–8 weeks  
**Time-log budget:** 240–320 hours  
**Cumulative timeline:** weeks 1–8

Although the repository bootstrap is complete, the production platform foundation is not. This phase establishes the cross-cutting controls required before financial and operational vertical slices are safe.

| Development log category                                            |       Hours |
| ------------------------------------------------------------------- | ----------: |
| Resolve foundation ADRs and contract conventions                    |       24–32 |
| Database migration, transaction, and repository foundations         |       48–64 |
| Authentication and permission-based authorization baseline          |       48–64 |
| Audit, idempotency, outbox, and worker foundations                  |       64–80 |
| Observability, CI validation, security hardening, and documentation |       56–80 |
| **Phase total**                                                     | **240–320** |

Expected deliverables:

- selected HTTP router and database migration tooling;
- API module registration conventions;
- business migration and transaction foundations;
- Storefront and Operations authentication baseline;
- backend permission enforcement;
- append-only audit records;
- request idempotency support;
- PostgreSQL outbox and recoverable worker processing;
- structured request correlation, metrics, and operational health signals;
- initial staging deployment pipeline.

Exit criteria:

- migrations can be applied and tested repeatably;
- authenticated public and Operations requests have defined boundaries;
- privileged commands are authorized and audited;
- retry-sensitive commands can be made idempotent;
- outbox work survives process restarts and retries;
- repository validation and staging smoke tests pass.

### 16.5 Phase 1 — Commerce foundation

**Estimated duration:** 12–16 weeks  
**Time-log budget:** 480–640 hours  
**Cumulative timeline:** weeks 7–24

This phase delivers the first complete business vertical slice:

```text
Qurban Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Basic Operations Dashboard
```

| Development log category                                           |       Hours |
| ------------------------------------------------------------------ | ----------: |
| Event and Offering domain behavior and persistence                 |      80–104 |
| Party, Purchase, Payment, and participant activation               |     112–144 |
| Public and Operations OpenAPI contracts and API handlers           |       48–64 |
| Storefront catalogue, checkout, payment, and tracking flow         |       72–96 |
| Operations verification workflow and dashboard projection          |       72–96 |
| Authorization, audit, tests, concurrency, observability, and fixes |      96–136 |
| **Phase total**                                                    | **480–640** |

Expected deliverables:

- annual Event lifecycle and event-scoped historical configuration;
- Offering catalogue, pricing snapshot, availability, and quota rules;
- explicit purchaser, payer, and participant relationships;
- canonical Common Purchase lifecycle;
- payment submission, verification, rejection, and traceable ledger entries;
- Sohibul Qurban activation after eligibility;
- Storefront purchase tracking;
- Operations payment-verification queue;
- basic read-optimized Operations dashboard;
- public and Operations API contracts, migrations, audit records, and automated tests.

Exit criteria:

- the common-purchase journey works end to end in staging;
- duplicate requests and payment retries do not create duplicate financial effects;
- quota contention produces one correct winner and explicit conflicts for stale requests;
- public APIs do not expose Operations-only data;
- every privileged payment change is authorized and audited;
- dashboard values can be rebuilt from authoritative records.

### 16.6 Phase 2 — Alternative purchasing

**Estimated duration:** 10–14 weeks  
**Time-log budget:** 400–560 hours  
**Cumulative timeline:** weeks 17–38

| Development log category                                     |       Hours |
| ------------------------------------------------------------ | ----------: |
| Saving Account, installment ledger, balance, and conversion  |     112–152 |
| Giveaway Program, funding, review, approval, and assignment  |     104–144 |
| Cross-channel Purchase and financial reconciliation behavior |       64–88 |
| Storefront Saving and Giveaway journeys                      |       48–72 |
| Operations Saving and Giveaway administration                |       40–56 |
| Authorization, privacy, audit, idempotency, tests, and fixes |       32–48 |
| **Phase total**                                              | **400–560** |

Expected deliverables:

- Saving Account lifecycle and append-oriented installment ledger;
- verified balance, funding status, and conversion eligibility;
- idempotent Saving-to-Purchase conversion;
- Giveaway Program, sponsor funding, candidate intake, review, and recipient assignment;
- duplicate-recipient and privacy controls;
- canonical Purchase creation from eligible Saving and Giveaway sources;
- cross-channel reporting and reconciliation.

Exit criteria:

- fully funded Saving Accounts remain distinct from converted Purchases;
- conversion is transactionally safe and cannot run twice;
- a Giveaway application cannot activate a Purchase before approved assignment;
- sponsor, applicant, payer, recipient, and Sohibul Qurban remain traceable independently;
- correction and approval actions are authorized and audited;
- approved price-lock, refund, transfer, expiry, and selection rules are covered by tests.

### 16.7 Phase 3 — Livestock and allocation

**Estimated duration:** 10–14 weeks  
**Time-log budget:** 400–560 hours  
**Cumulative timeline:** weeks 27–52

| Development log category                                    |       Hours |
| ----------------------------------------------------------- | ----------: |
| Livestock registry, inspection, readiness, and location     |      96–128 |
| Allocation lifecycle, capacity, release, and reassignment   |     112–152 |
| Public/Operations API contracts and handlers                |       48–64 |
| Operations livestock and allocation interfaces              |      80–112 |
| Safe batch operations, manifests, and exports               |       24–40 |
| Capacity/concurrency tests, audit, authorization, and fixes |       40–64 |
| **Phase total**                                             | **400–560** |

Expected deliverables:

- unique livestock operational identity;
- intake, classification, weight, health inspection, readiness, and location history;
- provisional and confirmed allocations;
- individual and shared-livestock capacity handling;
- release, reassignment reason, and immutable history;
- allocation manifests and safe bulk operational actions;
- Operations views for unallocated participants and available livestock.

Exit criteria:

- livestock cannot be allocated beyond its configured capacity;
- concurrent allocation attempts are transactionally safe;
- only eligible Purchases or participants can be allocated;
- reassignments preserve reasons, authorization, and audit history;
- manifests reconcile against authoritative participant and livestock records;
- load tests cover expected allocation volume and contention.

### 16.8 Phase 4 — Event-day operations

**Estimated duration:** 12–16 weeks  
**Time-log budget:** 480–640 hours  
**Cumulative timeline:** weeks 37–68

| Development log category                                    |       Hours |
| ----------------------------------------------------------- | ----------: |
| Event readiness and participant/livestock check-in          |      80–104 |
| Slaughter sessions, stations, queues, and milestones        |     112–144 |
| Outbox consumers, projections, polling, and SSE             |      88–120 |
| Fast Operations Web event-day interfaces                    |      88–120 |
| Incident, hold, exception, and degraded-mode behavior       |       48–64 |
| Load, concurrency, recovery, observability, and field tests |       64–88 |
| **Phase total**                                             | **480–640** |

Expected deliverables:

- event-readiness checks;
- participant and livestock check-in;
- slaughter sessions, stations, assignments, and queue transitions;
- queued, called, in-process, completed, held, and cancelled milestones;
- operational incident and exception records;
- read-optimized live projections;
- polling baseline and high-value SSE updates;
- projection freshness and outbox-backlog visibility;
- event-day runbooks and recovery procedures.

Exit criteria:

- contested queue transitions cannot silently overwrite each other;
- dashboard freshness generally remains under the documented 10-second target at expected load;
- reconnecting clients recover current state after missed updates;
- projection lag never changes authoritative queue or slaughter truth;
- field workflows have been exercised in a realistic event simulation;
- incident, hold, override, and recovery procedures are documented and tested.

This phase has the highest operational risk. A low-connectivity or offline mutation mode is not included unless the open product decision explicitly approves and designs it; adding that requirement can materially increase this phase.

### 16.9 Phase 5 — Distribution and reporting

**Estimated duration:** 8–12 weeks  
**Time-log budget:** 320–480 hours  
**Cumulative timeline:** weeks 49–80

| Development log category                                        |       Hours |
| --------------------------------------------------------------- | ----------: |
| Distribution planning, entitlement, preparation, and completion |      88–128 |
| Storefront and Operations distribution workflows                |       56–80 |
| Operational reports, projections, and exports                   |       64–96 |
| Certificates, evidence, document rendering, and object storage  |       40–72 |
| Historical archive, privacy, retention, and access behavior     |       24–40 |
| Tests, performance verification, reconciliation, and fixes      |       48–64 |
| **Phase total**                                                 | **320–480** |

Expected deliverables:

- distribution plans and prerequisites;
- portion or entitlement records;
- beneficiary, pickup, or delivery assignments;
- collection/delivery confirmation, proof, and exception handling;
- participant documents or certificates when approved;
- purchasing, payment, participant, livestock, allocation, slaughter, and distribution reports;
- annual event archive and historically stable read access.

Exit criteria:

- distribution cannot complete before its approved prerequisites;
- entitlements and beneficiaries follow approved product rules;
- completion reversals are privileged and audited;
- reports reconcile against transactional records;
- generated evidence and documents follow privacy and retention policies;
- archived events remain readable without inheriting later event configuration.

### 16.10 Phase 6 — Pilot and production readiness

**Estimated duration:** 6–8 weeks  
**Time-log budget:** 240–320 hours  
**Cumulative timeline:** weeks 57–88

| Development log category                                         |       Hours |
| ---------------------------------------------------------------- | ----------: |
| End-to-end regression, contract, and migration testing           |       56–72 |
| Security, privacy, backup, restore, and disaster-recovery checks |       48–64 |
| Performance, burst, contention, and endurance testing            |       40–56 |
| Production deployment, monitoring, alerting, and runbooks        |       48–64 |
| User acceptance, training support, pilot, and launch fixes       |       48–64 |
| **Phase total**                                                  | **240–320** |

Expected deliverables:

- production infrastructure and isolated secrets;
- backup and verified restore process;
- alerting for API errors, database pressure, outbox lag, payment failures, allocation conflicts, and event-day degradation;
- full regression and critical end-to-end test suite;
- realistic peak-load and event simulation results;
- operator runbooks, support escalation, and incident procedures;
- user acceptance testing and controlled pilot event;
- release, rollback, and post-event review process.

Exit criteria:

- production readiness review has no unresolved critical finding;
- backup restoration is demonstrated rather than assumed;
- expected peak traffic and concurrency targets pass with acceptable margins;
- operations staff complete realistic workflows during UAT;
- monitoring and incident response are exercised;
- release and rollback procedures are verified before the live event.

### 16.11 Sequential milestone timeline

The ranges below are cumulative for one developer working largely sequentially. The lower bound assumes timely decisions and limited rework; the upper bound allows normal discovery and integration complexity.

| Milestone                               | Earliest completion | Conservative completion |
| --------------------------------------- | ------------------: | ----------------------: |
| Platform foundation ready               |              Week 6 |                  Week 8 |
| Common commerce vertical slice ready    |             Week 18 |                 Week 24 |
| Saving and Giveaway ready               |             Week 28 |                 Week 38 |
| Livestock and Allocation ready          |             Week 38 |                 Week 52 |
| Event-day Operations ready              |             Week 50 |                 Week 68 |
| Distribution and Reporting ready        |             Week 58 |                 Week 80 |
| Pilot and production readiness complete |             Week 64 |                 Week 88 |
| Management commitment with 15% reserve  |             Week 74 |                Week 101 |

Phases may overlap at their boundaries for discovery, API design, UI design, and test preparation. The totals do not assume major parallel implementation by one person.

### 16.12 Team-size scenarios

Engineer-weeks describe effort; calendar duration depends on team composition. Adding people does not reduce time linearly because domain discovery, coordination, reviews, shared contracts, and integration remain sequential constraints.

| Delivery team              | Indicative calendar range  | Conditions                                                                                 |
| -------------------------- | -------------------------- | ------------------------------------------------------------------------------------------ |
| One full-time developer    | 64–88 weeks before reserve | Baseline used by this estimate                                                             |
| Two full-time developers   | Approximately 38–52 weeks  | Clear ownership split between backend/domain and frontend/experience, with shared testing  |
| Three full-time developers | Approximately 30–42 weeks  | Backend/domain, Storefront, and Operations ownership plus disciplined contract integration |

A separate product owner or domain expert does not replace engineering capacity, but can substantially reduce blocked time and rework by resolving policies before implementation.

### 16.13 Recommended development time-log format

Actual time should be logged against a phase, vertical slice, and work category so future estimates can be recalibrated.

Recommended categories:

```text
DISCOVERY       requirements, field observation, domain modeling, ADRs
DESIGN          API contracts, schema, state machines, UI and interaction design
BACKEND         Go domain, application, adapters, migrations, workers
STOREFRONT      public React routes, workflows, accessibility, PWA behavior
OPERATIONS      internal React workflows, dashboards, event-day interfaces
INTEGRATION     payment, identity, messaging, email, storage, documents
TEST            unit, integration, contract, end-to-end, concurrency, load
SECURITY        authentication, authorization, privacy, threat remediation
OBSERVABILITY   logs, metrics, alerts, traces, dashboards, runbooks
DEVOPS          CI, environments, deployment, backup, restore, rollback
DOCUMENTATION   technical, operational, support, and training documents
REWORK          defects, requirement changes, production or UAT corrections
```

Each logged entry should record:

- date;
- phase and vertical slice;
- category;
- issue or task reference;
- short description of completed work;
- hours spent;
- whether the work was planned, defect correction, or scope change;
- blocker or decision dependency when applicable.

Example:

```text
Date: 2026-09-14
Phase: 1 — Commerce foundation
Slice: Payment verification
Category: BACKEND
Task: Implement idempotent payment verification transaction
Hours: 6.5
Type: Planned
Blocker: None
```

At the end of each two-week iteration, compare:

- estimated versus logged hours;
- completed versus deferred acceptance criteria;
- planned work versus rework;
- unresolved decisions and blocked time;
- remaining phase forecast;
- defect and test-failure trends.

Reforecast a phase when its actual consumption reaches 50% of the budget or when an approved requirement materially changes. Do not hide scope growth inside the original estimate.

### 16.14 Schedule risks

The largest timeline risks are:

1. unresolved Offering, Saving, Giveaway, personal-slaughter, and Distribution rules;
2. authentication and payment-provider selection occurring late;
3. discovering event-day low-connectivity requirements only after the online workflow is built;
4. insufficient field testing before the annual event;
5. underestimating data privacy, audit retention, and financial reconciliation;
6. introducing microservices, Redis, or a message broker before measured need;
7. treating dashboard projections or frontend caches as transactional truth;
8. skipping concurrency and idempotency testing until production;
9. relying on one developer without scheduling review, leave, support, and recovery capacity;
10. targeting the live annual event without a smaller controlled pilot and operational rehearsal.

The delivery forecast should be reviewed after every phase. Actual phase logs, approved scope changes, defect rates, and stakeholder decision latency should replace these initial assumptions as evidence becomes available.
