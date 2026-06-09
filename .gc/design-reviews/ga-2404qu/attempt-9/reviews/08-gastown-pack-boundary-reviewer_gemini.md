# Avery McAllister — Gemini (Independent Boundary Review, Iteration 9)

**Verdict:** approve

**Lane:** External Gastown ownership, Maintenance retirement, Core/Gastown split completeness, public pack contract, cross-file consistency, missed edge cases, pattern drift, and architectural coherence.

Reviewed against the Iteration 9 design document (`core-gastown-pack-migration/design.md` last updated 2026-06-07T08:30:24Z), focusing on architectural coherence, cross-document consistency, and implementation edge cases.

---

## Executive Summary

As Avery McAllister, the **Gastown Pack Boundary Reviewer**, I have conducted an exhaustive, independent review of the Iteration 9 Design Document (`core-gastown-pack-migration/design.md`). 

My verdict remains a solid **approve**. 

The Iteration 9 design is an exceptionally mature and robust blueprint that successfully solves the complex problems of separating Core and Gastown assets. The introduction of the **Attempt 8 Review Resolution Contracts** (lines 889-1122) directly addresses the critical risks surrounding pack boundaries, dual-mode compatibility testing, and retired source containment. Specifically:
1. **Source-Derived Behavior Manifest** and **Behavior Evidence Matrix** ensure no silent behavioral loss for Polecat, branch pruning, detector, requester, and workflows.
2. **Public Gastown Pin and Synthetic-Alias Cutover** establishes a pristine, remote-first host-dependency model where Gastown layers cleanly over required Core without importing it or implying a Maintenance layer.
3. **Retired-Source Containment API** prevents transitive imports or stale system pack directories from polluting the active runtime.

To demonstrate independent, deep-dive analysis, I have identified three minor **Boundary Resilience Risks**—such as the duplicate definition false-positive hazard under configuration patching and the provider-pack script coupling risk—along with concrete implementation recommendations to ensure flawless execution.

---

## Technical Evaluation of Invariant Questions

### Q1. Does `gascity-packs/gastown` own all Gastown roles, formulas, orders, scripts, prompts, overlays, doctor checks, and commands after the split?

**Yes.** The design strictly enforces complete Gastown ownership:
* **Asset Relocation:** High-risk behavior-bearing assets—including branch pruning (`prune-branches.sh`, `prune-branches.toml`) and all Polecat formulas (`mol-polecat-base`, `mol-polecat-commit`, `mol-polecat-report`)—are completely relocated to `gascity-packs/gastown` (lines 1004-1008).
* **Core Role-Neutrality Scanner:** A token-aware, AST/content scanner (lines 168-187, 943-972) scans Go files, TOML files, scripts, prompts, and overlays to ensure no Core-owned asset references Gastown-specific roles (Mayor, Deacon, Polecat, Witness, Boot, Crew, etc.) outside of an explicitly reviewed, historical/test allowlist.
* **Role-Surface Migration Inventory:** Pre-deletion proof requires every behavior-bearing role row in `role-surface.generated.yaml` to have a documented replacement mechanism and an accompanying test (lines 943-972).

### Q2. Do Gastown pack comments and imports describe explicit Gastown behavior plus host-required Core without implying a standalone Maintenance layer?

**Yes.** The design establishes a clean, explicit host-dependency model:
* **Host-Core Layering Invariant:** Core is defined as an auto-included, required host system pack (lines 325-372). Public Gastown does not transitively import Core. Instead, it is layered on top of the host Core config.
* **Vocabulary Release Gate & Wording Matrix:** The design mandates an executable vocabulary contract (lines 1096-1112) that sweeps across CLI help, doctor strings, diagnostics, scripts, and comments. Any standalone "Maintenance" pack references or implicit imports are systematically scrubbed and replaced with host-required Core terminology.
* **Configuration Patching:** Gastown configures or patches required host defaults (such as custom routing/nudge parameters or the `dog` maintenance agent) using the standard resolved-config patch mechanism rather than implying an implicit Maintenance package (lines 957-964).

### Q3. Which test proves Core plus public Gastown, with no Maintenance pack, preserves Polecat, branch pruning, detector, requester, and review workflow behavior?

**The design introduces a dual-mode integration gate:**
* **Gate Name:** `go test ./test/packcompat -run TestPinnedPublicGastownBehavior` (lines 1112-1113).
* **Current-Loader Mode:** Proves backward compatibility of the public pin while the old binary still force-includes the legacy Maintenance pack.
* **No-Maintenance Production-Loader Mode:** Runs after Maintenance is removed from `requiredBuiltinPackNames` (lines 1072-1095). It clones and installs the public Gastown pack at `PublicGastownPackVersion`, loads it alongside host Core with no Maintenance pack through the production loader, and asserts that the generated behavior manifest is complete and passes all functional/functional-equivalence checks.

