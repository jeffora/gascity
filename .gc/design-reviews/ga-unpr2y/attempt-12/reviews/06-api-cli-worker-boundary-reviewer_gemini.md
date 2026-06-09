# Sarah Chen — DeepSeek V4 Flash (Independent Review, Attempt 12)

**Verdict:** block

**Review focus:** API/CLI projection reads, worker boundary routing, response compatibility, and layering — with direct evidence from the active codebase. Evaluated against the Attempt 12 iteration of `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-12/design-before.md`), `REQUIREMENTS.md`, the scoped `AGENTS.md`, and the active checkout source.

---

## Overview

The Attempt 12 iteration of `internal/session/DESIGN.md` makes outstanding structural progress. Formalizing "Slice 0" as an absolute preflight boundary, committing to clear requirements-to-slice mappings in `SCENARIO_PARITY.yaml`, and mapping Huma and CLI routing surfaces are excellent engineering responses. 

However, from the lane of **API/CLI Projection & Boundary Routing**, the design still presents significant vulnerabilities, ungrounded assumptions, and critical gaps that prevent safe decomposition. Crucially, the read-side projection adoption has no dedicated slice and relies on a near-vacuous guard that lets numerous raw metadata reads slip through in production. Parity between the CLI's dual execution paths (API-routed and local-fallback) remains completely unanchored in requirements. Furthermore, the "session command factory" exception path lacks any mechanical enforcement, meaning a backdoor is left wide open in the codebase with only advisory prose to restrict its growth.

Until these blockers are resolved in the design and proven in code, decomposition must remain blocked.

---

## Top Strengths

1. **Clear Division of Session Eligibility and Reconciler Policy:** Strictly partitioning lifecycle eligibility (session-owned deciders) from operational metrics such as desired demand, pool scaling, provider health, and alert deduplication (controller-owned, lines 405–418) prevents domain contamination and keeps the core simple.
2. **Disciplined Mapping of API Compatibility resolver chains:** Explicitly documenting the resolving order (direct session bead ID → open exact `session_name` → open exact current `alias` → allow-closed fallback, lines 438–440) protects the wire from breaking changes and provides developers with clear, actionable criteria for endpoint migration.
3. **Rigorous Stale-Proof Quarantine:** Refusing to accept missing test files (`scale_from_zero_test.go`, `provider_health_gate_test.go`, `session_progress_test.go`) as active evidence forces the extraction plan to remain grounded in reality, not historical memory.

---

## Critical Risks & Blockers

### 1. [Blocker] Read-Side Projection Adoption Under-Owned and Sheltered by a Near-Vacuous Guard
The core invariant of the Session architecture is that `ProjectLifecycle` is the single source of truth, and that callers must not scatter raw `state` or `sleep_reason` interpretation (lines 60–62 in `REQUIREMENTS.md`).

However, the active codebase relies heavily on raw metadata reads in critical user-facing paths. For example, `Metadata["state"]` is read directly on the following production lines:
* **`internal/api/handler_status.go:325`**: `state := strings.TrimSpace(bead.Metadata["state"])`
* **`internal/api/handler_status.go:371`**: `switch strings.TrimSpace(b.Metadata["state"])`
* **`internal/api/handler_status.go:488`**: `state := session.State(strings.TrimSpace(b.Metadata["state"]))`
* **`internal/api/huma_handlers_sessions_query.go:300`**: `if b, bErr := store.Get(id); bErr == nil && b.Metadata["state"] == "creating"`
* **`internal/api/session_resolution.go:414`**: `state := session.State(b.Metadata["state"])`

Despite these heavy production raw reads, **the design assigns no slice to own display-path and read-side projection conversion.** Worse, the cited oracle for `SESSION-LIFE-008`—`TestLifecycleUserFacingConsumersStayOnProjectionHelpers` in `lifecycle_projection_test.go:929`—is merely a denylist matching four exact string literals across four hardcoded test files. 
* It is completely blind to any of the production API files above because they use different variables (such as `bead.Metadata["state"]` or `b.Metadata["state"]`).
* It does not scan or restrict the CLI (`cmd/gc`), where approximately 83 raw metadata reads continue to re-derive state directly.

* **Impact:** Without a dedicated migration owner and an allowlist-driven static guard, the pipeline's mechanical freshness check for `SESSION-LIFE-008` will pass easily (since the brittle denylist test technically exists), while user-facing API status views, query endpoints, and CLI commands continue to bypass `ProjectLifecycle` and leak unvalidated raw states to operators.
* **Required Fix:** Expand Slice 0 to include a read-side deliverable: a generated inventory of raw reads on user-facing API and CLI surfaces. Replace the vacuous denylist in `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` with an AST-based allowlist guard, or explicitly assign read-path projection conversion to a named migration slice.

---

### 2. [Blocker] CLI Dual-Path Parity Lacks a Requirements Anchor
The CLI uses a dual-path execution model: commands attempt to route over the API using `apiClient()` and, on failure or environment preference, fall back to local direct execution (as seen in `cmd/gc/apiroute.go`). 

