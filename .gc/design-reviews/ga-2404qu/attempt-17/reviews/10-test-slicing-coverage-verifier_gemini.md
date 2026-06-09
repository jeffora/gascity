# Tomas Park — Test-Slicing & Coverage Verifier (Iteration 17 / Attempt 17, Independent DeepSeek V4 Flash Style Review)

**Verdict:** BLOCK

**Scope:** Test-Slicing, Migration Gate Ordering, Coverage Continuity, and Fixture Drift Detection.
**Reviewed design:** Iteration 17 Design Document (`updated_at: 2026-06-09T02:00:53Z`, 3220 lines), comparing with requirements, codebase realities, and historical wave findings.

---

## Executive Summary

As Tomas Park, the **Test-Slicing & Coverage Verifier**, I have conducted a rigorous, independent review of the Iteration 17 Design Document (`design-before.md`). While other peer lanes (e.g., Nadia Volkov / Behavior Preservation Auditor) have approved from their narrow perspectives, my lane requires strict validation of the technical execution, commit-level dependency ordering, and continuous test passability across the entire migration boundary.

My independent audit, styled with DeepSeek V4 Flash precision, reveals that while Iteration 17 makes excellent architectural progress—especially with the introduction of the authoritative `slice-gates.generated.yaml` contract and the behavior witness floor—the actual work plan contains **one critical blocker contradiction, two major consistency/completeness gaps, and a minor example-coexistence risk**.

Specifically:
1. **The Core Role-Asset De-Referencing and Removal Slice is Unassigned (Blocker):** Current Core contains extensive references to Polecat, Refinery, Witness, Mayor, and Gastown in formulas and skills (§2512). Although `mol-polecat-*` formulas are re-homed to Gastown in Slice 1, the authoritative rollout plan fails to assign the de-referencing of Core formulas (such as `mol-do-work`) or the split of the `gc-dispatch` SKILL to any specific slice or commit. This mathematically guarantees a broken/failing intermediate state where Core refers to missing/duplicated formulas during slices 2–4.
2. **Multiple Contradictory Rollout and Slicing Representations (Major Risk):** The design document presents at least three overlapping and non-reconciled representations of the slicing: the Attempt-9 Slice-Gate Matrix (§1489–1496), the Attempt-14 Activation Commit Ordering A–F (§2263–2271), and the Prose `## Rollout` Slices 1–7 (§3077–3162). While §2249 declares that `slice-gates.generated.yaml` is the "only binding source of truth," that file is currently unwritten and does not exist. An implementer has no clear, authoritative way to map intermediate commits.
3. **Self-Certified Manifest Discovery (Major Risk):** Behavior manifest completeness rests entirely on the manifest generator's discovery walk (§115–116, §1739). If the generator has a bug or fails to walk certain dynamic scripts or nested helper imports, those assets will be silently omitted from BOTH the generated and checked-in copies, passing all CI gates undetected.
4. **Examples Coexistence Risk (Minor Risk):** The `examples/` tree is expected to stay green during Slices 2–4 (§3111), yet stale local packs still cross-import `../maintenance` or rely on behavior-test assertions against files that are being actively moved or deleted. The transition window's green state remains asserted rather than technically demonstrated.

Accordingly, I must issue a **BLOCK** verdict. Resolving these concerns requires minor text and contract updates to the design document to ensure flawless execution.

---

## Technical Evaluation of Invariant Questions

### Q1. Is the migration sliced so each individual commit keeps `make test-fast-parallel` passing, or does any step require a cross-repo state that only exists after both repos land?
**Verifier Finding: AT RISK (BLOCKING GAPS).**
While the conceptual **Double-Pin Rollout Model** is highly robust, the implementation plan fails to guarantee that every individual commit remains green:
1. **Unassigned Core De-Referencing (Blocker):** Core formulas like `mol-do-work` currently reference Gastown-bound `mol-polecat-*` formulas (§2516). Since the removal/de-referencing step is unassigned, any intermediate commit between the removal of Polecat from Core and the activation of the public Gastown pin will fail formula-composition checks (§3021).
2. **Slice 5 Lack of Intra-Slice Commit Ordering (Major Risk):** Slice 5 bundles several high-risk moves, pin changes, and registry deletions (§3128–3146). Because duplicate-active definitions are fatal (§770) and there is no inactive-asset loader, any naive commit sequence within Slice 5 will break tests. A strict single-commit folding sequence is required.

### Q2. Do rewritten builtinpacks and embed tests assert loaded behavior — orders resolve, formulas parse, dog agent configures — rather than just counting files or checking include paths?
**Verifier Finding: SATISFACTORY (PASSING).**
The design’s **Behavior-Oriented Witness Floor** (§313–318, §1396–1400) is a major strength. It explicitly mandates that simple list checks or file counts are insufficient. Rewritten tests must prove formula composition, molecule step construction, hook target resolution, and pack-relative script execution, providing a robust empirical baseline.

### Q3. How does the proposed internal/bootstrap testdata core fixture stay in sync with the real internal/packs/core, and what CI gate detects divergence?
**Verifier Finding: EXCELLENT (MITIGATED BY DECOUPLING).**
The design elegantly resolves the drift risk of copying on-disk Core directories into `internal/bootstrap` by banning them entirely (§511–515) and mandating strictly minimal, inline `_test.go` `fstest.MapFS` mock/synthetic fixtures. This is paired with the executable `TestBootstrapFixtureIsMinimal` guard gate (§529, §917, §2537) to prevent future dependency accretion.

