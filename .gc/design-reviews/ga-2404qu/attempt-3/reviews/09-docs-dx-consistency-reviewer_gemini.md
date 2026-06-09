# Felix Moreau — DeepSeek V4 Flash (Docs & DX Consistency Review, Attempt 3)

**Verdict:** block

**Scope:** Cross-file consistency, operator terminology drift, missed edge cases in runtime-state migration, pattern drift between code and design, cross-pack contract correctness, and documentation completeness gaps that prior reviews partially flagged but left under-specified.

---

## Top strengths

- The design correctly identifies `docs/tutorials/01-cities-and-rigs.md:189` as the highest-signal terminology hazard and proposes reframing `dog` as Core's configurable maintenance agent.
- The import-state doctor messaging is explicitly in scope, with `cmd/gc/import_state_doctor_check.go:194`'s "supplied implicitly" wording identified for replacement.
- The rollback section acknowledges order-skip-list name preservation as a real operator concern.
- The behavior-inventory table (`requirements.md:317–368`, `design.md:149–157`) is thorough at the file level.

---

## Critical risks

### [Blocker] Runtime-state migration is still underspecified — three distinct failure modes not addressed

Prior reviews flagged that `.gc/runtime/packs/maintenance/` state is unaddressed. This remains the most dangerous gap, but the specifics deserve sharper enumeration:

1. **Silent data bifurcation.** After migration, scripts moved to Core will write new state to `.gc/runtime/packs/core/` (via `GC_PACK_STATE_DIR` set by `citylayout.PackRuntimeEnv(cityPath, "core")`). Existing cities retain JSONL archive repos, state files, and storm-count ledgers under `.gc/runtime/packs/maintenance/`. If `jsonl_archive_doctor_check.go` is updated to look only under `packs/core/`, operators lose visibility into their existing archive. If it keeps looking only under `packs/maintenance/`, new archives after migration are invisible. The script `jsonl-export.sh` already has a `LEGACY_ARCHIVE_REPO` and `LEGACY_STATE_FILE` fallback pattern, but the Go doctor check at `cmd/gc/jsonl_archive_doctor_check.go:50–110` does **not** — its `resolveStateFile()` and `resolveArchiveRepo()` hardcode `filepath.Join(runtime, "packs", "maintenance")` as the fallback path and have no dual-path lookup. The design must specify either: (a) dual-path lookup in the doctor check during transition, (b) in-place migration of existing state by `gc doctor --fix`, or (c) explicit manual-migration instructions in the release notes.

2. **`spawn-storm-detect.sh` has no legacy fallback at all.** While `jsonl-export.sh` has `LEGACY_ARCHIVE_REPO`/`LEGACY_STATE_FILE`, `spawn-storm-detect.sh` only has `PACK_STATE_DIR="${GC_PACK_STATE_DIR:-${GC_CITY_RUNTIME_DIR:-$CITY/.gc/runtime}/packs/maintenance}"`. After migration, existing storm-count ledgers under `packs/maintenance/` will never be read because the controller will set `GC_PACK_STATE_DIR` to `packs/core`. This is a silent state drop for any city that has accumulated storm counts.

3. **The `dolt-target.sh` helper in Maintenance references `packs/dolt` state.** Line 30: `DOLT_STATE_DIR="${GC_CITY_RUNTIME_DIR:-$GC_CITY_PATH/.gc/runtime}/packs/dolt"`. This isn't a `maintenance→core` migration issue, but it shows the pattern: every script that hardcodes a pack name in a state path has the same migration problem. The design should enumerate **all** hardcoded `packs/<name>` state paths across Maintenance scripts and specify whether each migrates, dual-reads, or is retired alongside the pack.

### [Blocker] The Gastown→Core dependency is implicit, uncheckable, and has no failure mode

Prior reviews noted this, but the specific technical gap deserves more precision. After migration:

- Gastown's `pack.toml` currently has `[imports.maintenance]` at line 19, with `source = "../maintenance"`.
- The design says "Gastown should not import Core; Core remains a required host system pack" (`design.md:203–204`).
- Gastown's `[[patches.agent]] name = "dog"` at `pack.toml:29` patches an agent defined in the (retired) Maintenance pack, which will now be defined in Core.
- Gastown's `[[named_session]]` entries at lines 33–48 reference `mayor`, `deacon`, `boot`, `witness`, `refinery` — all Gastown-local, fine.
- Gastown's `formulas/mol-polecat-work.toml` extends `mol-polecat-base` which is currently in Core.

The problem: there is **no manifest mechanism** for Gastown to declare "I require host Core providing agents {dog} and formulas {mol-polecat-base, mol-polecat-commit}." If a city imports Gastown but somehow lacks Core (a misconfiguration, a future test city, a minimal dev setup), the following break silently:

- The `[[patches.agent]] name = "dog"` patch fails to resolve — `dog` is not found in any imported pack, and there is no `dog` in Gastown itself.
- The `extends = ["mol-polecat-base"]` in `mol-polecat-work.toml` fails formula resolution because `mol-polecat-base` is only in Core.
- Gastown formulas that reference `$PACK_DIR` resolve to Gastown's pack directory, but `$GC_PACK_STATE_DIR` for housekeeping orders resolves to Core's state directory.

The design must either: (a) add a declared `[dependencies.core]` or `[requires]` mechanism to pack.toml schema, (b) add a runtime doctor check that validates that all Gastown formula `extends` and agent `patches` resolve against the city's actually-loaded packs, or (c) have Gastown define its own `dog` fallback (with documentation that it's a Core overlay, not a replacement).

### [Blocker] The `publicSubpathForPack` mapping and `builtinpacks.All()` are inconsistent with the target state

`internal/builtinpacks/registry.go:133` has:
```go
func publicSubpathForPack(name string) (string, bool) {
    switch name {
    case "gastown", "maintenance":
        return name, true
    default:
        return "", false
    }
}
```

After migration, `maintenance` is retired as a standalone pack and `gastown` is an external public pack. This function currently maps **both** `gastown` and `maintenance` to themselves as public subpaths. But:

1. `maintenance` will no longer exist as a builtin pack after the registry entry is removed. The `publicSubpathForPack("maintenance")` branch becomes dead code that can never match a valid builtin pack name, yet synthetic cache construction still includes it.
2. `gastown` is currently a builtin pack sourced from `examples/gastown/packs/gastown`, but the design says it should become an external public pack at `gascity-packs/gastown`. The synthetic cache for Gastown will need to point at `PublicRepository` + `gastown` subpath instead of `Repository` + `examples/gastown/packs/gastown`.
3. `All()` currently returns a pack with `{Name: "maintenance", Subpath: "examples/gastown/packs/maintenance", FS: maintenance.PackFS}`. Removing this entry without updating the embedding (importing the `maintenance` package and referencing `maintenance.PackFS`) will break the build.

The design must specify: (a) remove the `maintenance` import and `PackFS` reference from `registry.go`, (b) remove the `maintenance` entry from `All()`, (c) either remove `gastown` from `All()` (making it purely external) or keep it as a builtin with the new external-pack source, (d) update `publicSubpathForPack` to reflect the new state, and (e) update `syntheticPackLayouts()` and `NameForSource()` so that existing synthetic caches with the old `examples/gastown/packs/maintenance` subpath are recognized as stale and regenerated.

### [Major] The role-name guard surface is broader than prior reviews identified

Prior reviews flagged `mayor`/`deacon` in `reaper.sh`, `mol-dog-jsonl`, and `mol-shutdown-dance`. Additional contamination in current Core-bound assets that the design marks as `core`:

- **`internal/bootstrap/packs/core/formulas/mol-prompt-synth.toml`** references failure notification to "Witness" (lines 25–27 of the description).
- **`internal/bootstrap/packs/core/skills/gc-dispatch/SKILL.md`** references `polecat`, `refinery`, "the polecat's", and Gastown-specific dispatch patterns.
- **`internal/bootstrap/packs/core/formulas/mol-do-work.toml`** references `mol-polecat-work` from the "gastown pack" and `refinery review` — these are Gastown-specific workflow references in what should be a role-neutral Core formula.
- **`internal/bootstrap/packs/core/formulas/mol-polecat-base.toml`** and **`mol-polecat-commit.toml`** contain the string "polecat" in their filenames and descriptions. The design marks these as `core`, but the formula names embed a Gastown role. Either they need renaming to role-neutral names (breaking change for order skip lists), or the design needs to explicitly grandfather `mol-polecat-*` as pre-existing formula IDs that are not role names in the guard's sense.

The role-name guard specification must enumerate the full file-type surface: TOML formula filenames, TOML `[vars].*.default` values, embedded bash command text in step descriptions, `.sh` scripts, `.template.md` prompt fragments, `SKILL.md` skills, `doctor.toml` descriptions, `agent.toml` metadata, and overlay JSON. A Markdown-only guard will miss all of these.

### [Major] `examples/gastown/city.toml` and `examples/gastown/pack.toml` are not addressed as migration targets

The design focuses on `examples/gastown/packs/*` but does not address:

- `examples/gastown/city.toml:15` states "Maintenance still supplies the fallback dog shape and shared dog formulas/prompts that gastown reuses." This comment will be false after migration.
- `examples/gastown/city.toml:8-9` — the three-pack comment block calls Maintenance "generic infrastructure." After retirement, this needs rewriting.
- `examples/gastown/pack.toml:5-6` — `source = "packs/gastown"` is a local path import. The design says Gastown should be an explicit public pack import, but this example file still models a local path. Should it become a non-local-import example referencing the public `gascity-packs/gastown` pack?
- `examples/gastown/pack.toml:19-20` — `[imports.maintenance]` and `source = "../maintenance"` — this import must be removed, but doing so also removes the mechanism by which Gastown resolves `dog` and the `propulsion-dog` / `architecture` / `following-mol` template fragments. The design needs to specify where Gastown gets these template fragments from after the import is removed (they're currently in `examples/gastown/packs/maintenance/template-fragments/`).

### [Major] Template-fragment ownership is unspecified after the Maintenance→Core move

The `dog` prompt at `examples/gastown/packs/maintenance/agents/dog/prompt.template.md` uses three template fragments:
- `{{ template "propulsion-dog" . }}` (from `maintenance/template-fragments/propulsion.template.md`)
- `{{ template "architecture" . }}` (from `maintenance/template-fragments/architecture.template.md`)
- `{{ template "following-mol" . }}` (from `maintenance/template-fragments/following-mol.template.md`)

Gastown's `propulsion.template.md` defines `propulsion-mayor`, `propulsion-deacon`, `propulsion-witness`, `propulsion-refinery`, `propulsion-polecat`, `propulsion-crew`, and `propulsion-dog`. After migration:

- The dog-specific `propulsion-dog` must move to Core (since `dog` moves to Core).
- The Gastown role-specific templates (`propulsion-mayor`, etc.) must stay in Gastown.
- But `propulsion.template.md` is a single file in Gastown that defines all variants. Splitting it means either: (a) Gastown's `propulsion.template.md` keeps `propulsion-dog` as a duplicate of the Core version, (b) Core defines its own `propulsion-dog` and Gastown depends on the host Core to provide it, or (c) `propulsion-dog` is generalized into a `propulsion-worker` that both Core and Gastown can use.

Similarly, `architecture.template.md` and `following-mol.template.md` in Maintenance contain both generic content and Gastown-specific references (`Dogs run cleanup`, `Gas City Maintenance Context`). The design says to "move SDK-generic content into Core" but does not specify the template-fragment splitting strategy.

### [Major] The `requiredBuiltinPackNames` comment is stale and will mislead implementers

`cmd/gc/embed_builtin_packs.go:236` currently reads:
```go
func requiredBuiltinPackNames(cityPath string) []string {
    required := []string{"core", "maintenance"}
```

The comment at line 34 says "Required packs (core, maintenance, and the provider-dependent bd/dolt)". After migration, `maintenance` is removed from the required list and folded into `core`. But the design doesn't call out that this comment must be updated in lockstep, and the code change from `[]string{"core", "maintenance"}` to `[]string{"core"}` has cascading effects on:

- `builtinPackIncludes()` which appends paths for these required packs.
- `requiredBuiltinPackSet()` which determines which packs get force-refreshed.
- The `MaterializeBuiltinPacks` comment about "operator edits are preserved only for non-required packs."
- Test assertions in `embed_builtin_packs_test.go` that assert `"core + maintenance"` membership.

The design's test section (`design.md:376`) says to test that `builtinPackIncludes` yields "Core and bd" (non-bd) or "Core only" (no provider), which is correct, but does not call out the comment/commentary updates needed in `embed_builtin_packs.go` itself.

### [Minor] `internal/hooks/hooks.go` still imports `internal/bootstrap/packs/core`

`internal/hooks/hooks.go:23` imports `"github.com/gastownhall/gascity/internal/bootstrap/packs/core"` and references `core.PackFS` at lines 174, 177, 185, and 782. After the migration moves Core to `internal/packs/core`, this import path must change. The design mentions updating `internal/hooks/hooks.go` and `internal/builtinpacks/registry.go` to import `internal/packs/core`, but the hooks file has four separate `core.PackFS` references that all need the new import, and the embed directive pattern (`//go:embed pack.toml all:assets formulas orders all:overlay skills`) must be verified to match the new `internal/packs/core` directory structure.

### [Minor] `internal/bootstrap/bootstrap.go` embeds `packs/**` globally and tests override `BootstrapPacks` with `AssetDir: "packs/core"`

The design says `internal/bootstrap/packs/core` will be "removed or reduced to a compatibility shim." But:

- `internal/bootstrap/bootstrap.go:46` has `//go:embed packs/**` which embeds everything under `internal/bootstrap/packs/`, including the current `core/` directory.
- Several tests (e.g., `internal/bootstrap/bootstrap_test.go`, `internal/bootstrap/collision_test.go`) reference `AssetDir: "packs/core"` or test against bootstrap pack structures.
- The `RetiredBootstrapPacks` list in `bootstrap.go` has the old retired packs (`gc-core`, `gc-import`, `gc-registry`) but does not include any new retired entries for the maintenance pack.

The design must specify: (a) whether the `//go:embed packs/**` directive is removed entirely (since `internal/packs/core` will have its own `embed.go`), (b) whether `BootstrapPacks` is now always empty (the design says "BootstrapPacks is empty in production"), and (c) whether `packs/core` under bootstrap becomes an empty directory or is deleted entirely.

### [Minor] The `examples/dolt/pack.toml` dependency on Maintenance is acknowledged but not in the update plan

`requirements.md:384` flags `examples/dolt` as a downstream reference, and `requirements.md:388` lists it under "Downstream References To Update." But the design's update plan (`design.md:332–358`) does not include `examples/dolt/pack.toml` in its file-by-file list. The dolt pack currently says "Dog-backed formulas and orders rely on the city's maintenance pack" (`examples/dolt/pack.toml:6`). This comment must be updated to reference Core.

---

## Cross-pack contract issues not previously detailed

### The `[[patches.agent]]` cross-pack resolution needs a semantics specification

The Gastown pack currently patches `maintenance.dog` via:
```toml
[[patches.agent]]
name = "dog"
wake_mode = "fresh"
work_dir = ".gc/agents/dogs/{{.AgentBase}}"
```

After migration, `dog` comes from Core (a required system pack, not an explicit import). The config loader must resolve `name = "dog"` in a `[[patches.agent]]` block against the agent's source pack. Currently, patches resolve against the pack's explicit imports. Gastown does not (and per the design, should not) import Core. This means either:

1. The config loader gains new semantics: patches can target agents from required-but-not-imported system packs.
2. Gastown defines its own `dog` agent that overrides Core's — but this contradicts the design's intent that Gastown should "patch" Core's dog for theming.
3. Gastown imports Core explicitly — contradicting `design.md:203–204`.

The design must specify which semantics the config loader will support and add a corresponding test case.

### Formula `extends` cross-pack resolution

`examples/gastown/packs/gastown/formulas/mol-polecat-work.toml` has `extends = ["mol-polecat-base"]`. Currently this resolves because Gastown imports Maintenance, which imports nothing, but `mol-polecat-base` is in Core. After migration, with no import of Core, this formula resolution fails unless the config loader resolves extends against all loaded packs (including required system packs). The design must specify the resolution order and test it.

---

## Canonical wording matrix

The design should define and wire one shared terminology set. At minimum:

| Concept | Canonical wording | Surfaces in |
|---|---|---|
| Core pack | "Core is a required system pack" | Doctor messages, FixHints, pack comments |
| Core inclusion | "auto-included by the system" (not "implicit" or "supplied implicitly") | Doctor messages, `requiredBuiltinPackNames` comment |
| Gastown pack | "Gastown is an explicit public pack import" | Docs, pack comments, init output |
| Maintenance retirement | "Maintenance is retired; its behavior moved to Core and Gastown" | Doctor removal messages, migration docs |
| Store/Dolt maintenance | "store maintenance" and "Dolt maintenance" | Troubleshooting, script docs |
| `dog` agent | "Core's default configurable maintenance agent" | Tutorials, comments, prompt templates |
| Runtime state path | `.gc/runtime/packs/core/` for new state; `.gc/runtime/packs/maintenance/` is legacy | Scripts, doctor checks, troubleshooting |

---

## Questions

- Should `jsonl-export.sh` and `spawn-storm-detect.sh` use `$GC_PACK_STATE_DIR` exclusively (removing the `packs/maintenance` fallback), or retain a legacy fallback for existing cities? If the latter, for how many releases?
- Should `jsonl_archive_doctor_check.go` gain a dual-path lookup during the transition, or should `gc doctor --fix` migrate existing state from `packs/maintenance/` to `packs/core/`?
- Is `[[patches.agent]] name = "dog"` expected to work across required-but-not-imported system packs, or does Gastown need its own `dog` agent after migration?
- For order qualified names, is the intent that deployed cities silently accept both `maintenance.gate-sweep` and `core.gate-sweep`, or that `gc doctor --fix` rewrites old qualifiers?
- Should `examples/gastown/city.toml` become a non-local-import example referencing the public Gastown pack, or remain a local-development fixture with a comment explaining it's for in-tree testing only?
- Are formula filenames like `mol-polecat-base.toml` in scope for the role-name guard, or are they pre-existing formula IDs exempt from the guard's prohibition on Gastown role names?
- What is the template-fragment splitting strategy for `propulsion.template.md`, `architecture.template.md`, and `following-mol.template.md`?
- Should `BootstrapPacks` remain empty in production, and should the `//go:embed packs/**` directive in `bootstrap.go` be removed entirely?

---

## Summary of evidence

All findings are grounded in specific file paths, line numbers, and code quotations from the current source tree. The most impactful risks are:

1. **Data continuity** (`packs/maintenance/` runtime state with no migration plan)
2. **Cross-pack contract** (patches and formula extends with no resolution semantics for required-but-not-imported packs)
3. **Registry consistency** (`builtinpacks.All()` and `publicSubpathForPack` stale entries)
4. **Role contamination** in Core-bound assets beyond what prior reviews enumerated
5. **Template-fragment ownership** (single-file definitions that span Core and Gastown ownership)
6. **Comment/code drift** (requiredBuiltinPackNames, city.toml, pack.toml comments all describing retired state)
