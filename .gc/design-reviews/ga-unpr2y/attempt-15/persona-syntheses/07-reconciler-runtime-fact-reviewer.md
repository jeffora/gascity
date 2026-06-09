# Liam Okonkwo

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The high-level reconciler/session/runtime split is directionally correct, but `BOUNDARY_MATRIX.yaml` is not specified enough to make that split enforceable. All sources want policy and demand to stay outside `internal/session`; Codex requires a row schema that records source owner, policy owner, freshness, unknown/stale/partial/provider-error behavior, destructive-action policy, proof selectors, and forbidden session imports or fields.
- [Blocker] Provider-health and progress semantics are not anchored to active proof. Claude says the design asserts health/progress ownership against behavior that is absent on this branch and should be anchored to the ref where it exists; Codex requires explicit fail-open/fail-closed rows for red, absent, stale, unknown, startup grace, attached, pending interaction, and precedence behavior; DeepSeek requires restoring or replacing the missing provider-health, progress, and scale-from-zero tests before Slices 5 or 6 proceed.
- [Blocker] Wake-cause production is the key seam and remains unnamed. Claude requires the design to state that the reconciler/worker adapter produces the `[]WakeCause` candidate set from demand and schedule facts while `internal/session` only consumes those facts and applies lifecycle blockers. Codex makes the same point for `WakeCauseScaleDemand`: it must be an opaque controller-supplied fact, not permission for session deciders to compute pool demand.
- [Blocker] Runtime fact and destructive-action handling needs concrete row-level semantics before extraction. Codex requires rows for runtime observation fallback, partial work/session snapshots, drain cancellation, and runtime-missing cleanup. Unknown, stale, partial, provider-error, failed-query, and timeout states must say whether they fail open, fail closed, block destructive actions, or remain advisory.
- [Major] Pure fact isolation still needs physical guards and one clock-purity decision. Claude praises the existing `LifecycleInput` fact-isolation shape, Codex asks for forbidden-import/field rows for deciders, and DeepSeek blocks on the current `ProjectLifecycle` zero-`Now` fallback to `time.Now().UTC()`. The synthesis assessment is that mutation-feeding deciders must receive `now` as an explicit non-zero fact, and enrolled decider files need AST/import guards.
- [Major] RuntimeIntent is bounded in intent but still has two lane-specific edge risks. Claude flags `provider family` as routing-only data that must never drive session lifecycle policy. DeepSeek flags `session_key` as an observation token unless the design specifies deterministic prepare-time generation. Both need explicit wording before runtime-start slices depend on the intent shape.
- [Major] Drain and runtime-missing ownership remain unclear. Claude asks whether `sleep_reason="runtime-missing"` stays reconciler policy or becomes the future Atomic Command Contract cleanup operation; Codex requires drain rows split across session generation/stale-drain eligibility, controller work facts, runtime liveness facts, and adapter provider stop actions.
- [Minor] The design's warning against a single large `SessionFacts` struct could be misread. Claude notes that per-operation fact structs such as `LifecycleInput` are the correct purity pattern and should be explicitly blessed.

**Disagreements:**
- There is no verdict disagreement: Claude, Codex, and DeepSeek all block.
- The sources frame the provider-health/progress gap differently. Claude says the design may be correct but must anchor rows to the ref where the gates exist and record their absence on this branch. Codex and DeepSeek emphasize that the active checkout lacks the cited proofs and must restore, replace, or retire them. My assessment: the plan needs both branch-state disclosure and a hard exit gate before related slices.
- DeepSeek makes the zero-`Now` fallback a primary blocker; Claude treats `LifecycleInput` as evidence that fact isolation is already mostly healthy, and Codex focuses on matrix enforcement. My assessment: this is a blocker only for any mutation-feeding decider path; read-only rendering compatibility can be handled separately if explicitly excluded.
- DeepSeek uniquely argues that `session_key` in `RuntimeStartIntent` is a prepare/observe category error. Claude and Codex do not independently prove that point, but it is aligned with their bounded-intent concern and should be resolved in the RuntimeIntent contract.
- Claude wants the wake-cause production seam named; Codex asks for a broader matrix schema. These are compatible: the wake-cause seam should be one required row family in the broader matrix.

**Missing evidence:**
- A concrete `BOUNDARY_MATRIX.yaml` schema with source owner, policy owner, allowed session input/output, freshness rule, unknown/stale/partial/provider-error policy, destructive-action policy, current code selectors, current test selectors, and forbidden session imports or fields.
- Rows for provider health, progress health, runtime observation fallback, pool/cold-start demand, wake-cause production, partial work/session snapshots, drain cancellation/completion, runtime-missing cleanup, and adapter provider actions.
- Branch-state proof for provider-health, progress, and scale-from-zero behavior: which tests and implementation paths exist on this branch, which exist only on another ref, and what must be restored, replaced, or intentionally retired.
- A statement that production of all non-explicit wake causes stays reconciler-side and crosses into session code only as already-computed facts.
- A split for drain decisions: session-owned generation/stale-drain eligibility, controller-owned assigned-work and partial-query facts, runtime-owned liveness/stop result, and adapter-owned provider stop execution.
- A pure-decider file set and automated guard for direct wall-clock reads plus forbidden store, runtime, config, event, API/CLI rendering, mail/work, provider-health, progress, pool-demand, and scheduling imports or fields.
- A RuntimeIntent contract that resolves `provider family` as routing-only data and `session_key` as either deterministic prepare-time identity or post-start observation, not both.
- A disposition for existing reconciler observation-to-durable writes such as runtime-missing sleep conversion.

**Required changes:**
- Define the `BOUNDARY_MATRIX.yaml` row schema in the design before approving behavior-moving slices. Include owner, freshness, unknown/stale/partial/provider-error semantics, destructive-action safety, proof selectors, and forbidden session imports/fields.
- Add explicit matrix rows for provider health and progress health preserving red, absent, stale, unknown, startup grace, attached, pending-interaction, and precedence behavior. Restore, replace, or retire the missing proof files before Slices 5 or 6 decompose.
- Name the wake-cause production seam: reconciler/worker adapters produce candidate wake causes from demand and schedule facts; `internal/session` consumes those causes and applies lifecycle blockers. State that Slice 3 moves only explicit wake and not the other wake-cause producers.
- Document `WakeCauseScaleDemand` and future pool-related session inputs as opaque controller-supplied facts. Forbid session deciders from receiving raw work beads, pool desired counts, scale-check output, dependency state, rig-store reachability, restart budgets, or alert state.
- Add rows for partial work/session snapshots stating that partial snapshots are controller-owned, suppress orphan/drain/destructive close decisions, and may cross into session code only as an opaque "facts incomplete, reject destructive action" fact.
- Split drain ownership into session generation/stale-drain eligibility, controller assigned-work and query-completeness facts, runtime liveness/stop facts, and adapter provider-stop execution.
- Bless per-operation fact structs such as `LifecycleInput` as the required decider input pattern while rejecting one giant `SessionFacts` aggregate and any store/reconciler helper dependency inside deciders.
- Add the RuntimeIntent caveats: provider family is routing/identity data only, session deciders must not branch lifecycle policy on it, and `session_key` must be either deterministic prepare-time data or a post-start observation fact.
- Decide whether runtime-missing durable conversion remains reconciler policy or becomes a session-owned cleanup command, and add the corresponding matrix and command-contract rows.
- Enforce pure-decider guards, including rejection of direct `time.Now()` reads in enrolled decider code and mandatory explicit `now` facts for mutation-feeding projections.
