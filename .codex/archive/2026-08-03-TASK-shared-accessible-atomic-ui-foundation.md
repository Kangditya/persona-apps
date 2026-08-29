# Task: Shared Accessible Atomic UI Foundation

## Executed

## Objective

Establish a shared, accessible atomic component foundation for both frontend
applications:

```text
apps/operations-web
apps/storefront-web
```

Use the following layers together with clear, non-overlapping responsibilities:

```text
Tailwind CSS       → design tokens, layout, variants, responsive styling
shadcn/ui          → editable component patterns and composition conventions
React Aria         → accessible behavior, keyboard interaction, focus, and ARIA semantics
packages/ui        → shared framework- and domain-agnostic components
applications       → feature-specific compositions and business-facing UI
```

Generic primitives, semantic tokens, and reusable atomic components must be
owned by:

```text
packages/ui
```

This task establishes a reusable UI foundation only. It must not implement final
Figma designs, redesign every route, add qurban business behavior, or move API
logic into UI components.

It also establishes an independent Progressive Web App (PWA) foundation for
each frontend. The PWA foundation must provide installability, a correctly
scoped web-app manifest, safe service-worker lifecycle behavior, and an
explicitly conservative caching policy without treating cached data as
authoritative operational or transactional truth.

---

## Required Workflow

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

Before implementation:

* inspect both frontend applications, `packages/ui`, root workspace configuration,
  Tailwind v4 setup, TypeScript aliases, Vite configuration, CSS entrypoints,
  and package-manager conventions;
* inspect existing components, styling utilities, and package exports before
  creating replacements;
* inspect shadcn/ui's current supported installation pattern for the repository's
  React, Tailwind, and monorepo versions;
* inspect the current Vite configuration, public assets, deployment assumptions,
  and browser-support constraints before selecting a PWA integration;
* determine whether an existing PWA plugin, service worker, manifest, icon set,
  cache policy, update prompt, or offline fallback already exists;
* determine whether each proposed interactive primitive should use an existing
  shadcn dependency, React Aria Components, or a documented existing equivalent;
* do not place both Radix/shadcn behavior and React Aria behavior on the same
  interactive primitive unless the plan proves their responsibilities do not
  overlap;
* identify dependency ownership, import/export strategy, Tailwind source
  scanning, Vite workspace consumption, and PWA manifest/service-worker scope;
* identify public versus operations data that must remain network-only and must
  never be persisted by the service worker;
* produce a planned file-change table and an accessibility ownership matrix;
* identify assumptions and unresolved styling or interaction requirements;
* complete the plan and wait for explicit approval before BUILD.

---

## Source of Truth

Follow:

```text
AGENTS.md
.codex/AGENTS.md
.codex/CURRENT_STATE.md
docs/ARCHITECTURE.md
docs/DECISIONS.md
```

Relevant constraints:

* Tailwind CSS v4 through the Vite integration is accepted.
* `packages/ui` is reserved for shared presentational primitives; business
  workflows and API access do not belong there.
* Storefront and Operations remain independent React applications.
* React Router owns navigation; shared UI components must not own routes.
* Existing repository conventions take precedence over generic shadcn or React
  Aria defaults.
* The Go API remains the authoritative source of business, financial, and
  operational state; a PWA cache is never transactional truth.
* Storefront and Operations have separate deployment, API, and security
  boundaries; each requires its own manifest, service-worker scope, and cache
  policy.

---

## Ownership and Dependency Direction

### Shared package

Use `packages/ui` for reusable, domain-agnostic UI source:

```text
packages/ui/
├── components.json                 # only when required by shadcn tooling
├── package.json
├── src/
│   ├── components/
│   │   ├── atoms/
│   │   ├── molecules/
│   │   └── patterns/
│   ├── lib/
│   │   ├── cn.ts
│   │   └── accessibility.ts         # only if a shared helper is justified
│   ├── styles/
│   │   └── tokens.css
│   └── index.ts
└── README.md
```

The exact layout may differ when repository conventions justify it. Do not add
empty folders or abstractions without active components that use them.

### Atomic design boundaries

| Layer | Owns | Must not own |
|---|---|---|
| Tokens | semantic colors, typography, spacing, radii, shadows, motion, focus treatment | feature names, route-specific styling, business data |
| Atoms | buttons, text inputs, labels, badges, separators, icons, skeletons | API calls, route navigation, product workflows |
| Molecules | field groups, alerts, cards, menu triggers, dialogs, sheets | page layouts, business-specific validation rules |
| Patterns | accessible data display and feedback patterns reusable across applications | application pages or qurban terminology |
| Applications | page and feature composition, route-specific presentation, business language | generic primitives duplicated from `packages/ui` |

