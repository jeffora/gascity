# Design Review Synthesis

## Overall Verdict: block

Nine of ten persona syntheses returned `block`; the remaining persona returned `approve-with-risks`, so worst-verdict-wins yields a global `block`. The design has improved in staging, required-Core thinking, and public-pack intent, but reviewers still found unresolved contracts that can brick upgraded cities, silently lose Gastown behavior, or let role-specific logic remain in the SDK.

## Consensus Strengths
- Reviewers repeatedly praised the move toward a staged multi-slice rollout instead of a flag-day migration.
- Multiple personas approved the direction of making Core a required system pack with explicit validation rather than relying on incidental imports.
- Reviewers agreed that public Gastown should be pinned to immutable content and tested before in-tree Gastown/Maintenance assets are deleted.
- The design now identifies important risk areas: Maintenance retirement, Core/Gastown split ownership, doctor safety, stale generated directories, and provider pack continuity.
- The proposed behavior inventory, packcompat gate, scanner, and docs-lint concepts are directionally correct; reviewers blocked because they are not yet precise or enforceable enough.

## Critical Findings

### [Blocker] Behavior Preservation Gate Is Not Executable Or Source-Derived
**Sources:** Nadia Volkov; Avery McAllister; Tomas Park; Felix Moreau; Ritu Raman
**Issue:** The design still relies too much on curated inventories, representative tests, or existence checks. Reviewers identified behavior-bearing sources that can be missed: formulas, orders, TOML defaults, prompt fragments, shell helpers, mail/nudge targets, daemon identity, route metadata, named sessions, helper paths, and acceptance tests. A row that says a behavior exists does not prove the original trigger still produces the same observable output through host Core plus public Gastown.
**Required change:** Replace the curated inventory with a source-discovery-based manifest and an executable `test/packcompat` gate. Every moved, split, generalized, helper-dependent, or retired behavior-bearing asset must map to a preserved behavior test, a Core/Gastown replacement, or an approved intentional-removal row.

### [Blocker] Public Gastown Host-Core Contract Is Unresolved
**Sources:** Avery McAllister; Felix Moreau; Ingrid Kovac; Yuki Hayashi; Tomas Park
**Issue:** The design has not chosen how public Gastown depends on Core. It says Core is a required host system pack, Gastown should not import Core, and Gastown may patch Core's `dog` agent. That leaves `[[patches.agent]]`, prompt fragments, `session_live` hooks, formulas, missing-host diagnostics, and `dog` ownership ambiguous. Existing stale `../maintenance` imports can also make `gc doctor` or normal config loading fail before migration can run.
**Required change:** Add a concrete "Public Gastown host-Core contract" section that chooses explicit Core import, auto-included Core patching, or Gastown-owned replacement assets. Include tests for host Core present, host Core absent, no Maintenance present, `dog` patch resolution, prompt/template resolution, and stale local Gastown/Maintenance import repair.

### [Blocker] Required Core Loading Still Lacks A Complete Identity And Bypass Contract
**Sources:** Elias Sato; Tomas Park; Ritu Raman; Marcus Driscoll
**Issue:** Required Core inclusion is not fully proven by path/name provenance alone. Reviewers require separate fatal checks for materialized pack integrity and resolved-config participation, collision protection when another pack named `core` appears, and migration of direct production `config.Load*` callers through config-returning wrappers. The controller reload path and no-refresh/partial-read paths remain specifically called out.
**Required change:** Define a typed required-system-pack participation contract with content identity, materialized path, resolved layer identity, and collision semantics. Inventory production loader call sites, migrate normal paths to asserting wrappers, define narrow tested partial-read exceptions, and add scanner coverage for raw loader bypasses.

### [Blocker] Doctor And Migration Fixes Are Not Failure-Atomic Or Pre-Resolution Safe
**Sources:** Sofia Khoury; Avery McAllister; Felix Moreau; Tomas Park; Yuki Hayashi
**Issue:** `gc doctor --fix` can only be operator-safe if it can diagnose and repair broken legacy imports before full config resolution fails, and if it does not mutate manifests before reachability, lockfile, install, provenance, and Core materialization preflights succeed. Reviewers also found unresolved concurrency, live-controller, comment-preserving TOML editing, air-gapped failure, local development path, and generated/unmodified provenance concerns.
**Required change:** Specify a scoped diagnostic/pre-resolution load path or raw TOML migration path for stale imports. Make all mutating fixes preflight-first, scoped, idempotent, failure-atomic, provenance-aware, and concurrency-safe, with byte-identical failure tests and preserving TOML golden tests.

