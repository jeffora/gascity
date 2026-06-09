# Ravi Krishnamurthy — DeepSeek V4 Flash (Independent Review, Attempt 12)

**Verdict:** block

**Persona:** migration sequencing, legacy-new coexistence, rollback slices, worker-boundary collision, cross-document consistency.

**Reviewed against:** `internal/session/DESIGN.md` (Attempt 12, matching `/data/projects/gascity/internal/session/DESIGN.md` / `.gc/design-reviews/ga-unpr2y/attempt-12/design-before.md`), `REQUIREMENTS.md` (45 scenario rows), `internal/session/AGENTS.md` (root and scoped instructions), and the active repository checkout state.

---

## Overview

The Attempt 12 iteration of `internal/session/DESIGN.md` maintains a robust and disciplined engineering posture. Transitioning the "Slice 0" preflight from passive prose to a mandatory, non-mutating delivery package—complete with physical YAML inventories, a static AST-based guard, and a build-time test suite—is a masterclass in Technical Debt Defense. It locks down the "before" state of the codebase before any mutation-owning changes can slide into production.

Additionally, Attempt 12's inclusion of **"Data-direction convergence proof"** in the response table represents a highly sophisticated distributed-systems mindset, ensuring that legacy code safely tolerates and recovers over metadata written by the new paths during a mid-bake revert.

However, from the strict perspective of **migration sequencing and coexistence safety**, the Technical Design still suffers from critical logical contradictions, unaddressed boundary collisions, and un-sequenced "ghost" slices. The "done-criteria" for Slice 2 and Slice 3 remain mathematically unsatisfiable under the document's own invariants, and key API/CLI exception routes contain structural escape hatches that could lead to overlapping boundaries.

Until these sequencing and coexistence defects are repaired in the normative text, this design must remain in the `iterate` stage.

---

## Top Strengths

1. **Rigorous Materialization of Slice 0 Deliverables (`DESIGN.md:127–140`):**
   By forcing the physical creation and commit of `SLICE0_BASELINE.md`, `BOUNDARY_INVENTORY.md`, `SESSION_BOUNDARY_SYMBOLS.yaml`, `SCENARIO_PARITY.yaml`, and `VOCABULARY_CHECKPOINTS.yaml` before any mutation-owning slice can begin, the design prevents "document-only" passes and provides a compiler-verifiable baseline.

2. **Durable-Scan Convergence Recovery (`DESIGN.md:53`, `78`, `588`):**
   Subordinating in-process event delivery to best-effort diagnostic aids, and placing the entire burden of critical recovery on durable reconciler-driven fact scans, is an exceptionally robust application of NDI (Nondeterministic Idempotence).

3. **Coexistence Contract and Token Schema (`DESIGN.md:442–458`):**
   Requiring that every implementation bead carry twelve distinct `session_design.*` metadata fields ensures that workflow gates can dynamically verify migration safety and translates abstract architectural safety into a concrete CI gate.

---

## Critical Risks & Blockers

### 1. [Blocker] Reconciler-level `session_key` clears bypass Slice 3 validation during the bake window
* **Evidence:** `internal/session/DESIGN.md:810` (`W-009`)
* **Why it matters:** The design assigns `session_reconcile.go` (`W-009`) to Slices 6 and 7. However, the reconciler loop performs raw `SetMetadata` clears on `session_key` during unexpected death handling and session churn. Meanwhile, Slice 3 (`W-005` and `W-012`, "runtime-identity setter/clearer and continuation-reset paths") is the sole designated owner of the `session_key` family. If Slice 3 is implemented and enters its bake window with a "no parallel writer" constraint, the reconciler's raw writes on `session_key` will bypass validation, creating a **split-brain write window** and violating the "one-writer per key family during bake" invariant.
* **Required Change:** Explicitly state that `W-005`/`W-012` governs over the file-level exception of `W-009`. Reconciler-level `session_key` clears must be fenced and migrated into Slice 3 (or a narrow, shared validation helper in `internal/session` called by both) before Slice 3 can be declared complete.

---

### 2. [Blocker] Wake/hold/drain key family is split-owned across Slices 2 and 5, making one-writer done criteria unsatisfiable
* **Evidence:** `internal/session/DESIGN.md:774` and `1411`
* **Why it matters:** The key family for wake, hold, and drain is grouped as a single family: `[wake_request, wake_requested_at, wake_mode, last_woke_at, pin_awake, held_until, quarantined_until, wait_hold, sleep_intent, drain_*, handoff_*, assigned_work_*]`. The coexistence rule states: *"No slice can be marked done while old and new writers mutate the same owned key with different validation."* However, W-004 converts in **Slice 2**, whereas W-006 and W-007 convert in **Slice 5**. Because Slices 2 and 5 share ownership of this family, and Slice 2 is implemented first, old and new writers remain active concurrently on this key family, making Slice 2's one-writer done-criterion mathematically unsatisfiable. Furthermore, the wake applier has no durable fence token to prevent races.
* **Required Change:** Sub-split the wake/hold/drain row in the mutation-boundary and key-family tables into distinct per-key families with unique slice ownership: e.g., Slice 2 owns `wake_request`, `wake_requested_at`, and wake blockers; Slice 5 owns `drain_*`, `pin_awake`, `held_until`, sleep/wait-hold keys. Assign `W-011`'s nudge keys explicitly. Define unique revert rules and name the wake-side fence or convergence argument covering Slice 2's bake window.

