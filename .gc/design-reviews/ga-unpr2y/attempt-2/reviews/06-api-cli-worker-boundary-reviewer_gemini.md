# Marcus Webb — DeepSeek V4 Flash (Independent Review, Attempt 2)

**Verdict:** block

**Review focus:** Cross-file consistency, missed edge cases, pattern drift, and architectural coherence between the API/CLI/worker boundary and the design document. Independent adversarial scrutiny of assumptions carried forward from other reviews.

---

## Strengths (Carried Forward From Attempt 1)

The iterated design has genuine structural improvements over the first draft:

- The per-operation permission matrix (DESIGN.md:140–180) makes the classification contract reviewable and ties caller classes to specific operations with allowed/disallowed cells. This is auditable.
- The command atomicity contract (DESIGN.md:260–276) with precondition, written keys, ordering, conflict, and retryability columns gives implementers concrete contract terms.
- The worker-boundary migration sequencing (DESIGN.md:480–510) correctly states session-owned commands do not give `cmd/gc` permission to bypass `worker.Handle`.
- The mutation landscape inventory (DESIGN.md:100–140) names real files and maps them to owner slices. This is the right structure for guard baseline and retirement criteria.
- The scenario traceability matrix (DESIGN.md:209–248) ties every slice to `SESSION-*` rows, current proof paths, and freshness gates. Missing proof honestly blocks extraction.

---

## Critical Risks

### [Blocker] The API has six distinct routing patterns for lifecycle operations — the design specifies none of them

Verified in source code. Every lifecycle operation in the API routes through a different boundary surface:

| Operation | Legacy handler | Huma handler | Boundary path |
|---|---|---|---|
| Create | N/A | `worker.Handle.Create()` | `worker.Handle` ✓ |
| Wake | `session.WakeSession()` + `handle.Start()` | `session.WakeSession()` + `handle.Start()` goroutine | Direct mutation + worker runtime |
| Close | `handle.CloseDetailed(ctx)` | `mgr.CloseDetailed(id)` | Mixed: legacy through handle, Huma through `session.Manager` |
| Suspend | `handle.Stop(ctx)` | `mgr.Suspend(id)` | Mixed: legacy through handle, Huma through `session.Manager` |
| Rename | `handle.Rename(ctx, title)` | `mgr.Rename(id, title)` | Mixed: legacy through handle, Huma through `session.Manager` |

The design's worker-boundary section (DESIGN.md:480–510) states:

> Session-owned commands do not give `cmd/gc` permission to bypass `worker.Handle`.

This is specifically scoped to `cmd/gc`. The design's shared call-site plan lists `internal/api/session_manager.go` as an *"Active root-documented exception until retired"* and `internal/api/huma_handlers_sessions_command.go` must *"preserve Huma typed-wire behavior."* But neither entry specifies whether the API should reach session commands through `worker.Handle`, through `session.Manager`, or through session-owned command constructors directly. The root `AGENTS.md` migration text explicitly documents `internal/api/session_manager.go` and `internal/api/session_resolution.go` as active exceptions, but the design does not extend that documentation into a per-caller-class routing table.

This matters because the design introduces new command surfaces (decider + applier) but never specifies which surface the API calls. Without a stated target:

1. The close path's inconsistency (`handle.CloseDetailed` vs `mgr.CloseDetailed`) is not flagged as a bug, a migration artifact, or a design decision — it is simply undocumented.
2. The suspend path routes through `session.Manager.Suspend()`, which performs the full lifecycle mutation internally (lock, validate, runtime stop, metadata write) — exactly the pattern the design says external callers should not use. But `session.Manager` *is* `internal/session`, so this is "inside the boundary" while simultaneously being "outside the command-applier model." The design does not address this.
3. The wake path commits wake metadata through `session.WakeSession()` (a package-level function, not on `Manager`), then dispatches runtime start through `handle.Start()` in a goroutine. The metadata commit and runtime dispatch are decoupled — a design the atomicity contract (DESIGN.md:260–276) explicitly says must be addressed: "validate -> commit wake metadata -> optional event/trace -> reconciler scan starts runtime." The current code commits metadata and then *immediately* starts runtime in the same handler, with no recovery path for a start failure after metadata commit.

