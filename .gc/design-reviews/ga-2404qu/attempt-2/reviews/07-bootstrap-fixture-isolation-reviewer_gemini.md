# DeepSeek V4 Flash — Bootstrap Fixture Isolation Review

**Lane:** 07-bootstrap-fixture-isolation-reviewer
**Scope:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden dependency discovery, cross-file consistency, pattern drift, architectural coherence.
**Verdict:** block

## Summary

The design identifies the correct migration target — remove the production `//go:embed packs/**` from `internal/bootstrap/bootstrap.go`, isolate test fixtures, and eliminate the dual-embed of Core — but leaves six structural gaps that will cause implementation drift, false-green tests, or production regressions. The most serious are: (1) the production `bootstrapAssets` variable has no specified post-removal default and the current code reads from it whenever `BootstrapPacks` is non-empty, meaning a nil or accidentally-empty replacement causes panics or silent dead paths; (2) the test fixture contract is so underspecified that implementers will likely copy production Core "to make tests realistic," re-introducing the coupling the migration removes; (3) the `bootstrapManagedImportNames` list in `internal/config/compose.go` hardcodes `"core"` and `"registry"` but the design never mentions this second `core` reference or its migration; (4) `requiredBuiltinPackNames` hardcodes `"maintenance"` but the design's required-pack transition plan is incomplete; (5) the `GC_BOOTSTRAP=skip` early return is not retired or contained, leaving a production-visible bypass that conflicts with the "no environment variable" stance; (6) `internal/hooks/hooks.go` directly imports `internal/bootstrap/packs/core` for `core.PackFS` — the migration must change this import, but the design does not specify how hooks survive the move without breaking provider overlay installation.

I also found eight cross-file consistency issues and edge cases that prior reviews partially address but leave underspecified.

---

## Blocking Findings

### B1. `bootstrapAssets` post-removal default is unspecified; nil causes panics

`internal/bootstrap/bootstrap.go:22-25` declares:

```go
var embeddedBootstrapPacks embed.FS
var bootstrapAssets fs.FS = embeddedBootstrapPacks
```

When `//go:embed packs/**` is removed, `embeddedBootstrapPacks` becomes an empty `embed.FS{}`. The design says "remove the production `//go:embed packs/**` dependency" (L490) but never specifies what `bootstrapAssets` becomes. Three outcomes are possible, all problematic:

- **nil**: Every `fs.Stat(nil, …)` or `fs.WalkDir(nil, …)` call in `collectAssetFiles`, `copyEmbeddedTree`, `bootstrapPackRevision` (L170, L194, L217, L220, L241, L254) panics.
- **empty embed.FS{}**: `fs.Stat(emptyFS, "packs/core")` returns `os.ErrNotExist`, producing a different error path than today. Code that sets `BootstrapPacks` in tests (collision_test.go:111,149,172; skills_test.go:207,345) will hit this error and fail with an opaque "reading embedded bootstrap pack" message instead of a clear "no fixture injected" error.
- **Left as-is with a "tiny compatibility embed"**: The design hedges at L491–493 with "If a tiny compatibility embed remains…" which contradicts the anti-regression guard at L494–497 that should reject any `AssetDir: "packs/core"` outside the new fixture tests.

The Claude and Codex reviews both flag this nil-FS hazard. I want to underscore a different angle: even an empty `embed.FS{}` is wrong because it silently satisfies the `fs.FS` interface while returning `ErrNotExist` for all reads. The design must specify an explicit **named empty-FS type** (e.g., `errFS` struct in `bootstrap.go`) that returns `fs.ErrNotExist` for all operations, so that production `BootstrapPacks` (always empty) never accidentally exercises the materialization path, and test-only overrides use `os.DirFS("testdata/packs")` or `fstest.MapFS` injected from `_test.go` files.

**Required:** Specify that `bootstrapAssets` defaults to an explicit empty-FS implementation that returns `fs.ErrNotExist` for all reads, not nil. Name the type. Add a production-guard assertion: `TestProductionBootstrapAssetsIsEmpty` walks `bootstrapAssets` and fails if any entry is reachable. Test-only overrides use `os.DirFS("testdata/packs")` or `fstest.MapFS`.

