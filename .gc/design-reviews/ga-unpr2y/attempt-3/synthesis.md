# Design Review Synthesis

## Overall Verdict: block

All ten persona syntheses returned `block`, so worst-verdict-wins makes the global verdict `block`. The design is directionally aligned with session-owned lifecycle commands, typed projection boundaries, durable work recovery, and slice-based migration, but the current document still leaves several load-bearing contracts unenforceable: mutation ownership, target compatibility, atomicity, event recovery, runtime facts, and per-slice coexistence.

## Consensus Strengths
- Multiple reviewers recognized meaningful progress toward moving lifecycle mutation into session-owned command/decider surfaces rather than leaving callers to write raw metadata.
- The traceability matrix, operation/site/reason/outcome language, event recovery tiering, and bulk-snapshot performance intent are useful scaffolding for an implementable plan once they are made binding and slice-specific.
- Reviewers generally accept that slice 1 can remain first if it is narrowed to explicit classifier/adopter surfaces and kept independent of pool scheduling, runtime liveness, provider health, progress, and later mutation slices.
- The existing reconciler durable-fact scan model can be a valid convergence backstop for critical reactions, provided the design declares it authoritative and specifies cadence, completeness, idempotency, and tests.

## Critical Findings

### [Blocker] Mutation Boundary And Guards Are Not Enforceable
**Sources:** Elena Marchetti; Natasha Volkov; Sarah Chen; Ravi Krishnamurthy; Ingrid Holm; Claude, Codex, DeepSeek/Gemini

**Issue:** The design still lacks a canonical production writer inventory, exhaustive session-owned key taxonomy, parseable exception allowlist, and concrete static guard strategy. Reviewers found or disputed many direct writers and mutation paths, including `cmd/gc/session_reconcile.go`, `cmd/gc/session_reconciler.go`, `cmd/gc/cmd_session_wake.go`, `cmd/gc/session_circuit_breaker.go`, `cmd/gc/soft_reload.go`, `cmd/gc/cmd_nudge.go`, `cmd/gc/cmd_bd_store_bridge.go`, `internal/api/session_resolution.go`, `internal/api/handler_session_create.go`, `store.Update`, `RepairEmptyType`, and exported patch helpers. The current guard language does not handle dynamic metadata batches, generic bridges, session-bead discrimination, patch-builder use, API-layer bypasses, or shrink-only exception enforcement.

**Required change:** Replace the mutation landscape with a canonical call-site inventory and key-family taxonomy. Define an AST/symbol/package-boundary guard, a shrink-only allowlist with owner slice and expiry, and per-slice done criteria that retire specific inventory entries and guard exceptions.

### [Blocker] Target Classification Would Change Existing API, CLI, Mail, And Extmsg Behavior
**Sources:** Amara Diallo; Natasha Volkov; Sarah Chen; Kwame Asante; Claude, Codex, DeepSeek/Gemini

**Issue:** The proposed precedence table and result vocabulary do not preserve current resolver behavior. Configured named identity currently resolves before live `session_name` or alias in key paths, exact bead IDs can return closed session facts before operation-layer rejection, extmsg materializes configured named identities on miss, path aliases use state filtering plus most-recent-created tie-breaking, and mail has a separate taxonomy including `human`, configured recipients, unknown, and ambiguity. The design also lacks result kinds for configured-name conflict and rejected-by-config, and it mixes classification with operation policy such as materialization, forbidden targets, allow-closed, and read-only access.

**Required change:** Rewrite the classifier contract to mirror current behavior or explicitly mark behavior changes for requirements approval. Specify one-result versus all-candidates semantics, operation-policy inputs, result/negative kinds, retry/fallthrough protocols, and parity tests for API, CLI, mail, extmsg, assignee normalization, nudge, attach, inspect/log/transcript, and materialization paths.

### [Blocker] Behavior Parity Proof Is Too Weak And Some Cited Proof Is Stale
**Sources:** Natasha Volkov; Sarah Chen; Ingrid Holm; Elena Marchetti; Claude, Codex, DeepSeek/Gemini

