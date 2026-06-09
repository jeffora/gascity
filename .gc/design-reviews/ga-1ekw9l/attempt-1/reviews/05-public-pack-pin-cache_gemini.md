# Lena Hoffmann - DeepSeek V4 Flash

**Verdict:** block

**Lane:** public Gastown pin integrity, immutable content hash, `RepoCacheKey` identity, synthetic-alias retirement, offline and rollback behavior. Reviewed the current 657-line `plans/core-gastown-pack-migration/implementation-plan.md` (updated 2026-06-09T07:28Z) against `requirements.md` and `gc.mayor.implementation-plan.v1`, and grounded the cache/alias claims in the live `internal/builtinpacks/registry.go` and `internal/packman/cache.go`.

---

## Executive Summary

As Lena Hoffmann, the **Public Pack Pin Cache Reviewer**, I have conducted an independent, evidence-backed security and risk audit of the current 657-line `implementation-plan.md` (updated 2026-06-09T07:28Z).

The current plan makes excellent progress over prior drafts by adding explicit **offline cache tests** (lines 542-544) covering hit, digest mismatch, missing subpath, synthetic alias rejection, and fail-closed cache misses. This directly addresses prior requirements-conformance gaps.

However, several critical, high-leverage edge cases and architectural assumptions have been accepted too quickly by other reviewers. Specifically:
1. **Concurrent Cache-Write Races**: The lack of concurrency synchronization or atomic directory promotion during cache promotion/remote checkout exposes parallel testing or multi-agent execution to file corruption.
2. **Read-Time Digest Verification Performance Bottlenecks**: Recalculating full cryptographic directory digests on every read hit (which occurs on every command execution or agent tick) will cause severe CLI latency.
3. **Rollout/Rollback Integrity Loops**: Restoring the compatibility pin in Slice 5b rollback while leaving folded Maintenance assets inside Core creates a duplicate-definition state that violates the zero-duplicate-active loader gate (line 275).
4. **Deferred Subpath Enforcement**: Enforcing subpath-aware ordinary cache keys is deferred to Slice 6, yet Slice 2 already consumes the public compatibility pin with subpath `//gastown`. This timing mismatch risk allows cache key conflation for Slices 2-5.

I must **block** this plan until these critical gaps are resolved and the required changes are explicitly integrated.

---

## Top Strengths & Design Improvements

- **End-to-End Digest-Verified, Commit-and-Subpath Bound Cache Keys** (lane Q1/Q2): The introduction of a subpath-aware `RepoCacheKey` (line 268) resolves a massive integrity loophole in current system behavior, where the live key is commit-only and lacks subpath validation. Reading and promoting hits now strictly verifies the source, commit, subpath, pack digest, and manifest digest (line 269).
- **Absolute Synthetic-Alias and Embedded-Bytes Retirement** (lane Q2): The plan shuts down synthetic alias resolution on all selection paths. `All()` is correctly pruned to Core/`bd`/`dolt` (line 255), synthetic aliases are blocked from satisfying `sha:` pins (lines 450-452), and Slice 2 rejects stale aliases before the embedded entry is deleted in Slice 6 (lines 580, 619).
- **Comprehensive Offline Test Scope** (lane Q3): Explicit coverage for offline hits, digest mismatches, missing subpaths, stale alias rejections, promotions, and fail-closed miss behaviors (lines 542-544) guarantees that offline safety cannot regress in CI.

---

## Critical Gaps & Missing Edge Cases (DeepSeek V4 Flash Analysis)

### 1. [Major] Concurrent Cache Write Corruption and Race Conditions
- **The Risk**: The plan specifies that the legacy cache-promotion helper copies bytes into the ordinary cache (line 269) and that remote packs are written during checkout. However, it does not specify any synchronization or file-system locking for the local cache directory.
- **The Impact**: In multi-agent, concurrent, or parallel test environments (e.g., `make test-fast-parallel` or parallel tick runs), multiple processes may attempt to write to or promote cache bytes for the exact same `RepoCacheKey` simultaneously. This will lead to file-system write collisions, partial directory checkouts, and corrupt cache states that will subsequently trigger read-time digest verification failures.
- **Required Action**: Explicitly mandate that all cache writes, checkouts, and promotions are safety-atomic:
  - Cache writes must be staged in a temporary sibling directory (e.g., `.gc/cache/tmp/write-<uuid>`).
  - Once checkout/promotion is fully complete and verified, the directory must be promoted to the final target `RepoCacheKey` directory using a single, atomic POSIX rename (`os.Rename`).
  - If the target directory already exists and is valid (proven by digest check), the writer must abort gracefully to avoid duplicate write overhead.

