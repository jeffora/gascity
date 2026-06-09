# Natasha Volkov — DeepSeek V4 Flash (Independent Review, Attempt 13)

**Verdict:** block

**Review focus:** REQUIREMENTS scenario parity, regression prevention, characterization tests, proof freshness — with evidence drawn from verification of `internal/session/DESIGN.md`, `internal/session/REQUIREMENTS.md`, and the active workspace.

---

## Overview

Attempt 13 introduces a highly rigorous and safety-critical defense mechanism: **Slice 0 (session-boundary executable preflight)**. By shifting the entire extraction strategy to block all behavior-moving work until machine-checkable schemas, validators, inventories, and allowlists are checked into the codebase, the design addresses the fundamental risks of boundary drift and unprovable contracts head-on. This is an exceptional, state-of-the-art response to previous concerns.

However, from the perspective of a Behavior Parity Guardian, the design contains several critical gaps where requirements are omitted from the traceability matrix, the approval authority for behavior changes remains undefined, and characterization test definitions remain dangerously loose. These gaps would allow an implementer to skip parity proof for key scenarios or write brittle white-box mocks that miss real behavioral regressions.

---

## Top Strengths

- **The Slice 0 Hard Preflight Gate:** Freezing all subsequent mutation-owning, resolver-adopting, or reconciler-policy beads until Slice 0's validators are checked in and proven to fail on missing/unowned evidence is a textbook example of defensive engineering.
- **Machine-Checkable Schemas:** Standardizing on YAML ledgers (`BOUNDARY_INVENTORY.md`, `SCENARIO_PARITY.yaml`, etc.) validated against JSON Schema (`.schema.json`) files ensures that correctness is enforced mechanically rather than through fragile human reviews.
- **Strict Single-Owner Rule for Start Keys:** Consolidating `pending_create_*`, `instance_token`, `session_key`, and hash metadata into a single shippable Slice 3 prevents split-brain writes and ensures atomic start/rollback transitions.

---

## Critical Risks (Blockers)

### [Blocker] The Scenario Traceability Matrix omits critical scenario rows from REQUIREMENTS.md

While the `## Scenario Traceability Matrix` (DESIGN.md:1071-1080) maps most requirements scenarios to specific slices, it completely omits several active scenario rows defined in the requirements ledger:

- **`SESSION-LIFE-001` (Legacy compatibility states are projected):** Ensures `awake` and `drained` states are mapped correctly. Not assigned to any slice.
- **`SESSION-LIFE-002` (Pending create claim):** Covers the critical wake cause for pending-create claims. Not assigned to any slice.
- **`SESSION-LIFE-006` (Missing config):** Prevents silent waking when config is missing. Not assigned to any slice.
- **`SESSION-LIFE-008` (User-facing projection guard):** Enforces that CLI/API/doctor consumers use projection helpers. Not assigned to any slice.
- **`SESSION-ID-001` (Explicit session names) & `SESSION-ID-002` (Aliases):** Name-validation syntax and live alias collision rules. Not assigned to any slice.
- **`SESSION-RUNTIME-004` (Stop turn):** Dictates when stop-turn interrupts active/pool sessions. Not assigned to any slice.

Without assigning these rows to a specific extraction slice, there is no guarantee that an implementer will test them, write validators for them, or protect them from regression.

**Required change:** Map every single row in `REQUIREMENTS.md` to an extraction slice. Specifically:
- Add `SESSION-LIFE-001`, `SESSION-LIFE-002`, `SESSION-LIFE-006`, and `SESSION-LIFE-008` to **Slice 5** (wake/hold/drain eligibility) or **Slice 0** (as read-only projection properties).
- Add `SESSION-ID-001` and `SESSION-ID-002` to **Slice 1** (target classification).
- Add `SESSION-RUNTIME-004` to **Slice 3** (runtime start) or **Slice 4** (close).

---

### [Blocker] Lack of a designated behavior-change approval authority

TR-001 specifies that refactors must preserve the scenario ledger unless a product ambiguity is identified, and that any product behavior change adds/updates rows. However, the design does not define *who* has the authority to approve behavior changes or requirements updates. 

Without a designated owner-approval gate, any developer or agent can modify `REQUIREMENTS.md` to fit whatever buggy or incomplete refactored behavior they produce, rendering the requirements ledger a subordinate "scratchpad" rather than an authoritative contract.

