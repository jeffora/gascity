# Elias Sato — DeepSeek V4 Flash (Required Core Loading Invariant Review, Iteration 2, Independent)

**Verdict:** block

**Lane:** Required Core inclusion, config provenance, production loader-bypass containment, loud failure on corrupt/partial Core, escape-hatch leakage.

Reviewed against the revised design-after (updated 2026-06-05T20:30, with review-gated migration invariants, `assertRequiredSystemPackProvenance`, scanner modeled on `TestGCNonTestFilesStayOnWorkerBoundary`, Core maintenance contract, and doctor safety contract) and grounded in the live codebase in `cmd/gc/`, `internal/config/`, `internal/builtinpacks/`.

---

## Top Strengths

1. **The migration-invariant gate structure is well-calibrated.** Requiring each slice to prove Core provenance, Gastown behavior preservation, and doctor failure-atomicity before proceeding is the correct incremental structure. The design now states these explicitly rather than leaving them implicit.

2. **The manifest-integrity mechanism is already battle-tested.** `packContainsEmbeddedState` (`embed_builtin_packs.go:154`) iterates every embedded file, checks `os.Lstat` for existence and permissions, and compares byte content. `unusableRequiredBuiltinPackNames` (embed_builtin_packs.go:149-160) uses this to detect corrupt or partial packs. The design does not need to reinvent integrity checking — it needs to make the existing check fatal on the production path.

3. **The scanner/allowlist enforcement model has a proven precedent.** `TestGCNonTestFilesStayOnWorkerBoundary` is enforced in CI and prevents import regressions. Extending this pattern to `config.Load*` call sites is structurally sound.

4. **The provenance substrate is real and inspectable.** `config.Provenance.Sources []string` (compose.go:76) accumulates every source file, and `builtinPackIncludes` deterministically produces the pack directory paths that get expanded and appended during composition.

---

## Critical Risks

### [Blocker] The `LoadWithIncludes` dedup-by-pack-name silently drops the validated Core pack when any other import reaches a pack named "core"

This remains the highest-impact finding in the review, and prior wave-1 reviewers underweighted how the actual dedup code path terminates:

`LoadWithIncludes` appends system-pack includes after user includes, then runs `ExpandCityPacks` which calls `resolvedPackNames` on the combined list. `resolvedPackNames` discovers pack names by reading each pack's `pack.toml` name field. When a system-pack include has the same pack name as a pack already reachable through the root import closure, the system-pack include is **silently skipped** (`internal/config/compose.go:410-419`).

`TestLoadWithIncludes_SkipsSystemPackWhenReachableFromRootImport` (`compose_test.go:98-169`) proves this: it passes `system/maintenance` as an include, and the test explicitly asserts that the unqualified `dog` agent from the system pack is **absent** from the resolved config, because the root import already reaches a pack with the same name.

The design's `assertRequiredSystemPackProvenance` checks for the Core system-pack **path** in `prov.Sources`. But when dedup drops the system-pack include, the pack's `pack.toml` is never parsed, so the Core system-pack path is **never appended** to `prov.Sources`. A path-based assertion correctly **fails** in this case.

However: the design says "A successful materialization without resolved provenance is a load failure." This is correct only if the assertion always runs on the production path. But the design also permits `LoadWithIncludes` to silently skip a Core pack that happens to collide with a user import named "core" — which is exactly the scenario where the assertion should fail. The design must state whether a user importing a pack named "core" from another source is:

- **A hard error**: the composition fails because a required system pack collides with a user import.
- **A diagnostic with a fix**: composition proceeds but doctor reports the conflict.
- **Silently accepted**: the user's "core" replaces the system Core, and the assertion passes on the user's pack (which is unvalidated).

The correct answer per the design's own principles is (a): a name collision with the required system Core pack is a hard error. But the design does not state this, and the current dedup behavior in `ExpandCityPacks` silently skips system Core when any user import chain reaches a pack also named "core". The v0.15.1 collision gate (`compose.go:440-450`) only hard-stops on **implicit-import** collisions, not on system-pack collisions. A user who explicitly imports a pack named "core" from an external source gets the user's pack, not the system Core — and the provenance assertion would either pass on the wrong pack or fail without an actionable error message.

