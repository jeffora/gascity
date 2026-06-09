# Design Review Synthesis

## Overall Verdict: block

The global verdict is `block` because the caller-integration persona returned `block`, and worst-verdict-wins applies. The design direction is broadly sound: reviewers agree on `[requires]` as the capability declaration surface, compiler-owned accepted artifacts, conservative alias migration, pack-level reproducibility, and convergence moving off a private subset parser. The blocking issue is that several central integration contracts are still prose or sketches rather than compile-safe APIs, executable inventories, and zero-write fixtures.

## Consensus Strengths
- Multiple personas praised the core ownership model: formulas declare requirements, while the active Gas City binary owns requirement normalization, host satisfaction, diagnostics, v2 construct interpretation, and accepted compile artifacts.
- Reviewers agreed that `[requires] formula_compiler` is the right canonical surface, with legacy `contract = "graph.v2"` treated as a compatibility alias during a measured migration window rather than as a future compiler selector.
- The accepted-artifact boundary, resolver-owned provenance, pack revision and lockfile identity, and formula-owned projections were consistently viewed as the right direction for eliminating caller drift.
- The generated validation matrix approach is strong: grammar rows, raw-shape rows, diagnostic golden tests, contribution traversal, construct registry coverage, projection parity, and caller preflight inventories are the right guardrails.
- Documentation and diagnostics are correctly treated as release gates, not follow-up chores, with reviewers repeatedly asking for local commands and report artifacts that maintainers and pack authors can run.
- Convergence retiring its subset parser and consuming compiler-owned artifacts is the correct architecture, provided the projection boundary is made mechanically safe.

## Critical Findings

### [Blocker] Shared caller/source-workflow API is not yet a safe migration boundary
**Sources:** Yuki Patel persona synthesis, supported by Nadia Sorenson and Felix Berger findings.
**Issue:** The proposed source-workflow integration point is not yet compile-safe or behaviorally complete. Reviewers found a sketch where `WorkflowRootFacts` appears as both type and function, missing criteria and kind types, and a raw `beads.ListQuery` abstraction that cannot express the current union of canonical, dual-stamped, legacy, graph-v2-only, closed-history, and source-scoped roots. Current callers already differ on exact, trim, and case-fold semantics, so migrating them without a typed API and parity fixtures risks preserving ad hoc filters or silently dropping roots.
**Required change:** Replace the raw query proposal with a typed operation owned by `internal/sourceworkflow`, such as `ListWorkflowRoots(store beads.Store, criteria WorkflowRootCriteria) ([]WorkflowRootSnapshot, error)`, plus separately named facts and predicate helpers. Add golden parity fixtures for canonical-only, dual-stamped, legacy-only, graph-v2-only, closed-history, source-scoped, whitespace-variant, and case-variant stores, and make the per-occurrence caller manifest a Phase 0 acceptance artifact.

### [Blocker] Convergence projection boundary can recreate subset-parser drift
**Sources:** Felix Berger persona synthesis, with related concerns from Nadia Sorenson and Yuki Patel.
**Issue:** `formula.CompiledConvergenceProjection` and `convergence.ConvergenceMetadata` overlap but diverge on step identity, required vars, retry nilability, runtime vars, requirements, provenance, and artifact refs. That translator can become a new private convergence subset model under another name, especially if convergence can branch on projected requirements or host capability fields.
**Required change:** Collapse `ConvergenceMetadata` into a thin view of the compiler-owned projection, or add a generated and fixture-locked mapping with a field equivalence table in CI. Remove requirements and host capability values from convergence metadata, or make them provenance-only and guard `internal/convergence` against branching on them.

### [Major] Durable producer writes are not fully protected by accepted artifacts
**Sources:** Nadia Sorenson, Yuki Patel, Lena Driscoll, Felix Berger.
**Issue:** The design intends `AcceptedCompileArtifact` to be the durable boundary, but production writers can still appear to call bare recipes, re-read host capability, shell out through runner-based `bd`, or write graph/convergence state before acceptance. Sling preflight and instantiation, graph apply, molecule cook/attach helpers, order dispatch, convergence, fanout, retry, on-complete, repair, and finalize paths all need explicit migration rows.
**Required change:** Add one manifest row per production occurrence with file, current behavior, replacement API, durable-write boundary, owner, expiry phase, rollback control, and blocking test. Add static guards that catch runner-based `bd` shell-outs and zero-write fixtures proving compile failures leave no root, child, hook, convoy, tracking, retry, or metadata writes behind.

