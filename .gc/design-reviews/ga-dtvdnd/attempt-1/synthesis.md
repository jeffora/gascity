# Design Review Synthesis

## Overall Verdict: block

All ten persona syntheses return `block`, so the global verdict is `block` under worst-verdict-wins. The requirements describe the right migration direction, but the artifact is not schema-compliant and leaves several product contracts unresolved: Core loading, public Gastown resolution, Maintenance retirement, behavior preservation, role neutrality, diagnostics, rollout, and proof coverage.

## Consensus Strengths
- Multiple personas agreed on the high-level target state: Core should be the only required Gas City-owned runtime pack, Maintenance should retire as a standalone pack, and Gastown behavior should move to the external `gascity-packs/gastown` pack.
- The artifact already states several important non-goals, including no new MEOW primitives, no Gastown role-model redesign, and no source comments preserving removed pack history.
- Reviewers recognized useful starting decisions such as `internal/packs/core` as the intended canonical Core source path, the public Gastown pack as the intended Gastown authority, and explicit preservation of Gastown behavior as a migration goal.

## Critical Findings

### [Blocker] Requirements artifact does not conform to the Mayor requirements schema
**Sources:** Mara Voss / requirements-schema-compliance-officer; Simone Kaye / external-pack-docs-reviewer; Camille Okafor / migration-rollout-reviewer; Priya Menon / pack-resolution-architect; Alistair Sterling / zfc-role-neutrality-guardian; Hugo Bautista / asset-classification-split-reviewer
**Actionability:** document-fixable
**Issue:** The artifact is marked `status: approved` but does not follow `gc.mayor.requirements.v1`. It lacks the required top-level section order, including `## W6H`, `## Example Mapping`, consolidated `## Acceptance Criteria`, and `## Open Questions`, and it mixes implementation-plan material such as file migration tables and downstream file assignments into requirements.
**Required change:** Rewrite `plans/core-gastown-pack-migration/requirements.md` to the required schema, move implementation-plan details out of the normative requirements body, change status away from `approved`, add W6H and example mapping with concrete happy/negative/edge scenarios, consolidate behavior-focused acceptance criteria, and use `Open Questions` for unresolved product decisions or `None` only when they are resolved.

### [Blocker] Core loading, pack resolution, and provider behavior are undefined
**Sources:** Priya Menon / pack-resolution-architect; Petra Novak / embed-materialization-build-reviewer; Faisal Khoury / doctor-diagnostics-safety-reviewer; Camille Okafor / migration-rollout-reviewer
**Actionability:** document-fixable
**Issue:** The requirements say Core is required for `gc` to run while also requiring doctor/import-state diagnostics for missing Core, but they do not define one coherent model for init, doctor, CLI load, runtime resolution, resolved-config provenance, and diagnostic no-fatal behavior. Deterministic precedence across Core, public Gastown, `bd`, `dolt`, root packs, city imports, default rig imports, locked remotes, system includes, and local overlays is also missing.
**Required change:** Define the Core loading model in product terms, including how Core is represented in resolved config, how doctor can diagnose absence without failing before it reports, how dev/test escape hatches work if any, and the deterministic resolution, shadowing, collision, and provider-conditioned `bd`/`dolt` rules.

### [Blocker] Public Gastown authority and no-fallback behavior are not specified
**Sources:** Simone Kaye / external-pack-docs-reviewer; Priya Menon / pack-resolution-architect; Petra Novak / embed-materialization-build-reviewer; Camille Okafor / migration-rollout-reviewer
**Actionability:** document-fixable
**Issue:** The requirements name `gascity-packs/gastown` but do not define the exact source syntax, registry alias, URL/ref/version or pinning model, lock/cache/mirror behavior, offline behavior, or proof that `gc init --template gastown` resolves through public/external authority rather than local `examples/gastown/packs/gastown` or `.gc/system/packs/gastown` fallback.
**Required change:** Specify the canonical external Gastown import contract, version and cache policy, offline or fail-fast behavior, and an acceptance criterion proving `gc init --template gastown` uses the public/external source with no silent in-tree or system-pack fallback.

