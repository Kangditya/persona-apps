# Task: Frontend API Layers for Operations and Storefront Web
## `Executed`

## Objective

Implement maintainable, typed API layers for both frontend applications:

```text
apps/operations-web
apps/storefront-web
```

The API layers must give each application a clear boundary between pages/features and remote data access while preserving the repository's existing React, TypeScript, Vite, React Router, and monorepo conventions.

The task is focused on frontend API integration. It must not implement speculative qurban business capabilities, invent production endpoints, or move authoritative business rules into either web application.

The API layer should support the first concrete vertical slice that the repository can verify against the existing Go API and/or an explicitly documented development provider. It must be straightforward to extend as public and operations contracts gain real endpoints.

---

## Required Workflow

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

Before modifying code:

* inspect both frontend applications and their existing route, layout, feature, and library conventions;
* inspect `apps/api`, `contracts/openapi/storefront.yaml`, and `contracts/openapi/operations.yaml`;
* identify which endpoints are actually available and which remain placeholders;
* inspect the shared `packages/api-client` and `packages/contracts` shells before creating duplicate code;
* identify configuration, authentication, error, testing, and environment-variable patterns;
* produce a planned file-change table;
* identify assumptions, endpoint gaps, and unresolved product decisions;
* do not begin implementation until the plan is complete and explicitly approved.

If a required backend endpoint or contract is unavailable, record the limitation and use only a clearly isolated development-safe adapter or mock boundary. Do not fabricate a backend response and present it as implemented product behavior.

---

## Source-of-Truth Constraints

Follow these repository documents and instructions:

```text
docs/PRD.md
docs/PRODUCT_MAP.md
docs/ARCHITECTURE.md
docs/DECISIONS.md
.codex/CURRENT_STATE.md
.codex/AGENTS.md
```

Relevant architectural rules:

* `apps/api` is the system of record for business invariants, authorization, transactions, persistence, and API contracts;
* Storefront Web uses the public API surface and must not access the database or internal operational fields;
* Operations Web uses the operations API surface and must not bypass backend authorization or mutate state through analytics endpoints;
* public and operations OpenAPI contracts remain separate;
* React Router owns navigation and route registration;
* TanStack Query is the approved owner of remote API/server state once a concrete API slice is introduced;
* frontend validation may improve usability but is not authoritative domain validation;
* do not add speculative multitenancy, microservices, or infrastructure;
* do not commit secrets, generated clients, build output, or `.env` files.

---

## Scope

### In scope

Implement the frontend API boundaries for both applications, including as applicable:

* runtime API base-URL configuration through the established Vite environment mechanism;
* a shared or application-owned HTTP transport, selected after inspecting existing package boundaries;
* typed request and response models derived from the approved contract or existing source types;
* request cancellation and timeout behavior where supported by the existing stack;
* normalized API errors suitable for loading, empty, unauthorized, forbidden, validation, unavailable, and unknown-error UI states;
* authentication/session-header integration at the transport boundary without embedding credentials in source or logs;
* public endpoint modules for Storefront Web;
* protected operations endpoint modules for Operations Web;
* TanStack Query installation and provider configuration only if the selected vertical slice requires remote server state;
* query keys, query functions, mutations, invalidation, and explicit stale/loading/error behavior for the implemented slice;
* integration with the smallest relevant existing routes or pages so the API layer is exercised by real application code;
* focused unit and integration tests using the repository's existing Vitest tooling;
* local development and environment-variable documentation.

### Out of scope

* implementing new qurban domain behavior in the Go API unless separately approved as part of the plan;
* inventing endpoints, fields, permissions, metrics, or workflow rules;
* final Figma visual implementation or broad page redesign;
* replacing React Router or introducing a second routing system;
* adding a global state-management library for remote data;
* direct database access from either frontend;
* exposing operations data through Storefront endpoints;
* generated API clients unless the repository already has an approved generation workflow;
* authentication-provider implementation when no backend contract or provider is available;
* broad refactoring unrelated to API access.

---

## Application Boundaries

### Storefront Web

Location:

```text
apps/storefront-web
```

The Storefront API layer may access only public API capabilities such as event discovery, offering catalogue, public purchase interaction, payment interaction, purchase tracking, and participant-facing documents when those endpoints are confirmed by contract.

It must not import Operations Web modules, depend on internal API fields, or assume purchaser and Sohibul Qurban are always the same person.

### Operations Web

Location:

```text
apps/operations-web
```

The Operations API layer may access protected operational capabilities such as event operations, purchasing administration, payment verification, participant operations, livestock, allocation, distribution, and reporting when those endpoints are confirmed by contract.

It must not rely on client-side authorization as a security boundary, calculate authoritative financial balances locally, or silently overwrite contested updates.

### Shared packages

Inspect before changing:

```text
packages/api-client
packages/contracts
```

Place code in a shared package only when it is genuinely transport- or contract-wide and does not leak public/internal boundary concerns. Otherwise keep API modules within their owning application.