---

### 3. [Blocker] Slice 3's no-parallel-writer condition depends on an unscheduled "repair slice," forcing an implicit flag-day
* **Evidence:** `internal/session/DESIGN.md:820`, `823`, `825`, `826`, `828`, `830`, `832`
* **Why it matters:** Several safety-critical runtime identity and continuation backfills (like `W-022`, `W-031`, and the lock-free `instance_token` backfill in `chat.go:292–302`) are assigned to "repair slice" or "Slice 3, repair slice". Yet, the "repair slice" has no position, owner, or proof set in the backlog (Slices 0-7). Because these backfills continue to write directly to `instance_token` and `session_key`, Slice 3's "no parallel writer" condition is unsatisfiable unless a massive, unplanned flag-day occurs where the entire repair slice is implemented simultaneously, violating the gradual delegation model.
* **Required Change:** Formally schedule the "repair slice" in the backlog sequence with position, owner, and proof set—or fold the `W-012`/`W-022`/`W-031` backfill conversion entirely into Slice 3's atomic change. Reconcile this with the six-phase model.

---

### 4. [Major] Shared call-site plan for `session_manager.go` / `session_resolution.go` lacks coordination with worker-boundary ledger
* **Evidence:** `internal/session/DESIGN.md:1431–1438`
* **Why it matters:** Both this migration and the root worker-boundary migration are attempting to retire or narrow `session_manager.go` and `session_resolution.go`. However, the design still contains legacy route options like "an approved session command factory" (W-015/W-025) without mechanical enforcement or coordination rules. This dual ownership and lack of clear boundary default is a direct recipe for boundary collision and split-brain routing.
* **Required Change:** Reconcile W-014/W-015/W-016/W-025/W-029 and the shared call-site plan with the worker-boundary ledger. Require worker-boundary migration notes to be updated in the same change that retires a shared exception, and name which document is authoritative. Limit command factories to the API layer, enforcing `worker.Handle` as the sole production entry point for CLI session mutations.

---

### 5. [Major] Data-direction rollback remains prose without metadata representation
* **Evidence:** `internal/session/DESIGN.md:447–458`
* **Why it matters:** The design's response table mentions "Data-direction convergence proof", but this was never propagated to the twelve-field metadata schema at lines 447–458, nor to the per-slice coexistence gates. If a slice is reverted mid-bake, legacy readers encountering unexpected metadata fields (like stale prepare tokens, phase markers, or transitional close facts) may crash or misinterpret state due to lack of passive tolerance or scrubbing procedures.
* **Required Change:** Add `session_design.data_direction_proof` to the required metadata list in `DESIGN.md:442–458`. Add legacy-over-new-data convergence tests or a documented manual repair/key-scrubbing script (e.g., `gc rollback-repair --slice=3`) to each slice's bake/revert rule before bake begins.

---

## Lane Question Answers

1. **How does the plan sequence this extraction with the in-flight worker-boundary migration on overlapping `cmd/gc` and `internal/api` call sites?**
   - The migrations are sequenced as orthogonal, additive boundaries. The worker-boundary migration mandates execution routing (CLI/API → `worker.Handle`), while the mutation-boundary migration enforces metadata write safety. A call site is only considered "migrated" when both direct `session.Manager` creation imports are eliminated and direct raw `beads.Store` metadata writes are replaced with validating commands. However, the shared call-site plan for `session_manager.go` and `session_resolution.go` must be reconciled with the worker-boundary ledger to prevent overlapping retirement authority.

2. **During partial adoption, what prevents legacy patch-map callers and new command callers from split-brain writes to the same metadata fields?**
   - The combination of: (a) the per-key owner matrix which defines the single command slice owning each metadata family, (b) the shrink-only allowlist enforced by the build-failing static guard, and (c) the mandatory "rollback contract" attached to each implementation bead. Any un-inventoried direct write or double-write will immediately fail the build. However, as noted above, this protection is compromised by reconciler-level bypasses on `session_key` (W-009) and unprotected backfills (W-022/W-031), which must be resolved.

3. **Which single slice is independently shippable and revertible, and what proves it does not silently require the next slice?**
   - Slice 1 (Target Classification) is completely independent and safe. Because it acts as a read-only resolver that wraps the existing target oracle, it introduces zero metadata or write-path changes. Reverting Slice 1 requires only a routing-switch toggle back to the old resolver path, with zero data-direction risk.

---

## Questions for the Author

1. Who owns the "repair slice," and does it land before, with, or after Slice 3?
2. Can Slice 2 close while reconciler, waits, and nudge wake-writers remain live? If yes, exactly which keys does Slice 2 own at close?
3. Is the runtime-start family conversion accepted as a flag day, and what is the agreed maximum revert blast radius for Slice 3?
4. Which document is authoritative for the API exception end-route — this design or the root worker-boundary migration ledger — and what mechanism keeps them synchronized?
5. After a mid-bake revert of any slice, what heals new-path metadata (stale prepare tokens, phase markers, transitional close facts) under legacy code, and which tests prove it?
6. Does the AST static guard run as part of the local pre-commit hooks (`.githooks/pre-commit`) or solely in the remote CI environment? Running it locally is highly recommended.