---

## Critical Risks & Gaps (DeepSeek Focus)

### 1. [Blocker] Unsequenced and Unassigned Core Role-Asset Removal
* **The Contradiction:** The design explicitly dictates that Core must not own any Gastown role behavior, specifically listing `mol-polecat-*` formulas (§2516) and `gc-dispatch` skill references (§2518). However, the prose `## Rollout` (§3077–3162) never assigns the removal/rewrite of these assets from Core to any slice. 
* **The Risk:** If `mol-polecat-*` leaves Core in Slice 1 to land in public Gastown, but `mol-do-work` in Core still references them, formula-composition checks will fail instantly. An asset move will land without a passing intermediate test state.
* **Required Recommendation:** Explicitly assign the in-tree Core role-asset removal and de-referencing of `mol-do-work` (and related skill splits) to a named slice. Sequence it so the referenced formulas are re-homed in public Gastown *before or in the same commit* that removes them from Core.

### 2. [Major] Lack of Reconciliation Between Multi-Slicing Representations
* **The Contradiction:** The design defines three independent and conflicting slicing sequences: the Attempt-9 Matrix (§1489–1496), the Attempt-14 A–F Commit Ordering (§2263–2271), and the Prose Rollout Slices 1–7 (§3077–3162). 
* **The Risk:** There is no crosswalk mapping these sequences. For example, Attempt-14 Step E matches Prose Slice 6, and Step F matches Prose Slice 7, but Steps A–D are bundled in Prose Slice 5 without clear boundaries. This ambiguity guarantees developer confusion during implementation.
* **Required Recommendation:** Add a definitive, canonical crosswalk mapping prose slices 1–7 to the Attempt-14 A–F commit steps and the Attempt-9 matrix. State explicitly which representation is authoritative and why.

### 3. [Major] Self-Certified Behavior Manifest Discovery Completeness
* **The Contradiction:** Traceability and coverage rely on the `behavior-manifest.generated.yaml` (§115, §1739). However, both the generator output and the freshness check run on the same walk.
* **The Risk:** A discovery false-negative (e.g. failing to walk a nested helper or dynamic reference) will result in missing rows that pass all linter checks silently.
* **Required Recommendation:** Add an independent, VCS-level completeness check (e.g. comparing manifest rows against a simple git-level file list of moved or deleted assets) to guarantee 100% coverage completeness in CI.

---

## Required Gates and Rollout Traceability

| Slice | Focus Area | Required Process/Integration Shards | Key Verifier Controls |
| :--- | :--- | :--- | :--- |
| **Slice 1** | Candidate public Gastown | `gascity-packs` suite | Front-loaded ownership audits; build manifest and wording matrix generators; no Core code deletion. |
| **Slice 2** | Public-pin & packcompat | `make test-fast-parallel` | Ordinary remote install of exact pin; `packcompat` suite runs in hermetic mode; examples rewired away from local packs. |
| **Slice 3** | Core extraction | `make test-fast-parallel` | Move assets to `internal/packs/core`; bootstrap fixture isolation; introduce `core.maintenance_worker` binding. |
| **Slice 4** | Core loading/doctor | `make test-cmd-gc-process-parallel`<br>`make test-integration-shards-parallel` | Preflight integrity; doctor golden/failure-atomic tests; daemon reload and config repair coverage. |
| **Slice 5** | Public activation & folding | `make test-cmd-gc-process-parallel`<br>`make test-integration-shards-parallel` | Atomic activation-pin flip and Maintenance removal; single-commit asset folding; no-Maintenance loader gate. |
| **Slice 6** | Registry & cache cleanup | `make test-cmd-gc-process-parallel`<br>`make test-integration-shards-parallel` | Registry/cache negative tests; retired alias rejection; public pin remote-cache digest check. |
| **Slice 7** | Source deletion & docs | `make test-fast-parallel` | Final source cleanup; docs wording lint and goldens; post-deletion stale-path scan. |

---

## Required Changes for Finalization

1. **Assign Core Role-Asset Removal (Blocker):** Assign the in-tree Core role-asset removal (`mol-polecat-*` deletion, `gc-dispatch` split, and `mol-do-work` de-referencing) to a named slice. Sequence it so the referenced formulas are re-homed in public Gastown *before or in the same commit* that removes them from Core.
2. **Reconcile Slicing Representations (Major):** Add a canonical crosswalk mapping prose slices 1–7 to the Attempt-14 A–F commit steps and the Attempt-9 matrix. State which representation is authoritative and why.
3. **Verify Manifest Completeness Independently (Major):** Add an independent, VCS-level completeness check that reconciles behavior manifest rows against a VCS-level list of moved/deleted/added pack assets (e.g. `git diff --name-status` over pack roots).
4. **Resolve Examples Green-State Coexistence (Minor):** Specify how `go test ./examples/...` stays green during slices 2–4—e.g., delete or skip the local-pack behavior tests when the example city is rewired, or delete/rewire the stale local packs in the same slice.
