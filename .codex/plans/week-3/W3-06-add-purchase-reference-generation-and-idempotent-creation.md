# Task: W3-06 Add Purchase Reference Generation and Idempotent Creation

## Status

Completed and verified on 2026-08-21. The executed task is archived at
`.codex/archive/2026-08-21-TASK-w3-06-add-purchase-reference-generation-and-idempotent-creation.md`;
no commit or push was requested.

## Tracker

- Week: Week 3
- Epic: Common Purchase
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Add Purchase reference generation and idempotent creation.
- Tracker dependencies: W1-06 and W3-03.
- Execution dependencies: W3-01 through W3-05; this task composes their outputs
  into the public checkout command.
- Acceptance summary: a safe human-readable reference is generated and retries
  cannot create duplicate Purchases.

## Objective

Publish the contracted guest Common Purchase checkout as one atomic,
idempotent command. Generate the approved non-secret Purchase reference and
one-time access token, resolve explicit Party roles, capture snapshots, reserve
quota, persist history/outbox/replay, and return the original encrypted replay
response for a valid retry.

## Source of Truth

- `docs/PRD.md`, Phase 1 direct checkout, guest access, security, and money
  rules;
- `docs/PRODUCT_MAP.md` and `docs/ARCHITECTURE.md`;
- `docs/DECISIONS.md`, especially ADR-013, ADR-014, ADR-018 through ADR-020,
  ADR-024, ADR-033, ADR-038, and ADR-041 through ADR-045;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- `docs/security/AUTHENTICATION.md` and `docs/security/PERMISSIONS.md`;
- `docs/database/ERD.md` and migrations `0002`, `0005`, and `0006`;
- `contracts/openapi/storefront.yaml`;
- the existing `database`, `idempotency`, `outbox`, `httpx`, CORS, request-ID,
  recovery, and public rate-limit source;
- approved and verified W3-01 through W3-05 outputs.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate only after all dependencies and decisions are accepted. Do not commit
or push unless requested.

## Resolved Decisions

### Purchase reference format

ADR-050 sets `QRB-<event-year>-<16 uppercase unpadded Base32 characters>` from
10 `crypto/rand` bytes, with at most five whole-transaction retries on a unique
reference collision. The UUID remains the canonical API ID; the reference is a
display/support identifier, not authentication or an idempotency key.

### Guest idempotency caller scope

ADR-050 sets the stable scope to the literal
`storefront.purchase.create.guest`. It does not use a request ID, mutable IP,
browser header, Party contact, or Purchase token as an invented identity.

### Checkout replay retention

ADR-050 sets `Command.Retention == 0` to durable replay (`expires_at IS NULL`).
The executor must retain the encrypted result until a separately approved
cleanup/key-rotation policy exists; a copied 24-hour Event/Offering window is
not used.

W3-02's explicit role-linking payload and W3-04's total calculation are
accepted dependencies; this plan does not reopen them.

## Existing State

- `POST /api/public/v1/purchases` is fully drafted in the Storefront OpenAPI
  contract but has no runtime route.
- The response returns a raw Purchase token only at checkout; later reads do
  not return it. The server stores only its SHA-256 hash.
- The platform idempotency executor already namespaces commands, hashes
  canonical input, stores status/headers/body, and AES-GCM encrypts successful
  replay bodies. It must be reused because the checkout response contains the
  raw token.
- Public middleware already provides request IDs, recovery, no-store behavior,
  proxy-aware client handling, and rate limiting, but current CORS only permits
  GET and does not allow `Content-Type` or `Idempotency-Key` for checkout.
- `outbox_events`, Purchase history, all Purchase/Party/snapshot/reservation
  tables, and their constraints already exist.

## Scope

### In Scope

- Implement and register `POST /api/public/v1/purchases` with the existing
  public middleware and contract envelope.
- Require and validate `Content-Type: application/json` and
  `Idempotency-Key`; use bounded strict JSON decoding with unknown/trailing
  content rejected.
- Resolve/create Party records and role links exactly as approved by W3-02,
  without inferred contact deduplication.
