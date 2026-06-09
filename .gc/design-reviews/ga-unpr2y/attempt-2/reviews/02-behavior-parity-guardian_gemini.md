# Natasha Volkov — DeepSeek V4 Flash (Independent Review, Attempt 2)

**Verdict:** block

**Review focus:** Cross-file consistency, missed edge cases, pattern drift, and architectural coherence — with evidence drawn from code verification.

The iteration addresses most of attempt 1's structural blockers: the mutation landscape inventory, scenario traceability matrix, command atomicity contract, runtime ordering table, reconciler fact contract, target classification precedence table, and operability contract are all real, load-bearing additions. The stale-proof blocking rule (`DESIGN.md:21–23`) is correctly enforced. These additions move the design from an outline toward an implementable contract.

However, the iteration introduces new inconsistencies between the design and the current codebase, leaves several attempt-1 findings only partially resolved, and contains cross-reference gaps that would let an implementer skip parity proof for scenarios the design's own vocabulary says the slice touches.

---

## Top Strengths

- The field-level mutation boundary table (`DESIGN.md:67–77`) replaces vague "lifecycle and identity metadata" prose with named key families. This is the right granularity for a guard and for per-slice retirement tracking.
- The command atomicity contract (`DESIGN.md:260–276`) with per-operation precondition/commit/conflict/retry columns is a genuine structural improvement. It gives implementers concrete terms for stale-fact defense, partial-write handling, and conflict semantics.
- The scenario traceability matrix (`DESIGN.md:209–248`) ties every slice to `SESSION-*` rows, current proof, and freshness gates. This directly addresses attempt 1's central finding that backlog slices lacked row-level mapping.
- The runtime intent and recovery ordering table (`DESIGN.md:278–295`) assigns a recovery authority to each operation, making crash-after-commit convergence a design-level contract rather than an implementation detail.
- The target classification precedence table and operation policy matrix (`DESIGN.md:130–178`) are detailed enough to write adversarial parity tests against. The `candidates[]` and `negative_kind` result fields are a real improvement over returning bare error strings.

---

## Critical Risks

### [Blocker] The scenario traceability matrix omits scenario rows that the command atomicity contract touches

The matrix for slice 2 (explicit wake) lists `SESSION-LIFE-007`, `SESSION-START-003`, `SESSION-START-007` but omits `SESSION-START-001` (create state sequence), `SESSION-START-002` (stale create rollback), `SESSION-LIFE-002` (pending create claim), and `SESSION-LIFE-004` (creating staleness). This was the highest-severity finding in all three attempt-1 reviews and it is still not fixed.

The atomicity contract defines "Prepare runtime start" and "Commit runtime start" command clusters that write `pending_create_*`, `generation`, `instance_token`, and `creation_complete_at` fields — precisely the fields governed by those four rows. The mutation landscape assigns `cmd/gc/session_lifecycle_parallel.go` to slices 2, 5, and 6. An implementer who checks only the matrix rows for slice 2 will not test create/start parity, stale-create rollback, or pending-create claim clearing, because the matrix does not require it.

Similarly, `SESSION-STATE-001` and `SESSION-STATE-002` (legal/illegal transitions) are mentioned in the vocabulary table ("Commands must delegate to or prove parity with this table" at `DESIGN.md:73`) but do not appear in any slice's matrix rows. Every mutating command slice should carry an explicit transition-parity obligation.

**Required change:** Add `SESSION-START-001`, `SESSION-START-002`, `SESSION-LIFE-002`, and `SESSION-LIFE-004` to the slice(s) that convert `cmd/gc/session_lifecycle_parallel.go` runtime-start paths. Add `SESSION-STATE-001` and `SESSION-STATE-002` as explicit parity obligations for every mutating command slice, or state that transition-parity proof is delegated to the existing `state_machine_test.go` and is not repeated per slice.

### [Blocker] The mutation landscape inventory is incomplete: production writers of session-owned keys are missing

