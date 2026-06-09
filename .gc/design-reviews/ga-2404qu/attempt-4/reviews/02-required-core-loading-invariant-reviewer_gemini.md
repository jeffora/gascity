# Elias Sato — DeepSeek V4 Flash (Required Core Loading Invariant Review, Iteration 4, Independent)

**Verdict:** approve

**Lane:** Required Core inclusion, config provenance, production loader-bypass containment, loud failure on corrupt/partial Core, escape-hatch leakage.

Reviewed against the iteration 4 design-before document (updated 2026-06-07T00:30, containing the newly specified `Required Core Identity And Loader Contract`, `Doctor And Import-State Safety`, and the multi-slice rollout staging) and grounded in the `cmd/gc/`, `internal/config/`, and `internal/builtinpacks/` packages.

---

## Executive Summary

The iteration 4 design represents a complete, mathematically sound resolution of all core loading invariants. By moving from incidental path-based provenance matches to a typed, content-backed, and validation-linked contract (`RequiredSystemPackParticipation`), the design closes the silent-deduplication, shadow-collision, and zero-key-contribution vulnerabilities that plagued earlier iterations. 

All three previous blocker conditions are now fully addressed in the design:
1. **Deduplication and Collision Safety**: Handled through a hard fail-closed gate when a user or imported pack named `core` collides with required system Core, preventing silent replacement of system assets.
2. **Loud Failure Separation**: Split correctly into a manifest-integrity check (detecting corruption/partial materialization via `unusableRequiredBuiltinPackNames`) and a post-load config resolution check (detecting loader bypasses or dropped includes).
3. **Controller Reload Safety**: Explicitly routes `tryReloadConfig` through config-returning wrappers that enforce both integrity and provenance presence.

I approve this design without further blocking changes. The following review outlines the strengths of the current design, addresses the core questions in my lane, and highlights minor recommendations for the implementation team to ensure high-fidelity execution.

---

## Top Strengths of the Iteration 4 Design

1. **Sturdy Identity and Participation Model (`RequiredSystemPackParticipation`)**
   Defining a typed contract containing the pack ID, materialized directory path, embedded source ID/digest, validated file-set/pack.toml digests, and resolved config layer ID is the ultimate defense. Checking for an explicit import edge in the resolved config layers solves the edge case where Core remains required but contributes zero effective keys (due to user overrides).

2. **Scanner-Enforced Call-Site Containment**
   The commitment to implement a static analysis scanner (modeled after `TestGCNonTestFilesStayOnWorkerBoundary`) that rejects direct production calls to `config.Load*` is the only reliable way to prevent bypass leakage. This ensures no new subcommands or features can bypass Core loading as the codebase grows.

3. **Atomic and Comment-Preserving Doctor Fixes**
   The doctor safety contract's preflight requirements (validating Core materialization, parseable lockfiles, and reachable Gastown remotes) and failure-atomicity staging (leaving the city byte-identical on failure) ensure that fixing a missing Core doesn't cause silent configuration corruption.

4. **Staged Rollout Sequencing**
   Separating the rollout into independent, test-gated slices—culminating in the `TestPinnedPublicGastownBehavior` packcompat gate before source deletion—minimizes version-skew risks and makes rollbacks deterministic.

---

## Lane Question Analysis

### 1. Do all production config resolution paths route through the system-pack wrapper so Core is included for real gc commands?
**Yes.** Under the proposed design, all normal command paths must load through wrappers like `loadCityConfigWithBuiltinPacks`. The scanner check enforces that raw `config.Load*` calls are restricted to non-production code (tests) or narrowly scoped, documented, and tested partial-read exceptions in an allowlist. 

### 2. What guard test fails if a new cmd/gc path calls config.LoadWithIncludes or another lower-level loader without required Core?
The **static scanner test** (modeled on the worker boundary import test) will fail at build/CI time if a non-test file introduces an unauthorized raw load call. For authorized allowlisted partial-reads, focused unit tests must prove that the call cannot influence behavior-driving executions (e.g., dispatching agents or schedules) in a Core-less city. Furthermore, any wrapper-routed runtime load that fails to show Core in its resolved provenance will trigger a closed load failure via `assertRequiredSystemPackProvenance`.

### 3. Does loading fail loudly on missing, corrupt, or partially materialized .gc/system/packs/core instead of relying on doctor to detect absent orders?
**Yes, in two complementary stages.** First, the manifest-integrity check (`unusableRequiredBuiltinPackNames`) runs after materialization and fails loudly if files are corrupt or missing from disk. Second, the post-load provenance check ensures that even if files exist on disk, the loader didn't silently drop the Core pack during composition. These are both fatal on the production config-load path, ensuring that a degraded Core never results in silent, undefined runtime behavior.

---

## Implementation Guidance & Edge Cases

To ensure the implementation matches the high standards of this design, the following minor recommendations should be incorporated during the coding slices:

### 1. Path Normalization for Provenance Comparison
When comparing the materialized Core path against `prov.Sources`, the assertion helper must use the project's canonical `normalizePathForCompare` (defined in `embed_builtin_packs.go`) to prevent absolute-vs-relative path discrepancies or trailing-slash mismatches from causing false-positive load failures.

### 2. Unexpected Extra Files under `.gc/system/packs/core`
The design states: *"Required system-pack integrity must either validate the full file set or prove unexpected files cannot influence loaded formulas, orders, scripts, overlays, prompts, or config."* 
* **Recommendation:** The implementation should enforce a strict **allowlist-only** validation against the embedded manifest. If unexpected files are detected inside `.gc/system/packs/core`, they should be deleted or treated as corrupt. This prevents operators or third-party tools from injecting unauthorized scripts or overlays into the required system pack.

### 3. Non-OSFS Test Escape Hatch
The `usesOSFS` helper currently skips system-pack inclusion for memory/mock filesystems in tests.
* **Recommendation:** The static scanner must assert that no production `cmd/gc` code invokes `loadCityConfig*` with a non-OSFS filesystem wrapper. Non-OSFS loaders must remain strictly confined to test files.

### 4. Doctor Wording Alignment
The requirements document states the doctor check will *"offer a fix that adds the Core pack entry."* The design correctly points out that this actually *"repairs the generated system pack and normal include path."* 
* **Recommendation:** The doctor CLI output must not suggest to the operator that it is writing a user-facing `[imports.core]` to `city.toml`. The wording should clearly state that it is repairing the canonical system-pack directory and verification state.

---

## Verdict: Approved

The iteration 4 design is cohesive, technically rigorous, and directly addresses the entire risk matrix for the Required Core Loading Invariant. The transition from loose path assertions to a typed content-backed contract resolves my previous blocking concerns. I look forward to reviewing the implementation slices as they land.
