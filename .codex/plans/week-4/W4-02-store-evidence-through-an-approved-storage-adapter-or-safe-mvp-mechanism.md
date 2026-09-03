# Task: W4-02 Store Evidence through an Approved Storage Adapter or Safe MVP Mechanism

## Status

Ready for plan review on 2026-09-03, but blocked for BUILD by the object-storage
decision gate. This is planning-only. The live tracker remains `Backlog`; no
tracker, active-task, implementation, archive, commit, or push change is
authorized.

## Tracker

- Week: Week 4
- Epic: Payment & Activation
- Application: API
- Category: Backend
- Priority: P0
- Estimate: 4 hours
- Tracker objective: Store evidence through an approved storage adapter or safe
  MVP mechanism.
- Dependencies: W4-01 and W1-02.
- Acceptance summary: Payment evidence is stored safely and its immutable
  reference/digest is traceable from the Payment.

## Objective

Implement and wire one private evidence-storage adapter behind the narrow W4-01
port. The adapter must accept a bounded stream, store it immutably under an
opaque internal key, and support authorized retrieval/deletion needed by the
command consistency protocol without exposing a public URL or provider model.

Prefer the smallest approved mechanism. A standard-library private filesystem
adapter is the recommended MVP default only if the deployment decision accepts
a persistent private volume and its single/multi-replica limits. Otherwise use
the selected object provider through an equally narrow adapter.

## Context

Canonical rules require evidence bytes in private object storage and only an
opaque reference plus filename/media-type/size/SHA-256 in PostgreSQL. The
architecture still lists the object-storage provider, evidence retention, and
production configuration as unresolved. No storage SDK, bucket, configuration,
port, adapter, or evidence retrieval endpoint exists.

External storage cannot participate atomically in the PostgreSQL transaction.
The plan must therefore define ordering, rollback cleanup, retry identity,
crash-orphan handling, and retrieval authorization explicitly.

## Source of Truth

- all W4-01 sources and resolved gates;
- `docs/ARCHITECTURE.md`, external integrations, deployment, transactions, and
  architecture decisions required;
- `docs/DECISIONS.md`, especially ADR-020, ADR-024, ADR-026, ADR-027, ADR-030,
  ADR-038, ADR-041, and ADR-042;
- `docs/deployment/STAGING.md` private evidence configuration/security rules;
- `docs/database/ERD.md` and migration 0005 evidence fields;
- `contracts/openapi/storefront.yaml` and `contracts/openapi/operations.yaml`;
- current config/bootstrap/logging/PWA/security source;
- `.codex/CURRENT_STATE.md`, roadmap, and live tracker W4-02 metadata.

## Required Workflow

`DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW`

Before BUILD, record an accepted ADR naming the MVP mechanism, environment
support, consistency/cleanup model, private read authorization, configuration,
retention owner, and migration path. Provider choice cannot be hidden in code.

## Scope

### In Scope

- Finalize a small `EvidenceStore` contract for put, authorized open/read, and
  best-effort removal of one immutable object.
- Implement exactly one approved adapter and deterministic configuration
  validation; fail closed where the selected mechanism is unsafe/unavailable.
- Generate opaque unguessable object keys independent of original filenames,
  Purchase references, Party/contact data, and idempotency keys.
- Stream at most 10 MiB plus one sentinel byte while computing SHA-256 and
  trusted media type; avoid whole-file buffering.
- For a filesystem MVP, create private directories/files, prevent traversal and
  symlink escape, use exclusive/atomic creation, and keep storage outside the
  repository and web roots.
- Put the object only for the winning idempotent command. Commit its opaque
  reference and metadata with the Payment; remove it on command rollback when
  possible.
- Define bounded orphan detection/cleanup for a crash after object creation and
  before database commit without deleting referenced evidence.
- Provide internal retrieval to W4-06 through the same port. The Go API, not a
  browser-visible bucket path, owns authorization and safe response headers.
- Add safe health/readiness behavior only if the accepted mechanism requires it.
- Add focused adapter, configuration, composition, failure, and security tests.

### Out of Scope

- Multiple provider implementations, runtime provider switching, mirroring,
  replication, CDN delivery, public ACLs, image transformation, virus scanning,
  OCR, media transcoding, or a general asset service.
- Persisting evidence bytes, paths, credentials, presigned URLs, or provider
  payloads in PostgreSQL/API/audit/outbox.
- Inventing an evidence retention/deletion duration; retention remains a
  canonical product/compliance decision.
- Payment business rules, verification, eligibility, activation, or frontend
  workflows beyond the storage boundary required by W4-01/W4-06.
- New infrastructure or dependencies not required by the accepted adapter.

## Existing State

- No Go storage abstraction or adapter exists.
- `payment_records.evidence_reference` is a non-public text field with complete
  metadata/digest constraints but no reference uniqueness constraint.
- Staging documentation reserves private storage credentials/bucket names for a
  future adapter and prohibits public evidence URLs.
- The Go module has no object-storage dependency. Standard library filesystem,
  streaming, hashing, MIME sniffing, path, and permission primitives are
  available.
- The Operations Payment contract exposes metadata only. It has no evidence
  byte/download route, so W4-06 cannot review evidence until a safe private read
  contract is approved.

## Target State

- A configured adapter stores one immutable object before the Payment reference
  is committed and returns one opaque key plus trusted metadata.
- Same-intent replay does not write or read a second object.
- Database failure attempts cleanup; a documented orphan scanner can identify
  unreferenced old objects after crash windows without touching referenced data.
- Only an authenticated and authorized Operations API path can retrieve bytes.
  Storefront responses never reveal storage identity.

## Adapter Contract and Consistency

- Input is a context, opaque generated object identity, trusted media type, and
  bounded reader. Output is the stored opaque reference and exact byte count.
