# Ritu Raman - Claude

**Verdict:** approve-with-risks

Lane: bootstrap embed cleanup, deterministic test fixtures, test-only no-Core
path containment, hidden-dependency discovery. Iteration-18 review of the live
`design.md` (latest section now **Attempt 17**), re-grounded against
`internal/materialize/`, `internal/doctor/`, and `internal/bootstrap/` in the
tree.

My prior-iteration **[Major]** — the hidden-dependency guard was path-string-only
and blind to production *mechanism-coupled* consumers — is **resolved** by the
new Attempt-17 "Bootstrap Embed Removal And Hidden Dependency Guard." The guard
is now "source-symbol guarded, not path-string-only" and scans imports/symbols
for exactly the couplings I flagged (`bootstrap.BootstrapPacks`,
`bootstrap.PackNames`, `bootstrapSkillDirs()`, `GC_BOOTSTRAP`, skill
materialization, cache helpers), and "Core skill resolution moves to
Core/systempack sources" with a materialization test that fails if any
production path reads `internal/bootstrap/packs/core` (2609-2628). That closes
the silent-breakage hole. The remaining items are Minor: the *prose inventory*
and *slice dispositions* have not caught up to the new symbol guard, so I do not
block.

**Top strengths:**
- **Fixture drift is eliminated by construction, and the no-Core path is
  structurally test-only.** Tests inject inline `fstest.MapFS` with synthetic
  paths (`packs/test-core`), `AssetDir` never `packs/core`,
  `TestBootstrapFixtureIsMinimal` failing on production-only dirs, and real Core
  fidelity relocated to `internal/systempacks` strict integrity (2315-2334,
  3157-3169). Nothing is copied from shipped Core, so nothing can silently
  diverge (red flag #1 closed; lane-Q2 satisfied). No-Core behavior is a
  lower-level `internal/config` loader call, not a flag/env var (3045-3048, 2325),
  closing red flag #2.
- **The guard is now symbol-aware, which is the correct discovery mechanism for
  this lane.** Attempt-17 (2609-2614) scans imports, symbols, tests, docs,
  generated references, and fixtures — not just path strings — so a consumer that
  couples via `bootstrap.PackNames()`/`bootstrapSkillDirs()`/`GC_BOOTSTRAP`
  instead of a literal path is caught. Paired with the "fails if any production
  path reads `internal/bootstrap/packs/core`" materialization test (2628), this is
  a behavioral guard that catches couplings regardless of helper name.
- **Production `bootstrapAssets` is now specified as empty *and never read*.**
  Attempt-17 (2616-2619) states the production FS is a private non-nil empty
  filesystem returning `fs.ErrNotExist` and "must be empty and never read,"
  upgrading the earlier empty-only contract; `TestProductionBootstrapAssetsIsEmpty`
  proves the empty/erroring property (3160-3161).

**Critical risks:**
- **[Minor] The explicit hidden-dependency inventory has not been reconciled with
  the new symbol guard, so the three mechanism-coupled production files are caught
  reactively but never enumerated with a disposition.** The guard names the
  *symbols* (2611-2612), but the prose inventory (3180-3185) and the Attempt-14
  table (2321-2329) still list only path-string-era dependents
  (`cmd/gc/prompt_test.go`, `internal/config/bundled_import_test.go`,
  `examples/gastown/precompact_hook_test.go`, `test/packlint/*`). Verified still
  live and unnamed in the tree:
  - `internal/materialize/skills.go:46,206,684-685` — production skill assembly
    imports `internal/bootstrap` and resolves Core skills through
    `bootstrapSkillDirs()` → `bootstrap.PackNames()`. The skills move is a
    behavior change; 2627 states it generically but no row names this file or its
    slice.
  - `internal/doctor/implicit_import_cache_check.go:47,166,236-248` — reads
    `bootstrap.BootstrapPacks`/`RetiredBootstrapPacks` and save/unset/restores
    `GC_BOOTSTRAP`, then calls `EnsureBootstrap`. When `BootstrapPacks` empties and
    `GC_BOOTSTRAP` is retired, this env dance is vestigial or wrong.
  - `internal/bootstrap/collision.go:46-48` — `PackNames()` derives from
    `BootstrapPacks` and returns empty in production once it empties; its
    post-empty role (retained only for the fixture-collision test vs. dead code)
    is unstated.
  The guard makes these safe (CI fails if coupling remains), so this is not a
  safety gap — but for a "hidden-dependency discovery" mandate the inventory
  should name them so they are migrated proactively, not discovered at
  source-deletion time.
- **[Minor] `GC_BOOTSTRAP` disposition is still "delete OR narrow," and that
  choice is load-bearing for the unnamed doctor consumer.** Attempt-17 (2623)
  keeps "deleted from production semantics unless it can be proved to skip only
  empty bootstrap fixture materialization." The safety invariant (cannot suppress
  Core/systempacks materialization, integrity, classifier, collision, typed
  participation) is firm (2624-2626, 3192-3193), so this is not a safety gap — but
  `implicit_import_cache_check.go:236-245` manipulates `GC_BOOTSTRAP` in
  production, so delete-vs-narrow decides whether that code is removed or rewritten.

**Missing evidence:**
- No slice assignment for updating `internal/materialize/skills.go`,
  `internal/doctor/implicit_import_cache_check.go`, and
  `internal/bootstrap/collision.go`; 2627 states Core skill resolution moves to
  systempacks but does not name the file or the slice that rewrites its
  `bootstrap`-coupled path.
- No statement of `internal/bootstrap/collision.go`'s role after `BootstrapPacks`
  empties — whether required-`core` collision detection moves wholly to
  `internal/systempacks` (leaving `PackNames()` dead in production) or the
  bootstrap collision path is retained only for the synthetic fixture binding.
- No test named to prove the symmetric **skill** installation reads from
  `internal/packs/core`/systempacks (the hook-installation test is named at
  3194-3195; the materialization test at 2628 is a negative read-guard, not a
  positive skill-resolution assertion).

**Required changes:**
- Reconcile the explicit inventory (3180-3185 / Attempt-14 table) with the
  Attempt-17 symbol guard: add `internal/materialize/skills.go`,
  `internal/doctor/implicit_import_cache_check.go`, and
  `internal/bootstrap/collision.go` (plus their `_test.go` siblings) as named
  rows with a disposition (source Core skills from systempacks / remove
  `GC_BOOTSTRAP` + `BootstrapPacks` doctor coupling / state collision home) and a
  slice assignment.
- Add a positive skill-installation/resolution test proving Core skills resolve
  from `internal/packs/core`/systempacks after the move, symmetric to the
  hook-installation test at 3194-3195.
- Make the `GC_BOOTSTRAP` decision singular (delete the production branch) and
  remove the `implicit_import_cache_check.go` env manipulation in the same slice;
  or, if narrowed to test-only, state it is a no-op in every production path
  including that doctor check.

**Questions:**
- After the move, does `internal/materialize/skills.go` source Core skills from
  `internal/packs/core`/systempacks, and is that rewrite in the Core-extraction
  slice or the loading/doctor slice?
- Does required-`core` collision detection move entirely to `internal/systempacks`,
  leaving `internal/bootstrap/collision.go`/`PackNames()` as production-dead code
  retained only for the fixture binding, or does it stay live for non-Core
  bootstrap names?
- Is `GC_BOOTSTRAP` deleted or retained as a test-only token, and is the
  production doctor env manipulation
  (`implicit_import_cache_check.go:236-245`) removed in the same slice?
