# Formula Compiler Requirements v0

| Field | Value |
|---|---|
| Status | Proposed |
| Date | 2026-05-09 |
| Author(s) | Codex |
| Issue | [gastownhall/gascity#1760](https://github.com/gastownhall/gascity/issues/1760) |
| Supersedes | `contract = "graph.v2"` as the user-facing formula compiler requirement |

Design for replacing the formula `contract` field with an explicit
requirements table that declares the minimum formula compiler capability a
formula needs.

## Problem

Formula v2 currently uses:

```toml
formula = "code-review-loop"
version = 2
contract = "graph.v2"
```

This mixes three separate ideas:

- formula artifact revision
- formula file/compiler compatibility
- the current graph compiler implementation name

After review, formula artifact revision should not live on individual formula
files. The existing integer `version` field should not be repurposed into
formula semver. Formulas are distributed through packs, and pack version, pack
ref, or pack commit SHA is the durable way to identify the exact authored
workflow a consumer is using.

That leaves the current `contract` field. Its real purpose is not to describe a
runtime contract and not to select a compiler at execution time. It is a
minimum requirement: this formula uses constructs that require a compiler with
formula-v2 capability.

## Goals

1. Give formulas a standard way to declare minimum compiler requirements.
2. Make clear that formulas do not select the compiler implementation.
3. Remove formula-level artifact versioning from the design.
4. Preserve pack-level version/ref/SHA as the consistency mechanism for formula
   consumers.
5. Keep the current `[daemon] formula_v2` feature flag as the operator opt-in
   for formula-v2 compilation.
6. Provide a migration path from `contract = "graph.v2"` without breaking
   existing formulas immediately.

## Non-Goals

- Versioning individual formulas independently from their containing pack.
- Adding runtime compiler selection.
- Adding multiple formula compiler implementations.
- Changing formula layer resolution or filename-based shadowing.
- Changing pack semver, pack imports, or remote pinning behavior.

## Decision

Add a top-level `[requires]` table to formula files:

```toml
formula = "code-review-loop"

[requires]
formula_compiler = ">=2"
```

`requires.formula_compiler` declares the minimum formula compiler capability
needed to parse and compile the formula. It is a requirement expression, not a
selector.

The active Gas City binary decides which compiler path to use. For v0, the
mapping is simple:

| Requirement | Meaning |
|---|---|
| omitted | Formula only requires the default compiler capability |
| `>=2` | Formula requires formula-v2 graph compiler capability |

If `[daemon] formula_v2 = false`, a formula requiring compiler capability `>=2`
must fail with an actionable error that tells the operator to enable
`[daemon] formula_v2` or use a formula that does not require compiler v2.

## Why `requires`

`requires` is the standard package/config term for constraints that must be
satisfied by the host toolchain. It communicates the important boundary:

- the formula declares what it needs
- the host decides whether it can satisfy that need
- the formula does not choose the compiler at execution time

This is more precise than `contract`, which sounds like an interface agreement,
and more precise than `schema`, which suggests the physical file format.

## Formula Artifact Versioning

Formula files should not carry their own semver field.

Consumers who need consistency should pin the containing pack:

- by semver when using a released pack version
- by tag or branch when appropriate
- by commit SHA when exact reproducibility matters

This keeps the existing pack model as the single distribution and revision
boundary. A formula's identity remains its resolved formula name and winning
file in the formula layer stack. Its authored revision is the pack revision
that provided that winning file.

## Proposed TOML Surface

### Default compiler capability

```toml
formula = "simple-review"

[[steps]]
id = "review"
title = "Review the change"
description = "..."
```

No `[requires]` table means the formula can compile with the default formula
compiler capability.

### Formula-v2 compiler capability

```toml
formula = "code-review-loop"

[requires]
formula_compiler = ">=2"

[[steps]]
id = "review"
title = "Review the change"
description = "..."

[steps.retry]
max_attempts = 3
on_exhausted = "soft_fail"
```

The exact v2-only fields are validated by the formula package. If a formula uses
v2-only constructs without declaring `requires.formula_compiler = ">=2"`,
validation should fail with a direct message that names the missing requirement.

## Requirement Expression

V0 supports only the expression forms needed for current formula compiler
capabilities. The parser must reject every other value instead of accepting a
future range optimistically.

| Expression | Meaning |
|---|---|
| omitted | Default compiler capability, equivalent to `>=1` |
| `>=1` | Default compiler capability |
| `>=2` | Formula-v2 graph compiler capability |

Resolved public syntax choices:

- Exact `formula_compiler = "2"` is invalid syntax in v0.
- `">= 2"`, `" >=2"`, `">=2 "`, `">=2.0"`, and `">=2.1"` are invalid syntax
  in v0.
- `">=3"` is syntactically recognizable as a future monotonic minimum, but it
  is unsupported until the binary implements compiler capability 3.
- `requires.formula_compiler` must be a TOML string. Integer, float, boolean,
  array, table, and dotted-table values fail validation.
- An omitted `[requires]` table and an empty `[requires]` table both mean the
  default compiler capability.
- Unknown keys under `[requires]` fail validation. `formula_compiler` is the
  only supported requirement axis in v0.

Future compiler capability expressions can be added only by extending this
grammar and its validation matrix in the same change.

The parser must preserve raw TOML key names, types, and source positions until
diagnostics are built. Decoding directly into the final typed `Formula` shape is
not sufficient for `[requires]` because it can erase whether an author supplied
an integer, a dotted table, an unknown key, or an empty table.

The parser therefore has two stages:

```go
type RawRequirementField struct {
    TablePath  string
    Key        string
    ValueKind  string
    ValueText  string
    SourcePath string
    Line       int
    Column     int
}
```

Stage one captures raw requirement and legacy `contract` fields before typed
formula decoding. Stage two converts only the accepted raw fields into
`NormalizedRequirements`. Diagnostics always point at the raw field, never at a
lossy decoded zero value.

## Parser And Validation Contract

<!-- REVIEW: added per parser-validation-semantics -->

Requirement validation is owned by `internal/formula` and runs before any
caller creates beads, wisps, or workflow roots. The deterministic diagnostic
order is:

1. TOML decoding errors.
2. Unknown `[requires]` keys and invalid TOML types.
3. Malformed `requires.formula_compiler` syntax.
4. Syntactically valid but unsupported future `requires.formula_compiler`
   expressions.
5. Legacy `contract` normalization and `contract`/`requires` conflicts.
6. Missing compiler requirement for v2-only constructs.
7. Host capability failures such as `[daemon] formula_v2 = false`.

Accepted TOML shape matrix:

| TOML shape | Normalized result | Diagnostic |
|---|---|---|
| no `[requires]` table | default capability, source `omitted` | none |
| empty `[requires]` table | default capability, source `omitted` | none |
| `[requires] formula_compiler = ">=1"` | default capability, source `requires` | none |
| `[requires] formula_compiler = ">=2"` | compiler capability 2, source `requires` | none when host satisfies it |
| `[requires] formula_compiler = ""` | none | `formula.compiler_requirement_invalid_syntax` |
| `[requires] formula_compiler = "2"` | none | `formula.compiler_requirement_invalid_syntax` |
| `[requires] formula_compiler = ">= 2"` | none | `formula.compiler_requirement_invalid_syntax` |
| `[requires] formula_compiler = ">=2 "` or `" >=2"` | none | `formula.compiler_requirement_invalid_syntax` |
| `[requires] formula_compiler = ">=2.0"` or `">=2.1"` | none | `formula.compiler_requirement_invalid_syntax` |
| `[requires] formula_compiler = ">=3"` | none | `formula.compiler_requirement_unsupported_future` |
| `[requires] formula_compiler = 2`, `2.0`, `true`, `[]`, or `{}` | none | `formula.requirement_invalid_type` |
| `[requires.formula_compiler]` dotted or nested table | none | `formula.requirement_invalid_type` |
| `[requires] unknown_axis = "x"` | none | `formula.requirement_unknown_axis` |

V2-only construct registry:

| Construct | TOML locations scanned | Trigger predicate |
|---|---|
| Step `check` | `[[steps]]`, `children`, `loop.body`, expansion `template` | field is present, even if disabled by conditions |
| Legacy internal `ralph` | `[[steps]]`, `children`, `loop.body`, expansion `template` | field is present |
| Step `retry` | `[[steps]]`, `children`, `loop.body`, expansion `template` | retry table or retry metadata is present |
| Step `on_complete` | `[[steps]]`, `children`, `loop.body`, expansion `template` | on-complete formula/action is present |
| Workflow-control metadata key | step metadata, child metadata, loop body metadata, expansion template metadata | key is one of `gc.scope_name`, `gc.scope_role`, `gc.scope_ref`, `gc.continuation_group`, or `gc.on_fail` |
| Workflow-control metadata value | same metadata locations | `gc.kind` value is one of `scope`, `cleanup`, `scope-check`, `workflow-finalize`, `retry`, `retry-run`, `retry-eval`, `ralph`, `run`, or `check` |
| Step `expand` | root steps, children, loop bodies, inline expansion templates | expansion field is present and materializes a subgraph contribution |
| `expand_vars` | same locations as step `expand`, plus compose expansion/map vars | compile-time expansion vars are present and affect graph materialization |
| Expansion/aspect contribution | `compose.expand`, `compose.map`, `compose.aspects`, transitive imports | contributed formula or template contains any construct in this registry |

`needs`, `depends_on`, `children`, `loop`, `gate`, `condition`,
`compose.expand`, `compose.map`, and `compose.aspects` are not v2-only by
themselves in this design. They require compiler v2 only when the resolved
formula, expansion, aspect, or nested body also contains one of the v2-only
constructs above.

Normalization rules:

- Requirements are normalized after formula layer resolution and inheritance,
  before control-flow, expansion, retry, Ralph, and graph-control transforms.
- Parent requirements are inherited like the current `contract` field. A child
  may raise the requirement, but it may not lower a parent requirement.
- A child, expansion, aspect, or import that explicitly declares a lower
  requirement than the inherited or contributing context emits
  `formula.compiler_requirement_conflict` with both source locations. A default
  or empty `[requires]` in a contributed formula inherits when the formula has
  no independent v2-only constructs; it does not lower the compiled maximum.
- Inline children and `loop.body` inherit the containing formula's normalized
  requirement because they have no independent formula header.
- Expansion and aspect formulas are parsed and normalized independently. The
  compiled root requirement is the maximum requirement of the root formula and
  every expansion or aspect that contributed steps.
- If a formula, expansion, or aspect contains a v2-only construct but does not
  declare `requires.formula_compiler = ">=2"` or legacy `contract = "graph.v2"`,
  validation fails at that source formula with a diagnostic that names the
  missing requirement.

The validation matrix is generated from
`internal/formula/testdata/compiler_requirements_matrix.yaml`; the table below
is a rendered excerpt, not a hand-maintained checklist. The fixture file is the
normative source and every row names:

- requirement source (`omitted`, `requires`, `contract`, `dual`)
- raw TOML shape and raw value kind
- legacy `contract` state
- legacy `version` state (`omitted`, `1`, `2`, or invalid type)
- v2 construct registry entry and `min_compiler_capability`
- construct location, contribution path, and caller path
- host compiler capability
- normalized requirement and source attribution
- ordered diagnostics with source path, key, value, and severity

<!-- REVIEW: added per parser-validation-contract-v21 -->

The fixture is an executable contract, not a sample list. Its checked-in schema
is:

```yaml
schema_version: 1
dimensions:
  source_format: [toml, json]
  requirement_shape:
    [omitted, empty_table, string, integer, float, boolean, array, inline_table, dotted_table, duplicate_table, unknown_axis]
  requirement_value:
    [omitted, ">=1", ">=2", "", "2", ">= 2", " >=2", ">=2 ", ">=2.0", ">=2.1", ">=3"]
  contract_value: [omitted, graph.v2, graph.v1, other]
  version_value: [omitted, 1, 2, string, invalid_type]
  host_capability: [1, 2, invalid]
  construct_identity:
    [none, check, ralph, retry, on_complete, workflow_metadata_key, workflow_metadata_value, step_expand, expand_vars]
  construct_location:
    [root_step, child_step, loop_body_step, inline_expansion_in_step, inline_expansion_in_child, inline_expansion_in_loop_body, compose_expand_template, compose_map_template, aspect_advice]
  contribution_path:
    [root_file, child_body, loop_body, inline_expansion, compose_expand, compose_map, aspect, transitive_import]
  caller_path:
    [validate_only, show_preview, cook_root, cook_attach, sling_cli, sling_api, order_dispatch, controller_validation, fanout_fragment, convergence_create, convergence_retry, convergence_next_iteration, convoy_source_scan, dashboard_preview]
rows:
  - id: requires-v2-host-disabled
    input:
      source_format: toml
      source_path: formulas/root.toml
      requirement_shape: string
      requirement_value: ">=2"
      contract_value: omitted
      version_value: omitted
      host_capability: 1
      construct_identity: retry
      construct_location: root_step
      contribution_path: root_file
      caller_path: cook_root
    expect:
      normalized:
        formula_compiler: 2
        source: requires
        source_path: formulas/root.toml
        source_key: requires.formula_compiler
      diagnostics:
        - order: 1
          code: formula.compiler_requirement_unsatisfied
          severity: error
          source_path: city.toml
          source_key: daemon.formula_v2
          source_value: "false"
      diagnostic_count: 1
```

<!-- REVIEW: added per DR33-validation-matrix-bounded -->

The generator does not expand the full cross-product. The `dimensions` block is
the vocabulary, while executable suites select bounded combinations with named
coverage intent. CI locks the generated case count for each suite; adding a
requirement axis, construct, raw shape, or caller path must either add cases to
the owning suite with an updated count or declare a checked `unsupported` or
`impossible` row. This keeps the matrix auditable instead of turning it into an
unreviewable product of every dimension.

Normative suites:

| Suite | Covers | Count lock |
|---|---|---|
| `grammar` | omitted, empty, accepted byte-exact strings, rejected strings, unsupported future `>=3`, and invalid host capability | exact generated rows plus golden diagnostic order/count |
| `raw-shape` | TOML duplicate table, duplicate scalar key, dotted table, nested table, inline table, array, bool, integer, float, parser-boundary failures, and JSON equivalents | exact generated rows with TOML source position or JSON pointer expectations |
| `legacy-alias` | legacy `contract`, dual declarations, conflicts, unsupported contract values, legacy `version`, and warning placement | exact generated rows and warning/fatal precedence |
| `construct-registry` | every v2-only construct identity and decoded-field classification | at least one positive and one missing-requirement row per registry entry |
| `contribution-traversal` | root, child, loop body, inline expansion, compose expansion/map, aspect, and transitive import propagation | pairwise rows across construct location and contribution path |
| `caller-preflight` | validate/show previews plus every durable writer path | one fatal row per durable writer proving zero writes |
| `projection-parity` | CLI, API, API-routed CLI, dashboard, and Event Bus projections | golden rows for diagnostic field preservation, ordering, and count |

The raw-shape suite is the only owner of parser-boundary behavior. Duplicate
TOML tables or scalar keys that the TOML library rejects remain plain parser
errors with no structured formula diagnostic. Valid TOML/JSON values with the
wrong decoded shape become structured formula diagnostics. JSON cases use JSON
pointers as source keys and must be paired with the nearest TOML equivalent.
A reviewer must be able to add a new requirement axis or v2-only construct and
see exactly which suite counts, diagnostic fixtures, docs rows, and tests must
change.

<!-- REVIEW: added per validation-matrix-independent-dimensions -->

The dimensions are intentionally independent. `construct_identity` says what
semantic feature requires compiler capability 2. `construct_location` says
where that feature appeared in the parsed formula tree. `contribution_path`
says how the construct reached the final compiled recipe. `caller_path` says
which production or preview path consumed the compile output. The generator
must not collapse those axes into a single "formula source" label because that
would hide bypasses such as a `step.expand` inside a loop body, `expand_vars`
on an inline expansion, an aspect that contributes retry metadata, or a
transitive import that raises the maximum requirement.

Unsupported and impossible rows are checked fixtures, not comments. Examples:

| Row kind | Example | Required assertion |
|---|---|---|
| `unsupported` | JSON formula contains dotted-table equivalent for `[requires.formula_compiler]` | emits the same diagnostic code as the TOML invalid-type row, using a JSON pointer source key |
| `unsupported` | `expand_vars` contains a non-string compile-time value | emits a formula-var diagnostic before durable writes |
| `impossible` | `construct_location=aspect_advice` with `contribution_path=root_file` | generator rejects the row as invalid fixture input |
| `impossible` | `caller_path=convoy_source_scan` with a source-only requirement parse error | convoy scan never compiles source formulas and must use metadata predicates only |

Generator self-tests fail closed. They must prove every decoded field in
`internal/formula.Formula`, `Step`, inline expansion, compose expansion/map,
aspect advice, and import contribution is either registered with
`min_compiler_capability`, explicitly v1-compatible, or rejected as unsupported.
They must also prove each caller path has at least one fatal-diagnostic fixture
whose assertion includes "zero durable writes" when that caller can create
beads, wisps, convoys, dependencies, hooks, retry metadata, convergence state,
or root metadata.

<!-- REVIEW: added per legacy-version-bypass -->

Legacy `version` is never a compiler requirement. The matrix must cross
`version` omitted, `version = 1`, and `version = 2` with every v2-only
construct and with omitted, empty-table, `>=1`, and `>=2` requirement states.
All rows where a v2-only construct appears with only omitted/default
requirements fail with `formula.compiler_requirement_missing`, regardless of
whether legacy `version` is absent, `1`, or `2`. `version = 2` is preserved
only as legacy metadata and may produce `formula.version_deprecated` on
validation/display surfaces; it does not imply graph capability.

Fixtures that currently expect `version = 1` plus a v2-only construct such as
`[steps.check]` to compile as a legacy molecule are invalid under this design.
They must be removed or rewritten to dual-declare `contract = "graph.v2"` and
`[requires] formula_compiler = ">=2"` during the alias window. CI fails if any
fixture, generated matrix row, or golden diagnostic preserves a path where
legacy `version` satisfies or suppresses `requires.formula_compiler`.

`go generate ./internal/formula` regenerates the rendered matrix and the golden
diagnostic fixtures. CI fails when the generated Markdown table, golden
diagnostics, and fixture rows disagree.

Rendered validation matrix excerpt:

| `contract` | legacy `version` | `[requires] formula_compiler` | V2-only construct | Host `formula_v2` | Result |
|---|---|---|---|---|---|
| omitted | omitted, `1`, or `2` | omitted or empty table | no | either | default capability, optional `formula.version_deprecated` on validation/display |
| omitted | omitted, `1`, or `2` | `>=1` | no | either | default capability, optional `formula.version_deprecated` on validation/display |
| omitted | omitted, `1`, or `2` | `>=2` | no or yes | true | graph capability, no error except optional `formula.version_deprecated` |
| omitted | omitted, `1`, or `2` | `>=2` | no or yes | false | `formula.compiler_requirement_unsatisfied` |
| omitted | omitted, `1`, or `2` | omitted, empty table, or `>=1` | yes | either | `formula.compiler_requirement_missing` |
| `graph.v2` | omitted, `1`, or `2` | omitted | no or yes | true | graph capability plus `formula.contract_deprecated`, optional `formula.version_deprecated` on validation/display |
| `graph.v2` | omitted, `1`, or `2` | `>=2` | no or yes | true | graph capability, source `dual`, plus `formula.contract_deprecated`, optional `formula.version_deprecated` |
| `graph.v2` | omitted, `1`, or `2` | `>=1` | no or yes | either | `formula.compiler_requirement_conflict` |
| other value | any | any | any | any | `formula.compiler_requirement_unsupported` on `contract` |
| omitted | any | invalid syntax such as `2`, `>= 2`, `>=2.0`, or empty string | any | any | `formula.compiler_requirement_invalid_syntax` |
| omitted | any | unsupported future minimum such as `>=3` | any | any | `formula.compiler_requirement_unsupported_future` |
| omitted | any | wrong TOML type, dotted table, or nested table | any | any | `formula.requirement_invalid_type` |
| omitted | any | unknown `[requires]` key | any | any | `formula.requirement_unknown_axis` |

If several defects are present, diagnostic ordering follows the list above. For
example, an unknown requirement axis is reported before a missing v2 compiler
requirement, and a conflict between `contract` and `[requires]` is reported
before host capability satisfaction.

Combined-defect precedence examples:

| Input defects | First diagnostic |
|---|---|
| malformed TOML plus unsupported requirement | TOML decoding error |
| unknown `[requires]` key plus v2-only construct without `>=2` | `formula.requirement_unknown_axis` |
| wrong TOML type plus legacy `contract = "graph.v2"` | `formula.requirement_invalid_type` |
| unsupported future `formula_compiler = ">=3"` plus disabled host | `formula.compiler_requirement_unsupported_future` |
| `contract = "graph.v2"` plus `formula_compiler = ">=1"` plus disabled host | `formula.compiler_requirement_conflict` |
| `version = 1` or `version = 2` plus v2-only construct without `>=2` | `formula.compiler_requirement_missing` |
| missing `>=2` for v2-only construct plus disabled host | `formula.compiler_requirement_missing` |

Diagnostic count and ordering rules:

- Invalid TOML/JSON syntax stops formula validation because raw source
  locations are unavailable.
- For a valid raw file, all independent requirement diagnostics are emitted,
  sorted by precedence group, then source path, line, column, source key, and
  diagnostic code.
- Host satisfaction runs only when normalization produced a usable requirement
  and no fatal requirement syntax, type, axis, or conflict diagnostic exists for
  that source formula.
- Missing-requirement diagnostics are emitted once per source formula that owns
  the undeclared v2-only construct, not once per call site that imported it.
- `formula.contract_deprecated` and `formula.version_deprecated` warnings are
  appended after fatal diagnostics but keep their own source order when there
  are no fatal diagnostics.
- The matrix row's `diagnostic_count` is authoritative; tests fail if code
  short-circuits, duplicates warnings, or adds downstream diagnostics not named
  by the row.

The implementation must add table-driven tests for accepted strings, rejected
strings, invalid TOML types, unknown `[requires]` keys, empty tables, legacy
contract compatibility, conflicts, missing requirements, inherited
requirements, expansion/aspect aggregation, loop bodies, children, and
unsupported future requirements.

Raw file decode failures are classified before formula validation:

| Raw input failure | Diagnostic class |
|---|---|
| invalid TOML syntax, duplicate table, or invalid array/table structure | plain TOML parse/decode error; no formula diagnostic code |
| valid TOML with wrong `[requires]` value type | structured `formula.requirement_invalid_type` |
| valid TOML with malformed byte-exact string such as `">= 2"` or `"2"` | structured `formula.compiler_requirement_invalid_syntax` |
| valid TOML with syntactically valid future minimum such as `">=3"` | structured `formula.compiler_requirement_unsupported_future` |
| valid TOML with unknown `[requires]` key | structured `formula.requirement_unknown_axis` |
| valid TOML plus conflicting legacy `contract` | structured `formula.compiler_requirement_conflict` |

The v0 accepted grammar is byte-exact: only omitted, `">=1"`, and `">=2"` are
accepted. The parser may recognize `>=<integer>` only to emit the
machine-distinguishable `formula.compiler_requirement_unsupported_future`
diagnostic for future minimums above the host binary's known maximum. No
trimming, semver parsing, numeric coercion, or optimistic future-range
acceptance is allowed in v0.

V2-only construct registry completeness is CI-enforced. The registry fixture
enumerates every syntax location the decoder can produce for steps, children,
loop bodies, expansion templates, aspects, and imported formulas. A
reflection-backed unit test fails when a new decoded formula field or
metadata-controlled workflow construct is added without one of these outcomes:
registered as v2-only, explicitly classified as v1-compatible, or rejected as
unsupported input. This prevents silent graph-only syntax from bypassing
`requires.formula_compiler`.

## Canonical Compile Result

<!-- REVIEW: added per canonical-requirement-ownership -->

`internal/formula` is the only package that may interpret raw `contract`,
`requires.formula_compiler`, `version`, or v2-only construct strings. All
behavioral callers must consume a normalized compile result.

The required internal shape is:

```go
type CompilerCapability int

const (
    CompilerCapabilityDefault CompilerCapability = 1
    CompilerCapabilityV2      CompilerCapability = 2
)

type RequirementSource string

const (
    RequirementSourceOmitted  RequirementSource = "omitted"
    RequirementSourceRequires RequirementSource = "requires"
    RequirementSourceContract RequirementSource = "contract"
    RequirementSourceDual     RequirementSource = "dual"
)

type DiagnosticSeverity string

const (
    DiagnosticSeverityError   DiagnosticSeverity = "error"
    DiagnosticSeverityWarning DiagnosticSeverity = "warning"
)

type NormalizedRequirements struct {
    formulaCompiler CompilerCapability
    source          RequirementSource
    sourceFormula   string
    sourcePath      string
    deprecated      bool
}

type Diagnostic struct {
    Code                  string
    Severity              DiagnosticSeverity
    Formula               string
    SourcePath            string
    SourceKey             string
    SourceValue           string
    NormalizedRequirement string
    HostCapability        string
    Message               string
    Remediation           string
    OnceKey               string
}

type HostCapabilities struct {
    FormulaCompiler CompilerCapability
    Source          string
}

type CompileOptions struct {
    Vars             map[string]string
    HostCapabilities HostCapabilities
    ValidateRuntimeVars bool
}

type CompileIdentity struct {
    HostCapabilities HostCapabilities
    SearchPathsHash   string
    OptionsHash       string
    VarsHash          string
}

type CompileResult struct {
    Recipe       *Recipe
    Requirements NormalizedRequirements
    GraphWorkflow bool
    Diagnostics  []Diagnostic
    Provenance   Provenance
}
```

<!-- REVIEW: added per host-capability-satisfaction -->

The public compile path may keep a convenience `Compile(...) (*Recipe, error)`
wrapper, but callers that branch on graph behavior or project diagnostics must
use `CompileResult`, and callers that create workflow roots or stamp durable
metadata must use an `AcceptedCompileArtifact`. `Recipe` may carry the
normalized fields for instantiation, but it must not force consumers to
re-parse raw TOML strings.

`CompileResult` is a preview and inspection object. It is not proof that the
compile is acceptable for a durable write because a caller can still ignore
fatal diagnostics, skip host-capability satisfaction, or reuse the result with
different vars/search paths. Durable producers must first exchange a
`CompileResult` for a compiler-minted accepted artifact, described below.

`HostCapabilities` is an explicit input, not package-global state. Existing
`SetFormulaV2Enabled`/`IsFormulaV2Enabled` shims may remain only as temporary
wrappers that build `CompileOptions` at the CLI/API edge. They may not be read
inside requirement normalization or satisfaction logic.

`HostCapabilities` has one authoritative capability field. The legacy
`[daemon] formula_v2` boolean is translated at the CLI/API/controller edge into
`CompilerCapabilityDefault` or `CompilerCapabilityV2`; it is not carried as a
second boolean inside the formula package because that would allow contradictory
host state. `Source` is diagnostic attribution only, for example
`city.toml:[daemon].formula_v2`.

Valid host capability values are exactly `CompilerCapabilityDefault` and
`CompilerCapabilityV2` in v0. `HostCapabilitiesFromFormulaV2` is the only
production constructor for the legacy boolean. Any other value is rejected by
`CheckRequirements` as an internal configuration error diagnostic and never
panics. Future values such as `CompilerCapability(3)` are unsupported until the
same change adds typed parser support, diagnostics, metadata, docs, and tests.
Satisfaction is computed for every call from
`CompileOptions.HostCapabilities`; it is not cached on formula identity,
process-global config, or pack resolution state.

`NormalizedRequirements` is constructible only inside `internal/formula`.
Callers outside the package use accessor methods such as
`FormulaCompiler()`, `Source()`, `SourcePath()`, and `Deprecated()`. Test-only
constructors live in `_test.go` files or behind a `testonly` helper. Production
callers may not synthesize normalized requirements from raw TOML, root
metadata, or command flags.

Required formula-domain APIs:

```go
func CompileWithResult(ctx context.Context, name string, searchPaths []string, opts CompileOptions) (*CompileResult, error)
func CheckRequirements(req NormalizedRequirements, host HostCapabilities) []Diagnostic
func AcceptCompileResult(result *CompileResult, identity CompileIdentity) (AcceptedCompileArtifact, []Diagnostic)
func ValidateAcceptedArtifact(artifact AcceptedCompileArtifact, intent CompileWriteIntent) []Diagnostic
func HostCapabilitiesFromFormulaV2(enabled bool, source string) HostCapabilities
func RootMetadataFacts(metadata map[string]string) (RootMetadataFacts, []Diagnostic)
```

`CheckRequirements` returns diagnostics and never creates beads, writes
metadata, or mutates global state. `AcceptCompileResult` is the only function
that can mint an `AcceptedCompileArtifact`; it calls or verifies requirement
satisfaction, rejects any fatal diagnostic, verifies the compile identity, and
copies provenance from the compiler-owned result. Every caller that can create
a root, wisp, attached molecule, expansion fragment, order wisp, or convergence
instance must call `CompileWithResult` and `AcceptCompileResult` before the
first durable write. Unit tests must prove two compiles in the same process can
evaluate the same formula against different host capabilities deterministically,
including concurrent compiles with host capability `1` and `2`.

`GraphWorkflow` is true only when the accepted normalized compiler capability
is `CompilerCapabilityV2`; it is false for default-capability formulas even if
arbitrary metadata contains legacy-looking strings. `Recipe.GraphWorkflow` is a
read-only projection of the compile result for execution code that already
receives a recipe. Mutable recipe metadata is a serialization boundary only and
must not be used as a caller-side compiler API.

Workflow-root query authority is split deliberately and has one persistence
owner:

| Owner | Authority |
|---|---|
| `internal/formula` | Defines normalized requirement semantics, graph-workflow facts, and exact metadata key/value meanings |
| `internal/sourceworkflow` | Sole owner of workflow-root bead-store query criteria and post-fetch predicates |
| CLI/API/order/convoy/dashboard callers | Call the shared predicate/query helper only; no direct metadata filters |

Required `internal/sourceworkflow` APIs:

```go
type WorkflowRootCapabilityKind string

const (
    WorkflowRootNotWorkflow       WorkflowRootCapabilityKind = "not_workflow"
    WorkflowRootDefaultCapability WorkflowRootCapabilityKind = "default_capability_workflow"
    WorkflowRootKnownGraph        WorkflowRootCapabilityKind = "known_graph_workflow"
    WorkflowRootUnknownFuture     WorkflowRootCapabilityKind = "unknown_future_capability_workflow"
)

type WorkflowRootFacts struct {
    Kind                WorkflowRootCapabilityKind
    FormulaName         string
    CompilerCapability  int
    Requirement         string
    RequirementSource   string
    ArtifactRef         string
    Diagnostics         []formula.Diagnostic
}

func WorkflowRootFacts(metadata map[string]string) WorkflowRootFacts
func IsWorkflowRootMetadata(metadata map[string]string) bool
func IsGraphWorkflowMetadata(metadata map[string]string) bool
func WorkflowRootQuery(kind WorkflowRootKind, source SourceWorkflowScope) beads.ListQuery
func AcceptedCompileArtifactRef(metadata map[string]string) (string, bool)
```

`internal/formula.RootMetadataFacts` may classify the formula-owned metadata
keys, but it does not query beads and it is not the workflow-root predicate.
The `internal/sourceworkflow` package owns the predicate and query helper for
all callers. New packages may not create a successor owner without first
deleting or moving the old one in the same change.

<!-- REVIEW: added per DR33-typed-root-facts -->

Boolean predicates are compatibility helpers over `WorkflowRootFacts`, not the
primary model. Callers that need behavior choose from the typed kind:
non-workflow, default-capability workflow, known graph workflow, or unknown
future capability workflow. Old binaries that see valid canonical workflow
metadata with an unsupported compiler capability classify it as
`unknown_future_capability_workflow`, keep it visible for observation and
operator cleanup, and refuse to perform graph-specific child, fanout,
continuation, retry, or new-compile behavior they do not understand.

The query helper must expose typed criteria for all persistence scans that
currently need workflow roots: any workflow root, graph workflow roots,
canonical-only roots, legacy-only roots during the alias window, and roots that
need migration warnings. Tests must prove SQL/API/order/convoy selectors match
the same rows as the in-memory predicate.

Host capability plumbing is explicit at every production entry point:

| Entry point | Capability source and lifetime |
|---|---|
| `gc sling`, `gc formula`, `gc order` | loaded once from the active city config for that command invocation |
| HTTP/API handlers | derived from the request's resolved city config and passed through Huma handlers to formula adapters |
| order dispatch loop | snapshotted at the start of each order scan tick; one tick uses one snapshot |
| dispatch fanout | inherited from the parent dispatch operation and re-used for every fragment compile in that fanout transaction |
| convergence create/retry/speculative wisp | snapshotted before convergence validation and used until either zero writes happen or the accepted compile artifact is persisted |
| reconciler/controller tick | snapshotted per tick; existing accepted roots use persisted metadata, new formulas use the tick snapshot |

Caller inventory and required replacement behavior:

| Surface | Current risk | Required behavior |
|---|---|---|
| `internal/formula` parser, validation, and graph transforms | Raw `Contract` and `declaresGraphV2Contract` are load-bearing | Replace with `NormalizedRequirements` and `GraphWorkflow` |
| `gc formula show` | Preview can compile through legacy wrappers and omit structured diagnostics | Use `CompileWithResult` for display, keep runtime-var validation display semantics, and print/provide diagnostics without writing |
| `gc formula cook` | Root cook can bypass host checks before bead creation | Require `AcceptedCompileArtifact` before `molecule.Cook` writes root or children |
| `gc formula cook --attach` | Attached molecule can create child beads from a raw `Recipe` | Require `AcceptedCompileArtifact` before `molecule.Attach` writes the sub-DAG, dependency, hook, or attachment metadata |
| `internal/molecule` cook/cook-on/attach and graph apply | Root metadata can be stamped from raw contract or preview result | Stamp from `AcceptedCompileArtifact` only |
| `cmd/gc/cmd_sling.go` and `internal/sling` routing helpers | Graph routing and workflow attachment can branch on `gc.formula_contract` | Use shared workflow-root predicates backed by normalized metadata |
| `cmd/gc/cmd_order.go` and `cmd/gc/order_dispatch.go` | Order wisps can emit divergent errors | Preflight and accept a compile artifact; emit the shared diagnostic event on failure |
| `cmd/gc` helpers that shell out to `bd` or legacy formula probes | Shell-out output can bypass typed compiler diagnostics and durable preflight | Treat shell-outs as validation-only probes; no shell-out result may authorize root, child, hook, convoy, retry, or fanout writes |
| `internal/dispatch/fanout.go` and fanout expansion fragments | Runtime fragment compilation can happen after a durable root exists | Compile each fragment through `CompileWithResult` before fanout writes any child, convoy, or continuation metadata |
| `cmd/gc/cmd_convoy_dispatch.go` and convoy cleanup | Graph-only roots can be discovered by legacy metadata only | Use shared workflow-root predicate that accepts new and legacy metadata during migration |
| `internal/api` workflow-root read models | Duplicate predicates can diverge from CLI/sourceworkflow scans | Use `internal/sourceworkflow.WorkflowRootQuery` and typed facts for every workflow-root API projection |
| `internal/api/handler_sling.go`, formula endpoints, order feeds, and convoy projections | HTTP status and dashboard-facing errors can diverge | Project `Diagnostic` without hand-written JSON or string parsing |
| `internal/convergence` create/retry/next iteration | Subset validation can drift from full compiler semantics or reuse stale artifacts | Accept a compiler artifact, project typed convergence fields, and reuse only persisted artifacts whose identity matches the existing root |
| Dashboard generated types | UI can infer graph state from legacy metadata | Use API-projected typed fields and diagnostics |

<!-- REVIEW: added per caller-migration-executable -->

Executable call-site migration:

| Current file/helper | Target API | Required tests |
|---|---|---|
| `internal/formula/compile.go`: `isGraphWorkflow`, `declaresGraphV2Contract`, package-global `formulaV2Enabled` | `NormalizeRequirements`, `CheckRequirements`, and `CompileWithResult` with explicit `HostCapabilities` | `TestCompileWithResultHostCapabilitiesArePerCall`, concurrent enabled/disabled host tests |
| `internal/formula/types.go`: `requiresExplicitGraphContract`, `metadataRequiresGraphContract` | V2-only construct registry that returns source formula/path/key diagnostics | Parser matrix tests for every registry entry and nested location |
| `cmd/gc/cmd_formula.go`: `newFormulaShowCmd` | `CompileWithResult` in preview mode; no accepted artifact because it never writes | golden show diagnostics with required vars preserved |
| `cmd/gc/cmd_formula.go`: `newFormulaCookCmd` root and `--attach` branches | `CompileWithResult` -> `AcceptCompileResult` -> `molecule.CookAccepted` or `molecule.AttachAccepted` | root and attach tests prove disabled host writes no root, child, dependency, hook, or metadata |
| `internal/molecule.Cook`, `CookOn`, `Attach`, graph-apply build path | Accept `AcceptedCompileArtifact`; stamp recipe/root metadata from artifact requirements and provenance | Cook/attach/graph-apply tests assert canonical keys plus legacy key during alias window and compile-artifact refs when metadata is too large |
| `internal/sling.isGraphSlingFormula`, `validateSlingFormulaRuntimeVars`, `AttachFormula`, `LaunchFormula` | One preflight returning `AcceptedCompileArtifact`; use `GraphWorkflow` and shared diagnostics for conflict and runtime-var paths | CLI and `internal/sling` tests for enabled, disabled, force replacement, and no partial root on unsatisfied host |
| `cmd/gc/cmd_sling.go` graph decoration | Use accepted artifact provenance and canonical workflow-root metadata before route decoration | Existing graph sling tests updated to assert canonical metadata and no raw-contract branch |
| `cmd/gc/cmd_order.go` and `cmd/gc/order_dispatch.go` formula order cook | Preflight and accept artifact; emit one registered failure event on fatal diagnostics; continue later orders | Order tests for unsatisfied host, deprecation warning suppression, and successful later order |
| `cmd/gc` or internal helpers invoking `bd` for formula behavior | Remove from runtime authorization path; keep only explicit release-validation probes with typed native parity | static guard fails on durable writer paths that consume shell-out output; probe tests are non-interactive and validation-only |
| `internal/dispatch/fanout.go`: `CompileExpansionFragment` and any continuation fanout helper | Accept a parent artifact and compile/accept fragment formulas through the same host snapshot before durable fanout writes | Fanout tests prove disabled host capability creates zero child beads, convoys, continuation beads, or metadata |
| `cmd/gc/cmd_convoy_dispatch.go` source-workflow scans | Use `internal/sourceworkflow.IsWorkflowRoot` backed by canonical metadata first, legacy fallback second | Convoy tests with canonical-only, legacy-only, dual-stamped, and mixed-store roots |
| `internal/sourceworkflow.IsWorkflowRoot` and `ListLiveRoots` | Keep `internal/sourceworkflow` as the sole persistence predicate that calls formula metadata helpers | Predicate parity tests shared by sling, convoy, API, and dispatch |
| `internal/graphroute.IsCompiledGraphWorkflow` | Read `Recipe.GraphWorkflow` or `CompileResult.GraphWorkflow`, not root metadata strings | Graph route tests prove metadata changes do not affect compiled result semantics |
| `internal/api/handler_sling.go` and `handler_formulas.go` | Project `Diagnostic` into Huma response structs; never parse stderr/error strings | API tests for HTTP 400 diagnostic payloads and generated OpenAPI/type updates |
| `internal/api/orders_feed.go`, `handler_convoy_dispatch.go`, and workflow-root read-model helpers | Use shared workflow-root predicate for closed and open roots | API projection tests for canonical-only roots and legacy fallback; static guard against direct metadata filters |
| `internal/convergence/create`, retry, next-iteration, speculative wisp adapters | Preflight and accept artifact before any convergence bead/wisp write; keep convergence-only validation as post-compile domain checks | Convergence tests for disabled host capability with zero durable writes and artifact reuse after host downgrade |
| Dashboard generated types and state | Consume typed API diagnostics and workflow requirement fields | Dashboard tests for diagnostic rendering without metadata/string inference |

The migration is not complete while any non-test production code outside
`internal/formula` or the shared persistence predicate branches on
`Contract`, `declaresGraphV2Contract`, `Requires.FormulaCompiler`, or
`gc.formula_contract`.

<!-- REVIEW: added per caller-manifest-grep-derived -->

The first implementation PR must commit a grep-derived manifest at
`engdocs/release/formula-compiler-caller-manifest.md` or a machine-readable
equivalent under `docs/release/`. It is seeded from the current tree with
targeted searches for `CompileWithoutRuntimeVarValidation`,
`CompileExpansionFragment`, `SetFormulaV2Enabled`, `IsFormulaV2Enabled`,
`gc.formula_contract`, `contract = "graph.v2"`, `ValidateForConvergence`,
`molecule.Cook`, `molecule.Attach`, `GraphWorkflow`, `gc.fanout_state`,
`exec.Command("bd"`, and workflow-root metadata filters. Each row names the
owner package, current helper, durable-write
ability, migration target, and blocking test. At minimum it contains these
rows:

| Row | Current location | Durable write? | Migration owner/test |
|---|---|---|---|
| formula show preview | `cmd/gc/cmd_formula.go:newFormulaShowCmd` | no | CLI golden diagnostics and no-store-open test |
| formula cook root | `cmd/gc/cmd_formula.go:newFormulaCookCmd`, `internal/molecule.Cook` | yes | accepted-artifact root cook zero-write test |
| formula cook attach | `cmd/gc/cmd_formula.go:newFormulaCookCmd --attach`, `internal/molecule.Attach` | yes | accepted-artifact attach zero-write dependency/hook test |
| sling launch | `internal/sling.InstantiateSlingFormula`, `cmd/gc/cmd_sling.go` | yes | no partial root/child/hook on unsatisfied host |
| graph attachment predicate | `internal/sling.IsGraphWorkflowAttachment` | reads durable metadata | shared `internal/sourceworkflow` predicate parity test |
| order cook | `cmd/gc/cmd_order.go`, `cmd/gc/order_dispatch.go` | yes | grouped event and successful-later-order tests |
| durable shell-out bridge | any `bd` or legacy formula probe invocation reachable from launch/order/fanout/convergence | yes if it authorizes writes | validation-only probe or deletion plus static guard |
| fanout fragment | `internal/formula.CompileExpansionFragment`, `internal/dispatch/fanout.go` | yes | compile all fragments before first fragment write |
| convergence subset | `internal/convergence/formula.go:ValidateForConvergence` and create/retry callers | yes | projection and artifact-reuse tests |
| API sling/formulas | `internal/api/handler_sling.go`, `handler_formulas.go` | launch yes, preview no | typed Huma diagnostic and OpenAPI in-sync tests |
| API workflow roots | `internal/api/orders_feed.go`, `handler_convoy_dispatch.go`, workflow-root read models | reads durable metadata | shared predicate static guard and projection fixtures |
| convoy/source workflow | `cmd/gc/cmd_convoy_dispatch.go` | reads durable metadata | canonical-only/legacy-only/dual-stamped selector test |
| dashboard state | generated TS and dashboard panels | no direct write | generated-client compile and no-metadata-inference tests |

The manifest is updated in every caller-migration PR. A row can be removed only
when the static guard proves no production code path remains for that raw
consumer and the removal commit names the replacement test.

CI must include a static guard that fails on new behavioral uses of raw
`contract = "graph.v2"`, `declaresGraphV2Contract`, `Requires.FormulaCompiler`,
or `gc.formula_contract` outside the parser, compatibility metadata writer,
shared workflow-root predicate, tests, fixtures, and docs.

<!-- REVIEW: added per compiler-authority-enforcement -->

The guard starts from a checked-in allowlist, not from an aspirational grep
comment. The initial migration commit creates
`internal/formula/testdata/raw_consumer_allowlist.yaml` with this shape:

```yaml
schema_version: 1
allowed:
  - package: internal/formula
    file: internal/formula/compile.go
    symbols: [NormalizeRequirements, writeCompatibilityMetadata]
    reason: parser and alias writer own raw compatibility fields
  - package: internal/sourceworkflow
    file: internal/sourceworkflow/*.go
    symbols: [IsWorkflowRootMetadata, WorkflowRootQuery]
    reason: shared predicate owns legacy metadata fallback during alias window
```

`go test ./internal/formula ./internal/sourceworkflow ./cmd/gc` runs
`TestNoNewFormulaRawConsumers`, which scans production Go files for raw
`Contract`, `declaresGraphV2Contract`, `Requires.FormulaCompiler`,
`Version`, `gc.formula_contract`, `gc.formula_compiler_requirement`, and
`gc.formula_compiler_capability` reads. Additions fail unless the same change
updates this allowlist with a file, symbol, owner, expiry condition, and test
that proves the read is parser, compatibility-writer, or shared-predicate
logic. The allowlist is report-only only in rollout phase 1; it becomes
blocking before any durable producer switches to accepted compile artifacts.

Durable writer APIs must make the accepted compile identity impossible to
skip. New or modified APIs that can write roots, children, wisps, dependencies,
convoys, hooks, retry metadata, continuation metadata, or convergence state may
not accept a bare `*formula.Recipe`, `*formula.CompileResult`, or caller-built
`NormalizedRequirements` as proof of acceptance. They must accept
`formula.AcceptedCompileArtifact` or an entry-point-specific struct that embeds
that artifact and records the operation-specific write intent. Tests must prove
a caller cannot construct durable metadata from raw `contract`, raw
`[requires]`, legacy `version`, `gc.formula_*` metadata, global formula flags,
or a preview-only `CompileResult` without going through the compiler-owned
acceptance function.

Durable preflight contract:

| Entry point | Must complete before first durable write | Writes forbidden on fatal diagnostic |
|---|---|---|
| Root molecule or wisp launch | `CompileWithResult` -> `AcceptCompileResult` -> `ValidateAcceptedArtifact` | root bead, child beads, root metadata, convoy, hook |
| Attached molecule | compile/accept attached formula with current host capability and validate artifact against attach intent | attached child bead, dependency, hook, attachment metadata |
| Formula-backed order | compile/accept selected formula before marking the order fired | wisp root, child beads, order fired metadata |
| Retry or `on_complete` that starts a formula | compile/accept target formula before retry/on-complete state mutation | retry-run bead, attached molecule, continuation metadata |
| Fanout fragment | compile/accept every fragment formula in the fanout transaction and validate every artifact before fragment expansion | fragment child beads, convoy links, `gc.fanout_state`, continuation group metadata |
| Convergence create/retry/speculative wisp | compile/accept before convergence-specific validation writes | convergence root, iteration bead, missing-child state, retry metadata |

No production path may accept a pre-parsed `NormalizedRequirements` argument
from a caller as proof of satisfaction. Passing a full `CompileResult` is
allowed only between formula package helpers and `AcceptCompileResult`; durable
writers must receive the accepted artifact.

Fanout preflight is transactional. A fanout operation first compiles and
accepts the parent and every selected fragment using the same host-capability
snapshot, search-path hash, and vars/options identity. Only after every
fragment has an accepted artifact may the dispatcher mutate `gc.fanout_state`,
create fragment children, link convoys, or write continuation metadata. If any
fragment fails, the whole fanout returns the shared diagnostic and leaves the
store unchanged.

<!-- REVIEW: added per durable-artifact-identity-v21 -->

When a durable producer accepts a compiler-minted artifact before writing
beads, the accepted artifact identity is part of the contract:

```go
type AcceptedCompileArtifact struct {
    artifactVersion   int
    compileID         string
    formulaName       string
    hostCapabilities  HostCapabilities
    searchPathsHash   string
    optionsHash       string
    varsHash          string
    contentHash       string
    provenance        Provenance
    requirements      NormalizedRequirements
    diagnostics       []Diagnostic
    createdBy         string
    createdAt         time.Time
    proof             acceptedCompileProof
}
```

<!-- REVIEW: added per DR33-accepted-artifact-proof -->

The unexported `proof` field is intentional and is not a marker-only field:

```go
type acceptedCompileProof struct {
    nonce        [16]byte
    identityHash string
    accepted     bool
}
```

Code outside `internal/formula` cannot construct a non-zero accepted proof, and
the zero value of `AcceptedCompileArtifact` is invalid. Durable writer entry
points must call the formula-owned validation method before writing:

```go
type CompileWriteIntent struct {
    HostCapabilities HostCapabilities
    SearchPathsHash   string
    OptionsHash       string
    VarsHash          string
    ContentHash       string
    DurableProducer   string
}

func ValidateAcceptedArtifact(artifact AcceptedCompileArtifact, intent CompileWriteIntent) []Diagnostic
```

`ValidateAcceptedArtifact` fails closed when `proof.accepted` is false,
the compile ID is empty, the proof identity hash does not match the stored
identity fields, any fatal diagnostic is present, or the write intent disagrees
with host capability, search paths, compile options, vars, content hash, or
provenance. Callers may store and reload accepted artifacts only through
formula-owned serialization that revalidates the proof identity against the
persisted payload. They may not reconstruct an artifact from root metadata or
a `CompileResult`.

The only exported readers on the artifact return copied values; no caller can
mutate requirements, diagnostics, provenance, or identity after acceptance.

`SearchPathsHash` covers the ordered formula search path after pack/import
resolution. `OptionsHash` covers compile options that can affect expansion,
runtime-var validation, graph transforms, and diagnostic projection. `VarsHash`
covers only values used while compiling; secrets are hashed or omitted by a
typed redaction policy. A caller may submit a `CompileResult` for acceptance
only when these identities match the operation it is about to perform.
Otherwise it must compile again with current options before writing.

Required negative tests:

| Case | Assertion |
|---|---|
| zero-value artifact passed to root cook, attach, order, fanout, or convergence writer | validation diagnostic and zero durable writes |
| caller-built struct or stale deserialized payload with no accepted proof | validation diagnostic and zero durable writes |
| accepted artifact with stale host capability, search paths, options, vars, or content hash | validation diagnostic and zero durable writes |
| accepted artifact whose provenance/content hash no longer matches the selected formula | validation diagnostic and zero durable writes |
| accepted artifact containing any fatal diagnostic | validation diagnostic and zero durable writes |
| host-disabled compile result submitted for acceptance | no artifact minted; durable writer cannot be called without failing validation |

Retention semantics:

- Root, convergence, order-wisp, fanout, retry, and `on_complete` producers
  persist an accepted artifact whenever the full provenance or diagnostic set
  cannot fit in root metadata.
- The artifact is immutable and retained at least as long as the root bead or
  workflow history that references it.
- Garbage collection may delete an artifact only after no live or archived
  root metadata references `gc.formula_compile_artifact`.
- Reusing an artifact after a host downgrade is allowed only for already-created
  roots whose metadata references that artifact. Selecting or compiling a new
  formula after the downgrade must run `CompileWithResult` against the current
  host capability and can fail with zero writes.

Entry-point wrappers are thin and explicit:

```go
type RootLaunchCompile struct {
    Accepted formula.AcceptedCompileArtifact
    RootKind  string
}

type FanoutFragmentCompile struct {
    Parent   formula.AcceptedCompileArtifact
    Fragment formula.AcceptedCompileArtifact
    TargetID string
}
```

The wrapper may add write-specific context, but it must not weaken the accepted
artifact contract or allow callers to substitute raw metadata for compiler
acceptance.

## Compatibility With `contract`

<!-- REVIEW: added per migration-compatibility-contract -->

`contract = "graph.v2"` is a deprecated compatibility alias for:

```toml
[requires]
formula_compiler = ">=2"
```

V0 behavior:

- If only `contract = "graph.v2"` is present, compile as
  `requires.formula_compiler = ">=2"` and emit `formula.contract_deprecated`.
- If both are present and agree, compile and emit
  `formula.contract_deprecated`.
- If both are present and disagree, fail validation with
  `formula.compiler_requirement_conflict`.
- If `contract` has any value other than `graph.v2`, fail validation.
- The warning must preserve the source spelling and point to the exact
  replacement:

```text
contract = "graph.v2" is deprecated; use [requires] formula_compiler = ">=2"
```

Compatibility matrix:

| Formula source | Old binary that only knows `contract` | New binary with `formula_v2 = false` | New binary with `formula_v2 = true` |
|---|---|---|---|
| No requirement and no v2-only constructs | Works | Works | Works |
| `contract = "graph.v2"` | Works when old graph support is enabled | Fails with `formula.compiler_requirement_unsatisfied` | Works with deprecation warning |
| Dual `contract` plus `[requires] formula_compiler = ">=2"` | Works because old binary reads `contract` | Fails with `formula.compiler_requirement_unsatisfied` | Works with deprecation warning |
| `[requires] formula_compiler = ">=2"` only | Not supported for v2 formulas; old binary can miss the requirement | Fails with `formula.compiler_requirement_unsatisfied` | Works |
| Unsupported future requirement such as `>=3` | Not supported | Fails with `formula.compiler_requirement_unsupported_future` | Fails with `formula.compiler_requirement_unsupported_future` |

Runtime compatibility decision:

The live production runtime uses the native Go formula compiler only. Sling,
API sling, formula-backed orders, convergence, fanout, controller ticks, and
dashboard/API preview paths must not shell out to `bd` to decide formula
requirements or graph-workflow state. The historical `bd`/`GC_NATIVE_FORMULA`
path is a migration compatibility probe only: it may run in CI or explicit
operator validation, but no durable runtime entry point may depend on it after
this design lands. If the probe is still present, first-party graph formulas
remain dual-declared until the probe is removed or proven to parse `[requires]`
identically.

<!-- REVIEW: added per compatibility-corpus-v21 -->

This design supersedes conflicting rollback language in
`engdocs/proposals/formula-migration.md`, especially any text that treats
`GC_NATIVE_FORMULA=false` as a production runtime fallback after requirement
normalization ships. That proposal must be updated or marked superseded in the
same PR stack that introduces user-visible `[requires]` diagnostics. The only
permitted rollback after user-visible diagnostics ship is to restore
dual-declared formula sources and dual root metadata, or to disable the new
producer path before it writes durable state.

Per-entry-point behavior is therefore fixed:

| Entry point | Old or legacy-only `contract` formula | Dual-declared formula | Requires-only formula | Host `formula_v2=false` |
|---|---|---|---|---|
| `gc sling --formula` and API sling | native compiler accepts alias and warns | native compiler accepts and warns | native compiler accepts if host satisfies | returns typed diagnostic before root/child/hook writes |
| formula-backed orders | native compiler accepts alias and records bounded warning | same | same | records one grouped order compile-failure event for the scan tick; no wisp or fired metadata |
| convergence create/retry/speculative wisp | native compiler accepts alias before convergence validation | same | same | returns typed diagnostic before convergence root, retry metadata, missing-child state, or wisp writes |
| fanout fragments and `on_complete` formulas | native compiler accepts alias before fragment mutation | same | same | returns typed diagnostic before fragment children, convoy links, or continuation metadata |
| validation/report commands | reports alias, dual, or requires source | reports dual source | reports requires source | may still validate syntax and reports unsatisfied host separately |
| `bd` compatibility probe | must handle alias or dual through `contract` | must handle dual through `contract` | not a supported first-party distribution until minimum binary floor | not used for runtime writes |

First-party built-in and example graph formulas must stay dual-declared until
the minimum supported Gas City binary understands `[requires]` and either the
native compiler path is the only production path or the `bd` shell-out path is
proven to accept equivalent requirements. External SHA-pinned formulas that use
legacy `contract = "graph.v2"` remain valid through the alias window.

JSON formula policy:

TOML remains the canonical authoring surface. If an existing JSON formula
loader is enabled for legacy formulas, it must normalize through the same raw
requirement pipeline instead of bypassing it.

| JSON shape | Behavior |
|---|---|
| no `contract` and no `requires` | default compiler capability |
| `"contract": "graph.v2"` | deprecated alias for capability 2 with `formula.contract_deprecated` |
| `"requires": {"formula_compiler": ">=2"}` | JSON equivalent of the TOML `[requires]` table |
| both JSON `contract` and JSON `requires` agree | source `dual`, same warning as TOML |
| invalid JSON `contract` value | `formula.compiler_requirement_unsupported` with a JSON-pointer source key |
| malformed JSON `requires`, non-string value, unknown axis, or unsupported string | same structured diagnostic code as the equivalent TOML row |
| host disabled for any graph-capability JSON formula | `formula.compiler_requirement_unsatisfied` before durable writes |

First-party JSON formulas may not become requires-only until the minimum binary
floor artifact says all supported readers understand the JSON `requires`
object, or the JSON formula loader is formally retired.

The alias window is not time-based. The release captain for this migration may
remove legacy `contract` support only after all of these are true:

1. Every first-party pack, example, test fixture, and tutorial has shipped with
   `[requires]` for at least two minor releases.
2. `gc formula validate --all-packs --legacy-contract-report` reports zero
   first-party legacy-only formulas.
3. The release checklist records that externally pinned legacy formulas remain
   supported by either the alias or a documented compatibility branch.
4. CI's stale-guidance check rejects new docs that teach `contract` as the
   canonical surface.

<!-- REVIEW: added per migration-gates-measurable -->

Additional compatibility gates:

| Scenario | Required behavior before first-party formulas become requires-only |
|---|---|
| Old binary reading old pack | No change; legacy `contract` formulas keep working as today |
| Old binary reading dual-declared pack | Works because `contract` remains present and old binaries ignore `[requires]` |
| Old binary reading requires-only graph pack | Not a supported first-party distribution until the minimum binary floor is enforced |
| New binary reading old external SHA-pinned pack | Works through the `contract` alias and emits bounded deprecation diagnostics |
| New binary with `formula_v2=false` reading any graph-capability formula | Fails before durable writes with `formula.compiler_requirement_unsatisfied` |
| Mixed-version controllers sharing a bead store | New roots are dual-stamped until all supported readers use canonical keys |
| Native compiler path | Uses `CompileWithResult` and canonical diagnostics |
| `GC_NATIVE_FORMULA=false` or `bd` shell-out path | Optional validation-only probe; if marked active, preserve dual `contract` declarations or prove byte-level `[requires]` parity before source conversion |

<!-- REVIEW: added per alias-window-executable-gates -->

Alias-window gates are executable reports:

| Report or fixture | Must cover | Blocks |
|---|---|---|
| `--legacy-contract-report` | legacy-only, dual-declared, requires-only, empty `[requires]`, invalid `contract`, and source path/provenance | first-party requires-only conversion while first-party legacy-only rows remain |
| `--legacy-version-report` | `version` omitted, `1`, `2`, string, wrong type, and coexistence with v2-only constructs | diagnostics rollout when `version` still implies compiler behavior anywhere |
| inherited requirement conflict fixture | parent raises to `>=2`, child attempts default or `>=1`, expansion/aspect raises beyond root | caller migration when conflicts do not produce deterministic source-attributed diagnostics |
| first-party dual-declaration guard | built-in packs, examples, tutorials, and test fixtures during alias window | release when a first-party graph formula drops `contract` before the floor allows it |
| external pinned-pack inventory | local path, git SHA, tag, branch, registry source, transitive import, and shadowed formula | alias removal while any supported external path still depends on legacy `contract` |
| old-reader/probe corpus | latest supported old binaries, current binary, and optional active probe | first-party requires-only conversion while old readers are still in support |

`gc formula validate --all-packs --legacy-contract-report` is a release-gate
command, not best-effort text. It returns JSON by default when `--json` is set
with this shape:

```json
{
  "legacy_only": 0,
  "dual_declared": 0,
  "requires_only": 0,
  "unsupported": [],
  "items": []
}
```

Each item records formula name, source path, pack binding, import source, ref,
locked revision or local hash, requirement source, and whether the formula is
first-party or external. Exit code `0` means no first-party legacy-only formulas
were found; exit code `2` means first-party legacy-only formulas remain; exit
code `1` is reserved for I/O or malformed-formula failures. The minimum binary
floor is recorded in the release checklist and `docs/reference/config.md` before
any first-party requires-only conversion lands.

Old-reader and native-vs-`bd` parity are exercised by a pinned corpus under
`internal/formula/testdata/compat_corpus/`. The corpus contains at least these
named fixtures:

| Fixture | Required readers | Expected result |
|---|---|---|
| `legacy-only-graph-v2` | latest two published Gas City minor releases, current `main`, `bd >= 1.0.0` probe | old readers and new host-enabled readers accept; new host-disabled reader fails before writes |
| `dual-declared-graph-v2` | same | old readers accept through `contract`; new readers report source `dual` and deprecation warning |
| `requires-only-graph-v2` | current `main` only until minimum binary floor | old readers are allowed to fail; first-party distribution blocked until floor is enforced |
| `unsupported-future-requirement` | current `main` | deterministic `formula.compiler_requirement_unsupported_future` |
| `invalid-shape-non-string` | current `main` | deterministic `formula.requirement_invalid_type` |
| `diagnostic-source-attribution` | direct CLI, API-routed CLI, Huma endpoint, dashboard fixture | identical code, source path/key/value, remediation, and host capability |

The supported reader set for a release is materialized in
`docs/release/formula-compiler-compatibility.yaml`. It records exact binary
versions or SHAs for the latest two supported Gas City minor releases, current
`main`, and any optional compatibility probe. For this design baseline,
production is native-only; a `bd` or `GC_NATIVE_FORMULA=false` probe is a
release-validation tool only when a checked-in test proves the command still
exists and runs non-interactively. If that proof does not exist, the
compatibility artifact must mark the probe `status: not-needed` and no release
gate may require it.

<!-- REVIEW: added per executable-migration-gates -->

The three release artifacts are executable inputs with owners and schemas:

```yaml
# docs/release/formula-compiler-compatibility.yaml
schema_version: 1
owner: release-captain
updated_by: gc formula validate --compat-corpus internal/formula/testdata/compat_corpus --write-compatibility-artifact
supported_readers:
  - name: gas-city
    kind: binary
    version: 0.x.y
    sha: <git-sha>
    required: true
  - name: bd-probe
    kind: compatibility_probe
    status: not-needed # active only while a checked-in probe command exists
    version: 1.0.0
    required_until: native-only-runtime
fixtures:
  - id: dual-declared-graph-v2
    expected: accept
    diagnostics: [formula.contract_deprecated]
```

```json
{
  "schema_version": 1,
  "owner": "release-captain",
  "updated_by": "gc formula validate --all-packs --write-min-floor docs/release/formula-compiler-min-floor.json",
  "minimum_gc_for_requires_only": "0.x.y",
  "first_gc_with_requires": "0.x.y",
  "first_gc_with_canonical_root_metadata": "0.x.y",
  "first_gc_without_bd_runtime_fallback": "0.x.y",
  "dual_declared_compatibility_window_minor_releases": 2,
  "first_party_requires_only_allowed": false
}
```

`docs/release/formula-compiler-external-support.md` is a release checklist,
not prose-only background. It must contain frontmatter with
`schema_version`, `owner`, `status` (`active`, `expired`, or `not-needed`),
`supported_old_binaries`, `support_strategy`, `expires_after_release`, and
links to legacy inventory reports. CI blocks alias removal while the file is
missing, has `status: active`, or names supported readers that still require
legacy `contract`.

<!-- REVIEW: added per DR33-rollout-ownership -->

The initial implementation seeds these artifacts with conservative blocking
values before any first-party source conversion:

| Artifact | Initial owner | Conservative seed |
|---|---|---|
| `docs/release/formula-compiler-compatibility.yaml` | release-captain | current `main` required, older supported readers listed, optional `bd` probe `not-needed` unless a checked owner test proves it active |
| `docs/release/formula-compiler-min-floor.json` | release-captain | `first_party_requires_only_allowed: false` and no minimum floor until release evidence exists |
| `docs/release/formula-compiler-external-support.md` | release-captain | `status: active`, alias support as the default strategy, and explicit expiration unset |

"Two minor releases" means two completed Gas City minor releases after the
first release that shipped dual declarations, canonical metadata writers, and
the docs/example bundle. Pre-releases, nightly builds, failed releases, and
commits that were not published to the supported distribution channel do not
count. The release checklist records the two release tags, dates, supported
reader set, legacy inventory path, and compatibility-corpus result before a
gate depending on two minor releases may pass.

Legacy `version` handling has an explicit report owner. The same release
captain owns `gc formula validate --all-packs --legacy-version-report --json`,
which records `version` omitted, `1`, `2`, string, wrong type, coexistence
with v2-only constructs, and whether any code path still treats `version = 2`
as compiler behavior. User-visible diagnostics and first-party source
conversion are blocked while that report finds behavioral `version` coupling.

The `GC_NATIVE_FORMULA=false` and `bd` paths are validation-only probes after
this design lands, and they are optional. They are not supported runtime
fallbacks for sling, orders, convergence, fanout, controller validation, API,
or dashboard preview. Keeping either path as an active release gate requires a
checked-in owner test such as `TestLegacyFormulaProbeExists`, a documented
non-interactive command, and byte-level corpus parity: the native and probe
runs must match normalized requirement, source attribution, diagnostic code,
diagnostic order, exit code, and provenance for every fixture that first-party
distribution depends on. Without that proof, the only valid production
precondition is native Go compiler support.

```bash
gc formula validate --compat-corpus internal/formula/testdata/compat_corpus --json
# Optional only when docs/release/formula-compiler-compatibility.yaml marks the probe active:
GC_NATIVE_FORMULA=false gc formula validate --compat-corpus internal/formula/testdata/compat_corpus --json
```

The first command is authoritative for production behavior. The second command
is a release probe only when the artifact marks it active; if it is absent or
not-needed, requires-only conversion is governed by the minimum binary floor,
legacy inventory, and native corpus. If an active probe diverges on first-party
dual-declared formulas, requires-only conversion remains blocked. Pass/fail is
strict: accepted fixtures must match normalized requirement, source
attribution, diagnostics, exit code, and provenance; rejected fixtures must
match diagnostic code, count, ordering, and source attribution.

Measurable release gates:

| Gate | Artifact | Blocks |
|---|---|---|
| Legacy inventory | `gc formula validate --all-packs --legacy-contract-report --json` saved in release artifacts | first-party requires-only conversion while `legacy_only > 0` |
| Provenance inventory | `gc formula validate --all-packs --provenance --json` | source conversion when any first-party graph formula lacks pack binding, locked revision, or content hash |
| Minimum binary floor | `docs/release/formula-compiler-min-floor.json` plus release-note entry | publishing requires-only first-party packs |
| External pinned-pack plan | release checklist entry naming alias support, compatibility branch, or LTS binary | removing `contract` alias |
| Mixed-store compatibility | test report for old/new controllers reading dual-stamped and canonical-only roots | retiring dual root stamps |
| `bd`/native parity | golden corpus result for native and the optional active probe | source conversion only while the compatibility artifact marks a probe active |
| Stale guidance | CI report for docs/examples generated help | user-visible diagnostics |

The external support plan is a release-owned artifact at
`docs/release/formula-compiler-external-support.md`. It names the release
captain, supported old binary versions, whether external SHA-pinned packs are
served by alias support, a compatibility branch, or an LTS binary, opt-in
instructions for users, removal criteria, and the artifact location for each
legacy inventory report. Alias removal is blocked while this artifact is
missing or says external support is active.

`formula-compiler-min-floor.json` records the lowest Gas City version allowed
to consume first-party requires-only graph formulas, the first release that
understands `[requires]`, the first release that writes canonical root metadata,
and the release that stopped needing the `bd` compatibility path. Rollback is
to restore dual declarations and dual root stamps, not to reinterpret
requires-only formulas in old binaries.

First-party graph packs may become requires-only only when their pack metadata
also declares a compatible binary floor:

```toml
[pack]
requires_gc = ">=<minimum-floor-from-formula-compiler-min-floor.json>"
```

Pack resolver and import loading must reject any pack whose `[pack]
requires_gc` exceeds the active binary before formula selection, staging,
shadow resolution, or durable writes. First-party requires-only graph packs
must set the floor from `formula-compiler-min-floor.json`; external packs set
their own floor, but the resolver still rejects them fail-closed. Tests cover
built-in packs, direct local packs, installed packs, git/registry sources,
imported pinned packs, transitive imports, local dirty packs, and shadowed
formulas where a losing source would have required a higher floor.

Dual-stamp retirement criteria:

1. All supported readers use the shared canonical workflow-root predicate.
2. Mixed-version shared-store tests pass with canonical-only roots.
3. The external pinned-pack plan has shipped.
4. The previous two minor releases wrote dual stamps and emitted no known
   compatibility incident.

Alias removal criteria are stricter than dual-stamp retirement: the parser may
stop accepting legacy `contract` only after the enforced minimum binary floor
artifact says all supported readers understand canonical `[requires]`, the
configured dual-declared compatibility window has elapsed, the external plan
expires, and the legacy inventory gate has reported zero first-party
legacy-only formulas for two consecutive minor releases.

## Host Capability And Diagnostics

<!-- REVIEW: added per operator-diagnostics-projection -->

`[daemon] formula_v2` remains the host capability gate. A formula whose
normalized requirement is `CompilerCapabilityV2` must fail before creating
any new runtime state when the host capability is disabled.

At config load, `[daemon] formula_v2 = true` maps to
`HostCapabilities{FormulaCompiler: CompilerCapabilityV2}`; false maps to
`CompilerCapabilityDefault`. The boolean exists only as legacy config
vocabulary at the edge; `FormulaCompiler` is the canonical typed capability
used by `CheckRequirements`.

Required diagnostic codes:

| Code | Severity | Meaning |
|---|---|---|
| `formula.requirement_unknown_axis` | error | `[requires]` contains an unsupported key |
| `formula.requirement_invalid_type` | error | A requirement value has the wrong TOML type |
| `formula.compiler_requirement_invalid_syntax` | error | The expression is malformed or outside the byte-exact v0 grammar |
| `formula.compiler_requirement_unsupported_future` | error | The expression is syntactically valid but requires a future compiler capability |
| `formula.compiler_requirement_unsupported` | error | A legacy compatibility field such as `contract` has an unsupported value |
| `formula.compiler_requirement_conflict` | error | `contract` and `[requires]` disagree |
| `formula.compiler_requirement_missing` | error | A v2-only construct lacks a v2 requirement |
| `formula.compiler_requirement_unsatisfied` | error | Host config cannot satisfy the normalized requirement |
| `formula.host_capability_invalid` | error | Production code passed an invalid host capability value |
| `formula.contract_deprecated` | warning | Legacy `contract` spelling was accepted |
| `formula.version_deprecated` | warning | Legacy formula `version` was present and preserved only as metadata |

Diagnostic projection matrix:

| Code | Canonical message/remediation | CLI | API | Dashboard | Events |
|---|---|---|---|---|---|
| `formula.requirement_unknown_axis` | `unsupported [requires] key <key>` / remove it or upgrade to a binary that supports the axis | fatal stderr, exit 1 | HTTP 400 typed diagnostic | formula preview validation error | no event except controller/order failure wrapper |
| `formula.requirement_invalid_type` | `<key> must be a TOML string` / use `formula_compiler = ">=2"` | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_invalid_syntax` | `invalid formula compiler requirement <value>` / use omitted, `>=1`, or `>=2` exactly | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_unsupported_future` | `formula compiler capability <N> is not supported by this binary` / upgrade Gas City or use `>=1` or `>=2` | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_unsupported` | `unsupported legacy formula contract <value>` / use `[requires] formula_compiler = ">=2"` | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_conflict` | `contract and [requires] disagree` / make both declare graph v2 or remove `contract` | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_missing` | `v2-only construct requires formula_compiler = ">=2"` / add `[requires] formula_compiler = ">=2"` | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_unsatisfied` | `host formula compiler capability 1 does not satisfy >=2` / enable `[daemon] formula_v2` or choose a v1 formula | fatal stderr, exit 1; no beads created | HTTP 400 for validation/preview or 409 for launch conflict plus diagnostic body | disabled-capability state with remediation | registered compile-failure wrapper event |
| `formula.host_capability_invalid` | `invalid host formula compiler capability <value>` / construct host capabilities at the config edge | fatal stderr, exit 1 | HTTP 500 unless caused by request-local test input | internal error state with remediation | registered producer failure event |
| `formula.contract_deprecated` | `contract = "graph.v2" is deprecated` / use `[requires] formula_compiler = ">=2"` | bounded warning stderr by source/key | warning diagnostic in 200/preview response | non-blocking warning | none; warnings are never Event Bus payloads |
| `formula.version_deprecated` | `formula version is legacy metadata` / use pack version/ref/SHA for artifact identity | warning only on validate/show, never on launch success | warning diagnostic on formula endpoints | non-blocking warning | none |

Projection rules:

<!-- REVIEW: added per DR33-diagnostics-policy -->

- Warning diagnostics do not publish Event Bus events. They are returned on
  synchronous CLI/API/dashboard validation surfaces and may be counted in
  producer-local state, but `formula.contract_deprecated` and
  `formula.version_deprecated` never append warning events.
- Compile-failure events are fatal-diagnostic wrappers only. Their event
  constants are limited to named background or durable producers and must carry
  the same typed diagnostic payload as the synchronous surface.
- `config_generation` is the monotonically increasing generation assigned by
  the city config loader after a successfully parsed config snapshot is
  installed. A failed reload does not advance it. Direct CLI commands use the
  snapshot generation they loaded for that invocation; order, controller, and
  convergence loops use the generation captured at the start of the scan or
  operation. Suppression and grouping reset when the generation changes.
- CLI commands that compile formulas print warnings to stderr once per
  `(code, source path, source key)` per command invocation. Fatal diagnostics
  exit non-zero and include the code, source path, offending value, normalized
  requirement, host capability, and remediation.
- API endpoints return typed diagnostic fields in Huma-registered response
  bodies. User-correctable formula input errors use HTTP 400; internal I/O
  failures remain HTTP 500.
- Dashboard state is derived from the API diagnostic projection, not by parsing
  stderr or root metadata strings.
- Interactive preview/validate calls are synchronous by default. Direct CLI,
  API-routed CLI, and Huma preview/validate endpoints return typed diagnostics
  to the caller and do not publish Event Bus events unless they are invoked by
  a named durable/background producer listed below.
- Order dispatch emits at most one typed order failure event per
  `(order id, formula, OnceKey, host capability, config generation)` in a scan
  series and continues scanning later orders. Repeated scans increment a
  grouped occurrence counter on the producer state or next emitted event; they
  do not create wisps, mark the order fired, or append an unbounded event
  stream.
- Controller and convergence paths wrap the same `Diagnostic` code and fields
  in their existing error/event surfaces. They must not synthesize alternate
  wording that loses the remediation.
- Deprecation warnings are diagnostics attached to the compile result. CLI,
  API, dashboard, and controller projections decide whether and where to show
  them, but they all preserve the same code and source spelling.
- Event payloads for compile failures are typed and registered in
  `events.KnownEventTypes`. Payload fields include `code`, `severity`,
  `formula`, `source_path`, `source_key`, `source_value`,
  `normalized_requirement`, `host_capability`, `message`, `remediation`, and
  `once_key`. No event path may hand-write JSON or project diagnostics through
  `map[string]any`.
- `OnceKey` shape is
  `<code>|<source_path>|<source_key>|<source_value>|<host_capability>`. CLI
  suppression is per process invocation. Background suppression is an
  in-memory LRU per city and producer, capped at 4096 keys with oldest-entry
  eviction, reset on process restart and config reload. Warning diagnostics are
  suppressed by `OnceKey`; fatal diagnostics in background loops are grouped by
  `(producer, subject id, OnceKey, host capability, config generation)` until
  the formula source, config generation, or subject id changes. Evictions are
  observable through a producer-local counter and debug log field so operators
  can tell the difference between genuinely new failures and cache churn.

<!-- REVIEW: added per diagnostics-wire-contract -->

Typed HTTP responses use Huma structs generated from Go types. The formula
validation, preview, sling-launch conflict, order preview, and convergence
preview routes all embed the same diagnostic payload:

```go
type FormulaDiagnostic struct {
    Code                  string `json:"code" doc:"Stable diagnostic code"`
    Severity              string `json:"severity" enum:"error,warning"`
    Formula               string `json:"formula,omitempty"`
    SourcePath            string `json:"source_path,omitempty"`
    SourceKey             string `json:"source_key,omitempty"`
    SourceValue           string `json:"source_value,omitempty"`
    NormalizedRequirement string `json:"normalized_requirement,omitempty"`
    HostCapability        string `json:"host_capability,omitempty"`
    Message               string `json:"message"`
    Remediation           string `json:"remediation,omitempty"`
    OnceKey               string `json:"once_key,omitempty"`
}

type FormulaDiagnosticsBody struct {
    Diagnostics []FormulaDiagnostic `json:"diagnostics"`
}
```

No formula diagnostic HTTP path may use `map[string]any`, `json.RawMessage`,
hand-written JSON, or stderr parsing. `TestOpenAPISpecInSync`,
dashboard generated-client compilation, and a golden OpenAPI excerpt test must
fail when this shape changes without regenerated schemas and TypeScript.

CLI exit code mapping is fixed: `0` for accepted input with zero fatal
diagnostics, `1` for process/internal I/O failures, and `2` for formula
diagnostics on validation/report commands. Runtime launch commands such as
`gc sling --formula` and `gc order` use exit `1` for a failed operation but
still print the same diagnostic code and fields. API-routed CLI commands must
preserve the server diagnostic code, message, remediation, source attribution,
and status class instead of rewriting them locally.

Surface parity contract:

| Surface | Required projection | Parity tests |
|---|---|---|
| Direct CLI (`gc formula`, `gc sling`, `gc order`) | Prints the canonical diagnostic code, source path, source key/value, normalized requirement, host capability, and remediation | direct CLI golden stderr and exit-code tests |
| API-routed CLI | Forwards the typed Huma diagnostic without rewriting code, message, remediation, or source attribution | same formula fixture through direct and API-routed CLI produces byte-normalized diagnostics |
| Huma HTTP endpoints | Use typed response structs with `diagnostics []FormulaDiagnostic`; no `map[string]any`, `json.RawMessage`, or hand-written JSON | OpenAPI in-sync and generated-client compile tests |
| Generated TS client and dashboard | Render the generated diagnostic fields; never infer graph state or remediation from metadata strings | dashboard state tests using generated fixtures |
| Controller/order/convergence events | Wrap the same diagnostic in a registered typed payload and envelope source fields | event registry test plus repeated failure/dedup tests |

Registered event payload contract:

| Event constant | Emitted by | Payload |
|---|---|---|
| `formula.compile_failed` | named controller/background validation job only; not direct CLI/API preview | `FormulaDiagnosticPayload` plus producer id and occurrence count |
| `order.formula_compile_failed` | formula-backed order dispatch | `FormulaDiagnosticPayload` plus order id, scan generation, and occurrence count |
| `convergence.formula_compile_failed` | convergence create/retry/speculative-wisp paths | `FormulaDiagnosticPayload` plus convergence id when one already exists |
| existing sling/API failure wrapper | sling launch and API launch conflicts | `FormulaDiagnosticPayload` embedded in the typed error body/event payload; Event Bus emission only when a durable producer owns the launch |

Every event constant above must be present in `events.KnownEventTypes` and have
`events.RegisterPayload(constant, FormulaDiagnosticPayload{})` or a typed
wrapper payload registered before it can ship. Envelope fields carry the city,
agent/session when known, and source operation; payload fields carry the formula
diagnostic. The deduplication key is the diagnostic `OnceKey` plus producer,
subject id, host capability, and config generation for automatic loops.
Interactive operator attempts are never hidden from their caller, but they do
not publish repeated Event Bus entries.

Required repeated-loop tests:

| Scenario | Assertion |
|---|---|
| same disabled-capability order scanned three times with unchanged config | one grouped `order.formula_compile_failed`, occurrence count `3`, no wisp roots, child beads, or fired metadata |
| same controller validation failure across config reload | first generation grouped separately from second generation |
| direct `gc formula validate` repeated three times | three synchronous stderr/API results, zero Event Bus events |
| API preview called by dashboard polling | typed response every call, zero Event Bus events |

<!-- REVIEW: added per diagnostics-parity-fixtures -->

Diagnostic parity is fixture-driven. Each fixture under
`internal/formula/testdata/diagnostics/` contains the source formula, optional
city config, expected normalized requirement, host capability, ordered
diagnostics, and projections for every surface:

```yaml
schema_version: 1
id: disabled-v2-host
source: formulas/graph.toml
host_capability: 1
expect:
  diagnostics:
    - code: formula.compiler_requirement_unsatisfied
      severity: error
      source_path: city.toml
      source_key: daemon.formula_v2
      normalized_requirement: ">=2"
      host_capability: "1"
      remediation: enable [daemon] formula_v2 or choose a v1 formula
  cli_exit_validate: 2
  cli_exit_launch: 1
  api_status_preview: 400
  api_status_launch: 409
  event_constants: [order.formula_compile_failed]
  dashboard_group:
    key: formula.compiler_requirement_unsatisfied|formulas/graph.toml|daemon.formula_v2|1
    title: Formula compiler requirement unsatisfied
    occurrence_count: 1
```

The golden tests compare direct CLI, API-routed CLI, Huma response bodies,
generated TypeScript fixtures, dashboard rendering state, and event payloads.
Every projection must preserve `Code`, canonical message, remediation,
source path/key/value, normalized requirement, host capability, severity, and
`OnceKey`. Surfaces may add local context such as HTTP status or producer id,
but they may not rewrite or drop formula diagnostic fields.

Dashboard-visible grouped failure state is explicit. Background producer
failures appear as grouped rows keyed by diagnostic `OnceKey`, producer,
subject id, host capability, and config generation. A config-generation change
starts a new group; a process restart may clear in-memory suppression but must
not delete persisted failure history attached to the producer or root. Preview
and polling failures are visible to the caller but do not create Event Bus
entries or durable grouped failure rows.

## Persisted Metadata And Provenance

<!-- REVIEW: added per metadata-provenance-contract -->

Workflow roots created from compiler-v2 formulas must be dual-stamped during
the compatibility window:

| Metadata key | Value | Status |
|---|---|---|
| `gc.formula_compiler_requirement` | `>=2` | Canonical |
| `gc.formula_requirement_source` | `requires`, `contract`, or `dual` | Canonical |
| `gc.formula_compiler_capability` | `2` | Canonical |
| `gc.formula_contract` | `graph.v2` | Deprecated compatibility stamp |

Auditable provenance for every new workflow root:

| Metadata key | Value |
|---|---|
| `gc.formula_name` | resolved formula name |
| `gc.formula_source_path` | winning formula file path |
| `gc.formula_pack_name` | containing pack name when known |
| `gc.formula_pack_source` | local path, git URL, registry source, or `builtin` |
| `gc.formula_pack_ref` | requested tag, branch, semver, or SHA when known |
| `gc.formula_pack_revision` | locked revision or content hash used for compilation |
| `gc.formula_reproducibility` | `pinned`, `local-dirty`, `local-clean`, or `local-not-reproducibly-pinned` |
| `gc.formula_compile_artifact` | durable path or bead id for the compile result/provenance snapshot when the metadata would exceed practical size |

`CompileResult.Provenance` has a structured shape, not a formatted string:

```go
type Provenance struct {
    FormulaName     string
    SourcePath      string
    LayerName       string
    LayerPriority   int
    Pack            PackProvenance
    Imports         []ImportProvenance
    ContentHash     string
    Reproducibility string
    Dirty           bool
}

type PackProvenance struct {
    Name                string
    Source              string
    RequestedRef        string
    RequestedConstraint string
    LockedRevision      string
    RequiresGC          string
    Dirty               bool
}

type ImportProvenance struct {
    ParentPack          string
    ImportSource        string
    RequestedRef        string
    RequestedConstraint string
    LockedRevision      string
    RequiresGC          string
    ContentHash         string
    Dirty               bool
    LayerPriority       int
    ContributedPath     string
}
```

<!-- REVIEW: added per provenance-owner-v21 -->

Formula resolution produces a typed `ResolvedFormulaSource`; compilation does
not rediscover pack context from paths:

```go
type ResolvedFormulaSource struct {
    FormulaName       string
    OriginalPath      string
    StagedPath        string
    LayerName         string
    LayerPriority     int
    Pack              PackProvenance
    Imports           []ImportProvenance
    RequestedRef      string
    RequestedConstraint string
    LockedRevision    string
    RequiresGC        string
    ContentHash       string
    Dirty             bool
    TransitiveSources []ResolvedFormulaSource
}
```

The resolver/import layer owns `ResolvedFormulaSource` construction. The
formula compiler consumes it and copies the relevant fields into
`CompileResult.Provenance`. No caller may reconstruct pack name, requested ref,
locked revision, dirty status, or transitive import chain from a staged path,
formula name, or bead metadata after compilation.

`ContentHash` is the lowercase hex SHA-256 of canonical source bytes after
formula resolution, before runtime variable substitution, and before any graph
transform. For a multi-file contribution such as expansion/aspect imports, the
hash input is a length-prefixed manifest of each original source path,
locked revision or local dirty marker, file mode, and raw bytes in deterministic
resolver order. The hash excludes absolute staging paths and generated
diagnostics so the same pinned pack content produces the same value across
machines.

Layer resolution must preserve pack binding before symlink staging. A staged
winner path under `.beads/formulas/` is not enough provenance by itself; the
compile result records the original layer, pack, import source/ref, locked
revision or local hash, dirty state, transitive import chain, and layer
priority that selected the winning file.

<!-- REVIEW: added per pack-provenance-compatibility -->

Pack compatibility is evaluated before formula selection, staging, or compile.
For every pack origin - built-in, local path, installed pack, git URL, registry
source, direct import, and transitive import - the resolver reads `[pack]
requires_gc` and rejects an incompatible pack before any formula from that pack
can win layer resolution. The rejection is a pack-resolution diagnostic, not a
formula diagnostic, but the provenance report must include the pack source,
requested ref or version constraint, locked revision, and parent import chain
that caused it.

Raising a formula's normalized `requires.formula_compiler` minimum is a pack
compatibility event. A released pack may raise that minimum only when the same
change raises or verifies `[pack] requires_gc` so consumers below the minimum
binary floor reject the pack at resolver time instead of selecting a formula
they cannot compile. This applies to first-party packs, examples, local packs
prepared for release, and external pack validation output.

Import upgrades must report requirement deltas before the upgraded pack is
accepted:

```bash
gc formula validate --pack-path ./packs/acme --all --requirement-diff old.lock new.lock --json
```

The JSON diff contains old and new pack source/ref/locked revision, each
formula whose normalized compiler requirement changed, the old and new
requirement source, content hash changes, transitive import changes,
`requires_gc` floor deltas, and whether the change is safe for the active
binary. CI for first-party imports fails when an import upgrade raises a
compiler requirement without the corresponding pack-floor and docs/release
artifact updates.

New producers must write the canonical keys. During the alias window they also
write `gc.formula_contract = "graph.v2"` for graph workflow roots so existing
readers continue to work. New consumers must use a shared predicate that reads
canonical metadata first and falls back to the legacy key only for compatibility.
If root metadata cannot carry the full provenance, the producer must write a
durable compile artifact before creating child beads and stamp
`gc.formula_compile_artifact` on the root. That artifact is immutable for the
root lifecycle and contains the full `CompileResult.Provenance`,
`NormalizedRequirements`, and diagnostics that were accepted at creation time.

Formula validation must expose a read-only provenance surface before built-in
packs are migrated:

```bash
gc formula validate <name> --provenance
gc formula validate --all-packs --provenance
gc formula validate --pack-path ./packs/acme --all --json
gc formula validate --pack-source https://example.com/acme.git --ref v1.2.3 --all --json
```

This command must not create or update beads. It reports:

- formula name and winning source file
- formula layer and pack binding that won resolution
- pack import source, ref, and locked revision when available
- local path imports as `local` with a content hash and dirty status when the
  path is under a VCS checkout
- local path imports outside a VCS checkout as `local-not-reproducibly-pinned`
- deprecated fields, invalid requirements, v2-only constructs, and normalized
  compiler requirement

External author validation is a stable user-facing surface, not only an
internal release gate. With `--json`, the response schema contains
`schema_version`, `pack`, `requested_ref`, `locked_revision`, `dirty`,
`reproducibility`, `formulas`, `imports`, `shadowed_formulas`, `diagnostics`,
and `migration_hints`. Exit code `0` means every selected formula is valid,
`2` means formula diagnostics were found, and `1` means the pack could not be
loaded. Tests must cover local clean packs, local dirty packs, imported winners,
transitive imports, lockfile-backed refs, remote SHA pins, and shadowed formulas
whose losing source still contains legacy `contract`.

Shadowed formula diagnostics default to warning, not silence. When a losing
layer has a stricter requirement than the winning formula, an unsupported
future requirement, or legacy-only `contract`, validation reports a
non-blocking shadow diagnostic with the winner path, losing path, layer
priority, and remediation. Release gates may promote that warning to failure
for first-party packs because a later layer-order change could expose the
shadowed formula.

`migration_hints` is stable JSON, not prose. Each hint contains:

```json
{
  "code": "formula.migration.add_requires",
  "severity": "info",
  "formula": "code-review-loop",
  "source_path": "packs/acme/formulas/code-review-loop.toml",
  "source_key": "contract",
  "current": "contract = \"graph.v2\"",
  "recommended": "[requires]\\nformula_compiler = \">=2\"",
  "first_party": false,
  "pack": "acme",
  "requested_ref": "v1.2.3",
  "locked_revision": "abc123",
  "requires_gc_floor": ">=<minimum-floor>",
  "safe_automatic_edit": true
}
```

Required hint codes are `formula.migration.add_requires`,
`formula.migration.keep_dual_declaration`,
`formula.migration.remove_legacy_contract_when_floor_allows`,
`formula.migration.raise_pack_requires_gc`, and
`formula.migration.pin_pack_revision`. JSON field names, codes, and severities
are versioned by `schema_version`; external pack tooling must not parse
human-readable diagnostic messages.

Operational meaning of reproducibility values:

| Value | Meaning | Policy effect |
|---|---|---|
| `pinned` | Source comes from a locked pack revision or exact SHA | acceptable for release gates |
| `local-clean` | Local source is under VCS with no changes and a content hash | acceptable for development; release gate requires explicit approval |
| `local-dirty` | Local source has uncommitted changes | validation reports warning; release gate fails |
| `local-not-reproducibly-pinned` | Local source is outside a known VCS checkout or lacks a stable ref/hash | validation reports warning; release gate fails |

Canonical workflow-root predicate semantics:

| Root metadata shape | `IsWorkflowRootMetadata` | `IsGraphWorkflowMetadata` | Query behavior |
|---|---|---|---|
| canonical keys only, compiler capability `2` | true | true | matched by all new queries |
| canonical keys only, compiler capability `1` | true | false | matched as workflow root, not graph workflow |
| canonical keys only, compiler capability `3` or higher on an old binary | true | false | matched as unknown future workflow root for observation/cleanup only; graph-specific writes fail closed |
| dual-stamped canonical plus `gc.formula_contract=graph.v2` | true | true | preferred migration shape |
| legacy-only `gc.formula_contract=graph.v2` | true during alias window | true during alias window | legacy fallback; validation reports migration warning |
| raw `contract` string copied into arbitrary metadata | false | false | not a durable workflow-root signal |
| no canonical or legacy workflow metadata | false | false | ignored by workflow-root queries |

Raw metadata query construction for workflow roots is owned by the shared
predicate package only. New production code outside that owner may not filter
directly on `gc.formula_contract`, `gc.formula_compiler_requirement`, or
`gc.formula_compiler_capability`; it calls the predicate/query helper so
legacy-only, dual-stamped, graph-v2-only, and requires-only roots stay visible
through the migration.

The `gc.formula_*` metadata namespace is owned by `internal/formula` and the
shared predicate package. Matching is exact after trimming surrounding
whitespace only where legacy data already requires it: canonical keys use exact
values (`">=2"`, `"1"`, `"2"`, `requires`, `contract`, `dual`), while the
legacy `gc.formula_contract` fallback accepts only the byte string `graph.v2`
after trimming historical whitespace. No caller may branch on prefixes,
substrings, case-folded variants, or arbitrary `gc.*` keys as formula signals.

`version` remains accepted as legacy input for now. It is preserved only as
legacy metadata and may produce `formula.version_deprecated` on user-facing
validation paths. It is not a compiler selector and not a formula artifact
version.

## Terminology And Documentation Contract

<!-- REVIEW: added per docs-terminology-version -->

Glossary:

| Term | Meaning |
|---|---|
| Formula compiler capability | Numeric capability level the formula requires; v0 supports `1` and `2` |
| Host capability | What the active Gas City binary and current config can satisfy |
| Compiler implementation | Internal code path chosen by the binary, never by the formula |
| `contract` | Deprecated legacy spelling for graph-v2 compiler requirement |
| `schema` | Physical config/schema description, not compiler capability |
| Pack revision | The artifact identity and reproducibility boundary for formulas |
| Formula `version` | Legacy formula metadata only; not semver and not a compiler selector |
| Pack `requires` / `requires_gc` | Pack-level compatibility constraints, distinct from formula-level `[requires]` |
| Graph workflow | Persisted workflow-root graph produced when normalized compiler capability is `2` |

Docs and examples must be updated in the same release phase that makes
diagnostics user-visible:

| File family | Required update |
|---|---|
| `docs/reference/formula.md` | Canonical `[requires]` syntax, grammar, diagnostics, migration notes |
| `docs/reference/config.md` and generated schema docs | Host `[daemon] formula_v2` capability wording and legacy `graph_workflows` interaction |
| `docs/reference/cli.md` and generated command help | `gc formula validate --provenance` and `--legacy-contract-report` |
| `engdocs/architecture/formulas.md` | Internal ownership of compile result, diagnostics, and host capabilities |
| `engdocs/proposals/formula-migration.md` | Compatibility gates, minimum binary floor, external SHA-pinned behavior |
| Tutorials and examples under `examples/` and first-party packs | Dual declarations during alias window, `[requires]` as canonical prose |
| Test fixtures under `internal/testfixtures/` and `test/` | Explicit old, dual, and requires-only cases |
| Dashboard generated types | Typed diagnostics and workflow requirement fields |
| PackV2 author docs | Pack-level compatibility constraints, import refs, and formula `[requires]` distinction |

User-visible diagnostic rollout is blocked until the docs/example bundle lands
in the same PR stack. A diagnostic is user-visible when it can appear in CLI
stderr, HTTP/API responses, dashboard state, generated client types, controller
logs, order events, or release validation output.

<!-- REVIEW: added per docs-rollout-gate -->

This gate is phase-blocking, not advisory. A PR that exposes any formula
requirement diagnostic through CLI, API, dashboard, generated client types,
controller logs, order events, or release validation output must include the
docs/example bundle below and the generated-help refresh in the same branch.
There is no feature-flag exception: hidden or experimental diagnostics still
count as user-visible once an operator can trigger or observe them.

`docs/reference/formula.md` has a required rewrite skeleton, not just a list of
topics:

1. Formula identity and pack revision boundary: formula `version` is legacy
   metadata only; pack ref/version/SHA is the reproducibility boundary.
2. Canonical `[requires]` grammar: omitted, `>=1`, `>=2`, wrong-type and
   malformed-value examples, and unsupported future `>=3` behavior.
3. Host capability: `[daemon] formula_v2` as an operator capability gate, not
   a compiler selector.
4. Graph workflow examples: modern requires-only example for new binaries and
   dual-declared alias-window example for first-party packs.
5. Migration from `contract`: deprecation warning, conflict behavior,
   legacy-only report command, and alias removal criteria.
6. Pack compatibility: `[pack] requires_gc`, import refs, pinned SHA behavior,
   and requirement-diff reports.
7. Diagnostics: stable code table, CLI/API/dashboard projection, and
   remediation wording.

The docs bundle must include these exact TOML snippets:

```toml
formula = "code-review-loop"

[requires]
formula_compiler = ">=2"
```

```toml
[pack]
name = "acme-workflows"
requires_gc = ">=<minimum-floor-from-formula-compiler-min-floor.json>"
```

The formula `version` policy is fixed for rollout: preserve the field as legacy
metadata, emit `formula.version_deprecated` only on validation/display
surfaces, never treat it as compiler capability, and do not schedule removal
until a separate deprecation design names reader support, release timing, and
conversion tooling.

Requirement-surface comparison that must appear in reference docs:

| Surface | Owner | Purpose | Example |
|---|---|---|---|
| Formula `[requires].formula_compiler` | formula author | Minimum formula compiler capability needed by this formula | `formula_compiler = ">=2"` |
| Host `[daemon].formula_v2` | city operator | Whether this binary/config may satisfy compiler capability 2 | `formula_v2 = true` |
| Pack `requires_gc` | pack author | Minimum Gas City binary or pack schema compatibility | `requires_gc = ">=0.x"` |
| Pack import `version`/ref | pack consumer | Which pack revision to resolve | tag, branch, semver, or SHA |
| Formula `version` | legacy formula metadata | Preserved for compatibility; not a selector | `version = 2` |
| Formula `contract` | legacy formula authoring | Deprecated alias for compiler capability 2 | `contract = "graph.v2"` |

Canonical modern example:

```toml
formula = "code-review-loop"

[requires]
formula_compiler = ">=2"
```

Alias-window dual-declared example for first-party graph formulas:

```toml
formula = "code-review-loop"
contract = "graph.v2"

[requires]
formula_compiler = ">=2"
```

Generated-help and stale-guidance gate:

| Check | Fails when |
|---|---|
| `docs/reference/formula.md` examples | a graph-v2 example lacks `[requires]` or presents `contract` as canonical outside migration notes |
| `docs/reference/cli.md` / generated command help | formula validation flags omit `--provenance` or `--legacy-contract-report` |
| `docs/reference/config.md` | `[daemon] formula_v2` is described as selecting a compiler implementation instead of declaring host capability |
| tutorials and `examples/` | new graph workflow snippets are not dual-declared during the alias window |
| first-party pack snippets | source conversion happens before the minimum binary floor gate |
| dashboard generated types | diagnostic or workflow requirement fields are stale against OpenAPI |

Generated source ownership:

| Artifact | Source of truth | Regeneration command |
|---|---|---|
| `docs/reference/config.md` and city schema | config structs and schema generator | `make generate` |
| `docs/reference/cli.md` and command help | Cobra command definitions | `make generate` or `go run ./cmd/gc gen-doc` |
| `internal/api/openapi.json`, `docs/schema/openapi.json`, dashboard generated TS | Huma route registrations and API structs | `make dashboard-check` |
| formula validation fixture table | `internal/formula/testdata/compiler_requirements_matrix.yaml` | `go generate ./internal/formula` |

Explicit exceptions are limited to tests for legacy behavior, migration notes
that label `contract` as deprecated, and historical architecture text under
`engdocs/archive/`. The stale-guidance check must print each matched file and
line so authors can fix drift without guessing.

`formula.version_deprecated` is emitted only by validation/display surfaces
(`gc formula validate`, `gc formula show`, and formula API previews). Launch,
order dispatch, retry, convergence, and controller paths preserve legacy
`version` as metadata silently so operational logs are not polluted by an
artifact-identity warning after a formula has already been accepted.

## Rollout Plan

<!-- REVIEW: added per reversible-rollout -->

Rollout is split so `main` can stay green and each phase has a narrow rollback.

1. Parser and model: add `Requires`, `NormalizedRequirements`, diagnostics, and
   validation tests. Keep existing callers on current behavior.
2. Compile result and acceptance: add `CompileResult`,
   `AcceptedCompileArtifact`, canonical metadata keys, dual-stamping, and the
   shared workflow-root predicate. Existing legacy consumers still work through
   the fallback.
3. Caller migration: move callers to the normalized result in reversible
   sub-phases. The static no-raw-consumer guard starts in report-only mode and
   becomes blocking only after the final caller sub-phase lands.
4. Compatibility bridge: keep first-party graph formulas dual-declared while
   any supported production path can still shell out to a compiler that only
   understands `contract`.
5. Docs and examples: after parser and caller support ships, update
   `docs/reference/formula.md`, architecture docs, tutorials, examples,
   testdata, config references, and generated CLI docs to teach `[requires]` as
   canonical. Legacy `contract` appears only in migration notes.
6. Pack-floor enforcement: land resolver/import enforcement for `[pack]
   requires_gc`, seed the compatibility/minimum-floor/external-support
   artifacts, and prove incompatible packs are rejected before formula
   selection, staging, shadow resolution, or durable writes.
7. First-party requires-only conversion: remove first-party `contract` stamps
   only after the minimum binary floor is enforced and the `bd` compatibility
   strategy is complete.
8. Alias removal: remove legacy `contract` support only after the measurable
   alias-window criteria above pass.

Rollback notes:

- Phases 1 and 2 are additive and can be reverted without changing formula
  source files.
- Phase 3 can fall back to legacy predicates because roots are dual-stamped.
- Phase 4 keeps dual source declarations, so old binaries still read built-in
  graph formulas.
- Phases 7 and 8 require an explicit release decision because they can affect
  externally pinned packs.

Caller migration sub-phases:

| Sub-phase | Scope | Required gate | Rollback |
|---|---|---|---|
| 3a shared result plumbing | `CompileResult`, `AcceptedCompileArtifact`, typed diagnostics, metadata writer, workflow-root predicate | formula unit tests and predicate parity tests | callers still use legacy compile wrapper until durable writers switch |
| 3b sling and CLI launch | `gc sling`, `internal/sling`, force replacement, runtime-var validation | enabled/disabled host tests and no-partial-root tests | fall back to dual-stamped legacy predicate |
| 3c orders and controller scans | formula-backed order dispatch and controller validation producers | repeated scan grouping test and successful-later-order test | disable new producer event while keeping synchronous diagnostics |
| 3d API and generated clients | Huma handlers, OpenAPI, generated dashboard TS | `make dashboard-check` and OpenAPI in-sync | keep legacy HTTP error projection behind a feature flag |
| 3e convoy/source workflow | source-workflow scans, convoy dispatch and cleanup | canonical-only, legacy-only, dual-stamped, mixed-store root tests | legacy fallback remains in shared predicate |
| 3f convergence and molecule execution | convergence create/retry/speculative wisp, molecule cook/cook-on, fanout fragments | zero-write tests for every durable write boundary | reject new convergence/fanout formulas while legacy roots continue |
| 3g dashboard rendering | dashboard state and generated diagnostic rendering | generated TS compile and dashboard state tests | dashboard hides diagnostics but API remains typed |
| 3h blocking static guard | production raw-consumer denylist becomes CI-blocking | repo-wide guard has only named exemptions | revert guard to report-only |

Documentation and source conversion are separate gates:

| Gate | Scope |
|---|---|
| Docs prose | reference docs, architecture docs, PackV2 docs, tutorials, generated CLI/config docs |
| Runnable examples | examples and tutorial commands that CI executes |
| First-party dual declarations | built-in packs, examples, testdata, and fixtures carry both `contract` and `[requires]` |
| First-party requires-only conversion | only after minimum binary floor, external support plan, and `bd` probe retirement/parity gates pass |

## In-Flight And Convergence Behavior

<!-- REVIEW: added per in-flight-convergence-behavior -->

Compiler requirements are evaluated when a formula is compiled for a new root,
wisp, attached molecule, expansion, or convergence instance. They are not
re-evaluated for already-created beads that are merely being observed,
dispatched, retried, closed, or finalized.

Rules:

- If `[daemon] formula_v2` changes from true to false, existing graph workflow
  roots and their already-created step beads continue through dispatcher,
  retry, scope-check, convoy cleanup, and workflow-finalize paths using their
  persisted metadata.
- New `gc sling --formula`, API sling, formula-backed order dispatch, new
  convergence root creation, and speculative wisp creation must preflight the
  formula with the current host capability. On
  `formula.compiler_requirement_unsatisfied`, they create no root, child bead,
  or partial convergence state.
- Retrying an existing step that does not compile a new formula continues even
  if the host flag has since been disabled.
- A retry or `on_complete` action that compiles a new formula uses the current
  host capability and can fail with the shared diagnostic before creating the
  new attached molecule.
- Convergence adapters must use the same diagnostic path for create, retry,
  and speculative-wisp entry points. A failed preflight produces no active
  loop with missing children.
- Dispatch fanout compiles expansion fragments before root mutation. If any
  fragment has an unsupported or unsatisfied compiler requirement, fanout
  returns the shared diagnostic and writes no fragment beads, convoy links, or
  continuation-group metadata.

Convergence implementation decision:

The convergence formula subset parser is not an independent compiler. It is
rewritten as a typed projection over generic `CompileResult` fields, followed
by convergence-only domain validation. Domain validation may check convergence
semantics such as allowed retry shape, but it may not interpret raw `contract`,
raw `[requires]`, graph metadata strings, or host capability. That keeps
formula syntax, requirement satisfaction, and diagnostic projection owned by
`internal/formula` while keeping convergence-specific API surface inside
`internal/convergence`.

`internal/formula` exposes only generic compiled facts needed by all
execution paths:

```go
type CompileResult struct {
    Recipe       *Recipe
    Requirements NormalizedRequirements
    GraphWorkflow bool
    Diagnostics  []Diagnostic
    Provenance   Provenance
    Steps        []CompiledStep
    RuntimeVars  []CompiledRuntimeVar
}
```

<!-- REVIEW: added per convergence-boundary -->

`internal/convergence` owns the projection API:

```go
type SourceValue[T any] struct {
    Value      T
    SourcePath string
    SourceKey  string
    Line       int
    Column     int
}

type ConvergenceStepIdentity struct {
    StepID    SourceValue[string]
    StepTitle SourceValue[string]
}

type ConvergenceMetadata struct {
    Enabled             SourceValue[bool]
    RequiredVars        []SourceValue[string]
    EvaluatePrompt      SourceValue[string]
    EvaluatePromptPath  SourceValue[string]
    SourceKey           string
    RelevantStep        ConvergenceStepIdentity
    ReservedStepIDs     []SourceValue[string]
    RetryPolicy         CompiledRetryPolicy
    RuntimeVars         []CompiledRuntimeVar
    SourceFormula       string
    SourcePath          string
    Provenance          Provenance
    Requirements        NormalizedRequirements
    CompileArtifactRef  string
}

func ProjectFormula(result *formula.CompileResult) (ConvergenceMetadata, []formula.Diagnostic) // preview only
func ProjectAcceptedFormula(artifact formula.AcceptedCompileArtifact) (ConvergenceMetadata, []formula.Diagnostic)
func ValidateProjection(meta ConvergenceMetadata) []formula.Diagnostic
```

`ProjectFormula` may report convergence-domain diagnostics such as a missing
evaluate prompt for preview paths, but it may not reinterpret raw formula
requirement fields or host capability. Durable convergence paths use
`ProjectAcceptedFormula`. The projection is the canonical source-attributed
carrier for convergence enablement, required vars, evaluate prompt content,
prompt path, convergence source key, and convergence step identity; no caller
may read those fields from raw TOML, root metadata, or a convergence-only
subset parser. `CreateConvergenceBead`, retry root creation, next iteration,
missing-child repair, speculative wisps, and the first `PourWisp` call all run
after this order: `CompileWithResult`, `AcceptCompileResult`,
`ProjectAcceptedFormula`, `ValidateProjection`, accepted artifact persistence
when needed, then durable root, metadata, wisp, retry, iteration,
missing-child, or child writes.

Existing convergence roots use persisted accepted compile artifacts for their
current formula even if the host capability is later downgraded. Any operation
that selects or compiles a new formula after the downgrade fails closed before
durable writes. This chooses artifact reuse for existing roots and fail-closed
semantics for new or changed formulas.

<!-- REVIEW: added per convergence-canonical-projection -->

`internal/convergence/formula.Formula` and `ValidateForConvergence` are
retired or rewritten as compatibility shims over `ProjectFormula`. They may not
parse TOML, read raw `[requires]`, inspect `contract`, or reconstruct required
vars from a formula subset. A static guard under `internal/convergence` fails
when production code imports the legacy subset parser, reads raw formula files,
or searches for convergence fields outside the compiled projection.

Convergence-specific validation runs only after formula acceptance:

1. Compile with current host capability and caller vars/options.
2. Accept the compile result, rejecting fatal diagnostics and persisting a
   compile artifact when required.
3. Project typed convergence fields from canonical compile output.
4. Validate convergence domain rules: `Enabled`, required vars, evaluate prompt
   content, reserved step IDs, retry policy, and convergence-relevant step
   identity.
5. Write convergence root, iteration, retry, missing-child, or speculative-wisp
   state.

The durable metadata contract for convergence contains
`gc.convergence_formula_compile_artifact`, `gc.convergence_formula_name`,
`gc.convergence_formula_content_hash`, `gc.convergence_formula_pack_revision`,
and `gc.convergence_formula_host_capability`. Retry and next-iteration paths
first read the existing artifact ref. If the retry reuses the same formula,
vars hash, options hash, and content hash, it reuses the persisted artifact
even after a host downgrade. If any identity changes, it must compile and accept
again against the current host capability before writing retry metadata.

Required migration rows:

| Path | Required signature/flow | Zero-write test |
|---|---|---|
| `internal/convergence` create | `CompileWithResult` -> `AcceptCompileResult` -> `ProjectAcceptedFormula` -> convergence validation -> write root | disabled host leaves store unchanged |
| convergence retry | reuse persisted accepted compile artifact for existing root, or compile/accept new target formula before retry write | disabled host leaves retry metadata unchanged when compiling new formula |
| missing-child repair | load and validate the persisted accepted artifact before deciding whether a child should exist; compile/accept again only when the selected formula identity changes | disabled host does not create replacement child from a new formula |
| speculative wisp creation | compile/accept all candidate formulas before first speculative write | disabled host writes no speculative root or child |
| `internal/dispatch/fanout.go` fragment expansion | parent accepted artifact plus fragment `CompileWithResult`/`AcceptCompileResult` calls before expansion writes | disabled host writes no fragment child, convoy, or continuation |
| next-iteration convergence/fanout path | compile/accept iteration formula before iteration bead creation | disabled host writes no next-iteration bead |

Required tests cover enabled and disabled host capability for CLI sling, API
sling, order-created wisps, convergence root creation, convergence retry,
scope fragments, expansion/aspect requirements, and continuation of an
already-created graph workflow after the flag is disabled.

## Forward Compatibility

<!-- REVIEW: added per forward-compatibility-axis -->

V0 is intentionally fail-closed. Unknown `[requires]` axes and unsupported
compiler expressions are hard validation errors, not ignored future hints. That
lets old binaries distinguish "upgrade required or misspelled key" from a
formula they can safely run.

Capability semantics:

| Topic | Rule |
|---|---|
| Compiler capability numbers | Monotonic minimums. A host at capability `N` satisfies formula requirement `<=N`. |
| Accepted v0 syntax | Closed grammar: omitted, `>=1`, and `>=2` only. |
| Unsupported future capability | `>=3` is `formula.compiler_requirement_unsupported_future` until the binary implements capability 3. |
| Invalid syntax | Bad spacing, exact numbers, decimals, empty strings, and wrong TOML types remain `formula.compiler_requirement_invalid_syntax` or type diagnostics, not upgrade hints. |
| Unknown axis | Always `formula.requirement_unknown_axis`; old binaries do not ignore future axes. |
| Extension namespace | No user-defined axes in v0. A future extension axis must be a first-class typed field, not `x-*` passthrough. |
| Zero value | `CompilerCapability(0)` is an internal bug and must never escape normalization. |
| Provenance | Each axis records source path/key/value and normalized value independently. |
| Host capability growth | Guarded named fields on `HostCapabilities`; no raw string map until at least two implemented axes prove a generic representation is simpler. |
| Requirement source | `RequirementSource` is attribution only; behavior uses normalized capability values, never source spelling. |
| Construct registry | Every construct registry entry has `min_compiler_capability` and source attribution; no boolean "is v2" helpers in new code. |

Every future requirement axis must add all of these in the same change:

- typed normalized state, not raw maps or raw TOML pass-through
- accepted grammar and rejected-value tests
- diagnostic codes, canonical remediation, and projection behavior
- provenance fields and persisted root metadata when the axis affects runtime
- docs, examples, generated schemas/help, and stale-guidance checks

<!-- REVIEW: added per future-capability-guardrails -->

`RequirementSource` misuse is CI-guarded. Production code outside diagnostic,
provenance, migration-report, and compatibility-warning paths may not branch on
`RequirementSourceRequires`, `RequirementSourceContract`, `RequirementSourceDual`,
or `RequirementSourceOmitted`. Runtime behavior branches on
`NormalizedRequirements.FormulaCompiler()` and host capability only. The guard
prevents a future bug where dual declarations, legacy aliases, or source
spelling accidentally change execution semantics.

The construct registry is shaped for future capabilities:

```go
type ConstructCapability struct {
    ID                    string
    MinCompilerCapability CompilerCapability
    Locations             []ConstructLocation
    DiagnosticSource      SourceAttribution
}
```

Adding capability `3` means adding registry entries or raising existing
entries from `2` to `3`, extending the matrix dimensions, updating docs and
pack-floor artifacts, and adding old-reader fixtures that prove older binaries
fail closed. Old readers that encounter persisted metadata for a future
capability, for example `gc.formula_compiler_capability=3`, must classify the
root as a workflow root when canonical workflow-root keys are otherwise valid
but must not classify it as a graph workflow whose semantics they understand.
They emit or surface an unknown-capability diagnostic instead of silently
running new semantics.

<!-- REVIEW: added per DR33-future-root-lifecycle -->

Future-root lifecycle rules for old binaries:

| Operation | `unknown_future_capability_workflow` behavior |
|---|---|
| Observation, list, API/dashboard read models, and convoy/source scans | include the root with an unknown-capability diagnostic and no graph-specific inference |
| Cleanup of already-closed or orphaned records | allowed only when cleanup does not inspect or synthesize future graph semantics |
| Retry, fanout, continuation, next iteration, missing-child repair, child creation, or `on_complete` launch | fail closed before writes with an unknown-capability diagnostic |
| Selecting or compiling a new formula that requires the future capability or future axis | fail closed through parser/host capability diagnostics before writes |
| Existing accepted artifact load | allowed only to display provenance and diagnostics; not accepted as authorization for graph-specific writes the old binary cannot understand |

Old-reader fixtures must cover persisted roots with
`gc.formula_compiler_capability=3`, future `[requires]` axes, missing artifact
refs, stale artifact refs, dual-stamped legacy metadata plus future canonical
metadata, and dashboard/API projections. The expected result is visibility for
operators, zero graph-specific durable writes, and stable remediation to
upgrade Gas City.

Every new compiler capability or requirement axis has a release checklist:

| Checklist item | Required evidence |
|---|---|
| parser and matrix | byte-exact grammar, rejected values, unsupported future rows |
| host capability | typed config edge conversion and invalid-value diagnostic |
| durable metadata | canonical keys, old-reader behavior, accepted artifact fields |
| pack compatibility | `requires_gc` floor and requirement-diff report |
| diagnostics | CLI/API/dashboard/event golden projections |
| docs and generated artifacts | reference docs, examples, schema/help/OpenAPI/TS updates |
| stale guards | raw-consumer, `RequirementSource`, and construct registry guards updated |

Compiler capability numbers are stable once released and are never reused for
different semantics. If a future capability such as `3` is introduced, older
binaries reject `>=3` with `formula.compiler_requirement_unsupported_future`;
newer binaries may accept it only after adding typed support, tests,
diagnostics, and provenance. If a future requirement is not monotonic, it must
use a different axis rather than overloading `formula_compiler`.

Worked second-axis example:

```toml
[requires]
formula_compiler = ">=2"
state_store = ">=2"
```

`state_store` is illustrative, not part of v0. Adding it requires
`HostCapabilities{FormulaCompiler: ..., StateStore: ...}`,
`NormalizedRequirements.StateStore()`, `formula.state_store_requirement_*`
diagnostics mirroring the syntax/type/future/unsatisfied split above,
provenance fields for `requires.state_store`, persisted metadata such as
`gc.formula_state_store_requirement`, validation-matrix dimensions crossing
`state_store` with formula compiler capability, pack floor artifacts naming
the first binary that satisfies the new axis, and workflow-root predicates that
continue to classify roots by graph workflow using formula compiler metadata
only. Old binaries fail closed with `formula.requirement_unknown_axis` instead
of silently running a formula whose state-store requirement they cannot satisfy.

## Consequences

- Formula consistency remains anchored at the pack revision boundary.
- The formula header declares requirements, not artifact identity and not
  runtime compiler choice.
- Existing graph-v2 formulas keep working through a measurable compatibility
  alias.
- Operators get one stable diagnostic contract across CLI, API, dashboard,
  orders, controller, and convergence.
- Implementation work is larger than a field rename because all behavioral
  consumers must move to the normalized result before first-party formulas can
  become requires-only.
