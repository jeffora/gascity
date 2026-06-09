# Tomas Park — Test-Slicing & Coverage Verifier (Iteration 18 / Attempt 18, Independent DeepSeek V4 Flash Style Review)

**Verdict:** BLOCK

**Scope:** Test-Slicing, Migration Gate Ordering, Coverage Continuity, and Fixture Drift Detection.
**Reviewed design:** Iteration 18 Design Document (`updated_at: 2026-06-09T08:40:42Z`, 3515 lines), comparing with requirements, codebase realities, and historical wave findings.

---

## Executive Summary

As Tomas Park, the **Test-Slicing & Coverage Verifier**, I have conducted an independent, technically rigorous review of the Iteration 18 Design Document (`design.md`). While other peer lanes (such as Nadia Volkov / Behavior Preservation Auditor and Sofia Khoury / Doctor Migration Safety Reviewer) have moved to approval or conditional approval, my lane mandates strict validation of the technical execution, commit-level dependency ordering, and continuous test passability across the entire migration boundary.

My independent audit, styled with DeepSeek V4 Flash precision, reveals that while Iteration 18 makes exceptional architectural progress—especially with the introduction of the authoritative `slice-gates.generated.yaml` contract (§2249) and the behavior witness floor (§1393)—the actual rollout plan contains **two critical blocker contradictions, one major completeness gap, and two significant technical slicing risks** that mathematically guarantee a broken intermediate or unverified state during execution.

Specifically:
1. **Prose Rollout vs. Slice Gate Matrix Gaps (Blocker):** Prose Rollout Slice 4 (§3414) completely omits running the massive process-level and integration-level sharded tests (`make test-cmd-gc-process-parallel` and `make test-integration-shards-parallel`), whereas the Slice Gate Matrix (§1493) explicitly mandates them. Omitting these from the prose rollout steps introduces a severe risk of merging regressed doctor or loading code without running integration suites.
2. **Prose Rollout vs. Matrix Stale-Cache Contradiction (Blocker):** Prose Slice 2 (§3405) mandates *"stale synthetic-cache rejection for retired aliases"*, but the Slice Gate Matrix (§1491) specifies *"stale synthetic cache ignored"*. Rejection (fatal error/gate failure) vs. Ignoring (silent bypass) are mutually exclusive behaviors. This must be reconciled before implementation starts.
3. **Floating / Unassigned First-Pass Evidence Generation (Major Risk):** The crucial first-pass evidence-generation phase (§2279), which produces `behavior-manifest.generated.yaml` and other essential audit assets before any destructive source move, is not assigned to any prose rollout slice in §3372. It is effectively floating without a clear execution boundary in the Gas City repo rollout.
4. **Maintenance-to-Core Folding Duplicate Definition Risk (Slicing Risk):** Commit Step A (§2265) lands Core-owned Maintenance behavior behind existing active owners, while Maintenance remains registered until Step D (§2268). This creates a multi-commit window where duplicate-active definitions are present, violating the strict ZFC constraint against duplicate definitions (§2272) and risking `make test-fast-parallel` failures unless modeled as an atomic per-asset move.
5. **Examples Coexistence Risk (Slicing Risk):** Rewiring the example city in Slice 2 (§3393) while local Maintenance and in-tree Gastown sources are not deleted or retired until Slices 5 and 7 risks breaking `go test ./examples/...` across Slices 2–4.

Accordingly, I must issue a **BLOCK** verdict. Resolving these concerns requires minor text and contract updates to the design document to ensure flawless execution.

---

## Technical Evaluation of Invariant Questions

### Q1. Is the migration sliced so each individual commit keeps `make test-fast-parallel` passing, or does any step require a cross-repo state that only exists after both repos land?
**Verifier Finding: AT RISK (BLOCKING GAPS).**
While the conceptual **Double-Pin Rollout Model** is highly robust, the implementation plan fails to guarantee that every individual commit remains green:
1. **Commit Steps A–D Duplication:** The transition from Step A to Step D (§2263–2271) creates an intermediate commit state where both Core-owned Maintenance and the legacy Maintenance pack are present, leading to duplicate active definitions which break config loading and `make test-fast-parallel`.
2. **Unassigned Core De-Referencing:** Core formulas and dispatch skills that reference Gastown-bound formulas/roles are not systematically assigned to a clean pre-activation or post-activation slice, creating broken intermediate states where reference checks fail.
3. **Prose/Matrix Shard Mismatches:** Prose Slice 4 (§3414) omits the process and integration sharded tests, while Matrix §1493 mandates them. If an implementer only runs `make test-fast-parallel` in Prose Slice 4, they will miss doctor-active and config-repair regressions that are only caught by the process/integration shards.

