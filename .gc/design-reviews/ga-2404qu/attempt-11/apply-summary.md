# Apply Summary

## Verdict

Global synthesis verdict was `block`, so this apply pass sets
`design_review.verdict=iterate`.

## Changes Applied

- Added `Attempt 11 Review Resolution Contracts` to
  `.gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Chose the host-Core/public Gastown ownership model: Core owns the
  `maintenance_worker` binding, public Gastown patches it symbolically, and
  omitted/renamed workers have explicit diagnostics and tests.
- Removed reliance on inactive compatibility assets by requiring compatibility
  pins to omit colliding active assets and activation proof to run through the
  normal production loader after Maintenance is no longer included.
- Made required Core loading deny-by-default with uniform pre-resolution
  file-set integrity and post-resolution `RequiredSystemPackParticipation`
  fatal gates, plus a generated production loader inventory.
- Expanded role-surface inventory to production Go, API/dashboard/generated
  artifacts, provider packs, examples, prompts, scripts, overlays, docs, and
  moved asset names.
- Added an enforceable `FixIntent` and `doctor.MutationCoordinator` protocol,
  including two-phase reporting, directory-fd advisory locking, post-publish
  reruns, and non-destructive runtime-state copy semantics.
- Made `public-gastown-pins.yaml` authoritative for pin/cache/registry/offline
  behavior and separated active materialization from retired-source diagnostics.
- Bound docs, wording, generated schemas/references, moved assets, and
  extension-agnostic stale-path scans to the same release gates as the behavior
  they describe.

## Unfixable Items

None. All blocker and major findings from the attempt 11 synthesis were
addressed as design contracts, explicit ownership choices, executable gates, or
slice-level proof requirements.

## Verification

- Ran `git diff --check -- .gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Saved `design-after.md`.
- Saved `design.diff` against `attempt-11/design-before.md`.
