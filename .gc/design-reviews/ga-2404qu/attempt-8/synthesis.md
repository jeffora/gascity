# Design Review Synthesis

## Overall Verdict: block

All ten persona-level syntheses resolve to `block`, so the global verdict is
`block` by worst-verdict-wins. The revised design has the right direction in
several places, especially behavior-manifest thinking, role-neutral Core goals,
pre-resolution doctor diagnostics, and staged rollout intent, but reviewers
still found migration states that can pass while loading wrong content,
preserving unsafe runtime state, or violating the no-hardcoded-roles contract.

## Consensus Strengths

- Multiple personas praised the behavior manifest and packcompat direction as
  the right way to prove moved Gastown/Maintenance behavior rather than relying
  on path counts or narrative assertions.
- The staged rollout, public-pin plan, and old/new compatibility framing are
  safer than a flag-day migration if the slice gates are made executable.
- The design recognizes the correct high-risk surfaces: required Core loading,
  public Gastown pin integrity, stale retired packs, doctor preservation,
  runtime-state migration, docs vocabulary, provider pack continuity, and role
  neutrality.
- Reviewers agreed that pre-resolution doctor diagnostics are necessary because
  a retired or corrupt system pack can break full config resolution before the
  operator gets useful guidance.
- The notification/requester abstraction and Core role-neutrality goals are
  directionally correct, but need asset-by-asset contracts and tests.

## Critical Findings

### [Blocker] Required Core Participation Can Still Be Falsely Proven
**Sources:** Elias Sato; Ritu Raman; Marcus Driscoll; Tomas Park
**Issue:** The design still mixes a strong required-system-pack contract with
weaker path/provenance checks. A materialized Core path does not prove Core
identity, manifest integrity, freshness, repair status, collision safety, or
resolved participation. Several production surfaces can still bypass or weaken
the contract, including internal dispatch routing, controller reload, no-refresh
loaders, partial config reads, and `GC_BOOTSTRAP=skip`.
**Required change:** Define a typed, content-backed
`RequiredSystemPackParticipation` contract and make production loading enforce
two fatal gates: strict required-pack integrity before participation, and
post-load proof that the validated host Core layer participated in the resolved
config. Widen bypass enforcement beyond `cmd/gc`, settle Core-name collision
behavior, and retire or narrow `GC_BOOTSTRAP=skip` so it cannot suppress
materialization, pruning, integrity, provenance, collision, or doctor cleanup
checks.

### [Blocker] Role Neutrality Is Claimed While Role Logic Remains in Go and Pack Assets
**Sources:** Ingrid Kovac; Nadia Volkov; Avery McAllister; Marcus Driscoll; Felix Moreau
**Issue:** The design gates source deletion on role neutrality, but does not
resolve production Go and asset role surfaces. Reviewers identified
role-bearing behavior in tmux themes/icons, default city scaffolding, prompt
fallback, warmup mail defaults, sling formula-name heuristics, API `crew`
classification, provider `dog` bindings, Core notification targets, TOML
defaults, scripts, prompt fragments, comments, and formula/order metadata. A
role-token scanner can go green while Go decision logic or Core routes still
encode Gastown-specific roles.
**Required change:** Add a Go and asset role-surface migration table with file
or asset, current behavior, final owner, replacement mechanism, rollout slice,
and proof test. Either add a Go de-roling slice or explicitly defer Go de-roling
and remove it as a source-deletion gate. Decide whether `dog` is a stable Core
compatibility contract, a configurable maintenance-worker target, or a Gastown
convention, then update Core orders, provider packs, docs, scanner policy, and
renamed/omitted-worker tests to match.

