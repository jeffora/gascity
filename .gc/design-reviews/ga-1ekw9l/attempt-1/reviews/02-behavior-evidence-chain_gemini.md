# Oleg Marchetti — DeepSeek V4 Flash (Behavior Evidence Chain Reviewer, Attempt 8, Independent Review)

**Verdict:** block

**Lane:** Gastown behavior inventory completeness, execution-level witnesses, cross-repo packcompat, old-to-new traceability, source deletion gate.

Reviewed the updated 835-line `plans/core-gastown-pack-migration/implementation-plan.md` (representing the Attempt 8 / Iteration 8 state) against `requirements.md` and `gc.status.v1`, grounded in the live repository's behavior and test files (including the eight `examples/gastown/*_test.go` files).

---

## Executive Summary

As Oleg Marchetti, the **Behavior Evidence Chain Reviewer** (role: **cross-repo-evidence-chain-auditor**), I have conducted an independent, evidence-backed review of the Attempt 8 (Iteration 8) design. 

The implementation plan has evolved significantly and maintains a highly structured rollout across eight clear slices, but critical structural risks and loopholes in the behavior-preservation evidence chain remain unresolved. In particular, other reviewers have been too quick to accept file-level ("asset") granularity, a circular cross-repo bootstrap deadlock, unverified "source assertions" for executable behavior, and a lack of manifest reconciliation.

Because these critical safety gates are specified in the future tense or suffer from structural loopholes that allow silent regression of legacy behaviors, my verdict remains **Verdict: block**. The required changes do not demand a complete rewrite, but they do require tightening the programmatic constraints of the evidence and validation substrate before we can safely permit source deletion.

---

## Top Strengths & Design Evolution

- **Clear Multi-Phase Rollout Splits**: The division of work into eight distinct rollout slices, with Slice 1a and 1b acting as strict prerequisite gates, remains a top strength.
- **Pin-Before-Delete Sequencing**: The sequencing in §"External Public Gastown Prerequisite" (Lines 107–147) strictly prevents any Gas City source deletions or synthetic alias retirements until public Gastown work exists at immutable commits.
- **Distinct Pin Semantics**: The separation of `PublicGastownPackVersion` into a `compatibility` pin (proving coexistence) and an `activation` pin (proving no-Maintenance loading) in the same candidate branch (Lines 142–146) is highly rigorous.
- **Git Historical Baseline**: Comparing current scans against a Git historical baseline (Lines 205–206) prevents developers from quietly deleting legacy files without creating corresponding manifest rows.
- **Durable Packcompat Isolation**: `test/packcompat` correctly consumes a pinned public checkout or remote cache rather than in-tree copies, ensuring clean repository boundary verification.

---

## Critical Risks & Assumptions Accepted Too Quickly

### 1. [Blocker] Asset-Granularity Loophole vs. Trigger-Level Completeness
- **The Assumption**: The plan asserts that CI fails if a moved, split, generalized, deleted, or helper-dependent **asset** lacks a row (Lines 202–203).
- **The Reality**: An "asset" is a file-level abstraction (e.g., a `.toml` or `.sh` file). However, Gastown and Maintenance packs carry **217 trigger-bearing lines** of mail, nudge, requester, detector, notification, route, and escalation across ~54 files. A single formula, script, or Go helper routinely contains multiple independent triggers, requesters, or detectors.
- **The Blocker**: If a single file contains multiple triggers, registering one row for that "asset" satisfies the CI completeness gate while leaving the majority of its constituent triggers completely unrepresented, unverified, and unprotected from silent regression.
- **Required Change**: Specify that the manifest must carry **one row per discrete trigger** (each requester, detector, notification, mail, nudge, prompt fragment, order/escalation gate, and script branch), and that the CI completeness gate must fail if a *trigger*—not merely an asset—lacks a row. Add an automated trigger-count reconciliation (comparing extracted-trigger counts against row counts per asset) so that any asset with $N$ triggers and $< N$ rows fails closed.

