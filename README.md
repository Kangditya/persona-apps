# Brand Commerce

Monorepo foundation for a future single-brand commerce platform
combining:

- back-office operations;
- point of sale;
- public storefront;
- shared catalogue and inventory.

## Current status

Initial architecture and repository bootstrap.

The following are not implemented:

- authentication;
- product catalogue;
- inventory;
- POS checkout;
- customer ordering;
- payment processing;
- production deployment.

## Applications

- `apps/operations-web`
- `apps/storefront-web`
- `apps/api`

## Local requirements

- Node.js
- pnpm
- Go
- Docker with Docker Compose
- Git

## Development

```bash
pnpm install
docker compose -f infrastructure/compose.yaml up -d postgres
make dev
