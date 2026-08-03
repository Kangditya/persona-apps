# Product and Architecture Realignment Implementation Plan

> **For Hermes:** Execute only after this plan is reviewed and explicitly approved. Follow DISCOVER → PLAN → BUILD → VERIFY → REVIEW.

**Goal:** Align the repository documentation, agent context, README, and frontend placeholder terminology with the accepted Qurban Commerce and Operations Platform without implementing business functionality.

**Architecture:** Preserve the existing pnpm/Turborepo monorepo, two React/Vite frontends, Go modular monolith, PostgreSQL infrastructure, and separate public/operations OpenAPI contracts. This change is documentation and placeholder-shell alignment only; it must not introduce domain modules, migrations, authentication, payment behavior, multitenancy, or microservices.

**Tech Stack:** Markdown, React Router/TypeScript route placeholders, pnpm, Go tooling, Make, OpenAPI 3.1.

---

## Current discovery

- The repository has no root `AGENTS.md`; the active repository guidance is currently in `.codex/AGENTS.md`.
- Canonical product/architecture content currently exists in mixed locations:
  - `.codex/PRD.md` (untracked);
  - `docs/product/PRODUCT_MAP.md`;
  - `docs/architecture/ARCHITECTURE(1).md`;
  - `docs/decisions/DECISIONS.md`;
  - older supporting files under `docs/product/`, `docs/architecture/`, and `docs/decisions/`;
  - stale `.codex/DECISIONS.md` still exists.
- The intended canonical root files `docs/PRD.md`, `docs/PRODUCT_MAP.md`, `docs/ARCHITECTURE.md`, and `docs/DECISIONS.md` are not yet all present.
- The frontend shells still use generic bootstrap terms:
  - Operations: `/login`, `/management`, `/pos` and `ManagementPage`/`PosPage`.
  - Storefront: `/`, `/products`, `/cart` and `ProductsPage`/`CartPage`.
- `contracts/openapi/storefront.yaml` and `contracts/openapi/operations.yaml` already exist with placeholder `paths: {}` and must remain separate.
- The working tree already contains user changes in `.codex/AGENTS.md`, `.codex/ARCHITECTURE.md`, `.codex/CURRENT_STATE.md`, and documentation additions/deletions. BUILD must preserve those changes and avoid broad rewrites.

## Constraints and decisions to preserve

1. All `COMMON`, `SAVING`, and `GIVEAWAY` purchasing channels converge into one canonical Purchase lifecycle.
2. Sohibul Qurban is an outcome of eligible purchasing, not a purchasing channel.
3. Purchaser, payer, saving-account holder, sponsor, giveaway applicant, giveaway recipient, and Sohibul Qurban remain distinct roles.
4. Saving Account becomes a Purchase only after conversion; giveaway application becomes a Purchase only after approval and assignment.
5. Livestock is a lifecycle-managed physical entity, and Allocation is a first-class transactional domain.
6. Dashboard data is a projection, not transactional truth.
7. Initial deployment remains a Go modular monolith with PostgreSQL.
8. Do not introduce hypothetical SaaS multitenancy, mandatory `organisation_id`, microservices, payment gateways, authentication, or other speculative infrastructure.
9. Do not start the first business vertical slice in this task. The first slice remains Event → Offering → Common Purchase → Payment Verification → Sohibul Qurban Activation → Basic Operations Dashboard.
10. Do not commit or push.

## Proposed canonical documentation layout

```text
docs/
├── PRD.md
├── PRODUCT_MAP.md
├── ARCHITECTURE.md
├── DECISIONS.md
├── adr/                 # optional future ADR detail, only if needed
└── domain/              # optional future domain references, only if needed
```

The four root documents are the only canonical product/architecture sources. Existing nested documents should either be migrated into these files, reduced to clearly labelled supplementary references, or removed when they duplicate canonical content. No canonical document should remain under `.codex/`.

## Placeholder route terminology

Keep placeholders deliberately non-functional but capability-aligned:

### Operations Web

| Current       | Proposed                | Placeholder purpose                                             |
| ------------- | ----------------------- | --------------------------------------------------------------- |
| `/login`      | `/operator-login`       | Identity & Access placeholder; no authentication implementation |
| `/management` | `/event-dashboard`      | Event Operations and projection placeholder                     |
| `/pos`        | `/payment-verification` | Payment & Funding operations placeholder; no payment processing |
| —             | `/purchasing`           | Purchasing operations placeholder                               |

### Storefront Web

| Current     | Proposed             | Placeholder purpose                                    |
| ----------- | -------------------- | ------------------------------------------------------ |
| `/`         | `/`                  | Qurban event landing placeholder                       |
| `/products` | `/offerings`         | Qurban Offering Catalogue placeholder                  |
| `/cart`     | `/purchase-tracking` | Canonical Purchase tracking placeholder; no cart state |

Use centralized `src/routes/paths.ts` and `src/routes/routes.tsx`. Rename page components and navigation labels to match the proposed concepts. Do not add business data fetching or checkout behavior. The route tests should assert the new route set.

