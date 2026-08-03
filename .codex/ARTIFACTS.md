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

## Current frontend direction

The approved direction is React Router with Remix-style routing conventions in
the existing Vite-served React SPAs, and TanStack Query for future remote
API/server state. This does not add a Remix server runtime, SSR, server actions,
routing code, a TanStack Query dependency, providers, queries, mutations, or
cache behavior.

The Go API remains authoritative for business rules, authorization,
transactions, audit, idempotency, and contested state. Client cache data is not
transactional truth.
