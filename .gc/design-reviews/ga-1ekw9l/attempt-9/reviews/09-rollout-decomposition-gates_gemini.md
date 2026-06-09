# Iris Kowalski — DeepSeek V4 Flash (Rollout and Decomposition Gates Reviewer, Attempt 9, Independent Review)

**Verdict:** block

**Lane:** independently deployable slices, decomposition readiness, prerequisite honesty, exact gates, cross-repo sequencing and test coverage.

Reviewed the updated 835-line `plans/core-gastown-pack-migration/implementation-plan.md` (representing the Attempt 9 / Iteration 9 state) against `requirements.md` and `gc.mayor.implementation-plan.v1`, grounded in the live repository's structure and cross-repository dependencies.

---

## Executive Summary

As Iris Kowalski, the **Rollout and Decomposition Gates Reviewer** (role: **rollout-sequencing-auditor**), I have conducted an independent, evidence-backed review of the Attempt 9 (Iteration 9) design.

While the rollout architecture has significantly evolved, incorporating multi-slice rollout tables, explicit rollback plans, and sharded integration test targets, critical structural defects remain in the sequencing, prerequisite definition, and gate execution. Specifically, other reviewers have accepted a circular cross-repo bootstrap deadlock, paper-only decomposition gates without executable enforcement, missing start-gate dependencies, and overly bundled "big-bang" rollout slices.

Because these defects allow premature bead creation, unsequenced pin consumption, and silent downgrade regressions, my verdict is **Verdict: block**. The required changes do not demand restructuring the entire plan, but rather tightening the programmatic start gates, resolving the circular staging deadlock, and separating the densest rollout slices to ensure safe, independent, and verifiable deployment.

---

## Top Strengths & Design Evolution

- **Granular Multi-Slice Table**: The slice-to-gate table (§"Rollout and Recovery", lines 793-808) establishes clear start gates, merge gates, and one-way boundaries, providing a strong baseline for independent deployment.
- **De-batched Landing Sequence**: De-batching the migration into eight distinct slices (1a-1c, 2, 3, 4a-4c, 5a-5b, 6, 7) directly addresses the risk of "one fragile landing."
- **Rigorous Multi-Pin Separation**: Defining distinct pins (`compatibility` pin in Slice 2 vs `activation` pin in Slice 5a) allows testing co-existence and no-Maintenance modes in a controlled candidate branch before full activation.
- **Tiered Test Policy**: Requiring sharded process and integration tests (`make test-cmd-gc-process-parallel` and `make test-integration-shards-parallel`) for high-risk slices ensures we do not lean on fast unit tests alone.

---

## Critical Risks & Assumptions Accepted Too Quickly

### 1. [Blocker] Missing Slice 1b Dependency in Slice 2's Starting Gate
- **The Assumption**: The gate table (Line 800) states that Slice 2 "May start when AC15/AC16 cache schema proof exists."
- **The Reality**: Slice 2's objective is to "Update `internal/config/PublicGastownPackVersion` to the public compatibility commit" (Line 740). This compatibility commit is only created, tested, and published during Slice 1b!
- **The Blocker**: If Slice 2 starts when only the cache schema proof is present, it is scheduled to adopt a public pin that does not yet exist. This breaks prerequisite honesty and independent slice deployability.
- **Required Change**: Update Slice 2's "May start when" gate in the table to read: `"Slice 1b compatibility pin is merged/immutable AND AC15/AC16 cache schema proof exists."`

