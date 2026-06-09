# Design Review Synthesis

## Overall Verdict: block

The design still has credible release-blocking gaps in operator diagnostics, caller coverage, durable writer sequencing, and rollout ordering. Several lanes found the conceptual direction sound, but the remaining issues affect user-visible behavior, durable state safety, and whether the migration can be executed without bypasses.

## Consensus Strengths
- Multiple personas praised the core boundary: formulas declare compiler requirements, while the active Gas City binary and host capability decide whether those requirements can be satisfied.
- Reviewers agreed `internal/formula` should own requirement normalization, legacy `contract` alias handling, host satisfaction, typed diagnostics, construct registry interpretation, and accepted compile artifacts.
- The accepted-artifact and typed-projection model is the right durable-write boundary, provided callers cannot keep using raw metadata, bare `Recipe` values, or preview/delegator symbols as authority.
- The alias-window plan is directionally strong: legacy `contract = "graph.v2"` remains accepted during migration, conflicts with `[requires]` fail deterministically, and alias removal is tied to release evidence rather than only calendar intent.
- The validation-matrix approach is substantially improved: byte-exact grammar, fixture-locked diagnostics, caller-path coverage, construct registry rows, zero-write assertions, and rendered-doc drift checks are the right testing shape.
- Ecosystem review agreed that pack ref/SHA, locked revision, content hash, import binding identity, and pack provenance are the reproducibility boundary; formula `version` should not become formula artifact semver.
- Convergence reviewers agreed the target direction is correct: convergence should stop owning a formula-subset parser and should consume accepted compiler output plus a typed convergence view before writing durable state.
- Future-capability review agreed the construct registry is the right place to centralize capability floors, old-reader behavior, and future construct safety.

## Critical Findings

### [Blocker] Operator Diagnostics Do Not Yet Provide One Bounded Remediation Surface
**Sources:** 04-operator-diagnostics-gate / Claude, Codex; reinforced by 05-caller-integration-inventory
**Issue:** A disabled `formula_v2` host can still produce many operator-visible failures without a defined rollup row, deterministic burst budget, or persisted cadence state. The design also promises surface parity, but host source line/column is not fixture-locked across Huma JSON, generated TypeScript, dashboard, events, direct CLI, and API-routed CLI. Background warning occurrence updates also conflict with immutable accepted artifacts unless mutable cadence state is separated.
**Required change:** Define the shared host/config diagnostic rollup projection, grouping key, child subject details, config-generation behavior, typed wire shape, event cadence, persisted counters, restart/reload behavior, and burst fixtures. Add host source line/column to the typed diagnostic and event payloads or narrow the parity contract explicitly. Keep accepted artifacts immutable and put repeat occurrence state in producer-owned diagnostic state.

### [Blocker] Caller Inventory And Durable Writer Fences Still Allow Bypasses
**Sources:** 05-caller-integration-inventory / Claude, Codex; reinforced by 01-compiler-boundary-invariant, 09-convergence-subset-reviewer
**Issue:** The raw-consumer and workflow-root guards are scoped too narrowly for the current tree. Behavioral callers exist across `internal/graphroute`, `internal/sling`, `internal/api`, `internal/dispatch`, `internal/molecule`, `internal/convergence`, first-party packs, examples, tutorials, and prompts. First-party prompt commands such as `gc bd mol wisp` and `gc bd mol burn` can act as producer paths even when no Go shell-out exists.
**Required change:** Commit a generated caller manifest that includes raw metadata reads, direct compiles, fragment compiles, convergence parser use, API/dashboard projections, store materialization, and first-party prompt/formula command references. Make production guards repository-wide with narrow owner-package allowlists, add deletion tests for retired helpers, and either rewrite, block, or explicitly classify legacy `gc bd mol ...` producer commands before requires-only formulas are distributed.

### [Blocker] Rollout Sequencing Contradicts The Required Safety Gates
**Sources:** 06-rollout-sequencing-reviewer / Claude, Codex; reinforced by 02-contract-migration-guardian, 08-docs-dx-terminology
**Issue:** Phase 2 and Phase 3 disagree about when first-party dual declarations land relative to visible diagnostics, generated help, API fields, events, validation reports, and metadata writers. Phase 4 also does not prove shared molecule durable-writer hardening lands before migrated producers such as sling, orders, retry/on-complete, fanout, convergence, and dashboard projections.
**Required change:** Choose one Phase 2/3 sequence: either move first-party dual declarations and their green inventory gate before any visible/runtime surface, or keep Phase 2 dormant until Phase 3 passes. Add a mandatory writer-lockdown gate before producer migration, covering accepted-artifact-only writer signatures, no bare `Recipe` or `CompileResult` durable APIs, and zero-write tests for roots, children, hooks, convoys, metadata, retry, fanout, and convergence.

