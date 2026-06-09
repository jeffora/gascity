# Design Review Synthesis

## Overall Verdict: block

The design has a strong direction: requirements are moving toward a compiler-owned, typed, source-attributed contract, and multiple reviewers praised the intent to use accepted artifacts, typed diagnostics, compatibility inventories, and fixture-driven rollout gates. The review still blocks because several safety contracts are not yet normative or mechanically testable, especially alias removal, raw-consumer enforcement, validation coverage, and old-reader handling for future requirement axes. The blockers are design-contract gaps, not implementation polish.

## Consensus Strengths
- Multiple personas praised the central move from raw `contract` interpretation toward normalized formula compiler requirements owned by `internal/formula`.
- Reviewers broadly supported `CompileResult` / `AcceptedCompileArtifact` as the right boundary for durable writers, projections, diagnostics, and convergence.
- The design's compatibility posture is directionally sound: dual declaration, legacy inventory, release-floor evidence, old-reader fixtures, and stale-doc checks are the right classes of safeguards.
- Typed diagnostics across CLI, API, dashboard, Event Bus, and controller paths were repeatedly called out as the correct operator model.
- Reviewers agreed that documentation, examples, generated help/schema, and stale-guidance checks must be part of the migration, not an afterthought.
- The PackV2 reproducibility boundary is pointed in the right direction by removing formula-level artifact semver and relying on pack revisions, lockfiles, and immutable content identity.

## Critical Findings

### [Blocker] Alias Removal And Background Legacy Use Are Not Release-Safe
**Sources:** Elias Vega; reinforced by Marta Hidalgo, Lena Driscoll, Avery Brooks
**Issue:** Parser alias removal is not tied to enough observable evidence. First-party `contract = "graph.v2"` declarations, dual-declared formulas, externally supported pinned packs, and background legacy compiles can remain active without a fail-closed release gate or operator-visible migration signal.
**Required change:** Make alias removal require zero first-party `legacy_only` and zero first-party `dual_declared` sources, no active external-support row that still needs parser alias support, concrete old-reader versions or SHAs, and a stated visibility policy for accepted legacy-alias compiles across order dispatch, convergence, fanout, retry, controller, API, dashboard, and validation surfaces.

### [Blocker] Validation Matrix Cannot Prove Its Coverage Contract
**Sources:** Priya Zielinski; reinforced by Ibrahim Park, Marta Hidalgo, Felix Berger
**Issue:** The validation plan relies on prose promises that are not represented in machine-checked fixtures. Workflow-control `gc.*` metadata, `gc.kind` values, inheritance/contribution conflicts, source-paired diagnostics, and byte-exact requirement grammar boundaries can drift without breaking CI.
**Required change:** Add a centralized workflow-control metadata registry and a machine-readable matrix schema with suites, row kinds, coverage intent, count locks, dimension ownership, and zero-write assertions. Include fixtures for metadata keys, parent/child and expansion/aspect/import contributions, malformed and future requirement values, legacy contract edge cases, and generated-control metadata.

### [Blocker] Raw-Consumer Guard And Caller Migration Rollout Conflict
**Sources:** Yuki Patel; reinforced by Nadia Sorenson and Lena Driscoll
**Issue:** The design contradicts itself on when `TestNoNewFormulaRawConsumers` becomes blocking. One section requires it before durable producers migrate, while the rollout leaves it report-only until the final caller phase, allowing new raw consumers to enter during the migration.
**Required change:** State that phase 3a blocks new production raw consumers once the canonical compile result and workflow-root predicate exist. Existing consumers may remain only on an owned, expiring allowlist with replacement tests; later phases must shrink that allowlist. Extend the caller inventory to retry, control, Ralph continuation, fanout, convergence, `molecule.Instantiate`, fragment instantiation, API globals, and legacy graph-contract helpers.

### [Blocker] Future Requirement Axes Lack Durable Old-Reader Protection
**Sources:** Ibrahim Park; reinforced by Nadia Sorenson and Felix Berger
**Issue:** Source parsing fails closed for unknown future requirement axes, but persisted workflow roots created by a newer binary do not carry a v0-readable contract that lets older binaries identify unsupported future semantics before performing graph-specific writes.
**Required change:** Add a durable requirements contract for workflow roots, such as an axis manifest, requirements schema version, minimum-reader capability, accepted-artifact version, or equivalent typed signal. Extend `WorkflowRootFacts` so unknown axes or unsupported artifact/schema requirements classify as future-capability roots where observation and safe cleanup may proceed but retry, fanout, continuation, missing-child repair, child creation, and new compiles fail closed.

