# Natasha Volkov — Gemini (Independent Review, Attempt 4)

**Verdict:** block

**Review focus:** Cross-file consistency, missed edge cases, pattern drift, and architectural coherence — with evidence drawn from code verification.

The Attempt 4 iteration of the session boundary design makes notable progress in structuring the extraction backlog and aligning slice gates. The scenario traceability matrix, command atomicity contract, runtime recovery ordering, and static-guard specifications are now clear frameworks. 

However, critical gaps remain. We are still encountering unmapped requirement rows, missing production writers in our inventory, and unresolved behavior ambiguities that would allow significant drift during implementation. This review details these findings and blocks the design until the required changes are addressed.

---

## Top Strengths

- **Traceability Foundation:** The addition of the scenario traceability matrix (`DESIGN.md:300–308`) provides a concrete per-slice row mapping, which ensures that developers do not arbitrarily select which requirements to implement and test.
- **Granular Field Taxonomy:** The field-family taxonomy (`DESIGN.md:127–134`) clearly isolates session-owned keys (`state`, `instance_token`, `generation`, etc.) from other domains, establishing a reliable baseline for static verification.
- **Precedence & Fallback Rules:** The target classification precedence table and operation policy matrix (`DESIGN.md:270–288`) are highly detailed and provide robust protection against ambiguous session targeting and factory-fallback drift.
- **Coexistence Safeguards:** Coexistence gates (`DESIGN.md:491–497`) specify clear, non-overlapping constraints on legacy vs. converted writers per slice, ensuring the build remains green throughout the long extraction phase.

---

## Critical Risks

### [Blocker] The scenario traceability matrix leaves 13 requirements rows completely unmapped

A comparative audit between `REQUIREMENTS.md` and the scenario traceability matrix (`DESIGN.md:300–308`) reveals that exactly 13 scenario rows are completely unmapped to any implementation slice. An implementer executing the slices as written has no obligation to prove parity or run characterization tests for these 13 areas:

1. `SESSION-LIFE-001` (Legacy compatibility states are projected)
2. `SESSION-LIFE-002` (Pending create claim)
3. `SESSION-LIFE-006` (Missing config)
4. `SESSION-LIFE-008` (User-facing projection guard)
5. `SESSION-STATE-001` (Legal transition table)
6. `SESSION-STATE-002` (Illegal transitions reject)
7. `SESSION-STATE-003` (UI affordances follow reducer)
8. `SESSION-ID-001` (Explicit session names)
9. `SESSION-ID-002` (Aliases)
10. `SESSION-START-004` (Suspend named session)
11. `SESSION-RECON-001` (Worker boundary)
12. `SESSION-WORK-003` (Orphan pool step beads)
13. `SESSION-RUNTIME-004` (Stop turn)

Of these, `SESSION-RECON-001` is the critical architectural boundary enforcing that CLI commands route through `internal/worker/handle.go`. Omitting it means no slice is accountable for verifying that the direct manager-construction bypasses are retired. Similarly, `SESSION-START-004` (governing materialization when suspending configured named sessions) remains entirely unallocated, despite being a major driver of the target classifier's materialization contract.

**Required change:** Assign all 13 unmapped `SESSION-*` rows to their appropriate slices in the scenario traceability matrix. Specifically:
- Map `SESSION-ID-001` and `SESSION-ID-002` to Slice 1 (Target classification).
- Map `SESSION-LIFE-001`, `SESSION-LIFE-002`, `SESSION-LIFE-006`, and `SESSION-LIFE-008` to Slice 5 (Wake/hold/drain eligibility) or Slice 1 where they affect target classification.
- Map `SESSION-START-004` to Slice 3 (Runtime start prepare/commit/rollback) or Slice 4 (Close/retirement).
- Map `SESSION-STATE-001`, `SESSION-STATE-002`, and `SESSION-STATE-003` to Slice 2 or explicitly as cross-cutting invariants across all mutating command slices (Slices 2–7).
- Map `SESSION-RECON-001` to Slice 6 (Reconciler fact extraction).
- Map `SESSION-WORK-003` to Slice 4 (Close and identity retirement) or Slice 5.
- Map `SESSION-RUNTIME-004` to Slice 7 (Provider health/progress) or Slice 5 (Wake/hold/drain).

### [Blocker] State transition invariants are unassigned to any mutating command slice

`SESSION-STATE-001` (Legal transition table) and `SESSION-STATE-002` (Illegal transitions reject) define the core state-machine reducer rules. Despite Slices 2–7 performing state mutations (Explicit wake, Runtime start, Close/retire, etc.), neither row is mapped to any mutating slice in the matrix. Without this mapping, a slice can rewrite metadata transitions or bypass state validation entirely because the design does not enforce transition-parity checks per slice.

