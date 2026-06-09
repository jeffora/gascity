# Felix Moreau — DeepSeek V4 Flash (Docs & DX Consistency Review, Iteration 7, Independent)

**Verdict:** block

**Scope:** Cross-document consistency, operator terminology drift, tutorial integrity, maintenance-word disambiguation, runtime state continuity, and verification contracts. Reviewed against the current Iteration 7 design (`.gc/design-review-inputs/core-gastown-pack-migration/design.md`, L1045-1098), grounded in the live `docs/` tree, `docs/docs.json`, `cmd/gc` strings, `internal/events`, and `examples/`.

---

## Executive Summary

While the Iteration 7 design has matured significantly regarding rollout staging, behavior preservation manifests, and config-loading invariants (integrating Attempt 3/4 review resolution contracts), it remains critically under-scoped on the **documentation and DX migration tracks**.

The design is still trying to "update" files that do not exist, misses active references that will break post-migration, under-scopes rewrites of complex troubleshooting guidance, ignores downstream examples containing retired paths, and lacks a verification contract to bind the CLI, doctor, and docs to a single canonical verbatim vocabulary.

To prevent terminology drift and broken operator paths, **the design is blocked** until the following key gaps are resolved.

---

## Blocking Findings

### 1. [Blocker] System Packs Anchor Missing (Divergence on `system-packs.md`)
The design specifies "updating" `docs/reference/system-packs.md` (L1081-1083) to clarify that Core is the required system pack while Maintenance and Gastown are retired from the system set.

**The Reality:** 
- The file `docs/reference/system-packs.md` **does not exist** on disk under `docs/reference/`.
- The file is **not registered** in the navigation structure inside `docs/docs.json`.

Without an explicit creation step and registration in `docs.json`, the single canonical reference meant to anchor the new system pack hierarchy has no home and will remain invisible to operators.

**Required Action:** 
Add an explicit **create** instruction for `docs/reference/system-packs.md` in the update plan, outline its full canonical content structure, and include a step to register it under the Reference group in `docs/docs.json`.

---

### 2. [Blocker] Divergent Hand-Edit List on `shareable-packs.md`
The design states: "Update `docs/guides/shareable-packs.md`: remove 'core and maintenance stay implicit' guidance" (L1084-1085).

**The Reality:** 
- The actual file `docs/guides/shareable-packs.md` contains **no such text** regarding "implicit" guidance.
- What it **actually contains** (L103-105) is a live, copy-pasteable configuration example:
  ```toml
  [imports.maintenance]
  source = "../maintenance"
  export = true
  ```

If an operator copies this example post-migration, they will import a retired and deleted pack, breaking their configuration immediately. Furthermore, the `export = true` setting represents a transitive re-export contract hazard that the design completely ignores.

**Required Action:** 
Correct the instruction for `docs/guides/shareable-packs.md`. Explicitly direct the developer to replace the live `[imports.maintenance]` example with an active, non-retired pack example, and explain how the transition from a re-exported `maintenance` dependency to host-supplied Core configuration affects downstream consumers.

---

### 3. [Blocker] Under-Scoped `troubleshooting.md` Rewrite
The design limits its updates for `docs/getting-started/troubleshooting.md` to merely replacing `.gc/runtime/packs/maintenance` and `.gc/system/packs/maintenance` path strings (L1086-1088).

**The Reality:** 
The file contains several load-bearing prose references that personify the retired Maintenance pack or give obsolete guidance:
- **Line 146 ("Maintenance skips GitHub gate checks..."):** This treats Maintenance as an active actor. After retirement, this is highly confusing as the doctor will report that Maintenance is retired, yet troubleshooting claims it is actively skipping checks.
- **Line 241 ("The maintenance pack runs `jsonl-export` every 15 minutes..."):** This describes the `jsonl-export` cron behavior that is moving to the Core `dog` agent. It must be updated to attribute this behavior to Core.
- **Line 275 (`ARCHIVE=$(gc config get state_dir)/packs/maintenance/jsonl-archive`):** This is an operator-facing command that points to the retired path directory.
- **Line 351 ("...`GC_JSONL_MAX_PUSH_FAILURES=99` in the maintenance pack's environment"):** This instructs operators on how to configure environment variables. It must be updated to point to Core's environment.

