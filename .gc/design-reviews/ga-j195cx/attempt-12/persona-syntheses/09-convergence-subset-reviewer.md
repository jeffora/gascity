# Felix Berger

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The convergence preflight boundary is not executable enough to guarantee zero durable writes. Both reviews require convergence create, retry, and speculative-wisp paths to run canonical formula capability validation before any convergence root, child bead, metadata, or partial state is written. Codex explicitly blocks on the risk that current creation can write a `type=convergence` root before `Store.PourWisp` reaches compilation; Claude identifies the same gap through the missing `molecule.Cook`/`CookOn` plumbing.
- [Major] The relationship between `internal/convergence/formula.go`, `ValidateForConvergence`, and `CompileResult` is unresolved. The design must say whether the convergence subset is retired, kept only for post-compile domain checks, or rewritten as a typed projection from canonical `internal/formula` output. Leaving a raw subset decoder in place preserves the drift class the design is meant to eliminate.
- [Major] The convergence wire for host capabilities is unspecified. The design names `internal/formula` as the owner of `[requires]` and compiler requirements, but it does not define whether convergence extends `molecule.Options`, passes explicit `HostCapabilities`, precompiles via `CompileWithResult`, or hands a `CompileResult` into instantiation. Without one chosen API, disabled `daemon.formula_v2` behavior cannot be implemented or tested reliably.
- [Major] Convergence root provenance and persisted metadata are ambiguous. Reviews agree the design must state whether convergence roots carry canonical `gc.formula_compiler_*` requirement/provenance metadata, an immutable compile artifact, or only child wisp metadata. This matters because retries and future iterations are driven from convergence root state.
- [Major] Typed diagnostics through convergence surfaces are under-specified. The design promises shared diagnostic codes and fields, but it does not enumerate which convergence events or wrappers carry `formula.compiler_requirement_*` diagnostics for create, retry, speculative wisp, and iteration failures.
- [Major] Mid-loop host-capability disable behavior is not settled. A convergence root can pour new wisps over time, so the design must define what happens when `[daemon] formula_v2` changes from enabled to disabled while a graph-v2 convergence loop remains active.
- [Major] The required test matrix is missing convergence-specific coverage. Both reviews call for tests that prove disabled host capability produces the shared diagnostic and zero durable writes across create, retry, speculative-wisp, and iteration paths, plus coverage for v2-only constructs without `[requires]`.
- [Minor] Public docs and acceptance criteria need an explicit formula-header ownership rule: `[requires]` is a top-level formula requirement owned by `internal/formula`, not a convergence-only key and not something a convergence subset parser may silently drop.

**Disagreements:**
- Claude verdict is `approve-with-risks`; Codex verdict is `block`. Assessment: the synthesized persona verdict is `block` because the zero-durable-write requirement is a hard invariant, and the design does not yet specify an implementation boundary that can guarantee it.
- Claude presents several possible acceptable wires for `molecule.Cook`/`CookOn` and `CompileResult`; Codex requires the design to make late `PourWisp` validation invalid. Assessment: preserving a convenient cook path is fine only after the design names the preflight ordering and typed handoff that prevents root creation before capability failure.
- Claude emphasizes diagnostic event projection and mid-loop policy; Codex emphasizes pre-root creation ordering and provenance. Assessment: these are complementary gaps. The design needs both the preflight contract and the runtime failure semantics for later convergence iterations.
- Claude notes `ValidateForConvergence` may currently be dead on the create path; Codex treats the unresolved migration as a live drift risk. Assessment: the design should close the ambiguity either way, by deleting it from the live contract or constraining it to post-compile domain validation.

**Missing evidence:**
- No concrete convergence create/retry sequence or function signature proving `CompileWithResult` runs before `CreateConvergenceBead`, metadata writes, or `PourWisp`.
- No stated API for threading `HostCapabilities` or an accepted `CompileResult` from the controller/runtime layer into convergence and molecule creation.
- No migration rule for `internal/convergence/formula.go` and `ValidateForConvergence`.
- No explicit metadata/provenance contract for `type=convergence` roots that launch graph-v2 or compiler-v2 formulas.
- No enumeration of convergence event types or payload registry entries that carry shared `Diagnostic` payloads.
- No policy for active convergence roots when host capability flips off before a subsequent wisp pour.
- No required tests instrumenting convergence storage to prove zero root, child bead, metadata, or partial state writes on `formula.compiler_requirement_unsatisfied`.
- No docs acceptance criterion preventing convergence subset documentation from implying a separate TOML schema that can ignore `[requires]`.

**Required changes:**
- Add a convergence-specific preflight contract with an ordering guarantee. It must name the actual entry points, including `cmd/gc/convergence_store.go` `pourWisp`/`PourSpeculativeWisp`, retry paths, and any create handlers, and it must require canonical formula compilation before any durable convergence write.
- Define the chosen API boundary for compiler requirements in convergence: extend `molecule.Options`, pass explicit `HostCapabilities`, require precompilation via `CompileWithResult`, or pass `*CompileResult` into instantiation. The design must make validation only inside a late `PourWisp` call invalid for create/retry paths.
- Specify how convergence-specific validation composes with `CompileResult`: canonical formula parsing and requirement validation first, convergence-only domain checks second, with no raw `[requires]` interpretation in `internal/convergence`.
- State whether convergence roots persist canonical `gc.formula_compiler_*` metadata, an immutable compile artifact, both, or neither, and how retries/status/debugging consume that data without reinterpreting raw formula headers.
- Enumerate convergence event surfaces that wrap shared diagnostics for create, retry, speculative-wisp, and iteration failures, and register the payloads needed for `formula.compiler_requirement_*`.
- Pick the mid-loop policy for `[daemon] formula_v2` changing from enabled to disabled while a convergence root is active: terminate, park, or fail subsequent iterations with typed diagnostics.
- Expand required tests to cover disabled host capability with zero durable writes for convergence create, retry, speculative wisp, and next-iteration pour; enabled-host success for the same canonical path; v2-only constructs without `[requires]`; speculative-wisp conflict diagnostics; and inherited requirements for the controller-injected `evaluate` step.
- Update convergence subset documentation acceptance criteria so `[requires]` is described as a canonical formula requirement owned by `internal/formula`, while convergence fields such as `convergence`, `required_vars`, and `evaluate_prompt` remain domain metadata only.
