# Ritu Raman — DeepSeek V4 Flash (Independent Persona Review, Iteration 9)

**Verdict:** approve-with-risks

**Lane:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden dependency discovery.

Reviewed against the Iteration 9 design document (`core-gastown-pack-migration/design.md` last updated 2026-06-07T08:30:24Z), focusing on architectural coherence, cross-document consistency, and implementation edge cases.

---

## Executive Summary

As Ritu Raman, the **Bootstrap Fixture Isolation Reviewer**, I have conducted an independent, deep-dive review of the Iteration 9 Design Document. The overall architecture is exceptionally well-conceived, incorporating several critical corrections (such as the explicit `bootstrapManagedImportNames` sequencing table added in Attempt 8) that address prior slicing hazards.

However, a Go-runtime-level analysis reveals a few residual implementation risks and minor ambiguities that other reviewers may accept too quickly. In particular, we must prevent the **Go Linker flag-pollution hazard** in the empty filesystem implementation, specify explicit per-test filesystems disposition rules to prevent silent regression, and eliminate the vestigial `GC_BOOTSTRAP` variable from production semantics.

---

## Technical Evaluation of Invariant Questions

### Q1. Does `internal/bootstrap` stop embedding production Core while keeping bootstrap tests deterministic through explicit isolated fixtures?
* **Yes, but with an implementation hazard:** The design specifies that `bootstrapAssets` defaults to a private, empty `fs.FS` returning `fs.ErrNotExist`. However, developers must be strictly forbidden from using `testing/fstest.MapFS{}` for this in production code (`bootstrap.go`). Importing `testing/fstest` in a non-test file implicitly imports the Go `testing` package, causing standard testing flags (e.g. `-test.v`, `-test.run`) to be registered into the production binary's global `flag.CommandLine`. This pollutes the production `gc` CLI command line flags.
* **Deterministic Fixtures:** Using inline mocks instead of copying real Core files preserves absolute test determinism, satisfying the Bitter-Lesson and Zero Framework Cognition (ZFC) principles.

### Q2. How is fixture drift against the shipped Core pack detected without causing low-level config tests to exercise production assets accidentally?
* **Separation of Concerns is sound:** Fixture minimality is successfully enforced by the new `TestBootstrapFixtureIsMinimal` test, which explicitly fails if production-only directories (like `formulas/`, `orders/`, etc.) are included.
* **Drift containment:** Drift detection for real Core asset loading is correctly offloaded to the integration level (`internal/packs/core` tests), meaning fast config tests remain hermetic and decoupled from production asset churn.

### Q3. Are tests needing no-Core behavior using structurally test-only lower-level loaders rather than runtime flags or environment switches?
* **Yes:** Production configuration loaders fail loudly if typed `RequiredSystemPackParticipation` records are absent. Low-level configuration tests that require bypassing this required system-pack check use compile-time test-only raw loading helpers or inline mock filesystems.
* **No backdoors:** `GC_BOOTSTRAP=skip` is correctly narrowed to a legacy-only no-op that cannot bypass required Core materialization or integrity validation.

---

**Top strengths:**
- **Go Linker safety-first Empty FS:** Definitively prevents standard `//go:embed` asset bloating in production binaries by moving `packs/**` assets entirely to `internal/packs/core` and using a zero-dependency empty `fs.FS` fallback in `internal/bootstrap/bootstrap.go`.
- **Slice-Accurate Disposition Table:** The Attempt 8 table explicitly gates emptying `bootstrapManagedImportNames` until *after* both required-Core collision enforcement and bootstrap fixture isolation are test-green, closing a severe intermediate skill-loading suppression bypass.
- **Strict Minimality Enforcement:** `TestBootstrapFixtureIsMinimal` guarantees that inline test fixtures remain minimal mock structures, preventing accidental leakage of complex Core behavior into the bootstrap test suite.

**Critical risks:**
- **[Minor] The Go Linker Flag-Pollution Hazard:** If standard standard-library mocks like `testing/fstest.MapFS` are imported in production files (`bootstrap.go` or `compose.go`) to represent the empty `bootstrapAssets` fallback, the Go `testing` package init logic will pollute production CLI flag definitions.
- **[Minor] Vestigial `GC_BOOTSTRAP=skip` Survival:** Although successfully narrowed to not bypass Core loading, keeping this environment variable in production semantics serves no operational purpose (its only valid residual use is test-internal) and risks confusing operators.
- **[Minor] Per-Test Fixture Disposition Ambiguity:** The design's statement that affected tests are rewritten "for the new `internal/packs/core` path or fixture model" is too relaxed. Without a strict rule distinguishing content tests from mechanism tests, content tests could be fixturized with synthetic mocks, silently destroying our CI signal for real Core regressions.

**Missing evidence:**
- No concrete list of which of the five listed test files (`cmd/gc/prompt_test.go`, `internal/config/bundled_import_test.go`, `examples/gastown/precompact_hook_test.go`, etc.) map to real `internal/packs/core` reads vs. synthetic inline fixtures.
- No named artifact or explicit registration table defined for the fixture allowlist consumed by the path-string guard.

**Required changes:**
- Explicitly mandate that the empty `bootstrapAssets` production fallback must be implemented as a custom private empty struct to avoid importing `testing/fstest` in production code.
- Remove `GC_BOOTSTRAP=skip` entirely from production command semantics, moving its remaining lifecycle exclusively to test-internal helpers.
- Formulate a clear file-by-file disposition matrix: content/behavior tests must use real `internal/packs/core` assets; bootstrap-mechanism tests must use inline `fstest.MapFS` fixtures.
- Run the path-string search for `internal/bootstrap/packs/core` as a pre-slice discovery command to find any uninventoried consumers before source deletion.

**Questions:**
- Can `TestBootstrapFixtureIsMinimal` be strengthened to validate against an explicit allowlist of expected files rather than a denylist of forbidden directories?
- When the bootstrap Core extraction is complete, are we ready to deprecate the compatibility layer that allows `Entry.Name = "core"` in bootstrap entirely?
