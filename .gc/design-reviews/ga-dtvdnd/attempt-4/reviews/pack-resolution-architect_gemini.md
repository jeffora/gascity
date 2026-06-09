# Priya Menon — DeepSeek V4 Flash (Pack Resolution Architect Review)

**Verdict:** approve-with-risks

**Scope:** Required Core loading, pack registry behavior, import resolution mechanics, legacy import retirement, and multi-pack resolution precedence.

> ### Lane Note (Verify-Don't-Copy + Dual Placement)
> 1. **Re-grounding & Independence:** This review is an independent DeepSeek V4 Flash evaluation. While I have reviewed the prior Attempt 1 files and Claude's Attempt 3 draft, my findings are fresh, technically grounded in the live codebase, and focused on exposing unvoiced assumptions and systemic resolution paradoxes that other lanes/reviewers may accept too quickly.
> 2. **Dual-Placement Strategy:** Due to a known workflow defect documented in `attempt-2/synthesis.md` (where `gc.attempt=1` on beads causes them to write to `attempt-1/reviews/` and block attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/pack-resolution-architect_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-4/reviews/pack-resolution-architect_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 4 synthesis.

## Executive Summary
The Attempt 3 requirements document represents a major step forward in separating the SDK Core from user-supplied orchestration. By establishing a required-Core contract and retiring standalone Maintenance, it moves Gas City toward a pure role-neutral model.
However, from the perspective of a **Pack Resolution Architect**, several deep-seated architectural paradoxes and edge cases remain unaddressed. Specifically, we identify a **Bootstrapping Repair Paradox** where diagnostic repair commands are required to fix config errors but may themselves depend on config resolution, and a **Transitive Legacy Shadowing** issue where third-party packs can re-introduce retired dependencies. 
To prevent downstream implementation drift, I award this document an **APPROVE-WITH-RISKS** verdict and mandate four critical architectural pins.

## Top Strengths
* **Diagnostic Loader Path (AC11/Negative Path 3):** Elevating the non-fatal boot mode for `gc doctor` and `gc import-state` to a first-class requirement successfully breaks the startup dependency loop on missing Core.
* **First-Class Deterministic Precedence (AC3):** The explicit enumeration of all resolution participants (Core, `bd`/`dolt`, locked/cached remotes, local overlays) creates a testable matrix contract.
* **Strict Fallback Prevention (AC4/AC5):** Prohibiting implicit materialization of in-tree legacy paths (`examples/gastown/packs/gastown`, `.gc/system/packs/gastown`) ensures parity between test, dev, and production.

## Critical Risks & Assumptions

### 1. The Bootstrapping Repair Paradox (Resolving the Unresolved)
* **The Assumption:** W6H and AC10 demand that diagnostics provide "explicit idempotent repair actions," but Open Question 5 acknowledges that the actual repair command is undefined. Other reviewers treat this as a minor gap.
* **The Risk:** In a missing-Core or corrupt-import state, the `gc` CLI must boot to run the repair command. If the repair command relies on the standard resolution engine, and that engine crashes or fails to initialize because Core is missing, the operator can never run the repair command. This is a circular deadlock.
* **The Pin:** The resolution engine must support a minimal **"zero-config" or "bootstrap-only" loading mode**. This mode must bypass full import graph resolution and allow the CLI to execute basic diagnostics and mutations (e.g. `gc repair` or `gc doctor --repair`) without requiring a resolved Core pack.

### 2. Transitive Legacy Shadowing & Nested Skew
* **The Assumption:** AC3 prevents a city's direct imports from shadowing required pack identities, assuming direct import validation is sufficient.
* **The Risk:** A user city might import a well-formed third-party pack. However, that third-party pack may internally contain transitive nested imports pointing to legacy/retired paths (`packs/maintenance` or `.gc/system/packs/maintenance`). If the resolver only validates top-level imports, legacy paths can be silently re-introduced into the active graph, leading to non-deterministic behavior.
* **The Pin:** Transitive validation is mandatory. The deterministic resolution engine must recursively parse the entire import graph, and `gc doctor` must explicitly trace and output the **complete nested import chain** (e.g., `City -> Pack A -> legacy/packs/maintenance`) when attributing retired import diagnostics.

### 3. The Offline Cache-Keying & Determinism Gap
* **The Assumption:** Edge Case 7 assumes offline resolution can succeed deterministically if the pinned pack is in the repo cache.
* **The Risk:** If the public pack import uses mutable git references (like branch names `main` or tags `v1`), determining if the cached copy is up-to-date requires a network handshake. Offline, the resolver cannot verify if the cache is fresh. If it silently uses a stale cache, it breaks determinism; if it fails, it breaks offline capability.
* **The Pin:** Offline resolution must only be permitted for **fully-pinned, immutable version specifications (exact git commits / SHAs)**. If an import uses a mutable reference, offline resolution must fail-closed with a clear diagnostic indicating that mutable imports require an active connection.

### 4. Bounding the Dev/Test Escape Hatch (AC2 Security)
* **The Assumption:** AC2 allows a "clear dev/test escape hatch if tests need to construct partial configs."
* **The Risk:** If this escape hatch is exposed via standard environment variables or runtime flags, production configurations could accidentally (or maliciously) trigger it to bypass required Core checks. This would create unverified, silent, non-standard runtime environments in real cities.
* **The Pin:** The AC2 escape hatch must be **mechanically restricted to compilation or test-harness level** (e.g. Go build tags `//go:build test` or internal test-only function signatures) and must never be exposed as an undocumented CLI flag or standard environment variable in production binaries.

## Missing Evidence
* **Nested Import Diagnostics:** No golden-test output or schema exists showing how a nested transitive legacy import path is traced back to its root config layer.
* **Resolver Precedence for Split Assets:** There is no explicit rule resolving collisions of same-named split assets (e.g., custom scripts or template fragments) between required Core and explicitly imported external Gastown.
* **Fresh Offline Initialization Contract:** There is no stated policy for a fresh offline `gc init --template gastown` with no pre-existing cache (it must fail-closed to preserve role-neutrality).

## Required Changes
1. **Resolve the Bootstrapping Paradox:** Add a clause to AC11 specifying that `gc doctor` and any future repair commands can execute in a bootstrap-only mode that does not require a fully resolved pack import graph.
2. **Mandate Transitive Diagnostics:** Amend AC11 to require that missing-Core and retired-import diagnostics recursively trace and display the complete import chain for nested transitive imports.
3. **Pin Immutable Offline Resolution:** Specify under Edge Case 7 that offline resolution of external packs is strictly conditional on immutable commit/SHA locks.
4. **Isolate the Escape Hatch:** Add a constraint to AC2 stating that the dev/test escape hatch must be non-addressable via production CLI flags or production environment variables.
5. **Asset-Level Precedence Pin:** Add an explicit clause in AC3/AC8 declaring that in a city importing both required Core and external Gastown, the explicitly imported pack (Gastown) has precedence over same-named assets in Core, avoiding non-deterministic shadowing.

## Questions
* **How are transitive dependencies with conflicting version locks handled?** If two independent imports depend on different pinned SHAs of the same external pack, does the resolver fail, or does it attempt to unify them?
* **Will the `gc repair` command be built into the standard `gc` binary, or will it be a script packaged with required Core?** (Drives the bootstrap-only loading requirements).
