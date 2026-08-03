# Architecture Decision Register

## Qurban Commerce and Operations Platform

**Canonical location:** `docs/DECISIONS.md`  
**Status:** Active  
**Repository:** `Kangditya/persona-apps`

This register records accepted architecture and product-structure decisions that affect multiple applications or domains.

Each decision should be updated through a new ADR when its direction changes materially. Historical decisions should normally be marked `Superseded` rather than silently deleted.

---

## ADR-001: Use a monorepo

**Status:** Accepted

Use one repository for:

- Storefront Web;
- Operations Web;
- Go backend;
- shared frontend packages;
- API contracts;
- local infrastructure;
- product documentation;
- architecture documentation;
- agent instructions.

Current application structure:

```text
apps/
├── storefront-web/
├── operations-web/
└── api/
```

### Rationale

The project is currently maintained as one product with tightly related contracts and domain rules.

A monorepo keeps:

- frontend and backend changes aligned;
- API contracts versioned with implementations;
- shared tooling consistent;
- architectural documentation close to code;
- local development straightforward;
- cross-application refactoring visible.

### Consequences

- Repository-level validation must cover all applications.
- Shared packages must remain intentionally scoped.
- Application-specific code must not be moved into shared packages solely for convenience.
- Go correctness must not depend on JavaScript tooling.

---

## ADR-002: Use two frontend applications

**Status:** Accepted  
**Supersedes:** Previous POS-oriented interpretation of ADR-002

Use:

```text
apps/storefront-web
apps/operations-web
```

### Storefront Web

The public application supports:

- event and offering discovery;
- common purchasing;
- saving purchasing;
- giveaway participation;
- participant data entry;
- payment interaction;
- purchase tracking;
- participant-facing event information.

### Operations Web

The internal application supports:

- event configuration;
- purchase administration;
- saving and giveaway operations;
- payment verification;
- Sohibul Qurban management;
- livestock management;
- allocation;
- slaughter operations;
- distribution;
- reporting;
- privileged administration.

### Rationale

The two applications differ in:

- audience;
- authentication requirements;
- authorization model;
- data exposure;
- operational risk;
- deployment cadence;
- navigation and interaction design.

### Consequences

- Public and internal DTOs may differ.
- Operations-only fields must not leak into Storefront APIs.
- Business rules must remain in the backend rather than being duplicated across frontends.
- Shared UI packages should contain stable primitives, not application-specific pages.

---

## ADR-003: Use Vite and React

**Status:** Accepted

Both frontend applications use:

- Vite;
- React;
- TypeScript;
- React Router with Remix-style routing conventions.

### Rationale

The applications are API-driven and do not initially require server-side rendering.

Vite provides:

- fast local development;
- simple static deployment;
- direct compatibility with the current React stack;
- low framework coupling;
- independent deployment for each frontend.

### Consequences

- Public SEO requirements must be reassessed before they become critical.
- Server rendering should not be introduced without a documented requirement.
- Backend contracts remain independent of frontend framework details.
- Remix-style route hierarchy, layouts, route boundaries, navigation state, and
  route-data requirements are designed in the frontend while runtime delivery
  remains a Vite-served SPA. This does not adopt a Remix server runtime.

---

## ADR-004: Use central route registries

**Status:** Accepted

Each frontend application owns centralized route definitions.

Recommended structure:

```text
src/routes/
├── paths.ts
├── routes.tsx
└── guards.tsx
```

### Rationale

Central route ownership prevents duplicated path strings and inconsistent authorization behavior.

### Consequences

- URL builders must use centralized path definitions.
- Route guards must not replace backend authorization.
- Feature modules may contribute routes through explicit registration.
- Route modules must declare their public or operations API-surface ownership.
- Route data requirements must not expose operations-only DTOs through
  Storefront.

---

## ADR-005: Use Tailwind CSS

**Status:** Accepted

Use Tailwind CSS v4 through the Vite integration.

Initial setup should remain minimal:

```css
@import "tailwindcss";
```

### Rationale

