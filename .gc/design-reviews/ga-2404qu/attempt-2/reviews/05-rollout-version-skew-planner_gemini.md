# Yuki Hayashi — DeepSeek V4 Flash (Rollout Version Skew Review)

**Verdict:** block

**Persona focus:** Two-repo rollout sequencing, public pack pin integrity, intermediate state safety, rollback granularity. This review examines whether the design-after correctly handles the version-skew matrix, immutable pin validation, intermediate states between slices, and rollback paths. Iteration 2 re-examines the design-after (updated 2026-06-05T20:30Z) and the current codebase, with awareness of the iteration-1 findings from the other reviewer lanes.

---

## Top strengths

- The release compatibility matrix (old binary × old pack, old binary × new pack, new binary × old pack, new binary × new pack, rollback) is the right framework. It forces implementers to reason about every cross-repo state combination, and the "new binary × old pack" row correctly identifies that doctor must report actionable version-skew diagnostics rather than silently mutating content.
- Landing `gascity-packs` first (slice 1) and updating `PublicGastownPackVersion` only after the public commit is available (slice 6) is the correct sequencing. This avoids the "Gas City points at an unavailable or untested public Gastown commit" risk signal.
- The stale directory preservation policy (`.gc/system/packs/maintenance` and `.gc/system/packs/gastown` are ignored but never deleted) is conservative and correct.
- The preflight requirement before doctor mutations is the right contract to prevent partial-state corruption during version skew.

---

## Critical risks

### [Blocker] The "new binary × old pack" row has no implementation mechanism — doctor cannot produce the version-skew diagnostic the design promises

The compatibility matrix says:

> new binary | old pack | Fresh Gastown init pins the new pack, but existing locks to old public commits continue to load **or** doctor reports an actionable version-skew diagnostic without mutating custom content.

This is a disjunction: either old packs load cleanly, or doctor produces a diagnostic. The design provides no specification for:

1. **How the version-skew diagnostic is detected.** The import-state doctor (`cmd/gc/import_state_doctor_check.go:legacyPublicPackForSource`) only matches legacy local paths (`.gc/system/packs/gastown`, `examples/gastown/packs/gastown`). It does not match public Gastown imports at an old commit. A city with `source = "https://github.com/gastownhall/gascity-packs.git//gastown"` and `version = "sha:abc123"` (where `abc123` is an old pinned commit) will pass the `legacyPublicPackForSource` check because the source is remote — `isRemoteImportSource` returns true and the function bails out at line 218.

2. **How "actionable" is defined.** The design says "actionable version-skew diagnostic" but does not specify what the diagnostic contains or what action the operator should take. The current doctor output model is a `CheckResult` with `Name`, `Status`, `Message`, `FixHint`, and `Details`. What should `FixHint` say for version skew? "Update your Gastown import to the latest pinned version"? But that might not be safe if the operator has custom Gastown overlays.

3. **What happens when the old public pack is incompatible.** After slice 5 (Maintenance folding), Core drops Maintenance-specific behavior. A city locked to an old public Gastown commit that still imports `../maintenance` will have that import resolve to nothing (the in-tree Maintenance pack is gone). The Gastown pack's own `pack.toml` `imports` may reference `../maintenance` which resolves to a nonexistent local path in the new binary. This is a pack-composition failure, not a version-skew diagnostic — and nothing in the design specifies how this is handled.

The design must specify: (a) a new doctor check or import-state doctor extension that detects public Gastown imports pinned to commits older than `PublicGastownPackVersion`, (b) the exact diagnostic content and fix hint for the version-skew case, (c) what happens when the old public pack's `imports` reference a pack that no longer exists in the new binary, and (d) a test proving the "new binary × old locked pack" path is exercised.

### [Blocker] `PublicGastownPackVersion` is a hardcoded SHA constant with no materialization-time verification — the design requires immutable content verification but provides no mechanism

