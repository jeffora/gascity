# Elena Marchetti - DeepSeek V4 Flash

**Verdict:** block

Lane: session lifecycle ownership, external metadata write audit, patch-map escape hatches, CI guard coverage. This reviews the current `DESIGN.md` (the attempt-19 revision, located at `.gc/design-reviews/ga-unpr2y/attempt-19/design-before.md`), alongside `REQUIREMENTS.md` and the existing codebase. Findings are validated against the live check-out on this branch; inline citations and code-level references are provided below.

---

### Top strengths:
- **Outstanding Iterative Design Rigor:** The comprehensive transition table in the "Attempt 17 Review Response" (`DESIGN.md:54-81`) shows excellent engineering discipline, tracking and addressing previous feedback point-by-point.
- **Smarter, Practical Staging of Slice 0:** Splitting Slice 0 artifacts into a minimal Slice 1 close gate versus schema-only, non-normative inventories (`DESIGN.md:220-237`) is a highly pragmatic staging decision. This avoids blocking initial read-only progress with unrelated downstream validation details.
- **Quarantining of Side-Effect Repair:** Formally classification-blocking `RepairEmptyType` from read-only lookup paths and classifying it as a diagnostic `repair_pending` (`DESIGN.md:359-368`) prevents silent state mutation during simple query lookups.

---

### Critical risks:

#### 1. [Blocker] Manifests and Exception Ledgers Exist Only in Prose
- **Evidence:** `DESIGN.md` lines 200-218 (Universal Slice 0 artifacts), lines 625-637 (Huma routes / exception ledger).
- **Why it matters:** The design hinges entirely on machine-readable manifests (`WORKER_BOUNDARY_EXCEPTIONS.yaml`, `SESSION_BOUNDARY_SYMBOLS.yaml`, `MIGRATION_COHABITATION.yaml`) to act as CI gates and prevent bypasses. However, a physical inspection of the workspace reveals that **none of these files exist** in the current checkout. An unmaterialized manifest cannot prevent drift or enforce boundaries; the design remains a paper tiger until these files are committed and validated.

#### 2. [Blocker] Absence of VCS-Anchored Time / Drift Protections in Build-Enforced Expiries
- **Evidence:** `DESIGN.md` lines 210-216, lines 263-267 ("expired exception row... relative to current system time").
- **Why it matters:** The design mandates that Slice 0 validators fail if any exception row's `expiry` date is in the past relative to "current system time". Relying on ambient system clock time during local `go test` runs or pre-commit git hooks introduces a critical vulnerability: developers can bypass the build failure simply by drifting their local system clock backward, or isolated offline/CI environments with drifted clocks may falsely pass/fail.
- **Mitigation:** Expiry validation must be anchored to a deterministic, monotonically increasing source of time within the repository (such as the timestamp of the latest git commit or a pinned VCS anchor timestamp), rather than raw ambient local system clock.

#### 3. [Major] Severe Store-Level Race Vulnerability (TOCTOU) Across CLI/API/Controller Processes
- **Evidence:** `DESIGN.md` lines 505-516 (coexistence fences) and lines 576-586 (store capability matrix).
- **Why it matters:** `DESIGN.md` explicitly notes that `bdstore` "exposes separate read and write calls with no compare-and-swap, expected-revision, or token primitive, so no fence may be assumed". Despite admitting that "reread then write" is not a valid strategy due to TOCTOU races, the design still permits transitional cohabitation where legacy blind writers and new fenced command appliers coexist. Without store-level atomic primitives, any transaction or conditional write is an illusion, risking silent state clobbering across concurrent CLI/API/controller processes.
- **Mitigation:** Reject any transitional coexistence for a given key-family. The migration of any key family must be atomic across all active production surfaces, or store-level CAS/transaction primitives must be prioritized.

