# Design Review Synthesis

## Overall Verdict: block

Nine persona syntheses returned `block`, and one returned `approve-with-risks`, so the global verdict is `block` by worst-verdict-wins. The design direction is broadly sound, but the current document is not yet safe to decompose because its source inventory, behavior proof, command atomicity, surface compatibility, recovery, and diagnostics gates are still too ambiguous to constrain implementation beads.

## Consensus Strengths
- Multiple personas praised the move toward session-owned command boundaries, one-writer ownership, and explicit migration gates instead of scattered metadata mutation.
- Reviewers agreed that read-only target classification is the right core direction when kept separate from per-surface policy, repair, and materialization.
- The durable-fact recovery model is directionally correct: events should be factual accelerators or diagnostics, while scans provide crash recovery for critical reactions.
- The design correctly identifies the session/reconciler boundary as the key seam and tries to keep provider health, progress, scale, and restart policy out of `internal/session`.
- Vocabulary checkpoints, provisional bounds, owner matrices, and parity gates are useful controls, even though they need executable enforcement before task generation.

## Critical Findings

### [Blocker] Mutation writer inventory and guards are not source-complete
**Sources:** Elena Marchetti, Takeshi Yamamoto, Ravi Krishnamurthy, Ingrid Holm, Sarah Chen

**Issue:** The design still depends on a writer inventory that reviewers found incomplete. Omitted or under-specified writers include `internal/api/handler_sessions.go`, `internal/api/handler_beads.go`, `cmd/gc/session_resolve.go`, `cmd/gc/session_reconcile.go`, `cmd/gc/session_name_lookup.go`, `cmd/gc/adoption_barrier.go`, `cmd/gc/cmd_prime.go`, `cmd/gc/session_lifecycle_parallel.go`, `cmd/gc/session_beads.go`, `internal/session/chat.go`, generic bead update/close/metadata paths, `RepairEmptyType`, patch-map extension, package-level session mutators, and runtime identity setter/clearer paths. The proposed mutation-boundary guard is still a promise rather than a named checked-in artifact with fixtures, scan roots, allowlist schema, CI command, and shrink-only enforcement.

**Required change:** Add a non-mutating preflight/Slice 0 that generates or verifies the writer/clearer/repair inventory from the checkout and installs an additive static guard before any mutation-owning slice starts. The guard must cover direct owned-key writes, dynamic metadata batches, raw bead mutations, manager and package-level mutators, exported patch builders, patch-map aliasing/extension, generic bridge writes, `Create`/`Update`/`Close`, and repair helpers, with expiring allowlist rows keyed to concrete inventory IDs.

### [Blocker] Behavior parity and citation freshness are not enforceable
**Sources:** Natasha Volkov, Sarah Chen, Liam Okonkwo

**Issue:** The parity artifacts disagree with each other: the full baseline claims all 45 `REQUIREMENTS.md` rows, while the Scenario Traceability Matrix and backlog proof bullets omit active rows and carry weaker obligations. Several cited current-proof rows appear stale or phantom, including `SESSION-RECON-006`, `SESSION-RECON-007`, and `SESSION-WORK-003`; all reviewed commit-hash citations were reported as non-ancestors of HEAD. API and CLI compatibility proof is also not decomposable enough to preserve status codes, error bodies, request IDs, async behavior, stdout/stderr, JSON shape, and exit codes.

**Required change:** Reconcile the scenario matrix, full parity gate, and backlog proof bullets into one canonical bead-generation source that accounts for all requirement rows. Add a citation-freshness gate with a named path and command that fails on missing paths, commit-only current proof, non-ancestor commits used as current evidence, and assertions that no longer cover the cited behavior. Require operation-level API/CLI golden or typed-response tests for every migrated surface.

### [Blocker] Session command atomicity and failure ordering are under-specified
**Sources:** Takeshi Yamamoto, Amara Osei, Ravi Krishnamurthy

**Issue:** The design still relies on "re-read before write" in places even though the store surface has no general CAS/if-match primitive and multi-key metadata batches are not proven atomic across active stores. Close and runtime-start remain especially risky: provider stop can succeed while durable close fails, provider start can succeed while metadata or event handling fails, `CloseDetailed` and `closeBead` have divergent behavior, and unfenced `instance_token` or `session_key` backfills undermine single-writer identity claims.

**Required change:** Define a command-applier contract for every command cluster with snapshot preconditions, token/revision or phase markers, allowed write primitives, conflict reasons, stale-success handling, partial-state matrices, repair or compensation authority, event ordering, and race tests. Fence close and runtime-start durably, either with recoverable intent markers before provider side effects or with proven compensation that cannot restart stopped sessions, orphan runtimes, or strand work.

### [Blocker] Target classification, adapter policy, repair, and materialization remain conflated
**Sources:** Amara Diallo, Sarah Chen, Kwame Asante, Elena Marchetti, Ingrid Holm

