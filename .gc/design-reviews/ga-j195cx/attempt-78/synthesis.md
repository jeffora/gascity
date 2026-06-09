# Design Review Synthesis

## Overall Verdict: block

One persona returned `block`, so the global verdict is `block` under worst-verdict-wins. The design has a strong core direction: requirements are normalized inside `internal/formula`, the active binary remains the compiler authority, pack revision stays the ecosystem identity, and the rollout uses typed artifacts plus generated checks. The blocking gap is that operator-facing diagnostics depend on host-capability provenance the current constructor cannot carry, and several major migration, caller, convergence, provenance, and docs gates need to be pinned before implementation can safely proceed.

## Consensus Strengths
- Multiple personas praised the compiler boundary: formulas declare requirements, while the active Gas City binary decides what capabilities it can satisfy.
- The typed compile and accept model gives callers `CompileResult`, accepted artifacts, and workflow-root facts instead of letting projections interpret raw formula text.
- The closed `formula_compiler` grammar and matrix-driven parser validation strategy are the right shape for avoiding best-effort parsing drift.
- The design correctly keeps pack refs, lockfile revisions, content hashes, and accepted artifact provenance as the ecosystem identity rather than creating formula-level semver.
- The docs and rollout plan include useful gates: doctests, stale-guidance scans, generated help/schema/API/TS refreshes, first-party inventory reports, old-reader probes, and release artifacts.
- Convergence's intended destination is sound: retire the subset parser and consume accepted compiler artifacts plus typed convergence projection validation.

## Critical Findings

### [Blocker] Host capability provenance cannot support deterministic operator diagnostics
**Sources:** Marta Hidalgo; Nadia Sorenson
**Issue:** The design still centers a constructor shaped like `HostCapabilitiesFromFormulaV2(enabled bool, source string)`. That cannot deterministically carry source kind, path/key, raw value, position, config generation, omitted-default versus explicit false/true, deprecated `graph_workflows` promotion, or test override provenance. As a result, CLI, Huma JSON, API-routed CLI, dashboard state, and Event Bus payloads can diverge on the same disabled-host or downgraded-host condition.
**Required change:** Replace or extend the host-capability edge input with a structured type that preserves normalized capability plus diagnostic provenance. Add golden fixtures for omitted default, explicit false, explicit true, deprecated alias promotion, test override, config reload, and host downgrade with same-identity artifact reuse across all operator-visible surfaces.

### [Major] Durable write callers can still bypass normalized compile and accepted artifacts
**Sources:** Yuki Patel; Felix Berger; Nadia Sorenson
**Issue:** Legacy `Store.MolCook`/`MolCookOn`, bd-backed materialization, exec-store `mol-cook`, convergence subset parsing, fanout, order dispatch, API sling, and dashboard/generated-client inference are not all fenced behind `CompileWithResult`, `AcceptCompileResult`, and accepted-artifact validation. During migration, requires-only formulas or v2-required formulas could still authorize durable writes through old paths.
**Required change:** Add a transitional no-bypass gate before any durable writer accepts requires-only graph formulas. Retire or statically guard `MolCook` helpers, require shadow compile before convergence writes, block raw metadata and TOML parsing outside approved owners, and prove zero durable writes for rejected formulas across CLI, API, order dispatch, convergence, and fanout.

### [Major] The migration compatibility contract is not fully executable
**Sources:** Elias Vega; Priya Zielinski; Lena Driscoll; Saoirse Raman
**Issue:** The alias window for `contract = "graph.v2"` is directionally preserved, but post-removal parser behavior, noncanonical legacy spellings, dual declarations, empty `[requires]`, omitted `[requires]`, explicit `formula_compiler = ">=1"`, first-party CI symmetry, and external pinned-pack handling are not all fixture-locked. This leaves room for abrupt external breakage or inconsistent diagnostics.
**Required change:** Add compatibility-matrix rows and old-reader/probe corpus coverage for exact legacy aliases, whitespace/case variants, invalid strings, omitted and empty `[requires]`, default compiler requirements, and agreeing or conflicting dual declarations. Add a first-party guard requiring dual declarations during the alias window and a typed external migration report with concrete support-expiration evidence.

