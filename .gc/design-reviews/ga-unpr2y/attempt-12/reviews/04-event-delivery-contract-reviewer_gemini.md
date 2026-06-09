# Amara Osei — Gemini (Independent Review, Attempt 12)

**Verdict:** block

**Review focus:** Factual session events, idempotent subscribers, crash recovery backstop, close/work-release guarantee, successor safety, and event identity context consistency. Evaluated against the Attempt 12 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-12/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 12 revision of `internal/session/DESIGN.md` maintains a robust architectural posture by decoupling safety-critical event delivery from at-most-once in-process event emission. Formalizing the `work_release_pending=true` state, a dedicated release identity snapshot, and mapping critical versus best-effort subscriber classifications (lines 1186–1194) represent excellent distributed-systems practices.

However, from the perspective of the **Event Delivery Contract Reviewer**, several load-bearing recovery mechanisms remain purely theoretical or contain structural ordering violations. The active checkout continues to lack any executable Slice 0 test files (such as `slice0_artifact_test.go`), guard fixtures, or mock crash-recovery tests. Furthermore, a detailed review of the operation sequence reveals a critical "intent-before-side-effect" ordering violation in the Close command, as well as a complete lack of timeouts for the transition of draining states.

Until these blocking issues are resolved in the design and proven in code, decomposition must remain blocked.

---

## Top Strengths

1. **Durable Scans as the Resilient Backstop:** Explicitly stating that "no safety-critical reaction may depend only on at-most-once in-process event delivery" (lines 1208–1210) and that "durable scans are the mandatory backstop" is an exemplary distributed-systems design choice. This protects against lost, delayed, or duplicated event delivery after crashes.
2. **Preventing Orphaned Work on Crash:** Writing a transitional `work_release_pending=true` marker and a release identity snapshot before clearing live identity fields prevents a crash from leaving old work permanently stranded.
3. **Structured Reaction Matrix:** The subscriber classification table (lines 1186–1194) clearly segregates critical retryable reactions (work release, runtime reconciliation) from best-effort, observability-only concerns (SSE, trace, dashboard), establishing clear architectural boundaries.

---

## Critical Risks & Blockers

### 1. [Blocker] Continued Lack of Executable Slice 0 Recovery Code and Test Fixtures
* **Evidence:** `internal/session/` (active checkout)
* **Pattern Comparison:** `internal/session/lifecycle_projection_test.go` and `manager_test.go` contain robust synchronous tests but no crash-recovery or restart test coverage.
* **Why it matters:** The design requires comprehensive crash recovery proofs (including crash-after-close-commit, skipped event emission, duplicate scanner pass, partial-query retry, stale-successor, and terminal identity-retirement tests) before any external close path delegates. Yet, there are still no physical test files (such as `slice0_artifact_test.go` or equivalent crash/recovery tests) in the active checkout. Relying on prose to solve safety-critical distributed-systems crash recovery guarantees is a direct violation of TDD and GUPP principles. Without physical, runnable test cases that simulate process crashes, we cannot verify that the system converges correctly.
* **Required Fix:** Implement and commit the skeleton for `internal/session/slice0_artifact_test.go` and at least one robust integration/unit test simulating a crash-after-close-commit and successful backstop recovery before approving decomposition.

### 2. [Blocker] Critical "Intent-Before-Side-Effect" Ordering Violation in Close Operation
* **Evidence:** `internal/session/DESIGN.md:1190`
* **Why it matters:** The operation table defines the close sequence as: `provider stop success when required -> commit close/retire facts -> release work scan -> event/trace`. This executes the side-effect (provider stop) *before* committing the close intent or facts to the durable store. If the controller process crashes immediately after the provider stops but before the close/retire facts are committed, the database will still show the session as `active`. On reboot, the reconciler will query the database, see an active session with a missing runtime (since the provider was stopped), and under `SESSION-LIFE-005` will blindly attempt to restart/respawn the provider process. This completely overrides the user's explicit close request and causes an infinite resurrection loop of a closed session.
* **Required Fix:** Re-order the Close sequence to commit a transitional closing intent to the store *before* executing the provider stop side-effect: `commit closing intent (state=closing) -> provider stop side-effect -> commit closed/retired terminal state -> release work`.

