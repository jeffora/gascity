# Lena Driscoll - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design now separates parser/model support, canonical metadata, caller migration, source conversion, docs, requires-only conversion, and alias removal. That sequencing avoids the obvious flag day where parser, callers, docs, and formulas all have to move in one commit.
- Built-in and example formula compatibility is treated as a release gate, not a calendar promise: first-party graph formulas stay dual-declared, roots are dual-stamped, and alias removal waits for measurable legacy-contract reports and a minimum binary floor.
- In-flight behavior is explicit enough for rollback: existing graph roots keep running from persisted metadata after `[daemon] formula_v2` changes, while only new formula compiles are gated by current host capability.

**Critical risks:**
- [Major] Phase 3 is still too broad to prove the "each step leaves main working" claim. It moves sling, orders, API, convoy, convergence, molecule, dashboard projections, tests, and the static raw-consumer guard in one named phase. If dispatch regresses, the rollback note only says roots are dual-stamped and legacy predicates can be used; it does not define whether each caller can be reverted independently, whether the legacy bd/native execution path remains callable, or whether the guard lands only after the last consumer migrates.
- [Major] Dual-declared first-party pack compatibility is asserted but not backed by an executable old-binary or bd-shell-out gate in the rollout lane. The document says old binaries read `contract` and ignore `[requires]`, and it separately says `GC_NATIVE_FORMULA=false` / bd shell-out must preserve dual declarations or be proven to parse `[requires]` identically. Before any built-in or example formula gets `[requires]`, the release should prove the previous supported binary and the bd fallback path can parse every dual-declared first-party graph formula.
- [Minor] The docs/examples phase conflates prose updates with executable formula-source migration. Teaching `[requires]` as canonical is safe after parser and caller support, but runnable examples and first-party packs must be explicitly dual-declared until the minimum binary floor is enforced. Without that split, a well-intended docs/example update could publish requires-only formulas earlier than Phase 6.

**Missing evidence:**
- A per-surface Phase 3 landing order showing which caller migrates first, what tests gate that slice, and how that slice rolls back without reverting unrelated docs or formula sources.
- A compatibility test plan for "old supported binary reads dual-declared pack" and "bd shell-out fallback reads dual-declared pack", using the actual first-party formulas currently carrying `contract = "graph.v2"`.
- The exact place where the minimum supported Gas City binary floor is recorded before first-party requires-only conversion, and how pack/example publishing is blocked when that floor is not met.

**Required changes:**
- Split Phase 3 into explicit reversible substeps: shared preflight/result plumbing, then sling, orders, API, convoy, convergence, dashboard projections, and finally the static raw-consumer guard. For each substep, name the required test gate and the rollback mechanism.
- Add a release gate before Phase 4/Phase 5 executable formula edits: validate every dual-declared first-party graph formula against the previous supported Gas City binary and the configured bd shell-out fallback, or state that the fallback remains legacy-contract-only and source conversion is blocked until native-only production.
- Separate "documentation prose teaches `[requires]`" from "runnable formulas/examples add `[requires]`" and "first-party formulas become requires-only". The latter two need explicit compatibility gates; requires-only remains Phase 6 only.

**Questions:**
- Is `GC_NATIVE_FORMULA=false` still the intended rollback switch for caller migration, or has this design replaced it with a different compatibility mechanism?
- Which release artifact defines "minimum supported Gas City binary" for packs: generated config docs, pack metadata, release checklist, or all three?