### Allowed direction

```text
apps/operations-web  → packages/ui
apps/storefront-web  → packages/ui
```

### Forbidden direction

```text
packages/ui → apps/operations-web
packages/ui → apps/storefront-web
apps/operations-web ↔ apps/storefront-web
```

`packages/ui` must not import API clients, routes, authentication state, domain
entities, environment configuration, or qurban workflow code.

---

## Accessibility and Component Strategy

For every interactive component selected in the plan, record one owner for its
behavior and accessibility semantics:

| Component | Visual/composition owner | Behavior and ARIA owner | Notes |
|---|---|---|---|
| Button | Tailwind + shared component | native HTML / React Aria where needed | TBD during plan |
| Input and field | Tailwind + shared component | React Aria or native HTML | TBD during plan |
| Dialog | shadcn pattern + Tailwind | React Aria or a selected shadcn primitive | one owner only |
| Menu | shadcn pattern + Tailwind | React Aria or a selected shadcn primitive | one owner only |
| Sheet | shadcn pattern + Tailwind | React Aria or a selected shadcn primitive | one owner only |
| Table | Tailwind + shared component | native HTML / React Aria patterns as needed | TBD during plan |

The plan must explain the selected ownership for every component added. Do not
combine React Aria hooks/components with Radix-based shadcn behavior on the same
control merely to satisfy both libraries.

Implementation requirements:

* use semantic HTML before adding an ARIA role;
* use React Aria Components or hooks for shared controls that need accessible
  composite interaction, keyboard navigation, focus management, or dynamic ARIA
  state and are not already correctly supplied by the selected shadcn primitive;
* use shadcn/ui as editable source patterns, not an opaque design-system
  dependency;
* use Tailwind utility classes and semantic tokens for styling rather than
  unstructured inline styles or route-specific shared CSS;
* maintain visible `:focus-visible` treatment, pointer and keyboard support,
  disabled behavior, accessible names, validation descriptions, and error
  announcements;
* respect `prefers-reduced-motion` for nonessential animation;
* avoid color-only meaning and preserve sufficient contrast;
* provide component APIs that keep accessible labels and descriptions explicit;
* avoid introducing a global accessibility provider unless the selected React
  Aria APIs require it and the plan documents why.

---

## PWA Foundation Requirements

Create an installable PWA foundation for both applications while preserving
their independent identities and security boundaries:

```text
apps/operations-web
apps/storefront-web
```

The implementation plan must document:

```text
Selected Vite PWA/service-worker integration and version compatibility
Manifest identity, names, theme colors, display mode, start URL, and scope
Icon source, required sizes, maskable icon behavior, and ownership
Service-worker registration and lifecycle strategy
Update detection, user-visible update behavior, and activation strategy
Static asset precache policy
Runtime caching policy by resource and API sensitivity
Offline fallback behavior
Development versus production service-worker behavior
Deployment path/base-URL assumptions
Browser support and verification procedure
```

Implementation requirements:

* use a supported Vite-compatible PWA integration or an explicitly justified
  repository-native alternative;
* create separate manifests, icons, service-worker registrations, names, and
  scopes for Storefront and Operations;
* precache only immutable build assets and a small application shell required to
  start the application;
* treat `/api/**`, authenticated content, operational records, financial data,
  participant data, tokens, cookies, request bodies, and error responses as
  network-only and do not persist them in a service-worker cache;
* do not implement offline command queues, background sync, mutation replay,
  payment submission, privileged operations, or offline authorization;
* provide a clear offline/unavailable state rather than displaying stale cached
  records as current data;
* provide a visible, accessible update prompt or documented update mechanism;
* avoid unconditional `skipWaiting` or automatic reload behavior that can
  interrupt an in-progress form, command, payment, or privileged workflow;
* ensure the PWA shell, manifest, and update UI use the shared accessible atomic
  components and Tailwind tokens where appropriate;
* configure development behavior so service workers do not create confusing
  stale local assets unless explicit PWA-development testing is enabled;
* do not add a cache strategy for an endpoint without a documented freshness,
  privacy, authorization, and invalidation rationale.

The initial PWA scope is installability and a safe offline application shell.
Offline business workflows remain deferred until an approved architecture and
domain design define synchronization, conflict handling, authorization, audit,
and idempotency behavior.

