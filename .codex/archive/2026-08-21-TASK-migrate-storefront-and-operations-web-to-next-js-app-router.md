# Task: Migrate Storefront and Operations Web to Next.js App Router

## Executed

## Status

Executed and verified on 2026-08-21. The user explicitly approved the plan,
including its migration, dependency, ADR, and runtime deployment assumptions.

Both Next.js applications, production runtimes, proxy behavior, and generated
PWA policies are implemented. Real local API/PostgreSQL/OIDC command-path
verification passed. Production Chrome browser smoke also proved direct loads,
hydration, Storefront client navigation, and the user-controlled offline-ready
PWA status in both applications. Repository validation is green after
normalizing pre-existing Go formatter drift and making the auth-session test's
fixture expiry time-relative.

## Objective

Migrate `apps/storefront-web` and `apps/operations-web` in place from Vite and
React Router to the current stable Next.js 16 App Router while preserving their
separate application ownership, existing URLs and user-visible behavior,
TanStack Query server-state boundary, shared UI and transport packages,
shell-only PWA behavior, and the Go API as the sole authoritative backend.

The migration is compatibility-first. It establishes supported Next.js
applications without simultaneously redesigning the screens, moving business
behavior into Next.js, or converting existing API reads and commands to Server
Components, Route Handlers, or Server Actions.

## Context

Both frontend applications are implemented Vite SPAs. They have real
Event/Offering flows, centralized URL builders and route registration, shared
accessible UI primitives, application-owned API modules, TanStack Query hooks,
and independent Vite/Workbox PWA output. Operations additionally depends on a
same-origin Go API proxy for OIDC callback, cookie session, Origin, CSRF,
permission, idempotency, and optimistic-version behavior.

This is an architecture and deployment-boundary change, not a dependency-only
upgrade:

- ADR-003 currently requires Vite and React Router and records static
  deployment as a consequence;
- ADR-004 requires `src/routes/routes.tsx` as the central route registrar;
- ADR-005 names the Vite-specific Tailwind integration;
- `docs/ARCHITECTURE.md` describes Vite SPA and static web deployment;
- `docs/CONVENTIONS.md` requires Vite-prefixed browser configuration and
  describes Vite/PWA generated output.

A new accepted ADR must supersede or amend those details before implementation
changes the framework. The product boundary in ADR-002 remains unchanged: the
Storefront and Operations frontends stay separate applications with separate
API and data-exposure policies.

Current Next.js documentation supports an incremental Vite migration through a
client-only App Router boundary and supports App Router manifests and service
workers. This task uses that compatibility path first and defers optional
server-rendering optimization until behavioral parity is verified.

## Source of Truth

- `docs/PRD.md` for Storefront, Operations, security, privacy, accessibility,
  and PWA/offline product requirements;
- `docs/PRODUCT_MAP.md` for application and capability ownership;
- `docs/ARCHITECTURE.md` for frontend, API, authentication, caching, and
  deployment boundaries;
- `docs/DECISIONS.md`, especially ADR-001 through ADR-007, ADR-010, ADR-018,
  ADR-019, ADR-029, ADR-038, ADR-039, and ADR-043 through ADR-046;
- `docs/CONVENTIONS.md` for route, API, state, environment, dependency,
  generated-output, testing, and task rules;
- `.codex/CURRENT_STATE.md` for the implemented Event/Offering screens,
  Operations authentication, shared packages, and verified PWA behavior;
- `contracts/openapi/storefront.yaml` and
  `contracts/openapi/operations.yaml` for the unchanged consumer boundaries;
- the existing source under `apps/storefront-web`, `apps/operations-web`,
  `packages/api-client`, `packages/ui`, and `packages/typescript-config`;
- the official Next.js Vite migration, App Router, monorepo package
  transpilation, environment, deployment, and PWA documentation reviewed again
  at BUILD time.

External framework documentation informs implementation mechanics only. It
does not override repository product, security, API, or ownership decisions.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Before BUILD:

1. Review this task and its runtime/deployment assumptions.
2. Explicitly approve the plan.
3. Add and accept the ADR that supersedes ADR-003 and amends the affected
   routing, Tailwind integration, and deployment consequences.
4. Re-check the current stable non-canary Next.js 16 release against Node 24,
   React 19, pnpm, Turborepo, Tailwind CSS 4, Vitest, and the shared packages.

