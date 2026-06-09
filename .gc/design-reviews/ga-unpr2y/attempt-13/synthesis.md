# Design Review Synthesis

## Overall Verdict: block

All ten persona syntheses returned `block`, so worst-verdict-wins yields a global block. The design has converged on the right strategic direction: a non-mutating Slice 0 preflight with machine-checkable inventories, ledgers, and guards. It is not yet safe to decompose beyond that, and even Slice 0 needs one authoritative artifact/proof contract before it can close without encoding contradictions.

## Consensus Strengths
- Reviewers consistently endorsed non-mutating Slice 0 as the correct first scheduling boundary before any session-boundary behavior moves.
- The proposed inventory-and-ledger direction is sound: `SCENARIO_PARITY`, `COMMAND_APPLIERS`, `BOUNDARY_MATRIX`, `API_CLI_ROUTE_INVENTORY`, `DIAGNOSTICS_MANIFEST`, `VOCABULARY_CHECKPOINTS`, and writer/read inventories are the right shape of controls.
- The precompute-then-decide architecture, pure decider goal, command-applier fencing model, and durable-scan convergence model are the right primitives when made executable.
- The design correctly recognizes that repair/backfill, API/CLI route parity, worker-boundary coexistence, target identity classification, and diagnostics are not side notes; they are migration gates.

## Critical Findings

### [Blocker] Slice 0 Is Not A Single Executable Gate
**Sources:** Mutation Boundary Auditor, Behavior Parity Guardian, Decider Atomicity Enforcer, API/CLI Worker Boundary Reviewer, Reconciler Runtime Fact Reviewer, Migration Coexistence Strategist, YAGNI Contract Scope Reviewer, Operability Performance Diagnostics
**Issue:** Multiple Slice 0 artifact/proof lists still coexist, the canonical backlog can still be read as scheduling Target Classification first, and the named schemas, validators, guard fixtures, and freshness tests are not physically present. Several personas also found contradictions over whether Slice 0 needs live proof, planned proof, source-only proof, or explicit blocked fixture states.
**Required change:** Make Slice 0 the only first decomposable backlog item. Collapse every Slice 0 artifact/proof list into one authoritative contract, materialize the schemas and validators, and add a close/finalizer gate that blocks any non-Slice-0 session-boundary bead without referenced inventory IDs, one-writer evidence, rollback evidence, and proof-command metadata.

### [Blocker] Mutation And Repair Writer Inventory Is Incomplete
**Sources:** Mutation Boundary Auditor, Decider Atomicity Enforcer, Target Identity Classifier, Migration Coexistence Strategist, Operability Performance Diagnostics
**Issue:** The static guard scope conflicts between "all production cmd/internal Go" and narrower package roots; blanket repair/doctor/migration exclusions create self-labeling bypasses; patch-map detection is builder-name-based and incomplete; repair/backfill writers such as `RepairEmptyType`, runtime identity backfills, continuation backfills, and session key writers have no scheduled owner. Broad writer rows remain planning seeds rather than source-complete implementation rows.
**Required change:** Scan every non-test, non-generated production Go file under `cmd/` and `internal/`, or require exact expiring exclusions. Generate authoritative writer rows with target-bead proof, key family, receiver/helper chain, owner slice, expiry, guard rule, and retirement condition. Add a named repair slice or explicitly fold each repair/backfill conversion into its owning command slice.

### [Blocker] Runtime Start, Close, Drain, And Work Release Are Not Durably Safe
**Sources:** Decider Atomicity Enforcer, Event Delivery Contract Reviewer, Reconciler Runtime Fact Reviewer, Migration Coexistence Strategist
**Issue:** Runtime start still permits a no-token compatibility commit path; recovery is not keyed only on durable facts; close can still be interpreted as provider stop before durable close intent and release snapshot; drain generation and assigned-work events are not reconciled with public payloads; work-release scanner schemas and completion facts are incomplete.
**Required change:** Use durable intent before side effects for close and release. Remove normal no-token runtime-start commits; route them only through audited repair/adoption paths. Define command-applier rows and fault matrices for runtime start, close, drain, wake, rollback, repair, event emission, and partial metadata batches, including durable runtime discovery, stale-success handling, idempotency, event/wire obligations, and scanner completion facts.

