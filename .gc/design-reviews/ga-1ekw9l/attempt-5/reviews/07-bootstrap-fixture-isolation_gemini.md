# Hiroshi Tanabe — Gemini (Bootstrap Fixture Isolation Reviewer, Attempt 5, Independent DeepSeek V4 Flash Style)

**Verdict:** block

> **Lane:** production Core embed removal, non-nil empty bootstrap `fs.FS`, fixture-only tests, `GC_BOOTSTRAP=skip` containment, hidden-dependency inventory.
>
> Reviewed against the Attempt 5 design/implementation plan (`plans/core-gastown-pack-migration/implementation-plan.md`, 657 lines, `updated_at: 2026-06-09T07:28:00Z`) — §"Bootstrap Fixture Isolation" (lines 370–396), §"Current System" (lines 28–50), §"Proposed Implementation" (lines 51–284), §"Testing" (lines 485–560), and §"Rollout And Recovery" (lines 561–640).
>
> This independent review is produced using the DeepSeek V4 Flash persona, focusing specifically on cross-document consistency, missing edge cases, and assumptions other reviewers may accept too quickly.

---

## Executive Summary

As Hiroshi Tanabe, the **Bootstrap Fixture Isolation Reviewer**, I have conducted an independent, evidence-backed architectural and risk audit of the current 657-line `implementation-plan.md` (updated 2026-06-09T07:28:00Z) for Attempt 5.

While the high-level goal of removing production Core assets from the bootstrap layer is structurally correct and essential for SDK role-neutrality, the proposed implementation contains several critical compilation blindspots, unexamined rollout-slice race conditions, and test-dependency leaks. Specifically, the plan:
1. **Triggers a Hard Compilation Blocker**: Deleting `internal/bootstrap/packs/core` without explicitly removing or retargeting the `//go:embed packs/**` directive (line 44) will immediately crash the Go compiler during master branch builds with a fatal `pattern packs/**: no matching files found` error.
2. **Promotes a Production-Side Test-Dependency Leak**: Suggesting the use of standard library test concepts (like `fstest.MapFS{}`) inside production `internal/bootstrap/bootstrap.go` code introduces a forbidden dependency on `testing` packages, leaking command-line flags and bloating the production binary.
3. **Creates a Dangerous Rollout Window in Slice 3**: Deleting Core embeds in Slice 3 disables legacy bootstrap-level collision checks entirely before the new required-system-pack collision gates are introduced in Slice 4, exposing the system to silent configuration drift.
4. **Conflates Mechanism and Content Tests**: Asserting a blanket rule that tests must use an empty/minimal `fs.FS` fixture (line 379) directly contradicts content-fidelity tests (e.g., hooks overlays, prompts) which must read real relocated assets from `internal/packs/core`.

I must **block** this plan until these critical gaps are resolved and the required changes are explicitly integrated.

---

## Detailed Responses to Lane-Specific Questions

### Q1: After removing production Core from bootstrap embeds, what compile-time or CI check proves no production path imports the deleted bootstrap Core package?

**Answer:**
Once `internal/bootstrap/packs/core` is deleted in Slice 3, any production Go file that attempts to import `github.com/gastownhall/gascity/internal/bootstrap/packs/core` will immediately fail to compile (`go build` and `go test` fail with package-not-found errors). This provides a hard compile-time guarantee.

However, the plan's proposed "import guards" are poorly defined. To ensure absolute safety, the following must be explicitly mandated:
1. **Compile-Time Deletion Proof**: Package deletion is the primary, absolute compile-error proof for Go imports.
2. **Static Path-String Scanner**: For non-Go files (such as TOML configs, scripts, and documentation) and Go string literals (like `Subpath: "internal/bootstrap/packs/core"` in `registry.go:53`), a static path scanner test (integrated into `test/packlint` or modeled on `cmd/gc/worker_boundary_import_test.go`) must be added to grep for the deleted path string and assert zero references outside explicit historical/migration docs.

### Q2: Are tests that need Core assets using minimal fstest fixtures or the relocated system-pack wrapper, not copied production Core snapshots?

