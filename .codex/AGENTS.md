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