**Required change:** The design must specify that a pack-name collision between a required system Core pack and any user/imported pack named "core" is a hard composition error, not a silent dedup. Either: (a) add a collision check after `resolvedPackNames` that hard-stops when a required system pack name collides with a non-system source, or (b) change the system-pack inclusion to use a disambiguating namespace or path-based identity that cannot collide. Option (a) is consistent with the existing v0.15.1 collision-gate pattern. The design must also specify that `assertRequiredSystemPackProvenance` produces an actionable error message naming the colliding source when this failure mode triggers.

### [Blocker] `builtinPackIncludes` uses `packExists` (a `pack.toml` stat check) as the gate for inclusion, which means a partially-materialized Core directory with `pack.toml` present but content missing is included but broken

`builtinPackIncludes` (`embed_builtin_packs.go:272-285`) iterates `requiredBuiltinPackNames` and calls `packExists` for each, which checks `os.Stat(filepath.Join(dir, "pack.toml"))`. If `pack.toml` exists on disk, the pack path is included in the config-load includes. If `pack.toml` does not exist, the pack path is silently omitted from includes.

This means:

1. **A partially-materialized Core** (e.g., `pack.toml` present from a previous `MaterializeBuiltinPacks` but other files missing due to a crash or disk error) gets included in config loading. The `ExpandCityPacks` step reads the pack's `pack.toml`, appends its path to `prov.Sources`, and merges whatever content it finds. The provenance assertion passes because the Core path is in `prov.Sources`. But the config is incomplete — missing Core agents, orders, formulas, hooks, and overlays.

2. **A missing Core** (no `pack.toml` at all) gets silently excluded from includes. Config loads without Core. The provenance assertion fails. This is the correct failure mode, but the design never specifies what happens between `MaterializeBuiltinPacks` and `builtinPackIncludes` when `MaterializeBuiltinPacks` succeeds but `pack.toml` is still absent. Can `MaterializeBuiltinPacks` succeed and leave no `pack.toml`? Yes: if the atomic rename for `pack.toml` fails or the directory permissions prevent the write, `MaterializeBuiltinPacks` returns an error, but `cityConfigIncludesWithBuiltinPacks` (`cmd_config.go:29-38`) calls `MaterializeBuiltinPacks` independently of `ensureBuiltinPacksReadyForConfigLoad`, and its error path only checks for `MaterializeBuiltinPacks` failure — not for partial success.

The design's "loud failure on corrupt/partial Core" invariant depends on **both** `unusableRequiredBuiltinPackNames` (manifest integrity) and `assertRequiredSystemPackProvenance` (provenance presence) running on every production path. But the design's wrapper list includes `cityConfigIncludesWithBuiltinPacks`, which:

1. Calls `MaterializeBuiltinPacks` directly (not through `ensureBuiltinPacksReadyForConfigLoad`).
2. Never checks `unusableRequiredBuiltinPackNames`.
3. Returns `[]string` includes — no `*config.Provenance` to assert against.

Prior reviews (Claude, DeepSeek wave 1) both identified that two of the three named wrappers cannot host the assertion, but none identified that `cityConfigIncludesWithBuiltinPacks` also **skips the manifest-integrity check**. This means the controller-reload path (`tryReloadConfig` → `cityConfigIncludesWithBuiltinPacks`) materializes packs but does not verify their integrity before using them.

**Required change:** The design must specify that **every** production path that calls `MaterializeBuiltinPacks` must also run `unusableRequiredBuiltinPackNames` after materialization and fail loudly if it returns non-empty for required packs. This is not the same as the post-load provenance assertion. Both checks must be present and fatal. Either: (a) all production paths route through `ensureBuiltinPacksReadyForConfigLoad` (which already runs `unusableRequiredBuiltinPackNames`), or (b) the design adds an explicit manifest-integrity check to `cityConfigIncludesWithBuiltinPacks` and documents that both include-building wrappers enforce manifest integrity.

### [Blocker] `tryReloadConfig` is still a silent Core-bypass path after the design's changes