### [Blocker] Public Gastown Pin Gates Can Validate Bundled or Duplicate Content
**Sources:** Yuki Hayashi; Marcus Driscoll; Avery McAllister; Tomas Park; Nadia Volkov
**Issue:** The public-pin adoption slice can appear green while
`PublicGastownPackSource` still resolves through bundled synthetic aliases or
old cache namespaces, validating embedded bytes instead of an ordinary remote
commit. The rollout also permits windows where public Gastown and stale local
Maintenance/Gastown definitions can both be active, causing duplicate orders,
formulas, scripts, prompts, or base formulas.
**Required change:** Move public Gastown synthetic-alias bypass/removal into
the public-pin adoption slice before `PublicGastownPackVersion` is consumed.
Add failing-before/passing-after tests proving exact immutable remote checkout,
`gastown` subpath identity, ordinary remote cache use, stale synthetic-cache
diagnostics, and no duplicate active definitions across slices 2-6, rollback
states, stale local imports, and old-binary/new-pack combinations.

### [Blocker] `gc doctor --fix` Is Not Yet Preservation-Proven or Failure-Atomic
**Sources:** Sofia Khoury; Yuki Hayashi; Tomas Park; Felix Moreau
**Issue:** The doctor plan depends on preserving comments, unknown fields,
ordering, lockfile shape, installed-pack metadata, Core materialization,
public-pack provenance, and runtime-state paths, but it does not name an
executable preserving editor or a single commit boundary. Later instructions
still allow direct `MaterializeBuiltinPacks(cityPath)`, path-only provenance,
or partial live commits before final validation succeeds.
**Required change:** Define the doctor mutation coordinator before any
implementation bead starts: plan all manifest, lock, installed-pack, Core,
public Gastown, and runtime-state mutations; run operation-scoped preflight;
stage Core and public content; use a CST/span-preserving TOML editor or refuse
with manual steps; publish through a recoverable transaction or validate in a
staged overlay before live rename; and add fault-injection coverage for partial
writes, stale targets, controller-active cities, air-gapped repairs, failed
fetches, lockfile lexical precision, and repeated healthy no-op fixes.

### [Blocker] Retired Maintenance and Gastown Assets Can Still Affect Runtime
**Sources:** Marcus Driscoll; Avery McAllister; Felix Moreau; Ritu Raman; Tomas Park
**Issue:** Preserving stale `.gc/system/packs/maintenance` and
`.gc/system/packs/gastown` directories is safe only if every runtime path treats
them as retired. Reviewers found risks in direct or transitive imports, prompt
and template globbing, generated or bootstrap path references, stale cache
fallback, docs/examples that still point to retired paths, and materializer
behavior that may keep creating, refreshing, or pruning retired packs after the
no-Maintenance slice.
**Required change:** Add a central retired-source classifier used by config
load, lock validation, cached reads, import install/check, packman paths, and
pre-resolution doctor. Filter prompt/template discovery to active or required
packs, stop materializing retired system-pack directories in the no-Maintenance
slice, test absent/stale/customized retired directories through normal
production loaders, and gate docs/examples/scripts/generated references with an
inventory-backed lint that classifies every retained hit.

### [Blocker] Behavior Preservation Evidence Is Still Underspecified
**Sources:** Nadia Volkov; Avery McAllister; Tomas Park; Marcus Driscoll
**Issue:** The behavior manifest is promising but not yet a complete proof
contract. It misses or underspecifies notification/requester/detector
semantics, Git/Dolt/process authorship identity, moved helper dependencies,
provider database filters, TOML defaults, prompt fragments, session hooks,
dog-field ownership, and per-test migration evidence. Packcompat also lacks a
fixture matrix proving the exact public Gastown manifest consumed by Gas City at
the pinned commit.
**Required change:** Expand the manifest with row-level evidence classes,
schema/generator versions, generator owner and command, old-source and
behavior-bearing asset digests, semantic-delta approval records, public-pack
ownership, and test-function/subtest mappings. Add a packcompat matrix for
fresh init, upgraded locks, stale synthetic cache, ordinary remote cache,
offline no-cache, old binary/new pack, new binary/old pack, no-Maintenance
loader, host-Core patching, stale local Gastown imports, and provider pack
continuity.

