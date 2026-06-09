# Gastown Pack Boundary Review — Independent DeepSeek Analysis

**Verdict:** block

**Reviewer focus:** Cross-document consistency, missed edge cases, assumptions other reviewers may accept too quickly, and architectural coherence between requirements, design, and runtime code.

---

## Top Strengths

1. **The maintenance retirement runtime table is a significant improvement.** Mapping each runtime surface (`requiredBuiltinPackNames`, `builtinPackIncludes`, `publicSubpathForPack`, `MaterializeBuiltinPacks`, orders, skip lists, runtime state, doctor, and synthetic cache) to target behavior and required proof makes the migration auditable. This is the strongest structural addition in the revised design.

2. **Behavior-preservation-first gating is now a blocking prerequisite.** Requiring an immutable pinned commit, behavior inventory, and Gas City CI proof before any in-tree source removal is the correct ordering. Previous iterations left this as a retrospective check.

3. **The doctor safety contract is well-specified.** Preflight-before-mutation, scoped TOML edits, failure-atomic multi-file behavior, fork/custom-source preservation, and byte-identical healthy-city no-op address the most dangerous mutation path in the migration.

---

## Critical Risks

### [Blocker] Gastown pack.toml imports `../maintenance` and the design does not specify whether this import must be removed from the public Gastown pack

The current `examples/gastown/packs/gastown/pack.toml` (line 19) declares `[imports.maintenance]` with `source = "../maintenance"`. The design states that "Gastown should not import Core; Core remains a required host system pack" (`design-after.md:258-260`) and that "Gastown may continue to patch the Core `dog` agent for theming or work_dir behavior" (`design-after.md:263-264`).

This creates a three-way inconsistency:

- The current Gastown pack depends on Maintenance for dog definition, template fragments (`propulsion-dog`, `architecture`, `following-mol`), and session_live hooks (tmux-theme.sh, tmux-keybindings.sh).
- The design says Gastown should not import Core (the successor to Maintenance).
- The design says Gastown may patch Core's dog agent — but `[[patches.agent]] name = "dog"` in Gastown's pack.toml can only patch agents defined in imported packs or auto-included packs. Whether this resolves correctly depends on whether the config layering system treats auto-included system packs (via `builtinPackIncludes`) as patchable by non-importing packs.

The design asks for "the public `gascity-packs/gastown/pack.toml` should be updated so comments no longer describe an implicit Maintenance layer" (`design-after.md:258-262`) but does not require removing the `[imports.maintenance]` binding itself. If the public Gastown pack retains a `maintenance` import pointing at a retired pack, it fails to load. If it removes the import but doesn't add a Core import, it can't reference Core's dog or template fragments. If it adds a Core import, it contradicts "Gastown should not import Core."

**Required change:** The design must specify which of these is correct: (a) Gastown imports Core explicitly and the "should not import Core" rule is relaxed to "should not duplicate Core's auto-inclusion," (b) Gastown does not import Core and patches dog through config-layering over the auto-included system pack, or (c) Gastown supplies its own dog and dog-related assets without Core reference. Each option has different behavior-preservation implications, and the design currently specifies none of them.

### [Blocker] `builtinPackIncludes` silently loads stale packs that have `pack.toml` files, creating an invisible config-layering dependency

The current `builtinPackIncludes` function (`cmd/gc/embed_builtin_packs.go:272-283`) iterates `requiredBuiltinPackNames` and adds the system pack path only if `packExists` returns true — i.e., only if `.gc/system/packs/{name}/pack.toml` exists. After the migration, `requiredBuiltinPackNames` returns `["core"]` (plus provider packs), so maintenance is excluded from the includes list.

However, there is a secondary include path: `LoadWithIncludes` in config loading also processes `[imports]` from `pack.toml` and `city.toml`. If a stale `.gc/system/packs/maintenance/` directory contains a `pack.toml`, and any existing config file or import still references it transitively, the config layering will still load it. The design's Maintenance Retirement Runtime Table says `builtinPackIncludes` should "ignore stale generated Maintenance/Gastown directories," but the mechanism described is simply removing them from `requiredBuiltinPackNames`. This doesn't prevent config from discovering stale maintenance through its own import resolution if any surviving reference exists.

Conversely, the design says stale directories "may remain as ignored legacy state and are not deleted by `gc doctor --fix`" (`design-after.md:735-737`), but a stale `.gc/system/packs/maintenance/pack.toml` that is loaded via import resolution is not "ignored" — it's an active config layer with its own orders, formulas, and agent definitions that could conflict with the new Core-pack equivalents.

