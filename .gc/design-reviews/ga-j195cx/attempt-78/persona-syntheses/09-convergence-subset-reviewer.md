# Felix Berger

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] The destination direction is sound: convergence should retire its subset parser and consume accepted compiler artifacts plus typed convergence projection validation. Both reviews agree this eliminates the core `[requires]` drift risk if the design preserves that boundary through implementation.
- [Major] The closed convergence diagnostic surface is incomplete as written. The design lists missing and unreadable evaluate-prompt cases, but current convergence validation also checks prompt content markers such as `bd meta set` and `convergence.agent_verdict`. That rule needs an explicit diagnostic, projection owner, runtime owner, or deletion notice before the "emits only these" diagnostic list can be trusted.
- [Major] `ConvergenceRuntimeInputs` is too weak or underspecified for required-var validation. A plain `Vars map[string]string` cannot preserve duplicate inputs, defaulted values, caller/source attribution, redaction facts, or satisfaction evidence, yet the design's fixtures require those distinctions before convergence writes.
- [Major] The artifact-reference contract is ambiguous. The design references a generic `gc.formula_compile_artifact` repair path while later requiring `gc.convergence_formula_compile_artifact` for convergence durable behavior, and `convergence.formula_artifact_conflict` appears outside the closed `ValidateProjection` diagnostic list. Readers need one canonical precedence, migration, conflict, and repair rule.
- [Major] Compiler-owned subsystem projections need an explicit closure rule. `CompiledConvergenceProjection` is the right ownership move for this design, but without a bound or graduation criterion it creates a precedent for accumulating subsystem-specific projections in `internal/formula`.
- [Minor] The 4b-4e migration fence is asserted but not fixture-locked. While convergence still uses the legacy parser, a v2-requiring formula could bypass `[requires]` unless every convergence create/retry/iteration path shadow-compiles or is otherwise fenced before writing.
- [Minor] Host-capability lifetime is correct in pieces but not stated normatively in one place. The design should say which snapshot controls create, same-identity retry, new-identity retry, missing-child repair, and speculative wisps, and ensure any `CurrentHost` access is diagnostic-only rather than a second authorization gate.

**Disagreements:**
- There is no verdict disagreement: Claude and Codex both returned `approve-with-risks`.
- Claude prioritized evaluate-prompt content invariants, transition-window safety, host-capability lifetime, and the projection-precedent problem. Codex prioritized runtime-var evidence, artifact-ref ambiguity, artifact-conflict diagnostics, and `CurrentHost` branching risk. These are complementary rather than contradictory.
- Codex treats `convergence.formula_artifact_conflict` as a diagnostic-phase placement issue, while the broader artifact-key ambiguity is a larger durable-state contract issue. This synthesis treats the durable-state ambiguity as major and the missing diagnostic registration as part of that required fix.
- Claude frames the host-capability concern as scattered normative prose; Codex frames it as `CurrentHost` access conflicting with the no-branching guard. Assessment: the fix should do both, by centralizing the lifetime rule and fixture-locking diagnostic-only reads.

**Missing evidence:**
- A mapping for the current evaluate-prompt content-marker checks to a new diagnostic code, projection field, runtime invariant, or explicit retirement decision.
- A `formula_projection_equivalence.yaml` or equivalent fixture row proving evaluate-prompt content rules are preserved or intentionally moved outside projection validation.
- The concrete `ConvergenceRuntimeInputs` type and `ValidateProjection` signature, including supplied values, defaults, duplicate handling, value source, redaction, vars hash, and satisfaction facts.
- A single artifact-reference table for convergence roots covering `gc.formula_compile_artifact`, `gc.convergence_formula_compile_artifact`, dual-stamp migration, conflict precedence, repair-root behavior, and zero-write diagnostics.
- A checked transition fixture proving convergence cannot accept a formula requiring compiler v2 on a host where `[daemon] formula_v2=false` during sub-phases 4b-4e.
- A normative host-capability lifetime paragraph covering create, same-identity retry, changed-identity retry, missing-child repair, speculative wisp, host downgrade, and config-generation changes.
- A guard or fixture proving production convergence code cannot read raw `evaluate_prompt`, `required_vars`, `convergence`, `[requires]`, or host-capability fields outside the accepted projection path during the migration.
- A closure rule for when a subsystem gets a compiler-owned typed projection versus consuming generic compiled step or metadata structures.

**Required changes:**
- Resolve evaluate-prompt content validation explicitly: add a convergence diagnostic, project a structured validation fact, move the rule to a named runtime owner, or document its deliberate removal. The closed diagnostic list must remain truthful after the change.
- Define `ConvergenceRuntimeInputs` next to `ValidateProjection` with typed runtime-var satisfaction evidence. Do not flatten declarations, defaults, duplicate diagnostics, or redaction facts into `map[string]string`.
- Specify one convergence artifact-ref contract, including canonical key, accepted aliases if any, migration stamping, read precedence, conflict diagnostics, repair behavior, and zero-write fixture coverage.
- Add `convergence.formula_artifact_conflict` to the owning diagnostic phase or move conflict detection to a separately documented pre-projection phase with payload and fixture coverage.
- Add a 4b-4e transition fixture or change the migration order so convergence cannot write from the legacy parser when a shadow compile or accepted artifact would reject the formula.
- Extend the convergence static guard to block raw reads of evaluate prompt, required vars, convergence metadata, `[requires]`, and host capabilities outside the projection path during the transition as well as after 4f.
- Centralize host-capability lifetime rules and make diagnostic-only `CurrentHost` access enforceable by tests or guards.
- Add a projection closure rule that bounds allowed compiler-owned subsystem projections or defines the criteria for adding one.
