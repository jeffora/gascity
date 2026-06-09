# Kwame Asante — Gemini (Independent Review, Attempt 4)

**Verdict:** block

**Review focus:** minimal vocabulary, facade creep, event-log deferral, and backlog scope control — with evidence drawn from code verification.

The Attempt 4 iteration of the session boundary design makes substantial progress in structuring vocabulary checkpoints and limiting the shared-type surface. The "Rule of Two" shared-type graduation check and the explicit gating by "First allowed caller" are excellent YAGNI safeguards.

However, from a contract-scope and backlog control perspective, significant risks of speculative complexity and scope-creep remain. Specifically, event payloads are being contorted for a deferred event-sourcing model that the design itself postpones, and the read-only Slice 1 is gated by a complex AST-based mutation static guard. This review details these findings and blocks the design until the required changes are resolved.

---

## Top Strengths

- **The Rule of Two Shared-Type Guard (`DESIGN.md:76–78`):** Preventing the introduction of shared fact/result types until exactly two implemented deciders consume them is the strongest possible defense against monolithic facade creep. It ensures types remain localized and minimal.
- **Staged Vocabulary Checkpoints (`DESIGN.md:100–108`):** The checkpoint table binds proposed types directly to their introduction slices (`TargetCandidate`/`TargetSelection` in Slice 1, `SessionCommandConflict` in Slice 2, `RuntimeStartIntent` in Slice 3, etc.). This ensures we do not pollute the early slices with future vocabulary.
- **Durable Recovery Priority over Event-Replay (`DESIGN.md:370–374, 409–412`):** Keeping safety-critical recovery anchored on durable facts and scans (e.g., the durable assigned-work scan for work release) rather than on speculative in-process events is an excellent Bitter Lesson application.

---

## Critical Risks

### [Blocker] Event payloads are being heavily contorted for speculative event sourcing (TR-007 overreach)

The design explicitly defers event sourcing as a non-goal ("Do not make event sourcing the first implementation step", `DESIGN.md:885`), states that in-process events are only for acceleration/observability (`DESIGN.md:370–372`), and makes recovery durable-scan based (`DESIGN.md:372–374, 409–412`). 

Despite this, the event contract at `DESIGN.md:387–389` mandates that payloads *"carry stable IDs, generation/instance token when relevant, and idempotency keys for subscriber scans."* Furthermore, the per-event reaction matrix (`DESIGN.md:396–407`) assigns each event a payload-level "Duplicate key" and requires fields like `generation/token if start is prepared`, `wake cause`, and `captured output ref, generation/token` on the event payload.

Because critical subscribers and recovery loops re-read actual generation, liveness, and work facts directly from the durable bead store (`DESIGN.md:372–374`), none of these payload-carried keys are necessary for the in-process acceleration phase. Forcing their design and implementation now is a clear YAGNI violation. It shapes current in-process event APIs based on a future durable-event-log model that is postponed.

**Required change:**
1. Explicitly decouple the current event payloads from the deferred event log. State that for the current in-process phase, event payloads remain thin diagnostic/acceleration signals (carrying only the envelope, minimal subject identity, and trace context).
2. Move the "generation, instance token, and idempotency key" payload requirements out of the active contract and into a clearly-labeled "Deferred Event Sourcing Phase" note.
3. Relabel the reaction matrix's "Duplicate key" column (`DESIGN.md:396–407`) to specify that these are scan-side or consumer-computed keys, rather than payload-side fields.

### [Blocker] Read-only Slice 1 is gated by a complex AST/symbol-based mutation static guard

The static guard section (`DESIGN.md:182–186`) states: *"Before a slice is considered implemented, add or tighten a failing-build guard that rejects new production bypasses. The guard should be a Go AST/symbol guard, not a comment convention."* It requires scanning all production packages for calls to `SetMetadata`, `Update`, `Create`, or `Close` (`DESIGN.md:198–202`).

While this is appropriate for mutation-owning slices (Slices 2–7), applying this global requirement to Slice 1 is a major backlog scope-control failure. Slice 1 (Target classification) is strictly read-only and does not perform any metadata mutations. 

Forcing implementers of Slice 1 to build a complex, multi-package Go AST/symbol static scanner to analyze database writers before they can ship the basic candidate classifier will unnecessarily stall the entire migration's first step.

**Required change:**
1. Scope the static-guard requirement explicitly by slice type and field family.
2. State that the mutation-bypass AST/symbol scanner is required only when the first mutation-owning slice (Slice 2, Explicit wake command) begins.
3. For Slice 1, require only a target-classification adoption guard (e.g., a simple symbol check to prevent the creation of new raw resolution fallbacks) or comprehensive parity tests.

