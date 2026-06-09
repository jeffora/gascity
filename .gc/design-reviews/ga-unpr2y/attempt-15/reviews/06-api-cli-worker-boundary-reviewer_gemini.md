# Sarah Chen — DeepSeek V4 Flash (Independent Review, Attempt 15)

**Verdict:** block

**Lane:** API and CLI projection, worker boundary routing, response compatibility, layering.

---

## Overview

As the **06-api-cli-worker-boundary-reviewer** persona (Sarah Chen), this review covers the Attempt 15 iteration of the minimal `internal/session/DESIGN.md` (matching `.gc/design-reviews/ga-unpr2y/attempt-15/design-before.md`) alongside the canonical requirements in `internal/session/REQUIREMENTS.md` and the modular boundaries in `internal/session/AGENTS.md`.

The Attempt 15 revision has successfully streamlined the overall design compared to older attempts, focusing on an incremental slice-by-slice approach (Slices 1-6) rather than a monolithic rewrite. The pure read-only Target Classification Contract (lines 134-197) is an elegant first step that cleanly isolates resolution from side effects.

However, from the perspective of the **API and CLI projection and Worker Boundary** lane, the current design contains critical structural loopholes, un-designed integration gaps, and a severe case of "document creep" that directly threatens developer velocity and the **DX-wins** principle. 

Therefore, this lane assigns a verdict of **block**.

---

## Top Strengths

1. **Side-Effect-Free Classifier Taxonomy:** The Target Classification Contract (lines 134-197) correctly enforces that the raw classifier is entirely read-only (no store writes, no session materialization, no provider execution). This keeps the classification model highly testable and decoupled from operational state mutations.
2. **Explicit Fact/Decision Splitting:** The "Reconciler, Runtime, and Session Split" matrix (lines 288-312) correctly draws the boundary between pure deciders and external callers. Keeping work demand, scheduling, budgets, and provider policy in the reconciler prevents domain pollution.
3. **Structured Surface Precedence Matrix:** Explicitly cataloging the current target-resolution precedence for each surface (lines 187-197) provides a clear blueprint for preventing regression during target delegation.

---

## Critical Risks & Architectural Gaps

### 1. [Blocker] Deferral of Worker Boundary Integration to Exceptions File
* **The Gap:** The design states: *"Wake and drain slices must decide whether the operation stays store-level or grows `worker.Handle`; the decision must be recorded in `WORKER_BOUNDARY_EXCEPTIONS.yaml` before delegation."* (lines 273-275).
* **Sarah's Critique:** This is not an architectural decision; it is a permission to bypass architecture. The worker boundary (`internal/worker/handle.go`) is the load-bearing gate for all process execution in Gas City, strictly enforced by `TestGCNonTestFilesStayOnWorkerBoundary`. 
  If `Wake` (Slice 3) or `Drain` (Slice 5) bypasses `worker.Handle` and executes directly at the store level from the CLI or API, they break the single-entry-point invariant of the worker runtime. Simply recording this bypass as an exception in a YAML file does not solve the architectural tension. The design must explicitly mandate that all mutating operations remain wrapped by `worker.Handle` to maintain layering integrity, rather than deferring the decision to individual implementation slices.

### 2. [Blocker] "Document Creep" and Metadata Sync Overhead (Slice 0 Preflight)
* **The Gap:** Slice 0 (lines 98-133) now requires the creation of **11 separate YAML/MD files** containing metadata, Symbol lists, Route lists, boundary matrices, and diagnostics manifests before a single line of target-classification logic can be merged.
* **Sarah's Critique:** While rigorous traceability is excellent, this extreme level of **document-driven development (DDD)** introduces massive developer friction and directly violates the project's core **DX-wins** principle (*"When the DX conflicts with the docs, DX wins"*). 
  Specifically, files like `BOUNDARY_INVENTORY.md`, `SESSION_BOUNDARY_SYMBOLS.yaml`, `BOUNDARY_MATRIX.yaml`, and `API_CLI_ROUTE_INVENTORY.yaml` track overlapping facts about session boundary surface areas. Forcing developers to manually synchronize symbols across 11 files during a refactoring iteration is an anti-pattern. This overhead must be consolidated; otherwise, it will lead to constant CI build failures over stale docs and stall the migration.

