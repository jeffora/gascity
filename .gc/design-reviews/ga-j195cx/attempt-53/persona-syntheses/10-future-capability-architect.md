# Ibrahim Park

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] The v0 `[requires]` direction is sound, but the forward-compatibility invariants need to be stated as invariants, not implied behavior. Capability numbers must be monotonic and additive only: no released construct removal, no semantic redefinition, and no value reuse.
- [Major] Released requirement-axis grammar must be frozen. Codex requires the syntax-vs-support split so `>=3` is an unsupported-future capability instead of malformed authoring; Claude agrees that later ranges, minor versions, caret/tilde operators, or bounded expressions must not mutate the released `formula_compiler` grammar and should become new typed axes instead.
- [Major] Old-reader behavior for persisted future workflow roots is central and still needs stronger cross-product evidence. Both reviews require observation-only classification for future compiler capability or future axes, with graph-specific writes refused under a stable `unknown_future_capability_workflow` style classification.
- [Minor] Schema-version coupling is underspecified. The design should say explicitly that capability bumps inside the existing `formula_compiler` axis do not bump `gc.formula_requirements_schema_version`, while adding a persisted-behavior axis does.
- [Minor] The construct capability registry is still shaped around one axis. Future axes such as `state_store`, `tools`, or `runtime` need either a multi-axis capability map or a documented per-construct/per-axis row model before the first second-axis PR.
- [Minor] The compatibility corpus needs fixture parity for default-capability inputs and multi-axis future roots. Omitted `[requires]`, empty `[requires]`, and explicit `formula_compiler = ">=1"` should be proven behaviorally identical except for provenance, especially against every v2-only construct.

**Disagreements:**
- Both reviewers return `approve-with-risks`, so there is no verdict disagreement.
- Codex says unknown `[requires]` axes, default canonicalization, syntax-vs-support diagnostics, and persisted future-root metadata are missing. Claude's review treats those same areas as largely present in the current design: closed axes, unsupported-future diagnostics, normalized requirement values, and old-reader metadata are listed as strengths. Assessment: this looks like review-time skew or different design snapshots. Do not block on the older missing-feature claims if the Claude-reviewed text is authoritative, but keep fixture and invariant requirements where both reviews still expose risk.
- Codex leaves room for future expression growth inside `formula_compiler`, such as minor versions or ranges, if old readers classify unsupported values correctly. Claude argues this would make old-reader diagnostics unreliable and should be forbidden. Assessment: choose Claude's stricter rule. Freeze released axis grammar and model richer future syntax as a new axis.
- Codex asks for a future-axis example such as `tools`, `runtime`, or `pack_format`; Claude says the updated design uses `state_store` as the worked second-axis example and bans user-defined `x-*` passthrough. Assessment: `state_store` is enough as an example only if the design also generalizes registry shape, schema-version rules, and per-axis metadata ownership.
- Claude questions whether explicit `formula_compiler = ">=1"` should exist at all; Codex requires omitted, empty, and explicit default inputs to canonicalize. Assessment: do not require deprecating `>=1` for this design review. Require normalized behavioral parity and diagnostic-only provenance.

**Missing evidence:**
- No Gemini review artifact is present for this persona; Gemini is optional for this step.
- No explicit additive-only capability invariant covering construct removal, construct renaming, semantic changes to released constructs, and capability value reuse.
- No explicit frozen-grammar rule for released axes saying future ranges, minor versions, or alternate operators require a new axis rather than an expanded `formula_compiler` grammar.
- No exact rule for whether same-axis capability bumps affect `gc.formula_requirements_schema_version`, contrasted with the rule for adding a new persisted-behavior axis.
- No generalized `ConstructCapability` representation for multi-axis requirements, such as `MinCapabilityByAxis`, or an equivalent per-construct/per-axis registry model.
- No construct-registry fixture matrix pairing every v2-only construct with omitted `[requires]`, empty `[requires]`, and explicit `formula_compiler = ">=1"` and asserting identical normalized behavior plus the expected missing-requirement diagnostic.
- No cross-product old-reader fixture for a root stamped with both future compiler capability and a future axis manifest, proving observation remains possible and graph-specific writes refuse consistently.
- No explicit migration policy for a future need to retire a released construct; the reviews imply the answer should be "never remove, only deprecate via diagnostics," but the design does not state it.
- No final scaling policy for future-axis metadata namespaces: repeated `gc.formula_<axis>_*` keys versus a consolidated accepted-artifact payload.

**Required changes:**
- Add a Forward Compatibility "Invariants" subsection that states: compiler capability evolution is additive only; released constructs are not removed or semantically redefined; released requirement-axis byte grammars are frozen; richer future grammar is modeled as a new typed axis; and `RequirementSource` is diagnostic/provenance-only.
- State explicitly that `formula_compiler` capability bumps do not bump `gc.formula_requirements_schema_version`, while adding any new axis that affects persisted behavior does.
- Generalize `ConstructCapability` for multi-axis constraints, either with a `MinCapabilityByAxis` shape or a documented one-row-per-construct-per-axis model.
- Add construct-registry fixtures proving omitted `[requires]`, empty `[requires]`, and explicit `formula_compiler = ">=1"` are behaviorally identical for every v2-only construct and emit identical normalized results and `formula.compiler_requirement_missing` diagnostics.
- Add an old-reader cross-product fixture for a root containing both future compiler capability and a future axis, asserting `unknown_future_capability_workflow` classification, observation-only behavior, graph-write refusal, and unsupported-axis dashboard/API projection.
- Clarify the future construct retirement policy. If removal is forbidden, say so and direct deprecation to diagnostics only.
- Clarify the future-axis metadata scaling strategy so the first second-axis implementation does not need to invent whether each axis gets dedicated `gc.formula_<axis>_*` keys or whether accepted artifacts carry a consolidated requirements payload.
