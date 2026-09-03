# Task: W4-06 Add Operations Payment-Verification Queue and Detail

## Status

Ready for plan review on 2026-09-03. Planning-only; live tracker remains
`Backlog`. BUILD depends on verified W4-03 commands and an approved private
evidence retrieval contract from W4-02.

## Tracker

- Week: Week 4
- Epic: Payment & Activation
- Application: Operations
- Category: Frontend
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Add Operations payment-verification queue and detail.
- Dependencies: W4-03.
- Acceptance summary: authorized operators can find submitted evidence and make
  a verification decision through Operations.

## Objective

Replace the `/payment-verification` placeholder with a permission-aware,
mobile-readable Operations workflow that lists submitted Payments, opens an
exact detail, retrieves private evidence through the Go API, and submits verify
or reasoned reject commands without treating frontend state as authoritative.

## Context

The Operations OpenAPI drafts Payment list/detail and verify/reject routes.
Current backend source implements none of them, and the contract exposes only
evidence metadata—there is no authorized byte retrieval route. Operations
already provides OIDC session state, memory-only CSRF, permission presentation,
credentialed API transport, TanStack Query patterns, centralized paths, shared
UI primitives, and Purchase list/detail screens. The payment page is a
13-line placeholder.

No visual design artifact for Payment verification was supplied. Use the
existing Operations shell and information hierarchy; do not invent dashboard
KPIs, bulk actions, reconciliation totals, or provider fields.

## Source of Truth

- W4-02/W4-03 approved contracts and verified runtime;
- canonical PRD Payment, Operations, security, privacy, and audit sections;
- Product Map Payment Verification Queue and Operations ownership;
- Architecture frontend/API/auth/cache boundaries and ADR-002, ADR-010,
  ADR-018, ADR-019, ADR-024, ADR-038, ADR-039, ADR-041, ADR-044, and ADR-047;
- `docs/security/AUTHENTICATION.md` and `PERMISSIONS.md`;
- `contracts/openapi/operations.yaml` Payment schemas/routes;
- current Operations session, API, Purchase query/screen, layout, path, PWA,
  and test patterns;
- `.codex/CURRENT_STATE.md`, roadmap, and live tracker W4-06 row.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

At activation, compare the final W4-03 handler JSON and W4-02 evidence read
route with OpenAPI. Fix any mismatch at its owning backend/contract boundary;
the frontend must not use database/storage details or fake missing fields.

## Scope

### In Scope

- Keep `/payment-verification` as the centralized list route and add one encoded
  Payment detail route/builder under that path.
- Add an Operations Payment API module with strict runtime response guards for
  list, detail, private evidence response, verify, and reject.
- Add feature-owned query keys/hooks for status-filtered cursor pages and
  Payment detail, plus mutations for decisions.
- Require `payment.read` for list/detail metadata. Show decision controls and
  retrieve evidence bytes only with the approved `payment.verify` permission.
- Default the queue to `SUBMITTED`; support only contract-defined status filter,
  limit, and opaque cursor behavior.
- Present Payment reference, Purchase link, payer summary, amount/currency,
  method, status, submitted/decision times, rejection reason, and safe evidence
  metadata exactly as contracted.
- Retrieve evidence from the authenticated Go API as a short-lived in-memory
  blob/object URL, revoke it on replacement/unmount, and never persist/cache it.
- Render safe inline preview for supported browser-renderable content and a
  controlled download fallback with server-provided safe headers.
- Generate one idempotency key per verify/reject intent, reuse it only for that
  retry, send current in-memory CSRF, and require a reason for reject.
- After success, invalidate exact Payment list/detail and affected Purchase
  queries; do not optimistically mark financial state.
- Provide distinct loading, empty, unauthorized, forbidden, not-found,
  validation, conflict, unavailable, decision-pending, and success states.
- Add focused API/query/path/component/PWA/accessibility tests and real browser
  smoke over the real Go API and disposable PostgreSQL.

### Out of Scope

- Client-side authorization, direct PostgreSQL/object storage access, public
  evidence URLs, presigned-provider fields, or caching evidence in TanStack/PWA/
  browser storage.
- Bulk decisions, keyboard shortcuts that bypass confirmation, payment edits,
  refunds, adjustments, provider reconciliation, exports, dashboard totals,
  polling/SSE, or audit-log UI.
- Free-text/payer/contact/date-range search absent from the contract.
- A table/form/router/preview library or generated API client.
- W6 full commerce control plane and W13 realtime projections.

## Existing State

- `src/app/payment-verification/page.tsx` already maps the route to a placeholder
  screen, and navigation/path constants already exist.
- Operations transport sends `credentials: include`; mutations can reuse the
  existing session/CSRF and normalized error boundaries.
- Purchase list/detail demonstrates exact DTO guards, cursor queries, responsive
  desktop/card presentation, private-root cache cleanup, and no CSRF on GET.
