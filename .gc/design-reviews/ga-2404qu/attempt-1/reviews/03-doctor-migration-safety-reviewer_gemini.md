# Sofia Khoury — DeepSeek V4 Flash Independent Review (Core and Gastown Pack Split)

**Verdict:** BLOCK

**Lane:** `gc doctor --fix` idempotency, legacy import rewrites, custom data preservation, operator-safe diagnostics.

Reviewed strictly against the revised design document (`design-after.md` / Attempt 1) and requirements (`requirements.md`) in the live repository workspace.

---

## Executive Summary

As the Doctor and Migration Safety Reviewer, I have performed a rigorous, first-principles audit of the updated design document (`design-after.md`). Other reviewers (such as Gemini) have approved this design too quickly, accepting high-level safety assertions without stress-testing the underlying filesystem, process namespace, and concurrency invariants. 

While the design makes commendable progress—specifically around the isolation of stale pack directories (ignoring rather than deleting them) and the introduction of a centralized `doctor.MutationCoordinator`—it contains **three critical safety-critical blockers** and several high-risk assumptions that violate the core integrity of the doctor's diagnostic and repair boundary.

Specifically, the design:
1. **Self-contradicts on plain-doctor read-onliness**, directing a plain diagnostic run to call a materializing configuration loader that performs implicit side-effects.
2. **Creates a "Concurrency Illusion"** where an advisory lock on the city directory is assumed to protect a live running controller from reading torn, half-written configuration states when the controller itself does not participate in or respect the lock.
3. **Defines a false model of "idempotency"** that permits redundant write system calls (updating file modification times and inodes), which will cause infinite file-watcher thrashing and process reloads in live environments.

Consequently, my verdict is a firm **BLOCK**. Below is the detailed breakdown of the blockers, critical risks, and required design modifications.

---

## 1. Safety-Critical Blockers

### Blocker 1.1: Direct Contradiction on Plain-Doctor Read-Onliness (Lines 2109–2120 vs. Lines 2763–2766)
* **The Claim:** The Attempt-14 "Read-Only Doctor Diagnostic Boundary" (Lines 2106–2120) mandates that plain `gc doctor` (without the `--fix` flag) is strictly read-only, must call only `LoadRuntimeCityNoRefresh` / `ValidateRequiredFileSetsNoRefresh`, and **"must not call materializing runtime loaders, ... repair helpers, ... or quarantine/prune writers."**
* **The Contradiction:** Under the "Core Presence Doctor" implementation section (Lines 2755–2766), the design instructs the plain-doctor check to:
  - *"Verify the full required Core file set using the same strict integrity gate as runtime config loading"* (Lines 2763–2764)
  - *"Load resolved config through `internal/systempacks.LoadRuntimeCity`"* (Line 2766)
* **The Hazard:** `LoadRuntimeCity` is explicitly defined as the *materializing* runtime loader that **"repairs missing/corrupt expected files and prunes or quarantines unexpected effective files before validation"** (Lines 2709–2715). 
* **The Impact:** If an operator runs a plain, report-only `gc doctor` diagnostic on a city where the Core directory has drifted or is partially corrupted, the doctor check will invoke `LoadRuntimeCity`, which will silently materialize files, prune files, or move unexpected files to quarantine. This is a direct, silent mutation of the filesystem during a read-only diagnostic run—bypassing the `--fix` gate and the `doctor.MutationCoordinator` entirely. It completely violates the read-only diagnostic boundary.

### Blocker 1.2: The Concurrency Lock Illusion (Advisory FD Locks vs. Unparticipating Controllers)
* **The Claim:** The design asserts that concurrent mutation and read safety are guaranteed because the `MutationCoordinator` takes an OS-level advisory lock (`flock` / `fcntl`) on the city directory file descriptor (Lines 445–448) and refuses if another live doctor is active (Lines 1669, 2828).
* **The Oversight:** Advisory locks only work if **all** readers and writers cooperatively acquire the lock before accessing the target files. 
* **The Hazard:** The live Gas City controller (`gc start` / daemon) does not acquire this city-directory advisory lock before loading or reloading configurations. Furthermore, Go's runtime config loading path is a simple file-read operation.
* **The Impact:** If `gc doctor --fix` fails mid-publish (e.g., after writing `city.toml` but before writing the rig `pack.toml` or updating lockfiles), there is a significant, unprotected window where the active controller can read a torn, inconsistent multi-file state. The advisory lock on the city directory provides zero protection to the controller. A concurrent reader will experience undefined behavior, potentially crashing or spawning incorrect agent sessions.

