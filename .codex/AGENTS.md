# Repository Agent Instructions

## Purpose

This repository implements a **Qurban Commerce and Operations Platform** for one operating organization managing recurring annual qurban events.

The product consists of:

- a public Storefront Web;
- an internal Operations Web;
- a Go modular-monolith backend;
- shared API contracts and frontend packages;
- PostgreSQL infrastructure;
- product and architecture documentation.

The platform supports three purchasing channels:

```text
COMMON
SAVING
GIVEAWAY
```

All eligible purchasing channels converge into one canonical Purchase lifecycle and may activate one or more Sohibul Qurban records.

The following roles must remain distinct even when represented by the same person:

- purchaser;
- payer;
- saving-account holder;
- sponsor;
- giveaway applicant;
- giveaway recipient;
- Sohibul Qurban.

---

## Source of Truth

Before planning or implementation, read the applicable canonical documents:

```text
docs/PRD.md
docs/PRODUCT_MAP.md
docs/ARCHITECTURE.md
docs/DECISIONS.md
docs/CONVENTIONS.md
.codex/CURRENT_STATE.md
.codex/TASK.md
```

`MVP-DELIVERY-ROADMAP.md` is the approved delivery-plan baseline. It does not
override canonical product or architecture decisions and is not implementation
evidence.

Document responsibilities:

| Document | Responsibility |
|---|---|
| `docs/PRD.md` | Product goals, actors, requirements, rules, and success criteria |
| `docs/PRODUCT_MAP.md` | Capability hierarchy, roadmap, application ownership, and open requirements |
| `docs/ARCHITECTURE.md` | Technical boundaries, runtime, data, API, and deployment architecture |
| `docs/DECISIONS.md` | Accepted and superseded architecture decisions |
| `docs/CONVENTIONS.md` | Engineering implementation rules, defaults, and enforcement |
| `.codex/CURRENT_STATE.md` | Current implementation state and known gaps |
| `.codex/TASK.md` | Current implementation objective and task-specific constraints |

`.codex/TASK.md` is the active task file. Completed tasks are historical
records under `.codex/archive/` and must not remain as the active task.

`docs/CONVENTIONS.md` applies the established product, architecture, and
decision boundaries. Keep it as the canonical engineering rulebook instead of
duplicating broad implementation rules here.

`.codex/plans/` may contain inactive future-task drafts. A draft is neither
active nor approved for execution merely because it exists. Before activation,
revalidate its dependencies, decisions, file paths, and acceptance criteria
against current canonical documents and source, then copy exactly one reviewed
draft into `.codex/TASK.md`. Execute only the active task and archive only an
executed `.codex/TASK.md`; do not archive unexecuted drafts.

When documents conflict, use this authority model:

```text
PRD and Product Map
→ Architecture
→ Accepted decisions
→ Conventions for implementation defaults
→ Existing implementation and state records
```

The current TASK defines the approved work scope but cannot silently override
an accepted product, architecture, or decision document. Do not silently
resolve material conflicts; record them as plan risks or decision gaps.

---

## Task Completion and Archival

When an implementation task has been completed and its verification confirms
that the objective was implemented, update the active `.codex/TASK.md` before
finishing the task:

1. Add this exact marker immediately below the task's H1 title:

   ```text
   ## Executed
   ```

2. Preserve the complete executed task content, including its objective,
   constraints, plan requirements, verification requirements, and final status.
3. Create the historical directory when it does not exist:

   ```bash
   mkdir -p .codex/archive
   ```

4. Copy the executed task into the archive using the canonical filename:

   ```text
   YYYY-MM-DD-TASK-<h1>.md
   ```

   `<h1>` is the task title text after `# Task:`, normalized into a stable
   lowercase hyphen-separated filename component. For example:

   ```text
   # Task: Frontend API Layers for Operations and Storefront Web
   → .codex/archive/2026-08-03-TASK-frontend-api-layers-for-operations-and-storefront-web.md
   ```

   Use the execution date in the local repository timezone. The archived file
   must retain the `## Executed` marker.

