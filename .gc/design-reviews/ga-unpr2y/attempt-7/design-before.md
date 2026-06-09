# Session Boundary Design

| Field | Value |
|---|---|
| Status | Draft backlog |
| Behavior source | `REQUIREMENTS.md` |
| Scope | Technical architecture and extraction plan for session ownership |
| Latest design-review disposition | Attempt 6 global verdict `block`; this revision is an `iterate` response |

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

## Attempt 4 Review Response

<!-- REVIEW: added per attempt-4-global-synthesis -->

Attempt 4 still blocked decomposition. This revision keeps
`design_review.verdict=iterate` and makes the remaining review concerns explicit
implementation gates. No implementation bead may be spawned from this design
until the gate for its slice names the source inventory, scenario rows, caller
surfaces, mutation keys, recovery authority, diagnostic proof, and rollback
contract it owns.

## Attempt 5 Review Response

<!-- REVIEW: added per attempt-5-global-synthesis -->

Attempt 5 still blocks decomposition. This revision keeps
`design_review.verdict=iterate` and treats the remaining issues as exit gates,
not guidance. No implementation bead may be generated from this design until
the gate for its slice is backed by source inventory, parity proof, diagnostics,
rollback instructions, and performance evidence in the same change.

| Finding | Design response | Exit gate |
|---|---|---|
| Writer inventory and mutation guards are not source-complete. | The source inventory must be regenerated from AST/search output, split by concrete call site, and checked by a shrink-only guard before any writer conversion. | A source-verified inventory artifact names every production writer, exact or dynamic key class, target-bead proof, allowlist row, guard fixture, owner slice, and retirement condition. |
| Command concurrency and atomicity are not enforceable enough. | The selected model is tokened or revisioned commit markers plus command-time revalidation and bounded repair; in-process locks are only duplicate-work reduction. | Each command cluster proves its store primitive, conflict result, partial-state matrix, stale-success handling, provider compensation, and event ordering before delegation. |
| Target classification conflates identity, policy, repair, and materialization. | Slice 1 is raw candidate collection only; operation policy, audited repair, and materialization are separate adapters or commands with command-time revalidation. | Each surface has a transition table preserving resolver order, side effects, recipient-set behavior, closed lookup, and materialization before adopting the classifier. |
| Events and recovery do not yet guarantee convergence. | Current events are hints unless a slice explicitly migrates typed payloads; critical recovery converges from durable fact scans. | Every critical reaction names scan authority, cadence, store coverage, completeness guard, idempotency key, stale-event rule, and crash-after-commit tests. |
| Reconciler and session boundaries still leak controller policy. | The boundary is pre-decider demand assembly, session eligibility mask, then reconciler-owned demand resolution and health/progress/budget gates. | Pure session deciders live in a mechanically guardable file set and consume immutable facts only; demand, pool scaling, provider health, progress, budgets, and alert dedupe remain controller-owned. |
| Behavior parity proof is not decomposable. | Scenario rows must be propagated into every slice and proof command before code moves. | A citation-freshness gate fails on missing paths, stale assertions, and commit-only proof that is not backed by current tests or owner-approved requirement amendments. |
| Operability diagnostics are not concrete enough. | Diagnostics must map to named trace, doctor/session-inspect, API, CLI, event, or log surfaces. | Each new reason/outcome has centralized vocabulary, rendering tests, data source, fixability, and partial-fact behavior. |
| Migration coexistence and rollback are too loose. | Each slice owns a per-key owner matrix and bake/revert rule for cross-family operations. | Bead metadata names converted callers, bake exceptions, retired rows, guard changes, proof commands, rollback files, and one-writer restoration criteria. |
| Scope and vocabulary controls are advisory. | Vocabulary checkpoints are required metadata for every new shared fact, result, event, or diagnostic field. | A type or field cannot leave a private slice until first caller, exact fields, rule-of-two evidence, non-goals, and future-slice bounds are recorded. |
| Performance and backpressure lack numeric proof. | Hot-path slices must add query-count, benchmark, or source-proof budgets with default caps where no historical benchmark exists. | A slice touching lookup, scans, runtime observation, event emission, subscriber fan-out, or reconciler loops fails review without numeric or baseline-relative evidence. |

## Attempt 6 Review Response

<!-- REVIEW: added per attempt-6-global-synthesis -->

Attempt 6 still blocks decomposition. This revision keeps
`design_review.verdict=iterate` and adds a non-mutating Slice 0 preflight before
any mutation-owning implementation slice. The preflight is mandatory because the
review found that prose inventory rows, stale citations, and advisory
vocabulary checks are not enough to constrain decomposition.

| Finding | Design response | Exit gate |
|---|---|---|
| Writer inventory and mutation guards are not source-complete. | Slice 0 must generate or verify a source inventory from the checkout, install the additive static guard, and fail on unmapped writers before any writer conversion starts. | `internal/session/BOUNDARY_INVENTORY.md`, `cmd/gc/testdata/session_boundary_guard/allowlist.yaml`, guard fixtures, and `go test ./cmd/gc -run TestSessionBoundaryGuard -count=1` exist and pass with shrink-only allowlist rows. |
| Behavior parity and citation freshness are not enforceable. | `internal/session/SCENARIO_PARITY.yaml` becomes the canonical bead-generation source for all `REQUIREMENTS.md` rows and slice owners. | `go test ./internal/session -run TestScenarioParityFreshness -count=1` fails on missing paths, stale assertions, commit-only current proof, or non-ancestor commit citations. |
| Command atomicity and failure ordering are under-specified. | Every command cluster needs a command-applier row with snapshot preconditions, token or phase marker, allowed write primitive, stale-success behavior, partial-state matrix, repair authority, event order, and race tests. | Runtime-start and close cannot delegate callers until their applier rows prove provider-side-effect compensation and durable fencing. |
| Target classification conflates facts, adapter policy, repair, and materialization. | The raw classifier is read-only candidate collection; adapter-owned policy creates the only authoritative `selected` result and writable adapters perform repair/materialization after revalidation. | Per-surface fixtures preserve resolver chain, recipient set, side effects, error shape, materialization, closed handling, alias history, and ambiguity before delegation. |
| Reconciler/session ownership still leaks policy. | Session owns only lifecycle and identity eligibility over immutable facts; reconciler owns demand, scale, health, progress, restart budget, circuit, alert dedupe, trace policy, and idle-sleep decisions. | Pure-decider import guards and runtime-observation completeness tests fail destructive close, drain, rollback, release, cleanup, and restart branches on stale or partial facts. |
| Event recovery, diagnostics, and performance gates are not executable. | Durable scans are the mandatory backstop for critical reactions; events and synchronous cascades are latency or diagnostic aids. Diagnostics and performance gates name concrete trace, doctor, API/CLI, query-count, and benchmark artifacts. | Event-miss and crash-after-commit tests prove scan convergence; hot-path slices include query-count or benchmark commands and subscriber backpressure budgets. |
| Migration coexistence and rollback remain too loose. | Every implementation bead records converted callers, bake allowances, retired rows, guard-row changes, rollback files, proof commands, and one-writer proof. | A slice cannot close while any owned key has two production writers with different validation after its guard lands. |
| Scope and vocabulary controls are advisory. | Vocabulary checkpoint metadata is enforced by a failing test and each shared type names first caller, demanded fields, rule-of-two status, non-goals, provisional bounds, and owning slice. | `go test ./internal/session -run TestVocabularyCheckpoints -count=1` blocks shared types or fields without checkpoint evidence. |
| Current-attempt persona artifacts are stale. | This is workflow plumbing outside `internal/session/DESIGN.md`; the design records the defect but cannot repair the formula artifact writer. | The design-review workflow must either write persona syntheses under the current `$ATTEMPT_DIR` or materialize a source manifest; until then this design consumes only bead-stamped global synthesis paths. |

The source-verified inventory, static guard, scenario parity source, vocabulary
checkpoint file, performance baselines, and workflow artifact writer are not
created by this design document. This document therefore continues to block
decomposition until those artifacts exist; it does not mark those concerns
resolved by prose.

### Source-Verified Mutation Inventory Gate

Slice 0 is a non-mutating preflight. It does not move a writer, change
behavior, or introduce a new command API. It must land before any
mutation-owning slice and must be safe to run repeatedly on the current
checkout.

