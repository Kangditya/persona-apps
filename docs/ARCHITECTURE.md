# Architecture

## Qurban Commerce and Operations Platform

**Status:** Initial architecture baseline  
**Repository:** `Kangditya/persona-apps`  
**Architecture style:** Monorepo with web applications and a Go modular monolith

---

## 1. Architecture Context

The repository currently provides:

- `apps/storefront-web` — public Next.js React application;
- `apps/operations-web` — internal Next.js React application;
- `apps/api` — Go API shell;
- shared workspace packages;
- PostgreSQL local infrastructure;
- pnpm and Turborepo-based JavaScript tooling.

The initial system should evolve from this structure instead of introducing distributed services prematurely.

The target is a modular backend with clear domain ownership and separate public and internal API surfaces. Modules may later be extracted into independently deployable services only when operational, organizational, or scaling evidence justifies it.

---

## 2. Architectural Goals

- Preserve one authoritative domain model across Storefront and Operations.
- Support three purchasing channels without duplicating purchase logic.
- Maintain strong transactional integrity for payments, quota, and allocation.
- Provide near-real-time operational visibility.
- Support annual event isolation and historical reporting.
- Enable independent frontend development.
- Make external integrations replaceable.
- Permit future backend separation without requiring it initially.
- Keep local development and verification straightforward.

---

## 3. System Overview

```text
┌─────────────────────────┐
│     Storefront Web      │
│ Public purchasing flows │
└────────────┬────────────┘
             │ HTTPS / JSON
             ▼
┌───────────────────────────────────────────────────────────┐
│                         Go API                            │
│                                                           │
│  Public API Surface          Operations API Surface       │
│  /api/public/*               /api/operations/*            │
│          │                           │                    │
│          └────────────┬──────────────┘                    │
│                       ▼                                   │
│              Application Services                         │
│                       ▼                                   │
│                Domain Modules                             │
│                                                           │
│ Event · Identity · Offering · Purchasing · Payment        │
│ Saving · Giveaway · Participant · Livestock · Allocation  │
│ Slaughter · Distribution · Notification · Reporting       │
│                                                           │
│                       ▼                                   │
│ Repository Ports · Integration Ports · Event Outbox       │
└───────────────┬───────────────────────┬───────────────────┘
                │                       │
                ▼                       ▼
         ┌────────────┐        ┌─────────────────┐
         │ PostgreSQL │        │ External Systems │
         └────────────┘        │ payment/message  │
                               │ storage/email    │
                               └─────────────────┘
                ▲
                │ near-real-time read APIs or event stream
┌───────────────┴─────────────┐
│       Operations Web        │
│ Internal operational tools  │
└─────────────────────────────┘
```

---

## 4. Repository Target Structure

The target structure should remain close to the current monorepo:

```text
persona-apps/
├── apps/
│   ├── storefront-web/
│   ├── operations-web/
│   └── api/
│       ├── cmd/
│       │   ├── api/
│       │   └── worker/
│       ├── internal/
│       │   ├── platform/
│       │   ├── event/
│       │   ├── identity/
│       │   ├── offering/
│       │   ├── purchasing/
│       │   ├── payment/
│       │   ├── saving/
│       │   ├── giveaway/
│       │   ├── participant/
│       │   ├── livestock/
│       │   ├── allocation/
│       │   ├── slaughter/
│       │   ├── distribution/
│       │   ├── notification/
│       │   └── reporting/
│       ├── migrations/
│       └── go.mod
├── packages/
│   ├── typescript-config/
│   ├── api-client/
│   ├── contracts/
│   ├── ui/
│   └── validation/
├── docs/
│   ├── PRD.md
│   ├── ARCHITECTURE.md
│   ├── adr/
│   └── domain/
├── infra/
├── package.json
├── pnpm-workspace.yaml
└── turbo.json
```

The exact module list may evolve. Domain modules should be added only when they own meaningful behavior, not merely to create folders.

