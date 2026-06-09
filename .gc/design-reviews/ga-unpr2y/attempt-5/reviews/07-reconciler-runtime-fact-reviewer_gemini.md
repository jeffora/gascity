# Liam Okonkwo — DeepSeek V4 Flash (Independent Review, Attempt 5)

**Verdict:** block

**Persona:** reconciler boundary, runtime intent adapter ownership, fact isolation, health gate split.

**Reviewed against:** `internal/session/DESIGN.md` (attempt 5, `.gc/design-reviews/ga-unpr2y/attempt-5/design-before.md`), `cmd/gc/compute_awake_set.go`, `cmd/gc/compute_awake_bridge.go`, `cmd/gc/build_desired_state.go`, `cmd/gc/pool_desired_state.go`, `cmd/gc/session_circuit_breaker.go`, `cmd/gc/session_reconciler.go`, `internal/session/lifecycle_projection.go`, and attempt-1 and attempt-2 reviews (all reconciler-runtime-fact-reviewer lanes).

---

## Top Strengths

- **Clear Separation of Scheduling/Capacity Logic**: Sticking strictly to the core principle that "reconciler keeps scheduling, budgets, and trace output" (`DESIGN.md:200`, `701-705`) prevents the session package from being bloated with complex pool-scaling, work-aggregation, and multi-tier queue-shaping algorithms.
- **Explicit Pure Decider Fact Invariant**: Prohibiting pure deciders from importing store, runtime, or provider-health components (`DESIGN.md:205-207`) enforces strict design boundaries. This isolation ensures that `internal/session` deciders can be completely verified using mock-free, high-fidelity table-driven unit tests.
- **Robust Command Atomicity and Stale Fact Defenses**: Requiring commands to re-read and validate preconditions immediately before a transaction/commit (`DESIGN.md:601-614`) is a stellar strategy against cross-process concurrency races across CLI, API, controller, and reconciler processes.
- **Grandfathering Exception Model**: Pragmatically listing W-009 through W-013 exceptions (`DESIGN.md:437-441`) provides an achievable migration baseline, acknowledging live code reality rather than assuming a single, risky "flag day" sweep.

---

## Critical Risks (Blockers)

### [Blocker] 1. `AwakeInput` Composition Struct Still Fuses Disjoint Ownership Domains

The design's state-ownership table explicitly assigns "pool demand, work aggregation, scale checks, nested caps" to "Controller/reconciler policy; not a session primitive" (`DESIGN.md:200-202`). However, `DESIGN.md:352` (Wake selection) anchors on `AwakeInput` for slice 5 without establishing an isolated input structure.

In the codebase, `AwakeInput` (`cmd/gc/compute_awake_set.go:21-36`) mixes session-lifecycle facts with controller-domain demand inputs:
1. **Session eligibility facts**: `SessionBeads`, `ReadyWaitSet`, `RunningSessions`, `AttachedSessions`, `PendingSessions`, `ChatIdleTimeout`, `Now`.
2. **Controller demand facts**: `ScaleCheckCounts`, `NamedSessionDemand`, `NamedSessionWorkQ`, `WorkSet`, `WorkBeads`, `Agents.MinActiveSessions`.

If the entire `AwakeInput` struct is imported or copied into `internal/session`, then reconciler vocabulary such as `ScaleCheckCounts` and `WorkBeads` will cross the package boundary. This violates Liam's core red flag: *session deciders must not import reconciler helper types or bead store queries.*

* **Required Change**: Explicitly define an `EligibilityInput` (or equivalent) in `internal/session` carrying only session-lifecycle facts (state, holds, quarantine, explicit wake). Document that reconciler-specific fields (such as `ScaleCheckCounts`, `WorkBeads`, and `MinActiveSessions`) remain reconciler-owned demand facts and do not move into `internal/session`. Detail how the bridge layer (`cmd/gc/compute_awake_bridge.go`) builds both halves from raw reconciler data.

---

### [Blocker] 2. Bidirectional Eligibility ↔ Demand Co-dependence Lacks Seam Contract

`ComputeAwakeSet` (`cmd/gc/compute_awake_set.go`) does not compute eligibility and demand in separate, sequential passes. Instead, the code demonstrates bidirectional co-dependencies:
- **Demand passes are gated by eligibility**: `collectActiveBeads` (which populates the scaled:demand and work-query slots) excludes drained, closed, dependency-only, and manual beads (`compute_awake_set.go:572-589`).
- **Eligibility consumes demand results**: The idle-sleep exemption checks `desired[name] != "assigned-work" && desired[name] != "min-active"` (`compute_awake_set.go:386-388`), reading the demand pass's output.

The design's desired-running precedence lists five ordered steps (`DESIGN.md:731-739`) but presents them as one flat fold. It fails to specify how to split this co-dependence across a slice-5 session decider and a slice-6 demand adapter without creating tight coupling.

* **Required Change**: Define an explicit multi-pass execution contract to handle this co-dependence cleanly:
  1. **Pass 1 (Session Eligibility Decider)**: Emits a pure eligibility mask per bead (terminal, held, quarantined, drained, or eligible-for-wake).
  2. **Pass 2 (Reconciler Demand Adapter)**: Computes scaling and work demand only against eligible beads from Pass 1.
  3. **Pass 3 (Session Override Decider)**: Evaluates idle-sleep, pins, and attachments given both the Pass 1 eligibility mask and the Pass 2 demand results.
  State clearly that Pass 3 owns the idle-sleep exemption decision.

