# Engineering Conventions

## Purpose and status model

This document is the canonical rulebook for how engineers and coding agents
change this repository. It applies the product, architecture, and decision
documents; it does not redefine product behavior or make a new architectural
decision.

Classify every repository-wide convention as one of:

| Level       | Meaning                                                                                                       |
| ----------- | ------------------------------------------------------------------------------------------------------------- |
| REQUIRED    | Follow unless a meaningful exception is documented.                                                           |
| RECOMMENDED | Default approach; a concrete local reason may justify a different approach.                                   |
| EMERGING    | Direction is established, but repository adoption is incomplete. Apply it forward without a cleanup refactor. |
| TBD         | Evidence is insufficient or an explicit decision is still needed. Do not invent a policy.                     |

When a distinction matters, work items and reviews must label statements as a
verified fact, repository inference, assumption, or TBD. Existing/manual state
is baseline; do not revert it merely because it differs from a newer pattern.

## 1. Authority and documentation ownership

| Owner                          | Canonical responsibility                                                  |
| ------------------------------ | ------------------------------------------------------------------------- |
| README.md                      | Human entrypoint, setup, supported commands, and navigation.              |
| docs/PRD.md                    | Product goals, actors, requirements, and rules.                           |
| docs/PRODUCT_MAP.md            | Product capabilities, application ownership, roadmap, and open questions. |
| docs/ARCHITECTURE.md           | Runtime, boundaries, data, API, and deployment design.                    |
| docs/DECISIONS.md              | Accepted and superseded architecture and technology decisions.            |
| docs/CONVENTIONS.md            | Engineering implementation rules, defaults, enforcement, and exceptions.  |
| docs/domain and docs/security  | Detailed accepted lifecycle, permission, and authentication contracts.    |
| docs/database                  | Schema design and database lifecycle guidance.                            |
| .codex/CURRENT_STATE.md        | Implemented and incomplete repository state.                              |
| AGENTS.md and .codex/AGENTS.md | Coding-agent operating instructions.                                      |
| .codex/TASK.md                 | One active, bounded work item.                                            |

Authority flows from product intent to implementation:

    PRD and PRODUCT_MAP
        -> ARCHITECTURE
        -> accepted DECISIONS
        -> CONVENTIONS
        -> implementation and verification
        -> task-specific execution

REQUIRED:

- Start substantial work with the applicable product, architecture, decision,
  state, agent, and active-task documents.
- Treat an active task as execution scope, not permission to silently override
  an accepted product, architecture, or ADR decision.
- Escalate a material contradiction: record it as a plan risk or decision gap,
  update the authoritative product or architecture document when approved,
  then implement.
- Keep rules in their canonical owner. Link instead of copying large rule
  bodies into README files, agent instructions, tasks, or code comments.
- Add an ADR for a cross-module contract, infrastructure, security/data
  integrity, deployment-boundary, or long-term operational decision. Do not
  create ADRs for ordinary local implementation choices.

## 2. Work, task, and exception workflow

REQUIRED workflow:

    DISCOVER -> PLAN -> BUILD -> VERIFY -> REVIEW

Before BUILD, the plan must state objective, current and target behavior,
ownership, file changes, dependencies, API and database impact, authorization,
audit, idempotency, concurrency impact, verification, risks, and unresolved
decisions. BUILD begins only after the plan is reviewed and explicitly
approved.

Use the smallest coherent diff. Reuse an existing local pattern before adding
an abstraction or dependency. Do not create empty domain modules, speculative
infrastructure, generic utility packages, or broad compliance refactors.
Do not commit or push unless explicitly requested.

Agents inspect before assuming, reuse before abstracting, plan before
modifying, verify actual behavior, report failures exactly, and do not silently
expand scope or invent unavailable product/backend behavior.

Substantial new tasks should use:

    # Task: <name>
    ## Status
    ## Objective
    ## Context
    ## Source of Truth
    ## Required Workflow
    ## Scope
    ### In Scope
    ### Out of Scope
    ## Existing State
    ## Target State
    ## Constraints
    ## Implementation Requirements
    ## Planned File Changes
    ## Acceptance Criteria
    ## Testing
    ## Verification
    ## Deliverables
    ## Final Report Requirements

