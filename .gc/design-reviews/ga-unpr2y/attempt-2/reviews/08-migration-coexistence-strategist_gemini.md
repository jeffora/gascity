# Ravi Krishnamurthy — DeepSeek V4 Flash (Independent Review, Attempt 2)

**Verdict:** block

**Review focus:** Cross-file consistency, missed edge cases, and assumptions the Claude and Codex reviews accepted too quickly. Evidence drawn from source verification against the design document and all attempt-2 reviews.

---

## Top Strengths

- The mutation boundary table at field-family granularity (`DESIGN.md:67–77`) is the right contract level. Named key families let a static guard check specific writes and let reviewers verify per-slice retirement — this is a genuine structural improvement over attempt 1.
- The scenario traceability matrix (`DESIGN.md:209–248`) is a material advance. Each slice carries its touched rows, current proof, and a freshness gate that can block extraction. This was the highest-severity gap in attempt 1 and it is now present in form.
- The command atomicity contract (`DESIGN.md:260–276`) with precondition/written-keys/ordering/conflict/retry columns gives implementers concrete terms. The stale-fact and recovery-authority columns are particularly useful — they turn vague "validate current state" prose into auditable contract terms.
- The reconciler fact contract (`DESIGN.md:278–295`) anchors on `ProjectLifecycle` and `ComputeAwakeSet`, which are already pure functions consuming precomputed inputs. This is the correct coexistence baseline.

---

## Critical Risks

### [Blocker] The close command writes three field families in separate store operations, but the per-slice ownership model assigns those families to different slices

`Manager.CloseDetailed` (`internal/session/manager.go:862–921`) acquires `withSessionMutationLock`, then performs three distinct store-write operations:

1. `m.sp.Stop(sessName)` — runtime side effect (no store write, but a side effect between lock acquisition and store writes)
2. `m.clearWakeAndHoldOverrides(id)` — writes `pin_awake`, `held_until`, `sleep_intent` (wake/hold/drain family, slice 4)
3. `m.retireConfiguredNamedSessionIdentifiers(id, b)` — writes `alias`, `session_name`, `session_name_explicit`, `alias_history`, `synced_at`, `held_until`, `quarantined_until`, `wait_hold`, `sleep_intent`, `sleep_reason`, `retired_named_identity` (identity family, slice 1/3, plus wake/hold/drain overlap)
4. `m.store.Close(id)` — bead close (lifecycle family, slice 3)

The mutation boundary assigns lifecycle close to slice 3, wake/hold/drain to slice 4, and identity retirement to slice 3. But a single `CloseDetailed` call writes all three families in three separate `SetMetadataBatch` calls inside one lock span. The design's command atomicity contract says "stop succeeds before bead close when stop is required; close metadata and bead close are treated as one command unit" — but `clearWakeAndHoldOverrides` and `retireConfiguredNamedSessionIdentifiers` happen between stop and close, and neither is mentioned in the close command cluster's "Written keys" column.

This is not just a documentation gap. During coexistence, if slice 3 extracts the close command but slice 4 has not yet extracted wake/hold/drain, the close command must still call `clearWakeAndHoldOverrides` — which means slice 3 temporarily owns a wake/hold/drain write. The design must either:

- (a) specify that each command may call other-field-family writes as long as those writes go through the owning slice's command API when available (creating a command-call graph across slices), or
- (b) specify that the close command absorbs these writes and they are extracted later when the owning slice is implemented, with a documented handoff plan.

The current design assigns one owner slice per row in the landscape table, but `CloseDetailed` proves that a single production operation can write keys from three different families. Without a cross-family write policy, the static guard will either block a legitimate close path or allow an undocumented cross-family write.

**Required change:** Add a cross-family write policy. For each command that writes keys from another slice's family, document: (1) which cross-family keys it writes, (2) whether it calls the owning slice's command or writes directly, and (3) the handoff condition (when the owning slice extracts its command, the cross-family write redirects through it). The close command cluster's "Written keys" column must list `pin_awake`, `held_until`, `sleep_intent`, and the identity retirement keys.

### [Blocker] The mutation landscape table assigns `cmd/gc/session_beads.go` to slice 3 only, but `closeFailedCreateBead` writes fields from slices 2, 3, and 4

`closeFailedCreateBead` (`cmd/gc/session_beads.go:1736–1756`) applies:

- `ClosePatch(now, "failed-create")` — writes `state`, `state_reason`, `closed_at` (lifecycle family, slice 3)
- `pending_create_claim: ""` — create/start family, slice 2
- `pending_create_started_at: ""` — create/start family, slice 2
- `sleep_intent: ""` — wake/hold/drain family, slice 4

Then it calls `store.Close(id)` (lifecycle) and `cancelStateAssignedToRetiredSessionBead` (identity). This single function writes fields from three families. The landscape table's "Owner slice" column says "3" for this entry, implying that slice 3 owns all field families written by `session_beads.go` close paths. This is incorrect for the create/start and wake/hold/drain fields.

