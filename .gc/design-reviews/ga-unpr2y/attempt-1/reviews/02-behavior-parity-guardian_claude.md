# Natasha Volkov - Claude

**Verdict:** block

Lane: REQUIREMENTS scenario parity, regression prevention, characterization
tests, proof freshness. Scope reviewed: `internal/session/DESIGN.md` against
`REQUIREMENTS.md` and `internal/session/AGENTS.md`. The active global verdict is
already `block`; from the parity lane I concur. Slice 0 is non-mutating, so it
cannot itself ship a regression — but the design text does **not yet guarantee
that Slice 0 delivers provable, fresh parity**. I grounded this against the
working tree: the requirements ledger currently cites three test paths that
never existed in git history (below), the evidence-repair mechanism lacks a
same-behavior constraint, and one of my persona's named red flags (backlog
slices carry no scenario-row mapping) is triggered. Those are parity-proof
defects in the document under review, so I cannot clear it from this lane until
the load-bearing gates in Required changes land. The findings are specific and
fixable; this is a block to be lifted, not a rejection of the approach.

**Top strengths:**
- Characterization-first is structurally embedded, not just asserted: Refactor
  Rules step 3 captures characterization before step 6 moves a caller and step 7
  keeps old behavior absent owner-approved requirement change; the per-surface
  delegation matrix holds every non-adopter surface at "characterization only";
  the Atomic Command Contract requires "characterization tests for the current
  caller" plus "parity tests after delegation." This directly answers lane Q2.
- The design already diagnosed the exact proof-freshness rot my lane exists to
  catch: it names SESSION-RECON-002/003/006/007 as stale and forbids a row from
  being "green because a file exists" (DESIGN.md:175-179). I verified those four
  are genuinely broken, so the diagnosis is real, not theatre.
- Behavior-changing requirements edits are gated on a durable owner-approval
  artifact carrying the changed scenario rows, exact selectors, and approval
  record together (DESIGN.md:48-51) — the right defense against red-flag 3
  (ledger edited to fit the refactor).

**Critical risks:**
- **[Blocker] The requirements ledger already cites proof paths that never
  existed — proof the freshness gate is needed and not yet real.** Verified on
  this branch: `cmd/gc/scale_from_zero_test.go`,
  `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go`
  (the Evidence column for SESSION-RECON-002/003/006/007) have **no git history
  at those paths** — `git log -- <path>` returns nothing; they were never there.
  The cited commits (`a2b2da046`, `b5a7f3be3`, `dbda1e380`) are all real, so the
  behavior shipped but the test-path evidence is fabricated. SESSION-RECON-003
  cites *only* such a non-existent test (no commit fallback) → it currently has
  **zero executable evidence**. The design mandates repair, which is correct, but
  this proves rows reach the ledger green without their selector ever running, so
  Slice 0's gate must force the repaired selector to *execute and assert the
  original behavior* — see next item.
- **[Blocker] Evidence "repair" has no same-behavior constraint.**
  DESIGN.md:175-179 requires the repaired selector to execute *or* an
  owner-retirement record, but never requires the repaired proof to assert the
  **same product behavior** the original commit established. Risk: Slice 0
  re-points SESSION-RECON-006 to a passing-but-narrower test, closing the gate
  while silently shrinking the requirement. The owner-approval rule (line 48-51)
  catches *intentional* changes; it does not catch *accidental* behavior-
  narrowing during repair. With three fabricated paths already in the ledger,
  this is not hypothetical — it is the most likely way Slice 0 closes "green"
  while parity is not actually preserved. This is red-flag 3 through the side
  door.
- **[Major] Freshness enforcement is a manual "minimum proof command," not a
  proven always-on gate.** `TestScenarioParityFreshness` and siblings
  (DESIGN.md:191-193) do not exist yet (expected — Slice 0 builds them), but the
  design never states they run in the always-on unit tier / CI on every change.
  A freshness test that only fires when someone runs the minimum command goes
  stale exactly when it is needed — which is how three never-existed paths
  reached the ledger. Contrast the CI-enforced invariant the repo already trusts
  (`TestEveryKnownEventTypeHasRegisteredPayload`, present at
  `internal/api/event_payloads_coverage_test.go`); the parity validators need the
  same standing or they will rot identically.