### [Major] Bootstrap Extraction and Registry/Cache Boundaries Need Slice-Accurate Gates
**Sources:** Ritu Raman; Marcus Driscoll; Elias Sato; Tomas Park
**Issue:** Bootstrap and registry cleanup still have cross-slice hazards:
`bootstrapManagedImportNames` can be emptied before required-Core collision
enforcement lands, `bootstrapAssets` lacks a concrete non-nil empty production
`fs.FS`, hidden old `internal/bootstrap/packs/core` references remain in tests
and docs, and helper dispositions for `All()`, `requiredBuiltinPackNames`,
`publicSubpathForPack`, `PublicRepository`, `RepoCacheKey`,
`SyntheticContentHash`, `MaterializeSyntheticRepo`, and pruning are not fully
assigned to slices.
**Required change:** Land managed-name, collision, skill-suppression, bootstrap
fixture, hidden-dependency, registry, materializer, synthetic-cache, and old-path
changes in coherent green slices. Add `internal/packs/core` gates when that
package appears, old-cache/offline provider-continuity tests, and a broad
path-string guard for stale bootstrap Core references.

### [Major] Docs and Operator Vocabulary Are Not Yet an Executable Contract
**Sources:** Felix Moreau; Avery McAllister; Marcus Driscoll; Sofia Khoury
**Issue:** The design names `docs/reference/system-packs.md` as the canonical
terminology anchor, but that page does not yet exist or appear in navigation.
Known stale surfaces remain across tutorials, migration guides, troubleshooting
MDX, generated references, schemas, examples, scripts, CLI/help text, doctor
strings, pack comments, and public Gastown docs. Troubleshooting cannot be
truthful until runtime-state and `dog`/Core patch semantics are chosen.
**Required change:** Create and nav-register the canonical system-pack reference
or designate an existing page. Make one executable wording matrix the source of
truth for Core, provider system packs, public Gastown, retired Maintenance,
Core utility worker terminology, stale generated directories, store/Dolt
maintenance, and ordinary maintenance prose. Pair docs, lint, generated refs,
doctor strings, and golden tests with the first slice that changes
operator-facing behavior, or mark those slices as a non-release sequence.

### [Minor] Attempt-8 Persona Synthesis Artifacts Are Stamped Into the Wrong Directory
**Sources:** Global synthesizer assessment; persona synthesis bead metadata
**Issue:** All ten attempt-8 persona synthesis beads closed with
`gc.attempt=8`, but their `design_review.output_path` metadata points to
`.gc/design-reviews/ga-2404qu/attempt-1/persona-syntheses/...`. The files are
present and have attempt-8 timestamps, so synthesis could proceed, but
attempt-local artifact discovery remains fragile.
**Required change:** Fix persona-synthesis output path calculation so attempt N
writes to `attempt-N/persona-syntheses/`. Add a workflow assertion before
global synthesis that every required attempt-local persona synthesis exists and
matches the current attempt metadata.

## Disagreements

- Persona-level verdicts are unanimous: all ten syntheses returned `block`.
  Disagreements are within persona lanes, where some individual model reviews
  rated mature areas as `approve` or `approve-with-risks` while others blocked.
- Reviewers disagree on whether Go de-roling must be in this migration. My
  assessment: either in-scope or explicitly deferred can work, but the current
  design is inconsistent because it treats Go cleanup as a deletion gate without
  assigning replacement mechanisms, slices, or scanner coverage.
- Reviewers disagree on `dog`: stable Core compatibility pool, configurable
  Core utility target, or Gastown convention. My assessment: any one model can
  be viable, but the design must choose one because provider packs, Core orders,
  docs, and SDK self-sufficiency tests depend on it.
- Reviewers disagree on required-pack extra files: strict full-file-set
  rejection versus deterministic proof/pruning. My assessment: either can work
  if it is applied before config resolution and covered by negative tests.
- Reviewers disagree on rollback tolerance across the Maintenance-removal
  boundary. My assessment: full compatibility and a declared one-way boundary
  are both defensible, but the matrix must state the tested outcome per slice
  and provide operator recovery guidance.
- Reviewers differ on whether explicit docs file lists or an inventory/lint
  should be authoritative. My assessment: the inventory/lint must be
  authoritative; file lists should only call out known hotspots.

