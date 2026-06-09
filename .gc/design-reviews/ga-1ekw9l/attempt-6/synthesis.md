# Design Review Synthesis

## Overall Verdict: block

Worst-verdict-wins yields `block`: nine persona syntheses block and one persona approves with risks. The implementation plan now conforms to the required `gc.mayor.implementation-plan.v1` top-level shape, but the design is still not ready for full decomposition because required loader, behavior-evidence, public-pack, role-neutrality, doctor/runtime-state, bootstrap, and rollout gates remain too implicit or depend on missing external artifacts.

## Consensus Strengths

- The high-level target is correct: Core becomes the required Gas City runtime pack, Maintenance retires as an active standalone pack, and Gastown behavior moves to an explicit public pack.
- Multiple personas praised the fail-closed posture for missing, corrupt, shadowed, stale, or retired pack sources instead of silently loading fallback behavior.
- The plan's direction toward generated proof artifacts is sound: behavior manifests, role-surface scans, acceptance proof matrices, pin ledgers, diagnostics schemas, and packcompat transcripts are the right evidence shape.
- The implementation-plan artifact matches the required Mayor implementation-plan front matter and top-level sections: Summary, Current System, Proposed Implementation, Data And State, Testing, Rollout And Recovery, and Open Questions.
- The doctor and runtime-state sections have the right safety instincts: non-interactive repair, idempotence, atomic publication, refusal on unsafe state, TOML preservation, and operator-visible diagnostics.

## Critical Findings

### [Blocker] Required-pack participation and loader authority are not yet provable

**Sources:** 01-required-core-loading, 06-pack-boundary-containment, 07-bootstrap-fixture-isolation, 09-rollout-decomposition-gates
**Actionability:** document-fixable
**Issue:** Required Core/provider-pack participation is still described around paths, names, digests, and post-resolution checks rather than one resolver-owned trusted provenance API. The plan does not bind Gate 1 file-set validation to Gate 2 participation validation for the exact trusted target, and it does not define a canonical runtime loader plus AST/type-aware bypass guard or runtime participation stamp for all behavior-driving paths.
**Required change:** Define the trusted required-pack descriptor flow from `internal/systempacks` into `internal/config`, the typed participation records returned during resolution, and the exact Gate 1/Gate 2 binding. Name the canonical loader APIs, no-refresh/live-reload failure behavior, runtime snapshot or guard object, behavior-changing entrypoints, bypass inventory/scanner, and tests for copied Core, user-imported `core`, manual Core paths, provider-pack selection, stale/corrupt materialization, and API/controller bypasses.

### [Blocker] Behavior preservation and public-pack proof are not executable acceptance gates

**Sources:** 02-behavior-evidence-chain, 05-public-pack-pin-cache, 06-pack-boundary-containment, 09-rollout-decomposition-gates
**Actionability:** mixed
**Issue:** The plan depends on a behavior-evidence chain, public Gastown manifests, packcompat, pin ledgers, ownership rows, and exact public-pack cache proof, but the contracts are still future-tense or absent. Trigger-level behavior identity, one-row-per-trigger extraction, evidence-class rules, old/new witness mapping, generated freshness tests, exact pinned public checkout proof, and cross-repo bootstrap staging are not concrete enough for decomposition.
**Required change:** Define the Behavior Evidence contract end to end: manifest schema, generator command, row identity, source-kind extractors, digest/freshness fields, evidence classes, approval records, old/new witnesses, source-deletion gate, public-pack reconciliation, and packcompat command/fixture/failure semantics. If public `gascity-packs` commits, manifests, ownership rows, or transcripts do not yet exist, mark them as external prerequisites and block dependent Gas City slices until they do.

### [Blocker] Public Gastown pin/cache authority is sequenced too late

**Sources:** 05-public-pack-pin-cache, 06-pack-boundary-containment, 09-rollout-decomposition-gates, 10-operator-docs-schema
**Actionability:** document-fixable
**Issue:** Slice 2 consumes a subpathed public `//gastown` pin before subpath-aware ordinary cache identity, lock/cache provenance, digest validation, and synthetic-alias rejection are required. Current source+commit cache identity can conflate subpaths, and stale synthetic aliases can still satisfy public-looking imports unless rejection lands before the first public pin is consumed.
**Required change:** Move subpath-aware proof identity and synthetic-alias rejection into the first slice that consumes the public Gastown pin. Specify lock/cache schema migration, old-schema behavior, `RepoCacheKey` or proof identity, digest algorithms and scopes, read-hit validation, validation markers if used, atomic cache publication, network-disabled cache tests, compatibility versus activation pins, rollback/downgrade rows, and one executable pin-coherence gate spanning `PublicGastownPackVersion`, ledgers, fresh init, lock/cache provenance, pack digest, and behavior-manifest digest.

