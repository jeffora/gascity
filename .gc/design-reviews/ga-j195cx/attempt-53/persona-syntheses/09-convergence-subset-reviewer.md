# Felix Berger

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The accepted-artifact contract does not prove convergence can project from persisted compiler output alone. Both reviews identify that convergence fields, runtime vars, evaluate prompt data, retry policy, source attribution, and step facts must survive raw decode and acceptance; otherwise `ProjectAcceptedFormula` must reopen formula source, keep the subset parser alive, or lose data needed for retry, next iteration, and repair.
- [Major] The convergence projection boundary is underspecified. `formula.CompiledConvergenceProjection` and `convergence.ConvergenceMetadata` overlap on enabled state, runtime variables, evaluate prompt fields, retry policy, source attribution, and requirements, but the design does not pin which package owns each field or whether projection wraps a compiler-owned payload or re-derives from steps.
- [Major] The legacy shim and static guard are not mechanical enough. Claude calls out the missing symbol-level contract for `internal/convergence/formula.Formula`, `ValidateForConvergence`, and helpers such as `ResolveEvaluateStep`, `ValidateEvaluatePrompt`, and `ValidateVarKey`; Codex agrees that production convergence must be barred from raw TOML reads, legacy subset imports, and required-var reconstruction outside compiler projection.
- [Major] Convergence artifact reuse identity is ambiguous. The design says retries can reuse persisted artifacts after host downgrade, but does not state which of formula name, content hash, vars hash, options hash, search paths hash, host capabilities, config generation, and provenance binding identity must match for retry, next iteration, or missing-child repair.
- [Major] The zero-write contract is not yet enforceable at the current convergence write boundary. The design names create, retry, next iteration, missing-child repair, and speculative wisp paths, but needs tests proving fatal formula diagnostics leave no root bead, closed placeholder, metadata, child, hook, convoy, wisp, or artifact ref. Codex specifically flags the current root-before-`PourWisp` shape as a migration hazard.
- [Major] Operator diagnostics and event identity need a pinned contract. Claude flags that `convergence.formula_compile_failed` allows failures before a convergence id exists, but the design does not say whether pre-create failures emit an event, how dashboards deduplicate them, or how the event relates to shared formula compiler diagnostics.
- [Minor] The caller coverage lock is asserted instead of derived. The `caller-preflight.count_lock: 18` can pass while omitting a convergence durable-writer path unless it is generated from the caller-path dimension and row-kind matrix.

**Disagreements:**
- Claude returned `approve-with-risks`; Codex returned `block`. Assessment: the reviewers agree on the problem shape, but Codex's accepted-artifact objection blocks the central promise that convergence can stop reading a subset of formula TOML. The persona verdict is therefore `block`.
- Claude describes `internal/convergence/formula.Formula` and `ValidateForConvergence` as possibly retired or rewritten as shims if the contract is exact. Codex pushes harder toward a projection-only API. Assessment: deletion is the default safe path; any temporary shim must be named, one-call-deep, unable to own a convergence-shaped struct, and covered by an expiry-bound guard.
- Claude emphasizes artifact reuse identity, empty-id event behavior, and matrix coverage. Codex emphasizes persisted projection payload and zero-write write ordering. Assessment: these are complementary, not conflicting; they are all manifestations of the same missing boundary between compiler acceptance and convergence mutation.
- Codex's blocker is not stated as a blocker by Claude, but Claude's overlapping-struct and projection data-flow findings point at the same failure mode. Assessment: resolve it by making persisted compiler output sufficient for projection from an existing accepted artifact, not by relying on reviewer interpretation during implementation.

**Missing evidence:**
- No field-level accepted-artifact schema shows persisted compiled steps, runtime vars, retry policy, relevant convergence step, evaluate prompt source/path, `CompiledConvergenceProjection`, or enough equivalent data for `ProjectAcceptedFormula` to run without source access.
- No fixture proves `ProjectAcceptedFormula` works from a persisted artifact when the original formula source is unavailable.
- No named fixture covers a convergence formula that declares `[requires] formula_compiler = ">=2"` without v2-only constructs and proves that voluntary requirement survives projection.
- No convergence-specific concurrent downgrade fixture proves an accepted root can retry from a persisted artifact while a sibling create fails closed under the downgraded host capability.
- No active legacy-root migration fixture proves artifact stamping is idempotent under concurrent retry attempts.
- No explicit artifact-reuse identity table defines required matches for retry, next iteration, and missing-child repair.
- No checked allowlist or symbol/regex matrix defines `TestNoConvergenceSubsetParserUse`.
- No event grouping or suppression rule exists for `convergence.formula_compile_failed` before a convergence id has been minted.
- No derived count lock proves every convergence durable writer has a zero-write caller-preflight row.

**Required changes:**
- Define the accepted-artifact serialization, or a referenced compiler-owned projection snapshot, as versioned and hash-checked data sufficient for `ProjectAcceptedFormula` to populate convergence metadata without reading formula source.
- Add a package-boundary data-flow contract from convergence create/retry/iteration/reconciler/repair entry points through `CompileWithResult`, `AcceptCompileResult`, `ProjectAcceptedFormula`, `ValidateProjection`, and durable writes. State which package owns every field on `CompiledConvergenceProjection` and `ConvergenceMetadata`.
- Replace the vague shim language with an exported-symbol migration table for `internal/convergence/formula.Formula`, `ValidateForConvergence`, `ResolveEvaluateStep`, `ValidateEvaluatePrompt`, `ValidateVarKey`, and any related helpers. Prefer deletion; otherwise constrain any shim to forwarding to compiler projection only.
- Specify `TestNoConvergenceSubsetParserUse` as a checked allowlist, modeled on the raw-consumer allowlist, covering forbidden imports, raw formula file reads, `convergence.Formula{}` construction, raw `[requires]`/`contract` inspection, and required-var parsing outside compiler-owned projection.
- Add the missing convergence fixtures: persisted-artifact projection with missing source, voluntary `[requires]` preservation, retry/create across host downgrade, concurrent legacy-root artifact stamping, and zero-write behavior for every durable convergence path.
- Pin the artifact reuse identity table for retry, next iteration, and missing-child repair, including the treatment of search paths, host capabilities, config generation, and provenance binding identity.
- Define the empty-id contract for `convergence.formula_compile_failed`: either suppress pre-create Event Bus events and rely on the synchronous diagnostic, or emit them with a stable dedup key such as formula, host capability, once key, config generation, and source path.
- Derive `caller-preflight.count_lock` from the caller-path dimension and row-kind matrix so CI detects any missing convergence durable-writer row.