Use the existing status vocabulary: Draft, Ready for Planning, Plan Approved,
In Progress, Blocked, Executed, and Verified. Keep an already-active task as
baseline even if it uses an older heading form.

If work requires an unplanned file or behavior, record a PLAN VARIANCE before
touching it: planned work, unexpected requirement, necessity, impact, and why
it remains within the approved objective. Unrelated discoveries become a
follow-up task.

A meaningful convention exception records the convention, reason, scope, risk,
temporary or permanent status, and follow-up when temporary. Architectural
exceptions belong in an ADR; local exceptions belong in the relevant task or
code documentation. Do not require exception paperwork for trivial choices.

After verified implementation, mark the active task Executed, archive it using
the documented date-and-title filename, verify the copy, and remove only the
active task file. Do not archive incomplete work.

## 3. Repository ownership and package boundaries

| Location                          | Owner and boundary                                                                                                                  |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| apps/api                          | REQUIRED authoritative Go API, business invariants, authorization, transactions, persistence, audit, idempotency, and integrations. |
| apps/storefront-web               | REQUIRED public user experience and public-contract client only.                                                                    |
| apps/operations-web               | REQUIRED internal operator user experience and operations-contract client only.                                                     |
| contracts/openapi/storefront.yaml | REQUIRED public API contract.                                                                                                       |
| contracts/openapi/operations.yaml | REQUIRED Operations API contract.                                                                                                   |
| packages/api-client               | REQUIRED shared transport primitives and normalized transport errors only; endpoint modules remain application-owned.               |
| packages/ui                       | REQUIRED domain-neutral, accessible presentation primitives and tokens only.                                                        |
| packages/typescript-config        | REQUIRED shared compiler defaults for TypeScript applications.                                                                      |
| packages/eslint-config            | EMERGING reserved shared lint configuration; Oxlint is currently used directly.                                                     |
| packages/contracts                | TBD; this package does not exist and must not be created without an approved reusable contract need.                                |
| infrastructure                    | REQUIRED local infrastructure configuration, not application business behavior.                                                     |

Promote code into packages only when it is application-neutral, genuinely used
or clearly needed by more than one application, and stable enough to share.
Do not move feature pages, routes, authentication flows, environment reads,
public/internal DTOs, or qurban business vocabulary into a package for
convenience.

The shared UI package exports its public root surface and tokens. Applications
must use those exports rather than deep source imports or duplicate primitive
sets. Generated shadcn-style component source remains editable in that package;
the package does not own routing, API access, sessions, or product rules.

## 4. Backend, API, and data conventions

### 4.1 Go structure and dependency direction

REQUIRED:

- Use the existing Go module under apps/api with Gin as the canonical HTTP
  framework and router at the HTTP adapter/bootstrap boundary. Public and
  Operations routes register explicitly through separate Gin route groups;
  net/http remains valid for the server/runtime foundation.
- Keep HTTP and integration adapters above application behavior, and domain
  behavior above PostgreSQL/provider details. Domain code must not import HTTP
  frameworks, PostgreSQL drivers, provider SDKs, frontend contracts, or
  generated transport types.
- Keep business rules, authorization, conflict checks, transactions, audit,
  idempotency, and domain status transitions in the backend. Frontend
  validation and cache state are never authoritative.
- Wrap errors with useful context and preserve the cause. Use standard Go
  naming: lower-case package directories, concise package names, exported
  PascalCase identifiers, and initialism-aware names such as URL, HTTP, UUID,
  and ID.

EMERGING:

- A business vertical slice may use internal/<module>/domain, application, and
  adapter layers only where it owns meaningful behavior. The architecture
  describes this direction, but no business module is implemented yet.
- Keep transport request/response DTOs, application inputs/outputs, domain
  objects, persistence rows, and frontend view models distinct at their
  boundaries. Map them explicitly; do not leak table rows or transport structs
  as domain contracts.
- The existing empty internal/modules directory is baseline, not a required
  module pattern. Do not add empty module shells or refactor it solely for
  symmetry.

