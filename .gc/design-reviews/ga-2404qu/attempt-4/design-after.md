---
plan_slug: core-gastown-pack-migration
phase: design
rig: gascity
rig_root: /data/projects/gascity-fresh-main-20260604-VLKm8c
artifact_root: /data/projects/gascity-fresh-main-20260604-VLKm8c/plans
requirements_file: /data/projects/gascity-fresh-main-20260604-VLKm8c/plans/core-gastown-pack-migration/requirements.md
status: draft
created_at: 2026-06-04T15:07:35Z
updated_at: 2026-06-07T02:23:08Z
---

# Design: Core and Gastown Pack Split

## Summary

Move Gas City's required Core pack to `internal/packs/core`, retire the
standalone Maintenance pack, and make `gascity-packs/gastown` the only source
of Gastown behavior. Core remains a required system pack that `gc` materializes
and injects into normal config loads. Gastown remains an explicit external pack
import.

The migration has two coordinated tracks:

- Gas City repo: own Core, pack loading, doctor/fix behavior, source-tree
  cleanup, and tests/docs that describe system packs.
- `gascity-packs` repo: own Gastown roles, formulas, orders, scripts, prompts,
  and any Gastown behavior stripped from Core-bound assets.

Core is mandatory for real cities. There is no production opt-out. Tests and
low-level config fixtures may bypass Core only by using lower-level loading
helpers that do not represent the normal `gc` runtime path.

Gastown behavior preservation is a blocking prerequisite, not a retrospective
cleanup item. Before Gas City removes or generalizes any in-tree Gastown or
Maintenance behavior, the matching `gascity-packs/gastown` change must exist at
an immutable commit, Gas City must pin that exact commit through
`internal/config/PublicGastownPackVersion`, and CI must prove the pinned public
pack preserves every behavior edge that Gas City is about to stop owning.
<!-- REVIEW: added per blocker-behavior-preservation -->

### Review-Gated Migration Invariants
<!-- REVIEW: added per blocker-rollout-and-test-slicing -->

Every implementation slice must satisfy these invariants before the next slice
starts:

- The tree is test-green after the slice-specific focused suite and
  `make test-fast-parallel`.
- Normal production `gc` config loading proves Core was materialized and
  included in resolved provenance.
- Public Gastown behavior that will be removed from Gas City has already landed,
  has a generated manifest row, and is verified from the exact pinned commit.
- `gc doctor --fix` is failure-atomic and byte-identical on healthy manifests.
- Stale `.gc/system/packs/maintenance` and `.gc/system/packs/gastown`
  directories are ignored, diagnosed, and preserved, not deleted.
- Core assets remain role-neutral outside explicitly allowed Core maintenance
  configuration.

### Attempt 3 Review Resolution Contracts
<!-- REVIEW: added per blocker-behavior-preservation -->
<!-- REVIEW: added per blocker-required-core-loading -->
<!-- REVIEW: added per blocker-doctor-safety -->
<!-- REVIEW: added per blocker-core-role-neutrality -->
<!-- REVIEW: added per blocker-pack-registry-cache -->
<!-- REVIEW: added per blocker-maintenance-retirement-runtime -->
<!-- REVIEW: added per blocker-bootstrap-fixture -->
<!-- REVIEW: added per blocker-docs-dx-consistency -->
<!-- REVIEW: added per blocker-rollout-and-test-slicing -->
<!-- REVIEW: added per major-behavior-tests -->
<!-- REVIEW: added per major-provider-pack-continuity -->
<!-- REVIEW: added per major-cross-pack-ownership -->

The implementation beads must treat the following contracts as gates. They are
not optional documentation polish; a slice that cannot satisfy one must stop and
record the blocker before moving or deleting behavior.

#### Source-Derived Behavior Manifest

The behavior manifest is generated from source discovery, not hand-curated
memory. The generator walks every old Gas City behavior-bearing source under
Core, Maintenance, Gastown examples, hook overlays, formulas, orders, prompts,
skills, shell scripts, doctor strings, route metadata, notification templates,
runtime state helpers, and tests. It also follows helper references from those
assets so behavior implemented by shared scripts or prompt fragments gets its
own row.

Each manifest row must contain:

- stable row id
- old owner, old path, asset kind, and helper dependencies
- trigger, requester, detector, route metadata, mail/nudge target, prompt
  fragment, script branch, runtime state path, or named-session behavior
- old witness: test, fixture, golden output, command transcript, or explicit
  source assertion
- new owner: Core, public Gastown, provider pack, docs-only, or intentionally
  retired
- new path and new witness
- immutable `gascity-packs` commit when the row moves to public Gastown
- consuming `internal/config/PublicGastownPackVersion` value when Gas City
  stops owning the old behavior
- semantic-equivalence assertion, or an approved delta/removal record with
  owner, reason, and replacement/operator impact

CI must fail when a moved, split, generalized, retired, or helper-dependent
asset lacks a row, when a row lacks old and new witnesses, or when a semantic
delta lacks an approved record. `test/packcompat` consumes this manifest and
executes one assertion per row against the exact public pin before any dependent
Gas City source move, role-neutral rewrite, registry removal, or source
deletion lands.

#### Required Core Identity And Loader Contract

Required Core proof must be typed and content-backed; path or pack-name string
matches are not enough. Normal runtime config loading returns a
`RequiredSystemPackParticipation` record for every required system pack with at
least:

- pack id (`core`, `bd`, or `dolt`)
- materialized system-pack directory
- embedded source id and content manifest digest
- validated file set and pack.toml digest
- resolved-config layer id and import edge proving the pack participated in the
  final resolved config, even if the pack contributes no agents
- repair status and freshness timestamp for the generated system pack

Production loaders fail closed if:

- a required system pack is missing, corrupt, stale, partial, shadowed, or has
  unexpected effective files;
- a user or imported pack is named `core` and would collide with required Core;
- Core was materialized but no typed participation record appears after normal
  config resolution;
- a production `cmd/gc` path reaches `config.Load*` without the required-system
  wrapper or a documented partial-read allowlist entry.

The production loader scanner starts from an inventory of current call sites,
including controller reload, API routing/state, import and doctor diagnostics,
init readiness, convoy/sling/wait flows, and Dolt publication. Each allowlisted
partial read names the file, function, reason, and focused test proving it does
not represent normal runtime config resolution.

#### Doctor And Import-State Safety

`gc doctor --fix` may not mutate anything until all preflight checks pass:
public Gastown source reachability, immutable pin installability, lock
generation, Core materialization and content validation, parseable lockfiles,
editable manifests, and generated-source provenance for any import it plans to
remove or rewrite.

The concurrent-mutation policy is:

- If a controller for the same city is running, discovered from live runtime
  state rather than status files, automatic fix refuses with manual guidance.
- Immediately before every temp-file rename, doctor re-reads the target file and
  fails if its content changed since preflight.
- Multi-file fixes are staged so a failed preflight leaves all manifests,
  lockfiles, and installed pack directories byte-identical.

Doctor must preserve operator content. If scoped TOML editing cannot preserve
comments, unknown tables, unknown fields, array order, formatting, and unrelated
lock entries, the automatic fix refuses instead of whole-file re-encoding.
Generated/example Maintenance imports can be removed only when provenance proves
they are generated or known examples. Custom forks and edited local packs are
diagnostic/manual. After any fix, doctor revalidates typed Core participation and
the ordinary remote install of the exact public Gastown pin.

#### Role-Neutral Core And SDK Self-Sufficiency

Role neutrality covers Go and assets. A scanner must inspect Go, TOML, shell,
Markdown, templates, generated command text, API classifications, tmux theme
helpers, default scaffolding, warmup mail defaults, prompt fallbacks, formula
name heuristics, overlays, metadata, tests, and docs. Core-owned behavior may
not contain or branch on `mayor`, `deacon`, `witness`, `refinery`, `polecat`,
`boot`, `crew`, or `gastown` except in reviewed historical/test allowlist rows.

`dog` is allowed only as Core pack configuration for a configurable maintenance
worker. It is not an SDK primitive and Go must not require it for controller
infrastructure. Tests must prove:

- Core-only cities load and normal SDK infrastructure works with the
  maintenance worker renamed.
- Core-only cities load and controller-owned SDK operations still work when the
  maintenance worker is omitted.
