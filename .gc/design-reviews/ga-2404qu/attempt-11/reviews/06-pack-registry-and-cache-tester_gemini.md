# Marcus Driscoll — DeepSeek V4 Flash Perspective Independent Review (Iteration 11 / Attempt 11)

**Verdict:** approve-with-risks

**Scope:** Builtin pack registry identity, synthetic cache pruning, system pack materialization, and provider-dependent pack continuity.

This review evaluates the Iteration 11 draft of `design.md` (`updated_at: 2026-06-07T14:05:04Z`) against `requirements.md` and the existing codebase behavior.

---

## Executive Summary

The Iteration 11 design details a highly robust, secure, and clear plan for migrating to the new system pack structure. Decoupling the legacy `maintenance` and `gastown` packs from the embedded registry, moving the Core pack to `internal/packs/core`, and defining strict verification rules for synthetic and remote caches are massive wins for security, maintainability, and clean architecture.

However, from a meticulous **Registry, Cache, and Materialization Testing** perspective, there are several subtle but significant risks and omissions in the current Iteration 11 design draft:
1. **Offline Upgrade Continuity and Cache Promotion Failure:** Upgrading to a new binary invalidates the `IsSource(PublicGastown)` check for the old synthetic cache namespace. Without a transition-time promotion mechanism, offline/air-gapped cities will fail when attempting to fetch public Gastown remote.
2. **Failure-Recovery Deadlock on Corrupted Core Assets:** While `MaterializeBuiltinPacks` is specified to *overwrite* expected files on repair—the design lacks a requirement to prune unexpected or malicious files from required packs, leading to validation failure-recovery deadlock.
3. **Prompt Globbing Vulnerability from Preserved Stale Directories:** Stale preserved `maintenance` and `gastown` directories under `.gc/system/packs` can still be globbed by the template engine, shadowing the new active Core templates.
4. **Synthetic Content Hash Invalidation Disk I/O Storm:** Removing retired packs from the registry shifts the `SyntheticContentHash`, triggering an immediate, concurrent self-heal re-materialization across all local cities upon upgrade.
5. **Provider-Pack Role Leakage and `dog` Pool Name Ambiguity:** Hardcoded references to `pool = "dog"` in provider packs like `dolt` contradict role-neutrality and zero-hardcoded-roles requirements unless the provider packs can resolve this dynamically.

---

## Top Strengths

- **Decoupled Builtin Pack Registry (§824–838):** Stripping `maintenance` and `gastown` from `All()` and `requiredBuiltinPackNames` ensures a minimal, secure, and generic core SDK.
- **Content-Based Stale Cache Rejection (§690–705, §833):** Rejecting old synthetic cache namespaces and migrating public Gastown to the `source+version` namespace prevents stale/malicious caching.
- **Preservation of Stale System-Pack Directories (§213, §836):** Mandating that `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` are ignored and preserved rather than deleted prevents accidental deletion of operator modifications or custom local overrides.
- **Fail-Safe Repair of Required Packs (§57–60, §349–356):** Enforcing that required packs (`core`, `bd`, `dolt`) are always refreshed and validated protects the SDK from corrupted dependencies.

---

## Critical Risks & Gaps

### 1. Offline Upgrade Continuity and Cache Promotion
- **The Risk:** In the new design, public Gastown is resolved through its ordinary remote path and version, making `IsSource(PublicGastown)` false. This means any attempt to fetch or load it will fall through to a remote git checkout, which requires internet access.
- **The Gap:** For an offline/air-gapped city, if they upgrade to the new binary, they will no longer be able to resolve `gascity-packs/gastown` because their existing locks reference the old synthetic cache namespace (`bundled-synthetic-v1\x00`). Without a promotion/migration mechanism that copies un-tampered synthetic caches to the ordinary remote cache path, the upgrade will fail on offline machines.
- **Required Recommendation:** Mandate that the cache/registry layer implements a one-time "cache promotion" helper. If a valid synthetic cache exists for the pinned public Gastown commit, and the machine is offline, the loader should promote (copy) the validated synthetic cache files into the ordinary remote cache path before disabling synthetic-source resolution.

### 2. Failure-Recovery Deadlock on Corrupted/Extra Files
- **The Risk:** If there are unexpected or malicious files under required packs (like `core`), config loading fails closed because synthetic cache validation fails.
- **The Gap:** `MaterializeBuiltinPacks` is specified to *overwrite* expected files on repair—but it doesn't *prune* unexpected ones, leading to recovery deadlock where the system remains invalid on successive starts.
- **Required Recommendation:** Amend the materialization contract to explicitly require that `MaterializeBuiltinPacks` (and the underlying `materializeFS` or `pruneStaleGeneratedPackFiles`) removes any unexpected files or directories from required system pack directories (`core`, `bd`, `dolt`) during repair.

