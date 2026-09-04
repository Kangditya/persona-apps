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

**Status:** Superseded by ADR-047

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

**Status:** Accepted, amended by ADR-047

Each frontend application owns centralized URL definitions and route files.

Recommended structure:

```text
src/routes/
└── paths.ts

src/app/
├── layout.tsx
└── <route>/page.tsx
```

### Rationale

Central route ownership prevents duplicated path strings and inconsistent authorization behavior.

### Consequences

- URL builders must use centralized path definitions.
- Next.js App Router filesystem entries register routes; do not duplicate them
  in a parallel route table.
- Route guards must not replace backend authorization.
- Route modules must declare their public or operations API-surface ownership.
- Route data requirements must not expose operations-only DTOs through
  Storefront.

---

## ADR-005: Use Tailwind CSS

**Status:** Accepted, integration amended by ADR-047

Use Tailwind CSS v4 through the framework-supported PostCSS integration.

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

---

## ADR-040: Use an explicit Go-owned database lifecycle command

**Status:** Accepted

Use `golang-migrate/migrate/v4` from `apps/api/cmd/db` to execute the existing
numbered PostgreSQL migration pairs. Keep migration execution, inspection,
creation, and bounded rollback as explicit commands; do not run them from API
startup.

Seed execution is a separate ordered registry under
`apps/api/internal/database/seeder`. Reference and development seeds use
separate groups and independent `schema_seeds` history. Development seeds are
allowed only in development/test, while staging/production `--all` selects
reference seeds only. Rollback outside development/test requires an explicit
`ALLOW_DESTRUCTIVE_DB_COMMANDS=true` opt-in.

### Rationale

The existing `NNNN_name.up.sql`/`.down.sql` files already match
`golang-migrate`'s PostgreSQL source format. Reusing them avoids a second SQL
engine and preserves historical migration files. Separate seed history keeps
deterministic bootstrap data independent from schema version state.

### Consequences

- `schema_migrations` and PostgreSQL advisory locking provide migration state
  and serialization.
- `schema_seeds` records successful seed names only; changed seed definitions
  require a new immutable name because checksums are not needed yet.
- `db setup` applies migrations and reference seeds only.
- Dirty-version recovery (`force`) and arbitrary navigation (`goto`) remain
  deferred until an operational recovery policy and disposable-DB verification
  exist.

---

## ADR-041: Use command-scoped idempotency with domain-specific duplicate guards

**Status:** Accepted

Retry safety belongs to the command that owns a business effect. Do not add an
`idempotency_key` column to every aggregate, history table, audit record, or
outbox event.

Use `idempotency_records` as the shared replay ledger for retry-sensitive API
commands. Its `(namespace, idempotency_key)` primary key identifies one caller
intent, while `request_hash` prevents the same key from being reused for a
different request. The namespace must identify the API surface, command, and
stable caller scope without introducing hypothetical tenancy.

### Command replay contract

- Authenticate and authorize the caller before returning a stored response.
- The first request inserts its idempotency record and performs the domain
  mutation in one PostgreSQL transaction.
- Domain state, status history, audit records, outbox events, and the replayable
  response are committed together.
- The same namespace, key, and request hash returns the stored status and body
  without executing the command again.
- The same namespace and key with a different request hash returns a conflict.
- The primary-key conflict serializes concurrent duplicates; after the winning
  transaction commits, the duplicate reads and replays its result. A rolled
  back transaction leaves no completed replay record.
- Transient infrastructure failures are not stored as completed outcomes.
- Retention is command-specific and must not expire a key while a duplicate
  business effect would still be unacceptable.

### Domain-specific guards

- `financial_ledger_entries.idempotency_key` remains the direct effect-level
  guard for append-only financial writes. A command producing multiple entries
  derives a unique entry key for each leg from the command key.
- Provider references, purchase-source uniqueness, participant sequence,
  active-allocation constraints, current-location constraints, and one-record
  operational constraints remain natural duplicate guards.
- Version columns, row locks, and atomic conditional updates handle stale or
  contested state; they complement idempotency rather than replace it.
- Status histories and `audit_log` do not receive independent idempotency keys.
  They are written once inside the owning command transaction.
- `outbox_events.id` is the delivery identity. Each consumer deduplicates by
  event and consumer, or uses a naturally idempotent projection update. HTTP
  response replay records are not reused as a generic consumer inbox.