### 2. [Blocker] Lack of a Named Executable Enforcer for Decomposition Readiness (AC17)
- **The Assumption**: The plan limits implementation work to prerequisite-producing decomposition until support artifacts exist and are cited (Lines 691-699), and AC17 is listed as a "design gate before bead creation" (Line 722).
- **The Reality**: The requirements for AC17 explicitly demand that `"at least one executable gate such as a make target, pre-commit entry, or CI job must enforce the matrix before decomposition."` The plan fails to name or define any executable command, target, or pre-commit hook that validates `support/acceptance-proof-matrix.yaml` and mechanically blocks bead creation.
- **The Blocker**: Without a named executable gate, this boundary remains purely manual and "paper-based." A developer could easily bypass the gate and create/work implementation beads prematurely.
- **Required Change**: Define and name a concrete executable gate in the plan (e.g., `make check-decomposition-readiness` or a pre-commit check) that validates `support/acceptance-proof-matrix.yaml` evidence availability and blocks bead creation if any prerequisite is missing.

### 3. [Blocker] Cross-Repository Staging Deadlock / Circular Dependency
- **The Assumption**: Slice 1c (activation proof) requires generating a "host-Core/no-Maintenance transcript" in `gascity-packs` (Line 734) as an external prerequisite. Slice 5a (activation candidate) in Gas City consumes this pin to run and verify the no-Maintenance loader (Lines 766-770).
- **The Reality**: This is a classic circular deadlock. The public pack CI cannot produce the "host-Core/no-Maintenance transcript" because a Gas City binary that supports loading in that mode does not exist yet (it is only built and verified in Slices 4a and 5a). Conversely, Gas City cannot merge or verify Slice 5a because the activation pin from Slice 1c has not landed.
- **The Blocker**: Programmatic CI checks in both repositories will block progress indefinitely due to mutually unfulfilled immutable SHA and transcript constraints.
- **Required Change**: Document and define a concrete **staging/bootstrap workflow** (such as a local multi-workspace development setup, a local file-path checkout fallback, or a mutable draft branch) that allows local/temporary validation of no-Maintenance loading to break this circular dependency before final immutable SHAs are published.

### 4. [Major] High-Risk "Big-Bang" Landings in Slices 5b and 7
- **The Assumption**: Slices are assumed to be sufficiently decomposed, but Slices 5b and 7 remain overly dense.
- **The Reality**: Slice 5b couples removing Maintenance from required builtin packs, moving Core-owned Maintenance assets into Core, and consuming Gastown assets from the public pack. Slice 7 couples deleting the legacy examples/packs source tree with updating all operator-facing documentation, tutorials, and doctor goldens.
- **The Blocker**: These represent "fragile landings." If a regression occurs, identifying whether it was caused by builtin pack removal, asset rehoming, or public-pack consumption is difficult.
- **Required Change**: Further split these slices to minimize risks:
  - **Slice 5b.1 (Asset Rehoming)**: Move Core-owned Maintenance assets into `internal/packs/core`.
  - **Slice 5b.2 (Maintenance Fold)**: Remove Maintenance from `requiredBuiltinPackNames` and transition to public-pack consumption.
  - **Slice 7a (Documentation & Goldens)**: Update documentation, tutorials, and doctor golden files.
  - **Slice 7b (Source Deletion)**: Permanently delete `examples/gastown/packs/*`.

### 5. [Major] Rollback and Downgrade Verification Loophole
- **The Assumption**: The rollback compatibility row (Line 818) allows: `"Doctor-mutated manifests are either readable by old binaries or release notes name explicit downgrade limits."`
- **The Reality**: This "either/or" lacks a clear criterion, and the plan lacks any explicit test in Slices 4b, 4c, or 5a that programmatically executes a downgrade (e.g., running a legacy `gc` binary against a mutated `city.toml`) to verify readability.
- **The Blocker**: If a new binary mutates `city.toml`, and the operator downgrades, the legacy binary could crash or silently ignore critical configuration.
- **Required Change**: Mandate an explicit **downgrade test gate** in Slice 4c's must-pass-before-merge criteria. This test must programmatically run a legacy `gc` binary against a doctor-mutated `city.toml` and verify that it either loads safely (with appropriate ignore semantics) or fails with a clean, operator-readable diagnostic, rather than crashing or swallowing state.