**Required change:** The design must specify: (a) that `gc doctor --fix` removes any `[imports.maintenance]` binding and any transitive maintenance import from pack.toml and city.toml (not just legacy public-pack sources), and (b) that config loading must not resolve stale system-pack directories as import sources even if a `pack.toml` exists. The current `legacyPublicPackForSource` only handles the Gastown/maintenance sources it explicitly enumerates; it does not handle a stale system-pack directory discovered through `[imports]` resolution.

### [Blocker] `mol-shutdown-dance.toml` hardcodes `"deacon"` as the default requester — a ZFC violation that the role-token scanner must catch in TOML formula defaults

The design's Core Maintenance and Notification Contract says "no Go code may contain role names as a control decision" (`design-after.md:203-204`), and the role-token scanner should cover "Core TOML" (`design-after.md:222`). But the design does not specifically call out formula variable defaults as a ZFC concern.

`examples/gastown/packs/maintenance/formulas/mol-shutdown-dance.toml` defines:

```toml
[vars.requester]
description = "Who filed the warrant (deacon, witness, etc.)"
default = "deacon"
```

This `default = "deacon"` is a hardcoded Gastown role name in a Core-bound formula's TOML configuration. If `mol-shutdown-dance` moves to Core with this default, the Core pack encodes a Gastown-specific role assumption. The design says "detector/requester examples" should be "role-neutral" in Core (`design-after.md:210`), but a formula variable default is not an "example" — it's a runtime value that determines behavior when no override is provided.

Similarly, `mol-dog-jsonl.toml` and `mol-dog-reaper.toml` contain `gc mail send mayor/` and `gc session nudge deacon/` in their step descriptions. The design classifies these as `core-renamed`, which is correct, but the design's behavior inventory does not enumerate the `requester` default specifically.

**Required change:** Add to the behavior inventory: formula variable defaults that contain Gastown role names must either be made configurable (empty default, filled by formula caller), renamed to a generic name (`"maintenance-agent"`, `"patrol-agent"`), or moved to Gastown formula overlays. The role-token scanner must include TOML default values in its scan scope, and the design must specify the disposition of each formula variable default that contains a role name.

### [Blocker] The Gastown `session_live` hooks and dog prompt template fragments have no specified Core destination

The current Gastown pack (`pack.toml:23-25`) defines `session_live` hooks:

```toml
session_live = [
    "{{.ConfigDir}}/assets/scripts/tmux-theme.sh {{.Session}} {{.Agent}} {{.ConfigDir}}",
    "{{.ConfigDir}}/assets/scripts/tmux-keybindings.sh {{.Session}} {{.Agent}} {{.ConfigDir}}",
]
```

These reference `tmux-theme.sh` and `tmux-keybindings.sh` in the Gastown pack's `assets/scripts/`. The design's behavior inventory lists `tmux-theme.sh` and `cycle.sh` under Gastown-owned assets but does not specify where `session_live` configuration lands. If Gastown doesn't import Core and Core doesn't define `session_live`, who owns the session-live hooks for Gastown agents?

Similarly, the maintenance dog prompt references three template fragments from `examples/gastown/packs/maintenance/template-fragments/`:

| Fragment | Content |
|---|---|
| `propulsion-dog` | Dog-specific propulsion instructions (`template-fragments/propulsion.template.md:203`) |
| `architecture` | Generic architecture guidance |
| `following-mol` | MOL-specific following instructions |

The design's Core Maintenance Notification Contract addresses prompt fragments generically ("Remove Gastown examples or move them to Gastown-owned assets"), but does not specify the disposition of each fragment. If `architecture` and `following-mol` move to Core, they must be role-neutral. If `propulsion-dog` moves to Gastown, the Core dog prompt can't `{{ template "propulsion-dog" . }}` without the fragment being available. If all three move to Core, they must be stripped of Gastown references.

**Required change:** The design must specify the disposition of each template fragment and the `session_live` hook configuration. At minimum: (a) which fragments are Core-bound and must be made role-neutral, (b) which fragments are Gastown-bound and must be provided by the Gastown pack, and (c) how the Core dog prompt resolves its `{{ template "propulsion-dog" . }}` reference when that fragment is owned by Gastown.

---

## Major Risks

### [Major] `GastownCity()` creates a config with no explicit Core import — the auto-inclusion path is an untested assumption