The design says:

> Is `PublicGastownPackVersion` pinned to immutable content with materialization-time verification rather than a mutable branch or tag?

The current implementation in `internal/config/public_packs.go:11` is:

```go
PublicGastownPackVersion = "sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b"
```

This is a SHA-pinned constant, which satisfies the immutability requirement at the source level. However:

1. **No materialization-time verification exists.** `MaterializeSyntheticRepo` (`internal/builtinpacks/registry.go:148`) writes the embedded pack tree to a cache directory and validates it against a manifest derived from the current binary's embedded FS. But this only validates bundled packs (core, maintenance, gastown, bd, dolt). The public Gastown pack is fetched via `packman.EnsureRepoInCache` → `EnsureRepoInCache` → `ensureRepoInCacheLocked`, which clones from GitHub and checks out the pinned commit. There is no content hash verification step — a MITM or corrupted cache could serve different content at the same commit SHA, and the only check is `validateCachedPackRoot` which verifies `pack.toml` existence, not content integrity.

2. **The synthetic cache validation does not cover public packs.** `builtinpacks.ValidateSyntheticRepo` (the validation path for bundled packs) validates file content and mode byte-by-byte against the binary's embedded manifest. But public packs go through the git checkout path, which only validates that `pack.toml` exists at the root. A public pack cache entry that passes `validateCachedPackRoot` could have arbitrary content differences from the intended commit.

3. **The pin update path has no verification gate.** Slice 6 says "update `PublicGastownPackVersion`." But the design specifies no mechanism to verify that the new pin's content is the content that was actually tested in `gascity-packs`. The `TestPinnedPublicGastownBehavior` packcompat test is proposed but does not exist yet, and the design doesn't say what it should validate beyond "behavior inventory completeness."

The design must specify: (a) a content verification step when materializing the public Gastown pack that hashes the pack tree and compares it against an expected content hash stored alongside the commit SHA, or (b) a rationale for why the git commit SHA alone is sufficient integrity verification, considering the `MaterializeSyntheticRepo` precedent validates content byte-by-byte, or (c) at minimum, a `TestPinnedPublicGastownBehavior` specification that includes content hash verification, not just behavioral equivalence.

### [Blocker] Slice 5 (Maintenance folding) creates an intermediate state where neither Maintenance nor Core provides Maintenance-specific behavior

The rollout specifies slices in order 1→2→3→4→5→6→7. Slice 5 says:

> Move Core-owned Maintenance assets into Core, move Gastown-owned Maintenance assets to public Gastown, and apply the runtime retirement table.

This means during slice 5 deployment:

1. The Gas City binary removes Maintenance from the embedded pack set (`builtinpacks.All()` drops `maintenance`) and `requiredBuiltinPackNames` stops including `"maintenance"`.
2. Core gains the moved Maintenance assets.
3. Public Gastown gains the Gastown-specific Maintenance assets.

But consider the intermediate state during a rolling upgrade or during the `PublicGastownPackVersion` update lag:

- **New binary, old Gastown lock:** The old public Gastown pack imports `../maintenance`. After slice 5, the binary no longer embeds or materializes a `maintenance` pack. The Gastown pack's import of `../maintenance` resolves to `.gc/system/packs/maintenance` — which is now a stale directory that is ignored by config loading. Gastown's Maintenance-dependent formulas, orders, and scripts will fail at runtime because their pack imports cannot be resolved.

- **The design says stale directories are "ignored, diagnosed, and preserved, not deleted."** But "ignored" means the config loader skips them, so Gastown's `../maintenance` import resolves to nothing. The design needs to specify what happens when a pack import cannot be resolved — is this a load error, a diagnostic warning, or a silent degradation?

