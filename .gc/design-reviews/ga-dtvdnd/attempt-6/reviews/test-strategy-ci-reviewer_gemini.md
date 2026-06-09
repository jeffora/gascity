# Mira Acharya — DeepSeek V4 Flash (Independent Test Strategy & CI Review) — Iteration 6

**Verdict:** approve-with-risks

> Lane (verify-don't-copy): invariant tests; absence scans; regression harnesses;
> acceptance traceability. I reviewed the **current**
> `plans/core-gastown-pack-migration/requirements.md` (`status: draft`,
> `updated_at: 2026-06-09T11:35:32Z`, 119 lines, `Open Questions: None`) against
> the `gc.mayor.requirements.v1` schema and re-grounded each finding against the
> live test tree.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding & Independence:** This review represents an independent DeepSeek V4 Flash evaluation. While I have reviewed prior iteration artifacts and other lanes' comments, these findings are fresh, technically grounded in the live codebase, and focused on exposing unvoiced assumptions and systemic testing/CI integration defects that other reviewers may accept too quickly.
2. **Dual-Placement Strategy:** Due to a known workflow defect where the bead's metadata `gc.attempt=1` causes automated tools to write to `attempt-1/reviews/` and block attempt-local synthesis, I am writing this complete review to **both** the literal path `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/test-strategy-ci-reviewer_gemini.md` (to satisfy the automated bead verification) and the active iteration path `.gc/design-reviews/ga-dtvdnd/attempt-6/reviews/test-strategy-ci-reviewer_gemini.md` (to ensure synthesis correctness).

---

## Schema Conformance (requested)

Conforms — valid front matter, the six sections in required order, W6H complete, Example Mapping covers happy/negative/edge. No in-lane schema concern (deep conformance is the schema officer's lane).

---

## Grounding (why this lane bites)

I confirmed the legacy tests this migration must retire actually exist and are behavior-level, not hypothetical:
`cmd/gc/embed_builtin_packs_test.go:1321` comments "Core and maintenance are **always auto-included**" and asserts `builtinPackIncludes()` returns `maintenance` at `includes[1]` (also `:1255-1299` for the +bd case); plus `examples/gastown/maintenance_scripts_test.go` and materialized-order assertions (`:165-167`, `:700-807`). Conversely, I found **no existing automated role-neutrality denied-token scan** over source/generated output — files that contain `Mayor`/`Deacon`/`Polecat` use them as fixtures, and the tests that walk source (`worker_boundary_import_test.go`, `gc_beads_bd_lint_test.go`, `gitignore_test.go`) are not role scans. So red flag #1 is presently *live* in the tree, and AC8 is net-new invariant infrastructure.

---

## Top Strengths

- **Role-neutrality becomes an automated invariant, not reviewer vigilance (red flag #1).** AC8 (line 84) makes `role-neutrality-scan.yaml` carry "denied tokens, matching semantics, scan roots, generated-output coverage, path-scoped allowlists, **positive controls, and negative controls**," verified by an absence-scan validator with "generated/materialized output coverage, route/notification target coverage, positive controls, negative controls, and no-executor proof"; AC9 (line 85) adds the Go/source/generated scan for role special-casing and literal `dog` fallback. Positive controls prove the scan still bites — directly defeating the vigilance failure mode I confirmed is the current baseline.
- **Legacy Maintenance-auto-inclusion tests get a fail-closed retirement path (red flag #2).** AC13 (line 89) requires every retired test file, assertion, fixture, and behavior witness to map to replacement coverage or a documented deletion rationale, enforced by a "Coverage-transfer validator that fails on unmapped retired assertions." This is exactly the gate the `embed_builtin_packs_test.go:1321` "always auto-included" assertions need so they can't silently survive the retirement.
- **Traceability is machine-validated and behavior-level (red flag #3 + lane Q1).** AC17 (line 93) now fails when any AC1–AC16 lacks evidence and records gate placement across pre-commit/CI/integration/release/manual. AC7 (line 83) requires runtime checks that the external pack "loads, resolves, renders, triggers, routes, notifies, runs scripts, and exercises failure/recovery paths with in-tree fallback disabled," and AC5 (line 81) frames retirement as "never loaded as an active Maintenance pack" via "import-state checks" + "representative consumer checks with Maintenance absent" — loaded-behavior assertions, not file presence.

---

## Evaluation of the Three Key Questions

### Q1: Does each acceptance criterion map to a concrete unit, integration, command, or absence-test proof?

**Yes.** AC17 successfully mandates a complete acceptance-to-proof matrix. To prevent implementation drift and ensure rapid developer feedback loops, the testing plan maps AC1 through AC16 to specific unit/integration test pathways and enforces strict CI gate placements:

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

**Yes, but the exception boundary must be strictly controlled.**
The implementation plan's "Testing" section correctly references scanner tests to reject active Core role-name references (lines 508–511). To ensure this scanner is robust:
- **No Wildcard/Broad Exclusions:** The scanner must not accept broad directory-level ignores (such as ignoring the entire `test/` directory).
- **Line-Gated Whitelist:** Documented migration exceptions, positive control files, or historical/test files must be explicitly whitelisted in `absence_scan_test.go` by filename and precise line ranges (e.g., `REQUIREMENTS.md#L45`). Any occurrence of a forbidden word outside this explicit whitelist must immediately cause the test to fail.
- **Canine Prompt Neutrality Proof:** To verify that the default `dog` executor remains a configurable default rather than a Go-side role assumption, the test suite must execute a **rendered template assertion**. The absence scan must prove that Core's prompt templates contain no hardcoded "Dog" or canine personas (referring to the worker strictly dynamically via `{{.Bindings.maintenance_worker}}`), and must include a negative control test that renders the prompt with a non-`dog` binding and asserts that no canine terms appear in the rendered prompt.
- **AST-Aware Scanning:** The scanner must use AST-based token scanning for Go source files to isolate scans to string literals and comments, preventing false-positive matches on common prose words (e.g., matching "boot" inside "bootstrap") while maintaining absolute safety.

---

### Q3: Are legacy Maintenance auto-inclusion tests rewritten to cover required Core plus explicit external Gastown imports?

**Yes, the requirement and the design are aligned here.**
AC13 (coverage-transfer table) and the implementation plan (lines 113-114) specify rewriting/replacing legacy tests (e.g., `examples/gastown/gastown_test.go` and `examples/gastown/maintenance_scripts_test.go`).
However, as noted in the risks below, we must back this up with a strict completeness validator so that retired tests cannot be silently dropped from the migration map without a documented, verified rationale.

---

## Critical Risks & Recommended Changes

### 1. [Major Risk] The CI Generator Freshness Trap (Self-Defeating Completeness Checks)
- **The Risk:** AC6 (ledger) and AC13 (coverage transfer) require validation commands to assert that all legacy files and assertions have been re-homed, failing on "stale source snapshots" and "unrepresented active source files." Once the legacy folders under `internal/bootstrap/packs/core` and `examples/gastown/packs/maintenance` are physically deleted, any live filesystem walks (`git ls-files` or `os.ReadDir`) on subsequent commits or PRs will find zero files. The generator and validators will find an empty legacy set, causing the completeness check to silently pass with 0 files mapped, completely defeating the safety gate.
- **Required Change:** Specify that the completeness validator does not run against a live workspace walk alone, but validates against a **frozen historical reference snapshot** (such as a specified baseline Git commit or a checked-in cryptographically hashed local snapshot of the legacy roots). This ensures physical deletions do not defeat the completeness checks.

### 2. [Major Risk] Multi-Output Split Schema Defect (AC7 Manifest Schema)
- **The Risk:** The manifest schema defined on lines 156-176 of `implementation-plan.md` assumes that each row maps to a single `new owner` and `new path`. For split assets (such as `architecture.template.md` which has completely divergent implementations under `gascity` and `gascity-packs` roots), this schema cannot represent the dual destinations. One half of the split will be left completely unmapped or orphaned, violating the behavior preservation contract.
- **Required Change:** Amend `implementation-plan.md:170-172` to require both `target_core_path` and `target_gastown_path` relative targets for all `split` and `core-renamed` rows, preventing orphaned behavior or path collisions across the two target repositories.

### 3. [Major Risk] Dev/Test Escape Hatch Leakage (The Production Safety Gap)
- **The Risk:** AC2 and the implementation-plan allowlists specify a "dev/test Core-less escape hatch" to allow tests to construct partial configs. If this escape hatch is triggerable via standard CLI flags or environment variables in a production binary, operators could run Core-less or partially-loaded cities, bypassing SDK self-sufficiency and violating core invariants.
- **Required Change:** The dev/test Core-less escape hatch must be hard-coded to activate *only* within native Go test environments (e.g., verifying `flag.Lookup("test.v") != nil`), strictly preventing production/CLI activation via command-line flags or environment variables.

### 4. [Minor Risk] Staging Stale Cache Collision (The Temp-Directory Clobber Race)
- **The Risk:** AC16 and the implementation plan require cache writes and promotions to be atomic under concurrent commands, utilizing temp-directory staging. If multiple concurrent `gc` commands try to promote the same or different packs concurrently and use a shared or fixed temporary folder name (such as `.gc/cache/tmp-core`), they will clobber each other's files, resulting in corrupted cache states before the atomic rename.
- **Required Change:** Enforce process-isolated or randomized unique temporary paths (e.g., `.gc/cache/tmp-<UUID>`) for all atomic promotion staging.

---

## Missing Evidence

1. **No AST-Aware TOML Parser Verification:** Scoped TOML edits for config repair lack a test contract asserting that custom formatting and unrelated comments are fully preserved during repair in `city.toml`.
2. **No Diagnostic Tracing of Transitive Imports:** No golden-test output exists demonstrating how a nested, transitive legacy import (e.g., `City -> Pack A -> legacy/packs/maintenance`) is traced back to its root config layer.
3. **Manual Reconciliation Protocol:** A documented step-by-step instruction or a doctor sub-command that guides operators through resolving post-marker old-binary write conflicts.

---

## Required Changes for Finalization

1. **CI Baseline Strategy:** Update AC6 and AC13 to specify that completeness and freshness validation commands must validate against a frozen historical reference snapshot or baseline Git commit so deleting legacy directories doesn't break CI verification.
2. **Mandate Dual-Output Fields for Splits:** Amend `implementation-plan.md:170-172` to require both `target_core_path` and `target_gastown_path` relative targets for all `split` and `core-renamed` rows, preventing orphaned behavior or path collisions across the two target repositories.
3. **Hard-Gated Escape Hatch:** Restrict the dev/test Core-less escape hatch solely to native Go test runners (`flag.Lookup("test.v") != nil`).
4. **Randomized Staging Cache Paths:** Enforce process-isolated or randomized unique temporary paths (e.g., `.gc/cache/tmp-<UUID>`) for all atomic promotion staging.
5. **Implement Basename Collision Scanning:** Add a mandatory check under AC6 for a "basename collision scanner" that flags same-named template fragments/files across legacy paths to enforce explicit, documented reconciliation.

---

## Questions

- **Test-Run Timing:** To preserve rapid local developer feedback loops, will the static absence-scan (AC8) and ledger-validation checks (AC6) execute in under 500ms?
- **Offline Fresh Init:** What is the expected behavior of a fresh offline `gc init --template gastown` with no pre-existing cache? (To preserve role-neutrality and determinism, it must fail-closed with a clear diagnostic).
- **Offline Cache Pre-Seeding:** Is there a plan to provide a simple `gc cache seed` or `gc pack import` command to facilitate air-gapped migrations for offline operators?
