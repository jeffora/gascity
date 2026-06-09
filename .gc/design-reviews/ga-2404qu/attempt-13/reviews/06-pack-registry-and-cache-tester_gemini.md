# Marcus Driscoll — DeepSeek V4 Flash Perspective Independent Review (Iteration 13 / Attempt 13)

**Verdict:** approve-with-risks

**Scope:** Builtin pack registry identity, synthetic cache pruning, system pack materialization, and provider-dependent pack continuity.

This review evaluates the Iteration 13 / Attempt 13 draft of `design.md` (`updated_at: 2026-06-07T14:05:04Z`) against `requirements.md` and the existing codebase behavior, focusing specifically on the new registry, cache, materializer, and offline gates established in the **Attempt 11/12 Review Resolution Contracts**.

---

## Executive Summary

The Iteration 13 design represents a monumental leap forward in addressing the core architectural challenges of the Core/Gastown pack split. By introducing the **Concrete Core Maintenance-Worker Binding**, the design elegantly resolves the previous `dog` pool contract ambiguity and hardcoded role leakage using dynamic target bindings (`core.maintenance_worker`). Furthermore, the **Required Core Loading Fatal Gates** and the unified cache identity rules establish robust, dual-stage security and validation baselines that prevent configuration loading under compromised states.

However, from a rigorous **Registry, Cache, and Materialization Testing** perspective, several critical risks and logical contradictions remain:
1. **Failure-Recovery Deadlock on Required Pack Tampering:** The pre-resolution file-set integrity gate is strict, but required pack materialization does not explicitly prune unexpected files, causing a permanent startup deadlock if an un-tracked file is injected.
2. **Offline Upgrade Continuity and Cache Promotion Deficit:** Disabling synthetic fallback for public Gastown and enforcing strict network-only remote resolution offline leaves existing air-gapped cities stranded upon upgrading, as their namespaced caches are retired with no automated migration path.
3. **The Provider Pack De-Roling vs. Byte-Continuity Contradiction:** Mandating that required provider packs (`bd` and `dolt`) are de-roled while simultaneously forbidding any byte-level modifications creates a direct contradiction, as active provider formulas still contain hardcoded role targets (such as `deacon` in `mol-dog-backup.toml`).
4. **Hermetic Offline Execution of `test/packcompat`:** The rollout requires running `test/packcompat` against the exact pinned public Gastown commit on GitHub, which will fail on restricted or air-gapped CI/CD build servers without local fallback.

---

## Top Strengths

- **Symbolic Core Maintenance-Worker Binding:** Abstracting the maintenance agent assignment into `gc.bindings.maintenance_worker` with configurable defaults (`dog`) and supporting symbolic patching via `target_binding` is an outstanding design decision. It preserves absolute ZFC/role neutrality in Go source code while maintaining standard operations.
- **Strict Required Pack Loading Fatal Gates:** Implementing a dual-stage fatal loading model—verifying pre-resolution required file-set integrity against exact manifests and digests, followed by post-resolution typed `RequiredSystemPackParticipation` validation—ensures that the core SDK is tamper-proof and fully participating.
- **Unified Cache Identity and Synthetic Alias Retirement:** Restricting `SyntheticContentHash` exclusively to bundled host packs (`core`, `bd`, `dolt`) and forcing public Gastown to resolve via standard remote namespaces (keyed by normalized source, commit, and subpath) completely eliminates the historical risk of synthetic cache hijacking.
- **Advisory Directory Locks for Self-Heal Contention:** Introducing advisory directory locks to handle concurrent self-heal contention across different cities/rigs on a host is a highly robust mitigation for the disk I/O storms triggered by `SyntheticContentHash` invalidation during binary upgrades.

---

## Critical Risks & Gaps

### 1. Failure-Recovery Deadlock on Required Pack Tampering
- **The Risk:** The design introduces a strict pre-resolution required file-set integrity gate that validates files against exact manifests and digests. If an unexpected file or malicious script is placed in `.gc/system/packs/core/`, validation fails and config loading fails closed, preventing the CLI or controller from starting.
- **The Gap:** The required system pack materialization contract (`MaterializeRequiredPacks`) is specified to repair/overwrite missing or corrupted *expected* files, but it does not specify pruning of *unexpected* files. Since the unexpected file is never deleted during startup materialization or normal self-heal, successive runs will continue to fail the integrity check, leading to a permanent failure-recovery deadlock.
- **Recommendation:** Mandate that `MaterializeRequiredPacks` (and its underlying materialization helpers) performs directory pruning to actively delete unexpected files and folders from the `core`, `bd`, and `dolt` system pack directories before validating file-set integrity.

