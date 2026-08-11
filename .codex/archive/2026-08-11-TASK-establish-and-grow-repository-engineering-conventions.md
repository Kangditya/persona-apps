# Task: Establish and Grow Repository Engineering Conventions

## Executed

## Verified

## Objective

Establish a durable, repository-wide engineering convention system for the monorepo so that humans and coding agents have one clear, evolving rulebook for how the project is structured, changed, tested, documented, and operated.

The primary deliverable is a canonical conventions document:

```text
docs/CONVENTIONS.md
```

The task must not merely copy current habits into prose. It must:

1. discover the conventions that already exist in the repository;
2. reconcile them with the intended architectural direction described in the source-of-truth documents;
3. formalize missing conventions where the project already has a clear preferred direction;
4. identify ambiguous or conflicting conventions instead of silently choosing one;
5. connect conventions to practical enforcement mechanisms where justified;
6. make the convention system easy to extend as the monorepo grows;
7. avoid broad refactoring solely to make existing code conform to the newly documented rules.

The result should function as the engineering rulebook for future backend, frontend, database, infrastructure, testing, documentation, and agent-driven work.

---

## Core Principle

The convention system must preserve the following ownership model:

```text
Product requirements define what should exist.
Architecture defines structural boundaries.
Decisions explain why important choices were made.
Contracts define communication boundaries.
Applications own application behavior.
Packages own genuine reusable capabilities.
Backend owns authoritative business truth.
Tests verify behavior.
Conventions define how work is performed consistently.
TASK.md defines one bounded change only.
```

No lower-level document may silently override a higher-level architectural or product decision.

---

## Required Workflow

```text
DISCOVER → PLAN → REVIEW PLAN → BUILD → VERIFY → REVIEW
```

### DISCOVER

Before writing or changing conventions:

* inspect the repository root and workspace structure;
* inspect all existing application directories under `apps/`;
* inspect reusable packages under `packages/`;
* inspect API contracts under `contracts/`;
* inspect repository documentation under `docs/`;
* inspect `.codex/AGENTS.md`, `.codex/CURRENT_STATE.md`, and any nested `AGENTS.md` files;
* inspect root and application/package-level `README.md` files where relevant;
* inspect the root `Makefile` and any nested Makefiles;
* inspect package manager/workspace configuration;
* inspect Go workspace/module files and established Go project structure;
* inspect frontend TypeScript/React project structure;
* inspect lint, formatting, typecheck, test, build, and CI configuration;
* inspect database migration and seeding infrastructure if present;
* inspect environment configuration patterns and `.env.example` files;
* inspect existing route, API client, DTO/model, repository, service/use-case, error, and test patterns;
* inspect OpenAPI naming and serialization conventions;
* identify conventions that are already enforced by tooling rather than documentation;
* identify inconsistent patterns that require a decision instead of guessing.

Prefer repository evidence over assumptions.

### PLAN

Before modifying files, produce a plan containing:

```text
Summary
Observed repository conventions
Proposed canonical conventions
Conflicts / inconsistencies
Decisions required
Planned file-change table
Enforcement opportunities
Documentation cross-reference changes
Out-of-scope refactors
Verification plan
Risks
```

The planned file-change table must include at minimum:

| File | Action | Purpose | Risk |
|---|---|---|---|
| `docs/CONVENTIONS.md` | create/update | canonical engineering rulebook | low |
| `.codex/AGENTS.md` | inspect / minimal update if needed | reference canonical conventions without duplicating them | medium |
| `docs/ARCHITECTURE.md` | inspect / targeted update only if needed | resolve structural inconsistency | medium |
| `docs/DECISIONS.md` | inspect / targeted update only if needed | record architectural decisions, not ordinary conventions | medium |
| root `README.md` | optional | expose conventions entrypoint | low |
| `Makefile` | optional | enforce canonical developer commands where justified | medium |

Do not begin implementation until the plan is complete and explicitly approved.

### REVIEW PLAN

Before BUILD, verify that the plan:

* does not invent repository structure that does not exist;
* does not convert preferences into mandatory rules without evidence or rationale;
* distinguishes architecture decisions from coding conventions;
* does not require broad rewrites of working code;
* preserves public/internal API boundaries;
* preserves application/package ownership boundaries;
* defines how exceptions and future evolution are handled;
* includes a migration path for conventions that cannot be enforced immediately.

