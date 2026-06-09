# Hiroshi Tanabe — DeepSeek V4 Flash (Bootstrap Fixture Isolation Reviewer, Attempt 1, Independent Review)

**Verdict:** block

**Lane:** production Core embed removal, non-nil empty bootstrap fs, fixture-only tests, GC_BOOTSTRAP skip containment, hidden dependency inventory.

Reviewed against the Iteration 1 / Attempt 1 draft of `plans/core-gastown-pack-migration/implementation-plan.md` and `plans/core-gastown-pack-migration/requirements.md` in the active repository workspace.

---

## Executive Summary

As Hiroshi Tanabe, the **Bootstrap Fixture Isolation Reviewer**, I have conducted an independent, evidence-backed safety and risk audit of the Iteration 1 (Attempt 1) design for the Core and Gastown Pack Split. While other reviewers may move to `approve-with-risks` or overlook critical implementation details, my verdict is **Verdict: block**.

While the high-level goal of removing production Core assets from the bootstrap layer is correct and overdue, the proposed implementation contains several unexamined assumptions, a severe rollout-slice race condition, and Go-level compilation blindspots. Specifically, the plan:
1. Triggers an immediate compile-time error due to the removal of embed sources without resolving directive syntax.
2. Promotes a test-dependency leak by using `fstest` concepts in production code without clarifying compiler boundary limitations.
3. Leaves a dangerous rollout window in Slice 3 where bootstrap-level collision checks run in a completely toothless, disabled state before replacement gates exist.

These structural and syntactic deficiencies must be resolved before this lane can be approved.

---

## Top Strengths

- **Structural Cleanliness of Core Extraction**: Relocating the required Core assets from `internal/bootstrap/packs/core` to `internal/packs/core` (Lines 17–19) and severing the direct bootstrap dependency is structurally sound.
- **Narrowing of `GC_BOOTSTRAP=skip`**: The commitment to structurally constraint `GC_BOOTSTRAP=skip` (Lines 265–269) so that it only gates legacy empty bootstrap fixture materialization, while preventing it from skipping `internal/systempacks` validation and participation, effectively plugs a major legacy bypass route.
- **Fixture Guard Containment**: The intent to introduce fixture guard tests that fail if test paths attempt to copy production Core subtrees is an excellent mechanism to prevent test-only crutch drift.

---

## Critical Risks & Assumptions Accepted Too Quickly

### 1. [Blocker] Unresolved Empty Embed Directive Causes Hard Compilation Failure
- **The Assumption**: The plan states that "The Core extraction slice deletes the `//go:embed packs/**` directive and the `embeddedBootstrapPacks` variable" (Lines 259–260) to avoid build errors.
- **The Reality**: In `internal/bootstrap/bootstrap.go`, the variable `bootstrapAssets` is declared as a package-private `fs.FS` backed by this embed. Simply deleting the `//go:embed` directive without specifying how `bootstrapAssets` is initialized will cause immediate compile-time errors in any file importing `internal/bootstrap` or referencing `bootstrapAssets` (specifically at lines `:176/:217/:220/:241/:254` of `bootstrap.go`).
- **The Blocker**: If `bootstrapAssets` is left as `nil` or pointing at the deleted embed, standard calls like `fs.Stat`, `WalkDir`, or `ReadFile` will panic instead of returning expected file-not-found errors. The plan fails to specify the exact Go-level structure that `bootstrapAssets` adopts post-removal.
- **Required Change**: Specify that `bootstrapAssets` is re-declared as a concrete, non-nil empty `fs.FS` type that safely returns `fs.ErrNotExist` for all operations.

### 2. [Blocker] Production-Side Test-Dependency Leak (The `fstest` Dilemma)
- **The Assumption**: Other reviewers have accepted that `bootstrapAssets` can "default to an empty `fs.FS` implementation" such as `fstest.MapFS{}`.
- **The Reality**: `fstest` is a sub-package of `testing`. Under Go compilation invariants, importing `testing` or `testing/fstest` in production code (like `internal/bootstrap/bootstrap.go`) is strictly forbidden because it leaks testing command-line flags, slows compilation, and bloats the production binary.
- **The Blocker**: If a developer implements the empty `fs.FS` default using `fstest.MapFS{}`, the build will fail production-import validation gates.
- **Required Change**: The plan must explicitly mandate that the empty `fs.FS` is written as a custom, private, lightweight struct in `internal/bootstrap/bootstrap.go` (e.g., `type emptyFS struct{}` implementing `fs.FS`) to ensure it is completely isolated from test-only packages.

### 3. [Major] The Slice 3 Bootstrap Collision Protection Gap (Rollout Race Condition)
- **The Assumption**: The plan states that the migration "Keeps existing bootstrap collision protection... active through Slice 3 (Core Extraction) and removed only in Slice 4" (Lines 210–213).
- **The Reality**: Deleting the embedded `packs/core` files in Slice 3 means there is no longer any bootstrap-embedded Core to collide with!
- **The Blocker**: Because the bootstrap-level collision logic relies on checking the embedded pack set, the moment the files are deleted in Slice 3, the collision check becomes a complete no-op (toothless) for Core. There is a dangerous "collision protection gap" during Slice 3 before the new `internal/systempacks` required-pack collision gates are introduced in Slice 4.
- **Required Change**: Explicitly resolve this rollout gap. The plan should either accelerate the introduction of `internal/systempacks` collision gates to Slice 3 or acknowledge this temporary gap and document the explicit mitigation or risk acceptance.

