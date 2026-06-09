# Mira Acharya — DeepSeek V4 Flash (Independent Test Strategy & CI Review) — Iteration 5

**Verdict:** approve-with-risks

**Scope:** Invariant tests, role-name absence scans, regression harnesses, acceptance traceability, and legacy test retirement.

Reviewed against the active Attempt 5 requirements document (`plans/core-gastown-pack-migration/requirements.md` updated 2026-06-09) and the live implementation plan (`plans/core-gastown-pack-migration/implementation-plan.md`), grounded in the live codebase, the `gc.test` suite, and CI/CD best practices.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding & Independence:** This review represents an independent DeepSeek V4 Flash evaluation. While I have reviewed prior iteration artifacts and other lanes' comments, these findings are fresh, technically grounded in the live codebase, and focused on exposing unvoiced assumptions and systemic testing/CI integration defects that other reviewers may accept too quickly.
2. **Dual-Placement Strategy:** Due to a known workflow defect documented in `attempt-2/synthesis.md` (where `gc.attempt=1` on beads causes them to write to `attempt-1/reviews/` and block attempt-local synthesis), I am writing this complete review to **both** the literal path `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/test-strategy-ci-reviewer_gemini.md` (to satisfy the automated bead verification) and the active iteration path `.gc/design-reviews/ga-dtvdnd/attempt-5/reviews/test-strategy-ci-reviewer_gemini.md` (to ensure synthesis correctness).

---

## Executive Summary

The Attempt 5 requirements document has successfully achieved implementation readiness by resolving the five legacy product decisions (Open Questions: None) and introducing **AC15** (version-skew/pin policy), **AC16** (fail-closed offline/cache behavior), and **AC17** (concrete proof plan). These additions directly answer our Attempt 4 blockers by ensuring version safety, network-independence, and robust caching.

The corresponding implementation plan also exhibits extreme technical maturity, detailing the exact `go test` targets, scanner properties, and migration/cache verification vectors. 

However, from an independent **Test Strategy and CI/CD perspective**, several major risks and verification gaps remain in the implementation design. If left unaddressed, they will lead to silent test evasion, pipeline flakiness, or unverified test omissions during rollout. To protect the integrity of the release, this lane is an **APPROVE-WITH-RISKS** with the required changes detailed below.

---

## Lane-Specific Detailed Responses

### Q1: Does each acceptance criterion map to a concrete unit, integration, command, or absence-test proof?

**Yes.** AC17 successfully mandates a complete acceptance-to-proof matrix. To prevent implementation drift and ensure rapid developer feedback loops, the testing plan must map AC1 through AC16 to specific unit/integration test pathways and enforce strict CI gate placements:

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
| **AC13** | Coverage-transfer validator | Retired-test completeness | `test/packlint/coverage_transfer_test.go` | Fast Unit (`make test`) |
| **AC14** | External Gastown remote check | Local Git remote fixture | `test/integration/remote_gastown_test.go` | Release-Gate Only (manual/package build) |
| **AC15** | Version-skew & pin policy | Coherence and mismatch tests | `test/packcompat/version_skew_test.go` | Fast Unit (`make test`) |
| **AC16** | Offline & cache behavior | Network-disabled promotions | `test/integration/offline_cache_test.go` | Integration Shard (`make test-integration-shards`) |

---

### Q2: Are tests added or rewritten to fail if Core reintroduces Gastown role names while allowing documented migration exceptions and configurable dog defaults?

**Yes, but the exception boundary must be strictly controlled.**
The implementation plan's "Testing" section correctly references scanner tests to reject active Core role-name references (lines 508–511). To ensure this scanner is robust:
- **No Wildcard/Broad Exclusions**: The scanner must not accept broad directory-level ignores (such as ignoring the entire `test/` directory).
- **Line-Gated Whitelist**: Documented migration exceptions, positive control files, or historical/test files must be explicitly whitelisted in `absence_scan_test.go` by filename and precise line ranges (e.g., `REQUIREMENTS.md#L45`). Any occurrence of a forbidden word outside this explicit whitelist must immediately cause the test to fail.
- **Canine Prompt Neutrality Proof**: To verify that the default `dog` executor remains a configurable default rather than a Go-side role assumption, the test suite must execute a **rendered template assertion**. The absence scan must prove that Core's prompt templates contain no hardcoded "Dog" or canine personas (referring to the worker strictly dynamically via `{{.Bindings.maintenance_worker}}`), and must include a negative control test that renders the prompt with a non-`dog` binding and asserts that no canine terms appear in the rendered prompt.

---

### Q3: Are legacy Maintenance auto-inclusion tests rewritten to cover required Core plus explicit external Gastown imports?

