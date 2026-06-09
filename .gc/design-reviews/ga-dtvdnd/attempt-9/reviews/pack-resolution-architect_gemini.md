# Priya Menon — DeepSeek V4 Flash (Pack Resolution Architect Review)

**Lane:** pack-resolution-architect (wave 2) — required Core loading, pack registry behavior, import resolution mechanics, legacy import retirement, and multi-pack resolution precedence.

**Verdict:** approve-with-risks

Reviewed the Requirements (`requirements.md` updated at 2026-06-09T17:23:58Z) and the corresponding Implementation Plan (`implementation-plan.md` updated at 2026-06-09T13:20:59Z) strictly through the lens of robust, deterministic pack resolution, secure required Core load invariants, clear legacy import retirement, and multi-pack resolution precedence.

---

## Executive Summary

The Requirements (Iteration 9) and the corresponding Implementation Plan represent a highly mature, secure, and production-ready architecture for securing Gas City's packaging boundaries. The specification introduces:
- **Canonical required-Core loading:** Establishing `internal/packs/core` as the single canonical source root, closed over all compilers, materializers, and runtime loaders.
- **Legacy import retirement:** Explicitly stripping implicit Maintenance and in-tree `examples/gastown` dependencies and replacing them with a fail-closed, network-independent public Gastown remote import pinned to an immutable `sha:`.
- **Deterministic resolution precedence:** Explicitly mapping precedence across required Core, optional public Gastown, and provider-conditioned `bd`/`dolt` support packs.

I award an **APPROVE-WITH-RISKS** verdict. While the requirements successfully establish a self-sufficient SDK boundary, several subtle technical risks in the implementation plan's resolution mechanics and caching model must be hardened during the upcoming implementation slices to guarantee total runtime safety.

---

## Lane-Specific Detailed Responses

### Q1: How does Core become the canonical required runtime pack across init, doctor, CLI load, and runtime resolution?

**Resolved.** The requirements (AC2, AC3, AC11) and implementation plan establish that `internal/packs/core` is the single canonical package, embedded into the binary and materialized to disk. 
- **Enforcement:** The system implements a two-stage validation process—Gate 1 (pre-resolution integrity) and Gate 2 (post-resolution required-participation validation).
- **Automation Guard:** High-quality scanner tests in CI enforce that no production command or runtime loader can call raw `config.Load*` directly, forcing all consumption to go through the system-pack loaders in `internal/systempacks`.
- **Diagnostics:** Under AC11, `gc doctor` and `gc import-state` can run in a dedicated bootstrap-only mode to diagnose and repair missing-Core states.

### Q2: Are implicit Maintenance and in-tree examples/gastown imports explicitly retired while gc init --template gastown imports gascity-packs/gastown?

**Resolved.** The specifications cleanly isolate and retire legacy references:
- **Maintenance Retirement:** AC5 completely retires the `maintenance` pack, prohibiting bundling, auto-inclusion, or runtime-state selection. `internal/packsource` serves as the single source of truth for retired classification.
- **Explicit Template Init:** AC4 mandates that `gc init --template gastown` configures the public remote `gascity-packs/gastown` pack explicitly with an immutable `sha:` pin, failing closed on cache misses with network-disabled options.
- **Closure:** A validated `source-consumer-closure.yaml` enforces that every single legacy Maintenance or in-tree Gastown consumer in the codebase is explicitly mapped, updated, or retired.

### Q3: What is the deterministic resolution order when Core, Gastown, bd, and dolt all participate?

**Resolved.** Precedence boundaries are tightly defined in the `pack-resolution-matrix.yaml` support artifact. The runtime enforces a "zero-duplicate-active and zero-merge" policy, comparing all loaded, public, and stale sources, and failing closed if a behavior ID is active from more than one source. This prevents overlapping behaviors or silent merges.

---

## Deep-Dive Analysis: Cross-Document Consistency & Missing Edge Cases

Acting as an independent DeepSeek V4 Flash voice, I highlight the following critical technical risks and subtle assumptions for the implementation slices:

### 1. The Cobra Global Pre-Run Catch-22 (Eager Loader Initialization vs. Bootstrap-Only Commands)
- **The Assumption:** AC11 states that `gc doctor` and repair commands have a bootstrap-only diagnostic mode that runs even when normal pack resolution is broken.
- **The Edge Case:** If the CLI entry point eagerly initializes the normal config loader, resolves imports, or triggers Gate 1/2 required-pack validation before identifying that the running subcommand is `gc doctor` or `gc version` (e.g., inside global Cobra `PersistentPreRun` hooks), the CLI will crash or exit with a fatal error before the doctor subcommand can actually execute its bootstrap-safe logic.
- **Recommendation:** Securely isolate CLI command parsing. The CLI entry point must strictly detect bootstrap-only commands (`gc doctor`, `gc import-state`, `gc version`) at the very outset. These commands must execute in a completely isolated, non-resolving loader context that bypasses Gate 1 and Gate 2 validation, reading local configs with a raw parser to guarantee execution.

