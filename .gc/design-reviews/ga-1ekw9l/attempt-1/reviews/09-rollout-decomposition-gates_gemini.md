# Iris Kowalski - Gemini

**Verdict:** block

I have reviewed `plans/core-gastown-pack-migration/implementation-plan.md` against `requirements.md` and the `gc.mayor.implementation-plan.v1` schema, focusing specifically on my lane: independently deployable slices, rollout timelines, test-suite integrity, exact gates, and cross-document/cross-repo sequencing. 

While the 7-slice rollout is exceptionally thoughtful about avoiding unverified mega-commits and has robust validation matrices, the timeline contains a dangerous collision-detection blind spot in Slice 3, a sequencing contradiction with `dolt-target.sh` in Slice 2, and a high-probability breakage vector for the fast unit test suite under the proposed `GC_BOOTSTRAP` semantics. Consequently, I must block.

---

## Top Strengths

- **Exceptional Slicing Hygiene:** Splitting the migration into 7 distinct, sequentially validated slices directly addresses the anti-batching requirements. The distinction between public compatibility pin (Slice 2) and activation pin (Slice 5) is robust.
- **Rigorously Gated Transitions:** Prohibiting a test-only loader bypass for the no-Maintenance gate is a phenomenal safety invariant. Requiring candidate builds to validate real production loading paths before merge prevents silent regressions.
- **Clear Rollback Boundaries:** The release compatibility matrix explicitly covers old/new binary × old/new pack skew, including explicit rollback boundaries and one-way upgrade limits tied directly to release notes.

---

## Critical Risks

### 1. [Blocker] The Slice 3 Shadowing Blind-Spot (Timeline Discrepancy)
- **The Gap:** The `Bootstrap Cleanup` section (lines 2523-2526) states: *"Remove `core` and `registry` from `bootstrapManagedImportNames` in `internal/config/compose.go` in the same slice that permanently empties `BootstrapPacks` [Slice 3, Core Extraction]"*.
- **The Conflict:** However, the strict pre-resolution file-set integrity, typed `RequiredSystemPackParticipation`, and collision gates are not introduced until Slice 4 (*"Core loading/doctor slice"*, lines 2765-2773).
- **The Risk:** If you remove `core` and `registry` from `bootstrapManagedImportNames` in Slice 3, but the new loader validation and collision checks are not yet active until Slice 4, you create a temporary shadowing blind spot. During Slice 3, any user pack or import could shadow `core` or `registry` undetected. To maintain continuous safety invariants, these checks cannot be decoupled; the removal from `bootstrapManagedImportNames` must happen *after* or *simultaneously* with the activation of the Slice 4 collision/participation gates.

### 2. [Blocker] The `GC_BOOTSTRAP=skip` Test-Suite Breakage Risk
- **The Gap:** `cmd/gc/main_test.go` sets `GC_BOOTSTRAP=skip` as a suite-wide default (lines 45, 62) to keep minimal, fake-backend, and offline tests running fast and cleanly without full environment materialization.
- **The Conflict:** Under the `Bootstrap Cleanup` section (lines 2540-2544), the plan states: *"Retire `GC_BOOTSTRAP=skip` as a production behavior switch. If retained for tests, it may skip only empty bootstrap fixture materialization; it must not skip `internal/systempacks` materialization, required Core file-set integrity, retired-source classification, collision checks, or typed participation validation."*
- **The Risk:** If `GC_BOOTSTRAP=skip` is narrowed so that it no longer skips `internal/systempacks` materialization and validation, then **every single test** in `main_test.go` and testscript suites will attempt full required Core materialization and participation checks. Since many fast unit tests run in highly isolated, empty directories or fake environments, they do not have a valid Core pack or a global cache. Forcing them to execute these production gates will cause mass failures across hundreds of previously green unit tests, breaking the fast unit baseline.

### 3. [Major] The `dolt-target.sh` Rollout Sequencing Mismatch
- **The Gap:** Slice 2 states that it will *"update `examples/dolt` formulas and references that depend on legacy Maintenance scripts or role targets"* (lines 2744-2747).
- **The Conflict:** `examples/dolt/assets/scripts/port_resolve.sh` currently hardcodes a path pointing to `packs/maintenance/.../dolt-target.sh`. However, the required `dolt-target.sh` asset is not extracted/moved into Core (`internal/packs/core`) until Slice 3 (Core Extraction), or permanently folded until Slice 5.
- **The Risk:** During Slice 2, the `internal/packs/core` package does not exist yet. If Slice 2 attempts to rewire `examples/dolt` to point to Core's version of `dolt-target.sh`, those references will break immediately. The plan must clarify what `examples/dolt` resolves `dolt-target.sh` from during the Slice 2 gap, or sequence the extraction of `dolt-target.sh` into an earlier prerequisite slice.

### 4. [Minor] Vestigial `ensureBootstrapForDoctor` Hidden Dependency Leak
- **The Gap:** `internal/doctor/implicit_import_cache_check.go` contains `ensureBootstrapForDoctor` (lines 235-249), which actively unsets, saves, and restores `GC_BOOTSTRAP` around a direct `EnsureBootstrap` call.
- **The Conflict:** Once `BootstrapPacks` is permanently empty and bootstrap assets are a non-nil empty filesystem in Slice 3, calling `EnsureBootstrap` becomes a complete no-op in production.
- **The Risk:** Leaving this unhandled results in a dead-code mutation and active environmental manipulation for a no-op function. The plan should explicitly schedule the deprecation and removal of `ensureBootstrapForDoctor` or its adaptation to the new `internal/systempacks` infrastructure.

---

## Missing Evidence

- **Proof of Fast Unit Test Preservation:** There is no evidence or proof-of-concept showing how minimal offline/fake-backend tests (which rely on `GC_BOOTSTRAP=skip`) will remain green and performant without full Core materialization under the narrowed semantics.
- **Concrete Transition Plan for `ensureBootstrapForDoctor`:** No explicit timeline or transition steps are provided to clean up or adapt the doctor's bootstrap unsetting/restoring logic once bootstrap is permanently empty.

---

## Required Changes

1. **Synchronize Collision Gates:** Move the removal of `core` and `registry` from `bootstrapManagedImportNames` from Slice 3 to Slice 4, aligning it directly with the introduction of `internal/systempacks` strict pre-resolution file-set integrity and typed participation gates.
2. **Preserve `GC_BOOTSTRAP=skip` Test Isolation:** Allow `GC_BOOTSTRAP=skip` to continue skipping full system pack materialization and typed participation validation specifically for test suites/testscript environments, or explicitly provide a lightweight, pre-populated memory/mapFS mock for `internal/systempacks` to prevent suite-wide test breakage.
3. **Correct `dolt-target.sh` Sequencing:** Defer rewiring `examples/dolt`'s `dolt-target.sh` references until Slice 3 (when `internal/packs/core` is actually created and populated), or extract `dolt-target.sh` to a standalone prerequisite location before Slice 2.
4. **Deprecate `ensureBootstrapForDoctor`:** Explicitly add the cleanup/removal of `ensureBootstrapForDoctor` to Slice 4 or Slice 7 as part of the overall bootstrap and doctor refactoring.

---

## Questions

1. How will we ensure that highly-isolated, fake-backend unit tests remain green without having to materialize or fetch Core assets on every test run if `GC_BOOTSTRAP=skip` is retired from skipping system pack loading?
2. Can we extract `dolt-target.sh` to Core in Slice 1 or 2 to allow the `examples/dolt` formula rewiring in Slice 2 to remain valid and independent?