**Answer:**
The current plan conflates two different axes: mechanism tests and content tests.
1. **Mechanism-only tests** (such as bootstrap FS checks) must use a custom, private, non-nil empty `fs.FS` or minimal inline fixtures.
2. **Content-fidelity tests** (such as `hooks_test.go` asserting overlays or `prompt_test.go` asserting prompt content) *cannot* run against an empty FS. They must read relocated assets from `internal/packs/core`.

The plan's blanket instruction that "tests that need bootstrap assets use an empty `fs.FS` fixture or minimal inline fixture" (lines 378–380) contradicts the needs of these content-fidelity tests. The plan must split the guidance: mechanism-tests use empty/minimal `fs.FS` inline, while content-tests read from the relocated `internal/packs/core`. Furthermore, the audit list must be split into env-axis versus path-axis to capture all path-coupled dependents (such as `prompt_test.go` and `bundled_import_test.go`) that a simple `GC_BOOTSTRAP` grep would miss.

### Q3: Is GC_BOOTSTRAP=skip narrowed to fixture or no-Core tests and structurally unreachable as a production required-system-pack bypass?

**Answer:**
No, not yet in the current text. While the plan asserts that `GC_BOOTSTRAP=skip` is narrowed to skip only empty bootstrap fixture setup, it fails to provide a negative test or address suite-wide default behaviors.
To achieve true containment, the plan must:
1. **Add a Negative Containment Test**: Define a test script or command test that sets `GC_BOOTSTRAP=skip`, runs with missing/corrupt required Core, and asserts that the command still fails closed (proving that `internal/systempacks` required-pack validation completely ignores the flag).
2. **Address Suite-Wide Defaults**: Address the suite-wide `GC_BOOTSTRAP=skip` default in `cmd/gc/main_test.go:45`. Once the flag is narrowed, this default becomes a near-no-op, and the post-retirement disposition must be explicitly documented.
3. **Resolve Doctor Mutation**: The doctor check's unset/restore dance in `implicit_import_cache_check.go:236-245` must be cleaned up or deleted, as it becomes vestigial once `GC_BOOTSTRAP` is retired as a production behavior switch.

---

## Critical Risks & Assumptions Accepted Too Quickly

### 1. [Blocker] Unresolved Empty Embed Directive Causes Hard Compilation Failure
* **The Assumption**: Other reviewers assume deleting `internal/bootstrap/packs/core` is simple.
* **The Reality**: `internal/bootstrap/bootstrap.go:22` is `//go:embed packs/**`. `internal/bootstrap/packs/` contains *only* `core`. Deleting `internal/bootstrap/packs/core` means `packs/**` matches zero files, which causes a hard Go compilation error: `pattern packs/**: no matching files found`.
* **The Blocker**: The plan states "Production bootstrap no longer embeds Core assets" (line 378) but does not state that the `//go:embed` directive is removed or retargeted, nor that `bootstrapAssets` is re-declared. This causes immediate compilation failure on the master branch.
* **Required Change**: State explicitly that the `//go:embed packs/**` directive is removed/retargeted, and that `bootstrapAssets` is re-initialized to a concrete, non-nil empty `fs.FS`.

### 2. [Blocker] Production-Side Test-Dependency Leak (The `fstest` Dilemma)
* **The Assumption**: The plan suggests defining a non-nil production `bootstrap.EmptyFS` implementation (lines 384–386).
* **The Reality**: If developers use `fstest.MapFS{}` to implement this in production `internal/bootstrap/bootstrap.go`, they will import `testing/fstest` in production code. Importing any part of the `testing` package in production code is strictly forbidden as it leaks command-line flags and bloats the production binary.
* **Required Change**: The plan must explicitly mandate that `EmptyFS` is implemented as a custom, lightweight, package-private struct in `internal/bootstrap/bootstrap.go` (e.g., `type emptyFS struct{}` implementing `fs.FS`) to prevent importing any `testing` sub-packages in production.

