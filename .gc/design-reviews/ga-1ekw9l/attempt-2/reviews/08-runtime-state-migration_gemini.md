# Yelena Markovic — Gemini (Maintenance to Core Runtime State Migration Reviewer, Attempt 2, Independent Review)

**Verdict:** block

**Lane:** Maintenance to Core runtime state migration, JSONL and archive state, spawn-storm ledgers, non-destructive markers, old/new binary concurrency, downgrade continuity.

Reviewed against the Iteration 2 / Attempt 2 draft of `plans/core-gastown-pack-migration/implementation-plan.md` and `plans/core-gastown-pack-migration/requirements.md` in the active repository workspace.

---

## Executive Summary

As Yelena Markovic, the **Maintenance to Core Runtime State Migration Reviewer**, I have conducted an independent, evidence-backed safety and risk audit of the Iteration 2 (Attempt 2) design. My verdict is **Verdict: block**.

While the Iteration 2 draft has adopted the required schema layout and incorporated the necessary vocabulary ("shared lock acquisition," "staged archive-copy digest checks," "push-cursor reconciliation," "post-marker old-binary write detection"), these concepts are currently treated as slogans rather than specifications. 

The Doctor Mutation Safety lane (Leah Okafor) approved the general mutation machinery in Attempt 2. However, that approval assumes that the domain-specific runtime-state migration has been safely specified. Under closer inspection from a runtime-state perspective, the design contains critical, self-contradictory locking rules, unaddressed split-brain count windows during version skew, undefined push reconciliation algorithms for git-over-shared-remotes, and a lack of source-grounded file and field mappings. 

We cannot approve decomposition for this lane until these safety-critical gaps are closed with concrete engineering specifications.

---

## Top Strengths

- **Explicit Recognition of Legacy State Families (§266–268, §377–380)**: The plan correctly maps all at-risk runtime-state categories: JSONL archive state, spawn-storm ledgers, refs/remotes, escalation fields, pending archive push state, and order skip/tracking compatibility aliases.
- **Introduction of the Migration Marker (§269–271, §377–381)**: Recording schema version, old/new paths, and staged archive digests inside a dedicated marker provides a solid foundation for detecting post-migration conflicts.
- **Commit-Point and Staged Digest Checks (§260–264, §269–271)**: Staging archive copies and performing pre-rename digest checks protects the migration from treating a truncated or interrupted file copy as a completed state.
- **Non-destructive Directory Preservation (§388–391)**: Agreeing not to auto-delete stale legacy `.gc/runtime/packs/maintenance` directories prevents catastrophic operator data loss during upgrading or rollback.

---

## Critical Risks & Gaps (The Blockers)

### 1. Self-Contradictory Migration Trigger & Lock Domain (§256–257 vs §268–269)
The design asserts that the mutation coordinator "refuses with manual guidance" if "a controller for the same city is running" (§256–257). However:
- **The Conundrum**: Runtime-state migration is a startup-time concern. When a new-binary controller boots and detects legacy old-path state, it must migrate it before any orders, scripts, or agents execute. 
- **The Contradiction**: If the controller triggers the coordinator at startup, then the controller is running, which violates the coordinator's own rule to refuse automatic fixes.
- **The Race**: If migration is strictly doctor-only (`gc doctor --fix`), the plan must explicitly state that upgrading requires the controller to be stopped. It must also specify what a freshly started new controller does when it boots on un-migrated legacy state: does it refuse to run, does it orphan the legacy history and start a split-brain fresh ledger, or does it silently execute without history?
Currently, the "shared lock acquisition" (§268) is undefined and never integrated with the coordinator's city advisory lock.

### 2. Slogan-Level "Push-Cursor Reconciliation" (§267, §378)
The JSONL archive is a live git repository pushing to a shared remote (typically containing `refs/remotes/origin/main`). 
- **The Risk**: If an operator downgrades or runs an interleaved old binary, the old binary will check its legacy state file, find `pending_archive_push` is true (or set it true on action), and push commits to the shared remote. Meanwhile, the new binary in Core-owned paths will do the same.
- **The Gap**: The plan names "push-cursor reconciliation" as a recorded field, but never specifies the actual *rule* to prevent duplicate offsite pushes. It must define whether the migration repoints/clears the legacy remote on copy, or makes the push routine idempotent against a shared remote marker. Merely suppressing the new binary's version-skew diagnostic does not prevent double-pushes of the same commit refs.

### 3. Spawn-Storm Ledger Split-Brain Throttling Bypass (§380–381)
The spawn-storm ledger is a per-path sliding counter (`spawn-storm-counts.json`). 
- **The Risk**: During version skew (where old and new binaries run concurrently or alternate), the old binary reads/increments the legacy ledger while the new binary reads/increments the Core ledger.
- **The Gap**: The design states that "retained legacy state is ignored unless the marker or digest checks show conflict" (§380–381). If the new binary ignores the legacy ledger, then neither binary sees the combined count. A genuine spawn storm driven by both binaries will stay under their respective individual thresholds, completely bypassing the storm-throttling safety mechanism. 
- **Required Specification**: The new binary must read-union the legacy ledger counts into its threshold evaluation until the skew window has closed.

