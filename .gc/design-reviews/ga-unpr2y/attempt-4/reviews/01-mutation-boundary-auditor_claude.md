# Elena Marchetti - Claude

**Verdict:** block

Scope of this review is the mutation boundary only: session lifecycle/identity
write ownership, the patch-map escape-hatch surface, the canonical writer
inventory, and the static-guard enforceability. The section is materially
improved over a prose boundary, but it is blocked in this lane because the guard
as specified is defeated by the codebase's dominant write pattern and the
"canonical" writer inventory is provably incomplete. Both are exactly the
failure modes this persona exists to prevent.

**Top strengths:**
- The owned-key taxonomy (Mutation Boundary table, DESIGN.md:127-134) is concrete
  and field-family based, and it correctly mirrors the real keys in the source
  (`state`, `instance_token`, `pending_create_claim`, `session_key`, `alias`,
  `session_name`, `generation`, drain/hold keys). It also names the right
  primary enforcement rule — owned-key-set detection on session-bead store
  writes (DESIGN.md:206-209) — and the right discrimination requirement
  (non-session beads such as mail/extmsg/convoy/order are not violations,
  DESIGN.md:217-220).
- Exceptions are keyed by inventory ID with owner slice + retirement condition,
  and "shrink-only allowlist behavior" is stated (DESIGN.md:211-215). That is the
  correct shape for bounding doctor/migration/repair escape hatches.
- The guard is explicitly additive to and must-not-weaken the existing
  `cmd/gc/worker_boundary_import_test.go` (DESIGN.md:222-223), so it composes
  with the active worker-boundary migration rather than colliding with it.

**Critical risks:**
- **[Blocker] The patch-map guard rule is defeated by the existing call pattern;
  enumerating 5 of 21 patch builders is the "patch helpers remain reachable"
  red flag in concrete form.** `internal/session/lifecycle_transition.go` exports
  **21** builders that return `MetadataPatch` (verified: `RequestWakePatch`,
  `RequestExplicitWakePatch`, `PreWakePatch`, `ContinuationResetWakePatch`,
  `ClearWakeBlockersPatch`, `ClearExpiredHoldPatch`, `ClearExpiredQuarantinePatch`,
  `ConfirmStartedPatch`, `CommitStartedPatch`, `BeginDrainPatch`,
  `DrainAckStopPendingPatch`, `SleepPatch`, `AcknowledgeDrainPatch`,
  `CompleteDrainPatch`, `RestartRequestPatch`, `ConfigDriftResetPatch`,
  `ArchivePatch`, `ClosePatch`, `RetireNamedSessionPatch`, `QuarantinePatch`,
  `ReactivatePatch`). The Static Guard bullet (DESIGN.md:204-205) names only
  `RequestWakePatch`, `PreWakePatch`, `CommitStartedPatch`, `ClosePatch`,
  `RetireNamedSessionPatch` — leaving 16 exported builders (notably
  `BeginDrainPatch`, `SleepPatch`, `ArchivePatch`, `QuarantinePatch`,
  `ConfigDriftResetPatch`, `RestartRequestPatch`) as un-guarded escape hatches a
  future production caller can apply. Worse, the **type-based** version of the
  rule is also defeated by indirection: `MetadataPatch` is `map[string]string`
  (lifecycle_transition.go:19), and patches reach the store through the generic
  helper `setMetaBatch(store beads.Store, id string, batch map[string]string, ...)`
  (`cmd/gc/session_beads.go:1725`) — the named type is erased at that boundary,
  and any caller can hand-roll an inline `map[string]string` of owned keys and
  pass it without ever touching a patch builder. The owned-key-set rule
  (DESIGN.md:206-209) is the only one that survives, yet it is presented as
  co-equal to the defeatable patch-name rule rather than as the load-bearing
  rule.
- **[Blocker] The "Canonical Production Writer Inventory" is demonstrably
  incomplete, which invalidates downstream single-owner claims.** Two production
  (non-test) files write owned lifecycle/identity keys via inline maps and do not
  appear anywhere in W-001..W-021:
  - `cmd/gc/adoption_barrier.go:168-174` writes `session_name`, `state="active"`,
    `generation`, `continuation_epoch`, `instance_token` (Lifecycle state +
    Create/start lease + Runtime identity + Identity families). Note `state="active"`
    is set directly, bypassing the transition table referenced by SESSION-STATE-001.
  - `cmd/gc/session_name_lookup.go:189-200` writes `state=start-pending`,
    `pending_create_claim`, `pending_create_started_at`, `session_origin`,
    `generation`, `continuation_epoch`, `instance_token`, `session_name`, `alias`.

  Both write `instance_token`, yet Slice 3 stakes "`instance_token` ... must be
  owned by one runtime-start command slice" (DESIGN.md:130) and the inventory
  attributes instance_token writers only to W-005 (`session_lifecycle_parallel.go`)
  and W-012 (`cmd_prime.go`). A grep for `instance_token` writes returns at least
  `cmd/gc/session_beads.go:1035`, `cmd/gc/session_name_lookup.go:198`,
  `cmd/gc/adoption_barrier.go:173`, and `cmd/gc/session_lifecycle_parallel.go:888`
  in production. Until the inventory is regenerated from a mechanical source scan,
  the single-owner claim and the W-005/W-012 "retire together" gate cannot be
  trusted, because the set of writers to retire is unknown.
