# Priya Zielinski - Claude

**Verdict:** approve-with-risks

The grammar is well-bounded, the diagnostic order is deterministic, and the
v2-only construct registry is the right primitive. The risks cluster in three
places: (1) the v2-only registry has a malformed row that two implementers
will read differently; (2) accepted/rejected expressions are listed by
example rather than by closed grammar, which leaves whitespace and Unicode
edge cases to implementer judgment; and (3) the test contract names topics
but never mandates a Cartesian matrix, which is exactly the persona's
"hand-written per case" red flag. None block implementation. Each invites
silent drift between what the document says and what the parser accepts.

**Top strengths:**

- The rejected-form list is closed, not open: `>=1` and `>=2` are the only
  accepted strings and every neighboring spelling (`"2"`, `">= 2"`,
  `" >=2"`, `">=2 "`, `">=2.0"`, `">=2.1"`, `">=3"`) is enumerated as a
  rejection. The "reject every other value instead of accepting a future
  range optimistically" rule is the right default and is stated explicitly.
- The deterministic diagnostic order (TOML errors → unknown axes/types →
  unsupported expression → contract conflict → missing v2 requirement →
  host capability) gives test authors and operators a single canonical
  surfacing sequence and eliminates a class of "which error wins" bugs
  across CLI, API, and dashboard projections.
- v2-only construct enforcement is anchored in a registry that lists both
  the construct and the locations scanned (`steps`, `children`,
  `loop.body`, expansion `template`), and the rule that "expansion and
  aspect formulas are parsed and normalized independently" with the
  compiled root taking the maximum requirement prevents an aspect from
  sneaking v2 fields past a v1 root.

**Critical risks:**

- **[Major] The v2-only construct registry has a duplicated, ambiguous
  row.** Two rows are titled "Graph workflow metadata keys": one lists
  only locations and no key names (`steps`, `children`, `loop.body`,
  expansion `template`), and another lists specific key names
  (`gc.scope_name`, `gc.scope_role`, `gc.scope_ref`,
  `gc.continuation_group`, `gc.on_fail`). A reader cannot tell whether
  the first row means "any metadata key in those locations triggers v2"
  (a wildly broad rule that would make the named-key row redundant) or
  whether it is an editing stub that should have been deleted. Two
  implementations will diverge from the same document. This is the
  registry that the entire `formula.compiler_requirement_missing`
  diagnostic depends on, and the persona's red flag "v2-only fields
  compile without a requirement" lands directly in this gap.
- **[Major] The accepted/rejected expression list is example-driven, not
  exhaustive.** The Resolved public syntax block enumerates rejected
  forms but the validation matrix at the same time says "unsupported
  string **such as** `2`, `>= 2`, `>=2.0`, `>=3`, or empty string." The
  phrase "such as" leaves an implementer free to write a permissive
  parser that accepts `>=02`, `>=2\t`, `>=2.0.0`, `>=v2`, `>=2-rc1`,
  `>= 2 ` (multiple internal whitespaces), or ` >=2` (non-breaking
  space). Any of these is the "unknown ranges accepted silently" red
  flag. The accepted v0 surface is two byte sequences: `>=1` and
  `>=2`. The doc must say so as a byte-exact rule, not enumerate
  rejected variants.
- **[Major] Test categories are named but matrix coverage is not
  enforced.** The contract section lists "table-driven tests for
  accepted strings, rejected strings, invalid TOML types, unknown
  `[requires]` keys, empty tables, legacy contract compatibility,
  conflicts, missing requirements, inherited requirements,
  expansion/aspect aggregation, loop bodies, children, and unsupported
  future requirements." That is a list of axes, not a matrix. Without a
  stated Cartesian product (expression value × TOML type × `[requires]`
  presence × `contract` value × source location × v2-construct presence
  × host capability), implementers will hand-write per-case tests and
  miss combinations such as "unknown `[requires]` key in an expansion
  that also has a v2-only construct" or "valid `>=2` on an aspect that
  contributes no v2 constructs to a v1 root." Hand-written-per-case is
  exactly the persona's third red flag.
- **[Major] The empty value `formula_compiler = ""` is unspecified.** The
  Resolved public syntax block lists rejected forms but never the empty
  string. The closest rule is "An omitted `[requires]` table and an
  empty `[requires]` table both mean the default compiler capability,"
  which addresses table presence, not value emptiness. A naive
  implementation could either treat `""` as omission and accept default
  capability or reject it as unsupported. The validation matrix lists
  empty string under "unsupported," but the prose contradicts that
  reading; a strict reader cannot tell which is canonical.
- **[Major] No positive enforcement that every caller calls
  `CompileWithResult` before durable writes.** The static guard catches
  *new behavioral uses of legacy raw fields* (`Contract`,
  `declaresGraphV2Contract`, `Requires.FormulaCompiler`,
  `gc.formula_contract`), which is necessary but not sufficient. It
  cannot catch a new bead-creating call site that simply does not
  preflight at all. The caller inventory and migration table enumerate
  today's callers, but the document has no mechanism — typed wrapper,
  generated registry, or test that walks `Store.Create` callers from
  formulas — to guarantee tomorrow's caller is gated. The lane question
  "regardless of caller path" loses the moment an unrelated subsystem
  adds a formula consumer without preflight.
