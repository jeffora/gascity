# Apply Summary

Attempt: 21
Verdict applied: block -> iterate
Design document: `engdocs/design/formula-compiler-requirements.md`
Synthesis: `.gc/design-reviews/ga-j195cx/attempt-21/synthesis.md`

Addressed findings:

- [Blocker] Parser and validation contract: added an executable
  `compiler_requirements_matrix.yaml` schema, complete coverage dimensions,
  diagnostic count/order rules, and a v2 construct registry completeness test.
- [Major] Compiler boundary and durable writes: named `internal/sourceworkflow`
  as the sole workflow-root predicate owner, removed the ambiguous successor
  wording, and added accepted compile artifact identity plus retention rules.
- [Major] Rollout, compatibility, and alias retirement: added a pinned
  compatibility corpus, supported-reader artifact, `bd >= 1.0.0` probe floor,
  strict native-vs-`bd` parity rules, and a supersession rule for stale
  `GC_NATIVE_FORMULA=false` runtime rollback language.
- [Major] Pack provenance and external migration: added `ResolvedFormulaSource`
  ownership, stable `migration_hints` JSON, and `[pack].requires_gc` enforcement
  for first-party requires-only graph packs.
- [Major] In-flight, diagnostics, and documentation gates: preserved the
  existing accepted-artifact downgrade policy, zero-write boundaries, typed
  diagnostic/event projection, generated docs/help gates, and stale-guidance
  checks while tightening the executable contracts around them.

Unfixable items: none documented; all [Blocker] and [Major] synthesis findings
were addressed in the design text.
