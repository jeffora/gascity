# Design Review Synthesis

## Overall Verdict: block

Worst-verdict-wins yields `block`: six persona syntheses blocked and the remaining four approved only with risks. The design is directionally right to move compiler capability declarations into formula `[requires]` while keeping the active Gas City binary as the authority, but it still lacks enforceable contracts for host capability satisfaction, caller migration, diagnostics, provenance, and release gates.

## Consensus Strengths
- Multiple personas praised the central boundary: formulas declare requirements; the active Gas City binary decides whether those requirements can run.
- Reviewers agreed that raw `contract = "graph.v2"` checks should converge on one canonical `internal/formula` compile or preflight path.
- The compatibility bridge for legacy `contract = "graph.v2"` formulas is necessary; no reviewer recommended an abrupt break for existing packs.
- Structured diagnostics, normalized requirements, provenance, and stale-guidance checks are the right primitives if they become concrete contracts instead of prose.
- The design is right to avoid formula-level artifact semver and keep reproducibility anchored in pack revision, source ref, locked SHA, or local content identity.

## Critical Findings

### [Blocker] Host Capability And Compile Preflight Are Not Enforceable
**Sources:** Nadia Sorenson/Claude+Codex; Yuki Patel/Claude+Codex; Felix Berger/Claude+Codex; Ibrahim Park/Claude+Codex; Marta Hidalgo/Claude+Codex
**Issue:** The design does not yet define one typed host-capability input and one formula-domain satisfaction contract that every durable write path must use. `HostCapabilities` can still represent contradictory sources of truth, callers can construct `NormalizedRequirements` without compiling, and runtime paths such as dispatch fanout fragments and convergence can still create beads or state before formula capability failure is known.
**Required change:** Define a single `internal/formula` API boundary, such as `CompileWithResult` plus typed host capabilities, and require production root, wisp, molecule, order, retry, fanout, and convergence paths to consume the resulting `CompileResult`. Remove or edge-confine contradictory `FormulaV2` state, make caller-constructed requirements impossible outside tests, and add no-partial-write tests for disabled host capability across all durable entry points.

### [Blocker] Parser, Requirement Grammar, And V2 Construct Registry Are Not Mechanical
**Sources:** Priya Zielinski/Claude+Codex; Nadia Sorenson/Claude+Codex; Elias Vega/Claude+Codex; Ibrahim Park/Claude+Codex
**Issue:** The v2-only construct registry still mixes TOML locations, metadata keys, and metadata values in ways an implementer cannot encode or test exhaustively. The accepted `requires.formula_compiler` grammar, empty/default semantics, unknown axis handling, contract/requires truth table, inheritance/aggregation behavior, raw TOML preservation, and diagnostic ordering are not pinned tightly enough.
**Required change:** Replace the checklist with normative tables: accepted strings, TOML shape matrix, legacy `contract` compatibility matrix, construct-presence registry, inheritance/aggregation rules, host-capability outcomes, and combined-defect precedence. Preserve raw keys/types/source positions before typed decoding, and require table-driven fixtures for positive and negative cases.

### [Blocker] Diagnostic Projection Can Diverge Across Operator Surfaces
**Sources:** Marta Hidalgo/Claude+Codex; Avery Brooks/Claude+Codex; Felix Berger/Claude+Codex; Yuki Patel/Claude+Codex
**Issue:** The same formula diagnostic can still surface differently through direct CLI, API-routed CLI, Huma endpoints, generated clients, dashboard cards, controller/order events, convergence wrappers, and SSE/event streams. Event projection is ambiguous, repeated warning/fatal cadence is not specified, and generated dashboard/API consumers are not structurally prevented from inferring state from raw metadata strings.
**Required change:** Add a cross-surface diagnostic matrix by code. For each code, define severity, canonical message, remediation, CLI exit behavior, HTTP status/body, generated-client rendering, dashboard state, event constants, registered typed payloads, envelope fields, source attribution, and deduplication key. Add parity tests for direct versus API-routed CLI and repeated order/controller failures.

