# Design Review Synthesis

## Overall Verdict: block

The design is directionally strong: reviewers repeatedly endorsed the accepted compile artifact boundary, typed compiler requirements, fail-closed future capability model, and the move away from raw formula metadata consumers. It cannot be approved yet because several executable contracts remain contradictory or underspecified, especially around parser validation, host diagnostics, active legacy-root repair, rollout ordering, and pack provenance. These are design-level blockers because independent implementers could follow the document and still ship incompatible behavior.

## Consensus Strengths
- Multiple personas praised the compiler boundary direction: formula declarations are interpreted by `internal/formula`, while durable callers consume accepted artifacts and typed projections instead of raw formula evidence.
- The `requires.formula_compiler` model is intentionally narrow and conservative: omitted, empty, and explicit `>=1` normalize to compiler capability 1, while future or unknown requirements fail closed.
- The accepted-artifact write boundary, projection validation, and source/provenance concepts are the right shape for removing stale graph metadata authority from convergence, sling, fanout, orders, and API paths.
- The release strategy is evidence-driven rather than calendar-driven, with alias removal, stale-guidance checks, compatibility corpus reports, and docs gates treated as release artifacts.
- The terminology model separates formula compiler capability, host capability, compiler implementation, legacy `contract`, formula `version`, pack revision, pack `requires_gc`, and schema versions.
- External-pack handling is cautious: SHA-pinned and unknown external consumers default to supported legacy status rather than silent breakage.
- Future-axis handling is typed rather than generic pass-through, which avoids premature abstraction while still leaving room for additional capability axes.

## Critical Findings

### [Blocker] Validation Matrix Is Not Mechanically Enforceable
**Sources:** Priya Zielinski, Elias Vega, Nadia Sorenson, Saoirse Raman, Avery Brooks
**Issue:** The design depends on a generated validation matrix, but the current contract does not fully name the fixture path, row schema, generator contract, generated Go tests, diagnostic golden outputs, or literal/computed count locks for every normative suite. JSON formula behavior is unresolved while JSON loading remains live, and TOML/parser boundary behavior is not pinned tightly enough for duplicate keys, unsupported shapes, source attribution, future minima, and legacy `check` compatibility.
**Required change:** Define the validation matrix as an executable artifact: fixture paths, row schema, generated test names, golden diagnostics, count-lock policy, generator self-tests, TOML boundary rows, future-minimum grammar rows, public `check` compatibility rows, and a firm JSON decision. If JSON formulas stay supported, add duplicate-member, pointer/source attribution, parity, and CI fixtures; if not, require explicit rejection before durable writes with a named diagnostic.

### [Blocker] Host-Capability Diagnostics Can Hide Distinct Failures
**Sources:** Marta Hidalgo, Priya Zielinski, Lena Driscoll, Avery Brooks
**Issue:** Fatal host-capability grouping is unsafe as specified. The `OnceKey` and dashboard group key can collapse a newly failing formula into an old disabled-host occurrence when config generation is unchanged. The design also requires grouped diagnostic state writes while saying fatal diagnostics create zero protected state, without clearly separating allowed diagnostic-state writes from forbidden order/convergence/fanout/runtime writes.
**Required change:** Redesign the grouping contract so operator-action grouping preserves formula identity, producer, subject, requirement source path/key/value, host-capability provenance, and content hash as explicit dimensions or subrows. Add fixtures for multi-producer disabled-host failures, repeated scans, config reloads, LRU eviction, dashboard grouping, API/Huma JSON, generated TypeScript, CLI rendering, event payloads, and zero protected writes with allowed diagnostic-state persistence.

### [Blocker] Active Legacy-Root Repair Has Contradictory Contracts
**Sources:** Yuki Patel, Elias Vega, Lena Driscoll, Felix Berger, Nadia Sorenson
**Issue:** The design conflicts on whether graph-control mutators must auto-compile and stamp legacy-only active roots before the first graph-specific write, or must fail closed until an operator runs `gc formula repair-root-artifact`. The repair command itself is underspecified, and caller migration lacks symbol-level ownership for raw consumers, package-global capability accessors, API bridges, sling helpers, convergence paths, and artifact-ref readers.
**Required change:** For every caller path, record exactly one repair mode: `auto_repair`, `operator_repair_required`, or `read_only_fail_closed`. Specify `gc formula repair-root-artifact` inputs, source resolution, validation, atomic write behavior, idempotency, conflict handling, exit codes, and concurrent repair behavior. Seed the raw-consumer manifest, add monotonic-shrink and post-4g ceiling checks, collapse `CompileResult` to one canonical shape, and forbid durable writers from using preview-only fields or raw compiler outputs directly.

