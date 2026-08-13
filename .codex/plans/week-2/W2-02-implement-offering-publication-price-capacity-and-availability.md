# Task: W2-02 Implement Offering Publication, Price, Capacity, and Availability

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
- Tracker objective: Implement Offering publication, price, capacity, and
  availability.
- Dependency: W2-01 Event lifecycle behavior; verified Week 1 platform and
  contract foundations.

## Objective

Implement the event-scoped Offering domain and application behavior that owns
publication, exact pricing, per-Purchase capacity, total quota, advisory public
availability, and historical-safe mutation rules.

## Context

An Offering is the commercial abstraction for a package, share, or category;
it is not a physical Livestock record. Migration 0001 already creates
`offerings`; migration 0005 already adds optional `participant_quota`. The
Storefront and Operations contracts already contain Offering DTOs and routes,
but no runtime module exists.

## Source of Truth

- `docs/PRD.md` sections 6, 7.1, 10, 11.2, 12, and 13;
- `docs/PRODUCT_MAP.md` sections 2, 13, 14, and 15;
- `docs/ARCHITECTURE.md` sections 7.3, 9, 10, 16, 17, and 18;
- `docs/DECISIONS.md` ADR-009, ADR-011, ADR-032, ADR-033, ADR-038,
  ADR-041, ADR-042, ADR-044, and ADR-046;
- `docs/domain/COMMERCE_LIFECYCLES.md` Offering and quota policy;
- `docs/security/PERMISSIONS.md` Offering route mapping;
- migrations 0001, 0002, and 0005 for Offering and quota relationships;
- both OpenAPI contracts' Offering schemas;
- `.codex/CURRENT_STATE.md` for the implementation gap.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate only after this draft and W2-01 are reviewed. Revalidate the current
schema and contracts before BUILD. Do not commit or push unless asked.

## Scope

### In Scope

- Offering value types, validation, lifecycle guards, and framework-neutral
  application behavior.
- Create, update, publish, mark unavailable, republish, and archive behavior.
- Confirmed commercial-edit policy:
  - create in `DRAFT` for a non-closed/non-archived Event;
  - edit price, description, and quota only in `DRAFT` or `UNAVAILABLE`;
  - require a `PUBLISHED` Offering to become `UNAVAILABLE` before commercial
    edits;
  - keep `ARCHIVED` immutable.
- Publication when the parent Event is `PUBLISHED`, `ACTIVE`, or `SUSPENDED`;
  public visibility still requires an `ACTIVE` Event.
- Exact integer-minor-unit price and uppercase ISO-style three-letter currency
  validation; no floating point.
- Positive immutable-per-Purchase `participant_capacity` and optional total
  `participant_quota`.
- Advisory availability derived from Event and Offering quota usage.
- Application orchestration for append-only audit/outbox/replay behavior;
  W2-03/W2-05 provide and prove the concrete PostgreSQL/HTTP composition.

### Out of Scope

- Physical Livestock inventory, allocation, package composition, or species
  rules.
- An exhaustive Offering-kind enum. `offering_kind` remains a bounded,
  non-empty descriptive code with no kind-specific behavior.
- Checkout/reservation creation, expiry worker, Purchase snapshots, or quota
  consumption commands; those belong to later weeks.
- Cart, multiple Offerings per Purchase, discounts, tax, dynamic pricing,
  promotions, media galleries, or search engine optimization.
- Dedicated Offering version-history tables.

## Existing State

- SQL stores Event reference, unique event-scoped code, name, unconstrained
  kind, optional description, `price_minor`, currency, capacity, optional
  quota, lifecycle status, `published_at`, and timestamps.
- SQL validates non-negative price, uppercase three-letter currency, positive
  capacity, non-negative optional quota, and status vocabulary.
- `quota_reservations` already records participant units and supports counting
  `RESERVED` and `CONSUMED` attempts.
- No runtime calculates or exposes `available_participant_units`.
- The Operations patch contract currently cannot clear nullable description or
  quota and exposes no expected version.

## Target State

- A meaningful Offering module owns the commercial rules without importing
  Gin, SQL, or frontend types into the domain layer.