- Exact payment-webhook inbox fields remain deferred until a provider contract
  defines the provider event identity, authentication, and reordering rules.

### Consequences

- Public and operations OpenAPI contracts expose `Idempotency-Key` only for
  retry-sensitive commands, not every mutation.
- A frontend or integration reuses one key for retries of the same user intent
  and generates a new key for a new intent.
- Business references are not treated as response-replay keys unless the
  command contract explicitly defines them that way.
- The existing schema is sufficient for the generic transactional replay
  pattern and append-only ledger guard; this decision does not modify a
  historical migration.
- Provider inboxes, notification delivery attempts, and per-consumer receipts
  are added only with the vertical slice that owns their behavior.

---

## ADR-042: Fix Phase 1 common-purchase commerce rules

**Status:** Accepted

Phase 1 common purchasing uses these rules:

- Event lifecycle is
  `DRAFT -> PUBLISHED -> ACTIVE <-> SUSPENDED -> CLOSED -> ARCHIVED`. Only
  `PUBLISHED` becomes active; active events may be suspended or closed;
  suspended events may be reactivated or closed; closed and archived events
  cannot accept commerce commands. At most one Event is active.
- An MVP Offering is an event-scoped sellable package, share, or category and
  remains separate from physical Livestock.
- One direct checkout creates one Purchase for one Offering. No Shopping Cart
  or purchase-item aggregate is introduced.
- Checkout snapshots Offering identity, price, currency, participant capacity,
  and intended participant names.
- Quota is participant units against both Event and Offering limits. Checkout
  atomically creates a pending Purchase and a 24-hour reservation under row
  locking or an equivalent atomic database guard.
- Submitted payment evidence pauses reservation expiry until review.
  Activation consumes the reservation. Expiry, cancellation, and rejection
  release it. Resubmission after release must reacquire quota atomically.
- Common-purchase evidence is append-oriented and stored as a private
  object-storage reference with filename, media type, byte size, and SHA-256
  metadata. PostgreSQL does not store evidence bytes.
- Evidence accepts JPEG, PNG, or PDF up to 10 MiB and must declare the exact
  outstanding amount. Lower or higher submissions are rejected; partial-payment
  and balance policy is deferred.
- Finance or Operations Managers verify or reject evidence. Verification locks
  Payment, Purchase, and quota; revalidates amount, currency, status, and
  capacity; then atomically marks Payment `VERIFIED`, records Purchase `PAID`
  and `ELIGIBLE`, consumes quota, activates intended Sohibul Qurban exactly
  once, and writes audit and outbox effects.
- Rejection records a reason, releases quota, and leaves the Purchase pending.
  A later evidence attempt must reacquire quota.

### Consequences

- W1-03 must add an additive schema migration for Offering quota, intended
  participants, quota reservation attempts, evidence metadata, and required
  integrity constraints without editing historical migrations.
- W1-04 must define complete transition and permission matrices.
- Retry-sensitive commands follow ADR-041. Participant activation additionally
  uses natural uniqueness on `(purchase_id, sequence_no)`.
- Storefront availability is advisory; checkout and quota reacquisition are
  authoritative transactional commands.
- Saving, Giveaway, refunds, provider callbacks, evidence retention,
  multi-offering checkout, and livestock allocation remain separate decisions.

---

## ADR-043: Reuse standard HTTP and add OIDC-backed operations sessions

**Status:** Accepted; its router choice is superseded by ADR-046

### Decision

The HTTP-router choice in this ADR is superseded by ADR-046. The authentication,
session, CSRF, Purchase-token, and migration decisions below remain accepted.

- Keep ADR-040's `golang-migrate/migrate/v4` command, numbered SQL pairs, and
  separate seed lifecycle. API startup never runs migrations, and migrations
  `0001` through `0004` remain historical files.
- Phase 1 Storefront browsing and checkout remain guest-accessible. Purchase
  creation returns a random opaque Purchase access token once, stores only its
  SHA-256 hash, and requires the raw token as a Purchase-scoped Bearer
  credential for tracking, cancellation, and evidence submission.
- Operations uses provider-neutral OpenID Connect Authorization Code flow with
  PKCE. The Go API uses `github.com/coreos/go-oidc/v3/oidc` for provider
  discovery and ID-token verification and `golang.org/x/oauth2` for
  Authorization Code and PKCE exchange.
- The API explicitly validates state, nonce, and PKCE; maps the verified
  single-issuer `sub` claim to `operator_users.external_subject`; and denies
  unknown or inactive operators.
