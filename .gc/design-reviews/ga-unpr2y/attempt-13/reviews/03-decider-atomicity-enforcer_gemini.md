# Takeshi Yamamoto — DeepSeek V4 Flash (Independent Review, Attempt 13)

**Verdict:** block

**Review focus:** Pure deciders, optimistic concurrency, commit-event-intent ordering, stale fact defense — with direct evidence from the active codebase. Evaluated against the Attempt 13 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-13/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 13 iteration of the Session Boundary Design (`internal/session/DESIGN.md`) represents a highly mature evolution of Gas City's architecture. Transitioning the "Slice 0" preflight into a hard, schema-validated scheduling gate with comprehensive AST tests, JSON schemas (`command_appliers.schema.json`, etc.), and negative build-failing fixtures is a world-class posture for enforcing system correctness.

However, from the perspective of the **Decider Atomicity Enforcer**, the design document still retains several critical distributed-systems loopholes, physical contradictions, and direct divergence from the active codebase. Pure deciders remain structurally compromised by active local wall-clock reads in `lifecycle_projection.go`. The "Commit runtime start" command continues to license an unfenced compatibility loophole that undermines token-based concurrency isolation. The crash recovery matrix relies on "ghost facts" that cannot survive a process crash. And the critical `instance_token` backfills in the active source remain unprotected, lock-free blind writes.

Until these blockers are resolved in the design and proven in code, decomposition must remain blocked.

---

## Top Strengths

1. **Strict Slice 0 Gatekeeper Posture:** Elevating Slice 0 to a hard scheduling gate and preventing any downstream mutation-owning or behavior-moving beads from being generated until Slice 0 validates successfully prevents pattern drift and ensures uniform guard coverage early.
2. **Comprehensive Schema validation:** Mandating strict schema verification against `command_appliers.schema.json` and `boundary_matrix.schema.json` guarantees that all contract-level artifacts are structured correctly before being consumed by the test suites.
3. **Trace Outcomes Realism:** Demanding that trace outcomes report physical convergence states (such as `durable-scan-converged` or `durable-scan-deferred`) rather than fabricating best-effort event delivery (lines 397-400) enforces a rigorous, realistic distributed systems posture.

---

## Critical Risks & Blockers

### 1. [Blocker] Loophole in "Commit runtime start" licenses Unfenced Commit
* **Evidence:** `design-before.md` line 1174
* **Why it matters:** Despite previous warnings about token fencing, the Command Atomicity Contract table at line 1174 still contains an active compatibility loophole in its primary validation check:
  `re-read prepared bead; commit only when token matches or legacy no-token compatibility path applies`
  This contradiction permits an unfenced commit to proceed under the primary "Commit runtime start" command. A concurrent, older provider process could successfully commit its start without an `instance_token`, blindly overwriting a successor attempt that had already rotated the token and claimed the slot.
* **Required Fix:** Stale, unfenced compatibility paths must be completely expunged from the primary command validation point. No-token identity backfills must be routed strictly through audited, standalone repair flows.

---

### 2. [Blocker] "Ghost Facts" in Crash Recovery Matrix
* **Evidence:** `design-before.md` line 375
* **Why it matters:** The recovery matrix for `Provider start succeeds, commit fails` (line 375) lists the durable facts as `prepared token plus observed runtime identity`, and dictates a recovery of `Retry commit if token still current; otherwise stop or orphan-handle the runtime by token`.
  However, because the commit failed, the "observed runtime identity" was never committed to the database. And because the process crashed, that observation—which lived only in the crashed process's memory—is completely lost. To any surviving reconciler querying the database, the durable facts of this session are indistinguishable from "Crash before provider start" (line 373), which lists the durable facts as `prepared token, no runtime identity` and dictates `Roll back stale creating after grace`.
  On crash-after-start-success-but-before-commit, the reconciler will blindly execute the rollback recovery, leaving the orphaned runtime running on the filesystem. A successor prepare attempt will then yield two live, running agents on the same work directory, leading to catastrophic concurrent writes (Split-Brain).
* **Required Fix:** Redesign the recovery matrix to eliminate reliance on uncommitted, process-local memory. The recovery must instead be able to durably discover the running provider process strictly by the `instance_token` via standard operating system queries (such as scanning command-line arguments or environmental variables) before rollback can occur.

---

### 3. [Blocker] Unprotected, Lock-Free `instance_token` Backfills in `chat.go`
* **Evidence:** `internal/session/chat.go` lines 292-302 and lines 400-409
* **Why it matters:** The design positions `instance_token` as the authoritative identity for start attempts (line 364). However, the active checkout source performs backfill writes to this field outside of any lock, validation, or conditional write guards in `internal/session/chat.go`:
  ```go
  292: 	instanceToken := b.Metadata["instance_token"]
  293: 	if instanceToken == "" {
  294: 		instanceToken = NewInstanceToken()
  295: 		if err := m.store.SetMetadata(id, "instance_token", instanceToken); err != nil {
  296: 			return fmt.Errorf("storing instance token: %w", err)
  297: 		}
  ```
  If two concurrent reconciler loops or command threads race on a bootstrap session that lacks an `instance_token`, both will observe an empty token, generate different tokens, and blindly perform sequential, lock-free `SetMetadata` writes. The second writer overwrites the first, violating the core invariant that `instance_token` uniquely identifies a single start attempt.
