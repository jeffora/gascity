# Design Review Synthesis

## Overall Verdict: approve-with-risks

All ten persona syntheses returned `approve-with-risks`, so the global verdict is `approve-with-risks` under the workflow's worst-verdict-wins rule. The design direction is consistently viewed as sound: formulas declare requirements, the active Gas City binary decides capability, and accepted compiler artifacts become the durable-write boundary. The remaining risk is that several load-bearing contracts are still prose-level and need executable manifests, fixtures, and release gates before implementation can proceed safely.

## Consensus Strengths

- Multiple personas praised the central compiler-boundary model: formula source declares requirements, while the active Gas City host decides whether it can satisfy them.
- Reviewers consistently supported moving durable writes behind normalized requirements, `CompileResult`, `AcceptedCompileArtifact`, projection snapshots, and accepted-artifact validation.
- The migration is not a flag-day plan. Personas credited the phased parser/model, accepted-artifact, caller migration, docs, resolver enforcement, first-party conversion, and alias-removal checkpoints.
- The design treats pack revision, content/provenance, and lockfile identity as the reproducibility boundary instead of reviving formula-level semver.
- The closed requirement grammar, raw-capture parser boundary, v2 construct registry, generated validation matrix, and typed diagnostics were repeatedly identified as the right controls against heuristic drift.
- External pack migration is treated as an operational workflow with validation, migration hints, release evidence, and conservative gates rather than a changelog-only change.
- The terminology work is much stronger than prior docs and gives the project a usable distinction between formula compiler capability, host capability, pack compatibility, pack import versions, legacy `contract`, and formula `version`.

## Critical Findings

### [Blocker] Host-Downgrade Continuation Is Not Implementable Yet

**Sources:** Yuki Patel / Claude+Codex; Felix Berger / Claude+Codex; Ibrahim Park / Claude+Codex; Nadia Sorenson / Claude+Codex

**Issue:** The design says already-accepted graph roots may continue after `[daemon] formula_v2 = false`, while `ValidateAcceptedArtifact` appears to fail when write intent disagrees with current host capability. That leaves operators and implementers without one normative rule for retry, continuation, repair, sibling create, and graph-specific mutation under host downgrade.

**Required change:** Define one host-downgrade contract. Either permit a narrow same-identity accepted-artifact reuse mode, with exact identity fields and operator-visible diagnostics, or state that downgrade halts graph-specific mutation until reenabled. Add fixtures for reuse, identity mismatch, unsupported artifacts, zero-write fail-closed behavior, concurrent retry/create, and dashboard/API visibility.

### [Major] Durable Writers Need an Accepted-Artifact-Only Boundary

**Sources:** Yuki Patel / Claude+Codex; Felix Berger / Claude+Codex; Nadia Sorenson / Claude+Codex; Lena Driscoll / Claude+Codex

**Issue:** The design names `AcceptedCompileArtifact` as the boundary, but it does not yet enumerate every existing durable writer and writable helper that must require an accepted artifact. Current paths such as molecule cook/attach/instantiate helpers, sling, fanout, order dispatch, convergence create/retry/iterate/repair, CLI cook paths, and preview/test helpers can otherwise preserve bare-recipe write authority.

**Required change:** Add a durable-writer migration table with a per-symbol end state: accepted-artifact-only, unexported internal helper, test-only/preview-only helper, compatibility writer, or deleted. Include blocking tests that fail while any durable write path can persist root, child, hook, convoy, wisp, retry, convergence metadata, or artifact refs from a bare `*formula.Recipe`.

### [Major] Caller Inventory and Legacy Helper Retirement Are Not Yet Executable

**Sources:** Yuki Patel / Claude+Codex; Nadia Sorenson / Claude+Codex; Felix Berger / Claude+Codex; Lena Driscoll / Claude+Codex