### [Blocker] Pure Decider And Reconciler Boundaries Still Leak Policy
**Sources:** Decider Atomicity Enforcer, Reconciler Runtime Fact Reviewer, YAGNI Contract Scope Reviewer
**Issue:** Demand aggregation, pool capacity, scale/min-active policy, health/progress policy, runtime observations, and scheduling decisions are not cleanly separated from session-owned facts. Cause-sensitive holds are not precise enough to preserve `held_until`, `quarantined_until`, and `wait_hold` semantics. `time.Now()` zero-`Now` fallbacks remain in lifecycle projection and are not owned.
**Required change:** Make `BOUNDARY_MATRIX.yaml` normative and operation-specific. It must list immutable session inputs, adapter-owned facts, reconciler-owned policy, forbidden facts/imports, unknown/partial behavior, proof tests, and destructive-action semantics. Carry demand only as precomputed immutable facts or controller-composed results, define cause-sensitive masks, and add a pure-decider guard that rejects direct clocks and effect-capable inputs while allowing typed time values.

### [Blocker] Behavior Parity And Requirement Amendment Proofs Contradict Each Other
**Sources:** Behavior Parity Guardian, API/CLI Worker Boundary Reviewer, Reconciler Runtime Fact Reviewer, Migration Coexistence Strategist
**Issue:** Current proof is alternately described as source paths, issues, commits, runnable behavior tests, static assertions, or planned future proof. Scenario ownership tables disagree, some cited oracles are stale or absent, existing tests are not bound at assertion level, and requirements edits lack an owner-approved amendment artifact.
**Required change:** Make `SCENARIO_PARITY.yaml` the single source consumed by bead generation. Cover all `SESSION-*` rows with status, owner slice, exact selector or static assertion, current-oracle status, touched surfaces, baseline ref, compatibility shape, and freshness command. Add amendment metadata for intentional requirements changes and fail row-set, row-content, source-only, skipped-test, zero-match, and stale-proof cases.

### [Blocker] Target Identity Classification Omits Live Compatibility Behavior
**Sources:** Target Identity Classifier, API/CLI Worker Boundary Reviewer
**Issue:** The raw classifier no-write contract conflicts with current repair-on-read and materialize-on-miss behavior. Historical aliases are wrongly too narrow for scheduling/pool-resume/orphan-release routing. Per-surface error handling, no-store mail fallback, API rejected-by-config post-filtering, path-alias semantics, and candidate ordering/leak rules are not recorded in a way adapters can preserve.
**Required change:** Have the raw classifier emit read-only facts such as `repair-needed`, while adapter/command policy performs audited repair after revalidation. Add per-surface disposition tables for not-found, ambiguity, rejected-by-config, store error, repair-needed, and forbidden-kind. Add same-token collision matrices, historical-alias read-for-routing rules, no-raw-candidate leak tests, and exact adopting-surface proof before each delegation.

### [Blocker] API/CLI/Dashboard Route And Read Surfaces Lack Completeness Oracles
**Sources:** API/CLI Worker Boundary Reviewer, Target Identity Classifier, Operability Performance Diagnostics
**Issue:** `API_CLI_ROUTE_INVENTORY.yaml` tracks mutators but not raw lifecycle/identity reads, and it does not name enumeration authorities for Huma/OpenAPI routes, legacy mux routes, Cobra commands, `apiClient()` fallback sites, generated-client wrappers, dashboard/SSE projections, or local fallback behavior. Typed-wire and dashboard obligations are not mechanically tied to target/command diagnostics.
**Required change:** Add route/read inventory coverage for API, CLI, doctor, dashboard/SSE, and generated-client surfaces. Define completeness oracles per surface class, add `fixture_absent` or equivalent `delegation_blocked` row states, seed exact exceptions for current `genclient` and direct-manager paths, and require typed Huma structs, OpenAPI sync, generated-client checks, and `make dashboard-check` for API-visible changes.

