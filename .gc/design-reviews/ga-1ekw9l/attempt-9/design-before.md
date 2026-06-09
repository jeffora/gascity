---
plan_slug: core-gastown-pack-migration
phase: implementation-plan
rig: gascity
rig_root: /data/projects/gascity
artifact_root: /data/projects/gascity/plans
requirements_file: /data/projects/gascity/plans/core-gastown-pack-migration/requirements.md
status: draft
created_at: 2026-06-04T15:07:35Z
updated_at: 2026-06-09T13:20:59Z
---

# Implementation Plan: Core and Gastown Pack Split

## Summary

Move Gas City's required Core pack from `internal/bootstrap/packs/core` to
`internal/packs/core`, retire the standalone Maintenance system pack, and make
`gascity-packs/gastown` the only maintained Gastown behavior source. Core
remains a mandatory host system pack for normal `gc` runtime loading, while
Gastown remains an explicit public pack import pinned to immutable external
evidence. The migration is blocked from source deletion or Maintenance removal
until generated behavior evidence, public Gastown compatibility artifacts, and
exact-pack compatibility tests prove behavior preservation. This plan is ready
only for prerequisite-producing decomposition until the AC6, AC7, and AC14-AC17
support artifacts exist and pass; Gas City behavior-changing slices remain
blocked behind those gates. The implementation lands in small cross-repo slices
with fail-closed loader gates, failure-atomic doctor fixes, and
operator-visible diagnostics.

## Current System

`internal/builtinpacks/registry.go` currently registers embedded packs for
`core`, `bd`, `dolt`, `maintenance`, and `gastown`. The Core entry points at
`internal/bootstrap/packs/core`; Maintenance points at
`examples/gastown/packs/maintenance`; Gastown points at
`examples/gastown/packs/gastown`.

`cmd/gc/embed_builtin_packs.go` materializes embedded packs into
`.gc/system/packs/{name}`. `requiredBuiltinPackNames` requires Core and
Maintenance for normal cities, then adds `bd` or `dolt` according to the beads
provider. `builtinPackIncludes` appends those generated directories before
calling `config.LoadWithIncludes`, so Maintenance is still an implicit runtime
layer.

`internal/hooks/hooks.go` imports `internal/bootstrap/packs/core` directly and
reads provider overlays from `core.PackFS`. `internal/bootstrap/bootstrap.go`
still embeds `packs/**`; production `BootstrapPacks` is empty, but tests still
override it with `AssetDir: "packs/core"`, so Core extraction must also replace
test bootstrap fixtures.

`cmd/gc/import_state_doctor_check.go` rewrites legacy local Gastown imports to
`config.PublicGastownPackSource` and removes Maintenance imports with messaging
that says Core and Maintenance are supplied implicitly. That mutation path is
not yet protected by a single multi-file transaction coordinator.

Fresh Gastown initialization already writes explicit public
`gascity-packs/gastown` imports through
`internal/config.GastownCityWithProviders`, `PublicGastownPackSource`, and
`PublicGastownPackVersion`. The public pack no longer imports Maintenance, but
comments and patches still assume the host runtime supplies an implicit
Maintenance/Core utility layer.

The in-repo `examples/gastown` tree still depends on local pack sources:
`examples/gastown/pack.toml` and `examples/gastown/city.toml` import
`packs/gastown`, while `examples/gastown/packs/gastown/pack.toml` imports
`../maintenance`. `examples/gastown/packs/maintenance` still owns Dog,
cleanup orders, maintenance scripts, and role-specific notification behavior.

Tests and docs currently assert the legacy shape. Examples include
`internal/builtinpacks/registry_test.go`,
`cmd/gc/embed_builtin_packs_test.go`, `cmd/gc/controller_test.go`,
`cmd/gc/import_state_doctor_check_test.go`,
`examples/gastown/gastown_test.go`, `docs/reference/system-packs.md`,
`docs/guides/shareable-packs.md`, and
`docs/getting-started/troubleshooting.md`.

## Proposed Implementation

### Pack Ownership And Source Layout

Create `internal/packs/core` as the canonical Core package.
`internal/packs/core/embed.go` embeds Core assets with `PackFS`.
`internal/builtinpacks/registry.go`, `internal/hooks/hooks.go`, hook tests, and
bootstrap tests import this package instead of
`internal/bootstrap/packs/core`. After the replacement fixtures and import
guards are live, remove `internal/bootstrap/packs/core` as an asset source.

Core owns SDK-generic behavior only: CLI usage skills, generic provider hook
overlays, generic formulas and orders, prompt/template helpers that do not
depend on Gastown roles, and a configurable Core maintenance worker declared in
pack configuration. `dog` is allowed only as that default Core pack
configuration value; Go code must not require `dog` or any other configured
agent name for controller-owned SDK infrastructure.

Retire Maintenance as a standalone host system pack. Maintenance assets are
classified row-by-row into Core, public Gastown, provider pack, docs-only, or
approved retirement before they move. Do not create a replacement
`maintenance` system pack.

Make public `gascity-packs/gastown` the only maintained Gastown behavior
source. `examples/gastown` remains only as an example city that imports the
public pack. Gas City tests that previously validated local Gastown behavior
must either move to `gascity-packs` or become Gas City wiring tests that prove
public import, lock, and host-Core compatibility.

### External Public Gastown Prerequisite
<!-- REVIEW: added per behavior-evidence-public-pack-gate -->

No Gas City source deletion, Core role-generalization, Maintenance removal, or
public synthetic alias retirement may land before the matching public Gastown
work exists at immutable commits. The `gascity-packs` prerequisite must produce
all of the following before Gas City consumes the activation pin:

- moved Gastown-owned formulas, orders, scripts, prompts, overlays, and docs;
- tests replacing or mapping `examples/gastown/gastown_test.go` and
  `examples/gastown/maintenance_scripts_test.go`;
