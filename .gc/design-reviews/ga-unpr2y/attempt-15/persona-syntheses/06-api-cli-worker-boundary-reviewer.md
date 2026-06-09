# Sarah Chen

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Wake and drain cannot remain ordinary store-level production lifecycle operations by local slice decision. The design must require `worker.Handle` routing for production CLI/API mutating session operations, or require root-approved, exact, expiring exceptions with parity and retirement proof.
- [Blocker] API close routing is inconsistent across legacy and Huma surfaces. The legacy route uses `worker.Handle.CloseDetailed`, while the typed Huma route calls `session.Manager.CloseDetailed` directly. The design must either unify Huma close through the worker boundary or prove temporary exception parity for worker-operation events, cleanup behavior, response shape, OpenAPI/generated-client behavior, and dashboard/API compatibility.
- [Blocker] Error-to-wire compatibility for target classification is not pinned. The first read-only classifier adopter can regress existing `errors.Is` sentinel mappings for not-found, ambiguous, configured conflict, and default 500 cases unless `writeResolveError` and `humaResolveError` parity tests are mandatory.
- [Major] The first-adopter extraction source and materializing sibling behavior are under-specified. The design should name `resolveSessionTargetIDWithContext`, identify the exact read-only endpoints, and require parity proof for both read-only and `materialize:true` callers before carving out a side-effect-free classifier.
- [Major] The route inventory needs completeness and compatibility fields, not just freshness. A new session-reaching Huma, mux, or Cobra route with no row must fail CI, and each row must record HTTP status/problem shape, generated-client method, OpenAPI/dashboard impact, CLI stdout/stderr/JSON/exit-code contract, fallback classification, touched session rows, and proof selectors.
- [Major] `WORKER_BOUNDARY_EXCEPTIONS.yaml` is too narrow for real bypasses. It must track exported lifecycle helpers such as wake, exported repair helpers, patch constructors, direct lifecycle metadata writes, Huma manager lifecycle calls, runtime drain paths, CLI local fallback, repair/doctor/migration paths, and tests.
- [Major] Caller-specific resolver overlays are not fully represented as requirements or integration fixtures. Config-named reopening, template target rejection, API-side path-alias matching, closed lookup allowances, and other surface-specific precedence rules need explicit rows or mandatory multi-surface parity tests before delegation.
- [Major] Slice 0 artifact sprawl risks stale duplicated metadata. The preflight inventory should be consolidated or generated enough that developers do not have to manually synchronize overlapping boundary facts across many files.

**Disagreements:**
- Claude rated the design `approve-with-risks`; Codex and DeepSeek rated it `block`. Claude's approval rests on Slice 0 being non-mutating and the first adopter being narrowly scoped. I assess the block verdict as controlling because later mutating slices are already allowed to choose store-level wake/drain, and current API close paths already diverge at the boundary.
- DeepSeek treats Slice 0 document creep as a blocker, while Claude and Codex focus more on missing compatibility contracts and boundary exceptions. I assess artifact sprawl as a major required-change item, not the primary blocker, because consolidation can be handled without weakening the inventory requirement.
- Claude sees the worker-boundary posture as broadly accurate, with wake routing deferred as a minor load-bearing decision. Codex and DeepSeek object that deferral itself can bless bypasses. I agree with Codex and DeepSeek: worker-boundary routing for mutating production operations should be decided in the design, not left as a per-slice option.

**Missing evidence:**
- A worker-boundary guard plan that fails on `session.WakeSession`, exported repair helpers, patch constructors, direct lifecycle `SetMetadataBatch`, Huma manager lifecycle calls, and other non-worker mutating paths unless they have exact expiring exceptions.
- Route inventory schema with concrete compatibility fields for Huma responses, legacy mux responses, generated-client methods, OpenAPI/dashboard obligations, CLI output, exit codes, JSON shape, stderr/stdout, fallback behavior, and touched session rows.
- Tests proving Huma typed routes, legacy mux routes, and CLI `apiClient()` fallback paths preserve lifecycle outcomes for close, wake, drain, suspend, nudge, attach, transcript, mail, and generic bead routes that can reach sessions.
- Explicit first-adopter source and endpoint list, including `resolveSessionTargetIDWithContext`, the read-only endpoint(s), and all `materialize:true` callers that must remain behavior-compatible.
- Wire no-delta gates for the read-only adopter, including `TestOpenAPISpecInSync` and `make dashboard-check` where generated API/dashboard types can be affected.
- Requirements rows or mandatory integration fixtures for caller-specific resolver overlays: config reopening, template target rejection, active path-alias fallback, qualified-basename behavior, and closed lookup allowances.
- A plan for preventing mutating CLI commands from bypassing local worker-boundary logic through direct generated-client API calls.

**Required changes:**
- State that production CLI/API mutating session lifecycle operations must route through `worker.Handle` unless a root-approved, exact, expiring exception row permits a temporary bypass with owner, expiry, operation, allowed helper, retirement proof, and same-change update rule.
- Add API close to the boundary plan. Route the Huma typed close handler through `worker.Handle.CloseDetailed`, or add a temporary exception with tests for operation events, wait-nudge cleanup, response shape, OpenAPI/generated-client behavior, and dashboard compatibility.
- Require an error-translation parity suite for `writeResolveError` and `humaResolveError` so every classifier result preserves existing HTTP status and typed response behavior, including default 500 handling.
- Name `resolveSessionTargetIDWithContext` as the extraction source and require route inventory rows plus parity proof for read-only and materializing callers before Slice 1 adopts the classifier.
- Make `API_CLI_ROUTE_INVENTORY.yaml` completeness-enforced and extend each row with concrete API, CLI, generated-client, fallback, dashboard/OpenAPI, touched-session-row, and proof-selector fields.
- Expand `WORKER_BOUNDARY_EXCEPTIONS.yaml` to cover the actual bypass classes: exported lifecycle helpers, exported repair helpers, patch constructors, direct lifecycle metadata writes, Huma manager lifecycle methods, runtime drain, CLI fallback, repair/doctor/migration paths, and tests.
- Consolidate or generate overlapping Slice 0 manifests so the boundary facts have one authoritative source and stale duplicated metadata does not become a routine CI failure source.
