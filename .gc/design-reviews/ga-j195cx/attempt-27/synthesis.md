# Design Review Synthesis

## Overall Verdict: block

The design is directionally strong: reviewers consistently endorsed moving formula compiler requirements into a canonical `internal/formula` path, keeping formulas declarative, and preserving pack-level compatibility as the release boundary. The review blocks on one concrete validation gap: the proposed matrix can still miss a legacy `version` path that allows v2-only constructs without an explicit `requires.formula_compiler >=2`, while several migration, diagnostics, convergence, and documentation gates remain major risks that must be made executable before implementation.

## Consensus Strengths
- Multiple personas praised the central architectural direction: formulas declare requirements, the active Gas City binary decides capability satisfaction, and role or caller-specific decision logic stays out of Go.
- The shared typed diagnostic shape, normalized requirement model, source attribution, remediation strings, and zero-write preflight intent were repeatedly identified as the right foundation.
- The rollout concept of dual-declared compatibility, old-reader support, pack floors, and release gates was seen as the correct migration shape, even though the gates need concrete schemas and owners.
- Reviewers agreed that convergence, sling, orders, CLI, API, dashboard, and controller paths should consume compiler-owned typed results rather than reinterpreting raw `contract`, raw `[requires]`, or workflow metadata.
- The fail-closed stance for unknown requirement axes, byte-exact accepted grammar, and source provenance is strong and aligned with future capability growth.

## Critical Findings

### [Blocker] Validation matrix can preserve a legacy version bypass
**Sources:** Priya Zielinski / `03-parser-validation-matrix` synthesis; Codex raw review as summarized by Priya

**Issue:** The proposed executable matrix does not explicitly cross omitted `version`, `version = 1`, and legacy `version = 2` with every v2-only construct and omitted, empty, or `>=1` requirements. Priya reports a concrete risk that an existing `version = 1` plus `[steps.check]` expectation can still compile as a legacy molecule, allowing implementation to keep a known bypass while satisfying the visible parser matrix.

**Required change:** Add formula-version coverage to the compiler requirements matrix for omitted version, `version = 1`, and legacy `version = 2` crossed with every v2-only construct and requirement state. Remove or update legacy implementation expectations that contradict the new requirement grammar, including any `version = 1` plus v2-only construct fixtures. Any v2-only construct must require `requires.formula_compiler >=2` unless the design explicitly reclassifies that construct as v1-compatible.

### [Major] Compiler authority and durable-write boundaries are not enforceable enough
**Sources:** Nadia Sorenson / `01-compiler-boundary-invariant`; Yuki Patel / `05-caller-integration-inventory`; Felix Berger / `09-convergence-subset-reviewer`; Priya Zielinski / `03-parser-validation-matrix`

**Issue:** The design says `internal/formula` owns normalization and host-capability satisfaction, but several caller and durable-write surfaces can still reconstruct decisions or bypass the accepted compile identity. The reviews call out `Compile(...) (*Recipe, error)`, `Recipe.GraphWorkflow`, raw `Contract`, raw `Requires`, `gc.formula_*` reads, global formula flags, bare `*formula.Recipe` durable APIs, and convergence projection as drift risks.

**Required change:** Freeze a checked-in static raw-consumer allowlist at the start of caller migration and make additions fail CI. Require durable writers to receive an accepted compile artifact or `CompileResult`, not an unchecked bare recipe. Add compiler-owned durable-write helpers or artifact verification, with parity tests proving durable producers do not branch on raw requirement, contract, metadata, or host-capability fields.

### [Major] Migration and release gates are named but not executable
**Sources:** Elias Vega / `02-contract-migration-guardian`; Lena Driscoll / `06-rollout-sequencing-reviewer`; Saoirse Raman / `07-pack-versioning-ecosystem`; Yuki Patel / `05-caller-integration-inventory`; Ibrahim Park / `10-future-capability-architect`

**Issue:** The migration relies on compatibility artifacts, old-reader support, dual declarations, `bd` or `GC_NATIVE_FORMULA` probes, pack floors, external support, and alias-removal criteria, but those gates lack schemas, owners, commands, supported-reader sets, and pass/fail behavior. The alias-removal clock is especially weak if it starts from first-party `[requires]` conversion rather than enforcement of a minimum binary floor and a measured compatibility window.

