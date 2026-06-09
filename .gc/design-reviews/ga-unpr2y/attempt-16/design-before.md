# Session Refactor Design

| Field | Value |
|---|---|
| Status | Draft backlog |
| Behavior source | `REQUIREMENTS.md` |
| Scope | Small refactors that move session decisions into `internal/session` |
| Latest design-review disposition | Attempt 15 global verdict `block`; this revision is an `iterate` response |
| First executable work | Slice 0 universal evidence preflight only; per-slice contracts land with the first slice that uses them |
| Archive | `DESIGN_REVIEW_NOTES.md` is historical and non-normative unless text is copied here |

This is not a new architecture program. It is a refactor plan for making
session behavior easier to find, test, and change without changing product
behavior. The active design-review verdict is still `block`, so no
behavior-moving slice is approved. Slice 0 may only collect source-complete
evidence, validate artifact freshness, and repair or owner-retire stale
requirement evidence. Every later slice must add its own operation preflight
before it moves behavior.

## Goal

Move session-specific decisions into `internal/session` one operation at a
time.

Today some callers in API, CLI, worker, and reconciler code know too much about
session state, target resolution, wake rules, and lifecycle metadata. The
refactor should make those callers thinner:

```text
caller gathers facts -> internal/session decides -> caller executes or renders
```

The caller may still gather external facts, call runtime providers, write
non-session domain state, or render API/CLI output. The session module owns only
the session rule.

## Product Rule

`REQUIREMENTS.md` is the behavior source of truth. A refactor must preserve the
scenario rows it touches.

If current code and `REQUIREMENTS.md` disagree:

- update code when the requirement is right
- update `REQUIREMENTS.md` when current behavior is the intended product rule
- ask for a decision when neither is clear

Behavior-changing requirements edits require a durable owner-approval artifact
in the implementing bead or plan. An implementation may not relabel behavior
drift as a requirements update unless the changed scenario rows, exact proof
selectors, and approval record land together.

## Attempt 15 Review Response

<!-- REVIEW: added per attempt-15-global-synthesis -->

Attempt 15 still blocks decomposition. This revision keeps
`design_review.verdict=iterate` and narrows the approved next work to a
self-validating, non-mutating Slice 0. Future command, event, runtime, and
diagnostic contracts are not authorized by Slice 0 just because they are named
there; each becomes normative only in the slice that first delegates a caller to
it.

| Finding | Active design response |
|---|---|
| Slice 0 gate can pass with stale or missing proof. | Split Slice 0 into universal evidence and per-slice preflights. Add a meta-validator that fails on missing, skipped, or zero-match validators and require owner-approved repair or retirement for stale `SESSION-*` evidence before dependent slices cite it. |
| API query target classification does not preserve current resolver semantics. | Replace the broad first-adopter row with the exact `resolveSessionTargetIDWithContext` precedence, including configured named-session handling before live aliases, config-orphan rejection, path-alias by `Title`, closed lookup gates, and `RepairEmptyType` quarantine. |
| Mutation ownership still has escape hatches. | Require runtime bridge fencing or session-command routing for session-owned metadata families, create-time metadata, direct status/type writes, close/reopen paths, repair helpers, and helper-returned patch maps. |
| Atomic command semantics are not enforceable. | Require immutable operation fact snapshots with mandatory `now`, config facts, token/revision/value preconditions, store capability rows, post-write verifiers, and failure-injection tests per command. |
| Event recovery and diagnostics are underspecified. | Add current `session.*` event inventory requirements, stable identity payload rules, durable scan ownership, idempotency keys, diagnostic-to-trace mappings, renderer proof, and negative event/scan tests. |
| Reconciler/runtime/session fact boundaries are implicit. | Make `BOUNDARY_MATRIX.yaml` row schema explicit, including source owner, policy owner, freshness, unknown/stale/partial/provider-error behavior, destructive-action safety, proof selectors, and forbidden session imports/fields. |
| Hot-path cost gates miss target resolution scans. | Add per-hot-path budget rows for target resolution and reconciler fact compilation; `resolveLiveSessionByPathAlias` must be indexed, removed, or explicitly budgeted before delegation. |
| Worker/API/CLI compatibility needs hard choices. | Production API/CLI mutating lifecycle operations must route through `worker.Handle` unless an exact root-approved, expiring exception row exists with parity, OpenAPI/generated-client/dashboard proof, and retirement criteria. |
| Migration coexistence and rollback are under-specified. | Add cross-migration ordering for shared API/CLI/reconciler files, per-field ownership transfer rules, raw-writer retirement conditions, rollback data-direction rules, and cross-process fences when writers coexist. |
| `RepairEmptyType` is still an under-owned side effect. | Read-only classifier paths return `repair-needed` instead of silently mutating. Any actual repair goes through the audited repair owner, propagates persistence failures, and records before/after trace evidence. |
| Historical notes and human event messages could become implicit criteria. | Historical notes remain rationale only unless copied here. Event message text is operator-only; subscriber data must be typed payload or envelope fields. |

