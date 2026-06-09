# Marcus Driscoll - Claude

**Verdict:** approve-with-risks

Lane: builtin registry identity, synthetic cache pruning, system-pack
materialization, and provider-dependent pack continuity. I reviewed the current
design (`updated_at` 2026-06-09T08:40:42Z) and grounded every claim against live
source: `internal/builtinpacks/registry.go`, `cmd/gc/embed_builtin_packs.go`,
`examples/dolt`, and the Maintenance scripts the migration moves into Core. The
integrity model is strong and the existing code is a good baseline — content-
and-provenance validation already exists and the design preserves it. I am not
blocking. The residual risks are concrete: a provider-pack test and a script
fallback that are coupled to the Maintenance→Core move, an unstated role
division for the *combined* synthetic hash, and a provider notification rewrite
that can collapse two channels. These must be nailed before the
Maintenance-folding and registry/cache-cleanup slices (slices 5–6).

**Top strengths:**
- **Integrity is content+provenance, not existence or path-count — and the
  baseline already enforces it.** `validatePackFiles` (`registry.go:305`) does
  `bytes.Equal(got, want.data)` (323) and rejects unexpected files (340);
  `validateSyntheticRepoFileSet` (349) rejects extras (382); existing tamper
  tests are content/mode/symlink-based (`TestValidateSyntheticRepoRejectsTamperedContent`
  /`...Mode`/`...SymlinkRoot`), not path counts. The design extends this with
  per-pack `pack.toml` + full file-set + content-manifest digests and
  prune/quarantine before the fatal gate (~L2067, ~L3009, ~L3328), and states
  outright "Path, count, and name assertions are insufficient" (~L290). This
  closes my "tamper tests count paths" red flag — provided the existing
  content-based tests are carried forward, not regressed, when Core moves.
- **Registry identity and alias retirement are explicit, testable, and match
  source.** `All()` is today the 5-pack set with Core at
  `internal/bootstrap/packs/core` (`registry.go:53-57`) and `publicSubpathForPack`
  aliases exactly `gastown`/`maintenance` (`registry.go:126-133`). The design
  retires both, moves Core to `internal/packs/core` (~L2762, ~L2938), asserts
  only `core`/`bd`/`dolt` remain (~L3290), and requires negative
  `IsSource`/`NameForSource`/lock/install tests for retired Maintenance sources
  and historical public Gastown aliases (~L2968). This answers lane Q1.
- **Provider continuity is a first-class gate with deterministic selection and a
  Core-repair isolation proof.** `RequiredPackPlan` makes provider selection
  deterministic and immune to public Gastown / retired Maintenance / stale
  materialized dirs (~L2429, ~L2680); "Core repair isolation | untouched provider
  file digest proof after Core repair and offline self-heal" (~L2681); the
  provider exception ledger (~L2380) enumerates the `dolt` role-cleaning rewrites
  with required old/new witnesses. This is the right shape for lane Q2.

**Critical risks:**

- **[Major] A provider-pack test and a position-dependent script fallback are
  coupled to the Maintenance→Core move, and the design names neither.** Core's
  future `dolt-target.sh` sources the *dolt-provider* `port_resolve.sh` via
  absolute `${GC_SYSTEM_PACKS_DIR:-…}/dolt/assets/scripts/port_resolve.sh`
  (`dolt-target.sh:153`) **and** a co-location fallback that walks
  `$SCRIPT_DIR/../../../../../dolt/assets/scripts` (`dolt-target.sh:154-157`).
  That relative depth is anchored to the script's *current* Maintenance location;
  moving it to `internal/packs/core/...` changes the depth and silently breaks
  the fallback unless updated. Worse, `examples/dolt/port_resolve_test.go:148` —
  a test *inside the dolt provider pack* — hardcodes
  `…/gastown/packs/maintenance/assets/scripts/dolt-target.sh`, so the move breaks
  a provider test and couples provider continuity to the Core move. The witness
  row (~L2674) requires "old/new script execution, pack-relative path, env" but
  does not name `port_resolve_test.go`'s migration or pin which resolution path
  (`GC_SYSTEM_PACKS_DIR` vs the fallback) must survive. A path-anchored provider
  test silently dropped here is exactly the path-vs-content failure my lane
  guards. `examples/dolt/pack.toml:6` ("rely on the city's maintenance pack")
  also still names Maintenance and must be rewired to Core with a
  behavior-manifest helper-dependency row.