- **[Minor] Whitespace handling is enumerated case-by-case rather than
  by rule.** The design rejects `">= 2"`, `" >=2"`, and `">=2 "` but
  does not state the underlying rule (byte-exact match against the
  accepted set). Tabs (`"\t>=2"`), non-breaking spaces (`" >=2"`),
  trailing newlines (`">=2\n"`), and Unicode-equivalent whitespace are
  not addressed. A natural Go implementation will reach for
  `strings.TrimSpace`; without an explicit "no normalization" directive,
  the rejection rule and the implementation will silently drift.
- **[Minor] `version` is missing from the deterministic diagnostic
  order.** The diagnostic codes table introduces
  `formula.version_deprecated` as a warning, and the Persisted Metadata
  section confirms `version` is preserved only as legacy metadata. But
  the six-step diagnostic order in Parser And Validation Contract never
  mentions where `version` is detected. The document should say whether
  `version` is checked before or after the unknown-axis check, and
  whether a formula that fails the expression check still emits the
  `version_deprecated` warning.
- **[Minor] Construct-presence semantics for v2-only fields are
  undefined.** Does `check = {}` (empty inline table) trigger
  `formula.compiler_requirement_missing`? Does
  `retry = { max_attempts = 0 }`? Does `metadata = { "gc.kind" = "" }`?
  The persona's lane question "v2-only fields fail when the requirement
  is omitted" cannot be answered as written: presence-of-key vs.
  presence-of-non-zero-value is not specified. A test matrix cannot be
  generated until this is pinned.
- **[Minor] No diagnostic code is reserved for "child requirement lower
  than parent."** The inheritance section says a child "may not lower a
  parent requirement," but no code in the diagnostic registry covers it.
  Either it folds into `formula.compiler_requirement_conflict` (then say
  so) or it needs its own code. The validation matrix has no row for
  this case.
- **[Minor] "Default compiler capability, equivalent to `>=1`" creates a
  third entry that may or may not be a user-facing accepted form.** The
  Requirement Expression table accepts `>=1` as an explicit value, but
  it gives the same behavior as omitted and an operator who writes it
  learns nothing. The design should state explicitly whether `>=1` is
  accepted silently, accepted with an informational diagnostic, or
  treated as a stylistic preference.

**Missing evidence:**

- No formal grammar for accepted expression strings. The accepted set is
  small enough to be expressed as the regex `^>=(1|2)$`, which would be
  unambiguous in a way the prose is not.
- No statement of whether validation is fail-fast on first defect or
  collects all diagnostics in a single pass. The ordering rule is
  consistent with both interpretations; the matrix coverage and
  `OnceKey` suppression semantics depend on which one is canonical.
- No mechanism (typed interface, generated registry, fixture file, CI
  guard, parity test) is named for keeping the v2-only construct
  registry in sync with the actual graph-compiler behavior. If a new
  `gc.kind` value is added in `internal/molecule`, what fails until the
  registry table is updated?
- No specification of how `[requires]` raw position/type preservation is
  implemented. `BurntSushi/toml`'s `MetaData` API or a manual two-pass
  decode is implied but never named.
- No explicit count of total required matrix combinations. The matrix
  has 11 rows; the documented axes (3 `contract` states × 7 `[requires]`
  states × 2 v2-construct × 2 host) imply 84 cells. Which subset is the
  floor and which are derivable by composition?
- No statement on TOML inline-table form. Both `[requires]` and
  `requires = { formula_compiler = ">=2" }` parse to the same
  structure, but the document never says whether both spellings are
  accepted.
- No worked example for the inheritance edge case "v1 root + v2
  expansion compiles to a v2 root requirement." The two rules
  (inheritance and aspect aggregation) together are correct, but a
  fixture demonstrating their interaction is not called out.

**Required changes:**

1. **Resolve the duplicate "Graph workflow metadata keys" row** in the
   v2-only construct registry. Either merge them into a single row with
   a complete key list and explicit scan locations, or rename the
   location-only row to make its scope unambiguous. The likely intent is
   a single row listing the named keys (`gc.scope_name`,
   `gc.scope_role`, `gc.scope_ref`, `gc.continuation_group`,
   `gc.on_fail`) and where each is scanned; pick that and remove the
   redundancy.
2. **Rewrite the requirement expression spec as a closed byte-exact
   grammar.** Add one sentence: "Comparison is byte-exact against the
   accepted set `{">=1", ">=2"}`. No whitespace normalization, case
   folding, Unicode normalization, or substring matching is performed.
   The implementation must not call `strings.TrimSpace`,
   `strings.ToLower`, or any equivalent normalization on the value
   before comparison." Then drop the per-whitespace-variant rejection
   list in favor of "all other strings, including any whitespace
   variation, fail validation."