Unfixable in this document: attempt-15 persona child beads were stamped with
some output paths under `attempt-1`, though the global synthesis normalized
copies into `attempt-15/persona-syntheses`. Artifact placement is workflow
plumbing outside `internal/session`; this design consumes the global synthesis
path for attempt 15.

## Attempt 14 Review Response (Historical)

<!-- REVIEW: added per attempt-14-global-synthesis -->

Attempt 14 blocked decomposition. This block is retained as provenance for the
rules that attempt 15 supersedes and tightens. The active review response is
the attempt 15 section above.

| Finding | Active design response |
|---|---|
| Target classification contract is undefined. | Add the Target Classification Contract below with typed candidate/result kinds, no-side-effect rules, closed/historical handling, a first adopter, and a per-surface precedence matrix. |
| Behavior parity proof is unenforceable. | Slice 0 creates `SCENARIO_PARITY.yaml` mapping every active `SESSION-*` row to touched surfaces, exact selectors, proof commands, current oracle, and owner-approved amendment state. |
| Mutation ownership lacks a writer boundary. | Slice 0 creates a source-complete Mutation Ownership Ledger plus a shrink-only CI guard for external `SetMetadata*`, patch-map application, patch constructors, repair, doctor, migration, API, CLI, worker, and test exceptions. |
| Mutating commands lack atomic recovery semantics. | Every mutation-moving slice must define an Atomic Command Contract before decomposition. The contract names facts, stale/rejected outcomes, write primitive or fence, event/runtime ordering, partial states, recovery owner, and crash/race tests. |
| Worker/API/CLI routing is unsettled. | The worker-boundary section below names read-only store-level work, worker-routed lifecycle work, existing API exceptions, and the route inventory required before any new delegation. |
| Reconciler and runtime fact ownership is underspecified. | The reconciler/session split matrix below keeps work demand, scheduling, budgets, provider health, progress, alerts, pool sizing, and provider policy outside `internal/session`. |
| Event and recovery rules are not normative. | The Event and Recovery Contract below makes events post-commit facts, not commands, and requires durable scans for safety-critical convergence. |
| Operator diagnostics and cost budgets are missing. | Slice 0 creates `DIAGNOSTICS_MANIFEST.yaml` with structured reason/outcome codes, renderer coverage, query/subprocess budgets, and large-city hot-path proof. |
| Scope authority and vocabulary need tightening. | `DESIGN_REVIEW_NOTES.md` is historical only. Future vocabulary remains private or provisional until a real adopting caller proves exact fields and non-goals. |

Unfixable in this document: the design-review workflow wrote some attempt-14
persona syntheses under an older attempt path. That is workflow plumbing outside
`internal/session`; the workflow formula must fix artifact placement. This
document consumes the bead-stamped global synthesis path for attempt 14.

## Authority And Entry Gates

<!-- REVIEW: added per attempt-14-scope-authority -->

No section in `DESIGN_REVIEW_NOTES.md` authorizes work. If an old note is still
true, copy the rule into this file or into a Slice 0 artifact.

No behavior-moving bead may start until it depends on a closed Slice 0 bead,
cites the exact universal artifact rows it consumes, and includes the
operation-specific preflight rows for the caller it is about to move. A
preflight row cannot be satisfied by a broad future-facing contract; it must
name current code paths, current behavior proof, and the exact write or
rendering surface in scope. Required bead metadata for later slices:

- `session_design.slice`
- closed Slice 0 bead ID
- referenced inventory IDs
- referenced `SESSION-*` rows
- one-writer evidence for touched key families
- rollback and data-direction evidence
- proof command and test selector metadata
- route, wire, or rendering parity rows when public surfaces change
- unresolved-risk metadata when behavior is intentionally deferred
- per-slice preflight artifact IDs for operation contracts introduced by the
  slice

## Slice 0 Evidence Preflight

<!-- REVIEW: added per attempt-15-slice0-gate -->

Slice 0 is non-mutating. It does not move a caller, introduce a public command
API, change target resolution, materialize sessions, repair metadata, add event
payloads, or alter reconciler policy.

