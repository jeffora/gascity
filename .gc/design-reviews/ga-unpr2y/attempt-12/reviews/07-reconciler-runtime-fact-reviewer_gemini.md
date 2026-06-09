# Liam Okonkwo — Gemini (Independent Review, Attempt 12)

**Verdict:** block

**Persona:** Reconciler boundary, runtime intent adapter ownership, fact isolation, health gate split.

**Reviewed against:** `internal/session/DESIGN.md` (Attempt 12, matching `/data/projects/gascity/internal/session/DESIGN.md` / `.gc/design-reviews/ga-unpr2y/attempt-12/design-before.md`), `REQUIREMENTS.md` (45 scenario rows), `internal/session/AGENTS.md` (root and scoped instructions), and the active repository checkout state.

---

## Overview

The Attempt 12 iteration of `internal/session/DESIGN.md` is a highly structured, conceptually robust iteration. Transitioning "Slice 0" from a passive planning concept into a mandatory, non-mutating delivery gate comprising physical symbols, key lists, and static tests is an exceptional defensive step. This is precisely what is needed to lock down the codebase before any mutation-owning changes are allowed to land.

However, from the perspective of the **Reconciler Runtime Fact Reviewer (Liam Okonkwo)**, the Technical Design cannot be approved for decomposition. It still contains critical fact-isolation violations, category errors at the prepare/observe boundary, and un-reconciled policy leaks. The pure session projection continues to fall back to machine wall-clock queries in `lifecycle_projection.go`. The awake-set decider still evaluates pool-level scheduling policy directly in `compute_awake_set.go` via `countMinActiveCovered`. The `RuntimeStartIntent` straddles intent and observation by carrying the post-start `session_key` field at prepare-time. Furthermore, critical operational evidence like W-013 detached probes lacks freshness bounds, supersession rules, and fail-closed timeout definitions.

To safeguard the core session primitive boundary and prevent runtime policy from leaking into `internal/session`, **this review must sustain a strict BLOCK.**

---

## Top Strengths

1. **Exemplary Seven-Kind Eligibility Mask (`DESIGN.md:1265-1277`):**
   The taxonomy of the eligibility mask (separating `forced-runnable`, `conditional-runnable`, `idle-suppressible`, etc.) is excellent. It allows the reconciler to make high-fidelity scheduling and capacity decisions without needing direct read-access to the internal metadata of session beads.

2. **Clean field-level `AwakeInput` disposition mapping (`DESIGN.md:1278-1296`):**
   Explicitly detailing the source, session ownership, and controller/reconciler ownership for each `AwakeInput` field provides a stellar architectural contract that guards against incremental boundary erosion.

3. **Robust Destructive-Action Matrix (`DESIGN.md:1331-1340`):**
   Forcing destructive branches to explicitly consume and validate the completion state of runtime observations (`complete-missing`, `stale-observation`, etc.) aligns perfectly with the GUPP and NDI (Nondeterministic Idempotence) principles.

---

## Critical Risks & Blockers

### 1. [Blocker] Fact Isolation Violation: `ProjectLifecycle` clock fallback inside active codebase
* **Evidence:** `internal/session/lifecycle_projection.go:379-382` and `607-610`
* **Why it matters:** The design mandates that pure session deciders must live in a mechanically guardable set and consume only immutable, already-materialized fact structs (`DESIGN.md:1311-1315`). However, `ProjectLifecycle` (the central, load-bearing projection function) actively reads the system clock when `input.Now` is zero:
  ```go
  now := input.Now
  if now.IsZero() {
  	now = time.Now().UTC()
  }
  ```
  The exact same wall-clock fallback exists in `creatingStateIsStale` (lines 607–610). If a caller omits the timestamp, the projection becomes non-deterministic, destroying regression replayability and violating decider purity.
* **AST Guard Correction:** The design's proposed static guard ("Pure session deciders must live in a guardable file set with no imports of... clocks", lines 1311–1314) is too blunt. Deciders must import `time` to parse, store, and compare `time.Time` values. The static guard must be redefined to inspect the AST to reject direct calls to `time.Now()` inside the pure decider files rather than attempting a blanket package import filter.
* **Required Change:** Remove the `time.Now().UTC()` fallback completely from `lifecycle_projection.go`. Enforce `input.Now` as a mandatory, non-zero field, and fail fast or return an error if it is missing.

---

### 2. [Blocker] Policy Leak: Min-Active Wake evaluation in the awake decider
* **Evidence:** `cmd/gc/compute_awake_set.go` (lines 302, 519, 524)
* **Why it matters:** The design correctly assigns min-active pool coverage policy and capacity calculations to the controller/reconciler layer (`DESIGN.md:1250-1252` and `1282`). However, the active decider implementation in `cmd/gc/compute_awake_set.go` continues to evaluate min-active wake demand directly via the internal `countMinActiveCovered` helper. This leaks pool-level scheduling policy into the awake decider. If `ComputeAwakeSet` is extracted into `internal/session` as-is, the session package will absorb the pool's min-active scheduling rules.
* **Required Change:** Add an explicit refactoring row for `countMinActiveCovered`. Refactor the pool instance counting and min-active threshold evaluation out of the decider. Precompute the min-active protection deficit on the reconciler/controller side and pass it to the eligibility decider as a simple `min-active-deficit` count or as a conditional wake fact.

