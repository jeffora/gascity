# Amara Osei

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Major] The event model is directionally sound: session events are post-commit facts, not commands or durable truth, and safety-critical convergence must come from durable facts even when event emission or subscriber execution is skipped.
- [Major] `NoPayload` session events have an identity and generation blind spot. The event envelope exposes only `Subject` as identity, so envelope-only events such as `session.woke`, `session.draining`, `session.undrained`, `session.suspended`, and `session.stranded` cannot carry bead ID plus generation/instance token for idempotent replay or stale-event suppression.
- [Major] Slice 0 event inventory must be closed-world, not a minimum list. Every registered `session.*` event and session request-result event needs classification, including events outside the current design list such as `session.created`, `session.deleted`, `session.compacted`, `session.message`, `session.submit`, and `session.jsonl`.
- [Major] Critical recovery paths need concrete durable-scan contracts. Work release, drain, runtime start, identity retirement, and close/retire recovery must name the durable scan owner, idempotency key, stale/duplicate behavior, partial-query behavior, completion marker, and failure fixtures.
- [Major] "Generation/instance token when relevant" is too permissive for lifecycle, identity-reuse, work-release, and critical subscriber events. For any event where a subscriber could act on a reused identity, the generation or instance token is required.
- [Minor] Public wire impact is not optional. Typed payload additions or NoPayload decisions must include OpenAPI/SSE/generated-client/dashboard proof where the event is externally visible.
- [Minor] Machine-readable data must not be hidden in event `Message` text. If a subscriber needs work IDs, identity fields, or generation data, those fields belong in a typed payload.

**Disagreements:**
- There is no verdict disagreement: Claude, Codex, and DeepSeek all return `approve-with-risks`.
- Codex says no design-text blocker remains and frames the issues as Slice 0 close gates. Claude and DeepSeek want several clauses added to the design text now, especially envelope capability and NoPayload constraints. My assessment: the design shape can proceed, but those clauses or equivalent Slice 0 contract rows are required before Slice 0 closes.
- Claude and DeepSeek require generation/instance token for work-release-class, identity-retirement, or all lifecycle events; Codex asks for row-by-row proof that envelope-only events are sufficient. My assessment: generation/instance token is mandatory for lifecycle events that can be replayed across reused identities or drive critical reactions; purely diagnostic events may remain NoPayload only with explicit proof.
- DeepSeek emphasizes defining the idempotency key as bead ID plus generation. Claude asks for the same key for work release specifically; Codex asks for row-level idempotency proof. I assess bead ID plus generation as the right default for work-release and identity-reuse events unless a row justifies a stronger durable key.

**Missing evidence:**
- A complete inventory of current `session.*` and session request-result events with committed fact, emission owner, payload type, canonical identity fields, public visibility, criticality, durable scan owner, idempotency key, stale/duplicate behavior, and proof selectors.
- A documented envelope capability row stating that `Event.Subject` is the only envelope identity field and that generation/instance token cannot be supplied by `NoPayload`.
- A subscriber audit showing whether any current in-process or out-of-process subscriber treats `session.*` events as primary triggers for work release, alias release, cleanup, drain, or other safety-critical actions.
- Proof for each NoPayload session event that envelope fields and operator message text fully capture the semantics for all intended consumers.
- The authoritative generation/instance token source for pool, configured-named, and respawned identities, and whether that token is present on the bead at emission time.
- Work-release convergence tests for crash-after-close, skipped event emission, nil recorder, recorder failure, duplicate events, stale events, duplicate durable scans, partial assigned-work query failure, and store-query failure.
- OpenAPI, generated-client, SSE frame, and dashboard type proof for any public session event payload changes.

**Required changes:**
- Add an envelope-capability constraint: the event envelope carries only `Subject` as identity, so stable identity fields such as bead ID, session name, alias, configured identity, runtime session key, and generation/instance token must travel in typed payloads when subscribers need idempotent replay or reconciliation.
- Make the Slice 0 event inventory exhaustive over all current `session.*` and session request-result events, not only the listed minimum set; classify each event as critical, accelerator, diagnostic, or best-effort with payload adequacy and public-wire impact.
- Add a gate that no NoPayload session event may back a critical subscriber reaction unless the envelope alone is proven sufficient; otherwise the event must gain a typed payload and generated-client/OpenAPI proof.
- Replace "generation/instance token when relevant" with a mandatory token rule for lifecycle, work-release, identity-retirement, and identity-reuse events that can be replayed or consumed by critical subscribers.
- Define the work-release idempotency key, defaulting to bead ID plus generation, and require scan-driven and event-driven paths to share that durable key.
- Add negative fixtures for nil recorder, recorder failure, skipped event delivery, duplicate event delivery, stale event delivery, crash-after-commit-before-event, duplicate scans, partial work-query failure, and store-query failure.
- Keep event `Message` operator-only: any work IDs, identity data, or recovery inputs needed by subscribers must be typed payload fields.
