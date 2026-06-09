# Design Review: 08-runtime-state-migration

Reviewer: Yelena Markovic persona, Codex lane
Verdict: Block

Reviewed `.gc/design-reviews/ga-1ekw9l/attempt-6/design-before.md` for Maintenance-to-Core runtime state migration, JSONL/archive state, spawn-storm ledgers, non-destructive markers, old/new binary concurrency, and downgrade continuity. I did not read the Claude review output for this item before writing this review.

## Findings

### Blocker: The migration table is not yet an implementable old/new path ledger

The plan says runtime-state migration moves JSONL archive state, spawn-storm ledgers, refs/remotes, escalation fields, pending archive push state, formula environment state, and order skip/tracking aliases under Core-owned paths (`design-before.md:318`). It also says the migration table is part of the implementation, not review prose (`design-before.md:465`).

However, the table still uses placeholders such as `legacy archive directories`, `Core archive directory`, `old throttle ledger`, `Core throttle ledger`, `old order keys`, and `stable order keys plus aliases` (`design-before.md:467`). Those are categories, not paths or keys. The current tree has concrete legacy state surfaces, for example `GC_PACK_STATE_DIR` defaulting to `.gc/runtime/packs/maintenance`, legacy JSONL archive fallback `.gc/jsonl-archive`, legacy JSONL state fallback `.gc/jsonl-export-state.json`, pack state `jsonl-export-state.json`, `pending_archive_push`, `consecutive_push_failures`, `push_failure_escalated`, `spawn-storm-counts.json`, and `[orders].skip`. The plan does not map those exact paths/keys to exact Core-owned destinations.

Without that ledger, implementers cannot prove half-copied archives are detected, order aliases suppress the same logical order, or escalation/pending-push fields survive rollback. It also leaves room for destination-absent copy logic that races two processes or treats a partially copied `.git` archive as complete.

Required change: replace the generic migration table cells with exact old path/key to exact new path/key rows. Include JSONL archive repo path, JSONL state file and each state key, archive refs/remotes, spawn-storm ledger path and legacy fallback path, order skip configuration key and any order-tracking bead labels/metadata, escalation markers, formula env state, marker path, marker schema, and per-row copy/merge/conflict rules.

### Blocker: Post-marker old-binary write detection is asserted but not specified

The plan says the migration marker records old-binary post-marker write detection, and that if an old binary writes after the marker the new binary reports version-skew and requires manual reconciliation or deterministic re-upgrade (`design-before.md:322`). The data section repeats that retained legacy state is ignored unless the marker or digest checks show conflict (`design-before.md:459`).

That is not enough to implement the safety contract. The plan does not say which retained legacy files are fingerprinted, how directory trees such as archive refs/remotes are compared, whether mtimes are trusted or content digests are required, when the new binary checks for divergence, or how a retained legacy write is distinguished from legacy state intentionally preserved before the marker. It also does not define whether pre-marker operation reads a union from old and new locations or blocks until doctor migration completes.

Required change: define the marker fields and detection algorithm. At minimum, record per legacy path/key source digest, destination digest, copy completion status, last observed legacy digest after commit, and conflict classification. Specify that new binaries check the retained legacy fingerprints before behavior-changing operations that use migrated state, and that any post-marker legacy delta blocks with a typed version-skew diagnostic until reconciliation or re-upgrade completes.

### Major: The required old/new/old round-trip test is too implicit

The test section covers staged publish failure injection, old-binary post-marker writes, push-cursor reconciliation, retained legacy state, and downgrade/manual recovery guidance (`design-before.md:537`). That is close, but it does not name the persona's full round-trip scenario: interrupted archive copy, in-flight push state, duplicate old/new writers, and rollback to an older binary.

Required change: add one explicit round-trip test row for a city with a legacy Maintenance JSONL archive, origin refs/remotes, `pending_archive_push=true`, a spawn-storm ledger, order skips/tracking aliases, and escalation markers. The test should interrupt archive copy, resume migration, simulate an old binary writing the legacy state after the marker, run a new binary that blocks with version-skew, then roll back to an older binary or follow the documented downgrade path without deleting retained legacy state or duplicating pending pushes.

## What Passes

The direction is correct: migration is doctor-owned, controller/API/reload paths detect and refuse rather than mutate, fixes hold the city advisory lock, and stale Maintenance/Gastown runtime directories are retained instead of deleted (`design-before.md:311`, `design-before.md:482`). The plan also correctly names the critical state categories and avoids silent fallback after conflict.

The remaining blockers are about making that policy executable. Once the plan names exact state paths/keys, defines post-marker divergence detection, and pins the round-trip race/downgrade test, this persona should pass.
