---
plan_slug: core-gastown-pack-migration
phase: implementation-plan
rig: gascity
rig_root: /data/projects/gascity
artifact_root: /data/projects/gascity/plans
requirements_file: /data/projects/gascity/plans/core-gastown-pack-migration/requirements.md
status: draft
created_at: 2026-06-04T15:07:35Z
updated_at: 2026-06-09T01:20:00Z
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
exact-pack compatibility tests prove behavior preservation. The implementation
lands in small cross-repo slices with fail-closed loader gates, failure-atomic
doctor fixes, and operator-visible diagnostics.

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

Gas City adoption of `internal/config/PublicGastownPackVersion` has two
meanings. The compatibility pin proves the new public pack can coexist with the
current loader while in-tree sources still exist. The activation pin is consumed
only in the same candidate branch that removes Maintenance from required host
packs and proves no-Maintenance production loading.

### Behavior Evidence Contract
<!-- REVIEW: added per behavior-evidence-contract -->

Add a generated Behavior Evidence manifest and packcompat gate before any
dependent move. The Gas City side stores generated evidence under
`plans/core-gastown-pack-migration/artifacts/` during planning and then moves
the checked-in generator, schemas, and tests into the implementation-owned
paths selected by the first task slice.

Each manifest row must include:

- stable row id;
- old owner, path, asset kind, and helper dependencies;
- trigger, requester, detector, route metadata, mail/nudge target, prompt
  fragment, script branch, runtime-state path, or named-session behavior;
- old witness: source assertion, test, fixture, golden output, or command
  transcript;
- new owner: Core, public Gastown, provider pack, docs-only, or approved
  retirement;
- new path and new witness;
- immutable public Gastown commit for public-pack rows;
- consuming `internal/config/PublicGastownPackVersion` value;
- semantic-equivalence assertion, or approved delta/removal record with owner,
  reason, replacement, and operator impact.

The generator walks old Gas City behavior-bearing sources under Core,
Maintenance, Gastown examples, hook overlays, formulas, orders, prompts,
skills, shell scripts, doctor strings, route metadata, notification templates,
runtime-state helpers, tests, and helper references. CI fails if a moved,
split, generalized, deleted, or helper-dependent asset lacks a row, if a row
lacks old and new witnesses, or if a semantic delta lacks an approved record.

`test/packcompat` consumes the manifest and the exact public pin. It installs
public Gastown through the ordinary remote-pack path or a validated ordinary
remote cache, composes moved formulas and orders, resolves moved scripts using
pack-relative paths, verifies hook overlays and configured agents, and checks
one assertion per manifest row.

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

Normal command paths and behavior-driving production `internal/` paths call
`LoadRuntimeCity` or `LoadRuntimeCityNoRefresh`; they do not hand-build
required-pack include lists or call `config.Load*` directly. Low-level
`internal/config` tests may still call lower-level loaders because they test
config behavior, not full `gc` runtime behavior.

Runtime loading runs two fatal gates. Before config resolution it materializes
and validates the required file set for Core plus provider packs (`bd` and
`dolt` as selected today). After config resolution it validates typed
`RequiredSystemPackParticipation` records proving every required pack
participated in the resolved config. A generated or repaired Core tree without
typed participation is a load failure.

No-refresh reloads do not repair. If Core is missing, corrupt, stale, shadowed,
or missing participation, the controller keeps the last-known-good runtime
config only for read-only status/reporting, publishes an event and diagnostic,
and pauses or refuses behavior-changing dispatch, formula, order, hook, prompt,
and agent-start operations until a successful refreshed load occurs. API and
CLI paths surface the same diagnostic instead of silently continuing with
current invalid config.

Add scanner tests modeled on `cmd/gc/worker_boundary_import_test.go` to reject
production direct calls to `config.Load`, `config.LoadCity`,
`config.LoadWithIncludes`, aliases, wrappers, method/function values, and
manual required-pack include assembly in `cmd/gc` and behavior-driving
`internal/` packages. Any exception is generated into a partial-read allowlist
with file, function, call kind, fields consumed, reason, and a focused test
proving it cannot drive normal runtime behavior.

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

Do not delete stale `.gc/system/packs/maintenance`,
`.gc/system/packs/gastown`, or `.gc/runtime/packs/maintenance` directories
during startup or `gc doctor --fix`. They are ignored by active discovery and
reported as legacy state because they may contain operator edits.

### Doctor And Runtime-State Mutation Safety
<!-- REVIEW: added per doctor-runtime-state-safety -->

Replace legacy direct `Check.Fix(ctx)` mutation with a `FixIntent` plus mutation
coordinator API before public-pin, import-rewrite, or runtime-state fixes are
enabled. Checks either return a structured intent or are refused for automatic
fix with manual guidance. The coordinator is the only path that writes city
manifests, lockfiles, installed pack directories, runtime-state migrations, or
import rewrites.

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

