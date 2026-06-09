# Owen Gallagher - Claude

**Verdict:** approve-with-risks

I reviewed only my lane: the Core/Gastown ownership split, retired Maintenance
source containment, a single active-discovery classifier, and no duplicate
active behavior. I also schema-checked the artifact. The containment
*architecture* is sound and decomposable — one `internal/packsource` classifier,
an `ActiveRootsFor(kind)` discovery boundary, a zero-duplicate-active runtime
gate, preserved-but-ignored stale directories, and a packcompat activation mode
that fails when behavior resolves from a retired directory. The risks are
specification gaps that decide whether that architecture actually holds: the
classifier and the enumeration API are not reconciled to one authority, the
production loader's named gates don't include retired-import rejection, and the
anti-glob scanner is asserted over a large, real surface (114 raw-walk sites
today, including a literal pack glob) without a decidable scoping rule or a
current-system inventory. None is a redesign; all are bounded edits.

**Top strengths:**
- **A single classifier is named as the authority for retired-source decisions.**
  `internal/packsource` is "the sole authority for retired Maintenance and
  Gastown source classification," and "all load, install, cache, lockfile,
  materialization, discovery, doctor, docs-lint, generated-reference-lint, and
  public-source normalization paths use this classifier instead of duplicating
  string checks" (L320–326). That directly targets my "independent globs/string
  checks bypass classification" red flag.
- **Duplicate-active is a fatal runtime gate across the risky views.** The
  zero-duplicate/zero-merge gate compares "active bundled, active public, stale
  generated, synthetic cache, ordinary remote cache, compatibility pin,
  activation pin, and old/new binary views," and "if the same behavior id is
  active from more than one source… runtime loading fails before any behavior
  executes" (L347–351). Combined with packcompat activation mode failing when an
  assertion resolves from `examples/gastown`, `.gc/system/packs/gastown`,
  `.gc/system/packs/maintenance`, or a synthetic alias (L214–219), this closes
  most of "duplicate active during compatibility."
- **Stale directories are contained, not deleted.** `.gc/system/packs/maintenance`,
  `.gc/system/packs/gastown`, and `.gc/runtime/packs/maintenance` "are ignored by
  active discovery and reported as legacy state" and never auto-deleted
  (L353–356, L598–601). That answers my "can stale dirs stay on disk without
  entering active behavior" question for the non-rollback case.

**Critical risks:**

- **[Major] Two discovery authorities (`internal/packsource` classifier and
  `ActiveRootsFor`) are not reconciled, so "one classifier" is not actually
  guaranteed.** The classifier list (L320–326) covers load/install/cache/lock/
  docs-lint/normalization; the separate `ActiveRootsFor(kind)` mandate (L418–424)
  covers prompt/formula/order/script/hook enumeration. Both live in
  `internal/packsource`, but the plan never states that `ActiveRootsFor` derives
  its roots *solely from* classifier verdicts. If `ActiveRootsFor` has any
  independent root-selection logic, there are two sources of truth for "what is
  active" that can diverge — re-opening the containment hole one level up. The
  plan must state `ActiveRootsFor(kind)` returns only roots the classifier marks
  active.

- **[Major] The production loader's named gates omit retired-source rejection for
  the single-import case.** `LoadRuntimeCity`'s two fatal gates are file-set
  integrity and typed participation (L259–271); the zero-duplicate gate
  (L347–351) only fires when a behavior id is active from *more than one* source.
  A city that imports only a retired source (e.g. a stale `[imports.maintenance]`
  pointing at `.gc/system/packs/maintenance`, no duplicate) is not obviously
  rejected by any named loader gate — yet AC5 and my lane require active
  resolution to reject it before behavior discovery. The classifier is said to be
  used by "load" paths (L320), but it is not wired into `LoadRuntimeCity`'s gate
  sequence. Add retired-source rejection as an explicit pre-resolution loader
  gate, not only an install/lock/docs concern.

