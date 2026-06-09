# Design Review Synthesis

## Overall Verdict: block

The design is moving in the right direction, and reviewers consistently praised the shift toward typed diagnostics, normalized requirements, bounded migration artifacts, and a compiler-owned acceptance boundary. It is not ready to approve because several core contracts are still under-specified: accepted compile artifacts remain potentially forgeable or stale, the validation matrix is not implementable as written, and convergence still lacks a canonical projection over accepted formula output. Worst-verdict-wins applies: Nadia Sorenson, Priya Zielinski, and Felix Berger returned `block`.

## Consensus Strengths
- Multiple personas found the alias-window strategy for `contract = "graph.v2"` plus formula-level `[requires] formula_compiler = ">=2"` directionally sound.
- Typed diagnostics with source attribution, host-capability fields, canonical remediation text, and cross-surface projection were repeatedly identified as the right foundation.
- Reviewers agreed that normalized compiler requirements should be owned by `internal/formula` and consumed by callers rather than re-derived in CLI, API, controller, order, fanout, or convergence code.
- The proposed compatibility artifacts, legacy reports, and release gates are the right shape for migrating built-in and external packs without a flag day.
- The closed v0 requirement grammar, fail-closed unknown axes, and future-axis intent were praised, provided the future-reader and persisted-root behavior is pinned.

## Critical Findings

### [Blocker] Accepted Compile Artifacts Do Not Yet Prove Compiler Ownership
**Sources:** Nadia Sorenson; reinforced by Yuki Patel and Felix Berger

**Issue:** The design does not yet prove that durable writers can only consume compiler-minted, non-forgeable accepted compile state. `AcceptedCompileArtifact` and `AcceptCompileResult` need stricter private proof fields, immutable compile identity, zero-value fail-closed behavior, and rejection of stale or mismatched artifacts across host capability, search paths, options, vars, provenance/content hash, and fatal diagnostic state.

**Required change:** Redesign accepted artifacts so proof state is unexported and minted only by `internal/formula`; make durable writer APIs accept only accepted artifacts or sealed wrappers; add negative tests proving forged, zero-value, stale, mismatched, host-disabled, and fatal-diagnostic artifacts produce no durable writes.

### [Blocker] Validation Matrix Is Not Bounded or Auditable
**Sources:** Priya Zielinski; reinforced by Elias Vega and Marta Hidalgo

**Issue:** The promised validation matrix can explode into an impractical cross-product and contains contradictory raw-shape handling, especially for duplicate tables, duplicate scalar keys, dotted/nested/inline table forms, JSON equivalents, legacy contract spellings, and parser-boundary failures. Reviewers cannot tell which axes are exhaustive, pairwise, sentinel-only, impossible, or unsupported.

**Required change:** Replace the implicit full cross-product with named executable suites for grammar, raw shape, legacy alias, construct registry, contribution traversal, caller preflight, and projection parity. Lock generated-case counts in CI, define parser-boundary outcomes for every raw shape, and add golden rows for diagnostic ordering and count.

### [Blocker] Convergence Still Depends on an Undefined Projection Boundary
**Sources:** Felix Berger; reinforced by Yuki Patel and Ibrahim Park

**Issue:** `ProjectFormula` has no canonical, source-attributed accepted-output carrier for convergence enablement, required vars, evaluate prompt, prompt path, convergence source key, or convergence step identity. Without that contract, implementers must either keep raw convergence subset parsing or invent convergence-specific compiler leakage, reopening the drift this design is meant to close.

**Required change:** Define the canonical convergence projection input from `CompileResult` / `AcceptedCompileArtifact`, then require create, retry, next iteration, missing-child repair, and speculative-wisp paths to compile, accept/load, project, and validate before any durable write. Delete or shim the legacy convergence parser over canonical output and add a static guard against raw requirement or convergence TOML reads.