This is the same structural issue as the `CloseDetailed` finding, but it also affects the static guard: if the guard enforces slice-3-only key writes for this file, it will block the legitimate clearing of `pending_create_claim` and `sleep_intent`. If it doesn't enforce per-key-family ownership, the landscape table is misleading about what the guard actually checks.

**Required change:** Either add a "field families written" column to the landscape table, or split `session_beads.go` into per-family entries. For `closeFailedCreateBead`, document that slice 3's close command absorbs create/start and wake/hold/drain clears as part of terminal cleanup, with a handoff to the owning slice's command when available.

### [Blocker] `withSessionMutationLock` is unexported and only protects `Manager` methods — legacy writers can interleave with commands during coexistence

The command atomicity contract says "lock-plus-reread at the command boundary unless the store grows native revision tokens." All `Manager` methods that mutate session state (CloseDetailed, Suspend, Archive, Quarantine, BeginDrain, etc.) acquire `withSessionMutationLock(id)`. But this lock is an unexported package-level mutex — it is invisible to `cmd/gc/` code.

During coexistence, both paths are active:

1. A `Manager.CloseDetailed` call acquires the lock, calls `loadSessionBead` (re-read), validates state, calls `Stop`, writes `clearWakeAndHoldOverrides` and `retireConfiguredNamedSessionIdentifiers`, then closes the bead.
2. A reconciler tick reads session state, decides to update `sleep_reason`, and calls `store.SetMetadataBatch(session.ID, batch)` without any lock.

If (2) runs between `loadSessionBead` and the first `SetMetadataBatch` in (1), the command's re-read is stale: the reconciler's `sleep_reason` write is overwritten by the command's subsequent batch write that doesn't include `sleep_reason` at all. This is a classic read-modify-write race that the lock does not prevent because only one side holds it.

The prior reviews noted that legacy writers bypass the lock, but none observed that the design's atomicity contract specifically claims "lock-plus-reread" as the compare-and-swap substitute. If the lock doesn't protect against the most common interleaving (reconciler writing between command re-read and command write), then the atomicity contract is not actually satisfied.

**Required change:** One of:
- (a) Export the mutation lock or wrap it in a command-scoped `ReconcileAndValidate` helper that both commands and legacy paths can acquire, making the lock boundary cover all writers.
- (b) Add a store-level compare-and-swap (e.g., revision token) that both commands and legacy paths must respect. This is what the design already mentions as the long-term path.
- (c) Document that during coexistence, the command boundary accepts stale-read races for fields the command does not overwrite, and specify which fields are safe to interleave (idempotent clears, monotonic timestamps) vs. which require strict ordering (state transitions).

Option (c) is likely the correct coexistence answer — many of the cross-family writes are idempotent clears that lose no information on interleaving — but it must be stated explicitly, not left as an implicit assumption.

---

## Major Risks

### [Major] The prepare-runtime-start phase is not atomic — it uses two individual `SetMetadata` calls, not one batch

`session_lifecycle_parallel.go:814` writes `session_key` via `store.SetMetadata(session.ID, "session_key", sessionKey)`, and `:885` writes `instance_token` via `store.SetMetadata(session.ID, "instance_token", instanceToken)`. These are individual calls, not a batch. A crash between them leaves the bead with `session_key` set but `instance_token` unset. The command atomicity contract's "Prepare runtime start" row says "Written keys: state=creating, generation, instance_token, pending_create_started_at, reset keys" but the prepare phase actually writes `session_key` and `instance_token` as individual metadata updates before the main batch.

The design says "commit metadata before provider start" for the prepare phase, but the actual implementation writes two separate keys before the batch. The `CommitStartedPatch` row says "one batch for state and pending-create clear" which is correct for the commit phase, but the prepare phase is not one batch.

**Required change:** Either batch the prepare-phase writes (`session_key` and `instance_token`) into a single `SetMetadataBatch` call, or document in the atomicity contract that `session_key` and `instance_token` are idempotent-precondition fields that can be safely partially written (a stale `session_key` without `instance_token` will be overwritten on the next prepare attempt).

### [Major] `last_woke_at` is written in the start path but belongs to the wake/hold/drain family

`session_lifecycle_parallel.go:1580` clears `last_woke_at` via `store.SetMetadata(session.ID, "last_woke_at", "")`. The design's mutation boundary table assigns `last_woke_at` to the wake/hold/drain family (slice 4), but this write happens in the start-path code (`session_lifecycle_parallel.go`), which the landscape assigns to slices 2, 5, and 6. This is a cross-family write from slice 2's code path that the landscape table doesn't capture.