`tryReloadConfig` (`controller.go:895-905`) calls `cityConfigIncludesWithBuiltinPacks` and then `config.LoadWithIncludes`. As analyzed above:

1. `cityConfigIncludesWithBuiltinPacks` materializes packs without checking `unusableRequiredBuiltinPackNames`.
2. No post-load assertion runs on the result.
3. If `packExists` returns false for Core (e.g., transient I/O error, race with another `gc` process clearing `.gc/system/packs/core`), `builtinPackIncludes` returns an empty slice for Core, and the controller reloads a Core-less config.

The prior DeepSeek review identified `tryReloadConfig` as the weakest path. The prior Claude review also identified it. But neither review noted that `tryReloadConfig`'s call to `cityConfigIncludesWithBuiltinPacks` is the **only** include-building path that calls `MaterializeBuiltinPacks` directly (rather than through `ensureBuiltinPacksReadyForConfigLoad`), and therefore it is the only path where a partial Core is included without manifest validation.

The design names `tryReloadConfig` nowhere. The design's wrapper list names `cityConfigIncludesWithBuiltinPacks` but does not classify it or its callers.

**Required change:** `tryReloadConfig` must be explicitly in scope. The design must specify either: (a) `tryReloadConfig` routes through `loadCityConfigWithBuiltinPacks` which runs both the manifest-integrity check and the provenance assertion, or (b) `tryReloadConfig` runs a post-load assertion that includes both manifest integrity and provenance presence, with focused tests for each failure mode. Option (a) is simpler and more consistent.

### [Blocker] The `config.Load` (no-includes) surface is larger and more dangerous than the design acknowledges

The design's scanner says "Direct `config.Load`, `config.LoadCity`, `config.LoadWithIncludes`, and package aliases of those functions in production `cmd/gc` files should be rejected." But the design's wrapper list only names three functions that use `LoadWithIncludes`. It does not enumerate or classify the `config.Load` (no-includes) call sites.

A complete inventory of `config.Load` calls in production `cmd/gc/` files:

| File:Line | Function | Purpose | Classification |
|---|---|---|---|
| `apiroute.go:34` | API route discovery | Port discovery from raw config | Partial-read candidate |
| `apiroute.go:80` | API route readiness | Port check during startup | Partial-read candidate |
| `api_state.go:767` | `RawConfig()` | Serve live config to API | Behavior-driving (serves agent definitions to dashboard) |
| `cmd_agent.go:174` | `loadCityConfigForEditFS` | Raw config for writes | Partial-read (write-back, not runtime) |
| `cmd_config.go:634` | Quick pack check | Skip init if packs already installed | Partial-read (pre-init heuristic) |
| `cmd_import.go:1243` | Import pack parsing | Read pack config | Partial-read (import flow) |
| `cmd_import.go:1347` | Import pack validation | Validate imported pack | Partial-read (import flow) |
| `cmd_register.go:105` | Registration | Read city name for registration | Behavior-driving (registration affects runtime state) |
| `cmd_start_drift.go:390` | Drift detection | Config comparison for drift | Behavior-driving (drift triggers runtime actions) |
| `doctor_v2_checks.go:488` | Import-state doctor | Read raw config for doctor | Partial-read (diagnostic) |
| `doctor_v2_checks.go:1130` | Import-state doctor | Read raw config for doctor | Partial-read (diagnostic) |
| `init_provider_readiness.go:249` | Init preflight | Read config for provider readiness | Partial-read (pre-init) |
| `legacy_pack_preflight.go:16` | Legacy pack check | Quick check for existing packs | Partial-read (pre-init heuristic) |

At minimum, `api_state.go:767` (RawConfig serving expanded agent definitions to the dashboard) and `cmd_start_drift.go:390` (drift detection triggering runtime actions) are behavior-driving paths that use `config.Load` without any pack expansion. These paths can serve or act on a config that lacks Core's agents, overlays, and formulas.

The design must either: (a) classify each of these ~13 sites explicitly in the scanner allowlist with a justification and focused test, or (b) route the behavior-driving ones through `loadCityConfigWithBuiltinPacks`.

