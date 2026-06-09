# Design Review Synthesis

## Overall Verdict: block

The global verdict is `block` because the operator diagnostics persona raised a blocking defect in the order/background diagnostic grouping contract. The rest of the personas converged on `approve-with-risks`: the design direction is strong, but too many safety properties still depend on unstated store semantics, incomplete caller inventories, and release evidence that is not yet executable.

## Consensus Strengths
- The central compiler-boundary principle is correct: formulas declare minimum requirements, while the active Gas City binary and host capability snapshot decide whether those requirements can be satisfied.
- The typed compile and acceptance model is the right durable-write boundary, especially `NormalizedRequirements`, `HostCapabilities`, `CompileResult`, `AcceptedCompileArtifact`, and compiler-owned projections.
- The legacy `contract = "graph.v2"` alias window, dual-declaration support, and deterministic conflict handling are directionally sound.
- The byte-exact requirement grammar and generated validation-matrix approach are much stronger than hand-picked parser tests.
- The design correctly treats pack revision, resolved ref/SHA, lockfile revision, and content hash as the reproducibility boundary rather than inventing formula-level semver.
- Reviewers agree convergence should stop maintaining a subset parser and should write durable convergence state only after canonical compiler acceptance and typed projection validation.
- The docs/stale-guidance gate concept is useful and should become a release-blocking control before user-visible requirement diagnostics ship.

## Critical Findings

### [Blocker] Order/background diagnostic grouping is not implementable against the current persistence contract
**Sources:** Marta Hidalgo / Claude, Codex
**Issue:** The design relies on idempotent grouped diagnostic state for order/background compile failures, but it does not define a valid durable diagnostic subject or an atomic write mechanism available through the current Task Store API. That leaves repeated scans, concurrent dispatchers, restarts, config flaps, and write failures able to duplicate diagnostics, lose occurrences, or make an order look fired before the operator sees the fatal compile problem.
**Required change:** Define the order diagnostic subject and grouped-state write contract in terms of the actual Task Store API, or explicitly add and test the required store primitive. Move formula compile preflight before `order.fired`, fired metadata, wisp creation, route metadata, and user-visible history writes, then fixture-lock repeated-loop, concurrency, restart, failure, recovery, burst-budget, and config-flap behavior.

### [Major] Compiler authority and durable-write boundaries are still porous
**Sources:** Nadia Sorenson, Yuki Patel, Felix Berger / Claude, Codex
**Issue:** The design has the right typed acceptance model, but it does not yet prove that all production paths use it. Host capability construction lacks one edge adapter, formula metadata interpretation can still split across packages, raw TOML/JSON/root-metadata consumers are not fully guarded, and legacy materialization paths such as `Compile`, `MolCook`, `gc formula cook`, convergence repair, retry, fanout, hooks, convoys, and dashboard/API projections can remain reachable without accepted-artifact validation.
**Required change:** Add the host-capability edge adapter, make `internal/formula` or a formula-owned delegator the only interpreter of formula requirement facts, and add blocking static guards for formula-v2 globals, raw formula decoding, hand-rolled workflow-root queries, and convergence subset parser use. Durable writers must structurally require `AcceptCompileResult` and `ValidateAcceptedArtifact`, or fail closed with typed diagnostics and zero durable writes.

### [Major] Caller and validation matrices are not yet executable enough to enforce the design
**Sources:** Priya Zielinski, Yuki Patel, Felix Berger / Claude, Codex
**Issue:** The caller-path vocabulary is inconsistent with the 18-row preflight lock, durable-writer end-state cross-references are missing, `step.expand` semantics are ambiguous, construct predicates remain prose-driven, count locks and coverage intents are not fully checked, and convergence surfaces such as missing-child repair, speculative wisps, and active legacy-root repair are not explicitly covered.
**Required change:** Make `caller_paths.yaml` the single source for caller ids, generate the caller-to-durable-writer cross-reference, resolve whether `step.expand` itself requires compiler v2, replace prose construct predicates with decoded field locations or contribution traversal rules, and add blocking count-lock, coverage-intent, combined-defect, and convergence-bypass fixtures.

