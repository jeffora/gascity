# Marcus Driscoll — DeepSeek V4 Flash Perspective Independent Review (Iteration 12 / Attempt 12)

**Verdict:** approve-with-risks

**Scope:** Builtin pack registry identity, synthetic cache pruning, system pack materialization, and provider-dependent pack continuity.

This review evaluates the Iteration 12 / Attempt 12 draft of `design.md` (`updated_at: 2026-06-07T14:05:04Z`) against `requirements.md` and the existing codebase behavior, focusing specifically on the new registry, cache, materializer, and offline gates established in the **Attempt 11 Review Resolution Contracts (§1694–1951)**.

---

## Executive Summary

The Iteration 12 design represents a monumental leap forward in addressing the core architectural challenges of the Core/Gastown pack split. By introducing the **Concrete Core Maintenance-Worker Binding (§1703–1761)**, the design elegantly resolves the previous `dog` pool contract ambiguity and hardcoded role leakage using dynamic target bindings (`core.maintenance_worker`). Furthermore, the **Required Core Loading Fatal Gates (§1789–1821)** and the unified cache identity rules (§1893) establish robust, dual-stage security and validation baselines that prevent configuration loading under compromised states.

However, from a rigorous **Registry, Cache, and Materialization Testing** perspective, several critical risks and logical contradictions remain:
1. **Failure-Recovery Deadlock on Required Pack Tampering:** The pre-resolution file-set integrity gate is strict, but required pack materialization does not explicitly prune unexpected files, causing a permanent startup deadlock if an un-tracked file is injected.
2. **Offline Upgrade Continuity and Cache Promotion Deficit:** Disabling synthetic fallback for public Gastown and enforcing strict network-only remote resolution offline leaves existing air-gapped cities stranded upon upgrading, as their namespaced caches are retired with no automated migration path.
3. **The Provider Pack De-Roling vs. Byte-Continuity Contradiction:** Mandating that required provider packs (`bd` and `dolt`) are de-roled (§1831) while simultaneously forbidding any byte-level modifications (§2562) creates a direct contradiction, as active provider formulas still contain hardcoded role targets (such as `deacon` in `mol-dog-backup.toml`).

---

## Top Strengths

- **Symbolic Core Maintenance-Worker Binding (§1703–1761):** Abstracting the maintenance agent assignment into `gc.bindings.maintenance_worker` with configurable defaults (`dog`) and supporting symbolic patching via `target_binding` is an outstanding design decision. It preserves absolute ZFC/role neutrality in Go source code while maintaining standard operations.
- **Strict Required Pack Loading Fatal Gates (§1789–1821):** Implementing a dual-stage fatal loading model—verifying pre-resolution required file-set integrity against exact manifests and digests, followed by post-resolution typed `RequiredSystemPackParticipation` validation—ensures that the core SDK is tamper-proof and fully participating.
- **Unified Cache Identity and Synthetic Alias Retirement (§1893–1907):** Restricting `SyntheticContentHash` exclusively to bundled host packs (`core`, `bd`, `dolt`) and forcing public Gastown to resolve via standard remote namespaces (keyed by normalized source, commit, and subpath) completely eliminates the historical risk of synthetic cache hijacking.
- **Advisory Directory Locks for Self-Heal Contention (§1914–1915):** Introducing advisory directory locks to handle concurrent self-heal contention across different cities/rigs on a host is a highly robust mitigation for the disk I/O storms triggered by `SyntheticContentHash` invalidation during binary upgrades.

---

## Critical Risks & Gaps

### 1. Failure-Recovery Deadlock on Required Pack Tampering
- **The Risk:** The design introduces a strict pre-resolution required file-set integrity gate (§1799) that validates files against exact manifests and digests. If an unexpected file or malicious script is placed in `.gc/system/packs/core/`, validation fails and config loading fails closed, preventing the CLI or controller from starting.
- **The Gap:** The required system pack materialization contract (`MaterializeRequiredPacks` §2246) is specified to repair/overwrite missing or corrupted *expected* files, but it does not specify pruning of *unexpected* files. Since the unexpected file is never deleted during startup materialization or normal self-heal, successive runs will continue to fail the integrity check, leading to a permanent failure-recovery deadlock.
- **Recommendation:** Mandate that `MaterializeRequiredPacks` (and its underlying materialization helpers) performs directory pruning to actively delete unexpected files and folders from the `core`, `bd`, and `dolt` system pack directories before validating file-set integrity.

