# Avery McAllister

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek, Gemini

**Consensus findings:**
- [Blocker] Public Gastown's post-Maintenance dependency model is still not concrete enough. Claude and DeepSeek both find that the design depends on Core-owned `dog` behavior while public Gastown currently has Maintenance-era imports, patches, prompt fragments, and route assumptions. Codex requires a direct host-Core binding rule for public Gastown assets. The design must choose exactly one model: explicit Core import, host auto-inclusion plus patch layering, or public-Gastown-owned dog assets.
- [Blocker] Dog prompt preservation does not yet have an executable ownership mechanism. Claude identifies a contradiction between preserving Gastown-specific requester/detector/notification prose and only allowing Gastown to patch Core `dog` theming or `work_dir`. DeepSeek expands this to unresolved template-fragment ownership for `propulsion-dog`, `architecture`, and `following-mol`. Without a named mechanism, the split either loses Gastown behavior or puts Gastown role prose back into Core.
- [Blocker] Retired Maintenance must be rejected before config loading, not merely removed from `requiredBuiltinPackNames`. DeepSeek flags stale `.gc/system/packs/maintenance/pack.toml`, stale local Gastown imports, and transitive imports as paths that can still activate retired config. Claude and Codex also require pinned public-pack wording/path lints proving Maintenance imports and implicit-Maintenance comments are gone.
- [Blocker] Core-bound formulas, scripts, orders, and prompt assets still need role-neutral dispositions. DeepSeek calls out TOML defaults such as `default = "deacon"` in `mol-shutdown-dance.toml`; Claude flags script/formula/mail/nudge targets that still reference Gastown roles. These must become configurable, generic Core bindings, public-Gastown-owned assets, or approved semantic deltas.
- [Major] Public-pack behavior preservation needs an exact executable ledger, not broad prose. Codex and Claude both require a generated migration map tying old in-tree assertions to public `gascity-packs/gastown` tests or approved removals before Gas City deletes `examples/gastown` and Maintenance behavior coverage.
- [Major] The active public Gastown route boundary needs an explicit rule. Public Gastown prompts, formulas, pools, `gc.routed_to`, warrant metadata, commands, mail, and nudge targets must use the configured host-Core binding when targeting the Core maintenance worker, never hardcoded `dog` or `{{binding_prefix}}dog`, unless `dog` is a public-Gastown-owned agent.
- [Major] `session_live` hooks, dog `agent.toml` fields, prompt fragments, commands, doctor checks, branch pruning, Polecat formulas, and review workflow checks need final owners and paths. Reviewers agree the design has the right audit shape but still lacks enough row-level disposition for implementers to execute the split safely.
- [Major] Rollout and stale-path behavior remain underspecified. The activation pin may be a one-way version-skew boundary for older strict TOML loaders, and the design still needs cases for new binaries with stale local Gastown imports, downgrade/upgrade around stale generated pack directories, and intentional developer `examples/gastown/packs/*` checkouts.
- [Minor] Source-derived asset lists need cleanup. Claude flags `gastown/scripts/prune-branches.sh` versus the public pack's `gastown/assets/scripts/...` convention, and `mol-polecat-report.toml` as a named asset that does not exist in the current Core formula set.

**Disagreements:**
- Verdicts differ. Claude and DeepSeek block; Codex approves with risks; Gemini says the design is structurally sound with residual rollout risks. Assessment: this lane should block because the unresolved items decide whether public Gastown can load, route, and preserve behavior without Maintenance.
- Gemini asserts that the latest design already resolves dog fragment ownership, retired-source rejection, and TOML role scanning through `internal/packsource`, symbolic `target_binding`, and expanded scanners. Claude and DeepSeek still identify missing design rows and concrete current-path hazards. Assessment: the claimed mechanisms are acceptable only once the design records the exact import-chain decision, path classifier behavior, scanner scope, and packcompat witnesses.
- Reviewers propose multiple viable dog models: public Gastown patches an auto-included Core dog, public Gastown imports Core explicitly, Core accepts prompt-body patches, or public Gastown owns a dog-like agent. Assessment: any model could work, but the design must pick one and define collision behavior, fragment lookup, duplicate-agent rules, and renamed-worker tests.
- There is partial disagreement over stale retired directories. Some reviews tolerate stale files remaining on disk for rollback or operator preservation, while others ask for cleanup. Assessment: stale directories may remain only if they are provably unimportable and surfaced through doctor advisories or safe rewrites.
- Old-binary compatibility differs by pin phase. Assessment: compatibility-pin support and activation-pin support must be stated separately; activation should be documented as unsupported for old binaries unless tested.

