# Ibrahim Park

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] The design has the right fail-closed direction, but the long-term meaning of `formula_compiler = ">=N"` is still underspecified. It must state whether compiler capabilities are monotonically backward-compatible minimums, or whether the comparator-looking syntax is only a closed enum surface.
- [Major] Future requirement axes need a stronger contract before `[requires]` becomes public API. Unknown keys should fail closed, but the design still needs to choose whether that is permanent, whether any vendor or experimental namespace exists, and how future axes such as tools, runtime, pack format, or agents are reserved and introduced.
- [Major] Requirement provenance and aggregation are not pinned tightly enough for future axes. The design describes inherited, expansion, and aspect requirements, but the normalized/provenance model and validation matrix do not yet cover lowering, raising, multiple contributors, host-disabled aggregated requirements, or conditional expansion boundaries.
- [Major] Diagnostics are not actionable enough for forward compatibility. `formula.compiler_requirement_unsupported` conflates syntactically invalid values such as `>=2.0` with syntactically valid but unsupported future capabilities such as `>=3`, and unknown-axis diagnostics should expose the rejected axis, source key/value, and supported axis list.
- [Major] The compatibility matrix must explicitly cover `formula_compiler = ">=1"`. Omitted requirements, empty `[requires]`, and explicit `>=1` are intended to be behaviorally equivalent, but source attribution and v2-only construct handling are not pinned by rows or tests.
- [Major] The formula-domain host capability contract should not carry contradictory sources of truth. Codex identifies `HostCapabilities{FormulaCompiler, FormulaV2}` as a blocker because impossible states such as compiler v2 with `FormulaV2=false` are representable.
- [Minor] Capability constant naming should be revision-based, not feature-based. Both reviews object to `CompilerCapabilityGraphV2`; use a name such as `CompilerCapabilityV2` and keep graph-workflow meaning in comments/docs.
- [Minor] Numeric capability stability, zero-value handling, future-axis diagnostic code conventions, and snake_case axis naming are not documented clearly enough to guide future contributors.

**Disagreements:**
- Claude's verdict is `approve-with-risks`; Codex's verdict is `block`. Assessment: the persona verdict is `block` because the contradictory host-capability API and missing aggregation/provenance contract are structural issues that should be fixed in the design before implementation.
- Claude treats the current fail-closed grammar and normalized compiler capability as a strong enough base if the future semantics are documented. Codex is less comfortable with the scalar v0 model because future axes will likely require per-axis/per-source records. Assessment: the design can keep a narrow v0 implementation only if it explicitly marks the scalar shape as single-axis and defines what must change before another axis ships.
- Claude asks for an explicit vendor or extension namespace decision. Codex is comfortable with strict rejection if future axes are typed additions and no raw passthrough exists. Assessment: either choice is acceptable, but silence is not.
- Claude wants a clear answer on whether the `>=N` syntax is honest minimum-version syntax or an enum-like closed allowlist. Codex does not object to the string surface directly. Assessment: keep the string only if the monotonicity and transition policy are normative.

**Missing evidence:**
- No Gemini review artifact is present for this persona.
- No compatibility-matrix row for `[requires] formula_compiler = ">=1"` across old binary, new binary with formula v2 disabled, and new binary with formula v2 enabled.
- No validation rows for parent `>=2` with child `>=1`, parent `>=1` with child `>=2`, root omitted with expansion/aspect `>=2`, host-disabled aggregated requirements, duplicate maximum contributors, or non-contributing conditional expansions.
- No normative source/provenance table for omitted, empty `[requires]`, explicit `>=1`, explicit `>=2`, legacy `contract`, and dual declarations, including inherited source behavior.
- No formal accepted-string grammar covering whitespace, leading zeros, unicode whitespace, or future integer capabilities.
- No split remediation examples for invalid syntax, unsupported future compiler capability, and unknown requirement axis.
- No explicit policy for `CompilerCapability(0)` or for immutable persisted capability numbers.

**Required changes:**
- Remove `FormulaV2` from the formula-domain `HostCapabilities` type, or move it to an edge-only config conversion type so contradictory host states cannot be represented. Add tests for the config-edge mapping from `[daemon] formula_v2` into canonical compiler capability.
- Add a "Capability monotonicity" subsection. State whether `>=N` means a future host capability `N+1` always satisfies requirements for `N`, or replace the comparator-looking grammar with a non-comparator representation before publishing it.
- Split unsupported compiler requirement diagnostics into invalid syntax and syntactically valid but unsupported future capability. The latter should tell operators to upgrade `gc`, not edit the formula to a lower requirement.
- Add a forward-axis policy for `[requires]`: strict rejection forever or a precisely defined extension namespace, reserved future top-level snake_case axis names, typed normalized state, provenance, diagnostics, persisted metadata, docs, and tests for every new axis.
- Make requirement provenance per-axis/per-source now, or explicitly declare the v0 scalar model temporary and require it to be widened before any second axis ships. Behavioral callers should branch on normalized requirement satisfaction, not on source strings.
- Add validation and diagnostic matrix rows for `>=1`, v2-only constructs with `>=1`, inherited lowering/raising, expansion/aspect aggregation, host-disabled aggregation failure, duplicate maximum contributors, and conditional expansion contribution boundaries.
- Rename `CompilerCapabilityGraphV2` to a revision-level name such as `CompilerCapabilityV2`, with comments explaining that v2 currently enables graph-workflow compilation.
- Define the formal accepted-string grammar, rejection fixtures for whitespace/leading zeros/unicode variants, numeric capability stability, `CompilerCapability(0)` behavior, and generic versus per-axis diagnostic code conventions.
