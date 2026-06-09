# Amara Osei - Codex

**Verdict:** block

Reviewed `internal/session/DESIGN.md`, which matches `.gc/design-reviews/ga-unpr2y/attempt-4/design-before.md`, plus `internal/session/REQUIREMENTS.md`, scoped session instructions, and current event/session call sites.

**Top strengths:**
- The design now says the essential rule plainly: no safety-critical reaction may depend only on at-most-once in-process event delivery.
- The per-event matrix separates accelerator, diagnostic, retryable, and observability-only reactions, and it specifically keeps `SessionDrainAckedWithAssignedWork` out of the load-bearing recovery path.
- The API/event registration gate is correctly anchored on typed payload registration and `TestEveryKnownEventTypeHasRegisteredPayload`.

**Critical risks:**
- [Blocker] Runtime start still has an unowned crash window between provider start and durable correlation. The design orders runtime start as `commit creating/generation/token -> start provider -> commit active/session key -> event/trace`, while the recovery authority claims "provider runtime identity." For providers that need a generated session key, current code persists `session_key` before building the provider command in `cmd/gc/session_lifecycle_parallel.go`; if a slice moves `session_key` to the post-provider active commit and the process crashes after provider start but before that commit/event, the controller may not have the durable provider correlation needed to resume, stop, or reconcile the runtime. The design must either make provider correlation fields such as `session_key` part of the prepare record when required, or prove an alternate durable observation can recover the runtime after that crash.
- [Major] Close/work-release convergence is still a placeholder, not an implementation contract. The document says close does `commit close/retire facts -> release work scan -> event/trace`, and later says to "define whether" the close command performs synchronous work release or returns a post-commit fact for an idempotent scan. That leaves the exact recovery actor, trigger, idempotency key, store-ref scope, and retry cadence undecided for `SESSION-WORK-001` and `SESSION-WORK-002`. A crash after bead close but before the release scan must have a named scan path and test before this can be decomposed safely.
- [Major] Existing session events with `events.NoPayload` are not reconciled with the new payload requirements. `internal/api/event_payloads.go` currently registers `SessionWoke`, `SessionDraining`, `SessionUndrained`, `SessionQuarantined`, `SessionIdleKilled`, `SessionMaxAgeKilled`, `SessionSuspended`, `SessionUpdated`, and `SessionStranded` as no-payload events, while the design's matrix requires stable IDs, generation/token, reasons, projection facts, or assigned work refs for stale-event suppression and idempotent replay. The design needs an explicit legacy-vs-upgraded payload migration rule so implementation does not silently treat no-payload events as satisfying the new contract.

**Missing evidence:**
- Crash-after-provider-start-before-active-commit recovery proof for providers that use `session_key` or another runtime correlation key.
- A named close/work-release recovery loop and tests for crash after durable close, duplicate close scan, missed event, partial work query, and cross-store assigned work.
- A payload compatibility plan for existing no-payload session events, including API/SSE/OpenAPI effects and stale-event handling.
- Proof that command success does not depend on event recorder success, while still recording enough event emission result and recovery-path evidence for `gc trace` and doctor output.

**Required changes:**
- Move required provider correlation fields into the runtime-start prepare record, or document and test an equivalent durable recovery source for the provider-start-before-active-commit crash window.
- Replace the close/work-release "define whether" placeholder with a concrete contract: synchronous cascade vs controller scan, owner function/package, idempotency key, store-ref query shape, retry trigger, and failure behavior.
- Add explicit migration requirements for current no-payload session events: which remain observability-only legacy events, which gain typed payloads, and which tests protect API/SSE compatibility.
- Add recovery tests that simulate skipped event delivery and process crash after commit for wake, close/work release, drain, and runtime start.

**Questions:**
- For providers using `SessionIDFlag`, is `session_key` intended to be a prepare-time durable identity or an active-commit result?
- Which component owns the work-release scan after close: the session command itself, the controller tick, or a subscriber scheduled from returned post-commit facts?
- Are no-payload session events allowed to remain permanently for dashboard-only diagnostics, or must all session events named in the matrix gain typed payloads before the corresponding slice lands?
