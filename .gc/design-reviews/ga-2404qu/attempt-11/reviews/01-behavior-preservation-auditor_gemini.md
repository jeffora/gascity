# Nadia Volkov — DeepSeek V4 Flash Perspective Independent Review (Iteration 11 / Attempt 11)

**Verdict:** approve-with-risks

**Scope:** Behavior preservation lane only — Gastown behavior inventory, cross-repo evidence chains, requester/detector/notification continuity, and preventing silent capability loss.

This review evaluates the Iteration 11 draft of `design.md` (`updated_at: 2026-06-07T10:41:04Z`) against `requirements.md` and the existing codebase behavior.

---

## Executive Summary

The Iteration 11 design represents an exceptionally mature, production-ready blueprint for the Core/Gastown split. By incorporating the **Public Gastown Pin and Overlap Boundary (§1179–1215)** and the **Retired-Source Classifier as the Sole Gate (§1216–1235)**, the architecture successfully resolves the rollout-collision circularity that plagued earlier iterations. The transition from ad-hoc path comparisons to a centralized classification model is a major milestone for safety and maintainability.

However, from an exhaustive **Behavior Preservation Auditing** perspective, several critical contradictions, transition edge cases, and silent failure vectors remain unresolved. Specifically:
1. **The Provider Pack Role-Leakage vs. Byte-Continuity Contradiction:** Required provider packs (`dolt`, `bd`) still contain hardcoded references to Gastown roles (Mayor, Deacon), which conflicts with the "zero hardcoded roles" requirement for Core/system packs, yet the design mandates they remain byte-identical.
2. **Fuzzy Transition Mechanics for "Inactive Compatibility Assets":** The loader lacks any native mechanism to ship "disabled" or "inactive" formulas/orders, meaning a duplicate-active-definition crash remains a real threat during the compatibility phase.
3. **Silent Notification Failures on Missing Recipients:** Generalizing alert routing to dynamically evaluate configured recipients without robust preflight checks runs the risk of silent capability loss (swallowing failures under `|| true`) or cascading wisp failures.

---

## Top Strengths

- **centralized Classifier API (§1216–1235):** Replacing disparate file-path and glob checks with a single, unified retired-source classifier ensures that all validation, config-load, and diagnostic paths enforce identical containment boundaries.
- **Robust Behavior Witness Floor (§1324–1346):** The mandate that "the behavior manifest cannot downgrade evidence" ensures that any legacy behavior backed by an execution-level test is guaranteed to have equivalent execution-level verification on the other side of the migration.
- **Double-Pin Rollout Sequence (§1179–1215):** Splitting the public Gastown pin into a "compatibility pin" (coexisting with legacy Maintenance) and an "activation pin" (post-Maintenance retirement) provides a highly resilient, backward-compatible rollout vector.
- **Explicit Role-Surface Inventory (§944–973):** Compiling an asset-by-asset migration table guarantees that no implicit behaviors or prompt fragments are silently dropped or left unowned.

---

## Critical Risks & Gaps

### 1. The Provider Pack Role-Leakage vs. Byte-Continuity Contradiction
- **The Risk:** The design requires that `bd` and `dolt` provider packs remain strictly byte-identical, asserting that "materialized bytes and provenance remain unchanged except for expected manifest metadata" (§2304). However, active provider formulas like `examples/dolt/formulas/mol-dog-backup.toml` contain hardcoded notifications targeting `mayor/` and `deacon/` (§57–60).
- **The Gap:** Because these provider packs are required system packs for all provider-dependent cities, Gastown role targets will leak into Core-only cities through the provider layer. This directly violates the Core role neutrality and zero-hardcoded-roles requirements.
- **Required Recommendation:** Resolve this contradiction by explicitly bringing the provider packs (`bd` and `dolt`) into the role-surface and manifest scope. Update the testing byte-continuity rule to allow modifications for role cleaning, and refactor these formulas to resolve their notification targets dynamically from the city or pool configuration. Alternatively, declare a formal, time-bound exception/allowlist for provider-pack role targets in the role-surface inventory with a clear expiration date.

