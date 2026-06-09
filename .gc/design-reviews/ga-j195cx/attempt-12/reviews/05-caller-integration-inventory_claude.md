# Yuki Patel - Claude

**Verdict:** approve-with-risks

The design has clearly thought hard about caller integration. The Caller
Inventory table (lines 365-372), the Executable Call-Site Migration table
(lines 378-392), the In-Flight And Convergence Behavior section (lines
707-738), and the static-guard requirement (lines 399-402) collectively cover
the bulk of my mandate. But the inventory has gaps against the actual code,
the bd-path alignment row contradicts what is already in `main`, the in-flight
story does not address fanout-time fragment compiles, and the dead-code policy
omits two existing detectors. Each is fixable with focused edits to this
document, not a redesign.

**Top strengths:**
- The "Canonical Compile Result" section forces every behavioral caller through
  `CompileWithResult`/`CheckRequirements`, makes `HostCapabilities` a per-call
  input, and explicitly demotes `IsFormulaV2Enabled` to a CLI/API edge wrapper.
  This is exactly the seam needed to retire the `formulaV2Enabled` package
  global at `internal/formula/compile.go:483` and the per-package
  metadata-string sniffing.
- The "In-Flight And Convergence Behavior" rules (lines 716-731) draw the right
  line: persisted metadata governs continuation, current host capability
  governs new compiles. That single distinction kills the entire class of
  "what happens to a graph wisp when an operator flips the flag mid-flight"
  questions for already-materialized step beads.
- The static guard at lines 399-402 plus the `gc formula validate
  --legacy-contract-report` command at lines 474-494 give the alias-window
  removal a measurable exit criterion instead of a calendar date.

