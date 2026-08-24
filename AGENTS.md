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
- `docs/CONVENTIONS.md` — engineering implementation rules, defaults, and enforcement;
- `.codex/CURRENT_STATE.md` — implementation state and known gaps;
- `.codex/TASK.md` — current task constraints.

`MVP-DELIVERY-ROADMAP.md` is the approved delivery-plan baseline. It is a
planning artifact, not canonical product truth or implementation evidence.

`.codex/` contains agent execution context, not canonical product or architecture documents.
`docs/CONVENTIONS.md` applies those canonical boundaries; it does not override
product requirements, architecture, or accepted decisions.

`.codex/TASK.md` is the active task file. Completed tasks are historical
records under `.codex/archive/` and must not remain as the active task.

## Task completion and archival

After an implementation task has been completed and verification confirms that
its objective was implemented, archive the active task before finishing:

1. Add this exact marker immediately below the task's H1 title in
   `.codex/TASK.md`:

   ```text
   ## Executed
   ```

2. Preserve the full executed task content.
3. Create the archive directory if needed:

   ```bash
   mkdir -p .codex/archive
   ```

4. Copy the task using this canonical filename:

   ```text
   .codex/archive/YYYY-MM-DD-TASK-<h1>.md
   ```

   `<h1>` is the text after `# Task:`, normalized to lowercase hyphen-separated
   words. Use the execution date in the local repository timezone. For example:

   ```text
   # Task: Frontend API Layers for Operations and Storefront Web
   → .codex/archive/2026-08-03-TASK-frontend-api-layers-for-operations-and-storefront-web.md
   ```

5. Use `cp`, verify the archive copy, and then remove the active task:

   ```bash
   cp .codex/TASK.md .codex/archive/YYYY-MM-DD-TASK-<h1>.md
   test -s .codex/archive/YYYY-MM-DD-TASK-<h1>.md
   cmp .codex/TASK.md .codex/archive/YYYY-MM-DD-TASK-<h1>.md
   rm .codex/TASK.md
   ```

6. Confirm `.codex/TASK.md` is absent and do not overwrite an existing archive
   file. Stop and report a filename collision instead.

Archive only after the task's acceptance criteria have been verified and the
final review distinguishes implemented, verified, assumed, and deferred
behavior. If verification is incomplete or the task is abandoned, keep
`.codex/TASK.md` active and record the incomplete status. Start each new task
with a new `.codex/TASK.md`; do not reactivate an archived task.

## Product boundaries

- The initial deployment is a Go modular monolith backed by PostgreSQL.
- Storefront and Operations remain separate Next.js React applications.
- Purchasing channels are `COMMON`, `SAVING`, and `GIVEAWAY` and converge into one canonical Purchase lifecycle.
- Purchaser, payer, saving-account holder, sponsor, giveaway applicant, giveaway recipient, and Sohibul Qurban are distinct roles.
- Sohibul Qurban is an outcome of eligible purchasing, not a purchasing channel.
- Saving Account becomes a Purchase only after conversion; giveaway applications become Purchases only after approval and assignment.
- Livestock is a lifecycle-managed physical entity; Allocation is a first-class transactional domain.
- Dashboard data is a projection, not transactional truth.
- A Full Event-Day MVP Event has exactly three or four inclusive local
  execution days in an explicit IANA timezone.
- Field teams, memberships, shifts, station/location assignments, readiness,
  handovers, incidents, and support escalation are Event-scoped.
- Sohibul Qurban attendance is explicitly `SELF`, `PROXY`, or `NONE` when the
  Event enables it; attendance does not determine payment eligibility.
- Distribution supports explicit Sohibul entitlement and beneficiary portions,
  pickup or delivery, proof, exceptions, and completion.
- Polling is the realtime baseline; SSE serves high-value one-way updates.
  WebSocket requires a separate proven bidirectional need.
- The responsive PWAs are the mobile baseline. Only allowlisted non-financial
  field milestones may queue during degraded connectivity; sensitive commands
  remain online-only.
- Do not add hypothetical SaaS multitenancy, mandatory `organisation_id`, microservices, or speculative infrastructure without an accepted ADR.

## Engineering rules

- Keep business rules in the Go API, not in frontend placeholders.
- Use Next.js App Router filesystem routes under `src/app/` and centralized URL
  builders in `src/routes/paths.ts`. Do not recreate a client route registry.
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
