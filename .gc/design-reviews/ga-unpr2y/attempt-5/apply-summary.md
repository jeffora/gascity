# Apply Summary

Source synthesis: `.gc/design-reviews/ga-unpr2y/attempt-5/synthesis.md`
Global verdict: `block`
Applied verdict: `iterate`

## Changes Applied

- Updated `internal/session/DESIGN.md` latest disposition to attempt 5.
- Added an attempt-5 response table mapping every blocker and major finding to
  a design response and explicit exit gate.
- Expanded source-inventory requirements for the missed API, CLI, mail, chat,
  wait, manager, repair, patch-builder, and runtime-identity writer paths.
- Added inventory rows W-022 through W-028 and tightened static guard fixtures
  for raw store writes, manager calls, package-level mutators, dynamic batches,
  patch-map extension, generic bridges, and type/status mutation.
- Split target classification from operation policy, repair, materialization,
  and mail recipient-set behavior.
- Added slice-level proof coverage for transition reducer rows and reserved
  named-session materialization.
- Tightened command atomicity around tokened or revisioned commits,
  stale-success handling, provider compensation, no-token repair, and close
  path convergence.
- Reframed current events as hints unless a typed-payload migration slice is
  explicitly implemented; durable scans remain the critical recovery authority.
- Added the three-stage reconciler boundary, per-key rollback matrix, concrete
  doctor/session-inspect checks, enforceable vocabulary metadata, and numeric
  hot-path/backpressure budgets.

## Residual Risks

- The source-verified writer inventory and numeric performance baselines cannot
  be completed by editing `DESIGN.md` alone. The design now marks them as
  blocking pre-decomposition artifacts rather than resolved prose.
- No code or schema behavior was changed in this apply step.
