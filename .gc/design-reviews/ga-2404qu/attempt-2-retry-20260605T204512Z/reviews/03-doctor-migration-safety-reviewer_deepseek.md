# Sofia Khoury — DeepSeek V4 Flash (Independent Review)

**Verdict:** block

**Persona focus:** Doctor fix idempotency, legacy import rewrite safety, custom data preservation, operator-safe diagnostics. This review is grounded in a close reading of the current `cmd/gc/import_state_doctor_check.go`, `cmd/gc/embed_builtin_packs.go`, `internal/builtinpacks/registry.go`, `internal/doctor/stale_local_pack_dir_check.go`, and the config/pack write paths.

## Top strengths

- The design-after adds a substantive doctor fix safety contract (doctor fix safety contract section) that was absent from design-before: preflight-before-mutation, byte-identical healthy-city guarantee, scoped-edit-or-refuse rule, provenance-gated auto-rewrite, and explicit non-deletion of stale system/runtime pack directories. These are the right contractual invariants and map directly to the failure modes visible in the current codebase.
- The Maintenance Retirement Runtime Table is new and critical. It maps every affected runtime surface to target behavior and required proof. This is precisely the kind of per-surface contract needed to prevent silent regressions during the Maintenance→Core fold.
- The release compatibility matrix (5-cell table) is well-constructed and covers the version-skew interactions that matter. The "rollback from new to old" row explicitly names the risk that doctor-mutated manifests may not be readable by old binaries.

## Critical risk

### [Blocker] The import-state Fix mutates `pack.toml` before validating that post-mutation state is installable — with no rollback for `pack.toml`

Evidence from the current code (`cmd/gc/import_state_doctor_check.go:102–132`):

```
1. resolveWave1PublicPackImports   — static map lookup, no network/validation
2. rewriteLegacyPublicPackImportsFS — writes pack.toml and/or city.toml to disk
3. Re-read imports from mutated manifests
4. syncImports                      — first network call, can fail
5. writeImportLockfile              — can fail
6. installLockedImports             — can fail
```

If steps 4, 5, or 6 fail, `pack.toml` (and possibly `city.toml`) are already rewritten with no rollback path. The existing codebase has a snapshot/restore mechanism for `city.toml` via `internal/config/site_binding.go:358–378` (`writeCityAndRigSiteBindingsForEdit` snapshots `city.toml` + `.gc/site.toml`, restores on site-binding failure). But `pack.toml` is written through `writeCityPackManifest` → `fsys.WriteFileAtomic` with no snapshot.

The design-after requires "preflight must run before any mutation" and "fixes must be failure-atomic across city.toml, rig pack.toml files, lockfiles, and installed pack directories." But the current code path violates both requirements, and the design does not specify the concrete implementation change that closes this gap. The design's "temp-file-plus-rename for each file" phrasing addresses per-file atomicity (which the codebase already has via `WriteFileAtomic`), not cross-file transactional atomicity, which is the actual problem.

**Cross-file consistency issue:** `city.toml` writes get snapshot+restore. `pack.toml` writes get none. The import-state fix touches both. An inconsistent intermediate state is possible where `pack.toml` has been rewritten to remove `maintenance` and add `gastown` public source, but `city.toml` is still in its old form (or vice versa).

**Required change:** Restructure the import-state Fix so that one of these holds:
- **(a) Preflight-first:** Validate network reachability, installability, and lockability before any manifest mutation. If preflight fails, leave all manifests byte-identical. This eliminates the need for multi-file rollback in the common case.
- **(b) Snapshot-all:** Extend the `writeCityAndRigSiteBindingsForEdit` snapshot/restore pattern to cover `pack.toml` as well. On any post-mutation failure, restore all three files (`city.toml`, `pack.toml`, `.gc/site.toml`).

Option (a) is simpler and matches the preflight contract. Option (b) matches the existing pattern but is more complex. The design must specify which approach and require a test that injects failure at each step (4, 5, 6) and verifies byte-identical rollback.

### [Blocker] TOML round-trip is lossy for comments, unknown keys, and field ordering — the scoped-edit requirement has no implementation path and no detection mechanism

Evidence from the codebase:

1. `writeCityPackManifest` (`cmd/gc/cmd_import.go:1172–1213`) encodes a `cityPackManifestBody` struct via `toml.NewEncoder(&buf).Encode(body)`. Any `pack.toml` key not in this struct is silently dropped on write. Comments are lost. Field ordering follows struct field order, not the original file order.

