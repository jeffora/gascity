# Ingrid Kovac — DeepSeek V4 Flash (Role Neutrality Guardian, Iteration 17)

**Verdict:** approve-with-risks

**Lane:** Zero hardcoded roles, Core role neutrality, `dog` exception containment, SDK self-sufficiency, Go-source migration guard coverage, cross-document consistency.

Reviewed against the revised design document in Iteration 17 (`.gc/design-reviews/ga-2404qu/attempt-17/design-before.md` updated 2026-06-09T02:00:53Z) and grounded in the live codebase in `cmd/gc/`, `internal/config/`, `internal/packs/core/`, and `internal/systempacks`.

---

## Executive Summary

As Ingrid Kovac, the **Role Neutrality Guardian**, I am updating my verdict to **Verdict: approve-with-risks** for Iteration 17.

The Iteration 17 design is a masterpiece of architectural maturation. It successfully moves beyond the simplistic "token scanning" approach of early attempts, replacing ad-hoc code checks with a robust, type-safe symbolic binding model (`core.maintenance_worker`) and a comprehensive pre-resolution configuration resolution boundary. This is a monumental victory for the Zero Framework Cognition (ZFC) principle: role names are completely banished from Go control logic, and all roles are strictly user-supplied configuration.

However, a rigorous audit of the Proposed Design reveals a few subtle **role leakage risks**, **unspecified implementation surfaces**, and **logical contradictions** regarding prompt prose and deprecated API wrapping. These must be addressed before the design is finalized and implementation begins.

---

## Top Strengths & Design Evolution

- **The Symbolic Binding Model (§1796–1845):** The introduction of `[gc.bindings.maintenance_worker]` with dynamic pool, target, and step-metadata binding (`gc.run_target_binding = "core.maintenance_worker"`) is the gold standard of role-neutral design. It completely decouples the SDK's execution infrastructure from any assumption that a pool named `"dog"` exists on the user's system.
- **Strict Omission/Renaming Contracts (§1812–1818, §1842):** Mandating explicit tests that prove a Core-only city still loads and evaluates all non-agent controller operations when the maintenance worker is omitted or renamed is the ultimate proof of SDK self-sufficiency.
- **Path-, Token-, and Field-Aware Scanner (§2559–2564):** Mandating that the token scanner parses and tokenizes Go identifiers, string literals, TOML, shell scripts, Markdown, templates, and CLI help fixtures guarantees that no rogue Gastown roles creep back into the Core codebase.
- **The Provider Exceptions Ledger (§2370):** Permitting carefully scoped, tracked, and validated rewrites of legacy `mol-dog-*` provider assets to the generic `core.maintenance_worker` binding solves the byte-continuity paradox while preserving perfect role neutrality.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Does any Go change introduce role-conditional logic or a literal role name outside tests, migration docs, or pack configuration?
**No.** All production Go control branches and dispatching heuristics (such as the legacy `mol-polecat-*` prefix matching in `internal/sling`) are removed. Display metadata, tmux theme mappings, and icon classifications are moved entirely to config-supplied metadata or public Gastown assets, leaving the Go codebase 100% role-neutral.

### Q2: Does the Core role-name guard scan every asset type including scripts, overlays, orders, template fragments, doctor checks, metadata, and prompt snippets?
**Yes.** The specified scanner is path-, token-, and field-aware, recursively traversing Go files, TOML files, shell scripts, templates, overlays, skills, doctor strings, CLI help fixtures, and Markdown assets. It asserts that no new behavior-bearing references to Gastown roles land in Core.

### Q3: Can Core infrastructure still run when the default maintenance agent is removed or renamed by configuration?
**Yes, structurally.** The core bindings model decouples the maintenance worker. Omission of the worker disables only worker-bound maintenance orders at dispatch time (failing gracefully with `core.maintenance_worker.omitted`), while preserving all primary SDK infrastructure operations, config loading, and session management.

---

## Critical Risks & Architectural Inconsistencies

