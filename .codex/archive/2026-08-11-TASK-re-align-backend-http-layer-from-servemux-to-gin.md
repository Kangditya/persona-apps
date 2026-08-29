# Task: Re-align Backend HTTP Layer from ServeMux to Gin
## Executed
## `Planned`

## Objective

Re-align `apps/api` so **Gin becomes the canonical HTTP framework and router** for the backend.

The current backend trajectory introduced or retained `net/http` `ServeMux` routing. This task must remove that routing direction and migrate the HTTP adapter/bootstrap layer to Gin while preserving the existing modular-monolith architecture, domain boundaries, API contracts, behavior, and server lifecycle semantics.

This task also executes the repository formatting convention that was previously documented but not applied to the existing codebase: **all tracked source `*.go` and `*.tsx` files must be normalized to 4-space indentation**. This is an execution requirement, not a documentation-only convention update. The formatting sweep must be mechanical and must not introduce behavioral changes.

Target direction:

```text
HTTP Server
    ↓
Gin Engine
    ↓
Middleware
    ↓
Public / Operations Route Groups
    ↓
HTTP Handlers / Adapters
    ↓
Application Use Cases
    ↓
Domain
    ↓
Repositories / Infrastructure
```

Gin is an HTTP-edge dependency only. It must not leak into application, domain, repository, or persistence layers.

---

## Required Workflow

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

Before modifying code:

* inspect all HTTP bootstrap, router, handler, middleware, response, error, health, readiness, and server-lifecycle code under `apps/api`;
* locate every use of `http.NewServeMux`, `http.ServeMux`, `Handle`, `HandleFunc`, custom mux wrappers, route registration helpers, and ServeMux-oriented tests;
* inspect `go.mod`, `go.sum`, Makefiles, workspace files, Docker/development entrypoints, and test setup;
* inspect `contracts/openapi/storefront.yaml` and `contracts/openapi/operations.yaml` so route behavior is preserved;
* inspect architecture and decision artifacts for assumptions that currently imply or recommend ServeMux / standard-library routing;
* identify middleware already implemented or planned for request IDs, logging, recovery, CORS, authentication, authorization, error normalization, and observability;
* inventory every tracked `*.go` and `*.tsx` source file that is subject to the repository indentation convention;
* inspect `.editorconfig`, Prettier/Biome/ESLint configuration, Go editor/formatter settings, package scripts, Make targets, and CI checks that can enforce or overwrite indentation;
* identify generated/vendor/build-output paths that must not be mechanically reformatted;
* record current counts of in-scope `*.go` and `*.tsx` files so the formatting sweep can be verified as complete;
* produce a planned file-change table before implementation;
* explicitly distinguish routing-layer changes from business/domain changes;
* do not begin BUILD until the plan is complete and explicitly approved.

---

## Source-of-Truth Documents

Inspect and re-align as required:

```text
docs/PRD.md
docs/PRODUCT_MAP.md
docs/ARCHITECTURE.md
docs/DECISIONS.md
.codex/CURRENT_STATE.md
.codex/AGENTS.md
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
apps/api/go.mod
```

If any artifact currently describes ServeMux, `net/http` routing, or a still-open HTTP-framework decision, update it so the repository consistently states:

```text
Gin is the canonical backend HTTP framework and router.
```

Do not rewrite unrelated architecture documentation.

---

## Architectural Decision

### Canonical HTTP framework

Use:

```go
github.com/gin-gonic/gin
```

Gin owns:

* route registration;
* HTTP method/path matching;
* path/query parameter extraction;
* route grouping;
* HTTP middleware composition;
* request binding at the HTTP boundary;
* HTTP response rendering at the HTTP boundary;
* HTTP adapter context.

### `net/http` remains valid for

Do **not** interpret this migration as removing the Go standard HTTP package entirely.

`net/http` may still be used for:

* `http.Server`;
* status constants such as `http.StatusOK`;
* standard headers and cookies where appropriate;
* `context.Context` propagation through `c.Request.Context()`;
* graceful shutdown;
* transport-level tests or interfaces where standard-library compatibility is useful.

