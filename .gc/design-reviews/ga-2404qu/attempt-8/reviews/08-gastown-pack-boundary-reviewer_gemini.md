# Avery McAllister — Gemini 3.5 Flash (Independent Review, Iteration 8)

**Verdict:** pass

**Persona focus:** External Gastown ownership, Maintenance retirement, Core/Gastown split completeness, public pack contract, cross-file consistency, missed edge cases, pattern drift, and architectural coherence.

---

## Executive Summary

The transition of my verdict from **approve-with-risks** to a full **pass** in Iteration 8 is justified by the introduction of the highly comprehensive **Attempt 7 Review Resolution Contracts** (lines 590-888). The revised design directly addresses all previous pack boundary vulnerabilities. By mandating a rigorous **Behavior Manifest and Packcompat Evidence Contract** (lines 780-823), an explicit **Public Gastown Host-Core Contract** (lines 325-372), a **Test-Migration and Reconciliation Table** (lines 319-324, 800-803), and a robust **Retired-Source Classifier** (lines 749-779), the design completely eliminates risk of silent coverage loss, stale imports, vocabulary leakage, or cyclic dependencies. The resulting split model is exceptionally clean, robust, and maintains high architectural coherence.

---

## Evaluation of Avery's Critical Questions

### 1. Does `gascity-packs/gastown` own all Gastown roles, formulas, orders, scripts, prompts, overlays, doctor checks, and commands after the split?

**Yes. The design strictly enforces complete Gastown ownership:**
* **Complete Separation of Roles:** The design explicitly states that Core must not own any Mayor, Deacon, Polecat, Refinery, Witness, Boot, Crew, or Gastown role behavior (lines 1000-1003).
* **Asset Relocation:** High-risk assets like branch pruning (`prune-branches.sh` and `prune-branches.toml`) and all Polecat formulas (`mol-polecat-base`, `mol-polecat-commit`, `mol-polecat-report`) are fully moved to the `gascity-packs/gastown` repository (lines 1004-1008). 
* **Go-Level Role-Neutrality Scanner:** A token-aware Go test (lines 513-541, 650-685) scans Go, TOML, shell scripts, prompts, overlays, and markdown files to guarantee that no behavior-bearing Core assets contain hardcoded Gastown role names outside a highly restricted historical/test allowlist.

### 2. Do Gastown pack comments and imports describe explicit Gastown behavior plus host-required Core without implying a standalone Maintenance layer?

**Yes. The design establishes a clean host-dependency model:**
* **Host-Core Layering Invariant:** Core is defined as an auto-included host system pack owned by the Gas City binary (lines 328-331). Public Gastown does not import Core; instead, it is layered above required system packs and can safely patch host Core configuration (like the default `dog` maintenance agent) using the standard resolved-config patch mechanism (lines 332-347, 1088-1094).
* **Prose and Metadata Cleanup:** Comments and documentation within the public `gascity-packs/gastown/pack.toml` will be updated to explicitly remove all references to an implicit Maintenance layer, replacing them with host-required Core and explicit Gastown behavior descriptions (lines 1088-1090).
* **Vocabulary Enforcement:** A wording matrix in the docs lint (lines 570-584, 865-888) strictly separates Core, provider-dependent host system packs (`bd`/`dolt`), retired standalone Maintenance packs, and explicit public Gastown imports.

### 3. Which test proves Core plus public Gastown, with no Maintenance pack, preserves Polecat, branch pruning, detector, requester, and review workflow behavior?

**The design introduces a dual-mode integration gate:**
* **Gate Name:** `go test ./test/packcompat -run TestPinnedPublicGastownBehavior` (lines 1112-1113).
* **Current-Loader Mode:** Proves backward-compatibility of the public pin while the old binary still force-includes Maintenance.
* **No-Maintenance Production-Loader Mode:** Runs after Maintenance is removed from `requiredBuiltinPackNames` (lines 551-562, 1114-1120). It clones/installs the public Gastown pack at `PublicGastownPackVersion`, loads it alongside host Core with no Maintenance pack through the normal production loader, and asserts that the generated behavior manifest is complete and passes all functional checks.

