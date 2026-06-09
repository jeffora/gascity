# Hiroshi Tanabe - Claude

**Verdict:** approve-with-risks

Lane: production Core embed removal, non-nil empty bootstrap fs, fixture-only
tests, `GC_BOOTSTRAP` skip containment, hidden-dependency inventory. Scope is the
**Bootstrap Fixture Isolation** section (implementation-plan.md:459-493), its
Slice 3 gate (747-750, 801), and Testing rows (611-613, 624-627). Every
file:line below was verified against the live tree.

Not a block: production runtime is **neutral** to this section because
`BootstrapPacks` is empty in production (`bootstrap.go:35-39`; the
`len(BootstrapPacks) > 0` loop at `bootstrap.go:112` is a prod no-op), so embed
removal cannot regress a running city. Deletion of the old package is also gated
behind the Slice 3 scanner (plan:801). The risks below are enforcement-mechanism
gaps — the plan states the right invariants as prose but over-credits the
compile-time proof and under-specifies the skip/scanner gates — and should be
closed before decomposition so the beads encode gates, not intentions.

**Top strengths:**
- **Non-nil `bootstrap.EmptyFS` with asserted `Stat`/`WalkDir`/`ReadFile`**
  (plan:482-485) is a *structural* answer to the "nil `fs.FS` / `AssetDir
  packs/core` panic" red flag, not a runtime guard, and is net-new (no `EmptyFS`
  exists today).
- **The fixture guard bans copying production-only Core dirs** (`formulas/`,
  `orders/`, `overlay/`, `skills/`, `assets/prompts/`, plan:478-480) rather than
  freshness-checking copies — and those names map exactly onto Core's real
  top-level content (`internal/bootstrap/packs/core/{assets,formulas,orders,
  overlay,skills}`). Banning the copy kills the drift red flag at its source.
- **The hidden hooks dependency is surfaced and pinned to a test.** The plan
  names `internal/hooks/hooks.go` reading overlays from Core (verified:
  `hooks.go:21` import + `core.PackFS` reads at `hooks.go:174,177,185,782`) and
  requires `go test ./internal/hooks ./internal/bootstrap -run 'Core|Bootstrap|
  Hook'` (plan:611-613).

**Critical risks:**

- **[Major] "Package deletion is the compile-time proof" (plan:471) covers only
  the 2 Go imports; the production `Subpath` string and 4 test sites survive
  deletion.** Verified: only `internal/builtinpacks/registry.go:20` and
  `internal/hooks/hooks.go:21` import the package and break at compile. Everything
  else is a **string literal** the compiler never validates:
  - production `registry.go:53` — `{Name:"core", Subpath:"internal/bootstrap/
    packs/core", FS: core.PackFS}`. The `FS:` field forces an import edit, but the
    `Subpath` string can ship pointing at a deleted path;
  - **security** tamper/injection tests `registry_test.go:197,214,281` and
    `bundled_import_test.go:44,68` (they write `internal/bootstrap/packs/core/
    pack.toml` and `.../agents/injected/prompt.md "malicious"` as their baseline);
  - real-asset reads `prompt_test.go:781-782` (`.../core/assets/prompts/
    {pool-worker,graph-worker}.md`);
  - source-URL/subpath cases `remotesource_test.go:16,18` and
    `registry_test.go:23,52-58`.
  Post-deletion these compile and silently test/read a fictional path — the
  security tests being the worst case (vacuous pass). The plan lists "`Subpath`
  strings" in scanner scope (plan:472) but still credits compile-deletion as
  *the* proof; it must name the scanner as the gate for these specific sites and
  require they repoint to `internal/packs/core`.

