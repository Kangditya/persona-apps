# Task: W2-07 Build Storefront Event Landing, Offering List, and Offering Detail

## Executed

## Status

Completed and verified on 2026-08-14 after the user explicitly approved
execution. No Purchase control or client-authoritative commerce rule was added.

## Tracker

- Week: Week 2
- Epic: Event & Offering
- Week goal: Allow Operations to configure commerce and Storefront to discover
  it.
- Milestone: Public Event and Offering discovery.
- Weekly allocation: 42 hours shared across W2-01 through W2-08; the tracker
  does not assign per-task estimates, so this plan does not invent them.
- Tracker objective: Build Storefront Event landing, Offering list, and
  Offering detail.
- Dependency: W2-04 public Event/Offering queries.

## Objective

Replace Storefront catalogue placeholders with a thin, accessible, responsive
guest discovery flow that renders only the active Event and its published
Offerings from the real public API, while presenting registration and
availability as informative snapshots rather than transactional guarantees.

## Context

The Storefront React application already has centralized routes, TanStack
Query, a shared API client, PWA navigation rules, a layout, and placeholder
Home/Offerings pages. It has no Event/Offering API module, detail route, public
catalogue queries, or production data states.

The approved public contract intentionally exposes no images, species/package
composition, checkout availability lock, or currency exponent/format metadata.
This task must not invent those fields or imply that an advisory quota read
reserves capacity.

## Source of Truth

- `docs/PRD.md` Storefront actors, Event/Offering discovery, privacy, and
  accessibility requirements;
- `docs/PRODUCT_MAP.md` Storefront ownership, Offering Catalogue, and roadmap;
- `docs/ARCHITECTURE.md` frontend/API boundaries, public exposure, PWA, and
  deployment rules;
- `docs/DECISIONS.md` ADR-002, ADR-006, ADR-009, ADR-010, ADR-032,
  ADR-038, ADR-039, ADR-042, ADR-043, and ADR-046;
- `docs/CONVENTIONS.md` React, route, TanStack Query, accessibility, money,
  status, and API conventions;
- `contracts/openapi/storefront.yaml` after W2-04;
- current Storefront routes, API client, query client, layout, PWA source, and
  shared `@persona-apps/ui` primitives.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate after W2-04 is implemented and verified against real repository
queries. Re-read the final public DTOs and base-path behavior before BUILD. Do
not commit or push unless asked.

## Scope

### In Scope

- Home/Event landing state backed by `GET /api/public/v1/events/active`.
- Event-scoped Offering list backed by
  `GET /api/public/v1/events/{event_id}/offerings`.
- Offering detail route backed by
  `GET /api/public/v1/offerings/{offering_id}`.
- Centralized Offering detail path pattern and builder.
- Stable TanStack Query keys, cancellation, bounded read retry, and targeted
  invalidation/refetch behavior.
- Explicit loading, no-active-event, empty-catalogue, not-found, rate-limited,
  dependency-unavailable, generic-error, and success states.
- Event registration-window presentation using server UTC instants and an
  explicit browser timezone label.
- Offering code/name/kind/description, exact price representation,
  per-Purchase participant capacity, and advisory availability presentation.
- Baseline semantic headings, links, focus, status messaging, keyboard access,
  contrast-token use, and mobile/desktop layout.
- PWA/API boundary verification so public API responses are never served from
  the navigation fallback or cached as application shell data.

### Out of Scope

- Purchase creation, checkout forms, reservation guarantees, payment, saving,
  giveaway, accounts, carts, favorites, promotions, or disabled fake CTAs.
- Search, filters, sorting controls, public pagination, recommendations,
  analytics, personalization, or client-side business eligibility decisions.
- Invented Offering images, livestock details, location, slaughter schedule,
  stock badges, species taxonomy, or package composition.
- SEO SSR/prerendering, social metadata automation, offline catalogue data, or
  a service-worker API cache.
- A design-system redesign, animation framework, carousel, custom skeleton
  library, or new frontend dependency.

## Existing State

- `/`, `/offerings`, and `/purchase-tracking` are centrally registered.
- `HomePage` and `OfferingsPage` are placeholders; no Offering detail path
  exists.
- Storefront diagnostics already use TanStack Query and the shared API client.
- The API client defaults to `/api`; W2-04 must establish the one correct
  endpoint/base split and remove the local proxy rewrite that strips `/api`.
