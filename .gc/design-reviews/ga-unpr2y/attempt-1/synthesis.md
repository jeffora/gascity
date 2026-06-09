# Design Review Synthesis

## Overall Verdict: block

The design is directionally aligned with moving session behavior into typed,
session-owned commands and pure decision logic, but the current document is not
safe to implement. Nine of ten persona syntheses recommend `block`; the common
failure mode is that the design names the desired boundary without specifying
the enforceable contracts, migration sequence, concurrency model, recovery
guarantees, and parity proof needed to preserve current behavior.

## Consensus Strengths
- Multiple reviewers agree the functional-core direction is correct: session
  decisions should consume immutable facts and return typed outcomes rather than
  reading stores, runtime, or config directly.
- The existing code already contains useful migration anchors, especially
  `AwakeInput`/`ComputeAwakeSet`, `LifecycleInput`/`ProjectLifecycle`, existing
  transition validation, and existing manager methods that can inform the
  command/applier model.
- The design correctly identifies the central risks around direct metadata
  mutation, target classification, event payloads, and preserving behavior
  across API, CLI, reconciler, mail, worker, and external messaging surfaces.
- Reviewers found useful precedent in the active worker-boundary guard and
  existing trace vocabulary, but both must be extended or explicitly mapped
  rather than bypassed.

## Critical Findings

### [Blocker] Mutation Boundary Is Not Enforceable
**Sources:** Elena Marchetti, Sarah Chen, Ravi Krishnamurthy, Ingrid Holm
**Issue:** The design says lifecycle and identity metadata mutations should move
into session-owned commands, but it does not provide a current writer inventory,
a slice-by-slice retirement plan, or a failing-build guard for new production
bypasses. Exported patch helpers and direct metadata writers remain reachable
from `cmd/gc`, `internal/api`, reconciler paths, doctor/repair flows, and other
production code.
**Required change:** Add a current mutation landscape table covering each
non-test lifecycle/identity writer, field family, current mutation path, target
path, owner slice, and exception status. Add a static guard that prevents new
production files outside approved paths from writing boundary-owned keys or
applying session patch maps, with bounded exceptions for repair, migration,
fixtures, and low-level conformance utilities.

