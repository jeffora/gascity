# Takeshi Yamamoto - DeepSeek V4 Flash (Attempt 16 Review)

**Verdict:** block

**Review scope:** Pure decider enforcement, optimistic concurrency, commit-event-intent ordering, stale-fact defense, and boundary-inventory-enforceability for the Decider Atomicity Enforcer mandate. This reviews the current Attempt 16 iteration of `internal/session/DESIGN.md` (Attempt 16 input) against `REQUIREMENTS.md`, `AGENTS.md`, and the active checkout source.

---

## Top Strengths

- **Rigorous Coexistence and Rollback Mapping**: The newly introduced `Migration Coexistence And Rollback` section (DESIGN.md:430-451) is a major architectural milestone. Requiring a detailed migration row with predecessor/successor order, field-family ownership transitions, raw-writer retirement conditions, and cross-process fences addresses the high risk of split-brain writes on overlapping CLI and controller files.
- **Detailed Schema for Matrix Boundaries**: Adding explicit row-schema requirements to `BOUNDARY_MATRIX.yaml` (including policy owner, allowed inputs/outputs, destructive-action safety rules, and wake-cause production owner) ensures that the boundary between session deciders and reconciler adapters is fully audited rather than left to implicit understanding.
- **SSE/OpenAPI Event Alignment**: Section 8's new requirements for event taxonomy (specifically inventorying standard events like `session.woke`, `session.stranded`, and `session.drain_acked_with_assigned_work` against public SSE/OpenAPI client visibility) ensures that event tracking serves as a typed and observable system state projection rather than un-modeled log spam.

---

## Critical Risks

### 1. [Blocker] Pure Decider Purity Compromised by System Wall-Clock Reads
The core invariant of the decider model is that session deciders are pure functions consuming only immutable, explicit facts. However, `ProjectLifecycle`—the load-bearing decider—retains local wall-clock fallbacks when `input.Now` is zero (`internal/session/lifecycle_projection.go:380-382` and `608-610`):
```go
	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
```
If a caller fails to pass an explicit timestamp, the decider yields non-deterministic projection results based on the local OS clock. This breaks test replayability and makes state-machine predictability impossible. Purity must be call-level and absolute: the `now` field on `LifecycleInput` must be mandatory, and the system clock fallback must be completely removed.

### 2. [Blocker] Multi-Process TOCTOU Race and Unfenced Store Commit
The design positions `WithSessionMutationLock` as a primary tool to serialize metadata mutations and avoid concurrent modification races. However, `WithSessionMutationLock` in `internal/session/chat.go:165-199` is implemented as a purely **in-memory, single-process mutex lock** using Go's `sync.Mutex` and an in-process map:
```go
var (
	sessionMutationLocksMu sync.Mutex
	sessionMutationLocks   = map[string]*sessionMutationLockEntry{}
)
```
While this prevents concurrent writes within a single process (such as multiple threads inside the controller daemon), the **Gas City CLI runs in a completely separate OS process** from the controller daemon. When the operator runs a CLI command (like `gc stop`) and the daemon runs its reconciler loop concurrently on the same session bead:
1. Both processes have entirely independent memory spaces and cannot see each other's in-memory mutexes.
2. They will execute concurrent read-compare-write sequence blocks on `beads.Store` with no physical synchronization.
3. Because `beads.Store` does not support atomic compare-and-swap (CAS) or transaction-level revision locking on metadata keys, both processes will race, leading to lost updates or corrupted states.

The design doc avoids specifying how multi-process atomicity is actually achieved, pushing the problem to individual implementation slices via "Each slice must define an operation-specific command contract" (DESIGN.md:235-237). Without a global, cross-process synchronization standard (such as a filesystem/dolt-level lock, validation hooks inside the store, or conditional fence writes), this is a major architectural gap.

### 3. [Blocker] Sequential Non-Atomic Partial-State Metadata Writes
The underlying `beads.Store` implementations apply metadata batch mutations sequentially rather than atomically. In `internal/beads/bdstore.go:877-889`, `SetMetadataBatch` sorting and sequential processing performs unconditional writes:
```go
	for _, k := range keys {
		args = append(args, "--set-metadata", k+"="+kvs[k])
	}
```
Similarly, in `internal/beads/hqstore_core.go:393`, `SetMetadataBatch` merges keys directly under an in-process lock. If a process crashes or a store error occurs mid-loop, a session bead is left in a corrupted, partially mutated state (e.g. `state` updated to `creating` but the `instance_token` write is missing or incomplete). The design doc lists "partial-state matrix" as a required field on every operation-specific command contract, but does not provide a global, reusable repair mechanism for partial writes, leaving individual slices to handle complex recoverability independently.

