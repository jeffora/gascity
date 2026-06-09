# Design Review Synthesis

## Overall Verdict: block

All ten persona syntheses resolve to `block`, so the global verdict is `block`
by worst-verdict-wins. The reviewers broadly approve the architectural
direction - session-owned mutation commands, durable-fact convergence,
surface-specific target adapters, and preflight gates - but they repeatedly
find that the design is not yet concrete enough to decompose beyond a
non-mutating Slice 0.

## Consensus Strengths
- Reviewers consistently support moving lifecycle and identity mutation
  ownership into `internal/session` while keeping API, CLI, reconciler, and
  runtime layers as callers or adapters rather than competing owners.
- Multiple personas praise the durable-scan-first recovery model: session
  events should be factual diagnostic signals, while safety-critical work
  release and convergence must be recoverable from persisted session state.
- The raw target-classifier versus per-surface adapter split is directionally
  correct, provided raw classification stays read-only and policy-free.
- The design's emphasis on source inventories, parity ledgers, guard fixtures,
  diagnostics manifests, and rollback evidence is the right control plane for
  a staged migration.

## Critical Findings

### [Blocker] Slice 0 Is Not A Schedulable, Machine-Checkable Gate
**Sources:** Elena Marchetti; Natasha Volkov; Takeshi Yamamoto; Sarah Chen; Ravi Krishnamurthy; Kwame Asante; Ingrid Holm; Claude, Codex, and DeepSeek/Gemini sources
**Issue:** The only safe next work is a non-mutating preflight, but "Slice 0"
is still ambiguous and not backed by checked-in artifacts, schemas, validators,
or failing fixtures. Required items such as source inventories,
`SCENARIO_PARITY.yaml`, vocabulary checkpoints, `DIAGNOSTICS_MANIFEST.yaml`,
slice-close validation, guard fixtures, and proof commands remain design
intent rather than an executable gate.
**Required change:** Add one authoritative Slice 0 backlog item with a fixed
artifact list, file paths, schemas, closure conditions, CI/pre-commit proof
commands, negative fixtures, and a rule that no behavior-moving implementation
slice can proceed until it passes.

### [Blocker] Mutation Ownership And One-Writer Proof Are Not Enforceable
**Sources:** Elena Marchetti; Takeshi Yamamoto; Sarah Chen; Ravi Krishnamurthy; Ingrid Holm
**Issue:** Production writers are still described through broad package or
endpoint-family rows, while current code still has direct lifecycle,
identity, repair, runtime identity, wake, hold, drain, close, and command
factory bypass paths. Exception classes such as doctor, repair, migration,
API/Huma, bridge, and command factory paths can become permanent self-labeled
bypasses. The unscheduled repair/backfill slice also creates an implicit flag
day.
**Required change:** Generate a per-path, per-symbol, per-key writer inventory
over all non-test, non-generated production Go under `cmd/` and `internal/`.
Replace broad exemptions with exact owner/expiry/retirement allowlist rows,
schedule or fold the repair/backfill work into a named slice, and enforce
worker-boundary and command-construction rules with static guards.

### [Blocker] Behavior Parity And Target Semantics Are Not Pinned To Live Proof
**Sources:** Natasha Volkov; Amara Diallo; Sarah Chen; Liam Okonkwo
**Issue:** The parity baseline is not trustworthy enough for decomposition:
the syntheses report stale or absent proof files, missing scenario rows, a
missing checked-in `SCENARIO_PARITY.yaml`, and broad source citations where
runnable characterization is required. Target resolution is also not one
universal resolver: API targets, package resolution, mail send/query, assignee
normalization, extmsg, CLI fallbacks, pool/path aliases, and historical aliases
have different precedence, repair, materialization, and error behavior.
**Required change:** Pin the implementation baseline on a clean checkout, land
`SCENARIO_PARITY.yaml` covering every `SESSION-*` row, and require exact proof
commands that fail on zero intended matches. Add per-surface target adapter
contracts plus collision fixtures for exact IDs, aliases, path aliases,
configured names, historical aliases, materialization, repair-on-read, and
error-class fallthroughs before moving callers.