### Blocker 1.3: False Idempotency & File-Watch Thrashing (mtime/inode Inflation)
* **The Claim:** The design states that `gc doctor --fix` is idempotent because healthy cities remain "byte-identical" after a run (Lines 54, 452–454, 1977).
* **The Oversight:** Writing identical content to a file using the standard "write temporary file + atomic rename" pattern (Lines 449–451) is *not* a filesystem no-op.
* **The Hazard:** Even if the final bytes are identical, executing a `write` and `rename` syscall:
  1. Updates the file's modification time (`mtime`).
  2. Allocates a new `inode` on the filesystem.
* **The Impact:** Live production controllers and dev environments use file-system watchers (e.g., `fsnotify` or polling) on `city.toml` and pack configurations to trigger automatic reloads. By writing and renaming byte-identical files, `gc doctor --fix` will spuriously trip these watchers on every single invocation. This can lead to cascade reloads, thrashing, and active session terminations. True idempotency in a production SDK must mean **zero filesystem write/rename syscalls** on a healthy city.

---

## 2. Critical Risks & Missing Edge Cases

### Risk 2.1: Container PID and Namespace Blindness in Live Process Discovery
* **The Claim:** The doctor discovers running processes (like an active controller or another doctor) from "live runtime state rather than status files" to refuse mutation (Lines 180–181, 2827–2828).
* **The Gap:** The design does not specify *how* this discovery occurs. If it scans the process table (e.g., reading `/proc` or using `ps`), this approach is highly brittle and easily blinded in containerized or multi-tenant environments (Docker, Kubernetes namespaces, or systemd sandboxes).
* **The Hazard:** In a containerized deployment:
  - The doctor and the controller might run in different PID namespaces, making them mutually invisible via `/proc`.
  - `/proc` might be mounted read-only or restricted, causing the discovery scan to return empty results and silently bypass the safety check.
  - PIDs can clash across namespaces or recycle rapidly, leading to false positives (blocking valid repairs) or false negatives (permitting concurrent mutations).

### Risk 2.2: The Downgrade/Reverse Version-Skew Mutation Hazard
* **The Claim:** The rollback row promises that new-binary-mutated manifests remain readable by old binaries (Lines 1498–1504, 3176).
* **The Gap:** The design fails to address the reverse risk: what happens when an operator runs an *old* binary (`v1.2.1` or older) with the legacy `gc doctor --fix` command against a city that has *already* been migrated by the new binary.
* **The Hazard:** The legacy `v1.2.1` `--fix` code has no knowledge of the `MutationCoordinator`, the span-preserving TOML editor, public pinnings, or the migration markers. It sees the new public pin configuration as an unrecognized state or "legacy local import" and might blindly strip or overwrite `city.toml` using its legacy whole-file re-encoder, destroying all comment formatting and throwing the configuration back into a broken state.

### Risk 2.3: Unordered Preflight Actions on Offline Public Pins
* **The Claim:** If a public Gastown import is rewritten to a remote pin, the preflight check verifies reachability, installability, and lockability (Lines 444–448, 2817–2819).
* **The Gap:** The exact ordering of preflight checks relative to file edits is underspecified.
* **The Hazard:** If the doctor begins modifying `city.toml` or `pack.toml` *before* verifying that the remote public pin is reachable and installable, a network failure or a missing ordinary cache entry halfway through the run will leave the filesystem partially mutated and corrupted.
* **The Impact:** The entire preflight phase—including complete dry-run resolution, lockfile parsing, and remote pin validation—must occur in a strictly read-only sandbox *before* a single write or temporary file is created.

---

## 3. Evidence-Based Evaluation of Prior Claims