### [Blocker] Role neutrality and pack-boundary containment are not a single enforced contract

**Sources:** 04-role-neutrality-zfc, 06-pack-boundary-containment, 10-operator-docs-schema
**Actionability:** document-fixable
**Issue:** The design still allows broad `dog` occurrences in Core pack configuration, does not fully de-role warmup recipients, binding-prefix suffixes, tmux role themes, provider-pack escalation routes, formula/order targets, mail/nudge targets, and public/source classifier paths, and does not require every active behavior enumerator to consume classifier-filtered roots. Literal scanners alone cannot catch raw pack-root `ReadDir`, `WalkDir`, or `Glob` enumeration.
**Required change:** Narrow role-name allowance to exact configured-default binding keys and path-scoped fixtures. Define binding parser/resolver owners, config patch/merge paths, route provenance, required/optional binding metadata, literal-target conflict behavior, and diagnostics. Make `internal/packsource` the single retired-source classifier for load, install, cache, lock, materialization, docs lint, prompt/formula/order/script/hook enumeration, and doctor paths. Add a role-surface manifest, packsource adoption inventory, enumerator guard, state-by-operation matrix, expiry-backed allowlists, and negative fixtures.

### [Blocker] Doctor mutation and runtime-state migration lack a safe transaction/quiesce model

**Sources:** 03-doctor-mutation-safety, 08-runtime-state-migration, 09-rollout-decomposition-gates, 10-operator-docs-schema
**Actionability:** mixed
**Issue:** The proposed `internal/doctorfix` direction is right, but the design does not close existing direct `Fix(ctx)` and protected-write surfaces, define one shared city lock/generation boundary, or specify an idempotent multi-file recovery state machine. Runtime-state migration also misses real writers such as `jsonl-export.sh`, `spawn-storm-detect.sh`, legacy exec orders, old binaries, and sibling rigs; copy/digest checks cannot prove coherent state while those writers are active.
**Required change:** Inventory every auto-fix, import rewrite, pack install/update, required-pack repair/quarantine, runtime-state write, and cleanup path, then assign each to `FixIntent`, report-only, removal, or an explicit allowed escape. Define the doctor runner contract, lock/generation protocol, transaction journal path/schema, fsync/publish/commit ordering, rollback versus finish-forward rule, TOML scoped-edit verifier, and failure-injection tests. Add an exact path/key runtime-state migration table, marker schema, writer/quiesce policy, git-aware archive digest semantics, push reconciliation, post-marker old-binary detection, spawn-storm read-union, order alias compatibility, downgrade, rollback, and re-upgrade behavior.

### [Blocker] Bootstrap extraction and fixture isolation are not buildable enough

**Sources:** 07-bootstrap-fixture-isolation, 01-required-core-loading, 09-rollout-decomposition-gates
**Actionability:** document-fixable
**Issue:** Core extraction does not yet define the buildable end state for `internal/bootstrap/packs/core`, `//go:embed packs/**`, `bootstrapAssets`, `embeddedBootstrapPacks`, and `BootstrapPacks`. The old-bootstrap dependency inventory misses direct path/string consumers, and `GC_BOOTSTRAP=skip` containment is asserted without production-path negative tests.
**Required change:** State the exact slice that deletes or retargets bootstrap Core and make package deletion the compile-time proof for old Go imports. Add a production-safe non-nil empty `fs.FS`, scanner coverage for imports, path literals, constructed paths, `Subpath` strings, `AssetDir: "packs/core"`, docs, fixtures, examples, and generated files. Split mechanism fixtures from shipped-Core content tests, replace `AssetDir` dependencies with a named fixture seam, prove `GC_BOOTSTRAP=skip` cannot bypass production validation, and ensure replacement collision gates land no later than Core embed removal.

### [Blocker] Rollout gates and prerequisites are not honest enough for decomposition

