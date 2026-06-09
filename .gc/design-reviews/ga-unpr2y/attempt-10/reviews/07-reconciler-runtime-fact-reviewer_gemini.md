# Liam Okonkwo — Gemini (Independent Review, Attempt 10)

**Verdict:** block

**Review scope:** Reconciler boundary, runtime-intent adapter ownership, fact isolation, and health-gate split. Evaluated against the Attempt 10 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-10/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 10 iteration of the Session Boundary Design (`internal/session/DESIGN.md`) remains byte-identical (exactly 149,654 bytes) to the Attempt 8 draft. While the structural taxonomies (such as the seven-state eligibility mask, the detailed `AwakeInput` fields ownership boundaries, and the robust `work_release_pending` scanner recovery loop) are exceptionally strong architectural foundations, **none of the critical reconciler boundary and fact isolation blockers have been resolved.**

A thorough audit of the active workspace codebase and the unrevised design draft confirms that the fatal deterministic and boundary violations remain fully active. Specifically, the pure session projection logic still falls back to a non-deterministic wall-clock query via `time.Now()` inside `lifecycle_projection.go`, pool-level scheduling policy remains leaked into the awake decider via the `countMinActiveCovered` helper in `compute_awake_set.go`, and critical runtime/reconciler requirements remain completely unmapped in the traceability matrix.

To protect the core primitive boundary and ensure a clean, deterministic decomposition, **this review must sustain a strict BLOCK.** Decomposing the current state of the design would directly introduce non-deterministic state evaluation and cross-layer coupling into the session primitive layer.

---

## Top Strengths

1. **Elegant Eligibility Mask Taxonomy:** The seven-kind eligibility mask (`forced-runnable`, `idle-suppressible`, etc.) remains a stellar conceptual design. It allows the reconciler to make high-fidelity scheduling decisions without requiring direct read access to encapsulated session-bead metadata.
2. **Field-Level `AwakeInput` Disposition Table:** Defining the distinct ownership boundaries for each field of `AwakeInput` is an exceptional blueprint for preventing future coupling.
3. **Idempotence-First Convergence Mechanics:** Grounding liveness monitoring and orphan release in periodic, durable database scans rather than fleeting in-memory events beautifully respects the GUPP and NDI (Nondeterministic Idempotence) principles.
4. **Resilient Runtime Observation States:** The taxonomy for liveness queries (`complete-alive`, `stale-observation`, etc.) ensures robust transient-fault handling, protecting against split-brain scenarios.

---

## Critical Risks & Blockers

### 1. [Blocker] Fact Isolation Violation: `ProjectLifecycle` Fallback to `time.Now()` Breaks Decider Determinism
The design document requires that pure session deciders live in a mechanically guardable file set with no clock imports and may only consume already-materialized, immutable structs (`DESIGN.md:1211-1215`).

However, the canonical read-model baseline `internal/session/lifecycle_projection.go` at lines 381 and 609 still falls back to `time.Now().UTC()` when `input.Now` is zero:
```go
now := input.Now
if now.IsZero() {
	now = time.Now().UTC()
}
```
This introduces non-deterministic wall-clock side-effects directly into the projection. If a pure decider calls `ProjectLifecycle`, its execution is no longer deterministic, violating the core "pure decider" invariant.

* **Required Change:** Modify `ProjectLifecycle` to strictly reject zero-valued `input.Now` (e.g., returning an error or failing closed). The caller (the reconciler/adapter) must be the sole source of the wall-clock fact. Completely remove the `time.Now().UTC()` fallbacks from `lifecycle_projection.go`.

### 2. [Blocker] Min-Active Wake Leak: Inside-Decider `countMinActiveCovered` Violates Controller Policy Assignment
The design assigns min-active session policy and pool capacity rules to the controller/reconciler layer (`DESIGN.md:1182`, line 1250).

