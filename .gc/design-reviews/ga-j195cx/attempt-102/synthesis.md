# Design Review Synthesis

## Overall Verdict: block

Seven of ten persona syntheses returned `block`, so the global verdict is `block` by worst-verdict-wins. The design has the right overall direction - formula requirements owned by `internal/formula`, host capabilities decided by the active binary, typed compile artifacts, and pack revision as the ecosystem boundary - but the current text still leaves several release gates, migration boundaries, and user-facing contracts ambiguous enough that implementation would force ad hoc decisions in code.

## Consensus Strengths
- Multiple personas agree that formulas should declare minimum requirements while the active Gas City binary decides whether host capabilities satisfy them; the design is directionally aligned with Zero Framework Cognition.
- `internal/formula` is the right ownership boundary for requirement parsing, compatibility normalization, v2-only construct detection, diagnostics, compile results, and accepted artifacts.
- The pack-revision direction is sound: pack version, refs, SHAs, and provenance are the reproducibility boundary, not formula-level artifact semver.
- The future capability posture is broadly right: unknown requirement axes should fail closed, and omitted, empty, and explicit default compiler requirements should share runtime behavior.
- Reviewers generally accepted the need for typed diagnostics, zero-write fixtures, migration evidence, and stale-guidance checks; the remaining concerns are about making those contracts executable and internally consistent.

## Critical Findings

### [Blocker] Alias removal and rollout gates are internally inconsistent
**Sources:** 02-contract-migration-guardian/Claude, 02-contract-migration-guardian/Codex, 06-rollout-sequencing-reviewer/Claude, 06-rollout-sequencing-reviewer/Codex, 07-pack-versioning-ecosystem/Claude, 07-pack-versioning-ecosystem/Codex
**Issue:** The release gates for dual declarations, alias drain, and alias removal are not coherent enough to execute safely. Phase 8 can be read as requiring zero first-party dual-declared formulas before the phase that removes first-party dual declarations. The alias-removal schema omits artifacts named elsewhere in the rollout, and the alias-window clock can diverge across min-floor, alias-window-start, external-support, and drain evidence. Runtime accepted-alias counts are also not tied to a durable producer, retention rule, or reset semantics.
**Required change:** Split the rollout into a Phase 8 conversion gate and a Phase 9 alias-removal gate. Define one authoritative alias-window artifact and require related reports to embed its digest. Collapse alias-removal evidence into one command/schema/exit table that fails closed for missing, stale, conflicting, or rollback-ineligible evidence. Make accepted-alias counting durable or remove it from the gate.

### [Blocker] Caller inventory depends on ignored runtime state
**Sources:** 05-caller-integration-inventory/Claude, 05-caller-integration-inventory/Codex, 01-compiler-boundary-invariant/Codex
**Issue:** The design makes `.gc/system/packs` a mandatory CI scan input even though `.gc/` is ignored and absent from clean checkouts. That makes the no-bypass caller manifest unstable: it can fail in CI, miss materialized pack prompts, or depend on local runtime state. The same area has broad allowlist examples and a caller inventory that mixes intended future APIs with current symbols.
**Required change:** Remove `.gc/system/packs` as a mandatory CI input, or define a deterministic fixture-generation step from tracked sources and record a checked-in manifest artifact. Make local runtime `.gc` scans optional operator evidence only. Replace broad wildcard allowlists with exact file plus symbol or line-pattern constraints, and reconcile current versus net-new caller symbols in the manifest.

### [Blocker] Documentation and terminology still conflate distinct requirement surfaces
**Sources:** 08-docs-dx-terminology/Claude, 08-docs-dx-terminology/Codex, 07-pack-versioning-ecosystem/Claude, 09-convergence-subset-reviewer/Codex
**Issue:** The glossary and stale-guidance plan still blur formula `[requires]`, pack `[[pack.requires]]`, and pack `[pack].requires_gc`. The stale-guidance coverage also omits known formula-bearing paths even though the design claims broad inventory coverage. Several docs still teach legacy convergence or formula terminology, and the design path itself was reported as untracked by the persona synthesis.
**Required change:** Split glossary and common-confusion rows for formula `[requires]`, pack `[[pack.requires]]`, and pack `[pack].requires_gc`. Normalize "formula v2", "graph workflow", "compiler capability 2", and `[daemon] formula_v2` usage. Extend stale-guidance or inventory checks to formula testdata, tutorial golden sources, Ralph demo formulas, docs/reference/formula.md, PackV2 author docs, and architecture docs, with deliberate exemptions and expiry phases.