Slice 0 artifacts:

| Artifact | Purpose |
|---|---|
| `internal/session/BOUNDARY_INVENTORY.md` | Generated-or-verified inventory with one row per production writer, clearer, repair path, bridge, package mutator, manager method, exported patch builder, and raw bead mutation route that can affect session beads. |
| `cmd/gc/testdata/session_boundary_guard/allowlist.yaml` | Shrink-only allowlist keyed by stable inventory ID, owner slice, exact path/symbol, key family, reason, expiry condition, and retirement test. |
| `cmd/gc/testdata/session_boundary_guard/fixtures/` | Positive and negative fixtures for raw store writes, dynamic metadata batches, manager methods, package mutators, patch-map extension, generic bridge writes, top-level type/status mutation, repair helpers, and API/CLI manager construction. |
| `cmd/gc/session_boundary_guard_test.go` | Additive static guard that scans production roots and fails on unmapped session-boundary mutations. |

Slice 0 proof command:

```bash
go test ./cmd/gc -run 'TestSessionBoundaryGuard|TestSessionBoundaryInventoryFresh' -count=1
```

The guard may be introduced with legacy allowlist rows, but those rows are
exceptions, not approval. Adding a row after Slice 0 requires a design or scoped
`AGENTS.md` migration update; deleting or narrowing rows is always allowed.

The writer inventory below is a planning baseline, not an approval artifact.
Before any mutation-owning slice starts, the implementation bead must create or
update a source-verified inventory artifact checked by tests. The inventory row
schema is:

| Field | Meaning |
|---|---|
| stable ID | `W-###` or slice-local child ID used by guards and bead metadata. |
| source path and symbol | File, function, receiver, and helper chain that can write a session bead or route a write. |
| operation | create, close, metadata set, metadata batch, patch apply, update closure, repair, materialize, wake, drain, runtime-start, nudge, or bridge. |
| key family and exact keys | Owned taxonomy row plus string-literal keys or dynamic-key source. |
| target bead proof | How the code proves the target is a session bead, proves it is not a session bead, or is unsafe until converted. |
| current owner | Package or adapter currently deciding the mutation. |
| intended command owner | Session command slice that will own validation and mutation. |
| exception status | allowed, legacy allowed-during-bake, repair-only, non-session-only, test-only, or blocked. |
| retirement condition | Exact caller conversion, guard tightening, and tests needed to delete the exception. |
| guard rule | AST/symbol pattern that catches regressions for this row. |
| proof | Test, static guard, or source command that fails when the row drifts. |

The source scan must include at least these production paths before the row can
be marked source-complete:
`cmd/gc/session_reconcile.go`, `cmd/gc/session_reconciler.go`,
`cmd/gc/session_resolve.go`, `cmd/gc/session_name_lookup.go`,
`cmd/gc/session_circuit_breaker.go`,
`cmd/gc/soft_reload.go`, `cmd/gc/cmd_wait.go`, `cmd/gc/cmd_stop.go`,
`cmd/gc/cmd_nudge.go`, `cmd/gc/cmd_prime.go`,
`cmd/gc/adoption_barrier.go`,
`cmd/gc/cmd_bd_store_bridge.go`, `cmd/gc/session_lifecycle_parallel.go`,
`cmd/gc/session_beads.go`, `cmd/gc/cmd_session_wake.go`,
`cmd/gc/cmd_session.go`, `cmd/gc/cmd_handoff.go`,
`internal/api/session_resolution.go`,
`internal/api/handler_sessions.go`, `internal/api/handler_beads.go`,
`internal/api/handler_session_create.go`, `internal/api/session_manager.go`,
`internal/api/huma_handlers_sessions_command.go`,
`internal/session/chat.go`, `internal/mail/beadmail/beadmail.go`, and
`internal/extmsg`.

Attempt 5 named additional symbol-level sources that the inventory cannot
collapse into broad package rows:

- `session.RepairEmptyType` callers in API, mail, named-config, metadata
  candidate, and chat paths.
- `session.WakeSession`, `RequestWakePatch`, `PreWakePatch`,
  `CommitStartedPatch`, `ClosePatch`, and `RetireNamedSessionPatch` callers
  or external patch-map application sites.
- `Manager.CloseDetailed`, `Manager.Suspend`, `Manager.UpdatePresentation`,
  manager construction, and any direct manager method call outside approved
  worker/API construction points.
- Top-level session-bead type/status mutation through `beads.Store.Update`,
  `Close`, `Create`, `SetMetadata`, or `SetMetadataBatch`.
- Runtime identity backfills in `internal/session/chat.go`, including
  `session_key`, `instance_token`, hash, and continuation-reset keys.

A row is source-complete only when the inventory records the command used to
find these symbols, the resulting path and symbol set, and the reviewed
exception or retirement row for each production hit. A package-level row may
stay in the overview table, but the implementation gate must split it into
call-site child rows before code moves.

The scan must classify direct owned-key writes, dynamic metadata batches,
patch-map aliases, raw `Create`/`Update`/`Close`, generic bridge writes,
`RepairEmptyType`, runtime identity setters and clearers, manager methods,
package-level session mutators, and exported patch builders. Unknown dynamic
keys on a path that may carry a session bead are blocked until the inventory
proves the target bead type or constrains the key set.

Static guards are required before writer conversion, not after cleanup. They
must fail on new unmapped production calls to raw bead-store writes, patch map
application, local metadata wrappers, manager/package-level session command
bypasses, dynamic metadata batches that may reach session beads, and API or CLI
paths that construct session managers outside approved exceptions. The allowlist
is shrink-only: a slice may narrow or delete its own rows, but adding a bypass
row requires updating this design or the scoped `AGENTS.md` migration notes.

### Full Scenario Parity Gate

Scenario parity has one canonical source for decomposition:
`internal/session/SCENARIO_PARITY.yaml`. The YAML mirrors every scenario row in
`REQUIREMENTS.md`, assigns exactly one primary owner slice, lists secondary
slices that must preserve the row, and names the current proof command or path.
Task/bead generation must read this file instead of copying proof bullets from
this design, the requirements ledger, or backlog notes by hand.

The parity source row schema is:

| Field | Meaning |
|---|---|
| `scenario_id` | Exact `REQUIREMENTS.md` row ID. |
| `primary_slice` | Slice that owns behavior migration or `baseline` when no migration is planned. |
| `secondary_slices` | Slices that must prove they did not regress the row. |
| `current_oracle` | Runnable test, source path, issue, or owner-approved requirement amendment. |
| `operation_surfaces` | API, CLI, package, mail, extmsg, worker, reconciler, transcript, doctor, or dashboard surfaces the row constrains. |
| `compatibility_shape` | Required status codes, error body, request ID, async behavior, stdout/stderr, JSON shape, exit code, side effects, recipient set, and materialization behavior where applicable. |
| `freshness_command` | Command or guard that proves the oracle still exists and covers the behavior. |

Freshness proof:

```bash
go test ./internal/session -run TestScenarioParityFreshness -count=1
```

That test fails when a scenario row is missing from the parity source, a path no
longer exists, a commit hash is used as current proof without a live test or
source path, a cited commit is not an ancestor of `HEAD` but is treated as
current evidence, or an oracle assertion no longer covers the required
behavior. Commit-only proof may remain as historical context, but it cannot
unblock extraction without a current checkout proof or owner-approved
requirements amendment.

Each implementation bead must attach a parity artifact with one row per
`REQUIREMENTS.md` scenario it touches. The artifact must state the current
behavior oracle, exact preserved output, new proof, and whether the row is
out-of-scope. For API and CLI paths, exact output means status code, response or
error body, request ID behavior, stdout, stderr, JSON shape, and exit code. For
mail, extmsg, assignee normalization, nudge, attach, inspect, log, transcript,
and materialization paths, exact output means the current resolver chain,
fallback behavior, side effects, and diagnostic rendering.

Current row ownership baseline:

| Scenario row | Owning design slice or status | Proof gate |
|---|---|---|
| `SESSION-LIFE-001` | lifecycle projection baseline | Projection tests must remain the oracle. |
| `SESSION-LIFE-002` | runtime-start and reconciler eligibility | Pending-create tests plus runtime-start recovery tests. |
| `SESSION-LIFE-003` | wake/hold/drain eligibility | Hold/quarantine projection and reconciler parity tests. |
| `SESSION-LIFE-004` | runtime-start repair and reconciler eligibility | Stale-create rollback and projection tests. |
| `SESSION-LIFE-005` | runtime-start and runtime fact boundary | Runtime liveness parity and missing-runtime tests. |
| `SESSION-LIFE-006` | target/config eligibility | Missing-config projection and wake conflict tests. |
| `SESSION-LIFE-007` | wake, close, API lifecycle | Terminal wake conflict tests across package, API, and waits. |
| `SESSION-LIFE-008` | projection guard | Static guard for user-facing raw-state reads. |
| `SESSION-STATE-001` | state-machine baseline | Reducer table remains command oracle. |
| `SESSION-STATE-002` | command validation | Illegal transition tests for each command slice. |
| `SESSION-STATE-003` | UI/API affordances | Rendering tests consume reducer output, not raw metadata. |
| `SESSION-ID-001` | identity validation baseline | Name validation tests remain oracle. |
| `SESSION-ID-002` | alias validation baseline | Alias validation and collision tests remain oracle. |
| `SESSION-ID-003` | target classification | Package resolver parity including no template/config fallback. |
| `SESSION-ID-004` | target classification | Allow-closed resolver parity. |
| `SESSION-ID-005` | target classification and reconciler eligibility | Named reservation projection and materialization tests. |
| `SESSION-ID-006` | close/retire identity | Canonical/historical identity and duplicate tests. |
| `SESSION-ID-007` | wake and close/retire | Terminal named wake and successor materialization tests. |
| `SESSION-ID-008` | API/Huma target classification | API rejection of factory/config targets. |
| `SESSION-ID-009` | mail target classification | Mail send/query recipient and no-materialization tests. |
| `SESSION-ID-010` | identity/runtime-start | Concrete runtime identity tests for aliasless multi-session. |
| `SESSION-START-001` | runtime-start consistency cluster | Prepare/commit/rollback and pending-clear atomicity proof. |
| `SESSION-START-002` | runtime-start repair | Stale create rollback and start-pending migration tests. |
| `SESSION-START-003` | explicit wake | Wake intent, blocker clearing, and wait cancellation tests. |
| `SESSION-START-004` | API/Huma materialization | Reserved named suspend materialization tests. |
| `SESSION-START-005` | close/retire | Close cleanup, MCP snapshot, wake/hold override, and identity retirement tests. |
| `SESSION-START-006` | close/retire | Provider-stop failure keeps bead open. |
| `SESSION-START-007` | update/wake conflict | Template override safety tests. |
| `SESSION-START-008` | runtime-start ordering | Parallel lifecycle ordering tests. |
| `SESSION-RECON-001` | worker boundary | Existing worker-boundary guard plus new command-bypass guard. |
| `SESSION-RECON-002` | reconciler demand | Restore or replace missing scale-from-zero proof before extraction. |
| `SESSION-RECON-003` | reconciler demand | Restore or replace missing rig-session cold-wake proof before extraction. |
| `SESSION-RECON-004` | wake/hold/drain eligibility | Hold-vs-wake parity tests. |
| `SESSION-RECON-005` | reconciler demand | Pool demand-gating tests remain controller-owned. |
| `SESSION-RECON-006` | provider health boundary | Restore or replace missing provider-health proof before extraction. |
| `SESSION-RECON-007` | progress/circuit boundary | Restore or replace missing progress proof before extraction. |
| `SESSION-WORK-001` | close subscriber scan | Idempotent work-release tests from durable facts. |
| `SESSION-WORK-002` | close subscriber scan | Confirmed-dead release tests from complete runtime/provider facts. |
| `SESSION-WORK-003` | work recovery scan | Orphan pool step release proof from live work queries. |
| `SESSION-WORK-004` | drain recovery | No-wake drain cancel and assigned-work recovery tests. |
| `SESSION-RUNTIME-001` | runtime fact boundary | Provider observation fallback tests. |
| `SESSION-RUNTIME-002` | runtime fact boundary | Pending/respond missing runtime tests. |
| `SESSION-RUNTIME-003` | runtime-start/provider shell | ACP creation/resume/reroute tests. |
| `SESSION-RUNTIME-004` | submit/provider shell | Stop-turn tests remain provider-shell owned. |
| `SESSION-RUNTIME-005` | target/read-side lookup | Transcript lookup parity including provider fallback. |

Rows whose cited proof files are missing are blocked, not accepted on memory of
past behavior. The implementation must restore the deleted proof, replace it
with an equivalent current test, or update `REQUIREMENTS.md` with owner-approved
behavior before extraction.

### Atomicity And Runtime-Start Recovery Gate

The selected write model is tokened prepare/commit plus deterministic repair.
This is not a claim that every store operation is atomic. Each slice must prove
the concrete store primitive it uses:

| Store primitive | Default assumption | Required proof or mitigation |
|---|---|---|
| `SetMetadataBatch` | Atomic only for the active store path under test. | Unit or integration proof for all-session stores used by the slice, or repair rows for every visible partial key subset. |
| `Update` closure | Atomic only for one bead mutation in one store invocation. | Conflict tests for stale reads, concurrent writers, and closure errors. |
| `Close` | Atomic for bead status plus close metadata only if the store proves it. | Idempotent close tests and post-close subscriber scan recovery. |
| `Create` plus metadata writes | Not atomic across provider side effects or follow-up writes. | Prepare marker before side effect, tokened commit after side effect, and stale-create repair. |
| `BdStore.Tx` or backend transaction | Not assumed portable. | Slice may rely on it only after documenting backend coverage and fallback. |
| External store or bridge writes | Unknown. | Fail closed unless target bead type and owned-key set are proven. |

Runtime start is one fenced consistency cluster. The prepare command writes the
new `instance_token`, generation, pending-create marker, and start hashes before
provider start. The provider shell starts the runtime and returns observed
runtime identity. The commit command re-reads the bead, verifies the same
`instance_token`, writes active/session-key/hash facts, clears pending-create
markers, and emits only post-commit facts. Recovery handles these states:

| Failure point | Durable facts | Required recovery |
|---|---|---|
| Crash before provider start | prepared token, no runtime identity | Roll back stale creating after grace. |
| Provider start fails | prepared token plus provider error | Roll back or mark failed-create without clearing newer tokens. |
| Provider start succeeds, commit fails | prepared token plus observed runtime identity | Retry commit if token still current; otherwise stop or orphan-handle the runtime by token. |
| Commit partially writes active facts | token plus partial active/session-key/hash keys | Repair to committed active or rollback according to observed runtime identity. |
| Successor generation appears | newer token/generation | Old provider identity cannot mutate the bead; cleanup is best-effort by runtime identity. |

### Durable Reaction Recovery Gate

Events accelerate and explain; durable scans converge. Critical reactions must
declare their recovery matrix before relying on an event.

| Reaction | Source facts | Assignee/key set | Store routing | Cadence | Completeness guard | Stale-event handling | Duplicate behavior | Tests |
|---|---|---|---|---|---|---|---|---|
| Work release after close | closed bead, retired identity, open work assigned to bead ID/session name/configured identity | bead ID, current/historical names where requirements allow, configured identity | active store and rig stores from controller context | same controller tick or next reconciler scan | suppress close-driven release if work query is partial unless synchronous close path already proved complete | older event cannot release work for a reopened or successor identity | idempotent reopen/release | close with skipped event emission and duplicate scan. |
| Drain cancel/complete | drain generation, assigned-work query, worker ack metadata | session ID plus drain generation and assigned bead IDs | active store and rig stores | reconciler tick | cancel or defer destructive action on partial work query | stale generation ignored | duplicate cancel/complete no-ops | assigned-work reappears, query failure, duplicate ack. |
| Wake demand recovery | wake metadata, lifecycle projection, config identity, work demand | session ID, configured identity, template target | session store plus work-demand stores | desired-state rebuild each tick | config read failure blocks mutation; work query partial suppresses destructive sleep/drain | terminal or newer token supersedes wake event | duplicate wake request remains one desired action | durable metadata still wakes without event delivery. |
| Identity retirement cleanup | closed terminal facts, retired identity markers, bindings/work refs | configured identity, alias/session name, canonical bead ID | session store plus binding/work stores | close tick and periodic repair | partial binding/work query defers cleanup | newer canonical identity supersedes old event | cleanup is idempotent | terminal close, successor materialization, duplicate canonical. |
| Mail/extmsg binding cleanup | session facts plus binding records | mailbox identity, extmsg member target, session ID | mail/extmsg stores plus session store | adapter retry or controller repair | missing binding store logs and defers, not silent success | current binding generation supersedes old event | duplicate cleanup no-ops | skipped event emission and retry after store recovery. |

