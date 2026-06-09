# Design Review Synthesis

## Overall Verdict: block

Worst-verdict-wins produces `block`: eight persona syntheses block, and the two `approve-with-risks` lanes still require Slice 0 artifacts before any mutation-owning work proceeds. The design has moved in the right direction on durable scans, observation taxonomy, and deferring speculative vocabulary, but the current artifact set is not executable enough to protect session lifecycle ownership, parity, recovery, or operator visibility.

## Consensus Strengths
- Multiple personas praised the shift from event-dependent correctness to idempotent durable scans, with events treated as latency and diagnostic hints rather than the only recovery path.
- Reviewers consistently recognized the provider/runtime observation completeness taxonomy as a substantial improvement over boolean liveness facts.
- The non-mutating Slice 0 concept, shrink-only exception discipline, and per-key rollback framing are directionally sound migration controls.
- The target-classifier direction is useful if it remains read-only and surface adapters preserve existing CLI/API/mail/extmsg behavior.
- Reviewers agreed the design is better at deferring speculative vocabulary than earlier attempts, especially by avoiding a broad `SessionFacts` facade.

## Critical Findings

### [Blocker] Slice 0 gates are still prose, not executable protection
**Sources:** Elena Marchetti, Natasha Volkov, Takeshi Yamamoto, Sarah Chen, Ravi Krishnamurthy, Ingrid Holm
**Issue:** The required preflight artifacts are absent or not source-complete: boundary inventory, scenario parity, vocabulary checkpoints, static guard tests, allowlist fixtures, and proof commands. Existing inventory rows are too broad for decomposition, current proof is checkout-relative or stale, and several gates can appear to pass with no matching tests.
**Required change:** Make Slice 0 the first deliverable and block all mutation-owning beads until it lands. Pin the audited baseline, commit the requirements/design ledgers, generate a source-complete symbol/key/endpoint inventory, install failing guard fixtures and shrink-only allowlist checks, and make parity/vocabulary proof commands fail on missing artifacts or zero matched tests.

### [Blocker] Command atomicity and stale-writer safety are not physically guaranteed
**Sources:** Takeshi Yamamoto, Elena Marchetti, Amara Osei, Ravi Krishnamurthy
**Issue:** The design still implies conditional commit safety that the current store surface does not provide. Blind metadata writes, sequential `SetMetadataBatch`, missing CAS/revision primitives, lock-free `instance_token` backfills, and dual close paths leave stale-success and partial-state windows. `ProjectLifecycle` also violates pure-decider expectations with a wall-clock fallback.
**Required change:** State the real store primitive limits and either add a conditional-write primitive or explicitly classify each blind write as inert-when-lost or repair-converged. Remove the decider clock fallback, add stale-writer race tests, prove backend batch atomicity or add partial-state repair matrices, and unify or equivalently fence runtime-start, close, token backfill, and identity-retirement command paths.

### [Blocker] Close/work-release recovery and successor safety remain under-specified
**Sources:** Amara Osei, Takeshi Yamamoto, Ingrid Holm
**Issue:** Close/retire can commit terminal state or clear identity fields before work release and binding cleanup are durably recoverable. Events can be missed, old assignment keys may become unreconstructable, and delayed cleanup can collide with a live successor using the same name or alias.
**Required change:** Add a durable close/retire recovery contract: preserve release keys or use a transitional `closing`/`work_released=false` state, name the exact scanner/helper and cadence, define idempotency and supersession keys, and add crash-after-commit, event-miss, duplicate-scan, partial-query, and stale-successor tests.

### [Blocker] The session/reconciler boundary cannot preserve idle and forced-awake behavior yet
**Sources:** Liam Okonkwo, Takeshi Yamamoto, Ingrid Holm, Ravi Krishnamurthy
**Issue:** The proposed eligibility mask is too flat: `runnable` cannot distinguish forced-awake sessions from conditionally runnable sessions subject to idle sleep. `AwakeInput` still mixes session facts, controller demand, runtime observation, work aggregation, pool policy, and health/progress concerns. Provider-health and progress proof also depends on upstream code that is not present in this checkout.
**Required change:** Add a field-by-field `AwakeInput` disposition table, split desired-running precedence into session eligibility versus reconciler composition, return forced versus conditional runnable semantics or an equivalent structured mask, add idle-sleep rows to ownership tables, define typed runtime observation facts, and recover the pinned upstream provider-health/progress implementation and tests instead of re-deriving them from prose.

### [Blocker] Target resolution and API/CLI projection parity lack surface-specific contracts
**Sources:** Amara Diallo, Sarah Chen, Natasha Volkov, Elena Marchetti
**Issue:** Slice 1 can still mix read-only classification with repair, materialization, runtime-start, close/retire, and metadata normalization side effects. Existing API, CLI, mail, extmsg, assignee, nudge, attach, inspect, log, transcript, and pool-resume surfaces have different precedence, ambiguity, historical-alias, closed-ID, output, and wire-shape behavior that is not fully enumerated.
**Required change:** Add a target-collision matrix and per-surface adapter chains with terminal/fallthrough semantics. Split mail send from mail query, enumerate assignee-normalization variants, define historical-alias allow/deny rules, preserve `apiClient()` and local fallback behavior, and require endpoint/command rows with current status codes, problem bodies, request IDs, stdout/stderr/JSON shapes, exit codes, async events, and generated-client impact.

