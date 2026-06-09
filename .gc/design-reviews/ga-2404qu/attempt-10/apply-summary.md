# Apply Summary

## Verdict

Global synthesis verdict was `block`, so this apply pass sets
`design_review.verdict=iterate`.

## Changes Applied

- Added `Attempt 10 Review Resolution Contracts` to
  `.gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Replaced path/provenance-based Core proof with an importable
  `internal/systempacks` boundary, strict pre-resolution file-set integrity, and
  typed post-resolution `RequiredSystemPackParticipation`.
- Scheduled compatibility and activation public Gastown pins through
  `public-gastown-pins.yaml`, with activation required before Maintenance
  removal, source deletion, or registry/cache cleanup.
- Settled the host-Core/public Gastown model: public Gastown does not import
  Core, and Core maintenance-worker routing resolves from configuration rather
  than Go role constants.
- Centralized retired-source handling in an `internal/packsource` classifier
  covering load, install/check, lock/cache validation, packman, materialization,
  prompt/template discovery, docs lint, and generated references.
- Replaced direct `gc doctor --fix` writes with a coordinator-owned mutation
  plan covering preflight, staged publish, live-controller refusal, concurrent
  doctor detection, generated-source provenance, runtime-state migration, and
  failure-injection fixtures.
- Made behavior manifest and exact public replacement evidence the first
  implementation gate, including execution-level witnesses for channels,
  recipients, authorship, runtime state, scripts, prompts, Dog flows, Polecat,
  branch pruning, commands, doctor checks, providers, and examples.
- Finalized `GC_BOOTSTRAP=skip` as retired for production behavior and moved
  Core fidelity proof to `internal/systempacks`.
- Added generated artifact ownership/freshness contracts for behavior,
  role-surface, wording, and old-test/new-test mapping artifacts.
- Moved `examples/gastown` rewiring, docs, runtime-state guidance, tutorial
  goldens, generated references, and release gates into the slices that change
  the corresponding behavior.
- Added an attempt-local design-review artifact guard requiring persona output
  paths or consumed-source manifests to match current attempt metadata.

## Unfixable Items

None. All blocker and major findings from the attempt 10 synthesis were
addressed as design contracts, explicit rollout choices, executable gates, or
workflow follow-up tracking.

## Verification

- Ran `git diff --check -- .gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Saved `design-after.md`.
- Saved `design.diff` against `attempt-10/design-before.md`.
