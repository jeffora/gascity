# Ritu Raman

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek, Gemini

**Consensus findings:**
- [Blocker] The migration slice boundary is unsafe unless required-Core/systempacks loading lands before or with the removal of bootstrap Core. `internal/materialize/skills.go` still routes Core skill discovery through `bootstrapSkillDirs()` and `bootstrap.PackNames()`; emptying `BootstrapPacks` first can produce an intermediate build that compiles but installs zero Core skills.
- [Blocker] The skill catalog cache must be updated with the Core move. `currentBootstrapCatalogState()` currently keys transient-error state off `bootstrap.PackNames()`; once Core is no longer a bootstrap pack, the cache can ignore `.gc/system/packs/core` failures and propagate an empty or stale skill catalog.
- [Blocker] The hidden-dependency inventory is still incomplete for the source deletion slice. The plan needs file-level dispositions for `internal/materialize/skills.go`, `internal/doctor/implicit_import_cache_check.go`, `internal/bootstrap/collision.go`, `cmd/gc/prompt_test.go`, `internal/remotesource/remotesource_test.go`, `internal/builtinpacks/registry.go`, `publicSubpathForPack`, `requiredBuiltinPackNames`, `All()`, `MaterializeSyntheticRepo`, hook docs, architecture docs, and stale dual-embed comments.
- [Major] The production `bootstrapAssets` contract must be mechanically specified. Reviewers converge on deleting the production `//go:embed packs/**`, making `bootstrapAssets` a private non-nil empty/erroring filesystem, and proving production bootstrap assets are unreachable and never read.
- [Major] Test fixture isolation needs an exact helper and data contract. The design should use inline synthetic fixtures such as `fstest.MapFS` with an `AssetDir` like `packs/test-core`, forbid copied production Core content, and add negative coverage proving a Core-shaped bootstrap fixture cannot satisfy runtime required-Core participation.
- [Major] `GC_BOOTSTRAP=skip` remains load-bearing and unresolved. It must not bypass required Core/systempack materialization, retired-entry pruning, collision checks, typed participation, or normal command behavior; the doctor save/unset/restore path proves the current env behavior is not harmless.
- [Major] Positive coverage is still missing for the new Core source. The design needs named tests proving Core skills and hook overlays resolve from `internal/packs/core` or systempacks after the move, not from `internal/bootstrap/packs/core`.
- [Major] Retired pack directory behavior is underspecified. The design says stale `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories are preserved and diagnosed, but current pruning behavior must be explicitly changed and tested if those directories are not to be deleted.
- [Minor] Contributor-facing cleanup remains part of this lane. Prompt tests, hook README paths, architecture docs, registry/remotesource tests, doctor messages, and the old dual-embed comment all need concrete updates so contributors are not sent to the removed bootstrap Core path.

**Disagreements:**
- Verdict severity differs. Claude, Codex, and Gemini return `approve-with-risks`; DeepSeek returns `block`. Assessment: the persona verdict is `block` because Gemini's slice-order and cache findings are blocker-class even though its headline verdict is softer, and DeepSeek's blocking fixture/env/inventory findings are plausible.
- Claude considers the new source-symbol guard sufficient to resolve the previous hidden-dependency blocker, while Codex, DeepSeek, and Gemini still require explicit file ownership and concrete scanner tests. Assessment: keep the symbol guard, but do not use it as a substitute for slice-level dispositions.
- Reviewers differ on `GC_BOOTSTRAP=skip`: delete it, narrow it to fixture materialization, or document it as test-only/deprecated. Assessment: deletion best matches the no production no-Core invariant; any retained behavior must still run required-Core/systempack loading and retired-entry pruning.
- Fixture naming is unresolved. Assessment: use synthetic fixture paths like `packs/test-core`; keep `Name: "core"` only for tests whose purpose is core-name collision semantics, and state that exception explicitly.
- Fixture content is not agreed. Assessment: minimal inline `pack.toml` is enough for most bootstrap tests, but the design should state whether a dummy executable is required to preserve executable-permission coverage.
- `cmd/gc/prompt_test.go` could read from the new source path or from the Core embedded FS. Assessment: prefer the Core package/embedded FS so the test verifies shipped content rather than a relative source-tree path.

**Missing evidence:**
- A concrete slice assignment that updates `internal/materialize/skills.go` before or with the removal of `internal/bootstrap/packs/core`.
- A test proving Core skills install from `internal/packs/core` or systempacks after `BootstrapPacks` is empty.
- A `skill_catalog_cache.go` design/test showing required system packs, including Core, are tracked during transient filesystem errors.
- A complete import/symbol/path guard artifact covering `internal/bootstrap/packs/core`, `BootstrapPacks`, `PackNames`, `bootstrapSkillDirs()`, `GC_BOOTSTRAP`, hook overlays, generated references, docs, and fixture paths.
- A final `GC_BOOTSTRAP=skip` decision plus tests proving it cannot bypass materialization, retired-entry pruning, collision checks, typed participation, or command behavior.
- A disposition for `internal/doctor/implicit_import_cache_check.go` after the env-var behavior changes.
- A production-default `bootstrapAssets` test isolated from `_test.go` fixture overrides, plus a "never read production bootstrap assets" assertion.
- A named test-only fixture injection helper or API, including exact `AssetDir`, pack `Name`, fixture content, and parallel/race-safety rules.
- A negative test proving a synthetic bootstrap fixture named `core` cannot satisfy runtime required-Core participation.
- File-level migration rows for `cmd/gc/prompt_test.go`, `internal/remotesource/remotesource_test.go`, `internal/bootstrap/collision_test.go`, `internal/builtinpacks/registry_test.go`, `internal/hooks/config/README.md`, affected `engdocs/` references, and the stale dual-embed comment.
- A pruning rule and test for preserving retired `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories while excluding them from active loading.
- Updated doctor message text for retired Maintenance imports.

