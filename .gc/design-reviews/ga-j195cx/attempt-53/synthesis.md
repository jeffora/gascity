# Design Review Synthesis

## Overall Verdict: block

The design is pointed in the right direction: formulas declare capabilities, the active Gas City host decides whether it can satisfy them, and callers are meant to consume typed compiler results instead of duplicating TOML or metadata interpretation. It is not ready to implement. Multiple persona syntheses block on foundational contracts for diagnostics, rollout sequencing, and convergence/accepted-artifact reuse; those gaps can recreate the exact duplicated decision logic the design is trying to remove.

## Consensus Strengths
- Multiple reviewers praised the central compiler-boundary direction: requirement parsing and capability checks should live in `internal/formula`, with CLI, API, sling, orders, convergence, and dashboard consuming typed outputs.
- The pack-revision identity model is sounder than formula-level `version`: pack source/ref/revision, lockfile identity, binding identity, content hash, and accepted artifact references are the right reproducibility boundary.
- The design correctly treats docs, examples, generated API/dashboard artifacts, stale-guidance checks, and user-visible diagnostics as one hard rollout gate rather than optional follow-up work.
- The legacy `contract = "graph.v2"` migration is evidence-based in intent, preserving old formulas through an alias window instead of forcing a flag-day syntax change.
- Reviewers agreed that a matrix-driven fixture corpus is the right control for parser edge cases, durable writer preflights, projection parity, and old-reader compatibility.
- The future-capability direction is acceptable if capability evolution is made monotonic, additive, and fixture-proven.

## Critical Findings

### [Blocker] Diagnostics Lack A Single Typed Attribution Contract
**Sources:** Marta Hidalgo; reinforced by Elias Vega, Priya Zielinski, and Yuki Patel
**Issue:** Operator-facing diagnostics require both formula requirement attribution and host-capability attribution, but the core diagnostic model does not yet carry both as typed data. CLI, Huma/API payloads, generated TypeScript, dashboard state, and event payloads would have to reconstruct remedies from strings or side channels. The background grouping and warning-suppression rules also conflict on storage owner, grouping key, occurrence updates, restart behavior, and `OnceKey` shape.
**Required change:** Add one reusable typed diagnostic attribution model that carries primary source, requirement source, host source, source kind, path, key, and value. Derive CLI output, API responses, registered event payloads, dashboard projections, and warning keys from that object, with one disabled-host golden fixture covering direct CLI, API-routed CLI, HTTP body/status, generated TypeScript/dashboard state, and applicable events.

### [Blocker] Convergence Cannot Yet Project From Accepted Artifacts Alone
**Sources:** Felix Berger; reinforced by Nadia Sorenson, Yuki Patel, and Priya Zielinski
**Issue:** The accepted-artifact contract does not prove convergence can populate runtime vars, evaluate prompt data, retry policy, source attribution, requirements, and step facts without reopening formula source or preserving the legacy subset parser. Artifact reuse identity is also ambiguous across formula name, content hash, vars/options/search-path hashes, host capabilities, config generation, and binding identity. The current zero-write promise is not yet enforceable at all convergence durable-write boundaries.
**Required change:** Define a versioned, hash-checked accepted-artifact or compiler-owned projection snapshot that is sufficient for `ProjectAcceptedFormula` without source access. Pin the package data flow through `CompileWithResult`, `AcceptCompileResult`, `ProjectAcceptedFormula`, `ValidateProjection`, and durable writes; delete or strictly constrain the convergence shim; add missing-source, voluntary-requirement, host-downgrade, concurrent repair, and zero-write fixtures.

### [Blocker] Rollout Sequencing Leaves Compatibility Gaps
**Sources:** Lena Driscoll; reinforced by Yuki Patel, Avery Brooks, Elias Vega, and Saoirse Raman
**Issue:** The rollout still allows stale `engdocs/proposals/formula-migration.md` guidance to tell operators to use the old `GC_NATIVE_FORMULA=false -> Store.MolCook` rollback path. It also does not name the phase that writes `[pack] requires_gc` for first-party graph packs before resolver/import enforcement, does not make first-party dual declaration a cross-phase invariant, and does not specify the alias-window clock, release-captain ownership, active-root drain/waiver rule, or Phase 7 rollback clock behavior.
**Required change:** Supersede the old migration proposal before Phase 1 code lands. Add a named phase for first-party `[pack] requires_gc` floor writes, restate dual declaration as a Phases 2-7 invariant, add Phase 3 ordering and rollback controls, pin the two-minor-release and 60-day clock start, and require an active-root report, drain criterion, or explicit waiver before requires-only conversion.

