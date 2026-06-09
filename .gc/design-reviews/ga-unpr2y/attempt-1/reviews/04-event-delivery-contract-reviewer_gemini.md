# Amara Osei — DeepSeek V4 Flash (Independent Review, Attempt 16)

**Verdict:** approve-with-risks

**Review focus:** Factual session events, idempotent subscribers, crash recovery, work-release guarantee, successor safety, and event identity context consistency. Evaluated against the Attempt 16 iteration of `internal/session/DESIGN.md` (located at `.gc/design-reviews/ga-unpr2y/attempt-16/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source on this branch.

---

## Overview

The Attempt 16 revision of the Session Boundary Design continues to refine the architectural separation of concern. From the perspective of the **Event Delivery Contract Reviewer**, the proposed Event and Recovery Contract establishes critical boundaries needed to guarantee eventual consistency and crash safety. All core tenets are conceptually sound: events are strictly framed as post-commit facts (preventing imperative subscriber anti-patterns), and durable scans are established as the absolute convergence backstop (preventing split-brain states due to dropped in-process events).

However, while the model itself is structurally robust, a deep-dive grounding against the live codebase and cross-document requirements reveals several critical, latent risks in how these contracts are defined for Slice 0 and future event-bearing slices. Other reviewers may accept these risks too quickly, assuming that durable scans mitigate all event path weaknesses. If this SDK is to support reliable custom agent orchestration, we must address these gaps before proceeding.

---

## Top Strengths

1. **Post-Commit Fact Invariant:** Explicitly defining session events as post-commit facts (DESIGN.md:500) that are "not commands, locks, durable truth, or the only recovery mechanism" cleanly prevents the anti-pattern of using events as imperative commands. Limiting event `Message` text to operators and restricting subscriber consumption to typed payload fields and envelope fields (DESIGN.md:533-535) directly mitigates Red Flag #1.
2. **Durable Scans as the Absolute Convergence Backstop:** The non-negotiable rule that "Close, work release, wake, drain, runtime start... must converge from durable facts even when event emission or subscriber execution is skipped" (DESIGN.md:517-519) is the gold standard for distributed-system reliability. I verified that the current work release implementation matches this, converging through active and rig store reconciler sweeps in `cmd/gc/session_beads.go` / `cmd/gc/session_reconciler.go` rather than relying on in-process subscriber delivery.
3. **Structured Slicing Sequence:** Deferring mutating operations (Slices 3, 4, 5) and gating them on the completion of a non-mutating Slice 0 preflight ensures that all baseline schemas, validators, and negative tests exist before any production state is moved.

---

## Critical Risks & Gaps

### 1. The "NoPayload" Identity and Generation Blind Spot (Red Flag #3 Latent)
* **Risk:** The design requires events to carry "bead ID, session name, alias, configured identity, runtime session key, and generation/instance token" (DESIGN.md:507-508) to back idempotent replay. However, the live `Event` envelope (`internal/events/events.go:162-170`) carries only `Subject` as an identity field—there is no generation or instance token in the envelope.
* **Grounded Defect:** Under Slice 0, `session.stranded`, `session.draining`, `session.undrained`, `session.suspended`, and `session.woke` are registered as `events.NoPayload{}` in `internal/api/event_payloads.go:444`. Because they carry no typed payload, these events cannot carry a generation/instance token. 
* **Impact:** In multi-session and pool topologies (SESSION-ID-007, SESSION-WORK-002), slots and named session identities are regularly recycled and reused. A subscriber reacting to a `session.stranded` or `session.suspended` event gets only the `Subject` (session name) from the envelope, with **no generation context**. The subscriber cannot distinguish between an older, defunct generation and a newly spawned generation of the same identity. This exposes the system to cross-generation replay and stale-event interference if any reactive subscriber is attached—a critical safety risk that the scan backstop hides but does not solve.

### 2. Open-World Slice 0 Event Inventory
* **Risk:** DESIGN.md:521-526 inventories a minimum of 13 session events. However, a scan of the active event registry reveals additional registered `session.*` events, such as `session.created`, `session.deleted`, and `session.compacted` (with their corresponding request-result envelopes).
* **Impact:** For a robust event-delivery contract, the event space must be a closed-world inventory. Omitting `session.deleted` is highly risky, as it represents a terminal lifecycle fact that work-release or cleanup subscribers must react to. If any emitted session event remains unclassified, we cannot prove that uncontracted events will not reach subscribers and violate the safety rules.

### 3. "Generation/Instance Token When Relevant" is Too Permissive
* **Risk:** The contract qualifies the generation token requirement with "when relevant" (DESIGN.md:508). This introduces a subjective judgment call during implementation.
* **Impact:** For any lifecycle transition involving an identity that can be recycled, the generation token is **always** relevant for idempotent processing. Leaving this as conditional allows developers to omit it for safety-critical event payloads under the guise of simplicity.

### 4. Cross-Document Idempotency Key Inconsistency
* **Risk:** `REQUIREMENTS.md` (specifically under "Work Release And Drain Safety") mandates that "The release path is idempotent" (SESSION-WORK-001), but it does not specify how that idempotency is represented or verified. `DESIGN.md:530-533` requires the work-release convergence row to specify an "idempotency key," but does not explicitly define its structure.
* **Impact:** Without a strict, cross-document contract defining the idempotency key as **bead ID + generation**, separate implementations of the scan-driven and event-driven paths are highly likely to drift, leading to split-brain work-release conflicts on parallel ticks.

---

## Answers to Persona Questions

### 1. Are SessionEvent payloads facts with stable identifiers rather than commands to work, mail, trace, or extmsg?
**Answer:** Yes. The Event and Recovery Contract (DESIGN.md:500-501) explicitly establishes session events as post-commit facts. Furthermore, the design restricts subscribers to typed payload fields and envelope fields, while fencing event `Message` text to operator-only views (DESIGN.md:533-535). This prevents any attempt to embed imperative subscriber instructions in the event stream.

### 2. If the process crashes after a durable session mutation but before in-process event delivery, how does a safety-critical subscriber such as work release converge?
**Answer:** It converges via the **durable scan backstop** (DESIGN.md:517-519). The designated recovery authority (such as the controller reconciler scan in `cmd/gc/session_beads.go`) is responsible for periodically evaluating the live state of the store and driving convergence. Because the scan relies entirely on durable committed facts, memory loss during a crash is safely recovered on the subsequent scan tick, ensuring eventual consistency even if in-process event delivery fails entirely.

### 3. Which reactions are critical versus best-effort, and is the recovery path documented for each?
**Answer:** Conceptually yes. The design requires every event-bearing slice to segregate "best-effort subscribers versus critical recovery actions" and define the "durable scan owner for critical convergence" (DESIGN.md:512-513). However, the specific Slice 0 inventory does not yet explicitly document these relationships for the currently registered `session.*` events, representing a gap that must be closed before the Slice 0 preflight can be considered complete.

---

## Required Changes

To achieve a full and safe approval, the following modifications must be made to `DESIGN.md` and verified in Slice 0 deliverables:

1. **Envelope-Capability Constraint Clause:** Add a formal constraint to the Event And Recovery Contract stating that because the `events.Event` envelope carries only `Subject` as an identity field, the stable canonical identity field set (bead ID, session name, alias, configured identity, runtime session key, generation/instance token) **must** travel in a typed payload. Explicitly gate that **no NoPayload session event may back an idempotent-replay or reconciliation subscriber** without first being migrated to a typed payload.
2. **Exhaustive Closed-World Inventory:** Expand the Slice 0 event inventory list (DESIGN.md:521-526) to be exhaustive over all registered `session.*` (and request-result) events on the current branch, specifically including `session.created`, `session.deleted`, and `session.compacted`. Each must be mapped to its payload requirement and public SSE visibility.
3. **Mandatory Generation Token Contract:** Remove "when relevant" from line 508. Mandate that the generation/instance token is **required** for all session-related lifecycle events to prevent cross-generation replay on recycled identities.
4. **Define Work-Release Idempotency Key:** Update DESIGN.md:530-533 to explicitly define the work-release idempotency key as a composite of **bead ID + generation**. Add a matching requirement row in `REQUIREMENTS.md` under `SESSION-WORK-001` or as a new row `SESSION-WORK-005` to align the ledger with this technical constraint.

---

## Questions for the Author / Team

* **Current Subscriber Audit:** Do any active in-process or out-of-process subscribers currently treat a `session.*` event as their primary, non-scan-backed trigger for safety-critical actions (e.g., releasing assigned work or clearing runtime environments)? If yes, those events must be prioritized for typed payload migration immediately in Slice 0.
* **Stranded and Suspended Intent:** Is `session.stranded` intended strictly as an operator-facing alert, or will a recovery adapter eventually subscribe to it? If the latter, it cannot safely remain a `NoPayload` event.
* **Generation Token Source:** What is the authoritative source for the generation/instance token for auto-prov vs configured-named slots, and does the underlying bead schema support storing this token today?
