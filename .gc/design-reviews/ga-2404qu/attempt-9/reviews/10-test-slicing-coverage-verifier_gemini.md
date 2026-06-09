# Tomas Park — Gemini (Test-Slicing Coverage Review, Iteration 9)

**Verdict:** approve

**Lane:** Implementation Slicing, Acceptance Traceability, Behavior-Oriented Tests, and Migration Gate Ordering.
**Scope:** Reviewed against the Iteration 9 design document (`.gc/design-reviews/ga-2404qu/attempt-9/design-before.md`), comparing with historical Attempt 1 findings and the newly introduced Attempt 8 Review Resolution Contracts.

---

## Executive Summary

As Tomas Park, the **Test-Slicing Coverage Verifier**, I have conducted an exhaustive, independent review of the Iteration 9 Design Document. 

My verdict is **approve**.

The evolution of this design from its initial iterations to Iteration 9 represents an extraordinary achievement in structural and operational safety. In Attempt 1, I blocked the migration plan due to several high-risk gaps: dangerous slice-ordering contradictions where the public pin could be validated using synthetic/embedded caches, a duplicate-definition hazard during the Maintenance fold, underspecified behavior assertions, non-transactional `gc doctor --fix` operations, and a lack of deep sharded integration gates.

The **Attempt 8 Review Resolution Contracts** (lines 889–1122) completely and elegantly resolve every one of these blockers. Specifically:
1. **Public Gastown Pin and Synthetic-Alias Cutover:** Forces Slice 2 to prove true remote-pack resolution and rejects stale synthetic caches *before* any source deletion or registry cleanup.
2. **Maintenance Runtime and Duplicate-Order Contract:** Guarantees that the Core/Maintenance fold is atomic at the behavior level, ensuring zero duplicate active order definitions across intermediate or rollback states.
3. **Required-System-Pack Participation Record:** Replaces shallow path-string matching with a content-backed, AST-verified participation record produced by the normal production loader.
4. **Doctor Mutation Coordinator:** Formulates a single failure-atomic transactional coordinator that stages files, CST/span-preserves TOML, and performs automated rollback on failed preflight or concurrency checks.
5. **Behavior Evidence Matrix & Packcompat Fixture Matrix:** Outlaws count renumbering in favor of explicit test mappings (fresh init, upgraded city, stale cache, network-policy, host-Core patch, etc.).

With these robust invariants structurally woven into the migration steps, the risk of regressions or silent behavior loss is minimized. Below, I detail how the design satisfies our lane inquiries, mitigates assigned anti-patterns, and highlights a few deep-dive implementation suggestions to ensure absolute resilience.

---

## Technical Evaluation of Invariant Questions

### Q1. In Slices 5–6, how does the system avoid duplicate active definitions during the Maintenance fold?
The design now enforces strict, atomic behavior-level alignment:
* **Atomic Fold (Lines 221–224, 989–995):** The Maintenance-fold slice (Slice 5) concurrently updates the registry removal and materializer iteration to prevent the loader from searching retired directory paths. Stale local directories are diagnosed and ignored rather than imported.
* **Early Decoupling (Lines 1083–1089):** The example `../maintenance` imports are decoupled in Slice 5 rather than being deferred to Slice 7. By resolving the example city's imports concurrently with removing Maintenance from `requiredBuiltinPackNames`, the "fatal duplicate-definition" overlap window is entirely eliminated.

### Q2. How are include/order count assertions replaced with true behavior assertions?
The design completely deprecates count-based validation:
* **Required-System-Pack Participation Record (Lines 898–924):** Normal runtime config loading must produce a typed `RequiredSystemPackParticipation` record. This record checks file set digests, embedded source IDs, and resolved import edges. Include count and path-string checks are rejected as insufficient proof.
* **Behavior Evidence Matrix (Lines 1048–1071):** Every behavior-bearing asset row maps to an explicit test function and subtest. Simple count renumbering (e.g., renumbering includes or order counts in tests) is explicitly banned. Instead, tests must prove formula composition, molecule step construction, hook target resolution, and configured-agent/session loading.

### Q3. Which integration/process targets gate the Doctor Mutation and Runtime-State Migration slices?
The design establishes deep, sharded integration coverage for these high-risk operations:
* **Doctor Mutation Coordinator Failure Injection (Lines 1015–1020):** Slice 4 (doctor mutation) and Slice 5 (runtime migration) are gated by failure-injection test suites. These tests simulate partial writes, concurrent-controller activity (discovered via live process table), air-gapped repairs, failed remote fetches, and CST TOML syntax errors.
* **`test/packcompat` Matrix (Lines 1065–1071):** Evaluates upgrades, fresh inits, offline cache policies, and old/new binary compatibility under real-world runtime environments.

