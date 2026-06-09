# Ingrid Holm — DeepSeek V4 Flash (Independent Review, Attempt 15)

**Verdict:** block

**Lane:** Decision observability, trace and doctor diagnostics, fact read cost, and event fan-out load. Evaluated against the Attempt 15 iteration of `internal/session/DESIGN.md` (472 lines, "Draft backlog"), `internal/session/REQUIREMENTS.md`, `internal/session/AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 15 revision of `internal/session/DESIGN.md` makes laudable progress by introducing an explicit "Observability And Cost" section and framing the "Event And Recovery Contract." However, as an operability, performance, and diagnostics reviewer, I must **block** this design. 

The current design presents severe, unmitigated risks to production performance and operator-visible diagnostics:
1. It ignores existing unindexed all-session scans on the first-adopter path (`resolveLiveSessionByPathAlias`) that violate its own stated cost rules.
2. It fails to bind pure decider outputs to concrete trace/emission obligations, threatening a silent blackout of operator-visible `gc trace` signals.
3. It allows expensive `bd` subprocess spawning overheads on reconciler hot loops with no caching or incremental compilation.
4. It retains ownerless, error-swallowing read-path write side effects (`RepairEmptyType`) that bypass normal mutation and observability boundaries.

---

## Top Strengths

1. **Explicit "Observability And Cost" Section:** Cementing diagnostic requirements—such as operation ID, result kind, stable reason code, retryability, and missing/stale indicators (`DESIGN.md:343-353`)—as core design invariants represents a massive step forward from previous iterations.
2. **Post-Commit Event Semantics:** Correctly declaring that events are post-commit facts and ensuring that critical convergence (close, work release) runs from durable scans rather than synchronous event fan-out prevents event-log processing latency from polluting the reconciler hot-path.
3. **Vocabularies Managed by Central Manifest:** Placing ownership of trace mappings, doctor output, and adapter-rendering behaviors under `DIAGNOSTICS_MANIFEST.yaml` enables automated verification of diagnostic-coverage parity across surfaces.

---

## Critical Risks & Blockers

### 1. [Blocker] Unindexed All-Session Scans on First-Adopter Path (`resolveLiveSessionByPathAlias`)
* **Evidence:** `internal/api/session_resolution.go:392-427` implements `resolveLiveSessionByPathAlias` which is invoked by the first adopter (API query-side lookup, `DESIGN.md:191`). It calls `session.ListAllSessionBeads(store, beads.ListQuery{})` with an empty query.
* **Why it matters:** `ListAllSessionBeads` issues two full list queries to `beads.Store` (one by type, one by label). In `BdStore`, each list query forks and execs a `bd list` subprocess. This means a single lookup triggers **two subprocess invocations** and scans **every single session bead in the store**, filtering them in-memory in Go by `Title`. On a large city with thousands of historical or inactive sessions, this full scan is extremely slow. Since this resolver is also on background message-delivery hot-paths, high message throughput will trigger process-fork DDOSing on the host machine.
* **Required Change:** Mark `resolveLiveSessionByPathAlias` as a documented all-session scan and enforce a strict query-count/subprocess-count budget with large-city benchmarks in `DIAGNOSTICS_MANIFEST.yaml`. Alternatively, replace the full scan with an indexed metadata query or remove path-alias fallback lookup entirely from the Slice 1 first adopter's scope.

### 2. [Blocker] Diagnostic Signal Blackout: Decider Purity lacks Trace Emission Invariant
* **Evidence:** `DESIGN.md` lines 148-156 declare the classifier and deciders side-effect free (e.g., "no event emission", "no store writes").
* **Why it matters:** When decider logic moves from call-sites to pure functions in `internal/session`, the deciders themselves cannot log, trace, or write to event logs. Unless the design explicitly obligates the *calling adapter* to extract the decider's structured `diagnostic_code` and emit it to the `gc trace` framework at the former call-sites, those critical operator-facing "why did this decision occur" signals will silently vanish. Pure deciders without forced emission contracts guarantee diagnostic opacity.
* **Required Change:** Add a strict emission invariant: extracted deciders/classifiers must keep their former call sites as `gc trace` site/reason/outcome emission points. The calling adapter is explicitly responsible for mapping the decider's structured diagnostic result onto the trace record, ensuring trace detail parity with current production.

### 3. [Blocker] Ownerless, Error-Swallowing Read-Path Repairs (`RepairEmptyType`)
* **Evidence:** `internal/session/resolve.go:218-229` defines `RepairEmptyType` which is invoked on read/resolution paths. It executes `_ = store.Update(b.ID, ...)`—completely swallowing the write error—and mutates the in-memory bead state regardless.
* **Why it matters:** This is a hidden, unowned write side-effect inside read-only resolution flows. Swallowing the persistence error masks storage failures from operators, while patching in-memory state can lead to silent data corruption or unpredictable runtime-split behaviors. It violates the "no-side-effects read path" rule of target classification.
* **Required Change:** Either (a) keep `RepairEmptyType` behind a dedicated, audited repair owner with before/after trace evidence and full error propagation, or (b) have the read-path lookup return `repair-needed` as a diagnostic result kind, delegating the actual write to a separate, controlled repair/doctor command loop rather than performing in-line silent repair.

### 4. [Major] Reconciler Hot-Loop Subprocess Spawning Overhead
* **Evidence:** Reconciler ticks compilation (`cmd/gc/session_reconciler.go`) compiles facts on every loop using `ListAllSessionBeads` and metadata queries.
* **Why it matters:** Because `BdStore.List` executes a `bd list` subprocess, any multi-query resolution inside reconciler sweeps will execute multiple subprocesses per tick. Spawning OS processes repeatedly on every hot loop tick is unsustainable. The design does not specify caching, incremental fact-compilation, or strict subprocess execution caps for the hot path.
* **Required Change:** Establish a strict, non-zero reconciler subprocess execution budget and enforce snapshot/batch reads to prevent hot-loop CPU saturation.

### 5. [Major] Reason Code Disconnection from Canonical Projection Vocabulary
* **Evidence:** `diagnostic_code` on Candidate structs (`DESIGN.md:172`) is described as a stable reason code, but lacks a binding rule to reuse the canonical vocabulary.
* **Why it matters:** `REQUIREMENTS.md` already defines a rich, canonical vocabulary for blockers (e.g., `held`, `quarantined`, `missing-config`, `identity-conflict`, `duplicate-canonical`) and wake causes. If the classifier/decider introduces arbitrary new `diagnostic_code` strings, operators will have to maintain two parallel, conflicting glossaries to understand why a session is blocked or woken.
* **Required Change:** Require that all classifier/decider reason codes and `diagnostic_code`s map directly onto the canonical projection vocabulary from `REQUIREMENTS.md:44-52`. Any new terms must declare their explicit relationship to these base canonical states and blockers.

---

## Answers to Persona Questions

### 1. Which vocabulary types are required by slice 1 target classification versus introduced only for later slices?
* **Answer:** Slice 1 (read-only API lookup) requires only basic targeting vocabularies (`direct-id`, `live-session-name`, `live-alias`, `not-found`, and `repair-needed`). Mutation-related vocabularies (`held`, `quarantined`, `SessionCommandConflict`), runtime intent states, and durable event payloads are introduced only for later slices and must remain strictly marked as `provisional`, forbidden from appearing in Slice 1 codebase or public API schema.

### 2. Does TR-007 future durable-event compatibility shape current APIs more than today's in-process events require?
* **Answer:** No, the Attempt 15 text correctly defers event-sourcing as post-commit facts. However, to ensure future `TR-007` compatibility, the read-only lookup and classification APIs must avoid introducing any direct dependency on in-memory event-log streams, verifying convergence exclusively from durable fact states.

### 3. What stops SessionFacts from becoming a broad facade accumulating every field any decider might want?
* **Answer:** Purity of the decider and the Rule of Repeated Exact Use (`DESIGN.md:383-384`). To make this airtight, the Target Candidate struct must be restructured from a flat optional envelope (which currently violates `DESIGN.md:383`'s prohibition) into strict, per-kind typed structures.

---

## Consistency & Parity Report

* **Requirements Alignment:** Under `REQUIREMENTS.md`, exact target resolution order and precedence rules must be preserved. While the classifier precedence matrix in `DESIGN.md:191` aligns, the performance penalty of path-alias lookups creates an operational regression that threatens system stability.
* **Reviewer Interlock:** This review reinforces Elena Marchetti's (Mutation Boundary Auditor) focus on limiting unowned writes (specifically targeting the silent, swallowed write in `RepairEmptyType`) and Takeshi Yamamoto's (Decider Atomicity Enforcer) focus on pure decider inputs (enforcing the elimination of local clock fallbacks and binding emission call-sites).

---

## Required Changes Before Approval

1. **Mitigate Path-Alias Performance Regression:** Add a cost row in `Observability And Cost` and `DIAGNOSTICS_MANIFEST.yaml` that names the path-alias Title scan as an expensive all-session scan, and mandate a query-count budget or index before first adopter delegation.
2. **Define Trace Emission Contract:** Add an emission invariant requiring adapters/callers to map decider/classifier structured diagnostic output back to former `gc trace` site/reason/outcome records.
3. **Secure or Defer `RepairEmptyType`:** Require that `RepairEmptyType` be moved behind an audited repair owner with error propagation and before/after trace logging, or have the lookup return `repair-needed` and let an external command handle the write. Stop swallowing store update errors.
4. **Enforce Canonical Vocabulary Mapping:** Bind all `diagnostic_code` and reason codes to map onto the existing canonical projection vocabulary in `REQUIREMENTS.md`.
5. **Eliminate Flat Optional Candidate Contradiction:** Refactor the Target Candidate contract (`DESIGN.md:158-173`) to use tagged per-kind structures, resolving the flat optional envelope prohibition.

---

## Questions

1. Will the path-alias resolver receive a proper index on `Title`/path fields, or are we expected to run all-session scans under a measured large-city budget?
2. If `RepairEmptyType` fails to write, why does the code currently patch the in-memory bead anyway, and how do we prevent this split-brain scenario from confusing the calling adapter?
3. Since pure deciders cannot log or trace, who owns the mapping table between the pure decider's diagnostic result and the `gc trace` records—is it defined in `DIAGNOSTICS_MANIFEST.yaml` or inside the caller/adapter package?
