# Takeshi Yamamoto — DeepSeek V4 Flash (Independent Review, Attempt 12)

**Verdict:** block

**Review focus:** Pure deciders, optimistic concurrency, commit-event-intent ordering, stale fact defense — with direct evidence from the active codebase. Evaluated against the Attempt 12 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-12/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 12 iteration of `internal/session/DESIGN.md` maintains a robust and highly disciplined posture. Elevating the non-mutating "Slice 0" preflight to a hard scheduling entry gate and formalizing a clear, multi-model write-safety taxonomy are superb engineering steps.

However, from the perspective of the **Decider Atomicity Enforcer**, the design continues to contain critical distributed-systems contradictions, unsafe architectural assumptions, and direct misalignments with the active checkout. Pure deciders are compromised by active wall-clock reads in the projection layer. The "Commit runtime start" command licenses an unfenced compatibility loophole that directly undermines token-based isolation. The recovery matrix relies on "ghost facts" that do not survive a process crash. And the critical `instance_token` backfills in the active source remain unprotected, lock-free blind writes.

Until these blockers are resolved in the design and proven in code, decomposition must remain blocked.

---

## Top Strengths

1. **Durable Scans as the Decisive Backstop:** Shifting the critical correctness obligation from BEST-EFFORT in-process events (lines 1130–1135) to reconciler-driven, idempotent scans over durable session facts is an excellent, distributed-systems posture.
2. **Explicit Write-Safety Taxonomy:** Formally categorizing store mutations (conditional, tokened blind, and inert-when-lost writes, lines 345–351) provides developers with a clear and disciplined taxonomy for reasoning about store mutation boundaries on last-writer-wins backends.
3. **Authoritative Attempt Identity:** Elevating `instance_token` over simple monotonic integer sequences or timestamps as the primary and authoritative identifier for runtime-start attempts prevents sequencing drift, overlap, and collision across asynchronous reconciler sweeps.

---

## Critical Risks & Blockers

### 1. [Blocker] Pure Decider Purity Compromised by Local Clock Reads in active codebase
The core architectural principle of **Zero Framework Cognition (ZFC)** and pure deciders is that deciders must consume only immutable facts and perform *zero* clock reads or side effects.

However, `ProjectLifecycle`—the central, load-bearing decider cited by the design—directly violates this principle in `internal/session/lifecycle_projection.go:379–382` by reading the local machine clock when the input `Now` is zero:
```go
379: 	now := input.Now
380: 	if now.IsZero() {
381: 		now = time.Now().UTC()
382: 	}
```
The exact same clock fallback exists inside `creatingStateIsStale` at `internal/session/lifecycle_projection.go:607–610`:
```go
607: 	now := input.Now
608: 	if now.IsZero() {
609: 		now = time.Now().UTC()
610: 	}
```

* **Impact:** If a caller omits a validated timestamp (such as passing a zero `Now` fact), the decider yields non-deterministic projection results (such as blocker, timeout, or quarantine durations) based on the local machine clock. This breaks test determinism, regression replayability, and state-machine predictability. Decider purity must be call-level and absolute.
* **AST Guard Clarification:** The design's proposed guard ("Pure session deciders must live in a guardable file set with no imports of... clocks", lines 1311–1314) is structurally impossible to implement. Pure deciders *must* import Go's `time` package to parse, store, and compare `time.Time` fields on input structs. The static guard must instead inspect the AST to reject direct calls to `time.Now()` inside the pure decider set, rather than using blunt package-level import filters.
* **Required Fix:** Remove the clock fallbacks from `lifecycle_projection.go` completely. Make `input.Now` a mandatory, non-zero field, and fail fast if it is missing.

---

### 2. [Blocker] "Ghost Facts" in Crash Recovery Matrix
The recovery table on lines 371–378 contains a critical distributed-systems contradiction. 

For the "Provider start succeeds, commit fails" row (line 375), the design lists the durable facts as `prepared token plus observed runtime identity`, and dictates a recovery of `Retry commit if token still current; otherwise stop or orphan-handle the runtime by token`. 

However, because the metadata commit failed, the "observed runtime identity" was never committed to the database. And because the process crashed, that observation—which lived only in the crashed process's memory—is completely lost. To any surviving reconciler querying the database, the durable facts of this session are indistinguishable from "Crash before provider start" (line 373), which lists the durable facts as `prepared token, no runtime identity` and dictates `Roll back stale creating after grace`.

* **Impact:** On crash-after-start-success-but-before-commit, the reconciler will blindly execute the rollback recovery, leaving the orphaned runtime running on the filesystem. A successor prepare attempt will then yield two live, running agents on the same work directory, leading to catastrophic concurrent writes.
* **Required Fix:** Redesign the recovery matrix. The system must write a persistent `provider_started` or similar marker *before* executing the side effect, or must be able to deterministically discover the running provider process strictly by the `instance_token` via standard operating system queries (e.g., matching env/command line arguments).