- Core maintenance formulas that do need a worker resolve the configured worker
  name from pack/config data rather than a Go constant.

#### Public Gastown Pin, Registry, And Cache Gate

No Gas City behavior removal, Core generalization, registry alias retirement, or
source deletion may land before Gas City already pins the replacement public
Gastown commit. The pin gate requires:

- immutable public `gascity-packs/gastown` commit recorded in the manifest
- `internal/config/PublicGastownPackVersion` updated to that exact commit
- ordinary remote-pack install proof, not bundled synthetic cache proof
- old-binary/new-pack compatibility proof for the last supported released
  binary
- stale synthetic cache rejection for historical Gastown and Maintenance aliases
- offline behavior decision: either network-required fresh init with explicit
  diagnostics, or a tested compatibility cache keyed to the same public commit

Registry/cache removal is a later cleanup after the pin gate is green. It must
not be the first point where `PublicGastownPackVersion` changes.

#### Maintenance Runtime And Duplicate-Order Contract

Maintenance retirement includes runtime state and order identity, not only pack
imports. The implementation must add a state table covering at least:

| State or surface | Required decision |
| --- | --- |
| `.gc/runtime/packs/maintenance` | legacy root preserved; selected state migrates to Core; conflicts are manual; never deleted by `doctor --fix`. |
| JSONL archive state | migrates to `.gc/runtime/packs/core/jsonl-export-state.json` and `.bak` when Core destination is absent. |
| export cursors | preserved inside the migrated JSONL state file. |
| spawn-storm ledgers | migrates to `.gc/runtime/packs/core/spawn-storm-counts.json` when Core destination is absent. |
| order tracking and skips | preserved names, aliases, or explicit migration tests. |
| Maintenance-owned orders | Core generic owner or public Gastown owner, with no duplicate active order definitions. |
| public Gastown host-Core dependency | explicit no-import host dependency with patch behavior documented and tested. |

The Core/Maintenance fold must be atomic at the behavior level: no intermediate
slice may expose duplicate active order definitions or two owners for the same
script/formula behavior. Tests must assert zero duplicate active orders after
Core and public Gastown are loaded together.

#### Bootstrap And Fixture Contract

Production bootstrap must not embed or copy production Core through
`internal/bootstrap/packs/core`. Production `BootstrapPacks` remains empty.
Bootstrap tests use an explicit test-only fixture and the fixture is named
`core` only for a documented legacy-identity assertion.

`GC_BOOTSTRAP=skip` is not a production escape hatch from required Core. If the
variable remains, its semantics are narrowed to bootstrap fixture setup only and
normal `gc` config loading still materializes and validates Core. CI must fail
on production embeds of `packs/**`, old `AssetDir: "packs/core"` dependencies
outside the fixture allowlist, or hook overlays loading from the old bootstrap
Core path.

#### Docs, Tests, And Provider Integrity Gates

Docs are finalized only after runtime state, `dog`, order naming, public-source
semantics, and Maintenance retirement behavior are decided. The docs inventory
must cover Markdown, generated references, CLI help, examples, scripts, doctor
output, schema/reference text, tutorials, and troubleshooting. A navigable
`docs/reference/system-packs.md` or replacement canonical reference must exist,
be linked from docs navigation, and use the canonical wording matrix.

Path, count, and name assertions are insufficient. Replacement tests must prove
formula composition, molecule step construction, hook target resolution,
configured-agent/session loading, order ownership, prompt/template resolution,
pack-relative script execution, doctor idempotency, import-state diagnostics,
fresh init, upgraded-city behavior, and public Gastown compatibility.

Provider pack continuity is a gate for `bd` and `dolt`: bytes, provenance,
install locks, include behavior, formula/order resolution, and materialization
must remain correct when Core is repaired and Maintenance is absent. Required
system-pack integrity must either validate the full file set or prove unexpected
files cannot influence loaded formulas, orders, scripts, overlays, prompts, or
config.

### Attempt 4 Review Resolution Contracts
<!-- REVIEW: added per attempt4-global-blockers -->

Attempt 4 resolves the remaining design choices instead of carrying them as
implementation-time questions. These contracts supersede any earlier "decide
later" wording in this document.

#### Executable Source-Discovery Manifest
<!-- REVIEW: added per attempt4-behavior-preservation-gate -->

The canonical behavior manifest is machine-readable and source-derived. Human
documentation is rendered from it; it is not the source of truth.

Canonical Gas City path:
`plans/core-gastown-pack-migration/behavior-manifest.generated.yaml`.

Canonical public Gastown path:
`gastown/docs/behavior-manifest.generated.yaml`.

The generator must discover behavior-bearing assets from these roots before any
Core move, Maintenance fold, public pin update, registry cleanup, or source
deletion is allowed:

- `internal/bootstrap/packs/core`
- `examples/gastown/packs/maintenance`
- `examples/gastown/packs/gastown`
- `examples/gastown/*_test.go`
- `cmd/gc` config-load, doctor, order, formula, sling, hook, and prompt tests
- `test/packlint`
- shell helpers and template fragments referenced from any discovered TOML,
  Markdown, prompt, formula, order, or script

Discovery follows references instead of only scanning directories. Formula
`script`, template `{{template ...}}`, prompt-fragment, order, hook overlay,
mail/nudge target, requester/detector, runtime-state path, named-session, and
route metadata references each become rows or row dependencies.

Each generated row must contain:

- stable id derived from old owner, path, asset kind, and referenced symbol
- old path and old witness test or source assertion
- behavior trigger and observable output
- helper dependencies and runtime-state paths
- final owner: Core, public Gastown, provider pack, docs-only, or retired
- final path and final witness test
- required public Gastown commit when final owner is public Gastown
- consuming Gas City `PublicGastownPackVersion` when Gas City stops owning it
- explicit semantic delta/removal approval when behavior is not preserved

CI must fail if the generated manifest differs from the checked-in manifest,
if a moved/split/generalized/retired/helper-dependent asset lacks a row, or if
any row lacks both old and final witnesses. `test/packcompat` consumes the
manifest and executes one assertion per executable row. Path/count/existence
checks are not enough; rows must exercise the original trigger through normal
config resolution and verify the same observable behavior or the approved
delta.

Removed tests also require rows. `examples/gastown/maintenance_scripts_test.go`
is split into generated groups before implementation starts: Core-bound script
tests, public-Gastown script tests, provider-pack tests, and approved removals.
No test function from that file may disappear without a manifest row naming its
replacement test or removal approval.

#### Public Gastown Host-Core Contract
<!-- REVIEW: added per attempt4-public-gastown-host-core-contract -->

Chosen model: Core is an auto-included host system pack owned by the Gas City
binary. Public Gastown does not import Core. Public Gastown may patch host Core
configuration only through the normal resolved-config patch mechanism after
required system packs have been inserted into the base layer.

Config layer order for a Gastown city is:

1. required host system packs: Core plus provider packs
2. city and rig imports, including public Gastown
3. patches and overrides from later layers, including public Gastown patches
   against Core-defined agents
4. city and rig local overrides

`[[patches.agent]] name = "dog"` in public Gastown is valid only because `dog`
is a Core pack agent in layer 1. The patch target is the resolved agent key
after required system packs participate in config resolution. If Core is absent
or the target agent is renamed in Core without a compatibility alias, config
loading fails with a host-Core diagnostic; public Gastown does not silently
create a replacement Core agent.

Required tests:

- host Core present + public Gastown: `dog` patch applies and the effective
  work directory/theme fields match the Gastown patch.
- host Core absent: load fails with "required host Core missing" before any
  Gastown behavior executes.
- no Maintenance present: public Gastown plus host Core resolves agents,
  prompts, formulas, hooks, and orders through the normal production loader.
- stale local Gastown/Maintenance import: doctor can diagnose and fix before
  full config resolution.
- custom Core import named `core`: load fails as an unsupported collision.

Template fragments are owned by the pack whose prompt references them. Core
`dog` prompts may reference only Core-owned fragments; public Gastown prompts
may reference only public-Gastown fragments. Shared fragments must be duplicated
or moved to a provider-neutral Core fragment with an explicit Core prompt
reference; there is no implicit cross-pack template lookup from Gastown to Core
or Core to Gastown. `test/packcompat` must compile every Core and Gastown prompt
template after the split and fail on unresolved fragments.

