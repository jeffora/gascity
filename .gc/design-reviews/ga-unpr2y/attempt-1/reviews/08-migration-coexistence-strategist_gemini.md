# Ravi Krishnamurthy - DeepSeek V4 Flash

**Verdict:** block

Lane: migration sequencing, legacy-new coexistence, rollback slices, worker-boundary collision. This reviews the current `DESIGN.md` (the attempt-16 `iterate`-response revision, located at `.gc/design-reviews/ga-unpr2y/attempt-17/design-before.md`), alongside `REQUIREMENTS.md`, the in-flight worker-boundary migration, and the existing codebase. Findings are validated against the live check-out on this branch with precise code-level citations and inline references.

---

### Top strengths:
- **Clean Architectural Division of Labor:** The strict separation of concerns outlined in `DESIGN.md:550-563` is excellent. Keeping controller-level responsibilities (such as scheduling, budgets, desired pool sizes, and progress policies) completely outside `internal/session` ensures that the core session domain model remains thin and focused solely on pure state-transition and target-classification rules.
- **Query-Path Integrity and No-Repair Rule:** Enforcing that the target classifier is strictly side-effect free and bans silent repairs (`DESIGN.md:247-256`) is a crucial structural improvement. Pushing `RepairEmptyType` completely out of the read-only classification path and returning a typed `repair-needed` status prevents hidden state mutation during lookup operations and keeps query paths deterministic.
- **Robust Schema Progressive Activation:** Levering progressive config presence levels (Levels 0-8, `DESIGN.md:43-44`) as the universal activation mechanism provides a clear, highly-controllable pathway for rolling out features without requiring immediate global flag days.

---

### Critical risks:

#### 1. [Blocker] Unfenced Split-Brain Mutation Races During Partial Coexistence Window
- **Evidence:** `DESIGN.md:404-410` ("coexistence fences"), `DESIGN.md:521-523` ("migrations overlap"), and `DESIGN.md:539-542` ("No slice may leave both a legacy raw writer and a new session-owned command writing the same field family...").
- **Why it matters:** The design permits transitional "exception" periods where legacy raw metadata writers and new session-owned commands write to the same metadata fields. In reality, multiple concurrent processes (such as the in-process background reconciler, the CLI, and multiple Huma API server threads) can write to the same session beads. Legacy writers (such as `cmd/gc/session_beads.go:1737` and `session_reconciler.go:206`) perform **blind, unfenced writes** via `SetMetadataBatch` or direct `ClosePatch` application. They are completely unaware of the command-level optimistic concurrency fences (such as revision tokens or value preconditions) proposed in lines 417-465. If a legacy writer performs a blind status/close write concurrently with or after a fenced command, it will silently clobber the command's atomic update. The local database (`beads.db`) does not natively enforce field-level ownership or cross-process concurrency checks without coordination.
- **Mitigation:** The design must enforce that a key family's writer boundary is absolute and binary: any given field family (e.g., `alias`, `session_name`, or `state`) must have **exactly one** active writer path on any active production surface. A mixed-writer state must be forbidden. The transition of any field family must be atomic across all active production surfaces, or progressive activation levels (config Levels 0-8) must be used to strictly shut off legacy write paths when command paths are enabled.

#### 2. [Blocker] High-Risk Collision with the In-Flight Worker-Boundary Migration
- **Evidence:** `DESIGN.md:466-490` ("Worker/API/CLI Boundary"), `DESIGN.md:521-523` ("overlap"), and `cmd/gc/worker_boundary_import_test.go:11-49` (which enforces imports like `session.NewManager` are forbidden in non-test files under `cmd/gc`).
- **Why it matters:** There is an active, in-flight worker-boundary migration (started `12a0a848` on Apr 17 2026) that is aggressively routing all production `cmd/gc` operations through `worker.Handle`. However, the session refactor (Slices 1-5, e.g. lines 792-809) proposes moving target resolution, wake, close, and runtime start into `internal/session` deciders and commands. Because the session refactor will modify the exact same files and call sites (such as `resolveSessionTargetIDWithContext` and `materializeNamedSessionWithContext` in `session_resolution.go`), we have two active, partial refactors touching the exact same files and interfaces at the same time. If a slice of the session refactor alters `session_resolution.go` to introduce new classifier result types or command structures, it risks violating the worker-boundary imports (e.g., triggering `TestGCNonTestFilesStayOnWorkerBoundary` failure or introducing illegal imports).
- **Mitigation:** A strict sequencing rule is required: the worker-boundary migration must be **fully finalized** and `WORKER_BOUNDARY_EXCEPTIONS.yaml` must be cleared of temporary exceptions *before* any mutating session refactor slice (Slices 3-5) is allowed to touch those files. Alternatively, the two migrations must share a unified exception ledger.

#### 3. [Major] Lack of Downgrade Compatibility and Rollback Data-Direction Validation
- **Evidence:** `DESIGN.md:530-533` ("rollback data direction... tests proving old readers tolerate new fields...") and `DESIGN.md:718-732` ("Vocabulary Lifecycle").
- **Why it matters:** When a slice is deployed that introduces new metadata fields (such as `result_kind`, `match_vectors`, or provider-neutral `RuntimeIntent` fields) or state transition codes, and then the deployment has to be rolled back/reverted, the old version of the binary (the legacy reader) will run against the database containing these new fields. If the legacy code is not "downgrade-safe"—for instance, if it fails to parse because of unexpected keys in the JSON or if it misinterprets the state—the rollback will cause a catastrophic system failure (a "flag day" or a bricked database).
- **Mitigation:** The design must enforce a strict "two-phase schema migration" rule for all metadata changes:
  1. Phase 1: Deploy code that is *tolerant* of reading the new fields but still writes the old format (Read-Compatible).
  2. Phase 2: Deploy code that writes the new fields.
  Every slice must provide a rollback test suite proving that if the binary is downgraded to the previous slice's version, the database state remains valid and readable.