Trace outcomes must report observed convergence, not imaginary event delivery.
Use `durable-scan-converged`, `durable-scan-deferred`,
or `durable-scan-blocked` instead of claiming that the system observed skipped
event delivery. Use `event-emission-failed` only after the recorder API returns
an error that command code can observe; until then, missed or skipped events are
proved through durable-scan tests and reported as scan outcomes.

### Session Eligibility Versus Controller Demand Gate

`internal/session` may own lifecycle eligibility; it must not absorb controller
demand policy.

| Layer | Inputs | Allowed owner | Forbidden movement |
|---|---|---|---|
| Session eligibility | lifecycle projection, terminal status, holds, quarantine, missing config, identity conflict, explicit wake facts | `internal/session` deciders and commands | work aggregation, pool capacity, alert dedupe, restart budgets. |
| Controller demand | routed work, attached sessions, pool min/max, scale-check output, nested caps, in-flight reuse | controller/reconciler | session package must not decide how many workers a pool wants. |
| Runtime/provider gates | provider health, runtime liveness, progress, startup grace, circuit state | runtime adapter and reconciler | session package must not probe providers or consume restart budgets. |
| Destructive action proof | complete session/work/runtime/config facts | command adapter plus session command validation | destructive close/drain/rollback from stale, timed-out, or partial facts. |

Guards must reject store, runtime, provider-health, and work-query imports in
pure session deciders. Adapters may gather facts, but deciders receive immutable
fact structs only.

### Caller Routing And Command Construction Gate

Only `internal/session` may implement session commands. `internal/worker` may
expose production handles that delegate to those commands. Other production
callers must either route through `worker.Handle` or through an approved
session command factory named by the slice.

| Caller surface | Current risk | Required end route | Guard requirement |
|---|---|---|---|
| production CLI in `cmd/gc` | direct manager, patch, and metadata bypasses | `internal/worker.Handle` for lifecycle operations; read-only projection helpers for display | existing worker-boundary guard plus new command-bypass and raw-write guard. |
| legacy API handlers | direct manager construction and materialization | typed adapter -> worker/session command factory -> session command | API exception row must shrink as handlers convert. |
| Huma API handlers | typed wire can still route to raw helpers | Huma typed input/output -> adapter policy -> command/classifier | no raw JSON/map wire; OpenAPI/dashboard checks for any schema change. |
| reconciler/controller | mixed demand, runtime, and lifecycle mutation | fact readers and provider shell around narrow session commands | guard prevents pure decider imports from controller/runtime packages. |
| worker boundary | canonical production boundary | constructs command dependencies and calls session commands | allowed constructor sites listed by inventory ID. |
| mail/extmsg | target-resolution and binding side effects | classifier for session target identity; binding cleanup stays domain-owned | guard proves non-session bead writes or routes session writes through commands. |
| tests/fixtures | direct state construction | allowed when file is test/fixture | excluded from production guard, covered by test naming. |
| doctor/migration/repair | direct normalization writes | audited repair helper with trace/log evidence | repair-only allowlist with before/after and expiry. |

Internal classifier conflict details are not automatically wire-visible. Huma
and CLI adapters must choose typed, compatibility-preserving output fields
before exposing candidate, negative-kind, or conflict details.

### Slice Coexistence, Bead Metadata, And Revert Gate

Implementation beads must include these metadata fields so workflow gates can
verify migration ownership:

- `session_design.slice=<number-or-name>`
- `session_design.converts=<inventory IDs or caller surfaces>`
- `session_design.allows_during_bake=<legacy inventory IDs>`
- `session_design.retires=<inventory IDs>`
- `session_design.guard_rows_added=<inventory or allowlist IDs>`
- `session_design.guard_rows_removed=<inventory or allowlist IDs>`
- `session_design.scenarios=<REQUIREMENTS IDs>`
- `session_design.guard=<static guard test path>`
- `session_design.proof_commands=<exact commands>`
- `session_design.rollback=<files, switches, and caller routes to revert>`
- `session_design.one_writer_proof=<owned keys and surviving production writer>`
- `session_design.unresolved_risks=<accepted residual risks or none>`

No slice can be marked done while old and new writers mutate the same owned key
with different validation. Runtime start, close/retire, wake/hold/drain,
API materialization, and repair each ship as their own fenced cluster with a
single owner for the keys listed in that cluster.

The one-writer proof is mechanical: for each owned key family touched by the
slice, it lists every production writer before the slice, the converted route,
legacy rows allowed during bake, the guard row that prevents new bypasses, and
the exact condition that retires each legacy row. Worker-boundary compliance
does not prove mutation-boundary compliance; both guards must pass.

### Repair, Operability, And Budget Gate

Repair and normalization writes, including `RepairEmptyType`, are production
mutations when they can touch a session bead. They must go through one audited
helper or remain an explicitly expiring exception. The helper contract is:
read complete facts, prove target bead type, name the owned keys, validate
freshness/preconditions, write through the narrowest store operation, emit
trace/log evidence with before/after values, surface errors to the caller, and
render doctor evidence for applied, skipped, unsafe, partial, and failed repair
results.

Each operation must map `accepted`, `rejected`, `blocked`, `no-op`, `partial`,
`deferred`, `failed`, `repair-applied`, `repair-skipped`,
`durable-scan-converged`, and `durable-scan-deferred` to trace, logs, doctor,
API, and CLI where that surface exists. A diagnostic promise without a rendering
test does not satisfy the gate.

Hot-path slices must include a query-count test, benchmark, or source proof for
target lookup, path-alias fallback, reconciler fact materialization,
close/work-release scans, runtime-start repair scans, subscriber fan-out, event
emission, throttle writes, and snapshot rebuilds when they touch those paths.

### Vocabulary Checkpoint Gate

Every shared result or fact type needs a checkpoint before it leaves a private
slice package:

| Checkpoint field | Required content |
|---|---|
| first delegated caller | The first real caller and why its existing local type cannot remain private. |
| exact demanded fields | Fields populated for that caller, with per-kind population rules. |
| policy dimensions | Operation policy values the caller can set. |
| non-goals | Fields and decisions explicitly not carried by the type. |
| rule-of-two evidence | Second caller with the same exact field set, or a decision to keep the type private. |
| provisional bounds | Later-slice event/runtime fields marked non-binding until adopted. |

Broad optional envelopes are rejected. Use tagged unions or per-kind structs
when fields only make sense for one result kind.

The checkpoint must appear in implementation bead metadata before the type is
shared:

- `session_design.vocab=<type-or-field-name>`
- `session_design.vocab.first_caller=<path:symbol>`
- `session_design.vocab.fields=<exact fields introduced>`
- `session_design.vocab.rule_of_two=<second caller or private-until-needed>`
- `session_design.vocab.non_goals=<decisions and fields excluded>`
- `session_design.vocab.owner_slice=<slice that may expand or retire it>`
- `session_design.vocab.provisional_bounds=<future fields treated as closed upper bound>`

Executable checkpoint source:

| Artifact | Gate |
|---|---|
| `internal/session/VOCABULARY_CHECKPOINTS.yaml` | Records shared type or field, first delegated caller, exact demanded fields, rule-of-two status, non-goals, provisional bounds, and owner slice. |
| `go test ./internal/session -run TestVocabularyCheckpoints -count=1` | Fails when a shared type/field named by the slice lacks a checkpoint, carries fields not demanded by its first caller, or expands beyond its provisional bounds without an adopting caller. |

Future-slice vocabulary is a maximum bound, not scaffolding to build now. A
reviewer must reject a shared type that has no first delegated caller, carries
fields not populated by that caller, or exists only to make a later slice easier.

