# Session Boundary Design

| Field | Value |
|---|---|
| Status | Draft backlog |
| Behavior source | `REQUIREMENTS.md` |
| Scope | Technical architecture and extraction plan for session ownership |
| Latest design-review disposition | Attempt 3 global verdict `block`; this revision is an `iterate` response |

This document tracks the technical design for moving session behavior toward a
clearer ownership boundary. It is intentionally separate from
`REQUIREMENTS.md`: product behavior stays in the requirements ledger, while this
file holds design direction, backlog, and extraction sequencing.

## Attempt 3 Review Response

<!-- REVIEW: added per attempt-3-global-synthesis -->

Attempt 3 blocked on contracts that were too broad to decompose safely:
mutation ownership, target compatibility, scenario proof, command atomicity,
critical event recovery, reconciler fact ownership, slice coexistence,
operability, and shared vocabulary scope. This revision responds by making each
area a gate for implementation beads. It does not approve implementation; it
keeps the design in `iterate` until the inventories, tests, and guards below are
created or verified in code.

## Product Rule

Technical refactors must preserve the behavior in `REQUIREMENTS.md` unless code
inspection finds documented ambiguity, a real bug, or a product decision that
needs owner input. If behavior changes, update `REQUIREMENTS.md` with the new
scenario and proof.

Implementation is blocked for any slice whose cited proof is missing, stale, too
broad to show scenario parity, or not mapped to every scenario row the slice
touches. Replace the proof first, then extract the code.

## Problem

Session behavior is currently enforced across `internal/session`, `cmd/gc`,
`internal/api`, `internal/worker`, and runtime-facing adapters. That creates
poor encapsulation: callers can re-derive lifecycle, targeting, scaling,
work-release, and runtime-observation rules from raw metadata.

The target is not a broad session facade. The target is a small set of
session-owned decisions and commands that can be adopted one cluster at a time.

## Target Shape

Use a functional-core, imperative-shell model:

```text
Operation-specific fact reader -> Decider -> Session command/applier
    -> Post-commit facts/intents -> Subscribers/adapters
```

Roles:

- Fact reader: gathers immutable session, config, runtime, clock, and work facts
  needed by one operation.
- Decider: pure function that turns facts plus a requested operation into a
  typed decision.
- Session command/applier: re-reads or validates current state, applies durable
  session mutation, and returns post-commit session-domain facts and runtime
  intents.
- Subscribers/adapters: react to session-domain facts or runtime intents in
  their own domain.

The store is persistence, not a domain actor. It should not decide what session
events mean.

## Vocabulary And Existing Contracts

<!-- REVIEW: added per vocabulary-scope and existing-type-mapping -->

Avoid one broad `SessionFacts` type. Each slice introduces only the fact family
and result vocabulary needed for that operation. A shared type is allowed only
after two implemented deciders consume the same exact field set.

| Existing contract | Current type or path | Design term | Slice rule |
|---|---|---|---|
| Lifecycle projection | `LifecycleInput`, `LifecycleView`, `ProjectLifecycle` | Lifecycle facts/view | Reuse directly; do not replace in slice 1. |
| Wake selection | `AwakeInput`, `ComputeAwakeSet`, pool desired-state helpers | Awake facts/decision | Baseline for slices 5-7. |
| Named identity projection | `NamedIdentityInput`, `ProjectLifecycle` identity fields | Named identity facts | Reuse in target classification and close/retire slices. |
| Command transition table | `Transition`, `IllegalTransitionError`, `state_machine.go` | Transition validation | Commands must delegate to or prove parity with this table. |
| Metadata transition patches | `MetadataPatch`, `RequestWakePatch`, `PreWakePatch`, `CommitStartedPatch`, `ClosePatch`, `RetireNamedSessionPatch` | Patch builder implementation | May remain inside `internal/session`; external callers stop applying patch maps directly. |
| Resolution errors | `ErrSessionNotFound`, `ErrAmbiguous` | Target conflict result | Classifier wraps or maps these; callers keep compatible errors until a typed API change is approved. |
| Current event constants | `events.SessionWoke`, `SessionStopped`, `SessionCrashed`, `SessionDraining`, `SessionUndrained`, `SessionQuarantined`, `SessionIdleKilled`, `SessionMaxAgeKilled`, `SessionSuspended`, `SessionUpdated`, `SessionDrainAckedWithAssignedWork`, `SessionStranded`, `SessionWorkQueryFailed` | Session event facts | New events must be registered typed payloads and mapped to SSE/OpenAPI. |
| Reconciler diagnostics | `cmd/gc/session_reconciler_trace_*` | Decision/conflict trace | New decisions must map to centralized site/reason/outcome codes. |

Initial slice vocabulary:

- Slice 1 introduces target-classification facts, target-classification results,
  candidate conflict details, and a per-operation policy input.
- Slice 2 may introduce explicit-wake facts, wake command results, and wake
  conflicts.
- Later slices introduce lifecycle command facts, reconciler facts, and runtime
  intent results only when those slices begin.

Shared vocabulary checkpoints:

| Proposed vocabulary | First allowed caller | Required fields at introduction | Non-goals and expansion rule |
|---|---|---|---|
| `TargetCandidate` | Target classification adapter parity tests | kind, source surface, normalized token, session ID/name/alias/config identity, status, closed flag, conflict reason | Does not carry command mutation policy. Add fields only when a second surface needs the same field. |
| `TargetSelection` | API session target resolver adapter | selected candidate, ordered candidates considered, negative kind, retry/fallthrough directive | Not a universal session lookup facade. Each surface keeps its current resolver chain until parity tests approve delegation. |
| `SessionCommandConflict` | Explicit wake command | operation, session ID, precondition, current projection, retryable flag, trace reason | Not a catch-all error envelope for API or CLI formatting. Adapters map it to existing typed output. |
| `RuntimeStartIntent` | Runtime-start prepare/commit slice | bead ID, template, provider, work dir, generation, instance token, session key, config hash | Does not own provider execution or scheduling. The reconciler/worker remains the imperative shell. |
| `SessionFactEvent` | First post-commit command subscriber | event type, bead ID, generation/instance token when relevant, idempotency key, committed facts | Not a durable outbox until an outbox is explicitly designed. Critical recovery still scans durable facts. |

Flat optional envelopes are not acceptable for new shared types. Use tagged
result kinds or per-kind structs when only some fields are meaningful. A slice
must document the first delegated caller and tests before moving a vocabulary
type from private to shared.

## Mutation Boundary

<!-- REVIEW: added per mutation-boundary-enforceability -->

`internal/session` owns durable session lifecycle and identity mutation. In
production code outside `internal/session`, do not write session lifecycle or
identity metadata directly with bead-store calls.

Session-owned mutation includes these field families. The guard must treat exact
keys and documented prefixes as owned; unknown dynamic keys are unsafe when the
target bead may be a session bead.