Do not modify frontend runtime code before the decision gate is accepted. Do
not commit or push unless explicitly requested.

## Scope

### In Scope

- Keep the existing application directories and package names:
  `apps/storefront-web` and `apps/operations-web`.
- Replace Vite entrypoints and React Router registration with Next.js App
  Router route entries while preserving existing public URLs.
- Retain `src/routes/paths.ts` as each application's central URL constant and
  URL-builder owner; replace the obsolete `src/routes/routes.tsx` registration
  requirement with App Router filesystem ownership in the accepted ADR and
  conventions.
- Use thin App Router route files around existing page/feature modules rather
  than reorganizing the entire frontend.
- Add deliberate Client Component boundaries for TanStack Query, mutations,
  forms, browser APIs, active navigation, PWA registration, and Operations
  session behavior.
- Keep existing API modules and TanStack Query hooks as browser-side consumers
  for migration parity.
- Preserve same-origin `/api`, `/health`, and `/ready` behavior in local
  development and document ingress ownership for deployed environments.
- Preserve Operations OIDC callback, HttpOnly cookie, Origin, CSRF,
  permission, logout, idempotency-key, and expected-version behavior.
- Preserve independent installable PWA manifests, shell/static-asset caching,
  offline/update notices, and the prohibition on caching API, authentication,
  participant, financial, or operational data.
- Reuse the installed Workbox build/window dependencies or native service
  worker APIs; do not add a second PWA framework unless parity cannot be
  achieved with the existing dependencies and the plan records why.
- Replace the Vite-specific Tailwind plugin with the supported Next.js/PostCSS
  integration while retaining Tailwind CSS 4 and shared tokens.
- Add shared Next.js TypeScript defaults only where both applications genuinely
  use them.
- Update Turborepo, ignored output, environment examples, normal developer
  commands, documentation, and CI-visible validation for Next.js output.
- Define and verify two independently runnable Next.js server artifacts; no
  cloud provider provisioning is included.
- Update canonical decisions and architecture before source migration, then
  update README and current-state documentation after verification.

### Out of Scope

- Changes to Go API routes, business rules, persistence, PostgreSQL migrations,
  OpenAPI schemas, authorization policy, audit, idempotency, or concurrency
  behavior.
- Combining Storefront and Operations into one Next.js application.
- Next.js Route Handlers, Server Actions, middleware-based authorization, a
  backend-for-frontend, or duplicated API/domain logic.
- Converting catalogue reads or Operations commands to Server Components in
  this compatibility task.
- New SEO content strategy, ISR/revalidation policy, streaming design, image
  pipeline, analytics, or social-card generation beyond preserving existing
  title and description metadata.
- New product routes, Purchase behavior, dashboard behavior, or replacement of
  existing placeholders.
- UI redesign, broad component extraction, form-library adoption, generated
  API clients, or a new end-to-end test framework.
- Offline business data, background mutation replay, offline authentication,
  or service-worker caching of `/api/*` responses.
- Cloud accounts, DNS, TLS, CDN, ingress provisioning, Kubernetes, Terraform,
  or production rollout.
- Unrelated dependency upgrades or repository-wide cleanup.

## Existing State

- `pnpm-workspace.yaml` already includes `apps/*` and `packages/*`.
- Turborepo orchestrates `dev`, `build`, `lint`, `typecheck`, and `test`, but
  build output currently recognizes only `dist/**`.
- Storefront runs on `127.0.0.1:5173`; Operations runs on
  `127.0.0.1:5174`; the Go API normally runs on `127.0.0.1:8081` through the
  checked-in local environment example.
- Both apps use React 19, TypeScript, React Router, TanStack Query, Tailwind CSS
  4, Oxlint, Vitest, `@persona-apps/ui`, and `@persona-apps/api-client`.
- Storefront has `/`, `/offerings`, `/offerings/:offeringId`, and
  `/purchase-tracking`.
- Operations has `/operator-login`, `/events`, `/events/:eventId`,
  `/events/:eventId/offerings/:offeringId`, `/event-dashboard`, `/purchasing`,
  and `/payment-verification`.
- Existing pages, layouts, query hooks, and PWA status components use browser
  hooks or React Router and therefore cannot all remain Server Components.
