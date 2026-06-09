# Camille Sato — Gemini (Required Core Loading Invariant Reviewer, Attempt 4, Independent DeepSeek V4 Flash Style)

**Verdict:** approve-with-risks

> **Lane:** Required Core and provider pack loading, typed participation provenance, deny-by-default loaders, bypass containment, fail-closed behavior.
>
> Reviewed against the Attempt 4 design document (`.gc/design-reviews/ga-1ekw9l/attempt-4/design-before.md`, 657 lines, `updated_at: 2026-06-09T07:28:00Z`) — §"Required System Pack Loader" (193–251), §"Pack Registry, Cache, And Retired Source Authority" (252–285), §"Bootstrap Fixture Isolation" (370–396), and §"Data And State" (426–486).
>
> This independent review is produced using the DeepSeek V4 Flash persona, focusing specifically on first-principles trust boundaries, cross-document state consistency, and unstated runtime assumptions.

## Executive Summary
This iteration (Attempt 4) significantly strengthens the required-pack loading architecture by splitting the rollout into more granular slices (Slices 1a, 1b, 1c, 4a, 4b, 4c, 5a, 5b) and incorporating the concrete fail-closed reload semantics demanded in prior iterations. However, several first-principles trust-boundary gaps, bootstrapping paradoxes, and developer experience (DX) risks remain unaddressed in the core loader path. Specifically, the circular dependency between pre-resolution validation and provider selection, the vulnerability of the substring-based bypass scanner to AST-level circumvention, and the volatility of in-memory LKG states during process crashes pose serious system-integrity concerns. 

This review provides evidence-backed analysis and concrete resolutions to close these gaps and deliver an airtight, fail-closed system loader.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Which single production loader API all behavior-driving cmd/gc and config paths must use, and what test fails on direct config.Load bypasses?
**Answer:**
All production CLI commands and background worker paths must route config loading through `internal/systempacks` using either `LoadRuntimeCity` (which manages required pack materialization, fileset validation, and config resolution) or `LoadRuntimeCityNoRefresh` (used for live reload and read-only diagnostics).
The static linter test **`TestGCProductionLoaderBoundary`** (or `TestProductionLoaderBypassScanner`) must fail if any production (non-test) Go file outside `internal/systempacks` imports or invokes `config.Load`, `config.LoadCity`, `config.LoadWithIncludes`, or any manual include-list assembly. Any necessary exceptions must be explicitly registered in `loader-inventory.generated.yaml` with a narrow, tested justification.

### Q2: Can strict validation prove Core and provider pack participation from resolved config edges, not path or digest coincidence, before any orders, prompts, formulas, scripts, or API state read behavior?
**Answer:**
**Yes, but only if anchored to the configuration resolver's output.** Gate 2 (Post-Resolution) validates the typed `RequiredSystemPackParticipation` record. To prevent "digest coincidence" or spoofing (where a user-supplied local pack named `core` has the same files or is placed at a matching path), the check must assert that the resolved config graph contains a verified, unforgeable import edge stemming from the absolute system-managed path (`.gc/system/packs/core`) and that its resolved layer ID matches the system-materialized pack layer. This check must occur immediately after config composition and before any behaviors (orders, prompts, formulas) are parsed or executed.

### Q3: What degraded or operator-visible path exists when Core integrity fails during live reload, and does it avoid silently continuing with stale behavior?
**Answer:**
If Core fileset or participation validation fails during a reload, `LoadRuntimeCityNoRefresh` aborts and returns a detailed validation diagnostic. The controller continues serving the current in-memory last-known-good (LKG) configuration but transitions its operational state to `read_only_degraded`. In this mode, the API rejects all mutating transactions (e.g., dispatches, formula executions, and session starts) with a clear diagnostic, avoiding silent continuation with corrupted or stale configurations on disk.

---

## Critical Risks & Architectural Inconsistencies (DeepSeek V4 Flash Style)

### 1. [Major] The Provider Pack Bootstrapping Paradox (Circular Dependency)
- **The Risk:** Gate 1 (Pre-Resolution) validates filesystem integrity of required packs—specifically Core plus provider packs (`bd` and `dolt` as selected today)—*before* config resolution (220-222). However, determining which provider is active is a config-layer property declared in `city.toml`. 
- **The Impact:** To know which provider pack is required, the loader must parse `city.toml`. But to safely parse `city.toml`, the loader must first complete Gate 1. This is a classic bootstrapping circular dependency. If the loader reads and parses `city.toml` to extract the provider before validating the file sets, it parses unvalidated behavior-bearing assets, defeating Gate 1's "pre-resolution" integrity guarantee.
- **Resolution:** Eliminate the bootstrapping paradox by requiring Gate 1 to unconditionally validate the integrity of *all* built-in required system packs (`core`, `bd`, and `dolt`) at startup, regardless of the active provider. The overhead of hashing a few extra small templates is negligible, and verifying the entire set eliminates any pre-resolution config dependency.

