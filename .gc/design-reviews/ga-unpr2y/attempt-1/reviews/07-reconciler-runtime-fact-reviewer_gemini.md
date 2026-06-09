# Liam Okonkwo — DeepSeek V4 Flash (Attempt 16 Review)

**Verdict:** block

**Review scope:** Reconciler boundary, runtime intent adapter ownership, fact isolation, health gate split, and design-to-code alignment for the Reconciler Runtime Fact Reviewer mandate. Evaluated against the Attempt 16 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-16/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 16 revision of `DESIGN.md` delivers significant architectural improvements, especially on the migration and coexistence fronts. Adding the `Migration Coexistence And Rollback` section (lines 430–451) is a major step forward in addressing multi-process write safety on shared API/CLI/reconciler files. Furthermore, defining an explicit row-schema for `BOUNDARY_MATRIX.yaml` (lines 477–488)—including policy owner, destructive-action safety, and wake-cause production owner—directly addresses the core requirements of this persona.

However, from the perspective of the **Reconciler Runtime Fact Reviewer (Liam Okonkwo)**, several structural risks, boundary leaks, and missing safeguards remain. In particular, the core lifecycle projection continues to violate decider purity with direct local wall-clock reads. We also find that the preflight requirements of Slice 0 introduce a premature dependency on reconciler-owned requirements (`SESSION-RECON-*`), which couples the session boundary refactor to reconciler-side state too early.

Therefore, **this review sustains a strict BLOCK.**

---

## Top Strengths

1. **Explicit Schema for `BOUNDARY_MATRIX.yaml` (lines 477–488):**
   Requiring each row in the boundary matrix to explicitly track its source owner, policy owner, allowed session inputs/outputs, freshness rules, and destructive-action safety rules prevents the split from being treated as pure prose. This provides a clear, auditable structure for the reconciler/session boundary.
2. **Robust Multi-Writer Migration Fencing (lines 430–451):**
   The coexistence design is incredibly thorough. Mandating a cross-process fence (such as store-native CAS or token preconditions) if old and new writers coexist for a key family prevents split-brain writes on overlapping cmd/gc and internal/api paths.
3. **Structured Event-to-Fact Projection (lines 496–536):**
   Defining session events as post-commit facts (rather than commands or locks) and establishing standard event inventories (like `session.woke` and `session.stranded`) with typed subscriber-data expectations protects the system from distributed race loops.

---

## Critical Risks & Blockers

### 1. [Blocker] Fact Isolation Compromised: System Wall-Clock Reads inside `ProjectLifecycle` and `creatingStateIsStale`
* **Evidence:** `internal/session/lifecycle_projection.go` lines 380–382 and lines 607–610:
  ```go
  now := input.Now
  if now.IsZero() {
  	now = time.Now().UTC()
  }
  ```
* **Why it matters:** Pure session deciders must be completely deterministic and free of system side-effects. By falling back to `time.Now().UTC()` when `input.Now` is zero, the projection yields non-deterministic results based on the local OS clock, breaking test repeatability, trace auditability, and making state-machine behavior unpredictable.
* **Required Change:** Completely remove the `time.Now().UTC()` fallbacks from `lifecycle_projection.go`. Enforce `input.Now` as a mandatory, non-zero field in `LifecycleInput`. Fail fast or return an error if it is missing, and establish an AST guard to prevent future direct clock reads in `internal/session`.

### 2. [Blocker] Slice 0 Preflight Remains Pure Prose
* **Evidence:** Active checkout state
* **Why it matters:** The physical Slice 0 artifacts (`BOUNDARY_MATRIX.yaml`, `BOUNDARY_INVENTORY.md`, `SCENARIO_PARITY.yaml`, etc.) and the AST static guards required to validate decider purity do not yet exist on disk. Without these physical controls and guards actively integrated into the CI pipeline, there is no protection against pattern drift.
* **Required Change:** Complete the physical implementation of the Slice 0 preflight artifacts, AST guards, and tests as active CI gates before approving decomposition.

### 3. [Major] Smuggling `session_key` as Prepare-time Intent in `RuntimeStartIntent`
* **Evidence:** `DESIGN.md` lines 467–471
* **Why it matters:** The design permits `RuntimeStartIntent` to carry `runtime session key` (which is `session_key`). However, `session_key` is a dynamic, observer-driven runtime token generated *during or after the start* of the session (e.g., by the tmux provider socket or process PID scan), not during the preparation of the start intent. Carrying this observation token inside a prepare-time intent is a category error that violates the separation between intent-preparation and runtime-observation.
* **Required Change:** Decouple `session_key` from the prepare-time intent fields, or explicitly specify its deterministic pre-derivation rule (e.g., pre-generating a secure random token in the preparer) to preserve the boundary.

