# Design Review Synthesis

## Overall Verdict: block

The review blocks because the parser/validation matrix, caller migration, and PackV2 provenance lanes each found bypass-class gaps that can let graph-control behavior, durable writes, or external pack validation drift from the proposed compiler contract. The design direction is consistently praised, especially the move toward `internal/formula` as the central authority, but several contracts are not yet precise or enforceable enough to implement safely.

## Consensus Strengths
- Reviewers broadly support making `internal/formula` the single normalization, validation, diagnostics, and accepted-artifact authority for raw `contract`, `[requires]`, legacy `version`, v2-only constructs, and durable compile artifacts.
- The fail-closed accepted-artifact model, static raw-consumer guard, typed diagnostics, and shared CLI/API/dashboard/event projections are viewed as the right direction.
- The narrow byte-exact requirement grammar, normalized runtime behavior, and fail-closed handling of unknown `[requires]` axes give the design a sound forward-compatibility base.
- The alias-window strategy, first-party dual-declaration idea, release reports, and stale-guidance checks are strong foundations if converted into executable gates.
- Convergence projection from accepted compiler artifacts is the right replacement for the old convergence-specific formula subset.

## Critical Findings

### [Blocker] Workflow-Control Metadata Validation Is Not Closed
**Sources:** Priya Zielinski / Claude, Priya Zielinski / Codex
**Issue:** The v2-only trigger table is narrower than the workflow-control metadata surface generated and consumed by the system. Keys and values such as `gc.kind=fanout`, `gc.kind=spec`, `gc.control_for`, `gc.for_each`, `gc.bond`, `gc.bond_vars`, `gc.fanout_mode`, `gc.output_json_required`, `gc.spec_for`, and related Ralph/fanout metadata are not governed by one normative registry with byte-exact matching rules.
**Required change:** Replace the hardcoded trigger table with a generated workflow-control metadata registry seeded with every current graph-control key and value, then add positive and missing-requirement fixtures plus mandatory count locks for every validation suite.

### [Blocker] Graph-Control Durable Writes Can Still Bypass Accepted Artifacts
**Sources:** Yuki Patel / Claude, Yuki Patel / Codex
**Issue:** Active graph-control appenders are not all forced through accepted-artifact validation before creating graph-specific durable state. Retry, Ralph retry, control, fanout, `on_complete`, missing-child repair, next-iteration, hook, dependency, and continuation paths remain ambiguous or absent from the caller manifest.
**Required change:** Add a generic active-root rule: every graph-control child-creation or mutation path must load and validate the root accepted artifact before any graph-specific write. Legacy roots without accepted artifacts must compile, accept, and stamp the original formula first, or fail closed with zero writes.

### [Blocker] PackV2 Lockfile And Provenance Sources Conflict
**Sources:** Saoirse Raman / Claude, Saoirse Raman / Codex
**Issue:** The design introduces lockfile and provenance semantics that appear to conflict with existing `internal/packman` ownership and the current `packs.lock` schema. It also omits PackV2 import binding identity from provenance, persisted metadata, diagnostics, `migration_hints`, and requirement-diff output, even though binding paths are what authors can edit.
**Required change:** Reconcile the proposed lock/provenance schema with current PackV2 lockfile ownership and format, including versioning, read/write compatibility, migration behavior, and tests. Carry import binding identity through provenance, validation JSON, persisted metadata, migration hints, and requirement diffs.

### [Major] Accepted Artifact Identity And Durable Compile Boundaries Are Under-Specified
**Sources:** Nadia Sorenson / Claude, Nadia Sorenson / Codex; Felix Berger / Claude, Felix Berger / Codex; Ibrahim Park / Claude
**Issue:** `OptionsHash`, `VarsHash`, source/provenance identity, host capability identity, content hashes, redaction/projection rules, and accepted-artifact cross-version behavior are not defined precisely. The design also leaves tension between bare name/search-path compile helpers and resolver-owned source provenance.
**Required change:** Define the accepted artifact identity contract at byte-level precision, make resolver-owned source data the durable compile boundary, specify compile-time versus runtime vars, and add test-only artifact minting plus production import/static guards.