| Family | Exact keys and prefixes | Notes |
|---|---|---|
| Lifecycle state | `state`, `sleep_reason`, `state_reason`, `closing`, `closed_at`, `archived_at`, `close_reason`, `close_detail`, `last_state_change_at` | Includes compatibility states that project through `ProjectLifecycle`. |
| Create/start lease | `pending_create_claim`, `pending_create_started_at`, `start_requested_at`, `creation_complete_at`, `generation`, `instance_token`, `continuation_epoch`, `continuation_reset_pending`, `continuation_reset_at`, `continuation_reset_reason` | `instance_token` is the runtime-start attempt identity and must be owned by one runtime-start command slice. |
| Runtime identity | `session_key`, `transport`, `provider`, `provider_kind`, `builtin_ancestor`, `started_config_hash`, `started_command_hash`, `live_command_hash`, `started_*_hash`, `live_*_hash`, stored MCP snapshot keys | These keys decide whether a provider runtime belongs to the bead. |
| Wake/hold/drain | `wake_request`, `wake_requested_at`, `wake_mode`, `last_woke_at`, `pin_awake`, `held_until`, `quarantined_until`, `wait_hold`, `sleep_intent`, `drain_*`, `handoff_*`, `assigned_work_*` | Work assignment updates remain outside session, but drain metadata on session beads is session-owned. |
| Identity | `session_name`, `session_name_explicit`, `alias`, `alias_history`, `configured_named_session`, `configured_named_identity`, `configured_named_mode`, `session_origin`, continuity eligibility, retired identity markers | Ordinary config names and `template:<name>` remain config/factory targets, not live session identities. |
| Operational lifecycle evidence | provider-health episode markers, progress/restart/circuit-breaker metadata, detached-probe metadata, startup verification metadata | Ownership depends on the slice. Until extracted, these are reconciler-owned exceptions with explicit proof requirements. |

Patch helpers may remain internal implementation building blocks and test
fixtures. Production callers outside `internal/session` should call
session-owned commands instead of receiving a patch map and applying it
themselves.

Exceptions:

- tests and fixtures that construct explicit states
- doctor/migration/repair code that normalizes historical broken state and
  emits trace/log evidence for each direct repair write
- low-level bead-store conformance utilities unrelated to production session
  behavior
- current root-documented active migration exceptions in `internal/api`, until
  their slice retires them

### Canonical Production Writer Inventory

This inventory is the working baseline. A slice is not ready until this table is
updated from source inspection and every converted row has a retirement
condition. Rows marked "exception" must appear in the static guard allowlist
with the same ID, owner slice, and expiry.

| ID | Current mutation path | Field family | Target path | Owner slice | Exception status |
|---|---|---|---|---|---|
| W-001 | `internal/session/manager.go` create/start/close methods | lifecycle, create/start, identity, runtime identity | Session commands implemented here or below this package | all command slices | Canonical owner, not an exception. |
| W-002 | `internal/session/lifecycle_transition.go` patch builders | lifecycle, create/start, wake/hold/drain, identity | Keep as internal patch builders behind commands | all lifecycle slices | Internal only. |
| W-003 | `internal/session/waits.go` wake/wait updates on session beads | wake/hold/wait | Explicit wake and wait-hold commands | 2, 5 | Internal owner or converted command helper. |
| W-004 | `cmd/gc/cmd_session_wake.go` no-template wake fallback | wake, lifecycle | Explicit wake command/applier | 2 | Exception until slice 2 delegates. |
| W-005 | `cmd/gc/session_lifecycle_parallel.go` prepare/commit/rollback/runtime identity writes | create/start, lifecycle, runtime identity | Runtime-start command with provider shell | 3 | Exception until slice 3 owns all runtime-start keys. |
| W-006 | `cmd/gc/session_wake.go` drain cancellation/completion and wake-blocker updates | wake/hold/drain | Wake/hold/drain command set | 5 | Convert as one cluster. |
| W-007 | `cmd/gc/session_sleep.go`, `cmd/gc/cmd_session_pin.go`, wait integration | sleep, pin, wait-hold | Hold/pin/sleep commands | 5 | No new direct writers. |
| W-008 | `cmd/gc/session_beads.go` and close/sweep callers | close metadata, bead close, identity retirement, assigned-work release | Close/retire command plus work subscriber scan | 4 | Work release remains outside session; close mutation moves in. |
| W-009 | `cmd/gc/session_reconciler.go`, `cmd/gc/city_runtime.go`, pool desired-state helpers | lifecycle repair, drift, scale, provider health, progress, drain/restart decisions | Reconciler fact readers plus command calls | 6, 7 | Partial-read rules required before conversion. |
| W-010 | `cmd/gc/session_circuit_breaker.go` | circuit/restart signatures, assigned-work resolver keys | Reconciler-owned circuit state until a narrow command needs it | 7 | Exception; session must not absorb scheduling policy. |
| W-011 | `cmd/gc/soft_reload.go`, `cmd/gc/cmd_nudge.go`, `cmd/gc/cmd_bd_store_bridge.go` if they write owned keys | wake/nudge/runtime evidence or generic bridge writes | Inventory each concrete key before delegation | affected slice | Exception only for verified keys; unknown dynamic batches are blocked. |
| W-012 | `cmd/gc/cmd_prime.go` live-session `session_key` priming | runtime identity | Runtime-start/identity repair command or documented prime-only repair | 3 | Exception until a repair command exists. |
| W-013 | `cmd/gc/pool_session_name.go` detached-probe and session-name updates | runtime evidence, identity | Reconciler fact ownership or identity command | 6, 7 | Exception until ownership is decided. |
| W-014 | `internal/api/session_resolution.go` materialization, retirement, reassignment | identity, create/start | Target classifier plus worker/session command delegation | 1, 3, 4 | Root-documented exception until retired. |
| W-015 | `internal/api/session_manager.go` | manager construction for API handlers | Command factory or worker boundary | 4 | Root-documented exception until retired. |
| W-016 | `internal/api/handler_session_create.go` and Huma session command handlers | create/alias/session-name setup | Typed policy adapter over worker/session command | 1, 3, 4 | Must preserve Huma typed-wire behavior. |
| W-017 | `internal/worker/handle_lifecycle.go` | worker-boundary create/close delegation | Canonical production boundary | all command slices | Allowed boundary, not a bypass. |
| W-018 | `internal/session/mcp_state.go` | stored MCP snapshot lifecycle | Runtime identity/close command helper | 3, 4 | Internal helper only. |
| W-019 | `internal/mail/beadmail.go` `RepairEmptyType` during session reads | repair empty session type | Repair helper with trace/log evidence if promoted | repair slice | Bounded repair exception. |
| W-020 | `internal/extmsg/*`, mail stores, convoy/order stores | non-session bead metadata and closes | Stay outside session unless writing a session bead owned key | none | Guard must prove non-session bead discrimination. |
| W-021 | Generic `beads.Store.Update`, `SetMetadata*`, `Close`, `Create` bridges | unknown | Must be classified by target bead type and key set | per slice | Unsafe by default when a session bead can flow through. |

### Static Guard

Before a slice is considered implemented, add or tighten a failing-build guard
that rejects new production bypasses. The guard should be a Go AST/symbol guard,
not a comment convention.

Guard inputs:

- owned key taxonomy from this document, mirrored in one central test fixture
- writer allowlist keyed by inventory ID (`W-004`, `W-005`, ...)
- package roots to scan: `cmd/gc`, `internal/api`, `internal/worker`,
  `internal/mail`, `internal/extmsg`, and every package outside
  `internal/session` that imports `internal/session`
- excluded files: tests, fixtures, generated files, migrations, and explicit
  doctor/repair utilities

Guard behavior:

- Flag production calls to `SetMetadata`, `SetMetadataBatch`, `Update`,
  `Create`, or `Close` when the receiver is a `beads.Store`, a wrapper around a
  bead store, or an interface value that can carry a session bead.
- Flag direct application of `session.MetadataPatch` or maps returned by
  `RequestWakePatch`, `PreWakePatch`, `CommitStartedPatch`, `ClosePatch`, or
  `RetireNamedSessionPatch` outside `internal/session` and approved worker/API
  transition points.
- Inspect string-literal keys in map literals and helper calls. If the key is
  dynamic and the target bead might be a session bead, require an allowlist row
  with owner slice, reason, and retirement condition.
- Treat generic bridge code as unsafe unless the code first proves the target
  bead is not a session bead or proves the written keys are outside the owned
  taxonomy.
- Require shrink-only allowlist behavior: an implementation slice may delete or
  narrow its own rows; adding a row needs a matching update to this design or
  `AGENTS.md` migration notes.

Session-bead discrimination is part of the guard. A direct store write is not a
session-boundary violation when the guarded path can only reach mail, extmsg,
convoy, order, or ordinary work beads and does not apply a session-owned key.
Unknown target type plus owned or dynamic keys is a violation.

The existing `cmd/gc/worker_boundary_import_test.go` remains the worker-boundary
guard; this boundary guard is additive and must not weaken it.

## Target Classification Contract

<!-- REVIEW: added per target-classification-compatibility -->

Slice 1 is the first extraction. It must not impose one new global precedence
order over surfaces that currently behave differently. The first implementation
is a candidate collector plus compatibility adapters that reproduce each
existing resolver chain.

Candidate kinds:

| Kind | Meaning | Required evidence |
|---|---|---|
| `direct-session-id` | Identifier loads a session or repairable session bead by ID | Include bead status and whether the ID is closed. |
| `live-session-name` | Open exact `session_name` match | Preserve dual alias/session-name demotion from `ResolveSessionID`. |
| `live-alias` | Open exact current `alias` match | Lower than live `session_name` in the package resolver. |
| `configured-named-canonical` | Configured named identity already has its canonical bead | Include identity, canonical bead ID, and config match. |
| `configured-named-reserved` | Config declares a named identity with no canonical bead | Include whether the operation may materialize. |
| `configured-named-conflict` | Live bead conflicts with configured named identity | Map to the existing configured-name conflict error. |
| `rejected-by-config` | A named session bead exists but the backing config no longer declares it | Map to existing `errSessionTargetRejectedByConfig` behavior. |
| `path-alias` | API path alias/title match | Include state filter result and most-recent-created tiebreaker evidence. |
| `closed-session-name` | Closed exact `session_name` match | Read-only only unless a requirement says otherwise. |
| `closed-alias` | Closed exact current `alias` match | Read-only only. |
| `historical-alias` | Alias-history match | Mail/query or inspect-only where current code allows it. |
| `template-factory` | `template:<name>` target | Factory/config target, not a live session. |
| `ordinary-config-target` | Bare configured agent/template target | Config target, not a live session. |
| `not-found` / `ambiguous` | No compatible candidate, or multiple equal-precedence candidates | Include candidates and fallthrough/retry directive. |

Compatibility resolver chains:

| Surface | Required chain | Materialization and closed behavior |
|---|---|---|
| `internal/session.ResolveSessionID` | exact session ID -> open `session_name` with dual-match demotion -> open `alias` -> not found/ambiguous | No config, template, path alias, or historical fallback. |
| `internal/session.ResolveSessionIDAllowClosed` | normal package resolver -> closed `session_name` -> closed `alias` | Read-only compatibility lookup after live matches fail. |
| API/Huma live session target | reject `template:` -> exact session ID -> configured named lookup -> generic package resolver -> path alias -> not found | Configured named lookup may return canonical, conflict, rejected-by-config, or materialize only when the caller sets materialize. |
| API/Huma allow-closed target | API live chain -> reject unmaterialized configured named target -> package allow-closed fallback | Exact closed IDs can resolve, then operation policy may reject them. |
| API path alias | title match on active/awake/empty-state non-named session beads only | Most-recent-created wins; named-session beads are skipped; path alias never steals a prior API-chain match. |
| Mail send | `human` -> live configured-named mailbox basename -> API live target without materialization -> configured named mailbox address -> not found | Configured named mailboxes do not materialize sessions. Existing live named mailbox display wins for bare basename matches. |
| Mail query | `human` -> configured named identity recipient set -> API live target -> raw recipient fallback | Includes historical alias recipients where `apiSessionMailboxAddresses` currently does. |
| Extmsg member notification | existing target resolution -> materialize configured named session on miss -> nudge resolved session | This is a current extmsg behavior and must not be silently removed. |
| Extmsg display handle | package allow-closed lookup for display label -> fallback selector label | No materialization. |
| Attach/observe/log/transcript | existing target or allow-closed helper for that surface | Preserve provider-specific transcript fallback and closed lookup rules. |
| Assignee normalization and circuit-breaker resolution | bead ID/session-name/alias/configured identity resolver used by the current call site | Must preserve ambiguity and duplicate-key handling before centralization. |

Policy inputs are explicit and do not silently reorder a surface's compatibility
chain: `surface`, `allow_closed`, `materialize_named`, `allow_path_alias`,
`allow_historical_alias`, `read_only`, `allow_template_factory`,
`allow_ordinary_config_target`, and `reject_config_missing_named`. Policy can
accept, reject, materialize, or fall through; it cannot invent a candidate.

Typed result contract:

- one selected candidate or one negative result
- ordered candidates considered by the active compatibility chain
- optional all-candidates diagnostic snapshot for ambiguity output
- `negative_kind`: `not-found`, `ambiguous`, `forbidden-kind`,
  `requires-materialization`, `configured-name-conflict`, `rejected-by-config`,
  `closed-not-allowed`, or `ordinary-config-target`
- `fallthrough`: `stop`, `try-next-compatible-chain-step`, `retry-after-store`

Adversarial parity tests must cover same-token collisions across bead ID,
session_name, alias, configured named identity, configured conflict,
rejected-by-config named bead, path alias, closed lookup, historical alias, bare
config target, and `template:` target for every surface listed above.

## Scenario Traceability Matrix

<!-- REVIEW: added per scenario-parity-proof -->

Every implementation bead spawned from this design must carry the rows it
touches. "Current proof" must be a runnable test, source path, issue, or commit
that exists in the checkout at implementation time. Proof against one row is
not enough when a slice touches several rows; every touched row needs exact
evidence before code moves.

