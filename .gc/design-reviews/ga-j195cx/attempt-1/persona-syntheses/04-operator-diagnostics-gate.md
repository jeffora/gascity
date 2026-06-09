# Marta Hidalgo

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The operator diagnostic contract is still not executable across CLI, API/Huma, generated TypeScript, dashboard state, and event payloads. Codex flags missing response structs, JSON presence rules, CLI structured-output shape, and event schema; Claude flags that fatal diagnostics do not share the same canonical no-local-advice string contract already proposed for warnings. Without a single typed projection and canonical fatal remediation strings, disabled-host failures can drift across surfaces.
- [Blocker] Order diagnostic behavior is not pinned tightly enough for operators. Codex finds no exact event type, payload registration, compatibility story for existing order-failure consumers, or durable burst-budget semantics. Claude adds that diagnostic group lifecycle does not cover removed orders, so stale order rows can persist after config churn. The design needs immutable event append semantics plus a durable grouped state/read-model contract.
- [Major] Dashboard and dashboard-less rollup behavior needs a fixture-locked mapping. Claude identifies the missing diagnostic-code-to-`RollupClass` table that keeps host/config problems separate from formula/pack problems; Codex identifies the missing operator read surface for grouped diagnostics, occurrence counts, suppression state, and remediation. This is load-bearing because the persona goal is to make disabled capability errors look operator-fixable, not like parse failures.
- [Major] Warning and fatal cadence semantics remain under-specified at the edges. Claude asks for precise `--warnings=once` scope and interaction with `--quiet`; Codex asks for ownership, reset, LRU/eviction, occurrence counting, and first/last-seen rules for repeated daemon and CI loops. These decisions must be explicit for process-local CLI output, controller/order paths, restarts, config reloads, source edits, and host toggles.
- [Major] CLI and automation classification is too vague. Codex flags that "fatal diagnostics exit non-zero" does not distinguish user/formula diagnostics from process, filesystem, or infrastructure failures across `gc sling --formula`, order dispatch helpers, validation, and report commands. The design should either standardize a diagnostic exit class, preferably exit `2`, or explicitly require automation to consume typed output rather than `$?`.
- [Major] The disabled-host path can still create sequential remedy discovery. Claude notes that when a formula has both a fatal author-side requirement problem and `[daemon] formula_v2 = false`, host satisfaction is skipped and the operator only learns about the daemon flag after fixing the author error. A typed non-blocking host attribution should surface the known second remedy in the same response.
- [Minor] Several diagnostic lifecycle and edge cases are unnamed. Claude identifies no diagnostic code for unsupported stores that cannot provide singleton/CAS guarantees, no named diagnostic/remediation for `[daemon].graph_workflows` promotion, no fixture proving metadata-spillover diagnostics project correctly, and no owner/TTL for `expired` rollup rows.

**Disagreements:**
- Claude verdict is `approve-with-risks`; Codex verdict is `block`. I choose `block` because the missing typed wire/event/read-surface contracts would force implementers to make ad hoc choices in code, which is exactly what this operator-diagnostics gate is meant to prevent.
- Claude sees the normalized host gate and burst-budget structure as mostly sound, while Codex says order events and warning cadence can still flood operators. My assessment is that the concept is acceptable, but the design must name the event set, grouping state owner, reset rules, and compatibility behavior before implementation.
- Claude focuses on fatal-string parity, rollup class mapping, conflict precedence, unsupported-store diagnostics, and cleanup. Codex focuses on executable cross-surface contracts, order event semantics, durable suppression, exit classes, and read surfaces. These are complementary rather than contradictory.
- Claude raises API status-code rationale as a question, while Codex does not. This should remain a documentation/fixture decision unless the design relies on retry semantics from API clients.

**Missing evidence:**
- Golden parity fixtures for `formula.compiler_requirement_unsatisfied` and other fatal diagnostics across CLI stderr/stdout, Huma JSON, generated TypeScript, dashboard state, order event payloads, controller errors, and convergence errors.
- Presence/absence fixtures proving source values survive empty strings, omitted defaults, wrong TOML types, false-like host capability values, JSON nulls, and warning-only results across every operator surface.
- A generated `{diagnostic code -> RollupClass}` fixture, including host/config, formula source, pack source, and internal/store failures.
- Event registry tests and migration fixtures naming the formula diagnostic event set, payload registrations, correlation keys, and compatibility behavior for existing order-failure listeners.
- Dispatcher-loop fixtures showing repeated disabled-host formulas, repeated warnings, restarts, config reloads, source edits, host toggles, and order deletion do not flood logs/events or leave stale rows.
- Fixtures for `--warnings=once`, `--quiet`, JSON output, launch-command exit classes, and dashboard-less diagnostic status output.
- Metadata-spillover fixtures proving diagnostics stored in accepted artifacts or referenced compile artifacts remain visible in rollups after restart.
- A named unsupported-store diagnostic path that proves protected writes fail closed with zero partial writes.
- A named `[daemon].graph_workflows` promotion/deprecation diagnostic with canonical remediation.

**Required changes:**
- Define a presence-aware typed diagnostic projection for every operator surface: Go structs, Huma/OpenAPI schema, generated TypeScript, CLI structured output or stderr contract, dashboard state, and event payloads.
- Add canonical message/remediation constants or fixtures for every fatal diagnostic code, and state that no surface may append local advice except through a separate typed field.
- Add a generated and CI-locked `{diagnostic code -> RollupClass}` table; map disabled host capability failures to `host_config`, parse/type/conflict failures to `formula_source`, pack-floor failures to `pack_source`, and unsupported store failures to `internal`.
- Pin order diagnostic event behavior: exact event names, registered payloads, generic-versus-specific emission, correlation keys, SSE/generated-client coverage, burst-budget accounting, and compatibility for existing listeners.
- Define durable grouped diagnostic state: owner, singleton/CAS capability requirements, unsupported-store fail-closed diagnostic, occurrence counts, first/last-seen timestamps, LRU/eviction behavior, and reset rules for restart, no-op reload, formula-relevant config change, source change, host toggle, and order removal.
- Define one dashboard-backed API/read-model route and one dashboard-less CLI/status path for current grouped diagnostics, suppression state, occurrence counts, and remediation.
- Specify launch-command operator semantics for `gc sling --formula`, `gc order`, validation, report, and API-routed launches; prefer exit `2` for formula/host diagnostics and reserve exit `1` for process or infrastructure failures unless the design explicitly chooses typed-output-only automation.
- Add a fixture for disabled host plus author-side requirement conflict that returns the fatal author diagnostic and a typed informational host attribution in the same response.
- Reserve diagnostic codes and canonical remediation for unsupported diagnostic store behavior and `[daemon].graph_workflows` promotion/deprecation.
- Specify `--warnings=once` semantics, including accepted commands, scope per `OnceKey`, interaction with `--quiet`, and JSON-output behavior.