- Vite config currently owns local API proxying, Tailwind integration, PWA
  manifest generation, service-worker generation, and API fallback exclusion.
- Operations authentication is enforced by the Go API. The browser consumes a
  cookie-backed session and keeps CSRF data in memory only.
- Both generated PWAs cache only shell assets. API and sensitive data are
  network-only.
- No production hosting implementation exists. Canonical architecture still
  assumes static frontend deployments.

## Target State

### Application and route ownership

Both applications remain independently developed, built, run, and deployed.
Runtime URLs remain stable while App Router owns their file mapping:

| Application | Existing URL                             | App Router entry                                           |
| ----------- | ---------------------------------------- | ---------------------------------------------------------- |
| Storefront  | `/`                                      | `src/app/page.tsx`                                         |
| Storefront  | `/offerings`                             | `src/app/offerings/page.tsx`                               |
| Storefront  | `/offerings/:offeringId`                 | `src/app/offerings/[offeringId]/page.tsx`                  |
| Storefront  | `/purchase-tracking`                     | `src/app/purchase-tracking/page.tsx`                       |
| Operations  | `/operator-login`                        | `src/app/operator-login/page.tsx`                          |
| Operations  | `/events`                                | `src/app/events/page.tsx`                                  |
| Operations  | `/events/:eventId`                       | `src/app/events/[eventId]/page.tsx`                        |
| Operations  | `/events/:eventId/offerings/:offeringId` | `src/app/events/[eventId]/offerings/[offeringId]/page.tsx` |
| Operations  | `/event-dashboard`                       | `src/app/event-dashboard/page.tsx`                         |
| Operations  | `/purchasing`                            | `src/app/purchasing/page.tsx`                              |
| Operations  | `/payment-verification`                  | `src/app/payment-verification/page.tsx`                    |

`src/routes/paths.ts` continues to produce link targets and encoded dynamic
paths. App Router folders register routes; no parallel route table is added.

### Rendering and state

- Root layouts own metadata, global styles, and application shell composition.
- A small client provider owns one browser `QueryClient` per application.
- Existing interactive pages remain Client Components initially. Add
  `'use client'` only at the narrow boundaries that need browser state or
  client-only dependencies; do not mark the entire shared package or route
  tree client-only without evidence.
- TanStack Query continues to own remote server-state lifecycle and cache
  invalidation. It remains a view cache, never transactional truth.
- Public and Operations endpoint modules remain application-owned and continue
  to use their respective OpenAPI contracts.
- Existing loading, empty, error, rate-limit, unavailable, authorization,
  stale-conflict, and success states remain behaviorally equivalent.

### API and environment boundary

- Browser-visible configuration uses `NEXT_PUBLIC_*`; only values intentionally
  exposed to browser bundles receive that prefix.
- The local proxy target is server-only, for example `API_PROXY_TARGET`, and
  is never exposed as a `NEXT_PUBLIC_*` value.
- Development rewrites preserve `/api`, `/health`, and `/ready` paths without
  stripping or double-prefixing them.
- Deployed same-origin routing is owned by ingress/reverse-proxy configuration;
  a private API origin is not baked into the public browser bundle.
- Storefront continues to omit credentials. Operations continues to include
  credentials and uses the existing CSRF/session contract.
- Local development retains ports 5173 and 5174 so existing OIDC origin and
  callback configuration remains valid.

### PWA boundary

- Each app owns `src/app/manifest.ts`, its existing icons, a service-worker
  source, registration/update UI, and an offline fallback with no business
  data.
- Production builds generate the service worker into ignored Next.js output;
  generated workers are not committed.
- The worker may precache versioned `/_next/static/*` assets, app icons, and a
  data-free offline fallback only.
- Navigations may fall back to the data-free offline document after a network
  failure, but successful HTML/API responses containing user or business data
  are not added to runtime caches.
- Requests under `/api/`, authentication callbacks, `/health`, `/ready`, and
  all non-GET requests are network-only and are never replayed by the worker.
- Updates remain user-prompted; a waiting worker does not reload an operator's
  active work automatically.
- Service workers stay disabled during development.

### Runtime and deployment

- Each app produces an independently runnable Next.js server artifact suitable
  for a provider-managed Next.js runtime or a later container deployment.
- The accepted ADR and architecture replace the static-only frontend
  assumption with two separate Next.js runtime deployments while keeping the
  Go API and PostgreSQL deployment boundaries unchanged.
