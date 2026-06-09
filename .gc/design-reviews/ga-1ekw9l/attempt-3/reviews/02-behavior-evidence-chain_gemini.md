# Oleg Marchetti — DeepSeek V4 Flash (Behavior Evidence Chain Reviewer, Attempt 3, Independent Review)

**Verdict:** block

**Lane:** Gastown behavior inventory completeness, execution-level witnesses, cross-repo packcompat, old-to-new traceability, source deletion gate.

Reviewed against the updated design document in Attempt 3 (`plans/core-gastown-pack-migration/implementation-plan.md` / `plans/core-gastown-pack-migration/requirements.md`) and grounded in the live codebase, prior iterations, and other reviewer outputs.

---

## Executive Summary

As Oleg Marchetti, the **Behavior Evidence Chain Reviewer**, I have conducted an independent, evidence-backed review of the Attempt 3 (Iteration 3) design. While other reviewers (Claude and Codex) have moved to `approve-with-risks`, my verdict is **Verdict: block**. 

The design continues to carry fundamental, structural flaws and unexamined assumptions that other reviewers have accepted too quickly. Most notably, the proposed implementation contains a **Git-Historical Baseline blindspot** that allows developers to bypass the validation gate entirely, and a severe **Cross-Repo Bootstrap Lock** (circular dependency) that makes the rollout plan unworkable in practice. Until these structural gaps are closed with explicit, programmatic contracts, this design cannot be approved.

---

## Top Strengths & Design Evolution

- **External Public Gastown Prerequisite**: The introduction of the `gascity-packs` prerequisite section (Lines 104–128) is a powerful addition. It explicitly blocks any Gas City source deletion, role-generalization, or Maintenance removal until the public pack produces matching formulas, scripts, overlays, `behavior-preservation.yaml` evidence, and proof of no-Maintenance loading at immutable commits.
- **Two-Meaning Pinning Sequence**: The compatibility vs. activation pin division is highly sophisticated. Using a compatibility pin to prove coexistence and an activation pin to prove no-Maintenance production loading within the same candidate branch (Lines 124–128) provides an elegant, staged migration sequence.
- **Strict Packcompat Isolation**: `test/packcompat` correctly consumes the pinned public checkout rather than copied workspace assets (Lines 162–166), which ensures the integrity of the external behavior boundary.

---

## Critical Risks & Assumptions Accepted Too Quickly

### 1. [Blocker] Git-Historical Baseline Escape Hatch (The Passive Scanner loophole)
- **The Assumption**: The plan states that "The generator walks old Gas City behavior-bearing sources... CI fails if a moved, split, generalized, deleted, or helper-dependent asset lacks a row" (Lines 155-160).
- **The Reality**: The manifest check remains a passive workspace scanner. If a developer deletes a file, a static scanner walking the *current* codebase will simply see nothing at that path. It cannot know what files *used to exist* or what behaviors *were previously defined* unless it has a historical baseline. 
- **The Blocker**: Without an active Git-Delta scanner, a developer can delete a source file and its associated tests, and the generator will quietly succeed because there are no files left to walk. This undercuts the entire source-deletion gate.
- **Required Change**: The generator must be specified as an active **Git Delta-Assertion Scanner**. It must execute a git diff command (e.g., `git diff --name-status origin/main...HEAD`) to identify all deleted or modified files, and programmatically assert that every deleted or modified behavior-bearing asset has a corresponding active row in the checked-in `behavior-manifest.generated.yaml`.

### 2. [Blocker] Cross-Repo Bootstrap Lock / Circular Dependency
- **The Assumption**: The plan assumes `gascity-packs/gastown` will produce `behavior-preservation.yaml` and `public-gastown-pins.yaml` (Lines 115-118) as a prerequisite for Gas City's slices, while Gas City's generator will walk old sources to ensure manifest completeness (Lines 155-160).
- **The Reality**: This creates a tight circular dependency. `gascity-packs/gastown` must produce a preservation manifest matching Gas City's behavior. However:
  1. `gascity-packs` has no access to the Gas City source tree in its isolated CI, so it cannot "verify" or "generate" equivalence with old sources programmatically.
  2. Gas City's generator requires the immutable public Gastown commit and the new paths (Lines 150-151) to generate its own manifest rows. But that commit cannot exist and be immutable until `gascity-packs` completes its work and merges it.
