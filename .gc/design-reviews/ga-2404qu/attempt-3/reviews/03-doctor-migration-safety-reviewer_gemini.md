# Sofia Khoury — DeepSeek V4 Flash (Independent Review, Iteration 3)

**Verdict:** block

**Persona focus:** Doctor fix idempotency, legacy import rewrite safety, custom data preservation, operator-safe diagnostics, cross-file consistency, missed edge cases, pattern drift, and architectural coherence. This iteration re-examines the design-after (updated 2026-06-05T20:30Z) and the current codebase, with full awareness of iteration-1 and iteration-2 findings. The focus is on evidence that prior blockers have been resolved, newly surfaced gaps, and cross-file inconsistencies that remain unaddressed.

---

## Top strengths

- The design-after now carries a substantive doctor fix safety contract (preflight-before-mutation, byte-identical healthy-city guarantee, scoped-edit-or-refuse, provenance-gated auto-rewrite, no-deletion of stale system/runtime pack dirs). These are the correct contractual invariants.
- The Core presence doctor avoids injecting `[imports.core]` into user manifests and instead repairs the generated system pack path, which is architecturally cleaner.
- Stale-directory preservation is correctly conservative: `.gc/system/packs/maintenance` and `.gc/runtime/packs/maintenance` are diagnosed and preserved, never deleted by `--fix`.
- The review-gated migration invariants (slice-level gates, behavior inventory, pinned public commit) are the right structural approach.
- The seven-slice strategy with per-slice gates is well-structured.

---

## Critical risks

### [Blocker] The import-state Fix still has no rollback for `pack.toml` and no cross-file atomicity — the design-after's failure-atomic contract remains unimplementable with the current code path

The design-after promises "failure-atomic across city.toml, rig pack.toml, lockfiles, and installed pack directories" (design.md:449-451) and "temp-file-plus-rename for each file" (:450). The codebase confirms the gap persists from iteration 1:

1. **`pack.toml` has no snapshot/restore.** `writeCityPackManifest` writes `pack.toml` through `fsys.WriteFileAtomic` (`cmd/gc/cmd_import.go:1207`), which provides per-file atomicity (temp + rename) but no pre-write snapshot. `writeCityAndRigSiteBindingsForEdit` snapshots `city.toml` and `.gc/site.toml` and restores on site-binding failure (`internal/config/site_binding.go:347-386`), but `pack.toml` is entirely outside that rollback path.

2. **The import-state Fix writes `pack.toml` first, then `city.toml`.** In `rewriteLegacyPublicPackImportsFS` (`import_state_doctor_check.go:257-313`), `pack.toml` is written at line 278-280 before `city.toml` is written at line 313. If the `city.toml` write fails, `pack.toml` is already mutated with no rollback.

3. **Post-manifest mutations are unprotected.** After both manifest writes succeed, the Fix calls `syncImports`, `writeImportLockfile`, and `installLockedImports` sequentially (`import_state_doctor_check.go:113-131`). A failure at any of these steps leaves manifests already rewritten. The design-after requires preflight to validate reachability, lockability, and installability before any manifest write, but the current code has no such preflight — `resolveWave1PublicPackImports` is a static map lookup (`defaultWave1PublicPackImports`, lines 72-91), not a network validation.

4. **`writeImportLockfile` is also per-file atomic without rollback.** It calls `fsys.WriteFileAtomic` (`internal/packman/lockfile.go:99`), which provides temp+rename but no cross-file coordination.

The iteration-2 review identified this as a blocker with three proposed resolutions: (a) snapshot all target manifests before any mutation and restore on any post-mutation failure, (b) defer all manifest writes until after lock/install validation succeeds, or (c) delegate all mutations to `gc import install`. The design-after has not adopted any of these. The design must specify one.

### [Blocker] TOML round-trip remains lossy for comments, unknown keys, and field ordering — the scoped-edit-or-refuse requirement has no implementation mechanism

Both write paths decode TOML into Go structs and re-encode:

- `writeCityPackManifest` re-encodes through `cityPackManifestBody` → `toml.NewEncoder(&buf).Encode(body)` (`cmd/gc/cmd_import.go:1194-1203`). Any `pack.toml` key not in `cityPackManifestBody` is silently dropped. The read path (`loadCityPackManifestFS`, `cmd_import.go:1129-1166`) uses `toml.Decode` into the `cityPackManifest` struct without capturing `Undecoded()`, so unknown keys are silently ignored on read.

- `writeCityConfigForEditFS` → `cfg.MarshalForWrite()` → `clone.Marshal()` calls `toml.Encode` on a `City` struct. Unknown keys and comments are lost on re-encoding.

