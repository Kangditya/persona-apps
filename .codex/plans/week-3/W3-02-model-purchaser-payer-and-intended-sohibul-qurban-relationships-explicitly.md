# Task: W3-02 Model Purchaser, Payer, and Intended Sohibul Qurban Relationships Explicitly

## Status

Completed and verified on 2026-08-21. The executed task is archived at
`.codex/archive/2026-08-21-TASK-w3-02-model-purchaser-payer-and-intended-sohibul-qurban-relationships-explicitly.md`;
no commit or push was requested.

## Tracker

- Week: Week 3
- Epic: Common Purchase
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 5 hours
- Tracker objective: Model purchaser, payer, and intended Sohibul Qurban
  relationships explicitly.
- Dependency: W3-01.
- Acceptance summary: purchaser, payer, and intended Sohibul Qurban may be
  different Parties, while an explicitly shared person reuses one Party.

## Objective

Model the Common Purchase actors without collapsing financial, ordering, and
Qurban roles. Persist purchaser, payer, and intended-participant relationships
through the already-existing foreign keys and snapshots, and require an
explicit request representation whenever more than one role refers to the
same Party.

## Source of Truth

- `docs/PRD.md`, actor definitions, Common Purchase, Party Management, and
  Intended Participant requirements;
- `docs/PRODUCT_MAP.md` and `docs/ARCHITECTURE.md`;
- `docs/DECISIONS.md`, especially ADR-013, ADR-014, ADR-042, ADR-043, and
  ADR-044;
- `docs/domain/COMMERCE_LIFECYCLES.md`;
- `docs/security/PERMISSIONS.md` and `docs/security/AUTHENTICATION.md`;
- `docs/database/ERD.md` and migrations `0002` and `0005`;
- `contracts/openapi/storefront.yaml` and
  `contracts/openapi/operations.yaml`;
- W3-01's approved Party output and the current source.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Activate only after W3-01 is implemented and verified and the decision below
is approved. Do not commit or push unless requested.

## Contract Decision — Approved for BUILD

ADR-048 and the Storefront contract define the exact request representation:

- Purchaser is a `PartyDeclaration` with request-local `party_ref`, display
  name, and optional contact values.
- Payer is either a distinct `PartyDeclaration` or a `PartyReference` to the
  purchaser declaration.
- Each intended participant is either a `PartyReference` to a declared
  purchaser/payer Party or a `ParticipantNameInput` that remains unresolved.
- A declaration label is unique within the request. References must resolve to
  exactly one declaration; a duplicate declaration, unknown reference, or
  mixed reference/declaration object is rejected.
- `party_ref` is opaque and request-local, never a durable Party UUID. Equal
  names or contacts do not establish identity equality. The full payload,
  including labels, is the idempotency-hash input.

## Existing State

- `purchases.purchaser_party_id` is required and
  `purchases.payer_party_id` is nullable at the database level.
- `purchase_participants` stores ordered intended positions, a required
  display-name snapshot, and an optional `party_id`.
- Common checkout currently requires purchaser and payer contacts and one or
  more participant names, but provides no explicit same-Party link.
- Intended participants may remain unresolved before verification. Accepted
  lifecycle rules require every participant to resolve to a Party before
  automatic activation, not necessarily at checkout.
- Automatic identity deduplication is explicitly deferred in the tracker.
- No extra relationship table is required to represent the accepted roles.

## Scope

### In Scope

- Define domain role types for purchaser, payer, and ordered intended
  participants using Party UUID references and required snapshots.
- Enforce that a Common Purchase has a purchaser and payer at creation, even
  though the shared base schema remains nullable for other future workflows.
- Allow purchaser and payer to be distinct or to share one Party only through
  the approved explicit request representation.
- Allow an intended participant to reference an explicitly selected Party or
  remain name-only with `party_id = NULL` until verification.
- Preserve participant ordering with `sequence_no` and one display-name
  snapshot per intended position.
- After the contract decision, update the Storefront OpenAPI request and
  examples before or with the Go transport types.
- Map the relationships into the existing Purchase and participant columns;
  add focused validation and persistence tests.

### Out of Scope

- Guessing identity from equal names, email addresses, phone numbers, or
  normalized contact values.
- Automatic contact matching, merging, or duplicate resolution.
- Creating a Party from a name-only participant as if a display name proved a
  durable identity.
- Creating `sohibul_qurban` outcome records. Those are created only after the
  accepted payment/activation conditions, not during checkout.
- Saving-account holders, sponsors, giveaway applicants/recipients, or future
  channel-specific actor rules.
