# Takeshi Yamamoto - DeepSeek V4 Flash (Attempt 15 Review)

**Verdict:** block

**Review scope:** Pure decider enforcement, optimistic concurrency, commit-event-intent ordering, stale-fact defense, and boundary-inventory-enforceability for the Decider Atomicity Enforcer mandate. This reviews the current Attempt 15 iteration of `internal/session/DESIGN.md` (Attempt 15 input) against `REQUIREMENTS.md`, `AGENTS.md`, and the active checkout source.

---

## Top Strengths

- **Explicit Separation of Classification from Mutation**: Establishing the Target Classification Contract as a side-effect-free, read-only taxonomy is a highly welcome design pattern. Keeping the raw classifier isolated from store writes, session materialization, and API/CLI rendering decisions (DESIGN.md:148-156) guarantees that fact resolution remains a pure read-only projection.
- **Idempotency Over Transient Ephemeral Events**: Relying on durable fact scans as the definitive source of truth for critical lifecycle convergence (such as work release and session retirement) instead of solely trusting in-process, transient event propagation aligns perfectly with the Bitter Lesson. Treating events as latency optimizations rather than command vectors is structurally correct.
- **Rigorous Requirements Parity Framing**: Mapping each backlog slice to specific `SESSION-*` scenario rows in `SCENARIO_PARITY.yaml` with explicit proof command expectations ensures that behavior drift or unapproved product alterations are statically prevented prior to decomposition.

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
The underlying `beads.Store` implementations apply metadata batch mutations sequentially rather than atomically. In `internal/beads/exec/exec.go:417-424`, `SetMetadataBatch` is a simple iterative loop:
```go
func (s *Store) SetMetadataBatch(id string, kvs map[string]string) error {
	for k, v := range kvs {
		if err := s.SetMetadata(id, k, v); err != nil {
			return err
		}
	}
	return nil
}
```
If a process crashes or a store error occurs mid-loop, a session bead is left in a corrupted, partially mutated state (e.g. `state` updated to `creating` but the `instance_token` write is missing or incomplete). The design doc lists "partial-state matrix" as a required field on every operation-specific command contract, but does not provide a global, reusable repair mechanism for partial writes, leaving individual slices to handle complex recoverability independently.

### 4. [Blocker] Dual Close-Path Divergence and Concurrency Gaps
The codebase continues to contain two structurally divergent and concurrent close paths:
- **`CloseDetailed`** (`internal/session/manager.go:862`): stops the provider, cancels waits, clears wake/hold overrides, retires named session identifiers, and closes the bead.
- **`closeBead`** (`cmd/gc/session_beads.go:2144`): directly writes basic metadata and closes the bead without wait cancellation, override clearing, or identifier retirement.

Because `closeBead` is frequently invoked by the reconciler and lifecycle paths in `cmd/gc` without wait cancellation or override clearing, closed session beads will routinely retain active wake/hold overrides in the database, risking unintended recreations or misrouting on subsequent reconciler ticks. The design must mandate the complete unification of these close paths into a single session-owned command.

### 5. [Blocker] Lock-Free, Race-Prone `instance_token` Backfills
The design positions `instance_token` as the authoritative identity for start attempts. However, the active checkout source performs backfill writes to this field outside any lock, validation, or conditional write guards in `internal/session/chat.go:292–302` and `chat.go:400–409`. Concurrent reconciler sweeps can still execute concurrent backfills on a session that lacks a token, causing them to generate different tokens and overwrite each other.

### 6. [Blocker] Missing Concrete Boundary Artifacts
The non-mutating Slice 0 preflight (DESIGN.md:98-132) is a mandatory gate that requires the creation of several artifacts (`BOUNDARY_INVENTORY.md`, `allowlist.yaml`, `SCENARIO_PARITY.yaml`, and associated unit tests). Because these concrete files and automated guards do not fully exist in the repository checkout or run in CI as failing checks today, there is no active protection against regression or pattern drift.

---

## Required Changes

1. **Enforce Call-Level Decider Purity**: Completely remove the local clock fallback from `lifecycle_projection.go` and make the `now` timestamp a mandatory, non-zero field on `LifecycleInput`.
2. **Standardize Cross-Process Concurrency Control**: Define a global architectural standard for cross-process concurrency control (e.g. dolt-level conditional write fences, store-level validation hooks, or file locks) to resolve the single-process limitation of `WithSessionMutationLock`.
3. **Establish Reusable Partial-State Recovery**: Provide a unified, centralized recovery/repair helper rather than leaving individual slices to define custom partial-state handling for multi-key metadata batches.
4. **Unify Close Paths**: Fold `closeBead` and `CloseDetailed` into a single session-owned close command that guarantees wait cancellation, override clearing, and identity retirement are always applied atomically or resolved via an explicit, durable repair loop.
5. **Protect/Eliminate `instance_token` Backfills**: Move the `instance_token` backfills inside a validated transaction or the unified `PreWakePatch` batch.
6. **Materialize Slice 0 Artifacts**: Create and fully integrate the required Slice 0 preflight artifacts and tests in the codebase as active CI gates before approving decomposition.

---

## Questions

1. Since `WithSessionMutationLock` only serializes access inside a single Go process, how will we prevent the `gc` CLI process and the controller reconciler daemon process from concurrently mutating the same session bead's metadata?
2. Should we specialize the `beads.Store` interface to support an atomic metadata batch update (e.g. utilizing an underlying transaction or database lock) for compatible backends to prevent partial-state failures?
3. What is the rollback story if a caller adopts a new command API but the reconciler or CLI lacks a matching concurrency test to prove that they can recover safely from a partial write or a TOCTOU race?
