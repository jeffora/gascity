# Elena Marchetti — DeepSeek V4 Flash (Independent Review, Attempt 13)

**Verdict:** block

**Review focus:** Mutation boundary enforceability, session lifecycle/identity write ownership, patch-map escape hatches, canonical writer inventory completeness, and build-time static guard coverage. Evaluated against the Attempt 13 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-13/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 13 revision of the Session Boundary Design (`internal/session/DESIGN.md`) continues to mature the engineering boundaries of Gas City. Restricting the immediate next phase of development strictly to a non-mutating, schema-validated "Slice 0" preflight is a premium architectural safeguard. Elevating this preflight with a comprehensive suite of JSON schema paths (`internal/session/testdata/slice0_schemas/`), multi-package validators (`TestSessionBoundaryGuard`, `TestSessionRouteInventoryFresh`, etc.), and negative fixtures is a world-class posture to prevent un-fenced mutations from degrading the codebase.

However, from the perspective of the **Mutation Boundary Auditor**, several critical assumptions and structural risks remain unaddressed in the technical design contracts:

1. **Broad, Un-Decomposed Rows Persist in the Baseline Inventory (`W-026`, `W-029`, `W-030`):** By permitting broad package and file-level rows to remain in the planning baseline under the assumption they will be split "during implementation," the design shifts the hard work of call-site auditing away from the architectural gates.
2. **Unproven Feasibility of AST-Based Interface Receiver Guards:** The proposed static guard test (`cmd/gc/session_boundary_guard_test.go`) is tasked with flagging direct store writes when the receiver is an interface value (`beads.Store`) or a wrapper. In Go, pure AST-based static checkers cannot perform inter-procedural dataflow analysis or type-narrowing to determine if a dynamic string variable or an interface wrapper actually reaches a session bead at runtime. Without a committed prototype or a realistic false-positive budget, this gate is highly speculative.
3. **Local Unit Tests Cannot Enforce "Shrink-Only" and Expiring Exception Constraints:** The "shrink-only" allowlist constraint lacks a physical VCS-aware CI gate or a git-aware test in local checkouts. If a developer appends a new bypass row or pushes an expired date forward, local unit tests run in isolation and will pass, leaving the exception list to grow into a permanent bypass.

Until these structural concerns are directly resolved inside the design contracts, we must sustain a **block** on decomposition.

---

## Top Strengths

1. **Strict Slice 0 Schema Validation:** The addition of concrete JSON Schema definitions under `internal/session/testdata/slice0_schemas/` ensures that all Slice 0 YAML artifacts are validated against a strict, machine-readable contract before any code moves.
2. **Failing-Build Negative Fixtures:** Requiring Slice 0 to supply negative fixtures covering raw store writes, dynamic metadata batches, patch-map extensions, and Huma/legacy API bypasses ensures the validators fail closed on stale, missing, or unowned evidence.
3. **Hard Scheduling Boundary:** Treating Slice 0 as a hard scheduling gate and forbidding the creation of any mutation-owning or behavior-moving beads until the Slice 0 bead is closed is an outstanding project management control.

---

## Critical Risks & Blockers

### 1. [Blocker] Planning Baseline Contains Un-decomposed Broad Rows (`W-026`, `W-029`, `W-030`)
* **Evidence:** `design-before.md` lines 885, 888, 889
* **Pattern Comparison:** Compare with `W-004` (`cmd/gc/cmd_session_wake.go` no-template wake fallback) or `W-012` (`cmd/gc/cmd_prime.go` live-session `session_key` priming), which map to narrow, single concrete symbols.
* **Why it matters:** The design's exit gate requires a source-complete inventory that names every concrete production writer, symbol, and key. However, the planning baseline table uses extremely broad rows. For instance, `W-026` groups several files (`cmd/gc/cmd_session.go`, `cmd/gc/cmd_handoff.go`, etc.) under one row, `W-029` captures "legacy session REST handlers" in `internal/api/handler_sessions.go`, and `W-030` captures "bead update/close/metadata routes that can target session beads" in `internal/api/handler_beads.go`. Leaving these broad rows to be split "during implementation" is a severe risk (Red Flag #1/Red Flag #3) because developers lack an upfront, audited roadmap of what call sites are being converted or retired, allowing un-audited writes to persist.
* **Suggested Fix:** Fully decompose these broad planning rows into separate, symbol-level sub-rows in the design document before approving decomposition. Expand `W-029` to list each individual session handler endpoint (e.g., `/v0/session/{id}/suspend`, `/v0/session/{id}/close`), and list every distinct function in `handler_beads.go` that can write session-bead metadata.

---

### 2. [Blocker] High False-Positive Risk & Unproven Feasibility of AST-Based Interface Receiver Guards
* **Evidence:** `design-before.md` lines 910-930
* **Pattern Comparison:** Compare with `cmd/gc/worker_boundary_import_test.go` (which is a simple import-dependency linter).
* **Why it matters:** The proposed static guard test (`cmd/gc/session_boundary_guard_test.go`) is tasked with "flagging direct store writes when the receiver is a `beads.Store`, a wrapper around a bead store, or an interface value that can carry a session bead" and inspecting dynamic keys. In Go, pure AST-based static checkers cannot perform inter-procedural dataflow analysis or type-narrowing to determine if a dynamic string variable or an interface wrapper (like `beads.Store`) actually reaches a session bead at runtime. Without a committed proof-of-concept or a realistic false-positive budget, this gate is highly speculative. It risks either causing massive build-time noise (flagging every generic bead-store write in the system) or failing silently (missing actual session mutations).
* **Suggested Fix:** Provide a skeletal or prototype implementation of `session_boundary_guard_test.go` that demonstrates how the AST parser resolves interface-receiver types and discriminates session bead targets (e.g., via AST-level type assertion or annotation tagging) before approving this gate.

