# Ingrid Kovac

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash (current dependency output is the `_gemini.md` artifact; an older `_deepseek.md` artifact remains in the directory but was not used)

**Consensus findings:**
- [Blocker] The design still has an unresolved `agent_kind`/`crew` wire contradiction. The role-surface table says API/OpenAPI/dashboard `crew` vocabulary should be replaced with neutral grouping and regenerated schemas, while the testing section says the migration should not require dashboard changes. Reviewers identify live API, OpenAPI, generated TypeScript, and dashboard consumers, so the design must choose compatibility or a real wire/UI migration.
- [Blocker] Production Go role surfaces need explicit disposition rows. The reviews identify tmux role themes and icons, `ConfigureGasTownSession`, warmup mail defaults, embedded mayor prompt fallback, default/wizard city scaffolding, `mol-polecat-*` / `mol-refinery-patrol` sling heuristics, `crew` classification, and `gastown` init/config/public-pack wiring. A Core asset scanner alone cannot satisfy the project invariant that Go source must not retain hardcoded roles.
- [Blocker] `dog` containment is not yet enforceable end to end. The symbolic `core.maintenance_worker` binding model is sound, but the canonical implementation section does not require `target_binding` / `gc.run_target_binding` / `GC_CORE_MAINTENANCE_WORKER` for every worker-bound Core asset. Literal worker targets such as `pool = "dog"` must be rejected in Core unless covered by a narrow compatibility row.
- [Major] Core infrastructure is not classified as controller-owned versus optional worker-bound maintenance. Gate sweep, blocker-close cascade, orphan sweep, order tracking, spawn-storm detection, wisp compaction, reaper, JSONL export, binary doctor checks, formulas, orders, and scripts need per-asset ownership so required SDK operations do not depend on a configured maintenance worker.
- [Major] Scanner semantics still need a binding contract even though Claude credits the latest draft with improved camelCase/PascalCase detection. Codex and DeepSeek require explicit roots, file types, comment handling, sub-identifier matching, generated-wire handling, and field-scoped allowlists. Public Gastown assets should be inventoried for preservation, not rejected as Core role violations, and generated audit tables need narrow self-reference exemptions.
- [Major] If Go comments remain in scanner scope, live role examples in production comments must be dispositioned or rewritten. DeepSeek identifies `internal/api/handler_agents.go` as a concrete example; without a comment cleanup rule or field-scoped allowlist, the scanner can block on explanatory comments rather than behavior.
- [Major] Active `gastown` production literals require a policy decision. `gc init --template gastown`, `PublicGastownPackSource` / `PublicGastownPackVersion`, import-state doctor cases, and public-pack construction may be sanctioned knowledge of the canonical public pack or may need data/registry-driven resolution. The design cannot leave them to the generator while also claiming ZERO hardcoded roles.
- [Major] The sling formula-name heuristic replacement must preserve Core behavior. `mol-scoped-work` shares base-branch behavior with Gastown `mol-polecat-*` formulas; replacing role-name heuristics with formula metadata needs a witness row proving Core `mol-scoped-work` behavior survives.

**Disagreements:**
- Claude says the scanner now catches sub-identifiers and the remaining block is narrower; Codex and DeepSeek still treat scanner semantics and exemptions as underspecified. Assessment: keep the stronger scanner claim only if the design spells out the contract and fixtures in implementation-facing text.
- The reviewers offer two valid `crew` paths: preserve `crew` as a neutral structural UI bucket, or rename the wire/API/dashboard surface. Assessment: either path is acceptable only with a named rollout slice, tests, and the required dashboard/OpenAPI gates.
- Core `gastown` knowledge can be a sanctioned public-pack identity or a registry-driven lookup. Assessment: the design must document the policy and bound the allowed fields so `gastown` literals cannot leak into prompts, routing, formula selection, or role-conditioned behavior.
- `dog` can be a configurable default maintenance worker or purely a Gastown convention. Assessment: Core must behave through symbolic bindings either way, and omitted/renamed-worker tests must prove no fallback to literal `dog`.