### [Blocker] Diagnostics, Event Status, And Performance Budgets Are Not Machine-Checkable
**Sources:** Operability Performance Diagnostics, Event Delivery Contract Reviewer
**Issue:** `DIAGNOSTICS_MANIFEST.yaml` is load-bearing but lacks schema, owner, coverage semantics, renderer ownership, and vocabulary authority. The design overclaims event emission observability even though the current recorder does not return status. Performance budgets lack a counting backend, fake subscriber/runtime sources, cardinality fixtures, index/subprocess proof, and zero-match test-family enforcement.
**Required change:** Define the diagnostics manifest before Slice 0 decomposition. It must cover operation/check ID, owning slice, trace reason/outcome codes, wake causes, blockers, required facts, freshness/completeness, renderer surface, proof test, redaction keys, event payload relationship, and cost class. Rewrite event states to what current APIs can observe, and add executable query/subprocess-count budget proof with large-city fixtures and renderer goldens.

### [Major] Shared Vocabulary Is Still Too Broad And Too Early
**Sources:** YAGNI Contract Scope Reviewer, Reconciler Runtime Fact Reviewer, Operability Performance Diagnostics
**Issue:** Future-slice concepts such as `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent` can become active before a first real delegated production caller exists. Generic `committed facts` payloads and broad optional envelopes risk locking in abstractions that are not yet required.
**Required change:** Add row lifecycle states such as `documented`, `private`, `provisional`, and `delegating` for vocabulary artifacts. Keep target-classification vocabulary private to the first adopting adapter until a second surface proves identical fields. Remove or demote `SessionFactEvent` until a concrete slice and subscriber justify minimal typed payloads.

### [Major] Migration Sequencing, Worker Boundary, And Rollback Remain Mostly Prose
**Sources:** Migration Coexistence Strategist, API/CLI Worker Boundary Reviewer, Decider Atomicity Enforcer
**Issue:** The worker-boundary migration overlaps with session-boundary work, but exceptions around API session managers, session resolution, command factories, `worker.Handle` growth, and direct generated-client usage are not governed by one shared ledger. Rollback/data-direction convergence for transitional metadata is not testable.
**Required change:** Create an authoritative exception ledger with owner, allowed call sites, reason, expiry, retirement proof, and same-change update rules. Add named `worker.Handle` surface growth deliverables where needed. Add `session_design.data_direction_proof` or equivalent, plus reverted-direction convergence tests and rollback repair/scrubbing procedures.

### [Minor] Outcome And Diagnostic Vocabulary Drifts
**Sources:** Operability Performance Diagnostics, Event Delivery Contract Reviewer, Mutation Boundary Auditor
**Issue:** Outcome/status terms such as `durable-scan-converged`, `event-emission-failed`, `repair-applied`, `blocked`, `ambiguous`, `applied`, `skipped`, `partial`, and `failed` are not reconciled, and some tables mix naming conventions.
**Required change:** Declare the diagnostics manifest or vocabulary checkpoint file the single normative source, pick one naming convention, and define collision rules with existing trace outcome codes.

## Disagreements
- Several raw reviewers would approve only a tightly scoped, non-mutating Slice 0, while others block outright. Assessment: global block is still required because the design text can still drive behavior-moving work or a hollow Slice 0; approval can be reconsidered when Slice 0 is singular, physical, and enforceable.
- Some reviewers demand specific implementations such as prototype AST guards, `DiagnosticReport`, or incremental fact materialization. Assessment: the design may choose different local mechanisms, but it must prove the same properties: source-complete scans, typed diagnostics, bounded performance, and fail-closed ambiguity handling.
- Reviewers disagree about active-checkout versus `origin/main` evidence for scale, health, and progress proof files. Assessment: this is a baseline problem, not a product disagreement. The design must pin the baseline ref, ancestry rule, and reconcile/adopt plan.
- Close can be modeled with one canonical transitional shape or multiple close shapes. Assessment: either is acceptable only if each shape has durable intent, release snapshot, scanner predicate, completion fact, duplicate behavior, and crash tests.
- Event recovery language differs. Assessment: durable scans can be the convergence authority, but command diagnostics cannot claim event emission results unless a recorder/outbox API makes them observable.

