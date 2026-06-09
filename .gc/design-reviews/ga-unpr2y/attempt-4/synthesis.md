# Design Review Synthesis

## Overall Verdict: block

Nine persona syntheses returned `block`; only the YAGNI/contract-scope lane returned `approve-with-risks`, so worst-verdict-wins makes the global verdict `block`. The design is moving in the right direction around session-owned commands, durable-fact recovery, typed boundaries, and slice-scoped vocabulary, but reviewers still found implementation-blocking gaps in mutation ownership, behavior parity, atomicity, event recovery, target classification, worker/API routing, reconciler facts, and operability proof. The document should not be decomposed into implementation beads until those contracts are made concrete and testable.

## Consensus Strengths
- Multiple personas praised the direction of moving session lifecycle mutation behind session-owned command/decider contracts instead of letting API, CLI, reconciler, and helper paths mutate raw metadata directly.
- Reviewers saw real improvement in the target-classification inventory, slice vocabulary checkpoints, compatibility-chain thinking, and the rule that future shared types must be justified by actual callers.
- The durable-fact convergence model is a sound recovery direction when it is explicitly authoritative for critical reactions such as work release, identity retirement, drain recovery, and wake recovery.
- Several personas accepted the high-level separation between session eligibility, controller/reconciler demand policy, runtime providers, and user-facing projections, provided the plan redraws those boundaries as owned fact contracts rather than broad facades.
- The trace, doctor, event, parity, and performance sections contain useful scaffolding; reviewers mostly objected that the scaffolding is not yet precise enough to gate implementation.

## Critical Findings

### [Blocker] Mutation Ownership And Static Guards Are Still Not Enforceable
**Sources:** Elena Marchetti; Natasha Volkov; Takeshi Yamamoto; Sarah Chen; Ravi Krishnamurthy; Ingrid Holm; Claude, Codex, DeepSeek/Gemini

**Issue:** The mutation landscape is not source-complete and the static guard is not mechanically specified for the hard cases. Persona syntheses identify missing or under-specified writers across `cmd/gc/session_reconcile.go`, `cmd/gc/session_reconciler.go`, `cmd/gc/session_name_lookup.go`, `cmd/gc/session_circuit_breaker.go`, `cmd/gc/soft_reload.go`, `cmd/gc/cmd_wait.go`, `cmd/gc/cmd_stop.go`, `cmd/gc/cmd_nudge.go`, `cmd/gc/adoption_barrier.go`, `cmd/gc/cmd_bd_store_bridge.go`, `cmd/gc/session_lifecycle_parallel.go`, `cmd/gc/session_beads.go`, `cmd_session_wake.go`, `internal/api/session_resolution.go`, `internal/api/handler_session_create.go`, and repair paths such as `RepairEmptyType`. The guard language still does not define session-bead discrimination, dynamic metadata batch handling, generic bridge policy, patch-helper containment, shrink-only allowlist enforcement, or API/worker-boundary bypass coverage.

**Required change:** Replace the current mutation landscape with a canonical call-site inventory generated or verified from source. For every writer, record stable ID, operation, key family, target bead-type proof, current owner, intended command owner, exception status, retirement condition, and tests. Define a build-failing guard that covers raw store writes, patch construction/application, local wrappers, manager/package-level command bypasses, dynamic maps, and new unmapped resolver or lifecycle call sites.

### [Blocker] Behavior Parity And Target Classification Can Still Drift User-Facing Semantics
**Sources:** Natasha Volkov; Amara Diallo; Sarah Chen; Kwame Asante; Claude, Codex, DeepSeek/Gemini

**Issue:** The design still lacks reverse traceability from every `REQUIREMENTS.md` scenario row to a slice, preserved invariant, or explicit out-of-scope decision. Reviewers found unmapped or under-owned behavior for state transitions, suspend/materialization, wake-vs-hold/quarantine, path alias, CLI qualified-alias basename fallback, exact-ID closed behavior, `rejected-by-config`, assignee normalization, mail, extmsg, transcript/log lookup, and API/CLI response compatibility. Some requirements still cite deleted proof files, so the parity ledger is not a reliable implementation oracle.

**Required change:** Add a row-by-row parity matrix for every touched scenario and every adopted target surface. Bind API, CLI, mail, extmsg, assignee, nudge, attach, inspect/log/transcript, and materialization behavior to concrete compatibility chains, typed outcomes, and exact tests for status, body, error shape, request ID, stdout/stderr, JSON, and exit code. Restore or replace stale proof files before using the ledger as a gate.

### [Blocker] Command Atomicity, Cross-Process Concurrency, And Runtime-Start Recovery Are Underspecified
**Sources:** Takeshi Yamamoto; Ravi Krishnamurthy; Elena Marchetti; Ingrid Holm; Claude, Codex, DeepSeek

