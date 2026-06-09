# Marta Hidalgo - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The host capability model is explicit: `[daemon] formula_v2` is edge-only, maps to `HostCapabilities`, and `CheckRequirements`/`AcceptCompileResult` run before durable writes across CLI, API, orders, fanout, and convergence.
- The diagnostics surface is typed and testable: shared `FormulaDiagnostic`, projection parity fixtures, Huma/OpenAPI/generated TypeScript gates, and registered event payloads are all part of the contract.
- Warning and failure rate limiting are treated as operator-facing behavior, with per-invocation CLI suppression, background grouping, order occurrence counts, and zero-write guarantees.

**Critical risks:**
- [Major] The diagnostic grouping examples conflict with the `OnceKey` contract. The projection rules define `<code>|<source_path>|<source_key>|<source_value>|<host_capability>`, and the disabled-host diagnostic points at `city.toml` / `daemon.formula_v2`; the fixture's `dashboard_group.key` uses `formulas/graph.toml|daemon.formula_v2|1`, omitting `source_value` and switching source path from host config to formula source. If implementers follow different examples, CLI/API/event/dashboard grouping will not compare byte-identical diagnostics, and repeated disabled-host order failures may split or collapse incorrectly.
- [Major] Accepted deprecation-warning event behavior remains ambiguous. The projection matrix says `formula.contract_deprecated` has a "suppressed warning event only for daemon/order paths", but the registered event contract names only compile-failure events. This leaves operators without a clear answer on whether accepted legacy formulas in background producers create no event, a grouped warning event, a persisted producer diagnostic, or only a metric/debug log.
- [Minor] The earlier diagnostic projection table says formula diagnostic CLI cases exit `1`, while the later CLI mapping says validation/report commands exit `2` and runtime launch commands exit `1`. The fixture agrees with the later mapping, but the table is still a normative-looking contradiction for `gc formula validate`, API-routed CLI, and generated help tests.
- [Minor] The design uses `config generation`, `scan generation`, and process-restart grouping in the order/controller rate-limit contract, but does not define the generation source or persistence boundary. This matters because a config reload is supposed to start a new group while a process restart must not erase persisted failure history.

**Missing evidence:**
- No fixture snippet proves the same disabled-host formula preserves identical canonical diagnostic fields through direct CLI, API-routed CLI, Huma preview, dashboard grouping, and `order.formula_compile_failed` after the `OnceKey` field order is settled.
- No event payload/schema example shows how accepted warning diagnostics are represented, or states explicitly that they are never emitted as Event Bus events.
- No concrete definition is provided for `config generation` or `scan generation`, such as a content hash, revision counter, controller epoch, or city config version.

**Required changes:**
- Make `OnceKey` canonical in one place and update all examples and fixtures to match it exactly, including disabled-host source attribution. If host diagnostics point at `city.toml:[daemon].formula_v2`, dashboard and event grouping keys should use that same source path/key/value, or the contract should explicitly define a separate grouping key.
- Decide the accepted-warning policy for `formula.contract_deprecated` in daemon/order/controller paths. Either remove warning events and say accepted warnings are returned, persisted, or metriced only, or add a typed warning event constant, payload registration, dedupe key, occurrence-count rules, and tests.
- Reconcile CLI exit-code prose so the projection matrix, surface parity table, fixtures, and generated help all say formula validation/report diagnostics exit `2`, while runtime launch commands exit `1`.
- Define `config generation` and `scan generation` as typed values in the event/producer state contract, including reset and persistence semantics.

**Questions:**
- Should `formula.contract_deprecated` warnings from formula-backed orders be visible in the dashboard when the order succeeds, or only in explicit validation/report surfaces?
- For disabled host capability, is the operator-facing source intended to be the formula requirement location, the city config gate location, or both with one primary grouping source?
- Where is the occurrence counter stored for repeated order failures so a controller restart does not make a long-running disabled-host failure look new?
