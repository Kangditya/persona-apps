# `.codex/AGENTS.md`

# Repository Agent Instructions

## Purpose

This repository will become a single-brand commerce SaaS combining:

* point-of-sale operations;
* back-office management;
* public storefront;
* shared catalogue;
* shared inventory;
* customer ordering.

The repository is currently empty. The first milestone is limited to creating a clean, publishable monorepo foundation.

---

## Required Workflow

Every implementation task must follow:

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

### DISCOVER

Before planning:

1. Read:

   * `.codex/AGENTS.md`
   * `.codex/ARCHITECTURE.md`
   * `.codex/DECISIONS.md`
   * `.codex/CURRENT_STATE.md`
   * `.codex/TASK.md`
2. Inspect only files relevant to the current task.
3. Do not scan unrelated directories.
4. Do not modify files during discovery.

### PLAN

Before implementation, produce:

1. objective summary;
2. relevant architecture constraints;
3. proposed directory tree;
4. planned file-change table;
5. dependencies to add;
6. verification commands;
7. risks or unresolved decisions.

The planned file-change table must contain:

| File | Action               | Purpose |
| ---- | -------------------- | ------- |
| path | create/modify/delete | reason  |

Do not begin implementation until the plan has been reviewed and approved.

### BUILD

During implementation:

1. Maintain a concise todo checklist.
2. Modify only planned files.
3. Record plan variance before touching an unplanned file.
4. Keep changes limited to the current task.
5. Do not perform unrelated refactoring.
6. Do not add speculative abstractions.
7. Do not commit unless explicitly requested.

### VERIFY

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

Do not report a command as passing unless it was actually executed successfully.

### REVIEW

Finish with:

1. change summary;
2. file summary;
3. dependency summary;
4. plan variance;
5. verification results;
6. remaining risks;
7. recommended next task.

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
├── eslint-config/
└── typescript-config/

contracts/
└── openapi/

infrastructure/
docs/
.github/
.codex/
```

---

## Technology Constraints

### Operations frontend

Use:

* Vite
* React
* TypeScript
* React Router
* Tailwind CSS
* PWA-ready structure

The operations application will eventually contain:

* POS;
* back-office management;
* catalogue;
* inventory;
* orders;
* reports;
* organisation settings.

### Storefront frontend

Use:

* Vite
* React
* TypeScript
* React Router
* Tailwind CSS

The storefront will eventually contain:

* public product catalogue;
* cart;
* customer checkout;
* order status.

### Backend

Use:

* Go
* modular monolith architecture;
* PostgreSQL;
* standard-library-compatible HTTP architecture;
* explicit application, domain, transport, and persistence boundaries.

Do not introduce microservices.

### API contract

Use OpenAPI as the public frontend/backend contract.

Do not manually duplicate backend DTOs as frontend TypeScript interfaces when generated contracts become available.

---

## Architecture Rules

### Frontend

Keep routes centrally managed:

```text
src/routes/routes.tsx
src/routes/paths.ts
```

Do not scatter route strings throughout components.

Separate:

```text
routes
layouts
pages
features
components
lib
pwa
```

Do not place business logic inside route definitions.

### Backend

Use module-oriented boundaries:

```text
internal/modules/<module>/
├── domain/
├── application/
├── infrastructure/
└── transport/http/
```

Dependency direction:

```text
transport
    ↓
application
    ↓
domain

infrastructure
    → implements domain interfaces
