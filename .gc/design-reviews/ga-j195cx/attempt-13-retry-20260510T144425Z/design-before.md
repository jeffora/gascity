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

- Exact `formula_compiler = "2"` is not accepted in v0.
- `">= 2"`, `" >=2"`, `">=2 "`, `">=2.0"`, `">=2.1"`, and `">=3"` are not
  accepted in v0.
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
3. Unsupported `requires.formula_compiler` expressions.
4. Legacy `contract` normalization and `contract`/`requires` conflicts.
5. Missing compiler requirement for v2-only constructs.
6. Host capability failures such as `[daemon] formula_v2 = false`.

Accepted TOML shape matrix:

| TOML shape | Normalized result | Diagnostic |
|---|---|---|
| no `[requires]` table | default capability, source `omitted` | none |
| empty `[requires]` table | default capability, source `omitted` | none |
| `[requires] formula_compiler = ">=1"` | default capability, source `requires` | none |
| `[requires] formula_compiler = ">=2"` | compiler capability 2, source `requires` | none when host satisfies it |
| `[requires] formula_compiler = ""` | none | `formula.compiler_requirement_unsupported` |
| `[requires] formula_compiler = "2"` | none | `formula.compiler_requirement_unsupported` |
| `[requires] formula_compiler = ">= 2"` | none | `formula.compiler_requirement_unsupported` |
| `[requires] formula_compiler = ">=2 "` or `" >=2"` | none | `formula.compiler_requirement_unsupported` |
| `[requires] formula_compiler = ">=2.0"` or `">=3"` | none | `formula.compiler_requirement_unsupported` |
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
| Expansion/aspect contribution | `compose.expand`, `compose.map`, `compose.aspects` | contributed formula or template contains any construct in this registry |

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
- Inline children and `loop.body` inherit the containing formula's normalized
  requirement because they have no independent formula header.
- Expansion and aspect formulas are parsed and normalized independently. The
  compiled root requirement is the maximum requirement of the root formula and
  every expansion or aspect that contributed steps.
- If a formula, expansion, or aspect contains a v2-only construct but does not
  declare `requires.formula_compiler = ">=2"` or legacy `contract = "graph.v2"`,
  validation fails at that source formula with a diagnostic that names the
  missing requirement.

Required validation matrix:

| `contract` | `[requires] formula_compiler` | V2-only construct | Host `formula_v2` | Result |
|---|---|---|---|---|
| omitted | omitted or empty table | no | either | default capability, no diagnostic |
| omitted | `>=1` | no | either | default capability, no diagnostic |
| omitted | `>=2` | no or yes | true | graph capability, no error |
| omitted | `>=2` | no or yes | false | `formula.compiler_requirement_unsatisfied` |
| omitted | omitted or `>=1` | yes | either | `formula.compiler_requirement_missing` |
| `graph.v2` | omitted | no or yes | true | graph capability plus `formula.contract_deprecated` |
| `graph.v2` | `>=2` | no or yes | true | graph capability, source `dual`, plus `formula.contract_deprecated` |
| `graph.v2` | `>=1` | no or yes | either | `formula.compiler_requirement_conflict` |
| other value | any | any | any | `formula.compiler_requirement_unsupported` on `contract` |
| omitted | unsupported string such as `2`, `>= 2`, `>=2.0`, `>=3`, or empty string | any | any | `formula.compiler_requirement_unsupported` |
| omitted | wrong TOML type, dotted table, or nested table | any | any | `formula.requirement_invalid_type` |
| omitted | unknown `[requires]` key | any | any | `formula.requirement_unknown_axis` |

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
| unsupported `formula_compiler = ">=3"` plus disabled host | `formula.compiler_requirement_unsupported` |
| `contract = "graph.v2"` plus `formula_compiler = ">=1"` plus disabled host | `formula.compiler_requirement_conflict` |
| missing `>=2` for v2-only construct plus disabled host | `formula.compiler_requirement_missing` |