---

## Source of Truth

Inspect and preserve the intent of:

```text
docs/PRD.md
docs/PRODUCT_MAP.md
docs/ARCHITECTURE.md
docs/DECISIONS.md
docs/CONVENTIONS.md          # canonical after this task
.codex/CURRENT_STATE.md
.codex/AGENTS.md
TASK.md
```

### Intended responsibility hierarchy

```text
PRD.md
  What the product must do.

PRODUCT_MAP.md
  Product/domain capability map.

ARCHITECTURE.md
  Structural architecture and system boundaries.

DECISIONS.md
  Why significant architectural or technology choices were made.

CONVENTIONS.md
  How engineers and agents consistently implement work within those boundaries.

CURRENT_STATE.md
  What currently exists and what is incomplete.

AGENTS.md
  Operational instructions for coding agents.

TASK.md
  One bounded piece of work.
```

Do not duplicate large bodies of text across these files. Prefer cross-references.

---

# Scope

## In Scope

Formalize and document repository conventions for:

1. repository structure and ownership;
2. source-of-truth hierarchy;
3. development workflow;
4. planning and plan approval;
5. smallest-diff discipline;
6. application boundaries;
7. shared package boundaries;
8. backend layering;
9. frontend feature organization;
10. API access and transport;
11. public vs operations API separation;
12. API contracts and versioning;
13. DTO/model/view-model transformations;
14. Go naming and package conventions;
15. TypeScript/React naming and module conventions;
16. routing ownership;
17. server-state ownership;
18. query-key and cache invalidation conventions;
19. error taxonomy and propagation;
20. environment configuration;
21. authentication/session integration;
22. logging and sensitive-data handling;
23. testing strategy;
24. development/test mocks and providers;
25. database access boundaries;
26. database migration conventions;
27. seeding conventions;
28. Makefile/developer-command conventions;
29. documentation responsibilities;
30. `TASK.md` structure and lifecycle;
31. agent behavior and plan-variance reporting;
32. dependency introduction rules;
33. generated-code rules;
34. compatibility and deprecation rules;
35. convention exceptions and evolution;
36. enforcement through lint/test/build/CI where appropriate.

## Out of Scope

* broad codebase refactoring solely to satisfy the new convention document;
* redesigning product behavior;
* inventing qurban domain capabilities;
* implementing missing backend features merely to demonstrate a convention;
* replacing established frameworks without a separately approved architectural decision;
* adding speculative microservices, multitenancy, event infrastructure, observability stacks, or deployment infrastructure;
* renaming large portions of the repository for aesthetic consistency;
* changing database schemas unless directly necessary for an already-approved convention mechanism;
* committing or pushing changes.

---

# Canonical Convention Model

`docs/CONVENTIONS.md` must classify conventions so future engineers can distinguish hard rules from guidance.

Use a clear status model such as:

```text
REQUIRED      Must be followed unless an explicit exception is documented.
RECOMMENDED   Preferred default; deviations require a concrete reason.
EMERGING      Direction is established but repository adoption is incomplete.
TBD           Repository evidence is insufficient; requires a future decision.
```

Each meaningful convention should include, where useful:

```text
Rule
Rationale
Scope
Example
Anti-pattern
Enforcement
Exception process
Related architecture/decision reference
```

Do not make every trivial style preference verbose. Reserve detailed rationale for conventions that protect architectural boundaries, correctness, security, maintainability, or agent behavior.

---

# 1. Repository Structure Convention

The conventions must define ownership of repository-level locations after verifying actual structure.

Target conceptual model:

```text
apps/
  api/                    # authoritative backend application
  operations-web/         # protected/internal operations frontend
  storefront-web/         # public/customer frontend

packages/
  api-client/             # reusable transport/client primitives where justified
  contracts/              # reusable contract-derived types where justified
  ui/                     # genuinely reusable UI components if present/approved
  ...

contracts/
  openapi/
    storefront.yaml       # public API contract
    operations.yaml       # internal operations API contract

docs/
  PRD.md
  PRODUCT_MAP.md
  ARCHITECTURE.md
  DECISIONS.md
  CONVENTIONS.md

.codex/
  AGENTS.md
  CURRENT_STATE.md
```

Required rule:

> `apps/*` owns application behavior. `packages/*` owns genuinely reusable capabilities. Reuse must not erase domain or public/internal boundaries.