---

### [Blocker] 3. Desired-Running Precedence Table Collapses Execution Layers, Misleading Implementers

`DESIGN.md:731-739` lists desired-running precedence as five flat, sequential steps:
1. Terminal/closed blocks wake.
2. Holds/quarantine/missing-config/identity-conflict produce desired-blocked.
3. Pending-create, pin, attachment, pending, named-always, work, scale-demand, explicit wake produce desired-running.
4. Pool demand is clamped/budgeted by pool desired-state rules.
5. Provider-health red suppresses respawn.

However, in the actual reconciler:
- **Step 4 (Pool demand)** is computed **before** the decider — `poolDesired` feeds `ScaleCheckCounts` into `AwakeInput` via the bridge (`compute_awake_bridge.go:30`). It is a pre-decider input, not a precedence layer inside the decider.
- **Step 5 (Provider-health red)** is applied **after** the decider — health-red and circuit-breaker checks suppress respawns on the output of `ComputeAwakeSet` (`session_reconciler.go:1966-2054`). They are post-decider gates.

Listing all five as a flat list inside the session domain contract invites implementers to fold provider-health, circuit state, and budgets into the session eligibility decider.

* **Required Change**: Restructure the desired-running precedence in `DESIGN.md` into three distinct execution layers matching the actual runtime:
  1. **Reconciler Pre-Decider**: Computes pool demand, scale clamps, and budgets.
  2. **Session Eligibility Decider**: Decides steps 1-3 only, consuming pure lifecycle/config facts without provider-health, circuit state, or budgets.
  3. **Reconciler Post-Decider Gates**: Suppresses actual process start based on provider liveness, provider-health-red, circuit-breakers, and restart budgets.

---

### [Blocker] 4. `RuntimeStartIntent` Fields Remain Unattributed with Reversal Risk

`RuntimeStartIntent` (`DESIGN.md:376`) lists fields: "bead ID, template, provider, work dir, generation, instance token, session key, config hash".

The design does not state the ownership of these fields:
1. Are `provider`, `work dir`, and `config hash` upstream-resolved facts that the session command simply echoes, or are they session-derived selections?
2. If the session decider selects or resolves these fields, `internal/session` would need to import provider-routing logic — violating the core boundary ("Keep provider-specific runtime behavior behind `internal/runtime`" in `internal/session/AGENTS.md`).

* **Required Change**: Pin `RuntimeStartIntent` to "what, not how": explicitly document that `provider`, `work dir`, and `config hash` are upstream-resolved facts that the session command copies or echoes from the request, not session-derived selections. Detail that provider selection and API-control-plane rerouting remain behind `internal/runtime` and the reconciler's start executors, preventing reverse-coupling.

---

## Missing Evidence & Major Risks

- **Fact Completeness Gaps in `RuntimeFacts`**: `RuntimeFacts` in `internal/session/lifecycle_projection.go` remains a boolean snapshot (`Observed`, `Alive`, `Attached`, `Pending`). It lacks timestamps, provenance (which provider/adapter produced it), and completeness metrics. Stale or timed-out liveness probes must not be treated as durable session death; otherwise, destructive drain/close operations might trigger prematurely. The design must bridge `snapshotQueryPartial()` into the session-owned decider's input to fail-closed on incomplete runtime state.
- **Circuit Breaker Metadata Keys on Session Beads**: The reconciler writes nine `session_circuit_*` keys onto session beads (`session_circuit_breaker.go:43-53`). The state-ownership table assigns circuit state to reconciler policy, but storing them on the session bead creates a layering inconsistency. The design needs an explicit migration plan: either document these keys as an opaque, reconciler-owned namespace on session beads that `internal/session` ignores and the reconciler cleans up on close, or move them behind a dedicated session command.
- **Pool Scaling Partial Query Fail-Closed Behavior**: `DESIGN.md:713-718` details partial-fact fail-closed behavior for work queries but omits `ScaleCheckCounts`. If pool scale checks return partial templates, the decider must not make destructive drain/close decisions. `PoolScaleCheckPartialTemplates` must be bridged into the decider's input.

---

## Questions

1. After slice 5, is desired-running computed by one decider consuming all facts, or by composing the eligibility decider (slice 5) with the health/progress gates (slice 7)? If composed, which layer owns the final precedence fold?
2. For the eligibility↔demand co-dependence, does session own an eligibility predicate that the reconciler's demand passes call (a functional dependency from `cmd/gc` → `internal/session`, which is legal), or does the reconciler precompute an `eligible` flag per bead and hand it to the decider as a fact?
3. Is `AwakeDecision.HasAssignedWork` (`compute_awake_set.go:88`, consumed at `session_reconciler.go:1975`) the intended hand-off fact between the session-eligibility output and the controller demand/health gate, or does that crossing move?
4. Will `RuntimeStartIntent` be a closed enum so that adding a provider never edits `internal/session`?
5. For runtime-fact completeness, is the intent to move staleness detection into the decider (facts carry timestamps) or leave it in the adapter (the adapter decides what counts as "complete")?
