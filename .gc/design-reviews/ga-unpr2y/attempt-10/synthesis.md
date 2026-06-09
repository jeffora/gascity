# Design Review Synthesis

## Overall Verdict: block

Worst-verdict-wins produces `block`: all ten persona syntheses returned
`block`. The reviewers continue to endorse the broad direction, especially a
non-mutating Slice 0, durable-scan recovery, adapter-owned target policy, and
vocabulary restraint, but the design is not yet executable enough to
decompose safely beyond Slice 0.

## Consensus Strengths
- Multiple personas agree that a non-mutating Slice 0 preflight is the right
  first schedulable unit if it physically lands the inventory, parity,
  vocabulary, guard, allowlist, diagnostics, baseline, and proof artifacts.
- Reviewers repeatedly praised durable facts and controller scans as the
  correctness backstop, with event delivery and synchronous cascades treated
  as latency or diagnostic aids rather than the only recovery path.
- The raw target-classifier direction is viable if it stays policy-free and
  every API, CLI, mail, extmsg, assignee, pool, and worker-boundary surface
  keeps its own compatibility adapter and fixtures.
- The session/reconciler separation is directionally useful when session
  deciders receive immutable facts only and controller-owned demand, budgets,
  provider health, progress, scaling, and scheduling remain outside
  `internal/session`.
- The design's YAGNI posture is improving: several reviewers accepted
  provisional future vocabulary when it is kept out of active code until a
  first production caller proves exact field demand.

## Critical Findings

### [Blocker] Slice 0 Is Not Yet A Schedulable Executable Gate
**Sources:** Elena Marchetti, Natasha Volkov, Takeshi Yamamoto, Sarah Chen,
Liam Okonkwo, Ravi Krishnamurthy, Kwame Asante, Ingrid Holm
**Issue:** The design keeps treating missing Slice 0 artifacts as prose
conditions rather than committed, fail-closed gates. Required artifacts such as
`BOUNDARY_INVENTORY.md`, `SESSION_BOUNDARY_SYMBOLS.yaml`,
`SCENARIO_PARITY.yaml`, vocabulary checkpoints, guard fixtures, shrink-only
allowlists, baseline proof, diagnostics manifests, and freshness tests either
do not exist or are not yet mechanically defined. The "Slice 0" label is also
ambiguous between a standing traceability row and the only implementation
deliverable allowed to proceed.
**Required change:** Make Slice 0 one explicit backlog item with one artifact
list, one closure condition, and one proof command. No behavior-moving bead
should be generated until Slice 0 physically lands and fails correctly on
missing artifacts, stale paths, zero matched tests, unowned writers, expired
exceptions, undeclared vocabulary, and absent parity rows.

### [Blocker] Baseline And Parity Proof Are Not Trustworthy
**Sources:** Natasha Volkov, Elena Marchetti, Sarah Chen, Liam Okonkwo
**Issue:** The parity lane has no pinned clean baseline. Reviewers found the
active checkout can be hundreds of commits behind upstream, historical
requirement commits are not current proof, some cited files are missing in the
active checkout but exist upstream, and scenario ownership diverges between
tables. Source paths and commit hashes are still being used where current
runnable characterization tests, exact output contracts, or static assertions
are required.
**Required change:** Pin the clean implementation baseline ref and make
`SCENARIO_PARITY.yaml` the canonical source consumed by bead generation. Add
freshness checks for all `SESSION-*` rows that fail on non-ancestor commit
citations, missing proof paths, zero matched tests, source-only current proof,
row-set drift, missing scenario-ID markers, and behavior-ledger amendments
without owner approval.

