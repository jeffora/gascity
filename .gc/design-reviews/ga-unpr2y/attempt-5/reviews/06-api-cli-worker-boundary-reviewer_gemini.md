# API/CLI Worker Boundary Review — Independent DeepSeek Review

**Verdict:** block

**Review focus:** Cross-document consistency, missed edge cases, pattern drift, and assumptions other reviewers accepted too quickly. All claims grounded in source evidence.

---

## Top Strengths

- **Solid Architecture:** The functional-core / imperative-shell target shape (DESIGN.md:38–43) is the right decomposition: pure deciders behind typed commands, with subscribers reacting to post-commit facts. This cleanly separates session-domain decisions from I/O and scheduling.
- **Progressive Coexistence Plan:** The per-slice coexistence gates (DESIGN.md:760–771) with explicit "converted callers" / "legacy callers allowed" / "validation differences" / "guard update" columns give a concrete rollback path for each slice.
- **Detailed Mutation Inventory:** The mutation boundary inventory (DESIGN.md:420–450) with explicit exception rows and retirement conditions is materially better than prior iterations. The shrink-only allowlist rule prevents scope creep.
- **Shared Call-Site Plan:** The shared call-site table (DESIGN.md:772–781) anchors key integration points (`session_resolution.go`, `session_manager.go`, `worker_handle.go`) and sets clear end states.

---

## Critical Risks

### [Blocker] The API handler routing matrix is incomplete and the design conflates two independent handler surfaces

The design's shared call-site plan (DESIGN.md:772–781) lists key API files but does not distinguish or inventory the **two independent handler surfaces** that serve the same HTTP endpoints. Verified from source:

| Operation | Legacy handler routing (under `internal/api/handler_sessions.go`) | Huma handler routing (under `internal/api/huma_handlers_sessions_command.go`) | Wake mutation path |
|---|---|---|---|
| Suspend | `handle.Stop(ctx)` via `workerHandleForSession` | `mgr.Suspend(id)` via `sessionManager(store)` | Manager method |
| Close | `handle.CloseDetailed(ctx)` via `workerHandleForSession` | `mgr.CloseDetailed(id)` via `sessionManager(store)` | Manager method |
| Wake | `session.WakeSession(store, b, now)` direct package call | `session.WakeSession(store, b, now)` direct package call, then `handle.Start()` via `workerHandleForSession` | **Package-level function, not a Manager method** |
| Stop | — | `mgr.StopTurn(id)` via `sessionManager(store)` | Manager method |
| Kill | — | `mgr.Kill(id)` via `sessionManager(store)` | Manager method |

Three distinct routing patterns coexist in the API layer today:
1. **worker.Handle** (legacy suspend/close): `s.workerHandleForSession(store, id)` → `handle.Stop/CloseDetailed`
2. **session.Manager** (Huma close/stop/kill/suspend): `s.sessionManager(store)` → `mgr.CloseDetailed/StopTurn/Kill/Suspend`
3. **Direct package function** (wake, both handlers): `session.WakeSession(store, b, now)` + optional `handle.Start()`

The design's boundary principle (DESIGN.md:745–748) says "session-owned commands do not give `cmd/gc` permission to bypass `worker.Handle`" but says **nothing** about `internal/api`. The writer inventory (W-014, W-015, W-016) lists API files but does not state which routing pattern each handler uses today or which pattern it should converge to. The coexistence gates (DESIGN.md:760–771) only list `cmd/gc` converted callers — API handlers are absent from every slice's converted-caller column except slice 1 (target resolver).

**Required change:** Add a per-handler routing inventory to the design showing exactly how each API lifecycle handler currently reaches session mutation. State explicitly for each operation whether the end state routes through `worker.Handle`, through a session command factory, or through `session.Manager` — and which is the canonical path. Extend every coexistence gate to include the API handler class, not just CLI/reconciler.

---

### [Blocker] `session.WakeSession` is a package-level function, not a Manager method — the design's command routing assumption is wrong for wake

All prior reviews correctly identified that wake bypasses `worker.Handle`, but none noted that `session.WakeSession` is a **package-level function** in `internal/session/waits.go`, not a `session.Manager` method. Verified from source:

```go
// internal/api/huma_handlers_sessions_command.go:877
nudgeIDs, err := session.WakeSession(store, b, time.Now().UTC())
```

```go
// internal/api/handler_sessions.go:459 (legacy wake)
nudgeIDs, err := session.WakeSession(store, b, time.Now().UTC())
```