**Required change:** Define concrete schemas, owners, update commands, and blocking behavior for `docs/release/formula-compiler-compatibility.yaml`, `docs/release/formula-compiler-min-floor.json`, and `docs/release/formula-compiler-external-support.md`. Re-anchor alias removal to the enforced minimum-binary-floor artifact plus a defined dual-declared compatibility window. Decide whether `GC_NATIVE_FORMULA=false` and any `bd` path are removed, validation-only, or supported runtime fallback, and gate that decision with a byte-level parity corpus.

### [Major] Diagnostics, events, CLI, API, and dashboard parity is underdefined
**Sources:** Marta Hidalgo / `04-operator-diagnostics-gate`; Avery Brooks / `08-docs-dx-terminology`; Elias Vega / `02-contract-migration-guardian`; Lena Driscoll / `06-rollout-sequencing-reviewer`

**Issue:** The design has the right diagnostic concept but not a complete operator-facing contract. Host-capability failures can still look like parse failures, warning publication semantics are ambiguous, repeated background failure grouping is not implementable against an append-only Event Bus as written, and dashboard and generated TypeScript behavior are not specified with executable parity fixtures.

**Required change:** Define typed Huma response envelopes, OpenAPI/client coverage, event constants and registered payloads, CLI exit-code mapping, canonical message/remediation rendering, and dashboard surfaces for disabled host capability and legacy-contract warnings. Add cross-surface golden tests covering direct CLI, API-routed CLI, Huma responses, dashboard rendering, controller/order/convergence events, and generated dashboard types.

### [Major] Convergence integration needs a clearer package boundary and zero-write order
**Sources:** Felix Berger / `09-convergence-subset-reviewer`; Nadia Sorenson / `01-compiler-boundary-invariant`; Yuki Patel / `05-caller-integration-inventory`

**Issue:** Reviewers agree convergence should compile through canonical `CompileWithResult` before durable writes, but the proposed `ConvergenceMetadata()` boundary risks putting convergence-specific API surface into `internal/formula`. Existing convergence checks also need a compiled step set, evaluate-step and evaluate-prompt validation, artifact persistence, retry behavior, and fail-closed handling for corrupt accepted artifacts.

**Required change:** Keep formula output generic and let `internal/convergence` own convergence-specific projection and validation over compiled fields. Specify call order as compile, requirement satisfaction, convergence projection, convergence-domain validation, then root, metadata, wisp, retry, iteration, or child writes. Anchor accepted compile artifacts to convergence roots with a metadata key, atomic write timing, retry reuse path, and fail-closed diagnostics.

### [Major] Future capability growth is not yet a stable contract
**Sources:** Ibrahim Park / `10-future-capability-architect`; Priya Zielinski / `03-parser-validation-matrix`; Saoirse Raman / `07-pack-versioning-ecosystem`

**Issue:** The v0 grammar and registry risk becoming a boolean v2 boundary rather than a durable capability model. Unsupported future capability values such as `>=3` need machine-distinguishable treatment from malformed expressions, old readers need fail-closed behavior for future canonical metadata, and `HostCapabilities` lacks a chosen growth model for future axes.

**Required change:** Split diagnostics for malformed expressions, syntactically valid but unsupported future capability, and unknown axes. Generalize the v2-only construct registry to a minimum-capability registry. Choose a `HostCapabilities` growth model and include a worked second-axis example covering normalized state, provenance, diagnostics, metadata, validation-matrix dimensions, pack floors, and workflow-root predicate behavior.

### [Major] Documentation and terminology can lag user-visible behavior
**Sources:** Avery Brooks / `08-docs-dx-terminology`; Lena Driscoll / `06-rollout-sequencing-reviewer`; Elias Vega / `02-contract-migration-guardian`; Saoirse Raman / `07-pack-versioning-ecosystem`

**Issue:** The docs plan is directionally good but not an enforceable prerequisite. User-visible diagnostics could ship before reference docs, tutorials, examples, generated help, and stale-guidance checks explain formula `[requires]`, pack-level `requires_gc`, `city.toml` `[daemon] formula_v2`, legacy `contract`, and legacy `version` consistently.

**Required change:** Add a hard rollout gate requiring the docs, examples, generated help, and stale-guidance bundle before, or atomically with, the first user-visible diagnostic projection. Provide exact TOML examples for formula `[requires] formula_compiler = ">=2"`, pack-level `[pack].requires_gc`, and city `[daemon] formula_v2`. Decide whether legacy `version` is permanent legacy metadata or follows a measurable removal window.

### [Minor] External pack validation, shadowing, provenance, and warning details need sharper contracts
**Sources:** Saoirse Raman / `07-pack-versioning-ecosystem`; Marta Hidalgo / `04-operator-diagnostics-gate`; Elias Vega / `02-contract-migration-guardian`; Ibrahim Park / `10-future-capability-architect`

