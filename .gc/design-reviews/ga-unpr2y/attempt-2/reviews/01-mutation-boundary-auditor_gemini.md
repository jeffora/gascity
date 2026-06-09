# Elena Marchetti — DeepSeek V4 Flash (Independent Review, Attempt 2)

**Verdict:** approve-with-risks

The revision makes substantial progress on the mutation boundary. It adds a field-level ownership table, a mutation landscape inventory, a static-guard specification, a bounded exception list, command atomicity contracts, event/recovery ordering, a scenario traceability matrix, and per-slice proof gates. These are real, enforceable additions that move the boundary from aspirational prose toward an auditable contract.

The verdict is not `block` because the central structural gap — the absence of an enforceable mutation boundary — is now addressed in form, and the remaining issues are precision gaps and omissions that can be fixed without redesigning the boundary model. But the precision gaps are material: the inventory is incomplete, the static guard overstates what is statically decidable, key metadata fields are unclassified, and several production writer sites are missing from the landscape table. Each of these is a correctness risk for the guard that the design says must land before mutation-owning slices.

---

## Top Strengths

- The field-level mutation boundary (`DESIGN.md:67–77`) is the right granularity. Named key families let a guard check specific writes, let reviewers verify per-slice retirement, and let implementers know exactly which `SetMetadata`/`SetMetadataBatch` calls cross the boundary. This is a genuine improvement over attempt 1's prose-only boundary.
- The mutation landscape inventory (`DESIGN.md:90–117`) ties each writer cluster to an owner slice and exception status. This is the right structure: every row has a target path, and slices can be judged by whether their row is current and their exceptions have retired.
- The command atomicity contract (`DESIGN.md:260–276`) with per-operation tables for close, wake, drain, identity, and runtime start is a significant structural addition. The stale-fact, partial-write, and recovery-authority columns give implementers concrete contract terms rather than vague "validate current state" prose.
- The scenario traceability matrix (`DESIGN.md:209–248`) ties every slice to `SESSION-*` rows, current proof paths, allowed behavior changes, and freshness gates. This is what was missing in attempt 1 and it directly addresses the "stale proof" finding.

## Critical Risks

### [Major] The mutation landscape inventory omits production writers of session-owned keys

The inventory lists 13 areas. Verification against the current checkout finds production `SetMetadata`/`SetMetadataBatch` call sites writing session-owned keys that are not covered:

| Omitted site | Keys written | Approximate call count | Belongs to inventory row? |
|---|---|---|---|
| `cmd/gc/cmd_session.go:1592` | `held_until`, `sleep_intent`, `state` (metadata-only suspend) | 1 `SetMetadataBatch` | Not named. The "wait integration" row covers `cmd/gc/cmd_wait.go` but not `cmd_session.go` suspend. |
| `cmd/gc/cmd_stop.go:329` | `sleep_reason` | 1 `SetMetadata` | Not named. No row for CLI stop. |
| `cmd/gc/cmd_session_pin.go:125` | `pin_awake` | 1 `SetMetadata` | Not named. The "sleep, pin, wait-hold" row is too broad to verify. |
| `cmd/gc/cmd_prime.go:582` | `session_key` | 1 `SetMetadata` | Not named. The "reconciler" row does not cover `cmd_prime`. |
| `cmd/gc/session_name_lookup.go:192–230` | `state`, `generation`, `instance_token`, `session_name`, `alias` (create + rename) | Multiple `SetMetadata` + inline map | Not named. This is the primary session-materialization path but it has no row. |
| `cmd/gc/convergence_store.go:121` | `close_reason` on a convergence bead | 1 `SetMetadata` | Not named. A non-session domain writer of `close_reason` — exactly the kind of cross-domain collision the guard must handle. |
| `cmd/gc/nudge_dispatcher.go` (via `cmd/gc/cmd_nudge.go:988`) | `last_nudge_delivered_at` | 1 `SetMetadata` | Not named. The design does not classify `last_nudge_delivered_at` as inside or outside the boundary. |
| `cmd/gc/cmd_session_wake.go:83–90` | Six lifecycle keys in a raw map literal | 1 `SetMetadataBatch` | Named in the "cmd/gc/cmd_session_wake.go" row, but the row says "WakeSession plus fallback direct SetMetadataBatch for no-template wake" while the actual code clears `state`, `state_reason`, `pending_create_claim`, `pending_create_started_at`, `wake_request`, `wake_requested_at` — these overlap with lifecycle and create-start families, not just "wake". |

