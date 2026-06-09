# Design Review Synthesis

## Overall Verdict: block

Worst-verdict-wins makes the global verdict `block`: the embed/materialization persona synthesis blocks on unresolved Core embed-source authority, while the remaining persona syntheses are `approve-with-risks`. The requirements artifact is broadly schema-conformant and much stronger than prior attempts, but it still leaves several load-bearing contracts to downstream inference and to missing external proof artifacts.

Schema conformance is not the main blocker. The document has valid requirements front matter and the required top-level sections in order; the output-shape mismatches are cleanup and boundary issues, while the blocking risks are behavior/design authority, proof, rollout, and workflow-artifact integrity risks.

## Consensus Strengths
- Reviewers agree on the product direction: Core is required SDK runtime support, Maintenance is retired as an active pack, and Gastown behavior becomes explicit public pack configuration.
- AC6, AC7, AC14, AC15, AC16, and AC17 now name the right evidence chain: asset ledger, behavior-preservation proof, public Gastown pin/proof, version-skew policy, offline/cache behavior, and acceptance proof matrix.
- The draft no longer tries to inline a full file-by-file migration table in `requirements.md`; ownership and behavior proof are delegated to support artifacts.
- Doctor/import-state diagnostics are first-class requirements, including exact source attribution, report-only default behavior, non-interactive repair, bootstrap diagnostics, and live-state refusal.
- Role neutrality is directionally correct: `dog` is acceptable only as configurable pack data, and SDK infrastructure remains controller-owned rather than role-dependent.
- The requirements explicitly reject hidden fallbacks through in-tree examples, stale system packs, embedded synthetic aliases, or cache substitutions.

## Critical Findings

### [Blocker] Core Embed-Source Authority Is Still Ambiguous
**Sources:** embed-materialization-build-reviewer; pack-resolution-architect; external-pack-docs-reviewer; zfc-role-neutrality-guardian
**Actionability:** document-fixable
**Issue:** The requirements list `internal/bootstrap/packs/core` among legacy mixed roots while Core remains required and release-bundled. Existing Go embed, builtin registry, materialization, and public-alias seams can therefore keep a second active Core or Gastown authority alive unless the final Core source-root policy is explicit.
**Required change:** State the final canonical Core embed/source authority and require closure over Go embed packages, builtin registry functions, synthetic layouts/hashes, public subpath aliases, materialization commands, direct hook imports, and tests. If `internal/bootstrap/packs/core` remains, declare it canonical or prove it is a non-runtime fixture/shim.

### [Major] Required Support Artifacts Are Still Evidence Gaps
**Sources:** migration-rollout-reviewer; embed-materialization-build-reviewer; gastown-behavior-preservation-auditor; asset-classification-split-reviewer; test-strategy-ci-reviewer; requirements-schema-compliance-officer
**Actionability:** external-prerequisite
**Issue:** The requirements name the correct support artifacts, but the lane syntheses repeatedly note that the actual checked artifacts, schemas, snapshots, validators, proof commands, and gate wiring are not present yet. Without them, implementation can still invent policy or pass with stale, incomplete, or self-selected evidence.
**Required change:** Produce and validate the pack-resolution matrix, source-consumer closure, asset migration ledger, behavior-preservation manifest or harness, migration diagnostics schema, role-neutrality scan, public Gastown pin/version-skew/cache proofs, and AC17 acceptance-proof matrix before implementation approval.

### [Major] Behavior Preservation Is Still One-Sided And Too Coarse
**Sources:** gastown-behavior-preservation-auditor; asset-classification-split-reviewer; test-strategy-ci-reviewer; embed-materialization-build-reviewer
**Actionability:** document-fixable
**Issue:** AC7 protects a supported-Gastown before-state, but reviewers warn that generic Core/non-Gastown behavior can still be swept into Gastown-only rows and disappear from non-Gastown cities. File-level inventory is also too coarse for multi-target scripts, templates, notification targets, identity effects, requester/detector links, and failure/escalation paths.
**Required change:** Add a symmetric Core/non-Gastown behavior baseline or positive evidence rule for single-owner rows. Bind AC6 and AC7 bidirectionally at behavior-row or call-site granularity, with stable behavior/sub-asset IDs and explicit Core, Gastown, retired, or out-of-scope outcomes.

### [Major] Pack Resolution Needs More Product-Level Determinism
**Sources:** pack-resolution-architect; zfc-role-neutrality-guardian; external-pack-docs-reviewer; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The draft establishes deterministic resolution in principle, but still lacks exact `bd`/`dolt` provider-pack cardinality, healthy baseline precedence, release identity plus content digest for Core provenance, public Gastown declaration ownership or dedupe, overlay rules, same-named optional public-pack collision behavior, and local override support for cross-repo pin validation.
**Required change:** Tighten AC3/AC15/AC16 so init, doctor, CLI load, and runtime resolution agree on Core identity, source attribution, condition codes, pack layer ordering, public pin validation, overlay behavior, same-name conflicts, and transitive diamond diagnostics.

