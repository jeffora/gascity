# DeepSeek V4 Flash — Required Core Loading Invariant Review

**Verdict:** block

**Lane:** Required Core inclusion, config provenance, production loader bypass containment, loud failure on corrupt/partial Core, escape-hatch leakage.

Reviewed against the revised design (2026-06-05T20:30, with the `assertRequiredSystemPackProvenance` addition, scanner modeled on `TestGCNonTestFilesStayOnWorkerBoundary`, and review-gated migration invariants) and the live codebase in `cmd/gc/` and `internal/config/`.

---

## Top Strengths

- **The provenance substrate is real.** `config.Provenance.Sources []string` (`internal/config/compose.go:76`) accumulates every file that contributed to the resolved config, and `builtinPackIncludes` deterministically injects `.gc/system/packs/core` into that list. A post-load assertion that checks for the Core path in `Sources` is implementable and not speculative.
- **The materialization integrity mechanism is already strong.** `packContainsEmbeddedState` and `unusableRequiredBuiltinPackNames` do a byte-for-byte manifest comparison of embedded vs. on-disk content. The design can build on this rather than inventing parallel integrity checking.
- **The scanner/allowlist enforcement model is proven.** `TestGCNonTestFilesStayOnWorkerBoundary` exists and is enforced in CI. Extending it to ban direct `config.Load*` calls in production `cmd/gc` is structurally sound and directly addresses drift.
- **The review-gated migration invariants (added in the revised design) are the right idea.** Making each slice prove Core provenance, Gastown behavior preservation, and doctor failure-atomicity before proceeding is structurally correct even though the current design doesn't yet deliver on the provenance invariant.

## Critical Risks

### [Blocker] The proposed provenance assertion cannot detect partial Core materialization, yet the design demands it

