---
plan_slug: core-gastown-pack-migration
phase: design
rig: gascity
rig_root: /data/projects/gascity-fresh-main-20260604-VLKm8c
artifact_root: /data/projects/gascity-fresh-main-20260604-VLKm8c/plans
requirements_file: /data/projects/gascity-fresh-main-20260604-VLKm8c/plans/core-gastown-pack-migration/requirements.md
status: draft
created_at: 2026-06-04T15:07:35Z
updated_at: 2026-06-05T20:30:00Z
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
  has a behavior inventory row, and is verified from the exact pinned commit.
- `gc doctor --fix` is failure-atomic and byte-identical on healthy manifests.
- Stale `.gc/system/packs/maintenance` and `.gc/system/packs/gastown`
  directories are ignored, diagnosed, and preserved, not deleted.
- Core assets remain role-neutral outside explicitly allowed Core maintenance
  configuration.

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
- Keep `examples/gastown` only as an example city, if still useful, with public
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

Add a behavior inventory in the `gascity-packs` repo, required path:
`gastown/docs/behavior-preservation.md`.

The inventory must list every moved or generalized asset with:

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

This inventory is required for these high-risk moves:

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
longer describe an implicit Maintenance layer. Gastown should not import Core;
Core remains a required host system pack. Gastown may continue to patch the
Core `dog` agent for theming or work_dir behavior if that patch is the explicit
Gastown behavior, and the inventory should name that dependency.

Before any Gas City source deletion or Core generalization lands:
<!-- REVIEW: added per blocker-behavior-preservation -->

- The `gascity-packs` PR or branch must be named in the implementation bead.
- The replacement commit must be immutable and recorded in both the inventory
  and Gas City's `internal/config/PublicGastownPackVersion` value.
- The public pack must be fetched from the ordinary remote-pack install path,
  not from a bundled synthetic alias.
- Gas City CI must run a gate that installs that exact commit into a fresh test
  city, composes the moved formulas/orders, resolves moved scripts using
  pack-relative paths, verifies hook overlays and configured agents, and checks
  every behavior inventory row.
- The gate must fail if any test from `examples/gastown/gastown_test.go` or
  `examples/gastown/maintenance_scripts_test.go` is removed without an explicit
  row mapping it to a new `gascity-packs` test or a documented intentional
  behavior removal.

Recommended Gas City gate name:
`go test ./test/packcompat -run TestPinnedPublicGastownBehavior`.
The test should clone or install `github.com/gastownhall/gascity-packs/gastown`
at `PublicGastownPackVersion`, load it with host Core present and no
Maintenance pack, and assert that the behavior inventory is complete for the
removed in-tree assets.

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
- Fresh `gc init --template gastown` behavior must make a product decision:
  network-required public install is acceptable if the error clearly says the
  public Gastown pack could not be fetched. Offline fallback to embedded old
  Gastown is not allowed unless a deliberate compatibility cache is specified
  and tested against the same immutable public commit.

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
| Runtime state | Maintenance runtime paths are ignored-legacy, migrated to Core, or manual-diagnostic per state file. | Doctor diagnostics for each state-path class; no deletion on `--fix`. |
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
- Runtime state under `.gc/runtime/packs/maintenance` must be classified before
  docs or diagnostics claim it is gone. JSONL archives, export cursors, storm
  ledgers, order tracking, and `jsonl_archive_doctor_check.go` state either get
  an explicit Core destination, an ignored-legacy diagnostic, or a manual
  migration note. `gc doctor --fix` must not delete these state paths in this
  migration.

### Bootstrap Cleanup
<!-- REVIEW: added per blocker-bootstrap-fixture -->

`internal/bootstrap/bootstrap.go` should no longer depend on
`internal/bootstrap/packs/core`.

Use a single approach:

- Introduce a minimal synthetic fixture under
  `internal/bootstrap/testdata/packs/core`, used only by bootstrap tests.
- Make bootstrap tests inject a fixture filesystem explicitly; they must not
  read production Core assets from `internal/packs/core`.
- The fixture identity should remain `core` only if the test contract needs to
  prove legacy bootstrap metadata accepts a Core-shaped pack. Otherwise use
  `test-core` and update assertions to avoid implying it is production Core.
- The production `BootstrapPacks` default remains empty.
- Remove the production `//go:embed packs/**` dependency from
  `internal/bootstrap/bootstrap.go`. If a tiny compatibility embed remains, its
  path must name test/compatibility content and stay out of production
  bootstrap materialization.
- Add guards that fail on lingering imports of
  `internal/bootstrap/packs/core`, `AssetDir: "packs/core"` outside the new
  fixture tests, and file copies/hashes that still target the old production
  Core path.
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
| Dog prompt fragments | Core only for generic maintenance-agent behavior; Gastown notification/requester behavior moves to Gastown. | Behavior inventory rows and renamed-agent Core test. |
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
- Add behavior-preservation tests for all inventory entries.
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
   Gastown-owned assets, add the behavior inventory, update Gastown tests, and
   record the immutable commit/hash. Required gates: `gascity-packs` test suite,
   behavior inventory completeness, formula/order composition, script execution,
   and old-test to new-test mapping.
2. Gas City packcompat slice:
   add the Gas City pinned-public-pack compatibility test without deleting
   in-tree sources. Required gates:
   `go test ./test/packcompat -run TestPinnedPublicGastownBehavior`,
   `go test ./examples/...`, and `make test-fast-parallel`.
3. Core extraction slice:
   move Core assets to `internal/packs/core`, update hooks and builtin registry
   imports, and add bootstrap fixture isolation. Required gates:
   `go test ./internal/builtinpacks ./internal/hooks ./internal/bootstrap` and
   source guards for old bootstrap Core dependencies.
4. Core loading/doctor slice:
   add the post-load Core provenance assertion, production `config.Load*`
   scanner, Core doctor, and safe import-state doctor edits. Required gates:
   `go test ./cmd/gc -run 'BuiltinPack|ImportStateDoctor|CorePack|TryReloadConfig|MaterializeBuiltinPacks'`
   plus doctor golden/failure-atomic tests.
5. Maintenance folding slice:
   move Core-owned Maintenance assets into Core, move Gastown-owned Maintenance
   assets to public Gastown, and apply the runtime retirement table. Required
   gates: formula/order/script behavior tests, stale directory preservation
   tests, runtime-state diagnostics, and provider pack continuity tests.
6. Registry/cache slice:
   remove Maintenance and Gastown from the embedded pack registry, retire public
   synthetic aliases, update `PublicGastownPackVersion`, and verify stale cache
   rejection. Required gates: registry/cache negative tests and packcompat test
   against the exact public pin.
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
| new binary | old pack | Fresh Gastown init pins the new pack, but existing locks to old public commits continue to load or doctor reports an actionable version-skew diagnostic without mutating custom content. |
| new binary | new pack | Core, provider packs, and public Gastown load without Maintenance; packcompat and behavior inventory gates pass. |
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
  behavior inventory and Gastown pack tests are the mitigation.
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
