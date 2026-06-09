# Takeshi Yamamoto

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The Atomic Command Contract is not tied to an enforceable commit fence. The design requires exact pre-commit revalidation, token/revision/value preconditions, and stale-fact handling, but the current bead-store surface does not expose a clear compare-and-swap, revision, or expected-value update primitive for session metadata.
- [Blocker] Store capability is a prerequisite, not a deferred detail. Each persistence surface needs a matrix for atomicity, partial metadata update behavior, blind writes, close/reopen semantics, projection lag, cross-process visibility, and any conditional-write support before a mutating command can claim atomic semantics.
- [Blocker] Listing several fence strategies without choosing one per operation is unsafe. A command row must bind to exactly one strategy: store-native conditional write, value/token/revision fence with immediate pre-commit revalidation, or explicitly repair-converged blind write with eligibility limits and failure tests.
- [Major] Runtime side-effect ordering is still a checklist rather than an invariant. Runtime-start, wake, close, drain, stop/interrupt, repair, and identity operations need operation-specific durable mutation order, provider action order, event order, compensation or repair owner, and crash-scan recovery rules.
- [Major] Pure decider enforcement is not yet materialized. Future decider files need a guard that prevents store, runtime, config loading, event bus, API/CLI rendering, work/mail/extmsg, subprocess, and ambient clock access; `now` and relevant config must be explicit immutable facts for mutation-feeding deciders.
- [Major] The repair-converged blind-write path is currently too broad. If the design chooses NDI-style convergence instead of a store CAS, it must define which stale facts may be tolerated, which fields may be overwritten, what token or supersession key prevents older attempts from winning, and which durable scanner owns convergence.
- [Major] Existing compatibility paths are not mapped into command rows. The design must account for lifecycle patch helpers, process-local mutation locks, direct close paths, token backfills, manager close behavior, bead close behavior, reconciler paths, and partial metadata batches so advisory state transitions are not mistaken for durable command semantics.

**Disagreements:**
- Claude and DeepSeek return `block`; Codex returns `approve-with-risks` only for non-mutating Slice 0 decomposition. I choose `block` because this persona is about the atomicity contract that all later mutating slices depend on, and that contract currently lacks the store-capability answer and per-operation fence choice needed for safe decomposition.
- Claude frames the platform reality as "no CAS/revision today, so choose a new primitive or repair-converged blind writes." Codex accepts the design shape if every later row chooses a real strategy. DeepSeek pushes harder for a global cross-process synchronization or atomic batch standard. My assessment: either native conditional writes or tokened blind writes with deterministic repair can work, but the design must select and test one per operation.
- DeepSeek treats `ProjectLifecycle` wall-clock fallbacks as an immediate purity blocker. Codex requires a future pure-decider guard and notes those fallbacks must stay outside the pure set or be replaced. Claude emphasizes explicit `now` facts in the contract rather than citing the current fallback as the central defect. My assessment: mutation-feeding deciders must reject missing `now`; any legacy rendering fallback must be quarantined and excluded from command decider validation.
- DeepSeek uniquely calls out dual close paths, process-local mutation locks, token backfills, and sequential partial writes as blockers. The other reviews do not independently prove every implementation claim, but they align with the shared requirement that command rows map current behavior, concurrency assumptions, stale checks, and repair ownership before movement.
- DeepSeek also objects to reconciler evidence scope and attempt-directory mismatch. Those are workflow and parity concerns more than the core decider-atomicity lane; they should be carried as related risks but do not replace the store and command-contract blockers.

**Missing evidence:**
- A store capability row for the current persistence surfaces, including whether session commands use HQStore, BdStore, exec store, FileStore, MemStore, or all of them.
- A decision on whether Gas City will add a conditional/versioned bead-store write primitive or require all session-owned commands to use repair-converged blind writes.
- Per-operation command-applier rows for wake/start, close, drain, stop/interrupt, runtime-start prepare/commit/rollback, runtime-missing cleanup, identity retirement, token backfills, and repair/backfill.
- A pure-decider file/package set and import/call guard, with fixtures proving deciders cannot perform I/O, access stores/providers/events/config loaders, query work/mail/extmsg, render API/CLI responses, spawn subprocesses, or read ambient time.
- Failure-injection tests for stale snapshots, cross-process raced writers, partial metadata batches, duplicate commands, provider success followed by commit failure, commit success followed by event failure, skipped events, and durable scan convergence.
- A mapping from current lifecycle reducer/table behavior, patch helpers, manager close, bead close, reconciler paths, mutation locks, and token backfills to the future command contracts.
- Operation-order templates showing durable commit, provider action, event emission, compensation, retry, and repair ownership for each runtime-touching command.
- Materialized Slice 0 artifacts and CI validators: `COMMAND_APPLIERS.yaml`, `BOUNDARY_MATRIX.yaml`, store-capability inventory, pure-decider guard, and proof command.

**Required changes:**
- Record the real store-capability answer in Slice 0 artifacts: no implied CAS unless a concrete backend method exists, no implied revision token unless the schema provides one, and explicit treatment of blind writes, batch atomicity, partial failure, projection lag, and cross-process visibility.
- Choose the concurrency strategy before any mutating slice moves: add a conditional/versioned write primitive as a prerequisite, or remove unqualified CAS/token language and commit commands to repair-converged blind writes with deterministic scanners and race tests.
- Tighten the command-applier schema so each mutating row selects exactly one fence strategy and states method used, facts consumed, pre-commit reread point, expected token/revision/value, stale outcome, post-write verifier, repair owner, and failure-injection tests.
- Define eligibility limits for repair-converged blind writes, including stale-fact tolerance, overwrite rules, supersession token, safe partial states, and why destructive/runtime-identity operations do or do not require a stronger fence.
- Add operation-order templates for runtime-start, wake/start, close, drain, stop/interrupt, identity retirement, runtime-missing cleanup, and repair/backfill, including durable mutation order, provider action order, post-commit event order, provider-success/commit-failure behavior, commit-success/event-failure behavior, skipped-event behavior, and durable scan recovery.
- Enforce pure decider boundaries with a named file/package set, static import/call guard, positive and negative fixtures, mandatory non-zero `now` for mutation-feeding deciders, and explicit config facts instead of ambient config loading.
- Reconcile `SESSION-START-001` and other atomicity wording with the chosen strategy: define "atomic" as batch-atomic plus convergence, or gate the row on a new conditional-write primitive.
- Map existing close paths, lifecycle patch helpers, process-local locks, token backfills, and reconciler behavior into operation rows before creating implementation beads that move those callers.
