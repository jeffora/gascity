# Natasha Volkov — DeepSeek V4 Flash (Independent Review, Attempt 16)

**Verdict:** block

**Scope:** REQUIREMENTS scenario parity, regression prevention, characterization tests, proof freshness — with evidence drawn from verification of `internal/session/DESIGN.md`, `internal/session/REQUIREMENTS.md`, and the active workspace.

---

## Overview

Attempt 16 (represented by the revised draft in `internal/session/DESIGN.md`) is a thoughtful and highly detailed "iterate" response to the previous global blocks. It successfully narrows the scope of the approved next work to a self-validating, non-mutating Slice 0, and replaces the broad target classification adapter with a detailed, eight-step resolver precedence matrix that matches legacy behavior.

However, from the perspective of the Behavior Parity Guardian, this revision still contains critical **circular design paradoxes**, **deferred design-time mappings**, and **query-side resolution gaps** that will cause severe regressions if the design is decomposed as-is. Under project schedules, deferring traceability mapping to implementation files and hand-waving the read-only repair path will force implementers to either bypass validators or deploy breaking changes to production endpoints.

---

## Top Strengths

- **Detailed Target Resolver Precedence (DESIGN.md:220-245):** Specifying the exact eight-step target resolution order—including configured named-sessions, config-orphan rejections, path-alias by `Title` matching, and allow-closed query boundaries—is an excellent preservation of legacy behavior.
- **Strict Single-Owner Command Restrictions:** Prohibiting future command, event, and diagnostics contracts from being pre-authorized in Slice 0 ensures that mutating capabilities only land with the specific slices that implement their caller-delegations.
- **Pure, Side-Effect-Free Classifier Contract:** Ensuring the classifier never performs mutations, session materializations, or runtime provider calls protects the core target-resolution engine from race conditions and split-brain state mutations during query-side operations.

---

## Blocking Findings

### [Blocker] The Stale Evidence Paradox remains unresolved in Slice 0

The design rightly splits Slice 0 into universal evidence and per-slice preflights, and requires that universal Slice 0 "explicitly reconcile active requirement rows whose evidence is stale, missing, or on another ref" (`DESIGN.md:173-174`). At minimum, it must "repair or owner-retire the evidence for `SESSION-RECON-002`, `SESSION-RECON-003`, `SESSION-RECON-006`, and `SESSION-RECON-007` before a later slice cites those rows."

However, this requirement introduces a fundamental **structural contradiction**:
1. Slice 0 is defined as strictly **non-mutating and session-only**; it does not touch reconciler policy or write to any non-session store.
2. The stale/missing evidence paths cited in `REQUIREMENTS.md` (e.g., `cmd/gc/scale_from_zero_test.go`, `cmd/gc/provider_health_gate_test.go`, and `cmd/gc/session_progress_test.go`) belong to the **reconciler and provider-health sub-systems** (Layer 2-4), which are outside `internal/session`.
3. Repairing these tests requires restoring or writing complex reconciler scaling, health, and progress integration tests.
4. "Owner-retiring" these rows is unacceptable because they represent safety-critical production behaviors (such as cold-start clamping and health alerts) that must not be deleted or hidden just because their test evidence is currently missing on this branch.

Because Slice 0 cannot mutatively restore Layer 2-4 reconciler integration tests without violating its own "session-only, non-mutating" boundary, and cannot retire them without deleting product requirements, the Slice 0 validator (`TestScenarioParityFreshness`) is **guaranteed to fail immediately and permanently**, blocking all progress.

**Required change:** 
- Explicitly allocate the repair/restoration of `scale_from_zero_test.go`, `provider_health_gate_test.go`, and `session_progress_test.go` to a preceding or parallel **Reconciler Test-Hardening Slice (Slice 6/7 Backlog)** rather than the non-mutating Slice 0.
- Provide a machine-checkable **Transition Proof Allowlist** in Slice 0's validator allowing these specific reconciler rows to temporarily fall back to their historical commit citations (e.g., `commit a2b2da046`, `commit dbda1e380`) in `REQUIREMENTS.md` until the corresponding test-hardening slice restores the files.
- Annotate these four affected rows directly in `REQUIREMENTS.md` as `[STALE - REQUIRES SLICE 6/7 REPAIR]` so they are not mistaken for live executable evidence.

---

### [Blocker] Deferred design-time traceability prevents backlog verification

By removing the `## Scenario Traceability Matrix` entirely from `DESIGN.md` and deferring the row-to-slice allocation to a future `SCENARIO_PARITY.yaml` file to be created in Slice 0, the design backlog slices (Slices 1 to 6) **completely lack scenario-row mapping**. 

We cannot verify at this design gate that the refactor backlog is comprehensive, or that critical scenario rows (such as `SESSION-LIFE-001` legacy state projection, `SESSION-LIFE-002` pending-create claim, `SESSION-LIFE-006` missing-config block, or `SESSION-RUNTIME-004` stop turn) have assigned slices or owners. Reviewers are forced to accept on faith that the implementer will map them correctly during Slice 0.

