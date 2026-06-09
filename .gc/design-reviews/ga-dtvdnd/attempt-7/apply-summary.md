# Apply Summary

Source synthesis: `.gc/design-reviews/ga-dtvdnd/attempt-7/synthesis.md`

Global verdict: `block`

Recommended apply verdict: `iterate`

Applied verdict: `iterate`

Updated `plans/core-gastown-pack-migration/requirements.md` to address the
document-fixable blockers and majors from attempt 7:

- Declared `internal/bootstrap/packs/core` as the sole canonical source
  authority for release-bundled Core and required generated/materialized/cache
  paths to trace to it or be classified as non-runtime state.
- Tightened source-consumer closure over Go embed packages, builtin registry
  functions, materialization commands, hook/import paths, generated hashes,
  synthetic layouts, and tests.
- Strengthened pack-resolution determinism for provider-pack cardinality,
  condition-code consistency, local override pin validation, overlays, optional
  public-pack name conflicts, and transitive diamond conflicts.
- Bound AC6 and AC7 bidirectionally with stable behavior IDs, row/call-site
  witnesses, and a symmetric Core/non-Gastown behavior baseline.
- Narrowed role-neutrality exceptions for literal `dog` and retired role names
  to configured-default data keys, source-attribution diagnostics, migration
  docs, generated review artifacts, and test fixtures.
- Added concrete operator semantics for diagnostics and repair: shared
  condition-code registry, exit-code/stdout/stderr contracts, transaction
  backup or journal fields, inactive task-store reference handling, stale
  directory policy, rollback/downgrade/offline manual reconciliation, and
  live-state refusal evidence.
- Made validation gates less vacuous by requiring active execution evidence
  such as `go test -json`, skipped/no-op/empty witness failure, sanctioned
  mirror/redirect proof, and an executable AC17 gate before decomposition.
- Removed the previous inline `REVIEW` marker so the artifact remains clean
  `gc.mayor.requirements.v1` output.

Unfixable in this document:

- The support artifacts and proofs named by the synthesis remain external
  prerequisites: pack-resolution matrix, source-consumer closure,
  asset-migration ledger, behavior-preservation manifest, diagnostics schema,
  role-neutrality scan, public Gastown pin/version-skew/cache proof, and AC17
  acceptance-proof matrix.
- The design-review workflow defect remains outside `requirements.md`:
  attempt-local persona syntheses for attempt 7 are missing while fresh
  syntheses were written under another attempt directory. The workflow should
  fail before global synthesis when active-attempt persona artifacts are
  missing or stale.

Saved artifacts:

- `.gc/design-reviews/ga-dtvdnd/attempt-7/design-after.md`
- `.gc/design-reviews/ga-dtvdnd/attempt-7/design.diff`
- `.gc/design-reviews/ga-dtvdnd/attempt-7/apply-summary.md`
