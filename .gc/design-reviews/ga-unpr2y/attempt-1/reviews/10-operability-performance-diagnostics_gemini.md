# Ingrid Holm — DeepSeek V4 Flash (Independent Review, Attempt 16)

**Verdict:** block

**Lane:** Decision observability, trace and doctor diagnostics, fact read cost, and event fan-out load. Evaluated against the Attempt 16 iteration of `internal/session/DESIGN.md` (692 lines, "Draft backlog" / "iterate" response), `internal/session/REQUIREMENTS.md`, `internal/session/AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 16 revision of `internal/session/DESIGN.md` makes outstanding structural additions, specifically addressing the Attempt 15 findings with a dedicated `Migration Coexistence And Rollback` plan, explicit schema-level requirements for `BOUNDARY_MATRIX.yaml`, a defined `session.*` event taxonomy, and a formalized structure for hot-path budgets.

However, from the perspective of **operability, performance, and diagnostics**, the design remains a **BLOCK**. While the design documents *what* budgets and manifests must contain, it defers critical performance-safety choices (such as the physical indexing of all-session scans or caching of reconciler facts) to individual implementation slices. Furthermore, it leaves the execution and non-blocking recovery lifecycle of read-path repairs completely undefined, presenting a high risk of operational and availability regressions.

---

## Top Strengths

1. **Clean Read-Path Repair Separation (DESIGN.md:246-252):**
   Prohibiting `RepairEmptyType` from silently executing write side-effects on read-only target classification paths and instead returning a explicit `repair-needed` result kind is a major operability win. It preserves classifier purity and keeps write boundaries clean.
2. **Explicit Trace Mappings in `DIAGNOSTICS_MANIFEST.yaml` (DESIGN.md:561-564):**
   Requiring every diagnostic row to explicitly map to a `gc trace` site/reason/outcome record ensures that shifting logic to pure functions does not create a diagnostic signal blackout. This prevents silent failures and preserves operator-visible traces.
3. **Structured Boundary fresh/stale Rules (DESIGN.md:477-488):**
   Adding explicit freshness, unknown, stale, and provider-error fact-handling requirements to `BOUNDARY_MATRIX.yaml` ensures that the reconciler and session deciders have defined, predictable failure-handling behaviors rather than relying on ad-hoc error propagation.

---

## Critical Risks & Blockers

### 1. [Blocker] Unresolved All-Session Scan Hazard on first-adopter Path (`resolveLiveSessionByPathAlias`)
* **Evidence:** `DESIGN.md:581` lists `resolveLiveSessionByPathAlias` under "Required budget rows before delegation" with:
  > `Decision to index, remove, or keep with explicit scan budget. If kept, budget must name maximum session rows scanned and prove newest-created tiebreaker behavior on a large fixture.`
* **Why it matters:** While the design correctly identifies the hazard, it **fails to resolve it**, deferring the decision to individual implementation slices. `resolveLiveSessionByPathAlias` (invoked on the first-adopter API query path) currently executes an unindexed, full `beads.Store` scan, filtering every session bead in memory by `Title`.
  
  In `BdStore`, this triggers **two process forks (`bd list`)** per invocation. On any large city containing thousands of historical or inactive sessions, this unindexed all-session scan will saturate host CPU and disk I/O. Leaving this choice open means the first-adopter slice can legally ship with an unindexed full scan simply by "budgeting" it, violating our core performance-safety invariants.
* **Required Change:** The design must make a final architectural decision: **either remove path-alias resolution from the first adopter's scope entirely, or mandate a proper index on `Title` before delegation.** Simply budgeting a process-forking all-session scan on an API query hot path is unacceptable.

### 2. [Blocker] Unspecified Execution and non-blocking Lifecycle for `repair-needed` Reads
* **Evidence:** `DESIGN.md:246-252` states that when target classification encounters an empty-type session bead, the read path returns `repair-needed` and delegating the write to a separate, audited repair command.
* **Why it matters:** The design completely avoids specifying **how, when, and by whom the repair command is triggered**.
  1. If the API handler synchronously triggers the repair command, the read path is no longer side-effect-free, which violates the target classifier's fundamental contract (`DESIGN.md:210`).
  2. If the API handler rejects the query and returns a 404/500 until an external cron or human operator triggers the repair, this introduces a severe availability regression for previously readable sessions.
  3. If the repair runs asynchronously in the background, the design fails to specify what prevents a split-brain write race if a mutating command (like `gc wake`) targets that same unrepaired session bead concurrently.
* **Required Change:** Authoritatively define the repair execution lifecycle: specify whether repairs are triggered via non-blocking asynchronous worker queues, reconciler ticks, or background handlers, and detail the concurrency gates that protect the bead while a repair is pending.

### 3. [Blocker] Lack of concrete Caching or batching Mechanisms for Reconciler hot loops
* **Evidence:** `DESIGN.md:582` requires a budget row for Reconciler Fact Compilation:
  > `Store query count, subprocess count, runtime probe count, maximum session/work rows scanned, partial-snapshot behavior, proof command, and owner.`
* **Why it matters:** Every reconciler tick compiles facts by querying the store. Because `BdStore` list queries fork a `bd` subprocess, executing repeated queries per tick in a hot reconciler loop will lead to extreme CPU starvation and process table exhaustion on the host.
  
  Similar to the path-alias scan, the design lists this budget row as a *requirement* but **does not specify any caching, incremental fact compilation, or bulk read mechanism** at the architectural level. Defining a budget does not physically prevent process fork fatigue; the design must provide the structural mechanism to meet that budget.
* **Required Change:** Mandate a specific, shared fact-caching or snapshot-read mechanism (such as an in-memory TTL cache, a bulk session state reader, or an incremental fact compilation adapter) that reconciler ticks must use to prevent process fork fatigue.

### 4. [Major] Inappropriate Mixing of Reconciler-Specific requirements into Session Slice 0
* **Evidence:** `DESIGN.md:173-176` requires that Session Slice 0 (a non-mutating preflight) must repair or owner-retire evidence for:
  > `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007` before a later slice cites those rows.
* **Why it matters:** Under `DESIGN.md:459-465` (the Session/Reconciler split), pool scaling, provider health gates, and progress-aware thresholds are explicitly reconciler/pool behaviors that live *outside* `internal/session`.
  
  Forcing the Session-specific Slice 0 gate to block on, validate, and repair reconciler-specific requirement evidence violates the separation of concerns. It unnecessarily delays the delivery of non-mutating Session Slice 0 code by coupling it to complex, unrelated reconciler state machines.
* **Required Change:** Remove the reconciler-specific requirements (`SESSION-RECON-*`) from the Session Slice 0 entry criteria, moving them to their respective reconciler or pool-focused implementation backlogs.

---

## Answers to Persona Questions

### 1. Can operators explain why a session was blocked, woken, drained, or closed from decider output and trace evidence alone?
* **Answer:** With the Attempt 16 design, **yes, eventually**, but not in the initial slices. The addition of explicit `gc trace` site/reason/outcome mappings in `DIAGNOSTICS_MANIFEST.yaml` (DESIGN.md:561-564) and the structured decision inputs/outputs in `BOUNDARY_MATRIX.yaml` ensure that all transition "whys" are logged. However, because this manifest and its validators are only enforced by Slice 0 tests and do not require active CLI/API rendering proof until the respective behavior-moving slices land, operators will face a period of diagnostic asymmetry during partial adoption.

### 2. What do gc trace, conflicts, and event logs show when a decision is rejected or an event is missed?
* **Answer:** When a decision is rejected, `DIAGNOSTICS_MANIFEST.yaml` (DESIGN.md:164, 556) maps the structured `diagnostic_code` onto a `gc trace` site/reason/outcome record, ensuring the rejection is visible in `gc trace` and `doctor` surfaces.
  
  When an event is missed, the system relies on the **Durable Scan Contract** (DESIGN.md:513, 530). Critical actions (such as work release and close) are designed to converge from durable facts scanned periodically by a "durable scan owner," ensuring recovery even when in-process event delivery is completely lost.

### 3. What is the reconciler cost of materializing facts and emitting subscriber events across a large city?
* **Answer:** Extremely high under current architecture. Because each store query in `BdStore` forks an OS subprocess, compiling facts per-session on every reconciler tick will quickly saturate host resources. While `DESIGN.md:582` mandates a budget row for reconciler fact compilation, the design lacks concrete caching or incremental compilation mechanisms, making the real-world operational cost of fact materialization unsafe for large cities.

---

## Consistency & Parity Report

* **Requirements Alignment:** Under `REQUIREMENTS.md`, exact target resolution precedence must be preserved. While the classifier precedence matrix in `DESIGN.md:221-245` matches this perfectly, the performance penalty of path-alias lookups (`resolveLiveSessionByPathAlias`) introduces a severe operational regression that violates system stability.
* **Reviewer Interlock:** This review directly aligns with **Takeshi Yamamoto's** (Decider Atomicity Enforcer) focus on decider purity and the removal of local wall-clock reads, and **Ravi Krishnamurthy's** (Migration Coexistence Strategist) call for standardizing a global cross-process concurrency primitive instead of deferring it to individual slices.

---

## Required Changes Before Approval

1. **Eliminate Path-Alias Performance Hazard:** Completely remove `resolveLiveSessionByPathAlias` (and Title-based all-session scanning) from the Slice 1 API lookup scope, or mandate a proper database-level index on `Title` before delegation. Do not allow a process-forking all-session scan to be approved via a budget row alone.
2. **Define `repair-needed` Execution Lifecycle:** Detail the non-blocking execution model for the audited repair command when the read path encounters a `repair-needed` result. Specify how the repair is scheduled (e.g. background job, reconciler tick) and how concurrent writes are fenced while a repair is pending.
3. **Mandate Reconciler Fact Caching:** Add a concrete, reusable caching or bulk-reading mechanism to the Reconciler Fact Compilation design to prevent host process-fork exhaustion during ticks.
4. **Decouple Reconciler requirements from Session Slice 0:** Remove `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007` from the Session Slice 0 block list, keeping Slice 0 focused strictly on session-specific boundaries and inventories.

---

## Questions

1. If the API read-only classifier returns `repair-needed` for a session bead, will the API handler block the caller's request while it triggers an asynchronous background repair, or will it return a 404/not-found and depend on an external agent to resolve it?
2. To prevent reconciler hot loops from saturating host resources, can we introduce a bulk session status reader in `beads.Store` that returns all active session statuses in a single query rather than running per-session queries?
3. Why are reconciler-owned pool scaling and health-gate requirements (`SESSION-RECON-*`) placed as blocking gates for Session Slice 0, when they violate the Session/Reconciler separation of concerns?
