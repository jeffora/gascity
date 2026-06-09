# Amara Osei - Claude

**Verdict:** approve-with-risks

Scope of this review is the event-delivery and recovery contract: whether
session events are facts (not commands), whether safety-critical subscribers
(work release) converge after a crash between durable mutation and in-process
event delivery, and whether reactions are classified critical-vs-best-effort
with a documented recovery path. The recovery architecture is correct and is the
strongest part of the document. It is held at approve-with-risks because the
per-event payload-field and idempotency-key claims do not match the payloads
actually registered in code, which would mislead a subscriber implementer.

**Top strengths:**
- **Work release does not depend on observing an in-process event — verified in
  code.** The design's load-bearing rule (DESIGN.md:372-374, 409-412) is that
  critical reactions converge from durable session facts plus idempotent
  controller scans, and that no implementation may require
  `SessionDrainAckedWithAssignedWork` to fire. The source confirms this is
  already true: the event is emit-only from the reconciler
  (`cmd/gc/session_reconciler.go:145-154`), there is **no** production subscriber
  consuming session events to drive work release, and release/reassignment runs
  through durable bead-scan functions (`cmd/gc/session_beads.go:623`
  `unclaimWorkAssignedToRetiredSessionBead`, `:677`
  `reassignWorkAssignedToRetiredSessionBead`, `:733`
  `cancelStateAssignedToRetiredSessionBead`) invoked on the close/sync path. A
  crash after the durable mutation but before event delivery therefore converges
  on the next reconciler tick. This directly and correctly answers the lane's
  crash-recovery question.
- **Events are modeled as facts, not imperative subscriber instructions.** The
  registered `SessionDrainAckedWithAssignedWorkPayload`
  (`internal/api/event_payloads.go:398-405`) carries descriptive bead-side
  context (session ID, bead ID, template, bead status, reason) and its own doc
  string frames it as a "mechanism-only signal ... so pack-level subscribers can
  apply recovery policy without baking pack-specific knowledge into the SDK."
  TR-005 and TR-007 (DESIGN.md:694-708, 722-736) plus the event contract
  (DESIGN.md:385-394) reinforce facts-over-commands and forbid assuming the
  in-process mechanism is the permanent persistence authority. The red flag
  "payloads encode imperative subscriber instructions" is not present.
- **Critical-vs-best-effort is explicitly classified with a per-row recovery
  authority.** The Runtime Intent/Event/Recovery Ordering table
  (DESIGN.md:376-383), the per-event reaction matrix (DESIGN.md:398-407), and the
  subscriber classification table (DESIGN.md:416-421) each name a tier and a
  "Durable recovery authority," and correctly demote trace/SSE/dashboard to
  observability-only. This is a complete answer to "which reactions are critical
  versus best-effort, and is the recovery path documented for each."

**Critical risks:**
- **[Major] The per-event reaction matrix specifies payload fields and dedup keys
  that the registered payloads do not provide, and the design does not flag the
  required payload upgrades.** Three concrete mismatches:
  - `SessionDrainAckedWithAssignedWork`: matrix dedup key is "session ID + bead
    ID + **drain generation**" (DESIGN.md:404), but
    `SessionDrainAckedWithAssignedWorkPayload` has no drain-generation field
    (`internal/api/event_payloads.go:398-405`; builder signature
    `(sessionID, beadID, template, beadStatus, reason)` at `:412`). A subscriber
    cannot construct the stated dedup key from the event.
  - `SessionStranded`: matrix claims fields "session ID, assigned work refs, probe
    source, reason" and dedup key "session ID + work refs + probe generation"
    (DESIGN.md:406), but it is registered as `events.NoPayload{}`
    (`internal/api/event_payloads.go:455`). Only the envelope `Subject`/`Message`
    (`internal/events/events.go:162-170`) is available; work refs and probe
    generation exist nowhere on the event.
  - `SessionWorkQueryFailed`: matrix claims "store ref, query class, error,
    session ID if scoped" (DESIGN.md:407), but it reuses generic
    `SessionLifecyclePayload` (`:456`), which carries only SessionID, Template,
    Reason (`:245-249`) — no store ref, query class, or error.

  Because the design itself mandates that "New events must be registered typed
  payloads" with `events.RegisterPayload`, OpenAPI/SSE projection updates, and
  `TestEveryKnownEventTypeHasRegisteredPayload` parity (DESIGN.md:88, 391-392;
  the coverage test lives at `internal/api/event_payloads_coverage_test.go`), any
  enrichment of these three payloads is a typed-wire change with dashboard/client
  regeneration cost. The matrix presents the enriched fields as the contract
  without acknowledging that work. This is the lane's red flag "payloads omit
  identity context needed for idempotent replay or reconciliation" — mitigated
  only because the durable scan re-reads generation from the bead, so recovery
  still works; but a subscriber coded against the matrix would expect fields that
  do not exist.
