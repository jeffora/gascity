# Design Review Synthesis

## Overall Verdict: block

All ten persona syntheses reached `block`, so the global verdict is `block` by worst-verdict-wins. The design has the right goal, but it does not yet make the Core/Gastown/Maintenance split executable: behavior preservation, required-Core loading, doctor safety, role neutrality, runtime pack/cache migration, and release sequencing all need stronger gates before implementation can safely proceed.

## Consensus Strengths
- Multiple personas agreed the intended direction is correct: Core should become the required generated system pack, Gastown should move to an explicit public pack, and the standalone Maintenance pack should retire.
- Reviewers found useful starting points in the design's file inventory, public-pack pin concept, stale system-pack directory caution, and goal of keeping low-level tests able to exercise no-Core config loading.
- The proposed use of behavior-oriented smoke coverage, doctor checks, and role-name guards is directionally sound, but the reviewers consistently found those gates under-specified or too easy to bypass.

## Critical Findings

### [Blocker] Behavior preservation is not enforced before Gas City removes in-tree behavior
**Sources:** Nadia Volkov / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / Claude, Codex, DeepSeek V4 Flash; Tomas Park / Claude, Codex, DeepSeek V4 Flash; Yuki Hayashi / Claude, Codex, DeepSeek V4 Flash
**Issue:** The design relies on an out-of-repo public Gastown replacement and a mostly file-oriented inventory, but it does not require a pinned public-pack commit, behavior-edge manifest, replacement tests, and Gas City-side CI proof before removing or generalizing current Gastown and Maintenance behavior. Existing Gas City tests such as `examples/gastown/gastown_test.go` and `examples/gastown/maintenance_scripts_test.go` can be downgraded to wiring checks without a test-by-test replacement matrix.
**Required change:** Make public Gastown replacement-first a blocking prerequisite. Require an exact repo/PR/commit, `PublicGastownPackVersion` value, behavior inventory path, CI command, and required artifacts; add a Gas City gate that fetches the exact pinned commit and proves each trigger/action/target behavior edge and migrated test still passes.

### [Blocker] Required-Core loading can still be bypassed in production paths
**Sources:** Elias Sato / Claude, Codex, DeepSeek V4 Flash; Marcus Driscoll / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / DeepSeek V4 Flash; Tomas Park / DeepSeek V4 Flash
**Issue:** Materializing `.gc/system/packs/core` does not prove normal config resolution included Core. The design depends on an audit of direct `config.LoadWithIncludes` usage, but production `cmd/gc` code also has broader `config.Load*` bypasses and helpers such as no-refresh loading paths. Doctor is not a substitute for the normal command path failing closed.
**Required change:** Add a production wrapper invariant that asserts required Core participated in resolved config after loading. Add a CI scanner modeled on `TestGCNonTestFilesStayOnWorkerBoundary` that rejects direct production `config.Load` and `config.LoadWithIncludes` calls outside a documented allowlist with focused tests.

### [Blocker] Doctor and import-state fixes are not operator-safe
**Sources:** Sofia Khoury / Claude, Codex, DeepSeek V4 Flash; Felix Moreau / Codex, DeepSeek V4 Flash; Yuki Hayashi / Claude, Codex, DeepSeek V4 Flash; Marcus Driscoll / DeepSeek V4 Flash
**Issue:** The migration can rewrite local Gastown imports before proving the public pack is reachable, immutable, installable, and lockable. Automated Maintenance/Core import removals can affect operator-owned forks or edited system-pack directories, and whole-file TOML re-encoding can drop comments, ordering, unknown fields, or custom tables. Runtime state under `.gc/runtime/packs/maintenance` is also unresolved, so doctor output cannot be truthful yet.
**Required change:** Define a doctor fix safety contract: preflight before mutation, scoped TOML edits or manual fallback, failure-atomic multi-file behavior, byte-identical no-op on healthy cities, custom/fork provenance detection, final Core provenance revalidation, and explicit runtime-state migration or diagnostic policy.

### [Blocker] Core role neutrality and SDK self-sufficiency remain unresolved
**Sources:** Ingrid Kovac / Claude, Codex, DeepSeek V4 Flash; Nadia Volkov / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / Claude, Codex, DeepSeek V4 Flash; Felix Moreau / DeepSeek V4 Flash
**Issue:** Core-bound formulas, prompts, scripts, skills, overlays, metadata, and Go-adjacent runtime surfaces still appear to encode Gastown concepts such as `mayor`, `deacon`, `witness`, `refinery`, `polecat`, and `dog`. The design does not define whether `dog` is a Core-owned configurable maintenance agent, a user-supplied role, or Gastown-only behavior, so removed/renamed maintenance-agent acceptance criteria are untestable.
**Required change:** Add a per-operation Core maintenance/notification contract. Every Core mail, nudge, route, requester, detector, escalation, script, and prompt target must be parameterized, resolved from user configuration, moved to Gastown, or removed. Add a token/path/field-scoped role-name scanner across Core assets and tests proving Core-only cities work with the maintenance agent renamed or absent.

