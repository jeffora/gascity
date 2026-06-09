# Felix Berger

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- **Major:** The migration boundary for convergence durable writes is still too ambiguous. Both reviews require accepted-artifact preflight, host-capability satisfaction, projection validation, and artifact stamping to happen before any convergence root, retry, speculative/fallback wisp, hook, dependency, convoy, metadata, or artifact-ref write.
- **Major:** The design does not pin the retirement/quarantine of `internal/convergence/formula.go`, `ValidateForConvergence`, and the legacy `convergence.Formula` DTO to a concrete phase with an enforceable static guard. Leaving a temporary exported delegator or test-only parser path keeps the old subset parser culturally and mechanically available.
- **Major:** The convergence authoring schema is not locked. The design must decide whether current top-level `convergence`, `required_vars`, and `evaluate_prompt` fields are canonical compiler-projected fields, deprecated aliases, or replaced by a namespaced table, then fixture the raw shape, unknown-key behavior, source positions, expansion/aspect rules, and docs updates.
- **Major:** Evaluate-prompt handling is under-modeled. The live system has CLI/operator overrides, formula-authored prompt paths, retry metadata copy, prompt readability, path escape, symlink, and content/hash identity concerns; the design needs an explicit runtime input/projection identity model and fixtures for each surface.
- **Major:** Required var validation must remain compiler-owned and preserve the current identifier-shaped rejection set. Non-identifier keys such as empty, leading-digit, hyphen-containing, and non-ASCII names should produce a formula-layer diagnostic before convergence-domain validation runs.
- **Major:** The zero-write requirement needs named fixtures for disabled-host or unsatisfied `[requires]` convergence create/retry/speculative/fallback paths, including a v2-requiring convergence formula on `formula_v2=false`, with no convergence roots or secondary state writes.
- **Minor:** Diagnostic ordering stops too early. It must include accepted-artifact validation, projection, and convergence-domain validation after formula parse/acceptance and host-capability diagnostics so convergence errors cannot mask canonical compiler or host faults.
- **Minor:** Caller-path coverage must go beyond field projection parity. Each convergence durable caller should link to both projection field rows and zero-write fixtures so no path reconstructs fields from raw TOML or metadata.

**Disagreements:**
- Claude treats the legacy convergence subset parser as already having no production callers and therefore an end-state/quarantine risk; Codex treats it as a possible live bypass unless every writer is fenced. My assessment: the design should say both explicitly. Retire or quarantine the legacy parser with deny tests, while making the live migration target the actual create/retry/store call graph.
- Claude suggests the natural retirement point may be the end of sub-phase 4f, potentially behind a legacy build tag or `internal/convergence/legacy/`; Codex pushes for a stronger schema and API decision before implementation proceeds. My assessment: a phase is acceptable only if the design also forbids new production use immediately and makes the static guard executable before new convergence callers are added.
- Codex emphasizes the live `CreateHandler`/`CreateConvergenceBead` ordering risk: placing preflight only inside `PourWisp` would still leave a convergence root and metadata behind. Claude focuses more on `pourWisp -> molecule.Cook -> internal/formula` and the accepted-artifact handoff. My assessment: this is a real blocker; the design must name the pre-root boundary and signatures, not only the lower writer helper.
- Codex calls for modeling operator evaluate-prompt overrides as artifact identity. Claude calls out current convergence-only prompt validation and the no-city-context test behavior. These are additive: the projection contract must cover both identity/source attribution and retained or deliberately retired validation behavior.

**Missing evidence:**
- A before/after call graph for `gc converge create`, controller create handling, `CreateHandler`, `RetryHandler`, `Store.PourWisp`, `PourSpeculativeWisp`, fallback pour, pending adoption, burn, missing-child repair, and manual iterate.
- Concrete API/signature changes for `CreateParams`, `CreateHandler`, `RetryHandler`, `convergence.Store`, `convergenceStoreAdapter`, `cmd/gc/convergence_tick.go`, and/or `molecule.Cook` showing where accepted artifacts, host capabilities, search paths, prompt readers, runtime inputs, and diagnostics flow before root creation.
- Named zero-write fixtures for disabled-host or unsatisfied `[requires]` create/retry/speculative/fallback paths, including event subject keys and the absence of convergence root, metadata, vars, active-index, hook, dependency, convoy, child, retry, fanout, and artifact-ref writes.
- Named fixtures for persisted same-identity artifact reuse after `formula_v2` downgrades, and failure when the formula identity changes under that downgrade.
- A convergence raw-shape matrix for `convergence`, `required_vars`, `evaluate_prompt`, prompt path/content variants, unknown convergence fields, duplicate fields, source positions, inheritance/expansion/aspects, and docs stale-guidance.
- A projection parity fixture that covers current `ValidateForConvergence` behavior, including required var identifier validation and evaluate-prompt path/readability behavior.
- A stated policy for whether `internal/convergence/formula.Formula` is deleted or retained only as an unexported preview DTO, and when `internal/convergence/formula_test.go` is removed or converted to projection-driven tests.

**Required changes:**
- Bind convergence preflight to the pre-root create/retry boundary, not only to `PourWisp`. Name the helper/API and durable-writer allowlist that all convergence create, retry, next-iteration, speculative/fallback wisp, adoption/repair, and burn paths must pass through before writing durable state.
- Specify the signature/interface changes needed to carry `AcceptedCompileArtifact`, host capabilities, search paths, prompt reading, runtime inputs, diagnostics, and artifact stamping through the live integration point.
- Pin the retirement or quarantine of `internal/convergence/formula.go`, `ValidateForConvergence`, and `convergence.Formula` to a specific sub-phase, and add static guards that forbid raw TOML decode, parser imports, root-metadata reconstruction, and direct `convergence.Formula{}` construction in production convergence code.
- Decide and fixture-lock the convergence authoring schema under `internal/formula`, including top-level key policy and the invariant that no convergence-only TOML key may exist outside the canonical compiler recognition set.
- Introduce an explicit evaluate-prompt runtime input/projection identity model. Cover default prompt, formula-authored path, CLI override, retry metadata copy, changed override after root creation, unreadable prompt, path escape, symlink, same-identity reuse, and changed-identity recompile.
- Add compiler-owned required var identifier validation fixtures for empty, leading-digit, hyphen-containing, and non-ASCII keys, with a formula-layer diagnostic emitted before convergence-domain validation.
- Extend the deterministic diagnostic-order table to include accepted-artifact validation, typed projection, and convergence-domain projection validation after host-capability and `[requires]` diagnostics.
- Add the missing zero-write and persisted-artifact reuse fixtures to the convergence fixture matrices, routing them through both `cmd/gc/convergence_store.go:pourWisp` and the `internal/convergence/create.go` handler path.