- PWA configuration already excludes `/api/` from navigation fallback; this
  task must preserve and test that rule.
- Shared cards, badges, alerts, skeletons, buttons, and accessible field
  primitives exist; no catalogue-specific component abstraction exists.

## Target State

- `/` fetches the active Event and clearly renders either its public details or
  a truthful no-active-event state.
- `/offerings` first resolves the active Event, then fetches only that Event's
  public Offerings. It never requests an internal Event or displays hidden
  statuses.
- `/offerings/:offeringId` renders one published Offering under the active Event
  or an indistinguishable public not-found state.
- Query functions receive TanStack Query's `AbortSignal`; unmounted/superseded
  navigation does not leave avoidable requests running.
- Read retries are small and limited to transient network/`5xx` failures; no
  retry loop runs for `400`, `404`, or `429`.
- Registration state is a presentation label computed from returned instants:
  not configured, upcoming, open, or closed. It is explicitly non-authoritative
  and future checkout revalidates on the server.
- Availability renders as:
  - positive finite units: the advisory participant-unit count;
  - zero: currently unavailable, without promising permanent exhaustion;
  - `null`: no configured finite total quota, not infinite guaranteed stock.
- No Purchase button is rendered until a real W3 purchase route exists.

## Constraints

- Reuse React, React Router, TanStack Query, the shared API client, shared UI,
  CSS/Tailwind utilities, and browser `Intl`/date APIs. Add no dependency.
- Keep endpoint strings in one public catalogue API module and routes in the
  central registries.
- Do not fetch all Events and filter in the browser. The public API must own
  active/published visibility.
- Treat server content as text rendered through React; never use
  `dangerouslySetInnerHTML` for descriptions or errors.
- Do not calculate authoritative stock, registration eligibility, or commerce
  state in the browser.
- API-visible integer amounts remain JavaScript-safe, but the UI must retain
  the original integer and must not use floating-point math for business rules.
- Do not assume every currency has two decimals. The contract currently
  provides `price_minor` and `currency_code` but no exponent/display amount.

## Implementation Requirements

### API and query layer

- Add explicit public Event and Offering types matching only the reviewed
  Storefront contract; do not reuse Operations DTOs.
- Add narrow runtime guards for the required public envelope/resource fields;
  treat a malformed successful payload as a safe API failure instead of relying
  on a compile-time cast.
- Use canonical `/api/public/v1/...` endpoint paths under W2-04's tested base
  convention.
- Add stable keys for active Event, Event Offerings, and Offering detail. The
  Event Offerings query is disabled until a valid active Event ID exists.
- Forward `AbortSignal`, preserve normalized error code/request ID, and avoid
  exposing raw response bodies.
- Use the current query client defaults unless a public-catalogue-specific
  stale time is proven necessary. A cache hit must never be described as a
  quota guarantee.

### Home/Event landing

- Render Event name/year and registration-window state with the local timezone
  named when timestamps are present.
- Keep `data: null` as a normal no-active-event state, not an application error.
- Link to the Offering list only when an active Event exists; do not present a
  dead commerce control otherwise.
- Avoid adding fields absent from the public contract, including internal
  quota, audit, version, or Operations lifecycle history.

### Offering list

- Render deterministic API order without client-side resorting.
- Each item shows only public name/code/kind, optional description summary,
  exact price representation, per-Purchase participant capacity, advisory
  availability, and an accessible detail link.
- Distinguish an active Event with no published Offerings from an unavailable
  Event and from an API failure.
- Use responsive CSS grid/list markup and real headings/links; do not make a
  non-interactive card impersonate a button.

### Offering detail

- Resolve the ID from the centralized route and let the API enforce visibility.
- Render the same exact commercial facts as the list plus full public
  description. Never expose or infer internal quota composition.
- Provide a link back to Offerings and no checkout CTA until W3 supplies the
  corresponding route and server behavior.
- Map hidden/nonexistent Offering to the same not-found screen.

### Price, time, and availability presentation

- Until the contract defines a currency exponent or server display amount,
  show the exact integer with currency and an explicit “minor units” label,
  for example `IDR 125000 minor units`. Do not divide by 100 or treat the value
  as a major-unit currency amount.
- If product review accepts a currency-exponent policy before activation,
  amend the contract/plan first and implement one shared formatter with exact
  tests; do not add an ad hoc currency table in a page.
