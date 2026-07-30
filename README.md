# Persona Apps

Initial monorepo bootstrap for a future single-brand commerce platform with
employee operations, point of sale, and a public storefront.

## Current status

This repository contains application shells, shared tooling, a Go API shell,
and local PostgreSQL infrastructure. It does not yet implement:

- authentication;
- catalogue;
- inventory;
- POS checkout;
- customer ordering;
- payment processing;
- billing;
- production deployment.

## Applications

- `apps/operations-web`: employee-facing operations shell;
- `apps/storefront-web`: public storefront shell;
- `apps/api`: Go modular-monolith shell.

## Requirements

- Node.js 24 or newer;
- pnpm 10.30.0;
- Go 1.26 or newer;
- Docker with Docker Compose;
- Make.

## Local setup

```bash
cp .env.example .env
make install
make infra-up
make dev
```

The API listens on `http://localhost:8080`; PostgreSQL is exposed on port
`5433` unless `POSTGRES_PORT` is overridden in `.env`.

Stop PostgreSQL with:

```bash
make infra-down
```

## Validation

```bash
make validate
```

This runs formatting checks, linting, type checking, tests, builds, and Docker
Compose configuration validation.

Architecture and product boundaries are documented under `docs/`.
