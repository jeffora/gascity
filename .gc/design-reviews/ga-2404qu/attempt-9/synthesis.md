# Design Review Synthesis

## Overall Verdict: block

All ten persona-level syntheses return `block`, so the global verdict is `block`
by worst-verdict-wins. The design has converged on the right major safety
themes, especially content-backed Core participation, public-pack compatibility
gates, behavior evidence, role neutrality, and preservation-focused doctor
repairs, but reviewers still found unresolved contracts that can allow wrong
content to load, retired behavior to remain active, or user data to be mutated
without a proven recovery boundary.

## Consensus Strengths

- Reviewers consistently praised the move away from include counts and path
  checks toward behavior evidence, packcompat fixtures, and content-backed
  participation records.
- The design recognizes the highest-risk migration boundaries: required Core
  loading, public Gastown pin adoption, Maintenance retirement, stale local
  sources, doctor fixes, runtime-state migration, bootstrap fixture isolation,
  and docs/operator terminology.
- Several personas agreed the staged rollout can work if each slice is made a
  coherent green state with exact duplicate-definition, source-identity, and
  old/new binary compatibility gates.
- The doctor design is heading toward the right preservation model: scoped
  edits, provenance checks, live-controller awareness, staged writes, and
  failure-injection coverage.
- The role-neutral Core goal is well understood by the design, and reviewers
  agreed the scanner/manifest direction can become a useful CI contract once it
  covers the actual Go and asset surfaces.

## Critical Findings

### [Blocker] Required Core Loading Is Still Weaker Than the Stated Invariant
**Sources:** Elias Sato; Tomas Park; Marcus Driscoll; Ritu Raman
**Issue:** The constructive loader and doctor sections still leave room for
path/provenance-only proof. A materialized `.gc/system/packs/core` path or
successful file read does not prove the validated host Core was fresh, intact,
non-shadowed, and active in resolved runtime config. The bypass plan is also
too narrow: it centers `cmd/gc` while the changed production surfaces include
controller reload, API/config helpers, dispatch, doctor/configedit, packman,
import/check/install, no-refresh loaders, and other `internal/` paths.
**Required change:** Define a typed, content-backed
`RequiredSystemPackParticipation` contract with pre-resolution fatal integrity
and post-resolution fatal participation gates. Inventory every production
loader and wrapper, classify exceptions with focused tests, make user/imported
`core` collisions fatal or disambiguated, and extend bypass scanning beyond
`cmd/gc` to every production config-resolution surface in scope.

### [Blocker] Public Gastown Pin Adoption Can Load Duplicate or Wrong Content
**Sources:** Yuki Hayashi; Marcus Driscoll; Avery McAllister; Tomas Park; Nadia Volkov
**Issue:** The rollout can consume a public Gastown pin while bundled
force-required Maintenance is still active, creating duplicate orders,
formulas, scripts, or base formulas. Other paths can still validate synthetic
or stale cache content instead of an ordinary remote checkout, and the
distribution story can drift between `PublicGastownPackVersion`, direct source
resolution, registry selector resolution, lock entries, and subpath identity.
**Required change:** Resolve the bundle-versus-pin overlap with either a
backward-compatible first public pin or an atomic pin/removal boundary. Add
zero-duplicate-active-definition tests for current-loader plus candidate pin,
old-loader plus new pin, and rollback states. Prove that
`https://github.com/gastownhall/gascity-packs.git//gastown` resolves through
ordinary remote cache identity at the same immutable, reachable commit and hash
as registry-selector installation.

### [Blocker] Retired Maintenance and Gastown Sources Can Still Participate
**Sources:** Marcus Driscoll; Avery McAllister; Felix Moreau; Tomas Park
**Issue:** Removing Maintenance from required includes is not enough if stale
`.gc/system/packs/maintenance`, `.gc/system/packs/gastown`, old locks, local
imports, transitive imports, prompt/template globbing, synthetic cache fallback,
or packman/import paths can still read retired sources. Preserving stale
directories is safe only if every runtime path treats them as retired
diagnostics rather than active content.
**Required change:** Add one retired-source classifier with typed diagnostics
and use it before ordinary remote fallback in config load, lock validation,
cache reads, import install/check, packman paths, prompt/template discovery,
and pre-resolution doctor. Stop materializing, refreshing, pruning, or loading
retired system-pack directories in the no-Maintenance slice, and test absent,
stale, customized, and transitive-import cases through production loaders.