- Parse RFC 3339 timestamps defensively, display them in the browser locale,
  and show the resolved IANA browser timezone. Invalid server timestamps are a
  safe UI error and contract defect, not `Invalid Date` text.
- Label availability as advisory and explain that final availability is
  confirmed later. Do not animate countdowns or run second-by-second timers.

### States and accessibility

- Use skeletons or concise progress text with an accessible status, but keep
  page headings stable during loading.
- Error states include a retry control only for retryable reads and expose a
  request ID for support when available, without internal details.
- A rate-limited state shows a validated server retry delay when available and
  does not run an automatic retry loop.
- Preserve visible focus, logical heading order, descriptive links, and
  sufficient touch targets at narrow widths.
- Verify content at text zoom and reduced-motion settings; add no motion that
  requires a preference branch.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `apps/storefront-web/src/routes/paths.ts` | Modify | Add Offering detail pattern and safe builder. |
| `apps/storefront-web/src/routes/routes.tsx` | Modify | Register the public Offering detail page. |
| `apps/storefront-web/src/api/client.ts` | Modify if W2-04 has not already done so | Use the single verified canonical path convention. |
| `apps/storefront-web/src/api/catalogue.ts` | Create | Public DTOs and three exact endpoint functions. |
| `apps/storefront-web/src/api/catalogue.test.ts` | Create | Verify URLs, cancellation, DTO envelopes, and error behavior. |
| `apps/storefront-web/src/features/catalogue/queries.ts` | Create | Stable query keys/hooks for active Event and Offerings. |
| `apps/storefront-web/src/features/catalogue/presentation.ts` | Create only if reused | Minimal date, availability, and exact-price presentation helpers. |
| `apps/storefront-web/src/pages/HomePage.tsx` | Modify | Real active-Event landing/no-event states. |
| `apps/storefront-web/src/pages/OfferingsPage.tsx` | Modify | Real active Event catalogue states and detail links. |
| `apps/storefront-web/src/pages/OfferingDetailPage.tsx` | Create | Public Offering detail/not-found flow. |
| `apps/storefront-web/src/routes/paths.test.ts` and focused page/helper tests | Modify/create | Verify paths, states, presentation boundaries, and public field exposure. |
| `apps/storefront-web/src/styles/global.css` | Modify only if needed | Minimal catalogue responsive/state styling. |
| `apps/storefront-web/vite.config.ts` and PWA tests | Verify/modify only if W2-04 did not finish it | Keep `/api/` out of fallback/cache and preserve proxy paths. |
| `.codex/CURRENT_STATE.md` and `.codex/TASK.md` | Modify/create/archive when executed | Record verified discovery behavior. |

Small presentation helpers may remain in pages until a second caller exists.
Do not create a catalogue component library for three screens.

## Dependencies and Sequencing

- Requires W2-04's public endpoints, exposure tests, and base-path correction.
- Can be implemented alongside W2-06 once shared client conventions are
  stable; do not import code between the two applications.
- W2-08 closes remaining route/public-exposure regression gaps. This task still
  owns focused API/path/presentation tests and a real-browser smoke.
- W3 adds Purchase entry; this task deliberately leaves no fake CTA to wire up
  later.

## API and Database Impact

- No route, OpenAPI, or migration is introduced by this frontend task.
- Any missing public field or semantics must be resolved in the owning contract
  and W2-04, not by calling Operations endpoints from Storefront.
- Browser cache/TanStack Query never becomes authoritative transaction state.

## Privacy and Security

- The Storefront sends no Operations cookie requirement, CSRF token, operator
  permission, or internal ID beyond public resource IDs.
- Only Storefront DTO fields are rendered. Tests assert that version, internal
  quotas, lifecycle history, audit, reservation details, and operator data are
  absent.
- React text escaping remains intact; URLs are internal route builders, not
  caller-provided HTML or redirects.
- API errors are minimized and rate limits are respected; the UI does not
  automatically hammer a `429` endpoint.
- The service worker does not cache public business responses or error bodies.

## Acceptance Criteria

1. Home renders the active Event or a truthful no-active-event state from the
   real public API.
2. Offerings renders only the active Event's published catalogue and correctly
   distinguishes loading, empty, hidden/not-found, rate-limited, unavailable,
   and generic error states.
