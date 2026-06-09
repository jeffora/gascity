# Design Review Synthesis

## Overall Verdict: block

All ten persona syntheses returned `block`, so the global verdict is `block` by worst-verdict-wins. The design is still aimed at the right outcome - required Core as host infrastructure, public Gastown as user configuration, and retired Maintenance - but the current plan leaves behavior preservation, required-Core loading, doctor safety, role neutrality, registry/cache migration, runtime state, docs, and slice gates insufficiently specified for implementation.

## Consensus Strengths
- Multiple personas agreed that the target architecture is correct: Core should become the required generated system pack, Gastown should move to a public pack, and standalone Maintenance should retire.
- Reviewers consistently found useful direction in the behavior inventory, public-pack compatibility gate, doctor safety language, role-token scanning, and green-slice rollout goals.
- The design has started naming the right classes of risk: host-required Core, stale generated directories, public Gastown pinning, no-Core test isolation, operator documentation, and behavior-oriented test coverage.
- Several reviewers noted that strict review mode and cross-repo proof are the right release posture for this migration.

## Critical Findings

### [Blocker] Behavior preservation is still not executable or bidirectional
**Sources:** Nadia Volkov / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / Claude, Codex, DeepSeek, Gemini; Tomas Park / Claude, Codex, DeepSeek; Yuki Hayashi / Claude, Codex, DeepSeek V4 Flash
**Issue:** The design still relies on selected inventory rows and selected old tests rather than a source-derived inventory of every behavior-bearing asset, helper, trigger, detector, requester, route, prompt fragment, state path, and notification target. It also allows semantic deltas to be recorded without a machine-enforced approval rule, so behavior can be weakened in place while tests continue to pass.
**Required change:** Define a source-discovery behavior manifest with one required row per moved, split, generalized, retired, or helper-dependent behavior. Each row must prove old witness, new public-pack witness, immutable landing commit, consuming Gas City pin, semantic-equivalence assertion, and explicit approved-removal or approved-delta record.

### [Blocker] Required-Core loading can still be bypassed or falsely proven
**Sources:** Elias Sato / Claude, Codex, DeepSeek, Gemini; Marcus Driscoll / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / DeepSeek, Gemini; Tomas Park / DeepSeek
**Issue:** The proposed Core provenance assertion is not yet a safe identity contract. A path or name check can miss the validated system Core, accept a user/imported pack named `core`, or fail when Core has no effective keys. Controller reload and several production `cmd/gc` paths still use include-building or lower-level `config.Load*` paths that cannot assert post-load Core participation.
**Required change:** Add two required production checks: manifest/content integrity for missing, corrupt, stale, or partial Core, and typed required-system-pack participation after normal config resolution. Enforce production loader boundaries with a seeded `cmd/gc` scanner/allowlist and hard errors for required system-pack name collisions.

### [Blocker] Doctor and import-state fixes are not operator-safe
**Sources:** Sofia Khoury / Claude, Codex, DeepSeek V4 Flash; Felix Moreau / Claude, Codex, DeepSeek, Gemini; Yuki Hayashi / Claude, Codex, DeepSeek V4 Flash; Marcus Driscoll / DeepSeek V4 Flash
**Issue:** The legacy Gastown rewrite path can mutate `pack.toml` or `city.toml` before proving the public source is reachable, immutable, lockable, and installable. Automatic Core/Maintenance/Gastown import removal does not define generated-unmodified provenance, whole-file TOML rewrites can drop operator content, and mutating `gc doctor --fix` lacks a controller-active and concurrent-writer policy.
**Required change:** Rewrite the doctor fix contract around preflight-before-mutation, scoped TOML preservation or refusal, byte-identical no-op on healthy cities, failure-atomic multi-file behavior, custom/fork detection, explicit concurrency policy, and final Core provenance revalidation.

### [Blocker] Core role neutrality and SDK self-sufficiency remain unresolved
**Sources:** Ingrid Kovac / Claude, Codex, DeepSeek V4 Flash, Gemini-labeled re-review; Nadia Volkov / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / Claude, Codex, DeepSeek, Gemini; Felix Moreau / DeepSeek, Gemini
**Issue:** Reviewers found Go role literals and role-derived control paths, including tmux role theme APIs, default scaffolding, warmup mail defaults, prompt fallbacks, formula-name heuristics, and API agent-kind classification. Core-bound formulas, scripts, prompts, orders, metadata, and overlays still carry or depend on Gastown roles such as `mayor`, `deacon`, `witness`, `polecat`, and `dog`.
**Required change:** Add a Go and asset role-neutrality contract. Every role-bearing Go path and Core asset must be migrated, parameterized, moved to Gastown, or explicitly allowlisted with justification. Define whether `dog` is a Core configurable maintenance worker, a user-supplied role, or Gastown-only, and add Core-only tests with the maintenance worker renamed and absent.

