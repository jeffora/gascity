# Camille Sato — Required Core Loading Invariant Reviewer (Attempt 7, Independent DeepSeek V4 Flash Style)

**Verdict:** block

> **Lane:** Required Core and provider pack loading, typed participation provenance, deny-by-default loaders, bypass containment, fail-closed behavior.
>
> Reviewed against the Attempt 7 design document (`.gc/design-reviews/ga-1ekw9l/attempt-7/design-before.md`, 835 lines, `updated_at: 2026-06-09T13:20:59Z`) — specifically §"Required System Pack Loader" (lines 221–311), §"Pack Registry, Cache, And Retired Source Authority" (lines 313–357), and §"Data And State" (lines 534–545).
>
> This independent review is produced using the DeepSeek V4 Flash persona, focusing specifically on first-principles trust boundaries, cross-document state consistency, and unstated runtime assumptions.

---

## Schema Conformance

Conforms to `gc.mayor.implementation-plan.v1`. Front matter carries the required keys with `phase: implementation-plan` and no `design_file`; the eight required top-level sections appear once each in the required order, and `Open Questions` is `None`. No appended attempt/review prose in the artifact.

---

## Top Strengths of the Design

- **Formalized Trust Descriptors (Gate 1 & 2 binding):** The introduction of explicit `RequiredDescriptor` and `RequiredSystemPackParticipation` structures (lines 234–243) bound by `descriptor id` provides a robust, unforgeable cryptographic linkage. Binding Gate 1 (materialized fileset check) and Gate 2 (participation check) to the same descriptor ID prevents a user-supplied pack of matching name/digest from spoofing participation.
- **AST/Type-Aware Bypass Scanner:** Upgrading the substring-based scanner to a true AST-aware and type-aware validator (lines 296–303) that rejects package aliases, multiline calls, selector values, wrappers, and function variables directly mitigates previous bypass containment vulnerabilities.
- **Centralized Snapshot Guarding:** Standardizing on a typed `RuntimeSnapshot` or `RuntimeGuard` with `RequireReady(op)` gating across behavior-changing entry points (lines 281–287) ensures compile-time and runtime type safety, preventing developer oversights where new routes might run on degraded configurations.

---

## Critical Risks & Consensus Blockers (DeepSeek V4 Flash Style)

### 1. [Blocker] Go Package Dependency Cycle Risk (Systempacks vs Config Layering Invariant)
- **The Risk:** The plan specifies that `internal/systempacks` builds a `RequiredDescriptor` and passes it to `internal/config` via `config.WithRequiredSystemPacks` (lines 234–240). During resolution, `internal/config` returns `RequiredSystemPackParticipation` records keyed by descriptor ID (lines 240–243).
- **The Impact:** Under Gas City's layering invariants, "Layer N never imports Layer N+1". `internal/config` (Primitive #4) is a lower-layer primitive, while `internal/systempacks` is a higher-level coordinator. If `RequiredDescriptor` and `RequiredSystemPackParticipation` are defined in `internal/systempacks`, then `internal/config` cannot reference them without violating layering invariants or creating a circular package import (since `internal/systempacks` already imports `internal/config` to call `LoadRuntimeCity`).
- **Required Resolution:** The types `RequiredDescriptor` and `RequiredSystemPackParticipation` must be defined within `internal/config` (or a dedicated common leaf package like `internal/config/types` or `internal/config/provenance`), rather than `internal/systempacks`. `internal/systempacks` will then import these types from the config layer, preserving the strict downward dependency hierarchy.

### 2. [Blocker] Bricked-City Bootstrapping Risk (Loader-Closed Deadlock for Setup & Repair)
- **The Risk:** The plan requires that all normal command paths and behavior-driving paths call `LoadRuntimeCity` or `LoadRuntimeCityNoRefresh` (lines 245–247). Direct loading via `config.Load*` is rejected by the AST scanner (lines 288–294). If Core is missing, corrupt, stale, or shadowing occurs, `LoadRuntimeCity` fails closed and blocks the command (lines 273–279).
- **The Impact:** When an operator attempts to run `gc init` to set up a new city, or `gc doctor --fix` to repair a corrupted Core pack fileset, these commands themselves will fail closed and crash during CLI startup/initialization because the global required-pack loader fails. This creates a permanent system deadlock where a broken city cannot be repaired because the repair tool cannot run due to the broken city.
- **Required Resolution:** Specify an explicit bootstrap/bypass protocol for initialization and repair commands. Either `gc init` and `gc doctor --fix` must be registered as specialized allowed partial-reads in the `config_loader_allowlist.yaml` (using raw config loading to perform their setup/repair safely), or `LoadRuntimeCity` must support a dedicated "bootstrap mode" parameter that permits execution when Core is unmaterialized, strictly limiting behavior to materialization and doctor fixes.

