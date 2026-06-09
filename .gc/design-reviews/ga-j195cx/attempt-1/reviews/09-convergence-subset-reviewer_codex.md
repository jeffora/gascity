# 09 Convergence Subset Reviewer (Codex)

**Persona:** Felix Berger  
**Verdict:** block

The design now has the right high-level direction for convergence: convergence must not own a formula subset compiler, `[requires]` must be enforced by `internal/formula`, and durable convergence writes must flow through accepted artifacts. The remaining problems are at the migration boundary. The document still leaves enough ambiguity for an implementer to preserve the current convergence subset path or to add the accepted-artifact check too late in the current call graph.

## Findings

### Major: Convergence preflight is not bound to the current create/retry API boundary

The design requires convergence `create` to run `CompileWithResult -> AcceptCompileResult -> ValidateAcceptedArtifact -> ProjectAcceptedFormula -> ValidateProjection` before the first convergence root write, and lists `convergence root` as a forbidden write on failure (`.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:5493`). It also says `PourWisp` must call the preflight helper (`.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:5509`).

That is not enough against the live API. Today `CreateHandler` creates the convergence bead and writes state, formula, gate, city path, evaluate prompt, and vars before calling `Store.PourWisp` (`internal/convergence/create.go:59`, `internal/convergence/create.go:82`, `internal/convergence/create.go:101`, `internal/convergence/create.go:108`). The current `convergence.Store` interface exposes `CreateConvergenceBead`, `SetMetadata`, `PourWisp`, and `PourSpeculativeWisp`, but has no compile/preflight artifact boundary (`internal/convergence/handler.go:51`).

If the implementation follows the design literally by putting the helper at the start of `PourWisp`, a disabled host or malformed `[requires]` still leaves a convergence root and metadata behind. That violates the design's own zero-root-write contract and keeps convergence as a bypass relative to sling/order/fanout.

Required change: name the concrete signature/API changes for `CreateParams`, `CreateHandler`, `RetryHandler`, `convergence.Store`, `convergenceStoreAdapter`, and `cmd/gc/convergence_tick.go` so accepted-artifact preflight happens before `CreateConvergenceBead` and before retry root creation. Add zero-write fixtures that snapshot root beads, metadata, vars, active index changes, hooks, dependencies, convoys, and artifact refs for create, retry, speculative pour, and fallback pour.

### Major: Evaluate-prompt override is not modeled as artifact identity

The design says `evaluate_prompt` is part of artifact identity and hashes inline content or prompt paths (`.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:5599`). It also projects `EvaluatePrompt` and `EvaluatePromptPath` from the compiler (`.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:5682`).

The live system has more than one evaluate-prompt surface. `gc converge create --evaluate-prompt` is sent as request param `evaluate_prompt` (`cmd/gc/cmd_converge.go:61`) and described as an override (`cmd/gc/cmd_converge.go:115`). The controller passes that into `CreateParams.EvaluatePrompt` (`cmd/gc/convergence_tick.go:216`), `CreateHandler` persists it as convergence metadata (`internal/convergence/create.go:82`), and `pourWisp` passes it as a compile var named `evaluate_prompt` (`cmd/gc/convergence_store.go:169`). Separately, `ResolveEvaluateStep` treats `Formula.EvaluatePrompt` as a formula-declared prompt path (`internal/convergence/evaluate.go:31`).

The design does not decide whether the operator override is a path, inline content, compile var, runtime input, or formula source override. It also does not state how that override is source-attributed, path-normalized, symlink/path-escape checked, copied into an accepted projection, included in `VarsHash` versus `EvaluatePromptHash`, or handled when retry copies metadata from an old root.

Required change: introduce an explicit `ConvergenceEvaluatePromptInput` or equivalent in `ConvergenceRuntimeInputs` and the accepted write intent. Fixture-lock default prompt, formula-authored prompt path, CLI override, metadata retry copy, changed override after root creation, unreadable override, path escape, symlink, same-identity host-downgrade reuse, and changed-identity recompile. Without this, prompt changes can silently escape the accepted-artifact identity.

### Major: The convergence authoring schema is still not pinned

The design correctly says `internal/formula` owns convergence projection and `internal/convergence` must not read raw `[requires]`, `contract`, required vars, or root metadata (`.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:5627`, `.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:5651`). It defines `CompiledConvergenceProjection` with `Enabled`, `RequiredVars`, `EvaluatePrompt`, `EvaluatePromptPath`, `RelevantStep`, and `Retry` (`.gc/design-reviews/ga-j195cx/attempt-110/design-before.md:5682`).

But the design never locks the authored TOML shape for those fields. Current docs still teach a convergence-specific top-level subset owned by `internal/convergence/formula.go`, with `convergence`, `required_vars`, and `evaluate_prompt` (`docs/reference/formula.md:77`). The current `internal/formula.Formula` type has `Contract` and formula graph fields, but no convergence fields (`internal/formula/types.go:62`).

That leaves two unsafe implementation paths open: keep `internal/convergence/formula.go` as the real parser and only wrap it, or add ungoverned top-level fields to `internal/formula` without a raw-shape matrix. Either preserves the formula subset drift this review lane is meant to catch.

Required change: decide whether the current top-level `convergence`, `required_vars`, and `evaluate_prompt` keys are canonical compiler-projected domain fields, deprecated aliases, or replaced by a namespaced convergence table. Add raw scanner/type/unknown-key rows, source-position fixtures, inheritance/expansion/aspect rules, docs stale-guidance, and generated projection equivalence rows for the chosen surface.

### Minor: The projection parity fixture needs caller-path coverage, not only field parity

`formula_projection_equivalence.yaml` is a good field-level guard, but the risky convergence failures are caller-specific: create writes a root, retry creates a replacement root, speculative pour creates a hidden wisp, pending-wisp adoption mutates hook/convoy state, and burn/repair paths can delete or close children. The design lists these entry points, but the projection equivalence sample is one row per projection field.

Required change: require each convergence caller-path row to point to both a projection field row and a zero-write fixture. That keeps a future projection addition from being "covered" while one convergence entry point still reconstructs the field from metadata or raw TOML.

## Missing Evidence

- A before/after call graph for `gc converge create`, controller create request handling, `CreateHandler`, `RetryHandler`, `Store.PourWisp`, `PourSpeculativeWisp`, fallback pour, pending adoption, burn, missing-child repair, and manual iterate.
- A concrete API sketch showing how host capabilities, resolved formula source, compile vars, evaluate-prompt override, accepted artifact, write intent, and diagnostic subject flow before root creation.
- A convergence raw-shape matrix for `convergence`, `required_vars`, `evaluate_prompt`, prompt path/content variants, unknown convergence fields, duplicate fields, and source-position attribution.
- Zero-write fixtures that prove fatal formula diagnostics leave no convergence root for create and no retry/speculative/fallback state for existing loops.

## Required Changes

1. Bind the convergence preflight helper to the current create/retry API before `CreateConvergenceBead`, not just to `PourWisp`.
2. Model operator evaluate-prompt overrides as part of accepted artifact identity and projection validation.
3. Decide and fixture-lock the convergence authoring schema so `internal/formula` can own projection without recreating a subset parser.
4. Cross-link projection parity rows to every convergence durable caller and its zero-write fixture.

