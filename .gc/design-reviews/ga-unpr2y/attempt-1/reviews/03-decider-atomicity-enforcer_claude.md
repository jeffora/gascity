# Takeshi Yamamoto - Claude

**Verdict:** block

Lane: pure deciders, optimistic concurrency, commit-event-intent ordering,
stale-fact defense. Reviews the current `DESIGN.md` (attempt-15 review-response
revision) with `REQUIREMENTS.md` and `internal/session/AGENTS.md`. All store-API
claims verified against the checkout; citations inline. The block is narrow and
addressable: the Atomic Command Contract mandates a conditional-commit primitive
the persistence layer does not provide and the design never records that answer.

**Top strengths:**
- The decider/command split and pure-decider framing are exactly right for this
  lane. The Atomic Command Contract requires "immutable facts read by the decider,
  including mandatory `now`, current config hash ..., bead revision/token when
  available, and runtime observation timestamp" (DESIGN.md:365-369) and Boundaries
  requires "pure decisions that can be unit-tested without stores or providers"
  (DESIGN.md:627). `internal/session/lifecycle_projection.go` already proves a
  pure projection is feasible. Lane Q1 is well answered: clock and config enter as
  injected facts, not ambient reads.
- The post-write-verify half of the contract is real, not aspirational. The design
  requires a "post-write verifier and expected durable state" (DESIGN.md:380), and
  the platform already implements one: `BdStore.waitForUpdateProjection`
  (`internal/beads/bdstore.go:1037`) polls `Get(id)` until `updateProjectionMatches`
  before returning. An implementer has a concrete pattern to extend.
- The failure-injection matrix is thorough and lane-correct: stale snapshots,
  raced lifecycle operations, partial metadata writes, duplicate commands,
  "provider success followed by commit failure, commit success followed by event
  failure, skipped events, and crash recovery scan convergence"
  (DESIGN.md:393-396), backed by a durable-scan recovery owner (DESIGN.md:513-514).
  This is the right defense set — it just presumes a fence the store cannot supply.

**Critical risks:**
- [Blocker] **The Atomic Command Contract's pre-commit conditional has no store
  primitive, so "commit conditional on the validated snapshot" is structurally
  unachievable today — red flag #2 realized in the platform.** The contract
  mandates "exact pre-commit revalidation: token/revision/value preconditions
  checked immediately before commit against the same fact snapshot the decider
  consumed" (DESIGN.md:370-371) and a fence of "store-native compare-and-swap,
  value-embedded token checked immediately before commit, or ... repair-converged
  blind write" (DESIGN.md:346-348). Against the checkout: `beads.UpdateOpts`
  (`internal/beads/beads.go:46`) carries **no precondition/version/expected field**
  — every field is an unconditional setter; the `Bead` type has **no revision/etag
  column**, only `UpdatedAt time.Time`; `SetMetadata`/`SetMetadataBatch` and
  `Update` are unconditional blind writes across **every** backend (filestore,
  exec, hqstore, bdstore, memstore, caching). `hqstore.SetMetadataBatch`
  (`internal/beads/hqstore_core.go:393`) is batch-atomic under an **in-process**
  mutex but has no compare check; `BdStore.Update`
  (`internal/beads/bdstore.go:737`) shells out to `bd update --set-metadata k=v`
  with no `--if-version`/`--if-unchanged`, and `bdStoreTx.apply`
  (`internal/beads/bdstore.go:994`) batches without any conflict detection
  (`ErrConflict`/`RowsAffected`/`WHERE updated_at=…` all absent). Net: of the three
  fence options, (a) store CAS and (b) value-embedded token both require primitives
  that do not exist, leaving only (c) repair-converged blind write — and a
  "revalidate-read then unconditional write" sequence is a TOCTOU race, made wider
  by the projection lag that `waitForUpdateProjection` exists to absorb. The
  design presents three fences as equals and never records that two are
  unavailable. This contract is the design's answer to the attempt-15 finding
  "Atomic command semantics are not enforceable" (DESIGN.md:69); as written it does
  not enforce them — it specifies a mechanism the store cannot run.