## Planned file changes

| File                                                                     | Action                               | Purpose                                                                                                                                                                    |
| ------------------------------------------------------------------------ | ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `docs/PRD.md`                                                            | create/migrate                       | Place the canonical product requirements document at the required root path.                                                                                               |
| `docs/PRODUCT_MAP.md`                                                    | create/migrate                       | Move the capability map from `docs/product/PRODUCT_MAP.md` to the canonical root path.                                                                                     |
| `docs/ARCHITECTURE.md`                                                   | create/migrate                       | Move/reconcile the accepted architecture from `docs/architecture/ARCHITECTURE(1).md` to the canonical root path.                                                           |
| `docs/DECISIONS.md`                                                      | create/migrate                       | Move/reconcile the accepted decision register from `docs/decisions/DECISIONS.md` to the canonical root path.                                                               |
| `docs/product/PRODUCT_MAP.md`                                            | delete or replace with redirect note | Prevent a second competing product-map source. Choose one approach during BUILD based on existing repository conventions.                                                  |
| `docs/architecture/ARCHITECTURE(1).md`                                   | delete or replace with redirect note | Remove the duplicate/non-canonical architecture filename.                                                                                                                  |
| `docs/decisions/DECISIONS.md`                                            | delete or replace with redirect note | Remove the duplicate decision-register location after root migration.                                                                                                      |
| `docs/product/scope.md`                                                  | modify/delete                        | Reconcile or remove stale generic/bootstrap scope text; retain only if clearly labelled as supplementary.                                                                  |
| `docs/architecture/overview.md`                                          | modify/delete                        | Reconcile or remove duplicate architecture overview and point readers to `docs/ARCHITECTURE.md`.                                                                           |
| `docs/decisions/README.md`                                               | already deleted; verify              | Confirm no stale tracked reference remains after migration.                                                                                                                |
| `.codex/PRD.md`                                                          | delete after migration               | Ensure product requirements are not canonical under `.codex/`. Preserve content in `docs/PRD.md`.                                                                          |
| `.codex/ARCHITECTURE.md`                                                 | modify or supersede                  | Keep only agent/task context if needed; remove canonical-document duplication and reference `docs/ARCHITECTURE.md`.                                                        |
| `.codex/DECISIONS.md`                                                    | delete/supersede                     | Replace the stale generic decision register with the canonical `docs/DECISIONS.md` reference.                                                                              |
| `.codex/AGENTS.md`                                                       | modify                               | Keep execution rules and Qurban terminology consistent with the root guidance and canonical paths.                                                                         |
| `.codex/CURRENT_STATE.md`                                                | modify                               | Record the realignment state, canonical paths, remaining bootstrap behavior, route status, and verification date/results.                                                  |
| `.codex/TASK.md`                                                         | modify                               | Replace the obsolete bootstrap objective with the current realignment objective and reviewed workflow.                                                                     |
| `AGENTS.md`                                                              | create                               | Add repository-root instructions so agents discover the workflow and canonical documents from the repository root. Keep it concise and consistent with `.codex/AGENTS.md`. |
| `README.md`                                                              | modify                               | Rename the product from generic commerce/POS to Qurban Commerce and Operations Platform; describe implemented foundation and explicitly deferred qurban capabilities.      |
| `apps/operations-web/src/routes/paths.ts`                                | modify                               | Replace generic route constants with capability-aligned placeholder paths.                                                                                                 |
| `apps/operations-web/src/routes/routes.tsx`                              | modify                               | Register the aligned placeholder routes and renamed page components.                                                                                                       |
| `apps/operations-web/src/routes/paths.test.ts`                           | modify                               | Verify the new centralized route set.                                                                                                                                      |
| `apps/operations-web/src/layouts/OperationsLayout.tsx`                   | modify                               | Update navigation labels and application terminology.                                                                                                                      |
| `apps/operations-web/src/pages/LoginPage.tsx` or renamed equivalent      | modify/rename                        | Rename “employee login” placeholder to operator identity/access terminology without implementing auth.                                                                     |
| `apps/operations-web/src/pages/ManagementPage.tsx` or renamed equivalent | modify/rename                        | Rename generic management placeholder to event dashboard terminology.                                                                                                      |
| `apps/operations-web/src/pages/PosPage.tsx` or renamed equivalent        | modify/rename                        | Remove POS terminology and use payment verification placeholder terminology.                                                                                               |
| `apps/operations-web/src/pages/PurchasingPage.tsx`                       | create if needed                     | Add the minimal purchasing operations placeholder required by the aligned route set.                                                                                       |
| `apps/storefront-web/src/routes/paths.ts`                                | modify                               | Replace `products`/`cart` constants with `offerings`/`purchaseTracking`, retaining `/` for event landing.                                                                  |
| `apps/storefront-web/src/routes/routes.tsx`                              | modify                               | Register the aligned storefront placeholder routes and renamed page components.                                                                                            |
| `apps/storefront-web/src/routes/paths.test.ts`                           | modify                               | Verify the new centralized route set.                                                                                                                                      |
| `apps/storefront-web/src/layouts/StorefrontLayout.tsx`                   | modify                               | Update navigation labels from generic Products/Cart to Offerings/Purchase Tracking.                                                                                        |
| `apps/storefront-web/src/pages/HomePage.tsx`                             | modify                               | Use Qurban event landing terminology.                                                                                                                                      |
| `apps/storefront-web/src/pages/ProductsPage.tsx` or renamed equivalent   | modify/rename                        | Use Qurban Offering Catalogue terminology.                                                                                                                                 |
| `apps/storefront-web/src/pages/CartPage.tsx` or renamed equivalent       | modify/rename                        | Use Purchase Tracking terminology and explicitly state that tracking is not implemented.                                                                                   |
| `contracts/openapi/storefront.yaml`                                      | verify, modify only if needed        | Preserve the public contract location and update stale product naming metadata if required; keep `paths: {}` until meaningful endpoints exist.                             |
| `contracts/openapi/operations.yaml`                                      | verify, modify only if needed        | Preserve the operations contract location and update stale product naming metadata if required; keep `paths: {}` until meaningful endpoints exist.                         |
| `packages/api-client/README.md`                                          | modify only if needed                | Ensure it points to the canonical OpenAPI locations and does not claim generated clients exist.                                                                            |