**Issue:** The proposed re-read-before-write model is not an atomic commit fence, and the design does not choose a concrete writer topology or store-level concurrency mechanism. Multi-key lifecycle operations may use `SetMetadataBatch`, `Update`, `Close`, or multiple writes without proven atomicity, bounded partial-state ordering, commit markers, or repair rows. Runtime start remains especially risky: provider side effects can succeed before active/session-key metadata, event emission, or recovery correlation is durably complete.

**Required change:** Pick one enforceable model per command cluster: single-writer controller routing, store-level conditional mutation, transaction semantics, commit markers with deterministic repair, or another concrete alternative. For wake, close, drain, stale-resume cleanup, runtime start, identity retirement, and repair writes, state facts re-read, preconditions, lock or transaction scope, conflict result, partial-write recovery, event ordering, runtime compensation, and race tests. Make `instance_token` the authoritative runtime identity and define provider-start-success plus durable-commit-failure recovery.

### [Blocker] Event Delivery And Work Release Need A Durable Recovery Contract
**Sources:** Amara Osei; Ingrid Holm; Takeshi Yamamoto; Ravi Krishnamurthy; Claude, Codex, DeepSeek/Gemini

**Issue:** The design still blurs in-process event payloads, diagnostic events, scan-side idempotency, and load-bearing recovery. Critical reactions such as work release, identity retirement, drain recovery, binding cleanup, and wake recovery cannot depend only on at-most-once in-process event delivery. Reviewers also found mismatch between the proposed per-event matrix and current registered payloads, and `missed-event-recovered` is not a truthful trace outcome unless the system actually tracks event delivery.

**Required change:** Declare durable controller scans authoritative for all SDK-critical recovery unless a durable outbox/replay design is explicitly introduced. Add a per-reaction matrix naming source facts, assignee/key set, store routing, cadence, completeness guard, stale-event supersession, duplicate-run behavior, subscriber owner, event tier, payload fields, trace outcome, and tests. Replace or redefine `missed-event-recovered` with an outcome that reflects durable convergence rather than unobserved event loss.

### [Blocker] Reconciler, Runtime Fact, And Controller Policy Boundaries Are Not Yet Clean
**Sources:** Liam Okonkwo; Ingrid Holm; Natasha Volkov; Claude, Codex, DeepSeek

**Issue:** `AwakeInput` and desired-running logic still mix session lifecycle eligibility, runtime observations, controller demand, scale-check counts, work-query data, config identity, provider health, progress, restart budgets, and circuit state. Without an explicit eligibility/demand boundary, slices 5-7 can either move controller scheduling policy into `internal/session` or leave session deciders unable to reproduce current behavior. Runtime fact completeness is also missing for destructive decisions that must fail closed on stale, partial, timed-out, unobserved, or provider-unreachable data.

**Required change:** Define narrow session-domain fact inputs separately from reconciler/controller demand inputs. Rewrite desired-running as owned layers: reconciler demand and budget inputs, session eligibility and wake/hold rules, then reconciler provider-health/circuit/config/progress gates. Add fact completeness, provenance, staleness, partiality, and fail-closed rules for work, session, scale-check, runtime, provider-health, and progress data, plus guards that keep store/runtime/provider imports out of session deciders.

### [Blocker] Worker Boundary And API Routing Are Still Too Easy To Bypass
**Sources:** Sarah Chen; Elena Marchetti; Ravi Krishnamurthy; Amara Diallo; Claude, Codex, DeepSeek

**Issue:** The active worker-boundary guard only catches old `cmd/gc` bypass classes, not new session command APIs, manager method use, package-level functions such as `session.WakeSession`, or API-side lifecycle routing through `session.Manager` and direct session helpers. API read-side lifecycle projection is also unowned in files such as status/query/resolution handlers, and the classifier/conflict fields are not separated into internal-only versus wire-visible adapter contracts.

**Required change:** Add a per-caller routing table for CLI, legacy API, Huma API, reconciler/controller, worker, mail, and extmsg paths. State which package may construct and invoke session command values. Extend guards to block new disallowed production imports/calls and widen projection-helper checks for user-facing API read paths. Define the typed Huma/CLI adapter contract before exposing classifier candidate, negative-kind, or conflict details on the wire.

### [Major] Slice Sequencing Is Not Independently Shippable Or Revertible
**Sources:** Ravi Krishnamurthy; Takeshi Yamamoto; Elena Marchetti; Sarah Chen; Claude, Codex, DeepSeek/Gemini

