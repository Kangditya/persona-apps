# Task: W3-04 Snapshot Event, Offering, Price, and Participant-Capacity Data

## Executed

## Status

Completed and verified on 2026-08-21. No commit or push was requested.

## Tracker

- Week: Week 3
- Epic: Common Purchase
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Snapshot Event, Offering, price, and participant-capacity
  data.
- Dependencies: W2-01, W2-02, and W3-03.
- Acceptance summary: Purchase history preserves the accepted Event/Offering
  identity and commercial values captured at checkout.

## Objective

Capture the accepted immutable checkout facts on the Purchase so later
Event/Offering or Party changes cannot rewrite purchase history. Use the
snapshot columns already present in PostgreSQL and make every Purchase read use
those stored values rather than live catalogue joins.

## Source of Truth

- `docs/PRD.md`, Phase 1 Purchase snapshot and exact-money rules;
- `docs/PRODUCT_MAP.md` and `docs/ARCHITECTURE.md`, Event isolation and
  historical correctness;
- `docs/DECISIONS.md`, especially ADR-011, ADR-013, ADR-033, ADR-042, and
  ADR-044;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- `docs/database/ERD.md` and migration `0002_commerce_and_funding`;
- both OpenAPI contracts;
- W2 Event/Offering source and W3-03's approved Purchase aggregate.

ADR-042 is the most specific accepted snapshot rule: capture Offering
identity, name/kind, unit price, currency, participant capacity, and intended
participant names. The tracker wording “Event snapshot” is therefore bounded
to the immutable Event identity on the Purchase unless product/architecture
accepts additional Event snapshot fields before BUILD.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate only after W3-03 is implemented and the decision below is approved.
Do not commit or push unless requested.

## Decision Gate — Purchase Total Semantics

The schema names the captured price `offering_unit_price_minor` and stores both
`participant_count` and `total_amount_minor`, but the accepted product artifacts
do not state whether a direct Purchase total is:

- one Offering price for the Purchase; or
- Offering unit price multiplied by participant count.

This affects charged value, financial history, idempotency hashes, and later
ledger entries. The approved formula is `offering_unit_price_minor ×
participant_count`: existing two-participant Purchase fixtures consistently
persist twice the unit price, matching the stored unit-price name and
participant-unit quota model. The implementation must use checked integer
arithmetic and reject overflow.

## Existing State

- `purchases` already stores `event_id`, `offering_id`, Offering name, kind,
  unit price, participant capacity, participant count, total amount, currency,
  and Purchase status/version.
- `purchase_participants` stores the ordered intended display-name snapshot.
- No extra Event name/year/window/quota snapshot columns exist.
- Event and Offering records are lifecycle-managed and historically auditable,
  but Purchase reads must still use their own accepted commercial snapshots.
- Public Purchase responses already describe the captured Offering name, kind,
  unit price, capacity, total, currency, participant list, and reservation
  expiry. Operations DTOs expose a smaller accepted view.

## Scope

### In Scope

- Load the selected Event/Offering in the W3-06 transaction and verify their
  foreign-key relationship before building the snapshot.
- Copy the accepted Offering name, kind, unit price minor, participant capacity,
  and currency into the existing Purchase columns.
- Preserve Event and Offering identity with their UUID foreign keys.
- Copy each intended participant display name into the ordered participant
  snapshot rows.
- Calculate and store `total_amount_minor` only using the approved formula and
  checked exact-integer arithmetic.
- Ensure public and Operations Purchase DTO mapping reads stored snapshot
  columns rather than mutable Event/Offering display/configuration data.
- Add tests proving commercial snapshot stability after source Event/Offering
  changes and participant-name snapshot stability after a linked Party changes.
  Do not claim purchaser/payer Party summaries are snapshotted.

### Out of Scope

- Copying every Event or Offering field “just in case,” serializing a full
  catalogue object, or adding a generic snapshot framework/JSON blob.
- New Event name/year/window/quota snapshot columns without an accepted
  requirement or ADR.
- Floating-point money, currency conversion, tax, discounts, fees, donations,
  split tender, or payment ledger work.
- Mutating snapshot fields after Purchase creation.
- Quota availability/reservation logic, owned by W3-05.
- Public command/idempotency/token/reference composition, owned by W3-06.

## Snapshot Mapping

| Checkout fact | Historical Purchase representation |
| --- | --- |
| Event identity | `purchases.event_id` |
| Offering identity | `purchases.offering_id` |
| Offering display name | Offering-name snapshot column |
| Offering kind | Offering-kind snapshot column |
| Price | `offering_unit_price_minor` |
| Currency | Purchase currency snapshot |
| Participant capacity | Offering participant-capacity snapshot |
| Intended participant count | Purchase participant count |
| Intended participant names | ordered participant display-name snapshots |
| Total amount | exact stored total using the approved formula |

Use exact current column names from the migration during BUILD; do not create
aliases or duplicate columns to make this table prettier.

## Target State and Invariants

