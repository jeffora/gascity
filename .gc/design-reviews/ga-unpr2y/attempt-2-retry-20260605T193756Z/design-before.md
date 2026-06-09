# Session Boundary Design

| Field | Value |
|---|---|
| Status | Draft backlog |
| Behavior source | `REQUIREMENTS.md` |
| Scope | Technical architecture and extraction plan for session ownership |
| Latest design-review disposition | Attempt 1 global verdict `block`; this revision is an `iterate` response |

This document tracks the technical design for moving session behavior toward a
clearer ownership boundary. It is intentionally separate from
`REQUIREMENTS.md`: product behavior stays in the requirements ledger, while this
file holds design direction, backlog, and extraction sequencing.

## Product Rule

Technical refactors must preserve the behavior in `REQUIREMENTS.md` unless code
inspection finds documented ambiguity, a real bug, or a product decision that
needs owner input. If behavior changes, update `REQUIREMENTS.md` with the new
scenario and proof.

Implementation is blocked for any slice whose cited proof is missing, stale, or
too broad to show scenario parity. Replace the proof first, then extract the
code.

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
| Wake selection | `AwakeInput`, `ComputeAwakeSet`, pool desired-state helpers | Awake facts/decision | Baseline for slices 4-6. |
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

## Mutation Boundary

<!-- REVIEW: added per mutation-boundary-enforceability -->

`internal/session` owns durable session lifecycle and identity mutation. In
production code outside `internal/session`, do not write session lifecycle or
identity metadata directly with bead-store calls.

Session-owned mutation includes these field families:

| Family | Metadata keys |
|---|---|
| Lifecycle state | `state`, `sleep_reason`, `state_reason`, `closing`, `closed_at`, `archived_at`, `close_reason` |
| Create/start | `pending_create_*`, `start_requested_at`, `creation_complete_at`, `session_key`, `generation`, `instance_token`, `continuation_epoch`, `continuation_reset_pending`, reset metadata |
| Wake/hold/drain | `wake_request`, `wake_requested_at`, `wake_mode`, `pin_awake`, `held_until`, `quarantined_until`, `wait_hold`, `sleep_intent`, drain acknowledgements and completion metadata |
| Identity | `session_name`, `session_name_explicit`, `alias`, `alias_history`, configured named-session markers, continuity eligibility, retired identity markers |
| Runtime identity | provider session keys, transport, command hash/live hash fields that determine whether a runtime belongs to the bead |

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

### Current Mutation Landscape

This inventory is the working baseline reviewers should update when new writers
are discovered. A slice is not ready until its row is current.

| Area | Current mutation path | Field family | Target path | Owner slice | Exception status |
|---|---|---|---|---|---|
| `internal/session/manager.go` create/start/close methods | `Create*`, `SetMetadata*`, `Close` | lifecycle, create/start, identity, runtime identity | Session commands implemented here or split below this package | 2, 3 | Canonical owner, not an exception. |
| `internal/session/lifecycle_transition.go` | `MetadataPatch` builders | lifecycle, create/start, wake/hold/drain, identity | Keep as internal patch builders behind commands | all lifecycle slices | Internal only. |
| `cmd/gc/cmd_session_wake.go` | `WakeSession` plus fallback direct `SetMetadataBatch` for no-template wake | wake, lifecycle | Explicit wake command/applier | 2 | No lasting exception. |
| `cmd/gc/session_lifecycle_parallel.go` | pre-wake, commit-start, rollback, runtime identity writes | create/start, lifecycle, runtime identity | Start/wake command and runtime-intent applier | 2, 5, 6 | Must delegate as each start slice converts. |
| `cmd/gc/session_wake.go` | drain cancellation/completion and wake-blocker updates | wake/hold/drain, lifecycle | Wake/hold/drain command set | 4 | Must convert as a cluster. |
| `cmd/gc/session_sleep.go`, `cmd/gc/cmd_session_pin.go`, wait integration | sleep, pin, wait-hold metadata writes | wake/hold/drain, lifecycle | Hold/pin/sleep commands | 4 | No new direct writers. |
| `cmd/gc/session_beads.go` and close/sweep callers | session close, close metadata, assigned-work release | lifecycle, identity, work release | Close/retire command plus work subscriber scan | 3 | Work release remains outside session; close mutation moves in. |
| `cmd/gc/session_reconciler.go`, `cmd/gc/city_runtime.go`, pool desired-state helpers | sync, drift, scale, provider-health, progress, sweep decisions | lifecycle, create/start, wake/hold/drain | Reconciler fact readers plus session commands | 5, 6 | Partial-read rules must be explicit before conversion. |
| `internal/api/session_resolution.go` | direct materializing create for resolved named sessions | identity, create/start | Target classifier plus worker/session command delegation | 1, 3 | Active root-documented exception until retired. |
| `internal/api/session_manager.go` | constructs `session.Manager` for API handlers | command access | Worker/session command boundary | 3 | Active root-documented exception until retired. |
| `internal/api/huma_handlers_sessions_command.go` and legacy handlers | resolve/materialize/message/submit workflow | target, identity, create/start | Typed policy adapter over classifier and worker handle | 1, 3 | Must preserve Huma typed-wire behavior. |
| Doctor, migration, repair paths | repair empty type, normalize broken metadata | operational repair | Repair-only helpers with trace/log evidence | per affected slice | Allowed only with explicit reason and guard allowlist. |

