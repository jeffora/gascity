# Ritu Raman — DeepSeek V4 Flash Perspective Independent Review (Iteration 14 / Attempt 14)

**Verdict:** approve-with-risks

**Lane:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden dependency discovery.

This independent review evaluates the Iteration 14/Attempt 14 revision of `design.md` against `requirements.md` and the existing codebase behavior. It evaluates the progress made since earlier iterations, analyzes how the updated design addresses the core invariants of this lane, and highlights outstanding risks that must be closed prior to implementation.

---

## Executive Summary

The Iteration 14 design represents a mature and highly structured approach to the Core and Gastown pack split. By defining a clear **Bootstrap Extraction Completion Contract** (§503–540) and explicitly transitioning `internal/bootstrap` to a permanently empty state, the design successfully addresses the major architectural challenges of bootstrap embed cleanup and test fixture isolation.

Crucially, the decision to implement production `bootstrapAssets` as a private, non-nil, custom empty `fs.FS` returning `fs.ErrNotExist` (§506–509) represents a textbook application of Zero Framework Cognition (ZFC) and the Bitter Lesson. It prevents linker flag pollution from standard testing initialization while eliminating the risk of runtime nil-FS panics. Furthermore, the retirement of `GC_BOOTSTRAP=skip` as a production switch closes the low-level escape hatch that previously allowed unvalidated bypasses.

However, from a strict **Bootstrap Fixture Isolation and Hidden Dependency Discovery** perspective, several subtle compilation, runtime, and diagnostic gaps remain:
1. **The `GC_BOOTSTRAP` Env-Mutation Zombie Dance (§533–539):** The doctor function `ensureBootstrapForDoctor` in `internal/doctor/implicit_import_cache_check.go` still saves, unsets, and restores `GC_BOOTSTRAP` to trigger implicit import caching. While the design retires the production skip behavior, this doctor helper is unmentioned in the transition plan. Once `EnsureBootstrap` becomes a no-op apart from retired pruning, this environment-variable mutation dance is dead scaffolding and must be cleaned up.
2. **Vestigial Embed Compilation Failures (§519–522):** The design mandates removing the `//go:embed packs/**` directive from `internal/bootstrap/bootstrap.go`. However, leaving any vestigial references to the `embeddedBootstrapPacks` variable or empty directories without files will cause immediate Go compilation failures (`pattern packs/**: no matching files found`). This must be executed with atomic deletion of both the directive and its associated variable.
3. **Mock/Fixture Resilience under Symbolic Bindings (§1700–1720):** Core's new symbolic worker routing via `[gc.bindings.maintenance_worker]` introduces configuration dependencies. Because low-level config/loader tests use minimal inline `fstest.MapFS` fixtures that are completely stripped of these tables, the configuration resolution layer must be verified to handle the absence of `gc.bindings` gracefully without panicking or triggering fatal errors.
4. **Severe Command Latency from Manifest Integrity Verification (§1799):** The fatal gates mandate recursively checking required-pack file-sets against exact manifests and digests. Performing cryptographic hashing of dozens of templates and prompts on *every single CLI command* (e.g., `gc sling`, `gc prompt`) is a significant disk I/O bottleneck that will degrade CLI response times. High-performance caching or modification checking must be used instead.

---

## Technical Evaluation of Invariant Questions

### Q1. Does `internal/bootstrap` stop embedding production Core while keeping bootstrap tests deterministic through explicit isolated fixtures?
* **Yes:** Production `bootstrapAssets` is specified as a private, custom empty `fs.FS` returning `fs.ErrNotExist` for all paths. This completely extracts production Core assets from `internal/bootstrap`, preventing standard testing package pollution (e.g., `-test.v` flags) from affecting command execution.
* **Deterministic Fixtures:** Bootstrap tests are strictly confined to injecting explicit, inline `fstest.MapFS` fixtures (§511–515). This guarantees absolute isolation from the production filesystem and ensures that tests remain fully deterministic under all running environments.