### 1. [Major] Unspecified `classifyAgentKind` De-Roling Mechanism
* **The Risk:** Section §594 correctly lists `classifyAgentKind` in the migration inventory as a surface that must be de-roled. However, the Proposed Design section does not specify *how* this function is generalized, what its role-neutral signature looks like, or what configuration metadata replaces its hardcoded role-kind classifications (e.g., classifying `crew` or other role families).
* **The Hazard:** Without an explicit specification, implementation developers may resort to ad-hoc, hardcoded string exceptions or loose heuristics that violate ZFC.
* **Required Change:** Explicitly specify that `classifyAgentKind` resolves agent classifications dynamically from pack or city-level configuration metadata, and define the generic categories (e.g., `infrastructure` vs. `user-defined`) in the schema.

### 2. [Major] Hardcoded Prompt Prose in Core Maintenance Prompts (Model Confusability)
* **The Risk:** While §2520 allows Core formulas and orders to retain `dog` in their file names for compatibility, §2981 states that Dog prompt fragments move to Core for generic maintenance-agent behavior.
* **The Hazard:** If a Core-owned prompt fragment contains literal text like *"You are the dog maintenance worker,"* and an operator renames their worker to *"cat"* or *"ops-maintainer"*, the running LLM will receive contradictory instructions (e.g., *"Your session name is cat but you are the dog worker"*). This violates ZFC and causes severe model confusion.
* **Required Change:** Mandate that all Core-owned maintenance prompt templates completely abstract the agent name using Go `text/template` variables (e.g., `{{.AgentName}}`), dynamically injecting the resolved binding target name at materialization or prompt-rendering time.

### 3. [Minor] Legacy Go Theme API (`DogTheme`) Leaks
* **The Risk:** Section §1369 states that *"Go APIs such as `DogTheme` are removed or wrapped behind neutral compatibility data."* 
* **The Hazard:** Keeping deprecated Go-level theme functions like `DogTheme()` in production Go packages (even as wraps or aliases) violates the strict invariant that no production Go code may reference specific role names.
* **Required Change:** Completely remove `DogTheme()` from production Go packages. Tmux themes/icons must be 100% config-driven. Any backward-compatibility for old configs targeting the `"dog"` name must be handled entirely at config-load time by mapping `"dog"` to a default config-supplied theme, keeping the Go API pristine.

### 4. [System Safety] Allowlist Stale-Entry Detection (Anti-Rot)
* **.The Risk:** As the codebase evolves, files in the scanner allowlist (`role-surface.generated.yaml`) will be deleted or renamed.
* **The Hazard:** The allowlist can easily accumulate stale paths, leading to "allowlist rot" where unmonitored legacy exclusions clutter the schema.
* **Required Change:** The role-token scanner test must assert that every file path declared in the allowlist portion of `role-surface.generated.yaml` actually exists in the repository, failing loudly if any entry is stale or refers to a deleted asset.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Iteration 17 Design | Status |
| :--- | :--- | :--- |
| **Role-name guard allowlist grows to excuse convenient violations** | **Mitigated.** Blocked by requiring that each allowlist row in the schema has an explicit owner, narrow scope, expiration slice, and associated negative test fixtures. | **Pass** |
| **Core order, formula, or script assumes a specific configured agent pool exists** | **Mitigated.** Enforced by Gate 2 and the `core.maintenance_worker` symbolic binding model, which maps all Core dispatching to a configurable target. | **Pass** |
| **`dog` or any other role becomes hidden SDK behavior instead of user-supplied configuration** | **Mitigated.** SDK infrastructure operations are completely decoupled from `dog`. Omission and renaming tests enforce that `dog` is purely default configuration. | **Pass** |

---

## Actionable Requirements & Proposed Adjustments

1. **Specify `classifyAgentKind`:** Add a dedicated subsection under the Proposed Design defining how `classifyAgentKind` resolves classification metadata dynamically from config, eliminating hardcoded role mappings.
2. **Abstract Prompt Prose:** Mandate that Core-owned maintenance prompts resolve the worker's name dynamically via Go template interpolation (e.g., `{{.AgentName}}`) rather than containing literal `"dog"` text.
3. **Purge `DogTheme` from Go:** Delete `DogTheme()` from production Go code. Move all theme and icon configurations to the pack/city TOML schema and resolve them dynamically.
4. **Enforce Scanner Anti-Rot:** Update the scanner test contract to fail if any path in the `role-surface.generated.yaml` allowlist does not exist on disk.