### 4. [Blocker] Dual Close-Path Divergence and Concurrency Gaps
The codebase continues to contain two structurally divergent and concurrent close paths:
- **`CloseDetailed`** (`internal/session/manager.go:862`): stops the provider, cancels waits, clears wake/hold overrides, retires named session identifiers, and closes the bead.
- **`closeBead`** (`cmd/gc/session_beads.go:2144`): directly writes basic metadata and closes the bead without wait cancellation, override clearing, or identifier retirement.

Because `closeBead` is frequently invoked by the reconciler and lifecycle paths in `cmd/gc` without wait cancellation or override clearing, closed session beads will routinely retain active wake/hold overrides in the database, risking unintended recreations or misrouting on subsequent reconciler ticks. The design must mandate the complete unification of these close paths into a single session-owned command.

### 5. [Major] Scope Creep on Slice 0 - Reconciler/Pool Evidence Blocking Session Boundary
The non-mutating Slice 0 gate forces the preflight to repair or owner-retire evidence for `SESSION-RECON-002` (Cold pool scale from zero), `SESSION-RECON-003` (Existing rig session prevents cold wake), `SESSION-RECON-006` (Provider health gate), and `SESSION-RECON-007` (Progress-aware health) before dependent slices can proceed.
However, pool scaling, provider health gates, and progress thresholding are explicitly reconciler/pool adapter behaviors that the `Boundary Matrix` places *outside* `internal/session`. Forcing the non-mutating Session Slice 0 to validate and resolve stale reconciler-specific requirement evidence is an unnecessary scope creep. It creates a premature dependency on the reconciler's internal state machine.

### 6. [Major] Ongoing Workflow Metadata and Directory Mismatch
The physical subtask bead (`ga-vd97hl`) is created with `gc.attempt: 1` even though it is executing as part of logical attempt/iteration 16 (indicated by `gc.scope_ref`). This forces reviews to either be written to the obsolete `attempt-1/` directory or require manual directory overrides to be visible in `attempt-16/`. While this is seen as "just workflow plumbing" by other reviewers, it introduces serious cross-document and cross-directory review dispersion that hinders automated verification.

---

## Required Changes

1. **Enforce Call-Level Decider Purity**: Completely remove the local clock fallback from `lifecycle_projection.go` and make the `now` timestamp a mandatory, non-zero field on `LifecycleInput`.
2. **Standardize Cross-Process Concurrency Control**: Define a global architectural standard for cross-process concurrency control (e.g. dolt-level conditional write fences, store-level validation hooks, or file locks) to resolve the single-process limitation of `WithSessionMutationLock`.
3. **Establish Reusable Partial-State Recovery**: Provide a unified, centralized recovery/repair helper rather than leaving individual slices to define custom partial-state handling for multi-key metadata batches.
4. **Unify Close Paths**: Fold `closeBead` and `CloseDetailed` into a single session-owned close command that guarantees wait cancellation, override clearing, and identity retirement are always applied atomically or resolved via an explicit, durable repair loop.
5. **Protect/Eliminate `instance_token` Backfills**: Move the `instance_token` backfills inside a validated transaction or the unified `PreWakePatch` batch.
6. **Refactor Slice 0 Entry Scope**: Remove the requirement to resolve reconciler/pool requirement evidence (`SESSION-RECON-002`, `003`, `006`, `007`) from the Session Slice 0 entry criteria, moving that burden to pool/reconciler-specific slices.
7. **Materialize Slice 0 Artifacts**: Create and fully integrate the required Slice 0 preflight artifacts and tests in the codebase as active CI gates before approving decomposition.

---

## Questions

1. Since `WithSessionMutationLock` only serializes access inside a single Go process, how will we prevent the `gc` CLI process and the controller reconciler daemon process from concurrently mutating the same session bead's metadata?
2. Should we specialize the `beads.Store` interface to support an atomic metadata batch update (e.g. utilizing an underlying transaction or database lock) for compatible backends to prevent partial-state failures?
3. How does the design justify placing reconciler-owned requirements (`SESSION-RECON-*`) as blocking criteria for the Session Slice 0 gate, given that they violate the Reconciler/Session split?
