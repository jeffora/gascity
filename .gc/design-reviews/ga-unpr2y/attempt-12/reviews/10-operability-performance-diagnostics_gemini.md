# Ingrid Holm — DeepSeek V4 Flash (Independent Review, Attempt 12)

**Verdict:** block

**Review focus:** Decision observability, trace and doctor diagnostics, fact read cost, and event fan-out load. Evaluated against the Attempt 12 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-12/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 12 revision of `internal/session/DESIGN.md` demonstrates exceptional structural discipline. Introducing a machine-readable `DIAGNOSTICS_MANIFEST.yaml` and outcome registries (lines 628, 645, 1483) successfully prevents ad-hoc, untracked command-local string literals from compromising system observability. Furthermore, grounding critical recovery (like close/work-release and drain-cancel) in reconciler-driven scans of durable facts (lines 1130–1135) rather than best-effort in-process event delivery is an outstanding distributed-systems architecture pattern.

However, from the perspective of **Operability, Performance, and Diagnostics (Ingrid Holm)**, the technical plan contains significant structural gaps that will severely limit production debugging, degrade hot-loop latency, and compromise the integrity of the functional-core model:

1. **Pure deciders are opaque and blind by design**, lacking a structured diagnostics schema to return dynamic, rich, operator-actionable context (such as conflicting locks or exact elapsed times) without bloating the immutable facts or violating the ban on flat optional envelopes.
2. **Read-path repair mutations (`RepairEmptyType`) bypass command-applier guards**, mutating in-memory pointers even when database writes fail, which introduces severe split-brain risks and creates unobserved state mutations.
3. **Reconciler fact materialization cost is extremely expensive**, requiring full-store scans and complex aggregations across hundreds of sessions on *every single tick* without a concrete caching, indexing, or incremental-compilation contract.
4. **The `missed-event-recovered` diagnostic outcome is untraceable and speculative**, because idempotent background scans are state-blind and cannot distinguish a missed event from a normal, routine convergence action.

Until these operability, diagnostic, and performance boundaries are physically and structurally secured in the technical contracts, decomposition must remain blocked.

---

## Top Strengths

1. **Centralized Diagnostics Vocabulary and Manifest (Lines 628, 645, 1483):** Transitioning diagnostics from arbitrary, inline log lines to a centralized, machine-readable `DIAGNOSTICS_MANIFEST.yaml` and trace vocabulary registry is an excellent engineering choice. It guarantees that new trace, doctor, and inspect outcomes are systematically registered, preventing diagnostic drift and enabling robust golden-file testing.
2. **Durable Scan-First Recovery Posture (Lines 1130–1135, 1154–1159):** Correctly prioritizing reconciler-owned scans of durable facts as the source of truth for critical reactions (like close/work-release and drain-cancel) ensures that the system remains highly robust and converges correctly even if the in-process event bus suffers from skipped, delayed, or out-of-order event delivery.
3. **Explicit Performance budgets and Test Gates (Lines 1569–1579, 1583–1593):** Enforcing explicit, numeric budget limits on fact reads (e.g., precise queries for target lookup and mail recipient resolution) and backing them with executable, failing-build gates (like `Test.*QueryCount` and `Test.*Backpressure`) provides excellent protection against silent hot-path performance regressions in the reconciler.

---

## Critical Risks & Blockers

### 1. [Blocker] Decider Purity vs. Diagnostic Opacity: Lack of Writable Diagnostic Context in Pure Deciders
The functional-core model mandates that pure session deciders must live in a mechanically guardable file set with zero imports of bead stores, clocks, or loggers, consuming only pre-computed, immutable `SessionFacts` (lines 1311–1315).

While this ensures testing purity (lines 1648–1650), it introduces a critical operability blind spot. When a decider rejects or blocks an operation (e.g., returning a `blocked` or `ambiguous` outcome), an operator needs rich, dynamic, and action-oriented context to diagnose the issue (e.g., the exact IDs of conflicting locks, specific missing dependencies, or precise remaining cooldown/grace durations). 

Under the current design, a decider can only return simple, centralized outcome codes (lines 1457–1470) and a generic `trace reason` string (line 748). If we attempt to pack rich, dynamic diagnostic context into the decider's output, we either:
* Bloat the immutable input `SessionFacts` with speculative pre-computed data that the decider might not even need, violating TR-002 (lines 1639–1640).
* Construct a broad, flat, optional error envelope for `SessionCommandConflict` (line 748), directly violating the design's own strict ban on flat optional envelopes (lines 535–536, 752).

