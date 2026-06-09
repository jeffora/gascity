# Mira Acharya — DeepSeek V4 Flash (Independent Test Strategy & CI Review) — Iteration 7

**Verdict:** approve-with-risks

**Scope:** Invariant tests, absence scans, regression harnesses, and acceptance traceability. Judged against `gc.mayor.requirements.v1`.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding & Independence:** This review represents an independent DeepSeek V4 Flash evaluation of Iteration 7. I have re-grounded each finding against the live codebase, the updated requirements (`plans/core-gastown-pack-migration/requirements.md`, 135 lines, status updated 2026-06-09), and the proposed implementation plan (`plans/core-gastown-pack-migration/implementation-plan.md`, 835 lines).
2. **Dual-Placement Strategy:** Due to a known workflow defect where the bead's metadata `gc.attempt=1` causes automated tools to write to `attempt-1/reviews/` and block attempt-local synthesis, I am writing this complete review to **both** the literal path `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/test-strategy-ci-reviewer_gemini.md` (to satisfy the automated bead verification) and the active iteration path `.gc/design-reviews/ga-dtvdnd/attempt-7/reviews/test-strategy-ci-reviewer_gemini.md` (to ensure synthesis correctness).

---

## Schema Conformance (requested)

Conforms — valid front matter, the six sections in required order, W6H complete, and Example Mapping covers happy/negative/edge. No in-lane schema concern (deep conformance is the schema officer's lane).

---

## Grounding (why this lane bites)

I confirmed the legacy tests this migration must retire actually exist and are behavior-level, not hypothetical:
`cmd/gc/embed_builtin_packs_test.go:1321` comments "Core and maintenance are **always auto-included**" and asserts `builtinPackIncludes()` returns `maintenance` at `includes[1]` (also `:1255-1299` for the +bd case); plus `examples/gastown/maintenance_scripts_test.go` and materialized-order assertions (`:165-167`, `:700-807`). Conversely, I found **no existing automated role-neutrality denied-token scan** over source/generated output — files that contain `Mayor`/`Deacon`/`Polecat` use them as fixtures, and the tests that walk source (`worker_boundary_import_test.go`, `gc_beads_bd_lint_test.go`, `gitignore_test.go`) are not role scans. So red flag #1 is presently *live* in the tree, and AC8 is net-new invariant infrastructure.

---

## Top Strengths of Iteration 7

- **Frozen Baseline for Completeness Verification:** AC6 and AC13 have been updated to validate against a frozen historical reference snapshot or baseline Git commit rather than a live workspace walk. This directly mitigates the "CI Generator Freshness Trap" where deleting legacy files would cause subsequent sweeps to silently report 100% success on an empty set.
- **Robust Two-Repository Slice Sequence:** The seven-slice rollout sequence cleanly isolates prerequisite proof generation from Gas City side-effecting code deletions. It enforces that `behavior-preservation.yaml` and `public-gastown-pins.yaml` are verified external prerequisites before local Maintenance code can be retired.
- **Centralized Central-Coordinator Architecture:** Replacing legacy direct mutations with a centralized, lock-protected mutation coordinator (`internal/doctorfix`) ensures atomic preflight, post-commit checks, and deterministic recovery.

---

## Evaluation of the Three Key Questions

### Q1: Does each acceptance criterion map to a concrete unit, integration, command, or absence-test proof?

**Yes.** AC17 successfully mandates a complete acceptance-to-proof matrix. The implementation plan maps AC1 through AC16 to specific unit/integration test pathways and enforces strict CI gate placements:

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

**Yes, but with critical scanner guardrails required.**
The implementation plan's "Testing" section correctly references scanner tests to reject active Core role-name references (lines 624–627). To ensure this scanner is robust:
- **No Wildcard/Broad Exclusions:** The scanner must not accept broad directory-level ignores (such as ignoring the entire `test/` directory).
- **Line-Gated Whitelist:** Documented migration exceptions, positive control files, or historical/test files must be explicitly whitelisted in `absence_scan_test.go` by filename and precise line ranges (e.g., `REQUIREMENTS.md#L45`). Any occurrence of a forbidden word outside this explicit whitelist must immediately cause the test to fail.
- **Canine Prompt Neutrality Proof:** To verify that the default `dog` executor remains a configurable default rather than a Go-side role assumption, the test suite must execute a **rendered template assertion**. The absence scan must prove that Core's prompt templates contain no hardcoded "Dog" or canine personas (referring to the worker strictly dynamically via `{{.Bindings.maintenance_worker}}`), and must include a negative control test that renders the prompt with a non-`dog` binding and asserts that no canine terms appear in the rendered prompt.
- **AST-Aware Scanning:** The scanner must use AST-based token scanning for Go source files to isolate scans to string literals and comments, preventing false-positive matches on common prose words (e.g., matching "boot" inside "bootstrap") while maintaining absolute safety.

---

### Q3: Are legacy Maintenance auto-inclusion tests rewritten to cover required Core plus explicit external Gastown imports?

**Yes, the requirement and the design are aligned here.**
AC13 (coverage-transfer table) and the implementation plan (lines 641–643) specify rewriting/replacing legacy tests (e.g., `examples/gastown/gastown_test.go` and `examples/gastown/maintenance_scripts_test.go`).
To guarantee that physical deletions do not defeat the completeness checks, the validator must fail closed on empty post-deletion walks or unmapped assertions.

---

## Critical Risks & Recommended Changes

### 1. [Major Risk] The Skipped/No-Op Witness Blindspot (The Dummy Test Escape)
- **The Risk:** AC13 mandates that the `coverage-transfer.yaml` validator must fail closed on empty post-deletion walks, unmapped assertions, or behavior witnesses that only prove a no-op. However, if the validator only checks that a new witness test exists on the filesystem or matches a naming pattern, developers under pressure could map retired assertions to new tests that are skipped (`t.Skip()`), empty, or contain only dummy assertions. This would bypass the behavior-preservation guarantees.
- **Required Change:** The coverage-transfer and behavior-preservation validators must analyze active Go test execution transcripts (e.g., parsing the JSON stream of `go test -json`) during validation runs. They must verify that every mapped new witness is actively executed and passes, explicitly failing closed if any witness test is skipped, empty, or returns a no-op result.

### 2. [Major Risk] Multi-Repository Circular Pin-Lock (The SHA-Pin Deadlock)
- **The Risk:** The rollout policy (Slices 1b/1c/2/5a) defines a multi-repository promotion flow where `gascity` pins `gascity-packs/gastown` to an immutable `sha:`. However, the public Gastown activation commit cannot be fully tested in public CI without running against the final `gascity` binary that implements the Core split. This creates a circular dependency across the repositories: `gascity` cannot adopt the final pin until `gascity-packs` is committed, and `gascity-packs` cannot commit/verify the final activation mode until `gascity` is merged.
- **Required Change:** The pin-coherence validator and the `test/packcompat` harness must support a `--local-override` flag or path-level mapping during development and multi-repository CI pipeline execution. This allows validating the candidate `gascity-packs` commit locally in tandem with the local `gascity` workspace before the SHA is finalized on origin, breaking the circular dependency.

### 3. [Major Risk] The AC2 Dev/Test Escape Hatch Silent Paradox (The Production Safety Gap)
- **The Risk:** `requirements.md` AC2 introduces a "clear dev/test escape hatch if tests need to construct partial configs" (which lack Core). However, the implementation plan (`implementation-plan.md`) never mentions this escape hatch. If this escape hatch can be triggered in production (e.g., via CLI flags or naive env vars), operators could run Core-less or partially-loaded cities, violating SDK self-sufficiency and creating silent security and functional holes.
- **Required Change:** The dev/test escape hatch must be hard-coded to activate *only* within native Go test environments (e.g., verifying `flag.Lookup("test.v") != nil` or `flag.Lookup("test.run") != nil`), strictly preventing production/CLI activation via command-line flags or environment variables.

### 4. [Major Risk] Eager Loader Dependency in Bootstrap-Only Mode (The Doctor Boot-Loop)
- **The Risk:** AC11 and the implementation plan specify that `gc doctor` and repair commands have a bootstrap-only diagnostic mode that can run even when normal pack resolution is completely broken (e.g., when Core is missing or corrupted). However, if the CLI bootstrap process or any imported sub-package eagerly initializes the normal config loader or triggers required-pack validation before identifying that the command being run is a bootstrap-only repair command, the CLI will crash or exit with a fatal error. This results in a boot loop, completely blocking the operator from running `gc doctor --fix` to repair the broken state.
- **Required Change:** The CLI bootstrap path must strictly detect bootstrap-only commands (`gc doctor`, `gc import-state`, `gc version`) at the very entry point, before any config resolution, required-pack materialization, or Go-side initialization of system-packs occurs. These commands must execute in a completely dependency-isolated path that bypasses Gate 1 and Gate 2 validation.

---

## Missing Evidence

1. **No Active Test Run Verification Design:** No design is present showing how validators parse the output of `go test -json` to prove that new witnesses are actually running and passing rather than being skipped or dummy placeholders.
2. **No Diagnostic Tracing of Transitive Imports:** No golden-test output exists demonstrating how a nested, transitive legacy import (e.g., `City -> Pack A -> legacy/packs/maintenance`) is traced back to its root config layer.
3. **No Automatic Database/Config Backup on Repair:** Although the coordinator uses an advisory lock and supports idempotency, there is no automated backup step (`city.toml.bak` or `beads.db.bak`) before mutation writes occur.

---

## Required Changes for Finalization

1. **Automated Test Log Verification:** Require the coverage-transfer and behavior-preservation validators to inspect active test execution logs (`go test -json` transcripts) to guarantee that no mapped new witness is skipped, empty, or a no-op.
2. **Local Pin Overrides for Circular Dependencies:** Add a `--local-override` path mapping option to the pin-coherence and packcompat validators to support local cross-repo testing and break the circular SHA-pin deadlock during development.
3. **Hard-Gated Escape Hatch:** Restrict the dev/test Core-less escape hatch solely to native Go test environments (`flag.Lookup("test.v") != nil` or `flag.Lookup("test.run") != nil`).
4. **Bootstrap-Only Isolation:** Hard-gate CLI bootstrap so that bootstrap-only commands (`gc doctor`, `gc import-state`, `gc version`) are identified first and execute in an isolated environment bypass-mode, completely free of eager systempack or loader dependencies.

---

## Questions

- **Execution-Verified Witnesses Overhead:** Will parsing Go test JSON logs in the validators significantly increase CI runtimes, or can it be sharded with the existing `make test-fast-parallel` targets?
- **Offline Cache Pre-Seeding:** Is there a plan to provide a simple `gc cache seed` or `gc pack import` command to facilitate air-gapped migrations for offline operators?
- **Rollback Database Restore:** Should the mutation coordinator automatically restore `.gc/beads.db.bak` and `.gc/backup/city.toml.bak` when an explicit recovery/downgrade run is requested?