```

Domain packages must not depend on:

* HTTP frameworks;
* PostgreSQL drivers;
* infrastructure packages;
* frontend contracts.

### Database

The future SaaS will use shared-schema multitenancy.

Tenant-owned data must eventually be scoped using:

```text
organisation_id
```

However, business schemas are explicitly outside the bootstrap milestone.

---

## Code Quality Rules

1. Prefer simple and explicit implementations.
2. Avoid premature abstraction.
3. Avoid generic utility packages without demonstrated reuse.
4. Keep public APIs narrow.
5. Handle errors explicitly.
6. Do not suppress linting errors without justification.
7. Keep generated files separate from handwritten files.
8. Never commit credentials, tokens, secrets, or local environment files.
9. Never modify lockfiles manually.
10. Do not leave broken placeholder imports.

---

## Dependency Rules

Before adding a dependency:

1. state why it is needed;
2. confirm the standard library or existing dependency cannot reasonably cover it;
3. add it only to the narrowest applicable workspace package;
4. record it in the implementation summary.

Do not add future dependencies preemptively.

Examples of dependencies that must not be added during bootstrap unless explicitly required:

* payment SDKs;
* Redis clients;
* message brokers;
* ORM frameworks;
* charting libraries;
* complex state-management libraries;
* offline database libraries;
* authentication providers;
* email providers.

---

## Verification Standards

Frontend applications must eventually support:

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm build
```

The Go API must eventually support:

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

Repository validation must be reproducible from a clean clone.

---

## Prohibited Bootstrap Work

Do not implement during repository bootstrap:

* authentication;
* user registration;
* tenant resolution;
* business database tables;
* catalogue CRUD;
* inventory;
* cashier shifts;
* POS checkout;
* online ordering;
* payment processing;
* refunds;
* billing;
* offline mutation queues;
* background workers;
* message brokers;
* Kubernetes;
* production deployment;
* production observability.

---

## Git Rules

1. Do not commit unless explicitly requested.
2. Do not push unless explicitly requested.
3. Do not rewrite Git history.
4. Do not modify `.git` internals.
5. Do not include generated build output.
6. Do not include `.env`.
7. Keep each task reviewable as one coherent change set.

---

# `.codex/ARCHITECTURE.md`

# Canonical Architecture

## Product Definition

The future product is a single-brand commerce SaaS combining:

```text
Operations
├── Back-office management
└── Point of sale

Customer commerce
└── Public storefront

Shared platform
├── Catalogue
├── Pricing
├── Inventory
├── Orders
├── Sales
└── Customers
```

Each SaaS customer organisation will initially operate one brand.

The initial product assumption is:

```text
one organisation = one brand
```

Multiple brands per organisation are not part of the initial architecture.

---

## Current Milestone

The current milestone is repository bootstrap only.

The repository must become:

* structurally clear;
* locally buildable;
* suitable for publication on GitHub;
* ready for future vertical-slice implementation;
* intentionally free of business functionality.

---

## Monorepo Topology

```text
persona-apps/
├── apps/
│   ├── operations-web/
│   ├── storefront-web/
│   └── api/
│
├── packages/
│   ├── ui/
│   ├── api-client/
│   ├── eslint-config/
│   └── typescript-config/
│
├── contracts/
│   └── openapi/
│
├── infrastructure/
│   ├── compose.yaml
│   ├── nginx/
│   └── postgres/
│
├── docs/
│   ├── architecture/
│   ├── decisions/
│   └── product/
│
├── .github/
├── .codex/
├── Makefile
├── package.json
├── pnpm-workspace.yaml
└── turbo.json
```

---

## Application Boundaries

### `apps/operations-web`

Responsibilities:

* internal employee-facing application;
* future POS interface;
* future back-office interface;
* organisation and outlet context;
* operational workflows;
* installable PWA capabilities.

Initial routes:

```text
/login
/management
/pos
```

Canonical internal structure:

```text
src/
├── app/
├── routes/
├── layouts/
├── pages/
├── features/
├── components/
├── lib/
├── pwa/
├── styles/
└── main.tsx
```

### `apps/storefront-web`

Responsibilities:

* public customer-facing storefront;
* product discovery;
* customer cart;
* future checkout;
* future order tracking.

Initial routes:

```text
/
/products
/cart
```

Canonical internal structure:

```text
src/
├── app/
├── routes/
├── layouts/
├── pages/
├── features/
├── components/
├── lib/
├── styles/
└── main.tsx
```

### `apps/api`

Responsibilities:

* authoritative business logic;
* PostgreSQL access;
* authentication and authorisation later;
* catalogue;
* inventory;
* sales;
* orders;
* fulfilment;
* reporting.

The API remains one modular monolith.

Canonical structure:

```text
apps/api/
├── cmd/
│   ├── api/
│   └── migrate/
├── internal/
│   ├── app/
│   ├── config/
│   ├── platform/
│   └── modules/
├── migrations/
├── queries/
├── tests/
├── go.mod
└── Dockerfile
```

Initial endpoints:

```text
GET /health
GET /ready
```

`/health` verifies that the process is running.

`/ready` verifies that required dependencies, especially PostgreSQL, are reachable.

---

## Frontend Routing

Vite does not provide routing.

React Router owns runtime navigation.

Each frontend application must manage routes centrally:

```text
src/routes/paths.ts
src/routes/routes.tsx
```

`paths.ts` owns URL patterns and URL builders.

Example:

```ts
export const paths = {
  management: "/management",
  pos: "/pos",
  productDetail: {
    pattern: "/products/:productId",
    build: (productId: string) =>
      `/products/${encodeURIComponent(productId)}`,
  },
} as const;
```

`routes.tsx` owns the route tree, layouts, guards, and lazy page loading.

Literal route strings must not be duplicated across components.

---

## Shared Packages

### `packages/ui`

Contains reusable presentational primitives shared by frontend applications.

It must not contain:

* business workflows;
* API calls;
* POS state;
* storefront-specific logic;
* backend domain types.

### `packages/api-client`

Reserved for generated TypeScript clients derived from OpenAPI.

During bootstrap it may contain documentation or an empty public entry point.

### `packages/eslint-config`

Reserved for shared frontend linting configuration.

The repository may use Oxlint as the primary fast linter. ESLint should be introduced only if a required plugin or rule is unavailable.

### `packages/typescript-config`

Contains reusable TypeScript configuration bases.

Applications may extend these bases while retaining application-specific configuration.

---

## API Contracts

OpenAPI contracts live under:

```text
contracts/openapi/
├── operations.yaml
└── storefront.yaml
```

The contracts may eventually be served by the same Go API process.

They remain separated because:

* operational APIs are private;
* storefront APIs are public;
* authentication models differ;
* generated clients should expose only relevant endpoints.

No meaningful generated API client is required during bootstrap.

---

## Data Architecture

The future system will use PostgreSQL.

Expected future ownership hierarchy:

```text
Organisation
└── Outlet
    ├── Inventory
    ├── Employees
    ├── POS terminals
    └── Sales
```

Future tenant-owned tables must carry:

```text
organisation_id
```

Outlet-owned records may additionally carry:

```text
outlet_id
```

Future money values must use integer minor or smallest practical currency units.

For IDR:

```text
IDR 25,000 → 25000
```

Floating-point values must not be used for monetary storage or calculations.

---

## Future Domain Modules

The intended module boundaries are:

```text
identity
organisation
outlet
catalogue
pricing
inventory
shift
sale
customer
cart
order
payment
fulfilment
reporting
```

These directories should not all be generated with empty production code during bootstrap.

Introduce a module when its first vertical slice is implemented.

---

## Future Transactional Invariants

The following are architectural invariants for later work:

1. Checkout must be idempotent.
2. A completed sale must not be silently mutated.
3. Refunds and voids must use explicit compensating records.
4. Inventory must be represented through stock movements.
5. Historical sale items must retain product and price snapshots.
6. Tenant-owned repository operations must require organisation context.
7. POS and storefront must consume one inventory authority.
8. Payment state and fulfilment state must remain separate.
9. Cross-domain financial operations must run transactionally.
10. Public API identifiers should use opaque UUIDs.