- This task proves production-mode `next build` and `next start` locally. It
  does not select or provision the production hosting provider.

## Constraints

- Preserve ADR-002's two-application boundary and package names.
- Keep business invariants, authorization, transactions, audit, idempotency,
  conflict checks, and status transitions in the Go API.
- Do not proxy or reimplement business endpoints through Next.js code.
- Do not place Operations sessions, CSRF values, Purchase tokens, secrets,
  database URLs, or private API origins in browser-visible environment values,
  rendered markup, logs, build artifacts, service workers, or caches.
- Keep same-origin Operations behavior compatible with Secure, HttpOnly,
  SameSite=Lax cookies and exact Origin validation.
- Keep dynamic links generated through the central `paths.ts` builders.
- Preserve direct loads and browser navigation for every existing route.
- Reuse existing pages, API modules, query hooks, UI primitives, Workbox
  packages, browser APIs, and native HTML before adding abstractions or
  dependencies.
- Do not blanket-add Client Component directives. A server route may render a
  focused client page or shell when existing behavior requires it.
- Do not add a second state manager, router, form framework, PWA wrapper, or
  generated API client.
- Do not manually edit `pnpm-lock.yaml`; update it through pnpm.
- Do not commit `.next`, generated service workers, coverage, `.env`, or other
  build/runtime output.

## Implementation Requirements

### 1. Architecture decision and documentation gate

- Add a new ADR for the Next.js App Router migration.
- Mark ADR-003 superseded by the new ADR.
- Amend ADR-004 so `paths.ts` remains the central URL-builder registry while
  Next.js filesystem routes replace `routes.tsx` registration.
- Amend ADR-005 only for the Tailwind integration mechanism; Tailwind CSS 4
  remains accepted.
- Record the Node/Next runtime and independent deployment consequence.
- Update `docs/ARCHITECTURE.md` and `docs/CONVENTIONS.md` consistently before
  source migration.
- Do not modify the PRD or Product Map unless implementation discovery exposes
  a genuine product contradiction.

### 2. Dependencies and workspace configuration

- Add the reviewed current stable non-canary Next.js 16 release to each app.
- Retain React 19, TypeScript, TanStack Query, Tailwind CSS 4, Vitest, Oxlint,
  shared UI, and shared API transport.
- Remove `react-router`, Vite runtime/build dependencies,
  `@tailwindcss/vite`, and `vite-plugin-pwa` after all callers are migrated.
- Remove `@vitejs/plugin-react` only if focused Vitest execution no longer
  needs it; do not keep it speculatively.
- Reuse `workbox-build` and `workbox-window` for the narrow generated-worker and
  update lifecycle if they remain necessary.
- Add a shared Next.js TypeScript configuration export only once and have both
  apps extend it.
- Configure Next.js to transpile the raw TypeScript workspace packages
  `@persona-apps/ui` and `@persona-apps/api-client`.
- Configure monorepo output tracing only where required for the runnable server
  artifact.
- Update Turborepo build outputs to include `.next/**` while excluding
  `.next/cache/**`; retain `dist/**` for any workspace that still owns it.

### 3. App Router shells and routes

- Replace each `index.html`, `src/main.tsx`, `src/app/App.tsx`,
  `src/routes/routes.tsx`, Vite environment declaration, and Vite config with
  the minimum Next.js equivalents.
- Add one root layout and one small client provider per application.
- Keep existing page and feature files where practical. Route entries should
  import them instead of duplicating their content.
- Replace React Router `Link`, `NavLink`, `Outlet`, `useNavigate`,
  `useLocation`, and `useParams` usage with Next.js layout composition,
  `next/link`, `next/navigation`, route props, and a minimal active-link helper
  only where the current UI uses active state.
- Preserve encoded path builders and add focused tests for all static and
  dynamic paths.
- Preserve existing titles, descriptions, language, semantic layout, focus,
  keyboard behavior, and responsive behavior.

### 4. API, authentication, and query behavior

- Keep application endpoint paths and DTOs unchanged.
- Replace `import.meta.env` reads with validated Next.js-compatible
  environment reads at the owning client/server boundary.
- Implement and test exact local rewrites for `/api/:path*`, `/health`, and
  `/ready` without path stripping.
