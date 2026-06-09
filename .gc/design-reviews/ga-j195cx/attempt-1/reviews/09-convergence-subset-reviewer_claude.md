# Felix Berger - Claude

**Verdict:** approve-with-risks

**Top strengths:**
- The "Convergence implementation decision" section (≈L5432) and the
  package-boundary normative flow (L5539–5549) are unambiguous: the
  convergence subset parser is demoted to a typed projection over the
  canonical `CompileResult`, and `internal/formula` owns `[requires]`,
  legacy `contract`, host-capability satisfaction, and accepted-artifact
  authority. The field-ownership table (L5627–5640) closes the door on
  convergence reading raw TOML for any requirement-adjacent field, and
  the executable replacement pseudocode (L5562–5588) shows the exact
  call shape every convergence entry point must take.
- The transition fence (L5443–5451) explicitly blocks `[requires]` bypass
  through convergence during sub-phases 4b–4f: any caller that has not
  yet moved to `AcceptedCompileArtifact` must still shadow-compile and
  fail closed on requirement/host/projection/provenance diagnostics. That
  closes the most realistic drift channel between the canonical compiler
  and `ValidateForConvergence` while migration is in flight.
- Daemon-gate coverage is symmetric with dispatch: convergence
  create/retry/next-iteration/speculative-wisp paths consume the same
  `HostCapabilities` derived from `[daemon] formula_v2` (L3271–3275,
  L4761), and unsatisfied requirements emit `convergence.formula_compile_failed`
  (L3488, L5524) with stable subject keys (L5482–5489) and a typed
  `ConvergenceFormulaBlockedStatus`/`waiting_on_formula_requirement`
  state (L5515–5535). Same-identity persisted artifact reuse rules
  (L5415, L5479) keep already-active loops from being toppled by a host
  downgrade without giving them authority to absorb a new `[requires]`.

