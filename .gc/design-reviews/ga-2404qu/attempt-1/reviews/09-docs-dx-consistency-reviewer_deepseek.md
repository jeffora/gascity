# Felix Moreau — DeepSeek V4 Flash (Cross-Document Consistency & Edge-Case Review)

**Verdict:** block

**Scope:** Cross-document consistency, operator terminology, tutorial integrity, maintenance-word disambiguation, runtime state continuity, cross-pack contract correctness, and assumptions other reviewers may accept too quickly.

---

## Top Strengths

- The design correctly identifies the `dog` pool wording problem in `docs/tutorials/01-cities-and-rigs.md:189` and calls for reframing it as Core's configurable maintenance agent.
- The import-state doctor messaging is properly scoped — the current "supplied implicitly" wording at `cmd/gc/import_state_doctor_check.go:194` is identified as needing replacement.
- The rollback section correctly flags order-skip-list name preservation as a concern for renamed orders.
- Protecting `[maintenance.dolt]` store-maintenance terminology from careless global replacement is the right call.
- The stale-directory preservation policy (`gc doctor --fix` must not delete `.gc/system/packs/maintenance` or `.gc/system/packs/gastown`) is operator-safe.

---

## Blocking Findings

### [Blocker] Runtime state paths under `.gc/runtime/packs/maintenance/` are unaddressed, creating silent data loss risk

The design discusses `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` as stale directories to be ignored, but scripts that will move into Core currently write persistent state to `.gc/runtime/packs/maintenance/`:

- `jsonl-export.sh:22` hardcodes `PACK_STATE_DIR="${GC_PACK_STATE_DIR:-${GC_CITY_RUNTIME_DIR:-$CITY/.gc/runtime}/packs/maintenance}"`
- `spawn-storm-detect.sh:20` hardcodes the same path
- `cmd/gc/jsonl_archive_doctor_check.go:71,97` hardcodes `filepath.Join(runtime, "packs", "maintenance")` in both `resolveStateFile()` and `resolveArchiveRepo()`

After migration, Core-based scripts receive `GC_PACK_STATE_DIR` pointing to `.gc/runtime/packs/core/` (set by the controller via `citylayout.PackRuntimeEnv(cityPath, "core")`). But existing cities have JSONL archive repos, state files, and storm-count ledgers under `.gc/runtime/packs/maintenance/`.

The design's runtime-state table (design-after.md L467-472) lists three options (Core destination, ignored-legacy diagnostic, manual migration note) without choosing one. This is not a docs issue — it is a data-continuity issue. If scripts start writing to `.gc/runtime/packs/core/` while the doctor only reads `.gc/runtime/packs/maintenance/`, operators lose both push-history continuity and doctor visibility on the new path.

Further, `jsonl-export.sh` already has `LEGACY_ARCHIVE_REPO` and `LEGACY_STATE_FILE` fallback patterns that search `$CITY/.gc/jsonl-archive` and `$CITY/.gc/jsonl-export-state.json`. The design does not specify whether these legacy patterns are preserved, replaced with a `packs/core` primary path, or extended with a `packs/maintenance` fallback. Each choice produces different troubleshooting commands, doctor output, and rollback behavior.

**Required change:** Add a runtime-state migration table with one row per state class (JSONL archive repository, `jsonl-export-state.json` and push-failure counters, spawn-storm state, order-tracking state). For each, specify: owner after migration, current path, legacy read path, write path, doctor behavior, `gc doctor --fix` behavior, operator-facing command examples, and rollback behavior. Wire the docs and golden-test requirements to this table.

### [Blocker] The system-pack reference anchor `docs/reference/system-packs.md` does not exist and is not in docs navigation

The design's canonical-wording deliverable points to `docs/reference/system-packs.md` (design-after.md L537-539), but this file does not exist in the tree. The current `docs/reference/` directory contains `index.md`, `cli.md`, `config.md`, `formula.md`, `trust-boundaries.md`, `api.md`, `events.md`, `exec-session-provider.md`, and `exec-beads-provider.md` — no `system-packs.md`. The `docs/docs.json` navigation has no entry for it.

The design says "update" rather than "create," and never adds a `docs/docs.json` nav entry. The single page meant to anchor consistent Core/Gastown/retired-Maintenance language has no home.