The prohibited routing direction after this task is:

```go
http.NewServeMux()
*http.ServeMux
mux.Handle(...)
mux.HandleFunc(...)
```

No application route should remain registered through ServeMux.

---

## Dependency Rule

Gin must stop at the HTTP adapter boundary.

Allowed:

```text
cmd/api
platform/http
module/adapter/http
        ↓
      Gin
```

Not allowed:

```text
application → gin.Context
domain      → gin.Context
repository  → gin.Context
persistence → gin.Context
```

Application methods must continue to accept framework-neutral values, preferably:

```go
context.Context
commands / queries / DTO-neutral application inputs
```

Example boundary:

```go
func (h *Handler) Create(c *gin.Context) {
    ctx := c.Request.Context()

    // bind HTTP request
    // map to application command
    // call use case with ctx
    // map result/error to HTTP response
}
```

Do not pass `*gin.Context` into application services.

---

## Scope

### In scope

1. Replace all ServeMux-based application routing with Gin.
2. Add Gin to `apps/api` dependencies if not already present.
3. Introduce or normalize a canonical Gin router/bootstrap composition.
4. Preserve existing public and operations API boundaries.
5. Migrate health/readiness endpoints to Gin.
6. Migrate existing HTTP handlers to Gin-native handler signatures where appropriate.
7. Migrate middleware to Gin middleware where it is truly HTTP-layer middleware.
8. Preserve `context.Context` propagation from the incoming request.
9. Preserve graceful shutdown and existing `http.Server` lifecycle behavior.
10. Preserve existing request/response contracts unless a defect is discovered and separately documented.
11. Re-align HTTP/router tests to exercise Gin.
12. Update architecture/decision/current-state artifacts so Gin is no longer an unresolved or contradictory choice.
13. Update developer guidance for registering future endpoints through Gin route groups.
14. Remove obsolete ServeMux-specific helpers after all callers have migrated.
15. Execute repository-wide 4-space indentation normalization for all in-scope tracked `*.go` and `*.tsx` source files.
16. Align existing formatter/editor/lint configuration so new Go/TSX edits do not silently return to the previous 2-space convention.
17. Verify the indentation sweep is mechanical and does not alter application behavior.

### Out of scope

* qurban business-rule changes;
* database schema or migration changes;
* repository or persistence redesign;
* authentication-provider implementation unless required only to preserve an already-existing middleware boundary;
* new API endpoints unrelated to proving the Gin migration;
* OpenAPI contract redesign;
* frontend behavioral, feature, routing, or UI changes unrelated to the required `*.tsx` indentation normalization;
* microservice extraction;
* unrelated refactors;
* replacing `net/http` server lifecycle machinery simply because Gin is introduced.

---

## Discovery Requirements

Inventory the current HTTP layer before changing it.

At minimum search for:

```text
http.NewServeMux
http.ServeMux
ServeMux
HandleFunc(
.Handle(
.Handler
http.Handler
http.HandlerFunc
ListenAndServe
http.Server
/health
/ready
```

Also locate:

```text
router.go
routes.go
server.go
middleware.go
response.go
error.go
health*.go
main.go
```

For every discovered ServeMux-related file, classify it as one of:

```text
REPLACE     → migrate directly to Gin
ADAPT       → retain purpose but change interface/composition
KEEP        → valid standard-library HTTP usage not related to routing
REMOVE      → obsolete after Gin migration
DEFER       → unrelated or requires separate approved work
```

Include this classification in PLAN.

---

## Router Design

Create one clear Gin composition root, but **do not centralize all route declarations in one router file**.

The composition root owns engine construction and top-level middleware only. Route registration must be separated into focused files by API surface and, when the route count justifies it, by backend module/domain.

Preferred conceptual structure, adjusted to actual repository conventions:

