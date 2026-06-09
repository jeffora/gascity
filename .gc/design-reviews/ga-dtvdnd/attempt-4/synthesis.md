# Design Review Synthesis

## Overall Verdict: block

Attempt 4 remains blocked. The requirements artifact is now structurally valid as a `gc.mayor.requirements.v1` draft in `status: questions`, but the persona syntheses identify unresolved product decisions, missing proof artifacts, and release/diagnostic contracts that would force implementation agents to invent policy.

The current review workflow also still has an artifact-path defect: all ten persona synthesis scopes closed for attempt 4, but their synthesis files were written under `.gc/design-reviews/ga-dtvdnd/attempt-1/persona-syntheses/` while `.gc/design-reviews/ga-dtvdnd/attempt-4/persona-syntheses/` is empty. I used the ten current-timestamp persona syntheses from the misfiled path, plus the available attempt-4 raw reviews, and treat the path problem as a workflow defect rather than design evidence.

## Consensus Strengths
- The requirements now use the required schema shape: front matter, `Problem Statement`, `W6H`, `Example Mapping`, `Acceptance Criteria`, `Out Of Scope`, and `Open Questions`.
- Reviewers consistently praised the core product direction: Core becomes the only required Gas City-owned runtime pack, Maintenance is retired as an active standalone pack, and Gastown behavior is loaded explicitly from the public Gastown pack.
- AC6 and AC7 are the right anchors for preserving safety: an external asset migration ledger and behavior-preservation proof are needed instead of an inline file-by-file table in `requirements.md`.
- AC8 and AC9 make role neutrality explicit, including absence scanning and treating the default `dog` executor as configurable pack data rather than Go-side role knowledge.
- The diagnostics posture is directionally sound: report-only by default, no silent fallback, exact source attribution, structured JSON, and explicit non-interactive repair.

## Critical Findings

### [Blocker] Requirements Are Not Implementation-Ready While Open Questions Remain
**Sources:** requirements-schema-compliance-officer; doctor-diagnostics-safety-reviewer; pack-resolution-architect; migration-rollout-reviewer; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** The artifact is schema-valid as `status: questions`, but `Open Questions` still contains five material product decisions: ledger ownership, behavior-proof command/location, Gas City/public-Gastown version skew, in-flight retired-path sessions, and the exact repair workflow. The schema requires `Open Questions: None` for implementation-plan readiness.
**Required change:** Resolve the five product decisions in `plans/core-gastown-pack-migration/requirements.md`, update the relevant acceptance criteria and examples, and keep the artifact in `questions` or `blocked` until those answers are explicit.

### [Blocker] Core Resolution And Precedence Contract Is Undefined
**Sources:** pack-resolution-architect; embed-materialization-build-reviewer; migration-rollout-reviewer; doctor-diagnostics-safety-reviewer
**Actionability:** document-fixable
**Issue:** The requirements assert required Core and deterministic resolution but do not define Core's canonical source, representation in resolved config, provenance, materialization/cache behavior, runtime behavior when Core is missing, or precedence/collision rules across Core, `bd`, `dolt`, explicit Gastown, user imports, locks, caches, system packs, overlays, and same-named split assets.
**Required change:** Add a Core resolution contract and precedence table covering init, runtime load, doctor/import-state, import install, lock/cache, synthetic materialization, pack discovery, duplicate Core, required-pack shadowing, and stale materialized copies.

### [Blocker] AC6 Asset Ledger Cannot Yet Prove Safe File Ownership
**Sources:** asset-classification-split-reviewer; embed-materialization-build-reviewer; gastown-behavior-preservation-auditor; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** AC6 names an external ledger but does not pin the artifact path or generator, schema, owner, validator command, source snapshot, closed classification vocabulary, multi-output split representation, bidirectional validation, basename-collision handling, or gate that prevents unresolved `review` rows from proceeding.
**Required change:** Tighten AC6 so the ledger is generated or checked at a named path, validated against a deterministic source snapshot, covers both sources and downstream consumers, requires all split outputs, rejects stale/phantom rows, and blocks implementation while unresolved classifications remain.

### [Blocker] Behavior Preservation Proof Is Still Too Abstract
**Sources:** gastown-behavior-preservation-auditor; test-strategy-ci-reviewer; external-pack-docs-reviewer; migration-rollout-reviewer
**Actionability:** document-fixable
**Issue:** AC7 requires a manifest or harness, but the requirements do not define the authoritative Gastown behavior baseline, row schema, command, output path, owner, failure conditions, or CI/release gate. Reviewers specifically called out formulas, orders, scripts, prompts, template variables, notifications, requester/detector relationships, warning/failure/escalation paths, recovery flows, and identity side effects.
**Required change:** Define the AC7 proof contract and require executable behavior checks that load, resolve, render, trigger, route, notify, and exercise failure/recovery paths from the external public Gastown pack without in-tree fallback.