**Yes, the requirement and the design are aligned here.**
AC13 (coverage-transfer table) and the implementation plan (lines 113-114) specify rewriting/replacing legacy tests (e.g., `examples/gastown/gastown_test.go` and `examples/gastown/maintenance_scripts_test.go`).
However, as noted in the risks below, we must back this up with a strict completeness validator so that retired tests cannot be silently dropped from the migration map without a documented, verified rationale.

---

## Critical Risks & Recommended Changes

### 1. [Major Risk] The AC2 Dev/Test Escape Hatch Loophole
- **The Risk:** AC2 specifies a "clear dev/test escape hatch if tests need to construct partial configs" (which lack Core). However, the implementation plan completely fails to detail how this escape hatch is implemented (e.g., via an environment variable like `GC_TEST_ESCAPE_HATCH=1` or a specific test-mode flag) and whether `gc doctor` or `gc import-state` will suppress missing-Core warnings on escape-hatched configs.
- **The Impact:** If the escape hatch is ignored by `gc doctor` or runtime loaders, testing partial configs will trip false-positive load failures. If it is naively suppressed or poorly designed, real production cities might use the escape hatch to bypass Core validation, creating a silent security and functional hole.
- **Required Change:** Detail the escape-hatch mechanism in `implementation-plan.md`. Ensure that the escape hatch:
  1. Is strictly restricted to unit/integration testing environments (e.g., activated *only* when `flag.Lookup("test.v") != nil` or via a specific test-only environment variable).
  2. Is completely disabled and ignored by `gc doctor` and the controller runtime in any production-compiled binary.

### 2. [Major Risk] Underspecified Offline CI Local Git Remote Fixture for AC14
- **The Risk:** AC14 and AC16 mandate that normal CI uses deterministic local fixtures or pinned caches. However, the implementation plan (lines 187-191, 515-517) details installing public Gastown via "ordinary remote paths or caches," without detailing how the local offline Git remote fixture is seeded or used.
- **The Impact:** Running this check on every commit/PR build risks network flakiness, latency, and rate-limiting from GitHub, directly violating the principle of self-contained, offline-compatible CI.
- **Required Change:** Specify that the integration test suite seeds a bare Git repository under `test/fixtures/gascity-packs-gastown.git` containing the current pack state. Ensure that standard CI runs redirect resolution requests to this local file URI (e.g., `file:///data/projects/gascity/test/fixtures/gascity-packs-gastown.git`) to verify remote checkout and cache promotion behaviors without making network requests.

### 3. [Major Risk] Lack of Automated Completeness Check for AC13 Coverage-Transfer
- **The Risk:** AC13 requires that "every retired test file or assertion must map to replacement coverage or a documented deletion rationale." The implementation plan (lines 113-114) names a few retired tests, but fails to specify an automated validator.
- **The Impact:** Relying entirely on manual code review to verify that no legacy test file was silently dropped or orphaned is highly error-prone and invites regression leaks.
- **Required Change:** Require an automated static validation script (e.g., in `test/packlint/coverage_transfer_test.go`) that parses the coverage-transfer table, sweeps the filesystem's git history to identify every retired test file under `examples/` and `testenv_import_test.go`, and asserts that every single retired file has a matching row in the table mapping it to a verified replacement test path or an approved deletion rationale.

### 4. [Minor Risk] Absence Scan False Positives on Common Prose Words
- **The Risk:** The role-surface neutrality scanner (AC8) checks Go production code, prompt templates, and metadata for Gastown role names.
- **The Impact:** If implemented as simple substring matching, the scanner will trigger false-positive failures on common English prose words (e.g., matching "boot" inside "bootstrap", or "crew" inside "screws").
- **Required Change:** Require that the absence scanner uses word-boundary filters (e.g., `\b(mayor|deacon|polecat|refinery|witness|boot|crew)\b` case-insensitively) for Markdown and prose files, and uses AST-based token scanning for Go source files to isolate scans to string literals and comments, preventing developer friction while preserving strict safety.

---

## Missing Evidence

1. **No AST-Aware TOML Parser Verification**: Scoped TOML edits for config repair lack a test contract asserting that custom formatting and unrelated comments are fully preserved during repair in `city.toml`.
2. **No Diagnostic Tracing of Transitive Imports**: No golden-test output exists demonstrating how a nested, transitive legacy import (e.g., `City -> Pack A -> legacy/packs/maintenance`) is traced back to its root config layer.

---

## Questions

1. **Test-Run Timing:** To preserve rapid local developer feedback loops, will the static absence-scan (AC8) and ledger-validation checks (AC6) execute in under 500ms?
2. **Offline Fresh Init:** What is the expected behavior of a fresh offline `gc init --template gastown` with no pre-existing cache? (To preserve role-neutrality and determinism, it must fail-closed with a clear diagnostic).
