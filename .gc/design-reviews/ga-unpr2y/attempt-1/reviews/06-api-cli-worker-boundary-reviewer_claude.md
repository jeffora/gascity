# Sarah Chen - Claude

**Verdict:** approve-with-risks

Lane: API and CLI projection, worker-boundary routing, response compatibility,
layering. Reviews the current `DESIGN.md` (attempt-15 review-response revision)
with `REQUIREMENTS.md` and `internal/session/AGENTS.md`. All guard/handler claims
verified against the checkout; citations inline. The active global verdict is
`block` and only a non-mutating Slice 0 is authorized, so this disposition
governs whether the API/CLI/worker-boundary apparatus Slice 0 must build is sound.

**Top strengths:**
- The per-surface delegation matrix is appropriately conservative and is exactly
  the layering discipline this lane wants (DESIGN.md:284-293). Only read-only API
  query lookup is first adopter; CLI/`apiClient()`, mail, extmsg, assignee, nudge,
  attach, transcript, logs, and pool resume stay "characterization only" with an
  explicit "current precedence to preserve before delegation" per row. CLI parity
  is correctly scoped to "stdout, stderr, JSON shape, exit code, fallback order,
  and whether the local path or API path owns the session operation"
  (DESIGN.md:290).
- Wire-parity requirements are strong and grounded in infrastructure that already
  exists. The design requires no-delta tests for `writeResolveError`/
  `humaResolveError`, Huma typed responses, OpenAPI generation, generated-client
  compatibility, and dashboard/SSE parity (DESIGN.md:280-282,424-425). Verified
  present: `internal/api/openapi_sync_test.go`, `internal/api/genclient/genclient_test.go`,
  `writeResolveError` (`internal/api/handler_sessions.go:213`), `humaResolveError`
  (`internal/api/huma_handlers_sessions.go:24`). An implementer has real anchors.
- The eight-step resolver precedence (DESIGN.md:220-244) faithfully characterizes
  the current `resolveSessionTargetIDWithContext` behavior — configured
  named-session-before-live-alias, config-orphan rejection, path-alias-by-`Title`,
  allow-closed gating, `RepairEmptyType` quarantine, and 409/404/500 error
  projection. Every named symbol exists in `internal/api/session_resolution.go`,
  so Slice 1 is characterizing real behavior, not an idealized model.

**Critical risks:**
- [Major] **Red flag #1 is live and only partially guarded; the design
  over-relies on SESSION-LIFE-008 as if it were a comprehensive boundary.** API
  handlers re-derive lifecycle from raw `state` metadata today:
  `internal/api/handler_status.go:325,371,488`,
  `internal/api/session_resolution.go:414`, and
  `internal/api/huma_handlers_sessions_query.go:300` all read `Metadata["state"]`
  directly, and `ProjectLifecycle` is **not used anywhere in `internal/api`
  non-test code**. The guard that REQUIREMENTS.md SESSION-LIFE-008 leans on,
  `TestLifecycleUserFacingConsumersStayOnProjectionHelpers`
  (`internal/session/lifecycle_projection_test.go:929`), is a hand-maintained
  per-file string denylist — it only forbids specific strings in
  `handler_sessions.go`, `cmd_session.go`, and `doctor_session_model.go`, so it
  does not catch any of the reads above. `DESIGN.md` never mentions `handler_status`,
  raw state, `ProjectLifecycle`, or SESSION-LIFE-008 (grep-confirmed), and the
  backlog adds no slice to extract these reads. The design's own Goal names
  "callers ... know too much about session state ... and lifecycle metadata"
  (DESIGN.md:26-27), so these reads are in scope — but unscheduled and
  under-guarded.
- [Major] **The worker-boundary invariant is asserted for API but enforced only
  for CLI, and the only existing guard has no exception mechanism.** The design
  requires "Production API/CLI mutating lifecycle operations must route through
  `worker.Handle`" (DESIGN.md:73,405-406). The sole existing enforcement,
  `TestGCNonTestFilesStayOnWorkerBoundary`
  (`cmd/gc/worker_boundary_import_test.go`), scans `cmd/gc` **exclusively** (via
  `runtime.Caller`→`ReadDir` of its own dir) and is a flat forbidden-needle list
  with **no allowlist/exception-row mechanism**. `internal/api` manager
  construction and direct-create (`internal/api/session_resolution.go:328`
  `CreateAliasedNamedWithTransportAndMetadata`, `internal/api/session_manager.go`)
  are entirely unguarded by it — the documented "exceptions" are to a rule never
  enforced in that directory. So lane Q2 ("how does new code stay on
  `worker.Handle`?") has no enforcement answer on the API surface. The Slice 0
  `WORKER_BOUNDARY_EXCEPTIONS.yaml` + `TestWorkerBoundaryExceptionLedger` (both
  not-yet-existing) must therefore (a) extend enforcement to `internal/api` and
  (b) convert the flat denylist into a ledger-consuming allowlist — the design
  states neither.
- [Minor] **Wire/CLI parity is required "for touched scenario rows" but the
  backlog names zero concrete `SESSION-*` rows or routes.** The per-surface matrix
  is the right shape, but lane Q3's "what proves response shapes stay identical"
  is deferred entirely to per-slice preflight with no illustrative anchor — even
  for Slice 1's query endpoints whose precedence the design otherwise specifies in
  full. Consistent with the cross-persona finding that the backlog (DESIGN.md:656-679)
  names no call sites.

**Missing evidence:**
- No slice owns moving the API raw-`state` reads (`handler_status.go`,
  `session_resolution.go`, `huma_handlers_sessions_query.go`) onto
  `ProjectLifecycle`; the design is silent on `handler_status` classification.
- No statement that worker-boundary enforcement will extend to `internal/api`, or
  that the flat `cmd/gc` denylist becomes a ledger-consuming guard.
- No named scenario rows / route IDs per slice for parity proof; Slice 1's query
  endpoints are specified for precedence but not for parity-test coverage.

**Required changes:**
1. Acknowledge the live API raw-`state` reads
   (`internal/api/handler_status.go:325,371,488`,
   `internal/api/session_resolution.go:414`,
   `internal/api/huma_handlers_sessions_query.go:300`) and either schedule a slice
   that moves them onto `ProjectLifecycle` or record them as bounded, expiring
   exceptions. Require Slice 0 to upgrade SESSION-LIFE-008's enumerated denylist to
   a comprehensive scan for raw `Metadata["state"]`/`Metadata["sleep_reason"]`
   interpretation outside `internal/session` (or document why per-file enumeration
   is sufficient), so red flag #1 cannot silently regrow.
2. State that worker-boundary enforcement will cover `internal/api` (a guard that
   scans `internal/api`, not only `cmd/gc`) and that
   `TestGCNonTestFilesStayOnWorkerBoundary` will be converted to consume
   `WORKER_BOUNDARY_EXCEPTIONS.yaml` exact/expiring rows — otherwise the API
   worker-boundary rule is documentation without enforcement.
3. Give each API/CLI-touching slice concrete scenario rows + route IDs for parity
   proof, starting with Slice 1's query endpoints (get, pending, transcript,
   stream attach), not just the per-surface matrix.

**Questions:**
- Is the API worker-boundary rule meant to be CI-enforced via a new guard scanning
  `internal/api`, or documentation-only? Today only `cmd/gc` is guarded.
- Does Slice 1 (read-only target classification) also address the API raw-`state`
  lifecycle reads in `handler_status.go`, or are those a separate, unscheduled
  extraction?
- Will SESSION-LIFE-008's guard become a comprehensive scan, or remain a
  hand-maintained per-file denylist that new handlers can route around?