#### 4. [Major] Lack of Automated Linter Validation for Key Families on Direct Metadata Writes
- **Evidence:** `DESIGN.md` lines 603-612 (routing rules and exceptions).
- **Why it matters:** Direct metadata writes (e.g., `state` in `cmd_session_wake.go`) are planned as exceptions, but without an AST-based linter that parses Go code to check for specific key assignments, developers can easily add new unapproved direct writes by mimicking existing patterns.
- **Mitigation:** The static guard matcher (`TestSessionBoundaryGuard`) must parse the AST of external packages to explicitly verify that any call to `SetMetadata` or `SetMetadataBatch` targeting session key-families is blocked unless its file and line are registered in `WORKER_BOUNDARY_EXCEPTIONS.yaml`.

---

### Missing evidence:
- **No Concrete Backlog Mapping for `RepairEmptyType`:** While the design states that `RepairEmptyType` is retired from read-only paths, it still fails to list the 14 existing call sites or provide a concrete backlog mapping for their final retirement.
- **No VCS Anchor in Expiry Validation:** No evidence is provided showing how build-enforced expiries prevent clock drift or local clock manipulation.

---

### Required changes:
1. **Materialize the Draft Manifests:** Commit the initial schemas or empty instances of `WORKER_BOUNDARY_EXCEPTIONS.yaml`, `SESSION_BOUNDARY_SYMBOLS.yaml`, and `MIGRATION_COHABITATION.yaml` to the repository so the validators can run against actual files instead of prose.
2. **Anchor Clock Validation to VCS:** Update the validator clock rule (`DESIGN.md:263-267`) to compare row expiry dates against a deterministic repository metadata timestamp (e.g., the last commit time) instead of ambient system time.
3. **Commit to Atomic Store CAS/Transactions:** Prioritize adding a transaction or conditional-update primitive to the store before any mutating session-owned commands are merged.
4. **Mandate AST-Based Key-Family Linter:** Explicitly define that `TestSessionBoundaryGuard` will use AST parsing to identify and block unauthorized direct writes to session-owned key families.

---

### Questions:
- If a build fails due to an expired exception row in a critical offline CI environment, what is the approved emergency procedure?
- Is there an active ticket to introduce atomic compare-and-swap (CAS) primitives to `bdstore` before Slice 3 begins?

---

## Answers to Persona Questions

### 1. Which production call sites still write lifecycle or identity metadata directly, and does each backlog slice name the call sites it removes?
**Answer:** There are extensive direct lifecycle and identity metadata writers remaining, including:
- Direct metadata writes (such as `state` and `pending_create_claim` in `cmd/gc/build_desired_state_test.go:4279-4285`).
- External patch applications of `ClosePatch` (e.g., `cmd/gc/session_reconciler.go:206` and `cmd/gc/session_lifecycle_parallel.go:1823`).
- External applications of `RetireNamedSessionPatch` (e.g., `cmd/gc/session_beads.go:444`, `:514`, `:1502` and `internal/api/session_resolution.go:171`).
The backlog slices (lines 654-680) do **not** name the specific call sites or files they remove; the mapping is deferred to the Slice 0 symbols manifest, leaving the backlog too abstract to verify completion.

### 2. What failing-build guard prevents new non-session production code from applying session patch maps or SetMetadata lifecycle writes?
**Answer:** The design proposes `TestSessionBoundaryGuard` under `cmd/gc`, which is intended to fail the build on unauthorized external metadata writes. However, because it lacks type-directed analysis, there is a severe risk of either false positives (flagging non-session bead writes) or false negatives. The guard's matching mechanism remains under-specified and must be clarified in Slice 0.

### 3. How are doctor, migration, repair, and test exceptions bounded so ordinary callers cannot self-label into them?
**Answer:** Exceptions are bounded by machine-readable, schema-validated manifests (`WORKER_BOUNDARY_EXCEPTIONS.yaml` and `SESSION_BOUNDARY_SYMBOLS.yaml`) specifying the path, owner, reason, and expiry. However, because the design does not enforce expiry dates at build-time, these bounds are currently advisory and fail to prevent permanent bypasses.
