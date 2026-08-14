# Current Repository State

## Repository

```text
Repository: persona-apps
GitHub owner: Kangditya
Primary branch: main
```

## Current Condition

The repository bootstrap is implemented.

It currently contains:

- two frontend application shells;
- one Go API shell;
- shared workspace tooling;
- PostgreSQL local infrastructure;
- Phase 1 API contract baselines;
- documentation;
- CI and repository validation commands.

The Event and Offering domain cores, PostgreSQL repositories, additive
version/bounds migration, public catalogue routes, and permission-gated
Operations command/query routes are implemented. Operations Event/Offering and
Storefront public catalogue screens are implemented.

The product direction has changed from a generic single-brand commerce and POS platform into a:

> Qurban Commerce and Operations Platform

The canonical product and architecture documents now describe:

- annual Qurban Events;
- Common Purchasing;
- Saving Purchasing;
- Giveaway Purchasing;
- Sohibul Qurban;
- Payment and Funding;
- Livestock;
- Allocation;
- Event Operations;
- Distribution;
- Administration and Reporting.

---

## Canonical Documentation

The canonical documentation set is:

```text
docs/
├── PRD.md
├── PRODUCT_MAP.md
├── ARCHITECTURE.md
└── DECISIONS.md
```

Agent execution context is expected under:

```text
.codex/
├── AGENTS.md
├── CURRENT_STATE.md
└── TASK.md
```

The previous `.codex/DECISIONS.md` location is no longer canonical.

---

## Implemented Foundation

### Repository and Tooling

- Git repository initialization;
- repository-local Git identity;
- GitHub remote configuration;
- pnpm workspace;
- Turborepo command orchestration;
- repository Make targets;
- formatting, linting, testing, build, and validation workflows;
- GitHub-publishable repository baseline.

### Frontend Applications

Implemented application shells:

```text
apps/
├── operations-web/
└── storefront-web/
```

Current frontend stack:

- Vite;
- React;
- TypeScript;
- React Router with documented Remix-style routing conventions;
- TanStack Query as the approved future remote API/server-state library;
- Tailwind CSS;
- Oxlint;
- Vitest.

The applications now include a TanStack Query provider, typed API transport
boundaries, and a non-authoritative API-availability diagnostic. Storefront and
Operations use separate route registries and API surfaces. Operations has
authenticated Event/Offering list, create, detail, edit, lifecycle, conflict,
and logout flows against the W2-05 API. Storefront has active-Event landing,
published-Offering list/detail, safe public states, exact minor-unit price, and
advisory availability flows against the W2-04 API.

Both applications also consume `@persona-apps/ui`, which owns shared semantic
tokens and accessible atoms/molecules, and have independent production PWA
manifests and shell-only service workers. API and sensitive data remain
network-only.

Local development behavior:

- Storefront Vite server with HMR on `127.0.0.1:5173`;
- Operations Vite server with HMR on `127.0.0.1:5174`;
- Go API live reload through Air using `apps/api/.air.toml`;
- API default address `127.0.0.1:8080`.

Current placeholder routes:

#### Operations Web

```text
/operator-login
/event-dashboard
/purchasing
/payment-verification
```

#### Storefront Web

```text
/purchase-tracking
```

`/operator-login` now owns the real OIDC session/sign-in/logout state. The
remaining Operations routes and Storefront `/purchase-tracking` route in these
lists are capability-aligned placeholders. Event/Offering administration is
implemented under Operations `/events`; public discovery is implemented under
Storefront `/` and `/offerings`.

### Backend

Implemented:

```text
apps/api
```

Current backend capabilities:

- Go API process;
- `/health`;
- PostgreSQL-backed `/ready`;
- graceful shutdown;
- Air-compatible local live-reload configuration;
- PostgreSQL connectivity;
- Docker image build.
- guest `GET /api/public/v1/events/active`, Event Offering-list, and Offering
  detail routes with active/published filtering;
- request-correlated public JSON errors, recovery, no-store catalogue headers,
  bounded per-process guest rate limiting, and disabled-by-default proxy trust.
- permission-gated Operations Event/Offering list, detail, create, patch, and
  lifecycle routes under `/api/operations/v1`;
- strict 64 KiB JSON command decoding, exact Origin/CSRF checks, optimistic
  versions, 24-hour command replay, minimized audit, and lifecycle outbox
  writes in caller-owned transactions;
- exact-origin credentialed Operations CORS and optional exact-origin
  Storefront CORS, both disabled when no origins are configured.

