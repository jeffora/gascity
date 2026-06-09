# Apply Summary

## Verdict

Global synthesis verdict was `block`, so this apply pass sets
`design_review.verdict=iterate`.

## Changes Applied

- Added `Attempt 8 Review Resolution Contracts` to
  `.gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Replaced any remaining path-only Core proof ambiguity with a typed
  `RequiredSystemPackParticipation` record, two fatal loader gates, widened
  production bypass coverage, Core collision policy, and narrowed
  `GC_BOOTSTRAP=skip` semantics.
- Added a required Go/asset role-surface migration inventory with concrete
  starting surfaces and the `dog` decision as configurable Core
  maintenance-worker compatibility data.
- Tightened public Gastown pin adoption so synthetic aliases and bundled bytes
  are disabled before `PublicGastownPackVersion` is consumed, with duplicate
  active-definition gates across intermediate and rollback states.
- Defined a single `gc doctor --fix` mutation coordinator with staged overlay,
  CST/span-preserving TOML edits, compare-before-rename publication, and
  failure-injection coverage.
- Required a central retired-source classifier API for config loading, imports,
  cache reads, materialization, prompt/template discovery, docs lint, and error
  rendering.
- Expanded behavior-manifest and packcompat evidence to include digests,
  generator metadata, notification/requester/detector semantics, authorship,
  provider filters, TOML defaults, `session_live` hooks, dog-field ownership,
  and exact pinned public-manifest consumption.
- Added slice-accurate bootstrap, registry, cache, and materializer disposition
  gates, docs/operator vocabulary release gates, and a workflow artifact guard
  for attempt-local persona synthesis outputs.

## Unfixable Items

None. All blocker and major findings from the attempt 8 synthesis were
addressed as design contracts, executable gates, or explicit workflow follow-up
tracking.

## Verification

- Ran `git diff --check -- .gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Saved `design-after.md`.
- Saved `design.diff` against `attempt-8/design-before.md`.
