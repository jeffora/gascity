# Tomas Park — DeepSeek V4 Flash (Test-Slicing Coverage Review, Iteration 4)

**Verdict:** block

**Lane:** Implementation slicing, acceptance traceability, behavior-oriented tests, and migration gate ordering.
**Scope:** Reviewed against the latest `design.md` (updated 2026-06-07) and the live tree at `/data/projects/gascity`.

---

## Overall Assessment

The Iteration 4 design is a significant advancement. It addresses several feedback items from prior reviews, particularly the integration of a **Staged Rollout Approach** with 7 sequential slices, and the introduction of the **Required Core Identity and Loader Contract**. These are correct architectural bounds.

However, I must **block** the design from entering the implementation phase due to critical structural risks, unresolved dependency cycles, and deferred decisions that will cause global test breakages or circular rollout loops. 

---

## Critical Risks (Blockers)

### 1. [Blocker] Slice 3 Global Test Suite Breakage due to Hardcoded Test References
Slice 3 proposes moving Core pack assets from `internal/bootstrap/packs/core` to `internal/packs/core` and removing `internal/bootstrap/packs/core`. However, multiple test files outside the builtinpacks package still hardcode references to the old bootstrap path. Committed as planned, Slice 3 will instantly break these tests globally:

- **`cmd/gc/prompt_test.go` (Lines 781–782):**
  ```go
  "internal/bootstrap/packs/core/assets/prompts/pool-worker.md",
  "internal/bootstrap/packs/core/assets/prompts/graph-worker.md",
  ```
- **`internal/config/bundled_import_test.go` (Line 44):**
  ```go
  writeTestFile(t, cacheDir, "internal/bootstrap/packs/core/pack.toml", `...`)
  ```
- **`examples/gastown/precompact_hook_test.go` (Line 67):**
  ```go
  for _, root := range []string{"examples", "internal/bootstrap/packs"} {
  ```
- **`test/packlint` Suite Files:**
  - `test/packlint/bd_show_jq_test.go` (Line 45): `"internal/bootstrap/packs"`
  - `test/packlint/gc_nudge_form_test.go` (Line 28): `"internal/bootstrap/packs"`
  - `test/packlint/gc_session_peek_form_test.go` (Line 18): `"internal/bootstrap/packs"`

The Slice 3 "required gates" only execute `go test ./internal/builtinpacks ./internal/hooks ./internal/bootstrap`. They do **not** run the full `make test-fast-parallel` suite. This means the global breakages in `cmd/gc`, `internal/config`, and `test/packlint` will go uncaught at the slice gate, violating the core invariant that every slice must be test-green before moving to the next.

> [!IMPORTANT]
> **Required Change:** Update the design for Slice 3 to explicitly include updating all hardcoded test references. Expand Slice 3's required gates to execute the affected test suites: `go test ./cmd/gc`, `go test ./internal/config`, and `go test ./test/packlint/...`.

---

### 2. [Blocker] Deferring Audits to "Moves/Deletes" Slices Introduces Circular Rollout Loops
The design states: *"They [audits] must be resolved in the slice that moves or deletes the relevant source"* (lines 950–951). However, the rollout model freezes the public `gascity-packs/gastown` repository at an immutable pin (`PublicGastownPackVersion`) in Slice 1.

If an audit conducted during Slice 5 (Maintenance folding) or Slice 7 (Source deletion) reveals that a shared asset (such as `mol-review-quorum` or a Gastown-specific prompt fragment) actually belongs in Gastown, the developer will be forced to unfreeze `gascity-packs`, commit the change, create a new immutable commit, and backport the pin update to Gas City. This violates the rollout flow of Slice 1 and 2 and creates a circular loop.

> [!IMPORTANT]
> **Required Change:** All cross-pack ownership audits and decisions listed in the table must be fully resolved and committed to `gascity-packs` **before or during Slice 1**. Slices 1 and 2 must use a pin that already contains all final resolved Gastown assets.

---

### 3. [Blocker] Test Loader Incompatibility / Dependency Cycle in Slice 2
Slice 2 introduces `TestPinnedPublicGastownBehavior` to prove "Core plus public Gastown with no Maintenance pack." However, in Slice 2, the production loader still hardcodes `"maintenance"` in `requiredBuiltinPackNames` (which is not removed until Slice 5).

This creates a dependency cycle:
- If `TestPinnedPublicGastownBehavior` uses the production loader, it is impossible to assert behavior "with no Maintenance pack" because the loader will force-include Maintenance.
- If the test bypasses the production loader, it violates the requirement of executing normal config resolution paths.

> [!IMPORTANT]
> **Required Change:** Slice 2 must explicitly update the production config loader to dynamically support a test-only context or flag that disables Maintenance, or `TestPinnedPublicGastownBehavior` must be moved to Slice 6 (Registry/cache cleanup slice) where Maintenance is retired.

---

### 4. [Blocker] Unresolved State Fallback Migration for `jsonl_archive_doctor_check.go`
In `cmd/gc/jsonl_archive_doctor_check.go` (Lines 96 & 114), the doctor check's fallback logic for locating the export state and archive repository is hardcoded to `.gc/runtime/packs/maintenance`.

When Maintenance is retired, newly initialized Core-only cities will lack this directory, causing the doctor check to look in a non-existent path. The design under "Maintenance Retirement Runtime Table" (lines 540–552) says: *"Maintenance runtime paths are ignored-legacy, migrated to Core, or manual-diagnostic per state file"* but does not make a concrete decision.