### [Blocker] Rollout Ordering Exposes User-Visible Diagnostics Before Support Artifacts
**Sources:** Lena Driscoll, Avery Brooks, Elias Vega, Saoirse Raman
**Issue:** Phase 4 migrates CLI, API, sling, orders, convergence, fanout, dashboard, and controller producers to typed diagnostics before Phase 6 lands the docs, generated help/schema/API/dashboard artifacts, runnable examples, and stale-guidance material those diagnostics require. Rollback controls are also prose-only for durable producers, and API/dashboard generated-client coupling is not covered by a shared rollback or render-parity fixture.
**Required change:** Reorder or split the rollout so each visible diagnostic surface has its docs/help/schema/API/dashboard/generated artifacts before it can merge or ship. Add a caller-manifest rollback contract for every durable producer, including rollback owner, control, command/test proof, zero-write proof, render-parity proof where relevant, and rules for docs that remain landed while a producer is disabled.

### [Blocker] Packman Provenance And External-Pack Workflow Are Prerequisites But Not Sequenced
**Sources:** Saoirse Raman, Lena Driscoll, Elias Vega
**Issue:** Alias-removal evidence, external pinned-pack support expiration, imported-pack floor enforcement, reproducibility reports, and `--alias-removal-gate` depend on provenance that current `packs.lock` schema 1 cannot prove. External maintainers also lack a complete executable workflow for validation, requirement diffs, legacy reports, stable JSON schemas, exit codes, docs entry points, release notes, and stale-guidance checks.
**Required change:** Promote packman schema 2, or an equivalent packman-owned provenance contract, to an explicit rollout prerequisite before any resolver/import enforcement or alias-removal gate. Define lockfile fields, migration, schema-1 behavior, deterministic tests, owners, rollback unit, and compatibility rules. Define external-author commands and JSON schemas for `--pack-path`, `--pack-source`, `--provenance`, `--requirement-diff`, and `--legacy-contract-report`, including warning/error exit semantics and worked examples for pinned SHA, tags, branches, local dirty paths, transitive imports, shadowed formulas, unreachable sources, and registry mirrors.

### [Major] Compiler Boundary And Accepted-Artifact Proofs Are Still Porous
**Sources:** Nadia Sorenson, Yuki Patel, Felix Berger, Ibrahim Park
**Issue:** Several boundary types have no named package home, `acceptedCompileProof.identityHash` is underspecified, preview and accepted projection views are not structurally distinct, `HostCapabilities` construction can be bypassed through exported fields or global accessors, and workflow-root predicate ownership remains loose. Convergence also has competing artifact-ref keys and an undefined `ConvergenceRuntimeInputs` carrier.
**Required change:** Name the package home for source/provenance types, specify proof hashes byte-for-byte, make accepted projections distinguishable from previews, structurally constrain host capability construction, narrow raw-consumer allowlists, define one canonical convergence artifact ref with conflict diagnostics, and define runtime input validation against accepted projection facts.

### [Major] Alias-Removal And Compatibility Evidence Is Incomplete
**Sources:** Elias Vega, Saoirse Raman, Lena Driscoll, Priya Zielinski
**Issue:** The alias-removal gate is correctly evidence-based, but its JSON contract lacks release-history inputs, background accepted-alias counts, compatibility corpus artifact paths, rollback-clock semantics, external support transitions, and concrete baseline binaries. Legacy alias parsing also needs byte-exact fixture coverage and diagnostic precedence for unsupported future compiler requirements mixed with legacy contracts.
**Required change:** Add `--alias-removal-gate --json` fields and blocking exit semantics for release tag/date, legacy-only count, dual-declared count, corpus artifact path, background accepted-alias count, rollback-window status, support-row expiration, external pack class, and baseline binary digest or tag. Add fixtures for unsupported legacy values, future minima, external/transitive/shadowed/SHA-pinned remediation, and stale-guidance matching.

### [Major] Convergence Destination Is Sound But Transition And Diagnostics Are Incomplete
**Sources:** Felix Berger, Yuki Patel, Priya Zielinski, Marta Hidalgo
**Issue:** Reviewers support removing the convergence subset parser, but the design does not cover every current convergence validation rule, especially evaluate-prompt content checks. The migration window can leave convergence using the legacy parser while other entry points enforce accepted artifacts, allowing `[requires]` bypass. Pre-create compile failures, artifact-ref conflicts, and retry identity behavior are not fixture-locked.
**Required change:** Add or explicitly retire evaluate-prompt content diagnostics, define projection/equivalence fixtures, add a transition fence or shadow-compile rule for convergence during sub-phases 4b-4e, specify pre-create compile-failure behavior and subject keys, and add fixtures for same-identity retry, changed-identity retry, host toggles, and artifact-ref conflicts.