### [Blocker] Public Gastown pin, registry, and synthetic cache semantics are inconsistent
**Sources:** Yuki Hayashi / Claude, Codex, DeepSeek V4 Flash; Marcus Driscoll / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / DeepSeek, Gemini; Tomas Park / Codex, DeepSeek
**Issue:** The rollout can move or remove Gas City behavior before `PublicGastownPackVersion` points at the replacement commit. Slice 6 combines alias removal, registry changes, pin bumping, synthetic cache rejection, and ordinary remote-pack materialization without proving the new pin, new Core, and cache behavior together. Existing bundled synthetic caches, old `packs.lock` entries, and offline upgrades have no chosen behavior.
**Required change:** Choose and test the cache and pin migration strategy before registry removal. Specify `RepoCacheKey`, synthetic namespace/hash behavior, stale-cache handling, public-source alias retirement, offline-vs-network fresh init, and ordinary remote-pack install for public Gastown, with negative tests at every old `IsSource` and synthetic fallback call site.

### [Blocker] Maintenance retirement runtime behavior and host-Core dependency are not decided
**Sources:** Avery McAllister / Claude, Codex, DeepSeek, Gemini; Marcus Driscoll / Claude, Codex, DeepSeek V4 Flash; Sofia Khoury / Codex, DeepSeek V4 Flash; Felix Moreau / Claude, Codex, DeepSeek, Gemini; Tomas Park / Claude, DeepSeek
**Issue:** Required-pack lists, builtin includes, synthetic layouts, stale generated-pack pruning, `publicSubpathForPack`, order ownership, `legacyPublicPackForSource`, runtime state paths, and public Gastown's dependency on host Core are still open. The design does not decide how `.gc/runtime/packs/maintenance`, JSONL archive state, export state, storm ledgers, and Maintenance-owned orders move or remain readable.
**Required change:** Add a Maintenance retirement runtime table covering required packs, includes, public source recognition, stale directories, order ownership, legacy aliases, lock/install behavior, runtime state read/write paths, doctor diagnostics, and public Gastown's host-Core patch/import semantics. Include a zero-duplicate-order assertion for the Core/Maintenance fold.

### [Blocker] Bootstrap Core extraction lacks a safe production and fixture contract
**Sources:** Ritu Raman / Claude, Codex, DeepSeek, Gemini; Avery McAllister / DeepSeek, Gemini; Marcus Driscoll / DeepSeek V4 Flash
**Issue:** Removing production Core from `internal/bootstrap` is directionally correct, but the design does not fully define fixture identity, allowed contents, forbidden production-Core-like paths, post-removal `bootstrapAssets`, or hidden-dependency guards. `GC_BOOTSTRAP=skip` also conflicts with the stated no-environment-variable escape hatch for required Core.
**Required change:** Specify an explicit empty production `fs.FS`, a test-only synthetic fixture under testdata, removal of production `//go:embed packs/**`, hidden-dependency scanners for old bootstrap Core paths, a narrowed or retired `GC_BOOTSTRAP=skip` path, and hook overlay tests against the canonical Core package.

### [Blocker] Documentation and operator DX cannot be made accurate from unresolved semantics
**Sources:** Felix Moreau / Claude, Codex, DeepSeek, Gemini; Avery McAllister / Claude, Codex, DeepSeek, Gemini; Yuki Hayashi / Claude, Codex, DeepSeek V4 Flash
**Issue:** The design names a canonical `docs/reference/system-packs.md` that does not exist or appear in docs navigation, and its docs inventory is not authoritative across Markdown, MDX, generated references, CLI help, examples, scripts, doctor strings, and schema references. More importantly, runtime state, `dog` patching, order-name migration, and retired Maintenance semantics are not decided, so docs and doctor output cannot be truthful.
**Required change:** Create or designate the canonical system-pack reference, define a wording matrix for Core, public Gastown, retired Maintenance, store maintenance, and Dolt maintenance, and add generated inventory/lint/golden gates. Resolve runtime-state, `dog`, order alias/rename, first-run, tutorial, and upgraded-city stories before final docs wording.

### [Blocker] Test slicing and coverage gates are too vague for source deletion
**Sources:** Tomas Park / Claude, Codex, DeepSeek; Nadia Volkov / Claude, Codex, DeepSeek V4 Flash; Marcus Driscoll / Claude, Codex, DeepSeek V4 Flash; Avery McAllister / Claude, Codex, DeepSeek, Gemini
**Issue:** The proposed `test/packcompat` gate is the linchpin but lacks a hermetic contract, exact public pin placement, per-row executed assertions, and slice cadence. The source-deletion matrix covers only selected `examples/gastown` files while sibling example, import-wiring, command/status, doctor, overlay, prompt, tmux, branch pruning, shutdown, Polecat, detector/requester, and review-check behavior can be lost.
**Required change:** Make `test/packcompat` a hermetic per-row behavior gate required before the first dependent slice and at every protected boundary. Expand the test migration matrix to every relevant test file/function and require replacement coverage, retained SDK fixture coverage, or explicit intentional-removal decisions before deletion.

