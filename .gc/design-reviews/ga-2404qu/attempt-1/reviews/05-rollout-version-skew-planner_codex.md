# Yuki Hayashi - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design has the right two-phase public-pack rollout shape: a compatibility pin while bundled Maintenance is still active, then an activation pin consumed by the same candidate tree that removes Maintenance and runs no-Maintenance production-loader packcompat.
- Pin integrity is materially specified. Public Gastown must resolve through the ordinary remote source at an immutable commit, with subpath identity, manifest digest, active-asset digest, stale synthetic-cache rejection, ordinary remote-cache behavior, and old/new binary evidence.
- The activation and rollback ledger at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:2523` is strong: it requires `current_baseline`, `compatibility`, and `activation` rows, old/new binary evidence for each row, offline/cache results, rollback class, and one-way-boundary notes.

**Critical risks:**
- [Major] The final rollout slice says `public-gastown-pins.yaml` is written with only "compatibility and activation records" at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:3380`, but the stronger ledger contract requires three rows: `current_baseline`, `compatibility`, and `activation` at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:2523`. That baseline row is not optional: without it, reviewers cannot prove what fresh init and old binaries did before the migration or compare rollback behavior after the new pins.
- [Major] The final release compatibility matrix collapses compatibility and activation into a single "new pack" category at `.gc/design-review-inputs/core-gastown-pack-migration/design.md:3465`. Earlier sections correctly distinguish old binary plus compatibility pin from old binary plus activation pin, including the possibility of a named one-way boundary. The final matrix should keep those phases separate so an implementer does not accidentally certify compatibility-pin behavior and then reuse that evidence for activation.
- [Minor] The design implies, but does not explicitly state, the fresh-init behavior during the interval after the `gascity-packs` preservation branch lands but before Gas City updates `PublicGastownPackVersion`. The intended state appears to be: fresh `gc init --template gastown` still pins the current baseline until the Gas City compatibility slice lands; manual imports of the new public branch are outside supported migration proof. This should be written as an interval table to prevent a mistaken assumption that landing the public pack branch changes Gas City behavior by itself.

**Missing evidence:**
- No concrete `public-gastown-pins.yaml` exists yet, so this review could not verify actual commit ids, active-asset digests, old-binary transcripts, or offline/cache transcripts. That is acceptable only because the design blocks pin consumption until the ledger validator is green.
- The design does not yet name the baseline Gas City binary/build artifact used for old-binary evidence beyond the historical `v1.2.1` references in earlier sections. The generated ledger should make that exact binary source and digest explicit.

**Required changes:**
- Align the final rollout section with the ledger contract: require `current_baseline`, `compatibility`, and `activation` records in `public-gastown-pins.yaml`.
- Split the final release compatibility matrix by public-pack phase: current baseline, compatibility, activation, and rollback after doctor fix. Make old binary plus activation either a proven supported row or an explicit one-way boundary before `PublicGastownPackVersion` switches to activation.
- Add a short fresh-init interval table: before Gas City pin update, after compatibility-pin adoption, after activation-pin adoption, and rollback. Each row should state the pin selected, whether Maintenance is still active, whether the state is deployable, and which packcompat mode proves it.
- Make the Gas City `PublicGastownPackVersion` update conditional on a validator that reads the exact public-pack checkout/cache path and rejects a missing phase row, missing digest, or stale synthetic-cache proof.

**Questions:**
- If old binary plus activation pin is not supported, where does the operator see the one-way boundary before the activation pin is published: release notes only, `gc doctor`, or a packcompat/ledger validation failure?
- Does the compatibility pin omit all active assets that would collide with bundled Maintenance, or are any rows approved semantic deltas? The active-asset digest and duplicate matrix should make that distinction visible.
