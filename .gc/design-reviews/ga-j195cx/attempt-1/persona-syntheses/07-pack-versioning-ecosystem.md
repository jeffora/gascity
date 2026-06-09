# Saoirse Raman

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Info] Both reviews agree the design now rejects formula-level artifact semver as the reproducibility boundary. Pack revision is the artifact identity boundary, legacy formula `version` is inert metadata rather than a compiler selector, and stale-guidance checks are intended to keep docs from reviving formula semver semantics.
- [Info] Both reviews agree the external-author direction is sound: `gc formula validate`, provenance output, requirement diffing, legacy-contract reports, external-support evidence, and typed `migration_hints` give pack authors and release gates structured data instead of prose-only diagnostics.
- [Major] Resolver-to-compiler provenance remains load-bearing. Claude requires `ResolvedFormulaSource` threading to be an explicit Phase 2 deliverable with rollback, field-source mapping, and tests; Codex separately warns that Packman schema 2 or an equivalent typed provenance API must block later pack-floor and alias-removal phases if it slips.
- [Major] The author-facing validation/lint surface must be pinned to rollout timing. Claude finds `gc formula validate --pack-path` and `--pack-source --ref` well-specified but not assigned to a phase, CLI-doc update row, diagnostic-visibility predecessor, or explicit lint-surface sentence. Codex's pass depends on this command existing before any user-visible requirement diagnostic lands.
- [Minor] `[pack] requires_gc` comparator ownership needs a cleaner boundary. The design specifies comparator behavior while saying Packman owns it, so enforcement should wait for a ratified Packman contract and should not ship in the same PR as the first comparator or schema-2 implementation.
- [Minor] Some ecosystem fields and migration hooks are forward-looking without committed consumers. `safe_automatic_edit`, `registry_query`, and `registry_mirror` are acceptable only if documented as forward-compatible hooks with named follow-up work, or removed until they have owners.
- [Minor] Legacy formula `version` warning behavior needs an explicit policy. Suppressing `formula.version_deprecated` on launch surfaces is defensible, but the design should either state that merely stale `version` is acceptable forever or emit a once-per-launch warning.

**Disagreements:**
- Codex returns `PASS` with no required changes, while Claude returns `approve-with-risks` with concrete required changes. Assessment: choose `approve-with-risks`. Codex found no blockers in the architecture direction, but Claude identified rollout and ownership gaps that should be fixed before approval is treated as clean.
- Codex says the pack-author migration surface is sufficiently explicit. Claude agrees on the schema quality but finds the command timing and documentation placement underspecified. Assessment: adopt Claude's narrower requirement because it is a traceability gap, not a disagreement about the design direction.
- Codex treats Packman schema 2 / typed provenance as a watchpoint for later phases. Claude requires resolver-to-compiler provenance plumbing to be named in Phase 2. Assessment: these are complementary; Phase 2 can thread the schema-1/resolver fields now, while schema-2-dependent fields remain explicitly gated.
- Codex does not require changes for legacy formula `version` policy, migration-tool follow-up, outreach channel, builtin-root identity, or static guards. Assessment: these are not blockers, but they are valid missing-evidence items from the pack-versioning ecosystem lane.

**Missing evidence:**
- Kimi 2.6 review input was not present for this persona.
- A Phase 2 rollout item proving real workflow roots receive non-empty `PackProvenance.Source` and either `LockedRevision` or a documented local-VCS reproducibility value.
- A table separating `ResolvedFormulaSource` fields available from current schema-1 Packman plus resolver state from fields blocked on schema 2.
- A specific release phase for `gc formula validate --pack-path` and `--pack-source --ref`, plus matching CLI-doc, command-table, and diagnostic-visibility updates.
- A Packman-owned comparator contract for release, prerelease, nightly, source-build, dirty, and unknown active-binary versions.
- External-author guidance for the parse-but-do-not-enforce `requires_gc` window, especially that external packs should keep dual declarations through the alias window.
- A named outreach channel and owner for `public_notice_url`, `support_request`, and `manual_report` rows in the external-support artifact.
- A definition of builtin-pack root identity: what `gc.formula_pack_revision` means for `source_class = builtin`, whether the producing Gas City build is stamped, and how repair behaves after binary downgrade.
- A static guard such as `TestFormulaVersionIsNotABehavioralInput` proving production code does not branch on legacy formula `Version` outside parser/reporting paths.

**Required changes:**
- Name resolver-to-compiler `ResolvedFormulaSource` threading as an explicit Phase 2 deliverable, with rollback unit, schema-1 versus schema-2 field mapping, and an acceptance test for non-empty root provenance.
- Pin `gc formula validate --pack-path` and `gc formula validate --pack-source --ref` to a rollout phase no later than the first user-visible `[requires]` diagnostic. Add them to the diagnostic-visibility predecessor table, `docs/reference/cli.md` required-update row, and per-phase command table. Name `gc formula validate` as the canonical pack-author lint surface, or add a `gc formula lint` alias.
- Move `[pack] requires_gc` comparator semantics to a Packman-owned document, or mark them as proposed requirements pending Packman ratification. State that Phase 7b enforcement cannot ship in the same PR as the first Packman comparator implementation.
- Add a Strict Review Risk Register row or committed follow-up for first-party migration tooling before requires-only conversion. Document `safe_automatic_edit`, `registry_query`, and `registry_mirror` as forward-compatible hooks with owners, or remove them until they have consumers.
- State the external-pack outreach channel and the owner responsible for `support_request` / `manual_report` evidence rows.
- Add a compatibility-matrix row for external SHA-pinned requires-only graph formulas on binaries that parse but do not enforce `[pack] requires_gc`, with explicit dual-declaration guidance through the alias window.
- Either state that merely stale legacy `version` is acceptable forever, or emit `formula.version_deprecated` once per launch invocation through the same per-process LRU pattern used for `formula.contract_deprecated`.
- Add a static guard such as `TestFormulaVersionIsNotABehavioralInput` so legacy formula `Version` cannot become a behavioral input outside the parser, raw scanner, and legacy-version-report code.