Public Gastown pack scans must fail on hardcoded retired paths:
`.gc/system/packs/maintenance`, `.gc/runtime/packs/maintenance`,
`packs/maintenance`, and `../maintenance`, except in migration docs or tests
listed in the manifest allowlist.

#### Pre-Resolution Doctor And Legacy Import Recovery
<!-- REVIEW: added per attempt4-doctor-pre-resolution-safety -->

`gc doctor` gets a pre-resolution import migration phase that reads raw TOML
for `city.toml` and rig `pack.toml` files before full config expansion. This
phase is allowed to parse only import tables, source strings, versions, and
comments needed for a preserving rewrite. It must not evaluate pack formulas,
agents, hooks, templates, or orders.

The pre-resolution phase handles:

- local generated/system Gastown imports pointing at `.gc/system/packs/gastown`
  or known generated `examples/gastown/packs/gastown`
- local generated/system Maintenance imports pointing at
  `.gc/system/packs/maintenance` or known generated
  `examples/gastown/packs/maintenance`
- relative `../maintenance` imports inside legacy local Gastown packs
- stale locks for public Gastown commits older than
  `internal/config/PublicGastownPackVersion`

Fix behavior:

- Generated or known example Gastown imports are rewritten to
  `github.com/gastownhall/gascity-packs/gastown` at
  `PublicGastownPackVersion` only after the remote commit is reachable,
  installable, and lockable.
- Active local development paths are not rewritten automatically. A path under
  the current checkout, a dirty worktree, a non-generated pack, or a source
  outside the known generated/system set is reported as manual.
- Generated or known example Maintenance imports are removed only after Core is
  materialized and public Gastown lock/install validation succeeds.
- Relative `../maintenance` inside legacy Gastown is tolerated during the
  pre-resolution diagnostic pass only. Normal production config loading still
  fails on unresolved imports after doctor has had the chance to report the
  actionable fix.
- TOML edits are scoped and preserving. If comments, unknown tables, array
  order, or formatting cannot be preserved, `gc doctor --fix` refuses and gives
  manual steps.

All mutating fixes are preflight-first and failure-atomic. Doctor discovers
live controllers from runtime facts, not PID/status files, and refuses automatic
multi-file fixes while a controller for the same city is running. Immediately
before each rename, doctor re-reads the target file and aborts if it changed
since preflight.

#### Required Core Loader Bypass Inventory
<!-- REVIEW: added per attempt4-loader-callsite-inventory -->

The production loader scanner starts with this call-site inventory and shrinks
it during the Core loading slice. New direct production `config.Load*` calls are
forbidden unless they are added to the partial-read table with a reason and a
focused test.

| Current surface | Required disposition |
| --- | --- |
| `cmd/gc/cmd_import.go` import-edit reads | Use raw edit loader or pre-resolution import parser; never behavior-driving config. |
| `cmd/gc/init_provider_readiness.go` readiness probes | Route through required-system-pack wrapper unless explicitly checking partial config. |
| `cmd/gc/cmd_config.go` quick display reads | Split into `loadCityConfigRuntime` and `loadCityConfigPartialForExplain`. |
| `cmd/gc/cmd_supervisor_city.go` reload reads | Route normal reload through required-system-pack wrapper and assert Core participation. |
| `cmd/gc/cmd_rig.go` post-edit reloads | Route behavior-affecting reloads through wrapper; raw edit path remains partial. |
| `cmd/gc/cmd_init.go` template/copy validation | Use a testable wrapper for initialized cities; allow raw reads only before a city exists. |
| `cmd/gc/cmd_wait.go` legacy sanity load | Convert to wrapper or delete if redundant with API fallback. |
| `cmd/gc/dolt_runtime_publication.go` publication reads | Route through wrapper; provider pack participation is part of the proof. |
| `cmd/gc/cmd_start_drift.go` drift checks | Partial-read allowlist only if the file can be broken and still needs diagnostics. |
| `cmd/gc/legacy_pack_preflight.go` legacy import preflight | Replace with the pre-resolution TOML import parser. |

The refined interface names are:

- `loadCityConfigRuntime`: normal command path, materializes required packs and
  asserts participation.
- `loadCityConfigNoRefreshRuntime`: controller/reload path, validates existing
  required packs and asserts participation.
- `loadCityConfigPartialForEdit`: raw TOML edit path, no behavior decisions.
- `loadCityConfigPartialForDoctor`: pre-resolution diagnostic path, import
  tables only.

Scanner tests must reject direct calls to `config.Load`, `config.LoadCity`, and
`config.LoadWithIncludes` in non-test `cmd/gc` files outside these helpers.
They must also reject production use of non-OS filesystems for runtime loads.

#### Bootstrap Extraction Completion Contract
<!-- REVIEW: added per attempt4-bootstrap-fixture-isolation -->

Production bootstrap owns no Core assets and still must have a non-nil asset
filesystem. After removing `//go:embed packs/**`, `bootstrapAssets` defaults to
a private empty `fs.FS` implementation that returns `fs.ErrNotExist` for all
paths. It must never be `nil`.

Bootstrap tests use inline `fstest.MapFS` fixtures, not copied production Core
directories and not an on-disk `testdata/packs/core` tree. Fixture `Entry.Name`
may remain `core` when testing collision semantics, but `AssetDir` points at the
inline fixture path; the test must not imply production Core is still under
bootstrap.

Required changes in the Core extraction slice:

- remove `internal/bootstrap/packs/core/embed.go` and its dual-embed comments
- remove `"core"` and `"registry"` from `bootstrapManagedImportNames` in
  `internal/config/compose.go` when `BootstrapPacks` is permanently empty
- update the sync test so bootstrap-managed implicit imports are empty
- update `cmd/gc/prompt_test.go`,
  `internal/config/bundled_import_test.go`,
  `examples/gastown/precompact_hook_test.go`,
  `test/packlint/*`, and `internal/hooks/config/README.md` for the new
  `internal/packs/core` path or fixture model
- add `TestProductionBootstrapAssetsIsEmpty`
- add `TestBootstrapFixtureIsMinimal`, failing if fixture assets contain
  production-only directories such as `formulas/`, `orders/`, `overlay/`,
  `skills/`, or `assets/prompts/`

`GC_BOOTSTRAP=skip` is kept only as a bootstrap compatibility switch. It skips
legacy bootstrap materialization, which is empty in production. It must not
skip `MaterializeBuiltinPacks`, required Core materialization, required Core
integrity validation, or config provenance checks. Tests must prove a command
with `GC_BOOTSTRAP=skip` still materializes and verifies Core through the normal
system-pack path.

#### Concrete Runtime-State Migration
<!-- REVIEW: added per attempt4-maintenance-runtime-state -->

Retired Maintenance runtime state migrates to Core for continuity. The old
Maintenance path is never deleted automatically.

| Legacy state | New canonical state | Migration behavior |
| --- | --- | --- |
| `.gc/runtime/packs/maintenance/jsonl-export-state.json` | `.gc/runtime/packs/core/jsonl-export-state.json` | `gc doctor --fix` and the first Core `jsonl-export.sh` run copy/rename when Core destination is absent; if both exist and differ, report manual conflict. |
| `.gc/runtime/packs/maintenance/jsonl-export-state.json.bak` | `.gc/runtime/packs/core/jsonl-export-state.json.bak` | Same rule as primary state. |
| `.gc/runtime/packs/maintenance/jsonl-archive/` | `.gc/runtime/packs/core/jsonl-archive/` | Move only when Core destination is absent; preserve git remotes and refs; if both exist, diagnose and do not merge. |
| JSONL push fields: `pending_archive_push`, `consecutive_push_failures`, `push_failure_escalated`, `last_push_at`, `last_push_stderr` | Same keys in Core state file | Preserve exactly through the state migration. |
| `.gc/runtime/packs/maintenance/spawn-storm-counts.json` | `.gc/runtime/packs/core/spawn-storm-counts.json` | Move when Core destination is absent; if both exist, Core wins and legacy is reported as ignored. |
| order skip/tracking beads for `mol-dog-*` orders | existing bead metadata/order names | Preserve order names or add aliases; no runtime file migration. |
| script temp state passed explicitly by formula env | unchanged unless the script moves to Core-owned `GC_PACK_STATE_DIR` | Keep behavior row and test; do not silently reinterpret custom env paths. |
| `GC_PACK_STATE_DIR` for Core-owned moved scripts | `.gc/runtime/packs/core` | Formula/order environment sets the Core path; scripts default to Core, not Maintenance. |
| doctor JSONL fallback paths | Core first, then legacy Maintenance, then `.gc/jsonl-*` compatibility | `jsonl_archive_doctor_check.go` updates `resolveStateFile` and `resolveArchiveRepo` in the Maintenance folding slice. |