### [Major] Compiler Boundary And Graph Workflow Semantics Are Still Ambiguous
**Sources:** Nadia Sorenson; reinforced by Yuki Patel, Felix Berger
**Issue:** The design does not yet pin whether bare `requires.formula_compiler = ">=2"` declares graph topology or only host/provenance capability, and `Recipe.GraphWorkflow`, `internal/sourceworkflow`, convergence projections, and generic writer APIs can still become alternate authorities.
**Required change:** Decide the graph-workflow semantic rule and expose it only through compiler-owned accepted results. Define `CompiledStep` and `CompiledRuntimeVar`, add compiler-owned workflow-root metadata projection/classification, and require every durable writer to accept only an accepted artifact or wrapper with forged, stale, mismatched, host-disabled, and fatal-diagnostic zero-write tests.

### [Major] Operator Diagnostic Projection Is Not Fully Pinned
**Sources:** Marta Hidalgo; reinforced by Elias Vega and Lena Driscoll
**Issue:** Warning and fatal diagnostics have inconsistent grouping keys, incomplete source attribution, unclear Event Bus behavior for `formula.contract_deprecated`, and projection mismatches across CLI, Huma JSON, dashboard, and order events.
**Required change:** Define one canonical warning/fatal key policy, preserve both formula-requirement and host-gate attribution for unsatisfied requirements, decide whether background deprecation warnings emit events, and add golden fixtures for CLI, API, dashboard, order events, occurrence counts, remediation text, and LRU eviction behavior.

### [Major] Rollout, Documentation, And Rollback Sequencing Are Not Coupled
**Sources:** Lena Driscoll and Avery Brooks; reinforced by Nadia Sorenson and Saoirse Raman
**Issue:** Caller phases can expose new diagnostics before docs, tutorials, generated help/schema, stale-guidance checks, and PackV2 author docs land. Rollback language still implies production fallback to the old `bd` shell-out path or `GC_NATIVE_FORMULA=false`, conflicting with the new design.
**Required change:** Add a phase-by-phase diagnostics visibility matrix and make the docs/examples/generated-help bundle a predecessor for each user-visible diagnostic surface. Update or supersede stale live docs, especially `engdocs/architecture/formulas.md` and `engdocs/proposals/formula-migration.md`, before parser or diagnostic changes become user-visible.

### [Major] Convergence Projection And Legacy Root Policy Need A Contract
**Sources:** Felix Berger
**Issue:** The convergence path needs a named owner for decoding and projecting convergence-specific fields, an explicit `ValidateProjection` diagnostic list, and a policy for active legacy roots that lack accepted artifact metadata.
**Required change:** Name the package/file that owns convergence field projection, enumerate validation checks and diagnostic codes, add preflight zero-write rows for create/retry/next iteration/missing-child repair/speculative wisps, and define how legacy active roots behave during retries, host downgrades, and migration.

### [Major] PackV2 Lockfile And External Author Contracts Are Underspecified
**Sources:** Saoirse Raman; reinforced by Elias Vega and Avery Brooks
**Issue:** `LockedRevision`, `--requirement-diff`, `--pack-source --ref`, mutable refs, local paths, transitive imports, and `safe_automatic_edit` are referenced as safety mechanisms without an owning artifact, schema, CLI input contract, or documented author workflow.
**Required change:** Define the lockfile artifact and immutable revision semantics in PackV2 terms, classify branch/tag refs without lock entries as non-reproducible, specify validation CLI behavior for remote/local/direct/transitive inputs, and either define an autofix/apply contract for `safe_automatic_edit` or mark it advisory only.

