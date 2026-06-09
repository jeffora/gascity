# Design Review Synthesis

## Overall Verdict: block

The design is moving in a coherent direction: centralize formula capability interpretation in `internal/formula`, normalize requirements into typed compiler results, and migrate legacy `contract = "graph.v2"` through an explicit compatibility window. The review blocks because two implementation-critical areas remain underspecified: background/operator diagnostic behavior can still produce unbounded or contradictory events, and caller integration still leaves live ambiguity around `bd` fallback, workflow-root predicates, query selectors, fanout fragments, and convergence writes.

## Consensus Strengths
- Multiple personas agreed that `CompileResult`, normalized requirements, typed diagnostics, and a compiler-owned capability model are the right direction for eliminating caller-side formula interpretation.
- The compatibility-window approach for `contract = "graph.v2"` is directionally sound, especially the intent to dual-declare during migration, warn before removal, and keep legacy workflow-root metadata readable while callers migrate.
- The proposed diagnostic shape is a strong foundation for cross-surface parity, provided the design makes projection, event ownership, source attribution, and repeat-emission rules contractual.
- The provenance direction for packs, imports, formula layers, and workflow roots is useful and should become a typed flow rather than remaining path-only resolution data.
- Reviewers consistently supported generated validation matrices and static guards as the right way to prevent future drift, if the exact corpus, denied patterns, allowed exemptions, and gates are specified.

## Critical Findings

### [Blocker] Background Diagnostics Can Still Flood Or Contradict Operator Surfaces
**Sources:** Marta Hidalgo / operator-diagnostics-gate (block); Avery Brooks / docs-dx-terminology; Priya Zielinski / parser-validation-matrix; Lena Driscoll / rollout-sequencing-reviewer

**Issue:** The design defines warning suppression but not repeated fatal diagnostic cadence for automatic loops. A due order or controller path with `[daemon] formula_v2 = false` can emit the same fatal disabled-capability diagnostic on every poll. The design also contradicts itself about whether formula preview/validate paths publish Event Bus events or only return typed diagnostics synchronously. Without a single producer/consumer contract, CLI, API, dashboard, controller, order, and convergence paths can diverge or spam.

**Required change:** Define separate diagnostic emission policies for interactive attempts and automatic background loops. Resolve preview/validate event ownership, reserve Event Bus emissions for named durable/background producers, and add fixture-equality tests for `formula.compiler_requirement_unsatisfied` across CLI, API-routed CLI, API responses, dashboard rendering, order events, and controller/convergence surfaces. Add a repeated-order-scan test proving identical failures are rate-limited or grouped and do not create wisps.

### [Blocker] Caller Integration Still Allows Divergent Runtime Policy
**Sources:** Yuki Patel / caller-integration-inventory (block); Nadia Sorenson / compiler-boundary-invariant; Felix Berger / convergence-subset-reviewer; Lena Driscoll / rollout-sequencing-reviewer

**Issue:** The design keeps `bd` shell-out/fallback language without deciding whether it is a live production path. It also leaves workflow-root, graph-workflow, and formula-provenance predicates ambiguous, and it does not cover query-time selectors that currently filter raw metadata before any post-fetch predicate can run. Fanout expansion and convergence flows remain unclear when host capabilities change after a root exists.

**Required change:** Either remove `bd` shell-out fallback from the live runtime design or define exact per-entry-point behavior for old, dual-declared, and requires-only formulas across sling, API sling, orders, convergence, and fanout under both host capability states. Split formula provenance, workflow root, and graph workflow root into typed predicates/query criteria with explicit owning packages and static-guard exemptions. Specify fanout fragment compilation with current host capabilities and a zero-durable-write boundary before fragment children, dependencies, or fanout metadata are written.

### [Major] Parser And Validation Matrix Is Still Too Prose-Driven
**Sources:** Priya Zielinski / parser-validation-matrix; Ibrahim Park / future-capability-architect; Elias Vega / contract-migration-guardian

**Issue:** The design relies on examples and grouped categories rather than a normative generated matrix. It does not fully specify diagnostic ordering, warning-plus-fatal behavior, raw TOML failure classification, exact accepted `requires.formula_compiler` strings, legacy `contract` lexical behavior, expansion/aspect source attribution, or metadata trigger exactness.

