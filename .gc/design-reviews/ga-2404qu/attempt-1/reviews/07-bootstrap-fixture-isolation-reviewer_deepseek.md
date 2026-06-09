# DeepSeek V4 Flash — Bootstrap Fixture Isolation Review

**Lane:** 07-bootstrap-fixture-isolation-reviewer
**Scope:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden dependency discovery.
**Verdict:** block

## Summary

The design's bootstrap cleanup section (L474–499) correctly identifies the target — remove the production `//go:embed packs/**`, isolate test fixtures, and eliminate the dual embed — but leaves five structural gaps that will produce false-green tests, silent production regressions, or implementation dead-ends if handed to implementers as-is. The gaps are: (1) the `bootstrapAssets` variable has no specified post-removal default, and the code path that reads it is conditionally gated by `len(BootstrapPacks) > 0`, which means the design must specify what the default is and whether the gate is sufficient; (2) the `GC_BOOTSTRAP=skip` env var bypasses all bootstrap logic including retired-entry pruning, and the doctor fix path already unsets it before calling `EnsureBootstrap` — a contradiction with the design's "no environment variable" claim; (3) the hidden-dependency inventory misses two code paths that will silently break: the `publicSubpathForPack` mapping in `registry.go` and the `requiredBuiltinPackNames` function in `cmd/gc/embed_builtin_packs.go`; (4) the fixture contract is underspecified and the L491–493 "tiny compatibility embed" clause directly contradicts the L494–497 anti-regression guard; (5) the `cmd/gc/prompt_test.go` disk-read of Core prompt assets is a `cmd/gc` test that references production Core by absolute source path, which the design's test-only hatch does not cover.

I also found six cross-document consistency issues that prior reviews partially address but leave under-specified.

---

## Blocking Findings

### B1. `bootstrapAssets` post-removal default is unspecified; the conditional gate hides a latent nil-FS panic

`internal/bootstrap/bootstrap.go:25` declares `var bootstrapAssets fs.FS = embeddedBootstrapPacks`. When the `//go:embed packs/**` directive is removed (design L490), `embeddedBootstrapPacks` becomes an empty `embed.FS{}` value or must be deleted entirely. Three outcomes are possible:

1. **Deleted entirely.** The `var embeddedBootstrapPacks embed.FS` declaration cannot exist without an embed directive. The Go compiler requires `//go:embed` directives to reference at least one matching file. Since `internal/bootstrap/packs/core/` is the only directory under `packs/`, removing it means `//go:embed packs/**` has zero matches and becomes a compile error. The design must specify that `embeddedBootstrapPacks` is removed and `bootstrapAssets` receives a new default.

2. **Empty `embed.FS{}`.** This compiles but `fs.Stat(embed.FS{}, "packs/core")` returns `os.ErrNotExist`, producing opaque `fs.ErrNotExist` errors from `collectAssetFiles` and `copyEmbeddedTree` when `BootstrapPacks` is non-empty (tests only today). This is not a clear "no fixture injected" diagnostic.

3. **Nil.** `fs.Stat(nil, ...)` panics. This is the worst outcome and must be explicitly ruled out.

The design says "remove the production `//go:embed packs/**` dependency" but never specifies which of these three outcomes applies, nor does it specify what the production default for `bootstrapAssets` should be.

The subtlety that prior reviews miss: `BootstrapPacks` is empty in production (L35: "It is empty for the gc import launch path"), so the `len(BootstrapPacks) > 0` gate at L112 means `bootstrapAssets` is never read in production today. This is why removing the embed is safe — but it also means that any test that sets `BootstrapPacks` to non-empty (collision_test.go, cmd_start_test.go, init_provider_readiness_test.go) and doesn't also inject a fixture into `bootstrapAssets` will silently break. The design must specify that `bootstrapAssets` defaults to an explicit empty-FS implementation in production (not nil), and that test-only injection uses `os.DirFS` or `fstest.MapFS`.

**Required:** Specify that `bootstrapAssets` defaults to an `errFS` struct that returns `fs.ErrNotExist` for all operations (not nil, not an empty `embed.FS{}`). Add `TestProductionBootstrapAssetsIsEmpty` that walks `bootstrapAssets` and fails if any entry is reachable. Define the test injection mechanism (`os.DirFS("testdata/packs")` or `fstest.MapFS` via `_test.go`).