### [Major] Caller Migration and Durable Write Preflight Are Incomplete
**Sources:** Yuki Patel, Nadia Sorenson, Felix Berger, Priya Zielinski

**Issue:** The design does not yet provide a complete caller inventory and zero-write contract for sling launch, order dispatch, fanout fragments, retry/on-complete, convergence create/retry, API previews, dashboard paths, bd shell-outs, hooks, convoys, and child/root writes. Fanout is especially risky because metadata such as `gc.fanout_state` must not mutate before all fragments compile and validate.

**Required change:** Add a checked caller manifest and static guards for raw consumers and durable bd shell-outs. Require preflight-before-write tests for every durable producer, including root/control metadata, child beads, convoy links, continuation metadata, retry metadata, and fanout state.

### [Major] Migration and Rollout Gates Need Executable Ownership
**Sources:** Elias Vega, Lena Driscoll, Saoirse Raman, Avery Brooks

**Issue:** Compatibility and rollout gates are conceptually correct but lack concrete owners, schemas, commands, and phase ordering. Missing pieces include `[pack] requires_gc` resolver/import enforcement, first-party dual-declaration inventory and CI guard, old-reader compatibility artifacts, external support artifacts, `version = 2` plus v2-only legacy reports, and an explicit meaning for "two minor releases."

**Required change:** Add a pre-Phase-6 pack-floor enforcement phase, seed the compatibility/minimum-floor/external-support artifacts with conservative values, create CI guards for first-party dual declarations and stale legacy additions, and anchor alias removal to verifiable release evidence plus external support status.

### [Major] Diagnostics and Operator Projection Are Not Fully Pinned
**Sources:** Marta Hidalgo, Priya Zielinski, Avery Brooks

**Issue:** Background warnings, event emission, `config_generation`, suppression reset behavior, dashboard grouping, API state, TypeScript fields, and CLI exit-code policy remain ambiguous. Deprecation warnings could either spam controller/order loops, disappear from operator views, or diverge across CLI, API, dashboard, and Event Bus projections.

**Required change:** Decide whether warning diagnostics publish events. If they do, register event constants and payloads; if not, remove warning-event language. Define `config_generation`, `OnceKey`, grouping, occurrence counts, dashboard/API source of truth, command-class exit codes, and cross-surface golden fixtures for fatal and warning diagnostics.

### [Major] Future Capability and Persisted Root Semantics Are Too Implicit
**Sources:** Ibrahim Park; reinforced by Felix Berger

**Issue:** Future compiler capability roots and future requirement axes can be persisted before old binaries know how to execute them. A boolean workflow-root predicate can silently collapse unknown future workflow roots to non-workflows, while accepted artifact reuse and cleanup behavior for higher capabilities or unknown axes remains undefined.

**Required change:** Introduce typed root facts that distinguish non-workflow, default-capability workflow, known graph workflow, and unknown future capability workflow. Add old-reader fixtures for `gc.formula_compiler_capability=3` and future axes, and state lifecycle rules for observation, cleanup, retry, fanout, continuation, child writes, and new compiles.

### [Minor] Documentation and Terminology Gates Are Still Too Loose
**Sources:** Avery Brooks, Saoirse Raman, Lena Driscoll

**Issue:** The docs plan does not yet guarantee that `[pack].requires`, `[pack].requires_gc`, formula-level `[requires]`, and `[pack] schema = 2` are taught distinctly before user-visible diagnostics ship. Stale guidance in formula reference docs, PackV2 docs, guides, examples, generated help, architecture docs, and migration proposals can preserve the old mental model.

**Required change:** Move docs/examples/generated-help before or atomically with the first user-visible diagnostic projection. Add an executable stale-guidance gate with explicit scan roots, forbidden patterns, allowlists, owner, generated-artifact handling, and file:line failures.