**Required change:** Replace the checklist with a generated fixture table covering requirement source, raw TOML shape, legacy contract state, v2 construct registry entry, root versus contributed formula source, host capability, normalized result, diagnostic list, and source attribution. Make the v0 grammar byte-exact, preferably only `">=1"` and `">=2"`, and classify each malformed TOML edge as structured formula diagnostic or plain TOML parse/decode error.

### [Major] Legacy Contract, JSON, And External Compatibility Contracts Are Not Concrete Enough
**Sources:** Elias Vega / contract-migration-guardian; Saoirse Raman / pack-versioning-ecosystem; Lena Driscoll / rollout-sequencing-reviewer; Ibrahim Park / future-capability-architect

**Issue:** Alias removal depends on compatibility mechanisms that are not yet executable: the external support mode is still described as a compatibility branch, the `gc formula validate --all-packs --legacy-contract-report` contract lacks scan scope and stable schema, JSON formula migration is undefined, and old-binary/`bd` compatibility for dual-declared and requires-only formulas is not a release-blocking gate.

**Required change:** Define the external support plan, owner, supported versions, duration or removal criteria, user opt-in, and artifact location. Specify the validation/report JSON schema, exit codes, first-party/external classification, imported/transitive pack behavior, JSON formula policy, and CI or release-gate usage. Add gates proving previous supported `gc` and configured `bd` paths handle dual-declared formulas before first-party source conversion, or block conversion until those paths are removed.

### [Major] Workflow-Root And Graph-Workflow Authority Is Not Pinned
**Sources:** Nadia Sorenson / compiler-boundary-invariant; Yuki Patel / caller-integration-inventory; Elias Vega / contract-migration-guardian; Felix Berger / convergence-subset-reviewer

**Issue:** The design says callers should use typed compiler facts, but it leaves `Compile(...) (*Recipe, error)`, mutable recipe metadata, persisted `gc.formula_contract`, and raw query predicates in place. A broad predicate can over-match default-capability formula roots, while existing SQL/API/order/convoy selectors cannot be migrated by a post-fetch predicate alone.

**Required change:** Name one owner for graph workflow detection and expose both typed predicates and typed query criteria. Add `Recipe.GraphWorkflow` and `CompileResult.GraphWorkflow` or equivalent typed fields, define their relation to normalized requirements and legacy metadata, and make recipe metadata a projection boundary rather than a caller-side compiler API. Hide, remove, or statically forbid production use of compile wrappers where behavior, diagnostics, graph decisions, or metadata stamping depend on compiler facts.

### [Major] Host Capability Semantics Need Fail-Closed Construction And Stable Lifetimes
**Sources:** Nadia Sorenson / compiler-boundary-invariant; Felix Berger / convergence-subset-reviewer; Ibrahim Park / future-capability-architect; Yuki Patel / caller-integration-inventory

**Issue:** `HostCapabilities` represents the active binary boundary, but construction, valid values, source attribution, request lifetime, and per-call recomputation are still loose. Convergence, retry, speculative wisp, and fanout paths do not name how they receive a capability snapshot or accepted compile artifact.

**Required change:** Define valid capability values, reject invalid/future values without panics, and route production construction through compiler/config/controller adapters. State that satisfaction is computed per call from `CompileOptions.HostCapabilities`, not from package-global state or cached formula identity. Name the capability plumbing for sling, orders, fanout, convergence create/retry, speculative wisp, and reconciler ticks.

### [Major] Convergence Needs A Typed Projection And Zero-Write Contract
**Sources:** Felix Berger / convergence-subset-reviewer; Yuki Patel / caller-integration-inventory; Lena Driscoll / rollout-sequencing-reviewer

**Issue:** Convergence still risks preserving its local formula subset parser. The design does not specify a concrete `internal/formula` projection for convergence fields, preflight ordering before durable writes, mid-loop downgrade behavior, or what requirement/provenance/capability data is persisted on convergence roots.

**Required change:** Define `CompileResult.ConvergenceMetadata()` or an equivalent projection covering convergence enablement, required vars, evaluate prompt, relevant step identity, source/key provenance, and diagnostics. Require canonical compile and requirement validation before `CreateConvergenceBead`, retry root creation, metadata writes, missing-child markers, speculative wisps, or first `PourWisp`. Choose either persisted accepted compile artifact reuse or fail-on-downgrade semantics and test it.

