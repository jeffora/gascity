# Oleg Marchetti - Claude

**Lane:** gastown-behavior-preservation-auditor (wave 1) — Gastown behavior
preservation, before/after inventory, trigger and notification continuity.
Reviewed the current `plans/core-gastown-pack-migration/requirements.md`
(`updated_at` 2026-06-09T17:23:58Z) against the three lane questions, cross-checked
the named support artifact
`plans/core-gastown-pack-migration/support/maintenance-asset-classification.md`
and the live Maintenance assets it classifies, and verified
`gc.mayor.requirements.v1` schema conformance.

**Verdict:** approve-with-risks

The behavior-preservation machinery is strong and now closes the two Major risks
this lane raised against the prior revision (`updated_at` 15:35:47Z). AC7's
before-state denominator is now a "frozen, named, version-pinned list of
supported Gastown templates and workflows" drawn **from AC6**, and AC6 freezes
its deterministic source snapshot "before any in-tree Gastown or Maintenance root
is deleted or isolated" — the prior husk-against-itself circularity and the
self-defining "supported" set are both gone. A new symmetric Core/non-Gastown
baseline guards the opposite failure (generic SDK behavior vanishing into
Gastown-only rows), and AC5 now carries a concrete worked side-effecting *consumer*
closure (`port_resolve.sh` → `dolt-target.sh`). All three lane questions are
answered with named, validated, gate-wired artifacts. One in-lane Major remains:
trigger/notification continuity across the deliberately-split Core→Gastown seam
is proven only at its two endpoints, not as a composed cross-boundary witness —
and that seam is exactly this lane's mandate. Schema conformance passes (front
matter, section order, W6H, happy/negative/edge example mapping, testable ACs,
Out-of-Scope, `Open Questions: None`); AC6/AC7/AC14 are correctly framed as
acceptance evidence, not inline implementation design.

**Top strengths:**
- AC7 enumerates the behavior surface explicitly — formulas, orders, scripts
  (including `assets/scripts`), prompts and prompt fragments, template variables,
  notification targets, requester/detector relationships, identity side effects,
  success/warning/failure/escalation paths, and recovery flows. This is a behavior
  inventory, not a file-move map, and fully answers lane question 2.
- The prior denominator gaps are closed and made bidirectional: AC7's denominator
  is now "frozen, named, version-pinned" **and** "from AC6"; AC6's snapshot is
  frozen *before* any in-tree root is deleted/isolated; AC6↔AC7 equality plus the
  new symmetric Core/non-Gastown baseline mean behavior cannot shrink between
  capture and proof, nor be silently re-owned in either direction (lane questions
  1 and 3).
- A concrete worked side-effecting closure now exists — the `port_resolve.sh` →
  `dolt-target.sh` helper dependency (edge case + AC5) — with "the migration fails
  if any surviving Core/provider/public-pack asset still consumes a retired
  Maintenance path." That answers lane question 1 for the consumer-closure class
  with a real exemplar, not just prose.

**Critical risks:**
- **[Major] Cross-boundary trigger/notification continuity is proven only at
  endpoints, not across the seam.** The hardest behavior to preserve in this
  migration is a side effect deliberately *split* across the Core→Gastown boundary:
  a Core-owned detector whose escalation/notification target is Gastown-owned.
  This is not hypothetical — `examples/gastown/packs/maintenance/assets/scripts/spawn-storm-detect.sh:59`
  (Class A / Core-bound generic crash-loop detection) escalates via
  `gc mail send mayor/`, and `reaper.sh:627,636` notify `mayor/` and nudge
  `deacon/ "DOG_DONE:..."`. After the split the detector is Core, the
  `mayor/`/`deacon/` target is Gastown, and the support doc's own "Escalation
  Override Contract" admits the re-homing mechanism (resolved pack-asset lookup /
  pack-script PATH / env var) is *unresolved*. AC7's executable verification is
  scoped to "the **external Gastown pack** ... triggers, routes, notifies ... with
  in-tree fallback disabled" — i.e. the Gastown endpoint in isolation — and the
  Core detector half is covered separately by AC9/the Core baseline. Nothing
  requires the *composed* path (Core detector → real override resolution →
  Gastown target actually fires) to be a single witnessed row. If "side-effecting
  witness checks" is read narrowly, a broken override-resolution mechanic passes
  both endpoint witnesses while the real end-to-end notification silently dies —
  the precise failure this lane exists to catch. (Fairness: "side-effecting witness
  checks" is ambiguous and *could* be intended as end-to-end; the fix is to make
  that explicit, not to assume it.)
- **[Minor] The Core/non-Gastown baseline denominator lacks the explicit
  frozen-source anchor the Gastown denominator now has.** AC7 ties the
  supported-Gastown denominator to AC6's frozen-before-deletion snapshot, but the
  "symmetric Core/non-Gastown baseline" is asserted without naming the same (or
  any) frozen source set. Generic-but-currently-in-Maintenance behavior (the Class
  A list: `gate-sweep`, `order-tracking-sweep`, `wisp-compact`, `cross-rig-deps`,
  `orphan-sweep`, `prune-branches`, `spawn-storm-detect` and their orders) must
  land in Core, not be dropped. AC6's "unrepresented active source files" failure
  and AC6↔AC7 equality likely cover this transitively, but the Core side's
  provenance is implicit where the Gastown side is now explicit — an asymmetry
  worth removing.

**Missing evidence:**
- A worked Example-Mapping witness for a *notification/escalation/trigger* path
  carried across the Core→Gastown boundary (`spawn-storm-detect`→mayor escalation,
  `reaper`→mayor/deacon notification, or `mol-shutdown-dance` warrant trigger). The
  two concrete examples present cover a static template (architecture fragment) and
  a script-sources-helper closure — neither is a trigger/notification seam, so the
  highest-risk continuity class still has no exemplar.
- Proof that the override-resolution mechanism re-homing stripped escalation
  actually fires from a Core caller. AC7 assumes the external Gastown pack can
  re-home `mayor/`/`deacon/`/`DOG_DONE` routing; the support doc flags the
  resolution mechanic itself as undecided, and no AC requires witnessing it.
- The explicit source set behind the Core/non-Gastown baseline's "before" state.

**Required changes:**
- Make the cross-boundary seam an explicit witnessed AC7 row class: where a
  Core-owned detector/script's notification or escalation is stripped and re-homed
  to a Gastown-owned target/override, require an *end-to-end* witness (Core detector
  → real override resolution → Gastown target fires) with in-tree fallback
  disabled, rather than two independent endpoint proofs. State plainly that
  "side-effecting witness checks" means the composed path, not endpoints.
- Add at least one Example-Mapping row for a notification/escalation/trigger path
  crossing the boundary (e.g. the `spawn-storm-detect`→mayor or `reaper`→deacon
  seam), alongside the existing static-template and helper-closure examples.
- Anchor the AC7 Core/non-Gastown baseline denominator to the same AC6
  frozen-before-deletion snapshot the supported-Gastown denominator already uses,
  so neither side's "before" set can be computed loosely or late.

**Questions:**
- For a split like `spawn-storm-detect` (Core detects, Gastown overrides escalation
  to `mayor/`), does AC7's "side-effecting witness check" exercise the composed
  seam, or only the Gastown endpoint in isolation?
- Is the override-resolution mechanism that re-homes stripped escalation
  (pack-asset lookup / pack-script PATH / env var, per the support doc) in scope for
  an AC7 witness, or deferred entirely to implementation?
- Does the Core/non-Gastown baseline draw its "before" set from the AC6 snapshot,
  or from a separately — and possibly later — computed inventory?