### [Major] The compile/preflight authority boundary is not pinned
**Sources:** 01-compiler-boundary-invariant/Claude, 01-compiler-boundary-invariant/Codex, 05-caller-integration-inventory/Claude, 09-convergence-subset-reviewer/Claude, 09-convergence-subset-reviewer/Codex
**Issue:** The design still does not name one canonical compile/preflight API that carries resolved source, active host capabilities, diagnostics, compiled projection, provenance, and accepted artifact authority. Legacy `Compile(...)`, bare `*Recipe`, `Recipe.GraphWorkflow`, raw metadata predicates, and direct host-capability reads remain possible bypasses for durable or behavioral graph decisions. Workflow-root classification also lacks one named predicate with canonical-first behavior and conflict semantics.
**Required change:** Define the canonical `internal/formula` API and its result/accepted-artifact contract. Make legacy recipe and graph-state surfaces unsuitable for durable writer decisions through structural changes or static guards. Name the shared workflow-root predicate, its package, input type, canonical/legacy behavior, and fail-closed conflict handling.

### [Major] Operator diagnostics lack executable wire, event, and store contracts
**Sources:** 04-operator-diagnostics-gate/Claude, 04-operator-diagnostics-gate/Codex
**Issue:** The diagnostic contract promises parity across CLI, API, dashboard, SSE/events, and order failures, but key pieces are not implementable as written. Presence-sensitive fields can be dropped by wire tags, grouped diagnostic singleton/CAS behavior is not guaranteed by the current Task Store contract, `order.failed` compatibility is undefined, mutable warning cadence is mixed into immutable accepted artifacts, and no concrete rollup route/event/CLI surface is named.
**Required change:** Define a presence-aware typed diagnostic wire contract and regenerate OpenAPI/TypeScript/dashboard fixtures from it. Either promote grouped diagnostic singleton/CAS behavior into an explicit Task Store capability with backend conformance tests or redesign the grouping store. Pin the order-failure event migration, correlation keys, payload registrations, burst-budget accounting, rollup read surface, dashboard-less status surface, and launch-command exit semantics.

### [Major] Pack ecosystem compatibility and provenance are under-specified
**Sources:** 07-pack-versioning-ecosystem/Claude, 07-pack-versioning-ecosystem/Codex, 06-rollout-sequencing-reviewer/Claude, 06-rollout-sequencing-reviewer/Codex
**Issue:** External SHA-pinned packs can still break because the compatibility corpus preserves only byte-exact `contract = "graph.v2"` while the current compiler accepts case and whitespace variants. Resolver-to-compiler provenance is load-bearing but not a named rollout deliverable. Standalone `gc formula validate --pack-path` and `--pack-source --ref` lack a synthetic direct-pack binding identity. Old-reader pack-load compatibility for `[pack] requires_gc` is not proven before rollout steps rely on it.
**Required change:** Preserve every currently accepted legacy `contract` spelling as a deprecated alias during the compatibility window, or add explicit external-support evidence before rejecting any spelling. Make resolver/import provenance plumbing an explicit Phase 2 deliverable. Define direct-pack validation identity and fixtures. Add full old-reader pack-load fixtures for `[pack] requires_gc`, dual graph declarations, and actual floor values before first-party pack floors are written.

### [Major] Parser and validation matrices still have raw-shape gaps
**Sources:** 03-parser-validation-matrix/Claude, 03-parser-validation-matrix/Codex, 10-future-capability-architect/Claude
**Issue:** The parser matrix does not yet pin behavior for nested misplaced requirement tables, dotted-key TOML syntax, dotted/bracketed collisions, inline-table conflicts, combined-defect precedence, `condition_disabled` v2 constructs, JSON loader policy, unknown axes paired with satisfied known requirements, integer boundaries, or `CompilerCapability(0)`. The design also conflicts on whether explicit `>=1` author intent survives accepted-artifact deduplication.
**Required change:** Add named suites and count-locked fixture rows for combined defects, `condition_state`, nested non-top-level requirement tables, dotted-key syntax, unknown axes, duplicate/colliding shapes, and JSON caller preflight. Resolve explicit-`>=1` provenance by naming the durable authority, and state whether `ContentHash` hashes raw bytes or canonical post-normalization content.

### [Major] Convergence migration is not anchored to the live write path
**Sources:** 09-convergence-subset-reviewer/Claude, 09-convergence-subset-reviewer/Codex, 01-compiler-boundary-invariant/Claude
**Issue:** The convergence section implies the legacy subset parser is a live bypass, while the persona synthesis reports that production convergence already flows through `cmd/gc/convergence_store.go:pourWisp -> molecule.Cook -> internal/formula`. The real risk is the durable writer boundary and projection validation. Current evaluate-prompt checks, prompt path safety, accepted-artifact handoff, v2 convergence policy, and convergence diagnostics are not mapped tightly enough.
**Required change:** Anchor the migration to the current production call graph. Delete or quarantine retired subset parser paths with guards. Name the convergence preflight helper/API and durable-writer allowlist, map every evaluate-prompt validation to `ValidateProjection` diagnostics or retire it explicitly, and decide whether convergence formulas may use `[requires] formula_compiler = ">=2"` with v2 constructs.