- **[Major] The anti-glob `ActiveRootsFor` scanner is asserted over a large
  surface with no decidable scoping rule and no current-system inventory.** The
  plan says scanner tests "reject raw `fs.ReadDir`, `fs.WalkDir`, `filepath.Walk`,
  `Glob`, or string-prefix enumeration over pack roots" (L422–424), but "over
  pack roots" is a runtime property a static scanner cannot generally decide. In
  the live tree there are ~24 pack/behavior enumeration sites this must capture —
  `internal/orders/discovery.go:63,128`, `internal/formula/source.go:60`,
  `internal/overlay/overlay.go:75,148`, `internal/hooks/hooks.go:177`,
  `internal/packregistry/catalog.go:354`, `cmd/gc/cmd_formula.go:347`,
  `cmd/gc/prompt.go:246`, and a literal `filepath.Glob(packGlob)` at
  `cmd/gc/cmd_prompt.go:614` (exactly the red-flag shape) — out of 114 total
  raw-walk sites in non-test `internal`+`cmd`. None of these appears in the plan's
  Current System section. Without a typed `PackRoot` handle that makes "this walk
  is over a pack root" statically decidable, the scanner is either toothless
  (misses dynamically-constructed pack paths like `cmd_prompt.go:614`) or drowns
  in false positives across the other ~90 walks and becomes an allowlist sink.

- **[Major] Split assets risk a latent Core→Gastown reference.** `mol-shutdown-dance`
  keeps "generic stuck-session due process in Core, but Gastown
  detector/requester examples move to Gastown" (L449–451). The plan does not
  state how the Core remainder stays self-contained when the Gastown-owned
  detectors are absent. If Core's `mol-shutdown-dance` references a detector only
  Gastown supplies, that is a Core→Gastown dependency that violates the host-Core
  boundary (Core must never reference Gastown; Gastown patches Core). The plan
  must require every split asset's Core remainder to resolve with Gastown absent,
  proven by a "Core-only renders/composes the split asset" test plus the role
  scanner.

- **[Minor] The rollback narrative needs the pin and source reverted together to a
  packcompat-proven coexistence state.** Slice 5b rollback is "restore the
  compatibility pin and re-enable Maintenance" (L775–776). That is safe only
  because the compatibility pin is required to omit assets that collide with
  Maintenance; otherwise re-enabling Maintenance while the activation pin's
  Gastown assets remain installed trips the zero-duplicate gate and rollback
  fails closed (a deadlock, not a recovery). Make the pin↔source coupling
  explicit and add a rollback test asserting the post-rollback state is the
  duplicate-free compatibility coexistence already proven by packcompat.

**Missing evidence:**

- No current-system inventory of the order/formula/overlay/prompt/hook discovery
  call sites that `ActiveRootsFor` must capture; the Current System section names
  registry, embed, hooks, bootstrap, import-state, examples, and tests, but none
  of the enumeration paths above. The schema requires current-system claims to
  cite concrete files for the work being proposed.
- No statement that `ActiveRootsFor(kind)` is implemented on top of the classifier
  (single authority).
- No explicit loader-level retired-import rejection step inside `LoadRuntimeCity`.
- The inline `<!-- REVIEW: added per … -->` markers (e.g. L314, L345, L359, L460)
  are review provenance the schema says belongs in the workflow artifact
  directory, not in `implementation-plan.md` (Minor schema-hygiene; front matter,
  section set, and order otherwise conform).

**Required changes:**

1. State that `ActiveRootsFor(kind)` returns only roots the `internal/packsource`
   classifier marks active, so there is one authority for "what is active."
2. Add retired-source rejection as an explicit pre-resolution gate in
   `LoadRuntimeCity`/`LoadRuntimeCityNoRefresh`, covering the single-retired-import
   case (not only duplicate-active).
3. Make the anti-glob guard decidable: route all pack/behavior enumeration through
   a typed `PackRoot`/`ActiveRootsFor` handle and have the scanner flag any raw
   `WalkDir`/`ReadDir`/`Glob` whose argument is not such a handle. Add a
   current-system inventory of the ~24 enumeration sites (including
   `cmd/gc/cmd_prompt.go:614`) to Current System.
4. Require every split asset's Core remainder to resolve and render with Gastown
   absent, with a Core-only test and a no-Core→Gastown-reference scanner row.
5. Make the rollback pin↔source coupling explicit and add a duplicate-free
   post-rollback assertion.

**Questions:**

- Is `ActiveRootsFor(kind)` a thin projection of classifier verdicts, or does it
  carry independent root logic?
- Does the production loader fail closed on a city whose only retired reference is
  a single stale Maintenance/Gastown import with no duplicate?
- For split assets like `mol-shutdown-dance`, does the Core remainder have any
  reference resolvable only from Gastown, and what test proves Core-only
  resolution?