### 4. Regression of Source Grounding & Decomposition Readiness
The proposed implementation details (§266–274) describe the migration at a slogan level. 
- **The Gap**: It fails to name the concrete scripts (`jsonl-export.sh`, `spawn-storm-detect.sh`), Go files (`cmd/gc/jsonl_archive_doctor_check.go`), specific JSON fields (`pending_archive_push`, `consecutive_push_failures`), or paths (`.gc/jsonl-archive`, `.gc/runtime/packs/maintenance/jsonl-archive`, `spawn-storm-counts.json`).
- **The Impact**: A decomposer cannot safely generate implementation beads from this high-level text without guessing the target files and boundaries. Grounding must be restored to the Proposed Implementation section.

### 5. Overpromised "Deterministic Re-Upgrade Flow" (§272–273, §379–380)
The plan proposes a "deterministic re-upgrade flow" when an old binary has written to legacy paths after the migration marker.
- **The Gap**: This is mathematically unsound. If both the legacy append-style stores (JSONL archive, spawn-storm ledger) and the new Core-owned stores have post-marker writes, their histories have diverged. Merging them requires a manual reconciliation or a non-deterministic merge rule. 
- **Required Specification**: The design must restrict "deterministic re-upgrade" strictly to cases where the new Core-owned paths are completely untouched (pure rollback-then-re-upgrade), and mandate manual operator intervention if both sides have diverged.

### 6. Copy-vs-Move Ambiguity (§266 vs §380)
The plan asserts that migration "moves JSONL archive state" (§266), yet relies on "retained legacy state is ignored" (§380). If the state is truly *moved* (source deleted/renamed), then there is no legacy state left to be ignored or to detect old-binary writes against. If it is copied (non-destructive), it must be explicitly defined as a copy-then-mark protocol to maintain downgrade continuity.

---

## Required Changes

To unblock this lane for decomposition:

1. **Reconcile the Trigger & Lock Domains**: 
   - Define a single trigger for runtime-state migration (e.g., must run under `gc doctor --fix` with the controller stopped, or under a startup lock in the controller).
   - Specify how the "controller is running" check is bypassed *only* for the controller's own startup-migration path under the shared advisory lock.
   - Specify stage-to-temp-dir-plus-atomic-rename for the archive directory to prevent half-copied `.git` folders from being read as active.
2. **Define the Push Reconciliation Rule**:
   - Specify a concrete rule making duplicate offsite pushes impossible (e.g., "the migration repoints the legacy archive's remote to a dummy local URL to prevent legacy pushes, or clears its push flags").
3. **Specify Ledger Read-Union Throttling**:
   - Explicitly specify that the new binary's threshold evaluator read-unions the retained legacy ledger count during the version-skew window.
4. **Restore Source Grounding**:
   - Add the concrete file names (`jsonl-export.sh`, `spawn-storm-detect.sh`), JSON fields (`pending_archive_push`), and paths (`.gc/jsonl-archive`) to the Proposed Implementation section.
5. **Scope the Re-Upgrade and Clarify Copy-vs-Move**:
   - Limit the "deterministic re-upgrade" to the untouched-new-path case; otherwise, require manual reconciliation.
   - Define the migration as a non-destructive copy-then-mark protocol so that downgrade binaries can still function.

---

## Technical Audit: Lane-Specific Questions

### Q1: Are jsonl state, archive remotes and refs, spawn-storm ledgers, order skip lists, and escalation fields migrated or preserved with non-destructive markers?
**Answer**: The *intent* is represented in the data structure (§377–381), but the *mechanism* is contradictory. The plan claims to "move" the state (§266) while expecting it to be "retained legacy state" (§380). To safely preserve these fields under a non-destructive marker, the migration must be specified as a copy-and-mark protocol, leaving the legacy directories intact for downgrade binaries while shifting active operations to the Core paths.

### Q2: If an old binary writes legacy Maintenance state after the migration-completed marker, how does the new binary detect and handle divergence without data loss?
**Answer**: The design asserts that the new binary will compare the old paths against the marker and report a version-skew diagnostic (§271–273). However, it fails to specify *when* this check occurs. If it only occurs during a doctor check, a running controller will continue running with split-brain ledgers and archives. Furthermore, the "deterministic re-upgrade" is overpromised; if both sides have written, automatic deterministic merging is impossible without data loss, and manual operator reconciliation must be enforced.

### Q3: What round-trip test covers interrupted archive copy, in-flight push state, duplicate writers, and rollback to an older binary?
**Answer**: The testing section (§435–440) mentions failure injection and detecting post-marker writes, but lacks a high-fidelity concurrent test fixture. To prove this safety boundary, the test plan must specify a dedicated test case covering:
1. Interrupted copy resulting in a partial `.git` directory.
2. An old binary attempting a push to the shared remote concurrently with a new binary.
3. A complete rollback loop verifying that the old binary can successfully resume writing and pushing from its legacy paths.