### [Major] Packman provenance is a hidden prerequisite for several rollout gates
**Sources:** Lena Driscoll; Saoirse Raman; Nadia Sorenson
**Issue:** Imported-pack floor enforcement, external pinned-pack support expiration, alias-removal evidence, requirement-diff reports, and release provenance all depend on packman schema 2 or an equivalent provenance contract, but the design does not make that a first-class phase with owner, PR home, saved artifacts, rollback unit, and schema-1 fail-closed behavior.
**Required change:** Promote packman provenance readiness to an explicit prerequisite phase or split Phase 7 into provenance-contract and pack-floor-enforcement phases. Define required lockfile fields, schema migration, schema-1 diagnostics, cache/lockfile rollback behavior, and the artifacts that block alias removal.

### [Major] Parser, registry, and diagnostic matrices have unverifiable coverage counts
**Sources:** Priya Zielinski; Ibrahim Park; Marta Hidalgo
**Issue:** The matrix strategy is strong, but several counts and edges are not derivable from the design text. Missing rows include top-level `requires` as invalid non-table shapes, duplicate scalar keys, array-of-tables, combined-defect diagnostic counts, `caller-preflight count_lock: 18`, JSON loader ownership, construct-registry identity mapping, registry-minimum parsing, and multi-axis diagnostic ordering.
**Required change:** Enumerate the caller-preflight rows or derivation, add raw-shape and combined-defect matrix rows with golden diagnostic counts and ordering, name or remove JSON formula support, cross-walk v2 constructs to registry identities, and parse every axis minimum through its owning axis grammar.

### [Major] Convergence artifact and runtime projection contracts remain ambiguous
**Sources:** Felix Berger; Yuki Patel; Ibrahim Park
**Issue:** Convergence needs a sharper typed contract before it can safely leave the subset parser. Current evaluate-prompt content-marker checks are not mapped to a diagnostic, runtime owner, projection field, or retirement decision. `ConvergenceRuntimeInputs` is underspecified as a `map[string]string`, and artifact keys such as `gc.formula_compile_artifact` versus `gc.convergence_formula_compile_artifact` do not have one canonical precedence, migration, conflict, and repair rule.
**Required change:** Define `ConvergenceRuntimeInputs` with typed satisfaction evidence, defaults, duplicate handling, value source, redaction, and hashes. Specify one convergence artifact-ref contract with canonical key, aliases, dual-stamp migration, conflict diagnostics, repair behavior, and zero-write fixtures. Add transition fixtures proving convergence cannot write from the legacy parser when shared compile/accept would reject.

### [Major] Warning persistence, cadence, and alias-removal evidence are underspecified
**Sources:** Marta Hidalgo; Elias Vega; Saoirse Raman
**Issue:** Deprecation and accepted-warning behavior is split across prose without one durable key shape, write order, grouping lifecycle, LRU eviction contract, dashboard/reporting projection, or alias-removal report. Repeated scans, config reloads, polling producers, restarts, and cross-process CLI commands could either flood logs or lose evidence needed for release gates.
**Required change:** Define one accepted-warning persistence path and grouping key, including first occurrence, repeat scans, config generation changes, host toggles, producer policy, restart behavior, LRU eviction, dashboard projection, and release-report counters. Fixture-lock `formula.contract_deprecated`, `formula.version_deprecated`, and `formula.version_misuse` across direct CLI, API-routed CLI, API, dashboard, and background producers.

### [Major] Docs and proposal alignment are not pinned tightly enough to user-visible diagnostics
**Sources:** Avery Brooks; Lena Driscoll; Saoirse Raman
**Issue:** Context docs still teach pre-migration shapes, including formula `version` and no canonical `[requires]`. The superseded migration proposal is not explicitly tied to the Phase 2 docs bundle, stale-guidance checks may miss omission-only docs, and JSON formula loader status could fork the docs and fixture story.
**Required change:** Move proposal supersession into the same phase that first exposes requirement diagnostics, add positive-content stale-guidance checks for docs that mention `formula_v2`, add a common-confusion section distinguishing formula `[requires]`, `[pack].requires_gc`, and lockfile revisions, and decide whether JSON formulas get Phase 2 examples/fixtures or are retired first.