### [Major] Existing tests are still too path-, count-, and name-oriented
**Sources:** Tomas Park / Claude, Codex, DeepSeek; Nadia Volkov / Claude, Codex, DeepSeek V4 Flash; Marcus Driscoll / Codex, DeepSeek V4 Flash; Avery McAllister / Claude, Codex
**Issue:** Tests can pass by proving files, includes, names, or counts changed, without proving formulas compose, orders resolve from the right owner, hooks target correctly, scripts run with pack-relative paths, configured agents load, or doctor fixes preserve user content.
**Required change:** Upgrade acceptance to behavior proof: formula/molecule composition, step counts, hook target resolution, order owner resolution, configured-agent/session loading, prompt/template resolution, pack-relative script execution, doctor idempotency, import-state diagnostics, and fresh-init smoke behavior.

### [Major] Provider pack continuity and required-pack integrity are under-proven
**Sources:** Marcus Driscoll / Claude, Codex, DeepSeek V4 Flash; Elias Sato / DeepSeek, Gemini; Ritu Raman / DeepSeek, Gemini
**Issue:** The design asserts that `bd` and `dolt` continuity will survive Core repair and Maintenance retirement, but does not prove byte/provenance preservation, install locks, provider-dependent include behavior, or formula/order resolution when `dog` and maintenance formulas move. Unexpected files under required system packs may also remain loadable.
**Required change:** Add provider-pack matrix tests, byte-identical manifest/provenance checks, install/lock continuity tests, `dolt` formula/order resolution with Maintenance absent, and either full file-set integrity validation for required packs or compensating proof that unexpected files cannot influence loaded behavior.

### [Major] Cross-pack ownership decisions remain open
**Sources:** Avery McAllister / Claude, Codex, DeepSeek, Gemini; Ingrid Kovac / DeepSeek V4 Flash; Nadia Volkov / Codex, DeepSeek V4 Flash; Tomas Park / Claude, Codex, DeepSeek
**Issue:** Ownership and target-state decisions remain unresolved for `mol-review-quorum`, the Gastown Codex overlay, dog prompt fragments, Polecat formulas, review workflow checks, branch pruning, shutdown dance, named sessions, template fragments, session-live hooks, and role-themed tmux APIs.
**Required change:** Resolve every open ownership item before source moves or deletion, or create explicit blocking tasks. Add ownership and dependency tables for Maintenance orders, Core formulas, Gastown overlays, prompt fragments, scripts, hook overlays, named sessions, and tmux behavior.

### [Minor] Several migration artifacts need finer traceability
**Sources:** Nadia Volkov / Claude, Codex, DeepSeek V4 Flash; Ritu Raman / Gemini; Tomas Park / Claude, Codex, DeepSeek
**Issue:** The design does not yet require row-level old path, trigger, target, state path, semantic delta, new path, new test, public-pack landing commit, and Gas City pin for every migration row. Some slice gate lists also rely on the global fast-test invariant instead of naming example-tree coverage when examples are affected.
**Required change:** Add row-level traceability fields and explicit focused commands for affected examples, docs, registry/cache, loader, doctor, asset move, and public-pack compatibility slices.

## Disagreements
- Some underlying model reviews were less severe than the persona verdicts. Several Claude reviews used `approve-with-risks`, while Codex, DeepSeek, or Gemini blocked. Assessment: keep `block`; the repeated unresolved items affect data preservation, behavior equivalence, release sequencing, and role-neutral infrastructure.
- Reviewers differed on public Gastown offline behavior. One acceptable path is network-only fresh init with explicit diagnostics; another is dual-read or cache migration for existing synthetic fallback. Assessment: the design must make a product decision and test it at every former synthetic source call site.
- Reviewers differed on `dog` ownership. Assessment: a Core-provided maintenance worker can be acceptable only as an explicit configurable Core contract with renamed/absent tests; otherwise the behavior must live entirely in user/Gastown configuration.
- Reviewers proposed different doctor edit mechanisms: scoped TOML edits, mutation-deferred preflight, transactional rollback, or automated-fix refusal. Assessment: any mechanism is acceptable only if it preserves unrelated operator content and is failure-atomic across manifests, locks, and installed packs.
- Reviewers differed on bootstrap fixture drift. Assessment: the test fixture should remain deliberately minimal rather than production-Core-equivalent, but CI must prove it does not reintroduce production Core into bootstrap or drift into a misleading fixture.
- Reviewers differed on how aggressively historical docs should be rewritten. Assessment: operator-facing docs, generated references, command help, examples, scripts, doctor output, and troubleshooting are in scope; historical plans need an explicit allowlist.
- Reviewers differed on duplicate order severity during Maintenance folding. Assessment: treat it as a blocker because duplicate active order definitions can leave an intermediate state non-deterministic while name/count tests pass.

