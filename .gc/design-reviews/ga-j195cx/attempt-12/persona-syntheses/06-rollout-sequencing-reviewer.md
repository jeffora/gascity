# Lena Driscoll

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] Phase 3 is still too broad to satisfy the rollout claim that every step can land independently while keeping main green. Both reviews flag that sling, orders, API, convoy, convergence, molecule, dashboard projections, tests, and the static no-raw-consumer guard are grouped into one phase. As written, the guard can only land after all behavioral callers migrate, creating a large flag-day PR or an undefined multi-PR sequence.
- [Major] Compatibility for dual-declared first-party formulas is asserted but not fully executable. Both reviews require a concrete old-supported-binary and `bd` shell-out/fallback validation gate before source edits depend on `[requires]`, plus an explicit owner and artifact for proving parity or removing the fallback.
- [Major] The first-party requires-only conversion lacks a sufficiently enforceable release floor and rollback story. The design says a minimum binary floor must exist before Phase 6, but does not define the authoritative metadata field, CI enforcement, or a one-step rollback path for restoring removed `contract = "graph.v2"` declarations across packs, examples, fixtures, and tutorials.
- [Minor] Documentation, examples, and executable formula migration need clearer separation. Both reviews note that prose can teach `[requires]` after parser and caller support, but runnable examples and first-party formula sources must remain dual-declared until compatibility gates and the minimum binary floor are enforced.
- [Minor] Dual-stamping and alias-window exit criteria are mostly measurable, but the design does not name the phase where `gc.formula_contract` metadata stamping ends or the tests that prove alias-window dual-stamping and post-alias-removal canonical-only behavior.

**Disagreements:**
- Claude treats Phase 6 rollback as a separate major risk requiring either per-pack sub-phases or a revert command. Codex emphasizes the earlier gate before Phase 4/Phase 5 executable formula edits. These are compatible concerns: the rollout plan should add both an early compatibility gate for dual-declared edits and a later explicit rollback plan for requires-only conversion.
- Claude identifies additional missing test coverage around mixed-version controllers, in-flight wisps/convergence, metadata-key fallback during Phase 3, and Phase 1 validation-matrix scope. Codex does not dispute these; it simply focuses on release sequencing. I would keep Claude's extra test requirements because they directly exercise the migration window this persona is reviewing.
- Codex asks whether `GC_NATIVE_FORMULA=false` remains the intended rollback switch for caller migration. Claude frames the same area as an unanchored `bd` shell-out parity dependency. The synthesis is that the design must explicitly state whether the fallback remains supported, is proven compatible, or is removed before source conversion proceeds.

**Missing evidence:**
- No per-surface Phase 3 landing order showing which caller migrates first, what test gates each slice, and how each slice rolls back independently.
- No executable compatibility plan proving the previous supported Gas City binary and the configured `bd` shell-out fallback can parse all dual-declared first-party graph formulas.
- No authoritative location for the minimum supported binary floor, and no CI or release gate that blocks requires-only first-party formula conversion when that floor is absent.
- No phase or test that ends `gc.formula_contract` dual-stamping, nor tests proving dual-stamping during the alias window and canonical-only metadata after alias removal.
- No tested mixed-version-controller scenario for roots with dual-stamped metadata.
- No clear Phase 1 scope boundary for validation-matrix rows, diagnostic ordering, and the v2-only construct registry scan.
- No Gemini review was present for this persona.

**Required changes:**
- Split Phase 3 into named reversible sub-phases: shared normalized result plumbing, sling/CLI, orders, API, convoy, convergence/molecule execution, dashboard projections, and finally the static no-raw-consumer guard. For each sub-phase, document the required tests and rollback mechanism.
- Add a release gate before executable first-party formula edits that validates every dual-declared graph formula against the previous supported Gas City binary and the configured `bd` shell-out fallback, or explicitly blocks source conversion until that fallback is removed.
- Promote the minimum binary floor from release-checklist prose to an enforceable contract: define the pack/config metadata field, the release artifact that records it, and the CI test that fails Phase 6 changes without it.
- Add a Phase 6 rollback procedure: either split requires-only conversion by pack/artifact class or ship a tested revert command/script that restores dual declarations in one step.
- Separate documentation prose, runnable examples, dual-declared first-party formulas, and requires-only first-party formulas into distinct rollout gates.
- Name the phase that stops `gc.formula_contract` dual-stamping and add tests for alias-window dual-stamping, post-alias-removal canonical-only metadata, and mixed-version readers.
- Anchor the `bd` shell-out parity dependency to the relevant formula-migration phase or fallback-removal milestone.
- Tighten Phase 1 by listing which validation-matrix rows and diagnostic-order stages land there, including whether the v2-only construct registry scan is required in Phase 1.
