# Hugo Bautista — DeepSeek V4 Flash (Asset Classification & Split Review) — Iteration 5 / Attempt 5

**Verdict:** block

**Scope:** Split and core-renamed ownership; file-by-file migration table completeness; clean halves; resolution of review-marked assets.

---

> ### Lane Note (Verify-Don't-Copy + Path Alignment)
> 1. **Re-Grounding & Independence:** This review is an independent DeepSeek V4 Flash evaluation of the Attempt 5 requirements (`plans/core-gastown-pack-migration/requirements.md` updated 2026-06-09) and the proposed implementation plan (`plans/core-gastown-pack-migration/implementation-plan.md` updated 2026-06-09T07:28:00Z). I have not inherited conclusions from prior rounds. I evaluated AC6, AC7, and the Example Mapping, and verified the split/duplicate hazards against the live source tree to expose critical risks.
> 2. **Dual-Placement Strategy:** Due to the known `gc.attempt=1` path override on active beads (which causes them to write output metadata to `attempt-1/` and blocks attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/asset-classification-split-reviewer_gemini.md` and `.gc/design-reviews/ga-dtvdnd/attempt-5/reviews/asset-classification-split-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 5 synthesis.

---

## Executive Summary

The Attempt 5 requirements and the updated implementation plan represent a highly disciplined progression. Moving the inline file-by-file table out of the requirements document and delegating it to an external machine-validated **Asset Migration Ledger** (AC6) and a robust **Behavior Evidence Contract / Manifest** (AC7) successfully resolves the requirements schema conflict. 

However, from the perspective of an **Asset Classification & Split Reviewer**, the proposed design must be **BLOCKED**. Under the current `implementation-plan.md` specification, the manifest schema is structurally incapable of representing split assets, the validation command and execution model for the AC6 ledger remain undefined, and the lack of a source-level basename collision scanner for static/template assets leaves a critical silent overwrite risk unaddressed.

---

## Lane-Specific Detailed Responses

### Q1: Does the requirement define complete ownership for Core, Gastown, split, retired, and review-marked assets at file-by-file granularity?

**Yes in Requirements, but Gapped in Design.**
*   **Requirements:** AC6 (`requirements.md:81`) mandates a validated asset migration ledger (`ownership.yaml` or similar) containing a named path, generator/validator command, owner, current path, target owner, all target output paths, split boundary, fallback classification, and proof command. It must fail on unrepresented active source files, missing current paths, and unresolved `review` rows.
*   **Implementation Plan Gap:** The design defines `gascity-packs/gastown/ownership.yaml` (`implementation-plan.md:132`) as the artifact assigning ownership. However, while `implementation-plan.md` details the generator, schema, and tests for the AC7 *Behavior Evidence Contract* (lines 145-192), it leaves the AC6 *Asset Migration Ledger* validator as a conceptual black box. The plan fails to specify:
    1. What command or tool validates `ownership.yaml`.
    2. Whether this validator is part of the `gc` CLI or a standalone linter.
    3. How bidirectional validation is enforced (i.e., proving that every single file under the legacy roots maps to exactly one row, and failing on phantom/stale rows). 
    Without a concrete design for the AC6 validation command, we risk proceeding to implementation with a static YAML file that drifts from the live source snapshot.

### Q2: For split assets such as dispatch skills, maintenance docs, architecture fragments, following-mol, command glossary, and TDD discipline content, is each resulting Core and Gastown output explicit?

**No — Structural Schema Defect in Design.**
*   **Requirements:** AC6 and the Example Mapping (`requirements.md:61`) require split assets to record both successors (Core-neutral and Gastown-specific) along with stripped and retained behaviors.
*   **Implementation Plan Defect:** The manifest row schema defined in the implementation plan (`implementation-plan.md:156-176`) has a singular `new owner` and `new path` field:
    *   Line 170: `new owner: Core, public Gastown, provider pack, docs-only, or approved retirement;`
    *   Line 172: `new path and new witness;`
    This singular representation is physically incapable of mapping a split asset to *both* its Core-neutral output path and its public Gastown output path in a single row. If splits are modeled as two separate behavioral rows, the design fails to explain how the validator guarantees that both halves of the split exist and that no behavior is orphaned during the split. 

### Q3: Are review-marked assets resolved before downstream implementation depends on them?

**Yes in Requirements, partially addressed in Design.**
*   **Requirements:** AC6 explicitly fails validation on unresolved `review` rows, and Out-of-Scope (`requirements.md:112`) bars implementation approval until the ledger exists and passes.
*   **Implementation Plan:** The plan correctly requires `gascity-packs/gastown/ownership.yaml` (`implementation-plan.md:132`) to resolve all review-marked assets (such as `mol-review-quorum`, Dog prompts, and tmux behavior) before source moves can proceed. However, it does not define a clear timeline or gate to prevent developers from checking in intermediate code slices while `review` rows are still present in the ledger.

---

## Critical Risks & Gaps

### 1. [Blocker] Multi-Output Split Schema Defect (AC7 Manifest Schema)
*   **The Risk:** The manifest schema defined on lines 156-176 of `implementation-plan.md` assumes that each row maps to a single `new owner` and `new path`. For split assets (such as `architecture.template.md` which has completely divergent implementations under `gastown` and `maintenance` roots), this schema cannot represent the dual destinations. One half of the split will be left completely unmapped or orphaned, violating the behavior preservation contract.
*   **Mitigation:** Amend the schema in `implementation-plan.md` to require dual destination paths (e.g., `target_core_path` relative to the Gas City root, and `target_gastown_path` relative to the Gastown pack repository root) plus the specific content boundaries assigned to each target for any row classified as `split`.

### 2. [Major] Undefined AC6 Ledger Validator Command and Execution Model
*   **The Risk:** While `implementation-plan.md` details the AC7 `test/packcompat` and `internal/packevidence` generator, it treats the AC6 "Asset Migration Ledger" validator as an exercise for the implementer. Without a concrete design for the AC6 ledger validator, we cannot ensure that the ownership YAML remains synchronized with the live git repository or that phantom/stale rows are caught.
*   **Mitigation:** Define the AC6 ledger validator tool, its exact CLI command (e.g., `gc packlint --ledger`), and specify that it must perform strict bidirectional validation against a deterministic `git ls-files` snapshot of the legacy roots.

### 3. [Major] Basename Collision Risk for Static/Template Assets
*   **The Risk:** While the design implements "Zero-duplicate-active and zero-merge gates" (`implementation-plan.md:275`) at runtime based on behavior IDs, static files (such as operator docs, markdown template fragments, and diagrams) do not carry registered "behavior IDs". A basename collision (such as `architecture.template.md` existing under both `gastown` and `maintenance` roots) will result in silent file overwrites during materialization or cache promotion.
*   **Mitigation:** Require the build-time or ledger-time linter to incorporate a "basename collision scanner" that flags same-named files across the three legacy pack roots, forcing the design document to declare a definitive merge/override policy for each pair.

---

## Required Changes & Actions

1.  **Mandate Dual-Output Fields for Splits:**
    Amend `implementation-plan.md:170-172` to require both `target_core_path` and `target_gastown_path` relative targets for all `split` and `core-renamed` rows, preventing orphaned behavior or path collisions across the two target repositories.
2.  **Define the AC6 Ledger Validator Command:**
    Detail the command (e.g., `gc packlint --ledger`) and execution model for the AC6 Asset Migration Ledger, ensuring it runs as a pre-commit and CI gate that performs bidirectional validation against a deterministic git-backed snapshot.
3.  **Implement Basename Collision Scanning:**
    Add a mandatory check under AC6 for a "basename collision scanner" that flags same-named template fragments/files across legacy paths to enforce explicit, documented reconciliation.
4.  **Add Example Mapping Split Row:**
    Add a concrete Example Mapping row in `requirements.md` or a design walkthrough in `implementation-plan.md` demonstrating a real split asset (like `architecture.template.md`), illustrating its Core-neutral output, Gastown-specific output, validation evidence, and behavior-preservation proof.

---

## Questions

*   **Reconciliation of `architecture.template.md`:** Since the two legacy copies of this template are completely divergent, will the ledger merge them into a single file, keep separate outputs, or retire the maintenance copy in favor of a clean Core-neutral metadata definition?
*   **Validation of Static Assets:** How will the generator ensure that static files and documentation pages (which do not have executable behavior or triggers) are verifiably validated for completeness without manual reviewer oversight?
*   **Verification of Dual-Repo Sync:** How will CI guarantee that a commit in `gascity` and a commit in `gascity-packs` are perfectly synchronized during the intermediate compatibility-pin phase?
