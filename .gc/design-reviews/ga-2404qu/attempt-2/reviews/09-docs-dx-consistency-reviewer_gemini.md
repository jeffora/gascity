# Felix Moreau — DeepSeek (Docs & DX Consistency Review)

**Verdict:** block

**Scope:** Documentation consistency, operator terminology, tutorial integrity, maintenance-word disambiguation, runtime state path continuity, pack-name migration impact on qualified references, and cross-pack contract correctness.

---

## Top strengths

- The design correctly identifies the operator-facing terminology problem: `docs/tutorials/01-cities-and-rigs.md:189` calls `dog` "a background utility agent from the built-in maintenance pack" and proposes reframing it as Core's configurable maintenance agent.
- The import-state doctor messaging is in scope — `cmd/gc/import_state_doctor_check.go:194` currently says "should be removed; maintenance/core is supplied implicitly," which contradicts the target state where Core is a *required* system pack, not an implicit convenience.
- Protecting `[maintenance.dolt]` store-maintenance terminology from careless global replacement is the right call.
- The rollback section correctly identifies order-skip-list name preservation as a concern.
- The design's asset inventory table is thorough and covers the right files.

---

## Critical risks

### [Blocker] Runtime state paths under `.gc/runtime/packs/maintenance/` are unaddressed — operator data loss risk

The design mentions `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` as stale directories to ignore, but never addresses `.gc/runtime/packs/maintenance/`, where live state lives.

Current code hardcodes `packs/maintenance` in runtime paths:

- `examples/gastown/packs/maintenance/assets/scripts/jsonl-export.sh:22` — `PACK_STATE_DIR="${GC_PACK_STATE_DIR:-${GC_CITY_RUNTIME_DIR:-$CITY/.gc/runtime}/packs/maintenance}"`
- `examples/gastown/packs/maintenance/assets/scripts/spawn-storm-detect.sh:20` — same `PACK_STATE_DIR` pattern
- `cmd/gc/jsonl_archive_doctor_check.go:60,71,97` — `filepath.Join(runtime, "packs", "maintenance")` in both `resolveStateFile()` and `resolveArchiveRepo()`

After migration, the scripts will use `GC_PACK_STATE_DIR` (set by the controller via `citylayout.PackRuntimeEnv(cityPath, "core")`), writing new state to `.gc/runtime/packs/core/`. But existing cities have JSONL archive repos, state files, and storm-count ledgers under `.gc/runtime/packs/maintenance/`. The design never specifies:

1. Whether existing state is migrated in-place, read through a compatibility path, or reported with manual instructions.
2. Whether `jsonl_archive_doctor_check.go` gains a legacy fallback path. The script already has `LEGACY_ARCHIVE_REPO` and `LEGACY_STATE_FILE` patterns, but the Go doctor check does not.
3. What happens to operator documentation at `docs/getting-started/troubleshooting.md:275` that references `$(gc config get state_dir)/packs/maintenance/jsonl-archive`.

The `citylayout.PackStateDir()` infrastructure already provides pack-agnostic path resolution — the design should specify that migrated scripts and doctor checks use `GC_PACK_STATE_DIR` exclusively and that `jsonl_archive_doctor_check.go` gains a legacy-path fallback during the transition, rather than hardcoding `packs/maintenance`.

This is a **data-continuity** issue. Existing archive repos contain operator git history. If the script silently starts a fresh archive under `packs/core/` while the doctor check only looks at `packs/maintenance/`, operators lose both push-history continuity and doctor visibility.

### [Blocker] The `[[patches.agent]] name = "dog"` cross-pack resolution is unspecified

The Gastown pack currently patches `maintenance.dog` (`examples/gastown/packs/gastown/pack.toml:29`):

```toml
[[patches.agent]]
name = "dog"
wake_mode = "fresh"
work_dir = ".gc/agents/dogs/{{.AgentBase}}"
```

