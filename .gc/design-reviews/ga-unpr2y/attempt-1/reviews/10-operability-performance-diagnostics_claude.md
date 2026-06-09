# Ingrid Holm - Claude

**Verdict:** approve-with-risks

Lane: decision observability, trace and doctor diagnostics, fact read cost,
event fan-out load. Reviews the current `DESIGN.md` (attempt-15 review-response
revision) with `REQUIREMENTS.md` and `internal/session/AGENTS.md`. All trace,
resolver, and reconciler claims verified against the checkout; citations inline.
The active global verdict is `block` and only a non-mutating Slice 0 is
authorized, so this disposition governs whether the diagnostics/cost apparatus
Slice 0 must build is sound.

**Top strengths:**
- The Observability And Cost section answers all three lane questions head-on:
  structured diagnostic result fields — operation ID, result kind, reason code,
  retryability, selected identity, source facts, missing/stale fact indicator,
  renderer surfaces, redaction keys (DESIGN.md:544-554) — plus
  `DIAGNOSTICS_MANIFEST.yaml` owning trace mappings, doctor/inspect output, query
  and subprocess budgets, and a hot-path budget table (DESIGN.md:556-583). It
  correctly singles out `resolveLiveSessionByPathAlias` for index/remove/budget,
  and I confirm the concern is real: it calls
  `session.ListAllSessionBeads(store, beads.ListQuery{})` — an unbounded
  all-session scan with a per-bead `Metadata["state"]` read
  (`internal/api/session_resolution.go:392`) — on the query hot path.
- The diagnostics-to-trace requirement maps onto a real, rich subsystem, not an
  invented one. `gc trace` has a first-class `TraceSiteCode` enum (~30 sites) in
  `cmd/gc/session_reconciler_trace_types.go:57+`, including the exact deciders
  slices 3-5 will extract — `session_reconcile.wake_sleep`,
  `session_reconcile.drain_advance`, `session_reconcile.start_execution` — plus a
  `--reason` filter (`session_reconciler_trace_cmd.go:193`) and an existing
  `engdocs/contributors/reconciler-debugging.md` runbook. "Every diagnostics row
  must map to one `gc trace` site/reason/outcome rendering" (DESIGN.md:561) is
  therefore enforceable against a concrete schema.
- The event/recovery contract serves observability correctly: message text is
  operator-only and subscribers must consume typed payload + envelope fields
  (DESIGN.md:76,533-535), a concrete 13-event `session.*` inventory is required
  (DESIGN.md:521-528), and negative tests must prove "machine-readable data is
  not hidden only in message strings" (DESIGN.md:562-564), backed by
  skipped/duplicate/stale-event tests (DESIGN.md:515).

**Critical risks:**
- [Major] **Decider diagnostics are specified as a generic "reason code" and do
  not require carrying the typed `WakeCause`/blocker that operators rely on today
  — red flag #1.** `ProjectLifecycle` already produces a typed `WakeCause` enum
  (`internal/session/lifecycle_projection.go:119-130`: pending-create, pin,
  attached, pending, work, scale-demand, explicit) and blocker overlays (held,
  quarantined, missing-config, identity-conflict, duplicate-canonical), and
  `LifecycleDisplayReason` renders the operator-facing "why"
  (`internal/session/lifecycle_projection_test.go:623+`). But the design's
  diagnostic result fields (DESIGN.md:544-554) list only a generic "reason code,"
  and the Target Classification result schema (DESIGN.md:255-264) covers target
  resolution kinds, not lifecycle wake-cause/blocker. A wake/close/drain decider
  (slices 3-5) could satisfy the contract while collapsing the granular
  wake-cause/blocker into an opaque string, regressing "explain why this session
  was blocked/woken/drained/closed" from structured to prose.
- [Major] **On the production `bdstore` backend every `store.Get`/`SetMetadata` is
  a `bd` subprocess, so the design's "no subprocess loop" rule and its per-session
  reconciler reads/event writes are the same cost — and the design never says so.**
  `BdStore.Update`/`Get`/`SetMetadata` shell out to `bd` (`internal/beads/bdstore.go:737+`).
  The reconciler does `store.Get(session.ID)` in-loop
  (`cmd/gc/session_reconciler.go:219`) and writes the event throttle marker
  `stranded_event_emitted_at` via `store.SetMetadata` per session during the pass
  (`cmd/gc/session_reconciler.go:2478`); `resolveLiveSessionByPathAlias` scans all
  sessions per query. The budget table names "subprocess count"
  (DESIGN.md:582) — good — but the manifest must state that ordinary per-session
  store calls ARE subprocesses on the production backend; otherwise a "bounded
  indexed lookup" that still issues one subprocess per session silently violates
  the "no subprocess loop in ... reconciler hot loops" rule (DESIGN.md:572).
- [Minor] **Budget rows carry no current baseline, only future thresholds.**
  "Large-city baseline," "measured budget," and "threshold" (DESIGN.md:580-583)
  are deferred to Slice 0 with no current numbers, so there is no way to tell
  today whether `resolveLiveSessionByPathAlias`'s all-session scan or the
  reconciler's per-pass store-call count already exceeds an acceptable bound.
  Acceptable at design altitude, but the manifest should capture the current
  measured cost as the baseline, not only a target.

**Missing evidence:**
- No requirement that lifecycle decider diagnostics carry the typed
  `WakeCause`/blocker (the structured "why"), only a generic reason code.
- No acknowledgment that `bdstore` turns per-session store reads/writes into
  subprocesses, coupling the "no subprocess loop" rule to ordinary fact reads.
- No current measured baseline for any budget row (path-alias scan size,
  reconciler store-call count per pass, event fan-out size).

**Required changes:**
1. Require lifecycle deciders (wake/close/drain/identity-retirement) to emit the
   canonical typed `WakeCause` and blocker in their diagnostic result, with parity
   tests against the existing `LifecycleDisplayReason`/`WakeCause` rendering, and
   map each to its existing `gc trace` site (`session_reconcile.wake_sleep`,
   `.drain_advance`, `.start_execution`) so operator explainability does not
   regress.
2. Have `DIAGNOSTICS_MANIFEST.yaml` record, per hot path, that production
   `bdstore` store calls are subprocesses and budget "subprocess count" as the
   actual per-session store-call count, so a per-session `Get`/`SetMetadata` in a
   loop is counted as a subprocess loop. Require `resolveLiveSessionByPathAlias` to
   be indexed or removed (not merely "budgeted"), given it scans all sessions on
   the query hot path.
3. Record current measured baselines (scan row counts, store-call counts per
   reconcile pass, event fan-out size) in the budget rows, not only future
   thresholds, so regressions are detectable from day one.

**Questions:**
- Will lifecycle decider diagnostics carry the typed `WakeCause`/blocker enums, or
  a flattened reason string? The former preserves today's operator explainability;
  the latter regresses it.
- Given `bdstore` store calls are subprocesses, is the target an indexed/counting
  store lookup for path-alias and reconciler fact compilation, or an accepted
  measured subprocess budget? The design lists both as options without choosing.
- Does the cost budget count the existing per-pass `stranded_event_emitted_at`
  write, and does that lifecycle-adjacent hot-path metadata write move behind a
  session-owned command?