The 01-mutation-boundary review noted that `last_woke_at` should be added to the create/start key family, but it didn't call out that the landscape table's owner-slice assignment for `session_lifecycle_parallel.go` (slices 2, 5, 6) would be wrong if `last_woke_at` stays in the wake/hold/drain family.

**Required change:** Add `last_woke_at` to the landscape entry for `session_lifecycle_parallel.go`, or move it to the create/start family if the start path is the only writer that clears it on start. If it stays in wake/hold/drain, document the cross-family write in the command atomicity contract's "Prepare runtime start" cluster.

### [Major] `RepairEmptyType` is a write side effect in the classifier path, violating the TR-002 requirement that fact construction must not mutate session metadata

`ResolveSessionBeadByExactID` (`internal/session/resolve.go:48–67`) calls `RepairEmptyType(store, &b)` which mutates the bead via `store.Update(b.ID, beads.UpdateOpts{Type: &t})`. This is position 1 in the precedence table — the first thing the classifier does when resolving an exact bead ID. If the classifier is supposed to be a pure read-only classification function (TR-002: "Fact construction does not mutate session metadata"), then `RepairEmptyType` is a write side effect in the read path.

The 05-target-identity-classifier review noted this as a question, but the implications are more serious than a question: if the classifier has write side effects, it cannot be called idempotently, cannot be called from read-only contexts (API inspect, CLI status), and cannot be called from multiple surfaces on the same bead without coordinating the repair. A classifier that mutates data on first encounter is a repair tool, not a classification tool.

**Required change:** Either (a) make `RepairEmptyType` a pre-classification adapter step called by each surface before invoking the classifier, with the classifier itself remaining pure; or (b) document that the classifier has a write side effect for empty-type beads and specify the concurrency semantics (idempotent write, acceptable in read-only paths). Option (a) is consistent with TR-002 and the design's stated read-only classifier contract.

### [Major] `UpdatePresentation` writes identity-family fields via `store.Update` but is not in the mutation landscape

`Manager.UpdatePresentation` (`internal/session/manager.go:1114–1167`) writes `alias`, `alias_history` (via `UpdatedAliasMetadata`), and uses `store.Update` with metadata. The landscape table only lists `SetMetadata`/`SetMetadataBatch`/`Close`/`Create` call sites. `store.Update` with a `Metadata` field is an equally valid mutation path for session-owned keys, but the static guard specification only mentions `SetMetadata`, `SetMetadataBatch`, `Update`, `Create`, or `Close` when writing session-owned keys.

Since `UpdatePresentation` is a `Manager` method (inside `internal/session`), it's technically inside the boundary. But the landscape table should still list it so reviewers can verify that the command boundary eventually absorbs it, and so the static guard can track it as an internal writer that should not be called from outside `internal/session`.

**Required change:** Add `Manager.UpdatePresentation` to the mutation landscape as a `Manager`-internal writer of identity-family keys. State whether it becomes a session-owned command in slice 3 (identity retirement) or stays as an internal method.

### [Major] `AcknowledgeDrainPatch` and `DrainAckStopPendingPatch` clear create/start family fields, crossing the slice boundary

`AcknowledgeDrainPatch` (`lifecycle_transition.go`) clears `pending_create_claim`, `pending_create_started_at`, and `last_woke_at` — all create/start family or wake/hold/drain family fields. `DrainAckStopPendingPatch` also clears `pending_create_claim` and `pending_create_started_at`. Both are in the drain cluster (slice 4), but they write fields from the create/start family (slice 2).

This is the same cross-family pattern as `closeFailedCreateBead`: a terminal or state-transition patch clears create/start metadata as a "fresh start" or "abandon pending create" side effect. The design doesn't document this cross-family clearing pattern.

