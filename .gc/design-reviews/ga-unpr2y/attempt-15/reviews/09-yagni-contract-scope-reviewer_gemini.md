# Kwame Asante — DeepSeek V4 Flash (Independent Review, Attempt 15)

**Verdict:** block

**Lane:** Minimal vocabulary, facade creep, event-log deferral, and backlog scope control. Evaluated against the Attempt 15 iteration of `internal/session/DESIGN.md` (472 lines, "Draft backlog"), `internal/session/REQUIREMENTS.md`, `internal/session/AGENTS.md`, and the active checkout source.

---

## Overview

While the previous radical simplification to a lean ~137-line design was a massive win for the YAGNI principle, this 472-line Attempt 15 revision represents a severe and highly visible **review-driven re-inflation regression**. 

Under pressure from other review lanes, the design has grown a massive, speculative preflight "Slice 0" that demands 13 distinct machine-readable and human-readable artifacts (DESIGN.md:108-121) before a single line of behavior-moving code can land. Furthermore, the Target Classification Contract specifies a full 14-kind target taxonomy and covers 9 surfaces (8 of which are "characterization only") before Slice 1 (a simple, read-only API query lookup) even has an implementation to validate it.

This is the exact definition of **YAGNI failure**: creating complex upfront specifications and registries for future slices before one concrete, narrow slice is built to prove the pattern. 

Therefore, I must **block** this design. The design contains excellent internal control mechanisms (such as the Vocabulary Lifecycle and the Rule of Repeated Exact Use) but fails to apply them to its own newly added contracts.

---

## Top Strengths

1. **Explicit Deferral of Event Sourcing:** Keeping event sourcing out of the initial implementation steps (Non-Goals, DESIGN.md:467, "Do not make event sourcing the first implementation step") and declaring events as post-commit facts (DESIGN.md:317) is a crucial YAGNI guard that prevents massive speculative overhead in the first few slices.
2. **Vocabulary Lifecycle States:** The four-state vocabulary lifecycle (`documented`, `private`, `provisional`, `delegating` at DESIGN.md:370-384) is a brilliant, enforceable governance pattern. It correctly declares that future terms must stay `provisional` until a production caller proves the exact field set.
3. **Incremental Refactor Scoping:** Refactor Rules 5, 6, and 7 (DESIGN.md:431-435) safely bound the refactoring process to one current behavior cluster and one caller at a time, protecting the workspace from risky, monolithic flag-day migrations.

---

## Critical Risks & Blockers

### 1. [Blocker] Slice 0 Speculative Artifact Bloat and Pre-Execution Gates
* **Evidence:** `DESIGN.md` lines 108-121, requiring 13 complex artifacts including `COMMAND_APPLIERS.yaml` covering wake, close, retire, drain, runtime start, rollback, and repair, plus `BOUNDARY_MATRIX.yaml` covering all operations, before Slice 1 can begin.
* **Why it matters:** Slice 1 is a strictly read-only API query session target lookup. It has zero mutations, zero events, zero worker boundary bypasses, and zero provider runtime calls. Forcing the team to define detailed command contracts, rollback rules, and atomic appliers for wake, close, drain, and runtime start *before* building a simple read-only query classifier completely violates "one operation at a time." It forces speculative upfront design of complex mutation states before a single read-only helper is proven in production.
* **Required Change:** Split Slice 0 into a Universal Preflight (baseline, harness, source/writer inventories, scenario parity) and Per-Slice Preflights. `COMMAND_APPLIERS.yaml`, `BOUNDARY_MATRIX.yaml`, and other mutation-focused artifacts must be moved to the entry gates of the specific mutation-moving slices (Slices 3, 4, 5) that actually introduce those mutations.

### 2. [Blocker] The Target Classification Taxonomy is Speculative and Un-Gated
* **Evidence:** `DESIGN.md` lines 158-173, specifying 14 `kind` values and 9 `source_surface` values (including mail, extmsg, assignee, nudge, attach, transcript, and CLI) for the first behavior extraction.
* **Why it matters:** The design disclaims that all surfaces except `api-query` are "characterization only" (lines 189-196) and that future vocabulary stays `provisional` until a production caller proves it (line 381). Yet, the Target Classification Contract immediately activates a huge multi-surface taxonomy. This invites implementers to build a massive, complex classifier supporting all 14 kinds and 9 surfaces upfront, undoing the reset.
* **Required Change:** Explicitly mark all non-`api-query` surfaces, and any `kind`s or fields not strictly required by the read-only API lookup, as `provisional` in the Target Classification Contract. State that they must not be implemented in active code or exported types until those specific surfaces delegate in later slices.

### 3. [Blocker] Target Candidate Struct is a Prohibited Flat Optional Envelope
* **Evidence:** `DESIGN.md` lines 158-173 defines a single 12-field `candidate` struct where fields like `session_id`, `config_identity`, `alias`, `conflict_group`, and `retryable` are optional and only populated under specific discriminants. This directly contradicts line 383: *"Flat optional envelopes are not acceptable for new shared types; use tagged result kinds or per-kind structs when only some fields are meaningful."*
* **Why it matters:** Allowing this flat candidate structure to land as a shared type in Slice 1 reinstates the exact `SessionFacts` facade risk that the reset deleted. It allows a single struct to accumulate miscellaneous fields, which will inevitably grow as more surfaces are added.
* **Required Change:** Resolve this internal contradiction. Restructure the classification result as tagged per-kind structs / unions, or explicitly state that the candidate table in `DESIGN.md` is a provisional census and the production Go implementation must use strict, per-kind typed structures.