2. `writeCityConfigForEditFS` → `cfg.MarshalForWrite()` → `clone.Marshal()` calls `toml.Encode` on a `City` struct. Unknown keys and comments are lost.

3. The read paths (`loadCityPackManifestFS`, `loadCityConfigForEditFS` → `config.Load`) use `toml.Decode` into structs. The config package *does* have `CheckUndecodedKeys` (`internal/config/undecoded.go:27`) and `parsePackConfigWithMetadata` captures `toml.MetaData`, but neither the import-state doctor check nor `loadCityPackManifestFS` uses them to detect unknown keys before writing.

4. The design-after says "If the existing parser/editor cannot preserve unrelated content, doctor must refuse the automatic fix with manual guidance instead of whole-file re-encoding." This is conditional — it does not *require* detecting unknown content, and does not *require* refusing the fix when unknown content is present. An implementer could satisfy the letter by adding round-trip golden tests for known struct fields while still dropping comments and unknown keys, because the test corpus does not include comments or unknown keys.

**Required change:** The design must unconditionally require:
1. **Detection before mutation:** Before rewriting any manifest, parse it with `toml.Decode` + `Undecoded()` to detect unknown keys. If any exist, refuse the fix and emit manual guidance naming the unknown keys. The `CheckUndecodedKeys` function already exists in `internal/config/undecoded.go` — the design should require the import-state Fix to call it (or equivalent) before mutation.
2. **Comment preservation or refusal:** If the TOML file contains comments (detectable via a line-scanning pre-pass or by comparing round-tripped content), refuse the fix and emit manual guidance. The design should not defer this to "if the parser cannot preserve" — it must require detection and refusal when comments are present.
3. **Golden tests with comments and unknown keys:** The test suite must include `pack.toml` and `city.toml` fixtures containing comments, unknown `[section]` tables, and non-standard field ordering. These fixtures must prove that `gc doctor --fix` leaves them byte-identical when no legacy imports exist, and refuses the fix with explicit manual guidance when legacy imports exist alongside custom content.

### [Blocker] Legacy path-suffix matching in `legacyPublicPackForSource` cannot distinguish generated directories from operator forks

Evidence from `cmd/gc/import_state_doctor_check.go:216–236`:

```go
func legacyPublicPackForSource(cityPath, source string) (string, bool) {
    // ...
    for _, pack := range []string{"gastown", "maintenance"} {
        if source == ".gc/system/packs/"+pack || source == "examples/gastown/packs/"+pack {
            return pack, true
        }
    }
    // Also matches any path *ending* with those suffixes
    path = filepath.ToSlash(filepath.Clean(path))
    for _, pack := range []string{"gastown", "maintenance"} {
        for _, suffix := range []string{"/.gc/system/packs/" + pack, "/examples/gastown/packs/" + pack} {
            if strings.HasSuffix(path, suffix) {
                return pack, true
            }
        }
    }
    return "", false
}
```

This function classifies imports as "legacy" based on path suffix matching alone. It cannot distinguish:
- A `.gc/system/packs/maintenance` directory that was generated by `MaterializeBuiltinPacks` (safe to auto-rewrite) from one that an operator has customized (unsafe to auto-rewrite without confirmation).
- An `examples/gastown/packs/gastown` path that points to the embedded Gas City checkout (auto-rewrite is correct) from one that points to a fork or local experiment at a similar path (auto-rewrite would destroy intentional configuration).

The design-after requires "legacy local Gastown imports are auto-rewritten only when provenance matches known generated/example paths," but provides no mechanism for establishing provenance. Path-suffix matching is not provenance.

**Required change:** The design must specify a provenance detection mechanism. Options:
- **(a) Content-hash verification:** Before auto-rewriting a legacy import, verify that the referenced directory's file set and content match the known generated manifest (similar to `validatePackFiles` in `internal/builtinpacks/registry.go`). If the content differs, treat it as a custom fork and emit manual guidance.
- **(b) Strict path matching only:** Only auto-rewrite imports whose source is exactly `.gc/system/packs/{gastown,maintenance}` (relative to city root) or the embedded `Repository//` subpath form. All other matches should be diagnostic/manual.

Option (a) is more robust but requires maintaining a content manifest for each generated pack. Option (b) is simpler and covers the vast majority of cases. The design should specify at least one and require tests proving custom-fork detection.

## Major risk

### [Major] The `rewriteLegacyPublicPackImportMap` function has a TOCTOU race between reading and writing manifests — no file locking