Slice 0 has two artifact classes:

| Class | Normative scope |
|---|---|
| Universal evidence | Source-complete inventories, schemas, validators, negative fixtures, current behavior proof, and stale-requirement reconciliation that every later slice can cite. |
| Per-slice preflight | Operation-specific command, event, diagnostics, route, budget, and migration rows added by the first behavior-moving slice that needs them. Slice 0 may define the schema for these rows but must not pre-approve future behavior. |

Universal Slice 0 must create and validate these artifacts:

| Artifact | Purpose |
|---|---|
| `SLICE0_BASELINE.md` | Checkout baseline, upstream/ancestry facts, proof transcript, and explicit reason if an upstream ref is unavailable. |
| `SLICE0_CONTRACT.yaml` | Machine-readable close contract for artifacts, schemas, validators, fixtures, proof commands, workflow finalizer metadata, and a meta-check that fails when any named validator is absent or skipped. |
| `BOUNDARY_INVENTORY.md` | Human-readable source inventory for session lifecycle, identity, route, event, repair, and runtime fact boundaries. |
| `SESSION_BOUNDARY_SYMBOLS.yaml` | Machine-readable writer/read inventory with path, symbol, key family, dynamic-key source, receiver/helper chain, owner, exception, and retirement condition. |
| `SCENARIO_PARITY.yaml` | One row for every active `SESSION-*` scenario, touched surfaces, exact tests or static selectors, proof command, current oracle, amendment status, owner, and expiry for any accepted stale evidence. |
| `TARGET_CLASSIFICATION_CONTRACT.yaml` | Schema and first-adopter rows for typed result states, match vectors, bead/config state, diagnostics, resolver precedence, and no-side-effect proof. |
| `COMMAND_APPLIERS.yaml` | Schema for operation rows and universal inventory of current command-like writers. Operation-specific wake, close, drain, stop/interrupt, runtime-start, runtime-missing cleanup, identity-retirement, and repair/backfill rows are per-slice preflight unless the slice moves that operation. |
| `BOUNDARY_MATRIX.yaml` | Session/reconciler/runtime ownership matrix schema plus universal source inventory for wake, hold, wait, drain, runtime observation, pool demand, provider health, progress, repair, partial snapshots, and destructive actions. |
| `API_CLI_ROUTE_INVENTORY.yaml` | Huma routes, legacy mux routes, Cobra commands, `apiClient()` fallbacks, generated-client paths, dashboard/SSE projections, doctor, inspect, trace, mail, extmsg, transcript, attach, nudge, pool-resume, assignee, and generic bead routes that can reach sessions. |
| `WORKER_BOUNDARY_EXCEPTIONS.yaml` | Exact production exceptions to the worker boundary, including API manager construction and API session-resolution direct-create paths, with owner and expiry. |
| `DIAGNOSTICS_MANIFEST.yaml` | Operation/check IDs, reason/outcome codes, renderer surfaces, trace mappings, required facts, redaction, event relationships, query budgets, subprocess budgets, and hot-loop constraints. |
| `VOCABULARY_CHECKPOINTS.yaml` | Shared type/field lifecycle, first caller, exact demanded fields, non-goals, rule-of-two status, and expansion rule. |

Every machine-readable artifact must declare `schema_version`, row `id`,
`owner`, `source_paths`, `proof_selectors`, positive fixtures, negative
fixtures, and retirement or expiry metadata where applicable. Human-readable
artifacts may summarize, but CI gates may only consume the machine-readable
rows.

Slice 0 must explicitly reconcile active requirement rows whose evidence is
stale, missing, or on another ref. At minimum it must repair or owner-retire
the evidence for `SESSION-RECON-002`, `SESSION-RECON-003`,
`SESSION-RECON-006`, and `SESSION-RECON-007` before a later slice cites those
rows. A row cannot be "green" because a file exists; the proof selector must
execute or the owner-approved retirement/amendment record must name the product
decision.

Slice 0 validators must fail on missing artifacts, schema-invalid YAML, stale
paths, zero-match selectors, missing negative fixtures, broad exclusions,
unowned writers, missing renderer proof, unobservable event claims, absent
budget proof, skipped validators, or requirements rows without parity records.
Every validator referenced by `SLICE0_CONTRACT.yaml` must be invoked by the
minimum proof command and must assert that at least one positive fixture and one
negative fixture exercised the relevant selector.

Minimum proof command:

```bash
go test ./cmd/gc ./internal/session ./internal/api ./internal/events -run 'TestSlice0Contract|TestSessionBoundaryGuard|TestSessionBoundaryInventoryFresh|TestSlice0Artifacts|TestScenarioParityFreshness|TestVocabularyCheckpoints|TestSessionDiagnosticsManifest|TestSessionCommandApplierLedger|TestSessionBoundaryMatrix|TestSessionRouteInventoryFresh|TestWorkerBoundaryExceptionLedger|TestEveryKnownEventTypeHasRegisteredPayload' -count=1
```

## Target Classification Contract

<!-- REVIEW: added per attempt-15-target-classification -->

The first behavior extraction after Slice 0 is read-only session target
classification. It separates what an input token is from what an operation is
allowed to do with it.

First adopting surface: API query-side session lookup for read-only session get,
pending, transcript, stream attachment, and related query endpoints that already
call `resolveSessionTargetIDWithContext` or its allow-closed wrapper without
mutating session beads. Mutating API commands, CLI commands, mail, extmsg,
assignee normalization, nudge, attach, pool resume, and sling remain
characterization-only until their own surface rows pass.

The classifier is side-effect free:

- no store writes
- no session materialization
- no repair or reopen
- no runtime provider calls
- no event emission
- no work, mail, extmsg, or convoy mutation
- no API/CLI rendering decisions

First-adopter resolver precedence must preserve the current
`internal/api/session_resolution.go` behavior exactly:

1. Reject non-empty `template:<name>` targets as not found.
2. Resolve direct session bead ID with `ResolveSessionIDByExactID`.
3. Resolve configured named-session targets before live alias/session-name
   matches: canonical named bead wins, configured-name conflicts are conflict
   results, and non-materializing query paths return not found for reserved
   identities without a canonical bead.
4. Resolve ordinary live sessions through `session.ResolveSessionID`: open exact
   `session_name`, then open exact current `alias`, with dual alias/session-name
   demotion preserved.
5. If a live named-session bead resolves but its configured identity is absent
   from current config, reject it as the existing config-orphan not-found case.
6. Resolve path aliases by `Title` only after session-name/alias resolution,
   only for non-named session beads in state `active`, legacy `awake`, or empty
   state, with newest `CreatedAt` winning duplicate titles.
7. For allow-closed query endpoints, reject configured named-session targets as
   not found before closed lookup, then preserve `ResolveSessionIDAllowClosed`:
   direct ID, open session-name, open alias, closed session-name, then closed
   alias. Historical alias metadata is not a lookup source.
8. Preserve current error projection: `writeResolveError` and
   `humaResolveError` render ambiguity and configured-name conflicts as 409,
   not-found and config-orphan rejection as 404, and other errors as 500 with
   the existing code prefixes and response shapes.

`RepairEmptyType` is not allowed in the read-only classifier path. When current
helpers would repair an empty-type session bead as a side effect, the new
classifier returns `repair-needed` with the same match vector and an audited
repair command owns the write. Parity tests must prove whether the first
adopter preserves successful selection after a separate repair or intentionally
changes the result with owner approval.

Typed result schema:

| Field | Meaning |
|---|---|
| `result_kind` | `selected`, `not-found`, `ambiguous`, `rejected`, `repair-needed`, or `store-error`. |
| `source_surface` | Surface asking for classification, starting with `api-query`. |
| `normalized_token` | Trimmed target token after surface-local normalization. |
| `match_vectors[]` | Ordered matches with `vector_kind` (`template-prefix`, `direct-id`, `configured-name`, `live-session-name`, `live-alias`, `path-alias`, `closed-session-name`, `closed-alias`), candidate ID, order, and conflict group. |
| `bead_state` | For concrete matches: bead ID, status, type, labels, lifecycle state, session name, alias, title, created time, and whether the bead is repairable empty-type. |
| `config_state` | For configured targets: named identity, mode, canonical bead ID, conflict bead ID, reserved-unmaterialized flag, config-orphan flag, and materialization allowed flag. |
| `diagnostics` | Stable reason code, retryability, source facts, stale or partial fact marker, and renderer surfaces that must preserve current wire output. |
| `terminal_error` | Wrapped domain error class used only by adapters that must preserve `errors.Is` behavior and current Huma/legacy HTTP rendering. |

Typed selection result kinds:

- `selected`: exactly one candidate selected by a surface adapter
- `not-found`: no current surface-legal candidate
- `ambiguous`: multiple candidates remain after surface precedence
- `rejected`: candidate exists but the surface forbids it
- `repair-needed`: read path found metadata that a later audited repair command
  may fix after revalidation