Tailwind supports rapid interface development while the UI is being designed in Figma.

### Consequences

- Do not create a large design-token system prematurely.
- Reusable tokens should be introduced only when repeated design decisions become stable.
- Shared UI primitives should remain accessible and application-neutral.

---

## ADR-006: Use Oxlint as the primary frontend linter

**Status:** Accepted

Use Oxlint for JavaScript and TypeScript linting.

Use the TypeScript compiler separately for type validation.

### Consequences

- Introduce ESLint only when a required rule or plugin is unsupported.
- Linting and type checking remain separate repository validation steps.

---

## ADR-007: Use pnpm workspaces and Turborepo

**Status:** Accepted

Use pnpm for JavaScript package management.

Use Turborepo for workspace command orchestration.

### Consequences

- Shared JavaScript packages live under `packages/`.
- The Go backend remains a native Go module.
- Turborepo may orchestrate Go commands but must not become required for Go package correctness.

---

## ADR-008: Use a Go modular monolith

**Status:** Accepted

Use one initial Go backend under:

```text
apps/api
```

The backend is deployed as a modular monolith.

### Rationale

The core domains require strong transactional consistency across:

- purchasing;
- payments;
- quota;
- saving conversion;
- giveaway assignment;
- participant activation;
- livestock allocation.

Microservices would introduce distributed transactions and operational complexity before the product boundaries are proven.

### Consequences

- Modules must expose explicit application interfaces.
- Domain code must not depend on HTTP or PostgreSQL implementations.
- Direct cross-module database access is discouraged.
- Service extraction requires a documented ADR and measurable justification.

---

## ADR-009: Use PostgreSQL as the system of record

**Status:** Accepted

Use PostgreSQL as the authoritative transactional database.

### Rationale

The product requires:

- relational integrity;
- transactional updates;
- financial ledgers;
- event-scoped records;
- concurrency controls;
- audit history;
- reporting queries.

### Consequences

- Critical invariants should use database constraints where practical.
- Money must use integer minor units or exact decimal types.
- Floating-point values must not represent financial amounts.
- Database migrations must be versioned with the codebase.
- Analytics projections must not replace transactional records.

---

## ADR-010: Use separate OpenAPI contracts

**Status:** Accepted

Use separate API contracts for public and internal consumers:

```text
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
```

Possible future contract:

```text
contracts/openapi/integrations.yaml
```

### Rationale

Storefront and Operations expose different:

- authorization requirements;
- DTOs;
- workflows;
- sensitive fields;
- error conditions.

### Consequences

- Shared domain behavior may use different transport representations.
- Frontend clients may be generated once meaningful endpoints exist.
- Internal database structures must not become accidental API contracts.
- Public references should use stable UUID-based identifiers.

---

## ADR-011: Model the platform around annual Qurban Events

**Status:** Accepted

Use `QurbanEvent` as the top-level operational cycle.

Every event-scoped transactional aggregate must reference one event.

Examples:

- offerings;
- purchases;
- saving conversions;
- giveaway programs;
- Sohibul Qurban;
- livestock allocations;
- slaughter sessions;
- distribution records.

### Rationale

The program runs annually, while rules, pricing, capacity, dates, and operations may change between years.

### Consequences

- Historical records must not inherit future event configuration.
- Event-specific configuration must be snapshotted or versioned.
- Queries must use explicit event context.
- An archived event remains readable but normally immutable.

---

## ADR-012: Use three purchasing channels

**Status:** Accepted

The initial purchasing channels are:

```text
COMMON
SAVING
GIVEAWAY
```

### Definitions

- `COMMON`: direct purchasing and payment.
- `SAVING`: installments accumulated before conversion into a purchase.
- `GIVEAWAY`: sponsor-funded purchasing assigned to an approved recipient.

### Consequences

- Every canonical purchase belongs to exactly one channel.
- Channel-specific workflows may have separate aggregates.
- All successful channels converge into one canonical Purchase model.
- Adding another channel requires explicit eligibility and conversion rules.

---

## ADR-013: Use one canonical Purchase aggregate

