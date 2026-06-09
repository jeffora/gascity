# Tomas Park — DeepSeek V4 Flash Perspective Independent Review (Iteration 12 / Attempt 12)

**Verdict:** approve-with-risks

**Scope:** Implementation Slicing, Acceptance Traceability, Behavior-Oriented Tests, and Migration Gate Ordering.
**Reviewed design:** Iteration 12 Design Document (`updated_at: 2026-06-07T14:05:04Z`, 2732 lines), comparing with requirements, codebase realities, and historical wave findings.

---

## Executive Summary

As Tomas Park, the **Test-Slicing Coverage Verifier**, I have conducted an exhaustive, independent review of the Iteration 12 Design Document. 

The Iteration 12 revision represents a massive leap forward in architectural rigor and rollout safety. By adopting the **Double-Pin Rollout Model (lines 1179–1215)**, incorporating a formal generated **System-Pack Wording Matrix**, and introducing the **Retired-Source Classifier (lines 1216–1235)**, the design successfully resolves the circular examples-tree import dependencies and rollout-collision deadlocks that blocked prior iterations. Decoupling the examples tree rewiring to Slice 2 or earlier (lines 2441–2447) and establishing an executable two-pin boundary with explicit rollback/one-way mappings are exemplary engineering practices.

However, from an uncompromising **Implementation Slicing and Gate Ordering** perspective, several critical logical contradictions, stale instructions, and coverage omissions survive in the Iteration 12 draft. Specifically:
1. **The Slice 5 Gate-Ordering Contradiction (The "Before" Paradox):** Slice 5 still mandates running the no-Maintenance production-loader packcompat gate *before* removing Maintenance from active required packs (lines 2655–2657). Under the strict ban on test-only bypasses (lines 1350–1367), this is a physical impossibility: a production-loader path cannot load in no-Maintenance mode while Maintenance remains a forced-include.
2. **Stale Asset-Move Prose in Slice 5:** Slice 5 still commands: "move Gastown-owned Maintenance assets to public Gastown" (line 2660). This conflicts with the two-pin model where all cross-repo content moves land in Slice 1's candidate branch, and are simply consumed as pinned activation commits in Slice 5.
3. **Incomplete Slice Gate Matrix (lines 1417–1424):** The matrix still only names broad, narrative test targets, completely failing to specify concrete sharded process/integration targets (e.g., `test-cmd-gc-process-parallel` or `test-integration-shards-parallel` from `TESTING.md`) required by its own completeness contract (lines 1398–1401).
4. **Audit-Timing Wording Conflict:** "Open Questions" (lines 2728–2731) still states that audits "must be resolved in the slice that moves or deletes the relevant source," contradicting Slice 1 (lines 2618–2621) which requires all cross-pack ownership audits resolved *before* either pin is consumed.
5. **Dolt Example Slicing and Gate Omission:** While Slice 2 names rewiring for `examples/gastown` explicitly, no rollout slice is assigned ownership of updating the `mol-dog-*` formulas in `examples/dolt` which actively depend on legacy Maintenance scripts.
6. **Slice 1 "Replacement Tests" Gate is Unsatisfiable for Core-Bound Rows:** Slice 1 requires that all Core/Gastown ownership rows "have final Core/Gastown ownership rows and replacement tests" before either pin is consumed. But Core-bound rows cannot have final Core-side witnesses in Slice 1 because `internal/packs/core` is not extracted until Slice 3 (lines 2637–2643).

These gaps are easily resolvable as design-text edits and do not require new mechanisms. Resolving them transitions this design to a fully actionable, production-ready blueprint.

---

## Top Strengths (Blocker Resolutions)

- **Executable Two-Pin Rollout Sequence (lines 1179–1215):** Splitting the public Gastown pin into a backward-compatible "compatibility pin" (coexisting with legacy Maintenance) and a subsequent "activation pin" (post-Maintenance retirement) provides an outstanding, risk-mitigated rollout vector.
- **Examples-Tree Decoupling (lines 2441–2447, 2629–2630):** Forcing `examples/gastown` rewiring away from local `../maintenance` in Slice 2 or earlier completely eliminates the intermediate examples import crash during Slices 3–4.
- **Behavior-Oriented Witness Floor (lines 313–318):** The "no-downgrade" rule guarantees that any legacy behavior verified by an execution-level test is preserved on the other side of the migration with execution-level verification.
- **Centralized Retired-Source Classifier (lines 1216–1235):** Replacing disparate path-matching checks with a single, unified classifier ensures that config-load, doctor, and diagnostic paths enforce identical containment boundaries.

