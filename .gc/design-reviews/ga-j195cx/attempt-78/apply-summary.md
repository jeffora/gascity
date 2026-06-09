# Apply Summary

Verdict source: `.gc/design-reviews/ga-j195cx/attempt-78/synthesis.md`

Global verdict was `block`, so this apply pass addresses blocker and major
findings in `engdocs/design/formula-compiler-requirements.md` and sets the
workflow verdict to `iterate`.

Changes applied:

- Replaced the lossy host-capability constructor shape with structured
  `HostCapabilityInput` / `HostCapabilitySource` provenance and added parity
  fixtures for omitted defaults, explicit false/true, deprecated alias
  promotion, test overrides, config reloads, and host downgrade artifact reuse.
- Added a transitional no-bypass gate for requires-only graph formulas across
  `MolCook`, bd/exec materialization, convergence, fanout, orders, API, and
  dashboard/generated-client projection.
- Made matrix coverage more executable with caller-preflight count derivation,
  raw-shape and combined-defect rows, JSON loader disposition, construct
  identity crosswalks, and axis-owned minimum parsing.
- Tightened compatibility evidence for exact legacy `contract` spellings, dual
  declarations, omitted/empty/default requirements, and typed external support
  expiration fields.
- Promoted packman provenance readiness to an explicit rollout unit before
  pack-floor enforcement, with owner, PR home, saved artifacts, rollback, and
  schema-1 fail-closed behavior.
- Added durable accepted-warning persistence and cadence rules for
  deprecation/version warnings across producer state, dashboards, and release
  reports.
- Clarified docs/proposal alignment: Phase 2 must supersede the migration
  proposal, add positive stale-guidance checks, include common-confusion docs,
  and decide JSON formula support with fixtures or removal.
- Strengthened convergence runtime inputs and artifact-ref rules with typed
  runtime-var evidence, hashes, duplicate/default/redaction handling, canonical
  artifact key precedence, dual-stamp migration, and conflict repair behavior.

Unfixable items: none documented; all synthesis blocker/major findings were
addressed by design text additions.

Artifacts saved:

- `.gc/design-reviews/ga-j195cx/attempt-78/design-after.md`
- `.gc/design-reviews/ga-j195cx/attempt-78/design.diff`
- `.gc/design-reviews/ga-j195cx/attempt-78/apply-summary.md`