### Q2. How is fixture drift against the shipped Core pack detected without causing low-level config tests to exercise production assets accidentally?
* **Decoupled Verification by Construction:** Because the inline mock fixtures are forbidden from mimicking or copying real Core directory structures, there is no physical file structure to drift. Content and behavioral fidelity validation is shifted entirely to the `internal/systempacks` integration suite against the real embed.
* **Accidental Leakage Gate:** Low-level config/loader tests must execute within isolated sandboxed directories (e.g., setting `GC_HOME = t.TempDir()`). The design enforces this via the path-string guard and `TestBootstrapFixtureIsMinimal` (§529–531), which fails if any mock fixture contains production-only directories such as `formulas/`, `orders/`, `overlay/`, `skills/`, or `assets/prompts/`. This blocks developers from accidentally "creeping" real Core assets into mock fixtures.

### Q3. Are tests needing no-Core behavior using structurally test-only lower-level loaders rather than runtime flags or environment switches?
* **Yes:** The legacy environment variable switch `GC_BOOTSTRAP=skip` is completely retired as a production behavior switch (§533–536). Low-level config tests that require custom or no-Core configurations must invoke the lower-level loaders directly using custom `fstest.MapFS` mock FS fixtures or `_test.go` helpers. No production CLI command can bypass required Core materialization, retired-source classification, or fatal participation validation.

---

## Top Strengths of the Iteration 14 Design

- **Go Linker-Flag Containment (§506–509):** The transition of production `bootstrapAssets` to a custom, empty `fs.FS` returning `fs.ErrNotExist` is a robust, elegant solution that ensures `internal/bootstrap` has a non-nil filesystem while carrying zero bytes of production Core.
- **Strict Mock Invariant Assertion (§529–531):** The addition of `TestBootstrapFixtureIsMinimal` establishes a strong automated boundary. Failing tests if inline fixtures contain directories like `formulas/` or `skills/` prevents developers from reintroducing heavy, drift-prone mock setups.
- **Explicit Hidden Dependency Inventory (§523–527):** Listing `cmd/gc/prompt_test.go`, `internal/config/bundled_import_test.go`, `examples/gastown/precompact_hook_test.go`, and `test/packlint/*` in the transition plan ensures that the path-string migration is explicit, rather than being discovered through ad-hoc compilation failures.
- **Unified Retired-Source Classifier (§1548):** Centralizing retired path detection ensures that diagnostics, doctor cleanups, and loader exclusions always share a single source of truth.

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
* **The Gap:** The design retires `GC_BOOTSTRAP=skip` as a production behavior switch (§533). When `bootstrap.EnsureBootstrap` becomes an empty operation (apart from pruning retired imports), this saving, unsetting, and restoring of the environment variable in `ensureBootstrapForDoctor` is completely redundant dead code. Leaving it untouched creates cognitive debt for future maintainers.
* **Recommendation:** Explicitly slate the `ensureBootstrapForDoctor` helper for deletion or refactoring in the Core extraction slice. All doctor-level cached import validations should be routed directly through the central `doctor.MutationCoordinator` and `internal/packsource`.

### 2. Compilation Failure of Empty Go Embeds
* **The Code Evidence:** `internal/bootstrap/bootstrap.go:22-23` currently defines:
  ```go
  //go:embed packs/**
  var embeddedBootstrapPacks embed.FS
  ```
* **The Gap:** The design states: "After removing `//go:embed packs/**`, `bootstrapAssets` defaults to..." (§507–508) and "remove `internal/bootstrap/packs/core/embed.go`" (§519). In Go, if a `//go:embed` directive points to a pattern that matches no files (such as an empty `packs/` directory after files are moved), compilation fails immediately with `pattern packs/**: no matching files found`.
* **Recommendation:** The design must explicitly mandate **completely deleting** the `//go:embed packs/**` comment and the `embeddedBootstrapPacks` variable from `internal/bootstrap/bootstrap.go` in the same commit that moves the Core assets, preventing intermediate compilation breakages in the CI pipeline.

