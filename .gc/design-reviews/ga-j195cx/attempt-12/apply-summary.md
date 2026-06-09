# Apply Summary

Source synthesis: `.gc/design-reviews/ga-j195cx/attempt-12/synthesis.md`
Global verdict: `block`
Applied verdict: `iterate`

## Changes Applied

- Addressed host-capability and preflight blockers by making `HostCapabilities`
  a single typed capability input, making `NormalizedRequirements`
  formula-owned, adding a formula-domain satisfaction API, and defining
  zero-durable-write preflight contracts for root, attached molecule, order,
  retry, fanout, and convergence paths.
- Addressed parser/grammar blockers by adding a raw requirement capture model,
  accepted TOML shape matrix, detailed v2-only construct registry, and
  combined-defect precedence examples.
- Addressed diagnostic projection blockers by adding surface parity contracts,
  typed Huma/API/dashboard requirements, registered event payload contracts, and
  deduplication rules.
- Addressed migration blockers by adding measurable release gates, the minimum
  binary floor artifact, bd/native parity gate, external pinned-pack plan,
  dual-stamp retirement criteria, and alias removal criteria.
- Addressed provenance blockers by defining structured `CompileResult`
  provenance, durable root metadata semantics, reproducibility policy effects,
  and canonical workflow-root predicate behavior.
- Addressed convergence/fanout blockers by choosing convergence as a typed
  projection from canonical `CompileResult` plus convergence-only validation,
  and adding explicit fanout/convergence zero-write migration rows.
- Addressed major documentation and forward-compatibility findings by adding
  requirement-surface comparisons, modern and dual-declared TOML examples,
  stale-guidance gates, and monotonic capability/unknown-axis rules.
- Addressed minor cleanup by renaming the capability constant in the design to
  `CompilerCapabilityV2`, clarifying bounded warning cadence, and defining
  operational consequences for non-reproducible local formula inputs.

## Unfixable Items

None. All synthesis [Blocker] and [Major] findings were addressed in the design
document.

## Artifacts

- Design after changes: `.gc/design-reviews/ga-j195cx/attempt-12/design-after.md`
- Diff: `.gc/design-reviews/ga-j195cx/attempt-12/design.diff`