**Issue:** The caller migration depends on a current grep-derived manifest, but that manifest is still described as a future artifact. Legacy helpers such as `declaresGraphV2Contract`, `requiresExplicitGraphContract`, `metadataRequiresGraphContract`, `isGraphWorkflow`, `formulaV2Enabled`, `IsFormulaV2Enabled`, `SetFormulaV2Enabled`, and convergence subset wrappers have ambiguous outcomes.

**Required change:** Make the manifest a Phase 0 acceptance artifact with one row per production occurrence, file/line, current semantics, replacement API, durable-write boundary, owner, expiry phase, and blocking test. Replace "retired or rewritten as compatibility shim" with concrete delete, parser-shim, compatibility-writer, or temporary-delegator decisions for each helper.

### [Major] Requirement Grammar and Validation Matrix Need More Normative Rows

**Sources:** Priya Zielinski / Claude+Codex; Ibrahim Park / Claude+Codex; Marta Hidalgo / Claude+Codex

**Issue:** The closed grammar is the right model, but important edge cases remain prose-only: `>=0`, `>=-1`, `>=01`, `>=1.0`, `>=+2`, overflow, control bytes, Unicode lookalikes, TOML `[[requires]]`, datetime values, JSON malformed values, mixed unknown axes, and transitive v2 constructs. Without rows, implementers can disagree on diagnostic codes, ordering, and source attribution.

**Required change:** Add rendered fixture rows for malformed strings, unsupported future values, TOML/JSON shape errors, mixed-axis diagnostics, legacy `version` interactions, and transitive contribution cases. State how suite count locks are derived from dimensions or registries so CI fails when new caller paths or constructs are omitted.

### [Major] Operator Diagnostics and Grouping Semantics Are Underspecified

**Sources:** Marta Hidalgo / Claude+Codex; Elias Vega / Claude+Codex; Lena Driscoll / Claude+Codex

**Issue:** Background diagnostics promise bounded events and visible grouped state, but the design does not define "scan series", occurrence count updates, producer-state ownership, restart behavior, config reload reset triggers, dashboard/API projection, or direct CLI versus API-routed CLI warning cadence. This can lead to either event spam or hidden repeated failures.

**Required change:** Define the durable grouped-diagnostic state schema, grouping key, occurrence count fields, update cadence, reset rules, restart/controller-handoff behavior, and failure behavior if grouped-state writes fail. Add repeated-scan, unrelated config reload, API/Huma/dashboard, direct CLI, API-routed CLI, and report-mode fixtures.

### [Major] Alias Removal and External-Pack Release Gates Need a Single Auditable Contract

**Sources:** Elias Vega / Claude+Codex; Saoirse Raman / Claude+Codex; Lena Driscoll / Claude+Codex; Avery Brooks / Claude+Codex

**Issue:** The design has the right evidence-based intent, but the authoritative removal command, exit semantics, release artifacts, external-support acquisition path, maintainer outreach model, abandoned/unreachable pack policy, and SHA-pinned consumer handling are not concrete enough. The gate could become release judgment instead of executable evidence.

**Required change:** Make `gc formula validate --all-packs --alias-removal-gate --json` or another named command the single authoritative gate. Define exit behavior for `legacy_only`, `dual_declared`, unsupported future requirements, missing/unreadable artifacts, shadowed external formulas, auth failures, unreachable sources, and stale external-support rows. Seed release artifacts with conservative blocking defaults before diagnostics or first-party source conversion ship.

### [Major] Convergence Projection Ownership and Subset Parser Retirement Are Ambiguous

**Sources:** Felix Berger / Claude+Codex; Yuki Patel / Claude+Codex; Nadia Sorenson / Claude+Codex

**Issue:** `formula.CompiledConvergenceProjection` and `convergence.ConvergenceMetadata` overlap on enabled state, evaluate prompt fields, runtime variables, retry policy, source attribution, reserved step IDs, requirements, and artifact refs. The legacy convergence subset parser and helper exports are not mechanically mapped to delete or delegate, so convergence can keep a second parser/validation authority.