---

### 3. [Blocker] Unprotected, Lock-Free `instance_token` Backfills in `chat.go`
The design positions `instance_token` as the authoritative identity for start attempts (line 1062). However, the active checkout source performs backfill writes to this field outside of any lock, validation, or conditional write guards in `internal/session/chat.go:292–302` and `chat.go:400–409`:
```go
292: 	instanceToken := b.Metadata["instance_token"]
293: 	if instanceToken == "" {
294: 		instanceToken = NewInstanceToken()
295: 		if err := m.store.SetMetadata(id, "instance_token", instanceToken); err != nil {
296: 			return fmt.Errorf("storing instance token: %w", err)
297: 		}
...
```

* **Impact:** If two concurrent reconciler loops or command threads race on a bootstrap session that lacks an `instance_token`, both will observe an empty token, generate different tokens, and blindly perform sequential, lock-free `SetMetadata` writes. The second writer overwrites the first, violating the core invariant that `instance_token` uniquely identifies a single start attempt.
* **Required Fix:** These backfills must be wrapped inside `WithSessionMutationLock` or migrated exclusively into the pre-validated `PreWakePatch` batch.

---

### 4. [Blocker] Loophole in "Commit runtime start" licenses Unfenced Commit
The Attempt 5 loophole closure (lines 1077–1082) states that legacy no-token backfills are repair scenarios, not alternate command paths.

However, the actual **Command Atomicity Contract table** at line 1116 still contains an active compatibility escape hatch in the primary command validation point:
```markdown
re-read prepared bead; commit only when token matches or legacy no-token compatibility path applies
```

* **Impact:** This contradiction permits an unfenced commit to proceed under the primary "Commit runtime start" command. A concurrent, older provider process could successfully commit its start without an `instance_token`, blindly overwriting a successor attempt that had already rotated the token and claimed the slot.
* **Required Fix:** Stale, unfenced compatibility paths must be completely expunged from primary command validation points.

---

### 5. [Major] Blind Writes and the False Promise of Prevention
The design repeatedly claims that a newer token "prevents" or "cannot mutate" (lines 377, 1103–1104). Because `beads.Store` and its underlying implementations (like `BdStore` or `exec.Store`) only support blind writes, lack atomic compare-and-swap (CAS), and do not provide a `Get` method inside the `Tx` transaction callback, this prevention claim is physically false. Any writer can execute `SetMetadata` or `SetMetadataBatch` and blindly overwrite more recent metadata.

* **Impact:** The achievable guarantee is not physical prevention of writes, but rather *fenced stale write detection* and *repair-driven convergence*. The design under-specifies this for `Prepare runtime start` and `Commit runtime start`. If the commit step re-reads and verifies the `instance_token` but then performs a blind write *without reasserting the token as part of the write precondition*, a successor prepare landing between verify and write will result in a corrupted bead with mismatched tokens and active facts.
* **Required Fix:** The design must explicitly specify how stale writes are physically fenced or reconciled under last-writer-wins.

---

### 6. [Major] Missing Close Failure-Point Matrix
While the design provides a detailed 5-row failure-point matrix for `Runtime-start` (lines 371–378), it provides no equivalent failure-point matrix for the `Close` command.

* **Impact:** There is no specification for recovering from a process crash or store failure that occurs after stopping the provider but before committing the closed status to the bead. Without a formal matrix covering crash before stop, stop failure, stop success plus close commit failure, and stale close intent, the close recovery path remains under-specified and prone to leaving orphaned runtimes or stranded assigned work.
* **Required Fix:** Provide a detailed failure-point matrix for the Close command equivalent to the Runtime-start matrix.

---

### 7. [Blocker] Slice 0 Preflight Remains Pure Prose
None of the required Slice 0 preflight artifacts (including `internal/session/BOUNDARY_INVENTORY.md`, `allowlist.yaml`, `SCENARIO_PARITY.yaml`, and static AST guards) physically exist in the active checkout.

* **Impact:** The static protection against regression is non-existent. Implementation beads cannot proceed until Slice 0 is physically implemented and verified passing in the checkout.
* **Required Fix:** Complete the physical implementation of Slice 0 preflight files, AST guards, and tests before approving decomposition.

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
  - Aligns with Elena Marchetti's (Mutation Boundary Auditor) requirement for a strict, non-mutating Slice 0 preflight. Both reviews highlight that leaving critical guardrails as pure prose is a severe project risk.
- **Cross-File Integrity:**
  - The `instance_token` field usage described in `DESIGN.md` is correctly reflected in the codebase, but the unprotected backfills in `chat.go` directly break the design's concurrency invariants.
- **Inter-Reviewer Alignment:**
  - Strongly supports the findings of the Mutation Boundary Auditor. We agree that decomposition must be **blocked** until the physical Slice 0 baseline and guard files are checked in and verified passing.
