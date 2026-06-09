# Marcus Driscoll — DeepSeek V4 Flash Review (Iteration 2)

**Persona:** Pack Registry and Cache Tester
**Mandate:** Builtin registry identity, synthetic cache pruning, system pack materialization, provider-dependent pack continuity
**Verdict:** block

---

## Top Strengths

1. **Deterministic Builtin Registry Identity contract:** 
   The design cleanly establishes `All()` → `{core=internal/packs/core, bd=examples/bd, dolt=examples/dolt}` with no legacy Maintenance or Gastown aliases (lines 291-297). The complete removal of the `publicSubpathForPack` synthetic layout mapping for retired packs provides a crisp, testable target that prevents silent alias leaks by construction.

2. **Structured Rollout Plan & Milestones:** 
   The 7-stage Rollout plan (lines 656-702) is exceptionally well-calibrated. Segmenting the migration into candidate public Gastown, packcompat tests, Core extraction, Core loading/doctor, Maintenance folding, Registry/cache slice, and final source deletion ensures each commit keeps focused suites and `make test-fast-parallel` passing across repos.

3. **Maintenance Retirement Runtime Table:** 
   The inclusion of a dedicated surface-by-surface compatibility specification (lines 337-351) is highly professional. It lists the target behavior and required proof for every affected runtime surface (such as `requiredBuiltinPackNames`, `builtinPackIncludes`, `MaterializeBuiltinPacks`, orders/scripts, and stale directories), leaving no room for implementation ambiguity during the Maintenance → Core fold.

4. **Robust Behavior-Preservation Inventory requirement:** 
   Requiring `gastown/docs/behavior-preservation.md` in the `gascity-packs` repo to map every moved or generalized asset (old path, old trigger/requester, semantic delta, and corresponding new paths and tests, lines 226-287) is an excellent safeguard against the highest risk of this migration: losing Gastown behavior while generalizing Core.

---

## Blocker Findings

### [Blocker] `loadBaselinePrompt` globbing stale templates creates a massive prompt-override vulnerability

* **Evidence:** `cmd/gc/cmd_prompt.go:613-621`:
  ```go
  // 2. Pack defaults — scan all materialized packs.
  packGlob := filepath.Join(cityPath, ".gc", "system", "packs", "*", "agents", role, "prompt.template.md")
  if matches, err := filepath.Glob(packGlob); err == nil {
      sort.Strings(matches)
      for _, m := range matches {
          if data, err := os.ReadFile(m); err == nil {
              rel, _ := filepath.Rel(cityPath, m)
              return string(data), "pack default at " + rel, true
          }
      }
  }
  ```