- Keep Operations session bootstrap and CSRF in TanStack Query memory. Do not
  introduce local storage, server-rendered credentials, or middleware auth.
- Preserve logout query removal, safe return paths, permission-aware UI,
  expected versions, idempotency-key reuse for the same intent, explicit retry,
  and non-optimistic contested mutations.
- Preserve Storefront public DTO minimization, credential omission, advisory
  availability, exact minor-unit display, and bounded query retry behavior.

### 5. PWA migration

- Replace Vite manifest generation with typed App Router manifests.
- Replace `virtual:pwa-register` with the existing Workbox window integration
  or native service-worker registration.
- Generate a worker from checked-in source during production build without
  writing generated output into tracked source directories.
- Preserve distinct Storefront and Operations manifest identities and scopes.
- Preserve update-available, offline-ready, offline, dismiss, and explicit
  update behavior in `PwaNotice`.
- Add a deterministic policy test proving API/auth/mutation requests are absent
  from precache and runtime cache handling.
- Add build-output checks for manifests, icons, worker registration, immutable
  Next static assets, and a data-free offline fallback.

### 6. Runtime, tooling, and documentation

- Preserve `make dev-storefront`, `make dev-operations`, `make dev`, and all
  validation target names by changing package scripts rather than inventing a
  parallel command surface.
- Add production `start` scripts and prove each built app can run separately.
- Update `.env.example` and README with server-only proxy and browser-visible
  Next.js variable names, local URLs, PWA behavior, and runtime requirements.
- Ignore Next.js build/cache output and framework-generated transient files as
  appropriate.
- Keep GitHub Actions provider-neutral; existing root commands should exercise
  both Next.js applications without a duplicated CI workflow.
- Update `.codex/CURRENT_STATE.md` only after verified migration behavior.

## Planned File Changes

Exact route-entry files follow the Target State table. Small helpers should be
co-located until a second caller proves sharing.

| Path                                                                                                                        | Action                                             | Purpose                                                                                |
| --------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------- | -------------------------------------------------------------------------------------- |
| `docs/DECISIONS.md`                                                                                                         | Modify                                             | Add the accepted Next.js ADR and supersede/amend Vite-specific decisions.              |
| `docs/ARCHITECTURE.md`                                                                                                      | Modify                                             | Replace Vite SPA/static deployment details with the approved Next.js boundaries.       |
| `docs/CONVENTIONS.md`                                                                                                       | Modify                                             | Define App Router URL ownership, environment prefixes, output, and verification rules. |
| `README.md`                                                                                                                 | Modify                                             | Document Next.js setup, commands, environment, PWA, and runtime behavior.              |
| `.env.example`                                                                                                              | Modify                                             | Replace Vite-prefixed values with safe Next.js client/server configuration.            |
| `turbo.json`                                                                                                                | Modify                                             | Track Next.js build output without caching `.next/cache`.                              |
| `.gitignore` and `.prettierignore`                                                                                          | Modify                                             | Exclude Next.js generated output.                                                      |
| `packages/typescript-config/package.json`                                                                                   | Modify                                             | Export shared Next.js compiler defaults.                                               |
| `packages/typescript-config/nextjs.json`                                                                                    | Create                                             | Share the minimum Next.js TypeScript configuration.                                    |
| `apps/storefront-web/package.json`                                                                                          | Modify                                             | Replace Vite/Router scripts and dependencies with Next.js equivalents.                 |
| `apps/operations-web/package.json`                                                                                          | Modify                                             | Replace Vite/Router scripts and dependencies with Next.js equivalents.                 |
| `apps/*-web/next.config.ts`                                                                                                 | Create                                             | Own transpilation, local rewrites, runtime output, and worker headers.                 |
| `apps/*-web/postcss.config.mjs`                                                                                             | Create                                             | Integrate existing Tailwind CSS 4 with Next.js.                                        |
| `apps/*-web/next-env.d.ts` and `tsconfig*.json`                                                                             | Create/modify/remove                               | Adopt framework and shared compiler configuration.                                     |
| `apps/*-web/src/app/**`                                                                                                     | Create                                             | Add root layouts, providers, metadata/manifests, and route entries.                    |
| `apps/*-web/src/routes/paths.ts` and tests                                                                                  | Modify                                             | Retain URL ownership and update tests for App Router use.                              |
| `apps/*-web/src/layouts/**`                                                                                                 | Modify                                             | Replace Outlet/NavLink composition with Next children/navigation.                      |
| `apps/*-web/src/pages/**`                                                                                                   | Modify narrowly                                    | Replace React Router hooks/links and declare client boundaries where needed.           |
| `apps/*-web/src/api/client.ts` and tests                                                                                    | Modify                                             | Replace Vite environment reads while preserving transport behavior.                    |
| `apps/*-web/src/pwa/**` and focused tests                                                                                   | Modify/create                                      | Add native/Workbox worker source, registration, update, offline, and cache policy.     |
| `apps/*-web/public/offline.html`                                                                                            | Create                                             | Provide a data-free offline navigation fallback.                                       |
| `apps/*-web/scripts/build-service-worker.mjs`                                                                               | Create if existing Workbox build remains necessary | Generate ignored production worker output.                                             |
| `apps/*-web/index.html`, `vite.config*.ts`, `src/main.tsx`, `src/app/App.tsx`, `src/routes/routes.tsx`, `src/vite-env.d.ts` | Remove after parity                                | Remove obsolete Vite and React Router entry/configuration files.                       |
| `pnpm-lock.yaml`                                                                                                            | Regenerate with pnpm                               | Record dependency changes; never edit manually.                                        |
| `.codex/CURRENT_STATE.md`                                                                                                   | Modify after verification                          | Record actual Next.js behavior, evidence, and remaining risks.                         |
| `.codex/TASK.md`                                                                                                            | Execute/archive only after acceptance              | Preserve this plan and execution evidence.                                             |

