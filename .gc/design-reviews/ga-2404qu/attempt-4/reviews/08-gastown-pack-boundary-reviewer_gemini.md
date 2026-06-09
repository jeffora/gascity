# Gastown Pack Boundary Review — Independent DeepSeek V4 Flash Analysis

**Verdict:** block

**Reviewer focus:** External Gastown ownership, Maintenance retirement, Core/Gastown split completeness, public pack contract.

---

## Top Strengths

1. **Robust Gating and Pinned Commits:** The public-pin rollout model (`internal/config/PublicGastownPackVersion` pinned to an immutable git commit) and the mandatory behavior-preservation manifest are exceptional. Forcing public pack validation *prior* to in-tree source deletion prevents the "behavior gap" anti-pattern.
2. **Explicit Required-Core Verification:** The design correctly makes Core a typed, content-backed, required system pack verified via `RequiredSystemPackParticipation`. This prevents silent degradation of the CLI and core infrastructure if a config resolves without Core.
3. **Structured Maintenance Retirement Table:** Mapping each runtime surface (registry, includes, materialization, orders, state) to target behaviors and required proof is a masterclass in migration planning.

---

## Critical Risks

### [Blocker] `gc doctor` and `config.Load` deadlocks on un-migrated cities with missing `.gc/system/packs/gastown` imports
When an existing city is opened under the new binary, `MaterializeBuiltinPacks` no longer materializes the legacy `gastown` or `maintenance` packs under `.gc/system/packs/{name}` since they have been removed from `All()`.
However, that city's `city.toml` or `pack.toml` still contains `[imports.gastown]` pointing to the local system-pack path. 
When any command or doctor check runs, the config system attempts to load the configuration and resolve the imports. Since `.gc/system/packs/gastown` is missing (especially in clean workspaces or fresh checkouts), `config.Load` will fail closed with a directory-not-found or file-not-found error.
Because `gc doctor` itself relies on loading the configuration to analyze the city and run preflight checks, the system enters a deadlock:
1. `gc start` / `gc doctor` fails to run because the config cannot load due to the missing legacy pack.
2. The user cannot run `gc doctor --fix` to rewrite the imports because the doctor itself fails during config load.
This bricks the city with no automated recovery path.

**Required change:**
The config loader must be updated to support a lenient, partial loading mode specifically during diagnostic or doctor runs. Alternatively, the doctor must parse and edit the raw TOML AST of `city.toml`/`pack.toml` to migrate imports *prior* to attempting full configuration resolution.

### [Blocker] Untested and unspecified `[[patches.agent]]` resolution on auto-included system packs
The design requires that "Gastown should not import Core; Core remains a required host system pack." It also states that "Gastown may continue to patch the Core `dog` agent for theming or work_dir behavior."
This means Gastown's `pack.toml` will contain:
```toml
[[patches.agent]]
name = "dog"
...
```
However, the `dog` agent is defined in the Core system pack. Normally, the patch mechanism resolves patches against explicitly imported packs. Because Gastown does not import Core, whether `[[patches.agent]]` can resolve and apply patches to agents defined in auto-included system packs (loaded via `builtinPackIncludes` rather than explicitly declared imports) is an untested assumption. If this resolution fails or is disallowed by config-layering rules, the `dog` agent in Gastown cities will run unthemed and unpatched, breaking custom directory structures and behaviors.

**Required change:**
The design must explicitly specify and document how the config-layering system resolves patches against auto-included system packs without explicit imports. `TestPinnedPublicGastownBehavior` must include a concrete assertion proving that Gastown's patches to the `dog` agent are successfully applied when loaded alongside Core.

### [Blocker] Unresolved template fragment dependencies for the Core `dog` prompt
The Core `dog` prompt template relies on fragments (`propulsion-dog`, `architecture`, `following-mol`) that currently live in `examples/gastown/packs/maintenance/template-fragments/`. 
The design splits these fragments, moving some to Core and others to Gastown. 
However, if the Core `dog` prompt template still references fragments that are moved to Gastown, or if Gastown prompt templates reference fragments moved to Core, and there is no explicit import chain between Core and Gastown, the Go template executor will fail to resolve the templates at runtime.
This will cause the `dog` agent to crash or fail to initialize on startup.