- Successful login creates a random opaque server session. Only its SHA-256
  hash is stored; the raw token is sent in a `Secure`, `HttpOnly`,
  `SameSite=Lax`, host-only cookie. Sessions are revocable and expire no later
  than the verified identity session.
- Only allowlisted permission claims are snapshotted into the session. Backend
  policies authenticate and authorize each request before command idempotency
  lookup or replay.
- Unsafe cookie-authenticated requests require an explicitly allowed Origin and
  a matching `X-CSRF-Token`; only the CSRF token hash is stored.
- Local passwords, password reset, account recovery, MFA implementation, and
  provider administration remain outside the API. MFA may be required by the
  configured identity provider.

### Consequences

- W1-03 must add session and Purchase-token hash storage through a new additive
  migration; this decision does not edit existing migrations.
- W1-06 adds the selected Go dependencies and runtime behavior. They are not
  added during this documentation task.
- Provider issuer, client credentials, redirect URL, permission claim, allowed
  origins, cookie policy, and encryption keys are validated environment
  configuration, never committed values.
- Supporting multiple OIDC issuers requires a new identity-key decision because
  the current operator mapping assumes one configured issuer.
- Public accounts, Purchase-token recovery/rotation, operator provisioning,
  permission administration, and event-scoped permissions remain deferred.

---

## ADR-044: Define Phase 1 lifecycle and permission policy

**Status:** Accepted

### Decision

Commerce transitions, authorization requirements, guards, audit effects,
outbox effects, retry handling, and rejection outcomes are authoritative in
docs/domain/COMMERCE_LIFECYCLES.md. This policy implements the lifecycle
rules accepted in ADR-042 without adding runtime behavior.

The complete Phase 1 permission vocabulary is event.read, event.manage,
offering.read, offering.manage, purchase.read, payment.read, payment.verify,
participant.read, dashboard.read, audit.read, and admin.manage. Operations
roles are provisioned with the fixed grants listed in
docs/security/PERMISSIONS.md; no role management endpoint is introduced.

OIDC login, callback, session inspection, and logout use the authentication
and CSRF controls from ADR-043 rather than a business permission. Every
other operations endpoint requires its mapped permission after session,
Origin, and CSRF checks, and before idempotency replay. Storefront requests
remain role-free and may use only the scoped Purchase access token accepted
by ADR-043.

Payment verification and rejection retain ADR-042's transactional behavior.
Participant replacement or cancellation remains deferred; until a dedicated
permission is accepted, any exceptional Phase 1 participant write requires
admin.manage and an auditable reason.

### Consequences

W1-05 must expose only endpoints whose authorization maps to this vocabulary.
W1-06 implements the documented enforcement and command behavior. This ADR
does not add tables, Go dependencies, routes, or a generic idempotency store.

---

## ADR-045: Encrypt sensitive idempotent replay responses at rest

**Status:** Accepted

### Decision

Some successful retry-sensitive commands return a raw credential that is
intentionally unavailable from later read endpoints. Common-purchase checkout
returns the opaque Purchase Bearer token once, while ADR-041 requires an exact
retry to replay the committed response. Storing that response body as plaintext
would persist the raw token and violate ADR-043.

Store every replayable response body that contains raw credential material as
an AES-256-GCM encrypted envelope in `idempotency_records.response_body`. The
envelope records a key identifier, nonce, and ciphertext; it never stores a
plaintext response body or raw credential. Additional authenticated data binds
the command namespace, idempotency key, request hash, and response status.

`IDEMPOTENCY_RESPONSE_KEYS` is an ordered, secret key ring of
`key-id:base64-32-byte-key` values. The first key encrypts new records; all
configured keys may decrypt retained records. Key removal is allowed only
after every record encrypted with that key has expired. A same-request replay
whose retained key is unavailable returns `idempotency_conflict` and does not
execute the command again.

### Consequences

- The existing JSONB column is sufficient; no migration is needed.
- W1-06 implements encryption, decryption, key validation, and tests for
  tampering, rotation, and unavailable retained keys.
- Raw Purchase tokens remain absent from database plaintext, logs, audit data,
  errors, and read endpoints.
- Non-sensitive replay bodies may use the same envelope format so one
  idempotency decoder handles all successful responses.

---

## ADR-046: Use Gin as the canonical backend HTTP router

**Status:** Accepted

### Decision