### [Blocker] Atomicity, Close/Runtime Recovery, And Event Contracts Have Crash Windows
**Sources:** Takeshi Yamamoto; Amara Osei; Liam Okonkwo; Ingrid Holm
**Issue:** Runtime-start still permits a legacy no-token commit path and lacks
a durable way to recover/adopt/stop a started provider runtime after process
crash. Close and work-release ordering is not pinned tightly enough to prevent
crash-after-stop-before-commit resurrection or stranded work. Store write
safety is overstated without proven conditional writes or repair coverage, and
event diagnostics overclaim success/failure states the current recorder cannot
observe.
**Required change:** Materialize command-applier rows for runtime start, close,
wake, drain, retirement, and repair; include write primitive, fence, event
ordering, stale-success handling, and partial-state tests. Define a
close/work-release scanner contract with durable identity snapshots,
idempotency, supersession, completion markers, and crash tests. Replace event
diagnostic claims with truthful current states or assign an explicit
recorder/outbox migration.

### [Blocker] Decider, Fact, And Vocabulary Boundaries Can Still Become Facades
**Sources:** Liam Okonkwo; Kwame Asante; Amara Diallo; Ingrid Holm; Takeshi Yamamoto
**Issue:** Pure decider isolation is prose rather than a guarded file/package
set, and current lifecycle projection still has a zero-`Now` wall-clock
fallback if that code is enrolled as pure. Runtime intents, target candidates,
session fact events, diagnostics, and policy flags can still accumulate unused
future fields or caller policy before a production adapter needs them.
**Required change:** Add a machine-checkable boundary matrix for wake, hold,
sleep, drain, rollback, pool, health, progress, orphan release, and runtime
observation decisions. Split active next-slice vocabulary from provisional
future vocabulary, define fail-first checkpoint tests, name first production
callers for new shapes, and reject unused fields, flat optional envelopes, and
policy leakage into raw classifiers or session deciders.

### [Major] API, CLI, Dashboard, And Worker-Boundary Compatibility Need Route-Level Proof
**Sources:** Sarah Chen; Natasha Volkov; Amara Diallo
**Issue:** API and CLI paths still read raw lifecycle metadata and broad Huma
or legacy route rows do not prove response, error, JSON, exit-code,
fallback-reason, dashboard/SSE, or generated-client compatibility. Legacy
`/v0/session*` and `/v0/sessions` behavior is not clearly supported,
delegated, retired, or excluded.
**Required change:** Inventory session-affecting routes and commands at
handler/command granularity, including current mutators, response/output
contracts, owner slice, exceptions, expiry, and parity fixtures. Add CLI
API-route versus local-fallback requirements, generated-client wrapper rules,
and dashboard/OpenAPI impact checks where touched.

### [Major] Diagnostics And Performance Budgets Are Not Executable
**Sources:** Ingrid Holm; Liam Okonkwo; Amara Osei
**Issue:** The diagnostics vocabulary, doctor/session-inspect/trace renderer
ownership, event recovery states, wake-cause/blocker details, and performance
budgets are spread across prose. The budget substrate does not yet define
counting stores, backend behavior, fake subscribers, fake runtime sources,
large-city cardinalities, subprocess-count semantics, or non-flaky cap tests.
**Required change:** Make `DIAGNOSTICS_MANIFEST.yaml` the normative source for
operation/check IDs, owners, trace outcomes, wake causes, blockers, fact
fields, renderer surfaces, tests, redaction keys, event relationships, and
cost classes. Add golden renderer tests and measured query/subprocess budget
proof before diagnostics or performance claims become slice gates.

### [Minor] Artifact And Source Labeling Needs Cleanup
**Sources:** Amara Diallo; Ravi Krishnamurthy; local synthesis inspection
**Issue:** The ten persona-synthesis beads are stamped with `gc.attempt=12`,
but their `design_review.output_path` values point under
`attempt-1/persona-syntheses/`; the actual `attempt-12` directory has raw
review artifacts but no `persona-syntheses/` directory. Several syntheses also
refer to the third source as DeepSeek while the artifact filenames use
`_gemini.md`.
**Required change:** Fix the review workflow output-path and model-label
contract so attempt artifacts are written under the active attempt directory
and source labels are consistent. This synthesis used the ten output paths
stamped on the closed persona-synthesis beads because those are the available
required artifacts for this step.

