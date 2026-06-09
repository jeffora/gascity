# Alistair Sterling - DeepSeek V4 Flash (ZFC & Role-Neutrality Guardian)

**Lane:** zfc-role-neutrality-guardian (wave 1) — zero hardcoded roles, Core role neutrality, `dog` configurability, SDK self-sufficiency.

**Verdict:** approve-with-risks

Reviewed the Attempt 9 Requirements (`requirements.md` updated at 2026-06-09T17:23:58Z) and the corresponding Implementation Plan (`implementation-plan.md` updated at 2026-06-09T13:20:59Z) strictly through the lens of zero hardcoded roles, Core role neutrality, dynamic configurability, and SDK self-sufficiency, grounded against the live repository tree.

---

## Executive Summary

The Attempt 9 requirements introduce a critical architectural evolution: **shifting the Core pack end-state authority from `internal/bootstrap/packs/core` to `internal/packs/core`**, while designating the old bootstrap location as a legacy source root to be isolated or deleted. This is a massive victory for structural clarity and clean packaging, preventing the compile-time coupling of SDK core behaviors to bootstrapping mechanisms.

However, from the strict perspective of **Zero Framework Cognition (ZFC)** and the **Bitter Lesson**, several crucial implementation risks, precedence conflicts, and design leaks remain unresolved across the requirements and implementation plan. Specifically, the implementation plan carries a role-specific environment leak, a precedence conflict that renders environment variable overrides dead-code, and a "split-authority" transitional risk during the folder migration.

To protect the SDK against downstream role cognition and ensure a bulletproof migration, I award an **APPROVE-WITH-RISKS** verdict. The transition to implementation is approved, but the critical issues detailed below must be resolved in their respective implementation slices.

---

## Lane-Specific Detailed Responses

### Q1: Does any requirement introduce role-conditional logic or hardcoded role assumptions into Core assets or Go business logic?
**No in Requirements; Yes in Implementation Plan.**
The Attempt 9 Requirements (AC2, AC8, AC9) are role-neutral and explicitly prohibit role-conditional logic. However, the Implementation Plan (lines 432–435) introduces a ZFC leak by explicitly referencing and parsing a hardcoded, role-specific environment variable: `GC_CORE_MAINTENANCE_WORKER`. Compiling literal role-specific environment variables directly into Go business logic leaks role cognition to the framework.

### Q2: Is the default dog maintenance agent specified as replaceable pack configuration rather than a Go-side exception?
**Yes.**
AC9 specifies that the required Core pack may ship a default configured executor named `dog` purely as inert pack-level configuration data (`[system_packs.core.bindings]`), and Go code must treat it as user-supplied config. There are no Go-side fallbacks or hardcoded exceptions permitted for `dog`.

### Q3: Can SDK infrastructure still operate with only the controller and Core config, without assuming Gastown agents exist?
**Yes.**
AC2 and Happy-path 1 mandate that safety-critical, structural, and deterministic maintenance tasks are controller-owned and run natively in Go. LLM-executed maintenance work uses symbolic configurable bindings and is optional. If no executor/agent is available, the city loads and operates smoothly. However, the specific set of "SDK infrastructure operations" (such as gate evaluation, health patrol, bead lifecycle, order dispatch) must be exhaustively verified via a dedicated controller-only test surface to ensure no hidden role assumptions exist.

---

## Deep-Dive Analysis: Cross-Document Consistency & Missing Edge Cases

Acting as an independent DeepSeek V4 Flash voice, I highlight the following critical inconsistencies and edge cases between the Requirements and the Implementation Plan:

### 1. Precedence Conflict: Environment Injection vs. Core Pack Defaults
* **The Inconsistency:** The Implementation Plan (lines 439–443) specifies that "environment injection may supply a default only when neither city nor pack config names a binding." However, AC9 specifies that the required Core pack *always* ships a default configured executor named `dog` in its configuration data (`[system_packs.core.bindings]`). Because Core is mandatory and always loaded, this pack-level default will *always* be present in resolved config. Under the proposed precedence rule, the environment variable override will *never* fire, rendering it completely dead-code!
* **Recommendation:** Revise the precedence chain in the bindings parser so environment variable overrides have higher priority than pack-level defaults, but lower priority than city-scoped configuration. The correct order of resolution must be:
  `city.toml [gc.bindings.*]` ➔ `Environment Variable Overrides` ➔ `pack.toml [bindings] defaults`.

