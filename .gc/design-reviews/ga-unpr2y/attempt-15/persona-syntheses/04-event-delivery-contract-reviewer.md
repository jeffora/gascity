# Amara Osei

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Stable session-event identity is not enforceable for the existing public event set. Claude and Codex both found that the active contract asks for stable identity fields but does not inventory current `session.*` events, classify `NoPayload` usage, or prove that public lifecycle events carry a canonical session bead ID/generation context sufficient for replay, duplicate handling, and stale-event suppression. DeepSeek agrees this identity context is critical for future payload schemas, even though it does not block on the current text.
- [Blocker] Close/retire work-release convergence is not specified tightly enough for crash recovery. Codex blocks because the design does not require a durable release-identity snapshot, scanner trigger, completion marker, supersession key, duplicate-scan behavior, partial-query handling, or stale-successor suppression before identity metadata can be cleared or the bead closed. Claude identifies the same gap as a major risk and asks the design to name the controller tick orphan-release scan and bind it to `SESSION-WORK-001..004`.
- [Major] Existing `session.*` events need a machine-readable event recovery ledger. Claude asks for Slice 0 inventory of existing events with committed fact, critical-vs-best-effort classification, durable recovery owner, and stable identity fields. Codex asks for the same data in `EVENT_RECOVERY.yaml` or equivalent, including public SSE/OpenAPI impact and required tests. Without this ledger, future slices can preserve old subscriber-authority semantics or add event behavior that does not converge from durable facts.
- [Major] Current event docs still imply subscriber-owned recovery policy for safety-critical paths. Codex cites `session.drain_acked_with_assigned_work` and `session.stranded` documentation as saying subscribers choose recovery policy, which conflicts with the new contract that SDK safety-critical recovery must converge from durable scans. Claude's required changes similarly demand that work release have a named durable scan owner rather than an abstract future per-slice field.
- [Major] Idempotency must be shared between event subscribers and durable recovery scans. Claude specifically notes that per-slice "idempotency key or duplicate behavior" is insufficient unless the key is a pure function of durable facts shared by both the event-driven path and the durable scan path. Codex reaches the same point through duplicate scan, stale successor, and crash-after-close-before-scan requirements.
- [Minor] Machine-readable data should not be hidden in event messages. Codex found `session.stranded` uses `NoPayload` while embedding assigned work IDs in human message text. That should either become typed payload data or remain clearly non-machine-readable operator text.

**Disagreements:**
- Verdict severity differs. Codex returns `block`, Claude returns `approve-with-risks`, and DeepSeek returns `approve`. My assessment is `block` for this persona lane because the missing identity and recovery ledger affects events that already exist, not only future event-bearing slices; approving would let close/retire or subscriber-facing work proceed without an enforceable recovery contract.
- Claude frames the current Slice 0 work as safe because it is non-mutating and adds no events. Codex treats the current public event set as already part of the design surface and therefore requiring validation now. My assessment follows Codex on the gate: the first non-mutating slice can still be small, but it must inventory current event contracts before later mutating slices are decomposed.
- DeepSeek accepts the prose contract as sufficient and treats robust payload identity, bounded scans, and external cleanup ownership as implementation advisories. My assessment is that these are not optional advisories for this lane; they are the concrete checks that make "events are facts, durable scans converge" enforceable.

**Missing evidence:**
- A complete inventory of current `events.Session*` types with committed fact, emission owner, typed payload fields, canonical session ID source, critical-vs-best-effort classification, durable scan owner, and public SSE/OpenAPI impact.
- Proof that `NoPayload` session events are allowed only when envelope fields fully and stably capture the event semantics, and that lifecycle event `Subject` values use canonical session identity rather than display/template names.
- A close/retire recovery contract proving that durable facts retain every assignment identifier needed by `SESSION-WORK-001..004` after crash, including bead ID, session name, configured named identity, and old session names.
- The durable marker or query predicate that proves work-release recovery has completed for a closed or retired session identity.
- Negative tests for skipped event emission, duplicate events, stale events after successor creation, event-recorder failure, duplicate durable scans, partial work-query failure, and message-only data that should have been typed payload.
- Whether any subscriber beyond observability surfaces makes control decisions from `session.*` events.

**Required changes:**
- Add a machine-readable `EVENT_RECOVERY.yaml` artifact, or extend `COMMAND_APPLIERS.yaml`/`DIAGNOSTICS_MANIFEST.yaml`, to enumerate every current `events.Session*` type with committed fact, event owner, typed payload fields, canonical session ID source, idempotency key, critical-vs-best-effort classification, durable scan owner, public wire impact, and required tests.
- Add a rule that `NoPayload` is legal only when the event envelope alone carries all stable identity and semantics needed for idempotent replay and diagnostics; otherwise add typed payloads and generated-client proof.
- Name the durable scan owner for the work-release guarantee, currently the controller tick orphan-release scan, and cite `SESSION-WORK-001..004` as the parity oracle.
- Expand the close/retire command contract to persist or derive a release-identity snapshot before any identity-clearing mutation commits, including scanner trigger predicate, query shape, completion marker, duplicate behavior, partial-query behavior, stale-successor suppression, and crash-after-close-before-scan tests.
- Require event-bearing slices to use an idempotency key that is a pure function of durable facts and shared by both event-driven subscribers and durable recovery scans.
- Reclassify `session.drain_acked_with_assigned_work` and `session.stranded` as diagnostics/accelerators unless a controller-owned durable scan is defined in the same contract row, and remove hidden machine-readable data from human messages or promote it to typed payload fields.
