# Elias Vega

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- **[Major] External pinned-pack compatibility is not yet an enforceable contract.** Both reviews identify the "documented compatibility branch" as load-bearing but undefined. The design must state whether this is a long-lived binary branch, shim, frozen tag, or migration policy; who owns it; what it backports; how long it is supported; and how release approval proves SHA-pinned legacy formulas remain supported after alias removal.
- **[Major] `gc formula validate --all-packs --legacy-contract-report` is underspecified for a release gate.** The reviews agree that criterion 2 depends on this command, but the design does not define the full scan scope, first-party versus external classification, handling of imported packs, aspects, expansions, convergence subsets, dirty/local packs, intentional fixtures, report schema, exit codes, or whether the command gates CI or only a release checklist.
- **[Major] `contract` and `[requires]` interaction needs a canonical truth table.** The current matrix leaves ambiguous cells such as `contract = "graph.v2"` plus empty `[requires]`, unrelated keys only, explicit `>=1`, unsupported `>=3`, whitespace/case variants, over-declaration, and inheritance/lowering through parent-child formula composition. Different implementations could still silently diverge on conflict handling.
- **[Major] Invalid `contract` values lack a dedicated diagnostic contract.** Claude notes that `contract = "graph.v3"`, empty strings, or other invalid literals do not fit cleanly under the `[requires]`-oriented `formula.compiler_requirement_unsupported` code. Without a field-specific code and canonical remediation, CLI, API, dashboard, controller, and event projections can improvise inconsistent errors.
- **[Major] Deprecation warning reach is incomplete.** Both reviewers flag that users may not see `formula.contract_deprecated` before alias removal. Long-running cities may not recompile stable workflows, daemon/order paths have ambiguous event behavior, and `GC_NATIVE_FORMULA=false` or direct `bd cook --persist` consumers may bypass Gas City's new diagnostics entirely.
- **[Major] Dual-stamping lifecycle is undefined.** Workflow roots are required to carry `gc.formula_contract = "graph.v2"` during the alias window, but the design does not state whether this stops with source-alias removal, has separate readiness criteria, or depends on all root-consuming projections adopting the shared predicate.
- **[Minor] Warning deduplication and volume controls are not reviewable.** CLI `OnceKey` scope is described, but controller, order dispatch, convergence, API, dashboard, and HA behavior are not pinned. That leaves warning reach and event-bus volume dependent on implementation choices.
- **[Minor] Several migration policy details need to be made explicit.** The two-minor-release criterion should be framed as integrator soak time rather than usage evidence; over-declaration with `requires.formula_compiler = ">=2"` needs an allowed/disallowed answer; old-binary compatibility for dual-declared formulas needs a regression test or binary floor; and convergence should use one canonical compiler/preflight strategy instead of an "or" path.

**Disagreements:**
- **Persona verdict:** There is no verdict disagreement. Claude and Codex both recommend `approve-with-risks`; this synthesis adopts that verdict because the design has the right migration shape but several contract details must be specified before implementation.
- **External compatibility emphasis:** Claude frames the missing branch definition as the central problem. Codex additionally wants executable external counts and an explicit post-removal behavior for external legacy-only formulas. Assessment: require both a written compatibility policy and a machine-checkable release artifact; the branch phrase alone is not sufficient.
- **Warning reach scope:** Claude emphasizes long-running cities, order/daemon paths, and external `bd` consumers. Codex focuses on the `GC_NATIVE_FORMULA=false` or bd shell-out path and asks whether Gas City should run native validation before delegating. Assessment: the design should state which fallback paths project diagnostics and declare alias removal blocked until uninstrumented fallback paths are gone or mitigated.
- **Edge-case breadth:** Claude lists a broader `contract` x `[requires]` matrix and diagnostic taxonomy than Codex. Codex agrees on the conflict-prevention goal but calls out fewer cells. Assessment: include the broader matrix because this is the public compatibility contract that prevents projection drift.
- **Dual-stamping sunset:** Claude raises runtime `gc.formula_contract` sunset; Codex does not. Assessment: keep it as required evidence because source migration and workflow-root metadata migration can have different consumers and timelines.

**Missing evidence:**
- A precise definition of the documented compatibility branch or equivalent LTS policy, including owner, lifetime, backport scope, release artifact location, consumer migration step, and retirement criteria.
- The full contract for `gc formula validate --all-packs --legacy-contract-report`: source coverage, first-party/external classification, imported pack and pinned SHA handling, aspect/expansion/convergence coverage, fixture opt-outs, JSON schema, exit codes, and CI versus release-checklist use.
- A complete `contract` x `[requires]` compatibility table covering absent, empty/default, unrelated-key-only, `>=1`, `>=2`, unsupported expressions, invalid `contract` values, whitespace/case variants, parent-child inheritance, lowered child requirements, over-declaration, and host-capability failures.
- A dedicated diagnostic code and canonical message/remediation for invalid `contract` values.
- A warning-reach plan for stable long-running cities, daemon/order execution, controller event projection, `GC_NATIVE_FORMULA=false`, bd shell-out, and direct external `bd cook --persist` users.
- The binary floor for dual-declared formulas and evidence that the oldest supported `bd`/`gc` readers ignore or accept `[requires]`.
- The sunset rule for `gc.formula_contract` dual-stamping and the shared workflow-root predicate's legacy fallback.
- Projection-specific `OnceKey` deduplication scopes and HA semantics.
- A stated policy for over-declaration and one selected convergence validation strategy.

**Required changes:**
- Define the external compatibility contract before alias removal: branch or LTS form, owner, support duration, backport policy, release artifact path, migration instructions, and retirement criteria. Add a post-removal matrix row for "new binary plus external legacy-only formula" that states whether it is supported by mainline alias, compatibility branch, or intentional hard failure with remediation.
- Specify `gc formula validate --all-packs --legacy-contract-report` end to end, including scan scope, external/imported aggregate counts such as `external_legacy_only` and `external_dual_declared`, first-party zero criteria, exit-code semantics, and an explicit alias-removal decision field or equivalent release-gate output.
- Expand the compatibility section with a canonical truth table for all `contract` and `[requires]` combinations noted above, including deterministic diagnostic ordering for conflicts, unsupported expressions, invalid `contract` values, and host-capability failures.
- Add a field-specific diagnostic such as `formula.contract_invalid_value` with canonical message, remediation, and projection requirements.
- Add a warning-reach subsection. It should define periodic revalidation guidance for long-running cities, projection behavior for daemon/order/controller paths, and whether fallback execution through `GC_NATIVE_FORMULA=false` or bd shell-out runs native validation for diagnostics before delegating. If not, alias removal should be blocked until those fallback paths are gone.
- Define the `bd` compatibility story: whether `bd cook` and `bd mol wisp` must surface deprecation warnings, what `bd` version floor is part of the alias window, and how direct external `bd` consumers receive migration guidance.
- Define the sunset for `gc.formula_contract` dual-stamping and agreeing dual declarations, either tied to source-alias removal or to separate readiness criteria for every workflow-root consumer.
- Pin `OnceKey` deduplication scope per projection and state HA behavior.
- Clarify criterion 1 as soak time only, permit or reject over-declaration explicitly, and choose one convergence validation path that reuses the canonical compiler result.
- Add regression coverage for dual-declared formulas on the oldest supported `bd`/`gc` binary, plus tests for warning projection on fallback paths and external SHA-pinned legacy-only formulas during and after the alias window.