- **[Major] `GC_BOOTSTRAP=skip` "structurally unreachable" is asserted, not
  enforced — and a production path already defeats skip.** Plan:489-493 says what
  skip must not skip but adds no guard that `internal/systempacks` never reads
  `GC_BOOTSTRAP`. Today containment holds only by construction (skip read solely
  at `bootstrap.go:72`), but nothing stops a future `if GC_BOOTSTRAP=="skip"
  {return}` inside the new loader from reopening a production required-pack
  bypass that passes CI. Worse, `ensureBootstrapForDoctor`
  (`implicit_import_cache_check.go:235-246`) is a **production** path that
  `os.Unsetenv("GC_BOOTSTRAP")`/restores to force bootstrap to run regardless of
  skip — the plan names the function (plan:486) but not its env dance. And
  `cmd/gc/main_test.go` sets skip as a blanket testscript default. The lane needs
  a scanner forbidding any `GC_BOOTSTRAP` read under `internal/systempacks`, plus
  a positive test that `LoadRuntimeCity` fully materializes and validates required
  Core with `GC_BOOTSTRAP=skip` set.

- **[Major] The empty/minimal-fixture directive is conflated with tests that need
  *real* Core content.** Plan:476-480 routes "tests that need bootstrap assets" to
  empty/minimal fixtures, but `prompt_test.go` (reads actual prompt files) and the
  synthetic-repo security tests (`registry_test.go`, `bundled_import_test.go`,
  which encode the bundled Core tree as their tamper baseline) need the *real*
  relocated assets, not a fixture. These must repoint at `internal/packs/core`
  (the lane's "relocated system-pack wrapper"), which is distinct from the
  empty/minimal bootstrap fixtures the copy-ban governs. The section should split
  "needs bootstrap materialization" from "needs real Core assets."

- **[Minor] Deletion is gated to Slice 3 but not coupled to removing `//go:embed
  packs/**`.** Slice 3's one-way boundary (plan:801) authorizes deleting the old
  Core package after the scanner passes — so deletion *is* pinned, contrary to a
  pure "unpinned" reading. The real gap: deleting `internal/bootstrap/packs/core/`
  empties `internal/bootstrap/packs/`, turning `//go:embed packs/**`
  (`bootstrap.go:22`) into a `no matching files` build error. Deletion,
  `//go:embed` removal, and defaulting `bootstrapAssets`/`embeddedBootstrapPacks`
  to `EmptyFS` (`bootstrap.go:22-25`) must be the **same commit** or the tree
  won't build. This is a second compile-time forcing function the plan should
  credit and sequence explicitly.

- **[Minor] `EmptyFS.Open(".")` must return an `fs.ReadDirFile`.** Tests assert
  `fs.WalkDir` against it (plan:485) and doctor may traverse it; "returns an empty
  directory" is insufficient unless that file implements `ReadDir` returning
  empty (else `WalkDir` errors/panics). `testing/fstest.MapFS{}` satisfies this
  and is the natural minimal *test* fixture.

**Missing evidence:**
- No checked hidden-dependency inventory — plan:462-465 lists *areas* ("doctor
  checks, prompt tests, packlint, hook references…"), but the mandate is an
  enumerated inventory. Its seed rows are concrete and verified: 2 Go imports
  (`registry.go:20`, `hooks.go:21`) + 1 production `Subpath` string
  (`registry.go:53`) + 4 string/path test files (`registry_test.go`,
  `bundled_import_test.go`, `remotesource_test.go`, `prompt_test.go`) + 2
  `GC_BOOTSTRAP` readers (`bootstrap.go:72`, `implicit_import_cache_check.go:236`)
  + the `//go:embed packs/**` directive (`bootstrap.go:22`).
- No statement of the post-migration `GC_BOOTSTRAP` read-set end state ("empty-
  fixture materialization only; zero reads in `systempacks`"), nor whether
  `ensureBootstrapForDoctor`'s unset/restore is removed as vestigial.
- No freshness/anchor for the fixture-guard denylist: it enumerates 5 dir names
  but omits `pack.toml` (which `materializeBootstrapPack` and the collision tests
  require), so the denylist should be expressed as "no test tree may duplicate any
  file under `internal/packs/core`" rather than a fixed 5-name list.

**Required changes:**
- Reframe the Q1 proof as coupled, not import-only: in one named slice, delete
  `internal/bootstrap/packs/core/`, remove `//go:embed packs/**`, default
  `bootstrapAssets`/`embeddedBootstrapPacks` to `EmptyFS`, repoint
  `registry.go:53`'s `Subpath`, and run the path scanner over the four string-only
  sites (security tests especially). Add a check that `internal/bootstrap/packs/`
  carries no embeddable file.
- Enforce the skip invariant: a scanner failing on any `GC_BOOTSTRAP` reference
  under `internal/systempacks`, plus a positive test that `LoadRuntimeCity` fully
  validates required Core with `GC_BOOTSTRAP=skip` set. Decide the fate of
  `ensureBootstrapForDoctor`'s env dance and the blanket `main_test.go` skip.
- Split the test-isolation directive: bootstrap-materialization tests use
  empty/minimal `fs.FS` fixtures; real-Core-asset tests (`prompt_test.go`, the
  synthetic-repo security tests) repoint at `internal/packs/core`, moving their
  baselines in the same slice.
- Seed the hidden-dependency inventory with the file:line rows above so
  decomposition produces beads against an explicit list, not a prose audit.

**Questions:**
- Is the Slice 3 package deletion intended as the Q1 proof with the scanner as a
  re-introduction guard, given only 2 of 7 sites are compiler-visible — and does
  the relocation commit also drop `//go:embed packs/**` so the tree builds?
- Is `GC_BOOTSTRAP=skip` retained as-is, narrowed, or renamed (e.g.
  `GC_BOOTSTRAP_FIXTURE`) to make a production required-pack bypass impossible,
  and will `internal/systempacks` be contractually forbidden from reading it
  (test-enforced)? "If retained for tests" leaves this open while Open Questions
  says `None`.
- After the copy-ban, do the prompt and synthetic-repo security tests read the
  relocated `internal/packs/core`, and does the `bundled_import_test.go` injection
  baseline move to the new path in the same slice?

---

**Review-grounding note:** Every `file:line` above was independently re-verified
against the live tree this pass and all hold: `bootstrap.go` (`//go:embed
packs/**` L22, empty `BootstrapPacks` L39, `len(BootstrapPacks)>0` no-op L112,
`GC_BOOTSTRAP` skip read L72, and `fs.WalkDir(bootstrapAssets,…)` at L220/241
— confirming the `EmptyFS.Open(".")` must-be-`ReadDirFile` risk); the two Go
imports (`internal/builtinpacks/registry.go:20`, `internal/hooks/hooks.go:21`)
plus the production `Subpath` string (`registry.go:53`); and the four string-only
test sites. Two inventory rows omitted their package directory and are pinned
here for decomposition: the `GC_BOOTSTRAP` unset/restore dance is
`internal/doctor/implicit_import_cache_check.go:236-245` (`ensureBootstrapForDoctor`,
L235, invoked from a production check at L105), and the injection-baseline writes
are `internal/config/bundled_import_test.go:44,68`. The two production
`GC_BOOTSTRAP` readers are exactly `bootstrap.go:72` and
`implicit_import_cache_check.go:236`; the blanket testscript skip is
`cmd/gc/main_test.go`. No additional in-lane findings surfaced; the
approve-with-risks verdict and required changes stand as written.

**Schema conformance (secondary to my lane):** the artifact matches
`gc.mayor.implementation-plan.v1` — front matter has `phase: implementation-plan`,
a `requirements_file`, `status: draft`, no `design_file`; all seven required
top-level sections appear in order; no "Attempt N Review Response" sections; and
`Open Questions` resolves to `None`. One soft drift: the inline `<!-- REVIEW:
added per … -->` markers (e.g. plan:108,149,222,460,474) carry review provenance
that the schema places in the workflow artifact directory, not in
`implementation-plan.md`. Not a structural violation; worth scrubbing.