**Required change:** Add a per-caller-class routing table. For each caller class (CLI, API, reconciler), state the current boundary surface, the target end state, and the migration path. Then audit every current lifecycle operation in `internal/api/` against that table and list every deviation as a named migration entry with owner slice and retirement condition.

### [Blocker] `session.Manager.Suspend()` contradicts the design's command-applier model — and it is not listed in the mutation landscape

`session.Manager.Suspend()` at `internal/session/manager.go:754–849` does the following in a single method:

1. Acquires a per-session mutation lock
2. Reads the bead and checks `b.Status == "closed"`
3. Validates transition via `Transition(current, CmdSuspend)`
4. Calls `m.sp.Stop(sessName)` (runtime side effect)
5. Writes `state=suspended` + `suspended_at` directly via `m.store.Update(id, UpdateOpts{Metadata: map[string]string{"state": string(StateSuspended), "suspended_at": time.Now().UTC().Format(time.RFC3339)}})`

This is exactly the "validate, mutate metadata, side-effect" pattern the design's command-applier model (TR-004) requires session-owned commands to encapsulate. But:

- The method is on `session.Manager`, not a typed command struct.
- The design's command atomicity contract (DESIGN.md:260–276) does not list Suspend.
- The mutation landscape inventory (DESIGN.md:100–140) lists `internal/session/manager.go` as a monolithic entry with "lifecycle, create/start, identity, runtime identity" as field families and "Session commands implemented here or split below this package" as the target path. It does not call out Suspend as a specific mutation path.
- The Huma suspend handler calls `mgr.Suspend(id)` which bypasses `worker.Handle` entirely, and the design does not acknowledge this as an API-specific routing choice.

The same pattern applies to `Manager.CloseDetailed()` (lines 862–922): it acquires the lock, validates, stops runtime, cancels waits, clears overrides, retires named identity, and closes the bead — all within one method. The design says external callers should call session-owned commands, but the command-applier model does not specify whether `Manager` methods *are* those commands or whether they need to be decomposed into fact-reader → decider → applier.

**Required change:** State whether `session.Manager` methods become the command surface after extraction or are replaced by per-operation command structs. If they remain, add them to the mutation landscape with explicit ownership of which field families each method writes. If they are replaced, add a migration entry showing the transition. Then add Suspend (and any other unlisted methods) to the command atomicity contract with their precondition, written keys, ordering, and conflict columns.

### [Blocker] The Huma wake handler returns 500 for wake conflicts — a typed-wire regression with no design acknowledgment

The Huma wake handler at `internal/api/huma_handlers_sessions_command.go:881` calls `session.WakeSession()` and on error returns `huma.Error500InternalServerError(err.Error())`. The legacy handler at `internal/api/handler_sessions.go:456` checks `session.WakeConflictState(err)` and returns 409 Conflict. This is a concrete typed-wire regression: the Huma handler loses conflict semantics that the legacy handler preserves.

The design's TR-006 says "API, mail, bead assignee normalization, and extmsg callers stop re-deriving these categories." But the current API is *already* inconsistent in how it exposes wake conflicts. The Huma handler also checks `b.Status == "closed"` directly (line 877) instead of using `ProjectLifecycle` or the wake conflict state — a read-side bypass that the design's TR-006 and the `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` guard should catch, but the guard only covers 4 pinned strings in 4 files.

This is not a theoretical concern. The legacy handler's 409 response is what callers currently receive; the Huma handler's 500 response is what they will receive after migration. The design's "API, CLI, And Typed-Wire Gates" section (DESIGN.md:383–400) requires "API status codes, response body/error body shape, request IDs, and async result event behavior" as per-slice compatibility proof — but there is no proof that the Huma wake handler's error response matches the legacy handler's response contract.

**Required change:** Add a typed-wire compatibility proof for the Huma wake handler that demonstrates 409 Conflict is returned for wake conflicts, matching the legacy handler's behavior. Also add `b.Status == "closed"` direct metadata reads in `huma_handlers_sessions_command.go` to the read-side adoption list, or replace them with `ProjectLifecycle` checks.

---

## Major Risks

### [Major] Read-side metadata bypasses are more numerous than the design acknowledges