| Slice | Scenario rows | Current proof | New or updated proof | Allowed behavior change | Freshness gate |
|---|---|---|---|---|---|
| 1 Target classification | `SESSION-ID-003`, `SESSION-ID-004`, `SESSION-ID-005`, `SESSION-ID-007`, `SESSION-ID-008`, `SESSION-ID-009`, `SESSION-ID-010`, `SESSION-RUNTIME-005` where transcript lookup is touched | `internal/session/resolve_test.go`, `internal/session/named_config_test.go`, `internal/api/session_model_phase0_interface_spec_test.go`, `internal/api/session_resolution_path_alias_test.go`, `internal/api/handler_mail.go`, `internal/api/handler_extmsg.go` | Classifier candidate tests plus API, CLI, mail, extmsg, assignee, nudge, attach, inspect/log/transcript parity tests | None without `REQUIREMENTS.md` update | Tests must assert exact errors, result kinds, configured-name conflict, rejected-by-config, materialization, path-alias tiebreakers, and no factory fallback. |
| 2 Explicit wake command | `SESSION-LIFE-007`, `SESSION-START-003`, `SESSION-START-007` | `internal/session/waits_test.go`, `internal/session/lifecycle_transition_test.go`, `internal/api/session_model_phase0_lifecycle_spec_test.go` | Command/applier conflict, stale-fact, no-template fallback, wait-cancellation, and API/CLI response tests | None except documented bug fix | Proof must include terminal closed, archived, suspended, drained, no-template, pending-create, and wait cancellation cases. |
| 3 Runtime start prepare/commit/rollback | `SESSION-START-001`, `SESSION-START-002`, `SESSION-START-008`, `SESSION-LIFE-004`, `SESSION-LIFE-005`, `SESSION-RUNTIME-001`, `SESSION-RUNTIME-003` | `internal/session/lifecycle_transition_test.go`, `internal/session/manager_test.go`, `cmd/gc/session_lifecycle_parallel_test.go`, `cmd/gc/session_reconcile_test.go`, `internal/session/submit_test.go` | Runtime-start command tests for provider-start success followed by commit failure, crash after prepare, stale generation, stale/mismatched instance token, and partial metadata repair | None | Runtime-start keys have one owner; pending-create clear, runtime identity, config hashes, prepare, commit, and rollback cannot be split across slices. |
| 4 Close and identity retirement | `SESSION-START-005`, `SESSION-START-006`, `SESSION-WORK-001`, `SESSION-WORK-002`, `SESSION-ID-006`, `SESSION-ID-007` | `internal/session/manager_test.go`, `internal/api/session_model_phase0_lifecycle_spec_test.go`, `cmd/gc/session_beads_test.go`, commits cited in `REQUIREMENTS.md` | Close command tests, idempotent work-release scan tests, API/CLI error parity, missed-event recovery tests | None without owner approval | Must prove provider-stop failure leaves bead open and successful close releases assigned work without relying solely on events. |
| 5 Wake/hold/drain eligibility and commands | `SESSION-LIFE-003`, `SESSION-LIFE-004`, `SESSION-START-003`, `SESSION-RECON-004`, `SESSION-WORK-004` | `cmd/gc/session_reconcile_test.go`, `internal/session/lifecycle_projection_test.go`, `internal/session/waits_test.go`, `cmd/gc/session_wake_test.go` | Pure decider tests for unknown/partial/stale facts plus drain-cancel and drain-ack recovery parity | None | Destructive drain/close decisions must fail closed on partial work or runtime facts. |
| 6 Reconciler fact extraction | `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-005`, `SESSION-START-008` | `cmd/gc/build_desired_state_test.go`, `cmd/gc/session_lifecycle_parallel_test.go`; `cmd/gc/scale_from_zero_test.go` is cited by requirements but absent in this checkout | Restore or replace scale-from-zero proof before extraction, then add fact-reader complexity tests | None | Missing proof blocks this slice; controller scheduling policy stays controller-owned. |
| 7 Provider health, progress, and circuit state | `SESSION-RECON-006`, `SESSION-RECON-007`, `SESSION-RUNTIME-001`, `SESSION-RUNTIME-002` | `cmd/gc/session_reconciler_test.go`, `internal/session/manager_test.go`; `cmd/gc/provider_health_gate_test.go` and `cmd/gc/session_progress_test.go` are cited by requirements but absent in this checkout | Restore or replace provider-health/progress proof, then add stale runtime fact, alert-dedup, and budget tests | None | Missing proof blocks this slice; health/progress/circuit side effects remain reconciler-owned until a narrower command is proven. |

Citation freshness gate:

- Add a test or script that fails when a path cited in this matrix or
  `REQUIREMENTS.md` does not exist and is not an issue or commit reference.
- Mark a proof stale when it no longer asserts the row's required behavior, even
  if the file exists.
- Characterization tests must be committed before moving behavior out of
  reconciler, manager, API, CLI, mail, extmsg, worker, or session command paths.

Proof commands:

- Unit parity for changed packages: targeted `go test` package commands named
  in the slice.
- Broad baseline before handoff: `make test` or the sharded target from
  `TESTING.md`.
- API/schema changes: `make dashboard-check` and OpenAPI/client regeneration
  checks required by root `AGENTS.md`.

## Command Atomicity Contract

<!-- REVIEW: added per command-atomicity-stale-fact-defense -->

Session commands must defend against stale facts by re-reading or validating a
precondition immediately before commit. The selected cross-process model for
the migration is versioned commit markers plus idempotent repair:

- In-process locks may reduce duplicate work but are not a correctness
  mechanism across CLI, API, controller, and reconciler processes.
- `session.WithCitySessionIdentifierLocks` protects local alias/session-name
  creation races only; it is not the general mutation lock.
- Store-level transactions or CAS are not assumed unless a slice proves the
  specific store path provides them.
- Commands write prepare markers before provider side effects, re-read current
  facts before commit, and reject or repair when markers changed.
- Runtime-start uses `instance_token` as the authoritative attempt identity;
  `generation` is supporting sequence evidence, not sufficient alone.
- Multi-key metadata writes must be proven atomic for the active store or every
  visible partial state must have a repair row and test.

A command that cannot validate its preconditions must reject with a typed
conflict rather than write a best-effort partial transition.

