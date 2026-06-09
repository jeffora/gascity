# Design Review Synthesis

## Overall Verdict: block

Attempt 5 remains blocked by worst-verdict-wins: five persona syntheses returned `block`, while the remaining lanes returned `approve-with-risks`. The requirements artifact is now structurally close to `gc.mayor.requirements.v1`, but the review converges on unresolved executable contracts for asset ownership, public Gastown rollout, retired source disposition, diagnostics/repair, and proof gates that implementation agents would otherwise have to invent.

The active attempt directory still has a workflow artifact defect: `.gc/design-reviews/ga-dtvdnd/attempt-5/persona-syntheses/` is empty, while the ten current-timestamp persona syntheses were written under `.gc/design-reviews/ga-dtvdnd/attempt-1/persona-syntheses/`. I used those ten current syntheses, plus the attempt-5 raw reviews where helpful, and classify the path problem separately as a workflow defect.

## Consensus Strengths
- Reviewers agree the current requirements are far stronger than prior attempts: front matter is present, required top-level sections are in order, W6H and Example Mapping are concrete, inline review comments are gone, and `Open Questions` begins with `None`.
- Core is correctly framed as required Gas City runtime behavior, while Gastown behavior is explicit public pack configuration and Maintenance is retired as a standalone active pack.
- AC6, AC7, AC14, AC15, AC16, and AC17 are the right proof anchors: asset ownership, behavior preservation, public-pack validation, pin coherence, offline/cache safety, and acceptance traceability.
- The diagnostic posture is directionally sound: report-only by default, explicit non-interactive mutation, idempotence, atomicity, resumability, source attribution, bootstrap diagnostics, and live-state repair refusal.
- Role neutrality is broadly aligned with ZFC: the default `dog` executor is allowed only as configurable pack data, not as Go-side routing, prompt, or fallback knowledge.

## Critical Findings

### [Blocker] AC6 Asset Ledger Is Not Yet Executable
**Sources:** asset-classification-split-reviewer; requirements-schema-compliance-officer; gastown-behavior-preservation-auditor; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The requirements correctly require an asset migration ledger, but the accepted contract still leaves too much to downstream design: exact artifact path, validator command, ownership, source snapshot, bidirectional validation, split-row semantics, behavior IDs or sub-asset keys, classification vocabulary, and unresolved `review` row lifecycle. The current implementation-plan language also uses singular destination fields that cannot safely represent split assets.
**Required change:** Tighten AC6 so `plans/core-gastown-pack-migration/requirements.md` names or mandates a named support artifact, validator command, owner, closed classification vocabulary, stable behavior or sub-asset IDs, all Core/Gastown/retired outputs, basename collision checks, bidirectional validation, and a zero-unresolved-review gate before implementation planning, decomposition, or source moves depend on the ledger.

### [Blocker] Public Gastown And Retired Source Disposition Are Not Proven
**Sources:** external-pack-docs-reviewer; migration-rollout-reviewer; gastown-behavior-preservation-auditor; pack-resolution-architect
**Actionability:** external-prerequisite
**Issue:** The current public Gastown pin is not proven split-compatible, and reviewers found that the in-tree `examples/gastown/packs/{gastown,maintenance}` roots can remain a resolvable parallel authority. AC4, AC12, and AC14 can pass fresh-init and wording checks while an active in-tree fallback or stale `../maintenance` import still masks a broken external pack.
**Required change:** Produce or select a split-compatible public Gastown commit and pin ledger, prove behavior from the external checkout or pinned cache with in-tree fallback disabled, and require an enforceable source-tree disposition: delete the old pack roots or isolate them as non-resolvable test fixtures excluded from runtime resolution, docs authority, init templates, and public-pack proof.

