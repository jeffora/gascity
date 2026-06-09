# Marcus Driscoll — DeepSeek V4 Flash Review (Iteration 3)

**Persona:** Pack Registry and Cache Tester
**Mandate:** Builtin registry identity, synthetic cache pruning, system pack materialization, provider-dependent pack continuity
**Verdict:** block

---

## Top Strengths

1. **The review-gated migration invariants are well-structured.** The diff added a concrete invariant checklist (test-green, Core provenance, behavior-first, doctor safety, stale-directory preservation, role neutrality) that each slice must satisfy before the next starts. This is the right mechanism to prevent slice-skipping, and it directly addresses the most common failure mode in multi-step migrations.

2. **Registry identity is now precise and testable.** `All()` → `{core=internal/packs/core, bd=examples/bd, dolt=examples/dolt}` with no Gastown or Maintenance entries, plus the removal of the `publicSubpathForPack` case for those names, is a well-scoped target. Existing `TestAllAndSourceAreDeterministic` provides a named pin point.

3. **The release compatibility matrix is valuable.** The five-cell old/new binary × old/new public pack matrix with explicit rollback handling is the right shape for a two-repo migration. It forces thinking about each combination before landing code.

---

## Blocker Findings

### [Blocker] `RepoCacheKey` namespace flip has no migration, fallback, or diagnostic strategy

`RepoCacheKey` (pack_include.go:298–306) prefixes bundled sources with `bundled-synthetic-v1\x00`. After migration, `IsSource(PublicGastownPackSource)` returns false, so new installs compute the un-namespaced key. An existing city whose `packs.lock` references the old namespaced cache path will:
- Fail to find the cache at the new un-namespaced path during `resolveLockedRemoteImport`
- Appear "not installed" despite having valid cached content at the old path
- Require `gc import install` to re-fetch and re-cache at the new key

The design says "existing locks to old public commits continue to load or doctor reports an actionable version-skew diagnostic" (release matrix, row 4), but it does not specify which outcome wins, how doctor detects this specific key-mismatch, or what the user-facing error looks like. The Gemini and Claude reviews both flagged this; I agree it is blocking because it affects every existing Gastown-using city on upgrade.

**Evidence:** `RepoCacheKey` at pack_include.go:298–306; `EnsureRepoInCache` dispatch at cache.go:63; lock-resolution at pack_include.go:213.

### [Blocker] `TestSyncLockUsesBundledFallbackForPublicGastownWhenRemoteUnavailable` has no specified post-migration behavior

This test (check_test.go:49–91) verifies that `SyncLock` + `InstallLocked` + `CheckInstalled` works for `PublicGastownPackSource` when the network is unavailable, using the bundled synthetic fallback. After migration, `IsSource(PublicGastownPackSource)` is false, so `EnsureRepoInCache` takes the ordinary git-clone path (cache.go:77), which requires network. Offline Gastown init will fail.

The design must state one of:
- Offline Gastown init is no longer supported and the error message is clear.
- A new offline mechanism exists (pre-materialized cache, separate bundled Gastown binary, etc.).
- The synthetic fallback is preserved for public Gastown specifically until a deprecation period.

The current test will break on migration day, and the design does not enumerate it or specify the replacement behavior.

**Evidence:** check_test.go:49–91; cache.go:63–77.

### [Blocker] `validateInstalledRemoteCache` and `ReadCachedPackImports` accept synthetic caches for sources that will no longer be `IsSource`

`validateInstalledRemoteCache` (pack_include.go:213–230) dispatches on `IsSource(source)` to decide between synthetic validation and git-checkout validation. `ReadCachedPackImports` (install.go:49) does the same. After migration, public Gastown sources are no longer `IsSource`, but existing synthetic caches at the old namespaced key still exist on disk. The code will attempt git-checkout validation on a synthetic-cache-shaped directory, which will fail with a confusing error rather than a clear migration diagnostic.

The design needs call-site-specific behavior specifications for all five `IsSource` branches (pack_include.go:213, :304; cache.go:63; check.go:231; install.go:49), not just the registry-level identity change.

