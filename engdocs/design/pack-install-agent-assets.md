---
title: Pack Install for Agent-Facing Assets
description: Proposed lightweight pack install flow for skills and MCP definitions without requiring a configured Gas City runtime.
---

# Pack Install for Agent-Facing Assets

Status: Proposed design draft.

## Summary

Gas City packs should be usable as a lightweight distribution format for
agent-facing assets, starting with provider-neutral skills and MCP server
definitions. In this mode, `gc` acts as a pack manager and installer. The user
should not need to configure a city, rig, formula, order, daemon, or
orchestration runtime just to install reusable assets into Claude, Codex,
Gemini, or another supported provider.

The recommended first CLI surface is:

```sh
gc pack install <pack-ref> [--target <provider>] [--scope user|project] [--dry-run] [--yes]
gc pack uninstall <pack-or-install-id> [--target <provider>] [--scope user|project] [--dry-run] [--yes]
```

The first implementation should keep the install engine conservative:
built-in installers for known provider targets, explicit plans before writes
when target inference is uncertain, and no pack-defined arbitrary
install/uninstall hooks.

## Problem / Why This Matters

Packs are already the portable unit for Gas City definitions, but the current
mental model still assumes a city/runtime context. That is too much setup for
users who only want to reuse agent-facing assets such as:

- skills that can work across providers using the Agent Skills convention
- MCP server definitions that can be projected into provider-native config
- later, small command or profile templates once their runtime implications are
  clearer

If installing a skill requires users to understand rigs, formulas, orders, and
orchestration before they see value, packs will miss a simpler onboarding path.
The pack format can carry useful assets before the user has opted into Gas
City as an orchestrator.

## Product Stance

`gc pack install` is proposed as a distribution/install mechanism, not a hidden
city bootstrap.

The product stance is:

- A pack can be installed for asset delivery without being run as a city.
- Stage 1 should focus on provider-neutral agent-facing assets: skills first,
  MCP definitions next or alongside if the provider projection is ready.
- The installer should not require `city.toml`, `.gc/`, rig registration,
  formulas, orders, or a running supervisor.
- Installing formulas or orders outside a city/runtime is not the first goal.
  They remain orchestration constructs.
- Commands and agent profiles/templates may become installable later, but they
  have sharper runtime constraints and should not block the initial proposal.
- Provider-as-pack work is separate. The first installer can ship with known
  provider targets instead of requiring provider extensibility to be solved.

This gives packs a two-stage product story:

1. **Install useful assets.** `gc` resolves a pack and installs its exported
   asset surface into selected provider targets.
2. **Run orchestrated systems.** A configured city/rig/runtime consumes the
   broader pack graph for agents, formulas, orders, and orchestration.

## Proposed CLI Surface

### Install

```sh
gc pack install <pack-ref>
gc pack install <pack-ref> --target codex
gc pack install <pack-ref> --target claude --scope user
gc pack install <pack-ref> --target gemini --scope project --dry-run
gc pack install <pack-ref> --target codex --yes
```

`<pack-ref>` should use the same source vocabulary as imports where practical:
local directory, `file://`, HTTPS git URL, SSH git URL, or a GitHub shorthand if
that is already supported by pack resolution. Registry aliases can be added
later if a public registry lands.

Recommended behavior:

- Resolve the pack to a stable local source, using existing pack cache
  mechanics where possible.
- Read `pack.toml` and the pack's exported installable surface.
- Build a plan describing exactly which provider target, scope, and files will
  be written or updated.
- If the target is inferred rather than specified, show the plan and require
  confirmation unless `--yes` is present and inference is unambiguous.
- Refuse to run pack-provided installer scripts in v1.
- Record enough install metadata for safe uninstall and upgrade.

### Uninstall

```sh
gc pack uninstall <pack-or-install-id>
gc pack uninstall <pack-or-install-id> --target codex
gc pack uninstall <pack-or-install-id> --target claude --scope user --dry-run
```

Recommended behavior:

- Remove only files or config blocks previously installed and recorded as
  managed by `gc pack install`.
- Never delete hand-authored provider config outside the managed block.
- Show a plan before deletion unless the target and install record are
  unambiguous and `--yes` is present.
- Leave pack cache pruning to a separate command or later phase.

### Optional Supporting Commands

The first implementation can be useful with only `install` and `uninstall`.
If needed, the following commands are additive and should stay pack-scoped:

```sh
gc pack install --dry-run <pack-ref>
gc pack list --installed [--target <provider>]
gc pack upgrade <pack-or-install-id> [--target <provider>]
```

Avoid `gc skill install` or `gc mcp install` as the primary shape. The user is
installing a pack's exported asset surface, not choosing each underlying asset
type as the root concept.

## Installable Surface Area

This table classifies the pack definition surface from an installer point of
view. "Requires Gas City intermediation" means the installed thing only has its
intended behavior when invoked through `gc`, a Gas City runtime, or a
Gas-City-managed projection step.