### [Major] Pack Provenance And Author Tooling Are Not Yet Usable By External Authors
**Sources:** Saoirse Raman / pack-versioning-ecosystem; Elias Vega / contract-migration-guardian; Avery Brooks / docs-dx-terminology

**Issue:** The design wants pack semver, refs, SHAs, and provenance to be the reproducibility boundary, but current layer resolution is path-oriented and can lose pack binding, requested ref, locked revision, content hash, dirty state, import chain, and layer priority. The validation commands read as internal release gates, not stable external author workflows.

**Required change:** Specify a typed provenance flow from pack/import resolution through formula resolution, `CompileResult.Provenance`, and workflow-root metadata or durable compile artifacts. Add a first-class external validation surface for local pack paths and pinned remote source/ref or SHA, with stable JSON diagnostics, migration hints, reproducibility fields, and tests for imported winners, transitive imports, dirty local packs, lockfile-backed refs, and shadowed formulas.

### [Major] Rollout Phasing Still Hides Flag-Day Work
**Sources:** Lena Driscoll / rollout-sequencing-reviewer; Avery Brooks / docs-dx-terminology; Yuki Patel / caller-integration-inventory

**Issue:** Phase 3 groups sling, orders, API, convoy, convergence, molecule execution, dashboard projections, tests, and the static no-raw-consumer guard. As written, the static guard can only land after all callers migrate, making the plan either a large flag-day PR or an undefined multi-PR sequence. Runnable examples and first-party formulas are also not separated from docs prose and requires-only conversion.

**Required change:** Split caller migration into reversible sub-phases with tests and rollback for shared result plumbing, sling/CLI, orders, API, convoy, convergence/molecule execution, dashboard projections, and finally static guard enforcement. Separate documentation prose, runnable examples, dual-declared first-party formulas, and requires-only first-party formulas into distinct gates. Define the minimum binary floor artifact, CI enforcement, and rollback path before requires-only conversion.

### [Major] Documentation And Generated Help Are Not Tied To Diagnostics
**Sources:** Avery Brooks / docs-dx-terminology; Saoirse Raman / pack-versioning-ecosystem; Elias Vega / contract-migration-guardian

**Issue:** The design can expose `formula.contract_deprecated`, `formula.version_deprecated`, or `formula.compiler_requirement_unsatisfied` before the canonical docs, generated help, tutorials, examples, PackV2 docs, and first-party snippets teach `[requires]` consistently. It also does not name the source-of-truth files and regeneration commands for generated `docs/reference/config.md` and `docs/reference/cli.md`.

**Required change:** Add a release gate that blocks user-visible diagnostics until the documentation and example bundle lands. Replace category-level docs language with a file-level checklist including `docs/reference/formula.md`, generated config/CLI sources, `engdocs/architecture/formulas.md`, tutorials, public examples, first-party workflow formulas, testdata, compatibility fixtures, and PackV2 docs. Add a "Which requirement surface do I use?" comparison for formula `[requires]`, `[pack].requires_gc`, legacy pack `requires`, and `[imports.*].version`.

### [Minor] Diagnostic And Requirement Metadata Details Need Sharpening
**Sources:** Marta Hidalgo / operator-diagnostics-gate; Priya Zielinski / parser-validation-matrix; Ibrahim Park / future-capability-architect; Nadia Sorenson / compiler-boundary-invariant

**Issue:** Several lower-level details remain too loose: diagnostic severity is a free string, invalid syntax and unsupported future capability may share remediation, empty `[requires]` versus omitted provenance is not pinned, `gc.*` metadata namespace ownership is unclear, and `OnceKey` LRU size/eviction/reset behavior is not testable.

**Required change:** Enumerate diagnostic severities, split invalid syntax, unsupported future capability, and unknown axis codes, and define payload conventions for rejected axis and supported axes. Decide whether empty `[requires]` collapses to omitted or preserves authoring provenance. Reserve or narrow `gc.*` metadata triggers, state exact key/value matching, and set numeric bounds plus concurrency ownership for diagnostic suppression caches.

