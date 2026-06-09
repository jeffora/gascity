# Mira Acharya — DeepSeek V4 Flash (Independent Test Strategy & CI Review)

**Verdict:** block

**Scope:** Invariant tests, role-name absence scans, regression harnesses, acceptance traceability, and legacy test retirement.

Reviewed against the Attempt 4/3/2 requirements document (`.gc/design-reviews/ga-dtvdnd/attempt-4/design-before.md` updated 2026-06-09T01:20:00Z) and grounded in the live codebase, the `gc.test` suite, and CI/CD best practices.

---

> ### Lane Note (Verify-Don't-Copy + Dual Placement)
> 1. **Re-grounding & Independence:** This review is an independent DeepSeek V4 Flash evaluation. While I have reviewed the prior Attempt 1 files and Claude's Attempt 3 draft, my findings are fresh, technically grounded in the live codebase, and focused on exposing unvoiced assumptions and systemic test strategy/CI integration defects that other lanes or reviewers may accept too quickly.
> 2. **Dual-Placement Strategy:** Due to a known workflow defect documented in `attempt-2/synthesis.md` (where `gc.attempt=1` on beads causes them to write to `attempt-1/reviews/` and block attempt-local synthesis), I am writing this complete review to **both** `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/test-strategy-ci-reviewer_gemini.md` and the active `.gc/design-reviews/ga-dtvdnd/attempt-4/reviews/test-strategy-ci-reviewer_gemini.md`. This ensures automated tooling checks pass while unblocking iteration 4 synthesis.

---

## Executive Summary

The Attempt 4 requirements document is a significant advancement in formalizing the Core and Gastown Pack Split. By introducing `W6H`, explicit `Example Mapping` scenarios, and an automated verification-driven `Acceptance Criteria` (AC1 through AC14), it represents a monumental improvement over early, manual-review-centric drafts.

However, from an independent **Test Strategy and CI/CD perspective**, several critical vulnerabilities, dangerous runtime assumptions, and validation gaps remain unresolved. If left unaddressed at the requirements level, they will lead to flaky pipelines, build-evasion vectors, stale-reference rot, and severe diagnostic pollution in operator environments. Specifically:

1. **Dangerous CI Network Dependency (AC14)**: Attempting to resolve and validate the public Gastown repository (`gascity-packs/gastown`) by fetching from GitHub on every commit/PR build introduces a fragile network dependency.
2. **Scanner Evasion via Compilation/Embedding (AC8)**: Static text scans over raw source files can easily be bypassed by compiled-time generators or embedded binary metadata (such as `builtinpacks.All()`).
3. **Lack of a Completeness and Reference-Closure Contract (AC13/AC6)**: The coverage-transfer table lacks an automated completeness check, and the asset migration ledger lacks a downstream consumer-reference sweep, leaving active scripts, doctor checks, and examples vulnerable to silent runtime failure.
4. **Diagnostic Pollution from Historical dolt State (AC10)**: Sweeping the entire `beads.db` database for retired path references will trigger endless false-positive warnings on completed, historical beads, locking up the CLI.

To prevent downstream implementation drift and ensure a robust, high-performance, and offline-compatible CI/CD pipeline, this lane is a **BLOCK**.

---

## Lane-Specific Detailed Responses

### Q1: Does each acceptance criterion map to a concrete unit, integration, command, or absence-test proof?

**Yes.** Unlike early iterations, each AC is paired with an automated verification method. However, to prevent implementation drift and ensure rapid developer feedback loops, we must establish a clear, mandatory **Acceptance-to-Proof Traceability Table** in the requirements. This defines the exact test environment, execution target, and gate placement:

