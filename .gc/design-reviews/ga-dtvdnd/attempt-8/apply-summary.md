# Apply Summary

Source synthesis: `.gc/design-reviews/ga-dtvdnd/attempt-8/synthesis.md`

Global verdict: `block`

Recommended apply verdict: `iterate`

Applied verdict: `iterate`

Updated `plans/core-gastown-pack-migration/requirements.md` to address the
document-fixable blocker and major findings from attempt 8:

- Aligned the requirements artifact with the implementation-plan direction by
  making `internal/packs/core` the end-state canonical Core source authority
  and classifying `internal/bootstrap/packs/core` as legacy migration input,
  deletion candidate, or non-runtime fixture/compatibility shim.
- Added a supported existing-Gastown-city upgrade happy path with explicit
  preflight, atomic repair, journal/backup, post-verification, stale-state, and
  JSON/text success evidence.
- Added a worked support-pack closure case for a surviving Dolt script
  dependency on a helper formerly housed in retired Maintenance, including the
  requirement to rehome, inline, or intentionally retire such helpers before
  Maintenance removal.
- Tightened the pack-resolution matrix contract around provider-pack
  cardinality, operation classes, bootstrap-safe exceptions, source
  attribution, same-path/static asset collisions, public aliases, and diamond
  conflicts.
- Froze the AC6 source denominator before deletion/isolation and bound AC7 to
  that denominator, a symmetric Core/non-Gastown baseline, a named
  version-pinned supported Gastown template/workflow list, and side-effecting
  behavior witnesses.
- Narrowed role-neutrality and configurable-binding policy so literal `dog`
  is only inert configured-default pack data and cannot become an executable
  route, notification target, formula binding, prompt default, overlay,
  generated default, environment override default, or Go fallback.
- Strengthened repair, diagnostics, public-pack authority, active witness,
  pin-coherence, and AC17 evidence-contract requirements without adding
  implementation-plan content to the requirements artifact.

Unfixable in this document:

- The support artifacts and proofs remain external prerequisites:
  `pack-resolution-matrix.yaml`, `source-consumer-closure.yaml`,
  `asset-migration-ledger.yaml`, `behavior-preservation-manifest.yaml`,
  `migration-diagnostics.schema.json`, `role-neutrality-scan.yaml`,
  `docs-authority-audit.yaml`, `coverage-transfer.yaml`,
  `public-gastown-pin-ledger.yaml`, `version-skew-matrix.yaml`, public
  Gastown validation, offline/cache proof, and the AC17
  `acceptance-proof-matrix.yaml`.
- The design-review workflow defect remains outside `requirements.md`:
  attempt-local persona syntheses for attempt 8 were not written under
  `.gc/design-reviews/ga-dtvdnd/attempt-8/persona-syntheses/`, and the
  workflow should fail or repair that path mismatch before global synthesis.

Saved artifacts:

- `.gc/design-reviews/ga-dtvdnd/attempt-8/design-after.md`
- `.gc/design-reviews/ga-dtvdnd/attempt-8/design.diff`
- `.gc/design-reviews/ga-dtvdnd/attempt-8/apply-summary.md`
