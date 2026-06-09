# Apply Summary

## Verdict

Global synthesis verdict was `block`. The design was updated to address all
Blocker and Major findings, and this pass sets `design_review.verdict=iterate`.

## Changes Applied

- Updated `.gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Added an Attempt 14 contract section that supersedes stale wording around
  path-only Core proof, doctor mutation, retired-pack fallback, synthetic public
  Gastown activation, role-cleaning without provider provenance, broad slice
  gates, and docs-only behavior preservation.
- Specified resolver-produced required-pack provenance with stable layer ids,
  import edge ids, file-set digests, implicit required-host include semantics,
  partial-read exceptions, and a loader/failure-class matrix.
- Made plain `gc doctor` read-only by limiting it to validate-only and
  pre-resolution diagnostic APIs; all mutation is constrained to
  `doctor.MutationCoordinator` behind `--fix` with failure-injection coverage.
- Added a role-surface ownership table covering Go role literals,
  API/dashboard `crew` vocabulary, Core `dog` compatibility data, provider
  routes, formula metadata, mail/nudge recipients, prompt fragments, public
  Gastown assets, and historical docs/fixtures.
- Promoted retired Maintenance and local Gastown containment to a normal
  production resolution invariant with a classifier/rejection matrix for stale
  directories, explicit imports, transitive imports, custom forks, stale
  synthetic caches, prompt discovery, and lock entries.
- Split public pin evidence into `current_baseline`, `compatibility`, and
  `activation` phase records, including the shipped
  `sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b` baseline and old/new binary
  content-digest behavior.
- Added `slice-gates.generated.yaml` as the binding rollout gate artifact and
  defined commit-level activation ordering that keeps intermediate states green.
- Required first-slice generated evidence plus pilot behavior rows for
  shutdown-dance, spawn-storm, `DOG_DONE`, JSONL/reaper, Polecat handoffs,
  branch pruning, review checks, doctor scripts, prompt fragments,
  `session_live`, and public Gastown archive authorship.
- Tightened bootstrap fixture isolation with a concrete test-only fixture seam,
  non-bootstrap test migration rules, `GC_BOOTSTRAP=skip` closure, and old-path
  guards.
- Defined a generated docs/wording linter schema and the required contents of
  the canonical `docs/reference/system-packs.md` operator page.
- Reconciled provider-pack byte continuity with role-cleaning through a
  provider exception ledger for `bd`, `dolt`, provider locks, metadata, and
  generated comments.

## Unfixable Items

None. Every blocker and major finding from the attempt 14 synthesis was
converted into an implementation contract, artifact, table, or gate.

## Verification

- Ran `rg -n "[ \t]$" .gc/design-review-inputs/core-gastown-pack-migration/design.md`;
  no trailing whitespace matches were found.
- Reviewed the attempt-local diff against
  `.gc/design-reviews/ga-2404qu/attempt-14/design-before.md`.
- Saved `.gc/design-reviews/ga-2404qu/attempt-14/design-after.md`.
- Saved `.gc/design-reviews/ga-2404qu/attempt-14/design.diff`.