### [Major] Migration, alias removal, and rollout gates need executable release evidence
**Sources:** Elias Vega, Lena Driscoll, Saoirse Raman, Avery Brooks / Claude, Codex
**Issue:** The migration plan depends on first-party inventory, old-reader compatibility, minimum binary floor, external support state, public notice, stale-guidance CI, and alias-removal evidence, but the artifacts and gates are not yet specified tightly enough. Several rollout phases are too broad for safe rollback, and docs/proposal supersession is not pinned to the same phase as user-visible diagnostics.
**Required change:** Add a release clock-anchor gate that cannot start the alias window unless the required artifact bundle exists. Save first-party inventory, compatibility/min-floor/external-support, old Gas City and `bd` probe, stale-guidance, and alias-removal-gate artifacts; split broad phases into owner/rollback units; define the external public-notice/support window and sliding-window alias-drain rule; and land docs/reference/architecture/proposal updates before diagnostics are exposed.

### [Major] Operator projections and warning lifecycle can diverge across surfaces
**Sources:** Marta Hidalgo, Avery Brooks, Saoirse Raman / Claude, Codex
**Issue:** Warning suppression, accepted-warning persistence, `--warnings=once`, config reload, host toggles, LRU eviction, and background alias evidence are not one lifecycle. HTTP status rules also conflict for `formula.compiler_requirement_conflict`, and warning bodies for `formula.contract_deprecated`, `formula.version_deprecated`, and `formula.version_misuse` are not pinned across CLI, API-routed CLI, Huma JSON, generated TypeScript, dashboard, and reports.
**Required change:** Define one warning/group lifecycle and operation-aware status projection table, then add parity fixtures for direct CLI, API launch, API-routed CLI, OpenAPI/generated TypeScript, dashboard state, reports, and background producers.

### [Major] External-pack reproducibility and binary-floor behavior are underspecified
**Sources:** Saoirse Raman, Elias Vega, Lena Driscoll, Ibrahim Park / Claude, Codex
**Issue:** External-pack validation depends on packman provenance that is not yet contracted, including content hash, locked revision, parent binding, dirty state, transitive depth, registry mirror, and `[pack] requires_gc`. The active Gas City version comparator is undefined for release, prerelease/nightly, `main`, dirty, and source builds, and external-support artifacts, migration hints, and pinning recipes remain too informal for alias-removal gates.
**Required change:** Define packman schema 2 or equivalent provenance fields, schema-1 fail-closed diagnostics, `[pack] requires_gc` version-source/comparator behavior, executable external-support schema, external discovery/outreach workflow, `formula.migration.pin_pack_revision` JSON, and SHA-pinned fixture coverage through alias-removal blocking.

### [Major] Future capability and artifact identity semantics are incomplete
**Sources:** Ibrahim Park, Priya Zielinski, Nadia Sorenson / Claude, Codex
**Issue:** Runtime equivalence for omitted `[requires]`, empty `[requires]`, and `formula_compiler = ">=1"` does not settle content hash, compile id, accepted artifact refs, retry/downgrade/convergence reuse, or dedup identity. Future compiler minima and requirement axes also need byte-exact grammar, axis-id validation, multi-axis precedence, and clear `gc.formula_unsupported_axes` semantics.
**Required change:** Add a default-capability artifact-identity rule with fixture coverage, future-minimum rows for several unsupported positive integers, a canonical requirement-axis identifier grammar, multi-axis combined-defect precedence rows, and a durable-or-reader-projected decision for unsupported-axis metadata.

### [Minor] Documentation and terminology gates need sharper scope
**Sources:** Avery Brooks, Priya Zielinski, Elias Vega / Claude, Codex
**Issue:** The docs plan can still let stale examples pass if it only scans explicit TOML tokens. Public reference, tutorials, architecture docs, examples, first-party pack prose, generated help, and proposal docs must teach the same distinction between formula `[requires].formula_compiler`, pack `[pack].requires_gc`, pack revision, and lockfile identity.
**Required change:** Extend stale-guidance and positive-content checks to example Markdown and first-party pack prose, clarify doctest routing for formula versus pack/config TOML snippets, pin warning text, and name one canonical glossary/PackV2 anchor for the terminology.