- **The Blocker**: Neither repository can finalize its manifest or commit because each requires the other's immutable state as an input. 
- **Required Change**: The design must define a concrete **bootstrap/staging workflow** (e.g., using a local multi-workspace development mode, a mutable draft-pin branch, or a joint generator tool) to break this deadlock before enforcing immutable SHA checks.

### 3. [Major] Unfeasible / Undefined Packcompat Execution Model
- **The Assumption**: The plan states that `test/packcompat` "installs public Gastown... and checks one assertion per manifest row" (Lines 162-166).
- **The Reality**: Many Gastown behaviors (like tmux sessions, database migrations, specific agent loop transitions, and notification routes) depend on orchestration and complex mock setups that formerly lived in `examples/gastown/gastown_test.go` and `examples/gastown/maintenance_scripts_test.go`.
- **The Blocker**: Expecting `test/packcompat` in the `gascity` repository to execute a live behavioral assertion for every single row of the manifest without duplicating the entire external test harness and environment is unfeasible. The plan does not define what a "packcompat assertion" actually executes (e.g., is it config loading, dry-run parsing, or full orchestration?).
- **Required Change**: Define the execution boundaries of `test/packcompat` and introduce a **Witness-Kind Matrix** (distinguishing config loading, AST/parse verification, and full-orchestration execution) to prevent unachievable execution proof attempts.

### 4. [Major] "Source Assertion" Loophole Undercuts Execution-Level Witness Mandate
- **The Assumption**: The row schema allows "source assertion" as an acceptable old witness type (Line 145).
- **The Reality**: A source assertion is static (proving code exists). It does not prove the trigger actually fires or behaves correctly. 
- **The Blocker**: For legacy behaviors that lack automated tests, a developer can satisfy the CI completeness gate by writing a manual markdown/text description to satisfy the gate, moving the behavior with zero automated verification on either side.
- **Required Change**: Eliminate "source assertion" as an allowable witness type for migrated, generalized, or split behavior. If a behavior lacks an existing test, the plan must mandate that an automated test or trigger fixture be written *first* in Gas City to witness its execution, prior to any move or deletion.

---

## Missing Evidence

1. **Git Delta-Scanner Mechanism**: The exact git command or scanner logic that discovers deleted files to prevent passive scanner bypasses.
2. **Staging / Bootstrap Workflow**: A clear step-by-step description of how developers compile and test both repositories locally before merging immutable commits.
3. **Witness-Kind Matrix**: A matrix mapping each behavior kind (formulas, orders, prompts, scripts, helpers) to its minimum required witness type (e.g., executable test, golden output, or parse check).
4. **Generator's Trigger-Detection Parser**: The methodology (AST, regex, TOML parsing) the generator uses to enumerate multiple triggers inside a single formula or script.

---

## Required Changes

1. **Mandate Git Delta-Assertion Scanner**: CI must run an active Git-Delta scanner (`git diff --name-status`) to identify deleted files and assert that they are registered in the manifest.
2. **Solve Cross-Repo Deadlock**: Document a concrete bootstrap workflow allowing local multi-workspace paths or draft-branches during the staging phase.
3. **Introduce a Witness-Kind Matrix**: Explicitly define what constitutes a valid old/new witness for each behavior type (e.g. formulas require composition tests; scripts require trigger tests; prompts require render checks).
4. **Tighten "Source Assertion" Rules**: Permit static source assertions only with an explicit, approved "no executable witness possible" record naming owner, justification, and operator impact.

---

## Questions

1. **Circular Dependency**: How can a developer test and generate `behavior-preservation.yaml` locally across both repositories before the `gascity-packs` commit is finalized and immutable?
2. **Multi-Trigger Assets**: Does a single formula or order with multiple independent requesters/detectors generate a single manifest row or N manifest rows? If N, how does the generator discover and isolate them?
3. **Packcompat Scope**: What exact mock infrastructure will `test/packcompat` supply to execute complex, externalized Gastown orchestration assertions?