### [Major] Existing-City Repair And Diagnostics Need Concrete Operator Semantics
**Sources:** doctor-diagnostics-safety-reviewer; migration-rollout-reviewer; pack-resolution-architect; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The diagnostic direction is strong, but missing-Core repair targets, bootstrap-only command isolation, recursive import-chain tracing, exit codes, post-fix verification fields, transaction backup/journal behavior, stale directory guidance, inactive unresolved bead policy, rollback/downgrade handling, and offline cache seeding remain under-specified.
**Required change:** Define the `gc doctor`/`gc import-state` condition-code registry, exit-code matrix, stdout/stderr contract, recursive provenance model, bootstrap diagnose/fix boundary, transaction model, live-state refusal evidence, stale-directory cleanup guidance, task-store migration policy, and rollback/offline outcomes.

### [Major] Role-Neutral Binding And Scan Boundaries Are Not Tight Enough
**Sources:** zfc-role-neutrality-guardian; test-strategy-ci-reviewer; doctor-diagnostics-safety-reviewer; requirements-schema-compliance-officer
**Actionability:** document-fixable
**Issue:** AC8/AC9 permit a configurable default executor named `dog`, but the carve-out can still be misread as permission for literal role names in Core route targets, notification targets, formula bindings, prompt defaults, overlays, or Go fallbacks. Diagnostic use of retired identifiers also needs a narrow data surface so role scans do not become broad exceptions.
**Required change:** Limit literal `dog` and other role-name occurrences to exact configured-default data keys and tightly scoped diagnostic/migration fixtures. Require symbolic bindings, data-declared optionality and escalation behavior, generic override mechanisms, rendered-template checks, generated/materialized scan coverage, and positive/negative controls.

### [Major] Test And Validation Gates Can Still Pass Vacuously
**Sources:** test-strategy-ci-reviewer; pack-resolution-architect; gastown-behavior-preservation-auditor; asset-classification-split-reviewer
**Actionability:** document-fixable
**Issue:** Reviewers repeatedly flag validators that could pass by seeing mapped filenames, skipped tests, no-op witnesses, empty post-deletion walks, or broad scan allowlists. AC17 also needs a named validation command, make target, pre-commit entry, or CI job, not only an artifact name.
**Required change:** Require AC7/AC13 validators to parse active test execution evidence, such as `go test -json`, and fail on skipped, empty, or no-op witnesses. Pin AC6/AC13 denominators to frozen digest-verified baselines, require assertion- or behavior-level coverage transfer, and wire AC17 into an executable gate.

### [Major] Review Workflow Artifacts Are Still Misfiled
**Sources:** artifact check; persona syntheses; raw attempt-7 reviews
**Actionability:** workflow-defect
**Issue:** `.gc/design-reviews/ga-dtvdnd/attempt-7/persona-syntheses/` is empty even though ten fresh persona syntheses exist under `.gc/design-reviews/ga-dtvdnd/attempt-1/persona-syntheses/`. The raw attempt-7 review files also describe a known attempt-path defect and dual-write workaround. I used the ten fresh misplaced syntheses for substantive review, but the active attempt directory still lacks the required persona synthesis artifacts.
**Required change:** Fix the design-review workflow so persona synthesis output paths and metadata use the active attempt directory, and fail before global synthesis when attempt-local persona syntheses are missing or stale.

### [Minor] Requirements Schema And Approval Hygiene Need Cleanup
**Sources:** requirements-schema-compliance-officer; test-strategy-ci-reviewer; doctor-diagnostics-safety-reviewer
**Actionability:** document-fixable
**Issue:** The artifact conforms to `gc.mayor.requirements.v1`, but `<!-- REVIEW: added per pack-resolution-contract -->` is workflow residue, some Problem Statement wording is mechanism-heavy, and `status: draft` remains correct until explicit human approval. Support-artifact detail is acceptable as acceptance evidence, but concrete schemas, commands, owners, freshness rules, and sequence belong in the implementation plan or support artifact schemas.
**Required change:** Remove review residue, reword mechanism-heavy product text or explicitly declare it fixed product contract, keep `status: draft` until human approval, and move implementation data models and command details to downstream artifacts.

