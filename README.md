# Qurban Commerce and Operations Platform

Initial monorepo bootstrap for a Qurban Commerce and Operations Platform.

## Current status

This repository contains React application shells, shared workspace tooling, a Go modular-monolith API shell, PostgreSQL infrastructure, placeholder OpenAPI contracts, and canonical product and architecture documentation.

The documented frontend direction is React Router with Remix-style routing
conventions and TanStack Query for remote API/server state. Both Vite apps now
include a typed transport boundary and an API-availability diagnostic only; no
qurban business endpoint or workflow is implemented.

No qurban business functionality is implemented yet. The following remain deferred:

- Qurban Event and Offering Catalogue;
- Common, Saving, and Giveaway Purchasing;
- Payment Verification and Funding;
- Party, Participant, and Sohibul Qurban activation;
- Livestock and Allocation;
- Slaughter and Distribution operations;
- authentication and authorization;
- reporting projections and operational dashboards;
- payment gateway integration;
- production deployment.

The first planned business vertical slice is:

```text
Qurban Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Basic Operations Dashboard
```

## Applications

- `apps/storefront-web`: public event, offering, purchasing, payment interaction, and purchase-tracking shell;
- `apps/operations-web`: internal event, purchasing, payment verification, participant, livestock, allocation, and distribution operations shell;
- `apps/api`: Go modular-monolith API shell.

## Requirements

- Node.js 24 or newer;
- pnpm 10.30.0;
- Go 1.26 or newer;
- Docker with Docker Compose;
- Make;
- Air for Go API live reload: `go install github.com/air-verse/air@latest`.

## Installation

### 1. Clone and enter the repository

```bash
git clone https://github.com/Kangditya/persona-apps.git
cd persona-apps
```

### 2. Configure local environment

```bash
cp .env.example .env
```

The example configuration uses PostgreSQL host port `5433` so it can coexist with a local PostgreSQL server on `5432`. Adjust `.env` if required.

### 3. Install the frontend workspace

Install the JavaScript dependencies once from the repository root:

```bash
pnpm install
```

The frontend applications are workspace packages. Do not run separate `npm install` commands inside either frontend directory.

### 4. Install the API dependencies

```bash
cd apps/api
go mod download
cd ../..
```

### 5. Install the API live-reload tool

Air is a development-only tool and is not a runtime application dependency:

```bash
go install github.com/air-verse/air@latest
```

Make sure `$(go env GOPATH)/bin` is on `PATH` so `air` can be found by `make dev-api`.

### 6. Start PostgreSQL

```bash
make infra-up
```

Stop it with:

```bash
make infra-down
```

## Run each application separately

The local development servers are intentionally separate so each application can be developed and restarted independently.

Before starting the API, load `.env` in the shell if it is not already exported:

```bash
set -a
. ./.env
set +a
```

### Storefront Web

Run from the repository root:

```bash
make dev-storefront
```

Or run directly from the app directory:

```bash
cd apps/storefront-web
pnpm dev
```

URL: `http://127.0.0.1:5173`

Vite provides lightweight static serving and browser live reload through HMR.

### Operations Web

Run from the repository root:

```bash
make dev-operations
```

Or run directly from the app directory:

```bash
cd apps/operations-web
pnpm dev
```

URL: `http://127.0.0.1:5174`

Vite provides lightweight static serving and browser live reload through HMR.

### Go API

Run from the repository root:

```bash
make dev-api
```

Or run directly from the API directory:

```bash
cd apps/api
air
```

Air rebuilds and restarts the API when Go source files change. The API requests port `8080` by default. If that port is already occupied, `make dev-api` automatically selects the next available port and prints the address it chose.

To request a specific starting port:

```bash
HTTP_PORT=18080 make dev-api
```

The selected API address should be used by any local frontend API configuration that needs to call the server.

## Frontend API configuration

Both applications read these Vite variables at build time:

```text
VITE_API_BASE_URL=/api
VITE_API_PROXY_TARGET=http://127.0.0.1:8080
VITE_API_PROVIDER=api
```

For local development, Vite proxies `/api` to `VITE_API_PROXY_TARGET`; the
production value may instead be a same-origin API path or an API origin with
an appropriate CORS policy. `api` calls the existing Go API `GET /health`
endpoint. Set
`VITE_API_PROVIDER=development` to use a deterministic, non-authoritative
diagnostic adapter without credentials or a running API. It returns only
`{ "status": "development" }`; it does not represent qurban product data.

The Storefront and Operations endpoint modules remain application-owned.
They share only `@persona-apps/api-client`, which owns request serialization,
timeouts, cancellation, response parsing, and normalized errors. It exposes a
`getHeaders` extension point for a future session provider, but no credentials
are stored or logged because authentication is not yet contracted.

Future public endpoints belong in `apps/storefront-web/src/api`; operations
endpoints belong in `apps/operations-web/src/api`, using their respective
OpenAPI contract. Add a TanStack Query key beside each endpoint and invalidate
only affected keys after a successful, contracted mutation.

Health checks:

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/ready
```

`/ready` requires PostgreSQL to be running.

## Run the complete local stack

After installing dependencies, configuring `.env`, installing Air, and starting PostgreSQL:

```bash
make dev
```

This starts three independent development processes:

- Storefront Web on `127.0.0.1:5173`;
- Operations Web on `127.0.0.1:5174`;
- Go API with Air live reload on `127.0.0.1:8080`.

Use `Ctrl+C` to stop the development processes.

## Validation

```bash
make validate
```

This runs formatting checks, linting, type checking, tests, builds, Go checks, and Docker Compose configuration validation.

Canonical product and architecture documents are under `docs/`:

- `docs/PRD.md`;
- `docs/PRODUCT_MAP.md`;
- `docs/ARCHITECTURE.md`;
- `docs/DECISIONS.md`.
