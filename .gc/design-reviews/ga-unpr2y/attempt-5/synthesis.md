# Design Review Synthesis

## Overall Verdict: block

Nine persona syntheses returned `block`, and one returned `approve-with-risks`, so the global verdict is `block` by worst-verdict-wins. The design direction is widely viewed as sound, but the document is not yet decomposable because key contracts remain ambiguous exactly where implementation beads would need crisp source inventories, routing rules, concurrency guarantees, recovery guarantees, and proof obligations.

## Consensus Strengths
- Multiple personas praised the move toward session-owned command/decision contracts instead of scattered metadata writes.
- Reviewers agreed that durable facts and controller-side convergence are the right recovery foundation; events should accelerate or describe work, not be the only critical delivery path.
- The design's direction on `instance_token` as the runtime identity authority is stronger than the previous generation-centered model.
- The added row ownership baseline, compatibility matrices, and vocabulary checkpoint concepts are useful building blocks, even though they are not yet enforceable enough.
- Several reviewers viewed event sourcing, durable outbox work, and broad future vocabulary as correctly deferred when the document labels them as future rather than slice-1 requirements.

## Critical Findings

### [Blocker] Writer inventory and mutation guards are not source-complete
**Sources:** Elena Marchetti, Natasha Volkov, Takeshi Yamamoto, Sarah Chen, Ravi Krishnamurthy, Ingrid Holm

**Issue:** The design depends on a complete inventory of lifecycle, identity, repair, wake/hold/drain, create/start, close/retire, and patch-map writers, but reviewers found omitted or stale production call sites. Repeated examples include `RepairEmptyType` callers across API, CLI, mail, and session resolution; `session.WakeSession`; `CloseDetailed`; `Suspend`; `UpdatePresentation`; `PreWakePatch`; wake fallback metadata writes; `internal/api/session_resolution.go`; `cmd/gc/session_name_lookup.go`; `cmd/gc/adoption_barrier.go`; `cmd/gc/cmd_session.go`; `cmd/gc/cmd_handoff.go`; and `internal/session/chat.go` token backfills. The proposed guard also misses manager methods, package-level mutators, dynamic metadata batches, caller-side patch-map extension, generic store bridges, and top-level session-bead type/status mutation.

**Required change:** Regenerate a source-verified writer inventory before decomposition. Split broad W rows into concrete call-site rows with exact key sets or dynamic-key classes, target session-bead proof, owner slice, bake window, retirement condition, and proof test. Make the static guard a first-class gate that covers direct store writes, manager methods, package-level session mutators, patch builders and patch-map extension, dynamic batches, `Update`, `Close`, `Create`, and repair helpers, with a shrink-only allowlist and fixtures for every known bypass shape.

### [Blocker] Session command concurrency and atomicity are not enforceable
**Sources:** Takeshi Yamamoto, Ravi Krishnamurthy, Amara Osei, Ingrid Holm

**Issue:** The current contract still allows "re-read before write" followed by unconditional metadata updates, `Update`, `Close`, or sequential multi-key writes. That does not prevent races between validation and mutation, cross-process writers, or partially applied multi-key session commands. Close and runtime start are especially under-specified: provider side effects can happen before durable close/start commits are fenced, `CloseDetailed` and `closeBead` have divergent behavior, no-token `instance_token` backfills are unfenced, and provider-start success followed by metadata or event failure lacks a deterministic adoption/stop/retry rule.

**Required change:** Pick a testable concurrency model per command cluster: single-writer controller serialization, store-level conditional mutation/revision checks, command phase markers, post-write stale-success detection with bounded repair, or another explicit mechanism. Define commit preconditions, revalidated facts, conflict outcomes, partial-state matrices, lock or revision semantics, event ordering, provider compensation, and stale-success repair tests for close, wake, drain, runtime start, identity retirement, and repair.

### [Blocker] Target classification, policy, repair, and materialization are conflated
**Sources:** Amara Diallo, Sarah Chen, Ravi Krishnamurthy, Kwame Asante