**Issue:** The raw classifier is supposed to be read-only, but current resolver paths perform repair, materialization, open-state checks, named-session policy, and recipient expansion. Reviewers found missing or vague compatibility rows for CLI resolution, legacy and Huma API paths, mail send/query/count/inbox, extmsg, assignee normalization, attach/observe/log/transcript, nudge, resume/dispatch alias-history routing, circuit breaker, doctor, and materialization. Candidate kind ordering and policy-shaped results such as `forbidden-kind`, `requires-materialization`, and `closed-not-allowed` can also lead callers to reinterpret or re-rank classifier output.

**Required change:** Split raw target fact collection from per-operation policy, repair, materialization, and recipient-set behavior. Add per-surface transition tables and fixtures that compare old and new behavior for selected target, recipient set, side effects, error class/message, JSON shape, materialization, closed handling, historical aliases, and ambiguity. Define `selected` as the only selection authority, make diagnostic candidates non-authoritative, and decide whether policy is an explicit classifier input or adapter-owned post-filtering.

### [Blocker] Reconciler/session ownership is not narrow enough to prevent policy leakage
**Sources:** Liam Okonkwo, Takeshi Yamamoto, Ingrid Holm, Natasha Volkov

**Issue:** The `ComputeAwakeSet`/`AwakeInput` seam still mixes session eligibility with controller demand, work aggregation, scale counts, pool capacity, runtime observations, provider health, progress, restart budgets, circuit state, alert dedupe, trace behavior, and idle-sleep policy. The proposed purity guard is also not mechanically scoped because `internal/session` already contains store and manager code. Runtime observations lack a complete stale/unknown/partial/provider-error/unobserved/alive model for destructive branches.

**Required change:** Define a narrow session-owned eligibility contract over immutable lifecycle, identity, and session-domain facts only. Keep demand assembly, scale, health, progress, circuit, retry budget, alert dedupe, trace rendering, and idle-sleep policy in the reconciler/controller. Introduce a guardable pure-decider file set or subpackage, and add runtime observation completeness plus fail-closed rules for close, drain, rollback, release, cleanup, and restart branches.

### [Blocker] Event recovery, diagnostics, and performance gates are not executable
**Sources:** Amara Osei, Ingrid Holm, Sarah Chen, Takeshi Yamamoto

**Issue:** The close/work-release contract still contains contradictory decomposition-facing language about synchronous release versus durable scan recovery. Crash-after-commit and event-miss tests are not attached to affected slices, event diagnostics do not match the current best-effort recorder API, and thin events lack a typed internal fact/idempotency contract. Operability and performance promises are also not concrete enough: trace codes, doctor checks, renderers, query-count tests, large-city benchmarks, subscriber budgets, and reconciler hot-loop write budgets are not named with commands or fixtures.

**Required change:** Make durable-fact scans the mandatory backstop for every critical retryable reaction, with synchronous cascades only as latency optimizations. Define internal post-commit fact/idempotency keys, event-miss/crash-after-commit proof commands, close/work-release scan owner and cadence, live-successor guards, and failure behavior for partial queries. Add concrete trace/doctor/API/CLI diagnostic artifacts and executable performance gates for lookup, scans, fact materialization, event fan-out, subscriber backpressure, extmsg/mail expansion, and hot reconciler writes.

### [Major] Migration coexistence and rollback remain too loose
**Sources:** Ravi Krishnamurthy, Elena Marchetti, Sarah Chen

**Issue:** Worker-boundary compliance and mutation-boundary compliance are independent, but the migration plan can still imply that routing through `worker.Handle` proves metadata safety. Shared files such as `internal/api/session_resolution.go` span classify-only resolution, materialization, and identifier retirement across slices, and broad rows such as W-028 do not give caller-level bake behavior or rollback units.

**Required change:** Require every implementation bead to carry `session_design.converts`, `session_design.allows_during_bake`, `session_design.retires`, guard-row changes, rollback file/caller sets, and proof commands. Add a combined call-site/key owner matrix and mechanical one-writer proof so an owned key cannot retain multiple production writers after its guard lands.

### [Major] Scope and vocabulary controls are still advisory
**Sources:** Kwame Asante, Liam Okonkwo, Ingrid Holm

**Issue:** The target taxonomy, policy dimensions, command conflicts, diagnostics, event facts, and runtime intents can still be read as day-one shared vocabulary. Slice 1 risks becoming a universal target-resolution facade, and future-slice names such as `SessionFactEvent`, `RuntimeStartIntent`, and `SessionCommandConflict` are not tied tightly enough to first callers or rule-of-two evidence.

**Required change:** Make vocabulary checkpoint verification observable and failing. Each shared type or field must name its first delegated caller, exact demanded fields, rule-of-two status, non-goals, provisional bounds, and owning slice. Treat broad taxonomies as closed upper bounds that shrink unless an adopting surface proves the field is needed.

### [Minor] Attempt-scoped synthesis artifacts are still written to stale paths
**Sources:** Workflow artifact metadata, synthesis assessment

