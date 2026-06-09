# Camille Sato — Required Core Loading Invariant Reviewer (Attempt 8, Independent DeepSeek V4 Flash Style)

**Verdict:** approve-with-risks

> **Lane:** Required Core and provider pack loading, typed participation provenance, deny-by-default loaders, bypass containment, fail-closed behavior.
>
> Reviewed against the Attempt 8 design document (`.gc/design-reviews/ga-1ekw9l/attempt-8/design-before.md`, 835 lines, `updated_at: 2026-06-09T13:20:59Z`) — specifically §"Required System Pack Loader" (lines 221–312), §"Pack Registry, Cache, And Retired Source Authority" (lines 313–357), and §"Data And State" (lines 534–545).
>
> This independent review is produced using the DeepSeek V4 Flash persona, focusing specifically on first-principles trust boundaries, cross-document state consistency, and unstated runtime assumptions.

---

## Schema Conformance

Conforms to `gc.mayor.implementation-plan.v1`. Front matter carries the required keys with `phase: implementation-plan` and no `design_file`; the eight required top-level sections appear once each in the required order, and `Open Questions` is `None`. No appended attempt/review prose in the artifact.

---

## Top Strengths of the Design

- **Resolution of Circular Dependency:** By defining `RequiredDescriptor` and passing them via a typed option (`config.WithRequiredSystemPacks`), and returning `RequiredSystemPackParticipation` records keyed by descriptor ID, the design avoids upward dependency violations. Higher-level `internal/systempacks` coordinates and calls `internal/config` (Layer 0/1) cleanly.
- **AST/Type-Aware Linter and Allowlist:** The move to a type-aware AST scanner (lines 296–303) that checks for wrapper functions, multiline calls, aliases, and variable selector values is a massive security improvement. It completely mitigates the risk of direct `config.Load*` bypasses.
- **Graceful Bricked-City Deadlock Mitigation:** The distinction between behavior-changing entry points (requiring `ready` mode via `RequireReady`) and read-only status or administrative/repair commands (such as `gc doctor --fix` or `gc init` which can execute under `blocked` or `read_only_degraded` snapshots) resolves the bootstrapping deadlock while maintaining fail-closed operational safety.
- **Fail-Closed Live Reload:** When a reload fails Core integrity, keeping the last-known-good (LKG) config exclusively for read-only diagnostics while blocking behavior-changing dispatch, prompts, or session starts (lines 273–280) avoids the silent-continue hazard.

---

## Critical Risks & Remaining Gaps

### 1. [Major] The Bootstrapping Paradox of Pre-Resolution Provider Selection
- **The Risk:** In Gate 1 (lines 259–261), the system must validate the materialized fileset for Core plus provider packs (specifically `bd` and `dolt`) *before* config resolution. However, the selected beads provider is declared in the city's `city.toml` (or environment).
- **The Impact:** To know which provider pack to validate, the loader must parse `city.toml`. But under a deny-by-default model, we cannot trust or parse `city.toml` until the filesets driving the system are validated. If we parse `city.toml` first, we open a trust-boundary gap. If we don't, we have no secure way to know which provider fileset to validate.
- **Required Resolution:** The design should clarify that Gate 1 unconditionally materializes and validates the filesets for *all* registered/built-in provider packs (`bd` and `dolt`), without reading `city.toml` first. During Gate 2 (Post-Resolution), the config resolution graph's import edges will determine which provider actually participated, filtering out the inactive provider and confirming the active one's participation.

### 2. [Minor] AST Allowlist Scope-Creep Vulnerability
- **The Risk:** The AST scanner ensures that direct `config.Load*` bypasses are confined to the `config_loader_allowlist.yaml` (lines 296–303). However, the allowlist check is static. 
- **The Impact:** If a developer registers an exception for a source file (e.g., `cmd/gc/status.go`) because it only performs a partial read, there is no static constraint preventing that file from later being modified to include behavior-changing operations (such as starting a session) while still utilizing the bypass.
- **Required Resolution:** The AST scanner should statically enforce that any function/file in the `config_loader_allowlist.yaml` does not invoke mutating runtime actions (e.g., `RequireReady`, `worker.Create`, or session/dispatch triggers). Alternatively, a runtime guard should throw an error if a behavior-changing mutation is attempted on a `config.City` instance that lacks a verified participation record from `LoadRuntimeCity`.