- Lifecycle matches the accepted matrix:

```text
DRAFT -> PUBLISHED -> UNAVAILABLE -> PUBLISHED
DRAFT | PUBLISHED | UNAVAILABLE -> ARCHIVED
```

- Publication requires a parent Event in `PUBLISHED`, `ACTIVE`, or
  `SUSPENDED`. `CLOSED` and `ARCHIVED` Events reject Offering creation,
  publication, and commercial-field edits, but do not block marking a
  published Offering unavailable or archiving a non-archived Offering; those
  lifecycle transitions preserve administrative cleanup without restoring
  commerce.
- Public Offering reads require both Event `ACTIVE` and Offering `PUBLISHED`.
- Availability semantics are:
  - `participant_capacity`: maximum participants in one future Purchase;
  - `participant_quota`: optional total participant units for the Offering;
  - Event `participant_quota`: optional total participant units for the Event;
  - `RESERVED` and `CONSUMED` units count as used;
  - `RELEASED` and `EXPIRED` units do not;
  - finite availability is the non-negative minimum of configured remaining
    Event and Offering totals;
  - `null` availability means neither level has a finite configured quota.
- Public availability is advisory; later checkout must lock/revalidate totals.

## Constraints

- Reuse standard Go, `database/sql`, existing pgx, transaction, audit,
  idempotency, and outbox structures. Add no money, decimal, catalogue,
  inventory, or state-machine dependency.
- API-visible minor units and quotas must be non-negative integers no greater
  than JavaScript's safe integer maximum, `9_007_199_254_740_991`, so the JSON
  number contract remains exact across Go, OpenAPI, and TypeScript.
- `participant_capacity` maps to PostgreSQL `integer`, so its maximum is
  `2_147_483_647`, not the larger JSON-safe bigint bound.
- Require already-uppercase three-letter currency codes. Reject lowercase or
  malformed input rather than silently changing commercial data.
- An Offering kind is metadata only until an accepted requirement gives it
  behavior.
- Never derive authoritative availability from TanStack Query or a cached
  frontend value.

## Implementation Requirements

### Domain behavior

- Define `Offering`, `Status`, money/capacity inputs, transition methods, and
  narrow errors.
- Reject empty/oversized code, name, or kind; negative/unsafe price; invalid
  currency; capacity outside `1..2_147_483_647`; negative/unsafe quota; and
  invalid parent Event state.
- Keep code, kind, currency, and per-Purchase capacity immutable after
  creation in Week 2. Price, name, description, and total quota may change only
  under the confirmed status policy.
- Publishing sets `published_at` on first publication and preserves it on
  republish as the first-publication timestamp. If product needs last-published
  time, that requires a separate explicit field and decision.

### Availability query

- Calculate usage from authoritative reservations, not stored counters.
- Use one query snapshot that returns Event quota, Offering quota, and summed
  participant units to avoid mixing values from different reads.
- Count only `RESERVED` and `CONSUMED`; clamp remaining values at zero for safe
  display if legacy or concurrent data temporarily exceeds a configured
  quota.
- Do not add Redis or a denormalized availability counter before measurement.

### History and lifecycle effects

- Privileged create/update/transition commands write minimized Offering
  before/after audit rows in the same transaction.
- Lifecycle commands emit `OfferingPublished`, `OfferingUnavailable`, or
  `OfferingArchived` according to the authoritative matrix.
- Offering create and lifecycle commands use the existing encrypted replay
  executor after authorization; patch/update relies on expected version and is
  not replay-stored.
- Later Purchase snapshots, not mutable Offering rows, preserve price applied
  to a Purchase; W2 records catalogue change history but does not implement W3
  snapshots.

## Planned File Changes

| File | Action | Purpose |
| --- | --- | --- |
| `apps/api/internal/offering/domain.go` | Create | Offering model, validation, pricing/capacity vocabulary, and lifecycle. |
| `apps/api/internal/offering/service.go` | Create | Framework-neutral commands and availability query orchestration. |
| `apps/api/internal/offering/*_test.go` | Create | Focused lifecycle, validation, and availability tests. |
| `.codex/CURRENT_STATE.md` | Modify when executed | Record implemented behavior without claiming later checkout or screens. |
| `.codex/TASK.md` | Create, execute, archive | Preserve approved execution and final review. |

