# Design Review Synthesis

## Overall Verdict: block

Nine of ten persona syntheses returned `block`, so the global verdict is `block` by worst-verdict-wins. The strongest consensus is that the design is directionally right about making Gas City own formula compiler requirements, normalized diagnostics, and compatibility checks, but the current review cannot approve because the canonical design artifact is absent or mis-snapshotted and the rollout/compatibility boundary is still not executable.

## Consensus Strengths
- Multiple personas praised the core invariant that formulas declare required capability while the active Gas City binary, not the formula or `bd`, decides whether the host can satisfy it.
- Reviewers agreed that requirement normalization belongs in one `internal/formula` path and that callers should consume typed compile results, diagnostics, and accepted artifacts instead of raw `contract`, `version`, or metadata strings.
- The diagnostic direction is strong: one typed diagnostic shape projected to CLI, API/Huma, dashboard, generated TypeScript, and events is the right way to avoid contradictory operator guidance.
- The phased migration idea, dual-declaration bridge, and alias window are broadly pragmatic if they are backed by machine-readable evidence, old-reader compatibility fixtures, and reversible rollout units.
- Several personas agreed that pack revision, ref/SHA/content hash, and lockfile provenance are the correct reproducibility boundary, not formula-level `version` metadata.

## Critical Findings

### [Blocker] Canonical Design Artifact Is Missing Or Mis-Snapshotted
**Sources:** 03-parser-validation-matrix/Claude+Codex, 05-caller-integration-inventory/Claude+Codex, 07-pack-versioning-ecosystem/Claude+Codex, 08-docs-dx-terminology/Claude+Codex, 09-convergence-subset-reviewer/Claude+Codex, 10-future-capability-architect/Claude+Codex
**Issue:** Multiple personas report that `engdocs/design/formula-compiler-requirements.md` is absent from the review workspace or that attempt 94 captured `engdocs/proposals/formula-migration.md` instead. This prevents reviewers from confirming the authoritative design and creates artifact skew between Claude and Codex reviews.
**Required change:** Restore the tracked design document or repoint the source bead to the actual authoritative path, make the snapshot step fail when the configured design path is missing, and re-run review against the real artifact.

### [Blocker] Compatibility Boundary Is Not Fail-Closed
**Sources:** 02-contract-migration-guardian/Claude+Codex, 06-rollout-sequencing-reviewer/Claude+Codex, 07-pack-versioning-ecosystem/Claude+Codex, 10-future-capability-architect/Claude+Codex
**Issue:** The design does not yet define a complete fail-closed contract for legacy `contract`, `[requires]`, future requirement axes, unsupported fields, persisted roots, `GC_NATIVE_FORMULA=false`, `bd` fallback, and old-reader behavior. Newer or native-only formulas could be passed to older readers or fallback compilers and be miscompiled, ignored, or fail with incidental parser errors.
**Required change:** Add a normative compatibility contract covering omitted/default capability, legacy aliases, conflicts, unknown axes/keys, future capability requirements, fallback rejection before invoking `bd`, and persisted-root downgrade behavior, with stable diagnostics and zero-write fixtures.

### [Blocker] Caller And Runtime Inventory Is Incomplete
**Sources:** 01-compiler-boundary-invariant/Claude+Codex, 05-caller-integration-inventory/Claude+Codex, 09-convergence-subset-reviewer/Claude+Codex
**Issue:** The caller migration surface misses or underspecifies current compile/materialization paths, first-party prompt-template `gc bd mol` references, API/dashboard formula surfaces, graph-root predicates, convergence paths, dispatch/fanout, orders, source workflow scanners, and static-guard coverage. Existing raw metadata or fallback paths could continue making compiler-acceptance decisions outside `internal/formula`.
**Required change:** Generate a checked-in per-occurrence caller manifest and define one canonical `CompileResult`/accepted artifact API, one workflow-root predicate, repo-wide raw-consumer guards, dashboard TypeScript guards, and a migration plan that retires or fail-closes every bd-backed runtime path.

### [Blocker] Rollout Sequencing Can Break Old Readers And Shipped Examples
**Sources:** 02-contract-migration-guardian/Claude+Codex, 06-rollout-sequencing-reviewer/Claude+Codex, 08-docs-dx-terminology/Claude+Codex
**Issue:** The phase plan can be read as converting first-party files, examples, testdata, fixtures, or docs to `[requires]`-only before supported old binaries, pack floors, `bd` shell-outs, and public notice evidence are ready. That would violate the compatibility bridge and make rollback span parser, callers, docs, and packs.
**Required change:** Keep first-party graph formula files dual-declared during the alias window, split caller migration into reversible sub-phases, introduce and enforce `[pack] requires_gc` before requires-only conversion, and gate alias removal on machine-readable release/public-notice/legacy-root evidence.

