# Kwame Asante — DeepSeek V4 Flash (Independent Review, Attempt 9)

**Verdict:** block

**Review focus:** Minimal vocabulary, facade creep, event-log deferral, and backlog scope control for the Session Boundary Design. Evaluated against the Attempt 9 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-9/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 9 revision of `internal/session/DESIGN.md` introduces a highly detailed `## Attempt 8 Review Response` section (`DESIGN.md:568-611`). Establishing a strict "non-mutating Slice 0 preflight" as the sole schedulable implementation deliverable is an outstanding architectural guardrail for backlog scope control. Forbidding any behavior-moving, target-resolution, repair, or event-routing code until Slice 0 has landed is a major step toward locking down the boundaries.

However, from the perspective of the YAGNI Contract Scope Reviewer, **the design response remains purely rhetorical.** The critical tables, schemas, and structural definitions in the document have not been updated to enforce the boundaries described in the prose response. The design continues to pre-emptively codify active checkpoints for downstream slices, retains flat optional envelopes, permits a massive vocabulary backdoor via event payloads, leaks policy into the raw classifier, and leaves the Slice 0 zero-match rule in a state of logical contradiction.

Until these vocabulary creep and facade-creep issues are physically resolved in the technical design contracts (and not just acknowledged in the review response section), we must sustain a **block** on decomposition.

---

## Top Strengths

1. **Strict Slice 0 Implementation Boundary:** Elevating Slice 0 to a strict preflight that blocks all downstream mutation-owning slices is an exceptional mechanism for backlog scope control. It ensures that the static guard parser, scenario matrices, and vocabulary baselines are established before any code moves.
2. **Prose Realignment on Durable Recovery:** The response correctly treats in-process events as best-effort latency aids and delegates critical convergence guarantees to durable reconciler scans. This directly prevents early read-only phases from being bloated with complex event-sourcing abstractions.
3. **Physical Deliverables for Preflight:** Translating the Slice 0 preflight into a physical set of deliverables (such as `SLICE0_BASELINE.md`, `BOUNDARY_INVENTORY.md`, etc.) is a masterclass in backlog scope control. It ensures that static guards and inventories are fully established before any code is touched.

---

## Critical Risks & Blockers

### 1. [Blocker] Pre-emptive Downstream Vocabulary Checkpointing Violates YAGNI
The "Shared vocabulary checkpoints" table (`DESIGN.md:686-695`) still physically codifies types and fields for downstream slices on equal footing with Slice 1:
- `SessionCommandConflict` (Slice 2)
- `RuntimeStartIntent` (Slice 3)
- `SessionFactEvent` (Slice 4+)

While the prose response in Attempt 9 claims that active next-slice vocabulary is separated from design-only future bounds, leaving these rows in the active checkpoints table creates severe technical ambiguity. If `TestVocabularyCheckpoints` is run as specified, it will either fail on non-existent types or force developers to prematurely define stub Go structs in `internal/session` before their callers exist. This is the definition of vocabulary creep.

* **Required Change:** Physically remove `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent` from the active "Shared vocabulary checkpoints" table. Relocate them to a "Provisional Appendix: Future Slices" table, and explicitly declare that package `session` must contain zero code or Go type definitions for Slices 2–7 during Slice 1.

---

### 2. [Blocker] `committed facts` Backdoor Permits a Single Mega `SessionFacts` Type
Under "Vocabulary And Existing Contracts", the design rightly mandates: *"Avoid one broad `SessionFacts` type. Each slice introduces only the fact family and result vocabulary needed for that operation."* (`DESIGN.md:661-663`).

However, the active `SessionFactEvent` checkpoint (`DESIGN.md:694`) still requires a field named `committed facts`. If the event payload contains a flat, unvalidated block of "committed facts," developers will inevitably bypass compile-time safety by building a single, monolithic struct to satisfy this field, or pass `map[string]any`.

* **Required Change:** Explicitly forbid `committed facts` as a single generic field in `SessionFactEvent`. Specify that each event type (e.g., `events.SessionWoke`) must carry only the specific, minimal, typed subset of slice-specific fields required by its immediate subscribers, preventing any broad state facade from leaking under the event envelope.

---

### 3. [Blocker] Flat `TargetCandidate` and `TargetSelection` Structs are Flat Optional Envelopes
`TargetCandidate` (`DESIGN.md:690`) is specified to carry `kind`, `source surface`, `normalized token`, `session ID/name/alias/config identity`, `status`, `closed flag`, and `conflict reason` in a single flat struct.

This is a flat optional envelope. Since different calling surfaces (API, mail, nudge, assignee normalization) have radically different requirements (e.g., mail send does not care about `status` or `conflict reason`; assignee normalization does not care about `closed flag`), forcing all candidates into this single flat struct violates the YAGNI principle of "no premature abstraction." It pulls fields into the target classification boundary that are completely unrelated to some of the adopting surfaces.

* **Required Change:** Split `TargetCandidate` into specialized, minimal candidate types (e.g., `IdentityCandidate`, `AliasCandidate`, `ConfigCandidate`) or use a discriminated union/tagged result structure in Go. Do not allow a single flat struct to act as the universal target candidate.

---

### 4. [Blocker] Raw Classifier Negative Kinds Leak Adapter-Owned Policy
The design claims that *"policy is adapter-owned post-filtering, not raw classifier behavior"* (`DESIGN.md:599-600`). Yet, the raw classifier's result contract is permitted to return negative kinds like `ordinary-config-target` and `requires-materialization` (`DESIGN.md:816–817`, `932`).

