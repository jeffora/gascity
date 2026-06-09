# Mira Acharya — DeepSeek V4 Flash (Independent Test Strategy & CI Review)

**Verdict:** block

**Scope:** Invariant tests, role-name absence scans, regression harnesses, acceptance traceability, and legacy test retirement.

Reviewed against the Attempt 3 requirements document (`.gc/design-reviews/ga-dtvdnd/attempt-3/design-before.md` updated 2026-06-09T01:20:00Z) and grounded in the live codebase, the `gc.test` suite, and CI/CD best practices.

---

## Executive Summary

The Attempt 3 Requirements Document for the Core and Gastown Pack Split represents a massive, highly disciplined leap forward from Attempt 1. In the previous iterations, I blocked the requirements due to a complete absence of acceptance traceability, a total lack of Example Mapping, and a heavy reliance on human manual reviews instead of automated test suites to enforce Core neutrality and behavior preservation.

In Attempt 3, the authors have successfully maintained a structure fully compliant with the `gc.mayor.requirements.v1` schema. The structured `W6H`, `Example Mapping` with happy/negative/edge scenarios, and a dedicated `## Acceptance Criteria` section (AC1 through AC14) with concrete Verification columns represents an outstanding improvement.

However, from an independent **DeepSeek V4 Flash Test Strategy perspective**, several critical flaws, dangerous assumptions, and potential bypass routes remain unresolved. If left unaddressed at the requirements level, these gaps will lead to flaky CI builds, build evasion, and diagnostic pollution in real operator environments. Specifically:
1. **CI Network Dependency / Flakiness (AC14)**: Validating public Gastown repository imports by fetching from GitHub on every CI run introduces a fragile network dependency.
2. **Scanner Evasion via Pre-Compiled/Embedded Assets (AC8)**: Static code scanners can easily be bypassed by compilation-time generators or embedded binary metadata.
3. **Incomplete Legacy Coverage Transfer (AC13)**: The coverage-transfer table lacks a strict completeness failure contract, inviting silent coverage drops.
4. **Diagnostic Pollution from Historical Dolt State (AC10)**: Querying retired path references inside old, closed, or historical beads in `beads.db` will pollute diagnostics and generate false positives for active cities.

To close these testing and CI/CD gaps before design and implementation proceed, this lane is a **BLOCK**.

---

## Top Strengths

- **Excellent Acceptance Traceability (AC1–AC14)**: Every acceptance criterion is now paired with an explicit, automated verification method (such as resolved-config tests, pack-resolution precedence, absence scans, and golden outputs), eliminating reliance on human vigilance.
- **Automated Absence-Scan Invariant (AC8)**: Demands a real, programmatic scan with defined positive controls (fails red on stray tokens) and negative controls (passes green on allowed paths), making Core role-neutrality a continuous CI gate.
- **Explicit Test Coverage Migration (AC13)**: Mandates a formal mapping from retired legacy tests to their new homes, ensuring that foundational environment setup and pack-loading assertions are not lost during the split.

---

## Blocking Findings

### [Blocker] Dangerous Network Dependency in CI for Public Gastown Validation (AC14)
- **The Finding**: AC14 requires that "Public Gastown validation is part of acceptance... the public pack checkout or pinned cache must prove the roles, prompts... needed by supported Gastown templates."
- **The Impact**: If the test suite or CI pipeline attempts to fetch or resolve the remote Git repository `https://github.com/gastownhall/gascity-packs.git` on every single commit or PR build, the pipeline will suffer from severe flakiness due to GitHub rate-limiting, network latency, or remote downtime. This violates the core engineering principle of self-contained, high-performance, and offline-compatible test suites.
- **Required Change**: Refine AC14 to explicitly require a **local filesystem-based Git remote fixture** for normal CI runs:
  - The integration test suite must seed a bare Git repository under `test/fixtures/gascity-packs-gastown.git` containing the current state of the Gastown pack.
  - Standard CI runs must redirect resolution requests to this local file URI (e.g., `file:///data/projects/gascity/test/fixtures/gascity-packs-gastown.git`) to verify resolution, checkout, and cache behavior.
  - Fetching from the actual remote `github.com` must be strictly restricted to a final, separate release-gate command run only during deployment packaging, protecting the core CI loop from external dependencies.

### [Blocker] Absence-Scan Evasion via Materialized, Embedded, or Generated Assets (AC8)
- **The Finding**: AC8 requires Core role neutrality to be enforced across "Go production code, Core assets, formulas, orders, prompts... and route targets" via an "Absence-scan test" (lines 89-91).
- **The Impact**: If the absence-scan test is implemented as a simple static text scanner over raw source files in `internal/packs/core`, it can easily be bypassed. For example, a code generator could dynamically inject forbidden role names into generated Go metadata or embed files during compilation, or compiled virtual filesystems (e.g., `builtinpacks.All()`) could contain stale, role-specific metadata that is not present in the clean source tree. Static source scans alone will miss regressions introduced during pack materialization.
- **Required Change**: Require that the absence-scan test must run against **two distinct surfaces**:
  1. **Static Source Scan**: A case-insensitive, word-bounded regex scan of all files in the source tree under `internal/packs/core` (excluding documentation and allowed test fixtures).
  2. **Compiled/Materialized In-Memory Scan**: A runtime scan that programmatically inspects the virtual filesystems returned by `builtinpacks.All()` and the fully resolved in-memory configurations loaded by the session manager. This ensures that no forbidden role name is dynamically injected, embedded, or loaded from stale cache metadata during actual binary execution.

