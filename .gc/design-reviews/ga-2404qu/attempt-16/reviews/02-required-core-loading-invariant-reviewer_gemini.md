# Elias Sato — DeepSeek V4 Flash (Required Core Loading Invariant Reviewer, Iteration 16)

**Verdict:** approve-with-risks

**Lane:** Required Core inclusion, config provenance, production loader bypass containment, loud failure on corrupt/partial Core, escape-hatch leakage.

Reviewed against the revised design document in Iteration 16 (`.gc/design-reviews/ga-2404qu/attempt-16/design-before.md` updated 2026-06-09T02:00:53Z) and grounded in the live codebase in `cmd/gc/`, `internal/config/`, and `internal/systempacks`.

---

## Executive Summary

As Elias Sato, the **Required Core Loading Invariant Reviewer**, I am updating my verdict to **Verdict: approve-with-risks** for Iteration 16.

The evolution of the design since early iterations (such as the initial Attempt 1 `block` and Attempt 12 `approve-with-risks`) represents an outstanding, systems-aware maturation. The inclusion of the **Attempt 11 Fatal Gates**, the AST **Loader Scanner**, and the **Attempt 14 Resolver-Produced Provenance** has successfully elevated system pack integrity from an ad-hoc path-checking heuristic to a robust, type-safe compiler and runtime boundary.

However, a rigorous systems-level review of the Iteration 16 draft reveals **critical cross-document contradictions**, **bootstrapping paradoxes**, and **concurrency vulnerabilities** that have survived into the final Proposed Design section. These must be resolved before the implementation can be considered production-safe.

---

## Top Strengths & Design Evolution

- **Double Uniform Fatal Gates (Gates 1 & 2)**: The two-stage gating model—pre-resolution required file-set integrity followed by post-resolution typed participation (`RequiredSystemPackParticipation`)—is exceptionally strong. It permanently closes the vulnerability where Core is present on disk but omitted from the resolved configuration.
- **Deny-by-Default AST Scanner**: Extending static scanner enforcement to production `internal/` packages and explicitly checking against a generated `loader-inventory.generated.yaml` prevents developers from introducing rogue config-loading shortcuts in downstream work.
- **Read-Only Diagnostic Isolation**: The Attempt 14 contract clearly separates report-only diagnostic passes (`LoadRuntimeCityNoRefresh`, `ValidateRequiredFileSetsNoRefresh`) from mutative repairs, preserving the integrity of non-destructive `gc doctor` scans.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Do all production config resolution paths route through the system-pack wrapper so Core is included for real gc commands?
**Yes, structurally.** By enforcing a deny-by-default AST scanner that rejects direct lower-level loaders (such as `config.Load`, `config.LoadWithIncludes`, or `config.LoadWithIncludesOptions`) across all production `cmd/gc` and behavior-driving `internal/` paths, all configuration reads are forced through `internal/systempacks.LoadRuntimeCity` or `LoadRuntimeCityNoRefresh`.

### Q2: What guard test fails if a new cmd/gc path calls config.LoadWithIncludes or another lower-level loader without required Core?
The build fails during CI due to the **AST Scanner Test** (modeled on `TestGCNonTestFilesStayOnWorkerBoundary`). It recursively parses all production Go files in `cmd/` and `internal/` and rejects any unclassified direct calls or value-level aliases to `internal/config` resolving entrypoints that do not match the checked-in `loader-inventory.generated.yaml` allowlist.

### Q3: Does loading fail loudly on missing, corrupt, or partially materialized .gc/system/packs/core instead of relying on doctor to detect absent orders?
**Yes.** Gate 1 enforces a strict required file-set integrity check against embedded manifests and `pack.toml` digests *before* config resolution, ensuring that any corrupt, missing, or partially materialized Core files cause an immediate, loud load failure during the command startup sequence.

---

## Critical Risks & Architectural Inconsistencies

### 1. [Blocker] Stale Silent-Repair Doctor Instruction in Proposed Design
* **The Contradiction**: In the Proposed Design under **`### Core Presence Doctor`** (lines 2766-2767), the text instructs:
  > *"Load resolved config through `internal/systempacks.LoadRuntimeCity` and verify typed `RequiredSystemPackParticipation` includes Core."*
* **The Hazard**: This directly violates the binding Attempt 14 **`#### Read-Only Doctor Diagnostic Boundary`** (lines 2109-2113), which limits plain `gc doctor` to read-only APIs (`LoadRuntimeCityNoRefresh`). Because `LoadRuntimeCity` is a materializing and self-healing loader, calling it in a plain, read-only `gc doctor` run will silently write repairs to disk, masking the corruption and leaving nothing for `gc doctor --fix` to report or coordinate.
* **Required Change**: Align the Proposed Design text with the Read-Only Doctor contract. Explicitly mandate that the doctor check calls `LoadRuntimeCityNoRefresh` and translates its fail-closed errors into actionable diagnostics.