This is a one-way runtime-state migration for new binaries. Rollback guidance
must say that old binaries may continue reading legacy Maintenance state until
the operator manually moves Core state back; doctor-mutated city manifests must
remain readable by old binaries, but runtime-state rollback is advisory.

#### Go Role-Neutrality Scope And Scanner
<!-- REVIEW: added per attempt4-role-neutrality-scope -->

Go de-roling is in scope for SDK behavior, default scaffolding, and Core-owned
surfaces affected by this migration. Historical docs, Gastown examples, and
tests that explicitly model user-configured roles may remain only through an
allowlist.

The scanner is a Go test that walks Go, TOML, shell, Markdown, templates, docs,
generated CLI help fixtures, and Core assets. It tokenizes identifiers and
string literals and rejects behavior-bearing references to `mayor`, `deacon`,
`witness`, `refinery`, `polecat`, `boot`, `crew`, and `gastown` outside the
allowlist. `dog` is allowed only for Core maintenance-agent configuration and
explicit compatibility aliases.

The migration inventory must include:

- tmux theme/icon APIs such as `DogTheme`
- default city scaffolding and warmup defaults
- prompt fallback and generated prompt-help examples
- `internal/sling` formula-name heuristics
- `classifyAgentKind`
- mail/nudge targets and notification examples
- role-bearing TOML descriptions and generated docs

If an implementation slice cannot remove a role surface safely, it must narrow
the design claim for that slice and add an allowlist row naming the owner,
reason, and removal follow-up.

#### Rollout Gate Repairs
<!-- REVIEW: added per attempt4-rollout-gate-repairs -->

All cross-pack ownership audits happen before the public Gastown commit is
pinned. Slice 1 cannot finish until `mol-review-quorum`, provider overlays,
Dog prompt fragments, Polecat formulas, branch pruning, shutdown-dance examples,
review checks, and hardcoded role-theme/tmux APIs have final Core/Gastown
ownership rows and replacement tests.

`test/packcompat` is introduced in the public-pin slice, but it has two modes:

- current-loader compatibility: runs while the old binary still force-includes
  Maintenance, proving the public pin is backward-compatible
- no-Maintenance production-loader compatibility: runs in the first slice that
  removes Maintenance from `requiredBuiltinPackNames`, after Core-owned
  Maintenance behavior has moved to Core

The design must not claim a no-Maintenance production-loader gate has passed
until the normal production loader actually excludes Maintenance. No test-only
loader bypass may be used to satisfy that gate.

Every rollout slice runs `make test-fast-parallel` and `go vet ./...` unless a
slice-specific note names a narrower gate and the next slice cannot start until
the full gate is green. Slice 3 additionally runs `go test ./cmd/gc`,
`go test ./internal/config`, and `go test ./test/packlint/...` because it
updates hardcoded old bootstrap-Core paths outside `internal/bootstrap`.

#### Docs And Operator-DX Anchor
<!-- REVIEW: added per attempt4-docs-operator-dx -->

`docs/reference/system-packs.md` is the canonical operator reference. The docs
lint reads a wording matrix from the behavior manifest and checks doctor output,
CLI help, tutorials, generated docs, pack comments, and script comments against
these terms:

- Core: required host system pack
- `bd` and `dolt`: provider-dependent host system packs
- Maintenance: retired standalone pack
- Gastown: explicit public pack import
- Core `dog`: configurable Core maintenance agent, not an SDK primitive
- stale Maintenance paths: ignored legacy state or migrated runtime state,
  never silently deleted

Version-skew doctor output must compare the locked public Gastown commit with
`PublicGastownPackVersion`. Older pins get an actionable warning and a
non-mutating explanation; `gc doctor --fix` updates the pin only after the same
preflight and preserving-edit rules described above.

## Current System

`internal/builtinpacks/registry.go` defines the embedded pack set:

- `core`, sourced from `internal/bootstrap/packs/core`
- `bd`, sourced from `examples/bd`
- `dolt`, sourced from `examples/dolt`
- `maintenance`, sourced from `examples/gastown/packs/maintenance`
- `gastown`, sourced from `examples/gastown/packs/gastown`

`cmd/gc/embed_builtin_packs.go` materializes all embedded packs into
`.gc/system/packs/{name}`. `requiredBuiltinPackNames` currently always requires
`core` and `maintenance`, then adds `bd` or `dolt` depending on the beads
provider. `builtinPackIncludes` appends those materialized system pack
directories to `config.LoadWithIncludes`, so Core and Maintenance are implicit
normal config layers.

`internal/hooks/hooks.go` imports `internal/bootstrap/packs/core` directly and
reads provider overlays from `core.PackFS`.

`internal/bootstrap/bootstrap.go` still embeds `packs/**` even though
`BootstrapPacks` is empty in production. Several tests override
`BootstrapPacks` with `AssetDir: "packs/core"`, so removing
`internal/bootstrap/packs/core` also requires a replacement test fixture or an
injectable bootstrap test filesystem.

`cmd/gc/import_state_doctor_check.go` currently rewrites legacy local Gastown
imports to `config.PublicGastownPackSource` and removes Maintenance imports
with the message that Maintenance/Core are supplied implicitly. That wording and
fix behavior no longer match the target state.

Fresh Gastown init already writes explicit public `gascity-packs/gastown`
imports through `internal/config.GastownCityWithProviders`, using
`PublicGastownPackSource` and `PublicGastownPackVersion`. The public
`gascity-packs/gastown` pack no longer imports Maintenance, but its comments
and patches still assume the host runtime supplies an implicit maintenance/core
utility layer.

The in-repo `examples/gastown` tree still has local pack sources:

- `examples/gastown/pack.toml` imports `packs/gastown`.
- `examples/gastown/city.toml` default rig imports `packs/gastown`.
- `examples/gastown/packs/gastown/pack.toml` imports `../maintenance`.
- `examples/gastown/packs/maintenance` still owns Dog, cleanup orders, and
  maintenance scripts.

Tests and docs have many path and behavior assumptions:

- `internal/builtinpacks/registry_test.go` pins the embedded pack list and
  synthetic repo paths.
- `cmd/gc/embed_builtin_packs_test.go` asserts Maintenance order paths and
  builtin include counts.
- `cmd/gc/controller_test.go` expects orders from Core, Maintenance, `bd`, and
  `dolt`.
- `cmd/gc/import_state_doctor_check_test.go` expects Maintenance imports to be
  removed because Maintenance/Core are implicit.
- `examples/gastown/gastown_test.go` contains extensive Gastown pack behavior
  tests against local `examples/gastown/packs/gastown`.
- Docs such as `docs/guides/shareable-packs.md`,
  `docs/reference/system-packs.md`, and `docs/getting-started/troubleshooting.md`
  describe implicit Maintenance or `.gc/system/packs/maintenance`.

## Proposed Design

### Pack Ownership

Create `internal/packs/core` as the canonical Core package:

- Move Core pack assets from `internal/bootstrap/packs/core` to
  `internal/packs/core`.
- Add `internal/packs/core/embed.go` with `PackFS` embedding Core assets.
- Update `internal/hooks/hooks.go` and `internal/builtinpacks/registry.go` to
  import `github.com/gastownhall/gascity/internal/packs/core`.
- Remove `internal/bootstrap/packs/core` as an asset source.

Retire the Maintenance package:

- Remove `examples/gastown/packs/maintenance/embed.go`, `pack.toml`, and
  Maintenance-only tests after assets are moved.
- Do not create a replacement `maintenance` system pack.
- Keep "maintenance" Go/config names that refer to supervisor/store
  maintenance out of scope.

Move Gastown assets out of the Gas City source tree:

- Remove `examples/gastown/packs/gastown` as a maintained pack source.
- Keep `examples/gastown` as an example city with public
  `gascity-packs/gastown` imports rather than local `packs/gastown` imports.
- Move pack-specific tests from `examples/gastown/gastown_test.go` to the
  `gascity-packs` repo, or replace them with Gas City tests that assert init and
  import wiring only.