- **[Major] The `dolt` binding rewrite can collapse two distinct notification
  channels, and a legacy nudge masks delivery failure.** The moved scripts emit
  two channels with different severities: an escalation to `mayor/`
  (`reaper.sh`: `n mayor/ -s "ESCALATION: … [MEDIUM]"`; `jsonl-export.sh`:
  `n mayor/ -s "ESCALATION: … [HIGH]"`) and a summary/DOG_DONE to `deacon/`
  (`n deacon/ "n: $SUMMARY"`). The witness rows separate configured recipients
  (~L2675) and DOG_DONE nudges (~L2677) and ~L1379 forbids swallowing a
  configured-recipient delivery failure, but the design never states that the
  escalation recipient and the summary recipient must remain *two separately
  configurable* targets. If both resolve to one `core.maintenance_worker`, two
  severities merge; and any surviving fire-and-forget `2>/dev/null || true`
  still hides a dropped delivery unless the rewrite removes it. dolt
  backup/stale-db are store-critical, so this is a continuity decision, not a
  free no-op.

- **[Major] The combined `SyntheticContentHash` couples bd/dolt cache identity to
  Core/Maintenance bundle membership, and the design never states the role
  division or tests the shrink transition.** `SyntheticContentHash()`
  (`registry.go:252`) takes no pack argument — it hashes *every* bundled pack
  into one value stored in a single shared marker, `MaterializeSyntheticRepo`
  `os.RemoveAll`s the whole shared dir before rewrite (`registry.go:155`), and
  `ValidateSyntheticRepo` rejects on any global-hash mismatch (`registry.go:225`).
  So removing `maintenance`/`gastown` (and role-cleaning Core) necessarily
  changes the hash that `bd`/`dolt` cache validation depends on, invalidating
  every prior synthetic cache. The design says "Core, bd, and dolt continue to
  use `SyntheticContentHash()`" (~L2943) and asserts bd/dolt "byte-identical"
  continuity, but never states the role division (combined hash = cache identity
  only; per-pack file-set manifest = the tamper authority), never pins the hash
  domain to `{core, bd, dolt}`, and never requires a test that an old 5-pack
  combined-hash cache is treated stale and re-materialized — which is precisely
  the "offline old-cache-to-new-binary self-heal" gate the design itself demands
  (~L900, ~L2007). The mechanism is safe (re-materialize from embed works
  offline), but the continuity claim is unproven as written.

- **[Minor] Old `internal/bootstrap/packs/core` source recognition is left soft.**
  Source recognition derives purely from `All()` (`registry.go:90-139`), so
  dropping the old subpath makes `IsSource` false automatically — but the
  explicit negative-test list (~L2968) covers retired Maintenance sources and
  public Gastown aliases, *not* the old Core source path, and the design only
  says it "should be rejected or covered by explicit migration diagnostics"
  (~L3293). Require a hard negative: `IsSource`/`NameForSource`/lock/install must
  reject `internal/bootstrap/packs/core`, and `publicSubpathForPack` must no
  longer alias `gastown`/`maintenance` — asserted by tests, not left as
  "rejected or diagnostic." This closes the "old core path / retired alias
  accepted silently" red flag at the test layer.