### [Major] Parser, raw-shape, and JSON validation contracts are still too deferrable
**Sources:** Priya Zielinski, Ibrahim Park, Elias Vega.
**Issue:** The design defines the right closed grammar, but important outcomes remain unpinned: duplicate scalar keys, `[[requires]]`, nested `[requires.formula_compiler]`, inline tables, arrays, malformed numeric values, future ranges, unknown axes, duplicate JSON members, and JSON loader presence. The raw scanner also lacks a named fuzz, property, or differential testing contract against TOML metadata and JSON token behavior.
**Required change:** Seed literal count locks or equivalent generator self-tests for every normative suite. Pin every TOML and JSON raw-shape outcome to a diagnostic or parser-boundary error, add duplicate JSON-member fatal fixtures, and require a named raw-scanner property or differential test for key presence, raw bytes, decoded type, and source position.

### [Major] Operator diagnostics and background failure state are not executable enough
**Sources:** Marta Hidalgo, Elias Vega, Avery Brooks.
**Issue:** `FormulaDiagnosticGroupState`, warning suppression, event payload ownership, endpoint status semantics, and formula-backed order dispatch sequencing are not yet concrete. In particular, order dispatch currently risks recording fired/tracking state before formula acceptance, while the design requires fatal diagnostics to create no wisp/root/fired state and to update grouped failure state instead.
**Required change:** Define the durable grouped-diagnostic state owner, key, serialized fields, upsert API, atomic event/state update order, restart behavior, cleanup policy, dashboard/API projection, and write-failure semantics. Add a diagnostic-code by endpoint-class table, typed payload helpers usable from non-API packages, and order-dispatch fixtures proving compile failure leaves no fired/tracking/wisp state while repeated scans increment grouped state.

### [Major] Rollout gates and rollback controls are not yet release-operable
**Sources:** Lena Driscoll, Saoirse Raman, Elias Vega, Yuki Patel, Avery Brooks.
**Issue:** The migration avoids a flag day conceptually, but several phase gates are missing a concrete PR home, owner, command, or rollback unit. First-party dual declarations, `packman` schema 2, docs and generated help, alias-removal evidence, active legacy-root repair, pack-floor enforcement rollback, and API/dashboard coupling all need explicit sequencing.
**Required change:** Add a numbered first-party dual-declaration phase before diagnostic visibility, make `packman` lockfile schema 2 a hard prerequisite for gates that rely on content hash or binding identity, define per-phase rollback controls, require an active-root report before blocking legacy roots, and bind the alias-window clock and reset rules to concrete release conditions.

### [Major] Documentation and external-author migration gates need a precise local loop
**Sources:** Avery Brooks, Saoirse Raman, Elias Vega.
**Issue:** Reviewers support the docs-gated rollout, but the proposed stale-guidance matcher is too broad if it matches the literal word `version` across unrelated Go, Node, runtime, pack import, release, and CLI prose. External authors also need a canonical discovery path, worked examples, validate commands, and migration reports before alias removal can be considered safe.
**Required change:** Replace broad prose matching with structured or scoped checks: parse TOML blocks and formula files for formula-context `version =`, then use prose matchers only in known formula-reference documents. Define one local docs quality gate covering doctests, stale-guidance scans, first-party inventory, generated help/schema, examples, tutorials, and PackV2 author docs.

### [Major] Future compatibility metadata can become behavioral authority
**Sources:** Ibrahim Park, Nadia Sorenson, Felix Berger.
**Issue:** Requirement-source fields, normalized requirements, future axis manifests, host capability values, and provenance fields are observable across persisted metadata, projections, dashboard state, generated TypeScript, reports, alerts, and external tools. Without explicit invariants and guards, consumers can branch on provenance and recreate the same decision drift the design is trying to remove.
**Required change:** State that released construct requirements and requirement-axis byte grammars are immutable, richer syntax uses new axes or schema versions, and requirement-source fields are diagnostic/provenance-only on every surface. Add fixtures proving omitted `[requires]`, empty `[requires]`, and explicit `formula_compiler = ">=1"` have identical normalized identity except allowed provenance, and define canonical axis manifest encoding.

### [Minor] Several edge policies remain underspecified
**Sources:** Priya Zielinski, Marta Hidalgo, Saoirse Raman, Ibrahim Park, Avery Brooks.
**Issue:** Lower-than-supported values such as `>=0`, overflow behavior, deprecated `graph_workflows` promotion, offline `--pack-source` host capability context, registry-pinned reproducibility, placeholder TOML snippets, and post-alias-removal glossary treatment are smaller but still visible to operators and pack authors.
**Required change:** Add explicit fixtures or docs decisions for these edge policies before release gates become blocking, so authors see deterministic diagnostics and reviewers are not forced to infer intent from implementation details.