### Core Asset Contents

Core owns:

- CLI usage skills under `skills/gc-*`, after removing Gastown-only examples.
- Provider hook overlays under `overlay/per-provider/*` when they are generic
  Gas City hook behavior.
- Generic formulas such as direct work, scoped work, prompt synthesis, and
  review quorum only when role-neutral.
- Generic infrastructure orders and scripts from Maintenance, including gate
  sweep, route nudge, blocker-close cascade, orphan sweep, order tracking,
  spawn storm detection, wisp compaction, JSONL export, stale cleanup/reaper,
  and binary doctor checks.
- A default configurable `dog` maintenance agent for Core pack operations.
  `dog` is allowed only as pack configuration, not as a Go special case or SDK
  infrastructure dependency.
- `dolt-target.sh` for now, because Dolt remains a Core requirement until bd
  provider support is restored.

Core must not own:

- Mayor, Deacon, Polecat, Refinery, Witness, Boot, Crew, or Gastown role
  behavior.
- Branch pruning. Move `prune-branches.sh` and `orders/prune-branches.toml` to
  `gascity-packs/gastown`.
- Polecat formulas. Move `mol-polecat-base`, `mol-polecat-commit`, and
  `mol-polecat-report` to `gascity-packs/gastown`.
- Gastown examples inside `gc-dispatch` or other Core skills.

Dog-specific Core formulas may keep names such as `mol-dog-jsonl` and
`mol-dog-reaper` for compatibility, because `dog` is now the Core maintenance
agent. The implementation should rename only Gastown role names, not `dog`
itself. If a formula is renamed, add compatibility aliases or migration tests
for existing order skips and references.

`mol-shutdown-dance` should move to Core as generic stuck-session due process
for the Core `dog` agent. Its current detector table and examples mention
Deacon, Witness, Polecat, and Mayor. Remove those from the Core copy, then add
or preserve equivalent Gastown detector/requester instructions in Gastown-owned
prompts or formulas.

`mol-review-quorum` remains a design-time audit item. If its content is
role-neutral, keep it in Core. If it references Polecat, Refinery, Mayor, or
Gastown review flows, move it to `gascity-packs/gastown`.

#### Core Maintenance And Notification Contract
<!-- REVIEW: added per blocker-core-role-neutrality -->

Core maintenance behavior must be executable without hardcoded role knowledge in
Go and without assuming a Gastown role exists. The contract for each Core asset
is:

| Operation kind | Required design decision |
| --- | --- |
| Mail creation | Use configured recipients from the formula, order, or pack configuration; move Gastown-specific recipients to `gascity-packs/gastown`. |
| Nudge/route targets | Resolve through session/worker identity and configured agent names; no Go code may contain `mayor`, `deacon`, `witness`, `refinery`, `polecat`, `boot`, `crew`, or `gastown` as a control decision. |
| Detector/requester examples | Keep only role-neutral examples in Core; preserve Gastown-specific detector/requester prose in Gastown prompts, docs, or formulas. |
| Scripts | Accept target/session/filter values from environment or formula variables; scripts must not branch on Gastown role names. |
| Orders | Either target the Core maintenance agent declared by Core pack config or move the order to Gastown. Controller-owned SDK operations must not depend on this agent. |
| Prompt fragments and skills | Remove Gastown examples or move them to Gastown-owned assets. |

The Core `dog` name is a default pack configuration value for the Core
maintenance agent. Tests must prove a Core-only city can still load, run normal
SDK infrastructure, and evaluate non-agent controller operations when the
maintenance agent is renamed or omitted. Tests that execute Core maintenance
formulas must prove they resolve the configured maintenance agent name rather
than a Go constant.

Add a role-token scanner over Core TOML, shell, Markdown, templates, overlays,
skills, embedded command text, generated manifests, and Core tests. The scanner
must be path-, token-, and field-aware: it may allow historical docs and
explicit Gastown fixture names only through a reviewed allowlist, and it must
fail on new Core-owned behavior references to Gastown roles.

### Gastown Behavior Preservation

Generate the source-derived behavior manifest described in
`Executable Source-Discovery Manifest`, then render the public Gastown operator
view to `gastown/docs/behavior-preservation.md`.

The generated manifest must list every moved or generalized asset with:

- old path
- old behavior trigger, requester, detector, route metadata, notification
  target, script branch, or prompt fragment
- semantic delta, if any
- new Core path, if any
- new Gastown path, if any
- behavior intentionally removed from Core
- replacement Gastown behavior
- old Gas City test
- new `gascity-packs` test
- `gascity-packs` landing commit
- Gas City pin value that consumes the landing commit

This manifest is required for these high-risk moves:

- Maintenance `dog` prompt to Core, with Gastown notification/requester behavior
  preserved in Gastown prompts or formulas.
- `mol-shutdown-dance` to Core, with Gastown detector examples preserved in
  Deacon/Witness/Boot/Gastown documentation.
- `mol-dog-jsonl` and `mol-dog-reaper` to Core, with Mayor/Deacon/Gastown
  notifications preserved or replaced.
- `prune-branches` to Gastown.
- Core `mol-polecat-*` formulas to Gastown.
- Core `gc-dispatch` skill split.

The public `gascity-packs/gastown/pack.toml` should be updated so comments no
longer describe an implicit Maintenance layer. Gastown does not import Core;
Core remains a required host system pack inserted before Gastown in normal
config resolution. Gastown may continue to patch the Core `dog` agent for
theming or work_dir behavior only through the host-Core patch contract above,
and the generated manifest must name that dependency.

Before any Gas City source deletion or Core generalization lands:
<!-- REVIEW: added per blocker-behavior-preservation -->

- The `gascity-packs` PR or branch must be named in the implementation bead.
- The replacement commit must be immutable and recorded in both the manifest
  and Gas City's `internal/config/PublicGastownPackVersion` value.
- The public pack must be fetched from the ordinary remote-pack install path,
  not from a bundled synthetic alias.
- Gas City CI must run a gate that installs that exact commit into a fresh test
  city, composes the moved formulas/orders, resolves moved scripts using
  pack-relative paths, verifies hook overlays and configured agents, and checks
  every generated manifest row.
- The gate must fail if any test from `examples/gastown/gastown_test.go` or
  `examples/gastown/maintenance_scripts_test.go` is removed without an explicit
  row mapping it to a new `gascity-packs` test or a documented intentional
  behavior removal.

Recommended Gas City gate name:
`go test ./test/packcompat -run TestPinnedPublicGastownBehavior`.
At pin-adoption time the test runs in current-loader compatibility mode. Before
Maintenance is removed from required system packs, the same gate must run in
no-Maintenance production-loader mode: clone or install
`github.com/gastownhall/gascity-packs/gastown` at
`PublicGastownPackVersion`, load it with host Core present and no Maintenance
pack through the normal production loader, and assert that the generated
manifest is complete for the removed in-tree assets.

### Builtin Registry And Synthetic Cache
<!-- REVIEW: added per blocker-pack-registry-cache -->

Update `internal/builtinpacks/registry.go`:

- `All()` should return Core, `bd`, and `dolt`.
- Core subpath becomes `internal/packs/core`.
- Remove Maintenance and Gastown from the Gas City embedded set.
- Remove public synthetic aliases for `gastown` and `maintenance`, because
  Gastown is no longer embedded in the Gas City binary.
- Keep synthetic cache validation for the remaining bundled packs.
- Define synthetic cache namespaces as bundled-pack only. Core, `bd`, and `dolt`
  continue to use `SyntheticContentHash()` for bundled content. Public Gastown
  must use the ordinary remote-pack cache path keyed by repository source and
  immutable version.
- `RepoCacheKey` for public Gastown must include the normalized source
  (`github.com/gastownhall/gascity-packs/gastown`) and the exact
  `PublicGastownPackVersion` value. It must not collide with historical
  synthetic aliases for `gastown`.
- Existing synthetic cache entries for public Gastown or Maintenance are stale
  after this migration. They may remain on disk, but source resolution must not
  select them for a public `sha:` pin.
- Fresh `gc init --template gastown` is network-required unless the ordinary
  remote-pack cache already contains the exact public pin. If the public
  Gastown pack cannot be fetched, the error must say fresh Gastown
  initialization is unsupported offline without a pre-populated repository
  cache. Offline fallback to embedded old Gastown is not allowed.

