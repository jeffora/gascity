# Apply Summary

Attempt: 1
Global verdict: block
Applied verdict: iterate

Updated `engdocs/design/formula-compiler-requirements.md` to address every
[Blocker] and [Major] synthesis finding.

Changes applied:

- Defined the closed v0 parser grammar for `requires.formula_compiler`,
  including exact accepted strings, rejected strings, TOML type behavior,
  unknown-key policy, empty-table behavior, diagnostic order, v2-only construct
  registry, inheritance, expansion, and aspect normalization.
- Added the canonical `internal/formula` compile-result contract and caller
  inventory, with a required static guard against raw `contract`,
  `requires.formula_compiler`, and `gc.formula_contract` behavioral consumers.
- Replaced the undefined migration window with an old/new compatibility matrix,
  strict dual-declaration policy for first-party formulas, bd-shellout
  compatibility requirements, measurable alias-removal criteria, and an owner.
- Added a shared diagnostic contract with stable codes and CLI, API,
  dashboard, order, controller, and convergence projection rules.
- Resolved public open questions: exact `"2"` is rejected, deprecation warnings
  are compile diagnostics projected by callers, and formula `version` remains
  accepted only as legacy metadata with optional warning.
- Defined canonical persisted metadata, transitional dual-stamping,
  provenance/validation reporting, local path reproducibility reporting, and
  consumer migration requirements.
- Split rollout into reversible phases with rollback notes and docs/example
  sequencing.
- Defined in-flight, retry, order-wisp, and convergence behavior when
  `[daemon] formula_v2` changes.

Unfixable items: none.
