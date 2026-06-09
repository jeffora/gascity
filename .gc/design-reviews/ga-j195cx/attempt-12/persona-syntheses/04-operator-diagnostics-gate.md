# Marta Hidalgo

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- Blocker: Operator-facing diagnostic parity is not guaranteed across direct CLI, API-routed CLI, Huma endpoints, dashboard projections, order dispatch, controller paths, and convergence. Both reviews agree the structured `Diagnostic` fields are the right foundation, but the design still leaves enough projection choice that operators could see different messages, statuses, or remedies for the same disabled `formula_v2` condition.
- Blocker: Compile-failure event behavior is not a stable typed contract. Codex blocks on the design saying `formula.compile_failed` "or order failure event"; Claude similarly flags unbounded fatal events under poll-driven order dispatch. The design needs exact event constants, registered payload fields, envelope semantics, producer ownership, SSE/generated-type effects, compatibility for existing `order.failed` consumers, and repeat-emission policy.
- Major: HTTP and generated-client CLI classification for `formula.compiler_requirement_unsatisfied` is under-specified. Both reviews call out the 400-versus-409 split as too vague without endpoint/action mapping, response body fields, and dashboard contract tests. Codex additionally identifies the API-routed CLI path through `cmd/gc/apiroute.go:apiClient()` as a blocker because the same command can otherwise render canonical stderr without a supervisor and generic HTTP failure with one.
- Major: Warning and fatal diagnostic cadence is not pinned tightly enough for long-running controllers and poll-driven dispatch. The reviews agree `OnceKey` is useful, but "bounded LRU per city" lacks owner, capacity, eviction policy, concurrency expectations, tests across legitimate cardinality, and a policy for fatal repeated failures when `formula_v2 = false` keeps blocking a scheduled order.
- Major: Aggregated requirement provenance is insufficient for operator remediation. Codex requires diagnostics to show both the invoked root formula and the expansion/aspect or pack source that raised the requirement; Claude also notes the current state of controller and convergence error/event surfaces is not inventoried. Without this, operators can see "enable `[daemon] formula_v2`" but not which authored file, imported pack, or layer winner created the requirement.
- Major: Dashboard consistency depends on discipline rather than structural guards. Claude notes the migration table mandates typed dashboard projections but no CI gate prevents reading raw `gc.formula_contract` or other root metadata strings; Codex requires event/API generated types and parity tests so dashboard consumers do not infer disabled-capability state from strings or HTTP status alone.
- Minor: CLI warning wording is misleading. "Warning stderr once per invocation" should say "once per unique source/key per invocation" or equivalent, because a validate-all run across many legacy formulas can correctly emit many distinct warnings.

**Disagreements:**
- Claude verdict was `approve-with-risks`; Codex verdict was `block`. I choose `block` because this persona's gate is operator diagnostics. The unresolved API-routed CLI path, typed event choice, endpoint classification, and polling cadence leave high-risk ambiguity in exactly the surfaces operators depend on during incidents.
- Claude accepts fatal diagnostic emission once per failed dispatch/launch attempt if downstream deduplication is specified; Codex frames the event contract itself as the primary blocker. These are compatible concerns: first choose the typed event and payload contract, then define whether producers throttle repeated identical failures or consumers group by `once_key`.
- Codex focuses on API-routed CLI parity and event constants; Claude focuses more on LRU sizing, dashboard static guards, exit-code tables, and concrete dashboard cards. I treat the Codex-only items as required blockers and the Claude-only items as required hardening before approval, because all are part of the same cross-surface operator contract.
- Neither review requires Gemini input, and no Gemini artifact is present. That absence does not affect the synthesis because Claude and Codex were the required sources.

**Missing evidence:**
- A direct CLI versus API-routed CLI example showing identical diagnostic code, remediation, source fields, host capability, and exit class for disabled `formula_v2`.
- A table mapping every relevant Huma endpoint/action to HTTP status, response type, `Diagnostic` fields, and generated-client CLI handling.
- A chosen event projection for compile diagnostics, including event constants, payload structs, registration owner, envelope `Subject`/`Message`, SSE/OpenAPI/generated-type impact, and compatibility behavior for existing order-failure consumers.
- The warning-suppression owner/helper, capacity, eviction behavior, lifetime, concurrency expectations, and tests proving repeated legacy-contract warnings do not flood long-running controller/order scans.
- Fatal diagnostic repeat-emission policy under sustained poll-driven failure, especially whether producer backoff or dashboard/API grouping owns deduplication.
- Dashboard card/rendering contract for formula preview validation, disabled host capability, convergence preflight failure, and legacy-only roots with no canonical metadata.
- General `gc formula validate <name>` exit-code policy for fatal diagnostics and its relationship to `--legacy-contract-report`.
- Provenance shape for requirements raised by expansion/aspect formulas, including root formula, contributing source formula/path, pack or layer winner, and any compact reference to compile provenance artifacts.
- Inventory of existing controller, convergence, dashboard, and generated-client paths that currently emit typed diagnostics versus raw strings.

**Required changes:**
- Add a cross-surface projection matrix for formula diagnostics covering direct CLI, API-routed CLI, Huma endpoints, dashboard-consumed API responses, order dispatch, controller, convergence, and event/SSE streams. Include status/exit class, response/body fields, canonical rendering expectations, and test owners.
- Replace the ambiguous `formula.compile_failed` "or order failure event" language with an event contract table naming each event constant, payload fields, registration owner, envelope semantics, emitting surfaces, generated-type effects, and compatibility behavior.
- Define repeat-emission and deduplication policy for warning and fatal diagnostics. Pin the `OnceKey` owner, capacity, eviction behavior, reset behavior, and fatal failure cadence, with an order-dispatch test for disabled `formula_v2` across repeated polls.
- Require generated-client CLI parity tests that run the same disabled-capability launch with and without supervisor/API routing and assert the same diagnostic code, remediation, source fields, and exit class.
- Expand the endpoint/action matrix for `formula.compiler_requirement_unsatisfied`, including formula validate/preview, sling launch/attach, workflow launch, order-backed launch, convergence create/retry, and dashboard-consumed projections.
- Require diagnostics/events/API bodies for aggregated requirements to include invoked root formula, contributing source formula/path, pack or layer provenance, and a stable reference to full compile provenance when the compact fields are not enough.
- Add a CI/static guard preventing dashboard code from inferring disabled formula capability from raw root metadata such as `gc.formula_contract`; dashboard surfaces should consume typed diagnostics and generated fields.
- Document general `gc formula validate <name>` exit codes for fatal and warning diagnostics, and update the CLI warning wording to reflect one warning per unique source/key rather than one total warning per invocation.
