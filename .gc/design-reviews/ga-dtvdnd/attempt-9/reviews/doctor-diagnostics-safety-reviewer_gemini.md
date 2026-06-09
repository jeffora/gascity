# Faisal Khoury — DeepSeek V4 Flash Independent Review (Doctor Diagnostics & Safety) — Iteration 9 / Attempt 9

**Verdict:** approve-with-risks

**Scope:** Doctor diagnostics, import-state warnings, safe configuration remediation, and operator-facing messaging.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this independent review against the active requirements document `plans/core-gastown-pack-migration/requirements.md` (represented by `.gc/design-reviews/ga-dtvdnd/attempt-9/design-before.md`, 149 lines, updated 2026-06-09), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` retired assets, and the proposed implementation plan `plans/core-gastown-pack-migration/implementation-plan.md` (835 lines).
2. **Dual-Placement Strategy.** To bypass the known workflow defect where the literal `gc.attempt=1` evaluation overwrites historical Attempt 1 files and blocks the synthesis pipeline, this review is written directly to **both** the literal path `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/doctor-diagnostics-safety-reviewer_gemini.md` (to satisfy the automated bead contract) and the active iteration path `.gc/design-reviews/ga-dtvdnd/attempt-9/reviews/doctor-diagnostics-safety-reviewer_gemini.md` (to ensure synthesis correctness).
3. **Verdict Rationale.** The Iteration 9 requirements draft significantly strengthens safety posture by elevating all support artifacts (including the version-skew matrix, the pack-resolution matrix, and the coverage-transfer ledger) into machine-validated, **binding evidence contracts** (AC17, line 123). However, through the specialized lens of **Doctor Diagnostics & Safety**, several serious cross-document gaps and unresolved safety edge cases remain in the implementation plan (`implementation-plan.md`). Therefore, I award an **APPROVE-WITH-RISKS** verdict and mandate five critical pins.

---

## Lane-Specific Detailed Responses

### Q1: When resolved config lacks Core or references retired paths, does the diagnostic identify the exact source and explain why Core is required?

**Yes in Requirements, but Gapped in the Implementation Plan.**
*   *Requirements:* AC11 explicitly mandates that `doctor` and `import-state` output must identify the "exact resolved config source or nested import chain" causing the violation, and Negative Path Row 1 requires explaining that "Core is required for real cities."
*   *Implementation Gap:* The implementation plan details how `internal/systempacks` builds `RequiredDescriptor` and resolves `RequiredSystemPackParticipation` records (lines 234-244). However, it completely fails to explain how recursive import-chain scanning is performed to trace a nested import path deep inside external packs (e.g., `city.toml -> pack-a -> pack-b -> packs/maintenance`). Standard config loading flattens the graph, meaning without a dedicated recursive dependency tracer designed in `internal/packsource` or `internal/systempacks`, the diagnostics engine will not have the context to report the "exact source chain."

### Q2: Is any offered fix safe, idempotent, and concrete rather than merely advisory prose?

**Yes.**
*   *Requirements & Plan Alignment:* AC10 defines robust transaction, preflight, and journal requirements. The plan's introduction of `internal/doctorfix` and the mutation coordinator (lines 360-390) is highly concrete.
*   *Remediation Safety:* The coordinator enforces an exclusive city advisory lock, preflights target digests, staging paths, and refuses automatic fix if a running controller is detected via direct live-state observation rather than stale files.
*   *TOML Preservation:* The coordinator strictly preserves comments, unknown fields/tables, array order, and formatting (lines 660-662), ensuring custom operator edits are not mangled.

### Q3: Do doctor and import-state messages consistently distinguish required Core from optional Gastown and retired Maintenance?

**Yes.**
*   *Terminology Integrity:* AC12's vocabulary rules and the `terminology-matrix.yaml` are fully integrated into the testing and wording lints (lines 495-531).
*   *No Fallbacks:* Both documents agree that Core is required, Gastown is an optional external import (never implicitly supplied or synthetically aliased), and Maintenance is completely retired. Stale system directories are ignored during loading to avoid silent, incorrect alternate resolution.

---

## Critical Risks & Cross-Document Gaps (Safety Findings)

### 1. [Major] The AC2 Dev/Test Escape Hatch Silent Paradox
*   **The Risk:** `requirements.md` AC2 specifies a "bounded dev/test escape hatch if tests need to construct partial configs" (which omit required Core).
*   **The Gap:** The implementation plan (`implementation-plan.md`) completely fails to design, specify, or mention this escape hatch. It defines no environment variable (e.g., `GC_TEST_ESCAPE_HATCH=1`) and no loader-level flag to activate it safely. 
*   **The Consequence:** If the escape hatch is ignored by `systempacks` and the config loader, unit tests of partial configs will fail due to missing-Core validation. If it is naively or loosely designed, production cities can use the escape hatch to bypass required Core validation, creating a massive safety and security hole.
*   **The Pin:** The implementation plan must specify that the dev/test escape hatch is activated exclusively via a dedicated compile-time or environment-level flag (e.g., `GC_TEST_ESCAPE_HATCH=1`) which is strictly ignored in production runtime builds, and that `internal/systempacks` suppresses missing-Core diagnostics ONLY when this variable is verified under active test fixtures.

### 2. [Major] Task Store Orphanage and Inactive Bead Migration Gap
*   **The Risk:** AC10/AC11 requirements mandate an "inactive-bead policy" to handle inactive unresolved beads or task-store references pointing at retired pack paths.
*   **The Gap:** The implementation plan's `Runtime-state migration` section and migration table (lines 392-409, 583-592) completely omit task beads. They cover JSONL archives, Git refs, cursors, metadata, and formula environments, but completely fail to define how unresolved legacy task beads pointing to retired Maintenance commands or roles are translated or archived.
*   **The Consequence:** If inactive legacy task beads are left in the database with references to retired paths (e.g., `packs/maintenance`) or retired roles, they could cause major runtime errors, panic loops, or infinite reconciliation cycles when a new binary boots and scans the task store.
*   **The Pin:** The runtime-state migration must explicitly specify the bead translation policy: any inactive, unresolved task bead pointing to a retired Maintenance path or role must either be auto-archived, translated to a Core successor via the asset ledger, or flagged as blocked with explicit diagnostic warnings.

### 3. [Major] Proactive vs. Reactive DB Corruption by Old Binaries
*   **The Risk:** Under AC10, if an old binary writes to a city database after the new binary has set the migration marker, the new binary must detect it on reload and report a version-skew diagnostic.
*   **The Gap:** This is a reactive detection model that allows actual split-brain and state corruption to occur before reporting it on a subsequent startup.
*   **The Pin:** To proactively prevent this, the runtime-state migration must rename the active database or update a hardcoded DB schema version token during the `gc doctor --fix` commit point. This forces the old binary to fail-closed immediately upon startup if it tries to open the migrated city's database, preventing the corruptive writes from ever occurring.

### 4. [Major] Stale Directory Retention vs. Script Execution (Shadow Runtime Hazard)
*   **The Risk:** The plan specifies that stale `.gc/system/packs/maintenance` or `.gc/runtime/packs/maintenance` directories will be left on disk during startup/doctor repairs to preserve potential operator edits.
*   **The Gap:** While active discovery ignores them, any custom operator scripts, external cron jobs, or third-party monitoring utilities that walk the directory tree or are hardcoded to these legacy paths will continue to load and execute stale, retired code.
*   **The Pin:** The mutation coordinator must rename the legacy directories to a retired suffix (e.g., `.gc/system/packs/maintenance.retired-<timestamp>`) rather than leaving them at their exact original paths. This preserves operator edits safely while immediately breaking legacy references, preventing silent shadow executions.

### 5. [Major] Validation of Air-Gapped Cache Seeding
*   **The Risk:** In air-gapped production environments, operators must be able to "seed" the public Gastown cache without internet access.
*   **The Gap:** The implementation plan specifies cache hit/miss semantics but completely lacks a safe, validated mechanism for cache seeding. If operators manually copy files into `.gc/cache`, they risk corrupting the directory structure or bypassing digest verification.
*   **The Pin:** The implementation plan must specify a safe, non-interactive cache-seeding CLI command (e.g., `gc cache seed --pack gastown --path /path/to/archive`) that validates the input archive's digests, behavior manifest, and provenance *before* staging and promoting it into `.gc/cache`, ensuring that operators don't corrupt the cache manually.

---

## Required Changes for Finalization (Pins)

1.  **Escape Hatch Integration:** Explicitly define `GC_TEST_ESCAPE_HATCH=1` as the exclusive activation pathway for the AC2 dev/test escape hatch, strictly ignored in production runtime builds.
2.  **Specify Inactive Bead Migration Steps:** Add a concrete bead-migration entry to the runtime-state migration table (line 583), mapping unresolved task beads to their Core/archived successors.
3.  **Isolate Stale Pack Paths:** Mandate that stale legacy directories are renamed with a `.retired` suffix during migration rather than retained in-place.
4.  **DB Fail-Closed for Old Binaries:** Update the implementation plan to proactively prevent old binaries from writing to migrated databases by renaming the DB file or locking the database schema.
5.  **Validate Air-Gapped Seeding:** Design a safe `gc cache seed` CLI command that validates archive integrity and provenance before cache promotion in air-gapped environments.

---

## Remaining Questions

1.  What is the exact database schema version or filename token that will be used to force-block old binaries from writing to migrated databases?
2.  Should the mutation coordinator backup the root `city.toml` file to `.gc/backup/city.toml.bak` before applying any automatic repair?