- `gastown/behavior-preservation.yaml`, generated from old/new source and test
  witnesses;
- `gastown/public-gastown-pins.yaml`, naming compatibility and activation
  commits separately;
- proof that public Gastown runs with host Core present and no Maintenance pack;
- resolved ownership rows for `mol-review-quorum`, provider overlays, Dog
  prompt fragments, review workflow checks, shutdown-dance examples,
  `mol-polecat-*`, branch pruning, and role-theme/tmux behavior.

The prerequisite is tracked as external work, not inferred from Gas City source.
Its acceptance artifacts are:

- `gascity-packs/gastown/public-gastown-pins.yaml`, with `compatibility` and
  `activation` entries containing immutable commit, source URL, subpath, pack
  digest, behavior-manifest digest, generated-at timestamp, and approving PR;
- `gascity-packs/gastown/behavior-preservation.yaml`, generated by the public
  pack migration task from old Gas City sources and new public-pack witnesses;
- `gascity-packs/gastown/ownership.yaml`, assigning every behavior-bearing
  formula, order, prompt, script, overlay, docs page, and role-themed helper to
  Core, public Gastown, provider pack, docs-only, or retired;
- packcompat transcripts under
  `gascity-packs/gastown/artifacts/packcompat/` for compatibility-pin mode and
  host-Core/no-Maintenance activation mode.

Gas City adoption of `internal/config/PublicGastownPackVersion` has two
meanings. The compatibility pin proves the new public pack can coexist with the
current loader while in-tree sources still exist. The activation pin is consumed
only in the same candidate branch that removes Maintenance from required host
packs and proves no-Maintenance production loading.

### Behavior Evidence Contract
<!-- REVIEW: added per behavior-evidence-contract -->

Add a generated Behavior Evidence manifest and packcompat gate before any
dependent move. The first Gas City evidence slice creates
`internal/packevidence`, `cmd/gc/pack_evidence.go`, and schema fixtures under
`testdata/packevidence/`. Planning-only generated evidence may live under
`plans/core-gastown-pack-migration/artifacts/`, but decomposition must move the
generator, schema, sample rows, and freshness tests into checked-in
implementation paths before source deletion can be scheduled.

The generator command is executable and freshness-checked:

```bash
go run ./cmd/gc pack-evidence generate \
  --old-baseline <immutable-gas-city-commit> \
  --public-gastown <gascity-packs-checkout-or-cache> \
  --out plans/core-gastown-pack-migration/support/behavior-preservation-manifest.yaml
go run ./cmd/gc pack-evidence validate \
  --manifest plans/core-gastown-pack-migration/support/behavior-preservation-manifest.yaml
```

Each manifest row must include:

- stable row id;
- old owner, path, asset kind, and helper dependencies;
- source-kind enum: Go behavior, TOML formula/order, prompt/template, shell
  script, docs/operator text, generated reference, runtime-state behavior,
  test/golden, provider overlay, public-pack asset, or historical fixture;
- trigger, requester, detector, route metadata, mail/nudge target, prompt
  fragment, script branch, runtime-state path, or named-session behavior;
- old witness: source assertion, test, fixture, golden output, or command
  transcript;
- witness-kind enum: unit test, process/integration test, golden transcript,
  static scanner, packcompat assertion, public-pack test, manual proof
  transcript, or approved removal record;
- new owner: Core, public Gastown, provider pack, docs-only, or approved
  retirement;
- new path and new witness;
- immutable public Gastown commit for public-pack rows;
- consuming `internal/config/PublicGastownPackVersion` value;
- semantic-equivalence assertion, or approved delta/removal record with owner,
  reason, replacement, and operator impact.

Evidence classes are `preserved`, `moved`, `generalized`, `split`,
`retired-approved`, and `external-prerequisite`. A moved, split, generalized,
or deleted source row must have one old witness and one new executable witness.
Rows whose new witness lives in public Gastown cannot unblock Gas City deletion
until the public commit, pack digest, behavior-manifest digest, and packcompat
transcript are cited in the pin ledger.

The generator walks old Gas City behavior-bearing sources under Core,
Maintenance, Gastown examples, hook overlays, formulas, orders, prompts,
skills, shell scripts, doctor strings, route metadata, notification templates,
runtime-state helpers, tests, and helper references. CI fails if a moved,
split, generalized, deleted, or helper-dependent asset lacks a row, if a row
lacks old and new witnesses, or if a semantic delta lacks an approved record.
It also compares against a Git historical baseline so current-workspace-only
scans cannot miss deleted or moved legacy behavior.

`test/packcompat` consumes the manifest and the exact public pin. It installs
public Gastown through the ordinary remote-pack path or a validated ordinary
remote cache, composes moved formulas and orders, resolves moved scripts using
pack-relative paths, verifies hook overlays and configured agents, and checks
one assertion per manifest row.

Packcompat has two named modes. Compatibility-pin mode runs while local
Maintenance and in-tree Gastown sources still exist and proves the public pack
does not depend on hidden fallback. Activation-pin mode runs with Maintenance
removed from required host packs and fails if any assertion resolves from
`examples/gastown`, `.gc/system/packs/gastown`, `.gc/system/packs/maintenance`,
or a synthetic public alias.

### Required System Pack Loader
<!-- REVIEW: added per required-core-loader-contract -->

Create `internal/systempacks` as the single production boundary for required
host packs. It owns:

- `RequiredPackNames(cityPath string)`;
- `MaterializeRequiredPacks(ctx, cityPath, provider)`;
- `ValidateRequiredFileSets(ctx, cityPath)`;
- `RuntimeIncludes(ctx, cityPath, provider)`;
- `LoadRuntimeCity(ctx, cityPath, opts)`;
- `LoadRuntimeCityNoRefresh(ctx, cityPath, opts)`.

