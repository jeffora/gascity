# Sarah Chen

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Major] The API/CLI boundary model is correct in shape: read-only API target classification can be the first adopter, while mutating API commands, CLI fallback paths, mail, extmsg, assignee, nudge, attach, transcript/logs, and pool resume remain characterization-only until each surface has explicit rows.
- [Major] Worker-boundary enforcement must become machine-checked for API as well as CLI. The existing `cmd/gc` guard scans only CLI files and has no ledger-consuming exception mechanism; `internal/api` manager construction, direct create paths, wake paths, and lifecycle mutations need exact expiring exceptions or `worker.Handle` routing before mutating slices move.
- [Major] Current API raw lifecycle-state reads are in scope and under-scheduled. Direct `Metadata["state"]` and `Metadata["sleep_reason"]` interpretation in API status, query, and resolution paths bypasses `ProjectLifecycle`; these reads need a cleanup slice or bounded exceptions and stronger SESSION-LIFE-008 guard coverage.
- [Major] `API_CLI_ROUTE_INVENTORY.yaml` must enumerate exact surfaces, not broad classes. Rows need legacy mux handlers, Huma operation IDs, generated-client methods, dashboard/SSE consumers, Cobra commands, `apiClient()` fallback paths, touched `SESSION-*` rows, and proof selectors.
- [Major] Public wire parity must be part of the slice contract. API-touching slices need Huma response validation, `writeResolveError`/`humaResolveError` no-delta tests, OpenAPI sync, generated-client sync, dashboard checks where applicable, SSE wrapper obligations, and route-level response proof.
- [Major] Wake and drain cannot move with policy text alone. `worker.Handle` currently lacks explicit Wake/Drain operations, so wake/drain implementation slices must either add canonical worker-boundary operations or carry root-approved expiring exceptions with response-shape and retirement proof.
- [Major] CLI session commands need three-way parity: local fallback behavior, API-routed behavior through `apiClient()`, and direct API behavior must match for stdout, stderr, JSON shape, exit status, fallback reason logging, request/result events, and session state.
- [Minor] Slice 0 artifact volume is a DX risk. The inventory must remain complete, but overlapping facts should be consolidated or generated so boundary metadata does not drift across many manually synchronized files.

**Disagreements:**
- Claude and Codex return `approve-with-risks`; DeepSeek returns `block`. I choose `approve-with-risks` because the approved work remains non-mutating Slice 0 only, and all three reviews agree API/CLI mutation movement stays blocked until route inventory, worker-boundary exceptions, and typed-wire proof exist.
- DeepSeek treats Slice 0 artifact sprawl as a blocker and proposes reducing the preflight to a few consolidated files. Claude and Codex focus on completeness and enforcement. My assessment: consolidation is required where practical, but it must not weaken the closed-world inventory and exception-ledger requirements.
- Codex emphasizes that `worker.Handle` lacks Wake/Drain operations; Claude and DeepSeek emphasize that API enforcement does not exist today. These are compatible: before wake/drain slices move, the design must either add worker operations or write exact temporary exceptions, and the guard must cover API paths.
- Claude and DeepSeek flag raw API state reads as a boundary leak; Codex does not elevate that issue. I assess it as required for this lane because API projection should not keep independent lifecycle interpretation while the session model is being centralized.

**Missing evidence:**
- A worker-boundary guard plan that scans `internal/api` and `cmd/gc`, consumes `WORKER_BOUNDARY_EXCEPTIONS.yaml`, and fails unapproved use of lifecycle manager calls, `session.WakeSession`, direct lifecycle metadata writes, exported repair helpers, patch constructors, and other mutating bypasses.
- Exact `WORKER_BOUNDARY_EXCEPTIONS.yaml` rows for current API manager construction, API session-resolution direct-create paths, wake paths, drain paths, CLI local fallback, repair/doctor/migration code, and any temporary direct lifecycle calls.
- A map for API raw-state reads in `internal/api/handler_status.go`, `internal/api/session_resolution.go`, and `internal/api/huma_handlers_sessions_query.go`, including whether each is migrated to `ProjectLifecycle` or treated as an expiring exception.
- Exact first-adopter endpoint and route inventory for read-only target classification, including legacy and Huma routes, operation IDs, generated-client-visible methods, allow-closed/live-only behavior, and `materialize:true` sibling behavior.
- Public surface proof commands for each affected slice: OpenAPI sync, generated client sync, dashboard check when schema/types move, SSE frame validation, Huma response validation, and CLI local/API fallback parity.
- Concrete scenario row and route ID mapping for API/CLI-touching slices, starting with query endpoints such as get, pending, transcript, and stream attach.
- A consolidation or generation plan for overlapping Slice 0 manifests so route, symbol, boundary, exception, and proof metadata have clear ownership.

**Required changes:**
- State that production API/CLI mutating lifecycle operations must route through `worker.Handle` unless a root-approved exact expiring exception row permits a temporary bypass with owner, expiry, operation, allowed helper, parity proof, and retirement criteria.
- Extend worker-boundary CI enforcement to `internal/api` as well as `cmd/gc`, and convert the flat CLI denylist into a ledger-consuming guard that honors only exact expiring exception rows.
- Schedule the API raw-state reads for migration onto `ProjectLifecycle`, or record them as bounded exceptions; upgrade SESSION-LIFE-008 guard coverage beyond a hand-maintained per-file denylist.
- Add exact `API_CLI_ROUTE_INVENTORY.yaml` rows before Slice 1 moves code, including Huma operation IDs, legacy routes, generated-client methods, dashboard/SSE consumers, CLI command/fallback paths, touched `SESSION-*` rows, and proof selectors.
- Before wake/drain implementation slices, either add canonical `worker.Handle` operations with parity tests or add temporary exception rows that include exact route/command, response shape, OpenAPI/generated-client/dashboard proof, and retirement proof.
- Require public-surface proof commands in relevant slice contracts: `TestOpenAPISpecInSync`, `TestGeneratedClientInSync`, `make dashboard-check` when applicable, Huma route response validation, SSE wrapper tests, and CLI local/API fallback parity tests.
- Define CLI parity as a three-way contract between local fallback, API-routed CLI, and direct API behavior for stdout, stderr, JSON, exit status, fallback logging, request/result events, and session state.
- Consolidate or generate overlapping Slice 0 artifacts where possible while preserving closed-world route, boundary, and exception coverage.
