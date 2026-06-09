# Priya Menon — DeepSeek V4 Flash Perspective Independent Review (Iteration 7)

**Verdict:** approve-with-risks

**Scope:** Required Core loading, pack registry behavior, import resolution mechanics, legacy import retirement, and multi-pack resolution precedence.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this review against the current `plans/core-gastown-pack-migration/requirements.md` (specifically `.gc/design-reviews/ga-dtvdnd/attempt-7/design-before.md`, status updated 2026-06-09), the `plans/core-gastown-pack-migration/implementation-plan.md` (updated 2026-06-09), the `gc.mayor.requirements.v1` schema, and the live `examples/gastown/packs/maintenance` assets. I did not inherit prior conclusions; all findings are verified fresh against the repository tree.
2. **Dual-Placement Strategy.** Due to a known workflow defect where the bead's metadata `gc.attempt=1` causes automatic execution tools to write and read from `attempt-1/reviews/pack-resolution-architect_gemini.md` (which can block attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/pack-resolution-architect_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-7/reviews/pack-resolution-architect_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 7 synthesis.
3. **Verdict Rationale.** The Attempt 7 requirements and implementation plan represent a remarkably mature, cohesive, and robust architecture for securing Gas City's dependency boundaries. By isolating Core as the required runtime foundation, retiring Maintenance as an active standalone pack, and defining fail-closed, bootstrap-safe loader diagnostics, the system successfully eliminates untraceable legacy fallbacks. However, from the strict perspective of **Pack Resolution Architecture**, a few subtle edge cases and systemic assumptions must be pinned down (such as the doctor boot-loop risk during eager initialization of bootstrap-only commands, the circular pin-lock across repositories, and the blindspot of skipped or dummy witness tests). I award an **APPROVE-WITH-RISKS** verdict and mandate three critical pins to ensure complete runtime determinism.

---

## Evaluation of the Three Key Questions

### 1. How does Core become the canonical required runtime pack across init, doctor, CLI load, and runtime resolution?
**Architect Finding: Fully Resolved.**
The requirements and implementation plan successfully establish required-Core loading across all lifecycle phases:
- **Initialization & CLI Load:** `gc init` configures Core as a required system-pack identity in resolved config for real cities, with trusted provenance that user imports cannot shadow.
- **Loader Encapsulation:** The implementation introduces `internal/systempacks` as the single production boundary for required host packs, exposing `LoadRuntimeCity` and `LoadRuntimeCityNoRefresh`. Callers do not hand-build includes.
- **Verification Gates:** Runtime loading enforces two fatal gates—pre-resolution validation of Core/provider file-sets (Gate 1), and post-resolution validation of typed `RequiredSystemPackParticipation` records (Gate 2). Bypasses are prevented via scanner tests in CI that reject direct calls to `config.Load*` in `cmd/gc` and behavior-driving packages.
- **Diagnostics:** Under AC11, `gc doctor` and `gc import-state` can run in bootstrap-only mode when pack resolution is broken, reporting missing Core with exact source layer attribution.

### 2. Are implicit Maintenance and in-tree examples/gastown imports explicitly retired while gc init --template gastown imports gascity-packs/gastown?
**Architect Finding: Fully Resolved.**
The migration cleanly retires legacy dependencies:
- **Maintenance Retirement:** AC5 enforces that Maintenance is no longer bundled, public-source recognized, auto-included, materialized, or selected. `internal/builtinpacks/registry.go` is pruned.
- **Legacy Path Authority:** `internal/packsource` is established as the sole classifier for retired paths. Legacy directories on disk are safely ignored during startup and active discovery.
- **Explicit Gastown Init:** AC4 mandates that `gc init --template gastown` configures the public `gascity-packs/gastown` pack explicitly via `pack.toml` with an immutable `sha:` pin, with in-tree and implicit fallbacks strictly forbidden.

### 3. What is the deterministic resolution order when Core, Gastown, bd, and dolt all participate?
**Architect Finding: Fully Resolved.**
Precedence order across required Core, provider-conditioned support packs (`bd`/`dolt`), and remote/local imports is explicitly specified. The implementation plan's "zero-duplicate-active and zero-merge gates" compare active bundled, active public, and cached sources, failing closed if the same behavior ID is active from more than one source or if Core and public Gastown both claim an unresolved behavior row. To make this resolution fully bulletproof, we must enforce strict design-level pins for diamond dependencies and concurrent staging.

