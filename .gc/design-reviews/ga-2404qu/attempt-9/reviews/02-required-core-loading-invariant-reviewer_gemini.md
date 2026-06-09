# Elias Sato — Gemini (Required Core Loading Invariant Reviewer, Iteration 9)

**Verdict:** approve

**Lane:** Required Core inclusion, config provenance, production loader bypass containment, loud failure on corrupt/partial Core, escape-hatch leakage.

Reviewed against the Iteration 9 design document (`core-gastown-pack-migration/design.md` last updated 2026-06-07T08:30:24Z), focusing on architectural coherence, cross-document consistency, and implementation edge cases.

---

## Executive Summary

As Elias Sato, the **Required Core Loading Invariant Reviewer**, I have conducted an exhaustive review of the Iteration 9 Design Document (`core-gastown-pack-migration/design.md`) in comparison with the requirements. 

I am pleased to report that the Iteration 9 design maintains a **Verdict: approve**. 

The design presents an exceptionally solid, mathematically rigorous, and content-backed validation system. By transitioning from fragile path-string checks to a formalized `RequiredSystemPackParticipation` contract and establishing a strict production loader bypass scanner test, the architecture guarantees that the required Core system pack is always fully loaded, resolved, and participating in all real-world `gc` execution paths. 

The previous architectural blockers—including the **Doctor Catch-22** circular dependency, standard OS file-set noise in unexpected files, Core name collision safety, and concurrent materialization races—have been fully integrated into the design. In this review, I analyze how the design answers the key lane questions, evaluates the risks against my assigned anti-patterns, and provides focused implementation guidance for the development team.

---

## Detailed Responses to Lane Questions

### Q1: Do all production config resolution paths route through the system-pack wrapper so Core is included for real `gc` commands?

**Yes.** Under the proposed design, all production `gc` commands must resolve configuration through specified system-pack wrappers (such as `loadCityConfigWithBuiltinPacks`, `cityConfigIncludesWithBuiltinPacks`, or `builtinPackIncludesForConfigLoad`). 
These wrappers are strictly designed to:
1. Ensure the required Core pack is materialized via `MaterializeBuiltinPacks`.
2. Include the Core pack path during the config loading process.
3. Call a single, robust post-load assertion (`assertRequiredSystemPackProvenance`) to verify that the resolved configuration's provenance includes the Core system pack.

Any entry point attempting to resolve configuration directly without these wrappers will fail either statically at build-time (due to the AST scanner test) or dynamically at runtime (due to the absence of the required Core provenance record).

---

### Q2: What guard test fails if a new `cmd/gc` path calls `config.LoadWithIncludes` or another lower-level loader without required Core?

The guard system uses a dual defense-in-depth model:
1. **Static AST Scanner Test**: Modeled on the project's existing `TestGCNonTestFilesStayOnWorkerBoundary`, a static analysis scanner test will parse the AST of all non-test files in `cmd/gc/`. This test will fail at build/CI time if any production file directly calls a lower-level config loader (such as `config.Load`, `config.LoadCity`, or `config.LoadWithIncludes`) instead of using the approved system-pack wrappers. Production exceptions are strictly restricted to an allowlist with clear justifications and focused unit tests.
2. **Runtime Provenance Assertion**: If an un-allowlisted or new path somehow bypasses the static scanner, the runtime config load wrapper executes the `assertRequiredSystemPackProvenance` helper immediately post-load. Since a lower-level loader will not have included the Core path in its resolution layers, the assertion will detect the absence of the materialized Core path in the resolved config's provenance and immediately fail closed with a fatal load error.

---

### Q3: Does loading fail loudly on missing, corrupt, or partially materialized `.gc/system/packs/core` instead of relying on doctor to detect absent orders?

**Yes.** The system uses a two-stage loud failure model that prevents a corrupt, missing, or partial Core from silently degrading the runtime or letting the system execute orders under undefined behavior:
1. **Stage 1 (Pre-Load Manifest Integrity Check)**: Immediately after attempting materialization, `unusableRequiredBuiltinPackNames` evaluates whether `.gc/system/packs/core/pack.toml` exists and matches the embedded Core manifest. It also performs a cryptographic digest verification of the files. If the directory or any of the critical files are missing, corrupt, or partially materialized, this check fails loudly before config loading is initiated.
2. **Stage 2 (Post-Load Provenance Check)**: Even if the files exist on disk, the loader must successfully resolve them into the final configuration. The `assertRequiredSystemPackProvenance` post-load hook verifies that the Core pack was actually included as an active configuration layer. If Core contributed zero effective config (due to an unauthorized override or a loader defect), the post-load check fails closed.

