# Natasha Volkov — Gemini (Independent Review, Attempt 8)

**Verdict:** block

**Review focus:** REQUIREMENTS scenario parity, regression prevention, characterization tests, and proof freshness — with evidence drawn from codebase verification.

---

## Overview

The Attempt 8 iteration of the session boundary design makes substantial structural improvements. Elevating Slice 0 to a hard scheduling boundary—requiring source-complete symbol scans, executable static guards, and `SCENARIO_PARITY.yaml` before any mutation-owning work can be decomposed—is an excellent and highly rigorous design response. The "Current row ownership baseline" table also correctly enumerates all 45 scenario rows defined in the requirements ledger.

However, from the perspective of the Behavior Parity Guardian, the design document itself contains critical inconsistencies, unmapped requirements, and stale references that must be resolved before the design is approved for decomposition. Specifically, the actual **Scenario Traceability Matrix** covers only 36 unique requirements, leaving exactly 9 critical scenario rows completely unmapped. Furthermore, `REQUIREMENTS.md` continues to cite test files that have been deleted from `HEAD`, violating its own maintenance rules. 

Until these gaps are resolved, we cannot guarantee that developers will capture full behavioral parity or prevent regression during the extraction phase.

---

## Top Strengths

- **Slice 0 Hard Scheduling Boundary:** Elevating Slice 0 to a strict, non-negotiable entry gate prevents "decomposition creep" and ensures that static guards, AST scanners, and Yaml registries are fully established before any implementation begins.
- **Durable Scan Over Best-Effort Events:** The shift toward treating in-process event delivery as best-effort/latency aids and placing the critical convergence guarantee on durable reconciler fact scans is an exceptionally sound architectural decision that aligns with NDI (Nondeterministic Idempotence).
- **Clear Write-Safety Classification:** Categorizing write primitives (`Update` closures, `SetMetadataBatch`, etc.) and defining explicit compensation or repair paths for tokened blind writes provides a highly defensive mutation model.
- **Complete Baseline Enumeration:** The newly added "Current row ownership baseline" in `DESIGN.md:275–321` correctly lists all 45 `SESSION-*` rows from `REQUIREMENTS.md`, showing an awareness of the entire behavioral surface area.

---

## Critical Risks

### [Blocker] The Scenario Traceability Matrix STILL leaves exactly 9 requirements rows completely unmapped

A mathematical audit of the actual **Scenario Traceability Matrix** (`DESIGN.md:913–922`) reveals that it covers only 36 unique requirements out of the 45 listed in `REQUIREMENTS.md`. Exactly 9 scenario rows are completely omitted from the Matrix. A developer decomposing beads using only the Traceability Matrix as their source would have no obligation to implement or prove parity for these 9 critical behaviors:

1. **`SESSION-LIFE-001`** (Legacy compatibility states are projected)
2. **`SESSION-LIFE-002`** (Pending create claim)
3. **`SESSION-LIFE-006`** (Missing config)
4. **`SESSION-LIFE-008`** (User-facing projection guard)
5. **`SESSION-ID-001`** (Explicit session names)
6. **`SESSION-ID-002`** (Aliases)
7. **`SESSION-RECON-001`** (Worker boundary)
8. **`SESSION-WORK-003`** (Orphan pool step beads)
9. **`SESSION-RUNTIME-004`** (Stop turn)

The author of Attempt 8 attempted to resolve this in the prose "Current row ownership baseline" table (`DESIGN.md:275–321`) by classifying several of these as "baseline" (e.g., `SESSION-LIFE-001` as `lifecycle projection baseline`). However, because these baseline states/behaviors must remain unbroken throughout the migration, they must be explicitly mapped to the actual implementation slices in the Traceability Matrix as active validation constraints or secondary-slice verification targets. 

Specifically, leaving **`SESSION-RECON-001`** (Worker boundary) unassigned to any slice means no slice is actively accountable for ensuring that direct `session.Manager` construction bypasses are completely eliminated and replaced by `internal/worker/handle.go`. Similarly, leaving **`SESSION-WORK-003`** (Orphan pool step beads) unassigned means the boundary cleanup logic has no implementation or test owner.

