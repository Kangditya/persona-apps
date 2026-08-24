# Task: Replan Full Eid Event-Day MVP Artifacts and Tracker

## Executed

## Objective

Replace the remaining commerce-only MVP plan with an evidence-backed roadmap
that delivers a complete 3–4-day Eid al-Adha event operation from public
purchase through livestock intake, allocation, slaughter, distribution,
customer status, multi-team coordination, realtime visibility, mobile field
work, degraded connectivity, and release rehearsal.

## Approval and Boundaries

- The user explicitly approved changes to planning and documentation artifacts.
- Do not change application source, migrations, runtime OpenAPI behavior,
  generated clients, infrastructure, dependencies, or production data.
- Preserve completed tracker rows W1-01 through W3-06 and their archives.
- Preserve the untracked `.codex/config.toml` file.
- Do not commit or push.

## Source of Truth

- `docs/PRD.md`;
- `docs/PRODUCT_MAP.md`;
- `docs/ARCHITECTURE.md`;
- `docs/DECISIONS.md`;
- `docs/CONVENTIONS.md`;
- `.codex/CURRENT_STATE.md`;
- current source, contracts, migrations, completed task archives, and the live
  `Qurban MVP Project Tracker` spreadsheet.

## Required Planning Decisions

The revised artifacts must establish these baselines:

1. An Event has exactly 3 or 4 inclusive local execution days in an explicit
   IANA timezone; sessions, shifts, handovers, and recovery remain event-scoped.
2. Field teams, memberships, shifts, station assignments, readiness, incidents,
   and support escalation are first-class event-scoped operational records.
3. Sohibul Qurban attendance and personal/proxy slaughter are configurable and
   never inferred from purchase eligibility.
4. Distribution supports explicit Sohibul entitlement and beneficiary records,
   with pickup or delivery and traceable proof/completion.
5. The Go API and PostgreSQL remain authoritative. Operations uses polling
   first and SSE for valuable one-way event updates; WebSocket remains deferred
   without a proven bidirectional need.
6. The responsive web PWAs are the mobile baseline. Only allowlisted,
   non-financial field milestones may queue during degraded connectivity;
   configuration, payment, authorization, and sensitive mutations remain
   online-only.
7. Customer event-day views expose only purchase-token-scoped schedule,
   attendance, slaughter, distribution, notification, and document status.

## Deliverables

- Align applicable canonical product, architecture, decision, security,
  lifecycle, database-planning, README, current-state, and roadmap artifacts.
- Replace the obsolete two-month roadmap with a 16-week roadmap.
- Replan every unfinished tracker task while preserving native formulas,
  validation, formatting, completed evidence, and blank Actual Hours.
- Keep remaining `.codex/plans` deliberately header-only until one task is
  revalidated and activated; add/replace headers to match the tracker exactly.
- Update `Weekly Plan`, `Scope`, and `Dashboard` to the same task boundary.

## Acceptance Criteria

- Every requirement-coverage gap maps to one or more remaining tracker tasks
  and a canonical artifact statement.
- All task IDs are unique; every dependency resolves to an existing task and
  precedes its dependent task.
- Completed W1-01 through W3-06 rows, notes, status, and plan/archive files are
  unchanged.
- Every unfinished tracker row has a valid nonblank status, acceptance
  criterion, estimate, and matching header-only plan file.
- Weekly totals and dashboard formulas cover the final tracker row and all 16
  weeks; Actual Hours remain evidence-only.
- Scope no longer defers livestock, allocation, slaughter, distribution,
  polling/SSE, mobile field support, or bounded degraded-connectivity work.
- No application code, migration, runtime contract, dependency, or secret is
  changed.

## Verification

- Review the complete Git diff and `git diff --check`.
- Compare completed task files and tracker rows with their pre-change evidence.
- Validate Markdown formatting and repository task/header coverage.
- Re-read all written Google Sheet ranges with values, formulas, validation,
  and formatting metadata.
- Visually verify edited Sheet tabs or perform the strongest available native
  formatting inspection.
- Finish with implemented, planned, assumed, deferred, risk, and next-task
  distinctions.

## Final Review

### Implemented by this planning task

- Canonical and supporting artifacts now fix the Full Event-Day MVP requirement
  baseline and no longer defer or reopen its execution-day, team, attendance,
  distribution, realtime/mobile, or degraded-connectivity boundaries.
- The former eight-week document is superseded by the 16-week roadmap.
- Every unfinished tracker row was replanned and every unfinished row has a
  matching header-only draft.
- Tracker `Weekly Plan`, `Scope`, and `Dashboard` were synchronized and
  visually repaired without fabricating Actual Hours.

### Verified

- 127 unique tracker tasks through W16-09; 723 planned hours; 21 Done, 2 Ready,
  104 Backlog.
- Every dependency resolves to an earlier task.
- All 106 unfinished tracker rows have required fields, blank Actual Hours,
  strict status dropdown validation, and matching task headers.
- Sheet formulas cover Tracker rows 2–128 and Weeks 1–16; visual inspection
  confirmed readable Tracker, Weekly Plan, Scope, and Dashboard layouts.
- Prettier reports all matched files formatted; `git diff --check` passes.
- No application source, runtime contract, migration SQL, dependency, generated
  output, secret, or the existing untracked `.codex/config.toml` was changed.

### Assumed

- Estimates remain planning estimates for one-developer sequencing and require
  weekly reforecast from measured evidence.
- Exact supported device/browser versions will be selected and verified in
  W14-01 rather than invented during planning.

### Deferred

- All application/runtime implementation represented by W3-07 through W16-09.
- Saving, Giveaway, payment gateway automation, advanced financial adjustments,
  multiple-Offering checkout, automatic identity deduplication, native mobile,
  WebSocket, broker/Redis/microservices, route optimization, and advanced
  document/analytics platforms.

### Remaining risk and next task

- Repository `make format-check` still reports 18 pre-existing Go files that do
  not match the repository formatter. They were not touched because this task
  explicitly prohibited code changes.
- Activate and execute W3-07 next after revalidating its header-only draft
  against current source and copying one reviewed task into `.codex/TASK.md`.