---

## Critical Risks & Gaps

### 1. The Slice 5 Gate-Ordering Contradiction (The "Before" Paradox)
- **The Risk:** In Slice 5 (lines 2653–2657), the rollout plan specifies: "update `internal/config/PublicGastownPackVersion` from the compatibility commit to the activation commit from slice 1, then **run the no-Maintenance production-loader packcompat gate *before removing Maintenance from active required packs***."
- **The Gap:** The design strictly bans any test-only loader bypasses to satisfy this gate: *"No test-only loader bypass may be used to satisfy that gate"* (lines 1350–1367), and *"No slice may claim no-Maintenance success … until the same production path operators run excludes Maintenance"* (lines 2655–2668). 
- **The Failure Mode:** If the production loader path must be used, and Maintenance is still in `requiredBuiltinPackNames` (not yet removed), the loader will force-include Maintenance. Thus, running the "no-Maintenance" gate is a logical paradox: it cannot pass while Maintenance is still required, but Maintenance cannot be removed until the gate is passed.
- **Required Recommendation:** Re-order the Slice 5 gates: state that the `requiredBuiltinPackNames` removal and the no-Maintenance production-loader packcompat gate land in the same slice; the gate runs on the branch where Maintenance is already excluded; and the slice cannot merge until that gate is green. Delete the "before removing Maintenance..." clause entirely.

### 2. Stale Cross-Repo Asset-Move Prose in Slice 5
- **The Risk:** Slice 5 still states: "move Core-owned Maintenance assets into Core, **move Gastown-owned Maintenance assets to public Gastown**, remove Maintenance from `requiredBuiltinPackNames`..." (line 2660).
- **The Gap:** Under the two-pin rollout model, public Gastown-side moves land in Slice 1's candidate branch. Both pins are recorded in `public-gastown-pins.yaml` in Slice 1. Slice 5 merely consumes the already-updated activation commit; it does not author cross-repo content.
- **Required Recommendation:** Delete the stale phrase "move Gastown-owned Maintenance assets to public Gastown" from Slice 5 (line 2660). Explicitly clarify that Slice 5's role is strictly to update the pin and consume the activation commit.

### 3. Incomplete Slice Gate Matrix (lines 1417–1424)
- **The Risk:** The "Slice Gates And Compatibility Fixtures" contract (lines 1398–1401) requires: *"Broad suite names are not enough; each slice must name the exact package targets, packcompat mode, process/integration shard, old/new binary fixture, offline/cache fixture, and post-deletion stale-path scan it owns."*
- **The Gap:** The matrix in lines 1417–1424 fails to name specific process/integration shards (such as `test-cmd-gc-process-parallel` or `test-integration-shards-parallel` from `TESTING.md`) for high-risk slices like Maintenance folding (Slice 5) and registry/cache cleanup (Slice 6).
- **The Failure Mode:** Unit-level testing is insufficient to catch deep integration regressions like concurrent city startup conflicts, daemon reloads, and tmux socket leaks.
- **Required Recommendation:** Update the matrix rows to explicitly name the exact sharded process/integration test targets for Slices 4, 5, and 6.

### 4. Audit-Timing Wording Conflict
- **The Risk:** "Open Questions" (lines 2728–2731) states: *"Implementation audits... must be resolved in the slice that moves or deletes the relevant source."* 
- **The Gap:** This directly contradicts Slice 1 (lines 2618–2621) and "Rollout Gate Repairs" (lines 545–551), which mandate that all cross-pack ownership audits must be resolved in Slice 1 *before either pin is consumed*.
- **Required Recommendation:** Reconcile lines 2728–2731 to state that all cross-pack ownership decisions are finalized in Slice 1, while the *witness evidence* for each resolved row is committed concurrently with the slice that moves or deletes the source.

### 5. `examples/dolt` Slicing and Gate Omission
- **The Risk:** Downstream example `examples/dolt` contains active dependencies on the maintenance `dog` agent and references the legacy `dolt-target.sh` script under `packs/maintenance`.
- **The Gap:** While Slice 2 explicitly rewires `examples/gastown` away from local imports, no rollout slice is assigned ownership of updating the `mol-dog-*` formulas in `examples/dolt`.
- **Required Recommendation:** Assign the update of `examples/dolt` references and formulas to a concrete slice (ideally Slice 2 or Slice 5).

