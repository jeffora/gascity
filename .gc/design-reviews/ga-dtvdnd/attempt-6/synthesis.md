# Design Review Synthesis

## Overall Verdict: block

Attempt 6 is blocked by worst-verdict-wins: the pack-resolution persona returned `block`, while the other nine persona syntheses returned `approve-with-risks`. The requirements artifact now conforms closely to `gc.mayor.requirements.v1` and is much stronger than prior attempts, but one lane still identifies product-level resolution policy that implementation agents cannot safely infer.

The schema review is separate from the behavior/design risks: the document has the required front matter shape, required top-level sections, W6H, Example Mapping, Acceptance Criteria, Out Of Scope, and `Open Questions: None`. Remaining schema concerns are document polish and boundary clarity, not the reason for the global block.

## Consensus Strengths
- Reviewers agree the Core/Gastown/Maintenance product direction is correct: Core is required Gas City SDK runtime support, Maintenance is retired as an active pack, and Gastown behavior is explicit public pack configuration.
- The requirements now name the right support artifacts: pack-resolution matrix, source-consumer closure, asset migration ledger, behavior-preservation manifest or harness, role-neutrality scan, diagnostics schema, public pin/version-skew artifacts, and acceptance proof matrix.
- The migration no longer tries to embed a file-by-file move table in `requirements.md`; file ownership and behavior proof are delegated to validated support artifacts.
- The diagnostic and repair posture is materially stronger: report-only by default, `gc doctor --fix --non-interactive` for mutation, exact source attribution, bootstrap diagnostics, idempotent/atomic/resumable repair, and refusal for read-only, transitive, or live-state hazards.
- ZFC and role neutrality are directionally preserved: default `dog` is acceptable only as configurable pack data, and required deterministic SDK infrastructure remains controller-owned.
- Public Gastown authority is now explicit, including immutable `sha:` pinning, fallback-disabled proof, offline/cache behavior, docs terminology cleanup, and live-network release validation as a separate release gate.

## Critical Findings

### [Blocker] Pack Resolution Contract Still Leaves Critical Policy To Implementation
**Sources:** pack-resolution-architect; embed-materialization-build-reviewer; migration-rollout-reviewer; external-pack-docs-reviewer; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The requirements assert deterministic resolution and required Core, but the blocking persona still finds that the reviewed artifact does not pin one complete product contract for Core identity, Core provenance, resolved-config representation, stale Core behavior, public synthetic alias behavior, duplicate pack names, same-named assets, root/city/rig imports, locks, caches, overlays, provider-conditioned `bd`/`dolt`, and transitive diamond conflicts. If public Gastown imports can still be satisfied by embedded/in-tree aliases or if same-named Core/Gastown assets rely on implicit precedence, the migration can pass fresh-init checks while retaining a hidden second authority.
**Required change:** Add or tighten the Pack Resolution Contract so it names Core's canonical source/provenance, required load layer, lock/cache/materialized representation, normal runtime behavior when Core is missing, stale-copy policy, duplicate/collision behavior, public-alias disposition, provider-pack activation/protection, same-name asset behavior, and transitive pin conflict rule with full diagnostic attribution.

### [Major] Public Gastown Rollout And Legacy Authority Need Release Proof
**Sources:** external-pack-docs-reviewer; migration-rollout-reviewer; gastown-behavior-preservation-auditor; pack-resolution-architect
**Actionability:** external-prerequisite
**Issue:** Reviewers agree the requirements now demand public Gastown authority, but the actual split-compatible public pin, behavior-manifest digest, cache/lock provenance, external-pack proof, and two-repo release sequence are still missing evidence. Existing in-tree examples, synthetic aliases, stale system packs, and old locks can mask a broken public pack unless the release gate proves public Gastown works with in-tree fallback disabled.
**Required change:** Produce or select the split-compatible public Gastown commit and pin ledger, prove behavior from an external checkout or digest-verified cache with fallback disabled, define old/new binary and old/new pack states, and state the post-migration disposition of retained examples or fixtures so no maintained resolvable parallel authority remains.

### [Major] Asset And Behavior Proofs Need Bidirectional Traceability
**Sources:** asset-classification-split-reviewer; gastown-behavior-preservation-auditor; embed-materialization-build-reviewer; test-strategy-ci-reviewer
**Actionability:** external-prerequisite
**Issue:** AC6 and AC7 are the right gates, but the actual ledger/manifest evidence is not present. Reviewers require behavior-row or call-site granularity for split assets, scripts, prompts, template variables, notification targets, requester/detector relationships, identity side effects, warning/failure/escalation paths, recovery flows, and generic non-Gastown behavior that could be swept into Gastown-only rows.
**Required change:** Produce validated AC6/AC7 artifacts, or make the implementation plan require them before implementation approval, with stable behavior/sub-asset IDs, multi-output split representation, Core/Gastown/retired outputs, Core/non-Gastown baseline protection, AC6-to-AC7 reconciliation, and hard failure for orphaned or unresolved `review` rows.

