# Ibrahim Park - Codex

**Verdict:** block

**Top strengths:**
- The design correctly makes Gas City, not `bd`, the authoritative owner for future formula compilation work.
- The proposed native `formula.Compile` / `molecule.Instantiate` split leaves room for formula evolution without pushing compilation back into the task-store primitive.
- The phased migration and rollback toggle are pragmatic for moving a live orchestration surface without a flag day.

**Critical risks:**
- [Blocker] The design says all future formula features land in Gas City's `internal/formula/`, while also promising full backward compatibility and a `GC_NATIVE_FORMULA=false` fallback to the frozen `bd` compiler. It does not define a formula capability declaration, minimum-reader check, or unknown-field policy. A formula that uses a new Gas City-only field could be silently sent through the `bd` fallback or an older Gas City binary and either miscompile or fail with an incidental parser error.
- [Major] The design treats "Changing the `.formula.toml` file format" as a non-goal, but future-feature ownership necessarily creates a file-format compatibility boundary. Without an explicit rule for reserved future fields, unsupported top-level keys, and unsupported nested keys, the first post-migration formula feature will have to invent compatibility behavior under implementation pressure.
- [Major] The rollback toggle is underspecified for future capabilities. It says `GC_NATIVE_FORMULA=false` falls back to `Store.MolCook`, but it does not say whether formulas requiring native-only features must be rejected before fallback, skipped with an actionable diagnostic, or allowed to fail inside `bd`.
- [Minor] The context docs still describe `version` as an optional formula marker and do not document any reader/compiler capability metadata. The migration design does not state whether this work deliberately defers that cleanup or must update the docs before the native compiler becomes authoritative.

**Missing evidence:**
- No table defines the semantics for omitted capability metadata, default capability, native-only capability, or unsupported future capability declarations.
- No validation matrix covers unknown top-level keys, unknown nested keys under future capability metadata, or formulas authored for a newer compiler than the running binary.
- No caller behavior is specified for sling, orders, API, convergence, and fallback mode when a formula declares or uses an unsupported future capability.
- No golden fixture category proves that formulas with future-reserved fields are rejected deterministically instead of being ignored or passed to `bd`.

**Required changes:**
- Add a small compatibility contract before approving the migration: define how a formula declares the minimum compiler/reader capability it needs, what omission means, and how the active binary rejects unsupported future requirements.
- Specify the unknown-key policy at the formula boundary. If full backward compatibility requires preserving today's permissiveness, state exactly which known extension points remain permissive and which future capability roots are reserved and must fail closed.
- Define fallback behavior for `GC_NATIVE_FORMULA=false`: native-only formulas must be detected before invoking `bd`, with an actionable error that explains the formula requires the native Gas City compiler.
- Add tests for omitted/default capability, native-only capability under fallback, unknown future capability axes, and newer-than-supported compiler requirements across at least CLI sling, orders, API, and convergence.
- Update `docs/reference/formula.md` and `engdocs/architecture/formulas.md` to document the compatibility boundary or explicitly mark it as deferred work that must land before any post-migration formula feature.

**Questions:**
- Is this migration intended to preserve the existing formula surface only, with capability declarations deferred until the first new formula feature, or is the native compiler supposed to establish that boundary now?
- Should the fallback toggle be prohibited for formulas that require native-only features, or should the toggle be removed before any such feature ships?
- Are unsupported future fields supposed to be rejected by the TOML parser, by a normalized formula validation stage, or by caller-specific diagnostics?
