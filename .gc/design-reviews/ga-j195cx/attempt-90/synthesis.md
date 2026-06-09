# Design Review Synthesis

## Overall Verdict: block

The design has a coherent direction: formula source declares compiler requirements, the active Gas City binary evaluates host capability, and durable runtime state should be written only from compiler-owned accepted artifacts. It still cannot pass this review because multiple load-bearing contracts remain non-executable: parser coverage can miss real durable writers, operator diagnostics can flood or drift, and caller migration can bypass accepted artifacts through raw or preview data.

## Consensus Strengths
- Reviewers consistently praised the core boundary: formulas declare `requires.formula_compiler`, but formulas do not select compiler implementations.
- Multiple personas agreed that normalization belongs in `internal/formula`, with typed artifacts such as host capabilities, normalized requirements, compile results, and accepted compile artifacts consumed by CLI/API/sling/orders/convergence projections.
- Reviewers supported the legacy `contract = "graph.v2"` migration shape as a deprecated alias for `[requires] formula_compiler = ">=2"`, provided conflicts, warnings, and removal gates are fixture-locked.
- The pack ecosystem direction is sound: pack ref/SHA, locked revision, content hash, import binding, and pack provenance are the reproducibility boundary, not formula artifact semver.
- The rollout plan contains the right families of evidence: first-party inventory, external-support JSON, alias-removal reports, docs checks, old-reader compatibility probes, and strict release gates.

## Critical Findings

### [Blocker] Parser and Construct Coverage Can Give False Confidence
**Sources:** Priya Zielinski / parser-validation-matrix, Felix Berger / convergence-subset-reviewer, Ibrahim Park / future-capability-architect
**Issue:** The validation matrix is not yet tied to a repository-wide, per-occurrence caller manifest, and v2-only construct detection is not fully source-attributed across condition-disabled steps, `expand_vars`, retry/check/on_complete, compose/aspect/import contributions, workflow metadata, and convergence paths. A fixed caller count can pass while durable writers or prompt producers still bypass fatal requirement diagnostics.
**Required change:** Generate caller/preflight suites from a checked-in repository-wide manifest, define presence/materialization semantics before condition filtering, require raw source attribution for every triggering construct, and add named reflection/completeness tests plus edge-case fixtures for future capabilities, conflicts, multiple constructs, JSON behavior, and unregistered `gc.*` metadata.

### [Blocker] Operator Diagnostic Cadence and Rollups Are Not Executable
**Sources:** Marta Hidalgo / operator-diagnostics-gate
**Issue:** `FormulaDiagnosticBurstBudget`, producer policies, rollup lifecycle, event cadence, warning persistence, and shared payload ownership are referenced but undefined. Without exact budgets and bounded rollup rules, order dispatch, convergence, retries, fanout, CI, API, and dashboard surfaces can flood operators or show contradictory remediation state.
**Required change:** Define burst-budget fields, defaults, reset semantics, producer policies, typed payload/schema ownership, rollup bounds and lifecycle fields, and fixtures for repeated scans, restarts, reloads, config flaps, many-subject disabled hosts, legacy-contract parity, immutable accepted warning snapshots, and mutable producer warning counters.

### [Blocker] Durable Writers Can Still Bypass Accepted Artifacts
**Sources:** Yuki Patel / caller-integration-inventory, Nadia Sorenson / compiler-boundary-invariant, Felix Berger / convergence-subset-reviewer
**Issue:** The migration has not yet proven that every root, child, dependency, hook, convoy, retry, fanout, order, convergence, workflow-root, and prompt/formula producer writes through compiler-owned accepted artifacts rather than raw `Recipe`, preview `CompileResult`, caller-built metadata, shell-out output, or legacy prompt instructions.
**Required change:** Make accepted artifacts the only authority at durable writer APIs, commit a manifest and static guards that cover Go, dashboard TypeScript, packs, examples, tutorials, and prompt templates, classify all first-party `gc bd mol ...` producer references, and add zero-write host-capability transition fixtures for every multi-write entry point.

