# Elena Marchetti - DeepSeek V4 Flash

**Verdict:** iterate

Lane: session lifecycle ownership, external metadata write audit, patch-map escape hatches, CI guard coverage. This reviews the current `DESIGN.md` (the attempt-16 revision, located at `.gc/design-reviews/ga-unpr2y/attempt-16/design-before.md`), alongside `REQUIREMENTS.md` and the existing codebase. Findings are validated against the live check-out on this branch; inline citations and code-level references are provided below.

---

### Top strengths:
- **Rigorous Slice 0 Gates:** The `Authority And Entry Gates` section (lines 109-134) establishes a non-negotiable dependency ordering: no behavior-moving slice may start until it depends on a closed, validated Slice 0 bead. This correctly prevents premature refactoring before the guard infrastructure is fully baseline-integrated.
- **Explicit Separation of Concerns:** The reconciler/runtime split matrix (lines 452-495) cleanly isolates state-machine transitions and state eligibility in `internal/session` from reconciler policy decisions (like budgets, alerts, and desired-state pool scaling). This prevents domain-model bloat and keeps the core logic highly cohesive.
- **Strict Read-Path Repair Prohibition:** Rejecting side-effect mutations on query-side lookups (lines 246-251) and requiring the classifier to return a typed `repair-needed` result is correct. This directly addresses the risk of silent read-path mutation hiding persistence failures.

---

### Critical risks:

#### 1. [Blocker] Cosmetic Exception Expiries with Zero Build-Enforced Forcing Function
- **Evidence:** `DESIGN.md` lines 181-184 (validators list), lines 319-322 (guard failure triggers).
- **Why it matters:** The design relies heavily on "expiring exception rows" in `WORKER_BOUNDARY_EXCEPTIONS.yaml` and `SESSION_BOUNDARY_SYMBOLS.yaml` to ensure the bypass list actually shrinks over time. However, the validator failure criteria (lines 181-184) still do **not** fail the build when an exception row is past its stated expiry, nor when a writer persists after its retirement condition is met. Without an automated, build-breaking ratchet for expired exceptions, the massive existing baseline (~16 raw lifecycle writers, 16 `RepairEmptyType` sites, and multiple patch-application loops) becomes a permanent bypass list.
- **Mitigation:** Slice 0 validators must explicitly fail the build if:
  1. Any exception row's `expiry` date is in the past relative to current system time.
  2. The total number of exception/allowlist rows increases relative to the main branch (VCS/git-diff ratchet).

#### 2. [Major] Multi-Writer Concurrency and "Split-Brain" Risks During Transition
- **Evidence:** `DESIGN.md` lines 418-422 (Worker/API exceptions), and lines 345-350 (Coexistence fences).
- **Why it matters:** The design permits transitional "exceptions" where legacy raw writers and new fenced command appliers coexist. Specifically, `internal/api/session_resolution.go` constructs direct `session.Manager` instances and writes metadata directly, while new commands will utilize conditional updates. If a legacy writer performs a blind write (such as retiring an identity or updating a state) concurrently with or after a fenced write, it can silently clobber the fenced writer's updates.
- **Mitigation:** The design must enforce that *no key family* may have active concurrent writers from both the legacy and the fenced command paths. The migration of any key family must be atomic across all active production surfaces.

#### 3. [Major] Vague Static Guard Matching Mechanism and Over-Matching False Positives
- **Evidence:** `DESIGN.md` lines 319-322 and lines 331-332.
- **Why it matters:** The static guard matching mechanism remains under-specified. If `TestSessionBoundaryGuard` performs syntactic string matching on `SetMetadata` or `SetMetadataBatch` to prevent raw session-key writes, it will over-match and fail on legitimate non-session metadata writes in the same files (such as `gc.routed_to`, `workflow_id`, or task-specific metadata on other domains). Conversely, if it is too loose, it can be bypassed by wrapping the store or using dynamic map construction.
- **Mitigation:** Clarify that the key-family column in the symbols manifest is descriptive metadata for ownership, not the matcher. The linter must match by:
  1. Forbidding external package imports of raw patch-map symbols and constructors (e.g. `ClosePatch`, `RetireNamedSessionPatch`, `MetadataPatch`).
  2. Maintaining a strict, file-and-line-level exception inventory of approved raw `SetMetadata*` writers.

