# Tomas Park - Claude

**Verdict:** block

Lane: implementation slicing, acceptance traceability, behavior-oriented tests,
migration gate ordering. The *behavior-test posture* is genuinely strong — the
design repeatedly forbids path/count/name assertions and demands
formula-composition and script-execution witnesses, and it correctly forbids
satisfying the no-Maintenance gate with test-only loaders. But the *slicing* is
not deterministically orderable, and at least one asset move can land without a
passing intermediate test state: the in-tree Core *role*-asset removal that the
whole migration hinges on is never assigned to a slice, and removing those
assets breaks Core formula composition (`mol-do-work` → `mol-polecat-*`) at an
unspecified point. On top of that, the document carries three non-reconciled
slice/commit decompositions plus a fourth "only binding source of truth"
artifact that does not exist yet, and behavior-manifest completeness is
self-certified. These trigger my primary red flag — an asset move landing
without a passing intermediate test state. Citations are by section heading;
the document has grown past 3500 lines and absolute line numbers drift.

I re-grounded the premise against the live tree: current coverage is exactly
the count/path/name style the design says must be rewritten —
`cmd/gc/controller_test.go:2235` asserts order *names* are present
(`gate-sweep`, `wisp-compact`); `cmd/gc/embed_builtin_packs_test.go:165-167`
asserts Maintenance order *file paths* exist. The design's behavior-witness
mandate is therefore the right target; the question is whether the slicing
enforces it commit-by-commit, and there it falls short.

**Top strengths:**

