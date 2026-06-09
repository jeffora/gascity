# Nadia Volkov — DeepSeek V4 Flash Perspective Independent Review (Iteration 20 / Attempt 20)

**Verdict:** approve

**Scope:** Behavior preservation lane only — Gastown behavior inventory, cross-repo evidence chains, requester/detector/notification continuity, and preventing silent capability loss.

This review evaluates the Iteration 20 / Attempt 20 draft of `design.md` against `requirements.md` and the existing codebase behavior.

---

## Executive Summary

The Iteration 20 design document represents a masterclass in comprehensive, risk-aware, and highly disciplined software architecture. By systematically incorporating feedback from prior iterations, the current design completely resolves all previously highlighted major blockers:

1. **Resolution of the Provider-Pack Contradiction:** Section §303 explicitly introduces a table-driven byte-continuity exception for `bd` and `dolt`. This elegantly resolves the conflict between Core role-neutrality and the provider-pack byte-continuity rule by permitting targeted rewrites of active formulas, orders, scripts, and recipients (such as `mol-dog-*`) from hardcoded roles to dynamic `core.maintenance_worker` bindings.
2. **Prevention of Silent Alert Failures:** The generalized scripts now strictly enforce preflight validations (§1378–1380). Missing optional recipients are treated as safe no-ops, whereas empty or `/` targets in required recipient fields fail preflight immediately.
3. **Hermetic Offline Testing:** The `test/packcompat` fixture matrix (§870–880) now provides explicit "offline no-cache" and "ordinary remote cache" test targets, ensuring robust execution in air-gapped CI environments without relying on silent embedded fallback behavior.

With these major gaps closed, the design is fully mature and approved from the perspective of **Behavior Preservation Auditing**.

---

## Top Strengths

- **Source-Derived Executable Manifest (§88–120, §318–330):** Mandating that the behavior manifest is generated from actual source discovery rather than manual curation ensures 100% coverage. Requiring this to be checked in as a machine-readable artifact (`plans/core-gastown-pack-migration/behavior-manifest.generated.yaml`) allows continuous automated CI enforcement.
- **Strict Behavior Witness Floor (§313–318):** Every row in the behavior manifest is required to have both old and final witness tests, rejecting simplistic file-set, path, or include-count checks.
- **Coordinated 7-Slice Rollout Plan (§2351–2427):** The rollout sequence is structured into logical, intermediate, test-green slices that decouple high-risk steps and enforce clear pre-merge gates.
- **Fail-Closed Loading Invariants (§137–140):** Fail-closed behavior on missing, corrupt, stale, or partially materialized system packs ensures that core SDK execution is always protected.

---

## Nuanced Risks & Recommendations

While the design is approved, the following operational recommendations are offered to ensure flawless execution:

### 1. Verification of Default Recipients in Standard Templates
- **The Risk:** Since required recipient fields fail preflight if empty or `/` (§1379), there is a minor risk that a fresh city initialized with `gc init --template gastown` could fail preflight immediately if default templates do not ship with valid default recipient structures.
- **Recommendation:** Ensure that the default template configurations populated during `gc init` contain fully valid, out-of-the-box non-empty default recipient entries, or fall back to safe default-disabled states where the field is marked as optional.

### 2. Upstream Release Coordination Overhead
- **The Risk:** Pinned Gastown versions (`PublicGastownPackVersion`) are tied to immutable public commits. Any bugfix or minor change in Gastown requires a dual-repo dance (commit to Gastown -> update pin in Gas City -> release Gas City).
- **Recommendation:** Document a clear, streamlined "hotfix runbook" in the repository's contributor docs to minimize operator and developer friction when performing synchronized updates of the immutable pin.

---

## Evaluation of the Three Key Questions

### 1. Does the behavior inventory enumerate every Gastown-specific requester, detector, notification path, formula, order, script branch, and prompt fragment removed from Core?
- **Auditor Finding:** **Yes.** The expansion of the de-roling scan scope to cover required provider packs (`bd` and `dolt`), overlays, prompt fragments, and templates (§1823) is highly thorough. The narrow, table-driven byte-continuity exceptions (§303) ensure that active provider pack assets are fully cataloged, leaving no untracked role leakage.

### 2. Which concrete `gascity-packs/gastown` tests prove each restored behavior fires under the same trigger conditions rather than merely existing?
- **Auditor Finding:** **Satisfactory.** The `test/packcompat` fixture matrix (§870–880) and the "Strict Behavior Witness Floor" (§313) guarantee that path/existence checks cannot substitute for actual behavior execution. The requirement that every manifest row map to test-function and subtest granularity guarantees that execution-level coverage is strictly maintained.

### 3. Can reviewers trace each high-risk Maintenance or Core move to old path, new path, landing commit, and observable test evidence?
- **Auditor Finding:** **Yes.** The machine-readable manifest (`behavior-manifest.generated.yaml`) combined with the 7-slice rollout gates provides a clear, step-by-step audit trail. Every moved asset is mapped from old source digest to final asset digest, with a stable row ID that survives path renames.

---

## Recommendations for Finalization

1. **Mandate Preflight Validation Tests:** Explicitly require that the test harness includes test assertions covering both empty-recipient and slash-recipient preflight failure cases.
2. **Synchronized Hotfix Process:** Draft a brief guide in the repository's contributor docs on how to update and test `PublicGastownPackVersion` during local development when working across both `gascity` and `gascity-packs` repositories.
