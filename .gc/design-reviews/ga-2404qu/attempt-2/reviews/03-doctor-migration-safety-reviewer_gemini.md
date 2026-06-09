# Sofia Khoury — DeepSeek V4 Flash (Independent Review, Iteration 2)

**Verdict:** block

**Persona focus:** Doctor fix idempotency, legacy import rewrite safety, custom data preservation, operator-safe diagnostics. This review cross-references the design-after against the current codebase to surface inconsistencies, assumptions, and edge cases that may be underweighted in the other reviews. Iteration 2 re-examines the design-after (updated 2026-06-05T20:30Z) and the current codebase, with awareness of the iteration-1 findings.

---

## Top strengths

- The design-after now carries a substantive doctor fix safety contract (preflight-before-mutation, byte-identical healthy-city guarantee, scoped-edit-or-refuse, provenance-gated auto-rewrite, no-deletion of stale system/runtime pack dirs). These are the correct contractual invariants.
- Stale-directory preservation is correctly conservative: `.gc/system/packs/maintenance` and `.gc/runtime/packs/maintenance` are diagnosed and preserved, never deleted by `--fix`.
- The Core presence doctor avoids injecting `[imports.core]` into user manifests and instead repairs the generated system pack path, which is architecturally cleaner.
- The review-gated migration invariants (slice-level gates, behavior inventory, pinned public commit) are the right structural approach to phased rollout.

---

## Critical risks

### [Blocker] The import-state Fix has no rollback for `pack.toml` and no cross-file atomicity — contradicting the design-after's failure-atomic contract

The design-after promises "failure-atomic across city.toml, rig pack.toml, lockfiles, and installed pack directories" (`design.md:449-451`) and "temp-file-plus-rename for each file" (`:450`). The codebase confirms the gap:

1. **`pack.toml` has no snapshot/restore.** `writeCityPackManifest` writes `pack.toml` through `fsys.WriteFileAtomic` (`cmd/gc/cmd_import.go:1200-1209`), which provides per-file atomicity (temp + rename) but no pre-write snapshot. `writeCityConfigForEditFS` → `WriteCityAndRigSiteBindingsForEdit` snapshots `city.toml` and `.gc/site.toml` and restores on site-binding failure (`internal/config/site_binding.go:368-386`), but `pack.toml` is entirely outside that rollback path.

2. **The import-state Fix writes `pack.toml` first, then `city.toml`.** In `rewriteLegacyPublicPackImportsFS` (`import_state_doctor_check.go:257-313`), `pack.toml` is written at line 278-280 before `city.toml` is written at line 313. If the `city.toml` write fails, `pack.toml` is already mutated with no rollback.

3. **Post-manifest mutations are unprotected.** After both manifest writes succeed, the Fix calls `syncImports`, `writeImportLockfile`, and `installLockedImports` sequentially (`import_state_doctor_check.go:113-131`). A failure at step 4, 5, or 6 leaves manifests already rewritten. The design-after requires preflight to validate reachability, lockability, and installability before any manifest write, but the current code has no such preflight — `resolveWave1PublicPackImports` is a static map lookup (`defaultWave1PublicPackImports`, lines 72-91), not a network validation.

4. **`writeImportLockfile` is also per-file atomic without rollback.** It calls `fsys.WriteFileAtomic` (`internal/packman/lockfile.go:99`), which provides temp+rename but no cross-file coordination with the manifest writes.

The design-after's "failure-atomic" wording implies all-or-nothing semantics, but the implementation delivers independent per-file writes with no journal, no staging, and no rollback for `pack.toml`. The design must specify one of: (a) snapshot all target manifests before any mutation and restore on any post-mutation failure, (b) defer all manifest writes until after lock/install validation succeeds, or (c) delegate all mutations to `gc import install` which already owns the lockfile+install pipeline. Option (b) is simplest and directly matches the preflight contract the design-after already describes.

### [Blocker] TOML round-trip is lossy for comments, unknown keys, and field ordering — the scoped-edit-or-refuse requirement has no implementation mechanism

Both write paths decode TOML into Go structs and re-encode:

- `writeCityPackManifest` re-encodes through `cityPackManifestBody` → `toml.NewEncoder(&buf).Encode(body)` (`cmd/gc/cmd_import.go:1194-1203`). Any `pack.toml` key not in `cityPackManifestBody` is silently dropped.
- `writeCityConfigForEditFS` → `cfg.MarshalForWrite()` → `clone.Marshal()` calls `toml.Encode` on a `City` struct. Unknown keys and comments are lost on re-encoding.
- The read paths (`loadCityPackManifestFS`, `loadCityImportManifestFS`) both use `toml.Decode` into structs without capturing `Undecoded()`, so unknown keys are silently ignored on read.

The design-after (lines 455-457) says "If the existing parser/editor cannot preserve unrelated content, doctor must refuse the automatic fix with manual guidance instead of whole-file re-encoding." This is a conditional: it does not require implementing a scoped TOML editor, nor does it require refusing the fix when loss is possible. An implementer could satisfy the letter by adding golden tests that pass on round-trip for known struct fields while still dropping comments and unknown keys, because the test corpus doesn't include those cases.

**Required change:** The design must unconditionally require either (a) a comment/unknown-key-preserving scoped TOML editor with golden tests that assert unknown-field survival on operator-authored `city.toml`/`pack.toml`, or (b) refusal to auto-fix when the file contains out-of-schema content (detected via `toml.Decode` with `Undecoded()` capture). The current conditional "if the parser cannot preserve" leaves the door open for an implementer to skip the check entirely.

### [Blocker] Provenance detection for auto-rewrite vs. manual-only is path-suffix matching, which cannot distinguish generated from edited directories

The design requires provenance-matched auto-rewrites (lines 458-459) and specifies that operator forks, edited local packs, and custom public sources are diagnostic/manual-only (lines 460-463). But the codebase's only provenance detector is `legacyPublicPackForSource` (`import_state_doctor_check.go:216-244`), which matches by path suffix:

- `.gc/system/packs/{gastown,maintenance}` — matches any import whose source resolves to these relative paths
- `examples/gastown/packs/{gastown,maintenance}` — matches any import with these relative paths
- Absolute paths ending in `/` plus those suffixes — matches any absolute path with the same tail

This cannot distinguish:
- A pristine `.gc/system/packs/gastown` generated by `MaterializeBuiltinPacks` from an operator-edited version with custom scripts, patches, or prompt overrides.
- A deliberate custom Core import at `.gc/system/packs/core` from a generated one.
- A vendored Gastown fork living under `examples/gastown/packs/gastown` (a common pattern for local development).
- Any `.gc/system/packs/maintenance` directory where the operator has added site-specific doctor checks or formulas.

The design must specify a provenance detection mechanism. Options include: (a) a content hash or synthetic manifest installed by `MaterializeBuiltinPacks` that identifies generated content, (b) requiring the import source to match the exact `Repository//Subpath` or `PublicRepository//Subpath` form before auto-rewriting, (c) a `.gc-provenance` marker file inside generated pack directories, or (d) requiring `Undecoded()` capture on the manifest to detect operator edits before rewriting.

---

## Major risks

### [Major] The import-state Fix has no concurrency protection and can run simultaneously with the controller or a second doctor invocation

`gc doctor --fix` has no lock, mutex, or controller-active gate for mutating checks. The doctor loop (`internal/doctor/doctor.go:68-86`) runs checks sequentially and fixes them one at a time, but:

- The controller may be reading `.gc/system/packs/core` or other pack state while `MaterializeBuiltinPacks` is rewriting it. `MaterializeBuiltinPacks` does `os.RemoveAll(dst)` then regenerates (`embed_builtin_packs.go:68-72`), creating a window where Core is empty or partially populated.
- Two concurrent `gc doctor --fix` processes could both enter `rewriteLegacyPublicPackImportsFS` and both rewrite `pack.toml` and `city.toml`, with interleaved writes.
- The `builtinPackRefreshCache` sync.Map only prevents redundant materialization within a single process; it doesn't protect against concurrent processes.

The design-after must specify either: (a) `gc doctor --fix` refuses mutating fixes when the controller is active (using `IsControllerRunning` which already probes the flock), or (b) a file lock or staging protocol that concurrent readers and writers can safely tolerate, or (c) atomic Core-dir swap (materialize to temp dir + rename) instead of `os.RemoveAll` + regenerate.

### [Major] Core materialization under a live controller has a read-consistency gap