Framework-neutral Event and Offering modules enforce validated annual and
commercial configuration, version-aware lifecycle transitions, immutable
terminal states, and the audit/outbox/retry metadata required by the later
transactional adapter. Their concrete PostgreSQL repositories use bounded
keyset reads, conditional version updates, a caller-provided transaction, and
one-query advisory availability aggregation. Offering availability uses Event
and Offering quota snapshots with only reserved/consumed units counted; it
does not reserve quota. Public catalogue reads expose only minimized active and
published DTOs with advisory availability.

Implemented platform foundations:

- Gin as the canonical HTTP framework and router at the HTTP adapter/bootstrap
  boundary, with `net/http.Server` retained for lifecycle and transport;
- the existing ADR-040 `golang-migrate/migrate/v4` database lifecycle;
- provider-neutral Operations OIDC Authorization Code with PKCE when complete
  OIDC configuration is supplied;
- hashed, revocable server-side Operations sessions with permission snapshots,
  exact Origin checks, CSRF protection, and allowlisted permissions;
- request-ID middleware, structured error envelopes, transaction helper,
  caller-transaction audit writer, and encrypted idempotency replay executor.

ADR-043 through ADR-045 define these contracts. Event/Offering Operations
commands now compose the platform foundations transactionally; Storefront
Purchase access remains unimplemented. Auth and protected business routes exist
only when the complete fail-closed configuration is present; no secret value is
committed.

### Infrastructure

Implemented:

- Docker Compose PostgreSQL service;
- configurable PostgreSQL host port;
- local environment bootstrap;
- container configuration validation.
- provider-neutral staging deployment contract;
- CI PostgreSQL 18 migration, reference-seed idempotency, database-test,
  bounded rollback, and reapply coverage.

No cloud environment or production deployment configuration is implemented.

### Contracts and Shared Packages

Implemented as contracts, placeholders, or shells:

- separate Storefront and Operations OpenAPI contracts for the Phase 1
  catalogue, common Purchase, OIDC session, Event, Offering, payment-review,
  participant, dashboard, and audit surfaces;
- shared workspace packages;
- frontend TypeScript configuration;
- API client package shell;
- shared UI package with tokens and accessible primitives;
- application-owned PWA manifests and service-worker configuration.

The separate public and operations OpenAPI contracts define Phase 1 request,
response, permission, request-ID, CSRF, idempotency, and error behavior. Public
and Operations Event/Offering contract paths have matching API routes; Purchase
and later Operations groups remain contract-only. Storefront Event/Offering
discovery now consumes only the minimized public contract.

### Database lifecycle tooling

Implemented:

- explicit Go database CLI at `apps/api/cmd/db` using
  `golang-migrate/migrate/v4`;
- migration validation, status, version, up, bounded down, and skeleton
  creation against `apps/api/migrations`;
- separate ordered seed registry/history with development safety guards;
- `db setup` for migrations plus reference seeds only;
- repository Make targets for the complete normal CLI surface.

The proposed qurban business schema is documented in `docs/database/ERD.md`
and scripted in the six numbered migration pairs under
`apps/api/migrations`. The schema includes foundation, commerce/funding,
operations, audit, outbox, command idempotency, seed metadata, and Phase 1
commerce-safety constraints for Event suspension, Offering quota, intended
participants, quota reservations, evidence metadata, Purchase tokens, and
Operations sessions. Event/Offering repositories and both public reads and
privileged commands use the relevant tables.

The reference seed group is intentionally empty. The development group only
contains `development.sample-event`. No migration or seed has been run against
staging or production.

### Documentation

Implemented or revised:

- product requirements;
- product capability map;
- architecture baseline;
- architecture decision register;
- repository agent instructions;
- current-state tracking;
- database ERD, migration plan, operations guide, and design review;
- command-scoped idempotency ownership and development guidance.

The architecture artifacts document React Router with Remix-style routing
conventions and TanStack Query server-state ownership. The approved UI/API
foundation is now runtime code; qurban business capabilities remain deferred.

---

## Product Direction

The accepted product capabilities are:

```text
Qurban Platform
├── Storefront
├── Purchasing
├── Party & Participant
├── Payment & Funding
├── Livestock
├── Allocation
├── Event Operations
├── Distribution
├── Identity & Access
└── Administration & Reporting
```

The accepted purchasing channels are:

```text
COMMON
SAVING
GIVEAWAY
```

All eligible channels converge into one canonical Purchase lifecycle.

Sohibul Qurban is not a purchasing channel.

The following roles must remain distinct:

- purchaser;
- payer;
- saving-account holder;
- sponsor;
- giveaway applicant;
- giveaway recipient;
- Sohibul Qurban.

### Documented Phase 1 Commerce Rules

ADR-042 now fixes the Phase 1 Event lifecycle, one-active-event rule, MVP
Offering boundary, one-Offering direct checkout, participant-unit quota
reservation, append-oriented payment evidence, and atomic exactly-once Sohibul
Qurban activation.