* **Gemini's Approval of Plain Doctor (Gemini Review, Line 17):** Gemini claimed the read-only check path loader contradiction was resolved because "plain diagnostics remain strictly read-only."
  - *DeepSeek Re-evaluation:* **False.** Gemini accepted the high-level boundary description in Section 14 but failed to inspect the low-level implementation contract in Section 2755 (Core Presence Doctor), which explicitly mandates calling the materializing `LoadRuntimeCity` loader.
* **Codex's Advisory Lock Recommendation (Codex Review, Line 13):** Codex accepted that an advisory lock on the city directory fd is the primary path to concurrency safety.
  - *DeepSeek Re-evaluation:* **Flawed.** Codex did not analyze whether concurrent readers (like the controller) actually acquire or respect the same lock. An advisory lock that only writers respect provides no multi-file atomicity protection to the active controller.

---

## 4. Required Design Changes

To transition this design from **BLOCK** to **APPROVE**, the following concrete modifications must be incorporated into the text:

### 1. Fix the Plain-Doctor Core Presence Check
* **Change:** Under `cmd/gc/core_pack_doctor_check.go` (Lines 2755–2767), explicitly forbid calling `LoadRuntimeCity`.
* **Replacement:** Force the plain-doctor diagnostic path to load configuration strictly using `LoadRuntimeCityNoRefresh` and check the file set via `ValidateRequiredFileSetsNoRefresh`. If the Core directory has drifted or is missing, report the error state as a diagnostic only and defer all materialization and repair to `gc doctor --fix` via the coordinator.

### 2. Implement True Idempotency (Zero-Write Invariant)
* **Change:** Require the `MutationCoordinator` to compute content digests of the staged files and compare them against the active on-disk files *prior* to executing any write or rename system calls.
* **Replacement:** If the staged content of `city.toml` and lockfiles is byte-identical to the active state, the doctor must **skip all write/rename syscalls entirely**, logging "city is already healthy; no files modified." This preserves the file `mtime` and `inode`, preventing file-watcher reload storms.

### 3. Establish a Two-Tier Concurrency Lock
* **Change:** The advisory lock must not be limited to the doctor.
* **Replacement:** Mandate that both the live Gas City controller (`gc start` / `internal/session/manager`) and the `MutationCoordinator` must participate in the same advisory lock protocol on the city directory. The controller must take a **shared lock (SH)** during normal operation and configuration reloading, while the doctor must take an **exclusive lock (EX)** during `--fix` execution. This ensures the controller can never read a torn multi-file state during doctor mutations.

### 4. Hardhen Live Process Discovery
* **Change:** Define a robust, multi-layered process discovery protocol.
* **Replacement:** Discovery of live active controllers or doctors must use a combination of:
  - The shared/exclusive directory lock (the primary source of truth).
  - Explicit platform-agnostic socket/port verification (e.g., verifying if the API control plane port is active and bound to the current city path).
  - A fallback process scan that filters specifically on the normalized absolute city path, with explicit handling for empty or restricted `/proc` environments.

### 5. Document Downgrade Release Invariants
* **Change:** Add an explicit safety entry in the compatibility matrix for old-binary `--fix` runs.
* **Replacement:** Specify that the migrated `city.toml` will include an explicit comment marker or an unrecognized structure designed to cause older binaries to fail-fast with a clear "unsupported configuration version" error rather than blindly overwriting the file.

---

## Sofia's Evaluation Questions Answered

### 1. Is the Core presence doctor fix a proven no-op on a healthy city?
**No.** As written, it writes byte-identical files which updates `mtime` and `inodes`, causing file-watcher thrashing in the controller. It requires a zero-write digest compare gate.

### 2. What prevents `gc doctor --fix` from deleting user-added imports?
The proposed path-based provenance check is insufficient. It must be upgraded to content-digest and generation-marker verification.

### 3. Does rewrite-to-remote verify reachability in preflight?
The preflight ordering is too vague. The design must explicitly state that remote pin reachability and cache validation must occur in the read-only preflight phase before any filesystem writes are staged.

---

**Sofia Khoury blocks the design pending these critical safety modifications.**
