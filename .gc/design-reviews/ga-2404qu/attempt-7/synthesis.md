# Design Review Synthesis

## Overall Verdict: block

All ten persona-level syntheses resolve to `block`, so the global verdict is `block` by worst-verdict-wins. The revised design has strong structural direction, especially the behavior manifest, staged rollout, pre-resolution doctor concept, and role-neutral Core notification contract, but reviewers found several gates that can pass while validating the wrong content, preserving unsafe runtime state, or leaving retired Maintenance/Gastown assets active.

## Consensus Strengths

- Multiple personas praised the move toward a machine-readable behavior manifest and per-row evidence, which is the right mechanism for proving that moved Gastown/Maintenance behavior still exists after the split.
- The review-gated, seven-slice rollout is materially safer than a flag-day migration, and reviewers agreed that the pack-first/public-pin intent is directionally correct.
- The pre-resolution doctor concept addresses a real bootstrap problem: stale imports must be diagnosable before full config evaluation depends on the very packs being migrated.
- The Core notification contract table is a useful abstraction for replacing hardcoded role recipients with configured recipients, formula variables, or pack-owned behavior.
- The registry direction, retiring Maintenance/Gastown from the bundled system-pack set while preserving stale directories on disk, is the right data-loss-avoidance stance if every loader ignores those stale directories.
- The design recognizes the correct high-risk surfaces: required Core inclusion, public Gastown pinning, stale synthetic caches, docs vocabulary, doctor fixes, and public-pack compatibility.

## Critical Findings

### [Blocker] Required Core Participation Can Be Falsely Proven
**Sources:** Elias Sato; Sofia Khoury; Marcus Driscoll; Ritu Raman; Tomas Park
**Issue:** The design still mixes a strong required-system-pack contract with weaker path/provenance checks and include-builder helpers. A materialized path in provenance does not prove Core identity, manifest integrity, freshness, repair status, or resolved participation. Production bypass coverage is also too narrow: reviewers found current `internal/` resolver surfaces, controller reload paths, and `GC_BOOTSTRAP=skip` behavior that can avoid or weaken required Core participation checks.
**Required change:** Define a typed, content-backed `RequiredSystemPackParticipation` contract, produced by production config loaders and asserted after resolution. Required Core must have two fatal gates on every normal production path: strict manifest/file-set integrity before participation and post-load proof that the validated host Core layer actually participates. Widen bypass enforcement to production `internal/` loaders and either retire or narrow `GC_BOOTSTRAP=skip` so it cannot skip pruning, collision checks, Core materialization, or provenance validation.

### [Blocker] Role Neutrality Is Claimed While Go and Pack Surfaces Still Encode Roles
**Sources:** Ingrid Kovac; Nadia Volkov; Avery McAllister; Marcus Driscoll; Felix Moreau
**Issue:** The design claims zero hardcoded roles and gates source deletion on role cleanup, but no rollout slice resolves the production Go role surface or the Core/Gastown `dog` ownership model. Reviewers identified behavior-bearing role logic in tmux themes/icons, default city scaffolding, prompt fallback, warmup mail defaults, sling formula-name heuristics, API `crew` classification, provider `dog` bindings, Core notification targets, TOML defaults, comments, prompt prose, and moved order/formula assets.
**Required change:** Add a Go and asset role-surface migration table with file/function, current behavior, final owner, replacement mechanism, slice, and test. Either add an explicit Go de-roling slice or explicitly defer it and remove it as a source-deletion gate. Decide whether `dog` is a stable Core compatibility contract, a configurable maintenance-worker target, or a Gastown convention, then update Core orders, provider packs, docs, and renamed/omitted-worker tests to match.

### [Blocker] Public Gastown Pin Gates Can Validate Bundled or Duplicate Content
**Sources:** Yuki Hayashi; Marcus Driscoll; Avery McAllister; Tomas Park; Nadia Volkov
**Issue:** The public pin adoption flow can appear green while `PublicGastownPackSource` is still recognized as bundled synthetic content, so install/cache/include checks may validate embedded bytes instead of a real remote commit. Reviewers also found an intermediate-state risk where public Gastown and stale local Maintenance/Gastown definitions can both be active, causing duplicate orders, formulas, prompts, scripts, or patches.
**Required change:** In the public-pin slice, retire or bypass the public synthetic alias before consuming `PublicGastownPackVersion`. Add failing-before/passing-after tests proving ordinary remote cache use, exact commit checkout, `gastown` subpath identity, no duplicate active definitions, and no Maintenance/Gastown system-pack participation through stale local imports or preserved directories.