### [Blocker] `gc doctor --fix` Is Not Yet Preservation-Proven or Failure-Atomic
**Sources:** Sofia Khoury; Felix Moreau; Yuki Hayashi; Tomas Park
**Issue:** The design promises preservation of comments, unknown TOML fields,
ordering, lockfile shape, installed-pack metadata, Core materialization,
runtime-state paths, and public-pack provenance, but it does not yet name a
single enforceable commit boundary. Some later instructions still imply direct
`MaterializeBuiltinPacks(cityPath)`, path-only provenance, or live multi-target
mutation before final validation succeeds.
**Required change:** Define the doctor mutation coordinator before
implementation: build the mutation plan, run operation-scoped preflight, stage
Core and public content, use a CST/span-preserving TOML editor or refuse with
manual guidance, publish through a recoverable transaction or fully staged
overlay validation, and cover controller-active cities, air-gapped repairs,
partial writes, failed renames, stale targets, lockfile lexical precision,
runtime-state moves, and repeated healthy no-op fixes.

### [Blocker] Role Neutrality Is Not Proven for Go or Core-Owned Assets
**Sources:** Ingrid Kovac; Nadia Volkov; Avery McAllister; Marcus Driscoll
**Issue:** The design gates source deletion on zero hardcoded roles but does
not resolve production Go role logic or all asset role surfaces. Reviewers
called out tmux role maps/themes, default `mayor` scaffolding, prompt fallback,
warmup recipients, sling formula-name heuristics, API `crew` classification,
`dog` ownership, Core mail/nudge targets, TOML defaults, scripts, prompts,
descriptions, metadata, and comments. A token scanner can pass while branch
selection or routing still depends on Gastown identity.
**Required change:** Add a checked-in Go and asset role-surface migration table
with file/asset, current behavior, final owner, replacement mechanism, rollout
slice, and proof test. Decide `dog`, `crew`, tmux display identity,
non-Gastown `gc init`, prompt fallback, warmup targets, and sling branch needs
through configuration or explicit compatibility contracts. Make scanner scope,
token set, fields, comments/prose policy, allowlist schema, owners, reasons,
expiry, and negative fixtures precise enough for CI.

### [Blocker] Behavior Preservation Evidence Still Allows Silent Loss
**Sources:** Nadia Volkov; Avery McAllister; Tomas Park; Marcus Driscoll
**Issue:** The behavior manifest direction is strong, but it still needs a
witness-floor rule and complete source discovery. Reviewers found gaps around
requester, detector, notification, mail/nudge, script branch, prompt
instruction, route metadata, runtime-state mutation, authorship identity,
provider safety, session hooks, dog field ownership, commands, doctor checks,
and duplicated exec/formula code paths.
**Required change:** Require execution-level final witnesses whenever the old
behavior had execution-level evidence. Add semantic-delta approval records for
approved behavior changes, whole-repo and public-pack behavior discovery,
per-code-path manifest rows, behavioral-trigger packcompat fixtures, authorship
identity tests, provider-safety rows, and public Gastown same-commit proof for
the replacement behavior manifest.

### [Blocker] Bootstrap and Test-Fixture Escape Hatches Remain Ambiguous
**Sources:** Ritu Raman; Elias Sato; Tomas Park
**Issue:** `GC_BOOTSTRAP=skip` still has no final migration decision, and the
production `bootstrapAssets` fallback after removing `//go:embed packs/**` is
not specified enough to prevent nil, empty-embed, or `testing/fstest` leakage
into production code. The fixture plan also does not fully distinguish
bootstrap-mechanism tests from real Core content/behavior tests.
**Required change:** Choose the `GC_BOOTSTRAP=skip` outcome now: retire it,
narrow it, or make it test-only/deprecated. Specify a private production empty
`fs.FS` returning `fs.ErrNotExist`, use `_test.go` fixture injection only, add
an exact fixture allowlist, forbid bootstrap tests from importing or copying
production `internal/packs/core`, and run a pre-slice inventory for every old
`internal/bootstrap/packs/core` reference.

### [Blocker] Docs and Operator Vocabulary Are Too Late and Too Imprecise
**Sources:** Felix Moreau; Sofia Khoury; Avery McAllister; Marcus Driscoll
**Issue:** Release-critical docs, doctor output, import-state messages,
examples, generated schema/reference files, public Gastown comments, and
troubleshooting text are still deferred after behavior-changing slices. The
canonical terminology source is not concrete: `docs/reference/system-packs.md`
is named but does not exist or appear in navigation, and the design has
multiple non-identical term lists.
**Required change:** Create and register the canonical system-pack reference or
designate an existing registered page. Define one wording-matrix artifact with
schema, owners, allowed contexts, and consumers in both repos. Replace token
grep with manifest-derived inventory/lint that covers MD/MDX, TOML, Go, shell,
generated schemas/reference docs, navigation, CLI help, doctor strings,
examples, scripts, pack comments, public-pack docs, moved asset names, retired
paths, and store/Dolt maintenance disambiguation. Land docs and goldens in the
first slice that changes operator-facing behavior or mark the sequence
explicitly non-release.

