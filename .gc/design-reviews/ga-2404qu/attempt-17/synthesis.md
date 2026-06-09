# Design Review Synthesis

## Overall Verdict: block

The design has matured substantially, but it still leaves several migration invariants unresolved in executable terms. Multiple persona syntheses block on the same classes of risk: production loader bypasses, retired Maintenance activation paths, role-neutral Core boundaries, activation/version-skew safety, cache migration, and the exact slice gates needed to keep every intermediate commit green.

## Consensus Strengths
- Reviewers broadly praised the behavior-oriented witness floor: path counts, file existence, and include counts are explicitly insufficient; migrated behavior needs old and final execution witnesses under the same trigger conditions.
- The proposed generated artifacts are the right shape in principle: behavior manifests, role-surface inventories, loader inventories, public pin ledgers, slice gates, wording matrices, and test-migration maps are the correct durable control points.
- The symbolic binding direction for Core maintenance behavior is promising because it can make Core role-neutral if literals such as `dog`, `mayor`, and `deacon` are fully removed or classified.
- Test/fixture strategy is directionally sound: hermetic `test/packcompat`, minimal `fstest.MapFS` fixtures, no-Core containment in low-level config tests, and offline/cache fixtures are repeatedly identified as the right proof mechanisms.
- Several reviewers found the final design language stronger than prior iterations around preflight recipient validation, provider-pack rewrite exceptions, public pack pins, and same-slice docs/schema updates.

## Critical Findings

### [Blocker] Required Core Loading Is Not Yet Closed Over Every Production Path
**Sources:** Elias Sato; Sofia Khoury; Marcus Driscoll; Ritu Raman; Claude, Codex, DeepSeek, Gemini lanes.
**Issue:** The loader contract still permits or under-specifies bypasses. Persona syntheses call out missing real loader symbols such as `LoadWithIncludesOptions`, direct `config.Load` uses, no-refresh paths, controller reload, API/helpers, command completion/stop, raw TOML pre-reads, provider peeking, and wrapper functions. Required-pack integrity and typed participation are also not uniformly defined across missing, corrupt, partial, stale, extra-file, materialized-but-unresolved, repairable, no-refresh, and read-only cases.
**Required change:** Make a generated/default-deny loader inventory the source of truth for production config reads. Define resolver-produced `RequiredSystemPackParticipation` matched to validated file-set digests, choose a provider-pack selection algorithm, add focused allowlist rows for partial reads, and specify one failure-class by loader-class matrix. Plain diagnostics must use read-only/no-refresh APIs; mutation belongs only to explicit fix paths or a separately specified runtime repair protocol.

### [Blocker] Public Gastown, Core, And Retired Maintenance Boundaries Are Not Executable
**Sources:** Avery McAllister; Ingrid Kovac; Tomas Park; Marcus Driscoll; Nadia Volkov; Claude, Codex, DeepSeek, Gemini lanes.
**Issue:** The design does not yet choose one implementable model for dog prompt ownership and public Gastown's relationship to host Core. Stale Maintenance can still re-enter through local imports, generated system-pack directories, transitive imports, or public Gastown's current `../maintenance` import. Core-bound formulas, orders, scripts, TOML defaults, prompt fragments, and metadata still carry Gastown role assumptions without binding surfaces, relocation, or approved semantic deltas.
**Required change:** Pick and document the public Gastown/Core dependency model, including dog prompt patching or ownership, fragment resolution, duplicate-fragment behavior, and rendered prompt tests. Make retired Maintenance unimportable before config behavior discovery, rewrite or reject stale imports, add public-pack active-asset ledgers tied to the exact pin, and classify every Core/Gastown/provider asset as controller-owned, optional Core maintenance-worker, provider-pack, public-Gastown, compatibility, or retired.

