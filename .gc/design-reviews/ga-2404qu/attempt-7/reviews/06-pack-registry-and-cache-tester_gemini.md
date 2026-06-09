# Marcus Driscoll — DeepSeek V4 Flash (Independent Review, Iteration 7)

**Persona:** Pack Registry and Cache Tester
**Mandate:** Builtin registry identity, synthetic cache pruning, system pack materialization, provider-dependent pack continuity
**Verdict:** block

---

## Lane Context & Alignment

This independent review evaluates the latest migration design document (updated 2026-06-07T02:23:08Z, located at [design-before.md](file:///data/projects/gascity/.gc/design-reviews/ga-2404qu/attempt-7/design-before.md)) against the current Go codebase. The evaluation focuses on the built-in pack registry (`internal/builtinpacks/registry.go`), synthetic cache materialization and pruning (`internal/packman/cache.go`), and provider-dependent pack continuity (`bd`/`dolt` matrices).

While the design is exceptionally structured, several critical, high-risk gaps exist around un-namespaced prompt globbing, silent remote fallbacks for retired sources, expected-files-only Core validation, and global cache invalidation of provider packs. These issues must be addressed before the design can be approved.

---

## Top Strengths

1. **Deterministic and Exact Builtin Registry Identity:** 
   Updating `All()` to return only `{core=internal/packs/core, bd=examples/bd, dolt=examples/dolt}` and completely retiring the `maintenance` and `gastown` synthetic aliases is structurally clean. Reasserting exact string-equality and exact-count checks in `TestAllAndSourceAreDeterministic` prevents any silent re-addition of retired aliases.
2. **Stop-Generating-Don't-Delete Mechanics:** 
   The decision to leave stale directories on disk rather than deleting them (`MaterializeBuiltinPacks` only iterates over the active `All()`) eleganty avoids accidental operator data loss. It is robust, self-documenting, and requires no complex clean-up code.
3. **Explicit Synthetic Cache Namespace Separation:** 
   Restricting the `SyntheticContentHash()` verification strictly to bundled packs (`core`, `bd`, `dolt`) and routing `gascity-packs/gastown` through ordinary remote cache paths keyed by repository source and immutable version avoids cross-contamination of cache lookups.

---

## Critical Risks & Blockers

### 1. Stale `.gc/system/packs/` Prompts Loaded by Un-namespaced Glob Crawling
* **Risk Class:** Functional Regression / Behavior Drift
* **Grounded Location:** `cmd_prompt.go:613`; [design-before.md:866-870](file:///data/projects/gascity/.gc/design-reviews/ga-2404qu/attempt-7/design-before.md#L866-L870)

The design specifies that stale local `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories are ignored by `builtinPackIncludes` but preserved on disk on startup to protect potential operator edits.

However, the prompt template discovery helper `loadBaselinePrompt` (via `cmd_prompt.go`) crawls and globs baseline templates using `.gc/system/packs/*/agents/<role>/prompt.template.md` without filtering against the active or required pack lists. Since these stale directories remain on disk, their prompt templates (e.g., `mayor`, `deacon`, `polecat` under `gastown`) will still be discovered, loaded, and active as prompt baselines! An upgraded city will thus silently load stale prompts from the retired local directories, completely violating the de-coupling guarantee.

* **Required Resolution:** 
  The design must explicitly mandate that `loadBaselinePrompt` (and any other prompt/template discovery glob walks) filters walked directories, ensuring that only packs listed in the resolved config's active/required pack set are visited, or stale prompts from ignored packs are skipped.

---

### 2. Retired Bundled Sources Silently Reclassified as Ordinary Git Remotes
* **Risk Class:** Silent Configuration Failures / Opaque Error Messages
* **Grounded Location:** `internal/builtinpacks/registry.go:89-139` (`NameForSource`/`IsSource`); `internal/config/pack_include.go:227-229`

Once `All()` drops `maintenance` and `gastown`, and Core's subpath shifts to `internal/packs/core`, `IsSource` will return `false` for retired in-tree sources (e.g., `internal/bootstrap/packs/core` or `examples/gastown/packs/maintenance`).

Instead of triggering an explicit migration diagnostic, any existing lockfile or import reference pointing to these old sources will fall back to the ordinary Git remote import path. Because those legacy subpaths or repos do not exist or are inaccessible in that format, the operator will receive an opaque network-fetch error ("locked but not cached... run gc import install") instead of a clear, actionable migration message.

* **Required Resolution:** 
  The source recognition helpers must explicitly map known retired sources to a blocklist/retired list, returning a typed diagnostic (e.g., `ErrLegacySourceRetired`) so that `gc doctor` or the loader can output a clear, actionable migration prompt telling the operator exactly how to update their imports/locks.

---

### 3. Core Repair Doctor Check Uses Expected-Files-Only Validator, Permitting File Injection
* **Risk Class:** Security & Tamper Vulnerability / System Degradation
* **Grounded Location:** `cmd/gc/embed_builtin_packs.go:170-195` (`packContainsEmbeddedManifest`); [design-before.md:930-945](file:///data/projects/gascity/.gc/design-reviews/ga-2404qu/attempt-7/design-before.md#L930-L945)

The proposed Core presence doctor check uses `packContainsEmbeddedManifest` to verify `.gc/system/packs/core/pack.toml` and its files against the embedded manifest. 

This helper is an expected-files-only check: it asserts that every *expected* file is present and content-correct, but it does not walk the destination directory to detect unexpected files. An injected malicious formula, overlay, or stale script in `.gc/system/packs/core` will pass the doctor check completely unnoticed. This directly contradicts the strict security requirements of required system packs.

* **Required Resolution:** 
  Required host system packs (`core`, `bd`, `dolt`) must use the strict `validatePackFiles` approach that walking the directory and rejecting any unexpected files, treating any additions as corruption that triggers immediate repair.

---

### 4. Global `SyntheticContentHash` Coupling and Provider Cache Invalidation
* **Risk Class:** Performance Regression / Offline Upgrade Failure
* **Grounded Location:** `internal/builtinpacks/registry.go:252-274` (`SyntheticContentHash`); `internal/packman/cache.go:70-95`

`SyntheticContentHash` computes a single, monolithic SHA-256 digest covering *all* built-in pack layouts. When the list of built-in packs changes (removing Maintenance/Gastown and shifting Core to `internal/packs/core`), the global hash changes.

This invalidates *all* pre-existing synthetic cache directories on the operator's machine, including the critical caches for `bd` and `dolt` packs that provider cities depend on. On the first run of the upgraded binary, the system will trigger a complete on-disk re-materialization of these caches. If this occurs on an offline or air-gapped machine, and the system tries to validate or re-materialize these caches, it risks breaking provider-dependent pack continuity.

* **Required Resolution:** 
  The design must explicitly acknowledge this one-time cache invalidation and mandate that `MaterializeBuiltinPacks` ensures the re-materialization of `bd` and `dolt` remains 100% offline-capable, byte-identical, and testable from pre-migration cache states.

---

## Major Risks & Gaps

### 1. `dog` Pool-Name Hardcoding in Provider Pack Contradicts Role-Neutrality
* **Risk Class:** Architectural Incoherence / ZFC Violation
* **Grounded Location:** `examples/dolt/orders/mol-dog-stale-db.toml`; [design-before.md:168-187](file:///data/projects/gascity/.gc/design-reviews/ga-2404qu/attempt-7/design-before.md#L168-L187)

The design emphasizes that `dog` is a freely renameable/omittable agent pool (role-neutrality), and that Core maintenance formulas resolve the configured name from configuration rather than a Go constant.

However, Dolt's `mol-dog-stale-db.toml` hardcodes `pool = "dog"`. This creates a direct contradiction: if `dog` is renamed in Core to achieve role neutrality, Dolt's order will fail to resolve the agent, breaking provider pack continuity.

* **Required Resolution:** 
  The design must resolve this conflict by either (a) declaring `dog` as a stable, non-renameable contract that provider packs can bind to, or (b) making Dolt's orders dynamically resolve the maintenance worker name from active config, adjusting the role-neutrality tests accordingly.

---

### 2. Dead Code in `normalizeRepository` and `publicSubpathForPack`
* **Risk Class:** Technical Debt / Silent Regression
* **Grounded Location:** `internal/builtinpacks/registry.go:126-133`, `33`, `447-458`

`normalizeRepository` contains a special-case branch mapping `gascity-packs` URLs to `PublicRepository`. After the public Gastown synthetic alias is removed, this branch serves no purpose for synthetic cache recognition.

If left in place, this normalization branch and the unused `publicSubpathForPack` cases represent dead code that can silently re-introduce retired aliases if modified.

* **Required Resolution:** 
  The design must explicitly state whether `PublicRepository` and the `gascity-packs` normalization branch in `normalizeRepository` are deleted as dead code or if they are preserved for standard remote URL normalization.

---

### 3. Missing Behavior-Manifest Mapping for `TestBuiltinDatabaseEnumeratorsSkipManagedProbeDatabase`
* **Risk Class:** Provider-Continuity Regression
* **Grounded Location:** `embed_builtin_packs_test.go:183-203`

This test asserts that Dolt managed-probe-DB protection needles (e.g., `benchdb|testdb_*|beads_pt*`) inside `jsonl-export.sh` and `reaper.sh` are correctly filtered. Since both scripts are moving to Core as generic assets, this provider-safety witness must follow them.

Without mapping this test in the behavior manifest, a regression could occur where the generic Core reaper or export sweeps a `bd`/`dolt` managed probe DB.

* **Required Resolution:** 
  Add an explicit behavior-manifest row for `TestBuiltinDatabaseEnumeratorsSkipManagedProbeDatabase` pointing to the Core copies of `jsonl-export.sh`/`reaper.sh`, and assert the Dolt managed-probe-DB filter needles there.

---

## Minor Risks & Verification Gaps

### 1. Stale "Maintenance" Doc Comments
* **Grounded Location:** `embed_builtin_packs.go:54-59`, `266-271`

Doc comments inside `MaterializeBuiltinPacks` and `builtinPackIncludes` still refer to "Core and maintenance are always included." These comments must be updated to align with the new Core-only architecture.

---

### 2. Missing Negative Tests for Retired Core Subpaths
* **Grounded Location:** `internal/builtinpacks/registry_test.go:44-69`

`TestSourceRecognitionVariants` asserts various shapes of Core sources. The negative test suite must explicitly assert that `IsSource("https://github.com/gastownhall/gascity.git//internal/bootstrap/packs/core")` is false to prevent regression where the old path is accepted silently.

---

## Required Changes

1. **Glob Filtering in Prompt Crawler:** Update `loadBaselinePrompt` to restrict prompt template crawls to required and active config-resolved packs only, avoiding loading stale templates from preserved on-disk directories.
2. **Actionable Retired Diagnostics:** Add a retired-source recognizer that returns a typed `ErrLegacySourceRetired` for legacy paths to provide clear upgrade guidance instead of opaque network errors.
3. **Strict Core validation:** Re-use strict `validatePackFiles` full file-set validation for required system packs (`core`, `bd`, `dolt`), treating unexpected files as corruption that triggers immediate repair.
4. **Decouple/Document Synthetic Hash Invalidation:** Explicitly test the old-binary-cache to new-binary-cache migration path ensuring `bd` and `dolt` re-materialization remains entirely offline and byte-identical.
5. **Resolve the `dog` Pool-Name Contract:** Specify whether `dog` is a stable pool-name contract or if Dolt's orders must be updated to resolve the maintenance worker dynamically.
6. **Dead Code Cleanup:** Explicitly delete or document the retained purpose of `PublicRepository`, the `gascity-packs` normalization branch, and `publicSubpathForPack`.
7. **Database Enumerator Safeguard:** Map `TestBuiltinDatabaseEnumeratorsSkipManagedProbeDatabase` to the Core copy of reaper/export scripts in the behavior manifest.

---

## Questions

1. Does the Core presence doctor's `Run` reuse `packContainsEmbeddedManifest` (expected-files-only) or a full-file-set validator? If the former, an injected extra Core file reads healthy until the next `MaterializeBuiltinPacks` prune—is this an acceptable window?
2. Should `SyntheticContentHash` stay a single global digest, or be scoped per-pack so a Core-only change does not invalidate `bd`/`dolt` caches? This is the central provider-continuity decision and is currently implicit.
3. For a city whose lock pins the old `internal/bootstrap/packs/core` bundled source, what is the intended upgrade path—doctor rewrite, automatic re-resolve to the new subpath, or hard error with guidance? The doctor section addresses explicit imports but not locks.
4. Is an offline provider-city upgrade guaranteed never to need network access to re-materialize the bundled synthetic cache?
