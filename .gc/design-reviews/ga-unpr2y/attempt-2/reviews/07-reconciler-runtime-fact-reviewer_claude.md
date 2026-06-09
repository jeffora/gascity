# Liam Okonkwo - Claude

**Verdict:** approve-with-risks

**Top strengths:**

- The Reconciler Fact Contract anchors on surfaces that are genuinely pure and
  adapter-fed today. `ComputeAwakeSet` already consumes a precomputed
  `AwakeInput` ("All external I/O ... happens before this function is called",
  `cmd/gc/compute_awake_set.go:18-36`), the fact reader already lives in the
  adapter (`buildAwakeInputFromReconciler`,
  `cmd/gc/compute_awake_bridge.go:17`), and `ProjectLifecycle` consumes
  three-state runtime facts (`RuntimeFacts.Observed`,
  `internal/session/lifecycle_projection.go:143-148`). Baseline-first with "do
  not replace those with broader abstractions until a narrow slice proves
  parity" is the right reconciler-migration posture, and it directly answers
  the fact-isolation question: work counts, pool size, runtime liveness, and
  wait/attachment facts are precomputed and passed in, not queried from inside
  deciders.
- Runtime observations stay non-durable throughout. "Stale/unknown runtime
  cannot prove death for destructive action", partial work queries cancel
  no-wake drains, and unknown budget state blocks destructive restarts — all
  fail-closed for destruction, and all grounded in mechanisms that exist
  (`runtime.IsPartialListError` at `cmd/gc/session_beads.go:1861`,
  partial-template tracking in `cmd/gc/build_desired_state.go:280-312`,
  `events.SessionWorkQueryFailed`). The red flag "runtime missing observations
  become durable session truth" does not materialize in this design.
- The slice 5/6 freshness gates honestly self-report absent proof and block
  extraction on it ("Missing proof blocks this slice"), instead of letting the
  reconciler slices proceed on stale citations.

**Critical risks:**