### B2. `GC_BOOTSTRAP=skip` bypasses all bootstrap logic including retired-entry pruning; design claims "no environment variable"

The design states (L388–391) that the no-Core path is "not a CLI flag or environment variable." But `GC_BOOTSTRAP=skip` at `bootstrap.go:72` returns immediately from `EnsureBootstrapForCity`, skipping both materialization and **retired-entry pruning**. This is not a "no-Core" escape — it skips the entire function.

More importantly, `internal/doctor/implicit_import_cache_check.go:147–156` already works around this by unsetting `GC_BOOTSTRAP` before calling `EnsureBootstrap` in its fix path. This proves that `GC_BOOTSTRAP=skip` is already recognized as a problem for production code paths.

The prior reviews flag this env var but don't note the doctor workaround. The design needs an explicit migration decision:

- **Option A:** Retire `GC_BOOTSTRAP=skip`. Replace `cmd/gc/main_test.go:45,61–62` with a test-only injection mechanism. The doctor workaround becomes unnecessary.
- **Option B:** Narrow `GC_BOOTSTRAP=skip` to skip only `BootstrapPacks` materialization, but still run retired-entry pruning. This preserves the test-script convenience but closes the production-bypass hole.
- **Option C:** Document that `GC_BOOTSTRAP=skip` is test-only and will be removed in a future release. Add a deprecation log message.

The current design's "no environment variable" claim is factually wrong while `GC_BOOTSTRAP=skip` exists and is used by tests. The fix path must choose one option and update the code and tests accordingly.

**Required:** Add an explicit decision for `GC_BOOTSTRAP=skip` in the design. At minimum, add a test proving that `GC_BOOTSTRAP=skip` does not prevent Core from being materialized via the `MaterializeBuiltinPacks` path (which is the production path after this migration). If the env var is retired, update `cmd/gc/main_test.go` and `cmd/gc/cmd_start_test.go`.

### B3. Hidden-dependency inventory is incomplete — `publicSubpathForPack` and `requiredBuiltinPackNames` are not listed

The design names `internal/builtinpacks/registry.go` and `cmd/gc/embed_builtin_packs.go` as imports to update, but does not call out two critical functions whose behavior changes when the pack set changes:

1. **`publicSubpathForPack` (registry.go:88–94):** Maps pack names to public repo subpaths. Currently returns `"gastown"` and `"maintenance"` as public aliases. When Maintenance is retired from the embedded pack set, this function must stop mapping `"maintenance"` or it will produce invalid synthetic repo entries. This is not mentioned in the design.

2. **`requiredBuiltinPackNames` (cmd/gc/embed_builtin_packs.go:237–247):** Currently returns `["core", "maintenance"]` plus provider-dependent `"bd"` or `"dolt"`. When Maintenance is retired, this function must remove `"maintenance"` from the required set. The design mentions this implicitly through the runtime retirement table but never names the function.

3. **`All()` (registry.go:49–55):** Returns the hardcoded pack list including `"core"`, `"bd"`, `"dolt"`, `"maintenance"`, `"gastown"`. When Core moves and Maintenance is retired, both entries change. The `TestAllAndSourceAreDeterministic` test (registry_test.go:15–26) hardcodes the full expected list including `"core=internal/bootstrap/packs/core"` and `"maintenance=examples/gastown/packs/maintenance"`. This test must be updated in the same slice that changes the registry.

These are single-point-of-change locations that, if missed, will produce build errors or runtime misbehavior. They should appear in the design's hidden-dependency inventory with explicit dispositions.

**Required:** Add `publicSubpathForPack`, `requiredBuiltinPackNames`, and `All()` to the hidden-dependency inventory with their slice-specific dispositions. Note that `TestAllAndSourceAreDeterministic` must be updated in the registry/cache slice.

### B4. Fixture contract is contradictory — L491–493 "tiny compatibility embed" conflicts with L494–497 guard

The design says (L491–493):

> Tests that override `BootstrapPacks` with `AssetDir: "packs/core"` continue to find content through a tiny compatibility embed under `internal/bootstrap/testdata/packs/core` injected from `_test.go` files.

And then (L494–497):

> After the slice, a source guard rejects any lingering import, `AssetDir`, or path constant referencing `internal/bootstrap/packs/core`.