---

## Evaluation of Red Flags

* **Red Flag 1: Gastown still imports or documents Maintenance.**
  * *Status:* **Resolved.** The design mandates that public Gastown pack scans must fail on retired Maintenance paths (`.gc/system/packs/maintenance`, `packs/maintenance`, etc., lines 368-371). Furthermore, all public `pack.toml` comments and doc prose are scrubbed and verified by the executable wording contract.
* **Red Flag 2: Shared Core asset retains Gastown-specific conditional branches.**
  * *Status:* **Resolved.** Handled by the strict role-neutrality scanner (lines 170-187, 513-541). Additionally, complex shared assets like `mol-shutdown-dance` are stripped of Deacon, Witness, Polecat, and Mayor conditional branches upon being folded into Core as generic stuck-session helpers (lines 1016-1021).
* **Red Flag 3: Gas City removes in-tree Gastown tests before equivalent public-pack coverage exists.**
  * *Status:* **Resolved.** Deletion is completely blocked until the `test/packcompat` gate passes against the immutable `PublicGastownPackVersion` commit (lines 1095-1111).

---

## Resolution of Previously Raised Critical Risks

### 1. Cross-Pack Script Dependencies and Couplings (Risk #1)
* **Resolution:** The design mandates an **Executable Source-Discovery Manifest** (lines 269-324) that follows all formula `script`, prompt-fragment, order, and hook overlay references, cataloging them as rows or row dependencies. Any shared script or prompt fragment must be duplicated or refactored into a provider-neutral Core fragment (lines 360-367), completely eliminating implicit path-based cross-pack dependencies or scanner alerts.

### 2. Lack of Automated Test Completeness and Reconciliation Verification (Risk #2)
* **Resolution:** The design introduces a **Test-Migration Table** (lines 800-803) and a strict reconciliation check in Gas City CI: the CI gate fails if any test is removed from `gastown_test.go` or `maintenance_scripts_test.go` without a manifest row mapping it to a new `gascity-packs` test or a documented removal (lines 319-323, 1107-1110).

### 3. "Documents Maintenance" Wording and Prose Leakage in External Pack (Risk #3)
* **Resolution:** Addressed by updating the public `pack.toml` comments (lines 1088-1090), implementing public pack retired path failures (lines 368-371), and applying the executable vocabulary contract (lines 865-888) across CLI help, diagnostics, and documentation.

### 4. `fallback = true` in Core Dog Agent Changes Generic City Behavior (Risk #4)
* **Resolution:** The `dog` agent is decoupled from Go-level hardcoding and is treated as a configurable Core maintenance-worker target (lines 673-680). Tests prove that Core-only cities load and run normally when the maintenance agent is renamed or omitted entirely, ensuring that non-Gastown SDK consumers are not affected by hidden fallbacks.

### 5. Stale Maintenance Packs Loaded via Transitive Import Resolution (Risk #5)
* **Resolution:** A central **Retired-Source Classifier** (lines 749-779) is introduced. It distinguishes active required host system packs from generated retired Maintenance sources and preserved stale directories. Prompt and template discovery is restricted to active resolved packs and required host packs, explicitly ignoring stale system-pack directories even if they remain on disk (lines 767-771).

---

## Recommendations for the Implementation Phase

1. **Verify No Double-Nudges on Moved Orders:** During the transition where Maintenance-owned orders are migrated to Core generic or public Gastown ownership, ensure the `test/packcompat` suite verifies that no double-nudges or duplicate task allocations occur if an operator city is loaded with a combination of old state and new binaries.
2. **Explicit Host-Core Patch Diagnostics:** Provide highly descriptive error diagnostics when a public Gastown patch fails to apply because the host Core is absent or the target agent has been renamed, ensuring the operator receives clear, actionable instructions on aligning their system packs.
