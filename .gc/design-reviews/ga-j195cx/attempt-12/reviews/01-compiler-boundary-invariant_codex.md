# Nadia Sorenson - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design now states the right ownership boundary: `internal/formula` is the only package allowed to interpret raw `contract`, `[requires]`, `version`, and v2-only construct strings.
- `CompileResult`, `NormalizedRequirements`, `Diagnostic`, and explicit `HostCapabilities` give callers typed facts instead of asking CLI, API, sling, orders, convergence, or dashboard code to re-create compiler decisions.
- The caller migration table and static guard are correctly aimed at the current failure mode: raw `gc.formula_contract`, `declaresGraphV2Contract`, and `Requires.FormulaCompiler` checks leaking into projections.

**Critical risks:**
- [Major] `HostCapabilities` still has two fields for one compiler authority: `FormulaCompiler CompilerCapability` and `FormulaV2 bool`. The text says `FormulaCompiler` is canonical and `FormulaV2` is legacy vocabulary, but the exported struct admits contradictory states such as `{FormulaCompiler: GraphV2, FormulaV2: false}`. That lets caller construction, not the active binary, become the effective compiler decision. The design should either remove `FormulaV2` from the internal capability type or require a single constructor/normalizer plus tests proving `CheckRequirements` rejects or canonicalizes inconsistent host capabilities.
- [Major] `CheckRequirements(req NormalizedRequirements, host HostCapabilities)` invites behavioral callers to manufacture `NormalizedRequirements` without compiling. The prose says callers that create roots, wisps, fragments, orders, or convergence instances must use `CompileWithResult`, but the API surface does not encode that rule. Add a requirement that production callers outside `internal/formula` may only pass requirements taken from a `CompileResult`, or replace the public helper with a `CheckCompileResultRequirements`-style entry point. The static guard should also reject composite literals of `NormalizedRequirements` outside `internal/formula` test fixtures.
- [Minor] The static guard exception set needs a concrete allowlist. "Compatibility metadata writer" and "shared workflow-root predicate" are the right exceptions, but if the guard is regex-only and path-broad it can bless a second predicate in `internal/api`, `internal/sling`, or `internal/graphroute`. The design should name the exact package/function that owns persisted-root compatibility reads.

**Missing evidence:**
- No explicit invariant or test is specified for contradictory `HostCapabilities`.
- No enforcement detail shows that callers cannot synthesize `NormalizedRequirements` or call `CheckRequirements` as a substitute for `CompileWithResult`.
- The design does not name the final single shared workflow-root predicate package/function, even though the migration depends on eliminating duplicate raw metadata readers.

**Required changes:**
- Collapse host capability to one canonical value, or define a constructor/validation rule for `HostCapabilities` and add tests for every inconsistent field combination.
- Tighten the formula API contract so root/wisp/order/convergence callers cannot satisfy the compiler gate with caller-constructed normalized requirements.
- Add an explicit static-guard allowlist for raw legacy metadata reads and normalized-requirement construction.

**Questions:**
- Should `FormulaV2` remain only in `config.Daemon` and be translated at the CLI/API/controller edge into `CompilerCapability`, with no boolean field inside `internal/formula`?
- Is `internal/sourceworkflow.IsWorkflowRoot` intended to be the single compatibility predicate, or should the canonical predicate live in `internal/formula` with persistence adapters calling into it?