## Disagreements
- Yuki Patel had a verdict disagreement: Claude returned `approve-with-risks`, while Codex returned `block`. I adopt `block` because the source-workflow API issue is both compile-level and architectural, and the design cannot safely migrate callers without that boundary.
- Saoirse Raman had a verdict disagreement: Claude returned `approve-with-risks`, while Codex returned `approve`. I adopt `approve-with-risks` for that persona because the pack revision model is sound, but the packman schema dependency and alias-removal evidence are release-critical.
- Several personas differed on timing rather than facts. Codex often accepted design direction while Claude demanded concrete fixtures or phase gates earlier. My assessment is that implementation may remain incremental, but the design must name the executable artifacts that make each phase safe.
- Reviewers disagreed on whether some diagnostics and warning reach issues are major or minor. I treat headless legacy-alias visibility, background grouped state, and high-volume validation cadence as major because they affect operator upgrade safety.
- Kimi 2.6 artifacts were absent across the persona syntheses. The workflow allowed that lane to be skipped, so this is not a workflow failure, but it reduces independent model coverage.

## Missing Evidence
- Compile-safe `internal/sourceworkflow` API signatures, typed criteria and snapshot types, and parity fixtures for all root shapes and current caller semantics.
- A per-occurrence caller migration manifest covering sling, graph apply, workflow-root predicates, convergence, fanout, retry, finalize, repair, recipe writes, and runner-based `bd` shell-outs.
- Byte-exact `acceptedCompileProof.identityHash` fields, canonical ordering, encoding, and hash function.
- Structural guards for `HostCapabilities` construction, typed `SourceKind`, source-kind fixtures, and expiry for legacy global capability accessors.
- Literal count locks or generator self-tests for grammar, raw-shape, legacy-alias, construct-registry, contribution-traversal, projection-parity, and caller-preflight suites.
- Raw-scanner fuzz/property/differential tests for TOML and JSON source attribution, plus duplicate JSON member fatal behavior.
- Durable `FormulaDiagnosticGroupState` persistence contract, grouped-state API, order-dispatch failure sequence, and repeated-scan fixtures.
- Typed diagnostic payload ownership and event recording helpers for formula, order, and convergence compile failures.
- Diagnostic-code by endpoint-class status table covering CLI, API-routed CLI, Huma, dashboard, generated clients, preview, validation, launch, and background events.
- Named local docs quality gate, sample stale-guidance report, doctest fixture, generated help examples, and first docs PR inventory.
- `packman` lockfile schema 2 owner or design, worked external-consumer migration example, migration hint fixture paths, registry-pinned reproducibility class, and alias-window failure/reset semantics.
- Active legacy-root pre-4f report, `gc formula repair-root-artifact` ordering gate, and zero-write fixtures for all migrated durable producers.
- Convergence projection field-equivalence table, required-var satisfaction carrier, artifact reuse identity table, artifact-ref conflict behavior, and pre-create event grouping/suppression rule.
- Default-capability equivalence fixtures for omitted `[requires]`, empty `[requires]`, and explicit `formula_compiler = ">=1"`.

## Recommended Changes
1. Replace the source-workflow raw query sketch with a compile-safe typed API and make its parity fixture suite a blocking Phase 0 artifact.
2. Collapse or fixture-lock the convergence projection/metadata translator, and forbid convergence from branching on requirement or host capability values.
3. Convert the caller migration plan into a per-occurrence manifest with durable-write boundaries, rollback controls, owners, expiry phases, and zero-write tests.
4. Pin the parser and raw-shape matrix now: duplicate keys, TOML table shapes, malformed numbers, JSON duplicate members, unknown axes, and future capability cross-products.
5. Define the accepted-artifact identity and proof hash byte-for-byte, including whether normalized requirements or source spelling participates.
6. Specify durable grouped diagnostic state and formula-backed order dispatch sequencing before any controller, order, dashboard, or API path depends on it.
7. Add the diagnostic-code by endpoint-class table and typed event payload helpers so CLI, API, dashboard, generated clients, order dispatch, and convergence cannot drift.
8. Add a numbered first-party dual-declaration phase and require active legacy-root repair/reporting before migrated paths block or mutate existing graph roots.
9. Make `packman` lockfile schema 2 an explicit prerequisite for pack-floor enforcement, provenance-dependent release gates, and alias-removal checks.
10. Replace broad stale-guidance matching with scoped formula-context checks and publish one local docs quality gate that authors can run.
11. Define alias-removal evidence sources, headless operator reports, unreachable external-pack policy, and release-clock reset semantics.
12. Freeze future-compatibility invariants: released construct requirements are immutable, requirement axes have canonical byte grammar, and source/provenance fields are never behavioral authority.
