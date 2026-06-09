# Yuki Patel

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex; Kimi 2.6 not present

**Consensus findings:**
- [Major] Legacy `MolCook` and bd-backed materialization paths are not fully closed in the migration plan. Both reviews identify reachable `Store.MolCook`/`MolCookOn`, `BdStore`, or exec-store script-backed paths as the main caller-integration risk: they can still authorize molecule or wisp writes without going through `CompileWithResult`, `AcceptCompileResult`, and accepted-artifact validation. The design says shell-outs must not authorize durable writes, but it does not yet retire or preflight every still-reachable helper.
- [Major] The caller chokepoint guard does not cover every bypass shape. Claude notes that the design splits ownership between `internal/formula` for compile/write authorization and `internal/sourceworkflow` for workflow-root reads, but the named guard only scans raw formula metadata symbols. Codex adds that dashboard/generated-client inference must also be covered or prevented by a stricter typed API invariant.
- [Major] Transitional behavior for requires-only formulas is under-specified while legacy materialization remains compiled in. Both reviewers flag the migration window: external or pinned packs could reach old bd-backed behavior that ignores `[requires]` unless the plan explicitly rejects, native-preflights, or dual-declaration-gates those formulas before durable writes.
- [Minor] Convergence transition ordering needs one explicit write fence. Claude accepts the proposed shadow-compile fence but finds it does not state that the native shadow compile runs before any convergence root, iteration, missing-child, or speculative-wisp write. Without that ordering, old subset behavior could still persist state before the shared diagnostic is consulted.
- [Minor] Native-vs-legacy parity coverage is not concrete enough. Claude asks for byte-level parity coverage for every `[requires]` axis if any `bd` probe remains active, and Codex asks for fixtures that prove bd-backed and native paths behave consistently across sling, API sling, order dispatch, convergence, and fanout.

**Disagreements:**
- There is no verdict disagreement: Claude and Codex both return `approve-with-risks`.
- Claude frames the architecture as two valid helper owners, `internal/formula` and `internal/sourceworkflow`, and asks for a stronger shared-helper guard. Codex focuses less on the owner split and more on dashboard/generated-client inference. These are compatible concerns: the synthesis treats them as one guard gap across Go callers and frontend/API projections.
- Claude recommends explicitly retiring `beads.Store.MolCook`/`MolCookOn` and their bdstore and exec implementations, preferably around sub-phase 4f. Codex allows either native-preflighting or fail-closed behavior during the migration window. The synthesis requires both: define the migration-window behavior first, then schedule deletion or a narrowly expired compatibility state for the legacy methods.
- Claude raises `MinimumReaderCapability` reuse semantics as missing evidence; Codex does not mention it. This remains a valid unknown but is not the primary blocker for this persona lane because both reviewers put caller bypasses ahead of artifact-version nuance.

**Missing evidence:**
- No Kimi 2.6 artifact was present for this persona.
- No retirement-table or symbol-end-state row names `beads.Store.MolCook`, `beads.Store.MolCookOn`, `BdStore.MolCook`, `BdStore.MolCookOn`, `exec.Store.MolCook`, or `exec.Store.MolCookOn`.
- No static guard is named that fails when production code outside approved owners or shims calls `Store.MolCook` or `Store.MolCookOn`.
- No checked-in grep-derived caller manifest exists yet for the current tree.
- No fixture proves a legacy `Store.MolCook`, bd-backed, or exec-store `mol-cook` path cannot bypass `[requires]` before all caller migration phases complete.
- No concrete policy states whether external requires-only formulas are accepted, rejected, or required to be dual-declared while legacy materialization methods remain reachable.
- No explicit dashboard or generated-client gate proves graph/compiler state cannot be reconstructed from raw metadata after typed API fields exist.
- No explicit same-identity reuse case describes how `MinimumReaderCapability` participates in persisted-artifact validation after a binary upgrade.

**Required changes:**
- Add a transitional no-bypass gate: before any durable writer accepts requires-only graph formulas, every still-reachable `MolCook`, bd-backed, or exec-store materialization path must either native-preflight through `CompileWithResult`/`AcceptCompileResult` and write only from the accepted artifact, or fail closed with the shared typed diagnostic and zero durable writes.
- Add legacy-helper retirement rows for `beads.Store.MolCook` and `beads.Store.MolCookOn`, plus symbol-end-state rows for `internal/beads/bdstore.go:MolCook`/`MolCookOn` and `internal/beads/exec/exec.go:MolCook`/`MolCookOn`. Name the owning sub-phase, expected end state, and blocking test.
- Add a named guard such as `TestNoStoreMolCookProductionCalls` that fails when production code outside approved owner packages, test doubles, or time-bounded migration shims references `Store.MolCook` or `Store.MolCookOn`.
- Strengthen the raw-consumer guard into a full caller chokepoint test. It should cover raw formula metadata reads, hand-rolled workflow-root `beads.ListQuery` filters, and dashboard/generated-client metadata inference, or else state the typed API invariant that prevents frontend access to raw graph/compiler metadata.
- Add parity and negative fixtures for CLI sling, API sling, formula-backed order dispatch, convergence create/retry, and fanout fragments with `[daemon] formula_v2=false`, requires-only graph formulas, and dual-declared legacy formulas.
- Tighten the convergence transition fence to state that shadow `CompileWithResult` runs before any convergence root, iteration, missing-child, or speculative-wisp write, and that the legacy subset call is projection-only until the caller migrates to accepted artifacts.
