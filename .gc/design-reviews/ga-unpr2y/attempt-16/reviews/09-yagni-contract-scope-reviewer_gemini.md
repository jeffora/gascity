# Kwame Asante - DeepSeek V4 Flash (Independent Review, Attempt 16)

**Verdict:** block

**Lane:** Minimal vocabulary, facade creep, event-log deferral, and backlog scope control. This reviews the Attempt 16 iteration of `internal/session/DESIGN.md` (692 lines, "Draft backlog") against `REQUIREMENTS.md`, `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 16 iteration of the `DESIGN.md` maintains the 692-line structure introduced at the end of Attempt 15. While the introduction of the universal-vs-per-slice preflight split (DESIGN.md:143-149) is a conceptually sound compromise, a rigorous evaluation from the YAGNI perspective reveals that the design continues to suffer from **speculative front-loading** and **direct internal contradictions**. 

The design establishes excellent control mechanisms—such as the four-state Vocabulary Lifecycle (DESIGN.md:587-595) and the absolute prohibition on flat optional envelopes (DESIGN.md:598-599). However, it immediately violates these very rules in its own first-adopter contracts (Slices 0 and 1). Specifically, the Target Classification result schema is designed as a flat optional envelope containing fields that the read-only first adopter can neither populate nor consume. Furthermore, Universal Slice 0 continues to require the complete front-loading of mutation-specific registries, command appliers, and boundary matrix rows before a single mutating slice is authorized.

As the gatekeeper of the YAGNI principle, I must **block** this design. We cannot approve an architectural plan that violates its own structural invariants in its very first implementation step.

---

## Top Strengths

- **Enforceable Vocabulary Lifecycle (DESIGN.md:587-599):** The division of vocabulary into `documented`, `private`, `provisional`, and `delegating` states is a premier control pattern. Correctly classifying future concepts (such as command results, runtime intents, and session fact events) as `provisional` prevents them from prematurely leaking into public interfaces, client code, or event payloads.
- **Explicit Deferral of Event Sourcing (DESIGN.md:499-502, 617):** Defining session events strictly as post-commit facts rather than a primary storage substrate or coordination mechanism prevents early slices from being crushed by speculative infrastructure. Restricting Slice 0 to a best-effort event inventory (DESIGN.md:521-528) and requiring durable-scan recovery for critical convergence (DESIGN.md:513-520) is a highly mature YAGNI-aligned decision.
- **Micro-Slicing Refactor Loop (DESIGN.md:638-650):** The requirement to isolate a single caller and a single behavior cluster, prove parity via characterization tests, and only then delegate ensures that the refactoring remains manageable and avoids the high risks associated with large, monolithic migrations.

---

## Critical Risks & Blockers

### 1. [Blocker] Self-Contradiction: Target Classifier Result is a Prohibited Flat Optional Envelope
* **Evidence:** `DESIGN.md` lines 253-264 defines the `Typed result schema` with fields like `match_vectors[]`, `bead_state`, `config_state`, `diagnostics`, and `terminal_error` residing concurrently on the same flat struct. This is a direct violation of the design's own rule on lines 598-599: 
  > *"Flat optional envelopes are not acceptable for new shared types; use tagged result kinds or per-kind structs when only some fields are meaningful."*
* **Why it matters:** If the first concrete data structure introduced by the refactor is a flat optional envelope, it sets a bad precedent that subverts the type system from day one. Under this schema, a `not-found` result or a `store-error` result will carry empty, meaningless structs for `bead_state` and `config_state`, forcing downstream consumers to write unsafe nil-checks and duplicate conditional logic.
* **Required Change:** Resolve the contradiction. Either refactor the Target Classification contract to define a strict set of per-kind result structs (e.g., `SelectedResult`, `NotFoundResult`, `RepairNeededResult`) or explicitly mandate in the contract that the production Go implementation must use a tagged union or distinct, per-kind typed structures instead of a single flat struct.

### 2. [Blocker] Speculative Vocabulary Front-Loading in Slice 1 Schema
* **Evidence:** The `Typed result schema` (DESIGN.md:253-264) contains several sub-fields that the read-only API query first-adopter precedence (DESIGN.md:220-244) never exercises. Specifically:
  - `bead_state.labels` and `bead_state.lifecycle state` are not consumed by any query precedence rule.
  - `config_state.materialization allowed flag` is redundant beside the existing `reserved-unmaterialized` and `config-orphan` flags.
  - `diagnostics.stale or partial fact marker` represents runtime, reconciler, and command-level vocabulary that a pure, read-only memory classifier cannot compute.
* **Why it matters:** Landing these fields in Slice 1 violates the rule on lines 593-594 that provisional fields *"cannot appear in public API, generated clients, event payloads, or cross-slice contracts."* It forces developers to stub out speculative metadata and diagnostics fields before the corresponding reconciler health or command slices are even designed.
* **Required Change:** Explicitly designate all sub-fields of the result schema that are not strictly required to resolve the read-only API lookup as `provisional`. Declare that they must not be present in the active Go structs or OpenAPI types exported in Slice 1.

### 3. [Blocker] Premature Artifact Inflation in Universal Slice 0
* **Evidence:** `DESIGN.md` lines 150-165 requires the creation and validation of `COMMAND_APPLIERS.yaml` (inventory of command writers), the mutating rows of `BOUNDARY_MATRIX.yaml` (covering drain, repair, and destructive actions), and `WORKER_BOUNDARY_EXCEPTIONS.yaml` (covering mutating-lifecycle exceptions) as part of Universal Slice 0.
* **Why it matters:** Slice 1 is entirely read-only and mutates nothing (DESIGN.md:210-218). Forcing the team to write complete inventories of mutating commands, multi-process mutation locks, and exception lists before a single mutating slice has been approved violates backlog scope control and the Rule of Repeated Exact Use (DESIGN.md:383-384 in Attempt 15 notes, and L165). It turns Slice 0 into a massive upfront design program, delaying the delivery of real, proven value.
* **Required Change:** Apply the universal-vs-per-slice split to the artifacts themselves. Move `COMMAND_APPLIERS.yaml`, the mutating rows of `BOUNDARY_MATRIX.yaml`, and `WORKER_BOUNDARY_EXCEPTIONS.yaml` out of Universal Slice 0 and into the per-slice preflights of the specific mutation-moving slices (Slices 3, 4, and 5) that actually implement those writes.

### 4. [Major] Cross-Document Inconsistency: Missing TR-007 in Requirements
* **Evidence:** `DESIGN_REVIEW_NOTES.md` line 1823 contains an explicit, normative-sounding section `### TR-007: Keep The Event Log Path Open`. However, no corresponding `TR-007` requirement row exists in `REQUIREMENTS.md` or `DESIGN.md`.
* **Why it matters:** Because `DESIGN_REVIEW_NOTES.md` is declared non-normative (DESIGN.md:10), any load-bearing requirement regarding durable-event compatibility must reside in `REQUIREMENTS.md` or `DESIGN.md`. Leaving TR-007 as a disconnected, un-indexed brief topic creates cross-document confusion and leaves the engineering team without a clear test or verification criteria for event-log compatibility.
* **Required Change:** Either elevate `TR-007` to a formal requirement row in `REQUIREMENTS.md` (e.g., `SESSION-EVENT-007`) with explicit verification selectors, or explicitly state in the `DESIGN.md` Non-Goals that durable event-log compatibility is not a current constraint, and that in-process best-effort event delivery plus offline recovery scans is the sole target architecture.

