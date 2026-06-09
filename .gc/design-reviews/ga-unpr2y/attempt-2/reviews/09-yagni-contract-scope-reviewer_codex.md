# Kwame Asante - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The revision directly addresses the biggest YAGNI hazard from this lane: `internal/session/DESIGN.md:64`-`66` forbids a broad `SessionFacts` type and allows shared fact structs only after repeated exact use.
- Slice vocabulary is now explicitly staged. `internal/session/DESIGN.md:79`-`86` keeps slice 1 to target classification, defers wake facts to slice 2, and defers lifecycle/reconciler/runtime-intent vocabulary until those slices start.
- The event/recovery section avoids treating in-process events as a new primitive. `internal/session/DESIGN.md:265`-`288` keeps critical recovery on durable facts and scans, which is consistent with the current best-effort event log contract in `internal/events/events.go:1`-`9`.

**Critical risks:**
- [Major] The static guard still reads like a whole-program mutation gate for every slice, including slice 1, which is read-only target classification. `internal/session/DESIGN.md:143`-`155` says "Before a slice is considered implemented" the implementation must add or tighten a failing-build guard that detects `SetMetadata`, `SetMetadataBatch`, `Update`, `Create`, `Close`, `session.MetadataPatch` use, every temporary exception, and retirement conditions. That is correct for slices that move mutation, but it is over-scoped for `Backlog 1` (`internal/session/DESIGN.md:560`-`571`), whose explicit implementation is classifier result types and operation policy only. If implementers follow this literally, slice 1 can balloon into an AST scanner plus mutation inventory before any production mutation path changes. Required change: scope the guard requirement by field family and slice. For slice 1, require a target-classification adoption guard or parity tests that prevent new resolver fallbacks. Require the mutation bypass scanner only when a slice introduces or delegates session-owned commands for that metadata family.
- [Major] `path-alias` is promoted into the core classifier vocabulary without preserving how narrow and surface-specific it is today. The design's result table names `path-alias` as a classifier kind at `internal/session/DESIGN.md:177`, and the proof asks for collision tests for every result kind at `internal/session/DESIGN.md:207`-`209`. The current implementation is not a general session identity class: `internal/api/session_resolution.go:365`-`391` describes an API-specific pool Title fallback, active/awake/legacy-empty-state only, excluding named-session beads and using a newest-created tiebreaker for duplicate Titles. If this becomes a session-owned result kind without those constraints, a temporary API compatibility shim becomes durable session vocabulary. Required change: either keep path-alias fact gathering and acceptance entirely adapter-owned, with the classifier receiving only an adapter-supplied candidate, or spell out the exact current constraints in the slice 1 contract and state that no other surface may opt in without a requirement row.
- [Minor] TR-007 can still pull future event sourcing concerns into current command shape if read too broadly. The design says event names and payloads should be shaped so they "can later become durable event-log facts" at `internal/session/DESIGN.md:536`-`541`, while the current event package documents best-effort observability at `internal/events/events.go:1`-`6` and CI only requires typed payload registration for known events at `internal/events/events.go:129`-`142`. The recovery section mostly prevents overreach, but TR-007 should explicitly apply only when a slice emits or changes session events. A read-only classifier or command that does not emit events should not have to carry generation, idempotency, or future outbox fields just to satisfy a hypothetical event-sourcing path.

**Missing evidence:**
- There is no explicit statement that slice 1 can be completed without adding the mutation static guard. The document implies the opposite by putting the guard under the global mutation boundary and using "Before a slice is considered implemented."
- The path-alias proof list does not cite the exact API-only constraints from `resolveLiveSessionByPathAlias`: state filter, named-session skip, and most-recent tiebreaker.
- TR-007 does not name a concrete current slice that emits a new session event, so reviewers cannot tell whether it is an active acceptance criterion or only a design constraint for future event-emitting slices.

**Required changes:**
- Change the static-guard section to say the guard is required before a mutation-owning slice is complete, and that each slice tightens only the field family it owns. Add a separate slice 1 guard/test requirement for "no new factory/config fallback in target classification" if a guard is needed there.
- Amend the path-alias row to keep path aliases as adapter-scoped compatibility candidates, or add its current API constraints directly to the classifier contract and operation policy matrix.
- Narrow TR-007 acceptance criteria to event-emitting slices: "If this slice emits or changes session events..." Keep event sourcing as compatibility pressure, not a required field set for non-event work.

**Questions:**
- Is `path-alias` intended to be a first-class session-domain target kind, or only an API compatibility candidate that the classifier can rank after live session identity?
- Should the first implementation bead for target classification be allowed to land before any mutation-bypass scanner exists?
- Which planned slice is expected to emit the first new session event, if any?