### [Blocker] Rollout, Pinning, And Cache Transitions Still Permit Version Skew
**Sources:** Yuki Hayashi; Marcus Driscoll; Tomas Park; Sofia Khoury
**Issue:** Public Gastown can still be treated as bundled synthetic content while a `sha:` pin claims immutable remote content. Removing Gastown/Maintenance from the builtin registry changes cache validation and `All()` behavior, and old public locks or stale synthetic aliases can become low-level fetch failures instead of actionable diagnostics. The no-Maintenance public-pack proof also conflicts with slices where production loading still force-includes Maintenance.
**Required change:** Pair or order the public pin bump with resolver cutover so `PublicGastownPackSource` plus a full commit SHA cannot be satisfied by synthetic materialization. Add a transition matrix for registry, resolver, cache, `requiredBuiltinPackNames`, `publicSubpathForPack`, `legacyPublicPackForSource`, and init defaults, with old-binary/new-pack and new-binary/old-pack gates.

### [Blocker] Bootstrap Extraction Can Break Tests And Leak Production Core Into Fixtures
**Sources:** Ritu Raman; Tomas Park; Marcus Driscoll
**Issue:** Removing `internal/bootstrap/packs/core` is not yet paired with a concrete non-nil empty production filesystem, a safe fixture injection model, `GC_BOOTSTRAP=skip` semantics, `bootstrapManagedImportNames` updates, or a complete old-path inventory. Reviewers identified hardcoded references in `cmd/gc`, `internal/config`, `examples/gastown`, `test/packlint`, `internal/remotesource`, hooks docs, and engineering docs.
**Required change:** Define the production `bootstrapAssets` default, a synthetic test-only fixture contract, and scanner guards against old bootstrap-Core paths and copied production-Core directories. Update `bootstrapManagedImportNames`, the affected tests/docs, and Slice 3 gates so the fast baseline and all impacted packages stay green.

### [Blocker] Role Neutrality Is Claimed Without Resolving Go Role Surfaces
**Sources:** Ingrid Kovac; Nadia Volkov; Avery McAllister; Felix Moreau
**Issue:** The design focuses on Core asset scanning while live Go and pack surfaces still contain role-specific behavior: tmux themes/icons, `DogTheme`, default city scaffolding, warmup defaults, prompt fallback, `internal/sling` formula-name heuristics, `classifyAgentKind`, `dog` defaults, mail/nudge targets, and role-bearing TOML descriptions. Deleting the old role-name guard without a replacement can reduce coverage.
**Required change:** Decide whether Go de-roling is in scope. If it is, add a Go role-token inventory, migration table, replacements, and tests. If it is deferred, remove end-to-end role-neutrality claims from this design and keep asset-level neutrality explicitly bounded.

### [Blocker] Runtime State Migration For Retired Maintenance Is Undecided
**Sources:** Felix Moreau; Sofia Khoury; Tomas Park; Avery McAllister
**Issue:** `.gc/runtime/packs/maintenance` is not just stale naming. Reviewers identified JSONL archive repos, export cursors, push-failure counters, spawn-storm ledgers, order tracking, script state, `GC_PACK_STATE_DIR`, and doctor fallback paths. Without a chosen read/write/migrate/diagnose policy, docs, doctor output, scripts, and rollback guidance cannot be truthful.
**Required change:** Add a runtime-state migration table covering every retired Maintenance path. Choose Core migration, dual-read legacy fallback, advisory diagnostics, or intentional abandonment per state file, and update scripts, doctor checks, docs, tests, and rollback instructions accordingly.

### [Major] Docs And Operator DX Cannot Stabilize Until Core/Gastown Semantics Are Chosen
**Sources:** Felix Moreau; Avery McAllister; Marcus Driscoll
**Issue:** The design proposes canonical wording, docs lint, and a system-pack reference page, but the reference anchor is not concrete and the terminology still varies. Operator-facing examples still contain copy-paste traps around local Gastown, Maintenance imports, deleted paths, JSONL state, `dog`, order names, and public-pack setup.
**Required change:** Create or designate the canonical system-pack reference page, register it, and bind doctor output, docs, tutorials, pack comments, CLI reference, generated docs, and FixHints to one wording matrix after the runtime-state and host-Core contracts are settled.

