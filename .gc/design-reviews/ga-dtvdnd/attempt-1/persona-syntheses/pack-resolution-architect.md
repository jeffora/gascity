# Priya Menon

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Info] The requirements establish the right canonical-Core direction. All sources recognize release-bundled Core from `internal/packs/core` as the single required system layer, with `internal/bootstrap/packs/core` demoted to migration-input-only and no user, city, rig, cache, overlay, or transitive import allowed to satisfy required Core by name alone.
- [Info] Legacy import retirement is fail-closed rather than fallback-based. Claude and Codex agree that implicit Maintenance, in-tree `examples/gastown`, stale system-pack copies, and synthetic public Gastown aliases must not silently satisfy runtime behavior; DeepSeek agrees that `gc init --template gastown` should import public `gascity-packs/gastown` at an immutable pin.
- [Info] AC3 names the right resolution surfaces: required Core, provider-conditioned `bd`/`dolt`, explicit Gastown imports, root/city/rig imports, locks, caches, overlays, stale materialized copies, same-named assets, synthetic aliases, missing Core, and transitive diamond conflicts.
- [Major] The `bd`/`dolt` provider-pack policy is still under-specified at the requirements level. Claude finds no explicit activation condition, mutual-exclusion or co-activation rule, default non-Gastown behavior, or healthy no-conflict precedence. Codex treats these as high-risk matrix rows. DeepSeek says the implementation plan/matrix resolves them, but that is not enough if the requirements claim `Open Questions: None`.
- [Major] Cross-surface consistency needs an explicit acceptance criterion. The requirements imply a shared condition-code registry and source-attribution model, but Claude and Codex both stress that init, doctor, import-state, CLI load, and runtime resolution must agree for the same concrete city state.
- [Major] Bootstrap-only commands must be protected from eager normal pack resolution. DeepSeek identifies a real catch-22: if Cobra startup or global pre-run code resolves packs before dispatching `gc doctor`, `gc import-state`, or `gc version`, the diagnostic path can fail before it can explain or repair missing Core.
- [Major] Collision detection must happen before any parser or filesystem merge can erase evidence. DeepSeek flags that post-resolution duplicate-behavior checks are insufficient if TOML ingestion, directory overlay, or path merge behavior silently overwrites a formula, order, prompt, script, or fragment before validators see both sources.
- [Major] Transitive diamond conflicts require an explicit operator policy. DeepSeek recommends top-level override pins; Claude and Codex emphasize fail-closed deterministic resolution. The requirements should choose and document the policy rather than leaving implementation to infer it.
- [Major] Stale legacy directories are not harmless if shell-side scripts can still reach them by relative or hardcoded paths. DeepSeek's stale-directory shadow-execution risk means doctor repair should isolate or clearly retire old `.gc/system/packs/maintenance`, `.gc/system/packs/gastown`, or runtime-pack copies, not merely ignore them in Go-side discovery.
- [Minor] Bundled support-pack provenance is weaker than Core provenance. Claude notes that `bd`/`dolt` still appear tied to `examples/bd` and `examples/dolt` without a stated source-authority contract.
- [Minor] The AC2 dev/test Core-less escape hatch needs an enforcement witness proving it cannot drive production CLI, doctor, controller, runtime, session, dispatch, formula expansion, or city-state mutation.
- [Minor] Fresh offline `gc init --template gastown` with no seeded cache needs its own scenario now that the synthetic alias fallback is being retired.
- [Minor] Concurrent cache promotion needs a synchronization contract beyond process-unique staging and atomic rename when multiple processes promote the same public pack commit.

**Disagreements:**
- Claude says `bd`/`dolt` cardinality and baseline precedence are unresolved product policy. DeepSeek says the implementation plan and `pack-resolution-matrix.yaml` already define deterministic precedence. Codex says the requirement is broad enough but the future matrix must include executable rows. My assessment: the requirements must carry the product policy or explicitly list it as open; implementation-plan detail alone does not satisfy a requirements approval gate.
- Codex says no changes are required before requirements approval. Claude requires requirements tightening, and DeepSeek lists several hardening changes for implementation slices. My assessment: approve with risks, but carry the hardening items as required lane changes before implementation tasks are decomposed.
- DeepSeek recommends top-level overrides for diamond conflicts, while the other reviews emphasize fail-closed conflict handling. My assessment: either policy can be acceptable, but silent resolver choice is not. The requirements need an explicit top-level override contract or an explicit fail-closed/import-chain diagnostic contract.
- DeepSeek treats stale directory isolation as a doctor repair behavior. Claude and Codex focus on active resolver provenance and do not mention shell-side shadow execution. My assessment: the shadow-execution risk is credible because the migration includes scripts; it belongs in this lane's stale-copy acceptance criteria.

**Missing evidence:**
- Explicit `bd`/`dolt` activation conditions, mutual-exclusion or co-activation rules, default-city behavior, and healthy no-conflict layer precedence.
- A cross-surface golden set proving init, doctor, import-state, CLI load, and runtime resolution report the same Core identity, source attribution, and stable condition code for the same missing, duplicate, stale, or shadowed Core city.
- The exact command classes that are bootstrap-safe without normal Core resolution, and proof that normal behavior-changing commands are blocked without required Core.
- Source-authority and provenance expectations for bundled support packs `bd` and `dolt`.
- A production-boundary proof for the AC2 dev/test Core-less escape hatch.
- Fresh offline Gastown init behavior with no seeded cache after removing the synthetic public alias.
- Ingestion-level duplicate and static asset/path collision fixtures that prove collisions are caught before TOML or filesystem merges can hide them.
- Cache promotion evidence for concurrent processes promoting the same public pack commit.
- A documented transitive diamond conflict policy: explicit top-level override with attribution, or fail closed with import-chain diagnostics.
- A stale-directory repair/isolation policy that prevents retired Maintenance or in-tree Gastown scripts from being executed by relative paths.

**Required changes:**
- State the `bd`/`dolt` activation condition, mutual-exclusivity or co-activation rule, default-city cardinality, and healthy no-conflict layer precedence in AC3, and reconcile that with the support classification document's bd-owns-dolt proposal.
- Add a testable cross-surface consistency criterion requiring init, doctor, import-state, CLI load, and runtime resolution to agree on Core identity, source attribution, and stable condition code for the same city state.
- Define the bootstrap-safe command set and require those commands to bypass normal pack resolution until command dispatch has selected the bootstrap path; behavior-changing commands must still fail closed without required Core.
- State the source-authority and provenance expectation for bundled support packs `bd` and `dolt`, even if they remain optional bundled packs rather than required system layers.
- Add an enforcement witness for the AC2 dev/test escape hatch showing it is unreachable from the production `gc` binary and cannot mutate city state or drive runtime behavior.
- Add an Example Mapping edge case for fresh offline `gc init --template gastown` with no seeded cache: fail closed with cache-seeding guidance, not the retired synthetic alias.
- Require raw ingestion and filesystem-path collision checks for formulas, orders, prompts, scripts, fragments, and same-named static assets before any merge or overlay can discard duplicate evidence.
- Define cache-promotion synchronization for concurrent promotion of the same public pack commit.
- Decide and document the transitive diamond conflict policy: explicit top-level override with full source attribution, or fail closed with actionable import-chain diagnostics.
- Require `gc doctor --fix` or equivalent repair to isolate stale legacy pack directories so retired scripts cannot still execute through hardcoded or relative paths.