### [Blocker] Two-Repo Rollout And Version Skew Remain Underspecified
**Sources:** migration-rollout-reviewer; pack-resolution-architect; external-pack-docs-reviewer; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The draft names compatibility pins, activation pins, `PublicGastownPackVersion`, pack digests, behavior-manifest digests, and version-skew diagnostics, but it does not yet make those artifacts operationally precise. Reviewers also found old-binary/new-pack, new-binary/old-lock, offline cache-hit/cache-miss, downgrade, and rollback states that are not testable from the current contract.
**Required change:** Add a rollout matrix that defines compatibility and activation pins, their artifact paths and ownership, lock/cache relationship, supported and unsupported old/new binary states, downgrade expectations, offline migration-repair behavior, and exact fail-closed diagnostics.

### [Blocker] Doctor Repair And Rollback Contract Is Not Testable
**Sources:** doctor-diagnostics-safety-reviewer; migration-rollout-reviewer; pack-resolution-architect; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The requirements require safe repair, but the accepted contract still allows multiple possible mutation commands and does not define the diagnostic JSON schema, stable condition codes, recursive import-chain provenance, non-destructive TOML guarantees, bootstrap repair boundary, interrupted-run recovery, rollback/restore behavior, read-only or transitive import remediation, or affected live-session evidence.
**Required change:** Choose the canonical mutating repair command for each condition, define typed diagnostics and golden text/JSON output, require recursive import-chain tracing, state bootstrap diagnose/fix behavior, specify durable preflight/journal/backup or refusal semantics, preserve unrelated config/comments/ordering, and name live process/session evidence used for repair refusal.

### [Blocker] Maintenance And Dog Producer Behavior Still Lacks Closure
**Sources:** migration-rollout-reviewer; embed-materialization-build-reviewer; gastown-behavior-preservation-auditor; zfc-role-neutrality-guardian
**Actionability:** external-prerequisite
**Issue:** Reviewers re-identified producer and housekeeping behavior such as reaper, jsonl export, spawn-storm detection, `dolt-target.sh`, status-line tracing, and related Dog/default-executor flows that currently live in or reference retired Maintenance. The migration cannot safely retire Maintenance unless those behaviors are moved, generalized, split, or intentionally retired with witnesses.
**Required change:** Complete a source-plus-consumer closure inventory that records each current referencing file, old reference, replacement or retirement decision, owner, and proof command, then validate representative consumers with Maintenance absent from materialized system packs.

### [Major] Behavior Preservation Proof Needs AC6-Level Precision
**Sources:** gastown-behavior-preservation-auditor; asset-classification-split-reviewer; test-strategy-ci-reviewer; external-pack-docs-reviewer
**Actionability:** document-fixable
**Issue:** AC7 has the right scope, but it still lacks a concrete before-state denominator, artifact path, schema, generator/validator command, owner, witness types, failure conditions, and AC6-to-AC7 reconciliation rule. File-level coverage is too coarse for formulas, scripts, prompt fragments, notification paths, identity side effects, and multi-target assets.
**Required change:** Strengthen AC7 to require a generated or enumerated supported-Gastown baseline, row-level or call-site-level witnesses, scripts and `assets/scripts` in the public checkout proof set, Core/controller-owned outcomes as valid trace destinations, and fail-closed reconciliation for every behavior-bearing AC6 row.

### [Major] Pack Resolution And Core Identity Need A Precedence Contract
**Sources:** pack-resolution-architect; zfc-role-neutrality-guardian; doctor-diagnostics-safety-reviewer; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** AC3 asserts deterministic resolution but does not yet include an authoritative matrix for required Core, provider-conditioned `bd` and `dolt`, explicit Gastown, user imports, root/city/rig imports, locks, caches, system packs, local overlays, duplicate names, stale materialized copies, same-named static assets, corrupt caches, and transitive diamond conflicts. Core's trusted resolved-config identity and real-city missing-Core terminal behavior also need to be pinned.
**Required change:** Add or reference a resolution and collision matrix with winners, fail-closed cases, diagnostic fields, Core provenance representation, stale `.gc/system/packs/core` behavior, `bd`/`dolt` anti-shadow rules, cache integrity checks, and ordinary runtime behavior when Core is absent.

