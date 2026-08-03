# Current Repository State

## Repository

```text
Repository: persona-apps
GitHub owner: Kangditya
Primary branch: main
```

## Current Condition

The repository bootstrap is implemented.

It currently contains:

- two frontend application shells;
- one Go API shell;
- shared workspace tooling;
- PostgreSQL local infrastructure;
- placeholder API contracts;
- documentation;
- CI and repository validation commands.

No qurban business capability is implemented yet.

The product direction has changed from a generic single-brand commerce and POS platform into a:

> Qurban Commerce and Operations Platform

The canonical product and architecture documents now describe:

- annual Qurban Events;
- Common Purchasing;
- Saving Purchasing;
- Giveaway Purchasing;
- Sohibul Qurban;
- Payment and Funding;
- Livestock;
- Allocation;
- Event Operations;
- Distribution;
- Administration and Reporting.

---

## Canonical Documentation

The canonical documentation set is:

```text
docs/
├── PRD.md
├── PRODUCT_MAP.md
├── ARCHITECTURE.md
└── DECISIONS.md
```

Agent execution context is expected under:

```text
.codex/
├── AGENTS.md
├── CURRENT_STATE.md
└── TASK.md
```

The previous `.codex/DECISIONS.md` location is no longer canonical.

---

## Implemented Foundation

### Repository and Tooling

- Git repository initialization;
- repository-local Git identity;
- GitHub remote configuration;
- pnpm workspace;
- Turborepo command orchestration;
- repository Make targets;
- formatting, linting, testing, build, and validation workflows;
- GitHub-publishable repository baseline.

### Frontend Applications

Implemented application shells:

```text
apps/
├── operations-web/
└── storefront-web/
```

Current frontend stack:

- Vite;
- React;
- TypeScript;
- React Router with documented Remix-style routing conventions;
- TanStack Query as the approved future remote API/server-state library;
- Tailwind CSS;
- Oxlint;
- Vitest.

The current application shells do not yet implement route modules, a TanStack
Query provider, queries, mutations, API calls, or cache behavior. No TanStack
package is installed or declared until a concrete frontend vertical slice needs
remote data.

Local development behavior:

- Storefront Vite server with HMR on `127.0.0.1:5173`;
- Operations Vite server with HMR on `127.0.0.1:5174`;
- Go API live reload through Air using `apps/api/.air.toml`;
- API default address `127.0.0.1:8080`.

Current placeholder routes:

#### Operations Web

```text
/operator-login
/event-dashboard
/purchasing
/payment-verification
```

#### Storefront Web

```text
/
/offerings
/purchase-tracking
```

These routes are capability-aligned placeholders only. They do not implement authentication, catalogue, purchasing, payment verification, or dashboard behavior.

### Backend

Implemented:

```text
apps/api
```

Current backend capabilities:

- Go API process;
- `/health`;
- PostgreSQL-backed `/ready`;
- graceful shutdown;
- Air-compatible local live-reload configuration;
- PostgreSQL connectivity;
- Docker image build.

No qurban domain modules exist yet.

### Infrastructure

Implemented:

- Docker Compose PostgreSQL service;
- configurable PostgreSQL host port;
- local environment bootstrap;
- container configuration validation.

### Contracts and Shared Packages

Implemented as placeholders or shells:

- OpenAPI contract locations;
- shared workspace packages;
- frontend TypeScript configuration;
- API client package shell;
- shared UI package shell.

No meaningful qurban API contract has been implemented.
The separate public and operations OpenAPI contracts remain endpoint-free; no
route data requirement or query key is implemented yet.

### Documentation

Implemented or revised:

- product requirements;
- product capability map;
- architecture baseline;
- architecture decision register;
- repository agent instructions;
- current-state tracking.

The architecture artifacts document React Router with Remix-style routing
conventions and future TanStack Query server-state ownership. This is an
accepted direction, not a runtime implementation.

---

## Product Direction

The accepted product capabilities are:

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

The accepted purchasing channels are:

```text
COMMON
SAVING
GIVEAWAY
```

All eligible channels converge into one canonical Purchase lifecycle.

Sohibul Qurban is not a purchasing channel.

The following roles must remain distinct:

- purchaser;
- payer;
- saving-account holder;
- sponsor;
- giveaway applicant;
- giveaway recipient;
- Sohibul Qurban.

---

## Not Implemented

### Product Foundation

- Qurban Event;
- event lifecycle;
- event configuration;
- event-specific capacity;
- event-specific offering availability;
- event history and archival behavior.

### Offering

- Qurban Offering Catalogue;
- offering types;
- pricing;
- offering availability;
- offering publication;
- participant capacity rules;
- livestock share rules.

### Purchasing

- canonical Purchase aggregate;
- Common Purchasing;
- Saving Purchasing;
- Giveaway Purchasing;
- checkout;
- purchase validation;
- purchase confirmation;
- purchase cancellation;
- purchase history.

### Party and Participant

- person and organization identity;
- purchaser;
- payer;
- saving-account holder;
- sponsor;
- giveaway applicant;
- giveaway recipient;
- Sohibul Qurban;
- participant verification;
- duplicate detection.

### Payment and Funding

- payment methods;
- payment instructions;
- payment confirmation;
- payment verification;
- installment ledger;
- saving balances;
- sponsor funding;
- refund;
- reconciliation;
- payment gateway integration.

### Saving

- saving account;
- saving target;
- installment workflow;
- funding state;
- price-lock policy;
- conversion eligibility;
- saving-to-purchase conversion.

### Giveaway

- giveaway programs;
- sponsor funding;
- applicant or nominee intake;
- eligibility review;
- recipient selection;
- approval;
- purchase assignment.

### Livestock