## Missing Evidence

- Typed required-system-pack participation record, manifest-integrity result,
  assertion algorithm, Core collision policy, and fatal failure policy for all
  production loaders.
- Complete inventory and migrate-or-allowlist disposition for production
  `config.Load`, `config.LoadCity`, and `config.LoadWithIncludes` call sites in
  `cmd/gc` and relevant `internal/` packages.
- Go and asset role-surface migration table, scanner contract, allowlist schema,
  and replacement mechanisms for tmux display metadata, default init, prompt
  fallback, warmup mail, sling branch heuristics, API `crew`, and Core routes.
- Concrete `dog` ownership decision with Core-only, provider-pack, renamed, and
  omitted-worker compatibility tests.
- Public Gastown remote-fetch gate proving exact immutable commit checkout,
  durable reachability, `gastown/pack.toml` identity, ordinary remote cache use,
  and no synthetic alias/cached bundled content.
- Duplicate-definition tests for public Gastown plus bundled, stale, local,
  remote, synthetic, or materialized Maintenance/Gastown definitions.
- Preservation-proven doctor editor, composite transaction or staged-overlay
  model, content-backed generated-source provenance, controller-active policy,
  air-gap behavior, and failure-injection tests.
- Central retired-source classifier and prompt/template active-pack filter used
  by all load/install/check/read paths.
- Packcompat fixture matrix covering fresh init, upgraded locks, stale caches,
  offline states, old/new binaries, no-Maintenance loading, host-Core patching,
  provider continuity, and stale local imports.
- Behavior-manifest rows for notification/requester/detector semantics,
  authorship identity, `session_live` hooks, dog prompt fragments and agent
  fields, provider database filters, TOML role defaults, and moved helper
  dependencies.
- Machine-checkable current-test migration table for every touched or removed
  Gas City test and public-pack replacement test at the pinned commit.
- Bootstrap fixture allowlist, production empty `fs.FS`, `GC_BOOTSTRAP=skip`
  decision, managed-name sync, and stale `internal/bootstrap/packs/core` guard.
- Canonical system-pack docs page or registered equivalent, docs navigation,
  generated-reference regeneration/allowlist, executable wording matrix, and
  runtime-state troubleshooting contract.
- Attempt-local persona synthesis directory for attempt 8.

## Recommended Changes

1. Define and wire the required Core participation contract first, including
   manifest integrity, post-load participation, collision handling, production
   bypass inventory, controller reload, dispatch routing, and bootstrap skip
   semantics.
2. Resolve role neutrality explicitly: Go scope, scanner policy, `dog`
   ownership, `crew` classification, tmux metadata, sling branch metadata,
   default init/prompt fallback, and Core notification targets.
3. Rework public Gastown pin adoption so the pin is validated through ordinary
   remote resolution before it is consumed, with exact commit/subpath checks and
   duplicate active-definition gates.
4. Replace the doctor fix plan with a preservation-proven, transaction-aware
   migration model covering TOML, locks, installed packs, Core materialization,
   public Gastown install, runtime state, air-gap diagnostics, and rollback or
   staged validation.
5. Add retired-source containment across config loading, import/check/install,
   cache reads, materialization, prompt/template discovery, docs, examples, and
   public Gastown prose.
6. Expand the behavior manifest and packcompat matrix into executable proof
   contracts with row-level evidence, old/new test mappings, semantic-delta
   approvals, public-pack commit evidence, and provider continuity tests.
7. Fix slice ordering so Bootstrap/Core extraction, no-Maintenance loading,
   registry/cache cleanup, materializer behavior, source deletion, and docs
   updates each land with green production-path gates.
8. Promote docs/operator vocabulary to an executable lint/golden contract and
   create or designate the canonical system-pack reference before changing
   operator-facing diagnostics.
9. Fix the workflow artifact bug so attempt-8 persona syntheses are written
   under `attempt-8/persona-syntheses` and global synthesis fails early if
   attempt-local required inputs are absent.