### Static Guard

Before a slice is considered implemented, add or tighten a failing-build guard
that rejects new production bypasses. The guard should be AST-based or a
strict source scanner with tests, and it must:

- forbid production files outside `internal/session` and an explicit allowlist
  from calling `SetMetadata`, `SetMetadataBatch`, `Update`, `Create`, or `Close`
  when the call writes a session-owned key or applies a `session.MetadataPatch`
- forbid production callers outside `internal/session` from constructing and
  applying lifecycle patch maps directly
- list every temporary exception with owner slice, reason, and retirement
  condition
- keep tests, fixtures, migrations, doctor/repair utilities, and bead-store
  conformance utilities as bounded exceptions

The existing `cmd/gc/worker_boundary_import_test.go` remains the worker-boundary
guard; this boundary guard is additive and should not weaken it.

## Target Classification Contract

<!-- REVIEW: added per target-classification-precedence -->

Slice 1 is the first extraction. It separates pure classification from
operation policy. The classifier gathers candidates and returns a typed result;
API, CLI, mail, extmsg, nudge, attach, materialization, and assignee adapters
then apply per-operation permissions.

Authoritative precedence:

| Order | Result kind | Meaning | Notes |
|---|---|---|---|
| 1 | `direct-session-id` | Identifier is an existing session bead ID | Open or repairable session beads win before names. Closed IDs are returned only if policy allows closed targets. |
| 2 | `live-session-name` | Open exact `session_name` match | Preserves `SESSION-ID-003`. |
| 3 | `live-alias` | Open exact current `alias` match | Lower than live `session_name`. |
| 4 | `configured-named-identity` | Config declares a named identity with no live bead | Not a live target; materialization requires operation policy. |
| 5 | `path-alias` | Identifier is an accepted path alias for a surface that opts in | Never outranks live session identity. |
| 6 | `closed-session-name` | Closed exact `session_name` match | Read-only lookup after live matches fail and policy allows closed. |
| 7 | `closed-alias` | Closed exact current `alias` match | Read-only lookup after closed session-name matches. |
| 8 | `historical-alias` | Historical alias match | Inspect-only unless a future requirement says otherwise. |
| 9 | `template-factory` | `template:<name>` factory/config target | Not a live session target. |
| 10 | `ordinary-config-target` | Bare configured agent/template target | Not a live session target. |
| 11 | `not-found` or `ambiguous` | No allowed target, or multiple equal-precedence candidates | Must include candidate details for diagnostics. |

Typed result fields:

- `kind`, `operation`
- `input`, `normalized_input`
- `session_id`, `session_name`, `alias`, `template`, `config_target`
- `status` and `closed` when a bead was found
- `materializable` and `materialization_reason`
- `candidates[]` with kind, ID/name, status, and conflict reason
- `negative_kind` for `not-found`, `ambiguous`, `forbidden-kind`, or
  `requires-materialization`

Operation policy matrix:

| Operation surface | Direct/live targets | Configured named identity | Closed lookup | Template factory | Ordinary config target | Path alias |
|---|---|---|---|---|---|---|
| API/CLI wake, close, suspend, pin, nudge live session | allowed | materialize only when the existing requirement permits it | reject except documented terminal-wake behavior | reject | reject | reject unless explicitly mapped to a live session |
| API/CLI inspect, logs, transcript | allowed | show reserved-unmaterialized projection | allow read-only | reject as live session | reject as live session | allow only surface-owned read policy |
| Mail recipient | existing live named mailbox allowed | configured mailbox allowed without materializing | reject | reject | reject | reject |
| Extmsg bind/deliver | allowed by binding policy | no implicit materialization | allow only for historical binding inspection | reject | reject | binding service policy only |
| Attach/observe | allowed | materialize only through explicit create/session command | read-only inspect only | reject | reject | surface-owned policy |
| Assignee normalization | bead ID/session_name/alias normalized to durable session owner | configured named identity may normalize to configured mailbox/identity as today | no live ownership | no | no | no |

Adversarial parity tests must cover same-token collisions across bead ID,
session_name, alias, configured named identity, path alias, closed lookup,
historical alias, bare config target, and `template:` target.

## Scenario Traceability Matrix

<!-- REVIEW: added per scenario-parity-proof -->

Every implementation bead spawned from this design must carry the rows it
touches. "Current proof" must be a runnable test, source path, issue, or commit
that exists in the checkout at implementation time.

| Slice | Scenario rows | Current proof | New or updated proof | Allowed behavior change | Freshness gate |
|---|---|---|---|---|---|
| 1 Target classification | `SESSION-ID-003`, `SESSION-ID-004`, `SESSION-ID-005`, `SESSION-ID-007`, `SESSION-ID-008`, `SESSION-ID-009`, `SESSION-ID-010` | `internal/session/resolve_test.go`, `internal/session/named_config_test.go`, `internal/api/session_model_phase0_interface_spec_test.go`, `internal/api/session_resolution_path_alias_test.go` | Classifier table tests plus API/mail/extmsg adapter parity tests | None without `REQUIREMENTS.md` update | Tests must assert exact errors, candidate conflict details, and no factory fallback. |
| 2 Explicit wake command | `SESSION-LIFE-007`, `SESSION-START-003`, `SESSION-START-007` | `internal/session/waits_test.go`, `internal/session/lifecycle_transition_test.go`, `internal/api/session_model_phase0_lifecycle_spec_test.go` | Command/applier conflict, stale-fact, and API response tests | None except documented bug fix | Proof must include terminal closed, archived, suspended, drained, no-template, and wait cancellation cases. |
| 3 Close and identity retirement | `SESSION-START-005`, `SESSION-START-006`, `SESSION-WORK-001`, `SESSION-WORK-002`, `SESSION-ID-006`, `SESSION-ID-007` | `internal/session/manager_test.go`, `internal/api/session_model_phase0_lifecycle_spec_test.go`, `cmd/gc/session_beads_test.go`, commits cited in `REQUIREMENTS.md` | Close command tests, idempotent work-release scan tests, API/CLI error parity | None without owner approval | Must prove provider-stop failure leaves bead open and successful close releases work. |
| 4 Wake/hold/drain eligibility | `SESSION-LIFE-003`, `SESSION-LIFE-004`, `SESSION-START-003`, `SESSION-RECON-004`, `SESSION-WORK-004` | `cmd/gc/session_reconcile_test.go`, `internal/session/lifecycle_projection_test.go`, `internal/session/waits_test.go`, `cmd/gc/session_wake_test.go` | Pure decider tests for unknown/partial/stale facts plus drain-cancel parity | None | Destructive drain/close decisions must fail closed on partial facts. |
| 5 Pool scaling and cold-start demand | `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-005`, `SESSION-START-008` | `cmd/gc/build_desired_state_test.go`, `cmd/gc/session_lifecycle_parallel_test.go`; `cmd/gc/scale_from_zero_test.go` is cited by requirements but absent in this checkout | Restore or replace scale-from-zero proof before extraction, then add fact-reader complexity tests | None | Missing proof blocks this slice. |
| 6 Provider health and progress gates | `SESSION-RECON-006`, `SESSION-RECON-007`, `SESSION-RUNTIME-001`, `SESSION-RUNTIME-002` | `cmd/gc/session_reconciler_test.go`, `internal/session/manager_test.go`; `cmd/gc/provider_health_gate_test.go` and `cmd/gc/session_progress_test.go` are cited by requirements but absent in this checkout | Restore or replace provider-health/progress proof, then add stale runtime fact tests | None | Missing proof blocks this slice. |

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
precondition immediately before commit. The initial compare-and-set substitute is
lock-plus-reread at the command boundary unless the store grows native revision
tokens. A command that cannot validate its preconditions must reject with a
typed conflict rather than write a best-effort partial transition.

