# Avery Brooks

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] The docs rollout gate is directionally correct and must remain hard-blocking. Both reviews found the design appropriately blocks user-visible diagnostics until reference docs, examples, generated artifacts, stale-guidance checks, and dashboard/API surfaces are updated in the same PR stack.
- [Major] Existing docs still teach stale or contradictory concepts. `docs/reference/formula.md` still presents `version` as a normal top-level formula key, `engdocs/architecture/formulas.md` still describes the old `bd`/`MolCook` production path, and `engdocs/proposals/formula-migration.md` still frames `GC_NATIVE_FORMULA=false` as a production rollback path. These must be updated, explicitly superseded, or scoped as historical before diagnostics ship.
- [Major] First-party formula coverage is not yet proven. The design requires graph formulas, examples, tutorials, and fixtures to be inventoried and classified for dual declaration, v1 rewrite, or stale-fixture deletion, but neither review found a checked-in inventory or CI target that proves complete coverage.
- [Major] The `requires` terminology risks confusing formula-level `[requires] formula_compiler` with pack-level `[pack].requires_gc`. The glossary and rewrite skeleton need a side-by-side TOML example and clearer wording so authors do not invent a nonexistent pack `[requires]` key.
- [Major] `version` deprecation guidance is not concrete enough. Reference docs must remove `version` from canonical examples and top-level key tables, stale-guidance matchers must target real TOML/table patterns rather than broad prose mentions, and `formula.version_deprecated` remediation should tell users to remove the field and use pack revision/ref/SHA as the artifact identity boundary.
- [Minor] Release-oriented docs and diagnostics examples still contain wording or placeholders that could leak into shippable artifacts. Examples include unresolved `<minimum-floor...>` placeholders and inline remediation strings that look like copy-paste TOML but are not valid TOML.

**Disagreements:**
- No material verdict disagreement: Claude and Codex both returned `approve-with-risks`.
- Claude treated several docs/DX issues as separate major risks, including the misleading "Pack `requires` / `requires_gc`" glossary wording, missing supersession banner for the migration proposal, and missing native-compiler rewrite of the architecture doc. Codex grouped those under a broader "live docs still teach the old surface" risk. Assessment: keep the grouped consensus finding, but carry Claude's concrete required changes because they are actionable and reduce author confusion.
- Claude flagged diagnostic remediation string formatting as a minor issue; Codex only raised unresolved placeholders as minor. Assessment: both are valid shippability checks for docs/DX polish, so include them together as minor required cleanup.

**Missing evidence:**
- No generated stale-guidance config or CI output proves docs/examples scanning is active.
- No checked-in first-party formula inventory covers built-in packs, `.gc/system/packs`, examples, tutorials, and fixtures.
- No generated help, config reference, OpenAPI/dashboard type, formula reference, or migration-hint update is present yet.
- No explicit policy says whether `formula.version_deprecated` fires for all `version =` values or only selected legacy values.
- No clear external-pack-author guidance says whether authors should use requires-only syntax or dual declarations during the alias window.
- No operator-facing walkthrough shows the failure path when `[pack].requires_gc` exceeds the active binary.

**Required changes:**
- Update the `docs/reference/formula.md` rewrite contract so the minimal example omits `version`, the current `version` top-level-key row moves to a legacy-fields section naming `formula.version_deprecated`, and CI asserts the rewritten doc contains `[requires] formula_compiler` while rejecting canonical `version =` examples and the phrase "Optional formula version marker."
- Fix the glossary to name only pack `[pack].requires_gc`, add a distinct entry for the formula `[requires]` table, and add one combined TOML example that shows pack binary compatibility and formula compiler capability side by side.
- Explicitly update or supersede `engdocs/proposals/formula-migration.md` and `engdocs/architecture/formulas.md` in the same PR stack, including a top-of-file supersession/status marker for the proposal and native-compiler-as-production-path language in the architecture doc.
- Add a checked-in first-party formula inventory under `docs/release/` with one row per relevant pack/example/tutorial/fixture and a blocking assertion that every row has a classification.
- Tighten stale-guidance checks for `version` so TOML globs match `^version\s*=` and Markdown table checks match the top-level-key row, with legacy-section allowances only where intentional.
- Make `formula.version_deprecated` remediation directive: remove the `version` field and use the containing pack revision/ref/SHA as the artifact identity boundary.
- Add a release/docs check that rejects unresolved `<minimum-floor...>` placeholders in shippable docs, generated migration hints, and release report examples after `formula-compiler-min-floor.json` exists.
- Normalize diagnostic `Remediation` strings so they are either valid multi-line TOML snippets or plain prose, not inline shorthand that appears copy-pasteable but does not parse.
