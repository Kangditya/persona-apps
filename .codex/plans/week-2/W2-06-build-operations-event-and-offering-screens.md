# Task: W2-06 Build Operations Event and Offering Screens

## Status

Draft — awaiting review. The discovery defaults were confirmed on 2026-08-13,
but this task is not Plan Approved and must not enter BUILD until this file is
reviewed and explicitly approved.

## Tracker

- Week: Week 2
- Epic: Event & Offering
- Week goal: Allow Operations to configure commerce and Storefront to discover
  it.
- Milestone: Public Event and Offering discovery.
- Weekly allocation: 42 hours shared across W2-01 through W2-08; the tracker
  does not assign per-task estimates, so this plan does not invent them.
- Tracker objective: Build Operations Event and Offering screens.
- Dependency: W2-05 contract-aligned Operations API and authenticated session.

## Objective

Replace the relevant Operations placeholders with a thin, accessible, real-API
workflow for authorized operators to list, create, inspect, edit, and transition
Events and their Offerings without duplicating business rules in React.

## Context

The Operations React application already has centralized route registries,
React Router, TanStack Query, a shared API client, and accessible shared UI
primitives. It currently exposes only placeholder pages and diagnostics. No
business API module, authenticated-session bootstrap, Event/Offering query, or
mutation screen exists.

Week 2 needs a functional administration path, not the full Week 5 Operations
experience. The smallest complete path is an Event list/create page, Event
detail/edit/lifecycle page with its Offering list/create flow, and Offering
detail/edit/lifecycle page.

## Source of Truth

- `docs/PRD.md` Operations actors and Event/Offering behavior;
- `docs/PRODUCT_MAP.md` Operations application ownership and roadmap;
- `docs/ARCHITECTURE.md` frontend boundaries, auth, API, and deployment rules;
- `docs/DECISIONS.md` ADR-002, ADR-006, ADR-009, ADR-010, ADR-018,
  ADR-019, ADR-020, ADR-038, ADR-039, ADR-041, ADR-042, and ADR-046;
- `docs/CONVENTIONS.md` React, routes, query, forms, accessibility, money, and
  API conventions;
- `docs/security/PERMISSIONS.md` Event/Offering permission mapping;
- `docs/security/AUTHENTICATION.md` cookie session and in-memory CSRF model;
- `contracts/openapi/operations.yaml` after W2-05;
- current Operations routes, API client, query client, layout, pages, and
  shared `@persona-apps/ui` primitives.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate after W2-05 is implemented and contract-tested. Inspect the exact
response/error shapes again before coding. Do not use mocks to declare the task
complete. Do not commit or push unless asked.

## Scope

### In Scope

- A session bootstrap query using the existing cookie-backed
  `/api/operations/v1/auth/session` endpoint with credentials included.
- A sign-in state linking to the contracted login endpoint with a validated
  relative `return_to`; unauthenticated and forbidden states remain distinct.
- A visible logout control for authenticated operators using the existing
  CSRF-protected logout endpoint, followed by removal of private query data.
- Centralized Operations paths for Event list, Event detail, and Offering
  detail, including typed path builders for dynamic IDs.
- Event list with bounded cursor pagination and clear loading, empty, error,
  unauthorized, forbidden, and success states.
- Event create and edit forms for only the existing Week 2 fields.
- Explicit Event lifecycle actions shown only when locally plausible and
  enabled only with `event.manage`; the server remains authoritative.
- Event-scoped Offering list/create flow and Offering detail/edit/lifecycle
  actions with matching permission-aware UX.
- Version/expected-version handling, conflict recovery, mutation pending state,
  explicit confirmation for consequential lifecycle actions, and query
  invalidation after success.
- Stable per-intent idempotency keys for create/lifecycle retries.
- Browser-local datetime entry converted to UTC API instants with the active
  browser timezone named next to the controls.
- Exact integer handling for price/capacity/quota without floating-point form
  arithmetic.
- Accessible labels, descriptions, validation summaries, focus behavior,
  status announcements, keyboard operation, and baseline responsive layout.

### Out of Scope

- A new auth provider, token storage, role editor, operator provisioning,
  server authorization in React, or offline authentication.
- Full dashboard analytics, Purchase/payment screens, bulk editing, import,
  export, filters beyond the contracted API, or server-side total counts.
- Drag-and-drop, custom date/time pickers, rich text editors, optimistic
  lifecycle changes, autosave, background polling, or speculative shared form
  frameworks.
- Full visual-system polish, exhaustive browser automation, or advanced
  accessibility audit; those remain Week 5 and Week 7 work.
- Generated clients or runtime schema libraries.

## Existing State

