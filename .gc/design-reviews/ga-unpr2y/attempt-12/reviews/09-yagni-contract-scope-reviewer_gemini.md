# Kwame Asante — DeepSeek V4 Flash (Independent Review, Attempt 12)

**Verdict:** block

**Review focus:** Minimal vocabulary, facade creep, event-log deferral, and backlog scope control for the Session Boundary Design. Evaluated against the Attempt 12 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-12/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 12 revision of `internal/session/DESIGN.md` continues to showcase strong structural intent. Transitioning the non-mutating "Slice 0" preflight into a single, concrete backlog item (lines 623–632) with dedicated artifacts and an explicit zero-match prevention clause is a superb exercise in backlog scope control. Additionally, anchoring critical recovery in reconciler-driven scans of durable facts rather than best-effort in-process events is an outstanding distributed-systems design pattern.

However, from the perspective of the **YAGNI Contract Scope Reviewer**, the technical plan still suffers from speculative design, early vocabulary creep, and logical contradictions that will compromise decomposition. Downstream vocabulary checkpoints remain codified in active metadata tables without first caller evidence. Generic, unconstrained payload fields act as backdoors for monolithic state facades. The core target classification structure is designed as a flat optional envelope. And the mandatory Slice 0 proof command contains a logical contradiction that will prevent it from passing.

Until these vocabulary and scope boundaries are physically tightened in the technical contracts, decomposition must remain blocked.

---

## Top Strengths

1. **Disciplined Single-Item Backlog for Slice 0:** Consolidating the "Slice 0" preflight into a single, concrete backlog item with a pinned artifact checklist and a unified proof command (lines 623–632) represents an exceptional approach to backlog scope control. It guarantees that guards and baseline inventories land before any mutating logic is touched.
2. **Scan-First Crash Recovery Posture:** Correctly deferring critical recovery to reconciler-owned durable fact scans (lines 641–642) rather than speculative event sourcing prevents early implementation slices from being burdened with premature message-bus complexity.
3. **Prose Separation of Classifier and Adapters:** Mandating that the raw target classifier is a read-only, policy-free evidence collector (lines 952–954) ensures clean separation of concerns and avoids leaking surface-specific decisions into core session logic.

---

## Critical Risks & Blockers

### 1. [Blocker] Downstream Vocabulary Checkpoint Leakage Violates YAGNI
The "Shared vocabulary checkpoints" table (lines 744–750) still physically lists metadata and types that belong strictly to downstream slices:
* `SessionCommandConflict` (Slice 2 - Explicit Wake)
* `RuntimeStartIntent` (Slice 3 - Runtime Start)
* `SessionFactEvent` (Slice 4+ - Events)

Codifying these downstream types now violates the YAGNI principle of "no premature abstraction." It invites developers to implement stub Go structs inside the core `internal/session` package before their first delegated callers are created. This compromises the progressive capability model and leaks downstream vocabulary into Slice 1.

* **Required Fix:** Physically remove `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent` from the active "Shared vocabulary checkpoints" table. Relocate them to a "Provisional Appendix: Future Slices" section, and explicitly forbid defining any corresponding Go types in the codebase during Slice 1.

---

### 2. [Blocker] Generic `committed facts` Payload is a Facade-Creep Backdoor
The proposed checkpoint for `SessionFactEvent` (line 750) requires a generic field named `committed facts`. 

This is a major facade-creep vector. A flat, unconstrained `committed facts` block allows arbitrary internal session states to leak onto the event bus. To satisfy this field, developers will inevitably bypass compile-time safety by constructing a single monolithic state facade (such as a flat `SessionFacts` struct or a `map[string]any`). This directly contradicts the design's own ban on flat optional envelopes and broad facades.

* **Required Fix:** Explicitly forbid `committed facts` as an open-ended generic field in `SessionFactEvent`. Mandate that each discrete event payload (e.g., `events.SessionWoke`) must carry only the specific, minimal, strongly typed fields required by its immediate, registered subscribers.

---

### 3. [Blocker] `TargetCandidate` is a Flat Optional Envelope
The design specifies `TargetCandidate` (line 746) to carry `kind, source surface, normalized token, session ID/name/alias/config identity, status, closed flag, conflict reason` in a single shared struct.