| Command cluster | Precondition fields | Written keys | Ordering/atomicity assumption | Conflict reasons | Retryability and repair proof |
|---|---|---|---|---|---|
| Target classification | none; read-only snapshot plus policy | none | no mutation | ambiguous, not-found, forbidden-kind | Retry after config/store change. |
| Request explicit wake | bead status, projected lifecycle, held/quarantine, named identity, configured template presence, pending-create state | `wake_request`, blockers, optional dormant state normalization | re-read bead status/projection before one metadata batch | terminal, missing-config, identity-conflict, stale-state, pending-create-in-flight | Retriable when stale; tests race close, drain, and pending-create vs wake. |
| Prepare runtime start | state, generation, instance token, pending-create claim, current runtime liveness, config hash, budget/health gate result | `state=creating`, `generation`, `instance_token`, `pending_create_started_at`, reset keys, runtime-start hash markers | write prepare metadata before provider start; token identifies the attempt | stale-generation, stale-instance-token, already-running, provider-health-red, budget-exhausted, missing-config | Retriable if runtime absent and token/generation still match; repair rolls back stale creating. |
| Commit runtime start | generation, instance token, runtime identity, provider start result, pending-create marker | active/start-complete keys, `session_key`, hash fields, pending-create clear | re-read prepared bead; commit only when token matches or legacy no-token compatibility path applies | stale-generation, runtime-mismatch, pending-create-cleared, partial-start | Recovery can commit or roll back after observing runtime identity and token. |
| Roll back runtime start | generation, instance token, provider failure/crash marker, runtime liveness | failed-create/asleep state, pending-create clear, runtime identity preservation or clear per row | token must still identify the failed attempt | stale-token, runtime-alive-after-failure, partial-rollback | Retriable; repair converges from stale creating or failed-create. |
| Close session | bead open, lifecycle not already closed, provider stop outcome when required, work-release subscriber availability | close metadata, bead close, identity retirement keys | stop succeeds before bead close when stop is required; work release happens after close from durable facts | provider-stop-failed, stale-state, already-closed, work-release-deferred | Retriable; stop failure leaves bead open, successful close is idempotent. |
| Retire named identity | closed/terminal lifecycle, continuity eligibility, duplicate canonical facts | alias/session_name clears, retired markers | same command as terminal close or follow-up idempotent repair | duplicate-canonical, stale-identity, not-terminal | Retriable and idempotent. |
| Drain ack/complete/cancel | drain generation, assigned-work facts, runtime ack source, work-query completeness | drain state/reason keys, cleared ack metadata | reread assigned work before destructive close or no-wake completion | assigned-work-present, stale-drain-generation, partial-work-query | Retriable; partial facts cancel destructive action. |

Partial external-store failure must be handled as an explicit command result.
For multi-key metadata writes, tests must prove the store applies the batch as a
unit or that a repair path converges from every visible partial state.

## Runtime Intent, Event, And Recovery Ordering

<!-- REVIEW: added per runtime-ordering and event-delivery-convergence -->

No safety-critical reaction may depend only on at-most-once in-process event
delivery. Until a durable session outbox exists, critical reactions must be
driven by durable session facts plus idempotent controller scans or retained
synchronous cascades.

| Operation | Required order | Recovery authority | Subscriber class |
|---|---|---|---|
| Explicit wake request | validate -> commit wake metadata -> optional event/trace -> reconciler scan starts runtime | Durable wake metadata and controller scan | critical retryable |
| Runtime start | commit creating/generation/token -> start provider -> commit active/session key -> event/trace | Pending-create metadata, generation, provider runtime identity | critical retryable |
| Close | provider stop success when required -> commit close/retire facts -> release work scan -> event/trace | Closed bead and session metadata; work-release scan | critical retryable |
| Drain | commit drain intent -> worker ack -> reread assigned work -> complete/cancel | Drain metadata plus live work query | critical retryable |
| Identity retirement | close/retire metadata -> assignment and binding scans -> event/trace | Retired identity metadata and open work/binding queries | critical retryable |
| Trace/SSE/dashboard | emit after commit | Event log and trace store when available | observability-only |

Event contract:

- Event payloads describe facts that happened, not commands to attempt.
- Payloads carry stable IDs, generation/instance token when relevant, and
  idempotency keys for subscriber scans.
- Existing `events.*` constants remain the first mapping target. A new
  session event requires `events.RegisterPayload`, OpenAPI/SSE projection
  updates, and `TestEveryKnownEventTypeHasRegisteredPayload` parity.
- Duplicate, skipped, out-of-order, and crash-after-commit delivery must leave
  critical state recoverable from durable facts.

Per-event reaction matrix:

| Event/fact | Tier | Required payload fields | Idempotency/supersession | Durable recovery authority | Tests |
|---|---|---|---|---|---|
| wake requested / `events.SessionWoke` | accelerator + observability | session ID, target identity, generation/token if start is prepared, wake cause | Superseded by newer generation/token or terminal bead status | wake metadata plus controller desired-state scan | missed-event still starts or remains queued; stale event ignored |
| runtime stopped / `events.SessionStopped` | critical diagnostic, recovery accelerated by scan | session ID, runtime key, stop reason, generation/token when known | Superseded by newer live runtime identity or closed bead | provider observation plus session bead state | event miss does not strand assigned work |
| runtime crashed / `events.SessionCrashed` | critical diagnostic, recovery accelerated by scan | session ID, runtime key, crash reason, captured output ref, generation/token | Superseded by newer token or repaired state | provider observation, crash metadata, restart budget state | duplicate crash does not double-consume budget |
| drain state / `events.SessionDraining`, `events.SessionUndrained` | accelerator + observability | session ID, drain generation, reason | Superseded by drain generation or terminal state | drain metadata and assigned-work query | missed event still completes or cancels by scan |
| `events.SessionDrainAckedWithAssignedWork` | diagnostic signal, not recovery authority | session ID, assigned bead ID, template, bead status, reason | Duplicate key is session ID + bead ID + drain generation | durable assigned-work scan and wake-demand recomputation | event emitted once per episode; scan recovers if event missing |
| quarantine/idle/max-age/suspend/update events | observability + policy evidence | session ID, reason, policy source, projection facts | Superseded by newer projection timestamp/generation | lifecycle metadata and reconciler policy scan | stale events do not override current projection |
| `events.SessionStranded` | diagnostic | session ID, assigned work refs, probe source, reason | Duplicate key is session ID + work refs + probe generation | work/session scan | diagnostic can be missed without changing recovery |
| `events.SessionWorkQueryFailed` | diagnostic + fail-closed evidence | store ref, query class, error, session ID if scoped | Newer successful query supersedes diagnostic | next work query and controller scan | destructive action suppressed while query is partial/failed |

`SessionDrainAckedWithAssignedWork` is a diagnostic and acceleration signal. The
load-bearing recovery rule is the durable assigned-work scan; no implementation
may require that event to fire for work release, drain cancellation, or wake
demand to converge.

Subscriber classification:

| Domain | Classification | Contract |
|---|---|---|
| Work release and orphan assignment recovery | critical retryable | Scan durable work and session facts idempotently; do not depend only on an event. |
| Runtime reconciliation | critical retryable | Rebuild desired state from session/config/runtime facts each tick. |
| Mail/extmsg binding cleanup | retryable | Use session facts and binding records; event accelerates but is not sole authority. |
| Trace, logs, SSE, dashboard | observability-only | Missed event is a diagnostic loss, not a state convergence loss. |

## Reconciler Fact Contract

<!-- REVIEW: added per reconciler-runtime-facts -->

The reconciler migration baseline is the existing pure decision surface:
`ComputeAwakeSet`, `ProjectLifecycle`, and current pool desired-state code. Do
not replace those with broader abstractions until a narrow slice proves parity.

Controller scheduling and capacity policy remain controller-owned. In
particular, `ComputePoolDesiredStates` currently combines request tiers, nested
caps, in-flight reuse, named-session exclusion, scale-check demand, work
aggregation, and capacity decisions. A session slice may own lifecycle
eligibility facts or projection helpers, but it must not move demand shaping,
slot allocation, scale-check policy, or restart-budget consumption into
`internal/session` unless requirements are updated.

