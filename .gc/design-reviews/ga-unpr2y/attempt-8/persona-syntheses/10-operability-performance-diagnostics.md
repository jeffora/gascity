# Ingrid Holm

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The read-path repair write remains operator-invisible and unscheduled. Claude and DeepSeek both verify that `internal/session/resolve.go` still calls `RepairEmptyType` from read paths, swallows the persistence error, and mutates the in-memory bead type anyway. The design assigns this to a "repair slice" or repair ownership in multiple places, but the backlog contains no actual repair slice, so the fix has no schedulable owner.
- [Blocker] Operability requirements are not yet machine-checkable. Codex requires a diagnostics manifest mapping each operation to required trace fields, doctor/session-inspect mappings, renderers, and exact tests. Claude agrees the current diagnostic direction is better anchored in existing trace vocabulary, but the repair failure and fast-path/fallback cases still lack enforceable evidence. DeepSeek adds that speculative outcomes such as `missed-event-recovered` cannot be truthfully emitted without durable delivery tracking.
- [Major] Repair-path guard coverage is contradictory. Claude flags that doctor/repair utilities are excluded from the mutation-guard scan while the repair gate requires guard-visible allowlist rows. DeepSeek shows why this matters: direct repair helpers can hide mutations from a raw store-call scanner. The design must make repair writes visible to a guard or route them through one audited helper.
- [Major] Performance budgets lack current baselines and backend proof. Claude asks for production-backend proof that metadata filters are indexed rather than broad scans, plus large-city reconciler/resolution baselines. Codex asks for per-surface query budgets and lazy collection rules before Slice 1 delegates target lookup. DeepSeek adds write-side hot-loop risks, including synchronous throttle-marker writes and potential pool snapshot rebuild complexity.
- [Major] Trace outcome vocabulary is not mapped cleanly to current public tooling. Codex identifies gaps between proposed outcomes and existing `TraceOutcomeCode` values. Claude asks for a fast-path-versus-durable-scan distinction so operators can see when scans actually recover work. DeepSeek argues that "missed event recovered" is speculative without delivery tracking. The outcome model needs a compatibility table and truthful convergence terms.
- [Major] Doctor, inspect, and trace operator surfaces are under-owned. Codex notes that named checks such as `session.target-resolution`, `session.command-precondition`, `session.runtime-start`, `session.close-work-release`, `session.drain-assigned-work`, and `session.repair` are not tied to a registration or golden proof. DeepSeek adds that there is no session-scoped trace surface to answer why a specific session is blocked now.
- [Major] Event/subscriber recovery diagnostics still lack fan-out and backpressure contracts. Codex requires slow-subscriber, failed-subscriber, duplicate-event, skipped-event, and durable-scan retry tests. Claude asks for honest scan convergence reporting. DeepSeek reinforces that scans cannot claim missed-event recovery unless there is durable evidence of a missed event.

**Disagreements:**
- All three raw reviews verdict `block`, so there is no verdict disagreement. The only assessment choice is scope: the persona block is anchored first on the unresolved repair-write owner and second on missing executable diagnostics.
- Claude treats the trace/doctor direction as real progress because it uses existing centralized vocabulary. Codex says that is insufficient without a manifest and check registry. Assessment: keep the direction, but require machine-readable mapping and renderer tests before approval.
- Claude says the event-delivery story is mostly honest because recorder APIs cannot report every delivery failure. DeepSeek says `missed-event-recovered` is still speculative for the same reason. Assessment: keep durable-scan convergence, but remove missed-event claims unless the design adds durable delivery tracking.
- Claude focuses on read-cost and metadata index proof, Codex on per-surface lazy classifier budgets, and DeepSeek on write-side hot-loop costs and pool rebuild complexity. Assessment: the performance section must cover both read and write budgets, with baselines before hot-path slices move.
- Codex does not call out `RepairEmptyType` specifically, while Claude and DeepSeek make it the central blocker. Assessment: keep it as a blocker because it is source-verified, violates the design's own repair gate, and has no backlog owner.