This is a flat optional envelope. Disparate calling surfaces have completely distinct requirements—for instance, a simple read-only mail query does not care about `status` or `conflict reason`, while assignee normalization does not care about `closed flag`. Packing all of these fields into a single struct violates the YAGNI rule against premature abstraction. It forces irrelevant properties into the target classification boundary for simpler surfaces.

* **Required Fix:** Decompose `TargetCandidate` into specialized, minimal candidate types (e.g., `IdentityCandidate`, `AliasCandidate`, `ConfigCandidate`) or specify a discriminated tag-union structure in Go. Do not permit a single flat optional struct to serve as the universal candidate representation.

---

### 4. [Blocker] Raw Classifier Negative Kinds Conflate Policy with Classification
The raw classifier result contract (lines 971–979) is permitted to return negative kinds such as `ordinary-config-target` and `requires-materialization`.

This is a severe policy leak. Determining whether a target is an "ordinary config target" or "requires materialization" is not a read-only classification fact; it depends entirely on the calling surface's operational policy (e.g., `nudge` materializes configured named sessions on a miss, whereas mail query does not). If the classifier evaluates these kinds itself, it must be configured with surface policy. This directly violates the core design response that "policy is adapter-owned post-filtering, not raw classifier behavior."

* **Required Fix:** Remove `ordinary-config-target` and `requires-materialization` from the raw classifier's negative-kind vocabulary. The raw classifier must only collect and return physical, store-derived facts (e.g., `session-name-match`, `alias-match`, `config-entry-match`). The evaluation of materialization eligibility or rejection must occur exclusively inside surface-specific policy adapters.

---

### 5. [Blocker] Slice 0 Zero-Match Logical Contradiction Blocks Preflight Execution
The technical plan specifies that `TestVocabularyCheckpoints` is an exit gate and part of the Slice 0 proof command (lines 144, 630). It also mandates that every test in the Slice 0 command must report at least one matched case, and that a zero-match result is a failure (lines 147–149, 629).

However, because the design forbids creating checkpoint rows for new vocabulary before first caller evidence, and at Slice 0 time no new shared types or first delegated callers exist, `TestVocabularyCheckpoints` will have zero new types to verify. This creates a logical contradiction that causes the mandatory Slice 0 preflight command to fail closed.

* **Required Fix:** Explicitly state in the design text that Slice 0 seeds `VOCABULARY_CHECKPOINTS.yaml` with the nine existing-contract rows listed in the "Vocabulary And Existing Contracts" table (lines 721–731), mapping them to their current real callers and files. This provides the test with baseline active rows to match and resolves the zero-match preflight contradiction.

---

### 6. [Major] Slice 0 Nomenclature Ambiguity with Traceability Matrix Row 0
The Traceability Matrix still defines row `0 Transition reducer baseline` (line 1015) as a standing state-machine constraint. 

This creates naming ambiguity with the preflight backlog item, which is also identified as "Slice 0" (line 618). Because the entire scheduling and dependency-injection gate in the backlog is named-based ("must declare a dependency on the closed Slice 0 bead"), this overlap invites decomposition errors where implementers bind gates to the standing state-machine constraint rather than the preflight guard bead.

* **Required Fix:** Rename the Traceability Matrix row `0 Transition reducer baseline` to a non-slice identifier, such as `Reducer Oracle Baseline`, reserving the "Slice 0" designation exclusively for the preflight backlog item.

---

### 7. [Major] Unused Policy Flags Leaked into Slice 1
The technical plan lists nine policy fields (lines 955–958) to be applied by compatibility adapters in Slice 1. 

However, at least one flag—`allow_template_factory`—has no active setter or consumer in the Slice 1 adoption plan. API/Huma live session targets explicitly reject `template:` factory strings, and the surface that genuinely accepts them (CLI dispatch via `parseTemplateTarget`) is completely out of scope for Slice 1. Landing this policy flag now violates the "unused vocabulary lands before any slice needs it" red flag.

* **Required Fix:** Remove `allow_template_factory` (and any other policy flag lacking an active setting surface in the Slice 1 adoption plan) from the technical contract. Reintroduce them only when their adopting surfaces enter scope in downstream slices.

---

## Answers to Persona Questions

### 1. Which vocabulary types are required by slice 1 target classification versus introduced only for later slices?
**Answer:** Slice 1 target classification only requires `TargetCandidate` (and its specialized sub-types) and `TargetSelection`. These types are sufficient for read-only candidate collection and surface-specific adapter selection. In contrast, `SessionCommandConflict` (Slice 2), `RuntimeStartIntent` (Slice 3), and `SessionFactEvent` (Slice 4+) are introduced prematurely. They serve only downstream mutating operations and violate YAGNI at Slice 1 introduction.