### [Blocker] Scenario Parity Proof Is Stale And Too Broad
**Sources:** Natasha Volkov, Sarah Chen, Liam Okonkwo, Ingrid Holm
**Issue:** Requirements and design proof cite missing tests such as
`cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and
`cmd/gc/session_progress_test.go`. Backlog slices cite broad files rather than
exact `SESSION-*` rows, test names, commands, API response checks, CLI
stdout/stderr, error shapes, generated schema impacts, and trace expectations.
**Required change:** Add a scenario traceability matrix for every backlog slice:
touched `SESSION-*` rows, current behavior proof, new or updated proof, allowed
behavior change, exact tests or commands, and proof freshness requirements.
Restore, replace, or defer stale proof before implementing affected slices.

### [Blocker] Command Atomicity And Stale-Fact Defense Are Undefined
**Sources:** Takeshi Yamamoto, Ravi Krishnamurthy, Liam Okonkwo, Ingrid Holm
**Issue:** The design says commands validate current state before mutation, but
it does not choose a concrete stale-fact defense such as revision tokens,
generation preconditions, metadata hashes, identity-continuity checks, locks
with reread, or another compare-and-set substitute. It also leaves multi-key
metadata writes, partial external-store failure, conflict reasons, retry
semantics, and durable mutation ordering unspecified.
**Required change:** Add a command atomicity contract for each command:
precondition fields, commit validation, written keys, write ordering or atomicity
assumption, conflict reasons, retryability, partial-commit repair behavior, and
tests for concurrent close/wake/drain/identity/runtime races.

### [Blocker] Runtime Intent, Event, And Failure Ordering Are Not Specified
**Sources:** Takeshi Yamamoto, Amara Osei, Liam Okonkwo, Ingrid Holm
**Issue:** Decisions that produce both durable mutations and runtime intents do
not have defined ordering for commit, event emission, runtime execution,
subscriber reaction, rollback, or reconciliation. A runtime failure after a
durable mutation or a crash after mutation but before in-process event delivery
can strand work, double-start sessions, lose cleanup, or hide missed reactions.
**Required change:** Add per-operation ordering tables for close, wake, drain,
identity retirement, runtime start/stop, and work release. State the authoritative
recovery source for each reaction: durable outbox, replayable event log,
idempotent reconciler scan from durable facts, or retained synchronous cascade
until durable delivery exists.

### [Blocker] Event Delivery And Subscriber Convergence Are Unsafe
**Sources:** Amara Osei, Ingrid Holm, Takeshi Yamamoto
**Issue:** The design treats `SessionEvent` as if event delivery can safely
drive cross-domain reactions, but the current delivery substrate is not specified
as durable or replayable. Safety-critical work release, drain safety, identity
retirement cleanup, and runtime reconciliation cannot depend solely on at-most
once in-process delivery.
**Required change:** Add an event delivery and convergence contract. Classify
each subscriber domain as critical, retryable, best-effort, or observability-only;
define payload identities and idempotency keys; document duplicate,
out-of-order, skipped, and crash-after-commit behavior; and map proposed session
events to existing typed `events.*` constants and dashboard/SSE/event-log
visibility.

### [Blocker] Target Classification Lacks Precedence And Operation Policy
**Sources:** Amara Diallo, Sarah Chen, Kwame Asante, Natasha Volkov
**Issue:** TR-006 lists classifier categories but does not define the
authoritative precedence order across exact bead IDs, configured named sessions,
live `session_name`, live aliases, path aliases, closed lookups, historical
aliases, ordinary config targets, `template:` targets, not-found, and ambiguity.
It also does not separate pure classification facts from per-operation
permissions for API, CLI, mail, extmsg, nudge, attach, inspect, materialization,
and assignee normalization.
**Required change:** Add a target classification contract with an ordered
precedence table, typed result fields, negative result kinds, conflict details,
materialization facts, and a per-operation permission matrix. Split classifier
implementation from surface-policy adapter adoption and require adversarial
parity tests during migration.

### [Blocker] Reconciler Boundary And Runtime Facts Are Underspecified
**Sources:** Liam Okonkwo, Sarah Chen, Takeshi Yamamoto, Ingrid Holm
**Issue:** The design does not make the existing pure reconciler deciders the
migration baseline and does not define fact completeness, provenance, partial
query handling, stale runtime observations, desired-running composition, or
fail-closed rules. Destructive decisions such as drain, rollback, restart,
slot closure, circuit-open wake suppression, and wake suppression can regress if
partial reads collapse to false negatives.
**Required change:** Revise the reconciler backlog to anchor on
`ComputeAwakeSet`, `ProjectLifecycle`, and current pool desired-state code. Add
fact-family contracts for partial, unknown, stale, and probe-derived facts; pin
desired-running precedence; and define state ownership for circuit breakers,
health episodes, progress signatures, restart counts, and budget gates.

### [Blocker] Worker Boundary And Migration Sequencing Collide
**Sources:** Sarah Chen, Ravi Krishnamurthy, Elena Marchetti
**Issue:** The active worker-boundary migration requires production `cmd/gc`
lifecycle operations to route through `worker.Handle`, while the design also
asks callers to adopt session-owned commands. Shared files such as
`cmd/gc/session_beads.go`, `internal/api/session_resolution.go`, and
`internal/api/session_manager.go` have no ordered ownership, target replacement,
or exception retirement plan.
**Required change:** Add a migration and coexistence plan that phases each slice
through command introduction, first caller delegation, all caller delegation,
guard tightening, and legacy path removal. For each field family, either convert
all production writers in the slice or require every surviving writer to delegate
to the same validating command/applier path.

### [Blocker] Operability Contract Is Missing
**Sources:** Ingrid Holm, Liam Okonkwo, Sarah Chen, Amara Osei
**Issue:** `SessionDecision` and `SessionConflict` do not define operator-facing
diagnostic fields, and proposed event diagnostics are not mapped to the existing
trace substrate. Without stable site, reason, outcome, identity, blocker,
suppressed wake cause, fact freshness, conflict, event emission, and subscriber
failure evidence, the refactor can preserve behavior while making `gc trace`,
doctor, logs, API, and CLI conflicts less useful.
**Required change:** Add an operability contract mapping every decision and
conflict variant to existing or new trace site/reason/outcome codes, API/CLI
fields, logs, and doctor output. Include examples and tests for accepted,
rejected, blocked, no-op, drained, closed, ambiguous, and missed-event paths.

### [Major] Vocabulary Scope Is Too Broad For Slice 1
**Sources:** Kwame Asante, Amara Diallo, Liam Okonkwo, Ingrid Holm
**Issue:** The design introduces `SessionFacts`, `SessionDecision`,
`SessionMutation`, `RuntimeIntent`, `SessionEvent`, and `SessionConflict` before
the first slice needs most of them. A singular `SessionFacts` type risks
becoming the broad facade the design says it will avoid.
**Required change:** Add a per-slice vocabulary and fact-family map. Slice 1
should introduce only target-classification facts, classifier results, and the
needed conflict/error shape. Replace singular `SessionFacts` with
operation-specific fact families unless multiple implemented deciders require the
same exact field set.

### [Major] Existing Session Types And Errors Are Not Mapped
**Sources:** Kwame Asante, Ingrid Holm, Takeshi Yamamoto, Amara Diallo
**Issue:** Proposed vocabulary does not say whether it replaces, wraps, extends,
or coexists with existing `LifecycleInput`, `RuntimeFacts`,
`NamedIdentityInput`, `Transition`, `IllegalTransitionError`,
`ErrSessionNotFound`, `ErrAmbiguous`, trace codes, and typed event constants.
Callers may preserve raw-state interpretation because the new terms do not map
to current contracts.
**Required change:** Add a vocabulary mapping table for existing requirements,
code types, errors, trace codes, and event constants. Name the slice that
introduces or replaces each term.

### [Major] API, CLI, And Typed-Wire Obligations Are Under-Specified
**Sources:** Sarah Chen, Natasha Volkov, Amara Diallo
**Issue:** API and CLI compatibility proof does not cover response bodies, error
bodies, generated OpenAPI/client artifacts, dashboard checks, CLI output, exit
codes, and Huma typed-wire obligations for touched scenarios. Raw lifecycle
state readers also remain in API paths without a structural projection guard.
**Required change:** Add per-slice API/CLI compatibility gates. For API-visible
changes, require Huma-registered typed outputs, generated OpenAPI/schema/client
updates where applicable, `dashboard-check` when affected, and proof that API and
CLI surfaces remain projections over the session object model.

### [Major] Fact Materialization And Fan-Out Lack Performance Bounds
**Sources:** Ingrid Holm, Liam Okonkwo
**Issue:** The design constrains deciders to no I/O but does not constrain the
adapter side that materializes facts or emits subscriber work. Existing hot-loop
protections around bulk desired-state aggregation, scale checks, cached demand
reads, and named-session scans are not preserved as explicit requirements.
**Required change:** Add a performance contract for fact readers and subscribers:
reuse per-tick snapshots and bulk reads, avoid per-session subprocess fan-out,
cap synchronous subscriber work, record fact-build and subscriber durations, and
state expected read complexity for large cities.

### [Minor] Operational Metadata Boundary Needs Clarification
**Sources:** Elena Marchetti, Ingrid Holm
**Issue:** The boundary between lifecycle/identity fields and operational repair
or diagnostic metadata is not clear. Fields such as nudge delivery metadata,
close reasons, provider session keys, circuit-breaker metadata, template repair,
and non-lifecycle repair writes may remain outside the design's guard and trace
contracts.
**Required change:** Define which operational metadata fields are governed by
the session boundary, which are repair-only exceptions, and what trace/log
evidence each direct repair write must emit.

## Disagreements
- Nine personas recommend `block`; Kwame Asante recommends
  `approve-with-risks`. The non-blocking lane accepts the general direction but
  still finds major scope and vocabulary risks, so it does not weaken the global
  block verdict.
- Several personas had internal model severity disagreements. Elena, Sarah,
  Liam, and Ingrid each report at least one model giving `approve-with-risks`
  while the synthesis blocks. These are severity disagreements rather than
  factual disagreements; the same missing contracts recur across reviewers.
- Reviewers disagree on exact remedies for stale-fact defense and event
  convergence: CAS/revision tokens, lock-plus-reread, durable outbox, durable
  event replay, synchronous cascade retention, or idempotent reconciler scans.
  The design may choose any coherent mechanism, but it must choose and document
  one before implementation.
- Reviewers disagree on path alias placement and whether classifier scope should
  include API-specific compatibility behavior. The safe resolution is to keep
  pure classification separate from surface-policy adapters and document
  ownership explicitly.
- Reviewers cite different counts for current mutation call sites. The synthesis
  does not rely on those counts; it requires a fresh current inventory that can
  be verified by reviewers and guarded by CI.

## Missing Evidence
- Current inventory of all production lifecycle and identity metadata writers,
  including exported patch helpers, `SetMetadata*`, `Update`, `Create`, `Close`,
  manager methods, API direct-create exceptions, and doctor/repair paths.
- Static guard that fails when production code outside approved paths mutates
  session-owned metadata or applies session patch maps.
- Scenario traceability matrix tying every slice to exact `SESSION-*` rows,
  runnable proof commands, precise test names, API/CLI compatibility assertions,
  trace expectations, and stale-proof replacement rules.
- Concrete stale-fact, optimistic-concurrency, multi-key write, conflict,
  partial-commit, and runtime-failure tests.
- Event delivery, recovery, ordering, duplicate, replay, and crash-after-commit
  tests for critical reactions such as work release, close, identity retirement,
  wake, drain, and runtime start/stop.
- Target classification precedence table, typed result contract, negative
  result kinds, conflict candidate details, materialization policy, and
  per-operation permission matrix.
- Reconciler migration baseline tied to current `ComputeAwakeSet`,
  `ProjectLifecycle`, pool desired-state, circuit-breaker, provider-health, and
  progress code and tests.
- Worker-boundary sequencing for shared `cmd/gc` and `internal/api` call sites,
  including end states for `internal/session.Manager` methods and exported patch
  helpers.
- Mapping from proposed session events, decisions, conflicts, facts, and
  runtime intents to existing code types, errors, trace codes, event constants,
  API/CLI outputs, and dashboard/SSE visibility.
- Performance budget and instrumentation plan for fact materialization,
  runtime probes, config scans, store reads, and subscriber fan-out across large
  cities.

## Recommended Changes
1. Add the mutation landscape inventory and static boundary guard before or with
   the first slice that changes lifecycle or identity mutation behavior.
2. Add a scenario traceability matrix and fix stale proof paths in
   `REQUIREMENTS.md` and `DESIGN.md`; defer slices whose proof does not yet
   exist.
3. Add the command atomicity contract, including stale-fact defense,
   multi-key write strategy, conflict semantics, runtime intent ordering, and
   failure recovery.
4. Add event delivery and convergence contracts that make durable facts or
   replay, not in-process delivery alone, authoritative for critical reactions.
5. Define the target classification precedence table, typed result shape, and
   per-operation permission matrix before implementing classifier adoption.
6. Revise reconciler and runtime-fact slices to anchor on current pure deciders,
   preserve partial/unknown fail-closed behavior, and define desired-running
   composition and circuit/progress/budget ownership.
7. Add a migration and coexistence plan that sequences the mutation-boundary work
   against the active worker-boundary migration and prevents split-brain field
   writers.
8. Add an operability contract mapping decisions, conflicts, and events to
   existing trace codes, API/CLI diagnostics, logs, doctor output, and tests.
9. Add a per-slice vocabulary map that defers unused `SessionMutation`,
   `RuntimeIntent`, `SessionEvent`, and broad fact aggregates until the slice
   that consumes them.
10. Add API/CLI typed-wire and performance gates to every affected backlog slice,
    including generated schema/client checks and hot-loop fact materialization
    budgets where applicable.
