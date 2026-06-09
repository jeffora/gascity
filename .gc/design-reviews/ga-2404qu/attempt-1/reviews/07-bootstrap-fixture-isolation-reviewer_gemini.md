# Ritu Raman — DeepSeek V4 Flash Perspective Independent Review (Iteration 17 / Attempt 1)

**Verdict:** approve-with-risks

**Lane:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden dependency discovery.

## Summary

The proposed design for the Core and Gastown pack split successfully addresses the physical isolation of Core assets and test-fixture safety. Shifting bootstrap tests to inline synthetic `fstest.MapFS` fixtures and retiring the production skip behavior under `GC_BOOTSTRAP` are major architectural wins. 

However, from a DeepSeek-style forensic review perspective, several hidden mechanism-level dependencies, compilation risks, and cross-document inconsistencies remain unaddressed. If these are not resolved before downstream migration slices, they will lead to compilation failures or silent runtime regressions.

---

## Technical Evaluation of Invariant Questions

### Q1. Does `internal/bootstrap` stop embedding production Core while keeping bootstrap tests deterministic through explicit isolated fixtures?
* **Yes, in principle:** Removing `//go:embed packs/**` and pointing `bootstrapAssets` to an empty/erroring filesystem prevents the production binary from carrying duplicate Core assets.
* **Risks:** The transition must completely delete the embedded variable declarations and comments to avoid compile-time failures when no files match the `packs/**` glob pattern. Tests must explicitly use inline `fstest.MapFS` to guarantee determinism without on-disk file dependencies.

### Q2. How is fixture drift against the shipped Core pack detected without causing low-level config tests to exercise production assets accidentally?
* **Decoupled Verification:** Content fidelity and drift detection are moved entirely to systempacks integration tests, while low-level tests verify parsing and loading behavior on synthetic schemas.
* **Leakage Prevention:** To keep tests from accidentally reading real Core from the workspace, the design relies on isolated test setups (e.g., `t.TempDir()`) and a path-string guard. However, a path-string guard alone is insufficient; symbolic imports and global references must also be blocked via an import-graph and symbol scanner.

### Q3. Are tests needing no-Core behavior using structurally test-only lower-level loaders rather than runtime flags or environment switches?
* **Yes:** The environment variable `GC_BOOTSTRAP=skip` is retired for production commands. Tests requiring no-Core behavior must call lower-level config loader packages directly. 

---

## Critical Risks, Gaps & Hidden Dependencies

### 1. The Redundant `GC_BOOTSTRAP` Env-Mutation Scaffolding
* **The Code Path:** `internal/doctor/implicit_import_cache_check.go` (`ensureBootstrapForDoctor`) unsets, saves, and restores `GC_BOOTSTRAP` around `bootstrap.EnsureBootstrap`.
* **The Gap:** When `bootstrap.EnsureBootstrap` becomes a no-op (except for retired-entry pruning), this save/restore dance is obsolete and confusing.
* **Required Change:** Slate `ensureBootstrapForDoctor` for refactoring/deletion. All doctor-level implicit-import checking should use the central mutation coordinator and standard packsource resolution instead.

### 2. Compilation Vulnerability of Empty Go Embeds
* **The Code Path:** `internal/bootstrap/bootstrap.go` contains `//go:embed packs/**`.
* **The Gap:** Once `internal/bootstrap/packs/core` is deleted, the `packs/` folder is empty. Go compilation fails immediately if a `//go:embed` pattern matches no files.
* **Required Change:** Mandate the absolute deletion of the `//go:embed` directive and its associated `embeddedBootstrapPacks` variable in the Core extraction slice.

### 3. Missing Hidden-Dependency Inventory Rows
* **The Gap:** The design's inventory of hidden dependencies misses key functions:
  * `publicSubpathForPack` in `internal/builtinpacks/registry.go` (which maps public aliases).
  * `requiredBuiltinPackNames` in `cmd/gc/embed_builtin_packs.go` (which requires `"maintenance"`).
  * `cmd/gc/prompt_test.go` (which reads Core prompts via relative on-disk paths).
* **Required Change:** Add these files and functions to the migration checklist with explicit dispositions. `prompt_test.go` must be refactored to read prompts via `core.PackFS` instead of relative disk paths.

### 4. Contradictory Fixture Path Contract
* **The Gap:** The design allows a "tiny compatibility embed under `internal/bootstrap/testdata/packs/core`" but also states that the source guard "rejects any path constant referencing `internal/bootstrap/packs/core`."
* **Required Change:** Use a synthetic test-only path like `packs/test-core` and name `test-core` explicitly. Update collision tests to use `Name: "core"` to preserve collision semantics while keeping the path synthetic.

### 5. `pruneStaleGeneratedPackFiles` Behavior on Retired Directories
* **The Gap:** Stale generated directories (`.gc/system/packs/maintenance` and `.gc/system/packs/gastown`) are ignored on startup, but `pruneStaleGeneratedPackFiles` normally deletes unneeded pack files.
* **Required Change:** Explicitly modify `pruneStaleGeneratedPackFiles` to skip these retired directories so user-modified files in former system packs are not accidentally deleted.

---

## Required Changes Checklist

1. **Delete Doctor Env-Mutation Helper:** Add `internal/doctor/implicit_import_cache_check.go` (`ensureBootstrapForDoctor`) to the Core extraction slice. Route implicit checks through standard packsource classifiers.
2. **Complete Deletion of Embed Comment/Variable:** Explicitly delete the `//go:embed packs/**` comment and `embeddedBootstrapPacks` variable from `internal/bootstrap/bootstrap.go`.
3. **Require Mock-Resiliency for Bindings:** Add an explicit test invariant requiring the loader/config layer to handle the complete absence of `gc.bindings` table.
4. **Refactor `prompt_test.go` to use `PackFS`:** Mandate that `cmd/gc/prompt_test.go` uses embedded `core.PackFS` rather than direct `os.ReadFile`.
5. **Pruning Exclusions:** Update `pruneStaleGeneratedPackFiles` to ignore retired directories.
6. **Unified Doctor Wording:** Update `implicit_import_cache_check.go` removal wording to: *"Maintenance is retired; Core supplies generic maintenance and public Gastown supplies Gastown-specific behavior"* instead of *"supplied implicitly"*.

---

## Reflective Questions

* **Why maintain `internal/bootstrap`?** After extracting Core, `internal/bootstrap` becomes a skeleton package doing little more than retired pruning. Should its remaining prunings move to `internal/systempacks` or `internal/packsource`, allowing us to delete `internal/bootstrap` entirely?
* **Deterministic Air-Gapped Testing:** How do we guarantee packcompat tests can verify remote public Gastown versions in fully air-gapped CI environments? Is a local vendor-caching fallback planned?
