# Ravi Krishnamurthy — DeepSeek V4 Flash (Independent Review, Attempt 5)

**Verdict:** block

**Persona:** migration sequencing, legacy-new coexistence, rollback slices, worker-boundary collision, cross-document consistency.

**Reviewed against:** `internal/session/DESIGN.md` (attempt 5, `.gc/design-reviews/ga-unpr2y/attempt-5/design-before.md`), `internal/session/REQUIREMENTS.md`, `cmd/gc/cmd_session_wake.go`, `internal/session/resolve.go`, `internal/session/manager.go`, `internal/api/session_resolution.go`, and attempt-5 cross-persona reviews (Elena Marchetti, Natasha Volkov, Sarah Chen).

---

## Top Strengths

- **Explicit Per-Slice Coexistence Gates (`DESIGN.md:760–771`):** The inclusion of a dedicated coexistence table with "Converted callers", "Legacy callers allowed", "Validation differences", "Guard update", and "Bake and revert rule" columns is a highly professional addition. This provides a clear path for progressive adoption and helps prevent hidden flag days.
- **Detailed Mutation Inventory (`DESIGN.md:420–450`):** Granular tracking of all canonical writers and legacy exception paths (W-001 through W-021) with assigned owner slices is excellent. Defining shrink-only allowlist rules for the static guard ensures that legacy bypasses can only be retired, never added.
- **Scenario to Slice Mapping (`DESIGN.md:92–119`):** Explicitly linking each requirements scenario (e.g., `SESSION-LIFE-001` through `SESSION-LIFE-008`) to a specific extraction slice and naming the proof gate creates a verifiable traceability thread from requirements down to code.

---

## Critical Risks (Blockers)

### [Blocker] `cmd_session_wake.go` fallback is an unguarded direct-store write that bypasses both the session and worker boundaries

Verified in the source code at `cmd/gc/cmd_session_wake.go:82–94`:
```go
if !hasRunnableTemplate && sessionWakeRequestedCreate(b) {
	if err := store.SetMetadataBatch(id, map[string]string{
		"state":                     string(session.StateAsleep),
		"state_reason":              "",
		"pending_create_claim":      "",
		"pending_create_started_at": "",
		"wake_request":              "",
		"wake_requested_at":         "",
	}); err != nil {
		fmt.Fprintf(stderr, "gc session wake: updating metadata: %v\n", err) //nolint:errcheck
		return 1
	}
}
```
During slice 2 (wake) coexistence, this fallback path writes directly to the bead store. This write:
1. Does not route through `worker.Handle` or `session.Manager`.
2. Bypasses the `withSessionMutationLock` in-process mutex.
3. Mutates keys belonging to **three separate families** in a single call: lifecycle (`state`, `state_reason`), create/start lease (`pending_create_claim`, `pending_create_started_at`), and wake/hold/drain (`wake_request`, `wake_requested_at`).

This creates a high-severity split-brain hazard: if slice 2 has been adopted but the reconciler or another CLI command is concurrently executing, a legacy writer can interleave with the new wake command's validation, resulting in lost metadata writes. The AST-based static guard will fail to prevent this because it is listed as a legacy exception (W-004), but the design does not specify how this fallback write is retired or how it is validated during slice 2 execution.

**Required change:** Document a concrete coexistence and retirement plan for `cmd_session_wake.go:82–94` in the `DESIGN.md` writer inventory. The wake command must absorb this fallback behavior or route it through the command boundary so that no direct `SetMetadataBatch` calls remain in CLI production files.

---

### [Blocker] Cross-family writes are extensive, and the per-family ownership model lacks a unified close/retirement policy

The existing `CloseDetailed` method (`internal/session/manager.go:862–921`) writes metadata keys from four distinct families inside a single lock span:
- **Lifecycle state:** `state`, `close_reason`, `closed_at`, `synced_at`
- **Create/start lease:** `pending_create_claim`, `pending_create_started_at`
- **Wake/hold/drain:** `pin_awake`, `held_until`, `sleep_intent`
- **Identity:** `alias`, `session_name`, `session_name_explicit`, `alias_history`

Under the proposed design, these families are extracted in different slices (Slice 1 = identity, Slice 3 = runtime-start, Slice 4 = close/retire, Slice 5 = wake/hold/drain). 
Because `CloseDetailed` is a single atomic database operation, we face a major sequencing hazard:
- If Slice 4 (close/retire) is implemented before Slice 5 (wake/hold/drain), the close command must temporarily own and write the wake/hold/drain keys (`pin_awake`, etc.) directly, violating the key-family boundaries.
- If the close command instead calls individual slice commands to clear their respective keys, we introduce a complex cross-slice command call graph that `DESIGN.md` does not define.

Currently, `CloseDetailed` is **not mentioned** in `DESIGN.md`. 

**Required change:** Define a cross-family write policy in `DESIGN.md` that explicitly covers `CloseDetailed`, `Suspend`, and `UpdatePresentation`. Document which families each complex operation writes, whether it delegates to other slice commands during coexistence, and the exact handoff conditions when those downstream slices are extracted.

---

### [Blocker] The API handler routing matrix is missing, leading to dual-surface split-brain writes during partial adoption