**Required change:** Add a complete `config.Load` call-site inventory to the design, with explicit classification for each. Behavior-driving paths (`RawConfig`, drift detection, registration) must either be routed through the assertion path or explicitly documented as partial-read exceptions with focused tests proving they cannot dispatch agent/session/order behavior from a Core-less config.

---

## Major Risks

### [Major] The design conflates two separate failure modes — materialization failure and config-load bypass — and assigns them to a single mechanism that covers only one

Prior reviewers (Claude, prior DeepSeek, Gemini) all noted that the provenance-presence assertion detects absence but not corruption/partial content. The design's review-gated invariant says "missing, corrupt, or partially materialized Core" should fail at load. But the design's mechanism (`assertRequiredSystemPackProvenance`) only detects absence (Core path not in `prov.Sources`).

The manifest-integrity check (`unusableRequiredBuiltinPackNames`) detects corruption and partial materialization, but it is only called from `ensureBuiltinPacksReadyForConfigLoad`. It is not called from `cityConfigIncludesWithBuiltinPacks` or from the post-load path.

The design must specify these as two separate, both-fatal checks:

1. **Manifest integrity** (`unusableRequiredBuiltinPackNames` → fatal on production path): catches missing, corrupt, and partial Core.
2. **Provenance presence** (`assertRequiredSystemPackProvenance` → fatal on production path): catches "materialized but dropped from config" (e.g., dedup, `packExists` race, include-builder bug).

These are complementary, not redundant. The design currently describes both as aspects of a single "Core provenance" check, which is imprecise and will lead implementers to believe that one check covers both failure modes.

**Required change:** Define two separate check names and two separate failure-mode contracts. State explicitly: (1) `unusableRequiredBuiltinPackNames` returning non-empty for required packs after materialization is a hard failure; (2) `assertRequiredSystemPackProvenance` failing after a successful manifest check is a hard failure; (3) both checks must run on every production config-load path.

### [Major] The `materializeBuiltinPacksForConfigLoad` warning-downgrade path is a production reliability gap that the design does not close

`materializeBuiltinPacksForConfigLoad` (`embed_builtin_packs.go:136-148`) downgrades a failed refresh to a **warning** when existing content is "usable" (passes `unusableRequiredBuiltinPackNames`). After the migration, Core is the only always-required pack. A stale-but-usable Core means the on-disk content matches the embedded manifest but is from a previous binary version.

The design's review-gated invariant #2 says "Normal production `gc` config loading proves Core was materialized and included in resolved provenance." A stale Core from a previous binary may have missing orders, formulas, hooks, or overlays that the new binary depends on. "Stale but usable" means byte-identical to the old embedded manifest, not the current one.

The design must state one of:

- **(a) Hard failure on any `unusableRequiredBuiltinPackNames` result after materialization**: If Core fails manifest verification, the production load fails. Period. This means a transient I/O error during `MaterializeBuiltinPacks` that leaves Core incomplete is fatal, and the operator must retry.
- **(b) Warning for non-required packs, hard failure for required packs**: Downgrade is allowed for `bd`/`dolt` (provider packs) but not for Core.

Option (a) is simpler and more consistent with the "no production opt-out" posture. Option (b) is more tolerant of transient I/O on provider packs.

**Required change:** State the policy explicitly. My recommendation is (a): `unusableRequiredBuiltinPackNames` returning non-empty for any required pack after `MaterializeBuiltinPacks` + `ensureBuiltinPacksReadyForConfigLoad` is a hard failure on the production path. The current warning path is acceptable only for non-required packs, which no longer include Maintenance.

### [Major] `loadCityConfigWithoutBuiltinPackRefreshFS` skips `MaterializeBuiltinPacks` entirely and is underspecified

This function (`cmd_agent.go:66-80`) uses whatever packs are on disk without refreshing them. It is called from completion (`completion.go`) and stop (`cmd_stop.go`). The design says no-refresh helpers "must either run the same Core provenance assertion after loading or be documented as partial-read exceptions" but does not classify these callers.

After the migration, completion reads agent definitions that may come from Core. If Core is absent from disk (e.g., first run before `gc init`, or after a manual `.gc/system` cleanup), completion silently shows a Core-less config to the user.

