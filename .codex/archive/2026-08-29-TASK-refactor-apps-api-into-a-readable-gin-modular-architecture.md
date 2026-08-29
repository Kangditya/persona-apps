# Task: Refactor `apps/api` Into a Readable Gin Modular Architecture
## Executed
## `Pending`

## Problem Statement

The current Go API under:

```text
apps/api
```

is difficult to navigate and does not yet provide a sufficiently clear Gin-oriented modular structure for continuous development.

The primary problem is **code locality and readability**. A developer working on one business capability should not need to search through broad global folders, oversized router/bootstrap files, or unrelated packages to understand:

```text
route
→ handler
→ application/service logic
→ repository/persistence
→ domain types
→ request/response types
→ tests
```

This task must reorganize the API into a predictable modular structure while preserving current behavior and repository contracts unless a behavior change is separately identified, planned, and approved.

This is primarily an **architectural refactoring task**, not a qurban feature-development task.

---

## Objective

Refactor `apps/api` into a maintainable Gin-based modular backend that is easier to read, extend, test, and review.

The target architecture should favor **business-module locality** over global technical-layer folders.

Preferred direction:

```text
apps/api/
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── bootstrap.go
│   │   └── dependencies.go
│   │
│   ├── config/
│   │   └── ...
│   │
│   ├── modules/
│   │   └── <business-module>/
│   │       ├── domain/
│   │       ├── application/
│   │       ├── infrastructure/
│   │       │   └── persistence/
│   │       ├── transport/
│   │       │   └── http/
│   │       └── module.go
│   │
│   ├── platform/
│   │   ├── database/
│   │   ├── http/
│   │   ├── middleware/
│   │   ├── logging/
│   │   └── observability/
│   │
│   └── shared/
│       └── ...
│
├── migrations/
├── test/
├── go.mod
└── Makefile
```

This structure is a **target direction, not a blind folder-generation requirement**. During DISCOVER, inspect the real repository and retain simpler existing structures where they already provide good separation.

Do not create empty architectural layers merely to match this diagram.

---

## Primary Readability Goal

For a business capability such as:

```text
event
purchasing
payment
participant
livestock
allocation
distribution
reporting
```

a developer should be able to navigate primarily within:

```text
internal/modules/<module>/
```

and understand that capability without repeatedly jumping between repository-wide folders such as:

```text
handlers/
services/
repositories/
models/
dto/
routes/
```

Avoid a structure where all handlers for every feature live together, all services live together, and all repositories live together unless DISCOVER demonstrates a strong repository-specific reason to retain it.

Prefer:

```text
modules/
├── event/
│   ├── ...
├── purchasing/
│   ├── ...
└── livestock/
    ├── ...
```

instead of:

```text
handlers/
├── event.go
├── purchasing.go
└── livestock.go

services/
├── event.go
├── purchasing.go
└── livestock.go

repositories/
├── event.go
├── purchasing.go
└── livestock.go
```

The final structure must optimize for **finding related code quickly**.

---

## Required Workflow

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

### Mandatory planning gate

Do **not** begin the refactor immediately.

Before modifying source code:

1. inspect the entire current `apps/api` tree;
2. identify every Gin router, route group, middleware registration, handler/controller, service/use case, repository, model/entity, DTO, configuration package, database package, migration entrypoint, test package, and bootstrap file;
3. trace representative requests from route registration to persistence;
4. identify package import direction and circular-dependency risks;
5. identify duplicate helpers and cross-feature coupling;
6. identify oversized files and packages that contain unrelated responsibilities;
7. inspect the current OpenAPI contracts before moving transport types;
8. inspect `docs/ARCHITECTURE.md`, `docs/DECISIONS.md`, `.codex/AGENTS.md`, and `.codex/CURRENT_STATE.md`;
9. produce an **existing → target package mapping**;
10. produce a planned file-change table;
11. produce the migration/refactoring sequence;
12. identify behavior-change risks separately from structural changes;
13. stop after PLAN and request explicit approval before BUILD.

The PLAN must be sufficiently detailed that the user can review the final folder structure before any large file movement occurs.

---

## DISCOVER Deliverables

Before implementation, report at minimum:

### 1. Current API tree

Provide a concise tree of the current backend structure with important Go files.

### 2. Request-flow traces

Trace at least these kinds of flows where they currently exist:

```text
public read endpoint
protected operations read endpoint
protected mutation endpoint
authentication/session endpoint
database-backed endpoint
```

For each selected endpoint show:

```text
route
→ middleware
→ handler
→ service/use case
→ repository
→ database
```

### 3. Structural issues

Classify findings such as:

```text
oversized router files
mixed public and operations routes
global controller/handler bucket
global service bucket
global repository bucket
business logic inside handlers
SQL inside handlers
transport DTOs leaking into domain/application layers
database models used directly as API responses
cross-module imports
circular package pressure
shared helpers that actually contain domain behavior
inconsistent error handling
inconsistent dependency injection
package names that hide their responsibility
duplicate route registration
implicit global state
```

Do not assume these problems exist; verify them from the codebase.

### 4. Refactoring risk map

Classify each intended move:

```text
safe structural move
requires import rewrite
requires constructor change
requires test update
could alter runtime behavior
could alter API contract
could alter transaction behavior
```

---

## Architectural Direction

### 1. Keep `main.go` minimal

The executable entrypoint should primarily:

```text
load configuration
construct application
run application
handle shutdown
```

Avoid route registration, SQL setup details, business dependency construction, and large middleware configuration directly in `main.go`.

Preferred ownership:

```text
cmd/api/main.go
    ↓
internal/app
```

---

### 2. Make Gin the explicit HTTP framework boundary

The project must use Gin consistently for API routing if Gin is already the approved backend framework.

Do not introduce a parallel `net/http` `ServeMux` routing architecture.

If any `ServeMux`-based application routing remains from earlier scaffolding, identify it during DISCOVER and plan its removal or isolation.

Standard-library HTTP types may still be used where Gin itself depends on or interoperates with them; the restriction concerns competing application-router architecture, not ordinary Go HTTP primitives.

---

### 3. Separate router registration from handlers

Route registration must be easy to scan.

Avoid one giant router file containing all endpoints.

Prefer module-owned registration such as:

```text
internal/modules/event/transport/http/routes.go
internal/modules/purchasing/transport/http/routes.go
internal/modules/payment/transport/http/routes.go
```

Each module should expose a narrow registration/composition surface, for example conceptually:

```go
func (m *Module) RegisterRoutes(router gin.IRouter)
```

or another repository-consistent equivalent.

The exact API must be selected during PLAN.

---

### 4. Keep public and operations route boundaries obvious

The backend serves two distinct API surfaces:

```text
Storefront Web   → public API
Operations Web   → protected operations API
```

Their route registration must be visibly separated.

A reader should be able to identify where these groups are created, for example conceptually:

```text
/api/v1/public/...
/api/v1/operations/...
```

or the repository's already approved paths.

Do **not** change route paths merely to match this example.

Preserve existing OpenAPI contracts unless path changes are explicitly approved.

Authentication/authorization middleware must be applied at the appropriate route-group boundary rather than duplicated inconsistently across handlers where possible.

---

## Recommended Module Shape

The preferred full module shape is:

```text
internal/modules/<module>/
├── domain/
│   ├── entity.go
│   ├── errors.go
│   └── repository.go
│
├── application/
│   ├── service.go
│   ├── command/
│   └── query/
│
├── infrastructure/
│   └── persistence/
│       ├── repository.go
│       ├── queries.go
│       └── mapper.go
│
├── transport/
│   └── http/
│       ├── routes.go
│       ├── handler.go
│       ├── request.go
│       ├── response.go
│       └── mapper.go
│
└── module.go
```

However, **do not over-split small modules**.

For a simple CRUD capability, this may be enough:

```text
internal/modules/<module>/
├── domain.go
├── service.go
├── repository.go
├── handler.go
├── routes.go
└── module.go
```

The PLAN must choose the smallest structure that remains readable.

The goal is not maximum abstraction. The goal is predictable organization and sustainable growth.

---

## Dependency Direction

The intended dependency direction is:

```text
transport/http
      ↓
application
      ↓
domain
      ↑
infrastructure/persistence
```

Required constraints:

* handlers may depend on application services/use cases;
* handlers must not contain SQL;
* application logic must not depend on Gin;
* domain code must not depend on Gin;
* domain code should not depend on database-driver details;
* persistence implements repository contracts needed by the application/domain;
* persistence must not import HTTP handlers;
* transport-specific request/response types should stay at the HTTP boundary;
* avoid direct module-to-module imports unless the dependency is a deliberate domain relationship documented during PLAN.

If the existing application is intentionally simpler than clean/hexagonal layering, preserve useful simplicity while enforcing the same dependency principles.

---

## Module Composition

Each business module should have an explicit composition point.

Preferred concept:

```text
module.go
```

Responsibilities may include:

```text
construct repository
construct service/use cases
construct handler
expose route registration
```

Global application bootstrap should compose modules rather than instantiate every internal implementation detail inline.

Conceptual target:

```text
internal/app
    ↓
modules.NewEvent(...)
modules.NewPurchasing(...)
modules.NewLivestock(...)
```

Do not introduce a dependency-injection framework merely for this task.

Use explicit Go constructors unless an existing approved mechanism already exists.

---

## Platform vs Shared Code

### `internal/platform`

Use for technical infrastructure that is not owned by one business module, such as:

```text
database
Gin server setup
middleware
logging
metrics/tracing
cache
messaging
shutdown/lifecycle helpers
```

### `internal/shared`

Use sparingly for small domain-neutral helpers such as:

```text
pagination primitives
validation helpers
clock abstraction
identifier helpers
```

Do not create a generic dumping ground such as:

```text
utils/
common/
helpers/
```

without a clear responsibility.

If a helper contains business meaning, move it into the owning module.

---

## Handler Rules

Gin handlers should primarily perform HTTP-boundary work:

```text
bind request
validate transport input
extract authenticated context
invoke application service/use case
map result/error
write response
```

Handlers must not become the location for:

```text
large business workflows
SQL queries
multi-repository transaction orchestration
financial calculation rules
inventory allocation rules
authorization policy definitions
```

If existing handlers contain these responsibilities, move them incrementally during the refactor without changing externally observable behavior.

---

## Request and Response Types

Avoid one repository-wide DTO package.

Prefer module-local HTTP types:

```text
internal/modules/<module>/transport/http/request.go
internal/modules/<module>/transport/http/response.go
```

or equivalent co-located files for smaller modules.

Requirements:

* Gin binding tags remain at the HTTP boundary;
* JSON representation concerns remain at the HTTP boundary where practical;
* persistence models must not automatically become public API response models;
* public and operations response shapes must remain contract-compatible;
* do not rename JSON fields during structural refactoring unless separately approved.

---

## Persistence Rules

Database access should be owned by the relevant module unless genuinely platform-wide.

Prefer:

```text
internal/modules/<module>/infrastructure/persistence/
```

for module-specific SQL/repositories.

The persistence layer may contain:

```text
repository implementation
SQL/query constants
row scanning
persistence model mapping
transaction-aware repository variants
```

Do not move database-specific details into `domain` merely to reduce file count.

Do not rewrite working SQL purely for architectural aesthetics.

---

## Transaction Rules

Preserve existing transaction boundaries exactly unless a defect is identified and separately approved.

During DISCOVER:

* identify where transactions start and commit/rollback;
* identify multi-repository operations;
* identify whether transaction ownership currently lives in handler, service, or repository code;
* propose a consistent transaction boundary during PLAN.

If a shared transaction manager is required, it may live under:

```text
internal/platform/database
```

but do not invent an abstraction if the current requirements do not need it.

---

## Error Handling

DISCOVER the existing error flow before changing it.

The refactor should converge toward a predictable chain:

```text
repository error
    ↓
application/domain error
    ↓
HTTP error mapping
    ↓
contract-compatible response
```

Requirements:

* do not scatter HTTP status-code selection throughout repositories;
* domain/application code should not return Gin-specific errors;
* preserve existing API status codes and response bodies unless a change is explicitly planned;
* distinguish authentication, authorization, validation, not-found, conflict, and internal errors where the existing contract supports them;
* retain enough error context for logging without leaking sensitive data to clients.

---

## Middleware Organization

Inspect all existing Gin middleware and classify ownership.

Cross-cutting middleware may live under:

```text
internal/platform/middleware
```

Examples:

```text
request ID
structured logging
recovery
CORS
authentication
session loading
rate limiting
```

Business-specific authorization should remain close to the relevant authorization/application policy rather than becoming an unstructured global middleware collection.

Middleware order must be documented because route behavior can depend on ordering.

---

## Configuration and Bootstrap

Consolidate application initialization into a readable sequence.

A reader should be able to understand startup in this approximate order:

```text
load config
    ↓
initialize logger
    ↓
initialize database
    ↓
initialize external dependencies
    ↓
construct modules
    ↓
construct Gin engine
    ↓
register middleware
    ↓
register routes
    ↓
start server
    ↓
graceful shutdown
```

Avoid hidden package-level initialization and mutable global dependencies where practical.

Do not introduce service locators or dependency injection containers for this refactor.

---

## Naming and Package Rules

Use names that indicate responsibility.

