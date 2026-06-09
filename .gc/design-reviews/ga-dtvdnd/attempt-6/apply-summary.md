# Apply Summary

- Attempt: 6
- Synthesis: `.gc/design-reviews/ga-dtvdnd/attempt-6/synthesis.md`
- Global verdict: `block`
- Recommended apply verdict: `iterate`
- Applied verdict: `iterate`

## Requirements Changes

- Added a requirements-level pack-resolution product contract covering Core
  identity/provenance, required system-layer representation, missing-Core
  runtime behavior, stale materialized Core, public synthetic aliases,
  duplicate names, same-named assets, provider-conditioned support packs, and
  transitive diamond conflicts.
- Tightened W6H and AC3 so `pack-resolution-matrix.yaml` must define canonical
  Core provenance, resolved-config fields, lock/cache/materialized
  representation, diagnostic attribution, alias denial, same-pin dedupe, and
  conflicting-pin fail-closed behavior.
- Added an edge-case example for transitive pin conflicts and same-named assets.
- Tightened AC8 and AC9 role-neutrality boundaries around literal Gastown role
  names, literal `dog`, rendered templates, route/notification targets,
  diagnostic examples, and required versus optional symbolic bindings.
- Tightened AC10 and AC11 diagnostic and repair semantics with old-binary
  reconciliation, air-gapped cache seeding, direct live-state evidence,
  post-fix fields, manual reconciliation fields, and exit-code coverage.
- Tightened AC13, AC14, and AC16 so proof gates cannot pass vacuously: frozen
  historical baselines, two-repository release ordering, and randomized or
  process-unique cache staging are now explicit.

## Unfixable In This Apply Step

- External prerequisite: the split-compatible public Gastown pin, pack digest,
  behavior-manifest digest, lock/cache provenance, and fallback-disabled
  validation result still need to be produced and recorded.
- External prerequisite: AC6 asset migration ledger, AC7 behavior preservation
  manifest or harness, AC5 source-consumer closure, AC13 coverage-transfer
  mapping, AC15 pin/version-skew artifacts, and AC17 acceptance-proof matrix
  still need validated artifacts before implementation approval.
- Workflow defect: `.gc/design-reviews/ga-dtvdnd/attempt-6/persona-syntheses/`
  is empty while the ten current persona synthesis files are present under
  `.gc/design-reviews/ga-dtvdnd/attempt-1/persona-syntheses/`. This must be
  fixed in the design-review workflow; it is not a requirements content change.

## Result

`plans/core-gastown-pack-migration/requirements.md` remains a
`gc.mayor.requirements.v1` requirements artifact with `status: draft` and
`Open Questions: None`. The blocker class is now suitable for another review
iteration, but approval still depends on the external evidence artifacts and
the design-review workflow path fix.
