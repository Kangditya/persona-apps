# Repository Agent Instructions

This repository is the Qurban Commerce and Operations Platform in the `Kangditya/persona-apps` monorepo.

## Required workflow

Follow:

```text
DISCOVER → PLAN → BUILD → VERIFY → REVIEW
```

Do not begin BUILD until the plan is reviewed and explicitly approved. Do not commit or push unless explicitly requested.

## Source of truth

Read these canonical documents before planning or implementation:

- `docs/PRD.md` — product goals, actors, requirements, and rules;
- `docs/PRODUCT_MAP.md` — capabilities, application ownership, and roadmap;
- `docs/ARCHITECTURE.md` — runtime, boundaries, data, API, and deployment architecture;
- `docs/DECISIONS.md` — accepted and superseded decisions;
- `.codex/CURRENT_STATE.md` — implementation state and known gaps;
- `.codex/TASK.md` — current task constraints.

`.codex/` contains agent execution context, not canonical product or architecture documents.

## Product boundaries

- The initial deployment is a Go modular monolith backed by PostgreSQL.
- Storefront and Operations remain separate React applications.
- Purchasing channels are `COMMON`, `SAVING`, and `GIVEAWAY` and converge into one canonical Purchase lifecycle.
- Purchaser, payer, saving-account holder, sponsor, giveaway applicant, giveaway recipient, and Sohibul Qurban are distinct roles.
- Sohibul Qurban is an outcome of eligible purchasing, not a purchasing channel.
- Saving Account becomes a Purchase only after conversion; giveaway applications become Purchases only after approval and assignment.
- Livestock is a lifecycle-managed physical entity; Allocation is a first-class transactional domain.
- Dashboard data is a projection, not transactional truth.
- Do not add hypothetical SaaS multitenancy, mandatory `organisation_id`, microservices, or speculative infrastructure without an accepted ADR.

## Engineering rules

- Keep business rules in the Go API, not in frontend placeholders.
- Use centralized frontend route registries in `src/routes/paths.ts` and `src/routes/routes.tsx`.
- Keep public and operations OpenAPI contracts separate at:
  - `contracts/openapi/storefront.yaml`;
  - `contracts/openapi/operations.yaml`.
- Add migrations, audit records, idempotency, concurrency safeguards, and authorization when implementing their applicable business behavior.
- Do not add empty domain module shells without an active vertical slice.
- Never commit secrets, `.env`, generated build output, or generated API clients.

## Verification

Run the applicable repository checks, including:

```bash
make validate
cd apps/api && go vet ./... && go test ./... && go build ./...
docker compose -f infrastructure/compose.yaml config
```

Finish with a review that distinguishes implemented, verified, assumed, and deferred behavior, and identifies remaining risks and the next recommended task.
