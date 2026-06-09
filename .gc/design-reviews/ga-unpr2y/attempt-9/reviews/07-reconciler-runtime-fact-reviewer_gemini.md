# Liam Okonkwo — DeepSeek V4 Flash (Independent Review, Attempt 9)

**Verdict:** block

**Review scope:** Reconciler boundary, runtime-intent adapter ownership, fact isolation, and health-gate split. Evaluated against Attempt 9 of `internal/session/DESIGN.md` (which remains identical down to the byte to the Attempt 8 draft), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active codebase (specifically `internal/session/lifecycle_projection.go`, `cmd/gc/compute_awake_set.go`, `cmd/gc/session_reconciler.go`).

---

## Overview

The Attempt 9 revision of `internal/session/DESIGN.md` presents a draft that is identical (exactly 149,654 bytes) to Attempt 8. While the structural taxonomies established previously (including the seven-state eligibility mask and granular `AwakeInput` disposition mapping) remain solid foundations, **none of the critical architectural and boundary blockers identified in the Attempt 8 review have been resolved.**

A thorough audit of the active workspace codebase and the unrevised design draft confirms that the fatal deterministic and boundary violations remain fully active. Specifically, the pure session projection logic still falls back to a non-deterministic wall-clock query via `time.Now()`, pool-level scheduling policy remains leaked into the awake decider via the `countMinActiveCovered` helper, and critical runtime requirements remain completely unmapped in the traceability matrix.

To protect the core primitive boundary and ensure a clean decomposition, **this review must sustain a strict BLOCK.** Decomposing the current state of the design would directly introduce non-deterministic state evaluation and cross-layer coupling into the session primitive layer.

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

## Questions

1. **How will pure-decider import isolation be enforced?** Will the extracted decider live in a separate Go package (e.g., `internal/session/decider/`) so that the Go compiler or a simple directory-import linter can mechanically enforce the isolation boundary?
2. **For `SESSION-WORK-003`, who owns the orphan cleanup scan?** Does it run as a routine inside the reconciler's main loop or on a separate scheduler cadence?
3. **Who owns the `wait-only-wake` exception evaluation?** Does this logic live within the pure eligibility mask calculation or is it composed in Stage 3 by the reconciler?