`assertRequiredSystemPackProvenance` as specified (design lines 368-372, 406-409) verifies that the Core system pack *path* appears in `prov.Sources`. A partially-materialized Core directory (one where `pack.toml` is present but some orders/scripts/formulas are missing) still has its path contributed to `Sources` by `builtinPackIncludes`, so the assertion PASSES. The design simultaneously requires that "missing, corrupt, or partially materialized" Core be a load failure (review-gated invariant #2, design lines 54-57).

Provenance-presence catches *absence* — Core directory never loaded. It does not catch *partial content*. These are three distinct failure modes:

1. **Absent Core** — directory missing entirely. Provenance catches this.
2. **Corrupt Core** — known files with wrong content. `packContainsEmbeddedState` catches this.
3. **Partial Core** — some embedded files missing, but `pack.toml` present. Neither provenance-presence nor the current `unusableRequiredBuiltinPackNames` catches this, because `packContainsEmbeddedManifest` iterates embedded entries and checks them against disk — but a partially-materialized directory has `pack.toml` (passing `packExists`) and the manifest check only verifies files that *should* be there, not files that *are missing*.

Wait — actually, re-reading `validatePackFiles` in `registry.go:408-442`, the validation walks the embedded manifest AND checks that each manifest entry exists on disk with matching content. A missing file would be caught by `os.Lstat` returning an error. But `validatePackFiles` is only called from `validateSyntheticRepoFileSet` for synthetic cache validation, not from the `packContainsEmbeddedState` path used in `unusableRequiredBuiltinPackNames`. The code in `embed_builtin_packs.go` checks `packContainsEmbeddedState` which calls `manifestForFS` to get the expected file list and then `packContainsEmbeddedManifest` which iterates expected files and does byte comparison. A missing expected file would be caught by the `os.Lstat` inside `packContainsEmbeddedManifest` — so partial materialization IS caught by the existing manifest check, as long as the check is actually run on the critical path.

The real problem is: **the design does not specify that `unusableRequiredBuiltinPackNames` must be called on every production config load, nor that its result must be fatal on the normal path.** Currently `materializeBuiltinPacksForConfigLoad` (`embed_builtin_packs.go:140-148`) downgrades a partial refresh failure to a WARNING when prior content is "usable." The design adds a provenance-presence assertion but does not close the gap where `unusableRequiredBuiltinPackNames` returns names but the load proceeds anyway with a warning.

**Required change:** The design must state that if `unusableRequiredBuiltinPackNames` returns non-empty after `MaterializeBuiltinPacks` succeeds (or after a refresh attempt), the production load MUST fail, not warn. This is the "loud failure" contract. Provenance-presence is defense-in-depth for the "materialized but dropped from config" case; manifest-integrity is the "partial/corrupt" gate. Both must be fatal on the normal path.

### [Blocker] Two production include-builder families have divergent failure semantics, and the highest-value path uses the weaker one

`cityConfigIncludesWithBuiltinPacks` (`cmd_config.go:29-38`) calls `MaterializeBuiltinPacks` directly and hard-fails only if that returns an error. It never checks `unusableRequiredBuiltinPackNames` after a successful-but-partial materialization. This is the path used by `tryReloadConfig` (`controller.go:900-905`) — the controller's runtime re-resolution, which is the most important path in the system for long-running correctness.

`builtinPackIncludesForConfigLoad` (`embed_builtin_packs.go:83-92`) runs through `ensureBuiltinPacksReadyForConfigLoad`, which checks `unusableRequiredBuiltinPackNames` and downgrades to warning if usable content exists.

The design does not reconcile these. An implementer following the document literally would add the provenance assertion to both wrappers, but the controller reload path would still lack manifest-integrity enforcement because `cityConfigIncludesWithBuiltinPacks` never runs `unusableRequiredBuiltinPackNames`.

**Required change:** Both include-builder families must enforce identical Core semantics. The design must name `tryReloadConfig` explicitly and require either: (a) it routes through the same integrity-checked path as `builtinPackIncludesForConfigLoad`, or (b) it runs a post-load assertion that includes both provenance-presence and manifest-integrity, with focused tests for each failure mode.

### [Blocker] The design's wrapper list is incomplete and two named wrappers cannot host the post-load assertion

The design names three wrappers: `loadCityConfigWithBuiltinPacks`, `cityConfigIncludesWithBuiltinPacks`, and `builtinPackIncludesForConfigLoad`. Two of these return `[]string` includes — they have no `*config.Provenance` to assert against. The assertion must live in the config-returning functions, not in the include-building functions.

The actual config-returning entry points in production `cmd/gc/` that load full config are:

1. `loadCityConfig` (`cmd_agent.go:39-48`) — used by most agent commands
2. `loadCityConfigFS` (`cmd_agent.go:55-64`) — testable variant
3. `loadCityConfigWithoutBuiltinPackRefreshFS` / `loadCityConfigWithoutBuiltinPackRefresh` (`cmd_agent.go:66-80,83-86`) — used by completion and stop
4. `loadCityConfigWithBuiltinPacks` (`cmd_config.go:21-26`) — used by config and start commands
5. `loadConfigCommandCityConfig` (`cmd_config.go:16-18`) — config show/explain
6. `tryReloadConfig` (`controller.go:900-905`) — controller reload

Each of these must be classified: (a) must run the assertion, (b) is an allowed partial-read exception, or (c) must be migrated. The design does not provide this classification.

**Required change:** Replace the three-name wrapper list with a complete classification of every config-returning entry point in `cmd/gc/`. For each, state whether the post-load assertion runs, whether it is a partial-read exception with justification, or whether it must be migrated to a wrapper that runs the assertion.

### [Major] `effectiveCityName` materializes packs but does not include them — a concrete existing bypass bug

`effectiveCityName` (`cmd_supervisor_city.go:91-97`) calls `MaterializeBuiltinPacks` and then `config.LoadWithIncludes` WITHOUT the builtin pack includes. This is exactly the pattern the design should prevent: materialize to disk, then resolve config without the system packs, silently producing a Core-missing config. The design's scanner must catch this, but the design does not mention this specific site.

**Required change:** Classify `effectiveCityName` explicitly. If it only reads `city.Name` from root config, it may be an allowed partial-read exception — but that justification must be documented in the scanner allowlist with a focused test. If it drives any agent/session behavior, it must be migrated.

### [Major] The `config.Load` (no-includes) surface is unaddressed and larger than `config.LoadWithIncludes`

There are ~13 direct `config.Load` calls in non-test `cmd/gc/` that resolve config without any pack expansion. The design's scanner proposal only mentions `config.LoadWithIncludes`. Several `config.Load` sites serve behavior-driving code:

- `cmd/gc/api_state.go:767` — `RawConfig()` serves live config to API handlers without pack expansion.
- `cmd/gc/apiroute.go:34,80` — port discovery from raw config, but the same `RawConfig` function may serve other API reads.
- `cmd/gc/doctor_v2_checks.go:488,1130` — doctor checks reading raw config for import-state and other diagnostics.
- `cmd/gc/cmd_register.go:105` — registration reading raw config.
- `cmd/gc/cmd_start_drift.go:390` — drift detection reading raw config.

`config.Load` intentionally skips pack expansion (the Go doc says "Load intentionally skips include and pack expansion"). Any production path that uses `config.Load` to drive agent/session/order behavior is a Core bypass, even if it doesn't use `config.LoadWithIncludes`.

**Required change:** The scanner must ban direct `config.Load` in production `cmd/gc/` files outside an explicit allowlist, just like `config.LoadWithIncludes`. The design must enumerate the ~13 call sites and classify each.

### [Major] The no-refresh path `loadCityConfigWithoutBuiltinPackRefreshFS` is underspecified

This function (`cmd_agent.go:66-80`) skips `MaterializeBuiltinPacks` entirely, using whatever is on disk. It is called from completion (`completion.go:183,210`) and stop (`cmd_stop.go:395`). The design says "no-refresh or diagnostic config helpers must either run the same Core provenance assertion after loading or be documented as partial-read exceptions" but does not classify these callers.

Completion reads agent definitions (which may come from Core-provided packs). If Core is absent from disk, completion silently shows a Core-missing config to the user. Stop reads agent pools for shutdown targeting.

**Required change:** Classify each caller of `loadCityConfigWithoutBuiltinPackRefresh`:
- Completion: either run the assertion (fail loud if Core is absent) or document as partial-read with a focused test proving completion cannot dispatch formulas/orders with stale or absent Core.
- Stop: classify whether stop reads agent pools from config. If so, it needs Core included.

### [Minor] Path normalization for provenance comparison is unspecified

The provenance assertion and the Core doctor both need to compare "the materialized Core system pack path" against entries in `prov.Sources`. The codebase already uses `normalizePathForCompare` (`embed_builtin_packs.go`) for path comparison. The design does not address path normalization (absolute vs. relative, symlinks, trailing slashes) for the assertion.

**Required change:** State that the assertion must use the same path normalization as `normalizePathForCompare`, or define a canonical path form for the comparison.

### [Minor] Empty-pack provenance edge case

If a Core system pack contributes zero config keys (all its agents, orders, formulas are overridden by user packs), does `prov.Sources` still contain the Core path? The design assumes yes, but the code adds to `prov.Sources` at specific composition stages (line 279 for pack.toml, line 405 for fragments, line 485 for implicit imports). Builtin pack includes are injected via `packIncludes` at line ~430, and each included pack's `pack.toml` would be tracked via the pack expansion loop. If a pack contributes nothing after merge, its path is still in `Sources` because `Sources` records file provenance, not effective contribution. This should be verified, not assumed.

**Required change:** Add a test that a Core pack with all keys overridden by user packs still appears in `prov.Sources`, or document the verified behavior.

### [Minor] Scanner allowlist must cover indirect calls and function aliases

The design's scanner modeled on `TestGCNonTestFilesStayOnWorkerBoundary` must also catch:
- Helper functions that wrap `config.Load` or `config.LoadWithIncludes` and are themselves called from production code (e.g., `effectiveCityName`, `RawConfig`).
- `config.LoadCity` and other `config.Load*` variants.

**Required change:** The scanner should reject any function in non-test `cmd/gc/*.go` that calls `config.Load`, `config.LoadCity`, or `config.LoadWithIncludes` unless the function is in the allowlist. The allowlist must include the calling function, not just the file.

## Cross-Review Observations

- **Claude correctly identified the provenance-presence vs. manifest-integrity gap and the two-wrapper divergence.** I agree these are the two central structural weaknesses. Claude's analysis of the `materializeBuiltinPacksForConfigLoad` warning-downgrade is precise and correct.
- **Gemini correctly identified the `config.Load` bypass surface.** This is a gap the design and other reviews underweight. My count confirms ~13 direct `config.Load` calls in production `cmd/gc/`, several of which drive behavior.
- **Codex's concern about the provenance data model is valid.** The design must specify what field(s) in `Provenance` carry the assertion signal. Relying on `Sources []string` works for path-presence but not for integrity; the design should pair provenance-presence with manifest-integrity.
- **All three prior reviews agree the scanner/allowlist is the right enforcement shape but insufficiently scoped.** My assessment: the scanner must cover both `config.Load` and `config.LoadWithIncludes`, and the allowlist must include calling function, classification, reason, and focused test — not just file-level exceptions.

## Required Changes (Priority Order)

1. **Define the loud-failure contract as manifest-integrity (fatal) + provenance-presence (defense-in-depth).** State explicitly: `unusableRequiredBuiltinPackNames` returning non-empty after materialization is a fatal error on the production path, not a warning. Provenance-presence catches the "include built but Core dropped from config" case.
2. **Reconcile the warning-downgrade in `materializeBuiltinPacksForConfigLoad`.** Decide: stale-but-usable Core on a failed refresh is either a hard failure (recommended for required packs) or a warning (acceptable only for non-required packs). The design currently implies both. Apply the decision uniformly to both include-builder families.
3. **Name `tryReloadConfig` as in-scope.** The controller reload must run both manifest-integrity and provenance assertions, or be allowlisted with a focused test proving the exemption is safe.
4. **Replace the three-name wrapper list with a complete classification of all six config-returning entry points** (`loadCityConfig`, `loadCityConfigFS`, `loadCityConfigWithoutBuiltinPackRefreshFS`, `loadCityConfigWithBuiltinPacks`, `loadConfigCommandCityConfig`, `tryReloadConfig`). For each: assertion required, partial-read exception (with justification and test), or migrate-to-wrapper.
5. **Classify `effectiveCityName` explicitly.** It is a concrete existing bypass that materializes packs but does not include them.
6. **Extend the scanner to cover `config.Load` as well as `config.LoadWithIncludes`.** Enumerate all ~13 `config.Load` sites, classify each, and include calling-function-level allowlist entries.
7. **Classify the no-refresh path callers** (completion, stop) and either require the assertion or document the partial-read exception with a focused test.
8. **Specify path normalization** for the provenance comparison, referencing the existing `normalizePathForCompare`.
9. **Verify empty-pack provenance behavior**: add a test that a Core pack contributing zero effective keys still appears in `prov.Sources`, or document the verified behavior.

## Questions

1. For a present-but-stale Core after a failed refresh, is the intended behavior hard-fail or warn-and-proceed? The design implies both. (Same as Claude's Q1 — this is the central ambiguity.)
2. Should `tryReloadConfig` route through the integrity-checked `builtinPackIncludesForConfigLoad` path, or should it run an independent assertion after loading? The design does not specify.
3. Should `loadCityConfigWithoutBuiltinPackRefreshFS` remain in production after Core becomes mandatory? If so, which config fields may safely be read from a config that might lack Core overlays, and what prevents future callers from using it for behavior-driving reads?
4. Does `prov.Sources` include a pack path when the pack contributes zero effective config keys? The provenance assertion's correctness depends on the answer.
5. Is the loud-failure guarantee meant to cover *extra* unexpected files in `.gc/system/packs/core`? `packContainsEmbeddedManifest` iterates only embedded entries; `validatePackFiles` in `registry.go` does sweep for extra files but only for synthetic cache validation, not for the required-pack integrity check.