**Critical risks:**
- [Major] `ValidateForConvergence` / `internal/convergence/formula.Formula`
  retirement is described as an end state ("deleted or unexported preview
  DTO with no TOML decode methods", L5645) and a temporary delegator
  (L5646) but is **not pinned to a specific sub-phase**. Today
  `internal/convergence/formula.go` has zero production callers (grep
  finds only test callers); leaving the symbols exported with a
  "temporary delegator" through 4b–4f reopens the exact drift this
  design is supposed to close, because future PRs can reintroduce a
  raw-field read without tripping a phase boundary. The static guard
  `TestNoConvergenceSubsetParserUse` (L2221, L5651–5655) is described
  but its precise predicate ("imports parser packages directly, opens
  formula source files, scans raw `[requires]`, reads legacy `contract`,
  constructs `convergence.Formula{}` from TOML, reconstructs required
  vars outside the projection, or branches on root metadata") needs to
  be backed by a concrete check (import allowlist + AST literal scan +
  named-symbol deny list) and bound to a phase that fires *before* any
  new convergence callers are added. Without that, the design's
  "no convergence-specific exemption" invariant is enforced only by
  convention.
- [Major] Identifier-validation handoff for `required_vars` is not
  spelled out. The current `convergence.ValidateVarKey`
  (`internal/convergence/formula.go:88`) rejects keys that are not
  Go-identifier shaped — empty, leading digit, hyphens, non-ASCII, etc.
  The design delegates `required_vars` to the compiler projection
  (`CompiledRuntimeVar` in L5684 / L5699) but does not assert that
  compiler-side var-key validation preserves that rejection set. If the
  canonical compiler accepts `my-var` or `0var` where convergence used
  to fail, the projection silently widens a contract that callers rely
  on for prompt template safety. This needs an explicit fixture row
  ("identifier-shaped required_var keys; non-identifier keys produce
  `formula.runtime_var_invalid_identifier` before any convergence
  domain validation runs") and the diagnostic ordering needs to place
  it ahead of `ConvergenceRuntimeInputs` checks.
- [Minor] The deterministic diagnostic order table (L251–260) stops at
  step 7 (host capability) but the executable convergence flow
  (L5565–5588) interleaves four post-acceptance domain validators
  (`ValidateAcceptedArtifact`, `ProjectAcceptedFormula`,
  `ValidateProjection`, durable writer). The order table should be
  extended with steps 8–10 covering accepted-artifact validation, typed
  projection, and convergence-domain projection validation so that
  reviewers and implementers cannot rearrange convergence-domain
  diagnostics (reserved step id collision, evaluate-prompt readability,
  retry-shape conflict) ahead of host-capability or `[requires]`
  diagnostics. As written, two equally-correct readings of the spec are
  possible.
- [Minor] The design does not anchor an invariant against future
  convergence-only TOML keys. The current
  `internal/convergence/formula.go` enumerates four convergence-visible
  fields (`Convergence`, `RequiredVars`, `EvaluatePrompt`, `StepNames`)
  and the design extends this with `EvaluatePromptPath`, `RelevantStep`,
  and `Retry` (L5682–5692). A future contributor adding a
  convergence-only key (say `convergence.gate_strategy`) under the
  canonical parser's recognition set is fine; the same key parsed only
  by convergence is the drift this design is meant to prevent. A single
  load-bearing sentence — "no convergence-only TOML key may exist
  outside the canonical compiler's recognition set; convergence-domain
  validators consume only `CompiledConvergenceProjection` plus runtime
  inputs" — would make the invariant testable.

**Missing evidence:**
- No named fixture for the exact red-flag scenario in this lane:
  "convergence create against `formula = X` with
  `[requires] formula_compiler = ">=2"` on a host with
  `formula_v2=false` produces zero convergence/iteration/hook/dependency/
  convoy/retry/fanout/artifact-ref writes and emits one
  `convergence.formula_compile_failed` event with the diagnostic
  `OnceKey` and `producer=convergence_create`". The design references
  the contract repeatedly (L5476–5489) but never names the test.
- No named fixture asserting that an active convergence loop with a
  persisted v2 artifact continues iterating after `formula_v2` toggles
  to false at the same formula identity, and stops creating new
  iterations the moment the selected formula identity changes. The
  contract is stated (L5415, L5479–5480) but the corresponding test row
  is missing from the convergence column of the fixture matrices
  (L2219–2240, L5204–5210).
- No statement about which phase of the migration retires
  `internal/convergence/formula_test.go` or replaces it with
  projection-driven tests. Because the production callers are already
  zero (grep finds only test callers of `ValidateForConvergence`),
  leaving these tests in place quietly preserves a "this is the
  canonical way to validate a convergence formula" signal for
  contributors reading the package — exactly the cultural drift the
  ownership-migration table is trying to stop.
- No explicit decision about whether `ValidateForConvergence`'s implicit
  acceptance of `cityPath == "" && readFile == nil`
  (`internal/convergence/formula.go:42`) survives projection. The
  current code silently *skips* evaluate-prompt validation in tests; the
  design says readability is convergence-domain validation *after*
  acceptance (L5459), with `ValidateProjection` responsible for prompt
  path safety. The handoff for the "no city context" case needs a
  fixture: either projection always rejects, or the writer-side
  preflight asserts a non-empty cityPath before convergence-domain
  validators run.

**Required changes:**
1. Pin `internal/convergence/formula.go` retirement to a specific
   sub-phase. The migration table (L5641–5649) should add a column for
   "deleted / unexported by", and pick a phase (the natural fit is the
   end of sub-phase 4f, after all live callers move to
   `ProjectAcceptedFormula`). Until that phase, the file must carry a
   build-tag-guarded `//go:build legacy_convergence_subset` or be moved
   to `internal/convergence/legacy/` so a missing-import deny test is
   trivially expressible. Add the deny-test row to L2108 or L5651.
2. Add a row to the projection rules (L5694–5704) and the parser
   matrix (L261–L277, L1088 region) asserting that required_var key
   validation preserves the Go-identifier subset, with named fixtures
   for empty, leading-digit, hyphen-containing, and non-ASCII keys.
   Pin the diagnostic code to `internal/formula` (e.g.,
   `formula.runtime_var_invalid_identifier`) and forbid convergence
   from emitting an equivalent diagnostic after projection.
3. Extend the deterministic diagnostic-order table (L251–260) with the
   post-acceptance steps used by the executable flow (L5565–5588) so
   that convergence-domain diagnostics cannot be reordered ahead of
   host-capability or `[requires]` diagnostics. Add a fixture row in
   the L1003-region matrices for "convergence formula with reserved
   `evaluate` step *and* `[requires] formula_compiler = ">=2"` on
   `formula_v2=false`": only the host-capability diagnostic should be
   visible; convergence-domain diagnostics suppressed until acceptance.
4. Add the load-bearing invariant: "No convergence-only TOML key may
   exist outside `internal/formula`'s recognition set; convergence
   domain validators consume only `CompiledConvergenceProjection` plus
   `ConvergenceRuntimeInputs`." Anchor it to a new test row in
   `TestNoConvergenceSubsetParserUse` or a sibling test that scans
   `internal/convergence` for TOML decode / key-presence APIs.
5. Add the two missing named fixtures (host-disabled zero-write +
   active-loop persisted-artifact continuation) to the fixture matrices
   at L5206 and L5329, and route them through both
   `cmd/gc/convergence_store.go:pourWisp` and the
   `convergence.CreateHandler` entry point in
   `internal/convergence/create.go` so both legs of the live path are
   covered.

**Questions:**
- Does the design intend `internal/convergence/formula.Formula` to be
  deleted, or retained as an unexported `previewFormula` DTO? "deleted
  or unexported preview DTO" (L5645) reads as deferred to implementation,
  but the two end states have materially different drift surfaces — an
  unexported preview DTO can still grow a `Requires` field that nothing
  honors. If the intent is "deleted", say so explicitly.
- Is there a single `OnceKey` discriminator that separates "convergence
  preflight failed because `[requires]` is unsatisfied" from
  "convergence preflight failed because of a domain rule"? Today the
  proposed key shape includes `producer=convergence_create`, which is
  shared. If two diagnostics from the same convergence create are
  intentional (one formula-domain, one convergence-domain), the burst
  budget (L3685, 15-minute window) should document the expected
  cardinality so dashboards do not double-count.
- Does `gc.convergence_waiting_reason=formula_requirement` (L5613) have
  a sibling for convergence-domain failures (missing evaluate prompt,
  reserved step collision) that survive acceptance? The current text
  treats `formula_requirement` as the only waiting reason; if the
  convergence-domain diagnostic path can also park a loop, it needs its
  own metadata key and clearing rule, or a documented decision that
  convergence-domain failures only block create/retry and never an
  already-active loop.
- The design anchors to `cmd/gc/convergence_store.go:pourWisp ->
  molecule.Cook -> internal/formula` (L5422–5424). Today `molecule.Cook`
  does not appear to consult `HostCapabilities` or `[requires]` before
  writing the root and child beads. Is the first migration step adding
  a preflight inside `molecule.Cook`, or inserting a preflight in
  `convergence_store.go:pourWisp` before the `molecule.Cook` call? The
  table at L5498 says preflight runs *before* the durable writer, but
  the exact implementation seam for that insertion is not pinned, which
  affects whether the convergence path inherits the same gate as
  sling/order/fanout (`molecule.Cook` is shared) or gets its own
  parallel pre-Cook wrapper (which reintroduces drift risk).