### 2. [Major] The Layered Configuration & Pre-Read Paradox (Circular Dependency)
* **The Paradox**: Selecting provider-dependent required host packs (`bd` or `dolt`) requires knowing the "final effective beads provider" (defined in `city.toml` or includes) before Gate 1 integrity checks run.
* **The Trap**: The design assumes a raw partial-read exception (`peekBeadsProvider` / `configuredBeadsProviderValue`) can safely bypass the loader. However, in a layered configuration (progressive activation levels 0-8), the provider configuration can be defined in a nested include file. 
  - If the partial parser *does not* recursively resolve includes, it will miss the provider configuration, select the wrong required pack (or default to memory/json), and fail Gate 2 post-resolution.
  - If the partial parser *does* recursively resolve includes, it is performing a full, unsafe config resolution path *before* required Core has been validated by Gate 1, completely breaking our containment guarantees.
* **Required Change**: To break this circularity, **Gate 1 should validate the integrity of the entire embedded set of built-in host packs (`core`, `bd`, and `dolt`) on every run**. The validation overhead is negligible, and always validating the full embedded set ensures that whatever provider is eventually resolved, its required pack is already proven to be present and uncorrupted.

### 3. [Major] Unsynchronized Runtime Self-Healing Race Conditions
* **The Risk**: While the doctor mutation coordinator has an explicit advisory city-directory lock to prevent write collisions, the runtime self-heal loader does not.
* **The Hazard**: Normal runtime loads are instructed to "repair before publish, then fail closed if still invalid" (line 2087) on missing or corrupt files. If multiple ordinary `gc` commands run concurrently (e.g., concurrent hooks, background status runs, or parallel tests) on a city with a missing/damaged Core, they will simultaneously attempt to write, rename, prune, or quarantine files in `.gc/system/packs/core/` without synchronization. This will trigger spurious write collisions, partial reads, and catastrophic runtime loader failures.
* **Required Change**: Specify a lightweight, fast, non-blocking locking protocol (with a short, deterministic timeout like 200ms) for any self-healing or materialization step executed during a normal runtime config load.

### 4. [Systems Blocker] Read-Only Sandboxes and Immutable Filesystems
* **The Risk**: The design mandates that any required-pack materialization failure must fail the load closed.
* **The Hazard**: In modern cloud deployments, container sandboxes, or secure Kubernetes pods, filesystems are frequently mounted read-only to prevent tampering. If a container is deployed in a read-only environment and its on-disk `.gc/` directory is missing or damaged, `gascity` will permanently fail closed because it is physically blocked from writing repairs to disk—even though the complete, uncorrupted Core pack is fully embedded in the running `gc` binary.
* **Required Change**: The design must define an in-memory/VFS virtual fallback so that config resolution can resolve required system packs directly from the embedded bytes in the running binary when disk writes are blocked or the environment is classified as read-only.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Iteration 16 Design | Status |
| :--- | :--- | :--- |
| **Core is materialized on disk but absent from resolved config** | **Airtight.** Eliminated by Gate 2 (Post-Resolution typed participation proof), which validates deterministic layer and import-edge ids in the resolver output. | **Pass** |
| **Test-only no-Core escape hatch leaks into production code** | **Airtight.** `GC_BOOTSTRAP=skip` is retired as a production switch. All test-only no-Core behavior is strictly isolated to lower-level loaders in `_test.go` files. | **Pass** |
| **Core absence is detectable only by doctor after runtime behavior is already degraded** | **Airtight.** Gate 1 (Pre-Resolution required file-set integrity) runs on every runtime loader call, failing startup loudly at the boundary. | **Pass** |

---

## Actionable Requirements & Proposed Adjustments

1. **Fix Proposed Design Line 2766**: Replace `internal/systempacks.LoadRuntimeCity` with `internal/systempacks.LoadRuntimeCityNoRefresh` in the doctor check definition to preserve the read-only diagnostic boundary.
2. **Unify Pre-Resolution Integrity Checks**: Update Gate 1 to validate the integrity of the full closed-set of embedded builtin host packs (`core`, `bd`, `dolt`) pre-resolution. This solves the layered provider bootstrap paradox by removing the need for a complex, unsafe, recursive pre-parser.
3. **Introduce Runtime Self-Heal Lock**: Add a fast, non-blocking, short-timeout (e.g., 200ms) lock on the city directory when executing materialization/repair writes during a runtime config load.
4. **Specify a Read-Only/VFS Fallback**: Allow the loader to bypass disk materialization and resolve required packs directly from embedded memory/VFS when the filesystem is read-only.
