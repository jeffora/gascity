# Apply Summary

Attempt: 53
Source bead: ga-j195cx
Apply bead: ga-qyx4v8h
Global synthesis verdict: block
Applied verdict: iterate

## Changes Applied

- Added a single typed diagnostic attribution contract with primary, requirement, host, and pack source objects, then tied CLI, Huma/API, generated TypeScript, dashboard, events, warning keys, and release reports to that object.
- Tightened raw requirement capture: TOML uses BurntSushi as typed decoder and parser-error authority while source bytes, line, column, shape, and duplicates come from a compiler-owned raw pass; JSON stays in scope while the JSON loader exists and must use duplicate-aware raw capture.
- Added accepted projection snapshots to accepted artifacts so convergence, fanout, retry, scope-check, workflow-finalize, repair, and missing-child paths can project from persisted compiler output without source access or a subset parser.
- Clarified rollout sequencing with an explicit first-party pack-floor declaration phase before resolver enforcement, a Phases 2-7 dual-declaration invariant, Phase 4 caller-migration rollback controls, alias-window clock rules, and active-root drain/repair/waiver requirements before requires-only conversion.
- Split host-downgrade behavior between current-host validation for new or changed compiles and persisted-artifact reuse for same-identity roots; added operator-visible behavior for legacy roots, retry, fanout, scope-check, workflow-finalize, and convergence.
- Added deterministic first-party/external classification, a single `--alias-removal-gate` command, unreadable-inventory fail-closed behavior, per-occurrence caller manifests, first-party inventory artifacts, stale-guidance checks for formula `version`, and release placeholder checks.
- Made future-capability invariants binding: additive-only compiler capability evolution, frozen released axis grammars, schema-version rules, `RequirementSource` as provenance only, generalized multi-axis construct capability, and old-reader cross-product fixtures.
- Cleaned up residual details by defining deterministic compile-artifact spillover limits and requiring remediation text to be either plain prose or fixture-valid TOML/JSON.

## Finding Coverage

- [Blocker] Diagnostics Lack A Single Typed Attribution Contract: addressed.
- [Blocker] Convergence Cannot Yet Project From Accepted Artifacts Alone: addressed.
- [Blocker] Rollout Sequencing Leaves Compatibility Gaps: addressed.
- [Blocker] In-Flight Graph Workflow Semantics Are Not Crisp: addressed.
- [Major] Parser And Fixture Contracts Are Still Too Loose: addressed.
- [Major] Contract Migration And External Pack Evidence Are Incomplete: addressed.
- [Major] Documentation And Terminology Still Teach Old Behavior: addressed.
- [Major] Future-Capability Invariants Need To Be Binding: addressed.

No blocker or major finding was documented as unfixable.

## Validation

- Confirmed `engdocs/design/formula-compiler-requirements.md` matched `attempt-53/design-before.md` before editing.
- Saved `design-after.md`.
- Saved `design.diff`.
- Checked Markdown code fence balance: 94 fences, even parity.