```text
apps/api/internal/platform/http/
├── router.go                 # gin.Engine construction + global middleware only
├── routes_health.go          # /health and /ready
├── routes_public.go          # public API group composition only
├── routes_operations.go      # operations API group composition only
├── middleware.go
├── response.go
└── error.go

apps/api/internal/<module>/adapter/http/
├── handler.go
└── routes.go                 # module-owned route registration
```

Exact directories and names must follow the repository structure discovered during PLAN. Do not create a parallel architecture merely to match this example.

### Composition root

`router.go` should remain small and predictable:

```go
func NewRouter(deps Dependencies) *gin.Engine {
    router := gin.New()

    router.Use(
        gin.Recovery(),
        RequestIDMiddleware(),
        LoggingMiddleware(),
    )

    registerHealthRoutes(router, deps)
    registerPublicRoutes(router, deps)
    registerOperationsRoutes(router, deps)

    return router
}
```

It must not become a registry containing every endpoint in the application.

### Public route composition

The public route file should create the canonical public group and delegate registration to owning modules:

```go
func registerPublicRoutes(router *gin.Engine, deps Dependencies) {
    public := router.Group(publicAPIPrefix)

    eventhttp.RegisterPublicRoutes(public, deps.EventHandler)
    purchasinghttp.RegisterPublicRoutes(public, deps.PurchasingHandler)
}
```

### Operations route composition

The operations route file should create the protected operations group, attach operations-level middleware, and delegate to owning modules:

```go
func registerOperationsRoutes(router *gin.Engine, deps Dependencies) {
    operations := router.Group(operationsAPIPrefix)
    operations.Use(OperationsAuthMiddleware(deps.Auth))

    eventhttp.RegisterOperationsRoutes(operations, deps.EventHandler)
    purchasinghttp.RegisterOperationsRoutes(operations, deps.PurchasingHandler)
}
```

Do not copy these module names if they do not exist yet. Register only discovered/implemented modules.

### Module-owned routes

Where a module owns multiple handlers, keep its concrete endpoint declarations beside that module's HTTP adapter rather than adding them to the platform router:

```go
func RegisterOperationsRoutes(group *gin.RouterGroup, h *Handler) {
    events := group.Group("/events")

    events.GET("", h.List)
    events.GET("/:id", h.Get)
    events.POST("", h.Create)
}
```

Route files must contain routing/composition concerns only. They must not accumulate business logic, SQL, response mapping, or application orchestration.

### Router-file maintenance rules

* keep `router.go` limited to engine construction, global middleware, and top-level route registration calls;
* keep health/readiness route registration separate from domain routes;
* keep public and operations group registration in separate files;
* prefer module-owned `routes.go` files when a module has its own HTTP adapter package;
* do not create one file per individual endpoint; split by coherent API surface/module instead;
* avoid circular imports between platform HTTP composition and module HTTP adapters;
* centralize canonical API-prefix constants rather than scattering literal prefixes;
* do not duplicate the same route registration across public and operations surfaces;
* tests should target both the composed `gin.Engine` and focused module route registration where useful.

Do not use `gin.Default()` automatically if the repository already owns logging/recovery behavior that would become duplicated. Decide during PLAN whether to use:

```go
gin.New()
```

or:

```go
gin.Default()
```

and document the reason.

Default preference for this project:

```text
gin.New() + explicit middleware
```

because middleware ownership should remain visible and deliberate.

---

## Route Boundaries

Keep public and operations APIs visibly separated.

Expected conceptual structure:

```go
public := router.Group("/api/public/v1")
operations := router.Group("/api/operations/v1")
```

Use the repository's actual canonical prefixes discovered from OpenAPI and existing code. Do not invent or rename route prefixes merely to match this example.

### Public group

Public routes must not accidentally receive operations-only middleware or expose internal fields.

### Operations group

Operations routes must provide a clear insertion point for:

```text
authentication
authorization
audit context
operations-specific observability
```

when those capabilities exist.

Client-side authorization is never a substitute for backend authorization.

---

## Handler Migration

Where handlers currently use:

```go
func(w http.ResponseWriter, r *http.Request)
```

