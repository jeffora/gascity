---
plan_slug: core-gastown-pack-migration
phase: design
rig: gascity
rig_root: /data/projects/gascity-fresh-main-20260604-VLKm8c
artifact_root: /data/projects/gascity-fresh-main-20260604-VLKm8c/plans
requirements_file: /data/projects/gascity-fresh-main-20260604-VLKm8c/plans/core-gastown-pack-migration/requirements.md
status: draft
created_at: 2026-06-04T15:07:35Z
updated_at: 2026-06-07T14:05:04Z
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

`GC_BOOTSTRAP=skip` is retired as a production behavior switch. If retained for
tests, it may skip only legacy bootstrap fixture materialization, which is empty
in production. It must not skip `internal/systempacks` materialization,
required Core file-set integrity validation, retired-source classification,
collision checks, or typed participation checks. Tests must prove a command with
`GC_BOOTSTRAP=skip` still materializes and verifies Core through the normal
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

### Attempt 7 Review Resolution Contracts
<!-- REVIEW: added per attempt7-global-blockers -->

Attempt 7 keeps the verdict at `iterate` until the following contracts are
implemented or intentionally narrowed in this document. These contracts
supersede weaker path-only, name-only, or "decide during implementation"
language above.

#### Typed Required Core Participation
<!-- REVIEW: added per attempt7-required-core-participation -->

Required Core proof is a typed contract, not a materialized path check.
Production config loaders produce `RequiredSystemPackParticipation` records for
Core and provider-required system packs after materialization and after normal
config resolution. Each record contains:

- pack id and required-pack kind
- embedded source id, source commit or binary build id, and content-manifest
  digest
- full expected file-set digest, `pack.toml` digest, and validation result
- materialized directory and freshness or repair status
- resolved-config layer id and import edge proving that the validated host
  system pack participated in the final resolved config
- validation timestamp, diagnostic id, and fatal/nonfatal classification

Normal production paths have two fatal gates:

1. Pre-resolution integrity: required Core must be materialized from the
   expected embedded source, match the strict manifest/file-set policy, pass
   collision checks, and have no effective unexpected files. Unexpected files
   are fatal unless a central classifier proves the files cannot affect loaded
   formulas, orders, scripts, overlays, prompts, or config.
2. Post-resolution participation: the final resolved config must contain the
   typed participation record for the same validated Core layer. A successful
   materialization without this record is a load failure.

The bypass scanner covers non-test `cmd/gc` files and production `internal/`
loader surfaces that can drive runtime behavior, including controller reload,
API/session state loading, routing/state helpers, import/install/check paths,
doctor diagnostics after pre-resolution, init readiness, convoy/sling/wait
flows, and Dolt publication. Direct `config.Load`, `config.LoadCity`, or
`config.LoadWithIncludes` calls outside the named wrappers must either move to
the wrapper or appear in a partial-read allowlist with file, function, reason,
and a focused test proving it cannot drive normal runtime behavior.

`GC_BOOTSTRAP=skip` is not allowed to skip retired-directory pruning checks,
source-collision checks, Core materialization, strict manifest validation,
provider-required pack validation, or post-load participation assertions. If
that narrower meaning cannot be implemented without ambiguity, the environment
variable is retired and replaced by test-only fixture injection.

Required failing-before/passing-after tests:

- missing, corrupt, stale, partial, and extra-file Core materializations
- materialized Core absent from resolved participation
- user or imported pack named `core`
- controller no-refresh reload and API/session state paths
- `GC_BOOTSTRAP=skip` on a normal command
- production `internal/` direct-load bypasses

#### Go And Asset Role-Surface Migration Table
<!-- REVIEW: added per attempt7-role-surface-migration -->

Go de-roling is in scope for this migration because the design uses role
cleanup as a source-deletion gate. The implementation must generate and check
in `plans/core-gastown-pack-migration/role-surface.generated.yaml` before
deleting or moving source assets. Each row contains:

- file, function/table/key, and source span
- current role-bearing behavior
- final owner: Core, public Gastown, provider pack, docs-only, or approved
  fixture
- replacement mechanism: config value, formula variable, metadata field,
  prompt/template text, pack-owned asset, or deletion
- rollout slice and blocking test
- allowlist class and expiry, if retained

The inventory must cover tmux themes/icons, default city scaffolding, prompt
fallbacks, warmup mail defaults, sling formula-name heuristics, API `crew`
classification, provider `dog` bindings, Core notification targets, TOML
defaults, comments, generated help text, prompt prose, and moved order/formula
assets.

`dog` is the configurable Core maintenance-worker target. It is not an SDK
primitive and not a Gastown convention. Core orders and provider packs that need
an agent resolve the maintenance-worker name from pack or city configuration.
Public Gastown may patch the host Core `dog` configuration through normal
config layering, but Go must continue to work when the Core maintenance worker
is renamed or omitted. Tests must prove Core-only loading, controller-owned SDK
operations, provider-pack resolution, public Gastown host-Core patching, and
renamed/omitted maintenance-worker behavior.

Source deletion cannot proceed while any behavior-bearing role row lacks a
replacement mechanism and test. Historical docs and migration fixtures are
allowed only through explicit row classifications.

#### Public Gastown Remote Pin And Duplicate-Definition Gate
<!-- REVIEW: added per attempt7-public-gastown-pin -->

The public pin slice must retire or bypass the public synthetic Gastown alias
before `PublicGastownPackVersion` is consumed. `PublicGastownPackSource` must
resolve through the ordinary remote repository path, validate the exact
immutable commit, and verify the `gastown` subpath identity. It must never
select embedded bytes, historical synthetic cache entries, or a bundled
`gastown` alias for a public commit pin.

Pin adoption tests must fail before the resolver change and pass after it:

- ordinary remote-cache install for the exact `PublicGastownPackVersion`
- exact commit checkout and `gastown/pack.toml` identity
- stale synthetic Gastown cache present but ignored
- no embedded public Gastown bytes selected for a public pin
- generated behavior manifest loaded from the pinned public commit and digest
  matched against the Gas City lock/test fixture