### [Major] Existing-City Repair And Diagnostics Need Concrete Outcome Semantics
**Sources:** doctor-diagnostics-safety-reviewer; migration-rollout-reviewer; pack-resolution-architect; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The requirements cover the right diagnostic surfaces, but reviewers still need deterministic exit codes, bootstrap-only diagnose/fix boundaries, recursive import-chain provenance, post-fix verification fields, stale directory guidance, read-only/transitive handling, live-state refusal evidence, post-marker old-binary recovery steps, and air-gapped cache-seeding guidance.
**Required change:** Define diagnostic and repair outcome semantics in the requirements or implementation plan: exit-code matrix, text/JSON condition schema, recursive source-chain representation, bootstrap mutation policy, interrupted repair recovery, post-fix fields, manual reconciliation for old-binary writes, offline cache pre-seeding path, and live process/session evidence that avoids stale PID/status files.

### [Major] Role-Neutral Executor And Scan Controls Need Exact Boundaries
**Sources:** zfc-role-neutrality-guardian; test-strategy-ci-reviewer; requirements-schema-compliance-officer
**Actionability:** document-fixable
**Issue:** The artifact is ZFC-aligned in direction, but AC8/AC9 still need exact token, binding, and allowlist semantics. Literal `dog` must not become a route, notification, prompt, formula, or Go fallback hidden behind the configured-default carve-out, and diagnostic references to retired identifiers must not weaken role-neutrality scans.
**Required change:** State that literal `dog` and other Gastown role names may appear only in exact configured-default data or narrowly scoped diagnostic/migration allowlists; require symbolic bindings in Core assets; define required versus optional bindings in data; include init/template resolution, route targets, notification targets, generated/materialized outputs, rendered templates, and diagnostics in scan coverage.

### [Major] Test And CI Gates Must Prevent Vacuous Passes
**Sources:** test-strategy-ci-reviewer; asset-classification-split-reviewer; gastown-behavior-preservation-auditor; external-pack-docs-reviewer; embed-materialization-build-reviewer
**Actionability:** document-fixable
**Issue:** Several validators can become self-defeating if they walk only the post-deletion workspace or accept weak traceability. Reviewers repeatedly call out the risk that deleted legacy directories yield empty-pass completeness, AC13 maps retired assertions to no-op tests, public-pack CI uses synthetic resolver bypasses, and cache promotion tests miss concurrent staging races.
**Required change:** Require AC6/AC13 completeness against a frozen historical baseline or digest-verified snapshot, assertion- or behavior-level coverage transfer, centralized denied-token rot guards, basename/static-fragment collision scanning, deterministic offline tests that exercise the real resolver/cache/digest path, and randomized/process-unique cache staging in atomic promotion tests.

### [Major] Attempt-Local Persona Synthesis Artifacts Are Still Misfiled
**Sources:** artifact check; fanout metadata; current attempt timestamps
**Actionability:** workflow-defect
**Issue:** `.gc/design-reviews/ga-dtvdnd/attempt-6/persona-syntheses/` is empty even though ten current-timestamp persona syntheses exist under `.gc/design-reviews/ga-dtvdnd/attempt-1/persona-syntheses/`. The raw attempt-6 reviews also describe a known attempt-path defect and dual-writing workaround. I used the ten current syntheses from the misplaced path to avoid losing the review result, but the current attempt directory does not contain the required persona synthesis artifacts.
**Required change:** Fix the design-review workflow so persona synthesis output paths and metadata use the active attempt directory, and fail before global synthesis when attempt-local persona syntheses are missing or stale.

### [Minor] Requirements Schema And Product/Design Boundary Need Polish
**Sources:** requirements-schema-compliance-officer; doctor-diagnostics-safety-reviewer; asset-classification-split-reviewer; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The requirements document is schema-shaped and implementation-plan-ready in broad form, but several ACs sit close to the product/design boundary. AC1 is mainly a process gate, AC6/AC7 include field-level artifact detail, AC10/AC16 include mechanism vocabulary, and `status: draft` correctly remains until explicit human approval.
**Required change:** Before approval, keep observable product outcomes in `requirements.md`, move concrete data models and command internals into `implementation-plan.md` or support artifact schemas, preserve `status: draft` until approval, and make AC1 a process note if the product acceptance list should contain only externally visible outcomes.

## Disagreements
- Only the pack-resolution persona blocks. The other personas approve with risks because they consider the current requirements direction good enough if downstream artifacts and the implementation plan harden the details.
- Some reviewers treat AC6/AC7/AC14/AC17 as sufficient requirements-level obligations; others warn that missing commands, schemas, baselines, and validators can still let implementation invent unsafe policy. Assessment: the requirements may remain compact, but the next gate must make those artifacts executable and validated.
- DeepSeek/Gemini reviews often reference the implementation plan and conclude that several risks are already designed. Claude and Codex syntheses more often judge only the reviewed requirements artifact. Assessment: implementation-plan improvements reduce risk, but the source requirements and approval gates must still make load-bearing policies visible.
- Reviewers differ on compatibility shims for nested read-only legacy imports. Assessment: choose a bounded override/remediation policy or explicitly declare those states unsupported with actionable diagnostics; do not allow silent fallback.
- There is some severity disagreement on schema details. Assessment: schema conformance is mostly satisfied; the global block is about resolver authority and migration proof, not missing required sections.

