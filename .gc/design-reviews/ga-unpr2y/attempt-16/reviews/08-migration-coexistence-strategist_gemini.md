# Ravi Krishnamurthy — DeepSeek V4 Flash (Independent Review)

**Verdict:** block

**Lane:** Migration sequencing, legacy-new coexistence, rollback slices, and worker-boundary collision.

---

## Overview

As the **08-migration-coexistence-strategist** (Ravi Krishnamurthy), this independent review evaluates the current Attempt 16 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-16/design-before.md`) against `internal/session/REQUIREMENTS.md`, the architectural invariants in `AGENTS.md`, and the active checkout source of the `gascity` repository.

Attempt 16 introduces several excellent additions, specifically the new **Migration Coexistence And Rollback** section, a rigorous schema requirement for `BOUNDARY_MATRIX.yaml`, and a defined event/cost budget mapping. These additions are massive steps toward managing the extremely high risk of split-brain writes and concurrent corruption during incremental adoption.

However, from the perspective of **migration safety, write coexistence, rollback isolation, and worker-boundary integrity**, the design remains a **BLOCK**. It continues to defer critical concurrency decisions to individual implementation slices, tolerates compile-time test exceptions for the worker-boundary, and introduces highly risky rollback behaviors.

---

## Top Strengths

1. **Structured Migration Matrix (DESIGN.md:434-447):**
   Requiring each slice to define predecessor/successor slices, field-family ownership transitions, and old/new reader-writer compatibility before touching overlapping files is an outstanding defensive practice. It makes the "strangler migration" explicit and auditable.
2. **Explicit Event and Cost Budgets (DESIGN.md:521-528, 576-584):**
   Mandating an event taxonomy inventory for Slice 0 (including critical events like `session.woke`, `session.stranded`, etc.) and defining strict query/scanned-row limits for the hot path guarantees that the migration will not introduce performance regressions or unmodeled log spam.
3. **Purity of read-only Classifier (DESIGN.md:210-219):**
   Strictly prohibiting side-effects (no store writes, no materialization, no events) on the first adopter resolver ensures a clean, non-disruptive first step for Slice 1.

---

## Critical Risks & Coexistence Challenges

### 1. [Blocker] Deferral of physical cross-process concurrency standard
* **Evidence:** `DESIGN.md:441-442` (`cross-process fence when old and new writers coexist`) and `DESIGN.md:345-349` (`No slice may leave both a legacy raw writer and a new session-owned command writing the same field family unless the row names the fence...`).
* **Why it matters:** While the design strictly forbids unfenced writers, it **fails to specify the standardized fencing mechanism itself** at the architectural level. It pushes the choice of the concurrency fence onto individual slices.
  
  If Slices 3, 4, and 5 are left to define their own independent fencing mechanisms (e.g., custom file locks, custom Dolt-level metadata tag checks, or store-level hooks), we will end up with highly fragmented and inconsistent database concurrency logic across different files. Since the CLI and reconciler daemon run as completely separate OS processes, they have entirely independent memory spaces and cannot see each other's in-memory mutexes.
  
  A single, unified, and reusable cross-process concurrency primitive (such as cooperative file locks or store-level atomic metadata Compare-And-Swap) must be defined as a core Layer 0/1 SDK capability and mandated globally, rather than being treated as a slice-level design task.

### 2. [Blocker] High-Risk Rollback without State-Version Epochs
* **Evidence:** `DESIGN.md:443-446`:
  > `- rollback data direction: whether rollback preserves new fields, clears them, or runs a repair/backfill command`
  > `- tests proving old readers tolerate new fields and new readers tolerate old fields during rollback`
* **Why it matters:** Relying on ad-hoc "repair/backfill commands" during a rollback introduces a highly volatile window for silent data corruption. If a rolled-back system contains active, running sessions with mixed metadata (some written by the new command before rollback, some by the old legacy writers after rollback), the behavior is completely non-deterministic.
  
  The design lacks a clear **State Versioning/Migration Schema epoch marker** on the session bead. Without a monotonically increasing state-version tag on the session bead metadata, old and new readers/writers have no way to reliably detect and reject multi-version state mismatches, making safe rollbacks of active sessions impossible.

### 3. [Blocker] Worker-Boundary routing and exception loophole
* **Evidence:** `DESIGN.md:415-417` vs `DESIGN.md:402-406`:
  > `Wake and drain default to worker.Handle when exposed through production API or CLI mutation paths. A store-level route requires a root-approved exception row before delegation...`
* **Why it matters:** Permitting "store-level route" exceptions for wake/drain mutations bypasses the `worker.Handle` interface. This is a massive loophole that:
  1. Compromises the compile-time worker-boundary import guard (`TestGCNonTestFilesStayOnWorkerBoundary` in `cmd/gc/worker_boundary_import_test.go`), which strictly forbids direct imports of session managers.
  2. Bypasses the critical background cleanup, wait cancellations, and provider interactions that `worker.Handle` handles during session state transitions.
  
  Routing through `worker.Handle` must be **mandatory** for all mutating commands in production CLI/API code. Store-level bypass exceptions for mutations must be entirely prohibited to maintain worker-boundary integrity.

### 4. [Major] Availability Degradation on `RepairEmptyType` Quarantine
* **Evidence:** `DESIGN.md:246-252`:
  > `RepairEmptyType is not allowed in the read-only classifier path. When current helpers would repair an empty-type session bead as a side effect, the new classifier returns repair-needed with the same match vector and an audited repair command owns the write...`
* **Why it matters:** Currently, query-side handlers silently repair empty-type session beads on 12 different sites to keep the system running smoothly. Under the new design, when a read path finds an empty-type bead, it returns `repair-needed` and delegating to a separate, audited repair command.
  
  However, the design does not specify *when* and *how* this repair command is executed. If a query handler (like API Get Session) blocks and returns a 404/500 when encountering `repair-needed`, we introduce a severe availability regression for previously healthy sessions. If it triggers the repair synchronously, then the read-only path is no longer read-only, violating the classifier's side-effect-free contract. The execution lifecycle and non-blocking recovery path for empty-type repairs must be authoritatively defined.

---

## Answers to Persona Questions

### 1. How does the plan sequence this extraction with the in-flight worker-boundary migration on overlapping `cmd/gc` and `internal/api` call sites?
The plan sequences them by requiring that **all mutating session operations (wake, close, runtime start) are invoked exclusively via `worker.Handle` for production CLI/API code**. Read-only targets (like API query resolution) are isolated under the read-only target classifier. However, the loophole allowing "store-level route exceptions" for wake/drain in `WORKER_BOUNDARY_EXCEPTIONS.yaml` collides with the worker-boundary's compile-time import tests and must be removed.

### 2. During partial adoption, what prevents legacy patch-map callers and new command callers from split-brain writes to the same metadata fields?
The design introduces a `Migration Coexistence And Rollback` row to map field-family ownership and enforce a "cross-process fence." However, the lack of a standardized, globally reusable cross-process synchronization library and a state-version epoch tag on session beads means that split-brain writes cannot be physically prevented. Without these primitives, concurrent processes (CLI and reconciler daemon) will continue to race during partial adoption.

### 3. Which single slice is independently shippable and revertible, and what proves it does not silently require the next slice?
Only Slice 1 (API query target classification) is independently shippable and revertible. Because it is strictly read-only, side-effect-free, and handles empty-type repairs via a non-blocking `repair-needed` status without mutating the bead, it has zero data-corruption risks. Any rollback of Slice 1 simply reverts the query adapter back to `session.ResolveSessionID`. All subsequent slices (3, 4, 5) are highly coupled due to shared file modifications in the reconciler and CLI, and cannot be rolled back without cascading risks.

---

## Required Changes

1. **Standardize a Global Cross-Process Concurrency Primitive:**
   Define a unified, reusable concurrency primitive at the Layer 0/1 level (such as cooperative file locks or store-level atomic Compare-And-Swap) rather than deferring the cross-process fence implementation to individual slices.
2. **Implement State-Version Epoch Tags on Beads:**
   Add a mandatory `schema_version` or state-version epoch tag to the session bead metadata to allow old and new readers/writers to safely detect, reject, or handle multi-version state mismatches during rollbacks.
3. **Eliminate Store-Level Mutation Exceptions:**
   Remove the "store-level route exception" loophole for wake and drain. Mandate that **all** mutating session operations in production API/CLI code must route exclusively through `worker.Handle` to maintain the integrity of `TestGCNonTestFilesStayOnWorkerBoundary`.
4. **Define the Repair Execution Lifecycle:**
   Authoritatively define how the "audited repair command" is invoked when the read path encounters a `repair-needed` status to prevent read-side write side-effects while avoiding severe query availability regressions.