### [Blocker] Parser And Validation Matrix Is Not Mechanically Complete
**Sources:** 03-parser-validation-matrix/Claude+Codex, 10-future-capability-architect/Claude+Codex
**Issue:** The parser contract lacks fixture-locked outcomes for presence-preserving `[requires]` decoding, empty or malformed values, wrong TOML/JSON types, duplicate declarations, unknown keys and axes, unsupported future versions, integer overflow, v2-only construct detection before filtering/materialization, and cross-caller zero-write behavior.
**Required change:** Specify a raw presence-preserving decode layer, diagnostic ordering, matrix axes, impossible predicates, count locks, generated tests, and caller-path fixtures across CLI, API, orders, convergence, dispatch/fanout, preview, TOML, and JSON if JSON formulas remain supported.

### [Major] Diagnostic Projection And Operator Semantics Are Under-Specified
**Sources:** 04-operator-diagnostics-gate/Claude+Codex, 01-compiler-boundary-invariant/Claude+Codex
**Issue:** Host capability is not yet an explicit compile input at every entry point, API/Huma diagnostic wire shape is not pinned, order failure payload migration conflicts with the current `NoPayload` registration, grouped occurrence counters are ambiguous under the append-only Event Bus invariant, and launch commands collapse formula diagnostics into generic exit `1`.
**Required change:** Pass explicit `HostCapabilities`/`CompileOptions` from city config, define one Huma/OpenAPI/generated TypeScript diagnostic schema, register typed order failure payloads, keep grouped counts in producer/read-model state or append new summary events, and pin CLI exit/JSON/quiet/warning semantics for automation.

### [Major] Deprecation And Alias-Removal Evidence Is Not Machine-Canonical
**Sources:** 02-contract-migration-guardian/Claude+Codex, 04-operator-diagnostics-gate/Claude+Codex, 07-pack-versioning-ecosystem/Claude+Codex
**Issue:** Markdown checklists, prose release gates, and vague external support notes are insufficient for retiring legacy `contract` or formula `version`. Reviewers require JSON artifacts for alias drain, active legacy roots, external support status, legacy version reports, warning cadence, and public notice.
**Required change:** Define JSON schemas, paths, generators, owners, `schema_version`/`generated_at` envelopes, gate consumers, public-notice dates, release-window rules, and repair commands such as `gc formula repair-root-artifact` before alias removal can be approved.

### [Major] Pack Provenance And Author Tooling Are Aspirational
**Sources:** 07-pack-versioning-ecosystem/Claude+Codex, 06-rollout-sequencing-reviewer/Claude+Codex
**Issue:** Pack revision, lockfile, content hash, dirty state, transitive source evidence, `RequiresGC`, and external pack inventory are treated as available without a prerequisite plan proving the resolver and lockfile can supply them or fail closed.
**Required change:** Add resolver/lockfile provenance prerequisites, exact `gc formula validate`/lint command contracts, JSON schemas, diagnostic codes, exit statuses, requirement-diff behavior, all-pack scanning, and behavior when provenance evidence is unavailable.

### [Major] Documentation And Terminology Are Not Release-Gated
**Sources:** 08-docs-dx-terminology/Claude+Codex, 07-pack-versioning-ecosystem/Claude+Codex, 10-future-capability-architect/Claude+Codex
**Issue:** Docs still risk teaching stale `contract`, `version`, implementation-name, or `GC_NATIVE_FORMULA=false` concepts, and terminology conflates formula `[requires]`, pack `requires_gc`, compiler capability, host capability, compiler implementation, schema/artifact version, and pack revision.
**Required change:** Add a file-by-file docs/examples migration inventory, canonical glossary/copy source, generated help/API/dashboard doc gates, stale-guidance CI with path/pattern exceptions, and reference docs for diagnostic/remediation strings before user-visible behavior ships.

### [Minor] Artifact Placement For Persona Syntheses Is Inconsistent
**Sources:** Global synthesis input inspection
**Issue:** The attempt 94 persona-synthesis directory was empty, while the closed persona-synthesis beads recorded their outputs under `.gc/design-reviews/ga-j195cx/attempt-1/persona-syntheses/`. All ten required persona syntheses were present there and were used for this synthesis, but the mismatch reduces audit clarity.
**Required change:** Fix the workflow attempt directory propagation for nested retry/persona synthesis steps or record the actual persona-synthesis directory in a canonical attempt output field.

