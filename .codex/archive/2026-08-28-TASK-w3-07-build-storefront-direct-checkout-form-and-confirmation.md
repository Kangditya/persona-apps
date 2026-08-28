# Task: W3-07 Build Storefront Direct-Checkout Form and Confirmation

## Executed

## Status

Implemented and verified as of 2026-08-28. The user explicitly approved the
Penpot-reconciled W3 plan, authorized W3-07 BUILD, and confirmed that the
Penpot edits should be applied. The live tracker row W3-07 is now `Done`.
W3-08 remains `Ready`; W3-09 remains `Backlog`. No commit or push was made.

## Tracker

- Week: Week 3
- Epic: Party & Purchase
- Application: Storefront
- Category: Frontend
- Priority: P0
- Estimate: 7 hours
- Tracker objective: Build Storefront direct-checkout form and confirmation.
- Dependencies: W2-07, W3-02, W3-03, and W3-06.
- Acceptance summary: a public user submits one direct `COMMON` Purchase and
  receives the successful Purchase confirmation without exposing the one-time
  Purchase access token.

## Objective

Complete the smallest real Storefront checkout slice over the implemented
`POST /api/public/v1/purchases` command. The user selects one published
Offering, explicitly declares purchaser, payer, and intended-participant
relationships, submits one idempotent guest intent, and sees the committed
`PENDING_PAYMENT` Purchase facts returned by the Go API.

The Go API remains authoritative for Event/Offering eligibility, price,
participant capacity, quota, role validation, snapshots, totals, reservation,
reference generation, token generation, and idempotency.

## Context

W3-01 through W3-06 are implemented and verified. The backend now creates one
atomic `COMMON` Purchase with explicit Party roles, stored commercial and
participant snapshots, a 24-hour quota reservation, a non-secret support
reference, a one-time opaque access token, minimized outbox events, and durable
encrypted replay.

Storefront already has Next.js App Router, centralized URL builders, TanStack
Query, the shared API transport, the public Event/Offering catalogue, accessible
UI primitives, and shell-only PWA caching. It has no Purchase API module,
checkout route, checkout form, or confirmation UI. The public Purchase tracking
GET is contract-only and has no runtime handler, so this task must not pretend
that a reloadable tracking screen exists.

## Source of Truth

- `docs/PRD.md`, especially Common Purchasing, Storefront principles, Party
  distinctions, Purchase Management, security, privacy, and data integrity;
- `docs/PRODUCT_MAP.md`, Storefront, Purchasing, Party & Participant, and the
  first vertical slice;
- `docs/ARCHITECTURE.md`, application boundaries, frontend data ownership,
  Storefront access, API separation, transactions, caching, and concurrency;
- `docs/DECISIONS.md`, especially ADR-002, ADR-004 through ADR-006, ADR-010
  through ADR-014, ADR-018, ADR-033, ADR-038, ADR-039, ADR-041 through ADR-045,
  and ADR-047 through ADR-050;
- `docs/CONVENTIONS.md`;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- `docs/security/AUTHENTICATION.md` and
  `docs/security/PERMISSIONS.md`;
- `contracts/openapi/storefront.yaml`;
- completed W3-01 through W3-06 plans and execution evidence;
- current `apps/storefront-web`, `packages/api-client`, `packages/ui`, and
  `apps/api/internal/purchasing` source and tests;