### [Blocker] Slice 1 target classification is vulnerable to a "big bang" implementation due to lack of incremental surface delegation

The Target Classification Contract (`DESIGN.md:253–268`) lists 11 separate compatibility resolver chains (API target, ResolveSessionID, ResolveSessionIDAllowClosed, Mail send, Mail query, Extmsg, Attach, Assignee normalization, etc.) that must be unified in Slice 1. 

While documenting these chains as an inventory is highly valuable, the design does not specify an incremental adoption plan. Without a clear rule that surfaces must be migrated one-by-one—and that unchanged surfaces may continue to delegate to the old resolver as an oracle during the transition—the first slice will balloon into a massive, un-shippable pull request touching dozens of files across `cmd/gc`, `internal/api`, `internal/worker`, and `internal/mail` simultaneously.

**Required change:**
1. Add an explicit incremental-adoption contract for Slice 1.
2. State that the first implementation bead must deliver only the core candidate classifier, one compatibility adapter (e.g., API target resolver), and its parity tests.
3. Specify that other surfaces are migrated in subsequent, independent, shippable beads behind the legacy resolver oracle.

---

## Major Risks

### [Major] Shared vocabulary checkpoints contain provisional field lists that may freeze too early

The vocabulary checkpoints table (`DESIGN.md:100–108`) enumerates required fields for types like `RuntimeStartIntent` and `SessionFactEvent` which belong to Slices 3 and 4+. 

Because these slices have not yet been implemented, these detailed field lists are speculative estimates. If they are treated as frozen commitments, implementers will be forced to land these fields early or implement them exactly as listed even if actual slice development reveals a simpler, narrower vocabulary is sufficient.

**Required change:**
1. Add a clear disclaimer to the vocabulary checkpoints table stating that all field lists for non-Slice 1 types (Slices 2+) are provisional scaffolding and estimates, not commitments.
2. State that a slice may introduce fewer fields than enumerated when its first caller needs fewer, and that the lists represent upper bounds rather than rigid constraints.

---

## Missing Evidence

- **Type-Graduation Protocol:** There is no statement of where provisional, slice-local types live before they reach the "two exact-use callers" threshold required to graduate into shared `internal/session` exports. Without this, provisional types might be placed directly in the public API prematurely.
- **`closeFailedCreateBead` Exclusion in mutation landscape:** The mutation landscape table (`DESIGN.md:158–180`) still omits `closeFailedCreateBead` (`cmd/gc/session_beads.go:1736-1750`), which writes `pending_create_claim`, `pending_create_started_at`, and `sleep_intent` in a single call. Without mapping this cross-family writer, our static guard allowlist is incomplete.

---

## Required Changes

1. **Decouple Event Payloads from Speculative Sourcing:** 
   - State that in-process event payloads remain thin diagnostic/acceleration signals carrying minimal subject identity.
   - Move generation/token/idempotency requirements to a deferred outbox section.
   - Relabel the reaction matrix's "Duplicate key" column to refer to scan-side or consumer-computed keys.
2. **De-couple Mutation Static Guard from Slice 1:** 
   - Require the Go AST mutation-bypass scanner only when mutation-owning slices (Slice 2+) begin.
   - For Slice 1, require only a target-classification adoption guard.
3. **Establish an Incremental Surface Adoption Plan for Slice 1:** 
   - State that the classifier is implemented first and adopted surface-by-surface (e.g., API resolver first) in independent, shippable beads, using the old resolver as an oracle where necessary.
4. **Mark Future Slices' Field Lists as Provisional:** 
   - Label vocabulary lists for Slices 2+ as provisional scaffolding that may shrink or adapt at slice-start.
5. **Add Type-Graduation Rules:** 
   - Document that provisional fact/result types must remain private to their respective package or reside in a localized, non-exported space until the two-caller promotion threshold is achieved.
6. **Add `closeFailedCreateBead` to Mutation Landscape:** 
   - Include `closeFailedCreateBead` in the mutation landscape table (`DESIGN.md:158–180`) and allocate it to its appropriate converting slice.

---

## Questions

- **Provisional Type Location:** Where should provisional, un-graduated fact and result types be declared before they meet the two-caller promotion threshold?
- **First Delegated Caller:** Which target-resolution surface is intended to be the first delegated caller: API live target, `ResolveSessionID`, mail, or assignee normalization?
- **Current Event Constants:** Does TR-007 require changing any current event payload in Slice 1, or is it only a constraint on later command/event slices?