- `store-error`: classification could not complete

The raw classifier never decides operation policy. Each surface adapter owns the
authoritative selected result and must preserve current output behavior.

First-adopter proof must include positive and negative fixtures for every
precedence row above, plus exact wire parity for legacy mux and Huma query
handlers, generated client compatibility where schema is visible, and no-delta
tests for `writeResolveError` and `humaResolveError`.

Per-surface delegation matrix:

| Surface | Delegation state | Current precedence to preserve before delegation |
|---|---|---|
| API query-side read-only session lookup | First adopter after Slice 0 | The eight-step `resolveSessionTargetIDWithContext` precedence above, including configured named-session handling, config-orphan rejection, path-alias title matching, allow-closed behavior, `RepairEmptyType` quarantine, and existing Huma/legacy error rendering. |
| API mutating session commands | Characterization only | Preserve route-specific Huma status, problem body, request ID behavior, generated-client shape, materialization, close/wake semantics, and existing API exceptions until a command contract and worker-boundary row exist. |
| CLI local and `apiClient()` fallback paths | Characterization only | Preserve stdout, stderr, JSON shape, exit code, fallback order, and whether the local path or API path owns the session operation. |
| Mail send/query | Characterization only | Preserve configured named mailbox behavior without materializing a session, live named-session mailbox use, recipient-set behavior, and rejection of template factory targets where required. |
| Extmsg | Characterization only | Preserve binding cleanup and peer-notification behavior outside the classifier; the classifier may identify session targets but not touch extmsg state. |
| Assignee, sling, nudge, attach, logs, transcript, pool resume | Characterization only | Each surface needs its own closed/historical handling, ambiguity behavior, error projection, materialization rule, and no raw-candidate leakage proof before delegation. |

## Mutation Ownership Ledger

<!-- REVIEW: added per attempt-15-mutation-ownership -->

Before any mutation-owning slice, Slice 0 must enumerate every production writer
that can alter session lifecycle, identity, wake, runtime-start, close, drain,
repair, or transcript metadata.

Each ledger row must include:

- exact path and function
- called store or helper method
- session bead targeting proof
- exact key family and literal key, or dynamic-key source
- top-level bead field when the write changes `Type`, `Status`, assignee,
  labels, close/reopen state, or create-time metadata
- current owner
- intended owning slice
- allowed exception reason
- exception expiry and owner-approved retirement condition
- persistence-error handling
- trace, doctor, or proof evidence when operator-visible
- whether the write can run through a generic API/CLI bead mutation bridge

The shrink-only guard must fail on new external `SetMetadata*`,
`MetadataPatch` application, patch-constructor use, raw top-level status/type
mutation, `session.Manager.Create*` bypasses, repair/doctor/migration bypasses,
or generic bead mutation bridges unless an exact expiring row exists. Tests and
fixtures may be exceptions only when their path is exact and their fixture role
is declared.

Session-owned families include lifecycle (`state`, wake, hold, drain, pending
create/start, quarantine, archive, close reason), identity (`session_name`,
`alias`, named-session identity/mode/canonical/historical markers, alias
history, continuity eligibility), runtime-start/resume (`session_key`,
transport/provider runtime identity, startup/confirmed timestamps), and
session repair markers. Slice 0 must make this key-family list
machine-readable and share it between static guards and runtime bridge denial.

The guard must cover every mutation form, not only direct metadata calls:

- `beads.UpdateOpts.Metadata`, `MetadataPatch`, and dynamic-key map
  construction
- create-time metadata passed to session or bead constructors
- `Store.Close`, `Store.Reopen`, `Store.Update` status/type writes, and
  `CloseAll` metadata
- helpers that return patch maps, including lifecycle and retirement helpers
- repair helpers, doctor/migration paths, and runtime identity backfills
- API/CLI generic bead update routes that can target session beads

If old and new writers coexist for a key family, the design for that slice must
name the cross-process fence: store-native compare-and-swap, value-embedded
token checked immediately before commit, or an explicit repair-converged blind
write rule with crash/race tests. An exception that keeps two unfenced writers
for the same session-owned field is not allowed.

Repair writes, including empty-type repair and runtime identity backfills, must
have an audited owner. A read path may return `repair-needed`, but it may not
silently repair session-owned keys unless it is the named repair owner for that
key family and propagates persistence errors.

## Atomic Command Contract

<!-- REVIEW: added per attempt-15-atomic-command-contract -->

