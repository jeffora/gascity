# Avery McAllister - Claude

**Verdict:** block

Lane: external Gastown ownership, Maintenance retirement, Core/Gastown split
completeness, and the public pack contract. I reviewed the current revision
(`updated_at: 2026-06-09T08:40:42Z`, Attempt-17 contracts) and re-grounded every
load-bearing claim against both repos rather than trusting prose: the in-tree
`examples/gastown` sources, Gas City's `internal/config.PublicGastownPackVersion`,
and the pinned `gascity-packs` checkout itself.

The split *architecture* is strong and most of my lane is satisfied in principle:
behavior preservation is a blocking, execution-witnessed precondition; ownership
audits are front-loaded to slice 1 before either pin is consumed (3377-3388); and
the no-Maintenance packcompat gate is named and executable (2922-2930). But two
prose contracts that gate my lane directly contradict each other, and the one
they gate — slice-1 "Dog prompt fragments" ownership (3276) — cannot be decided
as written without either losing Gastown behavior or re-introducing Gastown role
prose into Core, which is exactly my red flag. The contradiction is concrete and
fixable, but until it is resolved the design's own precondition for consuming
either public pin is unmet.

A grounding note that frames the rest: the value Gas City currently pins,
`PublicGastownPackVersion = sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b`, is the
`gascity-packs` commit literally titled **"fix(gastown): rely on implicit
maintenance pack."** At that commit the public pack (a) still documents an
implicit maintenance/core layer in `gastown/pack.toml:5`
("The host Gas City runtime supplies the implicit maintenance/core utility"),
`gastown/formulas/mol-deacon-patrol.toml:30`, and
`gastown/agents/deacon/prompt.template.md:25`; and (b) does **not** yet own
`prune-branches`, `mol-polecat-base`, or `mol-polecat-commit`. That is correct
for a *compatibility* pin, and the design says so (2728-2730). It also means lane
Q1 and Q3 are satisfied only on paper, at an *activation* pin that does not exist
yet, against a behavior manifest that has not been generated. The design's gates
are the right shape; my block is about the specific contracts those gates depend
on.

**Top strengths:**
- Split completeness is auditable by construction. The requirements' file-by-file
  migration map, the Cross-Pack Ownership Decisions table (3266-3282) with
  per-row evidence, and the rule that slice 1 resolves `mol-review-quorum`,
  provider overlays, Dog prompt fragments, Polecat formulas, branch pruning,
  shutdown-dance examples, review checks, host-Core worker refs, and
  role-theme/tmux APIs **before** either pin is consumed (3377-3388) is a direct,
  enforceable answer to lane Q1.
- Lane Q3 has a named, executable answer:
  `go test ./test/packcompat -run TestPinnedPublicGastownBehavior` in
  no-Maintenance production-loader mode through
  `internal/systempacks.LoadRuntimeCity` (2922-2930, 3423-3441), backstopped by a
  `test-migration.generated.yaml` map that forbids removing any
  `gastown_test.go` / `maintenance_scripts_test.go` function without a named
  replacement or approved-removal row (2916-2919). With ~196 in-tree test
  functions at stake (70 in `gastown_test.go`, 126 in
  `maintenance_scripts_test.go`), an enforced map is the right mitigation for my
  red flag.
- The host-Core contract is clean on the import axis: public Gastown depends on
  host Core by contract, never imports Core or Maintenance (2459-2462), Core is
  inserted before Gastown in layer order (382-395), and `requiredBuiltinPackNames`
  requires Core plus provider packs only (2986). Maintenance is retired, not
  recreated as a system pack (2769-2773).

**Critical risks:**
- **[Blocker] The dog-prompt preservation contract contradicts itself, blocking
  slice-1 Dog-prompt-fragments ownership.** The ownership table (2470) lists
  "Gastown-specific dog prompt wording, ... requester/detector behavior, and
  notifications" as **public-Gastown-owned**, with binding rule "Patch host Core
  fields or own Gastown prompts." But the Gastown Behavior Preservation section
  (2897) restricts Gastown patching of the Core `dog` agent to "theming or
  work_dir behavior **only**," and the host-Core model makes `dog` a Core-owned
  agent (1606-1607, 2468). These cannot both hold: prompt *wording*,
  requester/detector prose, and notifications are not theme/work_dir fields, so
  "patch host Core fields" cannot carry them; and "own Gastown prompts" for a
  Core-owned `dog` agent is undefined — it implies either patching the prompt
  body (forbidden by 2897) or defining a second `dog` agent (a fatal duplicate
  per 1054-1059). The design does name "Gastown prompts or formulas" as the home
  (2884), but for the *dog* prompt specifically that only works if the stripped
  requester/detector/notification prose moves into **other** Gastown assets
  (deacon/witness prompts, Gastown formulas), which changes the acting role and
  is a semantic delta requiring an approved manifest record — a resolution the
  design never states for this asset. Until the design picks one mechanism and
  spells it out, the "Dog prompt fragments" ownership row (3276) that slice 1
  must decide before either pin is consumed (3377-3388) is undecidable, and
  "Gastown does not lose behavior" (requirements.md:228-235, 280) is
  unimplementable for the dog prompt without putting Gastown role prose back into
  Core.
