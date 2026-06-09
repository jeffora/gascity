# Elias Sato — DeepSeek V4 Flash (Required Core Loading Invariant Reviewer, Iteration 19)

**Verdict:** block

**Lane:** Required Core inclusion, config provenance, production loader bypass containment, loud failure on corrupt/partial Core, escape-hatch leakage.

Reviewed against the revised design document in Iteration 19 (`.gc/design-reviews/ga-2404qu/attempt-19/design-before.md` updated 2026-06-09T10:44:00Z) and grounded in the live codebase in `cmd/gc/`, `internal/config/`, and `internal/systempacks`.

---

## Executive Summary

As Elias Sato, the **Required Core Loading Invariant Reviewer**, I am issuing a **Verdict: block** for Iteration 19.

While the added **Attempt 17 Review Resolution Contracts** (specifically the *Default-Deny Production Loader Inventory* and the *Generated Artifact Contracts*) represent an extraordinary leap in formalizing the system-pack validation boundary, a rigorous, line-by-line systems-level comparison between these new contracts and the legacy Proposed Design sections reveals **critical cross-document contradictions**, **Circular Dependency paradoxes**, and **concurrency vulnerabilities**.

Because these gaps allow silent-repair side effects in report-only paths, leave concurrent CLI commands vulnerable to race-driven file corruption, and permanently fail-closed in secure read-only container sandboxes, the design is not yet ready for implementation. We must resolve these internal conflicts to make the loading invariant truly bulletproof.

---

## Top Strengths

- **Airtight Loader Inventory Contract**: The introduction of `loader-inventory.generated.yaml` (lines 2408–2418) as the sole, programmatic default-deny authority over both `cmd/` and `internal/` code blocks is a premier design. It prevents downstream developers from introducing unvalidated, low-level configuration loaders.
- **Resolver-Produced Participation Proof**: Mandating that required system packs prove participation via resolver-emitted `RequiredSystemPackParticipation` records with content-digest-derived IDs (lines 2426–2430) permanently closes the "materialized on disk but bypassed in resolved config" vulnerability.
- **Enforced Read-Only Diagnostic Boundaries**: The Attempt 17 safety contract clearly separates report-only diagnostic passes (`LoadRuntimeCityNoRefresh`) from the mutative, staged writes managed by the `doctor.MutationCoordinator` (lines 2550–2562).

---

## Critical Risks & Architectural Inconsistencies

### 1. [Blocker] Stale Silent-Repair Doctor Instruction in Proposed Design
* **The Contradiction**: Under **`### Core Presence Doctor`** (line 3061), the text instructs:
  > *"Load resolved config through `internal/systempacks.LoadRuntimeCity` and verify typed `RequiredSystemPackParticipation` includes Core."*
* **The Hazard**: This directly violates the binding Attempt 17 **`#### Doctor Mutation And Runtime-State Safety`** contract (lines 2550–2554), which explicitly mandates that plain `gc doctor` is report-only and *"must not materialize, repair, promote cache entries ... or rewrite TOML."* Because `LoadRuntimeCity` is an active, materializing, and self-healing loader, calling it in a plain, read-only `gc doctor` run will silently write repairs to disk, masking the corruption and leaving nothing for `gc doctor --fix` to report or coordinate.
* **Required Change**: Replace `internal/systempacks.LoadRuntimeCity` on line 3061 with `internal/systempacks.LoadRuntimeCityNoRefresh`, converting the validation failure directly into a structured diagnostic error.

### 2. [Blocker] Stale Symbol List and Missing Loader in Proposed Design
* **The Contradiction**: Under **`### System Pack Loading`** (line 3027), the design states:
  > *"Direct `config.Load`, `config.LoadCity`, `config.LoadWithIncludes` ... in production files are rejected by scanner tests."*
* **The Hazard**: This symbol list directly contradicts the Attempt 17 contract on line 2415, which correctly identifies `LoadWithIncludesOptions` and raw TOML pre-reads. It still lists `config.LoadCity` (which is a phantom function that does not exist in `internal/config`) and completely omits `config.LoadWithIncludesOptions` (a real, exported resolver in `internal/config/compose.go`). Because the scanner relies on this list, omitting `LoadWithIncludesOptions` from the Proposed Design section leaves a massive backdoor for unvalidated production reads.
* **Required Change**: Align line 3027 with line 2415. Delete the phantom `config.LoadCity` and explicitly include `config.LoadWithIncludesOptions`.

### 3. [Major] The Beads Provider Bootstrap Paradox
* **The Paradox**: Selecting provider-dependent host packs (`bd` or `dolt`) requires knowing the "final effective beads provider" (defined in `city.toml` or includes) before Gate 1 integrity checks run.
* **The Trap**: On line 3000, the Proposed Design section still assumes:
  > *"RequiredPackNames(cityPath) starts with Core and adds provider-dependent bd and dolt as today."*
  This relies on the legacy pre-read helper (`peekBeadsProvider`), which bypasses Gate 1. In a layered config, if this partial read does not resolve includes, it selects the wrong required pack (failing Gate 2); if it does resolve includes, it executes an unvalidated config load path before Core has been verified, breaking containment.
* **Required Change**: To break this circularity, Gate 1 must validate the integrity of the *entire* closed-set of embedded builtin host packs (`core`, `bd`, and `dolt`) pre-resolution. Always validating the full embedded set eliminates the need for an unsafe pre-load parser.
* **Gap in Core Presence Doctor**: Line 3058 instructs the doctor to verify only the Core file set, neglecting to verify the full set of embedded host packs (`bd` and `dolt`) when they participate as required host packs, creating a gap in diagnostic completeness.