**Required changes:**
- Combine the Core extraction and required-Core/systempacks loading work, or add a temporary same-slice fallback, so emptying `BootstrapPacks` cannot remove Core skill materialization.
- Update `cmd/gc/skill_catalog_cache.go` so transient-error state includes required system packs such as Core after Core leaves bootstrap.
- Expand the hidden-dependency inventory into an executable table covering every missed production, test, doc, and generated-reference surface, with owning slice, target behavior, and focused verification.
- Delete the production `//go:embed packs/**` and old `embeddedBootstrapPacks` dependency; initialize production `bootstrapAssets` to a private non-nil empty/erroring filesystem.
- Add tests proving production bootstrap assets are empty, cannot expose Core, and are not read in production flows.
- Make `GC_BOOTSTRAP=skip` singular. Prefer removing production semantics; if retained, narrow it so required Core/systempack behavior and retired-entry pruning still run.
- Remove or rewrite doctor env-mutation scaffolding in the same slice as the `GC_BOOTSTRAP` decision.
- Specify the fixture contract around inline `fstest.MapFS` synthetic paths such as `packs/test-core`; delete compatibility-fixture language that preserves `packs/core`.
- Add a negative fixture test proving synthetic bootstrap Core cannot satisfy runtime required-Core participation.
- Add positive tests proving Core skills and hook overlays resolve from the new Core/systempack source, not from `internal/bootstrap/packs/core`.
- Update command/runtime tests so no-Core behavior exists only in lower-level config tests, and update `cmd/gc/prompt_test.go` to read prompts via the Core package/embedded FS or the new Core path.
- Update registry/cache/remotesource tests and functions including `publicSubpathForPack`, `requiredBuiltinPackNames`, `All()`, `MaterializeSyntheticRepo`, and `TestAllAndSourceAreDeterministic`.
- Specify pruning behavior for retired Maintenance/Gastown system-pack directories and test that stale user-modified directories are preserved.
- Update stale docs, comments, and doctor wording so contributors are no longer pointed at `internal/bootstrap/packs/core` or told retired Maintenance is merely supplied implicitly.