### 2. [Major] Bypass Scanner AST Vulnerability (Substring Flaw)
- **The Risk:** Line 235 models the loader scanner on `cmd/gc/worker_boundary_import_test.go`, which is a simple substring/regex search on files. 
- **The Impact:** A substring scan keyed on `config.Load` or `config.LoadWithIncludes` is trivial to bypass. Developers can easily write multi-line calls (`config.\nLoad(...)`), alias the package import (`import cfg "github.com/.../internal/config"`), or assign and invoke function values (`fn := config.LoadWithIncludes; fn(...)`). The substring scanner will miss these entirely, silently leaking unvalidated direct loading paths into production while passing CI—directly violating the bypass containment mandate.
- **Resolution:** Explicitly require that `TestGCProductionLoaderBoundary` uses Go's `go/parser` and `go/ast` packages to traverse the AST. The scanner must resolve imports, trace package aliases, and reject any call expression referencing a function in the `internal/config` package, unless registered in the partial-read allowlist.

### 3. [Major] Last-Known-Good (LKG) In-Memory Volatility and Startup Deadlock
- **The Risk:** On reload failure, the controller retains the LKG config *only in-memory* (228) and transitions to read-only diagnostics mode.
- **The Impact:** If the controller process crashes, or if the host machine restarts while the on-disk Core fileset is in a corrupted state, the in-memory LKG is lost. Upon restarting, `LoadRuntimeCity` will fail Gate 1, and since there is no persisted LKG state, the controller will fail to boot entirely. This causes a total API blackout and prevents operators from even querying the "read-only status/reporting" or receiving diagnostics, leading to a severe system deadlock.
- **Resolution:** The plan must require that a validated, serialized copy of the last-known-good config be securely cached on disk (e.g., as `.gc/system/lkg.toml` or a sealed bead in `.gc/beads.db`) to survive controller restarts and allow booting into a degraded diagnostic state under all conditions.

### 4. [Major] Shadowing / Spoofing of Required Pack Participation (Gate 2 Provenance Loophole)
- **The Risk:** Gate 2 asserts that required packs "participated in the resolved config" (223). However, Gas City's progressive override model allows user-supplied config layers or overlays to shadow or override any properties or extension points declared by Core.
- **The Impact:** If a user-supplied pack shadows all Core-defined prompt templates or formula bindings, the Core layer still technically "participates" in the config resolution chain (i.e., its import edge is present), but its actual behavioral contributions are nullified. Furthermore, if the check only matches on the pack name "core", a user pack can easily spoof participation.
- **Resolution:** Specify that the post-resolution participation validator must:
  1. Verify that the resolved layer IDs of `core` and active provider packs match the absolute, system-managed materialization paths under `.gc/system/packs/`.
  2. Implement an override-integrity check ensuring that essential Core-declared system schemas and security hooks cannot be shadowed or overridden by user-supplied config layers without explicit approval.

### 5. [Minor] Ad-Hoc Reload Refusal and Lack of a Uniform Guard Surface
- **The Risk:** The plan lists multiple separate components that must refuse behavior-changing operations during degraded LKG mode: dispatch, formula, order, hook, prompt, and agent-start (230-231).
- **The Impact:** Relying on ad-hoc, prose-based checks scattered across multiple packages risks developer oversight. If a new API route or business-logic path is introduced and forgets to check the reload status, it may execute on corrupted or stale configurations.
- **Resolution:** Define a single, centralized in-memory guard object (e.g., `RuntimeConfigStatus`) returned by `LoadRuntimeCityNoRefresh`. All behavior-changing entry points must query this centralized status before executing. Add an entrypoint-coverage test to guarantee that any new route or execution path is structurally gated.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Attempt 4 Design | Status |
| :--- | :--- | :--- |
| **Core absence is only a doctor warning after behavior already loaded** | **Resolved.** Gate 1 pre-resolution fileset validation blocks loading before any files are resolved or parsed. | **Pass** |
| **Token scanners miss wrapper, alias, function-value, or generated API loader bypasses** | **Vulnerable.** Modeling the scanner on a substring test (`worker_boundary_import_test.go`) makes it highly vulnerable to AST-level bypasses. | **Fail-Closed Risk** |
| **A corrupted or partially materialized system pack is accepted because path and digest checks are conflated** | **Resolved.** Gate 1 asserts cryptographic fileset digests against the embedded manifest before loading. | **Pass** |

---

## Required Changes
1. **Unconditional Gate 1 Validation:** Specify that Gate 1 always materializes and validates the filesets of *all* built-in required system packs (`core`, `bd`, and `dolt`) at startup, avoiding any pre-resolution provider-lookup dependency.
2. **AST-Aware Bypass Scanner:** Require `TestGCProductionLoaderBoundary` to use AST-based analysis (`go/parser` and `go/ast`) to catch import aliases, multiline calls, and method/function-value assignments.
3. **Persisted LKG Config Cache:** Mandate that when a reload fails, the LKG configuration is serialized to a protected disk path (e.g., `.gc/system/lkg.toml`) so the system can boot into degraded diagnostic mode after a hard crash or restart.
4. **Absolute Path Participation Proof:** Specify that Gate 2 participation checks must verify that resolved layer IDs point to absolute, system-managed paths (`.gc/system/packs/`), preventing local user overrides from spoofing system packs.
5. **Centralized Reload Guard:** Establish a centralized `RuntimeConfigStatus` guard object that must be checked by all behavior-changing entry points (dispatch, formulas, orders, prompts, agent-start).

---

## Questions
- How does the pre-resolution loader determine which provider pack is active without reading or parsing `city.toml` first?
- Will the bypass scanner be implemented as a robust AST-based analyzer rather than the substring-matching model cited?
- If the controller crashes or is restarted during an active reload failure, how does it serve read-only status and diagnostics without a persisted LKG config?
