# Amara Osei — Gemini (Independent Review, Attempt 10)

**Verdict:** block

**Review focus:** Factual session events, idempotent subscribers, crash recovery backstop, close/work-release guarantee, successor safety, and event identity context consistency. Evaluated against the Attempt 10 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-10/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 10 revision of the Session Boundary Design introduces highly commendable conceptual upgrades to the event delivery and recovery architecture. Specifically, the introduction of the `work_release_pending=true` transitional state, a dedicated release identity snapshot, and a named controller scanner helper in `DESIGN.md:1105-1129` directly addresses the key gaps identified in the Attempt 8 review. This design ensures that a process crash after a close commit but before inline cleanup does not permanently strand assigned work.

However, from the perspective of the **Event Delivery Contract Reviewer**, these safety-critical recovery mechanisms remain purely theoretical. There are no executable Slice 0 test files (such as `slice0_artifact_test.go`), guard fixtures, or mock crash-after-commit tests in the active checkout. The design document itself explicitly admits this on lines 83–88:
> This document therefore continues to block decomposition until those artifacts exist; it does not mark those concerns resolved by prose.

Because a robust event delivery contract requires physical, executable proof of recovery convergence rather than prose guarantees, we must sustain a **block** on decomposition.

---

## Top Strengths

1. **Decoupling Event Delivery from Convergence:** Stating that "no safety-critical reaction may depend only on at-most-once in-process event delivery" (lines 1074–1075) and that "durable scans are the mandatory backstop" (line 1079) is a premium, resilient design pattern that respects Zero Framework Cognition (ZFC) and NDI.
2. **Elimination of the Clear-then-Release Race:** Writing a transitional `work_release_pending=true` marker and a release identity snapshot before clearing live identity fields prevents a crash from leaving old work unreleasable and orphaned.
3. **Comprehensive Per-Event Reaction Matrix:** The per-event reaction table (lines 1157–1167) clearly maps required durable fact fields, idempotency keys, and recovery authorities for all major events, ensuring that subscribers can handle duplicate or skipped events safely.

---

## Critical Risks & Blockers

### 1. [Blocker] Lack of Executable Slice 0 Recovery Code and Test Fixtures
* **Evidence:** `design-before.md:1125–1129`
* **Pattern Comparison:** `internal/session/manager_test.go` has robust synchronous tests but no crash-recovery or restart test coverage.
* **Why it matters:** The design mandates tests for "crash-after-close-commit, skipped event emission, duplicate scanner pass, partial-query retry, stale-successor, and terminal identity-retirement" before external close paths delegate. However, because no Slice 0 prototype code or test skeletons currently exist in the repository, we cannot verify that the AST parser or unit testing framework can physically simulate or check these conditions. Approving decomposition on prose alone risks introducing severe integration defects where a crash leaves work stranded.
* **Suggested Fix:** Implement and commit the skeleton for `internal/session/slice0_artifact_test.go` and basic crash-after-commit test cases under `cmd/gc` or `internal/session` before approving this gate.

---

### 2. [Blocker] Underspecified Schema for the "Release Identity Snapshot"
* **Evidence:** `design-before.md:1110–1111`
* **Pattern Comparison:** `internal/session/lifecycle_projection.go` has highly structured parsing for metadata keys.
* **Why it matters:** The design requires writing a "release identity snapshot" before clearing live identity fields, but does not define the exact metadata keys or the serialization format for this snapshot. Without an explicit schema in the design, developers are likely to use mismatched structures (e.g., flat metadata keys vs. JSON string blobs), causing decoding failures or pattern drift during recovery scans.
* **Suggested Fix:** Explicitly define the metadata keys for the release identity snapshot in the design (e.g., `gc.release_identity.session_name`, `gc.release_identity.alias`, `gc.release_identity.generation`) to ensure complete cross-file consistency.

---

### 3. [Major] Stale Backlog Wording Dilutes the Scanner Recovery Contract
* **Evidence:** `design-before.md:1368`
* **Why it matters:** The coexistence gate table for Slice 4 (line 1368) states: "Work-release scan may stay outside session but must consume committed close facts." This is weaker than the main text's requirement for a dedicated `work_release_pending` scanner owned by the controller. Mismatched requirements in the backlog tables can lead to developers skipping the implementation of the robust `work_release_pending` backstop in favor of simpler, less reliable post-commit fact listeners.
* **Suggested Fix:** Align the Slice 4 coexistence gate in line 1368 with the main text by explicitly requiring that the work-release scan is backed by the `work_release_pending` scanner.

---

### 4. [Minor] Ownership of Mail/Extmsg Binding Cleanup is Incomplete
* **Evidence:** `design-before.md:1179`
* **Why it matters:** The subscriber classification table lists Mail/extmsg binding cleanup as "retryable" and states that "event accelerates but is not sole authority", but does not assign a clear recovery owner or specify a periodic sweep cadence for missed notifications, leaving a small loophole of orphaned bindings.
* **Suggested Fix:** Explicitly state in the table that the controller or a named repair helper is the recovery authority for mail/extmsg bindings, or formally classify missed notifications as accepted best-effort.

---

## Answers to Persona Questions

### 1. Are SessionEvent payloads facts with stable identifiers rather than commands to work, mail, trace, or extmsg?
**Answer:** Yes. The updated design explicitly dictates that any added wire payloads must describe facts that happened (not commands to attempt) and carry stable IDs, generation/instance tokens, and idempotency keys (lines 1146–1148). Thin hints are strictly used as accelerators to trigger durable scans.

### 2. If the process crashes after a durable session mutation but before in-process event delivery, how does a safety-critical subscriber such as work release converge?
**Answer:** It converges via the `work_release_pending` scanner. The scanner—owned by the controller/reconciler—periodically queries closed session facts and open work. If a session bead is closed with `work_release_pending=true` and its release identity snapshot is present, the scanner releases the work and clears the pending flag, guaranteeing convergence even after a crash (lines 1112–1124).

### 3. Which reactions are critical versus best-effort, and is the recovery path documented for each?
**Answer:** Yes. The "Subscriber classification" table (lines 1173–1181) explicitly categorizes Work release and Runtime reconciliation as "critical retryable" with durable scan recovery. Mail/extmsg binding cleanup is categorized as "retryable", and Trace/SSE/dashboard is classified as "observability-only" (best-effort).

---

## Consistency Report

* **Pattern Alignment:**
  - Our findings strongly support the **Mutation Boundary Auditor (Elena Marchetti)**. The lack of physical, committed Slice 0 code files and test skeletons in the checkout is a shared blocker that prevents both the static mutation guards and the event recovery tests from being verified.
* **Cross-File Integrity:**
  - The proposed `work_release_pending` metadata state and release identity keys must be integrated into `internal/session/lifecycle_projection.go` once Slice 0 lands, ensuring they are not classified as unknown fields.