**Required Action:** 
Expand the `troubleshooting.md` rewrite plan. Detail specific text replacements for the personified "Maintenance" actor, re-attribute the `jsonl-export` cron explanation to the Core `dog` maintenance agent, and update the `GC_JSONL_MAX_PUSH_FAILURES` environment guidance and command on line 275 to target Core.

---

### 4. [Blocker] Missing Downstream Updates (Stale Paths in Swarm and Dolt Examples)
The design's documentation and example update list is highly focused on `examples/gastown/` (L1048-1058) but completely neglects other active examples containing retired terminology and broken script paths.

**The Reality:** 
- **`examples/swarm/city.toml:4`** references "picks up dog-based infrastructure from the system maintenance pack".
- **`examples/dolt/pack.toml:6`** references "Dog-backed formulas and orders rely on the city's maintenance pack".
- **`examples/dolt/assets/scripts/port_resolve.sh:6`** and **`examples/dolt/port_resolve_test.go:148`** explicitly reference `.gc/system/packs/maintenance/assets/scripts/dolt-target.sh` or `filepath.Join(root, "gastown", "packs", "maintenance", "assets", "scripts", "dolt-target.sh")`. This path will fail to resolve as soon as the Maintenance pack is deleted from system materialization.

Leaving these stale references in non-Gastown example directories creates trapdoors for operators using them as blueprints.

**Required Action:** 
Add `examples/swarm/city.toml`, `examples/dolt/pack.toml`, `examples/dolt/assets/scripts/port_resolve.sh`, and `examples/dolt/port_resolve_test.go` to the update plan with explicit, verbatim edit instructions to purge retired Maintenance path references.

---

### 5. [Blocker] No Verification Contract (No Exact-Phrasing Doctor or CLI Lints)
While the design mentions adding "golden tests for doctor output and first-run/tutorial text" (L1096-1097), it lacks any formal verification contract to bind the doctor messages, generated CLI help, and documentation pages to a single canonical verbatim vocabulary.

**The Reality:** 
Without a strict docs-lint or verification test asserting identical phrasing across all boundaries, the CLI help (`docs/reference/cli.md`), doctor messages (`cmd/gc/import_state_doctor_check.go`), manual pages, and tutorials will drift.

For example, the canonical block at L1068 says *"Core is the required host system pack"*, while L1081 says *"Core is the required system pack"* (dropping the word "host"). This terminology drift exists inside the design itself, proving that human vigilance alone is insufficient.

**Required Action:** 
Specify a test (e.g., in `test/packlint` or a dedicated docs-lint test) that asserts that the exact canonical terminology block is used verbatim across:
- The import-state doctor check message and its `FixHint`
- The system-packs anchor page
- CLI help snapshots (`docs/reference/cli.md`)
- `docs/reference/config.md`
- Tutorial 01's maintenance-agent explanation

---

## Major Findings

### [Major] Missing Updates in Key Guides (`coming-from-gastown.md` and `migrating-to-pack-vnext.md`)
Several other important documentation guides are completely omitted from the update list:
- **`docs/getting-started/coming-from-gastown.md:543-545`** contains literal permalinks to `examples/gastown/packs/gastown/pack.toml` in the main repo. This directory will be deleted by the migration, making these links 404.
- **`docs/guides/migrating-to-pack-vnext.md:134`** uses `includes = ["packs/gastown"]` and L534 uses `source = "./assets/imports/maintenance"`. Both represent obsolete, non-resolving paths post-migration.

**Required Action:** 
Add these files to the update list. Update the links in `coming-from-gastown.md` to point to the public pack repo, and replace the retired includes/imports in `migrating-to-pack-vnext.md` with active public pack import syntax.

### [Major] Walkthrough Error Paths (`gc-start-walkthrough.mdx`)
`docs/troubleshooting/gc-start-walkthrough.mdx:134-135` contains:
```
(from packs/gastown/pack.toml [agent #2]
 and .gc/system/packs/gastown/agents/mayor/agent.toml)
```
This references `.gc/system/packs/gastown` as a system pack. Since Gastown will be an external imported pack post-migration, this path will change, making the troubleshooting walkthrough inaccurate and confusing for new operators.

