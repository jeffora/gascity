# Saoirse Raman

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The design depends on pack revision provenance, but it does not define a durable formula-source provenance data model through layer resolution, compilation, and workflow-root creation. Both reviews identify the same failure mode: current path-oriented layer winners and staged symlinks can lose the pack binding, import source, requested ref or version, locked commit or content hash, transitive import edge, local dirty state, and layer priority needed to make `gc.formula_pack_*`, `gc.formula_reproducibility`, `CompileResult.Provenance`, and `gc formula validate --provenance` truthful months later.
- [Major] External pack-author migration is still underspecified. The design protects legacy `contract = "graph.v2"` through an alias window in principle, but it does not provide a standalone validation workflow for an external pack directory, remote source/ref, or SHA-pinned pack before publication, and the "documented compatibility branch" fallback has no owner, opt-in mechanism, support matrix, duration, or communication channel.
- [Major] Provenance and validation command contracts are uneven. Claude notes that `--legacy-contract-report` has a stronger schema and exit-code story than `--provenance`; Codex adds that `--all-packs --legacy-contract-report` still needs clearer scan scope, imported-pack classification, per-item pack fields, and command behavior for external authors. Pack-author CI cannot depend on prose-only provenance output.
- [Major] Legacy `version` handling can recreate the formula-semver confusion the design is trying to remove. Both reviews accept preserving it during migration, but the design needs canonical author guidance, accepted-value bounds, compatibility-matrix rows, stale-doc cleanup, warning or strict-mode behavior, and measurable criteria for eventual rejection or stripping.
- [Major] Metadata normalization rules for pack revision identity are missing. The design must say how semver constraints, `sha:` pins, `source#ref` imports, GitHub tree URLs, lockfile commits, transitive imports, and local path imports map to `gc.formula_pack_source`, `gc.formula_pack_ref`, `gc.formula_pack_revision`, and reproducibility status.
- [Minor] Reproducibility markers have too little operational effect. `local-not-reproducibly-pinned` is useful only if CI, launch policy, dashboard display, or another consumer can act on it rather than merely record it.
- [Minor] Pack composition and warning routing need clearer behavior. If pack A imports pack B and B owns the legacy declaration, the operator needs to know which pack receives the diagnostic, how imported-only findings are classified, and how warnings can be scoped while waiting on the upstream pack.

**Disagreements:**
- Claude's overall verdict is `approve-with-risks`; Codex's verdict is `block`. Assessment: block. The provenance gap is not a polish issue for this persona; without a concrete data model from formula-layer resolution to root metadata or a linked compile artifact, the central pack-versioning promise cannot be audited.
- Claude treats the existing auditable-provenance table as the right data-model shape on paper, while Codex argues the current formula layer model cannot populate it reliably. Assessment: the table is directionally right, but the design must add the missing data flow and types before approval.
- Claude emphasizes external-author alias-removal gates, migration tooling, and compatibility-branch definition; Codex emphasizes source/ref normalization and standalone external pack validation. Assessment: these are complementary requirements for the same ecosystem risk.
- Claude asks for explicit operational consumers for `local-not-reproducibly-pinned`; Codex focuses on whether local paths can ever be considered pinned. Assessment: first define the reproducibility taxonomy, then require at least one consumer that makes non-reproducible inputs visible or enforceable.

**Missing evidence:**
- No proposed `FormulaLayer`, `FormulaSource`, provenance index, or equivalent structure that replaces or augments raw directory slices and symlink-only winners.
- No defined `Provenance` type alongside `NormalizedRequirements` and `Diagnostic`, including where it is persisted and how callers consume it.
- No worked provenance examples for local clean packs, local dirty packs, semver imports resolved through `packs.lock`, SHA-pinned imports, `source#ref` imports, GitHub tree URLs, and transitive imported-pack winners.
- No end-to-end external SHA-pinned legacy `contract` example through the alias window and after alias removal, including the diagnostic, binary version, and remediation.
- No standalone external pack validation command contract that does not require the pack to already be imported into a consuming city.
- No clear statement whether `--all-packs` scans first-party packs only, imported third-party packs in the resolved layer set, or both, and how that affects alias-removal gates.
- No accepted-value bounds, warning/removal horizon, strict behavior, or compatibility-matrix rows for legacy `version`.
- No normalization rules for source identity, requested ref, resolved revision, local content hash, dirty state, or reproducibility status.

**Required changes:**
- Add a structured formula-source provenance design covering `ComputeFormulaLayers`, formula resolution, `CompileWithResult`, and workflow-root metadata or a durable compile artifact linked from the root. It must preserve formula identity, winning source path, layer priority, pack binding, import source, requested ref or version, locked revision or content hash, local dirty status, and transitive import attribution.
- Define exact normalization rules for `gc.formula_pack_source`, `gc.formula_pack_ref`, `gc.formula_pack_revision`, and `gc.formula_reproducibility` across local paths, semver imports, `sha:` pins, `source#ref` imports, GitHub tree URLs, lockfile-backed remote imports, and transitive imports.
- Specify `gc formula validate --provenance` with JSON schema, sample output, exit codes, and CI stability guarantees, parallel to the stronger `--legacy-contract-report` contract.
- Specify an external pack-author validation path, such as `gc formula validate --pack <path-or-ref> --legacy-contract-report --provenance`, with input assumptions, JSON output, exit codes, examples, and behavior for packs that are not yet imported by a city.
- Clarify `--legacy-contract-report` scan scope and accounting: first-party versus external packs, imported-only legacy usage, per-item pack fields, separate release-gate signals where needed, and whether `version` deprecations are included.
- Replace "documented compatibility branch" with a concrete external-pack support plan covering owner, supported `gc` versions, consumer opt-in mechanism, removal criteria, notice period or measurable adoption signal, and external-author communication channel.
- Add migration help for pack authors, at minimum a documented `--fix` or equivalent path from `contract = "graph.v2"` to `[requires] formula_compiler = ">=2"` with dual-declaration behavior during the alias window and TOML formatting/comment preservation expectations.
- Bound legacy `version`: canonical new-author guidance, accepted values, diagnostic text, docs/example updates, stale-guidance CI guard, compatibility-matrix rows, strict-mode behavior, and criteria for eventual hard error or removal.
- Specify operational consequences for non-reproducible inputs, such as a CI `--require-reproducible` mode, dashboard warning, or launch policy toggle.
- Add required tests for provenance on layer winners, including city pack versus local formula shadowing, rig pack winners, transitive import winners, lockfile-backed remote imports, local dirty imports, and legacy-contract report per-item pack fields.