### [Blocker] In-Flight Graph Workflow Semantics Are Not Crisp
**Sources:** Yuki Patel; reinforced by Felix Berger, Nadia Sorenson, and Priya Zielinski
**Issue:** The design says artifact-stamped roots can continue graph-specific writes after `[daemon] formula_v2 = false`, but operators may reasonably expect graph behavior to stop, and `ValidateAcceptedArtifact` can also fail on the same host-capability disagreement. The repair path for legacy roots is not scheduled, and caller migration is not mechanically enforceable because raw predicates are grouped by file instead of by each actual predicate/filter site.
**Required change:** Split validation into current-host validation for new or changed compiles and persisted-artifact reuse for same-identity roots. State operator-visible behavior for new compiles, artifact-stamped roots, legacy-only roots, fanout, retry, scope-check, workflow-finalize, and convergence while disabled. Schedule `gc formula repair-root-artifact` with dry-run JSON, diagnostics, idempotency, zero-write failure tests, and behavior for missing source, downgraded host, unsupported artifact schema, and already-stamped roots. Generate a per-occurrence caller manifest and CI guard.

### [Major] Parser And Fixture Contracts Are Still Too Loose
**Sources:** Priya Zielinski; reinforced by Ibrahim Park and Elias Vega
**Issue:** `RawRequirementField` requires source text, TOML/JSON shape, line, and column before typed decoding, but the selected BurntSushi/toml path is not specified as providing stable per-key positions or source slices for valid fields. Edge cases such as signed values, leading zeros, overflow, Unicode lookalikes, inline/nested/dotted tables, duplicate keys, contract spelling variants, and disabled-by-condition v2 constructs are named but not pinned to exact diagnostic rows. JSON scope and duplicate-member handling are undecided.
**Required change:** Specify the raw-source extraction implementation as an API contract, including duplicate parse-error boundaries and proof that typed decoding cannot bypass raw capture. Add exact fixture rows and expected diagnostics for all named grammar and raw-shape cases. Decide whether JSON formulas are in scope; if they are, add a JSON shape axis and duplicate-detecting raw pass. Require count locks or generated-count expressions for every normative suite, split by row kind.

### [Major] Contract Migration And External Pack Evidence Are Incomplete
**Sources:** Elias Vega and Saoirse Raman; reinforced by Avery Brooks
**Issue:** Alias removal is evidence-based in prose, but it lacks a single executable gate that blocks on first-party `legacy_only > 0`, first-party `dual_declared > 0`, unsupported future requirements, or unreadable inventory. First-party versus external pack classification is load-bearing but implicit. The offline `--pack-source --offline` behavior, external author migration walkthrough, external-support expiration evidence, release artifact schemas, tag/branch identity, and legacy `version` diagnostics are not fully pinned.
**Required change:** Add a deterministic first-party/external classification predicate with fixture-locked tests. Define one alias-removal command or CI job and cite it in removal criteria, rollout gates, and release checklists. Seed `formula-compiler-compatibility.yaml`, `formula-compiler-min-floor.json`, external-support, and stale-guidance artifacts with conservative blocking defaults and negative fixtures. Add an end-to-end external pack migration fixture and a byte-level offline acquisition matrix.

### [Major] Documentation And Terminology Still Teach Old Behavior
**Sources:** Avery Brooks; reinforced by Lena Driscoll, Yuki Patel, and Saoirse Raman
**Issue:** `docs/reference/formula.md` still presents `version` as a normal top-level key, `engdocs/architecture/formulas.md` still describes the old `bd`/`MolCook` production path, and `engdocs/proposals/formula-migration.md` still frames `GC_NATIVE_FORMULA=false` as a runtime rollback. The design also risks confusing formula `[requires] formula_compiler` with pack `[pack].requires_gc`, and there is no checked-in first-party formula inventory covering built-in packs, examples, tutorials, and fixtures.
**Required change:** Update or explicitly supersede the stale docs in the same PR stack that exposes diagnostics. Rewrite the formula reference so canonical examples omit `version`, add a glossary and side-by-side TOML example for formula `[requires]` and pack `[pack].requires_gc`, add a first-party formula inventory under `docs/release/`, and make stale-guidance CI reject canonical `version =` examples and unresolved release placeholders.

### [Major] Future-Capability Invariants Need To Be Binding
**Sources:** Ibrahim Park; reinforced by Priya Zielinski and Nadia Sorenson
**Issue:** The v0 `[requires]` direction is sound, but additive-only capability evolution, frozen released-axis grammar, same-axis versus new-axis schema-version rules, multi-axis construct capability representation, and old-reader behavior for future compiler capability plus future axes are still not binding enough.
**Required change:** Add a Forward Compatibility invariants section stating that compiler capability evolution is additive only, released constructs are not removed or semantically redefined, released axis byte grammars are frozen, richer future grammar becomes a new typed axis, and `RequirementSource` is diagnostic/provenance-only. Define schema-version rules, generalize construct capability for multi-axis constraints, and add old-reader cross-product fixtures.