**Issue:** Slice boundaries still follow broad field families more than actual consistency units. Runtime-start prepare, commit, rollback, pending-create cleanup, runtime identity, config hash, API materialization, and CLI fallback behavior must convert as one fenced cluster or have explicit typed coexistence conflicts. Cross-family operations such as `CloseDetailed`, `Suspend`, `UpdatePresentation`, `RetireNamedSessionPatch`, exact-ID repair, and API materialize-on-resolve can otherwise leave old and new writers mutating the same keys with different validation.

**Required change:** Re-anchor slices on concrete commands and written keys. Add coexistence tables for converted callers, legacy callers, allowed direct keys, validation differences, guard rows, bake criteria, rollback files, tests, and revert procedure. Require implementation bead metadata such as converts/allows-during-bake/retires so workflow gates can enforce migration ownership.

### [Major] Operability, Repair, Trace, Doctor, And Performance Proof Are Too Broad To Test
**Sources:** Ingrid Holm; Kwame Asante; Sarah Chen; Amara Osei; Claude, Codex, Gemini

**Issue:** The design promises rich operator evidence but does not map each operation result to trace site/reason/outcome codes, doctor/API/CLI rendering, logs, and tests. Repair/normalization writes like `RepairEmptyType` silently mutate durable session bead state from read paths without surfaced errors or centralized trace evidence. Performance budgets are also not concrete enough for reconciler ticks, target lookup, path-alias fallback, close/work-release scans, runtime-start repair scans, subscriber fan-out, event emission, throttle writes, or snapshot rebuild cost.

**Required change:** Route repair/normalize writes through one audited helper with before/after state, freshness/precondition data, surfaced errors, and trace/log evidence. Add per-operation diagnostic mappings and rendering tests for accepted, rejected, blocked, no-op, recovered, partial, deferred, failed, repair-applied, and repair-skipped cases. Add slice-level read/write/fan-out budgets and baseline query-count or benchmark proof before extraction.

### [Major] Shared Vocabulary Still Risks Becoming A New Facade
**Sources:** Kwame Asante; Amara Diallo; Liam Okonkwo; Takeshi Yamamoto; Claude, Codex, DeepSeek/Gemini

**Issue:** Classifier results, `SessionConflict`, diagnostic envelopes, `RuntimeStartIntent`, and event/fact vocabulary are still broader than the first adopting slices require. A flat classifier result with many optional fields, broad policy flags, and generic diagnostic payloads would recreate the facade the design is trying to remove. Future event sourcing and durable event-log compatibility also leak into current in-process event payload language.

**Required change:** Add slice vocabulary checkpoints that identify the first delegated caller, exact demanded fields, policy dimensions, non-goals, and rule-of-two evidence. Use tagged unions or per-kind population rules instead of broad optional result envelopes. Mark later-slice event and runtime field lists as provisional upper bounds, and keep slice 1 free of event vocabulary unless an adopting caller requires it.

### [Minor] Attempt-Scoped Persona Artifacts Are Still Written To The Wrong Directory
**Sources:** Workflow artifact inspection; all persona synthesis bead metadata

**Issue:** The ten persona synthesis beads for `gc.attempt=4` closed with `gc.outcome=pass`, but their `design_review.output_path` values point under `.gc/design-reviews/ga-unpr2y/attempt-1/persona-syntheses/`. The expected current-attempt directory `.gc/design-reviews/ga-unpr2y/attempt-4/persona-syntheses/` exists but is empty. This is the same artifact-placement defect called out in attempt 3, and it weakens auditability even though the stamped persona syntheses were available and used for this report.

**Required change:** Fix the persona synthesis output-path plumbing so attempt N writes under `attempt-N/persona-syntheses/`, and add a workflow check that fails the global synthesis when required current-attempt persona artifacts are absent or stamped to a stale attempt path.

## Disagreements
- The only persona-level verdict disagreement is Kwame Asante: YAGNI/contract-scope returned `approve-with-risks`, while the other nine personas returned `block`. I assess the global verdict as `block` because worst-verdict-wins applies and the blockers are implementation-safety issues, not preference differences.
- Several model lanes split on severity. Claude often rated a lane `approve-with-risks` where Codex or DeepSeek/Gemini rated the same contract gap `block`; in those cases the synthesis follows the stricter verdict when the issue affects command safety, behavior parity, or enforceable boundaries.
- Reviewers disagree on guard timing for slice 1. Mutation-boundary reviewers want an inventory freeze and baseline guard before adoption; the YAGNI lane warns that a read-only classifier should not be blocked on a speculative global AST scanner. The reconciled requirement is a slice-appropriate guard: target-resolution adoption/inventory freeze for slice 1, and mutation/write-path guards before mutation-owning slices.
- Reviewers also disagree on event payload richness. Some want explicit generation/token/idempotency fields; others see that as event-sourcing overreach. The practical resolution is to keep in-process payloads thin and diagnostic unless a specific load-bearing subscriber requires typed fields, while durable scans remain the correctness authority.
- A few concrete findings are single-source but code-grounded, including `RepairEmptyType`, `closeFailedCreateBead`, `session_circuit_*`, stale proof files, and `missed-event-recovered`. They should be verified rather than discarded, because each points at a specific file, key family, or trace contract.

