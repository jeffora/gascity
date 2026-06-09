# Design Review Synthesis

## Overall Verdict: block

Attempt 16 is materially stronger than prior iterations and is directionally safe for a non-mutating Slice 0 preflight. It still blocks overall approval because the design has not made parity proof, atomic mutation fencing, source-complete migration sequencing, and worker-boundary coexistence mechanically enforceable. Behavior-moving slices should not be decomposed until the missing gates are materialized and tied to exact source rows, routes, keys, and proof commands.

## Consensus Strengths

- Multiple personas praised the explicit split between non-mutating Slice 0 evidence gathering and later behavior-moving slices.
- The read-only target-classification first adopter is the right direction if it preserves current resolver precedence and keeps materialization, repair, and mutating callers out of scope.
- The event model is sound in principle: session events are post-commit facts and safety-critical convergence must come from durable state scans rather than in-process event delivery.
- The session/reconciler/runtime ownership split is directionally correct: session may own lifecycle projection and eligibility over immutable facts, while work demand, pool sizing, provider health, progress policy, restart budgets, and runtime observations remain outside `internal/session`.
- The new migration/coexistence section, vocabulary lifecycle, boundary-matrix schema, diagnostics/budget concepts, and micro-slicing loop are useful controls if they become executable artifacts rather than prose.

## Critical Findings

### [Blocker] Parity and evidence gates are not mechanically enforceable

**Sources:** Natasha Volkov / 02-behavior-parity-guardian; Elena Marchetti / 01-mutation-boundary-auditor; Ingrid Holm / 10-operability-performance-diagnostics; Liam Okonkwo / 07-reconciler-runtime-fact-reviewer

**Issue:** Slice 0 proof can still go green without proving the behavior it claims to protect. The reviewers identified stale or missing `SESSION-*` evidence, especially `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`, plus proof commands that are not yet required to fail on absent validators, zero matched tests, skipped tests, build-tagged-out tests, or stale selectors.

**Required change:** Make the Slice 0 proof command self-validating. Every active `SESSION-*` row must resolve to a current compiled and executed selector or carry an owner-approved amendment/retirement, and the scenario allocation artifact must map each row to exact slices, surfaces, routes, current oracle, characterization proof, and proof command.

### [Blocker] Atomic command semantics lack an enforceable store and fence contract

**Sources:** Takeshi Yamamoto / 03-decider-atomicity-enforcer; Ravi Krishnamurthy / 08-migration-coexistence-strategist; Elena Marchetti / 01-mutation-boundary-auditor; Liam Okonkwo / 07-reconciler-runtime-fact-reviewer

**Issue:** The design names pre-commit revalidation, token/revision/value preconditions, stale-fact handling, and cross-process fencing, but it does not bind mutating operations to a real store capability or one chosen fence strategy. The current code has process-local mutation locks and blind or sequential metadata updates, which are not sufficient for CLI, API, and controller processes racing across OS process boundaries.

**Required change:** Materialize a store-capability matrix before any mutating slice moves. Then choose, per operation, exactly one strategy: store-native conditional write, durable value/token/revision precondition with immediate reread, or explicitly repair-converged blind write with eligibility limits, supersession keys, post-write verification, and failure-injection tests.

### [Blocker] Migration coexistence and worker-boundary sequencing are not source-complete

**Sources:** Ravi Krishnamurthy / 08-migration-coexistence-strategist; Sarah Chen / 06-api-cli-worker-boundary-reviewer; Elena Marchetti / 01-mutation-boundary-auditor; Amara Diallo / 05-target-identity-classifier

**Issue:** The migration trigger is too narrow and the current inventory is source-inaccurate. Reviewers called out unresolved surfaces including `resolveSessionTargetIDWithContext` sharing read-only lookup with `materialize:true` named-session creation, Huma close constructing a session manager and calling `CloseDetailed`, multiple raw close writers, wake/nudge/sleep/wait/drain paths, create rollback, lifecycle parallel start, generic bead update bridges, and exported patch-map APIs.

**Required change:** Replace file-name based triggers with inventory-driven migration rows from `SESSION_BOUNDARY_SYMBOLS.yaml`, `API_CLI_ROUTE_INVENTORY.yaml`, and `WORKER_BOUNDARY_EXCEPTIONS.yaml`. Add exact expiring exception rows or `worker.Handle` routing for every API/CLI mutating lifecycle path, broaden close alignment to all production close writers, and require key-level before/during/after ownership plus rollback proof for each migrated field or dynamic key source.

### [Blocker] Target resolution, materialization, and repair semantics still collide

**Sources:** Ravi Krishnamurthy / 08-migration-coexistence-strategist; Amara Diallo / 05-target-identity-classifier; Natasha Volkov / 02-behavior-parity-guardian; Ingrid Holm / 10-operability-performance-diagnostics

