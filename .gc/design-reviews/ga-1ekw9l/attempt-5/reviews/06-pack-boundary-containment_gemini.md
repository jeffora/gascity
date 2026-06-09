# Owen Gallagher — Gemini (Pack Boundary Containment Reviewer, Attempt 5, Independent DeepSeek V4 Flash Style)

**Verdict:** approve-with-risks

> **Lane:** Core versus Gastown ownership split, retired Maintenance source containment, active discovery classifier, no duplicate active behavior.
>
> Reviewed against the Attempt 5 design document (`.gc/design-reviews/ga-1ekw9l/attempt-5/design-before.md`, 657 lines, `updated_at: 2026-06-09T07:28:00Z`) — §"Pack Registry, Cache, And Retired Source Authority" (lines 252–286), §"Rollout And Recovery" (lines 561–647), and §"Data And State" (lines 426–486).
>
> This independent review is produced using the DeepSeek V4 Flash style, focusing specifically on cross-document consistency, latent edge cases, and assumptions that other reviewers may accept too quickly.

---

## 1. Executive Summary

As Owen Gallagher, the **Pack Boundary Containment Reviewer**, I have conducted a rigorous, first-principles independent review of the Attempt 5 implementation plan. My final determination is **Verdict: approve-with-risks**.

While the plan outlines an elegant architecture for separating the required Core system pack from the now-retired Maintenance and public Gastown packs, it retains several dangerous containment escapes, rollout deadlocks, and circular dependencies that must be resolved prior to final decomposition.

The most critical findings are:
1. **The Doctor-Preflight Circular Dependency (Loader Deadlock):** The loader fail-closes on duplicate active behavior before any behavior executes. However, `gc doctor --fix` must load the config to analyze and fix the city. If the duplicate gate blocks loading entirely, the doctor can never bootstrap to apply the fix, locking the city in a permanent bricked state.
2. **The Compatibility-Pin Deadlock:** Adopting the compatibility-pin while embedded Maintenance is still required will trigger the `zero-duplicate-active` gate and block all loading across Slices 2-5a, halting rollout.
3. **AST-Bypassing Directory Enumeration:** Raw directory reads (e.g., `os.ReadDir` on `.gc/system/packs`) can enumerate and load legacy directories like `maintenance` without using forbidden string literals, bypassing AST scanners.
4. **Stale Directory "Dirty" Promotion:** Pristine stale directories can be falsely promoted to custom/fork status due to non-behavioral OS/IDE files (e.g., `.DS_Store`), causing duplicate behavior activation.

---

## 2. Lane-Specific Evaluation

### Q1: Is there one classifier API used before config load, install, cache lock, validation, discovery, docs lint, prompt, formula, order, script, and hook enumeration?

**Answer:** Yes, the plan defines `internal/packsource` as the sole authority (lines 259–266). However, the enforcement of this boundary relies on a "Scanner test [rejecting] ... retired-source string classifiers" (lines 508-511). 

This is a major security loophole. An AST-level string check only detects literal strings like `"maintenance"` or `"gastown"`. It cannot detect generic, dynamic directory enumeration. For example, any code that performs:
```go
files, _ := os.ReadDir(".gc/system/packs")
for _, f := range files {
    loadPack(f.Name()) // Dynamic loading of 'maintenance' without using the string "maintenance"
}
```
will completely bypass the string scanner, violating the containment invariant.

**Requirement:** The active enumerators (prompt, formula, order, script, hook) must be structurally bound to only iterate over directories returned by `systempacks.RuntimeIncludes`, or they must call `packsource.Classify` on every path before loading.

---

### Q2: Does a city running Core plus public Gastown and no Maintenance preserve Gastown-specific behavior without shared Core conditionals?

**Answer:** Yes, using symbolic bindings under `[gc.bindings.*]` (lines 343–357) and removing hardcoded role references in Go code (lines 336–342) structurally isolate Gastown behavior.