### [Major] Test, Scan, And Coverage Gates Are Still Too Narrative
**Sources:** test-strategy-ci-reviewer; zfc-role-neutrality-guardian; external-pack-docs-reviewer; embed-materialization-build-reviewer
**Actionability:** document-fixable
**Issue:** AC8, AC12, AC13, and AC17 name the right verification classes but do not yet force structured artifacts and validators. Reviewers require a strict role-token allowlist, generated/materialized scanning, route/notification target coverage, assertion-level coverage-transfer validation, deterministic public-pack CI fixtures, and a machine-validated acceptance-to-proof matrix.
**Required change:** Define AC17 as a structured artifact with a validator; tighten AC8 token matching and literal-`dog` exceptions; make AC12 path-classified and machine-readable; require AC13 assertion-level replacement or deletion rationale; and split deterministic offline CI from a saved live-network release gate.

### [Major] Attempt-Local Persona Synthesis Artifacts Are Still Misfiled
**Sources:** artifact check; prior attempt syntheses; current attempt timestamps
**Actionability:** workflow-defect
**Issue:** Attempt 5 expected ten persona synthesis files under `.gc/design-reviews/ga-dtvdnd/attempt-5/persona-syntheses/`, but that directory is empty. The ten current syntheses landed under `.gc/design-reviews/ga-dtvdnd/attempt-1/persona-syntheses/`, repeating the attempt-path defect already documented in attempts 2 and 4.
**Required change:** Fix the design-review workflow so persona synthesis output paths and metadata use the active attempt directory, and fail before global synthesis if attempt-local persona syntheses are missing.

### [Minor] Requirements Schema Shape Still Has Document-Fixable Mismatches
**Sources:** requirements-schema-compliance-officer; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The artifact is structurally compliant, but schema conformance is not fully clean for approval: `Open Questions` says `None` while retaining resolved-decision narrative, AC1 is mostly a process guardrail rather than a product outcome, AC6/AC7 contain design-field detail that may belong in support artifacts, and AC15 names a Go symbol as if it were product language.
**Required change:** Reduce `Open Questions` to `None`, relocate resolved-decision history, keep `status: draft` until explicit human approval, reframe AC1 or move it out of product outcomes, and rephrase `PublicGastownPackVersion` as the authoritative public pin/version source unless the symbol itself is intentionally contractual.

## Disagreements
- Requirements-schema, Gastown-behavior, pack-resolution, test-strategy, and ZFC lanes did not block; they mostly judged the requirements direction acceptable but risky. Asset-classification, doctor diagnostics, embed/materialization, external-pack docs, and migration rollout blocked because they found untestable or externally unproven contracts.
- Several disagreements are severity differences, not factual contradictions. Claude and Codex often accepted requirements-level obligations as enough for the next gate, while the third-source reviews more often treated missing commands, schemas, pins, and validators as blockers.
- Some reviewers want exact support-artifact paths and commands in `requirements.md`; others allow the implementation plan to own them. Assessment: the requirements can still pass only if they explicitly say which downstream artifact must define each path/command and that implementation planning cannot be approved until those artifacts exist and validate.
- Compatibility remediation for read-only or nested legacy imports is unsettled. One side favors bounded aliases or dependency overrides, while another warns that aliases become a second source of truth. The next draft must choose a policy or explicitly declare those states unsupported.
- AC8 scanning scope differs by reviewer. The synthesis treats operator-facing Go strings, generated/materialized outputs, routes, notifications, and rendered help/docs as in scope; comments should be scanned when they feed active guidance or generated artifacts, not merely as historical code comments.
- The raw artifact naming is inconsistent: some third-source reviews are stored as `_gemini.md` while self-identifying as DeepSeek V4 Flash, and two attempt-5 raw review files are not present in the active `reviews/` directory. The ten current persona syntheses are complete enough to synthesize, but the workflow artifact handling remains defective.