The inventory drives the guard baseline and per-slice retirement criteria. Omitted sites mean the guard will start with a baseline that undercounts bypasses, and slices 2–6 can be declared "ready" while these writers persist. Since the design says "a slice is not ready until its row is current" (`DESIGN.md:91`), every missing row is a readiness gap.

**Required change:** Add explicit rows for `cmd/gc/cmd_session.go` (suspend), `cmd/gc/cmd_stop.go`, `cmd/gc/cmd_session_pin.go`, `cmd/gc/cmd_prime.go` (session_key write), `cmd/gc/session_name_lookup.go` (materialization), `cmd/gc/convergence_store.go` (close_reason on convergence), and `cmd/gc/cmd_nudge.go`/nudge dispatcher (`last_nudge_delivered_at`). Group pin and stop under the "sleep, pin, wait-hold" row if desired, but name the exact files and call sites so reviewers can verify the baseline.

### [Major] The static guard is not statically decidable for key-aware writes

`DESIGN.md:143–155` requires the guard to "forbid production files outside `internal/session` and an explicit allowlist from calling `SetMetadata`, `SetMetadataBatch`, `Update`, `Create`, or `Close` when the call writes a session-owned key or applies a `session.MetadataPatch`."

The symbol-detection half (calling `session.MetadataPatch` builders and applying them via `SetMetadataBatch`) is statically decidable. The key-aware half ("when the call writes a session-owned key") is not, because production code builds `map[string]string` batches dynamically:

- `cmd/gc/session_reconciler.go:1715,1741,1859,1918` — multiple dynamic map constructions with `sessionpkg.RestartRequestPatch(...)` plus conditional keys, applied via `SetMetadataBatch`.
- `cmd/gc/session_reconcile.go:665` — dynamic `map[string]string{"restart_count": ...}`.
- `cmd/gc/cmd_session_wake.go:83–90` — raw map literal with six keys.
- `cmd/gc/cmd_session.go:1592` — raw map literal with three keys.
- `cmd/gc/session_lifecycle_parallel.go:1611,1712` — `CommitStartedPatch` with dynamic field additions.
- `internal/api/handler_session_create.go:472–481` — `persistSessionMeta` builds a batch from option metadata, then calls `SetMetadataBatch`.

A source scanner that checks symbol names and allowlisted files can detect *that* a `SetMetadataBatch` call occurs, but cannot determine *which keys* the map contains without data-flow analysis. The design should not claim the guard detects key-aware writes when it can only detect symbol-aware writes.

**Required change:** Restructure the guard specification into two enforceable parts:

1. **Symbol guard** (immediately enforceable, additive): Forbid production files outside `internal/session` and an explicit allowlist from importing or calling `session.*Patch` constructors and applying the result via `SetMetadataBatch`. This is a symbol check analogous to `TestGCNonTestFilesStayOnWorkerBoundary` and catches the escape-hatch pattern.
2. **Batch-key guard** (structural, per-slice): Require that production callers outside `internal/session` construct session-boundary batches only through typed patch builders or typed command methods, not through ad-hoc `map[string]string` literals. Enforce this structurally: when a slice introduces a command for a field family, the command must be the only way to construct a batch for that family outside `internal/session`. Verify by checking that no `map[string]string` literal in a non-test, non-exception file contains keys from the boundary family table.

Do not claim the guard detects "writes a session-owned key" without data-flow proof.

### [Major] The boundary scope for operational metadata is still incomplete

The design classifies five key families (lifecycle, create/start, wake/hold/drain, identity, runtime identity) but does not classify several keys that are written onto session beads by production code:

| Key | Writer(s) | Inside or outside boundary? |
|---|---|---|
| `last_nudge_delivered_at` | `cmd/gc/cmd_nudge.go:988`, `internal/session/manager.go:1526` (read) | **Unclassified.** Written by the nudge dispatcher outside `internal/session`. Constant defined in `internal/session/manager.go:69`. |
| `close_reason` (on convergence beads) | `cmd/gc/convergence_store.go:121` | **Must be outside** (it is convergence-domain, not session-domain), but the guard must not false-positive on it. The key name `close_reason` collides with the session-owned family. |
| `closed_at` (on extmsg membership beads) | `internal/extmsg/transcript_service.go:427` | **Must be outside**, same collision concern. |
| `state` (on convergence beads) | `cmd/gc/convergence_store.go:45–46` | **Must be outside**, same collision concern. |
| `restart_count` / `churn_count` / `wake_attempts` | `cmd/gc/session_reconcile.go:687,822,831`, `cmd/gc/session_reconciler.go:3013` | **Currently unclassified**. `restart_count` appears in the circuit-breaker row of the reconciler landscape but is not in the key-family table. `churn_count` and `wake_attempts` appear in patch builders but not in the family table. |
| `last_woke_at` | `cmd/gc/session_reconcile.go:615,767` clears it; `internal/session/lifecycle_transition.go:113,146,177,178,288,303,338,354,468` sets/clears it | **Unclassified.** Used in creating/start-requested projection but not named in the family table. |
| `detached_at` / `sleep_intent` | `cmd/gc/session_sleep.go:147,162,172,286,301,323` | Listed under "wake/hold/drain" family, but `sleep_intent` is also written by `cmd/gc/cmd_session.go:1594` and `cmd/gc/cmd_wait.go` without a named row. |

The guard cannot enforce a boundary whose membership is undefined. `close_reason`, `closed_at`, and `state` on non-session beads will false-positive any key-aware scanner. The reconciler counters need a classification decision.

**Required change:** Add a "Non-owned session-bead bookkeeping" classification (or equivalent) for keys like `restart_count`, `churn_count`, `wake_attempts`, `last_woke_at`, and `last_nudge_delivered_at` that are written to session beads but are not lifecycle or identity mutations. State explicitly whether each key is: (a) inside the boundary and moves behind a command in its owner slice, (b) outside the boundary as controller bookkeeping with a guard carve-out, or (c) TBD pending a specific slice. For `close_reason`, `closed_at`, and `state` on non-session beads, state that the guard's scope is limited to beads of type `session` (or equivalent type discriminator) and explain how the guard proves target-bead context.

### [Major] The exception policy is still intent-based, not structurally enforceable

`DESIGN.md:79–87` lists four exception classes:

1. tests and fixtures
2. doctor/migration/repair code that "normalizes historical broken state and emits trace/log evidence"
3. low-level bead-store conformance utilities
4. current root-documented active migration exceptions in `internal/api`

Problems:

- **Class 2 is intent-based.** A reviewer or CI guard cannot determine from code structure whether a production writer "normalizes historical broken state." The design should require that doctor/repair exceptions are annotated (file-level build tag, function-level comment with issue reference, or path-level allowlist entry) and that each exception has a retirement condition tied to a specific slice or issue.
- **Class 4 is correct but unbounded.** `internal/api/session_resolution.go:171` applies `RetireNamedSessionPatch` on the *normal* API resolution path (not in a doctor or repair context). The design lists it as an "active root-documented exception" but does not name its retirement slice explicitly. The landscape table row says "owner slice 1, 3" but the exception section says "until retired" without naming when.
- **No allowlist shape or CI enforcement is specified.** The static guard section (`DESIGN.md:143–155`) says the guard should "list every temporary exception with owner slice, reason, and retirement condition" but does not specify the allowlist format (file-level? function-level? build-tag-level?), who maintains it, or how CI fails when the allowlist grows.