The implementation must add table-driven tests for accepted strings, rejected
strings, invalid TOML types, unknown `[requires]` keys, empty tables, legacy
contract compatibility, conflicts, missing requirements, inherited
requirements, expansion/aspect aggregation, loop bodies, children, and
unsupported future requirements.

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

type NormalizedRequirements struct {
    formulaCompiler CompilerCapability
    source          RequirementSource
    sourceFormula   string
    sourcePath      string
    deprecated      bool
}

type Diagnostic struct {
    Code                  string
    Severity              string
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
wrapper, but callers that branch on graph behavior, project diagnostics, create
workflow roots, or stamp metadata must use `CompileResult`. `Recipe` may carry
the normalized fields for instantiation, but it must not force consumers to
re-parse raw TOML strings.

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
func HostCapabilitiesFromFormulaV2(enabled bool, source string) HostCapabilities
func IsWorkflowRootMetadata(metadata map[string]string) bool
func IsGraphWorkflowMetadata(metadata map[string]string) bool
```

`CheckRequirements` returns diagnostics and never creates beads, writes
metadata, or mutates global state. Every caller that can create a root, wisp,
attached molecule, expansion fragment, order wisp, or convergence instance must
use `CompileWithResult` before the first durable write. Unit tests must
prove two compiles in the same process can evaluate the same formula against
different host capabilities deterministically, including concurrent compiles
with host capability `1` and `2`.

Caller inventory and required replacement behavior:

| Surface | Current risk | Required behavior |
|---|---|---|
| `internal/formula` parser, validation, and graph transforms | Raw `Contract` and `declaresGraphV2Contract` are load-bearing | Replace with `NormalizedRequirements` and `GraphWorkflow` |
| `internal/molecule` cook/cook-on and graph apply | Root metadata can be stamped from raw contract | Stamp from `CompileResult` only |
| `cmd/gc/cmd_sling.go` and `internal/sling` routing helpers | Graph routing and workflow attachment can branch on `gc.formula_contract` | Use shared workflow-root predicates backed by normalized metadata |
| `cmd/gc/cmd_order.go` and `cmd/gc/order_dispatch.go` | Order wisps can emit divergent errors | Preflight with `CompileResult`; emit the shared diagnostic event on failure |
| `internal/dispatch/fanout.go` and fanout expansion fragments | Runtime fragment compilation can happen after a durable root exists | Compile each fragment through `CompileWithResult` before fanout writes any child, convoy, or continuation metadata |
| `cmd/gc/cmd_convoy_dispatch.go` and convoy cleanup | Graph-only roots can be discovered by legacy metadata only | Use shared workflow-root predicate that accepts new and legacy metadata during migration |
| `internal/api/handler_sling.go`, formula endpoints, order feeds, and convoy projections | HTTP status and dashboard-facing errors can diverge | Project `Diagnostic` without hand-written JSON or string parsing |
| `internal/convergence` formula subset | Subset validation can drift from full compiler semantics | Validate through `internal/formula` preflight or an adapter over `CompileResult` |
| Dashboard generated types | UI can infer graph state from legacy metadata | Use API-projected typed fields and diagnostics |

<!-- REVIEW: added per caller-migration-executable -->

Executable call-site migration:

| Current file/helper | Target API | Required tests |
|---|---|---|
| `internal/formula/compile.go`: `isGraphWorkflow`, `declaresGraphV2Contract`, package-global `formulaV2Enabled` | `NormalizeRequirements`, `CheckRequirements`, and `CompileWithResult` with explicit `HostCapabilities` | `TestCompileWithResultHostCapabilitiesArePerCall`, concurrent enabled/disabled host tests |
| `internal/formula/types.go`: `requiresExplicitGraphContract`, `metadataRequiresGraphContract` | V2-only construct registry that returns source formula/path/key diagnostics | Parser matrix tests for every registry entry and nested location |
| `internal/molecule.Cook`, `CookOn`, graph-apply build path | Accept `CompileResult` or call `CompileWithResult`; stamp recipe/root metadata from `CompileResult.Requirements` | Cook and graph-apply tests assert canonical keys plus legacy key during alias window |
| `internal/sling.isGraphSlingFormula`, `validateSlingFormulaRuntimeVars`, `AttachFormula`, `LaunchFormula` | One preflight returning `CompileResult`; use `GraphWorkflow` and shared diagnostics for conflict and runtime-var paths | CLI and `internal/sling` tests for enabled, disabled, force replacement, and no partial root on unsatisfied host |
| `cmd/gc/cmd_sling.go` graph decoration | Use `CompileResult` provenance and canonical workflow-root metadata before route decoration | Existing graph sling tests updated to assert canonical metadata and no raw-contract branch |
| `cmd/gc/cmd_order.go` and `cmd/gc/order_dispatch.go` formula order cook | Preflight `CompileWithResult`; emit one registered failure event on fatal diagnostics; continue later orders | Order tests for unsatisfied host, deprecation warning suppression, and successful later order |
| `internal/dispatch/fanout.go`: `CompileExpansionFragment` and any continuation fanout helper | Accept a parent `CompileResult` and compile fragment formulas through the same options before durable fanout writes | Fanout tests prove disabled host capability creates zero child beads, convoys, continuation beads, or metadata |
| `cmd/gc/cmd_convoy_dispatch.go` source-workflow scans | Use `internal/sourceworkflow.IsWorkflowRoot` backed by canonical metadata first, legacy fallback second | Convoy tests with canonical-only, legacy-only, dual-stamped, and mixed-store roots |
| `internal/sourceworkflow.IsWorkflowRoot` and `ListLiveRoots` | Move predicate to `internal/formula` or keep this package as the sole persistence predicate that calls formula metadata helpers | Predicate parity tests shared by sling, convoy, API, and dispatch |
| `internal/graphroute.IsCompiledGraphWorkflow` | Read `Recipe.GraphWorkflow` or `CompileResult.GraphWorkflow`, not root metadata strings | Graph route tests prove metadata changes do not affect compiled result semantics |
| `internal/api/handler_sling.go` and `handler_formulas.go` | Project `Diagnostic` into Huma response structs; never parse stderr/error strings | API tests for HTTP 400 diagnostic payloads and generated OpenAPI/type updates |
| `internal/api/orders_feed.go` and `handler_convoy_dispatch.go` | Use shared workflow-root predicate for closed and open roots | API projection tests for canonical-only roots and legacy fallback |
| `internal/convergence/create`, retry, speculative wisp adapters | Preflight through `internal/formula` before any convergence bead/wisp write; keep convergence-only validation as post-compile domain checks | Convergence tests for disabled host capability with zero durable writes |
| Dashboard generated types and state | Consume typed API diagnostics and workflow requirement fields | Dashboard tests for diagnostic rendering without metadata/string inference |

The migration is not complete while any non-test production code outside
`internal/formula` or the shared persistence predicate branches on
`Contract`, `declaresGraphV2Contract`, `Requires.FormulaCompiler`, or
`gc.formula_contract`.

CI must include a static guard that fails on new behavioral uses of raw
`contract = "graph.v2"`, `declaresGraphV2Contract`, `Requires.FormulaCompiler`,
or `gc.formula_contract` outside the parser, compatibility metadata writer,
shared workflow-root predicate, tests, fixtures, and docs.

Durable preflight contract:

| Entry point | Must complete before first durable write | Writes forbidden on fatal diagnostic |
|---|---|---|
| Root molecule or wisp launch | `CompileWithResult` plus `CheckRequirements` | root bead, child beads, root metadata, convoy, hook |
| Attached molecule | compile attached formula with current host capability | attached child bead, dependency, hook, attachment metadata |
| Formula-backed order | compile selected formula before marking the order fired | wisp root, child beads, order fired metadata |
| Retry or `on_complete` that starts a formula | compile target formula before retry/on-complete state mutation | retry-run bead, attached molecule, continuation metadata |
| Fanout fragment | compile every fragment formula before fragment expansion | fragment child beads, convoy links, continuation group metadata |
| Convergence create/retry/speculative wisp | canonical compile before convergence-specific validation writes | convergence root, iteration bead, missing-child state, retry metadata |

No production path may accept a pre-parsed `NormalizedRequirements` argument
from a caller as proof of satisfaction. Passing a full `CompileResult` is
allowed only when it came from the same `CompileOptions.HostCapabilities` and
search-path set used for the durable operation.

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
| Unsupported future requirement such as `>=3` | Not supported | Fails with unsupported requirement | Fails with unsupported requirement |

First-party built-in and example graph formulas must stay dual-declared until
the minimum supported Gas City binary understands `[requires]` and either the
native compiler path is the only production path or the `bd` shell-out path is
proven to accept equivalent requirements. External SHA-pinned formulas that use
legacy `contract = "graph.v2"` remain valid through the alias window.

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
| `GC_NATIVE_FORMULA=false` or `bd` shell-out path | Must either preserve dual `contract` declarations or be proven to parse `[requires]` identically before source conversion |

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

Measurable release gates:

| Gate | Artifact | Blocks |
|---|---|---|
| Legacy inventory | `gc formula validate --all-packs --legacy-contract-report --json` saved in release artifacts | first-party requires-only conversion while `legacy_only > 0` |
| Provenance inventory | `gc formula validate --all-packs --provenance --json` | source conversion when any first-party graph formula lacks pack binding, locked revision, or content hash |
| Minimum binary floor | `docs/release/formula-compiler-min-floor.json` plus release-note entry | publishing requires-only first-party packs |
| External pinned-pack plan | release checklist entry naming alias support, compatibility branch, or LTS binary | removing `contract` alias |
| Mixed-store compatibility | test report for old/new controllers reading dual-stamped and canonical-only roots | retiring dual root stamps |
| `bd`/native parity | golden corpus result for native and `GC_NATIVE_FORMULA=false` fallback | source conversion when fallback still needs legacy `contract` |
| Stale guidance | CI report for docs/examples generated help | user-visible diagnostics |

`formula-compiler-min-floor.json` records the lowest Gas City version allowed
to consume first-party requires-only graph formulas, the first release that
understands `[requires]`, the first release that writes canonical root metadata,
and the release that stopped needing the `bd` compatibility path. Rollback is
to restore dual declarations and dual root stamps, not to reinterpret
requires-only formulas in old binaries.

Dual-stamp retirement criteria:

1. All supported readers use the shared canonical workflow-root predicate.
2. Mixed-version shared-store tests pass with canonical-only roots.
3. The external pinned-pack plan has shipped.
4. The previous two minor releases wrote dual stamps and emitted no known
   compatibility incident.

Alias removal criteria are stricter than dual-stamp retirement: the parser may
stop accepting legacy `contract` only after the external plan expires and the
legacy inventory gate has reported zero first-party legacy-only formulas for
two consecutive minor releases.

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
| `formula.compiler_requirement_unsupported` | error | The expression is not supported in v0 |
| `formula.compiler_requirement_conflict` | error | `contract` and `[requires]` disagree |
| `formula.compiler_requirement_missing` | error | A v2-only construct lacks a v2 requirement |
| `formula.compiler_requirement_unsatisfied` | error | Host config cannot satisfy the normalized requirement |
| `formula.contract_deprecated` | warning | Legacy `contract` spelling was accepted |
| `formula.version_deprecated` | warning | Legacy formula `version` was present and preserved only as metadata |

Diagnostic projection matrix:

| Code | Canonical message/remediation | CLI | API | Dashboard | Events |
|---|---|---|---|---|---|
| `formula.requirement_unknown_axis` | `unsupported [requires] key <key>` / remove it or upgrade to a binary that supports the axis | fatal stderr, exit 1 | HTTP 400 typed diagnostic | formula preview validation error | no event except controller/order failure wrapper |
| `formula.requirement_invalid_type` | `<key> must be a TOML string` / use `formula_compiler = ">=2"` | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_unsupported` | `unsupported formula compiler requirement <value>` / use omitted, `>=1`, or `>=2` | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_conflict` | `contract and [requires] disagree` / make both declare graph v2 or remove `contract` | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_missing` | `v2-only construct requires formula_compiler = ">=2"` / add `[requires] formula_compiler = ">=2"` | fatal stderr, exit 1 | HTTP 400 | formula preview validation error | no event except controller/order failure wrapper |
| `formula.compiler_requirement_unsatisfied` | `host formula compiler capability 1 does not satisfy >=2` / enable `[daemon] formula_v2` or choose a v1 formula | fatal stderr, exit 1; no beads created | HTTP 400 for validation/preview or 409 for launch conflict plus diagnostic body | disabled-capability state with remediation | registered compile-failure wrapper event |
| `formula.contract_deprecated` | `contract = "graph.v2" is deprecated` / use `[requires] formula_compiler = ">=2"` | bounded warning stderr by source/key | warning diagnostic in 200/preview response | non-blocking warning | suppressed warning event only for daemon/order paths |
| `formula.version_deprecated` | `formula version is legacy metadata` / use pack version/ref/SHA for artifact identity | warning only on validate/show, never on launch success | warning diagnostic on formula endpoints | non-blocking warning | none |

Projection rules:

- CLI commands that compile formulas print warnings to stderr once per
  `(code, source path, source key)` per command invocation. Fatal diagnostics
  exit non-zero and include the code, source path, offending value, normalized
  requirement, host capability, and remediation.
- API endpoints return typed diagnostic fields in Huma-registered response
  bodies. User-correctable formula input errors use HTTP 400; internal I/O
  failures remain HTTP 500.
- Dashboard state is derived from the API diagnostic projection, not by parsing
  stderr or root metadata strings.
- Order dispatch emits one typed order failure event per failed dispatch
  attempt and continues scanning later orders. Repeated deprecation warnings
  use `OnceKey` suppression and must not flood the event bus.
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
  suppression is per process invocation. Controller, order dispatch, and daemon
  suppression is in-memory, bounded LRU per city, reset on process restart and
  config reload, and scoped to warnings only. Fatal diagnostics are emitted once
  per failed dispatch/launch attempt so operators can correlate failed work
  without warning floods.

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
| `formula.compile_failed` | formula preview/validate paths that run under the controller | `FormulaDiagnosticPayload` |
| `order.formula_compile_failed` | formula-backed order dispatch | `FormulaDiagnosticPayload` plus order id |
| `convergence.formula_compile_failed` | convergence create/retry/speculative-wisp paths | `FormulaDiagnosticPayload` plus convergence id when one already exists |
| existing sling/API failure wrapper | sling launch and API launch conflicts | `FormulaDiagnosticPayload` embedded in the typed error body/event payload |

Every event constant above must be present in `events.KnownEventTypes` and have
`events.RegisterPayload(constant, FormulaDiagnosticPayload{})` or a typed
wrapper payload registered before it can ship. Envelope fields carry the city,
agent/session when known, and source operation; payload fields carry the formula
diagnostic. The deduplication key is the diagnostic `OnceKey` plus the operation
kind for warnings. Fatal compile failures are not deduplicated across separate
operator attempts.

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
    Name           string
    Source         string
    RequestedRef   string
    LockedRevision string
}

