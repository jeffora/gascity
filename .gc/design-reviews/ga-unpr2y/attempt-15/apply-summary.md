# Apply Summary

Verdict: `iterate`

Updated `internal/session/DESIGN.md` for attempt 15.

Changes made:

- Marked the latest design-review disposition as attempt 15 global verdict
  `block` with this revision as an `iterate` response.
- Added `## Attempt 15 Review Response` with
  `<!-- REVIEW: added per attempt-15-global-synthesis -->`.
- Narrowed Slice 0 to universal, non-mutating evidence only and moved command,
  event, route, diagnostics, cost, and migration details into per-slice
  preflights.
- Required self-validating Slice 0 proof: schema versions, owners,
  source paths, exact proof selectors, positive and negative fixtures,
  no skipped validators, and stale `SESSION-*` evidence repair or retirement.
- Rewrote the first API query target-classifier adopter around the exact
  `resolveSessionTargetIDWithContext` order, including configured named
  sessions, config-orphan rejection, path-alias by `Title`, allow-closed
  lookup, Huma/legacy error rendering, and `RepairEmptyType` quarantine.
- Expanded mutation ownership to cover session-owned key families, create-time
  metadata, dynamic patch maps, direct status/type/close/reopen writes, repair
  helpers, and generic API/CLI bead mutation bridges.
- Added stricter atomic command requirements for immutable fact snapshots,
  mandatory `now` and config facts, pre-commit revalidation, store capability
  rows, post-write verifiers, and failure-injection tests.
- Strengthened worker-boundary rules so production API/CLI lifecycle mutations
  default to `worker.Handle` unless exact root-approved expiring exceptions
  exist.
- Added migration coexistence and rollback requirements for overlapping API,
  CLI, reconciler, and session mutation files.
- Made `BOUNDARY_MATRIX.yaml`, event recovery, diagnostics, and hot-path cost
  rows concrete, including current `session.*` event inventory and explicit
  budget handling for `resolveLiveSessionByPathAlias`.

Unfixable in this document:

- Attempt-15 child persona beads were stamped with some output paths under
  `attempt-1`; the global synthesis normalized copies into
  `attempt-15/persona-syntheses`. Artifact placement is workflow plumbing
  outside `internal/session`.

Saved artifacts:

- `.gc/design-reviews/ga-unpr2y/attempt-15/design-after.md`
- `.gc/design-reviews/ga-unpr2y/attempt-15/design.diff`
- `.gc/design-reviews/ga-unpr2y/attempt-15/apply-summary.md`