---

## Scope

### In scope

* create or adapt `packages/ui` as the shared component package;
* configure stable workspace imports and package exports for both Vite apps;
* configure shadcn tooling only where it supports the selected shared source
  structure and repository conventions;
* add `react-aria-components` or the narrowest supported React Aria dependency
  only when it is needed by selected accessible primitives;
* configure Tailwind v4 source scanning and global CSS only as needed for shared
  source files;
* add the narrowest Vite-compatible PWA dependency and configuration needed for
  separate Storefront and Operations PWA foundations;
* add separate web-app manifests, icons, service-worker registration, static
  asset precaching, and accessible update/offline UI for each application;
* define a minimal semantic token layer for color, typography, spacing, radius,
  shadow, focus, and motion;
* add a `cn` utility using an existing equivalent or `clsx` and `tailwind-merge`;
* implement a minimum shared set of atomic components:

  ```text
  Button
  Input
  Label / Field
  Textarea
  Badge
  Separator
  Skeleton
  Alert
  Card
  ```

* implement selected accessible molecular components after their behavior owner
  is approved:

  ```text
  Dialog
  DropdownMenu or Menu
  Sheet or Drawer
  Form field group with validation feedback
  ```

* add a reusable table or data-display pattern only if the plan identifies an
  existing placeholder route that can exercise it;
* integrate the smallest non-business showcase or existing placeholder route in
  each application to prove shared imports, tokens, and interactive behavior;
* document the process for adding a future atom, molecule, shadcn pattern, or
  React Aria-backed component;
* add focused tests using the existing Vitest tooling.

### Out of scope

* final page designs or Figma-driven visual implementation;
* complete dashboards, catalogue, checkout, payment, livestock, allocation, or
  reporting screens;
* API, backend, OpenAPI, authentication, or domain changes;
* caching API responses, authenticated content, operational records, financial
  data, participant data, or other sensitive/transactional data;
* offline mutation queues, background synchronization, payment flows, privileged
  operations, or offline authorization;
* Storybook, visual-regression infrastructure, or a separate documentation site
  unless already present and explicitly approved;
* duplicate primitive sets in either application;
* applying both React Aria and a shadcn/Radix behavior implementation to the
  same interactive component without an approved, non-overlapping rationale;
* speculative themes, white-label architecture, or multiple brands;
* broad refactoring of unrelated application components;
* committing or pushing changes.

---

## Package and Styling Requirements

The implementation plan must document:

```text
Shared package ownership
Atomic component classification
Public package exports and import paths
TypeScript aliases and Vite workspace consumption
Tailwind v4 source scanning and CSS entrypoints
Token ownership and light/dark-mode behavior
shadcn component-generation procedure
React Aria dependency and behavior ownership matrix
Focus, keyboard, form, overlay, and screen-reader behavior
PWA manifests, service-worker scopes, caching rules, update behavior, and offline fallback
Dependency placement and lockfile impact
```

The implementation must:

* keep shadcn-derived source editable inside `packages/ui`;
* expose components through stable workspace imports, not deep relative paths;
* use explicit exports when supported by the existing package setup;
* preserve or improve existing application branding without inventing a final
  brand system;
* avoid hard-coded application, route, or business terminology in shared
  primitives;
* include shared source files in Tailwind scanning or use the approved Tailwind
  v4 equivalent;
* ensure tokens are loaded exactly once per application entrypoint;
* keep application-specific composition inside the owning app;
* install dependencies at the narrowest valid workspace scope;
* not install every shadcn, Radix, or React Aria component preemptively.
* keep the PWA configuration application-owned; `packages/ui` may supply the
  presentational update/offline components but must not register service workers
  or own cache policy.

---

## Minimum Acceptance Criteria

1. `packages/ui` owns the selected generic atoms and molecules.
2. Both applications build while importing shared components through stable
   workspace package paths.
3. Neither application has a duplicate shared `components/ui` primitive set.
4. Every interactive primitive has a documented single behavior/ARIA owner.
5. React Aria is used where selected for behavior and accessibility, without
   duplicate Radix/shadcn interaction logic on the same primitive.
6. Tailwind v4 applies classes from `packages/ui` in development and production
   builds.
7. Shared semantic tokens render correctly in each consuming application.
8. The minimum atomic set renders with supported variants and disabled, invalid,
   loading, and focus-visible states where applicable.