### 4. [Major] Scope Creep: Reconciler-Owned Requirements Blocking Session Slice 0
* **Evidence:** `DESIGN.md` lines 173–179
* **Why it matters:** Slice 0 is forced to repair or owner-retire the requirement evidence for `SESSION-RECON-002` (Cold pool scale from zero), `SESSION-RECON-003` (Existing rig session prevents cold wake), `SESSION-RECON-006` (Provider health gate), and `SESSION-RECON-007` (Progress-aware health) before any subsequent session slice can proceed. However, pool scaling, provider health gates, and progress thresholding are reconciler/pool adapter behaviors that the `Boundary Matrix` places *outside* `internal/session`. Forcing the Session Slice 0 to validate reconciler-specific requirement evidence is a premature scope-coupling risk that halts session progress due to reconciler test gaps.
* **Required Change:** Remove the requirement to resolve `SESSION-RECON-*` evidence from the Session Slice 0 entry criteria, moving that burden to the pool/reconciler-specific slices (Slices 5 and 6) where they belong.

### 5. [Major] Concurrency Hazard: Unfenced "Runtime-Missing" Observation to Durable Write
* **Evidence:** `DESIGN.md` lines 473–475 and 489–494
* **Why it matters:** When the runtime provider fails to find a session (e.g., due to a slow VM boot or transient process-query error), this "missing" observation is translated into a durable session state write (e.g., `sleep_reason="runtime-missing"`). Without an explicit, time-bounded revalidation or provider-error fence, transient infrastructure hiccups are elevated into durable truth, leading to premature restarts or unintended cleanups.
* **Required Change:** Explicitly require in `BOUNDARY_MATRIX.yaml` that "runtime-missing" observations must survive a time-bounded quarantine or a multi-pass check before being committed as durable session truth.

---

## Missing Evidence

- **Pure Decider AST/Import Guard:** An automated test scanning the `internal/session` package to reject direct calls to `time.Now()` or any imports from `cmd/gc` or store query helpers.
- **Physical `BOUNDARY_MATRIX.yaml` Schema and Rows:** The machine-readable boundary matrix declaring owner, freshness, unknown/stale/partial/provider-error policy, and destructive-action safety.
- **Restored Reconciler-side Baseline Tests:** Missing test files for `scale_from_zero_test.go` and `provider_health_gate_test.go` in `cmd/gc/` to serve as a baseline before starting Slices 5 and 6.

---

## Required Changes

1. **Enforce Pure-Decider Clock Purity:** Remove `time.Now().UTC()` clock fallbacks from `lifecycle_projection.go`, making `input.Now` mandatory and non-zero.
2. **Refactor Slice 0 Entry Gate:** Move `SESSION-RECON-*` requirements from the Session Slice 0 entry criteria to reconciler-specific slices.
3. **Establish "Runtime-Missing" Observation Fences:** Require a multi-pass check or quarantine window in `BOUNDARY_MATRIX.yaml` before writing a transient missing observation as durable session truth.
4. **Decouple `session_key` from Intent:** Explicitly specify how the `session_key` is generated (deterministic pre-derivation vs post-start observation) to resolve the prepare vs observe category error.
5. **Materialize Slice 0 Preflight on Disk:** Physically create and validate the required preflight artifacts before approving decomposition.

---

## Answers to Persona Questions

### 1. Which wake, hold, drain, provider-health, and progress decisions move into session deciders, and which scheduling or budget responsibilities remain in the reconciler?
**Answer:**
- **Move to session deciders:** Evaluating basic transition eligibility (such as wake blockers, terminal states, configured identity conflicts, and hold/drain timeouts) based on immutable input facts.
- **Remain in the reconciler:** Aggregating work demand, computing desired pool counts, handling cold-start scaling, tracking provider health, progress policy, restart/rollback budgets, and orchestrating destructive actions.

### 2. Are work counts, pool size, runtime liveness, and progress facts precomputed by adapters instead of queried from deciders?
**Answer:** Yes. Pure session deciders perform zero store queries, network, or filesystem I/O. All operational facts—such as active work counts, pool sizing, and observed runtime liveness—must be pre-scanned and compiled by the reconciler/adapters and passed to the decider via copyable, immutable structures like `LifecycleInput`.

### 3. Can RuntimeIntent express adapter needs without smuggling provider policy into internal/session?
**Answer:** Yes. `RuntimeIntent` is a pure declarative state representation (e.g., specifying `session_id`, deterministic `session_key`, config hash, etc.) that defines *what* is intended. The runtime provider adapters (tmux, subprocess, k8s) consume this intent and translate it into provider-specific actions and policies, keeping `internal/session` completely free of provider leakages.

---

## Consistency Report

* **Pattern Alignment:**
  * **Decider Atomicity Enforcer (Takeshi Yamamoto):** We completely align with Takeshi's finding that the local clock fallback in `ProjectLifecycle` violates pure-decider invariants. We also support his blocker regarding the single-process limitation of `WithSessionMutationLock` (Go's `sync.Mutex` does not prevent CLI-to-daemon races).
  * **Wake Eligibility & Health Gates Code-Sharing:** Currently, wake eligibility and provider health gates share helper code in `cmd/gc/session_reconciler.go`. As we extract the session state-machine rules, this shared helper code must be cleanly split. We agree with Takeshi that the boundary matrix must explicitly separate these concerns to avoid messy, circular helper imports.
  * **Dual Close-Path Divergence:** We support Takeshi's demand to unify `closeBead` and `CloseDetailed` into a single session-owned close command, preventing closed beads from retaining orphaned wake/hold overrides in the database.