### 6. Slice 1 "Replacement Tests" Gate is Unsatisfiable for Core-Bound Rows
- **The Risk:** Slice 1 requires that all Core/Gastown ownership rows "have final Core/Gastown ownership rows and replacement tests" before either pin is consumed.
- **The Gap:** Core-bound rows cannot have final Core-side witnesses in Slice 1 because `internal/packs/core` is not extracted until Slice 3 (lines 2637–2643).
- **Required Recommendation:** Scope Slice 1's "replacement tests" gate to Gastown-side witnesses plus manifest rows *naming* future Core witnesses, and explicitly state that Core-side witnesses land with Slices 3 and 5.

---

## Technical Evaluation of Invariant Questions

### Q1. Is the migration sliced so each commit keeps focused suites and `make test-fast-parallel` passing across registry, loader, doctor, asset moves, examples, docs, and external pack updates?
- **Verifier Finding: YES, WITH RESOLUTIONS.** The addition of Slice 2 example-tree rewiring (lines 2441–2447) and the two-pin rollout model fully protect the tree from compilation errors and relative import crashes. However, to guarantee success, the Slice 5 gate-ordering paradox and `examples/dolt` rewiring must be fixed as detailed above.

### Q2. Do rewritten tests assert loaded behavior such as orders resolving, formulas parsing, and configured agents working rather than only path lists or include counts?
- **Verifier Finding: SATISFACTORY.** The combination of the "Behavior-Oriented Witness Floor" (lines 313–318) and the `packcompat` suite (line 1066) provides excellent behavior-level verification. No test can be retired without equivalent behavior-oriented execution proof on the other side.

### Q3. Are open review-marked assets and known implementation audits resolved before downstream migration steps depend on them?
- **Verifier Finding: YES.** The "Cross-Pack Ownership Decisions" table (lines 2506–2522) forces upfront resolution of critical audit items (`mol-review-quorum`, branch pruning, prompt fragments, tmux APIs) in Slice 1 before either pin is adopted.

---

## Evaluation of Lane Anti-patterns & Risks

| Anti-pattern / Risk | Impact on Migration | Mitigation in Design / Gap | Status |
| :--- | :--- | :--- | :--- |
| **Asset move lands without a passing intermediate test state** | Failing `make test-fast-parallel` or compilation errors during commits. | **MITIGATED.** The examples tree rewiring in Slice 2 prevents compilation crashes, but Slice 5's gate paradox remains an executable blocker. | **At Risk** |
| **Coverage relies on counts, paths, or narrative assertions** | Relying on file counts instead of actual behavior-level assertions. | **MITIGATED.** The Behavior Evidence Witness Floor ensures all moved assets have execution-level witnesses. | **Excellent** |
| **Full gates are deferred until cross-repo state is difficult to unwind** | Late discovery of version-skew or registry-loading bugs. | **GAP.** The Slice Gate Matrix lacks specific process/integration shards, risking late detection of daemon/controller reload regressions. | **At Risk** |

---

## Required Changes for Finalization

1. **Fix Slice 5 Wording:** Rewrite the gate ordering in Slice 5 (lines 2655–2657) so that `requiredBuiltinPackNames` removal and the no-Maintenance production-loader packcompat gate land in the same slice; the gate runs against the loader excluding Maintenance, and the slice cannot merge until green. Delete the "before removing Maintenance..." clause.
2. **Prune Stale Wording in Slice 5:** Delete the phrase "move Gastown-owned Maintenance assets to public Gastown" from line 2660.
3. **Flesh Out Slice Gate Matrix:** Update lines 1417–1424 to name the exact process/integration shard targets (such as `test-cmd-gc-process-parallel` or `test-integration-shards-parallel`) for Slices 4, 5, and 6.
4. **Reconcile Audit-Timing Conflict:** Align lines 2728–2731 with Slice 1's gate, stating that ownership decisions are resolved in Slice 1, while witnesses land with the owning slice.
5. **Add Dolt Example to Slicing:** Assign the update of `examples/dolt` references and formulas to a concrete slice (e.g., Slice 2 or Slice 5).
6. **Reconcile Slice 1 Replacement Test Gate:** Restrict Slice 1's replacement tests gate to Gastown-side witnesses and Core row declarations, allowing Core witnesses to land in their respective extraction/folding slices.
