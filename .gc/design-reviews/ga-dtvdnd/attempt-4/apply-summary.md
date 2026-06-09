# Apply Summary

Global verdict: `block`

Recommended apply verdict: `iterate`

Applied verdict: `design_review.verdict=iterate`

Updated `plans/core-gastown-pack-migration/requirements.md` in place to address
document-fixable blockers from the attempt-4 synthesis.

Changes applied:

- Kept the artifact in the Mayor requirements schema and changed front matter
  status from `questions` to `draft`.
- Removed inline `<!-- REVIEW: ... -->` workflow-provenance comments from the
  requirements body.
- Resolved the five prior product open questions in the `Open Questions`
  section while keeping implementation approval gated on supporting evidence.
- Tightened the W6H and examples for Core provenance, bootstrap-only
  diagnostics, explicit public Gastown imports, split-asset proof, duplicate
  Core, in-flight retired-path sessions, version skew, and offline cache
  behavior.
- Expanded acceptance criteria from AC1-AC14 to AC1-AC17, adding explicit
  requirements for the AC6 asset migration ledger, AC7 behavior-preservation
  proof, Core role neutrality, configurable maintenance executor behavior,
  repair semantics, public-pack pin/version-skew policy, offline cache
  fail-closed behavior, atomic cache promotion, and an acceptance-to-proof
  matrix.
- Documented report-only-by-default repair and the selected mutating repair
  surface: `gc doctor --fix --non-interactive` or a named replacement command
  documented in release notes.

Unfixable in this apply step:

- The AC6 asset migration ledger was not generated or validated.
- The AC7 behavior-preservation manifest/harness was not generated or validated.
- The public Gastown split-compatible checkout or pinned-cache proof was not
  produced.
- The review workflow defect that writes persona syntheses under the wrong
  attempt directory was not fixed by this requirements edit.

Artifacts saved:

- `.gc/design-reviews/ga-dtvdnd/attempt-4/design-after.md`
- `.gc/design-reviews/ga-dtvdnd/attempt-4/design.diff`
- `.gc/design-reviews/ga-dtvdnd/attempt-4/apply-summary.md`

Outcome: set `design_review.verdict=iterate` so the design-review workflow can
run another pass with the updated requirements and the remaining external
prerequisites made explicit.