### 2. Speculative Loader Support for "Inactive Compatibility Assets" (§1182–1206)
- **The Risk:** The design states that duplicate behavior assets in the public Gastown pack are "present only as inactive compatibility assets... until the no-Maintenance loader slice activates them" (§1197–1199).
- **The Gap:** The current Gas City loader has no native support for "inactive" or "disabled" assets; any `.toml` file present in the loaded directory is fully parsed and activated. If these duplicate definitions exist on disk alongside bundled Maintenance, they will trigger a fatal duplicate-active-definition crash.
- **Required Recommendation:** Eliminate the speculative "inactive compatibility assets" terminology. The compatibility pin should simply **omit** any colliding assets entirely, relying on the bundled Maintenance pack to provide them. The subsequent activation pin (adopted atomically in Slice 5 alongside Maintenance removal) will then introduce the complete set of assets.

### 3. Silent Alert Failures and Empty Recipient Misroutes (§1310–1312)
- **The Risk:** Generalized scripts like `reaper.sh` and `jsonl-export.sh` route alerts via `gc mail send {{RECIPIENT}}`. In a Core-only city where no recipient is configured, this evaluates to `gc mail send /` or an empty target, which fails.
- **The Gap:** If the calling formula swallows the error via `|| true` (retaining legacy behavior), it leads to **silent capability loss** where critical anomalies are never surfaced. If the error is not swallowed, the entire wisp/maintenance step fails because of an unconfigured optional notification.
- **Required Recommendation:** Mandate that generalized scripts perform a preflight validation on the recipient variable. If the recipient is empty or resolves to `/`, the script should gracefully log a warning to stderr and skip the mail execution, exiting with `0`. If a recipient *is* configured but the command fails, it must exit with a non-zero code to allow the calling formula to catch and handle the failure.

### 4. Hermetic Offline Execution of `test/packcompat`
- **The Risk:** `test/packcompat` is designed to execute against the exact public Gastown commit on GitHub.
- **The Gap:** In air-gapped, offline, or restricted CI environments, direct network requests to fetch the remote repository will fail, breaking Gas City's test baseline.
- **Required Recommendation:** Require that Gas City's test harness supports a hermetic, local-fixture fallback for `test/packcompat`. The pinned version of the Gastown pack should be vendored or cached locally in the test tree, allowing the compatibility suite to assert full behavioral preservation without requiring internet access.

---

## Evaluation of the Three Key Questions

### 1. Does the behavior inventory enumerate every Gastown-specific requester, detector, notification path, formula, order, script branch, and prompt fragment removed from Core?
- **Auditor Finding:** **Yes, with provider-pack exceptions.** The asset migration map and the newly added "Behavior Evidence First Slice" (§1606–1627) establish a highly detailed and exhaustive inventory framework. However, the inventory is currently incomplete regarding the `examples/dolt` and `examples/bd` notification paths. Once provider pack files are added to the scope, the inventory will be 100% complete.

### 2. Which concrete `gascity-packs/gastown` tests prove each restored behavior fires under the same trigger conditions rather than merely existing?
- **Auditor Finding:** **Satisfactory with recommendations.** The introduction of the `packcompat` suite (§1066) and the "Behavior Evidence Witness Floor" (§1324) ensures that compile-time and path checks cannot replace execution-level tests. However, to guarantee that conditional behaviors (such as wisp escalations or spawn-storm alerts) execute their warning branches, the test suite must mandate the use of **behavioral-trigger fixtures** that simulate these error states.

### 3. Can reviewers trace each high-risk Maintenance or Core move to old path, new path, landing commit, and observable test evidence?
- **Auditor Finding:** **Yes.** The 7-slice rollout plan (§2351–2427) is outstanding. It decouples high-risk steps, lists explicit gates for each slice, and enforces intermediate test runs (such as `go test ./test/packcompat` in Slice 2, 5, and 6), ensuring that no unverified code lands on the main branch.

---

## Required Changes for Finalization

1. **Resolve Provider Pack Contradiction:** Amend §2304 to permit modifications to `bd` and `dolt` bytes for role-cleaning purposes, or add a formal, time-bound exception row to the role-surface inventory.
2. **Clarify Compatibility Pin Contents:** Explicitly state that the compatibility pin omits colliding behavior assets rather than relying on "inactive compatibility assets" loader support.
3. **Enforce Script Preflight Validation:** Add a requirement that generalized scripts validate recipient variables, skipping execution with code `0` on empty/slash targets, and failing loudly on real failures.
4. **Mandate Behavioral-Trigger Fixtures:** Require `test/packcompat` to execute simulated failure/escalation states to verify warning/escalation pathways.
5. **Support Offline Test Mode:** Mandate a hermetic, local-fixture-backed execution path for `test/packcompat` to guarantee air-gapped CI success.
