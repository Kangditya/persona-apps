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
