# Ritu Raman — Gemini Perspective Independent Review (Iteration 13 / Attempt 13)

**Verdict:** approve-with-risks

**Lane:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden dependency discovery.

Reviewed against the Iteration 13 / Attempt 13 draft of `design.md` (`updated_at: 2026-06-07T14:05:04Z`) and `requirements.md` in the active repository workspace.

---

## Executive Summary

The Iteration 13 design demonstrates exceptional structural discipline by implementing the **Required Core Loading Fatal Gates (§1789–1822)** and transitioning `internal/systempacks` to a strict deny-by-default architecture. These additions successfully resolve many of the systemic, loader-level bypass risks that historically allowed cities to silently run without critical required configurations.

Transitioning production `bootstrapAssets` to a custom, non-nil empty `fs.FS` returning `fs.ErrNotExist` (§1358–1359) and enforcing inline `fstest.MapFS` mock assets via `TestBootstrapFixtureIsMinimal` (§1360–1363) are outstanding applications of Zero Framework Cognition (ZFC) and the Bitter Lesson. They prevent testing-package initialization flags (like `-test.v`) from polluting production `gc` runs and eliminate structural fixture drift.

However, a rigorous audit of the **Bootstrap Fixture Isolation and Hidden Dependency Discovery** boundary reveals several subtle technical risks, compilation vulnerabilities, and configuration gaps:

1. **The `GC_BOOTSTRAP` Env-Mutation Zombie Dance (§1350, §1631):** The production doctor utility `ensureBootstrapForDoctor` in `internal/doctor/implicit_import_cache_check.go` actively saves, unsets, and restores `GC_BOOTSTRAP` around a direct `bootstrap.EnsureBootstrap` call. This helper is unlisted in the transition plans and will become a dead environment-mutation wrapper around a no-op once `BootstrapPacks` is empty.
2. **Empty Go Embed Compilation Vulnerability (§469):** Removing the `packs/` directory contents while leaving any `//go:embed packs/**` patterns or references to `embeddedBootstrapPacks` in `internal/bootstrap/bootstrap.go` will cause immediate, fatal Go compiler errors. Go forbids embedding directives that resolve to no matching files.
3. **Mock/Fixture Resilience with Symbolic Worker Bindings (§1712–1719):** The newly added `[gc.bindings.maintenance_worker]` table in Core `pack.toml` creates a schema requirement. Loader/config tests running in "test-only fixture" mode (using minimal inline `fstest.MapFS` assets) must be proven to not fail or panic when resolving configurations that are completely stripped of these bindings.
4. **Pre-Resolution Hashing Performance Bottleneck (§1799):** Recursively walking and hashing entire required-pack file-sets (prompts, scripts, overlays, skills) on *every single* normal production command invocation (such as `gc sling`, `gc prompt`) introduces a major I/O performance risk that must be mitigated by a compiled manifest hash or file-modification caching.
5. **Fixture Contract Path Contradiction (§461–465, §494–497):** There is a logical contradiction between allowing a "tiny compatibility embed under `internal/bootstrap/testdata/packs/core`" and the strict path-string guard that "rejects any lingering import, `AssetDir`, or path constant referencing `internal/bootstrap/packs/core`."

---

## Technical Evaluation of Invariant Questions

### Q1. Does `internal/bootstrap` stop embedding production Core while keeping bootstrap tests deterministic through explicit isolated fixtures?
* **Yes:** Production `bootstrapAssets` is specified as a private, empty `fs.FS` returning `fs.ErrNotExist`. This cleanly extracts production Core assets from the `internal/bootstrap` package and prevents standard `testing` package initialization pollution.
* **Deterministic Fixtures:** Bootstrap tests are restricted to injecting explicit inline `fstest.MapFS` fixtures. This ensures complete isolation from the production filesystem and guarantees absolute test determinism.

