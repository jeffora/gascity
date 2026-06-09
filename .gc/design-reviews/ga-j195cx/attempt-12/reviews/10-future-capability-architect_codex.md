# Future Capability Architect Review

Persona: Ibrahim Park
Verdict: block

## Summary

The revised design is much stronger than a simple `contract` rename. It now
puts requirement parsing in `internal/formula`, defines host satisfaction as a
typed compile-time check, rejects unknown future syntax, preserves a
compatibility window, and names the caller migration surface. That is the
right direction.

I still would not implement from it yet. The remaining gaps are concentrated in
the contract that future requirement axes will copy. The v0 design says
`FormulaCompiler` is the canonical host capability, but the proposed
`HostCapabilities` type also carries the legacy `FormulaV2` boolean. It also
keeps requirement source and diagnostics scalar even though the document now
commits to future axes and aggregation across root, inherited, expansion, and
aspect formulas. Those choices are small to fix in the design and expensive to
unwind after caller migration.

## Findings

### Blocking: `HostCapabilities` has two sources of truth

The proposed core API defines:

```go
type HostCapabilities struct {
    FormulaCompiler CompilerCapability
    FormulaV2       bool
}
```

The text then says the boolean is legacy config vocabulary and
`FormulaCompiler` is the canonical typed capability used by
`CheckRequirements`. That contradiction should not enter `internal/formula`.
If both fields can be present, implementations and tests must answer what
happens for impossible states such as `FormulaCompilerGraphV2` with
`FormulaV2=false`, or default compiler with `FormulaV2=true`. If the answer is
"never construct that," the API should make it unrepresentable.

This matters for future axes. The first new requirement axis will likely copy
this pattern and carry both a normalized capability and a legacy feature flag,
which recreates the exact caller-specific decision logic the design is trying
to eliminate.

Required change: remove `FormulaV2` from the formula-domain
`HostCapabilities` type, or move it to an edge-only config conversion type.
`CheckRequirements` should receive one canonical host capability value for the
formula compiler axis. Add a test that the config edge maps
`[daemon] formula_v2` to `CompilerCapabilityDefault` or `CompilerCapabilityV2`
before entering `internal/formula`, and that contradictory host states cannot
be represented.

### Major: Requirement provenance is still scalar

`NormalizedRequirements` has one `FormulaCompiler`, one `Source`, one
`SourceFormula`, and one `SourcePath`. `Diagnostic` similarly has string fields
for one normalized requirement and one host capability. That is enough for the
simplest root-formula case, but the design also says:

- parent requirements are inherited and children may raise but not lower them
- expansion and aspect formulas are parsed independently
- the compiled root requirement is the maximum of every contributing formula
- every future requirement axis must add typed state and provenance

The scalar shape cannot naturally explain "root omitted, expansion A requires
compiler v2, aspect B later requires runtime X," nor can it preserve all source
references needed for a durable compile artifact. Implementers will either pick
one source and lose evidence, or add ad hoc side channels when the next axis
arrives.

Required change: either make v0 explicitly single-axis and state that the first
future axis must replace this scalar shape before shipping, or define a
per-axis/per-source structure now, for example entries with `Axis`,
`Requirement`, `Capability`, `Source`, `SourceFormula`, `SourcePath`,
`SourceKey`, and `SourceValue`. The durable compile artifact and diagnostic
projection should use the same structure rather than separate string fields.

### Major: Aggregation and lowering behavior needs diagnostic rows

The document states that child formulas may raise but not lower parent
requirements, and that expansion/aspect requirements aggregate into the root.
The validation matrix does not include those cases. It covers local
`contract`/`[requires]` combinations, but not:

- parent `>=2`, child explicit `>=1`
- parent `>=1`, child `>=2`
- root omitted, expansion/aspect `>=2`
- aspect/expansion `>=2` with host capability disabled
- multiple contributors with the same maximum requirement but different source
  paths
- whether a non-contributing conditional expansion is ignored or still checked

For compiler capabilities, "maximum wins" is probably correct. For future axes,
the rule may be minimum version, exact feature set, backend capability,
provider capability, or a disallowed combination. The design needs to say that
aggregation semantics are per axis and must be defined with tests when the axis
is introduced.

Required change: add explicit matrix rows and a diagnostic code for lowering or
cross-source conflicts, or state that `formula.compiler_requirement_conflict`
covers those cases. Add tests for inherited lowering, inherited raising,
expansion/aspect aggregation, host-disabled aggregation failure, and the
"contributed steps" boundary.

### Minor: Capability naming still bakes in the current implementation

The design names the numeric value `CompilerCapabilityGraphV2`. That makes the
capability level sound like a graph implementation instead of a formula
compiler capability. The prose says formulas declare minimum compiler
capability and do not select an implementation, so the constant should follow
that model.

Recommended change: use revision-level names such as
`CompilerCapabilityV1` and `CompilerCapabilityV2`, while keeping "graph
workflow" as a derived recipe/runtime outcome.

### Minor: Unknown-axis remediation is intentionally vague

Failing closed is correct, but the diagnostic still says effectively "remove it
or upgrade." That is acceptable for v0, but the future-axis contract should
require the diagnostic payload to include the rejected axis name, source key,
and the current binary's supported axis list. Otherwise operators cannot tell a
misspelling from a pack authored for a newer Gas City without consulting docs.

## Required Changes

1. Remove the legacy `FormulaV2` boolean from the formula-domain host
   capability contract, or otherwise make contradictory host states impossible.
2. Make requirement/provenance representation per-axis and source-aware, or
   explicitly mark the scalar v0 type as temporary and require it to be
   widened before any future axis ships.
3. Add validation/diagnostic matrix rows for inherited lowering, inherited
   raising, expansion/aspect aggregation, and host-disabled aggregated
   requirements.
4. Rename compiler capability constants away from graph-specific terminology.
5. Add supported-axis metadata to unknown-axis diagnostics.

## Residual Risk After Changes

Once those changes are made, the design is viable for v0. Future axes should be
able to extend the same pattern without raw TOML passthrough, caller-specific
checks, or metadata-string inference.