**Required change:** Re-introduce a high-level **Scenario Allocation Matrix** directly in `DESIGN.md` that maps groups of `SESSION-*` requirements to their target backlog slices (e.g., Slice 1 owns target classification, Slice 3 owns start/wake mutations, Slice 4 owns close detailed, Slice 5/6 owns reconciler fact extraction). This ensures completeness is proven at the design gate before implementation begins.

---

### [Blocker] Query-side "repair-needed" quarantine causes behavioral regression

The design specifies that `RepairEmptyType` is prohibited in the read-only classifier path and that the classifier must instead return a `repair-needed` result kind (`DESIGN.md:246-249`). 

However, the design fails to define what the query-side caller (such as `GET /sessions/<id>` or transcript/stream attachment GET endpoints) actually does when it receives `repair-needed`:
1. If the query endpoint returns an error (e.g., 404 or 409) or fails to resolve the session, this is a **severe behavioral regression**. Today, these endpoints silently repair the empty type and succeed.
2. If the query endpoint triggers the separate repair write synchronously, it violates safe HTTP GET method semantics and defeats the purpose of keeping the query path strictly read-only and side-effect-free.
3. If it falls back to the legacy resolution path, we introduce a dual-source of truth and defeat the purpose of centralizing target classification.

The hand-wave that "parity tests must prove whether the first adopter preserves successful selection after a separate repair or intentionally changes the result" (`DESIGN.md:249-251`) is unacceptable. The design must define this behavior before implementation.

**Required change:** Specify the exact adapter wrapper behavior for `repair-needed` in query-side endpoints. Specifically:
- Define that read-only query callers will **temporarily resolve the target using the unrepaired/raw metadata** returned in `bead_state` to preserve successful lookup parity, but must publish an asynchronous repair event or trigger a background repair task through the audited repair owner.
- Prohibit query-side GET handlers from blocking on database mutations or returning errors to clients for repairable empty-type sessions.

---

## Major Findings

### [Major] Absence of "Black-Box" assertion rules for Characterization Tests

The `Refactor Rules` (`DESIGN.md:652`) state that "The test should prove the behavior the user sees, not every internal branch." However, this is too weak to prevent regression. 

Without a strict prohibition against white-box mock assertions (such as mocking internal store interfaces or asserting internal function call chains), developers or agents will write brittle mocks that pass during the refactor even if the user-visible product behavior (exit codes, output payloads, or database commit states) is completely broken.

**Required change:** Add an explicit, non-negotiable rule to the `Refactor Rules`:
> "Characterization tests must be black-box, end-to-end, or integration-level tests asserting user-visible or system-level outputs (such as exit codes, stdout/stderr shape, API status codes, and database commit states) rather than white-box mocks of internal interfaces. The exact same characterization tests must run unchanged against both the legacy baseline and the refactored path to prove parity."

---

### [Major] Lack of assertion-level verification in Proof-Freshness validation

The Slice 0 validator `TestScenarioParityFreshness` checks if cited file paths exist (`DESIGN.md:181-182`). However, a file-existence check cannot detect if the tests inside that file have been renamed, gutted, or bypassed using `t.Skip()`. The fact that reconciler requirements were allowed to go completely missing while their citations remained in `REQUIREMENTS.md` proves that path-level checks are insufficient to prevent proof rot.

**Required change:** Require that `SCENARIO_PARITY.yaml` specifies both the file path and the **exact test function symbol(s)** (e.g., `TestSessionLifecycle/Wake_Held_Until`). The Slice 0 freshness validator must dynamically parse or execute the tests to verify that the named test functions exist and do not contain hardcoded skips.

---

## Minor Findings & Questions

- **Deferred decision on `resolveLiveSessionByPathAlias`:** The budget row for `resolveLiveSessionByPathAlias` (`DESIGN.md:581`) mentions a "Decision to index, remove, or keep with explicit scan budget." This is a deferred design choice. The design should state the chosen path (e.g., "This path is kept but explicitly indexed by Title") rather than deferring the decision itself.
- **Dynamic Key Static Guard:** How will the static guard specified in Slice 0 handle dynamic metadata key writes (e.g., loops writing variable key patterns)? Will it enforce that any `SetMetadata` with a non-literal key is a violation? We recommend that dynamic-key patterns must be explicitly registered as exceptions in `SESSION_BOUNDARY_SYMBOLS.yaml`.

---

## Summary of Required Changes

1. **Resolve the Stale Evidence Paradox:** Move reconciler integration test restoration to Slice 6/7 Backlog, add a **Transition Proof Allowlist** to Slice 0's validator allowing missing paths to fall back to historical commit hashes, and annotate the affected rows in `REQUIREMENTS.md`.
2. **Re-introduce the Scenario Allocation Matrix:** Add a high-level table to `DESIGN.md` mapping all `SESSION-*` requirements to their target backlog slices to ensure design-time coverage verification.
3. **Define Query-Side Repair-Needed Behavior:** Explicitly specify that query-side GET endpoints resolve empty-type sessions using the raw returned metadata to preserve successful lookup, and trigger background/asynchronous repair rather than blocking on writes or returning errors.
4. **Mandate Black-Box Characterization Tests:** Add a strict rule to the Refactor Rules forbidding white-box mock assertions.
5. **Verify Assertion-Level Freshness:** Update the Slice 0 freshness validator to check exact test function symbols and skip-states, not just file existence.