**Required change:** Assign `SESSION-STATE-001` and `SESSION-STATE-002` as explicit parity requirements across all mutating command slices (Slices 2–7), or state a global, non-negotiable rule in `DESIGN.md` that every mutating command slice must assert state-machine transition validation and reject invalid states.

### [Blocker] The canonical production writer inventory is incomplete: key production writers are missing

The production writer inventory (`DESIGN.md:158–180`) still omits major active production code files that write session-owned keys. Without these files mapped to explicit slices and retirement plans, the static guard cannot enforce boundary airtightness and behavior will silently drift:

- `cmd/gc/adoption_barrier.go:168–174` (writes `session_name`, `state="active"`, `generation`, `continuation_epoch`, `instance_token`).
- `cmd/gc/session_name_lookup.go:189–200` (writes `state=start-pending`, `pending_create_claim`, `pending_create_started_at`, `session_origin`, `generation`, `instance_token`, `session_name`, `alias`).
- `cmd/gc/cmd_session.go:1592` (writes `held_until`, `sleep_intent`, `state`).
- `cmd/gc/cmd_stop.go:329` (writes `sleep_reason`).
- `cmd/gc/cmd_wait.go` (writes `wait_hold`, `sleep_intent`, `closed_at`, `close_reason`).
- `cmd/gc/cmd_nudge.go:988` (writes `last_nudge_delivered_at`).

Furthermore, `adoption_barrier.go` writes `state="active"` directly to the store, bypassing the transition table (`SESSION-STATE-001`). Both `adoption_barrier.go` and `session_name_lookup.go` write `instance_token`, which directly contradicts Slice 3's claim that `instance_token` is owned solely by W-005 and W-012.

**Required change:** Add explicit rows or concrete inventory IDs for `cmd/gc/adoption_barrier.go`, `cmd/gc/session_name_lookup.go`, `cmd/gc/cmd_session.go` (suspend/resume path), `cmd/gc/cmd_stop.go`, `cmd/gc/cmd_wait.go`, and `cmd/gc/cmd_nudge.go`. Specify their owner slices, field-family classifications, and clear retirement or exception-status paths.

### [Blocker] REQUIREMENTS.md cites deleted test files as proof, violating its own maintenance rules

The requirements ledger cites several test files as active proofs of behavior, but these files have been deleted from HEAD and do not exist in the current checkout:

- `cmd/gc/scale_from_zero_test.go` (cited by `SESSION-RECON-002` and `SESSION-RECON-003`)
- `cmd/gc/provider_health_gate_test.go` (cited by `SESSION-RECON-006`)
- `cmd/gc/session_progress_test.go` (cited by `SESSION-RECON-007`)

This violates `REQUIREMENTS.md:159`: *"Keep evidence current. If a cited test is deleted, move or replace the row."* This is highly dangerous as future implementers reading the requirements ledger will believe these behaviors are actively tested when they are completely uncovered.

**Required change:** Annotate these four stale rows in `REQUIREMENTS.md` with explicit warnings that their cited test files are absent and currently uncovered. Add cross-references to the corresponding `DESIGN.md` blocking notes (Slices 6 and 7) so readers can find the blocks.

### [Blocker] TR-001 still permits too-lax parity assertions

`TR-001` (`DESIGN.md:446–454`) states that new technical APIs must prove parity with *"at least one existing scenario row."* This rule allows an implementer to verify parity against a single row while their slice touches multiple rows, leaving the other behaviors unverified and prone to regression.

**Required change:** Update `TR-001` to require parity with **every** scenario row the slice touches, cementing the scenario traceability matrix as the authoritative per-slice row list.

---

## Major Risks

### [Major] Wake-vs-hold parity remains unpinned

Current production behavior (as seen in `waits_test.go`) dictates that an explicit wake releases user holds and quarantine, while a rejected or failed wake preserves `held_until`. However, the command atomicity contract for "Request explicit wake" does not specify hold-conflict semantics. This ambiguity allows the new command implementation to either reject the wake on hold or silently discard the hold.

**Required change:** Add an explicit requirement row or extend `SESSION-START-003` to state: *"Explicit wake releases user hold and crash-loop quarantine; rejected terminal wakes preserve held_until."* Add both scenarios to Slice 2's freshness gate.

### [Major] CommitStartedPatch vocabulary ambiguity between single builder and two-phase command

The vocabulary table lists `CommitStartedPatch` as a single entry. However, the command atomicity contract defines a two-phase start: "Prepare runtime start" and "Commit runtime start". If a single patch builder is used for both phases, implementers are highly likely to conflate prepare-phase and commit-phase precondition checks, violating the atomicity contract.