### [Blocker] Repair, diagnostics, and performance gates are not operable yet
**Sources:** Ingrid Holm, Sarah Chen, Kwame Asante
**Issue:** `RepairEmptyType` remains an operator-invisible read-path repair write with swallowed persistence failures and no concrete slice owner. New trace/doctor/session-inspect vocabulary is not mapped to centralized constants or renderers. Query-count, subscriber-budget, backpressure, and hot-loop write budgets have no canonical measurement substrate or baseline.
**Required change:** Assign repair writes to a real repair slice or audited helper in Slice 0/1, propagate persistence errors, emit typed before/after trace evidence, and make direct repair writes guard-visible. Add a diagnostics mapping artifact with trace constants, reason/outcome codes, renderers, and tests. Define a shared counting store or benchmark fixture, commit baseline measurements, and add one subscriber budget helper plus reconciler-write budget rules.

### [Major] Migration and vocabulary scope are viable but need tighter delivery boundaries
**Sources:** Ravi Krishnamurthy, Kwame Asante
**Issue:** Slice 1 can be read as one oversized delivery unit, and later-slice vocabulary can still leak into active checkpoints before a first caller exists. Per-key-family rollback and one-writer proof are asserted more clearly than they are enforced.
**Required change:** State that Slice 1 lands incrementally per adopting surface, with taxonomy tables as closed bounds rather than a single implementation obligation. Split active slice vocabulary from provisional later-slice vocabulary, forbid exported types or checkpoint rows before first caller evidence, and require each implementation bead to list converted callers, allowed legacy rows, retired rows, guard changes, rollback path, proof commands, and one-writer proof.

### [Minor] Attempt artifact paths are inconsistent
**Sources:** Workflow artifacts observed during synthesis
**Issue:** The current attempt directory lacks `attempt-7/persona-syntheses/*.md`, while the ten persona synthesis beads for Attempt 7 stamped fresh outputs under `attempt-1/persona-syntheses/`. This synthesis used those ten child-stamped persona outputs because they contain the Attempt 7 content and all fragment finalizers closed pass.
**Required change:** Fix the child synthesis output path or attempt selection so future global synthesis steps read the current attempt directory without relying on stale-path metadata.

## Disagreements
- Verdict severity differed in several lanes. Some Claude reviews returned `approve-with-risks`, but Codex/DeepSeek/Gemini reviewers blocked the same areas when the issue affected decomposition safety, parity, or recovery. The global assessment follows worst-verdict-wins and treats those gaps as blockers before mutation-owning work.
- Reviewers disagreed on whether missing Slice 0 files are a design blocker or a first implementation deliverable. The practical conclusion is the same: only Slice 0 preflight can proceed, and no lifecycle mutation, close, wake, runtime-start, or repair conversion should be decomposed until those artifacts exist and fail correctly.
- Baseline evidence is disputed because the active checkout is stale relative to upstream. The design must pin the baseline and re-audit there rather than choose between active-checkout findings and upstream-main assumptions.
- Close/work-release reviewers proposed different recovery handles: release-key preservation, transitional close state, tokened work assignments, or a proven live-successor guard. Any mechanism is acceptable only if it is durable after restart and tested against stale successor reuse.
- The runtime observation taxonomy was widely praised, but reviewers disagreed on whether it is sufficient without a concrete typed fact surface. This synthesis treats the taxonomy as a strength and the missing implementable contract as a blocker.

## Missing Evidence
- Version-controlled Slice 0 artifacts: `BOUNDARY_INVENTORY.md`, `SCENARIO_PARITY.yaml`, `VOCABULARY_CHECKPOINTS.yaml`, static guard tests, allowlist fixtures, and CI/pre-commit proof commands.
- Pinned baseline ref, ancestry checks for cited commits, committed requirements/design ledgers, and a complete 45-row scenario parity mapping with exact oracle commands.
- Guard proof for dynamic metadata keys, generic bead routes, Huma/legacy API bridge writes, package-level mutators, patch-map extension, subprocess `bd` mutation, repair writes, and non-session-bead false positives.
- Race and partial-state tests for runtime-start, close, wake, drain, token backfill, duplicate prepare, stale no-token paths, batch metadata writes, provider-stop/write-fail, close-retire partial failures, event misses, and crash-after-commit recovery.
- Field-level reconciler/session disposition for `AwakeInput`, forced versus conditional runnable semantics, runtime observation facts, provider-health/progress baseline recovery, and idle-sleep decision/mutation ownership.
- Target/API/CLI parity matrices covering configured-name conflicts, historical aliases, exact closed IDs, mail send/query differences, assignee normalization variants, async Huma behavior, `apiClient()` routing, local fallback, and output/wire no-leak rules.
- Trace/doctor/session-inspect mapping, repair ownership, performance measurement substrate, baseline query counts, subscriber budget helper, backpressure behavior, and reconciler hot-loop write budgets.

## Recommended Changes
1. Limit approval to a non-mutating Slice 0 preflight and implement the inventory, parity, vocabulary, guard, allowlist, baseline, and proof-command artifacts before any mutation-owning decomposition.
2. Rewrite atomicity claims to match the actual store surface, then add stale-writer, partial-batch, decider-purity, close-path, token-backfill, and repair-convergence tests.
3. Define durable close/work-release recovery and successor safety with explicit persisted facts, scan owner, idempotency/supersession keys, and crash/event-miss tests.
4. Split the session/reconciler pipeline into field-disposition, eligibility-mask, runtime-observation, idle-sleep, health/progress, and policy-composition contracts.
5. Make target resolution and API/CLI projection surface-specific, with collision matrices, route inventories, golden output or typed-response oracles, and explicit no-leak rules.
6. Give repair, diagnostics, and performance a concrete owner, centralized vocabulary, rendered operator surfaces, measurement fixture, baseline budgets, and shared subscriber/reconciler budget helpers.
7. Constrain migration and vocabulary delivery per adopting surface and per owned-key family, with rollback and one-writer proof carried by each implementation bead.