The inventory (`DESIGN.md:90–117`) lists 13 areas. Code verification against the current checkout finds production `SetMetadata`/`SetMetadataBatch` call sites writing session-owned keys that are not enumerated:

| Omitted site | Session-owned keys written | Impact |
|---|---|---|
| `cmd/gc/cmd_session.go:1592` | `held_until`, `sleep_intent`, `state` (suspend) | Lifecycle + wake/hold/drain families. Not in any row. |
| `cmd/gc/cmd_stop.go:329` | `sleep_reason` | Lifecycle family. Not in any row. |
| `cmd/gc/cmd_wait.go` | `wait_hold`, `sleep_intent`, `closed_at`, `close_reason` | Wake/hold/drain + lifecycle families. Partially covered by "sleep, pin, wait-hold" row but the row does not list closed_at or close_reason. |
| `cmd/gc/session_circuit_breaker.go` | circuit breaker state keys (reset generation, circuit state) | Not classified as inside or outside the boundary. `continuation_reset_pending` is in the create/start family table; circuit breaker keys are not. |
| `cmd/gc/session_name_lookup.go:192–230` | `state`, `generation`, `instance_token`, `session_name`, `alias` (create + rename) | Create/start + identity families. The primary session-materialization path has no inventory row. |
| `cmd/gc/cmd_nudge.go:988` | `last_nudge_delivered_at` | Not classified. Defined in `internal/session/manager.go` but written from `cmd/gc`. |
| `cmd/gc/convergence_store.go:121` | `close_reason` on convergence beads | Cross-domain collision: same key name on non-session bead type. |
| `cmd/gc/cmd_convoy.go`, `cmd/gc/molecule_autoclose.go`, `cmd/gc/nudge_beads.go`, `cmd/gc/order_dispatch.go` | `close_reason` on convoy/molecule/order beads | Same cross-domain collision. |

The design's own stale-proof rule (`DESIGN.md:21–23`) says implementation is blocked for any slice whose proof is missing. The inventory drives the static guard baseline. Every omitted writer is a site where the guard will not track retirement, and a site where behavior can silently drift during extraction because the design does not say who owns those writes.

**Required change:** Add explicit rows for every production writer of session-owned keys, classifying each as in-scope for a specific slice, exception, or outside the boundary with justification. Specifically add: `cmd/gc/cmd_session.go` (suspend), `cmd/gc/cmd_stop.go`, `cmd/gc/cmd_wait.go` (closed_at and close_reason columns), `cmd/gc/session_circuit_breaker.go`, `cmd/gc/session_name_lookup.go` (materialization), `cmd/gc/cmd_nudge.go` (last_nudge_delivered_at), and all `close_reason` writers on non-session beads. Add a bead-type discriminator strategy for `close_reason`, `closed_at`, and `state` on non-session beads.

### [Blocker] TR-001 still allows parity with "at least one scenario row"

`TR-001` (`DESIGN.md:446–454`) says: "New technical APIs are covered by tests that prove parity with at least one existing scenario row." This was flagged in attempt 1 as enabling a failure mode where a slice author proves parity against one row while the slice touches five. The traceability matrix partially compensates, but the TR is the hard gate that CI and code review will cite. A reader who checks only TR-001 will conclude that one row is sufficient.

**Required change:** Replace "at least one existing scenario row" with "every scenario row the slice touches" in TR-001, or add a subordinate rule that the traceability matrix is the authoritative per-slice row list and that TR-001's floor applies only to unlisted scenarios discovered during implementation.

### [Blocker] REQUIREMENTS.md stale citations still violate the ledger's own maintenance rule