The import-state Fix reads and writes `pack.toml` and `city.toml` through separate read-modify-write cycles:
1. `loadCityPackManifestFS(fs, cityPath)` reads `pack.toml`
2. `rewriteLegacyPublicPackImportMap(...)` mutates the in-memory imports map
3. `writeCityPackManifest(fs, cityPath, manifest)` writes `pack.toml`

If another process (controller, `gc import`, another `gc doctor --fix`) modifies `pack.toml` between step 1 and step 3, those changes are lost. The `WriteFileAtomic` per-file atomicity does not prevent this — it only prevents partial writes, not lost updates.

The existing `writeCityAndRigSiteBindingsForEdit` has the same TOCTOU vulnerability for `city.toml`, but that function is typically called by single-operator commands (`gc agent add`, `gc rig add`). `gc doctor --fix` is more likely to run concurrently with a controller or warmup check.

The design-after requires "byte-identical on healthy cities, including repeated or concurrent runs with a controller active" but does not specify a concurrency contract or locking mechanism.

**Required change:** The design must specify one of:
- **(a)** `gc doctor --fix` refuses to mutate imports when the controller is active (detectable via PID file or process table query per AGENTS.md "No status files — query live state" principle).
- **(b)** A file-locking protocol for `pack.toml` and `city.toml` that prevents concurrent modification (e.g., `flock` on the file before read, released after write).
- **(c)** A compare-and-swap approach: read the file, compute changes, re-read before write, verify the file hasn't changed, write if unchanged, retry or fail otherwise.

The design should also require a concurrent-doctor test: run `gc doctor --fix` concurrently with a controller active and verify both byte-identical healthy-city output and no lost updates.

### [Major] The `refusing to overwrite existing import` guard is insufficient — it only checks the target binding, not other user-added imports

In `rewriteLegacyPublicPackImportMap` (line 357–359):

```go
if existing, ok := imports[binding]; ok && !sameImport(existing, target.Import) {
    if _, legacyTarget := legacyPublicPackForSource(cityPath, existing.Source); !legacyTarget {
        return false, nil, fmt.Errorf("refusing to overwrite existing %q import with source %q", binding, existing.Source)
    }
}
```

This guard prevents overwriting a *non-legacy* import at the target binding key (e.g., `gastown`). But it does not protect against:
1. An operator who has added additional imports that the rewrite function doesn't know about — these are preserved only because the function doesn't touch keys outside its target set, but this is implicit, not tested.
2. An operator who has a custom import at the *same* key as a legacy import (e.g., `[imports.maintenance]` pointing to a custom fork) — the `legacyPublicPackForSource` path-suffix match would classify this as legacy and delete it.
3. The `import_order` array manipulation in `replaceImportOrderWithTargets` — if the operator has reordered or added entries to the default rig import order, those may be silently lost.

**Required change:** The design should require tests proving that:
1. User-added imports outside the legacy set are preserved through a fix.
2. A custom-fork `[imports.maintenance]` import (non-standard source) is detected and treated as manual/diagnostic, not auto-removed.
3. Custom `default_rig_import_order` entries outside the legacy set are preserved.

### [Major] Core materialization uses `os.RemoveAll` + regenerate — creates a read-consistency gap under a live controller

`MaterializeBuiltinPacks` (line 61–80) iterates all embedded packs, writes each to `.gc/system/packs/{name}`, then prunes stale files. For required packs (Core, maintenance), it does *not* preserve operator edits and always refreshes. The materialization calls `pruneStaleGeneratedPackFiles`, which removes files not in the embedded manifest.

But the critical issue is: `pruneStaleGeneratedPackFiles` removes files *synchronously* before new files are written for the *next* pack in the iteration. If the iteration order is Core → bd → dolt → maintenance → gastown, and a controller is reading `.gc/system/packs/core` files while `MaterializeBuiltinPacks` is running for `bd`, the Core pack is already in its final state. However, if `MaterializeBuiltinPacks` is interrupted between `os.RemoveAll` and regeneration for a *single* pack, that pack's directory is empty or partially written.

The current code uses `fsys.WriteFileIfContentOrModeChangedAtomic` for individual files, which is per-file atomic. But the overall `MaterializeBuiltinPacks` is not atomic at the directory level — removing stale files and writing new ones are separate operations.

The design-after should specify a materialization swap pattern for Core: write to a temp directory (e.g., `.gc/system/packs/core.tmp.<pid>`) and rename the entire directory, rather than writing individual files and pruning individually. This eliminates the window where `.gc/system/packs/core` contains a mix of old and new files.

## Missing evidence