- `src/routes/paths.ts` and `routes.tsx` are the required centralized route
  owners.
- `EventDashboardPage` is a placeholder for later dashboard behavior; this
  task should add a distinct `/events` administration path rather than
  redefining dashboard projection semantics.
- The shared API client already supports typed requests, timeouts,
  cancellation, JSON/error normalization, and injectable `fetch` tests.
- TanStack Query is installed and mounted; no Event/Offering query keys exist.
- The default client base is `/api`, while W2-04 must prove that canonical
  `/api/...` endpoint paths are not double-prefixed or stripped.
- Shared Button, Input, Textarea, Field, Alert, Badge, Card, Dialog, Skeleton,
  and navigation primitives are available.
- Current Vitest tests are mostly pure/structural and no DOM testing framework
  is a direct dependency.

## Target State

- Operators reach `/events`, inspect a bounded Event list, and follow explicit
  links to `/events/:eventId` and
  `/events/:eventId/offerings/:offeringId` through centralized builders.
- Session state is loaded once through TanStack Query. CSRF is held only in
  query memory and passed explicitly to unsafe API calls; it is never written
  to local/session storage, URLs, logs, analytics, service workers, or errors.
- Each API resource has stable hierarchical query keys. Successful mutations
  invalidate only the affected detail/list/publication-related Operations
  keys; broad cache clearing is avoided.
- Mutations are pessimistic: disable duplicate submission, wait for the server,
  render the returned version, then invalidate/refetch. No contested state is
  changed optimistically.
- `409 stale_version` prompts a reload and preserves the user's form values for
  manual reconciliation; it does not silently overwrite or auto-merge.
- A create/lifecycle submission generates one `crypto.randomUUID()` key for
  that user intent. An explicit retry of an uncertain identical request reuses
  the key; changing payload/action creates a new intent/key.
- Permission checks improve navigation and controls but never replace backend
  enforcement. Direct route access still calls the API and handles `403`.

## Constraints

- Reuse React Router, TanStack Query, shared API client, shared UI primitives,
  browser APIs, and native form controls. Add no form, date, state, validation,
  money, ID, or notification dependency.
- Use `AbortSignal` supplied by TanStack Query for reads and the existing API
  client's cancellation support.
- Keep query keys serializable and deterministic; do not include freshly
  allocated objects/functions.
- Do not duplicate lifecycle/availability decisions as editable frontend
  truth. UI transition hints are derived from returned status solely to avoid
  obviously invalid buttons; server errors remain authoritative.
- Treat all form values as strings until validated. Convert exact integers
  only after per-field digit/range checks: capacity uses the contract's
  `2_147_483_647` maximum, while bigint money/quota/version values use
  `9_007_199_254_740_991`.
- Use `datetime-local` only as a browser-local editor. Convert valid values to
  RFC 3339 UTC, reject component round-trip changes such as nonexistent local
  times, and clearly display both the resolved browser timezone and computed
  UTC preview before submission.
- Never inject server text as HTML or expose raw backend error details.

## Implementation Requirements

### API and query layer

- Extend the Operations API module with explicit session, Event, and Offering
  request/response types matching the reviewed contract. Keep endpoint strings
  in this module, not scattered through pages.
- Add narrow runtime guards for required envelope/resource fields before pages
  use a response; a malformed successful payload becomes a safe API error, not
  an unchecked TypeScript cast or page crash.
- Use canonical `/api/operations/v1/...` paths under the base-path convention
  verified by W2-04/W2-05.
- Send `credentials: "include"` for session/Operations requests and attach
  `X-CSRF-Token`, `Origin` via the browser, and `Idempotency-Key` only where the
  contract requires them.
- Keep query hooks small: session, Event list/detail, Event Offerings list, and
  Offering detail. Use stable factories such as resource/list/detail key
  segments rather than a generic application cache abstraction.
- Default read retries must be bounded and disabled for `401`, `403`, `404`,
  and validation errors. Mutation retries are explicit only and retain the
  same input/key.

### Session and access states

- Load session before protected data. Render a non-destructive sign-in link to
  the existing login endpoint when unauthenticated; preserve a safe relative
  return path through the existing contract.
- Because fetching the session rotates CSRF, disable automatic focus/polling
  refetch for that query. Refresh it explicitly after an authentication/CSRF
  failure and require the operator to retry the mutation; never automatically
  replay an unsafe request with a newly rotated token.
- Show a clear forbidden state when the session exists but lacks read access.
- Hide or disable mutation controls without manage permission, with an
  explanation; never assume hidden controls secure the endpoint.
- On expired session, clear only relevant query state and show sign-in rather
  than entering a retry loop.
