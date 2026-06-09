# Sarah Chen — DeepSeek V4 Flash (Independent Review, Attempt 15 Response Iteration)

**Verdict:** block

**Lane:** API and CLI projection, worker boundary routing, response compatibility, layering.

---

## Overview

As the **06-api-cli-worker-boundary-reviewer** persona (Sarah Chen), this review covers the **Attempt 15 review-response revision** of `internal/session/DESIGN.md` alongside the canonical requirements in `internal/session/REQUIREMENTS.md` and the modular boundaries in `internal/session/AGENTS.md`.

The current iteration of the design has made major improvements by explicitly addressing several of the key structural concerns raised in Attempt 15. Specifically, it now mandates `worker.Handle` wrapping for mutating lifecycle operations, aligns API close with the worker boundary, and enforces strict wire-parity requirements for Huma and legacy error-resolution paths (`writeResolveError`/`humaResolveError`).

However, from the perspective of the **API and CLI projection and Worker Boundary** lane, several critical structural loopholes, un-designed integration gaps, and enforcement voids remain. Additionally, the design still suffers from extreme "document creep" in Slice 0, which directly threatens developer velocity and conflicts with the **DX-wins** principle.

Therefore, this lane assigns a verdict of **block** on the current revision.

---

## Top Strengths

1. **Explicit Mutation Routing Rules (DESIGN.md:404-406, 415-417):** The design now clearly states that mutating lifecycle operations (Wake, Drain, Runtime Start) exposed through the production API or CLI must route through `internal/worker/handle.go` unless an exact expiring exception is recorded. This preserves the single-entry-point invariant of the worker runtime.
2. **Unified API Close Plan (DESIGN.md:410-414):** Aligning the Huma and legacy API close paths through `worker.Handle.CloseDetailed` closes a major loophole where close behaviors and operational event emission were diverging.
3. **Mandatory Error Parity Suite (DESIGN.md:282, 423-425):** Requiring no-delta tests for `writeResolveError` and `humaResolveError` ensures that centralizing target classification does not silently break the wire contract by mapping existing client-visible error states (like 404, 409) into default 500 responses.

---

## Critical Risks & Architectural Gaps

### 1. [Blocker] Documentation-Only API Worker Boundary (No CI Enforcement)
* **The Gap:** The design relies heavily on `WORKER_BOUNDARY_EXCEPTIONS.yaml` and states that all production API/CLI mutating operations must route through `worker.Handle`.
* **Sarah's Critique:** In the current codebase, the only existing guard is `TestGCNonTestFilesStayOnWorkerBoundary` (`cmd/gc/worker_boundary_import_test.go`), which scans `cmd/gc` exclusively and has no exception mechanism. There is **zero active CI enforcement** scanning `internal/api` to prevent direct imports of `session.Manager` or bypasses of `worker.Handle`.
  Without converting `TestGCNonTestFilesStayOnWorkerBoundary` to scan `internal/api` and consume `WORKER_BOUNDARY_EXCEPTIONS.yaml` as an allowlist, the API worker-boundary rules are purely documentation. A developer can easily add direct mutating manager calls in a Huma handler, and CI will pass. This gap must be closed in Slice 0.

### 2. [Blocker] Unaddressed API Raw-`state` Reads Bypassing `ProjectLifecycle`
* **The Risk:** Multiple critical API handlers directly read and interpret raw metadata state:
  - `internal/api/handler_status.go:325,371,488`
  - `internal/api/session_resolution.go:414`
  - `internal/api/huma_handlers_sessions_query.go:300`
  They read `Metadata["state"]` or `Metadata["sleep_reason"]` directly instead of using the canonical `ProjectLifecycle` helper.
* **Sarah's Critique:** The design claims to clean up callers that "know too much about session state... and lifecycle metadata" (DESIGN.md:26-27), but the backlog (Slices 1-6) completely ignores these high-risk direct state reads. Leaving these raw reads unmapped and untracked means they will silently drift as the state machine refactor progresses. Slice 0 or Slice 1 must explicitly schedule their migration onto `ProjectLifecycle` or track them as bounded exceptions.

### 3. [Blocker] Undefined First-Adopter Endpoint Scope
* **The Gap:** Slice 1 (Target Classification) lists "API query-side read-only session lookup" as the first adopter (DESIGN.md:288).
* **Sarah's Critique:** The design fails to explicitly identify the target read-only endpoints (e.g. `GET /v1/sessions`, `GET /v1/sessions/{id}`) that will adopt the classifier first, nor does it list the specific `materialize:true` query-side callers that must remain behavior-compatible. Leaving the exact wire-parity boundaries of Slice 1 undefined makes verification of the first-adopter contract impossible to enforce in CI.

### 4. [Blocker] Unabated Slice 0 "Document Creep"
* **The Gap:** Slice 0 still mandates the creation and manual synchronization of **12 separate YAML and MD files** (DESIGN.md:151-166) before any behavior-moving code can be merged.
* **Sarah's Critique:** This extreme level of document-driven development forces developers to manually coordinate overlapping facts (e.g., symbols, routes, and boundary matrices) across a dozen different files during rapid refactoring. This introduces massive developer friction, directly violating the **DX-wins** principle (*"When the DX conflicts with the docs, DX wins"*). These files must be consolidated to prevent documentation stale-drift from paralyzing the migration.

---

## Required Action Items for Approval

To resolve these blockers and approve the design, the following updates must be made to `DESIGN.md`:

1. **Extend Worker Boundary Enforcement to API:**
   - Mandate that Slice 0 upgrades `TestGCNonTestFilesStayOnWorkerBoundary` to scan `internal/api` and convert the flat denylist into a ledger-consuming guard that parses `WORKER_BOUNDARY_EXCEPTIONS.yaml`.
2. **Backlog API Raw-State Cleanups:**
   - Explicitly schedule the migration of direct `Metadata["state"]` reads in `handler_status.go`, `session_resolution.go`, and `huma_handlers_sessions_query.go` onto `ProjectLifecycle` as part of the backlog, or record them as expiring exceptions.
3. **Specify Slice 1 Target Endpoints:**
   - Explicitly list the read-only endpoints (including `resolveSessionTargetIDWithContext` callers) that form the boundary of the Slice 1 classifier adoption, along with their expected parity test cases.
4. **Consolidate Slice 0 Artifacts:**
   - Reduce the 12-file Slice 0 preflight requirement to a maximum of 4 consolidated files (e.g., merge baseline, contract, exceptions, and route inventory into a single `SESSION_MIGRATION_MANIFEST.yaml`).

---

## Questions for the Team

1. Will the worker-boundary guard be extended to cover `internal/api` in CI as part of Slice 0, or are we comfortable with the API worker-boundary rule being documentation-only?
2. Why are we requiring developers to manually maintain overlapping facts across 12 separate files instead of using a single consolidated manifest?
