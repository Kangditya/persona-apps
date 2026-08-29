# Task: W1-05 Draft Storefront and Operations OpenAPI

## Status

Inactive draft. Activate only after W1-03 and W1-04 are executed and current
source is revalidated.

## Tracker

- Workstream: Platform Safety
- Estimate: 4 hours
- Dependencies: W1-03; W1-04
- Tracker objective: Draft Storefront and Operations OpenAPI endpoints and
  error shapes.

## Objective

Replace both endpoint-free OpenAPI placeholders with separate Phase 1 public
and operations contracts derived from resolved schema, lifecycle, permission,
audit, and idempotency decisions.

## Existing Foundation to Reuse

- Keep `contracts/openapi/storefront.yaml` and `operations.yaml` separate.
- Use architecture route prefixes `/api/public/v1` and `/api/operations/v1`.
- Reuse UUID public identifiers, human-readable references, request IDs, and
  ADR-041 command idempotency.
- Do not generate clients until a real vertical slice consumes the contracts.

## Storefront Contract

Define these operations:

```text
GET  /api/public/v1/events/active
GET  /api/public/v1/events/{event_id}/offerings
GET  /api/public/v1/offerings/{offering_id}
POST /api/public/v1/purchases
GET  /api/public/v1/purchases/{purchase_id}
POST /api/public/v1/purchases/{purchase_id}/cancel
POST /api/public/v1/purchases/{purchase_id}/payment-evidence
```

- Purchase creation accepts purchaser, payer, one Offering, and intended
  participants; it returns Purchase data plus the one-time access token.
- Purchase read, cancel, and evidence submission require the Purchase Bearer
  token and never expose its stored hash.
- Purchase creation, cancellation, and evidence submission require
  `Idempotency-Key`.
- Evidence submission is `multipart/form-data`: declared amount/currency plus
  one JPEG, PNG, or PDF file up to 10 MiB.
- Availability reads are snapshots; checkout may still return quota conflict.

## Operations Contract

Define these operation groups:

```text
GET  /api/operations/v1/auth/login
GET  /api/operations/v1/auth/callback
GET  /api/operations/v1/auth/session
POST /api/operations/v1/auth/logout

GET/POST      /api/operations/v1/events
GET/PATCH     /api/operations/v1/events/{event_id}
POST          /api/operations/v1/events/{event_id}/{publish|activate|suspend|close|archive}
GET/POST      /api/operations/v1/events/{event_id}/offerings
GET/PATCH     /api/operations/v1/offerings/{offering_id}
POST          /api/operations/v1/offerings/{offering_id}/{publish|unavailable|archive}
GET           /api/operations/v1/purchases
GET           /api/operations/v1/purchases/{purchase_id}
GET           /api/operations/v1/payments
GET           /api/operations/v1/payments/{payment_id}
POST          /api/operations/v1/payments/{payment_id}/{verify|reject}
GET           /api/operations/v1/participants
GET           /api/operations/v1/participants/{participant_id}
GET           /api/operations/v1/dashboard/summary
GET           /api/operations/v1/audit
```

- Every operation declares its W1-04 permission.
- Retry-sensitive status commands require `Idempotency-Key`; reads and logout
  do not.
- Payment verification/rejection requires a reason where policy requires it
  and returns the resulting Payment, Purchase, quota, and activation summary.
- Lists use bounded cursor pagination and deterministic default sorting.
- Operations responses do not return raw session hashes, CSRF hashes, Purchase
  token hashes, OIDC tokens, or public object-storage URLs.

## Common Headers and Errors

- Accept optional `X-Request-ID`; validate a bounded safe format or generate a
  UUID. Return it as `X-Request-ID` on every response.
- `Idempotency-Key` is a bounded opaque string and appears only on named
  retry-sensitive commands.
- Use this error envelope:

```json
{"error":{"code":"quota_unavailable","message":"...","request_id":"...","details":{}}}
```

- Define at least: `invalid_request`, `validation_failed`, `unauthenticated`,
  `forbidden`, `not_found`, `state_conflict`, `quota_unavailable`,
  `idempotency_conflict`, `evidence_too_large`, `unsupported_media_type`,
  `rate_limited`, `internal_error`, and `service_unavailable`.
- Map errors consistently to `400`, `401`, `403`, `404`, `409`, `413`, `415`,
  `422`, `429`, `500`, and `503` as applicable.
- Internal error details never expose SQL, stack traces, credentials, storage
  keys, operator permissions, or personal data not needed by the caller.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `contracts/openapi/storefront.yaml` | Modify | Define the Phase 1 guest catalogue, checkout, tracking, cancellation, and evidence API. |
| `contracts/openapi/operations.yaml` | Modify | Define auth, administration, verification, participant, dashboard, and audit APIs. |
| `docs/ARCHITECTURE.md` | Modify | Link route, error, request-ID, auth, and idempotency conventions. |
| `.codex/CURRENT_STATE.md` | Modify | Record drafted contracts without claiming runtime endpoints. |
| `.codex/TASK.md` | Create, execute, archive | Preserve the activated task and final review. |

## Acceptance Criteria

1. Both OpenAPI 3.1 documents parse and contain no placeholder `paths: {}`.
2. Public and operations schemas are separate and expose no privileged fields
   across the public boundary.
3. Every command maps to W1-01 lifecycle rules and W1-03 schema identifiers.
4. Every operations endpoint declares a W1-04 permission.
5. Retry-sensitive commands declare request ID and idempotency behavior,
   including replay and payload-mismatch responses.
6. Evidence type/size and common error envelopes are reusable components.
7. No runtime code, generated client, frontend, dependency, migration, or
   infrastructure change occurs.

## Verification

- Validate both documents with the repository's OpenAPI/format checks.
- Trace checkout, evidence, verify, reject, cancel, and replay success/error
  paths against W1-01 through W1-04.
- Search Storefront schemas for operations-only fields.
- Run `make validate`.

## Risks and Deferred Work

- Saving, giveaway, livestock, allocation, refund, provider webhook, and
  distribution endpoints are deferred.
- Evidence download/view policy is deferred until a private object-storage
  access pattern is selected.
- Generated clients wait for a consuming vertical slice.

## Final Review Requirements

Distinguish contracted behavior, parser/validation results, assumptions,
deferred endpoints, exposure risks, and the next dependency-ready task.
