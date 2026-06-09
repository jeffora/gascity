# Petra Novak — DeepSeek V4 Flash Perspective Independent Review (Iteration 5 / Attempt 5)

**Verdict:** approve-with-risks

**Scope:** Builtinpacks registry, embed path migration, Maintenance retirement, and downstream reference safety.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this review against the current `plans/core-gastown-pack-migration/requirements.md` (specifically `.gc/design-reviews/ga-dtvdnd/attempt-5/design-before.md`, 135 lines, status updated 2026-06-09), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` dog assets this migration retires, and the public `gascity-packs/gastown` pack source. I did not inherit prior conclusions; all findings are re-verified against the tree.
2. **Dual-Placement Strategy.** Due to a known workflow defect documented in `attempt-4/synthesis.md` (where `gc.attempt=1` on beads causes them to write to `attempt-1/reviews/` and block attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/embed-materialization-build-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-5/reviews/embed-materialization-build-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 5 synthesis.
3. **Verdict Rationale.** The Attempt 5 requirements and the corresponding implementation plan represent a massive structural improvement, successfully defining a deterministic path for Core loading, public Gastown resolution, and Maintenance retirement. However, from the strict perspective of **Embed, Materialization & Build Reviewing**, several critical compile-time leaks, file-system concurrency races, and portability edge cases remain un-gated by the current Acceptance Criteria. I award an **APPROVE-WITH-RISKS** verdict and mandate three critical pins to secure our compile-time and materialization safety.

---

## Evaluation of the Three Key Questions

### 1. When Core moves to its canonical embedded location, are all importers, embed.go files, registry entries, hook code, and generation commands updated together?
**Reviewer Finding: Yes.**
The requirements and implementation plan successfully align Core's canonical destination:
- **Canonical Path:** Core moves to `internal/packs/core` as the unified system pack.
- **Unified Embedding:** `internal/packs/core/embed.go` embeds all Core assets with `PackFS`.
- **Reference Closure:** All canonical importers—including `internal/builtinpacks/registry.go`, `internal/hooks/hooks.go`, hook tests, and bootstrap tests—are updated to import `internal/packs/core` instead of the legacy `internal/bootstrap/packs/core`, followed by physical deletion of the legacy folder.

### 2. Does builtin materialization stop embedding and auto-including retired Maintenance while still materializing Core, bd, and dolt correctly?
**Reviewer Finding: Yes.**
The Go-level compile-time changes are thoroughly scoped:
- **Registry Pruning:** `internal/builtinpacks/registry.go` is pruned of its `maintenance` compile-import (`:19`), its `All()` slice registration (`:56`), and its `publicSubpathForPack` alias case (`:128`).
- **Materialization Removal:** `cmd/gc/embed_builtin_packs.go` is updated to remove `"maintenance"` from `requiredBuiltinPackNames` (`:237`) and the implicit runtime load inclusion (`:265`). 
- **Isolation Protection:** AC3 and AC5 ensure that required Core and provider-conditioned `bd`/`dolt` resolve deterministically without any implicit fallback or dependency leakage.

### 3. Are downstream references to moved Maintenance scripts repointed to Core homes without dangling paths?
**Reviewer Finding: Partially Resolved.**
The AC6 Asset Ledger and AC7 Behavior Manifest successfully map all source files and logical behaviors to Core or public Gastown homes. However, from a strict build and runtime perspective, the *downstream consumers* of moved/retired scripts are not explicitly gated by acceptance criteria.
- **The Risk:** Sourced scripts (e.g., `examples/dolt/assets/scripts/port_resolve.sh` coupling with `dolt-target.sh`) and state-file doctor checks (e.g., `jsonl_archive_doctor_check.go:59-60` searching `packs/maintenance`) will fail silently at runtime if not cleanly repointed or retired. 
- **The Solution:** The implementation plan recognizes these sites, but we must enforce an explicit downstream reference closure verification gate to ensure zero stale string references survive the migration.

---

## Critical Risks & Materialization Gaps (Lane Findings)

### 1. [Major] The Compiling Test-Fixture Decay (The Mock Embed Leak)
* **The Risk:** Go tests often use mock registry configurations or custom mock assets to isolate tests from disk operations.
* **The Gap:** If these unit tests mock the `builtinpacks` registry but continue to include hardcoded strings or embedded assets representing the retired `maintenance` pack, the tests will continue to compile and pass. This creates a dangerous divergence where the test suite asserts and relies on retired behavior that is completely absent in production, masking compile-time drift.
* **The Pin:** All mock registry test fixtures and inline FS builders must be scanned and pruned of any reference to `maintenance`, verified by an automated CI scan (under AC13) that forbids mock test fixtures from referencing retired packs.

### 2. [Major] The Partial/Orphaned Materialization Race on SIGINT
* **The Risk:** During CLI startup or initialization, `gc` materializes embedded system packs by extracting files from Go `embed.FS` and writing them to `.gc/system/packs/{name}`.
* **The Gap:** If an operator terminates the execution (via `SIGINT` / Ctrl+C) mid-extraction, the `core` pack directory will be partially written. Because `gc` optimized startup by checking directory existence to skip materialization on subsequent runs, it will assume Core is fully present. This will lead to cryptic runtime failures (such as missing prompt templates or failed hook scripts) that are extremely difficult for operators to diagnose.
* **The Pin:** System pack materialization must be transaction-atomic. The extraction tool must write all files to a temporary staging directory (e.g., `.gc/system/packs/.tmp_core`) and atomically rename it to `.gc/system/packs/core` only after all assets have been fully written and verified against the embedded manifest.

### 3. [Major] The Symlink/Hardlink Portability Pitfall in Pack Storage
* **The Risk:** Moving scripts like `dolt-target.sh` and `reaper.sh` into Core or public Gastown might tempt developers to use symlinks or hardlinks to avoid code duplication.
* **The Gap:** Go’s `embed` package does not support symlinks, and materializing symlinks or hardlinks across different target platforms (such as Windows, MacOS, or nested Docker containers) frequently fails or loses executable permissions, causing critical runtime crashes.
* **The Pin:** The AC6 asset ledger validation tool must strictly prohibit symlinks, hardlinks, or relative pointer files in the embedded assets or materialized pack layouts, requiring all duplicated script logic to be resolved via explicit imports or clean Go helpers.

---

## Required Changes for Finalization (Actionable Pins)

1. **Mock Fixture Clean-up Scans:** Update AC13 to mandate that all mock registry test fixtures, test caches, and inline FS builders are scanned and pruned of `maintenance` or legacy `internal/bootstrap/packs/core` references, ensuring test assertions match production constraints.
2. **Atomic Materialization Staging:** Amend AC3 to require that system pack materialization utilizes temporary directory staging and atomic `os.Rename` operations to prevent corrupted, partial pack extractions on abrupt CLI termination.
3. **Symlink and Hardlink Prohibition:** Update AC6 to explicitly forbid symlinks or hardlinks within the embedded source files and materialized pack layouts, ensuring complete platform portability for materialized system packs.

---

## Open Questions for Implementation

* **How will the atomic staging directory handle permissions on Windows?** Go's `os.Rename` is atomic on POSIX, but on Windows, it fails if the destination folder already exists. The materializer must safely attempt to clean up or replace existing folders using non-interactive, fail-closed operations.
* **Should the test suite run with a read-only filesystem check?** A pre-commit test that runs `gc` with a read-only `.gc/system/packs` directory (forcing it to rely only on the embedded file-set where allowed) would be a powerful proof of Core's self-sufficiency.
