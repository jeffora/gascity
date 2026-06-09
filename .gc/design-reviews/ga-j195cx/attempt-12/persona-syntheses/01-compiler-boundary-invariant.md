# Nadia Sorenson

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The v2-only construct registry is not an unambiguous contract. Claude identifies a table that mixes scanned locations, metadata keys, and metadata values under the same column, while Codex agrees the boundary depends on callers consuming typed compiler facts rather than recreating raw checks. The parser, missing-requirement diagnostic, static guard, and future-capability guard need one clear registry before the design can enforce the compiler boundary.
- [Major] `HostCapabilities` still exposes more than one authority for compiler capability. `FormulaCompiler` is described as canonical, but `FormulaV2` remains reachable in the same internal struct and can represent contradictory states. Requirement satisfaction must be driven by one typed host capability input, not package-global state or caller-side daemon flag checks.
- [Major] The `CheckRequirements(req, host)` and `NormalizedRequirements` API shape lets behavioral callers manufacture compiler requirements without compiling. Production paths that create roots, wisps, molecules, orders, convergence state, retries, or fragments must consume requirements from `CompileResult`, or the API and static guard must make caller-constructed requirements impossible outside formula tests and fixtures.
- [Major] The shared workflow-root predicate is still not pinned to one package, symbol, signature, and metadata precedence rule. The current "or" between `internal/formula` and `internal/sourceworkflow` leaves room for duplicate canonical-vs-legacy fallback logic in sling, convoy cleanup, API projection, dashboard mapping, or future callers.
- [Major] Convergence remains a drift vector unless the design commits it to canonical `internal/formula` output. Both reviews require convergence create, retry, and speculative-wisp paths to preflight through `CompileWithResult` or a smaller helper exported by `internal/formula`, with any convergence-specific validation limited to post-compile domain checks.
- [Major] Raw `Recipe` exposure is unresolved. If public fields such as `Recipe.Contract`, `Recipe.Requires`, `Recipe.Version`, or `Recipe.GraphWorkflow` remain available, callers can bypass `CompileResult.Requirements` or branch on derived topology fields. The design must either remove or unexport raw inputs, or enforce field-access-aware guardrails.
- [Major] The static guard is necessary but underspecified and too narrow. It needs a concrete implementation mechanism, exact file or symbol allowlist, denied raw metadata reads, denied field accesses, and coverage for typed drift such as `HostCapabilities.FormulaV2`, `RequirementSource`, direct `CompilerCapability` comparisons, `NormalizedRequirements` composites, raw `Recipe` fields, and diagnostic-code based requirement decisions.
- [Minor] The compatibility and alias-removal plan is incomplete. The design does not define an automatable `bd` shell-out parity proof, does not sequence that proof against the native formula-migration proposal, and does not include a phase that stops new producers from writing legacy `gc.formula_contract`.
- [Minor] Several vocabulary and invariant details still invite future confusion: `CompilerCapabilityGraphV2` names an implementation-shaped capability, zero-valued capability structs have undefined behavior, and dependent constructs such as `needs`, `loop`, and `compose.*` are described as conditionally v2-only instead of letting the actual v2-only construct trigger the requirement.

**Disagreements:**
- Claude and Codex both label their overall verdict `approve-with-risks`, but Claude also marks the construct-registry ambiguity as a blocker and both reviews identify open issues in the persona's core boundary invariant. Assessment: this synthesis chooses `block` because the remaining ambiguity is not cosmetic; it prevents an implementer from knowing which package owns compiler interpretation and which callers are forbidden from rechecking it.
- Claude emphasizes the v2-only construct registry, convergence drift, raw `Recipe` fields, constant naming, zero values, and alias-window closure. Codex emphasizes contradictory `HostCapabilities`, caller-synthesized `NormalizedRequirements`, and a concrete allowlist for raw legacy metadata reads. Assessment: these are complementary. The design should resolve all of them before approval from this lane.
- Claude allows a shared workflow-root predicate either in `internal/formula` or in `internal/sourceworkflow` if the latter is the sole persistence predicate calling formula helpers. Codex asks for the exact final package and function. Assessment: the design may choose either package, but it must remove the "or" and name the owner, signature, allowed metadata input, and guard exemption.
- Claude focuses on broad static-guard deny coverage, including typed-field drift. Codex specifically calls out normalized-requirement construction and path-broad allowlists. Assessment: the guard must be AST-aware for field access and composite literals; string search is only sufficient for literal legacy metadata strings.