The design-after (lines 455-457) says "If the existing parser/editor cannot preserve unrelated content, doctor must refuse the automatic fix with manual guidance instead of whole-file re-encoding." This conditional does not require implementing a scoped TOML editor, nor does it require refusing the fix. An implementer could satisfy the letter by adding golden tests that pass on round-trip for known struct fields while still dropping comments and unknown keys, because the test corpus doesn't include comments or unknown keys. The iteration-2 review called this out as requiring an unconditional decision. The design-after has not resolved it.

The design must either: (a) unconditionally require TOML content preservation with `Undecoded()` detection and refusal on unknown content, (b) implement a scoped TOML editor that preserves comments and unknown keys (with tests), or (c) unconditionally refuse auto-fix when the manifest contains any unknown content, routing to manual guidance.

### [Blocker] `legacyPublicPackForSource` uses path-suffix matching that cannot distinguish generated from edited directories — provenance detection is still underspecified

The current provenance detection in `legacyPublicPackForSource` (`import_state_doctor_check.go:216-244`) matches imports by path suffix against `.gc/system/packs/{gastown,maintenance}` and `examples/gastown/packs/{gastown,maintenance}`. It also matches any absolute path ending in those suffixes. This cannot distinguish:

1. A pristine generated `.gc/system/packs/gastown` from an operator-edited one.
2. A vendored fork living under `examples/gastown/packs/gastown` from the genuine embedded pack.
3. An operator's deliberate custom Core pin at a similar path from a redundant generated import.

The design-after requires "provenance-gated auto-rewrite" (lines 458-463) and routes non-matching cases to "diagnostic/manual." But the only detection mechanism is path-suffix matching, which is not provenance. All three iterations of this review have identified this gap. The design must specify a provenance mechanism: content hash comparison against known-generated content, install-manifest provenance, or explicit clean-tree verification. Without it, the doctor fix can silently remove an operator's deliberate local import that happens to live under a matching path.

### [Blocker] Doctor check ordering and dependency — Core presence validation is not a guaranteed precondition for import-state fix

The design-after says redundant Core imports should be removed only after confirming Core is present (lines 434-437), and the Core doctor must materialize and revalidate Core (lines 412-417). But the doctor framework (`internal/doctor/doctor.go`) runs checks in registration order with no dependency mechanism. The `buildDoctorChecks` function in `cmd/gc/cmd_doctor.go` registers `importStateDoctorCheck` after the config-dependent block but before any explicit Core presence check.

If Core materialization fails (e.g., embedded pack FS error), the import-state fix could still remove a redundant Core import, leaving the city with no Core provenance at all. The design must specify either: (a) a check-dependency mechanism in the doctor framework, (b) a precondition check inside `importStateDoctorCheck.Fix()` that verifies Core provenance before removing Core imports, or (c) an explicit ordering contract with a test verifying the registration order and a precondition assertion.

---

## Major risks

### [Major] `MaterializeBuiltinPacks` does `os.RemoveAll` + regenerate for the Core pack directory — no swap pattern for live controllers

`MaterializeBuiltinPacks` (`cmd/gc/embed_builtin_packs.go:53-66`) iterates all builtin packs and for each one calls `materializeFS`, which creates directories and files. Inside `materializeFS` (`embed_builtin_packs.go:287-328`), `pruneStaleGeneratedPackFiles` removes stale files, but `MaterializeSyntheticRepo` (`internal/builtinpacks/registry.go:140-148`) does `os.RemoveAll(dst)` before writing the synthetic repo. A controller loading config concurrently can observe a missing or half-populated `.gc/system/packs/core` mid-fix. The design-after requires "byte-identical on healthy runs" but does not specify a swap pattern (write to temp dir + rename) for Core materialization, which would prevent this read-consistency gap.

### [Major] `.gc/runtime/packs/maintenance/` state migration is still unaddressed

The design mentions `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` as stale directories, but the codebase has runtime state under `.gc/runtime/packs/maintenance/`:

- `cmd/gc/jsonl_archive_doctor_check.go:60,71,97` hardcodes `filepath.Join(runtime, "packs", "maintenance")` in both `resolveStateFile()` and `resolveArchiveRepo()`.
- `examples/gastown/packs/maintenance/assets/scripts/jsonl-export.sh:22` hardcodes `PACK_STATE_DIR="${GC_PACK_STATE_DIR:-${GC_CITY_RUNTIME_DIR:-$CITY/.gc/runtime}/packs/maintenance}"`.