**Issue:** The read-only target classifier is a good first adopter only if it does not fork precedence from materializing and mutating resolver paths. Current helpers also perform `RepairEmptyType` style write-on-touch behavior and then continue selection from in-memory repaired beads, so returning `repair-needed` without a route-specific lifecycle can regress successful lookups or reintroduce read-side writes.

**Required change:** Add a shared-resolver sequencing rule. The design must either keep one source of target precedence for read-only and materializing modes until materialization migrates, or prove anti-drift with match-vector/result-kind parity tests. Define `repair-needed` per endpoint: who schedules the audited repair, whether reads retry or return a typed degraded result, how current successful lookups are preserved or owner-amended, and what fence protects repair.

### [Major] Target-classification output risks becoming a flat optional facade

**Sources:** Kwame Asante / 09-yagni-contract-scope-reviewer; Amara Diallo / 05-target-identity-classifier; Sarah Chen / 06-api-cli-worker-boundary-reviewer

**Issue:** The target-classification schema currently reads like a broad flat optional envelope while the design itself forbids flat optional shared types. Several fields are not consumed by the read-only API query first adopter, including lifecycle/label state, materialization flags, stale or partial diagnostics, and later-surface details.

**Required change:** Classify non-first-adopter fields as provisional and forbid them from active Go structs, OpenAPI/generated-client types, event payloads, and cross-slice contracts. Production output should use tagged or per-kind result structures, or the current table must be explicitly labeled as a provisional field census rather than an implementation contract.

### [Major] Event, recovery, and public-wire contracts need closed-world proof

**Sources:** Amara Osei / 04-event-delivery-contract-reviewer; Ingrid Holm / 10-operability-performance-diagnostics; Takeshi Yamamoto / 03-decider-atomicity-enforcer

**Issue:** The event model is directionally correct, but `NoPayload` session events have an identity and generation blind spot. Critical subscribers need durable idempotency keys, stale-event suppression, and scan recovery. Public SSE/OpenAPI/generated-client impacts are also not optional when payload shapes change.

**Required change:** Inventory every current `session.*` and session request/result event. For each, record committed fact, emission owner, payload type, canonical identity fields, public visibility, criticality, durable scan owner, idempotency key, duplicate/stale behavior, and proof selector. Require typed payloads when envelope fields are insufficient, especially for lifecycle, work-release, identity-retirement, and identity-reuse events.

### [Major] Runtime/reconciler boundaries, diagnostics, and budgets are not materialized

**Sources:** Liam Okonkwo / 07-reconciler-runtime-fact-reviewer; Ingrid Holm / 10-operability-performance-diagnostics; Sarah Chen / 06-api-cli-worker-boundary-reviewer

**Issue:** The design names the right boundaries but does not yet provide the executable rows and guards. Missing pieces include `BOUNDARY_MATRIX.yaml`, a pure-decider import/call guard, runtime-missing anti-flap rules, operation-specific `RuntimeIntent` schemas, `DIAGNOSTICS_MANIFEST.yaml`, typed trace/doctor/API/CLI diagnostic mappings, large-city performance evidence, and production `bdstore` subprocess accounting.

**Required change:** Materialize these artifacts before behavior-moving slices. Boundary rows must state owners, allowed inputs/outputs, forbidden imports/facts/policies, freshness and partial-state behavior, destructive-action safety, recovery owner, proof selector, and negative fixtures. Budgets must count store calls, subprocesses, scanned rows, runtime probes, event fan-out, durable scan size, fixture size, owner, threshold, and allowed delta.

### [Minor] Review artifact attempt paths are inconsistent

**Sources:** This synthesis observed the source bead metadata; Takeshi Yamamoto also flagged attempt-directory mismatch in the raw review signal.

**Issue:** The logical workflow attempt is 16, but the ten persona synthesis source beads stamped `design_review.output_path` under `.gc/design-reviews/ga-unpr2y/attempt-1/persona-syntheses/`. This synthesis used those stamped source outputs because they are the complete persona-level artifacts from the closed fanout, then wrote the global report to the required attempt 16 path.

**Required change:** Fix the persona-synthesis output path calculation so logical `gc.attempt` and artifact directories agree. Future global synthesis steps should not have to reconcile source-bead output paths against the current attempt directory manually.

## Disagreements