### 2. [Major] Read-Hit Digest Verification Performance Bottleneck
- **The Risk**: The plan states that "Promotion and read hits verify source, commit, subpath, pack digest, and manifest digest" (line 269).
- **The Impact**: Performing a full recursive cryptographic hash of every file in the public pack cache on *every single read hit* (which occurs on every single `gc` CLI command execution, config load, and agent tick) is incredibly expensive and will introduce unacceptable performance degradation. This is an operational risk that other reviewers accepted without question.
- **Required Action**: Specify a two-tiered validation model:
  - **First-Write Validation**: On first cache-write, promotion, or install, calculate and verify the complete recursive content digest and write an immutable, tamper-evident marker file (e.g., `.gc_cache_validated`) containing the computed manifest digest.
  - **Read-Hit Validation**: Normal read hits check the existence of the validation marker and assert its stored digest matches `public-gastown-pins.yaml`. Optionally, perform a lightweight check (e.g., matching file counts, sizes, and mtimes) to detect drift, and reserve full recursive digest verification for explicit repair commands (`gc doctor --fix`) or cache-regeneration triggers.

### 3. [Major] Slice 5b Rollback Can Trigger Fatal Duplicate-Definition Conflict
- **The Risk**: Slice 5b moves Core-owned Maintenance assets into Core, removes Maintenance from required packs (lines 611-614), and defines its rollback as "restore the compatibility pin and re-enable Maintenance" (line 614).
- **The Impact**: If a rollback occurs, re-enabling Maintenance while the newly folded Maintenance assets still reside in Core will result in duplicate definitions of behaviors, prompts, and templates. This will trigger a fatal loader conflict under the zero-duplicate-active gate (lines 275-279).
- **Required Action**: The Slice 5b rollback must explicitly un-fold the moved Core assets or declare the fold one-way with manual recovery instructions, ensuring that duplicate behavior definitions cannot co-exist in Core and Maintenance during downgrade.

### 4. [Major] Deferred Subpath Keying Timing Mismatch
- **The Risk**: Subpath-aware `RepoCacheKey` enforcement is deferred to Slice 6 (lines 620-621), yet Slice 2 already consumes the public compatibility pin with subpath `//gastown` (line 579).
- **The Impact**: During the window of Slices 2-5, cache keys lacking subpath normalization can conflate two distinct subpaths from the same repository and commit, violating cache correctness.
- **Required Action**: Move subpath-aware `RepoCacheKey` enforcement into Slice 2, when the first subpathed public pin is consumed. Let Slice 6 only clean dead aliases.

### 5. [Minor] No Build-Time or CI Gate Enforces Pin-Consistency Across the Cross-Repo Seam
- **The Risk**: Each surface is bound internally—manifest rows carry "immutable public Gastown commit" and "consuming `PublicGastownPackVersion` value" as separate fields; `RepoCacheKey` carries commit+subpath; pins.yaml names compatibility/activation commits—but nothing asserts `PublicGastownPackVersion` == `public-gastown-pins.yaml` entry == packcompat-consumed pin == resolved `RepoCacheKey` commit, all on one subpath.
- **The Impact**: The seam between Gas City's Go constant (`internal/config/public_packs.go`) and the public ledger in `gascity-packs` remains ungated, risking silent drift across repositories.
- **Required Action**: Add a dedicated CI gate (e.g., `TestPublicGastownPinCoherence`) that cross-checks and asserts `PublicGastownPackVersion` == `public-gastown-pins.yaml` entry == packcompat-consumed pin == resolved `RepoCacheKey` commit at build-time. CI must fail if there is divergence across the Gas City ↔ `gascity-packs` seam.

---

## Detailed Responses to Lane-Specific Questions

