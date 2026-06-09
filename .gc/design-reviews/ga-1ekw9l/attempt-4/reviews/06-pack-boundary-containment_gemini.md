# Owen Gallagher — Gemini (Pack Boundary Containment Reviewer, Attempt 4, Independent DeepSeek V4 Flash Style)

**Verdict:** approve-with-risks

> **Lane:** Core versus Gastown ownership split, retired Maintenance source containment, active discovery classifier, no duplicate active behavior.
>
> Reviewed against the Attempt 4 design document (`.gc/design-reviews/ga-1ekw9l/attempt-4/design-before.md`, 657 lines, `updated_at: 2026-06-09T07:28:00Z`) — §"Pack Registry, Cache, And Retired Source Authority" (lines 252–286), §"Rollout And Recovery" (lines 561–647), and §"Data And State" (lines 426–486).
>
> This independent review is produced using the DeepSeek V4 Flash persona, focusing specifically on cross-document consistency, latent edge cases, and assumptions that other reviewers may accept too quickly.

---

## 1. Executive Summary

As Owen Gallagher, the **Pack Boundary Containment Reviewer**, I have performed a rigorous, first-principles independent review of the Attempt 4 implementation plan. My final determination is **Verdict: approve-with-risks**.

This iteration (Attempt 4) represents a massive leap forward in addressing previous containment vulnerabilities. Specifically, the introduction of a granular, multi-stage rollout matrix (Slices 1a–7), the explicit definition of a single `internal/packsource` classification authority returning typed states, and the inclusion of a zero-duplicate-active validation gate represent exceptionally robust architectural guardrails.

However, several subtle deadlocks, bootstrapping paradoxes, and containment escape vectors remain unaddressed in the text:
1. **The Compatibility-Pin Deadlock:** Slices 2 through 4 adopt the compatibility pin while local Maintenance is still required. If the compatibility pin contains the migrated assets, the strict `zero-duplicate-active` gate will detect identical IDs from both sources and crash *every city load* across the entire compatibility window, halting the rollout.
2. **AST-Bypassing Directory Enumeration:** The proposed scanner test rejects banned string literals like `"maintenance"`. However, active discovery code often uses generic relative directory reads (e.g., `os.ReadDir(filepath.Join(cityPath, ".gc/system/packs"))`). A loop over this directory's children will find the `maintenance` folder without containing any string literal `"maintenance"`, bypassing the scanner entirely and entering active behavior.
3. **In-Flight Session and Molecule Lifecycle:** The plan declares "Open Questions: None", but fails to define the execution outcome for in-flight molecules or running sessions started under legacy in-tree Maintenance/Gastown during the transition in Slices 2-7.
4. **The "Dirty Directory" Custom Fork Promotion Vulnerability:** Stale local directories are ignored unless classified as `retired custom/fork`. However, if an operator or OS writes non-behavioral or temporary files (e.g., `.DS_Store` or `.tmp` files) inside the stale folder, a naive directory difference check will promote it to a custom fork and re-activate legacy files.

These risks must be closed with explicit, machine-testable gates before final approval.

---

## 2. Lane-Specific Evaluation

### Q1: Is there one classifier API used before config load, install, cache lock, validation, discovery, docs lint, prompt, formula, order, script, and hook enumeration?

**Answer:** Yes, the plan establishes `internal/packsource` as the sole authority (lines 259–266). However, there is a major loophole in how this is enforced. 

The plan relies on "Scanner tests [rejecting] ... retired-source string classifiers" (line 508) to guarantee containment. While checking for literal `"maintenance"` and `"gastown"` strings works for hardcoded imports, it is completely blind to AST-bypassing directory enumeration. For example, a developer can write:
```go
dirs, _ := os.ReadDir(filepath.Join(cityPath, ".gc/system/packs"))
for _, d := range dirs {
    // Reads 'maintenance' dynamically without using the string "maintenance"
    loadPack(d.Name()) 
}
```
Because the code contains no banned literals, the string-classifier scanner will pass, but the containment boundary is violated. 

**Requirement:** The plan must explicitly mandate that the five active behavior enumerators—prompt/template loading, formula discovery, order discovery, script resolution, and hook overlay enumeration—are structurally bound to only iterate over directories returned by `systempacks.RuntimeIncludes`, or they must call `packsource.Classify` on every subdirectory before loading.

---

### Q2: Does a city running Core plus public Gastown and no Maintenance preserve Gastown-specific behavior without shared Core conditionals?

**Answer:** Yes, the proposed architecture of symbolic bindings under `[gc.bindings.*]` (lines 343–357) and the removal of hardcoded roles in Go code (lines 336–342) structurally isolate Gastown behavior. 

However, we must guard against **implicit behavior leaking** where a Core-owned asset still contains a Gastown-shaped branch under a generalized name. For example, a generic stuck-session formula in Core could retain a Gastown-specific timeout layout. 

The proposed wording scanner and terminology matrix (lines 397–413) are excellent, but they must be complemented by a **provenance-sensitive packcompat test** (lines 187–192) that validates that no Gastown-specific prompt fragment, tool, or variable is loaded or executed from Core paths when running a city with Core and no Maintenance pack.

---

### Q3: Can stale in-tree Gastown or Maintenance directories remain on disk without entering active prompts, formulas, orders, scripts, hooks, or rollback states?

**Answer:** Yes, by ignoring stale `.gc/system/packs/maintenance` and `.gc/runtime/packs/maintenance` during active discovery (lines 280–285), they are rendered inert. 

