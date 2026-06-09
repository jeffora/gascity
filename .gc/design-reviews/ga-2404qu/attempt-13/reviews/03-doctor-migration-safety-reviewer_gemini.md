# Sofia Khoury — DeepSeek V4 Flash Independent Review (Iteration 13 / Attempt 13)

**Verdict:** approve-with-risks

**Lane:** doctor fix idempotency, legacy import rewrites, custom data preservation, operator-safe diagnostics.

Reviewed against the Iteration 13 / Attempt 13 draft of `design.md` (`updated_at: 2026-06-07T14:05:04Z`) and `requirements.md` in the active repository workspace.

---

## Executive Summary

The Iteration 13 / Attempt 13 design integrates major architectural advancements, notably the **Enforceable Doctor Mutation Protocol (§1852–1891)** and a crash-released **OS-level directory lock (§1873–1875)**. These changes successfully address several concurrency and failure-atomicity loopholes identified in earlier iterations.

However, from an independent, highly critical safety-and-risk auditing perspective, the current text still exhibits **severe rollout sequencing hazards, a permanent post-migration conflict loop, TOCTOU concurrency windows, and air-gap incompatibilities**. 

Crucially, the non-destructive copy-semantics intended to protect legacy state introduces a spurious, permanent manual conflict for healthy upgraded cities on subsequent doctor checks. Additionally, the rollout schedule leaves operators exposed to unsafe, lossy TOML rewriting during Slices 2 and 3.

This review presents an evidence-backed analysis of these critical risks and outlines the concrete, actionable changes required before the doctor migration subsystem can be safely implemented.

---

## Top Strengths

1. **OS-Level Crash-Released Directory Lock (§1873–1875):**
   Adopting a crash-released OS advisory lock on the city directory's file descriptor instead of relying on fragile, stale-prone PID files or persistent lock files is an outstanding, first-principles decision. This guarantees that concurrent doctor runs are serialized cleanly and that a crashed doctor session never leaves the city in a permanently locked state.
2. **True Non-Destructive Copy Semantics for Runtime State (§1884–1886):**
   Explicitly moving from "Move" to "Copy" semantics for JSONL state/archive and spawn-storm ledgers fully aligns the implementation with the core project invariant: *legacy Maintenance state is never deleted by `gc doctor --fix`*. This ensures that back-level binaries can be rolled back safely and continue executing against their frozen legacy states.
3. **Comprehensive Staged Mutation Coordinator (§1571–1605, §1852–1883):**
   Enforcing that all writes must return a structured `FixIntent` to be executed by a single, central `doctor.MutationCoordinator` is a major win for system stability. Preventing direct, ad-hoc writes from individual checks enforces a unified preflight and staging pipeline.
4. **CST/Span-Preserving TOML Refusal Gate (§1263–1266, §1596–1598):**
   The mandate to fail-closed and refuse automatic fixes if comments, custom formatting, array order, or unknown fields cannot be preserved via CST-span edits guarantees that user-curated manifests are never mangled by automated operations.

---

## Critical Risks & Gaps

### 1. [Major] The Post-Migration Spurious Conflict Loop (§1884–1888)
* **The Risk:** Under non-destructive copy-semantics, legacy JSONL state and spawn-storm files are copied to their new Core paths, but the legacy files remain in place as "ignored legacy state" (§1886). The protocol then mandates: *"If both sides exist and their digests differ, automatic fix refuses with a manual conflict diagnostic"* (§1887).
* **The Gap:** Once a city is successfully upgraded, it is active. The live Core files (JSONL logs, spawn-storm counts) will naturally be modified and appended to as the city executes workflows. The legacy ignored Maintenance files, being frozen, will remain unchanged. Consequently, their content digests **will inevitably differ** immediately after the first active run. On any subsequent routine execution of `gc doctor` (or automatic pre-command doctor checks), the doctor will find that both files exist and differ, triggering a **permanent, recurring manual conflict diagnostic that disables automatic fixes** for an entirely healthy, active city.
* **Required Change:** Restructure the conflict rule so that digest comparison is only executed if a migration has not yet occurred (i.e., if Core files do not yet exist or have not been initialized). Once the participation record or publish record proves the city was migrated, subsequent doctor checks must ignore differences between active Core state and legacy backups.

