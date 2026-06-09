# Design Review Synthesis

## Overall Verdict: block

The review remains blocked. Five of ten persona syntheses returned `block`, and the blocking findings converge on enforceability gaps: accepted compile state can still be bypassed or forged, caller migration and validation coverage are incomplete, convergence lacks an implementable canonical projection contract, and pack/version provenance is not durable enough to support the proposed migration. The design direction is broadly praised, but the current document still leaves too much behavior to caller discipline and post-hoc checklist enforcement.

## Consensus Strengths
- Multiple personas praised the move toward normalized requirements owned by `internal/formula`, with `contract = "graph.v2"` treated as a migration alias rather than a second long-term decision surface.
- Reviewers consistently approved the intent to use typed diagnostics, source attribution, host-capability checks, and zero-write preflight before durable workflow creation.
- The alias-window concept, first-party/external inventory, and minimum-floor direction are viewed as the right migration shape, provided they become executable gates rather than prose.
- The fail-closed posture for unknown requirement axes, byte-exact requirement grammar, and monotonic compiler capability model was broadly seen as a sound future-compatibility base.
- Reviewers supported retiring parallel formula parsers and raw predicate drift, especially for convergence and source-workflow classification, if the canonical compile output exposes enough typed data.

## Critical Findings

### [Blocker] Durable compile acceptance is still forgeable or bypassable
**Sources:** Nadia Sorenson/Claude, Nadia Sorenson/Codex, Yuki Patel/Codex, Felix Berger/Claude, Felix Berger/Codex

**Issue:** The design still allows production paths to depend on caller discipline around `CompileResult`, `Recipe`, raw metadata, or caller-constructed normalized requirements. Durable writers for roots, wisps, attached molecules, fanout fragments, retries, convergence iterations, and API/CLI launch paths need a compiler-minted acceptance proof; otherwise callers can skip fatal diagnostics, disabled host-capability checks, or provenance identity validation.

**Required change:** Define a non-forgeable accepted compile artifact minted only after normalization, requirement satisfaction, fatal-diagnostic filtering, host capability validation, and identity/provenance capture. Require every durable producer API to accept that artifact or an entry-point-specific struct embedding it; keep bare `CompileResult` for preview/inspection only.

### [Blocker] Caller migration inventory and parity gates are incomplete
**Sources:** Yuki Patel/Claude, Yuki Patel/Codex, Lena Driscoll/Claude, Lena Driscoll/Codex, Elias Vega/Claude

**Issue:** The migration table misses durable `gc formula` commands such as `show`, `cook`, and `cook --attach`, duplicate API workflow-root predicates, fanout preflight ordering, and convergence retry/iteration paths. The design also references `GC_NATIVE_FORMULA=false gc formula validate --compat-corpus` as a release gate even though that runtime toggle is not shown to exist as a production Go implementation.

**Required change:** Add a current, grep-derived caller/raw-consumer manifest that names every compile, preview, durable write, source-workflow reader, metadata predicate, and legacy helper with an owner and test. Either replace the nonexistent `GC_NATIVE_FORMULA=false` gate with a real, owned compatibility probe and corpus or explicitly remove that rollback story and state the native-only production precondition.

### [Blocker] Validation matrix has concrete bypasses and insufficient dimensions
**Sources:** Priya Zielinski/Claude, Priya Zielinski/Codex, Nadia Sorenson/Claude, Ibrahim Park/Claude, Ibrahim Park/Codex

**Issue:** The proposed matrix can claim completeness while missing existing construct contribution paths such as `step.expand`, `expand_vars`, inline expansion inside children or loop bodies, compose expansion, aspect contributions, transitive imports, and durable caller paths. Requirement grammar edge cases, duplicate/invalid TOML shapes, workflow-control metadata matching, JSON parity, multi-defect ordering, and future capability values are not pinned tightly enough.

**Required change:** Split construct identity, construct location, contribution path, caller path, and requirement shape into independently enumerable matrix dimensions with explicit unsupported/impossible rows. Add fail-closed generator self-tests, JSON-pointer parity, ordered diagnostic goldens, and rows for inline expansion, aspects, transitive imports, malformed/duplicate `[requires]`, future/invalid `>=N` values, and legacy `contract` combinations.

### [Blocker] Convergence cannot safely migrate without a typed canonical projection
**Sources:** Felix Berger/Claude, Felix Berger/Codex, Yuki Patel/Claude, Priya Zielinski/Claude

**Issue:** The design says convergence should stop parsing a formula subset, but it does not name typed, source-attributed compiled fields for convergence enablement, required vars, evaluate prompt, convergence-relevant step identity, or reserved-step validation inputs. It also leaves retry artifact reuse, vars/options/search-path identity, host downgrade behavior, and zero-write ordering ambiguous.