Duplicate active definitions are a fatal intermediate-state risk. Normal
production loading must reject or diagnose any city where public Gastown and
stale local Maintenance/Gastown sources both contribute active formulas,
orders, prompts, overlays, scripts, patches, or agents. Preserved
`.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories are
ignored unless explicitly imported; explicit generated legacy imports are
doctor-fixable, and custom forks are manual diagnostics. Tests must include
stale preserved directories, generated legacy imports, custom local forks,
duplicate order names, duplicate prompt fragments, duplicate scripts, and
duplicate patch targets.

#### Preservation-Proven Doctor Transaction
<!-- REVIEW: added per attempt7-doctor-transaction -->

`gc doctor --fix` uses a scoped byte-preserving TOML edit planner rather than
whole-file TOML re-encoding. The planner identifies editable spans for complete
legacy `[[imports.*]]` tables and scalar version/source lines, validates the
result with the normal TOML parser, and leaves all other bytes untouched. If a
manifest shape cannot be edited while preserving comments, unknown fields,
unknown tables, array order, whitespace, and unrelated lock entries, automatic
fix refuses with manual steps.

The doctor migration coordinator is one composite plan:

1. Read all target manifests, lockfiles, installed-pack metadata, and relevant
   runtime-state paths.
2. Prove generated-source provenance for any local import removal or rewrite.
3. Prove public Gastown reachability, exact commit installability, ordinary
   remote-cache identity, and lock generation.
4. Stage Core materialization in a temp directory and validate strict manifest
   integrity before touching user files.
5. Stage runtime-state moves with reversible names and conflict checks; never
   delete legacy Maintenance state.
6. Write manifests and locks last with compare-before-rename on every target.
7. Re-run normal production loading, typed Core participation, duplicate-source
   checks, and public Gastown validation after the writes.

If any preflight, staged operation, compare-before-rename, runtime-state move,
or final validation fails, files that were not yet committed remain
byte-identical, and the error names the exact failed operation. Failure
injection tests must cover mutation between preflight and rename, offline
no-cache, failed public fetch, stale lock, Core materialization failure,
runtime-state conflicts, and final validation failure.

#### Retired-Source Classifier And Runtime Containment
<!-- REVIEW: added per attempt7-retired-source-containment -->

Gas City gets one central retired-source classifier used by config load,
pre-resolution doctor, import check/install, cached reads, lock validation,
builtin materialization, prompt/template discovery, docs lint, and error
rendering. The classifier returns typed states:

- active required host system pack
- active public pack
- generated retired Maintenance source
- generated retired local Gastown source
- preserved stale system-pack directory
- custom local fork
- historical/generated docs fixture
- legacy runtime state
- invalid duplicate active definition

Prompt and template discovery must walk only active resolved packs and required
host packs. It must not glob preserved stale system-pack directories. Tests must
cover absent retired directories, stale generated directories with custom files,
explicit legacy imports, custom forks, stale prompt fragments, stale hook
overlays, and stale order/formula definitions.

Docs and examples are part of retired-source containment. The docs inventory
must include Markdown, MDX, TOML, shell, Go strings, generated references, CLI
help, doctor messages, examples, pack comments, and navigation. Every retained
Maintenance/Gastown path hit is classified as generated, historical, migration
fixture, legacy diagnostic, valid store-maintenance terminology, or active
public Gastown reference.

#### Behavior Manifest And Packcompat Evidence Contract
<!-- REVIEW: added per attempt7-behavior-packcompat-evidence -->

The behavior manifest is a proof input, not a checklist. Each row must include:

- manifest schema version and generator version
- old source digest, final asset digest, helper dependency digests, and
  behavior-bearing prompt/script/template fragment digests
- evidence class: trigger, observable output, formula composition, molecule
  step, prompt render, script side effect, runtime-state mutation,
  notification/requester/detector semantics, provider database filter, TOML
  default, or docs-only wording
- old witness and final witness at test-function and subtest granularity
- cross-repo owner and immutable public commit when behavior moves to public
  Gastown
- semantic-delta status: equivalent, intentionally generalized, intentionally
  retired, or docs-only, with owner and operator impact for non-equivalent rows
- generated row id stable across path renames when the behavior identity is
  unchanged

The generator also emits a test-migration table for every touched or removed
Gas City test function and subtest. A test may disappear only when the table
names the replacement Gas City test, replacement public Gastown test, or
approved removal row.

`test/packcompat` must cover this fixture matrix before source deletion:

| Fixture | Required proof |
| --- | --- |
| fresh init | public Gastown installs from ordinary remote cache or network at the exact pin. |
| upgraded locks | old local imports get pre-resolution diagnostics and safe fixes. |
| stale synthetic cache | historical public Gastown bundled cache is ignored. |
| ordinary remote cache | exact public pin resolves without network when cache is already populated. |
| offline no-cache | failure is explicit; no embedded fallback is selected. |
| old binary + new pack | public pack remains compatible with host Core and no Maintenance import. |
| new binary + old pack | doctor reports version skew without mutating custom content. |
| no-Maintenance production loader | host Core plus public Gastown resolves through normal production loading. |
| host-Core patch resolution | public Gastown patches Core maintenance-worker configuration only through normal layers. |

Gas City must prove it consumes the manifest from the exact pinned public
Gastown commit, not a copied local file, by checking the manifest digest through
the same remote-cache or checkout path used for pack installation.

#### Registry, Cache, And Materializer Slice Gates
<!-- REVIEW: added per attempt7-registry-cache-materializer -->

Registry cleanup decisions are slice-scoped and explicit:

| Surface | Delete, retain, or change | Slice gate |
| --- | --- | --- |
| `PublicRepository` and public source normalization | Retain for real public repo sources only; remove bundled/synthetic public Gastown alias behavior. | Public-pin slice before consuming `PublicGastownPackVersion`. |
| `publicSubpathForPack` | Retain only for ordinary public repo subpaths such as `gastown`; no synthetic Maintenance. | Remote exact-commit/subpath tests. |
| `RepoCacheKey` | Change public Gastown to source+version namespace; old synthetic keys are stale and ignored. | Stale-cache rejection and ordinary-cache hit tests. |
| `SyntheticContentHash` | Retain only for bundled Core, `bd`, and `dolt`. | Bundled-pack validation tests. |
| `requiredBuiltinPackNames` | Remove Maintenance only after Core-owned Maintenance behavior has moved and packcompat no-Maintenance mode is green. | Normal production loader test, not test-only includes. |
| `internal/systempacks.MaterializeRequiredPacks` | Materialize/repair Core, `bd`, and `dolt`; never refresh or prune retired Maintenance/Gastown directories. | Missing/corrupt provider repair and stale-directory preservation tests. |
| retired source recognizers | Convert to diagnostics, not active sources. | Import/check/install negative tests. |

Offline old-cache-to-new-binary tests must prove `bd` and `dolt` still
self-heal without network and are byte-identical aside from intentional
generated markers. No slice may claim "no Maintenance" until the same
production path operators run excludes Maintenance from required includes.

#### Bootstrap Escape-Hatch Closure
<!-- REVIEW: added per attempt7-bootstrap-escape-hatch -->

Production bootstrap has a private non-nil empty `fs.FS`; it never embeds,
copies, hashes, or fixture-loads production Core. Test fixtures use explicit
inline `fstest.MapFS` allowlists.

The Core extraction slice adds:

- `TestProductionBootstrapAssetsIsEmpty`
- `TestBootstrapFixtureIsMinimal`
- `TestBootstrapManagedImportNamesInSync`, expecting no production bootstrap
  managed imports after `BootstrapPacks` is empty
- a path-string guard for `internal/bootstrap/packs/core`, `AssetDir:
  "packs/core"`, old hook overlay paths, active engineering docs, generated
  docs, and test helpers outside the fixture allowlist
- a normal-command test proving `GC_BOOTSTRAP=skip` still materializes and
  validates Core through `internal/systempacks`

If `GC_BOOTSTRAP=skip` cannot be limited to empty bootstrap fixture behavior,
the variable is removed from production command semantics in the same slice.

#### Docs Vocabulary As Executable Contract
<!-- REVIEW: added per attempt7-docs-vocabulary-contract -->

The source-deletion/docs slice creates and nav-registers
`docs/reference/system-packs.md`. If an existing page is designated instead,
the design must name that path before implementation starts and the docs-nav
test must require it.

The wording matrix is executable. It distinguishes:

- retired standalone Maintenance pack
- lower-case maintenance as ordinary English
- store maintenance and `gc.store.maintenance.*` events
- Core maintenance-worker configuration
- stale legacy `.gc/system/packs/maintenance` and
  `.gc/runtime/packs/maintenance` paths
- active public Gastown imports

Docs lint and golden tests cover doctor `FixHint`, doctor JSON/text output, CLI
help, config/schema generated docs, tutorials, troubleshooting, examples, pack
comments, script comments, generated references, and docs navigation. Exact
phrases are preferred for operator-facing warnings; semantically equivalent
wording is allowed only when a golden test names the accepted phrase.

### Attempt 8 Review Resolution Contracts
<!-- REVIEW: added per attempt8-global-blockers -->

Attempt 8 keeps the review verdict at `iterate` until these contracts are
implemented or explicitly narrowed by a follow-up design change. These contracts
supersede any older wording that allows path-only proof, role-token-only proof,
synthetic public Gastown resolution, whole-file TOML rewrites, stale-pack
runtime reads, or release-time documentation cleanup.

#### Required-System-Pack Participation Record
<!-- REVIEW: added per attempt8-required-core-participation -->

Required system-pack proof is represented by one typed record:
`RequiredSystemPackParticipation`. The record is produced for Core and every
provider-required host pack by the normal production loader and contains:

- required pack id and host-pack kind
- embedded source id, source digest, binary build id, and expected file-set
  digest
- materialized path, repair/freshness status, and strict integrity result
- resolved layer id, import edge id, and post-resolution participation status
- collision result for same-name user/imported packs
- bypass class: runtime, no-refresh runtime, partial edit, partial doctor, or
  test-only fixture
- diagnostic id and fatality

Two fatal gates run for every runtime load:

1. Strict required-pack integrity before participation. Missing, stale,
   corrupt, shadowed, partial, extra-effective-file, wrong-source, wrong-digest,
   or same-name-collision states fail before behavior can be read.
2. Post-load participation. The final resolved config must reference the same
   validated host Core layer through a typed participation record. Materialized
   paths, provenance strings, include counts, or successful file reads are not
   sufficient proof.

The bypass guard covers `cmd/gc` and production `internal/` entrypoints that can
drive behavior: controller reload, API state/config helpers, dispatch routing,
session resolution, import/check/install paths, doctor after pre-resolution,
init readiness, convoy/sling/wait flows, Dolt publication, no-refresh loaders,
and package helpers used by those surfaces. A direct `config.Load*` call in
those paths is a bug unless it is in the partial-read allowlist with file,
function, reason, fatality boundary, and a focused test proving it cannot drive
runtime behavior.

Core name collisions are fatal in production. A user pack, imported pack,
synthetic alias, lock entry, or materialized retired directory named `core` may
only be diagnosed or edited before resolution; it cannot replace, shadow, patch
around, or satisfy required host Core. `GC_BOOTSTRAP=skip` cannot suppress any
required-pack materialization, pruning/classifier check, integrity proof,
collision proof, participation proof, or doctor cleanup. If that cannot be
enforced unambiguously, the variable is removed from production command
semantics and replaced with test fixture injection.

#### Role-Surface Migration Inventory
<!-- REVIEW: added per attempt8-role-surface-inventory -->

Role neutrality is proved by a generated inventory, not by a token scanner
alone. The implementation creates
`plans/core-gastown-pack-migration/role-surface.generated.yaml` and a rendered
table in the implementation PR. Each row names file or asset, current behavior,
final owner, replacement mechanism, rollout slice, proof test, and allowlist
expiry if retained.

The initial table must include at least:

| Surface | Current risk | Final owner and replacement |
| --- | --- | --- |
| tmux theme/icon helpers such as `DogTheme` | Go API encodes a role-shaped default. | Core only as configurable maintenance-worker display metadata; otherwise config or Gastown asset. |
| default city scaffolding and warmup mail | Generated defaults can imply Gastown roles. | Core provides role-neutral infrastructure defaults; Gastown-specific defaults move to public Gastown templates. |
| prompt fallback and generated help text | Fallback text can preserve Mayor/Deacon/Witness/Polecat assumptions. | Core prompt fallbacks become role-neutral; Gastown examples live in public Gastown assets/docs. |
| `internal/sling` formula-name heuristics | Formula names can encode role decisions in Go. | Formula selection comes from config/metadata; Gastown formulas live in public Gastown. |
| API `crew` classification | API can classify a role family as infrastructure. | Replace with config-supplied grouping or neutral session metadata. |
| provider `dog` bindings | Provider packs may assume a hardcoded worker. | Resolve Core maintenance-worker target from pack/city config. |
| Core notification/order targets | Routes can target Gastown roles from Core. | Core uses configured recipients; Gastown-specific recipients move to public Gastown. |
| TOML defaults, scripts, prompt fragments, comments, formula/order metadata | Asset text can keep behavior even when Go is clean. | Final owner is Core, public Gastown, provider pack, docs-only, or approved fixture with a manifest row. |

`dog` is the Core maintenance-worker compatibility target. It is configurable
Core pack data, not an SDK primitive and not a Gastown convention. Tests must
prove Core-only cities load and controller-owned SDK operations work when that
worker is renamed or omitted, while Core maintenance formulas that need a worker
resolve the configured name. Source deletion is blocked until every
behavior-bearing role row has a replacement mechanism and proof test.

#### Public Gastown Pin And Synthetic-Alias Cutover
<!-- REVIEW: added per attempt8-public-pin-synthetic-alias -->

Public Gastown pin adoption starts by disabling bundled/synthetic Gastown
resolution for public sources. `PublicGastownPackSource` must resolve through
the ordinary remote repository path and validate:

- exact immutable commit equals `PublicGastownPackVersion`
- `gastown/pack.toml` exists at that commit and has the expected pack identity
- the ordinary repo cache key is source plus version, not a historical synthetic
  alias
- no embedded bytes, synthetic cache entry, old cache namespace, local fallback,
  or materialized `.gc/system/packs/gastown` directory can satisfy the pin
- the behavior manifest consumed by Gas City is read from the same checkout or
  cache path used for pack installation

This cutover is in the public-pin slice before any source deletion or registry
cleanup. Duplicate active definitions are fatal across every intermediate and
rollback state: public Gastown plus bundled, stale local, generated legacy,
remote cache, synthetic cache, materialized Maintenance/Gastown, old-binary/new
pack, and new-binary/old pack combinations. The duplicate check covers agents,
patches, prompts, prompt fragments, formulas, base formulas, orders, scripts,
hooks, notification targets, and runtime-state owners.

#### Doctor Mutation Coordinator
<!-- REVIEW: added per attempt8-doctor-mutation-coordinator -->

`gc doctor --fix` is implemented through a single mutation coordinator. No
implementation bead may add ad hoc direct calls to
`MaterializeBuiltinPacks(cityPath)`, path-only provenance checks, whole-file
TOML rewrites, or partial live commits outside this coordinator.

The coordinator performs one operation-scoped preflight over manifests,
lockfiles, installed-pack metadata, Core materialization targets, public
Gastown install/cache state, retired-source classifications, and runtime-state
paths. It stages Core and public Gastown content outside the live city, edits
TOML through a CST/span-preserving editor, plans runtime-state moves, validates
the staged overlay, and then publishes through compare-before-rename writes. If
the editor cannot preserve comments, unknown fields, unknown tables, array
order, lockfile lexical precision, and unrelated lock entries, automatic fix
refuses with manual steps.

Failure-injection coverage is required for partial writes, mutation between
preflight and rename, stale targets, controller-active cities discovered from
live runtime state, air-gapped repairs, failed fetches, lockfile precision,
runtime-state conflicts, final validation failure, and repeated healthy no-op
fixes. Healthy cities must remain byte-identical after `gc doctor --fix`.

#### Retired-Source Containment API
<!-- REVIEW: added per attempt8-retired-source-containment-api -->

Retired Maintenance and in-tree Gastown sources are handled by one central
classifier API. Config load, lock validation, import install/check, packman
paths, cached reads, builtin materialization, pre-resolution doctor,
prompt/template discovery, docs lint, generated-reference lint, and error
rendering must use that API instead of independent path tests or filesystem
globs.

The classifier states are active host pack, active public pack, generated
retired Maintenance, generated retired local Gastown, preserved stale generated
directory, stale synthetic cache, custom local fork, historical fixture, legacy
runtime state, duplicate active definition, and invalid collision. Prompt,
template, hook, formula, order, script, and overlay discovery may read only
active resolved packs and required host packs. The no-Maintenance slice stops
materializing retired Maintenance/Gastown directories; it preserves existing
directories and diagnoses them, but never refreshes, prunes, imports, or globs
them as behavior sources.

Docs/examples/scripts/generated references are gated by an inventory-backed
lint. Every retained `maintenance`, `packs/maintenance`,
`.gc/system/packs/gastown`, `.gc/system/packs/maintenance`, or
`.gc/runtime/packs/maintenance` hit is classified as active public Gastown,
legacy diagnostic, generated fixture, historical reference, valid
store-maintenance terminology, or a failing stale reference.

#### Behavior Evidence Matrix
<!-- REVIEW: added per attempt8-behavior-evidence-matrix -->

The behavior manifest schema expands to cover behavior that is not visible from
file moves alone. Rows must include schema and generator versions; generator
owner and command; old-source digest; final asset digest; helper dependency
digests; behavior-bearing prompt/script/template fragment digests; semantic
delta approval; public-pack ownership; pinned public commit; and old/new
test-function plus subtest mappings.

Required evidence classes include notification/requester/detector semantics,
Git/Dolt/process authorship identity, provider database filters, TOML defaults,
prompt fragments, `session_live` hook behavior, dog-field ownership, moved
helper dependencies, formula composition, molecule step construction, prompt
rendering, pack-relative script execution, runtime-state mutation, docs-only
wording, and approved retirement.

`test/packcompat` must prove the exact public Gastown manifest consumed by Gas
City at the pinned commit across this matrix: fresh init, upgraded locks, stale
synthetic cache, ordinary remote cache, offline no-cache, old binary/new pack,
new binary/old pack, no-Maintenance loader, host-Core patching, stale local
Gastown imports, duplicate active definitions, and provider-pack continuity for
`bd` and `dolt`.

#### Slice-Accurate Bootstrap, Registry, And Cache Gates
<!-- REVIEW: added per attempt8-slice-accurate-bootstrap-registry -->

Bootstrap, registry, cache, and materializer cleanup use this disposition table:

| Surface | Required disposition | Earliest slice |
| --- | --- | --- |
| `bootstrapManagedImportNames` | Empty only after required-Core collision enforcement and bootstrap fixture isolation are green. | Core extraction. |
| `bootstrapAssets` | Non-nil empty production `fs.FS`; tests use inline fixtures only. | Core extraction. |
| `internal/bootstrap/packs/core` references | Active code/docs/generated refs removed or allowlisted as fixtures. | Core extraction. |
| `All()` | Returns only Core, `bd`, and `dolt` after replacement behavior is pinned and proven. | Registry/cache cleanup. |
| `requiredBuiltinPackNames` | Removes Maintenance only after Core-owned behavior moved and no-Maintenance production-loader packcompat is green. | Maintenance folding. |
| `publicSubpathForPack` | Retains ordinary public repo subpaths only; no synthetic Maintenance/Gastown alias. | Public-pin slice. |
| `PublicRepository` | Supports real public sources only, not bundled public Gastown. | Public-pin slice. |
| `RepoCacheKey` | Source-plus-version namespace for public Gastown; old synthetic namespaces ignored. | Public-pin slice. |
| `SyntheticContentHash` | Bundled Core, `bd`, and `dolt` only. | Registry/cache cleanup. |
| `MaterializeSyntheticRepo` | Never materializes retired Maintenance/Gastown for a public pin. | Public-pin slice. |
| `internal/systempacks.MaterializeRequiredPacks` | Repairs Core and provider host packs; never refreshes or prunes retired directories. | Maintenance folding. |

Each row needs a failing-before/passing-after test in the same slice, plus
offline old-cache provider-continuity coverage. No slice may claim no-Maintenance
success through test-only loaders, copied fixtures, or hidden
`internal/bootstrap/packs/core` paths.

#### Docs And Operator Vocabulary Release Gate
<!-- REVIEW: added per attempt8-docs-release-gate -->

`docs/reference/system-packs.md` is the canonical system-pack reference and is
nav-registered in the first slice that changes operator-facing behavior. If an
existing page is substituted, this design must name the exact path and the nav
test must enforce it before implementation starts.

The wording matrix is a generated contract shared by docs lint, doctor strings,
CLI help, generated references, schemas, examples, scripts, public Gastown
docs, and troubleshooting. It covers Core, provider host packs, explicit public
Gastown, retired Maintenance, Core maintenance-worker terminology, stale
generated directories, store/Dolt maintenance, version skew, runtime-state
migration, and host-Core patch semantics. Any slice that changes
operator-facing behavior without updating the matrix, golden outputs, and docs
navigation is marked non-release until those gates pass.

#### Design-Review Artifact Guard
<!-- REVIEW: added per attempt8-workflow-artifact-guard -->

The attempt-8 persona synthesis metadata bug is a workflow-system follow-up, not
a Core/Gastown migration design choice, but it must be tracked before another
design-review iteration is trusted. Persona synthesis output paths must include
the current attempt directory, and global synthesis must fail before review if
any required attempt-local persona synthesis is absent or has mismatched
`gc.attempt` metadata.

### Attempt 9 Review Resolution Contracts
<!-- REVIEW: added per attempt9-global-blockers -->

Attempt 9 keeps the review verdict at `iterate` and turns the remaining review
gaps into implementation gates. These contracts supersede any older wording that
allows provenance-only Core proof, bundled/public Gastown overlap, retired pack
fallback, ad hoc doctor writes, role-token-only scanning, weak behavior
witnesses, production bootstrap escape hatches, late docs, or broad slice gates.

#### Required-System-Pack Participation Contract
<!-- REVIEW: added per attempt9-required-core-loading -->

`RequiredSystemPackParticipation` is the only proof that required Core, `bd`,
or `dolt` participated in a runtime load. The record is produced by one
required-system-pack wrapper and contains:

- required pack id, pack kind, binary build id, embedded source id, and content
  manifest digest
- strict expected file-set digest, `pack.toml` digest, and unexpected-file
  classification
- materialized directory, freshness/repair status, and pre-resolution
  integrity diagnostic
- resolved layer id, import edge id, and post-resolution participation
  diagnostic for the same validated layer
- collision result for same-name user packs, imported packs, lock entries,
  synthetic aliases, and retired materialized directories
- loader class: runtime, no-refresh runtime, partial edit, partial doctor, or
  test-only fixture

The wrapper has two fatal gates:

1. Pre-resolution integrity: missing, stale, corrupt, partial, wrong-source,
   wrong-digest, shadowed, same-name-colliding, or effective-extra-file Core
   fails before any behavior-bearing config, formula, order, script, prompt, or
   overlay can be read.
2. Post-resolution participation: the final resolved config must contain a
   typed participation record for the same validated host-system-pack layer.
   Successful materialization, file reads, path matches, include counts, or
   provenance strings are not enough.

The bypass scanner is whole-production, not `cmd/gc`-only. It scans non-test Go
under `cmd/gc` and production `internal/` packages that can drive config
resolution or runtime behavior, including API/config helpers, controller reload,
dispatch routing, session resolution, doctor/configedit after pre-resolution,
import/check/install, packman, cache/lock validation, no-refresh loaders,
convoy/sling/wait flows, Dolt publication, and helper packages called by those
surfaces. Any direct `config.Load*` call in scope must either move to the
wrapper or appear in a generated partial-read allowlist naming file, function,
loader class, reason, fatality boundary, owner, expiry, and focused test.

User, imported, synthetic, lock, or materialized retired entries named `core`
are fatal collisions for runtime loads. They may be diagnosed or rewritten only
by partial doctor/edit paths before full resolution; they cannot satisfy,
replace, shadow, or patch around required host Core.

#### Public Gastown Pin And Overlap Boundary
<!-- REVIEW: added per attempt9-public-gastown-pin-overlap -->

The chosen rollout model is a backward-compatible first public pin followed by
an activation pin. Gas City may not consume a public Gastown pin that activates
definitions already active from bundled Maintenance or in-tree Gastown under
the current loader.

The first public pin is the compatibility pin:

- it resolves only through
  `https://github.com/gastownhall/gascity-packs.git//gastown` at the immutable
  commit named by `PublicGastownPackVersion`