**Missing evidence:**
- Exact scanner contract: roots, file types, TOML fields, metadata fields, generated help/schema artifacts, scripts, templates, prompt snippets, prohibited token set, word-boundary and camelCase/PascalCase rules, comment handling, allowlist format, owner/expiry fields, and negative fixtures.
- `role-surface.generated.yaml` rows for the real production Go surfaces: theme APIs, role icon maps, init/default city scaffolding, warmup mail default, prompt fallback, sling formula heuristics, API/dashboard `crew`, public-pack constants, active `gastown` template wiring, and Go comments if comments are in scope.
- Per-operation asset classification as `controller_owned`, `optional_core_maintenance_worker`, `provider_pack`, `public_gastown`, `compatibility`, or `retired`.
- End-to-end renamed-worker and omitted-worker tests proving controller-owned operations still run, and no dispatch, event, bead metadata, mail, nudge, pool, route target, formula step, or order falls back to literal `dog`, `mayor`, or `deacon`.
- Field-scoped compatibility rows for retained active role names such as `mol-dog-*`, provider scripts, docs/CLI output, order-skip compatibility, routing metadata, and `crew` wire labels, each with binding-based behavior evidence and expiry or permanent rationale.
- The policy and tests for active `gastown` init/config/public-pack constants.
- A behavior-manifest row preserving `mol-scoped-work` base-branch behavior in Core after replacing formula-name heuristics.
- API/OpenAPI/dashboard impact analysis and checks for replacing or temporarily preserving `agent_kind="crew"`.
- The controller-owned-vs-worker-bound fixture showing that omitting the maintenance worker does not break SDK infrastructure or spam diagnostics.
- Comment-scope cleanup policy for active Go files, plus a self-reference exemption for generated audit manifests that must record forbidden tokens as evidence.

**Required changes:**
- Resolve the `agent_kind`/`crew` contradiction explicitly. Either preserve `crew` as a field-scoped neutral structural compatibility value with tests proving it drives no role behavior, or schedule the API/OpenAPI/generated TypeScript/dashboard rename and run `make dashboard-check`.
- Add a Go role-surface migration table with exact dispositions, owners, rollout slices, replacement mechanisms, and tests for the production surfaces identified by the reviewers. If Go de-roling is deferred, narrow any source-deletion gate that implies it is solved.
- Add a `role-surface.generated.yaml` disposition and prose policy for active `gastown` init/config/public-pack wiring: accepted canonical public-pack identity with a documented carve-out, or data/registry-driven template resolution.
- In the Core Maintenance and Notification Contract, require `[gc.bindings.maintenance_worker]`, `target_binding`, `gc.run_target_binding`, and `GC_CORE_MAINTENANCE_WORKER` as the only Core worker resolution path. State that literal worker targets in Core assets are defects except for field-scoped compatibility rows.
- Extend the behavior manifest or role-surface table with mandatory per-asset classification for every Core order, formula, script, moved helper, provider-pack behavior, and public Gastown behavior. Tie controller-owned rows to omitted-worker execution tests and worker-bound rows to clean inert behavior.
- Specify scanner behavior by ownership: reject forbidden role tokens in Core/provider-owned active assets, inventory public Gastown roles for behavior preservation, and exempt role-surface/audit artifacts narrowly enough that they can record violations without becoming broad allowlists.
- Require scanner fixtures for sub-identifier matching, comments if comments are in scope, string literals, TOML metadata, scripts, templates, generated-wire fields, and audit-manifest exemptions.
- If comments stay in scope, require cleanup or explicit disposition rows for active Go comments that mention role examples, including `internal/api/handler_agents.go`.
- State the tmux disposition directly: delete dead `MayorTheme` / `DeaconTheme` / `DogTheme` APIs if unused, and replace role icon/status behavior with config-supplied display metadata.
- Add a behavior-manifest row preserving `mol-scoped-work` base-branch behavior in Core when sling formula-name heuristics are replaced by formula metadata.