**Critical risks:**
- [Major] **Caller inventory misses two store-query sites.** Production code
  has at least two places that use raw `gc.formula_contract` / `gc.kind` as a
  *store query filter*, not a post-fetch predicate:
  `internal/api/orders_feed.go:103,226` (passes
  `{"gc.kind":"workflow","gc.formula_contract":"graph.v2"}` as a query
  metadata map) and `internal/api/convoy_sql.go:453` (raw SQL
  `JSON_EXTRACT(i.metadata, '$."gc.kind"') = 'workflow'`). The shared
  `IsWorkflowRoot(b)` predicate the design proposes is post-fetch only — it
  cannot replace a query selector. Without an explicit query-criteria
  contract (e.g. a `WorkflowRootQuery()` helper or a documented "query by
  `gc.kind`, post-filter through `IsWorkflowRoot`" policy), these two sites
  are guaranteed to either keep the legacy `gc.formula_contract` filter
  indefinitely or silently miss requires-only roots after the alias window
  closes.
- [Major] **Compatibility row 472 ("`GC_NATIVE_FORMULA=false` or `bd`
  shell-out path") contradicts the current tree.** Production code shows no
  `MolCook`/`MolCookOn`/`GC_NATIVE_FORMULA` references; phase 4 of
  `engdocs/proposals/formula-migration.md` has already happened. Every
  formula compile in the runtime now goes through
  `formula.CompileWithoutRuntimeVarValidation` (see
  `internal/sling/sling_core.go:764`, `cmd/gc/order_dispatch.go:397`,
  `internal/molecule/molecule.go:130`, `cmd/gc/cmd_order.go:479`,
  `internal/api/handler_formulas.go:183`). The only surviving bd-cook
  invocation I found is `examples/gastown/gastown_test.go:221`. As written,
  the design treats the bd shell-out path as a live runtime concern, which it
  isn't. Either delete the row and tighten the alias-window criteria
  accordingly, or enumerate the bd-cook invocation surfaces (test, tutorial,
  external) the design is keeping alive — the alias-window decision is being
  made on stale assumptions.
- [Major] **In-flight expansion fragment compile is undefined under flag
  flip.** `internal/dispatch/fanout.go:129` calls
  `formula.CompileExpansionFragment(...)` at *runtime*, not at root creation,
  for every fanout item of an in-flight molecule.
  `internal/formula/fragment.go:116` already snapshots
  `IsFormulaV2Enabled()` at that compile. If `[daemon] formula_v2` flips
  from true to false while a graph wisp is in flight, this fragment compile
  re-enters the host-capability check with the new (lower) capability. The
  design at line 716 says "Existing graph workflow roots and their
  already-created step beads continue ... using their persisted metadata,"
  but fanout produces *new* beads from a *new* compile, so the rule at line
  727 ("a retry or `on_complete` action that compiles a new formula uses the
  current host capability") would kill the in-flight expansion. Whichever
  outcome is intended, the design must say it explicitly and add it to the
  Required tests at line 735, which currently lists "expansion/aspect
  requirements" generically without naming the runtime-fanout case.
- [Major] **`Recipe.GraphWorkflow` is referenced as if it exists.** Line 388
  says `internal/graphroute.IsCompiledGraphWorkflow` should "Read
  `Recipe.GraphWorkflow` or `CompileResult.GraphWorkflow`, not root metadata
  strings." Today `Recipe` carries no such field; the existing
  `IsCompiledGraphWorkflow` at `internal/graphroute/graphroute.go:84` reads
  `recipe.Steps[0].Metadata["gc.kind"]` and `["gc.formula_contract"]`.
  Without an explicit Recipe-shape change (and a statement that legacy
  metadata reads remain valid during phase 3), the migration will land
  call-sites pointing at a field the compiler hasn't been taught to populate.
- [Minor] **`requiresExplicitGraphContract` and
  `metadataRequiresGraphContract` are not on the static-guard exemption
  list.** Both functions live at `internal/formula/types.go:882,942` and are
  exactly the v2-only construct detectors the design says to replace with the
  registry. Either they need to appear in the deletion plan (Rollout phase 3
  says "move callers" but not "delete detectors") or they need to be on the
  static-guard exemption list as transitional. Today they are neither, so
  the static guard as drafted would either miss a real regression vector or
  fire on the legacy detectors themselves.
- [Minor] **Two duplicate inline graph-v2 checks in `internal/sling`.**
  `internal/sling/sling.go:783` and
  `internal/sling/sling_attachment.go:101-102` both reimplement the
  `IsWorkflowRoot` logic inline instead of calling
  `sourceworkflow.IsWorkflowRoot`. The design lists "internal/sling routing
  helpers" generically; naming these two functions in the migration table
  would force the implementer to delete the duplicates rather than add a
  third reader of canonical metadata next to them.

**Missing evidence:**
- No enumeration of bd-cook invocation sites that survive after phases 3-4 of
  the existing formula-migration. The design treats the bd shell-out path as
  abstract; the alias-window criteria depend on knowing exactly which call
  sites still hit `bd cook` / `bd mol wisp`.
- No statement of how the canonical metadata (`gc.formula_compiler_capability
  = "2"`, `gc.formula_compiler_requirement = ">=2"`) flows into the API SQL
  query at `internal/api/convoy_sql.go:453`. The persisted-metadata table
  (lines 575-582) describes what is *written*, not what is *queried*.
- No definition of behavior when a formula is required by an *open*
  convergence loop whose root was created at capability `2`, the daemon flag
  is then flipped to false, and a *speculative wisp* fires. The design says
  "convergence adapters must use the same diagnostic path" (line 730) but
  speculative-wisp creation is the path most likely to surprise an operator
  mid-incident.
- No call out for `examples/gastown/gastown_test.go:221`. If first-party
  examples invoke `gc bd mol wisp` on graph-only formulas, dual-declaration is
  not optional for those formulas — it is a hard requirement of the test
  suite. The "first-party packs must dual-declare" criterion (line 442) does
  not cite this test as evidence.

**Required changes:**
- Add a query-criteria row to the Caller Inventory and Executable Call-Site
  Migration tables that names `internal/api/orders_feed.go:103,226` and
  `internal/api/convoy_sql.go:453`. Specify either "query by `gc.kind` only,
  post-filter through `IsWorkflowRoot`" or "introduce
  `WorkflowRootQueryCriteria()` returning a metadata map." Whichever is
  chosen, the static guard at lines 399-402 must accept the chosen exemption.
- Reconcile the bd-path row at line 472 with the actual tree. Either
  (a) delete the row and add a one-sentence note that bd shell-out for
  formula compilation was retired in formula-migration phase 4, with the
  exception of `examples/gastown/gastown_test.go:221`; or
  (b) enumerate the bd-cook invocation sites that survive (tests, tutorials,
  external) and require those formulas to remain dual-declared explicitly.
- Add a row to the In-Flight And Convergence Behavior rules that names
  `formula.CompileExpansionFragment` and states whether a flag flip kills an
  in-flight fanout. Add a corresponding test entry to line 735.
- State explicitly that `Recipe.GraphWorkflow` (and
  `CompileResult.GraphWorkflow`) is a new field added in rollout phase 1 or
  2, that `formula.toRecipe` populates it from
  `NormalizedRequirements.FormulaCompiler`, and that
  `IsCompiledGraphWorkflow`'s legacy metadata read remains valid during phase
  3 so call-sites can migrate gradually.
- Either list `requiresExplicitGraphContract` and
  `metadataRequiresGraphContract` as deletion targets in rollout phase 3
  (preferred) or add them to the static-guard exemption list with a
  deprecation note.
- In the Executable Call-Site Migration table, name
  `internal/sling/sling.go:isWorkflowRoot` (line 783) and
  `internal/sling/sling_attachment.go:isWorkflowRoot` (line 101-102) as
  delete-and-replace targets pointing at `sourceworkflow.IsWorkflowRoot`.

**Questions:**
- Is `bd cook` / `bd mol wisp` still a supported runtime path that gas city
  binaries can shell out to, or is it strictly a CLI-tier artifact for human
  invocation? The answer determines whether the alias window must wait on
  external pack consumers or only on first-party migration.
- Should the API SQL query at `convoy_sql.go:453` be migrated to canonical
  metadata, kept on `gc.kind=workflow` permanently (since `gc.kind` is not
  scheduled for deprecation), or rewritten to use a typed
  `WorkflowRootQueryCriteria` helper? The decision affects how
  `internal/sourceworkflow` is shaped after phase 3.
- When a controller running with `formula_v2 = false` ingests a workflow root
  that another controller (with `formula_v2 = true`) created, dispatch reads
  persisted `gc.formula_compiler_capability = "2"` per line 580. Does the
  receiving controller refuse to dispatch the persisted graph workflow, or
  does it dispatch the already-materialized step beads while refusing only
  *new* compile requests? Line 716 implies the latter, but the
  Compatibility-matrix row "Mixed-version controllers sharing a bead store"
  (line 470) only requires dual-stamping, not behavior.
- Is the speculative-wisp creation path inside `internal/convergence` covered
  by the same `CompileWithResult` preflight as `convergence_store.go:177`'s
  `molecule.Cook` call, or does it have its own compile entry? The design
  says "speculative wisp creation must preflight" (line 723) but does not
  name the file or function.