5. Use `cp` to create the archive copy, verify that the copy exists and matches
   the active task, then remove the active task file:

   ```bash
   cp .codex/TASK.md .codex/archive/YYYY-MM-DD-TASK-<h1>.md
   test -s .codex/archive/YYYY-MM-DD-TASK-<h1>.md
   cmp .codex/TASK.md .codex/archive/YYYY-MM-DD-TASK-<h1>.md
   rm .codex/TASK.md
   ```

6. Confirm that `.codex/TASK.md` is absent and that the archived filename is
   unique. Do not overwrite an existing archive; choose a corrected title or
   stop and report the collision.

Do not archive a task merely because code was changed. Archive only after the
implementation has been verified against the task's acceptance criteria and
the final review distinguishes implemented, verified, assumed, and deferred
behavior. If verification is incomplete or the task is abandoned, leave it as
the active `.codex/TASK.md` and record the incomplete status instead.

When starting a new implementation task, create a new `.codex/TASK.md` from
the approved objective. Do not edit an archived task back into an active task.

---

## Required Workflow

Every implementation task must follow:

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

---

## DISCOVER

Before planning:

1. Read:
   - this `AGENTS.md`;
   - `.codex/AGENTS.md` when present;
   - `.codex/TASK.md`;
   - `.codex/CURRENT_STATE.md`;
   - `docs/CONVENTIONS.md` when present;
   - the relevant sections of canonical documents under `docs/`.

2. Inspect only files relevant to the task.

3. Prefer targeted inspection:
   - identify the owning application or backend module;
   - inspect one existing implementation pattern to mirror;
   - inspect contracts, tests, migrations, and callers affected by the change.

4. Do not scan unrelated directories.

5. Do not modify files during discovery.

6. Confirm whether the task affects:
   - Storefront;
   - Operations;
   - public API;
   - operations API;
   - domain rules;
   - persistence;
   - OpenAPI contracts;
   - generated clients;
   - audit;
   - background processing;
   - product or architecture documents.

7. Identify unresolved product decisions before implementation.

---

## PLAN

Before implementation, produce:

1. objective summary;
2. relevant product and architecture constraints;
3. owning capability and module;
4. current behavior;
5. proposed behavior;
6. proposed directory tree when structure changes;
7. planned file-change table;
8. dependencies to add;
9. database and migration impact;
10. API contract impact;
11. audit, authorization, concurrency, and idempotency impact;
12. verification commands;
13. risks and unresolved decisions.

The planned file-change table must contain:

| File | Action | Purpose |
|---|---|---|
| path | create/modify/delete | reason |

Do not begin implementation until the plan is reviewed and approved, unless the task explicitly authorizes direct execution.

---

## BUILD

During implementation:

1. Maintain a concise todo checklist.
2. Modify only planned files.
3. Record plan variance before touching an unplanned file.
4. Keep changes limited to the current task.
5. Do not perform unrelated refactoring.
6. Do not add speculative abstractions.
7. Do not commit unless explicitly requested.
8. Preserve existing public contracts unless the task explicitly changes them.
9. Update OpenAPI contracts when transport behavior changes.
10. Update canonical documents when a product or architecture decision changes.
11. Add or update tests with the implementation.
12. Preserve auditability for privileged and financial changes.
13. Preserve idempotency for retry-sensitive commands.
14. Preserve transactional integrity for quota, payment, conversion, and allocation.
15. Keep generated files separate from handwritten files.

---

## VERIFY

Run every verification command defined by the repository and the current task.

At minimum, when applicable:

```text
format
lint
typecheck
test
build
docker compose config
```

Do not report a command as passing unless it was executed successfully.

If a command cannot run:

- state the exact command;
- state why it could not run;
- do not mark it as passed;
- provide the remaining verification risk.

---

## REVIEW

Finish with:

1. change summary;
2. capability and module summary;
3. file summary;
4. dependency summary;
5. API and schema summary;
6. plan variance;
7. verification results;
8. remaining risks;
9. unresolved decisions;
10. recommended next task.

The review must distinguish:

- implemented behavior;
- verified behavior;
- assumed behavior;
- deferred behavior.

---

## Repository Scope

The canonical monorepo structure is:

```text
apps/
├── operations-web/
├── storefront-web/
└── api/

packages/
├── ui/
├── api-client/
├── contracts/
└── typescript-config/

contracts/
└── openapi/
    ├── operations.yaml
    └── storefront.yaml

infrastructure/
docs/
.github/
.codex/
```

