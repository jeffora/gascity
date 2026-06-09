# Priya Zielinski - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The v0 grammar is now fail-closed and byte-exact enough to reject `2`, `>= 2`, whitespace variants, future-looking ranges, empty strings, and non-string TOML values.
- The design explicitly requires raw TOML key/type/source-position preservation before final typed decoding, which closes the main parser blind spot for `[requires]`.
- The design centralizes diagnostics and host capability satisfaction in `internal/formula`, with enough caller inventory to prevent CLI/API/order/convergence drift.

**Critical risks:**
- [Major] The v2-only construct registry is still not mechanically encodable. The registry repeats "Graph workflow metadata keys", has one row that names scanned locations but no exact keys, one row that names `gc.kind` values but not the containing TOML path, and one row that names keys but not the scanned locations. An implementer cannot generate exhaustive positive/negative tests from this table without making new judgment calls.
- [Major] The required validation matrix is still compressed enough to miss combinations. Rows such as "omitted or empty table", "no or yes", "wrong TOML type, dotted table, or nested table", and "any" are useful summary prose, but they are not a normative matrix across expression, TOML shape, legacy `contract`, v2-only construct location, inherited or aggregated requirements, and host capability.
- [Major] Diagnostic precedence is only partially testable. Unknown axes and invalid types share one ordering tier, and warning diagnostics such as `formula.contract_deprecated` and `formula.version_deprecated` are not placed relative to fatal diagnostics for validate/show paths. Multi-defect files can still produce unstable output unless the design defines tie-breaking by source position, code, or both.
- [Minor] Empty/default provenance is under-specified. The design says omitted `[requires]` and empty `[requires]` both mean default capability, but it does not say whether an explicit empty table is recorded as `RequirementSourceOmitted`, `RequirementSourceRequires`, or another source shape for provenance, diagnostics, and `OnceKey` behavior.

**Missing evidence:**
- An encodable v2-only construct table with columns for construct id, exact TOML paths scanned, trigger predicate, positive fixture, and negative fixture.
- A separate TOML-shape matrix for absent table, empty table, inline table, dotted table, nested table, duplicate declaration, unknown sibling key, multiple unknown keys, invalid value type, unsupported string, and accepted string.
- A source-location matrix proving the same v2-only construct rule for root steps, children, `loop.body`, expansion templates, aspect formulas, and aggregated expansion/aspect output.
- Combined-defect precedence cases, including `contract = "graph.v2"` plus unsupported `[requires]`, invalid type plus unknown axis, missing v2 requirement plus disabled host capability, and `contract` conflict plus disabled host capability.
- Expected normalized provenance for omitted requirements, empty `[requires]`, explicit `>=1`, explicit `>=2`, inherited parent requirements, expansion/aspect maximum requirements, and attempted lowering.

**Required changes:**
- Rewrite the v2-only construct registry as a test-generating table: `construct_id`, `toml_path`, `predicate`, `requires_graph_v2`, `positive_fixture`, and `negative_fixture`. Split metadata key presence from metadata value matching so `gc.kind` values and `gc.scope_*` keys are unambiguous.
- Expand the validation matrix into normative sub-matrices instead of grouped summary rows: requirement expression/TOML shape, legacy `contract` compatibility, v2-only construct location, inheritance/aggregation, host capability, and combined-defect precedence.
- Define deterministic diagnostic sorting inside each precedence tier. At minimum, specify source position first, then diagnostic code, and state whether validation returns all diagnostics or first fatal diagnostic for each caller class.
- Place `formula.contract_deprecated` and `formula.version_deprecated` in the diagnostic ordering model for validate/show surfaces, including whether warnings are returned when fatal requirement diagnostics also exist.
- Add explicit matrix rows for `contract = "graph.v2"` combined with invalid or unsupported `[requires]`, multiple unknown `[requires]` axes, invalid type plus unknown axis, missing requirement plus host disabled, expansion/aspect requirement raising, and attempted lowering below a parent requirement.
- Define the normalized source/provenance behavior for an explicit empty `[requires]` table and explicit `formula_compiler = ">=1"` so API projections, metadata, and deprecation/warning suppression cannot drift.

**Questions:**
- Should an explicit empty `[requires]` table be observable in provenance as a user-authored requirement source, or normalized exactly like omission?
- For validate/show paths, should warning diagnostics be emitted alongside fatal parser diagnostics, or only after fatal requirement validation succeeds?
- Is `gc.kind = "ralph"` a historical metadata value that must stay in the v2-only registry, and can the registry describe it purely as a metadata value rather than a construct name?
