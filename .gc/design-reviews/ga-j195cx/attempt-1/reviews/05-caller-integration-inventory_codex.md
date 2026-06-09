# Yuki Patel - Codex

**Verdict:** block

**Top strengths:**
- The per-occurrence caller manifest is now tied to generated preflight rows, zero-write fixtures, static raw-consumer scans, and prompt/tutorial producer rows. That is the right shape for preventing a half-migrated caller surface.
- The runtime compatibility decision is clear: production is native-only, and any `bd` or `GC_NATIVE_FORMULA=false` path is a validation-only probe that cannot authorize durable writes.
- In-flight and fanout behavior is much more executable than earlier versions: same-identity accepted-artifact reuse, changed-identity recompilation, legacy-only observation, and fragment zero-write behavior are all described in one place.

**Critical risks:**
- [Blocker] `CompileResult` still has two incompatible contracts. The canonical shape at `engdocs/design/formula-compiler-requirements.md:1331` contains `Recipe`, `Requirements`, `GraphWorkflow`, `Diagnostics`, `Provenance`, and `Projection`. The later convergence section says it is the "same canonical `CompileResult` shape" but adds `Steps []CompiledStep` and `RuntimeVars []CompiledRuntimeVar` at `engdocs/design/formula-compiler-requirements.md:5662`. Callers cannot be migrated safely if convergence, fanout, and durable writers disagree about whether those fields live directly on `CompileResult`, under `Projection`, or behind an accepted-artifact projection API.
- [Blocker] The source-workflow API transition is internally inconsistent and does not ledger current behavior changes. The design requires new APIs such as `ClassifyWorkflowRoot(metadata)`, `IsWorkflowRootMetadata(metadata)`, and `ListWorkflowRoots(...)` at `engdocs/design/formula-compiler-requirements.md:1678`, but later says current `internal/sourceworkflow.IsWorkflowRoot` and `ListLiveRoots` should be kept as the sole predicate boundary at `engdocs/design/formula-compiler-requirements.md:1890`. The live package currently exposes `IsWorkflowRoot(beads.Bead)` and `ListLiveRoots(...)`, and `IsWorkflowRoot` case-folds both `gc.kind` and `gc.formula_contract` (`internal/sourceworkflow/sourceworkflow.go:44`). The design also says legacy fallback trims whitespace and "never case-folds" at `engdocs/design/formula-compiler-requirements.md:1770`. That is a narrowing behavior change, but there is no per-call-site compatibility ledger saying which callers intentionally change, which keep compatibility wrappers, and when those wrappers expire.
- [Major] API and projection callers are named, but the design still needs one concrete source-workflow migration row that proves duplicated predicates converge. Live API code still has local workflow-root predicates and direct metadata queries, for example `internal/api/orders_feed.go:100` and `internal/api/handler_convoy_dispatch.go:204`. The design says "use `internal/sourceworkflow.ListWorkflowRoots`", but does not show how API closed-history, graph-only, source-scoped, and fallback query semantics map to the new criteria and typed facts.
- [Major] The mandatory prompt/template inventory still lacks a stable source binding for workflow-pack prompt snippets outside this repo. The design lists "workflow prompt template source directories" and checked-in generated prompt fixtures at `engdocs/design/formula-compiler-requirements.md:2061`, and excludes ignored `.gc/system/packs` as a CI input. If the workflows pack is first-party but not vendored in the Gas City checkout, CI needs a lockfile, pinned source path, or deterministic fixture-generation contract that names exactly what is scanned.
- [Major] The context docs still contradict the native-only caller story. `engdocs/architecture/formulas.md` says `bd` is the production path and `Store.MolCook`/`MolCookOn` are the runtime seam, while `docs/reference/formula.md` still documents `Store.MolCook*` as user-facing formula instantiation. The design says `engdocs/proposals/formula-migration.md` must be superseded, but the architecture and reference docs are also direct stale guidance that can reintroduce `bd` and convergence-subset assumptions into caller migration work.

**Missing evidence:**
- A single authoritative `CompileResult`/projection snippet used by convergence, fanout, retry, API preview, and durable writers.
- A before/after table for every current source-workflow API: `IsWorkflowRoot`, `ListLiveRoots`, `WorkflowMatchesSource`, API `isWorkflowRoot`, sling attachment checks, convoy dispatch, and order/feed projections.
- Golden rows showing case-sensitive canonical metadata, whitespace-only legacy compatibility, and current case-folded legacy roots, with an explicit decision for whether existing case variants are still visible or become migration diagnostics.
- A concrete checked-in caller manifest example that covers at least one API direct metadata query, one sourceworkflow predicate, one fanout fragment compile, one prompt-template producer, and one validation-only `bd` probe.
- A pinned source or fixture contract for external-but-first-party workflow pack prompts and formulas.

**Required changes:**
- Define `CompileResult` once. If convergence needs `Steps` and `RuntimeVars`, put them in `PreviewProjectionSnapshot` or add them to the canonical shape and update every consumer contract to use the same access path.
- Resolve the source-workflow API plan: either keep `IsWorkflowRoot`/`ListLiveRoots` as named compatibility wrappers over `WorkflowRootFacts`, or replace them with `IsWorkflowRootMetadata`/`ListWorkflowRoots`; do not require both without wrapper semantics, expiry phase, and tests.
- Add an explicit current-behavior ledger for workflow-root predicates, including exact match vs trim vs case-fold, `kind` only, `contract` only, `kind && contract`, and source-scoped closed-history queries.
- Add one generated caller-manifest excerpt to the design, not just schema prose, and make it cover the duplicated API/sourceworkflow predicates and prompt producer references that are most likely to drift.
- Add `engdocs/architecture/formulas.md` and `docs/reference/formula.md` to the stale-guidance blocking set, or update/mark them superseded in the same phase that exposes user-visible `[requires]` diagnostics.

**Questions:**
- Should current legacy roots with `gc.formula_contract=Graph.V2` remain visible during the alias window, or is the new trim-only/no-case-fold rule intentionally a breaking cleanup?
- Is the workflows pack considered a Gas City first-party input for this migration? If yes, where is its reproducible source revision recorded for CI scans?
- Will `ListLiveRoots` remain a supported API for source-scoped conflict checks, or is the target to force all callers through `ListWorkflowRoots` criteria even for launch-lock conflict detection?