### [Minor] Persona artifact paths were inconsistent for this attempt
**Sources:** Workflow artifacts
**Issue:** `attempt-79/persona-syntheses/` was empty, while all ten closed persona synthesis beads recorded fresh outputs under `.gc/design-reviews/ga-j195cx/attempt-1/persona-syntheses/`. The timestamps and bead metadata show these were the current persona syntheses, but the path mismatch is workflow artifact drift.
**Required change:** Fix the persona synthesis output-path derivation so future attempts write under `attempt-${gc.attempt}` and stamp matching `design_review.output_path` values.

## Disagreements
- The only verdict disagreement that changes the global result is in the operator diagnostics lane: Claude rated it `approve-with-risks`, while Codex rated it `block`. I agree with the `block` assessment because the missing durable diagnostic subject and atomic/idempotent write mechanism are prerequisites for the promised operator behavior, not merely fixture cleanup.
- Several lanes split on whether missing release artifacts are design blockers or implementation evidence to produce later. I treat them as major risks, not the global blocker, because the rollout can proceed only if the design names the artifacts, owners, and gates before implementation teams encode judgment calls.
- Reviewers disagreed on whether `step.expand` is itself a v2-only construct or only a traversal mechanism for contributed formulas. The design must choose one rule; either is workable if the matrix and construct registry encode it mechanically.
- Claude often emphasized process and release governance, while Codex emphasized executable contracts, static guards, and exact fixtures. These are complementary: the release process needs the executable artifacts, and the static guards need the release process to decide when they become blocking.
- Kimi 2.6 reviews were absent. This matches `skip_gemini=true` / optional-lane behavior for the attempt and does not change the synthesis verdict.

## Missing Evidence
- A concrete Task Store-compatible grouped diagnostic write strategy and durable order diagnostic subject.
- A complete host-capability edge adapter API and blocking guards for formula-v2 globals, raw formula decoding, `RequirementSource` misuse, hand-rolled workflow-root queries, and convergence subset parser use.
- A generated Phase 0 caller manifest seeded from the current tree, with caller-to-durable-writer cross-references and explicit convergence repair/wisp paths.
- Saved first-party formula inventory, old-reader compatibility, `bd` probe, minimum-binary-floor, external-support, stale-guidance, and alias-removal-gate artifacts.
- Packman provenance contract, `[pack] requires_gc` comparator rules, external-support schema, external discovery/outreach process, and SHA-pinned validation fixture.
- Documentation doctest schema and stale-guidance sample report proving stale files are caught without broad false positives.
- Fixture coverage for accepted-warning lifecycle, cross-surface diagnostics, future compiler minima, requirement-axis grammar, multi-axis precedence, artifact identity for default capability spellings, and convergence host-disabled zero-write behavior.

## Recommended Changes
1. Fix the operator diagnostics blocker: define the durable subject and idempotent grouped-state write mechanism, move order compile preflight before all fired/history writes, and add the required background-producer fixtures.
2. Lock compiler authority down with one host-capability edge adapter, formula-owned requirement facts, accepted-artifact-only durable writer APIs, and static guards against raw consumers.
3. Make the caller and validation matrices generated, cross-referenced, and count-locked before exposing user-visible diagnostics.
4. Split the rollout into small owner/rollback units and add the saved release artifacts that gate alias-window start, requires-only conversion, and alias removal.
5. Define external-pack provenance, `requires_gc` comparison, support/outreach schema, migration hints, and SHA-pinned validation fixtures.
6. Pin warning lifecycle and projection parity across CLI, API, generated dashboard types, dashboard UI, reports, and background producers.
7. Fence convergence migration with per-PR bypass tests, no subset parser guard, convergence-specific artifact metadata rules, and host-disabled zero-write fixtures.
8. Resolve future capability identity and axis semantics, including default-capability artifact identity and multi-axis diagnostic precedence.
9. Land docs/reference/architecture/tutorial/example/proposal updates in the same phase as diagnostics, with stale-guidance and doctest gates.
10. Repair the design-review workflow artifact path bug so persona syntheses for an attempt are written and stamped under the current attempt directory.