**Missing evidence:**
- A real repair slice or explicit assignment of the `RepairEmptyType` fix to an existing slice with sequencing, proof obligations, trace evidence, and doctor rendering tests.
- Complete production call-site inventory for `RepairEmptyType`, including `internal/session/resolve.go` call sites and non-test callers in `cmd/gc`, `internal/api`, and `internal/session`.
- A rule that repair helpers propagate persistence errors and do not mutate in-memory state when the store write fails.
- Guard coverage for repair/doctor files, or a separate executable check proving repair files mutate owned keys only through one audited helper.
- Proof that production `store.List` metadata filtering for `session_name` and `alias` is index-backed, or a revised budget that measures the actual backend behavior.
- Current large-city baseline measurements for target lookup, API path alias fallback, mail recipient resolution, runtime-start repair, close/work-release scans, reconciler fact materialization, subscriber fan-out, and throttle-marker writes.
- A shared query-counting or benchmarking substrate for in-process `beads.Store` calls, replacing ad hoc counting stores.
- A diagnostics manifest mapping operations to required trace site/reason/outcome codes, required fields, doctor/session-inspect checks, API/CLI/log renderers, and exact tests.
- A trace outcome compatibility table covering existing outcomes versus new durable-scan, repair, blocked, deferred, and no-op outcomes, including `gc trace` filter behavior.
- Registered doctor/session-inspect check-name proof for the session checks named in the design, or explicit out-of-scope markers per slice.
- Subscriber fan-out and backpressure tests for slow subscriber, failed subscriber, duplicate event, skipped event, queue overflow, and durable scan retry/convergence.
- A session-scoped operator surface such as `gc trace show --session <id>` or equivalent `gc session inspect <id>` output for blocked and recovered states.
- Budget and routing rules for synchronous reconciler writes such as stranded/degraded throttle markers.
- Source proof or benchmarks for pool snapshot map rebuild behavior and pre-commit re-read query shape.

**Required changes:**
- Add or assign a sequenced repair work slice that consolidates `RepairEmptyType` behind one audited helper. The helper must return persistence errors, avoid in-memory repair on write failure, and emit typed before/after trace evidence plus doctor rendering proof.
- Add `internal/session/resolve.go` to the Slice 0 scan list and enumerate every production `RepairEmptyType` caller in `SESSION_BOUNDARY_SYMBOLS.yaml` or the equivalent inventory artifact.
- Resolve the repair guard contradiction: include doctor/repair files in mutation-guard scanning with expiring allowlist rows, or add a separate guard proving repair writes only flow through the audited helper.
- Add a machine-readable diagnostics manifest and a failing test that requires touched operations to declare trace fields, doctor/session-inspect mappings, rendering surfaces, and exact proof tests or explicit out-of-scope status.
- Add a doctor/session-inspect check registry or golden file for the named session checks, with per-slice implemented/out-of-scope status.
- Replace speculative `missed-event-recovered` terminology with truthful durable convergence outcomes unless durable event-delivery tracking is explicitly added.
- Add a trace outcome compatibility table and renderer/filter tests for existing and new outcome codes.
- Define per-surface query-count budgets before Slice 1 delegates target lookup paths, including direct ID, `session_name`, alias, configured named lookup, API path alias, CLI qualified-alias basename, mail query expansion, and allow-closed lookup.
- Prove metadata-list indexing for the production backend or restate the budget around measured backend behavior.
- Promote a shared counting/benchmark substrate and large-city baseline fixtures to Slice 0 before hot-path behavior moves.
- Define subscriber fan-out and durable-scan diagnostics with idempotency keys, retry cadence, queue/backpressure behavior, and tests for missed, duplicate, slow, and failed reactions.
- Add a reconciler-write budget and routing rule for throttle-marker writes, including whether they become asynchronous, queued, or boundary-command driven.
- Require a session-scoped trace or inspect surface that can answer why a named session is blocked, deferred, repaired, woken, drained, closed, or skipped.
- Pin pre-commit re-reads to cheap single-bead `Get` operations unless a slice explicitly proves and budgets a broader query.
- Add source proof or benchmark requirements for pool snapshot map rebuild behavior under large-city churn.
