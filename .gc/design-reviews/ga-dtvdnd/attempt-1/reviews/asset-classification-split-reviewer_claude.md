# Hugo Bautista - Claude

**Lane:** asset ownership classification; split assets; review-marked decisions;
the file-by-file migration table. Re-grounded against the current
`plans/core-gastown-pack-migration/requirements.md` (`status: draft`,
`updated_at: 2026-06-09T17:23:58Z`, Open Questions = `None`), the read-only
`gc.mayor.requirements.v1` schema, and the draft
`support/maintenance-asset-classification.md` that feeds the AC6 ledger.

**Verdict:** approve-with-risks

The doc keeps the right structural posture for my lane: the file-by-file map is
deliberately lifted out of the requirements body (the schema forbids
implementation file assignments) into an external **validated asset-migration
ledger** (AC6), while keeping complete ownership a hard prerequisite for
implementation approval (Out Of Scope). Since the prior iteration two things
materially closed the behavior-loss risk in *both* directions: AC7 now carries a
symmetric Core/non-Gastown baseline (generic behavior cannot vanish into
Gastown-only rows) plus single-owner positive-evidence, and the Example Mapping
adds a concrete side-effecting closure (`port_resolve.sh` → `dolt-target.sh`,
AC5 + edge case). On that basis I independently downgrade last round's Major
("no classification rule is bound to AC6") to Minor: AC8 role-neutrality, AC9's
executor policy, and the Problem Statement's explicit three-class split
(SDK-required / generic-maintenance / Gastown-specific) now supply a *testable*
criterion, and AC7's bidirectional baseline backstops both misclassification
directions even where the rule is not named. The residual items are
auditability/determinism tightenings plus one concrete unresolved ownership
decision.

**Top strengths:**
- File-by-file completeness is a hard, behavior-based gate, not a bucket move
  (red flags #1/#3). AC6 fails on unrepresented active source files, unresolved
  `review` rows, duplicated/orphaned split behavior, phantom rows, basename
  collisions, stale source snapshots, and ledger-to-output drift; the W6H "how
  much" row pins the denominator (every active asset under the Core source root
  and the legacy Maintenance/Gastown roots) so completeness has a defined
  universe.
- Split outputs are explicit and two-sided (lane Q2). AC6 requires, per split
  row, *all* Core/Gastown/retired output paths plus the split boundary and
  stable behavior/sub-asset IDs; AC7 requires row- or call-site-level witnesses
  for every behavior-bearing row and a symmetric Core/non-Gastown baseline, with
  single-owner rows carrying "positive evidence that the other owner has no
  behavior to preserve." Neither wholesale-to-Gastown nor wholesale-to-Core can
  strip behavior silently.
- Review-marked resolution is a gate, not a hope (lane Q3). AC6 fails on
  unresolved `review` rows and Out-of-Scope bars implementation approval until
  AC6/AC7 "exist and pass." The draft `maintenance-asset-classification.md`
  demonstrates the classification is tractable (gate-sweep → Core, reaper →
  split, `mol-shutdown-dance` → Gastown) and surfaces precisely the rows AC6
  must resolve.