However, this boundary is vulnerable to two distinct failure modes during rollout and rollback:
1. **The Rollback Version-Skew Loophole:** Slices 5a and 5b consume the activation pin (which contains the migrated Maintenance assets) and remove Maintenance from required host packs. If an operator rolls back a city to an old binary, the old binary still force-includes Maintenance via `requiredBuiltinPackNames` (line 37). Because the public activation pin already contains those same assets, the old binary will load both, causing duplicate active behaviors. Since the old binary lacks the `zero-duplicate-active` gate, this will result in silent, non-deterministic shadowing and toxic map-merging.
2. **The "Dirty Directory" False-Positive:** The classifier must ignore stale directories unless they are classified as a `retired custom/fork` (line 263). If this classification is based on simple directory presence or a naive file-count comparison, untracked OS-generated files (such as `.DS_Store`, `.tmp` files, or IDE metadata) will cause a pristine stale folder to be classified as a custom fork, re-activating all legacy behaviors.

---

## 3. Critical Risks & Edge Cases

### 3.1. The Compatibility-Window Rollout Deadlock (Slices 2–4)
The plan defines a `zero-duplicate-active` gate that fails runtime loading "if the same behavior id is active from more than one source" (lines 277–279). 
* **The Vulnerability:** Slice 2 adopts the compatibility pin and rewires examples to public imports, while local Maintenance remains a required host pack in `requiredBuiltinPackNames` until Slice 5b.
* **The Deadlock:** If the compatibility-pin public pack contains the migrated Maintenance/Gastown assets (as implied by Slice 1b, line 568), then throughout Slices 2–4, those assets will be active from *both* the required in-tree Maintenance pack and the imported public pack. The `zero-duplicate-active` gate will trigger on every city load, crashing the runtime and preventing any operator from completing the compatibility window.
* **Resolution:** The plan must explicitly state that the compatibility-pin public pack *excludes* assets that duplicate still-required in-tree Maintenance (only the activation pin absorbs them), or define a soft warning diagnostic + deterministic precedence in `packcompat` during the compatibility window.

### 3.2. Stale Directory "Dirty" Custom Fork Promotion
* **The Edge Case:** An operator runs a manual diagnostic, leaving a `.tmp` file, or the OS writes a `.DS_Store` or `.thumbs.db` inside `.gc/system/packs/maintenance/`.
* **The Vulnerability:** If the classifier's detection of a "custom/fork" is based on simple filesystem diffs, it will falsely promote this directory to `retired custom/fork`. Instead of being ignored, the entire stale Maintenance folder (with its outdated, duplicated formulas) will be loaded, violating containment.
* **Resolution:** The classifier must calculate cryptographic SHA-256 digests of *only* registered behavior-bearing files (ignoring OS-specific, non-behavioral, or untracked paths) before promoting a stale directory to a "custom/fork."

### 3.3. In-Flight Session/Molecule Execution Survival
* **The Edge Case:** A complex, multi-agent molecule is composed and begins execution under legacy in-tree Maintenance/Gastown. Slices 2-7 are applied mid-execution, removing or ignoring those source paths.
* **The Vulnerability:** The plan is silent on whether step definitions and prompt fragments are snapshotted at molecule composition time (persisted as beads in the task store, making execution immune to source removal), or read live from the pack directory during step execution (causing immediate failures).
* **Resolution:** Define the execution outcome for in-flight molecules (e.g., "Step definitions are snapshotted in the task store at composition time; running sessions continue to resolve against the snapshotted state, while new sessions must select the migrated paths").

---

## 4. Required Plan Updates

To resolve these risks before final decomposition, the implementation plan must be updated to include:

1. **Explicit Compatibility-Pin Exclusion Contract:** State explicitly that the compatibility-pin commit *excludes* assets that duplicate still-required in-tree Maintenance, with only the activation-pin commit absorbing them; or specify that the `zero-duplicate-active` gate operates in "report-only/precedence" mode during the compatibility window.
2. **Cryptographic Custom-Fork Filter:** Require that the `retired custom/fork` state is determined solely by comparing the SHA-256 digests of registered behavior-bearing files against a pre-compiled pristine hash list, ignoring untracked or OS-specific paths.
3. **Mandatory Downgrade Boundary Release Notes:** Specify that adopting the activation pin is a **hard one-way boundary** for old-binary rollback, and make the downgrade-limit release notes mandatory rather than optional. Add the "old binary + activation pin" row to the rollout compatibility matrix.
4. **Enumerator Containment Contract:** Mandate that all five active behavior enumerators (prompt, formula, order, script, hook) either route through `packsource` or are structurally bound to only iterate over directories returned by `systempacks.RuntimeIncludes`.
5. **In-Flight Molecule Snapshotting Contract:** Explicitly state the persistence policy for mid-migration running sessions and in-flight molecules.

---

## 5. Missing Evidence

The following machine-testable artifacts or details are missing from the current design:
1. **The Pristine Hash Registry:** No pre-compiled or generated registry of cryptographic digests for legacy Maintenance/Gastown files.
2. **Active Behavior Enumerator Inventory:** Lacks a checked-in inventory of all filesystem-walking sites (e.g., `os.ReadDir`, `iofs.WalkDir`) in the codebase to prove that none bypasses `systempacks.RuntimeIncludes` or `packsource`.
3. **In-Flight State Persistence Spec:** Lacks documentation of whether `internal/beads` snapshots step definitions on molecule creation.

---

## 6. Schema Conformance

The plan **fully conforms** to the `gc.mayor.implementation-plan.v1` schema:
* **Front Matter:** Complete, correct, and matching the proposed schema.
* **Heading Order:** Standardized and unaltered (`Summary` $\rightarrow$ `Current System` $\rightarrow$ `Proposed Implementation` $\rightarrow$ `Data And State` $\rightarrow$ `Testing` $\rightarrow$ `Rollout And Recovery` $\rightarrow$ `Open Questions`).
* **Content:** Substantive, precise, and completely free of chronological attempt-history pollution.
