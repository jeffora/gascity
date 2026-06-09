# Oleg Marchetti - Claude

**Verdict:** block

Reviewed `plans/core-gastown-pack-migration/implementation-plan.md` against
`requirements.md` (AC6, AC7, AC13, AC14) and the non-empty
`implementation-plan.schema.md`. Lane: Gastown behavior-inventory completeness,
execution-level witnesses, cross-repo packcompat, old→new traceability, and the
source-deletion gate.

Grounded against the live tree, not the document alone. The retiring trees under
`examples/gastown/packs/{maintenance,gastown}` carry a large behavior-bearing
surface (3 + 7 formulas, 9 + 1 orders, 11 + 13 scripts, 8 gastown prompt
templates, plus dog/witness/deacon/boot/mayor/refinery/polecat overlays), and
~40 of those files contain mail/nudge/requester/detector/notify/escalation
triggers. The old witnesses `examples/gastown/gastown_test.go` and
`maintenance_scripts_test.go` exist and cover **scripts** plus a subset of
formula behaviors at execution level — but prompt/template behavior and the
relational mail/nudge/requester→detector triggers are witnessed only by existence
checks (`TestPromptFilesExist`, `TestAllFormulasExist`,
`TestAllPromptTemplatesExist`). None of `internal/packevidence`,
`cmd/gc/pack_evidence.go`, `test/packcompat`, or the `support/` artifacts exist
yet, so the contract prose in this plan is the **entire** spec the decomposition
will encode; a weak clause is not caught later by any validator on disk.

I block because my lane's central one-way guarantee — "CI prevents deleting or
generalizing a behavior before the gascity-packs trace row + landing commit are
present, and the inventory enumerates *every* trigger" — is defeated by a
grounded, verified counterexample, and the binding CI rule is the weaker of two
contradictory granularity statements.

**Top strengths:**
- Evidence is generator-driven, freshness-checked, and diffed against a **Git
  historical baseline** (lines 159-206), so a file deleted from the workspace
  cannot make its legacy behavior invisible — the correct backbone for an old→new
  trace.
- The deletion gate is fail-closed and staged: **activation-pin packcompat
  rejects** any assertion resolving from `examples/gastown`,
  `.gc/system/packs/{gastown,maintenance}`, or a synthetic alias (lines 216-219);
  deletion cannot unblock until commit + pack digest + behavior-manifest digest +
  packcompat transcript are cited (lines 195-197); one-way slice boundaries are
  explicit (lines 795-808); and the two-repo release order forbids shipping a Gas
  City pin ahead of validated public behavior (110-112, Slices 1a→1c→2→5a).
- New-side witnesses must be **executable**, checked one-assertion-per-row against
  the exact pinned public pack (192-194, 208-213), and the planning generator must
  be promoted into checked-in implementation paths before deletion is scheduled
  (154-157).