## Disagreements
- Embed/materialization verdicts split: Codex blocks, while Claude and DeepSeek approve with risks. Assessment: `block` is correct because the Core source-root ambiguity can preserve an undeclared active authority.
- Behavior preservation severity differs. Codex and DeepSeek are more willing to accept downstream artifact gates; Claude stresses that "supported Gastown" and Gastown-triggered Maintenance-origin behavior can be self-selected. Assessment: downstream proof is acceptable only if the denominator and AC6/AC7 equality rules fail closed.
- Schema reviewers agree the artifact is schema-shaped, but differ on whether review residue and mechanism-heavy wording must be fixed before approval. Assessment: these are document-fixable hygiene issues, not the global block reason.
- Reviewers disagree on how concrete split-row schema must be in requirements. Assessment: the exact representation can live in support schemas, but requirements must require every split/core-renamed row to prove all outputs and retirements.
- Offline/public-pack reviewers balance fail-closed cache misses against mirror support for air-gapped operators. Assessment: keep fail-closed behavior and add a sanctioned mirror or Git redirect path; do not fall back to embedded or in-tree content.
- Some DeepSeek findings reference the implementation plan and treat planned mechanisms as evidence. Assessment: carry useful design pins forward, but do not treat unvalidated implementation-plan text as satisfying the missing support-artifact gates.

## Missing Evidence
- Canonical Core embed/source-root policy and builtin registry/materialization closure proof.
- Checked `source-consumer-closure.yaml`, `asset-migration-ledger.yaml`, `behavior-preservation-manifest.yaml`, `pack-resolution-matrix.yaml`, `migration-diagnostics.schema.json`, `role-neutrality-scan.yaml`, `public-gastown-pin-ledger.yaml`, `version-skew-matrix.yaml`, and `acceptance-proof-matrix.yaml`.
- Validated schemas, owners, freshness rules, source snapshots, generator/validator commands, and gate placement for those artifacts.
- Public Gastown candidate commit, immutable `sha:` pin, pack digest, behavior-manifest digest, lock/cache provenance, fallback-disabled validation result, and cross-repo release order proof.
- Symmetric Core/non-Gastown behavior baseline and AC6-to-AC7 row/call-site equality proof.
- Existing-city happy-path upgrade example, repair transaction model, exit-code matrix, recursive import-chain tracing, post-fix JSON fields, inactive unresolved bead policy, and rollback/offline cache-seeding story.
- Exact AC8 denied-token inventory, allowed paths/keys, generated/materialized and rendered-output scan coverage, and non-`dog` binding controls.
- Active witness execution evidence proving mapped replacement tests run and fail when moved behavior is absent or broken.
- Cross-repo documentation audit of the public Gastown pack at the accepted pin.
- Attempt-local persona synthesis files under `.gc/design-reviews/ga-dtvdnd/attempt-7/persona-syntheses/`.

## Convergence Assessment
- Remaining blocker class: mixed
- Recommended apply verdict: iterate
- Reason: another design-doc edit can directly resolve at least one blocking document-fixable issue, especially the Core embed-source authority, AC6/AC7 behavior-baseline rules, pack-resolution determinism, diagnostic/repair semantics, role-scan boundaries, and AC17 gate wording. However, final approval will still require non-design work: checked support artifacts, public Gastown proof, executable validators, and a workflow path fix.
- Next non-design work: create and validate the support artifacts and proof commands; produce the public Gastown pin/proof and cross-repo docs audit; wire AC17 into CI/pre-commit or named local gates; fix the design-review workflow so persona syntheses are written to the active attempt directory.

## Recommended Changes
1. Resolve the Core embed/source-root policy and make Go embed, builtin registry, synthetic alias, and materialization closure an explicit acceptance target.
2. Produce or gate on the named support artifacts, with schemas, validator commands, owners, freshness rules, and AC17 wiring before implementation approval.
3. Strengthen AC6/AC7 with symmetric Core/non-Gastown behavior preservation, behavior-row/call-site granularity, multi-output split representation, and bidirectional proof.
4. Tighten AC3/AC15/AC16 for provider-pack cardinality, Core release identity plus digest, public Gastown declaration ownership, overlays, same-named optional public-pack conflicts, local pin overrides, and cross-surface condition-code consistency.
5. Define existing-city diagnostic and repair semantics end to end: bootstrap-only isolation, missing-Core repair targets, recursive provenance, exit codes, transaction backup/journal, live-state refusal, inactive unresolved bead policy, stale-directory guidance, rollback, and offline cache seeding.
6. Narrow role-neutrality exceptions around literal `dog` and retired identifiers to exact data keys and diagnostic fixtures, with symbolic bindings and generated/rendered-output scans.
7. Make validators non-vacuous: frozen baselines, assertion/behavior-level coverage transfer, active `go test -json` witness checks, negative controls, and no broad scan exclusions.
8. Add public-pack authority details: retire or isolate embedded Gastown sources, define retained `examples/gastown` behavior, audit public-pack docs at the pin, and support sanctioned mirror/redirect behavior without hidden fallback.
9. Clean requirements output shape before approval: remove the review marker, reduce mechanism-heavy Problem Statement prose or mark it as product contract, and leave `status: draft` until explicit approval.
10. Fix the review workflow's active-attempt output paths for persona syntheses before relying on the next review loop as clean evidence.