- registry-selector install, direct source resolution, lock generation, and
  packcompat must prove the same reachable commit, subpath identity, and
  manifest digest
- historical synthetic aliases, embedded bytes, stale synthetic cache entries,
  local fallback paths, and `.gc/system/packs/gastown` cannot satisfy this pin
- behavior rows that would duplicate still-bundled Maintenance definitions are
  present only as inactive compatibility assets or approved semantic deltas
  until the no-Maintenance loader slice activates them

The activation pin is consumed only after Core-owned Maintenance behavior has
moved, Maintenance has been removed from required system packs, and
no-Maintenance packcompat is green through the normal production loader. If the
public pack cannot split compatibility and activation safely, the Gas City
slice must use an atomic pin/removal boundary and cannot land a public pin while
bundled Maintenance remains active.

Duplicate active definitions are fatal in every intermediate and rollback
state. Tests must cover current loader plus candidate pin, old `v1.2.1` binary
plus new pin, new binary plus old pin, rollback after pin adoption, stale local
imports, stale materialized system-pack directories, old synthetic cache, and
ordinary remote cache. The duplicate matrix covers agents, patches, prompt
fragments, prompts, formulas, base formulas, orders, scripts, hooks,
notification targets, runtime-state owners, and moved helper dependencies.

#### Retired-Source Classifier As The Sole Gate
<!-- REVIEW: added per attempt9-retired-source-classifier -->