### [Major] Documentation Gates Need To Become Real Build Inputs
**Sources:** Avery Brooks, Lena Driscoll, Saoirse Raman, Elias Vega
**Issue:** The docs model is strong, but the current tree still contains stale formula `version`, old `formula_v2`/`bd --graph` behavior, stale rollback prose, and legacy examples. The proposed doctest, stale-guidance baseline, first-party inventory, docs-check report, generated CLI/schema/OpenAPI/TS artifacts, and `make formula-docs-check` gate are not yet defined as executable blockers.
**Required change:** Define a single docs gate that runs formula doctests, stale-guidance scans, first-party inventory validation, generated help/schema checks, and docs report generation. Rewrite `docs/reference/formula.md` early, disambiguate formula `[requires]` from pack `[pack].requires_gc` and convergence `required_vars`, resolve JSON docs scope, and add parser-backed stale-guidance fixtures before making the scan CI-blocking.

### [Minor] Several Edge Contracts Need Tightening Before Implementation
**Sources:** Priya Zielinski, Elias Vega, Marta Hidalgo, Lena Driscoll, Ibrahim Park
**Issue:** Minor but recurring gaps include TOML parser version authority, dashboard remediation affordance boundaries, Phase 5 deliverable naming, future `formula_compiler` grammar wording, default-capability artifact identity, and whether source spelling affects `compileID`, projection snapshots, `OnceKey`, or root metadata.
**Required change:** Pin fixture behavior as authority or assert the TOML library version, define dashboard UI limits, give Phase 5 a concrete report deliverable or demote it to an invariant check, constrain future grammar growth to monotonic integer minima, and fixture omitted/empty/explicit `>=1` identity behavior.

## Disagreements
- Several persona syntheses had Claude at `approve-with-risks` and Codex at `block`. I side with `block` where the issue is an executable contract gap: parser validation, diagnostic grouping, active-root repair, rollout ordering, and pack provenance can all produce incompatible implementations if left as prose.
- Some reviewers treated docs, generated artifacts, stale-guidance matchers, and release reports as implementation details. I assess them as design requirements because the design makes them release-blocking evidence and user-visible diagnostic prerequisites.
- Reviewers disagreed mainly by omission on JSON formula handling, legacy public `check` behavior, sourceworkflow parity deltas, `CompileResult` shape, and convergence artifact refs. I treat those omissions as additional gaps rather than contrary evidence because each affects a named contract surface.
- Kimi 2.6 artifacts were absent across the persona syntheses. The persona contracts allow that absence, so it is not a workflow failure, but it reduces independent coverage.

## Missing Evidence
- Kimi 2.6 review artifacts were absent for all persona lanes.
- No complete generated validation matrix contract with fixture paths, schemas, generated tests, golden outputs, and count locks.
- No firm JSON formula decision with parity or rejection fixtures while the JSON loader remains live.
- No packman schema-2 artifact, migration test, owner, phase, or rollback unit.
- No worked external-pack lifecycle covering pinned SHA, lockfile updates, requirement diffs, old/new consumers, and successful revalidation.
- No seeded raw-consumer manifest with symbol-level rows for sling, sourceworkflow, graphroute, API global capability bridge, convergence, fanout, orders, convoy, and compatibility writers.
- No executable active-root repair fixture matrix for source-present, source-missing, host-downgraded, unsupported-artifact, and concurrent-repair cases.
- No pre-tag alias-window bundle check or `formula-compiler-min-floor.json` schema proving release bundle completeness and patch-release semantics.
- No local `make formula-docs-check` output or parser-backed stale-guidance fixture set.
- No convergence artifact-ref conflict fixture, `ConvergenceRuntimeInputs` type contract, or pre-create compile-failure behavior.
- No identity fixtures for omitted `[requires]`, empty `[requires]`, and explicit `formula_compiler = ">=1"` across compileID, projection snapshots, once keys, and root metadata.

## Recommended Changes
1. Land the validation matrix contract first: paths, schema, generated tests, golden diagnostics, count locks, TOML boundary rows, JSON decision, future-minimum grammar, and legacy `check` compatibility.
2. Resolve the active legacy-root repair contract and seed the caller/raw-consumer manifest before migrating durable producers.
3. Promote packman schema 2 or equivalent provenance to an explicit prerequisite, with migration, deterministic tests, owners, and rollback.
4. Redesign host-capability diagnostic grouping and warning suppression, then fixture CLI/API/dashboard/TS/event behavior plus zero protected writes.
5. Reorder the rollout so docs, generated artifacts, examples, and stale-guidance gates precede every user-visible diagnostic surface.
6. Define the external-author validation workflow and stable JSON schemas with worked lifecycle examples for pinned and transitive packs.
7. Specify accepted-artifact proof hashes, projection type separation, convergence artifact refs, and `ConvergenceRuntimeInputs`.
8. Complete alias-removal gate evidence, rollback-clock semantics, byte-exact alias parsing, compatibility corpus baselines, and stale-guidance matcher rules.
9. Turn the docs bundle into an executable local gate and rewrite stale formula reference material before exposing diagnostics.
10. Add future-capability identity and axis-independence fixtures so default-capability spelling and future axes cannot drift into behavioral decisions.
