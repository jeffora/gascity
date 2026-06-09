# Mira Acharya — DeepSeek V4 Flash (Independent Test Strategy & CI Review) — Iteration 9 / Attempt 9

**Lane:** test-strategy-ci-reviewer (wave 1) — invariant tests, absence scans, regression harnesses, and acceptance traceability.

**Verdict:** approve-with-risks

---

## Lane & Context Note (Dual-Placement Strategy)

1. **Re-Grounding & Independence:** This review represents an independent DeepSeek V4 Flash evaluation of Iteration 9. I have re-grounded each finding against the live codebase, the updated requirements (`plans/core-gastown-pack-migration/requirements.md` / `.gc/design-reviews/ga-dtvdnd/attempt-9/design-before.md`, 149 lines), the `gc.mayor.requirements.v1` schema, and the proposed implementation plan (`plans/core-gastown-pack-migration/implementation-plan.md`, 835 lines, updated 2026-06-09).
2. **Dual-Placement Strategy:** Due to the known workflow defect where the bead's metadata `gc.attempt=1` forces automated tools to write to `attempt-1/reviews/` and blocks active synthesis in the current iteration (`attempt-9/`), I am writing this complete review to **both** the literal path `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/test-strategy-ci-reviewer_gemini.md` (to satisfy the automated bead contract) and the active iteration path `.gc/design-reviews/ga-dtvdnd/attempt-9/reviews/test-strategy-ci-reviewer_gemini.md` (to ensure synthesis correctness).

---

## Executive Summary

The Iteration 9 requirements and implementation plan establish an exceptionally rigorous, evidence-based roadmap for the Core/Gastown pack split. By elevating the various support files (the version-skew matrix, the pack-resolution matrix, and the coverage-transfer ledger) into **machine-validated, binding evidence contracts** (AC17, line 722), the plan ensures that correctness does not rely on manual vigilance. This perfectly aligns with the Zero Framework Cognition (ZFC) principles and the Bitter Lesson.

I award an **approve-with-risks** verdict. The test strategy and validation gates are comprehensive, but we must enforce strict technical pins to guarantee that development-time edge cases, circular SHA-pins, and test escape hatches do not compromise production safety or boot-loop the doctor utility.

---

## Evaluation of the Three Key Questions

### Q1: Does each acceptance criterion map to a concrete unit, integration, command, or absence-test proof?

**Yes.** The implementation plan includes a complete AC17 acceptance-to-proof matrix (lines 704–723) that maps every single acceptance criterion to an executable test pathway or validated support artifact:

| ID | Focus | Verification Mode | Target Test Class / Path | CI / Gate Placement |
| --- | --- | --- | --- | --- |
| **AC1** | Schema Compliance | Static schema validation | `cmd/gc/doctor_test.go` | Pre-commit & Fast Unit (`make test`) |
| **AC2** | Core loading & self-sufficiency | Core required-pack tests | `internal/systempacks/load_test.go` | Fast Unit (`make test`) |
| **AC3** | Precedence & resolution matrix | Precedence test suite | `internal/config/resolution_test.go` | Fast Unit (`make test`) |
| **AC4** | Fresh template init with remotes | Integration template test | `test/integration/init_template_test.go` | Integration Shard (`make test-integration-shards`) |
| **AC5** | Maintenance retirement | Auto-include/bundle checks | `internal/builtinpacks/registry_test.go` | Fast Unit (`make test`) |
| **AC6** | Ledger validation check | Ledger schema & source audit | `cmd/gc/ledger_test.go` | Pre-commit & Fast Unit (`make test`) |
| **AC7** | Behavior preservation harness | Packcompat validation | `test/packcompat/preservation_test.go` | Integration Shard (`make test-integration-shards`) |
| **AC8** | Role-neutrality absence scan | Static & materialized scans | `test/packlint/absence_scan_test.go` | Pre-commit & Fast Unit (`make test`) |
| **AC9** | Configurable maintenance executor | Binding precedence & overrides | `internal/config/bindings_test.go` | Fast Unit (`make test`) |
| **AC10** | Existing-city upgrade matrix | Live upgrade-matrix test | `test/integration/upgrade_matrix_test.go` | Integration Shard (`make test-integration-shards`) |
| **AC11** | Text/JSON diagnostic attribution | Structured output checks | `cmd/gc/doctor_test.go` | Fast Unit (`make test`) |
| **AC12** | Documentation/help text grep | Terminology grep audit | `test/packlint/docs_audit_test.go` | Fast Unit (`make test`) |
| **AC13** | Coverage-transfer validator | Retired-test completeness | `test/packlint/coverage_transfer_test.go`| Fast Unit (`make test`) |
| **AC14** | External Gastown remote check | Local Git remote fixture | `test/integration/remote_gastown_test.go` | Release-Gate Only (manual/package build) |
| **AC15** | Version-skew & pin policy | Coherence and mismatch tests | `test/packcompat/version_skew_test.go` | Fast Unit (`make test`) |
| **AC16** | Offline & cache behavior | Network-disabled promotions | `test/integration/offline_cache_test.go` | Integration Shard (`make test-integration-shards`) |