Retired source handling is centralized before ordinary remote fallback or cache
selection. Config load, pre-resolution doctor, import check/install, packman,
cache reads, lock validation, builtin materialization, prompt/template
discovery, hook/formula/order/script discovery, docs lint, generated-reference
lint, and error rendering must call the same classifier API instead of using
independent path tests or globs.

The classifier returns typed states: active host system pack, active public
pack, generated retired Maintenance, generated retired local Gastown,
preserved stale system-pack directory, stale synthetic cache, legacy runtime
state, custom local fork, historical/generated fixture, duplicate active
definition, invalid Core collision, and unsupported transitive retired import.

The no-Maintenance slice stops materializing, refreshing, pruning, importing,
or loading retired Maintenance/Gastown directories. Existing directories are
preserved and diagnosed; prompt/template/formula/order/script discovery walks
only active resolved packs and required host packs. Tests must cover absent,
stale, customized, explicit legacy import, custom fork, transitive import,
stale prompt fragment, stale hook overlay, stale order/formula, and stale cache
cases through production loaders and import/check/install paths.

#### Doctor Mutation Coordinator And Commit Boundary
<!-- REVIEW: added per attempt9-doctor-transaction-boundary -->

All `gc doctor --fix` writes for this migration go through one mutation
coordinator. Implementation beads may not add direct ad hoc
`MaterializeBuiltinPacks(cityPath)` calls, path-only provenance checks,
whole-file TOML rewrites, live multi-target mutation, or partial commits outside
this coordinator.

The coordinator has one operation-scoped preflight and one publish boundary:

1. Read city manifests, rig manifests, lockfiles, installed-pack metadata,
   runtime-state files, existing system-pack directories, and live-controller
   runtime facts.
2. Refuse automatic multi-file fixes when a controller for the same city is
   running.
3. Prove generated-source provenance for every local import removal/rewrite.
4. Prove public Gastown reachability, exact commit installability, ordinary
   remote cache identity, registry/direct-source equality, and lock generation.
5. Stage Core and public Gastown content outside the live city and validate
   strict required-pack manifests before touching user files.
6. Plan runtime-state moves with conflict checks and reversible temporary names;
   never delete legacy Maintenance state.
7. Edit TOML and lockfiles with a CST/span-preserving editor. If comments,
   unknown tables, unknown fields, array order, whitespace, lexical lockfile
   precision, installed-pack metadata, or unrelated lock entries cannot be
   preserved, automatic fix refuses with manual steps.
8. Validate the staged overlay by running the normal production load, typed
   participation gates, duplicate-definition checks, retired-source classifier,
   and public Gastown pin proof.
9. Publish only after validation, using compare-before-rename on every target.

Before step 9, all live files must remain byte-identical. If any publish step
fails, the error names the exact operation and any already-renamed file; the
recovery instructions use the preserved original and staged temp paths rather
than deleting stale state. Failure-injection tests cover controller-active
cities, air-gapped repair, offline no-cache, failed public fetch, stale locks,
partial writes, failed renames, mutation between preflight and rename, runtime
state conflicts, final validation failure, repeated healthy no-op fixes, and
downgrade after doctor fixes.

#### Go And Asset Role-Surface Migration Table
<!-- REVIEW: added per attempt9-role-surface-table -->

Role neutrality is proved by a checked-in generated table plus scanner, not by
a token scan alone. The canonical table is
`plans/core-gastown-pack-migration/role-surface.generated.yaml`; the public
Gastown companion table is
`gastown/docs/role-surface.generated.yaml`. Each row contains file/asset,
source span, current role-bearing behavior, final owner, replacement mechanism,
rollout slice, proof test, allowlist class, owner, reason, and expiry.

The final decisions are:

- `dog`: Core maintenance-worker compatibility data. It is configurable pack
  data, not an SDK primitive, and Go cannot require it for controller-owned
  infrastructure. Core maintenance formulas resolve the configured worker name
  or fail with a formula/config diagnostic.
- `crew`: not an infrastructure role class. API/session output uses neutral
  config-supplied grouping or session metadata; no Go branch may classify
  `crew` as a special family.
- tmux display identity: role display, icons, and themes are config-supplied
  metadata or Gastown assets. Go APIs such as `DogTheme` are removed or wrapped
  behind neutral compatibility data.
- non-Gastown `gc init`: produces Core/provider infrastructure only. Gastown
  roles, prompts, warmup recipients, and examples come only from the explicit
  public Gastown import.
- prompt fallback and generated help: Core fallbacks are role-neutral.
  Mayor/Deacon/Witness/Polecat/Refinery/Boot/Crew/Gastown prose lives only in
  public Gastown docs/assets or explicit historical fixtures.
- warmup and Core mail/nudge targets: resolved from formula, order, pack, city,
  or session metadata. Missing configured recipients mean no warmup/nudge, not
  a hardcoded fallback.
- sling branch and formula selection: driven by formula metadata or explicit
  config, not role-name heuristics.

Source deletion, Core asset movement, or role-bearing docs cleanup cannot
proceed while any behavior row lacks a replacement mechanism and proof test.
The scanner covers Go identifiers and string literals, TOML, shell, Markdown,
templates, generated CLI help, comments/prose, prompts, scripts, descriptions,
metadata, Core assets, examples, and tests. Allowlist rows must be narrow,
owned, justified, expiring, and covered by negative fixtures that prove a
behavior-bearing role branch fails CI.

#### Behavior Evidence Witness Floor
<!-- REVIEW: added per attempt9-behavior-witness-floor -->

The behavior manifest cannot downgrade evidence. If old behavior had
execution-level evidence, the final row must have execution-level evidence for
the replacement or an approved semantic-delta/removal record. Source existence,
path moves, include counts, or rendered docs are insufficient replacements for
old execution tests.

The generator discovers behavior across the whole Gas City repo and the pinned
public Gastown checkout, including requester, detector, notification, mail,
nudge, route metadata, prompt instructions, scripts, commands, doctor checks,
authorship identity, provider safety/database filters, session hooks such as
`session_live`, dog-field ownership, runtime-state mutation, duplicated
exec/formula paths, and public-pack docs/comments. Each row is per behavior and
per code path when behavior has multiple execution paths.

Every non-equivalent row needs a semantic-delta record with owner, rationale,
operator impact, replacement, release note requirement, and approval status.
`test/packcompat` executes behavioral-trigger fixtures against the exact public
Gastown commit consumed by Gas City and verifies the manifest digest through
the same checkout/cache path used for pack installation.

#### Bootstrap Escape Hatch Final Decision
<!-- REVIEW: added per attempt9-bootstrap-decision -->

`GC_BOOTSTRAP=skip` is retired as a production behavior switch. Production
commands must ignore it for required-system-pack operations or emit a
deprecation diagnostic, but it cannot skip bootstrap fixture setup, Core
materialization, provider-pack materialization, strict integrity validation,
retired-source classification, collision checks, post-load participation, or
doctor cleanup. Any remaining branch that changes production config behavior
based on `GC_BOOTSTRAP=skip` must be deleted in the Core extraction slice.

Production `bootstrapAssets` is a private non-nil empty `fs.FS` whose `Open`
returns `fs.ErrNotExist`. Test fixtures are `_test.go`-only inline
`fstest.MapFS` definitions with this exact allowlist: `pack.toml`, one minimal
agent table when required by the test, and no production-only `formulas/`,
`orders/`, `overlay/`, `skills/`, `assets/prompts/`, scripts, or copied
`internal/packs/core` content. Tests must fail on production imports of
`internal/bootstrap/packs/core`, copied/hardlinked production Core fixtures,
`AssetDir: "packs/core"` outside the fixture allowlist, nil `bootstrapAssets`,
empty-embed leakage into production, and hook overlays resolving from the old
bootstrap Core path.

#### Docs And Operator Vocabulary First-Slice Gate
<!-- REVIEW: added per attempt9-docs-wording-gate -->

`docs/reference/system-packs.md` is the canonical system-pack reference and is
registered in docs navigation in the first slice that changes
operator-facing behavior. No substitute page is allowed unless this design is
edited to name the exact path and the docs-nav test enforces that path before
implementation starts.

The wording matrix is generated from
`plans/core-gastown-pack-migration/system-pack-wording.generated.yaml` and the
pinned public Gastown companion artifact. It has schema version, owners,
allowed contexts, forbidden contexts, exact recommended phrases, consumers, and
golden-test ids. It covers Core, provider host packs, explicit public Gastown,
retired Maintenance, stale system-pack directories, runtime-state migration,
host-Core patching, version skew, store/Dolt maintenance terminology,
Core-maintenance-worker terminology, public-pack comments, and generated
schema/reference text.