**Missing evidence:**
- Whether `MaterializeRequiredPacks` repairs each required pack in an isolated
  staged directory or in place on the shared `.gc/system/packs` tree. Today
  `MaterializeSyntheticRepo` removes `dst` unconditionally (`registry.go:155`)
  and `MaterializeBuiltinPacks` prunes to a desired set
  (`embed_builtin_packs.go:70`); the new classifier-driven repair (~L3009) runs
  over the shared tree. The Core-repair isolation digest proof (~L2681) is named
  but the staging model that makes it true is not.
- A materialization+integrity fixture for the real upgrade state where
  `.gc/system/packs` is pre-populated with stale `maintenance`/`gastown` dirs:
  only "include-list tests with stale directories present" (~L2987) is named —
  not a test proving the active set is `core`/`bd`/`dolt`, the stale dirs are
  neither pruned nor included, and the bundled synthetic cache is rebuilt.
- The prune-vs-quarantine behavior for required packs. `pruneStaleGeneratedPackFiles`
  (`embed_builtin_packs.go:70`) today *deletes* unexpected files in
  `.gc/system/packs/{core,bd,dolt}`; the design now wants operator-edited/
  unclassifiable files *quarantined*, not pruned. That is a behavior change for
  provider dirs and needs a test (an operator file under `.gc/system/packs/bd`
  is quarantined, not silently deleted).
- Whether the bundled synthetic cache *rejects* a modified-manifest or
  unexpected-file entry for the remaining bundled packs (lane Q3), or relies
  solely on downstream materialization integrity as the backstop.

**Required changes:**
- Name `examples/dolt/port_resolve_test.go` in the test-migration map as breaking
  on the Core move, with a content/sourcing replacement at the Core path; specify
  that Core's `dolt-target.sh` resolves dolt-provider `port_resolve.sh` via
  `GC_SYSTEM_PACKS_DIR` and that the co-location fallback (`dolt-target.sh:155`)
  is updated for the new Core depth (or removed). Rewire `examples/dolt/pack.toml:6`
  to name Core with a behavior-manifest helper-dependency row.
- State the `SyntheticContentHash` role division (combined hash = cache identity
  only; per-pack file-set manifest = tamper authority), pin its domain to
  `{core, bd, dolt}`, and add a test that an old 5-pack combined-hash synthetic
  cache is treated stale and re-materialized offline with `bd`/`dolt` bytes and
  provenance unchanged.
- Add an explicit materialization-isolation contract and fixture: state whether
  `MaterializeRequiredPacks` stages per-pack or repairs in place, and add a test
  (tamper/delete Core → repair → assert `bd`/`dolt` materialized bytes +
  provenance unchanged and prune/quarantine scoped to the Core directory only).
- Convert the old Core source path to a hard negative: `IsSource`/`NameForSource`/
  lock/install must reject `internal/bootstrap/packs/core`, and
  `publicSubpathForPack` must drop the `gastown`/`maintenance` aliases —
  asserted by tests.
- Expand the `dolt` exception rows so the `mayor/` escalation and the `deacon/`
  summary/DOG_DONE remain two distinct configurable recipients (not collapsed),
  with author identity (`--from controller`) and the deacon health-JSON consumer
  each carrying old/new witnesses; remove `2>/dev/null || true` masking on any
  rewritten nudge so a dropped delivery is a non-zero failure (per ~L1379).

**Questions:**
- Does `MaterializeRequiredPacks` stage each required pack in isolation, or repair
  in place on the shared system-pack tree? Provider byte-continuity hinges on it.
- After `dolt-target.sh` moves to Core, does Core's copy resolve dolt-provider
  `port_resolve.sh` purely via `GC_SYSTEM_PACKS_DIR` (robust to the move), and
  what replaces the provider test at `examples/dolt/port_resolve_test.go:148`?
- After the activation pin renames/omits the Core maintenance worker, what is
  dolt's `pool = "dog"` / escalation outcome — binding-resolved, fail with
  `core.maintenance_worker.*`, or compatibility-alias with expiry? dolt
  backup/stale-db are store-critical, so "disabled when omitted" is a continuity
  decision, not a free no-op.
