# Anand Krishnaswamy

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek

**Consensus findings:**
- [Blocker] Core still has live role-literal defaults with no neutral end-state in the design. Claude and DeepSeek both identify `DefaultCity` and `WizardCity` scaffolding `mayor`, `prompts/mayor.md`, and a `mayor` named session in `internal/config/config.go`. AC8 covers init/template resolution, but the plan does not say whether the SDK default city emits zero agents, symbolic bindings, or another role-neutral scaffold.
- [Blocker] `gc start` warmup still has a Go fallback to `mayor`. Claude and DeepSeek both cite `cmd/gc/cmd_start_warmup.go` using `defaultWarmupMailTo = "mayor"`. That contradicts the plan's own "No Go fallback may substitute `mayor`" rule unless the warmup recipient is moved to a symbolic binding or skipped with a typed diagnostic when unbound.
- [Blocker] Runtime tmux surfaces remain compile-bound to role names. Claude and DeepSeek both cite the `roleEmoji` / `roleIcons` map, role-named theme constructors, and `ConfigureGasTownSession`. These are Go runtime surfaces, not pack assets, so "move to pack manifest rows" is not a sufficient de-roling plan.
- [Major] Behavior-bearing target fields need structural binding evidence, not only token scanning. Codex requires every route, mail, nudge, warmup, formula, prompt fallback, and route metadata consumer to record binding-resolution provenance and to reject direct literal targets unless the surface is documented inactive/historical data. Claude and DeepSeek reach the same risk from the scanner side: a denied-token list can miss embedded names, arbitrary new role names, map keys, identifiers, and generated outputs.
- [Major] The `dog` and retired-identity exceptions are too broad or incomplete for a zero-hardcoded-roles gate. Codex says the `dog` exception must be pinned to exact config keys, rendered-output expectations, and negative fixtures. DeepSeek notes `dog` is omitted from the denied set. Claude also asks for the exact allowed position for literal `dog` and calls out unclassified `coordinator` / `health-check` role-key surfaces.
- [Major] Required-versus-optional binding semantics risk a ZFC violation unless declared as data. Claude and DeepSeek both note that Go must not decide a binding is required or optional from its name, provider, role, or purpose. Formula, order, and pack metadata must carry this classification, and dispatch must read it as pure transport.
- [Major] Scanner and allowlist contracts remain under-specified. Claude notes `boot` substring false positives, `gastown` module-path false positives, embedded role names in identifiers and formula names, and missing map-key coverage. DeepSeek adds missing allowlist-expiry enforcement. Codex adds that public Gastown role strings must pass only in the pinned public Gastown pack, never in Core, stale materialized packs, generated SDK defaults, or Go routes.
- [Major] Provider-pack routes and prompt/template suffixes are not fully covered. DeepSeek flags hardcoded `mayor` / `deacon` mail and nudge routes in required `bd` / `dolt` provider-pack behavior and suffix role names hidden behind `binding_prefix`; Claude similarly finds the current binding model maintenance-worker-specific rather than a general symbolic target mechanism.

**Disagreements:**
- Verdicts differ: Claude blocks, Codex approves with risks, and DeepSeek blocks. My assessment follows the blocking reviewers because the disputed items are live Core/SDK role-leak surfaces with no specified neutral end-state.
- Codex is more optimistic about the design direction and treats the binding model as acceptable if strengthened with auditable resolution evidence and narrow exceptions. Claude and DeepSeek require the design to name the concrete de-roling end-state for current Go defaults, warmup, and tmux before implementation begins. These are compatible requirements, not mutually exclusive positions.
- Claude focuses on concrete live Go violations and scanner semantics; Codex focuses on structural target-field validation and Core/public-Gastown scan modes; DeepSeek adds prompt suffixes and provider-pack routes. The combined persona finding is that the design must remove today's role leaks and add structural validation so new literal roles cannot bypass a historic deny list.