| Fact family | Source/provenance | Unknown or partial handling | Owner |
|---|---|---|---|
| Session bead snapshot | bead store list/get with partial-query marker | Never perform destructive close/drain/rollback from a partial snapshot | adapter |
| Lifecycle projection | `ProjectLifecycle` over snapshot metadata | Unknown runtime projects through existing rules; terminal bead status wins | `internal/session` |
| Config and named identities | resolved city config plus named-session config | Missing config blocks wake; config read failure blocks mutation | adapter |
| Runtime liveness | provider list/probe/process observation | Stale/unknown runtime cannot prove death for destructive action; may preserve existing fail-open health behavior where required | runtime adapter/reconciler |
| Work demand and assignments | active store and rig stores, with store refs | Partial work query cancels no-wake drains and suppresses destructive orphan cleanup | work/reconciler adapter |
| Provider health | health observations and episode state | Absent/stale/unknown health fails open per `SESSION-RECON-006`; red suppresses respawn without consuming budget | reconciler |
| Progress signatures | transcript/progress observation | Unknown does not kill; stale over threshold may drain only with complete exemption facts | reconciler |
| Circuit breakers/restart counts/budgets | durable metadata plus in-memory budget state where current code requires it | Unknown budget state prevents additional destructive restart | reconciler |

State ownership:

| State | Immutable session fact? | Stateful controller fact? | Boundary |
|---|---|---|---|
| terminal lifecycle, identity projection, wake blockers | yes | no | `internal/session` projection and commands own semantics. |
| runtime liveness observation | no | yes, from provider/runtime adapter | Session deciders consume snapshots; providers gather them. |
| provider health episode and alert dedupe | no | yes | Reconciler owns fail-open/red behavior and alert dedupe. |
| progress signature and inactivity threshold | no | yes | Reconciler owns transcript/progress observation and exemptions. |
| restart count, circuit breaker, budget | no | yes | Reconciler owns budget accounting; session may receive conflict facts. |
| pool demand, work aggregation, scale checks, nested caps | no | yes | Controller/reconciler policy; not a session primitive. |

Desired-running precedence must remain explicit:

1. terminal/closed/archived-not-continuity-eligible blocks wake
2. holds/quarantine/missing config/identity conflict produce desired-blocked
3. pending create, pin, attachment, pending interaction, named-always, targeted
   work, scale demand, and explicit wake produce desired-running when not
   blocked
4. pool demand is clamped and budgeted by existing pool desired-state rules
5. provider-health red suppresses respawn but does not mark the identity
   terminal

## Worker Boundary And Migration Sequencing

<!-- REVIEW: added per worker-boundary-migration-collision -->

The active worker-boundary migration remains authoritative for production
`cmd/gc` lifecycle operations. Session-owned commands do not give `cmd/gc`
permission to bypass `worker.Handle`; they define the validating path that the
handle, API adapters, and reconciler should delegate to.

Each slice phases in this order:

1. Document exact scenario rows, writer inventory, and proof commands.
2. Introduce the typed decider and command behind existing behavior.
3. Delegate the first caller without changing user-visible behavior.
4. Delegate all production writers for that field family, or make every
   surviving writer call the same command/applier.
5. Tighten static guards and retire exceptions.
6. Remove dead helper paths after proof passes.

Per-slice coexistence gates:

| Slice | Converted callers | Legacy callers allowed during bake | Validation differences allowed | Guard update | Bake and revert rule |
|---|---|---|---|---|---|
| 1 Target classification | API target resolver adapter first; then mail/extmsg/CLI/assignee helpers one surface at a time | Existing resolver functions stay as oracle until parity tests pass | None; exact errors and materialization behavior preserved | Read-only guard prevents new resolver copies after adoption | Revert by switching adapter back to old resolver; no metadata migration. |
| 2 Explicit wake | `cmd/gc/cmd_session_wake.go`, API wake path, wait/no-template fallback | Reconciler implicit wake and runtime-start code untouched | None except approved bug rows | Retire W-004 exception when converted | Bake requires trace/API/CLI parity and close-vs-wake race tests. |
| 3 Runtime start | `cmd/gc/session_lifecycle_parallel.go` prepare/commit/rollback, worker create path where applicable | No parallel writer for `pending_create_*`, `instance_token`, `session_key`, or config hash keys after conversion begins | None without requirement update | Retire W-005 and W-012 together | Revert only as whole runtime-start slice; do not split prepare/commit/rollback ownership. |
| 4 Close/retire | worker close, API close/retire, session close sweep | Work-release scan may stay outside session but must consume committed close facts | None; provider-stop failure still leaves bead open | Retire W-008 close-mutation exception | Bake requires duplicate close, missed-event, and work-release recovery tests. |
| 5 Wake/hold/drain | `cmd/gc/session_wake.go`, sleep/pin/wait-hold paths | Controller can still compute demand; direct drain metadata writers shrink per command | None; partial work query remains fail-closed | Retire W-006/W-007 exceptions | Revert command cluster as a unit if drain cancellation changes. |
| 6 Reconciler facts | fact readers for lifecycle/config/runtime/work snapshots | `ComputePoolDesiredStates` scheduling remains controller-owned | None; missing/partial facts preserve current fail-open/fail-closed behavior per row | Guard forbids new raw lifecycle decisions in adapters | Bake requires query-count and partial-fact tests. |
| 7 Health/progress/circuit | provider health/progress facts into narrow decisions if proven | Alert dedupe, budgets, restart counts, and circuit state remain reconciler-owned unless separately approved | None; absent/stale/unknown health fails open | Guard only after missing proof files are restored/replaced | Revert by returning to reconciler-owned decision helpers. |

Shared call-site plan:

| Call site | Current role | End state |
|---|---|---|
| `cmd/gc/session_beads.go` | CLI/reconciler close and work-release bridge | Calls worker/session close command; work release remains subscriber scan. |
| `cmd/gc/session_lifecycle_parallel.go` | reconciler runtime-start shell | Calls prepare/commit runtime-start command; keeps provider execution. |
| `internal/api/session_resolution.go` | API target resolution and direct named materialization | Uses classifier and delegates materialization to worker/session command. |
| `internal/api/session_manager.go` | API manager construction exception | Retired or narrowed to a command factory after API callers delegate. |
| `internal/worker/handle.go` | canonical worker boundary | Exposes the command surface needed by production CLI/API without leaking raw patch maps. |

## Operability Contract

<!-- REVIEW: added per operability-contract -->

Every decision, command, and conflict result must carry operator-facing evidence
that maps to existing trace, API, CLI, log, and doctor surfaces.

Required diagnostic fields:

- operation, site, reason, outcome
- session ID, session name, alias, template/config target, store ref
- lifecycle projection, blocker, wake cause, target kind
- generation, instance token, runtime session key when relevant
- fact freshness: snapshot age, partial-query flags, runtime probe source,
  config revision, work-query completeness