The trusted descriptor flow is part of the API, not an implementation detail.
`internal/systempacks` builds a `RequiredDescriptor` for each required host
pack with pack id, source kind, selected provider, materialization root,
expected content-manifest digest, expected file-set digest, required
`pack.toml` digest, and descriptor id. It passes those descriptors to
`internal/config` through a typed option such as
`config.WithRequiredSystemPacks`. Config resolution returns
`RequiredSystemPackParticipation` records keyed by the same descriptor id,
including the resolved layer id, import edge, effective contribution digest,
collision result, and allowed-use mode. <!-- REVIEW: added per required-pack participation and loader authority -->

Normal command paths and behavior-driving production `internal/` paths call
`LoadRuntimeCity` or `LoadRuntimeCityNoRefresh`; they do not hand-build
required-pack include lists or call `config.Load*` directly. Low-level
`internal/config` tests may still call lower-level loaders because they test
config behavior, not full `gc` runtime behavior.

`LoadRuntimeCity` returns a typed runtime result, not just a config pointer:
`Config`, `Includes`, `RequiredParticipation`, `Diagnostics`, `Mode`, and
`Freshness`. `Mode` is `ready`, `read_only_degraded`, or `blocked`. Only
`ready` may drive dispatch, formula expansion, order evaluation, hook
rendering, prompt resolution, or session start. API/controller callers must
check this mode through the systempacks result and may not reinterpret
diagnostics locally.

Runtime loading runs two fatal gates. Before config resolution it materializes
and validates the required file set for Core plus provider packs (`bd` and
`dolt` as selected today). After config resolution it validates typed
`RequiredSystemPackParticipation` records proving every required pack
participated in the resolved config. A generated or repaired Core tree without
typed participation is a load failure.

Gate 1 and Gate 2 are bound by descriptor id. Gate 1 proves that the exact
materialized directory for each descriptor matches the embedded manifest and
provider selection before any `pack.toml` is trusted. Gate 2 proves that the
same descriptor, not a same-named user import or copied Core tree, contributed
to the final config layer graph. A path match without descriptor participation,
or participation without a matching validated file set, is `blocked`.

No-refresh reloads do not repair. If Core is missing, corrupt, stale, shadowed,
or missing participation, the controller keeps the last-known-good runtime
config only for read-only status/reporting, publishes an event and diagnostic,
and pauses or refuses behavior-changing dispatch, formula, order, hook, prompt,
and agent-start operations until a successful refreshed load occurs. API and
CLI paths surface the same diagnostic instead of silently continuing with
current invalid config.

Behavior-changing entry points receive a `RuntimeSnapshot` or `RuntimeGuard`
from `internal/systempacks` and call `RequireReady(op)` before dispatch,
formula expansion, order evaluation, hook rendering, prompt resolution,
session start, worker creation, API mutation, or controller scheduling.
Read-only status/reporting entry points may use `read_only_degraded` snapshots
but must surface the diagnostics unchanged.

Add scanner tests modeled on `cmd/gc/worker_boundary_import_test.go` to reject
production direct calls to `config.Load`, `config.LoadCity`,
`config.LoadWithIncludes`, aliases, wrappers, method/function values, and
manual required-pack include assembly in `cmd/gc` and behavior-driving
`internal/` packages. Any exception is generated into a partial-read allowlist
with file, function, call kind, fields consumed, reason, and a focused test
proving it cannot drive normal runtime behavior.

The bypass inventory lives in a generated allowlist such as
`internal/systempacks/testdata/config_loader_allowlist.yaml`. Its validator is
AST/type-aware: aliases, selector values, function variables, wrapper
functions, and package-local helper names count as bypasses unless the
allowlist row proves the call is a partial read that cannot start sessions,
expand formulas, render hooks/prompts, dispatch work, evaluate orders, or write
city state.

Provider-conditioned required-pack selection lives only in `internal/systempacks`.
Callers pass city path and selected beads provider; they do not branch on Core,
`bd`, or `dolt` include behavior outside that boundary.

Keep existing bootstrap collision protection for `core`, `maintenance`, and
`gastown` until replacement required-system-pack collision gates and retired
source classification are live. Remove legacy entries only in the slice that
proves the replacement gates.

### Pack Registry, Cache, And Retired Source Authority
<!-- REVIEW: added per pack-boundary-retired-source-contract -->

Update `internal/builtinpacks/registry.go` so `All()` returns only Core, `bd`,
and `dolt`. Core's source identity becomes `internal/packs/core`. Maintenance
and Gastown are removed from the embedded set after the activation gate passes.

Create `internal/packsource` as the sole authority for retired Maintenance and
Gastown source classification. All load, install, cache, lockfile,
materialization, discovery, doctor, docs-lint, generated-reference-lint, and
public-source normalization paths use this classifier instead of duplicating
string checks. The classifier returns typed states such as active bundled,
active public, retired generated/example, retired custom/fork, stale cache,
historical fixture, and invalid collision.

Public Gastown never resolves from bundled synthetic cache aliases after this
migration. `RepoCacheKey` includes normalized source, exact commit, and subpath.
Promotion and read hits verify source, commit, subpath, pack digest, and
manifest digest. Durable public refs may keep SHAs fetchable, but they never
replace immutable SHA plus digest validation. Stale synthetic Gastown or
Maintenance cache entries may remain on disk, but new lock generation and
runtime resolution must not select them.

The subpath-aware lock/cache proof lands before the first Gas City slice that
updates `internal/config/PublicGastownPackVersion`. Lock entries store source,
commit, subpath, canonical pack-tree `sha256`, behavior-manifest `sha256`,
cache entry id, and validation timestamp. Old lock or cache records that lack
subpath or digest fields are read only far enough to emit a version-skew or
repair diagnostic; they cannot satisfy a new `sha:` pin. The pin-coherence gate
compares `PublicGastownPackVersion`, `public-gastown-pins.yaml`,
fresh-init output, lockfile provenance, cache proof, pack digest, and
behavior-manifest digest in one command before the pin is consumed.
<!-- REVIEW: added per public-pack pin/cache sequencing -->