Update tests in `internal/builtinpacks/registry_test.go`:

- Expected identities become `core=internal/packs/core`, `bd=examples/bd`, and
  `dolt=examples/dolt`.
- Source-recognition variants use `internal/packs/core`.
- Tamper and unexpected-file tests write under `internal/packs/core`.
- Public Gastown synthetic cache tests move out of Gas City or change to assert
  no bundled synthetic cache is available for public Gastown sources.
- Add negative tests that `IsSource`, `NameForSource`, install-lock generation,
  and materialization reject retired Maintenance sources and historical public
  Gastown synthetic aliases for new locks.
- Add stale-cache tests where an old synthetic public Gastown cache exists but a
  new `PublicGastownPackVersion` install must fetch or use the ordinary
  remote-pack cache for the pinned commit.

Do not aggressively delete stale `.gc/system/packs/maintenance` or
`.gc/system/packs/gastown` directories on startup. They should simply stop
being generated and stop being included. A separate doctor advisory can report
unused legacy system pack directories if needed. This avoids deleting operator
edits in formerly non-required generated packs.

#### Maintenance Retirement Runtime Table
<!-- REVIEW: added per blocker-maintenance-retirement-runtime -->

| Surface | Target behavior | Required proof |
| --- | --- | --- |
| `requiredBuiltinPackNames` | Requires Core plus provider packs only; never Maintenance or Gastown. | Unit tests for bd, dolt, and exec `gc-beads-bd` provider matrices. |
| `builtinPackIncludes` | Includes generated Core and provider packs; ignores stale generated Maintenance/Gastown directories. | Include-list tests with stale directories present. |
| `publicSubpathForPack` and source recognition | No public synthetic Maintenance; no public synthetic Gastown for new pins. | Negative `IsSource`/`NameForSource` and lock/install tests. |
| `MaterializeBuiltinPacks` | Generates/repairs Core, `bd`, and `dolt`; does not refresh or prune Maintenance/Gastown. | Missing/corrupt Core repair tests plus stale custom-file preservation tests. |
| Orders and scripts | Core-owned generic orders resolve from Core; Gastown-owned orders resolve from public Gastown. | Formula/order composition tests and pack-relative script execution tests. |
| Order skip lists | Preserved names continue to match; renamed orders have aliases or migration tests. | Existing skip-list compatibility tests. |
| Runtime state | JSONL state/archive and spawn-storm ledgers migrate to Core; conflicting legacy state is manual; other stale Maintenance paths are ignored legacy. | Doctor diagnostics and first-run script migration tests; no deletion on `--fix`. |
| Doctor import fixes | Generated/example Maintenance imports removed; custom forks preserved with manual guidance. | Golden TOML preservation and fork-provenance tests. |
| Synthetic cache validation | Remaining bundled packs keep validation; retired aliases are never selected for new resolutions. | Stale synthetic-cache rejection tests. |

### System Pack Loading

Update `cmd/gc/embed_builtin_packs.go`:

- `requiredBuiltinPackNames` starts with `[]string{"core"}`.
- Keep provider-dependent `bd` and `dolt` inclusion as today.
- Remove Maintenance from required pack refresh and `builtinPackIncludes`.
- Update comments from "Core and maintenance" to "Core".
- Keep `MaterializeBuiltinPacks` as the single materialization entrypoint for
  required system packs.

Update config-loading call sites:
<!-- REVIEW: added per blocker-required-core-loading -->

- Normal command paths must use `loadCityConfigWithBuiltinPacks`,
  `cityConfigIncludesWithBuiltinPacks`, or `builtinPackIncludesForConfigLoad`.
- The production wrappers must call a single post-load assertion, recommended
  helper name `assertRequiredSystemPackProvenance`, which verifies the resolved
  config provenance contains the materialized Core system pack path for the
  active city. A successful materialization without resolved provenance is a
  load failure.
- Direct `config.Load`, `config.LoadCity`, `config.LoadWithIncludes`, and
  package aliases of those functions in production `cmd/gc` files should be
  rejected by a scanner test modeled on
  `TestGCNonTestFilesStayOnWorkerBoundary`.
- Any production exception must be listed in an allowlist with the file,
  function, reason, and a focused test proving it intentionally reads
  partial/broken config and does not represent normal runtime config
  resolution.
- No-refresh or diagnostic config helpers must either run the same Core
  provenance assertion after loading or be documented as partial-read exceptions
  in that scanner allowlist.
- Low-level `internal/config` tests may continue to call `config.LoadWithIncludes`
  directly because those tests exercise the config package, not full `gc`
  runtime behavior.

The dev/test escape hatch is not a CLI flag or environment variable. Production
`gc` commands always include Core. Tests that need no-Core behavior should call
the lower-level config loader directly or use a `_test.go` helper that bypasses
`cityConfigIncludesWithBuiltinPacks`.

### Core Presence Doctor

Add a new doctor check, recommended file:
`cmd/gc/core_pack_doctor_check.go`.

The check should:

- Materialize intent: Core is required for real cities.
- Verify `.gc/system/packs/core/pack.toml` exists and matches the embedded Core
  manifest using the same manifest comparison behavior already used by
  `unusableRequiredBuiltinPackNames`.
- Load resolved config through the normal system-pack wrapper and verify
  provenance includes the Core system pack path.
- Return Error or Warning according to the existing doctor severity convention
  for required repairable checks.
- Include details naming the missing condition: missing materialized Core,
  invalid materialized Core, or Core absent from resolved config provenance.
- Set `FixHint` to run `gc doctor --fix`.

`Fix` should:

- Call `MaterializeBuiltinPacks(cityPath)`.
- Re-run the Core manifest/provenance checks.
- Return an error if Core is still absent, because production opt-out is not
  supported.

This fix "adds Core" by repairing the generated system pack and normal include
path. It should not write `[imports.core]` to `city.toml` or `pack.toml`.

Update existing doctor/import-state behavior:
<!-- REVIEW: added per blocker-doctor-safety -->

- In `cmd/gc/import_state_doctor_check.go`, change Maintenance messaging from
  "maintenance/core is supplied implicitly" to "maintenance is retired; Core
  supplies generic maintenance and Gastown supplies Gastown-specific behavior".
- Continue rewriting legacy Gastown local imports to public
  `gascity-packs/gastown`.
- Continue removing legacy Maintenance imports when the source is
  `.gc/system/packs/maintenance` or `examples/gastown/packs/maintenance`.
- Add or adjust tests so doctor fix no longer implies a standalone Maintenance
  layer.
- Flag explicit durable Core imports pointing at generated or legacy sources,
  such as `.gc/system/packs/core` or old `internal/bootstrap/packs/core`, as
  redundant and fixable. The fix may remove those imports after confirming the
  required Core system pack is present.
- If a city imports a custom `core` source, report it as unsupported/manual.
  Core is not user-replaceable in production; users can override Core-provided
  pack configuration through normal config layering, but not replace the
  required Core system pack itself.

Doctor fix safety contract:

- Preflight must run before any mutation. It verifies the public Gastown source
  and immutable version are reachable, installable, and lockable; the generated
  Core system pack can be materialized; existing lockfiles are parseable; and
  the city manifests can be edited without dropping unknown syntax.
- Fixes must be failure-atomic across `city.toml`, rig `pack.toml` files,
  lockfiles, and installed pack directories. Use temp-file-plus-rename for each
  file and leave the city byte-identical if any preflight step fails.
- Healthy cities must be byte-identical after `gc doctor --fix`; add golden
  tests for comments, ordering, unknown tables, unknown fields, array formatting,
  and unchanged lockfile contents.
- TOML edits must be scoped. If the existing parser/editor cannot preserve
  unrelated content, doctor must refuse the automatic fix with manual guidance
  instead of whole-file re-encoding.
- Legacy local Gastown imports are auto-rewritten only when provenance matches
  known generated/example paths. Operator forks, edited local packs, or custom
  public sources are diagnostic/manual unless the user explicitly opts into a
  separate migration command.
- Maintenance imports are auto-removed only for generated or known example
  Maintenance sources. Custom Maintenance-like imports are diagnostic/manual.
- After a fix, doctor must re-run Core materialization/provenance validation and
  public Gastown lock/install validation. A fix that leaves Core absent or the
  public pack unresolved is an error.
