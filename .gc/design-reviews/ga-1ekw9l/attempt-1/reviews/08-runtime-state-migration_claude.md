# Yelena Markovic - Claude

**Verdict:** block

Reviewed `plans/core-gastown-pack-migration/implementation-plan.md` (`updated_at`
2026-06-09T13:20:59Z) against the current `requirements.md` (`updated_at`
2026-06-09T17:23:58Z, AC10) and `implementation-plan.schema.md`. Lane:
Maintenance→Core runtime-state migration — JSONL/archive state, spawn-storm
ledgers, order skip/tracking, escalation fields, non-destructive markers,
old/new binary concurrency, downgrade continuity.

Grounded against the live tree: the migrated state is real and produced by
concrete Maintenance-pack orders —
`examples/gastown/packs/maintenance/orders/{mol-dog-jsonl, spawn-storm-detect,
order-tracking-sweep, mol-dog-reaper}.toml` (all four confirmed present;
`spawn-storm-detect.toml` is a `cooldown`-triggered `spawn-storm-detect.sh`
script) — and owned by `internal/sessionlog` (JSONL readers), `internal/config`
(order skip list `config.go:1481`, throttle `config.go:1843`), and
`internal/runtime/process_control.go` (escalation). No migration-marker or
mutation-coordinator infra exists today; Slice 4c builds it greenfield, which is
exactly why the detection and merge rules must be precise now. This blocker is
AC10/lane-substantive and is *not* affected by the later requirements update —
the runtime-state content of the plan is unchanged.

**Top strengths:**
- Migration is correctly scoped as doctor-owned, single-writer (city advisory
  lock + no-running-controller refusal discovered from live state, not status
  files — lines 379-383, 392-397; `controllerRunning` discovery already exists at
  `cmd/gc/cmd_doctor.go:240-243`) and non-destructive: legacy state is retained
  and "ignored unless the marker or digest checks show conflict" (578-579); refs
  are reconciled and unknown refs are never deleted (586).
- The per-artifact merge table (583-591) enumerates every surface in this lane
  with an explicit rule and a fail-closed "conflicts block" default; Slice 4c is
  gated late behind a proven 4b coordinator with "marker, quiesce, old-binary,
  downgrade, and re-upgrade tests" (804).
- Old/new concurrency is acknowledged with post-marker old-binary write detection
  + version-skew halt (404-407), and staged copy with source+destination digests
  guards against treating a half-copied archive as complete (402, 583) — the
  right design *intent* for red flags #1 and #2.

