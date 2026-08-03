# Task: Remix-Style Routing and TanStack Query Artifact Alignment

## Objective

Align the repository’s canonical documentation, execution context, and README
with the approved frontend direction:

- React Router with Remix-style routing conventions;
- TanStack Query (`@tanstack/react-query`) for future remote API/server state.

This task changes artifacts only. It does not implement routing, server state,
or qurban business functionality.

## Required workflow

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

The implementation plan must be reviewed and explicitly approved before BUILD.
Do not commit or push.

## Canonical documents

The canonical documents live at:

```text
docs/PRD.md
docs/PRODUCT_MAP.md
docs/ARCHITECTURE.md
docs/DECISIONS.md
```

`.codex/` is execution context only and must agree with the canonical
artifacts without becoming a competing source of architecture decisions.

## Scope

1. Record the SPA-compatible Remix-style routing conventions in the architecture
   and decision artifacts.
2. Record TanStack Query as the approved future server-state boundary.
3. Preserve centralized `src/routes/paths.ts` and `src/routes/routes.tsx`
   ownership in the documented architecture.
4. Document separate Storefront and Operations route-data/API-contract
   ownership.
5. Document cache limitations, explicit stale/error/conflict states, and Go API
   authority for contested operations.
6. Update execution context and README so they distinguish documented direction
   from the current unimplemented shell state.
7. Deprecate stale noncanonical artifact snapshots that contradict the accepted
   Qurban and frontend architecture.

## Explicit non-scope

Do not change:

- application source or tests;
- package manifests or `pnpm-lock.yaml`;
- dependencies, including adding `@tanstack/react-query`;
- Vite configuration;
- route files, route modules, loaders, API calls, query clients, queries,
  mutations, cache behavior, or providers;
- Remix server runtime, SSR, server actions, or deployment topology;
- OpenAPI endpoints, generated API clients, database migrations, or Go code;
- authentication, authorization, payments, or qurban business behavior;
- commits or pushes.

## Architecture constraints

- Runtime delivery remains two independent Vite-served React SPAs. “Remix-style”
  refers to route hierarchy, layouts, route boundaries, navigation state, and
  route-data requirements through React Router; it does not introduce Remix
  runtime features.
- TanStack Query represents only remote API/server state. Local UI state remains
  feature or component-local.
- Query cache data is not transactional truth and must not determine payment
  status, quota, allocation capacity, saving balance, or queue position.
- Contested operations are always revalidated by the Go API. A `409 Conflict`
  response must be rendered as a conflict state, not resolved from cached data.
- Storefront and Operations keep separate OpenAPI contracts and must not share
  sensitive operations DTOs through frontend routing or query code.
- Product rules remain unchanged: `COMMON`, `SAVING`, and `GIVEAWAY` converge
  into canonical Purchase; Sohibul Qurban and financial/participant roles remain
  distinct; dashboard data is a projection.

## Verification

At minimum:

```bash
pnpm run format:check
pnpm run lint
pnpm run typecheck
pnpm run test
pnpm run build
make validate
git diff --check
git status --short
```

Review the final diff to confirm it contains only approved artifact files and
no source, manifest, lockfile, OpenAPI, database, or infrastructure changes.