**Required change:** Classify each `loadCityConfigWithoutBuiltinPackRefresh` caller:
- **Completion**: If completion reads agent names from config (it does), it needs Core included. Either route through `loadCityConfig` (which refreshes), or add a post-load assertion that Core is present in provenance, or document as a partial-read exception with a test proving completion cannot dispatch formulas/orders from stale/absent Core.
- **Stop**: If stop reads agent pools for shutdown targeting (it does), it needs Core included. Same options apply.

### [Major] The `effectiveCityName` bypass materializes packs but never includes them

`effectiveCityName` (`cmd_supervisor_city.go:86-96`) calls `MaterializeBuiltinPacks` and then `config.LoadWithIncludes` **without** `builtinPackIncludes`. It materializes packs to disk but never adds their paths to the include list. This is exactly the pattern the design must prevent: a production path that goes through materialization but produces a Core-less config.

This is a behavior-driving path because `effectiveCityName` is called from `registeredCityName` (`cmd_supervisor_city.go:107`), which is called during city startup and registration — production paths that affect runtime state.

**Required change:** Fix `effectiveCityName` to pass `builtinPackIncludes(cityPath)` to `config.LoadWithIncludes`, or route it through `loadCityConfigWithBuiltinPacks`. Add this to the design's required changes list and include it in the scanner's seeded allowlist only if it is a justified partial-read exception.

### [Major] The design's "no production opt-out" posture conflicts with the `usesOSFS` guard in `builtinPackIncludesForConfigLoad`

`builtinPackIncludesForConfigLoad` (`embed_builtin_packs.go:83-92`) checks `usesOSFS(fs)` before including packs. If the filesystem is not `OSFS`, it returns `nil, nil` — no includes, no error. This means any test or utility that uses a non-OS filesystem (e.g., a memory FS) silently gets a Core-less config.

The design says "The dev/test escape hatch is not a CLI flag or environment variable. Production `gc` commands always include Core." But the `usesOSFS` check is a filesystem-type escape hatch that is not a `_test.go` helper — it's in production code. If any production path accidentally uses a non-OSFS filesystem wrapper, it silently bypasses Core.

**Required change:** The design should specify that `usesOSFS` is an intentional test-only escape hatch, and the scanner should verify that no production `cmd/gc` code calls `builtinPackIncludesForConfigLoad` or `loadCityConfig*` with a non-OSFS filesystem. Alternatively, the design should specify that `usesOSFS` is removed and test isolation uses explicit mock-pack fixtures instead.

---

## Minor Risks

### [Minor] Path normalization for provenance comparison

The provenance assertion and Core doctor need to compare "the materialized Core system pack path" against entries in `prov.Sources`. The codebase already uses `normalizePathForCompare` (`embed_builtin_packs.go`) for path comparison. The design does not address path normalization (absolute vs. relative, symlinks, trailing slashes).

**Required change:** State that the assertion must use the same path normalization as `normalizePathForCompare`, or define a canonical path form for the comparison.

### [Minor] Empty-pack provenance edgecase

If a Core system pack contributes zero effective config keys (all its agents, orders, formulas are overridden by user packs), does `prov.Sources` still contain the Core path? The code adds pack paths at `compose.go:279` when `pack.toml` is parsed, regardless of whether the pack contributes any keys. But this is not tested or documented.

**Required change:** Add a test proving that a Core system pack contributing zero effective keys still appears in `prov.Sources`.

### [Minor] `cmd_convoy.go` has three direct `LoadWithIncludes` calls without pack materialization or assertion

`cmd_convoy.go:144`, `cmd/gc/cmd_convoy.go:561`, and `cmd/gc/cmd_convoy.go:699` call `config.LoadWithIncludes` directly without `MaterializeBuiltinPacks` or any assertion. These are convoy create/inspect paths. If packs are already materialized from a prior `gc` command, these work fine. If not (e.g., first run), they load a Core-less config.

**Required change:** Classify these in the scanner allowlist or route them through `loadCityConfig`.

### [Minor] `cmd_init.go` has `LoadWithIncludes` calls during init that may not have Core on disk yet