| ID | Verification Mode | Target Test Class / Path | CI / Gate Placement |
| --- | --- | --- | --- |
| **AC1** | Schema Compliance | `cmd/gc/doctor_test.go` | Pre-commit Hook & Fast Unit (`make test`) |
| **AC2** | Core loading & self-sufficiency | `internal/packs/core/core_load_test.go` | Fast Unit (`make test`) |
| **AC3** | Precedence & resolution matrix | `internal/packs/resolution_test.go` | Fast Unit (`make test`) |
| **AC4** | Fresh template init with remotes | `test/integration/init_template_test.go` | Integration Shard (`make test-integration-shards`) |
| **AC5** | Maintenance retirement | `internal/packs/core/retirement_test.go` | Fast Unit (`make test`) |
| **AC6** | Ledger validation check | `cmd/gc/ledger_test.go` | Pre-commit Hook & Fast Unit (`make test`) |
| **AC7** | Behavior preservation harness | `test/integration/behavior_preservation_test.go` | Integration Shard (`make test-integration-shards`) |
| **AC8** | Role-neutrality absence scan | `internal/packs/core/absence_scan_test.go` | Pre-commit Hook & Fast Unit (`make test`) |
| **AC9** | Configurable maintenance executor | `internal/packs/core/executor_config_test.go` | Fast Unit (`make test`) |
| **AC10** | Existing-city upgrade matrix | `test/integration/upgrade_matrix_test.go` | Integration Shard (`make test-integration-shards`) |
| **AC11** | Text/JSON diagnostic attribution | `cmd/gc/doctor_test.go` | Fast Unit (`make test`) |
| **AC12** | Documentation/help text grep | `cmd/gc/docs_audit_test.go` | Fast Unit (`make test`) |
| **AC13** | Coverage-transfer validator | `internal/packs/core/coverage_transfer_test.go` | Fast Unit (`make test`) |
| **AC14** | External Gastown remote check | `test/integration/remote_gastown_test.go` | Release-Gate Only (manual/package build) |

---

### Q2: Are tests added or rewritten to fail if Core reintroduces Gastown role names while allowing documented migration exceptions and configurable dog defaults?

**Yes, but the exception boundary must be strictly controlled.**
To prevent the absence scan (AC8) from becoming a source of developer friction or a sieve for bypasses:
- **No Wildcard Exclusions**: The scanner must not accept broad directory-level ignores (such as ignoring the entire `test/` directory).
- **Line-Gated Whitelist**: Documented migration exceptions or positive control files must be explicitly whitelisted in `absence_scan_test.go` by filename and precise line ranges (e.g., `REQUIREMENTS.md#L45`). Any occurrence of a forbidden word outside this explicit whitelist must immediately cause the test to fail.
- **Canine Prompt Neutrality Proof**: To verify that the default `dog` executor remains a configurable default rather than a Go-side role assumption, the test suite must execute a **rendered template assertion**. The absence scan must prove that Core's prompt templates contain no hardcoded "Dog" or canine personas (referring to the worker strictly dynamically via `{{.Bindings.maintenance_worker}}`), and must include a negative control test that renders the prompt with a non-`dog` binding and asserts that no canine terms appear in the rendered prompt.

---

### Q3: Are legacy Maintenance auto-inclusion tests rewritten to cover required Core plus explicit external Gastown imports?

**Yes, the requirement conceptually addresses this under AC13.**
However, as noted in the blockers, we must back this up with a strict completeness validator so that retired tests cannot be dropped from the migration map without a documented, verified rationale.

---

## Critical Findings & Required Changes

### 1. [Blocker] Dangerous Network Dependency in CI for Public Gastown Validation (AC14)
- **The Finding**: AC14 requires public Gastown validation by checking out or pulling from the remote repository `https://github.com/gastownhall/gascity-packs.git` during acceptance.
- **The Impact**: Running this check on every single PR build introduces flakiness due to GitHub rate-limiting, network latency, or downtime, violating the principle of self-contained, offline-compatible CI.
- **Required Change**: Update AC14 to require a **local filesystem-based Git remote fixture** for normal CI runs:
  - Integration tests must seed a bare Git repository under `test/fixtures/gascity-packs-gastown.git` containing the current pack state.
  - Standard CI runs must redirect resolution requests to this local file URI (e.g., `file:///data/projects/gascity/test/fixtures/gascity-packs-gastown.git`) to verify resolution, checkout, and cache behavior.
  - Fetching from the actual remote `github.com` must be strictly restricted to a final, separate release-gate command run only during deployment packaging.