Operators will be left blind, forced to manually parse raw JSONL streams or query raw bead store states to explain why a decision was made.

* **Required Fix:** Introduce a concrete, structured diagnostic metadata schema for decider return values. Specifically, allow deciders to return a `DiagnosticReport` containing strongly-typed, kind-specific diagnostic fields (e.g., `LockConflictReport`, `DependencyGapReport`), with dedicated serialization and rendering rules, rather than relying on a generic `trace reason` string or forcing all fields into a flat `SessionCommandConflict` struct.

---

### 2. [Blocker] Read-Path Repair Mutations (`RepairEmptyType`) Introduce Split-Brain and Static Guard Bypass
The design lists `RepairEmptyType` (W-019, W-024, W-025) as a read-side normalization exception (lines 820, 825, 826). During target lookup, classification, and list filtering, `RepairEmptyType` is executed to write missing type fields back to the store.

This introduces two severe operability and consistency risks:
1. **Split-Brain Memory State:** If the database write fails (e.g., due to lock contention, transient network timeout, or read-only database replica mode), the write error is silently swallowed (e.g., `_ = store.Update(...)`), but the function continues to mutate the in-memory object pointer. This leaves the in-memory session object in a repaired state while the underlying store remains empty/broken, leading to unpredictable, hard-to-debug split-brain behaviors.
2. **Static Guard Bypass:** Because these mutations occur during read-only paths (like target classification or candidate selection), they bypass the command-applier boundary. Simple AST/symbol static guards designed to detect direct store mutations on write paths will fail to catch these hidden "dark writes."

* **Required Fix:** Explicitly forbid `RepairEmptyType` or any read-path normalization helper from mutatively updating in-memory pointers when the underlying store update fails. All repair actions must propagate write errors, and any successful write must be forced to write to a specialized audit trail log in the bead store and emit a critical `session.repair.apply` diagnostic event (line 1481) so that operators have full visibility.

---

### 3. [Major] Severe Fact Compilation Overhead on Reconciler Hot Loops
The design requires the reconciler to assemble demand, pool, work, config, runtime, health, progress, budget, and trace facts from live sources on every tick to feed the eligibility deciders (lines 1255–1257, 1314–1315).

In a large city with hundreds or thousands of active sessions and work items, querying and aggregating these facts from the store on every single reconciler tick is extremely expensive. While the design specifies a large-city performance target of $O(N)$ (lines 1560–1562), it does not provide any caching, indexing, or incremental-materialization contract to restrict the fact reader's CPU and store read latency. Without this, the reconciler loop is highly vulnerable to CPU starvation and database read saturation.

* **Required Fix:** Define a strict caching/indexing contract for fact reader materialization. The fact compiler must leverage durable metadata indexes and incremental caching (e.g., dirty-state flags on sessions/work) so that the reconciler loop does not perform full store scans on every tick. Additionally, add a concrete limit on event fan-out frequency (e.g., event throttling or backpressure budgets) to prevent system-wide saturation.

---

### 4. [Major] The `missed-event-recovered` Outcome is Untraceable and Speculative
The operability contract defines a distinct diagnostic outcome `missed-event-recovered` (line 1467) to be logged when background convergence heals an orphan state after an event was missed.

However, because background scans are designed to be purely idempotent, stateless, and fact-driven (lines 1130–1132, 1275), they simply read the current snapshot of durable facts and align them to the desired state. They have no temporal memory or knowledge of whether an event was successfully published or missed. Attempting to emit `missed-event-recovered` is speculative, as the scanner cannot tell the difference between a routine background convergence tick and a genuine event recovery, leading to false-positive logs that spam operators during normal, healthy ticks.

* **Required Fix:** Downgrade `missed-event-recovered` to a normal converged outcome (e.g., `session_exists_converged` or `convergence_reconciled`) that honestly reflects state-blind convergence rather than speculating on event-delivery failures.

---

## Answers to Persona Questions

### 1. Can an operator explain why a session was blocked, woken, or drained from decider output alone, or does diagnosis require reading raw bead metadata?
**Answer:** Currently, diagnosis still requires reading raw bead metadata. Because deciders are pure functions operating over static snapshots, they lack logging capabilities. If they return a simple outcome code like `blocked` or `blocked-by-hold`, they cannot output the rich, dynamic runtime context (e.g., which specific dependency is blocking it, or how many seconds of quarantine remain) unless that data is crammed into the static `SessionCommandConflict` struct, which violates the ban on flat optional envelopes.