## Missing Evidence
- Authoritative Pack Resolution Contract covering canonical Core identity, provenance, resolved-config representation, stale/materialized Core behavior, public alias disposition, provider pack rules, duplicate/collision behavior, same-named assets, lock/cache precedence, and diamond conflicts.
- Split-compatible public `gascity-packs/gastown` commit, immutable pin ledger, pack digest, behavior-manifest digest, cache/lock provenance, and fallback-disabled validation result.
- Validated AC6 asset migration ledger with source snapshot, closed vocabulary, split representation, behavior/sub-asset IDs, bidirectional validation, basename collision checks, and zero-unresolved-review gate.
- Validated AC7 behavior-preservation manifest or harness with an authoritative before-state denominator and runtime witnesses for triggers, rendering, routing, notifications, scripts, identity effects, failure/escalation, and recovery paths.
- Source-consumer closure for retired Maintenance, stale Core, in-tree Gastown, Dolt/provider scripts, status-line tracing, doctor/runtime-state paths, mock registries, fixtures, docs, tests, and Go embed/registry consumers.
- Diagnostic text/JSON schema, exit-code matrix, recursive import provenance, bootstrap-only repair semantics, post-fix verification fields, stale directory guidance, live-state refusal evidence, and old-binary reconciliation steps.
- Role-neutrality denied-token source, exact allowlist keys/paths, generated/materialized scan coverage, rendered-template controls, symbolic binding proof, no-executor proof, and data-declared binding optionality.
- AC13 coverage-transfer validator proving retired assertions map to behavior-equivalent replacement coverage or approved deletion rationale.
- AC17 acceptance-proof matrix with concrete commands, fixtures, artifact validators, local/integration/release/manual gate placement, and fail-closed handling for missing external evidence.
- Attempt-local persona synthesis files under `.gc/design-reviews/ga-dtvdnd/attempt-6/persona-syntheses/`.

## Convergence Assessment
- Remaining blocker class: mixed
- Recommended apply verdict: iterate
- Reason: at least one remaining blocker is document-fixable: the pack-resolution contract can be made explicit in `plans/core-gastown-pack-migration/requirements.md` or in a gated implementation-plan/support-artifact requirement. Another design-doc iteration can move the review forward, but approval will still require external artifacts and a workflow path fix before the next run can be expected to pass cleanly.
- Next non-design work: create and validate the AC6 ledger, AC7 behavior manifest/harness, AC17 proof matrix, public Gastown pin/proof artifacts, and source-consumer closure; fix the design-review workflow's active-attempt persona synthesis path; use `<none>` only after those artifacts exist or are deliberately declared out of scope.

## Recommended Changes
1. Add a complete Pack Resolution Contract that settles Core identity, provenance, precedence, alias retirement, provider-pack rules, duplicate/same-name behavior, lock/cache/system-pack ordering, stale-copy behavior, and transitive pin conflict diagnostics.
2. Produce or gate on the split-compatible public Gastown pin and fallback-disabled behavior proof, with a two-repo release-ordering and version-skew matrix.
3. Produce and validate AC6, AC7, AC5, and AC17 support artifacts before implementation approval; make missing, stale, or internally inconsistent artifacts block the gate.
4. Strengthen AC6/AC7 traceability with frozen historical baselines, behavior/sub-asset IDs, multi-output split rows, Core/non-Gastown preservation, bidirectional behavior witnesses, and collision scanning.
5. Define existing-city repair and diagnostic semantics: exit codes, recursive source attribution, bootstrap-only mode, interrupted recovery, post-fix fields, stale-state guidance, live-state refusal, old-binary reconciliation, and offline cache seeding.
6. Tighten role-neutrality controls around literal `dog`, diagnostic retired identifiers, symbolic bindings, data-declared optionality, init/template resolution, route/notification targets, rendered templates, and generated/materialized outputs.
7. Make test gates fail closed: AC13 assertion-level coverage transfer, denied-token rot guard, frozen-snapshot completeness, real resolver/cache/digest offline tests, and randomized cache staging concurrency tests.
8. State the post-migration disposition of retained examples and fixtures so no old `examples/gastown/packs/*` or `.gc/system/packs/*` path remains a maintained resolvable authority.
9. Polish schema-boundary items before approval: keep `status: draft` until human approval, move implementation mechanics into design/support schemas, and reframe AC1 if needed.
10. Fix the review workflow so current-attempt persona syntheses are written under `.gc/design-reviews/ga-dtvdnd/attempt-<n>/persona-syntheses/` before global synthesis runs.