- `.codex/CURRENT_STATE.md` and `MVP-DELIVERY-ROADMAP.md`.
- the user-supplied
  [QurbanPlus Penpot file](https://design.penpot.app/#/workspace?team-id=81f57451-85cc-819d-8008-7c2bac979fc9&file-id=81f57451-85cc-819d-8008-7c2c1dbf6c2a&page-id=81f57451-85cc-819d-8008-7c2c1dbf6c2b),
  specifically
  `13 · Storefront · 04 Checkout Participant` and
  `16 · Storefront · 07 Purchase Confirmation`, as a visual and information-
  architecture reference subordinate to the canonical product, API, privacy,
  and implementation artifacts above.

## Penpot Design Evidence and Reconciliation

The inspected Penpot file contains one `862 × 1650` board for checkout and one
`862 × 1650` board for confirmation:

- Checkout uses a narrow single-column layout, a non-interactive three-stage
  progress indicator (`Data`, `Pembayaran`, `Selesai`), cards for
  `Data pembeli`, `Sohibul Qurban`, and `Ringkasan paket`, a minimum-data
  privacy notice, one primary action, and a help block.
- Confirmation uses a success heading, `Menunggu pembayaran` status,
  non-secret reference/date/total, package and participant summary, next-step
  guidance, a payment-instructions action, and a help block.

Adopt the card order, visual hierarchy, compact summary, privacy reassurance,
success/reference/status presentation, and help placement. Reconcile these
design-only assumptions as follows:

- The accepted payload also requires an explicit payer. Keep same/distinct
  payer controls inside the buyer card even though the board does not show that
  contract detail.
- Checkout participants are intended qurban participants, not yet activated
  Sohibul Qurban. User copy must not imply activation before verified payment.
- Penpot's `Rp` rendering is not a currency-exponent contract. Retain the
  repository's exact minor-unit presentation until a canonical display policy
  is accepted.
- The payment-instructions CTA and automatic-verification copy belong to Week
  4/5 behavior. Omit them in W3-07 instead of shipping a dead control or false
  lifecycle promise.
- Replace Penpot's `Lanjutkan ke pembayaran` submit label with truthful
  Purchase-creation wording; W3-07 creates the pending Purchase but cannot
  navigate into an unimplemented payment flow.
- The help block may contain truthful static guidance. Add no `Hubungi kami`
  action until a real, configured contact target exists.
- Penpot copy is Indonesian while the current Storefront is English and has no
  accepted localization workflow. Adopt the hierarchy, not a one-screen
  language switch; localization remains a separate product task.
- Penpot provides no 360px board. Use the same single-column hierarchy with
  the existing responsive/accessibility conventions and verify the narrow
  adaptation separately.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

At activation, re-read the final Storefront Purchase contract and current
handler response before coding. If the public command or response no longer
matches this plan, return to PLAN instead of working around the mismatch in
React.

## Scope

### In Scope

- Add a thin App Router checkout entry under the selected Offering and a
  centralized checkout URL builder.
- Add an application-owned Storefront Purchase API module that sends the exact
  contract payload, `Idempotency-Key`, non-credentialed request, and
  `AbortSignal` where applicable.
- Add narrow runtime validation for the successful Purchase envelope and
  one-time access token before UI code consumes it.
- Collect purchaser identity, choose same or distinct payer, and collect one or
  more ordered intended participants as purchaser, payer, or name-only entries.
- Translate user choices into the accepted request-local `party_ref` contract;
  opaque labels are implementation details and are never editable durable IDs.
- Use one idempotency key per frozen submit intent. A manual retry of the same
  uncertain payload reuses the key; an edited/new intent gets a new key.
- Render explicit loading, field validation, semantic validation, conflict,
  quota-unavailable, rate-limit, service-unavailable, unexpected-error, and
  success states.
- Replace the submitted form with an inline confirmation using only safe
  Purchase fields from the committed checkout response.
- Follow the inspected Penpot checkout/confirmation information hierarchy
  without reproducing future-payment actions or noncanonical domain wording.
- Invalidate the affected public Offering list/detail queries after success so
  later advisory availability reads can refresh.
- Preserve accessible labels, error summary/focus, pending state, status
  announcements, keyboard operation, and representative narrow-screen layout.

### Out of Scope

- Public Purchase GET/tracking, cancellation, token recovery/rotation, payment
  instructions, evidence submission, payment status, or Sohibul activation.
- Persisting the raw Purchase token in a URL, query string, fragment,
  `localStorage`, `sessionStorage`, cookie, IndexedDB, service worker, logs,
  analytics, error text, screenshots, or documentation.
- A reloadable confirmation route or global credential store. W5-05 owns the
  reviewed Purchase tracking and durable token-handling policy.
- Shopping Cart, multi-Offering checkout, quantity, discounts, tax, fees,
  donations, currency conversion, or alternate pricing modes.
- Automatic Party matching, contact deduplication, account creation, Party
  search, or inventing participant identity from equal names or contacts.
- Client-authoritative Event status, registration, price, quota, capacity, or
  total calculation.
- Generated clients, a form/state/schema library, a DOM/E2E framework, or a
  new shared business component package.

## Existing State

- `src/app/offerings/[offeringId]/page.tsx` and
  `src/screens/OfferingDetailPage.tsx` render the real published Offering and
  currently say checkout is future work.
- `src/routes/paths.ts` owns `/offerings`,
  `/purchase-tracking`, and the Offering detail builder.
- `src/api/catalogue.ts` and `features/catalogue/queries.ts` demonstrate the
  public DTO, runtime-guard, request-cancellation, retry, and query-key pattern.
- `createStorefrontApi` forces `credentials: "omit"` and the shared client
  serializes JSON and normalizes contract errors.
- The service worker excludes API, authentication, and business data.
- The installed frontend stack has no form, validation, DOM, or E2E dependency.
- `POST /api/public/v1/purchases` is the only implemented public Purchase
  runtime route. The response includes the access token once; there is no
  runtime token-authenticated read yet.

## Target State

- `/offerings/[offeringId]/checkout` loads the selected public Offering for
  context and presents one direct-checkout form.
- The form exposes meaningful role choices rather than raw `party_ref` fields:
  purchaser details, same/distinct payer, and ordered participant rows whose
  source is purchaser, payer, or a name-only person.
- The single-column content order follows the Penpot reference: progress
  context, buyer/payer data, intended participant data, captured Offering
  summary, minimum-data reassurance, primary submit, and help.
- The form may use the currently displayed Offering capacity for immediate
  guidance, but the server revalidates every participant count and all quota.
  Advisory availability never authorizes or permanently blocks submission.
- One submit creates one frozen payload/key intent. Automatic mutation retries
  are disabled. Only a clearly labeled manual retry reuses an uncertain
  identical intent.
- A successful response replaces the form in the same client screen with the
  Penpot-aligned success hierarchy: safe status, Purchase reference/creation
  time, exact minor-unit total/currency, captured Offering, ordered participant
  names/count, reservation expiry, and help.
- The raw token remains only in ephemeral JavaScript memory associated with the
  checkout result. It is never rendered, serialized into navigation, persisted,
  or copied into TanStack query data. Reload/navigation loss is explicit
  deferred behavior until W5-05 accepts a durable tracking policy.
- The confirmation offers only safe navigation back to the catalogue. It does
  not link to the placeholder tracking route, show a payment-instructions CTA,
  or claim that payment verification is implemented.

## Constraints

- Use Next.js App Router filesystem registration and keep the route file thin.
- Keep URL construction in `src/routes/paths.ts`.
- Keep raw transport in `src/api`; the screen must not call `fetch`.
- Use TanStack Query only for remote request lifecycle. Do not move checkout to
  a Server Action, Route Handler, middleware, or Server Component data command.
- Reuse React state, native form controls, `crypto.randomUUID()`, the existing
  API client, and `@persona-apps/ui`. Add no dependency.
- Treat form values as strings until validated. Do not perform financial
  floating-point arithmetic or reconstruct the authoritative total.
- Render server-provided content as escaped React text. Map only recognized
  error-detail keys to fields; never dump arbitrary details.
- Keep all API responses and Purchase credentials network-only under the PWA
  policy.
- Use existing semantic tokens and shared primitives for the Penpot hierarchy.
  Do not add Rubik, hardcoded page-specific brand tokens, or a broad Storefront
  redesign for two screens.
- When Penpot conflicts with product/API/security artifacts, preserve the
  canonical rule and record the visual variance rather than changing business
  behavior to match a mockup.

## Implementation Requirements

### API boundary

- Define exact request types for `PartyDeclaration`, `PartyReference`,
  `ParticipantNameInput`, and `CreatePurchaseRequest`.
- Send `POST /api/public/v1/purchases` with `Content-Type: application/json`
  through the shared client and one valid `Idempotency-Key` header.
- Preserve `credentials: "omit"`; do not add Operations cookies, CSRF, or
  authorization headers.
- Parse only the Storefront response fields. Reject malformed UUID/reference,
  status, token shape, timestamps, currency, participant sequence/count, unsafe
  integers, or inconsistent response totals with a safe API error.
- Do not expose the token from a generic formatter, error, debug object, or
  public query key.

### Role form and validation

- Purchaser is always a declaration with the fixed internal label
  `purchaser`, display name, and optional email/phone.
- Same payer maps to `{ party_ref: "purchaser" }`. A distinct payer is a
  declaration with the fixed internal label `payer` and its own fields.
- Each participant is one ordered choice: a reference to a declared
  purchaser/payer or a name-only object. Never send a mixed object.
- When payer and purchaser are shared, avoid presenting two UI choices that
  silently resolve to the same Party.
- Require at least one participant, preserve row order, and enforce contract
  string bounds. Capacity validation is helpful client feedback only.
- Preserve user-entered data after `400`, `409`, `422`, `429`, network, and
  server errors. Reset sensitive contact form state only after success.

### Idempotency and errors

- Store the frozen request object and key together as one retry intent.
- Reuse that pair only for an explicit retry after an uncertain transport or
  dependency outcome. A form edit or fresh submit creates a new key.
- Do not automatically retry checkout, quota conflict, validation failure,
  rate limit, or idempotency conflict.
- Distinguish `quota_unavailable`, `state_conflict`, and
  `idempotency_conflict` from validation and transient failures.
- Show validated request IDs and bounded `Retry-After` information when the
  shared client supplies them.

### Confirmation, accessibility, and privacy

- Keep the Penpot section order and clear success/status emphasis, but use
  semantic headings, lists, definition data, and actual buttons/links rather
  than non-interactive card affordances.
- Render a stable heading, semantic form, visible labels, input purpose where
  appropriate, per-field errors, and a focusable error summary.
- Disable only the submitted action while pending and prevent duplicate
  submission without making the page unnavigable.
- Announce pending/error/success changes through concise live regions.
- On success, move focus to the confirmation heading and show no raw contact
  data beyond what is necessary for the user to verify participant names.
- Never render, log, persist, or place the access token in the DOM or URL.
- Verify at a representative 360px width, keyboard-only navigation, text zoom,
  and normal desktop width.

## Plan Variance — Shared UI Tailwind Source

- Planned behavior: reuse `@persona-apps/ui` primitives and semantic tokens for
  the Penpot-aligned checkout.
- Unexpected evidence: browser computed styles showed shared Card utilities
  such as `p-6`, `gap-1.5`, and `bg-card` were absent. The Storefront
  `@source "../../../packages/ui/src"` path resolves below `apps/` instead of
  the repository-level `packages/ui/src` directory.
- Necessary change: correct only the Storefront source path to
  `../../../../packages/ui/src` and verify the generated styles in the browser.
- Impact: one additional planned Storefront CSS file; no token, component,
  dependency, API, or business-behavior change. The matching Operations path
  is recorded in W3-08 rather than changed during W3-07.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `apps/storefront-web/src/routes/paths.ts` | Modify | Add the centralized Offering checkout builder. |
| `apps/storefront-web/src/routes/paths.test.ts` | Modify | Prove encoded checkout URLs and active-path behavior. |
| `apps/storefront-web/src/app/offerings/[offeringId]/checkout/page.tsx` | Create | Thin App Router entry for the client checkout screen. |
| `apps/storefront-web/src/api/purchases.ts` | Create | Exact public Purchase command types, request, and guarded response mapping. |
| `apps/storefront-web/src/api/purchases.test.ts` | Create | Verify URL, method, headers, credentials, payload branches, response guards, and errors. |
| `apps/storefront-web/src/features/purchases/checkout.ts` | Create | Minimal pure role/form-to-contract validation and frozen-intent helpers. |
| `apps/storefront-web/src/features/purchases/checkout.test.ts` | Create | Verify shared/distinct roles, participant order, bounds, and retry-intent identity. |
| `apps/storefront-web/src/screens/CheckoutPage.tsx` | Create | Accessible direct-checkout form, explicit states, and inline safe confirmation. |
| `apps/storefront-web/src/screens/OfferingDetailPage.tsx` | Modify | Replace future-checkout copy with the real centralized checkout link while retaining advisory availability language. |
| `apps/storefront-web/src/styles/global.css` | Modify | Correct the proven shared-UI Tailwind source path so existing primitives render their declared spacing and semantic colors. |
| `.codex/CURRENT_STATE.md` | Modify after verified execution | Record implemented checkout UI and the still-deferred tracking/token-persistence boundary. |

Keep small one-use presentation helpers inside `CheckoutPage.tsx`. Do not add a
component directory, credential context, or generic form framework for this
single screen. Keep the one checkout mutation in `CheckoutPage.tsx`; do not add
a `features/purchases/mutations.ts` layer for one caller. Reuse the existing
catalogue query hooks for Offering context instead of creating parallel
Purchase query infrastructure.

## API, Database, and Dependency Impact

- API route: no new backend route; consume the implemented public POST only.
- OpenAPI: no planned change. Any mismatch returns to the owning contract/API
  plan before frontend work continues.
- Database/migration: none.
- Backend authorization/audit/outbox: none added by the frontend. Guest
  checkout remains role-free; the backend transaction owns outbox and replay.
- Dependency/lockfile: none.
- PWA/cache: preserve network-only API and sensitive-data policy.

## Acceptance Criteria

1. A published Offering has a centralized, directly loadable checkout route.
2. The form represents same and distinct purchaser/payer relationships and
   ordered purchaser/payer/name-only participants exactly as ADR-048 defines.
3. A valid submit sends one exact `COMMON` checkout request with no credentials,
   one idempotency key, and no client-invented business fields.
4. Manual retry of one unchanged uncertain intent reuses its key and payload;
   edited or new intent uses a new key; no automatic mutation retry occurs.
5. Client validation is accessible and helpful, while the API remains
   authoritative for Event/Offering state, price, count, capacity, quota,
   snapshots, total, and reservation.
6. `400`, `409`, `413`, `415`, `422`, `429`, `500`, `503`, cancellation, and
   malformed-success states render safe, distinct outcomes without losing
   entered values.
7. Success renders the committed non-secret Purchase facts and exact captured
   minor-unit amount without recalculation.
8. The one-time access token is absent from URLs, rendered markup, logs,
   persistent browser stores, query data, service-worker caches, and errors.
9. The screen does not claim payment, tracking, cancellation, token recovery,
   account, or multi-Offering behavior that is not implemented.
10. Checkout and confirmation preserve the inspected Penpot information
    hierarchy while explicitly omitting its unsupported payment CTA,
    verification promise, and active-Sohibul implication.
11. Focus, labels, error summary, live status, keyboard behavior, text zoom, and
    representative 360px/862px layouts are verified.
12. Existing public Event/Offering pages and PWA exclusions remain intact.
13. No new dependency, generated client, backend workaround, or duplicated
    business authority is introduced.

## Testing

- Pure tests for role mapping, required/bounded values, participant ordering,
  capacity guidance, frozen intent/key reuse, and new-intent detection.
- Injected-fetch API tests for the exact URL, POST body, headers, omitted
  credentials, success envelope, token shape, safe integer/timestamp guards,
  cancellation, and normalized error classes.
- Path tests for encoded Offering checkout URLs and direct route ownership.
- Structural tests with the existing React/Vitest capability for semantic
  labels, safe confirmation fields, and explicit absence of token/payment/
  tracking claims.
- Structural/browser checks for the Penpot-backed section order, status/
  reference/summary/help content, and explicit absence of the deferred payment
  action and verification copy.
- Real-browser smoke against the real local API and disposable PostgreSQL for
  shared payer, distinct payer, name-only participant, validation, quota
  conflict, same-intent retry, successful confirmation, keyboard flow, 360px
  layout, and token/storage/network inspection.

Do not add a DOM or E2E framework merely to shallow-render this screen. W3-09
may close a demonstrated interaction gap with the smallest existing-capability
test or recorded browser proof.

## Verification

Run from the repository root:

```bash
pnpm --filter @persona-apps/storefront-web test
pnpm --filter @persona-apps/storefront-web typecheck
pnpm --filter @persona-apps/storefront-web lint
pnpm --filter @persona-apps/storefront-web build
make validate
docker compose -f infrastructure/compose.yaml config
```

Run the real browser/API/PostgreSQL checkout smoke described above. Report
focused frontend checks and unrelated repository-wide failures separately; do
not claim a mock-only or build-only result completes the task.

## Deliverables

- One real direct-checkout route and Offering-detail entry point.
- Exact Storefront Purchase API boundary with guarded response mapping.
- Explicit-role form, stable manual retry intent, and safe inline confirmation.
- Focused tests plus real API/browser evidence.
- Current-state update and canonical task archival only after execution is
  approved, implemented, and verified.

## Risks, Decisions, and Deferred Work

- Accepted: one checkout contains one Offering and prices per intended
  participant; the browser never recalculates the authoritative total.
- Accepted: request-local Party references express role reuse; equal names or
  contacts do not.
- Accepted task boundary: confirmation is inline. A separate URL would either
  lose the response or tempt token-bearing navigation before the tracking API
  and credential policy exist.
- Penpot is adopted for composition, not as a transport or lifecycle contract:
  payer input remains explicit, participants remain intended rather than
  activated, exact minor units remain authoritative, and its payment CTA/copy
  is deferred.
- Deferred risk: navigation/reload loses the ephemeral Purchase token. W5-05
  must accept a safe durable/resumable policy before any persistence is added.
- Deferred: payment instructions/evidence/status (Week 4), completed Storefront
  flow and tracking (Week 5), and all non-`COMMON` channels.

## Execution Report

### Implemented

- Added `apps/storefront-web/src/api/purchases.ts` and its focused tests for
  the exact public `POST /api/public/v1/purchases` contract, omitted
  credentials, bounded response guards, and safe error mapping.
- Added `apps/storefront-web/src/features/purchases/checkout.ts` and tests for
  purchaser declarations, same/distinct payer mapping, ordered purchaser /
  payer / name-only intended participants, bounded strings, and frozen
  idempotency intents.
- Added the thin App Router entry
  `apps/storefront-web/src/app/offerings/[offeringId]/checkout/page.tsx` and
  `apps/storefront-web/src/screens/CheckoutPage.tsx`.
- Updated `src/routes/paths.ts`, route tests, `OfferingDetailPage.tsx`, and
  the proven Tailwind shared-UI source path in `global.css`.
- The success state is inline and uses only safe Purchase facts. The one-time
  access token remains ephemeral JavaScript state and is not rendered,
  serialized into navigation, persisted, logged, placed in query data, or
  cached by the service worker.
- Applied the authorized Penpot copy corrections on pages `13 · Storefront ·
  04 Checkout Participant` and `16 · Storefront · 07 Purchase Confirmation`.
  No Penpot page was deleted; all 28 pages remain in the file.

### Verified

- Storefront focused suite: 10 files, 31 tests; typecheck, Oxlint, and
  production build passed, including the generated dynamic checkout route.
- `make lint typecheck test build compose-check` passed across both frontends
  and the API. The database-enabled Go suite passed against a disposable
  PostgreSQL database, and `git diff --check` passed.
- Browser evidence covered shared/distinct payer and name-only participant
  submission, quota conflict with preserved form values, safe success without
  payment/tracking controls, keyboard/focus behavior, 44px submit target, and
  a 360px viewport with no horizontal overflow.
- Penpot screenshots verified the updated checkout and confirmation hierarchy;
  the live page tree still lists every original page.
- The tracker was updated narrowly at `Tracker!I23` and `Tracker!N23`, then
  read back as `W3-07 = Done`.

### Assumed

- The existing public catalogue query remains the source of Offering context;
  the API remains authoritative for eligibility, price, totals, capacity,
  quota, snapshots, reservation, reference, and token generation.
- Indonesian Penpot copy is treated as visual hierarchy guidance because the
  Storefront has no accepted localization workflow; runtime copy remains the
  existing English product language.

### Deferred / Risks

- `make validate` is not green in this run because of pre-existing repository
  Go formatter drift; unrelated Go files were not reformatted.
- Payment instructions/evidence/status, Purchase tracking, cancellation,
  token recovery, Sohibul activation, non-`COMMON` channels, and Operations
  Purchase list/detail remain out of scope. W3-08 and W3-09 are still open.
- Reload/navigation loses the ephemeral token until W5-05 accepts a reviewed
  durable tracking policy.

## Final Report Requirements

The execution report must list exact files, role-to-payload behavior,
idempotency retry behavior, token handling, API/browser evidence, accessibility
checks, PWA/storage inspection, plan variance, and all deferred behavior. It
must distinguish implemented, verified, assumed, and deferred behavior and
must not call Purchase tracking or payment complete.
