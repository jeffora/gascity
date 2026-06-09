# Ritu Raman — Bootstrap Fixture Isolation Perspective Independent Review (Iteration 19 / Attempt 19)

**Verdict:** approve-with-risks

**Lane:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden-dependency discovery.

This independent review evaluates the Iteration 19 draft of the Core and Gastown Pack Split design (`.gc/design-review-inputs/core-gastown-pack-migration/design.md` / Attempt 19-retry `design-before.md`) against the `requirements.md` and the live codebase at the `rig_root` (`/data/projects/gascity`).

---

## Executive Summary

As Ritu Raman, the **Bootstrap Fixture Isolation Reviewer**, I am issuing a **Verdict: Approve-With-Risks** for the Iteration 19 draft. 

The current design represents an exceptionally clean blueprint for decoupling the production binary from legacy global implicit imports. Transitioning to a **source-symbol guarded scanner** (§271–276) and defining `bootstrapAssets` as a private, non-nil empty filesystem returning `fs.ErrNotExist` (§506–509) represents excellent structural hygiene.

However, a deep technical analysis of the codebase against the proposed 7-slice rollout sequencing reveals **one critical, unmitigated dependency gap** in the implementation slicing, **one severe operational blindspot** in the transient-error caching layer, and a few lingering vestigial code paths. 

If this design is executed as-is, Slice 3 (Core Extraction) will compile successfully but will silently **strip all Core skills from all running agents**, causing catastrophic production runtime failures that are only detected late in the cycle. This review provides the analytical evidence and the concrete changes required to bridge this gap safely.

---

## Top Strengths of Current Design

* **Absolute Production Binary Containment (§257–260):** Deleting the `//go:embed packs/**` directive and the `embeddedBootstrapPacks` variable ensures that no Core asset leak can occur in production binaries.
* **Hermetic, Non-Nil Private fs.FS (§506–509):** Forcing `bootstrapAssets` to default to a custom, non-nil filesystem returning `fs.ErrNotExist` prevents standard loader code from throwing nil-pointer panics while cleanly reporting absent fixtures.
* **Decoupled Verification via Synthetic Fixtures (§511–515):** Requiring bootstrap tests to use inline `fstest.MapFS` synthetic fixtures (e.g. `packs/test-core`) ensures config/bootstrap parsing remains fully testable without copying mutable production directories.
* **Negative Asset-Presence Guard (§528–531):** Implementing `TestBootstrapFixtureIsMinimal` (failing if fixture assets contain production-only folders like `formulas/`, `orders/`, or `skills/`) guarantees that synthetic test fixtures do not secretly grow to include real production behaviors.

---

## Critical Risks & Gaps (Evidence-Based)