## Missing Evidence
- Canonical production writer inventory for all session-owned lifecycle, identity, wake/hold/drain, runtime, repair, circuit, nudge, sleep, config-drift, and operational metadata writes.
- Exhaustive session-owned key taxonomy, including excluded or opaque reconciler-owned namespaces with rationale and cleanup rules.
- Build-failing guard design covering raw store writes, dynamic metadata maps, local wrappers, patch helpers, manager/package-level bypasses, API/CLI boundaries, and shrink-only exceptions.
- Reverse traceability from every `REQUIREMENTS.md` scenario row to slice ownership, preserved invariant, out-of-scope decision, and exact proof artifact.
- Current proof or restored replacements for deleted test files cited by the requirements ledger.
- Target classifier compatibility contract for exact ID, closed sessions, configured targets, rejected-by-config, path alias, qualified alias basename, mail, extmsg, assignee normalization, materialization, and historical aliases.
- Per-handler API and CLI routing inventory, with current route, target route, allowed package boundary, response compatibility proof, and OpenAPI/dashboard impact where relevant.
- Store atomicity and command concurrency contract for `SetMetadataBatch`, `Update`, `Close`, `BdStore.Tx`, external stores, and composed operations.
- Runtime-start recovery matrix for provider-start success followed by active metadata failure, session-key failure, event failure, process crash, or successor generation.
- Durable recovery scan contracts for work release, identity retirement, drain recovery, wake recovery, extmsg/wait cleanup, and binding cleanup.
- Session/reconciler fact-boundary design for eligibility, demand, provider health, progress, circuit state, runtime liveness, partial reads, and stale observations.
- Per-slice coexistence, bake, rollback, and revert drills for runtime-start, close/retire, API materialization, wake, repair, and cross-family writers.
- Operator diagnostics mapping for trace, doctor, logs, API, CLI, and session inspect outputs.
- Performance budgets and baseline measurements for target lookup, path alias, reconciler fact materialization, recovery scans, subscriber fan-out, event emission, throttle writes, and snapshot rebuilds.
- Correct attempt-scoped persona artifact placement under `.gc/design-reviews/ga-unpr2y/attempt-4/persona-syntheses/`.

## Recommended Changes
1. Freeze implementation decomposition until the mutation inventory, owned-key taxonomy, guard/allowlist contract, and API/worker-boundary routing table are source-complete and testable.
2. Tighten TR-001 into a row-by-row parity gate. Every touched requirement row needs exact proof, and stale proof citations must be restored, replaced, or marked proof-missing before slice 1 proceeds.
3. Rework target classification as a surface-by-surface adoption plan with a legacy oracle, not a big-bang replacement. Start with one proven caller, and define exact typed outcomes and compatibility tests before adopting more surfaces.
4. Choose the command concurrency and store atomicity model before any mutation-owning slice. Include conflict types, lock/transaction/revision behavior, partial-write repair, cross-process races, and event-after-commit ordering.
5. Fence runtime start as one consistency cluster or explicitly block all dependent slices until prepare, commit, rollback, pending-create cleanup, runtime identity, API materialization, and CLI fallback writers have one owner.
6. Make durable scans the authoritative recovery mechanism for critical reactions unless a durable outbox/replay design is approved. Events may accelerate and observe; they must not be the only correctness path.
7. Split reconciler/controller demand policy from session eligibility facts. Define owned layers for desired-running, runtime facts, provider health, progress, circuit state, and fail-closed behavior.
8. Add per-slice coexistence and rollback tables with converted callers, legacy callers, guard rows, bake criteria, and exact revert instructions.
9. Route repair/normalize writes through one audited helper and add trace/doctor/API/CLI rendering tests for repair, conflict, partial fact, recovery, and budget-deferral outcomes.
10. Scope shared vocabulary by first caller. Use tagged or per-kind result contracts and mark later-slice event/runtime fields as provisional upper bounds.
11. Add performance budgets and baseline query/benchmark proof before moving hot reconciler, target lookup, recovery scan, subscriber, or snapshot paths.
12. Fix the workflow artifact bug so persona syntheses for attempt 4 and future attempts are written under the current attempt directory, then add a guard that catches stale-attempt output paths.
