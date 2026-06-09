# Saoirse Raman

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] The design correctly anchors reproducibility at the pack revision boundary rather than formula-level `version`. Both reviews agree the current shape names the right durable identity surfaces: pack source/ref/revision, lockfile key, binding identity, content hash, root provenance metadata, and compile artifact references.
- [Major] External pack validation and migration surfaces are directionally strong, but the read-only `--pack-source --offline` contract is still underspecified. The reviews converge that external CI and release gates need one byte-level rule for cache lookup, immutable lock/cache hits, content-hash verification, ambiguous entries, cache misses, stale cache data, and whether city-local installed-pack cache is readable but never mutated.
- [Major] External author migration still needs a worked, executable walkthrough. Both reviews call for evidence that `contract = "graph.v2"`, `formula.migration.add_requires`, `formula.migration.raise_pack_requires_gc`, `[pack] requires_gc`, `--requirement-diff`, republish, and consumer lockfile update compose for SHA-pinned packs and mutable tag/branch imports.
- [Major] The alias-removal policy protects unknown external consumers better than earlier designs, but the ecosystem notification and timing story remains weak. Claude flags that consumer-local warnings may never reach a pack author; Codex asks for concrete external-support artifact examples. The design needs stronger evidence that external authors and SHA-pinned consumers have a realistic path before alias parsing disappears.
- [Major] The legacy `version` field remains a possible footgun for new authors. Claude identifies the risk that a new formula can still use `version = 3` as if it were semver; Codex does not dispute it. The pack-boundary docs reduce the risk, but diagnostics should make new misuse visibly wrong.
- [Minor] Release-owned artifacts and provenance persistence need sharper fixtures. Codex asks for negative schema examples for compatibility/floor/external-support artifacts; Claude asks for a deterministic `gc.formula_compile_artifact` spillover threshold and explicit behavior when lockfile schema data is not yet available.

**Disagreements:**
- There is no verdict disagreement: Claude and Codex both return `approve-with-risks`. This synthesis keeps that verdict because the pack-versioning model is conceptually sound, while treating the acquisition, migration, release-artifact, and legacy-`version` gaps as required follow-ups before implementation depends on those interfaces.
- Claude emphasizes outbound ecosystem coordination: pack-author notification, alias-window duration, abandoned maintainers, and single-file linting for first-time authors. Codex emphasizes reproducible machine contracts: offline acquisition matrix, release-artifact validation fixtures, and tag identity. Assessment: both are part of the same ecosystem contract; the design should cover author workflow and deterministic CI behavior together.
- Claude asks to decide whether the lockfile schema bump is in scope; Codex is satisfied that `internal/packman` owns lockfile semantics but still wants acquisition/provenance fixtures. Assessment: no new owner is required if `packman` remains authoritative, but the design must state schema version, fallback behavior, and which release gates may rely on persisted versus derived fields.

**Missing evidence:**
- No Gemini review artifact was present for this persona.
- No exact offline acquisition matrix for `gc formula validate --pack-source <url> --ref <ref> --all --offline --json`.
- No end-to-end external pack example covering original source, validation JSON, migration hints, pack-floor edit, republish or compatibility branch, requirement diff, and consumer lockfile update.
- No single-file or pre-pack lint entry point is named for an author validating a first formula outside a complete pack.
- No concrete initial contents or negative fixtures are shown for `docs/release/formula-compiler-external-support.md`, `formula-compiler-compatibility.yaml`, or `formula-compiler-min-floor.json`.
- No tag-resolution fixture states whether reproducibility identity is the tag object, peeled commit, fetched tree, or canonical content hash.
- No deterministic threshold is specified for switching from inline root provenance metadata to `gc.formula_compile_artifact`.
- No explicit v0 behavior is pinned for offline validation or release gates when the available lockfile schema lacks content hash, `requires_gc`, parent import, or binding fields.

**Required changes:**
- Define the read-only/offline acquisition contract for `--pack-source`: lookup order, allowed cache locations, lockfile or city-root selection, content-hash verification, ambiguous-entry behavior, cache miss/stale-cache diagnostics, exit codes, and confirmation that validation does not mutate installed-pack cache or `packs.lock`.
- Add a worked external-author migration fixture for a third-party legacy `contract = "graph.v2"` formula, including stable JSON hints, binding identity, pack-floor update, republish or compatibility-branch choice, `--requirement-diff`, and consumer lockfile/report update.
- Add release-artifact schema fixtures, including malformed and placeholder examples that block first-party requires-only conversion or alias removal.
- Add a first-formula lint path, either by documenting `--pack-path` behavior for directories without `[pack]` metadata or by adding `gc formula validate <path-to-formula.toml>`.
- Tighten the legacy `version` diagnostic policy so new misuse is visibly different from tolerated stale metadata, such as a stronger warning for unsupported/non-legacy values that points authors to `[requires]` and pack revision identity.
- Pin the `gc.formula_compile_artifact` boundary with a fixed byte budget, always-artifact rule, or equivalent deterministic producer contract, plus boundary fixtures.
- State how current and future `packs.lock` schemas carry or derive content hash, `requires_gc`, binding path, and parent binding fields, including read compatibility, write behavior, and release-gate degradation for older schemas.
- Add documentation or release-process language for external pack notification and alias-window calibration, including what happens for unknown, abandoned, or SHA-pinned external packs when alias removal is being considered.