However, the risk of **implicit behavior leaking** remains high. For example, generic formulas in Core (such as `mol-prompt-synth.toml` or `mol-shutdown-dance`) may keep Gastown-specific timeout curves or escalation rules under generalized names.

The terminology matrix (lines 405–413) must be supplemented by a **provenance-sensitive packcompat test** verifying that no Gastown prompt fragment or variable is loaded or evaluated from Core paths when running in Core-only mode.

---

### Q3: Can stale in-tree Gastown or Maintenance directories remain on disk without entering active prompts, formulas, orders, scripts, hooks, or rollback states?

**Answer:** Yes, by ignoring stale `.gc/system/packs/maintenance` and `.gc/runtime/packs/maintenance` during active discovery (lines 280–285), they are rendered inert.

However, this boundary is vulnerable to two distinct failures:
1. **The Rollback Version-Skew Loophole:** Slices 5a and 5b consume the activation pin (which contains the migrated Maintenance assets) and remove Maintenance from required host packs. If an operator rolls back a city to an old binary, the old binary still force-includes Maintenance via `requiredBuiltinPackNames` (line 37). Because the public activation pin already contains those same assets, the old binary will load both, causing duplicate active behaviors. Since the old binary lacks the `zero-duplicate-active` gate, this will result in silent, non-deterministic shadowing and toxic map-merging.
2. **The "Dirty Directory" False-Positive:** The classifier must ignore stale directories unless they are classified as a `retired custom/fork` (line 263). If this classification is based on simple directory presence or a naive file-count comparison, untracked OS-generated files (such as `.DS_Store`, `.tmp` files, or IDE metadata) will cause a pristine stale folder to be classified as a custom fork, re-activating all legacy behaviors.

---

## 3. Critical Risks & Edge Cases

### 3.1. The Doctor-Preflight Circular Dependency (Loader Deadlock)
* **The Vulnerability:** The `zero-duplicate-active` gate fail-closes on duplicate active behavior before any behavior executes (lines 277-279). It also states that `gc doctor --fix` is the only way to resolve legacy imports and perform runtime-state migrations (lines 313-316).
* **The Deadlock:** Running `gc doctor --fix` requires loading the city's runtime config to check status. If the loader fail-closes on duplicates before doctor preflight can run, then a city with duplicates can never be loaded by the doctor, preventing the automatic fix from ever being applied!
* **Resolution:** The loader must have a designated "diagnostic/doctor mode" that allows schema/import-resolution loading without executing active behaviors, or doctor must bypass the duplicate active gate during preflight.

### 3.2. The Compatibility-Window Rollout Deadlock (Slices 2–4)
* **The Vulnerability:** Slice 2 adopts the compatibility pin, while local Maintenance remains a required host pack in `requiredBuiltinPackNames` until Slice 5b.
* **The Deadlock:** If the compatibility-pin public pack contains the migrated Maintenance/Gastown assets, those assets will be active from *both* the required in-tree Maintenance pack and the imported public pack. The `zero-duplicate-active` gate will trigger on every city load, crashing the runtime and preventing any operator from completing the compatibility window.
* **Resolution:** State explicitly that the compatibility-pin public pack *excludes* assets that duplicate still-required in-tree Maintenance (only the activation pin absorbs them), or define a soft warning diagnostic + deterministic precedence in `packcompat` during the compatibility window.

### 3.3. Stale Directory "Dirty" Custom Fork Promotion
* **The Edge Case:** An operator runs a manual diagnostic, leaving a `.tmp` file, or the OS writes a `.DS_Store` or `.thumbs.db` inside `.gc/system/packs/maintenance/`.
* **The Vulnerability:** If the classifier's detection of a "custom/fork" is based on simple filesystem diffs, it will falsely promote this directory to `retired custom/fork`. Instead of being ignored, the entire stale Maintenance folder (with its outdated, duplicated formulas) will be loaded, violating containment.
* **Resolution:** The classifier must calculate cryptographic SHA-256 digests of *only* registered behavior-bearing files (ignoring OS-specific, non-behavioral, or untracked paths) before promoting a stale directory to a "custom/fork."