### 2. Does TR-007 future durable-event compatibility shape current APIs more than today's in-process events require?
**Answer:** Yes, it does. By codifying a broad, generic `committed facts` field on the proposed `SessionFactEvent` struct, the design shapes the current API to accommodate speculative durable event-sourcing. This future-proofing forces an open-ended payload envelope that leaks internal session state on the wire, bending today's API shape against the "no premature abstraction" and "no broad facade" rules.

### 3. What stops SessionFacts from becoming a broad facade accumulating every field any decider might want?
**Answer:** Today, nothing in the code base prevents this. While the design response offers a strong rhetorical guideline against a broad `SessionFacts` type, the guidelines remain advisory. Unless this boundary is mechanically enforced by physically removing downstream vocabulary from the active checkpoints table and wrapping the static AST checks (`TestVocabularyCheckpoints`) around strict "rule-of-two" validation, developers will naturally construct a monolithic `SessionFacts` struct to simplify state-gathering across deciders.

---

## Consistency Report

* **Pattern Alignment:**
  - Aligns with Elena Marchetti's (Mutation Boundary Auditor) requirement for a rigid, non-mutating Slice 0 preflight. Both reviews highlight that leaving critical guardrails as pure prose or logical contradictions is a severe project risk.
* **Cross-File Integrity:**
  - Audited against `REQUIREMENTS.md`. The design's classification and adapter contracts are compatible with the resolution precedence defined in `SESSION-ID-003`, but the raw classifier's inclusion of negative kinds like `requires-materialization` violates the global invariant that product policy must be separated from raw classification.
* **Inter-Reviewer Alignment:**
  - Directly supports the blockers raised by Liam Okonkwo (Reconciler Runtime Fact Reviewer). Liam highlights that pure deciders are violating isolation by falling back to local clock reads and evaluating pool demand directly. These leaks occur because the vocabulary boundaries and checkpoints are not strictly separated, proving the need to lock down the checkpoints and remove provisional next-slice vocabulary before any code is generated.

---

## Required Changes

1. **Move Downstream Checkpoints to Appendix:** Physically relocate `SessionCommandConflict`, `RuntimeStartIntent`, and `SessionFactEvent` from the active "Shared vocabulary checkpoints" table (lines 744–750) to a "Provisional Appendix: Future Slices" section.
2. **Forbid Generic `committed facts` Payload:** Remove `committed facts` from `SessionFactEvent`. Mandate that each discrete event payload carry only the specific, minimal, typed subset of fields required by its immediate subscribers.
3. **Decompose `TargetCandidate` Struct:** Decompose `TargetCandidate` into specialized, minimal candidate types (e.g., `IdentityCandidate`, `AliasCandidate`) to avoid creating a flat optional envelope.
4. **Remove Policy Decisions from Raw Classifier:** Strip `ordinary-config-target` and `requires-materialization` from the raw classifier's negative-kind vocabulary. Ensure the raw classifier returns only store-derived facts, leaving policy rejections to the adapters.
5. **Establish Seeding Contract for Slice 0:** Explicitly state that Slice 0 seeds `VOCABULARY_CHECKPOINTS.yaml` with the nine existing-contract rows from lines 721–731 to resolve the zero-match test contradiction.
6. **Disambiguate "Slice 0" Nomenclature:** Rename Traceability Matrix row `0 Transition reducer baseline` to `Reducer Oracle Baseline` to reserve "Slice 0" exclusively for the preflight backlog item.
7. **Prune Unused Policy Flags:** Remove `allow_template_factory` from the policy flag list until its adopting surface (CLI dispatch) enters the adoption plan.

---

## Questions

1. Since `TargetSelection` is adapter-owned, should the raw classifier itself be entirely private to the `internal/session` package, exposing only the public surface adapters to external callers?
2. Should `SESSION-RECON-001` (Worker boundary) checks be integrated directly into Slice 0's static-guard tests (`TestSessionBoundaryGuard`) to enforce that all CLI commands route through the allowed worker boundary?
3. Should the "Per-event reaction matrix" be entirely removed from the DESIGN document and deferred to Slice 4 (Event migration) to reduce the upfront cognitive load of Slice 1?