These invariants are documented now but not implemented during bootstrap.

---

## Deployment Boundary

The initial local topology is:

```text
Operations Vite app
Storefront Vite app
        ↓
Go API
        ↓
PostgreSQL
```

Docker Compose will initially provide PostgreSQL.

The Go API and frontend applications may run directly on the host during development.

Production deployment is explicitly outside the bootstrap milestone.

---

## Initial Quality Gates

The repository should eventually expose:

```bash
make install
make dev
make build
make lint
make typecheck
make test
make format
make validate
```

`make validate` should become the canonical pre-push verification command.

---

# `.codex/DECISIONS.md`

# Architecture Decision Register

## ADR-001: Use a monorepo

**Status:** Accepted

Use one repository for:

* operations frontend;
* storefront frontend;
* Go API;
* shared frontend packages;
* API contracts;
* local infrastructure;
* architecture documentation.

This keeps contracts, tooling, and versioning aligned while the project is maintained by one developer.

---

## ADR-002: Use two frontend applications

**Status:** Accepted

Use:

```text
apps/operations-web
apps/storefront-web
```

`operations-web` combines POS and back-office management.

`storefront-web` contains the public customer storefront.

POS and management remain one application because they share:

* employee authentication;
* outlet context;
* catalogue;
* inventory;
* operational permissions.

The storefront remains separate because it has:

* public access;
* different caching requirements;
* different user journeys;
* different deployment risk;
* customer-focused presentation.

---

## ADR-003: Use Vite and React

**Status:** Accepted

Both frontend applications use:

* Vite;
* React;
* TypeScript;
* React Router.

Vite is preferred because:

* the backend is implemented in Go;
* the application is primarily API-driven;
* the POS is client-oriented;
* central route configuration is required;
* the authenticated application does not depend on server-side rendering.

---

## ADR-004: Use central route registries

**Status:** Accepted

Each frontend application owns:

```text
src/routes/routes.tsx
src/routes/paths.ts
```

Route strings, URL builders, guards, and layout hierarchy must be managed centrally.

---

## ADR-005: Use Tailwind CSS

**Status:** Accepted

Use Tailwind CSS v4 through the Vite integration.

Initial setup should remain minimal:

```css
@import "tailwindcss";
```

Do not create a large design-token system during repository bootstrap.

---

## ADR-006: Use Oxlint as the primary frontend linter

**Status:** Accepted

Use Oxlint for fast JavaScript and TypeScript linting.

Use the TypeScript compiler separately for type validation.

Introduce ESLint only if a concrete unsupported plugin or rule is required.

---

## ADR-007: Use pnpm workspaces and Turborepo

**Status:** Accepted

Use pnpm for JavaScript package management.

Use Turborepo for cross-workspace command orchestration.

The Go backend remains a native Go module and must not depend on JavaScript tooling for correctness.

---

## ADR-008: Use a Go modular monolith

**Status:** Accepted

Use one Go API process under:

```text
apps/api
```

Do not introduce microservices during initial development.

Modules should use explicit domain, application, infrastructure, and transport boundaries.

---

## ADR-009: Use PostgreSQL

**Status:** Accepted

Use PostgreSQL as the authoritative database.

During bootstrap, implement only:

* local PostgreSQL infrastructure;
* database connectivity;
* readiness checking.

Do not create business tables during bootstrap.

---

## ADR-010: Use OpenAPI contracts

**Status:** Accepted

Use separate contracts:

```text
contracts/openapi/operations.yaml
contracts/openapi/storefront.yaml
```

Generate frontend clients only after meaningful endpoints exist.

---

## ADR-011: Design for SaaS but launch with one brand

**Status:** Accepted

The future system supports multiple customer organisations.

The initial business assumption is:

```text
one organisation = one brand
```

Do not implement multi-brand organisations until required.

---

## ADR-012: Use shared-schema multitenancy later

**Status:** Accepted

Future tenant-owned records will carry:

```text
organisation_id
```

Repository APIs must require organisation context for tenant-owned operations.

PostgreSQL row-level security may later be introduced as defence in depth.

No tenancy tables or policies are implemented during bootstrap.

---

## ADR-013: Keep the repository publishable but intentionally incomplete

**Status:** Accepted

The initial GitHub publication represents an architecture and application bootstrap.

The README must explicitly state that the following are not implemented:

* authentication;
* catalogue;
* inventory;
* POS checkout;
* ordering;
* payment processing;
* billing;
* production deployment.

---

## ADR-014: Do not optimise for permanent zero-cost production

**Status:** Accepted

The initial project should support zero-cost local development and a publishable source repository.

Production reliability, hosting, backups, monitoring, email, domains, and payment processing will require future budget.

The architecture must not be distorted solely to remain permanently free.

---

# `.codex/CURRENT_STATE.md`

# Current Repository State

## Repository

```text
Repository: persona-apps
GitHub owner: Kangditya
```

## Current condition

The repository is intentionally empty except for `.codex` governance documents.

No application, package, infrastructure, or business files currently exist.

---

## Implemented

* Git repository initialization
* Repository-local Git identity
* GitHub remote preparation
* Canonical architecture decisions
* Agent workflow documentation

---

## Not implemented

### Repository tooling

* root `package.json`
* pnpm workspace
* Turborepo
* Makefile
* editor configuration
* linting
* formatting
* testing
* CI

### Operations application

* Vite application
* React application shell
* React Router
* Tailwind CSS
* PWA structure
* placeholder routes

### Storefront application

* Vite application
* React application shell
* React Router
* Tailwind CSS
* placeholder routes

### Backend

* Go module
* API process
* health endpoint
* readiness endpoint
* PostgreSQL connectivity

### Infrastructure

* Docker Compose
* PostgreSQL service
* environment example
* local development commands

### Product functionality

* authentication
* organisations
* outlets
* catalogue
* inventory
* POS
* sales
* customers
* carts
* orders
* payments
* fulfilment
* reporting

---

## Current milestone

Create a clean and GitHub-publishable monorepo foundation.

The milestone must include only:

* application shells;
* baseline routes;
* Go API shell;
* PostgreSQL local infrastructure;
* shared tooling;
* documentation;
* verification commands.

---

## Current verification state

No verification commands are available yet because the repository is empty.

Expected verification after bootstrap:

```bash
pnpm install
pnpm lint
pnpm typecheck
pnpm test
pnpm build

cd apps/api
go fmt ./...
go vet ./...
go test ./...
go build ./...

docker compose -f infrastructure/compose.yaml config
```

---

## Known constraints

1. The developer is working alone.
2. The initial budget is minimal.
3. The repository must remain understandable and maintainable.
4. The first task must not implement business features.
5. PostgreSQL may already run locally for another project.
6. The bootstrap PostgreSQL container must use a configurable host port.
7. No secrets may be committed.
8. The repository must be suitable for public GitHub publication.

---

## Next task

Execute `.codex/TASK.md` after a plan-only review.

---

# `.codex/TASK.md`

# Task: Bootstrap Single-Brand Commerce Monorepo

## Objective

Create a clean, buildable, and GitHub-publishable monorepo foundation for a future single-brand commerce SaaS combining:

* point-of-sale operations;
* back-office management;
* public storefront;
* shared catalogue;
* shared inventory.

The repository is currently empty.

This task is limited to repository structure, application shells, shared tooling, local PostgreSQL infrastructure, documentation, and baseline verification.

---

## Required Workflow

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

Planning must be completed and reviewed before implementation begins.

---

## Repository Structure

Create the following canonical structure:

```text
.
├── apps/
│   ├── operations-web/
│   ├── storefront-web/
│   └── api/
├── packages/
│   ├── ui/
│   ├── api-client/
│   ├── eslint-config/
│   └── typescript-config/
├── contracts/
│   └── openapi/
├── infrastructure/
│   ├── compose.yaml
│   ├── nginx/
│   └── postgres/
├── docs/
│   ├── architecture/
│   ├── decisions/
│   └── product/
├── .github/
│   └── workflows/
├── .codex/
├── .editorconfig
├── .env.example
├── .gitignore
├── LICENSE
├── Makefile
├── README.md
├── package.json
├── pnpm-workspace.yaml
└── turbo.json
```

The exact file set must be proposed in the plan before implementation.

---

## Operations Frontend

Create:

```text
apps/operations-web
```

Use:

* Vite
* React
* TypeScript
* React Router
* Tailwind CSS v4
* Oxlint-compatible configuration
* PWA-ready directory structure

Required source structure:

```text
src/
├── app/
├── routes/
│   ├── paths.ts
│   └── routes.tsx
├── layouts/
├── pages/
├── features/
├── components/
│   └── ui/
├── lib/
├── pwa/
├── styles/
│   └── global.css
├── main.tsx
└── vite-env.d.ts
```

Required placeholder routes:

```text
/login
/management
/pos
```

Requirements:

* Route definitions must live in `src/routes/routes.tsx`.
* URL paths must live in `src/routes/paths.ts`.
* Pages may display simple placeholder content.
* Do not implement authentication or route protection.
* Do not implement POS behaviour.
* Do not implement PWA service-worker behaviour yet.
* Tailwind must be configured and visibly verified.

---

## Storefront Frontend

Create:

```text
apps/storefront-web
```

Use:

* Vite
* React
* TypeScript
* React Router
* Tailwind CSS v4
* Oxlint-compatible configuration

Required source structure:

```text
src/
├── app/
├── routes/
│   ├── paths.ts
│   └── routes.tsx
├── layouts/
├── pages/
├── features/
├── components/
│   └── ui/
├── lib/
├── styles/
│   └── global.css
├── main.tsx
└── vite-env.d.ts
```

Required placeholder routes:

```text
/
/products
/cart
```

Requirements:

* Route definitions must live in `src/routes/routes.tsx`.
* URL paths must live in `src/routes/paths.ts`.
* Pages may display simple placeholder content.
* Do not implement product fetching.
* Do not implement customer accounts.
* Do not implement cart state.
* Do not implement checkout.

---

## Go API

Create:

```text
apps/api
```

Initialize the Go module as:

```text
github.com/Kangditya/persona-apps/apps/api
```

Required structure:

```text
apps/api/
├── cmd/
│   ├── api/
│   │   └── main.go
│   └── migrate/
├── internal/
│   ├── app/
│   ├── config/
│   ├── platform/
│   │   ├── database/
│   │   ├── logger/
│   │   ├── middleware/
│   │   └── validator/
│   └── modules/
├── migrations/
├── queries/
├── tests/
├── go.mod
├── go.sum
└── Dockerfile
```

Required endpoints:

```text
GET /health
GET /ready
```

Behaviour:

* `/health` returns success when the API process is running.
* `/ready` checks PostgreSQL connectivity.
* Use explicit timeouts.
* Return structured JSON responses.
* Handle shutdown signals gracefully.
* Do not implement business modules.
* Do not create business migrations.

---

## PostgreSQL Infrastructure

Create:

```text
infrastructure/compose.yaml
```

Requirements:

* Use a PostgreSQL container.
* Do not specify a fixed `container_name`.
* Use a project-local named volume.
* Use a configurable host port.
* Default host port must avoid conflicting with a local PostgreSQL on `5432`.

Recommended default:

```text
POSTGRES_PORT=5433
```

Inside Docker, PostgreSQL remains on:

```text
5432
```

Required environment variables:

```dotenv
POSTGRES_PORT=5433
POSTGRES_DB=persona_apps
POSTGRES_USER=persona_apps
POSTGRES_PASSWORD=persona_apps
DATABASE_URL=postgres://persona_apps:persona_apps@localhost:5433/persona_apps?sslmode=disable
```

