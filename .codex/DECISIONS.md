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