**Issue:** TR-001 still allows proof against "at least one existing scenario row" even when a slice touches multiple rows. Runtime-start extraction is not pinned to create/start, stale-create rollback, pending-create clearing, projection freshness, and state-transition rows. Slice 1 lists broad visible surfaces while its proof mostly covers classifier/adaptor parity. The requirements ledger still cites absent proof files such as `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go`.

**Required change:** Require proof for every scenario row each slice touches, with exact row IDs and exact tests or commands. Restore or replace stale proof paths, add a citation-freshness check, and require characterization tests before moving behavior from reconciler, manager, API, CLI, mail, extmsg, worker, or session command paths.

### [Blocker] Atomicity, Concurrency, And Runtime-Start Recovery Are Underspecified
**Sources:** Takeshi Yamamoto; Ravi Krishnamurthy; Ingrid Holm; Claude, Codex, DeepSeek

**Issue:** The "lock-plus-reread" model does not identify the lock provider, scope, process model, lock order, conflict behavior, or whether CLI/API/controller/reconciler writes are serialized. The in-process session mutation lock cannot protect cross-process writers. `SetMetadataBatch` and `BdStore.Tx` may not provide the transactional semantics assumed by command rows. Runtime-start recovery remains unclear after provider side effects, especially when provider start succeeds but active/session-key metadata, event emission, or process crash happens before durable recovery information is complete.

**Required change:** Pick one enforceable command concurrency model: single-writer routing, real store-level conditional writes/transactions, safe sequential commit markers plus repair, or another concrete alternative. For each command, name facts re-read under the mechanism, write preconditions, typed conflicts, partial-write repair behavior, and race tests. Runtime start needs a recovery contract centered on `instance_token` as runtime identity and a single owner for prepare/commit/rollback/runtime-identity metadata.

### [Blocker] Critical Event Reactions And Work Release Lack A Durable Recovery Contract
**Sources:** Amara Osei; Ingrid Holm; Takeshi Yamamoto; Ravi Krishnamurthy; Claude, Codex, DeepSeek

**Issue:** Interim in-process session events are at-most-once and cannot be the only mechanism for work release, identity retirement, drain/assigned-work recovery, binding cleanup, wake recovery, or pack-level recovery. `SessionDrainAckedWithAssignedWork` remains ambiguous as diagnostic signal versus load-bearing recovery. The design does not reconcile new `SessionEvent` language with existing `events.Event` constants, registered payloads, SSE/dashboard projection, trace, and doctor output.

**Required change:** Add a per-event/reaction matrix with tier, payload fields, idempotency key, generation/instance token needs, authoritative durable recovery source, subscriber owner, stale-event supersession rule, and tests. Declare durable controller scans authoritative unless a durable outbox/replay marker is introduced for load-bearing subscribers.

### [Blocker] Runtime Facts, Desired-Running Composition, And Controller Ownership Are Not Separated
**Sources:** Liam Okonkwo; Ingrid Holm; Natasha Volkov; Claude, Codex, DeepSeek

**Issue:** `ComputePoolDesiredStates` combines controller scheduling policy with lifecycle eligibility, including request tiers, nested caps, in-flight reuse, named-session exclusion, scale-check demand, work aggregation, and capacity decisions. Provider health, progress, circuit-breaker state, restart counts, reset generation, alert deduplication, and budgets are stateful and idempotence-critical, but the design does not assign them cleanly to immutable session facts versus reconciler-owned state. Desired-running composition across terminal state, missing config, hold/quarantine, provider health, progress, demand, wake cause, and budgets is not specified.

**Required change:** Keep controller scheduling and capacity policy controller-owned unless the requirements intentionally change the boundary. Define fact completeness, provenance, timestamp, stale/partial/unknown/probe-derived semantics, fail-closed rules, desired-running precedence, and ownership of provider-health/progress/circuit side effects before creating implementation slices for these areas.

### [Blocker] The Slice Plan Is Not Independently Shippable Or Revertible
**Sources:** Ravi Krishnamurthy; Elena Marchetti; Sarah Chen; Takeshi Yamamoto; Claude, Codex, DeepSeek/Gemini