- Generate a cryptographically random 32-byte opaque Purchase access token,
  encode it with the accepted base64url representation, store only SHA-256,
  and expose raw token only in the 201/replayed checkout response.
- Generate the approved human-readable unique Purchase reference with bounded
  collision handling.
- In one transaction, validate/lock Event and Offering, capture snapshots,
  create Parties/relationships/Purchase/participants/status history, reserve
  quota, write accepted outbox events, and store encrypted replay response.
- Return `201` for the first success and the same stored status/body for a
  valid same-key/same-request replay; return `409 idempotency_conflict` when the
  same namespace/key is reused for different validated input.
- Extend public CORS/preflight only for POST checkout and the exact required
  headers; do not enable credentials or wildcard private headers.
- Add end-to-end handler/transaction/replay/rollback/security tests.

### Out of Scope

- Public Purchase tracking, cancellation, evidence upload, token rotation,
  payment verification, activation, or expiry worker.
- Operations commands or frontend forms/screens.
- Raw-token recovery outside encrypted idempotency replay.
- Sequential database IDs or references that expose volume, token/reference
  reuse, deterministic tokens, or storing/logging plaintext tokens.
- New idempotency, encryption, rate-limit, transaction, outbox, repository, or
  command-bus frameworks.
- Automatic Party matching/deduplication.

## Command Order and Atomicity

Trust-boundary order:

1. Apply public recovery, request ID, no-store, proxy handling, rate limiting,
   and exact CORS policy.
2. Validate method/media type/body size, require the idempotency key, strictly
   decode one JSON object, and validate the approved semantic request.
3. Build the ADR-041 namespace with the approved stable guest scope and hash
   canonical validated input, not raw JSON formatting.
4. Start the existing idempotency executor transaction/replay flow.
5. For a new command, lock Event then Offering and run W3-05 eligibility/quota
   checks.
6. Resolve explicit Parties/roles, create the approved reference and token/hash,
   build snapshots/total, and persist Purchase, participants, initial history,
   and attempt-1 reservation.
7. Write `PurchaseReserved` and `PurchasePendingPayment` outbox events with
   minimized payloads; write no guest audit row unless an assisted privileged
   command is separately introduced.
8. Build the contract response, encrypt/store its replay record with approved
   retention, and commit once.

Any failure in steps 5–8 must roll back every row and return a stable error.
Auth is not required for guest checkout, but possession of an existing
Purchase token must never influence creation replay scope.

## Request and Response Requirements

- Request shape must match the approved W3-02 role-linking contract and preserve
  the current one-Offering/no-cart constraint.
- Enforce documented UUID, string, list, participant-count, money, and body
  bounds in OpenAPI, Go, and database layers.
- Reject unsupported media type, oversized body, unknown fields, trailing JSON,
  invalid relationship forms, unavailable quota, and invalid state with stable
  contract error codes.
- The response contains Purchase UUID/reference, `COMMON`, stored snapshots,
  intended participants, exact total/currency, `PENDING_PAYMENT`, stored
  reservation expiry, and the raw token only at initial/replayed checkout.
- Never include token hash, idempotency namespace/key/hash, database errors,
  internal topology, or full unrelated Party data.

## Idempotency and Secret Invariants

- The idempotency namespace includes the accepted public surface, checkout
  command, and approved stable guest scope.
- Canonical request hashing uses validated semantic values so JSON whitespace
  and object-key order do not change identity; semantic changes conflict.
- The idempotency key is stored only in its owned record and is not copied to
  logs, audit, outbox, domain tables, or API responses.
- Replay lookup/creation and all business effects share one transaction.
- Successful response status, safe headers, and encrypted body replay exactly.
- Encryption/authentication failure is a server error and never falls back to
  issuing another Purchase or token.
- Logs and telemetry never contain raw Purchase tokens, token hashes, raw
  idempotency keys, full contacts, or encryption keys.

## Planned File Changes

Expected minimum after decisions are approved:

- `apps/api/internal/purchasing/public_http.go` and focused tests;
- W3-01 through W3-05 owned files only where final command composition is
  required;
- `apps/api/internal/app/server.go`, router constructor/wiring, and
  `apps/api/internal/app/routes_public.go`;
- `apps/api/cmd/server/main.go` only if current dependency construction requires
  it;
- `contracts/openapi/storefront.yaml` for approved payload/error/retention
  clarifications;
- the canonical decision/product document(s) that close the three gates;
- existing platform idempotency code only if approved retention cannot be
  represented by its current API.

No new table or dependency is currently planned.

## Public Route and CORS Impact

- Add POST and OPTIONS for `/api/public/v1/purchases`.
- Permit only the configured Storefront origins, `POST`, and the exact headers
  needed by checkout (`Content-Type`, `Idempotency-Key`, and `X-Request-ID` if
  the existing policy accepts it).
- Keep public CORS non-credentialed and existing public GET discovery intact.
- Keep `Cache-Control: no-store` on Purchase responses.
- Apply the existing public rate limiter before expensive database/encryption
  work and retain stable `429` behavior.

## Acceptance Criteria

- One valid request creates exactly one `COMMON`/`PENDING_PAYMENT` Purchase with
  approved actors, snapshots, participants, history, 24-hour reservation,
  token hash, reference, outbox events, and encrypted replay in one transaction.
- A valid retry with the same approved namespace/key/request returns the exact
  original `201` response, including the same one-time raw token, and creates no
  additional business rows.
- Same key with different semantic input returns `409 idempotency_conflict` and
  changes no business state.
- Concurrent same-key requests and concurrent last-unit requests remain
  duplicate-free and quota-safe.
- Any injected Party/Purchase/participant/history/reservation/outbox/replay
  failure leaves no partial checkout.
- Raw token is returned only by initial/replayed checkout, only its SHA-256 hash
  is stored, and neither form leaks to logs/outbox/errors.
- Contract, CORS/preflight, rate-limit, error, and no-store behavior match the
  public runtime.
- The three W3-06 decisions plus W3-02/W3-04 decisions are recorded in canonical
  artifacts before BUILD.

## Verification

```bash
cd apps/api
go test ./internal/purchasing/... ./internal/platform/idempotency/... ./internal/app/...
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
docker compose -f infrastructure/compose.yaml config
```

Also:

- validate `contracts/openapi/storefront.yaml` and run its exposure checks;
- run migrations and end-to-end command tests against disposable PostgreSQL;
- run concurrent same-key and quota-contention cases repeatedly;
- inspect logs/test captures for token, contact, SQL, and idempotency-key leaks;
- verify CORS preflight and the exact first/replayed response bytes/headers.

Report known pre-existing validation failures separately from regressions.

## Dependencies and Sequencing

- W3-01: Party persistence.
- W3-02: approved explicit actor relationship contract.
- W3-03: canonical Purchase aggregate, persistence, history, and reads.
- W3-04: accepted snapshots and total formula.
- W3-05: transactional reservation invariant.
- W1-06: encrypted idempotency/outbox/transaction foundations.
- W3-07 consumes this public route; W3-08 consumes W3-03 reads; W3-09 expands
  cross-cutting regression coverage but does not replace focused tests here.

## Decisions, Assumptions, and Deferred Work

- Accepted: ADR-050 reference format, stable guest caller scope, and durable
  replay retention; ADR-048 role-link payload; and ADR-049 total formula.
- Accepted: 32 random token bytes, SHA-256 at rest, raw token only at checkout,
  and AES-GCM encrypted replay.
- Accepted: checkout is guest-accessible and creates one Purchase for one
  Offering without a cart.
- Deferred: tracking/cancellation/evidence/payment/activation/expiry and all
  non-Common channel workflows.

## Final Review Requirements

The final report must list every decision adopted, route/contract changes,
transaction effects, outbox events, replay retention, security evidence,
contention/rollback tests, and deferred lifecycle behavior. Do not call the
task or Week 3 backend complete while any gate or atomicity check is unresolved.
