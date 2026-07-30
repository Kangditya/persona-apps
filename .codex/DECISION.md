# Architecture Decisions

## ADR-001: Repository structure

Use a monorepo containing:

- operations frontend;
- public storefront;
- Go API;
- shared frontend packages;
- OpenAPI contracts;
- local infrastructure.

## ADR-002: Frontend stack

Use Vite, React, TypeScript, React Router, and pnpm.

## ADR-003: Frontend applications

Use two frontend applications:

- `apps/operations-web` for back-office and POS;
- `apps/storefront-web` for the public storefront.

## ADR-004: Backend stack

Use one Go modular monolith under `apps/api`.

Do not introduce microservices.

## ADR-005: Database

Use PostgreSQL.

During bootstrap, implement connectivity and readiness checks only.
Do not create business schemas.

## ADR-006: Product boundary

The future product is a single-brand commerce SaaS combining:

- POS;
- back-office operations;
- public storefront;
- shared catalogue;
- shared inventory.

The bootstrap task must not implement those features yet.