**Issue:** Slice 1 is described as read-only target classification, but existing resolver paths can repair empty session types, retire named identifiers, and materialize sessions. The proposed classifier also mixes raw identity facts with policy decisions such as forbidden kind, closed-not-allowed, materializable, and requires-materialization. Current CLI, API, mail, extmsg, assignee, nudge, attach, inspect, log, transcript, and worker fallback surfaces have different precedence chains and side effects, and mail needs recipient sets and historical-alias read-only semantics rather than a single selected candidate.

**Required change:** Split raw target identity classification from per-operation policy, repair, and materialization. Add per-surface transition and routing tables for CLI, legacy API, Huma API, mail send/query, extmsg, assignee normalization, wake/nudge, attach/observe/log/transcript, and materializing resolver paths. State whether repair is a separate audited command or a typed repair-needed result, and require command-time revalidation before writable or materializing side effects.

### [Blocker] Event and recovery contracts cannot yet guarantee critical convergence
**Sources:** Amara Osei, Ingrid Holm, Takeshi Yamamoto

**Issue:** The per-event matrix promises facts and idempotency inputs that current registered payloads do not carry. Several session events are `NoPayload`, sparse lifecycle payloads lack fields promised by diagnostics, and `SessionDrainAckedWithAssignedWork` lacks drain-generation context. The close/work-release guarantee still leaves ambiguity between durable-fact scans and retained synchronous cascades. A synchronous cascade may reduce latency, but it cannot be the crash-recovery mechanism. The design also promises `event-emission-failed` even though the current recorder is best-effort and does not return failures to command code.

**Required change:** Decide whether event matrix fields are typed wire payload fields or durable-fact identities read by scans. If wire payloads change, add payload structs, registry entries, OpenAPI/SSE/client/dashboard regeneration, and rendering tests. If events remain thin triggers, label them as hints and define scan-side convergence keys. Critical work release must have a mandatory durable-fact backstop with owner, cadence, coverage, partial-read behavior, duplicate-run behavior, stale-event supersession, and crash-after-commit tests for CLI/API and controller close paths.

### [Blocker] The reconciler/session boundary still leaks controller policy
**Sources:** Liam Okonkwo, Takeshi Yamamoto, Ingrid Holm, Ravi Krishnamurthy

**Issue:** `AwakeInput` and desired-running language mix session lifecycle facts with controller demand, work aggregation, scale checks, pool capacity, provider health, progress, restart budgets, circuit state, config identity, alert dedupe, and trace behavior. Reviewers agreed that demand computation and eligibility currently influence each other, so an implementer could move reconciler policy into `internal/session` while satisfying the current prose. Runtime observation facts also lack a complete unknown/stale/partial/provider-error model for destructive decisions.

**Required change:** Define a narrow session-owned eligibility fact contract, separate from reconciler-owned demand assembly and post-decider health/circuit/budget gates. A workable model is reconciler pre-decider demand and scale inputs, a session eligibility mask, then reconciler-owned demand resolution and post-decider suppression or restart gates. Add runtime observation completeness fields and fail-closed rules for unknown, stale, partial, provider-error, observed-missing, and alive states. Put pure deciders in a mechanically guardable subpackage or file set.

### [Blocker] Behavior parity proof is not yet decomposable
**Sources:** Natasha Volkov, Sarah Chen, Amara Diallo, Ingrid Holm

**Issue:** The row ownership baseline is stronger than before, but the same obligations are not propagated into per-slice matrices, backlog proof bullets, and concrete fixture names. Reviewers found missing or weak proof ownership for `SESSION-STATE-001`, `SESSION-STATE-002`, `SESSION-STATE-003`, `SESSION-START-004`, `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007`. Several current-lineage behaviors may be absent or unproven, including provider-health fail-open, progress exemptions, startup grace, scale-from-zero wake, and CLI/API compatibility details. Citation freshness is still a promise rather than a named gate.

**Required change:** Propagate scenario-row ownership into every slice that moves behavior, then name exact endpoint, generated-client, CLI stdout/stderr/JSON, exit-code, status-code, error-body, request-ID, and dashboard/schema proof fixtures. Add a citation-freshness gate that fails on missing paths, non-ancestor commit citations used as current proof, and assertions that no longer prove the row. For divergent reconciler rows, restore current-lineage behavior with tests or amend `REQUIREMENTS.md` with owner approval before decomposition.

