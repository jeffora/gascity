# Apply Summary

Verdict: iterate

Updated `plans/core-gastown-pack-migration/requirements.md` in place to address
document-fixable blockers from the attempt-1 synthesis.

Changes applied:

- Converted the requirements body to `gc.mayor.requirements.v1` section order:
  `Problem Statement`, `W6H`, `Example Mapping`, `Acceptance Criteria`,
  `Out Of Scope`, and `Open Questions`.
- Changed front matter status from `approved` to `questions` because material
  product decisions and supporting evidence still block implementation-plan
  readiness.
- Removed the file-by-file migration table from the normative requirements
  body and replaced it with a requirement for a validated asset migration
  ledger.
- Added W6H coverage for owners, timing, affected surfaces, required Core
  behavior, explicit public Gastown imports, retired Maintenance behavior, and
  migration scale.
- Added happy, negative, and edge examples for fresh Core-only operation,
  fresh Gastown init, missing Core diagnostics, legacy imports, role-neutrality
  scans, configurable maintenance executors, offline public-pack resolution,
  and stale system-pack state.
- Added an acceptance traceability matrix mapping product outcomes to concrete
  verification modes.
- Grounded the public Gastown import contract in the current source and pin:
  `https://github.com/gastownhall/gascity-packs.git//gastown` and
  `sha:d3617d1319a1206ac85f69ba024ec395c49c6f4b`.
- Added explicit Open Questions for the validated ledger owner, behavior
  manifest/release gate, version-skew policy, in-flight retired-path sessions,
  and the explicit repair command/workflow.

Unfixable in this apply step:

- The validated asset migration ledger is an external prerequisite. The
  requirements now define its contract, but this apply step did not generate or
  validate the ledger.
- The behavior-preservation manifest or harness is an external prerequisite.
  The requirements now define its contract, but this apply step did not produce
  executable proof across Gas City and public Gastown.

Artifacts saved:

- `.gc/design-reviews/ga-dtvdnd/attempt-1/design-after.md`
- `.gc/design-reviews/ga-dtvdnd/attempt-1/design.diff`
- `.gc/design-reviews/ga-dtvdnd/attempt-1/apply-summary.md`