These two statements are contradictory. If the guard rejects `AssetDir: "packs/core"`, then the compatibility embed at `testdata/packs/core` cannot use that `AssetDir` value. If the compatibility embed uses `AssetDir: "packs/test-core"`, then every test that sets `AssetDir: "packs/core"` must be updated to `AssetDir: "packs/test-core"`, which is a different kind of change.

The Claude review correctly identifies this contradiction and proposes deleting the L491–493 clause. The Gemini review proposes `AssetDir: "packs/test-core"` with `Name: "core"`. Both are viable, but the design itself must resolve the contradiction, not just the reviewers.

Additionally, the design says "fixture assets do not hash/copy production Core" (L615) but never specifies what the fixture *does* contain. The collision tests only need `pack.toml` with `name = "core"` (or `"test-core"`). But `materializeBootstrapPack` also checks for `pack.toml` existence (bootstrap.go:194), sets executable permissions on `.sh`/`.py`/`.bash` files (bootstrap.go:262–276), and hashes asset content (bootstrap.go:169–176). A `pack.toml`-only fixture exercises the hash and rename paths but never the executable-permission branch. The design should state whether this is intentional.

**Required:** Delete or replace L491–493. Specify that test fixtures use `AssetDir: "packs/test-core"` (or similar) and the L494–497 guard allows `AssetDir: "packs/test-core"` but rejects `AssetDir: "packs/core"`. State the fixture content explicitly (minimal `pack.toml` only, or include a dummy file to exercise the exec-permission branch). Clarify whether collision tests use `Name: "core"` (preserving collision semantics) or `Name: "test-core"` (making the fixture synthetic).

### B5. `cmd/gc/prompt_test.go` reads production Core prompts from disk — not covered by the test-only hatch

`cmd/gc/prompt_test.go:781–782` reads two Core prompt files from disk:

```go
"internal/bootstrap/packs/core/assets/prompts/pool-worker.md",
"internal/bootstrap/packs/core/assets/prompts/graph-worker.md",
```

This is a `cmd/gc` test that references the production Core path by string literal. The design's test-only hatch (L384–386) allows `internal/config` tests to call `config.LoadWithIncludes` directly, but `cmd/gc` tests are in the boundary layer that should use `MaterializeBuiltinPacks`. This test reads the file directly from the source tree on disk, not through `PackFS`.

After Core moves to `internal/packs/core`, this test will fail with `os.ReadFile` returning `ENOENT`. The design's hidden-dependency inventory does not list this file, and the test-only hatch does not cover it because it's a direct `os.ReadFile` of a source-tree path, not a config-loading call.

**Required:** Add `cmd/gc/prompt_test.go` to the hidden-dependency inventory. Specify whether this test reads from `internal/packs/core/assets/prompts/` (new path) after migration, or uses `core.PackFS` (the embedded FS) instead of `os.ReadFile`. The latter is more robust because it tests the embedded content rather than the source tree.

---

## Major Findings

### M1. `MaterializeSyntheticRepo` subpath coupling affects all bundled pack tests

`bundled_import_test.go:44,68` writes synthetic content at `internal/bootstrap/packs/core/pack.toml` and `.../agents/injected/prompt.md` inside a synthetic repo cache. The `MaterializeSyntheticRepo` function creates a repo-shaped directory tree using `Subpath` from `All()`. When Core's Subpath changes from `internal/bootstrap/packs/core` to `internal/packs/core`, these test paths must change too. The prior reviews mention this file, but don't emphasize that `MaterializeSyntheticRepo` writes *all* packs into a single directory tree, so the subpath change affects the entire synthetic repo structure, not just the Core entry.

The `TestSourceRecognitionVariants` test in `remotesource_test.go` also hardcodes `internal/bootstrap/packs/core` as a subpath. This is a URL-parsing test, not a Core-specific test, but the hardcoded string must be updated.

### M2. Hook overlay README references the old path

`internal/hooks/config/README.md:18,59` references `internal/bootstrap/packs/core/overlay/per-provider/...`. After migration, this must point to `internal/packs/core/overlay/per-provider/...`. The design's source-deletion/docs slice should include this file, but it's not in the inventory. The Claude review mentions this; I want to emphasize it as a developer-facing doc that will send contributors to the wrong path.

### M3. Architecture docs reference `internal/bootstrap/packs/core/formulas/...`