### [Blocker] Mutation Boundary Guard Coverage Is Source-Incomplete
**Sources:** Elena Marchetti, Sarah Chen, Ravi Krishnamurthy, Ingrid Holm
**Issue:** The proposed guard can miss production mutation paths that do not
look like direct `internal/session` imports: generic bead-store bridges,
Huma/API routes, manager methods, package mutators, dynamic metadata batches,
patch-map extension/application, helper-returned patches, repair helpers,
wrapper/interface receivers, and subprocess `bd` mutation routes. Broad
doctor, repair, migration, and API/Huma exclusions remain self-labeling
escape hatches.
**Required change:** Scan all production Go files under `cmd/` and
`internal/`, or require explicit proven-unreachable rows for exclusions.
Replace package-level and endpoint-family rows with exact path/symbol/key
inventory rows, add target-bead classification proof, and enforce expiring
allowlist rows with owner, reason, expiry, source hit, design anchor,
retirement condition, and guard fixtures for dynamic and generic bypasses.

### [Blocker] Command Atomicity And One-Writer Migration Are Unspecified
**Sources:** Takeshi Yamamoto, Ravi Krishnamurthy, Elena Marchetti, Sarah Chen
**Issue:** Runtime-start, close, wake, drain, identity retirement,
`session_key`, `instance_token`, wake/hold/drain markers, and API command
routes still lack a concrete command-applier contract. The design references
lock-plus-reread, tokened commits, and repair convergence without naming the
cross-process primitive, write keys, entry points, partial-write behavior, or
stale-success handling. It also requires one writer during bake while leaving
legacy writers alive in overlapping slices and unscheduled repair paths.
**Required change:** Materialize command-applier rows before delegation. Each
command family needs preconditions, validation points, write primitive,
fence/token fields, stale-success semantics, provider side-effect ordering,
partial-state matrices, repair authority, race tests, and per-key ownership
with allowed legacy writers and expiry. Either prove a real conditional store
primitive or classify blind writes as detect-and-converge with tests for every
visible partial state.

### [Blocker] Close, Work Release, And Event Recovery Are Not Durable Enough
**Sources:** Amara Osei, Takeshi Yamamoto, Ingrid Holm
**Issue:** The durable-scan-first recovery model is sound, but the close and
work-release contract is still not executable. The design does not pin the
scanner/helper owner, trigger predicate, query shape, store coverage, cadence,
boundedness, completion marker, duplicate behavior, partial-query behavior,
restart behavior, stale-successor suppression, or release identity schema. It
also risks using event terminology such as missed-event recovery without
durable delivery evidence.
**Required change:** Define a concrete close/work-release scanner contract.
Write or preserve release identity before identity clearing, including bead ID,
session name or alias, configured identity, generation or `instance_token`,
store ref, idempotency key, and supersession key. Add crash-after-close-commit,
skipped-event, duplicate-scan, partial-query, partial-release, restart,
zero-match, stale-successor, and terminal-completion tests. Keep runtime
intents as command results, not event payloads.

### [Blocker] Target Classification Still Collapses Surface Policy
**Sources:** Amara Diallo, Kwame Asante, Sarah Chen, Natasha Volkov
**Issue:** The classifier boundary conflicts with live behavior for
`RepairEmptyType`, historical aliases, CLI qualified-alias basename fallback,
mail send/query, beadmail routing, assignee materialization, rejected-by-config
post-filtering, exact non-session bead IDs, dual alias/session-name demotion,
and template/config targets. Raw candidate facts can also become accidental
selection authority if callers rank `candidates[]` instead of using
adapter-owned `selected`.
**Required change:** Keep the raw classifier policy-free and diagnostic unless
a single first production surface proves exact demand. Add per-surface
compatibility tables and fixtures for API/Huma, CLI, mail, extmsg, assignee,
pool/orphan routing, nudge, attach, inspect, log, transcript, generated-client
impact, local fallback, materialization, repair, historical aliases, closed
targets, fallthrough errors, and output/wire shapes. Enforce by type or tests
that adapter `selected` is the only action authority.

