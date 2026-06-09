# Design Review Synthesis

## Overall Verdict: block

The design is directionally sound: reviewers broadly agree that `[requires]` is the right user-facing concept, that formulas must declare host requirements rather than select compiler implementations, and that pack revision should remain the durable artifact boundary. The review blocks because the parser and validation contract is still not normative enough to implement safely; without an exhaustive machine-readable matrix, deterministic diagnostics, and explicit provenance/inheritance semantics, callers and tests can still diverge on the central behavior.

## Consensus Strengths
- `[requires] formula_compiler = ">=2"` is a clearer capability declaration than `contract = "graph.v2"` and preserves the boundary where the active Gas City binary decides how to compile.
- The design correctly avoids formula-level artifact semver and keeps reproducibility anchored in pack version, ref, or commit SHA.
- Reviewers praised the fail-closed intent for requirement parsing: only omitted, `>=1`, and `>=2` should be accepted in v0.
- The shared diagnostic shape and canonical compiler result are the right foundations for CLI, API, dashboard, event, order, convergence, and molecule projections.
- The migration strategy has the right broad ingredients: dual declaration during the alias window, old-reader compatibility checks, static guards against raw consumers, and docs/stale-guidance gates.

## Critical Findings

### [Blocker] Parser And Validation Contract Is Not Yet Normative
**Sources:** Priya Zielinski/Claude, Priya Zielinski/Codex; echoed by Ibrahim Park, Nadia Sorenson, Felix Berger
**Issue:** The accepted grammar, raw TOML/JSON shape matrix, diagnostic precedence, provenance model, v2-only construct registry, inheritance behavior, and contributed-formula aggregation are not specified as an exhaustive machine-readable contract. This leaves room for hand-picked parser tests, nondeterministic multi-defect diagnostics, ambiguous dual `contract` plus `[requires]` source attribution, and inconsistent treatment of children, expansions, aspects, and future axes.
**Required change:** Add a normative `compiler_requirements_matrix.yaml` fixture schema with complete example rows, CI-enforced coverage over accepted and rejected values, TOML/JSON value kinds, host capability, legacy `contract`, v2-only construct locations, inherited/contributed sources, normalized result, source attribution, and ordered diagnostics. Define the closed byte-exact v0 grammar, diagnostic count/order rules, provenance fields, inheritance/lowering behavior, and v2 registry completeness test before implementation proceeds.

### [Major] Compiler Boundary And Durable-Write Contracts Remain Porous
**Sources:** Nadia Sorenson/Claude, Nadia Sorenson/Codex, Yuki Patel/Claude, Yuki Patel/Codex, Felix Berger/Claude, Felix Berger/Codex
**Issue:** Callers can still appear to consume raw `Contract`, `Requires.FormulaCompiler`, `Recipe.GraphWorkflow`, legacy metadata, or convenience compile wrappers instead of a canonical `CompileResult` produced under the same host capabilities and formula search path used for the durable write. Workflow-root predicate ownership is also unresolved between `internal/formula`, `internal/sourceworkflow`, and a successor package.
**Required change:** Name one workflow-root predicate owner outside the compiler package, make new raw-consumer references CI-blocking with a shrinking allowlist, and specify post-migration signatures for molecule cook, fanout, graph-apply, orders, sling, convergence, retry, and durable preflight paths. Durable producers must either compile internally from current options or consume an accepted artifact that carries host capability, search-path/options identity, formula provenance, diagnostics, and retention semantics.

### [Major] Rollout, Compatibility, And Alias Retirement Gates Are Not Executable
**Sources:** Elias Vega/Claude, Elias Vega/Codex, Lena Driscoll/Claude, Lena Driscoll/Codex, Yuki Patel/Claude, Yuki Patel/Codex, Ibrahim Park/Claude, Ibrahim Park/Codex
**Issue:** The design asserts old-reader and `bd` compatibility but does not name the supported binary set, minimum `bd` version, pinned corpus, commands, pass/fail rule, or release artifact. The `GC_NATIVE_FORMULA=false` rollback story also conflicts with the newer statement that `bd` is only a compatibility probe, and alias removal is measured from first-party conversion rather than minimum-binary-floor enforcement.
**Required change:** Add a release-blocking old-reader and native-vs-`bd` parity gate covering legacy-only, dual-declared, requires-only, unsupported-future, invalid-shape, and diagnostic-source cases. Supersede or update `engdocs/proposals/formula-migration.md`, define the permitted rollback lever for each phase, require first-party dual declarations until the minimum binary floor is enforced, and anchor alias removal to a machine-readable floor plus a defined dual-declared compatibility window.