**Issue:** Remote pack validation lacks fetch/cache/lockfile/mutation/security limits, pinning versus layer shadowing lacks default diagnostics and exit-code behavior, `ContentHash` is not precisely defined, and OnceKey/LRU warning grouping lacks producer ownership, reset triggers, capacity rationale, and eviction observability.

**Required change:** Add an external validation execution contract, a shadowing diagnostics policy, a precise `ContentHash` definition, resolver-to-compiler provenance handoff tests, and OnceKey/LRU semantics with observable eviction behavior.

## Disagreements
- Priya's lane had the only verdict disagreement: Claude recommended `approve-with-risks`, while Codex recommended `block`. I side with `block` because the legacy `version` bypass is concrete and can let the implementation pass an incomplete matrix.
- Marta's reviewers disagreed on whether host-capability failures require different HTTP statuses or can share statuses with a typed discriminator. The exact status split is less important than stable machine-readable classification and parity tests across CLI, API, dashboard, and events.
- Felix's reviewers disagreed on whether `formula.CompileResult` should expose `ConvergenceMetadata()`. I assess the safer boundary as generic compiled fields in `internal/formula` plus convergence-owned projection and validation.
- Saoirse's reviewers framed shadowed pinned-pack formulas differently: optional cleanup detail versus default validation signal. The design should pick an explicit default safety signal and may put verbose cleanup behind an option.
- Ibrahim's reviewers offered two viable future-axis shapes for `HostCapabilities`: guarded named fields or a typed-axis representation. Either can work, but the design must choose one and prove it with a second-axis example.
- Avery's reviewers differed on terminology emphasis. Use one human-facing concept name, while preserving exact TOML key spelling wherever keys are discussed.

## Missing Evidence
- Gemini artifacts were absent for all persona syntheses in this attempt; this is acceptable under the workflow because Gemini was skipped, but it leaves no third-source tie-breaker.
- The current attempt directory `.gc/design-reviews/ga-j195cx/attempt-27/persona-syntheses` was empty. The 10 required persona syntheses were read from the closed persona beads' `design_review.output_path` values under `.gc/design-reviews/ga-j195cx/attempt-1/persona-syntheses`.
- No compatibility corpus results prove supported old Gas City binaries or any `bd` probe accept dual-declared formulas and reject incompatible requires-only formulas as intended.
- No checked-in schemas or examples exist for the release compatibility artifacts that the rollout treats as gates.
- No current caller inventory exists for raw requirement, contract, metadata, global flag, graph workflow, and bare-recipe durable API consumers.
- No cross-surface golden fixtures prove identical diagnostics through direct CLI, API-routed CLI, Huma endpoints, dashboard, order events, convergence, and generated TypeScript.
- No concrete docs inventory or stale-guidance gate defines scanned paths, denied patterns, and allowed migration, legacy, historical, or fixture exceptions.
- No future-axis appendix demonstrates how another requirement axis flows through normalized state, host capabilities, diagnostics, provenance, persisted metadata, validation matrix, pack floors, and workflow-root predicates.

## Recommended Changes
1. Fix the blocking validation gap by expanding `compiler_requirements_matrix.yaml` to cover omitted, legacy `version = 1`, and legacy `version = 2` combinations with every v2-only construct and requirement state; remove any tests that preserve the bypass.
2. Make compiler-owned authority enforceable with a blocking raw-consumer guard, accepted compile artifacts for durable writes, and parity tests for caller paths and durable producers.
3. Turn rollout gates into executable artifacts with schemas, owners, commands, supported-reader sets, and blocking behavior; re-anchor alias removal to an enforced minimum binary floor plus a measured dual-declared window.
4. Define the diagnostic wire and event contract across CLI, API, dashboard, orders, convergence, and generated clients, then add cross-surface golden fixtures.
5. Resolve convergence boundaries by keeping formula output generic, moving convergence-specific projection to `internal/convergence`, and enforcing zero-write preflight order.
6. Complete the forward-capability story: typed unsupported-future diagnostics, a minimum-capability construct registry, old-reader fail-closed behavior, and a chosen `HostCapabilities` growth model.
7. Gate user-visible diagnostics on docs, examples, generated help, and stale-guidance checks, including canonical TOML snippets for formula `[requires]`, pack-level `requires_gc`, and city `[daemon] formula_v2`.
8. Specify external pack validation, layer shadowing diagnostics, provenance `ContentHash`, warning dedupe, and OnceKey/LRU operator observability.
