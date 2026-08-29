# Deprecated Artifact Snapshot

This file is a noncanonical historical artifact snapshot. It is retained only
for repository history and must not be used as a source of product,
architecture, implementation, or agent-execution guidance.

Use these current sources instead:

- `AGENTS.md` for repository instructions;
- `.codex/AGENTS.md` and `.codex/TASK.md` for current execution context;
- `docs/PRD.md` and `docs/PRODUCT_MAP.md` for product scope;
- `docs/ARCHITECTURE.md` for technical architecture;
- `docs/DECISIONS.md` for accepted and superseded decisions;
- `.codex/CURRENT_STATE.md` for implemented and deferred state.

The current frontend runtime is Next.js App Router with centralized URL
builders and TanStack Query. The Go API remains authoritative for business
rules, authorization, transactions, audit, idempotency, and contested state.

Use `MVP-DELIVERY-ROADMAP.md` for approved task sequencing. Do not add current
guidance to this deprecated snapshot.