### [Major] Operability diagnostics are promised but not specified as concrete surfaces
**Sources:** Ingrid Holm, Amara Osei, Sarah Chen, Takeshi Yamamoto

**Issue:** The design promises operators can understand why sessions were blocked, woken, drained, closed, repaired, or deferred, but it does not define enough trace schemas, centralized reason/outcome codes, doctor checks, rendering tests, or typed event payloads. Several outcome vocabularies still appear across the design, and richer diagnostics such as `SessionWorkQueryFailed` exceed the current typed payload shape.

**Required change:** Add a trace vocabulary migration table and parity test for every decider reason and outcome. Define doctor or session-inspect check names, statuses, detail fields, data sources, fixability, and rendering tests. Tie every diagnostic promise to one concrete surface: trace, doctor, typed event payload, OpenAPI/SSE/dashboard client, or log.

### [Major] Migration coexistence and rollback contracts are too operationally loose
**Sources:** Ravi Krishnamurthy, Sarah Chen, Elena Marchetti, Takeshi Yamamoto

**Issue:** Partial adoption can leave old and new writers mutating the same key families with different validation. Runtime-start, close/retire, wake/hold/drain, API routing, materialization, repair, and cross-family operations need exact bake windows, reverted files/callers, guard rows, test commands, and one-writer restoration criteria. The in-process mutation lock does not serialize direct CLI/API store writers unless those paths route through the protected boundary.

**Required change:** Add a per-key owner matrix for every affected key family and a per-operation coexistence policy for cross-family commands. Every implementation bead should state what it converts, allows during bake, retires, and proves. Bake and rollback instructions must identify exact files, call sites, command APIs, guard exceptions, tests, and minimum exit criteria.

### [Major] Scope and vocabulary controls remain advisory
**Sources:** Kwame Asante, Liam Okonkwo, Ingrid Holm

**Issue:** The 14-kind target taxonomy, broad policy dimensions, command conflict vocabulary, runtime intents, event fact types, and diagnostic fields can still be read as day-one scaffolding instead of per-slice vocabulary introduced by first use. This risks replacing scattered metadata logic with a broad facade or mega-slice.

**Required change:** Add a vocabulary-checkpoint hook to task or bead metadata. For each new type or field, record first delegated caller, exact demanded fields, rule-of-two status, non-goals, provisional bounds, and reviewer evidence. Mark future-slice vocabulary as maximum bounds that may shrink, and reject future-slice scaffolding before its owning slice and first caller exist.

### [Major] Performance and backpressure gates lack numeric proof
**Sources:** Ingrid Holm, Liam Okonkwo, Amara Osei

**Issue:** Reviewers repeatedly asked for query-count, benchmark, or threshold tests for target classification, named lookup, mail recipient expansion, session snapshots, close/work-release scans, runtime-start repair, extmsg notification, event emission, subscriber fan-out, scale-check snapshots, and reconciler hot-loop writes. Subscriber limits and critical/optional queue behavior remain qualitative.

**Required change:** Add bounded-query tests or benchmarks for hot paths and large-city scans. Define synchronous subscriber caps, queue sizes, retry/defer/drop semantics, critical versus optional tier behavior, and trace evidence for deferred, skipped, failed, retried, or shed subscriber work.

### [Minor] Several artifact and terminology inconsistencies still need cleanup
**Sources:** Natasha Volkov, Ravi Krishnamurthy, Kwame Asante, Ingrid Holm

**Issue:** Some paths and labels are stale or inconsistent, such as `internal/mail/beadmail.go` versus `internal/mail/beadmail/beadmail.go`, `_gemini.md` files whose content identifies DeepSeek, and outcome vocabulary lists that still diverge between artifacts. `path-alias`, `historical-alias`, `live-session-name`, and `live-alias` also need clearer surface and projection terminology.

**Required change:** Normalize stale paths, source labels, outcome names, and target vocabulary before task generation so implementers do not decompose from ambiguous or obsolete terms.