- conflict details: candidate list, precondition that failed, retryable flag
- event emission result and subscriber/recovery path used

Outcome vocabulary:

| Outcome | Meaning | Surfaces |
|---|---|---|
| accepted | command committed or decision scheduled a safe action | trace, logs where applicable |
| rejected | invalid target, forbidden operation, illegal transition | API/CLI typed error, trace |
| blocked | current facts valid but policy blocks mutation | API/CLI typed conflict, trace, doctor if persistent |
| no-op | already in target state | API/CLI success or idempotent no-op as current behavior requires |
| drained/closed | terminal lifecycle completion | event/trace plus work-release evidence |
| ambiguous | multiple equal-precedence candidates | API/CLI conflict with candidates |
| missed-event-recovered | scan repaired work after missed or skipped event | trace and doctor evidence |

Per-operation diagnostic mapping:

| Operation | Accepted site/reason | Rejected/conflict reasons | Required rendering tests |
|---|---|---|---|
| Target classification | `session.target.resolve` / selected candidate kind | not-found, ambiguous, forbidden-kind, configured-name-conflict, rejected-by-config, closed-not-allowed | API/Huma error body, CLI stderr/JSON, mail/extmsg logs, `gc trace` candidate rendering |
| Explicit wake | `session.wake.request` / wake cause | terminal, missing-config, identity-conflict, pending-create-in-flight, stale-state | API status/body, CLI exit/text, trace, no-template wake |
| Runtime start prepare/commit/rollback | `session.runtime_start.prepare`, `.commit`, `.rollback` | stale-generation, stale-instance-token, runtime-mismatch, partial-start, provider-start-failed | trace with token/generation, provider error logs, doctor partial-start output |
| Close/retire | `session.close.commit`, `session.identity.retire` | provider-stop-failed, already-closed, stale-state, duplicate-canonical, work-release-deferred | CLI/API close output, trace, doctor close/work-release evidence |
| Drain complete/cancel | `session.drain.complete`, `session.drain.cancel` | assigned-work-present, partial-work-query, stale-drain-generation | trace, `SessionDrainAckedWithAssignedWork` payload, doctor assigned-work output |
| Repair path | `session.repair.apply` | repair-unsafe, partial-query, unknown-key, non-session-bead | doctor and trace output proving each direct write |

Trace mapping rule:

- Use existing centralized trace site/reason/outcome codes in
  `cmd/gc/session_reconciler_trace_types.go` when available.
- Add new codes only in the centralized trace vocabulary with tests for
  `gc trace` rendering.
- API-visible conflict fields must be represented by typed Huma outputs or
  errors; do not smuggle structured data through raw JSON maps.

## API, CLI, And Typed-Wire Gates

<!-- REVIEW: added per api-cli-typed-wire-obligations -->

Any slice touching `internal/api`, `cmd/gc`, events, OpenAPI, dashboard types,
or generated clients must satisfy the root API control-plane rules:

- HTTP and SSE paths use Huma-registered typed request/response/event types.
- No hand-written JSON, `map[string]any`, or `json.RawMessage` on wire types
  except documented existing exceptions.
- OpenAPI is generated, not hand-written.
- Generated dashboard/client types are regenerated when schema changes.
- `make dashboard-check` passes when API/schema/dashboard files are touched.

Per-slice compatibility proof must include:

- API status codes, response body/error body shape, request IDs, and async
  result event behavior
- CLI stdout/stderr text, JSON output, and exit codes
- generated schema/client impact or a statement that no wire shape changed
- proof that API and CLI remain projections over the session object model, not
  parallel reimplementations of raw metadata decisions

## Performance And Fan-Out Contract

<!-- REVIEW: added per fact-materialization-performance -->

Deciders do no I/O, but fact materialization and subscriber work can still
regress hot loops. Each slice must state expected complexity and preserve
existing bulk-read behavior.

Requirements:

- Reuse per-tick session snapshots, config snapshots, runtime observations, and
  work-demand reads.
- Do not add per-session subprocess fan-out or per-session config scans in the
  controller hot path.
- Cap synchronous subscriber work; long or retryable reactions run through
  existing controller/reconciler loops.
- Record fact-build and subscriber durations in trace when a slice touches the
  reconciler hot path.
- Large-city target: fact materialization is O(session beads + relevant work
  beads + configured agents) per tick unless the current implementation already
  has a lower bound that must be preserved.

Slice-level budgets:

| Path | Budget |
|---|---|
| Package target resolution | Exact ID lookup plus bounded metadata list calls for `session_name` and `alias`; no all-session scan in the hot path. |
| API configured named resolution | Bounded metadata-filtered list calls for configured identity; no duplicate `session_name` full scan. |
| API path alias fallback | At most one all-session scan after prior exact/config/live resolution fails; state filter and most-recent-created tiebreaker happen in memory. |
| Mail recipient resolution | No session materialization for send/query; configured mailbox lookup uses bounded identity lists. |
| Extmsg member notification | Materialization is only on member miss and must be bounded by membership count; errors are logged without aborting unrelated members. |
| Runtime-start repair scan | Reuses reconciler session snapshot; no provider subprocess fan-out per session beyond existing runtime observation. |
| Close/work-release scan | List assigned open/in-progress work by assignee and store ref; idempotent retry allowed, no whole-city work scan when indexed query is available. |
| Event subscribers | Critical subscriber work is idempotent and capped; long reactions defer to controller/reconciler ticks. |

Every slice touching a hot path must include either a query-count test, a
benchmark, or an explicit source proof that the budget is preserved.

## Technical Requirements

### TR-001: Preserve Product Semantics

Technical refactors must preserve the current scenario ledger unless the change
explicitly identifies a product ambiguity or bug.

Acceptance criteria:

- A touched product scenario keeps its expected behavior, evidence, and tests
  aligned.
- New technical APIs are covered by tests that prove parity with every
  scenario row the slice touches.
- Any product behavior change adds or updates scenario rows with evidence and a
  short rationale.
- Cited proof files or commands exist at implementation time, or the slice is
  blocked until proof is restored or replaced.
- Characterization tests for current API, CLI, mail, extmsg, worker, manager,
  and reconciler behavior land before the behavior is delegated.

### TR-002: Define Operation-Specific Facts

`internal/session` defines read-only fact types that deciders consume. Fact
types describe already-gathered state; they do not perform store, runtime,
config, or event I/O.

Minimum fact families by slice:

- target classification facts
- explicit wake facts
- runtime-start facts
- close/retire facts
- awake/hold/drain eligibility facts
- reconciler lifecycle eligibility facts
- provider-health/progress facts

Acceptance criteria:

- Fact structs are copyable inputs to pure tests.
- Fact readers can live in adapters while fact types live in `internal/session`.
- Fact construction does not mutate session metadata.
- Broad shared fact structs are introduced only after repeated exact use.

### TR-003: Add Pure Deciders One Operation At A Time

Deciders are pure functions that take a requested operation plus immutable facts
and return a typed decision.

Acceptance criteria:

- Decider tests are table-driven and do not require bead stores, runtimes, or
  event recorders.
- Each decider covers one cohesive operation or cluster before the next cluster
  is extracted.