Canonical documentation:

```text
docs/
├── PRD.md
├── PRODUCT_MAP.md
├── ARCHITECTURE.md
├── DECISIONS.md
├── adr/
└── domain/
```

Agent operating context:

```text
.codex/
├── AGENTS.md
├── CURRENT_STATE.md
├── TASK.md                 # active task only
├── plans/                  # inactive future-task drafts
└── archive/                # executed task records
    └── YYYY-MM-DD-TASK-<h1>.md
```

Do not create canonical product or architecture documents under `.codex/`.

---

## Product Capability Boundaries

The canonical product capabilities are:

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

A capability map is not automatically a folder structure. Create a backend module only when it owns meaningful behavior and invariants.

### Storefront

Supports:

- event landing;
- Qurban Offering Catalogue;
- offering search and details;
- common purchasing;
- saving purchasing;
- giveaway participation;
- payment interaction;
- purchase tracking;
- participant profile and documents.

### Purchasing

Owns:

- canonical Purchase lifecycle;
- channel;
- event;
- offering;
- purchaser and payer references;
- eligibility;
- participant activation coordination;
- cancellation;
- history.

### Party & Participant

Owns:

- person and organization identity;
- contextual roles;
- Sohibul Qurban;
- participant verification;
- duplicate detection.

Do not collapse all roles into a generic `Customer`.

### Payment & Funding

Owns:

- payment methods and instructions;
- payment confirmation and verification;
- installment ledger;
- saving balance;
- sponsor funding;
- refund;
- reconciliation.

A payment gateway is an adapter, not the domain.

### Livestock

Owns:

- physical livestock identity;
- classification;
- inspection;
- weight;
- readiness;
- location;
- operational lifecycle;
- history.

Do not model livestock as generic inventory only.

### Allocation

Owns:

- purchase or participant allocation;
- livestock capacity;
- provisional and confirmed assignment;
- release;
- reassignment;
- allocation manifests.

Allocation is a first-class domain and must not be hidden in generic order fulfillment.

### Event Operations

Owns:

- annual qurban event;
- event readiness;
- participant and livestock check-in;
- slaughter schedule;
- queue;
- station;
- live status;
- incidents;
- completion.

For the Full Event-Day MVP, one Event has exactly three or four inclusive local
execution days in an explicit IANA timezone. Field teams, memberships, shifts,
station/location assignments, readiness, handovers, incidents, and support
escalation are Event-scoped records. Attendance supports `SELF`, `PROXY`, or
`NONE` and never determines financial eligibility.

Dashboards are projections over authoritative event and operational records.

### Distribution

Owns:

- planning;
- portion preparation;
- beneficiary assignment;
- pickup;
- delivery;
- proof;
- completion.

The Full Event-Day MVP covers both explicit Sohibul Qurban entitlement and
beneficiary portions, with pickup or delivery, proof, exceptions, and
completion. Route optimization remains out of scope.

### Identity & Access

Owns:

- public and operator authentication;
- user profile;
- role and permission management;
- event or location access scope;
- audit access.

### Administration & Reporting

Provides internal workflows over authoritative domains.

It must not become one unrestricted `admin` domain containing unrelated business logic.

---

## Technology Constraints

### Operations Web

Use:

- Vite;
- React;
- TypeScript;
- React Router with Remix-style routing conventions;
- TanStack Query for remote API/server state when a concrete vertical slice
  requires it;
- Tailwind CSS;
- Vitest;
- PWA-ready structure only when required.

The Operations application may contain:

- event dashboard;
- purchasing operations;
- payment verification;
- saving operations;
- giveaway operations;
- participant operations;
- livestock operations;
- allocation operations;
- slaughter operations;
- distribution operations;
- administration and reporting.

### Storefront Web

Use:

- Vite;
- React;
- TypeScript;
- React Router with Remix-style routing conventions;
- TanStack Query for remote API/server state when a concrete vertical slice
  requires it;
- Tailwind CSS;
- Vitest.

The Storefront may contain:

- event landing;
- Offering Catalogue;
- common purchase journey;
- saving journey;
- giveaway journey;
- payment interaction;
- purchase tracking;
- participant profile and documents.

Do not introduce a Shopping Cart unless multi-offering checkout is confirmed.

