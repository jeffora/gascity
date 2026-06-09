# Camille Okafor — DeepSeek V4 Flash Perspective Independent Review (Iteration 9 / Attempt 1)

**Verdict:** approve-with-risks

**Scope:** Existing-city upgrades, legacy local paths, compatibility shims, version skew, rollback/downgrade, offline cache seeding, and two-repo rollout mechanics.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this review against the current `plans/core-gastown-pack-migration/requirements.md` (represented by `.gc/design-reviews/ga-dtvdnd/attempt-9/design-before.md`, 149 lines, updated 2026-06-09 with the new binding-contract clause in AC17), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` dog assets this migration retires, the public `gascity-packs/gastown` pack source, and the updated `plans/core-gastown-pack-migration/implementation-plan.md` (835 lines).
2. **Dual-Placement Strategy.** To ensure complete compliance with automated workflow tooling while unblocking the active iteration 9 synthesis, I am writing this complete independent review to **both** of the following paths:
   - `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/migration-rollout-reviewer_gemini.md` (the physical file matching the bead's metadata-derived target)
   - `.gc/design-reviews/ga-dtvdnd/attempt-9/reviews/migration-rollout-reviewer_gemini.md` (the active synthesis directory)
3. **Verdict Rationale.** The Iteration 9 / Attempt 1 requirements draft elevates all auxiliary support files (such as the version-skew matrix, the public pin ledger, and the pack-resolution matrix) into machine-validated, **binding evidence contracts** (AC17, line 123), which represents an exceptional advancement in rollout safety and compliance. However, from the specific lens of **Existing-City Upgrade and Rollout/Recovery**, the actual proposed implementation plan (`implementation-plan.md`) remains unchanged from Attempt 8 and continues to carry several critical architectural risks and omissions. I award an **APPROVE-WITH-RISKS** verdict and mandate four critical pins.

---

## Evaluation of the Three Key Questions

### 1. For an existing city referencing packs/maintenance or packs/gastown local paths, what exact migration steps and diagnostics does the requirement demand?
**Auditor Finding: Highly Detailed and Actionable.**
* AC10 and AC11 specify detailed diagnostics for existing-city upgrades, identifying legacy paths, transitive dependencies, stale system packs, and custom overlays.
* The diagnostics run in a bootstrap-only mode when normal pack resolution is broken and output stable condition codes with exact source attribution.
* The mutating repair is report-only by default and uses `gc doctor --fix --non-interactive`, which is failure-atomic, idempotent, and backup-guarded.

### 2. Is any compatibility shim or grace behavior scoped and sequenced so pinned or in-flight references degrade gracefully?
**Auditor Finding: Yes.**
* Under AC15 and the version-skew-matrix, compatibility-pin windows are explicitly temporary and scoped.
* In-flight active sessions started from retired Maintenance prevent mutating repairs until they are drained, avoiding runtime split-brain.
* Stale system pack directories are ignored during loading but reported as legacy state to prevent silent alternate resolution sources.

### 3. How does the rollout handle public pack pinning, intermediate two-repo states, rollback, and operator repair expectations?
**Auditor Finding: Structurally Strong but Lacks Automated Operator Recovery.**
* AC14 requires a strict two-repository release sequence, ensuring that Gas City never ships an activation pin that lacks the validated public Gastown behavior manifest.
* Offline/cache promotion uses digest-verified remote caches and fails closed on cache misses or digest mismatches rather than falling back to stale local copies.
* Rollback during implementation slices is well-defined, but operator-visible automated rollback paths (e.g., restoring modified configuration files and task databases) are delegated to manual operator instructions rather than being natively automated.

---

## Top Strengths of the Iteration 9 Requirements

- **Binding Evidence Contracts (The AC17 Strengthening):** Explicitly declaring the support files (AC3, AC5, AC6, AC7, AC8, AC11, AC12, AC13, AC15, AC16, and AC17) as binding evidence contracts, not merely illustrative examples, prevents downstream implementation deviation and guarantees that the safety matrices are fully validated.
- **Exhaustive Version-Skew Matrix:** Defining clear, deterministic outcomes for old-binary/new-pack and new-binary/old-lock states closes the version-skew gap and avoids operator confusion during rollouts.
- **Robust Multi-File Transaction Guarantees:** Implementing a crash-released city advisory lock and a structured `FixIntent` plan inside the mutation coordinator prevents partial, corrupted writes.
- **Fail-Closed Offline Resolution:** Eliminating in-tree examples, system packs, or synthetic aliases as offline tie-breakers secures cache-integrity and forces operators to resolve cache-misses explicitly.

---

## Critical Risks & Rollout Gaps (Architectural Findings)

### 1. [Major] Reactive vs. Proactive Old-Binary Write Detection (The Silent Post-Marker DB Corruption)
* **The Risk:** The plan specifies that if an old binary writes after the migration marker is set, the new binary will detect it on the next reload and report a version-skew diagnostic, requiring manual reconciliation. However, because the database file name (`beads.db`) and basic schema are shared between binaries, the old binary is allowed to boot and write corruptive or obsolete records directly into the state database before any detection happens. 
* **The Gap:** This is a reactive detection model that allows actual split-brain and state corruption to occur before reporting it on a subsequent startup.
* **The Pin:** To proactively prevent this, the runtime-state migration must rename the active database or update a hardcoded DB schema version token during the `gc doctor --fix` commit point. This forces the old binary to fail-closed immediately upon startup if it tries to open the migrated city's database, preventing the corruptive writes from ever occurring.

### 2. [Major] Task Store Orphanage Omission in the Implementation Plan (The Inactive Legacy Bead Gap)
* **The Risk:** AC10/AC11 requirements mandate an "inactive-bead policy" to handle inactive unresolved beads or task-store references pointing at retired pack paths. However, the `Runtime-state migration` specification in the implementation plan (lines 583-592) completely omits beads and tasks from the migration table. It covers JSONL archives, refs/remotes, push cursors, escalation fields, spawn-storm ledgers, and order skip tracking, but fails to define how the coordinator translates or archives pending task beads.
* **The Gap:** A major cross-document inconsistency between requirements (which demand inactive-bead policy tests) and the implementation plan (which fails to specify how the task-store is actually migrated).
* **The Pin:** The implementation plan's runtime-state migration must explicitly define the bead translation policy: any inactive, unresolved task bead pointing to a retired Maintenance command or role must either be auto-archived, translated to a Core successor via the asset ledger, or flagged as blocked with explicit diagnostic warnings.

### 3. [Major] Stale Directory Retention vs. Custom Script Execution (The Shadow Runtime Hazard)
* **The Risk:** The plan explicitly retains stale `.gc/system/packs/maintenance` or `.gc/runtime/packs/maintenance` directories on disk to preserve "operator edits" and avoid automated deletion. While the core CLI ignores these during active discovery, any custom operator scripts, external cron jobs, or third-party monitoring utilities that walk the directory tree or are hardcoded to these legacy paths will continue to load and execute stale, retired code.
* **The Gap:** This creates a dangerous shadow runtime where the main CLI runs on Core but other automated maintenance hooks run on retired code.
* **The Pin:** The mutation coordinator must rename the legacy directories to a retired suffix (e.g. `.gc/system/packs/maintenance.retired-<timestamp>`) rather than leaving them at their exact original paths. This preserves operator edits safely while immediately breaking legacy references, preventing silent shadow executions.

### 4. [Minor] Lack of Automated Operator-Facing Rollback/Restore Path
* **The Risk:** The rollout and rollback policy delegates the operator-level rollback of modified configurations and database state to "release notes manual recovery." 
* **The Gap:** Requiring operators to manually restore TOML files and databases from a journal/backup directory is slow and error-prone, violating the "non-interactive, safe remediation" posture.
* **The Pin:** The mutation coordinator should support a simple, automated rollback command (e.g., `gc doctor --rollback` or `gc doctor --restore-backup`) that parses the active recovery journal and atomically restores the staged backups.

---

## Required Changes for Finalization

1. **DB Fail-Closed for Old Binaries:** Update the implementation plan to proactively prevent old binaries from writing to migrated databases by renaming the DB file or locking the database schema.
2. **Specify Inactive Bead Migration Steps:** Add a concrete bead-migration entry to the runtime-state migration table (line 583), mapping unresolved task beads to their Core/archived successors.
3. **Isolate Stale Pack Paths:** Mandate that stale legacy directories are renamed with a `.retired` suffix during migration rather than retained in-place.
4. **Automate Operator Rollback:** Specify an automated, single-command operator rollback mechanism in the mutation coordinator.

---

## Open Questions for Implementation

- Will the doctor diagnostics warn the operator if any custom user-configured script in `city.toml` contains path-literal references to the retired `packs/maintenance` directory?
- What are the exact manual steps recommended to an operator if the database state must be reconstructed after an old binary write has bypassed reactive detection?