Since the moved scripts (`jsonl-export.sh`, `reaper.sh`) will now run from Core, their state file resolution must be consistent with the doctor check. Leaving the path undecided at the design phase risks cross-file drift.

> [!IMPORTANT]
> **Required Change:** Make a concrete design decision: state files and archives must be migrated to Core (`.gc/runtime/packs/core`), and both the scripts and the doctor fallbacks (`resolveStateFile`, `resolveArchiveRepo`) must be updated in Slice 5 to use the new Core path.

---

### 5. [Blocker] Missing Call-Site Inventory and Partial-Read Loader Interface
The design bans direct `config.Load*` calls in production `cmd/gc/*.go` files using a scanner, but does not inventory the existing call sites. There are at least **18 production call sites** that currently call direct loader functions (e.g., `cmd/gc/apiroute.go`, `api_state.go`, `doctor_v2_checks.go`).

A blanket ban on day 1 will cause immediate compilation or test failures. The design mentions "starts from an inventory" but does not define a refined loader interface (e.g., `loadCityConfigPartial`) to keep the allowlist small and manageable.

> [!IMPORTANT]
> **Required Change:** Inventory the 18 call sites, define a refined loader interface for legitimate partial-read exceptions, and specify which ones will be migrated to the system-pack wrapper vs. allowlisted.

---

## Major Risks

### 6. [Major] Missing Behavior Inventory Mapping for `maintenance_scripts_test.go`
The design requires that "every removed test function must map to a new `gascity-packs` test," but does not provide this mapping or specify which tests belong in Core vs. Gastown.

`maintenance_scripts_test.go` contains **~50 distinct test functions** asserting complex runtime behaviors (e.g., orphan sweeps, stale session cleanups, Dolt port resolution, and session liveness). Many of these scripts are moving to Core rather than Gastown, meaning their test coverage must move to Core-level tests.

> [!WARNING]
> **Required Change:** Provide an exhaustive test-migration mapping for all 50 functions, classifying each as Core-bound or Gastown-bound.

---

### 7. [Major] Underspecified Provenance for `legacyPublicPackForSource`
The function `legacyPublicPackForSource` in `import_state_doctor_check.go` checks whether a source string matches `.gc/system/packs/gastown` or `examples/gastown/packs/gastown` without distinguishing how that directory was created (embedded materialization vs. public pack install).

The design says: *"Specify whether this is done via the import Source field... or another mechanism"* (lines 271–273), but does not choose or specify the mechanism in the proposed design section.

> [!WARNING]
> **Required Change:** Choose the import `Source` field check as the canonical mechanism and specify its exact implementation details.

---

### 8. [Major] Underspecified Role-Token Scanner Specification
The design mandates a "role-token scanner" over Core TOML, shell scripts, templates, overlays, and tests, but provides no concrete specification. Without defining the tool, rules, or test fixtures, implementing this scanner will cause friction or false positives/negatives during Slice 5.

> [!WARNING]
> **Required Change:** Define the exact tool (e.g., static analysis script/regex rules) and the reviewed allowlist files/fixtures.

---

## Minor Risks

- **Inconsistent Vet/Lint Gates:** Slice 7's required gate includes `go vet ./...`, but earlier slices do not. If intermediate slices introduce vet violations, they will not be caught until the very end, breaking the "independently deployable and clean" assertion. `go vet` should run on every slice.
- **Allowed Callers Allowlist Update:** `internal/builtinpacks/registry_test.go` (Line 130) contains a hardcoded allowed-caller map: `allowed := map[string]bool{"internal/packman/cache.go": true}`. When the registry package changes in Slice 6, this test may break and require updates.

---

## Detailed Slicing Audit Table

| Slice | Focus | Verification Gap / Expected Breakage | Tomas's Verdict / Gating Feedback |
|---|---|---|---|
| **Slice 1** | Candidate public Gastown | Audit of deferred assets (`mol-review-quorum`) is missing, leading to future unfreezing of the repo. | **Block** — Move all ownership audits and decisions into Slice 1 before pinning the commit. |
| **Slice 2** | Gas City packcompat | `requiredBuiltinPackNames` still forces Maintenance to load, making a "no Maintenance" check impossible. | **Block** — Postpone or adjust loader hooks to make no-Maintenance validation possible. |
| **Slice 3** | Core extraction | Hardcoded paths in `prompt_test.go`, `bundled_import_test.go`, and packlint tests break globally. | **Block** — Add test updates to Slice 3; expand required gates to run affected test suites. |
| **Slice 4** | Core loading/doctor | `config.Load` scanner lacks call-site inventory, immediately rejecting 18 production files. | **Block** — Inventory all 18 call sites and introduce cleaner loader wrapper patterns first. |
| **Slice 5** | Maintenance folding | Fallback in `jsonl_archive_doctor_check.go` looks for retired path; ~50 tests in `maintenance_scripts_test.go` lack explicit mapping. | **Block** — Update doctor check fallback to use `packs/core`; provide an exhaustive test-migration map. |
| **Slice 6** | Registry/cache | Potential failure in `registry_test.go` due to allowed-caller map. | **Pass with Caveats** — Monitor and update allowed-caller map if needed. |
| **Slice 7** | Source deletion/docs | None. | **Pass** — Verify all docs use unified system-pack terminology. |