**Status:** Accepted

Use one canonical Purchase model for all purchasing channels.

Conceptually:

```text
Common Checkout ───────────────┐
                               │
Saving Conversion ─────────────┼──► Purchase
                               │
Giveaway Assignment ───────────┘
```

### Rationale

The downstream lifecycle is shared:

- participant activation;
- quota consumption;
- payment or funding traceability;
- livestock allocation;
- event execution;
- reporting.

### Consequences

- Do not create three unrelated order implementations.
- Channel source references must be explicit.
- Exactly one purchasing-channel source applies to a purchase.
- Channel-specific data remains owned by its source module.
- Shared lifecycle rules belong to Purchasing.

---

## ADR-014: Separate Sohibul Qurban from financial actors

**Status:** Accepted

The following concepts must remain independent:

- purchaser;
- payer;
- saving-account holder;
- sponsor;
- applicant;
- nominee;
- giveaway recipient;
- Sohibul Qurban.

The same person may fulfill multiple roles, but the system must not assume this.

### Rationale

The three purchasing channels involve different relationships between funding and qurban participation.

### Consequences

- Sohibul Qurban registration is not a purchasing channel.
- Participant activation occurs only after channel eligibility is satisfied.
- APIs and database schemas must use role-specific references.
- Reporting must distinguish who funded, purchased, received, and performed qurban.

---

## ADR-015: Treat Saving as funding before purchase conversion

**Status:** Accepted

A Saving Account is not a completed Purchase.

Saving owns:

- target;
- installment ledger;
- balance;
- funding lifecycle;
- conversion eligibility.

A canonical Purchase is created or activated only after conversion.

### Consequences

- Installment entries must be traceable and append-oriented.
- Fully funded does not necessarily equal converted.
- Price-lock, refund, transfer, expiry, and overpayment policies remain configurable product decisions.
- Conversion must be idempotent and transactionally safe.

---

## ADR-016: Treat Giveaway approval as a source of purchase eligibility

**Status:** Accepted

A giveaway application or nomination is not itself a Purchase.

Giveaway owns:

- program;
- sponsor or funding source;
- candidate or nominee;
- review;
- approval;
- recipient assignment.

A canonical Purchase is created or activated only after approved assignment.

### Consequences

- Applicant, sponsor, and recipient may be different parties.
- Duplicate recipient controls are required.
- Giveaway data requires stricter privacy controls.
- Approval and assignment changes require audit history.

---

## ADR-017: Use ledger-based financial records

**Status:** Accepted

Represent payments, installments, refunds, transfers, and adjustments as traceable entries.

Do not rely only on a mutable balance field.

### Rationale

The platform must reconcile:

- common payments;
- saving installments;
- sponsor funding;
- refunds;
- corrections;
- conversion balances.

### Consequences

- Derived balances may be stored for performance but must be reconcilable.
- Financial corrections require compensating or adjustment entries.
- Sensitive financial changes require audit records.
- Duplicate provider callbacks must be idempotent.

---

## ADR-018: Separate public and operations API surfaces

**Status:** Accepted

Use route groups such as:

```text
/api/public/v1/*
/api/operations/v1/*
```

### Rationale

Public and internal users have different security and data-access boundaries.

### Consequences

- Shared application services may be reused.
- Transport DTOs and authorization policies remain separate.
- Internal override endpoints must never appear in the public contract.
- Public endpoints require rate limiting and safe error disclosure.

---

## ADR-019: Use permission-based authorization

**Status:** Accepted

Use permissions rather than relying only on hard-coded role names.

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

### Consequences

- Roles become collections of permissions.
- Authorization must be enforced in backend application policies.
- Frontend route guards improve usability but are not security controls.
- Event or location scope may be added to permission evaluation.

---

## ADR-020: Use an outbox for asynchronous side effects

**Status:** Accepted

Persist domain events to an outbox in the same transaction as authoritative state changes.

Use a worker process from the same Go codebase:

```text
apps/api/cmd/worker
```

Initial asynchronous responsibilities:

- notifications;
- projection updates;
- provider callbacks;
- report generation;
- retryable integration work.

### Rationale

External failures must not invalidate successful transactional changes.

### Consequences

- Consumers must be idempotent.
- Outbox age and failures must be observable.
- A dedicated message broker is not required initially.
- Broker adoption requires a documented operational need.

---

## ADR-021: Use projections for operational dashboards

**Status:** Accepted

The Operations dashboard uses read-optimized projections derived from authoritative transactional records.

### Rationale

Operational dashboards require aggregated and near-real-time views that should not force complex transactional queries.

### Consequences

- Projections are rebuildable.
- Projection lag must be visible.
- Dashboard data must not be used as the transactional source of truth.
- Commands must always operate against authoritative modules.

---

## ADR-022: Prefer polling first and Server-Sent Events when needed

**Status:** Accepted

Use:

1. polling for the first dashboard implementation;
2. Server-Sent Events for high-value one-way live updates;
3. WebSocket only when bidirectional real-time communication is required.

### Rationale

Polling and SSE provide adequate near-real-time behavior with lower complexity.

### Consequences

- Dashboard freshness targets should be explicit.
- Reconnection and missed-event behavior must be handled.
- Real-time transport must not carry authoritative state transitions without normal command validation.

---

## ADR-023: Preserve livestock allocation capacity transactionally

**Status:** Accepted

Allocation must prevent participant capacity from exceeding livestock or offering rules.

### Consequences

Use one or more of:

- database constraints;
- row-level locking;
- conditional updates;
- optimistic concurrency;
- idempotency records.

Reassignment and override actions require:

- authorization;
- reason;
- audit history.

---

## ADR-024: Maintain append-only audit records for privileged actions

**Status:** Accepted

Audit records must capture:

- actor;
- timestamp;
- permission context;
- request identifier;
- action;
- target;
- relevant before and after data;
- reason where required.

### Audited Examples

- payment verification override;
- installment adjustment;
- giveaway approval;
- participant replacement;
- livestock reassignment;
- queue priority override;
- distribution reversal.

### Consequences

- Normal operators cannot modify audit records.
- Sensitive data in audit payloads must be minimized.
- Retention policy remains a documented product and compliance decision.

---

## ADR-025: Design for one operating organization initially

**Status:** Accepted  
**Supersedes:** Previous SaaS and shared-schema multitenancy decisions

The initial system supports one operating organization running one or more annual qurban events.

Do not implement generic SaaS multitenancy during the initial product phases.

### Rationale

Current requirements describe one qurban operation, not a validated multi-tenant SaaS platform.

### Consequences

- No mandatory `organisation_id` is added to every table solely for hypothetical SaaS support.
- Ownership boundaries should remain clear enough to support future tenant design.
- Multi-organization support requires a new ADR and data migration strategy.
- Security must still isolate public, operational, and privileged access.

---

## ADR-026: Keep external providers behind ports and adapters

**Status:** Accepted

External systems include:

- payment gateways;
- bank reconciliation;
- messaging;
- email;
- object storage;
- document rendering;
- identity providers.

### Consequences

- Provider payloads must not become domain models.
- Webhook processing must authenticate providers and remain idempotent.
- Provider replacement should not require rewriting domain logic.
- Manual operational workflows may exist before provider automation.

---

## ADR-027: Do not introduce microservices without evidence

**Status:** Accepted

A module may be extracted only when justified by measurable needs such as:

- materially different scaling;
- independent release cadence;
- security or compliance isolation;
- dedicated ownership;
- availability boundaries;
- integration throughput;
- operational failure isolation.

### Consequences

- Service extraction requires a new ADR.
- Data ownership and consistency strategy must be documented first.
- Core purchase, payment eligibility, quota, and allocation should remain transactionally close during initial development.

---

## ADR-028: Keep canonical documentation under `docs/`

**Status:** Accepted

Canonical product and architecture documentation lives under:

```text
docs/
├── PRD.md
├── ARCHITECTURE.md
├── DECISIONS.md
├── adr/
└── domain/
```