### 3. [Blocker] Missing Drain Timeouts and Dangling `draining` State
* **Evidence:** `internal/session/DESIGN.md:1191`
* **Why it matters:** The drain operation relies on `commit drain intent -> worker ack -> reread assigned work -> complete/cancel` with recovery authority `Drain metadata plus live work query`. However, the design does not specify any timeout or lease duration for drain intents. If a worker process crashes, stalls, or fails to ack the drain, the session will remain trapped in the `draining` state indefinitely. There is no automated timeout, cleanup, or lease-expiration mechanism to recover a dangling drain.
* **Required Fix:** Introduce an explicit `drain_timeout` or lease duration in the drain intent metadata, and specify that the reconciler sweeps or forces closure of sessions stuck in `draining` past this threshold.

### 4. [Major] Gaps in "Release Identity Snapshot" Schema Specification
* **Evidence:** `internal/session/DESIGN.md:1166-1167`
* **Why it matters:** The design mentions writing a "release identity snapshot" before clearing live identity fields but still does not define the exact schema, serialization format, or metadata keys. On last-writer-wins backends using flat metadata key-value stores (like `beads`), this lack of structure leads to pattern drift, where different subsystems or developers use different keys (e.g. JSON string blobs vs flat prefixed keys), leading to decoding errors during recovery scans.
* **Required Fix:** Formally specify the exact metadata keys for the release identity snapshot in `DESIGN.md` (e.g. `gc.release.session_name`, `gc.release.alias`, `gc.release.generation`).

### 5. [Minor] Lack of Recovery Owner for Mail/Extmsg Binding Cleanup
* **Evidence:** `internal/session/DESIGN.md:1192-1193`
* **Why it matters:** The subscriber table still classifies Mail/extmsg binding cleanup as "retryable" but does not explicitly assign a recovery authority or periodic sweep routine, leaving a loophole for orphaned bindings after a crash.
* **Required Fix:** Formally document that the controller's periodic reconciler scan is the recovery authority for mail/extmsg bindings, or state that missed cleanups are acceptable best-effort.

---

## Answers to Persona Questions

### 1. Are SessionEvent payloads facts with stable identifiers rather than commands to work, mail, trace, or extmsg?
**Answer:** Yes. The updated design explicitly dictates that any added wire payloads must describe facts that happened (not commands to attempt) and carry stable IDs, generation/instance tokens, and idempotency keys (lines 1202–1204). Thin hints are strictly used as accelerators to trigger durable scans.

### 2. If the process crashes after a durable session mutation but before in-process event delivery, how does a safety-critical subscriber such as work release converge?
**Answer:** It converges via the `work_release_pending` scanner. The scanner—owned by the controller/reconciler—periodically queries closed session facts and open work. If a session bead is closed with `work_release_pending=true` and its release identity snapshot is present, the scanner releases the work and clears the pending flag, guaranteeing convergence even after a crash (lines 1168–1175).

### 3. Which reactions are critical versus best-effort, and is the recovery path documented for each?
**Answer:** Yes. The "Subscriber classification" table (lines 1186–1194) explicitly categorizes Work release, Runtime start, Explicit wake, and Runtime reconciliation as "critical retryable" with durable scan recovery. Mail/extmsg binding cleanup is categorized as "retryable", and Trace/SSE/dashboard is classified as "observability-only" (best-effort).

---

## Consistency Report

* **Pattern Alignment:**
  - **Mutation Boundary Auditor (Elena Marchetti):** Our findings strongly align with Elena's concerns. The lack of physical, committed Slice 0 code files and test skeletons in the checkout is a shared blocker that prevents both the static mutation guards and the event recovery tests from being verified.
  - **Decider Atomicity Enforcer (Takeshi Yamamoto):** Takeshi's block on Pure Decider clock reads and unsafe concurrency is perfectly matched by our blocker regarding "Intent-Before-Side-Effect" ordering. Both reviews highlight that when a process crashes, the in-memory state is lost, making it impossible to perform safe recovery without durable intent committed beforehand.
* **Cross-File Integrity:**
  - The proposed `work_release_pending` metadata state and release identity keys must be integrated into `internal/session/lifecycle_projection.go` once Slice 0 lands, ensuring they are not classified as unknown fields.