Before moving code into `packages/`, require evidence that the abstraction is application-neutral and stable enough to share.

---

# 2. Source-of-Truth and Documentation Convention

Formalize this hierarchy:

```text
Product requirements
        ↓
PRD / PRODUCT_MAP
        ↓
Architecture
        ↓
Architectural Decisions
        ↓
Engineering Conventions
        ↓
Application implementation
        ↓
Tests / verification
        ↓
Task-specific execution
```

A `TASK.md` must not silently override `ARCHITECTURE.md` or `DECISIONS.md`.

If a task requires an architectural change:

```text
identify conflict
    ↓
record/approve decision
    ↓
update architecture/decision documentation
    ↓
implement
```

---

# 3. Development Workflow Convention

Canonical workflow for non-trivial work:

```text
DISCOVER → PLAN → REVIEW PLAN → BUILD → VERIFY → REVIEW
```

### DISCOVER must answer

* What exists?
* What is the nearest canonical implementation?
* What boundaries are affected?
* What contracts are authoritative?
* What assumptions remain?

### PLAN must define

* intended behavior;
* files expected to change;
* dependency impact;
* contract impact;
* migration impact where applicable;
* test strategy;
* risks;
* unresolved decisions.

### BUILD must

* follow the approved plan;
* preserve smallest coherent diff;
* record variance before expanding scope.

### VERIFY must

* run focused checks first;
* run repository-level checks where applicable;
* report exact command failures and blockers.

### REVIEW must

* compare implementation with the approved plan;
* identify variances;
* check for boundary leaks;
* check tests and documentation;
* identify remaining risks.

---

# 4. Smallest-Diff Convention

Canonical rule:

> Change the smallest coherent surface that satisfies the approved requirement.

Do not combine feature work with unrelated cleanup, renaming, dependency replacement, formatting churn, or architectural refactors.

When adjacent technical debt is discovered:

```text
record it
    ↓
continue current bounded task
    ↓
create/recommend a separate follow-up task
```

Exceptions are permitted only when the adjacent change is necessary for correctness, buildability, security, or completion of the approved requirement.

---

# 5. Application Boundary Convention

## `apps/api`

The backend is authoritative for:

* business invariants;
* authorization;
* persistence;
* transactions;
* workflow state;
* authoritative validation;
* financial calculations;
* API contracts;
* concurrency-sensitive decisions.

## `apps/storefront-web`

Storefront owns public presentation and interaction only.

It must not:

* access the database directly;
* depend on operations-only fields;
* import Operations Web modules;
* implement authoritative business rules;
* assume purchaser and Sohibul Qurban are always the same actor unless the product contract explicitly establishes that invariant.

## `apps/operations-web`

Operations Web owns internal operational presentation/workflows.

It must not:

* treat client-side authorization as security enforcement;
* bypass backend authorization;
* calculate authoritative balances or workflow state locally;
* mutate state through reporting/analytics interfaces unless explicitly contracted.

---

# 6. Shared Package Convention

Before placing code in `packages/*`, verify that it is genuinely reusable.

Good shared candidates:

```text
HTTP transport primitives
contract-derived common types where boundaries permit
UI primitives with no application/domain ownership
shared build/lint configuration
stable utility code used identically across applications
```

Poor shared candidates:

```text
storefront-specific workflows
operations-specific authorization behavior
domain services owned by apps/api
feature modules that merely look similar
premature abstractions created for hypothetical reuse
```

Required rule:

> Duplicate a small amount of application-specific code before creating a shared abstraction that erases ownership boundaries.

---

# 7. Backend Layering Convention

Discover the existing Go architecture and document the actual canonical dependency direction.

The convention should preserve a structure conceptually similar to:

```text
HTTP / transport
      ↓
application / use-case
      ↓
domain
      ↓
repository interface
      ↓
persistence adapter
      ↓
database
```

Document where the repository actually places:

* handlers/controllers;
* request/response DTOs;
* application services/use-cases;
* domain entities and invariants;
* repository interfaces;
* persistence implementations;
* configuration/bootstrap wiring.

Do not force a textbook architecture if the repository uses a different but coherent structure.

---

# 8. Frontend Feature Convention

Discover the current frontend organization and document one preferred pattern for new features.

A candidate feature structure may resemble:

```text
features/<feature>/
  api/
    <feature>.api.ts
    <feature>.keys.ts
    <feature>.types.ts
  hooks/
  components/
  routes/
  utils/
```

