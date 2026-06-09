# Ravi Krishnamurthy

**Verdict:** block

**Top strengths:**
- The design recognizes the active worker-boundary migration and explicitly says new session commands do not let production `cmd/gc` bypass `worker.Handle`.
- The per-slice coexistence table is the right shape: converted callers, legacy callers, guard updates, bake rules, and revert rules are separated by key family instead of hidden behind a broad facade.
- The design now treats one-writer proof, source-complete inventory, guard tightening, parity fixtures, and rollback metadata as closure conditions rather than cleanup work.

**Critical risks:**
- [Blocker] Slice 0 is declared as the only schedulable implementation item, but it is not integrated into the actionable backlog. The document has a "Slice 0 backlog item" in the Attempt 10 response, then later `TR-006: Start With Target Classification` and the `Backlog` section begin with `### 1. Target Classification`. A decomposition pass can still read the lower backlog and schedule Slice 1 first. For migration sequencing, that is a hard failure: the source inventory, guard, metadata validator, diagnostics baseline, and proof transcript must be an explicit `Backlog 0` item with every later slice depending on it.
- [Blocker] The per-key owner matrix and `session_design.*` metadata are required, but there is no machine-checkable schema or close validator. The Attempt 10 table says to "add a slice-close validator or workflow finalizer", but the proof command and artifact list do not name one. Without a validator, a slice can close while old and new writers still mutate the same key family with different validation. The current worker-boundary guard only catches a small set of `cmd/gc` bypass strings; it does not enforce the broader mutation boundary across API, repair, patch-map, generic bridge, and direct store writes.
- [Major] Slice 1 target classification still has hidden dependencies on mutation slices. The design says the raw classifier is read-only, but the first adopting API/Huma surface includes configured named lookup, materialization policy, rejected-by-config handling, and reserved-named suspend behavior. Those operations touch create/start, identity, repair, and close/retire routes. Slice 1 cannot safely adopt materializing API paths unless the materialization command already exists or the slice is explicitly limited to read-only classification and non-materializing adapters.
- [Major] Rollback is route-focused but not yet data-direction-complete. Several planned fields are new durable facts (`work_release_pending`, close generation, release identity snapshot, command conflict vocabulary, runtime intent fields). The design says rollback must prove old and new metadata tolerance, but the per-slice table does not require fixtures that run old-path readers against new metadata and new-path readers against legacy metadata. That leaves a partial-adoption window where revert restores the route but leaves state that the restored route misreads.
- [Major] The "legacy callers allowed during bake" column is too coarse for broad rows such as W-026, W-029, W-030, and W-031. Those rows explicitly require symbol-level splitting before source-complete status, but the bake table still allows surface-level language. A slice can appear revertible while a helper or generic bridge under the same broad row continues to write owned keys outside the command.

**Missing evidence:**
- An explicit `Backlog 0: Slice 0 Preflight` section with artifacts, proof commands, and dependency rules for every later backlog item.
- A schema and validator for `session_design.*` bead metadata and the per-key owner matrix.
- A proof command for the slice-close validator, including failure fixtures for missing converted callers, stale legacy rows, missing rollback paths, and duplicate writers.
- A surface-by-surface API/Huma migration plan that separates read-only classification from materialization, repair, wake, suspend, close, and presentation updates.
- Rollback compatibility fixtures showing old code tolerates new metadata and new code tolerates legacy metadata for each key family.
- A source-split inventory for broad rows before they are used as bake allowances.

**Required changes:**
- Add `### 0. Slice 0 Preflight` to the main `Backlog` and change `TR-006` from "Start With Target Classification" to "Target Classification Is The First Behavior Slice After Slice 0".
- Define a machine-readable slice metadata artifact, for example `internal/session/SESSION_MIGRATION_SLICES.yaml`, with required fields for converted callers, untouched callers, allowed legacy rows, retired rows, guard deltas, owned keys, active/provisional vocabulary, rollback route, data compatibility, and proof commands.
- Add a validator test or workflow finalizer to the Slice 0 proof command. It should fail when any implementation bead lacks required `session_design.*` metadata, cites an unsplit broad inventory row, omits one-writer proof, or has no rollback/data-compatibility fixture.
- Split Slice 1 into read-only candidate classification plus per-surface adapters. Any API/Huma path that materializes, repairs, suspends, closes, or updates presentation should either stay legacy or depend on the relevant command slice.
- Replace broad bake allowances with inventory-child IDs before implementation. W-026, W-029, W-030, and W-031 should not be acceptable as closure evidence until they are split by path, symbol, key family, target-bead proof, and expiry.
- Add data-direction rollback fixtures for each key family. Revert proof must include route restoration and persisted metadata compatibility, not only "switch adapter back to old resolver."

**Questions:**
- Which command or workflow component will enforce `session_design.*` metadata before a slice can close?
- Is the API/Huma target-classification first surface meant to be read-only only, or does it include configured-named materialization and reserved-name suspend paths?
- Are Slice 0 dependencies encoded in beads, or is the current "no later slice before Slice 0" rule only prose in `DESIGN.md`?