### Q2. Do rewritten builtinpacks and embed tests assert loaded behavior — orders resolve, formulas parse, dog agent configures — rather than just counting files or checking include paths?
**Verifier Finding: EXCELLENT (MITIGATED BY DESIGN).**
The design’s **Behavior-Oriented Witness Floor** (§1393) and **Behavior Manifest** (§88) are excellent. They explicitly require executing actual witnesses, verifying formula composition, order resolution, and prompt rendering rather than just doing file counts or existence checks.

### Q3. How does the proposed internal/bootstrap testdata core fixture stay in sync with the real internal/packs/core, and what CI gate detects divergence?
**Verifier Finding: SATISFACTORY (MITIGATED BY DESIGN).**
The design decouples bootstrap from production Core entirely, using minimal, inline `fstest.MapFS` mock/synthetic fixtures instead of copying the whole core. This is guarded by `TestBootstrapFixtureIsMinimal` (§917, §1492) and a stale-path scan to detect any dependency drift.

---

## Critical Risks & Gaps (DeepSeek Focus)

### 1. [Blocker] Prose Rollout vs. Slice Gate Matrix Gaps
* **The Contradiction:** Prose Rollout Slice 4 (§3414) specifies running only `make test-fast-parallel` and `go vet ./...` plus specific package tests. However, the Slice Gate Matrix (§1493) explicitly mandates running `make test-cmd-gc-process-parallel` and `make test-integration-shards-parallel` shards that cover controller reload and config repair.
* **The Risk:** An implementer following the step-by-step prose rollout of Slice 4 will skip running the massive process and integration sharded tests. Regressions in daemon reloading and configuration repair will be missed, allowing broken code to land before the full suite is executed in Slice 5.
* **Required Recommendation:** Align Prose Rollout Slice 4 (§3414) with the Slice Gate Matrix (§1493) by explicitly adding `make test-cmd-gc-process-parallel` and `make test-integration-shards-parallel` to its required gates list.

### 2. [Blocker] Prose Rollout vs. Matrix Stale-Cache Contradiction
* **The Contradiction:** Prose Slice 2 (§3405) mandates *"stale synthetic-cache rejection for retired aliases"*, but the Slice Gate Matrix (§1491) specifies *"stale synthetic cache ignored"*.
* **The Risk:** Rejection (raising a fatal error/failing the gate) vs. Ignoring (bypassing silently) are mutually exclusive behaviors. If the system rejects it, it raises a fatal error/rejection. If the system ignores it, it silently bypasses it. This contradiction will lead to divergent test implementations and inconsistent CI gates.
* **Required Recommendation:** Standardize on one behavior. Update both sections to consistently mandate either "rejection" (raising an error) or "ignored" (silently bypassed) for stale synthetic caches.

### 3. [Major] Floating / Unassigned First-Pass Evidence Generation
* **The Contradiction:** Section §2279 ("The first implementation slice produces evidence before any destructive source move... It generates and validates: behavior-manifest.generated.yaml, independent old-tree baseline transcript, role-surface.generated.yaml...") defines a crucial initial phase that must happen *before* any other destructive changes.
* **The Risk:** In Prose Rollout (§3372-3458), there is no "Slice 0" or explicit pre-requisite slice in Gas City assigned to this phase. Slice 1 (§3377) is focused on `gascity-packs` Gastown candidate branch landing, leaving the Gas City first-pass evidence-generation unassigned to any explicit prose rollout slice.
* **Required Recommendation:** Explicitly define "Slice 0: First-Pass Evidence Generation" in the Prose Rollout (§3377) to ensure the generation and validation of `behavior-manifest.generated.yaml`, `role-surface.generated.yaml`, etc., are completed as a strict prerequisite before any destructive changes or pin updates are executed in the Gas City repo.