migrate the HTTP adapter to:

```go
func(c *gin.Context)
```

without moving domain/application logic into the handler.

Preferred responsibility:

```text
Gin handler
  ├─ read path/query/header/body
  ├─ bind and perform transport-level validation
  ├─ obtain context.Context
  ├─ map request → application command/query
  ├─ invoke application layer
  ├─ map known application errors → HTTP errors
  └─ serialize response
```

Not allowed:

```text
Gin handler
  ├─ SQL
  ├─ transaction orchestration belonging to application layer
  ├─ authoritative domain calculations
  └─ persistence-specific decisions
```

---

## Request Binding and Validation

Use Gin binding only at the transport boundary.

Example:

```go
var req CreateRequest
if err := c.ShouldBindJSON(&req); err != nil {
    // normalize as HTTP validation/bad-request response
    return
}
```

Do not assume Gin validation replaces domain validation.

Separate:

```text
Transport validation
    ↓
required JSON shape / parseability / basic field format

Domain validation
    ↓
business invariants and authoritative rules
```

Preserve the existing validation approach if the repository already defines canonical validators or error types.

---

## Response and Error Strategy

Inspect existing response envelopes before changing them.

If the repository already has a canonical shape such as:

```json
{
  "data": {},
  "error": null
}
```

or another envelope, preserve it.

Do not introduce a new envelope only because Gin provides `c.JSON`.

Centralize reusable HTTP response/error translation where justified, for example:

```go
func WriteError(c *gin.Context, err error)
```

or repository-equivalent naming.

Errors should remain distinguishable where currently supported:

```text
400 bad request
401 unauthorized
403 forbidden
404 not found
409 conflict
422 validation/domain rejection if already canonical
500 internal server error
503 unavailable
```

Do not expose stack traces, SQL errors, secrets, access tokens, cookies, or internal infrastructure details.

---

## Middleware Migration

Inventory each current middleware and classify whether it should become Gin-native.

Potential middleware:

```text
recovery
request ID
structured request logging
CORS
authentication
authorization
rate limiting
tracing
metrics
audit context
```

Canonical Gin form:

```go
func Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // before
        c.Next()
        // after
    }
}
```

Requirements:

* preserve execution ordering;
* preserve abort semantics;
* use `c.Abort()` / `c.AbortWithStatusJSON(...)` only where appropriate;
* continue propagating request-scoped values through `context.Context` when application code requires them;
* do not make `gin.Context` the cross-layer request context;
* do not duplicate Gin recovery/logger middleware with repository-owned equivalents.

---

## Context Propagation

This migration must explicitly preserve framework-neutral context propagation.

HTTP adapter:

```go
ctx := c.Request.Context()
result, err := useCase.Execute(ctx, input)
```

If middleware adds request-scoped values needed below the HTTP layer, prefer attaching them to the request context using a typed key or an existing repository abstraction rather than forcing downstream packages to depend on Gin.

Example direction:

```go
ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
c.Request = c.Request.WithContext(ctx)
```

Only introduce new context values when actually required by existing behavior.

---

## Health and Readiness

Migrate existing health endpoints from ServeMux to Gin without semantic changes.

Example only:

```go
router.GET("/health", healthHandler)
router.GET("/ready", readinessHandler)
```

Preserve:

* current status codes;
* readiness dependency checks;
* response shapes;
* database availability semantics;
* test behavior.

Do not turn `/health` and `/ready` into aliases if they currently represent different guarantees.


---

## Repository-Wide Indentation Refactor

The previous convention work documented the desired indentation but did not normalize the existing source tree. This task must **execute the convention across the codebase**.

### Required source coverage

Normalize every tracked, human-maintained source file matching:

```text
**/*.go
**/*.tsx
```

Target indentation:

```text
4 literal spaces per indentation level
no 2-space indentation retained as the repository convention
```

This applies repository-wide, not only to files touched by the Gin migration.

Exclude only paths that are clearly not owned source code, such as discovered equivalents of:

```text
.git/
node_modules/
vendor/          # when vendored/generated and repository policy says not to edit it
dist/
build/
coverage/
.generated/
gen/
```

Do not invent exclusions. Inspect the repository and list every exclusion in PLAN with a reason.

### Mechanical-change requirement

The indentation sweep must be a formatting-only transformation.

Do not intentionally change during the sweep:

```text
identifiers
control flow
expressions
function signatures
API contracts
business rules
SQL
JSX structure
component behavior
imports unless required by the repository formatter
string contents
comments except indentation
```

Do not combine opportunistic cleanup with the indentation pass.

If an existing line is already correctly indented with 4 spaces, leave its content unchanged.

### Go-specific requirement

The project convention requested by this task is **4 literal spaces** for `*.go` indentation.

Go's canonical `gofmt` normally uses tabs for leading indentation. Therefore, before executing the sweep, DISCOVER and PLAN must identify any tooling that automatically runs `gofmt`, `goimports`, editor format-on-save, or another formatter that would immediately rewrite the requested literal-space convention.

The task must not silently claim both of these incompatible states are enforced:

```text
A. exact 4 literal spaces in Go source
B. canonical gofmt leading-tab output
```

For this task, **A is the explicit repository requirement**. Where project tooling currently forces B, update the repository-owned formatting/editor/CI convention where feasible, document the compatibility limitation, and do not run a repository-wide formatter that undoes the requested indentation after normalization.

`go test`, `go vet`, and `go build` remain required; Go source must remain syntactically valid and compile successfully.

### TSX-specific requirement

For `*.tsx`, align the existing formatter configuration rather than creating competing formatters.

Where the repository uses Prettier or compatible settings, the effective convention should resolve to the equivalent of:

```text
useTabs: false
tabWidth: 4
```

Where `.editorconfig` owns the convention, use the equivalent of:

```ini
[*.tsx]
indent_style = space
indent_size = 4
```

If another formatter such as Biome owns TSX formatting, adjust the existing canonical configuration instead of adding Prettier solely for this task.

Do not change unrelated formatting rules such as quote style, semicolons, trailing commas, JSX wrapping, or line width unless the existing formatter necessarily applies them and the resulting variance is documented before execution.

### Configuration source of truth

Inspect first, then update only the canonical configuration already used by the repository. Potential files include:

```text
.editorconfig
.prettierrc*
prettier.config.*
biome.json*
eslint.config.*
.vscode/settings.json
Makefile
package.json
pnpm workspace scripts
CI workflow formatting checks
```

Do not add duplicate formatter configuration when an existing source of truth already exists.

### Execution strategy

To keep the Gin migration reviewable despite the large mechanical formatting diff:

```text
1. Complete and verify semantic Gin migration changes first.
2. Record the semantic diff before the formatting sweep.
3. Execute the repository-wide indentation normalization.
4. Re-run backend and frontend verification.
5. Review the final diff both normally and with whitespace ignored.
```

Use a whitespace-insensitive review such as the repository-equivalent of:

```bash
git diff --ignore-all-space
git diff --ignore-space-change
```

to confirm that the indentation sweep itself did not introduce hidden semantic edits.

Do not commit intermediate states; the task still ends with an uncommitted working tree unless the user separately requests a commit.

### Formatting verification

After normalization, verify all in-scope tracked files were considered. Use repository-safe commands based on discovered paths, for example:

```bash
git ls-files '*.go' '*.tsx'
```

Perform an automated indentation audit capable of identifying remaining leading indentation based on 2-space levels or tabs in files that are required to use literal 4 spaces. The audit must avoid false positives from intentionally aligned multiline literals, raw strings, generated code, or content where leading whitespace is data.

Report:

```text
Go files scanned
TSX files scanned
Go files changed
TSX files changed
Excluded files/paths
Remaining violations
Formatter/config files changed
Known formatter compatibility caveats
```

A zero remaining-violation result is required for code indentation that is safely machine-detectable. Any unavoidable exception must be listed file-by-file with justification.

