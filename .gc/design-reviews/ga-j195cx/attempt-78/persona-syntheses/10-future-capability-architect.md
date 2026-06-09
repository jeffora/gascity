# Ibrahim Park

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex (Kimi 2.6 absent)

**Consensus findings:**
- [Major] Default-capability and future-axis behavior is directionally correct but needs stronger persisted-state coverage. Both reviewers support omitted/empty/`>=1` normalizing to compiler capability 1 and unknown `[requires]` axes failing closed, but they require explicit fixtures for old readers encountering default compiler capability plus unsupported future axes.
- [Major] Future axes must stay tied to typed, owner-validated grammar rather than becoming a raw string escape hatch. Claude flags the global "flat scalar string" rule as potentially over-constraining future axes, while Codex flags `AxisMinimum.Minimum` as a possible second grammar source unless it is parsed by the owning axis parser or replaced by typed minima.
- [Major] Old-reader safety should cover more than graph-specific writes. A persisted root with `formula_compiler = ">=1"` plus a future axis such as `state_store`, `tools`, or `runtime` must remain visible but produce typed upgrade/remediation diagnostics and must not perform durable writes whose safety could depend on the unknown axis.
- [Minor] Multi-axis diagnostic behavior is under-specified. Both reviews ask for matrix or precedence coverage when more than one future-facing condition is present, including a host satisfying one axis while failing or not recognizing another.
- [Minor] The current grammar leaves some future-extension details ambiguous. Claude asks either to reserve or permanently reject non-`>=` operators, and Codex asks whether unknown-axis diagnostics should precede unsupported future compiler minima or whether mixed future cases should report multiple diagnostics.

**Disagreements:**
- Default-capability artifact identity: Claude says omitted `[requires]`, empty `[requires]`, and explicit `formula_compiler = ">=1"` may still diverge through `ContentHash`, `compileID`, accepted artifact refs, and root metadata unless identity normalization is explicitly specified. Codex reads the design as already requiring identity/runtime fixtures for those spellings. Assessment: treat this as a real review signal and require an explicit identity policy plus a fixture; behavioral equivalence is not enough if artifact identity remains source-byte-sensitive.
- Future-axis shape: Claude challenges the blanket scalar-string restriction because plausible axes may naturally need lists or structured values. Codex does not object to scalar axis values generally, but objects to registry minima remaining untyped. Assessment: the minimum required change is to validate every registry minimum through the owning parser; the design should also either justify scalar-only axes with old-reader fail-closed reasoning or define a registered typed-shape envelope for future non-scalar axes.
- Operator reservation: Claude wants reserved-operator diagnostics or a permanent rule that all non-`>=` operators are invalid. Codex does not raise this as a required change. Assessment: this is not blocking, but a short permanence/reservation statement now would prevent future grammar drift.
- Metadata shape: Claude questions empty `gc.formula_requirement_axes` handling and whether the five-key forward-compat metadata block advances independently or in lock-step. Codex does not flag this. Assessment: keep as missing evidence unless the design already has no-requirements manifest fixtures elsewhere.

**Missing evidence:**
- Kimi 2.6 review was not present in the attempt's `reviews/` directory.
- No fixture was cited for old readers observing `formula_compiler = ">=1"` with an unknown future sibling axis across list/API/dashboard observation, dispatch, retry, fanout, convergence, write attempts, and cleanup.
- No registry self-test was cited proving every `AxisMinimum.Minimum` is accepted by the owning axis grammar and maps to the same typed normalized value used by `NormalizedRequirements`.
- No explicit fixture was cited for the chosen artifact identity policy across omitted `[requires]`, empty `[requires]`, and explicit default capability.
- No fixture was cited for an empty requirement-axis manifest, nor for multi-axis diagnostic ordering when one axis is satisfied and another is unsupported or unknown.

**Required changes:**
- Add explicit forward-compatibility rows and fixtures for default-capability roots that carry future axes. Expected behavior should be visibility plus typed upgrade/remediation diagnostics and zero durable writes for operations whose safety could depend on the unknown axis.
- Choose and document whether `ContentHash`, `compileID`, accepted artifact refs, `gc.formula_compile_artifact`, retry-run reuse, and convergence reuse normalize omitted/empty/`>=1` or intentionally distinguish source spelling; add a fixture that locks the decision.
- Either change `AxisMinimum` to store typed per-axis minima or add CI that parses every registry minimum through the owning axis parser and rejects any raw string outside that axis grammar.
- Justify the scalar-string axis constraint with concrete old-reader fail-closed reasoning, or relax it into an explicit typed-shape envelope for future axes.
- Add multi-axis combined-defect coverage for `formula_compiler = ">=2"` plus a future axis such as `state_store = ">=2"` against hosts that satisfy only one axis. Specify whether diagnostics are per-axis, combined, or precedence-ordered.
- State whether non-`>=` operators are permanently invalid for `formula_compiler` or reserved for future diagnostics, and specify how mixed `unknown_axis` plus `unsupported_future_minimum` cases should report upgrade guidance.
