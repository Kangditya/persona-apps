# Task: W3-08 Build Operations Purchase List and Detail

## Status

Draft prepared on 2026-08-24 and ready for review. The live tracker status is
`Ready` because W3-03 and W3-06 are complete. This remains an inactive plan:
do not begin BUILD until it is explicitly approved and copied to a new
`.codex/TASK.md`. Do not commit or push unless requested.

## Tracker

- Week: Week 3
- Epic: Party & Purchase
- Application: Operations
- Category: Frontend
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Build Operations Purchase list and detail.
- Dependencies: W3-03 and W3-06.
- Acceptance summary: an operator with `purchase.read` can find bounded
  Purchases and inspect the contract-exposed Purchase snapshot and explicit
  Party relationships without database access.

## Objective

Replace the Operations `/purchasing` placeholder with the smallest useful,
read-only Purchase workflow over the implemented
`GET /api/operations/v1/purchases` and
`GET /api/operations/v1/purchases/{purchase_id}` routes.

The workflow must preserve the Operations session boundary, enforce
`purchase.read` in the Go API, use only the Operations contract, paginate and
filter exactly as the backend supports, and present stored Purchase facts
without inventing mutations, search modes, or lifecycle authority.

## Context

The backend already exposes permission-gated cursor-paginated Purchase list and
detail reads. List order is `created_at DESC, id ASC`. Supported filters are
`event_id` and `status`; there is no purchase-reference, contact, participant,
or free-text search endpoint.

Operations already has Next.js App Router, an authenticated OIDC session query,
memory-only CSRF, a reusable `AccessBoundary`, centralized paths, TanStack
Query, the credentialed API client, and working Event/Offering list/detail
patterns. `/purchasing` is still a placeholder and no Purchase frontend module
exists.

## Source of Truth

- `docs/PRD.md`, especially Operations scope, role separation, Purchase
  Management, privacy, pagination, security, and data integrity;
- `docs/PRODUCT_MAP.md`, Purchasing and Administration & Reporting ownership;
- `docs/ARCHITECTURE.md`, Operations and Go API boundaries, query/frontend
  ownership, authorization, caching, and concurrency;
- `docs/DECISIONS.md`, especially ADR-002, ADR-004 through ADR-006, ADR-010
  through ADR-014, ADR-018, ADR-019, ADR-033, ADR-038, ADR-039, ADR-043,
  ADR-044, ADR-047, ADR-048, and ADR-049;
- `docs/CONVENTIONS.md`;
- `docs/security/AUTHENTICATION.md` and
  `docs/security/PERMISSIONS.md`;
- `contracts/openapi/operations.yaml`;
- completed W3-03 and W3-06 plans and execution evidence;
- current `apps/operations-web`, `packages/api-client`,
  `apps/api/internal/purchasing/operations_http.go`, Purchase repository, and
  composed API integration tests;
- `.codex/CURRENT_STATE.md` and `MVP-DELIVERY-ROADMAP.md`.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

At activation, compare the live handler JSON and final Operations OpenAPI
schemas again. A mismatch is fixed in its owning API/contract layer through a
recorded plan variance; the frontend must not query PostgreSQL, call the public
API, or guess missing fields.

## Scope

### In Scope

- Keep `/purchasing` as the centralized Purchase-list route and replace its
  placeholder screen.
- Add `/purchasing/[purchaseId]` as a thin App Router detail entry and a
  centralized UUID-safe detail URL builder.
- Add an Operations Purchase API module with exact list/detail query
  construction and narrow runtime response validation.
- Add feature-owned TanStack Query keys/hooks for filtered cursor pages and
  Purchase detail, forwarding cancellation signals.
- Require `purchase.read` through the existing Operations access/session
  presentation while preserving backend enforcement.
- Render a bounded Purchase page with Event UUID/status filters, resettable
  filter state, cursor previous/next navigation, and distinct loading, empty,
  invalid-filter, unauthorized, forbidden, not-found, transient, and success
  states.
