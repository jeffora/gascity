# Takeshi Yamamoto - DeepSeek V4 Flash (Attempt 7 Review)

**Verdict:** block

**Review scope:** Pure decider enforcement, optimistic concurrency, commit-event-intent ordering, stale-fact defense, and boundary-inventory-enforceability for the Decider Atomicity Enforcer mandate. This reviews the current Attempt 7 iteration of `internal/session/DESIGN.md` (Attempt 7 input) against `REQUIREMENTS.md`, `AGENTS.md`, and the active checkout source.

---

## Top Strengths

- **Realistic Stale-Fact Mitigation Philosophy**: The design response to Attempt 5 blocks (DESIGN.md:51-53) and Attempt 6 blocks (DESIGN.md:73-82) represents an excellent, mature posture for a decentralized system. The design correctly rejects simple, non-atomic in-memory locks or raw `SetMetadata` read-write sequences, demanding tokened phase markers, prepare-commit-rollback transitions, and bounded reconciler-led repair.
- **Idempotency Over Ephemeral Event Convergence**: Moving away from a dependency on at-most-once, synchronous, in-process event delivery for critical operations like work release, and instead centering the system's reliability on idempotent, durable store scans (DESIGN.md:320-322, 914-917) is an excellent alignment with the Bitter Lesson. The design correctly treats events as latency optimizations and scans as the true correctness backstop.
- **Robust Attempt Identity with `instance_token`**: Prioritizing a random, highly entropy-rich `instance_token` as the primary and authoritative attempt identity for runtime start (DESIGN.md:846-847, 885-888) prevents generation wrap-around and sequencing errors that inevitably plague simple monotonic integer sequences in distributed, crash-prone environments.

---

## Critical Risks

### 1. [Blocker] Pure Decider Purity Compromised by System Wall-Clock Reads
The fundamental primitive of TR-003 is that session deciders are pure functions consuming only immutable facts and no clocks. However, `ProjectLifecycle`—the canonical, load-bearing decider cited by the design—directly violates this by reading the local machine clock on the zero-`now` fallback path (`internal/session/lifecycle_projection.go:381-382`):
```go
	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
```
Furthermore, the design's proposed guard ("Pure session deciders must live in a guardable file set with no imports of ... clocks", DESIGN.md:1024) is structurally impossible to implement. Deciders must import Go's `time` package to parse, store, and compare `time.Time` fields on the `LifecycleInput` and `LifecycleView` structs. An import-level block is therefore structurally blunt.

**Impact**: If a caller fails to pass a validated `now` timestamp, the decider yields non-deterministic projection results (such as blocker or quarantine durations) based on the local machine's current wall clock. This breaks test determinism, regression replayability, and state-machine predictability. Decider purity must be call-level and absolute: `now` must be a mandatory, non-nullable field, and the internal wall-clock fallback must be completely removed.

### 2. [Blocker] Multi-Process TOCTOU Race and Unfenced Store Commit
The design specifies that session commands defend against stale facts by re-reading or validating a precondition immediately before commit (DESIGN.md:834). However, `beads.Store` and its underlying implementations (such as `exec.Store` or `BdStore`) provide **no atomic compare-and-swap (CAS), conditional write primitives, or revision-locked commits**. 

The `Tx` transaction surface in `internal/beads/beads.go` lacks a `Get` method, meaning a read-modify-write cannot be expressed inside a backend database transaction. Consequently, every command-applier read-compare-write sequence (such as `asyncStartIdentityMatches` in `cmd/gc/session_lifecycle_parallel.go:1417`) is a TOCTOU (time-of-check to time-of-use) window:
1. Process A reads the session bead and verifies that `instance_token` still matches.
2. Process B (e.g., a concurrent reconciler sweep or command-caller) rotates the token or closes the bead.
3. Process A executes `SetMetadataBatch` or `Update`, blindly overwriting the newly committed token/status.

