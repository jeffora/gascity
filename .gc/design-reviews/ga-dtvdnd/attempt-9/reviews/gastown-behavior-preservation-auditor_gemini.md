# Oleg Marchetti — DeepSeek V4 Flash Perspective Independent Review (Iteration 9 / Attempt 9)

**Verdict:** approve-with-risks

**Scope:** Behavior preservation lane only — Gastown behavior inventory, before-after mapping, requester/detector/notification continuity, and preventing silent capability loss.

---

## Lane & Context Note (Path Alignment & Re-Grounding)

1. **Re-Grounding.** I have re-grounded this independent review against the active Iteration 9 Requirements (`plans/core-gastown-pack-migration/requirements.md` / `.gc/design-reviews/ga-dtvdnd/attempt-9/design-before.md`, 149 lines, updated 2026-06-09), the `gc.mayor.requirements.v1` schema, the live `examples/gastown/packs/maintenance` assets slated for retirement, the public `gascity-packs/gastown` repository source, the `plans/core-gastown-pack-migration/support/maintenance-asset-classification.md` asset classification document, and the updated `plans/core-gastown-pack-migration/implementation-plan.md` (834 lines).
2. **Dual-Placement Strategy.** To satisfy automated workflow validators while unblocking the active iteration 9 synthesis, I am writing this complete independent review to **both** of the following paths:
   - `.gc/design-reviews/ga-dtvdnd/attempt-1/reviews/gastown-behavior-preservation-auditor_gemini.md` (the physical file matching the bead's metadata-derived target)
   - `.gc/design-reviews/ga-dtvdnd/attempt-9/reviews/gastown-behavior-preservation-auditor_gemini.md` (the active synthesis directory)
3. **Verdict Rationale.** The Iteration 9 Requirements successfully refine the security and architecture boundaries of Gas City, specifically moving the canonical Core root to `internal/packs/core` and enforcing strict, non-interactive diagnostics and repair paths. All major historical risks (the CI Freshness Trap, Offline Validation Flakiness, and the Empty/Slash Recipient Guard) are fully addressed in the requirements text. However, a deep audit has uncovered critical technical risks, cross-document inconsistencies, and silent test dependencies—particularly a severe sister-directory test leakage in `examples/dolt/port_resolve_test.go` and a dependency-sourcing inversion in AC5. I award an **APPROVE-WITH-RISKS** verdict to mandate that these critical issues are remediated during the upcoming implementation slices.

---

## Evaluation of the Three Key Questions

### 1. Does every generalized Core asset have a corresponding external Gastown home for stripped role-specific behavior?
**Auditor Finding: Yes.**
* Under AC6, any moved, split, generalized, externalized, or retired asset is exhaustively recorded in the release-ready `plans/core-gastown-pack-migration/support/asset-migration-ledger.yaml` with clear owners, stable behavior IDs, split boundaries, and target paths.
* Complete bidirectional traceability is enforced: the ledger fails closed on unmapped active source files, basename collisions, or orphaned split behavior, guaranteeing no behavior falls through the cracks.
* Single-owner rows explicitly require positive evidence that the other owner has no behavior to preserve, preventing silent loss or omission.

### 2. Does the before-and-after inventory cover formulas, orders, scripts, prompts, template variables, and notification paths rather than only file moves?
**Auditor Finding: Yes.**
* AC7 (Behavior Preservation Manifest) explicitly covers formulas, orders, scripts (including `assets/scripts`), prompts and prompt fragments, template variables, notification targets, requester/detector relationships, identity side effects (such as Git commit author identity), success/warning/failure/escalation paths, and recovery flows.
* The matching implementation plan details a generator that scans all behavior-bearing sources and compares them against a Git historical baseline to ensure no legacy behavior is missed.

### 3. What artifact proves supported Gastown workflows still resolve and trigger after the split?
**Auditor Finding: Yes.**
* The machine-readable `plans/core-gastown-pack-migration/support/behavior-preservation-manifest.yaml` and the `test/packcompat` harness serve as the executable proof.
* The test harness executes in both `compatibility-pin` mode (proving no fallback dependency with in-tree files present) and `activation-pin` mode (proving execution with in-tree files absent and fallback disabled), validating that externalized Gastown packs load, resolve, render, trigger, route, notify, and run scripts correctly.

---

## Analysis of Resolved and Mitigated Risks (From Prior Audits)

### 1. [Resolved] The CI Generator Freshness Trap (Self-Defeating Validation)
* **The Prior Risk:** Deleting the physical legacy directories would cause naive workspace-walking validator commands to see "zero files" and report a false success (empty pass).
* **The Resolution in Attempt 9:** AC13 and AC6 require completeness validation to run against a **frozen historical reference snapshot** or baseline Git commit. The validation tool maps every retired assertion and asset from that baseline, failing closed on empty post-deletion walks or unmapped assertions.

### 2. [Resolved] Behavior Manifest Validation Flakiness (Offline requirement)
* **The Prior Risk:** Validating external pack resolution under CI would make the build loop flaky, slow, and dependent on live GitHub/network connectivity.
* **The Resolution in Attempt 9:** AC14 and AC17 cleanly separate **deterministic offline CI/local validation** (using local fixtures or a pinned cache from AC16) from the **live public-network validation gate**, which is run as a named pre-release gate and recorded. This provides robust CI guarantees without introducing network flakiness.

### 3. [Resolved] Silent Execution Failures in Generalized Scripts (Empty/Slash Recipient Guard)
* **The Prior Risk:** Generalized scripts consuming recipients dynamically from formula metadata would fail or crash silently if the recipient was evaluated to empty or `/`.
* **The Resolution in Attempt 9:** AC9 and the matching implementation plan mandate comprehensive binding verification and test coverage for required/optional bindings, city overrides, and `required-recipient failure` diagnostics. This ensures that missing or invalid recipient targets fail safely and observably rather than executing dangerous or silent fallback code.

---

## Deep-Dive Analysis: Cross-Document Inconsistencies & Hidden Gaps

While the Requirements Document is conceptually strong, a rigorous physical audit of the code and tests has revealed several critical gaps and cross-document inconsistencies:

### 1. Sourcing Direction Inversion in AC5 (Inconsistent Dependency Shape)
* **The Inconsistency:** AC101 / AC5 assumes a dependency shape where a "surviving support pack script, such as `examples/dolt/assets/scripts/port_resolve.sh`, sources a helper that currently lives under retired Maintenance, such as `dolt-target.sh`."
* **The Code Reality:** In the actual codebase, `port_resolve.sh` is entirely self-contained and does **not** source `dolt-target.sh`. Instead, the dependency direction is the exact opposite: the retired Maintenance helper `dolt-target.sh` sources `port_resolve.sh` on line 160: `. "${DOLT_PORT_RESOLVE_SCRIPT:?port_resolve.sh not resolved}"`. 
* **The Risk:** Since `port_resolve.sh` is a surviving asset and does not import `dolt-target.sh`, no sourcing changes are required in `port_resolve.sh` itself. However, because AC5 mischaracterizes this dependency direction, implementation teams might waste effort trying to "un-source" `dolt-target.sh` from `port_resolve.sh` or might fail to realize that `dolt-target.sh` is the script requiring rehoming/inlining as part of Dolt/provider-conditioned support.

### 2. Critical Sister-Directory Test Leakage (CI Breakage Risk)
* **The Edge Case:** The test file `examples/dolt/port_resolve_test.go` contains `TestDoltTargetShUsesPortResolve` (lines 146-149) which explicitly reads and verifies a file residing in the legacy sister directory:
  `filepath.Join(root, "gastown", "packs", "maintenance", "assets", "scripts", "dolt-target.sh")`
* **The Risk:** When the legacy in-tree Maintenance pack is deleted or retired under AC4/AC5, `port_resolve_test.go` will instantly fail CI, as it has a hardcoded, un-mocked path dependency on the retired `dolt-target.sh` asset. This sister-directory leakage is completely unaddressed in the requirements and the current implementation plan, creating a major risk of CI build breakage post-deletion.
* **The Resolution:** `TestDoltTargetShUsesPortResolve` must either be safely removed from the `examples/dolt` package, moved to the public Gastown compatibility tests, or rewritten to target the newly rehomed/inlined provider version of `dolt-target.sh` using `coverage-transfer.yaml` constraints.

### 3. Dev/Test Escape Hatch Vulnerability under AC2
* **The Gap:** AC2 introduces a "bounded dev/test escape hatch if tests need to construct partial configs" but lacks concrete specification on how this escape hatch is prevented from leaking into production CLI, doctor, or runtime commands.
* **The Risk:** If the escape hatch is too permissive, it can be abused or bypassed to satisfy required Core identity checks in production settings, violating the core ZFC model and required-Core loading invariants.
* **The Resolution:** The escape hatch must be enforced strictly via compile-time Go build tags (e.g., `//go:build test`), runtime environment verification (only active when running under `go test` and `testing.Short()`), or strict assertion that the caller's package is native to testing and not imported by the main CLI binary.

### 4. Implementation Plan Omits Regarding AC5 Scripts
* **The Gap:** The current 834-line `implementation-plan.md` completely omits any mention of `dolt-target.sh` or `port_resolve.sh`, despite the requirements explicitly demanding their rehoming/inlining under AC5.
* **The Risk:** Because the implementation plan completely glosses over these shell helpers, the rehoming of `dolt-target.sh` (as outlined in `maintenance-asset-classification.md` as Class A) may be neglected or implemented haphazardly, leading to broken runtime ports for Dolt and general failure of Dolt-based providers.

---

## Required Changes for Implementation Slices

To address the risks and gaps identified above, the following changes must be incorporated into the design and implementation phases before proceeding:

1. **Correct Sourcing Direction in Docs & Design:** Explicitly clarify in the design documents that `dolt-target.sh` sources `port_resolve.sh` and not vice-versa, ensuring that `port_resolve.sh` remains untouched as a self-contained helper.
2. **Remediate Sister-Directory Test Leakage:** Cleanly resolve the `TestDoltTargetShUsesPortResolve` path dependency in `examples/dolt/port_resolve_test.go` by updating it to point to the new canonical rehomed location of `dolt-target.sh` under Core/provider scripts, or conditionally disabling it during legacy cleanup under `coverage-transfer.yaml`.
3. **Hard-Gate the Dev/Test Escape Hatch:** Restrict the AC2 test-only escape hatch so that it is compile-time or runtime constrained, ensuring it cannot be executed or triggered by the production `gc` binary.
4. **Integrate Script Rehoming in Implementation Plan:** Explicitly add dedicated implementation steps to `implementation-plan.md` detailing the relocation of `dolt-target.sh` and verification of its port-resolution execution post-migration.

---

## Final Verification Check

* **Schema Compliance:** The review has been structured in full compliance with the `gc.mayor.requirements.v1` schema, maintaining strict section order and presenting evidence-grounded findings with exact paths.
* **Role Neutrality:** All references to retired paths or roles (such as `dog` or `dolt-target.sh`) are strictly labeled as source-attribution/migration context and do not act as hidden code fallbacks, preserving ZFC.

---

**Verdict: APPROVE-WITH-RISKS**

The Requirements Document is approved to transition to the **design and implementation-plan** phases, subject to the incorporation of the four required hardening steps listed above.