### [Minor] Artifact, Suppression, And Remediation Details Need Cleanup
**Sources:** Marta Hidalgo, Saoirse Raman, Avery Brooks, and Ibrahim Park
**Issue:** Warning suppression has multiple key definitions, background diagnostic grouping has no durable state model, `gc.formula_compile_artifact` spillover lacks a deterministic threshold, and some remediation strings look copy-pasteable while not being valid TOML.
**Required change:** Use one canonical warning key, define grouping storage and reset behavior, choose an artifact spillover boundary, and make remediation strings either valid TOML snippets or plain prose.

## Disagreements
- Marta Hidalgo: Claude returned `approve-with-risks`; Codex returned `block`. Assessment: the global synthesis follows the block because both reviews converge on source attribution and surface parity being foundational, even if they label severity differently.
- Lena Driscoll: Claude returned `block`; Codex returned `approve`. Assessment: the block stands because rollout sequencing, stale rollback guidance, pack floor writes, and rollback clocks are exactly this persona's mandate and must be operational before implementation planning.
- Felix Berger: Claude returned `approve-with-risks`; Codex returned `block`. Assessment: the block stands because the accepted-artifact/projection gap can keep convergence dependent on raw source or a subset parser, undermining the core design.
- Yuki Patel adopted `approve-with-risks` despite several blocker-level findings. Assessment: those findings reinforce the global blockers around accepted-artifact reuse, host downgrade semantics, repair, stale docs, and caller-manifest enforcement.
- Priya Zielinski and reviewers disagree only on framing for JSON scope and count-lock maturity. Assessment: this is not a blocker if the design makes a binding JSON scope decision and fixtures the row-kind locks before implementation.
- Ibrahim Park noted review-time skew: some items Codex saw as missing may be present in the design Claude reviewed. Assessment: do not block solely on skew, but require explicit invariants and fixtures wherever reviewers still found ambiguity.
- Elias Vega reviewers differ on branch/tag-pinned external packs. Assessment: the design must state whether mutable refs count as pinned and document a lockfile escape hatch if they do not.

## Missing Evidence
- No Gemini review artifact was present for the current persona syntheses; this is allowed by the workflow settings but leaves only Claude and Codex perspectives.
- No raw-source extraction strategy proves TOML and optional JSON requirement fields can retain source text, line, column, and raw shape before typed decoding.
- No accepted-artifact schema or compiler-owned projection snapshot proves convergence can run from persisted compiler output when formula source is unavailable.
- No generated caller or durable-writer manifest proves every graph-workflow predicate, durable write path, repair path, continuation path, and convergence path is represented.
- No disabled-host golden fixture pins CLI stderr/exit, API-routed CLI behavior, HTTP status/body, generated TypeScript/dashboard state, registered events, warning co-emission, and remediation bytes together.
- No release artifact seed files or negative fixtures prove alias-removal, min-floor, compatibility, external-support, stale-guidance, and release-checklist gates fail closed.
- No first-party formula inventory covers built-in packs, `.gc/system/packs`, examples, tutorials, fixtures, and Ralph demo formulas.
- No offline external pack acquisition matrix or worked third-party migration example proves `--pack-source --offline`, SHA/tag/branch refs, requirement diffs, republish, and consumer lockfile updates compose.
- No active-root drain report, waiver policy, or repair inventory proves legacy workflow roots are safe before requires-only conversion.
- No old-reader cross-product fixtures prove future compiler capability plus a future axis remains observable-only and refuses graph writes consistently.

## Recommended Changes
1. Define the typed diagnostic attribution model and make CLI, API, events, dashboard, warning keys, and canonical remediation fixtures derive from it.
2. Define the accepted-artifact/projection schema, transaction ordering, storage substrate, and artifact reuse identity table, then fixture convergence projection without source access and zero-write failures.
3. Supersede stale migration docs before Phase 1, add first-party `[pack] requires_gc` floor writes, restate dual declaration as a cross-phase invariant, and pin rollout clock/rollback rules.
4. Specify host-downgrade semantics for new compiles, accepted roots, legacy-only roots, fanout, retry, scope-check, workflow-finalize, and convergence; add `gc formula repair-root-artifact`.
5. Generate caller and durable-writer manifests from code, require per-occurrence rows, and add deletion guards for legacy raw-decision helpers, subset parsers, and global formula-v2 capability reads.
6. Specify raw requirement source extraction and expand grammar, raw-shape, JSON-scope, contract, disabled-condition, and count-lock fixture matrices.
7. Add the first-party/external classification predicate, alias-removal CI gate, release artifact schemas, conservative seed files, and external-support expiration evidence.
8. Update formula reference, architecture, migration proposal, generated help/API/dashboard docs, glossary, examples, and stale-guidance checks in the same PR stack that exposes diagnostics.
9. Add future-capability invariants, schema-version rules, generalized multi-axis construct capability representation, and old-reader future-axis fixtures.