- Render detail fields available in the Operations contract/runtime: Purchase
  UUID/reference, Event/Offering IDs, captured Offering name, channel, status,
  purchaser, optional payer, participant count, ordered participant-name
  snapshots, exact total/currency, and timestamps.
- Preserve accessible headings, filter labels, table/list semantics, links,
  focus, status messaging, keyboard use, and representative narrow layouts.

### Out of Scope

- Free-text, Purchase-reference, Party-contact, participant-name, full Event
  selector, date-range, sorting, export, saved-search, or server total-count
  behavior absent from the contract.
- Purchase mutation, correction, cancellation, payment verification, quota
  override, participant replacement, or audit UI.
- Client-side authorization, direct database access, public Purchase token
  access, or Storefront DTO reuse.
- Displaying snapshots or relationships the Operations response does not
  expose, including internal token/hash/idempotency/outbox/reservation data.
- Polling, dashboard projections, optimistic updates, bulk selection, infinite
  scrolling, TanStack Table/Form/Router, or a new frontend dependency.
- Completing the richer W6-03 search/filter/exception workflow.

## Existing State

- `src/app/purchasing/page.tsx` is a thin route to the placeholder
  `src/screens/PurchasingPage.tsx`.
- `src/routes/paths.ts` owns `paths.purchasing` but has no Purchase detail
  builder.
- `src/api/session.ts` and `features/events/AccessBoundary.tsx` implement the
  current session, sign-in, and permission presentation pattern.
- `src/api/events.ts` and `features/events/queries.ts` demonstrate exact
  credentialed reads, guarded DTOs, cursor pages, and stable query keys.
- The Operations client always sends `credentials: "include"`. Safe GET reads
  do not require CSRF or an idempotency key.
- The backend list accepts only `cursor`, `limit`, `event_id`, and `status`.
  Valid list statuses are `PENDING_PAYMENT`, `PAID`, `ELIGIBLE`, and
  `CANCELLED`.
- List responses omit participant rows; detail responses include ordered
  participant display-name snapshots. Runtime Party summaries currently expose
  only UUID and display name.

## Target State

- An authenticated operator navigating to `/purchasing` is challenged through
  the existing session flow and sees data only when `purchase.read` is present.
- The page requests one server-bounded page at a time. Query keys include
  normalized Event filter, status filter, and opaque cursor.
- Filters map only to exact contracted query parameters. Invalid UUIDs are
  rejected before the request with an accessible message; no broad or guessed
  query is sent.
- The list shows a concise summary and a centralized detail link for each
  Purchase. It does not load each detail or participant list, avoiding N+1
  requests.
- The detail route validates the path ID before enabling its query, handles
  `404` distinctly, and renders explicit purchaser, payer, and ordered intended
  participant roles without calling them one generic customer.
- All displayed commercial values come from the Operations Purchase response.
  The browser does not rejoin live Offering data or recompute totals.
- Query cache remains a private view cache and is removed by the existing
  logout `privateRoot` behavior.

## Constraints

- Use App Router filesystem routes and keep route files thin.
- Keep URL builders in `src/routes/paths.ts` and endpoint strings in
  `src/api/purchases.ts`.
- Use the current Operations client, session query, `AccessBoundary`,
  `safeErrorMessage`, TanStack Query, React state, native form controls, and
  shared UI. Add no dependency.
- Do not relocate the existing session/access code merely because it currently
  lives under `features/events`; a cross-feature auth refactor is outside this
  task.
- Treat API identifiers, cursors, timestamps, statuses, and integers as
  untrusted until the Purchase response guard validates them.
- Render backend strings through React text only. Do not expose arbitrary error
  details or use `dangerouslySetInnerHTML`.
- Do not attach CSRF or idempotency headers to these read-only GET requests.

## Implementation Requirements

### API and response boundary

- Define separate summary/detail frontend types when endpoint guarantees differ.
- Construct list queries with `URLSearchParams` and only non-empty
  `event_id`, `status`, `cursor`, and bounded `limit` values.
