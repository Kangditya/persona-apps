# Current Repository State

## Repository

```text
Repository: persona-apps
GitHub owner: Kangditya
```

## Current condition

The repository bootstrap is implemented. It contains application shells,
shared tooling, local infrastructure, contracts, documentation, and CI.

No business functionality exists yet.

---

## Implemented

* Git repository initialization
* Repository-local Git identity
* GitHub remote preparation
* Canonical architecture decisions
* Agent workflow documentation
* pnpm workspace and Turborepo commands
* Vite, React, React Router, Tailwind CSS, TypeScript, Oxlint, and Vitest shells
* operations routes: `/login`, `/management`, `/pos`
* storefront routes: `/`, `/products`, `/cart`
* Go API process with `/health`, PostgreSQL-backed `/ready`, and graceful shutdown
* Docker Compose PostgreSQL service with configurable host port
* OpenAPI placeholders, shared package shells, documentation, and validation CI

---

## Not implemented

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

## Completed milestone

The clean and GitHub-publishable monorepo foundation includes only:

* application shells;
* baseline routes;
* Go API shell;
* PostgreSQL local infrastructure;
* shared tooling;
* documentation;
* verification commands.

---

## Current verification state

Verified on 2026-07-30:

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
docker build -f apps/api/Dockerfile apps/api
make validate
```

Live verification confirmed `/health` remains successful while `/ready`
changes from success to unavailable when PostgreSQL stops.

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

Define the next narrow vertical slice in a separate reviewed task.

---
