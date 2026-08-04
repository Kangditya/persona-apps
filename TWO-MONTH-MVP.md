# Two-Month Qurban Platform MVP

## 1. Objective

Deliver a production-oriented MVP in eight weeks for one complete Common Purchasing journey:

```text
Qurban Event
→ Offering Catalogue
→ Common Purchase
→ Manual Payment Verification
→ Sohibul Qurban Activation
→ Purchase Tracking
→ Basic Operations Dashboard
```

The MVP validates the platform's central business model without attempting to implement every capability in the PRD.

## 2. Delivery assumptions

- One experienced full-time developer using coding-agent assistance.
- Eight weeks and approximately 320 human engineering hours.
- Product decisions and simple UI flows are approved during Week 1.
- One Offering is purchased through one direct checkout; no cart.
- Payment verification is manual; no payment-gateway automation.
- The existing React, Go, PostgreSQL, OpenAPI, PWA, and monorepo foundations remain in place.
- Work follows vertical slices and keeps the Go API authoritative.
- Scope is frozen after Week 1 unless an existing acceptance criterion is impossible or unsafe.
- Week 8 is reserved for stabilization and pilot preparation, not new features.

## 3. MVP scope

### 3.1 Storefront Web

The Storefront MVP must allow a public user to:

- view the active Qurban Event;
- browse published Offerings;
- view price, participant capacity, and availability;
- start one direct Common Purchase;
- enter purchaser, payer, and Sohibul Qurban information separately;
- receive payment instructions;
- submit a payment reference or evidence;
- view Purchase, payment, and eligibility status;
- retrieve a Purchase using a human-readable reference.

### 3.2 Operations Web

The Operations MVP must allow an authorized operator to:

- sign in;
- create, publish, suspend, close, and read an Event;
- create, publish, unpublish, and read Offerings;
- view and filter Common Purchases;
- inspect purchaser, payer, and participant relationships;
- view submitted payment evidence;
- verify or reject a Payment;
- activate Sohibul Qurban after Purchase eligibility is satisfied;
- view basic Event metrics and payment-verification backlog;
- inspect audit history for privileged changes.

### 3.3 Go API

The API owns:

- Event lifecycle and event-scoped configuration;
- Offering publication, pricing snapshot, participant capacity, and quota;
- Party identities and explicit purchaser, payer, and Sohibul Qurban relationships;
- canonical Purchase aggregate with `COMMON` channel;
- append-oriented Payment records and manual verification;
- Sohibul Qurban activation after eligibility;
- public and Operations authorization policies;
- PostgreSQL transactions and constraints;
- idempotency for retry-sensitive commands;
- append-only audit records;
- structured errors and request correlation;
- basic dashboard queries or rebuildable projections.

### 3.4 API contracts

The MVP must add meaningful endpoints to:

```text
contracts/openapi/storefront.yaml
contracts/openapi/operations.yaml
```

Public and Operations APIs may share application services but must use separate DTOs, authorization, and data-exposure rules.

### 3.5 Data and reliability

The MVP includes:

- versioned PostgreSQL migrations;
- UUID-based public references;
- exact money representation;
- foreign keys and constraints for critical invariants;
- pricing snapshots for historical truth;
- quota concurrency protection;
- idempotency records;
- payment and status history;
- append-only audit records;
- database backup and demonstrated restore procedure.

### 3.6 Verification and deployment

The MVP includes:

- domain unit tests;
- application-service tests;
- PostgreSQL repository integration tests;
- API contract tests;
- authorization and audit tests;
- idempotency and quota-concurrency tests;
- Storefront and Operations component/route tests;
- one complete end-to-end Common Purchase scenario;
- staging deployment;
- structured logs, request IDs, health, readiness, and basic alerts;
- controlled user-acceptance test and pilot rehearsal.

## 4. Explicitly deferred

The following capabilities are outside the two-month MVP:

- Saving Purchasing and installment conversion;
- Giveaway Programs and recipient assignment;
- payment-gateway callbacks and automatic bank reconciliation;
- refunds, transfers, and advanced financial adjustments;
- Livestock Registry, inspection, and location management;
- participant-to-livestock Allocation;
- slaughter sessions, queues, stations, and live execution;
- Server-Sent Events and WebSocket updates;
- low-connectivity or offline event-day mutations;
- Distribution, pickup, delivery, and proof;
- certificates and advanced document generation;
- advanced analytics, exports, and historical archive UI;
- multiple Offerings in one checkout;
- automatic identity deduplication;
- generalized administration or SaaS multitenancy;
- Redis, message brokers, and microservices.

Deferred capabilities must not be represented by fake workflows or placeholder business data in the production MVP.

## 5. Eight-week delivery plan

