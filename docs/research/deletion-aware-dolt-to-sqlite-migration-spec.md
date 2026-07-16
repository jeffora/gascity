---
title: Deletion-Aware Dolt→SQLite Migration for Existing Stores — Spec/Design
date: 2026-07-16
scope: gastownhall/beads (rigs/beads, HEAD 1d959d34e9432eb9ec7064d85cab366667236c54) + gascity source (rigs/gascity, HEAD 7ebadf76bf486848cc44407fe2f33da13ff1c35b)
supersedes-context: rigs/gascity/docs/research/beads-backend-migration-report-v2.md (be-yht/be-bc6), rigs/beads/docs/research/uniformity-and-cross-scope-backend-fleet.md (be-bc6, "existing-store migration is unsolved" carry-forward)
---

# Deletion-Aware Dolt→SQLite Migration for Existing Stores — Spec/Design

**Follow-on to:** be-bc6 (fleet decided: uniform SQLite going forward). This closes the one blocker be-bc6 explicitly left open: safely migrating **existing** Dolt stores that have delete history, without resurrecting deleted beads.

**Method:** source reading with file:line citations in both repos at the SHAs above, plus a live proof-of-concept on scratch stores (never touching any real fleet store) that directly demonstrates both the resurrection failure mode and the fix. All commands and output below are reproducible; scratch directories were deleted after the session.

---

## Recommendation up front

**Phased both, per be-bc6's own framing — but with one correction to be-bc6's stated risk, found by hands-on testing:**