- [Major] **Slice 6's baseline behavior is absent from this branch, not just
  its proof.** The provider-health gate (`b5a7f3be3`) and progress-aware
  health predicate (`dbda1e380`) are not ancestors of this checkout
  (`git merge-base --is-ancestor` fails for both), and no provider-health
  gating code exists in `cmd/gc` production source — only the separate
  no-progress respawn circuit breaker (`cmd/gc/session_circuit_breaker.go`).
  Yet the fact table rows for "Provider health" and "Progress signatures",
  desired-running precedence step 5 ("provider-health red suppresses
  respawn"), and the runtime-liveness row's "may preserve existing fail-open
  health behavior where required" all describe this as current behavior with a
  current owner. Slice 6's gate is framed as "Restore or replace the missing
  ... proofs" — restoring the *test files* alone would produce failing tests
  against a missing implementation. The design must say the implementation
  itself needs restoring (or the REQUIREMENTS rows need re-scoping), and name
  the restoration source.
- [Major] **The provider-health/progress decision split is given three
  inconsistent answers.** (a) The fact table owner column assigns provider
  health, progress signatures, and budgets to "reconciler". (b) Slice 6 says
  "Feed provider/progress facts into session-owned health decisions, while
  the reconciler keeps scheduling, budgets, and trace output". (c) The
  desired-running precedence places provider-health red at step 5 of *desire*
  computation while describing it as respawn suppression — an action-level
  gate. These do not compose into one boundary. Note the two gates are
  different in kind: SESSION-RECON-007 progress is desire-level ("is not
  desired and can be drained") while SESSION-RECON-006 health is action-level
  (skip respawn, no budget consumption, identity not terminal); the design
  lumps them into one slice with one sentence. "One alert per red episode"
  additionally depends on episode state whose fact shape is undefined — if
  alert/dedup decisions move into a session decider, episode facts must be
  defined; if they stay in the reconciler, the fact-table row should say so.
  This is the lane's central question and the current text would let two
  implementers draw the boundary in different places.
- [Major] **The desired-running precedence list is an incomplete and
  ambiguous restatement of the baseline it declares normative.** The design
  says "Desired-running precedence must remain explicit", then omits:
  wait-hold demand suppression vs ready-wait wake
  (`cmd/gc/compute_awake_set.go:326-353` — `WaitHold` suppresses
  config/attached/pending/pinned wake but `ReadyWaitSet` still wakes),
  idle-sleep suppression and its exemptions (attached/pending/pinned/
  always-named/assigned-work/min-active), and min-active demand. Worse, the
  word "holds" conflates two mechanisms with *opposite* wait interactions:
  `held_until`/quarantine suppress everything last — including wait-ready
  wake (`cmd/gc/compute_awake_set.go:405-414`) — while SESSION-RECON-004's
  "holds suppress ... except wait-only wake" describes the `wait_hold`
  mechanism. An implementer building the slice 4 decider from this list plus
  RECON-004 could make `held_until` overridable by ready-wait, inverting
  current behavior. Parity tests against `ComputeAwakeSet` would catch it,
  but the design should not be the document that introduces the inversion.
- [Major] **Decider-relocation mechanics are unstated and invite the
  forbidden import direction.** `AwakeInput` and the pool desired-state types
  live in package `main`; `computePoolDesiredStates` threads a
  `*sessionReconcilerTraceCycle` through the decision path and emits trace
  records inline during cap rejection (`cmd/gc/pool_desired_state.go:72-90`,
  `applyNestedCaps`). A session decider in `internal/session` can import
  neither, and the operability contract's instruction to "use existing
  centralized trace site/reason/outcome codes in
  `cmd/gc/session_reconciler_trace_types.go`" is unsatisfiable from
  `internal/session`. The implied resolution — deciders return structured
  decision evidence; the cmd/gc adapter maps it to trace codes — is correct
  but appears nowhere, and slices 4/5 "Before implementation" lists do not
  mention moving or parity-redefining the package-main fact types or
  restructuring inline trace emission into decision outputs. This is exactly
  how "session deciders import reconciler helper types" (or trace vocabulary
  migrating wholesale into `internal/session`) happens under deadline.
- [Minor] **RuntimeIntent has no contract.** The term appears in the target
  shape and TR-005 but no such type exists anywhere in the tree, and the
  vocabulary section defers definition to later slices. TR-005's rule
  ("session may request runtime intents but does not embed provider-specific
  execution policy") is right but untestable without one constraint the
  design never states: runtime identity values (provider session keys,
  transport, command hash/live hash — listed as session-owned *mutation* in
  the boundary table) are computed by runtime adapters and passed to session
  commands as opaque facts; session stores and compares them but never
  derives them. Without that sentence, the commit-runtime-start command
  ("written keys ... session key/hash fields") reads as license for
  `internal/session` to compute provider hashes.
- [Minor] **Two opposite unknown-fact defaults coexist without a stated
  rule.** Unknown runtime liveness fails closed for destruction; unknown
  provider health fails open for respawn gating. Both are individually
  correct and required, but the runtime-liveness row cross-references the
  health fail-open behavior, smudging the line. One sentence — "unknown facts
  fail closed for destructive actions and fail open for availability gates" —
  would make the pattern auditable per fact family.

**Missing evidence:**

- `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`,
  and `cmd/gc/session_progress_test.go` are absent from this checkout
  (design acknowledges).
- The provider-health gate and progress-predicate *implementations* are also
  absent — commits `b5a7f3be3` and `dbda1e380` are not ancestors of this
  branch (design does not acknowledge; it describes the behavior as current).
- No `RuntimeIntent` type or consumer exists in the tree; the contract is
  design-prose only.
- No fact shape exists for provider-health episode state ("one alert per red
  episode" dedup) or for in-memory budget state crossing into decider facts.
- The design does not name which branch/commit serves as the restoration
  source for slice 6's behavior.

**Required changes:**

- Re-scope slice 6's gate from "restore the missing proofs" to "restore the
  provider-health and progress implementations plus their proofs, or re-scope
  the SESSION-RECON-006/007 rows", and mark the affected fact-table rows and
  precedence step 5 as describing restored-target behavior, citing the source
  commits.
- Add a slice 6 decision-ownership table mirroring the fact table: for each
  decision (respawn suppression, red-episode alert emission, progress-stall
  drain, exemption evaluation, budget consumption) name decider location
  (session decider vs reconciler scheduling) and fact inputs. Resolve whether
  provider-health red is an input to the session desired-state decider or a
  post-decider reconciler gate, and state that progress is desire-level while
  health is action-level if that is the intended split.
- Complete the desired-running precedence list (wait-hold suppression,
  ready-wait wake, idle-sleep suppression and exemptions, min-active demand)
  or explicitly mark `ComputeAwakeSet` as the normative precedence and the
  list as a summary. Disambiguate "holds": `held_until`/quarantine override
  everything including wait-ready; `wait_hold` suppresses demand wake but
  yields to ready-wait.
- State the decider-relocation rule: fact and decision types move (or are
  parity-redefined) into `internal/session`; deciders return structured
  decision evidence; trace-code mapping stays in the cmd/gc adapter; the
  trace vocabulary does not migrate into `internal/session`. Add the inline
  trace-emission restructuring to slice 5's "Before implementation" list.
- Add the runtime-identity opacity rule to TR-005 or the mutation boundary:
  adapter-computed, session-stored, never session-derived.

**Questions:**

- Is the slice 6 restoration source a cherry-pick of `b5a7f3be3`/`dbda1e380`,
  or a fresh implementation written against restored tests and the
  REQUIREMENTS rows?
- When slice 4 moves wake eligibility into `internal/session`, does
  `AwakeInput` move verbatim or get redefined? Specifically, does agent
  base-name resolution with duplicate detection
  (`cmd/gc/compute_awake_set.go:103-127`) — config-name policy — belong inside
  the session decider, or should the adapter resolve names and pass resolved
  references as facts?
- Where does red-episode/alert-dedup state live after extraction: durable
  session metadata, durable reconciler-owned state, or in-memory reconciler
  state passed in as facts with an explicit unknown-handling rule?
- Does "session-owned health decisions" in slice 6 mean the decider emits a
  drain/respawn-suppression *decision*, with the reconciler retaining the
  right to defer it for budget/scheduling reasons — and if so, which side's
  verdict is recorded in trace as authoritative?