### [Blocker] Session/Reconciler Boundary Still Leaks Controller Policy
**Sources:** Liam Okonkwo, Takeshi Yamamoto, Ingrid Holm
**Issue:** The design is not decision-by-decision specific enough about what
moves into `internal/session` versus what stays in `cmd/gc`. `ComputeAwakeSet`,
`AwakeInput`, runtime intents, provider health, progress, circuits, min-active
coverage, pool demand, work aggregation, create budgets, wave ordering,
partial observations, and destructive actions can still be read as session
decider responsibilities. `ProjectLifecycle` still has a zero-`Now`
`time.Now().UTC()` fallback if enrolled in the pure-decider set.
**Required change:** Add a boundary matrix for wake eligibility, holds, pins,
sleep, ready wait, drain, stale creating rollback, pool scaling, health,
progress, orphan release, and runtime observation. Each row must name session
inputs, session output, adapter-owned facts, reconciler-owned scheduling or
budget policy, forbidden facts, and proof tests. Define the pure-decider file
set and guard, remove or explicitly migrate zero-`Now` fallbacks, and restore
or replace health/progress/scale proof files before those slices move.

### [Blocker] Repair Is A Live Mutation Without An Owner
**Sources:** Ingrid Holm, Amara Diallo, Elena Marchetti, Natasha Volkov
**Issue:** `RepairEmptyType` is a read-path repair mutation that can discard
persistence errors and still normalize in-memory state. The design says repair
failures must not be hidden, but it lacks a concrete repair slice, audited
helper, production caller inventory, trace evidence, doctor rendering, and
guard coverage. Excluding doctor/repair/migration paths from the guard
contradicts the design's own rejection of broad exclusions.
**Required change:** Assign repair to a real slice or audited helper before
target-classifier and resolver adoption. Enumerate every production
`RepairEmptyType` caller, scan repair/doctor/migration paths with expiring
repair-only allowlist rows, propagate persistence errors, avoid in-memory
repair when persistence fails, and emit typed before/after trace plus doctor or
inspect evidence.

### [Blocker] Operability And Performance Gates Are Not Machine-Checkable
**Sources:** Ingrid Holm, Amara Osei, Sarah Chen, Liam Okonkwo
**Issue:** The design promises trace, doctor, inspect, API, CLI, event,
diagnostic, performance, and fan-out behavior without enough concrete artifact
paths, centralized outcome codes, renderer/filter tests, query-count budgets,
backend index proof, subscriber backpressure behavior, or large-city
baselines. Eligibility diagnostics also collapse wake causes and blockers into
too-broad categories or free text.
**Required change:** Add a diagnostics manifest, trace outcome compatibility
table, doctor/session-inspect registry, trace/API/CLI rendering fixtures,
central wake-cause and blocker codes, query-count or subprocess-count budgets,
metadata-index proof or measured backend baselines, subscriber backpressure
tests, durable-scan diagnostics, reconciler-write budgets, session-scoped
trace/inspect output, and large-city benchmarks for lookup, mail, scans,
runtime repair, pool snapshots, and event fan-out.

### [Blocker] Vocabulary And YAGNI Controls Still Permit Premature Shapes
**Sources:** Kwame Asante, Liam Okonkwo, Amara Diallo
**Issue:** Future-slice vocabulary such as `SessionCommandConflict`,
`RuntimeStartIntent`, `SessionFactEvent`, generic committed facts, policy
flags, and flat target candidate/result envelopes can still leak into active
implementation before a first production caller exists. `TestVocabularyCheckpoints`
is not yet defined to reject provisional future rows, unused fields, or fields
included only for later surfaces.
**Required change:** Split active next-slice vocabulary from provisional design
bounds. Seed existing-contract rows in `VOCABULARY_CHECKPOINTS.yaml`, but
require first delegated production caller evidence, exact demanded fields,
non-goals, package visibility, and tests before any new type, field, policy
flag, event payload, package skeleton, or diagnostic vocabulary lands. Keep
future events provisional until a typed subscriber proves exact fields.