A **same-instant** `bd export` (of an existing Dolt store's *current* live state) immediately followed by `bd import` into a fresh SQLite store does **not** resurrect deleted beads — I verified this directly (§0). `bd export` queries only live rows (`cmd/bd/export.go`), so a bead already deleted before the export runs is simply absent from the export; there is nothing to resurrect. **The real hazard is a *stale* export** — any JSONL snapshot generated *before* a delete that gets imported *after* that delete — which I also reproduced directly and confirmed resurrects the deleted bead exactly as be-bc6's report predicted, just via a narrower mechanism than "export→import is inherently unsafe."

Given that correction, the recommendation is:

1. **Now, to unblock cutover — (b) out-of-band reconciliation in gascity**, built on a Dolt-native technique this session validated works cleanly: enumerate the authoritative deleted-set directly from the source Dolt store's own `dolt_history_<table>` system table (every row version that ever existed, across all commits) minus the live table, then apply that as a **post-import cleanup delete** against the new SQLite store, gated behind a verification step, before cutover. No upstream dependency; small, testable gascity/scripts-level tooling.
2. **Durable fix, upstream (a), lower priority** — propose a narrowly-scoped `bd export --deleted-since=<ref>` (or an equivalent read-only diff-based command) that wraps the same `dolt_history` technique inside beads itself, so future migrations don't need hand-rolled SQL against the Dolt server. This is **not** a request to bring back the old tombstone/soft-delete system — that system was deliberately removed (`CHANGELOG.md`, "Phase 4: Remove tombstone/soft-delete system", part of the 8-phase Dolt-native cleanup) and reintroducing live tombstone rows would cut against that decision. A point-in-time, read-only, Dolt-only diff export is a different and much smaller ask, because it only needs to work on the one backend that actually has the history to answer the question.
3. **Procedural guardrail, no code required, apply immediately regardless of (1)/(2):** never migrate from a pre-existing/cached export file. Always generate the export used for cutover at the moment of cutover, from the live source store, and treat any older `.beads/issues.jsonl` (including the git-committed one) as unsafe input for a migration import.

---

## 0. Proof-of-concept: reproducing and fixing the resurrection failure mode

Built `bd` from `rigs/beads` at HEAD (`CGO_ENABLED=0 go build ./cmd/bd`, pure-Go, 103MB binary — confirms the pure-Go SQLite/Postgres/MySQL backends need no CGO, per be-bc6 §Q2). All commands below ran under `env -i PATH="$PATH" HOME="$HOME" ...` to strip this session's ambient `BEADS_*` env vars pointed at the real `rigs/beads` store (confirmed necessary: a naive `bd init` in a scratch dir without `env -i` found and refused to touch the real Dolt server — the guard rail worked as intended). A scratch `dolt sql-server` on `127.0.0.1:13399` backed a throwaway `--backend=dolt --server` project under this session's scratchpad; everything was torn down afterward and no fleet store was touched.

**Step 1 — created 4 issues, deleted one (`poc-th1`) immediately:**
```
$ bd delete poc-th1 --force
✓ Deleted poc-th1
```
Confirmed in `dolt_log`: the delete is its own ordinary Dolt commit, message `bd: delete poc-th1` (sibling to `bd: create poc-th1`, `bd init`, and the schema-migration commits) — deletion is not a special code path from Dolt's own versioning perspective, it's a normal committed transaction (matches §3 below).

**Step 2 — created a 5th issue (`poc-p1a`), exported while it still existed** (`stale-export.jsonl`, 3 issues: `poc-5gf`, `poc-ccd`, `poc-p1a`), **then deleted it from the live store**, then took a second, fresh export (`fresh-export.jsonl`, 2 issues: `poc-5gf`, `poc-ccd` — `poc-p1a` correctly absent).

**Step 3 — imported each into its own fresh SQLite store:**

| Import source | Result |
|---|---|
| `fresh-export.jsonl` (taken *after* the delete) | 2 issues — **no resurrection** |
| `stale-export.jsonl` (taken *before* the delete) | **3 issues — `poc-p1a` resurrected** |

This directly falsifies the literal framing in be-bc6's carry-forward note ("a naive `bd export` → `bd import`... resurrects every bead ever deleted") for the same-instant case, and directly confirms the real, narrower mechanism: **staleness of the export relative to the delete, not the export/import round-trip itself, is what causes resurrection.**

**Step 4 — enumerated the authoritative deleted-set directly from Dolt, no beads changes needed:**
```sql
SELECT DISTINCT h.id FROM dolt_history_issues h
WHERE h.id NOT IN (SELECT id FROM issues);
-- id
-- poc-p1a
-- poc-th1
```
`dolt_history_<table>` is Dolt's built-in system table exposing every historical row version across the table's *entire* commit history in one query — it correctly caught both deletions (`poc-th1`, deleted long before either export, and `poc-p1a`, deleted between the two exports) with no need to know which commits to diff. I initially tried a plain `dolt_diff(<baseline>, 'HEAD', 'issues')` between two single snapshots and found it **does not** work for this purpose in general: a row created and deleted entirely *within* the diffed range nets to zero difference and is invisible to a two-point diff. `dolt_history_<table>` avoids that trap because it doesn't diff two snapshots, it aggregates every version ever committed.

**Step 5 — applied the deleted-set to the resurrected store and verified convergence:**
```
$ bd delete poc-p1a --force        # against the SQLite store imported from the stale export
✓ Deleted poc-p1a
```
Diffed the resulting store's live ID set against the cleanly-imported (`fresh-export.jsonl`) store's ID set: **identical, zero difference.** Reconciliation converges the resurrected store to the correct one.

**Step 6 — idempotency check, and an important operational finding:** re-running `bd delete poc-p1a --force` a second time (after it's already gone) returned `Error: issue poc-p1a not found`, exit code 1. **`bd delete` is not idempotent on its own** — it hard-fails on any ID not present, and in batch form (`bd delete id1 id2 --force`) a single missing ID aborts the *entire* batch, including IDs that *do* exist and should have been deleted (reproduced directly: `bd delete poc-th1 poc-p1a --force` against the resurrected store errored on `poc-th1`, which was never in that store to begin with, and left `poc-p1a` — which *was* present and should have been deleted — untouched). **Any reconciliation tool must pre-filter the deleted-set against the target store's current live IDs before calling `bd delete`, and must treat "not found" as success, not failure, when re-run.** This is a concrete implementation requirement for §2, found only by testing, not documented anywhere in `bd delete --help`.

---

## 1. Approach comparison + recommendation (detail)

### (a) Deletion-aware export upstream in beads

**Current export/import source, read directly:**
- `cmd/bd/export.go`: `runExport` builds an `IssueFilter` and calls `store.SearchIssues(ctx, "", filter)` against the **live** `issues` table (confirmed no historical/`AS OF` query anywhere in this file) — a full re-materialization of current state on every run, not an append-only change log. No `--include-deleted`/`--deleted` flag exists (checked `--help` and grepped `IssueFilter`'s fields).
- `cmd/bd/import.go`: doc comment states "upsert semantics" (line ~27); `runImportFromReader` explicitly skips any line with `status == "tombstone"` (line ~244) — **dead compatibility code**, not an active tombstone mechanism: this exists purely to silently drop old-format tombstone lines from **pre-v0.50** exports, confirmed by the comment in `import_shared.go` ("deleted issues exported by older versions (pre-v0.50)... not valid for re-import"). Grepping all of `internal/storage` for a `DeleteIssue`/`DeleteIssues` call from the import path found none — import never deletes.
- `CHANGELOG.md` confirms tombstones **used to exist** (multiple entries: "Tombstones in JSONL export (GH#696)", "Auto-import tombstone handling", "Tombstone wins over closed in merge conflict resolution") and were **deliberately removed**: the `[0.51.0]` "Dolt-native cleanup (dolt-1s40)" 8-phase refactor lists "Phase 4: Remove tombstone/soft-delete system" alongside "Phase 5: Remove JSONL sync layer" and "Phase 6: Remove SQLite backend entirely." `cmd/bd/info.go` still documents this removal in its own version-info string.
- **Implication:** the SQLite/Postgres/MySQL backends present today (per be-bc6, reintroduced via PR #4601, "choose your own storage backend") were rebuilt on the shared `sqlkit`/`issueops` core *without* resurrecting the old tombstone system. Any upstream ask that reintroduces a persistent, live tombstone row (a `status="tombstone"` row that lives in the table indefinitely) is asking beads to reverse a recent, deliberate architectural decision, and should not be the shape of the ask.

**What a minimal, non-controversial upstream change would actually look like:** not a tombstone system, but a **read-only, point-in-time diff export**, e.g. `bd export --deleted-since=<dolt-ref>` (Dolt backend only — SQL family backends have no history to diff, so this flag would correctly no-op or error on non-Dolt backends). Internally it is exactly the query validated in §0 Step 4: `SELECT DISTINCT id FROM dolt_history_<table> WHERE id NOT IN (SELECT id FROM <table>)`, optionally scoped to rows whose last-seen version postdates `<dolt-ref>`, emitted as its own JSONL record type (e.g. `{"_type":"deleted","id":"..."}`) that `bd import` would learn to apply as an actual delete (not skip, as it does for the old `tombstone` status line today). This is additive, Dolt-only, and does not touch the SQL-family backends' code paths at all — the smallest change that closes the gap durably. I have not opened or drafted this as an actual PR; this section is a recommendation for what to propose, not a claim that it exists or is planned.

### (b) Out-of-band reconciliation in gascity

This is what §0 validated end-to-end and what should be built first, because it requires no upstream change and the mechanism is now proven:

1. Enumerate the deleted-set directly from the source Dolt store via `dolt_history_<table>` (§0 Step 4) — works today, no beads changes needed, and is more robust than pairwise `dolt_diff` (§0 Step 4's negative finding: a create+delete that both happen inside one diffed range is invisible to a two-point `dolt_diff`, but is caught correctly by `dolt_history`).
2. Run the normal `bd export` (fresh, at cutover time) → `bd import` into the new SQLite store.
3. Filter the deleted-set to IDs that actually exist in the new store's live table (§0 Step 6 finding — required for idempotency and to avoid `bd delete`'s all-or-nothing batch failure on any missing ID), then run `bd delete <filtered-ids> --force` against the new store.
4. Verify (§3) before cutover.

This is gascity-side tooling (a script or a small `internal/beads`-adjacent helper, not a beads core change) that talks to the Dolt server's SQL endpoint directly for step 1 (the same `mysql`/Dolt SQL connection any `bd`-adjacent tool already uses) and shells out to the built `bd` binary for steps 2–3.

### Recommendation: phased both, (b) first

Matches be-bc6's own suggestion ("phased both... (b) now to unblock, (a) upstreamed as the durable fix") — this spec adds the concrete, tested mechanism for (b) and narrows what (a) should actually ask for, given the tombstone-removal history.

---

## 2. Mechanism detail for the recommended path

**Where deletions are recorded in Dolt today:** no soft-delete column, no tombstone row, no separate deletion-event table. `internal/storage/issueops/delete.go`'s `deleteIssueRowInTx` issues a plain `DELETE FROM issues WHERE id = ?` (hard row delete), used identically by both the Dolt backend (`internal/storage/dolt/issues.go`'s `DeleteIssue`/`DeleteIssues`, which additionally wraps the delete in `DOLT_ADD` + `DOLT_COMMIT`) and the SQLite backend (`internal/storage/sqlkit/issues.go`'s `DeleteIssue`, calling the *same* `issueops.DeleteIssueInTx`). Cascaded rows (`dependencies`, `labels`, `comments`, `events`, snapshots) are deleted too — confirmed directly in §0 Step 1 (`dolt_log` shows the delete as one commit, `bd: delete poc-th1`) — **the only reason a Dolt-backed deletion is recoverable at all is that Dolt commits it to version-controlled history; on SQLite, the identical hard-delete leaves zero trace once the transaction commits, because SQLite has no version control** (`docs/architecture/storage-backends.md`, confirmed by source: no `deleted_at`/`is_deleted` column found anywhere in `issueops/delete.go`, and the current schema migrations were not read line-by-line to positively rule out a dormant unused column — flagging this as the one thing not independently re-verified beyond the delete path itself).

**`bead.deleted` is not a beads concept at all — it's a gascity `CachingStore` cache-reconciliation event**, defined at `internal/events/events.go` and emitted from `internal/beads/caching_store_writes.go` (`c.notifyChange("bead.deleted", deleted)`). This matters for enumerating a deleted-set: **do not use the fleet event log (`/Users/jeff.burn/Code/gasland/.gc/events.jsonl`) as the authoritative source.** It is lossy by construction — it only reflects what a `CachingStore` reconcile loop *detected after the fact*, not a first-class signal from the storage layer itself. Direct evidence: all 348 `bead.deleted` events in that log carry `actor: "cache-reconcile"` (100%) and `subject` IDs prefixed `gld-` (the HQ/city scope) exclusively — zero events for the `beads`, `gascity`, or `gascity-packs` rig scopes. That is *not* proof those rigs have no delete history; it only proves the city-level event bus never saw any, which is exactly what "lossy" predicts for a rig-scoped Dolt store that isn't wired into a `CachingStore`'s reconcile loop the same way the city store is.

**Authoritative enumeration — use Dolt directly, not the event log:** the validated technique (§0 Step 4) is
```sql
SELECT DISTINCT h.id FROM dolt_history_<table> h
WHERE h.id NOT IN (SELECT id FROM <table>);
```
run against **each** table that has independent delete-relevant identity — at minimum `issues`; extend to `comments`/`labels`/`dependencies` only if reconciling those independently of their parent issue ever matters (in practice, cascade-deleting the parent issue in the new store, per §2's step 3, already removes its cascaded rows the same way the source deletion did, so a separate deleted-set for those child tables is very likely unnecessary — flagged as an assumption, not verified with its own PoC pass).

**Applying the deleted-set — idempotency and ordering, both validated in §0:**
- **Ordering:** must run *after* import (the rows have to exist in the new store to be deleted) and *before* the new store is put into live/production use (i.e., before any cutover flips a rig's `.beads/metadata.json` backend pointer) — otherwise a live write racing the reconciliation pass could recreate or touch a row mid-cleanup.
- **Idempotency is not free** — `bd delete` itself hard-fails (`Error: issue <id> not found`, non-zero exit) on any ID absent from the target, and in batch form one missing ID aborts deletion of every other ID in that same invocation, confirmed directly in §0 Step 6. The reconciliation tool must: (1) query the target store's current live IDs first (e.g. `bd list --json` or the equivalent SQL `SELECT id FROM issues`), (2) intersect with the Dolt-derived deleted-set, (3) call `bd delete` only on IDs that survive that intersection, and (4) treat an empty intersection as a successful no-op rather than skipping the step. This makes the *tool* idempotent even though the underlying `bd delete` command is not.

---

## 3. Verification / audit (the trust gate)

Concrete, human-runnable checks, in order:

1. **Live-row count reconciliation:** `SELECT COUNT(*) FROM issues` on the source Dolt store immediately before generating the cutover export must equal `SELECT COUNT(*) FROM issues` on the new store immediately after import + reconciliation delete. (§0's own numbers: source had 2 live issues after all test deletes; both the cleanly-imported store and the reconciled formerly-stale store converged to 2 — demonstrated directly, not hypothetical.)
2. **Deleted-set closure check:** re-run the `dolt_history_<table>` query from §2 against the source; for every ID it returns, assert `SELECT COUNT(*) FROM issues WHERE id = '<id>'` is `0` on the new store. This must return zero rows for every deleted ID — any hit is a reconciliation failure.
3. **Full-fidelity diff (the gate that must come back empty):** dump `id` (or `id` + `updated_at`, or a full-row hash for stronger fidelity) sorted by `id` from both stores' live `issues` tables and diff them. §0 Step 5 performed exactly this (`diff` on sorted ID lists between the reconciled store and a cleanly re-imported reference store) and got a clean, empty diff — that is the pass criterion. A non-empty diff at this stage means either the reconciliation missed an ID or the import/export step itself diverged, and cutover should not proceed until it's empty.
4. **Open/closed/deleted-set three-way reconciliation:** `(live IDs in new store) ∪ (deleted-set applied)` should equal `(live IDs in source store at export time) ∪ (pre-export historical deletes already excluded by export)` — in practice this collapses to checks 1–3 above; called out separately here only because the bead's own ask (§3 in the original assignment) named it as its own bullet.

## 4. Rollback

The source Dolt store is **read-only for the entire procedure** — nothing in §0's validated mechanism writes to the source: the `dolt_history_<table>` query is a `SELECT`, `bd export` is a read, and every write (`bd import`, the reconciliation `bd delete` calls) targets only the new SQLite store. This was true in every command run during the PoC — no command in §0 mutated the source Dolt project after the initial test-data setup phase.

**Abort procedure if any §3 check fails:** discard the new store entirely (`rm` the new store's directory/DB file — it is disposable, nothing points at it yet) and do not touch the rig's `.beads/metadata.json` backend pointer. The rig continues operating on its existing Dolt store, unchanged, exactly as before the migration attempt started. Re-run from a fresh export once the failure is understood; because the source was never written to, there is no state to "roll back" on that side — only the disposable target needs to be discarded and rebuilt.

## 5. Scope boundaries (blast radius)

**This only affects existing Dolt stores that actually have delete history — but the fleet event log cannot answer "which scopes those are" on its own**, per §2's finding that `/Users/jeff.burn/Code/gasland/.gc/events.jsonl`'s 348 `bead.deleted` events are 100% HQ/`gld`-scope and reflect only what a `CachingStore` reconcile loop happened to observe. To classify a scope before migrating it, run the §2 `dolt_history_<table>` query **directly against that scope's own Dolt store** (not the city event log):

```sql
SELECT COUNT(DISTINCT h.id) FROM dolt_history_issues h
WHERE h.id NOT IN (SELECT id FROM issues);
```

- **Result `0`:** trivial migration — plain `bd export` (fresh) → `bd import` is safe as-is, no reconciliation step needed.
- **Result `> 0`:** needs the full §1(b)/§2/§3 deletion-aware path before cutover.

**What this spec did *not* do:** run that query against the fleet's actual configured rig stores (`rigs/beads`, `rigs/gascity`, `rigs/gascity-packs`) or the HQ store itself — doing so means connecting to each rig's real Dolt server and was out of scope for a read-only spec pass (and this session's own scratch-store guard-rail experience in §0 is a direct demonstration of why that requires care: ambient `BEADS_*` env vars point at the real stores by default, and stripping them via `env -i` is mandatory before running any exploratory query against a scratch copy, to avoid an accidental query — even a read-only one — against a live server through the wrong environment). **Before cutting over any specific scope, run the one-line classification query above against that scope's real Dolt server first**, as the literal first step of that scope's migration, not as an assumption inherited from this document. The 348-event count from the city log is suggestive that HQ has real delete history and the three rig scopes might not — but per §2, "might not" is an artifact of event-log blindness, not a confirmed finding, and should not be treated as one.

---

## Summary of citations

| Claim | Evidence |
|---|---|
| Export queries live rows only, no historical/tombstone query | `rigs/beads/cmd/bd/export.go` (`runExport`, `SearchIssues` call); confirmed empirically in §0 Steps 2–3 |
| Import is upsert-only, never deletes, skips dead `tombstone` status | `rigs/beads/cmd/bd/import.go` (doc comment, `status == "tombstone"` skip); `import_shared.go` comment on pre-v0.50 tombstone format |
| Deletion = hard `DELETE FROM issues`, identical code path Dolt/SQLite | `rigs/beads/internal/storage/issueops/delete.go` (`deleteIssueRowInTx`); `internal/storage/dolt/issues.go` (`DeleteIssue`, wraps in `DOLT_COMMIT`); `internal/storage/sqlkit/issues.go` (`DeleteIssue`, same `issueops` call) |
| Tombstones existed, were deliberately removed | `rigs/beads/CHANGELOG.md` ("Phase 4: Remove tombstone/soft-delete system", `[0.51.0]`); `cmd/bd/info.go` version-info string |
| `bead.deleted` is a gascity cache-reconcile concept, not a beads/Dolt signal | `rigs/gascity/internal/events/events.go` (`BeadDeleted` constant); `internal/beads/caching_store_writes.go` (`notifyChange("bead.deleted", ...)`) |
| Event log is lossy / HQ-only for deletion signal | direct analysis of `/Users/jeff.burn/Code/gasland/.gc/events.jsonl`: 348 `bead.deleted` events, 100% `actor:"cache-reconcile"`, 100% `gld`-prefixed subjects, zero for configured rig scopes |
| Fresh export→import does not resurrect; stale export→import does | §0 Steps 2–3, reproducible PoC on scratch stores |
| `dolt_history_<table>` correctly enumerates the full historical deleted-set; two-point `dolt_diff` misses creates+deletes inside the diffed range | §0 Step 4, direct SQL run against a scratch Dolt server |
| Reconciliation converges a resurrected store to match a clean one | §0 Step 5, ID-set diff, empty result |
| `bd delete` is not idempotent and fails the whole batch on any missing ID | §0 Step 6, reproduced directly (`Error: issue poc-p1a not found`, batch abort on `poc-th1 poc-p1a`) |