**Issue:** The attempt-6 `persona-syntheses/` directory is empty, while the ten closed persona-synthesis beads for `gc.attempt=6` stamped fresh outputs under `.gc/design-reviews/ga-unpr2y/attempt-1/persona-syntheses/`. This synthesis used those bead-declared outputs because they are complete, current, and metadata-linked to attempt 6, but the documented `$ATTEMPT_DIR/persona-syntheses` path remains unreliable.

**Required change:** Fix the design-review workflow so persona synthesis beads write under the current attempt directory, or have the global synthesis step materialize a manifest or copy of the exact source files it consumed.

## Disagreements
- Verdict disagreements were local, not global. Event delivery, target identity, API/CLI boundary, reconciler boundary, and scope lanes had model splits, but the persona syntheses selected `block` where the ambiguity affects decomposability or safety.
- Kwame Asante's scope lane returned `approve-with-risks`. I agree that the broad direction is acceptable there, but those risks become required controls because vocabulary creep can distort the first implementation slice.
- Reviewers allowed multiple acceptable atomicity models: a real conditional primitive, tokened phase markers, single-writer serialization, or convergence tests with bounded repair. The design does not need to choose CAS specifically, but it must reject unconditional writes after stale pre-read validation.
- Event reviewers disagreed on whether public payload enrichment is needed now. The synthesis is that public wire changes may be deferred, but subscribers still need stable internal post-commit facts, idempotency keys, and durable scan recovery.
- Target reviewers disagreed on whether repair can be moved out of the classifier without preserving side effects. The classifier should stay read-only, but the design must explicitly preserve repair in audited adapters or remove it through owner-approved requirements changes.
- Some reviewers focused on concrete omitted files while others focused on abstract invariants. These are complementary: the design needs both a source-generated inventory and invariant-level acceptance gates.

## Missing Evidence
- Source-generated writer/clearer/repair inventory with W-IDs, file/function references, operation type, exact key set or dynamic-key class, target session-bead proof, owner slice, exception status, expiry, and retirement condition.
- Static guard implementation path, allowlist schema, fixture directory, CI command, and failing fixtures for direct store writes, dynamic metadata batches, package mutators, manager methods, patch builders, patch-map extension, generic bridge writes, create/update/close, and repair helpers.
- Canonical all-row scenario parity source, primary owner for multi-slice rows, citation-freshness command, and live proof or owner-approved amendments for phantom or stale requirement rows.
- Operation-level API/CLI compatibility matrix with status codes, error bodies, request IDs, async behavior, stdout/stderr, JSON fields, exit codes, OpenAPI/generated-client impact, and dashboard checks where applicable.
- Command-applier contracts, partial-state matrices, race tests, and compensation rules for runtime start, close, drain, wake, identity retirement, repair, and multi-key metadata writes.
- Per-surface resolver/adaptation fixtures for CLI, API, mail, extmsg, assignee normalization, nudge, attach/observe/log/transcript, resume/dispatch, circuit breaker, doctor, and materialization.
- Runtime observation completeness model and branch-level fail-open/fail-closed rules for destructive reconciler behavior.
- Durable close/work-release scan owner, cadence, identity key preservation, live-successor guard, duplicate scan handling, partial-query behavior, and CLI/API crash-after-close proof.
- Internal event/fact idempotency contract and a current-compatible answer for event-emission diagnostics with a best-effort recorder.
- Trace vocabulary constants, doctor check schemas, renderer tests, operator-facing command/API contracts, and centralized outcome/reason tables.
- Query-count tests, benchmarks, large-city fixtures, subscriber caps, defer/drop/retry semantics, and hot reconciler write budgets.
- Current-attempt manifest or artifact fix for persona synthesis source files.

## Recommended Changes
1. Add the non-mutating preflight/Slice 0 for source inventory generation, static mutation guard installation, repair inventory, and citation-freshness enforcement.
2. Reconcile all behavior-parity artifacts into one canonical bead-generation source and require row ownership plus live proof before task generation.
3. Define the command-applier model for atomicity, stale snapshots, partial writes, conflicts, repair, compensation, provider side effects, and event ordering.
4. Split raw target classification from adapter policy, repair, materialization, and recipient-set semantics; add per-surface compatibility fixtures before replacing resolvers.
5. Redraw the reconciler/session boundary around narrow session eligibility facts, controller-owned demand and health policy, pure-decider guards, and runtime observation completeness.
6. Resolve close/work-release and other critical event reactions around mandatory durable scans with crash-after-commit and event-miss tests.
7. Turn API/CLI/typed-wire compatibility into operation-level proof matrices, including async command behavior and classifier no-leak tests.
8. Make migration coexistence mechanical: per-key one-writer proof, per-file conversion order, bake allowances, guard-row changes, rollback units, and proof commands on every bead.
9. Make vocabulary checkpoints fail promotion without first-caller evidence, exact demanded fields, rule-of-two status, non-goals, and owning slice.
10. Specify trace, doctor, API/CLI, event, and performance deliverables as concrete constants, schemas, renderers, tests, benchmarks, and large-city budgets.
11. Repair the design-review workflow artifact-path drift so current-attempt persona syntheses are discoverable under `$ATTEMPT_DIR/persona-syntheses` or manifest-listed.