Prefer:

```text
handler
service
repository
routes
persistence
middleware
config
```

Avoid ambiguous names such as:

```text
manager
processor
helper
common
util
misc
base
```

unless their exact responsibility is documented and justified.

Go package names must remain short, lowercase, and idiomatic.

Avoid names such as:

```text
product_service
productService
product-service
```

for Go packages.

---

## File Size and Readability Rules

This task must explicitly reduce files that are difficult to scan because they contain several unrelated responsibilities.

During PLAN identify candidate files for splitting when they contain combinations such as:

```text
router registration + handler implementation
handler + SQL
multiple unrelated modules
configuration + server lifecycle
request DTOs + persistence model + response mapping
```

Do not split files solely to satisfy arbitrary line-count limits.

A file is acceptable when it has one coherent responsibility even if it is relatively long.

---

## Formatting Convention

Preserve the repository's approved formatting conventions.

For Go source code:

```text
gofmt is authoritative
```

Do not manually force whitespace that conflicts with `gofmt`.

For frontend TypeScript/TSX files affected incidentally by contract/import moves, preserve the repository's approved **4-space indentation** convention and do not perform unrelated frontend formatting refactors.

---

## Refactoring Strategy

Do not attempt a risky repository-wide rewrite in one uncontrolled move.

The PLAN should define incremental migration slices.

Preferred sequence:

### Phase 1 — Foundation

```text
app bootstrap
config
Gin server/router foundation
cross-cutting middleware
platform database boundary
```

### Phase 2 — One reference module

Select one representative business module and move it completely:

```text
routes
handler
application/service
repository
DTO mapping
tests
```

Use this module to validate the architectural pattern before duplicating it.

### Phase 3 — Remaining modules

Migrate modules one by one, keeping the application compiling and tests runnable after each logical slice.

### Phase 4 — Remove obsolete structure

Only after all consumers are migrated:

```text
remove old global routers
remove old handler buckets
remove unused services/repositories
remove forwarding wrappers that no longer provide value
```

### Phase 5 — Documentation and review

Update architecture documentation and final tree after the implementation is verified.

The exact phase boundaries must be adapted to the current repository discovered by the agent.

---

## Git Safety

Before source changes:

1. inspect the current branch;
2. do not implement directly on `development` or `origin/development`;
3. when Git metadata is writable, create or switch to a dedicated task branch before BUILD;
4. if Git metadata is read-only in the execution environment, report that constraint and do not attempt branch/commit/stash operations;
5. do not commit or push unless explicitly requested.

Suggested branch naming direction:

```text
refactor/api-gin-modular-architecture
```

The exact branch name may follow repository convention.

---

## Behavior-Preservation Constraints

Unless specifically approved during PLAN, this refactor must **not** change:

```text
HTTP methods
route paths
request JSON shapes
response JSON shapes
status codes
authentication requirements
authorization rules
pagination semantics
filter semantics
sorting semantics
transaction boundaries
database schema
migration history
business invariants
```

If the current implementation contains a defect discovered during refactoring:

1. document it;
2. identify whether fixing it is required to complete the refactor;
3. classify the fix as a plan variance;
4. do not silently change behavior.

---

## OpenAPI Constraints

The existing contracts remain authoritative:

```text
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
```

Structural Go refactoring alone should require **no contract change**.

If implementation and contract are already inconsistent, report the mismatch during DISCOVER.

Do not automatically rewrite the contract to match accidental implementation behavior.

Any contract correction requires an explicit plan item.

---

## Testing Requirements

Preserve existing tests and relocate them with their owning packages where appropriate.

Add focused tests where the refactor exposes untested structural boundaries.

### Router tests

Verify where applicable:

```text
route registration
HTTP method
path
public vs protected grouping
middleware boundary
404/405 behavior
```

### Handler tests

Verify:

```text
request binding
validation
service invocation
response mapping
known error mapping
unexpected error mapping
```

### Application/service tests

Verify business orchestration independently of Gin where practical.

### Repository tests

Preserve or add tests around significant queries and mapping behavior where repository test infrastructure already exists.

### Architecture tests/checks

If feasible without introducing a heavy dependency, add lightweight checks or documented review rules that guard against regressions such as:

```text
domain importing Gin
application importing Gin
handler importing raw database package for normal business queries
modules importing another module's transport package
```

Do not add an architectural testing framework solely for this task unless approved in PLAN.

---

## Minimum Acceptance Criteria

The task is complete only when all applicable criteria are satisfied:

1. `apps/api` uses Gin consistently as the application HTTP router.
2. There is one obvious application/bootstrap entry path.
3. `main.go` is small and contains no business routing logic.
4. Business capabilities are organized primarily by module/feature rather than repository-wide technical buckets.
5. Each migrated module has an obvious route-registration entrypoint.
6. Route registration is split into readable module-owned files rather than one monolithic router.
7. Public and operations API route boundaries are immediately visible.
8. Handlers contain HTTP concerns, not SQL or large business workflows.
9. Application/business logic is testable without Gin where practical.
10. Module-specific persistence code is easy to locate from its owning module.
11. Transport request/response types do not unnecessarily leak across the application.
12. Dependencies flow inward predictably and do not create circular imports.
13. No new generic `utils`/`common` dumping package is introduced.
14. Existing API behavior and OpenAPI contracts remain compatible unless explicitly approved otherwise.
15. Existing migrations continue to work unchanged unless a separately approved migration fix is required.
16. Existing tests continue to pass.
17. New/refactored code passes `gofmt` and the repository's Go static-analysis checks.
18. The final repository tree is materially easier to navigate than the starting structure.
19. Documentation explains where a developer should add a new route, handler, service/use case, repository query, and module.
20. The final review includes remaining architectural debt rather than hiding unfinished migration work.

---

## Developer Navigation Acceptance Test

As part of REVIEW, answer these questions using only the final repository structure:

```text
Where do I add a new operations route for livestock?
Where is the handler for that route?
Where is the business workflow?
Where is the database query?
Where are its request/response types?
Where are its tests?
Where is authentication middleware applied?
Where are public storefront routes registered?
Where are operations routes registered?
Where is the application composed and started?
```

If any answer requires searching several unrelated global folders, the refactor has not fully achieved its readability objective.

---

## Verification

Run the repository's actual Go validation commands discovered from `Makefile`, CI, and project documentation.

At minimum, where applicable:

```bash
gofmt -w <changed-go-files>
go test ./...
go vet ./...
make validate
```

Also run any existing commands for:

```text
lint
staticcheck
race tests
integration tests
migration verification
OpenAPI validation
build
```

If `apps/api` is a nested Go module, run Go commands from the correct module directory.

Do not claim a command passed if it was not executed successfully.

### Manual verification

Confirm:

1. the server starts normally;
2. representative public endpoints still respond with the same contract;
3. representative protected operations endpoints still enforce authentication/authorization;
4. representative reads return equivalent data;
5. representative mutations preserve transaction behavior;
6. graceful shutdown still works;
7. middleware executes in the intended order;
8. no route is accidentally registered twice or omitted;
9. no old router path remains active unintentionally;
10. adding a small endpoint to one module follows an obvious path without editing unrelated modules.

---

## Documentation Updates

Update the relevant architecture/development documentation after implementation.

At minimum document:

### Final backend tree

Show the important `apps/api` structure.

### Adding a new module

Document the sequence conceptually:

```text
create module
→ define domain/application boundary
→ implement persistence
→ implement HTTP handler
→ register module routes
→ wire module in application bootstrap
→ add tests
```

### Adding a new endpoint to an existing module

Document exactly which files typically change.

### Dependency rules

Document allowed and forbidden package dependencies.

### Public vs operations routes

Document where each surface is registered and protected.

---

## Final Report

The completed task must end with a structured report containing:

```text
initial architecture summary
problems confirmed during DISCOVER
approved target architecture
modules migrated
files moved
files created
files removed
route-registration changes
dependency-direction changes
tests added/updated
verification commands and results
OpenAPI impact
behavior changes, if any
plan variances
remaining architectural debt
recommended next task
```

Also include a concise **before vs after** tree so readability improvement can be reviewed directly.

---

## Constraints

* Keep the task focused on `apps/api` architecture and the minimum documentation/tests required to support it.
* Do not implement unrelated qurban product features.
* Do not add speculative microservices.
* Do not split the current API into multiple deployables.
* Do not introduce a dependency injection framework without explicit approval.
* Do not introduce an ORM or replace the current database library merely as part of modularization.
* Do not rewrite working SQL unnecessarily.
* Do not change OpenAPI contracts just to fit a preferred code structure.
* Do not perform broad frontend refactors.
* Do not create empty folders/layers with no implementation value.
* Prefer explicit constructors and readable Go over abstraction-heavy patterns.
* Record plan variances before modifying files outside the approved plan.
* Stop after PLAN for explicit approval before BUILD.
* Do not commit or push changes unless explicitly requested.
