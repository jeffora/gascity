# Apply Summary

## Verdict

Global synthesis verdict was `block`, so this apply pass sets
`design_review.verdict=iterate`.

## Changes Applied

- Added `Attempt 7 Review Resolution Contracts` to
  `.gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Defined typed, content-backed `RequiredSystemPackParticipation` with fatal
  pre-resolution integrity and post-resolution participation gates.
- Added a required Go and asset role-surface migration table, including the
  `dog` ownership decision as configurable Core maintenance-worker target.
- Tightened public Gastown pin adoption so it must use ordinary remote
  resolution, exact commit/subpath validation, and duplicate active-definition
  tests before source deletion.
- Replaced whole-file doctor mutation with a scoped byte-preserving TOML edit
  planner and a composite failure-atomic migration coordinator.
- Added a central retired-source classifier and prompt/template discovery
  containment rules for stale Maintenance/Gastown directories.
- Expanded behavior-manifest and packcompat evidence requirements with row
  digests, evidence classes, test-function/subtest mapping, and the required
  fixture matrix.
- Made registry, cache, materializer, bootstrap, and docs-vocabulary gates
  slice-specific and executable.

## Unfixable Items

None. All blocker and major findings from the attempt 7 synthesis were
addressed as design contracts or explicit gates.

## Verification

- Ran `git diff --check -- .gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Saved `design-after.md`.
- Saved `design.diff` against `attempt-7/design-before.md`.