Use this only if compatible with existing repository structure.

The convention must define:

* what belongs in application-level `lib/`;
* what belongs in a feature;
* what belongs in `packages/`;
* where route components live;
* where API endpoint modules live;
* where feature-specific types live;
* where reusable UI primitives live.

---

# 9. Frontend API Access Convention

Pages and UI components must not own raw transport details.

Preferred dependency flow:

```text
Page / Feature
      ↓
Query or Mutation Hook
      ↓
Endpoint Module
      ↓
Application/Public API Client
      ↓
Shared HTTP Transport
      ↓
HTTP
```

Avoid:

```text
Page → raw fetch → endpoint string
Component → token lookup → request
Component → response-envelope parsing
```

Centralize:

* endpoint paths;
* request serialization;
* response parsing;
* authentication/session integration;
* cancellation/timeout;
* normalized transport errors.

---

# 10. Public vs Operations API Convention

The public and internal API surfaces must remain intentionally separate.

Canonical contracts:

```text
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
```

The implementation should expose clearly separate client boundaries, for example conceptually:

```text
storefrontClient
operationsClient
```

Do not create a generic application-wide client that makes accidental crossing of public/internal boundaries trivial.

---

# 11. Contract-First Convention

When consuming an existing endpoint:

* conform to the approved contract;
* preserve response/error semantics unless an explicit adapter exists;
* avoid manually duplicating generated/approved contract representations.

When introducing or changing an endpoint, document:

```text
method
path
authentication
request parameters/body
success response
error responses
pagination/filtering/sorting
timestamp/timezone representation
compatibility/deprecation impact
```

If backend behavior is unavailable or undecided:

```text
TBD
```

or:

```text
not implemented
```

Do not fabricate production behavior.

---

# 12. DTO, Domain Model, and View Model Convention

Document clear transformation boundaries.

Conceptually:

```text
API Contract DTO
      ↓
optional explicit adapter
      ↓
application/domain or view model
```

Avoid silently changing transport semantics inside generic clients.

Where transformations are necessary, prefer explicit named transformations such as:

```text
toDomain(...)
toResponse(...)
toViewModel(...)
```

Use repository-established terminology when it already exists.

---

# 13. Go Naming Convention

Verify and document actual repository conventions for:

```text
package names
exported/unexported identifiers
request/response DTOs
repository interfaces
application services/use-cases
constructors
errors
constants
files
_test.go placement
```

Default target where compatible:

```text
package               lowercase
exported identifiers  PascalCase
unexported identifiers camelCase
request DTO            XxxRequest
response DTO           XxxResponse
repository              XxxRepository
service/use-case        follow existing application naming
```

For JSON field naming, preserve the canonical API contract. If the repository standard is `snake_case`, document that as the default and identify explicit exceptions rather than mixing naming styles accidentally.

---

# 14. TypeScript / React Naming Convention

Verify and document the existing convention for:

```text
components
hooks
API modules
query-key modules
types
schemas
utilities
routes
constants
```

Candidate defaults where compatible:

```text
Component.tsx          PascalCase component file if repository uses this style
use-thing.ts           hooks
thing.api.ts           endpoint module
thing.keys.ts          query keys
thing.types.ts         feature types
thing.schema.ts        runtime validation schema
thing.utils.ts         local utilities
```

Identifiers:

```text
functions/variables    camelCase
types/interfaces       PascalCase
static constants       UPPER_SNAKE_CASE where appropriate
```

Do not rewrite existing files merely to normalize naming.

---

# 15. Routing Convention

React Router remains the navigation and route-registration authority unless an approved architectural decision changes it.

Document:

* route declaration location;
* protected/public route handling;
* route parameter ownership;
* search/query parameter handling;
* navigation helpers;
* feature-route boundaries.

Do not introduce a competing routing abstraction.

---

# 16. Server-State Convention

TanStack Query is the preferred owner of remote/server state once a real API slice is integrated.

Remote state includes:

```text
API responses
request lifecycle
cache
refetching
staleness
mutations
invalidation
```

Do not mirror server state into React Context, Zustand, Redux, or ad-hoc local state without a concrete reason.

Document when local component state is appropriate versus server state.

---

# 17. Query-Key and Cache Convention

Query keys must be deterministic and centralized per feature/domain.

Document:

* root query keys;
* list/detail key derivation;
* parameter inclusion;
* invalidation after mutation;
* stale-time defaults where established;
* sensitive operations-data caching restrictions;
* cancellation and stale-update behavior.

Avoid scattered string-literal query keys.

---

# 18. Error Convention

Define a normalized application-facing error taxonomy based on existing backend/transport behavior.

Target categories where applicable:

```text
validation
unauthorized
forbidden
not_found
conflict
unavailable
timeout
cancelled
unknown
```

Document:

* where normalization happens;
* which backend fields must be preserved;
* how field validation errors are represented;
* how UI distinguishes retryable vs non-retryable states;
* how unexpected errors are logged without leaking sensitive data.

Do not inspect raw error-message strings in UI components when a structured error type can be used.

---

# 19. Environment Configuration Convention

Document:

* environment-variable naming;
* frontend Vite exposure rules;
* backend environment configuration;
* `.env.example` ownership;
* local-development defaults;
* secret-handling rules.

Required rules:

```text
.env.example       may be committed
.env               must not be committed
.env.local         must not be committed
credentials        must not be embedded in source
```

Environment variables configure deployment/runtime differences. They must not become a substitute for authoritative product/business configuration.

---

# 20. Authentication and Session Convention

Authentication/session integration should be centralized at the transport/application boundary.

Frontend pages/components must not independently retrieve credentials and construct authentication headers for ordinary API calls.

Document:

* credential/session source;
* header/cookie integration point;
* unauthorized handling;
* refresh/session-expiry behavior if currently implemented;
* what remains TBD if authentication is not finalized.

Never log:

```text
access tokens
refresh tokens
cookies
passwords
payment credentials
sensitive personal data
full sensitive request bodies
```

---

# 21. Logging Convention

Discover existing logging infrastructure and document rules for:

* log levels;
* request correlation identifiers if present;
* structured logging fields;
* sensitive-data redaction;
* expected error vs unexpected error logging;
* frontend development logging;
* production logging restrictions.

Do not introduce a logging framework solely for documentation consistency unless justified and approved.

---

# 22. Testing Convention

Document the testing pyramid actually supported by the repository.

Preferred principle:

```text
many focused unit tests
fewer integration tests
few end-to-end tests for critical flows
```

For API-facing features, cover as applicable:

```text
success
empty result
validation failure
unauthorized
forbidden
not found
conflict
backend unavailable
unexpected error
timeout/cancellation
mutation invalidation
```

Tests must not require production credentials.

Use existing repository tooling. Do not introduce a second testing framework without an explicit decision.

---

# 23. Mock / Development Provider Convention

Mocks are development/test infrastructure, not product truth.

Preferred shape:

```text
interface / endpoint boundary
      ├── real HTTP implementation
      └── deterministic test/dev implementation
```

Rules:

* mock behavior must be clearly isolated;
* mock data must not be presented as verified backend behavior;
* unknown business fields remain TBD rather than invented;
* tests must be deterministic;
* production builds must not accidentally depend on development-only providers.

---

# 24. Database Boundary Convention

Required dependency direction:

```text
Frontend
   ↓
API
   ↓
Application / Domain
   ↓
Repository / Persistence
   ↓
Database
```

Frontend applications must never access the database directly.

Document backend rules for:

* SQL ownership;
* transaction boundaries;
* repository/persistence boundaries;
* mapping database rows into application/domain models;
* avoiding business-rule leakage into SQL unless intentionally designed.

Use existing backend architecture as the source of truth.

---

# 25. Migration Convention

Inspect the migration tooling already introduced or planned in the Go project.

Document a canonical migration lifecycle covering, where supported:

```text
create migration
status
apply/up
rollback/down
verify
```

Migration files must be:

* ordered deterministically;
* immutable after being applied to shared environments unless the repository explicitly supports amendment;
* reviewable;
* safe against accidental destructive behavior where practical;
* separate from runtime application startup unless an explicit architecture decision says otherwise.

Do not invent command names if the repository already has canonical commands.

---

# 26. Seeder Convention

Document the distinction between:

```text
schema migration
reference/bootstrap seed
local-development seed
test fixture
production data migration
```

Seeder behavior must be explicit about:

* environment applicability;
* idempotency;
* dependencies/order;
* destructive behavior;
* ownership;
* rollback/reset expectations.