The `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` test (at `internal/session/lifecycle_projection_test.go:929`) currently guards against exactly 4 ad-hoc reads in 4 files:

| Guarded file | Forbidden pattern |
|---|---|
| `internal/api/handler_sessions.go` | `b.Metadata["sleep_reason"]` |
| `internal/api/handler_sessions.go` | `strings.TrimSpace(b.Metadata["state"])` |
| `cmd/gc/cmd_session.go` | `if sr := b.Metadata["sleep_reason"]; sr != ""` |
| `cmd/gc/doctor_session_model.go` | `strings.TrimSpace(b.Metadata["state"])` |

But the current codebase has many more raw `Metadata["state"]` reads in `internal/api/` that are not guarded:

| File | Line | Pattern | Read purpose |
|---|---|---|---|
| `handler_status.go` | 325 | `bead.Metadata["state"]` | session state display |
| `handler_status.go` | 371 | `b.Metadata["state"]` | countSessions active/suspended |
| `handler_status.go` | 488 | `b.Metadata["state"]` | statusSessionState |
| `handler_status.go` | 419 | `b.Metadata["session_name"]` | status display |
| `handler_status.go` | 488 | `session.State(strings.TrimSpace(b.Metadata["state"]))` | session state display with compatibility mapping |
| `session_resolution.go` | 414 | `session.State(b.Metadata["state"])` | conflict state check |
| `huma_handlers_sessions_command.go` | 877 | `b.Status == "closed"` | closed-session check before wake |
| `huma_handlers_sessions_query.go` | 300 | `b.Metadata["state"] == "creating"` | creating-session skip for pending |
| `handler_sessions.go` | 470 | `b.Metadata["session_name"]` | crash history clear |
| `huma_handlers_sessions_command.go` | 888 | `b.Metadata["session_name"]` | crash history clear |
| `handler_extmsg.go` | 79 | `b.Metadata["session_name"]` | session name extraction |
| `handler_mail.go` | 225, 244 | `b.Metadata["session_name"]` | session name for message routing |

The design's read-side adoption rule says "production callers outside `internal/session` call session-owned command APIs" and the design's static guard says it should "forbid production call-sites outside `internal/session` and an explicit allowlist from reading session-owned metadata keys" — but the guard specification (DESIGN.md:117–140) only talks about write-side enforcement (`SetMetadata`, `SetMetadataBatch`, `Update`, `Create`, `Close`). It does not mention read-side keys at all. The `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` guard only covers 4 patterns in 4 files, leaving at least 10 unguarded raw reads.

The `handler_status.go` reads are particularly concerning: `statusSessionState()` at line 488 does its own compatibility mapping (`"awake"` → `StateActive`, `"drained"` → `StateAsleep`), which is exactly the kind of ad-hoc projection the design says should use `ProjectLifecycle`. The `countSessions()` method at line 371 counts active/suspended by raw state comparison, which will break if a new state is added.

**Required change:** Extend `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` to cover all identified raw metadata reads of session lifecycle keys, or add a separate read-side guard. List all unguarded reads in the design as explicit migration entries with owner slices.

### [Major] `handler_status.go` reimplements lifecycle projection — an unlisted pattern drift

`internal/api/handler_status.go` contains its own session-state logic:

- `statusSessionState()` (line 488) maps `"awake"` → `StateActive` and `"drained"` → `StateAsleep`, which is a partial copy of `ProjectLifecycle`'s compatibility mapping.
- `countSessions()` (line 371) switches on `b.Metadata["state"]` counting only `active` and `suspended`, missing `awake`, `drained`, `asleep`, and other projected states.
- `sessionStateLabel()` (line 325) returns raw `Metadata["state"]` or `"materialized"`, bypassing projection entirely.

None of these are in the design's mutation landscape or the read-side adoption plan. The `sessionStateLabel` function is used in status output that operators rely on, so it has a user-facing correctness requirement that the design's TR-001 (preserve product semantics) should cover. The `countSessions` function is used for pool status display, so undercounting sessions has operational impact.

**Required change:** Add `handler_status.go` to the read-side migration entries. Specify whether `statusSessionState` and `countSessions` should adopt `ProjectLifecycle` and if so, in which slice.

### [Major] The wake path decouples metadata commit from runtime start with no recovery specification