- Verdict severity differs across lanes. Three personas block; the remaining seven approve with risks, usually limited to non-mutating Slice 0. Under worst-verdict-wins, the global verdict is block.
- Several reviewers would allow Slice 0 to proceed as an inventory/preflight even while blocking behavior-moving decomposition. The synthesis agrees: non-mutating evidence gathering is acceptable, but mutation or routing behavior must wait for executable gates.
- Reviewers disagree on whether rollback safety requires a new state-version/schema-epoch tag. The synthesis position is that an epoch is not the only possible answer, but every key family needs equivalent version/tolerance proof; without that proof, an epoch is the cleanest default.
- Store-level mutation exceptions are disputed. Some reviewers want them eliminated; others allow exact, expiring, root-approved exceptions. The synthesis accepts exceptions only as temporary, precise rows with owner, expiry, allowed helper, route/symbol, parity proof, and retirement condition.
- Path-alias lookup mitigation differs. Indexing or removal is preferred; keeping the current all-session scan is acceptable only with a hard maximum, production `bdstore` subprocess budget, deterministic tiebreaker behavior, and large-city proof.
- Universal Slice 0 scope is disputed. The synthesis keeps source inventory, scenario indexing, baseline evidence, and vocabulary checkpoint mechanics universal, but moves detailed command appliers, mutating boundary contracts, worker-boundary exception details, and mutation diagnostics to the slices that consume them unless they are strictly inventory.

## Missing Evidence

- Materialized `SESSION_BOUNDARY_SYMBOLS.yaml`, `API_CLI_ROUTE_INVENTORY.yaml`, `WORKER_BOUNDARY_EXCEPTIONS.yaml`, `SCENARIO_PARITY.yaml`, `COMMAND_APPLIERS.yaml`, `BOUNDARY_MATRIX.yaml`, `TARGET_CLASSIFICATION_CONTRACT.yaml`, `DIAGNOSTICS_MANIFEST.yaml`, and store-capability inventory.
- Self-validating proof commands that fail on absent validators, skipped validators, build-tagged-out validators, zero matched tests, stale selectors, and missing negative fixtures.
- Current executable proof or owner-approved amendment/retirement for every active stale `SESSION-*` evidence row.
- Exact API/Huma/legacy mux/Cobra/generated-client/dashboard/SSE route and operation inventory for the first target-classification adopter and for every later API/CLI-touching slice.
- Exact worker-boundary exception rows for current API manager construction, Huma close, API materializing named-session create, wake/drain paths, CLI local fallback, repair/doctor/migration paths, and temporary direct lifecycle calls.
- Store-level atomicity and cross-process visibility evidence for all persistence surfaces used by session commands.
- Per-operation command rows for wake/start, close, drain, stop/interrupt, runtime start, runtime-missing cleanup, identity retirement, token backfills, and repair/backfill.
- Pure-decider file/package set and CI guard proving no store/runtime/config/event/work/mail/extmsg/subprocess/provider-policy imports or ambient clock reads in mutation-feeding deciders.
- Route-specific `repair-needed` behavior and repair scheduling/fencing for empty-type sessions.
- Exhaustive `session.*` event inventory, subscriber audit, typed payload adequacy proof, idempotency keys, and public-wire sync proof.
- Diagnostic and trace mappings for accepted, rejected, blocked, deferred, stale-fact, partial-fact, repair-needed, event-failed/skipped/duplicated, scan-recovered, and scan-failed outcomes.
- Numeric large-city budgets and current baselines for API target resolution, path-alias lookup, reconciler fact compilation, event fan-out, and durable recovery scans.
- Attempt artifact path consistency between source bead `gc.attempt` metadata and `design_review.output_path`.

## Recommended Changes

1. Limit the next approval to non-mutating Slice 0 preflight only; explicitly block behavior-moving decomposition until the gates below pass.
2. Make parity proof self-validating and repair or owner-retire every stale active `SESSION-*` evidence row.
3. Materialize source-complete boundary, route, exception, scenario, diagnostics, event, and store-capability inventories with validators.
4. Decide and test the store/fence strategy per mutating operation before moving any metadata writer.
5. Sequence target classification with the worker-boundary migration so read-only lookup, materializing create, repair, and mutating resolver users cannot fork precedence.
6. Define `repair-needed` behavior per endpoint, including repair owner, non-blocking behavior, API/CLI result or retry semantics, trace evidence, persistence-error propagation, and concurrency fence.
7. Convert target-classification outputs to tagged/per-kind production types and keep non-first-adopter vocabulary provisional.
8. Extend worker-boundary enforcement to `internal/api` and require exact expiring exception rows or `worker.Handle` routing for every production lifecycle mutation.
9. Add exact key-level migration rows with before/during/after owners, old writer, new command owner, fence, validation parity, rollback behavior, and retirement condition.
10. Materialize event payload, durable recovery, diagnostics, trace, and performance-budget gates before any slice that changes runtime, reconciler, API/CLI, or event behavior.
11. Fix persona-synthesis attempt output paths so attempt 16 source artifacts are written under the attempt 16 directory.