**Required change:** Assign all 9 unmapped rows to appropriate slices in the Scenario Traceability Matrix (`DESIGN.md:913–922`):
- Map `SESSION-LIFE-001` and `SESSION-LIFE-008` as cross-cutting invariants across Slice 3 (Runtime start) and Slice 5 (Eligibility) to ensure projection helpers are used.
- Map `SESSION-LIFE-002` to Slice 3 (Runtime start prepare/commit/rollback) or Slice 6 (Reconciler facts).
- Map `SESSION-LIFE-006` to Slice 5 (Wake/hold/drain eligibility) or Slice 6 (Reconciler facts).
- Map `SESSION-ID-001` and `SESSION-ID-002` to Slice 1 (Target classification).
- Map `SESSION-RECON-001` to Slice 1 or Slice 4 (Close/Retire) to assert worker boundary routing is maintained.
- Map `SESSION-WORK-003` to Slice 4 (Close and identity retirement) or Slice 5 (Eligibility).
- Map `SESSION-RUNTIME-004` to Slice 5 (Wake/hold/drain eligibility) or Slice 7 (Provider health/progress).

---

### [Blocker] REQUIREMENTS.md STILL cites deleted test files as active proofs of behavior

The requirements ledger (`REQUIREMENTS.md`) continues to cite several test files that have been deleted from `HEAD` and do not exist in the checkout:
- `cmd/gc/scale_from_zero_test.go` (cited by `SESSION-RECON-002` and `SESSION-RECON-003`)
- `cmd/gc/provider_health_gate_test.go` (cited by `SESSION-RECON-006`)
- `cmd/gc/session_progress_test.go` (cited by `SESSION-RECON-007`)

While the design response in `DESIGN.md:323–326` says: *"Rows whose cited proof files are missing are blocked... The implementation must restore the deleted proof, replace it with an equivalent current test, or update REQUIREMENTS.md... before extraction,"* the actual `REQUIREMENTS.md` file has not been annotated or modified to reflect this. This violates the ledger's own Maintenance Rules (`REQUIREMENTS.md:159`): *"Keep evidence current. If a cited test is deleted, move or replace the row."* 

Allowing unannotated deleted citations to remain in the canonical behavior registry is dangerous, as future reviewers or developers reading the requirements ledger will assume these behaviors are actively covered when they are completely blind.

**Required change:** Annotate these four stale rows in `REQUIREMENTS.md` with warnings indicating that their cited tests are absent from the active checkout, and add cross-references to the corresponding `DESIGN.md` blocking notes (Slices 6 and 7) so readers can trace the dependency.

---

### [Blocker] Discrepancies between the prose "Row Ownership Baseline" and the "Scenario Traceability Matrix"

There are several outright contradictions between where a requirement is assigned in the prose "Row Ownership Baseline" table (`DESIGN.md:275–321`) and where it is assigned in the actual "Scenario Traceability Matrix" (`DESIGN.md:913–922`):

- **`SESSION-LIFE-002` (Pending create claim):** Baseline table maps it to `runtime-start and reconciler eligibility`. Traceability Matrix completely omits it from both Slice 3 (Runtime start) and Slice 6 (Reconciler).
- **`SESSION-LIFE-006` (Missing config):** Baseline table maps it to `target/config eligibility`. Traceability Matrix completely omits it from Slice 1 (Target classification) and Slice 5/6 (Eligibility/Reconciler).
- **`SESSION-ID-007` (Terminal named identity wake):** Baseline table maps it to `wake and close/retire`. Traceability Matrix maps it to Slice 1 (Target classification) and Slice 4 (Close and retirement), but completely omits it from Slice 2 (Explicit wake).
- **`SESSION-WORK-003` (Orphan pool step beads):** Baseline table maps it to `work recovery scan`. Traceability Matrix completely omits it.
- **`SESSION-RUNTIME-004` (Stop turn):** Baseline table maps it to `submit/provider shell`. Traceability Matrix completely omits it.

These contradictions create severe ambiguity regarding which slice actually owns the tests and parity verification for these behaviors.

**Required change:** Reconcile these tables so that every assignment in the Baseline table has an exact, matching slice assignment in the Scenario Traceability Matrix.

---

## Major Risks

### [Major] The canonical production writer inventory still misses key production files

The "Canonical Production Writer Inventory" (`DESIGN.md:700–733`) STILL fails to individually map key active production code files that write session-owned keys:
- **`cmd/gc/cmd_stop.go`** (writes `sleep_reason`)
- **`cmd/gc/cmd_wait.go`** (writes `wait_hold`, `sleep_intent`, `closed_at`, `close_reason`)

While these files are listed in a prose scanning list in `DESIGN.md:179–180`, they do not have their own unique inventory IDs (like `W-###`) or explicit owner slices in the actual writer inventory table. This means the static guard allowlist configuration cannot define shrink-only rules or explicit retirement plans for these writers, leaving a major blind spot in the mutation boundary.