## Disagreements
- Three personas had Claude/Codex verdict disagreements that drove the global block: compiler boundary, parser validation matrix, and convergence subset. In each case, the block assessment is persuasive because the disagreement concerns implementability or a boundary escape, not a stylistic preference.
- Some reviewers saw the generated matrix and accepted-artifact direction as promising enough for `approve-with-risks`; others found the same areas insufficiently enforceable. The synthesis keeps the approach but requires executable contracts, static guards, and negative tests before approval.
- Personas differed on whether external-author migration requires a canonical fixer command. The design may either specify `gc formula migrate` / `gc formula fix --apply` with exact dry-run, rewrite, idempotence, and ownership rules, or remove `safe_automatic_edit` and make hints advisory.
- Reviewers split emphasis between event spam, dashboard state, and documentation sequencing for diagnostics. These are not conflicting findings: the design must define warning/event semantics first, then ensure every user-visible projection has docs and golden tests.
- Current-attempt artifact note: the attempt-33 `persona-syntheses/` directory was empty, while the current attempt's persona synthesis beads stamped their outputs under `attempt-1/persona-syntheses/`. This synthesis used those bead-declared outputs because they were the only complete set of ten persona syntheses for `gc.attempt=33`.

## Missing Evidence
- No Gemini artifacts were present for any persona; the workflow allowed this via `skip_gemini=true`, but there is no third-model tie-breaker.
- No complete durable writer inventory proves every root, child, hook, convoy, retry, fanout, order, convergence, and API/dashboard path consumes only accepted artifacts.
- No bounded validation-suite definition, generated-case counts, impossible-row mechanism, or checked decoded-field classification registry is shown.
- No raw fixtures cover duplicate/scalar/table/dotted/nested/inline/JSON requirement shapes, legacy contract whitespace/case behavior, and diagnostic order/count interactions.
- No concrete convergence projection API, artifact metadata contract, reuse identity checks, artifact failure diagnostics, or legacy-root migration rule is specified.
- No source-controlled schemas/templates are shown for compatibility, minimum-floor, external-support, requirement-diff, or external-author release artifacts.
- No pack-level `[pack] requires_gc` enforcement phase or old-reader corpus proves older binaries reject incompatible packs before formula selection and durable writes.
- No executable stale-guidance gate, PackV2 policy for `pack.requires`, or side-by-side documentation example disambiguates the three `requires` surfaces.
- No typed root-fact model or old-reader fixture covers future compiler capability roots and unknown future requirement axes across sourceworkflow, CLI/API, dashboard, controller/order scans, cleanup, retry, and continuation paths.

## Recommended Changes
1. Fix the compiler boundary first: make accepted artifacts non-forgeable, identity-bound, zero-value fail-closed, and required by every durable writer API.
2. Define convergence as a projection over canonical accepted formula output, then retire or shim the raw convergence subset parser and add zero-write preflight tests.
3. Replace the unbounded validation matrix with named executable suites, generated-case count locks, raw-shape parser-boundary rules, and a decoded-field classification registry.
4. Add the caller manifest and static guards for raw consumers, durable bd shell-outs, workflow-root predicates, and formula metadata writers.
5. Pin diagnostics end to end: warning-event policy, `config_generation`, `OnceKey`, grouping, dashboard/API state, CLI exit codes, Event Bus payload registration, and cross-surface golden fixtures.
6. Make rollout gates executable: `[pack] requires_gc` enforcement, first-party dual-declaration CI, old-reader compatibility corpus, external support artifact, release-floor artifact, and legacy-version/contract reports.
7. Define future capability observation and lifecycle semantics with typed root facts and old-reader fixtures for higher compiler capability and future axes.
8. Move docs, examples, generated help, stale-guidance checks, and terminology updates before any user-visible diagnostic or source-migration phase.
9. Decide whether external authors get an automatic fixer; either specify the command and exact rewrites or remove `safe_automatic_edit`.
10. Repair the design-review artifact path bug so persona synthesis beads for a review attempt write under that same attempt directory and downstream synthesis can rely on `$ATTEMPT_DIR/persona-syntheses`.