### 4. [Major] Vestigial Doctor Env-Mutation Dance Left Unaddressed
- **The Assumption**: The plan notes that `internal/doctor/implicit_import_cache_check.go` unsets and restores `GC_BOOTSTRAP` (Line 307).
- **The Reality**: Once `GC_BOOTSTRAP=skip` is narrowed to skipping only empty bootstrap fixture setup and Core materialization/validation is moved to `internal/systempacks`, unsetting `GC_BOOTSTRAP` inside the doctor check does absolutely nothing.
- **The Blocker**: Leaving the environment-mutation logic inside the doctor check creates dead, confusing, and potentially buggy state-manipulation paths.
- **Required Change**: Explicitly mandate the post-retirement cleanup of `implicit_import_cache_check.go`'s `GC_BOOTSTRAP` unset/restore logic in the Core extraction slice.

### 5. [Minor] Fragile Hardcoded Fixture-Guard Prohibited List
- **The Assumption**: The fixture guard test fails if test paths copy directories like `formulas/`, `orders/`, `overlay/`, `skills/`, or `assets/prompts/` (Lines 268–269).
- **The Reality**: If Core gains a new directory in a future release (such as `templates/` or `hooks/`), a hardcoded list in the guard test will silently fail to protect it, allowing test-only copy drift to re-emerge.
- **Required Change**: Mandate that the fixture-guard prohibited list is dynamically derived from the top-level directories of the live `internal/packs/core` package, rather than being hand-curated and prone to rot.

---

## Missing Evidence

1. **Concrete Go-level struct definition** for the production-side empty `fs.FS` in `internal/bootstrap/bootstrap.go`.
2. **Explicit disposition** of `implicit_import_cache_check.go`'s `GC_BOOTSTRAP` unset/restore logic.
3. **The exact injection seam** in `internal/bootstrap` that allows `collision_test.go` to inject synthetic fixtures without relying on hardcoded `AssetDir: "packs/core"` strings.
4. **Focused test script regression** proving a command executed with `GC_BOOTSTRAP=skip` still fails closed when required Core is absent.

---

## Required Changes

1. **Implement a Custom Empty FS**: Specify that the fallback `bootstrapAssets` is a custom, private empty `fs.FS` defined in `internal/bootstrap/bootstrap.go` with zero production imports of `testing` or `testing/fstest`.
2. **Close the Rollout Gap**: Explicitly address the Slice 3 collision protection gap; require that system-pack level collision gates are introduced in parallel with Core extraction or define the risk boundary.
3. **Clean Up Doctor Mutation**: Mandate that the `GC_BOOTSTRAP` environment-variable mutation dance in `implicit_import_cache_check.go` is completely deleted or cleaned up.
4. **Dynamic Fixture-Guard List**: Assert that the test-fixture guard derives its prohibited directories dynamically from the actual subdirectories of `internal/packs/core`.
5. **Name the Injection Seam**: Explicitly specify that the `bootstrapAssets` variable is exposed as an internal package injection seam to let `collision_test.go` and other tests override it safely.

---

## Questions

### Q1: After removing production Core from bootstrap embeds, what compile-time or CI check proves no production path imports the deleted bootstrap Core package?
**Answer**:
Once `internal/bootstrap/packs/core` is deleted in Slice 3, any production Go file that attempts to import `github.com/gastownhall/gascity/internal/bootstrap/packs/core` will immediately fail to compile (`go build` and `go test` fail with package-not-found errors). This provides a hard compile-time guarantee.
For non-Go files (such as TOML configs, scripts, and documentation), a static path scanner test (modeled on `cmd/gc/worker_boundary_import_test.go` and integrated into `test/packlint`) must be added to grep for the deleted path string and assert zero references outside explicit historical/migration docs.

### Q2: Are tests that need Core assets using minimal fstest fixtures or the relocated system-pack wrapper, not copied production Core snapshots?
**Answer**:
Yes. Bootstrap and config tests requiring Core assets must use an empty `fs.FS` or minimal inline `fstest.MapFS` fixtures where behavior is asserted. Copying production Core subtrees is strictly prohibited. The dynamic fixture-guard test ensures that any attempt by a test path to copy production Core subdirectories results in a build failure, preventing drift.

### Q3: Is GC_BOOTSTRAP=skip narrowed to fixture or no-Core tests and structurally unreachable as a production required-system-pack bypass?
**Answer**:
Yes. `GC_BOOTSTRAP=skip` is structurally narrowed to skip only empty bootstrap fixture setup. Because Core materialization, Required Core integrity validation, and Core config participation are moved to `internal/systempacks` (which does not read `GC_BOOTSTRAP`), the skip flag is completely unreachable as a bypass. Tests must prove that a command run with `GC_BOOTSTRAP=skip` still fails closed if required Core is missing or corrupted on disk.
