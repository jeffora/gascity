# Liam Okonkwo — DeepSeek V4 Flash (Independent Review, Attempt 7 / Iteration 7)

**Verdict:** block

**Review scope:** Reconciler boundary, runtime-intent adapter ownership, fact isolation, and health-gate split. Evaluated against the Attempt 7 revision of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-7/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and active reconciler code in `cmd/gc/compute_awake_set.go` and `cmd/gc/session_reconciler.go`.

---

## Top Strengths

- **Robust Taxonomy of Observation Completeness**: The formalization of provider liveness observation states (`complete-alive`, `complete-missing`, `stale-observation`, `unknown-provider`, `provider-error`, `partial-query`, and `successor-observed`) is an excellent engineering addition. It directly solves the risk of treating transient probe failures or stale facts as durable confirmation of session termination, ensuring destructive actions (close, drain, rollback) are strictly fenced. This is a great alignment with the Bitter Lesson.
- **Clear Separation of Lifecycle Eligibility vs. Controller Policy**: The design correctly limits the scope of `internal/session` to only evaluating core lifecycle and identity eligibility over immutable facts. By keeping work aggregation, pool capacity calculations, alert-deduplication, restart budgets, and progress/health policy on the controller/reconciler side, the design successfully prevents business-domain scheduling heuristics from leaking into the primitive session layer.
- **Idempotency Over Fragile Event Convergence**: Shifting the critical path of work release and recovery away from fragile, in-process, at-most-once events and instead relying on idempotent, periodic, durable scans of store state (such as work and session beads) ensures reliable convergence under crash-restart scenarios. Events are correctly treated as diagnostic or latency optimizations rather than correctness gates.

---

## Critical Risks & Blockers

### 1. [Blocker] Flat Eligibility Mask Overloads "Runnable," Creating a Fatal Information Gap for Idle-Sleep and Exemptions
The three-stage pipeline contract specifies that Stage 2 (Session decider) receives immutable lifecycle and identity facts and returns an eligibility mask containing flat states (`runnable`, `blocked with reason`, `terminal`, `repair-needed`, or `unknown/partial` in `DESIGN.md:1018`).
However, a session is conceptually "runnable" under two radically different categories:
1. **Forced Runnable**: Pinned (`pin_awake`), attached (`user attached`), pending interaction, or explicit wake. These sessions *must* remain awake regardless of active work demand.
2. **Conditionally Runnable**: Normal sessions. These sessions are eligible to run, but if they have zero active work demand, the reconciler must put them to sleep (idle-sleep).

Because the reconciler (Stage 3) is forbidden from directly reading session-bead metadata keys to prevent boundary leaks, it has no way of knowing *why* a session is `runnable`. If there is zero active work demand, the reconciler will see that the session is `runnable` and has no work, and it will incorrectly put pinned or attached sessions to sleep on the very next tick.

* **Required Change**: Expand the eligibility mask to return richer semantics. Either output distinct categories (e.g., `runnable_forced` with sub-reasons like `pinned`/`attached`/`explicit_wake` versus `runnable_conditional`), or return a structured eligibility record that explicitly bubbles up the session-domain wake overrides so the reconciler can safely evaluate idle-sleep exemptions without piercing the boundary.

---

### 2. [Blocker] Desired-Running Precedence List Directly Violates the Three-Stage Pipeline Separation
Despite establishing the three-stage pipeline under the reconciler contract section, the flat list under `Desired-running precedence must remain explicit` (`DESIGN.md:1065-1075`) remains unaligned and represents a severe pattern drift:
> 3. pending create, pin, attachment, pending interaction, named-always, targeted work, scale demand, and explicit wake produce desired-running when not blocked

This list flatly mixes session-domain eligibility facts (`pin`, `attachment`, `explicit wake`) with reconciler-domain scheduling facts (`targeted work`, `scale demand`). 
This flat structure directly breaks the pipeline's layering:
- If Step 2 (Session decider) evaluates Step 3 of the precedence list, it must import or inspect work demand or scaling counts, violating the import ban on controller demand helpers.
- If Step 3 (Reconciler) evaluates Step 3 of the precedence list, it must inspect pins, attachments, and explicit wake, violating the session-metadata encapsulation boundary.

* **Required Change**: Decompose and realign the Desired-Running Precedence list to match the three-stage pipeline. Explicitly specify which precedence rows are processed in Step 2 (producing pure session-eligibility and wake overrides) and which are processed in Step 3 (producing reconciler-side policy combination and final desired-state).

---

