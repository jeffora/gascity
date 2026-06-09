# Saoirse Raman - Codex

**Verdict:** block

**Top strengths:**
- The design is now clear that formula files do not carry independent artifact semver; pack version, ref, lock entry, or local content hash is the intended revision boundary.
- The compatibility window for legacy `contract = "graph.v2"` is measurable rather than calendar-based, and it explicitly keeps external SHA-pinned legacy formulas valid through the alias window.
- The proposed validate/report surfaces are pointed at the right pack-author problem: stale legacy declarations should be discoverable before first-party packs become requires-only.

**Critical risks:**
- [Blocker] The provenance contract is specified only at the output metadata layer, but the current formula layer model is path-only. `config.FormulaLayers` carries `[]string` directories, and `cmd/gc/formula_resolve.go` reduces winners to formula name -> source path symlinks. That loses the pack binding, import source, requested ref/version, lock commit, transitive import edge, and local dirty/hash state required by `gc.formula_pack_*`, `gc.formula_reproducibility`, `CompileResult.Provenance`, and `gc formula validate --all-packs --provenance`. Without an explicit data model and flow for formula-source provenance, layer winners cannot be tied back to the pack revision boundary the design relies on.
- [Major] The external pack-author workflow is still underspecified. `gc formula validate --all-packs --legacy-contract-report` is a release gate for the current city, but the design does not define how an external author validates an arbitrary pack directory, remote source/ref, or SHA-pinned pack before publishing or before a consumer imports it. That leaves third-party authors discovering stale `contract` snippets only after a city import/load path fails or after reading changelog prose.
- [Major] The design does not normalize PackV2's multiple revision inputs into the proposed metadata fields. `[imports.<name>].version` may be a semver constraint or `sha:<commit>`, remote source strings can carry `#ref`, GitHub tree URLs encode refs in the path, and `packs.lock` records the resolved commit. The design needs to say which value goes in `gc.formula_pack_ref`, which value goes in `gc.formula_pack_revision`, and how this works for transitive/re-exported packs and local path imports.

**Missing evidence:**
- No proposed `FormulaLayer`/`FormulaSource`/provenance index shape that replaces or augments `FormulaLayers`' raw directory slices.
- No example provenance records for a local clean pack, local dirty pack, semver import resolved through `packs.lock`, SHA-pinned import, and transitive imported pack whose formula wins a layer.
- No command contract for validating a standalone external pack without a fully initialized consuming city.

**Required changes:**
- Add a structured formula-source provenance design: how `ComputeFormulaLayers`, formula resolution, `CompileWithResult`, and root metadata carry pack binding, import source, requested ref/version, locked revision or content hash, layer priority, and winning source path.
- Define exact metadata normalization rules for `gc.formula_pack_source`, `gc.formula_pack_ref`, `gc.formula_pack_revision`, and `gc.formula_reproducibility` across local paths, semver imports, `sha:` pins, `source#ref` imports, GitHub tree URLs, and transitive imports.
- Specify an external author validation path, such as `gc formula validate --pack <path-or-ref> --legacy-contract-report --provenance`, with JSON shape, exit codes, and examples that do not require a consumer city to already have imported the pack.
- Add required tests for provenance on layer winners, including city pack vs local formula shadowing, rig pack winners, transitive import winners, lockfile-backed remote imports, local dirty imports, and the legacy-contract report's per-item pack fields.

**Questions:**
- Is `gc.formula_pack_ref` intended to preserve the user-requested constraint/ref, while `gc.formula_pack_revision` stores the resolved lock commit or local content hash, or should those be split into more explicit `requested_ref` and `resolved_revision` fields?
- Should local path packs be allowed to produce `pinned` provenance when clean under VCS with a commit hash, or are all local paths considered non-reproducibly pinned unless imported through a lockable remote source?
- Should validation operate on original formula layer directories rather than `.beads/formulas` symlinks so provenance cannot be lost before compile time?
