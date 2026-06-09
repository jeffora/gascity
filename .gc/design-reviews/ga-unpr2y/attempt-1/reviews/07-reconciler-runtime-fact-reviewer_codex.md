# Liam Okonkwo - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The owner split is now explicit and aligned with the persona mandate. `internal/session` may own lifecycle eligibility and identity classification over immutable facts, while controller/reconciler owns work demand, dependency state, dispatch scheduling, pool desired size, cold-start demand, nested caps, restart budgets, alert dedupe, progress policy, and idle-sleep policy (`internal/session/DESIGN.md:452`).
- Runtime facts are kept out of durable session truth. Runtime liveness, provider errors, process existence, and transcript/provider-specific facts belong to the runtime provider or worker/runtime adapter (`internal/session/DESIGN.md:459`).
- Provider-neutral runtime intent is bounded. The design allows only stable identity and launch-context fields when a command contract needs them, and explicitly excludes provider-specific scheduling, health, progress, budget, and alert decisions (`internal/session/DESIGN.md:467`).
- Destructive-action safety is called out directly: unknown, stale, partial, or provider-error runtime facts must reject destructive actions unless a current requirement and operation contract state a tested safe convergence rule (`internal/session/DESIGN.md:473`).
- The design recognizes the current reconciler shape. Desired-state building already precomputes assigned work, scale-check counts, named demand, partial query flags, and store refs outside session deciders (`cmd/gc/build_desired_state.go:443`, `cmd/gc/build_desired_state.go:636`), while current requirements keep provider health and progress policy in reconciler-owned behavior (`internal/session/REQUIREMENTS.md:124`).

**Critical risks:**
- [Major] The boundary matrix must be mechanically enforced, not only documented. The design says `BOUNDARY_MATRIX.yaml` rows include forbidden `internal/session` imports, fields, and policy decisions (`internal/session/DESIGN.md:477`), but Slice 0 should turn this into an import/file-set guard for pure deciders: no store queries, runtime provider calls, config traversal, event emission, work-demand compilation, alert/budget state, or direct wall-clock fallback inside the pure session decision set.
- [Major] Runtime intent could still become a policy tunnel if rows stay generic. Fields such as provider family, work directory, config hash, runtime session key, generation, and instance token are legitimate facts, but only in operation-specific contracts. The runtime-start, wake, drain, stop/interrupt, and cleanup rows must each name the exact allowed fields and prove that provider health, progress, budget, and alert policy remain outside `internal/session`.
- [Major] Partial-snapshot semantics need first-class proof. Current desired-state code suppresses drains when store or session queries are partial (`cmd/gc/build_desired_state.go:55`, `cmd/gc/session_reconciler.go:1248`), and current requirements rely on fail-open/fail-closed distinctions for provider health and progress (`internal/session/REQUIREMENTS.md:133`). Boundary rows must preserve those distinctions so a future session decider cannot turn incomplete work/runtime evidence into a close, drain, rollback, release, cleanup, or restart decision.
- [Minor] The design rightly requires hot-path budgets, but reconciler fact materialization is large and store-heavy today. Before delegation, budget rows need counting-store or benchmark evidence for session/work rows scanned, store calls, runtime probes, subprocesses, and partial-snapshot behavior (`internal/session/DESIGN.md:566`).

**Missing evidence:**
- No `BOUNDARY_MATRIX.yaml` rows exist yet for provider health, progress and idle thresholds, runtime observations, wake-cause production, partial snapshots, drain/drain-ack, runtime-missing cleanup, adapter provider actions, pool demand, or work-demand compilation.
- No pure-decider guard exists yet to fail imports or calls that would leak reconciler/runtime policy into `internal/session`.
- No concrete runtime-intent schema exists yet for any operation, so the field list is still a design upper bound rather than an enforceable contract.
- No large-city or counting-store proof exists yet for reconciler fact compilation budgets.

**Required changes:**
- No design-text blocker remains in this persona lane. The document names the right ownership boundaries and rejection rules.
- Before closing Slice 0, add `BOUNDARY_MATRIX.yaml` rows for every required boundary named in the design, with source owner, policy owner, allowed session inputs/outputs, forbidden imports/facts, freshness, stale/unknown/partial/provider-error result, destructive-action safety, recovery owner, and negative fixtures.
- Add a mechanical pure-decider guard for the file set that will contain session deciders. It should reject store/runtime/config/event/work-query imports, direct provider-health/progress/budget fields, and direct `time.Now` fallback unless the file is outside the pure set.
- For each operation-specific runtime intent, require an allowed-field list and tests proving provider-specific execution and policy remain in the worker/runtime adapter or reconciler.
- Add partial-fact fixtures proving destructive actions are rejected or deferred when work queries, session queries, runtime observations, or provider-health/progress facts are unknown, stale, partial, or provider-error, except where `REQUIREMENTS.md` explicitly preserves fail-open behavior.

**Questions:**
- What exact package/file set will be guarded as pure session deciders?
- Will `RuntimeIntent` be one shared type, or should each operation define its own intent/result type to avoid optional-field drift?
- Which component owns the durable freshness timestamp for runtime observations: runtime provider, worker adapter, reconciler snapshot, or the operation contract row?