---

### 3. [Blocker] Local Unit Tests Cannot Enforce "Shrink-Only" and Expiring Exception Constraints
* **Evidence:** `design-before.md` lines 931-933
* **Pattern Comparison:** Standard Go testing framework.
* **Why it matters:** The design requires that `allowlist.yaml` exhibits "shrink-only" behavior and that rows carry a retirement `expiry`. However, a static unit test executed on a local checkout has no temporal memory or access to git history. If a developer simply appends a new exception row or pushes an expired exception's date forward in their local branch, the test will pass because the allowlist is evaluated in isolation. Without a CI-level check comparing the allowlist against the merge-base (e.g., `origin/main`), the "shrink-only" and "expiry" rules are merely prose guidelines, not executable build gates. This directly risks the bypass list growing indefinitely (Red Flag #3).
* **Suggested Fix:** Explicitly require that the Slice 0 preflight includes a CI workflow step or a git-aware test (e.g., invoking `git diff` or checking against a mainline reference) that fails the build if the number of allowlist entries increases or if any entry's `expiry` date is in the past.

---

### 4. [Blocker] The Concurrency "Split-Brain" / Clobbering Vulnerability during Coexistence
* **Evidence:** `design-before.md` lines 330-353 (Command Atomicity) and lines 460-469 (One-Writer Migration).
* **Why it matters:** The design allows "legacy allowed-during-bake" exception rows where both legacy patch-map writers and new command-applier writers coexist. However, legacy writers perform blind writes of entire metadata maps (`SetMetadataBatch`), completely bypassing the new command-applier's tokened prepare/commit/revalidation fence. If a legacy writer executes a blind write concurrently with or after a new command-applier, it will completely overwrite/clobber the newer, fenced state. Fencing only works if *all* active writers participate in the protocol. Having active legacy writers during a "bake" phase introduces a critical, silent concurrency risk (Split-Brain) that can corrupt session beads.
* **Suggested Fix:** Forbid concurrent multi-writer "bake" states for any single owned key family. A key family must transition atomically: either 100% of its production writers are legacy (the old path) or 100% are new (the command-applier path, protected by static guards). Rollback must be binary (switching all routes back to legacy via config/flags), rather than having two active writing paths at the same time.

---

## Answers to Persona Questions

### 1. Which production call sites still write lifecycle or identity metadata directly, and does each backlog slice name the call sites it removes?
**Answer:** The planning baseline table lists 31 writer rows (`W-001` to `W-031`). However, several key rows remain broad package/file level classifications (such as `W-026`, `W-029`, and `W-030`) rather than symbol-level or endpoint-specific call sites. Because these rows are not physically split into symbol-level rows at the design phase, the backlog slices do not yet name the exact call sites they remove. While the design response states that these will be split during implementation, this violates the design's own exit gate requirement that the inventory must be source-complete before any mutation-owning slice starts.

### 2. What failing-build guard prevents new non-session production code from applying session patch maps or SetMetadata lifecycle writes?
**Answer:** The design proposes an additive Go AST/symbol guard (`cmd/gc/session_boundary_guard_test.go`) that scans production roots. It is specified to flag calls to `SetMetadata`, `SetMetadataBatch`, `Update`, `Create`, or `Close` on receivers that are `beads.Store`, wrappers, or interface values that can carry session beads. However, the design still fails to resolve the technical feasibility blocker regarding dynamic type-narrowing on interface receivers. Without a committed prototype showing how the AST-based test will differentiate session-bead mutations from generic store writes, this guard is highly speculative and risks either high false-positive build noise or silent failures.

### 3. How are doctor, migration, repair, and test exceptions bounded so ordinary callers cannot self-label into them?
**Answer:** Exceptions are bounded by a local `allowlist.yaml` file keyed by stable inventory IDs, owner slices, and explicit `expiry` conditions. However, the design specifies evaluating "shrink-only" and "expiry" rules via a local static unit test (`TestSessionBoundaryGuard`). Because local unit tests have no access to VCS/git history or previous states, a developer can easily append new exceptions or push expiry dates forward locally without detection. Without a CI-level or git-aware test (e.g., invoking `git diff` against a mainline reference), the allowlist cannot be enforced as "shrink-only" or expiring, allowing the exception list to grow into a permanent bypass (Red Flag #3).

---

## Consistency Report

* **Pattern Alignment:**
  * Checked the "Shared vocabulary checkpoints" against the active `09-yagni-contract-scope-reviewer_gemini.md` blockers. Agree completely that downstream vocabulary (such as `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent`) must be stripped from the active technical design table and moved to a provisional appendix.
* **Cross-File Integrity:**
  * Verified that the key taxonomy table perfectly aligns with the `LifecycleInput` fields and key families declared in `internal/session/lifecycle_projection.go`.
* **Inter-Reviewer Alignment:**
  * Our findings strongly support the **YAGNI Contract Scope Reviewer** and the **Reconciler Runtime Fact Reviewer**. The lack of a physical, committed Slice 0 prototype is the root cause of these cascading review blocks, proving that the design must remain in an `iterate` / `block` state until concrete fixtures and code files are generated.
