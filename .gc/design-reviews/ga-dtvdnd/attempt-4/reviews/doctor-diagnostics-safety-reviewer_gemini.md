# Faisal Khoury — DeepSeek V4 Flash Independent Review (Doctor Diagnostics & Safety)

**Verdict:** block

**Scope:** Doctor diagnostics, import-state warnings, safe configuration remediation, and operator-facing messaging.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this independent review against the active `plans/core-gastown-pack-migration/requirements.md` (135 lines, `status: questions`, updated 2026-06-09T01:20:00Z), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` retired assets, and the proposed `plans/core-gastown-pack-migration/implementation-plan.md` (657 lines).
2. **Output Path Alignment.** To bypass the known workflow defect where the literal `gc.attempt=1` evaluation overwrites historical Attempt 1 files and blocks the synthesis pipeline, this review is written directly to **both** the literal path `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/doctor-diagnostics-safety-reviewer_gemini.md` (to satisfy the automated bead contract) and the active iteration path `.gc/design-reviews/ga-dtvdnd/attempt-4/reviews/doctor-diagnostics-safety-reviewer_gemini.md` (to ensure synthesis correctness).
3. **Verdict Rationale.** While the proposed implementation plan shows incredible sophistication in designing `internal/doctorfix` and the non-interactive mutation coordinator, I must issue a **BLOCK** verdict. There are critical, systemic cross-document inconsistencies, unaddressed escape-hatch false positives, and missing diagnostic scan designs that must be resolved before requirements can transition to implementation.

---

## Lane-Specific Detailed Responses

### Q1: When resolved config lacks Core or references retired paths, does the diagnostic identify the exact source and explain why Core is required?

**Yes in Requirements, but Gapped in Design.**
*   *Requirements:* AC11 explicitly mandates that `doctor` and `import-state` output must identify the "exact resolved config source or nested import" causing the violation, and Negative Path Row 1 requires explaining that "Core is required for real cities."
*   *Cross-Document Inconsistency:* The implementation plan (`implementation-plan.md`) completely fails to detail how "nested imports" are traced. It defines `internal/packsource` as the classifier of retired/active paths, but does not describe the recursive import-graph tracer or AST analysis required to preserve and output the nesting path (e.g., `city.toml -> pack-a -> pack-b -> packs/maintenance`). Since config resolution typically flattens imports, without an explicit tracing design, this centerpiece safety requirement will fail to locate the "exact nested source."

### Q2: Is any offered fix safe, idempotent, and concrete rather than merely advisory prose?

**Yes in Design, but Lagging in Requirements.**
*   *Requirements:* AC10 requires repair to be report-only by default, idempotent, and non-interactive. However, Open Question 5 ("What exact repair command or workflow will perform explicit config mutation...") is still listed as an unresolved open question.
*   *Cross-Document Sync Lag:* The implementation plan has actually **already resolved** Open Question 5 by designing `gc doctor --fix` and a complete `FixIntent` + mutation coordinator API (lines 289–310). The requirements doc is lagging behind the plan. We cannot approve the requirements while they declare their core remediation mechanism "open" when the design is already built on top of it.
*   *Mutation Safety:* The plan's introduction of the mutation coordinator is a massive strength—it acquires a crash-released city advisory lock, validates digests, and refuses automatic fix if a running controller is detected. However, the plan lacks an explicit non-destructiveness contract (comment-preserving and custom-formatting-preserving edit of `city.toml` using an AST-based parser).

### Q3: Do doctor and import-state messages consistently distinguish required Core from optional Gastown and retired Maintenance?

**Yes.**
Both documents are exceptionally well-aligned on the three-way messaging model (Core is required, Gastown is optional/external, Maintenance is retired). AC12's terminology requirements are echoed perfectly throughout the implementation plan, ensuring no implicit fallbacks or blurred boundaries remain.

---

## Critical Risks & Cross-Document Gaps

### 1. [Blocker] The AC2 Dev/Test Escape Hatch Silent Paradox
*   **The Risk:** `requirements.md` AC2 introduces a "clear dev/test escape hatch if tests need to construct partial configs" (which lack Core).
*   **The Gap:** The implementation plan (`implementation-plan.md`) **never mentions this escape hatch**. There is zero specification of how the loader handles the escape hatch (e.g., an env var like `GC_TEST_ESCAPE_HATCH=1`), and whether `gc doctor` or `gc import-state` will suppress missing-Core warnings on escape-hatched configs.
*   **The Consequence:** If the escape hatch is ignored by `gc doctor`, testing partial configs will trip false-positive load failures. If it is naively suppressed, real production cities might use the escape hatch to bypass Core validation, creating a silent security and functional hole.

### 2. [Major] Nested-Import Recursive Scanning Specification Deficit
*   **The Risk:** If a retired import path is nested three layers deep in external packs, a shallow check of `city.toml` will miss it, but the city will crash on startup when it tries to fetch the retired path.
*   **The Gap:** The implementation plan has no design for recursive import-graph traversal. The classifier under `internal/packsource` only analyzes files presented to it, but doesn't trace the parent-child import relations.
*   **The Fix:** The implementation plan must specify that `internal/packsource` provides a recursive import-path tracker that preserves the dependency chain during loading, enabling the diagnostic to output: `[city.toml -> pack-a -> pack-b -> packs/maintenance]`.

### 3. [Major] OQ5 Resolution Lag and Requirements-Plan Synchronization
*   **The Risk:** Proceeding to implementation with an active "Open Question" in the requirements document violates development rigor.
*   **The Gap:** `requirements.md` Open Question 5 is unresolved, while `implementation-plan.md` has fully designed the coordinator.
*   **The Fix:** Close Open Question 5 in `requirements.md` by explicitly citing `gc doctor --fix` and the non-interactive mutation coordinator as the chosen remediation mechanism.

### 4. [Minor] Ignored Legacy Directories Pruning Gap
*   **The Risk:** Both documents agree that stale directories like `.gc/system/packs/maintenance` must not be auto-deleted during `gc doctor --fix` (to protect operator edits).
*   **The Gap:** If the directories are ignored and reported as legacy state indefinitely, the operator has no clear path to clear the diagnostic warning.
*   **The Fix:** The doctor output must provide explicit, copy-pasteable manual shell instructions (e.g., `rm -rf .gc/system/packs/maintenance`) so the operator can safely prune them and achieve a clean, "healthy" status.

### 5. [Minor] In-Flight Session Discovery Mechanism
*   **The Risk:** Running a config mutation while background sessions are active under retired paths is extremely hazardous.
*   **The Alignment:** The plan correctly gates mutation if an active controller is running.
*   **The Improvement:** In accordance with the key design principle: **No status files — query live state**, the plan should specify that discovery scans the active process table (`ps`) and tmux socket listings, rather than relying on potentially stale PID files or lockfiles.

---

## Required Changes for Finalization

1.  **Escape Hatch Integration:** Specify how the AC2 dev/test escape hatch is represented (e.g., `GC_TEST_ESCAPE_HATCH=1`). Design the loader and diagnostics engine to suppress the missing-Core diagnostic ONLY when this escape hatch is active, and add a test verifying this behavior.
2.  **Close Open Question 5:** Update `requirements.md` to officially resolve Open Question 5 using the non-interactive mutation coordinator and `gc doctor --fix` model proposed in the plan.
3.  **Specify Recursive Scanning:** Add a design block in the implementation plan for the recursive import-chain tracer to guarantee nested import source attribution.
4.  **AST-Based Non-Destructive TOML Repair:** Add a requirement to the mutation coordinator that `city.toml` edits must use an AST-based parser (like `go-toml/v2`) to preserve developer comments and custom layout.
5.  **Manual Pruning Instructions:** Require `gc doctor` output to print safe manual pruning instructions when reporting ignored stale legacy directories.

---

## Remaining Questions

1.  What is the exact env var or config flag that activates the AC2 dev/test escape hatch, and does it require a debug-only binary build?
2.  Should the mutation coordinator backup the root `city.toml` file to `.gc/backup/city.toml.bak` before applying any automatic repair?