### 4. [Major] Unsynchronized Runtime Self-Healing Race Conditions
* **The Risk**: The design instructs the ordinary runtime loader `LoadRuntimeCity` to execute materialization and repairs on missing/corrupt required-pack files.
* **The Hazard**: While the `doctor.MutationCoordinator` has an advisory lock (line 3108), the ordinary runtime loader does not. If multiple `gc` commands run concurrently (e.g., parallel git hooks, background status runs, or concurrent API requests) on a city with a damaged or missing Core, they will simultaneously attempt to write, rename, prune, or quarantine files in `.gc/system/packs/core` without synchronization, triggering catastrophic write collisions and partial-write corruption.
* **Required Change**: Specify a lightweight, fast, non-blocking locking protocol (with a short timeout like 200ms) for any self-healing or materialization step executed during `LoadRuntimeCity` to prevent concurrent write collisions.

### 5. [Major] Read-Only Sandboxes and Immutable Filesystems
* **The Risk**: The design mandates that any required-pack materialization failure must fail the load closed.
* **The Hazard**: In secure cloud environments (e.g., read-only Kubernetes pods), filesystems are mounted read-only to prevent tampering. If a container starts up and its on-disk `.gc/system/packs/core` is missing or damaged, `gascity` will permanently fail closed because it is physically blocked from writing repairs to disk—even though the complete, uncorrupted Core pack is fully embedded in the running `gc` binary.
* **Required Change**: Specify an in-memory/VFS virtual fallback so that the loader can resolve required system packs directly from the embedded bytes in the running binary when disk writes are blocked or the environment is read-only.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Do all production config resolution paths route through the system-pack wrapper so Core is included for real gc commands?
**Yes, structurally.** By enforcing a deny-by-default AST scanner that rejects direct lower-level loaders (such as `config.Load`, `config.LoadWithIncludes`, or `config.LoadWithIncludesOptions`) across all production `cmd/gc` and behavior-driving `internal/` paths, all configuration reads are forced through `internal/systempacks.LoadRuntimeCity` or `LoadRuntimeCityNoRefresh`.

### Q2: What guard test fails if a new cmd/gc path calls config.LoadWithIncludes or another lower-level loader without required Core?
The build fails during CI due to the **AST Scanner Test** (modeled on `TestGCNonTestFilesStayOnWorkerBoundary`). It recursively parses all production Go files in `cmd/` and `internal/` and rejects any unclassified direct calls or value-level aliases to `internal/config` resolving entrypoints that do not match the checked-in `loader-inventory.generated.yaml` allowlist.

### Q3: Does loading fail loudly on missing, corrupt, or partially materialized .gc/system/packs/core instead of relying on doctor to detect absent orders?
**Yes.** Gate 1 enforces a strict required file-set integrity check against embedded manifests and `pack.toml` digests *before* config resolution, ensuring that any corrupt, missing, or partially materialized Core files cause an immediate, loud load failure during the command startup sequence.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Iteration 19 Design | Status |
| :--- | :--- | :--- |
| **Core is materialized on disk but absent from resolved config** | **Airtight.** Eliminated by Gate 2 (Post-Resolution typed participation proof), which validates deterministic layer and import-edge ids in the resolver output. | **Pass** |
| **Test-only no-Core escape hatch leaks into production code** | **Airtight.** `GC_BOOTSTRAP=skip` is retired as a production switch. All test-only no-Core behavior is strictly isolated to lower-level loaders in `_test.go` files. | **Pass** |
| **Core absence is detectable only by doctor after runtime behavior is already degraded** | **Airtight.** Gate 1 (Pre-Resolution required file-set integrity) runs on every runtime loader call, failing startup loudly at the boundary. | **Pass** |

---

## Missing Evidence

* **Specific Lock Implementation details**: No locking mechanism or timeout (e.g., 200ms non-blocking advisory file lock) is detailed for concurrent runtime self-healing.
* **VFS Fallback details**: The design does not specify how the system behaves when attempting to repair a missing pack on a read-only filesystem.

---

## Required Changes

1. **Fix Proposed Design Line 3061**: Replace `internal/systempacks.LoadRuntimeCity` with `internal/systempacks.LoadRuntimeCityNoRefresh` in the doctor check definition to preserve the read-only diagnostic boundary.
2. **Fix Proposed Design Line 3027**: Remove `config.LoadCity` and explicitly include `config.LoadWithIncludesOptions` and package aliases in the scanner rejection list.
3. **Unify Pre-Resolution Integrity Checks**: Update Gate 1 to validate the integrity of the full closed-set of embedded builtin host packs (`core`, `bd`, `dolt`) pre-resolution. This solves the layered provider bootstrap paradox by removing the need for a complex, unsafe, recursive pre-parser.
4. **Introduce Runtime Self-Heal Lock**: Add a fast, non-blocking, short-timeout (e.g., 200ms) lock on the city directory when executing materialization/repair writes during `LoadRuntimeCity` to prevent concurrent write collisions.
5. **Specify a Read-Only/VFS Fallback**: Allow the loader to bypass disk materialization and resolve required packs directly from embedded memory/VFS when the filesystem is read-only.
6. **Extend Doctor Verification**: Ensure that the Core Presence Doctor check (line 3058) also validates `bd` and `dolt` required host packs when they are configured.

---

## Questions

* **Custom Provider Resolution**: If an operator configures an offline, third-party provider that is not bundled, will the system fallback to standard external import resolution, and does Gate 1 bypass integrity verification for such unbundled providers?