### 2. Offline Upgrade Continuity and Cache Promotion
- **The Risk:** Upgrading the Gas City binary changes `All()` and invalidates the `SyntheticContentHash`, rendering existing namespaced synthetic caches (`bundled-synthetic-v1\x00`) stale and un-selectable for new pins.
- **The Gap:** The design declares that fresh Gastown init is network-required offline and fallback to synthetic fallback is forbidden. For an existing air-gapped city, if they upgrade to the new binary, they will be unable to resolve their pinned public Gastown commit because their existing `packs.lock` uses the old cache namespace and they cannot reach GitHub. They will be left stranded.
- **Recommendation:** Require a one-time "cache promotion" helper. Under offline constraints, if a valid, un-tampered synthetic cache for the pinned public Gastown commit is present on disk, the loader should safely copy and re-key its contents into the ordinary remote cache path before disabling synthetic-source resolution, ensuring seamless offline upgrade continuity.

### 3. Required Provider Pack De-Roling vs. Byte-Continuity Contradiction
- **The Risk:** Required provider packs (`bd` and `dolt`) are explicitly included in the role-surface inventory and de-roling scope. Active formulas in those packs (such as `examples/dolt/formulas/mol-dog-backup.toml`) contain hardcoded roles (targeting `deacon/` notifications).
- **The Gap:** The design continues to assert that `bd` and `dolt` materialized bytes and provenance must "remain unchanged except for expected manifest metadata". This is a logical contradiction: we cannot clean hardcoded roles from provider-pack formulas if we are forbidden from modifying their bytes.
- **Recommendation:** Update the specification to explicitly permit modifications of provider-pack bytes for role-cleaning and target-binding rewrites, or declare them as formal, time-bound exceptions in the role-surface inventory.

### 4. Hermetic Offline Execution of `test/packcompat`
- **The Risk:** The rollout requires running `test/packcompat` against the exact pinned public Gastown commit on GitHub.
- **The Gap:** In restricted or air-gapped CI/CD build servers, direct git checkouts of remote repositories will fail, leading to build breakage.
- **Recommendation:** Mandate that the Gas City test harness provides a hermetic, local-fixture-backed mode for `test/packcompat` that utilizes cached or vendored public pack assets when network access is unavailable.

---

## Evaluation of the Three Key Questions

### 1. Do registry and embed tests assert that only the intended built-in packs remain, with Core sourced from the new path and no Gastown or Maintenance aliases?
- **Finding:** **Yes.** The design-after specifications explicitly prune `All()` and requiredBuiltinPackNames to contain only `core`, `bd`, and `dolt`. Registry tests (`registry_test.go`) are updated to verify these exact identities, asserting that `core` is sourced from `internal/packs/core` and verifying that historical Gastown or Maintenance aliases are rejected.

### 2. Does MaterializeBuiltinPacks repair missing or tampered Core while preserving provider-dependent bd and dolt behavior exactly as before?
- **Finding:** **Satisfactory with Risks.** The materializer correctly refreshes and validates `core`, `bd`, and `dolt`, preserving their provider-dependent behavior. However, to prevent recovery deadlock, the materialization contract must be extended to prune unexpected files rather than only overwriting expected ones.

### 3. Do synthetic cache tests reject modified manifests, unexpected files, and stale retired pack sources rather than checking file existence only?
- **Finding:** **Yes.** The design specifies rigorous content-validation tests, including stale-cache rejection, stale directory preservation, manifest verification, and testing of integrity and provenance (via content hashes and manifest validation) rather than mere file existence.

---

## Evaluation of Red Flags

### 1. Old `internal/bootstrap/packs/core` or retired aliases are still accepted silently
- **Analysis:** This is successfully mitigated. The design explicitly defines the retirement of these aliases and specifies unit tests to reject any attempt to load them.

### 2. `bd` or `dolt` pack behavior changes while repairing Core
- **Analysis:** High risk of drift/contradiction. The materialization layer must treat `bd` and `dolt` with byte-for-byte fidelity while de-roling changes are performed. We must reconcile the byte-continuity constraint with the role-cleaning requirements.

### 3. Tamper tests count paths instead of validating content and provenance
- **Analysis:** Successfully addressed in the testing specifications, which require cryptographic digest and content-level verification rather than simple file/path counts.

---

## Required Changes for Finalization

1. **Enforce Directory Pruning on Required Pack Repair:** Require that `MaterializeRequiredPacks` prunes unexpected files and folders from the `core`, `bd`, and `dolt` system pack directories during materialization.
2. **Implement Legacy-to-Remote Cache Promotion:** Add a requirement for a one-time "cache promotion" helper to transition valid legacy namespaced synthetic caches of public Gastown to standard remote cache paths for offline upgrade continuity.
3. **Resolve Provider Byte-Continuity Contradiction:** Amend the specification to allow target-binding rewrites and role-cleaning within `bd` and `dolt` pack assets, or formally register them as exceptions.
4. **Hermetic Test Harness Fallback:** Mandate support for a local-fixture-backed fallback in `test/packcompat` to ensure deterministic execution in air-gapped CI environments.