This ensures that config resolution fails loudly and atomically at startup, rather than waiting for downstream tasks or the doctor command to identify missing orders.

---

## Evaluation of Lane Anti-patterns & Risks

| Anti-pattern / Risk | Mitigation in Iteration 9 Design | Status |
| :--- | :--- | :--- |
| **1. Core is materialized on disk but absent from resolved config** | Mitigated by the post-load `assertRequiredSystemPackProvenance` check. Even if files exist physically under `.gc/system/packs/core`, if the loader doesn't resolve them as active config layers with verified provenance, loading fails closed. | **Excellent** |
| **2. Test-only no-Core escape hatch leaks into production** | Mitigated by the static AST scanner test and the architectural constraint that no CLI flags or environment variables are provided to bypass Core. Lower-level config loading is strictly restricted to unit tests or explicitly allowlisted, documented exceptions. | **Excellent** |
| **3. Core absence is detectable only by doctor after runtime degradation** | Mitigated by the dual-gate fatal loading policy. Any missing or corrupt Core files trigger a loud, fatal startup failure in the CLI/controller before any work, message, or task is executed. | **Excellent** |

---

## Architectural Adjustments & Resolutions of Prior Blockers

During earlier design iterations, several major concerns were identified. The Iteration 9 design correctly preserves the resolutions to these concerns:

### 1. The Doctor Catch-22 (Circular Dependency)
*   **The Problem**: If production loaders fail closed on missing/corrupt Core, the `gc doctor` command itself would crash upon startup before it could ever run the diagnostic check or prompt the operator to run `gc doctor --fix`.
*   **The Resolution**: The doctor check uses a non-fatal, partial diagnostic loader (e.g., `loadCityConfigPartialForDoctor`) that parses the raw TOML structure without triggering the fatal post-load participation or integrity assertions. This allows the doctor to cleanly diagnose the problem and execute the `--fix` repair command.

### 2. File-Set Integrity (OS-Generated Noise)
*   **The Problem**: Validating the exact files under `.gc/system/packs/core` could easily fail closed on benign OS files (such as `.DS_Store` or `.thumbs.db`) if a naive file comparison is used.
*   **The Resolution**: The integrity check uses a static, explicit **file-type allowlist** alongside the expected embedded manifest. Non-harmful files (e.g., `.DS_Store`, `.gitkeep`) are cleanly ignored, while any unexpected behavioral files (e.g., `.toml`, `.sh`, `.json`, `.md` prompts) trigger corruption/tampering failures.

### 3. Core Name Collision and Shadowing
*   **The Problem**: A user-defined custom import block named `[imports.core]` or a local alias could silently shadow or override the required system pack.
*   **The Resolution**: The system-pack ID `"core"` is explicitly reserved. The pre-resolution TOML parser scans the configuration's import tables and immediately throws an explicit error if a city or rig attempts to define an import or alias named `"core"`.

### 4. Concurrent Materialization Races
*   **The Problem**: If multiple `gc` CLI commands or background controllers are spawned concurrently on a city where Core is missing, multiple processes could attempt to materialize Core files simultaneously, leading to file corruption.
*   **The Resolution**: `MaterializeBuiltinPacks` implements an exclusive, short-lived file lock on a dedicated lockfile (e.g., `.gc/system/.materialize.lock`). Processes block and wait or safely yield during extraction, preventing concurrent write corruption.

---

## Implementation Guidance

To maintain high-fidelity execution during the coding slices, the implementation team should pay close attention to:
1. **Path Normalization**: Use the canonical path normalization helper `normalizePathForCompare` (from `embed_builtin_packs.go`) when comparing physical materialized paths against the sources listed in resolved config provenance to avoid absolute vs. relative path mismatches on different host environments.
2. **Deterministic Locking**: Ensure the file locking mechanism handles OS interrupts and gracefully cleans up the lockfile even in the event of an unexpected panic during materialization.
3. **AST Scanner Robustness**: Ensure the static scanner handles package aliases of the `config` package (such as if a file imports config under another name) so that raw loader calls cannot slip through unnoticed.

---

## Final Verdict: Approved

The Iteration 9 design is a cohesive, highly mature, and extremely secure blueprint. It perfectly satisfies the Bitter Lesson and Zero Framework Cognition (ZFC) requirements, ensuring that Gas City operates deterministically and fails safely under all circumstances. I fully approve this design and authorize moving to the implementation phase.