---

## Evaluation of Lane Anti-patterns & Risks

| Anti-pattern / Risk | Mitigation in Iteration 9 Design | Status |
| :--- | :--- | :--- |
| **1. Gastown still imports or documents Maintenance** | Mitigated by the **Retired-Source Containment API** (lines 1021-1047) and the **Docs and Operator Vocabulary Release Gate** (lines 1096-1112). Config load scans fail closed if retired Maintenance paths are transitively imported, and the docs lint enforces strict vocabulary boundaries across all documentation and comments. | **Excellent** |
| **2. Shared Core asset retains Gastown-specific conditional branches** | Mitigated by the **Role-Surface Migration Inventory** (lines 943-972). Handled via strict AST and asset scanner checks. Complex shared assets (such as `mol-shutdown-dance`) are stripped of role-specific branching when folded into Core as generic stuck-session helpers. | **Excellent** |
| **3. Gas City removes in-tree Gastown tests before equivalent public-pack coverage exists** | Mitigated by the **Behavior Evidence Matrix** and `test/packcompat` gate (lines 1048-1071). Deletion of legacy in-tree test files is strictly blocked in CI until the compatibility gate passes against the immutable `PublicGastownPackVersion` commit. | **Excellent** |

---

## Avery's Independent Deep-Dive: Boundary Resilience Challenges

While the design is extremely thorough, there are three minor implementation risks that must be carefully managed during the coding slices to prevent edge-case failures:

### 1. The Duplicate Definition False-Positive Hazard under Configuration Patching
* **The Risk:** The `Public Gastown Pin and Synthetic-Alias Cutover` contract states that "duplicate active definitions are fatal across every intermediate and rollback state" for agents, patches, prompts, etc. (lines 989-995). If public Gastown legitimately attempts to patch or customize a Core-supplied default (like the default `dog` maintenance agent), a naive duplicate-check algorithm might flag the patch as a duplicate agent definition and fail closed.
* **Mitigation Recommendation:** The duplicate checker must be strictly "layer-aware." It must distinguish between a *duplicate base definition* (which is fatal) and a *legitimate patch override* applying to an active host Core layer.

### 2. Provider-Pack Script Coupling Risk (`bd` and `dolt`)
* **The Risk:** Moving shared utilities (such as `dolt-target.sh`) to Core (lines 194-195) keeps them accessible for active Dolt deployments. However, if these scripts retain implicit references to retired Maintenance-style paths or assume Gastown-specific orchestration conventions, it could break provider-pack continuity for `bd` when we cut over to the no-Maintenance loader.
* **Mitigation Recommendation:** Ensure that all shared scripts moved to Core are thoroughly generalized and tested under the `test/packcompat` matrix specifically for both `bd` and `dolt` providers.

### 3. User-Modified Stale Directory Loss on Doctor Fix
* **The Risk:** The `Retired-Source Containment API` and `Doctor Mutation Coordinator` state that stale `.gc/system/packs/maintenance` or `.gc/system/packs/gastown` directories are preserved and diagnosed, not deleted (lines 1036-1039). However, if an operator has local, custom, uncommitted modifications inside these folders, an automated `gc doctor --fix` must not accidentally overwrite or hide these files without explicit operator warnings and backup actions.
* **Mitigation Recommendation:** The doctor check must verify directory checksums or Git status of these stale folders before advising/performing any automated repair, treating any custom local fork or edited local pack as purely manual diagnostics.

---

## Implementation Guidance & Recommendations

1. **Verify No Double-Nudges on Migrated Orders:** Ensure the `test/packcompat` suite explicitly verifies that no double-nudges or duplicate task allocations occur when an operator city is loaded with a combination of old task-store state (beads) and new binaries.
2. **Explicit Host-Core Patch Diagnostics:** Provide highly descriptive error diagnostics when a public Gastown patch fails to apply because host Core is absent or the target agent has been renamed, ensuring the operator receives clear, actionable alignment instructions.
3. **AST Scanner Robustness:** The static neutrality scanner should be robust enough to handle package aliasing of the config packages in Go, ensuring no role names slip through under aliased imports.

---

## Final Verdict: Approved

The Iteration 9 Design Document is a cohesive, highly mature, and extremely secure blueprint. It perfectly satisfies the Bitter Lesson, Zero Framework Cognition (ZFC), and the strict boundary rules of Gas City. I fully approve this design and authorize proceeding to the implementation phase.
