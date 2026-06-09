# Amara Osei - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design states the core event boundary correctly: session events are post-commit facts, not commands, locks, durable truth, or the only recovery mechanism (`internal/session/DESIGN.md:496`).
- Event-bearing slices must now name the committed fact, emission owner, stable session identity fields, typed payload registration, public SSE/OpenAPI/generated-client obligations, critical-vs-best-effort handling, durable scan owner, idempotency, and crash/skipped/duplicate/stale-event tests (`internal/session/DESIGN.md:503`).
- Critical convergence no longer depends on in-process delivery. Close, work release, wake, drain, runtime start, provider-health reactions, and identity retirement must converge from durable facts even when event emission or subscriber execution is skipped (`internal/session/DESIGN.md:517`).
- The current session event inventory is explicitly bounded. Slice 0 must account for `session.woke`, `session.stopped`, `session.crashed`, drain events, kill/suspend/update events, `session.drain_acked_with_assigned_work`, `session.stranded`, and `session.work_query_failed` (`internal/session/DESIGN.md:521`).
- The design aligns with the API control-plane typed-wire rule: events are first-class wire data, every `events.KnownEventTypes` constant needs a registered payload, and envelope-only events must use `events.NoPayload` only when envelope fields capture the semantics (`engdocs/architecture/api-control-plane.md:436`, `engdocs/architecture/api-control-plane.md:467`).

**Critical risks:**
- [Major] Existing `NoPayload` registrations must not be treated as proof that envelope-only identity is sufficient. Today several session events are registered as `NoPayload`, including `session.woke`, `session.draining`, `session.undrained`, `session.quarantined`, `session.idle_killed`, `session.max_age_killed`, `session.suspended`, `session.updated`, and `session.stranded` (`internal/api/event_payloads.go:438`). Some current producers use `Subject` as template/display data rather than a canonical session bead ID, while `SessionLifecyclePayload` exists specifically because `Subject` has mixed meaning (`internal/api/event_payloads.go:238`). Slice 0 needs row-by-row proof that each envelope-only event carries enough stable identity for replay, reconciliation, and public clients; otherwise it should move that event to a typed payload.
- [Major] Work-release recovery is correctly required, but the durable scanner contract still needs concrete ownership before any event-bearing slice is approved. The requirement ledger says successful close releases open/in-progress work assigned by bead ID, `session_name`, or configured identity, and that the path is idempotent (`internal/session/REQUIREMENTS.md:136`). Current code performs direct release/reassign scans and updates (`cmd/gc/session_beads.go:640`, `cmd/gc/session_beads.go:677`), so the design must ensure event rows cannot replace those scans with subscriber-only behavior.
- [Minor] The event bus itself is best-effort: recording errors are logged and not returned (`internal/events/events.go:1`). The design requires skipped-event tests, but Slice 0 should make recorder-failure and nil-recorder cases explicit for critical convergence rows, not only duplicate/stale-event cases.

**Missing evidence:**
- No `DIAGNOSTICS_MANIFEST.yaml`, event inventory, or command-applier rows exist yet that classify each current `session.*` event as critical, accelerator, or best-effort.
- No row yet proves which `NoPayload` events are safe to keep envelope-only versus which need a typed payload with canonical session ID, configured identity, runtime key, and generation/instance token.
- No crash-after-commit, recorder-failure, missed-event, duplicate-event, stale-event, partial-query, or store-query-failure proof exists yet for work release, drain, runtime start, or identity retirement.
- No public-wire proof exists yet for session event payload changes: OpenAPI regeneration, generated Go client compatibility, generated dashboard/TS types, and SSE frame schema coverage.

**Required changes:**
- No design-text blocker remains in this persona lane. The document names the right event-delivery and recovery contract.
- Before Slice 0 closes, add one row per current `session.*` event with committed fact, emission owner, identity fields, payload type, public visibility, criticality, idempotency key, durable scan owner, stale/duplicate behavior, and proof selectors.
- Add negative fixtures showing that critical work release and drain recovery converge when the recorder is nil, recorder writes fail, event delivery is skipped, the event is duplicated, the event is stale, or assigned-work queries partially fail.
- For every session event kept as `NoPayload`, prove that `Actor`, `Subject`, `Message`, `Seq`, `Type`, and `Ts` are sufficient for all intended consumers. Otherwise assign a typed payload and include OpenAPI/generated-client/dashboard proof.

**Questions:**
- Which current `session.*` events are public client contract, and which are local diagnostics only?
- Should session events that identify concrete sessions converge on a shared lifecycle payload carrying canonical session bead ID, configured identity, alias, runtime key, generation, and instance token?
- What is the named durable scan owner and cadence for missed close/work-release, missed drain, missed runtime-start, and missed identity-retirement convergence?
