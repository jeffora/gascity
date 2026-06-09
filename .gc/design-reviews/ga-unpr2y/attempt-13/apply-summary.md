# Apply Summary

Verdict: `iterate`

Updated `internal/session/DESIGN.md` for attempt 13.

Changes made:

- Marked the latest design-review disposition as attempt 13.
- Added `## Attempt 13 Review Response` with a `<!-- REVIEW: added per attempt-13-global-synthesis -->` marker.
- Made Attempt 13 the controlling design contract when earlier attempt-response tables disagree or omit required proof.
- Added `internal/session/SLICE0_CONTRACT.yaml` as the single machine-readable Slice 0 close contract for artifacts, schemas, validators, fixtures, proof commands, and workflow finalizer metadata.
- Tightened Slice 0 into the only first decomposable backlog item and added a backlog item 0 for the non-mutating executable preflight.
- Added hard gates for source-complete writer/read/route inventories, repair/backfill ownership, runtime-start no-token repair-only paths, close/work-release scanner recovery, pure-decider boundary matrices, scenario parity and amendment proof, per-surface target compatibility, route/read completeness, diagnostics/event truthfulness, performance budgets, worker-boundary exceptions, rollback, and data-direction proof.
- Added vocabulary lifecycle states (`documented`, `private`, `provisional`, `delegating`) and demoted `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent` until a real delegated production caller proves minimal fields.
- Updated target-classification language so it is the first behavior extraction only after the closed Slice 0 preflight.

No behavior-moving work is approved by this revision. The only schedulable implementation work remains the non-mutating Slice 0 executable preflight.