Both the legacy and Huma wake handlers commit wake metadata via `session.WakeSession()` and then start the runtime in a separate `handle.Start()` call. In the Huma handler, `handle.Start()` runs in an untracked goroutine (`go func() { ... }()`). The design's command atomicity contract for explicit wake (DESIGN.md:273) specifies the ordering as:

> validate -> commit wake metadata -> optional event/trace -> reconciler scan starts runtime

But the current code does the first two steps eagerly and dispatches runtime start without any failure path visible to the caller. If `WakeSession` succeeds and `handle.Start()` fails, the session bead is in `wake_requested` state with no running process. The design says "critical retryable" reactions must converge from durable facts — and durable wake metadata *is* the recovery path. But the current Huma handler does not emit a trace, event, or reconciler hint for this failure path. The goroutine just logs and discards.

This is especially important because the design's event contract says "event payloads describe facts that happened, not commands to attempt." A successful `WakeSession` metadata commit is a fact that happened, but the API response already returned 200 before the runtime start attempt. The caller has no way to distinguish "wake committed and runtime starting" from "wake committed and runtime failed to start."

**Required change:** Specify in the command atomicity contract what happens when `handle.Start()` fails after `WakeSession` commits. Options include: (a) the reconciler catches this on the next tick via desired-state convergence (the current implicit path), (b) the wake command should not dispatch runtime start but instead set wake metadata and let the reconciler own start, or (c) the wake handler must await runtime start outcome before responding. State which is the target and document the recovery path.

### [Major] `session_resolution.go` writes session-owned metadata keys — not just reads

The design's mutation landscape lists `internal/api/session_resolution.go` as an "Active root-documented exception until retired" for "target resolution and direct named session materialization." But `session_resolution.go` also writes session-owned keys:

- Line 173: `store.SetMetadataBatch(b.ID, patch)` where `patch` comes from `session.RetireNamedSessionPatch(now, "continuity-ineligible-replacement", spec.Identity)`.
- Line 213: `store.Update(item.ID, beads.UpdateOpts{Assignee: &newID})` for work reassignment.

The first is a write of lifecycle identity metadata (`state`, `continuity_eligible`, `alias_history`, named-session markers) via a session-owned patch helper, applied from outside `internal/session`. The second is a work-domain mutation (assignee change) which is outside session ownership entirely but happens in the same code path. The mutation landscape's entry for `session_resolution.go` only describes "direct materializing create for resolved named sessions" — it does not mention the retirement mutation path or the work reassignment.

**Required change:** Update the `session_resolution.go` mutation landscape entry to cover both the create/materialization path and the retire/reassignment path. Specify whether the retire mutation is within scope for the target classifier (slice 1) or the close/identity retirement (slice 3).

---

## Minor Risks

### [Minor] The `session.WakeSession` function is not a method on `Manager` — its placement is inconsistent with the command model

`WakeSession` at `internal/session/waits.go:203` is a package-level function, not a method on `session.Manager`. This is the only lifecycle mutation function in the session package that is not a method on `Manager`. All other lifecycle operations (Suspend, Close, CloseDetailed, Rename, RequestFreshRestart) are methods on `Manager`. The design's command-applier model implies a uniform command surface, but `WakeSession` takes a `beads.Store` and `beads.Bead` directly rather than going through a session-aware command. This means the API can call it without constructing a `Manager`, which is architecturally inconsistent.

**Suggested:** Either move `WakeSession` onto `Manager` or acknowledge in the design that it is a transition artifact that will be absorbed into the wake command when slice 2 extracts it.

### [Minor] `huma_handlers_sessions_query.go:300` reads `Metadata["state"] == "creating"` as a raw string comparison

The pending-interaction handler at `internal/api/huma_handlers_sessions_query.go:300` checks `b.Metadata["state"] == "creating"` directly to decide whether to skip pending-interaction resolution. This is a read-side bypass of `ProjectLifecycle` — the creating-state check should use lifecycle projection, especially since `ProjectLifecycle` already has creating-state staleness logic (`fresh-creating` vs `stale-creating`). A raw `"creating"` check does not distinguish between a fresh and stale create, which could lead to incorrect pending-interaction behavior.

**Suggested:** Add this to the read-side migration entries under the query path slice or the target classifier.

