# Design Review Synthesis

## Overall Verdict: block

Multiple personas agree that the design is much stronger than earlier iterations, especially around the two-pin rollout shape, behavior manifests, wording-matrix direction, and bootstrap/Core separation. It is still not implementation-ready: several blockers remain in the maintenance-worker/dog contract, required Core loading invariants, no-Maintenance activation proof, role-neutrality surface, and public Gastown/Core boundary.

## Consensus Strengths
- Reviewers repeatedly praised the two-pin public Gastown rollout model as the right high-level shape, provided compatibility and activation pins are made asset-level, immutable, and mechanically gated.
- The behavior-manifest and witness-floor direction is strong because it moves preservation proof from narrative file inventories toward execution-level evidence.
- The retired-source classifier and centralized system-pack wording matrix are the right mechanisms for consistent diagnostics, docs, and stale-source handling.
- The mutation-coordinator direction for doctor fixes is safer than direct ad hoc writes and aligns with CST preservation, preflight, staged validation, and post-publish reruns.
- The bootstrap target architecture is sound: production bootstrap should stop embedding Core, tests should use minimal explicit fixtures, and real Core fidelity should move to system-pack validation.

## Critical Findings

### [Blocker] Maintenance-worker, dog, and public Gastown/Core ownership are not concrete
**Sources:** 01-behavior-preservation-auditor, 04-zfc-role-neutrality-guardian, 08-gastown-pack-boundary-reviewer, 10-test-slicing-coverage-verifier
**Issue:** The design still does not specify the concrete Core maintenance-worker binding: key name, config table, order/formula syntax, rename/omit behavior, diagnostics, public-pack patchability, and whether Gastown owns a replacement worker or rebinds Core. This leaves formula pools, `gc.routed_to`, mail/nudge targets, prompt fragments, dog defaults, and requester/detector behavior ambiguous.
**Required change:** Choose and document the host-Core/public-Gastown dependency model end to end. Update `pack.toml`, behavior manifests, diagnostics, template-fragment rules, duplicate-definition gates, rendered-prompt tests, and execution witnesses to prove the selected model works with renamed and omitted maintenance-worker configurations.

### [Blocker] The no-Maintenance activation proof is not executable as written
**Sources:** 01-behavior-preservation-auditor, 05-rollout-version-skew-planner, 08-gastown-pack-boundary-reviewer, 10-test-slicing-coverage-verifier
**Issue:** Multiple personas found contradictory gate ordering: the plan asks for no-Maintenance packcompat proof before the production loader can actually exclude Maintenance. The design also refers to inactive compatibility assets, but the loader does not have a defined inactive asset mechanism, so duplicate active definitions remain a real risk.
**Required change:** Rewrite the rollout sequence so approval-grade no-Maintenance proof runs through the normal production loader after the candidate slice removes Maintenance from active required packs. Replace inactive-asset prose with an asset-level compatibility/activation pin table, preferably omitting colliding assets from the compatibility pin and introducing them only in the activation pin.

### [Blocker] Required Core loading and production bypass controls are incomplete
**Sources:** 02-required-core-loading-invariant-reviewer, 06-pack-registry-and-cache-tester, 07-bootstrap-fixture-isolation-reviewer
**Issue:** The design has not made required Core failure one uniform fatal production contract. Production config load inventories remain incomplete, internal behavior-driving loaders are not fully classified, provider-required pack selection can diverge from the final effective beads provider, and `bootstrapManagedImportNames`/`GC_BOOTSTRAP=skip` sequencing can still allow shadowing or bypass of required-Core validation.
**Required change:** Define two fatal gates for production runtime loads: pre-resolution required file-set integrity and post-resolution typed required-pack participation. Publish a deny-by-default production loader inventory and scanner, define `RequiredSystemPackParticipation`, classify diagnostic/no-refresh paths, and sequence bootstrap managed-name removal only after replacement collision and implicit-import checks are live.

### [Blocker] Role-neutrality and provider-pack surfaces are not fully inventoried or guarded
**Sources:** 01-behavior-preservation-auditor, 04-zfc-role-neutrality-guardian, 06-pack-registry-and-cache-tester, 08-gastown-pack-boundary-reviewer, 09-docs-dx-consistency-reviewer
**Issue:** Required or embedded provider packs and examples still contain hardcoded `dog`/`mayor`/`deacon` behavior while the design also claims provider byte-continuity. Reviewers also flagged production Go role surfaces, dashboard/API `agent_kind`/`crew` projections, TOML defaults, prompt text, scripts, overlays, docs, and generated artifacts as either uncovered or underspecified by the scanner.
**Required change:** Bring `examples/dolt`, `examples/bd`, provider formulas/scripts, production Go, dashboard/OpenAPI/generated TypeScript, TOML fields, prompts, overlays, and docs into the role-surface inventory. Either replace hardcoded role targets with configured resolution or create narrow, owned, expiring compatibility rows with release gates and tests.

### [Major] Doctor migration safety still needs an enforceable write protocol
**Sources:** 03-doctor-migration-safety-reviewer, 05-rollout-version-skew-planner, 06-pack-registry-and-cache-tester
**Issue:** The mutation coordinator is not yet an enforceable boundary over existing doctor fix paths. Direct writes can still bypass preflight, generated/custom provenance, compare-before-rename, controller exclusion, staged validation, and post-publish reruns. Runtime-state copy/move semantics, offline repair, doctor-vs-doctor races, and healthy no-op byte identity are also not fully specified.
**Required change:** Add concrete `MutationCoordinator`/`FixIntent` contracts, forbid migration-affecting direct writes in `Check.Fix`, define generated provenance and accept/reject matrices, choose non-destructive runtime-state semantics, add crash-released concurrency exclusion or equivalent, and require two-phase doctor reporting with post-publish reruns.

