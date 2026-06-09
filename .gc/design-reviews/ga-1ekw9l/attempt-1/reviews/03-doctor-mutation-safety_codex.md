# Design Review: 03-doctor-mutation-safety (Codex)

Persona: Leah Okafor  
Focus: doctor fix coordinator atomicity, byte-preserving TOML, live-controller
concurrency, advisory locks, idempotent recovery  
Artifact reviewed: `plans/core-gastown-pack-migration/implementation-plan.md`  
Verdict: block

I did not read the Claude lane output before writing this review.

## Summary

The plan has moved in the right direction: `internal/doctorfix`, a mandatory
coordinator boundary, an inventory of existing mutation paths, a city advisory
lock, durable recovery state, and doctor-owned runtime-state migration are the
right primitives for this risk area. However, two parts are still not
decomposition-safe. The TOML preservation rule appears as a test expectation
rather than an API invariant, and the old-binary/concurrent-writer story only
has explicit divergence detection for runtime-state markers, not for every
surface the coordinator mutates.

## Findings

### BLOCKER: TOML byte-preservation/refusal is not part of the `FixIntent` contract

The implementation plan says existing direct `Check.Fix(ctx)` mutation will be
replaced with `FixIntent`, `Plan`, `Stage`, `Publish`, `Recover`, and `Refuse`
operations, and that the coordinator is the only writer for manifests,
lockfiles, installed pack directories, runtime-state migrations, and import
rewrites (`implementation-plan.md:361-368`). It also requires a doctor-fix
inventory (`implementation-plan.md:370-377`) and tests that scoped TOML edits
preserve comments, unknown tables, unknown fields, array order, formatting, and
unrelated lock entries or refuse (`implementation-plan.md:659-662`).

That is not yet enough to prevent the known unsafe shape. The current import
doctor path still demonstrates the exact risk: `importStateDoctorCheck.Fix`
collects imports, rewrites legacy public imports, writes `packs.lock`, and
installs locked imports (`cmd/gc/import_state_doctor_check.go:103-132`).
`rewriteLegacyPublicPackImportsFS` mutates parsed manifest/config maps
(`cmd/gc/import_state_doctor_check.go:246-278`), and `writeCityPackManifest`
re-encodes the entire `pack.toml` via `toml.NewEncoder`
(`cmd/gc/cmd_import.go:1129-1207`). A coordinator around that flow can make the
write serialized and recoverable, but it does not make the TOML edit
byte-preserving.

The design needs to make preservation/refusal a first-class property of the
intent, not only a test assertion. Required change:

- Define a `ScopedTomlPatch` or equivalent intent payload that is generated
  from original bytes, names the exact TOML file, import key/table spans, and
  expected original digest, and carries an "outside bytes unchanged" proof.
- Require `Stage` to render bytes and compare untouched regions before any
  publish step. If the patcher cannot preserve comments, unknown fields,
  table order, array order, or formatting outside the intended edit, the
  intent must be `Refuse`.
- Permit full-file TOML re-encoding only for files classified as generated and
  owned by Gas City, with that classification recorded in the doctor-fix
  inventory.
- Name existing helpers such as `writeCityPackManifest` and
  `writeCityImportManifestFS` as unsafe for scoped user TOML rewrites unless
  they are replaced or wrapped by the byte-preserving patch path.

Without this, AC10 can pass for selected fixtures while another doctor path
still normalizes or drops operator-authored TOML during a "safe" automatic fix.

### BLOCKER: Old-binary and concurrent-writer recovery is incomplete outside runtime-state migration

The plan correctly requires the new coordinator to acquire a crash-released
city advisory lock, repeat validation after the lock, and refuse automatic fix
when a live controller for the same city is detected from live runtime state
(`implementation-plan.md:379-383`). It also says multi-file fixes write durable
recovery state before publish, re-read target digests before each temp-file
rename, and define a single commit point (`implementation-plan.md:385-390`).

The unresolved part is that old binaries and existing direct mutators will not
honor the new advisory lock. Compare-before-rename protects only the next file
about to be renamed. It does not detect a concurrent old binary mutating an
already-published file while the coordinator is publishing the next file. The
plan explicitly records old-binary post-marker write detection for
runtime-state migration (`implementation-plan.md:399-407`,
`implementation-plan.md:575-579`), but the coordinator mutates more surfaces
than runtime-state markers: city manifests, lockfiles, installed pack
directories, and import rewrites (`implementation-plan.md:361-368`).

This matters for a migration where `pack.toml`, `city.toml`, `packs.lock`,
installed pack directories, cache proof, and runtime-state markers must agree.
A concurrent legacy `gc doctor --fix`, `gc import install`, or live controller
reload could leave a city with a new public Gastown import, an old lockfile, and
partially installed pack content. The current plan says post-commit validation
checks Core participation, public-pin installability, lock contents, and
runtime-state marker state (`implementation-plan.md:388-390`), but it does not
state how a divergence on manifests/lock/cache after one rename is detected,
journaled, refused, or repaired before behavior resumes.

Required change:

- Extend the recovery marker/journal schema to all doctor-mutated surfaces, not
  only runtime-state migration. It should record preflight and staged digests
  for `pack.toml`, `city.toml`, `packs.lock`, installed pack directories,
  cache proof files, and runtime-state marker files when they participate in
  one fix.
- Define the commit marker mechanics: which durable write is the commit point,
  what "prepared but not committed" means, and what "committed but
  post-validation failed" means.
- Require final all-surface digest validation before marking the recovery
  record complete. A post-publish drift in any touched surface must leave a
  version-skew or concurrent-write diagnostic that blocks behavior-changing
  runtime operations until manual reconciliation or deterministic re-upgrade.
- Add failure-injection tests with a simulated old binary or direct writer that
  changes an already-published file between publish steps.

Until this is specified, the plan still relies too much on per-file rename
atomicity for a cross-file migration.

## Non-Blocking Notes

The "doctor-owned runtime-state migration" constraint is strong and should
stay: controller startup, API handlers, and reload paths may diagnose and
refuse, but should not mutate runtime state (`implementation-plan.md:392-397`).

The stale generated pack directories rule is also correct: startup and doctor
should report them as legacy operator state and not delete them automatically
(`implementation-plan.md:353-356`, `implementation-plan.md:598-601`).

## Required Updates Before Decomposition

1. Add TOML patch/refusal semantics to the `internal/doctorfix` API contract
   and the doctor-fix inventory acceptance rules.
2. Extend recovery and old-binary divergence detection from runtime-state
   markers to every manifest, lock/cache, installed-pack, and runtime-state
   surface the coordinator can mutate.
3. Add the corresponding proof rows to AC10 and the slice 4b/4c tests,
   including concurrent old-writer failure injection across publish steps.