### [Major] Rollout Sequencing Has Contradictory or Overloaded Gates
**Sources:** Lena Driscoll / rollout-sequencing-reviewer, Saoirse Raman / pack-versioning-ecosystem, Elias Vega / contract-migration-guardian
**Issue:** Phase 2 bundles too many user-visible surfaces without a clear rollback or visibility control, Phase 4a says writers are accepted-artifact-only while callers still use legacy compile wrappers, and Phase 8 requires-only conversion is not cleanly separated from Phase 9 alias removal. Pack-floor and alias-window clocks also depend on provenance evidence that may not exist when the clock starts.
**Required change:** Split or gate Phase 2 around old-reader/probe compatibility, resolve the Phase 4a transition model, define Phase 4b-4g order or independence fixtures, separate conversion and alias-removal gates, require non-placeholder minimum-floor values before pack-floor declarations, and prevent alias-window start until packman schema 2 or equivalent provenance evidence is recorded.

### [Major] Migration Evidence Artifacts Need Schemas and Owners
**Sources:** Elias Vega / contract-migration-guardian, Saoirse Raman / pack-versioning-ecosystem, Lena Driscoll / rollout-sequencing-reviewer
**Issue:** Several release gates depend on artifacts whose contracts are still incomplete, including `gc formula repair-root-artifact`, active-root reports, external-support JSON, alias-window-start records, alias-drain reports, min-floor evidence, public notice, and supported-reader lists.
**Required change:** Specify command contracts, JSON schemas, paths, generator commands, owners, exit codes, refusal diagnostics, release-captain fields, and fixture coverage for every artifact consumed by migration or alias-removal gates. Gates must consume JSON as canonical input, with Markdown generated only as documentation.

### [Major] Convergence Boundary and Repair Semantics Are Ambiguous
**Sources:** Felix Berger / convergence-subset-reviewer, Priya Zielinski / parser-validation-matrix, Yuki Patel / caller-integration-inventory
**Issue:** The design does not yet pin the convergence compile/accept/cook sequence, artifact reuse identity, `evaluate_prompt` ownership and hashing, blocked-loop metadata, or the conflict between generic `gc.formula_compile_artifact` and convergence-specific artifact keys.
**Required change:** Add before/after convergence pseudocode, define `CompileWithResult -> AcceptCompileResult -> ProjectAcceptedFormula -> ValidateProjection` ownership, include `evaluate_prompt` in artifact identity, choose canonical repair artifact keys or dual-stamp compatibility with precedence, and fixture status/API/dashboard/reconciler/manual retry behavior for formula requirement failures.

### [Major] Documentation and Terminology Gates Remain User-Visible Risks
**Sources:** Avery Brooks / docs-dx-terminology, Ibrahim Park / future-capability-architect, Saoirse Raman / pack-versioning-ecosystem
**Issue:** Public docs still risk teaching stale `version` or `contract` patterns, the glossary blurs formula `[requires]` with pack `requires_gc`, and user-visible diagnostics could ship before reference docs, generated help, dashboard/API surfaces, doctests, and stale-guidance checks land.
**Required change:** Make the docs/check bundle a hard predecessor to user-visible diagnostics and `[requires]` behavior; fix glossary terms; add remediation for legacy `version = 2`; add stale-guidance fixtures for `contract`, `version`, `GC_NATIVE_FORMULA=false`, and `formula_v2`; and require checked placeholder/doctest behavior.

### [Major] Future Capability Contracts Need Release Invariants
**Sources:** Ibrahim Park / future-capability-architect, Priya Zielinski / parser-validation-matrix
**Issue:** The v0 requirement grammar is directionally simple, but released construct capability rows, explicit `>=1` provenance, unknown axes, future non-scalar axes, `RequirementSource`, and integer bounds are not pinned strongly enough to avoid reinterpretation later.
**Required change:** Make released construct capability rows immutable, validate first-party graph packs with the latest registry-owning binary before publication, separate behavioral identity from provenance for explicit defaults, define v0 flat-scalar limits and schema-extension rules, fixture unknown axes and mixed manifests, keep `RequirementSource` append-only and non-runtime, and pin integer overflow diagnostics.