### 2. [Major] Rollout Sequencing Exposes Legacy Unsafe Fixes (Slices 2–4, §2625–2648)
* **The Risk:** Slice 2 updates `PublicGastownPackVersion` to the compatibility commit, updating the target that existing cities must adopt (§2625). However, the `doctor.MutationCoordinator`, preflight validation, and CST-preserving TOML editor do not land until Slice 4 (§2644).
* **The Gap:** Because implementation slices must be independently deployable, an operator running a Slice 2 or Slice 3 binary who encounters a legacy import diagnostic will run `gc doctor --fix` using the **legacy, unsafe doctor machinery** (`cmd/gc/import_state_doctor_check.go`). This old code rewrites TOML manifests via whole-file re-marshalling (destroying operator comments/formatting) and lacks reachability checks, OS-level locking, and transactional atomicity. Operators will have their manifests corrupted and their comments destroyed by the legacy fixer while attempting to adopt the Slice 2 public pin.
* **Required Change:** Move the `doctor.MutationCoordinator` and CST-preserving TOML editor into Slice 2, or explicitly disable/freeze automatic imports-rewriting in the legacy doctor check during Slices 2 and 3, requiring manual operator guidance until the safe coordinator lands in Slice 4.

### 3. [Major] Advisory OS Locking TOCTOU Race Condition (§1873–1879)
* **The Risk:** The coordinator takes a crash-released OS advisory lock in Step 3, but this occurs *after* the Step 1 phase-one report and Step 2 preflight phases are completed (§1867–1875).
* **The Gap:** This sequencing introduces a classic Time-of-Check to Time-of-Use (TOCTOU) race condition:
  1. Two concurrent `gc doctor --fix` processes (A and B) run. Both complete Step 2 (preflight) successfully, reading identical file digests.
  2. Process A enters Step 3, acquires the OS directory lock, and proceeds to stage and publish its renames.
  3. Process B blocks at Step 3, waiting for the lock.
  4. Process A completes, writes its changes to the manifest, and releases the lock.
  5. Process B acquires the lock. Since its preflight was already completed *before* Process A wrote, Process B proceeds to stage and publish its stale candidates, silently overwriting Process A's changes and corrupting the configuration.
* **Required Change:** The OS advisory lock must be acquired **before** Step 2 (preflight) begins, or the coordinator must perform an explicit post-lock verification re-reading all preflight file digests immediately before staging and writing candidates.

### 4. [Major] Silent Overwrite of Operator Edits During Core Repair (§1909–1910, §2381–2384)
* **The Risk:** The doctor is mandated to repair missing, corrupt, or tampered Core files to maintain strict required-Core integrity. 
* **The Gap:** If an operator mistakenly edited files within `.gc/system/packs/core` (e.g., custom scripts, prompt overlays, or experimental formulas), the automated repair will silently regenerate the Core pack and overwrite their work, leading to permanent, silent loss of custom data.
* **Required Change:** The Core materializer/repair subsystem must back up any tampered files (e.g., staging them in a `.gc/staged-recovery/` or `.gc/system/packs/core.bak/` directory) before regenerating Core, or the doctor must report a warning and require an explicit `--force-repair` flag if custom-edited files are detected.

### 5. [Major] Air-Gapped Network Reachability vs. Cache-Satisfied Contradiction (§1869, §1902–1906)
* **The Risk:** Step 2 of the coordinator preflight mandates that the doctor *"validate public Gastown reachability"* (§1869). 
* **The Gap:** In a strictly air-gapped or disconnected environment, any outbound network reachability check to a remote registry will fail unconditionally. However, if the operator has pre-populated the local ordinary remote cache with the exact pinned Gastown commit, the city is fully installable and safe offline. The hard reachability check will block offline operators from executing repairs even when they possess a complete, lock-consistent local cache.
* **Required Change:** Clarify that preflight "reachability" is satisfied if *either* the remote registry is reachable *or* the exact pinned public Gastown commit is already present and validated in the local ordinary remote cache.

