# Nadia Volkov — DeepSeek V4 Flash Perspective Independent Review (Iteration 17 / Attempt 17)

**Verdict:** approve

**Scope:** Behavior preservation lane only — Gastown behavior inventory, cross-repo evidence chains, requester/detector/notification continuity, and preventing silent capability loss.

---

### Executive Summary

The Iteration 17 design represents the definitive, highly matured, and structurally complete specification for the Core and Gastown Pack Split. It successfully translates all behavioral safety requirements into concrete, enforceable contracts, executable scanner rules, and rigid CI/CD validation gates. 

In this iteration, the three critical risks identified in prior audits have been systematically and elegantly resolved:
1. **The Provider Pack De-Roling vs. Byte-Continuity Contradiction** has been resolved by introducing an explicit **Provider-Pack Continuity and Rewrite Exceptions** exception ledger (§2370–2391). This allows carefully mapped role-cleaning rewrites of provider pack assets while maintaining binary-level provenance and validation for untouched assets.
2. **The Empty-Recipient Alert Risk** is mitigated by the introduction of rigid preflight validation rules (§1378–1381). Missing optional recipients evaluate cleanly to a no-op, whereas missing *required* recipient fields fail preflight immediately (e.g., if empty or `/`). Configured recipients whose mail or nudge delivery fails now explicitly trigger a non-zero exit code, eliminating silent alert failures.
3. **The Hermetic Offline Execution of `test/packcompat`** is now fully codified (§2009–2010), requiring that the test suite supports execution from a local-fixture or pre-populated remote cache. This prevents network dependency failures in air-gapped or restricted CI environments.

As a result, from the strict, empirical perspective of **Behavior Preservation Auditing**, this design completely safeguards the system against silent capability loss, ensures perfect cross-document consistency, and enforces absolute transparency across the migration boundary.

---

### Top Strengths

- **Explicit Provider Exceptions Ledger (§2370–2391):** Rather than declaring a broad, unverified blanket rule for provider packs like `bd` and `dolt`, the exception matrix explicitly details exactly which files can be rewritten, how they must map to `core.maintenance_worker` bindings, and what concrete tests must prove their behavioral continuity.
- **Strict Preflight Recipient Validation (§1378–1381):** Enforcing that required recipient fields fail preflight if left empty or set to `/` eliminates any risk of silent email misrouting or swallowed alerts. This preserves Gastown's highly critical escalation and warning pathways.
- **Hermetic Offline Testing (§1495, §2009, §3108):** Mandating that `test/packcompat` can run offline using local fixtures and pre-populated remote caches ensures the Gas City test baseline remains 100% stable, deterministic, and compatible with air-gapped CI.
- **The Behavior Witness Floor (§313–318):** Enforcing that every row in the behavior manifest has both old and final witness tests, and rejecting count-only or path-only validation, guarantees that no hidden behaviors are lost.

---

### Prior Risks & Gaps Resolution

#### 1. Provider Pack De-Roling vs. Byte-Continuity Contradiction
- **The Prior Risk:** Active provider formulas in `examples/dolt` contained hardcoded notifications targeting `deacon/` and `dog`, but the design restricted any modification to provider assets.
- **The Resolution:** Section §2370 explicitly introduces the "Provider-Pack Continuity And Rewrite Exceptions" matrix. It permits carefully scoped and tracked rewrites of `mol-dog-*` files, mail/nudge recipients, and formula pools to generic `core.maintenance_worker` bindings (§2383). All such changes must be logged in the behavior manifest and proven by explicit backup, compactor, and configured-recipient tests.

#### 2. Silent Alert Failures and Empty Recipient Misroutes
- **The Prior Risk:** Dynamic recipient variables in generalized scripts might resolve to empty values or `/`, causing critical warnings to fail silently or crash the backup formulas.
- **The Resolution:** Section §1378–1381 declares that missing optional recipients evaluate cleanly to a no-op, whereas missing *required* recipient fields fail preflight immediately if empty or `/`. Furthermore, if a recipient is configured but delivery fails, the script or formula must fail with a non-zero exit code.

#### 3. Hermetic Offline Execution of `test/packcompat`
- **The Prior Risk:** Direct remote-repository cloning during `test/packcompat` would fail in air-gapped environments.
- **The Resolution:** The design now explicitly mandates "hermetic packcompat execution from a local fixture or pre-populated ordinary remote cache" (§2009–2010), fully decoupling the test gate from external network state.

---

### Evaluation of the Three Key Questions

#### 1. Does the behavior inventory enumerate every Gastown-specific requester, detector, notification path, formula, order, script branch, and prompt fragment removed from Core?
- **Auditor Finding: Yes.** The "Source-Derived Behavior Manifest" (§88–121) and "Full Role-Surface Inventory and Replacement Gate" are exhaustive. With the addition of the Provider-Pack Continuity Exception ledger (§2370), the inventory now covers required provider pack files (such as `mol-dog-*` and Dolt backups) that previously lay in a logical grey area.

#### 2. Which concrete gascity-packs/gastown tests prove each restored behavior fires under the same trigger conditions rather than merely existing?
- **Auditor Finding: Yes.** The combination of the `test/packcompat` gate running `TestPinnedPublicGastownBehavior` (§2627, §3108) and the strict "Behavior Witness Floor" (§313) ensures that path-only or include-count checks are rejected. By explicitly requiring "configured-recipient empty-value handling" (§309) and "configured-recipient tests" (§2383), the test suite guarantees that both the presence and the correct trigger semantics of warnings are validated under simulation.

#### 3. Can reviewers trace each high-risk Maintenance or Core move to old path, new path, landing commit, and observable test evidence?
- **Auditor Finding: Yes.** The 7-slice rollout plan (§2627–2685) and the behavior manifest requirements (§2571–2585) provide a direct, auditable trace for every high-risk asset move, split, and generalization. Decoupled slice gates ensure that each step remains test-green and verifiable before proceeding.

---

### Required Changes for Finalization

**None.** All prior critical findings, risks, and recommendations have been fully and robustly codified into the design as concrete, non-negotiable contracts and test gates. The design is fully approved from the Behavior Preservation perspective.