---

## API-Layer Design Requirements

The implementation plan must explicitly document:

```text
API base URL and environment configuration
Transport ownership and dependency direction
Public versus operations client boundaries
Authentication/session handling
Request serialization and response parsing
Error normalization
Timeout and cancellation behavior
Query and mutation ownership
Cache and invalidation policy
Development/mock-provider strategy
Contract and versioning strategy
```

The implementation must:

* keep pages and UI components unaware of raw `fetch` or transport details;
* keep endpoint paths and query keys centralized rather than scattered as string literals;
* use typed endpoint inputs and outputs;
* preserve server response envelopes and error semantics unless an explicit adapter is required;
* make authorization failures distinguishable from validation and availability failures;
* avoid logging access tokens, cookies, personal data, payment data, or full request bodies;
* avoid caching sensitive operations data beyond the selected query policy;
* avoid duplicate requests and stale updates where the selected query library supports those safeguards;
* leave an explicit extension point for future authentication and API-contract growth;
* provide a safe, deterministic test transport without requiring production credentials.

If an endpoint is not available, mark it as `TBD` or `not implemented` in the plan and do not add a fake production implementation.

---

## Contract Requirements

The public and operations contracts must remain separate:

```text
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
```

When the implementation consumes existing endpoints, align types and tests to those contracts. When the task introduces or changes an endpoint contract, update the appropriate OpenAPI document and document:

* method and path;
* authentication requirement;
* request parameters or body;
* success response;
* error responses;
* pagination, filtering, or sorting behavior where applicable;
* timestamp and timezone representation;
* compatibility expectations.

Do not create an incompatible manually maintained duplicate when an approved shared contract or generation approach exists.

---

## Minimum Functional Acceptance Criteria

The completed implementation must demonstrate, for both applications where applicable:

1. API base URL is read from the established environment mechanism.
2. Pages/features call endpoint modules rather than raw transport functions.
3. Successful responses are typed and usable by application code.
4. Loading, empty, unauthorized/forbidden, validation, unavailable, and unexpected-error states are representable.
5. Requests can be cancelled or safely ignored when the owning view unmounts.
6. Authentication handling is centralized and does not expose secrets.
7. Public and operations clients cannot accidentally cross their API boundary.
8. Query keys and cache invalidation are deterministic for implemented reads and mutations.
9. A development/test provider can exercise the API layer without production credentials.
10. Existing placeholder routes continue to compile and render.
11. No invented business data is presented as authoritative backend data.
12. The API layer can accept additional endpoints without rewriting the application shell.

The exact endpoint slice must be selected during DISCOVER and PLAN from available contracts and approved requirements.

---

## Testing

Add focused tests for the selected implementation, including as applicable:

### Transport and client tests

* base-URL resolution;
* request method, path, headers, and serialization;
* response parsing;
* timeout or cancellation;
* normalized API errors;
* unauthorized and forbidden responses;
* absence of secrets and sensitive data in logs;
* public/operations client boundary.

### Query and feature tests

* initial loading;
* successful data rendering boundary;
* empty response;
* stale or refetch behavior;
* mutation success and invalidation;
* validation failure;
* unavailable backend;
* cleanup on unmount;
* deterministic development/test transport.

Use existing repository test tooling. Do not add a new testing framework.

---

## Documentation and Deliverables

The completed task must include:

1. API-layer implementation for `apps/operations-web`;
2. API-layer implementation for `apps/storefront-web`;
3. shared transport, contract, or type changes only when justified by the plan;
4. relevant route or feature integration;
5. focused tests;
6. environment-variable and local-development documentation;
7. OpenAPI updates only when the task changes a real contract;
8. a final report containing:
   * implemented behavior;
   * verified behavior;
   * assumed behavior;
   * deferred behavior;
   * files changed;
   * dependencies added;
   * contract impact;
   * plan variances;
   * remaining risks and recommended next task.

---

## Verification

Run the applicable repository checks, including:

```bash
make validate
pnpm run lint
pnpm run typecheck
pnpm run test
pnpm run build
```

Also run focused application tests and any contract validation introduced by the implementation. If a command cannot run, report the exact command and blocker; do not mark it as passed.

Manually verify:

1. both applications start with documented environment configuration;
2. each implemented API call reaches the intended API boundary;
3. success and failure states are visible at the consuming route or feature;
4. request cleanup does not leave stale updates or duplicate active work;
5. no frontend code directly accesses the database or embeds secrets;
6. adding another endpoint follows the documented extension pattern.

---

## Constraints

* Keep changes limited to frontend API layers and their necessary integration.
* Do not implement final visual designs.
* Do not invent unavailable backend behavior.
* Do not treat placeholder contracts as implemented capabilities.
* Do not duplicate public and operations domain boundaries.
* Do not add dependencies without documenting why they are necessary.
* Reuse repository conventions before creating abstractions.
* Record any scope or plan variance before modifying unplanned files.
* Do not commit or push changes.