`internal/config/config.go:3712` defines `GastownCity()` which returns a `City` with an explicit `gastown` import but no `core` import. The design specifies that Core is auto-included via `builtinPackIncludes`. However, `GastownCity()` is a config constructor, not a config loader. If any code path creates a `GastownCity` config and loads it through a non-standard path (diagnostic, test, or init command that doesn't call `MaterializeBuiltinPacks` + `builtinPackIncludes`), Core would be absent from the resolved config.

The design's required-Core loading invariant addresses this at the command level, but `GastownCity()` itself is a Go function that can be called from anywhere. The design should specify whether `GastownCity()` needs an explicit `[imports.core]` entry for clarity, or whether the auto-inclusion contract is sufficient and should be tested as a unit-level invariant on `GastownCity()`.

### [Major] `legacyPublicPackForSource` matches `examples/` paths that may be intentional local dev setups

`cmd/gc/import_state_doctor_check.go:216-241` recognizes both `.gc/system/packs/{gastown,maintenance}` and `examples/gastown/packs/{gastown,maintenance}` as legacy public pack sources. After the migration, the `examples/gastown/packs/` directory structure changes (maintenance removed, gastown replaced by public import). But if a developer has a local checkout with `examples/gastown/packs/gastown` still present, `gc doctor --fix` would auto-rewrite that import to the public source — even if the developer intentionally pointed at their local checkout.

The design's doctor safety contract says "operator forks, edited local packs, or custom public sources are diagnostic/manual." But `legacyPublicPackForSource` matches `examples/gastown/packs/gastown` by path suffix, not by provenance. A local dev checkout IS an `examples/gastown/packs/gastown` path match. The current code doesn't distinguish between "generated system-pack Gastown" and "developer's local checkout Gastown."

### [Major] `pruneStaleGeneratedPackFiles` has no mechanism to detect stale directories from removed packs

After Maintenance is removed from `All()`, `MaterializeBuiltinPacks` never iterates over `.gc/system/packs/maintenance`, so `pruneStaleGeneratedPackFiles` never runs on it. This means stale generated files in that directory persist indefinitely — which is fine if they're truly "ignored." But the design doesn't address what happens when an old binary's generated files and a new binary's generated files coexist in `.gc/system/packs/`. If a user downgrades, the old binary would re-materialize maintenance (since the old binary still has it in `All()`). If they upgrade again, those re-materialized files would be stale but never cleaned up.

More importantly, the design says `MaterializeBuiltinPacks` "does not refresh or prune Maintenance/Gastown." But `pruneStaleGeneratedPackFiles` only runs within `MaterializeBuiltinPacks` for packs in `All()`. There is no separate cleanup pass for retired pack directories. The design should specify whether a one-time migration step should remove generated files from `.gc/system/packs/maintenance` (leaving any operator-created files), or whether the stale directory is truly permanent and should be documented as such.

### [Major] The design's rollout matrix lacks a case for `new binary + stale local Gastown import`

The release compatibility matrix (`design-after.md:685-694`) covers five combinations but omits: a new binary with an existing city that has a local `[imports.gastown]` pointing at `.gc/system/packs/gastown` (the bundled source, not the public pack). The design says doctor rewrites this to the public source, but what if the user hasn't run `gc doctor --fix` yet? Does the new binary's config loading resolve `.gc/system/packs/gastown` as a valid import even though it's no longer in `All()` and no longer materialized? If so, the stale pack's content could diverge from what the system expects. If not, the city fails to load with no clear diagnostic.

The `legacyPublicPackForSource` function would match this path, but only when doctor runs. Normal `gc start` config loading would attempt to resolve the import from the file path and might succeed (if the directory still exists) or fail (if `MaterializeBuiltinPacks` has stopped refreshing it and it's become stale). The design should specify the expected behavior for this case.

### [Major] Dog agent definition fields have no specified Core disposition

`examples/gastown/packs/maintenance/agents/dog/agent.toml` defines:

```toml
scope = "city"
fallback = true
nudge = "Check your hook for work assignments."
idle_timeout = "2h"
min_active_sessions = 0
max_active_sessions = 3
```

The design says `dog` is "a default configurable `dog` maintenance agent for Core pack operations" (`design-after.md:155-157`), and that the name is "a default pack configuration value." But it doesn't specify which of these fields are Core defaults and which are Gastown overrides. Currently, the Gastown pack patches dog with `wake_mode = "fresh"` and `work_dir`. If Core defines these base fields, Gastown's `[[patches.agent]]` must correctly overlay them. If the base definition stays in Gastown, it's not a Core agent.

The `fallback = true` field is particularly important: it means any unmatched session request can be routed to a dog agent. If Core defines this, it changes the semantics of every city — every city would have a fallback agent by default. This is a behavioral change that the design doesn't call out.

---

## Missing Evidence

- The design does not specify whether the public Gastown pack must remove its `[imports.maintenance]` binding and, if so, what replaces it (explicit Core import, config layering, or Gastown-owned dog definition).
- No behavior inventory row for `mol-shutdown-dance.toml` `default = "deacon"` or the `requester` variable default pattern across all Core-bound formulas.
- No disposition specified for `session_live` hooks in the Gastown pack — whether they stay in Gastown, move to Core as generic session hooks, or are replaced by per-provider configuration.
- No test for `builtinPackIncludes` when a stale `.gc/system/packs/maintenance/pack.toml` exists and a transitive import resolves it.
- No test specified for `GastownCity()` confirming that the generated config resolves with Core present in provenance.
- No specification of whether `legacyPublicPackForSource` should stop matching `examples/gastown/packs/` paths after the migration (since those local paths may be intentional dev setups, not generated system-pack paths).
- No one-time cleanup or advisory for `.gc/system/packs/maintenance` generated files (as opposed to operator-edited files) that will never be pruned again.

---

## Required Changes

1. **Resolve the Gastown import chain.** Specify whether the public Gastown pack imports Core, relies on auto-inclusion, or provides its own dog agent. Add a behavior inventory row and packcompat test for the chosen option.

2. **Add formula variable defaults to the role-token scanner scope.** Enumerate every TOML default that contains a role name (`requester`, mail/nudge targets in step descriptions) and specify whether it becomes configurable, generic, or Gastown-owned.

3. **Specify template fragment dispositions.** For each of `propulsion-dog`, `architecture`, and `following-mol`, state whether it moves to Core (and must be made role-neutral) or stays in Gastown (and the Core dog prompt must not reference it). Add a behavior inventory row for the dog prompt's template references.

4. **Specify `session_live` hook ownership.** State whether Gastown `session_live` hooks stay in the Gastown pack's `pack.toml` (which doesn't import Core) or are injected through config layering.

5. **Test `builtinPackIncludes` with stale maintenance pack.toml present.** Prove that a stale `.gc/system/packs/maintenance/pack.toml` does not get loaded as a config layer through import resolution or any other path.

6. **Clarify `legacyPublicPackForSource` path matching.** Specify whether `examples/gastown/packs/` paths should still be auto-rewritten after the migration, or whether only `.gc/system/packs/` paths should be treated as legacy generated sources.

7. **Specify one-time stale directory advisory.** Document that `.gc/system/packs/maintenance` generated files will persist indefinitely and add a doctor advisory that reports stale generated system-pack directories. State whether `pruneStaleGeneratedPackFiles` should be extended to also clean retired-pack directories (with operator-file preservation), or whether this is explicitly out of scope.

8. **Add a release matrix case for `new binary + stale local Gastown import`.** Specify whether config loading succeeds (resolving from the stale directory), fails (with a diagnostic), or triggers doctor.

9. **Specify dog agent.toml field dispositions.** State which fields are Core defaults and which are Gastown patches. Add a test that a Core-only city with a renamed maintenance agent still loads and runs SDK infrastructure.

---

## Questions

- Does the config layering system allow a pack to `[[patches.agent]]` an agent defined in a pack it doesn't explicitly import but that is auto-included as a system pack? This determines whether Gastown can patch Core's dog without importing Core.

- Should `GastownCity()` add an explicit `[imports.core]` entry for safety, or is the auto-inclusion contract considered sufficient? If the latter, what test proves that `GastownCity()` config always resolves with Core present?

- Is `fallback = true` in the dog agent definition intended to be a Core default? If so, what happens in a Core-only city that doesn't define a dog agent — does the lack of a fallback agent change session routing behavior?

- After the migration, will `examples/gastown/packs/maintenance/` be removed from the repository entirely (since it's a pack source, not just a generated directory)? If so, `legacyPublicPackForSource` matching `examples/gastown/packs/maintenance` would only match historical paths in existing config files, not current development paths. This affects whether the `examples/` path matching should be retired.

- The design says `dolt-target.sh` stays in Core "for now, because Dolt remains a Core requirement until bd provider support is restored." Is there a plan or issue tracking the eventual removal of `dolt-target.sh` from Core, or is this a permanent exception?