## Missing Evidence
- Checked-in Slice 0 artifacts, schemas, validators, negative fixtures, and proof commands for the boundary inventory, writer/read inventory, scenario parity, command appliers, boundary matrix, route inventory, diagnostics manifest, vocabulary checkpoints, and allowlists.
- Source-complete writer and read inventories with nonzero scan assertions over production `cmd/` and `internal/` files.
- A machine-readable session-owned key taxonomy with exact keys, prefixes, regexes, dynamic-key rules, non-session classifications, owner slices, and default-deny behavior.
- A scheduled repair/backfill plan covering `RepairEmptyType`, runtime identity, continuation, no-token identity, and session-key paths.
- Exact command-applier rows and fault matrices for runtime start, close, wake, drain, rollback, event emission, provider start/stop, and partial metadata batches.
- A baseline ref and freshness policy for active checkout, tracked `origin/main`, offline/shallow/detached states, and stale or missing proof files.
- Assertion-level parity bindings for all requirements rows, plus owner-approved amendment metadata and row-content drift detection.
- Surface completeness oracles for API/Huma, legacy mux, Cobra CLI, `apiClient()` fallbacks, generated-client wrappers, dashboard/SSE, doctor, inspect, trace, mail, extmsg, nudge, attach, transcript/log, pool-resume, and assignee flows.
- Typed-wire, OpenAPI, generated-client, and dashboard proof for any API-visible target, command, conflict, candidate, or diagnostic field.
- Diagnostics renderer ownership and golden tests for trace, doctor, inspect, CLI, API, and event surfaces.
- Performance budget substrate with store/backend coverage, subprocess counting, fake subscribers, fake runtime observations, large-city fixtures, index proof or measured-baseline-relative limits, and zero-match test-family failures.
- Worker-boundary exception ledger and required `worker.Handle` growth plan.
- Data-direction, rollback, and revert convergence tests for transitional metadata.

## Recommended Changes
1. Make Slice 0 the explicit first and only first decomposable work item, consolidate its artifact contract, and add the close/finalizer gate that enforces `session_design.*` metadata and referenced artifact IDs.
2. Generate the source-complete mutation/read/route inventories and exact allowlists; remove blanket repair/doctor/migration exclusions.
3. Define the machine-readable session key taxonomy, repair/backfill owner slice, and one-writer/per-key coexistence rules before any behavior-moving slice.
4. Materialize `SCENARIO_PARITY.yaml` with all `SESSION-*` rows, exact proof selectors, stale/missing states, baseline ref, and owner-approved amendment workflow.
5. Define `COMMAND_APPLIERS.yaml` and the runtime-start/close/drain/wake/repair fault matrices, using durable intent before side effects and durable facts for recovery.
6. Make `BOUNDARY_MATRIX.yaml` operation-specific and enforce pure-decider isolation, demand carriage, cause-sensitive hold behavior, runtime-observation provenance, and zero-`Now` ownership.
7. Specify target-classification compatibility one surface at a time, including repair-needed facts, historical-alias read-for-routing, no-store mail behavior, API rejection post-filtering, path aliases, and raw-candidate no-leak tests.
8. Add API/CLI/dashboard completeness oracles, direct `genclient` and session-manager exception rows, and typed-wire/OpenAPI/dashboard proof requirements.
9. Define `DIAGNOSTICS_MANIFEST.yaml`, truthful event observability states, renderer owners, trace outcome compatibility, and performance-counting proof.
10. Reconcile migration sequencing with the active worker-boundary migration, including `worker.Handle` surface growth, rollback/data-direction proof, and a generated-inventory-vs-prose precedence rule.