Six `engdocs/` files reference `internal/bootstrap/packs/core/formulas/...` paths:
- `engdocs/architecture/formulas.md` (2 references)
- `engdocs/architecture/v1-formula-audit.md` (6 references)
- `engdocs/proposals/skill-materialization.md` (1 reference)
- `engdocs/proposals/skill-materialization-handoff.md` (2 references)
- `engdocs/proposals/skill-materialization-implementation-plan.md` (2 references)

These are not in the hidden-dependency inventory and are not test code, so they won't be caught by source guards scanning `_test.go` files. They should be listed in the source-deletion/docs slice.

### M4. `TestProductionBootstrapAssetsIsEmpty` must exclude `_test.go` overrides

The proposed `TestProductionBootstrapAssetsIsEmpty` assertion (B1 above) must verify that `bootstrapAssets` contains no entries reachable via `fs.WalkDir`. But if `_test.go` files reassign `bootstrapAssets` in an `init()` or TestMain, the assertion could pass in isolated unit tests but fail in a test binary that includes both the production assertion and the fixture override. The design should specify that the assertion runs in a separate test binary or uses a build tag to ensure it only sees the production default.

### M5. Doctor `EnsureBootstrap` call path is partially documented

The design mentions `internal/doctor` as a consumer of bootstrap logic, but the specific code path in `implicit_import_cache_check.go` is more nuanced than the design acknowledges. The `ensureBootstrapForDoctor` function (L147–156) unsets `GC_BOOTSTRAP` before calling `EnsureBootstrap(gcHome)`. This means:
- `EnsureBootstrap` is called without the `GC_BOOTSTRAP=skip` bypass, so retired entries *are* pruned.
- But `BootstrapPacks` is empty in production, so the materialization loop at L112–136 is a no-op.
- After the migration, if `BootstrapPacks` stays empty, this call path only prunes retired entries and does nothing else. This is correct but should be documented as the canonical production path.

### M6. The `embed.go` comment claims dual embed is intentional

`internal/bootstrap/packs/core/embed.go` contains the comment:

> The same content is also reachable through the bootstrap's global packs/** embed, but exposing a dedicated PackFS lets cmd/gc's per-city MaterializeBuiltinPacks pipeline handle core uniformly with bd, dolt, maintenance, and gastown.

This comment documents the current dual-embed design as intentional. The design's L616 assertion ("production embeds no Core tree under `internal/bootstrap`") must be verified against this comment — if the embed.go comment survives the migration without update, it will document a stale rationale. The design should specify that this comment is updated or removed.

---

## Cross-Document Consistency Issues

1. **Requirements say "no Core, Maintenance, or Gastown pack source under `examples/gastown/packs/`" but the design's test-only hatch allows `config.LoadWithIncludes` to be called without Core in tests.** The requirements (L58–62) say Core must be available without importing Gastown. The design (L384–386) allows `internal/config` tests to call `config.LoadWithIncludes` directly, bypassing Core inclusion. These are consistent (tests vs production), but the requirements should note the test-only exception explicitly.

2. **Requirements say "`dog` is a configurable default" but the design's Core extraction slice doesn't specify how the `dog` agent definition migrates from Maintenance to Core.** The Maintenance pack owns `dog` today. The design's Maintenance folding slice (slice 5) should name the `dog` agent definition as an asset that moves to Core, but it doesn't.

3. **The design's review-gated invariants (L20–34) say "Core assets remain role-neutral outside explicitly allowed Core maintenance configuration."** The `dog` agent definition in the current Maintenance pack references Gastown-specific requesters and notification paths (per the requirements' ownership table). The design must specify how Core's `dog` agent is generalized to remove these Gastown-specific references.

4. **The requirements say "existing cities with legacy Maintenance imports get those imports removed by `gc doctor --fix`" (L80–82) but the design's doctor migration section (L445–467) says `gc doctor --fix` removes Maintenance imports with a message that they're "supplied implicitly."** After the migration, Maintenance is retired and Core is the replacement. The doctor message must say "Maintenance is retired; Core supplies generic maintenance behavior" rather than "supplied implicitly." The design's L458 says the wording needs updating but doesn't give the new text.