### 3. [Blocker] Untested Wire-Compatibility in API Error Mapping (Slice 1 Target Classification)
* **The Risk:** Centralizing target classification inside `internal/session` will return unified error types. However, standard API routes and Huma handlers (`internal/api/huma_handlers_sessions.go`) perform explicit type assertions or `errors.Is` checks on errors to return structured JSON responses (e.g., mapping `session.ErrSessionNotFound` to a `404 Not Found` with a specific JSON body).
* **Sarah's Critique:** If the new centralized classifier returns a wrapped, nested, or custom error structure (e.g. to attach rich diagnostic context) that fails standard `errors.Is` matching, the API handlers will silently hit default fallback blocks, mapping valid `404` or `409` errors into `500 Internal Server Error` responses. The design lacks any static compilation guard or mandatory unit-test contract that asserts every classification error state is correctly translated by the API adapter. Without this, Slice 1 is highly likely to break the public API wire contract.

### 4. [Blocker] Caller-Specific Resolver Overlays Remain Unmapped by REQUIREMENTS
* **The Gap:** The design assumes that `SCENARIO_PARITY.yaml` mapping to `REQUIREMENTS.md` rows is sufficient to prevent behavioral regression.
* **Sarah's Critique:** The complex resolver logic currently in the CLI (`cmd/gc/session_resolve.go`) and API (`internal/api/session_resolution.go`) contains behavior overlays—such as config-named reopening, template target rejection, and API-side active path-alias matching—that **are not represented as rows in REQUIREMENTS.md**. 
  Because these are caller-specific overlays, a developer could easily refactor the target classifier, break qualified-basename resolution or path-alias fallback, and still pass every "Product Rule" requirement row. The design must mandate that these caller-specific layers are explicitly rowed as requirements or covered by mandatory multi-surface integration test fixtures before delegation begins.

### 5. [Minor] Silent Bypasses of `worker.Handle` via CLI-API Direct Client Calls
* **The Risk:** CLI commands in `cmd/gc/*.go` currently import `internal/api/genclient` and call API endpoints directly.
* **Sarah's Critique:** If a CLI command bypasses the local `worker.Handle` by calling the API client directly to perform an operation, it circumvents the worker boundary. The design does not address or audit this direct-client bypass, which could allow developers to mutate sessions without routing through the local worker boundary logic.

---

## Required Action Items for Approval

To resolve these blockers and approve the design, the following updates must be made to `DESIGN.md`:

1. **Mandate `worker.Handle` Wrapping for Mutations:**
   - Formally state that `worker.Handle` is the sole execution wrapper for all mutating session operations (Wake, Close, Drain, Runtime Start). Bypassing `worker.Handle` in CLI production code to make direct store-level writes is strictly forbidden.
2. **Consolidate Slice 0 Artifacts:**
   - Consolidate the 11-artifact preflight list into a maximum of 4 high-level artifacts to reduce document creep and eliminate sync duplication (e.g., merge symbols, inventory, exceptions, and route inventory into a single `SESSION_BOUNDARY_MANIFEST.yaml`).
3. **Establish an Error-Translation Unit Test Suite:**
   - Require that Slice 1 defines a strict unit-testing gate for `writeResolveError` and `humaResolveError` to assert that every error produced by the centralized classifier maps to the exact expected HTTP status codes and response bodies.
4. **Row CLI/API Overlays in REQUIREMENTS.md:**
   - Explicitly add requirements rows for caller-specific overlays (config reopening, template target rejection, and path-alias resolution) to `REQUIREMENTS.md` before Slice 1 target classification work is scheduled.

---

## Questions for the Team

1. Why are we tracking overlapping facets of the boundary across 11 different Slice 0 files instead of centralizing them in a single machine-readable manifest?
2. How will we prevent the `gc` CLI from bypassing the local `worker.Handle` via direct `genclient` API calls for mutating operations?
