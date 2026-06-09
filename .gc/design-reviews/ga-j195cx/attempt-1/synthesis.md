# Design Review Synthesis

## Overall Verdict: block

Worst-verdict-wins yields `block`: six persona syntheses block, and the remaining four approve only with risks. The design has the right direction, especially moving formula capability declarations into a canonical `[requires]` surface, but it still lacks the typed host-capability contract, caller migration proof, diagnostic projection rules, and release gates needed before implementation can proceed safely.

## Consensus Strengths
- Multiple personas praised the core boundary: formulas should declare requirements, while the active Gas City binary decides whether it can satisfy them.
- Reviewers consistently agreed that raw `contract = "graph.v2"` checks should be replaced by one canonical `internal/formula` compile or preflight path.
- The design is directionally right to reject formula-level artifact semver and keep reproducibility anchored at pack revision, ref, SHA, or equivalent locked provenance.
- The compatibility bridge for legacy `contract = "graph.v2"` formulas is necessary and recognized; no reviewer asked for an abrupt break.
- The proposed structured diagnostic codebook and provenance concepts are good foundations if they become concrete contracts rather than prose guidance.

## Critical Findings

### [Blocker] Host Capability Satisfaction Contract Is Missing
**Sources:** Nadia Sorenson/Claude+Codex, Ibrahim Park/Claude+Codex, Felix Berger/Claude+Codex, Yuki Patel/Claude+Codex
**Issue:** The design defines normalized requirement outputs but not the typed host-capability input, `CheckRequirements` function, or equivalent formula-domain contract that all callers must use. That leaves room for package-global formula-v2 state, caller-side `[daemon] formula_v2` checks, or separate convergence/API/order interpretations.
**Required change:** Add a "Host capability and requirement satisfaction" section that defines the typed host capability input, the formula-domain satisfaction function or compile option, and the invariant that all behavioral callers consume `CompileResult` or a helper backed by it. Add tests proving two compiles in one process can evaluate different host capabilities deterministically.

### [Blocker] Caller And Convergence Migration Is Not Executable
**Sources:** Yuki Patel/Claude+Codex, Nadia Sorenson/Claude+Codex, Felix Berger/Claude+Codex
**Issue:** The caller inventory is still category-level. Raw `Contract`, `Requires.FormulaCompiler`, `gc.formula_contract`, duplicated workflow-root predicates, graph workflow metadata, and convergence-local validation can remain live in sling, sourceworkflow, API projections, graphroute, molecule stamping, order dispatch, formula endpoints, and convergence.
**Required change:** Add a call-site migration table naming current helpers/files, target `internal/formula` APIs, owner packages, and tests. Define the canonical persisted workflow-root predicate and root-listing/query helpers. Require convergence create/retry/speculative-wisp paths to preflight through canonical formula semantics before any root, child bead, or partial convergence state is written.

### [Blocker] Operator Diagnostic Projection Is Underspecified
**Sources:** Marta Hidalgo/Claude+Codex, Yuki Patel/Claude+Codex, Avery Brooks/Claude+Codex, Felix Berger/Claude+Codex
**Issue:** `formula.compiler_requirement_unsatisfied` and deprecation diagnostics do not have stable transport behavior. CLI exit/stderr, Huma JSON/status, dashboard state, controller/order events, convergence wrapping, message/remediation text, and repeated-warning suppression can diverge or flood operators.
**Required change:** Define a diagnostic projection matrix by code: severity, canonical message, remediation, CLI behavior, HTTP status/body, dashboard state, event projection, typed event payload/registration, source attribution, and suppression key. Specify `OnceKey` storage owner, key shape, lifetime, reset behavior, bounded memory behavior, and fatal-event cadence.

