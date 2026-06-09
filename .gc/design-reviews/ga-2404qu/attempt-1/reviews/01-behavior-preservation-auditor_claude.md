# Nadia Volkov - Claude

**Verdict:** approve-with-risks

Lane: Gastown behavior inventory, cross-repo evidence chains,
requester/detector/notification continuity, no silent capability loss. Findings
below were verified against the live tree, not just the prose: the
`current_baseline` pin `sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b` (L2214)
exactly matches `internal/config/public_packs.go:11`, and every named high-risk
asset and the warrant coupling described below exist in source.

**Top strengths:**
- **Silent loss is structurally prevented, not just discouraged.** The Attempt-17
  "Generated Artifact Contracts And Independent Completeness" contract (L2580-2604)
  is the right invariant for my mandate: the behavior generator "cannot
  self-certify solely from the current tree" and must diff old-tree
  moves/deletions/additions, old execution transcripts, current rows, the pinned
  public checkout, and the test-migration map — and *any source move, deletion,
  removed test assertion, or changed operator output without a row fails CI*. That
  closes my top red flag ("inventory lists files without observable behavior
  tests"): an asset cannot evaporate while a generic row shows green, because
  completeness is checked against VCS history, not the post-move tree.
- **The cross-repo sequence is airtight and grounded.** The three-phase pin ledger
  (`public-gastown-pins.yaml`, L2520-2536) with mandatory old/new-binary
  transcripts per row, the A–F activation commit ordering where source deletion is
  last (L2263-2270), "activation pin switch + Maintenance removal are one candidate
  gate unless the ledger proves coexistence," and the explicit ban on embedded /
  synthetic fallback satisfying a public pin (L68, L875, L2958) collectively defuse
  my red flag "Gas City lands before the public Gastown replacement commit is
  available." The `current_baseline` sha matches the live `public_packs.go` value,
  so the ledger is grounded, not aspirational.
- **Witness rows demand execution, name final paths, and block on unknowns.** The
  witness-row floor (L1393-1414, L2654-2685) requires first-pass rows for branch
  pruning, Polecat formulas, shutdown-dance detectors/requesters, warrants, and
  notifications with named final public paths
  (`gastown/orders/prune-branches.toml`, `gastown/formulas/mol-polecat-*.toml`) and
  a `blocked:path-unresolved` state (L2668) that forbids moving a source whose
  destination is unknown. `requester`/`detector`/`notification`/`DOG_DONE` and
  configured-recipient empty-value handling are first-class witness classes tested
  by execution, not existence.

**Critical risks:**
- **[Major] The cross-pack warrant field contract is named as evidence but the
  round-trip is never required.** Verified in source: the Core-bound
  `mol-shutdown-dance.toml` (`core-renamed`) extracts exactly
  `{target, reason, requester, warrant_id}` from wisp metadata at runtime
  (`gc bd show … | jq '.[0].metadata'`; vars L52-65, extract step L76-78) and
  notifies `{{requester}}/` (`DOG_DONE`/`PARDON` mail, L93/131/167), while the
  *filers* stay in Gastown and emit that exact shape:
  `mol-deacon-patrol.toml:166-277` and `mol-witness-patrol.toml:453-454` both
  `--label=warrant --metadata '{"target":…,"reason":…,"requester":"deacon"|"witness",
  "gc.routed_to":"{{binding_prefix}}dog"}'`. The design names "warrant metadata"
  as an evidence class (L1613, L2660), but the shutdown-dance pilot row (L2296)
  requires only "old formula/prompt render plus route/request transcript" and "Core
  generic render plus public Gastown detector/requester preservation" — two
  *independent* renders. Independent witnesses with their own fixture data both
  pass even if the Core executor's extracted field names drift from the pinned
  Gastown filers' emitted shape. Because the coupling is runtime bead metadata,
  drift produces no load-time error: warrant-driven session termination silently
  stops firing. This is precisely "trigger/requester semantics weakened during
  generalization," and it is the one high-risk seam the otherwise-strong manifest
  machinery does not force to a round-trip.
- **[Major] Semantic-delta approval authority is undefined — the one sanctioned
  path to capability loss has no independent gate.** Equivalence is machine-proven
  and deletion is blocked without witnesses, so the *only* way behavior may legally
  shrink is an "approved semantic-delta record" (L858, L1410-1411, L2588). CI only
  fails when a delta "lacks an approved record," but the design never names *who*
  approves a behavior delta or what distinguishes an approved record from a
  self-asserted one. An implementer can weaken a notification target or drop a
  requester, write the delta row with owner/rationale/operator-impact, mark it
  approved, and pass CI. The visibility requirements make a loss loud, not gated.
  For a "no silent capability loss" mandate, an implementer-settable approval flag
  is the residual silent-loss vector.
- **[Minor] orphan-sweep carries incident-derived Gastown-format behavior but is
  not in the named first-pass witness list.** `orphan-sweep.sh` is classified
  generic `core` (L2796) and appears in no high-risk move list, pilot row, or
  Cross-Pack Ownership row. Yet it encodes documented production-incident handling
  for `Agent.QualifiedName()` import-binding dot-strip and pool-spawned ephemeral
  assignees, with a comment recording a regression where "an ephemeral assignee
  like gastown__polecat-gc-XXXXX got stripped on every tick" (L106-128). The
  ownership-aware scanner will force its `gastown__*` literals onto an allowlist
  (reviewer attention) and the test-migration map catches removed assertions, so
  this is caught *reactively*; it is not *proactively* named, so the first
  generator pass may not emit an execution witness preserving the
  qualified/ephemeral assignee coverage under the same trigger inputs.

**Missing evidence:**
- No statement that the warrant manifest's "final witness" must consume the Gastown
  filer's *actual emitted metadata* (round-trip), versus a Core-only synthetic
  warrant. Only the round-trip catches field-name drift across the seam.
- The design forces a required-vs-optional recipient decision in general (L1378-1379
  "required recipient fields fail preflight if empty or `/`"; L2675 empty-value
  tests) but never records shutdown-dance's `requester` class specifically. If
  generalization lets it default optional-empty, `DOG_DONE`/`PARDON` mail to
  `{{requester}}/` degrades to best-effort with no failing gate.
- No execution witness row ties orphan-sweep / cross-rig-deps city-qualified and
  ephemeral assignee handling to a test or an approved semantic delta.
- `test/packcompat`, `behavior-manifest.generated.yaml`, and the compatibility /
  activation public pins do not exist in-tree yet (expected for a design). The
  preservation guarantee is therefore contractual today; no individual preserved
  behavior is traceable to test evidence until slice 1 lands.

**Required changes:**
- Freeze `{target, reason, requester, warrant_id}` as a named cross-pack warrant
  contract in `behavior-manifest.generated.yaml`, and upgrade the shutdown-dance
  pilot row (L2296) from "route/request transcript" to a **round-trip** packcompat
  witness: a warrant filed by the pinned public Gastown deacon/witness patrol must
  be extracted and executed by the Core `mol-shutdown-dance` unchanged, including
  an empty/`/` `requester` negative case.
- Record shutdown-dance's `requester` class explicitly as **required**
  (preflight-fail on empty/`/`) in its manifest row so completion notifications
  cannot silently vanish.
- Name the semantic-delta approval authority and make delta approval an
  *independent* gate: specify a named design-review owner / human approver distinct
  from the implementing bead, and make the manifest CI treat "approved" as a
  reviewer-attested field, not an implementer-settable boolean.
- Add orphan-sweep and cross-rig-deps to the first-generator-pass witness list
  (L2659) with an execution witness preserving city-qualified (`gastown.deacon`)
  and ephemeral (`gastown__…`) assignee handling, or an approved semantic-delta row
  referencing the original regression.

**Questions:**
- Is the warrant manifest "final witness" required to be a cross-pack round-trip
  (Gastown filer output → Core executor input), or are two independent renders
  acceptable? Only the round-trip catches field-name drift across the seam.
- Who approves a semantic-delta/removal record, and is that approval auditable
  independently of the implementer who wrote the row?
- For orphan-sweep moving to Core: are city-qualified and ephemeral Gastown
  assignee formats preserved behavior (role-neutralized in name only) with a
  retained execution witness once the scanner forbids `gastown.*` literals in its
  tests, or is dropping them an approved semantic delta?