**Missing evidence:**
- No Gemini review artifact is present for this persona.
- The final `internal/formula` compile/preflight API is not specified, including how typed host capabilities enter compilation and how package-global formula-v2 state is retired.
- The exact public `CompileResult`, `Recipe`, `NormalizedRequirements`, and `HostCapabilities` surfaces are not specified after migration.
- The canonical workflow-root predicate's package, symbol, signature, metadata precedence, fallback behavior, and static-guard exemption are not specified.
- The convergence entry points that must compile through `internal/formula` are named only generally, not as an enforceable caller-by-caller inventory.
- The static guard lacks a named implementation mechanism, exact allowlist, denied symbols, denied field accesses, denied metadata reads, and failing examples.
- Tests are missing for contradictory host capabilities, zero-valued capabilities, caller-constructed normalized requirements, unknown `[requires]` axes, and per-entry-point no-side-effect behavior when compiler requirements are unsatisfied.
- The `bd` compatibility proof and the phase that stops writing legacy `gc.formula_contract` are not defined.

**Required changes:**
- Rewrite the v2-only construct registry so scanned locations, metadata keys, and metadata values are distinct and unambiguous. Make that registry the named source of truth for parser detection, missing-requirement diagnostics, static guards, and future-capability checks.
- Collapse host compiler capability to one canonical internal value, or require a constructor/normalizer that rejects or canonicalizes inconsistent `HostCapabilities`. Add tests proving `CheckRequirements` depends only on the canonical capability.
- Tighten the formula API so behavioral callers cannot satisfy compiler gates with caller-constructed `NormalizedRequirements`. Production root, wisp, molecule, order, retry, convergence, and fragment paths should consume `CompileResult` end to end.
- Name and place the single workflow-root predicate, including signature, canonical-first and legacy-fallback behavior, relationship to workflow-kind metadata, and the precise static-guard allowlist entry.
- Commit convergence to `internal/formula` compilation or an `internal/formula` exported helper backed by the same `CompileResult`. Forbid convergence-local TOML decoding or raw `contract`, `requires`, or `version` interpretation.
- Resolve raw `Recipe` exposure by removing or unexporting raw input fields, or by adding a field-access-aware analyzer that blocks direct use outside the parser and explicitly allowed compatibility code.
- Define the invariant between `NormalizedRequirements.FormulaCompiler`, `CompileResult.GraphWorkflow`, and any `Recipe.GraphWorkflow` field, then test it for all construction paths and test helpers.
- Specify the static guard as an AST-aware test, vet analyzer, or equivalent enforceable tool with exact denied symbols, denied field accesses, denied metadata reads, denied composite literals, and exact file or symbol allowlists.
- Add tests for inconsistent host capabilities, zero-valued capability structs, unknown `[requires]` axes such as implementation selectors, and all entry points that must avoid partial side effects when compiler requirements are unsatisfied.
- Define the automatable `bd` compatibility criterion, sequence it with the native formula-migration proposal, add the alias-removal firing condition, and include a phase that stops new producers from dual-stamping `gc.formula_contract`.
- Rename implementation-shaped capability vocabulary such as `CompilerCapabilityGraphV2` to an implementation-neutral name such as `CompilerCapabilityV2`, and remove conditional v2-only treatment from dependent constructs that are not themselves requirement triggers.