### 3. [Blocker] Seam through `ComputeAwakeSet` Lacks a Field-Level Disposition and Cross-Seam Rule Ownership
The current `AwakeInput` struct in `cmd/gc/compute_awake_set.go` combines fields from all three layers (session, controller demand, and runtime observations). Despite proposing a three-stage pipeline, the design lacks an explicit field-by-field disposition table showing which fields map to Stage 1 (assembly), Stage 2 (session eligibility input), or Stage 3 (reconciler policy).
Furthermore, critical cross-seam suppression rules are unassigned:
- **`attached-pierces-drain`**: An attached user currently bypasses the drained hold state, keeping the session awake. Which layer owns this override?
- **`holds-suppress-attached`**: A hold suppresses user attachments. Is this hold evaluation done at the eligibility mask layer or the reconciler policy composer layer?

Without this detailed disposition, nothing blocks an implementer from moving `ComputeAwakeSet` wholesale into the session package (bringing along `ScaleCheckCounts` and `WorkBeads` and violating the import ban), or leaving session-bead metadata reads spread across the reconciler.

* **Required Change**: Add a comprehensive field-level disposition table for `AwakeInput` mapping every field to its respective stage of the pipeline. Designate clear architectural owners (mask vs. reconciler composer) for the cross-seam rules (`attached-pierces-drain`, `holds-suppress-attached`, and `wait-only-wake` exceptions). Add this disposition as a blocking requirement for Slice 5 and 6 implementation.

---

### 4. [Blocker] Designed-Against Provider-Health and Progress Code is Missing from this Checkout
As noted in the prior review iteration, the workspace checkout's version of `cmd/gc/session_reconciler.go` lacks the provider-health and progress gate implementations that landed upstream (e.g., `b5a7f3be3`, `dbda1e380`). The design continues to refer to files like `cmd/gc/provider_health_gate_test.go` and `cmd/gc/session_progress_test.go` as "absent in this checkout" (DESIGN.md:805, 806).
The current unblock instruction ("Restore or replace provider-health/progress proof," DESIGN.md:806, 1118) is dangerously passive. It invites an implementer to re-derive health/progress logic from REQUIREMENTS prose, which will inevitably fork the boundary from the actual upstream implementation.

* **Required Change**: Re-scope the Slice 7 unblock condition to explicitly forbid re-deriving the health/progress gate from prose. Instead, require the implementation team to audit the provider-health/progress implementation on the pinned upstream baseline, import the actual code and tests into this workspace, and rebase the workspace onto that baseline before proceeding with the extraction.

---

### 5. [Blocker] Idle-Sleep Ownership and Decision/Mutation Split Remain Unstated in Tables
Prose responses indicate that the reconciler owns "idle-sleep decisions" and timers, but the main layers table (DESIGN.md:345) and the state-ownership table (DESIGN.md:1058) omit idle entirely. Today, the reconciler decides idle and applies the mutation directly via `CompleteDrainPatch` (bypassing the mutation boundary). The design lacks an explicit contract detailing how this split is resolved.

* **Required Change**: Add an explicit idle-sleep row to the session eligibility vs. controller demand layers table and the state-ownership table. Specify that the reconciler owns the idle decision and timers, while the session command owns the resulting sleep/drain mutation, ensuring the legacy `CompleteDrainPatch` is correctly retired.

---

## Required Changes

1. **Richer Eligibility Mask Semantics**: Modify the eligibility mask output to distinguish between "conditionally runnable" sessions (subject to idle-sleep) and "forced runnable" sessions (exempt from idle-sleep due to pins, user attachments, or pending interactions).
2. **Decompose the Precedence List**: Split the flat desired-running precedence table into Step 2 (session eligibility and overrides) and Step 3 (reconciler policy combination) to avoid coupling leaks.
3. **`AwakeInput` Seam Field-Disposition**: Add a field-disposition table mapping every `AwakeInput` field to its pipeline stage, and explicitly assign ownership of the cross-seam overrides.
4. **Re-scope Slice 7 Baseline Recovery**: Replace the passive "Restore or replace proof" gate with an active requirement to audit, import, and rebase onto the actual upstream provider-health and progress implementations.
5. **Codify Idle-Sleep Split**: Add explicit rows for idle-sleep decisions to the state-ownership and layering tables, detailing the separation of reconciler-side decision and session-side mutation.

---

## Questions

1. **How should global configuration facts like `ChatIdleTimeout` be modeled?** Does this timeout belong to session eligibility (since it is a manual session configuration) or does the reconciler-side idle-sleep timer check it, and is it passed as an input to the decider?
2. **How should `repair-needed` sessions be treated by reconciler scheduling?** If a session is runnable but marked `repair-needed`, does the reconciler suppress all work scheduling, or can best-effort display/interactions proceed during the repair phase?
3. **Who owns the `wait-only-wake` exception?** Does this rule (where a wait-hold is bypassed under specific conditions) live in the session eligibility decider, or is it evaluated during the reconciler's policy composition phase?