Use `github.com/gin-gonic/gin` as the canonical HTTP framework and router for
the Go API. Gin is confined to the HTTP adapter and bootstrap boundary:

- `gin.Engine` owns route registration, method/path matching, route groups,
  request binding, middleware composition, and HTTP response rendering;
- public and Operations surfaces register through separate canonical route
  groups;
- module HTTP adapters own their concrete endpoint declarations;
- application, domain, repository, and persistence packages remain framework
  neutral and receive `context.Context`, not `*gin.Context`;
- `net/http` remains valid for `http.Server`, transport types, status
  constants, headers, cookies, `httptest`, request contexts, and graceful
  shutdown.

The engine uses explicit middleware with `gin.New()`; it does not adopt
`gin.Default()` or duplicate repository-owned request logging.

### Consequences

- ADR-043's first router bullet is superseded; its authentication and
  persistence decisions remain unchanged.
- The API keeps `/api/public/v1` and `/api/operations/v1` route boundaries.
- Gin is a new API-module dependency; no database migration or OpenAPI
  contract change is required.
- Future endpoints must register through the appropriate route group and
  preserve the public/Operations contract boundary.

---

## ADR-047: Use Next.js App Router for both web applications

**Status:** Accepted

### Decision

Migrate `apps/storefront-web` and `apps/operations-web` in place from Vite and
React Router to the current stable Next.js 16 App Router. The applications
remain separate workspace packages and independently runnable deployments as
required by ADR-002.

Use a compatibility-first migration:

- App Router filesystem entries own route registration while each
  application's `src/routes/paths.ts` remains the central URL constant and
  dynamic URL-builder registry;
- existing interactive pages, TanStack Query hooks, forms, and session flows
  remain Client Components and browser-side API consumers until a separate
  decision justifies server rendering;
- the Go API remains the sole authoritative backend; do not add Next.js Route
  Handlers, Server Actions, middleware authorization, direct database access,
  or duplicated business rules;
- local Next.js rewrites preserve the same-origin `/api`, `/health`, and
  `/ready` paths; deployed routing is owned by ingress or reverse proxy;
- browser-visible configuration uses `NEXT_PUBLIC_*`, while the optional
  local API proxy target is server-only;
- Tailwind CSS 4 uses its PostCSS integration;
- each app owns a typed manifest and a production-only service worker that may
  cache immutable framework assets, icons, and a data-free offline fallback,
  but never API, authentication, participant, financial, or operational data.

Initial pages remain client-rendered for parity. Server Components may provide
route and layout boundaries, but SSR/RSC data loading, SEO optimization,
revalidation policy, and middleware guards require separate evidence and work.

### Consequences

- ADR-003 is superseded. React and TypeScript remain; Vite and React Router are
  removed from both application runtimes.
- ADR-004 is amended: `paths.ts` remains centralized, while `src/app/**`
  filesystem routes replace `routes.tsx` registration.
- ADR-005 is amended only in its framework integration mechanism; Tailwind CSS
  4 remains accepted.
- Both web applications require a supported Node.js runtime and emit separate
  Next.js server artifacts instead of static-only frontend output.
- Local ports remain 5173 for Storefront and 5174 for Operations so the
  existing OIDC origin and callback contract remains stable.
- Provider selection, cloud provisioning, frontend containers, SSR/SEO
  optimization, and production rollout remain deferred.

---

## ADR-048: Use request-local Party references for guest Common Purchase

**Status:** Accepted

### Decision

The guest Common Purchase request uses opaque, request-local `party_ref`
labels to make role reuse explicit. A Party declaration includes a unique
label and contact-backed display name. Purchaser is always a declaration;
payer is either a distinct declaration or a reference to purchaser. An
intended participant is either a reference to one declared purchaser/payer
Party or a name-only unresolved participant.

`party_ref` is not a Party UUID, customer identifier, or durable handle. It
is valid only within one request. Each declaration label must be unique and
every reference must resolve to exactly one declaration in that request. A
duplicate declaration, unknown reference, or object mixing a reference with
declaration/name fields is invalid. Equal names, emails, and phones never
create a reference or merge identities.

The server creates Party rows only from purchaser/payer declarations. A
resolved participant writes the referenced Party UUID and its current display
name snapshot; a name-only participant writes a null `party_id` with the
provided display-name snapshot. The labels are part of the request payload and
therefore part of the canonical idempotency hash.

### Consequences