### [Blocker] Migration And Compatibility Gates Are Not Measurable
**Sources:** Elias Vega/Claude+Codex, Lena Driscoll/Claude+Codex, Saoirse Raman/Claude+Codex, Yuki Patel/Claude+Codex
**Issue:** The alias window, minimum binary floor, external SHA-pinned pack support, bd/native parity, and `gc formula validate --all-packs --legacy-contract-report` gate are not operational artifacts. The proposed raw-`contract` guard also conflicts with the required dual-declaration bridge unless the design separates Go-consumer guards from formula-source migration lint.
**Required change:** Replace migration-window prose with a compatibility matrix and release gates covering old/new binaries, old/new packs, external pinned packs, bd fallback/native paths, dual-declared formulas, requires-only formulas, and mixed-version shared stores. Specify the legacy-contract report command contract, support owner/duration for external packs, minimum-floor artifact, and dual-stamp retirement criteria.

### [Blocker] Parser Grammar And Validation Matrix Are Incomplete
**Sources:** Priya Zielinski/Claude+Codex, Elias Vega/Claude+Codex, Ibrahim Park/Claude+Codex
**Issue:** The design does not fully pin the accepted `requires.formula_compiler` grammar, TOML type behavior, empty/default semantics, unknown-axis handling, contract/requires truth table, diagnostic precedence, inheritance/aggregation behavior, or v2-only construct registry. Typed TOML decoding may also erase the raw data needed to emit structured diagnostics.
**Required change:** Specify byte-exact accepted expressions, preferably `{">=1", ">=2"}` for v0, and reject empty strings, unsupported comparators, wrong types, dotted/nested misuse, and unknown axes with typed diagnostics. Add a table-driven validation matrix covering contract state, requires state, source location, v2 construct presence, host capability, inheritance/expansion/aspect aggregation, attempted lowering, and combined-defect precedence.

### [Major] Persisted Metadata And Provenance Are Not Auditable Enough
**Sources:** Saoirse Raman/Claude+Codex, Yuki Patel/Claude+Codex, Nadia Sorenson/Claude+Codex, Lena Driscoll/Claude+Codex
**Issue:** The design does not say exactly what new workflow roots stamp, when `gc.formula_contract` stops being produced, how legacy readers remain compatible, or how an operator reconstructs the formula source, pack binding, ref/revision, pin kind, and local/dirty state used months earlier.
**Required change:** Define the persisted metadata migration contract and provenance surface. Either persist winning formula provenance on each new root or link a durable compile artifact from the root. Include formula identity/source path, pack binding/source/ref, locked revision or local hash/dirty marker, and reproducibility status.

### [Major] Docs, Terminology, And Legacy `version` Behavior Remain Open
**Sources:** Avery Brooks/Claude+Codex, Saoirse Raman/Claude+Codex, Priya Zielinski/Claude+Codex
**Issue:** `requires.formula_compiler`, formula-level `[requires]`, pack-level `requires`, `pack.requires_gc`, compiler capability, host capability, legacy `contract`, PackV2 `schema`, graph workflow, and legacy formula `version` are not tied together in one terminology contract. The design can emit deprecation diagnostics before docs, examples, generated help, and tutorials teach the replacement.
**Required change:** Add a glossary and a file-by-file docs/examples inventory. Resolve `formula.version_deprecated` deterministically: either define its compile/validate/projection/suppression contract or preserve `version` silently and remove the diagnostic from the required codebook. Gate user-visible diagnostics on updated reference docs, generated help, public examples, and stale-guidance checks.

### [Major] Forward Compatibility Strategy Is Too Implicit
**Sources:** Ibrahim Park/Claude+Codex, Priya Zielinski/Claude+Codex, Saoirse Raman/Claude+Codex
**Issue:** `[requires]` is presented as future multi-axis infrastructure, but v0 is still compiler-only and scalar. The design does not choose whether unknown axes are permanently hard errors, whether any extension namespace exists, how future axes get typed normalized state, or how old binaries distinguish "upgrade required" from "misspelled key."
**Required change:** Add a forward-compatibility section requiring every future axis to add typed normalized state, validation rules, diagnostics/remediation, provenance, docs, and tests. State that raw maps, raw TOML pass-through, and ignored unknown keys are not allowed. Document numeric compiler capability stability and `CompilerCapability(0)` behavior.