---

### Q2: Are tests added or rewritten to fail if Core reintroduces Gastown role names while allowing documented migration exceptions and configurable dog defaults?

**Yes, but with several scanner guardrails required to prevent evasion:**
The implementation plan specifies a wording scanner (AC8) and static absence scans to reject Gastown role names and literal `dog` routing (lines 410–431). To make this scanner watertight against "evasion" or false positives:
- **AST-Aware Scanning:** The Go-source absence scan must parse source files using the `go/ast` and `go/token` packages. It must specifically target Go string literals, comments, and identifiers. This prevents false-positive hits on common prose terms (such as matching the word "boot" inside the valid identifier "bootstrap") while ensuring that actual role assignments fail the build.
- **Line-Gated Allowlists:** The scanner's allowlist must be strictly line-gated and file-gated (e.g., `REQUIREMENTS.md#L45`). General directory-level or file-wide ignores are forbidden. If a forbidden word is added to a file, the scan must fail unless that exact line is whitelisted.
- **Canine Prompt Neutrality Proof:** To prove that the default `dog` executor remains a configurable default rather than a Go-side role assumption, the test suite must execute a **rendered template assertion**. The absence scan must prove that Core's prompt templates contain no hardcoded "Dog" or canine personas (referring to the worker strictly dynamically via `{{.Bindings.maintenance_worker}}`), and must include a negative control test that renders the prompt with a non-`dog` binding and asserts that no canine terms appear in the rendered prompt.

---

### Q3: Are legacy Maintenance auto-inclusion tests rewritten to cover required Core plus explicit external Gastown imports?

**Yes.** AC13 specifies a `coverage-transfer.yaml` ledger to map all legacy assertions (such as `examples/gastown/gastown_test.go` and `examples/gastown/maintenance_scripts_test.go`) to active replacement coverage (lines 641–643).
- To guarantee that physical deletions do not defeat the completeness checks, the validator must fail closed on empty post-deletion walks or unmapped assertions.

---

## Deep-Dive Analysis: Cross-Document Consistency & Missing Edge Cases

From an independent, evidence-based testing perspective, I have identified six critical risks and cross-document inconsistencies:

### 1. [Major Risk] The Skipped/No-Op Witness Blindspot (The Dummy Test Escape)
*   **The Risk:** AC13 mandates that the `coverage-transfer.yaml` validator must fail closed on empty post-deletion walks, unmapped assertions, or behavior witnesses that only prove a no-op. However, if the validator only checks that a new witness test exists on the filesystem or matches a naming pattern, developers under pressure could map retired assertions to new tests that are skipped (`t.Skip()`), empty, or contain only dummy assertions. This would bypass the behavior-preservation guarantees.
*   **Required Technical Pin:** The coverage-transfer and behavior-preservation validators must analyze active Go test execution transcripts (e.g., parsing the JSON stream of `go test -json`) during validation runs. They must verify that every mapped new witness is actively executed and passes, explicitly failing closed if any witness test is skipped, empty, or returns a no-op result.

### 2. [Major Risk] Multi-Repository Circular Pin-Lock (The SHA-Pin Deadlock)
*   **The Risk:** The rollout policy (Slices 1b/1c/2/5a) defines a multi-repository promotion flow where `gascity` pins `gascity-packs/gastown` to an immutable `sha:`. However, the public Gastown activation commit cannot be fully tested in public CI without running against the final `gascity` binary that implements the Core split. This creates a circular dependency across the repositories: `gascity` cannot adopt the final pin until `gascity-packs` is committed, and `gascity-packs` cannot commit/verify the final activation mode until `gascity` is merged.
*   **Required Technical Pin:** The pin-coherence validator and the `test/packcompat` harness must support a `--local-override` flag or path-level mapping during development and multi-repository CI pipeline execution. This allows validating the candidate `gascity-packs` commit locally in tandem with the local `gascity` workspace before the SHA is finalized on origin, breaking the circular dependency.

### 3. [Major Risk] The AC2 Dev/Test Escape Hatch Silent Paradox (The Production Safety Gap)
*   **The Risk:** `requirements.md` AC2 introduces a "clear dev/test escape hatch if tests need to construct partial configs" (which lack Core). However, the implementation plan (`implementation-plan.md`) never mentions this escape hatch. If this escape hatch can be triggered in production (e.g., via CLI flags or naive env vars), operators could run Core-less or partially-loaded cities, violating SDK self-sufficiency and creating silent security and functional holes.
*   **Required Technical Pin:** The dev/test escape hatch must be hard-coded to activate *only* within native Go test environments (e.g., verifying `flag.Lookup("test.v") != nil` or `flag.Lookup("test.run") != nil`), strictly preventing production/CLI activation via command-line flags or environment variables.

