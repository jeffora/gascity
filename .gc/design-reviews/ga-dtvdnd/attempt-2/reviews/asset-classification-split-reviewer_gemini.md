# Hugo Bautista — DeepSeek V4 Flash (Asset Classification & Split Review)

**Verdict:** approve-with-risks

**Scope:** Split and core-renamed ownership; file-by-file migration table completeness; clean halves; resolution of review-marked assets.

Reviewed against the Attempt 2 requirements document (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md` updated 2026-06-09T01:20:00Z).

---

## Executive Summary

The Attempt 2 requirements document for the Core and Gastown Pack Split represents an exceptional and highly disciplined evolution from Attempt 1. In Attempt 1, I issued a strict **BLOCK** due to a stale, static file-by-file table embedded in the requirements body. This table not only contained multiple nonexistent/phantom files but also violated the requirements schema and left critical split boundaries and divergent duplicates completely un-reconciled.

In this revision, the authors have elegantly resolved these structural and behavioral issues:
- The requirements document has been restructured to comply fully with the `gc.mayor.requirements.v1` schema, successfully purging the raw implementation-level file list from the requirements body.
- In its place, **AC6** and **AC7** establish a dynamic, automated, and machine-validated **Asset Migration Ledger** and **Behavior-Preservation Manifest** contract.
- These contracts mandate that the design phase must programmatically validate file presence, ensure 100% directory coverage, resolve all `review` rows, prevent duplicate or orphaned behavior, and verify triggerability before any implementation begins.

By transitioning the file-by-file mapping from a static, unverified prose list to a dynamic, executable design contract, this requirements document aligns perfectly with the Zero Framework Cognition (ZFC) and Bitter Lesson principles. I am pleased to upgrade my verdict to **APPROVE-WITH-RISKS**. Downstream implementation and design planners must now carefully address a few minor edge cases and validation-chain boundaries.

---

## Top Strengths (Grounded Evidence)

1. **Elegant Resolution of Schema Conflict (AC1 & AC6)**: Purging the inline raw markdown table and delegating it to an external ledger contract fully resolves the compliance blocker raised by Mara Voss (Requirements Schema Compliance) while retaining strict, auditable file-granularity ownership as a mandatory gate for implementation planning.
2. **Robust Automated Ledger Validation Contract (AC6)**: Rather than hoping a developer manually tracks every file correctly, AC6 requires that the ledger-validation engine programmatically fail on:
   - **Missing current paths** (completely eliminating the phantom/deleted-file defect from Attempt 1).
   - **Unrepresented active source files** (guaranteeing that 100% of files are classified and preventing wholesale, un-audited bucket moves).
   - **Unresolved `review` rows** (preventing deferred decisions from sliding into implementation).
   - **Duplicated or orphaned split behavior**.
3. **Explicit Behavior-Preservation Manifest (AC7)**: Mandating a behavior-preservation manifest that verifiably asserts that stripped Core behavior is re-homed and triggerable under Gastown ensures Oleg Marchetti's (Behavior Preservation) invariants are satisfied prior to merging any code.
4. **Deterministic Handling of Conditionals (AC6 fallback classification)**: Requiring a `fallback classification` field in AC6 ensures that any conditionally-classified `core` asset (e.g., "keep only if role-neutral") has a pre-determined, non-interactive resolution path if a role-neutrality violation is detected.

---

## Critical Risks & Gaps

### 1. [Major] Lack of Git-Tracked Baseline Definition for Ledger Validation
* **The Risk**: AC6 requires that the ledger validation check fails on "unrepresented active source files" to ensure 100% coverage. However, the requirements do not specify *how* the validator discovers the set of active source files. If the validator scans raw directories on disk, local untracked files, local test cache files, or git-ignored assets will cause the validator to fail in local developer environments while passing in CI, or vice versa.
* **Mitigation Recommendation for Design**: The ledger validator must use a deterministic, git-backed command (such as `git ls-files` or a specific commit tree snapshot) to establish the authoritative set of active files under the legacy pack directories, explicitly excluding ignored paths, `.git`, and build outputs.

### 2. [Major] Path Mismatch and Collision in Split Asset Outputs
* **The Risk**: For files classified as `split` (such as template fragments and skills) or `core-renamed`, the asset is divided into a Core generic output and a Gastown-specific output. Because Core and Gastown reside in two completely separate repositories (`gascity` and `gascity-packs/gastown`), a single `target output path` field in the ledger schema is structurally insufficient. If an implementer provides only a single path, one side of the split will be left undefined, resulting in orphaned behavior or path collisions.
* **Mitigation Recommendation for Design**: The Asset Migration Ledger schema must explicitly require **two distinct output relative paths** for any file classified as `split` or `core-renamed`: a `target_core_path` relative to the Gas City repository root, and a `target_gastown_path` relative to the Gastown pack repository root, with explicit split boundaries defined for each.

### 3. [Minor] Basename Collision Risk for Divergent Duplicate Fragments
* **The Risk**: Three critical prompt-template fragments exist in both legacy packs with identical basenames but divergent content (e.g. `architecture.template.md`, `following-mol.template.md`, and `propulsion.template.md`). A naive ledger validator might verify that both files exist and have rows, but miss that they represent divergent duplicates. Without an explicit basename collision check, the extraction script could silently overwrite one with the other or select the incorrect canonical copy.
* **Mitigation Recommendation for Design**: The ledger validation tool must incorporate a "basename collision scanner" that flags same-named files across the three legacy pack roots, forcing the design document to declare a definitive merge/override policy for each pair.

---

## Required Changes & Actions

1. **Define Git-Tracked Baseline for Validation**:
   Refine the ledger validation contract under AC6 to specify that the set of "active source files" must be derived strictly from git-tracked files (`git ls-files` under the legacy directories) at a specific snapshot, preventing local untracked file pollution from breaking developer builds.
2. **Mandate Dual-Destination Fields for Splits**:
   Specify that the Asset Migration Ledger schema must declare both `target_core_path` and `target_gastown_path` relative targets for all `split` and `core-renamed` rows, preventing orphaned behavior or path collisions across the two target repositories.
3. **Incorporate Basename Collision Scanning**:
   Require the design-stage validation tool to actively scan for and flag duplicate filenames across legacy pack paths to enforce explicit, documented reconciliation policies for divergent template fragments.

---

## Questions

- Will the Asset Migration Ledger and its validation commands support explicit exclusion patterns (e.g., ignoring `.gitignore`, README files, or non-behavioral test fixtures), or must every single file under the legacy pack roots have an explicit row in the ledger?
- For `split` template fragments, will the behavior-preservation harness verify that the resulting split prompts render and parse correctly in their respective target packs under real-world compilation/runtime states?