**Required change:** Define `internal/convergence` projection and validation over canonical compile output: compile, check requirements, project convergence fields, validate convergence domain rules, then write durable state. Retire or rewrite `internal/convergence/formula.Formula` and `ValidateForConvergence`, add a static guard against raw requirement/convergence TOML reads, and specify artifact metadata, retry reuse, downgrade behavior, and fail-closed diagnostics.

### [Blocker] Pack revision and runtime provenance are not durable enough
**Sources:** Saoirse Raman/Claude, Saoirse Raman/Codex, Lena Driscoll/Codex, Ibrahim Park/Codex

**Issue:** The design removes formula-level semver but does not durably stamp enough compile provenance to reconstruct the winning formula, pack binding, requested ref/version constraint, locked revision, content hash, dirty state, and transitive import chain after imports or caches change. It also does not classify raising `requires.formula_compiler` as a breaking pack compatibility event or define requirement-diff reports for import upgrades.

**Required change:** Add a persisted formula provenance contract for root, wisp, and molecule creation. Define pack compatibility policy for compiler capability increases, resolver-level `[pack] requires_gc` rejection for every pack origin before formula selection/staging, byte-exact `ContentHash` rules, and import upgrade/check output that reports compiler requirement deltas.

### [Major] Cross-surface diagnostics and background events are not executable contracts
**Sources:** Marta Hidalgo/Claude, Marta Hidalgo/Codex, Avery Brooks/Claude, Avery Brooks/Codex

**Issue:** The typed diagnostic direction is strong, but CLI exit/status behavior, Huma response bodies, API-routed CLI, dashboard rendering, generated TypeScript clients, order/controller/convergence events, deprecation warnings, grouping keys, recurrence counts, and config-generation reset behavior are not specified or tested as one parity surface.

**Required change:** Add a projection matrix and golden fixtures requiring diagnostic `Code`, message, remediation, source fields, normalized requirement, and host capability to survive every surface unchanged. Decide which diagnostics emit events, register typed payloads for emitted events, name no-event cases, and define dashboard-visible grouped failure state.

### [Major] Docs, examples, generated help, and stale-guidance checks are not a hard rollout gate
**Sources:** Avery Brooks/Claude, Avery Brooks/Codex, Lena Driscoll/Claude, Saoirse Raman/Codex

**Issue:** The design can expose new `[requires]`, `requires_gc`, diagnostic, metadata, and validation behavior before the public documentation contract is complete. `docs/reference/formula.md` still needs a concrete rewrite skeleton, examples, terminology boundaries, legacy `version` policy, and valid replacement snippets.

**Required change:** Make docs, examples, generated help, generated schema/client updates, and stale-guidance checks atomic with or prerequisites for the first user-visible behavior. Include exact TOML examples for formula `[requires] formula_compiler = ">=2"` and pack `[pack] requires_gc`, and decide whether `version` is preserved legacy metadata, warned, or scheduled for removal.

### [Major] Alias-window and external ecosystem gates remain partly prose-only
**Sources:** Elias Vega/Claude, Elias Vega/Codex, Saoirse Raman/Claude, Lena Driscoll/Claude

**Issue:** Legacy `contract`, `version = 2`, empty `[requires]`, inherited requirement conflicts, external SHA-pinned packs, old binary parsing, `bd` probe behavior, first-party dual declaration, and alias removal across major/minor releases are not all backed by executable reports and fixture gates.

**Required change:** Add legacy-contract and legacy-version reports, old-reader/probe compatibility fixtures, inherited requirement conflict diagnostics, first-party dual-declaration CI guards, external pinned-pack inventory, and a lifecycle for the external-support artifact with owner, transition criteria, and public deprecation policy.

### [Major] Future capability growth needs guardrails now
**Sources:** Ibrahim Park/Claude, Ibrahim Park/Codex, Priya Zielinski/Codex

**Issue:** The design supports current `>=2` requirements but does not fully specify future axes, future canonical metadata, `RequirementSource` behavioral misuse, construct registry shape for capability 3, or old-binary behavior for artifacts and metadata produced by newer binaries.

**Required change:** Treat `RequirementSource` as attribution-only and add a CI guard limiting behavioral reads. Pin future `[requires]` axis shape, generalize construct registry entries to `min_compiler_capability` or equivalent, define old-reader handling for future/malformed metadata, and add a release gate/checklist for any new compiler capability or requirement axis.

