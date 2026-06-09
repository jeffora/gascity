# Leah Okafor - Claude

**Verdict:** block

Lane: doctor `--fix` coordinator atomicity, byte-preserving TOML, concurrency
with live controllers, advisory locks, idempotent recovery. Reviewed against
`plans/core-gastown-pack-migration/implementation-plan.md` (`updated_at`
2026-06-09T13:20:59Z). I verified the toolchain rather than trusting prose: the
only TOML library available is `github.com/BurntSushi/toml`, used everywhere via
`toml.NewEncoder(...).Encode(...)` (`internal/configedit/configedit.go:582`,
`cmd/gc/cmd_import.go:1201`, `cmd/gc/cmd_init.go:576`,
`internal/config/site_binding.go:484`, `pack_fetch.go:142`, `config.go:3751`);
the canonical mutator `internal/configedit` documents its own limitation —
*"TOML comments and key ordering are not preserved — that is a limitation of the
underlying decode/encode round trip"* (`configedit.go:547-549`) and special-cases
mtime to avoid *"losing comments on a no-op rewrite"* (`configedit.go:52`); no
shared city-mutation flock exists (only `.gc/controller.lock`, controller-only,
`cmd/gc/controller.go:105`; `internal/packman/install.go` and `check.go` take
none); and the generic mutation driver `internal/doctor/doctor.go:69-74` runs
`c.Fix(ctx)` with no lock/staging/recovery across ≥10 files that advertise
`CanFix()=true` (`cmd/gc/doctor_v2_checks.go` `v2ImportFormatCheck`/
`v2AgentFormatCheck`/`v2FormulasDirCheck`, `cmd/gc/import_state_doctor_check.go:103`,
`doctor_codex_hooks.go`, `internal/doctor/checks.go`, `checks_beads_role.go`,
`checks_semantic.go`, `autofix_skills.go`, ...).

The design's mutation-*safety* skeleton is sound — corruption is contractually
fail-closed and atomicity is modeled as a journal, not as rename semantics. But
three load-bearing mechanisms in my lane are asserted as guarantees while being
unspecified or unbuildable on the current tree, each mapping to one of this
lane's assigned red flags, and the artifact still declares `Open Questions: None`.
That combination is a block in the mutation-safety seat, not approve-with-risks.

**Top strengths:**
- Single lock-first `FixIntent` coordinator as the sole mutation path (lane Q1).
  `internal/doctorfix` (`FixIntent/Plan/Stage/Publish/Recover/Refuse`) is "the
  only path that writes city manifests, lockfiles, installed pack directories,
  runtime-state migrations, or import rewrites," takes a crash-released city
  advisory lock *before* digest preflight, re-validates digest+provenance after
  the lock when a report-only phase read earlier, and re-reads target digests
  before each temp-file rename. Compare-before-rename is the right concurrent-
  writer mitigation, and `support/doctor-fix-inventory.yaml` forces every legacy
  `Fix(ctx)` to be classified before new mutations turn on.