---

## Server Lifecycle

Gin should be mounted as the `http.Server` handler rather than replacing established graceful-shutdown infrastructure.

Preferred conceptual composition:

```go
router := httpadapter.NewRouter(deps)

server := &http.Server{
    Addr:              cfg.HTTPAddr,
    Handler:           router,
    ReadHeaderTimeout: ...,
    ReadTimeout:       ...,
    WriteTimeout:      ...,
    IdleTimeout:       ...,
}
```

Preserve existing timeout values and shutdown behavior unless a current defect is identified during DISCOVER.

Do not regress signal handling or graceful shutdown.

---

## Dependency Management

Add Gin using the repository's Go dependency workflow.

Expected module:

```text
github.com/gin-gonic/gin
```

Requirements:

* use the version selected by the current Go module tooling;
* run `go mod tidy` in the correct module/workspace context;
* do not add unrelated dependencies;
* document newly introduced indirect dependencies in the final report only when relevant;
* do not manually edit generated dependency checksums.

---

## Documentation Re-alignment

The task must eliminate contradictory architectural guidance.

Update only applicable statements in:

```text
docs/ARCHITECTURE.md
docs/DECISIONS.md
.codex/CURRENT_STATE.md
.codex/AGENTS.md
```

Expected outcome:

```text
Before:
HTTP framework undecided / ServeMux trajectory / standard-library router examples

After:
Gin is the canonical HTTP framework and router.
net/http remains the server/runtime foundation where appropriate.
Gin is confined to HTTP adapters and bootstrap code.
```

If `docs/DECISIONS.md` uses ADR-style decisions, add or amend the relevant decision instead of silently deleting historical context.

Do not rewrite unrelated decisions.

---

## OpenAPI Contract Impact

This is an implementation-layer migration.

Default expectation:

```text
NO API CONTRACT CHANGE
```

Do not change:

* HTTP methods;
* canonical paths;
* request fields;
* response fields;
* status semantics;
* public vs operations exposure;

unless an actual pre-existing mismatch is discovered.

If a mismatch is discovered:

1. record it in PLAN;
2. do not silently fix it as part of router migration;
3. classify it as a separate contract issue unless fixing it is strictly necessary to preserve existing behavior.

---

## Testing Requirements

### Router tests

Verify at minimum:

* known routes resolve through Gin;
* HTTP methods are enforced correctly;
* unknown routes return the expected not-found behavior;
* unsupported methods preserve expected behavior;
* path parameters are available to handlers;
* query parameters are preserved;
* public and operations route groups do not cross-register endpoints.

### Handler tests

Use Gin test mode where appropriate:

```go
gin.SetMode(gin.TestMode)
```

Prefer `httptest` against the actual Gin engine for routing behavior.

Test existing behavior including:

```text
success
bad request
not found
internal error
health
readiness
```

plus authentication/authorization cases if those middleware already exist.

### Middleware tests

Verify where applicable:

* middleware order;
* abort behavior;
* request ID propagation;
* context propagation;
* recovery behavior;
* sensitive data is not logged;
* operations-only middleware is not applied to public routes accidentally.

### Regression requirement

Existing API behavior must remain functionally equivalent after the migration.

Do not rewrite tests merely to make failures disappear. Update tests where their only obsolete assumption is the ServeMux implementation detail.

---

## Minimum Functional Acceptance Criteria