### [Major] Pin, cache, registry, and offline-upgrade semantics need mechanical gates
**Sources:** 05-rollout-version-skew-planner, 06-pack-registry-and-cache-tester, 03-doctor-migration-safety-reviewer
**Issue:** The public Gastown pin ledger, durable-ref policy, old-binary evidence, public remote cache identity, synthetic alias retirement, active materialization set, stale synthetic cache rejection, and offline/air-gapped behavior are not yet precise enough. Several reviewers also noted rematerialization I/O/concurrency risk when synthetic content hashes change.
**Required change:** Make `public-gastown-pins.yaml` authoritative with phase, immutable commit, digest, cache identity, rollback/one-way-upgrade status, old-binary evidence, and transcript fields. Align `config` and `packman` cache keys, disable public synthetic fallback in the correct slice, distinguish active materialization from retired-source diagnostics, and add offline/cache and concurrent self-heal fixtures.

### [Major] Docs, DX, and wording gates are promising but incomplete
**Sources:** 09-docs-dx-consistency-reviewer, 04-zfc-role-neutrality-guardian
**Issue:** The wording matrix is the right mechanism, but concrete scan commands and named docs scope are too narrow. The design misses `.mdx`, `.json`, `.txt`, docs navigation, generated schemas, moved asset names such as `prune-branches`, and troubleshooting prose whose causal model changes even if tokens are cleaned up.
**Required change:** Use extension-agnostic stale-path and moved-asset scans, name the wording-matrix artifact/schema/owner/freshness test, add allowed contexts for store-maintenance terminology, and bind docs/navigation/generated-reference updates to the same rollout slices as the behavior they describe.

## Disagreements
- Several personas approve with risks while others block. Worst-verdict-wins makes the global verdict `block` because the blocking lanes identify executable-order contradictions and role/loading contracts that would make implementation beads unsafe or ambiguous.
- Reviewers agree the two-pin rollout shape is directionally correct, but disagree on whether the current text is sufficient. My assessment: keep the shape, but do not approve until the compatibility/activation asset table and production-loader gate ordering are explicit.
- Provider-pack role leakage severity varies from Major to Blocker. My assessment: it is block-relevant because required host packs carrying hardcoded Gastown targets would violate SDK self-sufficiency, while silently rewriting them would violate behavior preservation unless the change is inventoried and tested.
- Empty recipient behavior is unsettled. Some reviewers prefer skipping with an observable warning; others require escalation failures to fail loudly. The design should distinguish intentionally unconfigured optional recipients from configured delivery failure.
- Offline upgrade and downgrade requirements differ by reviewer. The minimum acceptable outcome is an executable, tested operator story: either cache-backed/offline recovery works, or a one-way/manual limitation is ledger-enforced and release-blocking.

## Missing Evidence
- Concrete maintenance-worker binding contract, public Gastown host-Core declaration, and tests for renamed/omitted worker configurations.
- Asset-level compatibility and activation pin contents, including duplicate-definition avoidance and no-Maintenance production-loader proof.
- Deny-by-default production config-load inventory, typed required-pack participation model, and provider-required pack selection against the final effective config.
- Full role-surface inventory and scanner roots covering provider packs, Go, dashboard/API/generated types, TOML fields, prompts, scripts, overlays, and docs.
- Packcompat execution harness, exact public pin provenance, offline/cache fixtures, old-binary fixtures, and behavioral-trigger witnesses for warning/escalation branches.
- Doctor mutation coordinator interfaces, generated/custom provenance matrix, concurrency exclusion, runtime-state semantics, and healthy no-op byte-identity scope.
- Wording-matrix artifact/schema/owner, extension-agnostic docs scans, generated schema updates, and moved-asset-name lint coverage.

## Recommended Changes
1. Specify the Core maintenance-worker/public Gastown dependency model first, because it unblocks dog ownership, route/mail/nudge targets, prompt fragments, role-neutrality scanning, and behavior witnesses.
2. Rewrite the rollout sequence and pin table so the no-Maintenance production-loader packcompat gate is executable and inactive compatibility assets are replaced by explicit compatibility/activation asset ownership.
3. Define and test the required Core loading boundary: pre-resolution integrity, post-resolution typed participation, deny-by-default loader bypass scanner, provider-pack selection, and bootstrap/skip sequencing.
4. Expand the role-surface manifest and scanner to provider packs, examples, Go, API/dashboard/generated artifacts, TOML fields, prompts, scripts, overlays, docs, and generated references.
5. Add the doctor mutation protocol and generated/custom provenance matrix before any slice consumes a new public Gastown pin or rewrites migration-affecting imports.
6. Make the pin/cache/registry contract mechanical: authoritative ledger, durable ref checks, exact remote cache identity, stale synthetic rejection, active materialization boundary, offline/cache fixtures, and old-binary evidence.
7. Bind docs, wording matrix, tutorials, troubleshooting, generated schemas, and moved-asset-name lint to the rollout slice that changes each behavior.