### [Blocker] Lack of a Failure Contract for the Coverage-Transfer Table (AC13)
- **The Finding**: AC13 requires a "Coverage-transfer table mapping old tests to new tests," but does not specify a failure contract (lines 94-95).
- **The Impact**: As written, a developer could submit a coverage-transfer table that is partially empty or silently drops critical integration assertions (e.g., omitting the retired `examples/gastown/packs/maintenance/pack_test.go` assertions). Because there is no rule enforcing completeness, this "Manual Inventory Fallacy" will allow coverage gaps to slip through unnoticed.
- **Required Change**: Refine AC13 to demand a **strict completeness validator**:
  - The coverage-transfer table or its validation script must assert that *every single* retired test file under `examples/` and `testenv_import_test.go` has a matching row in the table mapping it to a verified replacement test path or a documented deletion rationale.
  - The validation check must fail the build if any retired test file is unrepresented in the coverage-transfer table.

### [Blocker] Diagnostic Pollution from Historical Dolt Database State (AC10)
- **The Finding**: AC10 and AC11 specify that doctor and import-state diagnostics must report legacy local imports, stale system packs, and retired Maintenance paths (lines 91-92).
- **The Impact**: In real-world cities, the Dolt task store (`beads.db`) contains hundreds of closed, completed, or historical beads. If the doctor diagnostics perform a naive scan over the entire database, they will encounter historical beads that reference retired `packs/maintenance` or legacy roles. This will trigger "retired path" warnings on historical records, locking up the doctor CLI with endless legacy alerts on completed work, even though the active city configuration is perfectly clean.
- **Required Change**: Refine AC10 and AC11 to specify **strict database-scanning boundaries**:
  - The `gc doctor` and `gc import-state` commands must only scan the active/resolved configuration files on disk and *active, in-progress* beads in the database.
  - The diagnostics engine must explicitly ignore historical, closed, or completed beads in the Dolt task store to prevent diagnostic pollution and false-positive warnings.

---

## Lane-Specific Detailed Responses

### Q1: Does each acceptance criterion map to a concrete unit, integration, command, or absence-test proof?

**Yes.** Unlike Attempt 1, every AC now has a corresponding verification mode. To prevent implementation drift and ensure rapid developer feedback loops, we must establish a clear **Acceptance-to-Proof Traceability Table** defining the exact test environment, execution target, and execution timing (gate):

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
To prevent the absence scan (AC8) from becoming a source of build frustration or a vector for bypasses:
- **No Wildcard Exclusions**: The scanner must not accept broad directory-level ignores (such as ignoring the entire `test/` directory).
- **Line-Gated Whitelist**: Documented migration exceptions or positive control files must be explicitly whitelisted in `absence_scan_test.go` by filename and precise line ranges (e.g., `REQUIREMENTS.md#L45`). Any occurrence of a forbidden word outside this explicit whitelist must immediately cause the test to fail.

---

### Q3: Are legacy Maintenance auto-inclusion tests rewritten to cover required Core plus explicit external Gastown imports?

**Yes, the requirement conceptually addresses this under AC13.**
However, as noted in the blockers, we must back this up with a strict completeness validator so that retired tests cannot be dropped from the migration map without a documented, verified rationale.

---

## Missing Evidence

To ensure the test strategy is fully executable without downstream inference, the requirements document must add:
- An Example Mapping scenario showing **Diagnostic Boundaries on Historical Beads**:
  - *Input*: The `beads.db` contains a historical bead created in 2025 with the title "Assigned to Mayor."
  - *Expected*: `gc doctor` runs, scans the active config and current active sessions, and reports **green/clean**, completely ignoring the historical "Mayor" string in the closed bead.
- A specification of a **bare Git remote fixture** under `test/fixtures/` to simulate the external `gascity-packs/gastown` repository, enabling completely offline and deterministic CI.
- A requirement that the **role-neutrality absence scan (AC8) must execute as a pre-commit hook** and run in under 500ms to guarantee immediate feedback.

---

## Required Changes & Actions

1. **AC8 (Role Neutrality Scan)**: Update the text to state that the absence scan must run against both the static source assets and the compiled, materialized virtual filesystems in memory (e.g., via `builtinpacks.All()`).
2. **AC10 & AC11 (Database Boundaries)**: Update to specify that diagnostics and repair commands must only analyze resolved configurations and active, in-progress beads, explicitly ignoring historical, closed, or completed beads in Dolt.
3. **AC13 (Coverage Transfer)**: Update the criterion to require a strict completeness validation check that fails the build if any retired test file is missing from the transfer matrix without a documented deletion rationale.
4. **AC14 (Public Gastown Validation)**: Update the verification contract to require a local, bare Git remote repository fixture (`test/fixtures/...`) for deterministic offline CI runs, restricting live external network calls to the final release-gate command.
5. **Cross-Document Consistency**: Ensure all references to the public Gastown repository are standardized consistently to `https://github.com/gastownhall/gascity-packs.git//gastown` as pointed out by the Role Neutrality Guardian, resolving the organization naming mismatch.

---

## Questions

1. To protect local developer feedback loops, can we package the static absence-scan (AC8) and the ledger-validation checks (AC6) into a lightweight pre-commit script that runs instantly upon `git commit`?
2. If an operator-initiated repair (AC10) encounters a read-only or immutable configuration file (e.g., in a sandboxed or containerized environment), what is the expected fallback behavior (e.g., should `gc doctor` output copy-pasteable manual overrides)?
3. How will we coordinate the public release of the `gc` binary and the external Gastown pack to ensure that the pinned version in fresh init (AC4) points to a tag/SHA that is already publicly available and pushed to GitHub?
