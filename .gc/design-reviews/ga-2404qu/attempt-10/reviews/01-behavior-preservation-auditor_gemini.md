# Nadia Volkov — Gemini Independent Review (Iteration 10 / Attempt 10)

**Verdict:** approve-with-risks

**Scope:** Behavior preservation lane only — Gastown behavior inventory, cross-repo evidence chains, requester/detector/notification continuity, and preventing silent capability loss.

This review evaluates the Iteration 10 draft of `design.md` (`updated_at: 2026-06-07T10:41:04Z`) against `requirements.md` and the existing codebase behavior.

---

## Executive Summary

The Iteration 10 design draft represents a mature, highly secure architectural specification for splitting Core and Gastown. By introducing the **Public Gastown Pin And Overlap Boundary (§1178–1214)** and the **Retired-Source Classifier As The Sole Gate (§1215–1230)**, the design directly resolves the rollout-collision circularity identified in earlier iterations. The transition from an ad-hoc path comparison to an explicit, centralized classification model is an outstanding improvement in safety and robustness.

However, from a strict behavioral preservation auditing perspective, several **operational transition risks, commit authorship gaps, and conditional execution edge cases** remain. Specifically, the introduction of a dual-pin rollout (compatibility pin vs. activation pin) shifts the complexity burden to cross-repo release orchestration and introduces subtle risk surfaces that must be closely monitored.

---

## Top Strengths

- **Resolved Coexistence Collision (§1178–1214):** The split of the public Gastown pin into a "compatibility pin" (inactive, non-overlapping definitions) and an "activation pin" is an elegant and robust mechanism to prevent the duplicate-active-definition fatal check from crashing Slices 2–4.
- **Centralized Classifier API (§1215–1230):** Replacing individual file-path globs with a single retired-source classifier API ensures that all validation paths (including doctor, config load, and packman) enforce identical containment boundaries.
- **Robust Behavior Witness Floor (§1324):** Incorporating a generated manifest matching every removed/generalized asset with concrete execution witnesses ensures that no behavior is lost during the Core/Gastown split.
- **Strict Host-Core Priority Layering (§328–347):** Requiring Gastown to patch Core's `dog` configuration via normal resolved-config layering instead of importing Core or silently recreating it prevents circular dependencies.

---

## Critical Risks & Gaps

### 1. Cross-Repo Synchronization and Version-Control Complexity (§1181–1205)
- **The Risk:** The newly added "compatibility vs. activation pin" boundary solves the duplicate active definition error, but creates significant release-management complexity. In the `gascity-packs/gastown` repository:
  - Maintainers must support a compatibility state where behavior rows are present but kept inactive.
  - Then, in Slice 5, they must shift to an activation state.
- **The Gap:** The design does not specify how these compatibility assets are kept "inactive" under the old loader while being ready to be activated. Is this achieved via feature-flag conditions in formula metadata, or by branching in the `gascity-packs/gastown` repository? If it requires managing multiple branches (e.g., `compat-v1` and `main`), any drift between these branches could cause a failure when transitioning from Slice 2 to Slice 5.
- **Required Recommendation:** The `gascity-packs/gastown` release orchestration must be documented. If branching is used, the design should mandate that CI in the `gascity-packs` repo automatically validates that both the compatibility branch and the activation branch satisfy the identical behavioral-trigger evidence matrix before the compatibility pin is checked in.

### 2. Unresolved Git Commit Authorship Configuration (§40–42, §491–507)
- **The Risk:** Moving scripts like `jsonl-export.sh` and `reaper.sh` to Core-owned generic maintenance strips Gastown specific commit-author metadata (e.g., `--author 'reaper <reaper@gastown.local>'`). While the design acknowledges this, it does not explicitly declare how Core-neutral scripts obtain their commit-author identities.
- **The Gap:** If the authorship variables default to generic values, Gastown's historical commit-metadata footprint is lost. If they are left unset or empty, Git/Dolt commits may fallback to environment-derived or system identities (e.g., `ubuntu@localhost`), violating the requirement for reproducible, traceable audit logs in Dolt databases.
- **Required Recommendation:** Core maintenance scripts must resolve commit authorship through explicit config variables (e.g., `core.maintenance.git_author_name` and `core.maintenance.git_author_email`), defaulting to a generic Core maintenance identity. Public Gastown can then override these through normal patches without hardcoding role names in Go or scripts.

### 3. Broken Command Misroute in Generalized Alerts (§1269–1272)
- **The Risk:** In legacy `reaper.sh`, alerts were hardcoded to route to the mayor or daemon. Generalizing these alerts requires evaluating `gc mail send {{RECIPIENT}}`.
- **The Gap:** If a Core-only city runs this script and no recipient is configured, the command evaluates to `gc mail send /`. In legacy code, failures were swallowed via `|| true`. This means the broken command would fail silently, creating a silent capability loss where the operator is never notified of a failing reaper or jsonl-export run.
- **Required Recommendation:** Document that the generalized scripts must perform a preflight check on the recipient variable. If the variable is empty or resolves to `/`, the script must log a descriptive error to stderr and exit with a non-zero code *before* attempting the mail command, ensuring that the calling formula/molecule catches the failure rather than swallowing it with `|| true`.

---

## Evaluation of the Three Key Questions

### 1. Does the behavior inventory enumerate every Gastown-specific requester, detector, notification path, formula, order, script branch, and prompt fragment removed from Core?
- **Auditor Finding:** **Yes.** The combination of the source-derived behavior manifest (§269–317) and the role-surface migration table (§651–685) represents a complete inventory framework. The addition of the "Docs Vocabulary As Executable Contract (§866–888)" ensures that even human-facing comments and generated references are cataloged and validated against pattern drift.

### 2. Which concrete `gascity-packs/gastown` tests prove each restored behavior fires under the same trigger conditions rather than merely existing?
- **Auditor Finding:** **Satisfactory with recommendations.** The introduction of the dual-mode `test/packcompat` suite (§805–818) ensures compatibility tests run under both legacy and no-Maintenance production loaders. However, to guarantee that conditional behaviors (such as wisp compaction escalations or spawn-storm detections) actually execute their warning branches, the matrix must include explicit **behavioral-trigger fixtures** that simulate these exact error states rather than merely checking file compilation and load paths.

### 3. Can reviewers trace each high-risk Maintenance or Core move to old path, new path, landing commit, and observable test evidence?
- **Auditor Finding:** **Yes.** The 7-slice rollout plan (§2052–2116) is clean and logical. Each slice is decoupled and lists precise required gates and package-focused tests (e.g., `go test ./test/packlint/...` in Slice 3, `packcompat` TestPinnedPublicGastownBehavior in Slices 2, 5, and 6). This prevents broad, unverified commits from landing.

---

## Required Changes for Finalization

To finalize the design, the following minor refinements should be incorporated into the next revision of `design.md`:
1. **Define Compatibility-Pin Mechanism:** Specify how the public Gastown pack keeps overlapping definitions inactive during Slices 2–4 (e.g., through conditional TOML loading or dedicated compatibility branching).
2. **Standardize Core Commit-Authorship Config:** Declare the schema and default values for Core maintenance-git authorship to prevent fallback to system identities.
3. **Mandate Preflight Validation for Alert Recipients:** Ensure that generalized scripts explicitly reject empty or invalid recipients with descriptive errors instead of swallowing broken `gc mail send /` executions.
4. **Enforce Behavioral-Trigger Fixtures in Packcompat:** Require that `test/packcompat` executes simulated failure/escalation conditions to verify that conditional notification paths in Gastown fire correctly.
