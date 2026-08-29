# Task: Implement Radix and Phosphor Icon System
## Executed

## Objective

Introduce a consistent icon system for the frontend monorepo using:

```text
@radix-ui/react-icons
@phosphor-icons/react
```

Use Radix Icons for compact interface controls and Phosphor Icons for expressive, status, navigation, and qurban-domain visuals.

The first implementation target is the simplified mobile purchase-tracking page in:

```text
apps/storefront-web
```

The implementation must preserve the existing React Aria, shadcn, Tailwind CSS, React Router, and monorepo conventions.

---

## Required Workflow

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

Before coding:

- inspect `apps/storefront-web`, `apps/operations-web`, and existing shared UI packages;
- identify all currently installed icon libraries and direct SVG/icon implementations;
- inspect existing shadcn and React Aria component conventions;
- determine whether icon dependencies belong in a shared UI package or application package;
- produce a planned file-change table;
- do not begin BUILD until the plan is approved.

---

## Icon Ownership Rules

### Radix Icons

Use for small utility and control icons, including:

- chevrons;
- close and menu actions;
- copy actions;
- external links;
- overflow menus;
- compact form and disclosure controls.

### Phosphor Icons

Use for semantic and expressive icons, including:

- qurban and livestock concepts;
- location and schedule information;
- QR and tracking states;
- participant and transaction information;
- status, success, warning, and help visuals;
- navigation icons where a stronger visual symbol is useful.

Do not mix both libraries for equivalent icons within the same feature.

---

## Scope

### In scope

- install the selected icon packages in the correct workspace location;
- establish a small shared icon convention or wrapper only when justified by existing package boundaries;
- standardize icon sizing, stroke/weight, alignment, and decorative color behavior with Tailwind classes;
- migrate the Storefront purchase-tracking page from text symbols, inline SVGs, or inconsistent icons;
- apply icons to the simplified tracking information hierarchy:
  - transaction reference and copy action;
  - QR access;
  - customer information;
  - venue and Maps action;
  - order information;
  - animal code and slaughter schedule;
  - track action;
  - collapsible secondary details;
- preserve React Aria interaction and accessibility behavior;
- add focused tests for icon-only controls and accessible labels;
- document the icon-selection convention for future contributors.

### Out of scope

- broad redesign of unrelated pages;
- replacement of React Aria or shadcn components;
- custom illustration work;
- changes to backend APIs or business logic;
- introducing a third icon library;
- migrating every icon in the monorepo unless explicitly approved in the plan.

---

## Implementation Requirements

- Import icons directly from their package entry points unless repository conventions require a shared export.
- Avoid dynamic icon lookup maps unless runtime selection is required.
- Keep bundle impact limited through tree-shakeable named imports.
- Use Tailwind sizing such as:

```text
size-3.5
size-4
size-5
size-6
```

- Prefer Phosphor `regular` weight for normal content and `bold` only for emphasized states.
- Decorative icons must use:

```tsx
aria-hidden="true"
```

- Icon-only interactive controls must have an accessible name through React Aria or `aria-label`.
- Do not use icon color as the only status indicator.
- Preserve visible text for primary actions such as:

```text
Track
Lihat di Maps
Salin
Perbesar QR
```

- Do not place raw package imports throughout feature code when an existing shared UI abstraction already owns equivalent controls.
- Remove replaced icon dependencies only when no remaining usages exist.

---

## Tracking Page UX Constraints

The page must remain a simple tracking experience, not become a dashboard.

Use icons to improve scanning, not to add decoration. Prioritize:

1. transaction identity;
2. QR access;
3. customer information;
4. venue and Maps action;
5. order summary;
6. qurban item, animal code, slaughter schedule, and Track action.

Replace wide mobile tables with stacked semantic rows or cards where required by the existing redesign task. Optional or low-priority details should remain collapsible.

---

## Acceptance Criteria

1. Radix Icons and Phosphor Icons are installed in the correct workspace package.
2. Their responsibilities are clearly separated and documented.
3. The tracking page no longer uses inconsistent text glyphs or ad hoc SVG icons for migrated elements.
4. All icon-only controls have accessible names.
5. Decorative icons are hidden from assistive technology.
6. Icons align consistently with adjacent text and controls.
7. Primary actions remain understandable without relying on the icon alone.
8. Existing React Aria keyboard and focus behavior remains intact.
9. No unnecessary third icon library is introduced.
10. Existing routes continue to compile and render.

---

## Verification

Run applicable repository checks:

```bash
pnpm run lint
pnpm run typecheck
pnpm run test
pnpm run build
```

Also verify manually:

- tracking page at mobile widths;
- keyboard operation for copy, Track, Maps, QR enlargement, and disclosures;
- accessible names in the browser accessibility tree;
- no layout shift caused by icon sizing;
- no duplicate or unused icon dependency remains after migration.

If a command cannot run, report the exact command and blocker.

---

## Discovery

### Current `packages/ui` state

`@persona-apps/ui` is already the shared source for framework- and
domain-agnostic primitives. It currently contains:

- atoms: `Alert`, `Badge`, `Button`, `Card`, `Input`, `Label`, `Separator`,
  `Skeleton`, and `Textarea`;