- Put is create-only. Existing key with different content fails; normal writes
  never overwrite or mutate evidence.
- W4-01 calculates/validates digest and metadata while streaming; the adapter
  must not trust filename or client path as storage identity.
- The winning idempotency transaction may hold its bounded database claim while
  storing the object. After successful put, it writes Payment/history/outbox/
  replay and commits. Any failure before commit triggers best-effort delete.
- A crash can leave an unreferenced object but must never leave a committed
  Payment pointing to an object that was never durably finalized. The accepted
  ADR must specify durability acknowledgement and orphan grace period.
- Retrieval uses the exact database-owned reference after authorization; it
  rejects malformed keys and never resolves arbitrary user paths.

## Filesystem MVP Requirements, If Approved

- Configure one absolute storage root outside the checkout and any served web
  directory. Reject root, relative, missing, world-writable, symlinked, or
  otherwise unsafe configuration according to the accepted environment policy.
- Create directories with private permissions and files with mode `0600` under
  the deployment account's restrictive umask.
- Write a temporary file in the target directory, sync/close it, then atomically
  rename without replacement to the final opaque key.
- Resolve and validate every final path remains under the configured root.
- Never put original filename, Purchase reference, Party name, payment reference,
  or media type in the path.
- Document the ceiling: a shared durable volume is required for more than one
  API instance; otherwise production-like multi-replica use fails closed.

## Planned File Changes

- the W4-01 payment evidence port and tests;
- `apps/api/internal/payment/storage_filesystem.go` plus focused tests if the
  filesystem MVP is accepted, or one provider-specific adapter file otherwise;
- `apps/api/internal/config/config.go`, tests, and safe `.env.example` names;
- `apps/api/cmd/server/main.go`, `internal/app/server.go`, and public route
  composition for fail-closed construction;
- `contracts/openapi/operations.yaml` for the approved private evidence read
  route consumed by W4-06;
- `docs/DECISIONS.md`, `docs/deployment/STAGING.md`, and current state for the
  accepted storage/retention boundary;
- focused tests near each changed boundary.

No migration is planned unless the accepted consistency/cleanup design proves
that existing opaque-reference metadata cannot represent it.

## Acceptance Criteria

1. One approved adapter stores evidence immutably and returns an opaque internal
   reference traceable from exactly one Payment.
2. Content is bounded at 10 MiB, streamed, hashed, and stored without trusting
   filenames, client paths, or declared MIME alone.
3. Traversal, symlink escape, overwrite, public ACL/URL, unsafe root, missing
   credentials, and unsupported environment configurations fail closed.
4. Same-intent replay performs no second put; failed commands do not commit a
   Payment/reference and perform documented cleanup.
5. A crash-window orphan is detectable after a bounded grace period; referenced
   evidence is never deleted by cleanup.
6. A committed Payment never points to an unfinalized object according to the
   selected adapter's durability contract.
7. Retrieval is private, no-store, content-type safe, and authorization-gated;
   neither public API nor frontend receives a provider key or credential.
8. Evidence bytes/references/digests/credentials remain absent from logs, errors,
   audit, outbox, fixtures, screenshots, PWA caches, and Git.
9. The selected mechanism, environment ceiling, configuration, cleanup, and
   retention ownership are recorded canonically before implementation.

## Testing

- Port contract tests for create-only put/open/delete, cancellation, short read,
  limit+1, digest, and same-key collision.
- Filesystem tests, if selected, use a temporary directory and cover private
  modes, traversal, symlink, atomic finalization, cleanup, and restart reads.
- Composition tests prove missing/invalid storage configuration prevents only
  the evidence route/capability according to the approved fail-closed policy.
- Integration tests inject storage failures before and after put and database
  failures after put; assert exact database/object outcomes and no duplicate on
  replay/concurrency.
- Authorized retrieval tests cover missing, forbidden, wrong Payment, missing
  object, safe headers, and content bytes without logging them.

## Verification

```bash
cd apps/api
go test ./internal/payment/... ./internal/config/... ./internal/app/... -count=1
go vet ./...
go test ./...
go build ./...
cd ../..
make validate
docker compose -f infrastructure/compose.yaml config
```

Also run disposable-PostgreSQL command tests with a temporary/private adapter,
restart the process and retrieve a committed object, inspect permissions and
paths, and scan Git/log/test output for evidence or credential leakage.

## Decision Gates and Risks

- **Provider/mechanism:** accept filesystem plus persistent private volume, or
  name a concrete object-storage provider and SDK/configuration.
- **Retrieval permission:** decide whether evidence bytes require
  `payment.verify` (recommended least privilege) or `payment.read`; metadata may
  remain visible under `payment.read`.
- **Consistency:** accept the bounded in-transaction put plus best-effort rollback
  cleanup/orphan scan, or approve a larger staging/finalization schema.
- **Retention:** define retained duration, legal/operational hold, and deletion
  authority. W4-02 must not invent it.
- **Replica ceiling:** local filesystem without a shared volume is single-instance
  only. It must be documented and rejected in incompatible environments.
- Malware scanning is not an accepted Phase 1 requirement. If compliance later
  requires it, add a quarantined state and asynchronous scanner deliberately.

## Deliverables

- Accepted evidence-storage ADR and deployment/configuration update.
- One private adapter, W4-01 composition, safe retrieval port, focused tests,
  and exact consistency/cleanup evidence.
- No provider proliferation, public URL, or speculative storage platform.

## Final Report Requirements

Report the selected mechanism, configuration names (never values), durability,
object permissions/ACL, key format properties, transaction/storage ordering,
rollback/orphan behavior, retrieval authorization, exact tests, validation,
known ceiling, and deferred retention/provider work.
