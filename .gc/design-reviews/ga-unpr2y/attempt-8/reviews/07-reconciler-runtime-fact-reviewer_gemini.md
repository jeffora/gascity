# Liam Okonkwo — DeepSeek V4 Flash (Independent Review, Attempt 8)

**Verdict:** block

**Review scope:** Reconciler boundary, runtime-intent adapter ownership, fact isolation, and health-gate split. Evaluated against Attempt 8 of `internal/session/DESIGN.md`, `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active codebase (specifically `internal/session/lifecycle_projection.go`, `cmd/gc/compute_awake_set.go`, `cmd/gc/session_reconciler.go`).

---

## Overview

The Attempt 8 revision of `internal/session/DESIGN.md` makes outstanding structural and conceptual strides in clarifying the boundaries between the core session primitives and the controller/reconciler scheduler. Key improvements, such as the introduction of a seven-kind eligibility mask (`forced-runnable`, `conditional-runnable`, `idle-suppressible`, etc.) and the granular, field-by-field `AwakeInput` disposition table, directly address the fatal information gaps identified in prior reviews.

However, from the strict, first-principles perspective of the Reconciler and Fact-Isolation lane, there remain several deep architectural contradictions, boundary category errors, and direct violations of the "pure decider" invariant that must be resolved before approval can be granted. 

Most critically, the current baseline projection library (`lifecycle_projection.go`) directly violates the pure-decider contract by querying the wall-clock via `time.Now().UTC()` when `input.Now` is zero, rather than failing closed or strictly enforcing input parameter completeness. Additionally, the min-active sessions wake computation remains embedded in the "pure" awake decider despite the design explicitly assigning min-active policy to the controller/reconciler layer.

Until these boundaries are cleaned up and the unmapped requirements are integrated into the Scenario Traceability Matrix, the design cannot be safely decomposed.

---

## Top Strengths

1. **Richer Eligibility Mask Semantics:** Replacing a flat `runnable` bit with a seven-state mask (including `forced-runnable`, `conditional-runnable`, and `idle-suppressible`) elegantly solves the idle-sleep and user-attachment exemption gap. It allows the reconciler to evaluate scheduling policy without having to leak session-bead metadata reads or pierce encapsulated states.
2. **Comprehensive field-level `AwakeInput` disposition table:** Explicitly mapping each field of the `AwakeInput` struct to its respective owner and source provides a high-fidelity blueprint for developers. This ensures that the eventual extraction maintains perfect separation of concerns.
3. **Robust Runtime Observation Completeness Rules:** The seven-state taxonomy for liveness queries (`complete-alive`, `complete-missing`, `stale-observation`, etc.) ensures that transient network hiccups or rate limits do not become durable session truth, guarding against destructive split-brain actions.
4. **Idempotence-First Recovery Contract:** Grounding critical work release and drain safety in periodic, durable database scans rather than fleeting, in-process events is an exceptional application of the GUPP and NDI (Nondeterministic Idempotence) principles.

---

## Critical Risks & Blockers

### 1. [Blocker] Fact Isolation Violation: `ProjectLifecycle` Fallback to `time.Now()` Breaks Decider Determinism
The design document states that pure session deciders must live in a mechanically guardable file set with no imports of clocks and that deciders may only consume already-materialized, immutable structs (`DESIGN.md:1211-1215`).

However, the canonical read-model baseline `internal/session/lifecycle_projection.go` contains the following implementation at lines 379–382 and 608–610:
```go
now := input.Now
if now.IsZero() {
    now = time.Now().UTC()
}
```
This is a direct violation of the pure decider contract. If `input.Now` is omitted or zero, falling back to a non-deterministic wall-clock query introduces side effects and breaks unit-test reproducibility. If the projection is called in a pure decider context, it is no longer pure.

* **Required Change:** Modify `ProjectLifecycle` to strictly reject zero-valued `input.Now` or fail closed. The caller (the reconciler/adapter) must be the sole source of the wall-clock fact. Remove the `time.Now().UTC()` fallbacks from `lifecycle_projection.go`.

### 2. [Blocker] Min-Active Wake Leak: Inside-Decider `countMinActiveCovered` Violates Controller Policy Assignment
The design assigns min-active session policy and pool capacity rules to the controller/reconciler layer (`DESIGN.md:1182`, line 1250: *"pool demand... is controller/reconciler policy; not a session primitive"*).

However, the active decider `cmd/gc/compute_awake_set.go` implements the `countMinActiveCovered` helper internally (lines 287–316, 524–546) and evaluates min-active wake demand directly. If `ComputeAwakeSet` is moved into `internal/session` as a pure decider, the session package will absorb the pool's min-active scheduling policy. If `ComputeAwakeSet` is split, the design lacks a row indicating which layer owns the `countMinActiveCovered` logic and how the resulting `min-active` wake demand is carried.

* **Required Change:** Add an explicit row to the `AwakeInput` disposition table mapping `MinActiveSessions`. Specify that the min-active pool coverage calculation must be precomputed on the controller/reconciler side and passed to the decider as a simple `conditional-runnable` demand fact, or as an explicit `min-active-deficit` count, rather than letting the session decider count pool instances and evaluate min-active thresholds directly.

### 3. [Blocker] Critical Reconciler & Runtime Requirements Left Unmapped in Scenario Traceability Matrix
A behavior-parity audit of the Scenario Traceability Matrix (`DESIGN.md:913–922`) reveals that several critical reconciler and runtime boundary requirements defined in `REQUIREMENTS.md` have been completely left out of the implementation slices:
- **`SESSION-RECON-001` (Worker boundary):** Essential for enforcing that all production code routes through `worker.Handle`.
- **`SESSION-WORK-003` (Orphan pool step beads):** Critical for reconciler cleanup and boundary safety.
- **`SESSION-RUNTIME-004` (Stop turn):** Dictates provider stop-turn behavior.

Omitting these from the Traceability Matrix means no slice is actively accountable for writing and proving tests for these behaviors during extraction.

* **Required Change:** Assign these critical boundary requirements to specific slices in the Traceability Matrix (e.g., `SESSION-RECON-001` to Slice 1 or 4, `SESSION-WORK-003` to Slice 4, and `SESSION-RUNTIME-004` to Slice 5 or 7).

### 4. [Blocker] Prepare vs. Observe Boundary Category Error: `RuntimeStartIntent` Straddles the Boundary
The vocabulary contract for `RuntimeStartIntent` (`DESIGN.md:649`) requires it to carry the `session_key` field at prepare-time. 

However, the `session_key` is a dynamic, unique runtime token generated and written during the *observed start* of the session, not during the preparation of the start intent. Forcing the prepare-time intent to carry an observation token that can only be known after the provider successfully launches the session is a category error that violates the separation between intent-preparation and runtime-observation.

* **Required Change:** Split `session_key` out of `RuntimeStartIntent` and define it strictly as an observation fact captured during Stage 1 (assembly) or commit-time. Alternatively, explicitly document the deterministic pre-start runtime-identity generation rule that allows `session_key` to be pre-derived.

---

## Major & Minor Risks

### [Major] Durable Observation Evidence (W-013) Lacks Freshness/Supersession and Fail-Closed Binding
The W-013 detached-probe metadata persists observations directly onto session beads. Today, these probes gate destructive orphan release. However, the design does not define any freshness, TTL, or supersession keys (such as an observation timestamp or probe generation) for this persisted metadata. This creates a severe risk where a stale, cached probe observation is treated as current durable truth, leading to incorrect destructive actions. 

Furthermore, the design fails to specify whether a probe timeout or error must fail closed (skipping release) or fail open.

* **Required Change:** Define a strict freshness/supersession rule (e.g., probe generation or timestamp) for W-013/operational evidence. Explicitly state in the text that detached-probe errors or timeouts must bind to a fail-closed disposition (skipping release) to prevent premature cleanup.

### [Major] Omitted Production Writers in the Canonical Writer Inventory
The writer inventory table (`DESIGN.md:700–733`) fails to individually register and assign key production files that actively write session-owned keys:
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