## Missing Evidence
- Exact public `gascity-packs/gastown` replacement commit, `PublicGastownPackVersion`, ordinary remote install proof, and CI command proving the candidate public pack before Gas City source removal.
- Source-derived behavior inventory covering triggers, requesters, detectors, route metadata, notification targets, transport, script branches, prompt fragments, state paths, daemon identity, helper paths, and named sessions.
- Machine-enforced semantic-delta approval for every behavior weakening or removal.
- Typed required-system-pack identity and post-load participation proof, plus manifest/content integrity checks for missing, corrupt, stale, partial, or shadowed Core.
- Complete production `config.Load*` call-site inventory and scanner/allowlist, including controller reload and partial-read exceptions.
- Doctor safety proof for preflight, scoped TOML preservation, custom/fork provenance, air-gapped failures, lock/install failure injection, concurrent mutation, live-controller behavior, and Core materialization ordering.
- Runtime-state migration table for `.gc/runtime/packs/maintenance`, JSONL archives, export state, storm ledgers, `GC_PACK_STATE_DIR`, rollback, doctor checks, and operator commands.
- Core role-token scanner contract across Go, TOML, shell, Markdown prompts, skills, overlays, metadata, template fragments, generated command text, and tests.
- Decision and tests for `dog` ownership, public Gastown patching against host Core, renamed maintenance worker behavior, and absent maintenance worker behavior.
- Synthetic cache and public source transition plan for `RepoCacheKey`, `SyntheticContentHash`, retired aliases, existing locks, stale bundled caches, offline fallback, and negative validation.
- Maintenance retirement runtime table for required packs, includes, order ownership, stale generated directories, public source recognition, and moved state paths.
- Bootstrap fixture contract, empty production filesystem default, `GC_BOOTSTRAP=skip` policy, hidden-dependency guards, and hook overlay proof against canonical Core.
- Complete AC-to-test traceability across Gas City and gascity-packs, including all `examples/gastown` tests and replacement public-pack behavior tests.
- Hermetic `test/packcompat` contract with per-row executed assertions, exact pin cadence, formula/molecule/hook/configured-agent coverage, and required slice placement.
- Current-tree docs/operator inventory, canonical wording matrix, navigable system-pack reference, docs lint/golden gate, and first-run/tutorial/upgraded-city acceptance stories.

## Recommended Changes
1. Define a source-derived behavior manifest and make semantic-equivalence or explicit approved-removal mandatory for every moved, split, generalized, retired, or helper-dependent behavior.
2. Make public Gastown replacement-first a hard gate: exact immutable commit, ordinary install proof, behavior inventory, packcompat CI, and Gas City pin proof before source removal.
3. Add required-Core production invariants: manifest/content integrity, typed system-pack participation after config load, name-collision failure, controller reload coverage, and a `cmd/gc` loader-boundary scanner.
4. Rewrite doctor/import-state migration around preflight-before-mutation, scoped TOML preservation or refusal, generated-unmodified provenance, failure atomicity, concurrency policy, air-gap behavior, and Core repair ordering.
5. Resolve Core role neutrality and `dog` ownership, then add Go and asset scanners plus Core-only tests with the maintenance worker renamed and absent.
6. Specify the registry/public-source/synthetic-cache transition before alias removal or pin bump, including old locks, offline behavior, stale cache rejection, ordinary remote install, and negative call-site tests.
7. Add a Maintenance retirement runtime table and make the Core/Maintenance fold atomic, with zero duplicate active orders and explicit legacy state handling.
8. Complete the bootstrap extraction contract: empty production `fs.FS`, test-only fixture, old-path guards, hook overlay migration, and `GC_BOOTSTRAP=skip` containment.
9. Turn `test/packcompat` into a hermetic per-row behavior gate and expand the test migration matrix to every affected example, command, doctor, overlay, prompt, formula, order, and script behavior.
10. Replace path/count/name assertions with behavior assertions for formula composition, hook targets, configured agents, order ownership, prompt resolution, script execution, doctor idempotency, and fresh init.
11. Resolve all cross-pack ownership decisions before move/delete slices and record them in ownership/dependency tables with blocking tasks for any remaining open item.
12. Finish docs/DX design only after runtime-state, `dog`, order naming, and public-source semantics are decided; then add a generated inventory, wording matrix, system-pack reference, and docs/golden gates.