### [Major] Convergence Identity, Projection, And Zero-Write Guarantees Need Sharpening
**Sources:** Felix Berger / Claude, Felix Berger / Codex; Nadia Sorenson / Claude
**Issue:** Convergence projection parity is not fixture-locked, legacy convergence roots without accepted artifact metadata can rebind silently after pack layering or shadowing changes, and `ValidateProjection` cannot prove required-var handling without caller vars or compiler-owned satisfaction results.
**Required change:** Add projection-parity fixtures for requirements, source, provenance, compile artifact reference, and host capability provenance; define legacy convergence-root identity and host downgrade behavior; change validation APIs so required-var checks do not revive raw subset parsing; and require zero-write tests for create, retry, next iteration, missing-child repair, and speculative wisps.

### [Major] Legacy Alias Evidence And Warning Surfaces Are Not Durable Enough
**Sources:** Elias Vega / Claude, Elias Vega / Codex; Marta Hidalgo / Claude, Marta Hidalgo / Codex
**Issue:** Alias usage evidence for accepted `contract = "graph.v2"` compiles may be lost behind producer-local warning suppression, process restarts, or release-time prose. Background producers also lack a complete bounded logging and warning policy for successful deprecation warnings.
**Required change:** Choose one authoritative, restart-safe source for accepted legacy-alias compile counts across orders, convergence, fanout, retries, controller paths, API paths, and dashboard/report projections. Add repeated-loop fixtures proving grouped warning behavior, zero warning events, and no per-tick log spam.

### [Major] Rollout, Docs, And Stale-Guidance Gates Are Not Yet Executable
**Sources:** Lena Driscoll / Claude, Lena Driscoll / Codex; Avery Brooks / Claude, Avery Brooks / Codex; Saoirse Raman / Claude, Saoirse Raman / Codex
**Issue:** First-party dual declaration, diagnostic exposure, docs updates, generated help/schema updates, and stale-guidance checks are described but not pinned to phases, PRs, commands, path globs, or CI failures. Live docs and examples still teach older behavior, and the proposal rollback guidance can be mistaken for current policy.
**Required change:** Make docs, examples, generated artifacts, first-party dual declarations, and stale-guidance scanning a blocking predecessor for every user-visible diagnostic surface. Name exact files, generated artifacts, path globs, exception policy, stale matchers, and rollout phase gates.

### [Major] Host Capability Provenance And Mutable Feature Globals Can Drift
**Sources:** Yuki Patel / Codex; Marta Hidalgo / Codex; Ibrahim Park / Claude
**Issue:** Omitted/default `formula_v2=false`, explicit false, deprecated `graph_workflows` promotion, graph-apply globals, and accepted-artifact host capability identity are not tied into one enforceable provenance and lifetime contract.
**Required change:** Add host-capability source fields and fixtures across CLI, API, dashboard, events, order dispatch, convergence, and fanout. Route graph-apply decisions through per-operation options or accepted-artifact/write-intent validation rather than package globals.

### [Major] Future Requirement Axes Need Normative Shape And Old-Reader Rules
**Sources:** Ibrahim Park / Claude, Ibrahim Park / Codex; Priya Zielinski / Claude, Priya Zielinski / Codex
**Issue:** Future `[requires]` axes, unknown schema versions, `>=1` semantics, integer boundary behavior, and unknown future-axis workflow-root classification are not fully specified. Parser edge cases such as signed, zero-padded, overflow-sized, NUL-containing, Unicode, nested-table, duplicate-key, and JSON-equivalent inputs are missing from the matrix.
**Required change:** State that future axes are flat scalar strings directly under `[requires]`, define axis ownership and schema-version bump policy, decide whether explicit `>=1` remains accepted provenance, and expand parser/validation fixtures with deterministic diagnostics.

