# Ritu Raman — DeepSeek V4 Flash Perspective Independent Review (Iteration 20 / Attempt 20)

**Verdict:** approve

**Scope:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden dependency discovery.

This independent review evaluates the Iteration 20 / Attempt 20 draft of the Core and Gastown Pack Split design (`.gc/design-reviews/ga-2404qu/attempt-20/design-before.md`) against the `requirements.md` and the active codebase at the `rig_root` (`/data/projects/gascity`).

---

## Executive Summary

As Ritu Raman, the **Bootstrap Fixture Isolation Reviewer**, I am issuing a **Verdict: Approve** for the Iteration 20 draft.

The current design represents an exceptionally clean, robust, and highly disciplined blueprint for decoupling the production binary from legacy global implicit imports. Transitioning to a **source-symbol guarded scanner** and defining `bootstrapAssets` as a private, non-nil empty filesystem returning `fs.ErrNotExist` (§506–509) represents excellent structural hygiene.

By systematically addressing the technical feedback and edge cases raised in prior iterations, the Iteration 20 document successfully eliminates all critical blocking gaps:
1. **Unspecified `bootstrapAssets` Default (Resolved §506–509):** The design explicitly mandates that `bootstrapAssets` must default to a private, non-nil, empty filesystem that returns `fs.ErrNotExist`, avoiding any latent nil-FS panics or opaque errors on standard loader paths.
2. **`GC_BOOTSTRAP=skip` Production Bypass (Resolved §533–539):** The environment variable is fully retired as a production switch, ensuring it cannot be abused as an escape hatch, while narrowing its test-only scope strictly to legacy fixture materialization.
3. **Contradictory Fixture Embed Paths (Resolved §511–515):** The contradictory "tiny compatibility embed" clause has been cleanly replaced by a strict inline `fstest.MapFS` requirement, preventing mutable production folders from leaking into low-level tests.
4. **Complete Hidden-Dependency Inventory (Resolved §523–527):** Critical single-point-of-change dependencies—including `cmd/gc/prompt_test.go`, `internal/config/bundled_import_test.go`, and hook overlays—have been formally added to the refactoring checklist for Slice 3.

With these rigorous technical guards in place, the bootstrap extraction and test-isolation model is fully mature.

---

## Top Strengths of Current Design

* **Absolute Production Binary Containment (§506–509):** Deleting the `//go:embed packs/**` directive and the `embeddedBootstrapPacks` variable ensures that no Core asset leak can occur in production binaries.
* **Hermetic, Non-Nil Private fs.FS (§506–509):** Forcing `bootstrapAssets` to default to a custom, non-nil filesystem returning `fs.ErrNotExist` prevents standard loader code from throwing nil-pointer panics while cleanly reporting absent fixtures.
* **Decoupled Verification via Inline `fstest.MapFS` (§511–515):** Requiring bootstrap tests to use inline synthetic fixtures ensures config and bootstrap parsing remains fully testable without copying mutable production directories, completely eliminating disk-drift.
* **Negative Asset-Presence Guard (§529–531):** Implementing `TestBootstrapFixtureIsMinimal` (failing if fixture assets contain production-only folders like `formulas/`, `orders/`, or `skills/`) guarantees that synthetic test fixtures do not secretly grow to include real production behaviors.
* **Explicit Downstream Migration Paths (§523–527):** Formally documenting that `cmd/gc/prompt_test.go`, `internal/config/bundled_import_test.go`, and the Hook overlays README must be migrated in the same slice prevents intermediate compile breaks.

---

## Nuanced Risks & Operational Recommendations

While the design is approved, the following highly focused technical recommendations are offered to ensure a smooth, regression-free implementation:

### 1. Centralize Mock Fixture Setup to Prevent Inline Duplication and Drift
* **The Risk:** Multiple distinct test files (including `bundled_import_test.go`, `prompt_test.go`, and collision tests) are mandated to use inline `fstest.MapFS` configurations. If these mock literals are defined separately in each file, any future changes to the required structure of `pack.toml` or schema parsing will require manual edits across numerous test files, creating maintenance friction and a vector for test-only schema drift.
* **Recommendation:** Create a centralized test helper (e.g., `testhelper.NewMinimalCoreFixture()`) inside a non-production-exported testing utility. This helper should build and return the canonical minimal `fstest.MapFS` required for config-parsing tests. All unit tests requiring a synthetic Core fixture should query this utility.

### 2. Verify Hook-Overlay Hook Isolation in Slice 3
* **The Risk:** The hooks package (`internal/hooks/hooks.go`) and associated tests interact with provider-specific overlays. When removing `internal/bootstrap/packs/core` in Slice 3, there is a risk that unstated asset dependencies or path constants inside the hook-loading subsystem could silently break without direct compile errors.
* **Recommendation:** Mandate that Slice 3 contains a specific test verification step for the `internal/hooks` suite to ensure provider overlays load cleanly from the new `internal/packs/core` path without assuming legacy paths.

---

## Technical Evaluation of Invariant Questions

### Q1. Does `internal/bootstrap` stop embedding production Core while keeping bootstrap tests deterministic through explicit isolated fixtures?
* **Ritu Raman Finding:** **Yes.** Deleting the `//go:embed packs/**` directive and reassigning `bootstrapAssets` to a custom private filesystem returning `fs.ErrNotExist` successfully purges production Core from the bootstrap package. Tests achieve full determinism by using inline synthetic `fstest.MapFS` structures that mimic the parsing requirements without relying on real disk assets.

### Q2. How is fixture drift against the shipped Core pack detected without causing low-level config tests to exercise production assets accidentally?
* **Ritu Raman Finding:** **Satisfactory.** Low-level tests are completely isolated using inline `fstest.MapFS` structures, preventing them from touching real Core. Fixture minimal-size verification is continuously enforced by `TestBootstrapFixtureIsMinimal` (§529–531), which fails if any production-only assets leak into the test fixtures. Any broader schema or behavioral drift is validated upstream at the integration/compilation layer via the `test/packcompat` suite without compromising low-level unit test hermeticity.

### Q3. Are tests needing no-Core behavior using structurally test-only lower-level loaders rather than runtime flags or environment switches?
* **Ritu Raman Finding:** **Yes.** The complete retirement of `GC_BOOTSTRAP=skip` as a production bypass (§533–539) ensures that the normal runtime loader pathways cannot be skipped via environment state. Any test requiring explicit "no-Core" behavior must load the configuration using low-level package-internal loaders directly, enforcing a strict structural separation.

---

## Recommendations for Finalization

1. **Implement `testhelper.NewMinimalCoreFixture`:** In the implementation plan, explicitly slate the creation of a centralized test fixture builder to prevent duplicate inline `fstest.MapFS` definitions.
2. **Explicit Hook Path Verification:** Ensure that the AST and import scanner validation covers `internal/hooks` in the Slice 3 gate checks.
