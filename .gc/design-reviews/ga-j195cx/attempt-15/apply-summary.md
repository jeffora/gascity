# Apply Summary

Attempt: 15
Verdict applied: block -> iterate
Design document: `engdocs/design/formula-compiler-requirements.md`
Synthesis: `.gc/design-reviews/ga-j195cx/attempt-15/synthesis.md`

Addressed findings:

- [Blocker] Background diagnostics can still flood or contradict operator
  surfaces: split synchronous preview/validate behavior from durable/background
  Event Bus producers, added grouped background fatal emission rules, bounded
  `OnceKey` cache semantics, event ownership, and repeated-loop tests.
- [Blocker] Caller integration still allows divergent runtime policy: made the
  live runtime native-compiler-only, constrained `bd` to a compatibility probe,
  added per-entry-point behavior, typed workflow-root query authority, host
  capability plumbing, and fanout/convergence zero-write boundaries.
- [Major] Parser and validation matrix: made the generated fixture matrix
  normative, added required fixture fields, raw TOML classification, and
  byte-exact grammar rules.
- [Major] Legacy contract, JSON, and external compatibility: added JSON formula
  policy, external support artifact ownership, exact compatibility/report gates,
  and minimum-binary-floor constraints.
- [Major] Workflow-root authority, host capability semantics, convergence,
  provenance, rollout, and docs: pinned typed authority/projections, invalid
  capability handling, convergence projection API and artifact reuse semantics,
  external author validation, reversible migration sub-phases, docs/generated
  help gates, severity enum, exact metadata matching, and numeric suppression
  bounds.

Unfixable items: none documented; all [Blocker] and [Major] synthesis findings
were addressed in the design text.