### [Major] In-Flight Fanout And Convergence Semantics Are Ambiguous
**Sources:** Felix Berger/Claude, Felix Berger/Codex, Yuki Patel/Claude, Yuki Patel/Codex, Nadia Sorenson/Codex
**Issue:** The design says accepted roots may continue from persisted metadata, but later retry, next-iteration, fanout, on-complete, speculative-wisp, and convergence paths also need to respect current host capabilities. It does not choose the stored-state outcome after a capability downgrade or define the zero-write boundary for forbidden fragments, wisps, roots, children, dependencies, metadata, and convoy writes.
**Required change:** Choose and test the downgrade policy. Either bind same-formula in-flight work to an immutable accepted compile artifact, or fail on downgrade with named diagnostics and no new durable writes. Specify API changes so convergence and molecule stores instantiate from accepted compile results or artifacts rather than recompiling by formula name, and add zero-write tests for create, retry, fanout, speculative promotion, and next-iteration paths.

### [Major] Diagnostics Need Typed Cross-Surface And Event Contracts
**Sources:** Marta Hidalgo/Claude, Marta Hidalgo/Codex, Elias Vega/Codex, Avery Brooks/Codex
**Issue:** The shared diagnostic shape is promising, but the design lacks concrete Huma envelopes, generated OpenAPI/client expectations, direct-CLI versus API-routed CLI parity, dashboard rendering rules, event lifecycle semantics, occurrence-count behavior, `config_generation` source, and HTTP status rationale. Host capability failures may still look like parse failures or ordinary `order.failed` events depending on the surface.
**Required change:** Define typed warning/error envelopes and Huma registration, then add an end-to-end fixture proving canonical diagnostic fields and remediation survive direct CLI, API-routed CLI, API response, dashboard card, order history/feed, Event Bus payload, controller loops, and convergence. Clarify whether host-gate failures emit `order.failed`, a new typed event, or both, and make repeated-loop grouping append-only-compatible with explicit counter ownership, reset behavior, and observability.

### [Major] Pack Provenance And External Author Migration Are Under-Specified
**Sources:** Saoirse Raman/Claude, Saoirse Raman/Codex, Elias Vega/Claude, Elias Vega/Codex, Avery Brooks/Claude
**Issue:** External pack authors and consumers do not yet have a stable validation workflow for local, installed, imported, or pinned remote packs. The design also does not define a concrete source record that preserves pack binding, requested ref, locked revision or content hash, dirty state, layer winner, original path, transitive imports, and reproducibility through formula resolution and compilation.
**Required change:** Specify `gc formula validate` pack-source inputs, fetch/cache/lockfile behavior, scan scope, exit codes, JSON stability, and `migration_hints` schema. Define the owner type for resolved formula provenance and require pack-aware diagnostics that distinguish first-party, local-authored, imported pinned, and transitive sources. Enforce `[pack].requires_gc` for first-party requires-only graph packs and add resolver/import/load tests for incompatible binaries.

### [Major] Documentation And User-Facing Timing Gates Are Still Too Loose
**Sources:** Avery Brooks/Claude, Avery Brooks/Codex, Lena Driscoll/Claude, Lena Driscoll/Codex, Saoirse Raman/Claude
**Issue:** Formula `[requires]` can be confused with pack-level `[[requires]]`, `requires_gc`, and import `version` constraints, and the docs plan does not name enough concrete files, generated sources, commands, or stale-guidance exceptions. The rollout still permits diagnostics, generated help, validation output, API/dashboard surfacing, or requires-only authoring to ship before the docs/example bundle.
**Required change:** Add a hard rollout gate tying the first user-visible compiler-requirement behavior to docs, examples, generated help, and stale-guidance checks in the same PR stack. Update `docs/reference/formula.md` sections explicitly, add a comparison table for formula requirements versus pack/import constraints, provide canonical modern and dual-declared migration examples, and either supersede stale proposal/architecture docs or scan them with precise allowed exceptions.