---

## 5. Application Boundaries

## 5.1 Storefront Web

### Responsibilities

- public event and offering discovery;
- common purchasing;
- saving account and installment interactions;
- giveaway application or recipient journeys;
- purchaser and participant data collection;
- payment initiation or evidence submission;
- purchase and event status;
- participant-facing notifications and documents.

### Restrictions

Storefront must not:

- contain authoritative business rules;
- directly access the database;
- expose internal operational fields;
- perform privileged status overrides;
- assume purchaser and Sohibul Qurban are always the same person.

---

## 5.2 Operations Web

### Responsibilities

- event configuration;
- purchase and payment administration;
- saving and giveaway operations;
- participant management;
- livestock, allocation, slaughter, and distribution;
- dashboards and exception queues;
- privileged corrections;
- audit and reporting.

### Restrictions

Operations Web must not:

- bypass backend authorization;
- calculate authoritative financial balances locally;
- mutate state through analytics endpoints;
- directly depend on database structure;
- silently overwrite contested operational updates.

---

## 5.3 Go API

The Go API is the system of record and owns:

- business invariants;
- authorization enforcement;
- transactions;
- persistence;
- integration orchestration;
- audit events;
- idempotency;
- domain status transitions;
- API contracts.

The initial backend is one deployable modular monolith with optional worker processes.

---

## 6. Backend Modularization

Each module should follow explicit layers without requiring excessive abstraction.

```text
internal/<module>/
├── domain/
│   ├── entity.go
│   ├── value_object.go
│   ├── policy.go
│   ├── errors.go
│   └── events.go
├── application/
│   ├── commands/
│   ├── queries/
│   └── ports.go
├── adapter/
│   ├── http/
│   ├── postgres/
│   └── integration/
└── module.go
```

A smaller module may use fewer files. Layer boundaries matter more than folder ceremony.

### Dependency Direction

```text
HTTP / Integration Adapters
            │
            ▼
Application Use Cases
            │
            ▼
Domain Model
```

Infrastructure implementations depend on domain or application ports. Domain code must not depend on HTTP frameworks, PostgreSQL drivers, or external provider SDKs.

---

## 7. Domain Boundaries

## 7.1 Event

Owns:

- annual event lifecycle;
- registration and operational windows;
- event status;
- event-level quota and configuration references.

Does not own:

- individual purchases;
- payment balances;
- livestock execution.

## 7.2 Identity

Owns reusable party identity:

- person;
- organization;
- contact points;
- deduplication references.

Role assignments belong to their contextual modules.

## 7.3 Offering

Owns:

- qurban package;
- species and category rules;
- price;
- participant capacity;
- publication and availability rules.

A purchase stores a pricing snapshot or immutable reference sufficient to preserve historical truth.

## 7.4 Purchasing

Owns the canonical purchase aggregate:

- channel;
- purchaser;
- payer reference;
- event;
- selected offering;
- participant intent;
- eligibility;
- purchase lifecycle.

It coordinates with channel-specific modules through application services.

## 7.5 Payment

Owns:

- payment intent;
- received payment;
- verification;
- refund;
- adjustment;
- provider reference;
- idempotency.

Payment data should be modeled as a ledger rather than mutable total-only fields.

## 7.6 Saving

Owns:

- saving account;
- target;
- installment schedule or policy;
- funding state;
- conversion eligibility;
- conversion reference.

Conversion creates or activates a canonical purchase through the Purchasing module.

## 7.7 Giveaway

Owns:

- program;
- sponsor or funding source;
- applicant or nominee;
- eligibility review;
- approval;
- recipient assignment.

Approval creates or activates a canonical purchase through the Purchasing module.

## 7.8 Participant

Owns:

- Sohibul Qurban records;
- participant display information;
- participant status;
- relationship to eligible purchases.

It must not infer payment completion independently.

## 7.9 Livestock

Owns:

- livestock identity;
- classification;
- intake;
- inspection;
- weight and readiness;
- physical location;
- operational lifecycle.

## 7.10 Allocation

Owns:

- participant or purchase allocation to livestock;
- capacity;
- provisional and confirmed assignment;
- reassignment history.

Allocation updates must be concurrency-safe.

## 7.11 Slaughter

Owns:

- event sessions;
- stations or teams;
- queue;
- execution milestones;
- hold and exception handling.

## 7.12 Distribution

Owns:

- entitlement;
- preparation;
- beneficiary or recipient;
- pickup or delivery status;
- completion evidence.

## 7.13 Notification

Owns:

- template selection;
- notification request;
- delivery attempt;
- provider status;
- deduplication.

Notification failures must not roll back completed domain transactions.

## 7.14 Reporting

Owns read-optimized projections and exports.

Reporting does not own transactional status and should be rebuildable from source records and events.

---

## 8. Purchasing Architecture

All channels converge into one canonical purchase model.

```text
Common Checkout ───────────────┐
                               │
Saving Account → Conversion ───┼──► Purchase ─► Sohibul Qurban
                               │
Giveaway Approval ─────────────┘
```

Recommended approach:

```go
type Channel string

const (
    ChannelCommon   Channel = "COMMON"
    ChannelSaving   Channel = "SAVING"
    ChannelGiveaway Channel = "GIVEAWAY"
)
```

Channel-specific source references should be explicit:

```text
purchase
├── channel
├── common_checkout_id?
├── saving_account_id?
└── giveaway_assignment_id?
```

Exactly one channel source should apply. Enforce this in application logic and, where practical, database constraints.

Do not implement three unrelated order tables with duplicated lifecycle rules unless a future requirement clearly demands it.

---

## 9. Data Architecture

## 9.1 Database

PostgreSQL is the initial transactional database.

Use one logical database for the modular monolith, with ownership conventions per module. Separate schemas may be introduced if they improve ownership without complicating migrations.

### Database Principles

- UUIDs for public identifiers;
- numeric or database-native keys may be used internally;
- exact numeric representation for money;
- UTC timestamps with explicit timezone handling;
- soft deletion only where genuine recovery or historical behavior requires it;
- unique constraints for business references;
- foreign keys for critical integrity;
- check constraints for stable invariants;
- immutable or append-only ledgers for financial events;
- status history for sensitive lifecycle transitions.

## 9.2 Event Isolation

Every event-scoped aggregate must carry an event reference.

Historical event data must remain stable even when:

- current prices change;
- offering definitions change;
- quota rules change;
- participant details are corrected;
- future event configuration changes.

Use snapshots or versioned references for historically significant configuration.

## 9.3 Transactions

Use database transactions for operations such as:

- quota reservation;
- purchase eligibility;
- payment verification;
- saving conversion;
- giveaway assignment;
- allocation confirmation;
- contested queue transitions.

Avoid distributed transactions. Use an outbox pattern for post-transaction integrations.

---

## 10. API Architecture

## 10.0 HTTP Server

Use `github.com/gin-gonic/gin` as the canonical HTTP framework and router at
the HTTP adapter/bootstrap boundary. Build the engine with `gin.New()`, attach
explicit middleware, and register separate public and Operations route groups.
The standard library remains the server/runtime foundation: `http.Server`,
`context.Context`, status constants, headers, cookies, and graceful shutdown
remain valid below or beside the Gin edge.

Gin must not cross into application, domain, repository, or persistence
packages. Route handlers map HTTP inputs to framework-neutral application
inputs and propagate `c.Request.Context()`.

Database migrations continue through the explicit
`golang-migrate/migrate/v4` command accepted by ADR-040. API startup never
runs migrations.

## 10.1 API Surfaces

Use separate route groups:

```text
/api/public/v1/*
/api/operations/v1/*
```

Potential future groups:

```text
/api/integrations/v1/*
/api/field/v1/*
```