### [Major] Pack Registry And Provider Continuity Need More Than Path Checks
**Sources:** Marcus Driscoll; Yuki Hayashi; Elias Sato; Ritu Raman
**Issue:** Provider packs (`bd`, `dolt`) and required Core need strict file-set integrity, cache self-heal behavior, prompt-baseline filtering for stale retired directories, and continuity proof for manifests, executable modes, provenance, include ordering, lock/cache behavior, and provider matrix behavior. Simple path/count checks are insufficient.
**Required change:** Add stale-source diagnostics, strict required-pack validation or pruning, prompt-baseline filtering, provider continuity tests, and synthetic-cache regeneration tests. Clarify the fate of `PublicRepository`, `normalizeRepository`, `RepoCacheKey`, and `publicSubpathForPack`.

### [Minor] Persona Artifact Paths Are Inconsistent
**Sources:** Global synthesizer artifact audit
**Issue:** The attempt-4 persona-synthesis beads closed with `gc.attempt=4`, but their `design_review.output_path` metadata points under `.gc/design-reviews/ga-2404qu/attempt-1/persona-syntheses/`, and `.gc/design-reviews/ga-2404qu/attempt-4/persona-syntheses/` is absent. The global synthesis used those recorded paths because all ten persona-synthesis beads closed `pass`, but per-attempt traceability is weak.
**Required change:** Fix the persona-synthesis output path computation so it writes to the current attempt directory and add a workflow test or controller assertion that output paths match `gc.attempt`.

## Disagreements
- Several personas had internal model disagreement, but the global result is not close: nine persona syntheses block and one approves with risks.
- Some reviewers would accept explicit Core imports from public Gastown; others prefer auto-included host Core with patching. The synthesis does not choose the model, but it treats choosing and testing one model as mandatory.
- Reviewers differed on whether offline fresh Gastown initialization must remain supported. The synthesis treats the policy choice itself as required: either support it with a deliberate cache/prepopulation model or reject it with precise diagnostics and tests.
- The `dog` worker can be a Core-provided configurable housekeeping pool, a Gastown convention, or a public Gastown-owned utility worker. The current design cannot proceed while it implies more than one of these.
- Some reviewers classify Go de-roling outside the asset migration; others treat it as in scope because the design claims role neutrality. The synthesis requires the design to either include that work or narrow its claim.

## Missing Evidence
- Machine-readable, source-derived behavior manifest with CI validation and executable witnesses for every moved, split, generalized, retired, or helper-dependent behavior.
- Hermetic `test/packcompat` design and slice cadence proving host Core plus public Gastown behavior with no Maintenance through production config loading.
- Complete test-function migration table for `examples/gastown/*_test.go`, `maintenance_scripts_test.go`, Core-owned script tests, Gastown-owned tests, retained fixtures, and approved removals.
- Required Core identity model, collision behavior, direct-loader call-site inventory, partial-read allowlist, and controller reload tests.
- Pre-resolution or lenient doctor path for stale local Gastown/Maintenance imports, plus failure-atomic TOML/lockfile/install/Core-materialization tests.
- Public pin and resolver proof showing remote fetch/materialization of a full immutable commit SHA rather than synthetic bundled bytes.
- Runtime-state table for JSONL archives, export state, push-failure counters, spawn-storm ledgers, order tracking, `GC_PACK_STATE_DIR`, legacy reads, current writes, doctor behavior, and rollback.
- Scanner contracts for Core assets, Go role-name surfaces, retired Maintenance paths, prompt fragments, TOML role defaults, script helper paths, generated docs, and allowed historical references.
- Concrete docs/reference anchor and terminology matrix for required Core, provider system packs, public Gastown, retired Maintenance, stale generated directories, `dog`, store/Dolt maintenance, and preserved event names.
- Current-attempt persona synthesis artifacts under the attempt-4 directory.

## Recommended Changes
1. Define the public Gastown host-Core contract, `dog` policy, stale-import repair path, and runtime-state migration table before changing more rollout slices.
2. Replace curated preservation prose with a generated behavior manifest plus hermetic executable `test/packcompat` rows.
3. Add the required Core identity, integrity, provenance, collision, and raw-loader bypass contract, then inventory and classify all production `config.Load*` call sites.
4. Redesign `gc doctor --fix` around preflight-first, scoped, failure-atomic, preserving edits, including air-gapped and concurrent-operation tests.
5. Pair public pin adoption with resolver/cache cutover tests that prove remote SHA materialization and retired synthetic-source diagnostics.
6. Make the bootstrap extraction slice self-contained: safe empty FS, synthetic fixture, `GC_BOOTSTRAP=skip` decision, `bootstrapManagedImportNames` update, old-path scan, and expanded gates.
7. Decide whether Go role de-roling is in scope; add the migration table and tests if yes, or narrow the design claim if no.
8. Convert the docs/DX plan into a registered canonical reference page, a wording matrix, an authoritative inventory/lint gate, and golden checks tied to the settled runtime contracts.