### [Minor] Several Edge Semantics Need Fixture-Level Decisions
**Sources:** Nadia Sorenson, Priya Zielinski, Marta Hidalgo, Ibrahim Park, Saoirse Raman
**Issue:** `>=1` semantics, integer boundary values, Unicode and NUL strings, test-only constructors, glossary `schema` wording, dashboard remediation fields, transitive ref collisions, and `safe_automatic_edit` behavior are not individually blocking once the major contracts are fixed, but they are likely regression points.
**Required change:** Convert these into explicit matrix rows, docs glossary rows, fixture cases, or advisory-field decisions before implementation begins.

## Disagreements
- Several persona verdicts differed by severity, not by facts. Elias, Priya, Yuki, and Ibrahim each had at least one Claude/Codex split where one model chose `block` and the other chose `approve-with-risks`; this synthesis adopts `block` where the disputed issue affects release safety, durable writes, or old-reader compatibility.
- Reviewers disagreed on whether `internal/sourceworkflow` is a healthy containment layer or a drift risk. My assessment: it is acceptable only if it structurally delegates metadata meaning to compiler-owned classifiers instead of keeping independent graph predicates.
- Reviewers disagreed on whether convergence-specific facts should live directly in canonical compile output or in a typed projection over canonical parser facts. My assessment: either can pass, but the design must choose one owner and forbid convergence/cmd paths from reading raw TOML as a parallel subset parser.
- Reviewers differed on warning transport: some preferred runtime events for background legacy compiles, while others allowed report-only visibility. My assessment: either policy can pass only if it covers every background producer path and is explicit enough for CLI/API/dashboard/Event Bus fixtures.
- Documentation reviewers differed on whether archive/superseded status is required for old proposal docs. My assessment: archival status is optional, but live docs must stop teaching rollback and pre-requirements behavior as current policy before user-visible diagnostics ship.

## Missing Evidence
- Gemini was absent for every persona in this attempt, consistent with `skip_gemini=true`, but leaving no third model to arbitrate edge cases.
- No v0-readable durable metadata contract proves older binaries can detect persisted roots with unknown future requirement axes.
- No complete raw-consumer manifest, allowlist, owner/expiry table, and blocking CI plan covers every durable writer and projection caller.
- No exact old-reader versions, tags, SHAs, or release-floor artifact proves first-party requires-only conversion and parser alias removal are safe.
- No path-keyed docs bundle checklist covers reference docs, tutorials, generated config/CLI/schema docs, PackV2 author docs, live engdocs, examples, fixtures, and first-party formulas.
- No machine-checked validation matrix schema demonstrates suite/count locks, workflow-control metadata registration, inheritance/contribution conflicts, and zero-write assertions.
- No diagnostic projection fixture shows the same `formula.compiler_requirement_unsatisfied` case across CLI stderr, Huma JSON, dashboard, order event, occurrence grouping, source attribution, and remediation.
- No convergence corpus shows create/retry/next-iteration/missing-child behavior for legacy active roots, host downgrades, missing accepted artifacts, or projection validation failures.
- No PackV2 lockfile schema, discovery rule, immutable revision mapping, transitive import behavior, or stable external-author validation example is specified.

## Recommended Changes
1. Resolve the raw-consumer rollout contradiction: make new raw consumers blocking from phase 3a, seed an owned expiring allowlist for existing ones, and inventory every durable writer/projection caller.
2. Add the durable old-reader requirements contract for future axes and the workflow-root lifecycle rules that fail closed for graph-specific writes.
3. Tighten alias removal into fail-closed release gates covering first-party legacy/dual declarations, external support evidence, release-floor artifacts, and background legacy-use visibility.
4. Replace prose validation promises with a checked metadata registry and executable matrix schema that includes zero-write assertions and fixture count locks.
5. Pin compiler-owned graph-workflow semantics, accepted-artifact reuse rules, workflow-root metadata projection, and durable writer signatures.
6. Normalize diagnostic projection keys, source attribution, warning/event transport, remediation text, dashboard fields, and occurrence-count behavior with golden fixtures.
7. Rework rollout sequencing so docs, examples, generated help/schema, stale-guidance scans, and rollback-language cleanup land before any matching user-visible diagnostic surface.
8. Define convergence projection ownership, validation diagnostics, preflight no-write coverage, and legacy active-root migration behavior.
9. Define PackV2 lockfile / `LockedRevision` semantics and external-author validation CLI contracts, including mutable refs, local paths, transitive imports, and `safe_automatic_edit`.