The writer inventory W-003 lists `internal/session/waits.go` as owning "wake/wait updates on session beads" targeting "explicit wake and wait-hold commands" in slices 2 and 5. But the design doesn't acknowledge that API handlers call this function **directly** — not through `session.Manager` and not through `worker.Handle`. This matters because:

1. The static guard (DESIGN.md:467–484) flags `SetMetadata`/`SetMetadataBatch`/`Update`/`Create`/`Close` calls, but `session.WakeSession` does its mutation internally. The guard would not catch an API handler that calls `session.WakeSession` directly because the bypass is at the function-call level, not the store-call level.
2. Slice 2's coexistence gate says `cmd/gc/cmd_session_wake.go` and "API wake path" are converted callers, but the actual conversion path is undefined — should the API call the wake *command* through `worker.Handle`, through a session command factory, or continue calling `session.WakeSession`?

**Required change:** In the writer inventory, split W-003 into the internal helper (which stays in `internal/session`) and the external call surface (`session.WakeSession` called from `internal/api` and `cmd/gc`). State explicitly which external callers of `session.WakeSession` exist today and which session command surface they should call in the end state. The static guard must cover direct `session.WakeSession` calls from outside `internal/session` as a bypass path once the wake command is introduced.

---

### [Blocker] The design's static guard specification cannot catch the actual API mutation patterns

The design says the guard should "flag production calls to `SetMetadata`, `SetMetadataBatch`, `Update`, `Create`, or `Close` when the receiver is a `beads.Store`" (DESIGN.md:469–471). But three of the four existing API mutation paths bypass the store entirely:

1. **`session.WakeSession(store, b, now)`** — mutates session metadata internally; the API handler never calls `SetMetadataBatch` directly. The guard sees a clean call chain.
2. **`mgr.CloseDetailed(id)`** — the Manager mutates internally; the API handler calls a Manager method. The guard sees a clean call chain.
3. **`store.SetMetadataBatch(b.ID, patch)` in `session_resolution.go:173`** — this IS a direct store write, but it uses a `RetireNamedSessionPatch` return value extended with an ad-hoc key (`patch["alias_history"] = ""`). The guard cannot determine whether `patch` contains owned keys because the patch is constructed by `session.RetireNamedSessionPatch` then mutated in-place by the caller.

Only pattern 3 is even theoretically catchable by the guard, and even there the ad-hoc key addition makes the key-inference problem harder than the design acknowledges. Patterns 1 and 2 are invisible to the guard as specified.

**Required change:** The guard specification must cover **three categories**, not just store calls:
- **Direct store writes** (current specification) — `SetMetadataBatch`/`SetMetadata`/`Update`/`Create`/`Close` on a `beads.Store`
- **Manager method calls** — `mgr.CloseDetailed`/`mgr.StopTurn`/`mgr.Kill`/`mgr.Suspend`/`mgr.Pending` called from outside `internal/session` and `internal/worker` should be flagged once a session command replaces the Manager method
- **Package-level function calls** — `session.WakeSession` called from outside `internal/session` should be flagged once the wake command exists

Each category needs its own allowlist and retirement condition. Without this, the guard will be a false negative for the primary API mutation paths.

---

### [Major] Raw lifecycle re-derivation in Huma wake handler violates the design's own principle

The Huma wake handler checks session closed status directly from raw metadata:

```go
// internal/api/huma_handlers_sessions_command.go:876
if b.Status == "closed" {
    return nil, huma.Error409Conflict("session " + id + " is closed")
}
```

This is a raw `Bead.Status` check, not a lifecycle projection. The design's Problem statement (DESIGN.md:310) says "callers can re-derive lifecycle, targeting, scaling, work-release, and runtime-observation rules from raw metadata" and the entire design exists to prevent this. Yet the Huma handler — the supposedly "typed" path — does exactly this.

The legacy wake handler does it differently: it calls `session.WakeConflictState(err)` after `WakeSession` returns an error, which maps the error to a conflict status. The Huma handler bypasses this by pre-checking `b.Status` before calling `WakeSession`, introducing a semantic fork: the Huma handler may return `409 Conflict` for "session is closed" before `WakeSession` even runs, while the legacy handler lets `WakeSession` handle the conflict and returns the error through `session.WakeConflictState`.

This means two handlers for the same operation can return different error shapes for the same state.

**Required change:** Add this site to the lifecycle projection adoption inventory. State whether the `b.Status == "closed"` check should use `ProjectLifecycle` or delegate to the wake command. If it stays as a pre-check, add it to the `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` guard.

---

### [Major] `handler_status.go` contains THREE unguarded raw-metadata lifecycle derivations

