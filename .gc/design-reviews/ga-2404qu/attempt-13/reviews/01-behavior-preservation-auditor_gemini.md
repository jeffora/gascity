# Nadia Volkov — DeepSeek V4 Flash Perspective Independent Review (Iteration 13 / Attempt 13)

**Verdict:** approve-with-risks

**Scope:** Behavior preservation lane only — Gastown behavior inventory, cross-repo evidence chains, requester/detector/notification continuity, and preventing silent capability loss.

This review evaluates the Iteration 13 / Attempt 13 draft of `design.md` against `requirements.md` and the existing codebase behavior. Because the design document remains identical to Attempt 12, the critical structural gaps, risks, and logical contradictions identified in previous reviews remain active and unaddressed.

---

## Executive Summary

The Iteration 13 design presents a structurally sound blueprint for partitioning the legacy monolithic SDK into a clean Core/Gastown split. Retiring the "inactive compatibility assets" loader abstraction in favor of pure clean-cut compatibility pins simplifies the system loader and prevents duplicate-active-definition runtime crashes. The inclusion of a comprehensive **Full Role-Surface Inventory and Replacement Gate (§1823–1850)** is a major step toward systematic de-roling.

However, from the strict, empirical perspective of **Behavior Preservation Auditing**, several critical risks and logical contradictions remain unresolved in this iteration:
1. **The Required Provider Pack De-Roling vs. Byte-Continuity Contradiction:** Required provider packs (`bd` and `dolt`) are explicitly brought into the de-roling scope (§1831), yet the design mandates that their materialized bytes and git provenance remain completely "unchanged except for expected manifest metadata" (§2562). Active provider formulas like `examples/dolt/formulas/mol-dog-backup.toml` contain hardcoded notifications targeting `deacon/` (§189). Under the current design, we cannot modify these bytes to dynamic variables without violating the byte-continuity constraint.
2. **Silent Alert Failures on Empty Dynamic Recipients:** Generalizing scripts (§1310–1312) to accept alert/warmup recipients from environment or formula variables (such as in `reaper.sh` and `jsonl-export.sh`) introduces a dangerous risk of silent capability loss. In Core-only cities with no recipient configured, these variables evaluate to empty or slash targets (e.g., `gc mail send /`). If errors are swallowed to retain legacy flow, critical anomalies fail silently.
3. **Hermetic Offline Execution of `test/packcompat`:** Air-gapped CI environments will fail to execute `test/packcompat` if direct remote repository checks are hard-gated without local cache/fixture fallback support.

---

## Top Strengths

- **Simplification of Loader Logic (§1763):** Moving away from the "inactive compatibility assets" loader support and requiring the compatibility pin to omit colliding assets entirely prevents duplicate-active-definition runtime crashes.
- **Role-Surface Inventory Expansion (§1823–1850):** Explicitly including required provider packs, examples, prompt fragments, and templates under the de-roling scan boundary ensures role leakage is caught before release.
- **Double-Pin Rollout Sequence (§1179–1200):** Transitioning through a backward-compatible "compatibility pin" (coexisting with legacy Maintenance) and a subsequent "activation pin" (post-Maintenance retirement) provides a highly resilient, low-risk migration path.
- **Behavior Witness Floor Rule (§313–318):** Mandating that every row in the behavior manifest has both old and final witness tests, and rejecting count-only or path-only validation, guarantees that no hidden behaviors are lost.

---

## Critical Risks & Gaps

### 1. The Provider Pack De-Roling vs. Byte-Continuity Contradiction
- **The Risk:** Active provider formulas like `examples/dolt/formulas/mol-dog-backup.toml` contain hardcoded notifications targeting `deacon/` (§189). This directly violates Core role neutrality because required provider packs are loaded in Core-only cities where Gastown roles do not exist.
- **The Gap:** The design claims that `bd` and `dolt` materialized bytes and provenance must remain strictly "unchanged except for expected manifest metadata" (§2562). If their bytes remain unchanged, they cannot be modified to resolve notification targets dynamically or remove hardcoded roles.
- **Recommendation:** Resolve this contradiction by updating the byte-continuity rule (§2562) to explicitly permit modifications of provider pack files for de-roling and role-cleaning purposes.

### 2. Silent Alert Failures and Empty Recipient Misroutes (§1310–1312)
- **The Risk:** In Core-only cities where no recipient is configured, dynamic recipient variables in generalized scripts like `reaper.sh` and `jsonl-export.sh` will evaluate to empty or slash targets (e.g., `gc mail send /`).
- **The Gap:** If the calling formula swallows the error (retaining legacy behavior), critical database backup, replication, or export failures will be swallowed silently. If the error is not swallowed, the entire wisp/maintenance step will fail due to an unconfigured optional notification.
- **Recommendation:** Mandate that generalized scripts perform preflight validation on recipient variables. If empty or resolving to `/`, the script should gracefully log a warning to stderr and skip mail execution (exiting with code `0`). If a recipient *is* configured but the alert fails, it must exit with a non-zero code.

### 3. Hermetic Offline Execution of `test/packcompat`
- **The Risk:** `test/packcompat` is designed to execute against the exact public Gastown commit on GitHub (§2169–2172).
- **The Gap:** In air-gapped or restricted CI environments, direct network requests to fetch the remote repository will fail, breaking Gas City's test baseline.
- **Recommendation:** Require that Gas City's test harness supports a hermetic, local-fixture-backed fallback for `test/packcompat` by caching or vendoring the pinned Gastown pack.

---

## Evaluation of the Three Key Questions

### 1. Does the behavior inventory enumerate every Gastown-specific requester, detector, notification path, formula, order, script branch, and prompt fragment removed from Core?
- **Auditor Finding:** **Yes, with provider-pack exceptions.** The asset migration map and the newly added "Full Role-Surface Inventory" (§1823) establish an exceptionally detailed framework. However, the inventory is currently incomplete regarding active provider-pack files (such as `mol-dog-backup.toml`). Once the provider-pack byte-continuity contradiction is resolved, the inventory will be 100% complete.

### 2. Which concrete `gascity-packs/gastown` tests prove each restored behavior fires under the same trigger conditions rather than merely existing?
- **Auditor Finding:** **Satisfactory with recommendations.** The introduction of `test/packcompat` (§2179) and the "Strict Behavior Witness Floor" (§313) ensures that path checks cannot replace execution-level tests. However, the test suite must mandate the use of **behavioral-trigger fixtures** that simulate failure or escalation states (such as warning/escalation pathways) to verify warning branches.

### 3. Can reviewers trace each high-risk Maintenance or Core move to old path, new path, landing commit, and observable test evidence?
- **Auditor Finding:** **Yes.** The behavior manifest structure (§299–310) and the 7-slice rollout plan (§2351–2427) are outstanding. It decouples high-risk steps, lists explicit gates for each slice, and enforces intermediate test runs, ensuring that no unverified code lands on the main branch.

---

## Required Changes for Finalization

1. **Resolve Provider Pack Contradiction:** Amend §2562 to permit modifications to `bd` and `dolt` bytes for role-cleaning purposes, or declare a formal exception row in the role-surface inventory.
2. **Enforce Script Preflight Validation:** Add a requirement that generalized scripts validate recipient variables, skipping execution with code `0` on empty/slash targets, and failing loudly on real failures.
3. **Mandate Behavioral-Trigger Fixtures:** Require `test/packcompat` to execute simulated failure/escalation states to verify warning/escalation pathways.
4. **Support Offline Test Mode:** Mandate a hermetic, local-fixture-backed execution path for `test/packcompat` to guarantee air-gapped CI success.