- Runtime state under `.gc/runtime/packs/maintenance` follows the concrete
  runtime-state migration table above. JSONL state/archive and spawn-storm
  ledgers move to `.gc/runtime/packs/core` when safe; conflicts are manual;
  ignored legacy paths are never deleted by `gc doctor --fix`.

### Bootstrap Cleanup
<!-- REVIEW: added per blocker-bootstrap-fixture -->

`internal/bootstrap/bootstrap.go` should no longer depend on
`internal/bootstrap/packs/core`.

Use a single approach:

- Make production `bootstrapAssets` a non-nil empty filesystem after the
  production `packs/**` embed is removed; it returns `fs.ErrNotExist` for all
  paths and is covered by `TestProductionBootstrapAssetsIsEmpty`.
- Make bootstrap tests inject an inline `fstest.MapFS` fixture explicitly; they
  must not read production Core assets from `internal/packs/core` and must not
  copy production Core into `testdata`.
- The fixture `Entry.Name` should remain `core` only if the test contract needs
  to prove legacy bootstrap collision metadata accepts a Core-shaped binding.
  `Entry.AssetDir` points at the synthetic fixture path and must not be
  `packs/core` unless the inline fixture intentionally defines that path.
- The production `BootstrapPacks` default remains empty.
- Remove the production `//go:embed packs/**` dependency from
  `internal/bootstrap/bootstrap.go`.
- Remove `core` and `registry` from `bootstrapManagedImportNames` in
  `internal/config/compose.go` in the same slice that permanently empties
  `BootstrapPacks`; update the sync tests so bootstrap-managed implicit imports
  are empty.
- Add guards that fail on lingering imports of
  `internal/bootstrap/packs/core`, `AssetDir: "packs/core"` outside the new
  fixture tests, and file copies/hashes that still target the old production
  Core path.
- Add `TestBootstrapFixtureIsMinimal`, failing if inline fixtures include
  production-only directories such as `formulas/`, `orders/`, `overlay/`,
  `skills/`, or `assets/prompts/`.
- Keep `GC_BOOTSTRAP=skip` only as a legacy bootstrap-materialization skip. It
  must not skip `MaterializeBuiltinPacks`, required Core materialization,
  required Core integrity validation, or config provenance checks.
- Add hook installation tests proving `internal/hooks` reads overlays from
  `internal/packs/core` after bootstrap no longer owns Core assets.

### Examples And Docs
<!-- REVIEW: added per blocker-docs-dx-consistency -->

Update `examples/gastown`:

- Keep it as an example city only if it can use public Gastown imports without
  vendoring `packs/gastown`.
- Remove `examples/gastown/packs/maintenance`.
- Remove `examples/gastown/packs/gastown`.
- Update comments to describe Core as the required system pack and Gastown as
  the explicit external pack.
- Update tests that currently read `examples/gastown/packs/gastown` to either
  move to `gascity-packs` or assert the public-import wiring.

Build the docs/operator inventory from the current tree before editing docs.
Recommended inventory command:
`rg -n "maintenance|system/packs|runtime/packs|gastown|PublicGastown|dog|Core" docs examples cmd internal -g '*.md' -g '*.toml' -g '*.go' -g '*.sh'`.
The resulting inventory must classify each hit as operator-facing, generated
reference, tutorial, command/doctor output, schema/help text, script comment,
historical archive, or test fixture.

Canonical operator wording:

- Core is the required host system pack.
- `bd` and `dolt` are provider-dependent host system packs.
- Maintenance is retired as a standalone pack.
- Gastown is an explicit public pack import from
  `github.com/gastownhall/gascity-packs/gastown`.
- Core maintenance-agent behavior is Core pack configuration; store maintenance
  settings such as `[maintenance.dolt]` are not the retired Maintenance pack.
- Stale `.gc/system/packs/maintenance`, `.gc/system/packs/gastown`, and
  `.gc/runtime/packs/maintenance` paths may remain as ignored legacy state and
  are not deleted by `gc doctor --fix`.

Update docs and generated references:

- `docs/reference/system-packs.md`: Core is the required system pack; `bd` and
  `dolt` remain provider-dependent system packs; Maintenance and Gastown are no
  longer system packs.
- `docs/guides/shareable-packs.md`: remove "core and maintenance stay
  implicit" guidance.
- `docs/getting-started/troubleshooting.md`: replace
  `.gc/runtime/packs/maintenance` and `.gc/system/packs/maintenance` references
  with Core or Gastown paths as appropriate.
- `docs/tutorials/01-cities-and-rigs.md`: describe the `dog` pool as Core's
  configurable maintenance agent, not the Maintenance pack.
- Keep `[maintenance.dolt]` store-maintenance docs intact unless they refer to
  the retired Maintenance pack.
- Update CLI help, doctor strings, examples, generated docs indexes, and script
  comments that are operator-facing. Historical docs may keep old wording only
  through an explicit allowlist checked by a docs lint test.
- Add golden tests for doctor output and first-run/tutorial text for both a
  minimal city and `gc init --template gastown`.

### Cross-Pack Ownership Decisions
<!-- REVIEW: added per major-cross-pack-ownership -->

These decisions must be resolved before deleting source trees or moving the
public Gastown pin:

| Asset or surface | Decision rule | Required evidence |
| --- | --- | --- |
| `mol-review-quorum` | Core only if role-neutral; otherwise Gastown. | Token scan plus formula composition test in final owner. |
| Gastown Codex overlay | Core only if it contains generic provider hook behavior; otherwise Gastown. | Overlay diff and hook installation test from final owner. |
| Dog prompt fragments | Core only for generic maintenance-agent behavior; Gastown notification/requester behavior moves to Gastown. | Generated manifest rows, prompt-template resolution tests, and renamed-agent Core test. |
| `mol-polecat-*` parent/child formulas | Gastown. | Gastown formula composition and migrated old-test mapping. |
| Review workflow checks | Core only for generic review quorum/check infrastructure; Gastown for role/prompt review flows. | Ownership table row plus packcompat test. |
| Branch pruning | Gastown. | Gastown order/script existence and execution test. |
| `mol-shutdown-dance` detector/requester examples | Generic due process in Core; Gastown detector examples in Gastown. | Core role-token scanner and Gastown preservation test. |
| Hardcoded role-theme/tmux APIs | Move role theming to config or Gastown assets; Go APIs cannot encode Gastown roles. | Source scanner and config-driven theme test. |

## Testing
<!-- REVIEW: added per major-behavior-tests -->

### Gas City Unit Tests

Add or update tests for:

- `internal/builtinpacks.All`: only Core, `bd`, and `dolt`; Core subpath is
  `internal/packs/core`.
- Source recognition for `internal/packs/core`; old
  `internal/bootstrap/packs/core` sources should be rejected or covered by
  explicit migration diagnostics.
- Synthetic cache materialization and validation for the remaining bundled
  packs.
- `MaterializeBuiltinPacks`: no Maintenance or Gastown materialization expected;
  Core includes moved maintenance orders and scripts.
- `builtinPackIncludes`: default includes Core and `bd`; non-bd includes Core
  only; exec `gc-beads-bd` includes Core, `bd`, and `dolt`.
- Required pack refresh: Core is repaired when missing or corrupted.
- Core presence doctor: missing/corrupt Core is reported and fixed.
- Import-state doctor: legacy Gastown imports rewrite to public
  `gascity-packs/gastown`; legacy Maintenance imports are removed with the new
  retired-pack wording.
- Hook installation: `internal/hooks` reads overlays from `internal/packs/core`.
- Core role-name guard: Core assets must not contain Mayor, Deacon, Polecat,
  Refinery, Witness, Boot, Crew, or Gastown outside allowed docs/tests.
  `dog` is allowed only under Core maintenance-agent assets.
- Production config-load guard: non-test `cmd/gc` files may not call
  `config.Load*` directly outside the documented partial-read allowlist.
- Post-load Core assertion: normal command helpers fail closed if Core is
  materialized but absent from resolved config provenance.
- Behavior proof for Core formulas and orders: compose representative formulas
  into molecules, assert expected step counts, resolve hooks, load configured
  agents and pools, and execute moved scripts from pack-relative paths.
- Doctor fix safety: idempotent healthy-city no-op, scoped TOML preservation,
  fork/custom-source manual diagnostics, failure-atomic preflight failure, final
  Core provenance revalidation, and runtime-state diagnostic classes.