---

## Critical Risks & Architectural Gaps (Architectural Findings)

### 1. [Major] Eager Loader Dependency in Bootstrap-Only Mode (The Doctor Boot-Loop)
* **The Risk:** AC11 and the implementation plan specify that `gc doctor` and repair commands have a bootstrap-only diagnostic mode that can run even when normal pack resolution is completely broken (e.g., when Core is missing or corrupted). However, if the CLI bootstrap process or any imported sub-package eagerly initializes the normal config loader or triggers required-pack validation before identifying that the command being run is a bootstrap-only repair command, the CLI will crash or exit with a fatal error. This results in a boot loop, completely blocking the operator from running `gc doctor --fix` to repair the broken state.
* **The Gap:** There is no explicit isolation between the eager loader initialization and the bootstrap-only command parsing at the CLI entry point.
* **The Pin:** The CLI bootstrap path must strictly detect bootstrap-only commands (`gc doctor`, `gc import-state`, `gc version`) at the very entry point, before any config resolution, required-pack materialization, or Go-side initialization of system-packs occurs. These commands must execute in a completely dependency-isolated path that bypasses Gate 1 and Gate 2 validation.

### 2. [Major] Multi-Repository Circular Pin-Lock (The SHA-Pin Deadlock)
* **The Risk:** The rollout policy (Slices 1b/1c/2/5a) defines a multi-repository promotion flow where `gascity` pins `gascity-packs/gastown` to an immutable `sha:`. However, the public Gastown activation commit cannot be fully tested in public CI without running against the final `gascity` binary that implements the Core split. This creates a circular dependency across the repositories: `gascity` cannot adopt the final pin until `gascity-packs` is committed, and `gascity-packs` cannot commit/verify the final activation mode until `gascity` is merged.
* **The Gap:** The implementation plan assumes the public pin SHA can be calculated and committed without a local verification escape during development.
* **The Pin:** The pin-coherence validator and the `test/packcompat` harness must support a `--local-override` flag or path-level mapping during development and multi-repository CI pipeline execution. This allows validating the candidate `gascity-packs` commit locally in tandem with the local `gascity` workspace before the SHA is finalized on origin, breaking the circular dependency.

### 3. [Major] The Skipped/No-Op Witness Blindspot (The Dummy Test Escape)
* **The Risk:** AC13 mandates that the `coverage-transfer.yaml` validator must fail closed on empty post-deletion walks, unmapped assertions, or behavior witnesses that only prove a no-op. However, if the validator only checks that a new witness test exists on the filesystem or matches a naming pattern, developers under pressure could map retired assertions to new tests that are skipped (`t.Skip()`), empty, or contain only dummy assertions. This would bypass the behavior-preservation guarantees.
* **The Gap:** Static presence or name-based checks in the validator cannot distinguish an active, asserting test from a skipped or placeholder test.
* **The Pin:** The coverage-transfer and behavior-preservation validators must analyze active Go test execution transcripts (e.g., parsing the JSON stream of `go test -json`) during validation runs. They must verify that every mapped new witness is actively executed and passes, explicitly failing closed if any witness test is skipped, empty, or returns a no-op result.

---

## Required Changes for Finalization

1. **Bootstrap-Only Isolation:** Hard-gate CLI bootstrap so that bootstrap-only commands (`gc doctor`, `gc import-state`, `gc version`) are identified first and execute in an isolated environment bypass-mode, completely free of eager systempack or loader dependencies.
2. **Local Pin Override for CI:** Add a `--local-override` path mapping option to the pin-coherence and packcompat validators to support local cross-repo testing and break the circular SHA-pin deadlock during development.
3. **Execution-Verified Witnesses:** Require the coverage-transfer and behavior-preservation validators to inspect active test execution logs (`go test -json` transcripts) to guarantee that no mapped new witness is skipped or a no-op.

---

## Open Questions for Implementation

* **How will `gc doctor` display conflicting transitive imports?** It should print the full import paths (e.g., `City -> Pack A -> Pack C@sha1` vs `City -> Pack B -> Pack C@sha2`) so operators can easily copy-paste the required pin into `city.toml`.
* **Should `.gc_provenance.json` use SHA-256 for local cache integrity validation?** Yes, SHA-256 is recommended for consistency with existing remote-pack digests.
