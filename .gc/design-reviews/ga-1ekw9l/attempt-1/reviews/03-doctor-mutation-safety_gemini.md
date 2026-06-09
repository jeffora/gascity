# Leah Okafor — Doctor and Runtime-State Mutation Safety Reviewer (Iteration 9 / Attempt 9, Independent DeepSeek V4 Flash Style)

**Verdict:** block

> **Lane:** Doctor `--fix` coordinator atomicity, byte-preserving TOML editing, concurrency with live controllers, advisory locks, idempotent recovery.
> 
> Reviewed against the Iteration 9 / Attempt 9 implementation plan (`plans/core-gastown-pack-migration/implementation-plan.md`, 835 lines, `updated_at: 2026-06-09T13:20:59Z`) — specifically §"Doctor And Runtime-State Mutation Safety" (lines 358–408), §"Data And State" (lines 532–591, with recovery schemas/ledger at lines 570–580), and §"Testing" (specifically Doctor and runtime-state tests at lines 657–667).
> 
> This independent review is produced using the DeepSeek V4 Flash style, focusing rigorously on first-principles trust boundaries, cross-document state consistency, and unstated runtime assumptions.

---

## Schema Conformance

**Conforms.** The Iteration 9 / Attempt 9 implementation plan includes all required top-level sections in the correct order, carries the required front matter (`phase: implementation-plan`), and `Open Questions` is correctly marked as `None`. The doctor-safety and recovery material is properly integrated into the Proposed Implementation, Data And State (including the recovery and migration marker schemas), Testing, and Rollout sections, rather than being appended as unstructured review prose.

---

## Top Strengths of the Design

- **Structured Fix Inventory & Linter Gate (lines 370–377):** The introduction of `plans/core-gastown-pack-migration/support/doctor-fix-inventory.yaml` is a highly robust architectural gate. By mapping every single legacy direct `Fix(ctx)` caller, import rewrite, and test-only helper to a declared Row with an assigned disposition (`FixIntent`, `report-only`, `remove`, or `allowed-test-helper`), the design forces clean closure of the legacy un-coordinated write surface.
- **Unified Mutation Boundary (lines 361–368):** Consolidating all system-modifying operations (manifest writes, lockfiles, installed pack directories, and runtime-state) under `internal/doctorfix` ensures that no raw writes bypass the transaction coordinator.
- **Lock-First Preflight Sequence (lines 379–381):** Requiring the directory advisory lock to be acquired *before* preflight digests are validated, and repeating digest and provenance verification after lock acquisition if a report-only phase did any pre-lock reads, closes potential TOCTOU race windows.
- **Durable Write-Ahead Recovery Journal Shape (lines 385–390, 570–573):** Writing a detailed journal schema (mutation intent ID, preflight file digests, staged paths, publish order, commit point, completed steps, rollback instructions, and final validation) prior to any publishing step provides a sound conceptual foundation for failure-atomicity.
- **No Destructive Defaults (lines 353–356):** Refusing to automatically delete stale `.gc/system/packs/maintenance` or `.gc/runtime/packs/maintenance` directories prevents catastrophic loss of operator-edited files.

---

## Critical Risks & Consensus Blockers

### 1. The Unnamed "Single Commit Point" Mechanism and Impossible Rollback (lines 385–390, 570–573)