| Surface | Install-compatible? | Requires Gas City intermediation? | Proposed stance |
|---|---:|---:|---|
| `skills/` | yes, stage 1 | no after install | Primary v1 surface. Copy or symlink exported skills into provider skill sinks using built-in target installers. |
| `agents/<name>/skills/` | maybe, stage 1 with explicit export | no after install | Useful when a pack wants to expose named skill bundles. Do not blindly install every private agent-local skill. |
| `mcp/` | yes, stage 1 if projection support is ready | no after projection, yes during install | Primary or near-primary v1 surface. Convert neutral MCP TOML into provider-native managed config. |
| `agents/<name>/mcp/` | maybe, later or explicit stage 1 | no after projection, yes during install | Same caution as agent-local skills. Only install if the pack exports the entries as part of its public install surface. |
| `commands/` | later | yes | Commands require `gc` invocation/intermediation and namespace decisions. Keep out of the first asset-installer contract. |
| `agents/` / agent definitions | not as-is | yes | Full agents are orchestration/session constructs. Do not install into provider CLIs directly in v1. |
| Agent profiles/templates | later | maybe | Only useful if the target provider has a profile/template selection mechanism. Treat as a later provider-specific install surface. |
| `formulas/` | no for v1 | yes | Formulas create molecules and dispatch work inside a Gas City runtime. Keep them as city/orchestration constructs. |
| `orders/` | no for v1 | yes | Orders are schedules/gates over formulas and event state. Not meaningful as lightweight standalone installs. |
| `[providers.*]` | no for v1 | yes | Provider definitions configure Gas City sessions. Provider-as-pack is separate and should not block asset install. |
| `doctor/` | maybe later | yes | Doctor checks are pack operational tooling. They can be exposed through `gc doctor` in a city/runtime context, not installed into provider CLIs in v1. |
| `template-fragments/` | no standalone | yes | Prompt fragments matter when Gas City renders prompts. They may support later agent template/profile installs but are not independently useful. |
| `overlay/` | no for v1 | maybe | Overlay files can mutate project state broadly. Keep out until target ownership, conflict rules, and safety UX are clearer. |
| `patches/` | no | yes | Patches modify imported agent prompts in a composed pack graph. Not a standalone install surface. |
| `assets/` | only as dependencies | depends | Opaque support files may be copied when referenced by an installed skill or MCP definition, but `assets/` itself is not convention-installed. |
| `[imports.*]` | not directly | depends | Imports are dependency metadata. Imported content is installed only if exported through the installing pack's public install surface. |
| `[agent_defaults]` / named sessions | no for v1 | yes | These configure Gas City session behavior, not provider asset installation. |

## Pack Author Contract

The proposed authoring contract is intentionally small.

### Skills

A pack exposes installable skills by placing skill directories under
`skills/` and, optionally, by declaring which agent-local skills are part of
the public install surface.

Recommended default:

- `skills/<name>/SKILL.md` is public installable surface.
- `agents/<agent>/skills/<name>/SKILL.md` is private to that agent unless the
  pack explicitly exports it.
- Skill names are the directory names and should be valid for all supported
  provider sinks.

Open contract question: whether explicit exports live in `pack.toml` or in a
small per-surface manifest. A possible `pack.toml` shape is:

```toml
[install]
skills = ["code-review", "test-runner"]
mcp = ["github-tools"]

[[install.agent_skill]]
agent = "reviewer"
name = "reviewer-workflow"
as = "reviewer-workflow"
```

If no `[install]` block exists, the recommended v1 default is to export the
top-level `skills/` and `mcp/` catalogs only. That keeps common packs simple
while avoiding accidental exposure of private agent-local assets.

### MCP

A pack exposes installable MCP definitions by placing neutral MCP TOML files
under `mcp/`.

Recommended default:

- `mcp/<name>.toml` and `mcp/<name>.template.toml` are public installable
  surface.
- `agents/<agent>/mcp/<name>.toml` is private to that agent unless explicitly
  exported.
- Provider projection is performed by `gc`, not by pack-defined scripts.
- Provider-specific projection failures should be hard errors before writes.

The MCP install plan should show the neutral definition name, the projected
provider config file, and whether `gc` will create, merge, or update a managed
block.

### Ownership Metadata

Every install should record managed ownership outside the pack source, likely
under a user-level `gc` state directory. The record should include:

- pack source and resolved revision, if known
- installed pack identity and version
- target provider and scope
- installed skill directories or projected MCP entries
- content fingerprints or enough data to detect drift
- uninstall metadata for managed files/config blocks

Provider-visible files should also carry a lightweight marker where the target
format allows it. The source of truth for uninstall should be the `gc` install
record, not a best-effort filesystem scan.

## Imported Pack Behavior

The default stance should be explicit and closed.

Installing a pack installs that pack's exported installable surface. Imported
or transitive pack content is installed only when the installing pack exports
it under its own install surface. `gc pack install` should not walk arbitrary
private dependencies and spray their skills or MCP definitions into provider
targets by accident.

Recommended rules:

- Top-level `skills/` and `mcp/` in the requested pack are installable by
  default.
- Imported pack content is not installed just because it is reachable through
  `[imports.*]`.
- If a pack wants to redistribute an imported skill or MCP definition, it must
  re-export it explicitly or surface it under its own top-level installable
  catalog.
- The install plan should identify whether each asset is native to the pack or
  re-exported from an import.
- Name collisions are resolved within the exported install surface before any
  provider writes happen. Ambiguous collisions are hard errors.

This differs from city/runtime composition, where imports can contribute to a
resolved graph for orchestration. Asset install should optimize for safety and
predictability over maximum automatic transitivity.

## Target / Provider Model

The target is the provider destination that receives installed assets, not a
Gas City runtime provider definition. Stage 1 should use known built-in
installers for provider targets such as `claude`, `codex`, and `gemini` once
their current file locations and merge semantics are verified.

Recommended target flags:

```sh
--target claude|codex|gemini|...
--scope user|project
--project-dir <path>
--dry-run
--yes
```

Conservative default behavior:

- If `--target` is provided, install only for that target.
- If `--target` is omitted, detect safe candidates from the current directory,
  known provider config locations, and installed CLIs.
- If exactly one safe candidate is found, show the plan and ask for
  confirmation unless `--yes` is present.
- If multiple candidates are found, show choices and require `--target`.
- If no candidates are found, fail with examples rather than guessing.
- Never silently install into every provider that appears to be present.

Recommended scope behavior:

- `project` scope writes into the current project or `--project-dir` using the
  provider's project-local convention.
- `user` scope writes into the provider's user-level convention.
- If a provider lacks a safe project-local MCP surface, MCP install for that
  provider/scope should hard-error instead of falling back to a global write.
- Skill installs can prefer project scope when a project-local provider sink is
  known; otherwise require explicit user scope.

The plan output is part of the safety model. It should include:

- pack identity and resolved revision
- selected provider target and scope
- destination paths or config files
- files/config entries to create, update, leave unchanged, or remove
- conflicts requiring user action

## Non-Goals for the First Implementation

- No requirement to initialize or load a city.
- No rig registration, formula dispatch, orders, daemon, supervisor, or
  orchestration runtime.
- No install of formulas or orders as standalone user assets.
- No arbitrary pack-defined install/uninstall commands.
- No provider extensibility framework as a prerequisite.
- No silent transitive install of imported pack internals.
- No broad overlay installation.
- No automatic edits to provider config outside clearly owned managed blocks.
- No migration of existing city materialization behavior in this proposal.

## Implementation Sketch / Likely Phases

### Phase 0: Pack Resolution and Planning

- Add a `gc pack install --dry-run <pack-ref>` path that resolves a pack,
  reads `pack.toml`, discovers exported `skills/` and `mcp/`, and prints a
  no-write plan.
- Reuse existing pack cache/import resolution where practical, but do not
  require a root city or `packs.lock`.
- Define the install record schema before writing provider files.

### Phase 1: Skills Install / Uninstall

- Implement built-in provider target mappings for the first verified skill
  sinks.
- Install only top-level exported skills by default.
- Copy or symlink with a documented policy. Copying is safer across cache
  pruning; symlinking is better for local development. The first release may
  choose copy for remote/cached packs and symlink only for explicit local-dev
  mode.
- Record ownership and support uninstall of managed skills.

### Phase 2: MCP Projection

- Add provider-specific MCP projectors using the same managed-block discipline
  as runtime MCP materialization.
- Fail before writes if a target/scope lacks verified provider behavior.
- Preserve hand-authored config outside the managed block and snapshot/adopt
  only if the projection design requires it.
- Include MCP entries in install records and uninstall.

### Phase 3: Upgrade and List

- Add `gc pack list --installed` and `gc pack upgrade` once install records
  exist.
- Detect drift between the install record, provider files, and the currently
  resolved pack revision.
- Keep upgrade as a plan-first operation.

### Phase 4: Later Surfaces

- Revisit commands once command exposure and `gc` intermediation semantics are
  settled.
- Revisit agent profiles/templates only for providers with an explicit profile
  selection mechanism.
- Consider doctor checks as `gc`-scoped tooling, not provider-installed assets.

## Open Questions

- Should the public install surface be implicit by convention
  (`skills/`/`mcp/`) only, or should packs gain an explicit `[install]`
  manifest immediately?
- For local path packs, should the default skill install mode be copy or
  symlink?
- Where exactly should install records live, and how should they interact with
  existing `~/.gc` state?
- Which provider/scope pairs are verified enough for v1 skills and MCP?
- Should `gc pack install <pack-ref>` prompt interactively by default, or
  should no-target installs be dry-run-only until `--target` is specified?
- What is the minimum marker format needed in provider config files to make
  uninstall reliable without making the files feel owned by Gas City?
- How should re-exported imported assets be represented in `pack.toml` without
  overfitting to skills and MCP?
- Should project-scope installs require the target project to be a git repo, or
  is any explicit `--project-dir` sufficient?