Migration 0005 now scripts the supporting Offering quota, quota-reservation,
participant, evidence, token, session, and audit storage. No runtime command
currently enforces lifecycle, evidence, quota, payment-verification, or
participant-activation behavior. ADR-044, Commerce Lifecycles, and
Permissions now define the required Phase 1 lifecycle and authorization
policy; W1-05 and W1-06 own endpoint and runtime follow-up.

---

## Not Implemented

### Purchasing

- canonical Purchase aggregate;
- Common Purchasing;
- Saving Purchasing;
- Giveaway Purchasing;
- checkout;
- purchase validation;
- purchase confirmation;
- purchase cancellation;
- purchase history.

### Party and Participant

- person and organization identity;
- purchaser;
- payer;
- saving-account holder;
- sponsor;
- giveaway applicant;
- giveaway recipient;
- Sohibul Qurban;
- participant verification;
- duplicate detection.

### Payment and Funding

- payment methods;
- payment instructions;
- payment confirmation;
- payment verification;
- installment ledger;
- saving balances;
- sponsor funding;
- refund;
- reconciliation;
- payment gateway integration.

### Saving

- saving account;
- saving target;
- installment workflow;
- funding state;
- price-lock policy;
- conversion eligibility;
- saving-to-purchase conversion.

### Giveaway

- giveaway programs;
- sponsor funding;
- applicant or nominee intake;
- eligibility review;
- recipient selection;
- approval;
- purchase assignment.

### Livestock

- livestock registry;
- classification;
- inspection;
- weight;
- readiness;
- pen or location assignment;
- physical lifecycle;
- livestock history.

### Allocation

- purchase allocation;
- Sohibul Qurban allocation;
- shared livestock capacity;
- provisional allocation;
- confirmed allocation;
- release;
- reassignment;
- allocation manifests.

### Event Operations

- event readiness;
- participant check-in;
- livestock check-in;
- slaughter schedule;
- slaughter queue;
- slaughter station;
- live status tracking;
- incidents;
- event completion.

### Distribution

- distribution planning;
- portion preparation;
- Sohibul Qurban entitlement;
- beneficiary assignment;
- pickup;
- delivery;
- collection confirmation;
- delivery proof;
- distribution completion.

### Identity and Access

- storefront authentication;
- user profiles;
- roles;
- event-scoped access;
- audit access.

### Administration and Reporting

- purchasing administration;
- payment verification queues;
- saving administration;
- giveaway administration;
- participant administration;
- livestock administration;
- allocation administration;
- operational dashboards;
- reports;
- exports;
- audit log interface.

### Platform Capabilities

- disposable-PostgreSQL migration integration verification;
- background worker;
- notification delivery;
- operational projections;
- polling or SSE dashboard updates;
- production observability;
- production deployment.

---

## Previous Bootstrap Assumptions Now Superseded

The following previous assumptions are no longer accepted product direction:

- POS as a primary product capability;
- generic back-office commerce;
- generic catalogue and inventory;
- generic customer ordering;
- cart as a mandatory workflow;
- future SaaS as the default architecture;
- shared-schema multitenancy by default;
- mandatory `organisation_id` on future business tables;
- generic fulfillment as the primary operational abstraction.

Current replacements:

| Previous Assumption | Current Direction |
|---|---|
| Product catalogue | Qurban Offering Catalogue |
| Inventory | Livestock lifecycle |
| Customer | Explicit party and participant roles |
| Order | Canonical Purchase |
| POS operations | Qurban Event Operations |
| Fulfillment | Allocation, Slaughter, and Distribution |
| SaaS multitenancy | One operating organization initially |
| Mandatory cart | Direct checkout unless multi-offering checkout is confirmed |

---

## Current Verification State

Last recorded implementation verification: **2026-08-14**

The product/architecture and frontend-artifact alignment has been verified
with:

```bash
make validate
```

W2-05 through W2-08 are implemented and verified. The Event/Offering slice
passed `make validate`, the full Go suite with database integration enabled
against PostgreSQL 18, Go vet/build, focused race tests, a fresh-database
migration/seed/down/up cycle, Compose validation, and an API container build.
The composed Operations regression proves exact audit, outbox, and replay
effects plus public active/published visibility and historical Event quota and
Offering price values.

Redocly CLI 2.46.1 validates both OpenAPI contracts with no errors; six
pre-existing documentation warnings remain. Production pnpm audit reports no
known vulnerabilities. A current `govulncheck` 1.7.0 scan found six reachable
Go 1.26.5 standard-library advisories, so both the module and container builder
are now pinned to fixed Go 1.26.6. The 1.26.6 rescan reports zero reachable or
imported-package vulnerabilities. The sole module-only result is the uncalled,
unimported, upstream-unfixed `golang.org/x/crypto/openpgp` advisory.