Rows `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007` cite test files that do not exist in this checkout (`cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, `cmd/gc/session_progress_test.go`). The design correctly blocks slices 5 and 6 on restoring these proofs, but the requirements ledger still cites absent files as evidence. `REQUIREMENTS.md:159` says: "Keep evidence current. If a cited test is deleted, move or replace the row." The current citations violate this rule. A future implementer reading only `REQUIREMENTS.md` will not know these proofs are absent unless they also check `DESIGN.md` slice blocking notes.

**Required change:** Annotate the four stale rows in `REQUIREMENTS.md` with a note that the cited proof file is absent and the behavior is currently uncovered or covered only by commit citations. Add `DESIGN.md` cross-references so that readers of the requirements ledger can find the blocking note without searching the design backlog.

---

## Major Risks

### [Major] Wake-vs-hold parity is still unpinned

Current behavior: explicit wake releases user hold and crash-loop quarantine (`cmd/gc/cmd_session_wake.go` long help documents this; `internal/session/waits_test.go` proves `held_until`/`wait_hold` are cleared on wake, and that a rejected archived wake preserves `held_until`). The command atomicity contract for "Request explicit wake" lists `held/quarantine` as precondition fields but the conflict vocabulary (terminal, missing-config, identity-conflict, stale-state) has no hold conflict. This means a new implementation could either clear holds on wake (current behavior) or reject on hold (conflict vocabulary implies rejection) and both would satisfy the stated gate.

**Required change:** Extend `SESSION-START-003` or add a row stating: "Explicit wake releases user hold and crash-loop quarantine; rejected terminal wakes preserve `held_until`." Add both cases to slice 2's freshness gate. Decide explicitly whether hold release stays command-internal or becomes a typed conflict.

### [Major] `CommitStartedPatch` vocabulary ambiguity between single builder and two-phase command

The vocabulary table (`DESIGN.md:73`) lists `CommitStartedPatch` as one entry. The atomicity contract (`DESIGN.md:263–267`) defines two distinct command clusters: "Prepare runtime start" and "Commit runtime start". If the patch builder serves both phases, implementers will conflate precondition checks that belong to the commit phase with the prepare phase. If it is split, the vocabulary table should list both.

**Required change:** State whether `CommitStartedPatch` remains a single builder used by both prepare and commit command clusters, or will be split into `PrepareStartedPatch` and `CommitStartedPatch`. If single, add a note that prepare validation is separate from the patch builder.

### [Major] `last_woke_at` and `creation_complete_at` are in the boundary table but not in the mutation landscape

The field family table (`DESIGN.md:70`) lists `creation_complete_at` under create/start. The mutation landscape does not name which writers set it. Similarly, `last_woke_at` (not listed in the table but written by the reconciler in `cmd/gc/session_reconciler.go`) is read by `projectRuntimeProjection` to distinguish `fresh-creating` from `start-requested`. The boundary cannot be complete without these keys assigned to owner slices.

**Required change:** Add `last_woke_at` to the create/start key family (or the wake/hold/drain family, depending on ownership intent) and name its writer in the mutation landscape. Name `creation_complete_at`'s writer in the mutation landscape.

### [Major] `closed_at` and `close_reason` on non-session beads: no discriminator strategy

`close_reason` is written on session beads, convoy beads, molecule beads, order tracking beads, convergence beads, and nudge beads. The boundary guard must distinguish session-bead `close_reason` from non-session-bead `close_reason`, or accept that key-name enforcement is limited to patch builders and typed commands. The design lists `close_reason` in the lifecycle family (`DESIGN.md:69`) but does not address the cross-domain collision.

**Required change:** Add a bead-type discriminator note to the static guard specification: either the guard requires typed patch/command methods for session-bead lifecycle keys (making `SetMetadata("close_reason", ...)` on a session bead a violation regardless of which package calls it), or the guard uses a bead-type filter and the design must name which bead types are in scope.

### [Major] `cmd/gc/session_name_lookup.go` materialization path has no mutation landscape row

`session_name_lookup.go` writes `state`, `generation`, `instance_token`, `session_name`, and `alias` on session beads — keys from the create/start and identity families. It is the primary session-materialization path (creating a session from a name lookup). It is not in the mutation landscape. This is the highest-risk omission because the target classification slice (slice 1) is supposed to replace exactly this materialization logic, and the design does not track it.

**Required change:** Add a row for `cmd/gc/session_name_lookup.go` with field families (create/start, identity), target path (classifier + worker/session command delegation), owner slice (1 or 3 depending on whether classification subsumes materialization), and exception status.

### [Major] Reconciler fact contract leaves budget/circuit-breaker ownership ambiguous

The fact contract (`DESIGN.md:297–313`) says "Unknown budget state prevents additional destructive restart" but does not classify budget keys (restart counts, churn counts) as inside or outside the session boundary. If inside, they belong in the mutation landscape. If outside, the fact contract must state that the reconciler owns them and they are not subject to session-command extraction.

**Required change:** Classify `restart_count`, `churn_count`, `wake_attempts`, and circuit-breaker state keys as either inside the mutation boundary (add to landscape with owner slice) or outside (state explicitly that the reconciler owns them and they are not subject to command extraction).

### [Major] The design does not name expected new events per slice

The event contract (`DESIGN.md:285–295`) says new events require `events.RegisterPayload`, OpenAPI/SSE mapping, and test parity. The vocabulary table lists existing session events. But no slice names the events it will introduce. This means the design has an event registration gate but no way to verify that a slice has introduced all the events its commands need.

**Required change:** For each command cluster in the atomicity contract, name the session event(s) the command will emit or state that existing events suffice. At minimum, slices 2 and 3 should name their new events since they introduce commands that produce post-commit facts (wake accepted, wake blocked, close completed, identity retired).

---

## Minor Risks

- **Slice 1 claims `SESSION-ID-010` but omits its evidence.** The matrix lists `SESSION-ID-010` for target classification but does not include `internal/session/manager_test.go` or `cmd/gc/session_template_start_test.go` in its current-proof column, which are the evidence for that row in `REQUIREMENTS.md`. Either add the proof or remove the row.
- **Slice 5's current-proof column includes `cmd/gc/build_desired_state_test.go` which is the evidence for `SESSION-WORK-003` (orphan pool step beads), but `SESSION-WORK-003` is not in any slice's matrix.** Either assign it to a slice or state it is out of scope for current extraction.
- **The behavior-change approval bar is inconsistent.** Slice 2 allows "documented bug fix", slice 3 requires "owner approval", and the Product Rule (`DESIGN.md:17`) lets code inspection identify "a real bug". The person who classifies a deviation as bug-fix vs. product change is undefined. This is the seam through which a requirements ledger gets rewritten to fit a refactor.
- **Citation granularity is file-level only.** Freshness gates name case families ("must include terminal closed, archived, suspended...") but no test identifiers, so whether a gate is satisfied remains a judgment call rather than something a reviewer or CI guard can verify mechanically.
- **The shared call-site plan (`DESIGN.md:327–335`) lists 5 production sites but the mutation landscape lists 13 areas.** These two lists should be consistent. The call-site plan is the sequencing contract; the landscape is the guard baseline. Sites in the landscape but not in the call-site plan have no documented end state.

---

## Missing Evidence

- A complete generated inventory of all production `SetMetadata`/`SetMetadataBatch`/`Update`/`Close`/`Create` call sites that write session-owned keys, including the 19 `cmd/gc/` files and 3 `internal/api/` files verified in this review. The landscape table should cover each one with field family, owner slice, and exception status.
- A bead-type discriminator strategy for `close_reason`, `closed_at`, and `state` on non-session beads.
- Whether `cmd/gc/session_name_lookup.go` materialization is in scope for slice 1 (classification) or slice 3 (close/retire).
- Resolution of whether `CommitStartedPatch` is split or single, given the two-phase atomicity contract.
- Named new events for each command cluster, or a statement that existing events suffice.
- An evidence-existence guard: a test or CI check that parses the evidence columns of `REQUIREMENTS.md` and the design matrix and fails on missing file paths. The design's own stale-proof rule is enforced only by manual review, and this exact failure mode (three stale citations) was caught only by review, not by any automated check.
- Whether `SESSION-WORK-003` is in scope for any current slice or explicitly out of scope.

---

## Required Changes

1. **Add the missing scenario rows to the traceability matrix.** Add `SESSION-START-001`, `SESSION-START-002`, `SESSION-LIFE-002`, and `SESSION-LIFE-004` to the slice(s) that convert `cmd/gc/session_lifecycle_parallel.go` start paths. Add `SESSION-STATE-001` and `SESSION-STATE-002` as explicit parity obligations for every mutating command slice, or delegate them to `state_machine_test.go` with a stated assumption.

2. **Complete the mutation landscape inventory.** Add rows for `cmd/gc/cmd_session.go` (suspend), `cmd/gc/cmd_stop.go`, `cmd/gc/cmd_wait.go` (lifecycle keys), `cmd/gc/session_circuit_breaker.go`, `cmd/gc/session_name_lookup.go` (materialization), `cmd/gc/cmd_nudge.go` (`last_nudge_delivered_at`), `cmd/gc/cmd_prime.go` (`session_key`), `cmd/gc/soft_reload.go`, `cmd/gc/cmd_handoff.go`, `cmd/gc/pool_session_name.go`, and all non-session-bead `close_reason` writers. Classify each as in-scope, exception, or outside the boundary with justification.

3. **Tighten TR-001 from "at least one existing scenario row" to "every scenario row the slice touches."** The traceability matrix is authoritative for which rows a slice touches; TR-001 must not undermine it.

4. **Fix REQUIREMENTS.md stale citations.** Annotate `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007` with notes that the cited proof files are absent and the behavior is currently uncovered or covered only by commit citations. Cross-reference the design blocking notes.

5. **Pin wake-vs-hold behavior.** Add a `REQUIREMENTS.md` row or extend `SESSION-START-003` to state that explicit wake releases user hold/quarantine and that rejected terminal wakes preserve `held_until`. Add both cases to slice 2's freshness gate.

6. **Resolve the `CommitStartedPatch` vocabulary ambiguity.** State whether it remains a single builder serving both prepare and commit phases, or will be split. If single, add a note that prepare validation is a separate step.

7. **Classify budget/circuit-breaker keys and `last_woke_at`.** Add them to the mutation landscape with owner slice and exception status, or state explicitly that the reconciler owns them and they are outside the command extraction boundary.

8. **Add a bead-type discriminator strategy to the static guard specification.** State whether the guard enforces key names on session beads only (requiring typed commands for session-bead lifecycle keys) or uses a broader bead-type filter.

9. **Name expected new events per command cluster.** For slices 2 and 3 at minimum, list the session events that will be introduced.

---

## Questions

- Does the design intend `cmd/gc/session_name_lookup.go` to be part of slice 1 (target classification), slice 3 (close/identity retirement), or both? It writes both identity and create/start keys.
- Should `last_nudge_delivered_at` be inside the mutation boundary (nudge delivery is a session lifecycle event) or outside (it is an operational timestamp written by the nudge dispatcher)?
- Is `SESSION-WORK-003` (orphan pool step bead recovery) in scope for any current slice, or is it explicitly deferred?
- When slices 5 and 6 restore proofs, do the `REQUIREMENTS.md` rows keep their commit citations as historical anchors alongside the new test paths, or are commits retired as evidence?
- What is the intended end state for exported patch helpers (`RequestWakePatch`, `PreWakePatch`, `CommitStartedPatch`, `ClosePatch`, `RetireNamedSessionPatch`)? Are they unexported after all production callers in a field family route through commands, replaced with command-internal construction, or retained as building blocks behind a command API?