**Missing evidence:**
- The final public Gastown `pack.toml` import-chain decision replacing `[imports.maintenance] source = "../maintenance"`.
- A test proving public Gastown can patch an auto-included Core agent without importing Core, or a revised contract allowing explicit Core imports.
- Rendered prompt golden tests for Core-only and public-Gastown cities after same-name template-fragment shadowing is removed.
- Behavior inventory rows for dog prompt fragments, `session_live` hooks, dog `agent.toml` fields, TOML defaults with role names, mail/nudge targets, author identities, escalation scripts, commands, doctor checks, overlays, and review checks.
- Scanner fixtures covering role names in TOML defaults, TOML step descriptions, active scripts, orders, commands, prompt fragments, and public-pack comments/imports.
- Tests proving stale `.gc/system/packs/maintenance/pack.toml`, stale local Gastown imports, and transitive retired-pack imports cannot become active config layers.
- A `test-migration.generated.yaml` sample that maps old `examples/gastown/gastown_test.go` and `maintenance_scripts_test.go` assertions to public-pack tests, Core tests, provider-pack tests, documentation-only rows, or approved removals.
- Packcompat evidence against the exact pinned public Gastown checkout for Polecat, branch pruning, detector/requester behavior, review workflows, commands, doctor checks, overlays, prompt fragments, and role agents.
- Release matrix rows for old binary plus compatibility pin, old binary plus activation pin, new binary plus stale local Gastown import, and downgrade/upgrade with stale generated Maintenance directories.
- The canonical repo/path/writer/CI owner for `public-gastown-pins.yaml` and its activation-row behavior manifest.
- A `GastownCity()` invariant proving Core is present when generated Gastown config resolves through production loaders.

**Required changes:**
- Resolve the public Gastown/Core dependency model. Remove the lingering Maintenance import and state whether public Gastown imports Core, relies on host auto-inclusion plus patch layering, or owns dog assets itself.
- Specify dog prompt preservation end to end: prompt-body patching or relocation to named Gastown-owned assets, fragment ownership and lookup rules, duplicate-fragment behavior, duplicate-agent behavior, and rendered prompt tests.
- Make retired-pack handling enforceable before config loading. `gc doctor --fix` must remove or rewrite `[imports.maintenance]` and stale local references, and production loading must reject retired generated pack names even if directories still contain `pack.toml`.
- Add explicit public-Gastown route rules for every active asset that targets the Core maintenance worker, and test default, renamed, and omitted maintenance-worker fixtures.
- Add TOML defaults, step descriptions, active scripts, orders, commands, prompt fragments, and public-pack prose to role-token and retired-Maintenance scanner scope.
- Replace any Pack Ownership wording that permits behavior tests to become wiring-only tests with a strict migration contract: public-pack behavior tests must move to or be recreated in `gascity-packs/gastown`, and Gas City may retain only additional init/import wiring tests after the generated migration map proves equivalent public coverage or approved removals.
- Add public-pack active-asset and test ledgers to the pin artifact, tied to the exact checkout/cache path used for install.
- Specify ownership of `session_live` hooks and each dog `agent.toml` field, including whether `fallback = true` is a Core default.
- Split the release compatibility matrix by pin phase and document activation as a one-way boundary unless old-binary compatibility is proven.
- Clarify `legacyPublicPackForSource` behavior for `examples/gastown/packs/*` paths and add a stale generated Maintenance directory advisory or cleanup policy.
- Correct the asset-map facts for `prune-branches` paths and remove or annotate nonexistent `mol-polecat-report.toml`.