- Parse the `data` plus `page.limit`/`page.next_cursor` envelope and reject
  malformed cursors, UUIDs, timestamps, currency, unsafe integers, statuses,
  Party summaries, or inconsistent detail participant counts.
- Detail parsing must require the participant snapshot array described by the
  endpoint/runtime; a missing or malformed array is a safe contract-response
  error, not an empty invented relationship.
- Preserve optional payer behavior. Never infer payer from purchaser when the
  field is absent.
- Do not assume optional Party email/phone values exist; render only fields
  actually returned and validated.

### Queries, filters, and pagination

- Namespace keys under `["operations", "private", "purchases"]` and include
  normalized filters/cursor.
- Forward TanStack Query's `AbortSignal` to every read.
- Do not automatically retry `400`, `401`, `403`, or `404`. Keep transient
  retry bounded or use the current no-retry default with an explicit retry
  control.
- Reset cursor history when a filter changes or is cleared.
- Treat `next_cursor` as opaque. Never decode, manufacture, place business
  meaning on it, or request all pages in the background.
- The tracker word “search” is implemented only as the current contracted
  Event/status filtering plus bounded navigation. W6-03 owns additional search
  semantics.

### Access, errors, and presentation

- Wrap both list and detail screens in `AccessBoundary
  permission="purchase.read"`. Backend `401`/`403` remains authoritative on
  every request.
- Refresh the session query through the existing behavior when a backend read
  reports an expired/invalid session; do not enter a retry loop.
- List cards/rows expose the non-secret reference, status, Event, captured
  Offering name, purchaser, amount/currency, and created time with a detail
  link.
- Detail uses a definition list or similarly semantic grouping for identity,
  commercial snapshot, participant snapshots, and lifecycle timestamps.
- Display exact minor-unit values without floating-point conversion or an
  invented currency exponent.
- Show no mutation affordance, payment claim, reservation countdown, or current
  quota calculation.
- Maintain focus on page headings after direct navigation and expose request IDs
  only through the existing safe error presentation.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `apps/operations-web/src/routes/paths.ts` | Modify | Add centralized Purchase detail URL construction. |
| `apps/operations-web/src/routes/paths.test.ts` | Modify | Verify encoded Purchase paths and active navigation. |
| `apps/operations-web/src/app/purchasing/[purchaseId]/page.tsx` | Create | Thin App Router entry for Purchase detail. |
| `apps/operations-web/src/api/purchases.ts` | Create | Exact Operations Purchase list/detail client and guarded DTO mapping. |
| `apps/operations-web/src/api/purchases.test.ts` | Create | Verify filters, cursors, credentials, cancellation, envelopes, and malformed responses. |
| `apps/operations-web/src/features/purchases/queries.ts` | Create | Feature-owned private query keys and list/detail hooks. |
| `apps/operations-web/src/screens/PurchasingPage.tsx` | Modify | Replace the placeholder with permission-gated filters, cursor list, and explicit states. |
| `apps/operations-web/src/screens/PurchaseDetailPage.tsx` | Create | Permission-gated snapshot and Party/participant relationship detail. |
| `.codex/CURRENT_STATE.md` | Modify after verified execution | Record the implemented Operations Purchase reads and deferred W6 scope. |

Reuse `features/events/AccessBoundary.tsx` and `features/events/errors.ts`
without moving or copying them. Keep small display formatters in the owning
screen until a second Purchase screen proves reuse.

## API, Database, and Dependency Impact

- API route: no change; consume the two implemented Operations GET routes.
- OpenAPI: no planned change. A verified schema/runtime mismatch returns to
  PLAN and the owning W3-03 contract/API layer.
- Database/migration: none.
- Authorization: frontend affordance uses `purchase.read`; backend enforcement
  remains mandatory.
- CSRF/idempotency/audit/outbox: no impact because both operations are reads.
- Dependency/lockfile: none.
- PWA/cache: private Purchase data remains network-only and outside the service
  worker; TanStack Query remains in-memory private view state.

## Acceptance Criteria

1. `/purchasing` loads a bounded Operations Purchase page through the exact
   contract and `credentials: "include"`.