type ImportProvenance struct {
    ParentPack      string
    ImportSource    string
    RequestedRef    string
    LockedRevision  string
    LayerPriority   int
    ContributedPath string
}
```

Layer resolution must preserve pack binding before symlink staging. A staged
winner path under `.beads/formulas/` is not enough provenance by itself; the
compile result records the original layer, pack, import source/ref, locked
revision or local hash, dirty state, transitive import chain, and layer
priority that selected the winning file.

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
2. Compile result and metadata: add `CompileResult`, canonical metadata keys,
   dual-stamping, and the shared workflow-root predicate. Existing legacy
   consumers still work through the fallback.
3. Caller migration: move sling, orders, API, convoy, convergence, molecule,
   dashboard projections, and tests to the normalized result and shared
   diagnostics. Add the static no-raw-consumer guard.
4. Compatibility bridge: keep first-party graph formulas dual-declared while
   any supported production path can still shell out to a compiler that only
   understands `contract`.
5. Docs and examples: after parser and caller support ships, update
   `docs/reference/formula.md`, architecture docs, tutorials, examples,
   testdata, config references, and generated CLI docs to teach `[requires]` as
   canonical. Legacy `contract` appears only in migration notes.
6. First-party requires-only conversion: remove first-party `contract` stamps
   only after the minimum binary floor is enforced and the `bd` compatibility
   strategy is complete.
7. Alias removal: remove legacy `contract` support only after the measurable
   alias-window criteria above pass.

Rollback notes:

- Phases 1 and 2 are additive and can be reverted without changing formula
  source files.
- Phase 3 can fall back to legacy predicates because roots are dual-stamped.
- Phase 4 keeps dual source declarations, so old binaries still read built-in
  graph formulas.
- Phases 6 and 7 require an explicit release decision because they can affect
  externally pinned packs.

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
rewritten as a typed projection from `CompileResult` for fields convergence
needs, followed by convergence-only domain validation. Domain validation may
check convergence semantics such as allowed retry shape, but it may not
interpret raw `contract`, raw `[requires]`, graph metadata strings, or host
capability. That keeps formula syntax, requirement satisfaction, and diagnostic
projection owned by `internal/formula`.

Required migration rows:

| Path | Required signature/flow | Zero-write test |
|---|---|---|
| `internal/convergence` create | `CompileWithResult` -> convergence projection -> convergence validation -> write root | disabled host leaves store unchanged |
| convergence retry | reuse persisted accepted compile artifact for existing root, or preflight new target formula before retry write | disabled host leaves retry metadata unchanged when compiling new formula |
| speculative wisp creation | preflight all candidate formulas before first speculative write | disabled host writes no speculative root or child |
| `internal/dispatch/fanout.go` fragment expansion | parent compile result plus fragment `CompileWithResult` calls before expansion writes | disabled host writes no fragment child, convoy, or continuation |
| next-iteration convergence/fanout path | preflight iteration formula before iteration bead creation | disabled host writes no next-iteration bead |

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
| Unsupported future capability | `>=3` is `formula.compiler_requirement_unsupported` until the binary implements capability 3. |
| Invalid syntax | Bad spacing, exact numbers, decimals, empty strings, and wrong TOML types remain syntax/type diagnostics, not upgrade hints. |
| Unknown axis | Always `formula.requirement_unknown_axis`; old binaries do not ignore future axes. |
| Extension namespace | No user-defined axes in v0. A future extension axis must be a first-class typed field, not `x-*` passthrough. |
| Zero value | `CompilerCapability(0)` is an internal bug and must never escape normalization. |
| Provenance | Each axis records source path/key/value and normalized value independently. |

Every future requirement axis must add all of these in the same change:

- typed normalized state, not raw maps or raw TOML pass-through
- accepted grammar and rejected-value tests
- diagnostic codes, canonical remediation, and projection behavior
- provenance fields and persisted root metadata when the axis affects runtime
- docs, examples, generated schemas/help, and stale-guidance checks

Compiler capability numbers are stable once released and are never reused for
different semantics. If a future capability such as `3` is introduced, older
binaries reject `>=3` with `formula.compiler_requirement_unsupported`; newer
binaries may accept it only after adding typed support, tests, diagnostics, and
provenance. If a future requirement is not monotonic, it must use a different
axis rather than overloading `formula_compiler`.

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