### 4.2 HTTP contracts, errors, and compatibility

REQUIRED:

- Keep storefront and Operations OpenAPI contracts separate. A shared
  application service may serve both, but public and Operations DTOs,
  authorization, and data exposure remain separate.
- Update the relevant OpenAPI contract before or with a transport behavior
  change. Give operations an explicit versioned route prefix and preserve the
  existing public/internal boundary.
- Use UUID strings for stable API identifiers. Human-readable references such
  as purchase_ref are lookup/display aids, not substitutes for API identity.
- Return the contract error envelope with code, message, request_id, and
  details for API errors. Server errors must not disclose internals, credentials,
  tokens, evidence, or personal data. The existing httpx writer enforces
  sanitization for 5xx responses.
- Keep successful business responses in the data envelope documented by the
  owning contract. Collection envelopes add page only where that contract
  defines pagination; do not manufacture a generic total or page shape.
- Accept or generate X-Request-ID using the existing bounded format and return
  it on responses. It is correlation metadata, not authorization or an
  idempotency key.
- Treat health and readiness probes as infrastructure exceptions to the
  business response envelope; they currently return a simple status object.

EMERGING:

- The shared client normalizes cancelled, unauthorized, forbidden, validation,
  conflict, unavailable, and unknown errors. Frontend endpoints should preserve
  that taxonomy and render loading, empty, error, stale, and conflict states
  intentionally.
- Retry-sensitive commands use the accepted command-scoped idempotency
  contract. Authenticate and authorize before replay, use domain-specific
  duplicate guards where applicable, and persist replay-sensitive changes in
  the same transaction.

TBD:

- The existing v1 route prefix is not a complete compatibility, deprecation,
  or consumer-versioning policy. Until one is accepted, flag contract
  compatibility risk in the task and update the relevant contract; do not
  promise a migration period or introduce a versioning scheme silently.
- No repository-wide OpenAPI validator or generated-client workflow exists.
  Contract generation, validation tooling, and committed generated clients
  require an approved workflow.

### 4.3 Pagination, filters, identifiers, and time

REQUIRED:

- Operations list contracts use cursor and limit parameters, with limit between
  1 and 100 and an optional next_cursor. Preserve each operation's documented
  deterministic order.
- A successful empty collection is an ordinary successful response; a missing
  entity, validation failure, authorization failure, conflict, dependency
  outage, and unexpected server error remain distinct contract outcomes.
- Database migrations use UUID primary keys and timestamptz values. The schema
  baseline uses created_at and updated_at for mutable aggregates; append-only
  histories, ledgers, audit, outbox, and idempotency records do not gain
  updated_at.
- Store timestamp instants in UTC and expose API instants as OpenAPI date-time
  values. Use exact minor units plus uppercase ISO currency code for money;
  never use floating point for financial amounts.

RECOMMENDED:

- Choose filtering and sorting per endpoint and document the exact shape and
  deterministic order in OpenAPI. Do not impose a universal list model merely
  for consistency.

TBD:

- The public Offering list uses its endpoint-specific opaque cursor with a
  default limit of 50, maximum of 100, deterministic `code ASC, id ASC`
  order, and no total count. That decision does not standardize generic
  Storefront pagination, search/filter encoding, grouped-row pagination, or
  date-only representation.

### 4.4 Persistence, migrations, and seeds

REQUIRED:

- The Go API and PostgreSQL are the transactional system of record. Frontends
  never access PostgreSQL or rely on database column names as an API contract.
- Use a single transaction for a command's authoritative state, required
  histories, audit data, outbox effects, and replay record when those concerns
  apply. Use row locks, atomic conditional updates, constraints, version
  checks, and idempotency according to the owning lifecycle policy.
- Use numbered migration pairs under apps/api/migrations:

      NNNN_name.up.sql
      NNNN_name.down.sql

  Names use lower-case letters, numbers, hyphens, or underscores. Both files
  must be non-empty and share a version/name pair. Migration execution is an
  explicit apps/api/cmd/db operation; API startup never migrates.

- Make an additive migration for an already-shared schema change. Do not edit
  historical migrations after they have been shared or applied. Review both the
  forward and bounded rollback behavior.