### B2. Test fixture contract is unspecified; drift is undetectable

The design says "a minimal synthetic fixture under `internal/bootstrap/testdata/packs/core`" (L483) and "fixture assets do not hash/copy production Core" (L616), but never specifies:

1. **What files the fixture contains.** Is it `pack.toml` only? `pack.toml` plus a trivial placeholder? The design must enumerate the allowed fixture file set.
2. **What `name` and `schema` the fixture `pack.toml` carries.** Current collision tests (`collision_test.go:111,149,172`) set `Entry{Name: "core", AssetDir: "packs/core"}` and assert that `core` collides with a user import named `core`. If the fixture uses `name = "test-core"`, these collision semantics change. If the fixture uses `name = "core"`, the fixture *pretends* to be production Core, inviting drift where a contributor copies real Core content "to make it realistic."
3. **What the fixture must never contain.** Production Core has `formulas/`, `orders/`, `skills/`, `overlay/`, and `assets/prompts/`. The fixture must explicitly forbid these paths. The Claude review says this implicitly; I want it as a named CI contract.
4. **What happens when production Core's `pack.toml` schema version changes.** Today production Core is `schema = 2`. If a future migration bumps it to `schema = 3`, the fixture stays at `schema = 2` and no test catches the drift. The fixture contract must state whether it tracks the production schema version or deliberately does not.

The Codex review's finding #1 aligns here. I go further: the fixture contract should be a **separate test function** (`TestBootstrapFixtureIsMinimal`) that walks `internal/bootstrap/testdata/packs/core` and asserts:
- Only `pack.toml` exists (no subdirectories, no scripts, no prompts, no overlays).
- `name = "test-core"` in pack.toml (explicit synthetic identity).
- No file content matches any file in `internal/packs/core/` byte-for-byte (prevents copy-paste drift).

**Required:** Define the fixture contract with enumerated allowed paths, forbidden paths, identity choice (`name = "test-core"` recommended), drift posture (intentionally does NOT track production Core schema), and a CI guard function that fails the build if the fixture grows forbidden content or matches production Core files.

### B3. `GC_BOOTSTRAP=skip` is not retired or contained

`internal/bootstrap/bootstrap.go:72` has:

```go
if strings.EqualFold(strings.TrimSpace(os.Getenv("GC_BOOTSTRAP")), "skip") {
    return nil
}
```

This skips **all** of `EnsureBootstrap` — not just retired-entry pruning, but the entire function including any future Core-provenance work. The design says the test-only no-Core path is "not a CLI flag or environment variable" (L388–391) but never addresses the existing `GC_BOOTSTRAP=skip`. The Codex review's finding #3 aligns.

Three production callers depend on this env var being unset:
- `cmd/gc/main_test.go:45,61-62` sets `GC_BOOTSTRAP=skip` as a test default.
- `internal/doctor/implicit_import_cache_check.go:236-245` temporarily unsets it for doctor checks.

If the design removes `BootstrapPacks` materialization entirely (the intended end state), `GC_BOOTSTRAP=skip` becomes vestigial but still short-circuits any future `EnsureBootstrap` responsibility (like Core provenance checks). If it remains, it becomes an undocumented production bypass that contradicts the "Core is required" invariant.

**Required:** Add an explicit migration decision:
- Either retire `GC_BOOTSTRAP=skip` and replace `cmd/gc` testscript defaults with a test-only helper that doesn't set env vars, OR
- Narrow it to skip only retired-entry pruning and document that it does NOT bypass Core provenance checks. Add a test proving Core is still materialized when `GC_BOOTSTRAP=skip` is set.
- In either case, state explicitly that bootstrap fixture isolation does not reuse or extend `GC_BOOTSTRAP=skip`.

### B4. `bootstrapManagedImportNames` in `internal/config/compose.go` is a second hardcoded `"core"` not mentioned in the design

`internal/config/compose.go:839` declares:

```go
var bootstrapManagedImportNames = []string{"registry", "core"}
```

This is the collision-gate list used by `internal/config/compose.go`'s `CollidesWithBootstrapManagedImport` (used at config-load time) and is a separate hardcoded reference to `"core"` from the `requiredBuiltinPackNames` list in `cmd/gc/embed_builtin_packs.go:237`.

The design's "Cross-Pack Ownership Decisions" table and test-slicing section mention `requiredBuiltinPackNames` but never mention `bootstrapManagedImportNames`. When `"maintenance"` is removed from `requiredBuiltinPackNames`, the collision gate must also be updated: `bootstrapManagedImportNames` currently lists `"registry"` and `"core"` (but not `"maintenance"`), so it partially aligns but is not tested for sync with `BootstrapPackNames()`. The comment at L831 references a `TestBootstrapManagedNames_MatchesBootstrapPacks` test that **does not exist** in the current tree — I found no such test in `internal/bootstrap/` or `internal/config/`.

This is a real gap: when `BootstrapPacks` becomes permanently empty (the intended end state), `bootstrapManagedImportNames` should also become empty or be removed, but the design never mentions this path. If `BootstrapPacks` is removed but `bootstrapManagedImportNames` still lists `"core"`, the collision gate will block user imports named `core` even though there is no longer a bootstrap-managed `core` implicit import — an invisible breaking change.

**Required:**
- Add `bootstrapManagedImportNames` to the migration inventory.
- Either create the referenced `TestBootstrapManagedNames_MatchesBootstrapPacks` test or update the comment to remove the reference.
- Specify that when `BootstrapPacks` becomes empty, `bootstrapManagedImportNames` must also become empty (or be removed), with a sync test proving they agree.

### B5. `requiredBuiltinPackNames` hardcodes `"maintenance"` — transition plan is incomplete

`cmd/gc/embed_builtin_packs.go:237` currently lists:

```go
required := []string{"core", "maintenance"}
```

The design mentions removing Maintenance from the embedded pack registry (slice 6) but does not address the transition of `requiredBuiltinPackNames`. During slices 3–5, this list still says `"maintenance"`, meaning:
- `builtinPackIncludes` still returns `.gc/system/packs/maintenance` as a config include path.
- `unusableRequiredBuiltinPackNames` still reports maintenance as required.
- The comment at L268 says "Core and maintenance are always included."

If slice 5 moves Maintenance assets into Core but slice 6 removes Maintenance from `All()`, there is a gap between slices 5 and 6 where `requiredBuiltinPackNames` still lists `"maintenance"` but `All()` no longer includes a Maintenance pack, causing `MaterializeBuiltinPacks` to fail (it iterates `builtinPacks` which comes from `All()`, but required check looks at `requiredBuiltinPackNames`).

**Required:** Add `requiredBuiltinPackNames` to the migration inventory with explicit slice assignments:
- Slice 5: Remove `"maintenance"` from `requiredBuiltinPackNames` and update the comment.
- Verify that `builtinPackIncludes` and `unusableRequiredBuiltinPackNames` work correctly with Core-only required packs between slices 5 and 6.
- Add a test proving that `requiredBuiltinPackNames` matches the packs actually present in `All()`.

### B6. `internal/hooks/hooks.go` directly imports `internal/bootstrap/packs/core` — migration shape unspecified

`internal/hooks/hooks.go:21` imports:

```go
"github.com/gastownhall/gascity/internal/bootstrap/packs/core"
```

And uses `core.PackFS` at L174, L177, L185, and L782. The design says "add hook installation tests proving `internal/hooks` reads overlays from `internal/packs/core`" (L498–499) but does not specify the import change.

