# Elena Marchetti — Gemini (Independent Review, Attempt 7 / Iteration 7)

**Verdict:** block

**Review focus:** Mutation boundary enforceability, session lifecycle/identity write ownership, patch-map escape hatches, canonical writer inventory completeness, and build-time static guard coverage.

---

## Fix Validations
No functional bug-fix validation is applicable to this design-phase review. We are evaluating the technical design and architectural consistency of the Session Boundary Design document (`internal/session/DESIGN.md` / `attempt-7/design-before.md`) before decomposition.

---

## Findings

### 3) Change Impact / Blast Radius
- **Severity:** blocker
- **Confidence:** high
- **Quality dimension:** correctness
- **Gate impact:** blocker
- **Evidence:** `internal/session/DESIGN.md:630-639` (`W-026`, `W-029`, `W-030`, `W-031`)
- **Pattern comparison:** `W-004`, `W-005` (which map to single concrete command targets).
- **Why it matters:** The design's exit gate (lines 115-133) mandates that "each backlog slice name the call sites it removes and their exact retirement conditions." However, the inventory table baseline uses extremely broad rows. For instance, `W-029` captures "legacy session REST handlers" in a single broad row, and `W-030` captures "all bead routes that can target session beads" as a generic "per touched slice" entry. Without breaking these down into symbol-level, file-specific, and key-specific rows prior to decomposition, developers will lack a concrete roadmap of what call sites are being converted or retired in each slice, leading to un-audited map-updates and permanent bypasses.
- **Suggested fix:** Fully expand the planning baseline table to include dedicated sub-rows for each concrete endpoint inside `handler_sessions.go` (e.g., `/v0/session/{id}/suspend` under a dedicated `W-029a`, `/v0/session/{id}/close` under `W-029b`), and itemize the exact functions inside `handler_beads.go` (e.g., `resolveSessionAssignee`) that perform silent empty-type repairs.

### 10) Architectural Consistency
- **Severity:** blocker
- **Confidence:** high
- **Quality dimension:** correctness
- **Gate impact:** blocker
- **Evidence:** `internal/session/DESIGN.md:658-693` (Static Guard description)
- **Pattern comparison:** `cmd/gc/worker_boundary_import_test.go` (which is a simple import-dependency linter).
- **Why it matters:** The proposed static guard test (`cmd/gc/session_boundary_guard_test.go`) is tasked with "flagging direct store writes when the receiver is an interface value that can carry a session bead" and inspecting dynamic keys. In Go, pure AST-based static checkers (like those built with `go/ast` or `go/analysis`) cannot perform inter-procedural dataflow analysis or type-narrowing to determine if a dynamic string variable or an interface wrapper (e.g. `beads.Store`) actually reaches a session bead at runtime. Without a committed proof-of-concept or a realistic false-positive budget, this gate is highly speculative. It risks either causing massive build-time noise (flagging every generic bead-store write in the system) or failing silently (missing actual session mutations).
- **Suggested fix:** Provide a skeletal or prototype implementation of `session_boundary_guard_test.go` that demonstrates how the AST parser resolves interface-receiver types and discriminates session bead targets (e.g., via AST-level type assertion or annotation tagging) before approving this gate.

### 10) Architectural Consistency
- **Severity:** blocker
- **Confidence:** high
- **Quality dimension:** maintainability / security
- **Gate impact:** blocker
- **Evidence:** `internal/session/DESIGN.md:679-681` ("Require shrink-only allowlist behavior")
- **Pattern comparison:** Standard Go testing framework.
- **Why it matters:** The design requires that `allowlist.yaml` exhibits "shrink-only" behavior and that rows carry a retirement `expiry`. However, a static unit test (`TestSessionBoundaryGuard`) executed on a local checkout has no temporal memory or access to git history. If a developer simply appends a new exception row or pushes an expired exception's date forward in their local branch, the test will pass because the allowlist is evaluated in isolation. Without a CI-level check comparing the allowlist against the merge-base (e.g., `origin/main`), the "shrink-only" and "expiry" rules are merely prose guidelines, not executable build gates. This directly risks the bypass list growing indefinitely (Red Flag #3).
- **Suggested fix:** Explicitly require that the Slice 0 preflight includes a CI workflow step or a git-aware test (e.g., invoking `git diff` or checking against a mainline reference) that fails the build if the number of allowlist entries increases or if any entry's `expiry` date is in the past.

---

## Consistency Report
- **Patterns checked:**
  - Alignment of "Canonical Production Writer Inventory" table against custom/dynamic keys and files.
  - Feasibility of AST type-narrowing over interface receivers (`beads.Store` methods).
  - Validation of local vs CI execution limits for "shrink-only" allowlist enforcement.
- **Sibling files checked:**
  - `internal/session/REQUIREMENTS.md` vs `internal/session/DESIGN.md`.
  - `internal/session/lifecycle_projection.go` BaseState definitions vs DESIGN.md key families.
- **Propagation verified:**
  - Verified that all previously reported gaps (such as `handler_sessions.go` and `handler_beads.go:89`) are now listed as broad W-rows in `DESIGN.md`, but noted that their lack of symbol-level specificity creates a lingering structural gap.
- **Drift detected:**
  - Local unit test scope drifts from the described dynamic "shrink-only" constraint which can only be safely enforced at the CI/VCS level.