3. **State a normative test-coverage rule.** Replace the prose test
   list with an explicit matrix specification. Name the dimensions and
   the expected outcome for each Cartesian cell. Minimum dimensions:

   - **Expression value:** `omitted`, `""`, `">=1"`, `">=2"`, `">=2.0"`,
     `">=2.1"`, `">=3"`, `"2"`, `">= 2"`, `" >=2"`, `">=2 "`, `">2"`,
     `"<2"`, `"=2"`, `"~>2"`, `">=2,<3"`, `"\t>=2"`, `" >=2"`.
   - **TOML type:** string, integer, float, boolean, array, table,
     dotted-table, inline-table.
   - **`[requires]` table state:** absent, present-empty, present with
     `formula_compiler` only, present with unknown sibling key, present
     with unknown key only.
   - **`contract` field state:** absent, `"graph.v2"`, `"graph.v3"`,
     `""`, other-string.
   - **Source location:** root formula, child step, `loop.body`,
     expansion `template`, aspect formula, `extends` parent.
   - **v2-only construct presence:** absent, `check`, `retry`,
     `on_complete`, each `gc.kind` value, each named metadata key.
   - **Host capability:** `formula_v2 = true`, `formula_v2 = false`.

   Specify which cells are required to be tested and which are derivable
   by composition. Require a single table-driven harness in
   `internal/formula` that iterates a fixture struct naming every
   dimension explicitly, and a CI assertion that fails if a documented
   axis value lacks a row.
4. **Define the empty-value rule.** State explicitly: "`formula_compiler
   = ""` is rejected with `formula.compiler_requirement_unsupported`.
   An empty string is a malformed expression, not an omission. Empty
   `[requires]` table and absent `formula_compiler` key are the only
   forms that mean default capability." Add a row to the validation
   matrix for the empty-value case.
5. **Add a positive-enforcement mechanism for the caller-path
   guarantee.** Options: (a) make `CompileWithResult` the only
   constructor that returns a value any bead creator can consume, behind
   an unexported type that `Store.Create` callers cannot mint; (b) add a
   `go:generate` registry of every formula caller that the static guard
   cross-checks; (c) require an integration test that walks every
   `Store.Create`-from-formula site through the normalized pipeline.
   Pick one and name it.
6. **Define construct-presence precisely** for each v2-only field. For
   each row in the registry, state whether the trigger is "field key
   present in TOML" or "field has a non-default value." Add the
   corresponding rejected/accepted cases to the test matrix.
7. **Add a diagnostic for "child lowers parent requirement,"** or fold
   it explicitly into `formula.compiler_requirement_conflict` and add a
   matrix row that exercises it.
8. **Insert `version` into the deterministic diagnostic order.** Decide
   whether `formula.version_deprecated` fires before or after expression
   validation, and whether it suppresses on parse-failed formulas.
9. **State whether validation is fail-fast or aggregated.** Pick one and
   document the implication for `OnceKey` suppression and CLI exit
   behavior.
10. **Add a CI invariant** that asserts every entry in the v2-only
    construct registry has at least one positive test (construct
    present, no requirement → diagnostic) and one negative test
    (construct absent, v1 requirement → no diagnostic) for each
    location in `Locations scanned`.

**Questions:**

- The duplicated "Graph workflow metadata keys" row — what is its
  intended semantics? A reviewer cannot fix this without input from the
  author.
- Is `requires = { formula_compiler = ">=2" }` (TOML inline-table
  syntax) accepted alongside the `[requires]` table header? If yes,
  table-driven tests should cover both forms. If no, the document
  should say so.
- For `gc.kind` values not currently in the v2-only enumeration (e.g.,
  a future `gc.kind = "shadow-run"`), does the parser pass them through
  in a v1 formula, or does it require explicit registration? If passed
  through, "unknown gc.kind in v1" cannot trigger
  `formula.compiler_requirement_missing` — which seems correct but
  should be stated.
- Does `formula.compiler_requirement_unsupported` fire on inputs that
  also produce `formula.requirement_invalid_type`? A bare integer `2`
  fails the type check and would also fail an expression check; the
  diagnostic order says type wins. Please confirm in the test matrix.
- Is the v2-only construct scan applied recursively through `extends`
  chains? The design covers expansions and aspects but says nothing
  about parent formulas referenced via `extends`. If a child `extends` a
  parent that has `check`, does the child inherit the v2 requirement,
  or must the child re-declare it?
- Is `formula.compiler_requirement_missing` evaluated before or after
  expansion/aspect aggregation? The inheritance rules imply after, but
  the diagnostic ordering lists construct-presence as step 5 with no
  statement about scope-of-evaluation. A v2-only construct injected
  only by an aspect must surface a diagnostic that points at the aspect
  formula, not the root.
- Which TOML decoding strategy is canonical: a two-pass decode (raw
  `map[string]any` for shape and key checks, then typed decode for
  values), or `BurntSushi/toml`'s `MetaData` for source positions over a
  single typed decode? The choice determines what "preserve raw
  key/type/position" means in code.
- Who owns adding new entries to the `gc.kind` value enumeration —
  `internal/formula` or `internal/molecule`? The registry is described
  as a parser concern, but the values describe runtime kinds. Without
  naming the canonical owner, drift between the registry and runtime is
  the default outcome.