The migration must change this import from `internal/bootstrap/packs/core` to `internal/packs/core`. This is a single-point import change (the Claude review's cross-file finding #1 mentions this), but the design should name it explicitly because:
- The import is used in production code (not just tests).
- `desiredCodexPreCompactHook` at L782 reads `core.PackFS` directly as a fallback — this is a hidden dependency that a simple grep for `bootstrap/packs/core` would miss if the import is changed but the reference remains.
- The `installOverlayManaged` function at L174–190 walks `core.PackFS` at runtime. If the new `internal/packs/core.PackFS` embed does not include `overlay/per-provider/`, all provider hook installation silently breaks.

**Required:** Add the `internal/hooks/hooks.go` import change to the Core extraction slice. Specify that the import path changes from `internal/bootstrap/packs/core` to `internal/packs/core` and that `PackFS` must include the exact same `overlay/per-provider/` tree that the current embed provides. The hook-installation test must prove this by verifying at least one non-empty provider overlay installs successfully from the new path.

---

## Major Findings

### M1. Hidden dependency inventory is incomplete — five more files not listed

The design's test-update section names only `internal/builtinpacks/registry_test.go`. I found five additional in-tree files that hardcode `internal/bootstrap/packs/core` as a string or dependency:

1. **`cmd/gc/prompt_test.go:781-782`** — References `internal/bootstrap/packs/core/assets/prompts/{pool-worker,graph-worker}.md`. This is exactly the kind of `cmd/gc` test coupling the design should catch.
2. **`internal/config/bundled_import_test.go:44,68`** — Writes synthetic content at `internal/bootstrap/packs/core/pack.toml` and `.../agents/injected/prompt.md`. After the move, this test either fails or silently tests a dead path.
3. **`internal/remotesource/remotesource_test.go:16,18`** — Pins `internal/bootstrap/packs/core` as the canonical recognized subpath. Must update to the new path.
4. **`internal/hooks/config/README.md:18,59`** — Instructs contributors to add provider overlays under `internal/bootstrap/packs/core/overlay/per-provider/`. Must update to the new path.
5. **`internal/config/compose.go:839`** — `bootstrapManagedImportNames` references `"core"` (addressed in B4).

The design must enumerate all five in the bootstrap/Core-extraction slice with their dispositions (path rewrite, migration diagnostic, or removal).

### M2. `materializeBootstrapPack` uses `os.MkdirTemp` and `os.Rename` for atomic writes

`bootstrap.go:188-203` uses a staging directory pattern for atomic bootstrap pack materialization. This is consistent with the project's atomic-write convention. The design does not need to change this, but implementers should know that `materializeBootstrapPack` will still be called in test overrides after the migration — the staging logic must work with the fixture's minimal content. The fixture's `pack.toml` must be complete enough for `materializeBootstrapPack`'s `pack.toml` existence check at L195.

### M3. `builtinpacks.All()` Subpath and import changes are a single-point-of-change

`registry.go:53-57` maps each pack name to a `Subpath` and `FS`. When Core moves from `internal/bootstrap/packs/core` to `internal/packs/core`, the `Subpath` and import path change. This is a single-point-of-change, but:
- `publicSubpathForPack` (L113-119) returns `"gastown"` and `"maintenance"` as public subpaths. When Maintenance is removed from `All()`, `publicSubpathForPack("maintenance")` must also be removed or it will return a match for a pack that no longer exists in the embedded set.
- `syntheticPackLayouts()` iterates `All()` and builds layouts that include `PublicRepository` aliases. Removing Maintenance from `All()` without removing its public alias will cause `MaterializeSyntheticRepo` to skip it, but stale cache entries under the old subpath may linger.

**Required:** The Maintenance removal slice must also update `publicSubpathForPack` and verify `syntheticPackLayouts` no longer generates layouts for removed packs.

### M4. Collision test `Entry.Name` vs `AssetDir` semantics are conflated

Current collision tests (`collision_test.go:111,149,172`) set:

```go
Entry{Name: "core", AssetDir: "packs/core"}
```

The `AssetDir` field is resolved against `bootstrapAssets` (the embedded FS), while `Name` is the collision key used by `CollidesWithBootstrapPack`. These are independent: after the migration, `AssetDir` should point at the test fixture (e.g., `"packs/test-core"`), while `Name` can stay `"core"` for collision semantics. The design should state this explicitly to prevent implementers from conflating the two fields.

If the fixture uses `name = "test-core"` in its `pack.toml` (as I recommend in B2), the collision tests need to verify that the `Name` field (not the pack.toml name) is what drives collision detection — because `CollidesWithBootstrapPack` operates on `Entry.Name`, not on the pack.toml `name` field. This is a subtlety that could cause false-green tests.

### M5. `MaterializeSyntheticRepo` subpath coupling with bundled-import tests

`internal/config/bundled_import_test.go:44` writes cache content at `internal/bootstrap/packs/core/pack.toml`. `internal/builtinpacks/registry_test.go:197,214,281` does the same. After the migration changes the `Subpath` in `registry.go`, these tests must update their cache paths. The design mentions "source guards for old bootstrap Core dependencies" (L675) but should explicitly name `bundled_import_test.go` and `registry_test.go` as requiring subpath updates alongside the `Subpath` change.

### M6. Hook overlay docs reference old path

`internal/hooks/config/README.md:18,59` still instructs contributors to add provider overlays under `internal/bootstrap/packs/core/overlay/per-provider/`. These must update to `internal/packs/core/overlay/per-provider/` in the source-deletion/docs slice.

### M7. `EnsureBootstrapForCity` collision detection uses `PackNames()` which reads `BootstrapPacks`

`collision.go:47-48` iterates `BootstrapPacks` to get pack names. When `BootstrapPacks` is permanently empty (the intended end state), `PackNames()` returns an empty slice, and `CollidesWithBootstrapPack` with an empty `bootstrapNames` always returns nil — no collision detection at all. Meanwhile, `bootstrapManagedImportNames` in `compose.go` still lists `"core"` and `"registry"`, creating an inconsistency where config-load collision detection blocks `core` imports but bootstrap collision detection does not.

This is the same gap as B4 but viewed from the runtime behavior angle. When `BootstrapPacks` becomes empty, either:
- The bootstrap collision system is removed entirely (since there's nothing to collide with), OR
- It's replaced by a config-time collision check that uses `bootstrapManagedImportNames` (which already exists in `compose.go`).

**Required:** State the intended end state for `CollidesWithBootstrapPack` and `PackNames()` when `BootstrapPacks` is empty.

### M8. The `desiredCodexPreCompactHook` fallback reads `core.PackFS` directly at runtime

`internal/hooks/hooks.go:782` reads `core.PackFS` as a fallback when no managed hooks file exists:

```go
desired, err = iofs.ReadFile(core.PackFS, path.Join("overlay", "per-provider", "codex", ".codex", "hooks.json"))
```

This is a runtime dependency on the embed, not just an install-time dependency. After moving `core.PackFS` to `internal/packs/core`, this must still work. The design's hook-installation tests should cover this fallback path, not just `installOverlayManaged`.

---

## Cross-File Consistency Issues

1. **`bootstrapAssets` is a package-level var with no setter.** Tests that override `BootstrapPacks` (collision_test.go, skills_test.go) all set `AssetDir: "packs/core"`, which works because the embedded `packs/**` includes that content. After removing the embed, these tests need an injection point for `bootstrapAssets`. The design must specify whether to add a `SetBootstrapAssetsForTest` function or use `os.DirFS("testdata/packs")` in `_test.go` files.

2. **`builtinpacks.All()` is the single source of truth for the embedded pack set.** When Core moves from `internal/bootstrap/packs/core` to `internal/packs/core`, the import and `Subpath` change here. The design should note this is a single-point-of-change in `registry.go`, not a scattered update.

3. **`requiredBuiltinPackNames` and `builtinPackIncludes` both hardcode `"core"` and `"maintenance"`.** These are in `cmd/gc/embed_builtin_packs.go:237` and `:276`. The design should confirm that the `"core"` string constant continues to work (it will, because it references the pack name `"core"`, not the source path) and that `"maintenance"` is removed at the correct slice boundary.

4. **The sync test `TestBootstrapManagedNames_MatchesBootstrapPacks` referenced in `compose.go:831` does not exist.** Either create it or remove the comment reference.

5. **`legacyPublicPackImportDetails` in `import_state_doctor_check.go:181` still says "maintenance/core is supplied implicitly".** After the migration, this wording must change to reflect that only Core is implicit and Maintenance is retired. The design's import-state doctor section (L527–560) addresses this but should name this exact message string.

---

## Required Changes Summary

1. **Specify `bootstrapAssets` production default** (B1): empty `fs.FS` implementation, not nil. Add `TestProductionBootstrapAssetsIsEmpty`. Define test-only injection mechanism (`os.DirFS` or `fstest.MapFS` via `_test.go`).

2. **Define the fixture contract** (B2): `pack.toml` only, `name = "test-core"`, `schema = 2`, forbidden paths (no `formulas/`, `orders/`, `skills/`, `overlay/`, `assets/prompts/`, `agents/`), CI guard (`TestBootstrapFixtureIsMinimal`), explicit drift posture (no tracking of production Core schema).

3. **Contain `GC_BOOTSTRAP=skip`** (B3): Add explicit migration decision. At minimum, narrow it to skip only retired-entry pruning, add a test proving Core is still materialized when it is set, and document that it does not bypass Core provenance checks.

4. **Add `bootstrapManagedImportNames` to the migration inventory** (B4): Specify that when `BootstrapPacks` becomes empty, `bootstrapManagedImportNames` must also become empty. Create or remove the referenced sync test. Name the `compose.go:839` hardcoded `"core"` explicitly.

5. **Add `requiredBuiltinPackNames` to the migration inventory** (B5): Specify the slice in which `"maintenance"` is removed. Verify `builtinPackIncludes` and `unusableRequiredBuiltinPackNames` work with Core-only required packs. Add a test proving `requiredBuiltinPackNames` matches `All()`.

6. **Name the `internal/hooks/hooks.go` import change** (B6): `internal/bootstrap/packs/core` → `internal/packs/core`. Prove `PackFS` includes the `overlay/per-provider/` tree. Test `desiredCodexPreCompactHook` fallback, not just `installOverlayManaged`.

7. **Expand hidden-dependency inventory** (M1): Add `cmd/gc/prompt_test.go`, `internal/config/bundled_import_test.go`, `internal/remotesource/remotesource_test.go`, `internal/hooks/config/README.md`, and `internal/config/compose.go` to the migration inventory.

8. **State `Entry.Name` vs `AssetDir` semantics** (M4): After the migration, `AssetDir` points at the test fixture (`"packs/test-core"`), `Name` stays `"core"` for collision semantics. Update collision tests to verify collision detection uses `Entry.Name`, not pack.toml `name`.

9. **Name `MaterializeSyntheticRepo` subpath coupling** (M5): `bundled_import_test.go` and `registry_test.go` write cache content at the old subpath; these must update with the `Subpath` change.

10. **Update hook docs** (M6): `internal/hooks/config/README.md:18,59` references the old path.

11. **Specify end state for `CollidesWithBootstrapPack` and `PackNames()`** (M7): When `BootstrapPacks` is empty, state whether bootstrap collision detection is removed or replaced by config-time collision.

12. **Update `publicSubpathForPack`** (M3): Remove the `"maintenance"` case when Maintenance is removed from `All()`.

## Questions

- Should the fixture `pack.toml` carry `name = "core"` (preserving collision semantics) or `name = "test-core"` (making its synthetic nature explicit)? I recommend `name = "test-core"` with updated collision-test assertions, because `name = "core"` creates an ongoing temptation to copy real Core content into the fixture to "make tests realistic."
- What should `GC_BOOTSTRAP=skip` become? If narrowed to "skip retired-entry pruning only," it is safe to keep. If retired entirely, the `cmd/gc` testscript default must change to a test-only helper. The design should choose one.
- Are there any other callers of `bootstrap.EnsureBootstrap` or `EnsureBootstrapForCity` beyond `cmd/gc` and `internal/doctor`? The design should confirm the audit is complete.
- What happens to `CollidesWithBootstrapPack` and `PackNames()` when `BootstrapPacks` is permanently empty? Should the entire bootstrap collision system be removed?