After migration, new cities write state to `.gc/runtime/packs/core/` via `GC_PACK_STATE_DIR`, but existing cities have JSONL archives and state files under `.gc/runtime/packs/maintenance/`. The design never specifies whether existing state is migrated in-place, read through a compatibility path, or reported with manual instructions. The iteration-2 review identified this as a data-continuity issue. The `jsonl_archive_doctor_check.go` has no legacy fallback for `.gc/runtime/packs/maintenance/` — it only has a legacy fallback for `.gc/jsonl-export-state.json` at the city root.

### [Major] `[[patches.agent]] name = "dog"` cross-pack resolution is unspecified

The Gastown pack currently patches `maintenance.dog` via `[[patches.agent]] name = "dog"` in `examples/gastown/packs/gastown/pack.toml:29`. After migration, `dog` comes from Core, which is a required system pack injected by `builtinPackIncludes` — not an explicit import. The Gastown pack does not and cannot declare `[imports.core]` because Core is implicit. The design does not address how `[patches.agent]` resolves when the target agent comes from a required-but-not-imported system pack. This affects the primary Gastown use case.

### [Major] `requiredBuiltinPackNames` hardcodes `"maintenance"` but the design requires retiring it

`requiredBuiltinPackNames` (`cmd/gc/embed_builtin_packs.go:237`) currently returns `["core", "maintenance", ...]`. The design requires retiring Maintenance as a standalone pack, but the function still hardcodes `"maintenance"`. The migration slice that removes it must also update `requiredBuiltinPackSet`, `builtinPackIncludes`, `unusableRequiredBuiltinPackNames`, and the test assertions that verify `"maintenance"` membership. The iteration-2 bootstrap review (B5) identified this, and it remains unresolved in the design.

### [Major] Air-gap behavior is unspecified

On air-gapped hosts, `resolveWave1PublicPackImports` returns a static map (`defaultWave1PublicPackImports`) without network validation. The design-after requires preflight to validate reachability, lockability, and installability before any manifest write (lines 445-448), but the current code has no preflight. When the public Gastown pack is unreachable, the import-state fix should leave manifests byte-identical and emit manual guidance. The design must specify this as a first-class test case.

### [Major] `GastownCity` function directly constructs config instead of loading it — unclear if it needs the scanner test

`GastownCity` (`internal/config/config.go:3712`) directly constructs a `City` struct with `PublicGastownPackSource` and `PublicGastownPackVersion`. This is a factory function, not a `config.Load` path. The design's "production `config.Load*` scanner" must clarify whether `GastownCity` is a production path that needs the scanner test or a factory that's exempt. Current `gc init` calls this factory, so it is a production path — but it never goes through `loadCityConfigWithBuiltinPacks`, meaning it sidesteps the Core provenance assertion the design requires.

### [Major] Concurrency protection for `gc doctor --fix` is still unspecified

All three iterations have identified this. The design-after requires "byte-identical on healthy runs" and "failure-atomic across … installed pack directories" but does not specify whether `gc doctor --fix` can run concurrently with a controller, whether it should refuse mutating fixes when the controller is active, or what serialization protects `.gc/system/packs/core` during `MaterializeBuiltinPacks`. The current code has no locking. `MaterializeBuiltinPacks` is called from `ensureBuiltinPacksReadyForConfigLoad` on every `gc` startup, meaning a doctor fix and a controller startup can race on Core materialization.

---

## Minor risks

### [Minor] `publicSubpathForPack` in `internal/builtinpacks/registry.go` maps both `"gastown"` and `"maintenance"` as public packs

After Maintenance is retired, the `"maintenance"` case in `publicSubpathForPack` must be removed. The design mentions retiring Maintenance from the registry but does not call out this specific function.

### [Minor] `GastownCity` still references `"mayor"` in its agent list

`GastownCity` (`internal/config/config.go:3700-3710`) defines agents with `{Name: "mayor", PromptTemplate: "prompts/mayor.md"}`. The AGENTS.md principle states "ZERO hardcoded roles" and "If a line of Go references a specific role name, it's a bug." While `GastownCity` is a factory for Gastown-specific configuration and the `mayor` reference comes from the Gastown pack's expected agent set, this still hardcodes a role name in Go source. The design should clarify whether this is an acceptable pack-specific factory or whether it should be driven from pack configuration.

### [Minor] `legacyPublicPackForSource` matches absolute paths by suffix — too broad

The function matches any absolute path ending in `/.gc/system/packs/gastown` or `/examples/gastown/packs/gastown` (and the maintenance equivalents). This would match a user's custom directory at `/home/user/projects/gastown/.gc/system/packs/gastown` even if it's not related to the builtin packs. A canonical-path comparison against the resolved city directory would be more precise.

