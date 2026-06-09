# Avery McAllister — DeepSeek V4 Flash Perspective Independent Review (Iteration 12 / Attempt 12)

**Verdict:** approve-with-risks

**Lane:** External Gastown ownership, Maintenance retirement, Core/Gastown split completeness, public pack contract.

This independent review evaluates the Iteration 12/Attempt 12 draft of `design.md` (`updated_at: 2026-06-07T14:05:04Z`) against `requirements.md` and the existing codebase. It pays specific attention to the newly introduced **Attempt 11 Review Resolution Contracts (§1694–1951)** and their downstream impact on pack boundaries, dependency management, and old-binary compatibility during rollout.

---

## Executive Summary

The Iteration 12 design represents a masterful, highly robust blueprint for separating required Core SDK primitives from the Gastown orchestration domain. By introducing the **Concrete Core Maintenance-Worker Binding (§1703–1762)** and the **Full Role-Surface Inventory and Replacement Gate (§1823–1851)**, the design successfully dismantles the implicit lookup shadowing and uninventoried role-leakage issues that historically blocked previous iterations. 

The symbolic patch target contract (`target_binding`) provides public Gastown with a safe, decoupled customization channel, allowing the host Core to remain entirely role-neutral. Meanwhile, the dual-pin rollout architecture in `public-gastown-pins.yaml` establishes an auditable ledger for the transition, addressing critical rollout-collision hazards with real compiler gates.

However, from an exhaustive **Gastown Pack Boundary and Public Pack Contract** perspective, several critical interface gaps, stale decision remnants, and compatibility contradictions remain:
1. **The `PromptTemplate` Path Resolution Void (§1750):** While Gastown can symbolically patch the host-Core `maintenance_worker` to reference its own agent, the design never specifies how the `PromptTemplate` path within the patching pack is resolved. If the path does not resolve relative to the patching pack's filesystem root, cross-pack patching is unimplementable.
2. **Ghost Asset remnants in the Cross-Pack Table (§2515):** The `Gastown Codex overlay` remains listed in the Cross-Pack Ownership table as a pending `review` decision, despite having already been deleted in the live codebase.
3. **Undefined Overlay Destination Collision Semantics:** While duplicate formulas and orders are strictly banned and validated at loading, duplicate provider-specific overlays (e.g., Core and public Gastown trying to write to the same target hook file) have no merge or precedence rules defined.
4. **Version Skew Incompatibility on the Activation Pin (§2695):** The Release Compatibility matrix over-promises support for "old binary | new pack" by asserting that the new pack will remain backward compatible. This is false for the activation-pin pack, which utilizes `target_binding` patches that older binaries cannot parse, triggering immediate fatal undecoded TOML key warnings.

---

## Technical Evaluation of Invariant Questions

### Q1. Does gascity-packs/gastown own all Gastown roles, formulas, orders, scripts, prompts, overlays, doctor checks, and commands after the split?
* **Yes:** The design specifies complete relocation of all high-risk Gastown-bearing behaviors. Parent/child Polecat formulas (`mol-polecat-base`, `mol-polecat-commit`, etc.), branch pruning scripts and orders (`prune-branches.sh`/`.toml`), and role-specific templates are cleanly migrated to `gascity-packs/gastown` (§2063–2072). 
* **Role Neutrality Scanner:** A static source scanner (§2112–2117) enforces this split by ensuring that no files remaining in Core contain references to Gastown roles (`mayor`, `deacon`, `witness`, `refinery`, `polecat`, `boot`, `crew`) outside of explicit, allowlisted historical documentation.

### Q2. Do Gastown pack comments and imports describe explicit Gastown behavior plus host-required Core without implying a standalone Maintenance layer?
* **Yes:** The design enforces a strict non-transitive dependency model. Gastown does not explicitly import Core; Core remains an auto-included host system pack injected during config loading (§2151–2157). 
* **Vocabulary Matrix:** A dedicated wording matrix (§1920–1944) enforces that all diagnostics, comments, and tutorials are systematically scrubbed of legacy "Maintenance pack" terminology, replacing them with role-neutral "Core maintenance-worker" or "system pack" phrasing. Stale imports are explicitly rejected by the `internal/packsource` classifier before resolution.

### Q3. Which test proves Core plus public Gastown, with no Maintenance pack, preserves Polecat, branch pruning, detector, requester, and review workflow behavior?
* **The Dual-Mode Integration Gate:** Proved via `go test ./test/packcompat -run TestPinnedPublicGastownBehavior` (§2179–2188). 
* **Current-Loader Mode:** Verifies that the public compatibility pin runs safely alongside bundled Maintenance in existing cities without duplicate active definition errors.
* **No-Maintenance Mode:** Runs after Maintenance is retired from `requiredBuiltinPackNames`, proving that the normal production loader (`internal/systempacks.LoadRuntimeCity`) successfully resolves the pinned public Gastown pack (`activation` pin) and passes all functional verification checks for Polecat, branch pruning, detector, requester, and reviewer workflows.

---

## Top Strengths of the Current Design

