# Liam Okonkwo — DeepSeek V4 Flash (Independent Review, Attempt 6)

**Verdict:** block

**Persona:** reconciler boundary, runtime intent adapter ownership, fact isolation, health gate split.

**Reviewed against:** `internal/session/DESIGN.md` (Attempt 6, matching `.gc/design-reviews/ga-unpr2y/attempt-6/design-before.md` with Attempt 6 changes), `internal/session/lifecycle_projection.go`, `cmd/gc/compute_awake_set.go`, `cmd/gc/compute_awake_bridge.go`, and prior design iterations.

---

## Top Strengths

- **Successful Deconstruction of the `AwakeInput` Blocker**: 
  The introduction of the three-stage pipeline (`DESIGN.md:830-840`) is a major architectural victory. By splitting the hot reconciler pass into (1) Pre-decider fact assembly, (2) Pure session eligibility mask computation inside `internal/session`, and (3) Reconciler-side combination with policy, the design successfully prevents controller-domain scheduling and capacity facts (such as `ScaleCheckCounts` and `WorkBeads`) from leaking into the session package. This resolves the fusion risk from Attempt 5 while respecting the Bitter Lesson.

- **Mechanically Guardable Pure Decider Boundaries**: 
  Requiring pure deciders to reside in a mechanically guardable file set with strict import restrictions (no stores, runtime providers, clocks, files, work queries, etc.) provides a highly enforceable static structure. This ensures that the purity of `internal/session` deciders is maintained by static analysis rather than relying on developer discipline.

- **Grandfathering Exception and Coexistence Clarity**: 
  Expanding the exception list (W-022 through W-028) and adding clear coexistence gates and rollback matrices shows a realistic, production-minded transition strategy. It avoids risky "flag day" assumptions and accommodates the live codebase's complexity (such as `RepairEmptyType` and direct manager lifecycle methods).

---

## Critical Risks & Blockers

### [Blocker] 1. The Eligibility Mask Enum Overloads "Runnable," Breaking Idle-Sleep and Pins/Attachments

The newly proposed three-stage pipeline specifies:
> `internal/session` receives only immutable lifecycle and identity facts and returns an eligibility mask: runnable, blocked with reason, terminal, repair-needed, or unknown/partial.

However, this flat mask is insufficient for the reconciler's subsequent decision phase:
1. **The Overloading Problem**: Both pinned/attached/explicitly-woke sessions and normal idle sessions (with no assigned work) are not terminal, blocked, held, or quarantined. Under this contract, both will return an eligibility mask of `runnable`.
2. **The Information Gap**: In Step 3, the reconciler must combine the eligibility mask with work demand. If there is no work demand, the reconciler will put a normal session to sleep (idle-sleep). However, a pinned, attached, or explicitly-woken session *must* remain awake even with zero work demand.
3. **The Boundary Clash**: Since the reconciler is forbidden from reading session-bead metadata directly (the core session package boundary), the reconciler in Step 3 has no way of knowing *why* a session is `runnable`. It cannot inspect `pin_awake`, `attachment`, or `wake_request`.

Without a way to distinguish "forced runnable" from "conditionally runnable (subject to idle-sleep)," the reconciler will mistakenly put pinned or attached sessions to sleep on the very next tick.

* **Required Change**: Richer semantics must be returned by the eligibility mask. Either output distinct categories (e.g., `runnable_forced` with sub-reasons like `pinned`/`attached`/`explicit_wake` versus `runnable_conditional`), or expand the mask's struct to explicitly bubble up the session-domain overrides so the reconciler can safely evaluate idle-sleep exemptions.

---

### [Blocker] 2. Desired-Running Precedence Table Contradicts the 3-Stage Pipeline

Despite introducing the three-stage pipeline under `## Reconciler Fact Contract`, the design retains a flat 1-to-5 sequential list under `Desired-running precedence must remain explicit:` (`DESIGN.md:869-879`):
> 3. pending create, pin, attachment, pending interaction, named-always, targeted work, scale demand, and explicit wake produce desired-running when not blocked