Zero-duplicate-active and zero-merge gates compare active bundled, active
public, stale generated, synthetic cache, ordinary remote cache, compatibility
pin, activation pin, and old/new binary views. If the same behavior id is active
from more than one source, or if Core and public Gastown both claim an
unresolved behavior row, runtime loading fails before any behavior executes.

Do not delete stale `.gc/system/packs/maintenance`,
`.gc/system/packs/gastown`, or `.gc/runtime/packs/maintenance` directories
during startup or `gc doctor --fix`. They are ignored by active discovery and
reported as legacy state because they may contain operator edits.

### Doctor And Runtime-State Mutation Safety
<!-- REVIEW: added per doctor-runtime-state-safety -->

Replace legacy direct `Check.Fix(ctx)` mutation with a `FixIntent` plus mutation
coordinator API before public-pin, import-rewrite, or runtime-state fixes are
enabled. Create `internal/doctorfix` with `FixIntent`, `Plan`, `Stage`,
`Publish`, `Recover`, and `Refuse` operations. Existing doctor checks either
return a structured intent through that API or are marked report-only with
manual guidance. The coordinator is the only path that writes city manifests,
lockfiles, installed pack directories, runtime-state migrations, or import
rewrites.

The first doctor slice generates
`plans/core-gastown-pack-migration/support/doctor-fix-inventory.yaml` from
every `Fix(ctx)`, import rewrite, protected write, pack install/update,
required-pack repair/quarantine, runtime-state write, cleanup path, and
test-only mutation helper. Each row is assigned to `FixIntent`, `report-only`,
`remove`, or `allowed-test-helper`; production rows without an assignment fail
CI. Existing direct mutation paths are deleted or wrapped before any new public
pin, import rewrite, or runtime-state migration can mutate city state.

The coordinator acquires a crash-released city advisory lock before digest
preflight. If a report-only phase reads before the lock for operator messaging,
it repeats digest and provenance validation after the lock and before staging.
If a controller for the same city is running, discovered from live runtime state
rather than status files, automatic fix refuses with manual guidance.

Multi-file fixes write durable recovery state before the first publish step,
stage all edits, re-read target digests before each temp-file rename, and define
a single commit point. A process death before commit reruns deterministically or
rolls back from recovery state. A process death after commit converges by
revalidating Core participation, public-pin installability, lock contents, and
runtime-state marker state.

Runtime-state migration is doctor-owned. Controller startup, API handlers, and
runtime reload paths may detect legacy state, emit diagnostics, and refuse
behavior-changing operations, but they do not mutate runtime state. `gc doctor
--fix` performs the migration only when no controller for the city is running,
then holds the same city advisory lock used by pack install/update and import
rewrites.

Runtime-state migration moves JSONL archive state, spawn-storm ledgers,
refs/remotes, escalation fields, pending archive push state, explicit formula
environment state, and order skip/tracking compatibility aliases under
Core-owned paths only after shared lock acquisition and staged archive-copy
digest checks. The migration marker records schema version, old path, new path,
staged archive digest, completed steps, retained legacy path, fallback policy,
and old-binary post-marker write detection. If an old binary writes after the
marker, the new binary reports a version-skew diagnostic and requires manual
reconciliation or a deterministic re-upgrade flow before behavior resumes.

### Role Neutrality And Configurable Bindings
<!-- REVIEW: added per role-neutrality-contract -->

Generate a role-surface manifest covering Go, TOML, shell, Markdown,
templates, generated command text, API classifications, dashboard/OpenAPI
generated references, tmux theme helpers, default scaffolding, warmup mail
defaults, prompt fallbacks, formulas, overlays, metadata, tests, docs, and
public Gastown companion files.

Active behavior enumeration is also part of the role-neutral contract. Loaders,
installers, cache readers, docs/reference generators, prompt scanners, formula
and order expanders, script resolvers, hook overlay readers, and doctor checks
must obtain roots through `internal/packsource.ActiveRootsFor(kind)` or a typed
equivalent. Scanner tests reject raw `fs.ReadDir`, `fs.WalkDir`,
`filepath.Walk`, `Glob`, or string-prefix enumeration over pack roots unless an
allowlist row proves the path is a historical fixture or non-behavior audit.

Core-owned behavior may not contain or branch on `mayor`, `deacon`, `witness`,
`refinery`, `polecat`, `boot`, `crew`, or `gastown` except through explicit
historical/test allowlist rows with owner, justification, expiry, and negative
fixtures. `dog` is allowed only as Core's default configurable maintenance
worker in pack configuration and tests for that configuration.

Add parser/resolver support for `[gc.bindings.*]`,
`[system_packs.*.bindings]`, `target_binding`, `gc.run_target_binding`, and
`GC_CORE_MAINTENANCE_WORKER`. Core maintenance formulas and orders resolve the
configured binding rather than a Go constant. Tests prove Core-only cities load
and controller-owned SDK operations still work when the maintenance worker is
renamed or omitted.

Binding precedence is explicit: city `[gc.bindings.*]` overrides
`[system_packs.<pack>.bindings]`; formula `target_binding` and step
`gc.run_target_binding` resolve through that merged binding table; environment
injection may supply a default only when neither city nor pack config names a
binding. Missing optional bindings skip user-agent work with a typed diagnostic.
Missing required provider-pack escalation bindings fail the formula/order before
dispatch. No Go fallback may substitute `mayor`, `deacon`, `dog`, or another
concrete role name.