| Command cluster | Precondition fields | Written keys | Ordering/atomicity assumption | Conflict reasons | Retryability and repair proof |
|---|---|---|---|---|---|
| Target classification | none; read-only snapshot plus policy | none | no mutation | ambiguous, not-found, forbidden-kind | Retry after config/store change. |
| Request explicit wake | bead status, projected lifecycle, held/quarantine, named identity, configured template presence | `wake_request`, blockers, optional dormant state normalization | one `SetMetadataBatch` for wake/blocker keys | terminal, missing-config, identity-conflict, stale-state | Retriable when stale; tests race close vs wake. |
| Prepare runtime start | state, generation, instance token, pending-create claim, current runtime liveness, config hash | `state=creating`, `generation`, `instance_token`, `pending_create_started_at`, reset keys | commit metadata before provider start; generation/token identify the attempt | stale-generation, already-running, provider-health-red, budget-exhausted, missing-config | Retriable if runtime absent and generation unchanged; repair rolls back stale creating. |
| Commit runtime start | generation, instance token, runtime identity, provider start result | active/start-complete keys, session key/hash fields, pending-create clear | one batch for state and pending-create clear | stale-generation, runtime-mismatch, partial-start | Recovery can commit or roll back after observing runtime identity. |
| Close session | bead open, lifecycle not already closed, provider stop outcome when required | close metadata, bead close, identity retirement keys | stop succeeds before bead close when stop is required; close metadata and bead close are treated as one command unit | provider-stop-failed, stale-state, already-closed | Retriable; stop failure leaves bead open. |
| Retire named identity | closed/terminal lifecycle, continuity eligibility, duplicate canonical facts | alias/session_name clears, retired markers | same command as terminal close or follow-up idempotent repair | duplicate-canonical, stale-identity, not-terminal | Retriable and idempotent. |
| Drain ack/complete/cancel | drain generation, assigned-work facts, runtime ack source | drain state/reason keys, cleared ack metadata | reread assigned work before destructive close or no-wake completion | assigned-work-present, stale-drain-generation, partial-work-query | Retriable; partial facts cancel destructive action. |

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

## Technical Requirements

### TR-001: Preserve Product Semantics

Technical refactors must preserve the current scenario ledger unless the change
explicitly identifies a product ambiguity or bug.

Acceptance criteria:

- A touched product scenario keeps its expected behavior, evidence, and tests
  aligned.
- New technical APIs are covered by tests that prove parity with at least one
  existing scenario row.
- Any product behavior change adds or updates scenario rows with evidence and a
  short rationale.

### TR-002: Define Operation-Specific Facts

`internal/session` defines read-only fact types that deciders consume. Fact
types describe already-gathered state; they do not perform store, runtime,
config, or event I/O.

Minimum fact families by slice:

- target classification facts
- explicit wake facts
- close/retire facts
- awake/hold/drain eligibility facts
- pool demand facts
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

- Classification follows the precedence table in this document.
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

Proof:

- `internal/session/resolve_test.go`
- `internal/session/named_config_test.go`
- `internal/api/session_model_phase0_interface_spec_test.go`
- `internal/api/session_resolution_path_alias_test.go`
- New classifier collision tests for every result kind in the precedence table

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

### 3. Close And Identity Retirement

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

### 4. Wake/Hold/Drain Eligibility

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

### 5. Pool Scaling And Cold-Start Demand

Move pool desired-state and cold-start demand decisions behind session-owned
facts/deciders while leaving work queries and store aggregation in adapters.

Before implementation:

- Restore or replace the missing `cmd/gc/scale_from_zero_test.go` proof cited
  by `REQUIREMENTS.md`.
- State complexity and snapshot reuse expectations for large cities.

Proof:

- `cmd/gc/build_desired_state_test.go`
- `cmd/gc/session_lifecycle_parallel_test.go`
- Restored or replacement proof for `SESSION-RECON-002` and
  `SESSION-RECON-003`

### 6. Provider Health And Progress Gates

Feed provider/progress facts into session-owned health decisions, while the
reconciler keeps scheduling, budgets, and trace output.

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