The design acknowledges (in prior review responses) that read-side API sites re-derive lifecycle from raw metadata, but the current guard and inventory are still incomplete. Verified from source:

1. `handler_status.go:325`: `if state := strings.TrimSpace(bead.Metadata["state"]); state != ""` — raw state read for status display
2. `handler_status.go:371`: `switch strings.TrimSpace(b.Metadata["state"])` — raw state count for active/suspended
3. `handler_status.go:488`: `state := session.State(strings.TrimSpace(b.Metadata["state"]))` with custom mapping: `"awake"` → `StateActive`, `"drained"` → `StateAsleep`

Item 3 is particularly concerning because `statusSessionState` implements its own state mapping that differs from `ProjectLifecycle`. `ProjectLifecycle` maps `"awake"` to `StateActive`, but it also handles quarantine, hold, and other derived states. The custom function maps `"drained"` to `StateAsleep`, which is a different semantic than `ProjectLifecycle` might produce (drained sessions may have drain-specific projections). The `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` test (lifecycle_projection_test.go:929) only checks 4 files and does not include `handler_status.go`.

**Required change:** Add all three `handler_status.go` sites plus `huma_handlers_sessions_query.go:300` to the read-side adoption inventory with an owning slice. Extend the `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` guard to cover `handler_status.go`. If read-side projection adoption is an explicit non-goal, state that with rationale — but the design's Problem section says this is exactly what the design exists to fix.

---

### [Major] The `RetireNamedSessionPatch` application in `session_resolution.go:171–173` extends the patch with an ad-hoc key

```go
// internal/api/session_resolution.go:171-173
patch := session.RetireNamedSessionPatch(now, "continuity-ineligible-replacement", spec.Identity)
patch["alias_history"] = ""  // ad-hoc key addition
if err := store.SetMetadataBatch(b.ID, patch); err != nil {
```

The design lists `RetireNamedSessionPatch` in the vocabulary table as a "Patch builder implementation" that "may remain inside `internal/session`; external callers stop applying patch maps directly." But this call site does two things the design doesn't capture:

1. It applies the patch directly (not through a session command).
2. It extends the patch with `patch["alias_history"] = ""`, which is an **ad-hoc metadata mutation** that adds a session-owned key (`alias_history` is in the Identity family per DESIGN.md:402) outside the patch builder's contract.

The mutation guard (DESIGN.md:469–471) should flag the `SetMetadataBatch` call, but the ad-hoc key addition means the guard must also detect mutations to patch maps after construction, which is not in the guard spec.

**Required change:** The design must either: (a) fold `alias_history` clearing into `RetireNamedSessionPatch` itself so the patch builder owns all keys, or (b) explicitly document that callers may extend patches with keys from the same family and add this to the guard's scope. Option (a) is strongly preferred: it aligns with the design's principle that external callers stop applying patch maps directly.

---

### [Major] The Huma close handler has an inline named-session policy check that the design doesn't inventory

```go
// internal/api/huma_handlers_sessions_command.go:822-826
if b, getErr := store.Get(id); getErr == nil &&
    strings.TrimSpace(b.Metadata[apiNamedSessionMetadataKey]) == "true" &&
    strings.TrimSpace(b.Metadata[apiNamedSessionModeKey]) == "always" &&
    strings.Contains(strings.TrimSpace(b.Metadata[apiNamedSessionIdentityKey]), "/") {
    return nil, huma.Error409Conflict("configured always-on named sessions cannot be closed while config-managed")
}
```

This is a **named-session close policy** embedded directly in the Huma handler, reading three raw metadata keys (`apiNamedSessionMetadataKey`, `apiNamedSessionModeKey`, `apiNamedSessionIdentityKey`) and applying a business rule. The design's close/retire slice (slice 4) says "move close/retire identity mutation behind a session-owned command" but doesn't mention this policy check. If the close command absorbs this policy, the API handler must delegate it. If it stays in the API layer, it's a read-side metadata re-derivation that should be in the lifecycle projection inventory.

**Required change:** Add this named-session close-policy check to the slice 4 inventory. State whether the close command absorbs the always-on rejection policy or whether the API handler retains it. If retained, add the metadata reads to the read-side projection adoption inventory.

---

## Minor Risks