**Required change:** Either (a) add an explicit *create* step for `docs/reference/system-packs.md` with canonical wording, and register it in `docs/docs.json` navigation, or (b) designate an existing registered page (such as `docs/reference/config.md` or `docs/guides/shareable-packs.md`) as the canonical terminology anchor and update that page instead.

### [Blocker] The `[[patches.agent]]` cross-pack resolution is undefined for required-but-not-imported system packs

`examples/gastown/packs/gastown/pack.toml:25-28` currently patches `dog` by bare name:

```toml
[[patches.agent]]
name = "dog"
wake_mode = "fresh"
work_dir = ".gc/agents/dogs/{{.AgentBase}}"
```

After migration, `dog` comes from Core. Core is a required system pack injected via `builtinPackIncludes`, not an explicit `[imports.core]` declaration. The current patch-resolution logic in `internal/config/compose.go` matches by bare agent name against agents in the resolved config. The question is whether `[[patches.agent]] name = "dog"` in the public Gastown pack will match Core's `dog` when Core is not an explicit import.

Two options exist, both with docs/DX consequences:

1. **Gastown patches `core.dog`** — This requires packs to patch agents from required system packs they don't explicitly import. This is new semantics for `[[patches.agent]]` that the current config loader does not support. The design must specify this new behavior and test it.
2. **Gastown defines its own `dog`** — This duplicates the fallback dog from Core and removes the ability to patch a single shared definition. The comment in `examples/gastown/city.toml:15` ("Maintenance still supplies the fallback dog shape and shared dog formulas/prompts that gastown reuses") becomes inaccurate.

Without a resolution, the public `gascity-packs/gastown` pack will fail to apply its dog-patch during config resolution, breaking the primary Gastown use case. This is a semantic gap in the config system, not just a docs issue.

**Required change:** Specify how `[[patches.agent]]` resolves against required-but-not-imported system packs. Document the chosen resolution in the design and add config-resolution tests proving it works.

---

## Major Findings

### [Major] The `shareable-packs.md` import example is a copy-paste trap, not just stale wording

`docs/guides/shareable-packs.md:103-104` shows:

```toml
[imports.maintenance]
source = "../maintenance"
export = true
```

After the migration, this example instructs operators to import a deleted pack and re-export it. The design's docs section says "remove 'core and maintenance stay implicit' guidance" (L540-541), but the file contains no "implicit" text. The actual hazard is the live import example, which an operator could copy verbatim into their pack.toml, producing a broken config.

**Required change:** Replace the `[imports.maintenance]` example with one that uses a non-retired pack (e.g., `[imports.gastown]` with the public source), and add a note that Core and provider packs are implicitly included and should not be imported.

### [Major] The `pack-schema.json` includes description uses `"../maintenance"` as its canonical example

`docs/schema/pack-schema.json:872` contains:

```json
"description": "Includes lists other packs to compose into this one (V1 mechanism).\nEach entry is a local relative path (e.g. \"../maintenance\") or a\nremote git URL (SSH or HTTPS) with optional //subpath and #ref."
```

This is a generated schema file that serves as the API contract for pack configuration. Having `"../maintenance"` as the canonical example for the `includes` field is not just a stale doc reference — it's a stale contract definition. The design's inventory command (L515-520) scans only `*.md`, `*.toml`, `*.go`, and `*.sh` files, so generated JSON/TXT schema artifacts are not covered.

**Required change:** Expand the inventory/lint gate to include `docs/reference/cli.md`, `docs/reference/config.md`, `docs/schema/pack-schema.{json,txt}`, `docs/schema/city-schema.{json,txt}`, and `docs/docs.json` navigation entries. Specify whether each generated artifact is regenerated from source or allowlisted as historical.

### [Major] Tutorial 07 lists orders that will change ownership without specifying which city template

`docs/tutorials/07-orders.md:87-93` lists "built-in `mol-*` orders that ship with the tutorial template" including `beads-health`, `gate-sweep`, `mol-dog-jsonl`, `mol-dog-reaper`, `orphan-sweep`, `prune-branches`, `spawn-storm-detect`, and `wisp-compact`. After migration:

- `beads-health` stays in Core
- `gate-sweep`, `mol-dog-jsonl`, `mol-dog-reaper`, `orphan-sweep`, `prune-branches`, `spawn-storm-detect`, `wisp-compact` move from Maintenance to Core
- `mol-shutdown-dance` (a formula, not an order) is not listed but is Maintenance-owned

