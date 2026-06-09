# Apply Summary

Source synthesis: `.gc/design-reviews/ga-1ekw9l/attempt-3/synthesis.md`

Global verdict: `block`

Recommended apply verdict: `iterate`

Applied changes:

- Added concrete external prerequisite artifacts for public Gastown pins,
  behavior preservation, ownership rows, and packcompat transcripts.
- Strengthened the Behavior Evidence contract with planned generator/schema
  paths, source-kind and witness-kind enums, sample/freshness requirements, and
  a Git historical baseline.
- Defined the required system-pack runtime result, allowed-use modes, provider
  selection boundary, and caller obligations for fail-closed Core behavior.
- Added zero-duplicate-active and zero-merge gates for bundled/public/cache pin
  authority.
- Chose doctor-owned runtime-state migration, added `internal/doctorfix`
  coordinator ownership, and made controllers diagnostics/refusal-only for
  legacy runtime-state migration.
- Added binding precedence, required-recipient failure behavior, formula
  metadata replacement for role-name heuristics, bootstrap empty-FS behavior,
  wording matrix rules, runtime-state migration table, focused tests, and
  smaller rollout slices.

Unfixable in this document-only apply step:

- Workflow artifact defect: attempt 3 persona syntheses were not written under
  `.gc/design-reviews/ga-1ekw9l/attempt-3/persona-syntheses/`; the workflow
  needs attempt metadata/output-path repair before the next approval-seeking
  review can be fully trusted.
- External prerequisite: public `gascity-packs` commits, generated behavior
  evidence, pin ledgers, ownership rows, packcompat transcripts, and cache/offline
  proof are outside this implementation-plan edit. The plan now names them as
  prerequisite artifacts and gates dependent Gas City work on them.

Outcome: set `design_review.verdict=iterate` so the design-review workflow can
review the refined plan and separately address the workflow artifact defect and
external public-pack evidence.
