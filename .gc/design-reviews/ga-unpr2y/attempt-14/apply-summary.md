# Apply Summary

Verdict: `iterate`

Updated `internal/session/DESIGN.md` for attempt 14.

Changes made:

- Marked the latest design-review disposition as attempt 14 global verdict
  `block` with this revision as an `iterate` response.
- Added `## Attempt 14 Review Response` with
  `<!-- REVIEW: added per attempt-14-global-synthesis -->`.
- Made `DESIGN_REVIEW_NOTES.md` explicitly historical and non-normative unless
  copied into the active design.
- Added a non-mutating Slice 0 evidence preflight as the only schedulable first
  work, with required artifacts for scenario parity, target classification,
  mutation ownership, command appliers, API/CLI routing, worker-boundary
  exceptions, diagnostics, vocabulary, and validators.
- Added an active Target Classification Contract with typed candidate/result
  kinds, a no-side-effect classifier rule, a first read-only API query adopter,
  and a per-surface precedence/delegation matrix.
- Added Mutation Ownership Ledger, Atomic Command Contract,
  Worker/API/CLI Boundary, Reconciler/Runtime/Session Split, Event and
  Recovery Contract, Observability and Cost, and Vocabulary Lifecycle sections.
- Updated the backlog so Slice 0 precedes target classification and all later
  behavior-moving work depends on Slice 0 artifact rows and proof.
- Documented the attempt-14 persona artifact placement issue as unfixable in
  this design document and owned by the design-review workflow formula.

Note: the live `internal/session/DESIGN.md` at apply time was a compact
136-line draft, while `.gc/design-reviews/ga-unpr2y/attempt-14/design-before.md`
was the larger 2013-line artifact reviewed by the personas. This update edits
the live design in place and preserves its compact direction rather than
restoring the older dense artifact wholesale.

Saved artifacts:

- `.gc/design-reviews/ga-unpr2y/attempt-14/design-after.md`
- `.gc/design-reviews/ga-unpr2y/attempt-14/design.diff`
- `.gc/design-reviews/ga-unpr2y/attempt-14/apply-summary.md`