Frontend type checking, focused tests, linting, builds, PWA manifest output,
and service-worker precache policy are verified. Operations Event/Offering
screens were exercised in a real browser against the real local API and
PostgreSQL, including OIDC session bootstrap, create/edit/lifecycle commands,
stale-version recovery, rotated-CSRF recovery, logout, direct route loading,
and a 360-pixel layout. The temporary local OIDC verifier used only disposable
test identity data and was not shipped. Storefront Event/Offering screens were
also exercised against the real local API and PostgreSQL for populated, empty,
no-active-Event, not-found, rate-limited, dependency-failure, timezone, and
mobile states. Public fetches omit credentials, and generated service-worker
output excludes `/api/` from navigation fallback and runtime caching. Browser
fixtures, viewport changes, tabs, and the temporary verifier were cleaned up.
Physical-keyboard tab order and 200-percent browser zoom remain explicit
release-QA checks because the in-app harness could not inject them. All
Purchase behavior remains pending.

The API provides `/health`, PostgreSQL-backed `/ready`, graceful shutdown, the
bounded guest Event/Offering catalogue, and transactional Operations
Event/Offering commands. It has no Purchase command capability yet.

---

## Known Constraints

1. Development is currently handled by one developer.
2. The initial budget is minimal.
3. The repository must remain understandable and maintainable.
4. The application must support up to approximately 1,000 Sohibul Qurban per annual event.
5. The operational record count may reach thousands across purchases, payments, livestock, allocations, slaughter events, and distribution.
6. PostgreSQL may already run locally for another project.
7. The local PostgreSQL host port must remain configurable.
8. No secrets may be committed.
9. The repository may remain publicly accessible during early development.
10. Public repository status requires strict exclusion of credentials, production data, and private participant information.
11. Microservices are prohibited without an accepted ADR and demonstrated need.
12. Generic SaaS multitenancy is not part of the initial product.
13. Dashboard requirements are near-real-time, not hard real-time.
14. Product rules must remain adjustable while Figma and field requirements are still being refined.

---

## Open Product Decisions

The following decisions remain unresolved:

1. Whether offerings represent:
   - individual livestock;
   - categories;
   - packages;
   - cattle shares;
   - or a combination.

2. Whether one checkout may contain multiple offerings.

3. Whether Saving Purchasing locks:
   - the offering;
   - the price;
   - both;
   - or neither.

4. How giveaway recipients are selected:
   - sponsor selection;
   - committee selection;
   - manual approval;
   - random draw;
   - combined process.

5. Whether every Sohibul Qurban personally performs the slaughter.

6. Whether personal slaughter requires:
   - attendance registration;
   - check-in;
   - personal queue number;
   - assigned slaughter station;
   - proxy representation.

7. Whether distribution covers:
   - Sohibul Qurban entitlement;
   - beneficiaries;
   - pickup;
   - delivery;
   - or a combination.

These rules must not be invented during implementation.

---

## Current Risks

1. Event/Offering APIs and both owning frontend flows exist, but all Purchase
   flows remain unimplemented.
2. The public limiter is intentionally per process. Ingress/CDN enforcement is
   still required for a uniform multi-replica rate limit and key-filling abuse.
3. Migration/repository verification used only an explicitly disposable local
   PostgreSQL database; staging and production compatibility remain unproven.
4. OIDC-backed Operations sessions and Event/Offering permissions are enforced;
   operator/role administration and a real configured provider test identity
   remain deployment work.
5. The domain model has public-query and Operations-command proof but no
   Purchase vertical slice yet.
6. Week 2 writes transactional outbox rows but has no background publisher,
   retry, cleanup, or backlog monitoring yet.
7. CSP, HSTS, frame protection, and uniform multi-replica rate limiting depend
   on the eventual hosting/ingress configuration and are not proven by Vite.
8. Full WCAG, cross-browser, load/soak, penetration, disaster-recovery, audit
   retention/export, and staging/production verification remain outside Week 2.

---

## Recommended Next Task

Execute `.codex/plans/week-3/W3-01-implement-reusable-party-identity-records.md`
next. Reusable Party identity is the dependency-ready foundation for explicit
purchaser, payer, and intended Sohibul Qurban relationships before the canonical
Common Purchase task.

The recommended first slice remains:

```text
Qurban Event
→ Offering
→ Common Purchase
→ Payment Verification
→ Sohibul Qurban Activation
→ Basic Operations Dashboard
```

---

## Milestone Status

### Completed

```text
Repository bootstrap
Week 2 Event and Offering discovery/administration
```

### In Progress

```text
First qurban vertical slice
```

### Not Started

```text
Common Purchase vertical slice
```
