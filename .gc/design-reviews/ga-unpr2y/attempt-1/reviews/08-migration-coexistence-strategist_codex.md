# Ravi Krishnamurthy - Codex

**Verdict:** block

**Top strengths:**
- The design correctly narrows the only immediately executable work to non-mutating Slice 0, with `SESSION_BOUNDARY_SYMBOLS.yaml`, `API_CLI_ROUTE_INVENTORY.yaml`, `WORKER_BOUNDARY_EXCEPTIONS.yaml`, and shrink-only guards before later behavior moves.
- The mutation ownership ledger is pointed in the right direction: exact path/function, key family, dynamic-key source, current owner, intended owner slice, exception expiry, persistence-error handling, and generic API/CLI bridge coverage.
- The migration section has the right core rule: no legacy raw writer and new session-owned command may coexist for the same session-owned fields unless the row names a fence and proves it with raced-writer tests.

**Critical risks:**
- [Blocker] The worker/API close inventory is source-inaccurate. The design says close remains aligned with `worker.Handle.CloseDetailed` for production API and CLI callers, including Huma routes, but `internal/api/huma_handlers_sessions_command.go` still constructs `s.sessionManager(store)` and calls `mgr.CloseDetailed(id)`. That may be an allowed current exception, but the design only names broad API manager construction and API session-resolution direct-create exceptions. A migration plan cannot safely decompose close/identity work until this direct Huma close path is either routed through `worker.Handle` or listed as an exact expiring exception with route, response, cleanup, event, and retirement proof.
- [Blocker] The migration row trigger is too narrow. The design says worker-boundary and session-mutation-boundary migrations overlap in `internal/api/session_resolution.go`, `cmd/gc/session_reconciler.go`, and `cmd/gc/session_beads.go`, but source scan shows additional mutation-bearing or boundary-sensitive paths in Huma command handlers, legacy API handlers, session create rollback, wake, nudge, session lifecycle parallel start, sleep, wait, circuit breaker, soft reload, and multiple reconciler helpers. The broader mutation ledger may eventually inventory them, but the migration section currently risks creating rows only for three named files.
- [Major] Ownership transfer is specified at field-family level in the migration section, while the mutation ledger requires exact keys and dynamic-key sources. Field families such as wake/hold/drain, create/start, and runtime identity are too broad for coexistence decisions; two writers can collide on `wake_request`, `pending_create_*`, `session_key`, or close status while appearing to be in the same family. Migration rows need exact key-level ownership before/during/after the slice.
- [Major] Rollback data direction is named but not gated. The design says rows must state whether rollback preserves new fields, clears them, or runs repair/backfill, but it does not require proof that the slice can ship and revert independently without the successor slice. That leaves room for hidden flag-day dependencies where old readers do not actually tolerate new metadata, or new readers require a field the previous slice has not guaranteed.

**Missing evidence:**
- A materialized `WORKER_BOUNDARY_EXCEPTIONS.yaml` listing exact API/Huma manager-use exceptions, not just the helper that constructs a manager.
- A source-complete `API_CLI_ROUTE_INVENTORY.yaml` that maps every session-affecting Huma route, legacy mux route, Cobra command, `apiClient()` fallback, generic bead route, and dashboard/generated-client obligation to its current boundary owner.
- A `SESSION_BOUNDARY_SYMBOLS.yaml` row set proving every raw writer and helper-returned patch map is owned, including Huma close/update/permission mode/suspend paths, legacy wake, session create rollback, direct named-session materialization, wait metadata, and reconciler repair writes.
- Migration rows for the existing Huma close direct-manager path, API materializing named-session create path, explicit wake paths, close/delete cleanup, and reconciler wake/drain/repair writers.
- Tests proving old readers tolerate new fields and new readers tolerate old fields during rollback for each migrated key, plus raced old-writer/new-command tests where coexistence is temporarily allowed.

**Required changes:**
- Replace the three-file migration trigger with an inventory-driven trigger: any slice that touches a path or symbol in `SESSION_BOUNDARY_SYMBOLS.yaml`, `API_CLI_ROUTE_INVENTORY.yaml`, or `WORKER_BOUNDARY_EXCEPTIONS.yaml` with both worker-boundary and session-mutation relevance must add a migration row.
- Add exact exception rows for current API/Huma manager-use mutation paths, especially Huma close. Each row should name route, handler function, manager method, session-owned fields/top-level status touched, response compatibility proof, event/cleanup behavior, owner, expiry, and retirement condition.
- Require migration rows to transfer ownership at exact key or dynamic-key-source granularity, not only field family. The row may group keys for readability, but the validator must fail if a touched key lacks before/during/after owner, old writer, new command owner, fence, and rollback behavior.
- Strengthen rollback gates: every behavior-moving slice must prove it can ship alone and revert alone. The proof should include old-reader/new-field tolerance, new-reader/old-field tolerance, no required successor slice, and a repair/backfill path when rollback leaves durable metadata behind.
- Make worker-boundary and session-mutation-boundary close gates cross-check each other. A route cannot retire a raw writer, add a command call, or keep a manager exception unless both the worker-boundary exception ledger and session mutation ledger agree on the same owner, expiry, and tests.

**Questions:**
- Is the current Huma `mgr.CloseDetailed` path intended to be an active worker-boundary exception, or is it expected to route through `worker.Handle` before close/identity migration work starts?
- Should `internal/api/session_manager.go` be treated as one broad exception, or should every handler method that calls the returned manager have its own exception row?
- Which slice owns the API materializing named-session path in `internal/api/session_resolution.go`: target classification, worker-boundary cleanup, close/identity retirement, or a separate named-session materialization slice?