### 4. [Minor] Maintenance-to-Core Folding Duplicate Definition Risk
* **The Contradiction:** Commit Step A (§2265) lands Core-owned Maintenance behavior behind existing active owners (duplicating them into Core), while the Maintenance pack is not removed from `requiredBuiltinPackNames` until Step D (§2268).
* **The Risk:** This creates a multi-commit window where duplicate-active definitions are present across both packs, directly violating the duplicate-active definitions rule (§2272) and breaking `make test-fast-parallel` and config-loads unless modeled as an atomic single-commit change.
* **Required Recommendation:** Add a constraint to the Commit Steps stating that Steps A–D must be executed as a single, atomic, multi-asset commit or a highly coordinated single-branch change, ensuring no intermediate build is pushed to main with duplicate active definitions.

### 5. [Minor] Examples Coexistence Risk
* **The Contradiction:** Prose Slice 2 (§3393) requires rewiring `examples/gastown` away from local `../maintenance`.
* **The Risk:** The local Maintenance pack and in-tree Gastown sources are not deleted or retired until Slices 5 and 7. This intermediate partial rewiring window poses a high risk of breaking `go test ./examples/...` unless all stale references are handled atomically.
* **Required Recommendation:** Explicitly specify how the examples tree is verified green across the intermediate slices, or mandate that any legacy Maintenance/Gastown cross-imports are updated or skipped atomically during Slice 2.

---

## Required Gates and Rollout Traceability

| Slice | Focus Area | Required Process/Integration Shards | Key Verifier Controls |
| :--- | :--- | :--- | :--- |
| **Slice 0** | Evidence Generation | `make test-fast-parallel` | Generate and validate `behavior-manifest.generated.yaml`, `role-surface.generated.yaml`, and pilot rows (§2279). |
| **Slice 1** | Candidate public Gastown | `gascity-packs` suite | Front-loaded ownership audits; build manifest and wording matrix generators; no Core code deletion. |
| **Slice 2** | Public-pin & packcompat | `make test-fast-parallel` | Ordinary remote install of exact pin; `packcompat` suite runs in hermetic mode; examples rewired away from local packs. |
| **Slice 3** | Core extraction | `make test-fast-parallel` | Move assets to `internal/packs/core`; bootstrap fixture isolation; introduce `core.maintenance_worker` binding. |
| **Slice 4** | Core loading/doctor | `make test-cmd-gc-process-parallel`<br>`make test-integration-shards-parallel` | Preflight integrity; doctor golden/failure-atomic tests; daemon reload and config repair coverage. |
| **Slice 5** | Public activation & folding | `make test-cmd-gc-process-parallel`<br>`make test-integration-shards-parallel` | Atomic activation-pin flip and Maintenance removal; single-commit asset folding; no-Maintenance loader gate. |
| **Slice 6** | Registry & cache cleanup | `make test-cmd-gc-process-parallel`<br>`make test-integration-shards-parallel` | Registry/cache negative tests; retired alias rejection; public pin remote-cache digest check. |
| **Slice 7** | Source deletion & docs | `make test-fast-parallel` | Final source cleanup; docs wording lint and goldens; post-deletion stale-path scan. |

---

## Required Changes for Finalization

1. **Align Slice 4 Gates (Blocker):** Update Prose Rollout Slice 4 (§3414) to explicitly require `make test-cmd-gc-process-parallel` and `make test-integration-shards-parallel` in its list of required gates, resolving the contradiction with Matrix §1493.
2. **Reconcile Stale-Cache Behavior (Blocker):** Update §3405 and §1491 to use identical language ("ignored" or "rejected") for stale synthetic-cache behavior.
3. **Assign Slice 0 for Evidence Generation (Major):** Formally define "Slice 0: First-Pass Evidence Generation" in the Prose Rollout section, assigning the generation and validation of `behavior-manifest.generated.yaml`, `role-surface.generated.yaml`, etc., to an explicit initial rollout gate.
4. **Constrain Commit Steps A-D (Minor):** Add a strict constraint in §2261 requiring Steps A–D to be applied as a single, atomic commit or branch-merge to avoid intermediate duplicate definition errors in CI.
5. **Mitigate Examples Drift (Minor):** Detail the exact mechanism for keeping `go test ./examples/...` passing during the Slice 2-4 transition window when example cities are rewired but local sources are not yet deleted.
