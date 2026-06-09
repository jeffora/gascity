# Hugo Bautista — DeepSeek V4 Flash Perspective Independent Review (Iteration 7)

**Verdict:** approve-with-risks

**Scope:** Split and core-renamed ownership; file-by-file migration ledger completeness; clean halves; and resolution of review-marked assets.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding & Independence.** I have re-grounded this independent review against the active Attempt 7 requirements document (`plans/core-gastown-pack-migration/requirements.md` / `.gc/design-reviews/ga-dtvdnd/attempt-7/design-before.md`, updated 2026-06-09), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` retired assets, and the proposed implementation plan `plans/core-gastown-pack-migration/implementation-plan.md` (updated 2026-06-09). I evaluated the criteria and verified the asset-split and de-roling hazards against the live repository tree.
2. **Dual-Placement Strategy.** Due to the known workflow defect where the bead's metadata `gc.attempt=1` causes automated tools to write to `attempt-1/reviews/` and block attempt-local synthesis, I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/asset-classification-split-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-7/reviews/asset-classification-split-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 7 synthesis.
3. **Verdict Rationale.** The Attempt 7 requirements represent an exceptionally high level of structural maturity. Moving the inline file-by-file table out of the requirements document and delegating it to an external machine-validated **Asset Migration Ledger** (AC6) and a robust **Behavior Evidence Contract** (AC7) successfully resolves the requirements schema conflict. However, from the strict perspective of **Asset Classification & Split Review**, a few critical risks and design-level schema gaps remain in the proposed `implementation-plan.md` that must be addressed before the implementation plan can be approved. Therefore, I award an **APPROVE-WITH-RISKS** verdict and mandate four critical required changes to prevent orphaned behaviors, silent file collisions, and multi-repo transition drift.

---

## Evaluation of the Three Key Questions

### 1. Does the requirement define complete ownership for Core, Gastown, split, retired, and review-marked assets at file-by-file granularity?
**Reviewer Finding: Yes.**
The requirements successfully decouple the file-by-file ledger into `plans/core-gastown-pack-migration/support/asset-migration-ledger.yaml` (AC6). The criteria explicitly mandate that this ledger must have an owner, deterministic source snapshot, closed classification vocabulary, and stable asset/behavior IDs, and that it **must fail** on unrepresented active source files, missing current paths, and unresolved `review` rows. This ensures complete ownership granularity without violating the requirements schema.

### 2. For split assets such as dispatch skills, maintenance docs, architecture fragments, following-mol, command glossary, and TDD discipline content, is each resulting Core and Gastown output explicit?
**Reviewer Finding: Yes in Requirements, but Gapped in Design.**
* **Requirements:** AC6 and the Example Mapping (Row 77) require split assets to record both successors (Core-neutral and Gastown-specific) along with stripped and retained behaviors.
* **Implementation Plan Gap:** The manifest row schema defined in the implementation plan (`implementation-plan.md:170-192`) still contains a singular `new owner` and `new path` field (lines 184-186). This representation is physically incapable of mapping a split asset (which has clean halves) to *both* its Core-neutral output path and its public Gastown output path in a single row. If splits are modeled as two separate behavioral rows, the design fails to specify how the validator guarantees that both halves of the split exist and that no behavior is orphaned.

### 3. Are review-marked assets resolved before downstream implementation depends on them?
**Reviewer Finding: Yes.**
AC6 explicitly fails on unresolved `review` rows, and the Out of Scope section (`requirements.md:129-130`) strictly bars implementation approval until the ledger exists and passes. This ensures no downstream work can key off an unresolved classification.

---

## Critical Risks & Architectural Gaps

### 1. [Major] Multi-Output Split Schema Defect (AC7 Manifest Schema)
* **The Risk:** The manifest schema defined on lines 170-192 of `implementation-plan.md` assumes that each row maps to a single `new owner` and `new path`. For split assets (such as `architecture.template.md` which has completely divergent implementations under `gascity` and `gascity-packs` roots), this schema cannot represent the dual destinations. One half of the split will be left completely unmapped or orphaned, violating the behavior preservation contract.
* **The Fix:** Amend the schema in `implementation-plan.md` to require dual destination paths (e.g., `target_core_path` relative to the Gas City root, and `target_gastown_path` relative to the Gastown pack repository root) plus the specific content boundaries assigned to each target for any row classified as `split`.

### 2. [Major] No Core/non-Gastown behavior baseline — misclassification *toward* Gastown is undetectable
* **The Risk:** AC7 (line 99) records "the authoritative supported-**Gastown** before-state denominator" and traces AC6 rows to a Gastown successor, a Core/controller-owned outcome, or intentional retirement. There is no symmetric Core/non-Gastown denominator: the non-Gastown happy path (line 75) only checks that Core *loads* and the controller operates, and AC8 (line 100) only checks Core carries no role *names* (a Core that is *missing* generic behavior has no role names). AC6's orphan/duplicate checks operate on the ledger's *own* accounting, so if a behavior-entangled asset is recorded single-owner "Gastown", the ledger believes it is whole, AC7 only confirms Gastown still works (moving *more* into Gastown never breaks the Gastown baseline), and nothing proves the Core or Core-generalized successor reproduces the pre-split generic behavior. Generic behavior swept into Gastown is therefore caught by none of AC6/AC7/AC8, and every non-Gastown city silently loses it.
* **The Fix:** Update AC7 to be symmetric (two-sided) or require a Core behavior baseline that guarantees non-Gastown cities retain all pre-split generic behaviors.

### 3. [Major] Basename Collision Risk for Static/Template Assets
* **The Risk:** While the design implements "Zero-duplicate-active and zero-merge gates" (`implementation-plan.md:347`) at runtime based on behavior IDs, static files (such as operator docs, markdown template fragments, and diagrams) do not carry registered "behavior IDs." A basename collision (such as `architecture.template.md` existing under both `gastown` and `maintenance` roots) will result in silent file overwrites during materialization or cache promotion.
* **The Fix:** Require the build-time or ledger-time linter to incorporate a "basename collision scanner" that flags same-named files across the three legacy pack roots, forcing the design document to declare a definitive merge/override policy for each pair.

### 4. [Major] Dangling Split State during Multi-Repo Rollout
* **The Risk:** Slices 1a/1b/1c define a multi-step rollout where ownership and behavior evidence are mapped across repositories. If the Gas City PR is merged but the public Gastown commit has not been fully verified or finalized, there is a risk of a "dangling split" where half of an asset's behavior has been stripped from Gas City but its corresponding Gastown counterpart is not yet live or correctly version-pinned.
* **The Fix:** Implement an explicit "clean-halves atomic publish and verify gate" under AC6/AC14 that blocks release unless both halves of any split or core-renamed asset are verifiably published and resolvable under their respective versions.

---

## Required Changes for Finalization

1. **Mandate Dual-Output Fields for Splits:** Amend `implementation-plan.md:184-186` to require both `target_core_path` and `target_gastown_path` relative targets for all `split` and `core-renamed` rows, preventing orphaned behavior or path collisions across the two target repositories.
2. **Establish Symmetric Core Behavior Baseline:** Update AC7 to include a symmetric, verified non-Gastown Core behavior baseline so that generic behavior accidentally classified as Gastown-only is detected and rejected.
3. **Implement Basename Collision Scanning:** Add a mandatory check under AC6 for a "basename collision scanner" that flags same-named template fragments/files across legacy paths to enforce explicit, documented reconciliation.
4. **Clean-Halves Multi-Repo Gate:** Enforce an atomic publish and verify gate under AC14 that prevents any "dangling split" states where half of an asset is retired/stripped in Gas City before its Gastown counterpart is verifiably published and active.

---

## Questions

* **Reconciliation of `architecture.template.md`:** Since the two legacy copies of this template are completely divergent, will the ledger merge them into a single file, keep separate outputs, or retire the maintenance copy in favor of a clean Core-neutral metadata definition?
* **Validation of Static Assets:** How will the generator ensure that static files and documentation pages (which do not have executable behavior or triggers) are verifiably validated for completeness without manual reviewer oversight?