## Disagreements
- Several Claude reviews saw a newer or more complete compiler-requirements snapshot while several Codex reviews saw the older formula-migration artifact or a missing design file. Assessment: this is artifact skew, not a substantive disagreement; it is itself a blocker.
- 04-operator-diagnostics-gate returned `approve-with-risks` while all other personas except none returned `block`. Assessment: operator diagnostics are directionally sound but depend on the same host-capability, typed-wire, event, and fixture contracts that other personas found incomplete.
- Some reviewers considered existing normalized requirement direction structurally sound, while others considered the capability boundary missing. Assessment: if the compiler-requirements design owns that boundary, the migration/rollout design must normatively reference it and prove every caller enforces it; otherwise implementers do not have one contract.
- Claude and Codex differed on severity for rollout sequencing and future capability risks. Assessment: compatibility and old-reader failures are release-blocking because they can strand shipped packs, persisted roots, or operator rollbacks.
- Reviewers disagreed on whether convergence's old subset parser is a live production write path. Assessment: the design should anchor to the current call graph, delete or quarantine dead subset paths with static guards, and focus live migration on `molecule.Cook`/accepted artifact flow.

## Missing Evidence
- The authoritative current contents of `engdocs/design/formula-compiler-requirements.md`, or a corrected source path for the review.
- Proof that the design-review snapshot step fails when the configured design document is absent.
- A generated caller/raw-consumer manifest covering Go code, prompt templates, formulas, examples, dashboard TypeScript, API/generated clients, convergence, fanout, orders, root scanners, and `bd` materialization paths.
- A single `CompileResult`/accepted artifact definition and compile entry point that carries resolved source provenance, host capabilities, normalized requirements, diagnostics, compiled steps, and runtime variables.
- Machine-readable release evidence artifacts for alias removal, active legacy roots, external pack support, public notice, minimum binary floors, and legacy `version` reports.
- Cross-version fixtures proving old supported readers and any active `bd` fallback tolerate dual declarations while still honoring legacy `contract = "graph.v2"`.
- Matrix fixtures for malformed `[requires]`, unknown future axes, JSON formulas if retained, disabled-host zero-write behavior, v2-only construct detection before materialization, convergence behavior, and fallback rejection before invoking `bd`.
- One Huma/OpenAPI/generated TypeScript diagnostic schema, typed event payload registrations, and append-only-safe grouped occurrence read model.
- Resolver/lockfile provenance support for pack refs, SHAs, content hashes, dirty state, transitive sources, and `requires_gc`.
- File-by-file docs/examples/tutorial/generated-surface migration ownership and stale-guidance CI matcher fixtures.

## Recommended Changes
1. Restore or repoint the canonical design artifact and make snapshot failure explicit; then re-run the review against the intended design document.
2. Add the normative formula compatibility contract for `[requires]`, legacy `contract`, future axes, persisted roots, fallback mode, unknown keys, old readers, and zero-write diagnostics.
3. Generate and check in a caller/raw-consumer inventory, then define one canonical `CompileResult`/accepted artifact API and one workflow-root predicate consumed by all callers.
4. Split the rollout into reversible sub-phases, keep shipped graph formulas dual-declared during the alias window, and gate requires-only conversion on `[pack] requires_gc`, old-reader fixtures, and machine-readable evidence.
5. Build the parser/validation matrix as generated fixtures with count locks and caller-path coverage, including JSON behavior or a decision to retire JSON formulas.
6. Pin diagnostics end to end: host capabilities as explicit compile input, one API/dashboard schema, registered event payloads, append-only grouping semantics, CLI exit/JSON behavior, and canonical remediation strings.
7. Define JSON release artifacts and commands for alias drain, active root repair, legacy version reporting, external support, public notice, and minimum binary floors.
8. Add resolver/lockfile provenance prerequisites and the exact pack-author validation/lint contract before relying on pack revision as the reproducibility boundary.
9. Add a docs/examples migration inventory, glossary, stale-guidance CI, and generated-surface gates before any user-visible `[requires]` or deprecation diagnostic ships.
10. Fix nested workflow attempt-directory propagation so persona syntheses land in the current outer attempt directory or are referenced by a canonical output manifest.