* **Required Fix:** These backfills must be wrapped inside `WithSessionMutationLock` or migrated exclusively into the pre-validated `PreWakePatch` batch.

---

### 4. [Blocker] Pure Decider Purity Compromised by Local Clock Reads
* **Evidence:** `internal/session/lifecycle_projection.go` lines 379-382 and lines 607-610
* **Why it matters:** The core architectural principle of pure deciders is that they must consume only immutable facts and perform *zero* clock reads. However, `ProjectLifecycle`—the central, load-bearing decider cited by the design—directly violates this principle in `internal/session/lifecycle_projection.go:379-382` by reading the local machine clock when the input `Now` is zero:
  ```go
  379: 	now := input.Now
  380: 	if now.IsZero() {
  381: 		now = time.Now().UTC()
  382: 	}
  ```
  If a caller omits a validated timestamp (such as passing a zero `Now` fact), the decider yields non-deterministic projection results based on the local machine clock, breaking test determinism, regression replayability, and state-machine predictability.
* **Required Fix:** Remove the clock fallbacks from `lifecycle_projection.go` completely. Make `input.Now` a mandatory, non-zero field, and fail fast if it is missing.
* **AST Guard Clarification:** The design's proposed guard ("Pure session deciders must live in a guardable file set with no imports of... clocks") is structurally impossible to implement. Pure deciders *must* import Go's `time` package to parse and compare `time.Time` fields on input structs. The static guard must instead inspect the AST to reject direct calls to `time.Now()` inside the pure decider set, rather than using blunt package-level import filters.

---

### 5. [Major] Blind Writes and the False Promise of Prevention
* **Evidence:** `design-before.md` lines 350-356 and lines 1104-1108
* **Why it matters:** The design repeatedly claims that a newer token "prevents" or "cannot mutate". Because `beads.Store` and its underlying implementations only support blind writes, lack atomic compare-and-swap (CAS), and do not provide a `Get` method inside the `Tx` transaction callback, this prevention claim is physically false. Any writer can execute `SetMetadata` or `SetMetadataBatch` and blindly overwrite more recent metadata.
* **Impact:** The achievable guarantee is not physical prevention of writes, but rather *fenced stale write detection* and *repair-driven convergence*. If the commit step re-reads and verifies the `instance_token` but then performs a blind write *without reasserting the token as part of the write precondition*, a successor prepare landing between verify and write will result in a corrupted bead with mismatched tokens and active facts.
* **Required Fix:** The design must explicitly specify how stale writes are physically fenced or reconciled under last-writer-wins (e.g., asserting the token inside every written metadata map).

---

### 6. [Major] Missing Close Failure-Point Matrix
* **Evidence:** `design-before.md` lines 371-378 and line 702
* **Why it matters:** While the design provides a detailed 5-row failure-point matrix for `Runtime-start` (lines 371–378), it provides no equivalent failure-point matrix for the `Close` command, leaving the recovery path under-specified.
* **Required Fix:** Provide a detailed failure-point matrix for the Close command equivalent to the Runtime-start matrix, covering crash after stop but before close commit, and stale close intent.

---

## Answers to Persona Questions

### 1. Can each decider run with only immutable facts and no store, runtime, config, or clock reads?
**Answer:** Conceptually yes, but the active codebase violates this. `ProjectLifecycle` and `creatingStateIsStale` still read the system clock directly when the input `Now` is zero. Decider purity must be call-level and absolute, requiring the removal of these clock fallbacks and making the `now` timestamp a mandatory, non-zero field on `LifecycleInput`.

### 2. What compare-and-swap or revalidation happens immediately before mutation commit?
**Answer:** The design specifies re-reading the bead and verifying the `instance_token` match. However, because the underlying `beads.Store` only supports blind writes (no CAS/Tx Get), this revalidation is vulnerable to a TOCTOU race. If a newer prepare lands between the validation read and the write commit, the write will blindly overwrite the newer token. True fencing under last-writer-wins requires either store-level conditional write primitives or explicit repair-driven convergence tests.

### 3. When a decision returns both SessionMutation and RuntimeIntent, what is the exact commit, event emission, runtime execution, and failure ordering?
**Answer:** The design mandates writing the durable prepare marker/token *before* executing the runtime side-effect. If the runtime succeeds but the commit fails, the recovery must detect the discrepancy via the `instance_token` and either retry the commit or stop/orphan-handle the runtime. However, because the "observed runtime identity" is a ghost fact lost on crash, this ordering remains vulnerable to leaving orphaned runtimes during crash recovery unless the discovery mechanism is hardened.

---

## Consistency Report

- **Pattern Alignment:**
  - Strongly aligns with Elena Marchetti's (Mutation Boundary Auditor) requirement for a strict, non-mutating Slice 0 preflight. Both reviews highlight that leaving critical guardrails as pure prose is a severe project risk.
- **Cross-File Integrity:**
  - The `instance_token` field usage described in `DESIGN.md` is correctly reflected in the codebase, but the unprotected backfills in `chat.go` directly break the design's concurrency invariants.
- **Inter-Reviewer Alignment:**
  - Strongly supports the findings of the Mutation Boundary Auditor. We agree that decomposition must be **blocked** until the physical Slice 0 baseline and guard files are checked in and verified passing.
