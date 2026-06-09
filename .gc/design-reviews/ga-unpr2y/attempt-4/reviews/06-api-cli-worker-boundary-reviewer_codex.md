# Sarah Chen - Codex

**Verdict:** block

Reviewed `internal/session/DESIGN.md`, which matches `.gc/design-reviews/ga-unpr2y/attempt-4/design-before.md`, plus the API control-plane docs, Huma usage notes, scoped session instructions, and current API/CLI worker-boundary code.

**Top strengths:**
- The design now explicitly restates the control-plane rule: CLI and HTTP/SSE are projections over the session object model, and API-visible conflict data must use typed Huma outputs rather than raw JSON.
- The worker-boundary section correctly says session-owned commands do not give production `cmd/gc` code permission to bypass `worker.Handle`.
- Per-slice compatibility proof now includes API status/body shape, request IDs, async result event behavior, CLI stdout/stderr/JSON/exit codes, and generated schema/client impact.

**Critical risks:**
- [Blocker] The design does not add a failing-build guard for new API-side session-manager bypasses. The current worker-boundary guard in `cmd/gc/worker_boundary_import_test.go` scans only non-test files in `cmd/gc`; it does not cover `internal/api`. Existing exceptions are named in the design (`internal/api/session_manager.go` constructs `session.Manager`; `internal/api/session_resolution.go` calls `CreateAliasedNamedWithTransportAndMetadata` directly), but there is no rule that prevents a new API handler from adding another direct `session.Manager.Create*` path during the migration. The mutation guard covers raw store writes, not direct manager command bypasses, so this remains a hole in the API/worker boundary.
- [Major] The coexistence plan does not clearly distinguish user-facing CLI command files from controller/reconciler infrastructure files that happen to live under `cmd/gc`. The document says production `cmd/gc` lifecycle operations stay on `worker.Handle`, but it also plans for `cmd/gc/session_lifecycle_parallel.go` and `cmd/gc/session_beads.go` to call session commands directly as reconciler shells. Without an explicit category split and guard allowlist, either legitimate controller code will be blocked or user-facing CLI code will gain a path to import session commands directly.
- [Major] The API materialization end state is still too broad. `internal/api/session_resolution.go` currently performs target resolution, config/provider command construction, MCP metadata construction, identity locking, alias/session-name checks, direct manager creation, reassignment, and poking. The design says this will become "classifier plus worker/session command delegation", but it does not define the typed adapter boundary or the exact response/error parity tests for named materialization, configured-name conflict, rejected-by-config, transport validation, and create failure. That is where Huma response compatibility is most likely to drift.

**Missing evidence:**
- A guard that scans `internal/api` for direct `session.Manager.Create*`, `session.NewManager*`, and raw lifecycle/session command bypasses, with explicit temporary allowlist rows for current documented exceptions.
- A command-surface import rule for `cmd/gc`: which files may call session-owned commands as controller/reconciler infrastructure, and which user CLI files must go through `worker.Handle` or the API client fallback path.
- Endpoint-level parity cases for API named-session materialization and target resolution: status codes, problem details, request IDs, async events, generated OpenAPI/client changes, and CLI fallback behavior.
- Proof that new conflict/result types are represented as typed Huma structs and not folded into generic maps or untyped problem-detail extensions.

**Required changes:**
- Extend the boundary guard plan beyond store writes: forbid new direct `session.Manager.Create*` and `session.NewManager*` call sites outside `internal/session`, `internal/worker`, and explicit migration exceptions in `internal/api`.
- Split `cmd/gc` call sites into user CLI projection versus controller/reconciler infrastructure, then bind each category to an allowed boundary (`apiClient`, `worker.Handle`, or narrowly allowlisted session command).
- Specify the API adapter contract for named materialization and target classification, including exact legacy error/status mapping and Huma output types.
- Add acceptance criteria that API/schema/dashboard changes run `make dashboard-check` and that no-change slices state explicitly that OpenAPI/generated clients are unchanged.

**Questions:**
- Should `internal/api/session_resolution.go` materialization delegate to `worker.Handle` directly, or to a session command factory shared with worker?
- Which `cmd/gc` files are considered controller infrastructure exceptions rather than user-facing CLI projection files?
- Will the existing `cmd/gc/worker_boundary_import_test.go` be expanded, or should the new session-boundary guard own both API and CLI bypass detection?
