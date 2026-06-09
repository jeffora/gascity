# Iris Kowalski - Claude

**Verdict:** block

Reviewed `plans/core-gastown-pack-migration/implementation-plan.md` strictly in
my lane: independently deployable slices, decomposition readiness, prerequisite
honesty, exact gates, and cross-repo sequencing/test coverage. The rollout
*architecture* is strong — slice-to-gate table, per-slice one-way boundaries and
rollbacks, de-batched landings, and a tiered test policy that does not lean on
fast unit tests alone. I block on a verified exact-gate inconsistency, the
absence of any named executable enforcer for the decomposition-readiness gate,
and the requirements' explicit approval constraint — not on the rollout design
itself. The fixes are narrow.

Grounded against the tree: **0 of 12** named support artifacts exist under
`plans/core-gastown-pack-migration/support/` (only `maintenance-asset-
classification.md` is present), and **0 of 4** public prerequisites
(`gastown/public-gastown-pins.yaml`, `behavior-preservation.yaml`,
`ownership.yaml`, `artifacts/packcompat/`) exist in the gascity-packs worktree.
Front matter is `status: draft` with `Open Questions: None`. The proposed
packages (`internal/{systempacks,packsource,packevidence,doctorfix}`,
`cmd/gc/pack_evidence.go`, `test/packcompat`) do not exist yet, which is correct
for a pre-decomposition plan — but it means the gate that is supposed to
*authorize* decomposition is itself absent.

**Top strengths:**
- The slice-to-gate table (lines 795-808) gives every slice a start gate, a merge
  gate, and a one-way boundary, and the per-slice prose (lines 724-791) pairs each
  with a concrete rollback. Slices are independently deployable with explicit
  revert/one-way boundaries (Q1).
- Landings are **de-batched**, defeating the "one fragile landing" red flag: pin
  adoption (2 / 5a), Core extraction (3), loader (4a), doctor mutation (4b),
  runtime-state (4c), Maintenance fold (5b), cache cleanup (6), and deletion+docs
  (7) are separate slices with separate gates. No single slice fuses pin change +
  source deletion + doctor mutation + docs + activation.
- Test policy defeats the "fast unit tests as only proof" red flag at the
  policy level: `make test-fast-parallel` + `go vet` after every slice and
  *additionally* `make test-cmd-gc-process-parallel` +
  `make test-integration-shards-parallel` for high-risk loader/doctor/runtime-state
  slices (lines 681-685), with Slice 4a's bypass scanner and the `test/packcompat`
  ordinary-remote-path install (211, 287-302) forcing production loaders rather
  than copied fixtures or `config.Load` shortcuts (Q3).