### [Blocker] Public Gastown And Two-Repo Rollout Are Not Proven
**Sources:** external-pack-docs-reviewer; migration-rollout-reviewer; pack-resolution-architect; embed-materialization-build-reviewer
**Actionability:** external-prerequisite
**Issue:** The requirements name a public Gastown pin while retiring implicit Maintenance behavior, but reviewers found no proof that the pinned public pack is split-compatible or self-sufficient for supported Gastown templates. Release ordering, version skew, offline cache-hit/cache-miss behavior, fresh offline init, mutable refs, failed fetch/install, and fallback policy remain unresolved.
**Required change:** Produce or select a split-compatible public Gastown commit, define the Gas City/public-Gastown version-skew policy, validate the public pack from an external checkout or pinned cache with in-tree fallback disabled, and document the two-repo release and rollback behavior.

### [Blocker] Existing-City Migration And Repair Semantics Are Underspecified
**Sources:** migration-rollout-reviewer; doctor-diagnostics-safety-reviewer; pack-resolution-architect; embed-materialization-build-reviewer
**Actionability:** document-fixable
**Issue:** Existing-city repair lacks a complete before/after matrix across root imports, `city.toml`, default-rig imports, rig-local imports, relative legacy paths, `.gc/system/packs/*`, lockfiles, cached public packs, stale local pack directories, read-only third-party imports, and in-flight sessions. The exact repair command and mutation boundaries remain open.
**Required change:** Define the migration matrix and repair command semantics, including what mutates, what refuses, atomicity, rollback/resume after install failure, read-only handling, stale directory pruning guidance, live process/session gating, and post-repair verification.

### [Major] Diagnostic Contracts Need Stable Text, JSON, And Bootstrap Behavior
**Sources:** doctor-diagnostics-safety-reviewer; pack-resolution-architect; migration-rollout-reviewer; test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** AC10 and AC11 require exact source attribution, but the requirements do not yet define stable diagnostic codes, severity, source-reference format, nested import-chain representation, affected path, repair-action fields, mutation policy, idempotency status, or golden examples. Doctor/import-state also need a bootstrap-only diagnostic load path when normal pack resolution is broken.
**Required change:** Add diagnostic text/JSON contracts for missing Core and retired imports, recursive import-chain attribution, bootstrap-only diagnostic mode, no-default-substitution behavior, and golden tests that avoid historical closed-bead diagnostic pollution.

### [Major] Role-Neutral Maintenance Executor Boundary Is Too Loose
**Sources:** zfc-role-neutrality-guardian; test-strategy-ci-reviewer; doctor-diagnostics-safety-reviewer
**Actionability:** document-fixable
**Issue:** AC8 and AC9 prohibit hardcoded role logic but still need an exact boundary for the default `dog` configured data. Reviewers need a taxonomy of Core maintenance work proving required deterministic SDK infrastructure stays controller-owned and optional executor/LLM work remains configurable, diagnosable, and symbolically routed.
**Required change:** Define Core maintenance work categories, forbid Go-side `dog` fallback, restrict literal `dog` to exact configured-default fields, require symbolic bindings in Core formulas/orders/prompts/metadata, and add no-executor and renamed-executor verification cases.

### [Major] Test And CI Gates Are Not Concrete Enough
**Sources:** test-strategy-ci-reviewer; asset-classification-split-reviewer; gastown-behavior-preservation-auditor; external-pack-docs-reviewer
**Actionability:** document-fixable
**Issue:** AC1-AC14 are not mapped to concrete proof commands, fixture types, artifact paths, and gate placement. AC13 lacks a strict coverage-transfer failure contract, AC14 risks nondeterministic public-network validation, and AC8 needs scan coverage over source plus generated/materialized surfaces with positive and negative controls.
**Required change:** Add an acceptance-to-proof matrix, define AC13 completeness validation for retired tests/assertions, make normal public-pack CI deterministic through local fixtures or pinned caches, reserve live public checks for a named release gate, and require absence-scan controls over both source and materialized outputs.

### [Major] Attempt-Local Persona Synthesis Artifacts Are Misfiled
**Sources:** attempt-4 artifact check; fanout metadata; current persona synthesis timestamps
**Actionability:** workflow-defect
**Issue:** Attempt 4 expected ten persona synthesis files under `.gc/design-reviews/ga-dtvdnd/attempt-4/persona-syntheses/`, but that directory is empty. The ten current persona synthesis files landed under `.gc/design-reviews/ga-dtvdnd/attempt-1/persona-syntheses/`, repeating the attempt-path defect already described in the attempt-2 synthesis.
**Required change:** Fix the review workflow so persona synthesis output paths and metadata use the current attempt directory, and fail before global synthesis if attempt-local persona synthesis files are missing.