**Required change:** Define package-boundary data flow from `cmd/gc` through `internal/formula.CompileWithResult`, acceptance, projection, validation, and durable bead writes. Assign ownership of each overlapping field. Add an exported-symbol migration table and make `TestNoConvergenceSubsetParserUse` a checked allowlist forbidding raw TOML reads, raw `[requires]` inspection, `convergence.Formula{}` construction, subset imports, and required-var reconstruction outside compiler projection.

### [Major] Documentation and DX Gates Are Not Yet Copy-Paste Safe

**Sources:** Avery Brooks / Claude+Codex; Lena Driscoll / Claude+Codex; Yuki Patel / Claude+Codex; Elias Vega / Claude+Codex

**Issue:** The terminology map is strong, but the planned `docs/reference/formula.md` rewrite is still a skeleton and existing docs continue to teach stale `version`, `contract`, `Store.MolCook`, `bd mol cook`, and `GC_NATIVE_FORMULA=false` guidance. The docs rollout also needs to include contributor architecture docs in the same branch as user-visible diagnostics.

**Required change:** Make the reference-doc draft or doctest fixture a blocking Phase 2 artifact. Add a "which key do I edit" section, copy-paste-safe external-pack examples pairing `[requires] formula_compiler = ">=2"` with concrete `[pack] requires_gc`, structured stale-guidance checks with false-positive fixtures, and same-branch updates for `engdocs/architecture/formulas.md`, proposal docs, reference docs, CLI help, API, and dashboard copy.

### [Major] Forward-Compatibility and Accepted-Artifact Identity Need Normative Structs

**Sources:** Ibrahim Park / Claude+Codex; Felix Berger / Claude+Codex; Saoirse Raman / Claude+Codex

**Issue:** Future capability behavior is conceptually right but too implicit. Released constructs must be additive-only; released requirement-axis grammar must be frozen; schema-version rules must distinguish same-axis capability bumps from new axes; and `CompileIdentity` / `CompileWriteIntent` must match the identity contract used by accepted-artifact validation.

**Required change:** Add forward-compatibility invariants and normalize identity into one struct or derived contract containing formula/source identity, content hash, search/options/vars hash, host capability/config generation, binding identity hash, artifact version, requirements schema version, and projection snapshot version. Add negative fixtures for binding identity and projection snapshot mismatches, plus old-reader fixtures for future compiler capability combined with future axes.

### [Minor] Formula `version` and New-Author Misuse Need Sharper Diagnostics

**Sources:** Saoirse Raman / Claude+Codex; Avery Brooks / Claude+Codex; Priya Zielinski / Claude+Codex; Ibrahim Park / Claude+Codex

**Issue:** The design correctly avoids formula semver, but `version` remains a likely footgun for authors expecting `version = 2` to enable graph workflows. Current matrix excerpts and docs do not fully separate legacy tolerated metadata, invalid values, and new misuse diagnostics.

**Required change:** State whether formula `version` is preserved indefinitely or removed in a named deprecation phase. Add explicit matrix rows for omitted, valid legacy, string, invalid-type, and unsupported values, and consider a distinct `formula.version_misuse` warning pointing authors to `[requires]` and pack revision identity.

## Disagreements