- **[Major] The anti-regression guard lands too late to protect the migration's
  longest phase.** "Before a slice is considered implemented, add or tighten a
  failing-build guard" (DESIGN.md:184-185) defers the guard per-slice. Slice 1
  (Target Classification) is a read/resolution slice with no mutation, so by the
  letter of the plan it can land with **no** mutation guard in place, leaving the
  entire owned-key surface open to new bypasses for the duration of slices 1-2+.
  The guard is independent of any single behavior extraction: it only needs the
  owned-key fixture plus an allowlist seeded from the *current* inventory. It
  should land first, as a guard-only step, so new writers fail the build from day
  one rather than after the field family they touch is extracted.
- **[Major] "Shrink-only" and exception bounding are honor-system, not enforced.**
  The allowlist is "shrink-only" and "adding a row needs a matching update to this
  design or `AGENTS.md`" (DESIGN.md:211-215), but nothing described fails CI when
  the allowlist grows. Exceptions are also category/file based (tests, fixtures,
  "explicit doctor/repair utilities", DESIGN.md:141-149, 192-193), which is
  self-labelable: a caller dodges the guard by living in a doctor-named file or a
  `repair*` function. The repair requirement ("emits trace/log evidence for each
  direct repair write", DESIGN.md:146-147) is the right bound but is not stated as
  a testable obligation. Without a frozen, checked-in allowlist baseline that the
  guard test asserts can only shrink, and without funneling repair writes through
  a single audited helper, the third red flag — "exception list grows into a
  permanent bypass" — is unmanaged.

**Missing evidence:**
- No mutation-boundary guard exists in the checkout yet (only
  `cmd/gc/worker_boundary_import_test.go`). The "owned key taxonomy ... mirrored
  in one central test fixture" (DESIGN.md:189) is referenced but not present, so
  the guard's feasibility against `setMetaBatch`-style indirection and inline
  owned-key maps is asserted, not demonstrated. A spike test proving the guard can
  flag `adoption_barrier.go:173` and `session_name_lookup.go:198` would convert
  this from claim to evidence.
- The inventory's "Owner slice" / "Exception status" columns are asserted from
  prose, not from a source-derived call-site list. There is no count or
  completion criterion tying the 21 patch builders and the ~100 external
  `SetMetadata*` call sites down to the subset that actually writes session-bead
  owned keys, so "all writers converted" has no measurable definition of done.
- W-021 ("Generic `beads.Store.Update`, `SetMetadata*`, `Close`, `Create`
  bridges") is the row that actually covers `setMetaBatch` and inline-map writers,
  but it is marked "unknown" field family and "per slice" owner — i.e., the
  highest-risk, most-reachable path is the least specified.

**Required changes:**
1. Rewrite the Static Guard so the **owned-key-set write** is the single
   load-bearing rule and the patch-name list is explicitly demoted to a
   non-exhaustive convenience signal. The guard must flag any
   `SetMetadata`/`SetMetadataBatch`/`Update`/`Create`/`Close` (and local wrappers
   such as `setMeta`/`setMetaBatch`) whose written key set intersects the owned
   taxonomy when the target bead may be a session bead — regardless of whether the
   value came from a patch builder or an inline `map[string]string`. Either guard
   on the `MetadataPatch` return type for all 21 builders or, preferably, drop the
   name enumeration entirely in favor of the key-set rule.
2. Regenerate the Canonical Production Writer Inventory from a mechanical source
   scan and add the missing production writers, at minimum
   `cmd/gc/adoption_barrier.go` and `cmd/gc/session_name_lookup.go`, each with a
   field-family assignment and owner slice. Re-validate Slice 3's
   "single-owner `instance_token`" claim against the regenerated list before that
   slice is considered ready.
3. Move the boundary guard to a guard-only first step that lands before Slice 1's
   mutation work, seeded with the owned-key fixture and an allowlist mirroring the
   current (regenerated) inventory, so new bypasses fail the build during the
   entire migration rather than per-field-family.
4. Make shrink-only and exception bounding enforceable: a checked-in allowlist
   baseline the guard test asserts cannot grow without a paired design/`AGENTS.md`
   diff; replace file-path/naming-based exception detection with explicit
   allowlist rows keyed by call-site identity; and route doctor/migration/repair
   writes through one audited helper that emits trace evidence, asserted by test.

**Questions:**
- Does the active bead store apply a multi-key `SetMetadataBatch` atomically? The
  guard and the inventory both assume owned-key writes are the unit of
  enforcement, but `adoption_barrier.go`/`session_name_lookup.go` write 5-9 owned
  keys at once via `Create`; if these are not atomic, every such site also needs a
  Command Atomicity repair row, not just an inventory row.
- For W-021 generic bridges, what is the concrete session-bead discrimination
  signal the guard will use at a call site that only has a `beads.Store` and a
  bead ID (no type in hand)? Is it a bead-type lookup, a key-set heuristic, or a
  required typed wrapper? This determines whether the guard is decidable at AST
  time or needs a runtime/type-flow analysis.
- Should the 16 currently-unused-externally patch builders (everything except
  `Close`/`RetireNamedSession`, which already have external callers at
  `session_beads.go:444/514/1502/1737/2161` and
  `internal/api/session_resolution.go:171`) be made unexported until a command
  needs them, removing the escape hatch by construction rather than by guard?
