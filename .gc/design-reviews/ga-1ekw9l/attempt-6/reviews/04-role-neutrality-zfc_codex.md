# Design Review: 04-role-neutrality-zfc

Reviewer: Anand Krishnaswamy persona, Codex lane
Verdict: Block

Reviewed `.gc/design-reviews/ga-1ekw9l/attempt-6/design-before.md` for zero hardcoded roles, symbolic maintenance worker binding, SDK self-sufficiency, and ZFC judgment containment. I did not read the Claude review output for this item before writing this review.

## Findings

### Blocker: Maintenance is not covered by the active role-surface deny rule

The role-neutrality section says the generated role-surface manifest covers Go, TOML, shell, Markdown, templates, generated command text, API classifications, dashboard/OpenAPI generated references, tmux theme helpers, scaffolding, warmup mail defaults, prompt fallbacks, formulas, overlays, metadata, tests, docs, and public Gastown companion files (`design-before.md:331`). It then bans Core-owned behavior from containing or branching on `mayor`, `deacon`, `witness`, `refinery`, `polecat`, `boot`, `crew`, and `gastown`, with `dog` limited to the configurable maintenance-worker default (`design-before.md:337`).

That rule does not include `maintenance` or `Maintenance`, even though this migration is explicitly retiring Maintenance as a standalone pack and the persona contract asks whether Go, prompt assets, scripts, formulas, orders, generated help, or API routes still branch on Maintenance. The docs wording scanner distinguishes retired standalone Maintenance wording from valid lowercase maintenance and store-maintenance terms (`design-before.md:400`), but that is an operator-text gate, not an active behavior-surface gate. As written, a live Go/API/formula/order branch on a concrete `maintenance` identity can remain outside the role-surface deny rule as long as it is not caught by the separate retired-source classifier.

Required change: extend the role-surface manifest and scanner contract to classify `maintenance`/`Maintenance` tokens for active behavior surfaces, not just docs. Allow only explicitly typed contexts such as Core maintenance-worker binding vocabulary, Dolt/store-maintenance terminology, retired-pack diagnostics, or historical/test fixtures. Add a negative fixture proving a live Core branch, route, formula/order target, mail/nudge target, or generated help path containing `maintenance` fails unless it is in one of those typed contexts.

### Blocker: Allowlist expiry is required as data but not enforced as a gate

The plan requires historical/test allowlist rows to carry owner, justification, expiry, and negative fixtures (`design-before.md:337`, `design-before.md:477`). The test plan says scanner tests reject active Core role-name references outside generated allowlists (`design-before.md:508`).

That still leaves expiry as metadata only. There is no rule that an expired row fails CI, no date or release-bound comparison source, and no fixture that proves expired compatibility rows cannot continue authorizing live role references. This is exactly the failure mode the persona is meant to catch: role debt can become permanent under a neutrality banner.

Required change: specify the enforcement semantics for role-surface and wording allowlist expiry. The scanner should fail closed when `expires_at` or `expires_after_release` has passed, fail when an allowlist row lacks an expiry except for named permanent vocabulary classes, and include fixtures for an expired row and a live-behavior row that attempts to hide behind a historical/test exception.

### Major: Orders are not explicitly included in the role-surface scanner scope

The role-surface manifest list includes formulas and many generated/documented surfaces, but it does not explicitly include orders (`design-before.md:331`). Later sections mention Core maintenance formulas and orders resolving bindings (`design-before.md:343`) and representative Core formula/order tests (`design-before.md:520`), but those are behavior tests, not the manifest/scanner coverage that proves hardcoded route, mail, nudge, or target metadata cannot survive in order assets.

Required change: add orders and order metadata to the role-surface manifest scope and scanner fixtures. The test contract should fail on an order that hardcodes `dog`, `mayor`, `maintenance`, Gastown role names, or concrete route/mail/nudge targets instead of using declared binding metadata.

## What Passes

The binding model is moving in the right direction. The plan defines `[gc.bindings.*]`, `[system_packs.*.bindings]`, `target_binding`, `gc.run_target_binding`, and `GC_CORE_MAINTENANCE_WORKER`; it states that formulas and orders resolve configured bindings rather than Go constants; and it explicitly forbids Go fallback to `mayor`, `deacon`, `dog`, or another concrete role (`design-before.md:343`, `design-before.md:350`). It also requires tests proving Core-only cities load and controller-owned SDK operations still work when the maintenance worker is renamed or omitted (`design-before.md:346`).

The remaining blockers are about making the scanner and allowlist gates as strict as the prose. Once Maintenance is handled as an active behavior token, orders are in scanner scope, and allowlist expiry is executable, this persona should pass.
