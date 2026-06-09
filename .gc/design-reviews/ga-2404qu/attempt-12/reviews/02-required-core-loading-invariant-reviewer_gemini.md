# Elias Sato — Gemini (Required Core Loading Invariant Reviewer, Attempt 12, Independent DeepSeek V4 Flash Style)

**Verdict:** approve-with-risks

**Lane:** Required Core inclusion, config provenance, production loader bypass containment, loud failure on corrupt/partial Core, escape-hatch leakage.

Reviewed against the revised design document in Attempt 12 (`.gc/design-reviews/ga-2404qu/attempt-12/design-before.md` updated 2026-06-07T14:05Z) and grounded in the live codebase in `cmd/gc/`, `internal/config/`, and `internal/systempacks`.

---

## Executive Summary

As Elias Sato, the **Required Core Loading Invariant Reviewer**, I am updating my verdict to **Verdict: approve-with-risks** for Attempt 12.

The addition of the **Attempt 11 Review Resolution Contracts** (specifically the *Required Core Loading Fatal Gates* and the AST *Loader Inventory*) is an outstanding, first-principles leap forward. It structurally unifies the required-pack loading invariants and permanently seals the bypass vectors that previously degraded our containment guarantees.

However, while the structural contract is now exceptionally resilient, a rigorous cross-document and systems-level analysis reveals several critical edge cases, circular dependencies, and diagnostic contradictions that must be clarified before implementation proceeds.

---

## Top Strengths & Design Evolution (Attempt 11 to Attempt 12)

- **Two Uniform Fatal Gates**: The combination of Pre-Resolution required file-set integrity (Gate 1) and Post-Resolution typed participation proof (`RequiredSystemPackParticipation`, Gate 2) is a complete win. It directly eliminates the "Core materialized on disk but absent from resolved config" bypass, and rejects any hand-built paths or name-only checks as evidence of participation.
- **Deny-by-Default AST Scanner**: Extending the AST scanner to production `internal/` packages and explicitly banning `config.Load`, `config.LoadCity`, and aliases (instead of just `config.LoadWithIncludes`) ensures that no backdoor loading paths can creep in unnoticed.
- **Unified Loader Inventory**: Centralizing and classifying all production load paths in `loader-inventory.generated.yaml` with explicit metadata (fatality boundary, expiry slice, focused test) ensures that partial-read exceptions are strictly governed and auditable.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Do all production config resolution paths route through the system-pack wrapper so Core is included for real gc commands?

**Yes, structurally.** By enforcing a deny-by-default AST scanner that rejects direct lower-level loaders, and by routing all behavior-driving packages through the centralized `internal/systempacks` package, the design guarantees that all production paths call either `LoadRuntimeCity` or `LoadRuntimeCityNoRefresh`.

### Q2: What guard test fails if a new cmd/gc path calls config.LoadWithIncludes or another lower-level loader without required Core?

The build fails during CI due to the **AST Scanner Test** (modeled on `TestGCNonTestFilesStayOnWorkerBoundary`), which parses all production Go files in `cmd/gc` and `internal/` and rejects any unclassified direct calls to `config.Load*` or their package aliases that are not explicitly documented in the `loader-inventory.generated.yaml` allowlist.

### Q3: Does loading fail loudly on missing, corrupt, or partially materialized .gc/system/packs/core instead of relying on doctor to detect absent orders?

**Yes.** The introduction of Gate 1 (strict required file-set integrity check against embedded manifests and `pack.toml` digests) runs *before* config resolution, ensuring that any missing, corrupt, or partial Core files cause an immediate, loud load failure at startup.

---

## Critical Risks & Architectural Inconsistencies (DeepSeek V4 Flash Style)

### 1. [Major] The Beads Provider Bootstrap Paradox (Circular Dependency)

- **The Risk**: Gate 1 requires selecting and validating required host packs *before* any config is read. However, selecting provider-dependent packs (`bd` or `dolt`) requires knowing the "final effective beads provider" (defined in `city.toml` or includes). Resolving the final provider requires config resolution, which means reading config *before* Gate 1 has validated the host packs.
- **The Impact**: This is a classic bootstrapping circular dependency. If the loader resolves config to discover the provider, it reads behavior-bearing config before running Gate 1, violating Gate 1's safety guarantee. If it uses an ad-hoc pre-load raw parse, it will miss overrides or included fragments, violating correctness.
- **Recommended Action**: Explicitly require that Gate 1 validates *all* built-in packs (`core`, `bd`, and `dolt`) during the pre-resolution phase. Integrity overhead is negligible, and always validating the full embedded set eliminates the circular dependency entirely.