The design demands "local fallback behavior" parity, but **no scenario row in `REQUIREMENTS.md` anchors this requirement.** 

* **Impact:** Because there is no requirements-level anchor, `SCENARIO_PARITY.yaml` cannot map or enforce CLI dual-path equivalence. As independent slices modify the local execution paths (routing them through `worker.Handle`) or change the API route's underlying command, stdout, stderr, JSON shapes, exit codes, and side effects will silently diverge across the routed and fallback paths without triggering any test failures.
* **Required Fix:** Add a canonical scenario row to `REQUIREMENTS.md` stating that touched CLI commands must yield equivalent outputs (stdout, stderr, exit codes, and JSON schema) across both the API-routed path and the local execution fallback. Seed this row in `SCENARIO_PARITY.yaml` with `apiClientFallbackReason` behavior as the initial oracle.

---

### 3. [Blocker] approved Session Command Factory Escape Hatch Lacks Mechanical Enforcement
The Caller Routing gate (lines 420–435) allows callers to bypass `worker.Handle` and route "through an approved session command factory named by the slice," with Attempt 8 requiring an owner and expiry field per exception.

However, while `cmd/gc/worker_boundary_import_test.go` mechanically guards against direct session manager imports, **there is no static guard, allowlist schema, or enforcement artifact to verify that session command factory exceptions stay narrow and retire over time.**

* **Impact:** The factory exception mechanism is a prose-only rule with no mechanical teeth. Developers can freely instantiate new command factories or extend legacy ones without triggering static guard or test failures, creating a permanent bypass loophole that undermines the worker-boundary migration.
* **Required Fix:** Expand the static guard to inspect and assert factory construction. Define a named guard fixture and an allowlist schema field for approved factory exception sites. Until this is implemented, declare that `worker.Handle` is the only legal production entry point for CLI session mutations, limiting command factories to the API layer.

---

### 4. [Major] API Session-Command Inventory is Not Endpoint-Level Enough
The writer inventory maps package-level exceptions (such as W-015 and W-025) rather than pinning them to concrete Huma endpoint handlers.

Our active codebase reveals that Huma handlers still call manager and package mutators directly across many files. For example, in `internal/api/huma_handlers_sessions_command.go`:
* **Line 430**: `mgr.UpdatePresentation(...)`
* **Line 803**: `mgr.Suspend(...)`
* **Line 833**: `mgr.CloseDetailed(...)`
* **Line 881**: `session.WakeSession(...)`

* **Impact:** Package-level rows are too coarse to prove worker-boundary adoption or prevent regression on an endpoint-by-endpoint basis. If a single endpoint is partially converted, the broad package exception remains, leaving the remaining bypasses unmonitored.
* **Required Fix:** Expand the API inventory in Slice 0 to split package-level mutator rows into endpoint-level IDs. Each row must name the exact current mutator, intended route, owner slice, response/error contract, and retirement criteria.

---

## Missing Evidence

1. **User-Facing Raw Read Inventory:** A comprehensive file-and-line count of raw `Metadata["state"]` and `Metadata["sleep_reason"]` reads across user-facing API status, SSE dashboard, and CLI query surfaces.
2. **CLI Dual-Path Parity Fixtures:** Golden-file tests or schema validators asserting that standard commands (e.g., `gc status`, `gc stop`) output matching JSON, exit codes, and stream errors when executed via both the API routing layer and the local fallback fork.
3. **Mechanical Guard for Command Factories:** An AST-based check in `cmd/gc/worker_boundary_import_test.go` or a parallel guard that fails the build if an unmapped session command factory is constructed in production code.

---

## Required Changes

1. **Block Implementation Slices Beyond Slice 0:** Do not permit implementation of mutation-owning or routing-delegation slices to proceed until the Slice 0 preflight commits `BOUNDARY_INVENTORY.md`, `SESSION_BOUNDARY_SYMBOLS.yaml`, `SCENARIO_PARITY.yaml`, and `VOCABULARY_CHECKPOINTS.yaml`.
2. **Inventory and Guard Raw Reads:** Add a read-side inventory task to Slice 0, and expand `TestLifecycleUserFacingConsumersStayOnProjectionHelpers` to scan API and CLI roots, failing on unmapped raw reads.
3. **Anchor CLI Dual-Path Parity:** Commit a new requirement row to `REQUIREMENTS.md` enforcing dual-path CLI consistency, and require each CLI-affecting slice to register and pass corresponding parity fixtures.
4. **Enforce Factory Exceptions Mechanically:** Add an explicit AST import and construction guard for the session command factory, backed by a shrink-only allowlist with mandatory owner and expiry fields.
5. **Deconstruct Package-Level API Inventory Rows:** Force Slice 0 to expand broad rows (W-015, W-025) into endpoint-specific rows for every session-affecting Huma and legacy handler.