**Issue:** Slice ownership does not line up with real write topology. Runtime-start prepare, commit, rollback, pending-create cleanup, runtime identity, and runtime hash metadata are split across multiple slices even though they currently behave as one consistency unit. Cross-family operations such as `CloseDetailed`, `RetireNamedSessionPatch`, `Suspend`, `UpdatePresentation`, API materialization/retirement, and wake/start/close paths can write keys owned by multiple slices. The plan lacks per-slice coexistence tables, bake criteria, rollback procedures, and API coexistence for `internal/api/session_resolution.go` and `internal/api/session_manager.go`.

**Required change:** Re-anchor slices on specific commands and written keys, not broad field families. Add per-slice coexistence tables naming converted callers, legacy callers, validation differences, allowed direct keys, guard updates, retirement conditions, bake criteria, and revert procedures. Assign runtime-start to one slice or make all dependent slices wait.

### [Major] Operability, Trace, Doctor, And Performance Contracts Are Too Broad To Test
**Sources:** Ingrid Holm; Kwame Asante; Sarah Chen; Amara Osei; Claude, Codex, DeepSeek/Gemini

**Issue:** The design promises trace/API/CLI/log/doctor evidence broadly, but does not map each operation result to trace site/reason/outcome codes, record fields, typed conflicts, API/CLI formatting, logs, doctor output, and rendering tests. It also lacks slice-level read/fan-out budgets and proof for target lookup, path-alias resolution, close/work-release scans, subscriber fan-out, partial facts, and snapshot mutation cost.

**Required change:** Add per-operation diagnostic mappings and per-slice complexity budgets. Require `gc trace` or equivalent session-inspection rendering tests for accepted, rejected, blocked, no-op, recovered, partial, deferred, failed, and repair paths.

### [Major] Shared Vocabulary Risks Becoming A New Facade
**Sources:** Kwame Asante; Amara Diallo; Liam Okonkwo; Takeshi Yamamoto; Claude, Codex, DeepSeek

**Issue:** The classifier result, diagnostic envelope, `SessionConflict`, `RuntimeIntent`, and `SessionEvent` vocabulary are either too broad or not scoped to the first slice that needs them. A flat classifier with many optional fields and an "every result carries every diagnostic field" rule recreates the broad facade the design is trying to avoid.

**Required change:** Add slice vocabulary checkpoints. For every new shared type or policy dimension, record the first delegated caller, required payload fields, non-goals, and rule-of-two or demand evidence. Use tagged unions or per-kind population rules where a flat optional shape would invite expansion.

### [Minor] Artifact Placement For Attempt 3 Is Inconsistent
**Sources:** Workflow artifact inspection

**Issue:** The ten persona body beads for `gc.attempt=3` stamped `design_review.output_path` values under `attempt-1/persona-syntheses`, while `attempt-3` contained no `persona-syntheses/` directory. The artifacts were recent, closed with `gc.outcome=pass`, and used for this synthesis, but the directory mismatch weakens artifact auditability.

**Required change:** Fix the persona synthesis output-path plumbing so attempt-scoped source files live under the current attempt directory and future global syntheses can rely on `$ATTEMPT_DIR/persona-syntheses`.

## Disagreements
- Several personas had split model verdicts, often with Claude at `approve-with-risks` and Codex plus DeepSeek/Gemini at `block`. I assess the blocking verdicts as controlling because even the approving-with-risks reviews usually required changes to contracts that implementation beads would otherwise depend on.
- Reviewers offered different acceptable atomicity remedies: single-writer controller routing, store-level CAS/transactions, lock-plus-revision with explicit conflict behavior, or commit markers with repair. These are not incompatible; the design must choose one and apply it per command.
- Reviewers disagreed on classifier shape and ownership: selected result versus all candidates, path alias in the core classifier versus adapter-supplied candidate, and reserved `template:` rejection inside versus outside classification. Any of these can work if the design preserves current behavior and gives callers a typed protocol.
- Some DeepSeek/Gemini-specific findings were not independently verified by every model, including `RepairEmptyType`, `store.Update`, `closeFailedCreateBead`, missing proof files, and `gc trace` surface details. They are credible enough to require verification because they point at concrete files or contracts, but they should not be the sole basis for a blocker where broader evidence already exists.
- Reviewers agree the design direction is better than earlier versions. The disagreement is not whether the migration should proceed, but whether the current document is concrete enough to safely decompose into implementation work. It is not yet.

