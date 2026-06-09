# Simone Kaye — DeepSeek V4 Flash (External Pack & Docs Review) — Iteration 5 / Attempt 5

**Verdict:** approve-with-risks

**Lane:** external Gastown pack authority, registry behavior, source-tree cleanliness, documentation consistency.

---

> ### Lane Note (Verify-Don't-Copy & Dual-Placement Strategy)
> 1. **Re-Grounding & Independence:** This review is an independent DeepSeek V4 Flash evaluation of the proposed Requirements (`plans/core-gastown-pack-migration/requirements.md` updated 2026-06-09T08:58:00Z) and the corresponding Implementation Plan (`plans/core-gastown-pack-migration/implementation-plan.md` updated 2026-06-09T07:28:00Z). My findings are technically grounded, fresh, and re-verified directly against the repository's live state.
> 2. **Dual-Placement Strategy:** Due to the known workflow defect where the active bead's metadata targets `attempt-1/reviews/` while other active reviewers target `attempt-5/reviews/`, I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/external-pack-docs-reviewer_gemini.md` and `.gc/design-reviews/ga-dtvdnd/attempt-5/reviews/external-pack-docs-reviewer_gemini.md` to ensure all downstream synthesis steps resolve the file correctly.

---

## Executive Summary

The Requirements (Attempt 5) and corresponding Implementation Plan represent an impressive, robust architectural blueprint to decouple Gas City's Core SDK required behavior from optional, user-supplied Gastown behavior. By establishing `internal/systempacks` as the sole loading authority, retiring the standalone Maintenance system pack, and moving Gastown behavior to the public `gascity-packs/gastown` repository, the system successfully eliminates untraceable legacy fallbacks and framework-side role hardcoding.

However, from the strict perspective of **External Pack Authority & Documentation Consistency**, several key edge cases and systemic gaps remain unaddressed. Specifically, the Requirements document treats legacy source-tree cleanliness as a presentation constraint (AC12) rather than a hard structural acceptance criterion (AC), leaving a potential backdoor for legacy path resolution. Additionally, the shared cache promotion model lacks a global lock mechanism, exposing the system to concurrent write corruption, and the wording scanner's permissiveness in AC12 leaves a manual "grep" escape route.

To address these concerns, I award an **APPROVE-WITH-RISKS** verdict and mandate four critical required changes before requirements can transition to implementation approval.

---

## Evaluation of the Three Key Questions

### Q1: Is the transition of Gastown authority to `gascity-packs/gastown` robust and fail-closed?
**Yes, but with cache safety concerns.**
* **Strengths:** AC15 ensures that the public Gastown version, the pin ledger, packcompat proof, lock/cache provenance, and fresh-init output must agree on the source, subpath, exact commit, pack digest, and behavior-manifest digest. Branch or tag references are fetchability metadata only. If the exact pinned external pack is missing under offline conditions, AC16 correctly fails-closed with an actionable diagnostic and never falls back to in-tree examples or synthetic aliases.
* **Risks:** The implementation plan relies on city-scoped advisory locks for mutation safety. However, since the public pack cache is a shared, system-wide directory, concurrent CLI operations from different cities or parallel test runners in CI can collide during cache promotion, risking partial directory writes and subsequent silent corrupt cache-hit loads.

### Q2: Does the design ensure complete clean-up of legacy in-tree packs without breaking the codebase?
**Partially.**
* **Strengths:** AC4 guarantees that fresh Gastown initialization imports the public pack explicitly and does not rely on legacy in-tree example paths or bundled synthetic public aliases. AC3 ensures that any attempt by a user to shadow required Core with another pack named `core` fails-closed immediately with a collision diagnostic.
* **Gaps:** The Requirements document never explicitly mandates the deletion or demotion of `examples/gastown/packs/gastown` and `examples/gastown/packs/maintenance`. While the Implementation Plan plans to delete them in Slice 7, the Requirements treat their presence solely as a documentation presentation constraint (AC12: "do not present... as authoritative current sources"), leaving their structural disposition uncodified.

### Q3: Is the documentation consistency model robustly enforced?
**Partially.**
* **Strengths:** AC12 successfully establishes a terminology matrix defining Core as required, Gastown as external-optional, and Maintenance as retired, ensuring doctor messages, help screens, and CLI examples are updated in the same slice as the behavior change.
* **Gaps:** The verification column in AC12 allows a permissive manual "grep" run or an audit artifact rather than strictly mandating an automated, test-suite-enforced scanner. Furthermore, Go comment blocks and string literals are omitted from the wording scanner target list, allowing stale path references to rot in Go files.

---

## Critical Risks & Architectural Gaps

