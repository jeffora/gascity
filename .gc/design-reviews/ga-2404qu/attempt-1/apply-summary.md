# Apply Summary

Attempt: 1
Synthesis: `.gc/design-reviews/ga-2404qu/attempt-1/synthesis.md`
Global verdict: `block`
Applied verdict: `iterate`

Updated `.gc/design-review-inputs/core-gastown-pack-migration/design.md` to
address the global blocker and major findings. The design now adds:

- Review-gated migration invariants and green vertical rollout slices.
- Public Gastown replacement-first gates with immutable commit, pin, behavior
  inventory, and Gas City packcompat proof.
- Production Core-load provenance assertion and a `cmd/gc` direct
  `config.Load*` scanner.
- Doctor/import-state safety contract for preflight, scoped TOML edits,
  failure atomicity, fork/custom-source handling, and runtime-state diagnostics.
- Core role-neutral maintenance/notification contract and token scanner.
- Synthetic cache/public pin strategy for retiring bundled Gastown and
  Maintenance aliases.
- Maintenance retirement runtime table.
- Bootstrap test-only fixture contract and hidden-dependency guards.
- Current-tree docs/DX inventory, canonical wording, and golden/lint gates.
- Behavior-oriented tests, provider-pack continuity, Core integrity checks, and
  cross-pack ownership decisions.

Saved artifacts:

- `.gc/design-reviews/ga-2404qu/attempt-1/design-after.md`
- `.gc/design-reviews/ga-2404qu/attempt-1/design.diff`
- `.gc/design-reviews/ga-2404qu/attempt-1/apply-summary.md`