Every mutation-moving slice must define an operation-specific command contract
before decomposition. A generic `facts -> command` shape is not sufficient.

Required fields per operation:

- immutable facts read by the decider, including mandatory `now`, current
  config hash or relevant config facts, bead revision/token when available, and
  runtime observation timestamp when runtime facts are used
- freshness and stale-fact rules
- preconditions and terminal conflicts
- rejected result and retryability
- exact pre-commit revalidation: token/revision/value preconditions checked
  immediately before commit against the same fact snapshot the decider consumed
- store capability row for the current persistence surface: atomic update,
  partial metadata update, conditional update, close/reopen semantics, and blind
  write behavior
- write primitive, conditional fence, token, or explicit repair-converged blind
  write rule
- post-write verifier and expected durable state
- runtime provider ordering
- post-commit event ordering
- rollback or compensation rule
- partial-state matrix
- crash recovery owner
- duplicate and skipped-event behavior
- characterization tests for the current caller
- parity tests after delegation

Operations that require contracts before they can move: wake/start, close,
drain, stop/interrupt, configured identity retirement, runtime start
prepare/commit/rollback, runtime-missing cleanup, repair/backfill,
provider-health reactions, and reconciler lifecycle transitions.

Each operation row must include failure-injection tests for stale snapshots,
raced lifecycle operations, partial metadata writes, duplicate commands,
provider success followed by commit failure, commit success followed by event
failure, skipped events, and crash recovery scan convergence.

## Worker/API/CLI Boundary

<!-- REVIEW: added per attempt-15-worker-api-cli-boundary -->

Read-only classification may be store-level because it does not create,
destroy, wake, drain, or repair sessions. Lifecycle operations exposed through
production API or CLI code must route through `internal/worker/handle.go`
unless root `AGENTS.md` and `WORKER_BOUNDARY_EXCEPTIONS.yaml` list an exact,
expiring, owner-approved exception.

Current routing rules:

- Close remains aligned with `worker.Handle.CloseDetailed` for production API
  and CLI callers, including Huma routes. If any close path cannot use
  `worker.Handle`, the same change must add an exception row with event,
  cleanup, response-shape, OpenAPI/generated-client, dashboard, and retirement
  proof.
- Wake and drain default to `worker.Handle` when exposed through production API
  or CLI mutation paths. A store-level route requires a root-approved exception
  row before delegation, with parity tests and an expiry.
- Existing API manager construction and API session-resolution direct-create
  paths are documented exceptions, not precedent for new bypasses.
- New direct `session.Manager.Create*`, sessionlog, or worker-boundary bypasses
  outside tests are forbidden unless the exception ledger names exact path,
  owner, reason, and retirement proof.
- API-visible behavior must preserve Huma typed responses, OpenAPI generation,
  generated client compatibility, dashboard/SSE expectations, and fallback
  parity with CLI `apiClient()` callers.

Route-level parity tests must cover both local and API paths for every
session-affecting command that delegates to a new session-owned operation.

## Migration Coexistence And Rollback

<!-- REVIEW: added per attempt-15-migration-coexistence -->

Worker-boundary and session-mutation-boundary migrations overlap in
`internal/api/session_resolution.go`, `cmd/gc/session_reconciler.go`, and
`cmd/gc/session_beads.go`. Before a slice touches any of those files, its
preflight must add a migration row with:

- ordered predecessor and successor slices for the shared file
- field-family ownership before, during, and after the slice
- raw-writer retirement condition for each session-owned field family
- cross-process fence when old and new writers coexist
- rollback data direction: whether rollback preserves new fields, clears them,
  or runs a repair/backfill command
- tests proving old readers tolerate new fields and new readers tolerate old
  fields during rollback

No slice may leave both a legacy raw writer and a new session-owned command
writing the same field family unless the row names the fence and proves it with
raced writer tests.

## Reconciler, Runtime, And Session Split

<!-- REVIEW: added per attempt-15-reconciler-runtime-fact-ownership -->

`internal/session` may own lifecycle eligibility and identity classification
over immutable facts. It must not own controller policy.

| Input or decision | Owner |
|---|---|
| Lifecycle projection, terminal states, wake blockers, identity conflicts, configured-name canonical/historical status | `internal/session` |
| Work demand, dependency state, dispatch scheduling, pool desired size, cold-start demand, nested caps, restart budgets, alert dedupe, progress policy, idle-sleep policy | controller/reconciler |
| Runtime liveness observations, provider errors, process existence, transcript/provider-specific facts | runtime provider or worker/runtime adapter |
| API/CLI rendering, Huma error mapping, stdout/stderr/JSON shape | API/CLI adapter |
| Mail, extmsg, work release, convoy, and external-message policy | owning domain package or controller |