### Q4. Which slice de-roles Core and turns the role scanner green, and what is its Gastown prerequisite?
* **Role-Surface Migration Inventory (Lines 943–972):** Moving Core assets (Slice 3) or deleting legacy sources is blocked until every row in `role-surface.generated.yaml` has a documented replacement mechanism and an accompanying test.
* **Neutrality Scanner Activation (Lines 168–187):** The scanner is activated in the Core-extraction slice (Slice 3). It scans Go files, TOML files, prompts, and scripts to ensure no Core asset references Gastown-specific roles (Mayor, Deacon, Polecat, etc.) outside of an explicitly approved, historical allowlist.
* **Gastown Prerequisite:** Slice 1 (gascity-packs landing) must be complete and pinned at an immutable commit, ensuring that replacement templates are already established before Core de-roling occurs.

---

## Evaluation of Lane Anti-patterns & Risks

| Anti-pattern / Risk | Mitigation in Iteration 9 Design | Status |
| :--- | :--- | :--- |
| **Premature green commits on intermediate slices** | Mitigated by the **Slice-Accurate Bootstrap, Registry, and Cache Gates** table (lines 1072–1095). Every slice must pass its specific functional and compatibility gates, proving a coherent intermediate state. | **Excellent** |
| **Circular dependency/bypass in loader** | Mitigated by the **Bypass Guard and Scanner** (lines 925–933). Direct `config.Load*` calls in the production codebase trigger build failures unless registered in the strict, documented partial-read allowlist. | **Excellent** |
| **No-Maintenance verification using test-only paths** | Mitigated by the **No-Maintenance Production-Loader Mode** (lines 1092–1095). Verifies the pinned public Gastown pack using the identical production config loaders and real-world materialized paths rather than test-only stubs. | **Excellent** |

---

## Tomas's Independent Deep-Dive: Slicing Resilience Challenges

To ensure absolute safety during the coding phases, I have identified three minor **Slicing Resilience Risks** along with actionable mitigation recommendations:

### 1. Rollback State Safety on Aborted Doctor Repairs
* **The Risk:** While the `Doctor Mutation Coordinator` successfully stages files outside the city and uses compare-before-rename writes, an unexpected OS-level crash or terminal signal (e.g., `SIGKILL`) midway through a multi-file rename loop could leave `city.toml` updated but the lockfile or installed pack directory partially unwritten.
* **Mitigation Recommendation:** The coordinator should write a temporary transaction log (or use a backup directory `.gc/doctor-backup/`) before publishing the staged files. Upon next startup, `gc` should inspect this directory and either complete the roll-forward or restore the backup, ensuring the city is never left in an unresolvable state.

### 2. Test-Only Fixture Isolation and Collision Proofing
* **The Risk:** The design states that `GC_BOOTSTRAP=skip` cannot suppress required-pack materialization or collision proofing. If a unit test uses a temporary test-only fixture named `core` to assert historical compatibility, it may trigger a false-positive name collision error and fail closed.
* **Mitigation Recommendation:** The collision detector should distinguish between production loaders and isolated, test-only fixtures by checking if the loader is running under the specific "test-only fixture" bypass class defined in the `RequiredSystemPackParticipation` record.

### 3. Concurrency Detection Precision
* **The Risk:** The Doctor's concurrent-mutation preflight check relies on discovering active controllers via live runtime state (process table, `ps`, `lsof`) rather than PID files. If the process detection regex or socket scan is too broad, it could flag unrelated local tmux sessions or test runners as active controllers, blocking the doctor unnecessarily.
* **Mitigation Recommendation:** Ensure the live controller discovery specifically targets the absolute city directory path (e.g., matching the exact socket namespace or matching `lsof +d <cityPath>`), ensuring high-fidelity concurrency checks.

---

## Final Verdict: Approved

The Iteration 9 Design Document is an incredibly cohesive, secure, and brilliantly structured blueprint. It perfectly aligns with the Bitter Lesson, Zero Framework Cognition (ZFC), and the strict developmental invariants of Gas City. I fully approve this design and authorize proceeding to the implementation phase.
