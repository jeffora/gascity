# Nadia Sorenson

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Info] The design keeps the core compiler boundary in the right place: formulas declare requirements, the active binary decides what it can satisfy, and `internal/formula` owns raw `contract`, `requires.formula_compiler`, `version`, and v2-only construct interpretation.
- [Info] The typed compile and acceptance flow is the right durable-write boundary. `NormalizedRequirements`, `HostCapabilities`, `CompileResult`, `AcceptedCompileArtifact`, `WorkflowRootFacts`, and accepted-artifact validation give callers projections instead of raw formula authority.
- [Info] The proposed static guard stack is concrete enough to make the boundary reviewable in CI: raw-consumer allowlists, durable-writer accepted-artifact checks, convergence-subset checks, and workflow-root predicate parity fixtures are the correct enforcement shape.
- [Major] The host-capability contract is still too loose. `HostCapabilitiesFromFormulaV2(enabled bool, source string)` cannot express normalized capability, source kind, config generation, omitted-default versus explicit-false provenance, or deprecated `graph_workflows` promotion without hidden defaults. The design also does not yet block future reads of legacy globals such as `formulaV2Enabled` or `IsFormulaV2Enabled` inside `internal/formula` satisfaction logic.
- [Major] The raw-consumer guard does not cover all load-bearing symbols. It omits legacy global capability accessors, future canonical root metadata keys such as `gc.formula_requirements_schema_version` and `gc.formula_min_reader_capability`, formula-source TOML decoding outside `internal/formula`, and possible `beads.ListQuery` filters over `gc.formula_*` metadata outside the shared workflow-root predicate.
- [Major] Test-only construction remains a production bypass unless fenced. A Go package named `internal/formula/testonly` is importable from production code unless it has a build tag or a static guard blocking non-test imports.
- [Major] `Recipe.GraphWorkflow` and the `Recipe` struct surface can become a second behavioral authority. The design relies on assignment guards, but it does not yet make clear whether graph workflow state is compiler-owned, immutable after compile, read only through accepted-artifact validation, or exposed as mutable recipe state.
- [Major] Future compiler axes can ship with weaker enforcement than v2. The new-axis checklist does not yet require updating the raw-consumer symbol list, `RequirementSource` misuse guard, workflow-root metadata key set, or `workflow_control_metadata.yaml` count locks.
- [Major] Packman schema 2 is a blocking prerequisite but is buried in persisted-metadata details. Imported-pack floor enforcement, external pinned-pack support expiration, and alias-removal gates depend on it, so rollout can stall silently if this is not a named coordination gate.
- [Minor] Several edge fixtures are implied but not specified, especially config generation changing between `CompileWithResult` and `AcceptCompileResult`, and a newer binary reading an older accepted artifact version.

**Disagreements:**
- Claude and Codex agree on the verdict: both returned `approve-with-risks`.
- Codex treats explicit per-operation `HostCapabilities` as a strength but flags the listed constructor as inconsistent with the data the design depends on. Claude treats the same model as sound only if fields, globals, and edge adapters are structurally guarded. Assessment: the direction is correct, but the design must require a typed constructor or edge adapter plus static guards before implementation.
- Codex rates the `Recipe.GraphWorkflow` risk as minor if the design clarifies how writers should use it. Claude treats the recipe surface as a broader authority leak. Assessment: this should be handled as a major boundary issue because durable writers can drift toward mutable recipe state unless the type or guards make accepted artifacts the only write authority.
- Claude raises packman schema sequencing, test-only imports, formula TOML ownership, construct-registry source of truth, and repair-root-artifact attribution. Codex does not address those topics. Assessment: absence from Codex is not disagreement; these remain valid compiler-boundary risks.