- **The rollback path after slice 5 is unclear.** If an operator rolls back to a pre-slice-5 binary after Gastown's `pack.toml` has been updated (by the new binary's doctor) to remove the `../maintenance` import, the old Gastown pack that the rollback binary locks will still expect Maintenance. But the design says doctor removes Maintenance imports, so a rollback binary would see a Gastown pack that no longer imports Maintenance. The old binary expects Maintenance to exist as a system pack and will try to materialize it.

The design must specify: (a) the exact resolution behavior when a pack import cannot be satisfied (load error vs diagnostic vs silent), (b) whether Gastown's `pack.toml` removes the `../maintenance` import in the same commit as the new `PublicGastownPackVersion`, and if so, how the "new binary × old Gastown lock" case handles the missing import, and (c) whether rolling back from slice 5 requires a separate Gastown pin rollback, or whether the old binary is expected to tolerate the absence of Maintenance gracefully.

### [Blocker] The `requiredBuiltinPackNames` function hardcodes `"maintenance"` as required — removing it in slice 5 breaks `builtinPackIncludes` for every intermediate state

`cmd/gc/embed_builtin_packs.go:236-250`:

```go
func requiredBuiltinPackNames(cityPath string) []string {
    required := []string{"core", "maintenance"}
    // ...
}
```

`builtinPackIncludes` (line 262) includes exactly the packs in `requiredBuiltinPackNames`. After slice 5 removes `"maintenance"` from this list, `builtinPackIncludes` will only include Core (plus bd/dolt). This means:

1. **Any code path that calls `config.LoadWithIncludes` through `builtinPackIncludesForConfigLoad` will no longer auto-include the Maintenance pack.** During the intermediate state between slice 5 deployment and `PublicGastownPackVersion` update, existing cities that rely on implicit Maintenance inclusion will lose Maintenance formulas, orders, and scripts from their resolved config.

2. **The design says "Core is required" and "Maintenance is retired,"** but the retirement is supposed to be gradual: Core gains Maintenance behavior, and Gastown gains the rest. If the binary stops including Maintenance before the public Gastown pack has been updated to carry the Gastown-specific Maintenance behavior, there is a coverage gap.

3. **The `MaterializeBuiltinPacks` function** (line 46-77) materializes all `builtinPacks` (which includes maintenance) to `.gc/system/packs/`. If maintenance is removed from the pack set, `.gc/system/packs/maintenance` will not be refreshed. The stale directory preservation policy says it's "ignored" — but Gastown's `../maintenance` import resolves relative to the Gastown pack's location, which is in `.gc/system/packs/gastown/`. The relative path `../maintenance` resolves to `.gc/system/packs/maintenance/`. If this directory is stale (contains old content but is not refreshed), the config loader ignores it, but the Gastown pack's import resolution will fail because the import source is now stale and unmaintained.

The design must specify the exact sequencing of: (a) when `"maintenance"` is removed from `requiredBuiltinPackNames`, (b) when Maintenance assets land in Core and public Gastown, (c) when the public Gastown pack's `pack.toml` removes the `../maintenance` import, and (d) how `builtinPackIncludes` and import resolution handle the transition period. The current seven-slice rollout does not make this sequencing explicit.

---

## Major risks

### [Major] `defaultWave1PublicPackImports` is a static map with no network verification — air-gapped hosts will pass doctor preflight but fail at lock/install

`cmd/gc/import_state_doctor_check.go:72-91` shows `defaultWave1PublicPackImports` is a static map that returns `PublicGastownPackSource` and `PublicGastownPackVersion` with no reachability check. The design's preflight requirement says "verify the public Gastown source and immutable version are reachable, installable, and lockable." But the current code only does a static lookup.

An air-gapped host would:
1. Pass `legacyPublicPackImportDetails` (it has legacy imports).
2. Pass `defaultWave1PublicPackImports` (static map, no network).
3. Start `Fix()`, which rewrites `pack.toml` and `city.toml`.
4. Fail at `syncImports` → `packman.Install` when trying to clone the public repo.

