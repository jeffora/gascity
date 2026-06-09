# Anand Krishnaswamy - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The role-surface manifest scope is broad enough for this migration: Go, TOML, shell, Markdown, templates, generated command text, API classifications, dashboard/OpenAPI generated references, tmux helpers, scaffolding, warmup mail defaults, prompt fallbacks, formulas, overlays, metadata, tests, docs, and public Gastown companion files (`design-before.md:278`-`design-before.md:282`).
- Core-owned behavior is explicitly barred from concrete Gastown role names except through allowlist rows with owner, justification, expiry, and negative fixtures; `dog` is confined to default configurable Core pack data and tests for that configuration (`design-before.md:284`-`design-before.md:288`).
- The plan introduces symbolic binding surfaces and states that Core maintenance formulas/orders resolve configured bindings rather than Go constants, with tests for renamed and omitted maintenance workers (`design-before.md:290`-`design-before.md:295`).

**Critical risks:**
- [Major] Omitted-binding behavior is not concrete enough. The plan says controller-owned SDK operations still work when the maintenance worker is renamed or omitted, but it does not define what happens to agent-executed Core maintenance work when no binding resolves. To preserve SDK self-sufficiency and ZFC, this must not fall back to `dog` in Go or silently synthesize a role; it should produce durable visible work, a diagnostic, or a controller-owned no-agent path.
- [Minor] `GC_CORE_MAINTENANCE_WORKER` is listed as a binding input, but its precedence relative to city config and pack config is not specified. If the environment override wins silently, operators can get role routing that is not traceable to resolved config; if it is only a config-layer input with provenance, the plan should say that.
- [Minor] The allowlist expiry/negative-fixture model is good, but the plan should require live-behavior scanners to fail if a compatibility fixture path becomes importable or materialized into active Core behavior.

**Missing evidence:**
- A binding-resolution truth table for default `dog`, renamed worker, omitted worker, disabled worker, environment override, and public Gastown override cases.
- The diagnostic or durable-work behavior when a Core maintenance formula/order references an unresolved binding.
- A provenance requirement showing where the selected binding came from and why it is treated as user-supplied config.
- A scanner fixture proving allowed historical/test role names cannot be loaded through normal runtime paths.

**Required changes:**
- Define omitted/disabled binding behavior explicitly and prohibit Go-side fallback to any concrete role name.
- Specify binding precedence and provenance across `[gc.bindings.*]`, `[system_packs.*.bindings]`, `target_binding`, `gc.run_target_binding`, and `GC_CORE_MAINTENANCE_WORKER`.
- Add tests for default, renamed, omitted, disabled, and environment-overridden bindings that prove controller-owned SDK operations do not require a configured user agent.
- Require role-name allowlists to include an active-runtime exclusion test, not only a token scan exemption.

**Questions:**
- When no maintenance-worker binding resolves, does Core leave work unassigned, route it to a controller-owned operation, or produce a diagnostic-only bead?
- Is `GC_CORE_MAINTENANCE_WORKER` a runtime override, an init-time default, or a config-layer value with normal provenance?
- Which scanner root proves generated dashboard/OpenAPI/types references cannot reintroduce live role-conditioned behavior?