**Required change:**
The design must explicitly map the final destination of every template fragment used by `dog` and other core/Gastown agents. Ensure that any template references in Core-owned prompt files only target Core-owned fragments, and any references in Gastown target Gastown fragments. Add a packcompat template resolution test to prove cross-pack prompts compile and resolve correctly.

---

## Major Risks

### [Major] Stale `examples/gastown/packs/` path matching in `legacyPublicPackForSource` rewrites local developer setups
`legacyPublicPackForSource` matches local path prefixes like `examples/gastown/packs/` and rewrites them to the public remote repository. 
However, during local development or custom testing, a developer might intentionally point their imports to `examples/gastown/packs/` to test changes locally before pushing them. If `gc doctor --fix` blindly rewrites these paths to the public remote, it will disrupt and overwrite local developer workflows.

**Required change:**
`gc doctor --fix` must check if the matched `examples/` path represents an active, intentional local development path (e.g., matching a directory that exists in the current source checkout) before automatically rewriting it to a public remote.

### [Major] Stale public Gastown script and prompt references to `packs/maintenance` or `.gc/system/packs/maintenance`
While the design specifies updating the public Gastown `pack.toml` comments to remove references to `maintenance`, it does not mandate a complete audit of scripts, prompts, and formulas in the `gascity-packs` repository for hardcoded paths referencing the retired `.gc/system/packs/maintenance/` directory. 
If any such references (e.g., sourcing `_bd_trace.sh` or `dolt-target.sh`) remain in Gastown scripts, they will fail silently or loudly at runtime since the maintenance pack is no longer materialized.

**Required change:**
The behavior-preservation manifest and the `TestPinnedPublicGastownBehavior` gate must enforce a clean scan of all files in the public Gastown pack, asserting that no file contains a hardcoded `.gc/system/packs/maintenance/` or `packs/maintenance/` path, and that all references are updated to relative paths or the new `.gc/system/packs/core/` paths.

### [Major] Unmanaged state files in `.gc/runtime/packs/maintenance/` after retirement
Migrated scripts like `jsonl-export.sh` or `spawn-storm-detect.sh` will write state to Core-managed directories (like `packs/core/`). However, existing cities have legacy state files (e.g., jsonl-export cursors, archive metadata, spawn-storm ledgers) in `.gc/runtime/packs/maintenance/`. 
If these are ignored and left behind, upgraded cities will lose their operational history, causing the JSONL exporter to export from scratch or spawn-storm detectors to lose history and re-trigger false alerts.

**Required change:**
Specify whether `gc doctor --fix` or a one-time migration step will safely relocate existing state files from `.gc/runtime/packs/maintenance/` to `.gc/runtime/packs/core/` to preserve continuity for upgraded cities.

---

## Missing Evidence

- No specification of how `config.Load` behaves when a city imports `.gc/system/packs/gastown` but the directory does not exist on disk.
- No unit or integration test proving `[[patches.agent]]` works without an explicit `[imports.core]` entry.
- No audit of template fragment dependencies within the `dog` prompt template.
- No scan or validation of the public `gascity-packs` repository to ensure no hardcoded `.gc/system/packs/maintenance` paths exist in its scripts.

---

## Required Changes

1. **Deadlock Resolution:** Update the config loader to allow lenient loads during diagnostic/doctor runs, or update doctor to parse TOML ASTs before full config resolution.
2. **Auto-Included Patch Resolution:** Document and test the patch resolution mechanism for `[[patches.agent]]` on auto-included system packs.
3. **Template Fragment Audit:** Explicitly map every template fragment used by `dog` and ensure no cross-pack unresolved dependencies.
4. **Developer Local Path Protection:** Prevent `legacyPublicPackForSource` from rewriting active local development paths.
5. **Gastown Paths Audit:** Enforce a check that public Gastown scripts and prompts do not reference retired `maintenance` paths.
6. **State File Migration:** Specify a safe migration path for legacy runtime state under `.gc/runtime/packs/maintenance/`.

---

## Questions

- Does the config layering system allow a pack to `[[patches.agent]]` an agent defined in an auto-included system pack (loaded via `builtinPackIncludes`) without an explicit import?
- Can `gc doctor` load a city's configuration to fix it if an import inside `city.toml` is unresolved/missing?
- Should `legacyPublicPackForSource` ignore paths that represent local directories in a developer's clone?
- Will the public Gastown pack's scripts be audited for `.gc/system/packs/maintenance/` paths before the first release of the split?