**Required change:** Document the "terminal cleanup clears pending-create metadata" pattern explicitly. For each command that clears `pending_create_claim` or `pending_create_started_at`, state whether: (a) the clear is an integral part of the terminal transition (and thus owned by the terminal-command's slice), or (b) the clear should call a create-start command when available. The current code treats these clears as part of the terminal transition, which is the natural answer, but it must be stated so the static guard doesn't block them.

---

## Minor Risks

### [Minor] The `Suspend` method writes lifecycle and wake/hold/drain fields but is not in the landscape

`Manager.Suspend` (`internal/session/manager.go:754–758`) acquires the mutation lock and calls a transition, but the actual metadata write goes through `withSessionMutationLock` → `checkTransition(CmdSuspend)` → state transition. The suspend command writes `state` and `sleep_reason` (lifecycle family) and potentially `sleep_intent` (wake/hold/drain family). It's listed as a `Manager` method (canonical owner), but the landscape table's `internal/session/manager.go` row should mention `Suspend` as a cross-family write if it clears `sleep_intent`.

### [Minor] The reconciler writes `sleep_intent` in multiple paths without a single owner

`session_reconciler.go:2073` clears `sleep_intent` via `store.SetMetadata(target.session.ID, "sleep_intent", "")`. `session_reconcile.go:615` clears `last_woke_at`. Both are wake/hold/drain fields written by the reconciler, which the landscape assigns to slices 5/6. But the reconciler is the only production writer that clears `sleep_intent` on a wake — `session_wake.go` and `cmd_session_wake.go` also write `sleep_intent`, creating at least three production writers for the same key without a clear owner.

---

## Missing Evidence

- A cross-family write policy that specifies what happens when one command writes keys from another slice's family. Current code has at least five commands that do this (`CloseDetailed`, `closeFailedCreateBead`, `AcknowledgeDrainPatch`, `DrainAckStopPendingPatch`, `SleepPatch`), but the design treats each field family as having one owner slice.
- A concurrency model for the coexistence period that specifies how `withSessionMutationLock` protects command writes against legacy `SetMetadataBatch` writes that bypass the lock.
- A complete list of `store.Update`-based mutation paths in `Manager` methods, since the static guard mentions `Update` but the landscape only lists `SetMetadata`/`SetMetadataBatch`-based writers.
- Whether `RepairEmptyType` is a classifier side effect or a pre-classification adapter step, and what its concurrency semantics are during slice 1 extraction.
- The atomicity contract for the prepare-runtime-start phase, which currently uses individual `SetMetadata` calls rather than a single batch.

---

## Required Changes

1. **Add a cross-family write policy** to the mutation boundary section. For each command that writes keys from another slice's family (close writes wake/hold/drain and identity; drain ack writes create/start; suspend may write wake/hold/drain; `closeFailedCreateBead` writes create/start and wake/hold/drain), document: which cross-family keys it writes, whether it calls the owning slice's command or writes directly, and the handoff condition.

2. **Fix the landscape table owner-slice assignments for cross-family writers.** Either add a "field families written" column, or split entries like `cmd/gc/session_beads.go` into per-family entries. The current "Owner slice: 3" for `session_beads.go` is misleading because `closeFailedCreateBead` writes slices 2, 3, and 4 fields.

3. **Specify the concurrency model for the coexistence period.** Either export the mutation lock for legacy writer use, add a store-level revision mechanism, or document which field interleavings are safe (idempotent clears, monotonic timestamps) vs. which require strict ordering (state transitions). The current "lock-plus-reread" claim is unenforceable when legacy writers bypass the lock.

4. **Batch the prepare-runtime-start writes** (`session_key` and `instance_token`) or document in the atomicity contract that they are idempotent-precondition fields that can be safely partially written before the batch.

5. **Move `RepairEmptyType` to a pre-classification adapter step** or document it as a write side effect with specified concurrency semantics. The classifier contract in TR-002 says "fact construction does not mutate session metadata"; either the classifier must be pure, or TR-002 needs an exception.

6. **Add `Manager.UpdatePresentation` and `Manager.Suspend`** to the mutation landscape as `Manager`-internal writers of identity-family and cross-family keys, respectively.

7. **Document the "terminal cleanup clears pending-create metadata" pattern** explicitly. For each terminal or state-transition patch that clears `pending_create_claim`, `pending_create_started_at`, `sleep_intent`, or `last_woke_at`, state whether the clear is owned by the terminal-command's slice or delegated to the owning slice's command.

---

## Questions

- When slice 3 extracts the close command, should `clearWakeAndHoldOverrides` and `retireConfiguredNamedSessionIdentifiers` become separate commands called by the close command (creating a command-call graph across slices), or should the close command temporarily own these writes until slices 4 and 1/3 extract their own commands?
- Is `withSessionMutationLock` intended to be the long-term concurrency mechanism, or is it a placeholder until the store gains native revision tokens? If the former, it must be exportable or wrapped for legacy callers. If the latter, the design should state this explicitly and specify what protects commands during coexistence.
- Should `RepairEmptyType` be called by each surface before invoking the classifier (making the classifier pure), or should the classifier call it internally as an idempotent normalization (making the classifier impure but safe)?
- The design says `cmd/gc/session_lifecycle_parallel.go` "must delegate as each start slice converts." Does "delegate" mean per-operation (each operation in this file calls a command when available), per-field-family (each field family's writes route through the owning command), or per-file (the entire file converts at once)? The file contains start-path mutations (slice 2), close-path mutations (slice 3), and pool/desired-state logic (slices 5/6).
- During the coexistence period, if a legacy `SetMetadataBatch` in the reconciler writes `sleep_intent` between a command's `loadSessionBead` and `SetMetadataBatch`, the command's batch will overwrite the reconciler's write only if the command also writes `sleep_intent`. If the command doesn't write `sleep_intent`, the reconciler's write survives. Is this acceptable, or does the design need a "must-write-all-touched-fields" rule for commands?
