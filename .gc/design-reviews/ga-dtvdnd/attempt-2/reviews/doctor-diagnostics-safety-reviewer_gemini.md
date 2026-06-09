# Faisal Khoury — DeepSeek V4 Flash (Doctor Diagnostics & Safety Review)

**Verdict:** approve-with-risks

**Scope:** Doctor diagnostics, import-state warnings, safe configuration remediation, and operator-facing messaging.

Reviewed against the Attempt 2 requirements document (`.gc/design-reviews/ga-dtvdnd/attempt-2/design-before.md` updated 2026-06-09T01:20:00Z) and existing active Go codebase.

---

## Executive Summary

The Attempt 2 Requirements Document for the Core and Gastown Pack Split shows a highly commendable, disciplined progression from Attempt 1. In the previous iteration, I blocked the requirements due to several critical, systemic issues: the lack of a degraded bootstrap mode (causing an execution paradox where `gc doctor` could not boot if Core was missing), missing automation/stream-pollution safeguards, undefined/dangerous TOML mutations during auto-repair, and shallow diagnostic scanning over the import graph.

In Attempt 2, the authors have made substantial improvements that address my primary concerns:
- **W6H and Example Mapping** sections have been drafted from scratch, establishing a testable, and robust foundation for diagnostics and remediation.
- **AC10 and AC11** systematically establish that repair is report-only by default, require non-interactive operators for mutation, enforce strict idempotence, and mandate structured JSON output for automation contexts.
- **AC11** guarantees that doctor and import-state diagnostics identify the exact config source or nested import that triggered a warning, resolving generic-diagnostic concerns.

The requirements are now structurally sound and conceptually rich. I am pleased to upgrade my verdict to **APPROVE-WITH-RISKS**. While the overall framework is strong, several residual risks and subtle edge cases must be strictly bounded in the downstream design phase.

---

## Lane-Specific Detailed Responses

### Q1: When resolved config lacks Core or references retired paths, does the diagnostic identify the exact source and explain why Core is required?

**Yes.** AC11 explicitly requires that doctor/import-state output identifies the "exact resolved config source or nested import that caused a missing-Core or retired-path condition." Additionally, Negative Path Row 1 of the Example Mapping mandates that the report "identifies the config source/layer that omitted Core" and explains that "Core is required for real cities." This directly ensures that operators are never left with generic, unhelpful error messages.

*Risk mitigation recommendation for Design:* Downstream design must verify that the config resolver retains source-attribution metadata for all loaded configuration layers (including root `city.toml`, fragments, pack imports, and environment overrides). When a dependency violation is detected, the diagnostics engine must trace this back and report the exact filepath or import chain that introduced it.

### Q2: Is any offered fix safe, idempotent, and concrete rather than merely advisory prose?

**Yes.** AC10 and AC11 ensure that the fix is concrete, non-interactive, and idempotent. However, while AC10 guarantees idempotence, the requirements still lack an explicit **non-destructiveness guarantee** for custom TOML formatting.

*Risk mitigation recommendation for Design:* The auto-fix implementation must use a comment-preserving, format-preserving AST-based TOML parser (such as `github.com/pelletier/go-toml/v2` or similar non-destructive editors). Programmatic repair must only append or modify the specific Core import block, leaving existing developer comments, key ordering, and custom indentation completely untouched. It must fail gracefully with copy-pasteable manual instructions if the configuration is read-only or immutable.

### Q3: Do doctor and import-state messages consistently distinguish required Core from optional Gastown and retired Maintenance?

**Yes.** AC12 mandates a consistent, three-way messaging model across all documentation, examples, CLI help, doctor messages, and import-state outputs:
- **Core** is described as required.
- **Gastown** is described as external and optional.
- **Maintenance** is retired as a standalone pack.
This ensures there is no confusion or blurring of required vs. optional behavior.

---

## Deep-Dive Analysis: Cross-Document Consistency & Missing Edge Cases

To ensure a flawless implementation, downstream designers must address the following critical edge cases and cross-document nuances:

### 1. The Degraded Bootstrap Mode (Resolving the Chicken-and-Egg Paradox)
- *The Risk:* Real cities require the Core pack to run (`gc` reads Core for default skills and setup). If `gc` completely refuses to initialize when Core is missing, `gc doctor` and `gc import-state` cannot execute, rendering the diagnostic and remediation loop unreachable.
- *The Solution:* The CLI configuration loader must implement a degraded "bootstrap mode." When running diagnostic or repair commands (`gc doctor`, `gc import-state`), the CLI must bypass full, strict pack loading and template rendering, allowing the runtime to boot just enough to analyze the configuration, report the missing Core entry, and apply the repair.

### 2. Diagnostic Behavior Under the Dev/Test Escape Hatch (AC2)
- *The Risk:* AC2 introduces a "dev/test escape hatch" to construct partial (Core-less) configs for unit and integration testing. If this escape hatch is active, a naive `gc doctor` or import-state check might report a "missing Core" error, generating false-positive test failures.
- *The Constraint:* The design must specify how the diagnostics engine behaves when the escape hatch is active (e.g., via an environment variable like `GC_TEST_ESCAPE_HATCH=1`). The doctor and import-state checks must detect this state and either suppress the missing-Core warning entirely or downgrade it to a non-blocking informational note.

### 3. Recursive Scanning of Deep Import Chains
- *The Risk:* A legacy import reference might be nested multiple layers deep (e.g., `city.toml` -> imports `pack-a` -> imports `pack-b` -> imports `packs/maintenance`). A shallow, single-level scanner would fail to detect this.
- *The Constraint:* The diagnostics engine must perform a recursive scan of the fully resolved import graph. When a legacy or retired path is found, it must print the complete path chain (e.g., `[city.toml -> pack-a -> pack-b -> packs/maintenance]`) so the operator can locate and remediate the nested dependency.

### 4. Live Process Table Discovery for In-Flight Sessions
- *The Risk:* If an existing city has in-flight sessions running under retired `packs/maintenance` or in-tree Gastown, executing a repair command while background processes are active could lead to state corruption.
- *The Constraint:* In accordance with the key design principle: **No status files — query live state**. The repair command must discover running legacy sessions directly from the process table (`ps`, `lsof`, or tmux socket queries) and gracefully terminate them before isolating the directory, rather than relying on stale PID files.

---

## Verdict & Transition to Design

**Verdict: APPROVE-WITH-RISKS**

The Requirements Document is fully approved to transition to the **design and implementation-plan** phases. The open questions and residual risks identified above are minor design details that can be resolved as part of the initial design draft.