### Workflow Artifact Note

The attempt-scoped persona-output placement bug is workflow plumbing outside
`internal/session/DESIGN.md`. This design cannot fix it; the workflow formula
must write attempt-N persona syntheses under attempt-N and fail synthesis when
current-attempt artifacts are absent. Until that is fixed, this design treats
the global synthesis file, not stale persona paths, as the review input.

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
  candidate conflict details, adapter-owned policy inputs, and adapter
  selection results.
- Slice 2 may introduce explicit-wake facts, wake command results, and wake
  conflicts.
- Later slices introduce lifecycle command facts, reconciler facts, and runtime
  intent results only when those slices begin.

Shared vocabulary checkpoints:

| Proposed vocabulary | First allowed caller | Required fields at introduction | Non-goals and expansion rule |
|---|---|---|---|
| `TargetCandidate` | Target classification adapter parity tests | kind, source surface, normalized token, session ID/name/alias/config identity, status, closed flag, conflict reason | Does not carry command mutation policy. Add fields only when a second surface needs the same field. |
| `TargetSelection` | API session target resolver adapter | selected candidate, ordered candidates considered, negative kind, retry/fallthrough directive | The adapter-owned `selected` field is the only authority callers may act on. Diagnostic candidates are non-authoritative. Not a universal session lookup facade. Each surface keeps its current resolver chain until parity tests approve delegation. |
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
| W-019 | `internal/mail/beadmail/beadmail.go` `RepairEmptyType` during session reads | repair empty session type | Repair helper with trace/log evidence if promoted | repair slice | Bounded repair exception. |
| W-020 | `internal/extmsg/*`, mail stores, convoy/order stores | non-session bead metadata and closes | Stay outside session unless writing a session bead owned key | none | Guard must prove non-session bead discrimination. |
| W-021 | Generic `beads.Store.Update`, `SetMetadata*`, `Close`, `Create` bridges | unknown | Must be classified by target bead type and key set | per slice | Unsafe by default when a session bead can flow through. |
| W-022 | `internal/session/chat.go` runtime identity and continuation backfills | runtime identity, create/start lease | Runtime-start repair command or chat-specific audited repair | 3, repair slice | Internal but not source-complete until token and hash writes are inventoried separately. |
| W-023 | `internal/session/waits.go` `WakeSession` wait close/update plus session wake batch | wake/hold/wait | Explicit wake and wait-hold commands | 2, 5 | Package-level mutator; allowed only behind command or internal caller row. |
| W-024 | `internal/session/named_config.go` and `internal/session/metadata_candidates.go` `RepairEmptyType` | repair empty session type, identity read-side normalization | Audited repair helper or typed repair-needed result | repair slice | Internal repair exception; callers must not hide repair in read-only classification. |
| W-025 | `internal/api/huma_handlers_sessions_command.go` direct repair, wake, suspend, close, and presentation updates | repair, wake, lifecycle, close, identity presentation | Typed API adapter over worker/session commands | 1, 2, 4, repair slice | Root API exception until retired; must preserve Huma typed-wire outputs. |
| W-026 | `cmd/gc/cmd_session.go`, `cmd/gc/cmd_handoff.go`, `cmd/gc/session_resolve.go`, `cmd/gc/session_name_lookup.go`, `cmd/gc/adoption_barrier.go` target and lifecycle helpers | target lookup, wake/drain/identity evidence, handoff metadata | Worker/session command or read-only projection helper | per touched slice | Broad row only; implementation inventory must split by symbol and exact keys. |
| W-027 | External calls to `Manager.CloseDetailed`, `Manager.Suspend`, and `Manager.UpdatePresentation` | close, lifecycle, identity presentation | Worker/session command factory or approved API adapter | 2, 4, repair slice | Guard fixture required for every non-test caller outside approved construction points. |
| W-028 | External calls to `session.WakeSession` and patch builders | wake/hold/drain, create/start, close, identity retirement | Session command APIs; patch builders stay internal implementation details | 2, 3, 4, 5 | Guard fixture required for package-level mutator calls and patch-map extension. |
| W-029 | `internal/api/handler_sessions.go` legacy session handlers | target resolution, repair, lifecycle, close, presentation | Typed API adapter over worker/session commands and read-only classifier selection | 1, 2, 4, repair slice | Broad row only; implementation inventory must split by endpoint, symbol, status/error shape, and exact keys. |
| W-030 | `internal/api/handler_beads.go` bead update/close/metadata routes that can target session beads | raw bead mutation, close, metadata batch, repair | Reject or route session-bead owned keys through session commands; prove non-session targets for generic paths | per touched slice | Unsafe by default when the target may be a session bead; guard fixture required for generic bridge writes. |
| W-031 | Runtime identity setter and clearer helpers, including chat/session-key backfills and continuation reset paths | runtime identity, create/start lease | Runtime-start repair command or audited repair helper with token proof | 3, repair slice | Must split setter, clearer, hash, and continuation-reset writes by symbol before source-complete status. |

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
- Flag external production calls to manager lifecycle/presentation methods
  (`CloseDetailed`, `Suspend`, `UpdatePresentation`) and package-level mutators
  (`WakeSession`, `RepairEmptyType`) unless the caller has an inventory row and
  an expiry tied to a slice.
- Inspect string-literal keys in map literals and helper calls. If the key is
  dynamic and the target bead might be a session bead, require an allowlist row
  with owner slice, reason, and retirement condition.
- Flag direct top-level session-bead type or status mutation through
  `beads.UpdateOpts`, `Close`, or `Create`; changing `type`, labels, status, or
  close metadata is a session-boundary mutation when the target may be a
  session bead.
- Treat generic bridge code as unsafe unless the code first proves the target
  bead is not a session bead or proves the written keys are outside the owned
  taxonomy.
- Require shrink-only allowlist behavior: an implementation slice may delete or
  narrow its own rows; adding a row needs a matching update to this design or
  `AGENTS.md` migration notes.

The guard must include positive fixtures for every known bypass shape named in
the inventory: raw store calls, manager calls, package-level mutators, dynamic
metadata batches, patch-map extension, generic bridge writes, top-level
type/status mutation, and API/CLI manager construction. Each fixture must fail
before the allowlist row is added and pass only when the row carries owner
slice, reason, and expiry.

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

Policy is adapter-owned post-filtering, not raw classifier behavior. The raw
classifier receives target tokens and already-gathered session/config facts; it
returns candidate facts in evidence order and never returns `selected`.
Compatibility adapters apply explicit policy fields: `surface`, `allow_closed`,
`materialize_named`, `allow_path_alias`, `allow_historical_alias`, `read_only`,
`allow_template_factory`, `allow_ordinary_config_target`, and
`reject_config_missing_named`. Policy can accept, reject, materialize, or fall
through; it cannot invent or re-rank a candidate outside that surface's
compatibility chain.

Repair and materialization are not classifier behavior. The raw classifier
collects candidate facts and may return `repair-needed` or
`requires-materialization`; it must not call `RepairEmptyType`, create a
session, retire an identity, or rewrite metadata. A writable policy adapter or
command may choose to repair or materialize only after it revalidates the same
target and records the audited command outcome. Mail query remains a recipient
set operation, not a single selected target, and historical aliases stay
read-only unless the surface-specific row above says otherwise.

Raw classifier result contract:

- ordered candidate facts and negative facts only
- no `selected` field
- no repair, materialization, close, update, or metadata write side effects
- no surface-policy rejection except facts such as `template-factory`,
  `ordinary-config-target`, `closed-session-name`, or
  `requires-materialization`

Adapter selection result contract:

- one `selected` candidate or one negative result; this field is the only
  authority for caller action
