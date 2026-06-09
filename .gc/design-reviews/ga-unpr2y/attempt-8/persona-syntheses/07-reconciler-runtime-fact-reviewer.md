# Liam Okonkwo

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The eligibility/demand split is still under-specified around min-active behavior. Claude and DeepSeek both identify `countMinActiveCovered` and `MinActiveSessions` as controller demand policy currently embedded in `ComputeAwakeSet`; Codex raises the same issue as the broader risk that `ComputeAwakeSet` still mixes session eligibility with controller demand, work aggregation, pool/named-session demand, and direct wake decisions. The design needs an explicit owner and parity proof before this logic can move.
- [Blocker] Slices 6 and 7 cannot safely proceed until missing health, progress, and scale proof files are restored or replaced. All sources call out missing coverage for provider-health red/unknown/stale behavior, session progress behavior, and scale-from-zero or scale-demand preservation. The design correctly says some of this is blocked, but the gate needs to be non-negotiable and executable.
- [Blocker] Pure-decider fact isolation is not mechanically enforceable yet. Claude and DeepSeek identify `ProjectLifecycle` `time.Now().UTC()` fallbacks as direct violations if the projection is treated as a pure decision surface; Codex also flags the missing import/static guard. A prose-only "no clocks or I/O" rule is insufficient for this boundary.
- [Major] `RuntimeStartIntent` / `RuntimeIntent` is not yet a narrow bounded adapter contract. Claude and DeepSeek point to `session_key` as a prepare-time versus observe-time category error unless deterministic pre-start derivation is guaranteed. Codex generalizes the risk: without an exact field list and non-goals, the intent can become a back channel for provider probing, restart budgets, progress policy, scheduling, or durable runtime-missing truth.
- [Major] Runtime observations must stay non-durable truth, but durable probe evidence still lacks a freshness/supersession rule. Claude and DeepSeek both flag W-013 detached-probe metadata as a place where stale evidence could gate destructive orphan release; Codex asks for partial-fact tests that fail destructive close, drain, rollback, cleanup, release, and restart decisions when runtime/work/config facts are stale or partial.
- [Major] The new eligibility mask and `AwakeInput` disposition table are the right direction, but they need field-level proof. All sources support the seven-kind eligibility mask and adapter-precomputed facts, while also requiring tests for forced-runnable, conditional-runnable, idle-suppressible, blocked, terminal, unknown/partial, and repair-needed behavior across the split.

**Disagreements:**
- Claude and Codex verdicts are `approve-with-risks`; DeepSeek verdict is `block`. My assessment is `block` because the remaining risks are not just implementation cautions: min-active ownership, clock fallbacks in a named pure baseline, missing health/progress/scale proofs, and the intent prepare/observe boundary are prerequisites for decomposition of this lane.
- Claude classifies the `ProjectLifecycle` clock fallback as Major because `ComputeAwakeSet` itself is clean; DeepSeek classifies it as Blocker. I side with the blocker classification for the design gate unless the design explicitly excludes `ProjectLifecycle` from the pure-decider surface or removes the fallback before using it as baseline evidence.
- Codex says the runtime observation completeness model is a strength, while Claude and DeepSeek still require W-013 freshness and fail-closed binding. These are compatible: the taxonomy is useful, but it is not complete until persisted operational evidence has a supersession rule and destructive-action fixtures.
- DeepSeek uniquely raises unmapped traceability requirements (`SESSION-RECON-001`, `SESSION-WORK-003`, `SESSION-RUNTIME-004`) and omitted production writers (`cmd_stop.go`, `cmd_wait.go`). The other sources do not contradict this; the synthesis treats it as required follow-up evidence to verify before approval.

**Missing evidence:**
- Restored or replacement tests for `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go`, plus coverage for alert dedupe, restart-budget preservation, progress exemptions, provider-health suppression, and demand behavior.
- A named pure-decider package or file set and a concrete import/static guard proving no store, runtime provider, work query, config loader, provider-health helper, clock, or reconciler helper can enter pure session decision code.
- A zero-`Now` policy for `ProjectLifecycle` and proof that the current `time.Now().UTC()` fallbacks are removed, rejected, or kept outside the pure-decider surface.
- Field-disposition rows and parity fixtures for `MinActiveSessions`, min-active coverage, attached and pending interaction facts, pinned/named-always behavior, wait readiness, assigned-demand work, idle-sleep protection, work counts, pool size, runtime liveness, and progress facts.
- A concrete `RuntimeStartIntent` / `RuntimeIntent` field list with explicit non-goals and either no `session_key` at prepare time or a documented deterministic pre-start identity invariant.
- Freshness/supersession keys for W-013 and operational lifecycle evidence, plus explicit fail-closed behavior for detached-probe timeout or error outcomes.
- Verification that the scenario traceability matrix accounts for `SESSION-RECON-001`, `SESSION-WORK-003`, `SESSION-RUNTIME-004`, and any other currently unmapped boundary requirements.
- Granular writer inventory rows for active production writers such as `cmd/gc/cmd_stop.go` and `cmd/gc/cmd_wait.go`, if they still write session-owned keys.

**Required changes:**
- Add explicit `AwakeInput` disposition rows for `MinActiveSessions` and every idle-suppressible input fact, and state whether min-active coverage is controller-side demand composition, a precomputed protection fact, or a simple deficit fact consumed by the decider.
- Add parity tests for the extracted eligibility mask covering forced-runnable, conditional-runnable, idle-suppressible, blocked, terminal, unknown/partial, repair-needed, min-active, wait-only-wake, attachment, pending interaction, and pinned/named-always behavior.
- Keep slices 6 and 7 blocked until scale-from-zero, provider-health, and session-progress proof files are restored or replaced and the tests preserve fail-open, exemption, dedupe, budget, progress, and demand semantics.
- Define the pure-decider destination and guard as an executable artifact, then remove or explicitly isolate the `ProjectLifecycle` wall-clock fallbacks before the projection is used as pure decision evidence.
- Narrow `RuntimeStartIntent` / `RuntimeIntent` to adapter request fields only, with explicit non-goals for provider probing, restart budgets, scale decisions, progress thresholds, alert dedupe, and durable runtime-missing truth; resolve the `session_key` prepare/observe boundary.
- Define W-013 and operational evidence supersession, freshness, and fail-closed error/timeout behavior before any slice 6/7 reader conversion relies on persisted probe metadata.
- Update the scenario traceability matrix and writer inventory for the boundary requirements and active writers DeepSeek flagged, or document why each is out of this design's scope with a named proof path.