- [Major] **SessionMutation + RuntimeIntent ordering is required to be defined but
  the safe invariant and the idempotency that recovery depends on are not pinned
  (lane Q3).** The contract asks each command to state "runtime provider ordering,"
  "post-commit event ordering," "rollback or compensation rule," and a
  "partial-state matrix" (DESIGN.md:381-384), and slice 5 names
  prepare/commit/rollback for runtime start (DESIGN.md:674-676). But because there
  is no CAS, idempotent re-drive by the durable-scan owner is the *only* recovery
  path after a post-commit runtime failure — yet the design never states the
  load-bearing invariants: (i) durable intent must be committed before irreversible
  runtime side effects (or a named compensation exists), and (ii) every
  RuntimeIntent execution must be idempotent/retryable so the reconciler can
  re-drive it. Without (i)+(ii) explicit, "provider success followed by commit
  failure" (DESIGN.md:395) has no defined convergence — red flag #3.
- [Major] **`REQUIREMENTS.md` SESSION-START-001 promises atomicity the store
  cannot keep, and no slice reconciles the wording.** "Pending create metadata
  must be cleared atomically when start is confirmed or rolled back"
  (REQUIREMENTS.md:115) is satisfiable only as *batch-atomic* (all keys in one
  `SetMetadataBatch`), not as *conditional-atomic* ("clear only if still
  start-pending"). A stale-create reconciler write
  (`cmd/gc/session_reconcile.go`) racing a start-confirm write resolves
  last-writer-wins. Slice 5 lists "stale-token ... tests" (DESIGN.md:675), but
  "stale-token" presupposes a token the store does not have. The requirement's
  "atomically" must be defined as batch-atomic + convergence, or the row needs the
  new primitive.

**Missing evidence:**
- The store-capability answer for the current surface is never stated. The design
  requires a "store capability row for the current persistence surface: atomic
  update, partial metadata update, conditional update, close/reopen semantics, and
  blind write behavior" (DESIGN.md:372-374) but does not fill it in. The true row
  today is: atomic single-batch merge = yes (in-process); conditional/CAS update =
  **no**; revision token = **no**; cross-process fence = **none**; projection lag =
  **yes** (`waitForUpdateProjection`). Every per-command contract inherits this.
- No decision on whether a conditional/versioned write primitive will be added to
  the beads Store, or whether all session-owned commands adopt the repair-converged
  blind-write model. The backlog (DESIGN.md:656-679) contains no store-primitive
  slice and no such commitment.
- No statement of the canonical commit↔runtime ordering invariant, nor a required
  proof that RuntimeIntent execution is idempotent under re-drive.

**Required changes:**
1. Record the real store-capability answer in the Atomic Command Contract and in
   the Slice 0 `COMMAND_APPLIERS.yaml`/capability artifact: no store-native CAS, no
   revision token, batch-atomic-but-unconditional blind writes, observable
   projection lag. Then choose one and sequence it: **(a)** add a
   conditional/versioned write primitive to the beads Store (e.g.
   `UpdateIfUnchanged(id, opts, expectedUpdatedAt)` or a revision column with
   conflict error) as an explicit prerequisite slice before any command claims a
   "conditional fence"; or **(b)** drop the CAS/token language from per-command
   contracts and commit every session-owned command to the repair-converged
   blind-write model.
2. If (b), require **cross-process raced-writer** convergence tests per command
   (two independent store handles, not one in-process mutex), since the in-process
   batch atomicity of `hqstore`/`memstore` does not fence the multi-writer reality
   documented in the Mutation Ownership Ledger.
3. Pin the ordering invariants in the contract: durable intent committed before
   irreversible runtime side effects (or a named compensation), and a stated
   requirement that RuntimeIntent execution is idempotent/retryable so the durable
   scan owner can re-drive after a post-commit runtime failure.
4. Reconcile SESSION-START-001's "cleared atomically" to mean batch-atomic +
   convergence (or gate it on the new primitive), and replace slice 5's
   "stale-token" language accordingly.

**Questions:**
- Is the intended end-state optimistic concurrency (needs a store revision/CAS
  primitive — currently absent) or idempotent convergence (NDI, no fence, blind
  writes that converge)? The contract's language says the former; the platform
  only supports the latter. The design must pick one explicitly.
- For runtime start, is the canonical order durable-intent-then-spawn with
  reconciler re-drive on spawn failure, or spawn-then-commit with orphan-process
  cleanup? Slice 5 names prepare/commit/rollback but not the irreversibility order.
- Does `UpdatedAt` (wall clock, no monotonic guarantee, equal within resolution)
  count as a usable "value-embedded token," or is a real monotonic revision
  required before any command claims a token fence?