The plan asserts a "single commit point" (line 386) and records a "commit point" field in the recovery ledger (line 571), but never defines what specific filesystem operation linearizes the multi-file transaction. POSIX renames are atomic per-file, not across files. Without specifying the linearizing operation (e.g., writing or renaming a `committed` marker file in the recovery ledger directory after all staged data temps are fsync'd), a crash could leave target files half-renamed without a deterministic before/after boundary.
Furthermore, the recovery schema in lines 570–571 records "preflight file digests, staged paths... rollback instructions" but **carries no copy of the original file bytes**. Once a staged temp file is renamed over an existing file, the original bytes are gone. Without original-byte backups, the "rolls back" branch (line 388) is physically impossible with only a digest.
- **Required Resolution:** Define the single commit point filesystem operation (e.g., writing/renaming a specific commit marker file that gates post-marker idempotent roll-forward). State clearly that recovery is roll-forward-only (re-apply from staged temps), or add original-byte backups to the recovery record to make rollback physically possible.

### 2. Read-Path Required-Pack Repair Violates the Only-Writer Invariant (lines 536–539 vs 366–368)

This is a major cross-document inconsistency. Lines 366–368 assert: *"The coordinator is the only path that writes city manifests, lockfiles, installed pack directories, runtime-state migrations, or import rewrites."* Yet, lines 536–539 explicitly require that normal, read-path required-pack loading *"regenerates missing or corrupt expected files from the embedded manifest, prunes generated unexpected effective files, and quarantines operator-edited or unclassifiable files."* 
This regeneration, pruning, and quarantining are raw file mutations. Because this read-path repair executes *without* acquiring the city advisory lock, a doctor-fix transaction renaming a Core directory can race a concurrent read-path load that is regenerating/pruning files in the same tree, resulting in partial file writes, permission errors, or silent corruption.
- **Required Resolution:** Resolve this contradiction. Either route the load-path repair through the same city advisory lock, or completely remove automatic write-repair from the normal load path (making it report-only) and delegate all repairs to the doctor coordinator.

### 3. Lack of Continuous Advisory Lock Holding by the Controller & TOCTOU Window (lines 379–383)

The doctor refuses automatic fixes if a running controller is detected via live runtime state (lines 381–383). However, there is a clear TOCTOU race: a controller could start up *immediately after* the doctor checks the process table but *before* the doctor acquires the directory advisory lock.
Furthermore, the plan does not mandate that the running controller holds this city advisory lock *continuously* during its entire lifecycle. If the controller only acquires the lock transiently during pack install operations, it can perform concurrent file reads/writes while the doctor is executing mutations.
- **Required Resolution:** Mandate that any active controller must acquire and hold the city advisory lock continuously for its entire active lifecycle, and ensure that the doctor's liveness check relies on lock contention rather than volatile process table lookups.

### 4. Missing Provenance Refusal Rule for Custom/Fork Imports (lines 381, 402)

While the plan mentions repeating digest and provenance validation (line 381), it never states that the doctor *refuses* to write or rewrite imports that are classified as custom/fork or cannot be proven to be system-generated. Without an explicit provenance-based refusal gate, automatic fixes risk overwriting operator-customized packs or forked repositories.
- **Required Resolution:** Add an explicit refusal rule: the doctor must refuse automatic fixes on any import or pack source that is custom, forked, operator-edited, or has unproven/mismatched provenance.

### 5. Unspecified TOML Editing Substrate (lines 396–397)

Comment and formatting preservation is asserted as a test condition (lines 659–662) but has no corresponding mechanism named in Proposed Implementation. Since the existing standard Go TOML library (`BurntSushi/toml`) used in the tree cannot preserve comments and formatting, any attempt to perform "preserve or refuse" will default to refusing 100% of TOML edits, making the doctor auto-fix completely inert on real operator files.
- **Required Resolution:** Explicitly name a format-preserving TOML editor (or concrete token-span patcher) in the Proposed Implementation, or explicitly state that the initial doctor slice will ship refuse-only TOML mutation behavior.

### 6. Recovery Journal fsync Durability (lines 385–387, 570–573)

The recovery journal is the absolute source of truth for rolling forward or rolling back on crash. However, the plan does not specify that the recovery journal must be explicitly `fsync`'d to disk before the first publish step. If a power loss occurs while the write is buffered in page cache, the journal will be lost or torn, leaving the city in an unrecoverable, half-migrated state.
- **Required Resolution:** Specify that the recovery journal must be written and explicitly `fsync`'d to disk before any publish step is initiated.

---

## Missing Evidence

1. **Commit Point Operation:** The exact durable filesystem operation (e.g., commit marker file write/rename) that linearizes multi-file fixes.
2. **Original-Byte Backups:** Specification of whether original file bytes are copied to the recovery journal to make the promised rollback mechanism physically possible.
3. **Controller Lock Lifecycle:** Explicit assertion that the running controller holds the same city advisory lock continuously during its lifecycle.
4. **Explicit Refusal Rules:** Explicit refusal rules for doctor fixes targeting custom/fork or unproven imports.
5. **TOML Library Selection:** A named library or concrete span-patching strategy that will be used to achieve comment and array preservation.
6. **Durable fsync of Journal:** Explicit specification of `fsync` syncing of the write-ahead recovery journal prior to any publishing step.

---

## Required Changes

1. **Specify the Single Commit Point & Recovery Strategy:** Name the linearizing filesystem operation (e.g., a specific commit marker file write/rename). Classify target renames as pre-commit (rollback on crash) or post-commit (idempotent roll-forward on restart), and if rollback is supported, mandate copying original bytes to the recovery record.
2. **Reconcile Loader Repair with Only-Writer Invariant:** Eliminate the load-path write-repair bypass or route it through the coordinator's city advisory lock to prevent concurrent mutation races.
3. **Mandate Controller Lock Ownership:** Require that the running controller acquires and holds the city directory advisory lock continuously, ensuring true mutual exclusion.
4. **Add Explicit Provenance Refusal Rules:** The doctor must consult `internal/packsource` and refuse to rewrite any import/pack source that is custom/fork or cannot be proven system-generated.
5. **Name the TOML Strategy:** Name the format-preserving TOML parser/editor in the Proposed Implementation, or explicitly state that the initial slice will ship refuse-only TOML behavior.
6. **Mandate fsync of the Journal:** Specify that the recovery journal is written and explicitly `fsync`'d before any publish or edit operation is initiated.

---

## Responses to Lane-Specific Questions

### Q1: Do all doctor --fix paths stage FixIntent objects before mutation and hold a city advisory lock across stage, validate, compare-before-rename, and publish?

**Answer:** 
The proposed design consolidates all writes under `internal/doctorfix` and mandates acquiring a crash-released city advisory lock *before* checking preflight digests. It holds this lock continuously across staging, target-digest verification, and renaming.
However, this safety isolation is compromised because the normal load-path required-pack repair (lines 536–539) mutates the installed pack directories during normal execution *without* acquiring the lock, violating the only-writer invariant. For full safety, either load-path repair must be removed, or it must be routed through the same lock.

---

### Q2: When scoped TOML edits cannot preserve comments or unknown fields byte-for-byte outside intended changes, does the fixer refuse rather than rewrite whole files?

**Answer:** 
The plan asserts in Testing (lines 659–662) that the automatic fix is refused if comments, unknown tables/fields, array order, or formatting cannot be preserved.
However, because standard Go TOML libraries (`BurntSushi/toml`) do not support comment preservation, this "preserve or refuse" guard will collapse to refusing 100% of TOML edits in real-world environments. To avoid an inert automatic fix surface, the plan must explicitly name a format-preserving TOML parser or token patcher in the Proposed Implementation.

---

### Q3: What recovery is specified for crashes or concurrent old and new binaries between per-file renames so cities cannot remain half-migrated?

**Answer:** 
- **For Crashes:** The design introduces a durable write-ahead recovery ledger written before publication. If a crash occurs pre-commit, the system rolls back or reruns; if post-commit, it converges forward. However, rollback is unimplementable because the schema lacks original-byte backups. Recovery must be redefined as roll-forward-only, or original bytes must be backed up, and the commit point must be explicitly named as a single linearizing filesystem write.
- **For Concurrent Old Binaries:** The new binary records a migration marker that detects old-binary post-marker writes and raises version-skew diagnostics. However, because legacy binaries are not lock-aware, they can still write to a city *during* an active migration. An explicit lock or freeze file is needed to ensure old-binary exclusion.