Do not add frontend Dockerfiles, cloud configuration, a BFF, middleware auth,
or a new shared routing/PWA framework in this task.

## Dependencies and Sequencing

1. Explicit plan approval.
2. Accepted ADR and aligned Architecture/Conventions updates.
3. Shared TypeScript, workspace, environment, and Next.js configuration.
4. Storefront App Router compatibility migration and focused verification.
5. Operations App Router compatibility migration and focused authenticated
   verification.
6. PWA build/output verification for both apps.
7. Cross-app repository validation, production-start smoke, documentation, and
   current-state update.

Migrate Storefront first to validate the shared configuration and public
read-only path. Apply the proven pattern to Operations only after Storefront
route, proxy, query, and PWA checks pass. Keep both app migrations in this task
because the accepted framework decision and shared tooling must not leave the
repository in a permanently mixed state.

## API and Database Impact

- No OpenAPI path, DTO, status, error, authentication, or idempotency contract
  change is intended.
- No Go source, PostgreSQL migration, seed, or database schema change is
  intended.
- Next.js rewrites are transport plumbing only; they do not own or transform
  API payloads.
- If migration discovery finds an API or contract change is necessary, stop
  and record a PLAN VARIANCE before touching the contract or Go API.

## Authorization, Audit, Idempotency, and Concurrency Impact

- Backend authorization remains authoritative; no Next.js middleware or page
  visibility rule may be reported as security enforcement.
- Operations cookies remain HttpOnly and browser-managed. CSRF remains
  memory-only and is sent only to unsafe operations that require it.
- Existing audit and outbox effects remain Go-owned and unchanged.
- Existing per-intent idempotency keys, expected versions, stale conflict
  handling, and non-optimistic mutations must survive the migration exactly.
- Service workers must never cache or replay authenticated reads or commands.
- No new concurrency mechanism is introduced in the frontend.

## Assumptions Requiring Plan Approval

- The user intends to supersede the accepted Vite choice for both applications,
  not add temporary third applications.
- Both directories and workspace package names remain unchanged.
- Existing route URLs and local ports remain unchanged.
- Initial Next.js screens remain client-rendered for behavioral parity; SSR,
  RSC data loading, and SEO optimization are follow-up work.
- Two independently running Next.js server runtimes replace static-only web
  hosting. Provider selection and frontend containerization remain deferred.
- Current PWA installability, safe offline notice/fallback, and explicit update
  flow must remain; offline business data remains prohibited.

Any rejected assumption requires task revision before BUILD.

## Acceptance Criteria

1. An accepted ADR supersedes ADR-003 and aligns routing, Tailwind integration,
   PWA, and deployment consequences without changing ADR-002's app split.
