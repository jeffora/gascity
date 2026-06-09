# Hugo Bautista — DeepSeek V4 Flash Perspective Independent Review (Iteration 6 / Attempt 6)

**Verdict:** approve-with-risks

**Scope:** Split and core-renamed ownership; file-by-file migration ledger completeness; clean halves; and resolution of review-marked assets.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding & Independence.** I have re-grounded this independent review against the active Attempt 6 requirements document (`plans/core-gastown-pack-migration/requirements.md` / `.gc/design-reviews/ga-dtvdnd/attempt-6/design-before.md`, 119 lines, updated 2026-06-09), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` retired assets, and the proposed implementation plan `plans/core-gastown-pack-migration/implementation-plan.md` (657 lines, updated 2026-06-09). I evaluated the criteria and verified the asset-split and de-roling hazards against the live repository tree.
2. **Dual-Placement Strategy.** Due to the known workflow defect where the bead's metadata `gc.attempt=1` causes automated tools to write to `attempt-1/reviews/` and block attempt-local synthesis, I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/asset-classification-split-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-6/reviews/asset-classification-split-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 6 synthesis.
3. **Verdict Rationale.** The Attempt 6 requirements represent an exceptionally high level of structural maturity. Moving the inline file-by-file table out of the requirements document and delegating it to an external machine-validated **Asset Migration Ledger** (AC6) and a robust **Behavior Evidence Contract** (AC7) successfully resolves the requirements schema conflict. However, from the strict perspective of **Asset Classification & Split Review**, a few critical risks and design-level schema gaps remain in the proposed `implementation-plan.md` that must be addressed before the implementation plan can be approved. Therefore, I award an **APPROVE-WITH-RISKS** verdict and mandate three critical required changes to prevent orphaned behaviors and silent file collisions.

---

## Evaluation of the Three Key Questions

### 1. Does the requirement define complete ownership for Core, Gastown, split, retired, and review-marked assets at file-by-file granularity?
**Reviewer Finding: Yes.**
The requirements successfully decouple the file-by-file ledger into `plans/core-gastown-pack-migration/support/asset-migration-ledger.yaml` (AC6). The criteria explicitly mandate that this ledger must have an owner, deterministic source snapshot, closed classification vocabulary, and stable asset/behavior IDs, and that it **must fail** on unrepresented active source files, missing current paths, and unresolved `review` rows. This ensures complete ownership granularity without violating the requirements schema.

### 2. For split assets such as dispatch skills, maintenance docs, architecture fragments, following-mol, command glossary, and TDD discipline content, is each resulting Core and Gastown output explicit?
**Reviewer Finding: Yes in Requirements, but Gapped in Design.**
* **Requirements:** AC6 and the Example Mapping (Row 62) require split assets to record both successors (Core-neutral and Gastown-specific) along with stripped and retained behaviors.
* **Implementation Plan Gap:** The manifest row schema defined in the implementation plan (`implementation-plan.md:156-176`) still contains a singular `new owner` and `new path` field (lines 170-172). This representation is physically incapable of mapping a split asset (which has clean halves) to *both* its Core-neutral output path and its public Gastown output path in a single row. If splits are modeled as two separate behavioral rows, the design fails to specify how the validator guarantees that both halves of the split exist and that no behavior is orphaned.

### 3. Are review-marked assets resolved before downstream implementation depends on them?
**Reviewer Finding: Yes.**
AC6 explicitly fails on unresolved `review` rows, and the Out of Scope section (`requirements.md:113-114`) strictly bars implementation approval until the ledger exists and passes. This ensures no downstream work can key off an unresolved classification.

---

## Critical Risks & Architectural Gaps

### 1. [Major] The CI Generator Freshness Trap (Self-Defeating Completeness Checks)
* **The Risk:** AC6 (ledger) and AC13 (test coverage transfer) require validation commands to assert that all legacy files and assertions have been re-homed, failing on "stale source snapshots" and "unrepresented active source files."
* **The Gap:** Once the legacy folders under `internal/bootstrap/packs/core` and `examples/gastown/packs/maintenance` are physically deleted during Slice 5b/7, any live file-system walks (`git ls-files` or `os.ReadDir`) on subsequent commits or PRs will find zero files. The generator and validators will find an empty legacy set, causing the completeness check to silently pass with 0 files mapped, completely defeating the safety gate.
* **The Fix:** The requirements and design must specify that the completeness validator does not run against a live workspace walk alone, but validates against a **frozen historical reference snapshot** (such as a specified baseline Git commit or a checked-in cryptographically hashed local snapshot of the legacy roots). This ensures physical deletions do not defeat the completeness checks.

### 2. [Major] Multi-Output Split Schema Defect (AC7 Manifest Schema)
* **The Risk:** The manifest schema defined on lines 156-176 of `implementation-plan.md` assumes that each row maps to a single `new owner` and `new path`. For split assets (such as `architecture.template.md` which has completely divergent implementations under `gascity` and `gascity-packs` roots), this schema cannot represent the dual destinations. One half of the split will be left completely unmapped or orphaned, violating the behavior preservation contract.
* **The Fix:** Amend the schema in `implementation-plan.md` to require dual destination paths (e.g., `target_core_path` relative to the Gas City root, and `target_gastown_path` relative to the Gastown pack repository root) plus the specific content boundaries assigned to each target for any row classified as `split`.

### 3. [Major] Basename Collision Risk for Static/Template Assets
* **The Risk:** While the design implements "Zero-duplicate-active and zero-merge gates" (`implementation-plan.md:275`) at runtime based on behavior IDs, static files (such as operator docs, markdown template fragments, and diagrams) do not carry registered "behavior IDs." A basename collision (such as `architecture.template.md` existing under both `gastown` and `maintenance` roots) will result in silent file overwrites during materialization or cache promotion.
* **The Fix:** Require the build-time or ledger-time linter to incorporate a "basename collision scanner" that flags same-named files across the three legacy pack roots, forcing the design document to declare a definitive merge/override policy for each pair.

---

## Required Changes for Finalization

1. **Frozen CI Baseline:** Update AC6 and AC13 to specify that completeness and freshness validation commands must validate against a frozen historical reference snapshot or baseline Git commit so deleting legacy directories doesn't break CI verification.
2. **Mandate Dual-Output Fields for Splits:** Amend `implementation-plan.md:170-172` to require both `target_core_path` and `target_gastown_path` relative targets for all `split` and `core-renamed` rows, preventing orphaned behavior or path collisions across the two target repositories.
3. **Implement Basename Collision Scanning:** Add a mandatory check under AC6 for a "basename collision scanner" that flags same-named template fragments/files across legacy paths to enforce explicit, documented reconciliation.

---

## Questions

* **Reconciliation of `architecture.template.md`:** Since the two legacy copies of this template are completely divergent, will the ledger merge them into a single file, keep separate outputs, or retire the maintenance copy in favor of a clean Core-neutral metadata definition?
* **Validation of Static Assets:** How will the generator ensure that static files and documentation pages (which do not have executable behavior or triggers) are verifiably validated for completeness without manual reviewer oversight?
