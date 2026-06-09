# Apply Summary

Applied attempt 6 global synthesis to `internal/session/DESIGN.md`.

## Verdict

`design_review.verdict=iterate`

## Changes

- Updated the latest disposition to attempt 6 and added an attempt-6 review
  response table for all blocker and major findings.
- Added mandatory non-mutating Slice 0 preflight artifacts for source inventory,
  static mutation guard, shrink-only allowlist, fixtures, and proof command.
- Made scenario parity a canonical `SCENARIO_PARITY.yaml` source with a
  citation-freshness test that rejects missing paths, stale assertions,
  commit-only current proof, and non-ancestor commit evidence.
- Added command-applier requirements for stale-success handling, partial-state
  matrices, provider side-effect ordering, repair authority, event ordering,
  and race tests; fenced runtime-start and close explicitly.
- Split raw target classification from adapter-owned policy, repair,
  materialization, and authoritative `selected` results.
- Added runtime observation completeness states and fail-closed rules for
  destructive reconciler branches.
- Strengthened durable scan recovery, post-commit fact/idempotency keys,
  close/work-release scan ownership, diagnostics, and performance gates.
- Made migration coexistence mechanical with guard-row metadata, rollback files,
  proof commands, and one-writer proof on every implementation bead.
- Added executable vocabulary checkpoint artifacts and tests for shared types
  and fields.
- Documented the current-attempt persona artifact-path defect as workflow
  plumbing that this design cannot repair.

## Artifacts

- `.gc/design-reviews/ga-unpr2y/attempt-6/design-after.md`
- `.gc/design-reviews/ga-unpr2y/attempt-6/design.diff`
- `.gc/design-reviews/ga-unpr2y/attempt-6/apply-summary.md`