### [Minor] Transitional Source Spelling And Local Inputs Need Cleanup Rules
**Sources:** Marta Hidalgo/Claude+Codex, Elias Vega/Claude+Codex, Saoirse Raman/Claude+Codex
**Issue:** Diagnostics and reports need to preserve whether the author used legacy `contract`, direct `[requires]`, both, or `version`, and local path imports need explicit not-reproducibly-pinned handling.
**Required change:** Preserve source spelling/provenance in diagnostics and reports, define the lifecycle for agreeing dual declarations, and mark local path imports with a content hash/dirty status or an explicit local/not-reproducibly-pinned marker.

## Disagreements
- Verdicts differed inside several personas: Claude often returned `approve-with-risks`, while Codex returned `block`. My assessment follows worst-verdict-wins and treats the block verdicts as justified because the gaps affect executable contracts, not polish.
- Reviewers proposed different shapes for the canonical API: compile options, a requirement satisfaction helper, accessor predicates, or compile-result fields. The shape is less important than making it typed, owned by `internal/formula`, and mandatory for all behavioral consumers.
- External legacy support options differed: keep `contract = "graph.v2"` indefinitely for external pack sources, provide a maintained compatibility branch, or fail upgrades closed after local validation. Any option is acceptable only if the design chooses one and makes it measurable.
- Reviewers varied on the exact HTTP status, CLI exit code, and event names for diagnostics. The design need not adopt a specific reviewer suggestion such as HTTP 412, but it must provide stable per-code classifications and registered typed event payloads.
- Convergence can delete its local formula subset, keep it for post-compile domain checks, or rewrite it as a projection from canonical parsing. The unresolved live-path ownership is the problem.

## Missing Evidence
- Complete call-site inventory for raw `contract`, `graph.v2`, `Requires.FormulaCompiler`, `gc.formula_contract`, duplicated root predicates, and graph-workflow metadata consumers.
- Final host-capability input type and formula-domain requirement satisfaction contract.
- Parser implementation strategy that preserves raw `[requires]` keys and values long enough to emit structured diagnostics.
- Canonical `contract` x `[requires]` truth table, including invalid contract values and dedicated diagnostic codes.
- V2-only construct registry with positive and negative tests for every scanned location.
- Per-code diagnostic projection matrix across CLI, API, dashboard, controller/order events, convergence, and validation commands.
- `OnceKey` suppression semantics for CLI, API, daemon/controller, convergence, and order dispatch.
- bd/native parity statement for `GC_NATIVE_FORMULA=false`, bd-backed stores, and direct `bd cook --persist` consumers.
- Concrete `gc formula validate --all-packs --legacy-contract-report` schema, scope, imported-pack behavior, and exit-code contract.
- Enforceable minimum binary floor artifact and mixed-version shared-store compatibility matrix.
- Persisted workflow-root metadata and provenance contract.
- File-by-file docs/examples/generated-help inventory and stale-guidance CI spec.

## Recommended Changes
1. Define the typed host-capability input and single `internal/formula` requirement-satisfaction contract before any caller migration.
2. Build the exact parser/validation matrix, including accepted grammar, raw TOML preservation, contract/requires precedence, v2 construct registry, and diagnostic ordering.
3. Add a caller migration table and canonical workflow-root predicate/listing API, then migrate one caller family at a time with parity tests.
4. Lock down convergence preflight so disabled host capability fails before any durable convergence state is created.
5. Define the diagnostic projection matrix, typed event payloads, Huma response fields, dashboard-facing state, and `OnceKey` suppression semantics.
6. Specify bd/native compatibility and old/new binary-pack behavior before changing first-party formula sources.
7. Replace the migration window with executable release gates: minimum binary floor, legacy-contract report, external-pack support plan, dual-declaration lint, and dual-stamp retirement.
8. Persist or durably link formula provenance for new workflow roots, including pack source/ref/revision and local reproducibility state.
9. Resolve docs and terminology before diagnostics ship: glossary, formula/config references, generated help, examples, tutorials, fixtures, and stale-guidance CI.
10. Add a forward-compatibility rule for future `[requires]` axes so every new axis is typed, documented, tested, and fails closed on older binaries.