## Missing Evidence
- Canonical production mutation inventory with call-site IDs, operation type, key families, session-bead discrimination, owner slice, retirement condition, and exception status.
- Exhaustive session-owned key taxonomy covering lifecycle, identity, wake/hold/drain, circuit breaker, runtime identity, startup verification, nudge delivery, config drift, sleep/hold, provider repair, and operational metadata.
- Concrete static guard design with implementation strategy, false-positive/false-negative tests, patch-helper containment, generic bridge policy, and shrink-only allowlist.
- Scenario-row parity matrix for every slice, exact proof files or commands, and a CI check that cited proof paths still exist.
- Target classifier precedence/result contract and caller protocols for API, CLI, mail, extmsg, assignee normalization, nudge, attach, inspect/log/transcript, and materialization.
- Command concurrency and store atomicity contract, including cross-process race tests and partial-write repair matrices.
- Runtime-start recovery contract for provider-start success followed by commit/event/crash failures.
- Per-event/reaction delivery matrix and durable recovery contracts for work release, drain recovery, identity retirement, binding cleanup, wake recovery, and pack subscribers.
- Fact completeness and fail-closed semantics for work, session, scale-check, runtime liveness, provider health, progress, pool demand, and partial/degraded reads.
- Desired-running composition and provider/progress/circuit/budget ownership model.
- Per-slice coexistence, bake, and revert plans, including worker-boundary routing and API materialization coexistence.
- Per-operation trace/API/CLI/log/doctor mappings and rendering tests.
- Slice-level read/fan-out budgets and benchmark/query-count proof.

## Recommended Changes
1. Freeze implementation bead creation until the design adds a canonical mutation inventory, owned-key taxonomy, patch-helper inventory, and enforceable static guard/allowlist contract.
2. Rewrite slice ownership around commands and concrete written keys. Give runtime start a single shippable owner or make all dependent slices wait for one owner of prepare, commit, rollback, pending-create cleanup, runtime identity, and runtime hash metadata.
3. Tighten TR-001 so every touched scenario row requires exact proof, then fix stale requirement proof citations and add a citation-freshness guard.
4. Rework target classification to preserve current resolver behavior and typed failure modes, including configured-name conflict, rejected-by-config, exact closed-ID handling, path alias rules, extmsg materialization-on-miss, mail resolution, historical aliases, and reserved `template:` handling.
5. Define the command concurrency model and store atomicity requirements before any mutation-owning slice proceeds. Include conflict types, lock or transaction semantics, partial-write repair, and cross-process race tests.
6. Make durable scans or durable replay the authoritative recovery mechanism for critical reactions. Events may trigger or observe, but must not be the only path for work release, identity retirement, drain recovery, or binding cleanup.
7. Separate controller scheduling policy from session lifecycle eligibility. Define desired-running composition, fact completeness, stale/partial/unknown semantics, and fail-closed behavior before extracting runtime, provider-health, progress, or pool-demand logic.
8. Add routing and coexistence tables for CLI, legacy API, Huma API, reconciler/controller loops, mail/extmsg, and assignee normalization. Update guards so new session command APIs cannot bypass the active worker-boundary migration.
9. Scope shared vocabulary by slice. Replace broad optional result envelopes with tagged or per-kind contracts, and record first caller, payload fields, and non-goals for `SessionConflict`, classifier results, `RuntimeIntent`, and `SessionEvent`.
10. Add per-operation operability mappings and tests for trace, API/CLI output, logs, doctor, recovery, partial facts, no-ops, conflicts, and repair writes.
11. Add slice-level performance/read budgets, especially for target lookup, path alias fallback, close/work-release scans, subscriber fan-out, and snapshot mutation.
12. Fix the attempt-scoped artifact path bug so persona syntheses for attempt N are written under `attempt-N/persona-syntheses`.