### 3. [Major] Undecidable AST Scanner Semantic Proofs vs Human-in-the-Loop Allowlisting
- **The Risk:** Line 299 states that the AST validator counts bypasses as invalid "unless the allowlist row proves the call is a partial read that cannot start sessions, expand formulas, render hooks/prompts, dispatch work, evaluate orders, or write city state."
- **The Impact:** Requiring a static AST scanner to mathematically or logically *prove* that an arbitrary direct loading call cannot cause downstream side-effects (such as session starts or state writes) is an undecidable halting-problem-like task. If developers are forced to write static analysis solvers to pass CI, it will stall development.
- **Required Resolution:** Clarify that the AST scanner is a syntactic enforcement engine: it ensures that any direct `config.Load*` call matches an approved signature, package alias, file, and function registered in `config_loader_allowlist.yaml`. The semantic proof (that the call is a harmless partial read) is a peer-reviewed, human-certified statement documented in the allowlist row's `reason` field, not a compile-time proof synthesized by the validator itself.

### 4. [Major] Concurrent Materialization Lock Contention on CLI Commands
- **The Risk:** The plan specifies that `LoadRuntimeCity` materializes, validates, and potentially repairs the required file set. In production, multiple CLI commands (such as `gc status`, `gc list`, or custom operator scripts) can be invoked concurrently.
- **The Impact:** If `LoadRuntimeCity` acquires exclusive city advisory locks or writes to disk during routine read-only commands to verify/materialize required packs, concurrent invocations will experience severe lock contention, resulting in blocked processes, slow command performance, or transient execution failures.
- **Required Resolution:** Explicitly partition the locking/writing boundary. Read-only CLI commands and status routes must use `LoadRuntimeCityNoRefresh` (which performs Gate 1 and Gate 2 validation without acquiring exclusive repair locks or writing to disk). Only the controller daemon startup, explicit `gc doctor --fix` flows, or mutating commands may acquire the exclusive materialization/repair lock in `LoadRuntimeCity`.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Which single production loader API all behavior-driving cmd/gc and config paths must use, and what test fails on direct config.Load bypasses?

**Answer:**
All production CLI commands and background worker paths must route config loading through `internal/systempacks` using either `LoadRuntimeCity` (which manages required pack materialization, fileset validation, and config resolution) or `LoadRuntimeCityNoRefresh` (used for live reload and read-only diagnostics).

The linter test **`TestGCProductionLoaderBoundary`** (or `TestProductionLoaderBypassScanner`) must fail if any production (non-test) Go file outside `internal/systempacks` imports or invokes `config.Load`, `config.LoadCity`, `config.LoadWithIncludes`, or any manual include-list assembly. Any necessary exceptions must be explicitly registered in `config_loader_allowlist.yaml` with a narrow, tested justification.

---

### Q2: Can strict validation prove Core and provider pack participation from resolved config edges, not path or digest coincidence, before any orders, prompts, formulas, scripts, or API state read behavior?

**Answer:**
**Yes, but only if anchored to the configuration resolver's output.** Gate 2 (Post-Resolution) validates the typed `RequiredSystemPackParticipation` record. To prevent "digest coincidence" or spoofing (where a user-supplied local pack named `core` has the same files or is placed at a matching path), the check must assert that the resolved config graph contains a verified, unforgeable import edge stemming from the absolute system-managed path (`.gc/system/packs/core`) and that its resolved layer ID matches the system-materialized pack layer. This check must occur immediately after config composition and before any behaviors (orders, prompts, formulas) are parsed or executed.

---

### Q3: What degraded or operator-visible path exists when Core integrity fails during live reload, and does it avoid silently continuing with stale behavior?

**Answer:**
If Core fileset or participation validation fails during a reload, `LoadRuntimeCityNoRefresh` aborts and returns a detailed validation diagnostic. The controller continues serving the current in-memory last-known-good (LKG) configuration but transitions its operational state to `read_only_degraded`. In this mode, the API rejects all mutating transactions (e.g., dispatches, formula executions, and session starts) with a clear diagnostic, avoiding silent continuation with corrupted or stale configurations on disk.

---

## Evaluation Against Lane Anti-patterns

| Anti-pattern / Red Flag | Mitigation in Current Design | Status |
| :--- | :--- | :--- |
| **Core absence is only a doctor warning after behavior already loaded** | **Excellent.** Resolved via Gate 1 pre-resolution fileset validation, which blocks config parsing entirely if Core is missing/corrupted. | **Pass** |
| **Token scanners miss wrapper, alias, function-value, or generated API loader bypasses** | **Excellent.** The plan explicitly requires AST-based/type-aware validation (lines 296–303) rather than a substring scanner. | **Pass** |
| **A corrupted or partially materialized system pack is accepted because path and digest checks are conflated** | **Excellent.** Gate 1 asserts cryptographic fileset digests against the embedded manifest before loading. | **Pass** |

---

## Final Verdict: Block

The Attempt 7 required Core loader design is highly structured, and the inclusion of explicit AST-aware scanning, `RequiredDescriptor` bindings, and unified `RuntimeGuard` status types are monumental improvements. However, because the design introduces a severe Go package **dependency cycle risk** between `internal/config` and `internal/systempacks`, and a critical **bricked-city deadlock** for bootstrap commands (`gc init` / `gc doctor`), I must **Block** the plan. Moving the descriptor and participation types to the config layer, and defining an explicit bootstrap mode or allowlist for CLI repair commands, are necessary to make this fail-closed loader robust and correct.
