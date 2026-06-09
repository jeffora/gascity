# Design Review Synthesis

## Overall Verdict: block

Worst-verdict-wins produces `block`: all ten persona syntheses returned `block`. The design has useful direction in Slice 0 preflight, durable-scan recovery, surface-specific target resolution, runtime observation taxonomy, and vocabulary restraint, but the current artifact set is not executable enough to protect session lifecycle ownership, parity, recovery, migration safety, or operator visibility.

## Consensus Strengths
- Multiple personas praised the non-mutating Slice 0 idea as the right first step if it becomes a real inventory, guard, parity, vocabulary, and proof-command deliverable.
- Reviewers agreed that durable scans and persisted facts should be the correctness backstop, with events serving as latency and diagnostic hints rather than the only recovery mechanism.
- The runtime observation completeness taxonomy and seven-state eligibility-mask direction were seen as useful foundations for separating session decisions from reconciler policy.
- The per-surface target-classification direction is viable if the raw classifier stays policy-free and adapters preserve CLI, API, mail, extmsg, assignee, and pool behavior explicitly.
- The vocabulary and YAGNI posture is improving: several reviewers endorsed deferring future slice vocabulary, avoiding broad facades, and requiring first-caller evidence before new public shapes land.

## Critical Findings

### [Blocker] Slice 0 proof is still prose instead of executable protection
**Sources:** Elena Marchetti, Natasha Volkov, Takeshi Yamamoto, Sarah Chen, Liam Okonkwo, Ravi Krishnamurthy, Kwame Asante, Ingrid Holm
**Issue:** The required preflight artifacts are absent, incomplete, or not fail-closed: source-complete writer inventory, scenario parity, vocabulary checkpoints, static guards, shrink-only allowlists, baseline pinning, proof commands, diagnostics mapping, and performance baselines. Broad W-rows, prose key buckets, checkout-relative proof, and zero-match tests would let later implementation beads move behavior without real protection.
**Required change:** Make Slice 0 the only approved first deliverable. It must generate or commit machine-readable inventory, parity, vocabulary, guard, allowlist, diagnostics, baseline, and proof artifacts that fail on missing rows, stale paths, zero matched tests, allowlist growth, expired exceptions, and unowned session metadata writes.

### [Blocker] Mutation ownership and static guard coverage miss real bypass paths
**Sources:** Elena Marchetti, Sarah Chen, Ravi Krishnamurthy, Ingrid Holm
**Issue:** The proposed guard does not yet cover production bypass classes such as generic store handles, Huma/generic bead routes, manager method calls, package-level mutators, patch-map extension, subprocess `bd` mutations, repair helpers, and files outside the stated scan floor. Blanket doctor, repair, and migration exclusions are self-labeling escape hatches.
**Required change:** Define a source-complete scan over production `cmd/` and `internal/` paths that import or receive bead stores or session mutators. Split broad rows into exact path/symbol/key rows, remove class-level repair/doctor/migration exclusions, and prove guard behavior with fixtures for dynamic keys, wrapper receivers, interface receivers, patch-map application, manager calls, and generic bridges.

### [Blocker] Command atomicity, stale-writer safety, and one-writer migration are unresolved
**Sources:** Takeshi Yamamoto, Ravi Krishnamurthy, Elena Marchetti, Sarah Chen
**Issue:** Runtime-start, close, wake, drain, identity retirement, token backfill, wake/hold/drain metadata, and API mutation routes still lack physical commit contracts. The design assumes safety not proven by the current store surface, leaves blind writes and partial batch states under-defined, and simultaneously implies gradual migration and no parallel writers.
**Required change:** Materialize command-applier rows with preconditions, validation points, write primitives, fence markers, stale-success handling, partial-state repair, runtime side-effect ordering, and race tests. Either add/prove conditional writes or classify blind writes as detect-and-converge. For every key family and API/CLI surface, prove legacy callers are removed, delegated through the same applier, or read-only before bake.

### [Blocker] Close, work release, and event recovery are not durably recoverable enough
**Sources:** Amara Osei, Takeshi Yamamoto, Ingrid Holm
**Issue:** Close/retire can clear identity or commit terminal state before work release and binding cleanup are durably discoverable after restart. The pending-release concept is sound, but the design does not name the closed-bead query, scanner/helper, cadence, boundedness, release-key snapshot, completion fact, stale-successor suppression, or duplicate-scan semantics.
**Required change:** Add a durable close/work-release recovery contract. Preserve assignment keys or write a pending-release/closing fact before identity clearing, name the recovery scanner and exact query shape, define idempotency and supersession keys, and add crash-after-commit, skipped-event, duplicate-scan, partial-query, restart, and stale-successor tests.