### 2. Hardcoded Env Override `GC_CORE_MAINTENANCE_WORKER` is a ZFC Leak
* **The Inconsistency:** The Implementation Plan explicitly references and parses a literal environment variable named `GC_CORE_MAINTENANCE_WORKER` (lines 434, 441). This compiles role-specific naming directly into Go. If the SDK is truly zero-hardcoded-roles, the Go code should have zero awareness of specific role names or role-specific environment overrides.
* **Recommendation:** Replace `GC_CORE_MAINTENANCE_WORKER` with a generic, dynamic environment override mechanism: any environment variable prefixed with `GC_BINDINGS_` (e.g., `GC_BINDINGS_MAINTENANCE_WORKER`) is dynamically parsed and mapped to its corresponding binding. This remains 100% ZFC-compliant and scales infinitely without Go-side modifications when new bindings are introduced.

### 3. Split-Authority Risk: Transitioning from `internal/bootstrap/packs/core` to `internal/packs/core`
* **The Edge Case:** Shifting the Core authority path is an excellent development, but introduces a major risk of "dual-active folders" or transitional leaks. During implementation slices 3 through 6, if old paths remain on disk or are partially imported, the system could silently fall back to legacy bootstrap assets, making test results misleading.
* **Recommendation:** Ensure that throughout Slices 3 to 6, a strict import scanner and runtime participation validator block any resolution or fallback to the legacy `internal/bootstrap/packs/core`. Deletion of the old tree must be the final compile-time proof of isolation.

### 4. Data-Driving Binding Optionality (Preventing Go-Side Special Casing)
* **The Edge Case:** AC9 states that "optional LLM-executed maintenance uses symbolic configurable bindings whose required and optional keys are declared in pack data... escalation behavior, optionality, and override mechanisms are declared in data." However, the implementation plan lacks details on how optionality is declared. If the Go dispatcher/engine has to hardcode which bindings are "optional" vs. "required," this introduces Go-side special-casing of roles.
* **Recommendation:** Ensure that the formula, order, or pack configuration files explicitly declare the optional/required status of their expected bindings as metadata (e.g., `[bindings.maintenance_worker] optional = true`), allowing the engine to evaluate compliance data-drivenly and preventing any Go-side special casing.

### 5. Verification of User-Defined Bindings in `gc doctor`
* **The Edge Case:** If an operator renames or overrides the default `dog` executor in `city.toml` (e.g., mapping `maintenance_worker = "cleanup-agent"`), a typographical or mapping error in the binding name (e.g., `cleanup-agentt`) will result in silent GUPP failures because the dispatcher will search for an agent that does not exist.
* **Recommendation:** The `gc doctor` and `gc import-state` commands must validate that any user-configured bindings in `city.toml` map to agents that are actually declared in the active pool of configured agents, reporting a clear diagnostic with a stable condition code if a binding points to an undefined agent.

---

## Required Changes for Implementation Slices

1. **Fix Precedence Resolution for Environment Variables (Slice 4a):** Correct the precedence chain so environment variable overrides have a higher priority than pack-level defaults but are subordinate to city-scoped configuration.
2. **Generalize Environment Overrides (Slice 4a):** Replace the hardcoded `GC_CORE_MAINTENANCE_WORKER` environment variable with a generic, dynamic `GC_BINDINGS_<NAME>` parsing mechanism to prevent role cognition in Go.
3. **Prevent Split-Authority Transitional Fallbacks (Slice 3):** Ensure that throughout the folder transition, strict import scanners and runtime participation validators block any resolution or fallback to the legacy `internal/bootstrap/packs/core`.
4. **Data-Drive Binding Optionality (Slice 4a):** Specify that the optional/required status of all bindings must be declared as metadata in the formula, order, or pack configurations, rather than being determined via Go-side special-casing.
5. **Binding Validation in `gc doctor` (Slice 4b):** Mandate that `gc doctor` and `gc import-state` validate user-configured bindings against the active pool of configured agents.

---

## Verdict & Transition to Implementation

**Verdict: APPROVE-WITH-RISKS**

The Attempt 9 Requirements and Implementation Plan represent a highly robust and secure design. The shift of the Core authority path to `internal/packs/core` significantly strengthens the SDK's design. Moving to the implementation phase is approved, provided that the critical risks identified above are fully addressed during the upcoming development slices.