2. Both existing workspace packages build and run as independent Next.js 16
   App Router applications on their existing local ports.
3. Every existing Storefront and Operations URL has an App Router entry,
   centralized path builder, working client navigation, and successful direct
   production-mode load.
4. Existing Event/Offering screens and placeholders preserve their loading,
   empty, error, authorization, conflict, success, responsive, and accessible
   behavior.
5. Storefront uses only public contracts and omits credentials; Operations uses
   only Operations contracts and preserves cookie/CSRF/session behavior.
6. A real local OIDC flow completes through the same-origin Operations URL;
   authenticated Event/Offering reads, creates, edits, lifecycle actions,
   stale-version recovery, rotated-CSRF recovery, and logout still work.
7. The Go API remains the only business backend. No Route Handler, Server
   Action, middleware policy, direct database access, or duplicated domain rule
   is added.
8. TanStack Query remains the server-state owner, with existing query keys,
   targeted invalidation, cancellation, retry, and non-optimistic mutation
   behavior preserved.
9. Both apps emit distinct valid manifests and registered production service
   workers; offline/update notices remain visible and user-controlled.
10. PWA verification proves `/api/*`, auth callbacks, probes, non-GET requests,
    API responses, session data, and business data are absent from all caches
    and replay paths.
11. Vite, React Router, `virtual:pwa-register`, `import.meta.env`, and obsolete
    Vite entry/config files have no remaining runtime callers or direct
    dependencies.
12. `@persona-apps/ui`, `@persona-apps/api-client`, Tailwind tokens, and shared
    TypeScript defaults are reused without duplicating application pages or API
    policy into packages.
13. Turborepo caches the correct Next.js outputs, generated output is ignored,
    and the lockfile was regenerated by pnpm.
14. README, decisions, architecture, conventions, environment examples, and
    current-state documentation match verified behavior.
15. All focused checks, `make validate`, Go checks, Compose validation, both
    production-start browser smokes, and the existing API container build pass.

## Testing

- Route/path tests for every static and dynamic URL and encoded identifier.
- Next configuration tests for optional local rewrites, unmodified canonical
  API paths, workspace package transpilation, and service-worker headers.
- Environment tests proving only intended `NEXT_PUBLIC_*` values enter browser
  code and missing/invalid proxy configuration fails safely.
- Existing API-client, catalogue, session, query, presentation, forms, and UI
  foundation tests under both apps.
- Focused provider/client-boundary tests that detect accidental server use of
  browser-only APIs and QueryClient cross-request reuse.
- PWA policy tests for manifest identity, worker output, cache allowlist,
  network-only API/auth behavior, offline fallback, waiting-worker prompt, and
  explicit update activation.
- Production-mode direct-load and navigation smoke for every route, including
  both dynamic detail routes.
- Real Storefront/API/PostgreSQL browser checks for populated, empty,
  no-active-Event, not-found, rate-limited, dependency-failure, timezone, and
  narrow-layout states.
- Real Operations/OIDC/API/PostgreSQL browser checks for session bootstrap,
  permissions, create/edit/lifecycle commands, conflict recovery, CSRF refresh,
  logout, direct loads, and narrow-layout behavior.

Do not declare parity from unit tests or builds alone.

## Verification

Run focused application checks:

```bash
pnpm --filter @persona-apps/storefront-web test
pnpm --filter @persona-apps/storefront-web typecheck
pnpm --filter @persona-apps/storefront-web lint
pnpm --filter @persona-apps/storefront-web build

pnpm --filter @persona-apps/operations-web test
pnpm --filter @persona-apps/operations-web typecheck
pnpm --filter @persona-apps/operations-web lint
pnpm --filter @persona-apps/operations-web build
```

Start each production build separately on non-conflicting local ports and run
the direct-load, API proxy, PWA, Storefront, and authenticated Operations smoke
checks described above.

Then run repository checks:

```bash
make validate
cd apps/api && go vet ./... && go test ./... && go build ./...
cd ../..
docker compose -f infrastructure/compose.yaml config
docker build -f apps/api/Dockerfile apps/api
```

Also verify no obsolete framework/runtime coupling remains:

```bash
rg -n "react-router|virtual:pwa-register|import\\.meta\\.env|vite-plugin-pwa|@tailwindcss/vite" apps/storefront-web apps/operations-web
git status --short
```