- Atomicity is modeled as a WAL + reconvergence (red flag #1 avoided in shape).
  Multi-file fixes "write durable recovery state before the first publish step,
  stage all edits ... define a single commit point"; death-before-commit reruns
  or rolls back, death-after-commit converges by revalidating Core participation,
  pin installability, lock contents, and marker state. Healthy-city `gc doctor
  --fix` is required byte-identical and idempotent, so untouched cities are never
  rewritten.
- Live-state controller refusal + non-destructive defaults + marker-based runtime
  migration (red flags #2/#3 partial). Controller-running is "discovered from
  live runtime state rather than status files"; stale
  `.gc/system/packs/{maintenance,gastown}` are explicitly *not* auto-deleted
  because they "may contain operator edits"; the runtime-state marker records
  schema version / old / new / staged-archive-digest / completed-steps /
  old-binary post-marker write detection → version-skew diagnostic, and the
  migration table gives per-row merge rules ("reconcile, never delete unknown
  refs", "conflicts block").

**Critical risks:**
- [Major] Byte-preserving scoped TOML editing names no mechanism and is
  contradicted by the in-tree toolchain (lane Q2 / RF2). "Scoped TOML edits
  preserve comments, unknown tables, unknown fields, array order, formatting ...
  otherwise automatic fix refuses" appears only in Testing, never in
  `Proposed Implementation`, and the plan never mentions `internal/configedit` —
  whose own source says comments/key-ordering are not preserved
  (`configedit.go:547-549`), built on the only TOML lib in the tree. The refuse
  guard makes this *safe*, but with current tooling "preserve OR refuse"
  collapses to refuse-on-any-commented-file — an **inert** auto-fix on
  essentially every real city. AC10 makes `gc doctor --fix --non-interactive`
  the canonical mutating repair surface for existing cities, so "safe but inert"
  does not satisfy the requirement either. A bead author cannot tell whether to
  vendor a CST/format-preserving editor, hand-roll span edits, or ship
  refuse-only.
- [Major] Import rewrites match on legacy *name*, not provenance — operator-authored
  legacy-named imports are rewritten without proof they are system-generated (lane:
  rewrites without proof / RF2). The live `Fix` path keys the rewrite off
  `legacyPublicPackImportDetails` / `legacyPublicPackNames` and then calls
  `rewriteLegacyPublicPackImportsFS` (`cmd/gc/import_state_doctor_check.go:103-125`)
  with no check that a matched import was emitted by a prior `gc init`/migration
  versus deliberately authored by the operator (a local fork or an intentionally
  pinned legacy path). The plan routes this mutation "through the mutation
  coordinator" (Slice 2) but never adds a provenance/ownership gate, so RF2 —
  "doctor rewrites user imports or local forks without proof they are
  system-generated" — survives the redesign verbatim. AC10's "rewrites only
  supported local mutable configs ... without guessing policy" requires the
  fixer to *prove* the import is a supported system-generated shape (and the
  file is locally mutable and unmodified beyond that shape) before editing it;
  everything else must be report-only. The design names neither the criterion
  nor the failure mode for a name-match on an operator-authored import.
- [Major] "Rolls back from recovery state" is unimplementable as recorded (lane:
  idempotent recovery / RF1). The recovery record holds "preflight file digests,
  staged paths, publish order, commit point, completed steps, rollback
  instructions, final validation" — but **no copy of the original bytes**. Once
  temp #1 is renamed over original #1, a digest cannot reconstruct original #1,
  so the "rolls back" branch cannot restore an already-published file. AC10
  explicitly offers "durable preflight/journal/**backup** or refusal." Either add
  original-content backup to the record (true rollback), or drop "rolls back" and
  commit to roll-forward-only (durable staged temps + deterministic regeneration
  + idempotent re-apply). As written the two recovery models are conflated and
  neither is fully supported.
- [Major] "A single commit point" is underspecified for N>1 per-file renames
  (lane Q3 / RF1). POSIX has no atomic multi-file rename, and the real fix touches
  ≥4 mutations (import TOML, city.toml/pack manifest, lockfile, installed dirs —
  see the legacy `import_state_doctor_check.go:103` sequence: rewrite imports →
  `syncImports` → `writeImportLockfile` → `installLockedImports`). The plan must
  say whether the commit point is *one atomic marker rename that gates idempotent
  republish of all other files* (half-published pre-marker state never observable
  as valid by loaders), versus journal roll-forward, versus rollback. Without
  this an implementer can read "single commit point" as per-file rename
  atomicity — RF1 verbatim.
- [Major] Mutual exclusion vs controllers/old binaries is a TOCTOU check, not lock
  contention, and "coordinator is the only writer" is asserted but not enforced
  (lane Q1/Q3 / RF3). The plan invokes "the same city advisory lock used by pack
  install/update and import rewrites" as if it exists and already serializes those
  writers — it does not (packman takes no lock; the legacy `Fix` takes none; the
  controller holds a *separate* `.gc/controller.lock`). The plan never states that
  the controller, or a released old binary whose `--fix` predates the lock,
  contends on the *same* lock, so "no controller running" is a window-racy check
  and concurrent old+new `--fix` can interleave across the multi-file commit.
  Separately — unlike the loader lane, which models a `config.Load*` scanner on
  `worker_boundary_import_test.go` — no enforcement test forbids direct writes to
  manifests/lockfiles/pack dirs/runtime state outside `doctorfix`, while the real
  migration surface is the lock-less generic loop at `internal/doctor/doctor.go:74`
  driving ≥10 `CanFix()=true` checks.
- [Major] The loader's own required-pack repair is a second, unlocked writer to the
  surface the coordinator claims to own (lane Q1/Q3 / RF3). Proposed Implementation
  has runtime loading "regenerates missing or corrupt expected files from the
  embedded manifest, prunes generated unexpected effective files, and quarantines
  operator-edited or unclassifiable files" *before any formula/order/script/prompt/
  hook/overlay can be read* (impl plan lines 536-539) — i.e., the controller/read
  path mutates `.gc/system/packs/core` directly, with no city advisory lock named —
  while the doctor section asserts the coordinator "is the only path that writes
  city manifests, lockfiles, installed pack directories, runtime-state migrations,
  or import rewrites" (lines 366-368). Those two statements contradict on "installed
  pack directories": a `doctorfix` fix renaming a Core manifest can race a concurrent
  load-path repair regenerating/pruning the same tree, and the only-writer invariant
  is false on the read path before any enforcement test could catch it. Either the
  load-path repair contends on the same city lock (and the only-writer claim is
  narrowed to operator-content/migrations), or repair is removed from the read path
  and routed through the coordinator.
- [Minor] Recovery durability/idempotence under a *second* crash is unstated. The
  failure-injection test injects "after each staged publish step," but nothing
  states the convergence/replay pass is itself idempotent if the process dies
  *during* replay, nor that the recovery-journal write is fsync-durable before the
  first publish. The journal anchors all cross-file recovery; a torn journal write
  leaves replay with nothing to resume.

**Missing evidence:**
- The concrete byte-preserving TOML mechanism (or an explicit "refuse-only until X
  lands" statement), and the relationship between `internal/doctorfix` and the
  existing `internal/configedit` atomic mutator (two atomic city.toml mutators
  with different preservation guarantees must not clobber each other).
- Where the advisory lock is *released* relative to post-commit validation; the
  text gives acquire-before-preflight but no release point, so "lock held across
  stage/validate/compare/publish" is unverifiable.
- Whether the controller and old binaries contend on the same lock; an enforcement
  test proving `doctorfix` is the only writer of the protected surface.
- Whether the replay pass is idempotent under a crash during replay and whether the
  journal write is fsync-durable before first publish.
- `--non-interactive` behavior on a refuse decision: non-zero exit with guidance,
  or silent no-op?

**Required changes:**
1. Name the scoped-TOML edit substrate in `Proposed Implementation` and reconcile
   it with `internal/configedit` / the BurntSushi comment-loss limitation; add a
   preservation/refusal test. State whether refuse-only is the shipping behavior
   until a format-preserving editor lands, and confirm the AC10 legacy-import
   rewrite still functions under that rule.
2. Make the recovery record sufficient for its stated mode: add original-content
   backup (true rollback) or commit to roll-forward-only and remove "rolls back".
3. Define the commit primitive for N>1 files — one atomic marker rename gating
   idempotent republish, with half-published states unobservable to loaders.
4. Define the city-mutation lock as a real, named, shared OS flock (advisory,
   crash-released — not a PID/status file) and make the controller and any old
   binary contend on the *same* lock, so "controller running ⟹ fix refuses" holds
   by contention, not a liveness check; add an old+new concurrent-`--fix` fixture
   and state the lock release point. Add an enforcement scanner/test that
   `doctorfix` is the only writer, and enumerate the migrated-vs-report-only
   `Fix(ctx)` set behind `internal/doctor/doctor.go:74`. Resolve the contradiction
   that the loader's required-pack repair (impl plan lines 536-539) writes
   `.gc/system/packs/core` during normal loading with no lock while the coordinator
   claims sole ownership of installed pack directories — route that repair through
   the same lock or remove it from the read path.
5. State replay-pass idempotence under a second crash and journal write durability
   (fsync) before the first publish.
6. Add an import-provenance gate: rewrite only imports proven to be supported
   system-generated legacy shapes on locally-mutable, unmodified files; route
   operator-authored, locally-modified, read-only, or transitive imports to
   report-only with manual guidance. Name how provenance is established, since
   `rewriteLegacyPublicPackImportsFS` currently matches on legacy name alone
   (RF2 / AC10).
7. Resolve the above in the artifact rather than carrying `Open Questions: None`,
   since each changes how the 4b/4c doctorfix slices decompose.

**Questions:**
- Is `internal/doctorfix` built on `internal/configedit`, beside it, or a
  replacement — and how do two atomic city.toml mutators avoid clobbering each
  other with different preservation guarantees?
- Given BurntSushi cannot round-trip comments, is refuse-only the shipping
  behavior until a span/CST editor lands?
- Is the city-mutation lock the same flock as `.gc/controller.lock`, or a new lock
  the controller must also acquire at startup? Without one of these, doctor and a
  freshly started controller race.
- Does the recovery record persist the original bytes of every file it overwrites,
  or only digests plus textual rollback instructions?
- For "rollback from new to old binary": is a doctor-mutated city.toml/lockfile
  guaranteed readable by the prior release, or is the supported answer "downgrade
  limits + manual recovery"? The compatibility matrix states both without
  choosing.

---

**Review-grounding note:** The `file:line` evidence above was re-verified against
the live tree this pass — `internal/configedit/configedit.go` (comment/key-order
loss is documented in source and the `ErrUnmodified` mtime guard), the lock-free
generic `c.Fix(ctx)` loop in `internal/doctor/doctor.go`, the controller-only
`acquireControllerLock` flock on `.gc/controller.lock`, the absence of any OS
advisory lock in `internal/packman/{install,check}.go`, and the multi-file
rewrite→`syncImports`→`writeImportLockfile`→`installLockedImports` sequence in
`cmd/gc/import_state_doctor_check.go`. The RF2 finding on name-based (not
provenance-based) import rewriting and its required change were added this pass;
all other findings were confirmed, not taken on trust. The `block` verdict stands.

**Schema conformance (implementation-plan.schema.md):** Conforms to shape. Front
matter has all required keys with `phase: implementation-plan`, `requirements_file`
pointing at the approved `requirements.md`, `status: draft`, and no `design_file`;
the seven required body sections appear in order; the doctor/runtime-state material
is placed across Proposed Implementation, Data And State (recovery/marker schemas +
migration table), Testing, and Rollout And Recovery (slices 4b/4c), with no
`Attempt N Review Response` sections. One in-lane schema smell: the byte-preservation
*behavioral contract* lives only as a Testing assertion; for decomposition readiness
it must appear as a Proposed-Implementation mechanism, not be inferred from a test
row. The `<!-- REVIEW: added per ... -->` HTML comments are benign provenance
breadcrumbs but are review-process residue the schema would rather not carry.