### Backend

Use:

- Go;
- modular monolith architecture;
- PostgreSQL;
- Gin as the canonical HTTP framework and router at the HTTP adapter/bootstrap
  boundary, with net/http retained for server lifecycle and transport;
- explicit domain, application, adapter, and transport boundaries.

Do not introduce microservices without an accepted ADR.

### API Contracts

Use OpenAPI as the frontend/backend contract.

Maintain separate contracts:

```text
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
```

Do not manually duplicate backend DTOs as frontend TypeScript interfaces when generated contracts are available.

Use UUID-based public identifiers unless an accepted decision says otherwise.

---

## Frontend Architecture Rules

Keep routes centrally managed:

```text
src/routes/routes.tsx
src/routes/paths.ts
src/routes/guards.tsx
```

Do not scatter route strings throughout components.

Recommended structure:

```text
src/
├── app/
├── routes/
├── features/
├── entities/
├── shared/
│   ├── api/
│   ├── components/
│   ├── hooks/
│   ├── validation/
│   └── utilities/
└── main.tsx
```

Rules:

1. Organize business UI by feature.
2. Use React Router for Remix-style route hierarchy, layouts, route boundaries,
   navigation state, and route-data requirements without introducing a Remix
   server runtime, SSR, or server actions.
3. Keep server state separate from local UI state. TanStack Query owns remote
   request lifecycle, caching, and invalidation after successful API commands;
   it is not yet a dependency or runtime integration.
4. Keep `src/routes/paths.ts` and `src/routes/routes.tsx` as centralized,
   application-owned route registries. Feature routes register explicitly.
5. Do not place business rules inside route definitions or TanStack Query
   callbacks.
6. Do not make frontend validation or cached query data authoritative.
7. Handle loading, empty, error, stale, and `409 Conflict` states explicitly.
8. Do not expose operations-only fields through Storefront clients.
9. Shared UI packages contain stable primitives, not application pages.
10. Prefer direct checkout until cart requirements are confirmed.
11. Revalidate contested state with the Go API; cache data does not determine
    payment status, quota, allocation capacity, saving balance, or queue
    position.

---

## Backend Architecture Rules

Use module-oriented boundaries:

```text
apps/api/internal/modules/<module>/
├── domain/
├── application/
├── infrastructure/persistence/
├── transport/http/
└── module.go
```

A smaller module may use fewer folders. Boundary direction matters more than folder ceremony.

Dependency direction:

```text
HTTP and integration adapters
              ↓
      application services
              ↓
            domain
```

Infrastructure adapters implement ports defined by the domain or application layer.

Domain packages must not depend on:

- HTTP frameworks;
- PostgreSQL drivers;
- provider SDKs;
- infrastructure packages;
- frontend contracts;
- generated OpenAPI transport types.

### Initial Backend Modules

```text
apps/api/internal/
├── modules/
│   ├── event/
│   ├── identity/
│   ├── offering/
│   └── purchasing/
└── platform/
```

Do not create all modules as empty shells. Add a module under
`internal/modules` only when required by an implemented vertical slice.

---

## Domain Rules

### Qurban Event

Every event-scoped aggregate must explicitly reference a `QurbanEvent`.

Historical events must not inherit future changes to:

- pricing;
- quota;
- offering rules;
- operational dates;
- allocation rules;
- event configuration.

### Purchasing Channels

Supported channels:

```text
COMMON
SAVING
GIVEAWAY
```

Every canonical Purchase belongs to exactly one channel.

### Canonical Purchase

All successful channels converge into one Purchase lifecycle.

Do not implement three unrelated order systems.

Channel-specific sources remain explicit:

```text
common checkout
saving conversion
giveaway assignment
```

### Saving

A Saving Account is not a completed Purchase.

Saving becomes a Purchase only after a successful, idempotent, transactionally safe conversion.

### Giveaway

A giveaway application or nomination is not a completed Purchase.

A Purchase is created or activated only after approved recipient assignment.

### Sohibul Qurban

Sohibul Qurban registration is an outcome of eligible purchasing, not a purchasing channel.

Do not assume Sohibul Qurban is the purchaser or payer.

### Offering

Use `Offering` as the commercial abstraction.

An Offering may eventually represent:

- an individual livestock unit;
- a category;
- a package;
- a cattle share;
- a saving target;
- another event-specific qurban product.

Do not bind the catalogue directly to physical livestock unless the confirmed requirement requires it.

### Livestock

Livestock is a physical operational entity with lifecycle state.

Indicative lifecycle:

```text
REGISTERED
→ INSPECTED
→ READY
→ ALLOCATED
→ QUEUED
→ SLAUGHTERED
```

Final states must be defined by the owning task and domain model.

### Allocation

Allocation capacity must not be exceeded.

Use transactional safeguards such as:

- row-level locking;
- atomic conditional updates;
- optimistic concurrency;
- unique constraints;
- idempotency records.

Reallocation and override actions require authorization, reason, and audit history.

### Financial Records

Use ledger-based records for:

- payments;
- installments;
- refunds;
- transfers;
- adjustments;
- sponsor funding.

Do not use floating point for money.

Do not mutate financial history without traceable adjustment records.

---

## API Rules

Use route groups:

```text
/api/public/v1/*
/api/operations/v1/*
```

Rules:

1. Keep public and operations DTOs separate.
2. Use structured error responses.
3. Use pagination for potentially large lists.
4. Use deterministic filtering and sorting.
5. Use request correlation identifiers.
6. Use idempotency keys for retry-sensitive commands.
7. Do not expose database column names as accidental contracts.
8. Do not expose privileged override operations publicly.
9. Update OpenAPI whenever API behavior changes.
10. Validate authorization in backend policies, not only middleware or frontend guards.

---

## Database Rules

Use PostgreSQL as the transactional system of record.

Rules:

1. Use database constraints for critical invariants where practical.
2. Use exact numeric representation for money.
3. Use UTC timestamps with explicit timezone handling.
4. Use UUIDs for public references.
5. Keep financial and audit records append-oriented.
6. Version every schema change through migrations.
7. Preserve historical event configuration through snapshots or versioned references.
8. Do not implement hypothetical SaaS multitenancy.
9. Do not add `organisation_id` to every table without a confirmed multi-organization requirement.
10. Analytics and dashboard projections are not transactional truth.

---

## Asynchronous Processing Rules

Use an outbox for asynchronous side effects.

Potential worker responsibilities:

- notification delivery;
- projection updates;
- provider callbacks;
- report generation;
- retryable integration work.

Rules:

1. Persist outbox events in the same transaction as authoritative state changes.
2. Make consumers idempotent.
3. Do not let notification or reporting failures roll back valid transactions.
4. Do not add a message broker without a demonstrated requirement.
5. Prefer a PostgreSQL-backed outbox and worker initially.

---

## Real-Time Operations Rules

The dashboard requires near-real-time visibility, not hard real-time guarantees.

Preferred progression:

```text
polling
→ Server-Sent Events
→ WebSocket only when bidirectional communication is required
```

Rules:

1. Dashboard projections must be rebuildable.
2. Projection lag must be observable.
3. Dashboard state must not become the authoritative source.
4. Commands must operate against transactional modules.
5. Do not add WebSocket infrastructure merely for visual freshness.
6. Full Event-Day SSE must reconnect from `Last-Event-ID` or an equivalent
   durable cursor and retain polling as fallback.
7. Commands continue through authenticated HTTP APIs and revalidate
   authoritative state; they never mutate through a projection or stream.

### Mobile and degraded-connectivity rules

1. The responsive Next.js PWAs are the MVP mobile clients.
2. Every QR/barcode-assisted critical flow has manual code entry fallback.
3. Ordinary service-worker caching excludes API/authentication/business data.
4. Only ADR-055 allowlisted non-financial field milestones may queue locally.
5. Queued milestones use minimum non-sensitive payloads, bounded retention,
   idempotency, visible pending state, operator-confirmed replay, and conflict
   presentation.
6. Payment, evidence, identity, permission, Event/team configuration, capacity
   override, credentials, and other sensitive commands remain online-only.

---

## Audit and Security Rules

Sensitive actions must record:

- actor;
- timestamp;
- permission context;
- request identifier;
- action;
- target;
- reason where required;
- relevant before and after values.

Examples:

- payment verification override;
- installment adjustment;
- giveaway approval;
- participant replacement;
- livestock reassignment;
- queue priority override;
- distribution reversal.