**Critical risks:**
- **[Minor] (downgraded from the prior Major) The classification *test* is not
  bound to AC6 as the ledger's governing rule.** AC6 mandates the outputs
  (owner, both successors, split boundary, rationale, behavior IDs) but never
  names the criterion the ledger applies. AC8 (the role-neutrality denied-token
  scan) is a Core-side *backstop*, not a classifier — an asset carrying no role
  tokens can still be generic-Core or Gastown-only with no forced, consistent
  call. Impact is bounded to auditability/determinism rather than behavior loss
  (AC7's bidirectional baseline catches both misclassification directions), but
  two reviewers can still classify the same asset differently with equally
  plausible rationales. Bind AC6 to the role-neutral-mechanism → Core /
  role-specific-behavior → Gastown / mixed → per-behavior-split rule and require
  each row to record which side of that test it falls on.
- **[Minor] The mechanism-vs-instance seam for shared primitives is implicit.**
  AGENTS.md makes formulas/molecules/orders Core primitives, while W6H says
  "workflows ... belong to Gastown." The resolution — the engine stays Core,
  role-bound definition files (`following-mol`, `mol-shutdown-dance`,
  `mol-dog-*`) move to Gastown by role-neutrality — is correct and inferable
  from AC8, but is not stated, so `following-mol`-style rows lean on author
  judgment rather than a stated call-site rule.
- **[Minor] A live ownership decision is unresolved while Open Questions =
  `None`.** Whether the retired `mol-dog-*` order names survive as
  public-Gastown compatibility aliases for the renamed Core/provider orders, or
  disappear, is left open in the support draft ("keep a public Gastown
  compatibility alias only if needed" for `mol-dog-jsonl` / `mol-dog-reaper`).
  This changes a split row's output set (does the asset have a Gastown alias
  output path or not). AC6's unresolved-`review`-row gate forces resolution
  before implementation, so it cannot slip silently — but it is a product
  decision the requirements neither decide nor list.

**Missing evidence:**
- The governing classification rule AC6 must apply (beyond W6H prose), and a
  requirement that each ledger row record its test outcome.
- The enumerated closed target-owner vocabulary and a definition of the per-row
  "fallback classification" field — AC6 names both but enumerates neither, and
  "keep in Core as-is" vs "keep in Core but strip role references" are different
  decisions with different proof obligations.
- An explicit statement that the generic-maintenance *asset corpus* itself
  (`gate-sweep`, `wisp-compact`, `orphan-sweep`, `cross-rig-deps`,
  `order-tracking-sweep` and their orders) is Core-destined — inferable from the
  three-class Problem Statement + AC9, but not anchored as a classification rule.
- The resolution of the `mol-dog-*` compatibility-alias decision.

**Required changes:**
1. (clears the residual classification risk) Bind an explicit, testable
   classification rule to AC6: role-neutral SDK/mechanism → Core; role-specific
   behavior (roles, workflows, prompts, commands, overlays, recovery,
   notification targets) → Gastown; mixed → per-behavior split. Resolve the
   shared-primitive seam (formula/molecule/order/dispatch *engine* stays Core;
   role-bound definitions move to Gastown) so `following-mol`/`mol-*` rows have a
   deterministic, checkable home.
2. Enumerate the closed target-owner vocabulary (e.g.
   `core`/`core-renamed`/`gastown`/`split`/`retire`/`review`) and define how the
   "fallback classification" field resolves, so a conditional decision cannot be
   recorded as definitive.
3. Resolve — or list as the single Open Question — the `mol-dog-*`
   public-Gastown compatibility-alias decision, since it determines split-row
   output paths.
4. (Recommended) Name the known hard split categories (dispatch skills,
   maintenance docs, architecture fragments, `following-mol`, command glossary,
   TDD-discipline content) at the product level to anchor the AC6 ledger against
   the assets most prone to orphaned or under-split behavior; today only the
   architecture-template and helper-closure cases are exemplified.

**Questions:**
- For shared primitives (formulas/molecules/orders/dispatch, `following-mol`),
  is the intended rule "engine stays Core, only role-specific definitions move to
  Gastown"? If so, state it as the governing split rule rather than leaving it to
  the ledger author.
- Is role-neutrality (AC8) intended to be the sole classification tie-breaker,
  or may a `review` row escalate back to a product Open Question when
  role-neutrality alone cannot assign an owner?
- Do the renamed Core/provider orders keep `mol-dog-*` aliases in public
  Gastown, or are the old names retired outright?
- When a conditionally-Core asset turns out to carry role references, does the
  ledger record `core-renamed` (strip and keep) or `gastown` (move)?

**Schema conformance (read-only, vs `gc.mayor.requirements.v1`):** Conforms. Six
top-level sections in required order; valid front matter (`phase: requirements`,
`status: draft`; no forbidden `*_file` keys or bead IDs); Example Mapping has
happy/negative/edge scenarios with concrete input + expected behavior +
evidence; Acceptance Criteria are testable and behavior-focused; Open Questions
is a bare `None`. Moving the file-by-file table into the AC6 ledger is the
correct resolution of the schema's prohibition on implementation file
assignments. The only in-lane schema tension is the `mol-dog-*` ownership
decision sitting unresolved beneath `Open Questions: None` (Required change 3).