### [Major] Rollback And Coexistence Need Data-Direction Proof
**Sources:** Ravi Krishnamurthy, Sarah Chen, Takeshi Yamamoto
**Issue:** The migration plan has useful `session_design.*` metadata direction,
but no required finalizer, bead schema check, or test proves slice-close
completeness. Rollback mostly restores routes; it does not prove old code
tolerates new metadata or new code tolerates old metadata across
`instance_token`, `pending_create_*`, `session_key`, close/retire facts,
wake/drain markers, repair metadata, and partial adoption. API routes remain
ambiguous between `worker.Handle`, session command factory, direct manager
calls, generated-client paths, and legacy no-city endpoints.
**Required change:** Require every slice to carry converted caller lists,
allowed legacy rows, retired rows, guard deltas, one-writer proof,
data-direction rollback proof, proof commands, and exact API/CLI route
inventory. Add a slice-close validator or test for required
`session_design.*` metadata and update the worker-boundary ledger whenever a
shared API/CLI exception is created or retired.

### [Minor] Review Artifacts Are Still Written Outside The Attempt Directory
**Sources:** Workflow artifact inspection during synthesis
**Issue:** The attempt-10 persona synthesis directory is empty, while all ten
closed persona synthesis beads stamped fresh outputs under
`.gc/design-reviews/ga-unpr2y/attempt-1/persona-syntheses/`. This synthesis
consumed those bead-stamped source paths, whose mtimes matched the current
review wave, and wrote the global report to attempt 10.
**Required change:** Fix child synthesis attempt selection or output-path
stamping so future global synthesis steps can read the current
`$ATTEMPT_DIR/persona-syntheses/` directly without source-path recovery.

## Disagreements
- Several raw model reviews returned `approve-with-risks`, especially for the
  conceptual architecture, while every persona synthesis chose `block`. My
  assessment agrees with the persona verdicts: the architecture direction can
  proceed only as Slice 0, not as behavior-moving decomposition.
- Reviewers differed on whether missing Slice 0 artifacts are a design blocker
  or merely an implementation precondition. The synthesis treats them as a
  design blocker for any work beyond Slice 0 because decomposition cannot cite
  inventory IDs, parity rows, guard rules, or proof commands that do not exist.
- Baseline evidence remains disputed across active checkout, upstream state,
  and historical commits. The design must pin a clean baseline and re-audit
  there before classifying missing proof as restored, stale, unavailable, or
  owner-amended.
- Reviewers proposed different close/work-release mechanisms: pending-release
  flags, transitional close facts, release-key snapshots, tokened assignments,
  and stale-successor checks. Any mechanism is acceptable only if it is durable
  after restart and tested against partial release and successor reuse.
- API routing reviewers disagree on whether `worker.Handle` should be the
  universal end route or whether a narrow session command factory is acceptable.
  The design must choose per surface, document exceptions with owner and
  expiry, and keep the worker-boundary ledger authoritative.
- Event reviewers accept durable-scan recovery, but disagree on names such as
  `missed-event-recovered`. The synthesis requires truthful durable
  convergence terminology unless durable event-delivery evidence is added.
- Target vocabulary reviewers differ on whether candidate facts need per-kind
  structs or can remain tagged/private shapes. The required property is
  mechanical prevention of broad optional envelopes and raw-candidate action
  authority.
- Performance reviewers focus on different layers: read query/index cost,
  write-side marker cost, subscriber fan-out, and pool snapshot rebuilds. The
  common requirement is measurable budgets or baselines across all hot paths.

## Missing Evidence
- Committed Slice 0 artifacts: source-complete writer inventory,
  `SESSION_BOUNDARY_SYMBOLS.yaml`, `SCENARIO_PARITY.yaml`,
  `VOCABULARY_CHECKPOINTS.yaml`, static guard implementation and fixtures,
  shrink-only allowlist, baseline proof, diagnostics manifest, and
  artifact-freshness tests.
- Pinned clean baseline ref, upstream ancestry checks, full 45-row
  requirements parity, current oracle commands, scenario-ID markers, and
  owner-approved amendment evidence for behavior-ledger changes.
