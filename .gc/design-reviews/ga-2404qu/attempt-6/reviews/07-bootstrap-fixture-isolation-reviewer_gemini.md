# Gemini 3.5 Flash — Bootstrap Fixture Isolation Review (Iteration 6)

**Lane:** 07-bootstrap-fixture-isolation-reviewer
**Scope:** Bootstrap embed cleanup, deterministic test fixtures, test-only no-Core path containment, hidden dependency discovery.
**Verdict:** approve

### Summary

In Iteration 6, the design-before document (`.gc/design-reviews/ga-2404qu/attempt-6/design-before.md`) has achieved a complete, rigorous, and highly coherent specification for the bootstrap extraction and test-fixture isolation. Every single structural gap, edge case, and pattern drift identified in prior review cycles has been thoroughly resolved and codified as explicit contract requirements under the **Bootstrap Extraction Completion Contract** (L453-488) and **Bootstrap Cleanup** (L1007-1044).

Notably:
1. **Production `bootstrapAssets` Default (B1):** The design now explicitly mandates that production `bootstrapAssets` defaults to a private, non-nil, empty `fs.FS` implementation returning `fs.ErrNotExist` for all paths, eliminating any risk of nil-pointer dereference panics.
2. **CI Fixture Guard (B2):** An automated, scanner-enforced CI validation check (`TestBootstrapFixtureIsMinimal`) is defined to prevent developers from copying production-only Core directories into the synthetic test fixtures.
3. **`GC_BOOTSTRAP=skip` Narrowing (B3):** The semantics of `GC_BOOTSTRAP=skip` are narrowed strictly to legacy bootstrap materialization (which is empty in production). It is explicitly forbidden from bypassing required Core system-pack materialization, and a test is mandated to prove Core materializes normally under this flag.
4. **Composer Sync (`bootstrapManagedImportNames`) (B4):** The composer's hardcoded lists in `internal/config/compose.go` will remove `"core"` and `"registry"` in the same slice that clears `BootstrapPacks`, and matching synchronization tests are fully aligned.
5. **Dependency Inventory Completion (M1, M2, m1, m2):** All dead string-literal paths (e.g., `cmd/gc/prompt_test.go`, `internal/config/bundled_import_test.go`, `examples/gastown/precompact_hook_test.go`, `test/packlint/*`, and documentation paths like `internal/hooks/config/README.md`) are explicitly scheduled for path-replacement and update.

The design meets the highest standards of architectural cohesion, low coupling, and SDK self-sufficiency. This review grants full approval to the bootstrap and fixture extraction architecture.

---

### Evidence-Based Findings & Satisfied Criteria

#### 1. Isolation & Production Containment (Questions 1 & 3)
* **Design Grounding:** Under the newly added *Bootstrap Extraction Completion Contract*, `internal/bootstrap/bootstrap.go` completely removes the `//go:embed packs/**` directive. To prevent panics (previously raised as B1), the private empty filesystem fallback is explicitly specified.
* **Test Isolation:** Tests will utilize inline `fstest.MapFS` fixtures instead of copying production assets or using disk-based `testdata` directories. This guarantees that config/bootstrap tests remain deterministic and decoupled from the active production Core pack files.
* **Containment of escape hatches:** The design restricts the `GC_BOOTSTRAP=skip` env var to be a compatibility switch for legacy bootstrap behavior only. Core system-pack loaders are structurally incapable of being bypassed via runtime switches or flags, satisfying the required containment invariant.

#### 2. Detection of Fixture Drift & Guarding Minimal Fixtures (Question 2)
* **Automated Guard:** The addition of `TestBootstrapFixtureIsMinimal` directly targets the risk of fixture copy-paste drift. If a developer attempts to drag-and-drop production formulas, orders, overlays, skills, or prompts into the bootstrap test setup, the CI gate will fail immediately.
* **Dual-Embed Invariant:** `internal/bootstrap/packs/core/embed.go` and its comments are completely removed, establishing a clean "single-embed" model via `internal/packs/core.PackFS`.

#### 3. Cross-File & Composer Alignment
* **Syncing imports:** By removing `"core"` and `"registry"` from the composer's `bootstrapManagedImportNames` list, we eliminate pattern drift and prevent unit test failures in `collision_test.go`.

---

### Recommendations for the Implementation Phase (Non-Blocking)

While the design is fully approved, the following best practices should be observed during implementation:
1. **Private `errFS` Implementation:** In `internal/bootstrap/bootstrap.go`, implement the empty fallback FS as follows:
   ```go
   type errFS struct{}
   func (errFS) Open(name string) (fs.File, error) {
       return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
   }
   ```
2. **Minimal Inline Fixtures:** Keep the inline `fstest.MapFS` files as thin as possible (e.g., only containing a mock `pack.toml` with the bare minimum metadata required to pass the collision and parsing tests).
3. **Execution Verification:** Ensure `TestProductionBootstrapAssetsIsEmpty` walks `bootstrapAssets` and asserts that no entries are returned under production compilation builds.
