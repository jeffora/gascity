# Ibrahim Park

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] Default-capability identity is central to future compatibility and needs executable contracts at the same strength as the prose. Both reviews agree omitted `[requires]`, empty `[requires]`, and `formula_compiler = ">=1"` must normalize to capability `1`, share runtime behavior, and differ only in diagnostic/provenance display; Claude finds the named test/static-guard coverage missing, while Codex treats the rule as explicit but still depends on it holding across future API evolution.
- [Major] The second-axis path is directionally sound but under-specified at the API/provenance boundary. Claude asks for a worked non-monotonic axis and an axis-namespace policy; Codex flags that the canonical v0 `NormalizedRequirements` shape still exposes singular `source`, `sourcePath`, `Source()`, and `SourcePath()` fields that could become the wrong durable contract once per-axis provenance is needed.
- [Minor] Unknown axes and unsupported future minima fail closed with deterministic diagnostics, but the future-axis checklist must become an enforceable release gate. The design should prevent parser acceptance from shipping before docs, host-capability accessors, old-reader fixtures, persisted metadata decisions, diagnostic projections, and generated matrix rows are complete.
- [Minor] The narrow `formula_compiler` grammar is a future-compatibility feature, but the permanence of that grammar is easy to miss. The Forward Compatibility section freezes released axis byte grammars; the Requirement Expression section should say directly that richer syntax requires a new axis or schema-versioned extension, not a later relaxation of `formula_compiler`.
- [Minor] Documentation and fixture evidence is not yet reviewable in several places. Both reviews note placeholder/generated count locks or absent implementation artifacts; Codex also calls out stale public docs that still teach legacy `version` rather than `[requires]`.

**Disagreements:**
- Claude treats the missing omitted/empty/`>=1` runtime-equivalence test and `RequirementSource` branch guard as major required work; Codex does not elevate that as a top risk because the design's prose is clear. Assessment: carry Claude's requirement forward because the persona's mandate is future-capability drift, and prose-only equivalence is exactly where drift can enter.
- Codex's primary major risk is the singular v0 `NormalizedRequirements` provenance/API shape; Claude does not mention it. Assessment: this is compatible with Claude's per-axis provenance concerns and should be added as a required clarification rather than ignored.
- Claude asks for a concrete non-monotonic future-axis example and namespace policy; Codex accepts the v0 rejection of user-defined namespaces but wants the future-axis checklist promoted to a gate. Assessment: the minimal durable outcome is to both define who may own axis names and make any future-axis addition follow an enforceable typed-axis template.
- Codex frames the stale docs issue as non-blocking because the design already has a hard docs gate; Claude asks for a specific grammar-permanence sentence in the reader-facing Requirement Expression section. Assessment: both point to doc clarity as required before release, but neither makes it a design block.

**Missing evidence:**
- Kimi 2.6 review was not present in the attempt's `reviews/` directory.
- No named executable test or test family is identified for runtime equivalence across omitted `[requires]`, empty `[requires]`, and explicit `formula_compiler = ">=1"` over CompileID, accepted-artifact hash/dedup, workflow-root kind, convergence/fanout/retry validation, dispatch, and dashboard/API projection.
- No named static guard is identified for banning runtime branches on `RequirementSource`, boolean "is v2" helpers, or capability-literal comparisons outside the compiler boundary.
- Literal generated row/count locks for grammar, raw-shape, future-boundary, and caller-preflight suites are not visible in the raw reviews; the design still appears to contain placeholder count expressions in some snippets.
- No end-to-end second-axis fixture is shown for source parsing, unknown-axis old-reader behavior, normalized typed state, host capability, accepted artifact metadata, API/dashboard projection, release checklist output, and persisted metadata with both a known axis and an unknown future axis.
- No explicit axis-identifier namespace policy or reserved-future axis list is present; it is unclear whether all axes are forever Gas-City-owned or whether third-party names will ever be permitted.

**Required changes:**
- Add named executable contracts for default-capability equivalence. Reference a `TestDefaultCapabilityRuntimeEquivalence`-style test or test family next to the prose claim, and cover CompileID, accepted-artifact hash/dedup reference, workflow-root kind, convergence/fanout/retry validation, dispatch, and dashboard/API projection for omitted `[requires]`, empty `[requires]`, and explicit `formula_compiler = ">=1"`.
- Add or name static guards for future-drift hazards: runtime branching on `RequirementSource` outside diagnostic/provenance/migration/compat paths, boolean "is v2" helpers, and capability-literal comparisons outside `internal/formula`.
- Clarify that the singular `NormalizedRequirements` `source`/`sourcePath` fields and `Source()`/`SourcePath()` accessors are v0-only shorthand for the single `formula_compiler` axis. Adding any second axis must introduce typed per-axis requirement/provenance accessors and must not turn the structure into a generic raw map or reuse the singular accessors as generic authority.
- Make the future-axis checklist an explicit CI or release-gate artifact owned with the parser matrix, so parser acceptance for any future axis cannot land without docs, old-reader fixtures, persisted metadata schema decisions, generated projections, diagnostics, and release registry updates.
- Add one worked non-monotonic future-axis example covering grammar, owning package, host-capability accessor shape, diagnostic codes, persisted metadata key, projection behavior, and provenance fields.
- Add an explicit axis namespace policy. Either state that all requirement axes are forever Gas-City-owned, or define the reserved third-party naming shape and v0 rejection behavior.
- Add a sentence in the Requirement Expression section that the accepted/rejected byte grammar for `formula_compiler` is permanent; richer expressivity requires a new typed axis or schema-versioned extension, never a relaxed `formula_compiler` parser.
