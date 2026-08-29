# API Client

`@persona-apps/api-client` currently owns the application-neutral TypeScript
transport primitive: request serialization, timeout and cancellation handling,
response parsing, and normalized transport errors. Storefront and Operations
keep their endpoint modules, DTOs, credentials, and contract ownership inside
their respective applications.

The separate OpenAPI contracts remain at:

- `contracts/openapi/storefront.yaml`
- `contracts/openapi/operations.yaml`

No generated client is committed and no generation command exists. Introducing
generated clients requires an approved generation and stale-output verification
workflow.