Do not treat arbitrary production data changes as development seeders.

---

# 27. Makefile / Developer Command Convention

Treat the root Makefile as the preferred human- and agent-facing command interface when the repository already follows this direction.

Discover existing targets before adding new ones.

Potential canonical categories include:

```text
make setup
make dev
make lint
make typecheck
make test
make build
make validate
make migrate-up
make migrate-down
make migrate-status
make seed
```

Only document or add commands that are actually implemented or approved.

Underlying commands such as `pnpm`, `go test`, migration binaries, or scripts may remain implementation details.

A Make target should not hide important destructive behavior without an explicit name and documentation.

---

# 28. Dependency Convention

Before adding a dependency, document:

```text
problem being solved
why existing dependencies/utilities are insufficient
runtime vs development dependency
bundle/runtime impact where relevant
maintenance/security impact
alternative considered
```

Rules:

* reuse existing dependencies before adding another library with overlapping responsibility;
* do not add a framework to solve a single trivial utility problem;
* record architectural dependencies in `DECISIONS.md` when the choice materially shapes the system;
* ordinary implementation dependencies do not require an ADR unless they change architecture.

---

# 29. Generated Code Convention

Identify any generated artifacts in the repository.

Document:

* source file;
* generation command;
* whether generated output is committed;
* whether manual editing is forbidden;
* verification for stale generated artifacts.

Do not introduce generated API clients unless there is an approved generation workflow.

---

# 30. Compatibility and Deprecation Convention

Establish conventions for externally consumed or cross-application contracts.

Document:

* backward-compatible changes;
* breaking changes;
* API versioning strategy if one exists;
* deprecation markers/documentation;
* migration period expectations;
* database compatibility during rolling changes if relevant.

If the repository has no versioning policy yet, mark it `TBD` rather than inventing one silently.

---

# 31. Documentation Convention

Canonical responsibilities:

| Document | Responsibility |
|---|---|
| `README.md` | human entrypoint, setup, and navigation |
| `PRD.md` | product requirements |
| `PRODUCT_MAP.md` | product/domain capability map |
| `ARCHITECTURE.md` | system structure and boundaries |
| `DECISIONS.md` | significant architecture/technology decision rationale |
| `CONVENTIONS.md` | engineering implementation rules and defaults |
| `CURRENT_STATE.md` | implemented/incomplete repository state |
| `AGENTS.md` | coding-agent operational rules |
| `TASK.md` | one bounded work item |

Avoid copying the same rule into several documents. Reference the canonical owner instead.

---

# 32. TASK.md Convention

Establish a canonical structure for substantial work items.

Target template:

```markdown
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
```

### Existing State rule

Anything explicitly documented as existing/manual/pre-task state is baseline.

Agents must not revert it merely because it differs from an older pattern.

### Task lifecycle

Use a consistent status vocabulary after discovering current repository practice, for example:

```text
Draft
Ready for Planning
Plan Approved
In Progress
Blocked
Executed
Verified
```

Do not add status complexity if the repository does not benefit from it.

---

# 33. Agent Convention

Canonical agent principles:

```text
Inspect before assuming.
Reuse before abstracting.
Plan before modifying.
Prefer the smallest coherent diff.
Verify actual behavior.
Report failures exactly.
Do not silently expand scope.
Do not invent unavailable product/backend behavior.
Do not commit or push unless explicitly requested.
```

Agents must distinguish:

```text
verified fact
repository inference
assumption
TBD
```

when that distinction materially affects implementation.

---

# 34. Plan Variance Convention

If BUILD requires an unplanned file or behavior change:

```text
planned work
    ↓
unexpected required change
    ↓
record PLAN VARIANCE
    ↓
explain necessity and impact
    ↓
continue only if required to satisfy the approved task
```

The final report must list all plan variances.

Unrelated discoveries become recommended follow-up tasks.

---

# 35. Convention Exception Convention

The convention system must permit deliberate exceptions without creating silent inconsistency.

For a meaningful exception, document:

```text
convention being deviated from
reason
scope
risk
whether temporary or permanent
follow-up if temporary
```

Do not require exception paperwork for trivial implementation choices.

Architectural exceptions may require a `DECISIONS.md` entry; local implementation exceptions generally belong in the relevant task/code documentation.

---

# 36. Convention Evolution Convention

`docs/CONVENTIONS.md` is a living document.

New conventions should be added when:

* the same implementation decision recurs across features;
* repeated inconsistencies create maintenance cost;
* a boundary needs protection;
* tooling begins enforcing a behavior;
* a new application/package/workflow introduces a stable pattern.

Do not add conventions for one-off personal preferences.

Periodically remove or revise conventions that no longer match the architecture or tooling.

---

# 37. Enforcement Convention

For each REQUIRED convention, identify whether enforcement is:

```text
Documentation only
Code review
Lint/formatter
Type system
Tests
Build
Contract validation
Make target
CI
Database constraint
Runtime guard
```

Prefer mechanical enforcement where it is low-cost and reliable.

Do not add elaborate tooling merely to enforce stylistic preferences.

Create an enforcement matrix in `docs/CONVENTIONS.md` for high-value rules, for example:

| Convention | Level | Enforcement | Current State |
|---|---|---|---|
| no frontend DB access | REQUIRED | architecture/review | enforced structurally |
| public/internal OpenAPI separation | REQUIRED | contract structure/test where available | active |
| no committed secrets | REQUIRED | gitignore/review/CI if available | inspect |
| lint/typecheck before merge | REQUIRED | Makefile/CI | inspect |
| query keys centralized | RECOMMENDED/REQUIRED by app maturity | code review/tests | inspect |

Use repository evidence for the final values.

---

# 38. Security and Sensitive Data Convention

Formalize baseline rules for:

* credentials;
* secrets;
* access/session tokens;
* cookies;
* payment data;
* participant/purchaser personal data;
* logs;
* test fixtures;
* screenshots/debug artifacts;
* environment files.

Security-sensitive rules should be REQUIRED and mechanically enforced where practical.

Do not invent compliance certifications or regulatory requirements not established by the project.

---

# 39. Time, Date, and Identifier Convention

Inspect how the repository handles:

```text
UUIDs
numeric database IDs
public identifiers
timestamps
timezones
dates without time
created_at / updated_at / deleted_at
```

Document the canonical API representation and where conversion occurs.

If the repository has unresolved inconsistencies, list them as explicit decisions/TBDs rather than normalizing them silently.

---

# 40. Pagination, Filtering, and Sorting Convention

Inspect existing API behavior and formalize reusable rules where stable.

Document as applicable:

* page/limit vs cursor strategy;
* total-count semantics;
* grouped-row pagination semantics;
* filter encoding;
* search whitespace/special-character handling;
* deterministic sorting;
* empty-result semantics;
* frontend query parameter representation.

Do not impose one universal pagination strategy if different APIs intentionally require different models.

---

# 41. API Response and Empty-State Convention

Document existing response-envelope conventions.

Where supported by current contracts, distinguish:

```text
successful empty collection
not found entity
validation error
unauthorized/forbidden
backend unavailable
unexpected server error
```

Do not return errors for ordinary empty collections merely to simplify frontend handling unless the API contract explicitly requires that behavior.

---

# 42. Review Checklist Convention

`docs/CONVENTIONS.md` should end with compact review checklists for common work types.

At minimum:

### Backend/API change

```text
boundary respected
contract updated if required
DTO mapping explicit
business invariant remains backend-owned
persistence isolated
errors preserved/normalized intentionally
tests added
migration impact checked
```

### Frontend change

```text
route ownership respected
no raw transport in page/component
server state uses canonical mechanism
public/internal API boundary respected
loading/empty/error states covered
no secrets/internal fields exposed
focused tests added
```

### Database change

```text
migration created
rollback strategy considered
seed vs migration responsibility correct
application compatibility checked
migration command verified
```

### Documentation/agent task

```text
source-of-truth owner correct
no duplicated rules
existing state recorded
scope bounded
plan variances reported
verification commands exact
```

Keep checklists short enough to remain usable.

---

# Expected Deliverables

The completed task must produce:

1. `docs/CONVENTIONS.md` as the canonical engineering convention document;
2. a documented convention classification system (`REQUIRED`, `RECOMMENDED`, `EMERGING`, `TBD` or an approved equivalent);
3. repository/source-of-truth hierarchy documentation;
4. application/package ownership rules;
5. backend and frontend implementation conventions;
6. API/contract rules;
7. database migration and seeding conventions based on actual repository capabilities;
8. Makefile/developer-command conventions based on actual targets;
9. testing and error-handling conventions;
10. agent/TASK workflow conventions;
11. convention exception/evolution rules;
12. enforcement matrix for high-value rules;
13. minimal cross-reference updates to `.codex/AGENTS.md`, root `README.md`, `ARCHITECTURE.md`, or `DECISIONS.md` only when justified by DISCOVER/PLAN;
14. optional canonical task template only if the repository lacks one and the plan approves its creation;
15. final report of discovered conflicts, deferred decisions, and recommended future convention work.