### [Blocker] Role-Neutrality Enforcement Still Misses Live Production Surfaces
**Sources:** Ingrid Kovac; Avery McAllister; Tomas Park; Marcus Driscoll; Felix Moreau.
**Issue:** The proposed role scanner cannot prove zero hardcoded roles if it misses camelCase/PascalCase sub-identifiers and active Go surfaces such as theme functions, icon maps, prompt fallbacks, warmup defaults, `crew` API/dashboard vocabulary, `gastown` template wiring, public-pack constants, and dispatch heuristics. The design also lacks a surviving replacement for removed role-name guards and does not consistently split rejection rules for Core assets from preservation inventory for public Gastown assets.
**Required change:** Add a scanner contract with roots, file types, TOML fields, metadata fields, generated artifacts, scripts, templates, comments, case and sub-identifier matching, allowlist format, owners, expiry, and negative fixtures. Add a Go role-surface migration table with exact dispositions, rollout slices, replacement mechanisms, API/OpenAPI/dashboard handling, and tests, or explicitly narrow the migration claim if Go de-roling is deferred.

### [Blocker] Rollout, Activation Pin, Cache Migration, And Rollback Semantics Are Unsafe
**Sources:** Yuki Hayashi; Marcus Driscoll; Sofia Khoury; Felix Moreau; Tomas Park.
**Issue:** Activation sequencing is internally contradictory: one reading makes the activation pin and Maintenance removal atomic, while another permits an intermediate state where activation public Gastown coexists with bundled Maintenance. Old binary plus activation pin behavior is unresolved, public pin evidence is not normalized to a durable three-row model, existing synthetic cache and lock entries have no chosen online/offline migration path, and downgrade/rollback after doctor-mutated manifests or runtime state is not executable enough.
**Required change:** Normalize `public-gastown-pins.yaml` to `current_baseline`, `compatibility`, and `activation` rows with durable ref evidence and mandatory old/new binary results. Make activation pin switch plus Maintenance removal one candidate gate, or prove current-loader coexistence. Add generated rollback artifacts, cache migration matrices, legacy synthetic cache detection before ordinary git validation, stale-cache diagnostics or promotion rules, no-gap and zero-duplicate gates, and downgrade/upgrade state reconciliation.

### [Blocker] Doctor And Runtime Mutation Boundaries Can Still Mutate Or Tear State Unsafely
**Sources:** Sofia Khoury; Elias Sato; Felix Moreau; Marcus Driscoll.
**Issue:** Plain `gc doctor` is described as report-only but still points at materializing loaders. Multi-file publish semantics are not precise enough to prove rollback or safe resume after partial failure. Doctor/controller concurrency, lock scope, city identity, zero-write idempotency, generated-vs-custom pack detection, and runtime-state migration paths are all under-specified.
**Required change:** Split doctor into report-only and fix-intent phases. Route materialization, repair, cache promotion, lock generation, quarantine/prune, runtime-state migration, and config writes through a mutation coordinator under `--fix`. Define generation/swap or resumable partial-publish semantics, controller participation in concurrency control, process-table discovery, zero-write healthy `--fix`, digest/generation-marker custom detection, and runtime-state migration tables with operator commands.

### [Blocker] Source Deletion And Slice Gates Depend On Underspecified Generated Artifacts
**Sources:** Ritu Raman; Tomas Park; Nadia Volkov; Ingrid Kovac; Felix Moreau.
**Issue:** The migration relies on generated artifacts that do not yet have schemas, owners, generation status, digest/freshness tests, exact commands, fixture modes, forbidden intermediate states, or CI failure behavior. The behavior manifest can self-certify if discovery and freshness checks share the same blind spots. Rollout slices, attempt-9 matrix rows, attempt-14 A-F ordering, and `slice-gates.generated.yaml` are not reconciled.
**Required change:** Define contracts for `behavior-manifest.generated.yaml`, `role-surface.generated.yaml`, `loader-inventory.generated.yaml`, `slice-gates.generated.yaml`, `public-gastown-pins.yaml`, wording matrices, and test-migration maps. Add a canonical slice/commit crosswalk, make implementation beads cite generated row IDs, and add independent VCS-level and old-tree transcript reconciliation so moved/deleted/added assets cannot disappear through generator blind spots.

