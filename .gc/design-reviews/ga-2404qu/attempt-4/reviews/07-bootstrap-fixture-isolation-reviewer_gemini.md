# Gemini 3.5 Flash — Bootstrap Fixture Isolation Review (Iteration 4)

**Lane:** 07-bootstrap-fixture-isolation-reviewer
**Scope:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden dependency discovery.
**Verdict:** block

## Summary

In Iteration 4, the design-before document (`.gc/design-reviews/ga-2404qu/attempt-4/design-before.md`) has made significant progress in defining the test-isolation architecture for `internal/bootstrap/bootstrap.go`. Specifically, it now explicitly proposes introducing a minimal synthetic fixture under `internal/bootstrap/testdata/packs/core`, specifies utilizing a single test-injection approach, and ensures that production `BootstrapPacks` default remains permanently empty without Dual-Embed of Core.

However, three critical structural gaps and one newly-discovered pattern drift issue remain unaddressed in the Iteration 4 design. These issues would likely lead to compilation errors, runtime panics, or developer-friction drift during the migration. Most notably, the design is silent on the hardcoded `bootstrapManagedImportNames` collision filter in `internal/config/compose.go`, which is directly coupled to `BootstrapPacks` and will break existing sync tests if left uncoordinated.

To ensure complete bootstrap-test isolation and robust production containment, this review presents the evidence-based findings and required changes below.

---

## Blocking Findings

### B1. `bootstrapAssets` post-removal default is unspecified; nil FS causes panics

`internal/bootstrap/bootstrap.go:25` currently declares `var bootstrapAssets fs.FS = embeddedBootstrapPacks`. When the `//go:embed packs/**` directive is removed as proposed, `embeddedBootstrapPacks` will be deleted from the file (as it will have zero matching files on disk, and the Go compiler rejects empty embed patterns).

The design states "Remove the production `//go:embed packs/**` dependency from `internal/bootstrap/bootstrap.go`" (L691-692), but fails to define what `bootstrapAssets` defaults to in production.
- If it defaults to `nil`, any execution path (such as test overrides of `BootstrapPacks` that do not explicitly re-assign `bootstrapAssets` in test files) reaching `collectAssetFiles` (L170), `copyEmbeddedTree` (L194), or any standard `fs` utility calls (such as `fs.WalkDir`) will trigger a nil-pointer dereference panic at runtime.
- If it defaults to an empty `embed.FS`, operations such as `fs.Stat` will return `fs.ErrNotExist`, which must be handled cleanly.

**Required:** Specify that `bootstrapAssets` in production defaults to a dedicated, non-nil, empty filesystem implementation (such as a private `errFS` struct defined in `bootstrap.go` that returns `fs.ErrNotExist` for all operations). Add `TestProductionBootstrapAssetsIsEmpty` to walk `bootstrapAssets` and assert that no assets are reachable in production, while test-only files explicitly inject the synthetic test fixture using `os.DirFS` or `fstest.MapFS`.

### B2. Lack of scanner-enforced CI guard against copying production Core into test fixture

The design successfully mandates introducing a minimal synthetic fixture under `internal/bootstrap/testdata/packs/core` (L684) and states that bootstrap tests "must not read production Core assets from `internal/packs/core`" (L686-687). 

However, it lacks an explicit, scanner-enforced CI guard to prevent drift or regressions. Without an automated check, future contributors who need a quick fix for a failing bootstrap test will be tempted to copy-paste parts of the production Core pack (e.g., formulas, prompts, overlays) into the synthetic test fixture, completely undermining the dependency-coupling isolation this migration seeks to establish.

**Required:** Define a strict CI guard function (e.g., `TestBootstrapFixtureIsMinimal` in `internal/bootstrap/collision_test.go`) that walks the synthetic fixture directory (`internal/bootstrap/testdata/packs/core` or `test-core`) and fails the build if any production-only directories such as `formulas/`, `orders/`, `overlay/`, `skills/`, or `assets/prompts/` are present in the fixture.

### B3. `GC_BOOTSTRAP=skip` containment and narrowed semantics are unspecified

The current implementation of `EnsureBootstrapForCity` (L72) aborts immediately when the environment variable `GC_BOOTSTRAP=skip` is set. 
The design claims "The dev/test escape hatch is not a CLI flag or environment variable" (L590) but does not provide an explicit migration path or decision for the existing `GC_BOOTSTRAP=skip` code.
- If `GC_BOOTSTRAP=skip` is kept as a test-only convenience, its semantics must be narrowed to skip *only* bootstrap-level materialization (which is already empty in production), but it must never bypass Core system pack materialization (which now happens via `MaterializeBuiltinPacks` in `builtinpacks`).
- If it is retired, the design must specify updating the testscript setup in `cmd/gc/main_test.go` and removing the `os.Unsetenv("GC_BOOTSTRAP")` workaround in `internal/doctor/implicit_import_cache_check.go`.

**Required:** Document an explicit, concrete decision for `GC_BOOTSTRAP=skip` in the design. If kept, narrow its scope and add a test (e.g., in `internal/bootstrap`) proving that setting `GC_BOOTSTRAP=skip` does not prevent Core from being materialized and verified via `builtinpacks` and config-loading pathways.

### B4. Missing synchronization for `bootstrapManagedImportNames` in `compose.go`

In `internal/config/compose.go`, the composer defines:
```go
var bootstrapManagedImportNames = []string{"registry", "core"}
```
And `internal/bootstrap/collision_test.go` has a sync test (`TestBootstrapPackNamesMatchesEntries`) asserting that `PackNames()` matches `BootstrapPacks`.