### 3. Prompt Globbing Vulnerabilities and Preserved Stale Directories
- **The Risk:** Stale directories (like `maintenance`) left on disk can still have their prompt templates globbed and loaded, shadowing the active Core templates.
- **The Gap:** The design states that stale `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories are ignored, diagnosed, and preserved. However, unless globbing is explicitly filtered by `requiredBuiltinPackNames` or active resolved packs, the template engine will still discover and load prompts from these stale preserved directories.
- **Required Recommendation:** Require that the prompt/template loader is strictly constrained to walk only active resolved packs and required host packs. It must explicitly blacklist or avoid globbing any preserved stale system-pack directories.

### 4. Synthetic Content Hash Invalidation Disk I/O Storm
- **The Risk:** Removing `maintenance` and `gastown` from `All()` changes the `SyntheticContentHash()`.
- **The Gap:** This change invalidates all existing synthetic caches for `core`, `bd`, and `dolt` because the marker file's `content_hash` will no longer match the new binary's hash. This will force a one-time self-heal re-materialization of all synthetic caches upon upgrade, causing a simultaneous disk I/O storm of self-healing across concurrent cities/rigs on a host.
- **Required Recommendation:** Document this performance/migration risk in the upgrade guide and run tests to verify that concurrent city starts do not corrupt each other's in-flight atomic temp files during this self-heal storm.

### 5. `dog` Pool Name Ambiguity and Role-Neutrality
- **The Risk:** The `pool = "dog"` is hardcoded in provider packs (`dolt`), which contradicts the zero-hardcoded-roles / "renamable dog" requirement.
- **The Gap:** Since provider packs must remain byte-identical except for expected manifest metadata, they cannot easily be cleaned of hardcoded pool targets unless we allow provider-pack-specific role exceptions or dynamic target resolution.
- **Required Recommendation:** Provide a clear mechanism for provider packs (`dolt`, `bd`) to resolve their role target dynamically from the city or pool configuration, or explicitly register `dog` as an allowed Core role exception in the role-surface inventory.

---

## Evaluation of the Three Key Questions

### 1. Do registry and embed tests assert that only the intended built-in packs remain, with Core sourced from the new path and no Gastown or Maintenance aliases?
- **Auditor Finding:** **Yes, with slice gates.** The Registry, Cache, and Materializer Slice Gates (§824–838) ensure that `All()` is correctly pruned, and `requiredBuiltinPackNames` is updated to remove Maintenance only after the packcompat no-Maintenance mode is green. The design correctly specifies negative/rejection tests to prove stale sources and aliases are ignored/rejected.

### 2. Does MaterializeBuiltinPacks repair missing or tampered Core while preserving provider-dependent bd and dolt behavior exactly as before?
- **Auditor Finding:** **Satisfactory with recommendations.** The design mandates that required packs are always refreshed and validated, preserving provider-dependent `bd` and `dolt` behavior. However, to prevent recovery deadlock, we must ensure that `MaterializeBuiltinPacks` actively prunes unexpected files from these required pack directories.

### 3. Do synthetic cache tests reject modified manifests, unexpected files, and stale retired pack sources rather than checking file existence only?
- **Auditor Finding:** **Yes.** The design specifies rigorous content validation tests including stale-cache rejection, stale directory preservation, and validating content integrity and provenance (via content hashes and manifest validation) rather than just checking path existence.

---

## Required Changes for Finalization

1. **Cache Promotion Helper:** Add a requirement for a one-time "cache promotion" helper to handle offline upgrades by copying un-tampered synthetic caches to the ordinary remote cache path.
2. **Pruning Unexpected Files on Repair:** Mandate that `MaterializeBuiltinPacks` removes unexpected files/folders from required system pack directories during repair to avoid deadlock.
3. **Template Engine Directory Blacklist:** Add a rule constraining prompt/template globbing to active/resolved packs, ensuring preserved stale directories are completely ignored by the template loader.
4. **I/O Storm Warning:** Document the synthetic content hash invalidation and subsequent I/O storm in migration docs.
5. **Dynamic Role Resolution in Provider Packs:** Clarify how provider packs resolve their notification/pool targets dynamically, or register `dog` as a formal role exception.
