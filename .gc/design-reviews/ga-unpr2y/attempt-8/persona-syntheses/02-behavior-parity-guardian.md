# Natasha Volkov

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The parity audit has no safe baseline rule. Claude found this checkout is far behind `origin/main` and that several files classified as missing are present upstream; DeepSeek found every cited historical commit is a non-ancestor of active HEAD; Codex agrees commit-only proof must become a failing freshness condition. The design must pin the audit and implementation baseline ref before any generated parity artifact can be trusted.
- [Blocker] The executable parity source and freshness gates do not exist yet. `internal/session/SCENARIO_PARITY.yaml` and `TestScenarioParityFreshness` are still design intent rather than checked-in proof, so any extraction slice can drift from `REQUIREMENTS.md` unless Slice 0 lands first and fails closed.
- [Blocker] Scenario-to-slice ownership is not yet decomposable. Claude and DeepSeek identify rows whose baseline ownership and traceability matrix assignment diverge, including `SESSION-LIFE-002`, `SESSION-LIFE-006`, and `SESSION-WORK-003`; Codex is less concerned about row presence in prose but still requires a canonical YAML source before bead generation. The synthesis assessment is that prose tables are not enough for parity-safe decomposition.
- [Major] Current proof is not consistently behavioral proof. Source files and historical commits are being used where the lane requires runnable characterization tests, exact output contracts, or precise static assertions. This is especially risky for mail, API, CLI, extmsg, reconciler, and pool-work recovery behavior.
- [Major] Owner-approved requirements amendments are a required concept but not a verifiable artifact. Claude and Codex both flag that a refactor-convenient ledger edit and an approved behavior change would look identical to the proposed gates.
- [Major] The design has behavior not anchored by a requirements row. Claude found `RepairEmptyType` call sites, including resolver-internal call sites, whose current repair-on-read side effects are not represented in `REQUIREMENTS.md`; Slice 1 could therefore change production behavior without a parity owner.
- [Major] Audit inputs can be dirty or checkout-relative. Claude found the baseline artifact records `HEAD` but not a clean working tree, and the raw review context shows untracked or modified session artifacts can affect what gets treated as canonical proof.

**Disagreements:**
- Verdict severity differs: Claude and DeepSeek returned `block`, while Codex returned `approve-with-risks`. My assessment is `block` because this persona's lane is regression prevention, and all reviewers require Slice 0 parity infrastructure before mutation-owning work proceeds.
- Codex says the design names every `SESSION-*` row in prose; Claude and DeepSeek say the traceability matrix still omits or misassigns rows. My assessment is that the executable artifact, not prose, must be authoritative, and it does not exist yet.
- DeepSeek classifies `SESSION-RECON-006`, `SESSION-RECON-007`, and `SESSION-WORK-003` as absent or phantom behavior on active HEAD. Claude shows at least some related proof exists on `origin/main`, which means the conclusion depends on the chosen baseline. The required response is to pin the baseline and re-audit, not to accept either checkout-relative claim unconditionally.
- DeepSeek suggests banning commit-hash citations; Claude and Codex allow them as historical context when paired with live current-lineage proof. This synthesis adopts the latter: commits may explain history, but they cannot be current proof unless the ancestry gate and live proof both pass.
- DeepSeek says Slice 0 should exist before design-review approval; Codex frames the narrower requirement as allowing only Slice 0 to proceed before other implementation. The common blocker is that no extraction or behavior-moving bead should be generated until Slice 0 artifacts exist and pass.

**Missing evidence:**
- A recorded baseline ref and clean-worktree requirement for generating `SLICE0_BASELINE.md`, `BOUNDARY_INVENTORY.md`, and `SCENARIO_PARITY.yaml`.
- A checked-in `internal/session/SCENARIO_PARITY.yaml` covering all 45 `SESSION-*` rows with primary slice, secondary slices, operation surfaces, current oracle, compatibility shape, and freshness command.
- Implemented freshness tests that fail when proof paths are absent, commit citations are non-ancestors of the pinned baseline, test patterns match zero tests, or a row has only historical/source evidence.
- Characterization tests or exact output contracts for source-only proof areas, including API, CLI, mail, extmsg target handling, assignee normalization, nudge, attach, inspect/log/transcript, materialization, reconciler, and pool-work recovery.
- Owner-approval evidence format for intentional `REQUIREMENTS.md` behavior changes, including where the approval lives and how gates verify it.
- Requirements coverage and disposition for empty-type repair-on-read behavior across resolver, CLI pin, beadmail, and API callers.
- Re-audited status of `SESSION-RECON-006`, `SESSION-RECON-007`, and `SESSION-WORK-003` on the pinned baseline.

**Required changes:**
- Add a Slice 0 baseline rule: run parity and inventory generation on a clean worktree at the pinned implementation baseline ref, record that ref, and anchor commit-ancestry checks to it.
- Land `SCENARIO_PARITY.yaml` and `TestScenarioParityFreshness` before any mutation-owning extraction slice. Generated beads must consume the YAML, not prose tables.
- Reconcile the ownership baseline and traceability matrix into one source of truth. Explicitly assign all 45 rows, including `SESSION-LIFE-002`, `SESSION-LIFE-006`, `SESSION-WORK-003`, baseline/no-touch rows, and multi-slice secondary obligations.
- Treat commit hashes as historical unless `git merge-base --is-ancestor` passes against the pinned baseline and a live source/test oracle also exists.
- Treat production source paths as inventory unless paired with a characterization test, exact command/output contract, or precise static assertion for the required behavior.
- Define a durable requirements-amendment approval artifact, such as bead or mail metadata, and make the freshness/parity gate fail behavior-ledger changes without it.
- Add or amend requirements rows for `RepairEmptyType` repair-on-read behavior before Slice 1 changes resolver behavior; record owner-approved preserve or retire disposition.
- Re-audit `SESSION-RECON-006`, `SESSION-RECON-007`, and `SESSION-WORK-003` on the pinned baseline, then restore proof, re-cite live proof, or amend the requirements with owner approval.