The design's shared call-site plan (`DESIGN.md:772–781`) lists `internal/api/session_resolution.go` and `session_manager.go` but fails to address the **two independent API surfaces** serving the same session endpoints:
1. **Legacy handlers (`internal/api/handler_sessions.go`):** Route via `s.workerHandleForSession(store, id)` and call `handle.Stop/CloseDetailed`.
2. **Huma handlers (`internal/api/huma_handlers_sessions_command.go`):** Route via `s.sessionManager(store)` and call `mgr.CloseDetailed/StopTurn/Kill/Suspend`.

These two surfaces use different internal routing. When Slice 3 or Slice 4 is partially adopted, having some API endpoints routing through the legacy worker boundary while others route through the new session command boundary creates a critical concurrency and validation risk. If one caller suspends via the legacy manager while another closes via the new command boundary, they will validate against stale state snapshots.

**Required change:** Add a dedicated API routing matrix to `DESIGN.md` showing exactly how each API lifecycle endpoint maps to its legacy and target command implementations. Specify which boundary (`worker.Handle` or a command factory) is the single source of truth for each slice during the bake period.

---

### [Blocker] `session.WakeSession` is a package-level function called directly by API handlers, bypassing boundary controls

The design assumes that all session mutations route through `session.Manager` or `worker.Handle`. However, `session.WakeSession` is a package-level function in `internal/session/waits.go` that is called directly by both legacy and Huma API wake handlers:
```go
// internal/api/huma_handlers_sessions_command.go:877
nudgeIDs, err := session.WakeSession(store, b, time.Now().UTC())
```
Because the bypass occurs at the function-call level rather than a direct store write, AST/symbol-based static guards checking for `store.SetMetadata` or `store.Update` will **fail** to detect that these API handlers are bypassing the command boundary. This means API wake requests will continue to use the legacy, uncoordinated wake function during and after Slice 2's rollout.

**Required change:** Modify the Slice 2 coexistence gate to specify how the API wake path is converted. `session.WakeSession` must be deprecated, and the API handlers must route wake requests through `worker.Handle` or the new explicit wake command.

---

## Major Risks

### [Major] `RepairEmptyType` is a mutating write on the read path, violating TR-002 and running with zero lock coordination

The exact direct-ID resolver `ResolveSessionBeadByExactID` (`internal/session/resolve.go:50–63`) calls `RepairEmptyType(store, &b)` which mutates the bead's top-level `Type` field via `store.Update(b.ID, beads.UpdateOpts{Type: &t})` on the read path. 
This is a critical architectural and coexistence risk:
1. **Violation of TR-002:** "Fact construction does not mutate session metadata."
2. **Races on read-only endpoints:** Read-only operations like API inspect (`/sessions/{id}/inspect`) or CLI status queries can trigger database writes if they encounter an empty-type bead, which is highly dangerous on concurrent lookups.
3. **No Lock Coordination:** `RepairEmptyType` runs outside of `withSessionMutationLock` and any command boundary, risking database locking issues or conflicting writes during parallel execution.

**Required change:** State in `DESIGN.md` whether `RepairEmptyType` is a pre-classification adapter (called explicitly before classification, keeping the classifier pure) or if it must be wrapped in a specific, locked repair command with an explicit retirement schedule (W-019).

---

### [Major] Per-slice revert contracts are incomplete and non-operational

Slices 2 through 6 have incomplete revert rules. For example, Slice 3's revert rule is:
`Revert only as whole runtime-start slice; do not split prepare/commit/rollback ownership.`
This is a directive, not an operational rollback contract. A safe rollback plan must explicitly specify:
- Which files and API handlers must be rolled back.
- Which static guard exceptions must be restored in the allowlist.
- Whether the rollback requires coordination with the in-flight worker-boundary migration.

**Required change:** Expand the "Bake and revert rule" column in `DESIGN.md:760–771` to name the exact files, static-guard exception IDs, and configuration toggles required to execute a revert for each slice.

---

## Required Changes

1. **Incorporate `cmd_session_wake.go` Fallback:** Add the direct write in `cmd/gc/cmd_session_wake.go:82–94` to the writer inventory (W-004) and detail its conversion into the explicit wake command.
2. **Add a Cross-Family Write Policy:** Add a dedicated section in `DESIGN.md` defining how multi-family operations (like `CloseDetailed`, `Suspend`, and `UpdatePresentation`) validate and commit keys belonging to different owner slices.
3. **Document API Routing Matrix:** Detail the legacy and Huma API surfaces in the shared call-site plan and map their step-by-step migration to the new command boundary.
4. **Deprecate `session.WakeSession`:** Convert all direct package-level calls to `session.WakeSession` in the API layer into proper command-boundary calls.
5. **Formally Banish Reads with Side-Effects:** Relocate `RepairEmptyType` out of the read-only lookup path or explicitly wrap it in a locked repair slice command.
6. **Flesh out Revert Contracts:** Provide step-by-step rollback steps for every slice in the coexistence table.

---

## Questions

- When Slice 4 extracts the close command, will `clearWakeAndHoldOverrides` and `retireConfiguredNamedSessionIdentifiers` become sub-commands called by the close command, or will the close command temporarily absorb these writes until Slices 1 and 5 are fully implemented?
- Does the mutation boundary cover top-level bead fields like `Type` and `Status` (which `RepairEmptyType` and `session_beads.go` modify), or is it strictly scoped to `Metadata` keys?
- How do we serialize concurrent CLI direct writes and reconciler ticks during the coexistence period when `withSessionMutationLock` is only in-process and invisible to the CLI?