- **Decoupled Symbolic Worker Binding (§1703–1746):** Utilizing symbolic bindings (`[gc.bindings.maintenance_worker]`) rather than hardcoded role constants in Go completely eliminates template shadowing and namespace collisions. It is an outstanding application of Zero Framework Cognition (ZFC).
- **Comprehensive Behavior Manifest (§2118–2138):** Mapping every moved asset to its old path, delta, new path, and verification test ensures 100% auditable continuity of Gastown's feature set.
- **Enforceable Pre-Resolution File-Set Integrity (§1799–1804):** Checking required packs against exact manifests and `pack.toml` digests before any behaviors are read prevents silent corruption of core system assets.
- **Wording Matrix with CI Teeth (§1920–1927):** Scanning all Markdown, MDX, TOML, and Go comments for retired pack vocabulary prevents documentation and DX drift over the course of the staged rollout.

---

## Critical Risks, Gaps & Hidden Dependencies

### 1. The `PromptTemplate` Path Resolution Ambiguity
* **The Code Evidence:** `internal/config/config.go` defines `AgentOverride` with `PromptTemplate *string` (§16). The Attempt 11 contract introduces symbolic patching of the maintenance worker via `[[patches.agent]]` (§1750).
* **The Gap:** When public Gastown patches the host-Core `maintenance_worker` to use its custom prompt template, it must declare a `prompt_template` path (e.g., `prompt_template = "assets/prompts/deacon.template.md"`). However, the design does not specify the base directory against which this relative path is resolved. If the loader resolves it relative to the *city's* root, it will fail because the template only exists inside the public Gastown pack's directory. If the loader resolves it relative to the patching pack's root, this behavior must be explicitly coded and verified in the config layer.
* **Recommendation:** Explicitly specify that any `prompt_template` path supplied by an imported pack's patch layer must resolve relative to that patching pack's root directory, rather than the city's root directory or the host-Core pack's root directory.

### 2. Ghost Asset Remnants in the Cross-Pack Ownership Table
* **The Code Evidence:** The Cross-Pack Ownership Decisions table (§2515) includes the `Gastown Codex overlay` as a pending `review` decision.
* **The Gap:** The in-tree Gastown Codex overlay has already been deleted in the live codebase (as verified in §19–21 of Claude's prior review). Retaining this row in the design document creates an implementation hazard, as developers will waste cycles trying to audit and move a ghost asset.
* **Recommendation:** Completely delete the `Gastown Codex overlay` row from the Cross-Pack Ownership Decisions table before finalizing the design.

### 3. Undefined Overlay-File Collision Semantics
* **The Gap:** While the design details strict collision detection for duplicate formulas and orders, it does not define merge or precedence rules for file-set overlays (e.g. `overlay/per-provider/codex/.codex/hooks.json`). If both Core and an imported pack attempt to supply the same destination overlay file, the behavior is undefined. Will the imported pack overwrite the Core overlay, or will the loader raise a fatal duplicate error?
* **Recommendation:** Establish a clear precedence rule for overlays: imported pack overlays merge with or override host Core overlays of the same destination path, and any fatal collision checks are confined to exact duplicate content.

### 4. Version Skew Incompatibility on the Activation Pin
* **The Code Evidence:** The Release Compatibility matrix (§2695) says:
  ```markdown
  | old binary | new pack | Public pack remains compatible ... no reliance on new Gas City-only loader behavior |
  ```
* **The Gap:** This row holds true for the `compatibility` pin, but is fundamentally false for the `activation` pin. Under Attempt 11, the activation pin requires active public Gastown assets to utilize `target_binding` symbolic patches. Older binaries do not implement this configuration field. Because the old config loader is configured to detect unknown keys and fail (`fatalUndecodedWarnings`), running an old binary against the activation-pin pack will result in immediate load failures or silent omissions of the dog patch.
* **Recommendation:** Split the "old binary | new pack" compatibility matrix row into two: one for the compatibility pin (which remains fully compatible), and one for the activation pin (which triggers a version-skew diagnostic or warning, with manual recovery instructions).

---

## Required Changes for Finalization

1. **Specify Cross-Pack Template Resolution:** Update §1750 to mandate that the config layer resolves relative `prompt_template` paths inside imported patch layers against the root directory of the importing/patching pack.
2. **Purge Ghost Codex Overlay:** Remove the stale `Gastown Codex overlay` row from the Cross-Pack Ownership Decisions table (§2515).
3. **Define Overlay Collision Rules:** Add an explicit invariant to the loader design: "Two active packs shipping the same destination overlay file must merge their configurations if JSON/JSONL, or let the imported pack overlay take precedence over Core, without raising a fatal loading error."
4. **Correct the Version Skew Matrix:** Split the "old binary × new pack" row in §2690 to explicitly separate the `compatibility` pin from the `activation` pin, detailing the expected diagnostic failures on the activation pin for old binaries.

---

## Questions for the Implementation Team

1. Does the current config parser support relative path resolution for `PromptTemplate` strings that are declared in imported-pack `[[patches.agent]]` sections, and does it correctly map them to the pack's cached subpath?
2. If an operator runs `gc doctor --fix` in an air-gapped CI environment, how does the `doctor.MutationCoordinator` verify the reachability and integrity of the public Gastown `compatibility` or `activation` pins without failing on live network calls?
3. What mechanism is planned to prevent AST scanner false-positives on generic terms like `crew` or `boot` when scanning Core shell scripts?