This step mixes session-domain eligibility facts (`pin`, `attachment`, `explicit wake`) with reconciler-domain scheduling facts (`targeted work`, `scale demand`). 
This flat structure directly contradicts the pipeline:
- If Step 2 (Session decider) evaluates Step 3 of the precedence list, it must import or inspect work demand/scaling counts, violating the import ban.
- If Step 3 (Reconciler) evaluates Step 3 of the precedence list, it must inspect pins/attachments/explicit-wake, violating session-metadata ownership.

This flat list represents a classic case of pattern drift, where a new paragraph was added to address a blocker but the existing flat table immediately below was left unaligned.

* **Required Change**: Realign the Desired-Running Precedence list to match the three-stage pipeline. Explicitly specify which rows of the precedence list are processed in Step 2 (as pure session eligibility) and which are processed in Step 3 (as reconciler policy combination).

---

### [Blocker] 3. `RuntimeStartIntent` Vocabulary Checklist Remains Under-Attributed

The `RuntimeStartIntent` vocabulary checklist entry (`DESIGN.md:444`) still includes fields such as `provider`, `work dir`, and `config hash` without clarifying their provenance.
As noted in prior reviews, if the session command selects or resolves these fields internally, `internal/session` will be forced to import runtime provider routing or configuration-hash calculations, creating a major reverse-coupling violation.

* **Required Change**: The vocabulary checklist must explicitly state that `provider`, `work dir`, and `config hash` are upstream-resolved facts that the session command simply copies/echoes from the request, rather than session-derived selections. Document that provider selection and ACP routing stay behind `internal/runtime` and the reconciler's start executors.

---

### [Blocker] 4. Missing Provider-Health and Progress Proof Files Lacks a Recovery Plan

The design correctly recognizes that slices 6 and 7 are blocked because the cited proof files (`cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go`) are missing from the repository.
However, the design response merely acknowledges this block without outlining how the implementation team should recover. If these files are gone permanently, how are the health/progress gates to be verified?

* **Required Change**: Add an explicit recovery plan to `DESIGN.md:1300`. Define whether these files must be re-created from scratch or if they can be replaced by alternative, equivalent test coverage, specifying the exact minimum verification criteria for health/progress/scaling before slices 6 and 7 can proceed.

---

## Remaining Major Risks

- **Fact Completeness Gaps in `RuntimeFacts`**:
  `RuntimeFacts` in `internal/session/lifecycle_projection.go` remains a simple boolean snapshot (`Observed`, `Alive`, `Attached`, `Pending`). It still lacks timestamps, provenance, and completeness flags. As a result, the decider cannot distinguish between a timed-out probe and a successful probe confirming death, meaning destructive close/drain decisions remain vulnerable to premature triggering.

- **Circuit Breaker Metadata Keys Layering Leak**:
  The reconciler continues to write nine `session_circuit_*` metadata keys onto session beads. While grandfathered under exception W-010, storing reconciler-owned circuit state on session beads violates strict layering. The design lacks a clear path to eventually isolate or clean up these keys upon session close.

---

## Answers to Persona Questions

1. **Which wake, hold, drain, provider-health, and progress decisions move into session deciders, and which scheduling or budget responsibilities remain in the reconciler?**
   - *Answer*: Terminal status, holds, quarantine, missing config, and explicit wake decisions move into session deciders. Scheduling, pool scaling, work aggregation, provider health, progress, circuit state, and restart budgets remain in the reconciler. However, as noted in Blocker 1 and 2, the exact seam for pins and attachments remains structurally broken.
2. **Are work counts, pool size, runtime liveness, and progress facts precomputed by adapters instead of queried from deciders?**
   - *Answer*: Yes. The three-stage pipeline ensures the reconciler precomputes these facts and provides them as immutable inputs.
3. **Can RuntimeIntent express adapter needs without smuggling provider policy into internal/session?**
   - *Answer*: Yes, but only if the fields inside `RuntimeStartIntent` are strictly bound to being echoes of request facts, which still requires explicit documentation.