Before BUILD, compare this table with the actual working tree and record any necessary variance. Do not overwrite existing user edits without preserving their intent.

## Dependencies

No new dependencies are planned. Route renames use existing React Router/TypeScript code, and documentation migration uses repository files only.

## Database, API, authorization, audit, and concurrency impact

- Database: none. No migrations, schema changes, business tables, or seed data.
- API behavior: none. Existing `/health` and `/ready` behavior remains unchanged.
- OpenAPI: locations are verified; contracts remain placeholders with no generated clients.
- Authorization: none; operator login remains a non-functional placeholder.
- Audit/idempotency/concurrency: no runtime behavior is added. These remain requirements for the first vertical slice and must not be faked in the shells.
- Dashboard: only terminology changes; no dashboard projection or transactional endpoint is implemented.

## Build sequence after approval

1. Re-read all planned source files and current diffs; preserve user changes.
2. Reconcile/migrate the four canonical documents into `docs/` root.
3. Remove or clearly supersede duplicate/stale canonical documents under `.codex/` and nested `docs/` locations.
4. Create/update root `AGENTS.md`, then update `.codex/TASK.md` and `.codex/CURRENT_STATE.md`.
5. Update README terminology and deferred-capability statements.
6. Rename only the placeholder route/page terminology listed above; keep route ownership centralized.
7. Verify OpenAPI paths and references, without generating clients or adding business endpoints.
8. Run the verification commands below and review the diff for accidental generic terminology or scope expansion.

## Verification plan

Run from repository root:

```bash
pnpm install
pnpm run format:check
pnpm run lint
pnpm run typecheck
pnpm run test
pnpm run build
make validate
```

Also run the repository-specific checks required by the existing guidance:

```bash
cd apps/api && test -z "$(gofmt -l .)" && go vet ./... && go test ./... && go build ./...
cd ../..
docker compose -f infrastructure/compose.yaml config

git diff --check
git status --short
```

Targeted checks:

- `docs/PRD.md`, `docs/PRODUCT_MAP.md`, `docs/ARCHITECTURE.md`, and `docs/DECISIONS.md` exist.
- No canonical document remains at `.codex/DECISIONS.md` or a duplicate nested canonical path.
- Both OpenAPI files exist at `contracts/openapi/` and still contain valid OpenAPI metadata.
- Route tests assert only the capability-aligned placeholder paths.
- Search confirms no obsolete `/management`, `/pos`, `/products`, or `/cart` route constants remain, and no README claim describes generic POS/customer commerce as the product.
- No `.env`, credentials, build output, generated client, commit, or push is introduced.

## Risks and unresolved decisions

1. Existing working-tree documentation changes may contain the intended canonical text. Migration must preserve them rather than reconstructing documents from memory.
2. The exact long-term URL taxonomy is not a product decision. The proposed routes are temporary placeholder names and must not be treated as final API or navigation contracts.
3. Offering shape, multi-offering checkout, saving price lock, giveaway selection, personal slaughter flow, and distribution scope remain open product decisions. This task must not encode them.
4. `docs/PRD.md` content must be reconciled from the available `.codex/PRD.md` and current product direction before deleting the `.codex` copy.
5. The repository has a pre-existing dirty working tree. Review must distinguish pre-existing changes from this task's changes.

## Completion criteria for this planning task

- This plan is saved under `.hermes/plans/`.
- No repository implementation or documentation file other than this plan is modified.
- BUILD remains blocked until the user reviews and explicitly authorizes execution.
