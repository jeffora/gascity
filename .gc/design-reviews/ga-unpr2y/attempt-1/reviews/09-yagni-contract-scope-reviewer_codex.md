# Kwame Asante - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design now frames the work as a small refactor plan, not an architecture program. It explicitly says no behavior-moving slice is approved and Slice 0 may only collect evidence, validate freshness, and repair or retire stale requirement evidence (`internal/session/DESIGN.md:12`).
- Slice 0 is fenced as non-mutating: it may not move callers, introduce public command APIs, change target resolution, materialize sessions, repair metadata, add event payloads, or alter reconciler policy (`internal/session/DESIGN.md:135`).
- The design separates universal evidence from per-slice preflight and says Slice 0 may define schemas but must not pre-approve future behavior (`internal/session/DESIGN.md:143`).
- Vocabulary lifecycle is a real YAGNI guard. Future terms such as command result, runtime intent, session fact event, and generic conflict stay `provisional` until a production caller proves exact fields (`internal/session/DESIGN.md:585`).
- The API shape guidance rejects the two biggest facade paths: no broad `SessionService`, no one large `SessionFacts`, no generic command bus, and no event sourcing as the first step (`internal/session/DESIGN.md:601`).
- The backlog is sliced in the right order: first read-only API query classification, then other target-classification surfaces one at a time, then wake, close/identity, runtime start, and only narrow reconciler facts (`internal/session/DESIGN.md:654`).

**Critical risks:**
- [Major] Slice 0 is large enough to become an accidental architecture program if the close contract is weak. The artifact list includes baseline, contract, boundary inventory, symbol inventory, scenario parity, target classification, command appliers, boundary matrix, route inventory, worker exceptions, diagnostics, and vocabulary checkpoints (`internal/session/DESIGN.md:150`). That breadth is justified as evidence, but only if `SLICE0_CONTRACT.yaml` prevents behavior claims, exported APIs, command result types, runtime intent types, new payload shapes, and future-slice compatibility promises from landing as approved implementation surface.
- [Major] The target-classifier result schema is already fairly rich. Fields like `match_vectors`, `bead_state`, `config_state`, `diagnostics`, and `terminal_error` are plausible for the first API query adopter (`internal/session/DESIGN.md:253`), but they should remain private to that slice until a real caller proves each field. Otherwise the classifier becomes the first broad facade under a different name.
- [Major] Event compatibility can still slide into speculative event sourcing. The design correctly says events are post-commit facts and not durable truth (`internal/session/DESIGN.md:496`), but rows for stable identities, idempotency, durable scan owner, and crash tests must stay tied to current in-process event needs and current recovery scans. They should not introduce an outbox, event-sourced state model, or generic committed-fact payload until separately approved.
- [Minor] `DESIGN_REVIEW_NOTES.md` remains a large nearby archive. The design says it is non-normative unless copied here (`internal/session/DESIGN.md:10`), but Slice 0 and bead generation must enforce that rule so implementers cannot cite historical notes as acceptance criteria.

**Missing evidence:**
- No `VOCABULARY_CHECKPOINTS.yaml` exists yet proving which terms are `documented`, `private`, `provisional`, or `delegating`.
- No `TestVocabularyCheckpoints` or equivalent guard exists yet to fail future-only exported types, broad optional envelopes, unused result fields, or generic command/event abstractions.
- No Slice 0 close contract exists yet showing that schemas and inventories are evidence artifacts only, not authorization for behavior-moving code.
- No first-adopter proof exists yet that every target-classifier field is demanded by API query parity rather than included for later surfaces.

**Required changes:**
- No design-text blocker remains in this persona lane. The document now states the right scope controls.
- Before Slice 0 closes, make `SLICE0_CONTRACT.yaml` reject behavior-moving changes, new public command APIs, event payload additions, exported future-slice types, and implementation code that is not needed to build or validate the evidence artifacts.
- Add vocabulary checkpoint rows for every new shared term and fail the build if a term marked `provisional` appears in public API, generated clients, event payloads, cross-slice contracts, or exported package APIs.
- Keep first-adopter target classification private and operation-specific. Promote a field to shared vocabulary only when a production caller delegates to it and the checkpoint row names exact fields, non-goals, tests, and expansion rules.
- Require future slices to delete or shrink caller logic after parity is proven; do not let new deciders accumulate behind old duplicated behavior indefinitely.

**Questions:**
- Which Slice 0 artifacts are allowed to add code helpers, and which must remain data/schema/test artifacts only?
- For target classification, which fields are strictly required by API query-side parity on the first adopter, and which are provisional for later CLI/mail/extmsg surfaces?
- What exact test fails if a future slice introduces `SessionFacts`, `SessionService`, `RuntimeIntent`, or a generic command result before a production caller needs it?