The tutorial doesn't specify whether this is a minimal or Gastown city. For a minimal city, these Core orders would all appear. For a Gastown city, the list would also include Gastown-specific orders. The design's explicit docs list (L537-546) does not name `07-orders.md`.

**Required change:** Add `07-orders.md` to the docs update list. Specify the order enumeration for each template type (minimal vs Gastown) and label Gastown-specific orders accordingly.

### [Major] The `dog` pool description has a terminology error the design doesn't fix

`docs/tutorials/01-cities-and-rigs.md:189` says:

> The `dog` pool is a background utility agent from the built-in maintenance pack.

The design proposes changing "built-in maintenance pack" to "Core's configurable maintenance agent" (L545). But `dog` is a *pool* (a configurable scaling group), not an *agent*. The tutorial output on L175-177 shows `dog` as a scaled pool with `min=0, max=3`, not a named session. Calling it a "maintenance agent" swaps one inaccuracy for another.

**Required change:** Change the description to something like: "The `dog` pool is a background utility worker pool provided by the required Core system pack. It handles internal housekeeping like shutdown coordination and JSONL archival."

### [Major] Example comments in `examples/swarm/city.toml` and `examples/dolt/pack.toml` reference the retired Maintenance pack by name

- `examples/swarm/city.toml:4`: "picks up dog-based infrastructure from the system maintenance pack"
- `examples/dolt/pack.toml:6`: "Dog-backed formulas and orders rely on the city's maintenance pack"
- `examples/dolt/assets/scripts/port_resolve.sh:6`: "`.gc/system/packs/maintenance/assets/scripts/dolt-target.sh`"
- `examples/gastown/packs/gastown/assets/scripts/status-line.sh:15-16`: searches for `_bd_trace.sh` under `.gc/system/packs/maintenance`

These are operator-facing comments and script paths in functional examples. The design's inventory command (L515-520) should catch `.toml` and `.sh` files, but the design's explicit docs list (L537-546) does not name `examples/swarm/`, `examples/dolt/`, or the Gastown scripts.

**Required change:** Add `examples/swarm/city.toml`, `examples/dolt/pack.toml`, `examples/dolt/assets/scripts/port_resolve.sh`, and `examples/gastown/packs/gastown/assets/scripts/status-line.sh` to the inventory completion gate. Specify whether each is updated in place or added to a named allowlist with a classification (historical fixture, legacy diagnostic test, still-valid reference).

### [Major] The `coming-from-gastown.md` guide links to a file the migration deletes

`docs/getting-started/coming-from-gastown.md:545` says:

> then [examples/gastown/packs/gastown/pack.toml](https://github.com/gastownhall/gascity/blob/main/examples/gastown/packs/gastown/pack.toml)

The migration removes `examples/gastown/packs/gastown/` from the tree (design-after.md slice 7). This GitHub permalink will 404 after the source deletion slice lands.

**Required change:** Update the link to point to the public `gascity-packs/gastown` pack source, or restructure the onboarding guide to reference the public pack documentation.

### [Major] `migrating-to-pack-vnext.md` shows the retired include syntax alongside the retired import syntax

`docs/guides/migrating-to-pack-vnext.md:134` shows:

```toml
includes = ["packs/gastown"]
```

and L169 shows:

```toml
[imports.gastown]
```

The include syntax at L134 uses the old local path that won't exist after the in-tree Gastown sources are removed. The import syntax at L144-148 is correct for the public pack model. The design does not name this file in its update list.

**Required change:** Update the include example to use the public pack source, or mark it clearly as a legacy syntax that only works with local packs. Add this file to the docs inventory completion gate.

### [Major] The `gc-start-walkthrough.mdx` error message examples reference `.gc/system/packs/gastown`

`docs/troubleshooting/gc-start-walkthrough.mdx:134-135` shows:

```
(from packs/gastown/pack.toml [agent #2]
 and .gc/system/packs/gastown/agents/mayor/agent.toml)
```

After migration, Gastown is an external import, not a system pack. The error message format and paths will change. The walkthrough does not explain what these paths mean in the context of an external import, and the design does not name this file in its update list.

**Required change:** Add `gc-start-walkthrough.mdx` to the docs inventory. Update the duplicate-agent-name error example to show the new paths for externally-imported packs.

### [Major] The `troubleshooting.md` JSONL guidance references retired Maintenance in three places, not just path strings

The design says to replace `.gc/runtime/packs/maintenance` and `.gc/system/packs/maintenance` references in `troubleshooting.md` (L542-544). But the file has three distinct references that require different rewrites:

1. `:241` — "The maintenance pack runs `jsonl-export` every 15 minutes" — This personifies the retired pack. After migration, the Core dog agent runs it.
2. `:275` — `ARCHIVE=$(gc config get state_dir)/packs/maintenance/jsonl-archive` — This is an operator-facing command. The path cannot be rewritten until the runtime-state decision is made (see Blocker 1).
3. `:351` — "set `GC_JSONL_MAX_PUSH_FAILURES=99` in the maintenance pack's environment" — After migration, this env var is set in the Core dog agent's environment, not the maintenance pack's.

**Required change:** Expand the `troubleshooting.md` rewrite to cover all three references. Sequence the `:275` command rewrite after the runtime-state JSONL destination decision. Update `:241` and `:351` to reference Core's dog agent regardless.

### [Major] Order qualified-name migration affects pool resolution, not just skip lists

The design mentions that "existing order skip lists containing moved Core order names should continue to work when names are preserved" (design-after.md L531). But the concern is broader than skip lists. Orders from Maintenance are currently qualified as `maintenance.gate-sweep`, `maintenance.mol-dog-jsonl`, etc. The `orderroute.QualifyPool` function in `internal/orderroute/orderroute.go` resolves pool targets by matching the order's formula layer directory against agent source directories.

When `gate-sweep.toml` moves from `examples/gastown/packs/maintenance/orders/` to `internal/packs/core/orders/`, its `FormulaLayer` source directory changes from `packs/maintenance` to `packs/core`. This changes the `sourceDirHint` in `QualifyPool`, which changes how the `dog` pool is resolved for these orders. The design does not address this runtime resolution change.

**Required change:** Specify that moved orders retain their pool resolution behavior after migration. Add a test proving that `QualifyPool` resolves `dog` to the correct qualified name whether the order comes from `maintenance/orders/` or `core/orders/`.

---

## Minor Findings

### [Minor] "Maintenance skips GitHub gate checks" personifies the retired pack

`docs/getting-started/troubleshooting.md:146` says "Maintenance skips GitHub gate checks when the GitHub CLI is not installed." This is the Maintenance pack's gate-sweep order behavior, which moves to Core. After retirement, `gc doctor` says "maintenance is retired" while this line still treats Maintenance as an active agent. The design's disambiguation list does not cover this.

### [Minor] The `gc rig add --include packs/gastown` CLI example references a local path that won't exist after migration

`docs/reference/cli.md:2450-2452` shows:

```
gc rig add ./my-project --include packs/gastown
gc rig add ./my-project --include packs/gastown --start-suspended
```

After the migration removes in-tree `examples/gastown/packs/gastown`, this local path won't resolve. The example should show the public import syntax.

### [Minor] The inventory grep does not cover `.mdx` files

The design's inventory command (L515-520) uses `-g '*.md'` but the tree contains `.mdx` files in `docs/troubleshooting/gc-start-walkthrough.mdx`. The glob should include `*.mdx` or use a more inclusive pattern.

---

## Missing Evidence

- A runtime-state migration table for `.gc/runtime/packs/maintenance/` (JSONL archive, export state, storm ledgers, order tracking) specifying current path, legacy read path, write path, doctor behavior, and rollback.
- Whether `[[patches.agent]] name = "dog"` resolves against required-but-not-imported system packs, and what config loader change supports this.
- Whether `jsonl-export.sh` and `spawn-storm-detect.sh` should use `$GC_PACK_STATE_DIR` exclusively or retain a `packs/maintenance` legacy fallback, and what happens to the existing `LEGACY_ARCHIVE_REPO` and `LEGACY_STATE_FILE` patterns.
- A generated inventory of all docs, examples, command help, schema docs, scripts, pack comments, doctor strings, and generated references containing retired pack terminology.
- A real canonical system-pack reference destination: either a new `docs/reference/system-packs.md` wired into `docs/docs.json` or an existing doc path designated as the owner.
- A source-of-truth terminology matrix for required Core, explicit public Gastown, retired Maintenance, Core maintenance-agent behavior, store maintenance, and Dolt maintenance.
- The target first-run/tutorial story for minimal `gc init`, `gc status`, Core's `dog` pool, and `--template gastown`.
- A decision on whether moved order names keep `maintenance.*`, gain `core.*`, support aliases, or require `gc doctor --fix`/manual migration for `skip_orders` and `order_config`.
- The scope of the Core role-name guard and its allowlist for pack asset filenames, prompt content, formula filenames, skill docs, and test fixtures.
- A transition fixture plan for `examples/gastown/` so tests stay green between the public Gastown availability slice and the in-tree source deletion slice.
- Whether `examples/swarm/city.toml`, `examples/dolt/pack.toml`, and `examples/dolt/assets/scripts/port_resolve.sh` are in scope for terminology updates or on a named allowlist.

---

## Required Changes

1. **Resolve the runtime-state migration.** Add a table specifying disposition for each `.gc/runtime/packs/maintenance/` state class: owner, current path, legacy read path, write path, doctor behavior, `gc doctor --fix` behavior, operator command examples, and rollback. Wire docs and golden-test requirements to this table.
2. **Make the system-pack reference anchor concrete.** Create `docs/reference/system-packs.md` and register it in `docs/docs.json`, or designate an existing registered page.
3. **Specify `[[patches.agent]]` resolution against required system packs.** Define whether Gastown's `name = "dog"` patch matches Core's `dog` without an explicit `[imports.core]`, and add config-resolution tests.
4. **Fix the `shareable-packs.md` import example.** Replace `[imports.maintenance]` with a non-retired pack example. Remove the `export = true` guidance for retired packs.
5. **Expand the inventory/lint gate** to include `docs/reference/cli.md`, `docs/reference/config.md`, `docs/schema/pack-schema.{json,txt}`, `docs/schema/city-schema.{json,txt}`, `docs/docs.json`, `docs/troubleshooting/gc-start-walkthrough.mdx`, `docs/getting-started/coming-from-gastown.md`, `docs/guides/migrating-to-pack-vnext.md`, `docs/tutorials/07-orders.md`, `examples/swarm/city.toml`, `examples/dolt/pack.toml`, `examples/dolt/assets/scripts/port_resolve.sh`, `cmd/gc/jsonl_archive_doctor_check.go`, and `.mdx` files.
6. **Update all three `troubleshooting.md` Maintenance references** (personification at `:241`, path command at `:275`, env var at `:351`). Sequence the `:275` rewrite after the runtime-state decision.
7. **Fix the `dog` pool terminology** — call it a "worker pool," not an "agent."
8. **Add `07-orders.md` to the docs update list** with order enumeration by city template type.
9. **Update `coming-from-gastown.md:545`** to point to the public pack source instead of the deleted in-tree path.
10. **Update `migrating-to-pack-vnext.md:134`** include example to use public pack syntax.
11. **Update `gc-start-walkthrough.mdx:134-135`** error message paths for external import model.
12. **Define the `troubleshooting.md:146` rewrite** — "Core's gate-sweep order" instead of "Maintenance skips GitHub gate checks."
13. **Update `cli.md:2450-2452`** `--include packs/gastown` example to use public import syntax.
14. **Add order-qualified-name pool resolution test** proving `QualifyPool` resolves correctly after Maintenance-to-Core migration.
15. **Add `.mdx` to the inventory grep patterns** so `gc-start-walkthrough.mdx` is not missed.

---

## Questions

- Should `jsonl-export.sh` and `spawn-storm-detect.sh` use `$GC_PACK_STATE_DIR` exclusively (removing the `packs/maintenance` hardcoded fallback), or should they retain a legacy fallback for existing cities?
- Is the `[[patches.agent]] name = "dog"` mechanism expected to work across required-but-not-imported system packs, or does Gastown need to define its own `dog` agent after migration?
- For order qualified names, is the intent that deployed cities silently accept both `maintenance.gate-sweep` and `core.gate-sweep`, or that `gc doctor --fix` rewrites old qualifiers?
- Are formula filenames like `mol-polecat-base.toml` in scope for the role-name guard, or are they considered configuration artifacts (not Go code)?
- Should `jsonl_archive_doctor_check.go` gain a dual-path lookup during the transition, or should existing `packs/maintenance/jsonl-archive` state be migrated by `gc doctor --fix`?
- Should `examples/swarm/` and `examples/dolt/` comments be updated in the same slice that removes `examples/gastown/packs/maintenance/`, or are they on a separate timeline?