At this point, manifests are already mutated with no rollback (as the 03-review also identifies). But even with the proposed preflight fix, an air-gapped host cannot validate reachability without network access. The design must specify: either (a) preflight defers all manifest writes until after lock/install succeeds (option b from the 03-review), or (b) preflight explicitly checks for air-gapped conditions and refuses the auto-fix with manual guidance, or (c) the bundled synthetic cache serves as a fallback for the public Gastown pack so that air-gapped hosts can resolve the import without network access.

Option (c) is architecturally interesting because `builtinpacks.All()` already embeds the Gastown pack and `MaterializeSyntheticRepo` can create a local cache for it. After slice 6 removes Gastown from the embedded pack set, this fallback disappears. The design must specify whether air-gapped Gastown resolution is supported after the migration and, if so, how.

### [Major] The rollback compatibility matrix does not address `pack.toml` import order preservation

The design says:

> Existing order skip lists containing moved Core order names should continue to work when names are preserved. If any order is renamed, provide aliases or a migration test.

But `replaceImportOrderWithTargets` (`cmd/gc/import_state_doctor_check.go:375-399`) rewrites `DefaultRigImportOrder` by replacing legacy names with targets. This reordering is not order-preserving for existing entries that are not being rewritten:

```go
for _, name := range order {
    if target, ok := targetByLegacy[name]; ok {
        if target != "" && !seenTarget[target] {
            out = append(out, target)
            seenTarget[target] = true
        }
        continue
    }
    seenTarget[name] = true
    out = append(out, name)
}
```

If the existing order is `["core", "gastown", "maintenance"]` and targets are `maintenance→remove`, the result is `["core", "gastown"]`. This is correct. But if the existing order is `["gastown", "core", "maintenance"]`, the result is `["gastown", "core"]` — the relative order of non-rewritten entries is preserved. However, if the rewrite adds `gastown` (replacing a legacy source), it appends it at the end: `["core", "gastown"]` regardless of the original position. This could change import priority order, which affects overlay resolution.

The design should specify: (a) whether `DefaultRigImportOrder` order matters for overlay resolution, and if so, (b) whether doctor fix should preserve the original position of replaced entries rather than appending at the end.

### [Major] The synthetic cache namespace does not distinguish between old and new Gastown pack content

`internal/builtinpacks/registry.go` uses `SyntheticCacheNamespace = "bundled-synthetic-v1"` as a prefix for cache keys. After slice 6 removes Gastown from the embedded pack set, any existing synthetic cache entries for Gastown will have keys like `bundled-synthetic-v1\x00https://github.com/gastownhall/gascity.git//examples/gastown/packs/gastown\x00<commit>`.

But after slice 1 (public Gastown lands on a branch), `PublicGastownPackSource` points to `https://github.com/gastownhall/gascity-packs.git//gastown`. The cache key for this is computed differently: it uses `config.RepoCacheKey` with the normalized source URL, not `SyntheticCacheNamespace`.

This means after the migration:
- Old cache entries keyed under the in-tree Gastown source URL will be orphaned.
- New entries keyed under the public Gastown source URL will be fresh clones.
- There is no migration or cleanup for the old synthetic cache entries.

The design's slice 6 mentions "retire public synthetic aliases" but does not specify what happens to existing cache entries under the old key space. If an operator rolls back to a pre-slice-6 binary, the old binary will use the old synthetic cache key and may find stale content.

The design must specify whether old synthetic cache entries are cleaned up during migration or simply left to expire, and whether a rollback binary can safely reuse them.

### [Major] `GastownCity()` pins `PublicGastownPackVersion` in both `Imports` and `DefaultRigImports` — after slice 6, this creates a deprecation path that is not addressed

`internal/config/config.go:3712-3735` shows `GastownCity()` setting both `Imports["gastown"]` and `DefaultRigImports["gastown"]` to `PublicGastownPackSource`/`PublicGastownPackVersion`. After slice 6 updates the pin, all new `gc init --template gastown` cities will use the new pin. But:

1. **Existing cities with the old pin will not be automatically updated.** The import-state doctor only rewrites local Gastown imports (`.gc/system/packs/gastown`, `examples/gastown/packs/gastown`), not public imports at an old commit. There is no mechanism to nudge operators to update their lockfile to the new commit.

2. **The design says "existing locks to old public commits continue to load"** but does not address what happens when the old public commit's content becomes incompatible with the new binary (see the blocker about Gastown's `../maintenance` import).

3. **The `PublicGastownPackVersion` constant is the only version pin, and it's compiled into the binary.** This means updating the pin requires a new binary release. If a critical Gastown pack fix is needed between binary releases, operators cannot update their lockfile to the fixed commit without a new `gc` binary. The design should specify whether `PublicGastownPackVersion` should ever be overridable via config, and if not, what the release cadence expectation is.

---

## Minor risks

### [Minor] The `publicSubpathForPack` function only maps `gastown` and `maintenance` — after slice 6, `maintenance` has no public mapping

`internal/builtinpacks/registry.go:publicSubpathForPack` maps `"gastown"` and `"maintenance"` to public subpaths. After the migration, Maintenance is retired and should not have a public subpath mapping. The function should return `("", false)` for `"maintenance"` after slice 6, or be removed entirely if Gastown is also removed from the embedded set.

### [Minor] The `legacyPublicPackForSource` function matches `.gc/system/packs/gastown` as a legacy path — but after the migration, this path is a stale ignored directory, not a valid import source

The function at `cmd/gc/import_state_doctor_check.go:216-243` matches sources ending in `.gc/system/packs/gastown`. After the migration, this directory is stale and ignored by config loading. But if an operator manually edits their `pack.toml` to add `source = ".gc/system/packs/gastown"`, the doctor will still try to rewrite it as a public import. This is probably the right behavior, but the design should explicitly confirm it.

### [Minor] The design does not specify whether `gc init --template gastown` should pin to `PublicGastownPackVersion` or to the latest registry-available commit

The current `GastownCity()` uses `PublicGastownPackVersion` which is a hardcoded SHA. The design says "Fresh Gastown init pins the new pack" but does not clarify whether the pin should always be the exact SHA in the binary or the latest available commit from the registry. If it's always the binary SHA, operators who `gc init` with a month-old binary get a month-old Gastown pack, which may lack recent fixes. The design should state whether this is intentional or whether `gc init` should resolve to the latest commit.

---

## Missing evidence

1. **No test for the "new binary × old locked public pack" state.** The design specifies this as a compatibility matrix row but provides no test specification for it. There should be a test that creates a city with an old `PublicGastownPackVersion` lock, runs the new binary, and verifies either successful loading or actionable diagnostics.

2. **No specification for Gastown pack `imports` resolution when the imported pack does not exist.** After Maintenance is retired, Gastown's `../maintenance` import (if still present in the locked commit) cannot be resolved. The design must specify whether this is a load error, a warning, or silent degradation, and what the observable behavior is.

3. **No specification for the air-gapped host experience after the migration.** The current system resolves public Gastown from the bundled synthetic cache. After Gastown is removed from the embedded set, air-gapped hosts have no local cache for the public pack. The design must specify whether air-gapped operation is supported and how.

4. **No rollback test specification.** The compatibility matrix has a "rollback from new to old" row but no test that exercises it. A test should verify that a city initialized with the new binary can be downgraded to the old binary and still function (or produce actionable diagnostics).

5. **No specification for what happens to `builtinPackIncludes` during the transition period between removing Maintenance from the embedded set and the public Gastown pack carrying Maintenance-specific behavior.** The design says slice 5 folds Maintenance into Core and Gastown, but the `requiredBuiltinPackNames` change must be atomic with the Gastown pack update, and no test proves this atomicity.