Public and operations APIs may share application services but must have separate authorization, DTOs, and data exposure policies.

## 10.2 Contract Principles

- JSON over HTTPS;
- versioned route prefix;
- UUID-based public references;
- explicit pagination;
- deterministic filtering and sorting;
- structured errors;
- request correlation identifier;
- idempotency key for retry-sensitive commands;
- encrypted at-rest replay envelopes for responses containing a raw credential;
- no leaking database column names as accidental contracts;
- OpenAPI specification as the contract baseline.

## 10.3 Example Error Shape

```json
{
  "error": {
    "code": "PURCHASE_QUOTA_UNAVAILABLE",
    "message": "The selected qurban capacity is no longer available.",
    "details": {},
    "request_id": "req_..."
  }
}
```

## 10.4 Command and Query Separation

Strict CQRS is not required.

Use separate command and query handlers where it improves clarity:

- commands enforce invariants and mutate aggregates;
- queries serve purpose-built views;
- dashboards use read models rather than loading complex aggregate graphs.

---

## 11. Real-Time Operations

The Operations dashboard requires near-real-time updates, not necessarily hard real-time guarantees.

### Initial Approach

1. Domain transaction commits.
2. Outbox event is persisted in the same transaction.
3. Worker publishes or processes the event.
4. Dashboard projection is updated.
5. Operations Web receives updates using:
   - polling initially;
   - Server-Sent Events when live updates add value;
   - WebSocket only for genuinely bidirectional scenarios.

SSE is preferred over WebSocket for one-way dashboard updates due to lower complexity.

### Event Examples

- `PurchasePaid`
- `SavingAccountFullyFunded`
- `GiveawayRecipientApproved`
- `ParticipantActivated`
- `LivestockReady`
- `AllocationConfirmed`
- `LivestockQueued`
- `SlaughterCompleted`
- `DistributionCompleted`
- `OperationalIncidentRaised`

Events are internal contracts and should be versioned when consumed asynchronously.

---

## 12. Background Processing

Use a worker process from the same Go codebase:

```text
apps/api/cmd/worker
```

Initial responsibilities:

- outbox processing;
- notification delivery;
- payment callback processing;
- report generation;
- projection updates;
- retryable integration tasks.

A dedicated message broker is optional initially. A PostgreSQL-backed outbox and job queue may be sufficient for expected scale.

Introduce a broker only when throughput, delivery topology, or service extraction requires it.

---

## 13. Authentication and Authorization

### Storefront

Phase 1 browsing and checkout are guest-accessible. Purchase creation returns a
random opaque Purchase access token once and stores only its SHA-256 hash.
Tracking, cancellation, and evidence submission require that token as a Bearer
credential and remain scoped to one Purchase.

The Purchase token is not an operator identity, browser session, or permission
set. Public self-service accounts remain a later product decision.

### Operations

Use provider-neutral OpenID Connect Authorization Code flow with PKCE. The Go
API owns login, callback verification, operator mapping, and a revocable
server-side browser session.

- Discover the configured issuer and verify the ID token's issuer, audience,
  signature, and expiry with `github.com/coreos/go-oidc/v3/oidc`.
- Generate and validate state, nonce, and the PKCE verifier. Nonce comparison
  remains an explicit application check.
- Map the verified subject to `operator_users.external_subject`; reject
  unknown or inactive operators.
- Store only the SHA-256 hash of a random opaque session token. Send the raw
  token in a `Secure`, `HttpOnly`, `SameSite=Lax`, host-only cookie.
- Snapshot only allowlisted permissions. Authenticate and authorize every
  operations request in the backend.
- Require an allowed Origin and `X-CSRF-Token` for unsafe
  cookie-authenticated requests.
- Revoke the server session on logout and expire it no later than the verified
  identity session.

The API does not implement local passwords, password recovery, or MFA. Those
remain identity-provider responsibilities.

### Authorization Model