**Sources:** 09-rollout-decomposition-gates, 02-behavior-evidence-chain, 05-public-pack-pin-cache, 06-pack-boundary-containment
**Actionability:** mixed
**Issue:** The plan still reads as ready to decompose while AC6, AC7, AC14, AC15, AC16, and AC17 artifacts are absent or uncited. Cross-repo `gascity-packs` work lacks producing branches, PRs, beads, commands, immutable commits, artifact paths, and downstream dependency edges. Slice-to-gate traceability is scattered across prose, and several slices have unresolved safety forks around mutating legacy rewrites, duplicate-active activation, cache rejection ownership, and public-pack bootstrapping.
**Required change:** Add or cite the AC17 acceptance-to-proof matrix with exact paths, commands, owners, artifacts, gates, and first dependent slice for AC1 through AC16. Add a slice-to-gate table covering touched files, prerequisites, proof commands, generated outputs, negative tests, old/new binary behavior, cache/offline behavior, rollback, and one-way boundaries. Name the public-pack producing work and staging handle, resolve Slice 2 mutation timing, Slice 5a activation duplicate behavior, `GC_BOOTSTRAP=skip` test semantics, and split or constrain broad slices before bead creation.

### [Major] Operator docs, diagnostics, schemas, and generated references are not release-gated per behavior slice

**Sources:** 10-operator-docs-schema, 09-rollout-decomposition-gates, 04-role-neutrality-zfc
**Actionability:** document-fixable
**Issue:** The docs/schema plan is improved but not executable enough. `docs/reference/system-packs.md` is named as an existing file even though it must be created or retargeted, the terminology matrix/wording scanner lack checked-in paths and fixtures, and Slice 7 appears to defer docs and diagnostic goldens after earlier operator-facing behavior changes.
**Required change:** Resolve the vocabulary document path, then define terminology matrix, wording scanner, docs authority audit, migration diagnostics schema, condition-code table, generated reports, fixtures, goldens, and CI wiring. Gate wording lint, doctor/import-state text and JSON goldens, generated references, OpenAPI/dashboard docs-schema outputs, CLI help, tutorial transcripts, and operator recovery guidance after every operator-facing slice. Include tutorial 05/07 dispositions, stale pack/cache examples, exact non-interactive doctor repair command, and isolated tmux cleanup rules for tests.

### [Minor] Review artifacts for attempt 6 wrote persona syntheses to the compatibility attempt-1 path

**Sources:** workflow artifact inspection
**Actionability:** workflow-defect
**Issue:** The attempt-6 fanout scopes closed pass, but the completed persona syntheses for `.gc/design-reviews/ga-1ekw9l` are under `attempt-1/persona-syntheses`; `attempt-6/persona-syntheses` is empty. The files in `attempt-1/persona-syntheses` were updated during the attempt-6 window and are the only complete persona-synthesis set available.
**Required change:** Keep this synthesis at `attempt-6/synthesis.md`, but fix the workflow's attempt-directory propagation so future persona syntheses write to the current attempt directory and global synthesis can read the current attempt path without compatibility fallback.

## Disagreements

- Persona verdicts are almost unanimous: nine block and one approve-with-risks. The non-blocking operator-docs lane still raises major release-gate issues, so it does not change the global result.
- Several raw model reviews praised the direction and gave approve-with-risks, while persona syntheses blocked because the plan still lacks concrete contracts, artifact paths, gates, or external evidence. I agree with the stricter synthesis because this review feeds decomposition, where missing contracts become unsafe beads.
- Reviewers differ on acceptable mechanisms: WAL versus convergence scan, hash markers versus full digest reads, AST scanner versus runtime participation stamp, local draft checkout versus immutable pin, and report-only versus automatic repair for some legacy states. The plan can choose mechanisms, but it must name one enforceable outcome and the proof command.
- Some blockers are document-fixable by making the plan precise; others are external prerequisites because public-pack commits, generated manifests, pin ledgers, and packcompat transcripts may not exist yet. The implementation plan should distinguish those classes rather than treating all missing proof as future implementation detail.
- The schema conformance check is not a blocker this time: the reviewed implementation plan has the required front matter and top-level sections. The remaining schema-related risk is readiness language, not output shape.

## Missing Evidence