However, the active decider implementation in `cmd/gc/compute_awake_set.go` (lines 302, 519, 524) still evaluates min-active wake demand directly via the internal `countMinActiveCovered` helper. This leaks pool-level scheduling policy into the awake decider. If `ComputeAwakeSet` is moved into `internal/session` as a pure decider, the session package will absorb the pool's min-active scheduling policy. If it is split, the design lacks a row indicating which layer owns the `countMinActiveCovered` logic.

* **Required Change:** Add an explicit row to the `AwakeInput` disposition table mapping `MinActiveSessions`. Specify that the min-active pool coverage calculation must be precomputed on the controller/reconciler side and passed to the decider as a simple `conditional-runnable` demand fact, or as an explicit `min-active-deficit` count, rather than letting the session decider count pool instances and evaluate min-active thresholds directly.

### 3. [Blocker] Critical Reconciler & Runtime Requirements Left Unmapped in Scenario Traceability Matrix
A behavior-parity audit of the Scenario Traceability Matrix reveals that critical reconciler and runtime boundary requirements defined in `REQUIREMENTS.md` have been completely omitted:
- **`SESSION-RECON-001` (Worker boundary):** Enforces that all production code routes through `worker.Handle`.
- **`SESSION-WORK-003` (Orphan pool step beads):** Essential for reconciler cleanup and boundary safety.
- **`SESSION-RUNTIME-004` (Stop turn):** Dictates provider stop-turn behavior.

Omitting these from the Traceability Matrix means no slice is actively accountable for writing and proving tests for these behaviors during extraction.

* **Required Change:** Assign these critical boundary requirements to specific slices in the Traceability Matrix (e.g., `SESSION-RECON-001` to Slice 1 or 4, `SESSION-WORK-003` to Slice 4, and `SESSION-RUNTIME-004` to Slice 5 or 7).

### 4. [Blocker] Prepare vs. Observe Boundary Category Error: `RuntimeStartIntent` Straddles the Boundary
The vocabulary contract for `RuntimeStartIntent` requires it to carry the `session_key` field at prepare-time.

However, the `session_key` is a dynamic, unique runtime token generated and written during the *observed start* of the session, not during the preparation of the start intent. Forcing the prepare-time intent to carry an observation token that can only be known after the provider successfully launches the session is a category error that violates the separation between intent-preparation and runtime-observation.

* **Required Change:** Split `session_key` out of `RuntimeStartIntent` and define it strictly as an observation fact captured during Stage 1 (assembly) or commit-time. Alternatively, explicitly document the deterministic pre-start runtime-identity generation rule that allows `session_key` to be pre-derived.

---

## Major & Minor Risks

### [Major] Durable Observation Evidence (W-013) Lacks Freshness/Supersession and Fail-Closed Binding
The W-013 detached-probe metadata persists observations directly onto session beads. Today, these probes gate destructive orphan release. However, the design does not define any freshness, TTL, or supersession keys (such as an observation timestamp or probe generation) for this persisted metadata. This creates a severe risk where a stale, cached probe observation is treated as current durable truth, leading to incorrect destructive actions. 

Furthermore, the design fails to specify whether a probe timeout or error must fail closed (skipping release) or fail open.

* **Required Change:** Define a strict freshness/supersession rule (e.g., probe generation or timestamp) for W-013/operational evidence. Explicitly state in the text that detached-probe errors or timeouts must bind to a fail-closed disposition (skipping release) to prevent premature cleanup.

### [Major] Omitted Production Writers in the Canonical Writer Inventory
The writer inventory table fails to individually register and assign key production files that actively write session-owned keys:
- `cmd/gc/cmd_stop.go` (writes `sleep_reason`)
- `cmd/gc/cmd_wait.go` (writes `wait_hold`, `sleep_intent`, `closed_at`, `close_reason`)

Leaving these active writers unmapped in the inventory means they cannot be bound to shrink-only rules or explicit retirement plans, leaving a significant bypass window open in the mutation boundary.

* **Required Change:** Add granular, dedicated writer IDs (e.g., `W-015`, `W-016`) to the inventory table for the specific call sites in `cmd_stop.go` and `cmd_wait.go`, and assign them to their respective retirement slices.

