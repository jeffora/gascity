# Design Review Synthesis

## Overall Verdict: block

The design remains blocked because the reviewers found unresolved contracts at the exact boundaries the refactor intends to move: target identity, session mutation ownership, worker/API/CLI routing, parity proof, event recovery, and reconciler/runtime fact ownership. The first read-only target-classification slice is still viewed as directionally sound, but it is not decomposable until the design pins result taxonomy, adopting surface, behavior parity, diagnostics, and no-side-effect rules.

## Consensus Strengths
- Multiple personas agree the desired direction is right: caller-gathered facts, session-owned decisions, and operation-specific facts are preferable to scattered raw metadata interpretation.
- The read-only target-classification first slice is broadly acceptable if it stays side-effect free, names one adopting surface, and preserves caller-specific behavior through policy adapters.
- Reviewers support factual post-commit events and durable scan convergence as the right recovery model; the missing piece is making that rule normative in the active design.
- The design correctly avoids a monolithic `SessionFacts` facade and largely defers durable event-log work, which reduces the risk of recreating the previous overlarge contract system.
- Most disagreement is about severity and proof shape, not direction: reviewers converge on smaller active contracts, live proof, and guarded migration rather than a large speculative preflight package.

## Critical Findings

### [Blocker] Target Classification Contract Is Not Defined
**Sources:** Amara Diallo / Claude, Codex, DeepSeek; Sarah Chen / Claude, Codex, DeepSeek; Ingrid Holm / Claude, Codex, DeepSeek; Natasha Volkov / Claude, Codex, DeepSeek
**Issue:** The design does not define typed result kinds, selected identity authority, liveness/closedness, configured-name state, candidate ordering, conflict and ambiguity semantics, retryability, materialization or repair eligibility, or per-surface precedence for API, CLI, mail, extmsg, worker handles, logs, transcript, sling, and assignee-adjacent paths. Current outer resolver chains include configured-name conflicts, `template:` rejection, exact bead ID precedence, API path aliases, CLI qualified-alias basename fallback, allow-closed behavior, and materialization/reopen differences that are not protected by the first-slice description.
**Required change:** Add a Target Classification Contract with typed result/candidate structures, a no-side-effect classifier rule, closed/historical match handling, and a per-surface precedence matrix. Name the first adopting surface and require characterization tests before delegation plus parity tests after delegation.

### [Blocker] Behavior Parity And Requirements Proof Are Not Enforceable
**Sources:** Natasha Volkov / Claude, Codex, DeepSeek; Amara Diallo / Claude, Codex, DeepSeek; Sarah Chen / Claude, Codex, DeepSeek; Ravi Krishnamurthy / Claude, Codex, DeepSeek
**Issue:** The active requirements ledger contains stale or missing evidence for reconciler rows, and the backlog slices are not mapped to every active `SESSION-*` scenario row, caller surface, proof command, or exact test selector. Reviewers also found unrowed API/CLI resolver overlays and no owner-approval gate for requirements changes, so an implementation could relabel behavior drift as a requirements update.
**Required change:** Add a scenario parity matrix covering every touched active row and surface, repair or explicitly retire stale proof with owner approval, bind proof to exact tests/assertions and commands, and require a durable owner-approval artifact for behavior-changing requirements edits.

### [Blocker] Session Mutation Ownership Lacks A Guarded Writer Boundary
**Sources:** Elena Marchetti / Claude, Codex, DeepSeek; Ravi Krishnamurthy / Claude, Codex, DeepSeek; Sarah Chen / Claude, Codex, DeepSeek; Ingrid Holm / Claude, Codex, DeepSeek
**Issue:** The design does not provide a source-complete ledger of lifecycle and identity writers outside `internal/session`, and it does not prove a failing-build guard against new external `SetMetadata*`, `MetadataPatch`, patch-constructor, or repair/doctor/migration bypasses. Repair paths such as `RepairEmptyType` remain ownerless, and old raw writers can coexist with new validating commands over the same key families.
**Required change:** Add a Mutation Ownership Ledger and shrink-only CI guard before any mutation-owning slice. Enumerate allowed exceptions by exact path/function, owner, reason, and retirement condition; decide whether exported patch constructors are transitional or stable; and give repair writes an audited owner with persistence-error and trace/doctor evidence.

