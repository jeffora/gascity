# Priya Zielinski

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- **Major:** The parser validation matrix is not yet a single executable contract. Claude found two competing normative schemas for `internal/formula/testdata/compiler_requirements_matrix.yaml`, plus drift between `coverage_intent` prose and `coverage_intent_id`; Codex likewise requires a machine-readable validation matrix with literal count locks before parser implementation. The design needs one schema that carries suite identity, source shape, legacy/construct/caller/host dimensions, expected diagnostics, coverage ownership, and count locks.
- **Major:** Raw TOML requirement shapes are under-specified. Both reviews require fixture-locked behavior for bracketed `[requires]`, dotted `requires.formula_compiler`, inline tables, duplicate keys, dotted/bracket collisions, unknown axes, invalid scalar types, empty tables, arrays of tables, parser-boundary failures, and exact source attribution. Claude additionally calls out that BurntSushi/toml collision behavior must be pinned to a specific version or converted into checked fixture rows, while Codex notes the current parser path can collapse distinct TOML spellings before diagnostics.
- **Major:** Misplaced requirement declarations need complete paired coverage. Claude explicitly flags missing rows for `[children.requires]`, `[loop.body.requires]`, `[compose.map.requires]`, `[compose.aspects.requires]`, and `[expand_vars.requires]`; Codex flags `[steps.requires]`, `[[steps.requires]]`, `[children.requires]`, `[loop.body.requires]`, and `[compose.requires]`. The matrix must prove these nested declarations never satisfy the root requirement and that adjacent v2-only constructs still emit `formula.compiler_requirement_missing` rather than an ambiguous type or parser error.
- **Major:** V2-only construct enforcement is conceptually sound but not fully fixture-locked across contribution and condition states. Codex requires explicit cross-products for construct identity, authored location, condition state, caller path, and source attribution. Claude identifies a direct contradiction around `expand_vars`: one rule makes compose traversal v1-compatible until contributed content contains v2-only constructs, while another treats `expand_vars` on compose expansion/map vars as v2-only. The design must choose one policy and update registry, traversal rules, and fixtures together.
- **Major:** Diagnostic precedence for combined defects is too coarse. Codex requires rows that lock whether diagnostics are accumulated or suppressed, their order, fatal/warning split, and how parser-boundary errors interact with semantic `formula.*` diagnostics. Claude's schema concern reinforces this: the final row model must be rich enough to express diagnostic source path, line, column, count, and ordering without relying on prose.
- **Major:** The decoded-field classification guard cannot land without a v1-compatible manifest. Claude notes the design promises every decoded field is either v2-only, explicitly v1-compatible, or unsupported, but only the v2 registry has a file path. Without a checked-in `v1_compatible_fields.yaml` or equivalent seed rule, the reflection guard will fail or become subjective.
- **Major:** Future compiler requirement values need a host-capability axis. Claude points out that rows treating `>=3` as unsupported are only true for binaries whose known max is 2. The matrix needs `host_known_max` coverage, for example crossing known max `{2, 3}` with requirement values `{">=2", ">=3", ">=4"}`, so the accepted grammar can grow without changing historical meaning.
- **Minor:** Test-only parser constructors need an import guard. Claude flags that `internal/formula/testonly` can be imported by production packages unless it is moved under an internal-only path or protected by a static import test.
- **Minor:** Legacy JSON behavior remains undefined. Codex notes `.formula.json` is still parsed by the current parser, so JSON requirement rows must either cover omitted/valid/invalid/unknown/conflict behavior across durable caller paths or the JSON loader must be explicitly retired for this capability.

**Disagreements:**
- There is no verdict disagreement: both Claude and Codex return `approve-with-risks`.
- Claude focuses more on internal design contradictions and implementation gates: competing matrix schemas, `expand_vars` carrier semantics, BurntSushi/toml version pinning, v1-compatible field manifests, host-known capability growth, and test-only import safety. Codex focuses more on matrix execution risk: raw TOML spellings, condition/contribution cross-products, combined-defect precedence, source location, and JSON policy. Assessment: these are complementary findings; none conflict.
- Claude treats unsupported future minima as a release-versioned parser contract; Codex does not review that edge directly. Assessment: keep the distinct `formula.compiler_requirement_unsupported_future` diagnostic, but add the host-capability dimension so future releases can accept newly known capabilities without contradicting the matrix.

**Missing evidence:**
- No Kimi 2.6 artifact is present for this persona.
- No single canonical YAML schema currently reconciles the design's two matrix-row shapes, literal count locks, `coverage_intent_id`, and expected diagnostic projection.
- No checked fixture evidence pins BurntSushi/toml duplicate table, dotted/bracket collision, array-of-table, inline-table conflict, and parser-boundary behavior.
- No fixture set proves `condition_disabled` and materialized contribution behavior across root steps, children, loop bodies, expansion templates, map/aspect contributions, inline expansion, fragments, convergence, sling/API/order preflight, and source attribution paths.
- No `v1_compatible_fields.yaml` or equivalent manifest is named for the decoded-field classification guard.
- No matrix rows show host-known maximum capability changing from 2 to 3.
- No executable JSON policy is stated while `.formula.json` remains supported.

**Required changes:**
- Replace the competing matrix-row descriptions with one normative schema for `internal/formula/testdata/compiler_requirements_matrix.yaml` that includes `suite_id`, literal `count_lock`, `coverage_intent_id`, source/legacy/construct/caller/host dimensions, expected diagnostics, source path, line/column, fatality, and ordering.
- Add raw-shape fixtures for bracketed, dotted-key, inline-table, duplicate, collision, unknown-axis, invalid-type, malformed range, empty operand, overflow, array-of-table, and nested requirement declarations. Pin `github.com/BurntSushi/toml@v1.6.0` behavior or replace all parser-version-sensitive prose with fixture-locked rows.
- Resolve `expand_vars` semantics on compose entries, then update construct registry text, compose traversal rules, and fixtures for `compose.expand`/`compose.map` variables without contributed v2-only template content.
- Add misplaced-requirement rows for `[steps.requires]`, `[[steps.requires]]`, `[children.requires]`, `[loop.body.requires]`, `[compose.map.requires]`, `[compose.aspects.requires]`, `[compose.requires]`, and `[expand_vars.requires]`, each paired with a v2-only construct row proving the nested table does not satisfy the root requirement.
- Add a `condition_state` dimension with `authored_present`, `materialized`, and `condition_disabled`, and require construct-registry plus contribution-traversal suites to cover feasible construct/location/source/caller combinations.
- Add a combined-defect precedence suite that locks diagnostic accumulation or suppression, order, fatal/warning split, and parser-boundary versus semantic diagnostic precedence.
- Introduce a checked `v1_compatible_fields.yaml` or equivalent companion manifest before adding `TestEveryDecodedFormulaFieldClassified`.
- Add `host_known_max` rows for future compiler capability values so `>=N` behavior is explicitly binary-versioned.
- Move `internal/formula/testonly` under a production-unreachable path or add a static guard proving production packages do not import it.
- Decide the JSON formula policy in the parser PR: either add JSON requirement rows across durable caller paths or explicitly retire/document JSON support for compiler requirements.
