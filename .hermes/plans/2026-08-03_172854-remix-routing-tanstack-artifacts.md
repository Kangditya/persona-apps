# Remix-Style Routing and TanStack Query Artifact Alignment Plan

> **For Hermes:** Execute only after this plan is reviewed and explicitly approved. Do not modify application code, package manifests, lockfiles, generated clients, or runtime behavior.

**Goal:** Establish the frontend architectural direction for Remix-style routing and TanStack Query in the repository’s canonical and execution-context artifacts, while preserving the current Vite/React application shells and leaving implementation deferred.

**Architecture:** The Storefront and Operations applications remain separate Vite-served React SPAs backed by the Go modular monolith and their separate OpenAPI contracts. The artifact direction will document React Router’s Remix-style route module/data-navigation conventions for application navigation, and TanStack Query as the client-side server-state boundary. Until a vertical slice is approved, neither routing nor query-client runtime code, dependencies, data loaders, API calls, caches, mutations, or generated clients will be introduced.

**Tech Stack Direction:** Vite, React, TypeScript, React Router with Remix-style routing conventions, TanStack Query, Tailwind CSS, Vitest, pnpm/Turborepo.

---

## Current context

- Both applications currently declare `react-router` but do not declare a TanStack package:
  - `apps/storefront-web/package.json:13-17`
  - `apps/operations-web/package.json:13-17`
- The accepted architecture currently names React Router, requires centralized route registries, and distinguishes server state from local UI state:
  - `docs/DECISIONS.md:124-178`
  - `docs/ARCHITECTURE.md:733-765`
- The repository has no business API endpoints or generated clients. The separate OpenAPI contracts are intentionally empty placeholders:
  - `contracts/openapi/storefront.yaml`
  - `contracts/openapi/operations.yaml`
- The active frontend shells contain only capability-aligned placeholder routes. They must remain unchanged in this artifact-only task.
- Existing user-owned working-tree changes are present in `.gitignore` and `DESIGN.md`; this task must not overwrite or include them.

## Constraints to preserve

1. The initial deployment remains a Vite/React SPA architecture; this is not an authorization to add Remix server runtime, SSR, loaders that call the backend, or server actions.
2. Storefront and Operations remain separate applications with distinct public and operations API contracts.
3. The Go API remains authoritative for business rules, authorization, transactions, auditing, idempotency, concurrency, and lifecycle changes.
4. TanStack Query cache data is never transactional truth. It must not authoritatively calculate or resolve payment status, quota, allocation capacity, saving balance, or slaughter queue position.
5. Product data boundaries remain intact: `COMMON`, `SAVING`, and `GIVEAWAY` converge into canonical Purchase; Sohibul Qurban and all funding/participant roles remain distinct.
6. This task is artifacts only. No code, test, dependency, lockfile, OpenAPI endpoint, package-manifest, migration, generated-client, or route-file change is allowed.
7. Do not commit or push.

## Terminology and architectural decisions to record

### Routing

- “Remix-style routing” means route ownership follows explicit route-module conventions: route hierarchy, layouts, path parameters, route-level boundaries, navigation state, and deferred route data requirements are designed together.
- In the current SPA topology, those conventions are implemented through React Router rather than introducing a Remix server runtime.
- Central registries at `src/routes/paths.ts` and `src/routes/routes.tsx` remain required. Future route modules may be feature-owned but must register through the application-owned route registry.
- Route protection provides user-experience gating only; backend authorization remains mandatory.
- Route-level data requirements must specify public vs operations contract ownership and avoid exposing operations DTOs in Storefront.

### Server state

- “TanStack” is assumed to mean **TanStack Query** (`@tanstack/react-query`) for remote API/server state, rather than TanStack Router, Table, Form, or another TanStack library.
- TanStack Query will own query caching, request lifecycle, invalidation after successful commands, retries appropriate to idempotency, and explicit loading/error/stale/conflict rendering.
- Local UI state remains component/feature state. Forms remain local until a future approved form-library decision; no TanStack Form adoption is implied.
- Query keys must be application-scoped and based on stable public API identifiers and explicit event context. Cache invalidation is a UI refresh mechanism, not correctness or authorization enforcement.
- High-contention state must be revalidated by the Go API. A `409 Conflict` response must be displayed as a conflict and must not be resolved locally from cache.

## Planned artifact changes

