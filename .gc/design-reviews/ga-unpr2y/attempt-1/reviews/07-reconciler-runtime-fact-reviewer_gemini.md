# Liam Okonkwo — DeepSeek V4 Flash

**Verdict:** block

Lane: reconciler boundary, runtime intent adapter ownership, fact isolation, health gate split. This reviews the current `DESIGN.md` (the attempt-16 `iterate`-response revision, located at `.gc/design-reviews/ga-unpr2y/attempt-17/design-before.md`), alongside `REQUIREMENTS.md` and the existing codebase. Findings are validated against the live check-out on this branch; inline citations and code-level references are provided below.

---

### Top strengths:
- **Comprehensive Boundary Matrix Schema (lines 580–595):** The detailed matrix requirements—enforcing source ownership, policy ownership, freshness rules, and pure-decider import/call guards—provide a strong foundation for isolating internal session logic from volatile controller and provider states.
- **Strict Separation of Reconciler and Session Policies (lines 553–563):** The split matrix is exceptionally well-defined. It restricts `internal/session` to pure lifecycle eligibility and identity classification over immutable facts, while reserving the controller/reconciler's ownership of work demand, dispatch scheduling, pool sizing, and progress policy.
- **Provider-Neutral Intent Separation (lines 564–568):** Correctly limiting `RuntimeIntent` to provider-neutral fields (session ID, config hash, transport runtime identity) preventing the smuggling of provider-specific scheduling or health policy into `internal/session`.

---

### Critical risks:

#### 1. [Blocker] Fact Isolation Compromised: Load-bearing `ProjectLifecycle` clock fallback inside active codebase
- **Evidence:** `internal/session/lifecycle_projection.go:381` and `609` (clock fallback: `now = time.Now().UTC()`).
- **Why it matters:** The design mandates that pure session deciders receive immutable facts and do not perform ambient clock reads (`DESIGN.md:592`, `603-606`). However, `ProjectLifecycle` (the central, load-bearing projection function) still contains local wall-clock fallbacks when `input.Now` is zero:
  ```go
  now := input.Now
  if now.IsZero() {
  	now = time.Now().UTC()
  }
  ```
  If a caller fails to pass an explicit timestamp, the decider yields non-deterministic results based on the local OS clock, destroying test replayability and violating decider purity. Fact isolation must be call-level and absolute.
- **Required Change:** Completely remove the `time.Now().UTC()` fallbacks from `lifecycle_projection.go`. Enforce `input.Now` as a mandatory, non-zero field, returning an error or failing fast if it is missing. Redefine any static AST guards to inspect pure decider files to reject direct calls to `time.Now()`.

#### 2. [Blocker] Slice 0 Preflight and Minimum Proof Command Remain Pure Prose and Non-Executable
- **Evidence:** `DESIGN.md` lines 218-222 (the minimum proof command) and the active checkout state (lack of Go test files).
- **Why it matters:** The design proposes a minimum proof command:
  ```bash
  go test ./cmd/gc ./internal/session ./internal/api ./internal/events -run 'TestSlice0Contract|TestSessionBoundaryGuard|TestSessionBoundaryInventoryFresh|TestSlice0Artifacts|TestScenarioParityFreshness|TestVocabularyCheckpoints|TestSessionDiagnosticsManifest|TestSessionCommandApplierLedger|TestSessionStoreCapabilityMatrix|TestSessionBoundaryMatrix|TestSessionRouteInventoryFresh|TestWorkerBoundaryExceptionLedger|TestEveryKnownEventTypeHasRegisteredPayload' -count=1
  ```
  However, **none** of the referenced tests (such as `TestSessionBoundaryGuard`, `TestSlice0Contract`, `TestSessionBoundaryInventoryFresh`, etc.) exist in the Go codebase. Citing a minimum proof command containing 12+ non-existent test targets is a cosmetic success that would fail immediately in CI. Leaving this critical defense boundary as pure prose provides no physical protection against pattern drift or regressions as work proceeds.
- **Required Change:** Slices 0-1 cannot decompose or proceed until the physical Go test files, AST guards, and test skeletons are implemented and passing on the active branch.

#### 3. [Major] Naive Reconciler Fact Compilation Hot-Loop Cost and lack of Store-Scan Budgets
- **Evidence:** `DESIGN.md` lines 697-701 (cost rules), lines 713-714 (reconciler fact compilation budget requirement).
- **Why it matters:** The pure-decider boundary requires that all facts (such as work counts, pool size, active work beads) are pre-scanned and compiled before being passed into session-owned deciders (`DESIGN.md:603-606`). In a city with hundreds of sessions, pre-compiling beads/work counts for every single session in every tick of the reconciler loop before passing them to deciders will absolutely cripple performance if done naively via individual database queries or subprocess scans. Although the design requires a "budget row" for reconciler fact compilation, it does not specify a physical implementation mechanism (e.g. batch-querying the bead store, leveraging memory indices, or caching) to prevent a catastrophic $O(N)$ performance scaling issue where $N$ is the number of active sessions.
- **Required Change:** The design must explicitly commit to a *batch fact compilation protocol* where the reconciler fetches all active session and work beads in a single query/scan, rather than compiling facts individually per session.