**Critical risks:**
- **[Major] The decomposition-readiness gate has no named executable enforcer,
  and the artifact it would check does not exist.** The plan restricts work to
  "prerequisite-producing decomposition" until AC6, AC7, AC14-AC17 artifacts
  "exist, validate, and are cited by immutable commit or checked path" (Summary
  26-27; Decomposition Readiness Gate 691-699), and AC17 is labeled "design gate
  before bead creation" (line 722). Requirements AC17 explicitly demands "at least
  one executable gate such as a make target, pre-commit entry, or CI job must
  enforce the matrix before decomposition." The plan never names that concrete
  gate, and `support/acceptance-proof-matrix.yaml` itself is absent — so the
  boundary lives only in prose and the gate-table "May start when" column. A
  decomposer trusting `status: draft` + `Open Questions: None` (line 826) could
  create behavior-changing beads with nothing mechanically stopping it. This is
  exactly the "status says decomposition-ready while required generators/commits
  do not exist" red flag, and it is grounded: every entry condition in the gate
  table ("May start when AC6 ledger generator exists", "when AC3 matrix and AC11
  diagnostics schema exist", "when AC15/AC16 cache schema proof exists") points at
  an artifact that does not exist.
- **[Major] Slice 2's start gate omits its dependency on Slice 1b.** The table
  (line 800) says Slice 2 "May start when AC15/AC16 cache schema proof exists," but
  Slice 2's job is to "Update `internal/config/PublicGastownPackVersion` to the
  public compatibility commit" (line 740) — a commit that exists only once Slice 1b
  lands "the compatibility pin plus packcompat transcript" (729-732). The table
  expresses direct slice→slice dependencies everywhere else (1b←1a, 1c←1b,
  5a←1c+4a, 5b←5a, 6←5b), so omitting 1b from Slice 2's gate is an inconsistency
  that, decomposed literally, schedules Slice 2 to adopt a pin that may not yet
  exist. The gate must read "1b compatibility pin is immutable **and** AC15/AC16
  cache schema proof exists."
- **[Minor] Per-slice merge gates don't carry the integration-test requirement
  the Testing prose adds.** The table's "Must pass before merge" for the loader
  slice 4a lists only "descriptor gate tests, runtime guard, bypass scanner"
  (unit/scanner), even though lines 681-685 say high-risk loader slices also run
  the sharded process/integration targets. 5a and 4c do name integration-level
  proofs (packcompat, old/new binary matrix, quiesce/old-binary), but 4a does not.
  A bead decomposed from 4a's table row alone could merge on unit/scanner proof.
  Fold the process/integration targets into the high-risk slices' merge-gate rows.
- **[Minor] Conditional readiness is not encoded in a structured field.** Front
  matter is `status: draft` and Open Questions is `None`, but the real state is
  "decompose prerequisites now, behavior-changing slices only after gates." The
  schema's `status` enum (`draft|questions|approved|blocked`) cannot express
  "conditionally ready," so the distinction rests entirely on a reader honoring the
  Summary and gate table. An explicit per-slice "decompose-now / blocked-until-
  <gate>" label would keep the two phases from being conflated.
- **[Minor] Slices 5b and 7 are the densest landings.** Slice 5b couples removing
  Maintenance from `requiredBuiltinPackNames`, moving Core-owned assets into Core,
  and consuming Gastown assets from the public pack; Slice 7 couples stale-source
  deletion with final docs/tutorial/doctor goldens. Each has a coherent rollback,
  so this is not a blocker, but each is a candidate for further splitting if its
  merge gate proves heavy.

**Missing evidence:**
- The exact command/target that enforces the AC17 acceptance-proof matrix "before
  bead creation" is not stated; AC17's row cites the (absent) support artifact but
  not the enforcing gate.
- Slice 2 says "Unsafe legacy import rewrites are disabled or routed through the
  mutation coordinator" (743-745), but the coordinator (`internal/doctorfix`) does
  not exist until Slice 4b, which lands *after* Slice 2. The plan does not make
  explicit that "disabled" is the only available option at Slice 2's point in the
  sequence.
- The rollback-from-new-to-old matrix row (818) says doctor-mutated manifests are
  "either readable by old binaries or release notes name explicit downgrade
  limits" — the criterion that decides which, and which slice proves it, is
  unstated.
- The source of truth for "gascity-packs branch/commit availability" that the
  readiness gate checks (tracked external bead? pin-ledger entry? manual AC14
  sign-off?) is unstated.

**Required changes:**
- Name the concrete executable gate (make target / pre-commit / CI job) that
  enforces the AC17 readiness matrix and blocks behavior-changing bead creation
  until AC6/AC7/AC14-AC17 artifacts exist and are cited by commit/path, and create
  the `support/acceptance-proof-matrix.yaml` it checks before this plan is treated
  as decomposable beyond prerequisite-producing beads.
- Add "1b compatibility pin is immutable" to Slice 2's "May start when" gate (or
  state explicitly that the AC15 proof subsumes 1b's landed compatibility commit).
- Fold the sharded process/integration targets into the high-risk slices' (4a/4b/
  4c/5a) "Must pass before merge" rows so the per-slice gate is not unit/scanner-
  only.
- State that Slice 2 must *disable* unsafe legacy import rewrites (the
  coordinator-routed option is unavailable until Slice 4b), so the sequencing is
  unambiguous.
- Make the "prerequisite-producing decomposition only vs full decomposition"
  boundary explicit at the slice level (label each slice decomposable-now or
  blocked-until-gate) so `Open Questions: None` cannot be read as full readiness.

**Questions:**
- Per requirements Out-of-Scope (lines 143-144), the plan "must not be approved
  before AC6, AC7, AC14-AC17 proof artifacts exist and pass," and none exist yet.
  Is this design review intended to approve only *prerequisite-producing*
  decomposition (Slices 1a/1b plus support-artifact generators), with a second
  review gating the behavior-changing slices once artifacts land? If so, that
  scoping belongs in the plan.
- What is the source of truth for "gascity-packs branch/commit availability" — a
  tracked external bead / pin-ledger entry the readiness gate checks, or a manual
  release-gate sign-off (AC14)?
- Does any slice's merge gate verify that the *previous* slice's one-way boundary
  has not been crossed prematurely (e.g., that no pin was consumed before Slice 2),
  or is ordering enforced only by the "May start when" column?

**Schema conformance (bead-requested):** Front matter and the seven required
top-level sections appear in the exact order mandated by
`implementation-plan.schema.md`, and `Open Questions` is `None`. The inline
`<!-- REVIEW: added per … -->` provenance comments brush against the "must not
append review-attempt summaries" rule but are comments, not sections, so they
pass. The substantive schema tension is the one above: the `status`/`Open
Questions` signal reads "decomposable" while the body and (absent) gates say
behavior-changing decomposition is blocked — the schema's `blocked` status /
`blocked:prerequisite` stop more honestly fits the current artifact state.