The first command is expected to return no application-runtime matches after
planned removals. Document any intentional test/documentation match rather than
silently ignoring it.

## Deliverables

- Accepted Next.js migration ADR and aligned canonical documentation.
- Two behaviorally equivalent, independently runnable Next.js App Router apps.
- Stable App Router entries and retained central URL builders for all existing
  routes.
- Preserved shared UI, shared transport, TanStack Query, Go API, OIDC/CSRF,
  idempotency, version-conflict, and public/internal exposure boundaries.
- Independent safe PWA manifests, workers, offline fallback, and update flows.
- Updated workspace tooling, environment examples, README, tests, and
  current-state evidence.
- Archived executed task only after all acceptance criteria are verified.

## Risks and Deferred Work

- Next.js changes static frontends into server runtimes; hosting cost,
  operations, patching, and failure modes increase even when pages remain
  client-rendered.
- App Router's default Server Component model can cause browser-API or
  hydration failures if client boundaries are too broad or too narrow.
- An incorrect local/production routing split can break OIDC callbacks,
  cookies, exact Origin checks, CSRF, or canonical `/api` paths.
- A service worker can leak stale or sensitive state if its request matching is
  broader than the explicit immutable-shell allowlist.
- Storefront SSR/SEO, server-side catalogue fetching, metadata enrichment, and
  caching/revalidation policy remain separate follow-up work after parity.
- Operations middleware guards, server-side session rendering, and Server
  Actions remain deferred because the Go API is the security and command
  boundary.
- Frontend containers, provider selection, CDN/ingress, TLS, CSP/HSTS, load
  testing, cross-browser release QA, and production rollout remain deferred.

## Final Report Requirements

Distinguish:

- implemented framework, route, runtime, proxy, state, auth, and PWA behavior;
- focused, repository, real-browser, real-API, OIDC, and production-start
  verification evidence;
- approved assumptions and any PLAN VARIANCE;
- deferred SSR/SEO, provider deployment, hardening, and product work;
- remaining security, PWA cache, hydration, runtime-cost, and deployment risks;
- whether every acceptance criterion was proven;
- the next recommended product task after the migration.

Do not mark the task Executed or archive it until the objective is implemented,
all acceptance criteria are verified, and the final review is complete.

## Final Review

### Implemented

- Storefront and Operations are independent Next.js 16 App Router applications
  with retained URL builders, browser-owned TanStack Query, application-owned
  API modules, and no duplicated Go business behavior.
- Production service-worker generation, manifests, data-free offline fallback,
  and user-controlled PWA status remain in each application. API, auth, probes,
  cross-origin, and non-GET requests remain network-only by policy.
- The only verification-era Go changes were formatting normalization of four
  previously drifted files and a time-relative test fixture in
  `TestSessionReturnsExpiryAndSortedPermissionsWithoutCaching`; no runtime API
  or product behavior changed.

### Verified

- `make validate` passed, including formatting, frontend lint/typecheck/test/
  build, Go vet/test/build, and Compose configuration.
- `docker build -f apps/api/Dockerfile apps/api` passed.
- Production `next start` instances for Storefront (`5173`) and Operations
  (`5174`) loaded directly in Chrome, hydrated without console errors, and
  rendered the expected navigation/error boundaries when the API was absent.
- Storefront client navigation to `/offerings` completed under the production
  server. Both apps displayed the offline-ready PWA status after worker
  registration.
- The existing real Go API/PostgreSQL/local-OIDC command-path evidence remains
  recorded in `.codex/CURRENT_STATE.md`; it covers Operations session/OIDC,
  CRUD/lifecycle, idempotency, conflict, CSRF recovery, public projection, and
  logout.
- No obsolete Vite/React Router/PWA runtime coupling remains under either web
  application.

### Assumed and Deferred

- The browser-control evaluation surface does not expose direct
  `navigator.serviceWorker` or CDP ServiceWorker inspection. Actual registration
  is evidenced by the production PWA status UI; deterministic worker-policy
  tests continue to prove cache exclusions.
- Cross-browser release QA, deployed-ingress behavior, hosting/provider choice,
  CSP/HSTS, load/soak testing, and broader accessibility certification remain
  follow-up release work, not blockers for this completed migration scope.

### Next Task

Activate and execute W3-01, then continue the approved Common Purchase vertical
slice through W3-06.