Additionally, `docs/troubleshooting/gc-start-walkthrough.mdx:262` refers to `includes = ["packs/gastown"]`, which represents the legacy PackV1 include layout.

**Required Action:** 
Add `gc-start-walkthrough.mdx` to the update list, update the duplicate-agent-name error output to show the new external import paths, and replace legacy includes with active PackV2 imports.

### [Major] `gc.store.maintenance.*` Event Disambiguation
The design does not clarify whether the event definitions referencing `gc.store.maintenance.done` and `gc.store.maintenance.failed` inside `docs/schema/openapi.json` and `internal/events/events.go` are in scope for modification. These references are to **Dolt store maintenance**, not the retired Maintenance pack. Renaming them would be a wire-format breaking change for the SSE API.

**Required Action:** 
Explicitly declare that `gc.store.maintenance.*` events and their OpenAPI schemas are out-of-scope for the migration and add them to the docs-lint allowlist.

### [Major] Scope Rule for `docs/plans/` and `docs/runbooks/`
Several historical runbooks (like `docs/runbooks/managed-city-endpoints.md`) and design plans contain active references to the retired Maintenance pack and its directories. The design does not provide a classification or scope rule for these directories, leaving them in a gray area.

**Required Action:** 
Define a clear classification rule: either mark `docs/plans/` and `docs/runbooks/` as frozen/historical (and add them to a lint allowlist) or place them in-scope for active text replacement.

### [Major] `dog` Worker Pool Terminology Ambiguity
`docs/tutorials/01-cities-and-rigs.md:189` says:
> The `dog` pool is a background utility agent from the built-in maintenance pack.

The design proposes changing "built-in maintenance pack" to "Core's configurable maintenance agent" (L1089-1090). However, `dog` is a *pool* (a configurable scaling group), not an *agent*. Tutorial output on L175-177 shows `dog` as a scaled pool with `min=0, max=3`, not a single named session.

**Required Action:** 
Update the description to: "The `dog` pool is a background utility worker pool provided by the required Core system pack. It handles internal housekeeping like shutdown coordination and JSONL archival."

---

## Summary of Required Changes

1. **Resolve the system-packs anchor:** Add an explicit *create* step for `docs/reference/system-packs.md` and register it in `docs/docs.json` navigation.
2. **Correct `shareable-packs.md`:** Replace the live `[imports.maintenance]` example block with an active non-retired pack example, and explain transition impact.
3. **Expand `troubleshooting.md` rewrite:** Detail specific text replacements for the personified "Maintenance" actor, the `jsonl-export` re-attribution to Core's `dog` pool, the `GC_JSONL_MAX_PUSH_FAILURES` env-var guidance, and the `ARCHIVE` command path on line 275.
4. **Include downstream examples:** Add `examples/swarm/city.toml`, `examples/dolt/pack.toml`, `examples/dolt/assets/scripts/port_resolve.sh`, and `examples/dolt/port_resolve_test.go` to the update plan.
5. **Establish a verification contract:** Ensure a docs-lint/golden test asserts that the canonical vocabulary is used verbatim across CLI help, doctor messages, and docs.
6. **Include guide files:** Add `coming-from-gastown.md`, `migrating-to-pack-vnext.md`, and `gc-start-walkthrough.mdx` to the update plan with explicit edit instructions.
7. **Explicitly preserve Dolt store maintenance events:** State that `gc.store.maintenance.*` is out of scope.
8. **Define scope for plans and runbooks:** Declare whether historical markdown folders are frozen or actively updated.
9. **Fix `dog` pool terminology:** Define `dog` as a worker pool, not a single agent.

---

## Questions

- Should `jsonl-export.sh` and `spawn-storm-detect.sh` use `$GC_PACK_STATE_DIR` exclusively (removing the `packs/maintenance` hardcoded fallback), or should they retain a legacy fallback for existing upgraded cities?
- For order qualified names, is the intent that deployed cities silently accept both `maintenance.gate-sweep` and `core.gate-sweep`, or that `gc doctor --fix` rewrites old qualifiers?
- Are formula filenames like `mol-polecat-base.toml` in scope for the role-name guard, or are they considered configuration artifacts (not Go code)?
- Should `examples/swarm/` and `examples/dolt/` comments be updated in the same slice that removes `examples/gastown/packs/maintenance/`, or are they on a separate timeline?
