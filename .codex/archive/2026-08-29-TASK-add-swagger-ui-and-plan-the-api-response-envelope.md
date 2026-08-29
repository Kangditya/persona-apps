# Task: Add Swagger UI and Plan the API Response Envelope

## Executed

## Objective

Expose the existing Storefront and Operations OpenAPI contracts through a
Swagger UI served by the Go API, including local and containerized execution.
Produce an implementation plan for a future response envelope with `success`,
`data`, `error`, and tracing `metadata` without changing the current v1
payloads in this task.

## Canonical Sources

- `docs/PRD.md`
- `docs/PRODUCT_MAP.md`
- `docs/ARCHITECTURE.md`
- `docs/DECISIONS.md`
- `docs/CONVENTIONS.md`
- `.codex/CURRENT_STATE.md`
- `contracts/openapi/storefront.yaml`
- `contracts/openapi/operations.yaml`

## Current State

- The Go API uses Gin and exposes the documented public and Operations API
  routes.
- Both canonical OpenAPI contracts exist at `contracts/openapi/`.
- No Swagger UI route or Swagger UI dependency is currently present.
- The API currently returns the v1 response shapes documented by those
  contracts; a response envelope would be a compatibility-sensitive change.

## Scope

### Implement now

- Serve Swagger UI from the Go API at `/swagger`.
- Serve both canonical YAML contracts from API-owned documentation routes.
- Use the smallest dependency footprint possible; do not add generated clients
  or duplicate the canonical contracts.
- Package the contracts with the production API image and update the image
  build context used by CI.
- Add focused route tests and developer documentation.

### Plan only

- Define the response-envelope shape and migration strategy.
- Keep the current `/api/public/v1` and `/api/operations/v1` responses and
  OpenAPI contracts unchanged until a separate compatibility decision is
  approved.

### Out of scope

- Rewriting the existing API contracts for the envelope.
- Changing business handlers, frontend clients, health/readiness probe
  payloads, or idempotency response bodies for the envelope.
- Adding generated Swagger clients, a second API documentation source, or
  speculative authentication flows for Swagger UI.

## Acceptance Criteria

- `GET /swagger` reaches the Swagger UI page.
- The page loads the Storefront and Operations contracts from the same API.
- The served contracts are the canonical repository files, not copies.
- The API image contains the contracts required by the UI.
- Focused tests prove the UI route, both contract routes, content types, and
  missing-contract behavior.
- Documentation explains the local URL and the response-envelope migration
  plan, including compatibility, pagination, error, timestamp, and request-ID
  decisions.
- Repository verification passes for the affected Go and container paths.

## Verification

- Run Go formatting, vet, build, and tests.
- Run the repository validation target and Docker Compose configuration check.
- Build the API image with the updated context.
- Start the API with the test configuration and verify `/swagger`, both YAML
  routes, and `/health`.
- Review the final diff for scope, security, and preserved v1 behavior.