- **[Minor] The matrix conflates "event payload dedup key" with "durable-scan
  dedup key."** The design's own thesis is that dedup/idempotency is enforced by
  the scan reading bead facts (generation, assigned work). Labeling those keys as
  the event's "Duplicate key" (DESIGN.md:404, 406) invites an implementer to
  dedup on event contents, which is exactly the at-most-once dependence the
  design forbids elsewhere. The intent is right; the labeling undercuts it.

**Missing evidence:**
- No statement of which envelope field carries the subject identity for
  `NoPayload` events. `SessionStranded` relies on `Event.Subject`
  (`internal/events/events.go:167`) for session ID, but the design never says so,
  so "session ID" in the Stranded row is unverifiable from the document alone.
- The per-event matrix has no row-level proof column. Unlike the Scenario
  Traceability Matrix, the event rows cite no test proving that a missed/dropped/
  duplicated delivery still converges. The claim "scan recovers if event missing"
  (DESIGN.md:404) is asserted, not tied to a `session_reconciler_test.go` /
  `session_wake_test.go` case (SESSION-WORK-004 exists for drain-ack vs wake
  demand, but the matrix does not reference it).
- No idempotency-key definition for the `RequestResult*` async result events or
  for `events.SessionWoke` as it relates to a prepared runtime-start generation;
  the matrix says "generation/token if start is prepared" but the wake payload's
  carriage of that token is not shown.

**Required changes:**
1. Reconcile the per-event reaction matrix with the registered payloads. For each
   of `SessionDrainAckedWithAssignedWork`, `SessionStranded`, and
   `SessionWorkQueryFailed`, either (a) state that the listed fields are the
   durable-fact set the recovery scan reads from beads — and that the event
   payload is a thin diagnostic accelerator that need not carry them — or (b)
   commit to enriching the typed payload (add `drain_generation`; give Stranded
   and WorkQueryFailed dedicated typed payloads) as an explicit typed-wire change
   gated by `events.RegisterPayload`, OpenAPI/SSE regeneration, dashboard/client
   regeneration, and `event_payloads_coverage_test.go` parity.
2. Relabel the matrix's "Duplicate key" column to make clear the dedup key is
   computed by the durable scan from bead facts, not from the event payload, so no
   implementer dedups on at-most-once event contents.
3. Name the envelope field (`Event.Subject`) that carries session identity for
   `NoPayload` session events, and state the minimum identity context guaranteed
   on the envelope for every critical/diagnostic event.
4. Add a proof column (or reference the existing scenario rows, e.g.
   SESSION-WORK-004) to each critical event row demonstrating missed/duplicate/
   out-of-order delivery still converges via scan.

**Questions:**
- Is the intent that no new typed payloads are created in the first slices — i.e.
  the existing thin payloads stay and the matrix is purely a durable-fact
  specification? If so, the document should say so explicitly so reviewers do not
  read the matrix as a wire contract.
- For `SessionStranded` (NoPayload today), is upgrading to a typed payload in
  scope of this design, or is it deliberately left as envelope-only because the
  stranded condition is always re-derivable by the work/session scan?
- When runtime start commits `generation`/`instance_token` before emitting
  `SessionWoke`/start events, will those tokens be added to the event payloads (a
  wire change) or remain durable-only on the bead for the scan to read?