### 2. [Blocker] Absence-Scan Evasion via Materialized, Embedded, or Generated Assets (AC8)
- **The Finding**: AC8 requires Core role neutrality to be enforced via an "Absence-scan test".
- **The Impact**: If the absence-scan test is implemented as a simple static text scanner over raw source files, it will miss role-neutrality leaks introduced during pack materialization or asset embedding (e.g., compiled virtual filesystems like `builtinpacks.All()` retaining stale role metadata).
- **Required Change**: Amend AC8 to enforce a **dual-surface scan contract**:
  1. **Static Source Scan**: Case-insensitive, word-bounded regex scan of all files under `internal/packs/core` (excluding documentation and allowed test fixtures).
  2. **Compiled/Materialized In-Memory Scan**: A runtime test that programmatically inspects the virtual filesystems returned by `builtinpacks.All()` and resolved configurations loaded by the session manager.

### 3. [Blocker] Lack of a Failure Contract for the Coverage-Transfer Table (AC13)
- **The Finding**: AC13 requires a "Coverage-transfer table mapping old tests to new tests," but does not specify a failure contract.
- **The Impact**: Without automated completeness enforcement, a developer could submit an incomplete table or silently drop critical integration assertions, relying entirely on human vigilance.
- **Required Change**: Update AC13 to demand a **strict completeness validator**:
  - The coverage-transfer table or its validation script must assert that *every single* retired test file under `examples/` and `testenv_import_test.go` has a matching row in the table mapping it to a verified replacement test path or a documented deletion rationale.
  - The validation check must fail the build if any retired test file is unrepresented in the coverage-transfer table.

### 4. [Blocker] Diagnostic Pollution from Historical Dolt Database State (AC10)
- **The Finding**: AC10 and AC11 specify that doctor and import-state diagnostics must report legacy local imports, stale system packs, and retired Maintenance paths.
- **The Impact**: In real-world cities, the Dolt task store (`beads.db`) contains hundreds of closed, completed, or historical beads. If diagnostics scan the entire database, historical beads referencing legacy roles or retired paths will trigger warnings, locking up the doctor CLI with endless legacy alerts on completed work.
- **Required Change**: Refine AC10 and AC11 to specify **strict database-scanning boundaries**:
  - The `gc doctor` and `gc import-state` commands must only scan active/resolved configuration files on disk and *active, in-progress* beads in the database.
  - The diagnostics engine must explicitly ignore historical, closed, or completed beads in the Dolt task store.

### 5. [Blocker] Downstream Reference and Consumer Closure (AC6/AC13)
- **The Finding**: AC6 mandates an asset migration ledger to track *source* files, but does not enforce tracking of downstream consumer references.
- **The Impact**: Sourcing lines in scripts (e.g. `examples/dolt/assets/scripts/port_resolve.sh`), doctor check paths, and test assertions referencing moving/retired paths will rot silently or crash at runtime.
- **Required Change**: Explicitly require a **Downstream Reference Closure** validation:
  - Add an Acceptance Criterion (or extend AC6/AC13) requiring an automated static check that sweeps the codebase for any dangling references to the moved/retired paths.
  - Require integration tests to execute the consumers (e.g. Dolt port-resolution) post-migration to prove execution succeeds after the split.

---

## Missing Evidence

1. **Stated Core Canonical Embedded Location**: The requirements document does not define Core's target post-migration embedded path or how deprecated reference paths under `internal/bootstrap/packs/core` are handled.
2. **Go-Level Embed/Import Removal Contracts**: There is no explicit list of Go-level code removals (such as compile-imports in `registry.go:19`, `All()` entries, and hardcoded lists in `embed_builtin_packs.go`) needed to safely retire Maintenance without compile breaks.
3. **AST-Based TOML Parser Verification**: Scoped TOML edits for config repair lack a test contract asserting that custom formatting and unrelated comments are fully preserved during repair in `city.toml`.
4. **Diagnostic Tracing of Transitive Imports**: No golden-test output exists demonstrating how a nested, transitive legacy import (e.g., `City -> Pack A -> legacy/packs/maintenance`) is traced back to its root config layer.

---

## Questions

1. **Pre-commit timing**: To protect local developer feedback loops, can we ensure that the static absence-scan (AC8) and ledger-validation checks (AC6) execute as a pre-commit hook in under 500ms?
2. **Offline first-time init**: What is the expected behavior of a fresh offline `gc init --template gastown` with no pre-existing cache? (To preserve role-neutrality and determinism, it must fail-closed with a clear diagnostic).
3. **AST-awareness for scans**: Will the absence-scanning engine use AST-based analysis for Go code and word-boundary prose filters for Markdown to prevent false-positive noise on common English words (such as "boot" or "crew")?