---

## Answers to Persona Questions

### 1. Which vocabulary types are required by slice 1 target classification versus introduced only for later slices?
* **Answer:** Slice 1 (read-only API target lookup) only requires types representing the search input token, the resolved session ID, and basic result kinds (`selected`, `not-found`, `ambiguous`, `rejected`, `repair-needed`, `store-error`). Enums and fields representing mutation conflicts (`SessionCommandConflict`), runtime intents (`RuntimeStartIntent`), or durable session events (`SessionFactEvent`) are completely out of scope and must not be defined or stubbed.

### 2. Does TR-007 future durable-event compatibility shape current APIs more than today's in-process events require?
* **Answer:** Not in the current text. Because the design correctly treats events as post-commit facts and relies on durable-scan recovery for critical operations, the API is insulated from the speculative complexity of a durable event log. This insulation must be strictly preserved; no early slice should introduce event-sourcing schemas.

### 3. What stops SessionFacts from becoming a broad facade accumulating every field any decider might want?
* **Answer:** The primary barrier is the explicit prohibition of a single large `SessionFacts` struct (DESIGN.md:615) and the requirement that deciders remain small and operation-specific (DESIGN.md:603-611). To fully secure this boundary, we must also break up the Target Classifier's flat result struct into per-kind result types so that we do not build a "TargetFacts" facade by another name.

---

## Required Changes Before Approval

1. **Eliminate Flat Struct Contradiction:** Refactor the Target Classification Contract (DESIGN.md:253-264) to require per-kind result structures (e.g. tagged structs / union kinds) in the production Go implementation, resolving the contradiction with the flat-envelope ban on line 599.
2. **Prune Speculative Fields from Slice 1:** Mark all sub-fields in the `Typed result schema` that are not strictly used by the read-only lookup precedence (such as `bead_state.labels`, `bead_state.lifecycle state`, `materialization allowed flag`, and `stale or partial fact marker`) as `provisional`, forbidding their implementation in Slice 1.
3. **Decompress Universal Slice 0 Scope:** Move `COMMAND_APPLIERS.yaml`, `WORKER_BOUNDARY_EXCEPTIONS.yaml`, and the mutating/destructive matrix rows of `BOUNDARY_MATRIX.yaml` out of the universal baseline and into the per-slice preflights for Slices 3-5.
4. **Reconcile TR-007:** Explicitly move the durable event log rule (TR-007) from `DESIGN_REVIEW_NOTES.md` into `REQUIREMENTS.md` as an audited row, or formally declare it a Non-Goal in `DESIGN.md` to prevent un-testable design requirements.

---

## Questions

1. Since `DESIGN.md` explicitly forbids flat optional envelopes (line 599), why does the Target Classifier contract (lines 253-264) specify a flat multi-field struct where the majority of fields will be nil depending on the `result_kind`?
2. Can we explicitly restrict the active codebase in Slice 1 to implement only the `api-query` surface, treating all other 8 surfaces in the matrix as purely historical or provisional design boundaries?
3. What is the explicit verification path for TR-007, and should it be captured as a formal behavior row in `REQUIREMENTS.md` to ensure it is auditable?