After migration, `dog` comes from Core. But Core is a required system pack injected by `builtinPackIncludes` — not an explicit import. The Gastown pack does not and cannot declare `[imports.core]` because Core is implicit.

The design does not address how patch-qualified names resolve when the target agent comes from a required system pack that the patching pack does not explicitly import. Two options:

1. **Gastown patches `core.dog`** — Requires packs to patch agents from required system packs they don't import. This is a new semantics for `[patches.agent]` that the current config loader may not support.
2. **Gastown defines its own `dog`** — Duplicates the fallback dog definition from Core and removes the ability to patch a single shared definition. It also contradicts `examples/gastown/city.toml:15` which says "Maintenance still supplies the fallback dog shape."

Without a resolution, the public `gascity-packs/gastown` pack will fail to apply its dog patch during config resolution, breaking the primary Gastown use case.

### [Major] Order qualified names change from `maintenance.*` to `core.*` with no migration strategy

Maintenance currently provides 9 orders under `maintenance.*` qualified names (`maintenance.gate-sweep`, `maintenance.mol-dog-jsonl`, `maintenance.mol-dog-reaper`, `maintenance.mol-shutdown-dance`, `maintenance.orphan-sweep`, `maintenance.order-tracking-sweep`, `maintenance.prune-branches`, `maintenance.spawn-storm-detect`, `maintenance.wisp-compact`). After migration, these move to Core and become `core.gate-sweep`, etc.

The design says "existing order skip lists containing moved Core order names should continue to work when names are preserved" — but the pack-name qualifier *changes from `maintenance` to `core`*, so names are **not** preserved. Every `skip_orders` entry, `order_config` override, and `orders` reference in deployed cities that uses `maintenance.*` qualified names will break.

The design needs to specify one of:
- Alias support so `maintenance.gate-sweep` resolves to `core.gate-sweep`
- A `pack-name` override in Core's embedded `pack.toml` that preserves backward-compatible qualified names
- A doctor migration step that rewrites `maintenance.*` qualifiers to `core.*` in deployed config

The controller test at `cmd/gc/controller_test.go:2234` currently asserts "Maintenance pack orders (always included)" — this also needs updating when Maintenance orders become Core orders.

### [Major] Doctor message says "supplied implicitly" — contradicts "required system pack"

`cmd/gc/import_state_doctor_check.go:194` tells operators that Maintenance/Core is "supplied implicitly." After migration, Core is a *required* system pack — not implicit, not optional. The corresponding test assertions at `import_state_doctor_check_test.go:427,443` repeat this: "maintenance/core is implicit."

This is a terminology inversion that will confuse operators. Core should be described as "required and auto-included," not "implicit." The doctor's `FixHint` at line 60 ("rewrite legacy gastown imports and remove legacy maintenance imports") also needs updating — after migration, the message should explain that Maintenance is *retired* and its behavior is now in Core, not just that it was "removed."

### [Major] The role-name guard scope is ambiguous — pack asset content vs. Go source only

The design states "Core assets must not contain Mayor, Deacon, Polecat, Refinery, Witness, Boot, Crew, or Gastown outside allowed docs/tests." But the current Core pack already contains role names in pack assets:

- `internal/bootstrap/packs/core/formulas/mol-polecat-base.toml` — mentions "Polecat" in description, formula name, and step content (references to "refinery," "witness," "polecat" at lines 2, 4, 14, 21, 79, 100, 205)
- `internal/bootstrap/packs/core/formulas/mol-polecat-commit.toml` — mentions "polecat," "refinery," "witness"
- `internal/bootstrap/packs/core/formulas/mol-do-work.toml:10-11` — "use mol-polecat-work from the gastown pack"
- `internal/bootstrap/packs/core/skills/gc-dispatch/SKILL.md:24-25,30,43,85,92-93,98,101` — mentions "mayor," "polecat," "refinery"
- `internal/bootstrap/packs/core/formulas/mol-review-quorum.toml` — role-neutral, which the design acknowledges as an audit item

