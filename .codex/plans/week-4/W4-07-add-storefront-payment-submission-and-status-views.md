# Task: W4-07 Add Storefront Payment Submission and Status Views

## Status

Ready for plan review on 2026-09-03, blocked for BUILD by the public status and
Purchase-token handoff decisions below. Planning-only; live tracker remains
`Backlog`.

## Tracker

- Week: Week 4
- Epic: Payment & Activation
- Application: Storefront
- Category: Frontend
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Add Storefront payment submission and status views.
- Dependencies: W4-01, W4-02, and W4-03.
- Acceptance summary: a Purchase-token holder can submit evidence and view
  current Payment and Purchase status without internal fields.

## Objective

Add the smallest safe Storefront continuation from successful checkout to
manual-transfer evidence submission and current status. Keep the one-time
Purchase token ephemeral, send it only in Authorization headers, use the public
contract exclusively, and never duplicate amount/quota/eligibility rules in
the browser.

## Context

Checkout currently validates the returned Purchase token and deliberately
discards it before TanStack Mutation state; no implemented feature consumes it.
The Storefront has no Payment API/query/form or runtime Purchase tracking route.
OpenAPI drafts evidence submission and Purchase GET, but current Go routing
implements neither. `PublicPurchase` exposes Purchase status and `PublicPayment`
is returned only by submission; there is no contract-backed way to refresh the
current Payment after Operations review.

W5-04 later completes payment instructions/evidence UX and W5-05 owns resumable
tracking by safe reference/authentication policy. W4-07 should provide the
working immediate vertical slice without inventing token recovery, durable
browser credential storage, or bank instructions.

## Source of Truth

- W4-01 through W4-03 approved runtime/contracts;
- PRD Common Purchase, Storefront principles, Payment, security, privacy, and
  data-integrity sections;
- Product Map Storefront Payment Interaction and Purchase Tracking boundaries;
- Architecture Storefront/Go API/auth/cache/PWA rules and ADR-002, ADR-010,
  ADR-018, ADR-038, ADR-039, ADR-041 through ADR-045, ADR-047, and ADR-050;
- `docs/security/AUTHENTICATION.md` Purchase-token contract;
- `contracts/openapi/storefront.yaml`;
- current Storefront checkout API/form/screen, query client, routes, PWA, shared
  UI, and W3-07/W3-09 token-retention evidence;
- roadmap distinctions between W4-07, W5-04, and W5-05.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Before BUILD, approve how current Payment state is read and how the checkout
token is handed directly to the immediate payment flow without entering DOM,
URL, TanStack Query/mutation cache, browser storage, service worker, logs, or
analytics.

## Scope

### In Scope

- Extend the checkout continuation so a newly returned token is passed through
  a one-shot in-memory callback/ref to the payment step while the safe Purchase
  result remains the only mutation-cache value.
- Add a Storefront Payment API module that builds multipart evidence requests,
  sends a frozen idempotency key and Bearer token, and strictly validates the
  public response.
- Add the approved token-scoped current-status query/runtime contract needed to
  return both minimized Purchase and current Payment state after review.
- Keep status/evidence state scoped to the identified Purchase and ephemeral
  token; lose access safely on reload until W5-05 supplies a resumable policy.
- Render amount/currency from the captured Purchase; let the user select one
  JPEG/PNG/PDF up to 10 MiB; treat client validation as advisory and preserve
  backend errors.
- Generate one key per explicit submission intent, reuse it only for retry of
  the same file/amount/currency, and require a new intent after rejection.
- Render submitted, verified, rejected, Purchase pending/paid/eligible, quota
  conflict, validation, unsupported/oversized, auth, rate-limit, unavailable,
  retry, and token-lost states.
- Poll only if the approved status contract specifies a bounded interval and
  stop conditions; otherwise provide explicit user refresh. Never infer state
  from elapsed time or cached checkout data.
- Ensure all Payment/Purchase/API/auth data remains network-only in the PWA.
- Add focused API, intent, screen, route, accessibility, PWA, and real-browser
  tests.

### Out of Scope

- Persisting the Purchase token in local/session storage, cookies, URL/query/
  fragment, IndexedDB, TanStack cache, service worker, logs, analytics, or DOM.
- Token recovery/rotation, account login, durable tracking, safe-reference
  lookup, or reload resumption; W5-05 owns the accepted policy.
- Bank/account/payment instructions, polished completion UX, or broad status
  state coverage beyond the working slice; W5-04/W5-03 own completion.
- Frontend-authoritative amount, quota, Payment, Purchase, or eligibility logic.
- Operations fields, evidence references/digests, internal Party/operator data,
  audit/outbox/idempotency data, refunds, partial payments, or providers.
- Upload/query/form dependencies, background sync, offline submission, or API
  caching.

## Existing State

- Checkout API currently returns only the safe `CreatedPurchase` after validating
  and discarding the raw token.
- Checkout screen renders inline confirmation and explicitly says Payment and
  tracking are unavailable.
- `purchase-tracking` remains a placeholder and is owned by W5-05.
- Storefront service worker already excludes every API/auth/business request.
- Public contract GET Purchase lacks current Payment summary; submission replay
  would return the original `SUBMITTED` response forever and is not a status
  refresh mechanism.