### [Minor] Projection And Operator Details Need Final Decisions
**Sources:** Nadia Sorenson / Claude; Marta Hidalgo / Claude; Elias Vega / Claude; Avery Brooks / Codex
**Issue:** Several lower-level but user-visible choices remain unsettled: `acceptedCompileProof.nonce`, HTTP `409` versus `412`, warning LRU sizing, CLI suppression key shape, report schema versioning, non-formula `contract` terminology, and dashboard grouping of operator-fixable versus author-fixable diagnostics.
**Required change:** Resolve these as explicit design decisions with fixtures or static checks where they affect wire contracts, reports, dashboard state, or operator remediation.

## Disagreements
- Several disagreements are severity assignments, not factual conflicts. Priya, Yuki, and Saoirse each had one model rate the design lower because the gaps were bypass-class or source-of-truth conflicts; this synthesis agrees with the stricter assessment and sets the global verdict to `block`.
- Claude often emphasized rollout, documentation, operational visibility, and compatibility lifecycle risks; Codex often emphasized repository-grounded caller paths, PackV2 ownership conflicts, and enforceable static guards. These positions are complementary and should be merged rather than choosing one side.
- Ibrahim's lane split between `approve` and `approve-with-risks`. The forward-compatibility model is sound, but the risk verdict is more appropriate until future-axis shape, old-reader behavior, and boundary fixture requirements are normative.
- Gemini artifacts were absent throughout this attempt, which is acceptable because the workflow was configured with Gemini skipped; their absence is not treated as a workflow failure.

## Missing Evidence
- A single generated workflow-control metadata registry seeded with all current graph-control keys and `gc.kind` values.
- Machine-checkable count locks or generated-count expressions for every compiler requirements matrix suite.
- Grep-backed caller manifests and retirement tables proving every raw consumer, graph-control appender, workflow-root query, SQL predicate, graph-apply global, and legacy helper has a disposition.
- Byte-level accepted-artifact identity inputs, vars redaction/projection policy, source/provenance identity, host capability identity, and cross-version artifact reuse rules.
- Zero-write fixtures across order dispatch, convergence create/retry/repair, fanout, retry, Ralph, control, hooks, dependencies, and continuations.
- PackV2 lockfile compatibility tests and worked external-author validation JSON showing binding identity, provenance, diagnostics, and migration hints for pinned packs.
- Durable legacy-alias usage evidence that survives warning suppression and process restart.
- CI contract for docs/generated-help/schema/stale-guidance scanning, including path globs, exception policy, and v2-only construct matchers.
- Fixture rows for parser boundary cases, future-axis structured input, unknown schema versions, and non-formula `contract` namespace handling.

## Recommended Changes
1. Close the three blockers first: generated workflow-control metadata registry, universal accepted-artifact gates for graph-control durable writes, and PackV2 lockfile/provenance reconciliation with binding identity.
2. Specify accepted-artifact identity in one normative section, including hash inputs, vars projection, source/provenance identity, host capability identity, cross-version reuse, and test-only minting.
3. Convert the requirements validation matrix into an executable contract with mandatory suite count locks and explicit parser boundary rows.
4. Add a complete caller migration manifest with static guards for raw metadata reads, workflow-root predicates, SQL filters, graph-apply globals, and deprecated helper functions.
5. Define convergence identity and projection behavior, including legacy root handling, host downgrade behavior, projection validation inputs, and zero-write atomicity.
6. Make docs, examples, generated artifacts, first-party dual declarations, and stale-guidance scanning blocking gates before any user-visible diagnostics or warning surfaces ship.
7. Define durable legacy-alias reporting, warning suppression, external-support lifecycle, and schema-versioned release reports.
8. Settle future-axis and host-capability rules: flat scalar axes, axis ownership, requirements schema-version policy, `>=1` semantics, old-reader behavior, and capability provenance across all projections.