Move or rewrite active role-specific surfaces. Branch pruning and
`mol-polecat-*` move to public Gastown. `mol-shutdown-dance` may keep generic
stuck-session due process in Core, but Gastown detector/requester examples move
to Gastown. `mol-review-quorum`, provider overlays, Dog prompt fragments,
review checks, and tmux theme APIs are assigned by manifest rows before source
moves.
Formula-name heuristics such as `mol-polecat-*` and `mol-refinery-patrol` are
replaced by declared formula metadata owned by the pack that provides the
formula. Tmux cleanup examples and tests must target isolated sockets only and
must never use a default-server kill.

### Bootstrap Fixture Isolation
<!-- REVIEW: added per bootstrap-fixture-isolation -->

Audit every `GC_BOOTSTRAP` production and test dependency, including doctor
checks, command tests, testscript defaults, prompt tests, config bundled-import
tests, precompact hook tests, packlint, generated docs, hook references, and
helper paths.

The buildable end state is explicit: no production package may import
`internal/bootstrap/packs/core`; no production `//go:embed packs/**` may carry
Core assets; `bootstrapAssets`, `embeddedBootstrapPacks`, and `BootstrapPacks`
are deleted or backed only by `bootstrap.EmptyFS`; and package deletion of
`internal/bootstrap/packs/core` is the compile-time proof that old imports are
gone. Path-literal scanners cover `packs/core`, `AssetDir: "packs/core"`,
`Subpath` strings, constructed paths, docs, examples, fixtures, and generated
files. <!-- REVIEW: added per bootstrap fixture isolation -->

Production bootstrap no longer embeds Core assets. Tests that need bootstrap
assets use an empty `fs.FS` fixture or minimal inline fixture whose `Stat`,
`WalkDir`, and `ReadFile` behavior is asserted. Fixture guard tests fail if
allowed test paths copy production-only Core directories such as `formulas/`,
`orders/`, `overlay/`, `skills/`, or `assets/prompts/`.

Define a non-nil production `bootstrap.EmptyFS` implementation whose `Open(".")`
returns an empty directory and whose `Open` for all other names returns
`fs.ErrNotExist`. Tests assert `fs.Stat`, `fs.WalkDir`, and `fs.ReadFile`
against that implementation. `ensureBootstrapForDoctor` is either deleted in
the same slice or rewritten to call `internal/systempacks` diagnostics without
materializing bootstrap Core.

`GC_BOOTSTRAP=skip`, if retained for tests, may skip only empty bootstrap
fixture materialization. It must not skip `internal/systempacks`
materialization, strict required Core file-set integrity, retired-source
classification, collision checks, typed participation validation, provider
materialization, or doctor cleanup.

### Operator Docs And Generated References
<!-- REVIEW: added per docs-operator-release-gate -->

Add a generated wording/docs scanner covering Markdown, MDX, JSON, TXT, TS,
OpenAPI, dashboard generated files, docs/schema outputs, public Gastown docs,
generated help, CLI examples, scripts, prompts, tutorial transcripts, and
doctor output. Allowlist rows carry owner, justification, expiry, and negative
fixtures.

The vocabulary authority is
`plans/core-gastown-pack-migration/support/terminology-matrix.yaml`; generated
operator-facing evidence is written under
`plans/core-gastown-pack-migration/support/generated-references/`. If
`docs/reference/system-packs.md` does not exist in the implementation branch,
the slice creates it before docs lint runs instead of treating the path as an
optional note.

The wording scanner consumes a terminology matrix with token, class,
allowed-contexts, denied-contexts, owner, examples, false-positive rule, and
golden fixture. It distinguishes retired standalone Maintenance-pack wording
from valid lowercase maintenance, Dolt/store-maintenance terminology,
Core-maintenance-worker bindings, public Gastown docs, historical examples, and
operator recovery text. Generated OpenAPI, dashboard types, docs/schema files,
CLI reference/help, tutorial transcripts, and doctor output goldens regenerate
before wording lint runs.

Update `docs/reference/system-packs.md`, `docs/guides/shareable-packs.md`,
`docs/getting-started/troubleshooting.md`,
`docs/tutorials/01-cities-and-rigs.md`,
`docs/tutorials/05-formulas.md`, `docs/tutorials/07-orders.md`, CLI help,
doctor strings, generated references, examples, and script comments in the same
slice as the behavior change they describe. Canonical wording is: Core is the
required host system pack; `bd` and `dolt` are provider-dependent host system
packs; Maintenance is retired as a standalone pack; Gastown is an explicit
public pack import; stale Maintenance/Gastown generated paths are ignored
legacy state and are not automatically deleted.

## Data And State

Required system-pack state lives under `.gc/system/packs/core` and provider
pack directories such as `.gc/system/packs/bd` and `.gc/system/packs/dolt`.
Required-pack repair regenerates missing or corrupt expected files from the
embedded manifest, prunes generated unexpected effective files, and quarantines
operator-edited or unclassifiable files outside active discovery before any
formula, order, script, prompt, hook, or overlay can be read.

`RequiredSystemPackParticipation` is the typed runtime proof record. Each
record contains pack id, materialized directory, embedded source id, content
manifest digest, validated file-set digest, `pack.toml` digest, resolved config
layer id, import edge proving participation in final config resolution, repair
status, freshness timestamp, diagnostic id, and allowed-use mode.

Behavior evidence is durable generated data. Gas City planning artifacts live
under `plans/core-gastown-pack-migration/artifacts/` until implementation
chooses checked-in generator and schema paths. Public Gastown produces
`gastown/behavior-preservation.yaml` and `gastown/public-gastown-pins.yaml`.
CI treats those artifacts as stale if their source digests, old/new witness
digests, or public pin values do not match the current tree.