5. **The design says "stale `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories are ignored, diagnosed, and preserved, not deleted" (L26–27) but `MaterializeBuiltinPacks` in `embed_builtin_packs.go` calls `pruneStaleGeneratedPackFiles` which deletes files in non-required packs that don't match the embedded content.** The design's L26–27 invariant conflicts with the current behavior for non-required packs. The design must specify whether `pruneStaleGeneratedPackFiles` is changed to skip the `maintenance` and `gastown` directories entirely, or whether these directories are simply not generated in the first place after the registry change.

6. **The requirements list `internal/hooks/hooks.go` as importing `internal/bootstrap/packs/core` directly, but the design's hook section (L498–499) says "a new hook-installation test proving overlays load from `internal/packs/core`."** The current hooks code reads provider overlays from `core.PackFS` (not from disk). After the migration, the import changes from `internal/bootstrap/packs/core` to `internal/packs/core`, but the read mechanism (`fs.WalkDir(core.PackFS, ...)`) stays the same. The design should confirm this is a pure import-path change, not a behavioral change.

---

## Required Changes Summary

1. **Specify `bootstrapAssets` production default** (B1): `errFS` struct returning `fs.ErrNotExist`, not nil. Add `TestProductionBootstrapAssetsIsEmpty`. Define test injection mechanism.

2. **Resolve `GC_BOOTSTRAP=skip`** (B2): Choose one of retire/narrow/document. Add a test proving Core is still materialized via `MaterializeBuiltinPacks` when `GC_BOOTSTRAP=skip` is set. Note the doctor workaround at `implicit_import_cache_check.go:147–156`.

3. **Complete hidden-dependency inventory** (B3 + B5): Add `publicSubpathForPack`, `requiredBuiltinPackNames`, `All()`, `cmd/gc/prompt_test.go`, `internal/remotesource/remotesource_test.go`, `engdocs/` architecture files, and `internal/hooks/config/README.md`.

4. **Resolve fixture contract contradiction** (B4): Delete L491–493 "tiny compatibility embed." Specify `AssetDir: "packs/test-core"` with `Name: "core"` (or `"test-core"` with updated collision tests). State fixture content explicitly. State whether the exec-permission branch needs coverage.

5. **Add `cmd/gc/prompt_test.go` disposition** (B5): Specify whether this test reads from `internal/packs/core/` (new path) or uses `core.PackFS`.

6. **Name the `GC_BOOTSTRAP=skip` decision** in the Review-Gated Migration Invariants or as a same-slice note in the Core loading/doctor slice.

7. **Verify the single-embed invariant** (M6): Specify that `internal/bootstrap/packs/core/embed.go` is updated or removed, and its "dual embed" comment is deleted.

8. **Clarify `pruneStaleGeneratedPackFiles` behavior** for retired pack directories (cross-doc #5): Specify that stale `maintenance`/`gastown` directories under `.gc/system/packs/` are skipped by the pruning function, not deleted.

9. **Update doctor message** for retired Maintenance imports (cross-doc #4): Change "supplied implicitly" to "retired; Core supplies generic maintenance behavior."

## Questions

- Should the bootstrap fixture use `os.DirFS("testdata/packs")` (shipping test content in the test directory) or `fstest.MapFS` (no on-disk fixture at all)? `os.DirFS` means a `_test.go`-visible `testdata/` directory that could become a drift target; `fstest.MapFS` means fixture content is defined inline in the test file. The prior reviews assume `os.DirFS` but this choice has implications for the L616 invariant and for whether any test content ships in the production binary.
- Does the `MaterializeSyntheticRepo` function need to change its `Subpath` entries in the same slice as the Core extraction, or should it change in the registry/cache slice (slice 6)? The design doesn't specify. Changing it in slice 3 would break `bundled_import_test.go` before slice 6 updates the `All()` return value.
- Should `GC_BOOTSTRAP=skip` be retired entirely, or narrowed to skip only `BootstrapPacks` materialization while still running retired-entry pruning? The current behavior (skip everything) is convenient for testscript but blocks doctor's fix path.
- The collision tests use `Name: "core"` and `AssetDir: "packs/core"`. If the fixture uses `AssetDir: "packs/test-core"`, should `Name` stay `"core"` to preserve collision semantics, or change to `"test-core"` to make the fixture's synthetic nature explicit? The Gemini review recommends `"test-core"` with updated assertions; the Claude review recommends keeping `"core"` for collision tests only. The design should choose one.