**Required change:** Clarify whether `CommitStartedPatch` remains a single patch builder (with separate validation logic) or will be split into two discrete helpers (e.g. `PrepareStartedPatch` and `CommitStartedPatch`).

### [Major] Stored keys `last_woke_at` and `creation_complete_at` are missing from the mutation landscape

`creation_complete_at` is listed in the field-family table but its writers are not mapped. `last_woke_at` (written by the reconciler in `cmd/gc/session_reconciler.go` and read by `projectRuntimeProjection`) is omitted entirely. These keys must be assigned to owner slices to prevent unmonitored direct mutations.

**Required change:** Add `last_woke_at` to the field-family taxonomy under wake/hold/drain or create/start, and explicitly list its writers and owner slices in the mutation landscape. Map `creation_complete_at`'s writers in the landscape.

### [Major] Stated keys 'closed_at' and 'close_reason' on non-session beads lack discriminator strategy

`close_reason` and `closed_at` are written across multiple bead domains, including session, convoy, molecule, order-tracking, and nudge beads. The static guard must be able to distinguish session-bead updates from non-session-bead updates to avoid massive false-positives, yet no discriminator strategy is defined.

**Required change:** Add a bead-type discriminator strategy to the static guard specification. State whether the guard enforces key names only on session beads (e.g., using a type filter or typed wrappers) or whether a broader AST heuristic is used.

---

## Minor Risks

- **Slice 1 claims `SESSION-ID-010` but omits its evidence:** The matrix lists `SESSION-ID-010` for target classification but does not include `internal/session/manager_test.go` or `cmd/gc/session_template_start_test.go` in its current-proof column, which are the primary evidence. Either add the proof files or remove the row from Slice 1.
- **Slice 5 cites `build_desired_state_test.go` but omits `SESSION-WORK-003`:** `build_desired_state_test.go` is the primary evidence for `SESSION-WORK-003` (orphan pool step beads), but `SESSION-WORK-003` is not assigned to Slice 5's matrix. Ensure these are aligned.
- **Inconsistent behavior-change approval bars:** Slice 2 allows "documented bug fix", Slice 3 requires "owner approval", and the Product Rule allows arbitrary code inspection to identify "a real bug". The absence of a single clear authority for classifying a behavior deviation as a "bug fix" vs. a "product change" risks silent rewriting of requirements to fit a refactor.

---

## Missing Evidence

- **Mechanical Source Scan of Production Writers:** There is no evidence of a systematic, source-derived inventory of all production `SetMetadata` and `Create` call sites. The current inventory is hand-written and provably incomplete.
- **Guard Feasibility spike:** No proof-of-concept exists showing that the static guard can successfully flag inline-map writes or handle dynamic key indirection in `setMetaBatch` wrapper calls.
- **Bead-type discrimination proof:** No AST or runtime strategy has been demonstrated to show how the guard distinguishes session beads from other bead domains.

---

## Required Changes

1. **Map all 13 unmapped requirements rows** to their respective slices in the scenario traceability matrix (`DESIGN.md:300–308`).
2. **Assign `SESSION-STATE-001` and `SESSION-STATE-002`** as explicit parity obligations for all mutating command slices (Slices 2–7).
3. **Complete the mutation landscape inventory** by adding explicit rows for `adoption_barrier.go`, `session_name_lookup.go`, `cmd_session.go` (suspend), `cmd_stop.go`, `cmd_wait.go`, and `cmd_nudge.go`.
4. **Annotate the deleted/absent proof files** in `REQUIREMENTS.md` (`scale_from_zero_test.go`, `provider_health_gate_test.go`, and `session_progress_test.go`) as currently uncovered.
5. **Tighten TR-001** to require parity with **every** scenario row the slice touches.
6. **Pin wake-vs-hold behavior** by adding a requirement row or extending `SESSION-START-003` to define explicit wake hold-release semantics.
7. **Clarify the `CommitStartedPatch` ambiguity** and state whether it will be split.
8. **Add `last_woke_at` and `creation_complete_at`** to the mutation landscape.
9. **Define a bead-type discriminator strategy** in the static guard specification to handle non-session bead collisions.

---

## Questions

- Does the design intend `cmd/gc/session_name_lookup.go` to be part of Slice 1 (target classification), Slice 3 (close/identity retirement), or both, given that it writes both identity and create/start keys?
- Should `last_nudge_delivered_at` be inside the mutation boundary (nudge delivery is a lifecycle event) or outside (operational timestamp written by nudge dispatcher)?
- Is `SESSION-WORK-003` (orphan pool step bead recovery) in scope for any current slice, or is it explicitly deferred?
- When Slices 5 and 6 restore proofs, do the `REQUIREMENTS.md` rows keep their commit citations as historical anchors alongside the new test paths, or are commits retired as evidence?
