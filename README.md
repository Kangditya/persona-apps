# Qurban Commerce and Operations Platform

Initial monorepo bootstrap for a Qurban Commerce and Operations Platform.

## Current status

This repository contains two Next.js React applications, shared workspace
tooling, a Go modular-monolith API, PostgreSQL infrastructure, separate OpenAPI
contracts, and canonical product and architecture documentation.

Both web applications use Next.js 16 App Router with TanStack Query for remote
API/server state. They keep separate public and Operations API boundaries. The
Go API serves guest Event/Offering discovery, authenticated Operations
Event/Offering commands, Party persistence, authorized Purchase reads, and
atomic guest `COMMON` Purchase creation with captured snapshots, quota
reservation, reference/token safety, outbox effects, and durable encrypted
idempotency replay.

Event/Offering configuration, PostgreSQL persistence, guest catalogue
discovery, and the Week 3 Common Purchase backend are implemented. The
following remain unimplemented:

- Storefront checkout and Operations Purchase screens;
- Payment submission, Verification, and Sohibul Qurban activation;
- Livestock and Allocation;
- Slaughter and Distribution operations;
- field teams, shifts, readiness, check-in, incidents, and support escalation;
- customer event-day status, notifications, and completion documents;
- realtime projections, polling/SSE, and bounded degraded-connectivity field
  replay;
- reporting projections and operational dashboards;
- payment gateway integration;
- production deployment.

The approved Full Event-Day MVP sequence is:

```text
Qurban Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Livestock and Pen Assignment
→ Allocation
→ Slaughter Execution
→ Distribution
→ Customer Event-Day Status
→ Realtime Multi-Team Mobile Operations
```

The approved planning baseline is
[`MVP-DELIVERY-ROADMAP.md`](MVP-DELIVERY-ROADMAP.md). The old two-month plan is
superseded because it deferred the event-day capabilities required to operate
Eid al-Adha for 3–4 days.

## Applications

- `apps/storefront-web`: public event, offering, purchasing, payment interaction, and purchase-tracking shell;
- `apps/operations-web`: internal event, purchasing, payment verification, participant, livestock, allocation, and distribution operations shell;
- `apps/api`: Go modular-monolith API with public Event/Offering discovery.

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

Next.js provides the App Router development server and browser live reload.

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

Next.js provides the App Router development server and browser live reload.

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

## Shared UI and PWA foundation

Both applications consume the domain-agnostic `@persona-apps/ui` workspace
package. It owns Tailwind v4 semantic tokens and editable shadcn-style atoms,
molecules, and patterns. Native HTML owns simple controls; React Aria
Components owns the composite dialog, menu, and sheet behavior. The package has
no routes, API access, authentication, environment reads, or qurban rules.

Each application also has an independent production PWA configuration:

| App        | Manifest identity                     | Scope                        | Offline policy       |
| ---------- | ------------------------------------- | ---------------------------- | -------------------- |
| Storefront | Qurban Storefront (`/storefront-web`) | `/` on the storefront origin | immutable shell only |
| Operations | Qurban Operations (`/operations-web`) | `/` on the operations origin | immutable shell only |

Each manifest declares the app-owned `icon-192.svg` and `icon-512.svg` assets;
the 512px icon is marked `maskable` and has safe centered artwork. The service
workers precache immutable Next.js JavaScript/CSS/font assets, the icons, and a
data-free offline fallback. Successful HTML and API responses are not cached.
API requests, authentication, participant, financial, operational, and
mutation data have no cache or replay path. Offline mode shows an unavailable
notice; it never presents cached records as authoritative. Updates use a
visible prompt and do not activate or reload automatically during work.

Service workers are disabled during development. Production builds emit each
app's manifest, icon, and worker. Keep the applications on separate origins or
configure a distinct deployment base path before hosting them on one origin.

## Frontend API configuration

Browser-visible values use Next.js public variables; the proxy target remains
server-only:

```text
NEXT_PUBLIC_API_BASE_URL=
API_PROXY_TARGET=http://127.0.0.1:8081
NEXT_PUBLIC_API_PROVIDER=api
```

For local development, Next.js rewrites canonical `/api` paths unchanged and
also rewrites `/health` and `/ready` to `API_PROXY_TARGET`. Deployed same-origin
routing belongs to ingress or a reverse proxy, so omit that value when the
Next.js runtime should not proxy locally. Set
`NEXT_PUBLIC_API_PROVIDER=development` to use a deterministic,
non-authoritative health diagnostic without a running API. It returns only
`{ "status": "development" }`; it does not represent qurban product data.

The Storefront and Operations endpoint modules remain application-owned.
They share only `@persona-apps/api-client`, which owns request serialization,
timeouts, cancellation, response parsing, and normalized errors. Operations
keeps its HttpOnly session browser-managed and its rotated CSRF value in
TanStack Query memory. No credentials are stored or logged by the generic
client.

Future public endpoints belong in `apps/storefront-web/src/api`; operations
endpoints belong in `apps/operations-web/src/api`, using their respective
OpenAPI contract. Add a TanStack Query key beside each endpoint and invalidate
only affected keys after a successful, contracted mutation.

Health checks:

```bash
curl http://127.0.0.1:8081/health
curl http://127.0.0.1:8081/ready
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
- Go API with Air live reload on the `.env` HTTP address (`127.0.0.1:8081` in
  the example).

Use `Ctrl+C` to stop the development processes.

## Production-mode web runtime

Each web application builds and starts independently:

```bash
pnpm --filter @persona-apps/storefront-web build
pnpm --filter @persona-apps/storefront-web start

pnpm --filter @persona-apps/operations-web build
pnpm --filter @persona-apps/operations-web start
```

The commands listen on ports 5173 and 5174 respectively. Production hosting
must provide a supported Node.js runtime and same-origin ingress routing for
`/api`, `/health`, and `/ready`; this repository does not provision a provider
or frontend container.

## Validation

```bash
make validate
```

This runs formatting checks, linting, type checking, tests, builds, Go checks, and Docker Compose configuration validation.

Canonical product and architecture documents are under `docs/`:

- `docs/PRD.md`;
- `docs/PRODUCT_MAP.md`;
- `docs/ARCHITECTURE.md`;
- `docs/DECISIONS.md`;
- `docs/CONVENTIONS.md`.

Delivery and execution planning:

- `MVP-DELIVERY-ROADMAP.md`;
- `.codex/CURRENT_STATE.md`;
- `.codex/plans/` for inactive header-only task drafts;
- `.codex/TASK.md` for exactly one active reviewed task when present.