**Missing evidence:**
- No Kimi 2.6 review artifact was present for this persona.
- Concrete edge-builder API that converts city config into `HostCapabilities` while preserving normalized capability, diagnostic source, source kind, config generation, omitted default, explicit false, explicit true, deprecated `graph_workflows`, and test override provenance.
- Explicit statement that formula source TOML is decoded only inside `internal/formula`, with other packages consuming `CompileResult`, `AcceptedCompileArtifact`, or `WorkflowRootFacts`.
- Static evidence that production code outside `internal/sourceworkflow` cannot query or interpret `gc.formula_*` metadata directly.
- Exact `Recipe` struct surface and lifecycle: whether `Contract`, `Requires`, or `Version` fields exist or are forbidden, and whether `GraphWorkflow` is introduced, assigned, read, or mutated only through compiler-owned paths.
- Fixture coverage for a config generation change between compile and accept, and for a current binary that supports a newer artifact version than the artifact has.
- Proof that `workflow_control_metadata.yaml` is the sole source of truth for `min_compiler_capability`, or a generator/reflect test proving any Go-side constants cannot diverge.
- Explicit attribution rules for `gc formula repair-root-artifact` when current host capability and recovered legacy provenance differ.
- Operator story for intentionally freezing graph-specific writes on already accepted roots after disabling formula v2.

**Required changes:**
- Replace `HostCapabilitiesFromFormulaV2(enabled bool, source string)` with a typed constructor or edge adapter that requires normalized capability, diagnostic source, source kind, and config generation. Add tests for omitted default, explicit false, explicit true, deprecated `graph_workflows`, and test overrides, proving behavior branches only on normalized capability while diagnostics preserve provenance.
- Extend `TestNoNewFormulaRawConsumers` or add a sibling guard that scans `internal/formula/*.go`, excluding the named edge adapter and `_test.go`, for reads of `formulaV2Enabled`, `IsFormulaV2Enabled`, and `SetFormulaV2Enabled`. The edge adapter allowlist must name exact files and expiry phases.
- Add `//go:build testonly` to `internal/formula/testonly` or add `TestNoTestonlyImportsInProduction` that fails when any non-`_test.go` file imports test-only constructors, including accepted-artifact constructors.
- Add `gc.formula_requirements_schema_version`, `gc.formula_accepted_artifact_version`, `gc.formula_min_reader_capability`, `gc.formula_requirement_axes`, and `gc.formula_unsupported_axes` to the raw-consumer guard. Only `internal/formula` and the precise `internal/sourceworkflow` owner symbols should be allowed to read them.
- State that formula source TOML is decoded only inside `internal/formula`, then add guard coverage so dashboard, convoy scan, order dispatch, sling, API, and other projections cannot parse formula TOML to recover behavior.
- Extend the new compiler capability or requirement-axis checklist to require raw-consumer symbol updates, `RequirementSource` misuse-guard updates, workflow-root predicate metadata key updates, and `internal/formula/testdata/workflow_control_metadata.yaml` count-lock updates.
- Make `Recipe.GraphWorkflow` structurally compiler-owned, for example by unexporting the field and exposing only a read accessor or by requiring all durable writers to prove accepted-artifact validation before using graph state. Add a legacy helper retirement row or guard for the `Recipe` struct surface itself, covering `Contract`, `Requires`, and `Version` fields if they exist or are reintroduced.
- Promote packman schema 2 to a named rollout-plan coordination gate with an owner and explicit blocking relationship to imported-pack floor enforcement, external pinned-pack support expiration, and alias removal.
- Add durable preflight fixtures for config generation changing between `CompileWithResult` and `AcceptCompileResult`, requiring fail-closed behavior with zero protected writes and recompile on the current host snapshot.
- Add artifact compatibility fixtures for a newer binary reading an older accepted artifact, alongside the unsupported-newer-artifact and minimum-reader-capability cases.
- State whether `workflow_control_metadata.yaml` is the sole source of truth for `min_compiler_capability`. If Go-side constants remain, require generated code or reflect tests proving the YAML and Go constants cannot diverge.