- Behavior-witness floor is executable, not aspirational. The attempt-3 *Docs,
  Tests, And Provider Integrity Gates* contract ("Path, count, and name
  assertions are insufficient. Replacement tests must prove formula composition,
  molecule step construction, hook target resolution, configured-agent/session
  loading, order ownership, prompt/template resolution, pack-relative script
  execution…"), the attempt-9 *Behavior Evidence Witness Floor*
  ("execution-level old witnesses require execution-level final witnesses"), and
  the attempt-14 *Behavior Evidence First Slice And Pilot Rows* old/final
  witness table together satisfy lane question 2.

- The no-Maintenance proof cannot be faked. Attempt-11 *Activation Pin Has No
  Inactive Loader Fiction* requires it to "run through
  `internal/systempacks.LoadRuntimeCity` or `LoadRuntimeCityNoRefresh`; a copied
  fixture, direct `config.Load*`, hidden bootstrap Core path, or test-only
  include list cannot satisfy it." This closes the most common gate-laundering
  hole and is the right answer to lane question 1's hardest case.

- Review-marked ownership audits are front-loaded and gated before either pin is
  consumed. *Cross-Pack Ownership Decisions*, attempt-4 *Rollout Gate Repairs*,
  and the *Open Questions* close-out ("later move/delete slices … may not defer
  ownership decisions … only local witness wiring named by an already-resolved
  row") answer lane question 3 at the sequencing level.

**Critical risks:**

- [Blocker] **The slice that strips existing Gastown role assets from in-tree
  Core is unassigned, and removing them breaks Core formula composition at an
  unsequenced point.** Requirements record that current Core "contains Polecat,
  Refinery, Witness, Mayor, and Gastown references in formulas and skills," and
  the migration map sends `mol-polecat-base/commit/report` to Gastown while
  `mol-do-work.toml` is `core-renamed` to "remove references to Polecat,
  Refinery, and Gastown formulas." *Core Asset Contents* confirms "Core must not
  own … Polecat formulas. Move `mol-polecat-base`, `mol-polecat-commit`, and
  `mol-polecat-report` to `gascity-packs/gastown`." But the Rollout assigns
  slice 1 to land them in *public* Gastown, slice 3 to "move Core assets to
  `internal/packs/core`" (a wholesale relocation per *Pack Ownership*: "Move
  Core pack assets from `internal/bootstrap/packs/core` to
  `internal/packs/core`"), slice 5 to "Move Core-owned *Maintenance* assets into
  Core," and slice 7 to "remove in-tree `examples/gastown/packs/*` sources" —
  none of which is the Core pack's own `mol-polecat-*` (it lives under
  `internal/packs/core`, not `examples/gastown/packs/`). No slice names the
  deletion of the Core `mol-polecat-*` formulas, the `gc-dispatch` SKILL split,
  or the `mol-do-work` de-referencing. The moment `mol-polecat-*` leaves Core
  while `mol-do-work` still references it, a formula-composition test fails —
  exactly the proof *System Pack Loading* / Testing mandate. The expiring
  role-token allowlist keeps the *scanner* green during the gap but cannot keep
  *composition* green, and the allowlist's expiry slice is itself undefined
  because the cleanup slice is undefined.

- [Major] **The Maintenance fold of same-named `dog` assets has no stated
  duplicate-free, test-green intermediate sequence, and "green intermediate
  commits" is false at the pin-switch step.** *Core Asset Contents* keeps the
  names ("rename only Gastown role names, not `dog` itself"; `mol-dog-jsonl` /
  `mol-dog-reaper` retained). Attempt-14 *Binding Slice Gates And Activation
  Commit Ordering* step A lands "Core-owned Maintenance behavior **behind
  existing active owners**" while Maintenance stays in
  `requiredBuiltinPackNames` through step D. But attempt-3 *Maintenance Runtime
  And Duplicate-Order Contract* ("no intermediate slice may expose duplicate
  active order definitions or two owners for the same script/formula behavior")
  and attempt-11 ("there is no inactive compatibility-asset mechanism in the
  production loader") mean a Core `mol-dog-jsonl` landed *behind* a still-active
  Maintenance `mol-dog-jsonl` is a duplicate active definition the normal loader
  rejects — that commit cannot be test-green. The same defect recurs at step C
  ("switch `PublicGastownPackVersion` to activation pin") *before* step D
  ("remove Maintenance"): C's named gate is a fixture / old-new binary matrix,
  not the production loader, so `make test-fast-parallel` / normal production
  load is not provably green at C, yet A–F are labeled "green intermediate
  commits." The only duplicate-free reading is a *per-asset atomic move*
  (delete-from-Maintenance + add-to-Core in one commit), which step-A's wording
  and the A→D separation contradict.

- [Major] **Three non-reconciled slice decompositions, plus a binding artifact
  that does not exist.** The document presents the attempt-9 *Slice Gates And
  Compatibility Fixtures* matrix (six slices from "public compatibility pin"),
  the attempt-14 activation commit ordering A–F, and the prose `## Rollout`
  slices 1–7 — then declares
  "`plans/core-gastown-pack-migration/slice-gates.generated.yaml` is the only
  binding source of truth for rollout gates" (attempt-14), an artifact produced
  only *during* implementation. No crosswalk maps these onto each other, and
  they disagree on the highest-risk slices: the attempt-9 matrix requires
  `make test-cmd-gc-process-parallel` **and** `make test-integration-shards-parallel`
  for the Core-loading/doctor slice (which migrates controller-reload and
  API/session config-load call sites), but prose Rollout slice 4 omits both and
  lists only the `-run`-filtered unit suite + goldens + `make test-fast-parallel`
  + `go vet`. Likewise prose slice 3 modifies
  `examples/gastown/precompact_hook_test.go` (Bootstrap Cleanup hidden-dependency
  inventory) but omits `go test ./examples/...` from its gate list, a tier the
  design treats as distinct. An implementer cannot deterministically order
  commits or know which suite gates which boundary — the core deliverable of my
  lane.

- [Major] **Behavior-manifest completeness is self-certified.** Traceability
  rests on attempt-3's "CI must fail when a moved, split, generalized, retired,
  or helper-dependent asset lacks a row" and attempt-10's
  `TestBehaviorManifestFresh` comparing generated-vs-checked-in. But both the
  generator output and the freshness comparison derive from the *same* discovery
  walk. A discovery false-negative — an un-walked helper reference or an asset
  the generator never visits — produces no row, is identical in generated and
  checked-in copies, and passes both checks silently. The attempt-17 "compare
  old-tree file moves/deletions/additions … current generated rows" clause
  gestures at an independent check, but no concrete VCS-level reconciliation is
  specified, and the ~11 hand-curated pilot rows do not cover the full surface.
  The central acceptance claim ("every moved asset has old+new witnesses")
  cannot be proven complete — a traceability hole directly in this lane.

- [Minor] **The `examples/` tree has an asserted-but-undemonstrated green window
  across slices 2–4.** *Examples And Docs* / Rollout slice 2 rewires
  `examples/gastown/pack.toml`, `city.toml`, and "nested legacy imports to
  public Gastown," but `examples/gastown/packs/maintenance` is not removed until
  the slice-5 activation tree and `examples/gastown/packs/gastown` plus
  `gastown_test.go` not until slice 7. `go test ./examples/...` is a required
  slice-2 gate, yet the design never specifies how the rewired example city
  coexists with stale local packs that still cross-import `../maintenance` and a
  `gastown_test.go` still asserting behavior against the local pack. The green
  state across this window is claimed, not shown.

**Missing evidence:**

- No slice is named for the in-tree Core role-asset removal (`mol-polecat-*`
  delete, `gc-dispatch` split, `mol-do-work` / `mol-shutdown-dance` /
  `mol-review-quorum` generalization), and therefore no expiry slice is defined
  for the role-token allowlist rows covering tokens currently in Core.
- No per-commit duplicate-free assertion for the Maintenance fold: the
  zero-duplicate test is specified only for the *end* state ("after Core and
  public Gastown are loaded together"), never for the intermediate candidate
  tree (post-A, pre-D) the rollout calls green.
- No numbered slice owns generation/freshness of the Gas City evidence
  artifacts (`behavior-manifest.generated.yaml`, `role-surface.generated.yaml`,
  `loader-inventory.generated.yaml`, `slice-gates.generated.yaml`, etc.).
  attempt-14 calls this "the first implementation slice," but the numbered
  Rollout (slice 1 = `gascity-packs`; slice 2 = Gas City pin) never assigns it.
- No crosswalk reconciling prose Rollout 1–7, the attempt-9 matrix, the
  attempt-14 A–F ordering, and `slice-gates.generated.yaml`; the document never
  states which wins where they disagree.
- No independent (non-generator) completeness check for behavior-manifest source
  discovery — nothing detects a discovery false-negative.

**Required changes:**

- Assign the in-tree Core role-asset removal to a named slice and sequence it so
  composition stays green: `mol-polecat-*` must be re-homed in public Gastown
  *before or in the same commit* that removes them from Core, and `mol-do-work`
  (plus any sibling referencing them) must be de-referenced in that same commit.
  State the expiry slice for every role-token allowlist row covering a token
  currently present in Core.
- Specify the Maintenance fold as an explicit *per-asset atomic move* (remove the
  asset from the Maintenance pack and add it to Core in the same commit) for
  every same-named `dog` asset, rewrite step-A's "behind existing active owners"
  wording accordingly, and reconcile step C/D: either make pin-switch and
  Maintenance-removal one atomic commit or mark A–C as non-deployable scaffolding
  where the production-loader / `make test-fast-parallel` gate is deferred to D,
  and stop calling them "green intermediate commits." Add a per-commit test
  asserting the intermediate candidate tree loads through
  `internal/systempacks.LoadRuntimeCity` with zero duplicate active definitions.
- Add one canonical slice/commit crosswalk mapping prose slices 1–7 ↔ attempt-9
  matrix ↔ attempt-14 A–F ↔ `slice-gates.generated.yaml`, state which is
  authoritative, and reconcile the per-slice gate lists with the contract
  matrices — at minimum add `make test-cmd-gc-process-parallel` +
  `make test-integration-shards-parallel` to the Core-loading/doctor slice and
  `go test ./examples/...` to the Core-extraction slice.
- Add an independent manifest-discovery completeness check that reconciles
  manifest rows against a VCS-level enumeration of moved/deleted/added pack
  assets (e.g. `git diff --name-status` over pack roots), so a discovery
  false-negative fails CI rather than passing silently.
- Add an explicit first Gas City slice (or extend slice 2 with stated ordering)
  that generates and freshness-tests all Gas City `plans/...generated.yaml`
  artifacts before any destructive change, and state that slice 2's packcompat
  consumes that slice's `behavior-manifest.generated.yaml`.
- Specify how `go test ./examples/...` stays green during slices 2–4 — delete or
  skip the local-pack behavior tests when the example city is rewired, or
  rewire/remove the stale local packs in the same slice rather than slices 5/7.

**Questions:**

- Is `slice-gates.generated.yaml` intended to *supersede* the prose Rollout and
  the attempt-9 matrix, or augment them? If supersede, why keep the prose
  Rollout as the most readable plan?
- For same-named maintenance assets, is the fold a per-asset atomic move
  (duplicate-free) or an add-to-Core-then-bulk-remove (duplicate window)? The
  design currently supports both readings, and only the former is test-green.
- Is the Core role-token scanner introduced in the first evidence slice with
  expiring allowlist rows for the existing Core role assets, and which commit
  makes the scanner green *without* allowlisting?
- Does slice 2's "rewire nested legacy imports" include rewriting
  `examples/gastown/packs/gastown/pack.toml`'s `../maintenance` import, and if so
  does the local `packs/gastown` then import the public pack or get deleted
  earlier than slice 7?