* **The Problem:** The design explicitly specifies that stale `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories are **preserved on disk, ignored, and not deleted** on startup or during doctor fixes (lines 55-56, 331-335, 720-723). However, because `loadBaselinePrompt` uses a directory-level wildcard glob (`packs/*`), it will **still discover and read** these stale templates!
* **Impact:** For any role that was present in the retired packs (e.g., `mayor`, `dog`), an upgraded city will silently load stale prompt templates from the retired, no-longer-included directories on disk instead of falling back to the updated embedded Core templates or the newly installed remote Gastown pack templates. This completely violates the "ignored legacy state" guarantee.
* **Required Change:** Extend the "retired directory" contract to cover `loadBaselinePrompt`. Either:
  - (a) Modify `loadBaselinePrompt` to restrict its scan to directories listed in the resolved config imports/includes rather than globbing `*` directly.
  - (b) Filter matches in `loadBaselinePrompt` against the active `requiredBuiltinPackNames` so only active built-in packs supply baseline templates.
  - (c) Mandate that `gc doctor --fix` renames or moves stale prompt templates out of the glob path, or add a diagnostic warning.

---

### [Blocker] `RepoCacheKey` namespace key flip orphans existing namespaced public Gastown cache entries and breaks offline upgrades

* **Evidence:** `internal/config/pack_include.go:304-306` and `internal/packman/cache.go:63`.
* **The Problem:** Currently, bundled packs are keyed in the cache under the `bundled-synthetic-v1` namespace prefix via `RepoCacheKey`. After migration, `IsSource(PublicGastownPackSource)` will return `false`, meaning any new resolution of public Gastown will compute its cache key as the ordinary remote path (no prefix).
  The design states: *"Existing synthetic cache entries for public Gastown or Maintenance are stale after this migration. They may remain on disk, but source resolution must not select them for a public sha: pin."* (lines 307-309).
* **Impact:** If an upgraded binary runs **offline** on an existing city whose `packs.lock` was generated pre-migration, `resolveLockedRemoteImport` will compute the unprefix/un-namespaced key, look for it, fail to find it (as the folder was cached under the `bundled-synthetic-v1` prefix), and try to fetch from git. Since the operator is offline, the fetch will fail.
* **Required Change:** The design must specify a concrete upgrade/migration strategy for namespaced cache entries. We must choose and explicitly test one of:
  - **(a) Dual-read fallback lookup:** If the computed un-namespaced cache path is absent, check if a namespaced cache folder exists for the exact same source/version, and if so, fall back to reading from it or copy it.
  - **(b) One-time migration:** Make `gc doctor --fix` detect stale namespaced public Gastown caches and copy/migrate them to the new un-namespaced remote cache path.
  - **(c) Doctor diagnostic with online requirement:** Accept that offline load will fail, but require `gc doctor` to report the version-skew mismatch and instruct the operator to run `gc import install` while online.

---

### [Blocker] Lack of concrete list of packman/config tests to invert or delete (silently accepting retired aliases)

* **Evidence:** Five production sites branch on `IsSource(source)` (`pack_include.go:213, 304`; `cache.go:63`; `check.go:231`; `install.go:49`).
* **The Problem:** There are existing integration and unit tests that assert public Gastown synthetic acceptance and offline embedded fallbacks—chiefly `TestSyncLockUsesBundledFallbackForPublicGastownWhenRemoteUnavailable` (`internal/packman/check_test.go:49-91`) and `TestResolveImportPackRefAcceptsPublicGastownSyntheticCache` (`internal/config/bundled_import_test.go:98-113`).
* **Impact:** While the design names "install-lock generation" and "materialization" negative registry tests, it fails to enumerate these integration-level tests. If left unmodified, these tests will silently fail or encode the old, retired alias acceptance behavior.
* **Required Change:** Enumerate all five `IsSource` call sites in the test plan. Explicitly list `TestSyncLockUsesBundledFallbackForPublicGastownWhenRemoteUnavailable` and `TestResolveImportPackRefAcceptsPublicGastownSyntheticCache` as tests that must be deleted or rewritten to assert the new remote-only cache resolution path and offline fetch failures.

---

## Major Findings

### [Major] Required Core system pack file-set integrity is too weak

* **Evidence:** Lines 612-614: *"Core integrity: tampered Core is repaired; unexpected files either fail full file-set integrity validation or are proven unable to influence loaded formulas..."*
* **The Problem:** `MaterializeBuiltinPacks` already prunes non-manifest files for required packs via `pruneStaleGeneratedPackFiles`. However, the doctor's check path `packContainsEmbeddedManifest` only verifies that expected files exist and match. An extra injected file under `.gc/system/packs/core/` (such as a malicious order or script) will NOT be detected by the doctor check, but because the directory is included in config, that injected file is loadable and active!
* **Impact:** The escape hatch "proven unable to influence" is fragile for the primary required system pack.
* **Required Change:** Tighten required pack integrity. The doctor check for `core`, `bd`, and `dolt` must walk the on-disk directory and fail if any unexpected file is present. Prune-on-materialize and check-on-doctor must be in sync. Limit "preserve/ignore" semantics strictly to retired directories (`maintenance`/`gastown`).

---

### [Major] `PublicRepository` constant and `normalizeRepository` become dead-code registry clutter

* **Evidence:** `PublicRepository` constant (`internal/builtinpacks/registry.go:22`) and `normalizeRepository` branch (`registry.go:327-338`).
* **The Problem:** After the public-subpath and synthetic aliases are removed for gastown and maintenance, `publicSubpathForPack` will return false for everything. This makes `PublicRepository` and the `normalizeRepository` normalization branch dead code.
* **Required Change:** The design should decide whether to delete `PublicRepository` and the `normalizeRepository` branch, or add a negative registry test verifying that no remaining bundled pack exposes a public synthetic subpath (preventing accidental reintroduction of retired aliases).

---

## Minor Findings

1. **Cold-cache-boot latency spike on upgrade:**
   When users upgrade, `ValidateSyntheticRepo` will reject every existing synthetic cache because `SyntheticContentHash()` changes (the bundled set lost gastown and maintenance). `ensureBundledRepoInCacheLocked` will then re-materialize from scratch. For all three remaining bundled packs (core, bd, dolt), this is a cold re-materialization on first run. The design should acknowledge this one-time startup latency spike.

2. **SyntheticCacheNamespace bump:**
   The design should explicitly decide whether `SyntheticCacheNamespace` stays at `bundled-synthetic-v1` or bumps to `bundled-synthetic-v2`. Bumping is a defense-in-depth practice to prevent collision with stale directories.

---

## Cross-File Consistency Issues

* **`builtinPacks` package-level var vs. `All()` consistency:**
  `cmd/gc/embed_builtin_packs.go:41` defines `var builtinPacks = builtinpacks.All()`. This is initialized once at package load. If `All()` changes (returns 3 packs instead of 5), `builtinPacks` automatically reflects the change. But `TestBuiltinPacksUseCanonicalRegistry` (`embed_builtin_packs_test.go:317`) asserts that the two lists match by name only. If a pack is renamed (e.g., Core's subpath changes from `internal/bootstrap/packs/core` to `internal/packs/core`), the name-only comparison passes but subpath-dependent behavior may break. The test should also compare subpaths.

---

## Missing Evidence

1. **RepoCacheKey migration strategy:** how to handle existing namespaced cache entries for `PublicGastownPackSource` when the namespace prefix is removed. No test, no migration path, no doctor diagnostic.
2. **Packman/config call-site parity tests:** the five `IsSource` branch sites need explicit post-migration behavior tests, not just registry-level negative tests.
3. **`TestSyncLockUsesBundledFallbackForPublicGastownWhenRemoteUnavailable` migration plan:** this test will fail after migration and has no replacement specified.
4. **`TestResolveImportPackRefAcceptsPublicGastownSyntheticCache` migration plan:** this test will fail after migration and has no replacement specified.
5. **`loadBaselinePrompt` stale prompt template discovery:** no test proving stale Gastown prompt templates under `.gc/system/packs/gastown/agents/` are not used after migration.
6. **`packContainsEmbeddedManifest` unexpected-file gap:** no test proving the doctor check path detects (or explicitly accepts the risk of) extra files in required pack directories.

---

## Required Changes

1. **Address the `loadBaselinePrompt` stale prompt discovery risk:** Either filter the glob results against `requiredBuiltinPackNames` or restrict its scan to directories listed in resolved/active includes.
2. **Add a `RepoCacheKey` migration/upgrade strategy:** Choose, specify, and test one of: (a) backward-compatible lookup, (b) one-time migration, or (c) explicit doctor diagnostic requiring online reinstall.
3. **Enumerate and migrate all packman/config tests that assert synthetic acceptance for public Gastown:** Delete, invert, or rewrite: `TestSyncLockUsesBundledFallbackForPublicGastownWhenRemoteUnavailable` and `TestResolveImportPackRefAcceptsPublicGastownSyntheticCache`.
4. **Tighten `packContainsEmbeddedManifest` to reject unexpected files for required packs:** Extend the doctor check to walk the on-disk directory and reject files not in the embedded manifest for `core`, `bd`, and `dolt`.
5. **Decide whether to delete or retain dead public-synthetic registry plumbing:** Update `PublicRepository` and `normalizeRepository` accordingly.
6. **Update `TestBuiltinPacksUseCanonicalRegistry`:** Compare subpaths in addition to names.

---

## Questions

1. Is `loadBaselinePrompt` intended to ignore stale prompt templates left behind in the retired directories, or should those directories be fully pruned on startup?
2. How should upgraded cities handle existing `packs.lock` pins to public Gastown when running offline if their cache is keyed under the old synthetic namespace?
3. Should `PublicRepository` and the `normalizeRepository` branch for `gascity-packs` URLs be deleted outright after synthetic alias removal, or retained for potential future use?
