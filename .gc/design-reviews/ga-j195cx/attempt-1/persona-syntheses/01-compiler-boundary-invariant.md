# Nadia Sorenson

**Persona verdict:** block

**Sources:** Claude, Codex (Kimi 2.6 absent)

**Consensus findings:**
- [Info] Both reviews agree the core invariant is directionally correct: formulas declare minimum requirements, and the active Gas City binary decides whether its host capabilities satisfy them.
- [Info] Both reviews agree `internal/formula` is the right ownership boundary for requirement parsing, v2-only construct detection, compatibility normalization, diagnostics, and typed compile outputs.
- [Major] The active-binary decision point is still under-specified. Codex requires a canonical compile/preflight API that accepts the resolved formula source plus active host capabilities; Claude requires structural guarantees that invalid host capabilities and fatal diagnostics cannot reach durable writers.
- [Major] Legacy compile and recipe surfaces remain bypass paths. Both reviews flag `Compile(...)`, bare `*Recipe`, and especially `Recipe.GraphWorkflow` as ways for callers to infer graph acceptance without normalized diagnostics, provenance, host-capability proof, or accepted-artifact validation.
- [Major] Static guard coverage is incomplete. Existing guard concepts cover raw contract reads, but the synthesis of both reviews requires blocking direct production reads of host capability shims such as `formula_v2`, `FormulaV2Enabled`, deprecated metadata aliases, and behavioral `Recipe.GraphWorkflow` reads outside the approved compile/preflight or metadata-predicate path.
- [Major] Persisted workflow-root classification needs one named, tested authority. The design must pin package, function, input type, canonical-first behavior, legacy fallback, and conflict semantics so callers do not duplicate canonical/legacy metadata interpretation.
- [Major] Convergence transition is a boundary-risk hotspot. Claude requires a drift-detection fence comparing shadow `CompileWithResult` and legacy subset-parser decisions; Codex independently flags convergence as one of the high-risk projections that must consume normalized results rather than raw metadata.
- [Minor] The v2-only construct registry needs a tighter schema. Codex calls out ambiguous repeated entries and missing exact AST/model field locations; Claude's proposed fixture rows per construct registry entry align with making the registry compiler-enforceable.
- [Minor] Deprecated host-capability spellings are tightly scoped but not fully sunset. Claude specifically flags `graph_workflows` as lacking a scheduled removal phase, and both reviews push toward a single typed host-capability source.

**Disagreements:**
- Claude verdict was `approve-with-risks`; Codex verdict was `block`. Assessment: choose `block` because the missing canonical compile/preflight API leaves the central boundary implementable in multiple conflicting ways.
- Claude presents active-binary ownership as mostly anchored by named tests, while Codex says the decision point is not yet expressed as one API. Assessment: Claude's tests are necessary but insufficient until the API that carries host capabilities into `internal/formula` is explicit.
- Claude treats `Recipe.GraphWorkflow` read authority as a minor risk; Codex treats legacy `Compile(...)` and bare `*Recipe` as a major bypass class. Assessment: classify as major because durable or behavioral callers can bypass diagnostics and host-capability evidence without assigning to the field.
- Claude asks for convergence drift detection while Codex emphasizes normalized caller consumption. Assessment: the drift fence is the concrete rollout mechanism needed to make Codex's broader boundary rule true during sub-phases 4b-4f.

**Missing evidence:**
- No Kimi 2.6 artifact was present.
- No final compile/preflight signature shows `ResolvedFormulaSource`, active host capabilities, diagnostics, compiled recipe/projection, graph flag, runtime vars if any, and provenance flowing through one result.
- No explicit invariant proves host capabilities returned with fatal constructor diagnostics can satisfy no requirement and can reach no durable writer.
- No package, function name, input type, or conflict behavior is pinned for the shared workflow-root metadata predicate.
- No property-style parity test proves `internal/formula.RootMetadataFacts` and `internal/sourceworkflow.ClassifyWorkflowRoot` agree across canonical, dual-stamped, legacy, malformed, future-capability, and unknown-axis metadata.
- No static-guard evidence proves direct host-flag decisions, legacy `Compile(...)` behavioral use, `Recipe.GraphWorkflow` behavioral reads, raw metadata reads, and deprecated alias reads fail outside allowlisted adapters/tests.
- No scheduled removal phase is shown for the deprecated `graph_workflows` host-capability source.
- No guard or timed allowlist expiry is shown for the temporary `cmd/gc/feature_flags.go:applyFeatureFlags` compatibility shim after caller sub-phase 4a.
- No convergence fixture matrix proves every construct registry entry is shadow-compiled and that native-vs-legacy accept/reject drift fails closed before writes.
- No precise schema names the exact v2-only AST/model fields, metadata map locations, and walker contract.

**Required changes:**
- Define the canonical compile/preflight API in `internal/formula`, including how the active binary's host capability enters without callers interpreting requirements and without `internal/formula` importing upper-layer config concerns.
- Make legacy `Compile(...)`, bare `*Recipe`, and `Recipe.GraphWorkflow` unsuitable for durable or behavioral graph decisions, either by structural access changes or static guards that require `AcceptedCompileArtifact`/`CompileResult` authority at the correct boundary.
- Expand static guards to cover raw requirement metadata, raw `contract` fields, v2 construct strings, direct production reads of host-capability flags, deprecated aliases, and behavioral graph-state reads outside the approved adapter/predicate/test allowlist.
- Name the shared workflow-root predicate and specify package, function name, input type, canonical-first behavior, legacy fallback, and fail-closed/reporting behavior when canonical and legacy metadata disagree.
- Add a convergence transition fence: during sub-phases 4b-4f, compare shadow `CompileWithResult` and legacy subset-parser accept/reject decisions on every selected formula; any disagreement fails closed with a typed transition-fence diagnostic and zero durable writes.
- Add construct-registry fixtures proving every exact v2-only AST/model field and metadata key produces the same normalized requirement and diagnostic on every caller path.
- Pin a removal phase for `HostCapabilitySourceDeprecatedGraphWorkflows` and add a timed guard or allowlist expiry for `applyFeatureFlags` after caller sub-phase 4a.
- Add `TestWorkflowRootFactsAndMetadataPredicateAgree` or equivalent property coverage across canonical, dual, legacy, future-capability, malformed, and unknown-axis metadata maps.
- Add a structural host-capability invariant: any constructor result with fatal `formula.host_capability_invalid` or other error-severity diagnostics is rejected by `CheckRequirements`/`ValidateAcceptedArtifact` before any writer boundary.