---

## Missing Evidence

- **Fact Isolation Test Enforcer:** No named test or static-analysis tool (like an AST import-guard) is specified to mechanically enforce the pure decider import ban.
- **Missing Baseline Tests:** As self-reported, `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go` are missing from the current active checkout, making verification of `SESSION-RECON-002`, `-003`, `-006`, and `-007` impossible.

---

## Required Changes

1. **Purify Projection Clocks:** Remove the `time.Now().UTC()` fallbacks from `lifecycle_projection.go`. Enforce that any zero-valued `input.Now` results in an error or failed-closed evaluation.
2. **Move Min-Active Counting to Reconciler:** Refactor the min-active pool coverage check out of the awake decider. Define it as a reconciler/controller-level demand policy, passing the deficit to the decider.
3. **Map the 9 Omitted Requirements:** Fully integrate `SESSION-RECON-001`, `SESSION-WORK-003`, `SESSION-RUNTIME-004`, and the other 6 unmapped requirements into the Scenario Traceability Matrix.
4. **Fix `RuntimeStartIntent` Signature:** Remove `session_key` from the prepare intent, or specify the deterministic pre-start derivation rule.
5. **Add Supersession to Durable Evidence:** Document the freshness/TTL rule for W-013 metadata, and bind probe error/timeout to fail-closed release.
6. **Incorporate Missing Writers:** Map `cmd_stop.go` and `cmd_wait.go` into the Canonical Production Writer Inventory with clear retirement targets.
7. **Strict Block on Slices 5–7:** Add a non-negotiable exit gate specifying that Slices 5–7 cannot be decomposed until the deleted health, progress, and scale tests are successfully restored to `HEAD`.

---

## Answers to Persona Questions

### 1. Which wake, hold, drain, provider-health, and progress decisions move into session deciders, and which scheduling or budget responsibilities remain in the reconciler?
**Answer:** Pure decisions regarding session state transitions (e.g., wake cause matching, hold evaluation, drain eligibility, and transition to closed) move completely into pure, deterministic session deciders in `internal/session`. However, scheduling responsibilities (such as counting min-active deficits, managing concurrency limits or pool capacities, rate limits, and scheduling retries) along with overall coordination budget rules must remain strictly within the controller/reconciler layer to keep session logic decoupled from execution policy.

### 2. Are work counts, pool size, runtime liveness, and progress facts precomputed by adapters instead of queried from deciders?
**Answer:** Yes. Pure deciders must be supplied with all necessary external scheduling context as precomputed, immutable facts inside `AwakeInput` or `SessionFacts`. Reconciler adapters/readers are responsible for querying active processes, tracking the process table, measuring container liveness, counting current workloads, and calculating pool deficits *before* passing these facts to the decider, ensuring deciders make no external calls or bead-store queries.

### 3. Can RuntimeIntent express adapter needs without smuggling provider policy into internal/session?
**Answer:** Yes. `RuntimeIntent` acts as a pure, declarative state change request (e.g., `StartIntent{Alias: "X"}` or `StopIntent{SessionKey: "Y"}`) that is completely decoupled from implementation details. This agnostic intent is then interpreted by the provider runtime adapter (e.g., tmux, subprocess, or Kubernetes) which executes the concrete, provider-specific operations, keeping `internal/session` free of provider leaks.

---

## Consistency Report

- **Inter-Reviewer Alignment:**
  Our conclusions align perfectly with the **Mutation Boundary Auditor (Elena Marchetti)** and the **Event Delivery Contract Reviewer (Amara Osei)**. The lack of physical, committed Slice 0 code files and test skeletons in the repository is a shared blocker that prevents both the static mutation guards and the event recovery tests from being verified.
- **Cross-File Integrity:**
  - The proposed `work_release_pending` metadata state and release identity keys must be integrated into `internal/session/lifecycle_projection.go` once Slice 0 lands, ensuring they are not classified as unknown fields.