### [Minor] Workflow artifact hygiene is ambiguous for this attempt
**Sources:** persona synthesis bead metadata for attempt 102
**Issue:** The current attempt directory `.gc/design-reviews/ga-j195cx/attempt-102` did not contain the ten persona synthesis files. The closed persona-synthesis beads for `gc.attempt=102` recorded nonempty output paths under `.gc/design-reviews/ga-j195cx/attempt-1/persona-syntheses/*.md`, and this synthesis used those recorded outputs. That is sufficient to synthesize the review, but it makes the reviewed snapshot identity harder to audit.
**Required change:** Fix the workflow output-path calculation so persona syntheses for attempt N are written under `attempt-N/persona-syntheses`, or have the global synthesis step copy/manifest the exact source paths it consumed.

## Disagreements
- Several personas had Claude at `approve-with-risks` and Codex at `block`. I side with `block` where Codex identified concrete fail-open or contradictory contracts, especially alias-removal gating, ignored `.gc` scan inputs, stale-guidance coverage, and pack compatibility. Claude's findings usually reinforced the same risks at a different level of detail.
- Parser validation, convergence, and future capability personas returned `approve-with-risks`, not `block`. I treat their findings as required design work but not independently global-blocking; the global block is driven by the seven blocked personas.
- Some reviewers praised the broad architecture while blocking on executable contracts. My assessment is that the architecture should not be reopened, but the design must pin APIs, schemas, gates, and rollout evidence before implementation starts.
- The persona synthesis artifacts were stored under `attempt-1` despite `gc.attempt=102`. I treated the bead metadata output paths as authoritative source artifacts, but this should be fixed in the workflow so future attempts are auditable from the attempt directory alone.

## Missing Evidence
- Kimi 2.6 reviews were absent; the attempt was run with Claude and Codex sources only.
- Old-reader proof that supported binaries tolerate dual-declared formulas with `[requires]` and first-party packs with `[pack] requires_gc`.
- A unified alias-window, alias-drain, min-floor, external-support, active-root, and rollback-aware release-gate schema with digest linkage and freshness checks.
- Durable accepted-alias counter semantics, or a decision to remove runtime counts from release gates.
- A deterministic caller manifest from tracked sources or checked-in generated fixtures, without mandatory ignored `.gc` runtime state.
- Canonical compile/preflight API signatures, accepted-artifact validation boundaries, and static guards for raw metadata, legacy recipes, graph-state reads, and host-capability shims.
- Presence-aware diagnostic wire fixtures, Task Store grouping conformance tests, typed event migration fixtures, rollup surfaces, and launch-command exit semantics.
- Resolver/import provenance plumbing, direct-pack validation identity, and old-reader complete-pack fixtures.
- Parser matrix rows for nested requirement tables, dotted-key syntax, condition-disabled constructs, unknown axes, collision precedence, JSON loader paths, integer boundaries, and `CompilerCapability(0)`.
- Convergence call-site maps, projection diagnostics, durable-writer allowlists, and disabled-host zero-write fixtures for the chosen convergence v2 policy.
- Stale-guidance coverage for formula-bearing docs, examples, tutorials, testdata, and generated fixtures, plus evidence that the design document is tracked as the canonical artifact.

## Recommended Changes
1. Rewrite the rollout gates first: split Phase 8 conversion from Phase 9 alias removal, define one alias-window artifact, and attach old-reader proof before any first-party source or pack metadata edits land.
2. Define the canonical `internal/formula` compile/preflight and accepted-artifact API, then make every durable writer consume that authority instead of raw metadata, legacy recipes, or direct host flags.
3. Replace mandatory `.gc/system/packs` scanning with a deterministic tracked-source manifest or checked-in generated fixture, and tighten raw-consumer allowlists to exact file plus symbol or line constraints.
4. Fix the terminology contract by separating formula `[requires]`, pack `[[pack.requires]]`, and pack `[pack].requires_gc`, then wire stale-guidance checks to every formula-bearing docs/test/tutorial path.
5. Make operator diagnostics executable: presence-aware wire types, registered typed events, grouping-store capability or redesign, rollup/status surfaces, and clear exit-code semantics.
6. Protect external packs by preserving every currently accepted legacy `contract` spelling during the alias window, defining direct-pack validation identity, and making resolver/import provenance a named rollout deliverable.
7. Expand parser and matrix fixtures for TOML raw shapes, misplaced requirements, dotted syntax, unknown axes, condition-disabled constructs, JSON loader policy, and future-capability boundary cases.
8. Anchor convergence work to the live `pourWisp -> molecule.Cook -> internal/formula` path, then map prompt/path validation and v2 convergence policy into projection diagnostics and zero-write fixtures.
9. Resolve future-capability provenance and identity: explicit `>=1`, `ContentHash`, requirement-axis root metadata, integer-boundary guards, and the allowed subject matter for future axes.
10. Fix the design-review workflow artifact path bug so attempt 102 persona syntheses are discoverable under `attempt-102/persona-syntheses` or are explicitly manifest-listed by the global synthesis.