- **[Major] Backlog slices 1-6 carry no scenario-row mapping in the design
  (named red flag #1).** The Backlog (DESIGN.md:654-679) names slices but not the
  `SESSION-*` rows each touches. The mapping is deferred to
  `SCENARIO_PARITY.yaml` (which maps scenario→surface, not slice→scenario) plus
  per-bead `referenced SESSION-* rows` metadata. The entry gate mitigates it, but
  a slice author can still under-map (cite 1 row when 3 are touched) with nothing
  in the design to check against. This must be closed before any behavior-moving
  slice is decomposed.
- **[Major] Reconciler-fact extraction (Slice 6) lacks an explicit
  characterization gate.** The Atomic Command Contract's characterization
  requirement attaches to *mutation*-moving slices. Slice 6 extracts read-side
  "narrow lifecycle eligibility facts" out of the most parity-fragile code
  (`cmd/gc/session_reconciler.go` — the SESSION-RECON + SESSION-WORK rows plus the
  idempotence invariant at REQUIREMENTS.md:70-71), yet its characterization
  baseline rests only on prose Refactor Rules. Read-side extraction from the
  reconciler needs the same hard characterization-baseline requirement as
  commands.

**Missing evidence:**
- No statement of the freshness status of the commit-cited work-release rows
  SESSION-WORK-001..004 (`a3a3b9fcf`, `47b580e9f`, `8068393d8`, `d565a34e2`).
  They cite `cmd/gc/session_beads_test.go`, which exists, but the mandatory-repair
  list stops at the four RECON rows, so a reader cannot tell whether the WORK
  rows' selectors were confirmed to execute or merely assumed fresh. These rows
  guard the "closing must not strand assigned work" invariant; their freshness
  must be asserted, not implied.
- No oracle-staleness rule for the gap between Slice 0 baseline capture and a
  late slice (e.g., Slice 5 runtime-start) landing weeks later.
  `SLICE0_BASELINE.md` and `SCENARIO_PARITY.yaml`'s "current oracle" are captured
  once; the design does not say the oracle is re-executed (not just re-checked
  for selector existence) when a dependent slice lands.
- No statement that proof selectors must be symbol/test-name based rather than
  line based. The design itself notes the worker-boundary migration overlaps
  `internal/api/session_resolution.go` (DESIGN.md:433-436); line-anchored
  selectors there will drift under that migration and silently zero-match.

**Required changes:**
1. Make Slice 0's close gate require **every active `SESSION-*` row** (not only
   the four named RECON rows) to either execute its proof selector against a real
   test/static selector or carry an owner-approved retirement/amendment record.
   Add an explicit clause to DESIGN.md:175-179: repaired evidence must assert the
   **same product behavior** the prior commit established unless an owner-approved
   amendment row is attached.
2. State that `TestScenarioParityFreshness`, `TestSlice0Contract`, and the other
   validators in the minimum proof command run in the **always-on unit tier**
   (no integration build tag), execute on every implementation-file change, and
   that the Slice 0 close gate fails if any is skipped or tag-gated.
3. Require proof selectors in `SCENARIO_PARITY.yaml` to be **test-name or symbol
   based, not line based**, with a freshness assertion that the named
   test/symbol resolves to a real, compiled, executed selector (zero-match =
   fail) — explicitly covering files touched by the in-flight worker-boundary
   migration.
4. Add a per-slice **scenario-row map** (slice → touched `SESSION-*` rows) to the
   Backlog or Authority/Entry-Gate section, and require the entry gate to
   cross-check a slice's `referenced SESSION-* rows` metadata against the
   `SCENARIO_PARITY.yaml` surface→scenario rows so under-mapping fails the gate.
5. Extend the explicit characterization-baseline requirement (currently scoped to
   the Atomic Command Contract) to cover **read-side reconciler-fact extraction
   (Slice 6)** and runtime start, naming the current `cmd/gc/session_reconciler.go`
   selectors that must be characterized before any logic leaves the reconciler.
6. Make the `RepairEmptyType` quarantine parity test (DESIGN.md:246-251) a **hard
   gate**, not an "or": the first adopter must prove with positive+negative
   fixtures whether a previously selectable empty-type session stays selectable
   after a separate repair, or carry an owner-approved amendment row if selection
   intentionally changes. This is the single highest behavior-parity risk in the
   plan and the current wording leaves the change-without-approval door ajar.

**Questions:**
- Why is SESSION-RECON-003 in the mandatory-repair list when REQUIREMENTS.md
  shows it citing only a test (no commit)? I verified that path never existed —
  confirming the row has no current evidence — so the design should state whether
  003's *behavior* is still desired (repair the selector) or being retired (owner
  record), since there is no commit to fall back on.
- Are SESSION-WORK-001..004's selectors confirmed to execute today, or assumed
  fresh because `session_beads_test.go` exists? If unconfirmed, they belong in
  the Slice 0 mandatory-reconciliation set alongside the four RECON rows.
- When Slice 0 captures the "current oracle," is it re-executed when a dependent
  behavior-moving slice lands later, or only checked for selector existence?
  Parity depends on the former.