### [Blocker] Target resolution and projection parity remain surface-incomplete
**Sources:** Amara Diallo, Sarah Chen, Natasha Volkov
**Issue:** The classifier boundary conflicts with current repair, materialization, closed-ID, historical-alias, mail, assignee, API, and CLI behavior. Missing or ambiguous cases include CLI qualified-alias basename fallback, configured-name precedence, rejected-by-config post-filtering, mail send versus query, direct closed ID behavior, assignee materialization retry, API/Huma route differences, and output/wire compatibility.
**Required change:** Add per-surface compatibility chains and fixture matrices before Slice 1 adoption. The raw classifier should collect physical identity/type facts only; adapters must own materialization, repair, rejection policy, live-only enforcement, output shapes, status codes, `RequestID`, generated-client effects, stdout/stderr/JSON, exit codes, and no-leak rules.

### [Blocker] Session/reconciler and runtime-fact boundaries are not decomposition-ready
**Sources:** Liam Okonkwo, Takeshi Yamamoto, Ravi Krishnamurthy
**Issue:** `AwakeInput` still mixes session eligibility with controller demand, min-active policy, work aggregation, pool/named-session demand, direct wake, runtime observation, provider health, progress, and scale behavior. Missing health/progress/scale proof files and `ProjectLifecycle` wall-clock fallback contradict the pure-decider story.
**Required change:** Add field-level disposition rows and parity fixtures for eligibility, min-active, idle sleep, forced versus conditional runnable, wait-only wake, attached/pending interaction, pinned/named-always, runtime liveness, work counts, progress, provider health, and scale-from-zero. Define a guarded pure-decider file set, remove or isolate direct clock reads, and keep health/progress/scale slices blocked until proof files are restored or replaced.

### [Blocker] Repair, diagnostics, and performance are not operable
**Sources:** Ingrid Holm, Amara Diallo, Sarah Chen, Elena Marchetti
**Issue:** `RepairEmptyType` remains a read-path write that can swallow persistence errors and mutate in-memory state, yet no real repair slice owns it. Trace, doctor, session-inspect, and performance requirements are not machine-checkable, and proposed outcome vocabulary can overclaim missed-event recovery without durable delivery evidence.
**Required change:** Assign `RepairEmptyType` to a sequenced repair slice or audited helper, propagate write errors, avoid in-memory repair on failed persistence, and make repair writes guard-visible. Add a diagnostics manifest, trace outcome compatibility table, doctor/session-inspect registry or golden file, renderer/filter tests, query-count budgets, backend index proof or measured budgets, subscriber backpressure tests, and large-city baselines.

### [Blocker] Vocabulary and scope controls still permit premature abstractions
**Sources:** Kwame Asante, Ravi Krishnamurthy
**Issue:** Active design material still mixes next-slice vocabulary with future-slice names such as `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent`. `TestVocabularyCheckpoints` could pass vacuously, and Slice 1 target classification could become a universal resolver or flat optional envelope before a first production adapter proves each field.
**Required change:** Split active vocabulary from design-only future bounds. Seed Slice 0 checkpoints with existing-contract rows, require first delegated production caller evidence for new rows, reject undeclared or unused fields, keep future event vocabulary provisional, and constrain the first Slice 1 target-classification bead to one production surface and only the fields it needs.

### [Major] Migration and rollback contracts lack data-direction proof
**Sources:** Ravi Krishnamurthy, Takeshi Yamamoto, Sarah Chen
**Issue:** Rollback language restores routes more than it proves restored legacy code can converge over new-path metadata such as tokens, phase markers, close facts, wake normalization, `alias_history`, and runtime identity fields. API destination choices also remain split between `worker.Handle`, session command factory, direct manager, and legacy/Huma routes.
**Required change:** For each slice and adopting surface, require converted caller list, allowed legacy rows, retired rows, guard changes, rollback route, data-direction convergence proof, one-writer proof, and exact API/CLI route inventory. Prefer `worker.Handle` unless a narrow exception is explicitly documented with owner, expiry, and guard enforcement.

### [Minor] Attempt artifact paths are inconsistent
**Sources:** Workflow artifacts observed during synthesis
**Issue:** The ten persona synthesis beads for attempt 8 stamped fresh outputs under `.gc/design-reviews/ga-unpr2y/attempt-1/persona-syntheses/` while the current attempt directory `.gc/design-reviews/ga-unpr2y/attempt-8/persona-syntheses/` was empty. This synthesis mirrored those fresh child-stamped outputs into attempt 8 before reading them.
**Required change:** Fix child synthesis attempt selection or output-path stamping so future global synthesis steps can read the current attempt directory directly without artifact recovery.