**Required change:** Update TR-001 or the Product Rule (`DESIGN.md:726`) to explicitly state that any behavior change or requirements ledger update (`REQUIREMENTS.md`) requires explicit owner approval (e.g., the Mayor or Code Owner), and that pull request validators must fail if requirements are edited without a signed-off approval artifact.

---

### [Blocker] Ambiguity in "Characterization Test" definition allows brittle white-box tests

The design rightly mandates that "characterization tests" must be captured before moving logic out of current reconciler or manager code. However, it does not specify what constitutes a valid characterization test. 

If implementers write white-box tests that assert implementation shape (such as mocking internal interfaces, verifying helper function call counts, or asserting specific private struct field layouts), these tests will pass even if the user-visible product behavior diverges or crashes under real workloads.

**Required change:** Add a strict technical requirement to TR-001 stating:
*"Characterization tests must be black-box, end-to-end or integration-level tests asserting user-visible or system-level outputs (such as exit codes, output JSON payloads, state projections, and database commit states) rather than white-box mocks asserting internal function call chains. The exact same characterization tests must run against both the legacy code and the refactored command APIs to prove behavioral parity."*

---

## Major Risks

### [Major] Stale ledger citations are not marked in REQUIREMENTS.md

`REQUIREMENTS.md` currently cites `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go` as evidence for `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`. These files are absent in the active checkout.

While the design's backlog notes this, the requirements ledger itself contains no inline warnings. A developer reading `REQUIREMENTS.md` in isolation has no way of knowing that these citations are stale and that the corresponding behavior is currently uncovered.

**Required change:** Annotate the four affected rows directly in `REQUIREMENTS.md` under the "Evidence" column to mark them as "stale/missing proof," with a cross-reference to the Slice 6/7 backlog items.

---

### [Major] Semantic drift of cited proof tests

The proposed Slice 0 validator `TestScenarioParityFreshness` checks if a cited file path exists, but it cannot verify if the tests inside that file actually execute and assert the required behavior. A test can be renamed, gutted, or marked as skipped (`t.Skip()`) while the file remains, leading to a silent proof-freshness failure.

**Required change:** Require that `SCENARIO_PARITY.yaml` specifies not just the file path, but the exact test function names (e.g., `TestSessionLifecycle/Wake_Held_Until`). The Slice 0 validator must dynamically parse the test file to verify that the named test functions exist and do not contain hardcoded skips.

---

## Minor Risks & Questions

- **Dynamic key scan in static guard:** How will the static guard specified in Slice 0 handle dynamic metadata key writes (e.g., loops writing variable key patterns)? Will it enforce that any `SetMetadata` with a non-literal key is a violation?
- **Is `SESSION-WORK-003` (orphan pool step beads) in scope for Slice 6?** It is listed under Slice 6 in the traceability matrix, but the backlog item for Slice 6 does not mention orphan step beads. Ensure the backlog and matrix are fully aligned on Slice 6's scope.
- **Historical commit preservation:** When Slices 6 and 7 restore/replace the missing proofs, do the corresponding rows in `REQUIREMENTS.md` retain their historical commit citations as ancestry evidence, or are those commits retired? We recommend retaining them as historical support notes while appending the new live test paths.

---

## Required Changes

1. **Add the missing scenario rows** (`SESSION-LIFE-001`, `SESSION-LIFE-002`, `SESSION-LIFE-006`, `SESSION-LIFE-008`, `SESSION-ID-001`, `SESSION-ID-002`, and `SESSION-RUNTIME-004`) to their respective slices in the `## Scenario Traceability Matrix`.
2. **Define the behavior-change approval authority** in TR-001 to prevent unauthorized modifications to the requirements ledger.
3. **Tighten characterization test criteria** in TR-001 to mandate black-box, old-vs-new dual execution tests rather than white-box mock assertions.
4. **Annotate stale rows directly in `REQUIREMENTS.md`** to indicate that `scale_from_zero_test.go`, `provider_health_gate_test.go`, and `session_progress_test.go` are currently missing from the checkout.
5. **Add exact test function verification** to the Slice 0 proof-freshness gate to prevent tests from silently going stale or being skipped.
