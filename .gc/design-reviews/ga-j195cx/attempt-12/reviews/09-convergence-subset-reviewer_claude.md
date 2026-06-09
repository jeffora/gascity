# Felix Berger - Claude

**Verdict:** approve-with-risks

**Top strengths:**
- Single-parser invariant: "`internal/formula` is the only package that may
  interpret raw `contract`, `requires.formula_compiler`, `version`, or v2-only
  construct strings" structurally eliminates the convergence-subset drift class
  by removing convergence's standing to ever decode `[requires]` itself.
- Convergence is named explicitly in the host-capability gate list (create,
  retry, speculative-wisp), and the migration matrix demands "Convergence
  tests for disabled host capability with **zero durable writes**" — the
  fail-closed-before-bead-create posture is the right one for a subsystem
  whose root is durable.
- Inheritance rule ("Inline children and `loop.body` inherit the containing
  formula's normalized requirement") cleanly handles the controller-injected
  `evaluate` step: it has no formula header of its own, so the parent's
  normalized requirement covers it without convergence having to think about
  v2 capability for synthesized steps.

**Critical risks:**
- **[Major] `convergence.Formula` ↔ `CompileResult` coupling is unspecified.**
  The current Go subset (`Name`, `Convergence`, `RequiredVars`,
  `EvaluatePrompt`, `StepNames`) carries no normalized-requirement field. The
  design says convergence "validates through `internal/formula` preflight or
  an adapter over `CompileResult`" but never says whether
  `convergence.Formula` is retired, populated from `CompileResult`, or kept
  as a parallel decode path. Until that decision is in the design, the next
  contributor who needs convergence to know "is this a graph-v2 loop?" will
  add a private decode and the drift class re-emerges.
- **[Major] `molecule.Cook`/`CookOn` is the actual convergence wire and the
  design does not update its signature.** Production convergence flows
  through `cmd/gc/convergence_store.go.pourWisp` →
  `molecule.Cook(ctx, store, formula, paths, molecule.Options{...})`.
  `molecule.Options` has no `HostCapabilities` field, and `CompileOptions`
  (which does) lives in `internal/formula`. The gate the design promises only
  closes once that plumbing exists. The doc must say which it is: extend
  `molecule.Options`, take a new `HostCapabilities` arg, or require callers
  to pre-compile via `CompileWithResult` and pass `*CompileResult` into
  `molecule.Instantiate`. Without this, "Convergence tests for disabled host
  capability with zero durable writes" cannot be implemented from the design
  alone.
- **[Major] Diagnostic projection through convergence event surfaces is not
  enumerated.** The design states "Controller and convergence paths wrap the
  same `Diagnostic` code and fields in their existing error/event surfaces."
  Convergence has many typed events (create/start/iteration/terminated/
  failure variants in `internal/convergence/events.go`). Without naming which
  events carry the typed `Diagnostic` payload — and which payload registry
  entry covers `formula.compiler_requirement_*` for convergence — the
  "operators get one stable diagnostic contract" goal in §Consequences will
  silently drift the moment convergence emits a hand-rolled error. This is
  the same drift class the design otherwise eliminates for parsing, repeated
  one layer up.
- **[Major] Convergence retry/in-flight policy when `[daemon] formula_v2`
  flips off mid-loop is under-specified.** §In-Flight covers continuation of
  "already-created step beads," but a convergence root pours a *new* wisp
  every iteration. The doc says new wisp creation must preflight and "create
  no root, child bead, or partial convergence state," but does not answer:
  does the active convergence root keep iterating and emit a typed failed-
  iteration event each tick, terminate itself, or become `waiting_manual`?
  Each option has very different operator UX. The convergence persona needs
  the design to pick one.

**Missing evidence:**
- Disposition of `internal/convergence/formula.go` and `ValidateForConvergence`.
  In current `main` this code has zero non-test callers; the design treats it
  as an active subset-validator without acknowledging it is currently dead
  code on the create path. Either it becomes the post-compile domain checker
  (called after `CompileWithResult` succeeds) or it is removed.
- The migration matrix lists "`internal/convergence/create`, retry,
  speculative wisp adapters" but the production code lives in
  `cmd/gc/convergence_store.go` (`pourWisp`/`PourSpeculativeWisp`) and
  `cmd/gc/cmd_converge.go`, not under `internal/convergence`. The doc should
  name the actual files so reviewers can verify the gate is wired at the
  right seam.
- Required-test list does not include convergence-specific scenarios
  (graph-v2 convergence formula + flag off → zero beads; v2-only construct +
  missing `[requires]` in a convergence formula; speculative-wisp conflict;
  mid-loop disable behavior; inherited requirement on the controller-
  injected `evaluate` step).
- No statement about convergence-specific persisted metadata. Workflow roots
  get canonical `gc.formula_compiler_*` keys; should convergence roots
  (which already carry a separate `convergence.formula` namespace) also
  carry the canonical keys, or only the canonical keys via the shared
  workflow-root predicate? The shared-predicate compatibility window
  paragraph implies the latter, but convergence is the only subsystem where
  the root and the workflow-root concept fully coincide, and that should be
  said.
- No table-row in the §Persisted Metadata matrix for `convergence` roots
  vs. graph-workflow roots, even though §Caller Inventory names convergence
  separately.

**Required changes:**
- Add a "Convergence integration" subsection (peer to "Compatibility With
  `contract`" and "In-Flight And Convergence Behavior") that pins down: (a)
  the exact convergence Go entry points that must call `CompileWithResult`
  before any durable write — name the files: `cmd/gc/convergence_store.go`
  `pourWisp`/`PourSpeculativeWisp`, plus any `internal/convergence/create.go`
  or `retry.go` paths that pour wisps; (b) disposition of
  `internal/convergence/formula.go` (`Formula` struct +
  `ValidateForConvergence`) — retire, repurpose as post-compile domain
  checker, or extend with a `NormalizedRequirement` field sourced from
  `CompileResult`; (c) the `molecule.Cook`/`CookOn` API surface that
  threads `HostCapabilities` (or the rule that convergence pre-compiles via
  `CompileWithResult` and calls `molecule.Instantiate(*CompileResult)`); (d)
  the convergence event types that carry typed `Diagnostic` payloads, and
  the registry entries those payloads must use.
- Extend the "Required tests" enumeration with: convergence-create with
  graph-v2 formula + host disabled (zero durable writes, typed event);
  convergence-create with v2-only construct + missing `[requires]` (matches
  `formula.compiler_requirement_missing` deterministic order);
  speculative-wisp conflict diagnostic; inherited-requirement coverage of
  the controller-injected `evaluate` step; mid-loop disable behavior of an
  active convergence root.
- Pick and state the mid-loop policy for `[daemon] formula_v2` going
  `true` → `false` while a convergence root is `state=active`. The
  `formula.compiler_requirement_unsatisfied` diagnostic already exists; what
  the convergence subsystem does *with* it (terminate, stall, retry-loop)
  must be a settled invariant, not a behavior an operator infers.
- Add a row to the §Persisted Metadata table or a one-line note clarifying
  that `convergence` roots carry the canonical `gc.formula_compiler_*` keys
  via the same shared workflow-root predicate. If they additionally need a
  `convergence.formula_compiler_capability` mirror for convergence-specific
  views, say so explicitly; otherwise say they do not.
- Either delete `internal/convergence/formula.go` from the design's "subset"
  language and move convergence's row in the caller-migration table to point
  at `cmd/gc/convergence_store.go`, or explicitly retain
  `ValidateForConvergence` and define when it runs (post-compile, given a
  `CompileResult.Recipe`). Pick one; the current text supports both
  readings.

**Questions:**
- For the convergence wire: extend `molecule.Options` with
  `HostCapabilities`, change `Cook`/`CookOn` to take an explicit
  `HostCapabilities` argument, or require convergence (and other callers
  with subtle gate semantics) to compile via `CompileWithResult` and pass
  `*CompileResult` into `molecule.Instantiate`? The first preserves the
  one-call convenience API; the third matches the "all behavioral callers
  must consume a normalized compile result" principle most strictly.
- Does the convergence subset (`internal/convergence/formula.go`) survive
  the migration? If yes, what is its post-`CompileResult` job — only domain
  checks (reserved step name, evaluate-prompt content) on the compiled
  Recipe? If no, when is it deleted, and which PR removes its currently
  test-only references?
- For convergence rooted on a graph-v2 formula, when `[daemon] formula_v2`
  is disabled while the root is `state=active`: terminate immediately with
  a typed final event, stall in `waiting_manual`, or fail each subsequent
  iteration's pour with `formula.compiler_requirement_unsatisfied` and
  leave the root open? The right choice depends on whether the operator
  changing the flag is treated as an immediate eviction or a graceful
  drain, and that is a product decision.
- Are existing typed convergence events (`convergence.create_failed`,
  `convergence.iteration_failed`, etc.) sufficient carriers for the typed
  `Diagnostic` payload, or do we need a new
  `convergence.compile_requirement_failed` event class? §Operator
  Diagnostics requires no event for fatal CLI/API diagnostics except
  "controller/order failure wrapper" — convergence is a third wrapper and
  should be named.
- Does `gc formula validate --provenance` surface convergence-specific
  fields (`required_vars`, evaluate-prompt path, reserved-step assertions)
  alongside the `[requires]` requirement, or are those left to a separate
  `gc convergence validate` style command? The doc's "Read-only provenance
  surface" reads as canonical-fields-only.