## Missing Evidence
- Attempt-local persona synthesis files under `.gc/design-reviews/ga-dtvdnd/attempt-5/persona-syntheses/`.
- Concrete AC6 asset ledger path, schema, generator/validator command, source snapshot, closed vocabulary, split representation, behavior IDs or sub-asset keys, and bidirectional validation proof.
- Concrete AC7 behavior-preservation manifest or harness path, baseline denominator, row schema, generator/validator command, failure conditions, and AC6 reconciliation proof.
- Split-compatible public `gascity-packs/gastown` commit, pin ledger, pack digest, behavior-manifest digest, and proof that supported Gastown behavior works without in-tree fallback or implicit Maintenance.
- Enforceable source-tree disposition for `examples/gastown/packs/gastown` and `examples/gastown/packs/maintenance`.
- Version-skew, compatibility-pin, activation-pin, rollback, downgrade, offline repair, and cache-integrity matrices.
- Diagnostic JSON schema, condition codes, source-chain taxonomy, golden text/JSON outputs, recursive import tracer design, and canonical mutating repair command.
- Source-plus-consumer inventory for retired Maintenance, stale Core, in-tree Gastown, Dolt/provider script edges, status-line tracing, doctor/runtime state paths, and mock/test fixture references.
- Role-neutrality scan token list, matching semantics, literal-`dog` configured-default allowlist, diagnostic-data exceptions, and no-executor proof.
- Assertion-level AC13 coverage-transfer table and machine validator.
- Structured AC17 acceptance-to-proof matrix and validator.

## Convergence Assessment
- Remaining blocker class: mixed
- Recommended apply verdict: iterate
- Reason: another edit to `plans/core-gastown-pack-migration/requirements.md` can directly improve several blocking and major issues by pinning acceptance contracts, source-tree disposition, diagnostic/repair policy, AC6/AC7/AC17 obligations, scan semantics, and rollout matrices. However, the review cannot become fully approvable from requirements prose alone: external public-pack artifacts, support validators, implementation-plan updates, and a workflow path fix are also required before a retry should be expected to pass.
- Next non-design work: create or attach the AC6 asset ledger, AC7 behavior manifest/harness, AC17 proof matrix, public Gastown pin/proof artifacts, and source-plus-consumer closure inventory; update `implementation-plan.md` to make those contracts executable; fix the design-review workflow's attempt-local persona synthesis output path.

## Recommended Changes
1. Tighten AC6 into an executable ledger contract with exact support-artifact ownership, validator, schema, split behavior representation, bidirectional validation, collision scanning, Core/non-Gastown preservation, and zero-unresolved-review gate.
2. Add a source-tree disposition requirement proving old `examples/gastown/packs/*` roots are deleted or non-resolvable fixtures and cannot remain a maintained parallel authority.
3. Define the public Gastown compatibility/activation pin artifacts, version-skew matrix, cache/offline behavior, downgrade/rollback expectations, and release ordering.
4. Select one canonical repair command and define typed diagnostics, recursive source attribution, bootstrap diagnose/fix behavior, non-destructive mutation guarantees, interrupted repair recovery, and live-state refusal evidence.
5. Require source-plus-consumer closure for retired Maintenance, stale Core, in-tree Gastown, Dolt/provider script edges, status-line tracing, doctor/runtime state paths, mock registries, fixtures, docs, and tests.
6. Strengthen AC7 with row-level behavior witnesses, a generated supported-Gastown baseline, scripts in public-pack proof, Core/controller-owned trace outcomes, and fail-closed reconciliation with AC6.
7. Add an authoritative pack resolution matrix and Core provenance contract covering required packs, provider packs, imports, locks, caches, overlays, stale materialized copies, duplicate names, same-named static assets, and transitive conflicts.
8. Make AC8, AC12, AC13, and AC17 machine-validated rather than narrative: precise token matching, path-scoped allowlists, generated/materialized scanning, assertion-level coverage transfer, deterministic public-pack CI fixtures, and a validated proof matrix.
9. Clean remaining requirements-schema issues: reduce `Open Questions` to `None`, move resolved-decision history elsewhere, keep `status: draft` until human approval, reframe AC1, and avoid Go-symbol product wording unless intentional.
10. Fix the design-review workflow so current-attempt persona syntheses are written under the active attempt directory before global synthesis runs.
