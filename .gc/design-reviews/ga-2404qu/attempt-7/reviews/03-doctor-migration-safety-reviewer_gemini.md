# Sofia Khoury — DeepSeek V4 Flash (Independent Review, Iteration 7)

**Verdict:** block

**Persona focus:** Doctor fix idempotency, legacy import rewrite safety, custom data preservation, operator-safe diagnostics, cross-file consistency, missed edge cases, pattern drift, and architectural coherence. This iteration re-evaluates the latest design (updated 2026-06-07T02:23:08Z, identical to the previous attempt's design) against the current codebase. The focus is on determining if previously identified blocking issues have been resolved, and identifying newly surfaced edge cases or architectural coherence gaps in the current migration design.

---

## Top strengths

- **Comprehensive Preflight Gating:** The design (lines 145-149, 395-398, 980-983) requires extensive validation (network reachability, installability, lockfile parseability, manifest editability) *before* any disk mutation occurs. This protects cities from ending up in half-migrated, broken states on offline or misconfigured machines.
- **Strict Preservational Mandate for Legacy Directories:** The design explicitly guarantees that legacy directories like `.gc/system/packs/maintenance` and `.gc/runtime/packs/maintenance` are preserved and never deleted, preventing accidental operator data loss (lines 213, 883, 1005).
- **Explicit Loader Wrapper Framework:** The separation of config-loading surfaces (e.g., `loadCityConfigRuntime`, `loadCityConfigPartialForDoctor`, etc., lines 441-448) is an elegant architectural choice that prevents lower-level partial loaders from inadvertently driving behavior decisions.

---

## Critical risks

### [Blocker] State-Migration Atomicity & Rollback Gap: Manifests vs. Runtime State

The design specifies that runtime state (such as JSONL archive state and spawn-storm ledgers) migrates from `.gc/runtime/packs/maintenance` to `.gc/runtime/packs/core` (lines 213-217, 1002-1005). It also stages manifest/lockfile writes using a `temp-file-plus-rename` approach to ensure failure-atomicity across manifests (lines 157-158, 984-986), ensuring that any preflight or edit failure leaves the city byte-identical.

However, there is no coordination between manifest mutation and runtime state migration:
1. If `gc doctor --fix` copies/moves the JSONL state files and then a subsequent preflight step or manifest edit fails (e.g., `installLockedImports` fails or `city.toml` write fails), the manifests will remain byte-identical (rolled back).
2. However, the runtime state files on disk will have already been modified/moved into `.gc/runtime/packs/core`!
3. This leaves the city in a corrupt, mismatched state where manifests point to legacy maintenance, but active runtime state is stored in Core, leading to duplicate exports or state loss.

**Required Resolution:**
The design must explicitly defer any runtime state filesystem modifications (copies/moves/renames) until *after* all preflight checks have passed and manifest edits have been successfully committed, or define a fully-atomic transaction/rollback mechanism that restores state files if the TOML edits fail.

---

### [Blocker] The Scoped TOML Editor Contradiction

The design promises that "TOML edits must be scoped. If the existing parser/editor cannot preserve comments, unknown tables, unknown fields, array order, formatting, and unrelated lock entries, the doctor must refuse the automatic fix" (lines 160-162, 408-410, 990-992).

The gap is that the current Go codebase relies on `toml.NewEncoder` which re-encodes the whole file, dropping comments and custom fields. Since essentially every production `city.toml` or `pack.toml` in real-world environments contains comments or custom metadata, a strict enforcement of the "refuse if cannot preserve" contract means `gc doctor --fix` will refuse to run on nearly every real city. This directly contradicts the design's own goal of providing automatic `--fix` migrations.

As written, an implementer can satisfy the auto-fix promise (re-encode, lose operator content) or the preservation promise (refuse, fix nothing on real cities), but not both.

**Required Resolution:**
The design must move from a conditional "if it cannot preserve" stance to an explicit, executable mechanism. It must either:
- **(a) Explicitly adopt a CST-preserving Go TOML editor** (such as `github.com/pelletier/go-toml/v2`),
- **(b) Specify a line-oriented parser/regex editor** that only modifies the specific `[[imports]]` table lines, keeping the rest of the file bytes unchanged, or
- **(c) Downgrade the Backward Compatibility promise** to admit that automatic `--fix` is unsupported for commented manifests, routing them to manual diagnostics.

---

### [Blocker] Content-Based Provenance vs. Path Suffix Fallacy for Legacy Imports

The design-before asserts that "Legacy local Gastown imports are auto-rewritten only when provenance matches known generated/example paths. Operator forks, edited local packs... are diagnostic/manual" (lines 163-165, 399-401, 993-996).

Suffix-based path checks (e.g., matching `.gc/system/packs/maintenance`) are insufficient for provenance:
1. If an operator made custom edits directly inside their local generated pack directory (e.g., adding custom scripts/formulas to `.gc/system/packs/maintenance/`), a path-only check will classify this as a "known generated path."
2. The doctor will auto-remove the import, and the directory will be silently ignored by resolved config, dropping the operator's custom edits from active behavior.

**Required Resolution:**
Provenance detection must be content-aware. The design must specify that the doctor check validates the directory's content against a synthetic hash of the pristine embedded pack. If any drift or unexpected files are detected, the doctor must bypass the auto-fix, flag it as custom, and output explicit manual instructions.

---

### [Blocker] Sequenced Dependency Gap between Core Presence and Import-State Fixes

The design registers `importStateDoctorCheck` and `corePackDoctorCheck` as sequential checks within the doctor framework (lines 930-977).

If the Core materialization or presence check fails, `importStateDoctorCheck` can still run and remove legacy Core/Maintenance imports. If the doctor subsequently fails to materialize Core (due to disk I/O, permissions, or missing assets), the city is left with no Core import and no materialized Core pack, resulting in a completely un-bootable state.

**Required Resolution:**
The design must specify a strict sequencing guard: `importStateDoctorCheck.Fix()` must assert that Core is fully materialized and verified on disk *before* attempting any manifest or import removals. If Core is invalid or absent, the import-state fix must immediately abort.

---

## Major risks

### [Major] No Swap-Pattern for `MaterializeBuiltinPacks` Concurrent Loads

`MaterializeBuiltinPacks` writes files into `.gc/system/packs/core/` (lines 947-953). Under active-controller operations, a concurrent reload or config-read will fail or load a partial configuration if it hits the directory mid-write (especially since `MaterializeBuiltinPacks` deletes and rewrites files). This violates the "no-status-files, query-live-state" safety pattern and leads to transient runtime crashes.

**Required Resolution:**
Mandate a "directory swap" pattern: materialize the Core pack into a temporary directory (e.g., `.gc/system/packs/.tmp-core`), then perform an atomic rename/directory swap to replace `.gc/system/packs/core` in a single system operation.

---

### [Major] Required Core Loader Bypass Blindspot: The API and Session Layers

The required Core static call-site scanner (lines 418-450) is scoped exclusively to `cmd/gc/` files. However, the API layer (`internal/api/`) and session layer (`internal/session/`) construct session managers and load configurations directly (e.g., `session_manager.go` and `session_resolution.go`). If these are not included in the call-site scanner, they could load config without required Core, leading to silent behavior failures for API-driven actions.

**Required Resolution:**
Extend the static call-site check and required Core wrapper enforcement to encompass `internal/api/` and `internal/session/`, ensuring all backend and API-driven execution paths honor required Core participation.

---

### [Major] Lack of Air-Gap/Offline Fallback Guidance for Doctor Fix

The preflight check verifies remote public Gastown availability (lines 145-149, 395-398). If an existing legacy city is run in an air-gapped or offline environment, the `--fix` reachability preflight will fail. The design does not specify what helpful guidance is printed to help the operator manually vendor or pre-populate the pack cache.

**Required Resolution:**
Specify the precise failure message for offline preflight checks, detailing instructions on how the operator can manually cache the pinned Gastown version or run doctor in an air-gapped mode.

---

## Missing evidence

- **No Concrete Scoped TOML Editor Engine:** The design provides no name or reference to a Go package capable of performing scoped, comment-preserving TOML table modifications.
- **No Content-Hash Specification for Legacy Provenance:** The design fails to detail how the pristine template state hash is computed or stored for verifying that legacy directories are unedited.
- **No Idempotency Assertions for Materialized Core Freshness:** The design lacks an explicit assertion that a healthy, materialized Core pack directory is completely untouched (and its freshness timestamp unmodified) across repeated `gc doctor --fix` runs.

---

## Questions

1. Is there an intention to use `github.com/pelletier/go-toml/v2` to support round-trip comment preservation, or is the team planning to implement a regex/line-oriented editor specifically for the `[[imports]]` table?
2. Why is the static required-core loading call-site scanner limited to `cmd/gc/` when the API handlers under `internal/api/` and `internal/session/` also perform config loads?
3. How will `gc doctor --fix` protect custom user files or custom scripts placed inside `.gc/system/packs/maintenance/` from being silently ignored or dropped when the legacy import is removed?
