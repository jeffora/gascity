# Sarah Chen - API/CLI Worker Boundary Reviewer (Codex)

**Persona verdict:** approve-with-risks

Approve the design only for Slice 0 decomposition. Do not approve API/CLI
mutation movement, wake/drain movement, or public wire changes yet. The design
now has the right boundary language, but several items must become concrete
machine-checked rows before any behavior-moving slice starts.

## What is strong enough now

- The design respects the control-plane layering direction. The API control
  plane says the object model is central and the CLI/API are projections, not
  reimplementations (`engdocs/architecture/api-control-plane.md:17-30`). The
  design matches that: callers gather facts and render results, while
  `internal/session` owns only session rules (`internal/session/DESIGN.md:28-35`
  and `internal/session/DESIGN.md:460-465`).
- Read-only API target classification is separated from mutating lifecycle
  work. The design allows store-level classification only because it does not
  create, destroy, wake, drain, or repair sessions
  (`internal/session/DESIGN.md:402-406`). That is the right boundary for the
  first behavior slice after Slice 0.
- The design explicitly defers mutating API commands and CLI fallback paths
  until compatibility rows exist. The per-surface matrix requires route-specific
  Huma status, problem body, request ID behavior, generated-client shape,
  materialization, close/wake semantics, stdout/stderr/JSON shape, exit code,
  and `apiClient()` fallback parity (`internal/session/DESIGN.md:288-290`).
- The Worker/API/CLI section now encodes the root migration rule: production
  API/CLI lifecycle operations route through `internal/worker/handle.go` unless
  root docs and `WORKER_BOUNDARY_EXCEPTIONS.yaml` list an exact expiring
  exception (`internal/session/DESIGN.md:402-425`).
- The design acknowledges the typed-wire obligations that matter for this repo:
  Huma typed responses, OpenAPI generation, generated client compatibility,
  dashboard/SSE expectations, and CLI `apiClient()` fallback parity
  (`internal/session/DESIGN.md:423-425`). Those line up with the API control
  plane requirements that Huma registration generates the spec
  (`engdocs/architecture/api-control-plane.md:105-109`), spec files are
  generated (`engdocs/architecture/api-control-plane.md:189-195`), every
  compile-time-known shape is typed (`engdocs/architecture/api-control-plane.md:266-282`),
  and generated-client/spec drift is CI-covered
  (`engdocs/architecture/api-control-plane.md:550-561`).

## Required gates before behavior-moving slices

1. **Wake/drain cannot just say "use `worker.Handle`"; the handle shape is not
   there yet.** Current `worker.LifecycleHandle` exposes Start, Attach, Create,
   Reset, Stop, Kill, Close, CloseDetailed, Rename, and State
   (`internal/worker/handle.go:28-40`), but no explicit Wake or Drain command.
   Current CLI and API wake paths still call `session.WakeSession` and write
   metadata directly (`cmd/gc/cmd_session_wake.go:56-94`,
   `internal/api/handler_sessions.go:440-466`). Before the wake/drain slice is
   decomposed, it must choose one concrete path: add the canonical
   `worker.Handle` operation and parity tests, or add a root-approved expiring
   exception row with the exact route, reason, response shape, and retirement
   proof. The current design text is enough as a policy, not enough as an
   implementation contract.
2. **`API_CLI_ROUTE_INVENTORY.yaml` must enumerate exact surfaces, not classes
   of surfaces.** The API control-plane doc says registered routes are the
   exposed routes and shadow/path-rewrite mappings are bugs
   (`engdocs/architecture/api-control-plane.md:197-203`). Slice 0 therefore
   needs rows for each legacy mux handler, Huma operation ID, generated Go
   client method, dashboard or SSE consumer, Cobra command, and `apiClient()`
   fallback path that can reach a session. Rows should include the touched
   `SESSION-*` requirements and the exact parity tests. Without that, "API
   query-side lookup" and "CLI fallback paths" remain too broad to protect wire
   compatibility.
3. **Public API proof must include the real typed-wire gates, not just targeted
   Go unit tests.** Huma usage has sharp edges around operation IDs and
   middleware-vs-validation behavior (`engdocs/contributors/huma-usage.md:120-128`,
   `engdocs/contributors/huma-usage.md:230-247`), and SSE routes must use the
   local wrapper when precheck errors are possible
   (`engdocs/contributors/huma-usage.md:287-302`). Any slice touching
   `internal/api/`, OpenAPI, generated clients, dashboard types, or SSE must
   require the repo gates from root instructions: `make dashboard-check` when
   applicable, `TestOpenAPISpecInSync`, `TestGeneratedClientInSync`, and
   route-level response validation. The current design names these concerns,
   but the Slice 0 artifacts should turn them into commands.
4. **The current `cmd/gc` worker-boundary guard is necessary but not sufficient
   for the new mutation boundary.** `TestGCNonTestFilesStayOnWorkerBoundary`
   catches direct sessionlog, `session.NewManager`, `worker.SessionHandle`, and
   similar bypasses in non-test `cmd/gc` files
   (`cmd/gc/worker_boundary_import_test.go:21-48`). It does not catch direct
   `session.WakeSession`, session-owned `SetMetadata*`, or generic bead bridge
   writes. The design's additional mutation ledger and static guard must cover
   those paths before wake, close, wait, pin, nudge, bd-store bridge, or generic
   bead update code is touched.
5. **CLI fallback parity must be tested as a three-way contract.** The API
   control-plane doc says local CLI calls core directly when no supervisor is
   running, and routes mutations through HTTP/generated client when a local
   supervisor is running (`engdocs/architecture/api-control-plane.md:84-103`).
   For session-affecting commands, the proof rows need to compare local fallback
   behavior, API-routed behavior, and direct API behavior for stdout, stderr,
   JSON, exit status, fallback reason logging, request/result events, and
   session state. A pure handler-level test is not enough.

## Evidence checked

- `go test ./cmd/gc -run TestGCNonTestFilesStayOnWorkerBoundary -count=1`
  passed.
- `go test ./internal/api -run TestOpenAPISpecInSync -count=1` passed.
- `go test ./internal/api/genclient -run TestGeneratedClientInSync -count=1`
  passed.
- Static scan confirmed current wake paths still use direct session mutation
  APIs; this is acceptable as legacy behavior only because the design keeps
  wake/drain characterization-only until a worker-boundary row exists.

## Required changes before API/CLI implementation beads

- Add exact `API_CLI_ROUTE_INVENTORY.yaml` rows for the first adopter before
  Slice 1 moves code, including Huma operation IDs, legacy routes, generated
  client methods, dashboard/SSE consumers when present, CLI command/fallback
  paths, and proof selectors.
- Add exact `WORKER_BOUNDARY_EXCEPTIONS.yaml` rows for any production API/CLI
  session lifecycle path that cannot use `worker.Handle`, with owner, expiry,
  response-shape proof, OpenAPI/generated-client/dashboard proof, and retirement
  criteria.
- For wake/drain, either extend `worker.Handle` with the needed operation or
  explicitly approve a temporary exception before implementation starts.
- Require public-surface proof commands in the relevant slice contract:
  OpenAPI sync, generated client sync, dashboard check when generated dashboard
  types or API schema move, and CLI local/API fallback parity tests.

## Bottom line

The design is acceptable for non-mutating Slice 0. It should remain blocked for
API/CLI mutating lifecycle work until the route inventory, worker-boundary
exception ledger, and typed-wire proof commands exist as concrete artifacts.