Runtime-state migration moves JSONL archive state, spawn-storm ledgers,
refs/remotes, escalation fields, pending archive push state, and order
skip/tracking compatibility aliases under Core-owned paths only after shared
lock acquisition and staged archive-copy digest checks. The migration marker
records schema version, old path, new path, staged archive digest, completed
steps, and old-binary post-marker write detection. If an old binary writes
after the marker, the new binary reports a version-skew diagnostic and requires
manual reconciliation or a deterministic re-upgrade flow.

### Role Neutrality And Configurable Bindings
<!-- REVIEW: added per role-neutrality-contract -->

Generate a role-surface manifest covering Go, TOML, shell, Markdown,
templates, generated command text, API classifications, dashboard/OpenAPI
generated references, tmux theme helpers, default scaffolding, warmup mail
defaults, prompt fallbacks, formulas, overlays, metadata, tests, docs, and
public Gastown companion files.

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

Move or rewrite active role-specific surfaces. Branch pruning and
`mol-polecat-*` move to public Gastown. `mol-shutdown-dance` may keep generic
stuck-session due process in Core, but Gastown detector/requester examples move
to Gastown. `mol-review-quorum`, provider overlays, Dog prompt fragments,
review checks, and tmux theme APIs are assigned by manifest rows before source
moves.

### Bootstrap Fixture Isolation
<!-- REVIEW: added per bootstrap-fixture-isolation -->

Audit every `GC_BOOTSTRAP` production and test dependency, including doctor
checks, command tests, testscript defaults, prompt tests, config bundled-import
tests, precompact hook tests, packlint, generated docs, hook references, and
helper paths.

Production bootstrap no longer embeds Core assets. Tests that need bootstrap
assets use an empty `fs.FS` fixture or minimal inline fixture whose `Stat`,
`WalkDir`, and `ReadFile` behavior is asserted. Fixture guard tests fail if
allowed test paths copy production-only Core directories such as `formulas/`,
`orders/`, `overlay/`, `skills/`, or `assets/prompts/`.

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
status, and freshness timestamp.

Behavior evidence is durable generated data. Gas City planning artifacts live
under `plans/core-gastown-pack-migration/artifacts/` until implementation
chooses checked-in generator and schema paths. Public Gastown produces
`gastown/behavior-preservation.yaml` and `gastown/public-gastown-pins.yaml`.
CI treats those artifacts as stale if their source digests, old/new witness
digests, or public pin values do not match the current tree.

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
- Scanner tests reject production direct `config.Load*` bypasses and active
  Core role-name references outside generated allowlists.

Behavior and compatibility tests:

- `go test ./test/packcompat -run TestPinnedPublicGastownBehavior` runs first
  in compatibility-pin mode, then in no-Maintenance production-loader mode in
  the activation slice.
- `go test ./test/packlint/...` validates role-surface, retired-source,
  wording/docs, generated-reference, and old bootstrap path inventories.
- Representative Core formula/order tests compose molecules, assert resolved
  step content, verify run-target bindings and configured recipients, and
  execute moved scripts from pack-relative paths.
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

Docs and release tests:

- Golden tests cover doctor output, CLI help, first-run text, tutorial
  transcripts, and docs wording for minimal and Gastown-template cities.
- `make dashboard-check` runs only if API, dashboard, generated OpenAPI, or
  generated TypeScript files change.
- Broad gates after each code slice: `make test-fast-parallel` and
  `go vet ./...`. High-risk loader/doctor/runtime-state slices also run the
  sharded process and integration targets from `TESTING.md`, including
  `make test-cmd-gc-process-parallel` and
  `make test-integration-shards-parallel`.

## Rollout And Recovery

Slice 1 is the public Gastown prerequisite. Land the `gascity-packs` branch
with moved Gastown-owned assets, behavior-preservation manifest, pin ledger,
replacement tests, ownership rows, and host-Core/no-Maintenance proof. Gas City
source deletion and Maintenance removal wait until this prerequisite exists.
Rollback is to leave Gas City unchanged and not consume the pin.

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

Slice 4 adds `internal/systempacks`, strict required-pack validation, typed
participation, production loader scanner, partial-read allowlist, Core doctor,
pre-resolution import recovery, version-skew diagnostics, and the mutation
coordinator. Rollback is to route runtime loads back through the old wrapper
only if no mutation coordinator state has committed; otherwise rerun the
coordinator recovery path first.

Slice 5 consumes the public activation pin and folds Maintenance. In the same
candidate branch, update `PublicGastownPackVersion` to the activation commit,
remove Maintenance from `requiredBuiltinPackNames`, move Core-owned
Maintenance assets into Core, consume Gastown-owned assets from the public pack,
and run packcompat in no-Maintenance production-loader mode. If the activation
commit cannot satisfy old/new binary compatibility and duplicate-definition
requirements, stop and convert the slice into a paired cross-repo activation
boundary. Rollback is to restore the compatibility pin and re-enable
Maintenance before source deletion.

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