### 3.4. In-Flight Session/Molecule Execution Survival
* **The Edge Case:** A complex, multi-agent molecule is composed and begins execution under legacy in-tree Maintenance/Gastown. Slices 2-7 are applied mid-execution, removing or ignoring those source paths.
* **The Vulnerability:** The plan is silent on whether step definitions and prompt fragments are snapshot at molecule composition time (persisted as beads in the task store, making execution immune to source removal), or read live from the pack directory during step execution (causing immediate failures).
* **Resolution:** Define the execution outcome for in-flight molecules (e.g., "Step definitions are snapshotted in the task store at composition time; running sessions continue to resolve against the snapshotted state, while new sessions must select the migrated paths").

---

## 4. Required Plan Updates

To resolve these risks before final decomposition, the implementation plan must be updated to include:

1. **Doctor-Preflight Bypass/Diagnostic Mode:** Introduce a dedicated diagnostic/doctor mode in `systempacks.LoadRuntimeCity` that allows loading config structure and imports without enforcing the `zero-duplicate-active` execution gate.
2. **Explicit Compatibility-Pin Exclusion Contract:** State explicitly that the compatibility-pin commit *excludes* assets that duplicate still-required in-tree Maintenance, with only the activation-pin commit absorbing them; or specify that the `zero-duplicate-active` gate operates in "report-only/precedence" mode during the compatibility window.
3. **Cryptographic Custom-Fork Filter:** Require that the `retired custom/fork` state is determined solely by comparing the SHA-256 digests of registered behavior-bearing files against a pre-compiled pristine hash list, ignoring untracked or OS-specific paths.
4. **Mandatory Downgrade Boundary Release Notes:** Specify that adopting the activation pin is a **hard one-way boundary** for old-binary rollback, and make the downgrade-limit release notes mandatory rather than optional. Add the "old binary + activation pin" row to the rollout compatibility matrix.
5. **Enumerator Containment Contract:** Mandate that all five active behavior enumerators (prompt, formula, order, script, hook) either route through `packsource` or are structurally bound to only iterate over directories returned by `systempacks.RuntimeIncludes`.
6. **In-Flight Molecule Snapshotting Contract:** Explicitly state the persistence policy for mid-migration running sessions and in-flight molecules.

---

## 5. Missing Evidence

The following machine-testable artifacts or details are missing from the current design:
1. **Diagnostic/Doctor Mode Loader Spec:** Lacks documentation of how `LoadRuntimeCity` behaves when called by `gc doctor`.
2. **The Pristine Hash Registry:** No pre-compiled or generated registry of cryptographic digests for legacy Maintenance/Gastown files.
3. **Active Behavior Enumerator Inventory:** Lacks a checked-in inventory of all filesystem-walking sites (e.g., `os.ReadDir`, `iofs.WalkDir`) in the codebase to prove that none bypasses `systempacks.RuntimeIncludes` or `packsource`.
4. **In-Flight State Persistence Spec:** Lacks documentation of whether `internal/beads` snapshots step definitions on molecule creation.

---

## 6. Schema Conformance

The plan **fully conforms** to the `gc.mayor.implementation-plan.v1` schema:
* **Front Matter:** Complete, correct, and matching the proposed schema.
* **Heading Order:** Standardized and unaltered (`Summary` $\rightarrow$ `Current System` $\rightarrow$ `Proposed Implementation` $\rightarrow$ `Data And State` $\rightarrow$ `Testing` $\rightarrow$ `Rollout And Recovery` $\rightarrow$ `Open Questions`).
* **Content:** Substantive, precise, and completely free of chronological attempt-history pollution.