#### 4. [Major] Lack of Type-Level Closure Commit for Slices 3-5
- **Evidence:** `DESIGN.md` lines 334-344 (guard scope) and lines 671-676 (backlog).
- **Why it matters:** The design relies entirely on the static guard to prevent external patch map application, but it never commits to a type-level closure. As long as `type MetadataPatch map[string]string` remains a transparent alias, external callers can bypass constructors with raw map literals that static guards cannot easily catch.
- **Mitigation:** The backlog must explicitly commit that Slices 3-5 will **unexport** the individual `*Patch` constructors and make `MetadataPatch` an opaque, package-private type applied only through session-owned methods.

---

### Missing evidence:
- **Unhomed Identity Assignment Writes:** Slices 1-2 are read-only, slice 4 is identity *retirement*, but no slice in the backlog (lines 654-680) is homed to own identity *assignment* (the two `session_name` writers in `cmd/gc/session_name_lookup.go:227` and `cmd/gc/session_beads.go:1119`).
- **Unmapped `RepairEmptyType` Sites:** There are exactly 16 call sites for `RepairEmptyType` across the codebase (including 2 in `internal/mail/beadmail/beadmail.go` and 4 in huma commands). The design-before document does not provide a concrete backlog mapping or completion criteria for retiring these sites.
- **No Concrete Backlog Completion Criteria:** The backlog slices name high-level concepts but provide no concrete, file-or-symbol-level completion criteria to determine when a slice is complete and can be merged.

---

### Required changes:
1. **Automate Validator Expiry Failures:** Add "ledger/exception row past its stated expiry" to the Slice 0 validator failure criteria (lines 181-184), making expiry dates build-enforced.
2. **Commit to Type-Level Closure:** Explicitly commit in the backlog/Atomic Command text that Slices 3-5 will unexport the patch constructors and make `MetadataPatch` opaque.
3. **Map and Home Identity Assignment:** Create or assign a backlog slice to own and migrate the two raw `session_name` identity-assignment call sites.
4. **Enumerate `RepairEmptyType` Retires:** Enumerate the 16 `RepairEmptyType` sites in the symbols ledger and define their retirement slice/criteria in the backlog.
5. **Detail the Guard Matcher:** Explicitly define the matching strategy of `TestSessionBoundaryGuard` to prevent false positives on non-session bead metadata writes.

---

### Questions:
- If a build fails due to an expired exception row in a critical path, what is the approved emergency procedure?
- Is there any technical justification for keeping `MetadataPatch` exported in the final state, or is the end-goal a complete type-level closure?

---

## Answers to Persona Questions

### 1. Which production call sites still write lifecycle or identity metadata directly, and does each backlog slice name the call sites it removes?
**Answer:** There are extensive direct lifecycle and identity metadata writers remaining, including:
- Direct metadata writes (such as `state` and `pending_create_claim` in `cmd/gc/build_desired_state_test.go:4279-4285`).
- External patch applications of `ClosePatch` (e.g. `cmd/gc/session_reconciler.go:206` and `cmd/gc/session_lifecycle_parallel.go:1823`).
- External applications of `RetireNamedSessionPatch` (e.g. `cmd/gc/session_beads.go:444`, `:514`, `:1502` and `internal/api/session_resolution.go:171`).
The backlog slices (lines 654-680) do **not** name the specific call sites or files they remove; the mapping is deferred to the Slice 0 symbols manifest, leaving the backlog too abstract to verify completion.

### 2. What failing-build guard prevents new non-session production code from applying session patch maps or SetMetadata lifecycle writes?
**Answer:** The design proposes `TestSessionBoundaryGuard` under `cmd/gc`, which is intended to fail the build on unauthorized external metadata writes. However, because it lacks type-directed analysis, there is a severe risk of either false positives (flagging non-session bead writes) or false negatives. The guard's matching mechanism remains under-specified and must be clarified in Slice 0.

### 3. How are doctor, migration, repair, and test exceptions bounded so ordinary callers cannot self-label into them?
**Answer:** Exceptions are bounded by machine-readable, schema-validated manifests (`WORKER_BOUNDARY_EXCEPTIONS.yaml` and `SESSION_BOUNDARY_SYMBOLS.yaml`) specifying the path, owner, reason, and expiry. However, because the design does not enforce expiry dates at build-time, these bounds are currently advisory and fail to prevent permanent bypasses.
