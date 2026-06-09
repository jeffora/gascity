# Nadia Volkov - Claude

**Verdict:** approve-with-risks

**Top strengths:**
- **Silent loss is structurally prevented, not just discouraged.** The Attempt-17
  "Generated Artifact Contracts And Independent Completeness" contract is the
  right invariant for my mandate: the behavior generator "cannot self-certify
  solely from the current tree" and must diff old-tree moves/deletions/additions,
  old execution transcripts, current rows, the pinned public checkout, and the
  test-migration map — and *any source move, deletion, removed test assertion, or
  changed operator output without a row fails CI*. That closes my top red flag
  ("inventory lists files without observable behavior tests"): an asset cannot
  evaporate while a generic row shows green, because completeness is checked
  against VCS history, not the post-move tree.
- **The cross-repo sequence is airtight and grounded.** The two-pin
  (compatibility → activation) model, `public-gastown-pins.yaml` with mandatory
  old/new-binary transcripts per row, the "activation pin switch + Maintenance
  removal are one candidate gate unless the ledger proves coexistence" rule, and
  the "before any Gas City source deletion lands" preconditions collectively
  defuse my red flag "Gas City lands before the public Gastown replacement commit
  is available." The `current_baseline` sha still matches the live
  `internal/config/public_packs.go` value, so the ledger is grounded, not
  aspirational.
- **Witness rows demand execution, name final paths, and block on unknowns.** The
  Attempt-17 witness-row floor requires first-pass rows for branch pruning,
  Polecat formulas, shutdown-dance detectors/requesters, warrants, and
  notifications with named final public paths
  (`gastown/orders/prune-branches.toml`, `gastown/formulas/mol-polecat-*.toml`)
  and a `blocked:path-unresolved` state that forbids moving a source whose
  destination is unknown. `requester`/`detector`/`notification`/`DOG_DONE` and
  configured-recipient empty-value handling are all first-class witness classes
  tested by execution, not existence.

**Critical risks:**
- **[Major] The cross-pack warrant field contract is named as evidence but the
  round-trip is never required.** Verified in-tree: `mol-shutdown-dance.toml`
  (→ Core, `core-renamed`) extracts exactly `{target, reason, requester,
  warrant_id}` from wisp metadata at runtime (`gc bd show {{warrant_id}} | jq`;
  fields declared lines 52–65, read step lines 69–96), while the *filers* stay in
  Gastown and emit that exact `--metadata '{"target":...,"reason":...,
  "requester":...}' --label=warrant` shape: `mol-witness-patrol.toml` (L445–454)
  and `mol-deacon-patrol.toml` (L164–275). The design names "warrant metadata" as
  an evidence class (L1613) and "warrants" as a first-pass row (L2660), but the
  shutdown-dance pilot row (L2296) requires only an "old formula/prompt render
  plus route/request transcript" and "Core generic render plus public Gastown
  detector/requester preservation" — two *independent* renders. Independent
  witnesses with their own fixture data both pass even if the Core executor's
  extracted field names drift from the pinned Gastown filers' emitted shape.
  Because the coupling is runtime bead metadata, drift produces no load-time
  error: warrant-driven session termination silently stops firing. This is
  exactly "trigger/requester semantics weakened during generalization," and it is
  the one high-risk seam the otherwise-strong manifest machinery does not force
  to a round-trip.
- **[Minor] orphan-sweep is not in the named first-pass witness list despite
  carrying incident-derived Gastown-format behavior.** `orphan-sweep.sh` is
  classified generic `core` (design L2796) and appears in no high-risk move list,
  pilot row, or Cross-Pack Ownership row. Yet the script encodes documented
  production-incident handling for `Agent.QualifiedName()` import-binding
  dot-strip and pool-spawned ephemeral assignees (`gastown__polekitten-gc-...`),
  with an explicit comment recording a regression where "an ephemeral assignee
  like `gastown__polecat-gc-XXXXX` got stripped on every tick" (L106–128). The
  ownership-aware scanner will force its `gastown__*` literals onto an allowlist
  (good — reviewer attention) and the test-migration map catches removed
  assertions, so this is caught *reactively*; but it is not *proactively* named,
  so the first generator pass may not emit an execution witness preserving the
  qualified/ephemeral assignee coverage under the same trigger inputs.

**Missing evidence:**
- No statement that the warrant manifest's "final witness" must consume the
  Gastown filer's *actual emitted metadata* (round-trip), versus a Core-only
  synthetic warrant. Only the round-trip catches field-name drift.
- The design forces a required-vs-optional decision for recipients in general
  (L1378–1379 "required recipient fields fail preflight if empty or `/`"; L2165;
  empty-value tests L2675) but never records shutdown-dance's `requester` class
  specifically. If generalization lets it default optional-empty,
  `DOG_DONE`/`PARDON`/`EXECUTE_FAILED` mail to `{{requester}}/` degrades to
  best-effort with no failing gate.
- No explicit witness row tying orphan-sweep / cross-rig-deps city-qualified and
  ephemeral assignee handling to an execution test or an approved semantic delta.

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
- Add orphan-sweep and cross-rig-deps to the first-generator-pass witness list
  (L2659) with an execution witness preserving city-qualified (`gastown.deacon`)
  and ephemeral (`gastown__...`) assignee handling, or an approved semantic-delta
  row referencing the original regression.

**Questions:**
- Is the warrant manifest "final witness" required to be a cross-pack round-trip
  (Gastown filer output → Core executor input), or are two independent renders
  acceptable? Only the round-trip catches field-name drift across the seam.
- For orphan-sweep moving to Core: are city-qualified and ephemeral Gastown
  assignee formats preserved behavior (role-neutralized in name only) with a
  retained execution witness once the scanner forbids `gastown.*` literals in its
  tests, or is dropping them an approved semantic delta?
