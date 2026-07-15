# Beads Backend: Performance and Concurrency, Not Sync (v2)

**To:** Adjunct
**From:** Mayor's research assignment (ga-7f0, re-scoping ga-vgk / v1)
**Date:** 2026-07-15

## Summary up front

v1 (`beads-backend-migration-report.md`) recommended keeping Dolt as the default because `refs/dolt/data` sync was assumed to be this project's cross-machine sync mechanism. That premise does not hold: **sync is dormant on this fleet.** I verified this directly (see below) rather than just taking HQ's sweep on faith. With sync out of the picture, the decision is purely about **local performance and concurrent-writer behavior** on a single Mac running four managed Dolt SQL servers, one per city.

The headline correction to v1: **the SQLite/Postgres/MySQL backend support v1 described does not exist in any released version of beads.** The Homebrew-installed `bd` (v1.1.0 — the same version the fleet's `gc` links) hard-fails on `--backend=sqlite` ("DEPRECATED: The SQLite backend has been removed... Dolt is now the default (and only) storage backend") and on `--backend=postgres` ("only \"dolt\" is supported"). `OpenConfigured` and `OpenNativeBeadsStoreAt` — the APIs v1's plumbing analysis (PRs #4736/#4154) hangs on — grep to zero hits anywhere in the current gascity source; only `OpenNativeDoltStoreAt` exists (`internal/beads/native_dolt_store.go:188`). v1 did flag the pinned-unreleased-commit risk in its "open questions" section, but its summary and options table read as if the SQL backends are available today. They are not: they live only on an unmerged/unreleased beads commit. Everything below evaluates the decision on that corrected footing.

**Recommendation, unchanged in direction from v1 but now performance-justified rather than sync-justified:** stay on Dolt everywhere for now, because the SQL-backend alternative doesn't ship yet; when it does, prioritize SQLite for ephemeral single-writer pool-worker rigs (not because sync doesn't matter there — it never mattered — but because per-op subprocess-to-managed-server round trips are the dominant cost for a scope that lives seconds to minutes); leave HQ on Dolt, since Dolt's SQL-server mode already gives HQ what a hypothetical Postgres migration would buy (a concurrent-writer server), at zero additional operational cost.

## 1. Verifying the "sync is dormant" premise

I ran this fleet's own checks rather than trusting the HQ sweep verbatim:

- `gc doctor` on this city (gasland) reports `jsonl-archive — local-only mode — commits stay on this host, off-box backup disabled`, and `bd-backup-freshness` shows the last `bd backup sync` for both `rigs/gascity` and `rigs/gascity-packs` was **813h33m (~34 days) ago** — the backup pipeline, a mechanism distinct from Dolt remote sync, is stale exactly as HQ's sweep described.
- `.beads/config.yaml` at both the city root and the `gascity` rig carries only `gc.endpoint_origin` / `gc.endpoint_status` / `dolt.auto-start` / `export.auto` / `backup.enabled` keys — **no `sync.remote`, no push/pull config**, at either level.
- The gascity rig's own git remotes (`origin` = `https://github.com/jeffora/gascity.git`, a personal fork; `upstream` = `git@github.com:gastownhall/gascity.git`, the project this fork tracks) — `git ls-remote origin 'refs/dolt/*'` returns **nothing**: no `refs/dolt/data` on this rig's actual push/pull target. (`upstream` does carry a `refs/dolt/data` ref, but that's the upstream *project's own* beads tracking of its own issues — unrelated to this fleet's local store — so it doesn't contradict the finding; it's a different repository's data, not this rig's.)
- The city root (`/Users/jeff.burn/Code/gasland`) is **not a git repository at all** (`fatal: not a git repository`), so there is no git remote at that level for `refs/dolt/data` to live on regardless.
- `ps aux` shows **four independent `dolt sql-server` processes**, one per city (`gasland`, `darujhistan`, `unta`, `raraku`), all bound to `localhost` — confirming the single-host, one-server-per-city topology HQ described, with no evidence of any of them talking to a remote.

Verdict: confirmed. Nothing here contradicts HQ's sweep; if anything it's more specific (this rig's actual push/pull remote has zero Dolt refs, vs. a stale unrelated ref on `upstream`). **Do not weight cross-machine sync in this decision.**

## 2. Performance — measured, not estimated, where feasible

**Managed-server overhead (measured):** the four running `dolt sql-server` processes on this host show steady-state RSS of 330–731 MB each (`ps -o rss`: 330144 KB / 461600 KB / 716096 KB / 730784 KB), with uptimes from 3 hours to 6 days, and CPU ranging 0%–92% depending on activity at sample time. That's 1.4–2.9 GB of resident memory across the fleet just to keep four issue-tracker backends warm, before any application workload runs on top.

**Per-op latency (measured on this city, warm server, 3–5 runs each):**

| Operation | `bd` command | Wall time |
|---|---|---|
| Read (list) | `bd list --status=open --limit 5` | 0.09–0.11s |
| Read (show) | `bd show <id>` | 0.09–0.19s |
| Read (ready) | `bd ready` | 0.09–0.10s |
| Write pair (create + close) | `bd q "..." && bd close <id>` | 0.29–0.31s (~0.15s/write) |

Every one of these numbers is dominated by **process-spawn + subprocess-to-managed-server round trip**, not raw SQL execution — this is the CLI path, which is what most `gc`/`bd` invocations outside the controller's in-process fast path actually take. Notably: `bd remember` on this fleet already carries a live incident memory (`beads-version-compat-warn-bd-cli-vs-linked-lib`) describing exactly this cost category — a `bd` CLI/library version skew (`v1.0.5` binary vs `v1.1.0` linked library) previously tripped the native-store `version_compat` preflight gate and **silently downgraded RIG-scope stores from the in-process fast path to per-call `bd` subprocess**, for every `gc`/`bd` op on the affected rigs, with the memory noting "NO measured latency impact under warm Dolt (~0.57s)" at the time — i.e., this fleet has already experienced, and instrumented, exactly the perf-relevant control-plane/subprocess-fallback failure mode this research is meant to reason about. That skew is currently resolved (`bd --version` now reports 1.1.0 matching the linked library), but it is evidence that the in-process-vs-subprocess boundary — not the SQL-engine choice — has already been the fleet's actual observed perf lever.

**SQLite proxy benchmark (explicitly a proxy, not a real bd-on-sqlite measurement — bd's SQLite backend does not exist in the installed release, so I could not benchmark it directly):** a raw `sqlite3` CLI insert+update pair against a throwaway file (`/tmp/bench.sqlite`, two separate `sqlite3` process invocations, no server) measured **0.02s per insert+update pair**, vs. 0.15s/write for the equivalent `bd` operation against the managed Dolt server — roughly **7–15x**, even though the SQLite measurement pessimistically pays for two CLI process spawns and the Dolt measurement is against an already-warm server. The gap is structural: an in-process/embedded single-file engine has no network round trip and none of Dolt's versioned-storage (Noms-style) write amplification; a client-server round trip to a managed SQL process always pays connection + protocol + MVCC-commit overhead on top of the actual write. Treat the specific multiplier as illustrative, not a validated bd-vs-bd number — but the direction (embedded beats client-server for single-writer, low-volume ops) is not in question.

**If a full apples-to-apples benchmark isn't feasible:** it isn't, because the only backend the installed `bd` supports is Dolt — there is no `--backend=sqlite`/`postgres` path to init a real comparison store with the current binary. The estimate above is the best-grounded substitute: measured Dolt latency, a labeled proxy for the embedded case, and the fleet's own prior incident data point (subprocess-fallback added no *additional* measurable latency over the already-slow warm-Dolt baseline, meaning the "slow" number here — ~0.1-0.3s per bd invocation — is already close to Dolt's floor, not an artifact of some avoidable extra hop).

## 3. Local concurrent-writer behavior — the real remaining risk

This is the axis that actually matters now that sync is off the table, and it cuts differently for HQ vs. ephemeral rigs:

- **HQ's store is a genuine multi-writer environment.** Per this rig's own architecture docs (`AGENTS.md`), the controller, the control dispatcher, and N pool workers all read/write the same city/rig beads store concurrently — reconciliation reads, session-lifecycle writes, hook claims. A `dolt sql-server` is a real SQL server with normal client-connection concurrency (confirmed running and reachable in `gc doctor`: `dolt-server — reachable on 127.0.0.1:53577`) — **it already provides what a hypothetical move to Postgres would provide: a server all writers connect to, with MVCC-style concurrent-writer semantics.** There is no concurrency deficiency in the current Dolt-server setup that Postgres would specifically fix; the risk v1 implied ("Dolt's locking wrong for HQ") doesn't match how HQ's Dolt is actually deployed (server mode, not embedded single-file).
- **The SQLite single-writer-lock risk is real but currently moot in practice**, since the SQLite backend isn't shippable with the installed beads release at all. If/when it lands, the risk is specifically for scopes with concurrent writers — i.e. **not** the ephemeral single-agent pool-worker rigs this report otherwise recommends SQLite for. A pool-worker rig like the one processing this bead has exactly one writer (itself) for its lifetime; SQLite's single-writer lock is a non-issue there by construction. It would only matter if SQLite were ever pointed at a *shared* scope with multiple concurrent writers — which is precisely the HQ scope this report says should stay on a server backend regardless.
- Net: **the axis v1 buried under the sync argument doesn't actually change the recommendation** — it reinforces the same split (single-writer ephemeral → embedded/lightweight; shared multi-writer → server-backed), just for a different reason (perf/lock-contention fit, not sync-capability fit).

## 4. Local Dolt-dependency gaps — closed, both green

Grepped `internal/beads` and all callers (`cmd/gc/*.go`, `internal/api/*`, `internal/dispatch/*`) directly rather than relying on v1's "not audited, follow up" note:

- **No runtime dependency on Dolt history/time-travel.** Every hit for `History(`/`history` in non-test source is either agent/session-turn history (`internal/api/handler_maintenance.go:44`, `handler_session_stream.go`, `huma_handlers_sessions_stream.go:44` — all `worker` session handles, unrelated to the beads store) or the literal string "dolt log" used as an error-log-file label in the managed-server watchdog (`cmd/gc/dolt_scope_watchdog.go:193`, `dolt_start_managed.go:1174`) — not Dolt's `AS OF`/time-travel SQL feature. Zero hits for `AS OF`, `DoltHistory`, or a `SupportsHistory`-style capability check. **Green light: nothing depends on local Dolt versioning.**
- **Worktree isolation is purely git/filesystem-based, not Dolt-branch-based.** `.gc/worktrees/<rig>/<name>` is a real `git worktree` (`cmd/gc/bead_worktree_reaper.go:18-49`, `internal/git/git.go` `WorktreeList`/`WorktreeRemove`/`WorktreePrune`, `internal/runtime/t3bridge/provider.go:1065-1085` for creation via `git.createWorktree`). Zero hits for a Dolt-branch-per-worktree pattern anywhere in the source. **Green light: nothing depends on local Dolt branching.**
- **Store-opener reality check:** only `OpenNativeDoltStoreAt` (`internal/beads/native_dolt_store.go:188`, called from `cmd/gc/api_state.go:365` and `cmd/gc/main.go:1226`) exists in current gascity source. This confirms v1's PR-diff read was accurate as a description of what #4154 *would* change, but underscores that as of today the codebase is 100% Dolt-opinionated at this call site — consistent with, not contradicting, v1's own "PRs are open, not merged" framing.

## 5. Per-scope fit on this fleet's actual shape

| Scope | Writers | Sync need | Recommended backend (once SQL backends actually ship) |
|---|---|---|---|
| Ephemeral pool-worker rig (this bead's own session) | 1 (itself, for seconds–minutes) | None (confirmed dormant) | SQLite — no server to keep warm, no RSS footprint, no lock contention (single writer by construction) |
| CI / throwaway sandbox | 1, short-lived | None | SQLite, same reasoning |
| HQ / shared city store | Controller + dispatcher + N pool workers, concurrently, indefinitely | None (confirmed dormant) | Stay on Dolt server mode — it already is a concurrent-writer server; a Postgres move buys nothing HQ doesn't already have, at real migration cost |
| Product rig with a single long-lived maintainer session | Effectively 1, but long-lived | None | Dolt is fine (current default); SQLite is a plausible future downgrade but lower priority than the ephemeral-rig case since the RSS/server cost is amortized over a longer session |

The mixed outcome is intentional, per the task's own framing, and the justification for each cell is now perf/concurrency-shaped rather than capability-shaped: nothing here is graded on "does this backend support labels/deps/audit trail" (they all do, identically, per beads' `issueops` core) — it's graded on writer count and server-overhead amortization.

## 6. Migration and rollout reality (carried forward from v1, still applicable, one addition)

- `bd export`/`import` (JSONL) remains upsert-only and cannot represent deletes — unchanged from v1's finding, still the blocking gap for any live cutover of an existing scope's data.
- #4736 (beads)/#4154 (gascity) sequencing is unchanged, but **the more important open item is that #4736 is not just "pinned to a focused unreleased commit" — the installed release (1.1.0) actively describes SQLite as removed, past tense**, which reads as beads having *had* an older SQLite path that was deprecated before the new conformance-tested multi-backend core (that #4736 exposes via `OpenConfigured`) existed. Before treating "adopt SQLite for ephemeral rigs" as actionable, confirm with the beads maintainers which of these is true: (a) #4736's SQL-backend core is a from-scratch replacement for the removed legacy SQLite path and is on track to ship in a future release, or (b) the removal notice and #4736 are unrelated efforts that haven't been reconciled. This wasn't resolvable from this fleet's read-only vantage point and materially changes the timeline for every recommendation above that assumes SQLite becomes available.
- If Postgres is ever pursued for HQ specifically (this report does not recommend that now, given Dolt-server-mode already provides concurrent-writer semantics at zero extra cost): the credential story flagged in v1 stands unresolved — supervisor-managed processes need a non-interactive credential path (`_PASSWORD_COMMAND` / `BEADS_POSTGRES_URL` ladder), and no such wiring exists in this fleet today.

## Recommendation (final)

1. **Do not move any scope off Dolt today.** The SQL-backend alternative v1 evaluated does not exist in the installed beads release; `--backend=sqlite`/`postgres` both hard-error on the version this fleet actually runs.
2. **Drop cross-machine sync entirely from the decision criteria going forward.** It's confirmed dormant fleet-wide (this rig's own remote carries no `refs/dolt/data`; no `sync.remote` config; single-host, four-city topology with no second machine to sync to). Any future re-litigation of this topic should not resurrect the sync argument without first checking whether that's changed.
3. **When/if a released beads version actually ships working SQLite support**, adopt it first for ephemeral single-writer pool-worker rigs (perf/RSS win, zero concurrency risk given single-writer-by-construction) — not for HQ, which should stay on Dolt's server mode since it already delivers concurrent-writer semantics a Postgres migration wouldn't meaningfully improve on.
4. **Before committing to any timeline, get a direct answer from beads maintainers** on whether "SQLite backend removed" (current release note) and "SQL backend core via OpenConfigured" (#4736, pinned unreleased commit) are the same effort at different maturity or two unreconciled tracks — this determines whether "adopt SQLite for ephemeral rigs" is a near-term or speculative recommendation.
5. **The local-dependency gaps are closed, both green:** no runtime read of Dolt history/time-travel, and worktree isolation is git-based, not Dolt-branch-based. Neither blocks a future backend change for any scope.

## Sources

- This session's direct verification: `gc doctor` (gasland city), `.beads/config.yaml` (city root + gascity rig), `git remote -v` + `git ls-remote origin/upstream 'refs/dolt/*'` (gascity rig), `ps aux | grep dolt` (all 4 cities), `ps -o rss,vsz,%cpu,%mem` per Dolt server PID, `bd --version`, `bd init --backend=sqlite|postgres` (error output), timed `bd list`/`show`/`ready`/`q`+`close` (this city), a labeled `sqlite3` CLI proxy benchmark (`/tmp/bench.sqlite`, deleted after use).
- `bd recall beads-version-compat-warn-bd-cli-vs-linked-lib` (fleet persistent memory) — prior version-skew incident and its "~0.57s warm Dolt" latency note.
- Source grep of `internal/beads`, `cmd/gc`, `internal/api`, `internal/dispatch`, `internal/git`, `internal/runtime/t3bridge` (this session, via a read-only exploration pass) — history/time-travel usage, worktree isolation mechanism, store-opener call sites.
- v1 report (`beads-backend-migration-report.md`, this directory) — retained for diff; its PR/diff descriptions of #4736/#4154 and the migration-tooling gap (JSONL upsert-only) are reused here as still-accurate, and are not re-verified independently in this pass.
