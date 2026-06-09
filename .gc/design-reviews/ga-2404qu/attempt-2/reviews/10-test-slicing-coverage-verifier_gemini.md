# Tomas Park — DeepSeek V4 Flash (Test-Slicing Coverage Review, Iteration 3)

**Verdict:** block

**Persona Focus:** Tomas Park — Implementation slicing, acceptance traceability, behavior-oriented tests, and migration gate ordering.

---

## Top Strengths

- **Staged Rollout Approach:** Breaking the migration into 7 distinct, sequential implementation slices prevents a single high-risk "flag-day" release. Each slice defines target behaviors and required gates to prevent regression.
- **Traceability via Behavior Inventory:** The requirement for `gastown/docs/behavior-preservation.md` with explicit mapping of old path, trigger conditions, semantic delta, and landing commits provides excellent traceability for Gastown behavior.
- **Runtime Materialization Safeguards:** The `assertRequiredSystemPackProvenance` check enforces that Core is present in the final resolved config, transforming a silent loading failure into a loud, fast startup error.
- **Preservation over Deletion:** Stale `.gc/system/packs/maintenance` directories are ignored but preserved, ensuring operators' custom edits are not aggressively deleted by `gc doctor --fix`.

---

## Critical Risks (Blockers)

### [Blocker] Global Test Breakage in Slice 3 due to Hardcoded Test References

Slice 3 proposes moving the Core pack assets from `internal/bootstrap/packs/core` to `internal/packs/core` and updating builtin registry imports. However, multiple existing test suites contain hardcoded references to the old bootstrap path. Under the proposed plan, these tests will break globally as soon as Slice 3 is committed:

1. **`cmd/gc/prompt_test.go` (Lines 781-782):**
   ```go
   "internal/bootstrap/packs/core/assets/prompts/pool-worker.md",
   "internal/bootstrap/packs/core/assets/prompts/graph-worker.md",
   ```
2. **`internal/config/bundled_import_test.go` (Line 44):**
   ```go
   writeTestFile(t, cacheDir, "internal/bootstrap/packs/core/pack.toml", `...`)
   ```
3. **`examples/gastown/precompact_hook_test.go` (Line 67):**
   ```go
   for _, root := range []string{"examples", "internal/bootstrap/packs"} {
   ```
4. **`test/packlint` Suite Files:**
   - `test/packlint/bd_show_jq_test.go` (Line 45): `"internal/bootstrap/packs"`
   - `test/packlint/gc_nudge_form_test.go` (Line 28): `"internal/bootstrap/packs"`
   - `test/packlint/gc_session_peek_form_test.go` (Line 18): `"internal/bootstrap/packs"`

The Slice 3 "required gates" only list `go test ./internal/builtinpacks ./internal/hooks ./internal/bootstrap`. They do **not** run the full `make test-fast-parallel` test suite. Consequently, the global test breakages in `cmd/gc/prompt_test.go`, `bundled_import_test.go`, and the packlint tests will not be caught by the slice's focused gate, but will cause the global invariant (`make test-fast-parallel` passes) to fail immediately.

> [!IMPORTANT]
> **Required Change:** Slice 3's asset moves and path updates must explicitly include corresponding updates to these hardcoded test paths. The required gates for Slice 3 must also execute the affected test packages (`go test ./cmd/gc`, `go test ./internal/config`, `go test ./test/packlint/...`) to prove that intermediate commits keep the entire suite green.

---

### [Blocker] Deferring Cross-Pack Ownership Decisions to Slice 5 Creates Circular Rollout Loops

The design defers critical cross-pack ownership audits/decisions (such as `mol-review-quorum`, Gastown Codex overlay prompts, and Dog prompt fragments) to Slice 5 (the folding slice). However, the rollout model establishes in Slice 1 that the `gascity-packs/gastown` repository must be frozen and pinned at an immutable commit (`PublicGastownPackVersion`).

If an audit conducted during Slice 5 reveals that a shared asset (like `mol-review-quorum` or a Gastown-specific prompt fragment) actually belongs in Gastown, the developer will be forced to:
1. Un-freeze the public `gascity-packs` repository.
2. Land the audited changes there.
3. Generate a new immutable commit.
4. Backport and update the `PublicGastownPackVersion` pin in the Gas City codebase.

This violates the rollout requirement of Slice 1 and introduces a circular dependency between down-stream implementation steps.

> [!IMPORTANT]
> **Required Change:** All cross-pack ownership audits and decisions must be resolved **before or during Slice 1**. The public pack's pinned commit in Slice 1 must already contain all necessary Gastown role assets, prompts, and formula overrides so that downstream slices do not require repeated repository un-freezing and re-pinning.

---

### [Blocker] Test Loader Incompatibility / Dependency Cycle in Slice 2

Slice 2 introduces the compatibility test `TestPinnedPublicGastownBehavior` to verify "Core plus public Gastown with no Maintenance pack." However, in Slice 2, the host binary still hardcodes `"maintenance"` in `requiredBuiltinPackNames` (which is not removed until Slice 5). 

