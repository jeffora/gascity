# Camille Okafor — DeepSeek V4 Flash Perspective Independent Review (Iteration 7 / Attempt 7)

**Verdict:** approve-with-risks

**Scope:** Existing-city upgrades, legacy local paths, compatibility shims, version skew, rollback/downgrade, offline cache seeding, and two-repo rollout mechanics.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this review against the current `plans/core-gastown-pack-migration/requirements.md` (specifically `.gc/design-reviews/ga-dtvdnd/attempt-7/design-before.md`, updated 2026-06-09), the `plans/core-gastown-pack-migration/implementation-plan.md` (updated 2026-06-09), the `gc.mayor.requirements.v1` schema, and the live `examples/gastown/packs/maintenance` assets. All findings are verified fresh against the repository tree without relying on historical assumptions.
2. **Dual-Placement Strategy.** Due to a known workflow defect where the bead's metadata `gc.attempt=1` causes automatic execution tools to write and read from `attempt-1/reviews/migration-rollout-reviewer_gemini.md` (which can block attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/migration-rollout-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-7/reviews/migration-rollout-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 7 synthesis.
3. **Verdict Rationale.** The Iteration 7 / Attempt 7 requirements and implementation plan represent a highly robust, well-sequenced migration strategy. Gaps from prior loops (such as air-gapped cache seeding under AC10/AC16 and old-binary post-marker write diagnostics) have been explicitly integrated. However, from the perspective of **Existing-City Upgrade and Rollout/Recovery**, several critical risks must be addressed before implementation approval. These include the doctor boot-loop risk during bootstrap-only execution, task-store orphanage for inactive but unresolved legacy beads, and the lack of an automated rollback/restore path. I award an **APPROVE-WITH-RISKS** verdict and mandate three critical pins.

---

## Top Strengths of the Iteration 7 Plan

- **Comprehensive 7-Slice Rollout Architecture:** Splitting the rollout into distinct, manageable slices (from Slices 1a through 7) prevents the premature consumption of the public activation pin. Requiring `behavior-preservation.yaml` and `public-gastown-pins.yaml` as external prerequisites ensures a strict, evidence-backed gate before source deletion occurs.
- **Durable, Coordinator-Owned Repairs:** Moving direct mutations out of individual checks into a centralized, lock-protected mutation coordinator (`internal/doctorfix`) is an outstanding design. The use of a crash-released city advisory lock, preflight/post-commit validations, and multi-file transaction commit points eliminates the risk of partial edits.
- **Deterministic Live-State Detection:** Querying live process state (via the process table/lsof) rather than fragile status files to refuse mutating repairs on cities with active controllers or running tmux sessions is perfectly aligned with the project's core design principles.
- **Configurable Core Maintenance Executor:** Replacing hardcoded Go references to `dog` with a configurable maintenance worker contract and `[gc.bindings.*]` config precedence successfully preserves Zero Framework Cognition (ZFC) while maintaining SDK self-sufficiency.

---

## Critical Risks & Rollout Gaps (Architectural Findings)

### 1. [Major] Eager Loader Dependency in Bootstrap-Only Mode (The Doctor Boot-Loop)
* **The Risk:** AC11 specifies that `gc doctor` and repair commands have a bootstrap-only diagnostic mode that can run even when normal pack resolution is broken (e.g., when Core is missing or corrupted). However, if the CLI bootstrap process eagerly initializes the config loader or triggers required-pack validation before detecting that the command being run is a bootstrap-only repair command, the CLI will exit with a fatal error. This results in a boot loop, completely blocking the operator from running `gc doctor --fix` to restore the city.
* **The Gap:** There is no explicit boundary in the CLI entry point separating bootstrap-only commands from eager loader initialization.
* **The Pin:** The CLI bootstrap path must strictly detect bootstrap-only commands (`gc doctor`, `gc import-state`, `gc version`) at the very entry point, executing them in a completely dependency-isolated path that bypasses Gate 1 and Gate 2 validation.

### 2. [Major] Inactive Task Store Orphanage (The In-Flight Legacy Bead Hazard)
* **The Risk:** The plan correctly refuses mutating repairs when active, live sessions are running. However, inactive but unresolved beads/tasks may still exist in the task store. When the migration is executed, Maintenance is retired.
* **The Gap:** These inactive, unresolved beads may contain dependencies or triggers pointing at legacy Maintenance paths, commands, or roles. If the doctor migrates config and task-store schemas without a clear policy for these pending tasks, the new controller will either fail to reconcile them, crash, or permanently orphan them in the database.
* **The Pin:** The doctor-owned state migration must define a deterministic policy for unresolved, inactive task-store beads referencing retired paths—either auto-archiving them, translating them to Core equivalents using the asset ledger, or prompting the operator with actionable remediation.

### 3. [Major] One-Way Upgrade and Lack of Automated Rollback/Restore
* **The Risk:** The rollout matrix mentions "rollback from new to old: Doctor-mutated manifests are either readable by old binaries or release notes name explicit downgrade limits and manual recovery."
* **The Gap:** However, if the new binary writes tasks, archives, or state using the new Core-owned paths, and then the operator needs to rollback to the old binary, the old binary will expect the retired `packs/maintenance` or legacy local paths. Because the old binary does not know about the new Core-owned paths, it will fail to see the new state or crash due to missing legacy files. This is a classic "one-way street" upgrade where rollback is practically impossible or destructive without manual state reconstruction.
* **The Pin:** The rollout and rollback policy must explicitly mandate that `gc doctor --fix` creates a full, restorable backup of all modified configuration files AND the task store/state directory *before* mutating any files on disk. The manual recovery/downgrade notes must provide an automated or highly simple restore script (e.g., `gc doctor --restore-backup`) rather than leaving the operator to manually reconstruct the legacy state.

### 4. [Minor] Read-Only Transitive Dependency Deadlock (Upstream Lag)
* **The Risk:** If a city imports a third-party remote pack that has nested imports referencing retired paths like `packs/maintenance`, the city will fail closed during loading.
* **The Gap:** Because these third-party dependencies are read-only and nested, the local operator cannot run a `gc doctor --fix` to rewrite them. Without a local dependency patching/override mechanism (like cargo `[patch]` or npm `overrides`), the operator is completely blocked from upgrading Gas City until the upstream author publishes an update. While this is the decided "unsupported-state" policy, the rollout impact of locking operators out of their own cities must be highlighted.

---

## Required Changes for Finalization

1. **Loader Bypass for Bootstrap Commands:** Hard-gate the CLI entry point so that bootstrap-only commands (`gc doctor`, `gc import-state`, `gc version`) bypass required-pack loading, materialization, and config-resolution validation, executing in an isolated bootstrap path.
2. **Deterministic Legacy Bead Policy:** Add a requirement to AC10/AC11 stating that the doctor-owned task-store migration must reconcile inactive, unresolved beads referencing retired paths—either by archiving them, translating them, or warning the operator.
3. **Automate Pre-Fix Backups and Recovery:** Require the mutation coordinator to create a full, restorable backup of config and database files prior to writing, and define an automated restore path (`gc doctor --restore-backup` or similar) to support safe rollbacks and downgrades.

---

## Open Questions for Implementation

- What is the exact manual reconciliation flow for an operator whose city has triggered the "old-binary post-marker write" fail-closed diagnostic?
- Is there a plan to provide a simple `gc cache seed` or `gc pack import` command to facilitate air-gapped migrations for offline operators?
