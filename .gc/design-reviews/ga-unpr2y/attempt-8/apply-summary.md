# Apply Summary

**Workflow verdict applied:** `iterate`

**Global synthesis:** `.gc/design-reviews/ga-unpr2y/attempt-8/synthesis.md`

**Design document updated:** `internal/session/DESIGN.md`

## Changes

- Updated the design-review disposition header from Attempt 7 to Attempt 8.
- Added an `Attempt 8 Review Response` section with a review marker.
- Converted all Attempt 8 [Blocker] and [Major] findings into hard design gates for Slice 0 and later slices.
- Kept Slice 0 as the only approved next implementation step and explicitly made it non-mutating.
- Documented the attempt artifact path bug as unfixable from `internal/session/DESIGN.md` because it belongs to the design-review workflow formula.

## Required Next State

- `design_review.verdict=iterate`
- `design_review.workflow_status=iterating`
- No implementation or decomposition beyond non-mutating Slice 0 should proceed until the named artifacts, guards, parity files, diagnostics, vocabulary checkpoints, and proof commands physically exist and pass.

## Verification

- Saved `design-after.md`.
- Saved `design.diff`.
- No code tests were run; this attempt changed the design document only.