### [Blocker] Migration Gates And Compatibility Policy Are Not Measurable
**Sources:** Elias Vega/Claude+Codex; Lena Driscoll/Claude+Codex; Saoirse Raman/Claude+Codex; Yuki Patel/Claude+Codex
**Issue:** The alias window, dual declarations, `gc.formula_contract` dual-stamping, minimum binary floor, external SHA-pinned pack support, bd/native fallback behavior, and `gc formula validate --all-packs --legacy-contract-report` gate are not operational artifacts. The design cannot prove when it is safe to update first-party formulas, remove aliases, or require `[requires]` only.
**Required change:** Replace migration-window prose with a compatibility matrix and release gates covering old/new binaries, old/new packs, external pinned packs, bd fallback/native paths, dual-declared formulas, requires-only formulas, and mixed-version shared stores. Specify the legacy-contract report schema and exit codes, the external compatibility plan, minimum-floor artifact, dual-stamp retirement criteria, and rollback procedure.

### [Blocker] Pack Provenance And Workflow-Root Metadata Are Not Auditable
**Sources:** Saoirse Raman/Claude+Codex; Yuki Patel/Claude+Codex; Nadia Sorenson/Claude+Codex; Lena Driscoll/Claude+Codex
**Issue:** The design depends on pack revision as the reproducibility boundary but does not define a durable provenance data model from formula layer resolution through compilation and workflow-root creation. Path-oriented layer winners and staged symlinks can lose pack binding, import source/ref, locked revision, content hash, dirty state, transitive import attribution, and layer priority. Store queries can also miss canonical-only roots if they filter only on legacy metadata.
**Required change:** Define structured formula-source provenance in `CompileResult` and either persist it on workflow roots or link a durable compile artifact from each root. Add canonical workflow-root query/predicate semantics that cover legacy-only, dual-stamped, graph.v2-only, and requires-only roots, and guard new raw metadata query construction outside the owner package.

### [Blocker] Convergence And Fanout Remain Active Bypass Paths
**Sources:** Felix Berger/Claude+Codex; Yuki Patel/Claude+Codex; Nadia Sorenson/Claude+Codex
**Issue:** Convergence create, retry, speculative-wisp, and next-iteration paths need canonical capability validation before any durable convergence root, child bead, metadata, or partial state is written. Dispatch fanout expansion fragments also compile at runtime after the root is created and are missing from the executable migration inventory.
**Required change:** Add explicit migration rows and function signatures for convergence and `internal/dispatch/fanout.go`. State whether convergence deletes its subset parser, keeps it only for post-compile domain checks, or rewrites it as a typed projection from canonical formula output. Require disabled-host tests proving zero durable writes for convergence and fanout fragment paths.

### [Major] Documentation, Generated Help, And Terminology Are Not Release-Gated
**Sources:** Avery Brooks/Claude+Codex; Saoirse Raman/Claude+Codex; Priya Zielinski/Claude+Codex
**Issue:** User-visible diagnostics can ship before reference docs, generated config/CLI help, tutorials, examples, and first-party snippets teach `[requires]` as the canonical authoring surface. The design also leaves terminology collisions unresolved between formula `[requires]`, pack `requires_gc`, legacy pack `requires`, import `version`, formula `version`, compiler capability, host capability, and pack schema.
**Required change:** Gate the first user-visible diagnostics on a file-level docs/examples/generated-help inventory. Add canonical modern and dual-declared TOML examples, a "which requirement surface do I use" comparison, stale-guidance CI with explicit exceptions, and regeneration commands for generated docs/help.

### [Major] Forward Capability Semantics Are Too Implicit
**Sources:** Ibrahim Park/Claude+Codex; Priya Zielinski/Claude+Codex; Saoirse Raman/Claude+Codex
**Issue:** The public-looking `formula_compiler = ">=N"` syntax does not yet say whether capabilities are monotonic minimums or a closed enum-like surface. Future axes are fail-closed in spirit, but the design does not choose a permanent unknown-axis policy, extension namespace policy, per-axis provenance model, or diagnostic split between invalid syntax and unsupported future capability.
**Required change:** Add a forward-compatibility section that defines compiler capability monotonicity, numeric stability, zero-value behavior, unknown-axis policy, diagnostic conventions, and the rule that every future axis must add typed normalized state, provenance, diagnostics, docs, and tests before shipping.

