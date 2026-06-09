# Apply Summary

## Verdict

`design_review.verdict=iterate`

Attempt 16 remained `block`, so this pass addressed all [Blocker] and [Major]
findings in `internal/session/DESIGN.md` and keeps behavior-moving
decomposition blocked.

## Changes Applied

- Added an Attempt 16 review response with one design response per blocker,
  major finding, and the attempt-path minor finding.
- Made Slice 0 mechanically gated by `SLICE0_CONTRACT.yaml`, current executed
  `SESSION-*` proof selectors, zero-match/skipped/build-tag checks, negative
  fixtures, and owner-approved amendment or retirement rows.
- Added `STORE_CAPABILITY_MATRIX.yaml` and required every mutating operation to
  choose exactly one enforceable fence strategy before behavior moves.
- Replaced file-name migration triggers with inventory-driven rows from
  `SESSION_BOUNDARY_SYMBOLS.yaml`, `API_CLI_ROUTE_INVENTORY.yaml`, and
  `WORKER_BOUNDARY_EXCEPTIONS.yaml`.
- Added shared-resolver sequencing, endpoint-specific `repair-needed`
  semantics, and anti-drift proof requirements for read-only and materializing
  target resolution.
- Reclassified broad target-classification fields as provisional and forbade
  flat optional shared/public contracts for non-first-adopter fields.
- Tightened session event proof to require closed-world event inventories,
  payload adequacy proof, public-wire obligations, idempotency, stale/duplicate
  behavior, and durable recovery ownership.
- Materialized runtime/reconciler boundary requirements for `RuntimeIntent`
  rows, pure-decider guards, runtime-missing anti-flap behavior, diagnostics
  mappings, large-city budgets, and `bdstore` subprocess accounting.

## Unfixable Here

- Persona synthesis source beads for attempt 16 still stamped output paths
  under `attempt-1`. That is workflow formula plumbing outside
  `internal/session/DESIGN.md`; this design records the limitation and uses
  `.gc/design-reviews/ga-unpr2y/attempt-16/synthesis.md` as the authoritative
  synthesis for this pass.

## Artifacts

- `.gc/design-reviews/ga-unpr2y/attempt-16/design-after.md`
- `.gc/design-reviews/ga-unpr2y/attempt-16/design.diff`
- `.gc/design-reviews/ga-unpr2y/attempt-16/apply-summary.md`