`MaterializeBuiltinPacks` (`embed_builtin_packs.go:62-77`) calls `materializeFS` for each pack, which individually writes files via `WriteFileIfContentOrModeChangedAtomic` (temp + rename). But for required packs (`preserveOperatorEdits == false`), it also calls `pruneStaleGeneratedPackFiles` which removes files not in the current embedded pack. Between `os.RemoveAll(dst)` at the directory level and the per-file writes, a controller loading config from `.gc/system/packs/core` can observe missing or half-populated files. The design-after's Core presence doctor must specify a materialization swap pattern: write to a temp directory, verify, then rename over the target. The current `os.RemoveAll` + regenerate pattern is not safe under a live controller.

### [Major] `requiredBuiltinPackNames` hardcodes `"core"` and `"maintenance"` as required — the design-after says Maintenance is retired

`requiredBuiltinPackNames` (`embed_builtin_packs.go:236-250`) returns `[]string{"core", "maintenance"}` as always-required. The design-after says Maintenance is retired and its assets move to Core. After the migration, `"maintenance"` should no longer be required. The design must specify that the migration removes `"maintenance"` from `requiredBuiltinPackNames` and adds a transitional fallback for cities whose lockfile still references the maintenance pack. This is a code-level detail that the design-after references only implicitly through the runtime retirement table.

### [Major] The design-after's rollback compatibility matrix is incomplete for the `pack.toml` import binding rename from `maintenance` to `core`

When the import-state Fix removes the `[imports.maintenance]` binding and the Gastown pack moves from importing `../maintenance` to Core, any existing `pack.toml` that has order skip lists referencing `maintenance.gate-sweep` or `maintenance.jsonl-export` will break. The design mentions "order skip list name preservation" but does not specify: (a) whether `maintenance.*` qualified names are aliased to `core.*` during config loading, (b) whether `gc doctor --fix` rewrites skip list entries, or (c) whether the Gastown pack must carry both old and new qualified names during the transition.

### [Major] `PublicGastownPackVersion` is a SHA-pinned constant but the design-after does not specify an immutability validation

`PublicGastownPackVersion` (`internal/config/public_packs.go:8`) is `sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b`, which appears to be a concrete commit SHA. The design-after requires that the public Gastown pack is pinned to an immutable commit. But the preflight validation in the import-state Fix (`resolveWave1PublicPackImports`) is a static map lookup that returns the hardcoded source and version — it does not verify that the commit is reachable, installable, or lockable. An air-gapped host would pass preflight (no network call) and then fail at `syncImports` after manifests are already rewritten. The design must specify that preflight makes a network reachability check, or defer all manifest writes until after the lock/install validation succeeds.

---

## Missing evidence

1. **No provenance detection mechanism.** The design requires provenance-matched auto-rewrites but provides no specification for how provenance is established. The codebase uses path-suffix matching which cannot distinguish generated from edited directories.

2. **No concurrency test for doctor fix under a live controller.** The design requires byte-identical healthy-city behavior but does not specify or require tests for `gc doctor --fix` running concurrently with a controller reading `.gc/system/packs/core`.

3. **No test for partial-failure rollback.** The import-state Fix touches `pack.toml`, `city.toml`, `.gc/site.toml`, the lockfile, and installed pack directories. No test verifies that a failure at step 4, 5, or 6 leaves all files byte-identical. The existing test (`TestImportStateDoctorCheckFixRewritesLegacyPublicPackImports`) only tests the happy path with mocked `syncImports` and `installLockedImports`.

4. **No test for comment/unknown-key preservation through the fix path.** The design requires TOML content preservation but the codebase's round-trip path loses comments and unknown keys, and no test in the current suite verifies preservation.

5. **No specification of Core presence check registration order relative to the import-state check.** The doctor framework is registration-order-based with no dependency mechanism (`internal/doctor/doctor.go:68-86`). If the import-state Fix removes a redundant Core import before the Core presence doctor materializes Core, the city may be in an invalid intermediate state.

6. **No specification of what `PublicGastownPackVersion` immutability means in practice.** Is it always a commit SHA? Could it be a tag? The design says "immutable commit" but the constant is just a string — there is no validation that the value is a SHA or that it resolves to an immutable ref.

7. **No fault-injection test plan.** The design requires failure-atomic behavior but does not specify how to inject failures at each step of the fix pipeline to verify rollback.