Docs lint is manifest-derived and covers Markdown/MDX, TOML, Go strings,
shell, generated references, docs navigation, CLI help, doctor JSON/text
strings, examples, scripts, pack comments, public-pack docs, moved asset names,
retired paths, and troubleshooting. A slice that changes operator-facing
behavior without docs, wording matrix, golden output, generated references, and
navigation updates is explicitly non-release and cannot be the release branch.

#### Slice Gates And Compatibility Fixtures
<!-- REVIEW: added per attempt9-slice-gate-specificity -->

Every high-risk slice has a machine-checkable green state. Broad suite names
are not enough; each slice must name the exact package targets, packcompat mode,
process/integration shard, old/new binary fixture, offline/cache fixture, and
post-deletion stale-path scan it owns.

Compatibility fixtures use:

- old binary: `v1.2.1`
- new binary: the migration branch build under test
- compatibility pin: the first public Gastown pin consumed while Maintenance is
  still active
- activation pin: the public Gastown pin consumed after no-Maintenance
  production loading is green
- cache states: offline no-cache, ordinary remote cache hit, old synthetic
  cache present, stale generated system-pack directory present, and air-gapped
  repair

The slice gate matrix includes:

| Slice | Additional required gates |
| --- | --- |
| public compatibility pin | ordinary remote install, registry/direct hash equality, current-loader zero-duplicate matrix, old `v1.2.1` binary plus compatibility pin, stale synthetic cache ignored, docs marked non-release if wording not updated. |
| Core extraction | `go test ./internal/builtinpacks ./internal/hooks ./internal/bootstrap ./internal/config`, bootstrap fixture guards, old-path scan, `GC_BOOTSTRAP=skip` production-branch deletion/diagnostic test. |
| Core loading/doctor | whole-production loader scanner, API/controller/no-refresh/session/dispatch/import/check/install bypass tests, doctor CST preservation, staged overlay validation, controller-active refusal, failure-injection suite. |
| Maintenance folding | no-Maintenance production-loader packcompat, zero duplicate active definitions, runtime-state migration/conflict tests, stale directory preservation, provider `bd`/`dolt` byte/provenance continuity. |
| registry/cache cleanup | public synthetic alias retirement, source-plus-version repo cache key, old synthetic cache rejection, public pin remote-cache digest check, bundled Core/`bd`/`dolt` validation. |
| source deletion/docs | generated behavior and role table completeness, old-test to new-test mapping, docs wording lint/goldens/navigation, post-deletion stale-path scan, public-import wiring tests. |

Rollback and one-way boundaries are executable. Doctor-mutated manifests must
remain readable by `v1.2.1` unless the release notes name the exact one-way
boundary and manual recovery steps. Runtime-state moves are one-way for new
binaries; rollback guidance must say old binaries may continue reading legacy
Maintenance state until the operator manually moves Core state back.

#### Workflow Artifact Guard Follow-Up
<!-- REVIEW: added per attempt9-workflow-artifact-guard -->

The attempt-local persona-synthesis path issue is not a Core/Gastown migration
requirement, but it remains a required workflow follow-up. The design-review
workflow must write persona syntheses to
`.gc/design-reviews/<source-bead>/attempt-<N>/persona-syntheses/`, stamp every
artifact with `gc.attempt=N`, and fail global synthesis before review if any
required persona synthesis is absent from the current attempt directory or has
mismatched attempt metadata.

### Attempt 10 Review Resolution Contracts
<!-- REVIEW: added per attempt10-global-blockers -->

Attempt 10 closes the remaining ambiguity by making the implementation contract
canonical. These rules supersede older sections if any stale wording survives:
implementation beads must update the stale section in the same slice before
using it as work guidance.

#### Importable Required-System-Pack Boundary
<!-- REVIEW: added per attempt10-required-core-loading -->

Required system-pack loading moves behind an importable internal boundary named
`internal/systempacks`. `cmd/gc` may keep thin command helpers, but production
`cmd/gc` files and production `internal/` packages must call this boundary
instead of reassembling materialization, include lists, or post-load proof.

`internal/systempacks` owns:

- required-pack selection for Core plus provider packs;
- materialization and repair of active required packs only;
- strict pre-resolution file-set integrity;
- production runtime config loading with required includes injected;
- typed post-resolution participation validation;
- no-refresh validation for controller reloads;
- partial-read allowlist checks for diagnostics and editors.

The boundary exposes two fatal gates:

1. Before config resolution, every active required system pack must match the
   embedded source manifest exactly: `pack.toml` digest, full file-set digest,
   permissions, no missing files, and no unexpected effective files. User packs
   or imports named `core`, `bd`, or `dolt` collide with required host packs and
   fail before ordinary remote fallback or behavior discovery.
2. After config resolution, every active required system pack must have a typed
   `RequiredSystemPackParticipation` record. The record includes pack id,
   materialized directory, embedded source id, embedded digest, `pack.toml`
   digest, full file-set digest, resolved layer id, import edge, repair status,
   and load mode. Path-only checks, expected-file-only checks, and helper names
   such as `assertRequiredSystemPackProvenance` are not approval evidence.

The production loader inventory must cover all non-test `cmd/gc` surfaces and
behavior-driving production `internal/` surfaces, including `internal/dispatch`,
`internal/doctor`, `internal/configedit`, controller reload and no-refresh
paths, import/cache/lock validation, packman, prompt/template discovery, API
state helpers that compose behavior, and command helpers. Scanner tests fail on
direct `config.Load*` use outside `internal/systempacks` or a generated
partial-read allowlist row. Each partial-read row names file, function, call
kind, reason, allowed fields, and a focused test proving the call cannot drive
normal runtime behavior.

#### Public Gastown Activation Boundary
<!-- REVIEW: added per attempt10-public-gastown-activation -->

The two-pin model is retained, but both pins are explicit deliverables. Slice 1
in `gascity-packs` writes
`plans/core-gastown-pack-migration/public-gastown-pins.yaml` with two records:

- `compatibility`: public Gastown still supports the last old Gas City binary
  and can coexist with bundled Maintenance.
- `activation`: public Gastown is safe when Maintenance is absent and Gas City
  no longer bundles local Gastown behavior.

Each record contains the immutable commit, durable public ref, ordinary remote
source, subpath, remote-cache digest, behavior-manifest digest, registry/direct
identity proof, duplicate-definition matrix result, old/new binary matrix
result, offline/cache behavior, rollback or one-way-upgrade text, and
packcompat transcript path.

Gas City may update `PublicGastownPackVersion` to the compatibility commit only
in the public-pin compatibility slice. It may remove Maintenance, retire local
Gastown, or delete source only after the activation slice updates
`PublicGastownPackVersion` to the activation commit and the no-Maintenance
packcompat gate passes. If the public pack cannot support two safe commits, the
compatibility and activation slices collapse into one paired cross-repo
activation/removal boundary with release notes naming the one-way point.

#### Host-Core Worker Reference Contract
<!-- REVIEW: added per attempt10-host-core-routing -->

Chosen model: public Gastown depends on host Core being inserted by Gas City and
does not import Core. Core owns a configurable maintenance-worker binding with
default key `dog`; `dog` is compatibility configuration, not SDK role logic.

Under public-pack bindings:

- `[[patches.agent]] name = "dog"` may patch only the resolved host-Core
  maintenance-worker agent or its explicit compatibility alias.
- Formula pools, `gc.routed_to`, mail/nudge targets, warrant metadata, prompt
  examples, and runtime routing for Core maintenance behavior resolve the
  configured maintenance-worker key from Core pack/config data.
- Gastown-specific roles, routes, pools, prompts, mail, and nudges live in the
  public Gastown pack and are not Core fallbacks.
- Missing host Core, missing maintenance-worker binding for a Core maintenance
  formula, or a public Gastown patch targeting an omitted/renamed Core worker
  fails with a host-Core diagnostic before behavior executes.
- SDK infrastructure controlled by the controller must still run when the Core
  maintenance worker is renamed or omitted; only Core maintenance formulas that
  actually require a worker may fail with the diagnostic above.

Required tests cover host Core present, host Core missing, maintenance worker
renamed, maintenance worker omitted, public Gastown patch application, formula
pool resolution, `gc.routed_to` metadata, mail/nudge recipients, prompt
fragment ownership, and generated `pack.toml` defaults.

#### Retired-Source Classifier Call-Site Table
<!-- REVIEW: added per attempt10-retired-source-classifier -->

Retired-source handling is centralized in `internal/packsource`. All call sites
must consume the same classifier result before load, repair, install, cache, or
discovery logic can act on a source.

Classifier states are: active required host pack, active provider host pack,
active public Gastown, generated retired Maintenance, generated retired local
Gastown, historical public Gastown synthetic alias, historical public
Maintenance alias, stale synthetic cache, retired import, retired lock entry,
preserved stale system directory, preserved runtime state, custom local fork,
legacy fixture, historical docs, and unknown.

The call-site table covers config load, pre-resolution doctor/import recovery,
`gc pack install/check`, cache lookup, lock validation, packman, builtin
materialization, prompt/template discovery, hook/formula/order/script
discovery, docs lint, generated-reference lint, public-source normalization,
`publicSubpathForPack`, `PublicRepository`, `normalizeRepository`,
`requiredBuiltinPackNames`, `RetiredBootstrapPacks`, and materializer pruning.
Generated retired directories and locks are diagnosed or migrated before
ordinary remote fallback. Custom directories are preserved and excluded from
active content discovery; no automatic fix deletes them.

#### Doctor Mutation Coordinator
<!-- REVIEW: added per attempt10-doctor-coordinator -->

`gc doctor --fix` becomes coordinator-owned for this migration. Pack/Core/import
and runtime-state checks may report fix intents, but they must not directly
write manifests, lockfiles, installed pack directories, or runtime-state paths.

The doctor runner adds a `doctor.MutationCoordinator` that executes one staged
plan per run. A fix intent records id, owner check, preflight reads and content
digests, writes, lock/install operations, runtime-state copy or move
operations, validation hooks, rollback notes, offline guidance, and publish
record fields. The coordinator:

- refuses automatic fixes when a live controller for the city is detected from
  runtime facts;
- refuses when another live doctor process is mutating the same city, detected
  from live process state rather than persistent status or lock files;
- validates public Gastown reachability, exact pin installability, lock
  parseability, Core file-set integrity, generated-source provenance, editable
  TOML spans, and runtime-state conflicts before staging any write;