### [Major] Host Capability Authority And Requirement Provenance Need Blocking Guards
**Sources:** 01-compiler-boundary-invariant / Claude, Codex; reinforced by 10-future-capability-architect
**Issue:** `internal/formula` can still read package-global formula-v2 state in production paths, and `RequirementSource` is visible enough to become accidental runtime authority. Same-identity reuse after host downgrade also needs property-level proof that it cannot expand durable writes beyond the accepted projection snapshot.
**Required change:** Add blocking guards for production reads of legacy formula-v2 globals inside `internal/formula`, except explicitly named edge adapters. Add a `RequirementSource` provenance-only guard with a narrow diagnostic/provenance/migration allowlist. Add property tests for `CompilerCapability(0)` non-escape and same-identity host-downgrade reuse with write-intent limits.

### [Major] Parser Matrix And Construct Registry Semantics Are Not Fully Executable
**Sources:** 03-parser-validation-matrix / Claude, Codex; 10-future-capability-architect / Claude, Codex
**Issue:** The matrix still has governance gaps: combined-defect ownership, literal count locks, checked `coverage_intent`, caller-path vocabulary drift, BurntSushi/toml version guards, and generated-doc drift checks are not all fixture-locked. `step.expand` remains ambiguous: the design must decide whether field presence requires capability 2 or whether only contributed formulas drive escalation. Construct registry rows can also drift if `Minimum` and typed `Capability` disagree.
**Required change:** Add owning suites and count locks for combined defects, require literal count locks, replace free-text `coverage_intent` with checked ids, make caller paths a single source of truth, guard the TOML parser version, and add generated-doc staleness checks. Resolve `step.expand` semantics and add rows across root, child, loop body, inline expansion, compose expansion/map, aspect, and transitive imports. Add construct-registry invariants so `Minimum` and `Capability` cannot disagree.

### [Major] External Pack Evidence And Alias Exit Criteria Remain Underdefined
**Sources:** 02-contract-migration-guardian / Claude, Codex; 07-pack-versioning-ecosystem / Claude, Codex
**Issue:** Alias removal depends on external-support data, packman provenance, compatibility reports, and migration hints that are not yet specified as executable artifacts. Unknown and unreachable external consumers lack a deterministic public-notice and expiry path. `formula.migration.pin_pack_revision` is named but undefined, and `[pack] requires_gc` grammar compatibility must be settled before enforcement can safely run ahead of formula validation.
**Required change:** Define the external-support artifact schema, canonical storage path, malformed-row behavior, discovery inputs, outreach rules, public notice fields, unknown-expiry semantics, and alias-removal-gate consumption. Make the alias-window clock start only when the required migration bundle is present and non-stale, and require both two completed minor releases and 60 calendar days, with the longer floor binding. Define `formula.migration.pin_pack_revision`, add edit-target semantics for SHA-pinned packs, and resolve or fixture-lock `requires_gc` grammar compatibility.

### [Major] Convergence Needs A Stronger Preflight Boundary And Active-Loop Failure Contract
**Sources:** 09-convergence-subset-reviewer / Claude, Codex; reinforced by 05-caller-integration-inventory
**Issue:** The design does not yet prove preflight occurs before every durable convergence write. Current create/retry behavior may write roots or metadata before formula validation, and the protected caller set omits manual iterate, fallback pour, missing-state repair, missing-child repair, next-iteration repair, and speculative pending-wisp adoption or burn. Active loops that later encounter an unsatisfied compiler requirement also lack defined state, events, status, and cleanup behavior.
**Required change:** Specify the concrete preflight call sequence before create, retry, next iteration, manual iterate, speculative pour, fallback pour, missing-state repair, missing-child repair, and pending-wisp adoption. Add zero-write fixtures for disabled hosts and failure classes. Define active-loop unsatisfied-requirement state, waiting metadata, event/status payload, retry semantics, and pending-wisp cleanup. Replace raw convergence subset parsing with a blocking static guard and an accepted-projection write fence.

### [Major] Documentation And Terminology Gates Must Precede User-Visible Behavior
**Sources:** 08-docs-dx-terminology / Claude, Codex; reinforced by 03-parser-validation-matrix, 10-future-capability-architect
**Issue:** The rollout can still expose `[requires]`, diagnostics, generated help, API fields, dashboard previews, or migration reports before public references and stale-guidance checks are updated. Terms such as formula `[requires]`, pack `requires_gc`, pack requirements, host `formula_v2`, compiler capability, implementation, and graph workflow remain easy to conflate.
**Required change:** Make the docs/check bundle a hard predecessor for all user-visible behavior. Update `docs/reference/formula.md`, `engdocs/architecture/formulas.md`, PackV2 author docs or inventory rows, generated surfaces, and proposal supersession markers before diagnostics ship. Add stale-guidance matcher semantics and fixtures for `contract`, formula `version`, `GC_NATIVE_FORMULA=false`, and `formula_v2`, plus a checked placeholder registry for doctests.