## Week 1 — Decisions, contracts, and platform safety

### Goal

Freeze the MVP behavior and establish the minimum safe backend foundation.

### Work items

- Resolve MVP Event, Offering, direct-checkout, quota, payment-evidence, and participant-activation rules.
- Select and document the Go HTTP router, migration tooling, and authentication approach.
- Define Event, Offering, Party, Purchase, Payment, Sohibul Qurban, Audit, and Idempotency schemas.
- Define status transitions and permission matrix.
- Draft Storefront and Operations OpenAPI endpoints and error shapes.
- Establish migration runner, transaction boundary, module registration, request IDs, audit writer, and idempotency foundation.
- Configure staging environment and CI migration checks.

### Exit criteria

- No unresolved decision blocks the Common Purchase journey.
- Schemas, state transitions, endpoint list, and permissions are reviewed.
- A migration applies and rolls forward in a clean PostgreSQL database.
- Authenticated Operations and safe public API boundaries are demonstrated.

## Week 2 — Event and Offering vertical slice

### Goal

Allow Operations to configure commerce and Storefront to discover it.

### Work items

- Implement Event lifecycle and historical configuration.
- Implement Offering publication, price, capacity, and availability.
- Add Event and Offering migrations and repositories.
- Add public Event/Offering queries.
- Add Operations Event/Offering commands and queries.
- Build Operations Event and Offering screens.
- Build Storefront Event landing, Offering list, and Offering detail.
- Add unit, repository, API, route, authorization, and audit tests.

### Exit criteria

- An authorized operator can publish an Event and Offering.
- A public user can view only active and published data.
- Historical pricing/configuration does not silently follow later edits.

## Week 3 — Party, participant, quota, and Purchase

### Goal

Create a valid pending Common Purchase without collapsing financial and Qurban roles.

### Work items

- Implement reusable Party identity records.
- Model purchaser, payer, and intended Sohibul Qurban relationships explicitly.
- Implement canonical Purchase with `COMMON` channel.
- Snapshot Event, Offering, price, and participant-capacity data.
- Implement quota reservation/consumption policy.
- Add Purchase reference generation and idempotent creation.
- Build Storefront direct-checkout form and confirmation.
- Build Operations Purchase list and detail.
- Add validation, authorization, idempotency, and quota-contention tests.

### Exit criteria

- A valid direct checkout creates one traceable pending Purchase.
- Retrying the same request does not create a duplicate Purchase.
- Concurrent requests cannot consume more quota than configured.
- Purchaser, payer, and Sohibul Qurban can reference different people.

## Week 4 — Payment verification and activation

### Goal

Complete the authoritative commercial lifecycle.

### Work items

- Implement append-oriented Payment submission and evidence metadata.
- Store evidence through an approved storage adapter or safe MVP mechanism.
- Implement verification and rejection commands.
- Implement Purchase eligibility transition.
- Implement idempotent Sohibul Qurban activation.
- Add Operations payment-verification queue and detail.
- Add Storefront payment submission and status views.
- Record privileged actions in append-only audit history.
- Add duplicate-submission, verification, authorization, and activation tests.

### Exit criteria

- A submitted Payment can be verified or rejected by an authorized operator.
- A verified Payment makes the Purchase eligible under approved rules.
- Sohibul Qurban activates once and only once.
- Financial history and privileged decisions remain traceable.

## Week 5 — Storefront completion

### Goal

Deliver a usable public journey from discovery to tracking.

### Work items

- Complete mobile-first Event and Offering pages.
- Complete direct checkout and participant forms.
- Add explicit loading, validation, empty, error, stale, and conflict states.
- Complete payment instructions and evidence submission.
- Add Purchase tracking by safe reference and authentication policy.
- Add accessibility checks and responsive behavior.
- Add route, component, API-client, and critical-flow tests.
- Confirm the PWA does not cache sensitive API or financial data.

### Exit criteria

- A normal public user can complete the journey without operator assistance.
- Sensitive data is not cached or exposed through unsafe tracking.
- The flow is usable on representative mobile screen sizes.

## Week 6 — Operations completion and dashboard

### Goal

Deliver the minimum internal control plane needed to operate the MVP.

### Work items

- Complete operator login and route-access behavior.
- Complete Event and Offering administration.
- Complete Purchase search, filters, detail, and exception display.
- Complete payment-verification queue and decisions.
- Add Sohibul Qurban list and detail.
- Add dashboard metrics for Purchases, quota, payment backlog, eligibility, and activation.
- Use polling for dashboard freshness where needed.
- Display projection/query freshness and errors explicitly.
- Add Operations workflow and permission tests.

### Exit criteria

- Operators can run the full Common Purchase process without direct database access.
- The dashboard derives values from authoritative records.
- Frontend guards do not substitute for backend authorization.