| File                      | Action              | Purpose                                                                                                                                                                                     |
| ------------------------- | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `docs/DECISIONS.md`       | modify              | Supersede the React Router-only frontend decision with a precise SPA-compatible Remix-style routing decision; add an accepted TanStack Query decision and explicit non-adoption boundaries. |
| `docs/ARCHITECTURE.md`    | modify              | Define frontend route-module conventions, route/data boundary ownership, TanStack Query responsibilities, query-key/invalidation conventions, conflict behavior, and cache safety limits.   |
| `.codex/AGENTS.md`        | modify              | Replace React Router-only technology guidance with the approved routing and server-state conventions; retain centralized registry and backend-authority rules.                              |
| `.codex/ARCHITECTURE.md`  | modify              | Reconcile its duplicated frontend-architecture guidance with canonical `docs/ARCHITECTURE.md`, or replace it with a concise execution-context reference to prevent divergence.              |
| `.codex/CURRENT_STATE.md` | modify              | Record that the architecture is documented but no Remix-style route modules, TanStack Query provider, dependencies, or server-state integration are implemented.                            |
| `.codex/TASK.md`          | modify              | Replace the stale realignment-only objective with the artifact-alignment objective and the explicit no-code/no-dependency constraint.                                                       |
| `README.md`               | modify              | State the intended frontend architecture and accurately distinguish the documented direction from the unimplemented shell state.                                                            |
| `.codex/ARTIFACTS.md`     | modify or deprecate | Remove stale generic/POS frontend guidance and prevent it from contradicting the canonical docs; retain only a clearly labelled historical/reference role if it must stay.                  |

## Explicitly unchanged

- `apps/storefront-web/**`
- `apps/operations-web/**`
- `apps/*/package.json`
- `pnpm-lock.yaml`
- root `package.json`
- `contracts/openapi/storefront.yaml`
- `contracts/openapi/operations.yaml`
- `packages/**`
- `apps/api/**`
- database migrations and infrastructure

No dependencies will be installed or declared. The future implementation task will add the narrowest app-level dependency only after its first concrete remote-data vertical slice is approved.

## Build sequence after approval

1. Re-read the planned artifact files and their current diffs; preserve all unrelated user edits.
2. Update `docs/DECISIONS.md` first, assigning new ADR identifiers without renumbering historical decisions. Mark only the routing portion of ADR-003 superseded if the decision register’s established conventions require it; otherwise add a narrowly scoped successor ADR.
3. Update `docs/ARCHITECTURE.md` so its frontend architecture and caching sections conform to the accepted decisions.
4. Align the repository and execution-context artifacts: `.codex/AGENTS.md`, `.codex/ARCHITECTURE.md`, `.codex/CURRENT_STATE.md`, `.codex/TASK.md`, `.codex/ARTIFACTS.md`, and `README.md`.
5. Search all tracked Markdown artifacts for stale claims that React Router alone is the complete frontend architecture or that TanStack Query is already installed/implemented.
6. Validate Markdown formatting, repository checks that do not require code changes, and the final diff scope.

## Dependencies, API, database, and behavior impact

- Dependencies: none in this task. The documentation may name the future package `@tanstack/react-query`, but it must explicitly state that the package is not yet a repository dependency.
- Application code and runtime behavior: none.
- API and OpenAPI: none. Both contracts remain separate and endpoint-free placeholders.
- Database/migrations: none.
- Authorization/audit/idempotency/concurrency: none implemented. The architecture artifacts will make the required backend ownership and cache limitations explicit.

## Verification after artifact updates

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

Targeted review:

```bash
git diff -- docs/DECISIONS.md docs/ARCHITECTURE.md README.md .codex/AGENTS.md .codex/ARCHITECTURE.md .codex/CURRENT_STATE.md .codex/TASK.md .codex/ARTIFACTS.md
rg -n -i "react router|remix|tanstack|server state|query client|query cache" docs README.md .codex
```

Expected result: documentation consistently describes the future routing/server-state architecture; no code or dependency files appear in the diff; all validation commands still pass.

## Risks and open question

- “TanStack” is a library family. This plan assumes **TanStack Query**. If the request instead means TanStack Router, Table, Form, or multiple packages, the decision text and future implementation scope change materially.
- Remix-style routing is compatible with the retained Vite SPA topology only when it refers to routing conventions and React Router integration, not Remix server runtime/SSR. A later SSR requirement needs a separate ADR.
- No live API exists to validate query-key shapes, stale-time policy, mutation invalidation, or route data ownership. The artifacts must establish principles without inventing endpoints or domain DTOs.

## Completion criteria

- All planned documentation and agent-context artifacts consistently record the approved direction.
- The working tree contains no edits to source, test, manifest, lockfile, OpenAPI, database, or infrastructure files attributable to this task.
- Repository checks and diff hygiene pass.
- The review distinguishes documented architecture from deferred implementation.