- Machine-checkable command-applier ledger for runtime-start, close, wake,
  drain, identity retirement, token backfill, runtime side effects,
  stale-success handling, partial batches, and backend primitive support.
- Durable close/work-release scanner contract with release identity schema,
  query shape, owner, cadence, boundedness, completion fact, idempotency key,
  supersession key, restart behavior, stale-successor tests, and partial-query
  semantics.
- Per-surface target/API/CLI/mail/extmsg/assignee matrices covering resolver
  order, historical aliases, closed targets, materialization, repair,
  fallthrough errors, recipient sets, direct beadmail routing, stdout/stderr,
  JSON, status codes, Huma problem bodies, generated-client effects, and
  legacy no-city endpoints.
- Pure-decider guard, `ProjectLifecycle` zero-`Now` policy, `AwakeInput`
  disposition rows, forbidden operational facts, runtime-intent schema,
  health/progress/scale replacement tests, destructive-action matrix, and
  W-013 freshness/supersession rules.
- Repair owner for `RepairEmptyType`, production caller inventory, guard
  coverage for repair/doctor/migration paths, persistence-error propagation
  tests, before/after trace evidence, and doctor or inspect rendering.
- Diagnostics and operability artifacts: trace outcome compatibility table,
  doctor/session-inspect registry, session-scoped trace or inspect surface,
  canonical wake-cause and blocker codes, renderer/filter tests, API/CLI
  fixtures, and centralized outcome names.
- Performance and fan-out evidence: query-count or subprocess-count budgets,
  metadata-index proof or measured backend behavior, subscriber backpressure
  tests, durable-scan retry/cap rules, reconciler-write budget, large-city
  baselines, and pool snapshot rebuild proof.
- Vocabulary checkpoint tests that reject missing rows, undeclared fields,
  provisional future rows in active code, unused fields, fields without first
  production caller evidence, and universal optional envelopes.
- Enforceable migration metadata: converted callers, allowed legacy writers,
  retired rows, guard changes, one-writer proof, rollback files,
  data-direction convergence tests, and slice-close validator or finalizer.

## Recommended Changes
1. Approve only a non-mutating Slice 0 next; make it one backlog item with
   committed inventory, parity, vocabulary, guard, allowlist, diagnostics,
   baseline, and proof artifacts.
2. Pin the clean audit baseline and regenerate parity from that baseline,
   making `SCENARIO_PARITY.yaml` the bead-generation source for every
   `SESSION-*` row.
3. Expand the mutation guard to scan source-complete production paths and catch
   generic stores, Huma/API bridges, manager methods, patch maps, package
   mutators, repair helpers, and subprocess mutation routes.
4. Rewrite lifecycle command contracts around actual store primitives,
   including fences, stale-success handling, partial-state repair, runtime
   side-effect ordering, and one-writer ownership per key family.
5. Define durable close/work-release recovery with preserved release identity,
   scanner ownership, bounded queries, completion semantics, idempotency,
   supersession, stale-successor proof, and crash-window tests.
6. Keep target classification policy-free and adopt one production surface at
   a time with exact per-surface resolver, repair, materialization, output, and
   generated-client fixtures.
7. Split session eligibility from reconciler policy with field-level
   disposition rows, pure-decider guards, runtime observation facts, restored
   health/progress/scale proof, and explicit forbidden facts.
8. Assign repair, diagnostics, and performance to concrete slices or Slice 0
   artifacts, including `RepairEmptyType`, trace/doctor/session-inspect,
   canonical outcome codes, query budgets, subscriber caps, and large-city
   baselines.
9. Constrain vocabulary with active/provisional checkpoints, first-caller
   evidence, exact demanded fields, non-goals, package visibility, and tests
   that reject unused or future-only shapes.
10. Add a slice-close validator or workflow gate for `session_design.*`
    metadata, data-direction rollback proof, one-writer evidence, converted
    caller lists, and worker-boundary ledger updates.