9. Selected dialogs, menus, and sheets provide correct keyboard interaction,
   Escape behavior, focus handling, and accessible naming.
10. Existing placeholder routes continue to compile and render.
11. One small integration point in each application proves shared component
    consumption without introducing business workflows.
12. `packages/ui` contains no API access, routing, authentication, or business
    rules.
13. No final product data, visual specification, or unsupported design decision
    is invented.
14. The component-addition procedure is documented and repeatable.
15. Storefront and Operations each expose a valid, separately scoped web-app
    manifest and installable production build.
16. Each application precaches only approved immutable assets and the application
    shell; API and sensitive data remain network-only.
17. Each application provides an accessible update and offline/unavailable state
    without automatic interruption of user work.

---

## Testing

Add focused tests supported by the existing repository tooling.

### Shared component tests

* class merging, variants, and forwarded refs where applicable;
* semantic elements and accessible names;
* disabled, invalid, required, and loading states;
* focus-visible behavior and keyboard activation;
* React Aria-backed field descriptions and validation announcements;
* overlays open, close, handle Escape, and restore focus according to the
  selected behavior owner;
* stable shared exports resolve without deep source imports;
* reduced-motion behavior when the component includes animation.

### Application integration tests

* Operations Web renders a component imported from `packages/ui`;
* Storefront Web renders a component imported from `packages/ui`;
* shared Tailwind tokens and classes resolve;
* application compositions do not create application-to-application imports;
* no browser console errors result from duplicate interactive behavior.

### PWA foundation tests

* each production build emits the expected manifest and service-worker assets;
* each manifest has the correct application identity, scope, icons, and start
  URL;
* the service-worker configuration precaches approved static assets only;
* `/api/**` and authenticated or sensitive responses are excluded from runtime
  caching;
* update and offline UI are accessible and render without business data;
* development configuration does not leave an unintended persistent service
  worker active.

Use existing Vitest tooling. Do not add a test framework solely for this task.

---

## Verification

Run applicable repository checks:

```bash
make validate
pnpm run lint
pnpm run typecheck
pnpm run test
pnpm run build
```

Also run focused package and application checks when workspace scripts support
them.

Manually verify:

1. both applications start with documented local setup;
2. shared atoms render in both applications;
3. selected dialogs, menus, and sheets work with keyboard-only navigation;
4. focus is visible, managed correctly, and restored after overlays close;
5. screen-reader labels, descriptions, and validation feedback are available for
   the selected field components;
6. light/dark token behavior matches the support implemented by the plan;
7. production builds contain no unresolved aliases, package exports, or Tailwind
   classes;
8. `packages/ui` has no direct imports from applications or domain packages;
9. no primitive combines competing React Aria and shadcn/Radix behavior without
   the approved ownership rationale.
10. each production application build passes a browser PWA audit or equivalent
    manifest/service-worker inspection;
11. installation creates the correct Storefront or Operations application
    identity without cross-application scope conflicts;
12. offline mode renders only the approved application shell and unavailable
    state; it does not expose stale or sensitive API data;
13. a service-worker update can be detected and applied through the documented
    user-safe update flow.

If a command cannot run, report the exact command and blocker; do not mark it
as passed.

---

## Deliverables

The final report must include:

* selected shadcn patterns and React Aria dependencies;
* atomic component inventory and ownership matrix;
* package exports, aliases, and Tailwind/token integration;
* PWA integration, manifest identities, icons, service-worker scopes, cache
  policy, update behavior, and offline fallback;
* application integration points;
* dependencies and configuration changes;
* files changed;
* verification and accessibility results;
* plan variances;
* deferred components and styling decisions;
* recommended next UI implementation task.

---

## Constraints

* Keep changes limited to the shared accessible atomic UI foundation and minimal
  application integration.
* Reuse repository conventions before adopting shadcn or React Aria defaults.
* Keep generic primitives in `packages/ui` and business composition in the
  owning application.
* Treat Tailwind as the styling and token mechanism, shadcn as editable
  component patterns, and React Aria as selected accessible behavior—not three
  competing implementations of the same control.
* Treat the PWA as a static-shell and installability foundation, not an offline
  source of business or operational truth.
* Never cache or replay authenticated, financial, participant, operational, or
  mutation data without an explicit future architecture decision.
* Do not duplicate shared primitives in the applications.
* Do not implement final product screens or backend behavior.
* Do not add dependencies without an approved plan and documented reason.
* Do not commit or push changes.