### [Minor] `defaultWave1PublicPackImports` hardcodes only `"gastown"` and `"maintenance"` — no extensibility

The `defaultWave1PublicPackImports` function (`import_state_doctor_check.go:72-91`) is a static switch statement that only handles `"gastown"` and `"maintenance"`. If future packs need similar migration, this function must be extended. The design should note this as a maintenance point or generalize the mechanism.

### [Minor] The `writeCityPackManifest` function re-encodes TOML without preserving `DefaultRigImportOrder` in-band

`loadCityPackManifestFS` separately reads `DefaultRigImportOrder` from the file (via `config.LoadRootPackDefaultRigImports`) because the TOML struct doesn't have a corresponding field. `writeCityPackManifest` then passes `manifest.DefaultRigImportOrder` through `writeOrderedDefaultRigImports` which appends a `[defaults.rig.import_order]` section. If an operator has manually edited the import order in `pack.toml` with inline comments or formatting, the rewrite will lose those edits. This is a specific case of the general TOML-lossiness blocker.

---

## Required changes

1. **Resolve the cross-file atomicity gap.** The design must adopt one of: (a) snapshot all target manifests before any mutation and restore on any post-mutation failure, (b) defer all manifest writes until after lock/install validation succeeds (simplest, directly matches the preflight contract), or (c) delegate all mutations to `gc import install` which already owns the lockfile+install pipeline. This was identified in iterations 1 and 2 and remains unresolved.

2. **Resolve the TOML round-trip lossiness.** Unconditionally require one of: (a) `Undecoded()` detection with refusal to auto-fix when unknown content is present, (b) a scoped TOML editor with comment/unknown-key preservation tests, or (c) refusal to auto-fix any manifest with content outside the known struct schema. This was identified in iterations 1 and 2 and remains unresolved.

3. **Specify a provenance mechanism for `legacyPublicPackForSource`.** Replace path-suffix matching with content hash comparison against known-generated manifests, install-lock provenance, or explicit clean-tree verification. Auto-rewrite should only proceed when provenance is established; all other cases should route to manual guidance.

4. **Make Core presence validation a guaranteed precondition for import-state fix.** Either add a check-dependency mechanism to the doctor framework, require `importStateDoctorCheck.Fix()` to verify Core provenance itself before removing Core imports, or specify an explicit ordering contract with a test.

5. **Specify a materialization swap pattern for Core.** Core materialization should write to a temp directory and rename, not `os.RemoveAll` + regenerate, to prevent a read-consistency gap under a live controller.

6. **Specify `.gc/runtime/packs/maintenance/` state continuity.** Specify whether existing runtime state is migrated in-place, read through a compatibility path, or reported with manual instructions. Document what happens to `jsonl_archive_doctor_check.go`'s hardcoded `packs/maintenance` path. Add a legacy fallback path during the transition.

7. **Specify `[[patches.agent]]` resolution for required-but-not-imported system packs.** Either enable patches to target agents from implicitly-included system packs, or require Gastown to define its own `dog` agent.

8. **Specify concurrency protection for `gc doctor --fix`.** Either refuse mutating fixes when the controller is active, or define a file lock that concurrent doctor/import-install processes respect.

9. **Specify air-gap behavior.** On unreachable public Gastown, leave manifests byte-identical and emit manual guidance. This should be a first-class test case.

10. **Remove `"maintenance"` from `publicSubpathForPack`, `requiredBuiltinPackNames`, and related hardcoded lists** as part of the specified migration slice, and call out each function that must change.

---

## Questions

1. Should `gc doctor --fix` refuse to mutate imports when the controller is active, or should it define a concurrency contract for concurrent reads?
2. What existing artifact can prove a `.gc/system/packs/maintenance` directory is unmodified? Is there a generated manifest or content hash, or should the design require one?
3. Is `PublicGastownPackVersion` guaranteed to always be a commit SHA? If it could be a branch name or tag, the "immutable version" preflight check needs to resolve it to a commit first.
4. Should the import-state Fix own a multi-file transaction, or should it delegate all pack/import mutations to an existing `gc import install` path with proven atomicity?
5. What happens to existing `packs/maintenance` runtime state (JSONL archives, storm counts) after the migration? Is it migrated, read through a compatibility path, or reported as a manual step?
6. How does `[[patches.agent]] name = "dog"` resolve when `dog` is in Core (a required system pack) but Gastown does not explicitly import Core?