**Missing evidence:**
- The role-neutral end-state for `DefaultCity`, `WizardCity`, and `gc init` scaffolding.
- The warmup recipient contract, including whether warmup is Core SDK infrastructure or public Gastown behavior and how `gc start` behaves when the relevant binding is omitted, disabled, or renamed.
- The de-roling implementation for `internal/runtime/tmux`: icon selection, theme assignment, role-named API deletion or rename, `ConfigureGasTownSession` rename, and treatment of `coordinator` / `health-check`.
- A binding-resolution evidence schema covering consumer id, consumer kind, source pack, binding key, required/optional mode, binding source, resolved target, absence behavior, diagnostic id, and behavior-driving status.
- A parser-level validator for behavior-bearing target fields across Core and required system-pack formulas, orders, metadata, mail/nudge recipients, warmup defaults, prompt fallbacks, generated help defaults, API/dashboard target fields, scripts, and tmux selectors.
- A machine-readable operation manifest covering every Core formula, order, script, check, warmup, and maintenance workflow with execution owner, binding key, required/optional class, default source, no-executor result, diagnostic id, and proof command.
- The implementation contract for `[gc.bindings.*]`, `[system_packs.*.bindings]`, `target_binding`, `gc.run_target_binding`, and `GC_CORE_MAINTENANCE_WORKER`, including config structs, patch/merge behavior, precedence, provenance, empty values, environment fallback, and literal-target conflicts.
- How prefix-only `binding_prefix` is replaced or constrained so templates and scripts cannot keep hardcoded role suffixes.
- Required provider-pack treatment for `bd` and `dolt` mail, nudge, escalation, and notification routes.
- The scanner schema, generated path, generator/check command, freshness test, CI owner, allowlist format, expiry enforcement, active-vs-historical classification, token boundary rules, identifier/map-key matching, module-path false-positive rule, and positive/negative fixtures.
- The exact allowed context for `dog`, role-identity uses of `maintenance` / `Maintenance`, `gastown`, `boot`, `coordinator`, and `health-check`.

**Required changes:**
- Define the neutral default scaffold for `gc init`, `DefaultCity`, and `WizardCity`, and add an AC8 proof that the non-Gastown SDK default path contains no concrete role literal or prompt path.
- Replace `defaultWarmupMailTo = "mayor"` with a symbolic binding or a skip-with-typed-diagnostic path, and state whether warmup is Core SDK infrastructure or public Gastown behavior.
- De-role `internal/runtime/tmux`: move role-to-icon/theme selection into config/pack data or a generic deterministic mechanism, delete or rename role-named theme APIs, and rename `ConfigureGasTownSession` to a neutral API.
- Add structured validation and binding-resolution evidence for behavior-bearing target fields so Core active behavior cannot contain arbitrary literal agent/session/role targets, even when the literal is absent from the historic denied-token list.
- Add a machine-readable Core/provider-pack operation manifest that drives no-executor tests and rejects required user-agent bindings for controller-owned SDK infrastructure.
- Declare required/optional binding semantics in formula/order/pack data and validate them before dispatch; Go must not classify bindings by name, provider, or purpose.
- Expand the scanner contract to cover strings, identifiers, map keys, generated/materialized outputs, active roots, and provider packs; enforce allowlist expiry; include token-boundary and module-path false-positive rules; and add controls for arbitrary literal roles, embedded names, and historical fixture leakage.
- Narrow the `dog` exception to exact configuration paths and exact tests, with owner, reason, expiry or explicit permanent-design rationale, token kind, rendered-output expectation, and negative fixtures proving it cannot move into active routing, notification, prompt, formula, generated help, API/dashboard, or Go code.
- Add role-identity uses of `maintenance` / `Maintenance`, `coordinator`, and `health-check` to the role-surface taxonomy with narrow allowed contexts. Literal role names must fail in `target_binding`, `gc.run_target`, `gc.routed_to`, recipients, mail/nudge/warmup targets, script bodies, prompt suffixes, and theme targets unless the row is a documented non-active fixture.
- Rebind required provider-pack mail/nudge/escalation routes through symbolic bindings and fail CI on hardcoded literal role routes in required provider-pack active behavior.
