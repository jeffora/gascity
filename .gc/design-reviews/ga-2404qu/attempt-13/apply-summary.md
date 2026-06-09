# Apply Summary

## Verdict

Global synthesis verdict was `approve-with-risks`. Strict mode is enabled for
this design review, so this apply pass documents and addresses the risks it can
in the design and sets `design_review.verdict=iterate`.

## Changes Applied

- Updated `.gc/design-review-inputs/core-gastown-pack-migration/design.md`.
- Made Public Gastown activation-pin adoption and Maintenance removal a single
  candidate change, with the no-Maintenance production-loader gate run on that
  combined tree before merge.
- Added a provider-pack byte-continuity exception table for `bd` and `dolt`,
  allowing only explicit role-cleaning and target-binding rewrites with
  behavior witnesses and provenance.
- Defined required-pack prune/quarantine repair before strict validation so
  unexpected effective files cannot deadlock startup recovery or behavior
  discovery.
- Strengthened `gc doctor --fix` ordering: lock before mutating preflight,
  revalidate any report-only pre-lock reads, land the mutation coordinator before
  exposing public-pin/import rewrites, or freeze legacy automatic fixes.
- Made runtime-state migration non-destructive copy semantics with
  migration-completed markers, so retained legacy Maintenance divergence is not
  treated as a recurring conflict after Core is live.
- Clarified offline/cache behavior: public Gastown reachability can be network
  source or validated ordinary remote-cache entry for the exact source, commit,
  and subpath; retired synthetic cache can only be promoted, not selected.
- Tightened bootstrap cleanup: delete the empty embed directive/variable, use
  inline synthetic fixtures such as `packs/test-core`, and include
  `cmd/gc/prompt_test.go` in the hidden-dependency inventory.
- Replaced stale audit-timing wording so cross-pack ownership audits are
  resolved before either public pin is consumed.

## Unfixable Items

None. The remaining state is intentionally `iterate` because strict mode treats
the current approve-with-risks synthesis as requiring another review pass.

## Verification

- Ran `git diff --no-index --check -- .gc/design-reviews/ga-2404qu/attempt-13/design-before.md .gc/design-review-inputs/core-gastown-pack-migration/design.md`; no whitespace diagnostics were emitted.
- Ran `rg -n "[ \t]$" .gc/design-review-inputs/core-gastown-pack-migration/design.md`; no trailing whitespace matches.
- Ran a stale-phrase scan for circular rollout, broad byte-continuity, runtime
  move, and network-only wording; no targeted stale phrases remained.
- Saved `.gc/design-reviews/ga-2404qu/attempt-13/design-after.md`.
- Saved `.gc/design-reviews/ga-2404qu/attempt-13/design.diff`.