The `.codex/` directory is reserved for agent operating context and task execution artifacts.

Recommended `.codex/` contents:

```text
.codex/
├── AGENTS.md
├── CURRENT_STATE.md
└── TASK.md
```

### Consequences

- `.codex/DECISIONS.md` has been removed after migration.
- Agent instructions may reference `docs/DECISIONS.md`.
- Product and architecture decisions remain readable independently of a specific coding agent.

---

## ADR-029: Keep the repository publishable but explicitly incomplete

**Status:** Accepted  
**Supersedes:** Previous commerce/POS feature list

The repository may remain public during early development, but documentation must distinguish implemented capabilities from intended capabilities.

Initial missing or incomplete areas may include:

- authentication;
- authorization;
- qurban events;
- offerings;
- common purchasing;
- saving purchasing;
- giveaway purchasing;
- payment processing;
- Sohibul Qurban management;
- livestock management;
- allocation;
- slaughter operations;
- distribution;
- production deployment.

### Consequences

- README claims must match current implementation.
- Architecture documents describe intended direction, not completed functionality.
- Secrets, production data, and sensitive operational information must never be committed.

---

## ADR-030: Do not distort architecture solely for zero-cost hosting

**Status:** Accepted

The project should support low-cost local development and practical early deployment.

Production requirements such as:

- availability;
- backups;
- monitoring;
- domains;
- email;
- payment processing;
- object storage;
- incident response;

may require paid infrastructure.

### Consequences

- Provider cost is a decision factor, not the sole architecture driver.
- Critical data integrity and recoverability take precedence over permanent free-tier constraints.

---

## Superseded Decisions Summary

The previous decision register described a generic single-brand commerce platform with POS and future SaaS multitenancy.

The following previous interpretations are superseded:

| Previous Decision                                                    | Replacement                                                                                                         |
| -------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| Operations Web combines POS and back office                          | ADR-002: Qurban internal operations application                                                                     |
| Future SaaS is an accepted baseline                                  | ADR-025: One operating organization initially                                                                       |
| Shared-schema multitenancy is planned by default                     | ADR-025: Defer tenancy until validated                                                                              |
| Catalogue, inventory, POS, and billing are primary product domains   | Qurban Event, Purchasing, Saving, Giveaway, Participant, Livestock, Allocation, Slaughter, and Distribution domains |
| Sohibul Qurban registration is treated as a primary acquisition flow | ADR-012 through ADR-016: registration is an outcome of eligible purchasing                                          |

---

## Decision Process

A new ADR should be added when a decision:

- changes a cross-module contract;
- introduces infrastructure;
- affects security or data integrity;
- changes deployment boundaries;
- changes an accepted product-domain relationship;
- creates a long-term operational constraint.

Use this format:

```markdown
## ADR-XXX: Decision title

**Status:** Proposed | Accepted | Superseded | Rejected

### Context

Describe the problem and constraints.

### Decision

Describe the selected direction.

### Consequences

Describe benefits, costs, risks, and required follow-up.
```

---

## ADR-031: Use a capability map as the product decomposition baseline

**Status:** Accepted

Use `docs/PRODUCT_MAP.md` as the canonical hierarchy for product capabilities, application mapping, roadmap phases, and unresolved capability-level questions.

### Decision

The top-level capabilities are:

```text
Storefront
Purchasing
Party & Participant
Payment & Funding
Livestock
Allocation
Event Operations
Distribution
Identity & Access
Administration & Reporting
```

### Consequences

- Product epics should map to capabilities or meaningful sub-capabilities.
- Frontend navigation is not the canonical domain decomposition.
- A dashboard, screen, gateway, or adapter must not automatically become a domain.
- Architecture and PRD changes must remain consistent with the Product Map.

---

## ADR-032: Use Qurban Offering rather than Animal Catalogue as the commercial abstraction

**Status:** Accepted

Use `Offering` as the public commercial abstraction.

An Offering may represent:

- an individual livestock unit;
- a category;
- a package;
- a cattle share;
- a saving target;
- another event-specific qurban product.

### Consequences

- Storefront naming should use `Offering Catalogue`.
- Livestock remains a separate physical operational domain.
- Offering-to-livestock relationships may be resolved during allocation rather than catalogue browsing.
- Exact supported offering types remain a product decision.

---

## ADR-033: Do not require a Shopping Cart until multi-offering checkout is confirmed

**Status:** Accepted

A Shopping Cart is optional and must not be introduced as a core aggregate solely because the product resembles commerce.

### Consequences

- Initial common purchasing may use direct checkout.
- Multi-offering checkout requires a new or updated requirement.
- Purchase and payment design must not depend on cart existence.

---

## ADR-034: Treat Livestock as an operational lifecycle, not generic inventory

**Status:** Accepted

Livestock owns physical identity, inspection, readiness, location, allocation state, slaughter progression, and history.

### Consequences

- Generic inventory terminology should not define the core model.
- Stock quantity alone is insufficient.
- Livestock and Offering remain separate abstractions.
- Health, readiness, and physical status are domain concerns.

---

## ADR-035: Treat Allocation as a first-class domain

**Status:** Accepted

Allocation owns the relationship between eligible purchases or Sohibul Qurban and livestock capacity.

### Consequences

- Allocation must not be hidden inside generic order fulfillment.
- Capacity validation must be transactional.
- Provisional assignment, confirmation, release, and reassignment require explicit states.
- Reassignment and override actions require audit history.

---

## ADR-036: Treat Event Operations as a domain and dashboards as projections

**Status:** Accepted

Event Operations includes readiness, check-in, scheduling, slaughter queue, stations, status tracking, incidents, and completion.

The operational dashboard is a projection over these domains.

### Consequences

- Dashboard widgets do not own transactional state.
- Commands remain in authoritative modules.
- Real-time updates may use polling or SSE according to ADR-022.
- Operational incidents require explicit records rather than free-form dashboard notes only.

---

## ADR-037: Separate Distribution from order delivery

**Status:** Accepted

Distribution is a qurban-specific operational capability that may include:

- Sohibul Qurban entitlement;
- beneficiary assignment;
- portion preparation;
- pickup;
- delivery;
- proof;
- completion.

### Consequences

- Generic e-commerce delivery must not define the domain.
- Final entitlement and beneficiary rules remain open.
- Distribution may begin only after relevant slaughter and preparation conditions are met.

---

## ADR-038: Deliver the product through vertical slices

**Status:** Accepted

Prioritize executable business slices across backend, APIs, frontend, persistence, and verification.

The first recommended slice is:

```text
Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Basic Operations Dashboard
```

### Consequences

- Avoid implementing all database tables or all frontend shells before one complete flow works.
- Every slice must include authorization, audit, validation, and tests appropriate to its risk.
- Phase ordering may change based on discovery, but capability boundaries remain stable.

---

## ADR-039: Use TanStack Query for frontend server state

**Status:** Accepted

Use `@tanstack/react-query` for remote API/server state when the first
frontend vertical slice introduces real API reads or commands. React Router
continues to own navigation; TanStack Query owns request lifecycle, caching,
invalidation, and explicit server-state rendering.

### Rationale

The two applications need consistent handling for asynchronous API data
without treating component state or browser caches as domain truth. TanStack
Query provides a focused boundary between remote state and local UI state
while keeping the Go API authoritative.

### Consequences

- This decision does not add a dependency or runtime provider until a concrete
  vertical slice needs remote data.
- Query keys must use stable public API identifiers and explicit event context.
- Successful commands invalidate or update relevant query data only after a
  successful API response; cache invalidation is not a correctness mechanism.
- The UI must render loading, empty, error, stale, and `409 Conflict` states
  explicitly.
- Cache data must not authoritatively determine payment status, quota,
  allocation capacity, saving balance, or queue position; contested operations
  are always revalidated by the Go API.
- This decision does not adopt TanStack Router, Table, Form, or other TanStack
  libraries.