### 6. [Major] Generated-Source Provenance Verification Mechanism Gap (§1589, §1871)
* **The Risk:** The design requires "generated-source provenance" to prove a legacy import was generated by the binary (and not custom-edited) before permitting auto-removal.
* **The Gap:** The text fails to specify how this provenance is computed for directories and imports created by older binary versions, whose content digests cannot match the current binary's embedded digests. Without a specified mechanism (e.g., using recorded hashes in `install-lock.toml` or embedding a static table of legacy release digests), implementers are highly likely to fall back to forbidden path-based heuristics.
* **Required Change:** Explicitly define the provenance mechanism: either mandate checking recorded digests in `install-lock.toml`, or embed a historical table of content digests of the final versions of the retired Maintenance and local Gastown packs within the new binary.

---

## Evaluation of Sofia's Critical Questions

### 1. Is the Core presence doctor fix a proven no-op on a healthy city, including repeated or concurrent runs with a controller active?
**Yes, mostly, but with a critical caveat.** If the city is healthy, the pre-resolution pass and preflight checks will stage zero intents, resulting in a byte-identical no-op. Concurrent runs with an active controller are safely blocked by Step 3's concurrency check (§1873). However, concurrent doctor-vs-doctor runs are susceptible to the TOCTOU race detailed in Risk 3, which could result in a non-no-op collision if one process overwrites a newly written manifest with stale staged candidates.

### 2. When `gc doctor --fix` removes redundant Core or legacy Maintenance imports, what prevents it from deleting user-added imports or custom pack edits?
**The CST-preserving TOML editor (§1863) and the Retired-Source Classifier (§1225–1230) are effective guards.** User-added imports are classified as "custom local forks" and are bypassed by the automated editor, routing to manual diagnostics. The CST editor ensures that if untouched user-added tables, comments, or formatting cannot be preserved, the coordinator aborts and refuses the fix.

### 3. If a local Gastown import is rewritten to a public remote, does the fix verify reachability and immutable provenance or fail with explicit operator guidance?
**Yes.** Preflight (§1869–1871) requires validating exact pin installability, ordinary remote cache identity, and lock parseability. If these fail (e.g., in a disconnected environment without a pre-populated cache), the fix fails gracefully and outputs explicit offline operator guidance (§1862).

---

## Required Changes for Finalization

1. **Resolve Post-Migration Spurious Conflict Loop:** Update §1887 to specify that digest comparison and manual conflict checks for JSONL/spawn-storm files are only performed *prior* to initial migration. Once active Core state is initialized, differences with the ignored legacy backup must not trigger manual conflicts on subsequent doctor runs.
2. **Close Slice 2–3 Unsafe Fix Window:** Move the `doctor.MutationCoordinator` and CST-preserving editor into Slice 2, or freeze/disable automatic import rewrites in the legacy doctor check during Slices 2 and 3.
3. **Correct TOCTOU Lock Sequencing:** Mandate that the OS advisory lock (§1873) is acquired *before* Step 2 (preflight) begins, or enforce a post-lock re-verification of all preflight file digests before rename/publish.
4. **Protect Custom Core Modifications:** Require the Core repair system to back up tampered Core files to a `.gc/staged-recovery/` or `.gc/system/packs/core.bak/` directory before overwriting them.
5. **Define Air-Gap Preflight Cache Exception:** Enforce that the preflight reachability check is bypassed/satisfied if the exact pinned public Gastown commit is already present in the ordinary remote cache.
6. **Define Legacy Provenance Verification:** Explicitly specify the use of recorded digests in `install-lock.toml` or a historical digest dictionary embedded in the binary to verify generated-source provenance.

---

## Questions

1. **Does the post-publish revalidation (§1880) rerun the exact same diagnostics as preflight?** If so, how do we prevent the revalidation from throwing a false conflict on the newly copied JSONL/spawn-storm files, which now exist in both locations?
2. **What is the cleanup lifecycle of stale candidate files in `.gc/tmp` after a partial publish failure?** Does the next successful doctor run or a normal `gc` command execution automatically sweep them, or is it a manual operator task?
3. **If an operator has a dirty worktree in a custom local pack, does the doctor refuse all unrelated manifest fixes?** Or is the "dirty worktree" classification scoped strictly to the packs being modified or rewritten?