## Disagreements
- Marta Hidalgo and Yuki Patel block; the other eight personas approve with risks. I adopt `block` by worst-verdict-wins and because the two blockers affect runtime behavior, operator noise, and durable-write boundaries, not just documentation polish.
- Several Claude reviews were willing to accept producer behavior if dashboard grouping or compatibility documentation closed the loop, while Codex reviews more often required producer-side contracts and executable gates. I side with requiring named producer behavior and tests because this design is about removing projection drift.
- Reviewers disagreed on whether `bd` fallback is live. That ambiguity is itself a finding: the design must either delete fallback as production behavior or specify and test it as a first-class compatibility path.
- For convergence and in-flight work, reviewers differed between persisted accepted compile artifacts and fail-on-downgrade semantics. Either can be valid, but the design must split existing-root iteration semantics from new formula selection and prove zero-write behavior for rejected current compiles.
- Some reviewers treated `RequirementSource`, source paths, and `gc.formula_requirement_source` as diagnostic-only metadata; others worried they could become new routing inputs. The design should make them diagnostic attribution only and add static-guard coverage against branching on them outside `internal/formula`.
- Multiple personas noted no Gemini artifacts were present. Because the required sources were Claude and Codex and each persona synthesis had both, this is not a workflow failure, but it reduces independent model diversity.

## Missing Evidence
- Cross-surface fixture equality for disabled host capability across direct CLI, API-routed CLI, API responses, dashboard rendering, order events, controller paths, and convergence paths.
- Repeated-fatal cadence tests for background order/controller/daemon loops and a resolved rule for preview/validate Event Bus emissions.
- A generated parser/validation matrix with diagnostic ordering, grammar, raw TOML classification, legacy `contract` lexical behavior, v2 construct registry completeness, expansion/aspect validation, and metadata predicate exactness.
- A live-or-retired decision for `bd` shell-out fallback, plus parity or removal gates for dual-declared and requires-only formulas.
- Typed workflow-root/query criteria and graph-workflow authority, including exact static-guard exemptions for raw metadata selectors that must remain at persistence boundaries.
- Concrete `HostCapabilities` construction, lifetime, source, and per-entry-point threading through sling, orders, fanout, convergence, retry, speculative wisp, and reconciler paths.
- A convergence projection API and zero-durable-write tests before every root, retry, metadata, marker, wisp, and child-bead write.
- External pack-author validation command contract, provenance schema, support plan for SHA-pinned legacy packs, and compatibility branch/LTS/failure policy.
- JSON formula migration policy for legacy `"contract": "graph.v2"`, invalid JSON contract values, host-disabled behavior, and any accepted or rejected JSON `[requires]` equivalent.
- File-level docs migration plan, generated-doc source ownership, regeneration commands, stale-guidance CI corpus, and rendered canonical plus migration examples.
- Minimum binary floor artifact, CI enforcement, rollback procedure for requires-only conversion, and tests for old supported readers.

## Recommended Changes
1. Resolve the two blockers first: define background diagnostic/event emission rules and settle the live/retired status of `bd` fallback with per-entry-point runtime semantics.
2. Pin caller authority: add typed formula provenance, workflow-root, graph-workflow predicates, and typed query criteria; forbid raw metadata interpretation outside named persistence-boundary exemptions.
3. Define host-capability construction and threading, then specify fanout, convergence, retry, and in-flight behavior under capability changes with zero-write tests.
4. Replace prose parser requirements with generated validation fixtures and an encodable v2-only construct registry, including legacy `contract`, JSON, expansion/aspect, and raw TOML edge cases.
5. Make diagnostic projection contractual across CLI, API, dashboard, events, controller, orders, and convergence, including severity registry, remediation ownership, source attribution, OnceKey bounds, and repeat-fatal cadence.
6. Turn pack provenance and external validation into stable user-facing APIs with schemas, exit codes, support policy, and provenance fields that survive layer resolution and imports.
7. Split rollout phases into independently reversible slices and add release gates for old-binary/`bd` compatibility, minimum binary floor, dual-stamping sunset, and requires-only rollback.
8. Gate user-visible diagnostics on docs and examples landing together: canonical formula reference, generated CLI/config docs, tutorials, examples, first-party workflow formulas, PackV2 docs, and stale-guidance CI.