The production system loader cannot exclude the Maintenance pack in Slice 2. This creates a logical contradiction:
- If `TestPinnedPublicGastownBehavior` uses the production loader, it is impossible to assert behavior "with no Maintenance pack" because the loader will force-include Maintenance.
- If the test bypasses the production loader, it violates the "behavior-oriented loading assertion requirement" by using synthetic mocks instead of validating actual runtime config resolution.

> [!IMPORTANT]
> **Required Change:** The design must resolve this dependency cycle. The packcompat test should either be moved to a later slice (such as Slice 6, after Maintenance has been retired from the required list) or the production loader must be updated earlier to dynamically allow disabling Maintenance via a test-only config context.

---

### [Blocker] Missing Call-Site Inventory and Allowlist for `config.Load` Scanner

The proposed scanner test (modeled on `TestGCNonTestFilesStayOnWorkerBoundary`) rejects direct `config.Load`, `config.LoadCity`, and `config.LoadWithIncludes` calls in production `cmd/gc/*.go` files. However, the design does not inventory the existing production call sites. A quick search reveals at least **18 production call sites** that will instantly fail this scanner:

- `cmd/gc/apiroute.go` (2 direct `config.Load` calls)
- `cmd/gc/cmd_config.go` (1 `config.Load` for quick-check path)
- `cmd/gc/cmd_import.go` (2 `config.Load` + 3 `config.LoadRootPackDefaultRigImports` calls)
- `cmd/gc/api_state.go` (1 direct `config.Load` call)
- `cmd/gc/legacy_pack_preflight.go` (1 `config.Load` for quick-check path)
- `cmd/gc/doctor_v2_checks.go` (4 `config.Load`/`config.LoadSiteBinding`/`config.LoadPackGraphDirsForDoctor` calls)
- `cmd/gc/cmd_lint.go` (1 `config.LoadPackForLint` call)

An allowlist of 18 entries is too large and indicates that the scanner mechanism is too crude or the codebase is not yet ready for a strict blanket ban.

> [!IMPORTANT]
> **Required Change:** The design must explicitly inventory these 18 production call sites, classify each as "migrate to wrapper" versus "legitimate partial-read exception," and define a refined loader interface (e.g., `loadCityConfigPartial`) so the scanner can be successfully introduced without a massive, fragile allowlist.

---

## Major Risks

### [Major] Doctor Fallbacks Defaulting to Retired Path in `jsonl_archive_doctor_check.go`

In `cmd/gc/jsonl_archive_doctor_check.go` (Lines 64 & 87), the doctor check's fallback logic for locating the export state and archive repository is hardcoded to a retired path:
```go
func (c *jsonlArchiveDoctorCheck) resolveStateFile() string {
    ...
    base = filepath.Join(runtime, "packs", "maintenance")
    ...
}
```
For a newly initialized, Core-only city where the Maintenance pack is completely retired and `.gc/runtime/packs/maintenance` does not exist, this check will look in a non-existent directory. It fails to dynamically resolve the active required pack directory (which should be `packs/core`).

> [!WARNING]
> **Required Change:** The fallback logic in `jsonlArchiveDoctorCheck` must be updated to resolve state paths dynamically based on the active pack configuration (e.g., falling back to `packs/core`) rather than relying on a hardcoded, retired Maintenance pack path.

---

### [Major] Incomplete Behavior Inventory Mapping for `maintenance_scripts_test.go`

The behavior inventory in `design.md` only specifies 6 "high-risk moves" (Maintenance `dog` prompt, `mol-shutdown-dance`, `mol-dog-jsonl`, `mol-dog-reaper`, `prune-branches`, and `mol-polecat-*` formulas). However, `examples/gastown/maintenance_scripts_test.go` contains **~50 distinct test functions** asserting complex runtime behaviors (e.g., orphan sweeps, stale session cleanups, Dolt port resolution, and session liveness).

Requiring that "every removed test function must map to a new `gascity-packs` test" without an exhaustive list of all 50 test functions risks massive coverage gaps. Many of these scripts are moving to Core rather than Gastown, meaning they need Core-level tests rather than public pack tests.

> [!WARNING]
> **Required Change:** Expand the behavior inventory requirements to exhaustively classify all test functions within `maintenance_scripts_test.go`. Explicitly designate whether each test maps to a new Core test file or a new public Gastown test file.

---

### [Major] Lack of Role-Token Scanner Specification and Fixtures

The design mandates a "role-token scanner" over Core TOML, shell scripts, templates, overlays, and tests, but provides no concrete specification:
- No implementation exists in the current codebase.
- No test fixtures are defined to show what is accepted versus rejected.
- It is unclear how the scanner will parse indirect role references in shell scripts (e.g., `gc mail send mayor/` or `gc session nudge deacon/`) or template variables (e.g., `{{.AgentBase}}`).

---

## Minor Risks

- **Inconsistent Lint Gates:** Slice 7's required gate includes `go vet ./...`, but earlier slices do not. If intermediate slices introduce vet violations, they will not be caught until the very end, breaking the "independently deployable and clean" assertion. `go vet` should be ran on every slice.
- **Allowed Callers Allowlist Update:** `internal/builtinpacks/registry_test.go` (Line 130) contains a hardcoded allowed-caller map: `allowed := map[string]bool{"internal/packman/cache.go": true}`. When the registry package changes in Slice 6, this test may break and require updates, which is currently unmentioned.

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