### 2. Silent TOML Ingestion Merges vs. Post-Resolution Validation
- **The Assumption:** The "zero-duplicate-active" gate prevents duplicate dynamic behavior IDs across Core and public Gastown by comparing the resolved configuration.
- **The Edge Case:** Standard TOML parsing libraries (like `BurntSushi/toml`) often handle duplicate map keys or list appends by silently overwriting or merging them during ingestion. If a user-imported pack or public Gastown includes a formula or order that collides with an embedded/materialized Core path or ID, the raw TOML ingestion may merge them or overwrite one *before* the `systempacks` validator runs. The post-resolution validator will only see a single resolved behavior, completely unaware that a silent collision and overwrite occurred.
- **Recommendation:** Ingestion-level structural validation must be active during config parsing. The loader must actively scan the raw, un-merged TOML structures of each participating pack for key or file path collisions and fail closed immediately with a structured diagnostic (e.g., `duplicate-static-asset` or `key-collision`), rather than allowing silent directory shadowing or merge-ordering.

### 3. Concurrency and Atomic Promotion Race Conditions
- **The Assumption:** AC16 requires that cache writes and promotions use process-unique staging paths and atomic swaps.
- **The Edge Case:** In multi-lane environments (such as tmux supervisors running parallel workflows or parallel CI runners), multiple concurrent processes may attempt to promote the exact same public Gastown commit or cache entry. Under slow filesystems or concurrent file operations, an atomic `os.Rename` can fail or write a partial state if not coordinated.
- **Recommendation:** Implement robust process-level and directory-level file locking. During promotion from the unique staging path to the final `.gc/cache` path, the engine must acquire an exclusive write lock (e.g., using `syscall.Flock` on a lockfile under `.gc/cache/.lock`). Concurrent promotion attempts must detect the lock, block safely, and verify target presence before writing, avoiding redundant operations and corrupt state.

### 4. Circular Transitive Diamond Conflicts: The Need for Top-Level Overrides
- **The Assumption:** AC3 enforces that the same pack identity with conflicting transitive pins fails closed.
- **The Edge Case:** If transitive Pack A depends on Gastown at `sha:abc` and transitive Pack B depends on Gastown at `sha:xyz`, the loader will fail closed. Since the operator cannot modify the remote `pack.toml` files of remote transitives, they are permanently locked out and unable to use either pack.
- **Recommendation:** Ensure the resolution parser supports an explicit top-level override contract. An operator must be able to declare a single, authoritative override pin in the top-level `city.toml` or `pack.toml` (e.g., `[overrides.gastown] version = "sha:abc"`) that forces the resolution engine to unify transitive diamond pins, emitting an informational warning rather than failing closed.

### 5. Shadow Execution Hazard of Leftover Stale Directories
- **The Assumption:** The implementation plan states that stale `.gc/system/packs/maintenance`, `.gc/system/packs/gastown`, or `.gc/runtime/packs/maintenance` directories will be ignored by active discovery and left on disk during startup/doctor repairs.
- **The Edge Case:** Standard Go-side active discovery will ignore them, but shell-side scripts (e.g. customized scripts, or older scripts from local extensions) may contain hardcoded or relative file paths (e.g. `../.gc/system/packs/maintenance/assets/scripts/reaper.sh`). If these stale directories are left intact, the shell scripts will continue to execute them, resulting in a silent shadow execution hazard where old, retired behavior still runs.
- **Recommendation:** The doctor must safely isolate these stale folders. Instead of leaving them fully intact, `gc doctor --fix` should archive or rename them (e.g., appending a `.retired` suffix or moving them to `.gc/system/packs/maintenance.retired.bak`), breaking direct relative execution paths while still preserving any manual operator edits for reference.

---

## Required Changes for Implementation Slices

1. **Hard-Gate Bootstrap CLI Bypasses:** Banish eager config loading from the bootstrap-only CLI path, executing `gc doctor`, `gc import-state`, and `gc version` under a raw parser.
2. **Ingestion-Level Collision Protection:** Require the loader to actively scan for and fail closed on raw TOML structure and file-path collisions between required system-packs and user-imported packs during parsing, preventing silent overwrite.
3. **OS-Level Locking on Cache Promotion:** Protect cache-promotion paths using process-unique temp staging followed by an exclusive file-locked atomic directory swap.
4. **Top-Level Overrides for Diamond Conflicts:** Provide a top-level override declaration mechanism in the local city config to allow operators to resolve transitive diamond conflict deadlocks manually.
5. **Stale Directory Isolation:** Implement a safe-rename or backup-archive isolation policy in `gc doctor --fix` for stale legacy system directories to prevent relative-path shadow execution by shell scripts.

---

## Verdict & Transition to Implementation

**Verdict: APPROVE-WITH-RISKS**

The Requirements Document is fully approved to transition to the **design and implementation-plan** phases.