Acceptance support artifacts live under
`plans/core-gastown-pack-migration/support/` until a later slice promotes any
validator or schema into package-local `testdata`. The required support set is
`pack-resolution-matrix.yaml`, `asset-migration-ledger.yaml`,
`behavior-preservation-manifest.yaml`, `source-consumer-closure.yaml`,
`role-neutrality-scan.yaml`, `doctor-fix-inventory.yaml`,
`migration-diagnostics.schema.json`, `docs-authority-audit.yaml`,
`coverage-transfer.yaml`, `public-gastown-pin-ledger.yaml`,
`version-skew-matrix.yaml`, and `acceptance-proof-matrix.yaml`.

Public pack cache state uses ordinary remote-pack cache entries keyed by
normalized source, commit, and subpath. Cache entries store source provenance,
pack digest, manifest digest, and validation timestamp. Historical synthetic
Gastown and Maintenance aliases are non-authoritative retired state and cannot
satisfy a public `sha:` pin.

Doctor recovery state records mutation intent id, locked city, preflight file
digests, staged paths, publish order, commit point, completed steps, rollback
instructions, and final validation result. It is written before mutation and
removed or marked complete only after all post-commit validation passes.

Runtime-state migration records marker path, marker schema version, old paths,
new Core-owned paths, staged archive-copy digest, push-cursor reconciliation,
order skip/tracking aliases, post-marker old-binary write detection, rollback
state, and re-upgrade state. Retained legacy state is ignored unless the marker
or digest checks show conflict.

The migration table is part of the implementation, not review prose:

| Artifact | Old path or key | New Core-owned path or key | Merge rule |
| --- | --- | --- | --- |
| JSONL archives | legacy archive directories | Core archive directory | staged copy with source and destination digests |
| Git refs/remotes | legacy `.git` refs/remotes | retained Git state plus Core marker | reconcile, never delete unknown refs |
| Archive push cursors | pending push state | Core push cursor state | newest verified cursor wins, conflicts block |
| Escalation fields | legacy runtime metadata | Core runtime metadata | copy exact fields with schema version |
| Spawn-storm ledgers | old throttle ledger | Core throttle ledger | read-union before marker, Core-only after marker |
| Order skip/tracking | old order keys | stable order keys plus aliases | aliases suppress same logical order until retired |
| Formula environment | legacy env state | Core formula env state | copy exact key/value records, conflicts block |

Role-surface and docs wording manifests are generated data with allowlist rows
for historical/test exceptions. Allowlist rows require owner, reason, expiry,
source path, token kind, and a negative fixture proving new active Core
behavior cannot use the exception.

Stale `.gc/system/packs/maintenance`, `.gc/system/packs/gastown`, and
`.gc/runtime/packs/maintenance` directories are legacy operator state. Startup
and doctor do not delete them automatically. Doctor may report them as ignored
legacy paths with manual cleanup guidance.

## Testing

Focused Gas City unit tests:

- `go test ./internal/builtinpacks -run 'All|Source|Synthetic|Retired'`
  verifies only Core, `bd`, and `dolt` are embedded; Core source is
  `internal/packs/core`; retired Maintenance and Gastown sources cannot create
  new locks or synthetic cache entries.
- `go test ./internal/hooks ./internal/bootstrap -run 'Core|Bootstrap|Hook'`
  proves hooks read overlays from `internal/packs/core`, production bootstrap
  no longer owns Core assets, and bootstrap fixtures are minimal.
- `go test ./cmd/gc -run 'BuiltinPack|ImportStateDoctor|CorePack|DoctorFix|TryReloadConfig'`
  proves include lists, doctor diagnostics, mutation coordinator behavior, and
  no-refresh reload semantics.
- `go test ./internal/systempacks -run 'MaterializeRequiredPacks|RequiredSystemPackParticipation|LoadRuntimeCity'`
  proves strict pre-resolution validation, post-resolution participation, typed
  records, repair/quarantine behavior, and fail-closed missing/invalid Core
  cases.
- `go test ./internal/doctorfix ./cmd/gc -run 'FixIntent|Coordinator|ImportState|RuntimeStateMigration'`
  proves mutation coordination, report-only legacy fix handling, recovery state,
  lock checks, and controller-running refusal.
- Scanner tests reject production direct `config.Load*` bypasses, active Core
  role-name references, retired-source string classifiers, bare tmux cleanup,
  old bootstrap import paths, and copied production Core fixture directories
  outside generated allowlists.
- `go test ./internal/systempacks -run 'RequiredDescriptor|GateBinding|RuntimeGuard|BypassAllowlist'`
  proves descriptor-id binding, copied-Core rejection, user-imported `core`
  collision rejection, stale/corrupt materialization blocking, provider-pack
  selection confinement, and API/controller behavior-changing guard coverage.
- `go test ./internal/packsource ./test/packlint/... -run 'ActiveRoots|Enumerator|RetiredSource|RoleSurface'`
  proves active enumerators consume classifier-filtered roots and raw pack-root
  walks are rejected outside generated allowlists.
- `go test ./internal/packevidence ./test/packcompat -run 'BehaviorEvidence|PinnedPublicGastown|PinCoherence'`
  proves manifest freshness, one-row-per-trigger extraction, exact public pin
  provenance, synthetic-alias rejection, and compatibility/activation modes.

Behavior and compatibility tests:

- `go test ./test/packcompat -run TestPinnedPublicGastownBehavior` runs first
  in compatibility-pin mode, then in no-Maintenance production-loader mode in
  the activation slice.
- `go test ./test/packlint/...` validates role-surface, retired-source,
  wording/docs, generated-reference, and old bootstrap path inventories.
- Representative Core formula/order tests compose molecules, assert resolved
  step content, verify run-target bindings and configured recipients, and
  execute moved scripts from pack-relative paths.
- Binding tests cover city override, system-pack default, formula
  `target_binding`, step `gc.run_target_binding`, environment fallback,
  required-recipient failure, and omitted optional binding diagnostics.