### [Minor] Persona Artifacts For Attempt 80 Were Stamped Under The Wrong Attempt Directory
**Sources:** Synthesis artifact inspection; persona synthesis bead metadata
**Issue:** `.gc/design-reviews/ga-j195cx/attempt-80/persona-syntheses/` was empty, while the ten closed persona synthesis beads for `gc.attempt=80` stamped fresh outputs under `.gc/design-reviews/ga-j195cx/attempt-1/persona-syntheses/`. This appears to be existing workflow artifact-path drift, not a review-content disagreement.
**Required change:** Repair the design-review workflow so persona synthesis beads write under the current `$ATTEMPT_DIR/persona-syntheses` and downstream synthesis can rely on the documented path. This report used the bead-declared output paths because they were complete, current, and metadata-linked to attempt 80.

## Disagreements
- Several persona syntheses disagreed internally on severity. In 04, 05, and 06, Claude often returned `approve-with-risks` while Codex returned `block`; the persona syntheses chose `block` because the issues affect operator guarantees, durable write safety, and executable rollout order. I agree with those stricter classifications.
- 02 disagreed between `approve-with-risks` and `approve`. The stricter result is appropriate because alias-window ownership, clock anchoring, and unknown external consumer handling must be executable before removal decisions are safe.
- 03 disagreed mostly on focus rather than verdict: Claude emphasized matrix governance, while Codex emphasized stale caller vocabulary and `step.expand` ambiguity. Both are required because governance checks need stable semantics.
- 07 disagreed on which ecosystem risks matter most. Claude emphasized packman schema, migration hints, and external outreach; Codex emphasized pack-floor grammar and closed discovery. The design needs both machine-readable contract detail and an operational support inventory.
- 09 disagreed on projection placement. Claude objected to a subsystem-named convergence projection in `internal/formula`; Codex focused on durable writes needing accepted projections. Package placement matters because formula compilation should own generic facts and convergence should own convergence policy.
- Kimi 2.6 was absent from every persona synthesis. That is allowed by the persona bead contract because Claude and Codex were the required inputs, but it remains a missing third-model signal.

## Missing Evidence
- Executable diagnostic rollup fixtures, event cadence fixtures, OpenAPI/TypeScript/dashboard/event provenance fixtures, and store CAS/singleton conformance tests.
- A generated repository-wide caller manifest and matching production static guards covering Go code, first-party packs, examples, tutorials, and prompts.
- Phase gates proving first-party dual declarations and docs/check bundles precede visible surfaces, and proving writer-lockdown precedes producer migration.
- Property tests for host capability non-escape, requirement-source non-authority, accepted-artifact reuse under host downgrade, and immutable accepted warnings versus mutable producer cadence.
- Matrix fixtures for combined defects, checked `coverage_intent`, `step.expand`, overflow diagnostics, parser version changes, and generated/public-doc drift.
- External-support inventory schema, public notice and unknown-expiry rules, packman provenance contingency, and worked migration-hint JSON for pinned external packs.
- Convergence preflight call-site sketches, zero-write fixtures, active-loop failure semantics, and static guards against subset parser or raw metadata bypasses.
- Public docs, glossary, PackV2 terminology surfaces, stale-guidance YAML, doctest fixtures, and placeholder validation behavior.
- Current-attempt persona synthesis files under `.gc/design-reviews/ga-j195cx/attempt-80/persona-syntheses/`; the current outputs were found through the closed persona beads under `attempt-1/persona-syntheses/`.

## Recommended Changes
1. Block the rollout until operator diagnostics have a typed rollup model, bounded event cadence, host-source provenance on every promised surface, immutable accepted artifacts, and store conformance for any singleton/CAS state.
2. Commit the generated caller manifest and make raw-consumer, workflow-root, and durable-writer guards repository-wide, including first-party prompts and pack formulas.
3. Resolve rollout sequencing: first-party dual declarations and docs/checks must precede visible surfaces, and accepted-artifact-only writer hardening must precede migrated producers.
4. Add blocking compiler-boundary tests for host capability authority, `RequirementSource` provenance-only behavior, host-downgrade reuse, and `CompilerCapability(0)` non-escape.
5. Finish the parser and construct registry contract by resolving `step.expand`, locking matrix governance, and enforcing `Minimum`/`Capability` consistency.
6. Define executable external-pack migration evidence: external-support schema, public notice and expiry rules, packman provenance contingency, migration-hint examples, and alias-window clock eligibility.
7. Specify convergence preflight before every durable write and define active-loop unsatisfied-requirement behavior with zero-write fixtures.
8. Make documentation, terminology, stale-guidance, and doctest gates mandatory before any user-visible behavior ships.
9. Fix the design-review artifact-path bug so current-attempt persona syntheses are written under the current attempt directory.