2. Both list and detail require the existing session presentation and
   `purchase.read`, while backend `401`/`403` remains authoritative.
3. Event UUID/status filtering and opaque previous/next cursor navigation match
   the only currently supported server query semantics.
4. The list performs no per-row detail/participant request and renders distinct
   loading, empty, invalid-filter, auth, forbidden, transient, and success
   states.
5. Centralized detail links/direct loads validate the UUID and render safe
   not-found behavior.
6. Detail visibly distinguishes purchaser, optional payer, and ordered intended
   participant snapshots.
7. Captured Offering name, exact amount/currency, status, references, and times
   come from the Operations response; no live catalogue join or recalculation
   changes history.
8. No raw Purchase token/hash, idempotency data, reservation internals, SQL,
   audit/outbox payload, public DTO, or unrelated Party data is exposed.
9. Safe GETs send no CSRF/idempotency header and introduce no mutation or
   optimistic state.
10. Keyboard, focus, labels, semantic data grouping, request-state messaging,
    text zoom, and representative narrow/desktop layouts are verified.
11. Logout still removes all private Purchase query data through the existing
    private-root prefix.
12. No uncontracted search, generated client, new dependency, backend
    workaround, or W6 behavior is added.

## Testing

- Injected-fetch API tests for exact paths, credentials, filter encoding,
  cursor opacity, cancellation, list/detail envelopes, optional payer, summary
  versus detail participants, unsafe integers, timestamps, and normalized
  errors.
- Query-key tests or direct assertions proving Event/status/cursor isolation and
  the `privateRoot` logout prefix.
- Path tests for encoded Purchase detail URLs and `/purchasing` active state.
- Structural tests with the existing React/Vitest capability for role labels,
  safe snapshot fields, empty/error states, and absence of mutation/token
  content.
- Real-browser smoke with the real OIDC/session API and disposable PostgreSQL:
  unauthenticated, forbidden, permitted list, filters, cursor page, direct
  detail, missing Purchase, logout cache removal, keyboard flow, and 360px
  layout.

Do not add a table library, DOM harness, or browser E2E dependency solely for
this read-only slice. W3-09 audits and closes only demonstrated coverage gaps.

## Verification

Run from the repository root:

```bash
pnpm --filter @persona-apps/operations-web test
pnpm --filter @persona-apps/operations-web typecheck
pnpm --filter @persona-apps/operations-web lint
pnpm --filter @persona-apps/operations-web build
make validate
docker compose -f infrastructure/compose.yaml config
```

Run the real Operations/API/PostgreSQL browser smoke described above. Report
focused frontend results separately from unrelated repository-wide failures;
mock-only data does not satisfy the acceptance criteria.

## Deliverables

- Real permission-gated Purchase list at `/purchasing`.
- Centralized, directly loadable Purchase detail route.
- Exact Operations Purchase API/query boundary with guarded DTOs.
- Contract-backed filters, cursor navigation, explicit role/snapshot
  presentation, focused tests, and real browser evidence.
- Current-state update and canonical task archival only after execution is
  approved, implemented, and verified.

## Risks, Decisions, and Deferred Work

- Accepted: `purchase.read` is the only permission for these reads; frontend
  checks are presentation only.
- Accepted: current list “search” means Event/status filtering and cursor
  navigation. There is no free-text or reference-search API to call.
- Contract/runtime nuance: detail runtime includes participant snapshots and
  list runtime omits them; validate this endpoint-specific distinction instead
  of treating one shape as universal.
- Deferred: richer search, filters, exception display, and commerce control
  plane in W6-03; all Purchase mutations and Payment behavior remain Week 4+.
- Deferred: polling/dashboard projections. This screen reads authoritative
  queries on demand and does not imply realtime freshness.

## Final Report Requirements

The execution report must list exact files, query/filter semantics, response
fields, permission and session behavior, browser/API evidence, accessibility
checks, cache/logout behavior, plan variance, and deferred W6/payment work. It
must distinguish implemented, verified, assumed, and deferred behavior and
must not call Operations Purchase administration complete.
