# Apply Summary

Attempt: 4
Synthesis: `.gc/design-reviews/ga-unpr2y/attempt-4/synthesis.md`
Global verdict: `block`
Applied verdict: `iterate`

## Changes Applied

- Updated `internal/session/DESIGN.md` to record attempt 4 as the latest
  blocking review and keep the design in iterate state.
- Added an attempt-4 response section that blocks decomposition until each slice
  has source-verified mutation inventory, row-level requirements parity,
  command atomicity proof, durable reaction recovery, routing ownership,
  coexistence metadata, repair diagnostics, performance proof, and vocabulary
  checkpoints.
- Made the mutation inventory gate explicit about source paths, dynamic metadata
  batches, generic bridge writes, target bead discrimination, shrink-only
  allowlists, and static guard coverage.
- Added a full scenario ownership baseline for every `REQUIREMENTS.md` row and
  made missing/stale proof files blocking rather than acceptable by memory.
- Tightened runtime-start around tokened prepare/commit, store primitive proof,
  partial-write repair, provider-start-success/commit-failure handling, and
  successor-token cleanup.
- Added a durable recovery matrix for work release, drain, wake demand, identity
  cleanup, and mail/extmsg cleanup, with durable scans as the recovery
  authority.
- Split session eligibility from controller demand and runtime/provider gates,
  and added per-caller routing rules for CLI, API, Huma API, reconciler,
  worker, mail, extmsg, tests, and repair.
- Replaced the old event-loss diagnostic wording with durable-scan and
  event-emission outcome language.
- Documented the attempt-scoped persona artifact placement issue as workflow
  plumbing outside `internal/session/DESIGN.md`.

## Notes

- Saved `design-after.md` and `design.diff` in this attempt directory.
- No code or tests were changed; this was a design-document update only.