8. **No test for the `pack.toml` write path lacking snapshot/restore.** The existing `writeCityConfigForEditFS` path has snapshot/restore for `city.toml` and `.gc/site.toml` (`site_binding.go:368-386`), but `writeCityPackManifest` has no equivalent. No test exercises a `pack.toml` write failure during the import-state Fix.

9. **No specification of `.gc/runtime/packs/maintenance/` state migration.** The design mentions `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` as stale directories, but the existing `jsonl_archive_doctor_check.go` and maintenance scripts hardcode `packs/maintenance` in runtime state paths. The design does not address what happens to existing JSONL archive repos, state files, and storm-count ledgers under `.gc/runtime/packs/maintenance/`.

10. **No specification of `[[patches.agent]] name = "dog"` cross-pack resolution.** After migration, `dog` comes from Core, but Core is a required system pack injected by `builtinPackIncludes` — not an explicit import. The Gastown pack patches `dog` via `[[patches.agent]] name = "dog"`, but the current config loader may not resolve patch targets from required-but-not-imported system packs.

---

## Required changes

1. **Defer all manifest mutations until after preflight validation succeeds.** Restructure the import-state Fix so that network reachability, lockability, and installability are verified before any manifest write. If preflight fails, leave all manifests byte-identical and emit manual guidance. This eliminates the need for multi-file rollback in the common case and directly implements the design-after's stated contract.

2. **Add snapshot/restore for `pack.toml` or defer all writes.** Either extend `writeCityPackManifest` to snapshot and restore like `writeCityAndRigSiteBindingsForEdit` does for `city.toml`, or defer all manifest writes to after lock/install validation succeeds.

3. **Unconditionally require TOML content preservation or fix refusal.** Remove the conditional "if the parser cannot preserve" language. Either implement a scoped TOML editor with comment/unknown-key preservation tests, or require detection of out-of-schema content (via `Undecoded()` capture) and refusal to auto-fix when it is present.

4. **Replace path-suffix provenance matching with exact canonical path matching plus content verification.** Only auto-rewrite imports whose source is exactly `.gc/system/packs/{gastown,maintenance}` (relative to city) or the embedded `Repository//` form, and only when the directory content matches the known generated manifest. All other matches should be diagnostic/manual.

5. **Specify import-state check dependency on Core presence check.** Either add a check dependency mechanism to the doctor framework, or require the import-state Fix to call `MaterializeBuiltinPacks` itself before removing redundant Core imports.

6. **Specify air-gap behavior.** On unreachable public Gastown, leave manifests byte-identical and emit manual guidance. This should be a first-class test case.

7. **Specify a materialization swap pattern for Core.** Core materialization should write to a temp directory and rename, not `os.RemoveAll` + regenerate, to avoid a read-consistency gap under a live controller.

8. **Specify concurrency protection for `gc doctor --fix`.** Either refuse mutating fixes when the controller is active, or define a file lock that concurrent doctor/import-install processes respect.

9. **Address `.gc/runtime/packs/maintenance/` state continuity.** Specify whether existing runtime state is migrated in-place, read through a compatibility path, or reported with manual instructions. Document what happens to `jsonl_archive_doctor_check.go`'s hardcoded `packs/maintenance` path.

10. **Specify `[[patches.agent]]` resolution for required-but-not-imported system packs.** The Gastown pack currently patches `maintenance.dog`. After migration, `dog` is in Core. The design must specify whether patches can target agents from implicitly-included system packs, or whether Gastown must define its own `dog` agent.

---

## Questions

1. Should `gc doctor --fix` refuse to mutate imports when the controller is active, or should it define a concurrency contract for concurrent reads?
2. What existing artifact can prove a `.gc/system/packs/maintenance` directory is unmodified? Is there a generated manifest or content hash, or should the design require one?
3. Is `PublicGastownPackVersion` guaranteed to always be a commit SHA? If it could be a branch name or tag, the "immutable version" preflight check needs to resolve it to a commit first.
4. Should the import-state Fix own a multi-file transaction, or should it delegate all pack/import mutations to an existing `gc import install` path with proven atomicity?
5. What happens to existing `packs/maintenance` runtime state (JSONL archives, storm counts) after the migration? Is it migrated, read through a compatibility path, or reported as a manual step?
6. How does `[[patches.agent]] name = "dog"` resolve when `dog` is in Core (a required system pack) but Gastown does not explicitly import Core?
