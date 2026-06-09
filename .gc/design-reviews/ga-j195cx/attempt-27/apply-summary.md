# Apply Summary

Source synthesis: `.gc/design-reviews/ga-j195cx/attempt-27/synthesis.md`
Global verdict: `block`
Apply verdict: `iterate`

Updated `engdocs/design/formula-compiler-requirements.md` to address all
blocker and major findings from attempt 27:

- Added legacy `version` coverage to the executable validation matrix and made
  `version = 1` / `version = 2` unable to satisfy v2-only constructs.
- Split malformed requirement syntax from syntactically valid future
  capability requirements with distinct diagnostic codes.
- Added a checked-in raw-consumer allowlist contract and durable-writer API
  constraints so compile authority stays in `internal/formula`.
- Made migration gates executable with release artifact schemas, owners,
  update commands, `bd`/`GC_NATIVE_FORMULA=false` validation-only behavior,
  and minimum-floor/dual-window alias-removal criteria.
- Defined the Huma diagnostic payload, CLI exit-code mapping, generated-client
  parity, and event registration requirements.
- Moved convergence-specific projection ownership to `internal/convergence`
  and specified compile/project/validate/write ordering.
- Added a hard docs rollout gate and a worked second-axis example for future
  capability growth.
- Tightened provenance, `ContentHash`, shadowed formula diagnostics, and
  warning-dedup observability.

No unfixable blocker or major item remains in the design text. Because the
global verdict was `block`, the workflow should iterate for another review.
