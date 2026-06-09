# Apply Summary

## Verdict

`iterate`

## Applied Changes

- Preserved the required `gc.mayor.implementation-plan.v1` top-level schema and kept the plan as an implementation-plan artifact.
- Updated the readiness posture: only prerequisite-producing work may decompose until AC6, AC7, and AC14-AC17 evidence exists and passes.
- Added a trusted required-pack descriptor flow, Gate 1/Gate 2 binding, runtime ready guard, and AST/type-aware loader bypass inventory.
- Added executable Behavior Evidence commands, evidence classes, public-pack witness rules, and compatibility/activation packcompat semantics.
- Moved subpath-aware cache identity, lock/cache provenance, synthetic-alias rejection, and pin-coherence proof ahead of public pin consumption.
- Added doctor-fix inventory, active-root enumeration guard, bootstrap buildable end state, docs vocabulary authority, required support artifact set, focused proof commands, AC17 proof matrix, and slice-to-gate table.

## Unfixable Or External Items

- Public `gascity-packs` commits, ownership rows, `behavior-preservation.yaml`, `public-gastown-pins.yaml`, packcompat transcripts, and live public-pack validation remain external prerequisites. The plan now blocks dependent Gas City slices until those artifacts are cited by immutable commit or checked path.
- The workflow defect where attempt-6 persona syntheses were written under the compatibility attempt-1 path is not fixable in `implementation-plan.md`; it requires workflow attempt-directory propagation fixes outside this design artifact.

## Artifacts

- Design after: `.gc/design-reviews/ga-1ekw9l/attempt-6/design-after.md`
- Diff: `.gc/design-reviews/ga-1ekw9l/attempt-6/design.diff`
- Synthesis: `.gc/design-reviews/ga-1ekw9l/attempt-6/synthesis.md`