**Critical risks:**
- **[Blocker] The no-data-loss guarantee rests on detection the plan never
  fingerprints (lane Q2 / red flag #2).** The marker records the "staged
  archive-copy digest" — the digest of the *copy written to Core* — and lists
  "post-marker old-binary write detection" as a marker *field* (575-579), but
  never records a fingerprint of the legacy *source* at marker commit (per-path
  digest+size, or a monotonic write-generation counter). An old binary by
  definition does not know the marker exists; with no recorded source baseline the
  new binary has nothing to compare the legacy JSONL/throttle path against, so a
  post-marker append is undetectable. The plan's entire safety model is "detect
  conflict → block" (578-579), and retaining legacy bytes is inert for safety if
  the new binary never re-reads them against a baseline. Naming the goal
  ("detection") is not the mechanism; a faithful implementer could ship an mtime
  or no-op check that silently misses appends and still satisfy the plan. Red flag
  #2 ("post-marker old-binary writes are silently ignored") is therefore not
  closed and lane question 2 is unanswered. Because this is the lane's central
  mandate over operator archives and git state, it blocks decomposition, not
  merely risks it.
- **[Major] No Current-System grounding for the migrated state.** "Current System"
  (31-76) never mentions JSONL archives, spawn-storm/throttle ledgers, order
  skip/tracking, push cursors, or escalation fields. Verified that state is
  produced by the four locatable Maintenance-pack orders above and owned by
  `internal/sessionlog`, `internal/config`, and `internal/runtime/process_control.go`.
  The migration table specifies *destinations* for state whose current
  paths/formats the plan never cites, so a decomposer must rediscover them and
  behavior preservation cannot be proven against an unstated baseline — the
  schema's "not ready when it requires readers to infer the actual change from
  principle" rule.
- **[Major] Spawn-storm union has no dedup key and goes blind post-marker (red
  flag #1/#2).** "read-union before marker, Core-only after marker" (589): a union
  of old+Core throttle events with no stated dedup key can double-count
  (over-throttle) or mis-count across disjoint windows; and "Core-only after
  marker" makes a storm that a post-marker old binary records only in the OLD
  ledger invisible to throttling unless the (unspecified) detector explicitly
  monitors the old throttle-ledger path. The producer is a script
  (`spawn-storm-detect.sh`) reachable outside the controller, so "no controller
  running" does not obviously quiesce it.
- **[Major] Push-cursor reconciliation is underspecified and unproven against
  duplicate push (red flag #3).** "newest verified cursor wins, conflicts block"
  (587) gives no total order ("newest" by sequence or timestamp?), no definition
  separating "newest wins" from "conflict," and no proof that an already-pushed
  archive is not re-pushed. If push is not idempotent against the reconciled
  high-water mark, a wrong "newest wins" pick re-pushes pending state.
- **[Major] Order-skip alias retirement orphans old-key entries (red flag #3).**
  "aliases suppress same logical order until retired" (590) correctly stops a
  renamed order re-firing while the alias lives, but the skip list is keyed by the
  *old* order key (`config.go:1481`) and nothing re-keys the actual skip/tracking
  entries at retirement. When the alias retires, a previously-skipped order
  re-fires. Re-keying (or blocking retirement until live entries are migrated)
  must be specified.
- **[Major] No concurrent/duplicate-writer round-trip test, and archive copy lacks
  process-unique staging + single atomic commit (lane Q3 / red flag #1).** Testing
  (657-667) names failure-injection, push-cursor reconcile, old-binary, and
  downgrade tests — not a two-writer archive-copy race; the only "concurrent
  promotion" test (670) is the cache, not archives. Unlike AC16's cache writes
  (mandated "randomized or process-unique staging paths"), the runtime-state
  archive copy only says "staged copy with source and destination digests"
  (402, 583) with no process-unique staging, no destination-absent first-copy
  case, and no single-atomic-commit requirement — so a half-copy at a fixed
  staging path can be resumed as complete.
- **[Minor] Downgrade continuity asserted without a readability contract.**
  "supports downgrade or manual recovery guidance" (666-667, 818) never states
  whether Core-owned paths written by the new binary are readable by an old
  binary, or whether downgrade is strictly manual — the reverse of the
  new-binary-sees-old-write case. "Core-only after marker" implies post-marker
  new-binary state is stranded on downgrade; the plan must enumerate what is lost.

**Missing evidence:**
- The post-marker detection fingerprint (a per-path legacy-source baseline) and
  where it is recorded in the marker.
- Current on-disk paths, formats, and writers for each migration-table row, and
  proof that "no controller running" quiesces every writer (the spawn-storm and
  order-tracking producers are Maintenance-pack scripts/orders, not controller
  core).
- A single combined interrupted-copy + concurrent-writer + rollback round-trip
  test; lane Q3 asks for one, the plan lists separate pieces.
- The "deterministic re-upgrade flow" (407, 764) merge semantics — does it fold
  post-marker legacy JSONL entries into Core (no loss) or discard them?
- Known-ref conflict handling: line 586 only addresses unknown-ref deletion, not
  the same known ref pointing at different commits in legacy vs Core.

**Required changes:** *(all before Slice 4c is decomposed/approved)*
- Record a per-migrated-path legacy-source fingerprint (digest+size or
  write-generation counter) in the marker at commit time and define post-marker
  detection against it, proving the no-data-loss guarantee. Define behavior for
  append-only JSONL where any new record changes the fingerprint.
- Add Current-System grounding (paths, packages, writers) for every
  migration-table row, citing the four Maintenance-pack orders and the owning
  packages (`internal/sessionlog`, `internal/config`,
  `internal/runtime/process_control.go`).
- Define the spawn-storm union dedup key and wire the old throttle ledger as a
  monitored post-marker path so old-ledger storms are not lost.
- Make push-cursor reconciliation a single deterministic, idempotent policy with a
  stated conflict definition; add a mid-push interruption test proving no
  duplicate or dropped push.
- Specify skip/tracking re-keying at alias retirement (or block retirement until
  live entries are re-keyed), with a regression test that a previously-skipped
  order does not re-fire after retirement.
- Require process-unique/randomized staging and a single atomic commit
  (same-filesystem rename, or a documented non-atomic fallback) for runtime-state
  archive copy; define the destination-absent first-copy case; add an explicit
  two-writer / old-binary-controller-live refusal round-trip test.
- State whether new Core-owned runtime-state paths are readable by an old binary,
  or declare downgrade manual-only and enumerate the post-marker state lost.

**Questions:**
- What fingerprint detects post-marker old-binary writes, and is it recorded per
  migrated path at marker commit?
- Does the live-controller probe detect an OLD-binary controller (which never
  takes the new advisory lock), and what happens when liveness is indeterminate?
- Is archive push idempotent against the reconciled cursor, and what exactly
  separates "newest wins" from "conflict"?
- At alias retirement, what migrates existing skip/tracking entries keyed by the
  old order key?
- On downgrade, are new Core-owned runtime-state paths readable by an old binary,
  or is downgrade manual-only?

**Schema conformance (implementation-plan.schema.md, non-empty):** Conforms. Front
matter carries all required keys with `phase: implementation-plan`,
`requirements_file` pointing at the approved requirements, `status: draft`, no
`design_file`; the seven required body sections are present and in order; the
migration table sits correctly in `Data And State`; `Open Questions` is `None`.
Minor nit (not lane-driving): inline `<!-- REVIEW: added per … -->` provenance
comments are borderline against the "no appended review-attempt summaries" rule
but pass as comments rather than sections. The block above is on lane content, not
schema shape.