### [Blocker] Mutating Commands Need Atomic Commit And Recovery Semantics
**Sources:** Takeshi Yamamoto / Claude, Codex, DeepSeek; Amara Osei / Claude, Codex, DeepSeek; Liam Okonkwo / Claude, Codex, DeepSeek; Ingrid Holm / Claude, Codex, DeepSeek
**Issue:** Later wake, close, identity-retirement, runtime-start, drain, and reconciler slices do not have per-operation contracts for immutable facts, snapshot freshness, commit-time revalidation, conditional or tokened writes, runtime intent execution, event emission, rollback, partial-failure state, and reconciliation owner. Existing operation ordering differs by path, so a generic "facts -> decider -> command" shape is not enough.
**Required change:** Define an Atomic Command Contract for every mutation-moving slice before decomposition. Each contract must state facts, preconditions, stale/rejected outcomes, write primitive or fence, event ordering, runtime ordering, partial states, crash recovery owner, and characterization tests for rollback and skipped/duplicate event delivery.

### [Blocker] Worker/API/CLI Boundary Routing Is Unsettled
**Sources:** Sarah Chen / Claude, Codex, DeepSeek; Ravi Krishnamurthy / Claude, Codex, DeepSeek; Elena Marchetti / Claude, Codex, DeepSeek
**Issue:** The design does not decide how new session-owned lifecycle operations interact with the in-flight worker-boundary migration. Wake currently uses direct store-level session calls while close routes through `worker.Handle.CloseDetailed`; API construction/direct-create exceptions still exist; and CLI/API fallback and generated-client paths can diverge unless route-level parity is pinned.
**Required change:** Add a "Relation to the worker boundary" section naming which operations stay store-level, which move through `worker.Handle`, and how API exceptions are retained or retired. Add an API/CLI boundary inventory, exception schema, static guard coverage, and route/fallback parity tests for session-affecting commands.

### [Blocker] Reconciler And Runtime Fact Ownership Is Underspecified
**Sources:** Liam Okonkwo / Claude, Codex, DeepSeek; Takeshi Yamamoto / Claude, Codex, DeepSeek; Ingrid Holm / Claude, Codex, DeepSeek
**Issue:** Awake, drain, hold, quarantine, provider-health, progress, pool sizing, work demand, waits, runtime missing, idle sleep, and attached/pending behavior are not separated into session-owned lifecycle eligibility versus reconciler-owned scheduling and policy. Runtime unknown/stale/partial/provider-error facts are not specified for destructive actions, and runtime intent risks smuggling provider policy into `internal/session`.
**Required change:** Add a reconciler/session split matrix that classifies every input and transition. Keep work demand, scheduling, budgets, provider health, progress policy, and alerts outside `internal/session`; define provider-neutral runtime intent fields and destructive-action safety rules for unknown or stale facts.

### [Major] Event And Recovery Rules Are Not Normative In The Active Design
**Sources:** Amara Osei / Claude, Codex, DeepSeek; Takeshi Yamamoto / Claude, Codex, DeepSeek; Ingrid Holm / Claude, Codex, DeepSeek
**Issue:** Reviewers agree session events should be factual post-commit hints and durable scans should converge safety-critical reactions, but the active design does not assign event emission ownership or recovery authority for close, work release, wake, drain, runtime-start, provider-health, and event-bearing slices. Crash-after-commit windows could drop release or trace behavior if callers and commands split responsibility.
**Required change:** Add an active Event and Recovery Contract. State post-commit emission, stable identity payloads, facts-not-commands semantics, durable scan convergence, critical versus best-effort reactions, typed payload/registry obligations, and crash-window tests for close/retire and other event-bearing slices.

### [Major] Operator Diagnostics And Cost Budgets Are Missing
**Sources:** Ingrid Holm / Claude, Codex, DeepSeek; Sarah Chen / Claude, Codex, DeepSeek; Liam Okonkwo / Claude, Codex, DeepSeek
**Issue:** Extracted classifiers and deciders could hide the reason a session was blocked, woken, drained, closed, or rejected. The first target-classification slice lacks a diagnostic schema and query-count budget, and later command results lack fields needed by `gc trace`, doctor, inspect, API, CLI, and tests. Reviewers also require source proof that lookups avoid all-session scans or expensive subprocess loops on hot paths.
**Required change:** Add an Observability and Cost section with typed classifier and command diagnostic results, trace reason/outcome mappings, API/CLI rendering parity, lookup/query/subprocess-count budgets, and hot-loop constraints for reconciler fact materialization.

### [Minor] Scope Authority And Vocabulary Still Need Tightening
**Sources:** Kwame Asante / Claude, Codex, DeepSeek; Amara Diallo / Claude, Codex, DeepSeek; Amara Osei / Claude, Codex, DeepSeek
**Issue:** `DESIGN_REVIEW_NOTES.md` is still ambiguous authority, and future vocabulary such as wake, close, runtime-start, event-log, generic command/result, and broad facts can leak into the first classification slice. Later backlog items also lack entry criteria that prevent future slices from starting before target classification has landed with one adopter and proof.
**Required change:** State that archived review notes are historical and non-normative unless text is copied into the active design. Add a Slice 1 vocabulary guard, name the first adopter, add repeated-exact-use criteria for shared facts, and require prior-slice proof before later vocabulary expands.