**Required change:** Replace intent-based exception prose with a structural allowlist. For each exception, specify: file path, function or method name (or build tag), owner slice, reason, and retirement condition (slice number, issue number, or date). Tie allowlist entries to the guard so that a new `SetMetadataBatch` call in an allowlisted file is still detected and must be explicitly added. State that the allowlist is shrink-only: entries are removed when their owner slice completes, and the build fails if the allowlist grows without a corresponding design update.

### [Minor] `REQUIREMENTS.md` is missing scenario rows for functions used in `ProjectLifecycle` that the design depends on

The design's vocabulary table (`DESIGN.md:60–74`) depends on `ProjectLifecycle`, `ComputeAwakeSet`, and `LifecycleDisplayReason`, and the reconciler fact table depends on `projectContinuityEligibility`, `shouldResetContinuation`, and `countsAgainstCapacity`. None of these functions have `SESSION-*` rows in `REQUIREMENTS.md`:

- `projectContinuityEligibility` has a subtle inverted-allowlist implementation (`internal/session/lifecycle_projection.go:498`) that determines whether `closed`/`archived` beads are continuity-eligible. The orphaned + missing-config case is not covered.
- `shouldResetContinuation` determines stale-creating vs start-requested projection and is central to `SESSION-LIFE-004`, but its exact precondition (`pending_create_started_at` threshold) is not a scenario row.
- `countsAgainstCapacity` affects pool scaling and awake-set membership but has no explicit requirement row.
- `RuntimeProjectionUnknown` is not a scenario in `REQUIREMENTS.md`, yet the design's fact table depends on unknown runtime handling for destructive operations.

The design says "Technical refactors must preserve the behavior in `REQUIREMENTS.md`" but REQUIREMENTS.md does not cover the full behavior surface the design depends on. This is not a design bug, but it means the parity proof for slices 1, 4, and 6 will cite functions whose behavior is not scenario-gated.

**Required change:** Add scenario rows for `projectContinuityEligibility` (especially the orphaned + missing-config edge case), `shouldResetContinuation` (stale-creating threshold), `countsAgainstCapacity`, and `RuntimeProjectionUnknown` to `REQUIREMENTS.md`, or note in the design that these functions are implementation details of `ProjectLifecycle`/`ComputeAwakeSet` and parity is proven through the existing projection test surface.

---

## Cross-Document Consistency Issues

- **`last_woke_at` vocabulary gap between REQUIREMENTS.md and DESIGN.md.** `REQUIREMENTS.md` `SESSION-LIFE-004` mentions `pending_create_started_at` for stale-creating detection but not `last_woke_at`, which `projectRuntimeProjection` uses to distinguish `fresh-creating` from `start-requested` (`lifecycle_projection.go:475–480`). The design's reconciler fact table lists runtime liveness but does not name `last_woke_at` as a fact input. The key is written/cleared by the reconciler (`session_reconcile.go:615,767`) and used in projection, but it is not in any key family.
- **`REQUIREMENTS.md` `SESSION-RECON-006` (provider health gate) and `SESSION-RECON-007` (progress-aware health) cite proof that does not exist in this checkout.** The design's slice 6 acknowledges this (`DESIGN.md:543–548`) but the REQUIREMENTS rows still cite `cmd/gc/provider_health_gate_test.go` and `cmd/gc/session_progress_test.go`. The design and requirements are inconsistent: the design says "restore or replace the missing proof" but REQUIREMENTS.md treats those tests as current evidence.
- **`convergence_store.go:121` writes `close_reason` on a convergence bead, but REQUIREMENTS.md and DESIGN.md only discuss `close_reason` in the session lifecycle family.** The guard must distinguish session-bead `close_reason` from convergence-bead `close_reason`, or accept that key-aware enforcement is limited to typed patch builders.

---

## Missing Evidence