## Week 7 — Integration, security, and resilience

### Goal

Prove the MVP under expected failures and contested operations.

### Work items

- Run complete backend, frontend, contract, and integration suites.
- Add the end-to-end Common Purchase scenario.
- Test duplicate Purchase and Payment requests.
- Test quota contention and stale-version conflicts.
- Verify public/internal API data separation.
- Review permission enforcement and audit completeness.
- Test backup and restore.
- Add structured operational metrics and alerts.
- Run baseline load tests for public reads, transactional writes, and dashboard polling.
- Fix critical and high-severity defects.

### Exit criteria

- Critical end-to-end, authorization, audit, idempotency, and concurrency tests pass.
- Backup restoration is demonstrated.
- Expected MVP load meets documented latency targets or known exceptions are approved.
- No known critical security or data-integrity defect remains.

## Week 8 — UAT, pilot, and release readiness

### Goal

Stabilize the system and prove that real users can operate it safely.

### Work items

- Deploy a release candidate to staging.
- Seed representative non-sensitive test data.
- Run purchaser, finance-operator, and operations-manager UAT.
- Rehearse Event publication, Purchase creation, payment verification, activation, tracking, and dashboard monitoring.
- Test rollback, restore, and incident procedures.
- Fix release-blocking defects only; do not add features.
- Prepare operator runbook, support escalation, release checklist, and known limitations.
- Conduct a controlled pilot and capture follow-up work.

### Exit criteria

- UAT acceptance criteria pass.
- Operators complete the journey using documented procedures.
- Release, rollback, backup, restore, and incident paths are verified.
- Deferred scope and known limitations are communicated explicitly.
- A go/no-go decision is recorded for the controlled MVP release.

## 6. Milestones

| Milestone                                          | Target        |
| -------------------------------------------------- | ------------- |
| MVP scope and architecture frozen                  | End of Week 1 |
| Public Event and Offering discovery                | End of Week 2 |
| Pending Common Purchase creation                   | End of Week 3 |
| Payment verification and Sohibul Qurban activation | End of Week 4 |
| Storefront journey feature-complete                | End of Week 5 |
| Operations workflow feature-complete               | End of Week 6 |
| Integrated release candidate                       | End of Week 7 |
| UAT-approved controlled MVP                        | End of Week 8 |

## 7. Definition of done

The MVP is complete only when:

- Event and Offering management works through Operations APIs and UI;
- public users can discover an active Event and published Offerings;
- Common Purchase preserves purchaser, payer, and Sohibul Qurban distinctions;
- quota is protected transactionally;
- Payment history is traceable and manual verification is permission-controlled;
- eligible Purchases activate Sohibul Qurban idempotently;
- Storefront tracking safely exposes current public status;
- Operations can manage the workflow without database access;
- audit records exist for privileged changes;
- public and Operations OpenAPI contracts match implemented behavior;
- critical automated tests and `make validate` pass;
- staging, backup, restore, logging, alerts, and release procedures are verified;
- controlled UAT passes with no unresolved critical defect.

## 8. Project-tracker fields

The Google Sheets tracker uses these fields:

| Field               | Purpose                                                                        |
| ------------------- | ------------------------------------------------------------------------------ |
| ID                  | Stable tracker identifier                                                      |
| Week                | Planned delivery week                                                          |
| Epic                | Owning vertical slice or cross-cutting capability                              |
| Task                | Concrete deliverable                                                           |
| Application         | API, Storefront, Operations, Infrastructure, or Cross-cutting                  |
| Category            | Discovery, Design, Backend, Frontend, Test, Security, DevOps, or Documentation |
| Priority            | P0, P1, or P2                                                                  |
| Estimate Hours      | Initial effort estimate                                                        |
| Status              | Backlog, Ready, In Progress, Blocked, Review, Done, or Deferred                |
| Dependency          | Required predecessor or decision                                               |
| Acceptance Criteria | Verifiable completion condition                                                |
| Actual Hours        | Logged effort                                                                  |
| Owner               | Responsible developer or reviewer                                              |
| Notes               | Risks, blockers, decisions, and scope changes                                  |

## 9. Delivery controls

- Only one primary implementation item should be in progress per developer.
- Agent tasks may run in parallel only when their files and contracts are independent.
- Every agent-produced change requires diff review and relevant verification.
- A task is `Done` only when its acceptance criterion is verified.
- Blocked time is recorded separately from engineering hours.
- Scope additions after Week 1 require an explicit tradeoff: defer another item or move the release date.
- Reforecast at the end of every week using actual hours, completed acceptance criteria, defects, and blockers.
- Do not commit or deploy generated business code without human review.