- ordered candidates considered by the active compatibility chain
- optional all-candidates diagnostic snapshot for ambiguity output; diagnostic
  candidates are not selection authority
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
| 0 Transition reducer baseline | `SESSION-STATE-001`, `SESSION-STATE-002`, `SESSION-STATE-003` | `internal/session/state_machine.go`, `internal/session/state_machine_test.go` | Any command slice that validates state transitions must add reducer parity or direct reducer-delegation tests before replacing local checks. | None without `REQUIREMENTS.md` update. | Missing reducer proof blocks command extraction even if higher-level API tests pass. |
| 1 Target classification | `SESSION-ID-003`, `SESSION-ID-004`, `SESSION-ID-005`, `SESSION-ID-007`, `SESSION-ID-008`, `SESSION-ID-009`, `SESSION-ID-010`, `SESSION-START-004` when materialization/suspend is touched, `SESSION-RUNTIME-005` where transcript lookup is touched | `internal/session/resolve_test.go`, `internal/session/named_config_test.go`, `internal/api/session_model_phase0_interface_spec_test.go`, `internal/api/session_model_phase0_lifecycle_spec_test.go`, `internal/api/session_resolution_path_alias_test.go`, `internal/api/handler_mail.go`, `internal/api/handler_extmsg.go` | Classifier candidate tests plus API, CLI, mail, extmsg, assignee, nudge, attach, inspect/log/transcript, reserved-named suspend, and materialization parity tests | None without `REQUIREMENTS.md` update | Tests must assert exact errors, result kinds, configured-name conflict, rejected-by-config, materialization, reserved-named suspend, path-alias tiebreakers, and no factory fallback. |
| 2 Explicit wake command | `SESSION-LIFE-007`, `SESSION-START-003`, `SESSION-START-007` | `internal/session/waits_test.go`, `internal/session/lifecycle_transition_test.go`, `internal/api/session_model_phase0_lifecycle_spec_test.go` | Command/applier conflict, stale-fact, no-template fallback, wait-cancellation, and API/CLI response tests | None except documented bug fix | Proof must include terminal closed, archived, suspended, drained, no-template, pending-create, and wait cancellation cases. |
| 3 Runtime start prepare/commit/rollback | `SESSION-START-001`, `SESSION-START-002`, `SESSION-START-008`, `SESSION-LIFE-004`, `SESSION-LIFE-005`, `SESSION-RUNTIME-001`, `SESSION-RUNTIME-003` | `internal/session/lifecycle_transition_test.go`, `internal/session/manager_test.go`, `cmd/gc/session_lifecycle_parallel_test.go`, `cmd/gc/session_reconcile_test.go`, `internal/session/submit_test.go` | Runtime-start command tests for provider-start success followed by commit failure, crash after prepare, stale generation, stale/mismatched instance token, and partial metadata repair | None | Runtime-start keys have one owner; pending-create clear, runtime identity, config hashes, prepare, commit, and rollback cannot be split across slices. |
| 4 Close and identity retirement | `SESSION-START-005`, `SESSION-START-006`, `SESSION-WORK-001`, `SESSION-WORK-002`, `SESSION-ID-006`, `SESSION-ID-007` | `internal/session/manager_test.go`, `internal/api/session_model_phase0_lifecycle_spec_test.go`, `cmd/gc/session_beads_test.go`, commits cited in `REQUIREMENTS.md` | Close command tests, idempotent work-release scan tests, API/CLI error parity, event-independent recovery tests | None without owner approval | Must prove provider-stop failure leaves bead open and successful close releases assigned work without relying solely on events. |
| 5 Wake/hold/drain eligibility and commands | `SESSION-LIFE-003`, `SESSION-LIFE-004`, `SESSION-START-003`, `SESSION-RECON-004`, `SESSION-WORK-004` | `cmd/gc/session_reconcile_test.go`, `internal/session/lifecycle_projection_test.go`, `internal/session/waits_test.go`, `cmd/gc/session_wake_test.go` | Pure decider tests for unknown/partial/stale facts plus drain-cancel and drain-ack recovery parity | None | Destructive drain/close decisions must fail closed on partial work or runtime facts. |
| 6 Reconciler fact extraction | `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-005`, `SESSION-START-008` | `cmd/gc/build_desired_state_test.go`, `cmd/gc/session_lifecycle_parallel_test.go`; `cmd/gc/scale_from_zero_test.go` is cited by requirements but absent in this checkout | Restore or replace scale-from-zero proof before extraction, then add fact-reader complexity tests | None | Missing proof blocks this slice; controller scheduling policy stays controller-owned. |
| 7 Provider health, progress, and circuit state | `SESSION-RECON-006`, `SESSION-RECON-007`, `SESSION-RUNTIME-001`, `SESSION-RUNTIME-002` | `cmd/gc/session_reconciler_test.go`, `internal/session/manager_test.go`; `cmd/gc/provider_health_gate_test.go` and `cmd/gc/session_progress_test.go` are cited by requirements but absent in this checkout | Restore or replace provider-health/progress proof, then add stale runtime fact, alert-dedup, and budget tests | None | Missing proof blocks this slice; health/progress/circuit side effects remain reconciler-owned until a narrower command is proven. |

Citation freshness gate:

- Add a test or script that fails when a path cited in this matrix or
  `REQUIREMENTS.md` does not exist and is not an issue or commit reference.
- Treat commit citations as historical support, not current proof, unless the
  cited behavior is also asserted by a test or source path that exists in the
  checkout. Non-ancestor or unreachable commits cannot unblock extraction by
  themselves.
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

Attempt 5 closes two remaining loopholes:

- A plain read followed by `SetMetadata*`, `Update`, or `Close` is not a
  valid command implementation unless the slice proves a store-level
  conditional revision check or wraps the write in tokened phase markers and
  repair. Post-write "success" from stale facts must be detectable and
  classified as committed, repaired, compensated, or rejected.
- Legacy no-token runtime identity backfills, `CloseDetailed` versus
  lower-level close helpers, and provider-start success followed by commit
  failure are repair scenarios, not alternate command paths. They need explicit
  stale-success, orphan-runtime, stop-compensation, and retry tests before the
  runtime-start or close slice can delegate callers.

Every command cluster must define a command-applier row before implementation:

| Field | Required content |
|---|---|
| snapshot preconditions | Exact lifecycle, identity, runtime, config, work, and policy facts read before the command decides. |
| validation point | The re-read, token comparison, revision check, or phase marker verified immediately before mutation. |
| allowed write primitives | `Update`, `Close`, `SetMetadataBatch`, backend transaction, or command-specific helper, with store coverage proof. |
| token/revision/phase marker | `instance_token`, generation, drain generation, close intent, or other durable fence that identifies the attempt. |
| conflict reasons | Typed stale, terminal, missing-config, identity-conflict, provider-stop-failed, partial-query, budget, or runtime-mismatch outcomes. |
| stale-success handling | How a write that used stale facts is detected and classified as committed, compensated, repaired, rejected, or unsafe. |
| partial-state matrix | Every visible partial subset from multi-key writes and the repair or fail-closed behavior for that subset. |
| provider side-effect order | Whether the durable intent is written before the provider action and how success/failure is compensated. |
| repair authority | Which controller, command, or doctor path may repair and which facts it must prove first. |
| post-commit fact and event order | Facts returned or emitted only after durable commit, plus idempotency and supersession keys. |
| race tests | Required tests for close-vs-wake, close-vs-runtime-start, drain-vs-assigned-work, stale token, duplicate command, partial store failure, event miss, and crash after commit. |

Runtime-start and close require durable fences around provider side effects:

- Runtime-start writes a prepare token before provider start. If provider start
  succeeds and commit fails, repair either commits the same token from observed
  runtime identity or stops/orphan-handles that runtime by token. A newer token
  prevents old provider identity from mutating the bead.
- Close writes or proves a close intent before irreversible cleanup when the
  provider action can outlive the process. If provider stop succeeds and durable
  close fails, repair retries close without restarting the stopped runtime. If
  provider stop fails, the bead remains open and no close/retire facts are
  emitted.

| Command cluster | Precondition fields | Written keys | Ordering/atomicity assumption | Conflict reasons | Retryability and repair proof |
|---|---|---|---|---|---|
| Target classification | none; read-only snapshot; adapter-owned policy applies after candidate collection | none | no mutation | ambiguous, not-found, forbidden-kind | Retry after config/store change. |
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

Durable scans are the mandatory backstop for every critical retryable reaction.
Synchronous cascades may remain only as latency optimizations and must be safe
to skip, duplicate, or crash after commit. Each reaction row must name the scan
owner, cadence, store refs queried, completeness guard, idempotency key,
supersession key, partial-query behavior, and event-miss/crash-after-commit
tests.