### 4. [Major Risk] Eager Loader Dependency in Bootstrap-Only Mode (The Doctor Boot-Loop)
*   **The Risk:** AC11 and the implementation plan specify that `gc doctor` and repair commands have a bootstrap-only diagnostic mode that can run even when normal pack resolution is completely broken (e.g., when Core is missing or corrupted). However, if the CLI bootstrap process or any imported sub-package eagerly initializes the normal config loader or triggers required-pack validation before identifying that the command being run is a bootstrap-only repair command, the CLI will crash or exit with a fatal error. This results in a boot loop, completely blocking the operator from running `gc doctor --fix` to repair the broken state.
*   **Required Technical Pin:** The CLI bootstrap path must strictly detect bootstrap-only commands (`gc doctor`, `gc import-state`, `gc version`) at the very entry point, before any config resolution, required-pack materialization, or Go-side initialization of system-packs occurs. These commands must execute in a completely dependency-isolated path that bypasses Gate 1 and Gate 2 validation.

### 5. [Minor Risk] Tmux Clean-up Scan Defect
*   **The Risk:** The Implementation Plan (lines 456-457) correctly specifies that "Tmux cleanup examples and tests must target isolated sockets only and must never use a default-server kill." However, without an automated static scan to enforce this, future development could easily re-introduce raw `tmux kill-server` or `tmux kill-session` calls into scripts or tests.
*   **Required Technical Pin:** Add `tmux kill-server` and `tmux kill-session` (without explicit `-L` socket identifiers) to the denied-token set in the wording/absence scanners, making raw tmux cleanups a pre-commit static failure.

### 6. [Minor Risk] Interrupted doctor mutations and recovery verification
*   **The Risk:** Slices 4b/4c implement `internal/doctorfix` and automatic doctor repairs. While the plan mentions that failure injection after each staged publish step reruns or rolls back deterministically (lines 663-664), it lacks a concrete test that proves recovery after an *unclean process exit* (such as killing the test binary mid-write and verifying that next startup is repairable/idempotent).
*   **Required Technical Pin:** Require the doctor upgrade-matrix tests to include a crash-recovery simulation where a mutating doctor process is forcibly terminated (e.g. via `SIGKILL` or a test-level exit) mid-mutation, verifying that the lock remains intact, the journal captures the incomplete transaction, and a subsequent non-interactive doctor run can recover the city cleanly.

---

## Required Technical Guardrails for Slices

1. **Automated Test Log Verification:** Require the coverage-transfer and behavior-preservation validators to inspect active test execution logs (`go test -json` transcripts) to guarantee that no mapped new witness is skipped, empty, or a no-op.
2. **Local Pin Overrides for Circular Dependencies:** Add a `--local-override` path mapping option to the pin-coherence and packcompat validators to support local cross-repo testing and break the circular SHA-pin deadlock during development.
3. **Hard-Gated Escape Hatch:** Restrict the dev/test Core-less escape hatch solely to native Go test environments (`flag.Lookup("test.v") != nil` or `flag.Lookup("test.run") != nil`).
4. **Bootstrap-Only Isolation:** Hard-gate CLI bootstrap so that bootstrap-only commands (`gc doctor`, `gc import-state`, `gc version`) are identified first and execute in an isolated environment bypass-mode, completely free of eager systempack or loader dependencies.
5. **Static Scan for Tmux Safety:** Enforce the tmux isolation requirement statically by registering raw `tmux kill-server` (lacking socket parameters) as a denied token in the static absence scan.
6. **Crash-Recovery Verification:** Mandate a crash-recovery simulation test for `internal/doctorfix` where mutating repair is killed mid-write and recovery/idempotency is verified.

---

## Questions

- **Execution-Verified Witnesses Overhead:** Will parsing Go test JSON logs in the validators significantly increase CI runtimes, or can it be sharded with the existing `make test-fast-parallel` targets?
- **Offline Cache Pre-Seeding:** Is there a plan to provide a simple `gc cache seed` or `gc pack import` command to facilitate air-gapped migrations for offline operators?
- **Rollback Database Restore:** Should the mutation coordinator automatically restore `.gc/beads.db.bak` and `.gc/backup/city.toml.bak` when an explicit recovery/downgrade run is requested?

---

## Verdict & Transition to Implementation

**Verdict: approve-with-risks**

The Requirements Document and Implementation Plan are fully approved to transition to the **design and implementation-plan** phases once the critical technical pins specified above are incorporated into the appropriate development slices.
