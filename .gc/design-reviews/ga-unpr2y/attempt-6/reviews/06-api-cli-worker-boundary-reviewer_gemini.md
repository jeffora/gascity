# Jaxon Vance — DeepSeek V4 Flash (Independent Review, Attempt 6)

**Verdict:** pass

**Persona:** API and CLI projection, worker-boundary routing, response compatibility, and layering.

**Reviewed against:** `internal/session/DESIGN.md` (Attempt 6, matching `.gc/design-reviews/ga-unpr2y/attempt-6/design-before.md` with Attempt 6 changes), `/data/projects/gascity/internal/api/`, `/data/projects/gascity/cmd/gc/`, `/data/projects/gascity/internal/worker/`, and `REQUIREMENTS.md` scenarios.

---

## Top Strengths

- **De-conflated API/CLI Boundary & Core Decider Logic**: 
  The designer has added the **"Caller Routing And Command Construction Gate"** (`DESIGN.md:264–281`), mapping out the CLI, legacy API, Huma API, reconciler, and mail surfaces to their required end-state routes and static guard rules. This clearly articulates how the API/CLI projections sit around the core deciders and prevents bypass patterns.
  
- **Comprehensive Symbol-Level Mutation Inventory (W-022 through W-028)**: 
  The expanded mutation inventory explicitly covers previously untracked symbol-level mutators and API-side bypasses:
  - `W-023`: Tracks `session.WakeSession` wait updates and wake batches.
  - `W-024`: Tracks `RepairEmptyType` internal repair exceptions.
  - `W-025`: Tracks direct API commands in `huma_handlers_sessions_command.go`.
  - `W-027` & `W-028`: Capture external calls to `Manager` lifecycle methods and package-level mutators.
  This completely resolves the previous blocker regarding missing call-site visibility.

- **Robust Static Guard Specification Upgrade**: 
  The static guard description has been upgraded to scan not only raw store writes but also direct package-level function calls (`WakeSession`, `RepairEmptyType`), manager method calls (`CloseDetailed`, `Suspend`, `UpdatePresentation`), and dynamic metadata patch-map extensions (`DESIGN.md:551-554`). It also mandates positive test fixtures to prove the AST/symbol-level scanner fails before allowlist rows are added.

---

## Remaining Risks & Cleanup (Recommendations)

### 1. Folding Ad-Hoc Key Clearing into the Patch Builder (`session_resolution.go:171-173`)
In the active API layer, the retirement of Named Sessions is handled by:
```go
patch := session.RetireNamedSessionPatch(now, "continuity-ineligible-replacement", spec.Identity)
patch["alias_history"] = "" // Ad-hoc key addition
```
While row `W-028` of the inventory requires static guard coverage for patch-map extension, we strongly recommend that during the implementation of Slice 4, the clearing of `alias_history` be formally folded into the `session.RetireNamedSessionPatch` helper inside `internal/session`. This aligns with the core architectural principle that the session domain owns its key-family taxonomy, rather than exposing mutable maps for callers to arbitrarily extend.

### 2. Standardizing `handler_status.go` and Query Handlers on Projections
The current `handler_status.go:488` maps `"drained"` to `StateAsleep` via `statusSessionState`, which bypasses `ProjectLifecycle`. 
While `W-026` and `W-027` provide a bridge for status and target helpers during migration, the implementation team should ensure that `handler_status.go` and `huma_handlers_sessions_query.go` eventually migrate to the standard `ProjectLifecycle` projection. Furthermore, `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` in `internal/session/lifecycle_projection_test.go` should be expanded to cover `handler_status.go` read-sides once converted.

### 3. Verification of Inline Huma Close Policy Check
The close handler in `huma_handlers_sessions_command.go:822-826` performs inline metadata pre-checks to reject closing config-managed named sessions. During Slice 4's implementation, the Close command should absorb this named-session check entirely, so that both CLI and API surfaces automatically share the same rejection rules, preventing error-response drift.

---

## Answers to Persona Questions

1. **How do we ensure that user-facing CLI and API responses remain compatible during and after the extraction?**
   - The design guarantees response compatibility by preserving the existing, surface-specific resolver chains as separate oracles (`DESIGN.md:614-626`). It avoids imposing a single global precedence and utilizes Golden/Characterization tests committed before migration as a strict exit gate.

2. **Does the worker boundary or CLI bypasses allow direct store writes?**
   - No. The strict static guard rejects any direct writes to `beads.Store` from outside `internal/session` unless covered by a shrink-only allowlist row with an owner slice and expiry.

3. **Are external package-level calls and manager methods safely tracked?**
   - Yes, the updated AST guard covers external calls to `Manager` lifecycle/presentation methods (`CloseDetailed`, `Suspend`, `UpdatePresentation`) and package-level mutators (`WakeSession`, `RepairEmptyType`).

---

## Questions for the Author

1. Will the AST static guard be run as a pre-commit hook or part of the `make dashboard-check` pipeline to ensure no developer introduces raw `WakeSession` or `SetMetadataBatch` bypasses?
2. During Slice 4 implementation, can we commit to moving the named-session close-rejection policy from `huma_handlers_sessions_command.go` into the core session close command?