**Evidence:** The five `IsSource` call sites enumerated in the Claude review and confirmed in source.

---

## Major Findings

### [Major] `dog` pool-name contract ambiguity between Core ownership and dolt dependency

`examples/dolt/pack.toml:6` states: "Dog-backed formulas and orders rely on the city's maintenance pack." After migration, maintenance moves into Core. The design says tests must prove "a Core-only city can still load, run normal SDK infrastructure, and evaluate non-agent controller operations when the maintenance agent is renamed or omitted" (design-after.md:216–218). But `examples/dolt/orders/mol-dog-stale-db.toml` hardcodes `pool = "dog"`.

This creates a contradiction: if `dog` is "freely renameable" (ZFC/role-neutral), then dolt's `pool = "dog"` breaks on rename. If `dog` is a stable contract that provider packs can bind to, then Core cannot rename it. The design needs to declare which it is and adjust either the Core role-neutrality test or the dolt pack contract. The Claude review identified this; I'm confirming it as a major design-level inconsistency rather than a test gap.

### [Major] `SyntheticContentHash` is global over all packs in `All()`

`SyntheticContentHash` (registry.go:252–274) hashes all packs returned by `All()` into a single marker. Removing `maintenance` and `gastown` from `All()` changes this hash, invalidating every existing synthetic cache including `core`, `bd`, and `dolt`. The design says "materialized bytes and provenance remain unchanged except for expected manifest metadata" (design-after.md:609–614), but the content hash changes, so existing caches self-invalidate and require re-materialization.

This is actually safe — `ValidateSyntheticRepo` will reject stale caches and `MaterializeSyntheticPacks` will re-materialize them — but the design must explicitly acknowledge this one-time re-materialization event and add a test proving self-heal, not silent breakage. The Gemini review identified this; I'm confirming and adding that the design should state whether the re-materialization happens on next `gc` start or requires explicit user action.

### [Major] `MaterializeBuiltinPacks` prune semantics do not cover unexpected files in required packs

`MaterializeBuiltinPacks` (embed_builtin_packs.go:60–70) iterates `requiredBuiltinPackNames` and writes/refreshes only those packs. `pruneStaleGeneratedPackFiles` (embed_builtin_packs.go:480–538) prunes generated-only files that are no longer in the desired set. But `packContainsEmbeddedManifest` (embed_builtin_packs.go:178–195) only checks that embedded files are present and content-correct — it does not walk the directory to detect unexpected files.

For required packs (`core`, `bd`, `dolt`) included in normal config resolution, an injected file would survive between materializations. The design's "Core integrity" test requirement says "unexpected files either fail full file-set integrity validation or are proven unable to influence loaded formulas, orders, scripts, overlays, and prompts" — but does not commit to which. For required packs, unexpected files should be rejected/pruned. The current `validatePackFiles` in `registry.go` already does unexpected-file rejection for synthetic caches; this should be applied to on-disk system packs too.

### [Major] Stale retired `.gc/system/packs/gastown/` prompt templates can still affect prompt loading