### [Minor] Current Attempt Artifact Paths Are Drifted
**Sources:** global synthesis artifact inspection
**Issue:** `.gc/design-reviews/ga-j195cx/attempt-90/persona-syntheses/` is empty, while the ten closed persona synthesis beads for `gc.attempt=90` stamped fresh output paths under `.gc/design-reviews/ga-j195cx/attempt-1/persona-syntheses/`.
**Required change:** Repair the design-review workflow so persona synthesis beads write under the current `$ATTEMPT_DIR/persona-syntheses` and stamp matching output paths. This synthesis used the bead-declared outputs because they were complete, current, and metadata-linked to attempt 90.

## Disagreements
- Three personas returned `block` and seven returned `approve-with-risks`; worst-verdict-wins therefore makes the global verdict `block`.
- Parser validation, operator diagnostics, and caller integration had the sharpest Claude/Codex severity disagreement. In each case the stricter finding controls because the issue is not taste or sequencing preference; it is an executable-proof gap that could let invalid formulas write durable state, flood operators, or bypass host capability.
- Several personas found the architecture direction acceptable while objecting to missing fixtures. My assessment is that those fixture gaps are design gaps here because the proposal makes strict migration gates, release artifacts, and static guards load-bearing.
- Reviewers disagreed on whether some issues are conceptual blockers or rollout risks. The synthesis treats the compiler boundary as conceptually sound but blocks approval until the contracts that enforce it are named, owned, and testable.
- Kimi 2.6 was absent from all persona syntheses. That is allowed by the workflow when skipped, but it lowers diversity of review evidence.

## Missing Evidence
- Current-attempt persona synthesis files under `.gc/design-reviews/ga-j195cx/attempt-90/persona-syntheses/`; outputs were found only through the closed persona beads' stamped paths.
- Generated per-occurrence caller manifest and CI guard coverage for Go, dashboard TypeScript, first-party packs, examples, tutorials, prompt templates, and legacy `gc bd mol ...` producers.
- Exact `FormulaDiagnosticBurstBudget`, producer-policy, rollup, payload, event/no-event, and warning persistence schemas.
- Raw construct/source attribution contract and reflection-backed completeness test for every decoded formula field and contribution path.
- `gc formula repair-root-artifact`, active-root report, alias-window-start, external-support, alias-drain, min-floor, and supported-reader artifact schemas.
- Old-reader/probe compatibility corpus for dual-declared formulas, pack floors, JSON loader paths, installed packs, active `bd`/native probes, and pinned external packs.
- Convergence accepted-artifact pseudocode, artifact key policy, prompt hashing rule, blocked-loop metadata contract, and repair fixtures.
- Executable docs/check bundle, stale-guidance matcher fixtures, placeholder registry/doctests, and rewritten reference/architecture/author docs.
- Future-axis and capability-growth fixtures for immutable construct rows, unknown axes, mixed persisted manifests, explicit `>=1` provenance, integer overflow, and released `RequirementSource` durability.

## Recommended Changes
1. Land the repository-wide caller manifest and static guards first, then derive parser preflight coverage and durable-writer migration suites from that single manifest.
2. Define and fixture-lock accepted-artifact authority for every durable writer, including convergence, fanout, retry/on_complete, order scans, root repair, prompt-spawned successors, and first-party legacy mol producers.
3. Specify the operator diagnostic contract end to end: burst budgets, producer policies, typed payloads, rollup bounds, warning state split, CLI behavior, API/dashboard projection, and event decision.
4. Complete the parser construct registry and source-attribution contract, including condition-disabled constructs, compose/aspect/import contributions, workflow metadata, JSON behavior, and diagnostic precedence.
5. Resolve rollout sequencing: split Phase 2 preconditions, choose the Phase 4 transition model, separate Phase 8 conversion from Phase 9 alias removal, and anchor pack-floor declarations to concrete release evidence.
6. Add schemas, paths, owners, and fixtures for every release evidence artifact consumed by migration or alias-removal gates.
7. Replace convergence prose with executable before/after flow, artifact identity rules, prompt hashing behavior, blocked-loop metadata, and repair-key compatibility rules.
8. Make docs/reference/generated-surface cleanup a hard predecessor to user-visible diagnostics or `[requires]` behavior.
9. Add future-capability invariants for immutable released construct rows, latest-binary publication checks, explicit default provenance, unknown axes, schema extensions, and integer grammar limits.
10. Fix the design-review workflow artifact path drift so future global syntheses can rely on `$ATTEMPT_DIR/persona-syntheses`.