### [Minor] Requirements Artifact Contains Review Provenance Residue
**Sources:** requirements-schema-compliance-officer; embed-materialization-build-reviewer; external-pack-docs-reviewer; asset-classification-split-reviewer
**Actionability:** document-fixable
**Issue:** `requirements.md` still contains inline `<!-- REVIEW: ... -->` comments. Those are workflow provenance, not product requirements, and the schema says research notes and review reports belong in workflow artifacts.
**Required change:** Remove the review comments from `plans/core-gastown-pack-migration/requirements.md` and keep review provenance under `.gc/design-reviews/ga-dtvdnd/...`.

## Disagreements
- Several personas split on severity rather than substance. Claude often returned `approve-with-risks` while Codex or DeepSeek/Gemini returned `block`; the synthesis follows the stricter verdict when the issue defines product policy that implementation cannot safely infer.
- The zfc-role-neutrality and Gastown-behavior personas landed at `approve-with-risks`, but their major gaps are prerequisites for other blocking lanes, especially AC7, AC8, AC9, and in-flight-session migration behavior.
- Some DeepSeek/Gemini raw reviews referenced a later implementation plan and proposed concrete mechanisms such as symbolic bindings, local Git fixtures, AST-preserving TOML repair, and `gc doctor --fix`. I treat those as viable candidates, not as proof that the reviewed requirements artifact already contains the decisions.
- Reviewers disagreed on compatibility shims: some warn against any legacy alias becoming a second source of truth, while others allow narrowly scoped aliases for pinned/read-only remote packages. The requirements must make an explicit product decision either way.
- Reviewers also differed on offline Gastown behavior: bundled fallback, pinned immutable cache only, and fail-closed fresh offline init were all proposed. The next draft must choose one policy.

## Missing Evidence
- Final answers to all five open product questions in `requirements.md`.
- Concrete AC6 asset migration ledger path or generator output, closed schema, owner, validator command, deterministic source snapshot, and bidirectional validation proof.
- Concrete AC7 behavior-preservation manifest/harness command, output path, schema, owner, failure conditions, and CI/release gate.
- A split-compatible public `gascity-packs/gastown` commit and release artifact proving supported Gastown behavior without in-tree or implicit Maintenance fallback.
- Core resolution and precedence matrix covering init, runtime, doctor/import-state, locks, cache, materialization, imports, overlays, collisions, and stale system packs.
- Existing-city upgrade matrix with diagnostics and repair behavior for every legacy import/cache/system-pack shape.
- Stable text and JSON diagnostics, including nested import-chain source references and bootstrap-only diagnostic behavior.
- No-executor and renamed-executor proof for SDK self-sufficiency and configured maintenance executor routing.
- AC1-AC14 proof matrix naming commands, tests, fixtures, artifacts, and gate placement.
- Attempt-local persona synthesis artifacts under `.gc/design-reviews/ga-dtvdnd/attempt-4/persona-syntheses/`.

## Convergence Assessment
- Remaining blocker class: mixed
- Recommended apply verdict: iterate
- Reason: another requirements-document edit can resolve several blockers directly by making product decisions explicit, closing `Open Questions`, tightening AC6/AC7/AC10/AC11/AC13/AC14, and removing schema residue. Some blockers also require external artifacts and public-pack proof, but the reviewed document must first state the selected policies and gates.
- Next non-design work: create or attach the AC6 asset migration ledger, create or attach the AC7 behavior-preservation harness/manifest, validate a split-compatible public Gastown pack from an external checkout or pinned cache, and fix the review workflow's attempt-local persona synthesis output path.

## Recommended Changes
1. Resolve the five `Open Questions` and revise `status` only when the artifact is ready; otherwise keep it as `questions` or `blocked`.
2. Add a Core resolution, provenance, precedence, collision, and bootstrap-diagnostics contract covering all resolver surfaces.
3. Tighten AC6 into an executable ledger contract with path/generator, schema, owner, validator, deterministic snapshot, multi-output split rows, bidirectional validation, collision scanning, and hard unresolved-review gates.
4. Tighten AC7 into an executable behavior-preservation contract with supported-behavior baseline, row schema, command, output path, owner, failure conditions, and external-pack runtime proof.
5. Define the public Gastown rollout and version-skew policy, then validate against a split-compatible public pack with local in-tree fallback disabled.
6. Add the existing-city migration matrix and exact repair-command semantics, including atomicity, refusal cases, stale state, read-only/nested imports, and in-flight sessions.
7. Add stable diagnostic text/JSON contracts and golden tests for missing Core, retired imports, duplicate Core, version skew, nested import chains, and repair output.
8. Define the configured maintenance executor boundary, including no Go-side `dog` fallback, symbolic routing, no-executor behavior, and exact absence-scan exceptions.
9. Add the AC1-AC14 proof matrix and deterministic CI/release-gate placement, including AC13 coverage-transfer completeness and AC14 public-pack fixture strategy.
10. Remove inline `<!-- REVIEW: ... -->` comments from `requirements.md`.
11. Fix the design-review workflow so attempt-local persona synthesis files are written under the active attempt directory before global synthesis runs.