### [Minor] Several Edge Policies Need Final Decisions
**Sources:** Elias Vega/Claude, Marta Hidalgo/Claude, Priya Zielinski/Codex, Ibrahim Park/Claude, Saoirse Raman/Claude
**Issue:** Source `contract` whitespace exactness, JSON formula support, warning dedupe observability, `formula.version_deprecated` lifecycle, declared-but-unused capability behavior, future-axis diagnostic typing, HTTP 400 versus 409 usage, and `gc.formula_compile_artifact` retention remain underspecified.
**Required change:** Add explicit policy rows and fixtures for each edge case, or declare a surface unsupported and remove it from active examples and implementation plans.

## Disagreements
- Priya Zielinski's Claude review returned `approve-with-risks`, while Codex returned `block`. I side with `block`: the missing parser fixture schema, deterministic diagnostic rules, and inheritance/contribution semantics define the implementation contract, not optional polish.
- Reviewers differed on whether the mid-loop downgrade policy should preserve same-formula work from an accepted artifact or fail after host capability changes. Both are viable, but the design must choose one and make convergence/fanout/retry behavior executable.
- Claude was more comfortable with a conceptual provenance model; Codex was stricter that the current formula-layer boundary is still path-oriented. The design should keep the concept but add an owner type and tests proving provenance survives staging, imports, shadowing, and compilation.
- Some reviewers would allow an existing diagnostic code for unsupported future capability if payloads distinguish remediation; others wanted separate codes. The required outcome is machine-distinguishable malformed syntax, unsupported future compiler capability, and unknown requirement axis cases.
- Reviewers framed legacy `version` differently: removal/sunset versus permanent ignored metadata. The design can choose either, but it must make the lifecycle explicit and align docs, reports, strict modes, and examples.

## Missing Evidence
- Gemini was skipped or absent for all persona syntheses; this synthesis uses Claude and Codex only.
- No normative validation fixture schema, complete example row, or CI exhaustiveness rule exists for compiler requirements.
- No current caller inventory is cited for raw `Contract`, `declaresGraphV2Contract`, `Requires.FormulaCompiler`, `gc.formula_contract`, `Recipe.GraphWorkflow`, or compatibility helper use.
- No named old-reader compatibility probe or native-vs-`bd` parity corpus identifies versions, commands, outputs, and pass/fail rules.
- No exact API signatures show how `CompileResult`, accepted artifacts, host capabilities, and provenance flow through molecule cook, fanout, sling, orders, graph apply, convergence, retry, and speculative wisp paths.
- No Huma/OpenAPI/dashboard/event contract proves canonical diagnostic parity across user-visible and controller surfaces.
- No concrete `ResolvedFormulaSource` or equivalent provenance owner ties pack/import resolution to formula winner selection, compile result, root metadata, and validation JSON.
- No machine-readable minimum-binary-floor artifact, external-support plan, legacy-contract inventory, or legacy-version inventory sample is provided.
- No concrete stale-guidance CI spec names scanned paths, forbidden patterns, allowed exceptions, failure messages, and quality gate placement.
- Review artifact note: the closed attempt-21 persona synthesis beads stamped their `design_review.output_path` under `.gc/design-reviews/ga-j195cx/attempt-1/persona-syntheses/`, while `.gc/design-reviews/ga-j195cx/attempt-21/persona-syntheses/` is empty. I used those closed bead outputs because they carry `gc.attempt=21` and were updated during this review run.

## Recommended Changes
1. Make the parser/validation matrix normative and CI-enforced, including grammar, diagnostics, provenance, inheritance, contributed sources, JSON if supported, and v2 construct registry coverage.
2. Lock the compiler boundary: one workflow-root predicate owner, blocking raw-consumer guard, canonical fatal-diagnostic helper, and durable-write contracts based on current `CompileResult` or accepted artifacts.
3. Define executable old-reader, first-party dual-declaration, minimum-binary-floor, native-vs-`bd`, and alias-removal gates before source conversion.
4. Specify in-flight fanout, retry, on-complete, speculative-wisp, and convergence downgrade behavior with zero-write tests.
5. Add typed diagnostic projection contracts and parity fixtures for CLI, API, dashboard, events, order history, controller loops, and convergence.
6. Define pack provenance, external validation, `migration_hints`, `[pack].requires_gc` enforcement, and pack-aware diagnostics for pinned/imported/transitive sources.
7. Tie docs, generated help, examples, first-party pack snippets, stale-guidance scanning, and user-visible diagnostics into one rollout gate.
8. Decide the lifecycle for legacy `version`, future axes, declared-but-unused requirements, warning dedupe observability, HTTP status mapping, and compile-artifact retention.