Security rules:

1. Use permission-based authorization.
2. Enforce authorization in backend application policies.
3. Apply least privilege.
4. Do not log secrets or unnecessary personal data.
5. Rate-limit public endpoints where appropriate.
6. Keep audit records append-only.
7. Protect giveaway and beneficiary data from public disclosure.
8. Never commit credentials, tokens, secrets, or local environment files.

---

## Code Quality Rules

1. Prefer simple and explicit implementations.
2. Avoid premature abstraction.
3. Avoid generic utility packages without demonstrated reuse.
4. Keep public APIs narrow.
5. Handle errors explicitly.
6. Do not suppress linting errors without justification.
7. Keep generated files separate from handwritten files.
8. Never modify lockfiles manually.
9. Do not leave broken placeholder imports.
10. Avoid empty architecture scaffolding without an active vertical slice.
11. Mirror existing validated patterns where possible.
12. Add comments only when intent is not clear from code.
13. Keep domain terminology consistent with canonical documents.
14. Do not use generic commerce terminology when qurban-specific concepts exist.

---

## Dependency Rules

Before adding a dependency:

1. state why it is needed;
2. confirm the standard library or existing dependency cannot reasonably cover it;
3. add it only to the narrowest applicable workspace package;
4. record it in the implementation summary;
5. verify its license and maintenance status when material.

Do not add future dependencies preemptively.

Dependencies requiring explicit task justification include:

- payment SDKs;
- Redis clients;
- message brokers;
- ORM frameworks;
- charting libraries;
- complex state-management libraries;
- offline database libraries;
- authentication providers;
- email providers;
- WebSocket frameworks;
- workflow engines.

---

## Verification Standards

Frontend applications must support:

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm build
```

The Go API must support:

```bash
go fmt ./...
go vet ./...
go test ./...
go build ./...
```

Infrastructure must support:

```bash
docker compose -f infrastructure/compose.yaml config
```

Repository-level validation should use:

```bash
make validate
```

Verification must be reproducible from a clean clone.

Task-specific verification should also include, when applicable:

- migration up and down validation;
- OpenAPI validation;
- generated client consistency;
- repository integration tests;
- concurrency tests;
- idempotency tests;
- authorization tests;
- audit assertions;
- critical end-to-end flows.

---

## Delivery Rules

Implement through vertical slices.

Recommended first slice:

```text
Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Basic Operations Dashboard
```

A vertical slice should include, where applicable:

- domain model;
- application commands and queries;
- persistence;
- API contract;
- Storefront workflow;
- Operations workflow;
- authorization;
- audit;
- tests;
- observability.

Do not implement every database table, module shell, or frontend page before one complete business flow works.

---

## Work That Requires Explicit Task Scope

Do not implement the following unless the current task explicitly requests it:

- authentication provider integration;
- payment gateway integration;
- bank reconciliation;
- generic SaaS multitenancy;
- multi-brand support;
- microservices;
- Redis;
- message brokers;
- Kubernetes;
- production deployment;
- offline event mutation queues;
- WebSocket infrastructure;
- generalized accounting;
- route optimization;
- speculative mobile applications;
- speculative POS features;
- Shopping Cart without confirmed multi-offering checkout;
- distribution rules not yet confirmed;
- saving price-lock behavior not yet confirmed;
- automated giveaway selection without confirmed policy.

---

## Open Product Decisions

Before implementing affected capabilities, verify:

1. whether offerings represent individual livestock, categories, packages, cattle shares, or a combination;
2. whether one checkout may contain multiple offerings;
3. whether saving plans lock offering and price;
4. how giveaway recipients are selected;
5. whether each Sohibul Qurban personally performs slaughter and requires an individual attendance queue;
6. whether distribution covers Sohibul Qurban entitlement, beneficiaries, pickup, delivery, or a combination.

Do not invent these rules in code.

---

## Git Rules

1. Do not commit unless explicitly requested.
2. Do not push unless explicitly requested.
3. Do not rewrite Git history.
4. Do not modify `.git` internals.
5. Do not include generated build output.
6. Do not include `.env`.
7. Keep each task reviewable as one coherent change set.
8. Do not mix documentation migration with unrelated implementation.
9. Record plan variance before modifying unplanned files.