### 3. [Major] The Slice 3 Bootstrap Collision Protection Gap (Rollout Race Condition)
* **The Assumption**: The plan states that the migration "Keeps existing bootstrap collision protection... active through Slice 3 (Core Extraction) and removed only in Slice 4" (lines 247–249).
* **The Reality**: The moment the embedded `packs/core` files are deleted in Slice 3, the bootstrap-level collision logic has nothing to collide with!
* **The Blocker**: Because the old collision logic relies on comparing against the embedded asset set, deleting the files in Slice 3 makes the collision check a complete no-op for Core. This creates a dangerous rollout window in Slice 3 where no collision protection exists, before the new required-system-pack collision gates are introduced in Slice 4.
* **Required Change**: Close this gap. The plan must either introduce the required-system-pack collision gates in the same slice as Core embed removal (Slice 3) or explicitly document the temporary risk boundary and mitigation.

### 4. [Minor] Fragile Hardcoded Fixture-Guard Prohibited List
* **The Assumption**: The plan states that fixture guard tests fail if test paths copy directories like `formulas/`, `orders/`, `overlay/`, `skills/`, or `assets/prompts/` (lines 381–382).
* **The Reality**: If Core gains a new directory in a future release, a hardcoded list in the guard test will silently fail to protect it, allowing test-only copy drift to re-emerge.
* **Required Change**: Mandate that the fixture-guard prohibited list is dynamically derived from the actual subdirectories of the live `internal/packs/core` package, rather than being hand-curated and prone to rot.

---

## Missing Evidence

1. **Concrete Go-level struct definition** for the production-side empty `fs.FS` in `internal/bootstrap/bootstrap.go` (preventing `testing` imports).
2. **Explicit disposition** of `implicit_import_cache_check.go`'s `GC_BOOTSTRAP` unset/restore logic.
3. **The exact injection seam** in `internal/bootstrap` that allows `collision_test.go` to inject synthetic fixtures without relying on hardcoded `AssetDir: "packs/core"` strings.
4. **Focused test script regression** proving a command executed with `GC_BOOTSTRAP=skip` still fails closed when required Core is absent.

---

## Required Changes

1. **Implement a Custom Empty FS**: Specify that `bootstrapAssets` is a custom, private empty `fs.FS` defined in `internal/bootstrap/bootstrap.go` with zero production imports of `testing` or `testing/fstest`.
2. **Close the Rollout Gap**: Explicitly address the Slice 3 collision protection gap; require that system-pack level collision gates are introduced in parallel with Core extraction or define the risk boundary.
3. **Clean Up Doctor Mutation**: Mandate that the `GC_BOOTSTRAP` environment-variable mutation dance in `implicit_import_cache_check.go` is completely deleted or cleaned up.
4. **Dynamic Fixture-Guard List**: Assert that the test-fixture guard derives its prohibited directories dynamically from the actual subdirectories of `internal/packs/core`.
5. **Name the Injection Seam**: Explicitly specify that the `bootstrapAssets` variable is exposed as an internal package injection seam to let `collision_test.go` and other tests override it safely.
6. **Split Test Guidance & Inventory**: Split the fixture guidance into mechanism-tests (empty/minimal `fs.FS`) vs content-tests (`internal/packs/core`), and split the audit into env-axis vs path-axis, explicitly listing path-coupled sites.

---

## Key Open Questions for the Mayor

1. Should we completely remove `GC_BOOTSTRAP` from the codebase in Slice 4 once `internal/systempacks` is live, or does it have a long-term testing role?
2. What specific mechanism should be used to dynamically list subdirectories of the relocated Core pack inside the test-fixture guard?
3. During Slice 3, should we temporarily mock or disable the collision check if it fails due to empty embeds, or accelerate the Slice 4 checks?

---

## Schema Conformance (in scope)

The plan conforms perfectly to the `gc.mayor.implementation-plan.v1` schema:
- All required front-matter fields (including `plan_slug`, `phase`, and `requirements_file`) are present.
- All seven top-level sections are correctly named and ordered (Summary, Current System, Proposed Implementation, Data And State, Testing, Rollout And Recovery, Open Questions).
- The "Open Questions: None" section contains correct explanatory prose and references to external prerequisites (lines 652-656).