- Logout sends the current CSRF token without an idempotency key, clears the
  session and private Event/Offering query entries after success, and returns
  focus/navigation to the sign-in state.

### Event screens

- Add `/events` as the Event administration list/create route while preserving
  the existing `/event-dashboard` placeholder for later projection work.
- List event year, name, status, registration window summary, quota, and an
  accessible detail link. Cursor navigation uses server-provided cursors.
- Create only `event_year`, `name`, optional registration bounds, and optional
  participant quota; server creates `DRAFT` and version.
- Detail displays status/version/timestamps and includes editable current
  configuration plus explicit lifecycle actions allowed by the reviewed
  matrix.
- Require confirmation for activate, close, and archive; publish/suspend may
  use a clear inline action unless review asks for confirmation everywhere.
- Patch sends only changed fields plus current expected version; clearing an
  optional value sends explicit `null`, not omission.

### Offering screens

- Event detail owns the event-scoped Offering list and create form.
- Create only code, name, bounded descriptive kind, optional description,
  exact price minor units, currency, participant capacity, and optional total
  quota. Do not add species/package behavior absent from the contract.
- Offering detail displays availability as advisory and distinguishes finite
  zero from unbounded `null`.
- Edit only fields permitted by W2-02 and returned status. Immutable fields are
  rendered as text, not editable disabled inputs that imply future support.
- Lifecycle controls publish, mark unavailable, republish through publish, and
  archive. Archive requires confirmation.

### Forms, errors, and accessibility

- Use semantic forms, one visible label per control, descriptions for units and
  timezone, field-level errors, and a focusable error summary after failed
  client/server validation.
- Preserve entered values on validation, network, `409`, and `422` errors.
- Map only recognized server field-detail keys to known controls; unknown
  details remain in the safe summary and are never injected or dumped into the
  page.
- Announce pending/success/error state with appropriate live regions without
  duplicating every message to screen readers.
- Move focus to the page heading after navigation and to confirmation/error
  context after consequential actions where shared primitives do not already
  manage it.
- Disable a submitted action while pending, but do not disable the whole page
  or remove navigability.

## Planned File Changes

File splitting is a planning boundary; co-locate small types/helpers when that
keeps the implementation shorter.

| File | Action | Purpose |
| --- | --- | --- |
| `apps/operations-web/src/routes/paths.ts` | Modify | Add centralized Event/Offering patterns and builders. |
| `apps/operations-web/src/routes/routes.tsx` | Modify | Register Event list/detail and Offering detail routes. |
| `apps/operations-web/src/layouts/OperationsLayout.tsx` | Modify | Add permission-neutral Events navigation. |
| `apps/operations-web/src/api/client.ts` | Modify | Apply the verified canonical path/credential convention. |
| `apps/operations-web/src/api/session.ts` | Create | Session query and exact in-memory CSRF/permission shape. |
| `apps/operations-web/src/api/events.ts` | Create | Explicit Event/Offering DTOs and endpoint functions. |
| `apps/operations-web/src/api/events.test.ts` | Create | Verify URLs, headers, payloads, aborts, and error mapping with injected fetch. |
| `apps/operations-web/src/features/events/queries.ts` | Create | Stable query keys/hooks and bounded invalidation. |
| `apps/operations-web/src/features/events/forms.ts` | Create only if shared by multiple pages | Minimal integer/datetime conversion and validation helpers. |
| `apps/operations-web/src/pages/EventsPage.tsx` | Create | Event list, cursor navigation, and create flow. |
| `apps/operations-web/src/pages/EventDetailPage.tsx` | Create | Event edit/lifecycle plus Offering list/create. |
| `apps/operations-web/src/pages/OfferingDetailPage.tsx` | Create | Offering details, edit, availability, and lifecycle. |
| `apps/operations-web/src/pages/OperatorLoginPage.tsx` | Modify | Replace showcase copy with real session/sign-in state if still the owning entry page. |
| `apps/operations-web/src/routes/paths.test.ts` and focused page/helper tests | Modify/create | Verify builders, semantic states, conversion boundaries, and status/action mapping. |
| `apps/operations-web/src/styles/global.css` | Modify only if needed | Minimal responsive/form states not expressible with existing utilities. |
| `.codex/CURRENT_STATE.md` and `.codex/TASK.md` | Modify/create/archive when executed | Record verified screen behavior. |

Do not add a frontend-wide auth store, form framework, generated client, or
duplicate UI component package for these pages.

## Dependencies and Sequencing

- Requires W2-05's final contract, auth/session conformance, and live routes.
- W2-04/W2-05 must establish one base-path convention before endpoint modules
  are written.
