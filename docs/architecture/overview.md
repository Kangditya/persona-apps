# Architecture Overview

Persona Apps is a monorepo containing two React applications and one Go
modular monolith backed by PostgreSQL.

```text
Operations web ─┐
                ├─> Go API ─> PostgreSQL
Storefront web ─┘
```

Frontend URL paths and route trees are centralized within each application.
Future backend modules will live under `apps/api/internal/modules` and follow
transport → application → domain dependency direction, with infrastructure
implementing domain interfaces.

OpenAPI contracts are separated for private operations and public storefront
consumers. Generated clients will be introduced only after meaningful
endpoints exist.