**Critical risks:**
- **[Blocker] The binding completeness rule is per-asset, so the deletion gate
  passes while sub-asset triggers vanish.** Line 202 fails CI "if a moved, split,
  generalized, deleted, or helper-dependent **asset** lacks a row" (asset
  granularity), while Testing promises "one-row-per-**trigger** extraction" (line
  636). These contradict and the *enforceable* clause is the weaker one. Verified
  counterexample: `formulas/mol-shutdown-dance.toml` — which the plan itself
  splits (line 451: "generic stuck-session due process in Core, but Gastown
  detector/requester examples move to Gastown") — encodes in one file **6 steps,
  3 escalating health-check nudges (60s/120s/240s), ~5 distinct mail sends (PARDON
  ×3, EXECUTE_FAILED, DOG_DONE), a `gc.routed_to`/`dog` requester binding, and a
  PARDONED-vs-EXECUTE escalation branch**. Under rule 202 one row satisfies CI;
  packcompat then checks "one assertion per row" (213) = one assertion for the
  whole dance. The 120s-vs-240s escalation, per-attempt PARDON mails, and the
  EXECUTE_FAILED path can all be dropped with the gate green. This is the lane-1
  failure and the "existence-only" red flag, on the one-way deletion boundary.
- **[Blocker] The old-side witness floor is existence-level, so "preservation"
  degrades to "exists in new location."** Line 194 requires "one old witness and
  one new **executable** witness," but the old witness may be a "source assertion"
  (184-185) and the witness-kind enum admits "static scanner" / "manual proof
  transcript" / "approved removal record" (180-186) — none execution-level. For
  any behavior without a pre-existing test (much script/notification/escalation
  and all prompt-embedded behavior is untested today), a row can point at an old
  source line, prove only that the *new* side fires, and never compare the two.
  Nothing in the plan captures an old golden/transcript to diff against, so AC7's
  "success/warning/failure/escalation paths, and recovery flows" and the
  "semantic-equivalence assertion" (189-190) reduce to a human claim. My mandate is
  "the *same* trigger fires"; the plan proves "*a* trigger fires."
- **[Major] Row schema omits AC7's outcome-path and template dimensions.** The row
  fields (170-190) and generator walk (200-203) name
  trigger/requester/detector/mail-nudge/prompt-fragment/script-branch but have no
  field for outcome class (success/warning/failure/escalation) or recovery flow,
  and never name template *fragments/variables*. Grounded: the pack ships
  `template-fragments/{following-mol,propulsion,...}.template.md` and
  `agents/dog/prompt.template.md`, and `mol-shutdown-dance` is almost entirely
  escalation/recovery outcomes. AC7 (requirements line 113) requires both
  explicitly; the schema cannot express either, so they are silently droppable.
- **[Major] Compatibility-pin mode is provenance-blind — the FIRST gate (Slices
  1b/2) can pass on copied in-tree assets.** Lines 214-216 claim compatibility mode
  "proves the public pack does not depend on hidden fallback" while
  `examples/gastown/packs/{gastown,maintenance}` still exist, but — unlike
  activation mode (216-219) — it specifies **no** per-assertion resolution-provenance
  check. An assertion can resolve from the in-tree copy and pass, masking a broken
  or never-pushed pinned public pack. Normal CI also runs packcompat against "a
  validated ordinary remote cache" (209-210); if that cache is seeded from a local
  `gascity-packs` worktree and its digest computed from the same seed, the check is
  circular. The live-network step is only a "recorded" release-gate result (AC14,
  719), not wired as hard digest equality into the pin-coherence gate (343-344).
  This is the "consumes copied assets instead of verifying the pinned public
  checkout" red flag at the earliest gate.
- **[Major] AC6↔AC7 bidirectional traceability is required by both ACs but is not
  wired.** AC6 requires "a bidirectional link between source behavior IDs and AC7
  witnesses at behavior-row or call-site granularity"; AC7 requires "AC6-to-AC7
  bidirectional equality checks." The manifest row schema (170-190) has a "stable
  row id" but **no field referencing the AC6 asset-ledger behavior id**, and no
  validator enforcing ledger↔manifest row-level equality (both directions, failing
  on orphans) is named in Proposed Implementation, Data And State, or the AC17
  matrix. The artifacts are only sequenced (AC6→Slice 1a, AC7→Slice 1b, lines
  711-712).
- **[Major] Whole evidence classes are exempt from the witness floor.** The "one
  old + one new executable witness" rule (line 194) names only `moved, split,
  generalized, or deleted`; `retired-approved` and `external-prerequisite` rows
  (line 192) are exempt. A behavior can therefore be *retired* on an
  approved-removal record with no execution-level — or even consumer-closure —
  proof that nothing still depends on it, directly weakening the source-deletion
  gate this lane guards.
- **[Minor] Two behavior-preservation manifests, no coherence gate.** The plan
  generates `support/behavior-preservation-manifest.yaml` (164-165) AND consumes
  `gascity-packs/gastown/behavior-preservation.yaml` produced independently by the
  public task (133-134, 550). The pin-coherence gate (343-344) compares "a
  behavior-manifest digest" but never states the two are the same artifact or
  cross-validate. They can drift and break the trace.
- **[Minor] The deletion gate trusts citations, not re-executed proof.** Deletion
  unblocks once digests + transcript are "cited in the pin ledger" (195-197) — a
  citation references a past run; it is not re-run in the deleting change's CI. And
  Slice 7's "Must pass before merge" (808) lists only docs/help/tutorial/doctor
  goldens — it omits behavior-evidence freshness and activation-mode packcompat *at
  the exact point of source deletion*.

**Missing evidence:**
- **Frozen old baseline unanchored.** `--old-baseline <immutable-gas-city-commit>`
  (164) is a placeholder; nothing records where it is stored, proves it predates
  any deletion, or binds it to AC6's "deterministic source snapshot." If the
  baseline can advance past a deletion, the Git-baseline safety (205-206) is
  silently defeated.
- **No defined trigger-extraction contract per source-kind.** "One-row-per-trigger
  extraction" (636) names a test but no algorithm: how a "trigger" is identified in
  a TOML order vs a Markdown prompt vs a shell script, and whether extracted
  trigger *count* per asset is compared old-vs-new.
- **No worked end-to-end row.** With zero support artifacts on disk, one concrete
  sample row (old trigger → old witness → new owner → new path → new executable
  witness → public commit/digest → packcompat assertion), especially for the split
  `mol-shutdown-dance`, is the only way to prove the schema can express the chain.
- **No named AC6↔AC7 validator and no stated cross-manifest equality.**

**Required changes:**
- Reconcile 202 vs 636: make **per-trigger** row granularity the CI-enforced rule,
  and fail closed when an asset's extracted trigger count drops between the frozen
  baseline and the manifest (so a multi-trigger asset cannot collapse to one row).
  Add a fixture with a known trigger count (use `mol-shutdown-dance`) as the
  enforcement test.
- Add outcome-path-class (success/warning/failure/escalation) and recovery-flow
  fields to the row schema, add template-fragment/variable extraction to the
  generator walk (200-203), and map both explicitly to AC7.
- Require execution-level old witnesses (test, golden, or command transcript
  captured **at the frozen baseline**) for any moved/split/generalized/deleted row,
  with the new witness diffed against it; isolate source-assertion-only rows as a
  distinct lower-assurance class needing an approved written equivalence rationale;
  and **remove the `retired-approved`/`external-prerequisite` exemption** by
  requiring positive execution-level or consumer-closure evidence before retirement.
- Give compatibility-pin mode the same per-row resolution-provenance assertion
  (source+commit+subpath+pack-digest, never an in-tree path) that activation mode
  has, and require the CI cache digest to derive from at least one independent live
  fetch at the immutable commit, fed into the pin-coherence gate.
- Wire AC6↔AC7: add an AC6 asset-ledger behavior-id field to every manifest row and
  name the validator that fails on any ledger row without a matching manifest
  witness and any manifest row without a matching ledger id (both directions).
- Name one authoritative behavior-preservation manifest and make the other a
  digest-verified projection (or add an explicit bidirectional cross-manifest
  coherence check).
- Strengthen the deletion gate from citation-trust to proof-execution: re-run or
  freshness-verify the green activation-pin packcompat transcript at the exact cited
  commit in the deleting change's own CI, and add behavior-evidence freshness +
  activation-mode packcompat to Slice 7's must-pass-before-merge gates (808).
- Anchor `--old-baseline` to a checked-in frozen commit bound to AC6's source
  snapshot, and CI-validate that the generator ran against exactly that commit and
  that it predates any deletion.

**Questions:**
- Is a manifest "row" keyed per source asset or per behavioral trigger — and for
  the split `mol-shutdown-dance`, how does a single asset express which of its
  ~dozen triggers went to Core vs public Gastown?
- For untested old behavior, is a source-assertion/static-scanner old witness
  acceptable, or must a characterization test/transcript be captured at the frozen
  baseline so the new witness can be diffed against it?
- Which behavior-preservation manifest is authoritative — Gas City `support/` or
  `gascity-packs/gastown/` — and what gate keeps them coherent in both directions
  (and enforces AC6↔AC7)?
- Where is the frozen `--old-baseline` commit recorded, what binds it to AC6's
  source snapshot, and what proves the CI cache equals the pinned public checkout
  rather than a local copy?

**Schema conformance (bead-requested):** Front matter and the seven required
top-level sections appear in the exact order mandated by
`implementation-plan.schema.md`, and `Open Questions` is `None`. The only nit: the
inline `<!-- REVIEW: added per … -->` provenance comments sit close to the
schema's "must not append review-attempt summaries" rule — they are comments, not
sections, so they pass, but a strict reading may want them stripped. This is not a
basis for the block; the block is the behavior-evidence content above.