- Payment OpenAPI has list/detail/decision contracts but no evidence byte route.
- Customer Service has `payment.read` but not `payment.verify`; Finance and
  Operations Manager have both.

## Target State

- Authorized readers can inspect a bounded Payment page/detail but cannot
  decide without `payment.verify`.
- Authorized reviewers can view private evidence for the selected Payment and
  submit exactly one confirmed decision intent.
- Server response replaces stale client assumptions; list/detail/Purchase state
  refreshes after the transaction commits.
- Logout/private-root cleanup removes every Payment query and any in-memory
  evidence URL is revoked immediately.

## API and Security Requirements

- GET list/detail send credentials but no CSRF or idempotency header.
- Evidence GET uses the approved least-privilege permission, no-store,
  nosniff-safe content type, bounded content length, and safe content disposition.
- POST verify/reject send credentials, exact Origin through the browser, CSRF,
  and idempotency key. Reject sends only the bounded reason.
- Runtime guards reject unknown/internal fields, unsafe integers, bad UUID/time/
  enum values, and exposure of reference/digest/provider fields.
- UI never logs, analytics-captures, screenshots, stores, or service-worker
  caches evidence, session, CSRF, or idempotency values.

## Planned File Changes

- `apps/operations-web/src/api/payments.ts` and tests;
- `apps/operations-web/src/features/payments/queries.ts` and focused tests;
- `apps/operations-web/src/screens/PaymentVerificationPage.tsx` and tests;
- thin `src/app/payment-verification/[paymentId]/page.tsx` and centralized path
  builder/tests if detail uses its own route;
- existing session/private-query cleanup only if the current prefix does not
  already cover Payment keys;
- Operations service-worker tests to retain network-only evidence/payment data;
- W4-02/W4-03 backend/contract files only for a verified missing route/field.

No new frontend dependency is planned.

## Acceptance Criteria

1. `/payment-verification` loads only bounded contract-backed Payment pages and
   defaults to the actionable `SUBMITTED` queue.
2. `payment.read` grants list/detail metadata; only `payment.verify` exposes
   evidence bytes and decision controls, with backend enforcement authoritative.
3. Detail uses the Operations contract and reveals no storage reference/digest,
   token, idempotency, audit, SQL, or unrelated Party/contact data.
4. Evidence is loaded on demand, kept only in memory, rendered/downloaded with
   safe headers, revoked promptly, and never enters caches/storage/logs.
5. Verify and reject use frozen intent keys, current CSRF, accessible confirmation,
   reason validation, and no optimistic financial transition.
6. Success renders the server's Payment/Purchase/quota/activation outcome and
   invalidates affected queries; conflict/auth/unavailable states remain clear.
7. Filters/cursors match the server exactly; no unsupported search/KPI/bulk/
   realtime behavior is implied.
8. Logout and direct-return paths remove private data and evidence; 360px and
   desktop layouts have no inaccessible overflow or action loss.
9. API, route, component, PWA, accessibility, production build, and real browser
   tests pass against real authorization and PostgreSQL state.

## Testing

- API tests for exact paths/query/credentials/headers/bodies, response allowlists,
  binary headers, cancellation, errors, and unsafe values.
- Query tests for key isolation, permission gating, invalidation, no automatic
  decision retry, and private-root logout cleanup.
- Component tests for list/detail/evidence/decision states, reason requirement,
  focus management, keyboard use, and absence of internal fields.
- PWA tests prove every Payment/evidence/auth/API request is network-only.
- Browser smoke covers signed-out, payment.read-only, payment.verify, evidence
  open/revoke, verify, reject, replay/conflict, logout, and narrow layout.

## Verification

```bash
pnpm --filter @persona-apps/operations-web test
pnpm --filter @persona-apps/operations-web typecheck
pnpm --filter @persona-apps/operations-web lint
pnpm --filter @persona-apps/operations-web build
make validate
docker compose -f infrastructure/compose.yaml config
```

Also run W4-02/W4-03 focused Go/PostgreSQL tests and the production Next.js
proxy browser flow. Mock-only success does not satisfy acceptance.

## Risks and Decision Gates

- Approve the evidence retrieval endpoint and whether bytes require
  `payment.verify` (recommended) or `payment.read` before frontend work.
- Browser preview support differs by media type. Provide a download fallback;
  do not add a PDF/image viewer dependency.
- Payment list contract currently filters only status. Rich search and totals
  belong to W6 unless a separate backend contract is approved.
- Evidence access is sensitive even when Payment metadata is readable; frontend
  hiding cannot substitute for backend permission checks.

## Deliverables and Final Report

Deliver contract-guarded Payment API/query code, responsive queue/detail,
private evidence access, safe decision mutations, focused/full verification,
and browser evidence. Report permissions, headers, query/invalidation behavior,
evidence lifecycle, tested states, variances, and deferred W6/realtime work.
