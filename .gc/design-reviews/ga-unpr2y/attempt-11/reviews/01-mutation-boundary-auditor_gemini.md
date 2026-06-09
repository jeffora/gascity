# Elena Marchetti — DeepSeek V4 Flash (Independent Review, Attempt 11)

**Verdict:** block

**Review focus:** Mutation boundary enforceability, session lifecycle/identity write ownership, patch-map escape hatches, canonical writer inventory completeness, and build-time static guard coverage. Evaluated against the Attempt 11 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-11/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 11 revision of the Session Boundary Design (`internal/session/DESIGN.md`) continues to establish valuable and robust engineering boundaries. Restricting the immediate next phase of implementation strictly to a non-mutating "Slice 0" preflight is a premium architectural guardrail. It prevents premature and un-fenced mutation code from reaching production before the baseline guards and scenario matrices are established.

However, from the perspective of the **Mutation Boundary Auditor**, several critical assumptions and structural risks remain unaddressed in the technical design contracts. The baseline writer inventory still depends on broad, un-decomposed file and package-level rows (`W-026`, `W-029`, `W-030`), shifting the hard task of call-site auditing to the implementation phase. Furthermore, the feasibility of static AST type-narrowing over generic interface receivers (such as `beads.Store`) is unproven and highly speculative. Finally, the "shrink-only" allowlist constraint lacks a physical VCS-aware CI gate to prevent silent local drift.

Until these structural flaws are directly resolved inside the design contracts, we must sustain a **block** on decomposition.

---

## Top Strengths

1. **Strict Slice 0 Implementation Boundary:** Elevating Slice 0 to a strict, non-mutating preflight is an excellent architectural control. It guarantees that static guards, scenario maps, and baseline transcripts are committed and verified before a single line of production behavior is changed.
2. **Explicit Verification Command:** Pinned verification via the `go test ./cmd/gc ./internal/session` suite with a strict "zero-match is a failure" rule ensures that no guard or test is introduced as an inert placeholder.
3. **Comprehensive Key Taxonomy Mapping:** Mapping session-owned mutation fields into detailed taxonomy families (such as Create/start lease, Runtime identity, and Wake/hold/drain) provides developers with a clear and comprehensive list of restricted keys.

---

## Critical Risks & Blockers

### 1. [Blocker] Planning Baseline Baseline Contains Un-decomposed Broad Rows (`W-026`, `W-029`, `W-030`)
* **Evidence:** `design-before.md` lines 827, 830, 831
* **Pattern Comparison:** `W-004`, `W-005` (which map to single concrete command targets).
* **Why it matters:** The design's exit gate requires a source-complete inventory that names every concrete production writer, symbol, and key. However, the planning baseline table uses extremely broad rows. For instance, `W-029` captures "legacy session REST handlers" in a single broad row, and `W-030` captures "all bead routes that can target session beads" as a generic entry. Leaving these broad rows to be split "during implementation" is a severe risk because developers lack a concrete, audited roadmap of what call sites are being converted or retired in each slice, allowing un-audited writes to persist.
* **Suggested Fix:** Fully decompose these broad planning rows into separate, symbol-level sub-rows in the design document before approving decomposition. Specifically, expand `W-029` to list each individual session handler endpoint (e.g., `/v0/session/{id}/suspend`, `/v0/session/{id}/close`), and list every distinct function in `handler_beads.go` that can write session-bead metadata.

---

### 2. [Blocker] High False-Positive Risk & Unproven Feasibility of AST-Based Interface Receiver Guards
* **Evidence:** `design-before.md` lines 834-891
* **Pattern Comparison:** `cmd/gc/worker_boundary_import_test.go` (which is a simple import-dependency linter).
* **Why it matters:** The proposed static guard test (`cmd/gc/session_boundary_guard_test.go`) is tasked with "flagging direct store writes when the receiver is an interface value that can carry a session bead" and inspecting dynamic keys. In Go, pure AST-based static checkers cannot perform inter-procedural dataflow analysis or type-narrowing to determine if a dynamic string variable or an interface wrapper (like `beads.Store`) actually reaches a session bead at runtime. Without a committed proof-of-concept or a realistic false-positive budget, this gate is highly speculative. It risks either causing massive build-time noise (flagging every generic bead-store write in the system) or failing silently (missing actual session mutations).
* **Suggested Fix:** Provide a skeletal or prototype implementation of `session_boundary_guard_test.go` that demonstrates how the AST parser resolves interface-receiver types and discriminates session bead targets (e.g., via AST-level type assertion or annotation tagging) before approving this gate.

---

### 3. [Blocker] Local Unit Tests Cannot Enforce "Shrink-Only" and Expiring Exception Constraints
* **Evidence:** `design-before.md` lines 873-875
* **Pattern Comparison:** Standard Go testing framework.
* **Why it matters:** The design requires that `allowlist.yaml` exhibits "shrink-only" behavior and that rows carry a retirement `expiry`. However, a static unit test (`TestSessionBoundaryGuard`) executed on a local checkout has no temporal memory or access to git history. If a developer simply appends a new exception row or pushes an expired exception's date forward in their local branch, the test will pass because the allowlist is evaluated in isolation. Without a CI-level check comparing the allowlist against the merge-base (e.g., `origin/main`), the "shrink-only" and "expiry" rules are merely prose guidelines, not executable build gates. This directly risks the bypass list growing indefinitely (Red Flag #3).
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

- **Pattern Alignment:**
  - Checked the "Shared vocabulary checkpoints" against the active `09-yagni-contract-scope-reviewer_gemini.md` blockers. Agree completely with Kwame Asante that downstream vocabulary (such as `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent`) must be stripped from the active technical design table and moved to a provisional appendix.
- **Cross-File Integrity:**
  - Verified that the key taxonomy table perfectly aligns with the `LifecycleInput` fields and key families declared in `internal/session/lifecycle_projection.go`.
- **Inter-Reviewer Alignment:**
  - Our findings strongly support the **YAGNI Contract Scope Reviewer (Kwame Asante)** and the **Reconciler Runtime Fact Reviewer (Liam Okonkwo)**. The lack of a physical, committed Slice 0 prototype is the root cause of these cascading review blocks, proving that the design must remain in an `iterate` / `block` state until concrete fixtures and code files are generated.