### 3. [Minor] Indefinite Degraded State (Stale LKG Config)
- **The Risk:** If Core fileset validation fails during a live reload, `LoadRuntimeCityNoRefresh` transitions the controller into `read_only_degraded` mode, keeping the last-known-good (LKG) configuration active for status reporting.
- **The Impact:** While this prevents corrupt config execution, running on an increasingly stale in-memory LKG config indefinitely can lead to drift and confusion if the on-disk state is broken for a prolonged period.
- **Required Resolution:** Introduce a warning-escalation threshold or time-to-live (TTL) for the in-memory LKG configuration. If the system remains in `read_only_degraded` mode due to persistent reload failures for more than a configured duration (e.g., 10 minutes), the controller should publish a high-priority diagnostic event to alert operators to the prolonged out-of-sync state.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Which single production loader API all behavior-driving cmd/gc and config paths must use, and what test fails on direct config.Load bypasses?

**Answer:** 
All production command and worker paths must load configs via the `internal/systempacks` API using either `LoadRuntimeCity` or `LoadRuntimeCityNoRefresh`. They are strictly prohibited from calling direct loaders like `config.Load` or assembling includes manually.

The AST scanner test **`TestProductionLoaderBypassScanner`** (or package equivalent) will fail if any production Go code outside `internal/systempacks` bypasses this API, unless the call site is explicitly registered in `internal/systempacks/testdata/config_loader_allowlist.yaml` with a verified partial-read justification.

---

### Q2: Can strict validation prove Core and provider pack participation from resolved config edges, not path or digest coincidence, before any orders, prompts, formulas, scripts, or API state read behavior?

**Answer:**
**Yes.** This is achieved during Gate 2 (Post-Resolution) by checking the typed `RequiredSystemPackParticipation` records. Rather than relying on simple file path checks or digest coincidence (which a user-supplied pack of matching name/digest could spoof), the verification checks that the resolved config graph contains an unforgeable import edge stemming from the system-controlled materialization root and that the resolved layer ID matches the system-materialized pack layer. This check runs immediately after config composition and before any behaviors (prompts, formulas, scripts) are parsed or executed.

---

### Q3: What degraded or operator-visible path exists when Core integrity fails during live reload, and does it avoid silently continuing with stale behavior?

**Answer:**
If Core integrity fails during live reload, `LoadRuntimeCityNoRefresh` returns detailed validation diagnostics. The controller retains the last-known-good (LKG) config in memory but transitions into `read_only_degraded` mode. In this mode, behavior-changing entry points (dispatch, formula expansion, order evaluation, hook rendering, prompt resolution, session starts) are blocked or paused. Both the API and CLI surface the diagnostics instead of silently continuing with stale or corrupted disk configuration.

---

## Evaluation Against Lane Anti-patterns

| Anti-pattern / Red Flag | Mitigation in Current Design | Status |
| :--- | :--- | :--- |
| **Core absence is only a doctor warning after behavior already loaded** | **Excellent.** Blocked completely by Gate 1 pre-resolution fileset validation. | **Pass** |
| **Token scanners miss wrapper, alias, function-value, or generated API loader bypasses** | **Excellent.** Upgraded to an AST/type-aware static validator (lines 296–303). | **Pass** |
| **A corrupted or partially materialized system pack is accepted because path and digest checks are conflated** | **Excellent.** Conflation is avoided by separating Gate 1 (cryptographic fileset digests) from Gate 2 (config graph participation). | **Pass** |

---

## Final Verdict: Approve-with-Risks

The Attempt 8 design resolves all major blockers from previous iterations, particularly circular dependency package layering risks, doctor repair deadlocks, and linter evasion vectors. It is a highly robust and secure architecture. Adopting the recommended mitigations for the provider selection paradox and AST allowlist scope constraints will fully close the remaining trust boundaries.
