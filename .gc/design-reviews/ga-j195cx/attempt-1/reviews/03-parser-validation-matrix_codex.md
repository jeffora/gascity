# Parser Validation Matrix Review (Codex)

Persona: Priya Zielinski
Mandate: requirement grammar, parser edge cases, v2 construct enforcement, test matrix
Verdict: approve-with-risks

## Summary

The design is directionally correct: `internal/formula` owns requirement interpretation, callers consume typed results, unsupported v0 grammar is rejected, and host capability failures are separated from parser failures. I would approve the direction only with parser-matrix hardening before the first parser PR lands. The remaining risk is not conceptual; it is that several TOML and traversal edge cases can collapse to the same Go shape before diagnostics, making the promised stable diagnostic contract difficult to enforce.

## Findings

### Major: Raw TOML authoring shapes are not fixture-locked

The design says v0 accepts omitted, `>=1`, and `>=2`; rejects unsupported expressions and non-string requirement values; and treats unknown `[requires]` keys as errors (`design-after.md:156-179`). It also requires source spelling to survive into diagnostics (`design-after.md:270-281`, `design-after.md:393-412`). That is not enough to implement deterministic parser behavior with the current TOML stack, because `internal/formula.Parser.ParseTOML` decodes straight through `toml.Unmarshal` into the Go model (`internal/formula/parser.go:148-153`).

Required change: add a raw-shape matrix before implementation. At minimum it must state the exact behavior, diagnostic code, source key, source value, and source provenance for:

- canonical `[requires]` table with `formula_compiler`
- top-level dotted assignment `requires.formula_compiler = ">=2"`
- inline table `requires = { formula_compiler = ">=2" }`
- root `requires` with scalar, array, bool, int, float, table, and inline-table conflicts
- duplicate scalar keys and dotted/table collisions in both source orders
- unknown axes in table, dotted, and inline-table forms
- nested or misplaced tables such as `[steps.requires]`, `[[steps.requires]]`, `[children.requires]`, `[loop.body.requires]`, and `[compose.requires]`

Without this, the implementation can accidentally accept a non-canonical TOML shape, fail with an untyped decoder error where the design promises `formula.requirement_invalid_type`, or lose enough raw spelling that the CLI/API diagnostic contract cannot be projected consistently.

### Major: V2-only construct coverage lacks condition-state and contribution dimensions

The construct registry covers `steps`, `children`, `loop.body`, and expansion `template` (`design-after.md:196-211`), and normalization is supposed to run after resolution but before transforms (`design-after.md:213-227`). The current code already has multiple transform entry points and condition filters in the fragment path (`internal/formula/fragment.go:52-102`), while the existing graph-requirement scan only walks steps, children, loop bodies, and templates (`internal/formula/types.go:890-940`).

Required change: make the validation matrix explicitly cross these dimensions:

- construct identity: `check`/`ralph`, `retry`, `on_complete`, graph metadata key, graph metadata value
- authored location: root step, child, nested child, loop body, expansion template, expansion contribution, map contribution, aspect contribution, inline `expand`
- condition state: authored and materialized, authored but condition-disabled, and materialized through an expansion/aspect
- caller path: root compile, expansion fragment compile, sling/API/order preflight, convergence preflight
- source attribution: root formula, inherited parent, expansion formula, aspect formula, and loop/body source location

The design should also make the policy explicit: an authored v2-only construct behind a false condition either still requires `>=2` because validation is source-based, or it is intentionally materialization-based. Today the prose implies source-based validation, but the matrix does not lock the behavior. If this stays implicit, a formula can compile or fail depending on transform order and runtime variables.

### Major: Diagnostic precedence is too coarse for combined defects

The deterministic order is useful (`design-after.md:185-194`), but it is not precise enough to write count-locked tests. For example, a file can contain an unknown requirement axis, an invalid requirement type, a `contract` conflict, and a v2-only construct with a missing requirement. The design does not say whether the compiler emits all diagnostics or stops at the first class, how diagnostics are sorted within a class, or how TOML decoder errors from duplicate/collision syntax relate to semantic `formula.*` diagnostics.

Required change: define a `precedence-ordering` suite with concrete rows for multi-defect formulas. Each row should lock the expected diagnostic list, order, fatal/warning split, and whether later checks are suppressed after a decode failure. Add a `SourceLocation` or equivalent location field to `Diagnostic`; `SourceKey` alone cannot distinguish root `[requires]` from nested `loop.body[0].requires` or from a contributed expansion formula.

### Minor: Legacy JSON behavior remains undefined

The parser still supports `.formula.json` through `Parser.Parse` (`internal/formula/parser.go:130-145`), while the design frames the new requirement surface almost entirely as TOML (`design-after.md:166-176`). If JSON remains a supported formula input during the migration window, the parser matrix needs JSON rows for omitted requirements, valid `requires`, invalid types, unknown axes, `contract` compatibility, and conflict behavior. If JSON is intentionally being retired for this capability, that policy should be explicit and paired with tests/docs so callers do not get a silent divergence between TOML and JSON formulas.

## Required Changes Before Parser Implementation

1. Add a machine-readable validation matrix with literal expected counts per suite; reject placeholder/generated counts in checked-in artifacts.
2. Add raw TOML shape rows for table, dotted-key, inline-table, duplicate, collision, unknown-axis, invalid-type, and misplaced nested requirement declarations.
3. Add construct traversal rows for root, inherited, child, loop, expansion, map, aspect, inline expansion, fragment compile, and convergence paths.
4. Add condition-state rows and state whether validation is source-based or materialization-based.
5. Add combined-defect precedence rows that lock diagnostic order, suppression, and source attribution.
6. Decide and test the legacy JSON policy in the same parser PR.

## Questions Answered

- Unsupported expressions such as exact `2`, `>=2.1`, empty strings, and malformed ranges are rejected in principle, but the exact raw TOML spellings that produce those values are not yet fully specified.
- Empty `[requires]`, unknown keys, and unknown `contract` values are addressed in prose, but not yet locked to a complete fixture matrix.
- V2-only constructs are required to fail without a v2 requirement, but the design still needs explicit coverage for false conditions, contributed formulas, fragments, and caller paths.

## Final Verdict

Approve with risks. The design has the right ownership model and diagnostic vocabulary, but parser implementation should not start until the raw-shape, traversal, condition-state, and precedence matrices are concrete enough that tests can prove the public contract rather than merely sample it.