- The Storefront OpenAPI contract uses mutually exclusive declaration,
  reference, and name-only participant shapes with examples for shared and
  distinct roles.
- Existing Party and Purchase foreign keys represent the relationships; no
  relationship table, Party search endpoint, contact matching, or migration is
  needed.
- A later authenticated Party-selection or merge workflow requires its own
  contract and decision; it must not overload these guest request labels.

---

## ADR-049: Price Common Purchases per intended participant

**Status:** Accepted

### Context

A Common Purchase stores one Offering, a captured
`offering_unit_price_minor`, an intended `participant_count`, and
`total_amount_minor`. Phase 1 has no cart, purchase-item, or independent
quantity aggregate. Existing two-participant Purchase fixtures already record
twice the Offering unit price, while quota uses the same participant-unit
measure.

### Decision

For a Phase 1 Common Purchase, the total is the exact minor-unit product:

```text
total_amount_minor = offering_unit_price_minor × participant_count
```

The unit price, currency, capacity, count, and total are captured once at
checkout. The multiplication uses checked integer arithmetic and rejects any
result above the API/database exact-integer bound. A zero unit price remains
valid and yields a zero total. Taxes, discounts, fees, donations, conversion,
and a separate purchase quantity remain out of scope.

### Consequences

- `participant_count` is the commercial count as well as the quota unit for
  this one-Offering Common Purchase.
- The Purchase constructor validates the formula, so direct persistence and
  future public checkout cannot bypass it.
- No migration is required: existing immutable snapshot columns already hold
  the inputs and result.
- Any future Offering whose price is for a whole multi-participant package
  needs an explicit pricing-mode requirement and additive contract/schema
  design; it must not silently reinterpret this formula.

---

## ADR-050: Make guest Common Purchase creation durable and replayable

**Status:** Accepted

### Decision

The Purchase UUID remains the canonical API identifier. A Common Purchase also
gets a non-secret human/support reference in this exact form:

```text
QRB-<event-year>-<16 uppercase unpadded Base32 characters>
```

The suffix encodes 10 bytes from `crypto/rand` (80 bits). A database duplicate
reference rolls back the checkout and retries with a new candidate at most five
times; no sequential identifier or generated volume is exposed.

Guest checkout has no authenticated person or durable browser identity, so its
ADR-041 namespace is the stable literal
`storefront.purchase.create.guest`. It identifies the Storefront surface,
Purchase creation command, and the one global guest scope. Request IDs, IP
addresses, headers, Purchase tokens, and Party contacts are never used as
caller scope. The client-supplied high-entropy idempotency key represents one
guest intent and stays only in `idempotency_records`.

For Common Purchase checkout, `Command.Retention == 0` means durable replay:
`idempotency_records.expires_at` is `NULL` and the record is not removed by
normal expiry processing. This is required because issuing another Purchase for
the same retry would remain unacceptable after 24 hours. A configured
idempotency response key must remain available for every durable replay it
encrypted; key cleanup/rotation and any controlled record-retention process
need a separate approved operational policy.

### Consequences

- The encrypted response replay contains the one-time Purchase token but no
  plaintext token, contact, or raw idempotency key is copied to domain, audit,
  or outbox data.
- A successful guest checkout creates Parties, Purchase/history, reservation,
  outbox events, and the durable encrypted replay in one transaction. Failed
  attempts, including a reference collision, leave no completed replay record.
- Other commands retain their explicit positive replay windows. A negative
  retention is invalid.
- Purchase tracking, cancellation, evidence, payment, token recovery, and
  background cleanup remain separate lifecycle work.

---

## ADR-051: Model one Eid Event as three or four local execution days

**Status:** Accepted

### Decision

An executable Qurban Event declares one IANA timezone and exactly three or four
inclusive local execution dates. An execution day owns operating windows and
may contain sessions, shifts, station assignments, handovers, readiness gates,
and recovery periods. UTC timestamps remain the stored instants; local dates
and timezone are preserved as the operational calendar.

Registration windows and Event lifecycle status remain separate from the
execution calendar. Closing a shift or execution day does not silently close
the Event or discard unfinished work.

### Consequences

- Event-day records and APIs carry Event and execution-day scope.
- Validation rejects fewer than three, more than four, duplicate, unordered,
  or timezone-invalid execution dates.
- A later change to duration requires an explicit product decision and
  additive schema/contract change.
- Release evidence includes a continuous 72–96-hour soak and a complete
  three-or-four-day rehearsal.