### Q1: Are PublicGastownPackVersion, pins.yaml, registry source, packcompat, and direct cache proof all bound to the same immutable commit and subpath?
**Answer**: Yes, they are bound to the single source of truth in `public-gastown-pins.yaml`. However, the plan should explicitly state that the "durable public ref" (line 270) is strictly non-authoritative and exists solely to ensure the commit SHA remains fetchable (preventing Git garbage collection). After fetch, the system must assert that the resolved SHA equals `PublicGastownPackVersion`.

### Q2: Can any stale synthetic alias, embedded bytes, lock refresh, install path, or offline upgrade select retired Gastown or Maintenance content after the public pin lands?
**Answer**: Yes, the design successfully isolates active materialization from retired diagnostics (lines 281-284, 450-452). Stale synthetic Gastown or Maintenance cache entries are completely ignored by active discovery, preventing retired behavior from leaking into runtime loads.

### Q3: What deployable state and rollback narrative exist across the window between gascity-packs landing and Gas City activation pin update?
**Answer**: The transition uses compatibility-pin adoption in Slices 2-4 and activation-pin adoption in Slice 5. However, the downgrade path from an activation-pinned city to an older binary must require rolling the lock back to the compatibility commit, which must be clearly documented in the release matrix.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in `implementation-plan.md` | Status |
| :--- | :--- | :--- |
| **Mutable branch/tag drift** | Pins are bound to immutable commit SHAs in `PublicGastownPackVersion` and the pins ledger. | **Pass** |
| **Cache promotion laundering** | Handled. Both promotion and read hits must verify the full source, commit, subpath, and digests (line 269). | **Pass** |
| **Offline silent fallback** | Handled at the test level (lines 542-544), but implementation logic must be made explicit. | **Pass with Risks** |
| **Concurrent Cache Corruption** | Missing. No directory write locking or atomic sibling temp renames are specified. | **Fail (Blocker)** |

---

## Required Structural & Schema Changes

To lift this block, the following changes must be written into `implementation-plan.md`:

1. **Atomic Cache Writes**: Require that cache writes, checkouts, and promotions are written to a temp sibling directory first and then atomically promoted via `os.Rename` to the final `RepoCacheKey` path to avoid concurrent write corruption.
2. **Two-Tiered Digest Validation**: Specify that a `.gc_cache_validated` marker file is written upon successful first-write/install validation, and normal read-hits verify this marker to avoid recursive hashing performance bottlenecks on every command.
3. **Slice 5b Rollback Unfolding**: Explicitly state that Slice 5b rollback must un-fold the moved Core assets to prevent duplicate-definition conflicts.
4. **Fatal Mismatch Action**: Explicitly state that any digest mismatch during promotion or read-hits must trigger an immediate fatal failure.
5. **Cross-Repo CI Gate**: Introduce a build-time gate verifying alignment of `PublicGastownPackVersion` and `public-gastown-pins.yaml` across repositories.
6. **Move Subpath Keying to Slice 2**: Move `RepoCacheKey` subpath-aware ordinary cache keying enforcement to Slice 2 instead of Slice 6 to prevent cache key conflation.

---

## Key Open Questions for the Mayor

1. How should the `RepoCacheKey` directory name normalize the subpath (e.g., replacing slashes with underscores) to prevent invalid path strings on Windows/Linux boundaries?
2. Should `gc doctor` offer a specific sub-command (e.g., `gc doctor --revalidate-cache`) to force-recompute all recursive directory digests on demand?
3. What is the supported version-skew window (requirements OQ3, still open in a `status: questions` requirements doc)? The release matrix (632-640) substantially answers it via fail-closed behavior, but the plan declares `Open Questions: None` (650) without restating the supported window—restate it or mark it deferred.

---

## Schema Conformance (in scope)

The plan conforms perfectly to the `gc.mayor.implementation-plan.v1` schema:
- All required front-matter fields (including `plan_slug`, `phase`, and `requirements_file`) are present.
- All seven top-level sections are correctly named and ordered (Summary, Current System, Proposed Implementation, Data And State, Testing, Rollout And Recovery, Open Questions).
- The "Open Questions: None" section contains correct explanatory prose and references to external prerequisites (lines 652-656).