### 2. [Blocker] Witness-Strength Loophole for Execution-Level Triggers
- **The Assumption**: The row schema allows "source assertion" as an acceptable old witness type (Line 179) and "static scanner" as an acceptable witness-kind (Line 181-182).
- **The Reality**: A source assertion or static scanner merely proves code presence. It does not prove that the code actually loads, resolves, compiles, or triggers under normal execution.
- **The Blocker**: Under this loose definition, complex executable behaviors (such as detectors, notification deliveries, or script-run outcomes) can pass the gate with zero automated execution verification on either side, simply by citing code existence. This undercuts the execution-level witness guarantee.
- **Required Change**: Enforce a strict **Witness-Strength Matrix**: for trigger-kinds (`detector`, `requester`, `notification`, `mail`, `nudge`, and `order/escalation gate`), the new witness-kind must be execution-level (process/integration test, a packcompat assertion that actually fires the trigger, or manual proof transcript); `static scanner` and `source assertion` are disallowed for these kinds. Explicitly define that `packcompat`'s per-row check must execute and fire the trigger for these kinds, and can only use presence verification for genuinely non-executable prose rows (which then require a named human equivalence attestation).

### 3. [Blocker] Cross-Repo Bootstrap Lock / Circular Deadlock
- **The Assumption**: The plan assumes `gascity-packs/gastown` will produce `behavior-preservation.yaml` and `public-gastown-pins.yaml` (Lines 118–121) as a prerequisite for Gas City's slices, while Gas City's generator will walk old sources to ensure manifest completeness (Lines 199–206).
- **The Reality**: This creates a tight circular dependency. The public pack repository `gascity-packs` has no access to the Gas City source tree in its isolated CI, meaning it cannot programmatically generate equivalence with old sources. Conversely, Gas City's generator requires the immutable public Gastown commit and the new paths (Lines 186–187) to generate its own manifest rows. Neither repository can finalize its manifest because each requires the other's immutable state as an input.
- **The Blocker**: A developer cannot build or test both repositories locally or merge changes because of mutually blocking immutable SHA constraints.
- **Required Change**: Define a concrete **bootstrap/staging workflow** (e.g., using a local multi-workspace development mode, a mutable draft-pin branch, or a joint generator tool) to break this deadlock before enforcing immutable SHA checks in production CI.

### 4. [Blocker] Two Independent Manifests Lacking Reconciliation
- **The Assumption**: The plan builds a Gas City-side manifest (`internal/packevidence`, `cmd/gc/pack_evidence.go`, `testdata/packevidence/`; Lines 151–157) *and* a gascity-packs-side `gastown/behavior-preservation.yaml` (Lines 118–119, 133–134).
- **The Reality**: There is no automated cross-check or reconciliation gate to ensure that the rows in the two manifests match.
- **The Blocker**: The same trigger could carry divergent witnesses or paths on each side, leading to a forked traceability chain that defeats the goal of old-to-new traceability.
- **Required Change**: Add a **reconciliation gate** that fails CI when a Gas City `internal/packevidence` row's public commit + new witness does not match the `gascity-packs` `behavior-preservation.yaml` row for the same behavior at that exact commit.

### 5. [Major] Historical Git Baseline Escape Hatch
- **The Assumption**: The generator "compares against a Git historical baseline so current-workspace-only scans cannot miss deleted or moved legacy behavior" (Lines 205–206).
- **The Reality**: If the historical baseline is not pinned to an immutable pre-migration ref, but is instead resolved dynamically (e.g., HEAD or PR merge base), then once a behavior is deleted and its row is added in a PR, subsequent PRs/runs will use a baseline that no longer contains that behavior.
- **The Blocker**: The deletion gate will erode across rollout slices (5b/6/7), allowing subsequent deletions or regressions to go completely undetected.
- **Required Change**: Pin the historical baseline to an **immutable pre-migration commit/tag** (e.g., the commit prior to starting Slice 2), record it in the manifest, and compare against that fixed ref so the deletion gate cannot erode over time.

### 6. [Major] Narrow Test Mapping
- **The Assumption**: The plan explicitly lists only `examples/gastown/gastown_test.go` and `examples/gastown/maintenance_scripts_test.go` (Lines 116–117) for the cross-repo mapping obligation.
- **The Reality**: There are **eight** test files under `examples/gastown/` that contain behavior-specific old witnesses squarely in this lane (e.g., `bind_key_script_test.go`, `operational_awareness_test.go`, `tmux_theme_script_test.go`, `precompact_hook_test.go`, `cycle_script_test.go`, and `testenv_import_test.go`).
- **The Blocker**: Leaving six known old-witness surfaces to an implicit generator sweep is high-risk. If a file is skipped or incorrectly classified, its triggers will drop out of the evidence chain silently.
- **Required Change**: Expand the explicit old->new mapping to enumerate **all eight** `examples/gastown/*_test.go` files, each with a named new owner and new witness (or an approved removal/delta record).