Repository SQL belongs to W2-03; HTTP/OpenAPI correction belongs to W2-04 and
W2-05. Do not create empty adapters here.

## Dependencies and Sequencing

- Requires W2-01's Event-state vocabulary and policy.
- W2-03 supplies concrete repository queries, versioning, and availability SQL.
- W2-04 exposes public reads; W2-05 exposes Operations reads and commands.
- W2-06 and W2-07 consume the separate Operations and Storefront DTOs.
- W2-08 provides repository and cross-boundary coverage.

## API and Database Impact

- W2-03 must add `version` to Offering persistence and indexes justified by
  actual query order.
- W2-04 should make `available_participant_units` nullable to represent
  unbounded availability rather than inventing a number.
- W2-05 should make nullable patch clearing and expected-version behavior
  explicit and document first-publication timestamp semantics.

## Authorization, Audit, Idempotency, and Concurrency

- Reads require `offering.read`; changes require `offering.manage` on the
  Operations surface. Public reads remain guest-accessible.
- Parent Event state and expected Offering version are revalidated inside the
  owning transaction for commands whose accepted policy depends on the parent.
- Audit failure aborts privileged changes.
- Idempotency protects retry intent; optimistic versioning protects stale
  changes; later quota commands require row locking. These are complementary.

## Acceptance Criteria

1. Offering lifecycle, parent-Event guards, editability, republishing, and
   archival match the confirmed policy.
2. Money remains exact and bounded; currency, capacity, quota, code, name, and
   kind validation is explicit.
3. No Offering is publicly visible unless it and its Event have the required
   statuses.
4. Advisory availability correctly handles finite, mixed finite/unbounded, and
   fully unbounded quotas and counts only active/consumed reservation units.
5. Stale/invalid-state commands return stable conflicts, while invalid
   capacity/price/quota input returns validation failure, without SQL leakage.
6. Application tests prove the command boundary invokes state, audit, outbox,
   and replay effects exactly as declared; W2-05 owns real-database atomicity.
7. No Livestock, cart, kind-specific rule, cache, stored quota counter, or new
   dependency is introduced.

## Testing

- Table-test lifecycle and parent Event-state combinations.
- Test exact price/currency/capacity/quota boundaries, including the
  JavaScript-safe maximum.
- Test availability for no quota, only Event quota, only Offering quota, both
  quotas, released/expired attempts, exhausted quota, and overcommitted legacy
  data.
- Test update freeze rules, first-published timestamp preservation, stale
  versions, and side-effect orchestration with narrow fakes. W2-03 owns the SQL
  availability/concurrency proof; W2-05 owns PostgreSQL audit rollback and
  create/lifecycle replay proof.

## Verification

```bash
cd apps/api && go run ./cmd/format -check
cd apps/api && go vet ./...
cd apps/api && go test ./internal/offering/... ./internal/event/...
cd apps/api && go test ./...
cd apps/api && go build ./...
make validate
```

This task may archive after its scoped domain/application criteria pass. It
must not claim SQL availability, contested update, or transaction atomicity;
W2-03, W2-05, and W2-08 prove those against disposable PostgreSQL 18.

## Deliverables

- Reviewed Offering domain and application implementation.
- Explicit availability semantics and focused tests.
- Current-state update and verified task archive only after execution.

## Risks and Clarifications to Review

- Confirmed default: `offering_kind` remains bounded text without a dropdown
  enum or kind-specific behavior.
- Confirmed default: a published Offering must become unavailable before price
  or quota edits.
- Confirmed default: discovery stays visible outside the Event registration
  window; the UI labels the window state.
- Confirmed default: fully unbounded availability is `null`, not zero or an
  invented large number.
- Historical Purchase price snapshots remain W3-04 and must not be claimed as
  Week 2 runtime behavior.

## Final Report Requirements

Distinguish implemented Offering behavior, verified exact-money/availability
behavior, assumptions, deferred Purchase/Livestock work, plan variance,
remaining catalogue and concurrency risks, and the next task.