- Decider output names session-domain facts, not work/mail/extmsg commands.

### TR-004: Move Durable Mutation Into Session-Owned Commands

Production callers outside `internal/session` call session-owned command APIs.
Those commands apply durable session mutations themselves; external callers do
not apply session patch maps.

Acceptance criteria:

- Command APIs validate current session state before mutation.
- Events are emitted or returned only after the session mutation commits.
- Existing external `SetMetadata*`, `Update`, `Close`, or `Create` call sites
  for session lifecycle/identity metadata shrink as commands are adopted.
- Static guards prevent new bypasses outside approved paths.

### TR-005: Keep Subscribers Domain-Specific And Recoverable

Session events are session-domain facts. Subscribers decide what those facts
mean for their own domain, and critical subscribers must converge from durable
facts when events are missed.

Acceptance criteria:

- Session does not choose work beads to reopen, mail recipients to rewrite, or
  extmsg content to send.
- Work, mail, extmsg, trace, and notification reactions are driven by typed
  session events or current session facts.
- Runtime execution remains adapter-owned; session may request runtime intents
  but does not embed provider-specific execution policy.
- Critical reactions have idempotent scan/retry proof.

### TR-006: Start With Target Classification

The first extraction slice is session target classification. It separates what
an input token is from what an operation is allowed to do with it.

Acceptance criteria:

- Classification preserves the compatibility resolver chains in this document.
- API, mail, bead assignee normalization, and extmsg callers stop re-deriving
  these categories as the classifier is adopted.
- Existing target-resolution and Phase 0 interface tests continue to pass.
- Adversarial collision tests cover negative and ambiguous cases.

### TR-007: Keep The Event Log Path Open

The initial implementation may use post-commit in-process `SessionEvent`
values only for acceleration or observability. Design event names and payloads
so they can later become durable event-log facts if session state moves toward
event sourcing.

Acceptance criteria:

- Event payloads describe facts that happened, not commands to attempt.
- Event payloads carry stable identifiers and enough context for subscribers to
  react idempotently.
- No implementation assumes the current post-commit event mechanism is the
  permanent persistence authority for critical reactions.

## Backlog

### 1. Target Classification

Create a session-owned target classifier that current API, mail, bead assignee,
and extmsg callers can adopt without changing behavior.

Before implementation:

- Complete the target collision inventory for API, CLI, mail, extmsg, nudge,
  attach, inspect, materialization, and assignee normalization.
- Add the classifier result types and operation policy input only; do not add a
  broad session facade.
- Preserve each compatibility resolver chain in the target-classification
  contract before replacing a call site.

Proof:

- `internal/session/resolve_test.go`
- `internal/session/named_config_test.go`
- `internal/api/session_model_phase0_interface_spec_test.go`
- `internal/api/session_resolution_path_alias_test.go`
- New classifier collision tests for every result kind and every surface chain

### 2. Wake Command Slice

Extract explicit wake into a small decider plus session-owned command while
preserving current wake metadata, wait cancellation, terminal conflict, and API
response behavior.

Before implementation:

- Finish the wake writer inventory for CLI, API, wait, reconciler, and
  no-template fallback paths.
- Define stale close-vs-wake and wake-vs-drain conflict tests.

Proof:

- `internal/session/waits_test.go`
- `internal/session/lifecycle_transition_test.go`
- `internal/api/session_model_phase0_lifecycle_spec_test.go`
- New command atomicity tests for terminal, stale, and partial-write cases

### 3. Runtime Start Prepare/Commit/Rollback

Move runtime-start prepare, provider-start commit, rollback, pending-create
cleanup, `instance_token`, `session_key`, and runtime hash metadata into one
shippable command slice. Do not split these keys across wake, pool, provider
health, or progress slices.

Before implementation:

- Verify writer inventory W-005 and W-012 from source.
- Define prepare, commit, rollback, stale-token, crash-after-prepare, and
  provider-start-success/commit-failure tests.
- Prove batch atomicity or add repair rows for partial `pending_create_*`,
  `instance_token`, `session_key`, and hash writes.

Proof:

- `internal/session/lifecycle_transition_test.go`
- `internal/session/manager_test.go`
- `cmd/gc/session_lifecycle_parallel_test.go`
- `cmd/gc/session_reconcile_test.go`
- New runtime-start recovery and race tests

### 4. Close And Identity Retirement

Move close/retire identity mutation behind a session-owned command and publish
session-domain facts for subscribers.

Before implementation:

- Define whether the close command performs synchronous work release or returns
  a post-commit fact consumed by an idempotent scan in the same controller tick.
- Preserve worker-boundary routing for production CLI callers.

Proof:

- `internal/session/manager_test.go`
- `internal/api/session_model_phase0_lifecycle_spec_test.go`
- `cmd/gc/session_beads_test.go`
- New idempotent recovery tests for missed event and duplicate subscriber scan

### 5. Wake/Hold/Drain Eligibility

Extract wake/hold/drain eligibility from reconciler helper logic into a pure
session decider.

Before implementation:

- Anchor the fact set on `ProjectLifecycle`, `ComputeAwakeSet`, and current
  drain cancellation behavior.
- Define partial/unknown/stale fact handling for every destructive branch.

Proof:

- `cmd/gc/session_reconcile_test.go`
- `internal/session/lifecycle_projection_test.go`
- `internal/session/waits_test.go`
- `cmd/gc/session_wake_test.go`

### 6. Reconciler Facts, Pool Scaling, And Cold-Start Demand

Extract narrow lifecycle eligibility facts while leaving pool desired-state,
cold-start demand, request tiers, nested caps, work aggregation, and capacity
policy in the controller/reconciler.

Before implementation:

- Restore or replace the missing `cmd/gc/scale_from_zero_test.go` proof cited
  by `REQUIREMENTS.md`.
- State complexity and snapshot reuse expectations for large cities.

Proof:

- `cmd/gc/build_desired_state_test.go`
- `cmd/gc/session_lifecycle_parallel_test.go`
- Restored or replacement proof for `SESSION-RECON-002` and
  `SESSION-RECON-003`

### 7. Provider Health, Progress, And Circuit State

Feed provider/progress facts into narrow decisions only after stale-proof files
are restored or replaced. Reconciler keeps scheduling, alert dedupe, restart
budgets, circuit state, and trace output unless requirements approve moving a
specific rule.

Before implementation:

- Restore or replace the missing `cmd/gc/provider_health_gate_test.go` and
  `cmd/gc/session_progress_test.go` proofs cited by `REQUIREMENTS.md`.
- Preserve provider-health unknown/stale fail-open behavior and progress
  exemption rules.

Proof:

- `cmd/gc/session_reconciler_test.go`
- `internal/session/manager_test.go`
- Restored or replacement provider-health/progress tests

## Non-Goals

- Do not rewrite the reconciler wholesale.
- Do not introduce one broad session facade before a narrow contract is proven.
- Do not move work, mail, extmsg, or provider-specific runtime policy into
  `internal/session`.
- Do not make event sourcing the first implementation step.
- Do not weaken the active worker-boundary migration or its CI guard.