### [Minor] Secondary operational details still need precise ownership
**Sources:** Marta Hidalgo/Claude, Marta Hidalgo/Codex, Priya Zielinski/Claude, Saoirse Raman/Claude, Avery Brooks/Claude

**Issue:** `OnceKey` ownership, LRU capacity/eviction behavior, process-restart warning suppression, metadata artifact size cutover, `safe_automatic_edit` consumption, diagnostic code namespace conventions, duplicate TOML classification, `HostCapabilities.Source`, and `RootMetadataFacts` allowed callers are not yet crisp enough for implementation.

**Required change:** Add the owner, observable behavior, thresholds, fixtures, and docs for each of these operational surfaces before decomposition or implementation relies on them.

## Disagreements
- Five personas had Claude return `approve-with-risks` while Codex returned `block`; in each case the synthesis sided with `block` when Codex identified a concrete implementation bypass or impossible gate rather than a documentation risk.
- Elias, Marta, Lena, Avery, and Ibrahim all landed on `approve-with-risks`; their findings are not workflow blockers by themselves, but several are prerequisites for resolving the blocker lanes safely.
- Claude often emphasized rollout, old-reader compatibility, diagnostics, and migration governance; Codex often emphasized type/API enforceability, direct code-path inventory, static guards, and fixture coverage. These are complementary, not conflicting.
- The strongest disagreement is whether the design is close enough to decompose. My assessment is no: accepted-artifact enforcement, caller inventory, convergence projection, validation matrix, and pack provenance are still design-level contracts, not implementation details.

## Missing Evidence
- No Gemini artifacts were present for any persona in this attempt; the configured run skipped or omitted Gemini, so all syntheses are based on Claude and Codex only.
- The ten persona synthesis beads for `gc.attempt=30` stamped output paths under `.gc/design-reviews/ga-j195cx/attempt-1/persona-syntheses/`, while `.gc/design-reviews/ga-j195cx/attempt-30/persona-syntheses/` is empty. The files were present and timestamped during this run, so this synthesis used the stamped paths, but the artifact-path mismatch should be fixed.
- No concrete durable-writer signature inventory proves every production write path requires a compiler-minted accepted artifact.
- No current checked-in caller/raw-consumer manifest covers CLI, API, sling, orders, fanout, molecule writers, convergence, source-workflow, graphroute, convoy projections, and duplicate workflow-root predicates.
- No runnable compatibility evidence proves the referenced `GC_NATIVE_FORMULA=false` probe, old Gas City binaries, or `bd >= 1.0.0` parse and validate dual-declared `[requires]` formulas as assumed.
- No convergence call-order diagram, API signature, artifact metadata key, retry read path, or fail-closed diagnostic contract is shown.
- No persisted root/wisp/molecule provenance contract is shown for winning formula source, pack ref, locked revision, content hash, dirty state, and transitive imports.
- No rendered cross-surface diagnostic fixtures prove CLI, API, dashboard, generated TS, Event Bus, order/controller/convergence producers, and validation reports preserve typed diagnostics identically.
- No docs-bundle skeleton or stale-guidance inventory proves user-visible terminology will be updated atomically with behavior.

## Recommended Changes
1. Define and require a non-forgeable accepted compile artifact for every durable production write path; make bare `CompileResult` preview-only.
2. Produce a checked-in caller/raw-consumer manifest and static guard seeded from the current tree, then assign owner/test coverage for each migration row.
3. Replace or implement the `GC_NATIVE_FORMULA=false` compatibility gate, and state the native-runtime-only production precondition unambiguously.
4. Rework the validation matrix around construct identity, location, contribution path, caller path, requirement shape, and diagnostic ordering; add fail-closed generator self-tests.
5. Specify convergence projection over canonical compile output, artifact handoff/retry semantics, host downgrade behavior, and retirement of the legacy convergence subset parser.
6. Add durable formula provenance metadata and pack compatibility policy for compiler capability increases, including resolver-level `requires_gc` enforcement and import requirement-diff reporting.
7. Make diagnostics/event/dashboard/CLI/API/generated-client parity an executable contract with registered payloads, no-event cases, and rendering tests.
8. Bind docs, examples, generated help/schema/client changes, stale-guidance scans, `version` policy, and valid migration snippets to the rollout phase that first exposes user-visible behavior.
9. Add external ecosystem gates: old-reader/probe fixtures, legacy reports, first-party dual-declaration CI guard, external pinned-pack inventory, and an owned external-support lifecycle.
10. Add forward-capability rules for future requirement axes, `RequirementSource` attribution-only use, construct `min_compiler_capability`, and old-reader behavior for future artifacts/metadata.
