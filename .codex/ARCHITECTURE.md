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