### Q2. How is fixture drift against the shipped Core pack detected without causing low-level config tests to exercise production assets accidentally?
* **Decoupled Verification Path:** Low-level config and bootstrap loader tests are restricted to verifying mechanism behaviors (like collision detection, metadata parsing, and retired-source classification) using minimal, inline mock fixtures.
* **Drift Avoidance by Construction:** Because the allowed fixtures are prohibited from copying Core structures (enforced by `TestBootstrapFixtureIsMinimal`'s strict allowlist), there is zero file structure to drift. Content and behavioral fidelity validation is shifted entirely to the `internal/systempacks` integration suite against the real embed.
* **Accidental Production Leakage Gate:** Low-level loader/config tests must configure isolated runtime environments (e.g. setting `GC_HOME = t.TempDir()`). The design enforces this via the path-string guard and `TestBootstrapFixtureIsMinimal`, which prevents production Core from being loaded from the workspace in low-level tests.

### Q3. Are tests needing no-Core behavior using structurally test-only lower-level loaders rather than runtime flags or environment switches?
* **Yes:** `GC_BOOTSTRAP=skip` is completely retired as a production behavior switch (§1350). Low-level config tests that require custom or no-Core setups must invoke the lower-level loaders directly with custom mock fs fixtures, or use `_test.go` helpers. No production CLI command can bypass Core materialization, retired-source classification, or fatal participation validation.

---

## Top Strengths of the Current Design

* **Go Linker-Flag Containment (§1358–1359):** Implementing `bootstrapAssets` as a custom empty `fs.FS` prevents standard `testing` init flags (like `-test.v`) from polluting production `gc` commands, a robust and elegant solution.
* **Strict Deny-by-Default Fatal Gates (§1792–1810):** The requirement to classify all production config loads (runtime, no-refresh, partial doctor, partial edit, test-only) prevents unclassified code paths from accidentally bypassing system pack validation.
* **Unified Retired-Source Classifier (§1548):** Centralizing the identification of legacy local paths (such as old `internal/bootstrap/packs/core` or `examples/gastown`) ensures that diagnostics, doctor cleanups, and loader exclusions always share a single source of truth.
* **Symbolic Overrides for Bindings (§1747–1753):** Abstracting the maintenance agent to `gc.bindings.maintenance_worker` completely eliminates hardcoded agent roles from Go source files, satisfying ZFC principles perfectly.

---

## Critical Risks, Gaps & Hidden Dependencies

### 1. The `GC_BOOTSTRAP` Env-Mutation Zombie Dance in `ensureBootstrapForDoctor`
* **The Code Evidence:** `internal/doctor/implicit_import_cache_check.go:235-249` defines:
  ```go
  func ensureBootstrapForDoctor(gcHome string) error {
      prev, hadPrev := os.LookupEnv("GC_BOOTSTRAP")
      if err := os.Unsetenv("GC_BOOTSTRAP"); err != nil {
          return err
      }
      defer func() {
          if hadPrev {
              _ = os.Setenv("GC_BOOTSTRAP", prev)
              return
          }
          _ = os.Unsetenv("GC_BOOTSTRAP")
      }()
      return bootstrap.EnsureBootstrap(gcHome)
  }
  ```
* **The Gap:** The design mandates deleting "any remaining branch that changes production config behavior based on `GC_BOOTSTRAP=skip`" (§1355–1356). However, this production doctor function is unlisted in the design’s transition and rollout plans. When `bootstrap.EnsureBootstrap` becomes an empty operation (apart from pruning retired imports), this saving/unsetting/restoring environment dance is dead scaffolding. Furthermore, if `EnsureBootstrap`'s skip branch is removed, the environment variable no longer affects it, making this mutation completely redundant.
* **Recommendation:** Explicitly slate `ensureBootstrapForDoctor` for deletion in the Core extraction slice. All doctor-level implicit import prunings should route through the `doctor.MutationCoordinator` and the `internal/packsource` classifier.

### 2. Compilation Failure of Empty Go Embeds
* **The Code Evidence:** `internal/bootstrap/bootstrap.go` currently defines:
  ```go
  //go:embed packs/**
  var embeddedBootstrapPacks embed.FS
  ```
* **The Gap:** The design states: "After removing `//go:embed packs/**`, `bootstrapAssets` defaults to..." (§456–457) and "remove `internal/bootstrap/packs/core/embed.go`" (§469). In Go, if a `//go:embed` directive points to a pattern that matches no files (such as an empty `packs/` directory), compilation will fail immediately with `pattern packs/**: no matching files found`.
* **Recommendation:** The design must explicitly mandate **completely deleting** the `//go:embed packs/**` comment and `embeddedBootstrapPacks` variable from `internal/bootstrap/bootstrap.go` in the Core extraction slice, rather than leaving them in a vestigial state.

### 3. Mock/Fixture Resilience with New Worker Bindings
* **The Code Evidence:** The Attempt 11 contract introduces symbolic worker routing via `[gc.bindings.maintenance_worker]` in Core `pack.toml` (§1712–1719).
* **The Gap:** Low-level config tests and bootstrap loader tests use inline `fstest.MapFS` fixtures that are mandated by §1360 to contain *only* `pack.toml` and a single agent table when required. They will completely lack `gc.bindings.maintenance_worker` mappings. If the loader or config resolution paths assume these bindings are always present (or fail with a strict `core.maintenance_worker.missing` or similar diagnostic during general resolution), then all low-level loader mock tests will break.
* **Recommendation:** Explicitly require that the config resolution layer gracefully handles the complete absence of `gc.bindings` (and specifically `maintenance_worker`) during "test-only fixture" loads, ensuring they do not trigger fatal diagnostic errors or panic.

### 4. Severe Command Latency from Hashing Entire File-Sets
* **The Gap:** Section §1799 mandates checking required packs "against exact manifests and `pack.toml` digests" before any config, formula, or script is read. Hashing dozens of files (including prompts, templates, overlays, and skills) on *every single CLI invocation* (e.g., `gc sling`, `gc prompt`) is a significant disk I/O bottleneck that will severely degrade CLI responsiveness, violating Bitter-Lesson and DX invariants.
* **Recommendation:** Refine the fatal gates to check a single compiled metadata/manifest hash file (e.g. `manifest.json.sha256`) or use lightweight file-modification checking (mod time/size) during standard runtime commands, reserving full-file-set hashing for `gc doctor` and materialization/self-heal operations.

### 5. Contradictory Fixture Path Contract
* **The Gap:** The design contains a contradiction between allowing a "tiny compatibility embed under `internal/bootstrap/testdata/packs/core`" and the strict path-string guard that "rejects any lingering import, `AssetDir`, or path constant referencing `internal/bootstrap/packs/core`." If the guard rejects `AssetDir: "packs/core"`, the compatibility embed under `testdata/packs/core` cannot use that path value.
* **Recommendation:** Completely delete the "tiny compatibility embed under `internal/bootstrap/testdata/packs/core`" clause. Explicitly mandate that test fixtures must use a synthetic path name like `packs/test-core` or `packs/fixture-core`, ensuring the path-string guard allows `packs/test-core` but remains absolute for `packs/core`.

### 6. `cmd/gc/prompt_test.go` Direct Disk-Read Breakage
* **The Code Evidence:** `cmd/gc/prompt_test.go` reads production Core prompts from disk:
  ```go
  "internal/bootstrap/packs/core/assets/prompts/pool-worker.md",
  "internal/bootstrap/packs/core/assets/prompts/graph-worker.md",
  ```
* **The Gap:** This is a `cmd/gc` test that references the production Core path by string literal. Once Core moves to `internal/packs/core`, this test will immediately fail with `os.ReadFile` returning `ENOENT`. It is omitted from the design's hidden-dependency inventory.
* **Recommendation:** Add `cmd/gc/prompt_test.go` to the hidden-dependency inventory in the design. Specify that this test must be updated to read from `core.PackFS` or the new `internal/packs/core/` path.

---

## Required Changes for Finalization

1. **Delete Doctor Env-Mutation Helper:** Add `internal/doctor/implicit_import_cache_check.go` (`ensureBootstrapForDoctor`) to the Core extraction slice transition plan. Delete the environment-saving dance and route implicit-import cache checks through the central `doctor.MutationCoordinator`.
2. **Mandate Complete Deletion of Embed Directives:** Update §456 and §469 to require the absolute deletion of the `//go:embed packs/**` comment and `embeddedBootstrapPacks` variable from `internal/bootstrap/bootstrap.go` to prevent compiler errors.
3. **Require Mock-Resiliency for Bindings:** Add an explicit test invariant: "Config loader tests running with test-only inline fixtures must resolve successfully even when `gc.bindings` is completely absent from the mock metadata."
4. **Performance Gate for Fatal Integrity Checks:** Clarify §1799 to state that pre-resolution integrity validation must utilize a high-performance verification mechanism (such as single-manifest digest comparison) rather than recursively reading and hashing every individual prompt and asset file on every CLI run.
5. **Resolve Fixture Contract Contradiction:** Remove references to "tiny compatibility embed under `internal/bootstrap/testdata/packs/core`." Specify that all test fixtures use a synthetic, non-overlapping `AssetDir` (such as `packs/test-core` or `packs/fixture-core`), and update the path-string guard to reject `packs/core` absolutely.
6. **Add `cmd/gc/prompt_test.go` to Inventory:** Add `cmd/gc/prompt_test.go` to the hidden-dependency inventory and update it to resolve prompt assets from `core.PackFS` or the new path.

---

## Questions

* Since `internal/bootstrap` after the Core extraction will be a permanently vestigial package containing empty materialization logic, an erroring `fs.FS`, and a handful of retired prunings, why do we maintain it? Should its remaining prunings be moved into `internal/systempacks` or `internal/packsource`, and the `internal/bootstrap` package be completely deleted?
* For the retained `core`-named fixture used in bootstrap tests: exactly which legacy-identity assertion does it preserve, and against which component's collision gate is it asserted in the Slice 3 to Slice 4 window?
