# Research: per-role / per-task model selection, and an open-model (GLM/opencode/pi/oh-my-pi) pathway

Bead: ga-7w4. Research only, no code changes.

Repo root: `/Users/jeff.burn/Code/gasland` (not a git repo itself).
Main SDK: `/Users/jeff.burn/Code/gasland/rigs/gascity`, at commit `21ffdae5139bf9aa8049f52c0aa228cb40d7a9ef` (`fix(dispatch): tolerate torn-down worktree in ralph check gate (ga-myq)`) when this research was done, 2026-07-16.
Packs sibling tree: `/Users/jeff.burn/Code/gasland/rigs/gascity-packs`, at `21ffdae51` on the same clone (fork `origin`, upstream `gastownhall/gascity-packs`).
Live deployed city config: `/Users/jeff.burn/Code/gasland/city.toml` (this is the actual fleet config referenced in the prompt, not an example).

---

## A. Current state — how is a model assigned today?

### A.1 The field is `option_defaults["model"]`, not a dedicated `Model` field on `Agent`

There is **no** `Agent.Model` field. The hint in the task ("adding a config.Agent field touches AgentPatch/AgentOverride/applyAgentMutation/Agent.Clone") describes the general mechanism for *any* new agent-level field, but model selection already rides an **existing generic mechanism**: `Agent.OptionDefaults map[string]string` (`rigs/gascity/internal/config/config.go:2999`), keyed by the provider's `OptionsSchema` keys (`model`, `effort`, `permission_mode`, …).

Plumbing, source → runtime:

1. **Provider schema (builtin default)** — `rigs/gascity/internal/worker/builtin/profiles.go:164-172` (`claude` provider's `model` `OptionsSchema` entry). Each choice carries `FlagArgs` — the literal CLI args to inject, e.g.:
   ```go
   {Value: "sonnet", Label: "Sonnet", FlagArgs: []string{"--model", "claude-sonnet-5"}, FlagAliases: [][]string{{"-m", "claude-sonnet-5"}}},
   ```
   (line 169). `opus` → `claude-opus-4-8` (line 167). This is the "builtin claude model choices" file named in the task hint.
   - **ga-pai has already landed.** `bd show ga-pai` shows it as in-progress, but the described change (add `claude-sonnet-5` to the closed enum, mirroring the fable-5 precedent) is present in the checked-out source: commit `88edd472a` "feat(providers): add Claude Sonnet 5 to builtin claude model choices (#3867)". The bead's own notes record this as merged (fork commit + upstream PR #3867) with only fleet redeploy/rollout left as adjunct-owned follow-up. `git log --oneline -- internal/worker/builtin/profiles.go` (top 3): `e8f071037` (adds gpt-5.6 codex choices), `88edd472a` (sonnet-5, #3867), `1e5ad20b8`.
   - Model is **not** Anthropic-specific in the type system: `BuiltinOptionChoice.FlagArgs []string` (`internal/config/provider.go:37-40`) is an opaque arg-list injected into whatever `Command` the provider spec names. The same struct is used verbatim for the `codex`, `gemini`, `grok`, `kimi`, `kiro`, `cursor`, `copilot`, `amp`, `opencode`, `mimocode`, `cerebras`, `groq`, `auggie`, `pi`, `omp`, `antigravity` builtin provider specs (`profiles.go:103-700`, provider name list via `grep -n '^\t"[a-z-]*": {'`).

2. **Config load → resolve** (`rigs/gascity/internal/config/resolve.go`):
   - `ComputeEffectiveDefaults(schema, providerDefaults, agentDefaults)` (`internal/config/options.go:38-`) merges three layers in order: schema-declared default → provider-level `option_defaults` → agent-level `option_defaults` (agent wins). This is literally the "Layer 1/2/3" doc comment at `options.go:38-45`.
   - `resolve.go:756-806` `mergeAgentOverrides(rp *ResolvedProvider, agent *Agent)`: for `OptionDefaults`, agent keys are merged on top of `rp.EffectiveDefaults` (`resolve.go:799-806`).
   - `ResolvedProvider.EffectiveArgs()`-style logic (`provider.go:388-400`, `FlagArgs` emission) walks the effective default per option key and appends `choice.FlagArgs` to the final CLI argv.

3. **AgentPatch / AgentOverride plumbing** (the map the task hint pointed at) — confirmed but *not* model-specific:
   - `AgentPatch` (`internal/config/patch.go:23`), `AgentOverride` (`internal/config/config.go:653`), `applyAgentMutation` (`internal/config/patch.go:445`) all carry `OptionDefaults map[string]string` (`patch.go:159`, `config.go:785`) with additive-merge semantics (`patch.go:597-604`, `pack.go:2721-2734` via `toAgentPatch()`). `Agent.Clone()` (`config.go:3259`, copy at `config.go:3277`) deep-copies `OptionDefaults`. So model selection already flows through this exact map end-to-end without any dedicated field — `option_defaults = { model = "opus" }` on an `[[patches.agent]]` entry is today's supported surface for a per-agent model override (see A.3).

4. **Runtime dispatch** (`internal/runtime/` providers: `tmux`, `exec`, `acp`, `subprocess`, `herdr`, `k8s`, `ssh`, `hybrid`, `auto`): these are transport providers — they start/observe/nudge a session — and are model-agnostic. They receive the resolved `Args`/`FlagArgs` (already containing e.g. `--model claude-sonnet-5`) as an opaque argv/handshake payload; the exec provider (`internal/runtime/exec/exec.go:118,192`) execs `cfg.Command` with args built upstream in `config`; the ACP provider (`internal/runtime/acp/acp.go`) speaks JSON-RPC over stdio per the Agent Client Protocol (`internal/runtime/protocol.go` documents the "RPP" contract: version handshake, capability strings, one-op-per-invocation for exec). Nothing in `internal/runtime/*` special-cases "claude" or Anthropic model names — model plumbing is entirely a `internal/config` concern (schema → effective defaults → FlagArgs), and the runtime layer just launches `Command` with `Args`.

### A.2 Where the model choice is validated / rejected

`ValidateOptionDefaults` (`internal/config/options.go:23-35`) rejects any `option_defaults` value that isn't a declared `Choices[].Value` for that schema key — this is the "closed enum" the ga-pai bead investigated (adding `claude-sonnet-5` required a source change to `profiles.go`, not a config-only workaround).

### A.3 How "mayor=opus, others=sonnet" is expressed **today**, verbatim from the live `city.toml`

```toml
# city.toml (repo root)
[providers.claude]
base = "builtin:claude"
options_schema_merge = "by_key"
ready_delay_ms = 0
[providers.claude.option_defaults]
effort = "medium"
model = "sonnet"        # <- city-wide default for every agent on the claude provider
permission_mode = "auto"
...
[[patches.agent]]
dir = ""
name = "mayor"
[patches.agent.option_defaults]
effort = "high"
model = "opus"          # <- single agent-level override, additive on top of the provider default
```
(`/Users/jeff.burn/Code/gasland/city.toml:13-19` and `:53-59`.)

No "adjunct" agent exists in this city's config, `gascity-packs/gascity/roles/agents/*`, or the example cities under `rigs/gascity/examples/*/city.toml`; "adjunct" in the file is a person/role referenced only in a comment (`city.toml:5`) and is a pack under `rigs/gascity-packs/adjunct/`, not a named agent with its own model override. So today's *actual* split is exactly: one city-wide provider default (`sonnet`) plus one `[[patches.agent]]` override (`mayor` → `opus`) — the coarse split described in the prompt is not a special-cased code path, it's the smallest possible use of the general per-agent `option_defaults` override.

None of the `gascity-packs/gascity/roles/agents/*/agent.toml` files (design-author, publisher, issue-triager, design-implementation-reviewer, task-decomposer, run-operator, implementation-reviewer, implementation-worker, gap-analyst, design-test-risk-reviewer, requirements-planner, review-synthesizer) set `option_defaults` or `provider` at all — role model choice is 100% left to city-level config today.

---

## B. Per-role model levels — what exists vs. needed

**It already works, today, with existing config surface — nothing to build.** The mechanism used for `mayor` (`[[patches.agent]] name = "<role-or-agent-name>"` + `[patches.agent.option_defaults] model = "<choice>"`) generalizes to any agent/role name with zero code changes, because:
- `applyAgentPatch` (`internal/config/patch.go:408`) matches by `(Dir, Name)`, so a per-role split is one `[[patches.agent]]` block per role name.
- `option_defaults` is validated at config-load time against the provider's schema (`ValidateOptionDefaults`), so a typo'd model tier fails closed, not silently.

**Minimal config change to express "reviews/plans/requirements/design → opus, simple implementation → sonnet" fleet-wide** (recommended, no Go changes):

```toml
[[patches.agent]]
name = "requirements-planner"
[patches.agent.option_defaults]
model = "opus"

[[patches.agent]]
name = "design-author"
[patches.agent.option_defaults]
model = "opus"

[[patches.agent]]
name = "implementation-reviewer"
[patches.agent.option_defaults]
model = "opus"

# implementation-worker inherits the city-wide provider default (sonnet) — no patch needed.
```

This is literally more `[[patches.agent]]` blocks of the same shape already in production for `mayor`. If the same tiering should apply identically across every city that imports the `gc-roles` pack (`rigs/gascity-packs/gascity/roles/pack.toml`), the cleaner fleet-wide surface is to set `option_defaults.model` directly on each role's `agent.toml` in the pack (e.g. `rigs/gascity-packs/gascity/roles/agents/requirements-planner/agent.toml`) rather than repeating `[[patches.agent]]` per city — `Agent.OptionDefaults` is a first-class TOML field on the pack-authored `Agent` struct (`config.go:2999`) merged the same way (`applyOverrides`/`applyPackAgentPatches`, `pack.go:2337,2695`). That is a one-line-per-role pack change, still zero Go changes — purely config/data.

---

## C. Per-task/per-bead model selection

### C.1 A per-dispatch mechanism already exists and is live in `gc`

Bead metadata key **`opt_model`** (convention: `opt_<OptionsSchema key>`, prefix declared as `beadmeta.OptionMetadataPrefix = "opt_"`, `internal/beadmeta/keys.go:266-273`) is read at reconcile/dispatch time:

- `resolveTaskOptionOverrides(store, rp, assignees...)` (`cmd/gc/session_reconciler.go:5073-5099`) looks at the newest `in_progress` work bead(s) assigned to the candidate session and calls `workBeadOptionOverrides` (`session_reconciler.go:5104-5128`), which reads `b.Metadata["opt_model"]`, `b.Metadata["opt_effort"]`, etc., validates each against `config.ResolveExplicitOptions(rp.OptionsSchema, ...)`, and drops (with a log line) anything invalid.
- The result is consumed in `cmd/gc/session_lifecycle_parallel.go:867-880`: "Work beads may carry one-shot provider option overrides as `opt_<key>` metadata... Apply them after core/live hash calculation because they are dispatch inputs from the current work bead, not durable session config." These override the durable per-agent `option_defaults` for that one launch, and explicit session `template_overrides` still win over them per key.
- Landed in commit `4d36f6cb6` / `27b42186f`, **"Route work option metadata through OptionsSchema" (#3239), 2026-06-09** — i.e. *after* the coarse mayor/sonnet split was set up, and importantly *before* PR #58 was even opened (see D).

- **`gc.model` is a deprecated spelling** of the same idea and is explicitly documented as such: `docs/reference/specs/formula-spec-v2.md:583` — "Per-dispatch provider options ride `opt_*` step metadata (for example `opt_model`), validated against the provider's options schema at spawn; `gc.model` is a deprecated spelling that the `gc doctor` check `work-option-metadata-migration` migrates to `opt_model`." (also the deprecation table at line 1038). `internal/beadmeta/keys.go:125,141` still declares `ModelMetadataKey = "gc.model"` and `PerDispatchModelMetadataKey = "gc.per_dispatch_model"` as constants, but neither is *read* anywhere outside `internal/beadmeta/keys.go` itself and `cmd/gc/doctor_work_option_metadata.go` / its tests — they exist purely as migration-source spellings for `gc doctor`'s `work-option-metadata-migration` check (`cmd/gc/doctor_work_option_metadata.go:204`, `doctor_work_option_metadata_test.go`), not as a live dispatch input.

**So, to directly answer the question:** yes — a formula step (or anything that stamps `bd update <id> --set-metadata opt_model=opus`) already gets its model choice threaded to session spawn, validated against the resolved provider's schema, with no Go change required. This is functionally the workflow-tool `opts.model`/`opts.effort`-per-call analog the task asked about, at the dispatch layer rather than inside a single agent-loop call.

### C.2 Static per-formula/per-step model (option a) — supported today

A formula step can set `opt_model` directly as step metadata (`formula-spec-v2.md:582-584`), so "this formula step always runs on opus" is a config-only static choice, already wired end-to-end via C.1.

### C.3 Complexity-driven per-task selection (option b) — the judgment call, and where it must live

`AGENTS.md:304-308` states the principle directly:

> **Keep judgment out of Go.** Go handles transport, not reasoning. The framework moves work; it doesn't reason about it. If a line of Go contains a judgment call, it's a violation. **The test:** does any line of Go contain a judgment call? An `if stuck then restart` is framework intelligence. Move the decision to the prompt.

and the related Primitive Test at `AGENTS.md:295-297` (atomicity + gets more useful as models improve + keeps judgment out of Go).

Given C.1's existing seam, "simple impl → sonnet, complex impl → opus" decomposes cleanly without violating that principle:
- **Classification is a prompt/agent decision**, not a Go heuristic: a triage/dispatcher agent (or the mayor, or a dedicated advisor role) reads the bead/task, judges complexity, and calls `bd update <id> --set-metadata opt_model=<tier>` (or authors it into the formula step at plan time). No Go code inspects task content or applies an if/else on "complexity."
- **Go's only job stays what it already does**: read `opt_model` off the assigned work bead, validate against the schema, apply as a FlagArgs override at spawn (`resolveTaskOptionOverrides`/`workBeadOptionOverrides`). This is transport, not reasoning, and is already built.
- Sling-time human/mayor choice (option c) is also already possible today by hand via `bd update <id> --set-metadata opt_model=opus` before/at sling — no new surface needed, just workflow convention (e.g. document it in the mayor's prompt template, or add a `sling --model` UX convenience *in front of* the same metadata write — that would be the only net-new Go work, and it's a thin arg-to-metadata passthrough, not a judgment call).

**Conclusion for C:** the missing piece is not a Go mechanism — it's (1) a prompt/skill that teaches the mayor/triage agent the convention "stamp `opt_model` based on your judgment of task complexity," and (2) optionally, an evidence-driven advisor that recommends (never silently applies) via the same key. That is what PR #58 targets — see D below, with an important compatibility gap.

---

## D. Community pieces

### D.1 `gastownhall/gascity-packs#58` — "Add model-advisor pack — Bayesian model-tier selection for Gas City"

- `gh pr view 58 --repo gastownhall/gascity-packs`: **state OPEN**, author `jsgerman-oss` (Jay German), +17,329/−0, label `kind/feature`, created 2026-06-04, last updated 2026-07-14. `gh pr view 58 --repo gastownhall/gascity-packs --json mergeable,mergeStateStatus`: `"mergeable":"CONFLICTING"`, `"mergeStateStatus":"DIRTY"` — **not mergeable as-is** (conflicts with `main`, needs a rebase regardless of any other assessment).
- What it adds (from `gh pr diff 58`): a Python, stdlib-only pack `model-advisor/` implementing "Conservative Constrained Thompson Sampling" (CC-TS) — a per-`(provider, agent, shape, tier)` Beta-Bernoulli posterior learned from bead-closure + a `Stop`/`SubagentStop` telemetry hook, gated by a one-sided Wilson lower-confidence-bound vs. a baseline tier, choosing the cheapest tier that clears the gate. Surfaces: `advisor advise <agent> <shape>` (read-only recommendation), `advisor inspect` (evidence), `advisor apply <agent>` (writes the tier into the agent's config), `advisor auto-apply` (scheduled sweep). 71 tests claimed in the summary / 228 claimed in the README for the "advanced modes." Install is per-town or per-rig via `install.sh`, editing `city.toml`/`pack.toml` imports and merging a Stop-hook overlay.
- **How it wires model selection, and the key finding**: the PR's own `model-advisor/docs/INTEGRATION-FEASIBILITY.md` (bundled in the diff, `gh pr diff 58` around line 1653+) states its own limitation plainly: *"Per-AGENT: YES, today... Per-DISPATCH/per-task: NO — needs a gc-core change (no `--model` on `gc sling`, no `gc.model` consumption at spawn)"* and describes a "missing seam" of binding bead `gc.model` → an env var `GC_AGENT_MODEL` at launch. **This is now stale**: the actual gc-core seam that shipped is `opt_model` bead metadata consumed via `resolveTaskOptionOverrides`/`workBeadOptionOverrides` and applied as provider-schema `FlagArgs` (not an env var, not the `gc.model` key) — landed in commit `4d36f6cb6`/PR #3239 on **2026-06-09**, five days after this PR's repo was created (2026-06-04) and evidently not yet reflected in the PR's design docs, which still reference `gc.model` throughout (`grep` inside the diff shows ~15 occurrences of `gc.model`, 0 of `opt_model`). v1's own `apply` surface is also described in its README as "coarse: one tier per agent... Per-dispatch application... is the next step" — i.e. the PR author believes per-dispatch is future work; it is in fact already possible today via `opt_model`, just not what this pack targets or writes.
- **Fit for a fork at `jeffora/gascity-packs`** (the fork this city's `rigs.imports.gc` actually points at — `city.toml:37,44`: `source = "https://github.com/jeffora/gascity-packs/tree/main/gascity/roles"`): architecturally compatible (it's a self-contained pack, install/uninstall reversible, stdlib-only, no gc-core changes required to *land* it), but as authored it optimizes the wrong metadata key for per-dispatch application and needs non-trivial adaptation (swap `gc.model`/`gc.provider`/`gc.run_target` writes for `opt_model`/`opt_provider`-shaped keys validated against the resolved provider's `OptionsSchema`) before its "apply" surface can do anything beyond the coarse per-agent `option_defaults` edit it already supports.
- **Recommendation: adapt, don't merge as-is.** The statistical engine (CC-TS gate, Beta posteriors, telemetry hook, `advise`/`inspect` reads) is sound, self-contained, and orthogonal to the metadata-key question — it's a legitimate answer to C.3's "evidence-driven advisor" option. But (1) the PR needs a rebase (currently CONFLICTING/DIRTY), (2) its dispatch-time write path needs to target `opt_model` (the real, already-shipped seam) instead of `gc.model`/`GC_AGENT_MODEL` (a seam that doesn't exist and was never built that way), and (3) given it's a from-scratch clean-room Python implementation with an installer that edits `city.toml`/`pack.toml` and merges hook overlays, it deserves its own security/quality review pass before adoption — not evaluated here beyond structure/fit.

### D.2 `jsgerman-oss/model-advisor` (the upstream repo the pack PR mirrors)

`gh repo view jsgerman-oss/model-advisor`: public, MIT-licensed Python repo, description "Bayesian model-tier selection for Gas City agents", created 2026-06-04, 1 star, 1 fork, 2 open issues. Its README (fetched via `gh pr diff`, since the PR vendors the same content) differs from the pack PR only in provenance framing ("An implementation of..." vs. "A clean-room implementation of..." of the same "Conservative Constrained Thompson Sampling" paper at `jsgerman-oss/research`). It is the same artifact as D.1's pack contents; no separate integration surface. Same adopt/adapt call applies.

---

## E. Open-model pathway (GLM 5.2 via opencode / pi / oh-my-pi)

### E.1 The runtime is already provider-agnostic, and opencode/pi/omp are already builtin provider *presets* in gascity source

`internal/worker/builtin/profiles.go` already declares, as builtin `BuiltinProviderSpec` entries (not hypothetical — present in the checked-out source at commit `21ffdae51`):

- `"opencode"` (`profiles.go:513-540`): `Command: "opencode"`, `PromptMode: "flag"`/`--prompt`, `SupportsACP: true`, `ACPArgs: []string{"acp"}`, `InstructionsFile: "AGENTS.md"`, a `model` `OptionsSchema` with choices already routed through OpenCode's own model-string convention (`opencode/deepseek-v4-flash-free`, etc.).
- `"cerebras"` (`profiles.go:581-608`) and `"groq"` (`profiles.go:610-640`): both are **OpenCode with a different `OptionDefaults.model`/API backend**, same `Command: "opencode"` — i.e. gascity already treats "opencode pointed at a different model-serving backend" as just another builtin provider preset. Notably `cerebras`'s model choices already include `{Value: "cerebras/zai-glm-4.7", Label: "GLM 4.7", FlagArgs: []string{"--model", "cerebras/zai-glm-4.7"}}` (`profiles.go:604`) — **a GLM model is already a first-class model choice today**, just not GLM 5.2 and not via a GLM-native/Z.ai endpoint.
- `"pi"` (`profiles.go:662-684`): `Command: "pi"`, `Args: ["-e", ".pi/extensions/gc-hooks.js"]`, `SupportsHooks: true`, a `model` schema entry already wired to a non-Anthropic backend (`{Value: "ollama-cloud-gpt-oss-20b", ... FlagArgs: []string{"--provider", "ollama-cloud", "--model", "gpt-oss:20b"}}`).
- `"omp"` (`profiles.go:685-695`, "Oh My Pi (OMP)"): `Command: "omp"`, `Args: ["--hook", ".omp/hooks/gc-hook.ts"]`, `SupportsHooks: true`, `InstructionsFile: "AGENTS.md"`, `ResumeFlag: "--resume"` — present as a builtin provider preset but with **no `model` `OptionsSchema` populated yet** (unlike `pi`/`opencode`/`cerebras`).

None of these five providers were added by the in-flight ga-pai bead; `git log --oneline --all -S'"opencode": {'` / `-S'"pi": {'` / `-S'"omp": {'` on `profiles.go` all bottom out at refactor commits `1e4c7fe43`/`3cf3a24b3` ("refactor: move worker authority toward phase 4 boundary"), i.e. they predate this research and are already load-bearing, tested builtin specs.

The runtime contract these providers satisfy is the generic one at `internal/runtime/protocol.go` (RPP: version handshake + capability strings like `report-attachment`, `proc.exec`, `proc.stream`, `tty.attach`) plus, for ACP-capable providers, the JSON-RPC-over-stdio Agent Client Protocol spoken by `internal/runtime/acp/acp.go`. `SupportsACP *bool` (`internal/config/provider.go:98-101`) is a per-provider tri-state flag gating whether `Agent.Session = "acp"` is legal (`config.go:2961`); `claude`, `opencode`, and `mimocode` builtin specs set it true, `pi`/`omp`/`auggie` do not (they run over the plain CLI/exec transport instead, per their spec entries above). Neither transport is Anthropic-specific.

### E.2 External verification of opencode / pi / oh-my-pi (websearch, not gascity source)

- **OpenCode** ships a native `opencode acp` subcommand that starts an ACP-compliant JSON-RPC-over-stdio server, documented at opencode.ai and openclaw docs, and is listed by third parties (e.g. `formulahendry/vscode-acp`) as one of several ACP-compatible agents alongside Claude/Codex/Copilot/Qwen/Gemini/Kiro. *(Source: [ACP | OpenCode](https://opencode.ai/docs/acp/), [ACP - OpenCode Docs](https://open-code.ai/en/docs/acp), [vscode-acp README](https://github.com/formulahendry/vscode-acp).)*
- **Oh My Pi / `omp`** (`can1357/oh-my-pi` on GitHub, npm package `@oh-my-pi/pi-coding-agent`) is described as a fork of Mario Zechner's `pi-mono`/pi-coding-agent, extended with "40+ providers, 32 built-in tools, 14 LSP operations... ~55k lines of Rust core" and IDE integration; I could **not verify from search results whether `omp` itself speaks ACP** (unlike opencode, which explicitly documents an `acp` subcommand) — gascity's own builtin spec for `omp` (`profiles.go:685-695`) does not set `SupportsACP` or `ACPArgs`, consistent with this being unconfirmed/plain-CLI-only; flag this as **unverified, do not assume ACP support for omp** without checking its own docs/`--help`. *(Source: [can1357/oh-my-pi](https://github.com/can1357/oh-my-pi), [oh-my-pi npm](https://www.npmjs.com/package/@oh-my-pi/pi-coding-agent).)*
- **`pi` / pi-coding-agent** (`badlogic/pi-mono`, npm `@mariozechner/pi-coding-agent`, by Mario Zechner) is documented as supporting "a wide range of LLM providers—including Anthropic, OpenAI, Google Gemini, Groq, and others—via API keys or subscriptions." *(Source: [mariozechner.at post](https://mariozechner.at/posts/2025-11-30-pi-coding-agent/), [pi-coding-agent npm](https://www.npmjs.com/package/@mariozechner/pi-coding-agent).)* I could not confirm from search results whether `pi` supports a GLM/Z.ai backend specifically, or only the providers named in that description — **treat "can pi run GLM 5.2" as unverified** pending checking `pi`'s own provider list/config docs.
- **GLM-5.2** (Zhipu AI / Z.ai) is a real, current (as of this research) open-weight release: MoE, ~744B total / ~40B active params, 1M-token context, MIT license, weights on Hugging Face (`zai-org/GLM-5.2`) and ModelScope, API priced ~$1.40/$4.40 per Mtok in/out, first available 2026-06-13. *(Sources: [datanorth.ai](https://datanorth.ai/news/zhipu-ai-releases-glm-5-2), [eigent.ai](https://www.eigent.ai/blog/glm-5-2), [letsdatascience.com](https://letsdatascience.com/news/zai-releases-glm-52-tops-open-weight-rankings-7c27fe58).)* Since it's MIT-licensed and self-hostable/API-servable, it is a plausible drop-in `model` choice for any of gascity's already-existing OpenCode-backed providers (which route to arbitrary `provider/model` strings, as the `cerebras`/`groq`/plain-`opencode` specs already demonstrate) — but the exact model-ID string OpenCode/Cerebras/Groq/Z.ai would expose for GLM 5.2 was not verified in this pass and must be checked at implementation time (OpenCode's `cerebras/zai-glm-4.7` naming convention in `profiles.go:604` suggests the eventual string would look like `<backend>/zai-glm-5.2` or similar, but do not assume this without checking the serving backend's actual model catalog).

### E.3 Smallest viable open-model integration path

Given E.1, the runtime/config side needs **no new provider, no new runtime, and no Go changes** for the opencode route — the smallest viable path is a **config-only addition mirroring the existing `cerebras`/`groq` presets**:

1. Add a new builtin (or city/pack-level custom) provider spec, e.g. `"zai"` or extend `"opencode"`'s existing `model` schema, with a new `BuiltinOptionChoice` such as `{Value: "zai/glm-5.2", Label: "GLM 5.2", FlagArgs: []string{"--model", "zai/glm-5.2"}}` — same pattern as the existing `cerebras/zai-glm-4.7` entry (`profiles.go:604`). This can be done **without touching Go** by defining a custom provider in `city.toml`/pack `pack.toml` with `base = "builtin:opencode"` and a `[providers.<name>.option_defaults]`/`[[providers.<name>.options_schema]]` override — the same `options_schema_merge = "by_key"` mechanism already used for the `claude` provider's `permission_mode` custom choice in the live `city.toml:20-30`.
2. Set an agent's (or a new agent's) `provider = "<that-provider-name>"` and, if needed, `option_defaults = { model = "zai/glm-5.2" }` — the same `Agent.Provider`/`Agent.OptionDefaults` fields already covered in A/B.
3. Credentials/endpoint: OpenCode-backed providers pass model-serving credentials via OpenCode's own env/config, not gascity's `UpstreamBaseURLEnv`/`UpstreamAPIKeyEnv`/`UpstreamAuthTokenEnv` fields (those three are declared only on the `claude` builtin spec, `profiles.go:106-108`, for `ANTHROPIC_BASE_URL`/`ANTHROPIC_API_KEY`/`ANTHROPIC_AUTH_TOKEN`) — an OpenCode-fronted GLM agent would need its Z.ai/GLM API key supplied the way OpenCode itself expects (e.g. via its own auth config or `Env` map on the `Agent`/provider spec, which is a generic `map[string]string` already supported, `config.go` `Env` field).
4. For `pi`/`omp`, the same shape applies if/once their own provider-side model routing to GLM is confirmed (E.2 flags this unverified) — worst case, wrap GLM behind an OpenCode-fronted provider (proven route) rather than `pi`/`omp` directly.

**Gaps / what would break for a non-Anthropic model:**
- `ValidateOptionDefaults`/closed-enum validation (A.2) means **any** new model string — Anthropic or not — must be a declared `Choices[].Value` in the resolved provider's schema first; this is a generic safety rail, not an Anthropic-specific assumption, but it does mean GLM 5.2 can't be used by just typing a model string into `option_defaults` — the schema entry must exist first (config-only, per step 1 above).
- `TitleModel` (`provider.go:366-380`, e.g. `TitleModel: "haiku"` on the `claude` spec, `"cerebras/gpt-oss-120b"` on the `cerebras` spec) is a **per-provider** field already, so it is not an Anthropic-only assumption either — a new GLM-backed provider would just set its own `TitleModel`.
- Anthropic-specific env plumbing (`UpstreamBaseURLEnv: "ANTHROPIC_BASE_URL"` etc., only on the `claude` spec) does not apply to OpenCode-fronted providers and isn't inherited by them (each spec declares its own), so no leakage risk there — but it does mean **no existing gascity mechanism auto-provisions GLM/Z.ai credentials**; that's new config (an `Env` map entry or an OpenCode-native auth file staged via `PreStart`), not a gap in the runtime contract.
- Token/caching assumptions: not investigated in this pass — Claude Code's own prompt-caching behavior is Anthropic/Claude-CLI-specific, but gascity's runtime layer doesn't model or depend on caching semantics anywhere found in `internal/runtime` or `internal/config`; this is a soft "unverified, likely fine" rather than a known break.

---

## Recommended pathway (summary)

1. **Per-role tiers**: no code change. Add `[[patches.agent]]` blocks per role name (or set `option_defaults.model` directly on each `gascity-packs/gascity/roles/agents/<role>/agent.toml`) mirroring the existing `mayor` → `opus` pattern at `city.toml:53-59`. This is today's real config surface, exercised in production for exactly one role; extending it to N roles is copy-paste, not new plumbing.
2. **Per-task selection, judgment out of Go**: use the already-shipped `opt_model` bead-metadata seam (`resolveTaskOptionOverrides`/`workBeadOptionOverrides`, PR #3239, 2026-06-09). Author formulas with static `opt_model` step metadata for known-complexity steps (C.2), and/or have the mayor/a triage agent stamp `opt_model` on a bead via `bd update <id> --set-metadata opt_model=<tier>` as a prompt-driven judgment call (C.3) — Go only ever reads and validates the metadata key against the provider's schema, never classifies complexity itself, satisfying `AGENTS.md`'s "keep judgment out of Go."
3. **PR #58 / model-advisor: adapt, don't merge as-is.** Rebase to resolve the current CONFLICTING/DIRTY state, and redirect its dispatch-time write path from the stale `gc.model`/`GC_AGENT_MODEL` design (based on a pre-#3239 understanding of gc-core) to the real `opt_model` seam. Its statistical core (CC-TS advisor, `advise`/`inspect` read-only surfaces) is a reasonable, non-invasive complement to (2) and doesn't require gc-core changes to install — but its "apply"/"auto-apply" write paths need the metadata-key fix before they do anything useful per-dispatch, and the pack as a whole (installer, hook overlay merge, Python engine) warrants its own review before adoption.
4. **Open-model pathway**: smallest viable path is **config-only** — gascity already ships `opencode`, `cerebras` (OpenCode+GLM-4.7 today), `groq`, `pi`, and `omp` as builtin provider presets (`profiles.go:513-695`), and the runtime/ACP/exec transport layer is already provider-agnostic. Bringing up GLM 5.2 means adding one new `model` choice (or a new provider spec cloned from `cerebras`'s pattern) pointing at whatever backend serves GLM 5.2 (Z.ai API, or an OpenCode-compatible aggregator), plus wiring its credential via the provider's `Env`/auth config — no new runtime provider, no ACP/exec code changes. Confirm before implementing: (a) the exact model-ID string the chosen backend uses for GLM 5.2, (b) whether `pi`/`omp` need GLM routed through them directly or can be skipped in favor of the already-proven `opencode`-fronted route, both flagged unverified in E.2.

## Parts not fully verified

- Whether `omp` (oh-my-pi) speaks ACP — its gascity builtin spec sets no `SupportsACP`/`ACPArgs`, and web search didn't confirm ACP support either way.
- Whether `pi` (badlogic/pi-mono) or `omp` can serve GLM 5.2 specifically (vs. only the providers explicitly named in their own docs) — not found in search results.
- The exact model-ID string a GLM-5.2-serving backend (Z.ai, Cerebras, Groq, or an OpenCode aggregator) would expose to `--model`; inferred by analogy to the existing `cerebras/zai-glm-4.7` entry, not confirmed against any live backend.
- Whether prompt-caching/token-accounting assumptions elsewhere in gascity (outside `internal/runtime` and `internal/config`, not fully swept) depend on Claude-specific behavior.