### [Blocker] Bootstrap Embed Removal Has Hidden Dependencies And No Singular No-Core Disposition
**Sources:** Ritu Raman; Elias Sato; Marcus Driscoll; Tomas Park.
**Issue:** Hidden dependency discovery is broader than path-string scanning. Reviewers identify coupling through `bootstrap.BootstrapPacks`, `bootstrap.PackNames`, `bootstrapSkillDirs()`, `GC_BOOTSTRAP`, bootstrap imports, skill materialization, cache helpers, provider registry functions, and prompt tests. The compile-safe contract for deleting the `//go:embed packs/**` directive and replacing `bootstrapAssets` with a non-nil empty/erroring filesystem still needs explicit tests.
**Required change:** Add import-graph and symbol-level guards for production bootstrap ownership, delete or narrowly constrain `GC_BOOTSTRAP=skip`, prove production bootstrap assets are empty and never read, move Core skill resolution to the new Core/systempack source, and use synthetic inline fixtures for bootstrap tests. Add slice ownership for every named consumer and prevent command/runtime tests from regaining no-Core behavior through low-level loaders.

### [Blocker] Operator Docs Cannot Be Correct Until Runtime State And Terminology Are Bound
**Sources:** Felix Moreau; Avery McAllister; Yuki Hayashi; Marcus Driscoll.
**Issue:** Runtime-state migration for JSONL archives, export cursors, push-failure counters, spawn-storm ledgers, and related doctor checks is not specified, so troubleshooting paths cannot be safely rewritten. The canonical system-pack reference page is missing or unregistered, the wording inventory does not cover all operator-facing file types and generated artifacts, and terminology around `maintenance_worker`, Core maintenance worker, maintenance agent, `dog`, dog pool, retired Maintenance, store maintenance, and Dolt maintenance is inconsistent.
**Required change:** Add a runtime-state migration table with current path, legacy read path, Core write path, doctor/fix behavior, rollback, and operator commands. Create or designate a registered system-pack reference page, make the docs/wording linter extension-aware across Markdown, MDX, JSON, TXT, docs nav, schema/reference outputs, CLI/help, doctor text, examples, scripts, pack comments, and public Gastown docs, and sequence same-slice docs/schema/help updates with the behavior changes they describe.

### [Major] Behavior Preservation Evidence Is Strong In Principle But Not Yet Grounded In Exact Witness Rows
**Sources:** Nadia Volkov; Avery McAllister; Marcus Driscoll; Tomas Park; Felix Moreau.
**Issue:** Multiple lanes endorse the witness floor but still ask for exact public-pack paths, old and final tests, behavior rows, pilot rows, generated old-test to new-test maps, and packcompat witnesses for branch pruning, Polecat formulas, detectors/requesters, review workflows, commands, doctor checks, overlays, prompts, role agents, provider pack scripts, warrants, and notifications.
**Required change:** Before destructive source moves or source deletion, generate the behavior manifest, role-surface inventory, old-tree baseline transcript, public pin ledger, test-migration map, pilot rows, and exact packcompat fixtures. Every removed assertion or moved behavior must map to Core, public Gastown, provider-pack, docs-only, or approved-retirement evidence.

### [Major] Provider Pack Continuity Is More Than Byte Preservation
**Sources:** Marcus Driscoll; Nadia Volkov; Ingrid Kovac; Elias Sato.
**Issue:** Provider-pack continuity still has unresolved behavior around `dolt-target.sh`, cross-pack script sourcing, configured recipients, `pool = "dog"`, conditional escalation, `DOG_DONE` nudges, author identities, health JSON consumers, and provider-specific host-pack selection. Byte identity or path preservation is insufficient if behavior-bearing routes change.
**Required change:** Add exception-ledger rows and old/new witnesses for provider formulas, scripts, recipients, pool bindings, author fields, and health consumers. Prove Core repair isolation for `bd`/`dolt` bytes and provenance unless an explicit rewrite row authorizes the delta.

### [Minor] Stale Prose, Comments, And Source Labels Still Need Cleanup
**Sources:** Nadia Volkov; Ritu Raman; Felix Moreau; Tomas Park.
**Issue:** Several persona syntheses note weaker historical prose, stale comments, outdated examples, inconsistent source labels, and lingering references to old paths or retired Maintenance semantics. These are not independently blocking, but they can mislead implementation beads and operators if left outside the generated wording and slice gates.
**Required change:** Make implementation beads cite the final binding contracts and generated row IDs, regenerate docs/schema/help in the same slice as behavior changes, and classify retained historical terms through the wording matrix.