1. `apps/api` uses Gin as the canonical application router.
2. No application route remains registered through `http.ServeMux`.
3. No `http.NewServeMux()` remains in production API routing code.
4. Existing API paths and methods remain unchanged unless separately documented.
5. Public and operations APIs remain explicitly separated.
6. Health and readiness endpoints run through Gin.
7. Existing graceful shutdown still uses a correctly configured server lifecycle.
8. Application and domain layers do not import Gin.
9. Repositories and persistence do not import Gin.
10. Application use cases continue to receive `context.Context`, not `*gin.Context`.
11. Request/response/error semantics are preserved.
12. Existing middleware behavior is preserved or explicitly documented when intentionally changed.
13. Gin middleware ordering is deterministic and tested where material.
14. Router and handler tests use the real Gin composition rather than a parallel fake router.
15. `router.go` remains a small composition root and does not contain all concrete application endpoints.
16. Health, public, and operations route registration are separated into focused router/route files.
17. Module-specific endpoint declarations are owned by module HTTP adapter route files where the repository structure supports that boundary.
18. Route prefixes and group ownership are centralized and deterministic; route declarations are not scattered through handlers/bootstrap code.
19. Architecture and decision artifacts identify Gin as the canonical backend HTTP framework.
20. No contradictory ServeMux recommendation remains in active project guidance.
21. `go mod tidy` completes successfully.
22. Existing repository validation, lint, tests, and build continue to pass.
23. No unrelated business/domain changes are introduced.
24. Every in-scope tracked `*.go` source file has been processed by the 4-space indentation migration.
25. Every in-scope tracked `*.tsx` source file has been processed by the 4-space indentation migration.
26. Existing canonical formatter/editor configuration is aligned to 4-space indentation where repository tooling supports it.
27. No source file remains on the previous 2-space indentation convention except explicitly documented exclusions.
28. The formatting sweep is verified as behavior-neutral using tests/builds plus whitespace-insensitive diff review.
29. Backend Go tests/builds remain green after the indentation sweep.
30. Frontend lint/typecheck/tests/build remain green after the TSX indentation sweep where those commands exist.
31. The final report includes file counts, changed counts, exclusions, remaining violations, and formatter compatibility caveats.
32. No commit or push is performed.

---

## Removal Criteria

After all routes have migrated, remove ServeMux-specific code only when no longer referenced.

Candidates may include:

```text
ServeMux constructor helpers
ServeMux route registration helpers
http.HandlerFunc adapters created only for ServeMux
custom middleware chaining built solely around ServeMux
ServeMux-specific tests
obsolete routing comments/docs
```

Do **not** remove generic helpers simply because they use valid `net/http` types.

Examples that may remain:

```text
http.Server
http.Status*
httptest
http.Header
http.Cookie
context.Context
```

The goal is **full Gin routing**, not artificial elimination of the standard HTTP library.

---

## Planned Migration Sequence

The PLAN should refine this sequence against actual files:

```text
1. Inventory current router and HTTP adapter usage
2. Record current behavior with focused tests if coverage is insufficient
3. Add Gin dependency
4. Create canonical Gin engine/bootstrap with a small `router.go` composition root
5. Create separated health, public, and operations route-composition files
6. Establish module-owned route files for discovered HTTP modules where appropriate
7. Migrate global middleware
8. Migrate health/readiness routes
9. Migrate public routes through their owning route files
10. Migrate operations routes through their owning route files
11. Migrate remaining handlers/adapters
12. Reconnect Gin engine to existing http.Server lifecycle
13. Remove obsolete ServeMux routing helpers and monolithic route declarations
14. Update router, route-registration, handler, and middleware tests
15. Re-align architecture/decision/current-state/agent docs
16. Run full verification
17. Review imports to ensure Gin did not leak below HTTP boundary
18. Review router-file size/ownership to ensure routing remains maintainable
19. Inventory all tracked `*.go` and `*.tsx` files plus canonical formatter/editor configuration
20. Record/inspect the semantic Gin diff before mass formatting
21. Normalize all in-scope `*.go` files to 4-space indentation
22. Normalize all in-scope `*.tsx` files to 4-space indentation
23. Update canonical formatting/editor/lint configuration to preserve 4-space indentation where supported
24. Run automated indentation audit and resolve remaining safe-to-fix violations
25. Re-run complete backend and frontend verification after the formatting sweep
26. Review `git diff --ignore-all-space` / equivalent to detect unintended semantic changes hidden by the mechanical diff
27. Report exact file counts, exclusions, violations, and tooling caveats
```

Prefer incremental compilation/testing after each migration group rather than one large unverified replacement.

