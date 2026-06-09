# Marta Hidalgo

**Persona verdict:** block

**Source reviewed:** `.gc/design-reviews/ga-j195cx/attempt-1/design-after.md`

**Mandate:** daemon `formula_v2` gate; actionable errors; API, CLI, dashboard, and order consistency; warning rate limiting.

**Findings:**

- [Blocker] The design still does not define an executable cross-surface diagnostic contract. It names a Go-side `Diagnostic` shape and says CLI, API, dashboard, orders, controller, and convergence preserve the same fields, but it does not define the Huma response structs, JSON presence rules, generated TypeScript expectations, CLI structured-output shape, or event payload schema. This matters for operator diagnostics because required evidence such as an empty `SourceValue`, omitted `HostCapability`, or false/empty warning state can disappear or be rendered differently across surfaces.

- [Blocker] Order diagnostic behavior is underspecified and can still flood operators. The design says order dispatch emits one typed order failure event per failed dispatch attempt and that repeated deprecation warnings use `OnceKey` suppression, but it does not name the event type, payload registration, correlation with existing order failure events, or budget semantics. A poll-driven dispatcher can repeatedly hit the same disabled-host formula and emit repeated fatal events unless the design defines grouping, suppression lifetime, occurrence counting, and first/last-seen state for fatal diagnostics as well as warnings.

- [Major] Warning cadence is not durable enough for daemon and CI loops. CLI warnings are once per command invocation, which is fine for an interactive command but not for automation that repeatedly runs validation or dispatch. For controller/order paths, `OnceKey` exists as a field but there is no ownership boundary for storing suppression state, no reset rules for config reloads or source changes, and no LRU/eviction behavior. The current text can still produce log and event floods during normal polling.

- [Major] CLI exit semantics are too vague for automation. "Fatal diagnostics exit non-zero" does not let scripts distinguish formula/user diagnostics from process, filesystem, or infrastructure failures. If `gc sling --formula`, order dispatch helpers, and validation/report commands differ between exit `1` and exit `2`, operators get inconsistent remediation despite identical diagnostic codes.

- [Major] Dashboard and dashboard-less operator read surfaces are promised but not pinned. The design says dashboard state derives from API diagnostic projection, but it does not name the route, response type, dashboard card state, or any CLI/status command that lets an operator inspect current grouped diagnostics, suppression status, occurrence counts, and remediation without reading raw events.

- [Major] The canonical remediation contract is implied rather than specified. The design says projections must not synthesize alternate wording, but it does not provide canonical message/remediation values per diagnostic code or fixtures proving that `requires`, legacy `contract`, and dual-declared sources all render the same host-capability remedy when `[daemon] formula_v2 = false`.

**Missing evidence:**

- Golden fixtures for `formula.compiler_requirement_unsatisfied` across CLI stderr/stdout, API/Huma JSON, generated TypeScript, dashboard state, order event payloads, controller errors, and convergence errors.
- Presence/absence fixtures proving diagnostic source fields survive empty strings, omitted values, wrong TOML types, false-like host capability values, and warning-only results.
- Event registry tests naming every diagnostic/order event constant and registered payload, including compatibility behavior for existing `order.failed` consumers.
- Dispatcher-loop fixtures showing repeated disabled-host formulas and repeated deprecation warnings do not flood events or logs.
- Restart, config reload, formula source edit, and host toggle fixtures proving when warning and fatal diagnostic suppression resets.
- Exit-code fixtures for `gc sling --formula`, `gc order`, formula validation, API sling failure mapping, and controller/order dispatch failure paths.
- Dashboard and dashboard-less operator fixtures for current diagnostic state, occurrence counts, first/last seen timestamps, suppression state, and remediation.

**Required changes:**

- Define a presence-aware typed diagnostic projection for every operator surface: Go structs, Huma/OpenAPI schema, generated TypeScript, CLI structured output or stderr contract, dashboard state, and event payloads.
- Pin order diagnostic events: exact event type names, registered payloads, whether generic and formula-specific failure events both emit, correlation keys, burst-budget accounting, and compatibility behavior for existing listeners.
- Add durable diagnostic grouping/suppression rules for warnings and fatal host-capability failures, including state owner, `OnceKey` construction, reset conditions, occurrence counting, LRU/eviction behavior, and first/last-seen timestamps.
- Specify CLI exit classes for formula diagnostics versus infrastructure failures, and require all formula-launching commands to follow the same classification.
- Define the operator read surface for grouped diagnostics, including one dashboard-backed API route and one dashboard-less CLI/status path.
- Add canonical message/remediation fixtures for every diagnostic code, with parity across `requires`, legacy `contract`, and dual declarations.