3. Centralized links/direct loads render public Offering detail or an
   indistinguishable not-found state.
4. Registration labels use returned UTC instants, show browser timezone, and
   remain explicitly informational rather than checkout authorization.
5. Capacity and finite/zero/unbounded advisory availability are labeled
   accurately and never presented as a reservation guarantee.
6. Price is rendered exactly without an invented decimal exponent; any later
   localized-major-unit formatter requires an accepted contract policy.
7. Query keys, cancellation, bounded retry, base paths, and PWA exclusions are
   tested; `/api/` responses are never navigation-fallback or offline data.
8. Baseline semantic, keyboard, focus, text-zoom, responsive, and safe-error
   behavior is verified.
9. No Operations endpoint/data, mock catalogue, fake purchase CTA, invented
   product field, new dependency, or client-side business authority is added.

## Testing

- Path-builder and direct-route tests for valid IDs and fallback behavior.
- Injected-fetch API tests for exact paths, active `data: null`, Event
  Offerings/detail envelopes, cancellation, `404`, `429`, `503`, request IDs,
  and no credentials-only headers.
- Pure presentation tests for registration boundaries, invalid timestamps,
  finite/zero/null availability, exact safe-integer prices, and timezone label.
- Structural/SSR tests with the existing stack for headings, links, status
  text, descriptions, and absence of internal fields/fake purchase controls.
- Real-browser smoke at desktop/mobile widths for no active Event, empty list,
  populated list/detail, retryable error, keyboard navigation, text zoom, and
  API/PWA behavior.

## Verification

```bash
pnpm --filter @persona-apps/storefront-web test
pnpm --filter @persona-apps/storefront-web typecheck
pnpm --filter @persona-apps/storefront-web lint
pnpm --filter @persona-apps/storefront-web build
make validate
```

Then run the real Storefront against the real local/disposable API and inspect
network paths, returned public fields, service-worker behavior, responsive
layout, keyboard navigation, and empty/404/429 states. A mock-only or build-only
result is incomplete.

## Deliverables

- Real active-Event landing, Offering list, and Offering detail routes.
- Minimal catalogue API/query/presentation code with focused tests.
- Public exposure, accessibility/responsive, and real-browser verification
  evidence.
- Current-state update and archived executed task only after acceptance.

## Risks and Clarifications to Review

- Confirmed default: discovery remains visible outside the registration window
  and communicates upcoming/open/closed state without authorizing checkout.
- Confirmed default: `null` availability means no configured finite total,
  while all displayed availability remains advisory.
- Confirmed default: no fake Purchase CTA or invented product content is added.
- Remaining product clarification: the contract does not define a currency
  exponent/display amount. The safe Week 2 default is explicit raw minor-unit
  text; approve a currency policy before requesting localized major-unit price.
- Later SEO/SSR, richer visual polish, and end-to-end purchase flow remain
  planned work and are not hidden in this task.

## Execution Evidence and Plan Variance

- Real-browser checks used the real Go API and disposable PostgreSQL rows for
  populated list/detail, active-Event absence, empty published catalogue,
  not-found, registration timezone, finite availability, rate limiting, and a
  temporary network failure. Test rows were restored after each state.
- The mobile list/detail rendered at 375 pixels without horizontal overflow;
  semantic heading order, links, terms, status regions, and the absence of a
  Storefront Purchase/checkout control were inspected. Native links received
  browser focus; the in-app harness did not advance focus when injecting Tab,
  so a physical-keyboard tab-order pass remains release QA. No motion was
  added.
- Browser zoom alteration was unavailable in the in-app browser. Rem-based
  layout, wrapping, mobile width, and long-text-safe rendering were verified;
  a separate manual 200-percent zoom pass remains appropriate for release QA.
- A single bounded page of up to 100 Offerings is rendered, as planned. When a
  next cursor exists, the UI states that the Week 2 view is limited rather than
  inventing an unapproved pagination control.
- The generated service worker contains only shell precaching and a
  `^/api/` navigation denylist. A real Vite-proxied public response remained
  JSON with `Cache-Control: no-store`.

## Final Report Requirements

Distinguish implemented public screens, API/browser/PWA evidence, exact price
presentation, timezone/availability assumptions, deferred Purchase/SEO/polish
work, plan variance, remaining public UX/security risks, and the next test task.