The design says "move `mol-polecat-base`, `mol-polecat-commit`, and `mol-polecat-report` to `gascity-packs/gastown`," which addresses the formula content. But the design also needs to clarify whether the guard scans only Go source code or extends to:
- Pack asset filenames (e.g., `mol-polecat-base.toml`)
- Prompt content and skill docs (e.g., SKILL.md references)
- Formula variable defaults
- Script content (e.g., `jsonl-export.sh`'s "mail mayor" escalation)

Without this boundary, implementers will reasonably disagree about whether `mol-do-work.toml`'s "use mol-polecat-work from the gastown pack" cross-reference is a ZFC violation or acceptable documentation.

---

## Additional findings

### Doc references to retired paths and terminology

The design lists several doc files needing updates but misses these:

- `docs/guides/shareable-packs.md:103-104` — The `[imports.maintenance]` example with `source = "../maintenance"` references a retired pack. After migration, this example should use a different pack or explicitly note that Maintenance is retired.
- `docs/schema/pack-schema.json:572,872` and `docs/schema/pack-schema.txt:872` — The `includes` description uses `"../maintenance"` as an example, and the `source` description mentions `gastown` as a relative import example. These need updating.
- `docs/troubleshooting/gc-start-walkthrough.mdx:134-135` — The error message example references `.gc/system/packs/gastown/agents/mayor/agent.toml`. After Gastown becomes an external import, the path structure changes to a remote-pack cache layout, and this diagnostic should reflect the new resolution.
- `docs/reference/cli.md:2450` — `gc rig add ./my-project --include packs/gastown` uses a local path. After migration, new rigs should use the public pack source, not a local path.
- `docs/plans/dolt-port-resolve-helper.md:4,17,70,139,142` — Multiple references to "maintenance pack" and `.gc/system/packs/maintenance/` in an active plan. This plan needs terminology alignment.
- `examples/gastown/city.toml:5,10,15` — Comments reference "maintenance — generic infrastructure" and "Maintenance still supplies the fallback dog shape." These will be misleading after migration.
- `examples/gastown/packs/gastown/assets/scripts/status-line.sh:15-16` — Hardcodes `.gc/system/packs/maintenance/assets/scripts/_bd_trace.sh`. This script needs to use `GC_SYSTEM_PACKS_DIR` or the Core pack path after migration.
- `examples/gastown/packs/maintenance/assets/scripts/dolt-target.sh:153` — Uses `GC_SYSTEM_PACKS_DIR` for `dolt` but the fallback hardcodes `.gc/system/packs/`. The migration path for this script (moving to Core) needs to use the proper env var.

### Tutorial and troubleshooting terminology

- `docs/tutorials/01-cities-and-rigs.md:189` — "The `dog` pool is a background utility agent from the built-in maintenance pack." Should become "from the Core pack" with a note that `dog` is configurable.
- `docs/getting-started/troubleshooting.md:241` — "The maintenance pack runs `jsonl-export` every 15 minutes" — after migration, this is Core, not "the maintenance pack."
- `docs/getting-started/troubleshooting.md:275` — `ARCHIVE=$(gc config get state_dir)/packs/maintenance/jsonl-archive` — this path will change to `packs/core/jsonl-archive`.
- `docs/getting-started/troubleshooting.md:351` — "`GC_JSONL_MAX_PUSH_FAILURES=99` in the maintenance pack's environment" — should reference "Core pack" after migration.

### Test assertions referencing retired pack names

- `cmd/gc/embed_builtin_packs_test.go:165-167` — Asserts Maintenance order paths like `maintenance/orders/gate-sweep.toml`. These become `core/orders/gate-sweep.toml`.
- `cmd/gc/embed_builtin_packs_test.go:198-203` — Content needle assertions reference `"maintenance"` pack paths.
- `cmd/gc/embed_builtin_packs_test.go:669,700,724,783,807` — Multiple assertions reference `maintenance/orders/` and `maintenance/formulas/` paths.
- `cmd/gc/embed_builtin_packs_test.go:1253-1332` — `builtinPackIncludes` tests assert "core + maintenance" and "core + maintenance + bd" membership. After migration, these should assert "core" (which now includes former Maintenance content) without a separate "maintenance" pack.
- `cmd/gc/controller_test.go:2234-2237` — Comment and assertions reference "Maintenance pack orders (always included)." After migration, these are Core orders.
- `cmd/gc/import_state_doctor_check_test.go:427,443` — "maintenance/core is implicit" terminology needs updating.

### Pack comments and examples

- `examples/gastown/packs/maintenance/pack.toml:1,8,11` — Pack comment says "Maintenance — generic multi-agent infrastructure pack" and "No rig-scoped agents — maintenance is global infrastructure." The name `maintenance` and these comments will be misleading if assets move to Core.
- `examples/gastown/packs/maintenance/doctor/check-binaries/doctor.toml:1` — Description says "Verify required maintenance binaries are available." After moving to Core, this should say "required Core maintenance binaries" or similar.
- `examples/gastown/packs/maintenance/formulas/mol-dog-reaper.toml:80` — "until the maintenance pack has an explicit DB-to-`BEADS_DIR` routing map." After migration, this should reference "Core" not "the maintenance pack."
- `examples/gastown/packs/gastown/pack.toml:5-6,29` — Still imports `../maintenance` and patches `dog`. After migration, this import and patch need to target Core or define Gastown's own dog.
- `examples/gastown/pack.toml:3,6,14-15` — References gastown pack at local path `packs/gastown`. After migration, this example should import the public Gastown pack.
- `internal/builtinpacks/registry.go:56,128` — Registry lists "maintenance" as a builtin pack and in `publicSubpathForPack` maps both "gastown" and "maintenance" as public packs. The Maintenance entry must be removed and the gastown public subpath remapped.

### Canonical wording matrix

The design should define consistent terminology and wire it into doctor messages, FixHints, docs, and pack comments. At minimum:

| Concept | Canonical wording | Context |
|---|---|---|
| Core pack | "Core is a required system pack" | Doctor messages, FixHints |
| Core inclusion | "auto-included by the system" (not "implicit") | Doctor messages, comments |
| Gastown pack | "Gastown is an explicit public pack import" | Docs, pack comments, init output |
| Maintenance retirement | "Maintenance is retired; its behavior moved to Core and Gastown" | Doctor removal messages, migration docs |
| Store maintenance / Dolt maintenance | "store maintenance" and "Dolt maintenance" | Troubleshooting, script docs |
| `dog` agent | "Core's default configurable maintenance agent" | Tutorials, comments |

---

## Questions

- Should `jsonl-export.sh` and `spawn-storm-detect.sh` use `$GC_PACK_STATE_DIR` as their only state path (removing the `packs/maintenance` hardcoded fallback), or should they retain a legacy fallback for existing cities?
- Is `[[patches.agent]] name = "dog"` expected to work across required-but-not-imported system packs, or does Gastown need its own `dog` agent after migration?
- For order qualified names, is the intent that deployed cities silently accept both `maintenance.gate-sweep` and `core.gate-sweep`, or that `gc doctor --fix` rewrites old qualifiers?
- Are formula filenames like `mol-polecat-base.toml` in scope for the role-name guard, or are they configuration artifacts exempt from the Go-source guard?
- Should `jsonl_archive_doctor_check.go` gain a dual-path lookup during the transition, or should existing `packs/maintenance/jsonl-archive` state be migrated by `gc doctor --fix`?
- What should happen to `examples/gastown/city.toml` — should it become a non-local-import example referencing the public Gastown pack, or remain a local-development fixture with a comment explaining it's for in-tree testing only?