**Impact**: The system has no physical protection against a stale writer stomping on a newer token rotation. While the design leans on "idempotent repair converges," it fails to obligate each slice to provide a concurrency test that explicitly drives this TOCTOU race (e.g., interleaving a newer token commit between another process's read and write phases) to prove convergence.

### 3. [Blocker] Sequential Non-Atomic Partial-State Metadata Writes
The `Store` interface doc-comment (`internal/beads/beads.go:240-246`) explicitly states that external store implementations (such as `exec.Store` or `BdStore`) apply metadata batches sequentially rather than atomically, meaning partial application is possible on mid-batch failure. In `internal/beads/exec/exec.go:417-424`, `SetMetadataBatch` is implemented as a simple sequential loop that returns immediately on first error:
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
The Command Atomicity Contract (DESIGN.md:848-849, 877) correctly states that every visible partial state must have a repair row and test. However, the design under-specifies this for the runtime-start and close command clusters, which write multiple keys (such as `state`, `instance_token`, start hashes, and `pending_create_claim` clearing) in a batch.

**Impact**: A crash or store error mid-loop will leave a session bead in a corrupted, half-written state (e.g., state is marked `creating` but the `instance_token` write is missing or incomplete). Without an explicit, codified partial-state repair matrix and test expectations for every partial subset of keys, the controller and reconciler are highly likely to hang or trigger infinite loops on corrupted beads.

### 4. [Blocker] Dual Close-Path Divergence and Concurrency Gaps
The codebase contains two structurally divergent, concurrent close paths:
- **`CloseDetailed`** (`internal/session/manager.go:862-920`): executes within the `WithSessionMutationLock` boundary, stops the provider, cancels waits, clears wake/hold overrides via batch write, retires named session identifiers, and then closes the bead.
- **`closeBead`** (`cmd/gc/session_beads.go:2144-2175`): runs without any lock, performs a sequential `setMetaBatch` with `ClosePatch` (which only writes state/reason/closed_at), and closes the bead without wait cancellation, override clearing, or identifier retirement.

**Impact**: A session closed via the low-level `closeBead` (e.g., during stale-session cleanup, dead-runtime sweeps, or CLI-driven close loops) remains un-retired and retains its active wake/hold overrides in the database. On the next reconciler tick, the reconciler can observe these active overrides on the closed bead and attempt to recreate, wake, or misroute the session against operator intent. The design must mandate the unification of these close paths into a single session-owned command.

### 5. [Blocker] Lock-Free, Race-Prone `instance_token` Backfills
The design positions `instance_token` as the authoritative identity for start attempts (DESIGN.md:846-847). However, the active checkout source performs backfill writes to this field outside any lock, validation, or conditional write guards in `internal/session/chat.go:292–302` and `chat.go:400–409`:
```go
	instanceToken := b.Metadata["instance_token"]
	if instanceToken == "" {
		instanceToken = NewInstanceToken()
		if err := m.store.SetMetadata(id, "instance_token", instanceToken); err != nil {
			return fmt.Errorf("storing instance token: %w", err)
		}
		b.Metadata["instance_token"] = instanceToken
	}
```

**Impact**: If two concurrent reconciler sweeps or command-caller threads race on a session that lacks an `instance_token` (such as a bootstrap session), both will observe an empty token, generate different tokens, and perform back-to-back sequential `SetMetadata` writes. The second writer overwrites the first, violating the core invariant that `instance_token` uniquely identifies a single start attempt. These backfills must be unified into the `PreWakePatch` or strictly wrapped in the `WithSessionMutationLock`.

### 6. [Blocker] Missing Concrete Boundary Artifacts (Slice 0 Preflight is Pure Prose)
The Attempt 6 revision introduced the non-mutating Slice 0 preflight (DESIGN.md:65-87) as a mandatory gate before any mutation-owning implementation slices can proceed. This gate requires `internal/session/BOUNDARY_INVENTORY.md`, `cmd/gc/testdata/session_boundary_guard/allowlist.yaml`, static guard tests (`TestSessionBoundaryGuard`, `TestSessionBoundaryInventoryFresh`), scenario parity source (`SCENARIO_PARITY.yaml`, `TestScenarioParityFreshness`), and vocabulary checkpoint checks (`TestVocabularyCheckpoints`).
However, **none of these files, schemas, guards, or tests exist in the active repository checkout**. 

**Impact**: The Slice 0 preflight gate is purely descriptive prose. Because these concrete guardrails do not exist in reality, there is no static protection against regression or pattern drift as developers begin implementing subsequent slices. As stated in DESIGN.md:83-87, this design continues to block decomposition until those artifacts are created and verified in code.

---

## Required Changes

1. **Enforce Call-Level Decider Purity**: Remove the `time.Now()` fallback from `lifecycle_projection.go` (line 381) and make the `now` timestamp a mandatory, non-zero field on `LifecycleInput`. Update the TR-003 static guard specification to scan the AST for calls to `time.Now()` or store reads inside the decider file set, rather than relying on blunt import-level package filters.
2. **Mandate TOCTOU Concurrency Race Tests**: For every command-applier row in the Command Atomicity Contract (DESIGN.md:895), mandate a concurrency test that explicitly drives the read-compare-write race (e.g., a newer token committed between the snapshot read and the write phase) and proves that the next repair scan converges or that the write is unconditionally idempotent.
3. **Specify Partial-State Repair for Batch Metadata**: For all multi-key metadata writes (such as runtime-prepare and runtime-commit), require a codified partial-state repair matrix and tests that verify the reconciler can recover from a partial write of any subset of the keys.
4. **Unify the Dual Close Paths**: Fold `closeBead` and `CloseDetailed` into a single, session-owned close command that ensures wait cancellation, override clearing, and identity retirement are always applied atomically or resolved via an explicit repair loop.
5. **Protect/Eliminate `instance_token` Backfills**: Move the `instance_token` backfills in `chat.go` inside the `WithSessionMutationLock` boundary, or migrate them into the pre-validated `PreWakePatch` batch.
6. **Implement the Slice 0 Preflight**: Before marking this design approved, the Slice 0 preflight artifacts (`BOUNDARY_INVENTORY.md`, `allowlist.yaml`, AST boundary guard, `SCENARIO_PARITY.yaml`, and associated tests) must be created and verified passing in the checkout.

---

## Questions

1. Since `Tx` has no `Get` and conditional commits are not supported at the store layer, is tokened prepare/commit + deterministic repair the permanent, non-negotiable architectural model? If so, should the convergence proof of this model under races be a centralized, reusable test fixture rather than left to individual slices?
2. Does `BdStore.SetMetadataBatch` (or equivalent backend transaction) execute atomically at the `bd` or `dolt` layer? If yes, should we specialize the store interface to expose a proven atomic batch operation for compatible backends?
3. Is the clock fallback in `ProjectLifecycle` currently relied upon by any caller that passes a zero `now` value? If so, is making `now` mandatory a breaking behavior change that requires a new `REQUIREMENTS.md` row or owner-approved requirements amendment?