- Use the explicit database CLI and its thin Make targets. Rollback is
  destructive, bounded by steps, blocked outside development/test unless the
  explicit opt-in is supplied, and never a hidden default.
- Register seeds with a unique immutable name, group, and order. Reference
  seeds are production-safe; development seeds run only in development/test.
  Changed seed behavior uses a new seed name because seed history is
  append-oriented.

RECOMMENDED:

- Keep SQL and persistence code owned by the module that owns the invariant.
  Introduce repository ports only when a concrete module needs a boundary; do
  not create interfaces with one speculative implementation.

## 5. Frontend conventions

### 5.1 Application structure and naming

The current applications use app, api, layouts, pages, pwa, routes, styles,
and main entrypoint folders. This baseline is valid for the current shells.

EMERGING:

- New vertical slices may introduce feature-oriented folders and entities or
  shared application utilities when real business behavior justifies them.
  Adopt that structure forward; do not migrate the placeholders merely to
  match the architecture example.
- Use PascalCase component and page files, PascalCase component names, camelCase
  hooks and utilities, and a use prefix for hooks. Use descriptive Props names
  rather than an I prefix for component props.

REQUIRED:

- Keep public and Operations application ownership explicit. A Storefront page
  must not expose Operations-only fields or use an Operations contract.
- Keep routes centralized in src/routes/paths.ts and src/routes/routes.tsx.
  Use the path registry rather than scattering route strings through pages.
  A future Operations route guard improves navigation only; backend
  authorization remains the security boundary.
- Keep domain/business rules out of route definitions and UI callbacks.

### 5.2 API access and server state

REQUIRED:

- Keep application endpoint modules under each application's src/api boundary.
  Pages and shared presentation components do not make raw fetch calls.
- Reuse packages/api-client for generic serialization, timeout, cancellation,
  response parsing, and normalized errors. Keep application-specific endpoint
  paths, DTOs, credentials, and policy outside that package.
- Use the respective public or Operations OpenAPI contract before adding an
  endpoint module. Do not manually expose operations DTOs to Storefront.
- Use TanStack Query for real remote API/server state. It owns request
  lifecycle, caching, invalidation after a successful command, and explicit
  loading/error/stale/conflict presentation; it never determines financial,
  quota, allocation, payment, saving-balance, or queue truth.
- Define query keys alongside the endpoint/query boundary and namespace them
  by application and feature. Include stable public identifiers and event
  context when applicable. Do not introduce a global query-key abstraction
  before real endpoint reuse demonstrates the need.

EMERGING:

- Both applications currently mount a QueryClientProvider and use a
  namespaced health diagnostic key with retry disabled. Query stale-time,
  mutation invalidation, session integration, and feature-level key shapes
  remain to be established with the first business vertical slice.
- The development API provider is limited to a deterministic,
  non-authoritative health diagnostic. Do not use it as a substitute for
  backend business behavior or add a broad mock provider without a task.

TBD:

- The frontend API base-path composition for future versioned contract routes
  is not yet exercised by a real endpoint. Define and test it with the owning
  vertical slice instead of guessing a universal path prefix now.

### 5.3 Environment, sessions, and sensitive client state

REQUIRED:

- Server configuration is loaded and validated through the Go config package.
  Production-like environments fail closed when required OIDC and replay-key
  configuration is missing.
- Browser configuration uses only VITE-prefixed build-time values. Those values
  are public by design; never put secrets, private keys, database URLs,
  Operations session material, or Purchase tokens in them.
- Keep .env untracked and use .env.example for names and safe local examples
  only. Never place real credentials, raw sessions, CSRF values, Purchase
  tokens, payment evidence, participant data, or production exports in source,
  fixtures, screenshots, logs, or documentation.
- Keep Storefront Purchase bearer credentials and Operations OIDC sessions
  separate. Operations authentication, Origin, CSRF, permission, and
  revocation checks belong in the backend; frontend guards are not
  authorization controls.
- Do not cache or replay API, authentication, participant, financial, or
  operational data in application PWA service workers. The current PWA policy
  caches immutable shell assets only.