- **[Major] Pack Ownership still authorizes deleting in-tree Gastown behavior
  tests before public coverage exists.** 2777-2784 says pack-specific tests from
  `examples/gastown/gastown_test.go` can move to `gascity-packs` **"or"** be
  "replace[d] with Gas City tests that assert init and import wiring only." That
  "or ... wiring only" clause is precisely my red flag (removing behavior tests
  before equivalent public-pack coverage exists) and contradicts the strict
  mapping in 2916-2919 and 3335-3351. The precedence rule (2603-2604, "prose
  sections are advisory when they conflict with a generated row") does defuse
  this to a required edit rather than a standalone blocker — the generated
  test-migration artifact wins — but leaving a contradictory normative sentence
  in the section an implementer reads first is an avoidable hazard in the exact
  area my lane guards.
- **[Major] Active public-pack `dog`/worker routes are not yet de-roled at the
  pinned commit, and the design states the route rule only as generic witness
  rows.** The host-Core worker contract (1609-1623) requires Core maintenance
  routing to resolve the configured worker key, and the witness rows cover
  `pool = "dog"`, hardcoded worker targets, and `DOG_DONE` nudges (2672-2677).
  But the activation requirement that public Gastown's active prompts, formulas,
  orders, and warrant metadata must not target the host worker via
  `{{binding_prefix}}dog` / `name = "dog"` unless `dog` is a *public-Gastown-owned*
  agent is implied, not stated as an ownership rule. Given the pinned pack still
  carries implicit-maintenance assumptions, this needs to be an explicit boundary
  rule, not only a generic witness obligation, or the activation pin can pass
  path/row checks while still routing Gastown behavior at a host-Core compatibility
  name.
- **[Minor] Two concrete inaccuracies in the public-Gastown asset map undercut
  the "source-derived" auditability claim.** (1) Witness rows name
  `gastown/scripts/prune-branches.sh` (2663), but the pinned public pack's
  established convention is `gastown/assets/scripts/` (e.g.
  `gastown/assets/scripts/worktree-setup.sh`, `.../checks/...`); the design's own
  example path does not match the pack layout it must land in. (2) The design
  (2666, 2812) and requirements (249) name `mol-polecat-report.toml` as a Core
  asset to move to Gastown, but no such file exists in
  `internal/bootstrap/packs/core/formulas/` (only `mol-polecat-base.toml` and
  `mol-polecat-commit.toml`). A source-derived generator would simply emit no row
  for it, but the prose over-specifies a nonexistent asset, which will mislead an
  implementer reading the witness/preservation lists.

**Missing evidence:**
- The exact `gascity-packs` activation commit (and its `public-gastown-pins.yaml`
  `activation` row): immutable commit, the `gastown/...` final paths for
  `prune-branches`, `mol-polecat-base|commit`, and the relocated shutdown-dance
  detector/requester prose, and the behavior-manifest digest. None of this exists
  at the pinned compatibility commit, so lane Q1/Q3 are unproven today.
- A worked example showing how the stripped Maintenance-`dog` requester/detector/
  notification prose is preserved in Gastown: which asset receives it (dog patch,
  deacon/witness prompt, or a Gastown formula), the resulting actor, and the
  semantic-delta record if the actor changes.
- A wording-lint fixture that fails on the exact strings present at the pinned
  commit — `gastown/pack.toml:5` "implicit maintenance/core utility",
  `mol-deacon-patrol.toml:30`, `deacon/prompt.template.md:25` — proving the
  activation pin removes them, not just that imports are absent.

**Required changes:**
- Resolve the dog-prompt contradiction (2470 vs 2897) with one explicit
  mechanism. State whether Gastown preserves stripped dog-prompt requester/
  detector/notification behavior by (a) a host-Core prompt-body patch (then 2897
  must widen beyond theme/work_dir and define collision-free prompt patching), or
  (b) relocating that prose to named Gastown-owned assets (deacon/witness prompts
  or Gastown formulas) with a required semantic-delta manifest row for the changed
  actor. Make the "Dog prompt fragments" ownership row (3276) cite that mechanism.
- Delete the "or ... assert init and import wiring only" clause from 2777-2784 and
  replace it with the strict contract: public-pack behavior tests must move to or
  be recreated in `gascity-packs/gastown`; Gas City may retain only additional
  init/import wiring tests after `test-migration.generated.yaml` proves equivalent
  public coverage or approved removals.
- Add an explicit public-Gastown boundary rule (not only witness rows): any active
  public Gastown patch, formula pool, `gc.routed_to`, warrant metadata, mail/nudge
  target, prompt instruction, or command that targets the Core maintenance worker
  must resolve the configured host-Core binding, never `{{binding_prefix}}dog` or
  `name = "dog"`, unless `dog` is a public-Gastown-owned agent — tested with
  default, renamed, and omitted maintenance-worker fixtures.
- Correct the asset-map facts: use the pack's `gastown/assets/scripts/...`
  convention for moved scripts (or state the chosen convention once and apply it),
  and drop or annotate `mol-polecat-report.toml` since it does not exist in Core.

**Questions:**
- After activation, is every `dog` surface in public Gastown strictly host-Core
  compatibility data (no public-owned `dog` agent), or may public Gastown own a
  `dog` agent? The dog-prompt and route mechanisms above resolve differently
  depending on the answer, and the duplicate-definition rule (1054-1059) makes a
  public-owned `dog` agent collide with host Core unless explicitly aliased.
- Which `gascity-packs` test package is the canonical home for moved Gas City
  behavior tests, and does `test/packcompat` invoke it or only consume its
  generated manifest and the pinned assets? Lane Q3's guarantee depends on which.
- Will the activation pin's `gastown/pack.toml` and deacon assets have the
  implicit-maintenance comments removed as a *gated* condition, or is that left to
  best-effort cleanup? At the current pin they are still present.