### [Major] Slice Gates Are Not Specific Enough for Risky Intermediate States
**Sources:** Tomas Park; Yuki Hayashi; Ritu Raman; Marcus Driscoll
**Issue:** Several slices name broad suites while changing loader, doctor,
runtime-state, cache, registry, materialization, public-pack, and source
deletion behavior. Reviewers need exact green intermediate states, especially
for slice 5 example `../maintenance` decoupling, registry/materializer updates,
no-Maintenance production loading, public-pin current-loader overlap, and
post-deletion runs with local Gastown/Maintenance sources absent.
**Required change:** Add machine-checkable slice gates: zero duplicate active
definitions in the public-pin/current-loader window, focused package targets
for every loader and doctor surface, `go test ./internal/packs/core` once it
exists, process/integration shard gates for doctor/runtime/materialization, and
post-deletion packcompat/public-import tests plus stale-path scans.

### [Major] Rollback, Offline, and Old/New Binary Semantics Are Not Executable
**Sources:** Yuki Hayashi; Marcus Driscoll; Felix Moreau; Tomas Park
**Issue:** The design acknowledges cache-cold fetch, old binary plus new pack,
new binary plus old pack, rollback after doctor fixes, migrated locks, old
synthetic caches, runtime-state moves, and air-gapped installs, but does not
turn them into a concrete fixture set or explicit one-way boundary.
**Required change:** Name the compatibility binary set or tagged builds and add
fixtures for old/new pack/binary combinations, offline no-cache and cache-hit
states, old synthetic cache transition, registry/direct source identity,
downgrade after doctor fix, and runtime-state rollback. If any boundary is
one-way, document it as such and provide concrete recovery steps with preserved
files.

### [Major] Public Gastown and Core Dependency Model Is Not Settled
**Sources:** Avery McAllister; Ingrid Kovac; Felix Moreau; Marcus Driscoll
**Issue:** The design still does not choose the exact public Gastown dependency
model: no explicit Core import with patches against auto-included Core, an
explicit Core import, or public Gastown owning dog-related behavior. This
affects `[imports.maintenance]` removal, `[[patches.agent]] dog`, moved order
qualification, host-Core patch diagnostics, pack comments, docs, and no-Core or
incompatible-Core behavior.
**Required change:** Specify the public Gastown `pack.toml` model and test it:
host-Core patching if chosen, clear diagnostics when host Core is absent or
incompatible, removal/rejection of surviving Maintenance imports, ownership of
Dog notification/requester flows and fields, moved-order pool qualification,
and named tests for public Gastown commands and doctor checks.

### [Minor] Current Persona-Synthesis Artifacts Are Still Stamped Into the Wrong Directory
**Sources:** Global synthesizer assessment; persona synthesis artifacts
**Issue:** The attempt-9 global synthesis expects attempt-local persona
syntheses, but the complete ten-file set is under
`.gc/design-reviews/ga-2404qu/attempt-1/persona-syntheses/` while
`.gc/design-reviews/ga-2404qu/attempt-9/` contains only four raw review files.
The prior synthesis noted the same path-stamping issue. The artifacts are
present and current, so synthesis can proceed, but automated discovery remains
fragile.
**Required change:** Fix persona-synthesis output path calculation so review
attempt N writes to `attempt-N/persona-syntheses/`. Add a pre-global-synthesis
assertion that every required persona synthesis exists under the current
attempt directory and matches the current attempt metadata.

## Disagreements

- The global persona verdict is unanimous: all ten persona syntheses return
  `block`. Disagreements are inside individual persona lanes, where some raw
  model reviews approved while other models identified blocker-level gaps.
- Elias's Gemini source approved the required-Core plan, but Claude, Codex, and
  DeepSeek found contradictions between the intended dual-gate model and the
  implementable loader, scanner, and doctor sections. Assessment: the stronger
  model must be written as the implementation contract before this lane can
  approve.