#### 4. [Major] Underspecified Shared-Resolver Sequencing & Target Precedence Anti-Drift
- **Evidence:** `DESIGN.md:239-245` ("Shared resolver sequencing rule... If a slice cannot use one shared resolver implementation, it must add anti-drift tests...") and `internal/api/session_resolution.go:429-477` (`resolveSessionTargetIDWithContext`).
- **Why it matters:** The design acknowledges that read-only target classification and materializing resolution must share a single source of precedence, or use extensive "anti-drift tests" to compare match vectors and result kinds. However, `session_resolution.go` is heavily complex, involving:
  - configure-name resolution (lines 221-245)
  - exact session ID lookup (line 436)
  - alias/session-name lookup with alias demotion (line 446)
  - path-alias by `Title` matching (line 459)
  - allow-closed historical lookups (line 464).
  If read-only classification forks target resolution behavior even slightly, query endpoints and mutation endpoints will resolve the same token to different sessions. The design lacks concrete test specifications or an automated validation tool to enforce this anti-drift guarantee.
- **Mitigation:** Specify that the anti-drift tests must be automated in CI and run the entire target resolution matrix across both the new `internal/session` classifier and the legacy `session_resolution.go` resolver before any target-classification surface is delegated.

---

### Missing evidence:
- **No Concrete Backlog Completion Criteria:** The backlog slices name high-level concepts but provide no concrete, file-or-symbol-level completion criteria to determine when a slice is complete and can be merged.
- **Unmapped `RepairEmptyType` Sites:** There are exactly 14 call sites for `RepairEmptyType` across the codebase (including 2 in `internal/mail/beadmail/beadmail.go` and multiple in Huma commands). The design-before document does not provide a concrete backlog mapping or completion criteria for retiring these sites.

---

### Required changes:
1. **Define Atomic Key Family Cut-over:** Mandate that once a key family (such as `alias`, `session_name`, or `state`) begins migrating, all active production writers for that family must cut over atomically. No concurrent legacy and new writers are allowed to coexist on active production surfaces.
2. **Sequence Worker-Boundary Finalization:** Require that the worker-boundary migration is fully finalized and closed before Slices 3-5 of the session refactor can be merged.
3. **Mandate Two-Phase Metadata Migrations:** Enforce that any change to metadata fields or state codes must be deployed in a two-phase (Read-Compatible then Write-Active) sequence, and require downgrade compatibility tests.
4. **Detail Target Classifier Anti-Drift Verification:** Specify a concrete test command and fixture set in `TARGET_CLASSIFICATION_CONTRACT.yaml` that proves exact target precedence match parity between the read-only and materializing modes.

---

### Questions:
- If a rollback is triggered, what specific tooling or commands will be used to clean up/repair any new metadata fields written by the reverted version?
- How will we prevent the background reconciler from performing a blind write to `state` while the API is mid-flight processing an atomic command update?

---

## Answers to Persona Questions

### 1. How does the plan sequence this extraction with the in-flight worker-boundary migration on overlapping cmd/gc and internal/api call sites?
**Answer:** The current plan fails to specify a formal sequence for coordinating with the in-flight worker-boundary migration. Both migrations overlap directly on `internal/api/session_resolution.go`, `cmd/gc/session_reconciler.go`, and `cmd/gc/session_beads.go`. If behavior-moving slices are rolled out before the worker-boundary exceptions are retired, they will inject direct, raw manager/store-level mutations in files that `TestGCNonTestFilesStayOnWorkerBoundary` is attempting to restrict. To sequence this properly, the worker-boundary migration must be finalized and its exceptions retired *before* mutating session refactor slices (Slices 3-5) are allowed to touch the overlapping files.

### 2. During partial adoption, what prevents legacy patch-map callers and new command callers from split-brain writes to the same metadata fields?
**Answer:** In the current design, **nothing** robustly prevents split-brain writes during partial adoption. While the design specifies "coexistence fences" for command appliers, the legacy callers (such as `cmd/gc/session_lifecycle_parallel.go:1823` and `session_reconciler.go:206`) perform direct, unfenced updates via `SetMetadataBatch` or raw map merges. A legacy caller will blindly clobber any fenced, atomically-validated writes because the store itself does not enforce metadata ownership fields. To prevent this, the transition of any metadata key family must be binary and atomic across all production surfaces—either all writers use the legacy path, or all use the command path.

### 3. Which single slice is independently shippable and revertible, and what proves it does not silently require the next slice?
**Answer:** Only Slice 0 (universal evidence gathering) and Slice 1 (read-only target classification) are truly independently shippable and revertible. Because they are strictly read-only and side-effect free, they do not mutate state and can be reverted without risking backward-compatibility failures. To prove a mutating slice does not silently require the next slice, we must mandate downgrade/rollback compatibility tests in CI: running the previous version's test suite against data produced by the current slice's version.