### 5.4 Logging and diagnostic data

REQUIRED:

- Use the existing structured JSON slog foundation for API logs. Keep request
  correlation through the request ID and add module, action, error code, and
  safe latency/context fields when a real behavior needs them.
- Do not log raw credentials, session or CSRF tokens, idempotency keys with
  sensitive payloads, payment evidence, object-storage references, database
  URLs, secrets, or unnecessary participant/purchaser data.
- Keep API error messages safe for their public or Operations audience. Log
  internal causes only through the controlled server logger, not the response
  envelope.

## 6. Quality, dependencies, and generated artifacts

### 6.1 Testing and verification

REQUIRED:

- Use Go's standard testing package for API tests and Vitest for frontend
  tests. Keep focused tests near the behavior they protect and test contract,
  boundary, error, security, concurrency, idempotency, and migration behavior
  when the change introduces that risk.
- Use the canonical repository command:

      make validate

  It performs formatting checks, frontend lint/typecheck/test/build, Go vet,
  Go test/build, and Compose configuration validation. CI additionally proves
  the disposable PostgreSQL migration and reference-seed lifecycle.

- Treat a command as passed only when it was run successfully. Report the exact
  command, output/error, blocker, and whether the failure was caused by the
  task.

RECOMMENDED:

- Prefer the smallest test that proves the changed behavior. Do not add a
  redundant test framework, fixture system, or end-to-end harness before a
  concrete requirement needs it.

### 6.2 Developer commands

REQUIRED:

- Use the repository Makefile as the normal developer-command entrypoint.
  Implemented targets include install, infra-up, infra-down, dev,
  dev-storefront, dev-operations, dev-api, build, lint, typecheck, test,
  format, format-check, compose-check, and validate.
- Treat format as a mutating command and format-check or validate as
  verification commands. Use the documented database targets for migration and
  seed lifecycle instead of ad hoc SQL execution.
- Name destructive database work explicitly. The existing db-rollback and
  db-rollback-steps targets are bounded rollback commands, not general cleanup
  commands; review their environment guard and data impact first.
- Do not document or add a Make target for a command that does not exist.
  Lower-level pnpm and Go commands remain implementation details unless a task
  needs them directly.

### 6.3 Dependencies and generated output

REQUIRED:

- Before adding a dependency, record the problem, why standard library or
  existing dependencies are insufficient, narrowest workspace scope,
  runtime/development impact, maintenance/security impact, and alternatives.
  Record a material architecture choice in docs/DECISIONS.md.
- Do not manually edit lockfiles. Do not add a framework for a trivial utility
  or preemptively install a future dependency.
- Do not commit generated build output, node_modules, coverage, local database
  data, .env files, or generated API clients.

Current generation state:

| Artifact                          | Rule                                                                                                                                      |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Vite/PWA build output             | Generated into ignored dist output; never edit or commit it.                                                                              |
| Shared UI component configuration | packages/ui/components.json may guide shadcn-style additions; emitted source is editable and must receive a root export and focused test. |
| OpenAPI clients                   | No generated client is committed and no generation command exists. Do not introduce one without an approved workflow.                     |

Go source is normalized by `apps/api/cmd/format`, which uses the standard
library formatter for syntax-safe layout and then emits four literal spaces
for leading code indentation. Direct `gofmt` output uses tabs and therefore is
not the repository's final Go formatting result; editor format-on-save should
run the repository command instead. TSX indentation is owned by
`.editorconfig` and the existing Prettier workflow.

## 7. Enforcement matrix