## Disagreements
- Several raw model reviews returned `approve-with-risks`, but every persona synthesis chose `block` after weighing the cross-model evidence. The global verdict follows worst-verdict-wins and the fact that all persona-level verdicts are blockers.
- Reviewers disagreed on whether missing Slice 0 artifacts block design approval or merely limit implementation to Slice 0. The synthesis assessment is that Slice 0 may proceed, but no mutation-owning, behavior-moving, or decomposition work should proceed until Slice 0 artifacts exist and pass.
- Baseline evidence is disputed because some findings depend on active checkout state while others reference upstream or historical commits. The design must pin a clean baseline ref, re-audit there, and treat commit hashes as historical unless live proof and ancestry checks also pass.
- Close/work-release reviewers proposed different recovery handles: pending-release flags, transitional closing state, release-key snapshots, tokened assignments, or live-successor guards. Any mechanism is acceptable only if it is durable after restart and tested against stale successor reuse.
- API routing reviewers differ on whether the end route should be strictly `worker.Handle` or a narrower session command factory. The design must choose per surface and record exceptions in the worker-boundary ledger rather than letting both migrations coexist implicitly.
- Event reviewers accept durable-scan-first recovery direction, but disagree on terminology such as `missed-event-recovered`. The synthesis keeps durable convergence but requires truthful outcome names unless durable delivery tracking exists.
- Target vocabulary reviewers differ on whether the target candidate shape must be split into per-kind structs. The required property is mechanical prevention of a universal optional envelope, whether by tagged/private types or checkpoint enforcement.

## Missing Evidence
- Version-controlled Slice 0 artifacts: source-complete writer inventory, `SCENARIO_PARITY.yaml`, `VOCABULARY_CHECKPOINTS.yaml`, static guard implementation and fixtures, shrink-only allowlist, diagnostics manifest, baseline proof, and CI/pre-commit commands.
- Pinned clean baseline ref, ancestry checks for historical commits, complete 45-row requirements parity, current oracle commands, and owner-approval evidence for intentional requirements changes.
- Store primitive and lifecycle command matrices for runtime-start, close, wake, drain, identity retirement, token backfill, batch writes, provider side effects, stale writers, and partial-state repair.
- Closed-session pending-release scanner contract, durable release-key schema, completion/clearing semantics, idempotency keys, supersession keys, and stale-successor tests.
- Per-surface target/API/CLI/mail/extmsg/assignee matrices with precedence, closed-ID policy, historical-alias policy, materialization/repair ownership, output/wire oracles, and generated-client effects.
- Pure-decider guard, `ProjectLifecycle` zero-`Now` policy, `AwakeInput` disposition rows, min-active ownership, runtime observation facts, restored provider-health/progress/scale tests, and W-013 freshness/supersession rules.
- Repair ownership and guard coverage for `RepairEmptyType`, trace/doctor/session-inspect check registry, trace outcome compatibility table, session-scoped operator surface, query budgets, backend index proof, subscriber backpressure tests, and large-city baselines.
- Vocabulary checkpoint fixture list, first production caller for Slice 1 target classification, package visibility rules, field-exclusion rules, and mutation-slice overlap policy.

## Recommended Changes
1. Approve only a non-mutating Slice 0 preflight next: inventory, parity, vocabulary, guard, allowlist, diagnostics, baseline, proof commands, and workflow gates must physically exist and fail correctly.
2. Expand the mutation guard and inventory to cover all production writer and mutator bypasses, including generic stores, API/Huma bridges, manager methods, package mutators, patch-map extension, repair helpers, and subprocess mutation routes.
3. Rewrite lifecycle command contracts to match actual store primitives, then add stale-writer, partial-batch, runtime-start, close, wake, drain, token-backfill, provider-failure, and event-miss tests.
4. Define durable close/work-release recovery with preserved release keys or pending facts, scanner ownership, bounded queries, idempotency, supersession, completion, trace/doctor output, and stale-successor proof.
5. Make target resolution and API/CLI projection surface-specific, with compatibility chains, collision matrices, adapter-owned side effects, exact wire/output oracles, and no caller authority over diagnostic candidates.
6. Split session eligibility from reconciler policy through field-level `AwakeInput` disposition, pure-decider guards, runtime-observation facts, and restored health/progress/scale proof before those slices move.
7. Give repair, diagnostics, and performance concrete owners and executable artifacts: repair slice/helper, diagnostics manifest, trace constants, doctor/session-inspect registry, query-count fixture, backend proof, subscriber budgets, and large-city baselines.
8. Constrain vocabulary and YAGNI scope by separating active next-slice vocabulary from design-only future names, requiring first-caller evidence, and preventing broad optional envelopes or universal resolver behavior.
9. Resolve migration overlap explicitly: choose API/CLI end routes, one-writer ownership per key family, bake/revert criteria, data-direction rollback proof, and worker-boundary ledger updates before generating behavior-moving beads.
10. Fix the review workflow artifact path bug so child persona syntheses write to the current attempt directory and stamp `design_review.output_path` consistently.