---

# Acceptance Criteria

The task is complete only when:

1. engineers can identify the canonical owner of product, architecture, decision, convention, state, agent, and task documentation;
2. `docs/CONVENTIONS.md` clearly distinguishes mandatory rules from recommendations and unresolved areas;
3. repository structure and dependency boundaries are documented using actual repository evidence;
4. Storefront, Operations, API, and shared-package ownership rules are explicit;
5. public and operations API contracts remain separate;
6. frontend remote-state, API-access, and error-handling conventions are explicit;
7. backend layering and DTO/persistence conventions reflect the actual Go codebase;
8. migration/seeder conventions reflect actual tooling or are clearly marked TBD;
9. Makefile conventions distinguish implemented targets from desired/future targets;
10. testing conventions use existing tooling and do not introduce redundant frameworks;
11. no secrets, credentials, or sensitive data are added to documentation/examples;
12. no broad source-code refactor is performed merely for convention compliance;
13. conflicts between existing patterns are documented rather than hidden;
14. significant architecture changes are routed through `DECISIONS.md` instead of being buried in conventions;
15. `AGENTS.md` references conventions rather than duplicating large sections;
16. the convention document explains how to add, change, deprecate, or intentionally violate a convention;
17. high-value REQUIRED conventions identify practical enforcement mechanisms;
18. final verification commands are executed and failures are reported exactly;
19. final report includes implemented, verified, assumed, deferred, and TBD items;
20. future tasks can reference `docs/CONVENTIONS.md` instead of restating repository-wide engineering rules.

---

# Verification

Run applicable repository checks after documentation/tooling changes:

```bash
make validate
pnpm run lint
pnpm run typecheck
pnpm run test
pnpm run build
go test ./...
```

Use only commands applicable to the actual workspace discovered.

If the repository provides canonical Make targets, prefer them over duplicate lower-level commands.

For documentation-only changes, still verify:

* referenced paths exist;
* documented commands exist;
* documented package/application names match the repository;
* examples do not contradict contracts;
* links/cross-references are valid;
* no generated or secret files are accidentally included.

If a command cannot run, report:

```text
command
exit/error
exact blocker
whether caused by this task
```

Do not mark an unexecuted check as passed.

---

# Final Review Requirements

The final review must explicitly answer:

```text
What conventions were already present?
What conventions were newly formalized?
What existing patterns conflict?
Which conventions are REQUIRED?
Which remain RECOMMENDED or EMERGING?
Which decisions remain TBD?
What is mechanically enforced today?
What is documentation/review-only?
What files changed?
Were any architecture decisions changed?
Were any dependencies added?
Were any source-code files refactored?
What plan variances occurred?
What risks remain?
What convention area should be standardized next?
```

---

# Constraints

* Prefer repository evidence over generic best practices.
* Do not rewrite working code merely to satisfy documentation.
* Do not invent repository capabilities, commands, contracts, or architecture.
* Do not turn every coding preference into a mandatory convention.
* Preserve the smallest coherent diff.
* Keep product/domain truth in the authoritative backend and product documents.
* Keep public and operations boundaries explicit.
* Keep application-specific concerns out of shared packages unless justified.
* Avoid duplicate documentation; link to canonical owners.
* Treat manually existing state as baseline unless the task explicitly changes it.
* Record plan variance before touching unplanned files.
* Do not commit or push changes.

---

# Recommended Follow-Up Direction

After this task is complete, future convention work should be incremental rather than another broad rewrite.

Potential follow-up tasks should be derived from discovered gaps, for example:

```text
standardize backend error contracts
standardize OpenAPI validation/generation
standardize frontend feature layout
standardize migration CLI + Makefile lifecycle
standardize observability/logging
standardize CI quality gates
standardize testing fixtures/factories
standardize API pagination/filtering contracts
standardize release/versioning conventions
```

Do not implement these follow-ups automatically as part of this task unless they are explicitly required to make the convention system valid.