---

## ADR-052: Treat field teams, shifts, assignments, and incidents as event-scoped operations

**Status:** Accepted

### Decision

Livestock, Allocation, Slaughter, Distribution, Management, and Support teams
are first-class Event-scoped records. Membership, shift, station/location
assignment, handover, readiness, and incident ownership are explicit. A
frontend team selection never grants authority; backend permissions and the
active Event/team/shift assignment authorize each field command.

Operational incidents record severity, affected work, owner, escalation,
resolution, and handover state. Support diagnostics expose safe correlation,
connectivity, queue, projection-lag, and replay state without credentials or
unnecessary participant data.

### Consequences

- Team, membership, shift, assignment, handover, and incident storage uses an
  additive migration; historical migrations are not edited.
- Privileged membership, assignment, handover, incident, and support actions
  are audited and version/conflict protected.
- Volunteer payroll, generic workforce management, and organization-wide HR
  remain out of scope.

---

## ADR-053: Make Sohibul Qurban attendance configurable and independent of eligibility

**Status:** Accepted

### Decision

Purchase eligibility and Sohibul Qurban activation never imply attendance or
personal slaughter. Each Event may enable participant attendance, and each
Sohibul Qurban uses one explicit mode when applicable:

```text
SELF
PROXY
NONE
```

`SELF` and `PROXY` may require check-in and queue/station coordination. `NONE`
does not create a participant queue obligation. Changing an attendance mode
requires authorization, reason, history, and conflict protection.

### Consequences

- Payment verification can activate Sohibul Qurban without an attendance
  decision.
- Storefront exposes only the token-scoped participant's attendance and safe
  progress state.
- Proxy identity, where collected, follows minimum-data and retention rules.

---

## ADR-054: Support explicit Sohibul entitlement and beneficiary distribution

**Status:** Accepted

### Decision

Distribution is a Qurban operational domain, not ecommerce fulfillment. It
supports both Sohibul Qurban entitlement and beneficiary portions. Each record
identifies its subject and portion/entitlement, uses `PICKUP` or `DELIVERY`, and
tracks preparation, readiness, collection/delivery, proof, exception, and
completion.

Distribution begins only after the relevant slaughter and preparation guards
pass. Beneficiary data is Operations-only unless an explicit Purchase-token
scope authorizes the corresponding customer view.

### Consequences

- Beneficiary, portion, proof, and method storage is additive to the current
  minimal distribution schema.
- Pickup and delivery use the same authoritative lifecycle but may have
  different required evidence.
- Route optimization, generalized courier management, and public beneficiary
  lookup remain out of scope.

---

## ADR-055: Use online-authoritative mobile PWAs with polling, SSE, and bounded field replay

**Status:** Accepted

### Decision

The existing responsive Next.js Storefront and Operations PWAs are the MVP
mobile clients. The Go API and PostgreSQL remain authoritative.

Operational reads use bounded polling first. High-value one-way event updates
may use SSE with durable projection cursors, `Last-Event-ID`, reconnection, and
missed-event recovery. WebSocket is not part of the MVP.

A device-local queue may store only explicitly allowlisted, non-financial
field milestones with bounded retention, minimum non-sensitive payload,
idempotency keys, visible pending state, operator-confirmed replay, and conflict
presentation. Payment, evidence, identity, authorization, Event/team
configuration, capacity overrides, and other sensitive commands remain
online-only.

Native QR/barcode detection is an optional browser enhancement. Manual code
entry is always available.

### Consequences

- Dashboard projections are rebuildable and never command truth.
- Service workers continue to exclude API/auth responses from ordinary runtime
  caching; the approved field queue is a separate, narrowly owned mechanism.
- Multi-device and reconnect tests must prove duplicate-free replay and visible
  conflict behavior.
- A native mobile application, Redis, message broker, microservice, or
  WebSocket requires measured need and a separate accepted decision.

---

## ADR-056: Make manual-transfer evidence submission append-oriented and replayable

**Status:** Accepted

### Decision

Phase 1 Common Purchase evidence submissions use method `MANUAL_TRANSFER` and
create a new `SUBMITTED` Payment rather than mutating a rejected attempt.
Payment references use `PAY-` plus 128 random bits encoded as uppercase,
unpadded Base32. They are non-secret support identifiers, not authentication or
idempotency credentials.