### [Minor] Several secondary boundaries need explicit closure
**Sources:** Nadia Sorenson; Elias Vega; Avery Brooks; Ibrahim Park
**Issue:** `Recipe.GraphWorkflow` can become a second behavioral authority, interactive CLI warnings may need a quiet/once mode for CI, future non-`>=` operators need a reserved-or-invalid rule, and future compiler axes need a clear scalar-only justification or typed-shape envelope.
**Required change:** Make `Recipe.GraphWorkflow` compiler-owned or validated through accepted artifacts, document CLI warning controls or intentional noisiness, freeze the operator policy, and specify future-axis value shapes and mixed unknown-axis diagnostics.

## Disagreements
- Marta Hidalgo was the only persona to return `block`; the rest returned `approve-with-risks`. The block is accepted globally because deterministic operator diagnostics are a core release contract, not an implementation cleanup detail.
- Marta's Claude review returned `approve-with-risks` while Codex returned `block`. The synthesis sides with the stricter assessment because the boolean host-capability constructor cannot carry the structured provenance promised by the diagnostic design.
- Priya Zielinski, Saoirse Raman, and Avery Brooks had Claude/Codex splits between `approve-with-risks` and `approve`. The synthesis keeps the risk verdicts because their gaps are about executable gates, compatibility evidence, and docs accuracy rather than subjective polish.
- Some reviewers treated missing generated fixtures as implementation-time evidence; others required design-level row lists and counts. The synthesis requires enough design-level enumeration for implementers and reviewers to verify the gates before trusting generated artifacts.
- Reviewers differed on whether current docs, packman provenance, and external outreach should block the design. The synthesis treats them as major required changes because alias removal and external compatibility depend on those artifacts being measurable.

## Missing Evidence
- No Kimi 2.6 artifacts were present for this run; that matches the skipped-provider configuration but reduces independent review diversity.
- Generated matrix artifacts, literal count-lock outputs, generated Go tests, and golden diagnostic files are not present, so parser, caller-preflight, projection, and warning coverage counts cannot yet be audited.
- No concrete old-reader/probe artifact proves supported older Gas City binaries can read the exact Phase 3 first-party pack shape with `[pack].requires_gc` plus dual-declared graph formulas.
- No packman schema 2 tracking reference, owner, PR home, saved artifact set, or schema-1 diagnostic contract is linked to the gates that depend on provenance.
- No checked-in caller manifest proves all `MolCook`, bd-backed, exec-store, convergence, fanout, order, API, and dashboard paths are behind typed compile/acceptance.
- No convergence artifact-reference table or typed runtime-input definition pins convergence's durable-state behavior.
- No external-pack discovery and outreach process shows how supported external packs are identified, contacted, migrated, or declared expired.
- No actual docs-check, stale-guidance, doctest, generated-help, generated-schema, OpenAPI, generated TypeScript, or release-report outputs are available for the rewritten user-facing docs.

## Recommended Changes
1. Replace the host-capability constructor with a structured provenance type and add cross-surface diagnostics fixtures; this is the global blocker.
2. Add the transitional no-bypass gate for all durable writers, including `MolCook` retirement/guards, convergence shadow compile, and frontend/API raw-metadata prevention.
3. Promote packman provenance to a first-class rollout prerequisite with owner, phase, artifacts, schema migration, schema-1 diagnostics, and rollback behavior.
4. Complete the compatibility matrix and old-reader/probe corpus for legacy `contract`, `[requires]`, default compiler requirements, dual declarations, and external pinned packs.
5. Define convergence's typed runtime inputs, artifact key precedence, conflict diagnostics, repair behavior, and host-capability lifetime rules.
6. Make warning persistence and deprecation cadence durable and measurable, including dashboard/reporting evidence for alias-removal gates.
7. Pin docs/proposal supersession, stale-guidance positive checks, JSON loader disposition, and generated docs/schema/API/TS refreshes before any user-visible requirement diagnostic ships.
8. Add future-axis and parser registry self-tests for typed minima, mixed-axis diagnostics, default-capability artifact identity, and non-`>=` operator policy.
