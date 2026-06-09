# Apply Summary

Global synthesis verdict: `block`
Apply verdict: `iterate`

Updated `engdocs/design/formula-compiler-requirements.md` to address all
blocker and major findings from attempt 73.

## Changes Applied

- Added a mechanically enforceable validation-matrix contract with fixture
  paths, generated Go test names, generator self-tests, golden diagnostics,
  row schema, JSON-loader policy, and public `check` compatibility coverage.
- Split preview and accepted projection snapshots, constrained
  `HostCapabilities` construction, pinned accepted-artifact proof hashing, and
  named package homes for source/provenance types.
- Reworked active legacy-root handling so runtime mutators do not auto-repair;
  every graph-control caller now has an explicit `operator_repair_required` or
  `read_only_fail_closed` mode, and `gc formula repair-root-artifact` has
  inputs, validation order, atomic write, exit-code, idempotency, and
  concurrency contracts.
- Strengthened host-capability diagnostic grouping so formula identity,
  producer, subject, requirement source, host source, content hash, and config
  generation cannot collapse distinct failures; clarified diagnostic-state
  writes versus protected runtime writes.
- Made packman schema 2 or equivalent packman-owned provenance an explicit
  prerequisite for resolver/import enforcement, alias removal, and external
  support expiration; added stable external-author command/report contracts.
- Added alias-removal JSON report fields for release tag/date, rollback
  window, baseline binary, corpus paths, background accepted aliases, support
  rows, and external classifications.
- Tightened rollout ordering so docs/help/schema/OpenAPI/dashboard/generated
  artifacts must land before each visible diagnostic surface, with packman
  provenance sequenced before pack-floor enforcement.
- Added convergence transition fencing, legacy-rule parity coverage,
  `ConvergenceRuntimeInputs`, pre-create subject keys, canonical convergence
  artifact-ref conflict handling, and fixtures for retry identity/host toggles.
- Made `make formula-docs-check` a real implementation/CI prerequisite and
  captured minor edge contracts for parser authority, dashboard affordances,
  Phase 5 deliverables, future grammar, and default-capability identity.

## Unfixable Items

None. All blocker and major findings were addressed in the design text.

## Artifacts

- `.gc/design-reviews/ga-j195cx/attempt-73/design-after.md`
- `.gc/design-reviews/ga-j195cx/attempt-73/design.diff`