`cmd_init.go:1106` and `cmd_init.go:1371` call `LoadWithIncludes` during the init flow. At this point, Core may not be materialized yet (or may be partially materialized). The init flow explicitly calls `MaterializeBuiltinPacks` separately (`init_provider_readiness.go:42`), but the timing of when `LoadWithIncludes` runs relative to materialization is not guaranteed.

**Required change:** Verify that init-flow `LoadWithIncludes` calls always run after `MaterializeBuiltinPacks` and after `builtinPackIncludes`, or classify them as partial-read exceptions with a justification.

---

## Required Changes (Priority Order)

1. **Define two separate, both-fatal checks.** Manifest integrity (`unusableRequiredBuiltinPackNames` → fatal) catches missing, corrupt, and partial Core. Provenance presence (`assertRequiredSystemPackProvenance` → fatal) catches "materialized but dropped from config." Both must run on every production path.

2. **Add a pack-name collision gate for required system packs.** When `resolvedPackNames` discovers that a user/imported pack named "core" collides with the required system Core, the composition must hard-stop with an actionable error, not silently dedup. Extend the v0.15.1 collision-gate pattern.

3. **Route `cityConfigIncludesWithBuiltinPacks` through manifest-integrity checking.** Either make it call `ensureBuiltinPacksReadyForConfigLoad` (which checks `unusableRequiredBuiltinPackNames`), or add an explicit manifest check. The current path materializes without validation.

4. **Name `tryReloadConfig` as in-scope.** The controller reload must run both manifest-integrity and provenance-presence assertions, or route through `loadCityConfigWithBuiltinPacks`.

5. **Fix `effectiveCityName` to include system packs.** Pass `builtinPackIncludes(cityPath)` or route through `loadCityConfigWithBuiltinPacks`.

6. **Replace the three-name wrapper list with a complete classification of all six config-returning entry points and all ~13 `config.Load` (no-includes) sites.** Each must have: assertion required, partial-read exception (with justification and test), or migrate-to-wrapper.

7. **Close the `materializeBuiltinPacksForConfigLoad` warning-downgrade for required packs.** State explicitly: `unusableRequiredBuiltinPackNames` returning non-empty for Core after materialization is fatal. The warning path is acceptable only for non-required packs.

8. **Classify `loadCityConfigWithoutBuiltinPackRefresh` callers** (completion, stop) and either add a post-load Core assertion or document the partial-read exception.

9. **Classify `config.Load` (no-includes) call sites** with explicit justifications for each. Behavior-driving paths (`RawConfig`, drift detection, registration) must be in the scanner allowlist with focused tests.

10. **Address the `usesOSFS` filesystem-type escape hatch.** Either document it as a test-only bypass and add a scanner rule that no production code calls the affected functions with a non-OSFS filesystem, or specify an alternative test isolation strategy.

11. **Fix the cross-document inconsistency** between requirements ("offer a fix that adds it") and design ("repair the generated system pack and normal include path").

12. **Specify path normalization** for the provenance comparison, referencing the existing `normalizePathForCompare`.

13. **Add a test** proving that a Core system pack contributing zero effective keys still appears in `prov.Sources`.

---

## Questions

1. When a user explicitly imports a pack named "core" from an external source, should the composition hard-stop, warn and proceed with the user's pack, or silently dedup? The v0.15.1 collision gate hard-stops for implicit imports but silently dedups for explicit imports. The design's "no production opt-out" posture implies hard-stop.
2. Is `loadCityConfigWithoutBuiltinPackRefreshFS` intended to remain in production after Core becomes mandatory? If so, which config fields may safely be read from a config that might lack Core overlays?
3. Should `cityConfigIncludesWithBuiltinPacks` (which bypasses `ensureBuiltinPacksReadyForConfigLoad`) be merged into or replaced by `builtinPackIncludesForConfigLoad` (which includes the manifest-integrity check)?
4. For stale-but-manifest-valid Core after a failed refresh, is the intended behavior hard-fail or warn-and-proceed? The current code warns; the design says "no production opt-out."
5. Does `prov.Sources` include a pack path when the pack contributes zero effective config keys? The provenance assertion's correctness depends on the answer.
