# Nadia Sorenson

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Info] Both reviews agree the design has the right top-level compiler boundary: formulas declare requirements, the active host decides whether it can satisfy them, and callers should consume typed compiler outputs instead of raw `contract`, `requires`, or workflow-control metadata strings.
- [Major] The boundary is not yet structurally enforced everywhere inside `internal/formula`. Legacy raw-decision helpers such as `declaresGraphV2Contract`, `requiresExplicitGraphContract`, `metadataRequiresGraphContract`, and the global `formulaV2Enabled` path remain possible parallel authorities unless the design pins deletion and guard tests.
- [Major] `Recipe.GraphWorkflow` is load-bearing compiler truth but is only described as a read-only projection. If it remains an exported mutable field, external composite literals can synthesize graph-workflow state outside the compiler; it needs an unexported compiler-owned representation or equivalent factory boundary.
- [Major] Workflow-control metadata still has multiple semantic readers unless all `gc.kind` classifiers and graph-routing/topology/control helpers are explicitly migrated to generated registry-backed facts or typed compiler projections.
- [Major] Convergence migration needs blocking sequencing. A "retire or rewrite" option for `internal/convergence/formula.Formula` and `ValidateForConvergence` leaves room for a compatibility path that keeps the legacy subset parser alive after the new compiler contract is supposed to own the decision.
- [Major] Accepted compile artifacts are treated as authorization for future graph-specific writes, but the persistence owner and transaction model are not fixed. The design should choose the canonical substrate and specify ordering, recovery, stale-reference behavior, and portability.
- [Major] Legacy workflow-root metadata compatibility is underspecified. In particular, `gc.kind=workflow`-only roots need an explicit classification, migration diagnostic, parity tests, and expiry rule so old live roots are neither hidden nor treated as unbounded author truth.
- [Minor] `requires.formula_compiler = ">=2"` carries graph-workflow implications close to a selector unless the design explicitly states the invariant and guards that production code branches only on normalized requirement plus host capability, never the raw source spelling.

**Disagreements:**
- There is no verdict disagreement; Claude and Codex both return `approve-with-risks`.
- Claude focuses on in-package compiler-boundary escape hatches: legacy helper deletion, `Recipe.GraphWorkflow` mutability, `formulaV2Enabled` global reads, convergence guard timing, parser cache shape, and release-window clarity.
- Codex focuses on projection and persistence drift: registry-backed workflow-control readers, `gc.kind=workflow` legacy roots, accepted-artifact storage, and the "not a selector" semantics of `requires.formula_compiler`.
- Assessment: these are complementary findings. Claude's internal-boundary risks and Codex's projection/persistence risks should both be treated as required design hardening before implementation, because either class can recreate duplicated compiler authority.

**Missing evidence:**
- Gemini review was not present; this synthesis uses the required Claude and Codex inputs only.
- No committed caller manifest seed with the current raw metadata readers, including at least `internal/sling/sling.go`, `internal/sling/sling_attachment.go`, and `internal/graphroute/graphroute.go`.
- No full in-package replacement map for legacy graph-contract helpers and global capability reads inside `internal/formula`.
- No concrete generated-registry consumer API showing how graph route, topology, control-dispatcher, blocker, and binding code asks for workflow-control facts without duplicating `gc.kind` literals.
- No migration fixture for `gc.kind=workflow`-only roots across sourceworkflow, sling attachment, convoy/source workflow scans, API projections, and cleanup behavior.
- No accepted-artifact storage contract naming whether artifacts are beads, files, or another typed record, and how artifact persistence is ordered with root metadata stamping and child writes.
- No parser cache key statement proving host capability checks run after parse and do not leak through formula-name-only cache reuse.
- No reconciliation of the "two completed Gas City minor releases" alias-removal rule with current `0.x.y` versioning.
- No diagnostic-ordering rule for external packs whose `[pack] requires_gc` floor is lower than the floor implied by their own formula compiler requirements.

**Required changes:**
- Pin Phase 8 to a concrete deletion list for legacy in-package raw-decision symbols: `Formula.Contract`, `declaresGraphV2Contract`, `requiresExplicitGraphContract`, `stepsRequireGraphContract`, `stepsRequireDetachedGraphContract`, `stepRequiresGraphContract`, and `metadataRequiresGraphContract`.
- Make `Recipe.GraphWorkflow` compiler-owned state, for example an unexported field with `Recipe.IsGraphWorkflow()` or an equivalent internal factory, and add a blocking test that callers cannot synthesize it.
- Add an in-package guard that prevents `internal/formula` from reading `formulaV2Enabled.Load()` or `IsFormulaV2Enabled()` outside the dedicated edge adapter.
- Move the convergence legacy-subset-parser guard to blocking before convergence migration begins, not report-only after the fact.
- Add `internal/graphroute` and current `internal/formula` `gc.kind` classifiers to the caller manifest with target APIs that consume the generated workflow-control registry or typed compiler facts.
- Extend the workflow-root predicate table with `gc.kind=workflow`-only roots, including whether they are observable-only, repairable/mutable, or deprecated invalid metadata, and add parity tests against current `sourceworkflow` behavior.
- Choose and document the accepted-artifact persistence substrate and transaction model before durable writers are migrated.
- Add a "not a selector" guard test proving production behavior uses `NormalizedRequirements.FormulaCompiler()` plus host capability only, not `RequirementSource`, raw TOML, or persisted metadata strings except for diagnostics/provenance/shared predicates.
- Add a parser-cache invariant stating the cache key shape and confirming that host-capability checks run after parse.
