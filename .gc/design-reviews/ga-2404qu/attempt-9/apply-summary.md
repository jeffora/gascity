# Apply Summary

## Verdict

Global synthesis verdict was `block`, so this apply pass sets
`design_review.verdict=iterate`.

## Changes Applied

- Added `Attempt 9 Review Resolution Contracts` to
  `.gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Reframed required Core loading around a typed
  `RequiredSystemPackParticipation` contract, two fatal gates, full production
  loader-bypass scanning across `cmd/gc` and behavior-driving `internal/`
  surfaces, and fatal Core collision semantics.
- Chose a backward-compatible public Gastown compatibility pin followed by a
  no-Maintenance activation pin, with an atomic pin/removal fallback if the
  public pack cannot safely split the two states.
- Required ordinary remote source identity for public Gastown and duplicate
  active-definition tests across current loader, old/new binary, rollback,
  stale local import, stale system-pack directory, synthetic cache, and ordinary
  remote-cache states.
- Made one retired-source classifier the sole path for config loading,
  pre-resolution doctor, import/check/install, packman, cache/lock validation,
  builtin materialization, prompt/template discovery, docs lint, generated
  references, and error rendering.
- Defined a single `gc doctor --fix` mutation coordinator with operation-scoped
  preflight, staged Core/public content, CST/span-preserving edits, overlay
  validation, compare-before-rename publication, controller-active refusal, and
  failure-injection coverage.
- Added a checked-in Go and asset role-surface table contract and settled
  `dog`, `crew`, tmux display identity, non-Gastown init, prompt fallback,
  warmup targets, Core mail/nudge targets, and sling/formula selection.
- Strengthened behavior preservation with an execution-level witness floor,
  whole-repo plus pinned-public-pack discovery, per-code-path rows, semantic
  delta records, authorship/provider/session-hook coverage, and exact-pinned
  packcompat verification.
- Made `GC_BOOTSTRAP=skip` retired as a production behavior switch and specified
  the production empty `fs.FS`, `_test.go` fixture allowlist, and old bootstrap
  Core leakage guards.
- Moved docs and operator vocabulary into a first operator-facing slice gate
  through a canonical nav-registered `docs/reference/system-packs.md`, generated
  wording matrix, lint coverage, golden outputs, generated references, and
  release/non-release rule.
- Added slice-specific package, packcompat, old/new binary, offline/cache,
  runtime-state, registry/cache, docs, and source-deletion gates using `v1.2.1`
  as the old binary fixture.
- Preserved the workflow artifact path issue as an explicit workflow follow-up
  requiring attempt-local persona syntheses and metadata validation.

## Unfixable Items

None. All blocker and major findings from the attempt 9 synthesis were
addressed as design contracts, explicit rollout choices, executable gates, or
workflow follow-up tracking.

## Verification

- Ran `git diff --check -- .gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Saved `design-after.md`.
- Saved `design.diff` against `attempt-9/design-before.md`.