- All captured values come from one transactionally consistent checkout view.
- The selected Offering belongs to the captured Event.
- Price and total are non-negative exact minor-unit integers within both
  database and JSON-safe contract bounds.
- Currency is the accepted three-letter code captured with the price.
- Participant count equals the number of name snapshots and does not exceed
  the captured participant capacity.
- Snapshot values never change when source Event/Offering/Party rows change.
- API reads never reconstruct historical commercial values or intended
  participant names from live catalogue/Party joins. Current purchaser/payer
  summaries remain separate referenced identity data.

## Implementation Plan

1. Record the approved total formula in canonical product/decision source
   before changing runtime behavior.
2. Reconfirm exact migration/OpenAPI field names and bounds.
3. Add a small snapshot constructor on the Purchase side that accepts the
   selected Event identity, Offering state, intended participants, and approved
   total rule. Reuse existing money/bound helpers where present; otherwise use
   minimal checked integer arithmetic.
4. Call snapshot construction only after W3-05's consistent Event/Offering lock
   and eligibility checks so values cannot race with a concurrent mutation.
5. Persist the existing snapshot columns through W3-03's repository.
6. Map Storefront/Operations responses from the stored Purchase snapshots.
7. Add tests for exact capture, overflow/bounds, Event/Offering mismatch,
   participant mismatch, and source-record mutation after checkout.

## Planned File Changes

Expected minimum:

- focused Purchase snapshot code/tests in `apps/api/internal/purchasing/`;
- W3-03 PostgreSQL mapping and DTO files as needed;
- canonical product/decision document that records the approved total formula;
- OpenAPI only if the accepted formula or existing DTO meaning requires a
  contract clarification.

No migration is planned. If review accepts more Event snapshot fields, return
to PLAN and define the smallest additive migration and contract impact first.

## API and Database Impact

- API: no new route; existing Purchase responses gain runtime-backed snapshot
  values when W3-06 publishes creation.
- Database: reuse existing snapshot columns; no backfill because no Purchase
  runtime currently exists.
- Concurrency: snapshot construction must share W3-05's lock/transaction, not
  run as a separate preflight read.
- Audit/outbox/idempotency: W3-06 writes applicable command effects using the
  finalized snapshot result.

## Acceptance Criteria

- A valid Purchase stores exact Event/Offering identity, accepted Offering
  name/kind/price/currency/capacity, participant count/names, and approved
  total amount.
- Event/Offering mismatch or invalid/overflowing money/count data creates no
  Purchase.
- Later Event/Offering changes do not alter stored commercial snapshot fields,
  and later Party changes do not alter intended-participant name snapshots.
- DTO mapping uses stored snapshots, not live catalogue fields.
- No unapproved snapshot column, JSON blob, migration, or framework is added.
- The canonical source records the exact total formula and tests cover it.

## Verification

```bash
cd apps/api
go test ./internal/purchasing/...
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
docker compose -f infrastructure/compose.yaml config
```

Run disposable PostgreSQL integration coverage that creates a Purchase,
mutates source Event/Offering data and a linked participant Party, and proves
the commercial/participant snapshot fields remain unchanged. Validate both
OpenAPI contracts if either is edited.

## Decisions, Assumptions, and Deferred Work

- Accepted: `total_amount_minor` is the exact checked product of captured
  Offering unit price and intended participant count.
- Accepted: snapshot the ADR-042 fields and immutable Event/Offering identity.
- Assumption constrained by ADR-042/current schema: “Event snapshot” does not
  mean copying full Event configuration. Expand only through an accepted ADR
  and additive schema plan.
- Deferred: later financial ledger/evidence/payment behavior and any wider
  historical Event reporting requirements.

## Final Review Requirements

Report the accepted formula, exact source-to-snapshot mapping, schema changes
(expected none), historical-stability evidence, verification output, and any
remaining ambiguity. Do not mark complete while the formula is undecided.

## Execution Review

- Implemented: ADR-049 fixes the exact total as captured unit price multiplied
  by intended participant count. `NewPurchase` rejects mismatched, unsafe, and
  overflowing values, while `NewSnapshotPurchase` copies the locked Event and
  Offering identity/commercial values into the existing Purchase fields.
- Implemented: Storefront and Operations contract descriptions now make the
  captured total rule explicit. No migration, snapshot blob, new Event fields,
  or generic snapshot framework was added.
- Verified: focused purchasing tests, a PostgreSQL source-mutation regression,
  full API vet/test/build with disposable PostgreSQL, `make validate`, Compose
  configuration, `git diff --check`, and both OpenAPI lint runs passed. The
  OpenAPI linter retains its pre-existing one Storefront and five Operations
  warnings.
- Assumed: ADR-042/current schema bounds the Event snapshot to immutable Event
  identity. Current purchaser/payer Party summaries intentionally remain live
  referenced identity data.
- Deferred: W3-06 owns the public command, idempotency, reference/token, and
  outbox composition; later financial and richer historical Event reporting
  need separately accepted requirements.