- Public Party search, customer accounts, relationship graphs, or a generic
  role-assignment framework.

## Relationship Mapping

| Purchase meaning | Persisted representation |
| --- | --- |
| Purchaser | `purchases.purchaser_party_id` references one Party |
| Payer | `purchases.payer_party_id` references one Party for COMMON checkout |
| Same purchaser and payer | both columns contain the same explicitly selected Party UUID |
| Distinct purchaser and payer | the columns contain two distinct Party UUIDs |
| Resolved intended participant | ordered `purchase_participants` row with `party_id` and display snapshot |
| Unresolved intended participant | ordered row with `party_id = NULL` and display snapshot |
| Sohibul Qurban outcome | not created in Week 3; later lifecycle work writes `sohibul_qurban` |

The display-name snapshot preserves historical intent; it is not a second
source of identity truth and is never used to merge Parties.

## Target State and Invariants

- Role equality is represented by Party-ID equality that came from an explicit
  approved input choice, never from server-side contact inference.
- Role difference remains possible even when the same person could technically
  fill both roles; the server follows the explicit request contract.
- Participant sequence is contiguous, stable, and unique within the Purchase.
- Participant count equals the number of intended-participant rows created.
- Every intended row has a non-empty historical display snapshot.
- A name-only participant remains intentionally unresolved and cannot be used
  to satisfy later activation's Party requirement.
- No role data is flattened into the Purchase purchaser identity or copied into
  a separate payer/participant identity table.

## Implementation Plan

1. Apply the approved Storefront payload and response examples.
2. Update the OpenAPI contract first, including mutually exclusive/required
   branches, bounds, and examples for shared and distinct roles.
3. Add the minimum domain types and validation needed to produce the mapping
   above. Keep Party creation/reuse in W3-01 and Purchase persistence in W3-03.
4. Add repository mapping for purchaser/payer Party IDs and ordered participant
   rows without adding a new table or generic relationship abstraction.
5. Add tests for distinct roles, explicitly shared roles, mixed resolved and
   unresolved participants, order preservation, contradictions, and rollback.
6. Revalidate the OpenAPI contract and run focused/full Go checks.

## Planned File Changes

Expected minimum after the decision is approved:

- `contracts/openapi/storefront.yaml`;
- focused relationship/domain code in `apps/api/internal/purchasing/`;
- focused `_test.go` files;
- W3-03 repository files only if that approved task already owns the mapping.

Do not create an empty purchasing module merely to satisfy this plan. If task
execution is kept strictly sequential, the meaningful relationship types may
be added with W3-03 in the same active vertical slice and attributed clearly.

## API and Database Impact

- API: the Create Common Purchase request must change only after approval of
  the explicit role-linking representation.
- Database: reuse existing Party foreign keys and participant rows; no schema
  change is currently justified.
- Audit/outbox/idempotency: composed by W3-06, not emitted by this internal
  modeling task.

## Acceptance Criteria

- The approved contract can express different purchaser, payer, and intended
  participant Parties.
- The approved contract can explicitly express a shared Party for supported
  role combinations without duplicating identity data.
- Ambiguous or contradictory relationship inputs fail validation and create
  no rows.
- Name/contact equality never causes implicit Party reuse or merge.
- Participant order, display snapshots, nullable resolution state, and count
  are mapped exactly to the existing schema.
- No `sohibul_qurban` outcome is created during pending checkout.
- Contract and domain tests cover shared, distinct, unresolved, invalid, and
  rollback cases.

## Verification

```bash
cd apps/api
go test ./internal/identity/... ./internal/purchasing/...
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
docker compose -f infrastructure/compose.yaml config
```

Parse/lint the changed Storefront OpenAPI document with the same pinned parser
used by the executed W1-05 contract task. That parser command is not currently
persisted as a Make target, so establish or record the exact reproducible
command before claiming contract validation. Verify against disposable
PostgreSQL where repository mapping is implemented.

## Decisions, Assumptions, and Deferred Work

- Accepted: ADR-048 supplies the exact Storefront payload for explicit
  same-Party roles.
- Accepted: name-only intended participants may remain unresolved until the
  later verification/activation workflow.
- Accepted: automatic identity deduplication is deferred.
- Deferred: creation of Sohibul Qurban outcomes, payment verification, and
  later channel-specific roles.

## Final Review Requirements

The final report must reproduce the approved payload semantics, show the
role-to-column mapping, list tests and schema impact, and distinguish resolved
Parties from name-only intended participants. Do not call the task complete if
the contract decision remains open or contract/runtime behavior diverges.