## Target State

- A successful checkout can immediately continue to evidence submission using
  a token that exists only in the active client boundary and Authorization
  header.
- A successful upload renders safe Payment/Purchase status and can refresh it
  through one approved scoped query until the page reloads/token is lost.
- Verification/rejection updates come only from the Go API. Rejected evidence
  remains historical and a new submission intent can be created.
- Reload/direct navigation shows a clear credential-required state and does not
  attempt recovery or leak prior data.

## Public Contract Requirements

- Keep evidence POST exactly multipart with amount, currency, one file,
  `Authorization`, `Idempotency-Key`, and request ID behavior.
- Define one minimized current-status response. Recommended minimal shape:
  existing `PublicPurchase` plus an optional current/latest `PublicPayment`
  summary, with precise selection semantics when rejected history exists.
- Do not expose evidence metadata/reference, payer contact, operator, ledger,
  reservation internals, audit, outbox, or token hash.
- Authenticate token before lookup/replay. Return no-store and stable 401/403/
  404/409/422/429/503 envelopes.
- Do not overload submission replay as a query and do not implement W5-05
  reference lookup in this task.

## Planned File Changes

- `apps/storefront-web/src/api/payments.ts` and tests;
- `apps/storefront-web/src/features/payments/submission.ts` and focused tests;
- the existing Checkout screen/API boundary for one-shot token handoff and
  immediate Payment/status presentation;
- a feature query only if the approved status refresh uses TanStack Query, with
  the token supplied transiently outside the query cache/key;
- Storefront PWA tests and existing route/path files only if an approved new
  route is necessary;
- `contracts/openapi/storefront.yaml` and narrow Go public query/handler files
  for the approved current-status gap;
- no new dependency or generated client.

## Acceptance Criteria

1. A newly created Purchase can submit exactly one evidence intent with its
   ephemeral Bearer token and frozen idempotency key.
2. Raw token never enters rendered content, URL, query/mutation cache, browser
   storage, service worker, logs, analytics, errors, or screenshots.
3. Form accepts only the documented media/size/amount/currency shape and renders
   backend rejection without claiming client validation is authoritative.
4. Retry of the same intent reuses the key; editing file/amount/currency creates
   a new intent and never silently reuses the old key.
5. The approved scoped query returns current minimized Payment and Purchase
   status after Operations review; submission replay is not used for refresh.
6. Reload/token loss removes access and private state safely; no recovery or
   durable tracking is claimed before W5-05.
7. Verified/rejected/pending/paid/eligible and validation/conflict/auth/rate/
   unavailable states are explicit, accessible, and mobile-readable.
8. No internal evidence, storage, operator, ledger, audit, reservation, token,
   or idempotency fields are exposed.
9. Payment/evidence/status API requests remain network-only and are never queued
   offline.
10. Focused frontend, public API/PostgreSQL, production build, 360px/browser,
    and repository validation pass.

## Testing

- API tests for exact multipart parts/headers/path, token non-retention,
  idempotency intent, strict response allowlists, status query, cancellation,
  and normalized errors.
- Feature tests for file/media/size, edited-versus-retry intent, rejection/new
  attempt, and token-lost state.
- Screen tests for accessible labels/status/focus, no raw token/internal fields,
  and safe re-render/logout/reload behavior.
- PWA tests prove GET/POST Payment/Purchase and evidence bytes are network-only.
- Browser smoke uses real checkout, ephemeral handoff, evidence upload, Operations
  decision in the companion app, status refresh, reload loss, and 360px layout.

## Verification

```bash
pnpm --filter @persona-apps/storefront-web test
pnpm --filter @persona-apps/storefront-web typecheck
pnpm --filter @persona-apps/storefront-web lint
pnpm --filter @persona-apps/storefront-web build
make validate
docker compose -f infrastructure/compose.yaml config
```

Also run W4 public Go/PostgreSQL integration and secret-exposure checks. A
mock-only upload/status result is insufficient.

## Decision Gates and Risks

- **Current status:** define the public response/selection semantics for current
  Payment alongside Purchase status without consuming W5-05 reference lookup.
- **Token handoff:** approve an ephemeral one-shot checkout-to-payment channel.
  Any durable browser storage conflicts with the current security boundary and
  needs a separate authentication/retention decision.
- **Polling:** default to explicit refresh unless a bounded cadence/stop policy
  is accepted; Payment state is sensitive and non-authoritative in cache.
- **Zero price/name-only participant:** surface backend conflict honestly; do
  not mask unresolved W4-01/W4-05 policy in the UI.
- The W3-09 token-discard behavior was correct while no consumer existed. W4-07
  may revise the API return path only narrowly, keeping the token out of
  TanStack and every persistent/rendered surface.

## Deliverables and Final Report

Deliver the ephemeral credential handoff, public Payment API/intent logic,
submission/status UI, contract/runtime status query, network-only proof, and
real-browser flow. Report token lifetime/locations, exact requests/responses,
retry keys, states, accessibility, browser sizes, validation, variances, and
deferred W5 tracking/instructions.
