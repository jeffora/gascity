# Apply Summary

Applied attempt 7 global synthesis to `internal/session/DESIGN.md`.

## Verdict

`design_review.verdict=iterate`

## Changes

- Updated the latest disposition to attempt 7 and added an attempt-7 review
  response table for all blocker and major findings.
- Made Slice 0 the only allowed preflight before mutation-owning work, with
  pinned baseline, generated symbol/key/endpoint inventory, scenario parity,
  vocabulary checkpoints, shrink-only guard allowlist, fixture, and zero-match
  proof requirements.
- Reframed store writes around proven conditional primitives, tokened or
  phased blind writes, or inert advisory writes; blocked stale-sensitive blind
  writes by default.
- Added durable close/work-release recovery requirements for release identity
  preservation, transitional close facts, scanner cadence, idempotency,
  successor suppression, and crash/event-miss tests.
- Added a field-level `AwakeInput` disposition table and split eligibility into
  forced runnable, conditional runnable, idle-suppressible, blocked, terminal,
  unknown/partial, and repair-needed results.
- Added surface-specific terminal/fallthrough contracts for API, CLI, mail,
  extmsg, assignee normalization, nudge, attach, inspect, log, transcript, and
  pool-resume target resolution.
- Made `RepairEmptyType` and other repair writes operator-visible, guard-visible
  mutations with audited helper, repair-slice, exception, or non-session-only
  classifications.
- Required centralized diagnostics vocabulary, rendering tests, and a shared
  counting-store or benchmark substrate for performance and subscriber budgets.
- Tightened migration delivery to per-surface adoption with explicit active
  versus provisional vocabulary metadata, surface parity, and revert paths.

## Unfixable In This Design Artifact

- The attempt-path workflow issue is outside `internal/session/DESIGN.md`.
  The design keeps documenting that the design-review workflow must write
  current-attempt persona syntheses under the current attempt directory or emit
  a source manifest.

## Artifacts

- `.gc/design-reviews/ga-unpr2y/attempt-7/design-after.md`
- `.gc/design-reviews/ga-unpr2y/attempt-7/design.diff`
- `.gc/design-reviews/ga-unpr2y/attempt-7/apply-summary.md`