### [Blocker] Pack registry, synthetic cache, and public Gastown pin semantics are inconsistent
**Sources:** Yuki Hayashi / Claude, Codex, DeepSeek V4 Flash; Marcus Driscoll / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / DeepSeek V4 Flash; Tomas Park / DeepSeek V4 Flash
**Issue:** The design does not choose a coherent strategy for removing public Gastown and Maintenance synthetic aliases while updating `PublicGastownPackVersion`. A new public `sha:` pin might still be satisfied by embedded old synthetic content, while alias removal may retire offline `gc init --template gastown` behavior and shift cache namespaces. Removing packs from the bundled layout may also invalidate unrelated synthetic caches through `SyntheticContentHash()`.
**Required change:** Specify and test the cache migration strategy before pin bump or alias removal. Define `RepoCacheKey`, synthetic namespace/content-hash behavior, stale cache rejection, offline-vs-network init behavior, and ordinary remote-pack materialization for public Gastown. Add direct negative tests for retired sources in `IsSource`, `NameForSource`, lock/install paths, and legacy public source handling.

### [Blocker] Maintenance retirement mechanics are runtime decisions, not cleanup details
**Sources:** Avery McAllister / Claude, Codex, DeepSeek V4 Flash; Marcus Driscoll / Claude, Codex, DeepSeek V4 Flash; Ingrid Kovac / DeepSeek V4 Flash; Felix Moreau / Codex, DeepSeek V4 Flash
**Issue:** `requiredBuiltinPackNames`, `builtinPackIncludes`, `publicSubpathForPack("maintenance")`, `MaterializeBuiltinPacks`, stale generated-pack pruning, import-state doctor logic, order-name migration, and runtime state paths all determine whether existing cities boot correctly after Maintenance retires. The design does not yet state how stale `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories are ignored, diagnosed, preserved, migrated, or pruned.
**Required change:** Add a Maintenance retirement runtime table covering required packs, includes, public source recognition, order ownership, state paths, stale generated directories, doctor behavior, and synthetic cache validation. Tests must prove Core plus provider packs are required, Maintenance is not auto-included, stale directories with custom files survive, and moved orders/scripts resolve from their new owner.

### [Blocker] Bootstrap Core extraction lacks a safe fixture and hidden-dependency contract
**Sources:** Ritu Raman / Claude, Codex, Gemini; Avery McAllister / DeepSeek V4 Flash; Marcus Driscoll / DeepSeek V4 Flash
**Issue:** The design moves production Core away from `internal/bootstrap`, but does not fully specify removal of `//go:embed packs/**`, the replacement `bootstrapAssets` default, the synthetic test fixture contract, or guards against old `internal/bootstrap/packs/core` dependencies. Existing tests can still accidentally hash, copy, or validate the production Core tree through bootstrap fixture paths.
**Required change:** Commit to a test-only synthetic fixture under `internal/bootstrap/testdata/packs/core`, keep it out of the production binary, define its minimal manifest contract, set production bootstrap assets to an explicit empty filesystem, and add source guards for lingering embeds, imports, and `AssetDir: "packs/core"` test paths. Move hook overlay reads to the canonical Core package and prove hook installation still works.

### [Blocker] Rollout and test slicing do not keep every intermediate state green
**Sources:** Tomas Park / Claude, Codex, DeepSeek V4 Flash; Yuki Hayashi / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / Claude, Codex, DeepSeek V4 Flash; Sofia Khoury / Claude, Codex, DeepSeek V4 Flash
**Issue:** The plan is phased by repository rather than by independently deployable, test-green slices. It does not say whether asset moves, registry changes, Core path moves, Maintenance folding, public pin updates, doctor changes, docs, examples, and source deletion are atomic or separately deployable. Version-skew and rollback states involving old/new binaries, old/new public Gastown commits, existing locks, and doctor-mutated manifests are untested.
**Required change:** Add a slice-by-slice implementation plan with exact focused tests and broad gates before each next slice. Include a release compatibility matrix for old/new Gas City and old/new public Gastown, candidate-pack verification before pinning, rollback/downgrade policy, and a green-commit invariant requiring focused suites plus `make test-fast-parallel` to remain passing at each step.