### 6. [Minor] Excluded Slices from Sharded Test Coverage
- **The Assumption**: Only "high-risk loader/doctor/runtime-state slices" run the sharded process and integration tests (Lines 681-685).
- **The Reality**: Slices 5b (Maintenance fold) and 6 (registry/cache cleanup) change required builtin packs, cache keys, and alias resolution. These are highly risky loader and runtime state changes.
- **The Blocker**: Restricting sharded test sweeps to Slices 4a-4c means Slices 5b and 6 could merge with undetected integration or process-level regressions.
- **Required Change**: Explicitly mandate that Slices 5b (Maintenance fold) and 6 (registry/cache cleanup) must also run `make test-cmd-gc-process-parallel` and `make test-integration-shards-parallel` before merge.

### 7. [Minor] Lack of Structured Slice-Level Scoping Tags
- **The Assumption**: The plan's status is `draft`, but the individual slices lack structured phase metadata.
- **The Reality**: Slices 1a-1c are prerequisite-producing and decomposable now, while Slices 2-7 are behavior-changing and strictly blocked.
- **The Blocker**: Without structured tagging, a developer or automated tool parsing the plan may conflate the two phases, leading to premature work on blocked slices.
- **Required Change**: Prefix each slice in the rollout list with an explicit metadata tag, such as `[Phase: Prerequisite / Decomposable Now]` or `[Phase: Behavior-Changing / Blocked until AC1-AC16 Complete]`.

---

## Missing Evidence

1. **Named Executable Command for AC17**: The specific make target or pre-commit command that validates evidence availability in `support/acceptance-proof-matrix.yaml`.
2. **Bootstrap Setup Commands**: The CLI commands or configuration flags used to perform cross-repo staging and verify the no-Maintenance loader locally.
3. **Downgrade Validation Logic**: The programmatic assertions used to verify legacy binary compatibility with mutated `city.toml` files.

---

## Required Structural & Schema Changes

To lift this block, the following changes must be made to `implementation-plan.md`:

1. **Amend Slice 2 Start Gate**: In the Slice-to-Gate table (Line 800), add Slice 1b completeness as an explicit prerequisite.
2. **Define the AC17 Executable Gate**: Specify a named validation command/target (e.g., `make check-decomposition-readiness`) that programmatically enforces the AC17 readiness gate before bead creation.
3. **Incorporate Bootstrap Workflow**: Document the staging/local checkout commands to resolve the cross-repo deadlock between Slice 1c and 5a.
4. **De-batch Slices 5b and 7**: Split Slice 5b and Slice 7 into more granular sub-slices as detailed in Risk #4.
5. **Add Programmatic Downgrade Tests**: Add an explicit downgrade verification test to Slice 4c's merge gates.
6. **Extend Sharded Integration Coverage**: Explicitly require `make test-cmd-gc-process-parallel` and `make test-integration-shards-parallel` to pass before merging Slices 5b and 6.
7. **Add Slice-Level Phase Metadata**: Add explicit `[Phase]` tags to each slice to prevent phase conflation.

---

## Questions

1. **AC17 Enforcement**: What exact validation tool or script will `make check-decomposition-readiness` run, and will it be integrated into the git pre-commit hook?
2. **Staging / Local Testing**: Can we use a local `git replace` or directory-based override in `go.mod` or `city.toml` to safely execute and transcript-verify the no-Maintenance loader during the staging phase?
3. **Downgrade Boundaries**: What specific legacy binary version will be used for the downgrade tests in Slice 4c?
4. **Pruning of Stale Cache**: In Slice 6, how do we programmatically verify that the stale synthetic cache alias is completely inactive without relying on manual cache purges?

---

## Schema Conformance (in scope)

The plan conforms perfectly to the `gc.mayor.implementation-plan.v1` schema:
- All required front-matter fields are correctly defined.
- All seven top-level sections are correctly named and ordered.
- The `Open Questions` section is correctly structured with explanatory prose (lines 830-835).