- Avery's and Tomas's Gemini sources read later contracts as sufficient for
  host-Core patching, no-Maintenance verification, and duplicate-definition
  avoidance. Other sources found those contracts not yet bound to exact loader
  paths, slice gates, and behavior execution. Assessment: the design should
  keep the stronger intent but make it slice-owned and test-owned.
- Several lanes allow multiple safe choices for `dog`, `crew`, public Gastown
  Core dependency, `GC_BOOTSTRAP=skip`, rollback boundaries, and old cache
  handling. Assessment: the blocker is not that any one choice is mandatory,
  but that the design must choose and prove one.
- Reviewers differ on whether some items are implementation risks or design
  blockers. Assessment: because this review is gating an implementation plan,
  unresolved transaction boundaries, loader bypass contracts, source identity,
  role-surface ownership, and release-critical docs timing are design blockers.

## Missing Evidence

- A concrete `RequiredSystemPackParticipation` type or equivalent, its
  producer/assertion helpers, diagnostic IDs, fatality semantics, and exact
  production wrapper call sites.
- A complete production loader inventory across `cmd/gc` and relevant
  `internal/` packages with allowlist records and focused tests.
- Duplicate-definition matrices for bundled force-required Maintenance plus
  candidate public Gastown, old loader plus new pin, current loader plus
  candidate pin, stale local imports, and rollback states.
- A single retired-source classifier wired into all load, cache, lock, import,
  packman, prompt/template, and doctor paths.
- A doctor mutation coordinator with staged overlay or recoverable transaction
  proof, byte-preserving TOML edit mechanism, controller-active policy, and
  failure-injection suite.
- A Go and asset role-surface migration table, including `dog`, `crew`, tmux
  display, sling branch heuristics, default init, prompt fallback, warmup
  recipients, and Core notification targets.
- A behavior manifest witness-floor rule, semantic-delta schema, generator
  command, whole-repo/public-pack discovery, per-code-path rows, and
  execution-level packcompat fixtures.
- Final `GC_BOOTSTRAP=skip` semantics, production empty `fs.FS` implementation
  details, exact bootstrap fixture allowlist, and old-path inventory.
- Public Gastown dependency model, canonical source string usage, protected
  commit reachability, registry/direct source hash equality, and deterministic
  CI acquisition of the pinned public pack.
- A wording matrix artifact, canonical docs page, public-pack terminology lint,
  manifest-derived stale-reference inventory, runtime-state documentation
  table, and slice-timed docs/golden updates.
- Machine-checkable test migration table replacing count/name assertions with
  behavior assertions or approved retirement rationale.
- Old/new binary, offline/cache, air-gapped, stale cache, and rollback fixtures.

## Recommended Changes

1. Rewrite the required-Core loader and doctor contract around two fatal gates:
   strict required-pack integrity before resolution and typed participation
   after resolution, with a complete production loader inventory and scanner.
2. Resolve public Gastown rollout identity and overlap: choose the pin strategy,
   retire or bypass synthetic public aliases at pin adoption, prove ordinary
   remote `//gastown` checkout, and add zero-duplicate tests for current and old
   loaders before Maintenance removal hides the overlap.
3. Define the retired-source classifier and wire it through config load,
   import/cache/lock/packman, prompt/template discovery, and pre-resolution
   doctor before preserving stale directories on disk.
4. Specify the doctor mutation coordinator, commit boundary, TOML preservation
   mechanism, controller-active policy, staged Core materialization, and
   failure-injection gates before creating implementation beads.
5. Add a Go and asset role-surface migration table and settle `dog`, `crew`,
   tmux display, sling branch metadata, non-Gastown init, prompt fallback, and
   Core notification targets with tests.
6. Strengthen behavior preservation with execution-level witness floors,
   behavioral-trigger packcompat fixtures, semantic-delta approvals, whole-repo
   and public-pack source discovery, authorship identity tests, and per-code-path
   manifest rows.
7. Decide and test `GC_BOOTSTRAP=skip`, the production empty `bootstrapAssets`
   implementation, fixture isolation rules, production Core fixture leakage
   guard, and old bootstrap Core path inventory.
8. Move release-critical docs, canonical terminology, public-pack wording lint,
   stale-reference inventory, generated docs/schema updates, and golden outputs
   into the first operator-facing behavior-changing slice.
9. Make every high-risk slice a coherent green intermediate with exact package,
   process, integration, packcompat, old/new binary, offline/cache, and
   post-deletion gates.
10. Fix the workflow artifact path bug so persona syntheses are written and
    validated under the current attempt directory before global synthesis.