### 3. Mock/Fixture Resilience with New Worker Bindings
* **The Code Evidence:** The design introduces symbolic worker routing via `[gc.bindings.maintenance_worker]` in Core `pack.toml`.
* **The Gap:** Low-level config tests and bootstrap loader tests use inline `fstest.MapFS` fixtures that are mandated by §511–515 to contain *only* `pack.toml` and a single agent table when required. They will completely lack `gc.bindings` tables or `maintenance_worker` mappings. If the loader or config resolution paths assume these bindings are always present (or fail with a strict validation diagnostic), then all low-level config mock tests will break.
* **Recommendation:** Require that the config resolution layer gracefully handles the complete absence of `gc.bindings` (and specifically `maintenance_worker`) during "test-only fixture" loads, ensuring they do not trigger fatal diagnostic errors or panic.

### 4. Severe Command Latency from Hashing Entire File-Sets
* **The Gap:** Section §1799 mandates checking required packs "against exact manifests and `pack.toml` digests" before any config, formula, or script is read. Cryptographically hashing dozens of files (including prompts, templates, overlays, and skills) on *every single CLI invocation* (such as `gc sling` or `gc prompt`) introduces a significant disk I/O bottleneck that will severely degrade CLI responsiveness, violating Bitter-Lesson and DX invariants.
* **Recommendation:** Refine the fatal integrity gates to check a single compiled metadata/manifest hash file (e.g. `manifest.json.sha256`) or use lightweight file-modification checking (mod time/size) during standard runtime commands, reserving full-file-set hashing for `gc doctor` and materialization/self-heal operations.

### 5. `cmd/gc/prompt_test.go` Direct Disk Reads of Core Prompts
* **The Code Evidence:** `cmd/gc/prompt_test.go:781–782` reads two Core prompt files directly from disk via relative paths:
  ```go
  "internal/bootstrap/packs/core/assets/prompts/pool-worker.md",
  "internal/bootstrap/packs/core/assets/prompts/graph-worker.md",
  ```
* **The Gap:** When Core moves to `internal/packs/core`, these direct `os.ReadFile` calls will fail with `ENOENT`. While the design correctly lists `prompt_test.go` in the hidden-dependency inventory to be updated (§523), it doesn't specify *how* it should be updated.
* **Recommendation:** Require that `cmd/gc/prompt_test.go` is refactored to read these prompt assets through the embedded `core.PackFS` filesystem rather than hardcoded on-disk path strings. Reading from the embedded filesystem is far more robust because it validates the actual bytes packaged into the binary rather than the developer's local source tree.

---

## Required Changes for Finalization

1. **Delete Doctor Env-Mutation Helper:** Add `internal/doctor/implicit_import_cache_check.go` (`ensureBootstrapForDoctor`) to the Core extraction slice transition plan. Delete the environment-saving dance and route implicit-import cache checks through the central `doctor.MutationCoordinator`.
2. **Mandate Complete Deletion of Embed Directives:** Update §507 and §519 to require the absolute deletion of the `//go:embed packs/**` comment and `embeddedBootstrapPacks` variable from `internal/bootstrap/bootstrap.go` to prevent compiler errors.
3. **Require Mock-Resiliency for Bindings:** Add an explicit test invariant: "Config loader tests running with test-only inline fixtures must resolve successfully even when `gc.bindings` is completely absent from the mock metadata."
4. **Performance Gate for Fatal Integrity Checks:** Clarify §1799 to state that pre-resolution integrity validation must utilize a high-performance verification mechanism (such as single-manifest digest comparison) rather than recursively reading and hashing every individual prompt and asset file on every CLI run.
5. **Refactor `prompt_test.go` to use `PackFS`:** Explicitly mandate that the refactoring of `cmd/gc/prompt_test.go` uses embedded `core.PackFS` instead of direct `os.ReadFile` of disk paths, ensuring robust test-to-binary fidelity.

---

## Reflective Questions

- **Why maintain `internal/bootstrap`?** Since `internal/bootstrap` after the Core extraction will be a permanently vestigial package containing empty materialization logic, an erroring empty `fs.FS`, and a handful of retired prunings, why do we maintain it? Should its remaining prunings be moved into `internal/systempacks` or `internal/packsource`, and the `internal/bootstrap` package be completely deleted?
- **Air-Gapped Test Isolation:** In air-gapped CI environments where `test/packcompat` cannot reach GitHub to fetch the public Gastown repository, how do we guarantee deterministic test runs? Is there a local caching/fallback mechanism to vendor the companion public Gastown commit into the test tree?