| Convention                                         | Level                          | Enforcement                                | Current state                                                         |
| -------------------------------------------------- | ------------------------------ | ------------------------------------------ | --------------------------------------------------------------------- |
| Formatting                                         | REQUIRED                       | Prettier, apps/api/cmd/format, Make, CI    | Mechanically enforced; Go output uses literal four-space indentation. |
| Frontend lint and type safety                      | REQUIRED                       | Oxlint, strict TypeScript config, Make, CI | Mechanically enforced.                                                |
| Go vet, test, and build                            | REQUIRED                       | Make and CI                                | Mechanically enforced.                                                |
| Compose configuration                              | REQUIRED                       | Make and CI                                | Mechanically enforced.                                                |
| Migration naming, pairs, and non-empty files       | REQUIRED                       | database CLI validation and CI             | Mechanically enforced.                                                |
| Migration and reference-seed lifecycle             | REQUIRED                       | disposable PostgreSQL CI job               | Mechanically enforced.                                                |
| No committed local secrets/build output            | REQUIRED                       | .gitignore, review, CI-adjacent checks     | Partially mechanical; no secret scanner is installed.                 |
| Separate public and Operations contracts           | REQUIRED                       | repository layout and review               | Structural/review enforcement; no OpenAPI linter is installed.        |
| Backend-owned business truth and database boundary | REQUIRED                       | architecture, code review, focused tests   | Review/test enforcement.                                              |
| Request-ID and sanitized HTTP errors               | REQUIRED                       | httpx platform utility and handler tests   | Partially mechanical; future handlers must adopt it.                  |
| Query-key ownership and cache invalidation         | REQUIRED for real remote state | code review and focused tests              | Early adoption only.                                                  |
| Package promotion boundary                         | REQUIRED                       | code review                                | Review enforcement.                                                   |

Prefer low-cost, reliable mechanical enforcement. Do not add elaborate tooling
only to enforce stylistic preference.

## 8. Review checklists

### Backend or API change

- Boundary and owning module are clear.
- Contract changes use the correct public or Operations document.
- Transport, application, domain, and persistence mappings are intentional.
- Backend owns invariants, authorization, and contested-state checks.
- Errors, request IDs, audit, idempotency, transactions, and migrations are
  considered where applicable.
- Focused tests and applicable repository validation ran.

### Frontend change

- Application and route ownership are correct.
- No page or shared UI component performs raw transport work.
- The matching API contract and endpoint boundary are used.
- Server state uses the canonical query mechanism when applicable.
- Loading, empty, error, stale, and conflict states are intentional.
- No public page exposes internal fields or sensitive state.

### Database change

- Additive migration pair and rollback effect were reviewed.
- Migration command and seed responsibility are correct.
- Application compatibility, transaction, concurrency, and audit impact were
  checked.
- Migration validation and the applicable lifecycle checks ran.

### Documentation or agent task

- The canonical owner is correct and duplicate rule text was avoided.
- Existing state, assumptions, conflicts, and TBD items are explicit.
- Scope stays bounded and plan variances are listed.
- Paths, commands, and cross-references were verified.

## 9. Evolution, known gaps, and follow-up decisions

Add or revise a convention when a pattern recurs, inconsistency creates
maintenance cost, a boundary needs protection, an enforcement mechanism is
introduced, or a stable application/package workflow appears. Remove or revise
a convention that no longer matches source or architecture. Do not add a rule
for a one-off preference.

The following are intentionally not normalized by this document:

| Area                              | Current evidence and required follow-up                                                                                                                                                                                                                        |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Frontend feature layout           | Architecture recommends feature-oriented structure, while the shells use pages, layouts, and api. Apply feature organization only with a real vertical slice.                                                                                                  |
| TanStack Query adoption           | ADR-039 describes deferred installation, while both applications now install and mount it for diagnostics. Treat detailed server-state policy as EMERGING until a business slice establishes it; reconcile ADR/status wording separately.                      |
| Authentication and runtime status | Source contains an OIDC route registrar and validation foundation, while some status-oriented documents still describe it as unimplemented. Correct those status documents in a focused state-documentation task rather than silently relying on either claim. |
| Database CI wording               | Current CI exercises the disposable migration/seed lifecycle, while older database documentation describes that job as future work. Reconcile documentation in a focused state-documentation task.                                                             |
| API compatibility and generation  | No complete deprecation/versioning policy, OpenAPI validator, or generated-client workflow exists. These require explicit future decisions.                                                                                                                    |
| Pagination and filters            | Operations cursor pagination is specified; Storefront and generic search/filter policy remain endpoint-specific and incomplete.                                                                                                                                |

Do not resolve these gaps through broad refactoring or speculative tooling.