The design specifies that stale `{maintenance,gastown}` directories are "ignored by config loading after the migration" (design-after.md:680–682). This is true for `builtinPackIncludes` and `config.LoadWithIncludes`, but prompt loading may glob `agents/<role>/prompt.template.md` under all `.gc/system/packs/*/` directories. The design adds "filter the glob results against `requiredBuiltinPackNames`" (Gemini review finding #4), but does not commit to this in the design text itself.

If prompt loading does not filter by required pack names, a stale `.gc/system/packs/gastown/agents/mayor/prompt.template.md` would shadow Core's mayor prompt — a far more dangerous consequence than just stale config includes.

### [Major] `builtinPackIncludes` comment still says "Core and maintenance"

The current comment (embed_builtin_packs.go:263) says: "Core and maintenance are always included." After migration, this becomes "Core is always included (with provider-dependent bd/dolt)." The design's test requirement names updating `TestBuiltinPackIncludes_*`, but the comment, doc string, and any related error messages also need updating as a tracked deliverable, not just a test-byproduct.

### [Major] `normalizeRepository` still maps `gascity-packs` URLs to `PublicRepository`

`normalizeRepository` (registry.go:447–458) has a special case mapping `https://github.com/gastownhall/gascity-packs` to `PublicRepository`. After the public Gastown synthetic alias is removed, this branch serves no purpose for synthetic cache recognition — but it still normalizes `gascity-packs` clone URLs. If this normalization is needed for ordinary remote pack resolution, it should stay. If it only existed to support synthetic recognition, it should be removed. The design must state which, and the Codex review's question about this remains unanswered.

---

## Minor Findings

### [Minor] The old `internal/bootstrap/packs/core` source should be a named negative test case

The design says old Core sources should be "rejected or covered by explicit migration diagnostics" (design-before.md:368), but the updated design-after.md just says the old path is "rejected" without specifying whether `IsSource`, `NameForSource`, lock generation, or doctor produces the rejection. The negative test should explicitly verify that `IsSource("https://github.com/gastownhall/gascity.git//internal/bootstrap/packs/core")` returns false after migration.

### [Minor] `PublicRepository` constant and `PublicGastownPackSource`/`PublicGastownPackVersion` in `internal/config/public_packs.go`

After migration, `PublicGastownPackSource` still points to `https://github.com/gastownhall/gascity-packs.git//gastown` and `PublicGastownPackVersion` pins to a specific commit. These are used by `GastownCityWithProviders` (config.go:3728–3735) for fresh Gastown init. The design must state whether these remain as the ordinary remote import source (not synthetic) and whether `PublicRepository` in registry.go is retained for ordinary URL normalization or deleted.

### [Minor] `requiredBuiltinPackNames` starts with `[]string{"core", "maintenance"}` — the migration must change this to `[]string{"core"}`

This is obvious but not explicitly called out as a code change target in the design. The entire function body changes, which affects every test that mocks or overrides `requiredBuiltinPackNames`.

---

## Cross-Document Consistency Issues

### Consistency between design-after.md and existing reviews

1. **Cache-key strategy:** The design-after.md (line 672, registry/cache slice) says "retire public synthetic aliases, update `PublicGastownPackVersion`, and verify stale cache rejection." But it does not address the `RepoCacheKey` namespace flip or specify the migration/lookup/diagnostic strategy. The Gemini review's blocker on cache orphans is not reflected in the updated design.

2. **Offline behavior:** The design-after.md does not specify offline Gastown init/install behavior after synthetic fallback removal. The Gemini and Codex reviews both flagged the `SyncLockUsesBundledFallbackForPublicGastownWhenRemoteUnavailable` test; the design needs a section specifying the new offline behavior.

3. **`IsSource` call-site enumeration:** The design-after.md test requirements list "Source recognition for `internal/packs/core`; old `internal/bootstrap/packs/core` sources should be rejected" but do not enumerate the five production `IsSource` call sites and specify behavior at each. The Claude review listed all five; the design should incorporate that enumeration.

4. **Provider-continuity proof scope:** The design-after.md says "provider pack continuity: `bd` and `dolt` materialized bytes and provenance remain unchanged" but does not address the `dog` pool-name contract issue raised in the Claude review. Byte-level continuity is necessary but not sufficient; the `dog` pool resolution from Core must also be proven.

---

## Edge Cases Other Reviewers May Accept Too Quickly

### Edge case: concurrent binary upgrade during active import install

If one `gc` process is mid-import-install while another `gc` process upgrades the binary (changing `All()` and `SyntheticContentHash`), the `RepoCacheKey` for the same source+commit changes mid-operation. The write-lock in `EnsureRepoInCache` protects the cache directory, but the key change means the new process writes to a different directory while the old process may still be reading from the old one. This is likely safe (old process continues with old binary behavior; new process starts fresh), but the design should acknowledge concurrent-upgrade scenarios in the cache strategy.

### Edge case: `publicSubpathForPack` removal affects `syntheticPackLayouts`

`syntheticPackLayouts` (registry.go:103–118) builds layouts by iterating `All()` and then adding public aliases via `publicSubpathForPack`. Removing the `gastown`/`maintenance` cases from `publicSubpathForPack` means public Gastown is no longer in `syntheticPackLayouts`, which is correct — but it also means `NameForSource("https://github.com/gastownhall/gascity-packs.git//gastown")` returns false. This is the intended behavior, but it means any code that calls `NameForSource` on a public Gastown source will get a different result. The design should verify that no code path (including doctor check) relies on `NameForSource` recognizing public Gastown.

### Edge case: `examples/gastown/packs/gastown/pack.toml` imports `../maintenance`

The in-repo `examples/gastown/packs/gastown/pack.toml` imports `../maintenance`. After migration, this local import path is severed because `examples/gastown/packs/maintenance` moves its assets into Core and Gastown. The design addresses this for the public pack but not for the in-repo test tree. The source deletion slice (slice 7) should verify that the local import chain is also updated or that the local example tree is removed before the public Gastown pin change.

---

## Required Changes

1. **Specify the `RepoCacheKey` upgrade contract.** Choose one: (a) one-time migration that relocates namespaced cache entries, (b) backward-compatible dual-read that checks both keys, or (c) explicit doctor diagnostic + mandatory `gc import install`. Add a test for the chosen strategy covering old-key cache, new-key lookup, and offline failure behavior.

2. **Enumerate all five `IsSource` call sites** (pack_include.go:213, :304; cache.go:63; check.go:231; install.go:49) and specify the post-migration behavior at each. For public Gastown, each site must take the non-bundled path; specify the exact error or behavior for offline scenarios.

3. **Resolve the `dog` pool-name contract.** Declare `dog` as a stable pool-name contract that provider packs may bind to, or make dolt's `pool` configurable, or rename the concept out of Core entirely. Update `examples/dolt/pack.toml:6` accordingly.

4. **State that `SyntheticContentHash` changes on migration and add a self-heal re-materialization test.** The test should prove that after removing maintenance/gastown from `All()`, a city with an old synthetic cache re-materializes successfully on next `gc` start without data loss.

5. **Commit to unexpected-file rejection/pruning for required system packs** (`core`, `bd`, `dolt`) and reserve preserve/ignore semantics only for retired, non-included directories. Align `packContainsEmbeddedManifest` with `validatePackFiles` for required packs.

6. **Add a stale prompt-template diagnostic or filter.** Either filter the prompt glob to `requiredBuiltinPackNames` or add a doctor diagnostic for stale prompt templates under retired pack directories.

7. **Enumerate the test migration list for public Gastown synthetic tests.** Specifically: `TestSyncLockUsesBundledFallbackForPublicGastownWhenRemoteUnavailable`, `TestResolveImportPackRefAcceptsPublicGastownSyntheticCache`, and any test in `check_test.go` or `bundled_import_test.go` that asserts `IsSource(PublicGastownPackSource)` is true. For each, specify the new behavior and replacement test.

8. **Decide the fate of `PublicRepository`, `normalizeRepository`'s `gascity-packs` branch, and `publicSubpathForPack`.** Delete dead code or document its retained purpose.

9. **Add the concurrent-binary-upgrade edge case to the cache strategy** or acknowledge it as safe-by-construction with a brief rationale.

---

## Questions

1. For an existing city whose `packs.lock` references a namespaced public Gastown cache entry, what is the exact required outcome on a new-binary run offline? Hard fetch error, doctor version-skew diagnostic, or one-time compatibility cache?

2. Should offline `gc init --template gastown` fail with a clear "could not fetch public Gastown pack" error after migration, or is some offline-capable mechanism preserved?

3. Is `dog` a stable pool-name contract that provider packs may bind to, or must dolt's `pool = "dog"` become configurable so Core can rename its maintenance agent?

4. Should `PublicRepository` and the `gascity-packs` normalization branch in `normalizeRepository` be deleted after public synthetic alias removal, or retained for ordinary remote pack resolution?

5. After `publicSubpathForPack` returns false for all bundled packs, should a registry test assert that no bundled pack exposes a public synthetic subpath, preventing silent reintroduction of a retired alias?