Internal post-commit facts use this minimum contract before any public payload
change:

| Field | Required content |
|---|---|
| fact type | Existing `events.*` constant or slice-local internal fact name. |
| committed identity | Session bead ID plus current session name, alias, configured identity, generation, and `instance_token` when relevant. |
| idempotency key | Stable key for duplicate subscriber scans, usually operation + bead ID + generation/token or drain generation. |
| supersession key | Newer token/generation, terminal status, or successor canonical identity that makes an older fact non-authoritative. |
| scan authority | Controller/reconciler/adapter path that can converge from durable facts when the event is missed. |
| diagnostics | Trace/doctor outcome for converged, deferred, blocked, duplicate, stale, and failed scan results. |

For close/work-release, the controller owns the durable scan on the same
controller tick or the next reconciler scan. The scan queries closed session
facts and open/in-progress work by bead ID, current and historical names where
requirements allow, and configured identity. A partial work query defers
release unless the synchronous close path already proved complete; a live
successor or newer token suppresses old-identity release.

| Operation | Required order | Recovery authority | Subscriber class |
|---|---|---|---|
| Explicit wake request | validate -> commit wake metadata -> optional event/trace -> reconciler scan starts runtime | Durable wake metadata and controller scan | critical retryable |
| Runtime start | commit creating/generation/token -> start provider -> commit active/session key -> event/trace | Pending-create metadata, generation, provider runtime identity | critical retryable |
| Close | provider stop success when required -> commit close/retire facts -> release work scan -> event/trace | Closed bead and session metadata; work-release scan | critical retryable |
| Drain | commit drain intent -> worker ack -> reread assigned work -> complete/cancel | Drain metadata plus live work query | critical retryable |
| Identity retirement | close/retire metadata -> assignment and binding scans -> event/trace | Retired identity metadata and open work/binding queries | critical retryable |
| Trace/SSE/dashboard | emit after commit | Event log and trace store when available | observability-only |

Event contract:

- Current events are thin hints unless an event-migration slice explicitly
  changes typed payloads. The fields below are required durable fact identities
  for scans; they become wire payload fields only when the slice updates typed
  payloads, registry entries, OpenAPI/SSE projections, generated clients, and
  rendering tests.
- When payloads are added, they describe facts that happened, not commands to
  attempt, and carry stable IDs, generation/instance token when relevant, and
  idempotency keys for subscriber scans.
- Existing `events.*` constants remain the first mapping target. A new
  session event requires `events.RegisterPayload`, OpenAPI/SSE projection
  updates, and `TestEveryKnownEventTypeHasRegisteredPayload` parity.
- Duplicate, skipped, out-of-order, and crash-after-commit delivery must leave
  critical state recoverable from durable facts.

Per-event reaction matrix:

| Event/fact | Tier | Required durable fact or payload fields | Idempotency/supersession | Durable recovery authority | Tests |
|---|---|---|---|---|---|
| wake requested / `events.SessionWoke` | accelerator + observability | session ID, target identity, generation/token if start is prepared, wake cause | Superseded by newer generation/token or terminal bead status | wake metadata plus controller desired-state scan | durable metadata still starts or remains queued when event emission is skipped; stale event ignored |
| runtime stopped / `events.SessionStopped` | critical diagnostic, recovery accelerated by scan | session ID, runtime key, stop reason, generation/token when known | Superseded by newer live runtime identity or closed bead | provider observation plus session bead state | event miss does not strand assigned work |
| runtime crashed / `events.SessionCrashed` | critical diagnostic, recovery accelerated by scan | session ID, runtime key, crash reason, captured output ref, generation/token | Superseded by newer token or repaired state | provider observation, crash metadata, restart budget state | duplicate crash does not double-consume budget |
| drain state / `events.SessionDraining`, `events.SessionUndrained` | accelerator + observability | session ID, drain generation, reason | Superseded by drain generation or terminal state | drain metadata and assigned-work query | durable facts still complete or cancel by scan when event emission is skipped |
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

The extraction order is a three-stage pipeline:

1. The reconciler assembles demand, pool, work, config, runtime, health,
   progress, budget, and trace facts from live sources.
2. `internal/session` receives only immutable lifecycle and identity facts and
   returns an eligibility mask: runnable, blocked with reason, terminal,
   repair-needed, or unknown/partial.
3. The reconciler combines that mask with controller-owned demand, capacity,
   provider-health, progress, circuit, restart-budget, and alert-dedupe policy
   before starting, draining, or suppressing sessions.

Pure session deciders must live in a guardable file set with no imports of
bead stores, config loaders, runtime providers, event recorders, clocks, files,
work queries, provider-health readers, or controller demand helpers. Adapters
may gather those facts; deciders may only consume the already-materialized
structs.

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

Runtime observation completeness is explicit. A destructive branch must record
which observation state it consumed and why the branch is safe:

| Observation state | Meaning | Destructive branch rule |
|---|---|---|
| complete-alive | Provider/runtime facts identify the current token or runtime key as alive. | Do not close, release, or clean up as dead; restart policy may nudge or observe only. |
| complete-missing | Complete provider/runtime query proves the current token or runtime key is absent. | Close/release/cleanup may proceed only after work/config facts are also complete and command preconditions pass. |
| stale-observation | Query succeeded but is older than the slice's freshness bound or predates the current token/generation. | Treat as unknown; no destructive close, drain, rollback, release, cleanup, or restart. |
| unknown-provider | Provider lacks the fact, probe is inconclusive, or provider reports no process names. | Preserve existing fail-open or defer behavior for health, but fail closed for destructive branches. |
| provider-error | Provider query failed or timed out. | Defer destructive action, emit trace/doctor evidence, and retry through controller scan. |
| partial-query | Session, work, config, or runtime facts are incomplete. | Suppress destructive close/drain/rollback/release/cleanup/restart and return `durable-scan-deferred` or typed conflict. |
| successor-observed | A newer token/generation or canonical identity exists. | Old token cannot mutate the bead; cleanup is best-effort by runtime identity and cannot release successor-owned work. |

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

Every slice must also carry a per-key owner matrix in its bead or task
artifact. The matrix is the rollback contract; without it, the slice cannot be
marked done.

| Key family | One writer during bake | Required rollback rule |
|---|---|---|
| Create/start lease and runtime identity | Runtime-start command slice only after W-005/W-012/W-022 delegation begins. | Revert prepare, commit, rollback, and repair together; never leave old prepare with new commit or vice versa. |
| Wake/hold/drain | Wake/hold/drain command slice after W-004/W-006/W-007/W-023 delegation begins. | Revert the cluster as a unit if wait cancellation, no-template wake, or drain generation behavior changes. |
| Close and identity retirement | Close/retire command slice once provider-stop and work-release recovery tests pass. | Restore the old close caller and work-release scan together; provider-stop failure must continue leaving the bead open. |
| Target repair and materialization | Policy adapter plus audited repair/materialization command, not the raw classifier. | Revert to the previous resolver path for that surface; no metadata migration is allowed in the classifier-only slice. |
| Diagnostics and event payloads | Central trace/event vocabulary owner for the slice. | Remove new codes or payloads with their renderers and generated clients; never leave undocumented API/SSE fields. |
| Reconciler demand, health, progress, and budgets | Controller/reconciler until a requirement explicitly moves a rule. | Revert to reconciler-owned helpers; `internal/session` must not retain scheduling policy after rollback. |

Per-slice coexistence gates:

| Slice | Converted callers | Legacy callers allowed during bake | Validation differences allowed | Guard update | Bake and revert rule |
|---|---|---|---|---|---|
| 1 Target classification | API target resolver adapter first; then mail/extmsg/CLI/assignee helpers one surface at a time | Existing resolver functions stay as oracle until parity tests pass | None; exact errors and materialization behavior preserved | Read-only guard prevents new resolver copies after adoption | Revert by switching adapter back to old resolver; no metadata migration. |
| 2 Explicit wake | `cmd/gc/cmd_session_wake.go`, API wake path, wait/no-template fallback | Reconciler implicit wake and runtime-start code untouched | None except approved bug rows | Retire W-004 exception when converted | Bake requires trace/API/CLI parity and close-vs-wake race tests. |
| 3 Runtime start | `cmd/gc/session_lifecycle_parallel.go` prepare/commit/rollback, worker create path where applicable | No parallel writer for `pending_create_*`, `instance_token`, `session_key`, or config hash keys after conversion begins | None without requirement update | Retire W-005 and W-012 together | Revert only as whole runtime-start slice; do not split prepare/commit/rollback ownership. |
| 4 Close/retire | worker close, API close/retire, session close sweep | Work-release scan may stay outside session but must consume committed close facts | None; provider-stop failure still leaves bead open | Retire W-008 close-mutation exception | Bake requires duplicate close, skipped-event-emission, and work-release recovery tests. |
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
| durable-scan-converged | durable fact scan repaired or completed state without relying on event delivery | trace and doctor evidence |
| durable-scan-deferred | durable fact scan found incomplete facts and safely deferred mutation | trace, doctor, and retry evidence |
| repair-applied | audited repair wrote normalized session state with before/after evidence | trace, doctor, logs |
| repair-skipped | repair was unsafe, stale, partial, or unnecessary | trace, doctor, logs |

Per-operation diagnostic mapping:

| Operation | Accepted site/reason | Rejected/conflict reasons | Required rendering tests |
|---|---|---|---|
| Target classification | `session.target.resolve` / selected candidate kind | not-found, ambiguous, forbidden-kind, configured-name-conflict, rejected-by-config, closed-not-allowed | API/Huma error body, CLI stderr/JSON, mail/extmsg logs, `gc trace` candidate rendering |
| Explicit wake | `session.wake.request` / wake cause | terminal, missing-config, identity-conflict, pending-create-in-flight, stale-state | API status/body, CLI exit/text, trace, no-template wake |
| Runtime start prepare/commit/rollback | `session.runtime_start.prepare`, `.commit`, `.rollback` | stale-generation, stale-instance-token, runtime-mismatch, partial-start, provider-start-failed | trace with token/generation, provider error logs, doctor partial-start output |
| Close/retire | `session.close.commit`, `session.identity.retire` | provider-stop-failed, already-closed, stale-state, duplicate-canonical, work-release-deferred | CLI/API close output, trace, doctor close/work-release evidence |
| Drain complete/cancel | `session.drain.complete`, `session.drain.cancel` | assigned-work-present, partial-work-query, stale-drain-generation | trace, `SessionDrainAckedWithAssignedWork` payload, doctor assigned-work output |
| Repair path | `session.repair.apply` | repair-unsafe, partial-query, unknown-key, non-session-bead | doctor and trace output proving each direct write |

Doctor or session-inspect checks are concrete surfaces, not optional prose:

| Check name | Statuses | Data sources | Fixability and rendering proof |
|---|---|---|---|
| `session.target-resolution` | ok, ambiguous, rejected, repair-needed, materialization-required | classifier candidates, policy input, config snapshot | API/CLI parity and trace candidate rendering. |
| `session.command-precondition` | ok, blocked, stale, partial, retryable | command precondition facts and current bead projection | CLI/API conflict rendering with retryability. |
| `session.runtime-start` | prepared, committed, partial, orphan-runtime, rollback-needed | pending-create keys, `instance_token`, runtime observation | doctor partial-start output and trace token rendering. |
| `session.close-work-release` | converged, deferred, blocked, failed | closed bead facts plus assigned-work scan | work-release recovery rendering and duplicate-scan proof. |
| `session.drain-assigned-work` | complete, canceled, deferred, stale-generation | drain metadata plus live assigned-work query | `gc trace` and doctor assigned-work output. |
| `session.repair` | applied, skipped, unsafe, partial, failed | target bead proof, owned-key set, before/after values | doctor and trace output for every direct repair write. |

Trace mapping rule:

- Use existing centralized trace site/reason/outcome codes in
  `cmd/gc/session_reconciler_trace_types.go` when available.
- Add new codes only in the centralized trace vocabulary with tests for
  `gc trace` rendering.
- API-visible conflict fields must be represented by typed Huma outputs or
  errors; do not smuggle structured data through raw JSON maps.

Executable diagnostics gates:

| Artifact or command | Required proof |
|---|---|
| `cmd/gc/session_reconciler_trace_types.go` plus renderer tests | New site/reason/outcome codes are centralized and rendered by `gc trace`. |
| `go test ./cmd/gc -run 'TestSessionReconcilerTrace|TestSessionDoctor|TestSessionInspect' -count=1` | Trace, doctor, and inspect render accepted, rejected, blocked, deferred, repair, and scan outcomes for touched operations. |
| `go test ./internal/api -run 'Test.*Session.*(Command|Lifecycle|Interface|Error)' -count=1` | API/Huma status, error body, request ID behavior, and typed response shape stay compatible for touched surfaces. |
| `go test ./cmd/gc -run 'Test.*Session.*(CLI|JSON|Wake|Close|Drain|Resolve)' -count=1` | CLI stdout, stderr, JSON shape, and exit codes stay compatible for touched commands. |
| `go test ./internal/events -run TestEveryKnownEventTypeHasRegisteredPayload -count=1` | Any new public event payload is registered and typed. |

If a named test does not exist for a touched surface, the slice must add it
before implementation or mark the surface out-of-scope in the parity source.

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
- Default cap before a slice proves a different number: a command may perform
  at most one synchronous critical subscriber pass, bounded to 100 affected
  items or 250 ms of wall time before deferring to controller/reconciler repair.
- Record fact-build and subscriber durations in trace when a slice touches the
  reconciler hot path.
- Large-city target: fact materialization is O(session beads + relevant work
  beads + configured agents) per tick unless the current implementation already
  has a lower bound that must be preserved.
- If no historical benchmark exists, the slice must add a baseline test first
  and then prove the changed code performs no more than baseline plus one
  store query on the same fixture.

Slice-level budgets:

| Path | Budget |
|---|---|
| Package target resolution | Exact ID lookup plus at most two indexed metadata list calls for `session_name` and `alias`; no all-session scan in the hot path. |
| API configured named resolution | At most two bounded metadata-filtered list calls for configured identity; no duplicate `session_name` full scan. |
| API path alias fallback | At most one all-session scan after prior exact/config/live resolution fails; state filter and most-recent-created tiebreaker happen in memory. |
| Mail recipient resolution | No session materialization for send/query; configured mailbox lookup uses bounded identity lists and no more than baseline plus one store query. |
| Extmsg member notification | Materialization is only on member miss and is bounded by membership count; one member failure logs and does not abort unrelated members. |
| Runtime-start repair scan | Reuses reconciler session snapshot; zero new provider subprocess fan-out per session beyond existing runtime observation. |
| Close/work-release scan | List assigned open/in-progress work by assignee and store ref; no whole-city work scan when indexed query is available; duplicate scan is idempotent. |
| Event subscribers | At most one synchronous critical subscriber pass per command, capped at 100 affected items or 250 ms before deferred repair. |

Every slice touching a hot path must include either a query-count test, a
benchmark, or an explicit source proof that the budget is preserved.

Executable performance gates:

| Gate | Command or fixture |
|---|---|
| lookup query count | `go test ./internal/session ./internal/api -run 'Test.*(QueryCount|TargetResolutionBudget)' -count=1` |
| reconciler fact materialization | `go test ./cmd/gc -run 'Test.*(DesiredState|Reconciler).*QueryCount' -count=1` |
| close/work-release scan | `go test ./cmd/gc -run 'Test.*(WorkRelease|AssignedWork).*QueryCount' -count=1` |
| runtime-start repair scan | `go test ./cmd/gc ./internal/session -run 'Test.*RuntimeStart.*(Repair|Recovery)' -count=1` |
| subscriber cap and backpressure | `go test ./cmd/gc ./internal/events -run 'Test.*(Subscriber|FanOut|Backpressure)' -count=1` |
| large-city baseline | Slice-owned benchmark or fixture named in the bead, with before/after results and allowed delta. |

If a gate command matches no tests, that is not a pass. The implementation bead
must add the missing query-count, benchmark, or source-proof test before moving
hot-path behavior.

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
- Add raw classifier result types plus adapter-owned policy and selection
  result types only; do not add a broad session facade.
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
- New idempotent recovery tests for skipped event emission and duplicate subscriber scan

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
