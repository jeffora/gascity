# Hugo Bautista — DeepSeek V4 Flash Perspective Independent Review (Asset Classification & Split) — Iteration 9 / Attempt 9

**Verdict:** approve-with-risks

**Scope:** Split and core-renamed ownership; file-by-file migration ledger completeness; clean halves; and resolution of review-marked assets.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this independent review against the active Iteration 9 requirements document (`plans/core-gastown-pack-migration/requirements.md` / `.gc/design-reviews/ga-dtvdnd/attempt-9/design-before.md`, updated 2026-06-09), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` retired assets, and the proposed implementation plan `plans/core-gastown-pack-migration/implementation-plan.md` (updated 2026-06-09). I evaluated the criteria and verified the asset-split and de-roling hazards against the live repository tree.
2. **Dual-Placement Strategy.** Due to the known workflow defect where the bead's metadata `gc.attempt=1` causes automated tools to write to `attempt-1/reviews/` and block attempt-local synthesis, I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/asset-classification-split-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-9/reviews/asset-classification-split-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 9 synthesis.
3. **Verdict Rationale.** The Iteration 9 requirements represent an exceptionally high level of structural maturity. Moving the inline file-by-file table out of the requirements document and delegating it to an external machine-validated **Asset Migration Ledger** (AC6) and a robust **Behavior Evidence Contract** (AC7) successfully resolves the requirements schema conflict. Therefore, I award an **APPROVE-WITH-RISKS** verdict because while the requirements themselves are highly sound, several critical classification, override, and split-schema gaps remain in the proposed `implementation-plan.md` that must be addressed during design finalization.

---

## Lane-Specific Detailed Responses

### Q1: Does the requirement define complete ownership for Core, Gastown, split, retired, and review-marked assets at file-by-file granularity?

**Yes.**
The requirements successfully decouple the file-by-file ledger into `plans/core-gastown-pack-migration/support/asset-migration-ledger.yaml` (AC6) and `plans/core-gastown-pack-migration/support/source-consumer-closure.yaml` (AC5). The criteria explicitly mandate that this ledger must have an owner, deterministic source snapshot, closed classification vocabulary, and stable asset/behavior IDs, and that it **must fail** on unrepresented active source files, missing current paths, and unresolved `review` rows. This ensures complete ownership granularity without violating the requirements schema.

### Q2: For split assets such as dispatch skills, maintenance docs, architecture fragments, following-mol, command glossary, and TDD discipline content, is each resulting Core and Gastown output explicit?

**Yes in Requirements, but Gapped in Design.**
*   **Requirements:** AC6 and the Example Mapping (Row 89) require split assets to record both successors (Core-neutral and Gastown-specific) along with stripped and retained behaviors.
*   **Implementation Plan Gap:** The manifest row schema defined in the implementation plan (`implementation-plan.md:170-192`) still contains a singular `new owner` and `new path` field (lines 184-186). This representation is physically incapable of mapping a split asset (which has clean halves) to *both* its Core-neutral output path and its public Gastown output path in a single row. If splits are modeled as two separate behavioral rows, the design fails to specify how the validator guarantees that both halves of the split exist and that no behavior is orphaned.
*   **Asset Override Gap:** For split mechanics such as `assets/scripts/spawn-storm-detect.sh`, the generic detection moves to Core and calls a generic `escalate.sh` script/hook. To make public Gastown override escalation by filename, the implementation needs a resolved pack-asset lookup or pack-script PATH. The implementation plan does not explicitly design or wire this asset override PATH or lookup mechanic to prevent raw Core-relative `$PACK_DIR` paths from breaking override behaviors.

### Q3: Are review-marked assets resolved before downstream implementation depends on them?

**Yes.**
AC6 explicitly fails on unresolved `review` rows, and the Out of Scope section (`requirements.md:143-144`) strictly bars implementation approval until the ledger exists and passes. This ensures no downstream work can key off an unresolved classification.

---

## Critical Risks & Architectural Gaps

### 1. [Major] Multi-Output Split Schema Defect (AC7 Manifest Schema)
*   **The Risk:** The manifest schema defined on lines 170-192 of `implementation-plan.md` assumes that each row maps to a single `new owner` and `new path`. For split assets (such as `architecture.template.md` which has completely divergent implementations under `gascity` and `gascity-packs` roots), this schema cannot represent the dual destinations. One half of the split will be left completely unmapped or orphaned, violating the behavior preservation contract.
*   **The Fix:** Amend the schema in `implementation-plan.md` to require dual destination paths (e.g., `target_core_path` relative to the Gas City root, and `target_gastown_path` relative to the Gastown pack repository root) plus the specific content boundaries assigned to each target for any row classified as `split`.

### 2. [Major] Undefined Product Placement for Generic Maintenance Assets
*   **The Risk:** While AC9 covers the maintenance *executor/work*, the Maintenance pack's *assets* (prompts, formulas, docs) lack a clear product placement policy. If generic maintenance assets (like a template for system resource cleanup or generic health checking) are retired or swept entirely into the Gastown pack without generic Core equivalents, non-Gastown Gas City deployments will lose these utility assets.
*   **The Fix:** Explicitly define a product placement policy for all generic maintenance assets. If an asset represents generic, role-neutral maintenance helper utilities, it must have a Core-neutral placement/successor in `internal/packs/core` (or be marked for retirement with documented alternatives).

### 3. [Major] Basename Collision Risk for Static/Template Assets
*   **The Risk:** While the design implements "Zero-duplicate-active and zero-merge gates" (`implementation-plan.md:347`) at runtime based on behavior IDs, static files (such as operator docs, markdown template fragments, and diagrams) do not carry registered "behavior IDs." A basename collision (such as `architecture.template.md` existing under both `gastown` and `maintenance` roots) will result in silent file overwrites during materialization or cache promotion.
*   **The Fix:** Require the build-time or ledger-time linter to incorporate a "basename collision scanner" that flags same-named files across the three legacy pack roots, forcing the design document to declare a definitive merge/override policy for each pair.

### 4. [Major] Dangling Split State during Multi-Repo Rollout
*   **The Risk:** Slices 1a/1b/1c define a multi-step rollout where ownership and behavior evidence are mapped across repositories. If the Gas City PR is merged but the public Gastown commit has not been fully verified or finalized, there is a risk of a "dangling split" where half of an asset's behavior has been stripped from Gas City but its corresponding Gastown counterpart is not yet live or correctly version-pinned.
*   **The Fix:** Implement an explicit "clean-halves atomic publish and verify gate" under AC6/AC14 that blocks release unless both halves of any split or core-renamed asset are verifiably published and resolvable under their respective versions.

### 5. [Major] Dynamic Script Override Path Resolution
*   **The Risk:** Core-owned scripts that invoke helpers (such as `spawn-storm-detect.sh` invoking `escalate.sh`) are designed to be overridable by Gastown. However, if Core-relative `$PACK_DIR` paths are hardcoded in the scripts, the override mechanism will be bypassed, resulting in Gastown being unable to override Core behaviors dynamically.
*   **The Fix:** Specify an explicit search PATH or look-up command in the implementation plan to dynamically locate asset scripts based on active pack-hierarchy precedence, ensuring that public Gastown overrides are resolved correctly.

---

## Required Changes for Finalization (Pins)

1.  **Mandate Dual-Output Fields for Splits:** Amend `implementation-plan.md:184-186` to require both `target_core_path` and `target_gastown_path` relative targets for all `split` and `core-renamed` rows, preventing orphaned behavior or path collisions across the two target repositories.
2.  **Define Placement Policy for Maintenance Assets:** Explicitly establish in the design document a product placement policy that ensures generic, role-neutral maintenance assets receive symmetric Core-neutral successors.
3.  **Implement Basename Collision Scanning:** Add a mandatory check under AC6 for a "basename collision scanner" that flags same-named template fragments/files across legacy paths to enforce explicit, documented reconciliation.
4.  **Clean-Halves Multi-Repo Gate:** Enforce an atomic publish and verify gate under AC14 that prevents any "dangling split" states where half of an asset is retired/stripped in Gas City before its Gastown counterpart is verifiably published and active.
5.  **Design Asset-Script PATH Resolution:** Define a dynamic lookup PATH or command in `implementation-plan.md:107-115` that enforces pack priority precedence for shell helper invocations across Core and Gastown.

---

## Remaining Questions

*   **Reconciliation of `architecture.template.md`:** Since the two legacy copies of this template are completely divergent, will the ledger merge them into a single file, keep separate outputs, or retire the maintenance copy in favor of a clean Core-neutral metadata definition?
*   **Validation of Static Assets:** How will the generator ensure that static files and documentation pages (which do not have executable behavior or triggers) are verifiably validated for completeness without manual reviewer oversight?
