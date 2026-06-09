# Felix Berger - Codex

**Verdict:** block

**Top strengths:**
- The design correctly names `internal/convergence` as a formula-subset drift risk instead of treating convergence as an unrelated consumer.
- The canonical ownership rule is strong: only `internal/formula` may interpret raw `contract`, `[requires]`, `version`, or v2-only construct strings.
- The in-flight section explicitly includes convergence create, retry, and speculative-wisp paths in the shared host-capability behavior.

**Critical risks:**
- [Blocker] The convergence preflight boundary is not executable enough to guarantee zero durable writes. Current convergence creation creates a `type=convergence` root and writes metadata before `Store.PourWisp` reaches formula compilation. The design says `internal/convergence/create`, retry, and speculative wisp adapters must preflight through `internal/formula`, but it does not specify where the compile happens, how `HostCapabilities` and formula search paths reach that code, or what API prevents an implementer from only fixing `PourWisp`. That leaves the main red flag alive: a disabled `daemon.formula_v2` can still produce a partial convergence root if the preflight is placed too late.
- [Major] The relationship between `internal/convergence/formula.go` and the canonical compiler is under-specified. The subset validator currently models only convergence-specific fields and has no `Requires`/normalized requirement projection. The design should state whether this subset is derived from `CompileResult` after canonical TOML parsing, or whether the raw subset loader remains but is forbidden from accepting/dropping `[requires]`. Without that, convergence-specific validation and canonical formula validation can continue to drift.
- [Major] Convergence root provenance is ambiguous. The design requires workflow roots to be dual-stamped and says convergence instances evaluate compiler requirements, but it does not say whether the convergence root itself records the accepted `CompileResult`/requirement metadata or compile artifact. Since retries and future iterations are driven from convergence root metadata, the design needs a clear rule for what is persisted on the convergence root versus only on the child wisp/workflow root.
- [Minor] The docs update plan is broad, but the convergence-specific formula subset docs need an explicit acceptance criterion: `[requires]` is a top-level formula requirement validated by `internal/formula`, not a convergence-only key and not something the convergence subset may ignore.

**Missing evidence:**
- No concrete function signature or sequence for convergence create/retry preflight before `CreateConvergenceBead`.
- No stated test that instruments the convergence store and proves `CreateConvergenceBead`, `SetMetadata`, and `PourWisp` are not called on `formula.compiler_requirement_unsatisfied`.
- No migration rule for the existing `ValidateForConvergence` subset path: adapt from `CompileResult`, keep as post-compile domain validation, or retire raw TOML decoding.
- No explicit metadata/provenance contract for `type=convergence` roots that launch compiler-v2 formulas.

**Required changes:**
- Add a convergence-specific preflight contract with an ordering guarantee. For example: `CityRuntime.handleConvergenceCreate` and retry perform `internal/formula.CompileWithResult` with current `HostCapabilities` and formula search paths before calling `CreateHandler`, then pass the accepted compile/provenance information into `CreateParams`; or add an explicit convergence store/preflight interface that must run before `CreateConvergenceBead`. The design must make the late-`PourWisp` implementation invalid.
- Define how convergence-specific validation composes with `CompileResult`: canonical formula parsing and requirement validation first, convergence-only checks second, with no raw `[requires]` interpretation in `internal/convergence`.
- Add required tests for disabled host capability on convergence create, convergence retry, and speculative wisp creation that assert zero durable writes and shared diagnostic fields.
- Specify whether convergence roots persist canonical requirement/provenance metadata or an immutable compile artifact, and how that metadata is used for status/debugging without reinterpreting raw formula headers.
- Update the convergence subset documentation acceptance criteria so the public docs do not imply that the subset is a separate TOML schema that can drop `[requires]`.

**Questions:**
- Should convergence root creation compile once and pass the accepted result through to first-wisp creation, or is double compilation acceptable if both compiles are explicitly required to use the same host capability and source revision?
- Is `internal/convergence` allowed to import `internal/formula`, or should the controller/runtime layer own formula preflight and pass only typed convergence parameters into the lower-level handler?
- Should a convergence retry that compiles a new formula use the source bead's persisted accepted requirement as an audit baseline, or only the current formula source plus current host capability?
