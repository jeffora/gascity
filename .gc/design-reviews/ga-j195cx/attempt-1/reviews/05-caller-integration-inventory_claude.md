# Yuki Patel - Claude

**Verdict:** approve-with-risks

Reviewed `engdocs/design/formula-compiler-requirements.md` (6089 lines, on disk)
against the caller-integration lane: caller rewrite; bd-vs-native path
alignment; in-flight molecule behavior; dead-code policy. Cross-checked against
the live tree (`internal/formula/compile.go`, `internal/formula/fragment.go`,
`internal/sling`, `internal/graphroute`, `internal/api`,
`internal/dispatch/fanout.go`, `internal/convergence`, `internal/sourceworkflow`,
`internal/molecule`, `cmd/gc`) and `engdocs/proposals/formula-migration.md`.
This version of the design materially closes earlier prompt-template / bd-mol
gaps (it now requires the caller manifest to classify first-party
`gc bd mol …` references) and adds a real `--alias-removal-gate` precondition
for the requires-only milestone, so the bd-fallback red flag is largely
neutralized. Two in-flight-behavior gaps and one predicate-reconciliation gap
remain.

**Top strengths:**
- The anti-coexistence machinery is real, not aspirational: per-occurrence
  caller manifest, `raw_consumer_allowlist.yaml` with per-row owner + expiry
  phase + replacement test, `TestNoNewFormulaRawConsumers`,
  `TestCallerManifestDrivesPreflightMatrix`, named deletion tests
  (`TestDeclaresGraphV2ContractDeleted`, `TestValidateForConvergenceNotImported`,
  `TestNoLegacyMolProducerPromptReferences`), and the rule that "every later
  caller sub-phase must remove or narrow at least the rows it owns" directly
  answer "old and new requirement checks coexist indefinitely." The migration
  completion criterion ("not complete while any non-test production code outside
  `internal/formula` or the shared predicate branches on `Contract`,
  `declaresGraphV2Contract`, `Requires.FormulaCompiler`, or `gc.formula_contract`")
  is the right one.
- The bd-fallback red flag is closed structurally: the runtime is native-only
  (`Store.MolCook`/`MolCookOn` are already gone from the tree), every
  requirement interpretation is owned by `internal/formula` with
  `TestNoRawFormulaMetadataReadsOutsideOwners`, and the bd path is at most a
  `not-needed`-by-default release-validation probe that must prove byte-level
  `[requires]` parity if it is ever marked active. Prompt-template / tutorial /
  example references to `gc bd mol …` are explicitly first-class producer
  candidates that the manifest must classify before requires-only distribution.
- In-flight *end-state* semantics are coherent: requirements are evaluated at
  compile time for new roots/wisps/attachments/expansions and never
  re-evaluated for already-created beads; same-identity accepted-artifact reuse
  lets a disabled host keep mutating an already-accepted root, while any changed
  formula/vars/options/source recompiles against the current host and fails
  closed with zero writes; `gc formula repair-root-artifact` is a defined,
  idempotent operator path with a release-fixture dry-run JSON.

**Critical risks:**

- **[Major] "Previous behavior must match" is under-specified because the
  current workflow-root predicates disagree with each other, and the design's
  own "matching is exact … no case-folded variants" rule silently changes
  behavior at one of them.** Today the tree has at least five subtly different
  "is this a (graph) workflow root" predicates:
  - `internal/sling/sling.go:783` — `kind == "workflow" && formula_contract == "graph.v2"`, no trim, exact, AND.
  - `internal/sling/sling_attachment.go:118` (`IsWorkflowAttachment`) — `EqualFold(TrimSpace(kind), "workflow") || EqualFold(TrimSpace(formula_contract), "graph.v2")` — **case-folded, trimmed, OR**.
  - `internal/graphroute/graphroute.go:84` — `kind == "workflow" && formula_contract == "graph.v2"`, no trim, exact, AND.
  - `internal/api/orders_feed.go:111,237` — `isWorkflowRoot(bead) && TrimSpace(formula_contract) == "graph.v2"` — trimmed, exact value.
  - `internal/api/handler_convoy_dispatch.go:205` / `internal/api/convoy_sql.go:453` — `TrimSpace(kind) == "workflow"` / SQL `JSON_EXTRACT('$."gc.kind"') = 'workflow'` — `kind`-only, no contract check.

  The `OR` site (`sling_attachment.go`) matches *more* beads than the `AND`
  sites; the `kind`-only sites match a different set again. The design's
  Phase-0 `workflow_root_parity/` fixtures cover "whitespace-variant" and
  "case-variant" stores, but a single shared predicate cannot reproduce the
  output of all five sites simultaneously, so "the callers whose previous
  behavior must match" is impossible to satisfy as written. The design needs an
  explicit **reconciliation policy** — e.g. "the shared predicate is exact
  `gc.kind == "workflow"` for `WorkflowRootAny`, plus canonical keys or
  whitespace-trimmed `gc.formula_contract == "graph.v2"` (never case-folded) for
  the graph distinction; the `kind`-only convoy/SQL sites map to
  `WorkflowRootAny`, the `kind AND contract` sites map to
  `WorkflowRootKnownGraphOnly`/`dual_stamped`, the `EqualFold` `OR` site is a
  deliberate narrowing" — **plus a per-call-site behavior-change ledger** naming,
  for each of the five sites, whether the moved behavior is identical / narrower
  / wider and in which direction. The `EqualFold` + `OR` divergence in
  `sling_attachment.go` is exactly the silent-divergence case this design exists
  to kill; it should be enumerated as a named row, not subsumed under a generic
  "case-variant" fixture.

- **[Major] The in-flight behavior of fanout fragment expansion after a host
  downgrade is genuinely ambiguous — accepted-projection-snapshot reuse vs
  re-compile give opposite answers — and that ambiguity directly realizes "in-
  flight molecule enters an undefined state."** "Canonical Compile Result" /
  "DR53-accepted-projection-snapshot" says `AcceptedProjectionSnapshot` is
  populated "from the fully decoded, inherited, **expanded**, and
  source-attributed compiler output" and is "the only data durable convergence,
  **fanout**, retry, scope-check, workflow-finalize, and missing-child repair may
  use when they need formula facts after the original source may have moved or
  disappeared." That implies fanout fragments of an already-accepted root are
  *in* the snapshot and re-materializing them is "graph state already
  represented by the accepted projection snapshot" → same-identity reuse allowed
  even with the host disabled. But "Required migration rows" and the
  "In-Flight And Convergence Behavior" / host-downgrade table say fanout
  "compiles expansion fragments … with the current host capability" and
  "parent accepted artifact plus fragment `CompileWithResult` /
  `AcceptCompileResult` calls before expansion writes" → a not-yet-materialized
  fragment can never compile once `formula_v2` flips and the workflow stalls. So
  a graph workflow whose root was accepted under `formula_v2=true` either
  proceeds (snapshot reuse) or silently stalls (re-compile) — and a convergence
  loop in the identical situation explicitly reuses the persisted artifact and
  proceeds. The design must pick one and write it next to the convergence rule.
  If the answer depends on inline-expansion `template` (already inside the
  accepted artifact's compiled `Steps`) versus `compose.expand` of a separate
  formula file (a fresh formula selection), say so — the v2-only construct
  registry already draws that line (`step_expand` vs `compose_expand`), but the
  In-Flight section and the fanout row do not.

- **[Major] The "In-Flight And Convergence Behavior" section assumes
  infrastructure that does not exist until Phase 2 / sub-phase 4f, leaving a
  transitional hole.** Accepted artifacts land in Phase 2; `gc formula
  repair-root-artifact` is scheduled in caller sub-phase 4f. Between Phase 1 and
  4f, v2 roots are created the *current* way: no accepted artifact, only
  `gc.kind=workflow` + `gc.formula_contract=graph.v2`. The current fanout path
  (`internal/dispatch/fanout.go:129` → `formula.CompileExpansionFragment` →
  `isGraphWorkflow(resolved, IsFormulaV2Enabled())`), Ralph next-iteration, and
  retry/`on_complete` all **recompile a formula on each expansion**. If an
  operator disables `[daemon] formula_v2` mid-flight, the next fragment fails
  closed — with no accepted-artifact same-identity reuse path (artifacts don't
  exist yet, and the root has none) and no repair command (not built until 4f).
  The molecule stalls with no operator remedy except re-enabling the flag. The
  host-downgrade tables read as if the artifact/repair machinery is always
  present. Required: an explicit **"transitional in-flight behavior, Phases 1
  through 4e"** subsection — most likely "today's behavior is preserved
  unchanged; the only stuck state is disabling `formula_v2` while a v2 loop is
  mid-expansion, which is already the current behavior; repair becomes available
  in 4f" — or move `gc formula repair-root-artifact` earlier than 4f (the
  "Questions" below ask whether it can land in 4a alongside the artifact APIs).

**Missing evidence:**
- Whether `internal/sourceworkflow.IsWorkflowRoot`, `ListLiveRoots`,
  `ListWorkflowRoots`, `ClassifyWorkflowRoot`, `IsWorkflowRootMetadata`, and the
  `WorkflowRootKind` enum already exist. They do not — the package today
  contains only `CloseWorkflowSubtree`, `ConflictError`,
  `SourceStoreRefMetadataKey`. The "Executable call-site migration" row
  "`internal/sourceworkflow.IsWorkflowRoot` and `ListLiveRoots` | Keep
  `internal/sourceworkflow` as the sole persistence predicate" reads as if
  these are current; they are net-new deliverables of Phase 2/4a. Say so, so the
  manifest scan does not look for symbols that aren't there.
- A plain statement that the `bd`/`GC_NATIVE_FORMULA=false` runtime fallback is
  *already removed* (`formula-migration.md` Phase 4 is done in the tree). The
  `bd mol cook …` strings in `cmd/gc/cmd_sling.go` are operator-facing help
  text, not a runtime authorization path. The design hedges correctly
  (`status: not-needed`), but it never states this once up front, so a reader is
  left thinking there is a live parallel interpreter that must stay byte-
  identical with `[requires]` semantics.
- How `TestRawConsumerAllowlistExpiry` decides "the owning sub-phase is
  complete." There is no checked-in rollout-phase state for the guard to read,
  so a stale allowlist row whose sub-phase never formally "completes" can linger
  — the very coexistence failure the guard is meant to prevent. A
  `docs/release/formula-compiler-rollout-phase.json` (release-captain-owned,
  naming the highest landed sub-phase) would give the guard something concrete
  to check.
- Whether the caller-manifest seed grep list (the design enumerates
  `CompileWithoutRuntimeVarValidation`, `CompileExpansionFragment`,
  `SetFormulaV2Enabled`, `IsFormulaV2Enabled`, `gc.formula_contract`,
  `contract = "graph.v2"`, `ValidateForConvergence`, `molecule.Cook`,
  `molecule.Attach`, `molecule.Instantiate`, `GraphWorkflow`, `gc.fanout_state`,
  `gc.continuation_group`, `gc.kind`, `exec.Command("bd"`, API/global
  projections, legacy graph-contract helpers, workflow-root metadata filters)
  explicitly includes `bd mol`, `gc bd mol`, `mol wisp`, `mol bond`,
  `mol cook`, `mol cook-on` as literals. The prose requires prompt/tutorial/
  example classification of those commands, but the literal grep list does not
  name them — only `exec.Command("bd"`, which won't catch markdown prompt text.
- An explicit decision (and fixtures) for fanout fragments of an
  already-accepted root on host downgrade — reuse-snapshot vs fail-closed, and
  the inline-template vs compose-expand distinction — paired with the
  convergence next-iteration rule so the two cannot contradict.

**Required changes:**
- Add an explicit reconciliation policy for the divergent current workflow-root
  predicates plus a per-call-site behavior-change ledger (`sling.go:783`,
  `sling_attachment.go:118`, `graphroute.go:84`, `api/orders_feed.go:111,237`,
  `api/handler_convoy_dispatch.go:205`, `api/convoy_sql.go:453`): for each,
  whether the moved behavior is identical / narrower / wider, which direction,
  and which `WorkflowRootKind` it maps to. Make the Phase-0
  `workflow_root_parity/` golden encode the case-fold→exact and OR→AND/`kind`-only
  reconciliation, not just "case-variant store → some output."
- Resolve the fanout-vs-convergence in-flight inconsistency: state whether
  fanout fragments of an already-accepted root reuse the root's accepted
  projection snapshot / host capability or fail closed against the current host;
  distinguish inline-expansion `template` fragments (in the accepted `Steps`)
  from `compose.expand` of a separate formula file; add zero-write / no-stall
  fixtures for both directions of a `formula_v2` flip during fanout, retry/
  `on_complete`, attach, order scans, convergence create/retry, and root-only
  wisps. Keep the convergence rule and the fanout rule textually adjacent so a
  future edit cannot drift them.
- Add a "transitional in-flight behavior, Phases 1 through 4e" subsection
  describing host-downgrade behavior for legacy-only v2 roots created before
  accepted artifacts (Phase 2) and before `gc formula repair-root-artifact`
  (sub-phase 4f) exist — or move the repair command to a sub-phase that lands
  before any recompile-on-expansion path (fanout / Ralph / retry-`on_complete` /
  convergence-next-iteration) starts depending on accepted artifacts.
- Clarify net-new vs existing in `internal/sourceworkflow` (predicate/query API
  added in Phase 2/4a; package today owns only `CloseWorkflowSubtree`/
  `ConflictError`), and state once in "Runtime compatibility decision" that the
  `bd` shell-out runtime path is already removed and the `bd` probe is
  `not-needed` by default.
- Mark the four caller tables ("Caller inventory and required replacement
  behavior", "Executable call-site migration", grep-derived manifest minimum
  rows, "Required migration rows" for convergence/fanout) as rendered excerpts
  of the single generated per-occurrence manifest, and reconcile every cited
  symbol name against the tree: the design names `internal/sling.isGraphSlingFormula`
  and `InstantiateSlingFormula`, but the tree has `IsGraphWorkflowAttachment`
  (store-backed metadata predicate), `IsCompiledGraphWorkflow`/`graphroute`
  delegation, the inline check in `sling_attachment.go:118`, `Instantiate`/
  `InstantiateFragment` in `internal/molecule`, `declaresGraphV2Contract` /
  `isGraphWorkflow` / `IsFormulaV2Enabled` in `internal/formula`,
  `CompileWithoutRuntimeVarValidation`, `CompileExpansionFragment`, and
  `ValidateForConvergence`; the manifest must enumerate `convergence.Formula{`,
  `ResolveEvaluateStep`, and `ValidateEvaluatePrompt` as occurrences too, not
  just the one validate function.
- Add `bd mol`, `gc bd mol`, `mol wisp`, `mol bond`, `mol cook`, `mol cook-on`
  to the manifest seed grep list explicitly (not only `exec.Command("bd"`), so
  the prompt/tutorial/example producer requirement is mechanically enforced from
  the seed, not just from prose.
- Specify how `TestRawConsumerAllowlistExpiry` reads rollout-phase state
  (`docs/release/formula-compiler-rollout-phase.json` or equivalent) and tag
  each manifest row with its owning sub-phase explicitly — package boundaries do
  not line up with sub-phase boundaries (`internal/sling` owns both a launcher
  (4b) and the `IsGraphWorkflowAttachment` predicate (4e); `fragment.go`'s
  `isGraphWorkflow` is shared by compile (4a) and fanout (4f)), so the "4b–4e in
  parallel if write boundaries don't overlap" claim only holds if the manifest,
  not the package layout, is the arbiter.
- Minor doc-consistency: `CompileResult` is defined twice with different shapes
  — `{Recipe, Requirements, GraphWorkflow, Diagnostics, Provenance, Projection}`
  in "Canonical Compile Result" and `{… , Steps, RuntimeVars}` in "In-Flight"
  ("abbreviated"). The "abbreviated" form *adds* fields; `Steps`/`RuntimeVars`
  already live inside `Projection PreviewProjectionSnapshot`. Define it once and
  have convergence read `Steps`/`RuntimeVars` from `result.Projection`.

**Questions:**
- For a graph workflow whose root was accepted under `formula_v2=true`: when the
  flag is later disabled, does a not-yet-materialized fanout fragment (a) reuse
  the root's persisted accepted projection snapshot / host capability and
  proceed (parity with convergence next-iteration), (b) fail closed and stall
  the workflow, or (c) behave differently for inline-expansion `template`
  fragments vs separately resolved `compose.expand` formulas? Pick one and
  document it next to the convergence rule.
- Can `gc formula repair-root-artifact` land in caller sub-phase 4a alongside
  the accepted-artifact APIs instead of 4f? Bringing it forward closes the
  transitional in-flight hole without a separate caveat subsection.
- Does the Phase-1 PR land a behavior-preserving shim of the shared workflow-root
  predicate (delegating to the current `kind && contract` / `EqualFold` logic)
  so that 4e is a pure call-site swap, or does 4e change predicate semantics
  *and* call-site usage in the same step? The former is the reversible path the
  rollout plan implies; if it is the latter, name it.
- Which existing function becomes the single canonical workflow-root predicate —
  a net-new `internal/sourceworkflow.IsWorkflowRoot`, the store-backed
  `internal/sling.IsGraphWorkflowAttachment`, or `internal/graphroute`'s
  metadata predicate — and do `internal/graphroute`'s metadata predicate and its
  recipe predicate (`IsCompiledGraphWorkflow`) both survive, or does one fold
  into the other?
- Which first-party formulas, examples, tutorials, and `internal/testfixtures/`/
  `test/` files pair a v2-only construct with `version = 1` and no `[requires]`
  today (now declared invalid by this design), and is rewriting them part of
  rollout Phase 1 or a prerequisite to it?

```bash
bd update "$GC_BEAD_ID" --set-metadata design_review.output_path="$OUT" --set-metadata gc.outcome=pass --status closed
```