#### 4. [Major] Runtime-Missing Anti-Flap Rule Lacks Concrete Cooldown or Verification Protocol
- **Evidence:** `DESIGN.md` lines 593-594 (runtime-missing anti-flap rule in boundary matrix).
- **Why it matters:** Gascity forbids status files, mandating that process liveness is discovered by querying the live system (`ps`, `lsof`, port scans). Under heavy system load, transient query failures or rate-limiting can cause `ps` to return zero running processes. If the decider naively treats a transient "missing observation" as definitive proof that the session has crashed or stopped, it will write a durable session state mutation (`session.crashed` or `session.stopped`) to the Bead Store, triggering premature teardown or recovery loops (split-brain). While the design mentions a "runtime-missing anti-flap rule", it remains purely abstract (a list of fields to include in the boundary matrix) rather than a concrete, physical protocol.
- **Required Change:** Specify a concrete, physical cooldown window (e.g. minimum 3 ticks or 30 seconds of consecutive missing observations) and double-verification protocol (e.g. multiple distinct probe commands) before a missing observation can be written as durable session truth.

#### 5. [Major] Smuggling Provider Policy into `RuntimeIntent` through Implicit Destructive Action Rules
- **Evidence:** `DESIGN.md` lines 564-569 (RuntimeIntent definition) and lines 576-578 (destructive-action safety rule).
- **Why it matters:** The design states that `RuntimeIntent` must remain provider-neutral (no provider-specific scheduling or policy). However, destructive actions under "unknown, stale, partial, or provider-error runtime facts" require a "safe convergence rule". If the decider has to apply a generic convergence rule without knowing the provider type (e.g., tmux local vs k8s cluster), it has to choose between failing closed (orphaning resources) or failing open (risking dual-spawns). Translating provider-specific failures into provider-neutral intent actions forces `RuntimeIntent` to either smuggle provider policy or make unsafe assumptions.
- **Required Change:** Explicitly specify that the *provider adapter* (not the internal decider) owns the final execution logic and policy when encountering partial or provider-error states, keeping the internal decider strictly query-neutral.

---

### Missing evidence:
- **Pure Decider AST Guard Test:** An automated AST parser test in `internal/session` that scans the pure-decider files and fails the build if any direct call to `time.Now()` or store-query patterns are found.
- **Reconciler Batch Fact Compilation Proof:** Detailed budget specifications or mock implementations demonstrating how the reconciler can avoid individual store scans for each session on hot loops.
- **Physical Slice 0 Test Skeletons:** The Go test code files defining the minimum proof targets (`TestSessionBoundaryGuard`, `TestSlice0Contract`, etc.).

---

### Required changes:
1. **Enforce Absolute Decider Purity:** Remove `time.Now().UTC()` clock fallbacks from `lifecycle_projection.go`, making `input.Now` mandatory and non-zero.
2. **Materialize Slice 0 Test Targets:** Implement the physical Go test files for the minimum proof targets in `cmd/gc/`, `internal/session/`, and other relevant packages.
3. **Specify Batch Compilation:** Explicitly commit to a single, batch-based fact query compilation step in the reconciler to prevent $O(N)$ database scan overhead.
4. **Define Cooldown for Missing Observations:** Codify the runtime-missing anti-flap rule with a concrete cooldown threshold and verification protocol.
5. **Decouple Destructive Policy from Decider:** Ensure that the provider adapter holds absolute authority over handling transient provider errors during destructive operations.

---

### Questions:
- Is there any plan to support batch-queries in the `bdstore` package to facilitate low-cost pre-compilation of reconciler facts?
- Should the physical cooldown threshold for missing process observations be a configurable value in `city.toml` or remains a hardcoded threshold?

---

## Answers to Persona Questions

### 1. Which wake, hold, drain, provider-health, and progress decisions move into session deciders, and which scheduling or budget responsibilities remain in the reconciler?
**Answer:**
- **Move to session deciders:** Basic eligibility and transition rules, including determining terminal states, wake blockers, identity conflicts, and hold/drain timeouts based on immutable fact inputs.
- **Remain in reconciler:** Fact compilation, work-demand aggregation, pool sizing, dispatch scheduling, progress policy, restart/rollback budgets, and orchestrating destructive actions.

### 2. Are work counts, pool size, runtime liveness, and progress facts precomputed by adapters instead of queried from deciders?
**Answer:** Yes. Session-owned deciders are strictly pure and do not perform any I/O, database queries, or external process checks. All necessary facts, such as work counts, pool sizing, and observed runtime liveness, must be pre-scanned and compiled by reconciler/caller adapters and passed as copyable structures to the decider.

### 3. Can RuntimeIntent express adapter needs without smuggling provider policy into `internal/session`?
**Answer:** Yes. `RuntimeIntent` represents a provider-neutral declarative state (e.g. session ID, config hash, transport runtime identity) that states *what* is intended. The downstream runtime provider adapters (such as tmux or subprocess) consume this intent and translate it into provider-specific actions and policies, ensuring `internal/session` remains completely decoupled from provider details.

---

## Consistency Report

- **Pattern Alignment (with Elena Marchetti - Wave 1, Mutation Boundary Auditor):**
  - We strongly align with Elena's finding that the AST static guards are under-specified and currently exist only as prose. Citing non-existent test files like `TestSessionBoundaryGuard` in the minimum proof command is a massive consistency issue.
  - We also agree with Elena's identification of the multi-writer concurrency and split-brain risks during transition; our finding on the lack of a concrete cooldown/anti-flap rule for transiently missing runtime observations highlights a similar split-brain risk on the query-and-observe boundary.