### 2. Offline Upgrade Continuity and Cache Promotion
- **The Risk:** Upgrading the Gas City binary changes `All()` and invalidates the `SyntheticContentHash`, rendering existingNamespaced synthetic caches (`bundled-synthetic-v1\x00`) stale and un-selectable for new pins (§1907).
- **The Gap:** The design declares that fresh Gastown init is network-required offline and fallback to synthetic fallback is forbidden (§2211–2215). For an existing air-gapped city, if they upgrade to the new binary, they will be unable to resolve their pinned public Gastown commit because their existing `packs.lock` uses the old cache namespace and they cannot reach GitHub. They will be left stranded.
- **Recommendation:** Require a one-time "cache promotion" helper. Under offline constraints, if a valid, un-tampered synthetic cache for the pinned public Gastown commit is present on disk, the loader should safely copy and re-key its contents into the ordinary remote cache path before disabling synthetic-source resolution, ensuring seamless offline upgrade continuity.

### 3. Required Provider Pack De-Roling vs. Byte-Continuity Contradiction
- **The Risk:** Required provider packs (`bd` and `dolt`) are explicitly included in the role-surface inventory and de-roling scope (§1831). Active formulas in those packs (such as `examples/dolt/formulas/mol-dog-backup.toml` §189) contain hardcoded roles (targeting `deacon/` notifications).
- **The Gap:** The design continues to assert that `bd` and `dolt` materialized bytes and provenance must "remain unchanged except for expected manifest metadata" (§2562). This is a logical contradiction: we cannot clean hardcoded roles from provider-pack formulas if we are forbidden from modifying their bytes.
- **Recommendation:** Align with Nadia Volkov's lane by updating §2562 to explicitly permit modifications of provider-pack bytes for role-cleaning and target-binding rewrites, or declare them as formal, time-bound exceptions in the role-surface inventory.

### 4. Hermetic Offline Execution of `test/packcompat`
- **The Risk:** The rollout requires running `test/packcompat` against the exact pinned public Gastown commit on GitHub (§2169–2172).
- **The Gap:** In restricted or air-gapped CI/CD build servers, direct git checkouts of remote repositories will fail, leading to build breakage.
- **Recommendation:** Mandate that the Gas City test harness provides a hermetic, local-fixture-backed mode for `test/packcompat` that utilizes cached or vendored public pack assets when network access is unavailable.

---

## Evaluation of the Three Key Questions

### 1. Do registry and embed tests assert that only the intended built-in packs remain, with Core sourced from the new path and no Gastown or Maintenance aliases?
- **Finding:** **Yes.** The design-after specifications explicitly prune `All()` (§2194) and requiredBuiltinPackNames (§2243) to contain only `core`, `bd`, and `dolt`. Registry tests (`registry_test.go` §2217) are updated to verify these exact identities, asserting that `core` is sourced from `internal/packs/core` and verifying that historical Gastown or Maintenance aliases are rejected.

### 2. Does MaterializeBuiltinPacks repair missing or tampered Core while preserving provider-dependent bd and dolt behavior exactly as before?
- **Finding:** **Satisfactory with Risks.** The materializer correctly refreshes and validates `core`, `bd`, and `dolt` (§2246), preserving their provider-dependent behavior. However, to prevent recovery deadlock, the materialization contract must be extended to prune unexpected files rather than only overwriting expected ones.

### 3. Do synthetic cache tests reject modified manifests, unexpected files, and stale retired pack sources rather than checking file existence only?
- **Finding:** **Yes.** Section §2217 specifies rigorous content-validation tests, including stale-cache rejection, stale directory preservation, manifest verification, and testing of integrity and provenance (via content hashes and manifest validation) rather than mere file existence.

---

## Required Changes for Finalization

1. **Enforce Directory Pruning on Required Pack Repair:** Update §2246 to require that `MaterializeRequiredPacks` prunes unexpected files and folders from the `core`, `bd`, and `dolt` system pack directories during materialization.
2. **Implement Legacy-to-Remote Cache Promotion:** Add a requirement for a one-time "cache promotion" helper to transition valid legacy namespaced synthetic caches of public Gastown to standard remote cache paths for offline upgrade continuity.
3. **Resolve Provider Byte-Continuity Contradiction:** Amend §2562 to allow target-binding rewrites and role-cleaning within `bd` and `dolt` pack assets, or formally register them as exceptions.
4. **Hermetic Test Harness Fallback:** Mandate support for a local-fixture-backed fallback in `test/packcompat` to ensure deterministic execution in air-gapped CI environments.