## Disagreements
- Several model-level reviews were `approve-with-risks`, especially for the
  event-delivery, reconciler, API/CLI, and YAGNI lanes. The persona syntheses
  reasonably upgrade these to `block` because the unresolved issues are
  decomposition prerequisites, not implementation polish.
- Reviewers differ on mechanisms: conditional store primitives versus tokened
  blind writes with repair tests, concrete `DiagnosticReport` types versus
  kind-specific diagnostic payloads, and flat candidate structs versus
  per-kind/private target shapes. The design can choose the simplest local
  mechanism, but it must be machine-checkable and tied to first production use.
- Durable scans are widely accepted as recovery authority, while event
  diagnostics and `SessionFactEvent` vocabulary are disputed. The safe middle
  ground is to keep events thin and factual today, remove generic committed
  facts, and introduce richer payloads only when a typed subscriber needs exact
  fields.
- Some source-specific findings cite attempt-specific line numbers or files
  that may have drifted. Treat those as prompts to re-audit against the pinned
  baseline rather than as dismissals.

## Missing Evidence
- Checked-in Slice 0 artifact manifest, schemas, proof commands, CI/pre-commit
  integration, and negative fixtures.
- Source-complete writer and raw-read inventories for session-owned lifecycle,
  identity, runtime, repair, wake, hold, drain, close, event, diagnostics, and
  API/CLI surfaces.
- `SCENARIO_PARITY.yaml`, row statuses, owner-approved amendment artifacts,
  and exact freshness tests covering every `SESSION-*` scenario row.
- Machine-checkable command-applier ledger and backend write-primitive matrix
  for conditionality, partial writes, stale tokens, duplicate commands, skipped
  events, provider-side success with commit failure, and repair paths.
- Close/work-release and drain recovery contracts with durable identity
  snapshots, trigger predicates, query shape, cadence, idempotency,
  supersession, completion markers, restart behavior, and crash-window tests.
- Per-surface target-classification compatibility matrix and collision
  fixtures for API, package resolver, mail, assignee normalization, extmsg,
  CLI, nudge, attach, inspect/log/transcript, pool scheduling, and path aliases.
- Pure-decider enrollment guard, zero-`Now` migration evidence, reconciler
  boundary matrix, runtime-observation freshness/supersession rules, and
  replacement scale/health/progress proofs.
- `DIAGNOSTICS_MANIFEST.yaml`, trace/doctor/inspect renderer ownership,
  canonical wake-cause/blocker vocabulary, and query/subprocess performance
  budget substrate.

## Recommended Changes
1. Freeze behavior-moving decomposition and make one authoritative Slice 0
   backlog item with artifact paths, schemas, validators, proof commands, and
   failing fixtures.
2. Pin the implementation baseline on a clean checkout, then generate the
   source inventories, scenario parity file, vocabulary checkpoints,
   diagnostics manifest, and slice-close validator from that baseline.
3. Convert broad mutation and API rows into exact path/symbol/key/surface rows,
   including repair/backfill ownership, worker-boundary exceptions, expiry,
   rollback proof, and one-writer bake rules.
4. Define command atomicity and recovery contracts for runtime start, close,
   drain, wake, identity retirement, and work release before any caller
   delegates to those commands.
5. Split target classification into pure candidate collection plus
   per-surface adapters, and land the same-token collision matrix with
   characterization tests before adopting callers.
6. Add pure-decider, active-vocabulary, and YAGNI guards that reject direct
   clocks, store/runtime/config access, unused fields, provisional vocabulary
   in active code, and raw-classifier policy leakage.
7. Make diagnostics and performance proof executable through a normative
   manifest, renderer golden tests, truthful event status vocabulary, and
   measured query/subprocess budgets.
