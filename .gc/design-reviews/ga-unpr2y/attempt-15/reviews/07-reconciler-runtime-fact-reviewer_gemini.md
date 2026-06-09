# Liam Okonkwo — DeepSeek V4 Flash (Attempt 15 Review)

**Verdict:** block

**Review scope:** Reconciler boundary, runtime intent adapter ownership, fact isolation, health gate split, and design-to-code alignment for the Reconciler Runtime Fact Reviewer mandate. Evaluated against the Attempt 15 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-15/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 15 streamline of `./internal/session/DESIGN.md` is a massive step forward in architectural maturity. By replacing over-engineered, highly speculative, 1200+ line specifications with a focused, minimalist 472-line refactoring blueprint, the design successfully honors the YAGNI (You're Not Gonna Need It) principle. Deferring the complex mutating operations (wake, close, retire, drain, runtime start) to subsequent slices and gating them behind strict, operation-specific **Atomic Command Contracts** (lines 231-259) is excellent engineering.

However, from the perspective of the **Reconciler Runtime Fact Reviewer (Liam Okonkwo)**, this design cannot be approved for decomposition. It still retains critical fact-isolation violations, physical guard gaps, and baseline test deficits. In particular, the core lifecycle projection continues to violate decider purity by reading the local clock, and we lack the necessary physical Slice 0 artifacts and baseline reconciler tests to prevent regression.

Therefore, **this review must sustain a strict BLOCK.**

---

## Top Strengths

1. **Robust, Low-Coupling Reconciler/Session Split Matrix (lines 289–302):**
   The split matrix is exceptionally well-defined. It correctly restricts `internal/session` to pure lifecycle eligibility and identity classification over immutable facts, while preserving the controller/reconciler's ownership of policy and demand (work demand, dispatch scheduling, pool sizing, cold-start demand, restart budgets, and progress policy). This prevents structural policy leaks.
2. **Minimalist, Deferral-First Backlog Sequence (lines 439–460):**
   Prioritizing non-mutating Slice 0 preflights and side-effect-free Slice 1 read-only target classification before touching any mutating command or reconciler fact-moving path is a highly disciplined, TDD-friendly progression.
3. **Streamlined Destructive-Action Boundary (lines 309–312):**
   Forcing destructive actions with unknown, stale, partial, or provider-error runtime facts to be rejected unless explicitly covered by a safe convergence rule and test-proven prevents catastrophic out-of-sync or premature teardowns.

---

## Critical Risks & Blockers

### 1. [Blocker] Fact Isolation Compromised: Load-bearing `ProjectLifecycle` clock fallback inside active codebase
* **Evidence:** `internal/session/lifecycle_projection.go:381` and `609`
* **Why it matters:** The design mandates that pure session deciders must live in a mechanically guardable set and consume only immutable, copyable facts. However, `ProjectLifecycle` (the central, load-bearing projection function) still contains local wall-clock fallbacks when `input.Now` is zero:
  ```go
  now := input.Now
  if now.IsZero() {
  	now = time.Now().UTC()
  }
  ```
  If a caller fails to pass an explicit timestamp, the decider yields non-deterministic results based on the local OS clock, destroying test replayability and violating decider purity. Fact isolation must be call-level and absolute.
* **Required Change:** Completely remove the `time.Now().UTC()` fallbacks from `lifecycle_projection.go`. Enforce `input.Now` as a mandatory, non-zero field, returning an error or failing fast if it is missing. Redefine any static AST guards to inspect pure decider files to reject direct calls to `time.Now()`.

### 2. [Blocker] Slice 0 Preflight Remains Pure Prose
* **Evidence:** Active checkout state
* **Why it matters:** None of the physical Slice 0 preflight files (`BOUNDARY_INVENTORY.md`, `allowlist.yaml`, `SCENARIO_PARITY.yaml`, etc.), AST static guards, or test skeletons exist in the active checkout yet. Leaving this critical defense boundary as pure prose provides no physical protection against pattern drift or regressions as work proceeds.
* **Required Change:** Complete the physical implementation of Slice 0 preflight files, AST guards, and tests before approving decomposition.

### 3. [Major] Prepare vs. Observe Boundary on `session_key` in `RuntimeStartIntent`
* **Evidence:** `internal/session/DESIGN.md:303–307`
* **Why it matters:** The design states that provider-neutral runtime intent fields "may include... runtime session key" (which is `session_key`) (lines 303-305). However, the `session_key` is a dynamic, observer-driven runtime token generated during the *observed start* of the session, not during the preparation of the start intent. Carrying this observation token inside a prepare-time intent is a category error that violates the separation between intent-preparation and runtime-observation.
* **Required Change:** Explicitly state in the design that `session_key` must remain strictly an observation fact captured during or after start, unless a deterministic caller-side pre-generation formula is defined.

### 4. [Major] Reconciler-side Baseline Test Gaps for Slices 5 and 6
* **Evidence:** `internal/session/DESIGN.md:451–460` and active checkout state
* **Why it matters:** The active checkout still lacks the deleted reconciler-side provider-health, scale-from-zero, and progress tests (e.g. `scale_from_zero_test.go`, `provider_health_gate_test.go`, and `session_progress_test.go`) cited by `REQUIREMENTS.md`. Slices 5 (Runtime Start) and 6 (Reconciler Facts) cannot safely decompose or proceed without these baselines restored to HEAD.
* **Required Change:** Add an explicit, non-negotiable exit gate specifying that Slices 5 and 6 cannot decompose until the missing provider-health, progress, and scale-from-zero tests are restored to `cmd/gc/`.

---

## Missing Evidence

- **Pure Decider AST Guard Test:** An automated AST parser test in `internal/session` that scans the pure-decider files and fails the build if any direct call to `time.Now()` or store-query patterns are found.
- **Provider-Health/Scale Baseline Proof:** Restored or replacement test files for `scale_from_zero_test.go` and `provider_health_gate_test.go` in `cmd/gc/`.
- **Deterministic `session_key` Generation Spec:** Detailed documentation on whether the `session_key` is pre-derived deterministically (such as via a secure random generation in the preparer) or captured solely as a post-start runtime observation.

---

## Required Changes

1. **Enforce Absolute Decider Purity:** Remove `time.Now().UTC()` clock fallbacks from `lifecycle_projection.go`, making `input.Now` mandatory and non-zero.
2. **Materialize Slice 0 Artifacts:** Physical creation and integration of all Slice 0 preflight artifacts and tests in the codebase.
3. **Clean Prepare vs. Observe Boundary:** Decouple `session_key` from prep-time start intent fields, or explicitly specify its deterministic pre-derivation rule.
4. **Restore Reconciler Baseline Tests:** Restore the deleted reconciler tests (`scale_from_zero_test.go`, `provider_health_gate_test.go`, etc.) to HEAD before starting Slices 5 and 6.

---

## Answers to Persona Questions

### 1. Which wake, hold, drain, provider-health, and progress decisions move into session deciders, and which scheduling or budget responsibilities remain in the reconciler?
**Answer:**
- **Move to session deciders:** Determining basic transition eligibility (such as wake blockers, terminal states, configured identity conflicts, and hold/drain timeouts) based on immutable input facts.
- **Remain in reconciler:** Aggregating work demand, computing desired pool counts, handling cold-start scaling, tracking provider health, progress policy, restart/rollback budgets, and orchestrating destructive actions.

### 2. Are work counts, pool size, runtime liveness, and progress facts precomputed by adapters instead of queried from deciders?
**Answer:** Yes. Pure session deciders perform zero store queries or I/O. All required operational facts—such as active work counts, pool sizing, and observed runtime liveness—must be pre-scanned and compiled by callers/adapters and passed to the decider via copyable, immutable structures.

### 3. Can RuntimeIntent express adapter needs without smuggling provider policy into `internal/session`?
**Answer:** Yes. `RuntimeIntent` is a pure declarative state representation (e.g. specifying `session_id`, deterministic `session_key`, config hash, etc.) that specifies *what* is intended. The runtime provider adapters (tmux, subprocess, k8s) consume this intent and translate it into provider-specific actions and policies, keeping `internal/session` completely free of provider leakages.

---

## Consistency Report

* **Pattern Alignment:**
  * **Decider Atomicity Enforcer (Takeshi Yamamoto):** We completely align with Takeshi's finding that the local clock fallback in `ProjectLifecycle` violates pure-decider invariants. We agree that AST guards are necessary to enforce decider purity without restricting valid `time.Time` helper uses.
  * **Event Delivery Contract Reviewer (Amara Osei):** We align with Amara's approval of the streamlined Event Contract. Keeping events as pure post-commit facts is consistent with our demand that post-commit observations remain strictly decoupled from prep-time intents.