### [Blocker] Asset migration inventory is not reproducible or faithful enough to drive implementation
**Sources:** Hugo Bautista / asset-classification-split-reviewer; Oleg Marchetti / gastown-behavior-preservation-auditor; Petra Novak / embed-materialization-build-reviewer; Mira Acharya / test-strategy-ci-reviewer
**Actionability:** external-prerequisite
**Issue:** Persona reviewers found stale or nonexistent migration-table rows, unresolved `review` rows, conditional `core` rows without fallbacks, duplicated template fragments without reconciliation policy, and split/core-renamed assets that do not consistently name both resulting artifacts, exact target paths, stripped behavior, retained behavior, and proof. A requirements document cannot safely approve a file-by-file migration map without a generated or validated source-tree inventory.
**Required change:** Produce or require a validated asset migration ledger with current path, provenance, target owner, concrete target path(s), action, rationale, split boundary, fallback classification, and proof command. The ledger or validation must fail on missing current paths, unrepresented source files, unmarked historical rows, and unresolved `review` assets.

### [Blocker] Behavior preservation is asserted but not proven
**Sources:** Oleg Marchetti / gastown-behavior-preservation-auditor; Mira Acharya / test-strategy-ci-reviewer; Hugo Bautista / asset-classification-split-reviewer; Simone Kaye / external-pack-docs-reviewer
**Actionability:** external-prerequisite
**Issue:** The requirements promise that stripped Gastown behavior survives, but do not require behavior-level successor rows or executable checks for formulas, orders, scripts, prompts, template variables, notification paths, requester/detector relationships, and recovery flows. Dog notification and escalation paths are especially under-mapped: the current wording can remove Mayor/Gastown requesters, deacon `DOG_DONE` nudges, daemon identity, and escalation behavior from Core without naming the external Gastown successor.
**Required change:** Add a behavior-preservation manifest or harness that records current behavior, Core-after behavior, Gastown-after behavior, trigger/order/formula identity, script entrypoint, template variables, notification target, and proof command. Require runtime checks that supported Gastown workflows load, resolve, render, trigger, and deliver success/warning/failure/escalation notifications through the external pack.

### [Blocker] Role neutrality and the `dog` maintenance binding are underspecified
**Sources:** Alistair Sterling / zfc-role-neutrality-guardian; Mira Acharya / test-strategy-ci-reviewer; Oleg Marchetti / gastown-behavior-preservation-auditor
**Actionability:** document-fixable
**Issue:** The requirements do not make the controller-vs-agent boundary explicit enough. Core runtime behavior could still depend on a configured maintenance agent literally named `dog`, and Core formulas, orders, routing metadata, pools, dispatch targets, prompts, provider overlays, or generated metadata could preserve role assumptions outside Go source.
**Required change:** State that SDK infrastructure works with only the controller and no configured agent role. Define the configurable maintenance-worker contract for renamed, replaced, omitted, or disabled executors; require symbolic bindings for Core maintenance work; and prohibit literal Gastown role names or literal `dog` routing except in narrow configured-default or absence-test exceptions.

### [Blocker] Acceptance criteria are not traceable to proof
**Sources:** Mira Acharya / test-strategy-ci-reviewer; Petra Novak / embed-materialization-build-reviewer; Oleg Marchetti / gastown-behavior-preservation-auditor; Faisal Khoury / doctor-diagnostics-safety-reviewer
**Actionability:** document-fixable
**Issue:** The artifact lists important outcomes, but it does not map each acceptance criterion to a named unit test, integration test, command verification, absence scan, behavior harness, golden output, or manual check. Legacy Maintenance/Gastown test retirement is also not paired with a coverage-transfer table.
**Required change:** Add an acceptance traceability matrix that maps every AC to a verification method and, where appropriate, target test path or command. Include pack-loading/import-resolution matrices, role-neutrality absence scans with controls, behavior-preservation checks, documentation/diagnostics audits, and a replacement-coverage table for retired tests.

### [Major] Existing-city migration, compatibility shims, and rollout sequencing are unresolved
**Sources:** Camille Okafor / migration-rollout-reviewer; Priya Menon / pack-resolution-architect; Petra Novak / embed-materialization-build-reviewer; Faisal Khoury / doctor-diagnostics-safety-reviewer
**Actionability:** document-fixable
**Issue:** Existing cities that reference `packs/maintenance`, `packs/gastown`, in-tree examples, stale `.gc/system/packs/*` entries, old synthetic caches, pinned/cached public Gastown versions, or custom local overlays do not have exact before/after states. Compatibility-shim behavior is left open, and release sequencing across Gas City and `gascity-packs/gastown` is not specified.
**Required change:** Add an upgrade matrix covering legacy local imports, public pins/caches, stale generated state, custom local packs, rollback/downgrade expectations, in-flight work, and runtime state. Define whether repair is report-only, patch generation, direct mutation, or separate command; require idempotence, source preservation, proof after repair, and a bounded shim policy or an explicit breaking migration.

