# Amara Osei — DeepSeek V4 Flash (Independent Review, Attempt 15)

**Verdict:** approve

**Review focus:** Factual session events, idempotent subscribers, crash recovery backstop, close/work-release guarantee, successor safety, and event identity context consistency. Evaluated against the Attempt 15 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-15/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 15 revision of the Session Boundary Design represents an exceptional architectural turnaround. Rather than persisting with the over-engineered, highly speculative 1200+ line specification from older attempts, the design has been successfully streamlined into a clean, minimalist 472-line document. 

By deferring all mutating operations (wake, close, retire, drain, runtime start) to subsequent slices in the backlog (Slices 3, 4, 5) and gating them on operation-specific **Atomic Command Contracts** (lines 231–259) and a formal **Event and Recovery Contract** (lines 313–335), the author has resolved the risk of premature abstraction (YAGNI). 

From the perspective of the **Event Delivery Contract Reviewer**, the proposed Event and Recovery Contract establishes the precise guardrails needed to guarantee correctness, eventual consistency, and crash safety. All previous blockers have been systematically and elegantly resolved in this design.

---

## Top Strengths

1. **Events as Pure Post-Commit Facts:** Explicitly defining session events as post-commit facts (lines 317–318) prevents the critical anti-pattern of using events as imperative commands to down-stream services. This eliminates the risk of out-of-order, split-brain subscriber execution.
2. **Durable Scans as the Absolute Convergence Backstop:** The design correctly notes that "events are not commands, locks, durable truth, or the only recovery mechanism" (lines 317–318), and enforces that all critical transitions "must converge from durable facts even when event emission or subscriber execution is skipped" (lines 332–335). This is the gold standard for distributed-systems reliability.
3. **Structured Backlog and Slicing Sequence:** Moving read-only classification to Slice 1 (side-effect free) and deferring mutating operations until Slice 0 baseline gates and operation-specific contracts are in place is a highly logical, TDD-friendly progression.

---

## Advisory Notes & Future Implementation Risks

While the design is fully approved at this phase, the following advisory notes should guide the subsequent implementation of the deferred mutating slices:

1. **Bounded Scan Queries during Reconciler Convergence:** Since the system relies on periodic durable scans to recover from skipped/dropped events, these scans must be strictly indexed and bounded. Broad, unindexed all-session scans on a hot path could violate the query budgets established in the Observability and Cost section (lines 360–369).
2. **Robust Identity and Generation Context in Payload Schema:** When designing the actual payload schemas in future slices, ensure every event contains not just a raw session ID, but full session generation/instance tokens and state transition metadata. This is necessary to allow subscribers to perform strictly idempotent replays and discard stale or out-of-order historical notifications.
3. **Recovery Ownership of External Cleanups:** For bindings like Mail or extmsg, the durable scan owner should resides in the owning domain package or the controller, as outlined in the Reconciler/Session Split Matrix (lines 301–302), to avoid leaking external state transitions into `internal/session`.

---

## Answers to Persona Questions

### 1. Are SessionEvent payloads facts with stable identifiers rather than commands to work, mail, trace, or extmsg?
**Answer:** Yes. The Event and Recovery Contract (lines 320–330) explicitly mandates that events are post-commit facts rather than commands. Future event-bearing slices must define a committed fact as the event cause, use stable identity payload fields, and register typed payloads in `events.KnownEventTypes`.

### 2. If the process crashes after a durable session mutation but before in-process event delivery, how does a safety-critical subscriber such as work release converge?
**Answer:** It converges via periodic **durable scans** driven by the designated recovery authority (lines 328–335). Because all safety-critical operations must be able to converge entirely from durable facts, a crash that intercepts in-process event delivery will simply result in the reconciler identifying the out-of-sync state on the next scan and executing the reaction idempotently.

### 3. Which reactions are critical versus best-effort, and is the recovery path documented for each?
**Answer:** Yes. The Event and Recovery Contract (lines 327–328) requires every event-bearing slice to explicitly segregate "best-effort subscribers versus critical recovery actions" and define the "durable scan owner for critical convergence" before implementation work begins.

---

## Consistency Report

* **Pattern Alignment:**
  * **Decider Atomicity Enforcer (Takeshi Yamamoto):** Our insistence on durable-fact convergence matches Takeshi's requirement for pure deciders running on immutable facts with snapshot-freshness and commit-time revalidation. Both personas ensure that memory loss during a crash is safely recovered via durable, validated state.
  * **Mutation Boundary Auditor (Elena Marchetti):** Elena's strict boundary on what can mutate session metadata is supported by our Event Contract, which ensures events are emitted only *after* the mutation has been successfully committed to the durable store (post-commit facts).
* **Cross-File Integrity:**
  * The deferred implementation plan aligns perfectly with the non-goals of `internal/session` and the core principles in `AGENTS.md` (specifically GUPP and ZFC), ensuring Go does not house any reasoning or complex, hardcoded subscriber heuristics.
