# Ritu Raman

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek, Gemini (`_gemini.md` artifact is also headed DeepSeek V4 Flash)

**Consensus findings:**
- [Blocker] The bootstrap fixture is not yet a safe contract. Reviewers agree the design must define fixture identity, allowed contents, forbidden production-Core-like paths, schema/drift posture, and the distinction between `Entry.Name` collision semantics and `Entry.AssetDir` source semantics before implementation.
- [Blocker] The post-removal `bootstrapAssets` default is underspecified. Removing `//go:embed packs/**` is directionally correct, but the design must specify a non-nil empty production `fs.FS`, remove the orphaned embed/import path, and add a concrete assertion that production bootstrap exposes no Core tree.
- [Blocker] `GC_BOOTSTRAP=skip` conflicts with the stated "no environment variable" no-Core policy. The design must retire, narrow, or explicitly document it, and prove it cannot bypass required Core materialization or retired-entry pruning in production paths.
- [Major] The hidden-dependency inventory is incomplete. Missing or under-specified references include `cmd/gc/prompt_test.go`, `internal/config/bundled_import_test.go`, `internal/remotesource/remotesource_test.go`, `internal/hooks/config/README.md`, architecture docs, registry code/tests, synthetic repo/cache paths, and old-path string literals beyond Go imports.
- [Major] Hook overlay migration needs explicit proof. `internal/hooks` currently reads provider overlays from the old bootstrap Core package; the design should state the new `internal/packs/core` dependency shape and require tests across supported overlays, not a single-path smoke test.
- [Major] Registry and materialization coupling must be named. `builtinpacks.All()`, `publicSubpathForPack`, `MaterializeSyntheticRepo`, `requiredBuiltinPackNames`, and stale-pack pruning behavior all have Core/Maintenance path or retirement consequences that need same-slice dispositions.
- [Major] Retired implicit-import pruning must stay tested after production bootstrap assets are empty. `EnsureBootstrap` remains behaviorally relevant even when `BootstrapPacks` no longer materializes production Core.

**Disagreements:**
- Claude and Codex return `approve-with-risks`; DeepSeek and Gemini block. Assessment: the shared hazards are concrete compile, test, and runtime migration gaps, so this lane's verdict is `block`.
- Reviewers differ on fixture identity. Claude emphasizes keeping `Entry.Name: "core"` for collision semantics while changing the asset source; Codex asks for a narrow allowlist if any fixture is named `core`; DeepSeek/Gemini prefer an unambiguous synthetic fixture such as `test-core`. Assessment: the design must choose explicitly, and should separate the collision key from the fixture asset directory.
- Reviewers differ on fixture drift handling. Assessment: require a guard that prevents production-Core-like fixture contents, but do not require byte-for-byte drift tracking against shipped Core because that would re-couple bootstrap tests to Core assets.
- Codex leaves the hook overlay source as an open seam choice; DeepSeek/Gemini say direct import of `internal/packs/core` is acceptable. Assessment: direct import is acceptable for this migration if tests prove overlays load from the new package and the old bootstrap import is forbidden.
- DeepSeek recommends an `errFS` default specifically; other reviews only require a non-nil empty FS. Assessment: the design must name a concrete default and diagnostics; `nil` and leftover production embeds are unacceptable.

**Missing evidence:**
- Exact synthetic fixture contents and whether it includes only `pack.toml` or a small executable/asset file to preserve chmod/hash branch coverage.
- Fixture `pack.toml` identity and schema policy, including whether it uses `name = "core"` or `name = "test-core"` and whether it intentionally ignores future production Core schema drift.
- A CI/source guard proving the fixture does not contain production-Core-like directories such as `assets/prompts`, `formulas`, `orders`, `overlay`, or `skills`.
- Concrete production default for `bootstrapAssets` after `//go:embed packs/**` is removed, plus proof no Core files are reachable from production bootstrap.
- Complete scan scope for old-path references across Go, Markdown, TOML, shell/test fixtures, docs, string literals, `AssetDir` values, and source-path constants.
- Explicit `GC_BOOTSTRAP=skip` disposition and tests proving it does not become a no-Core runtime escape hatch.
- Hook overlay test matrix for supported providers and a guard against importing `internal/bootstrap/packs/core`.
- Disposition for Maintenance retirement interactions, including `requiredBuiltinPackNames`, doctor wording, and stale `.gc/system/packs/maintenance` / `gastown` pruning behavior.

**Required changes:**
- Define the bootstrap fixture contract: allowed files, forbidden production-Core paths, identity/schema stance, and explicit statement that bootstrap fixture drift detection against shipped Core is intentionally absent.
- Use an unambiguous test-only fixture path such as `internal/bootstrap/testdata/packs/test-core` or `fstest.MapFS`; inject it only from `_test.go` code and remove the "tiny compatibility embed" option from the design.
- Update bootstrap collision tests so `Entry.AssetDir` points at the synthetic fixture while `Entry.Name` preserves the intended collision key, if the design keeps testing collisions on `core`.
- Specify the production `bootstrapAssets` default as a concrete non-nil empty FS, remove the orphaned bootstrap embed/import, and add an assertion that production bootstrap exposes no Core tree or `pack.toml`.
- Add source guards for lingering `internal/bootstrap/packs/core`, old `AssetDir: "packs/core"` values, `//go:embed packs/**`, and old-path strings outside explicit migration allowlists.
- Expand the hidden-dependency inventory to cover prompt tests, bundled-import tamper tests, remote-source tests, registry code/tests, synthetic repo/cache paths, hook docs, architecture docs, and Markdown/TOML references.
- Move hook overlay reads to the canonical Core package and add tests over supported provider overlays that fail if the old bootstrap Core package remains in use.
- Name and schedule the registry/materialization changes for `All()`, `publicSubpathForPack`, `MaterializeSyntheticRepo`, `requiredBuiltinPackNames`, and stale retired-pack pruning.
- Preserve focused tests for `EnsureBootstrap` retired implicit-import pruning after production `BootstrapPacks` and bootstrap assets are empty.
- Add the `GC_BOOTSTRAP=skip` migration decision and tests proving Core loading still comes from the required builtin-pack path rather than a runtime no-Core bypass.