### [Major] Doctor diagnostics and remediation safety are too vague
**Sources:** Faisal Khoury / doctor-diagnostics-safety-reviewer; Priya Menon / pack-resolution-architect; Camille Okafor / migration-rollout-reviewer; Simone Kaye / external-pack-docs-reviewer
**Actionability:** document-fixable
**Issue:** The requirements require doctor/import-state warnings and a fix offer, but do not define resolved config source attribution, nested import tracing, output channels, JSON/stdout isolation, no-prompt behavior, idempotence, read-only config handling, duplicate-Core prevention, or post-fix verification.
**Required change:** Define diagnostic cases with condition, source attribution, severity, output channel, exact operator-facing message shape, remediation mode, and expected post-fix result. Require non-interactive behavior and golden-output or equivalent checks across doctor, import-state, CLI text, docs, and automation modes.

### [Major] Builtin materialization, stale state, and downstream references need explicit end-state contracts
**Sources:** Petra Novak / embed-materialization-build-reviewer; Priya Menon / pack-resolution-architect; Simone Kaye / external-pack-docs-reviewer
**Actionability:** document-fixable
**Issue:** The requirements do not define final behavior for `builtinpacks.All()`, source aliases, synthetic repo validation/materialization, required builtin lists, provider-dependent required packs, `.gc/system/packs/maintenance`, old synthetic cache layouts, generated formulas/orders, scripts, docs, and runtime state under retired pack namespaces.
**Required change:** Add materialization and downstream-reference acceptance criteria that fail if Maintenance remains bundled, public-source recognized, auto-included, or materialized as an active system pack. Define prune/ignore/regenerate/diagnose behavior for stale artifacts and require absence scans for non-historical references to retired paths.

### [Major] Documentation, examples, and terminology cleanup are not testable
**Sources:** Simone Kaye / external-pack-docs-reviewer; Faisal Khoury / doctor-diagnostics-safety-reviewer; Mira Acharya / test-strategy-ci-reviewer
**Actionability:** document-fixable
**Issue:** Documentation cleanup is described generically and focuses mainly on Maintenance imports, but the migration also needs to retire authoritative `packs/gastown`, `.gc/system/packs/gastown`, `.gc/system/packs/maintenance`, and `examples/gastown/packs/*` references. The artifact does not distinguish the retired Maintenance pack from legitimate store/supervisor maintenance terminology.
**Required change:** Add a docs/examples/help inventory with allowed archive or migration-history exceptions, a terminology matrix for Core required / Gastown external / Maintenance pack retired / store maintenance separate, and a verification command or audit artifact that prevents stale authority-path guidance from returning.

### [Minor] Some naming and status choices reinforce legacy role assumptions
**Sources:** Alistair Sterling / zfc-role-neutrality-guardian; Mara Voss / requirements-schema-compliance-officer
**Actionability:** document-fixable
**Issue:** Names such as `mol-dog-jsonl`, `mol-dog-reaper`, and similar Core maintenance assets reinforce `dog` as an operation identity rather than a configurable executor. The `approved` front matter status also misrepresents readiness.
**Required change:** Prefer function-centric names for Core maintenance formulas and orders where the implementation plan rewrites assets, and ensure the requirements status reflects draft/questions/blocked until schema and product gaps are resolved.

## Disagreements
- Several individual source reviewers gave `approve-with-risks` within specific persona lanes, but every persona synthesis ultimately assessed its lane as `block`. The global synthesis follows the persona syntheses, not the most optimistic model review inside a lane.
- Reviewers differed on whether schema nonconformance is a process issue or a substantive blocker. I assess it as substantive because the missing W6H, Example Mapping, Acceptance Criteria, and Open Questions sections are exactly where Core absence, public Gastown resolution, migration rollout, and role-neutrality scenarios need to become concrete and testable.
- Reviewers differed on whether the asset migration table should remain in requirements. My assessment is that requirements may state the need for a validated ledger and summarize product outcomes, but file-by-file migration assignments, proof commands, and generated inventories should live in a supporting design/implementation artifact unless the schema owner explicitly permits them in requirements.
- Reviewers did not fully agree on the offline Gastown behavior: some suggested cached or embedded fallback, while others only required an explicit product decision. The synthesis does not prescribe the answer; it requires the requirements to choose and test one behavior.
- Reviewers varied on whether compatibility shims should redirect, rewrite, warn, or be rejected. The synthesis treats this as an unresolved product contract: the requirements must define each supported legacy surface and must not leave a permanent alternate source of truth by accident.