Provider-neutral runtime intent fields may cross into a session-owned command
only when the command contract needs them. They may include stable session bead
ID, generation or instance token, provider family, runtime session key, work
directory, and config hash. They must not include provider-specific scheduling,
health, progress, budget, or alert decisions.

Destructive actions with unknown, stale, partial, or provider-error runtime
facts are rejected unless the current requirements and the operation contract
explicitly state a safe convergence rule and test it.

`BOUNDARY_MATRIX.yaml` rows must include:

- source owner and policy owner
- current source files and test selectors on this branch
- allowed session inputs and outputs
- forbidden `internal/session` imports, fields, or policy decisions
- freshness rule for facts and the stale/unknown/partial/provider-error result
- destructive-action safety rule: fail closed, fail open, or advisory only
- wake-cause production owner
- recovery or repair owner
- proof selector and negative fixture

Required rows before behavior-moving slices include provider health, progress
and idle thresholds, runtime observations, wake-cause production, partial
snapshots, drain and drain-ack, runtime-missing cleanup, adapter provider
actions, pool demand, and work-demand compilation. The row must distinguish
what exists on the current branch from behavior that exists only on another ref
or in historical notes.

## Event And Recovery Contract

<!-- REVIEW: added per attempt-15-event-recovery-contract -->

Session events are post-commit facts. They are not commands, locks, durable
truth, or the only recovery mechanism.

Event-bearing slices must state:

- committed fact that caused the event
- event emission owner
- stable canonical session identity fields: bead ID, session name, alias,
  configured identity, runtime session key, and generation/instance token when
  relevant
- typed payload registration in `events.KnownEventTypes`
- SSE/OpenAPI/generated-client obligations when the event is public
- best-effort subscribers versus critical recovery actions
- durable scan owner for critical convergence
- idempotency key or duplicate behavior
- crash-after-commit, skipped-event, duplicate-event, and stale-event tests

Close, work release, wake, drain, runtime start, provider-health reactions, and
identity retirement must converge from durable facts even when event emission or
subscriber execution is skipped.

Slice 0 must inventory current session events at least for
`session.woke`, `session.stopped`, `session.crashed`, `session.draining`,
`session.undrained`, `session.quarantined`, `session.idle_killed`,
`session.max_age_killed`, `session.suspended`, `session.updated`,
`session.drain_acked_with_assigned_work`, `session.stranded`, and
`session.work_query_failed`. Each row must say whether envelope fields are
sufficient, whether a typed payload is required, whether the event is critical
or best-effort, and whether public SSE/OpenAPI clients observe it.

Work-release convergence rows must include the release-identity snapshot,
scanner trigger, completion marker, duplicate/stale event behavior, partial
query behavior, idempotency key, and tests for crash after close, missed event,
duplicate event, and store query failure. Event `Message` text is
operator-only; subscribers may consume only typed payload fields and envelope
fields.

## Observability And Cost

<!-- REVIEW: added per attempt-15-diagnostics-and-cost -->

Classifiers, deciders, and commands must return structured diagnostics rather
than hiding decisions in local strings.

Diagnostic result fields:

- operation ID
- result kind
- reason code
- retryability
- selected session identity when safe to expose
- source facts used
- missing or stale fact indicator
- renderer surfaces that must show the result
- redaction keys

`DIAGNOSTICS_MANIFEST.yaml` owns trace mappings, doctor/session-inspect output,
API/CLI rendering parity, event relationships, and test selectors. New
diagnostic vocabulary must have centralized constants or manifest rows before
use.

Every diagnostics row must map to one `gc trace` site/reason/outcome rendering
or state that it is intentionally not trace-rendered. Rows must name source
facts, redaction, renderer tests, and negative tests proving machine-readable
data is not hidden only in message strings.

Cost rules:

- Hot-path classification and reconciler fact materialization must use bounded
  indexed lookups or counting-store proof.
- No all-session scan is allowed on a hot path unless the surface has a
  measured budget and a large-city baseline.
- No subprocess loop is allowed in classification or reconciler hot loops.
- Subscriber fan-out, event emission, and durable scans need default caps or
  benchmark-relative budgets.

Required budget rows before delegation:

| Hot path | Budget requirement |
|---|---|
| API target resolution | Query shape, maximum store calls, maximum scanned rows, fixture size, proof command, threshold, owner, and renderer parity tests. |
| `resolveLiveSessionByPathAlias` | Decision to index, remove, or keep with explicit scan budget. If kept, budget must name maximum session rows scanned and prove newest-created tiebreaker behavior on a large fixture. |
| Reconciler fact compilation | Store query count, subprocess count, runtime probe count, maximum session/work rows scanned, partial-snapshot behavior, proof command, and owner. |
| Event recovery scans | Scan interval or trigger, maximum rows, duplicate handling, idempotency key, completion marker, and crash-retry proof. |

## Vocabulary Lifecycle

Shared vocabulary has four states:

| State | Meaning |
|---|---|
| `documented` | Existing behavior or code vocabulary named for traceability only. |
| `private` | May exist inside one slice package or adapter, but is not a shared contract. |
| `provisional` | Design-only upper bound for a future slice; cannot appear in public API, generated clients, event payloads, or cross-slice contracts. |
| `delegating` | A real production caller has delegated to the type or field and its checkpoint row names exact fields, non-goals, proof tests, and expansion rules. |

Future terms such as command result, runtime intent, session fact event, and
generic conflict stay `provisional` until a production caller proves the exact
field set. Flat optional envelopes are not acceptable for new shared types; use
tagged result kinds or per-kind structs when only some fields are meaningful.

## Shape

Prefer small operation-specific APIs over a broad `SessionService`.

Good shape:

```text
Target facts -> read-only target classifier -> caller-specific adapter
Wake facts -> wake decider -> wake command
Close facts -> close decider -> close command
```

Avoid:

- one large `SessionFacts` struct
- a generic command bus
- event sourcing as the first step
- moving work, mail, extmsg, provider, or pool policy into `internal/session`

## Boundaries

`internal/session` should own:

- lifecycle projection and transition rules
- session identity and target classification rules
- session-owned lifecycle and wake metadata mutations after their command
  contracts exist
- pure decisions that can be unit-tested without stores or providers

Callers should own:

- API and CLI rendering
- work assignment and release policy outside session facts
- mail and external-message delivery policy
- runtime provider execution
- reconciler scaling, budget, progress, and alert policy

## Refactor Rules

For each operation:

1. Pick one current behavior cluster.
2. Read the matching `REQUIREMENTS.md` scenario rows.
3. Add or keep characterization tests for the current caller behavior.
4. Add or reference the Slice 0 artifact rows that cover the cluster.
5. Add a small session-owned decider or command only after its contract exists.
6. Move one caller to it.
7. Keep the old behavior unless the requirements row changes with owner
   approval.
8. Delete or shrink duplicated caller logic after parity is proven.

The test should prove the behavior the user sees, not every internal branch.

## Backlog

0. Slice 0 universal evidence preflight: create the non-mutating inventories,
   schemas, validators, negative fixtures, proof transcripts, stale-requirement
   repair/retirement rows, and workflow close gate.
1. API query target classification: adopt read-only target classification for
   API query-side lookup while preserving the exact
   `resolveSessionTargetIDWithContext` precedence, Huma/legacy error rendering,
   closed lookup rules, config-orphan rejection, path-alias behavior, and
   `RepairEmptyType` quarantine.
2. Additional target-classification surfaces: adopt CLI, mail, extmsg,
   assignee, nudge, attach, transcript, logs, and pool resume one surface at a
   time after each compatibility matrix passes.
3. Explicit wake: move wake eligibility/conflict decisions behind a
   session-owned operation after worker-boundary routing, per-slice atomic
   command rows, event rows, diagnostics rows, cost rows, and migration rows
   exist.
4. Close and identity retirement: keep close semantics, worker-boundary routing,
   event emission, and work-release recovery clear without scattering lifecycle
   metadata writes.
5. Runtime start: fold prepare/commit/rollback metadata ownership into one
   command after stale-token, partial-write, provider compensation, and recovery
   tests exist.
6. Reconciler facts: extract only narrow lifecycle eligibility facts; keep pool
   scaling, work demand, provider health, progress, budgets, and alerts in the
   reconciler.

## Non-Goals

- Do not rewrite the reconciler wholesale.
- Do not introduce a large facade before one small operation proves value.
- Do not move work, mail, extmsg, or provider-specific runtime policy into
  `internal/session`.
- Do not make event sourcing the first implementation step.
- Do not use design review feedback as a substitute for readable requirements
  and tests.
- Do not grow Slice 0 into behavior-moving preflight work; it proves boundaries
  and gates only.