1. **No fault-injection test plan.** The design requires failure-atomic behavior but does not specify how to inject failures at each step of the fix pipeline to verify rollback. The design should require at minimum: (a) network failure after manifest mutation, (b) lock write failure, (c) install failure, and (d) concurrent doctor + controller scenarios.

2. **No provenance detection specification.** The design requires provenance-matched auto-rewrites but provides no mechanism for establishing provenance. The current code uses path-suffix matching which cannot distinguish generated from edited directories.

3. **No test for concurrent doctor + controller runs.** The design requires byte-identical healthy-city behavior including concurrent runs but does not specify or require tests for `gc doctor --fix` running concurrently with a controller that is reading `.gc/system/packs/core`.

4. **No test for partial-failure rollback.** The import-state Fix touches `pack.toml`, `city.toml`, `.gc/site.toml`, the lockfile, and installed pack directories. No test verifies that a failure at any post-mutation step leaves all files byte-identical.

5. **No specification of import-state check dependency on Core presence check.** The doctor framework is registration-order-based with no dependency mechanism. If the Core presence check runs after the import-state check, the import-state fix could remove a legacy Maintenance import before Core is confirmed present. The design should specify the execution ordering or add a check dependency mechanism.

6. **No specification of air-gap behavior.** If `PublicGastownPackSource` is unreachable (air-gapped host), `syncImports` will fail after manifests are already rewritten. The design must specify: on air-gapped hosts, `gc doctor --fix` must leave manifests byte-identical and emit explicit manual guidance including the exact lines the operator should change.

7. **The `StaleLocalPackDirCheck` exists but is not integrated with the migration.** The existing `internal/doctor/stale_local_pack_dir_check.go` warns about `packs/<binding>/` directories alongside remote imports. The design should specify whether `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` will be covered by this check or by a new check. The current check targets `packs/` (not `.gc/system/packs/`), so it will not detect stale system pack directories.

## Required changes

1. **Defer manifest mutation until after preflight.** Restructure the import-state Fix so that network reachability, lockability, and installability are verified before any manifest write. If preflight fails, leave all manifests byte-identical and emit manual guidance. This eliminates the need for multi-file rollback in the common case.

2. **Add snapshot/restore for pack.toml, or defer all writes.** Extend the `writeCityPackManifest` path to snapshot and restore like `writeCityAndRigSiteBindingsForEdit` does for `city.toml`, or defer all manifest writes to after lock/install validation succeeds.

3. **Unconditionally require TOML content preservation or fix refusal.** Remove the conditional "if the parser cannot preserve" language. Either implement detection of unknown keys (using the existing `CheckUndecodedKeys`) and comments (via a line-scanning pre-pass), and refuse the fix when they are present, or implement a scoped TOML editor with preservation tests.

4. **Replace path-suffix matching with exact canonical path matching plus content verification.** Only auto-rewrite imports whose source is exactly `.gc/system/packs/{gastown,maintenance}` (relative to city) or the embedded `Repository//` form, and only when the directory content matches the known generated manifest. All other matches should be diagnostic/manual.

5. **Specify a concurrency contract for doctor + controller.** Either refuse to mutate when the controller is active, add file locking, or use compare-and-swap. Require a concurrent-doctor test.

6. **Specify import-state check dependency on Core presence check.** Either add a check dependency mechanism to the doctor framework, or require the import-state Fix to call `MaterializeBuiltinPacks` itself before removing redundant Core imports.

7. **Specify air-gap behavior.** On unreachable public Gastown, leave manifests byte-identical and emit manual guidance. This should be a first-class test case.

8. **Specify a materialization swap pattern for Core.** Core materialization should write to a temp directory and rename, not write individual files and prune individually, to avoid a read-consistency gap under a live controller.

9. **Integrate stale system pack directories with the existing StaleLocalPackDirCheck or add a new check.** The current check covers `packs/` but not `.gc/system/packs/`. The design should specify the check surface and required tests.

## Questions

1. Should `gc doctor --fix` refuse to mutate imports when the controller is active, or should it define a concurrency contract for concurrent reads?
2. What existing artifact can prove a `.gc/system/packs/maintenance` directory is unmodified? Is there a generated manifest or content hash, or should the design require one?
3. Is `PublicGastownPackVersion` guaranteed to always be a `sha:` commit hash? If it could be a branch name or tag, the "immutable version" preflight check needs to resolve it to a commit.
4. Should the import-state Fix own a multi-file transaction, or should it delegate to `gc import install` which already has atomicity for lockfile + install?
5. Should the `cityPackManifest` struct be extended to capture `toml.MetaData` for unknown-key detection before mutation, matching the pattern already used by `parsePackConfigWithMetadata`?