Only one `SUBMITTED` Payment may exist for a Purchase. A partial unique index is
the final different-key concurrency guard; rejected records remain immutable
and a later submission uses a new command key and Payment row.

The submission command authenticates the Purchase Bearer token before replay
lookup. Its namespace is scoped to the Storefront command and Purchase UUID,
and its semantic hash includes the Purchase, exact amount/currency, normalized
filename, trusted media type, byte size, and SHA-256 digest rather than raw
multipart boundaries. Successful replay is durable because duplicating the
same evidence intent remains unacceptable after an arbitrary client timeout.

The current Payment schema requires a positive amount. A zero-total Purchase
therefore cannot submit evidence and returns a state conflict. Whether such a
Purchase becomes eligible without Payment is a separate product decision.

### Consequences

- Evidence submission pauses an unexpired reservation or atomically expires
  and reacquires a new numbered reservation under Event-then-Offering locks.
- A reservation already paused for review rejects another submission before
  storage work.
- PostgreSQL stores only immutable evidence metadata and an opaque reference;
  W4-02 separately decides and wires the durable private storage adapter.
- The W4-01 handler remains unwired in production until that adapter exists.
- No payment gateway, partial payment, refund, zero-price eligibility, token
  recovery, or evidence-retention policy is introduced by this decision.

---

## ADR-057: Use a single-instance private filesystem EvidenceStore for the MVP

**Status:** Accepted

### Context

ADR-056 deliberately leaves its evidence-storage port unwired. Evidence is
private, immutable payment support material; PostgreSQL stores its opaque
reference and verified metadata, never bytes. The MVP needs a durable adapter
without adding an SDK, provider, public URL, ACL, listing, or second service.

### Decision

Use one Go-standard-library private filesystem `EvidenceStore` behind the
narrow W4-01 port. An empty `EVIDENCE_STORAGE_ROOT` disables only evidence
upload and download. When configured, startup requires
`EVIDENCE_STORAGE_ROOT`, `EVIDENCE_STORAGE_MODE=single-instance`, and
`API_REPLICA_COUNT=1`; replica count is explicit outside development/test.
Invalid values or an unsafe/unavailable root fail API startup.

The root is an existing absolute, non-root, non-symlink, owner-private
directory outside the Git checkout and served web roots. The adapter holds a
nonblocking lifetime root lease and is run by exactly one API process on a
persistent private volume in staging/production. The root lease and declared
replica count fail closed for declared or same-volume concurrency, but cannot
prove a dishonest separate-volume topology; deployment owns enforcement,
backup, and restore. More than one API instance requires a separately approved
object-provider adapter behind this same port.

References are flat, opaque, 256-bit random values. The adapter confines work
with `os.Root`, creates `0700` subdirectories and `0600` files, streams bounded
content while verifying SHA-256, size, and trusted MIME, syncs and closes a
staging file, publishes by atomic create-only hard link, then syncs the
objects directory before success. Startup performs a capability probe.

Purchase authorization occurs before multipart spooling or replay lookup, and
only the winning idempotency callback calls `Put`. A definite rollback or
panic after a successful `Put` starts a bounded detached best-effort delete.
`database.ErrCommitUncertain` preserves the object for reconciliation and wraps
the original commit cause. A collision never deletes the existing object.

Startup and hourly sweeps serialize with the single process. They use a
24-hour grace only for unreferenced objects and staging files, keep submission
lifetime below that grace, query PostgreSQL for the exact reference immediately
before final deletion, and abort a pass on query error. Referenced evidence is
never age-deleted: product/compliance still owns its retention, hold, and
deletion policy.

`GET /api/operations/v1/payments/{payment_id}/evidence` requires an Operations
session and `payment.verify`. It returns an attachment with `no-store`,
`nosniff`, trusted media type, and length; it never exposes or logs a provider
reference, digest, path, bytes, or credentials. Storefront submission CORS is
exact-origin and non-credentialed, and includes `Authorization` only when that
route is registered.

No schema migration is required now: existing Payment metadata represents an
opaque reference made unique by create-only publication. Any multi-replica or
provider migration remains behind this port and requires a new deployment
decision.

### Consequences

- The Go modular monolith remains authoritative for authorization, storage
  ordering, retrieval, and cleanup.
- Staging/production must provision and protect the one persistent private
  volume; this decision does not provide its topology, backups, or restore.
- No evidence retention duration, public storage URL, PostgreSQL evidence
  bytes, storage dependency, or provider is introduced.
