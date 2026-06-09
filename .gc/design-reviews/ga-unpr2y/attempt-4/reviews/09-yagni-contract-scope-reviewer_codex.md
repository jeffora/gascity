# Kwame Asante - Codex

**Verdict:** approve-with-risks

Reviewed `internal/session/DESIGN.md`, which matches `.gc/design-reviews/ga-unpr2y/attempt-4/design-before.md`, plus `internal/session/REQUIREMENTS.md` and scoped session instructions.

**Top strengths:**
- The design explicitly rejects a broad `SessionFacts` facade and says shared types graduate only after repeated exact use.
- Slice 1 is constrained to target-classification facts/results and a policy input; wake, runtime-start, close, reconciler, and event vocabulary are deferred until their slices begin.
- TR-007 keeps future durable-event compatibility open without making event sourcing the first implementation step.

**Critical risks:**
- [Major] Slice 1 can still become too wide if "complete the target collision inventory for API, CLI, mail, extmsg, nudge, attach, inspect, materialization, and assignee normalization" is interpreted as implementing all those surfaces before the first narrow extraction lands. The design should state that the broad work is an inventory/parity matrix, while implementation delegates one surface at a time behind the old resolver oracle.
- [Minor] The shared vocabulary checkpoint table names `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent` before their slices exist. The surrounding text defers them correctly, but task authors could still introduce the names early "because the design says so." Add an explicit rule that checkpoint names are design placeholders until a slice creates the first real caller and tests.
- [Minor] TR-007 could still be over-applied. "Design event names and payloads so they can later become durable event-log facts" should not require replay cursors, durable outbox interfaces, event-store APIs, or event-sourcing-specific IDs in the first in-process event pass.

**Missing evidence:**
- A slice-1 task boundary that separates inventory-only work from the first delegated implementation surface.
- A type-graduation rule that says where provisional slice-local types live and when they may move to shared `internal/session` vocabulary.
- A concrete "not yet" list for TR-007 so future-event-log compatibility does not pull durable outbox or replay API work into this migration.

**Required changes:**
- Make the slice-1 MVP explicit: candidate collector plus one compatibility adapter and parity tests, then surface-by-surface adoption.
- Mark non-slice-1 vocabulary as reserved design terms, not implementation deliverables.
- Add a durable-event deferral note: stable factual payloads are allowed now; outbox persistence, replay APIs, durable sequence contracts, and event-sourcing reducers require a separate approved design.

**Questions:**
- Which target-resolution surface is intended to be the first delegated caller: API live target, `internal/session.ResolveSessionID`, mail, or assignee normalization?
- Where should provisional fact/result types live before they have two exact-use callers?
- Does TR-007 require changing any current event payload in slice 1, or is it only a constraint on later command/event slices?