### 1. [Major] Missing Source-Tree Deletion/Demotion Acceptance Criterion (The Disposition Gap)
* **The Risk:** AC12 forbids documentation and examples from *presenting* `examples/gastown/packs/*` as authoritative current sources, but no Acceptance Criterion explicitly mandates the **complete physical deletion** or **exclusion from resolution** of the legacy source trees themselves.
* **The Gap:** Without a structural AC, an implementer can leave `examples/gastown/packs/gastown` (with its live `../maintenance` import) in the repository as a "silent fallback," relying solely on documentation constraints to prevent its use. If an operator configures a legacy city with a local import path, the engine will still resolve it, bypassing the external pack decoupling.
* **The Fix:** Create a new Acceptance Criterion (or amend AC5) to explicitly mandate that the `examples/gastown/packs` directory must be physically deleted from the repository (or demoted to non-resolvable, isolated test-only fixtures) before the migration is considered complete, with a CI absence check to prevent re-introduction.

### 2. [Major] Global Lock Deficit for Concurrent Shared Cache Promotion (The Multi-City Collision Gap)
* **The Risk:** The Implementation Plan specifies that the doctor and mutation coordinator acquire a "city advisory lock" before digest preflight. However, the remote pack cache is a shared, system-wide or user-wide directory, not isolated to a single city.
* **The Gap:** If multiple cities (or parallel, sharded integration test runners in CI) execute concurrently, they will attempt to fetch, promote, or write to the *same* shared cache path. Because the advisory lock is city-scoped (bound to `.gc/` inside the city), it cannot prevent concurrent write collisions, file-handle races, or folder-overwrite corruption in the system-wide cache.
* **The Fix:** Require the resolver/mutation coordinator to acquire a separate, global file-based lock (e.g., a write lock on the shared cache directory or cache promotion temp path) before executing any write, fetch, or promotion to the shared cache.

### 3. [Major] Partial-Download Cache Corruption Vulnerability (The Cache Integrity Gap)
* **The Risk:** AC16 requires cache writes and promotions to be atomic, utilizing temp-directory staging and atomic rename operations.
* **The Gap:** If a write/promotion process is abruptly terminated (system crash, power loss, `SIGKILL`) in the split second *after* the target cache directory is created but *before* all files are completely written, the folder will exist on disk. Subsequent runs will treat this existing folder as a valid cache-hit, loading a corrupted, partial pack and causing silent, hard-to-debug runtime failures.
* **The Fix:** The cache resolver must never assume a cached folder is healthy merely because it exists. Every cached pack must require a complete integrity marker file (such as `.gc_provenance.json` containing a file manifest, SHA-256 hashes, and a completion timestamp) written as the final atomic step of promotion. If this marker is missing or fails verification, the resolver must treat the hit as a cache-miss, delete the corrupted directory, and fail-closed.

### 4. [Minor] Permissive Grep Audit Escape Route in AC12 (The Automation Gap)
* **The Risk:** The verification of AC12 allows "Docs/examples/help grep or generated audit artifact," leaving a manual escape route.
* **The Gap:** Manual greps are notoriously prone to omission during releases, allowing retired terminology and legacy path references to leak over time.
* **The Fix:** Mandate that the wording and terminology consistency matrix must be validated exclusively by an automated, test-suite-enforced scanner (integrated in the CI pipeline) that fails the build on any denied contexts, and extend its scope to check string literals and comment blocks in Go source files (`.go`).

---

## Required Changes for Finalization

1. **Source-Tree Disposition Invariant:** Add an explicit Acceptance Criterion mandating the physical removal of legacy source-tree packs under `examples/gastown/packs/*` (or their total demotion to isolated, non-resolvable fixtures) to guarantee no parallel authority remains in the tree.
2. **Global Cache Lock:** Update AC16 to require a global file lock on the shared cache promotion temp path to protect concurrent multi-city/test-suite cache promotions from write collisions.
3. **Cache Integrity Provenance Marker:** Update AC16 to mandate a cryptographically verified completion/integrity marker file (e.g., `.gc_provenance.json`) for cached packs, causing the resolver to fail-closed and quarantine any partial or corrupted cache folders.
4. **Automate Wording Scanner & Include Go Comments:** Amend AC12 to require that wording validation runs as an automated CI test suite scanner, and extend its target scope to include comment blocks and string literals inside `.go` files.

---

## Remaining Questions

1. Will the global cache lock support a timeout and fail-closed behavior to prevent concurrent CLI commands from hanging indefinitely in a deadlock state?
2. When the wording scanner detects a denied terminology token in a historic changelog or migration guide file, does the requirements plan allow an explicit, path-scoped exception allowlist with expiry metadata?