- Provider pack continuity: `bd` and `dolt` materialized bytes and provenance
  remain unchanged except for expected manifest metadata, and provider matrices
  still resolve the correct required packs.
- Core integrity: tampered Core is repaired; unexpected files either fail full
  file-set integrity validation or are proven unable to influence loaded
  formulas, orders, scripts, overlays, and prompts.
- Bootstrap fixture isolation: test fixture assets do not hash/copy production
  Core, production bootstrap embeds no `packs/**` Core tree, and hook overlays
  load from `internal/packs/core`.

### Gastown Pack Tests

In `gascity-packs`:

- Move Gastown pack behavior tests from `examples/gastown/gastown_test.go` or
  recreate equivalent tests against `gastown/`.
- Add behavior-preservation tests for all generated manifest entries.
- Add a test-by-test migration matrix for
  `examples/gastown/gastown_test.go`,
  `examples/gastown/maintenance_scripts_test.go`, packlint/parser coverage,
  and any Maintenance auto-inclusion tests removed from Gas City.
- Verify branch pruning order/script exists in Gastown.
- Verify Polecat formulas moved from Core and still compose with
  `mol-polecat-work`.
- Verify Gastown detector/requester behavior removed from Core still exists in
  Gastown prompts or formulas.
- Verify Gastown can run with host Core present and without a Maintenance pack.

### Integration And Regression Coverage

Run focused suites first:

- `go test ./internal/builtinpacks ./internal/hooks ./internal/bootstrap`
- `go test ./cmd/gc -run 'BuiltinPack|ImportStateDoctor|CorePack|TryReloadConfig|MaterializeBuiltinPacks'`
- `go test ./test/packlint`
- `go test ./test/packcompat -run TestPinnedPublicGastownBehavior`
- `go test ./examples/...` until the example-tree source removal slice lands;
  after that slice, replace it with the focused public-import wiring tests.

Then run the repo gate documented in `TESTING.md`:

- `make test-fast-parallel`
- `go vet ./...`

Run `make dashboard-check` only if API, dashboard, or generated schema files
are touched. This migration should not require dashboard changes.

## Rollout
<!-- REVIEW: added per blocker-rollout-and-test-slicing -->

Use a staged rollout to keep Gas City and `gascity-packs` in sync.

1. Candidate public Gastown slice:
   land `gascity-packs` Gastown behavior preservation on a branch, move/add
   Gastown-owned assets, add the generated behavior manifest, resolve all
   cross-pack ownership audits, update Gastown tests, and record the immutable
   commit/hash. This slice must decide `mol-review-quorum`, provider overlays,
   Dog prompt fragments, Polecat formulas, branch pruning, shutdown-dance
   examples, review checks, and hardcoded role-theme/tmux APIs before the pin is
   consumed. Required gates: `gascity-packs` test suite, manifest completeness,
   formula/order composition, prompt-template resolution, script execution,
   clean scan for retired Maintenance paths, and old-test to new-test mapping.
2. Gas City public-pin adoption and packcompat slice:
   update `internal/config/PublicGastownPackVersion` to the immutable public
   `gascity-packs/gastown` commit from slice 1 and add the Gas City
   pinned-public-pack compatibility test in current-loader mode without
   deleting in-tree sources or claiming no-Maintenance production-loader proof.
   Required gates:
   ordinary remote-pack install for the exact pin,
   `go test ./test/packcompat -run TestPinnedPublicGastownBehavior`,
   old-binary/new-pack compatibility proof,
   stale synthetic-cache rejection for retired aliases,
   `go test ./examples/...`, `make test-fast-parallel`, and `go vet ./...`.
3. Core extraction slice:
   move Core assets to `internal/packs/core`, update hooks and builtin registry
   imports, and add bootstrap fixture isolation. Required gates:
   `go test ./internal/builtinpacks ./internal/hooks ./internal/bootstrap`,
   `go test ./cmd/gc`, `go test ./internal/config`,
   `go test ./test/packlint/...`, source guards for old bootstrap Core
   dependencies, `make test-fast-parallel`, and `go vet ./...`.
4. Core loading/doctor slice:
   add the post-load Core provenance assertion, production `config.Load*`
   scanner, call-site migration/allowlist, Core doctor, pre-resolution import
   recovery, version-skew diagnostics, and safe import-state doctor edits.
   Required gates:
   `go test ./cmd/gc -run 'BuiltinPack|ImportStateDoctor|CorePack|TryReloadConfig|MaterializeBuiltinPacks'`
   plus doctor golden/failure-atomic tests, `make test-fast-parallel`, and
   `go vet ./...`.
5. Maintenance folding slice:
   move Core-owned Maintenance assets into Core, move Gastown-owned Maintenance
   assets to public Gastown, remove Maintenance from `requiredBuiltinPackNames`,
   update `jsonl_archive_doctor_check.go` to prefer Core runtime state, and
   apply the runtime retirement table. Required gates: formula/order/script
   behavior tests, stale directory preservation tests, runtime-state migration
   and conflict diagnostics, provider pack continuity tests,
   `go test ./test/packcompat -run TestPinnedPublicGastownBehavior` in
   no-Maintenance production-loader mode, `make test-fast-parallel`, and
   `go vet ./...`.
6. Registry/cache cleanup slice:
   remove Maintenance and Gastown from the embedded pack registry, retire public
   synthetic aliases, and verify stale cache rejection. This slice must consume
   the already-updated `PublicGastownPackVersion` from slice 2; it must not be
   the first pin update. Required gates: registry/cache negative tests and
   packcompat test against the exact public pin, `make test-fast-parallel`, and
   `go vet ./...`.
7. Source deletion/docs slice:
   remove in-tree `examples/gastown/packs/*` sources only after replacement
   tests are green, migrate docs and examples, and apply docs lint/golden
   checks. Required gates: docs inventory/lint, tutorial/doctor output golden
   tests, focused public-import wiring tests, `make test-fast-parallel`, and
   `go vet ./...`.

Each slice must be independently deployable or explicitly marked as a paired
cross-repo change with rollback instructions. Do not batch source deletion,
pin changes, doctor mutation logic, and docs into one unverified commit.

Release compatibility matrix:

| Gas City binary | Public Gastown pack | Expected behavior |
| --- | --- | --- |
| old binary | old pack | Existing behavior unchanged. |
| old binary | new pack | Public pack remains compatible with host Core and absence of Maintenance imports; no reliance on new Gas City-only loader behavior. |
| new binary | old pack | Fresh Gastown init pins the new pack, but existing locks to old public commits continue to load only far enough for pre-resolution doctor diagnostics; doctor reports an actionable version-skew/import diagnostic without mutating custom content. |
| new binary | new pack | Core, provider packs, and public Gastown load without Maintenance; packcompat and generated-manifest gates pass. |
| rollback from new to old | any existing lock | Doctor-mutated manifests either remain readable by old binaries or the release notes name an explicit downgrade limitation and manual recovery path. |

Backward compatibility:

- Existing cities with legacy local Gastown imports get rewritten to the public
  Gastown pack by `gc doctor --fix`.
- Existing cities with legacy Maintenance imports get those imports removed by
  `gc doctor --fix`; Core supplies generic maintenance and public Gastown
  supplies Gastown-specific behavior.
- Existing stale `.gc/system/packs/maintenance` or `.gc/system/packs/gastown`
  directories are ignored by config loading after the migration. Doctor may
  report them as legacy ignored directories, but `gc doctor --fix` must not
  delete them because they may contain operator edits.
- Existing order skip lists containing moved Core order names should continue
  to work when names are preserved. If any order is renamed, provide aliases or
  a migration test.

Operational risk:

- The highest risk is losing Gastown behavior while generalizing Core. The
  generated behavior manifest and Gastown pack tests are the mitigation.
- The second risk is stale assumptions in `cmd/gc` direct config loads. The
  production call-site audit and Core presence doctor are the mitigation.
- The third risk is public pack version skew. Land `gascity-packs` first and
  update `PublicGastownPackVersion` only after the public commit is available.

## Open Questions

No shared requirements or policy questions remain.

Implementation audits are represented as blocking ownership decisions in
`Cross-Pack Ownership Decisions`. They must be resolved in the slice that moves
or deletes the relevant source, and the resolution must include the required
evidence named in that table.