- **Legacy handler untyped responses:** The legacy close handler returns `writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})` (handler_sessions.go:380), the suspend handler returns the same (line 343), and the wake handler returns `map[string]string{"status": "ok", "id": id}` (line 475). These are untyped wire shapes. The design's typed-wire gates require Huma-registered typed request/response types, but the design doesn't address the legacy handler migration path. Each legacy handler needs a per-handler migration entry.
- **Inconsistent Closed Checks:** The `humaHandleSessionWake` closed-session pre-check and the `WakeSession` internal check can produce different error shapes. The Huma handler returns `huma.Error409Conflict("session " + id + " is closed")` (line 876) before calling `WakeSession`, but if `WakeSession` itself detects a closed session, it returns a different error through `session.WakeConflictState`. The design doesn't specify which check is canonical and what happens if they disagree (e.g., a race where the session closes between the `store.Get` and the `WakeSession` call).
- **Unbounded Scan Performance:** The design's performance budget for "API path alias fallback" (DESIGN.md:566–567 of prior attempts) says "at most one all-session scan after prior exact/config/live resolution fails." But `resolveLiveSessionByPathAlias` (session_resolution.go:395–424) calls `session.ListAllSessionBeads` with no filter, which is an unindexed scan. The budget doesn't state whether this scan is bounded by any pagination or count limit, and the current code has none.
- **Unspecified Metadata in W-016:** The design lists `handler_session_create.go:481` as W-016 but doesn't specify what keys the `SetMetadataBatch` call writes. The guard needs to classify this call by key family. If it writes session-owned keys, the guard should have an exception row. If it only writes non-session keys, the guard should prove it.

---

## Missing Evidence

- No per-handler routing inventory showing which API lifecycle handler routes through `worker.Handle`, `session.Manager`, or direct `session.*` package functions.
- No guard coverage for Manager method calls (`mgr.CloseDetailed`/`mgr.Suspend`/etc.) or package-level function calls (`session.WakeSession`) from outside `internal/session`.
- No statement of whether `b.Status == "closed"` checks in Huma handlers should use `ProjectLifecycle` or remain as raw checks.
- No inventory of the `handler_status.go` raw-metadata derivations in any slice's adoption plan.
- No specification of how the `alias_history` key addition to `RetireNamedSessionPatch` in `session_resolution.go:172` should be handled — fold into patch builder or document as allowed extension.
- No named-session close-policy check in the slice 4 inventory.
- No legacy-handler typed-wire migration plan for `writeJSON(map[string]string{...})` responses.

---

## Required Changes

1. **API Handler Routing Inventory:** Add a per-handler API routing inventory to the design showing each lifecycle handler's current routing path (worker.Handle / session.Manager / direct session function) and target routing path. Extend every coexistence gate to include API handler classes.
2. **Wake Call Surface Separation:** Split W-003 into internal helper vs. external call surface. State which callers of `session.WakeSession` exist outside `internal/session` and which command surface they should call in the end state. Extend the static guard to cover Manager method calls and package-level session function calls from outside `internal/session`.
3. **Closed Status Checks Consolidation:** Add the `b.Status == "closed"` check in `humaHandleSessionWake` and the named-session close-policy metadata reads to the lifecycle projection adoption inventory.
4. **Read-Side Status Adoption:** Add `handler_status.go:325,371,488` and `huma_handlers_sessions_query.go:300` to the read-side adoption inventory with an owning slice. Extend `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` beyond its current four-file scope.
5. **Retire Patch Integrity:** Either fold `alias_history` clearing into `RetireNamedSessionPatch` or document it as a permitted patch extension and add it to the guard scope.
6. **Named-Session Close Policy Placement:** Add the named-session always-on close-policy check to slice 4's inventory. State whether the close command absorbs this policy.
7. **Legacy Handler Response Migration:** Add legacy-handler typed-wire migration entries for the three `writeJSON(map[string]string{...})` response patterns.

---

## Questions

- Should all API lifecycle handlers converge to `worker.Handle` as the canonical path, or should `internal/api` call session commands through a different boundary (e.g., a session command factory)?
- Is the `b.Status == "closed"` pre-check in `humaHandleSessionWake` intended to be absorbed by the wake command, or will the API handler retain its own closed-state guard? If retained, should it use `ProjectLifecycle`?
- What is the intended ownership of the named-session always-on close policy — session domain or API-domain? The current code embeds it in the API handler, but the design says close/retire moves into session-owned commands.
- How should the static guard handle `session.WakeSession` calls from outside `internal/session` once the wake command exists — flag them as a bypass, or allowlist them temporarily?
- Does the `alias_history` key addition in `session_resolution.go:172` indicate that `RetireNamedSessionPatch` is incomplete (missing a key it should own), or is this a legitimate caller-side extension that the design should formally permit?