### [Minor] Transitional Wording And Naming Need Cleanup
**Sources:** Nadia Sorenson/Claude+Codex; Marta Hidalgo/Claude+Codex; Ibrahim Park/Claude+Codex; Avery Brooks/Claude+Codex
**Issue:** Several names and messages still invite drift: `CompilerCapabilityGraphV2` is implementation-shaped, warning text says "once per invocation" when validate-all may emit one warning per source/key, and local non-reproducible markers lack a clear operational consequence.
**Required change:** Rename capability constants to revision-level names such as `CompilerCapabilityV2`, clarify warning cadence language, and define whether non-reproducible local inputs produce CI failures, dashboard warnings, launch policy effects, or report-only metadata.

## Disagreements
- Several personas had internal Claude/Codex verdict disagreements: Claude often returned `approve-with-risks` while Codex returned `block`. My assessment follows worst-verdict-wins and treats the block verdicts as justified because the gaps affect executable contracts, durable state ordering, and operator surfaces.
- Reviewers proposed different shapes for the canonical formula API: compile options, precompiled `CompileResult` handoff, requirement satisfaction helpers, or accessor predicates. The precise shape is less important than making it typed, owned by `internal/formula`, and mandatory for all behavioral consumers.
- External compatibility options differed: maintain mainline alias support longer, provide a compatibility branch/LTS binary, or fail upgrades closed after validation. Any option can work only if the design chooses one and defines owner, duration, opt-in, release artifact, and retirement criteria.
- Convergence could delete its local formula subset, keep it for post-compile domain checks, or rewrite it as a projection from canonical parsing. The unresolved ownership is the problem; the design must choose one.
- Reviewers varied on exact HTTP statuses, event names, and warning suppression mechanics. The design need not adopt a specific reviewer suggestion, but it must provide stable per-code classifications and registered typed event payloads.

## Missing Evidence
- Final `HostCapabilities`, `NormalizedRequirements`, `CompileResult`, and requirement-satisfaction API shapes.
- Checked-in call-site inventory for raw `contract`, `graph.v2`, `Requires.FormulaCompiler`, `gc.formula_contract`, workflow-root predicates, `CompileExpansionFragment`, and convergence formula paths.
- Encodable v2-only construct registry with TOML paths, trigger predicates, scanned locations, positive fixtures, and negative fixtures.
- Full `contract` x `[requires]` truth table, including invalid contract values, explicit `>=1`, unsupported future values, unknown axes, empty tables, and diagnostic precedence.
- Raw TOML parsing strategy that preserves keys, value types, and source positions for structured diagnostics.
- Cross-surface diagnostic projection matrix, event constants, registered payload structs, generated-client behavior, and `OnceKey` suppression semantics.
- `gc formula validate --all-packs --legacy-contract-report` and `--provenance` JSON schemas, scan scopes, imported-pack behavior, exit codes, and release-gate semantics.
- External pack-author validation flow for local directories, remote refs, SHA-pinned packs, and transitive imports.
- Minimum binary floor artifact, mixed-version shared-store matrix, bd/native fallback decision, and dual-stamp retirement gate.
- Structured formula provenance model from layer resolution to workflow-root metadata or durable compile artifact.
- Convergence and fanout tests proving disabled host capability fails before root, child bead, metadata, or partial state writes.
- File-level documentation, example, generated-help, and stale-guidance CI inventory.

## Recommended Changes
1. Define the typed host-capability input and single `internal/formula` requirement-satisfaction contract before migrating callers.
2. Build the parser/validation matrices: accepted grammar, raw TOML preservation, `contract`/`[requires]` precedence, v2 construct registry, inheritance/aggregation, host capability, and diagnostic ordering.
3. Add an executable caller migration table and canonical workflow-root query/predicate API, then migrate one caller family at a time with parity tests.
4. Lock down convergence and fanout preflight so disabled host capability fails before any durable state is created.
5. Define the diagnostic projection matrix, typed event payloads, Huma response fields, dashboard-facing state, generated-client behavior, and warning/fatal deduplication semantics.
6. Specify bd/native compatibility, external pack support, old/new binary behavior, and minimum binary floor before changing first-party formula sources.
7. Replace the migration window with executable release gates: legacy-contract report, provenance report, external-pack validation, dual-declaration lint, and dual-stamp retirement.
8. Persist or durably link formula provenance for new workflow roots, including pack source/ref/revision, local hash or dirty status, transitive attribution, and reproducibility state.
9. Gate user-visible diagnostics on updated docs, generated help, examples, first-party formula snippets, tutorials, and stale-guidance CI.
10. Add forward-compatibility rules for future `[requires]` axes so every new axis is typed, documented, tested, and fails closed on older binaries.