### 4. [Major] Diagnostic Manifest Speculative Vocabulary Registration
* **Evidence:** `DESIGN.md` lines 120 and 355-359 require `DIAGNOSTICS_MANIFEST.yaml` to own check IDs, reason/outcome codes, and budgets across all session-affecting surfaces in Slice 0.
* **Why it matters:** This creates a speculative central registry where future commands must pre-register diagnostic codes before they are designed, encouraging upfront vocabulary inflation.
* **Required Change:** Seed `DIAGNOSTICS_MANIFEST.yaml` only with existing behavior codes and the next adopting slice's codes. State that new codes move from `provisional` to active only when their concrete calling slice is decomposed and integrated with tests.

---

## Answers to Persona Questions

### 1. Which vocabulary types are required by slice 1 target classification versus introduced only for later slices?
* **Answer:** Slice 1 (read-only API target lookup) only requires types representing basic lookup inputs (like string tokens), a minimal `kind` set (e.g., `direct-id`, `live-session-name`, `live-alias`, `not-found`), and the resolved session ID or query result. Types for mutation conflicts (`SessionCommandConflict` - Slice 3), runtime intent (`RuntimeStartIntent` - Slice 5), or durable session events (`SessionFactEvent` - Slice 4) are completely out of scope and must not be defined or stubbed.

### 2. Does TR-007 future durable-event compatibility shape current APIs more than today's in-process events require?
* **Answer:** Not in the Attempt 15 text. Deferring event sourcing to post-commit facts and focusing on best-effort subscription prevents `TR-007` from shaping early APIs. However, the requirement that critical actions (close, work release) must converge from durable facts must be strictly enforced via offline recovery tests in Slice 0/1, ensuring no sneaky dependency on event logs is introduced.

### 3. What stops SessionFacts from becoming a broad facade accumulating every field any decider might want?
* **Answer:** The explicit prohibition of a single large `SessionFacts` struct (DESIGN.md:400) and the Rule of Repeated Exact Use (DESIGN.md:383-384, "Introduce broad or shared fact types only after repeated, exact use by multiple deciders/adapters, not in anticipation of sharing") are the primary barriers. To make this absolutely solid, the Target Candidate struct must also be broken up into per-kind types to set the right precedent from day one.

---

## Consistency & Parity Report

* **Requirements Alignment:** Under `REQUIREMENTS.md`, session resolution (e.g., `SESSION-ID-003` and `SESSION-ID-008`) must preserve exact precedence and fallback rules. The target classifier's precedence matrix (DESIGN.md:191) is aligned, but the assumption that "historical aliases reject live session targets" needs extreme care to prevent breaking existing CLI query behaviors where historical aliases might still be visible in some read-only listings.
* **Reviewer Interlock:** This review strongly aligns with Elena Marchetti's (Mutation Boundary Auditor) focus on limiting unowned writes and Takeshi Yamamoto's (Decider Atomicity Enforcer) focus on pure deciders. By pushing mutation contracts and atomic appliers out of Slice 0, we allow Elena and Takeshi to focus their gates exactly on the slices that introduce mutations, rather than forcing speculative reviews of un-executed mutation designs in Slice 0.

---

## Required Changes Before Approval

1. **Decompress Slice 0 Gate:** Move mutation-oriented command and boundary artifacts (`COMMAND_APPLIERS.yaml`, `BOUNDARY_MATRIX.yaml`, `WORKER_BOUNDARY_EXCEPTIONS.yaml`, and the mutation parts of `DIAGNOSTICS_MANIFEST.yaml`) out of Slice 0 and into the entry criteria for the specific mutation slices.
2. **Restrict Target Classifier Active Scope:** Explicitly mark non-API-query surfaces as `provisional` upper-bounds in the contract, forbidding active Go code or types from declaring fields or enums for them in Slice 1.
3. **Eliminate Flat Candidate Contradiction:** Refactor the Target Candidate contract (DESIGN.md:158-173) to use tagged per-kind structures, or add a binding rule that production Go code must use strict, per-kind types instead of a flat optional struct.
4. **Scope Diagnostics Manifest:** Seed diagnostics and route inventories with existing behavior and Slice 1 vocabulary only, growing them progressively per slice.
5. **Enforce Historical Note Non-Normative Status:** Add a explicit rule to the entry gates or Non-Goals stating that no uncopied section of `DESIGN_REVIEW_NOTES.md` can be treated as a normative requirement by any validator or implementation.

---

## Questions

1. Why does `DESIGN.md` explicitly forbid flat optional envelopes on line 383, but then immediately specify a flat, multi-field optional envelope for the Target Classifier on lines 158-173?
2. Can we explicitly define `api-query` as the *only* active surface in the target classification codebase for Slice 1, completely deleting the other 8 surface enums from early type definitions?