- Can be implemented alongside W2-07 after API stability, with shared API
  client changes coordinated rather than copied.
- W2-08 closes remaining route/auth/UI regression gaps; this task still owns
  focused URL/header/form/query tests and a real-API smoke check.

## API and Database Impact

- No API route or migration is introduced by the screen task.
- Any contract/runtime mismatch found here is fixed in the owning W2-05 layer,
  not worked around in React.
- Frontend cache is a view cache only and never changes database or lifecycle
  authority.

## Authorization, CSRF, Idempotency, and Concurrency

- Cookies remain `HttpOnly`; JavaScript holds only contract-returned session
  metadata and rotating CSRF in memory.
- UI permissions control affordances only. All requests tolerate authoritative
  `401`/`403` responses.
- Create/lifecycle retry controls reuse the same key only for the same frozen
  intent. Form edits or a new action create a new key.
- Patches and lifecycle commands send the currently displayed version. A stale
  response never auto-retries as a new write.
- No optimistic update is used for publication, activation, quotas, price, or
  availability.

## Acceptance Criteria

1. Authorized operators can list/create Events, open Event detail, edit valid
   configuration, and execute all accepted lifecycle actions through real API
   calls.
2. Authorized operators can list/create Offerings for an Event, inspect/edit
   permitted fields, and execute publication/unavailability/archive actions.
3. Routes and dynamic links come exclusively from the central registries and
   survive direct-load navigation under the SPA fallback rules.
4. Session, unauthenticated, forbidden, loading, empty, validation, network,
   stale-conflict, and success states are distinct and accessible.
5. CSRF is memory-only; cookies/credentials, CSRF, expected versions, and
   idempotency headers are sent only as required by the reviewed contract.
6. Authenticated operators can log out through the existing CSRF-protected
   endpoint, after which private query data is removed and protected screens
   return to sign-in state.
7. A repeated uncertain create/lifecycle retry retains its original key and
   payload; a changed intent gets a new key.
8. Integer and datetime inputs round-trip without silent floating-point or
   timezone reinterpretation; the browser timezone is visible.
9. Mutation success uses targeted invalidation/refetch and returned versions;
   no contested business state is optimistically changed.
10. Baseline keyboard, focus, label, live-status, contrast-token, and responsive
   behavior is verified without introducing new UI/framework dependencies.
11. No mock provider, fake business data, local token storage, duplicated
    domain rule, or unfinished navigation control is shipped as complete.

## Testing

- Pure tests for path builders, query keys, integer boundary parsing,
  datetime-local/UTC conversion, transition affordance mapping, and error
  normalization.
- Injected-fetch API tests for exact URLs, credentials, CSRF, idempotency,
  expected-version bodies, cancellation, and response envelopes.
- Structural/SSR tests with existing React/Vitest capabilities for meaningful
  headings, labels, links, state text, and no secret-bearing markup.
- Real-browser smoke with the real API: unauthenticated login handoff,
  authorized list/create/edit/transition, stale conflict, keyboard navigation,
  mobile-width layout, and direct route load.
- Do not add a DOM-test stack in this task solely for shallow render tests; add
  one later only if W2-08 identifies interaction coverage that cannot be proven
  through the existing stack and browser verification.

## Verification

```bash
pnpm --filter @persona-apps/operations-web test
pnpm --filter @persona-apps/operations-web typecheck
pnpm --filter @persona-apps/operations-web lint
pnpm --filter @persona-apps/operations-web build
make validate
```

Then run an authenticated real-API smoke against disposable/local data and
record which operator permissions were exercised. A successful build or mock
response alone does not satisfy the task.

## Deliverables

- Functional Event and Offering Operations routes and screens.
- Explicit session/query/API modules with focused tests.
- Accessibility/responsive and authenticated real-API verification evidence.
- Current-state update and archived executed task only after acceptance.

## Risks and Clarifications to Review

- Confirmed default: this is a thin functional Week 2 UI; full visual and E2E
  hardening remains scheduled later.
- Confirmed default: no optimistic updates for contested business state.
- Confirmed default: CSRF is memory-only and idempotency keys are per frozen
  retry intent.
- Confirmed default: `/events` remains distinct from the later dashboard
  projection route.
- Remaining execution dependency: a test operator with the exact Event and
  Offering permissions and a configured test OIDC identity/provider must exist
  in the local/disposable environment; this task does not create credentials or
  a mock provider silently.

## Final Report Requirements

Distinguish implemented screens, test/build evidence, authenticated real-API
evidence, permission assumptions, deferred polish/E2E work, plan variance,
remaining usability/security risks, and the next Storefront task.
