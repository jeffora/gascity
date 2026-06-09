# Alistair Sterling - DeepSeek V4 Flash (ZFC & Role-Neutrality Guardian)

**Lane:** zfc-role-neutrality-guardian (wave 1) — zero hardcoded roles, Core role neutrality, `dog` configurability, SDK self-sufficiency.

**Verdict:** approve

Reviewed the Attempt 8 Requirements (`requirements.md` updated at 2026-06-09T15:35:47Z) and the corresponding Implementation Plan (`implementation-plan.md` updated at 2026-06-09T13:20:59Z) strictly through the lens of zero hardcoded roles, Core role neutrality, dynamic configurability, and SDK self-sufficiency, grounded against the live repository tree.

---

## Executive Summary

The Attempt 8 requirements represent a complete and robust specification for securing Gas City's role-neutrality and SDK self-sufficiency. All core concepts conform perfectly to the ZFC (Zero Framework Cognition) rules and the Bitter Lesson:
- Legacy in-tree role behavior has been fully extracted to the external public pack authority (`gascity-packs/gastown`).
- Core is established as the sole required host-side pack, carrying only SDK-generic infrastructure and leaving all role definitions to user-supplied configuration.
- Safety-critical maintenance operations are confined natively within the controller, leaving LLM-executed maintenance purely optional and configurable.

I award a definitive **APPROVE** verdict. However, to prevent downstream ZFC erosion and ensure a flawless, bulletproof implementation, the following critical technical warnings, cross-document inconsistencies, and implementation-level edge cases must be strictly addressed during the upcoming slices.

---

## Lane-Specific Detailed Responses

### Q1: Does any requirement introduce role-conditional logic or hardcoded role assumptions into Core assets or Go business logic?
**No.** The requirements (AC2, AC8, AC9) explicitly prohibit role-conditional logic or hardcoded role assumptions in Core assets or Go business logic. The inclusion of an automated role-surface scanner with positive and negative controls (AC8) ensures that any accidental role-name leakage will immediately fail the build.

### Q2: Is the default dog maintenance agent specified as replaceable pack configuration rather than a Go-side exception?
**Yes.** AC9 specifies that the default `dog` executor is declared solely as a default configuration value inside the Core pack's configuration data, and Go code must treat it as user-supplied config. Go-side hardcoded fallbacks to `"dog"` are explicitly forbidden.

### Q3: Can SDK infrastructure still operate with only the controller and Core config, without assuming Gastown agents exist?
**Yes.** The requirements mandate that safety-critical, structural, and deterministic maintenance tasks remain controller-owned and run natively in Go. LLM-executed maintenance work uses symbolic configurable bindings and is treated as optional; if no executor is available, the city loads and operations run smoothly without causing Go-side exceptions or failures.

---

## Deep-Dive Analysis: Cross-Document Consistency & Missing Edge Cases

Acting as an independent DeepSeek V4 Flash voice, I highlight the following critical inconsistencies and edge cases between the Requirements and the Implementation Plan:

### 1. Precedence Conflict: Environment Injection vs. Core Pack Defaults
- **The Inconsistency:** The Implementation Plan states that "environment injection may supply a default only when neither city nor pack config names a binding" (lines 441–443). However, AC9 specifies that the required Core pack always ships a default configured executor named `dog` in its configuration data (declared in `[system_packs.core.bindings]`). Because Core is required and always loaded, this pack-level default will *always* be present. Under the proposed precedence rule, the environment variable `GC_CORE_MAINTENANCE_WORKER` will *never* be able to override anything, rendering the environment override completely dead-code.
- **Recommendation:** Revise the precedence chain in the bindings parser so environment variable overrides have a higher priority than pack-level defaults, but a lower priority than city-scoped configuration. The correct order of resolution must be:
  `city.toml [gc.bindings.*]` ➔ `Environment Variable Override` ➔ `pack.toml [bindings] defaults`.

### 2. Hardcoded Env Override `GC_CORE_MAINTENANCE_WORKER` is a ZFC Leak
- **The Inconsistency:** The Implementation Plan explicitly references and parses a literal, hardcoded environment variable named `GC_CORE_MAINTENANCE_WORKER` (lines 434, 441). By explicitly compiling a role-specific environment variable name in Go, the Go codebase retains role-specific cognition. If the design is truly zero-hardcoded-roles, the Go code should have zero awareness of specific role names or role-specific environment variables.
- **Recommendation:** Prohibit role-specific hardcoded environment variables in Go. Replace `GC_CORE_MAINTENANCE_WORKER` with a generic, dynamic environment override mechanism: any environment variable prefixed with `GC_BINDINGS_` (e.g., `GC_BINDINGS_MAINTENANCE_WORKER`) is dynamically parsed and mapped to its corresponding binding. This is completely ZFC-compliant and scales infinitely without Go-side modifications when new bindings are introduced.

### 3. Data-Driving Binding Optionality (Preventing Go-Side Special Casing)
- **The Inconsistency:** AC9 states that "optional LLM-executed maintenance uses symbolic configurable bindings whose required and optional keys are declared in pack data... escalation behavior, optionality, and override mechanisms are declared in data." However, the implementation plan lacks any concrete details on how this optionality schema is declared or verified. If the Go dispatcher/engine has to hardcode which bindings are "optional" vs. "required," this introduces Go-side special-casing of roles.
- **Recommendation:** Ensure that the formula, order, or pack configuration files explicitly declare the optional/required status of their expected bindings as metadata (e.g., `[bindings.maintenance_worker] optional = true`), allowing the engine to evaluate compliance data-drivenly and preventing any Go-side special casing.

### 4. Verification of User-Defined Bindings in `gc doctor`
- **The Edge Case:** If an operator renames or overrides the default `dog` executor in `city.toml` (e.g., mapping `maintenance_worker = "cleanup-agent"`), a typo in the binding name (e.g., `maintenance_workr`) will result in silent GUPP failures or failed dispatches because the dispatcher will search for an agent that does not exist.
- **Recommendation:** The `gc doctor` and `gc import-state` commands must validate that any user-configured bindings in `city.toml` map to agents that are actually declared in the active pool of configured agents, reporting a clear diagnostic with a stable condition code if a binding points to an undefined agent.

---

## Required Changes for Implementation Slices

1. **Fix Precedence Resolution for Environment Variables:** Revise the precedence chain so environment variable overrides have higher priority than pack-level defaults.
2. **Generalize Environment Overrides (ZFC Compliance):** Replace the hardcoded `GC_CORE_MAINTENANCE_WORKER` environment variable with a generic, dynamic `GC_BINDINGS_<NAME>` parsing mechanism to prevent role cognition in Go.
3. **Data-Drive Binding Optionality:** Specify that the optional/required status of all bindings must be declared as metadata in the formula, order, or pack configurations, rather than being determined via Go-side special-casing.
4. **Binding Validation in `gc doctor`:** Mandate that `gc doctor` and `gc import-state` validate user-configured bindings against the active pool of configured agents.

---

## Verdict & Transition to Implementation

**Verdict: APPROVE**

The Requirements Document is fully approved to transition to the **design and implementation-plan** phases.