## Missing Evidence
- Schema-compliant W6H, Example Mapping, consolidated Acceptance Criteria, and Open Questions for the Core/Gastown/Maintenance migration.
- A validated source-tree inventory or ledger proving every listed path exists or is explicitly historical and every relevant source file is represented exactly once.
- Exact Core loading, resolved-config provenance, pack precedence, collision, provider-conditioned `bd`/`dolt`, and doctor no-fatal diagnostic contracts.
- The canonical public Gastown source string, version/ref/pin policy, lock/cache/mirror/offline behavior, and no-in-tree-fallback proof.
- Behavior-level successor rows for stripped or moved Gastown behavior, especially Dog notification, requester, escalation, daemon identity, detector/recovery, formulas, orders, scripts, prompts, and template variables.
- Role-neutrality scan policy covering Go production code, Core assets, generated/materialized metadata, formulas, orders, prompts, provider overlays, notification targets, test exceptions, generated files, and positive/negative controls.
- Existing-city upgrade matrix, before/after configs, compatibility-shim lifecycle, rollback/downgrade policy, version-skew policy, stale state handling, and in-flight work policy.
- Acceptance-to-proof matrix, replacement coverage for retired tests, pack-resolution test matrix, behavior-preservation harness, diagnostics golden-output checks, and documentation/terminology audit.

## Convergence Assessment
- Remaining blocker class: mixed
- Recommended apply verdict: iterate
- Reason: Several blockers are directly document-fixable: the artifact must be normalized to `gc.mayor.requirements.v1`, the front matter status corrected, W6H/examples/ACs/Open Questions added, and Core/public-Gastown/doctor/rollout/role-neutrality contracts specified. Some blockers also require external evidence or supporting artifacts, especially the validated migration ledger and behavior/test proof, but another requirements iteration is still necessary and can materially improve the next review.
- Next non-design work: generate or attach a validated asset inventory/ledger, define or prototype the absence-scan and behavior-preservation proof approach, and identify existing tests/docs that must be moved or replaced before retrying; these can be supporting artifacts, but `requirements.md` must stop asserting unproven file-level ownership as approved requirements.

## Recommended Changes
1. Rewrite `requirements.md` to `gc.mayor.requirements.v1`, set status to `draft`, `questions`, or `blocked`, and add schema-compliant W6H, Example Mapping, Acceptance Criteria, Out Of Scope, and Open Questions.
2. Resolve or explicitly list product decisions for Core loading, doctor no-fatal diagnostics, Core opt-out/dev-test behavior, public Gastown source/version/offline policy, legacy import handling, compatibility shims, rollback, and in-flight runtime state.
3. Replace the static file migration table with a validated asset migration ledger or move it into a supporting artifact; require path provenance, target paths, split boundaries, fallback classifications, duplicate reconciliation, and proof.
4. Add behavior-preservation requirements for Gastown successors, including Dog notification/escalation paths, formulas, orders, scripts, prompts, template variables, route targets, requester/detector relationships, and runtime proof commands.
5. Define the controller-vs-executor boundary and configurable maintenance-worker contract so Core infrastructure does not depend on a literal `dog` agent or any Gastown role.
6. Add a deterministic pack-resolution and materialization contract covering Core, public Gastown, `bd`, `dolt`, builtins, system packs, locked remotes, local overlays, legacy paths, stale `.gc/system` state, and collision behavior.
7. Add an acceptance traceability matrix mapping every outcome to tests, commands, absence scans, behavior harnesses, golden outputs, docs audits, or explicit manual checks.
8. Add diagnostics/remediation requirements with exact source attribution, output-channel safety, non-interactive behavior, idempotent repair semantics, read-only handling, duplicate prevention, and post-fix verification.
9. Add docs/examples/terminology cleanup criteria that enumerate retired authority paths and distinguish the retired Maintenance pack from valid store or supervisor maintenance language.
10. Provide a coverage-transfer table for legacy Maintenance/Gastown tests and require public external Gastown validation so local in-tree fallback cannot mask a broken migration.