Store only example development values in `.env.example`.

Do not commit `.env`.

The Compose configuration must include a PostgreSQL health check.

---

## Shared Packages

Create minimal package shells for:

```text
packages/ui
packages/api-client
packages/eslint-config
packages/typescript-config
```

### `packages/ui`

May export one minimal shared presentational component.

Do not introduce a complete design system.

### `packages/api-client`

Create a placeholder package only.

Do not generate a client because no meaningful OpenAPI endpoints exist yet.

### `packages/eslint-config`

If Oxlint is used exclusively, this package may instead document or reserve shared lint configuration.

Do not add ESLint unless required by a concrete unsupported rule.

### `packages/typescript-config`

Provide reusable baseline TypeScript configurations for Vite React packages and shared libraries.

---

## Root Tooling

Use:

* pnpm workspace;
* Turborepo;
* Oxlint;
* TypeScript compiler;
* Prettier or another explicitly selected formatter;
* Go native tooling;
* Docker Compose.

Required root commands:

```text
pnpm dev
pnpm build
pnpm lint
pnpm typecheck
pnpm test
pnpm format
```

Required Make targets:

```text
make install
make dev
make build
make lint
make typecheck
make test
make format
make validate
make infra-up
make infra-down
```

Keep commands explicit and understandable.

---

## Documentation

Create:

```text
README.md
docs/architecture/overview.md
docs/product/scope.md
docs/decisions/README.md
```

The root README must state that the repository is currently an initial bootstrap.

It must explicitly state that these features are not yet implemented:

* authentication;
* catalogue;
* inventory;
* POS checkout;
* customer ordering;
* payment processing;
* billing;
* production deployment.

Include local setup instructions.

---

## GitHub Configuration

Create a minimal validation workflow:

```text
.github/workflows/validate.yml
```

The workflow should validate:

* frontend installation;
* linting;
* type checking;
* tests;
* frontend builds;
* Go formatting check;
* Go vet;
* Go tests;
* Go build.

Do not add deployment workflows.

---

## Constraints

Do not implement:

* authentication;
* authorisation;
* tenant middleware;
* organisation CRUD;
* outlet CRUD;
* product CRUD;
* inventory;
* cashier shifts;
* POS checkout;
* online orders;
* payment processing;
* refunds;
* subscriptions;
* billing;
* email;
* background workers;
* Redis;
* message brokers;
* microservices;
* Kubernetes;
* production deployment;
* production observability;
* offline mutation queues.

Do not:

* commit secrets;
* create `.env`;
* commit build output;
* commit generated API clients;
* perform unrelated work;
* commit or push changes.

---

## Planning Requirements

Before coding, produce:

1. concise architecture assessment;
2. final proposed directory tree;
3. dependency list with justification;
4. planned file-change table;
5. application bootstrap sequence;
6. verification plan;
7. risks and unresolved decisions.

Stop after planning for review.

---

## Verification Requirements

After implementation, run all applicable commands.

At minimum:

```bash
pnpm install
pnpm lint
pnpm typecheck
pnpm test
pnpm build

cd apps/api
gofmt -l .
go vet ./...
go test ./...
go build ./...

docker compose -f infrastructure/compose.yaml config
```

Also verify:

```bash
git diff --check
git status --short
```

`gofmt -l .` must produce no file paths.

---

## Completion Report

Finish with:

### Changes

Summarise the implemented foundation.

### Files

List created and modified files by area.

### Dependencies

List all added dependencies and why they were needed.

### Plan variance

List deviations from the approved plan.

Write:

```text
None
```

when there was no variance.

### Verification

Report each command and its actual result.

### Remaining risks

Document only real remaining bootstrap concerns.

### Next recommended task

Recommend one narrow next vertical task without implementing it.
