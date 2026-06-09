# Marta Hidalgo

**Persona verdict:** block

**Scope reviewed:** `engdocs/design/formula-compiler-requirements.md` for the operator diagnostics gate: `[daemon].formula_v2` failures, actionable remedies, CLI/API/dashboard/order consistency, and warning flood control.

## Findings

### Blocker: API-routed CLI diagnostics are not covered

The design specifies direct CLI stderr behavior and Huma API response behavior as separate projections, but Gas City's API control-plane invariant says CLI mutations can route through the local HTTP API via `cmd/gc/apiroute.go:apiClient()` when a supervisor is running. That means the same operator command can take two paths:

- no supervisor: direct CLI compiles and renders the formula diagnostic
- supervisor running: CLI receives a generated-client HTTP error from the API projection

The design does not require the API-routed CLI path to decode the typed diagnostic body and render the same code, message, remediation, source path, host capability, and exit behavior as the direct path. Without that rule, `gc sling --formula` can produce the canonical `formula.compiler_requirement_unsatisfied` stderr in one runtime mode and a generic HTTP 400/409 failure in another. That is exactly the operator inconsistency this persona gates.

Required change: add an explicit projection rule and acceptance test for the generated-client CLI path. The test should run the same disabled-`formula_v2` formula launch with and without supervisor routing and assert identical diagnostic code/remediation and exit class.

### Blocker: the event contract still leaves compile-failure projection ambiguous

The diagnostic matrix says `formula.compiler_requirement_unsatisfied` emits a registered `formula.compile_failed` "or order failure event". That "or" is not implementable as a stable operator contract. Gas City's event/SSE model requires exact event constants in `events.KnownEventTypes` and exactly registered payload shapes via `events.RegisterPayload`; `order.failed` is currently a typed event variant with `NoPayload`, while the design also requires compile-failure payload fields such as `code`, `source_path`, `host_capability`, and `remediation`.

The design must choose the event shape per surface:

- whether compile diagnostics use a new `formula.compile_failed` event, an upgraded `order.failed` payload, or both
- which surfaces emit which event: order dispatch, controller, convergence, API-triggered launch, and direct CLI if any
- what goes in the event envelope `Subject` and `Message` versus the typed payload
- how existing order-failure consumers remain compatible if `order.failed` gains a payload

Without this, dashboard/event-stream consumers cannot rely on a stable event type or payload, and implementers can satisfy the prose while still producing incompatible event projections.

Required change: replace the "or" wording with an event projection table naming each event constant, payload struct fields, registration owner, envelope semantics, OpenAPI/SSE/generated-type impact, and compatibility behavior for existing `order.failed` consumers.

### Major: HTTP status classification is still too endpoint-dependent to verify

For `formula.compiler_requirement_unsatisfied`, the matrix says "HTTP 400 for formula input or 409 for launch conflict". That is directionally useful, but it does not map the classification to named API operations. The ambiguous cases are exactly the user-visible ones: sling launch, attach, order-backed launch, convergence create/retry, formula preview/validate, and dashboard-consumed projections.

Required change: add an endpoint/action matrix that pins the status for each Huma endpoint and any generated-client CLI adapter path. This should include the response struct fields that carry `Diagnostic`, not just the status code.

### Major: expansion/aspect provenance is not guaranteed in operator-facing diagnostics

The design correctly says expansion and aspect formulas contribute to the maximum normalized requirement, and root metadata can persist compile provenance. But the operator-facing `Diagnostic` shape and event payload fields only require `formula`, `source_path`, `source_key`, and `source_value`. They do not require both the invoked root formula and the contributing expansion/aspect formula that raised the requirement, nor the resolution chain or pack binding.

For a disabled host capability, "enable `[daemon] formula_v2`" is only half the remedy. Operators also need to know which imported pack/formula introduced the requirement so they can choose a v1 formula, update a pack pin, or fix the right authored file.

Required change: require diagnostics/events/API bodies for aggregated requirements to include root formula, contributing source formula/path, and enough provenance to identify the pack or layer winner. If full provenance is too large for every error, define a compact summary plus a stable reference to the compile provenance artifact.

### Major: warning suppression is close, but the storage contract needs one more pin

The `OnceKey` shape and warning-only suppression rule address the main flood risk. The remaining gap is that "in-memory, bounded LRU per city" does not name the owning component, capacity, eviction behavior, or concurrency expectations. A bounded cache without a capacity is easy to implement inconsistently across controller, order dispatch, and daemon paths.

Required change: name the owner/helper for warning suppression and set the default capacity or config source. Tests should prove repeated legacy-contract dispatch warnings do not emit on every scan, while fatal compile failures still emit according to the chosen order/controller failure cadence.

## Positive Signals

- The diagnostic code catalog, canonical message/remediation table, and projection rules are a substantial improvement over earlier versions.
- Explicit `HostCapabilities` per compile removes the package-global decision point from the core requirement-satisfaction path.
- The design now clearly gates new roots/wisps/convergence instances before durable writes when `[daemon].formula_v2` is disabled.
- Warning suppression is correctly scoped to projections rather than hidden inside requirement normalization.

## Required Acceptance Tests

- Direct CLI versus API-routed CLI disabled-capability launch produce the same diagnostic code, remediation, source fields, and exit class.
- Huma endpoint tests cover each pinned status/body mapping for formula validation, sling launch/attach, and dashboard-consumed API projections.
- Event/SSE tests prove the chosen compile-failure event payload is registered, generated, and includes the diagnostic fields without hand-written JSON.
- Order dispatch with a legacy `contract = "graph.v2"` formula emits at most one deprecation warning per `OnceKey` window while continuing later orders.
- Expansion/aspect requirement aggregation reports both the invoked root and the contributing source that raised the requirement.
