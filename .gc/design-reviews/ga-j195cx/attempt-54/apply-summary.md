# Apply Summary

Attempt: 54
Source bead: ga-j195cx
Apply bead: ga-w8q6srq
Global synthesis verdict: approve-with-risks
Applied verdict: iterate

## Changes Applied

- Defined same-identity accepted-artifact reuse for host downgrade, including exact identity checks, current-host behavior, unsupported artifact handling, concurrency, and dashboard/API visibility fixtures.
- Normalized accepted-artifact identity into `CompileIdentity`, `AcceptedArtifactIdentity`, and `CompileWriteIntent`, and clarified that current host satisfaction is required for new or changed compiles but not for same-identity persisted-artifact reuse.
- Added per-symbol durable-writer end states, concrete legacy helper retirement decisions, and manifest requirements for `end_state`, owners, expiry phases, replacement APIs, and blocking tests.
- Expanded requirement grammar and validation matrix closure for `>=0`, `>=-1`, leading zeros, decimals, signed strings, overflow, control/Unicode values, TOML/JSON shape errors, mixed axes, and transitive v2 construct contributions.
- Added `formula.version_misuse` handling for `version = 2` paired with v2-only syntax and missing compiler requirements.
- Specified durable grouped diagnostic state, scan-series grouping, occurrence-count updates, reset rules, restart/controller handoff, grouped-state write failure behavior, and CLI/API/dashboard/report fixtures.
- Made alias-removal gate exit semantics auditable for legacy-only, dual-declared, unsupported future, unreadable inventory, stale external support, auth/unreachable sources, shadowed external formulas, SHA pins, stale guidance, and malformed input.
- Added doctest requirements for `docs/reference/formula.md`, a "which key do I edit" table, copy-paste-safe external-pack snippets, and same-branch generated artifact gates.
- Clarified convergence package-boundary flow, field ownership, exported-symbol migration, and `TestNoConvergenceSubsetParserUse` allowlist rules.
- Added a strict-review risk register documenting release-blocking executable artifacts that must exist before the workflow can approve under strict mode.

## Finding Coverage

- [Blocker] Host-Downgrade Continuation Is Not Implementable Yet: addressed with a normative same-identity reuse contract.
- [Major] Durable Writers Need an Accepted-Artifact-Only Boundary: addressed with per-symbol end states and zero-write assertions.
- [Major] Caller Inventory and Legacy Helper Retirement Are Not Yet Executable: addressed with manifest `end_state` requirements and helper-specific decisions.
- [Major] Requirement Grammar and Validation Matrix Need More Normative Rows: addressed with exact edge rows and count-lock derivation rules.
- [Major] Operator Diagnostics and Grouping Semantics Are Underspecified: addressed with `FormulaDiagnosticGroupState` and fixture rules.
- [Major] Alias Removal and External-Pack Release Gates Need a Single Auditable Contract: addressed with exit/report semantics and conservative missing-artifact behavior.
- [Major] Convergence Projection Ownership and Subset Parser Retirement Are Ambiguous: addressed with package-boundary flow, field ownership, and symbol migration.
- [Major] Documentation and DX Gates Are Not Yet Copy-Paste Safe: addressed with doctest and same-branch docs/generated artifact gates.
- [Major] Forward-Compatibility and Accepted-Artifact Identity Need Normative Structs: addressed with normalized identity structs and same-identity reuse checks.
- [Minor] Formula `version` and New-Author Misuse Need Sharper Diagnostics: addressed with `formula.version_misuse`.

No item was documented as unfixable. Because strict mode is enabled and the global verdict was `approve-with-risks`, residual risks are documented in the design and the applied verdict is `iterate`.

## Validation

- Confirmed `engdocs/design/formula-compiler-requirements.md` matched `attempt-54/design-before.md` before editing.
- Saved `design-after.md`.
- Saved `design.diff`.
- Checked Markdown code fence balance: 102 fences, even parity.
- Checked for trailing whitespace: none found.
- Ran `git diff --no-index --check` against the attempt baseline; no whitespace errors were reported.
