# Amara Osei - Claude

**Verdict:** approve-with-risks

Lane: factual session events, idempotent subscribers, crash recovery,
work-release guarantee. Scope reviewed: the Event And Recovery Contract
(DESIGN.md:496-536) and the work-release/drain rows in `REQUIREMENTS.md`, read
against the live event registry. I grounded the claims in code: the work-release
path and the event payload map below are what the branch actually does, not what
the document hopes.

From this lane the design's event model is sound — it correctly frames events as
post-commit facts, mandates durable-scan convergence independent of event
delivery (which I verified is how release works today), and requires per-event
critical/best-effort classification. The risks are gaps in the Slice 0 event
inventory, not flaws in the model, so I approve with the required changes below.

**Top strengths:**
- Events are declared post-commit facts, "not commands, locks, durable truth, or
  the only recovery mechanism" (DESIGN.md:498-499), and `Message` text is
  fenced as operator-only with subscribers limited to typed payload + envelope
  fields (line 533-535). That squarely forecloses my red flag #1 (imperative
  instructions in payloads).
- The work-release guarantee does not depend on observing an in-process event:
  "Close, work release, wake, drain, runtime start... must converge from durable
  facts even when event emission or subscriber execution is skipped"
  (DESIGN.md:517-519). I confirmed the current implementation matches — release
  converges through the reconciler scan in `cmd/gc/session_beads.go` /
  `cmd/gc/session_reconciler.go` (e.g. the "dead runtime session retained — open
  assigned work blocks alias release" path), not through a subscriber. This is
  the right defense against my red flag #2 and it is real today.
- The work-release convergence row requirements are thorough and answer lane Q2
  directly: release-identity snapshot, scanner trigger, completion marker,
  duplicate/stale behavior, partial-query behavior, idempotency key, and explicit
  tests for crash-after-close, missed event, duplicate event, and store-query
  failure (DESIGN.md:530-533).

**Critical risks:**
- **[Major] The required stable-identity field set cannot live in the envelope,
  and the safety-adjacent events are NoPayload — so they cannot support
  idempotent replay if a critical subscriber is ever attached.** The design
  requires events to carry "bead ID, session name, alias, configured identity,
  runtime session key, and generation/instance token when relevant"
  (DESIGN.md:506-508). But the live `Event` envelope
  (`internal/events/events.go:162-170`) carries only `Seq, Type, Ts, Actor,
  Subject, Message, Payload` — the sole identity field is one `Subject` string,
  with **no generation/instance token**. So those fields can only ride a typed
  `Payload`, yet I verified that `session.stranded`, `session.draining`,
  `session.undrained`, `session.suspended`, and `session.woke` are all
  registered as `events.NoPayload{}` (`internal/api/event_payloads.go:444-456`).
  A subscriber to `session.stranded` therefore gets only `Subject` + operator
  `Message` and **cannot distinguish a closed session from a respawned
  same-identity session** (pool/configured-named identities do reuse identity per
  SESSION-ID-007 / SESSION-WORK-002). This is safe only while those events stay
  best-effort and the durable scan is the real mechanism — which the design
  intends but does not pin down. This is my red flag #3 latent in the contract.
- **[Major] The Slice 0 event inventory is a minimum of 13, not exhaustive, so
  real session events reach no classification.** The registry defines more
  `session.*` events than the list at DESIGN.md:521-528 — verified extras include
  `session.created`, `session.deleted`, `session.compacted`, `session.message`,
  `session.submit`, `session.jsonl` (plus the `RequestResultSession{Create,
  Message,Submit}` results). `session.deleted` is a terminal lifecycle fact a
  work-release/cleanup subscriber could legitimately react to, yet "at least for"
  permits an inventory of exactly 13 that omits it. For a subscriber contract the
  inventory must be closed-world: every emitted session event classified
  critical/best-effort with payload adequacy, or no one can prove an event won't
  reach a subscriber uncontracted.
- **[Minor] "generation/instance token when relevant" is too soft for
  work-release-class events.** "When relevant" is a judgment call; for any event
  whose subscriber must reconcile against a possibly-reused identity, the
  generation/instance token is the linchpin of idempotent replay and is *always*
  relevant. As written, an implementer can omit it and still satisfy the prose.

**Missing evidence:**
- No statement of which current subscribers (if any) consume `session.stranded`,
  `session.draining`, `session.undrained`, or `session.suspended`. The
  best-effort-vs-critical safety of those NoPayload events hinges entirely on
  there being no critical consumer, and the document does not assert it.
- No envelope-capability row. The contract lists required identity fields without
  noting that the envelope provides only `Subject`, so a reader cannot see that
  NoPayload ⇒ "single-string identity only." Slice 0's inventory needs that fact
  recorded per event.
- No idempotency-key definition for the release operation itself — the design
  requires "idempotency key" as a field but does not say what makes two release
  attempts (scan-driven and event-driven) identical, i.e. keyed on bead ID +
  generation rather than session name alone.

**Required changes:**
1. Add an envelope-capability note to the Event And Recovery Contract: the
   `events.Event` envelope carries only `Subject` as identity, so the required
   stable-identity field set (bead ID, session name, alias, configured identity,
   runtime session key, generation/instance token) must travel in a **typed
   payload**. State explicitly that a NoPayload session event exposes only
   `Subject` + envelope fields and therefore **cannot** back an idempotent-replay
   or reconciliation subscriber.
2. Make the Slice 0 event inventory **exhaustive** over all registered
   `session.*` (and session request-result) events, not a minimum of 13.
   Explicitly classify `session.created`, `session.deleted`, and
   `session.compacted` critical/best-effort with payload adequacy and identity
   sufficiency.
3. Add a gate: **no NoPayload session event may be promoted to a critical
   subscriber reaction** without first being given a typed payload carrying bead
   ID + session name + generation/instance token. Tie this to the
   crash-after-commit / duplicate / stale-event tests the contract already
   requires.
4. Replace "generation/instance token when relevant" with a mandatory rule for
   work-release-class and identity-retirement events: the generation/instance
   token is required so a subscriber cannot act on a stale generation after the
   identity is reused.
5. Require the work-release convergence row to define its **idempotency key as
   bead ID + generation**, and to prove the scan-driven and event-driven release
   paths are mutually idempotent (the duplicate-event and store-query-failure
   tests the contract already mandates should assert this).

**Questions:**
- Does any current subscriber treat a session event as the *primary* trigger for
  a safety-critical reaction (work release, alias release, cleanup), or is every
  such reaction already scan-backed? The answer decides whether the NoPayload set
  is acceptable as-is or must be upgraded before Slice 0 closes.
- Is `session.stranded` intended as a pure operator signal (best-effort) or as a
  recovery trigger? If the latter, it needs a typed identity payload now.
- For identity-reuse cases (pool slots, configured named respawn), what is the
  authoritative generation/instance token source the inventory should cite, and
  is it present on the bead today?