When `BootstrapPacks` becomes permanently empty in production, `PackNames()` will return an empty list. If `bootstrapManagedImportNames` still hardcodes `"core"`, the lists will diverge, and the unit tests will FAIL. Furthermore, keeping `"core"` in `bootstrapManagedImportNames` when bootstrap no longer manages it causes pattern drift in the composer's collision gate.

**Required:** Update `internal/config/compose.go` to remove `"core"` (and `"registry"`) from `bootstrapManagedImportNames` in the same slice that permanently clears `BootstrapPacks`. Update the sync test `TestBootstrapPackNamesMatchesEntries` to assert that bootstrap-managed implicit imports are permanently empty, and verify that the composer's collision gate behaves correctly when `BootstrapManagedImportNames()` returns an empty list.

---

## Major Findings

### M1. Unlisted String Literal Dependencies in `prompt_test.go` and `bundled_import_test.go`

The hidden-dependency inventory of the design fails to list two Go test files that hardcode paths pointing directly into the old `internal/bootstrap/packs/core` structure:

1. **`cmd/gc/prompt_test.go:781-782`**
   Reads production Core prompts directly from disk:
   ```go
   "internal/bootstrap/packs/core/assets/prompts/pool-worker.md"
   "internal/bootstrap/packs/core/assets/prompts/graph-worker.md"
   ```
   After migration, these paths will be dead. Since `cmd/gc` tests are at the boundary layer, this test must be updated to either read from the new path `internal/packs/core/assets/prompts/` or utilize `core.PackFS` to extract the embedded prompt files.

2. **`internal/config/bundled_import_test.go:44,68`**
   Writes synthetic cache contents under `internal/bootstrap/packs/core/...`. Because `MaterializeSyntheticRepo` relies on `Subpath` values defined in `registry.go` (which will change from `"internal/bootstrap/packs/core"` to `"internal/packs/core"`), these test-setup paths must be updated to prevent test failures.

**Required:** Add `cmd/gc/prompt_test.go` and `internal/config/bundled_import_test.go` to the migration inventory. Ensure that the source-code guards and path-replacement checks cover these files.

### M2. Single-Embed Invariant Verification

The design proposes removing the production `//go:embed packs/**` from `bootstrap.go` (L691-692), while keeping `internal/packs/core.PackFS` as the sole source of Core assets (L332). 
However, `internal/bootstrap/packs/core/embed.go` currently carries the comment:
```go
// The same content is also reachable through the bootstrap's global packs/** embed...
```
This comment explicitly documents the dual-embed as intentional.

**Required:** Ensure that the old `embed.go` file under `internal/bootstrap/packs/core` is deleted completely, and add a test-level invariant verifying that Core assets are embedded exactly once in the compiled binary.

---

## Minor Findings

### m1. Collision Test `Entry.Name` vs `AssetDir` Disambiguation

In `internal/bootstrap/collision_test.go`, the collision tests utilize `AssetDir: "packs/core"` and `Name: "core"`.
- `Name` is the collision key.
- `AssetDir` is the directory matched against `bootstrapAssets`.

When introducing the synthetic test fixture, the test should update `AssetDir` to point to the fixture (e.g., `"packs/test-core"`), but `Name` should remain `"core"` (or be configurable) to preserve the exact collision-prevention semantics of the compiler. The design should explicitly point out this distinction to prevent implementers from conflating the two fields.

### m2. Hook Documentation Paths in `internal/hooks/config/README.md`

`internal/hooks/config/README.md` contains multiple references (L18, L59) instructing developers to add provider overlays under `internal/bootstrap/packs/core/overlay/per-provider/`. These documentation references must be updated in the final documentation slice to point to `internal/packs/core/overlay/per-provider/` to prevent developer confusion.

---

## Required Changes Summary

1. **Production `bootstrapAssets` Default (B1):** Specify an explicit, private, non-nil empty `fs.FS` type (such as an `errFS` struct returning `fs.ErrNotExist` for all operations) as the production fallback. Add a CI test to assert that `bootstrapAssets` remains empty in production.
2. **CI Fixture Guard (B2):** Add a strict CI validation test (`TestBootstrapFixtureIsMinimal`) to prevent any production Core subdirectories (such as `formulas/`, `orders/`, `overlay/`, `skills/`, or `assets/prompts/`) from being copy-pasted into the synthetic fixture.
3. **`GC_BOOTSTRAP=skip` Resolution (B3):** Formulate an explicit decision on the env var. If kept, narrow its scope to skip only bootstrap materialization and add a test to verify it does not bypass the core system-pack loading.
4. **`bootstrapManagedImportNames` Sync (B4):** Remove `"core"` and `"registry"` from the composer's `bootstrapManagedImportNames` slice in `internal/config/compose.go` in the same slice that clears `BootstrapPacks`. Update the matching sync tests.
5. **Complete Dependency Inventory (M1):** Update `cmd/gc/prompt_test.go` and `internal/config/bundled_import_test.go` to reference the new paths and registry subpaths.
6. **Single-Embed Comment Cleanup (M2):** Delete the old bootstrap-core `embed.go` and update any comments documenting the stale dual-embed architecture.

---

## Questions

- **Should the bootstrap fixture utilize an on-disk `testdata` directory or inline `fstest.MapFS`?** Utilizing `fstest.MapFS` inside `bootstrap_test.go` removes the need for an on-disk directory under `internal/bootstrap/testdata`, completely eliminating any possibility of a developer copy-pasting production files into the test directory. I highly recommend `fstest.MapFS` for maximum isolation.
- **Is `GC_BOOTSTRAP=skip` required for any downstream CI environments, or can we retire it entirely in favor of programmatic test-only loaders?** If no external scripts depend on it, retiring it entirely is the cleanest way to prevent production bypass.