## Disagreements
- Several persona lanes split between `approve-with-risks` and `block`, but the blocking reviewers identified missing contracts that are prerequisites for decomposition, not merely implementation cautions. I therefore treat the global verdict as `block`.
- Event reviewers mostly allow the read-only target-classification slice to proceed without event tables, while command/reconciler reviewers block later mutation slices. Assessment: target classification can avoid event work, but the design still needs a minimal active event/recovery rule now and per-event tables before event-bearing slices.
- Reviewers disagree on whether configured-name lookup, API path aliases, CLI qualified alias fallback, and materialization belong inside the classifier or adapters. Assessment: either placement can work if the design represents them in typed facts or explicit adapter steps and pins current precedence with tests.
- Reviewers disagree on proof heft. Some favor a small active contract; others point to richer preflight artifacts. Assessment: do not resurrect a large speculative artifact system, but do require a source-derived writer ledger, parity matrix, and failing-build guard before behavior-moving work.
- DeepSeek/Gemini reviews in multiple lanes appear to have reviewed an Attempt 14 Slice 0 preflight shape while Claude/Codex saw a simpler live design. Assessment: treat those details as desired proof-shape signal, not as evidence that the current design already contains the artifacts.

## Missing Evidence
- Ten persona syntheses for attempt 14 were not present under `.gc/design-reviews/ga-unpr2y/attempt-14/persona-syntheses`; the closed persona beads recorded current synthesis output paths under `.gc/design-reviews/ga-unpr2y/attempt-1/persona-syntheses`. This global synthesis used those recorded output paths, but the artifact-path mismatch should be fixed by the workflow.
- Complete scenario parity matrix for all active `SESSION-*` rows, touched surfaces, exact proof selectors, proof commands, and owner-approved behavior changes.
- Source-complete mutation writer ledger and shrink-only guard for lifecycle/identity metadata writers, external patch-map application, patch constructors, repair, doctor, migration, and test exceptions.
- Target-classification typed result taxonomy, per-surface precedence matrix, no-side-effect proof, first adopter, and parity fixtures.
- Worker-boundary/API/CLI inventory covering wake, close, identity retirement, API exceptions, generated-client/fallback behavior, Huma response compatibility, OpenAPI/generated type checks, and user-facing projection reads.
- Atomic command contracts for wake, close, retire, drain, runtime start, provider-health, runtime-missing, and reconciler lifecycle transitions.
- Event emitter/consumer inventory, post-commit event ownership, durable recovery authority, and crash-window tests for safety-critical reactions.
- Reconciler/session split matrix for lifecycle eligibility, work demand, waits, holds, quarantine, health/progress, pool sizing, runtime facts, and destructive-action safety.
- Observability and cost budgets for classifier/decider diagnostics, trace mappings, query counts, subprocess costs, hot-loop fact materialization, and API/CLI rendering.

## Recommended Changes
1. Add the Target Classification Contract and choose one first adopting surface; require no side effects and pin current precedence/error behavior with characterization and parity tests.
2. Repair the behavior parity foundation: map all active `SESSION-*` rows and touched surfaces to exact proof commands/selectors, mark or retire stale proof with owner approval, and add a behavior-change approval artifact.
3. Add the Mutation Ownership Ledger plus shrink-only CI guard for lifecycle/identity writers, patch-map application, exported patch constructors, and exact exception rows.
4. Add the worker-boundary/API/CLI routing contract, including wake versus close routing, API exceptions, Huma error compatibility, generated-client/fallback behavior, and route-level parity proof.
5. Define Atomic Command Contracts for every mutation-moving slice before decomposition, including stale facts, write fences, runtime/event ordering, partial states, and crash recovery.
6. Add the reconciler/session split matrix and provider-neutral runtime intent contract so scheduler, work-demand, health, progress, budget, and provider policy stay outside `internal/session`.
7. Add the active Event and Recovery Contract and gate event-bearing slices on emitter/consumer inventory, durable convergence tests, typed payload registry proof, and OpenAPI/SSE checks when public wire events change.
8. Add Observability and Cost requirements for structured diagnostics, `gc trace` mappings, doctor/inspect/API/CLI rendering, query-count budgets, and large-city hot-loop constraints.
9. Clarify that `DESIGN_REVIEW_NOTES.md` is historical/non-normative unless copied into the active design, and add slice-entry/vocabulary guards to prevent future abstractions from entering Slice 1.
10. Fix the design-review workflow artifact bug that wrote attempt-14 persona syntheses under `attempt-1/persona-syntheses`.