- Several personas had Claude return `approve-with-risks` while Codex returned `approve`; this occurred in compiler boundary, parser matrix, rollout sequencing, pack ecosystem, and related lanes. My assessment: Codex generally accepted the architecture direction, while Claude identified design-level evidence gaps. The global synthesis should carry the stricter risk posture because these gaps define migration safety.
- Caller integration had the strongest severity disagreement. Claude treated host-downgrade continuation as blocker-level; Codex considered in-flight behavior defined in principle. My assessment: the design cannot be implemented consistently until the downgrade contract is normative and testable, so this remains the highest-priority required change even though the persona verdict stayed `approve-with-risks`.
- Operator diagnostics reviewers differed on whether event-only warning visibility and HTTP status wording are core design issues or fixture-level polish. My assessment: the exact surface can be implementation-owned, but grouped state, reset behavior, warning cadence, and typed dashboard/API projection must be design-owned.
- Pack/versioning reviewers differed on how much external author workflow belongs in v0 design versus implementation sequencing. My assessment: a separate lint concept is unnecessary, but the design must name one concrete first-formula validation path and release-captain evidence lifecycle before alias-removal gates can be trusted.
- Convergence reviewers differed on whether compatibility shims are acceptable. My assessment: temporary shims are acceptable only as one-call-deep delegators to compiler projection with explicit deletion or allowlist rules.
- Kimi 2.6 was absent from all persona syntheses. The bead contract allows this when skipped or soft-failed, and the source bead has `design_review.skip_gemini=true`, so this is missing optional diversity rather than a workflow failure.

## Missing Evidence

- No Kimi 2.6 review artifacts were present for any persona.
- No checked-in current caller manifest or raw-consumer allowlist proves every production occurrence has been captured before migration begins.
- No durable-writer end-state table covers molecule helpers, sling, order dispatch, fanout, convergence, CLI cook/attach paths, preview/test-only paths, and compatibility writers.
- No seeded release artifacts are shown for compatibility, minimum-floor, external-support, stale-guidance, first-party inventory, or alias-removal evidence.
- No old-reader compatibility corpus output, external SHA-pinned pack fixtures, auth/unreachable-source fixtures, shadowed formula fixtures, or optional `bd` probe status is present.
- No rendered `docs/reference/formula.md` draft or doctest fixture proves the final docs order, copy-paste snippets, glossary wording, and stale-guidance false-positive behavior.
- No Huma/OpenAPI/generated TypeScript fixture exists for grouped producer failures or dashboard rendering of typed formula diagnostics.
- No convergence projection snapshot or persisted-artifact fixture proves projection can run without reopening original formula source.
- No host-downgrade read model, CLI/API/dashboard artifact, or golden fixture shows active graph roots under disabled host capability.
- No explicit external-support lifecycle defines owner, notification channels, outreach SLA, abandonment path, opt-out semantics, or who may expire stale external rows.
- No normative identity table covers retry, next iteration, missing-child repair, binding identity, projection snapshot, search paths, host capability, config generation, artifact version, and requirements schema version.

## Recommended Changes

1. Resolve host-downgrade semantics and accepted-artifact identity first; this drives caller migration, convergence, operator diagnostics, and artifact validation.
2. Add the Phase 0 caller and durable-writer manifests with per-symbol end states, owners, expiry phases, replacement APIs, and blocking tests.
3. Make convergence an accepted-artifact consumer only: define projection ownership, delete or delegate subset parser exports, and add zero-write fixtures for every durable convergence path.
4. Expand the requirement grammar, mixed-axis, transitive contribution, JSON/TOML shape, legacy `version`, and future-capability matrices with exact diagnostic codes, ordering, and source attribution.
5. Define the grouped diagnostic state contract and add fixtures for repeated scans, restarts, config reloads, CLI/API/dashboard parity, and warning/report cadence.
6. Name the authoritative alias-removal gate command and seed conservative release artifacts for compatibility, min-floor, external-support, stale guidance, and first-party inventory.
7. Specify the external-pack lifecycle, including release-captain ownership, outreach evidence, unreachable/auth-failed sources, SHA-pinned consumers, shadowed formulas, and first-formula validation.
8. Make docs a same-branch gate for user-visible diagnostics: reference docs, architecture docs, proposal supersession, CLI help, API/dashboard copy, and structured stale-guidance checks.
9. Add forward-compatibility invariants: additive-only constructs, frozen released axis grammar, schema-version rules, multi-axis construct capability representation, and old-reader observation-only behavior.
10. Decide and document the formula `version` policy, including misuse diagnostics and explicit matrix coverage for legacy and invalid values.
