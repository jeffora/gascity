# Kwame Asante - Claude

**Verdict:** approve-with-risks

Scope of this review is YAGNI and contract-scope discipline: whether vocabulary
is introduced only when a slice needs it, whether TR-007's future durable-event
ambition is shaping today's APIs more than in-process events require, and whether
anything is on a path to a broad `SessionFacts` facade. The facade and
unused-vocabulary red flags are well-controlled and not realized in code. The
one real scope risk is event-payload shaping for a deferred event log.

**Top strengths:**
- **Shared vocabulary is gated by an explicit "First allowed caller" and a
  two-caller promotion rule — the strongest YAGNI control in the document.** The
  checkpoints table (DESIGN.md:102-108) binds each proposed type to the slice
  that first needs it (`TargetCandidate`/`TargetSelection` → slice 1,
  `SessionCommandConflict` → slice 2, `RuntimeStartIntent` → slice 3,
  `SessionFactEvent` → later), and the rule "A shared type is allowed only after
  two implemented deciders consume the same exact field set" (DESIGN.md:76-78)
  plus "A slice must document the first delegated caller and tests before moving a
  vocabulary type from private to shared" (DESIGN.md:109-113) is a concrete gate,
  not a slogan. Verified that none of these types exist in the checkout yet
  (`grep` for `type SessionFacts|TargetCandidate|TargetSelection|...` returns
  nothing), so the design is enumerating intent, not landing speculative code.
- **The facade red flag is defended in four places and is consistent with the
  existing code shape.** "Avoid one broad `SessionFacts` type" (DESIGN.md:76),
  TR-002 "Broad shared fact structs are introduced only after repeated exact use"
  (DESIGN.md:665), Non-Goal "Do not introduce one broad session facade before a
  narrow contract is proven" (DESIGN.md:882), and `AGENTS.md`:34-35. The existing
  baseline vocabulary is already operation-specific — `RuntimeFacts`,
  `LifecycleInput`, `LifecycleView`, `NamedIdentityInput`
  (`internal/session/lifecycle_projection.go:143-176`), `PreWakePatchInput`,
  `CommitStartedPatchInput` — so the design's "reuse directly" rows
  (DESIGN.md:82-89) inherit narrow types, not a facade. The "one mega
  SessionFacts struct" red flag has no current foothold.
- **Slice 1's vocabulary is scoped to what target classification actually needs.**
  Slice 1 introduces exactly target-classification facts, results, candidate
  conflict details, and a per-operation policy input (DESIGN.md:91-98, 276-283),
  and later fact families are explicitly deferred ("Later slices introduce
  lifecycle command facts, reconciler facts, and runtime intent results only when
  those slices begin", DESIGN.md:96-98). That is the correct minimal first cut.

**Critical risks:**
- **[Major] Event payloads are being shaped for a deferred event log, against the
  design's own Non-Goal and durable-scan recovery model.** The design defers event
  sourcing ("Do not make event sourcing the first implementation step",
  DESIGN.md:885), states in-process events are "only for acceleration or
  observability" (TR-007, DESIGN.md:724-726), and makes recovery durable-scan
  based ("No safety-critical reaction may depend only on at-most-once in-process
  event delivery", DESIGN.md:372-374; "the load-bearing recovery rule is the
  durable assigned-work scan", DESIGN.md:409-412). Yet the event contract mandates
  that payloads "carry stable IDs, generation/instance token when relevant, and
  idempotency keys for subscriber scans" (DESIGN.md:391), and the per-event matrix
  assigns each event a payload-level "Duplicate key" (DESIGN.md:398-407). Because
  the durable scan re-reads generation/work facts from beads, none of those
  payload-carried keys are needed today; specifying them now is payload shape
  driven by the future durable-event-log model the design postpones. This is the
  lane's red flag "event payloads are contorted for speculative event sourcing,"
  and it is the concrete answer to lane question 2: TR-007's future-compat goal is
  shaping current event APIs beyond what today's observability-only events
  require. (Independently corroborated by the registered payloads being thin —
  e.g. `SessionStranded` is `events.NoPayload{}` — so the matrix is already
  ahead of code.)
- **[Minor] Full contracts for slices 5-7 are front-loaded before slices 1-2
  prove the pattern.** The document is 886 lines with ~19 top-level contracts plus
  detailed Command-Atomicity, Operability, and Performance rows for drain, close,
  health, and circuit work that belongs to slices 4-7. This is defensible as a
  gating backlog (each row is a gate, not implementation), but the detailed
  required-field enumerations for `RuntimeStartIntent` (slice 3) and
  `SessionFactEvent` (later) at DESIGN.md:107-108 risk being treated as frozen
  contracts and drifting before their slice starts.

**Missing evidence:**
- The checkpoints table does not state what happens if a slice reaches its "First
  allowed caller" and finds it needs only a subset of the enumerated required
  fields. There is no rule that the field list may shrink at slice start, so the
  enumerations read as commitments rather than estimates.
- No explicit statement that slices 5-7's contracts are provisional. The document
  would benefit from marking which contracts are frozen gates (slices 1-2) versus
  indicative scaffolding (slices 3-7), so reviewers and implementers do not treat
  the entire 886-line surface as equally binding now.
- TR-007's acceptance criteria do not state the minimum payload (envelope-only)
  that is acceptable for the in-process phase, so there is no lower bound that
  would stop a future implementer from "completing" the matrix's speculative
  fields as if required.

**Required changes:**
1. Decouple the event contract from the deferred event log. State that for the
   in-process phase, payloads remain thin diagnostic accelerators (envelope plus
   minimal subject identity), and that idempotency/dedup keys are computed by the
   durable scan from bead facts. Move the "generation/instance token/idempotency
   key" requirements out of the mandatory payload contract and into a
   clearly-labeled "if and when a durable outbox is designed" note, so no payload
   is enriched speculatively. Relabel the matrix's per-event "Duplicate key"
   column as a scan-side key, not a payload field.
2. Add an explicit shrink rule to the vocabulary checkpoints: a slice may
   introduce fewer fields than enumerated when its first caller needs fewer, and
   the enumerations are upper bounds, not commitments.
3. Mark slices 3-7 contracts (atomicity rows, operability mappings, performance
   budgets, and `RuntimeStartIntent`/`SessionFactEvent` field lists) as
   provisional scaffolding to be finalized at slice start, distinguishing them
   from the frozen slice-1/slice-2 gates.

**Questions:**
- Is the intent that the in-process event phase keeps the existing thin payloads
  unchanged (no new typed payloads in slices 1-2), with the matrix describing only
  the durable-fact set the scan reads? If so, say it explicitly so the matrix is
  not read as a wire-payload contract.
- For `RuntimeStartIntent` and `SessionFactEvent`, is enumerating required fields
  now (slices 3+) meant as a binding contract or as a sketch? The answer
  determines whether these belong in this document today or in the slice's own
  design when it starts.
- Does the two-caller promotion rule apply to policy-input types
  (per-operation policy input) as well as fact/result types, or can a policy
  struct be shared after a single caller? The document is explicit for facts but
  silent for policy inputs.