### [Minor] The `session_resolution.go` kill-on-retire pattern bypasses the worker boundary

At `internal/api/session_resolution.go:167`, the named-session replacement path does:

```go
if handle, err := s.workerHandleForSession(store, b.ID); err == nil {
    _ = handle.Kill(context.Background())
}
```

This kills the runtime process as a side effect of named-session retirement, before the retire metadata is committed (line 173). If `Kill` succeeds but `SetMetadataBatch` fails, the session is dead but its metadata still claims it owns the identity. This is a minor ordering issue (kill before metadata commit vs metadata commit before kill), but it is not documented in the design's command atomicity contract for close/retire operations.

---

## Missing Evidence

1. **Per-caller-class routing table:** Which caller class (CLI, API, reconciler) reaches session commands through which boundary surface, and what is the target end state.
2. **`session.Manager` target shape:** Whether it becomes a command factory, is replaced by command structs, or remains the API surface after extraction.
3. **Suspend atomicity contract:** Precondition, written keys, ordering, conflict reasons, and retryability for the Suspend operation.
4. **Complete read-side adoption owner:** Which slice (or non-goal) owns moving API raw `Metadata["state"]` reads onto `ProjectLifecycle` or the classifier.
5. **Huma wake handler typed-wire proof:** Demonstrating 409 Conflict responses for wake conflicts, matching legacy behavior.
6. **Recovery path for post-metadata-commit runtime failures in the wake path.**
7. **`handler_status.go` read-side migration plan:** Whether `statusSessionState`, `countSessions`, and `sessionStateLabel` adopt `ProjectLifecycle` and in which slice.

---

## Required Changes Summary

1. **Add a per-caller-class routing table** with current boundary surface, target end state, and migration path for every lifecycle operation in CLI, API, and reconciler. Audit the current codebase and list every deviation as a named migration entry.
2. **Add Suspend (and any other unlisted Manager methods) to the command atomicity contract** with precondition, written keys, ordering, and conflict columns.
3. **State whether `session.Manager` is the target command surface** after extraction or is decommissioned in favor of per-operation command structs.
4. **Fix or document the Huma wake handler's 500-for-conflict regression** — add a typed-wire compatibility proof showing 409 Conflict is returned for wake conflicts.
5. **Extend the read-side guard** (`TestLifecycleUserFacingConsumersStayOnProjectionHelpers`) to cover all identified raw `Metadata["state"]` and `Metadata["session_name"]` reads in `internal/api/`, or add a separate read-side guard specification to the design.
6. **Add `handler_status.go` to the read-side migration entries** with explicit adoption targets for `statusSessionState`, `countSessions`, and `sessionStateLabel`.
7. **Specify the recovery path for the wake handler's decoupled metadata-commit/runtime-start pattern** — document whether the reconciler catches this on the next tick, the wake command should not dispatch runtime start, or the handler must await runtime outcome.
8. **Update the `session_resolution.go` mutation landscape entry** to cover the retire/reassignment path, not just create/materialization.
9. **Acknowledge `WakeSession` as a package-level function** in the design and state whether it is a transition artifact or the intended long-term wake command surface.

---

## Questions

- Is `session.Manager` the target command API surface after extraction, or is it decommissioned in favor of per-operation command structs? The answer determines whether `Suspend`, `CloseDetailed`, and `Rename` are refactored internally or replaced wholesale.
- For the Huma close handler, which uses `mgr.CloseDetailed(id)` instead of `handle.CloseDetailed(ctx)` — is this an intentional design choice (the Manager is the canonical close path) or an oversight that should be aligned with the legacy handler's `handle.CloseDetailed(ctx)` pattern?
- Should the `WakeSession` package-level function be absorbed into `Manager` or remain a standalone function that the wake command applier calls? The current placement means the API does not need a `Manager` to perform wake mutations, which is inconsistent with all other lifecycle mutations.
- For `internal/api/session_resolution.go:167`, should `handle.Kill()` happen before or after the retire metadata commit? The current code kills before committing, which creates a window where the session is dead but metadata claims it owns the identity. The design's command atomicity contract for close/retire says "close/retire metadata -> assignment and binding scans -> event/trace" — is this ordering also required for retire-then-replace?