---

### 3. [Blocker] Prepare vs. Observe Boundary Category Error in `RuntimeStartIntent`
* **Evidence:** `internal/session/DESIGN.md:205-206` (W-022 / W-031 backfills)
* **Why it matters:** The vocabulary contract for `RuntimeStartIntent` requires it to carry the `session_key` field at prepare-time. However, the `session_key` is a dynamic, unique runtime token generated and written during the *observed start* of the session, not during the preparation of the start intent. Forcing the prepare-time intent to carry an observation token that can only be known after the provider successfully launches the session is a category error that violates the separation between intent-preparation and runtime-observation.
* **Required Change:** Split `session_key` out of `RuntimeStartIntent` and define it strictly as an observation fact captured during Stage 1 (assembly) or commit-time. Alternatively, explicitly document the deterministic pre-start runtime-identity generation rule that allows `session_key` to be pre-derived.

---

### 4. [Blocker] W-013 Detached-Probe and Operational Evidence Lacks Freshness/Supersession and Fail-Closed Binding
* **Evidence:** `internal/session/DESIGN.md:814` (`W-013`) and `1323–1327`
* **Why it matters:** The W-013 detached-probe metadata persists observations directly onto session beads. Today, these probes gate destructive orphan release. However, the design does not define any freshness, TTL, or supersession keys (such as an observation timestamp or probe generation) for this persisted metadata. This creates a severe risk where a stale, cached probe observation is treated as current durable truth, leading to incorrect destructive actions. Furthermore, the design fails to specify whether a probe timeout or error must fail closed (skipping release) or fail open.
* **Required Change:** Define a strict freshness/supersession rule (e.g., probe generation or timestamp) for W-013/operational evidence. Explicitly state in the text that detached-probe errors or timeouts must bind to a fail-closed disposition (skipping release) to prevent premature cleanup.

---

### 5. [Blocker] Slices 6 and 7 cannot decompose due to missing proof files in HEAD
* **Evidence:** `internal/session/DESIGN.md:1305–1310`
* **Why it matters:** As self-reported in `DESIGN.md:1305-1310`, `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go` are missing from the current active checkout, making verification of `SESSION-RECON-002`, `-003`, `-006`, and `-007` impossible.
* **Required Change:** Add a non-negotiable exit gate specifying that Slices 5–7 cannot be decomposed until the deleted health, progress, and scale tests are successfully restored to `HEAD`.

---

### 6. [Blocker] Slice 0 Preflight Remains Pure Prose
* **Evidence:** Active checkout state
* **Why it matters:** None of the required Slice 0 preflight files, AST static guards, or test skeletons exist in the active checkout. Leaving this vital defensive boundary as pure prose means we have no physical protection against regression or pattern drift.
* **Required Change:** Complete the physical implementation of Slice 0 preflight files, AST guards, and tests before approving decomposition.

---

## Lane Question Answers

1. **Which wake, hold, drain, provider-health, and progress decisions move into session deciders, and which scheduling or budget responsibilities remain in the reconciler?**
   - **Move into session deciders:** Evaluating transition eligibility (e.g., matching wake causes, hold durations, and drain completions) over already-assembled, immutable facts.
   - **Remain in reconciler:** Aggregating work demand, computing pool desired-counts (including min-active deficit, concurrency limits, and request tiers), tracking container progress, consuming restart/rollback budgets, and orchestrating the actual startup, drain, or shutdown sequences.

2. **Are work counts, pool size, runtime liveness, and progress facts precomputed by adapters instead of queried from deciders?**
   - **Yes.** Pure deciders must have no external I/O or state-query capabilities. Adapters must assemble these facts (e.g., counting assigned work, checking the container engine process table, evaluating health logs) *before* invoking the decider, passing them in as copyable, immutable fields on the input struct.

3. **Can RuntimeIntent express adapter needs without smuggling provider policy into `internal/session`?**
   - **Yes.** `RuntimeIntent` acts as a pure, declarative state change request (e.g., `StartIntent{Alias: "X"}` or `StopIntent{SessionKey: "Y"}`) that is completely decoupled from implementation details. This agnostic intent is then interpreted by the provider runtime adapter (e.g., tmux, subprocess, or Kubernetes) which executes the concrete, provider-specific operations, keeping `internal/session` free of provider leaks.

---

## Consistency & Alignment Report

- **Alignment with Takeshi Yamamoto (Decider Atomicity Enforcer):**
  This review strongly supports Takeshi's finding that `ProjectLifecycle`'s wall-clock fallback breaks the core decider purity invariant. We agree that AST guards must target and forbid `time.Now()` calls rather than attempting a blanket clock import ban.
  
- **Alignment with Ravi Krishnamurthy (Migration Coexistence Strategist):**
  We agree with Ravi that the split-ownership of the wake/hold/drain key family between Slices 2 and 5 must be resolved to make the "one-writer during bake" done criteria satisfiable. We also support the demand that the un-sequenced "repair slice" backfills (W-022 / W-031) be explicitly integrated into Slice 3's atomic change.
