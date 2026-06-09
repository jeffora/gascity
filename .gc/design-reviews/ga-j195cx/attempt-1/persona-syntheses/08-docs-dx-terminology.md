# Avery Brooks

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- **Blocker:** The pack/config documentation surface is not specified tightly enough for copy-paste-safe guidance. Claude found `[[pack.requires]]` named as a real parallel surface without a TOML example, parser ownership pointer, doctest coverage, or clear relationship to `[pack].requires_gc`; Codex found the required pack examples omit the required `[pack].schema` key, so the "runnable" examples are invalid under the existing pack loader. Together this means the design does not yet safely teach the distinction between formula `[requires]`, pack dependency declarations, pack GC compatibility floors, schema, and pack revision.
- **Blocker:** Warning transport is unresolved for user-visible deprecation and misuse diagnostics. Claude found surface policy only defined for `formula.version_deprecated`, while `formula.contract_deprecated` and `formula.version_misuse` lack a per-entry-point visibility contract. Authors can therefore hit fatal launch behavior without a reliable path to see the warning/remediation guidance that explains the terminology shift.
- **Major:** The docs gate coverage is internally inconsistent. Claude found no checked-in prose/snippet inventory with per-passage migration outcomes, and Codex found the stale-guidance matcher claims to protect TOML examples but scopes the `requires-positive-content` rule only to Markdown paths. The result can pass or fail in ways that do not match the stated docs/DX guarantee.
- **Major:** Several release-gate CLI/doc surfaces are referenced but not specified. Claude called out `--legacy-contract-report`, `--provenance`, `--requires-only-conversion-gate`, `--alias-removal-gate`, `--compat-corpus`, and `--all-packs` as required inputs without JSON shapes, exit-code contracts, or composition rules.
- **Major:** Canonical examples still contain unresolved placeholder policy. Claude found `requires_gc = ">=<minimum-floor-from-formula-compiler-min-floor.json>"` used in canonical-looking pack examples while the stale-guidance gate also bans that placeholder in release artifacts; Codex separately requires paired resolved fixtures when templates are shown.
- **Minor:** The design uses inconsistent terminology it intends to police elsewhere: "Formula v2", "formula-v2", "formula v2", and "compiler capability 2" appear with mixed meanings, and `[daemon] graph_workflows` is treated as user-visible legacy surface by stale-guidance rules but omitted from the glossary.
- **Minor:** The architecture-doc update requirement is under-specified. Claude found `engdocs/architecture/formulas.md` still teaches `version` as an optional marker and contains stale current-state framing; the design only requires adding new ownership content, not removing or reclassifying the conflicting text.

**Disagreements:**
- Claude chose `approve-with-risks` while Codex chose `block`. I assess Codex's verdict as the right persona outcome because the invalid pack examples and unspecified `[[pack.requires]]` surface directly violate the design's own docs/DX promises: readers would be given non-runnable or ambiguous TOML at the exact point the migration is supposed to reduce terminology confusion.
- Codex focused on pack snippet validity and matcher scope; Claude focused on broader glossary, warning transport, CLI flag, inventory, and placeholder risks. These are complementary, not contradictory. The combined signal is that the design has the right documentation intent but lacks enough executable contracts to keep docs, examples, diagnostics, and release gates aligned.

**Missing evidence:**
- No Kimi 2.6 review was present for this persona.
- No positive TOML example, doctest fixture, or PackV2 reference defines `[[pack.requires]]` syntax or states whether it is current or forward-looking.
- No proof that pack and city TOML snippets embedded in `docs/reference/formula.md` will be validated by the correct config loaders instead of only by the formula loader.
- No per-warning-code visibility matrix for `formula.contract_deprecated`, `formula.version_deprecated`, and `formula.version_misuse` across CLI, API, dashboard, controller logs, order events, and release validation.
- No docs/prose inventory enumerates existing passages that teach `contract` or `version` as canonical, including the stale `engdocs/architecture/formulas.md` language.
- No specified JSON shape, exit-code semantics, or example output for the release-gate CLI flags the rollout depends on.

**Required changes:**
- Make the pack compatibility examples valid and copy-paste-safe: include required pack metadata such as `schema = 2`, classify each TOML block by snippet kind/path, and validate formula snippets with the formula loader, pack snippets with the pack/config loader, and city snippets with the city config loader. If a template placeholder remains, pair it with a resolved runnable fixture.
- Define `[[pack.requires]]` completely or remove it from the glossary, common-confusion table, and requirement-surface comparison until a separate design owns it. A complete definition must include syntax, parser ownership, positive example, and doctest/config validation coverage.
- Extend the diagnostic visibility matrix and per-entry-point table with explicit rows for every warning code, especially `formula.contract_deprecated`, `formula.version_deprecated`, and `formula.version_misuse`.
- Specify the `gc formula validate` release-gate flags: input scope, JSON output shape, exit-code contract, and how the flags compose with each other.
- Add a checked prose/snippet inventory artifact for all docs, tutorials, architecture docs, examples, and fixtures that currently teach `contract` or `version`, with a migration outcome for each passage. Include `engdocs/architecture/formulas.md` and make the inventory a hard input to `make formula-docs-check`.
- Fix stale-guidance matcher scope so TOML-bearing examples/testdata are covered by the rule it claims to enforce, or split prose and parsed-TOML checks into separate named reports.
- Decide the placeholder policy for canonical examples and remove ambiguous placeholders from reader-facing runnable snippets.
- Normalize terminology in the design body: use "compiler capability 2" for executable contracts, reserve "Formula v2" only for explicit shorthand, and add `[daemon] graph_workflows` to the glossary as a deprecated host alias.
- Tighten the architecture-doc update requirement so stale "Optional formula version marker", old verification framing, and `bd` shell-out current-state prose are removed or reclassified in the same change.