## Disagreements
- Verdicts varied by persona and model, but worst-verdict-wins applies. The strongest approvals generally endorsed the direction while the blocking reviews identified concrete invariants that remain unimplemented or contradictory.
- Several dog ownership models are viable: public Gastown can own dog assets, patch host Core, explicitly import Core if rules permit it, or use host auto-inclusion plus symbolic patches. The design must choose one; the disagreement is about mechanism, not about whether the mechanism is required.
- Runtime repair can either fail closed and delegate mutation to `gc doctor --fix`, or use a short deterministic repair lock with atomic publish. Reviewers accept either if it is explicitly specified and tested across failure and concurrency cases.
- Activation can be one-way, or it can support old-binary compatibility. Reviewers disagree on which is intended, but agree that the pin ledger and release plan must record observed results, manual recovery, and rollback constraints before activation.
- Public Gastown and provider-pack scans must both reject forbidden Core role leakage and preserve Gastown behavior inventory. The synthesis assessment is that scanner behavior must differ by asset ownership instead of using one broad reject rule everywhere.

## Missing Evidence
- Generated artifact schemas, owners, entrypoints, exact commands, freshness/digest tests, CI failure modes, and row IDs for the behavior manifest, role surface, loader inventory, slice gates, public pins, wording matrix, and test-migration map.
- Complete loader inventory and scanner fixtures for production config paths, aliases, wrappers, raw TOML reads, partial reads, no-refresh paths, provider peeking, and new exported `*City`-returning resolvers.
- Resolver-produced typed required-pack participation matched to validated file-set digests, plus failure-class by loader-class behavior for broken required packs.
- Public Gastown dependency/import decision, dog prompt ownership model, rendered prompt goldens, and exact pinned-checkout asset/test ledger.
- Tests proving stale Maintenance directories, stale local Gastown imports, transitive retired imports, prompt/template discovery, and retired aliases cannot become active config layers.
- Public pin ledger rows and matrix evidence for current baseline, compatibility, activation, old binary plus activation pin, downgrade/upgrade, stale cache, offline cache, and no-cache states.
- Doctor report-only goldens, fix-path failure-injection tests, zero-write healthy `--fix`, doctor/controller concurrency matrix, generated-vs-custom pack provenance, and runtime-state migration table.
- Import-graph/symbol guard output for hidden bootstrap dependencies, plus tests for non-nil empty/erroring bootstrap assets and production no-Core containment.
- Executed tutorial/troubleshooting transcript coverage, registered system-pack reference navigation, generated schema/help/reference freshness, and public Gastown companion docs/comment lint.

## Recommended Changes
1. Close the Core loader invariant first: generate the loader inventory, define typed required-pack participation, choose provider-pack selection, and add the fatal/read-only/fix loader matrix.
2. Resolve the public Gastown/Core/Maintenance boundary: choose the dependency model, make retired Maintenance unimportable, specify dog prompt ownership, add binding surfaces, and seed the exact public-pack asset/test ledger.
3. Make role neutrality enforceable across Go, assets, API/OpenAPI/dashboard, scripts, templates, TOML, metadata, and generated artifacts with ownership-aware scanner behavior and expiring compatibility rows.
4. Normalize rollout and pin artifacts: three public-pin rows, activation gate semantics, old/new binary matrix, cache migration contract, rollback transcript, and no-gap/zero-duplicate slice gates.
5. Split doctor/report-only and fix-intent semantics, then define publish atomicity, concurrency, zero-write idempotency, custom-pack detection, runtime-state migration, and operator recovery.
6. Define all generated artifact contracts and reconcile rollout representations into one canonical slice/commit crosswalk used by implementation beads.
7. Add independent behavior-manifest completeness checks from VCS moves/deletions and old-tree execution transcripts before any destructive source move or source deletion.
8. Lock down bootstrap embed removal with import/symbol guards, synthetic fixtures, Core skill rematerialization, and a singular decision for `GC_BOOTSTRAP=skip`.
9. Expand docs/DX gates to include navigation, MDX, JSON/TXT schemas, generated references, CLI/help, doctor text/JSON/FixHint strings, tutorials, examples, scripts, pack comments, and public Gastown companion docs.
10. Re-run the design review after the design includes concrete generated rows, fixtures, and gate commands rather than only prose-level acceptance criteria.