- Trusted required-pack provenance API and tests binding file-set validation, resolved import edge, materialized directory, digests, layer id, effective contribution, and collision result.
- AST/type-aware loader bypass scanner or runtime participation-stamp guard, plus generated allowlist inventory for partial/non-behavior config reads.
- Behavior Evidence manifest schema, generator, freshness tests, row identity, evidence classes, source-kind extractors, old/new witness map, packcompat command contract, and public-pack reconciliation.
- Public `gascity-packs` branch/PR/bead handles, compatibility and activation commits, ownership rows, `behavior-preservation.yaml`, `public-gastown-pins.yaml`, and packcompat transcripts.
- Subpath-aware lock/cache schema, old-lock/cache migration policy, digest algorithms and scopes, read-hit validation, cache promotion atomicity tests, and pin-coherence gate.
- `internal/packsource` API, state enum, state-by-operation matrix, adoption inventory, enumerator guard, stable behavior-id schema, stale-directory/custom-fork provenance, and duplicate-active gate timing.
- Binding implementation details for `[gc.bindings.*]`, `[system_packs.*.bindings]`, `target_binding`, `gc.run_target_binding`, `GC_CORE_MAINTENANCE_WORKER`, warmup recipients, provider-pack escalations, and tmux theme de-roling.
- Existing doctor fix surface inventory, `FixIntent` runner contract, shared lock/generation protocol, recovery journal schema, TOML scoped-edit verifier, required-pack repair/quarantine write boundary, and failure-injection tests.
- Runtime-state path/key migration table, marker schema, writer/quiesce proof, git archive digest semantics, push reconciliation, old-binary post-marker detection, spawn-storm continuity, order alias compatibility, rollback, and re-upgrade tests.
- Bootstrap package/path dependency inventory, empty bootstrap FS implementation, `GC_BOOTSTRAP=skip` production-path negative tests, fixture-copy guard, replacement doctor bootstrap API, and collision-gate sequencing.
- AC17 acceptance-to-proof matrix, per-slice gate table, one-way rollback boundaries, exact old/new binary matrix, cache/offline matrix, and public-pack bootstrap staging workflow.
- Docs/reference vocabulary target, terminology matrix and wording scanner paths, migration diagnostic schema, condition-code table, generated reference goldens, tutorial transcripts, operator recovery guide, and per-slice docs/schema gates.

## Convergence Assessment

- Remaining blocker class: mixed
- Recommended apply verdict: iterate
- Reason: Another design-doc edit can resolve the document-fixable blockers by making the loader, packsource, role-binding, doctorfix, runtime-state, bootstrap, cache, docs, and rollout contracts explicit and schema-local. The same edit must also mark unavailable public-pack commits, generated manifests, pin ledgers, packcompat transcripts, and other proof artifacts as external prerequisites instead of implying full decomposition readiness.
- Next non-design work: public `gascity-packs` compatibility and activation artifacts, generated behavior/role/source inventories, pin ledgers, packcompat transcripts, cache/lock schema proof, doctor/runtime-state tests, docs/schema goldens, and process/integration validation artifacts.

## Recommended Changes

1. Update the implementation plan's readiness posture: either set `status: blocked` or state that only prerequisite-producing work may decompose until AC6, AC7, and AC14-AC17 artifacts exist and pass.
2. Add the AC17 acceptance-to-proof matrix and per-slice gate table with exact paths, commands, owners, artifacts, gates, and first dependent rollout slice.
3. Define the trusted required-pack loader/provenance contract and behavior-driving loader guard before any pack move, public pin, or source deletion slice depends on it.
4. Define the Behavior Evidence and packcompat contracts, including one-row-per-trigger extraction, old/new executable witnesses, immutable baseline, and public-pack reconciliation.
5. Move subpath-aware cache/proof identity, synthetic public-alias rejection, digest validation, and pin-coherence gates into the first slice that consumes the public Gastown pin.
6. Make `internal/packsource`, stable behavior ids, duplicate-active policy, stale-directory provenance, and all active enumerator guards explicit before compatibility or activation pins can load moved behavior.
7. Narrow role-name allowances and define symbolic binding resolution across config, formulas, orders, graph expansion, warmup, provider packs, tmux themes, mail/nudge targets, and diagnostics.
8. Specify doctorfix and runtime-state migration as one locked, recoverable transaction/convergence model with exact writer participation, marker schema, TOML verifier, archive push reconciliation, and high-risk race tests.
9. Make bootstrap Core extraction buildable and testable: delete or retarget old embeds, add empty FS semantics, scan all path/string consumers, isolate fixtures, and prove `GC_BOOTSTRAP=skip` cannot bypass production validation.
10. Gate operator-facing docs, diagnostics, generated references, CLI help, tutorial transcripts, and wording/schema goldens in the same slices that change the behavior they describe.