### 2. [Major] Stale Silent-Repair Instructions in Main Body (Doctor Contradiction)

- **The Risk**: Line 2315 in the main design body instructs the Core Presence Doctor check to call `internal/systempacks.LoadRuntimeCity`. However, `LoadRuntimeCity` is the active loader that materializes and self-heals Core. Calling it in a read-only `gc doctor` check means the check will silently mutate the filesystem and heal the corruption before reporting any error.
- **The Impact**: The corruption becomes invisible to `gc doctor`. The `--fix` option will have nothing to fix, violating the doctor safety contract and the read-only requirement of diagnostics.
- **Recommended Action**: Explicitly align line 2315 with the `Required Core Loading Fatal Gates` contract: the doctor check MUST call `LoadRuntimeCityNoRefresh` (which validates without materializing/repairing) and render the wrapper's fail-closed error as the diagnostic.

### 3. [Major] Ordinary-Command Lock Contention during Concurrent Self-Heal

- **The Risk**: The design mandates taking an advisory OS lock on the city directory descriptor to prevent self-heal contention during concurrent commands.
- **The Impact**: If a simple, fast, and high-frequency command like `gc hook` or `gc status` runs concurrently, it will contend for this lock. If the lock blocks indefinitely, a hung doctor or controller will freeze all concurrent agent processes. If it is non-blocking, it will cause spurious execution failures.
- **Recommended Action**: Define the locking strategy precisely. Self-heal execution should use a non-blocking lock with a short, deterministic timeout (e.g., 200ms). Furthermore, separate the read-only validation (which needs no lock or only a shared lock) from the active write/repair operation (which needs an exclusive lock), ensuring normal concurrent reads never contend with each other.

### 4. [Minor] Ungoverned Allowlist Aging & Metadata Fraud

- **The Risk**: The AST scanner validates that direct loader calls are registered in `loader-inventory.generated.yaml`. However, there is no automated validation of the metadata fields (such as owner, expiry slice, or focused test existence).
- **The Impact**: Over time, developers under pressure can add fake or copy-pasted rows with artificial expiries or non-existent tests, turning the allowlist into a permanent dumping ground for bypasses.
- **Recommended Action**: Specify an executable CI gate `TestAllowlistInventoryValidity` that parses `loader-inventory.generated.yaml`, asserts that every registered file/function actually exists in the AST, validates that no active row's expiry slice has passed, and verifies that the named focused test exists in the suite and is green.

### 5. [Minor] Extensibility of Provider Host Packs in Offline Environments

- **The Risk**: The design mentions "air-gapped provider self-heal for `bd` and `dolt`" using `SyntheticContentHash`.
- **The Impact**: If an operator configures a custom third-party provider that is not bundled in the binary, the offline/air-gapped self-heal contract is impossible to satisfy. The design should state whether provider host packs are strictly closed-set (limited only to bundled `bd` and `dolt`) or if an external caching mechanism is supported.
- **Recommended Action**: Clarify that provider host packs are strictly limited to the built-in, embedded set (`core`, `bd`, `dolt`), with all other third-party providers classified as external imports that do not participate in the pre-resolution fatal Gate 1.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Attempt 12 Design | Status |
| :--- | :--- | :--- |
| **Core is materialized on disk but absent from resolved config** | **Excellent.** Resolved via the uniform Post-Resolution Gate 2, which validates `RequiredSystemPackParticipation` (including layer id and import edge) after config resolution. | **Pass** |
| **Test-only no-Core escape hatch leaks into production code** | **Excellent.** All dev/test escape hatches are confined strictly to lower-level loaders in test files (`_test.go`), with `GC_BOOTSTRAP=skip` retired as a production switch. | **Pass** |
| **Core absence is detectable only by doctor after runtime behavior is already degraded** | **Excellent.** Pre-Resolution Gate 1 validates full file-set integrity and manifests before any config or asset is parsed, failing loudly at the loader boundary. | **Pass** |

---

## Final Verdict: Approved with Risks

The Attempt 12 design is a masterclass in secure and robust system-pack encapsulation, offering airtight loading boundaries and a clean separation of concerns. By addressing the deep-dive integration nuances—specifically the **provider circular dependency**, the **doctor self-healing contradiction**, and **high-frequency lock contention**—the team will deliver a bulletproof required Core loader substrate.

Proceed with implementation under these guidelines.