- Public Gastown tests in `gascity-packs` cover moved formulas, moved orders,
  branch pruning, Polecat formulas, Gastown detector/requester behavior,
  prompt-template resolution, host-Core compatibility, and no-Maintenance
  loading.

Doctor and runtime-state tests:

- Healthy-city `gc doctor --fix` is byte-identical and idempotent.
- Scoped TOML edits preserve comments, unknown tables, unknown fields, array
  order, formatting, and unrelated lock entries; otherwise automatic fix
  refuses.
- Failure injection after each staged publish step reruns or rolls back
  deterministically.
- Runtime-state migration detects old-binary post-marker writes, reconciles
  push cursors, preserves retained legacy state, and supports downgrade or
  manual recovery guidance.
- Offline cache tests cover exact-public-pin hit, digest mismatch, missing
  subpath, stale synthetic alias rejection, ordinary remote cache promotion, and
  fail-closed cache miss when network access is unavailable.

Docs and release tests:

- Golden tests cover doctor output, CLI help, first-run text, tutorial
  transcripts, and docs wording for minimal and Gastown-template cities.
- Generated freshness tests run OpenAPI, dashboard type, docs/schema, CLI help,
  tutorial transcript, doctor output, wording matrix, role-surface, and
  behavior-evidence generation before linting their outputs.
- `make dashboard-check` runs only if API, dashboard, generated OpenAPI, or
  generated TypeScript files change.
- Broad gates after each code slice: `make test-fast-parallel` and
  `go vet ./...`. High-risk loader/doctor/runtime-state slices also run the
  sharded process and integration targets from `TESTING.md`, including
  `make test-cmd-gc-process-parallel` and
  `make test-integration-shards-parallel`.

## Rollout And Recovery

### Decomposition Readiness Gate
<!-- REVIEW: added per rollout-decomposition-gates -->

Before full Gas City implementation decomposition, only external-prerequisite
and proof-producing beads may be created: public `gascity-packs` ownership and
pin work, support-artifact generators and validators, non-mutating inventory
scans, packcompat harness work, and docs/schema golden generation. Gas City
source deletion, Maintenance removal, public activation-pin consumption,
runtime-state mutation, automatic doctor repair, or behavior-changing loader
cutover waits until AC6, AC7, AC14, AC15, AC16, and AC17 support artifacts
exist, validate, and are cited by immutable commit or checked path.

AC17 acceptance-to-proof matrix rows are required before implementation
approval:

| AC | Proof artifact or command | Gate placement | First dependent slice |
| --- | --- | --- | --- |
| AC1 | Requirements schema review against `requirements.schema.md` | design gate | already complete before this plan |
| AC2 | `go test ./internal/systempacks ./cmd/gc -run 'RequiredCore|ControllerOnly|FreshCity'` | local CI | Slice 4a |
| AC3 | `support/pack-resolution-matrix.yaml` plus validator and collision tests | local CI | Slice 4a |
| AC4 | `go test ./cmd/gc ./test/packcompat -run 'GastownInitPublicPin|FreshGastown'` | deterministic public-pack CI | Slice 2 |
| AC5 | `support/source-consumer-closure.yaml` plus retired Maintenance registry/cache tests | local CI | Slice 5b |
| AC6 | `support/asset-migration-ledger.yaml` with source snapshot and split-row validator | manual audit plus local CI | Slice 1a |
| AC7 | `support/behavior-preservation-manifest.yaml` plus packcompat witness checks | deterministic public-pack CI | Slice 1b |
| AC8 | `support/role-neutrality-scan.yaml` and generated/materialized absence scans | local CI | Slice 3 |
| AC9 | Binding tests for default, renamed, omitted, disabled, and no-executor cities | local and process CI | Slice 4a |
| AC10 | Doctor upgrade matrix, repair idempotence, TOML preservation, and live-controller refusal tests | process/integration CI | Slice 4b |
| AC11 | `support/migration-diagnostics.schema.json` plus text/JSON golden diagnostics | local CI | Slice 4a |
| AC12 | `support/docs-authority-audit.yaml`, terminology matrix, and docs/help goldens | docs CI | first operator-facing slice |
| AC13 | `support/coverage-transfer.yaml` mapping retired tests to replacement coverage | local CI | Slice 5b |
| AC14 | Public Gastown checkout or pinned-cache validation transcript with in-tree fallback disabled | release gate | Slice 1b |
| AC15 | `support/public-gastown-pin-ledger.yaml`, `support/version-skew-matrix.yaml`, and pin-coherence gate | release gate | Slice 2 and Slice 5a |
| AC16 | Network-disabled cache hit/miss, digest mismatch, stale alias, missing subpath, and concurrent promotion tests | process/integration CI | Slice 2 |
| AC17 | `support/acceptance-proof-matrix.yaml` validating AC1-AC16 evidence availability | design gate before bead creation | all implementation slices |

Slice 1a is the public Gastown ownership prerequisite. Land ownership rows for
every Gastown, Maintenance, provider-overlay, Dog prompt, review-check,
shutdown-dance example, branch-pruning, role-theme, and tmux behavior surface.
Rollback is to leave Gas City unchanged and not consume any new pin.

Slice 1b is public Gastown compatibility proof. Land moved Gastown-owned assets,
replacement tests, `behavior-preservation.yaml`, and the compatibility pin plus
packcompat transcript proving coexistence with the current Gas City loader.
Rollback is to keep Gas City on the previous pin and leave local sources active.

Slice 1c is public Gastown activation proof. Land the activation pin and
host-Core/no-Maintenance transcript. Gas City source deletion and Maintenance
removal wait until this prerequisite exists. Rollback is to leave Gas City
unchanged and not consume the activation pin.