### 1. [MAJOR] The Slice 3 vs. Slice 4 Skill Materialization Dependency Gap
* **The Claim:** The design splits the migration into seven slices (§3. Core extraction slice vs. §4. Core loading/doctor slice). Slice 3 removes Core assets from `internal/bootstrap/packs/core` and empties `BootstrapPacks`. Slice 4 introduces the `internal/systempacks` loader and `config.Load*` scanner.
* **The Reality (Code Grounding):** 
  * [internal/materialize/skills.go:206](file:///data/projects/gascity/internal/materialize/skills.go#L206) loads Core skills via `bootstrapSkillDirs()`.
  * `bootstrapSkillDirs()` ([skills.go:676](file:///data/projects/gascity/internal/materialize/skills.go#L676)) queries `bootstrap.PackNames()` to match implicit imports.
  * In Slice 3, because `BootstrapPacks` is cleared, `bootstrap.PackNames()` returns empty. Consequently, `bootstrapSkillDirs()` returns `nil`, immediately dropping Core skills (`gc-work`, `gc-dispatch`, `gc-mail`, `gc-city`, etc.) from the loader's shared catalog.
  * However, the new required-Core loader / `internal/systempacks` path that restores these skills *only lands in Slice 4*.
* **The Operational Impact:** Any intermediate build on Slice 3 will compile and pass simple file-presence checks, but **will materialize absolutely zero Core skills into agent sinks**. Running agents will be left completely brainless.
* **Required Change:** Mandate that the required-Core systempacks loading fallback or the `internal/systempacks` loading mechanism itself must land **in the same slice** that removes the legacy bootstrap Core paths (combining Slice 3 and Slice 4, or maintaining a temporary bridge in Slice 3).

### 2. [MAJOR] The `skill_catalog_cache.go` Transient Error Blindspot
* **The Claim:** The design asserts that the system remains resilient to transient filesystem errors by relying on the skill catalog cache.
* **The Reality (Code Grounding):**
  * `currentBootstrapCatalogState()` in [cmd/gc/skill_catalog_cache.go:126–164](file:///data/projects/gascity/cmd/gc/skill_catalog_cache.go#L126-L164) is responsible for capturing the live state of implicit imports to decide if transient failures can be safely cached or ignored.
  * To do this, it filters imports against `bootstrap.PackNames()` ([skill_catalog_cache.go:135–138](file:///data/projects/gascity/cmd/gc/skill_catalog_cache.go#L135-L138)).
  * Once `BootstrapPacks` is empty, `bootstrap.PackNames()` is empty. As a result, the transient-error caching layer **completely ignores the new `.gc/system/packs/core` directory**.
  * If a transient filesystem or read-lock error occurs on `.gc/system/packs/core`, `skillCatalogCache` will fail to recognize it as a transient bootstrap-state change and will propagate an empty skill catalog, triggering an immediate and catastrophic config-drift/session-drain storm.
* **Required Change:** Update `currentBootstrapCatalogState()` in `cmd/gc/skill_catalog_cache.go` to explicitly query and track required system packs (including Core) alongside any remaining implicit imports.

### 3. [MINOR] Obsolete Environment Variable Mutation in Doctor Implicit Checks
* **The Claim:** `GC_BOOTSTRAP=skip` is permanently retired as a production behavior switch (§533–539).
* **The Reality (Code Grounding):**
  * [internal/doctor/implicit_import_cache_check.go:235–249](file:///data/projects/gascity/internal/doctor/implicit_import_cache_check.go#L235-L249) defines `ensureBootstrapForDoctor` which unsets, saves, and restores `GC_BOOTSTRAP` around `bootstrap.EnsureBootstrap`.
  * Since `EnsureBootstrap` becomes a no-op (except for pruning retired entries) and `GC_BOOTSTRAP` is no longer a production flag, this environment-variable dance is obsolete, misleading, and introduces maintenance debt.
* **Required Change:** Slate `ensureBootstrapForDoctor` and its env-variable manipulation for complete removal in Slice 3, routing implicit-import check scaffolding through the standard packsource classifiers instead.

---

## Technical Evaluation of Invariant Questions

### Q1. Does `internal/bootstrap` stop embedding production Core while keeping bootstrap tests deterministic through explicit isolated fixtures?
* **Yes:** The removal of `//go:embed packs/**` and the default initialization of `bootstrapAssets` to a private `errFS` returning `fs.ErrNotExist` successfully prevents the production binary from carrying duplicate Core assets. Tests achieve determinism by explicitly injecting inline `fstest.MapFS` fixtures.

### Q2. How is fixture drift against the shipped Core pack detected without causing low-level config tests to exercise production assets accidentally?
* **Decoupled Verification:** Core fidelity verification is moved entirely to systempacks integration tests, while low-level tests verify parsing/loading behavior on synthetic schemas.
* **Accidental Leakage Prevention:** The transition to AST source-symbol scanning (monitoring symbols like `bootstrap.PackNames` and `GC_BOOTSTRAP`) and the negative read-guard assertion (§2628) ensure that low-level config tests do not accidentally read or couple with real production Core.

### Q3. Are tests needing no-Core behavior using structurally test-only lower-level loaders rather than runtime flags or environment switches?
* **Yes:** Production CLI pathways do not support environment switches or runtime flags for skipping Core. Tests requiring no-Core behavior must call lower-level config loader packages directly.

---

## Required Changes Checklist

1. **[Blocker] Combine Slice 3 and Slice 4 (or Implement Skill Fallback):** Mandate that the transition to the new `internal/systempacks` loader must land in the same slice that empties `BootstrapPacks`, preventing a "zero Core skills" intermediate state.
2. **[Blocker] Refactor `skill_catalog_cache.go`:** Update `currentBootstrapCatalogState()` to explicitly query and track the availability of required system packs (including Core) alongside legacy implicit imports.
3. **[Blocker] Clean Up Stale Imports in Production Core-Coupled Files:** Ensure `internal/builtinpacks/registry.go` and `internal/hooks/hooks.go` are explicitly listed in the Slice 3 refactoring checklist to move their legacy `internal/bootstrap/packs/core` imports to the new `internal/packs/core` path.
4. **Retire `ensureBootstrapForDoctor`:** Delete the obsolete environment-variable save/restore and unset dance in `implicit_import_cache_check.go`.
5. **Clean Up Dual-Embed Comments:** Require the deletion of the stale comment in `internal/bootstrap/packs/core/embed.go` that describes the dual-embed design.