- A complete generated inventory of production session-owned metadata writers, including `cmd/gc/cmd_session.go` (suspend), `cmd/gc/cmd_stop.go`, `cmd/gc/cmd_session_pin.go`, `cmd/gc/cmd_prime.go`, `cmd/gc/session_name_lookup.go`, `cmd/gc/convergence_store.go`, and `cmd/gc/cmd_nudge.go`.
- A concrete static-guard implementation strategy that separates the symbol guard (immediately enforceable) from the batch-key guard (structurally enforceable per slice).
- A non-owned bookkeeping classification for `restart_count`, `churn_count`, `wake_attempts`, `last_woke_at`, `last_nudge_delivered_at`, and a discriminator strategy for `close_reason`/`closed_at`/`state` on non-session beads.
- A structural allowlist format (file, function, owner slice, reason, retirement condition) that CI can enforce as shrink-only.
- `REQUIREMENTS.md` scenario rows for `projectContinuityEligibility`, `shouldResetContinuation`, `countsAgainstCapacity`, and `RuntimeProjectionUnknown`.

---

## Required Changes

1. **Add the omitted mutation writer sites to the landscape table** with file, key families written, owner slice, and exception status. Specifically: `cmd/gc/cmd_session.go` suspend, `cmd/gc/cmd_stop.go`, `cmd/gc/cmd_session_pin.go`, `cmd/gc/cmd_prime.go` session_key write, `cmd/gc/session_name_lookup.go` materialization, `cmd/gc/convergence_store.go` `close_reason`, and `cmd/gc/cmd_nudge.go` `last_nudge_delivered_at`.

2. **Restructure the static guard into two enforceable parts:** a symbol guard that catches `session.*Patch` construction and application outside `internal/session` and an allowlist, and a structural batch-key guard that requires typed patch builders or command methods for session-bead mutation outside `internal/session`. Remove the claim that the guard detects "writes a session-owned key" without data-flow proof.

3. **Classify all unowned bookkeeping keys.** Add a "Non-owned session-bead bookkeeping" category for `restart_count`, `churn_count`, `wake_attempts`, `last_woke_at`, and `last_nudge_delivered_at`. State whether each is inside the boundary (moves behind commands), outside with a guard carve-out, or pending a specific slice decision. Add a note that `close_reason`, `closed_at`, and `state` on non-session beads require a bead-type discriminator in the guard.

4. **Replace intent-based exception prose with a structural allowlist.** For each exception, specify: file path, function or method name, owner slice, reason, and retirement condition. State that the allowlist is shrink-only and CI-enforced.

5. **Specify the exported patch helper end state.** Choose one of: (a) unexport after all production callers in a field family route through commands, (b) replace with command-internal mutation construction, or (c) retain as command-internal building blocks behind a command API that validates. Document the choice and add it to the static guard so that external `session.*Patch` construction shrinks as commands are adopted.

6. **Add `last_woke_at` to the create/start key family** in the boundary table, since it is written/cleared by the reconciler and read by `projectRuntimeProjection` to distinguish `fresh-creating` from `start-requested`.

7. **Add scenario rows to `REQUIREMENTS.md`** for `projectContinuityEligibility`, `shouldResetContinuation`, `countsAgainstCapacity`, and `RuntimeProjectionUnknown`, or state in the design that parity for these functions is proven through the existing projection test surface and no new rows are needed.

---

## Questions

- Does the design intend `cmd/gc/session_name_lookup.go` to be part of the slice 1 "target classification" migration, the slice 3 "close and identity retirement" migration, or both? It creates sessions with lifecycle and identity keys (lines 192–199) and mutates `session_name` (line 227), crossing two field families.
- Should `last_nudge_delivered_at` be inside the mutation boundary (nudge delivery is a session lifecycle event) or outside (it is an operational timestamp written by the nudge dispatcher, not a lifecycle transition)? The constant is defined in `internal/session/manager.go` but written from `cmd/gc/cmd_nudge.go`.
- The design says `cmd/gc/session_lifecycle_parallel.go` "must delegate as each start slice converts" — does that mean per-operation, per-field-family, or per-file? The file contains both start-path mutations (`CommitStartedPatch`) and close-path mutations (`ClosePatch`), and those belong to different slices (2 vs 3).
- When slice 1 (target classification) ships, should the guard detect `session.RetireNamedSessionPatch` construction in `internal/api/session_resolution.go` as an exception, or should the classifier wrap that call first and make the API path go through the classifier before calling the patch?
