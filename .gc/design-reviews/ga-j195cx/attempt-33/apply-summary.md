# Apply Summary

Verdict applied: `block` -> `iterate`.

Updated `engdocs/design/formula-compiler-requirements.md` to address all
attempt-33 blocker and major findings:

- Made accepted compile artifacts fail closed with an unexported proof,
  explicit write-intent validation, identity checks, and negative test
  requirements for forged, zero-value, stale, mismatched, host-disabled, and
  fatal-diagnostic artifacts.
- Replaced the unbounded validation-matrix cross-product with named executable
  suites, count locks, parser-boundary ownership, and projection parity
  fixtures.
- Defined typed workflow-root facts for unknown future capabilities and added
  old-reader lifecycle rules.
- Completed caller and durable preflight coverage for shell-out probes, fanout
  state, artifact validation, and zero-write guarantees.
- Added rollout ownership for compatibility artifacts, conservative seed
  values, pack-floor enforcement before first-party requires-only conversion,
  legacy-version reporting, and the concrete meaning of two minor releases.
- Pinned diagnostics policy: warning diagnostics never publish Event Bus
  events, fatal compile events are typed wrappers only, and
  `config_generation` is defined for grouping and suppression.
- Tightened convergence projection around source-attributed accepted output,
  missing-child repair, and durable path ordering.

No unfixable blocker or major item remains documented as accepted risk.
