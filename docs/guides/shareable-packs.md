---
title: "Shareable Packs"
description: Create, import, and customize PackV2 Gas City packs.
---

A pack is a portable definition of behavior: agents, prompt templates,
providers, formulas, orders, commands, doctor checks, overlays, skills, and
other reusable assets. A city is a root pack plus `city.toml` deployment
configuration and machine-local `.gc/` bindings.

PackV2 keeps those concerns separate:

- `pack.toml` names the pack and declares its pack dependencies.
- Pack directories such as `agents/`, `formulas/`, and `commands/` hold
  reusable definitions and support files.
- `city.toml` says how this city runs those definitions.
- `.gc/` stores local site bindings and runtime state managed by `gc`.

Use this guide when you want to create a reusable pack, import one into a city,
or customize an imported pack without editing the pack's files.

## Pack Layout

Pack structure is convention-based. Standard directories are loaded by name;
opaque helper files belong under `assets/`.

```text
code-review-pack/
├── pack.toml
├── agents/
│   └── reviewer/
│       ├── agent.toml
│       └── prompt.template.md
├── formulas/
│   └── review-change.toml
├── orders/
│   └── nightly-review.toml
├── commands/
│   └── status/
│       ├── help.md
│       └── run.sh
├── doctor/
│   └── check-review-tools/
│       └── run.sh
├── overlay/
├── skills/
└── assets/
    └── scripts/
        └── setup-reviewer.sh
```

The smallest useful `pack.toml` names the pack and records the schema:

```toml
[pack]
name = "code-review"
schema = 1
version = "1.0.0"
```

Agent definitions live in `agents/<name>/agent.toml`. The directory name is
the agent's local name:

```toml
# agents/reviewer/agent.toml
scope = "rig"
provider = "codex"
prompt_template = "prompt.template.md"
nudge = "Review the assigned change and leave findings."
idle_timeout = "30m"
pre_start = ["{{.ConfigDir}}/assets/scripts/setup-reviewer.sh {{.RigRoot}}"]
```

Use `scope = "city"` for agents that should load at the city level. Use
`scope = "rig"` for agents that should be stamped into rigs that import the
pack. If `scope` is omitted, the agent is eligible in both contexts.

## Import A Pack

An import is a named dependency. Durable imports use `source` plus an optional
`version` constraint.

```toml
[imports.review]
source = "../packs/code-review"
```

For a versioned source:

```toml
[imports.gastown]
source = "https://github.com/gastownhall/gascity-packs.git//gastown"
version = "^1"
```

The table name, such as `review` or `gastown`, is the local import binding. It
helps Gas City load dependencies deterministically, but it does not become part
of runtime agent names.

## Find Packs With A Registry

> **Review note:** The registry discovery commands in this section depend on
> #2351. Confirm the final command wording with Donna or Mabel before merging
> this guide.

A pack registry is a catalog. It tells `gc` which packs exist, what versions
are available, and where their durable sources live.

Configure a registry on the local machine:

```text
$ gc pack registry add main https://packages.example/registry.toml
$ gc pack registry refresh main
```

Search the cached registry records:

```text
$ gc pack registry search "code review"
```

Inspect one result:

```text
$ gc pack registry show main:gascity
```

Registry handles such as `main:gascity` are command-time lookup handles. Do
not persist them in `pack.toml`. Use the durable `source` and optional
`version` from the registry record when you write the import.

## Import Into A City

A city can import a pack at the root pack level:

```toml
# pack.toml
[pack]
name = "bright-lights"
schema = 1

[imports.review]
source = "../packs/code-review"
```

City-level imports make city-scoped agents and definitions available to the
city. If the imported pack defines a city-scoped `reviewer` agent, the runtime
agent name is `reviewer`.

Use rig-level imports when only one rig should receive a pack's rig-scoped
agents or formulas:

```toml
# city.toml
[[rigs]]
name = "backend"

[rigs.imports.review]
source = "../packs/code-review"
```

If the imported pack defines a rig-scoped `reviewer` agent, the runtime name is
`backend/reviewer`. The `backend` part comes from the rig `name`, not from the
import binding and not from the rig filesystem path.

Validate after changing imports:

```text
$ gc config show --validate
```

## Customize An Imported Pack

A reusable pack should work without editing its files. Customize from the city
that imports it.

Use city defaults for broad local policy. Defaults fill blank fields only:

```toml
# city.toml
[agent_defaults]
default_sling_formula = "review-change"
```

Use a patch when you need to change one city-level agent:

```toml
# city.toml
[[patches.agent]]
name = "reviewer"
provider = "codex"
session_setup_append = ["tmux set status-left '[review]'"]
```

For a rig-scoped agent, target the rig identity prefix with `dir`:

```toml
# city.toml
[[patches.agent]]
dir = "backend"
name = "reviewer"
provider = "gemini"
```

The `dir` value is the rig name. It is not the rig path.

Rig overrides are another rig-local customization point:

```toml
[[rigs]]
name = "backend"

[[rigs.overrides]]
agent = "reviewer"
provider = "gemini"
```

Prefer patches and rig overrides over forking. Fork only when you want to own a
new pack with a different public surface.

## Name Public Surface Carefully

Names are part of a pack's public API. A downstream city can route work to an
agent name, invoke a formula name, or call a command name.

Choose stable names for:

- agents under `agents/<name>/`
- formulas under `formulas/<name>.toml`
- orders under `orders/<name>.toml`
- commands under `commands/<name>/`

If two imported packs define the same city-level agent name, config load fails.
The same applies inside one rig. Rename one public definition or avoid
importing both packs onto the same surface.

## Update Or Remove A Pack

The checked-in import records the compatibility range. The lock and cache
record the exact resolved copy used locally.

After editing imports, run:

```text
$ gc config show --validate
```

If the pack came from a remote source, keep the lock and cache in sync with the
current import-management commands for your release. Registry discovery in
#2351 does not make registry handles durable pack coordinates; the committed
import remains `source` plus optional `version`.