### 2. Do SessionConflict values carry enough context for gc trace and doctor flows to identify the blocker without querying additional state?
**Answer:** No. The proposed `SessionCommandConflict` carries only `operation, session ID, precondition, current projection, retryable flag, trace reason` (line 748). While this is sufficient for basic error mapping, it lacks the structured, kind-specific details (such as conflicting lock holders or exact resource-budget exhaustion states) necessary for `gc trace` and `gc doctor` to diagnose and suggest fixes without re-querying the entire store.

### 3. After health gates and progress facts move into session-owned decisions, does the reconciler still emit enough trace output for operators to diagnose scheduling and budget decisions?
**Answer:** Yes, but only if the reconciler's own trace output is explicitly decoupled and budgeted. The design separates reconciler scheduling decisions from session eligibility (lines 1258–1263). However, because the fact readers are in the reconciler and deciders are in session, if the reconciler does not emit detailed trace logs detailing *why* it did not act on a session's eligibility mask (e.g., because of overall pool capacity limits or provider health-red suppression), operators will struggle to bridge the diagnostic gap between session eligibility and reconciler action.

---

## Consistency Report

* **Pattern Alignment:**
  - Aligns with Takeshi Yamamoto (Decider Purity and Atomicity Enforcer) regarding the purity of deciders. It highlights that keeping deciders pure must not come at the cost of operator visibility, and that read-path repairs (`RepairEmptyType`) are a severe violation of the decider-purity model.
* **Cross-File Integrity:**
  - Audited against `REQUIREMENTS.md`. The design's performance budgets and diagnostics are structurally aligned with the latency thresholds, but the high CPU cost of reconciler-loop fact compilation threatens the system's ability to maintain these thresholds under high-city load.
* **Inter-Reviewer Alignment:**
  - Directly supports Liam Okonkwo (Reconciler Boundary and Fact Isolation Auditor). Liam notes that extracting health and progress gates from the reconciler must not leave the reconciler blind. This review confirms that if fact readers do not have explicit caching and trace logging, the overall system operability will suffer.

---

## Required Changes

1. **Establish a Structured Diagnostic Metadata Schema for Deciders:** 
   - Define a strongly-typed `DiagnosticReport` or kind-specific diagnostic payload contract for decider results (e.g., `EligibilityResult` and `SessionCommandConflict`). This allows deciders to return rich, actionable error metadata without relying on opaque strings or violating the flat-envelope ban.
2. **Protect Against Read-Path Repair Split-Brain:**
   - Modify `RepairEmptyType` to ensure that if the database write fails, it does not mutate the in-memory pointer. Write errors must be propagated to the caller, and all successful writes must emit a centralized, typed `session.repair.apply` trace event with before/after evidence.
3. **Register all `RepairEmptyType` Call Sites in the Inventory:**
   - Update the writer inventory (lines 800–832) to explicitly register all 17 non-test call sites across `cmd/gc`, `internal/api`, and `internal/session` so that they are fully visible to security and boundary audits.
4. **Implement an Incremental Fact-Materialization Contract:**
   - Mandate that the fact compiler leverages caching and dirty-state tracking (e.g., metadata change timestamps on sessions/work) so that the reconciler loop does not perform full store scans on every tick.
5. **De-speculate the Event Recovery Diagnostic Code:**
   - Remove the unprovable `missed-event-recovered` outcome and replace it with a generic convergence outcome (e.g., `session_exists_converged` or `convergence_reconciled`) that honestly reflects state-blind convergence.
6. **Mandate Session-Specific CLI/API Diagnostic Commands:**
   - Require a concrete CLI addition—specifically `gc trace show --session <id>` or `gc session inspect <id>`—to allow operators to easily query and format the detailed, structured diagnostic reports returned by session deciders.

---

## Questions

1. Since the reconciler loop aggregates a massive set of disparate facts, should we introduce a centralized `FactCompiler` component that is independently benchmarked and has its own query-count and latency budgets?
2. To prevent reconciler writes (like `strandedEventEmittedKey` throttle markers) from stalling the hot path, should we mandate that all throttle and diagnostic state writes are deferred to a non-blocking background writer queue?
3. To ensure comprehensive golden-file coverage, should we mandate that every new `TraceReasonCode` must have a corresponding rendering test in `TestSessionReconcilerTrace` before the owning slice is merged?
