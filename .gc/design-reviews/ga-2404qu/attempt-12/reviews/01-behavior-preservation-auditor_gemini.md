# Nadia Volkov — DeepSeek V4 Flash Perspective Independent Review (Iteration 12 / Attempt 12)

**Verdict:** approve-with-risks

**Scope:** Behavior preservation lane only — Gastown behavior inventory, cross-repo evidence chains, requester/detector/notification continuity, and preventing silent capability loss.

This review evaluates the Iteration 12 / Attempt 12 draft of `design.md` (`updated_at: 2026-06-07T14:05:04Z`) against `requirements.md` and the existing codebase behavior.

---

## Executive Summary

The Iteration 12 design represents an exceptionally solid and realistic blueprint for the Core/Gastown pack split. By retiring the speculative loader "inactive compatibility assets" fiction (§1763) and establishing the **Full Role-Surface Inventory and Replacement Gate (§1823–1850)**, the architecture successfully resolves the major rollout-collision and role-leakage concerns raised in previous iterations.

However, from a rigorous **Behavior Preservation Auditing** perspective, there are two remaining risks and logical contradictions that must be called out:
1. **The Required Provider Pack De-Roling vs. Byte-Continuity Contradiction:** The required provider packs (`bd` and `dolt`) are explicitly brought into the de-roling scope (§1831), yet the design continues to mandate that their materialized bytes and provenance "remain unchanged except for expected manifest metadata" (§2562). Active provider formulas like `mol-dog-backup.toml` contain hardcoded notifications targeting `deacon/` (§189). This contradiction prevents clean role-neutrality.
2. **Silent Alert Failures on Empty Dynamic Recipients:** Generalizing scripts to accept alert/warmup recipients from environment or formula variables (§1310–1312) creates a risk of silent capability loss if an empty or invalid recipient is passed. If the calling formula swallows the error, critical database backup or export anomalies will fail silently.

---

## Top Strengths

- **Retirement of Inactive Loader Fiction (§1763):** Replacing the speculative "inactive compatibility assets" loader support with a concrete requirement that the compatibility pin must omit colliding assets entirely is an outstanding improvement. It simplifies the loader and prevents duplicate-active-definition crashes.
- **Comprehensive Role-Surface Inventory (§1823–1850):** Explicitly expanding the de-roling and role-surface scan scope to cover required provider packs, examples, prompt fragments, and templates ensures that role leakage is audited at the source level.
- **Double-Pin Rollout Sequence (§1179–1200):** Decoupling the rollout into a backward-compatible "compatibility pin" (coexisting with legacy Maintenance) and an "activation pin" (post-Maintenance retirement) provides a highly resilient and safe migration path.
- **Strict Behavior Witness Floor (§313–318):** Enforcing that every row in the behavior manifest has both old and final witness tests, and rejecting path-only or count-only validation, guarantees that no hidden behaviors are lost.

---

## Critical Risks & Gaps

### 1. The Provider Pack De-Roling vs. Byte-Continuity Contradiction
- **The Risk:** Active provider formulas like `examples/dolt/formulas/mol-dog-backup.toml` contain hardcoded notifications targeting `deacon/` (§189). This directly violates Core role neutrality because required provider packs are loaded in Core-only cities.
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