**Required change:** Add explicit, granular rows in the writer inventory table for the specific writers in `cmd/gc/cmd_stop.go` and `cmd/gc/cmd_wait.go`. Assign them to their appropriate owner slices and specify their retirement conditions.

---

### [Major] Traceability and characterization gaps for Slice 5–7 due to missing live proof in Head

Slices 5, 6, and 7 heavily touch reconciler, pool scaling, and provider health behaviors where key characterization tests (like `scale_from_zero_test.go`, `provider_health_gate_test.go`, and `session_progress_test.go`) are missing from the current active checkout. The Traceability Matrix maps these rows to Slices 5–7 and cites these deleted files as "current proof" while simply adding a note: *"...is cited by requirements but absent in this checkout."*

This is highly problematic because these slices cannot be safely implemented or reviewed if there is no live characterization test in `HEAD` to assert the "before" behavior. Decomposing these slices without first restoring or replacing these tests is an invitation for silent behavioral drift.

**Required change:** State an explicit rule in `DESIGN.md` that Slices 5, 6, and 7 are strictly blocked from being decomposed or worked on until the missing tests are successfully restored to `HEAD` (or replaced with equivalent, passing unit/integration tests) under the Slice 0 preflight or a dedicated non-mutating test-restoration slice.

---

## Minor Risks

- **Underdetermined Multi-Slice Ownership:** Several rows are assigned to multiple slices in the Traceability Matrix (e.g., `SESSION-START-003` to Slices 2 and 5; `SESSION-START-008` to Slices 3 and 6; `SESSION-RUNTIME-001` to Slices 3 and 7). If multiple slices modify the code that impacts these rows, there is a risk of regression or conflicting assumptions.
  - *Required change:* For each multi-slice row, explicitly declare which slice is the **primary owner** (accountable for the core behavior and tests) and which slices are **secondary consumers** (accountable only for verifying no regression).
- **Inconsistent Revert Paths:** The "Slice Coexistence, Bead Metadata, And Revert Gate" (`DESIGN.md:442–470`) mentions that each slice must provide a per-surface revert path, but does not specify where the revert scripts or verification tests are stored.
  - *Required change:* Define a standard directory pattern (e.g., `internal/session/revert/slice-###/`) for rollback scripts and verification tests.

---

## Missing Evidence

- **Baseline Live Proofs:** There is no evidence of a live, passing test run in `HEAD` for `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`.
- **Granular Scanner AST Design:** There is no evidence demonstrating that the static guard AST parser can successfully identify and block direct writes to the `state` key within inline-map parameters or dynamic `setMetaBatch` wrappers.

---

## Required Changes

1. **Map all 9 unmapped scenario rows** (`SESSION-LIFE-001`, `SESSION-LIFE-002`, `SESSION-LIFE-006`, `SESSION-LIFE-008`, `SESSION-ID-001`, `SESSION-ID-002`, `SESSION-RECON-001`, `SESSION-WORK-003`, and `SESSION-RUNTIME-004`) to their respective owner slices in the actual Scenario Traceability Matrix (`DESIGN.md:913–922`).
2. **Annotate the stale/deleted test citations** in `REQUIREMENTS.md` to explicitly state they are absent from `HEAD`.
3. **Resolve all contradictions** between the "Row Ownership Baseline" table and the "Scenario Traceability Matrix" table.
4. **Expand the writer inventory table** to include explicit rows and IDs for `cmd_stop.go` and `cmd_wait.go`.
5. **Add a strict block on Slices 5–7** preventing any decomposition of these slices until the missing tests are restored to `HEAD`.
6. **Define primary vs. secondary ownership** for all multi-slice requirements rows in the Traceability Matrix.
7. **Define a standard directory pattern** for rollback scripts and revert verification tests.

---

## Questions

- Does the design intend the restored `scale_from_zero_test.go` and other missing tests to be merged back into `cmd/gc/` under Slice 0, or will they be restored as a separate, non-mutating Slice 0.5?
- Since `SESSION-RECON-001` (Worker boundary) is critical for enforcing that all CLI calls route through `worker.Handle`, should its validation be integrated into Slice 0's static-guard tests (`TestSessionBoundaryGuard`)?
- For `SESSION-WORK-003` (Orphan pool step beads), does the reconciler's recovery scan execute on a fixed cron cadence, or is it triggered dynamically on every reconciliation cycle?