### [Blocker] Documentation and operator DX cannot be made consistent from the current design
**Sources:** Felix Moreau / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / Claude, Codex, DeepSeek V4 Flash; Yuki Hayashi / Claude, Codex, DeepSeek V4 Flash
**Issue:** The docs inventory names a missing `docs/reference/system-packs.md` and omits current operator-facing references, examples, schema docs, CLI references, troubleshooting paths, script comments, and doctor strings. There is no canonical wording matrix for Core, public Gastown, retired Maintenance, Core maintenance-agent behavior, store maintenance, and Dolt maintenance. Runtime-state and order-name decisions are not resolved, so docs and doctor guidance cannot be accurate yet.
**Required change:** Replace the representative docs list with a generated current-tree inventory and a docs/golden/lint gate. Create or name the canonical system-pack reference doc, define exact operator wording, specify first-run and tutorial behavior for minimal and Gastown-template cities, and document runtime-state, stale-directory, dog-patch, and order-name migration semantics.

### [Major] Behavior tests are still too path-, name-, and count-oriented
**Sources:** Tomas Park / Claude, Codex, DeepSeek V4 Flash; Nadia Volkov / Claude, Codex, DeepSeek V4 Flash; Marcus Driscoll / Codex, DeepSeek V4 Flash; Avery McAllister / Claude, Codex
**Issue:** Several proposed tests can pass by proving files exist, names appear, includes changed, or orders parse, without proving formulas compose into molecules, hooks resolve, configured agents and pools load, doctor fixes are idempotent, or moved scripts run with correct pack-relative paths.
**Required change:** Upgrade test acceptance to behavior proof: formula composition, expected step counts, hook target resolution, configured agent/session loading, hook installation, pack-relative script paths, doctor fix idempotency, import-state diagnostics, and fresh init smoke behavior.

### [Major] Provider pack continuity and Core file integrity need stronger proof
**Sources:** Marcus Driscoll / Claude, Codex, DeepSeek V4 Flash; Elias Sato / DeepSeek V4 Flash
**Issue:** Core repair may re-materialize bd and dolt alongside Core, but the design does not prove bd/dolt bytes and provenance remain identical. Required Core validation is ambiguous about unexpected files, which could leave rogue formulas, orders, scripts, or overlays loadable.
**Required change:** Add tampered-Core repair tests, bd/dolt byte-identical preservation tests, provider matrix tests for bd and dolt resolution, and either full file-set integrity validation for Core or compensating tests proving unexpected files cannot influence loaded behavior.

### [Major] Cross-pack ownership decisions remain open
**Sources:** Avery McAllister / Claude, Codex, DeepSeek V4 Flash; Ingrid Kovac / DeepSeek V4 Flash; Nadia Volkov / Codex, DeepSeek V4 Flash; Tomas Park / Claude, Codex, DeepSeek V4 Flash
**Issue:** The design leaves unresolved ownership for `mol-review-quorum`, the Gastown Codex overlay, dog prompt fragments, Polecat parent/child formulas, review workflow checks, branch pruning, shutdown-dance behavior, and hardcoded role-theme APIs. These items determine both package boundaries and tests.
**Required change:** Resolve every `review` or `split` item before source deletion or create explicit blocking tasks before implementation. Add ownership and dependency tables for Maintenance orders, Core formulas, Gastown overlays, prompt fragments, scripts, and tmux theme behavior.

### [Minor] Several migration artifacts need finer traceability
**Sources:** Nadia Volkov / Claude, Codex, DeepSeek V4 Flash; Ritu Raman / Gemini; Tomas Park / Claude, Codex, DeepSeek V4 Flash
**Issue:** The inventory lacks row-level landing commits, old-test to new-test mapping, fixture identity decisions, and explicit inclusion of `go test ./examples/...` for example-facing changes.
**Required change:** Add row-level old path, trigger, target, semantic delta, new path, new test, landing commit, and Gas City pin data. Decide whether the bootstrap fixture uses `core` or `test-core`, and include example-tree gates where examples are affected.

## Disagreements
- Some underlying model reviews were less severe than the persona synthesis verdicts: Claude returned `approve-with-risks` in several personas, while Codex, DeepSeek V4 Flash, or Gemini blocked. Assessment: keep `block`; the combined evidence shows unresolved runtime and release-gating decisions, not polish items.
- Reviewers disagreed on synthetic public Gastown behavior. One view accepts network-only fresh init if documented, another preserves offline fallback, and another emphasizes cache namespace/hash blast radius. Assessment: the design must make a product decision and then test it; implicit alias removal is not acceptable.
- Reviewers differed on whether a Core-owned `dog` is acceptable. Assessment: it can be acceptable only as an explicit, configurable, role-neutral Core contract with renamed/disabled tests; otherwise dog behavior must move to user/Gastown configuration.
- Reviewers differed on TOML edit strictness for doctor fixes. Assessment: automated `gc doctor --fix` must preserve unrelated manifest content or refuse the fix with manual guidance; dry-run alone is not enough.
- Reviewers differed on fixture drift. Assessment: the bootstrap fixture should not track production Core byte-for-byte, but CI must prove it remains deliberately minimal and does not reintroduce production Core assets into bootstrap.
- Reviewers differed on how much historical documentation must be rewritten. Assessment: operator-facing docs, generated references, examples, command help, doctor output, and troubleshooting are in scope; historical docs need an explicit allowlist.

