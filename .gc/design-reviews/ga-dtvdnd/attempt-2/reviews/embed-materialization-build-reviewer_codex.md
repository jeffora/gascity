# Petra Novak - Codex

**Verdict:** block

**Top strengths:**
- The requirements state the right target outcome for this lane: Core remains required, Maintenance is retired, and Gastown is loaded explicitly from the public pack.
- AC6 names the correct shape for an asset migration ledger, including current path, target owner, target output path or retirement action, rationale, and proof command.
- AC14 correctly guards against the highest-risk false pass: a local in-tree Gastown copy must not mask a broken public Gastown pack.

**Critical risks:**
- [Blocker] The requirements do not yet define Core's canonical embedded source identity or old-source compatibility behavior. Today `internal/builtinpacks/registry.go` embeds Core from `internal/bootstrap/packs/core`, and `internal/bootstrap/packs/core/embed.go` exposes the `PackFS` consumed by `cmd/gc/embed_builtin_packs.go`. The document says Core is required, but it does not say whether `internal/bootstrap/packs/core` remains canonical, moves to a new path, or becomes a diagnosed legacy alias. Without that decision, implementation can update registry entries, embed directives, generated system-pack paths, import-state diagnostics, and tests inconsistently.
- [Blocker] Maintenance retirement is a stated product outcome, but the required proof artifacts are still absent and open. The current code still embeds and materializes Maintenance as required through `builtinpacks.All()` and `requiredBuiltinPackNames()`, while `examples/gastown/city.toml`, `examples/dolt/pack.toml`, `examples/gastown/gastown_test.go`, `cmd/gc/order_dispatch_test.go`, and runtime-state checks still refer to `packs/maintenance`, `maintenance.dog`, or `.gc/runtime/packs/maintenance`. AC6 and AC7 require a ledger and behavior-preservation manifest, but the requirements remain in `questions` status and only ask who/what will produce them. That is not enough evidence to approve implementation planning for moving or deleting embedded assets.
- [Major] Public and bundled alias handling for retired packs is underspecified. Current `internal/builtinpacks/registry.go` maps both `gastown` and `maintenance` through the public `https://github.com/gastownhall/gascity-packs.git` synthetic-cache path, and `MaterializeSyntheticRepo` writes every synthetic layout. AC5 says Maintenance is not bundled or public-source recognized, and AC14 says public Gastown must not be masked by an in-tree copy, but the example mapping does not explicitly cover `gascity-packs.git//maintenance`, `gascity.git//examples/gastown/packs/maintenance`, or a synthetic-cache hit for `gascity-packs.git//gastown`. That leaves room for a migration that passes local import tests while retaining hidden fallback behavior.
- [Major] The ledger contract covers source files, but the requirements do not explicitly require downstream reference closure for generated commands, script paths, runtime state paths, docs, and tests that consume those assets. Examples include `cmd/gc/jsonl_archive_doctor_check.go` looking under `.gc/runtime/packs/maintenance`, `cmd/gc/embed_builtin_packs_test.go` asserting Maintenance materialization, and Gastown tests expecting `packs/maintenance/formulas`. The ledger must prove these consumers are repointed, retired, or intentionally preserved as legacy diagnostics, not only that source files have target owners.

**Missing evidence:**
- No validated asset migration ledger path or validation command exists yet.
- No behavior-preservation manifest or harness exists yet for supported Gastown formulas, orders, scripts, prompts, template variables, notification targets, and recovery flows.
- No explicit compatibility matrix states whether old Core, Maintenance, and Gastown sources are accepted, rejected, repaired, ignored, or diagnosed across local paths, system-pack paths, public-pack URLs, lockfiles, and synthetic caches.
- No proof command demonstrates that the pinned public Gastown pack can run with host Core and without an active Maintenance pack.

**Required changes:**
- Keep the requirements unapproved until the migration ledger and behavior-preservation proof exist, or add a blocking prerequisite section that names their concrete paths, owners, validation commands, and approval gate.
- Add an acceptance criterion or example that defines Core's canonical embedded source identity and the diagnostic treatment for the old Core source path.
- Extend AC5/AC10/AC14 with explicit cases for retired Maintenance public aliases, old `gascity.git//examples/gastown/packs/maintenance` aliases, stale synthetic caches, and public Gastown cache hits that would otherwise be satisfied by embedded in-tree content.
- Strengthen AC6 so the ledger must include downstream source references and generated/runtime paths, not only the pack files being moved or retired.
- Require a no-Maintenance materialization check proving `MaterializeBuiltinPacks`, `builtinPackIncludes`, import resolution, doctor/import-state, and fresh Gastown init cannot silently activate Maintenance after the migration.

**Questions:**
- What is the canonical Core source path after this migration, and should existing `internal/bootstrap/packs/core` references remain valid, warn, or fail?
- Can the `gc` binary continue to synthesize a cache for `gascity-packs.git//gastown`, or must fresh Gastown init prove a real public-pack checkout/cache instead?
- Which artifact owns cross-repo behavior preservation: the Gas City plan, the public Gastown pack release, or a shared release gate consumed by both?
- What exact repair action removes or ignores stale `.gc/system/packs/maintenance` and `.gc/runtime/packs/maintenance` without deleting operator-owned data accidentally?
