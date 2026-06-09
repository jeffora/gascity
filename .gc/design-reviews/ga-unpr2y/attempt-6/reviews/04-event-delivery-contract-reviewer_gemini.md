# Amara Osei — DeepSeek V4 Flash (Independent Review, Attempt 6)

**Verdict:** pass

**Persona:** factual session events, idempotent subscribers, crash recovery, work-release guarantee, cross-document consistency.

**Reviewed against:** `internal/session/DESIGN.md` (Attempt 6, matching `.gc/design-reviews/ga-unpr2y/attempt-6/design-before.md` with Attempt 6 changes), `internal/events/events.go`, previous attempt reviews, and `REQUIREMENTS.md` scenarios.

---

## Top Strengths

- **Resolution of Systemic Payload Mismatch ("The Durable-Fact Model")**: 
  The designer has beautifully addressed the blocker from Attempt 5 by decoupling wire payload migration from scan-side convergence. Labeling the matrix "Required durable fact or payload fields" and explicitly stating that current events are "thin hints" (with wire payload upgrades deferred to a future event-migration slice) completely resolves the payload mismatch risk. This avoids forcing massive, unnecessary OpenAPI/SSE regeneration churn in Slice 1 (respecting YAGNI), while still clearly defining what the required *durable fact* identity fields are for scan-side convergence.
  
- **Introduction of the "Durable Reaction Recovery Gate" Matrix**: 
  The added matrix (`DESIGN.md:228-240`) is an exceptional addition. Specifying the Cadence, Completeness guard, Stale-event handling, Duplicate behavior, and Tests for all 5 major reactions (Work release, Drain cancel/complete, Wake demand, Identity retirement, Mail/extmsg binding) perfectly implements the robust crash-recovery and convergence principles required by this persona.
  
- **Explicit Performance Boundaries with Default Caps**: 
  Specifying that critical subscriber passes must perform at most one synchronous pass, capped to 100 affected items or 250 ms of wall time before deferring to controller/reconciler repair, ensures that synchronous cascades cannot block the hot reconciler loop, satisfying the Bitter Lesson.

---

## Remaining Risks & Cleanup (Recommendations)

### 1. Stale Roadmap Placeholder (Line 1252)
Line 1252 of `internal/session/DESIGN.md` still carries the stale placeholder from earlier attempts:
```
- Define whether the close command performs synchronous work release or returns a post-commit fact consumed by an idempotent scan in the same controller tick.
```
**Required Action:** Since the "Durable Reaction Recovery Gate" table has already resolved this by declaring the cadence as "same controller tick or next reconciler scan" and defining the completeness guard, this placeholder is obsolete and should be removed/cleaned up to prevent implementer confusion.

### 2. Event Envelope Subject Routing for Thin Events
Since current events are defined as "thin hints" without full payload data, subscribers must route and filter scans using the event envelope's `Subject` field (which carries the target session/bead ID). 
**Required Action:** The implementation team should explicitly ensure that the event recorder populates the `Subject` field for all session-lifecycle events, avoiding the need for costly full-city scans.

---

## Answers to Persona Questions

1. **Are SessionEvent payloads facts with stable identifiers rather than commands to work, mail, trace, or extmsg?**
   - Yes. The updated design explicitly dictates that any added wire payloads must describe facts that happened (not commands to attempt) and carry stable IDs, generation/instance tokens, and idempotency keys.
2. **If the process crashes after a durable session mutation but before in-process event delivery, how does a safety-critical subscriber such as work release converge?**
   - It converges via an idempotent scan of durable session and work facts. The "Durable Reaction Recovery Gate" table guarantees that work release is backstopped by the reconciler/controller scan running either on the same controller tick or the next reconciler scan, ensuring complete event-independent crash recovery.
3. **Which reactions are critical versus best-effort, and is the recovery path documented for each?**
   - Yes, the "Durable Reaction Recovery Gate" and "Subscriber classification" tables explicitly classify Work release, Runtime reconciliation, Mail/extmsg binding cleanup, and Drain cancel/complete as critical retryable paths with documented recovery authorities, while Trace/SSE/dashboard are correctly categorized as observability-only (best-effort).

---

## Questions for the Author

1. Will the event envelope's `Subject` field be strictly mandated as the session ID for all thin session events to optimize subscriber scan scoping?
2. For the 250 ms / 100-item synchronous subscriber limit, will the first implementation slice include telemetry or assertion tests to verify this ceiling is not breached under large-city profiles?