## Missing Evidence
- Exact public `gascity-packs/gastown` prerequisite: repo, PR or branch, immutable commit, `PublicGastownPackVersion`, behavior inventory path, CI command, and required artifacts.
- A behavior-edge preservation manifest with one row per trigger, requester, detector, route metadata, notification target, script branch, prompt fragment, old test, new test, public-pack landing commit, and Gas City pin.
- A Gas City CI gate that fetches the exact pinned public Gastown commit and proves replacement behavior before source removal.
- A production config loading boundary guard and post-load Core inclusion assertion used by normal commands, controller reload, and doctor diagnostics.
- A full inventory and disposition of production `config.Load*` call sites, including allowed partial-read exceptions and tests.
- Doctor failure-atomicity proof for manifests, lockfiles, and installed pack directories; golden TOML preservation tests; and custom/fork provenance rules.
- Runtime-state policy for `.gc/runtime/packs/maintenance` JSONL archives, export state, storm ledgers, and `jsonl_archive_doctor_check.go`.
- A Core role-token inventory and scanner contract covering TOML, shell, Markdown prompts, skills, overlays, doctor assets, agent metadata, template fragments, variable defaults, and embedded command text.
- A decision on `dog` ownership and tests for Core-only cities when the default maintenance agent is renamed or absent.
- Version-skew and rollback matrix for old/new binaries, old/new public Gastown commits, existing locks, Core-only host state, and doctor-mutated manifests.
- Synthetic cache strategy covering `RepoCacheKey`, `SyntheticCacheNamespace`, `SyntheticContentHash`, stale cache rejection, public alias removal, and offline-vs-network `gc init --template gastown`.
- Maintenance retirement runtime table for required packs, includes, public source recognition, order ownership, stale system-pack directories, and moved state paths.
- Bootstrap fixture contract, production empty filesystem default, hidden-dependency guard, and hook overlay proof against the new canonical Core package.
- AC-to-test traceability across Gas City and gascity-packs, including replacement coverage for `examples/gastown/gastown_test.go`, `examples/gastown/maintenance_scripts_test.go`, packlint/parser tests, and Maintenance auto-inclusion tests.
- Current-tree docs/operator inventory, canonical wording matrix, canonical system-pack reference doc, docs lint/golden gate, and first-run/tutorial transition story.

## Recommended Changes
1. Add public Gastown replacement-first gating: exact commit, immutable pin, behavior manifest, required artifacts, and Gas City CI proof against the pinned public pack before any in-tree source removal.
2. Define the required-Core production loading invariant and enforce it with a post-load assertion plus a `cmd/gc` scanner/allowlist for direct `config.Load*` usage.
3. Rewrite the doctor/import-state design around preflight-before-mutate, scoped TOML edits, failure atomicity, custom/fork preservation, runtime-state policy, and final Core provenance revalidation.
4. Resolve Core role neutrality: per-operation target contracts, explicit `dog` ownership, token/path/field-scoped Core asset scanning, and Core-only renamed/absent maintenance-agent tests.
5. Specify the pack registry, public source, and synthetic cache migration strategy before changing `PublicGastownPackVersion` or removing aliases.
6. Add a Maintenance retirement runtime table and tests for required packs, includes, stale generated directories, order ownership, state paths, provider pack continuity, and legacy doctor handling.
7. Commit to the bootstrap Option 1 fixture: test-only synthetic fixture, no production `//go:embed packs/**`, explicit empty production `fs.FS`, hidden-dependency guards, and hook overlay tests.
8. Replace repository-level rollout phases with green vertical slices, a cross-repo candidate-pack verification gate, version-skew/rollback matrix, and per-slice focused and broad test commands.
9. Upgrade test strategy from path/count/name assertions to behavior proof for formulas, orders, hooks, agents, pools, scripts, doctor fixes, import-state diagnostics, and fresh init.
10. Complete the docs/DX design: current-tree inventory, canonical wording, runtime-state and stale-directory guidance, first-run/tutorial story, generated reference/doc lints, and explicit historical-doc allowlists.