Use permissions rather than hard-coded frontend roles.

Phase 1 permission vocabulary:

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

Authorization must be enforced in application services or dedicated policy components, not only in HTTP middleware.

The complete login, session, CSRF, and Purchase-token contracts are documented
in docs/security/AUTHENTICATION.md. The authoritative lifecycle and permission
details are documented in docs/domain/COMMERCE_LIFECYCLES.md and
docs/security/PERMISSIONS.md.

---

## 14. Audit Architecture

Sensitive changes require an audit record containing:

- actor identity;
- actor role or permission context;
- request identifier;
- timestamp;
- action;
- target type and identifier;
- reason where required;
- relevant before and after values;
- source application;
- optional client and network metadata.

Audit records should be append-only and protected from normal operator modification.

Examples:

- payment verification override;
- saving balance adjustment;
- recipient approval;
- participant replacement;
- livestock reassignment;
- queue priority override;
- distribution completion reversal.

---

## 15. External Integrations

Each provider is implemented behind an application port.

```text
application port
      │
      ▼
provider adapter
```

Examples:

- `PaymentGateway`
- `BankReconciliation`
- `MessageSender`
- `EmailSender`
- `ObjectStorage`
- `DocumentRenderer`

Provider payloads must not become domain models.

Webhook handlers must:

- authenticate the sender;
- store provider identifiers;
- be idempotent;
- tolerate duplicates and reordering;
- acknowledge quickly;
- defer retryable processing to workers.

---

## 16. Frontend Architecture

Both web applications use Next.js 16 App Router, React, TypeScript, Tailwind
CSS, Vitest, TanStack Query, and the shared monorepo tooling already established
in the repository. They remain separate applications and initially keep their
existing interactive pages and API reads client-rendered for migration parity.

Recommended source structure:

```text
src/
├── app/
│   ├── layout.tsx
│   └── <route>/page.tsx
├── api/
├── routes/
│   └── paths.ts
├── features/
├── entities/
├── shared/
│   ├── api/
│   ├── components/
│   ├── hooks/
│   ├── validation/
│   └── utilities/
└── styles/
```

### Frontend Principles

- organize by feature rather than technical file type;
- generated or centralized API client;
- server state kept distinct from local UI state;
- use App Router filesystem entries for route registration and layout
  composition;
- keep centralized `src/routes/paths.ts` as the application-owned URL constant
  and dynamic URL-builder registry rather than scattering path strings;
- let TanStack Query own remote request lifecycle, cache updates, and
  invalidation after successful API commands;
- no duplicated domain validation as authoritative logic;
- route-level access control for operations;
- accessible components;
- explicit loading, empty, error, stale, and conflict states;
- test critical workflows at component and end-to-end levels.

A shared UI package should contain stable primitives, not application-specific pages.

### Frontend Data Boundaries

- Next.js App Router owns navigation and route composition. The compatibility
  migration does not move API reads or commands into Server Components, Route
  Handlers, Server Actions, or middleware authorization.
- Next.js server output is a delivery runtime, not a business backend. The Go
  API remains authoritative for contracts, authorization, transactions,
  idempotency, audit, and domain behavior.
- TanStack Query owns only remote API/server state. Forms and transient UI
  state remain feature or component-local unless a separate decision changes
  that boundary.
- Query keys use stable public identifiers and explicit event context. Public
  and operations data must use their respective OpenAPI contracts and DTOs.
- Query cache data is never transactional truth. Payment state, quota,
  allocation capacity, saving balance, and queue position are revalidated by
  the Go API for every contested command.
- A `409 Conflict` response is rendered as a conflict state; the frontend must
  not resolve it from stale cache data or override backend authorization.

---

## 17. Caching

Caching is optional and evidence-driven.

Initial approach:

- use HTTP caching for public static or slowly changing content;
- use in-process caching only for safe configuration;
- use TanStack Query cache only as a client-side view of remote data once it is
  implemented, with invalidation after successful commands and explicit stale
  state handling;