- molecules: `Dialog`, `DropdownMenu`, `Field`, and `Sheet`;
- one pattern: `PwaNotice`;
- shared `cn` utility;
- shared Tailwind v4 semantic tokens in `src/styles/tokens.css`;
- root exports from `src/index.ts`;
- no package-level component tests;
- no icon dependency, icon wrapper, or icon-specific token convention.

Existing behavior ownership is consistent with the architecture: native HTML
owns simple controls and React Aria Components owns composite keyboard/focus
behavior. The package README explicitly says applications import from the
package root and that shared UI must remain application- and domain-neutral.

### Consumers and current gap

Both frontend applications depend on `@persona-apps/ui`, but current consumers
only exercise the existing primitives. The Storefront purchase-tracking route
is still a placeholder at `apps/storefront-web/src/pages/PurchaseTrackingPage.tsx`;
there are no existing text glyphs, inline SVGs, or migrated tracking controls
to replace yet.

The icon task therefore has two distinct boundaries:

1. establish the shared icon dependency/usage convention in `packages/ui`;
2. migrate the tracking UI only when the tracking page implementation exists.

The current workspace has no Radix or Phosphor package installed and no third
icon library to remove.

### Constraints confirmed

- `docs/DECISIONS.md` ADR-005 requires minimal Tailwind usage and defers a large
  token system until repeated decisions stabilize;
- `docs/ARCHITECTURE.md` requires shared UI to contain stable primitives, not
  application pages or business rules;
- the active task forbids implementation before plan approval;
- no API, database, authorization, audit, concurrency, or idempotency impact is
  expected for this UI-only task.

## Plan

### Proposed approach

Use the smallest package-boundary change that supports the stated ownership:

1. add `@radix-ui/react-icons` and `@phosphor-icons/react` to
   `packages/ui/package.json`;
2. add a small root-exported icon convention only if it prevents repeated
   package imports or centralizes shared sizing/accessibility behavior;
3. document the Radix-versus-Phosphor selection rule in `packages/ui/README.md`;
4. keep application-specific icon selection in the application until the
   tracking page is no longer a placeholder;
5. add focused application tests only for the actual icon-only controls after
   the tracking UI is implemented.

No broad icon registry, dynamic lookup map, custom SVG abstraction, or
application-page migration is planned at this stage.

### Planned file changes

| File | Action | Purpose |
|---|---|---|
| `packages/ui/package.json` | modify | Add the two approved icon dependencies if package ownership is confirmed during BUILD. |
| `packages/ui/src/index.ts` | modify | Export a shared icon convention only if needed by existing package boundaries. |
| `packages/ui/src/components/atoms/Icon.tsx` | create, conditional | Minimal shared icon sizing/accessibility helper only if direct imports would duplicate stable behavior. |
| `packages/ui/README.md` | modify | Document icon ownership, sizing, decorative behavior, and accessible-name rules. |
| `apps/storefront-web/src/pages/PurchaseTrackingPage.tsx` | modify, deferred | Migrate actual tracking controls after the placeholder is replaced by the approved tracking UI. |
| `apps/storefront-web/src/pages/PurchaseTrackingPage.test.tsx` | create, conditional | Test accessible names and keyboard behavior for real icon-only tracking controls. |
| `pnpm-lock.yaml` | modify, generated | Update only through the package manager when dependencies are added. |

### Verification

Before BUILD approval, no implementation commands are run. After approval,
run the applicable focused checks first, then the repository checks:

```bash
pnpm --filter @persona-apps/ui typecheck
pnpm --filter @persona-apps/storefront-web test
pnpm run lint
pnpm run typecheck
pnpm run test
pnpm run build
```

Manual checks remain limited to the implemented surface: keyboard/focus
behavior, accessible names, decorative `aria-hidden`, mobile sizing, and
bundle-safe named imports.

### Risks and open decisions

- The current tracking page is a placeholder, so page migration cannot be
  honestly completed in this task without inventing product UI.
- A shared `Icon` wrapper is not yet justified by current consumers; direct
  named imports may be the lower-complexity choice unless the first real page
  exposes repeated behavior.
- Exact icon choices should wait for the approved tracking composition and
  should not introduce qurban-domain vocabulary into `packages/ui`.

BUILD remains blocked until this plan is explicitly approved.

---

## Final Report

- dependencies added: `@radix-ui/react-icons`, `@phosphor-icons/react` in `packages/ui`; no dependencies removed.
- icon ownership documented in `packages/ui/README.md`; no wrapper or registry added.
- tracking page migration deferred because `PurchaseTrackingPage` remains a placeholder.
- files changed: `packages/ui/package.json`, `packages/ui/README.md`, `pnpm-lock.yaml`; task archived after verification.
- accessibility verification: convention documented; no new icon controls exist to test.
- commands passed: `pnpm run lint`, `pnpm run typecheck`, `pnpm run test`, `pnpm run build`; both app dev servers started successfully on ports 5173 and 5174.
- plan variance: none.
- deferred: actual tracking-page icon migration and focused icon-control tests until the page exists.

Provide:

- dependencies added or removed;
- icon ownership decision;
- migrated components and pages;
- files changed;
- accessibility verification;
- commands executed and results;
- plan variances;
- deferred icon migrations.

Do not commit or push changes.