- stages all file changes outside active paths, compares every target's
  preflight digest immediately before rename, and publishes with atomic renames;
- writes a publish record after success and validates typed Core participation,
  public Gastown lock/install state, retired-source exclusions, and runtime-state
  convergence before returning success;
- preserves generated/custom provenance and refuses automatic fixes for custom
  forks, edited retired directories, air-gapped missing pins, downgrade-unsafe
  manifests, and ambiguous runtime-state conflicts.

Failure-injection fixtures must cover preflight failure, publish failure after
some staged files exist, target mutation before rename, active controller,
concurrent doctor, offline public pin, custom fork, partial lock rewrite,
runtime-state copy/move conflicts, generated-source provenance mismatch, and
rerun convergence after an interrupted attempt.

#### Behavior Evidence First Slice
<!-- REVIEW: added per attempt10-behavior-evidence -->

The behavior manifest and public replacement evidence are the first blocking
implementation gate. No source move, role-neutral rewrite, registry/cache
retirement, public pin update, or docs claim may precede a complete generated
manifest row for the affected behavior.

Each row includes old path, old owner, new path, new owner, public landing
commit, consuming Gas City pin, trigger, requester semantics, detector
semantics, channel, recipient, author identity, route metadata, runtime-state
effect, script branch, prompt/template fragment, old witness, final witness,
packcompat witness, semantic delta, removal approval, and helper dependencies.
Execution-level old witnesses require execution-level final witnesses; source
or path assertions cannot replace command/formula/hook/script behavior proof.

Packcompat must run against the exact commit named by the consuming pin and
must prove requesters, detectors, notifications, channels, recipients,
authorship identity, route metadata, runtime-state mutation, script branches,
prompt fragments, Dog flows, Polecat and branch-pruning moves, commands,
doctor checks, providers, and examples that the row claims to preserve.

#### Bootstrap And Implicit-Import Final Rule
<!-- REVIEW: added per attempt10-bootstrap-implicit-import -->

`GC_BOOTSTRAP=skip` is retired as a production behavior switch. It may only skip
the now-empty bootstrap fixture materialization path in tests. Normal production
commands still run `internal/systempacks` materialization, strict file-set
integrity, retired-source classification, collision checks, and post-load typed
participation validation.

`bootstrapManagedImportNames` becomes empty in the Core extraction slice.
Legacy implicit imports named `core` or `registry` are classified before config
resolution: generated fixture entries are test-only, active user entries named
`core` collide with required host Core, and `registry` entries are stale legacy
diagnostics. Production `bootstrapAssets` is a private non-nil empty/erroring
`fs.FS`. All real Core content fidelity tests move to `internal/systempacks`;
bootstrap tests may use only inline `_test.go` fixtures and must fail if they
pull production Core assets from `internal/packs/core` or
`internal/bootstrap/packs/core`.

#### Generated Artifact Ownership And Freshness
<!-- REVIEW: added per attempt10-generated-artifacts -->

Generated artifacts are source contracts, not prose. Each generated file has an
entrypoint, schema, owner, digest test, and CI freshness test:

| Artifact | Entrypoint | Schema | Owner | Freshness test |
| --- | --- | --- | --- | --- |
| Behavior manifest | `go run ./scripts/gen-behavior-manifest` | `plans/core-gastown-pack-migration/behavior-manifest.schema.json` | Gas City migration owner plus public Gastown owner for public rows | `TestBehaviorManifestFresh` |
| Role-surface table | `go run ./scripts/gen-role-surface` | `plans/core-gastown-pack-migration/role-surface.schema.json` | Gas City migration owner | `TestRoleSurfaceFresh` |
| System-pack wording matrix | `go run ./scripts/gen-system-pack-wording` | `plans/core-gastown-pack-migration/system-pack-wording.schema.json` | docs owner | `TestSystemPackWordingFresh` |
| Old-test to new-test map | `go run ./scripts/gen-test-migration-map` | `plans/core-gastown-pack-migration/test-migration.schema.json` | test owner for each repo | `TestTestMigrationMapFresh` |

If an artifact is hand-owned, remove any `generated` label and add a direct
schema validation test instead. Generated docs, CLI/help text, doctor output,
schemas, tutorials, examples, pack comments, script comments, and public
Gastown companion docs consume the wording matrix or fail docs lint.

#### Examples, Docs, And Command-Level Gates
<!-- REVIEW: added per attempt10-examples-docs-gates -->

`examples/gastown` import rewiring and test splitting are assigned to the
public-pin compatibility slice or earlier; they cannot wait until after
Maintenance folding. Before Core-owned Maintenance behavior moves or
Maintenance stops loading, the example must no longer import local
`../maintenance`, and every removed `examples/gastown` or `examples/dolt` test
must have a test-migration row naming its public Gastown, Core, provider, or
approved-retirement replacement.

Docs and operator-facing text move in the same slice as the behavior they
describe. The canonical `docs/reference/system-packs.md` baseline, docs
navigation/index checks, runtime-state migration table, doctor output goldens,
CLI/help goldens, generated reference/schema lint, tutorial transcripts, public
Gastown companion artifact, and case-aware docs lint are release gates for the
first slice that changes the corresponding behavior. A slice may defer docs
only by marking itself non-release and adding a failing release gate.

#### Design-Review Artifact Guard
<!-- REVIEW: added per attempt10-workflow-artifact-guard -->

Future design-review approval evidence must be attempt-local. Persona synthesis
beads write outputs under the current attempt directory, and global synthesis
must either copy the exact source files it consumed into the current attempt or
write a manifest with bead id, path, digest, model, persona id, and
`gc.attempt`. A pre-global guard fails if any consumed file path, metadata, or
digest does not match the current attempt.

### Attempt 11 Review Resolution Contracts
<!-- REVIEW: added per attempt11-global-blockers -->

Attempt 11 keeps the review verdict at `iterate` and makes the remaining
contracts executable. These rules supersede any earlier section that still
allows inactive compatibility assets, path-only Core proof, ad hoc doctor
writes, role-token-only inventory, bundled public Gastown fallback, or deferred
docs.

#### Concrete Core Maintenance-Worker Binding
<!-- REVIEW: added per attempt11-maintenance-worker-binding -->

Chosen ownership model: Core owns the maintenance-worker binding. Public
Gastown does not define a replacement Core worker and does not import Core.
Public Gastown may set or patch the host-Core binding through normal config
layering, and Gastown-specific workers, routes, prompts, mail, nudges, formulas,
and examples remain public-Gastown assets.

Core `pack.toml` introduces one binding table:

```toml
[gc.bindings.maintenance_worker]
default_agent = "dog"
optional_for_controller = true
diagnostic = "core.maintenance_worker"
```

City, rig, and imported-pack layers may override only the binding value:

```toml
[system_packs.core.bindings]
maintenance_worker = "ops-maintainer"
```

An explicit empty value means the operator omitted the worker. Omission must
not prevent Core materialization, config loading, controller-owned SDK
operations, beads, events, formulas, or ordinary session management. It does
disable worker-bound Core maintenance orders and formulas at dispatch time with
diagnostic `core.maintenance_worker.omitted`. A configured value that names no
effective agent fails before dispatch with `core.maintenance_worker.missing`.

Formula and order routing uses binding references instead of role literals:

- order TOML uses `target_binding = "core.maintenance_worker"` for any
  worker-bound Core order.
- formula step metadata uses `gc.run_target_binding = "core.maintenance_worker"`
  instead of hardcoding `gc.run_target` or `gc.routed_to`.
- scripts receive `GC_CORE_MAINTENANCE_WORKER` from the resolved binding and
  must treat an empty value as "skip optional notification" or fail with the
  diagnostic above when the step requires a worker.
- prompt fragments referenced by Core maintenance formulas resolve from Core
  only; Gastown prompt fragments resolve from public Gastown only.

Patchability is symbolic. Public Gastown patches the binding target with:

```toml
[[patches.agent]]
target_binding = "core.maintenance_worker"
```

`[[patches.agent]] name = "dog"` remains a compatibility-row-only form for old
public packs and must be removed from active public Gastown assets by the
activation pin. A patch that targets an omitted binding or a missing host Core
fails with `core.maintenance_worker.patch_target_missing`; it does not create a
new Core agent. Tests must cover default `dog`, renamed worker, omitted worker,
public Gastown symbolic patching, old `name = "dog"` compatibility diagnostics,
formula pool resolution, `gc.routed_to` emission, mail/nudge recipient
resolution, prompt fragment ownership, and rendered prompt output.

#### Activation Pin Has No Inactive Loader Fiction
<!-- REVIEW: added per attempt11-activation-proof -->

There is no inactive compatibility-asset mechanism in the production loader.
The compatibility pin must omit colliding formulas, orders, prompt fragments,
scripts, agents, patch targets, hooks, and runtime-state owners from active
discovery paths. Compatibility-only notes may live in docs or manifest rows,
but not under active `formulas/`, `orders/`, `agents/`, `overlay/`, `assets/`,
or script paths that the loader can discover.

`public-gastown-pins.yaml` is an asset-level ledger. Each record contains:
phase (`compatibility` or `activation`), immutable commit, durable ref, public
source, subpath, pack digest, behavior-manifest digest, active asset digest,
ordinary remote-cache key, registry/direct-source equality proof, old-binary
evidence, new-binary evidence, offline/cache fixture result, duplicate-active
matrix result, rollback or one-way-upgrade status, and transcript path.

The no-Maintenance proof runs only after the candidate Gas City slice stops
including Maintenance through the normal production loader. The proof must run
through `internal/systempacks.LoadRuntimeCity` or
`LoadRuntimeCityNoRefresh`; a copied fixture, direct `config.Load*`, hidden
bootstrap Core path, or test-only include list cannot satisfy it. If the
compatibility public pack cannot omit colliding active assets, the project must
skip the compatibility pin and land a paired cross-repo activation/removal
boundary instead.

#### Required Core Loading Fatal Gates
<!-- REVIEW: added per attempt11-required-core-fatal-gates -->

`internal/systempacks` is deny-by-default. Every production config load must be
classified as runtime, no-refresh runtime, partial doctor, partial edit, or
test-only fixture before it can call any lower-level config loader. Unclassified
production calls fail the scanner.

Runtime and no-refresh runtime loads have two uniform fatal gates:

1. Pre-resolution required file-set integrity: Core and provider-required host
   packs must be selected after reading the final effective beads provider,
   materialized from the expected embedded source, checked against exact
   manifests and `pack.toml` digests, and collision-checked before any
   behavior-bearing config, formula, order, script, overlay, prompt, or patch is
   read.
2. Post-resolution typed participation: the final resolved config must contain
   `RequiredSystemPackParticipation` for the same validated pack layer. The
   record includes pack id, source id, embedded digest, full file-set digest,
   `pack.toml` digest, materialized path, resolved layer id, import edge id,
   load mode, repair state, collision state, and validation diagnostic.

The production loader inventory is generated and checked in as
`plans/core-gastown-pack-migration/loader-inventory.generated.yaml`. It covers
`cmd/gc`, controller reload/no-refresh, API state helpers that compose behavior,
session resolution, dispatch/sling, formula/order loading, prompt/template
discovery, import/check/install, cache/lock validation, packman, doctor after
pre-resolution, configedit, init readiness, convoy/wait flows, Dolt
publication, and helper packages called by those surfaces. Each allowlist row
names file, function, loader class, fields consumed, fatality boundary, owner,
expiry, and focused test. `bootstrapManagedImportNames` and any
`GC_BOOTSTRAP=skip` branch can be removed or narrowed only after collision
checks, implicit-import checks, and the fatal gates above are live.

#### Full Role-Surface Inventory And Replacement Gate
<!-- REVIEW: added per attempt11-role-surface-inventory -->

Role neutrality is proven by the generated role-surface table plus a scanner,
not by a token scan alone. The inventory roots are:

- production Go under `cmd/gc` and `internal/`
- dashboard/API/OpenAPI/generated TypeScript and generated schema artifacts
- `internal/packs/core`, provider packs, public Gastown checkout, overlays,
  formulas, orders, scripts, prompts, templates, skills, examples, and tests
- `examples/bd`, `examples/dolt`, `examples/gastown`, provider formulas and
  scripts, generated CLI help, docs navigation, docs, `.mdx`, `.json`, `.txt`,
  `.toml`, shell, and public-pack docs

Every row names source span, behavior, final owner, replacement mechanism,
proof test, allowed context, owner, and expiry. Hardcoded `dog`, `mayor`,
`deacon`, `witness`, `refinery`, `polecat`, `boot`, `crew`, and `gastown`
surfaces must either move to configured resolution or have a narrow compatibility
row. Compatibility rows are allowed only for historical fixtures, docs examples,
old-pack diagnostics, or Core `maintenance_worker` compatibility data, and each
row has a release gate that fails after its expiry slice.

The scanner covers API/dashboard projections such as `agent_kind` and `crew`,
TOML defaults, provider `dog` bindings, notification targets, prompt prose,
scripts, overlays, generated docs, generated schemas, and moved asset names
such as `prune-branches`. A source deletion, Core move, public pin update, or
role-neutral rewrite cannot land while any behavior-bearing row lacks a
replacement mechanism and executable witness.

#### Enforceable Doctor Mutation Protocol
<!-- REVIEW: added per attempt11-doctor-mutation-protocol -->

`Check.Fix` implementations for this migration may only return `FixIntent`
values. They must not directly mutate city manifests, rig manifests, lockfiles,
installed-pack directories, generated system-pack directories, public pack
cache entries, or runtime-state paths.

Each `FixIntent` records id, owner, generated/custom provenance proof,
preflight reads, expected digests, planned writes, lock/install operations,
runtime-state copy plan, validation hooks, publish order, rollback note,
operator-facing report fields, and post-publish rerun requirements.
`doctor.MutationCoordinator` accepts a set of intents and executes one staged
plan:

1. phase-one report: classify intended changes and refusal reasons without
   mutating live files;
2. preflight: validate public Gastown reachability, exact pin installability,
   ordinary remote-cache identity, Core strict file-set integrity, editable TOML
   spans, parseable locks, generated-source provenance, and runtime-state
   conflicts;
3. concurrency: refuse if a live controller owns the city; otherwise take a
   crash-released OS advisory lock on the city directory file descriptor, not a
   PID, status, or lock file;
4. stage: write all candidate outputs outside active paths and validate the
   staged overlay through the normal production loader;
5. publish: compare every target digest immediately before rename, then publish
   with atomic renames;
6. phase-two report: rerun Core participation, public Gastown pin proof,
   retired-source exclusion, duplicate-definition checks, runtime-state
   convergence, and doctor diagnostics before reporting success.

Runtime-state migration is non-destructive. JSONL state/archive and
spawn-storm ledgers are copied to Core when the Core destination is absent,
with source and destination digests recorded. Legacy Maintenance state remains
in place and is classified as ignored legacy state. If both sides exist and
their digests differ, automatic fix refuses with a manual conflict diagnostic.
Healthy cities must be byte-identical after `gc doctor --fix`; repeated fixes
must converge to the same phase-two report.

#### Pin, Cache, Registry, And Offline Gates
<!-- REVIEW: added per attempt11-pin-cache-registry -->

`public-gastown-pins.yaml` is authoritative for
`internal/config/PublicGastownPackVersion`. Gas City may not update that
constant unless the ledger validator proves the target phase, commit, digest,
ordinary remote-cache key, active asset set, duplicate matrix, old-binary
evidence, offline behavior, and transcript fields are present and match the
checkout used by pack installation.

`config`, `packman`, lock validation, import check/install, and cache lookup
must use the same public cache identity: normalized public source plus immutable
commit plus subpath. Historical synthetic aliases for public Gastown and
Maintenance are retired-source diagnostics only. They cannot satisfy a public
pin, lock refresh, direct source install, stale-cache repair, or offline init.
`SyntheticContentHash` remains only for bundled Core, `bd`, and `dolt`.

Active materialization and retired diagnostics are separate sets. The
materializer repairs only Core and provider host packs. Retired Maintenance,
retired local Gastown, stale synthetic caches, and stale system-pack directories
are classified before fallback, preserved on disk, and excluded from prompt,
template, formula, order, script, hook, overlay, and patch discovery. Fixtures
must cover ordinary remote cache hit, offline no-cache failure, stale synthetic
cache present, air-gapped provider self-heal for `bd` and `dolt`, concurrent
self-heal contention through the advisory directory lock, old binary/new pack,
new binary/old pack, and downgrade after doctor fixes.

#### Docs, Wording, And Moved-Asset Gates
<!-- REVIEW: added per attempt11-docs-wording-gates -->

The wording matrix remains
`plans/core-gastown-pack-migration/system-pack-wording.generated.yaml` and is
owned by the docs owner. It contains schema version, owner, source artifacts,
allowed contexts, forbidden contexts, exact recommended phrases, replacement
phrases, moved asset names, consumers, freshness test ids, and release-gate ids.

The docs scan is extension-agnostic and covers docs navigation, Markdown, MDX,
JSON, TXT, generated schemas, generated references, TOML, Go strings, shell,
TypeScript, CLI help, doctor JSON/text, examples, pack comments, script
comments, public-pack docs, troubleshooting, and tutorial transcripts. It must
classify every retained Maintenance/Gastown path or term as retired-pack
diagnostic, stale legacy state, runtime-state migration, valid store/Dolt
maintenance terminology, Core `maintenance_worker` terminology, public Gastown
reference, historical fixture, or error.

The matrix explicitly tracks moved asset names, including `prune-branches`,
`mol-polecat-*`, Dog prompt fragments, shutdown-dance examples, provider
`dog` formulas, and stale `.gc/system/packs/*` and `.gc/runtime/packs/*` paths.
A behavior-changing slice must update docs navigation, generated references,
schema text, doctor output goldens, CLI/help goldens, tutorial transcripts, and
public Gastown companion docs in the same slice, or mark the slice non-release
with a failing release gate.

#### Attempt 11 Unfixable Items
<!-- REVIEW: added per attempt11-unfixable-items -->

None. All blocker and major findings are addressed as explicit ownership
choices, config surfaces, fatal gates, generated inventories, mutation
protocols, ledger fields, or slice-level proof requirements.

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

Before any Gas City source deletion, Maintenance removal, or Core
generalization lands:
<!-- REVIEW: added per blocker-behavior-preservation -->

- The `gascity-packs` PR or branch must be named in the implementation bead.
- The replacement commit must be immutable and recorded in the manifest. Gas
  City's `internal/config/PublicGastownPackVersion` must already name the
  compatibility commit for coexistence work, and must name the activation commit
  before Maintenance is removed or source is deleted.
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
At compatibility-pin adoption time the test runs in current-loader compatibility
mode. At activation-pin adoption time, before Maintenance is removed from
required system packs, the same gate must run in no-Maintenance
production-loader mode: clone or install
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
| `internal/systempacks.MaterializeRequiredPacks` | Generates/repairs Core, `bd`, and `dolt`; does not refresh or prune Maintenance/Gastown. | Missing/corrupt Core repair tests plus stale custom-file preservation tests. |
| Orders and scripts | Core-owned generic orders resolve from Core; Gastown-owned orders resolve from public Gastown. | Formula/order composition tests and pack-relative script execution tests. |
| Order skip lists | Preserved names continue to match; renamed orders have aliases or migration tests. | Existing skip-list compatibility tests. |
| Runtime state | JSONL state/archive and spawn-storm ledgers migrate to Core; conflicting legacy state is manual; other stale Maintenance paths are ignored legacy. | Doctor diagnostics and first-run script migration tests; no deletion on `--fix`. |
| Doctor import fixes | Generated/example Maintenance imports removed; custom forks preserved with manual guidance. | Golden TOML preservation and fork-provenance tests. |
| Synthetic cache validation | Remaining bundled packs keep validation; retired aliases are never selected for new resolutions. | Stale synthetic-cache rejection tests. |

### System Pack Loading

Move the production system-pack loading boundary to `internal/systempacks`:

- `RequiredPackNames(cityPath)` starts with Core and adds provider-dependent
  `bd` and `dolt` as today.
- Maintenance and Gastown are never required host packs and never appear in
  active required-pack includes.
- `MaterializeRequiredPacks`, `ValidateRequiredFileSets`,
  `RuntimeIncludes`, `LoadRuntimeCity`, and `LoadRuntimeCityNoRefresh` are the
  single production entrypoints for required system-pack materialization,
  no-refresh validation, include injection, and typed participation proof.