In a pure, read-only token-classification model, whether a target requires materialization depends entirely on the caller's operational context (e.g., extmsg nudge materializes, while mail query does not). If the classifier makes this distinction itself, it must inspect or be configured with the caller's policy. This conflates candidate collection with policy decisions, violating the core design response of Slice 1.

* **Required Change:** The raw classifier must only collect and classify candidates by their physical existence and type in the store (e.g., `session-name`, `alias`, `template`, `config-entry`). The negative kinds and "rejections" (including `ordinary-config-target` and `requires-materialization`) must be evaluated entirely inside the surface-specific policy adapters.

---

### 5. [Blocker] The Slice 0 Zero-Match Contradiction Remains Unresolved in the Technical Plan
The Slice 0 preflight command is specified to run `TestVocabularyCheckpoints` (`DESIGN.md:144`), and the design mandates: *"Every named test in the Slice 0 command must report at least one matched production path, scenario row, vocabulary row, or fixture. A zero-match result is a failure, not a pass."* (`DESIGN.md:147–149`).

However, because the design forbids creating checkpoint rows for new vocabulary "before first caller evidence," and at Slice 0 time no new shared type or first delegated caller exists, the test cannot match any new vocabulary rows.

The design never states whether the nine existing contracts in the "Vocabulary And Existing Contracts" table seed `VOCABULARY_CHECKPOINTS.yaml` as baseline active rows. If they do not, the test will either fail the zero-match rule or force developers to invent premature vocabulary to make it pass.

* **Required Change:** Explicitly state in the design text that Slice 0 seeds `VOCABULARY_CHECKPOINTS.yaml` with the nine existing-contract rows (recording their current real callers and files) as baseline active vocabulary. This resolves the zero-match contradiction and makes `TestVocabularyCheckpoints` executable at Slice 0 without inventing premature vocabulary.

---

## Consistency Report

- **Pattern Alignment:**
  - Audited the "Shared vocabulary checkpoints" against active Slice 1 boundaries. Notice how Liam Okonkwo's independent review (Liam Okonkwo — DeepSeek V4 Flash) sustains a strict block because `lifecycle_projection.go` falls back to `time.Now()` (breaking decider determinism) and `compute_awake_set.go` directly computes min-active pool constraints. These violations are the direct result of not having hard, compiler-enforced boundaries. If we allow downstream vocabulary like `RuntimeStartIntent` to remain in the active technical tables, we invite even more leakage of runtime intents into prepare-time deciders (as Liam Okonkwo's blocker regarding `session_key` straddling prepare-vs-observe shows!).
- **Cross-File Integrity:**
  - `REQUIREMENTS.md` vs `DESIGN.md`. Note that `REQUIREMENTS.md` defines specific base states and wake causes, but `DESIGN.md`'s checkpoints do not enforce compile-time mappings for these, leaving a large gap for pattern drift.
- **Inter-Reviewer Alignment:**
  - Our blockers directly support the findings of the **Reconciler Runtime Fact Reviewer (Liam Okonkwo)**. Specifically, Liam flags that pure deciders are violating isolation by fallback to `time.Now()` and direct evaluation of min-active pool constraints. These leaks occur precisely because the vocabulary boundaries and checkpoints are not strictly separated, proving the need to lock down the vocabulary checkpoints and writer boundaries before decomposition.

---

## Answers to Persona Questions

1. **Which vocabulary types are required by slice 1 target classification versus introduced only for later slices?**
   - *Answer*: Only `TargetCandidate` and `TargetSelection` are required for Slice 1. Slices 2–4 types (`SessionCommandConflict`, `RuntimeStartIntent`, `SessionFactEvent`) must be stripped from the active technical design table and moved to a provisional appendix.

2. **Does TR-007's future durable-event compatibility shape current APIs more than today's in-process events require?**
   - *Answer*: Yes. The inclusion of `committed facts` in `SessionFactEvent` represents a major design leak that forces the event payload to carry a broad state facade.

3. **What stops SessionFacts from becoming a broad facade accumulating every field any decider might want?**
   - *Answer*: The explicit instruction to avoid one broad `SessionFacts` type. However, flat optional envelopes in `TargetCandidate` and `SessionFactEvent`'s `committed facts` threaten to bypass this guardrail.

---

## Required Changes

1. **Remove Downstream Vocabulary Checkpoints:** Move `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent` from `DESIGN.md:686-695` to a "Future Slices (Provisional)" appendix.
2. **Forbid `committed facts` Backdoor:** Explicitly forbid `committed facts` in `SessionFactEvent`. Require each event to carry only slice-specific fact families.
3. **Decompose `TargetCandidate` Struct:** Split `TargetCandidate` into specialized candidate types (e.g., `IdentityCandidate`, `AliasCandidate`) or use a discriminated union in Go to avoid a flat optional envelope.
4. **Move Negative-Kind Evaluation to Adapters:** Remove `ordinary-config-target` and `requires-materialization` from the raw classifier's negative-kind vocabulary. Let surface-specific adapters apply policy to raw candidate outputs.
5. **Seed existing-contract rows as Active Vocabulary:** Explicitly declare that Slice 0 seeds `VOCABULARY_CHECKPOINTS.yaml` with the nine existing-contract rows from `DESIGN.md:665–674` as baseline active vocabulary.

---

## Questions

1. If `TargetSelection` is adapter-owned, should the classifier itself be entirely private to the `internal/session` package, exposing only the public surface adapters?
2. Since `SESSION-RECON-001` (Worker boundary) is critical for enforcing that all CLI calls route through `worker.Handle`, should its validation be integrated into Slice 0's static-guard tests (`TestSessionBoundaryGuard`)?
3. Should the "Per-event reaction matrix" be moved entirely to Slice 4 (Event migration) to reduce the upfront cognitive load of Slice 1?
