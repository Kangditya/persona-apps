# Task: Product and Architecture Realignment

## Objective

Align the repository documentation, agent context, README, and frontend placeholder terminology with the accepted Qurban Commerce and Operations Platform.

This task does not implement qurban business functionality. It preserves the existing React/Vite shells, Go modular monolith API shell, PostgreSQL infrastructure, and placeholder OpenAPI contracts.

## Required workflow

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

The implementation plan must be reviewed and explicitly approved before BUILD. Do not commit or push.

## Canonical documents

The canonical documents must live at:

```text
docs/PRD.md
docs/PRODUCT_MAP.md
docs/ARCHITECTURE.md
docs/DECISIONS.md
```

`.codex/` contains agent context only. `.codex/DECISIONS.md` is superseded and must not remain canonical.

## Scope

1. Migrate or reconcile the canonical documents into `docs/`.
2. Update root `AGENTS.md` and this execution context.
3. Update `.codex/CURRENT_STATE.md` after the changes are verified.
4. Update README terminology from generic commerce/POS to qurban terminology.
5. Replace outdated placeholder route terminology:
   - Operations: `/operator-login`, `/event-dashboard`, `/purchasing`, `/payment-verification`.
   - Storefront: `/`, `/offerings`, `/purchase-tracking`.
6. Verify the separate OpenAPI contracts at `contracts/openapi/storefront.yaml` and `contracts/openapi/operations.yaml`.
7. Run `make validate` and the required repository checks.

## Explicit non-scope

Do not implement:

- qurban business modules;
- database migrations or business tables;
- authentication or authorization behavior;
- payment processing or gateway integration;
- generated API clients;
- cart state or checkout behavior;
- microservices or hypothetical SaaS multitenancy;
- commits or pushes.

## Product rules to preserve

- `COMMON`, `SAVING`, and `GIVEAWAY` are purchasing channels that converge into one canonical Purchase lifecycle.
- Sohibul Qurban is an outcome of eligible purchasing, not a channel.
- Purchaser, payer, saving-account holder, sponsor, giveaway applicant, giveaway recipient, and Sohibul Qurban remain distinct roles.
- Saving Account becomes a Purchase only after conversion.
- Giveaway application becomes a Purchase only after approval and assignment.
- Livestock is a lifecycle-managed physical entity.
- Allocation is a first-class transactional domain.
- Dashboard data is a projection, not transactional truth.

## Verification

At minimum:

```bash
pnpm run format:check
pnpm run lint
pnpm run typecheck
pnpm run test
pnpm run build
make validate
cd apps/api && test -z "$(gofmt -l .)" && go vet ./... && go test ./... && go build ./...
cd ../..
docker compose -f infrastructure/compose.yaml config
git diff --check
git status --short
```
