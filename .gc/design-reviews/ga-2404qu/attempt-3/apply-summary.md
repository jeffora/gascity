# Apply Summary

Attempt: 3

Global verdict: block

Applied verdict: iterate

Updated `.gc/design-review-inputs/core-gastown-pack-migration/design.md` to turn
the attempt-3 blocker and major findings into explicit implementation gates.

Changes made:

- Added `Attempt 3 Review Resolution Contracts` with source-derived behavior
  manifest requirements, row-level old/new witnesses, immutable public Gastown
  commit proof, consuming Gas City pin proof, and machine-enforced semantic
  delta approval.
- Added typed required-Core identity and loader participation requirements,
  including content-backed system-pack provenance, name-collision failures, and
  a seeded production `config.Load*` scanner/allowlist.
- Strengthened doctor/import-state safety with preflight-before-mutation,
  concurrent-controller refusal, final re-read before renames, scoped TOML
  preservation, generated-source provenance, and post-fix Core/public-pack
  revalidation.
- Added role-neutral Core and SDK self-sufficiency gates covering Go and asset
  scanners, role-derived control paths, configurable `dog` ownership, and
  renamed/absent maintenance-worker tests.
- Added public Gastown pin/cache sequencing: Gas City must adopt and prove the
  exact public `PublicGastownPackVersion` before behavior removal,
  generalization, registry cleanup, or source deletion.
- Added Maintenance runtime-state and zero-duplicate-order requirements for
  retired Maintenance paths, JSONL/export/storm/order state, host-Core public
  Gastown dependency, and doctor behavior.
- Added bootstrap fixture and `GC_BOOTSTRAP=skip` constraints to prevent
  production Core from remaining under `internal/bootstrap`.
- Added docs, behavior-test, provider-continuity, and required-pack integrity
  gates so path/name/count assertions are replaced by behavior proof.
- Updated the rollout ordering so the public-pin adoption and packcompat slice
  happens before Core/Maintenance behavior is moved or removed; registry/cache
  cleanup now consumes the already-updated pin.

Artifacts:

- Design after: `.gc/design-reviews/ga-2404qu/attempt-3/design-after.md`
- Diff: `.gc/design-reviews/ga-2404qu/attempt-3/design.diff`
- Synthesis: `.gc/design-reviews/ga-2404qu/attempt-3/synthesis.md`

Residual status:

The design must iterate because the global verdict was `block`. The newly added
contracts intentionally preserve the findings as implementation gates rather
than marking them resolved without proof.