---

## Verification

Run repository-standard checks discovered from the project.

At minimum, where applicable:

```bash
go test ./...
go vet ./...
go build ./...
go mod tidy
```

Also run project-level commands such as:

```bash
make validate
make test
make build
```

only when those targets exist.

If the monorepo has a workspace-specific command, use the repository's canonical command instead of inventing one.

Because this task intentionally reformats `*.tsx` across the repository, also run the existing frontend checks where available:

```bash
pnpm run lint
pnpm run typecheck
pnpm run test
pnpm run build
```

Do not add a new formatter/test framework merely to satisfy verification. Use existing workspace filters/scripts when the monorepo requires them.

### Indentation verification

Inventory the final source set:

```bash
git ls-files '*.go' '*.tsx'
```

Run a repository-safe automated indentation audit for the explicit 4-space convention and record any exclusions/false-positive classes. The final audit must demonstrate that the previous 2-space convention is no longer present in normal code indentation.

Also review the semantic diff with whitespace ignored:

```bash
git diff --ignore-all-space
git diff --ignore-space-change
```

Any non-whitespace change produced solely during the mass-formatting phase must be explained and either reverted or recorded as an approved variance.

### Static verification

After migration, search production backend code for remaining ServeMux usage:

```bash
rg 'NewServeMux|ServeMux|\.HandleFunc\(|\.Handle\(' apps/api
```

Review every match manually because `.Handle` may legitimately occur in unrelated code.

Also verify Gin does not leak into forbidden layers:

```bash
rg 'github.com/gin-gonic/gin|gin\.' apps/api/internal
```

Every match must be confined to approved HTTP adapter/platform locations.

### Manual verification

Start the API and verify:

1. server boots successfully;
2. `/health` preserves its prior response;
3. `/ready` preserves its prior response and dependency semantics;
4. at least one public endpoint routes correctly if currently implemented;
5. at least one operations endpoint routes correctly if currently implemented;
6. unknown route behavior is acceptable and documented;
7. method mismatch behavior is acceptable and documented;
8. graceful shutdown still completes correctly;
9. logs do not duplicate requests because both Gin and custom logging middleware are active;
10. no sensitive request data is newly emitted by middleware.

If a required route does not yet exist in the project, report it as not applicable rather than creating a speculative endpoint.

---

## Final Report

The completed task must report:

```text
Implemented behavior
Preserved behavior
Router/middleware architecture
Files changed
Files removed
Gin dependency/version
ServeMux usages removed
Remaining net/http usages and why they remain
Go files scanned / changed by indentation migration
TSX files scanned / changed by indentation migration
Formatting exclusions and remaining violations
Formatter/editor/CI configuration changes
Go gofmt/literal-space compatibility caveats
Tests added/updated
Verification commands and results
Documentation/ADR changes
OpenAPI impact
Plan variances
Deferred issues
Remaining risks
Recommended next task
```

Include an explicit boundary check:

```text
Gin imports outside HTTP adapter/bootstrap layers: NONE
```

or list every exception with justification.

---

## Constraints

* Gin is the target framework; do not retain ServeMux as an alternate production router.
* Do not run Gin and ServeMux side-by-side as permanent architecture.
* Temporary coexistence during BUILD is acceptable only as an incremental migration state and must be removed before completion.
* Preserve modular-monolith package boundaries.
* Preserve existing API contracts.
* Keep Gin out of domain/application/repository layers.
* Do not replace domain validation with Gin binding validation.
* Do not introduce speculative endpoints or business behavior.
* Do not refactor persistence merely to support the router change.
* Keep semantic changes focused on HTTP architecture and required documentation/tests; the repository-wide `*.go` / `*.tsx` indentation sweep is an explicitly approved mechanical exception and is expected to create a large whitespace-only diff.
* The 4-space normalization of tracked `*.go` and `*.tsx` source is already approved scope and must not be deferred as unrelated refactoring.
* Record scope variance before making non-formatting changes outside the approved plan.
* Do not commit or push changes.