## Disagreements
- Verdict disagreements were local, not global. Behavior parity, event delivery, API/CLI boundary, and reconciler-boundary lanes had split model verdicts, but the persona syntheses chose `block` because the disputed risks affect decomposability and safety.
- Kwame Asante's scope lane returned `approve-with-risks` rather than `block`. I agree that the design direction is acceptable there, but those risks still become required changes because vocabulary creep can undermine the first implementation slice.
- Reviewers allowed multiple acceptable concurrency models: CAS/revision checks, single-writer serialization, phase markers, or post-write verification with bounded repair. The design does not need to pick CAS specifically, but it must reject plain unconditional writes after a pre-read.
- Event reviewers disagreed on whether the per-event matrix should become wire payloads or durable scan facts. Either path is acceptable if the document labels it consistently and assigns the payload or scan proof work.
- Target-classification reviewers offered two acceptable repair strategies: keep classification pure and return repair-needed facts, or perform repair in an audited compatibility adapter. The current text must choose one because "read-only classifier" conflicts with existing repair/materialization behavior.
- Source emphasis varied by persona: some reviewers focused on concrete omitted call sites, while others focused on abstract invariants. These are complementary. The global assessment treats source inventory and invariant contract as mutually required, not alternative fixes.

## Missing Evidence
- Source-verified writer inventory for every session lifecycle, identity, repair, create/start, wake/hold/drain, close/retire, materialization, patch-map, and top-level session-bead mutation path.
- Static guard fixtures proving detection of direct store writes, manager calls, package-level mutators, dynamic batches, patch builders, patch-map extension, `Update`, `Close`, `Create`, and repair helpers outside approved boundaries.
- Store atomicity and concurrency proof for `SetMetadataBatch`, `Tx`, `Update`, and `Close`, including cross-process writer behavior.
- Per-surface resolver and routing matrices for CLI, legacy API, Huma API, mail, extmsg, assignee normalization, worker fallback, nudge/wake, attach/observe/log/transcript, and materialization.
- Current-lineage tests or owner-approved requirements amendments for provider health, progress, scale-from-zero, startup grace, and reconciler fail-open rows.
- Citation-freshness gate path, invocation, and failure semantics.
- Event payload compatibility table covering current registered payloads, final payload or durable-fact identity, idempotency/convergence keys, OpenAPI/SSE/dashboard impact, and migration plan.
- Durable work-release scan owner, cadence, store coverage, completeness guard, partial-read behavior, duplicate-run behavior, stale-event supersession rule, and CLI/API crash-after-close proof.
- Trace, doctor, and session-inspect schemas with centralized reason/outcome codes and rendering tests.
- Runtime observation completeness model and destructive-action fail-closed tests.
- Pure-decider demarcation mechanism and import/call guard for clock, config, files, stores, runtimes, event emitters, provider health, and controller demand access.
- Query-count tests, benchmarks, subscriber caps, queue/backpressure behavior, and reconciler hot-loop write budgets.
- Bake, rollback, and revert drills for each slice with exact files, callers, command APIs, guard rows, and proof commands.

## Recommended Changes
1. Regenerate and freeze the source writer inventory, split broad W rows, and make the mutation-boundary guard a required pre-decomposition deliverable.
2. Choose the session command concurrency model and define atomicity, partial-state, stale-success, provider-compensation, and event-ordering contracts per command cluster.
3. Split target identity classification from operation policy, repair, materialization, and per-surface routing; add transition tables and parity fixtures for every affected surface.
4. Decide the event contract: typed payload migration or durable-fact scan identities. Then make critical work release converge through a mandatory durable scan backstop.
5. Redraw the session/reconciler boundary around a narrow eligibility contract, reconciler-owned demand and health gates, runtime observation completeness, and pure-decider guards.
6. Propagate scenario-row ownership into slice proof bullets and add a citation-freshness gate before implementation beads are generated.
7. Turn operability promises into concrete trace, doctor, typed-event, OpenAPI/SSE/dashboard, and rendering-test deliverables with centralized outcome vocabulary.
8. Add structured migration metadata to every implementation bead: converts, allows during bake, retires, guard-row changes, proof commands, bake exit criteria, and rollback path.
9. Enforce vocabulary checkpoints so each new shared type or field has a first caller, demanded fields, rule-of-two evidence, and a declared non-goal.
10. Add performance and backpressure gates with numeric thresholds for lookup, scan, event, subscriber, and hot reconciler paths.