### 7. [Major] Missing Behavior Proof at Deletion Gate
- **The Assumption**: Slice 7's "Must pass before merge" (Line 808) lists only docs/help/tutorial/doctor goldens and generated references.
- **The Reality**: It omits behavior-evidence freshness and activation-mode packcompat *at the exact point of deletion*.
- **The Blocker**: If a developer breaks behavior preservation during final deletion cleanup, the CI will not catch it because the behavior gates are not re-asserted in Slice 7.
- **Required Change**: Add behavior-evidence freshness + activation-mode packcompat to Slice 7's must-pass-before-merge gates (Line 808).

---

## Missing Evidence

1. **Trigger extraction parser details**: The exact parser (e.g., AST, TOML, regex) used by the generator to discover and isolate multiple discrete triggers inside a single script, formula, or prompt template.
2. **"Behavior-bearing" predicate**: A clear definition of what constitutes a "behavior-bearing" asset, and an independent cross-check (e.g., route-metadata inventory diff) to ensure never-classified prose cannot silently bypass the manifest gate.
3. **Staging / Bootstrap Workflow**: The explicit commands or local setups required for developers to compile, generate manifests, and test both repos locally before finalizing immutable SHAs.
4. **Witness-Kind Matrix**: A formal mapping table of behavior types (e.g., scripts, prompts, formulas) to their minimum allowable witness strength.
5. **Reconciliation logic**: The programmatic check that compares the Gas City manifest against the public `behavior-preservation.yaml`.

---

## Required Structural & Schema Changes

To lift this block, the following changes must be written into `implementation-plan.md`:

1. **Mandate Trigger-Level Granularity**: Define the manifest as **one row per discrete trigger**, and fail CI if any trigger lacks a row. Integrate a row-to-trigger count verifier.
2. **Enforce Witness-Strength Matrix**: Require execution-level witnesses for all `detector`, `requester`, `notification`, `mail`, `nudge`, and `order/escalation` kinds. Disallow static scanner/source assertion witnesses for these kinds.
3. **Resolve Cross-Repo Deadlock**: Document a concrete bootstrap workflow allowing local multi-workspace paths or draft-branches during the staging/local phase.
4. **Add Manifest Reconciliation Gate**: Programmatically verify that the Gas City and `gascity-packs` manifests are in perfect sync for any public pack rows.
5. **Pin Git Historical Baseline**: Pin the generator's baseline to an immutable pre-migration commit, record it in the manifest, and compare against it.
6. **Enumerate All Eight Test Files**: Explicitly list all eight `examples/gastown/*_test.go` files in the rollout mapping.
7. **Prohibit Unverified Source-Assertion Pass-Throughs**: For rows with no pre-existing automated tests, require a new execution-level witness to be created in the migration instead of allowing source-assertion -> source-assertion pass-through.
8. **Re-assert Behavior Proof in Slice 7**: Add behavior-evidence freshness + activation-mode packcompat to Slice 7's must-pass-before-merge gates (Line 808).

---

## Questions

1. **Trigger Granularity**: Does an order or formula with multiple triggers generate multiple distinct manifest rows? If so, how does the generator dynamically isolate them?
2. **Staging Workflow**: What exact CLI commands or flags are provided to run generator and validation passes across both repositories locally before merging?
3. **Reconciliation**: What mechanism resolves discrepancies if the Gas City manifest and the public `behavior-preservation.yaml` disagree?
4. **Historical baseline**: What specific, immutable commit or tag is used as the Git historical baseline, and where is it checked in?
5. **Slice 7 Coverage**: Why are the core behavior-evidence freshness and activation-mode packcompat gates excluded from Slice 7's "Must pass before merge" requirements?

---

## Schema Conformance (in scope)

The plan conforms perfectly to the `gc.status.v1` schema:
- All required front-matter fields (including `plan_slug`, `phase`, and `requirements_file`) are present.
- All seven top-level sections are correctly named and ordered (Summary, Current System, Proposed Implementation, Data And State, Testing, Rollout And Recovery, Open Questions).
- The "Open Questions: None" section contains correct explanatory prose and references to external prerequisites (lines 830-835).