Slice 2 is Gas City compatibility-pin adoption. Update
`internal/config/PublicGastownPackVersion` to the public compatibility commit,
add packcompat in current-loader mode, reject retired synthetic cache hits, and
rewire `examples/gastown` to public imports while in-tree sources still exist.
Unsafe legacy import rewrites are disabled or routed through the mutation
coordinator before this slice exposes new diagnostics. Rollback is to restore
the previous pin and keep in-tree sources active.

Slice 3 extracts Core. Move assets to `internal/packs/core`, update hooks and
registry imports, replace bootstrap fixtures, add fixture guards, and keep
Maintenance required until the activation gate. Rollback is to restore old
imports and embedded Core path before deleting the old source tree.

Slice 4a adds `internal/systempacks`, strict required-pack validation, typed
participation, production loader scanner, partial-read allowlist, no-refresh
read-only degraded mode, and pre-resolution diagnostics. Rollback is to route
runtime loads back through the old wrapper before any mutation state commits.

Slice 4b adds `internal/doctorfix`, migrates direct `Check.Fix(ctx)` callers to
`FixIntent` or report-only guidance, and proves lock/recovery behavior. Rollback
is to disable automatic fixes and keep diagnostics report-only.

Slice 4c adds doctor-owned runtime-state migration, version-skew diagnostics,
marker schema, state table fixtures, and old-binary post-marker detection.
Rollback is to rerun the coordinator recovery path or follow manual
reconciliation guidance before restoring old behavior.

Slice 5a consumes the public activation pin in a candidate branch and runs
packcompat in no-Maintenance production-loader mode while local sources are
still present. If the activation commit cannot satisfy old/new binary
compatibility and duplicate-definition requirements, stop and convert the slice
into a paired cross-repo activation boundary.

Slice 5b folds Maintenance only after Slice 5a passes. Remove Maintenance from
`requiredBuiltinPackNames`, move Core-owned Maintenance assets into Core,
consume Gastown-owned assets from the public pack, and keep stale generated
paths as ignored legacy state. Rollback before source deletion is to restore
the compatibility pin and re-enable Maintenance; after activation, recovery is
manual or coordinator-driven and must not silently reselect retired embedded
behavior.

Slice 6 cleans registry and cache. Remove Maintenance and Gastown from the
embedded registry, retire public synthetic aliases, enforce subpath-aware
ordinary cache keys, and verify stale-cache rejection. Rollback is to restore
the alias only if activation has not shipped; after activation, recovery is a
doctor diagnostic plus manual cache cleanup guidance, not silent fallback to
embedded Gastown.

Slice 7 deletes stale sources and lands final docs. Remove
`examples/gastown/packs/*` only after replacement tests are green, complete
docs/tutorial/doctor-output goldens, and run generated wording/reference
freshness gates. Rollback is to restore the source tree from git and revert
docs in the same change.

Slice-to-gate table:

| Slice | May start when | Must pass before merge | One-way boundary |
| --- | --- | --- | --- |
| 1a public ownership | AC6 ledger generator exists | ownership rows, source snapshot, split-row validation | none in Gas City |
| 1b compatibility proof | 1a rows are closed | AC7, AC14 compatibility transcript, fallback-disabled proof | no Gas City pin change |
| 1c activation proof | 1b compatibility pin is immutable | host-Core/no-Maintenance transcript and activation pin row | no Gas City source deletion |
| 2 compatibility-pin adoption | AC15/AC16 cache schema proof exists | pin-coherence, stale-alias rejection, offline cache matrix | first public pin consumed |
| 3 Core extraction | AC8 role scan and bootstrap inventory exist | old import/path scanner, fixture guard, hook overlay proof | old Core package may be deleted only after scanner passes |
| 4a systempacks loader | AC3 matrix and AC11 diagnostics schema exist | descriptor gate tests, runtime guard, bypass scanner | behavior-changing loads route through `internal/systempacks` |
| 4b doctorfix | AC10 inventory exists | lock/journal/failure-injection tests and report-only conversion | automatic fixes use only `internal/doctorfix` |
| 4c runtime-state migration | 4b coordinator is active | marker, quiesce, old-binary, downgrade, and re-upgrade tests | marker schema published |
| 5a activation candidate | 1c activation proof and 4a loader pass | no-Maintenance packcompat, duplicate-active gate, old/new binary matrix | activation pin consumed in candidate branch |
| 5b Maintenance fold | 5a passes | source-consumer closure, coverage transfer, stale-state diagnostics | Maintenance removed from required host packs |
| 6 registry/cache cleanup | 5b passes | registry absence, synthetic alias rejection, subpath cache proof | aliases retired after activation |
| 7 stale source/docs deletion | all replacement tests and docs gates pass | docs/help/tutorial/doctor goldens and generated references | example pack sources deleted |

Release compatibility matrix:

| Gas City binary | Public Gastown pack | Expected behavior |
| --- | --- | --- |
| old binary | old pack | Existing behavior unchanged. |
| old binary | new compatibility pack | Public pack remains compatible with host Core and legacy coexistence. |
| new binary | old locked pack | Runtime loads only far enough for pre-resolution doctor diagnostics; behavior-changing operations fail closed with version-skew guidance. |
| new binary | new activation pack | Core, provider packs, and public Gastown load without Maintenance; packcompat and generated-manifest gates pass. |
| rollback from new to old | existing city | Doctor-mutated manifests are either readable by old binaries or release notes name explicit downgrade limits and manual recovery. |

Failures are noticed through load failures, doctor diagnostics, emitted
events, packcompat failures, generated-manifest freshness failures, and docs
golden failures. Recovery favors deterministic rerun, quarantine, or explicit
manual guidance over deleting operator content or silently falling back to
retired embedded behavior.

## Open Questions

None.

The public `gascity-packs` commits, generated manifests, pin ledger, ownership
rows, and packcompat transcripts are external prerequisites for Gas City source
deletion and Maintenance removal. They are not open design questions; the
rollout gates above block dependent implementation slices until those
artifacts exist and are cited by exact path and commit.