### [Blocker] Doctor `--fix` Is Not Failure-Atomic or Preservation-Proven
**Sources:** Sofia Khoury; Yuki Hayashi; Tomas Park; Nadia Volkov; Felix Moreau
**Issue:** The doctor plan promises scoped preservation of TOML comments, unknown fields, ordering, and custom content, but does not name an executable preserving editor. It also lacks one coordinated transaction boundary across `city.toml`, rig `pack.toml`, lockfiles, installed packs, Core materialization, runtime-state movement, and public Gastown install validation. Runtime state can move while manifests roll back, and legacy import removal can happen before Core/public Gastown validation is complete.
**Required change:** Replace whole-file TOML re-encoding with a concrete CST-preserving or line-scoped editor, or make `--fix` refuse with manual steps when preservation cannot be proven. Add a composite doctor migration coordinator: plan first, prove content provenance, run operation-scoped preflight, stage Core in a temp directory, move runtime state in a recoverable sequence, write manifests last with compare-before-rename, validate afterward, and cover failure-injection and air-gap paths.

### [Blocker] Retired Maintenance and Gastown Assets Can Still Affect Runtime
**Sources:** Marcus Driscoll; Avery McAllister; Felix Moreau; Tomas Park; Ritu Raman
**Issue:** Preserving stale `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories is safe only if every runtime and docs path treats them as retired. Reviewers found prompt/template globbing that can still load stale prompts, stale local imports that can reactivate retired packs, generated/bootstrap path references, downstream examples that point to retired scripts, and docs that teach users to import Maintenance.
**Required change:** Filter prompt/template discovery to active or required packs, add a central retired-source classifier with typed diagnostics, and test absent/customized/stale retired directories through the normal production loader and materializer. Expand the docs/example update plan into an inventory-backed lint gate covering Markdown, MDX, TOML, shell, Go strings, generated references, CLI help, doctor messages, examples, pack comments, and navigation.

### [Blocker] Behavior Preservation and Packcompat Evidence Are Underspecified
**Sources:** Nadia Volkov; Avery McAllister; Tomas Park; Marcus Driscoll
**Issue:** The behavior manifest is promising but not yet a sufficient proof system. It misses behavior-bearing dimensions such as daemon commit identity, notification/requester/detector semantics, moved helper dependencies, provider database filters, TOML defaults, prompt fragments, and test-function/subtest granularity. Packcompat is asked to cover many states without a concrete fixture matrix, and Gas City does not yet prove it consumes the exact public Gastown manifest from the pinned commit.
**Required change:** Add row-level evidence classes, schema/generator versions, old-source digests, behavior-bearing asset digests, semantic-delta status, and cross-repo ownership. Require a generated test-migration table for every touched or removed test. Define a packcompat matrix covering fresh init, upgraded locks, stale synthetic cache, ordinary remote cache, offline no-cache, old-binary/new-pack, new-binary/old-pack, no-Maintenance loader, and host-Core patch resolution.

### [Major] Registry, Cache, and Materializer Boundaries Need Slice-Accurate Gates
**Sources:** Marcus Driscoll; Ritu Raman; Tomas Park; Yuki Hayashi
**Issue:** Registry cleanup details remain implicit: retired source aliases, `PublicRepository`, `publicSubpathForPack`, public URL normalization, `RepoCacheKey` migration, global `SyntheticContentHash` invalidation, `requiredBuiltinPackNames`, and `MaterializeBuiltinPacks` iteration all affect whether a slice is truly no-Maintenance. The design can remove Maintenance from required includes before the materializer stops creating or refreshing retired directories.
**Required change:** Specify exact delete/retain decisions and slice ordering for registry helpers, source recognizers, synthetic cache keys, materializer iteration, and stale directory pruning. Add offline old-cache-to-new-binary tests proving `bd` and `dolt` self-heal without network and remain byte-identical aside from regenerated markers.

### [Major] Bootstrap Extraction Has Cross-Slice Escape Hatches
**Sources:** Ritu Raman; Elias Sato; Tomas Park
**Issue:** Bootstrap production asset removal depends on a clear empty `fs.FS` default, inline test fixtures, managed-name sync, and `GC_BOOTSTRAP=skip` semantics. Current design references a managed-name sync guard that does not exist, and path-string references to `internal/bootstrap/packs/core` remain in tests, hook docs, and active engineering docs.
**Required change:** Make production bootstrap assets a private non-nil empty `fs.FS`, define test-only fixture injection with strict allowlists, add the missing `BootstrapManagedImportNames` sync test, inventory managed-name consumers, and add a broad path-string guard for stale bootstrap Core references.

### [Major] Docs and Operator Vocabulary Are Not Yet a Verified Contract
**Sources:** Felix Moreau; Avery McAllister; Marcus Driscoll
**Issue:** The planned canonical `docs/reference/system-packs.md` page does not exist or appear in docs navigation, while live examples still show `[imports.maintenance]`, stale `.gc/system/packs/gastown` paths, and Maintenance-as-actor troubleshooting text. The design also lacks a single executable vocabulary source that distinguishes retired Maintenance from valid lower-case maintenance terms and Dolt/store maintenance events.
**Required change:** Create and nav-register the canonical system-pack reference page or designate an existing one. Add exact-phrase or equivalent docs-lint/golden coverage for doctor output, FixHint text, CLI/config/schema generated docs, tutorials, troubleshooting, examples, and allowed contexts such as `gc.store.maintenance.*`.

### [Minor] Several Review Artifacts Reveal Workflow Hygiene Gaps
**Sources:** Global synthesizer assessment; persona task metadata
**Issue:** The persona synthesis tasks for attempt 7 recorded output paths under `attempt-1/persona-syntheses`, even though their metadata has `gc.attempt=7` and the files were freshly written during this attempt. This did not block synthesis because all ten persona syntheses were present and readable, but it makes artifact discovery fragile.
**Required change:** Fix the persona-synthesis output path calculation so it uses the design-review attempt number, not the retry wrapper attempt number. Add a workflow assertion that attempt N has an attempt-local `persona-syntheses/` directory before global synthesis runs.

## Disagreements

- Persona-level verdicts are unanimous: all ten syntheses returned `block`. The main disagreements are inside persona sources, where some models rated mature areas as `approve-with-risks` while others blocked.
- Reviewers disagree on whether Go de-roling must be in scope for this migration. My assessment: either in-scope or explicitly deferred can work, but the design is currently inconsistent because it treats Go cleanup as a deletion gate without providing a slice, scanner contract, or replacement mechanisms.
- Reviewers disagree on `dog`: stable Core compatibility pool, configurable Core maintenance target, or Gastown convention. My assessment: any one model can be viable, but the design must choose one because provider packs, docs, Core orders, and SDK self-sufficiency tests depend on it.
- Reviewers disagree on whether extra unexpected files under required system packs must always be fatal or can be proven non-loadable. My assessment: required packs should either use strict full-file-set validation or document a deterministic pruning/proof model before config resolution.
- Reviewers disagree on whether generated/historical docs should be rewritten or allowlisted. My assessment: both are acceptable only if every retained hit is classified as generated, historical, migration fixture, legacy diagnostic, or still-valid maintenance terminology.

## Missing Evidence

- Typed required-system-pack participation record, assertion algorithm, manifest-integrity result, and fatal failure policy for all production loaders.
- Complete production direct-load inventory for `config.Load`, `config.LoadCity`, and `config.LoadWithIncludes` across `cmd/gc` and relevant `internal/` packages.
- Go role-surface migration table and replacement mechanisms for tmux themes/icons, default init, prompt fallback, warmup mail, sling formula heuristics, API `crew`, and role-bearing Go constants.
- Concrete `dog` ownership decision with Core-only, renamed, omitted, provider-pack, and public Gastown compatibility tests.
- Public Gastown remote-fetch gate proving exact immutable commit checkout and no synthetic alias/cached bundled content.
- Duplicate-definition tests for local stale Maintenance/Gastown plus public Gastown during intermediate slices.
- Scoped TOML editor choice, content-provenance hash strategy, composite doctor transaction plan, runtime-state migration table, and failure-injection tests.
- Central retired-source classifier used by lock validation, config load, import check/install, cached reads, and pre-resolution doctor.
- Prompt/template discovery test proving preserved stale system-pack directories cannot contribute baselines.
- Packcompat fixture matrix and exact pinned public manifest consumption check.
- Generated test-migration table for touched/removed Gas City tests and external pack tests.
- Docs inventory/lint covering Markdown, MDX, generated references, CLI help, schemas, examples, scripts, pack comments, doctor strings, and docs navigation.
- Bootstrap fixture allowlist, `GC_BOOTSTRAP=skip` narrowed/retired behavior, and stale `internal/bootstrap/packs/core` path guard.
- Process and integration gate list for doctor mutation, runtime-state migration, init/import/check, old/new binary compatibility, and source deletion.

## Recommended Changes

1. Define the required Core participation contract and retrofit production loaders, controller reload, dispatch routing, bootstrap skip behavior, and scanner/allowlist enforcement around it.
2. Resolve role neutrality explicitly: Go role-surface scope, `dog` ownership, `crew` classification, tmux display metadata, sling branch metadata, Core notification targets, and replacement tests.
3. Rework public Gastown pin adoption so it uses ordinary remote resolution before the pin is consumed, with exact commit/subpath validation and duplicate active-definition tests.
4. Replace the doctor fix design with a preservation-proven, transaction-coordinated migration plan covering TOML, locks, Core materialization, public Gastown install, runtime state, air-gap diagnostics, and rollback/failure injection.
5. Add retired-source and stale-directory containment across config loading, import/install/check, materialization, prompt/template discovery, docs, and examples.
6. Expand the behavior manifest into a real proof contract with evidence classes, per-row digests, semantic-delta rules, generator/schema versions, and public-pack manifest compatibility.
7. Add a concrete packcompat fixture matrix and old/new binary harness before source deletion or public pin consumption.
8. Tighten registry/cache/materializer slice ordering so no-Maintenance claims are tested on the same production paths operators run.
9. Finish bootstrap extraction design with non-nil empty production FS, strict inline fixture allowlists, managed-name sync, and `GC_BOOTSTRAP=skip` containment tests.
10. Promote docs/operator vocabulary to an executable lint/golden contract, create or designate the canonical system-pack reference, and update all stale examples and generated references.