- livestock registry;
- classification;
- inspection;
- weight;
- readiness;
- pen or location assignment;
- physical lifecycle;
- livestock history.

### Allocation

- purchase allocation;
- Sohibul Qurban allocation;
- shared livestock capacity;
- provisional allocation;
- confirmed allocation;
- release;
- reassignment;
- allocation manifests.

### Event Operations

- event readiness;
- participant check-in;
- livestock check-in;
- slaughter schedule;
- slaughter queue;
- slaughter station;
- live status tracking;
- incidents;
- event completion.

### Distribution

- distribution planning;
- portion preparation;
- Sohibul Qurban entitlement;
- beneficiary assignment;
- pickup;
- delivery;
- collection confirmation;
- delivery proof;
- distribution completion.

### Identity and Access

- storefront authentication;
- operator authentication;
- user profiles;
- roles;
- permissions;
- event-scoped access;
- audit access.

### Administration and Reporting

- event administration;
- offering administration;
- purchasing administration;
- payment verification queues;
- saving administration;
- giveaway administration;
- participant administration;
- livestock administration;
- allocation administration;
- operational dashboards;
- reports;
- exports;
- audit log interface.

### Platform Capabilities

- business database schema;
- database migrations;
- audit framework;
- permission enforcement;
- request idempotency;
- outbox;
- background worker;
- notification delivery;
- operational projections;
- polling or SSE dashboard updates;
- production observability;
- production deployment.

---

## Previous Bootstrap Assumptions Now Superseded

The following previous assumptions are no longer accepted product direction:

- POS as a primary product capability;
- generic back-office commerce;
- generic catalogue and inventory;
- generic customer ordering;
- cart as a mandatory workflow;
- future SaaS as the default architecture;
- shared-schema multitenancy by default;
- mandatory `organisation_id` on future business tables;
- generic fulfillment as the primary operational abstraction.

Current replacements:

| Previous Assumption | Current Direction |
|---|---|
| Product catalogue | Qurban Offering Catalogue |
| Inventory | Livestock lifecycle |
| Customer | Explicit party and participant roles |
| Order | Canonical Purchase |
| POS operations | Qurban Event Operations |
| Fulfillment | Allocation, Slaughter, and Distribution |
| SaaS multitenancy | One operating organization initially |
| Mandatory cart | Direct checkout unless multi-offering checkout is confirmed |

---

## Current Verification State

Last recorded full verification: **2026-08-03**

The product/architecture and frontend-artifact alignment has been verified
with:

```bash
make validate
```

This passed formatting checks, frontend linting, TypeScript type checking,
frontend tests, frontend builds, Go formatting, Go vet, Go tests, Go build,
and Docker Compose configuration validation.

The verification confirms artifact consistency only. It does not verify
Remix-style route modules, TanStack Query integration, API calls, caching, or
any qurban business behavior because none is implemented.

The API foundation remains limited to `/health`, PostgreSQL-backed `/ready`, and graceful shutdown. No qurban business capability has been implemented.

---

## Known Constraints

1. Development is currently handled by one developer.
2. The initial budget is minimal.
3. The repository must remain understandable and maintainable.
4. The application must support up to approximately 1,000 Sohibul Qurban per annual event.
5. The operational record count may reach thousands across purchases, payments, livestock, allocations, slaughter events, and distribution.
6. PostgreSQL may already run locally for another project.
7. The local PostgreSQL host port must remain configurable.
8. No secrets may be committed.
9. The repository may remain publicly accessible during early development.
10. Public repository status requires strict exclusion of credentials, production data, and private participant information.
11. Microservices are prohibited without an accepted ADR and demonstrated need.
12. Generic SaaS multitenancy is not part of the initial product.
13. Dashboard requirements are near-real-time, not hard real-time.
14. Product rules must remain adjustable while Figma and field requirements are still being refined.

---

## Open Product Decisions

The following decisions remain unresolved:

1. Whether offerings represent:
   - individual livestock;
   - categories;
   - packages;
   - cattle shares;
   - or a combination.

2. Whether one checkout may contain multiple offerings.

3. Whether Saving Purchasing locks:
   - the offering;
   - the price;
   - both;
   - or neither.

4. How giveaway recipients are selected:
   - sponsor selection;
   - committee selection;
   - manual approval;
   - random draw;
   - combined process.

5. Whether every Sohibul Qurban personally performs the slaughter.

6. Whether personal slaughter requires:
   - attendance registration;
   - check-in;
   - personal queue number;
   - assigned slaughter station;
   - proxy representation.

7. Whether distribution covers:
   - Sohibul Qurban entitlement;
   - beneficiaries;
   - pickup;
   - delivery;
   - or a combination.

These rules must not be invented during implementation.

---

## Current Risks

1. Remix-style route modules and TanStack Query runtime integration are
   documented but not implemented; their first use must be part of an approved
   vertical slice with real API behavior and tests.
2. Query-key shapes, stale-time policy, mutation invalidation, and route-data
   requirements cannot be finalized until meaningful public and operations
   endpoints exist.
3. The Go backend has no module registration pattern yet.
4. No migration tool has been selected.
5. No HTTP router decision has been finalized.
6. No authentication approach has been selected.
7. No initial database domain model has been validated.

---

## Recommended Next Task

Define and approve the first business vertical slice before adding routing or
server-state runtime code. That slice must identify its route modules, the
applicable separate OpenAPI contract, and the smallest required
`@tanstack/react-query` integration.

The recommended first slice remains:

```text
Qurban Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Basic Operations Dashboard
```

---

## Milestone Status

### Completed

```text
Repository bootstrap
```

### In Progress

```text
Product and architecture realignment
```

### Not Started

```text
First qurban vertical slice
```