- `cmd/gc/embed_builtin_packs.go` becomes a thin compatibility wrapper during
  migration, then either disappears or delegates entirely to
  `internal/systempacks`.
- Comments and diagnostics change from "Core and maintenance" to "Core and
  provider host packs" where provider packs are relevant.

Update config-loading call sites:
<!-- REVIEW: added per blocker-required-core-loading -->

- Normal command paths and behavior-driving production `internal/` paths must
  call `internal/systempacks.LoadRuntimeCity` or
  `internal/systempacks.LoadRuntimeCityNoRefresh`.
- The wrapper must run two fatal gates: strict required file-set integrity
  before resolution, then typed `RequiredSystemPackParticipation` validation
  after resolution. A successful materialization without typed participation is
  a load failure.
- Direct `config.Load`, `config.LoadCity`, `config.LoadWithIncludes`, package
  aliases of those functions, and hand-built required-pack include lists in
  production `cmd/gc` and production `internal/` files are rejected by scanner
  tests modeled on `TestGCNonTestFilesStayOnWorkerBoundary`.
- Scanner coverage includes `internal/dispatch`, `internal/doctor`,
  `internal/configedit`, controller reload/no-refresh paths, import/cache/lock
  validation, packman, prompt/template discovery, API state helpers that
  compose behavior, and command helpers.
- Any production exception must be generated into the partial-read allowlist
  with file, function, call kind, reason, fields consumed, and a focused test
  proving it intentionally reads partial/broken config and cannot drive normal
  runtime behavior.
- No-refresh or diagnostic config helpers must either run the same pre-load
  integrity and post-load participation gates or appear in that allowlist.
- Low-level `internal/config` tests may continue to call `config.LoadWithIncludes`
  directly because those tests exercise the config package, not full `gc`
  runtime behavior.

The dev/test escape hatch is not a CLI flag or environment variable. Production
`gc` commands always include Core. Tests that need no-Core behavior should call
the lower-level config loader directly or use a `_test.go` helper that bypasses
`internal/systempacks`.

### Core Presence Doctor

Add a new doctor check, recommended file:
`cmd/gc/core_pack_doctor_check.go`.

The check should:

- Materialize intent: Core is required for real cities.
- Verify `.gc/system/packs/core/pack.toml` and the full Core file set match
  `internal/systempacks` embedded manifests using the same strict integrity
  gate as runtime config loading.
- Load resolved config through `internal/systempacks.LoadRuntimeCity` and verify
  typed `RequiredSystemPackParticipation` includes Core.
- Return Error or Warning according to the existing doctor severity convention
  for required repairable checks.
- Include details naming the missing condition: missing materialized Core,
  invalid materialized Core, unexpected effective files, host-pack collision,
  or Core absent from typed resolved-config participation.
- Set `FixHint` to run `gc doctor --fix`.

`Fix` must not write directly. It should emit a coordinator intent that:

- stages required Core repair through `internal/systempacks`;
- validates public Gastown pin state when the same run also rewrites imports;
- re-runs strict file-set integrity and typed participation validation before
  publish completes;
- returns an error if Core is still absent, because production opt-out is not
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

- Pack/Core/import and runtime-state checks submit fix intents to the
  `doctor.MutationCoordinator`; direct per-check writes are forbidden for this
  migration.
- Preflight must run before any mutation. It verifies the public Gastown source
  and immutable version are reachable, installable, and lockable; generated Core
  can satisfy strict file-set integrity; existing lockfiles are parseable; city
  manifests can be edited without dropping unknown syntax; and runtime-state
  conflicts are classified.
- Fixes must be failure-atomic across `city.toml`, rig `pack.toml` files,
  lockfiles, installed pack directories, and migrated runtime-state paths. Use
  staged writes plus temp-file-plus-rename for each file and leave the city
  byte-identical if any preflight step fails.
- The coordinator refuses automatic fixes when a controller for the same city or
  another live doctor mutation process is discovered from live runtime state.
- Immediately before every rename, the coordinator re-reads the target and
  aborts if its content changed since preflight.
- After successful publish, the coordinator writes a structured publish record
  for diagnostics and revalidates typed Core participation, public Gastown
  install/lock state, retired-source exclusions, and runtime-state convergence.
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
- After a fix, doctor must re-run strict Core file-set integrity, typed
  participation validation, and public Gastown lock/install validation. A fix
  that leaves Core absent or the public pack unresolved is an error.
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
- Retire `GC_BOOTSTRAP=skip` as a production behavior switch. If retained for
  tests, it may skip only empty bootstrap fixture materialization; it must not
  skip `internal/systempacks` materialization, required Core file-set integrity,
  retired-source classification, collision checks, or typed participation
  validation.
- Add hook installation tests proving `internal/hooks` reads overlays from
  `internal/packs/core` after bootstrap no longer owns Core assets.

### Examples And Docs
<!-- REVIEW: added per blocker-docs-dx-consistency -->

Update `examples/gastown`:

- Keep it as an example city only if it can use public Gastown imports without
  vendoring `packs/gastown`.
- Rewire `examples/gastown/pack.toml`, `examples/gastown/city.toml`, and nested
  legacy imports to public Gastown in the public-pin compatibility slice or
  earlier. This must happen before Maintenance folding.
- Remove `examples/gastown/packs/maintenance` only after Core/public Gastown
  replacement witnesses and the activation-pin no-Maintenance gate are green.
- Remove `examples/gastown/packs/gastown` only after public-import wiring tests
  and old-test to new-test mapping are green.
- Update comments to describe Core as the required system pack and Gastown as
  the explicit external pack.
- Update tests that currently read `examples/gastown/packs/gastown` to either
  move to `gascity-packs` or assert the public-import wiring.
- Every removed or rewritten `examples/gastown` and `examples/dolt` test must
  have a generated test-migration row naming its new Core, provider, public
  Gastown, or approved-retirement witness.

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
  longer system packs. This page is the canonical baseline and must be linked
  from docs navigation before operator-facing behavior changes ship.
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
- Runtime-state migration tables, public Gastown companion docs, generated
  reference/schema lint, docs navigation/index checks, and tutorial transcripts
  are release gates in the same slice as the behavior they describe. A slice
  can defer them only by marking itself non-release and adding a failing release
  gate.

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
- `internal/systempacks.MaterializeRequiredPacks`: no Maintenance or Gastown
  materialization expected; Core includes moved maintenance orders and scripts.
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
- Production config-load guard: non-test `cmd/gc` and behavior-driving
  production `internal/` files may not call `config.Load*` directly outside the
  generated partial-read allowlist.
- Required-pack load gates: normal runtime helpers fail closed when strict
  pre-resolution file-set integrity fails or typed post-resolution
  `RequiredSystemPackParticipation` is missing.
- Behavior proof for Core formulas and orders: compose representative formulas
  into molecules, assert expected step counts, resolve hooks, load configured
  agents and pools, and execute moved scripts from pack-relative paths.
- Doctor fix safety: idempotent healthy-city no-op, scoped TOML preservation,
  fork/custom-source manual diagnostics, failure-atomic preflight failure, final
  Core provenance revalidation, and runtime-state diagnostic classes.
- Provider pack continuity: `bd` and `dolt` materialized bytes and provenance
  remain unchanged except for expected manifest metadata, and provider matrices
  still resolve the correct required packs.
- Core integrity: tampered Core is repaired; unexpected effective files fail
  strict full file-set integrity validation before behavior discovery.
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
- `go test ./cmd/gc ./internal/systempacks -run 'BuiltinPack|ImportStateDoctor|CorePack|TryReloadConfig|MaterializeRequiredPacks|RequiredSystemPackParticipation'`
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
   cross-pack ownership audits, update Gastown tests, and write
   `public-gastown-pins.yaml` with compatibility and activation records. This
   slice must decide `mol-review-quorum`, provider overlays, Dog prompt
   fragments, Polecat formulas, branch pruning, shutdown-dance examples, review
   checks, host-Core worker references, and hardcoded role-theme/tmux APIs
   before either pin is consumed. Required gates: `gascity-packs` test suite,
   manifest completeness, formula/order composition, prompt-template
   resolution, script execution, clean scan for retired Maintenance paths, and
   old-test to new-test mapping.
2. Gas City public-pin adoption and packcompat slice:
   update `internal/config/PublicGastownPackVersion` to the immutable public
   compatibility commit from slice 1 and add the Gas City pinned-public-pack
   compatibility test in current-loader mode without deleting in-tree sources or
   claiming no-Maintenance production-loader proof. Rewire `examples/gastown`
   away from local `../maintenance` in this slice or earlier.
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
   add `internal/systempacks`, strict pre-resolution file-set integrity, typed
   `RequiredSystemPackParticipation`, production `config.Load*` scanner,
   call-site migration/allowlist, Core doctor, pre-resolution import recovery,
   version-skew diagnostics, and the doctor mutation coordinator.
   Required gates:
   `go test ./cmd/gc ./internal/systempacks -run 'BuiltinPack|ImportStateDoctor|CorePack|TryReloadConfig|MaterializeRequiredPacks|RequiredSystemPackParticipation'`
   plus doctor golden/failure-atomic tests, `make test-fast-parallel`, and
   `go vet ./...`.
5. Public Gastown activation and Maintenance folding slice:
   update `internal/config/PublicGastownPackVersion` from the compatibility
   commit to the activation commit from slice 1, then run the no-Maintenance
   production-loader packcompat gate before removing Maintenance from active
   required packs. If the activation commit cannot support old/new binary and
   duplicate-definition requirements, stop and convert this to a paired
   cross-repo activation/removal boundary.
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
   the already-updated activation `PublicGastownPackVersion` from slice 5; it
   must not be the first pin update. Required gates: registry/cache negative
   tests and packcompat test against the exact public pin,
   `make test-fast-parallel`, and `go vet ./...`.
7. Source deletion/docs slice:
   remove in-tree `examples/gastown/packs/*` sources only after replacement
   tests are green, migrate any remaining docs and examples, and apply docs
   lint/golden checks. Behavior-changing docs and generated references should
   already have landed in the slice that changed that behavior; this slice is
   only the final stale-source and stale-doc cleanup. Required gates: docs
   inventory/lint, tutorial/doctor output golden tests, generated freshness
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