- avoid caching contested balances, quota, payment state, allocation capacity, or queue position as authoritative data;
- introduce Redis only when a concrete use case requires distributed cache, rate limiting, session storage, or short-lived coordination.

Database correctness must not depend on cache correctness.

---

## 18. Concurrency Control

High-risk contested resources include:

- offering quota;
- participant capacity;
- saving conversion;
- payment callback processing;
- livestock allocation;
- slaughter queue transitions.

Use one or more of:

- row-level locks;
- atomic conditional updates;
- unique constraints;
- version columns for optimistic concurrency;
- idempotency records.

The API should return a conflict response when a user acts on stale state.

---

## 19. Observability

### Logging

Use structured logs with:

- timestamp;
- severity;
- service and module;
- request or trace ID;
- actor ID where permitted;
- event ID;
- error code;
- latency.

Avoid logging secrets, payment evidence, or unnecessary personal data.

### Metrics

Technical:

- request rate, error rate, and latency;
- database connections and query latency;
- worker backlog;
- outbox age;
- integration failures.

Business:

- purchases by channel;
- pending payment verification;
- active and fully funded saving accounts;
- approved but unassigned giveaway recipients;
- quota utilization;
- unallocated participants;
- livestock readiness;
- slaughter throughput;
- distribution completion.

### Tracing

Add distributed tracing when asynchronous flow or extracted services make correlation difficult. It is optional for the first simple deployment.

---

## 20. Deployment Architecture

### Initial Deployment

A practical first deployment may contain:

- Storefront Web Next.js runtime;
- Operations Web Next.js runtime;
- Go API container;
- Go worker container;
- PostgreSQL;
- object storage;
- reverse proxy or managed ingress.

```text
Internet
   │
   ▼
CDN / Ingress
   ├── storefront domain ─► Storefront Next.js runtime
   ├── operations domain ─► Operations Next.js runtime
   └── API domain ────────► Go API
                                │
                         ┌──────┴──────┐
                         ▼             ▼
                     PostgreSQL      Worker
```

Operations access may be protected by additional network or identity controls.
Each web runtime is independently built and deployed. Ingress preserves
same-origin `/api`, `/health`, and `/ready` routing to the Go API; private API
origins are not exposed in browser bundles. Static assets may use a CDN, but
the applications are no longer static-only deployments.

### Environments

- local;
- development;
- staging;
- production.

Each environment must have isolated databases and secrets.

---

## 21. Verification Strategy

Repository-level validation should continue through the existing `make validate` workflow.

Target coverage:

### Backend

- domain unit tests;
- application service tests;
- repository integration tests;
- API contract tests;
- migration tests;
- concurrency and idempotency tests.

### Frontend

- component tests;
- route tests;
- API client contract tests;
- critical end-to-end journeys.

### Critical End-to-End Scenarios

- common purchase to participant activation;
- saving installments to purchase conversion;
- giveaway approval to participant activation;
- payment callback duplicate handling;
- quota contention;
- shared livestock capacity;
- reallocation;
- slaughter completion;
- distribution completion.

---

## 22. Architecture Evolution

Service extraction is not a goal by itself.

A module may become a separate service when one or more are true:

- independent scaling is repeatedly required;
- deployment cadence differs materially;
- security or compliance isolation is required;
- a dedicated team owns it;
- integration traffic overwhelms the modular monolith;
- availability boundaries justify isolation;
- data ownership can be made explicit without distributed transactions.

Likely future extraction candidates:

- notification;
- document generation;
- payment integration;
- reporting;
- live event projection.

Core purchasing, payment eligibility, quota, and allocation should remain transactionally close until there is strong evidence otherwise.

---

## 23. Architecture Decisions Required

ADR-040 and ADR-043 resolve migration tooling, HTTP routing, Phase 1
Storefront access, and Operations authentication. Remaining decisions include:

1. Public identifier strategy.
2. Money representation.
3. Outbox and background job implementation.
4. API contract generation.
5. Dashboard update transport: polling versus SSE.
6. Object storage provider.
7. Payment gateway integration.
8. Event configuration versioning.
9. Audit data retention.
10. Offline or degraded event-day operation.
11. Conditions for backend service extraction.

---

## 24. Guardrails

- Do not create separate business logic for each frontend.
- Do not equate Sohibul Qurban with purchaser or payer.
- Do not treat saving deposits as a completed purchase before conversion.
- Do not treat giveaway application as an approved purchase.
- Do not use floating point for money.
- Do not update financial history without traceability.
- Do not allocate beyond livestock capacity.
- Do not use analytics projections as transactional truth.
- Do not introduce microservices without a documented reason.
- Do not expose internal operations DTOs through public APIs.
- Do not let notification or reporting failures roll back valid domain transactions.

---

## 25. Current Baseline and Immediate Next Steps

The repository already has the correct high-level application split:

```text
apps/storefront-web
apps/operations-web
apps/api
```

Recommended next architecture work:

1. establish transaction boundaries and explicit module registration around
   the Gin route groups;
2. implement the accepted OIDC session, CSRF, permissions, and Purchase-token
   foundations;
3. implement Event, Identity, Offering, and Purchasing foundations;
4. define OpenAPI conventions;
5. establish outbox and audit foundations;
6. deliver the first vertical slice:
   common purchase → payment verification → Sohibul Qurban activation;
7. add operations dashboard projections after transactional records exist.

This sequence builds executable product capability while preserving the option to refine Figma flows and operational requirements.

---

## 26. Capability-to-Module Alignment

The product capability map is not a direct one-to-one folder mandate, but it defines the expected backend ownership boundaries.

```text
Product Capability              Backend Module
────────────────────────────────────────────────
Storefront                      transport/public + frontend
Purchasing                      purchasing
Party & Participant             identity + participant
Payment & Funding               payment + saving + giveaway
Livestock                       livestock
Allocation                      allocation
Event Operations                event + slaughter
Distribution                    distribution
Identity & Access               identity + platform/auth
Administration & Reporting      reporting + operations transport
```

### Rules

- A screen or dashboard is not automatically a domain module.
- Integration names such as `PaymentGateway` must remain adapter-level concepts.
- `ShoppingCart` is not a required aggregate until multi-offering checkout is confirmed.
- `Customer` must not replace role-specific party relationships.
- `Inventory` must not be used as the primary livestock abstraction.
- `OrderFulfillment` must not hide livestock allocation, slaughter, or distribution boundaries.

---

## 27. Canonical Product Documentation

The architecture now depends on four canonical documents:

```text
docs/
├── PRD.md
├── PRODUCT_MAP.md
├── ARCHITECTURE.md
└── DECISIONS.md
```

Their responsibilities are:

| Document          | Responsibility                                                         |
| ----------------- | ---------------------------------------------------------------------- |
| `PRD.md`          | Product goals, actors, requirements, rules, and success criteria       |
| `PRODUCT_MAP.md`  | Capability hierarchy, application mapping, roadmap, and open questions |
| `ARCHITECTURE.md` | Technical structure, boundaries, runtime, data, and integration design |
| `DECISIONS.md`    | Accepted and superseded architectural decisions                        |

`PRODUCT_MAP.md` is the source of truth for product capability decomposition. It must not be replaced by frontend navigation structure.

---

## 28. Delivery by Vertical Slice

Implementation should proceed through vertical slices rather than completing one technical layer across the entire product.

Example first slice:

```text
Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Operations Read Model
```

Each vertical slice should include:

- domain model;
- application commands and queries;
- persistence;
- public or operations API;
- frontend workflow where applicable;
- authorization;
- audit coverage;
- tests;
- observability.

This approach validates domain boundaries before expanding them.
