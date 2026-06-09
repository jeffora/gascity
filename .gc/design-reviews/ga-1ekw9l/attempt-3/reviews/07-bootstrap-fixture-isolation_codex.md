# Hiroshi Tanabe - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The Current System section correctly identifies the hidden bootstrap dependency: hooks import `internal/bootstrap/packs/core`, `internal/bootstrap/bootstrap.go` still embeds `packs/**`, and tests override bootstrap with `AssetDir: "packs/core"` (`design-before.md:43`-`design-before.md:47`).
- The plan moves Core to `internal/packs/core`, rewires registry/hooks/bootstrap tests, and removes `internal/bootstrap/packs/core` only after replacement fixtures and import guards are live (`design-before.md:79`-`design-before.md:84`).
- `GC_BOOTSTRAP=skip` is explicitly contained: it may skip only empty bootstrap fixture materialization and must not skip `internal/systempacks` materialization, required Core validation, retired-source classification, collision checks, typed participation, provider materialization, or doctor cleanup (`design-before.md:318`-`design-before.md:322`).

**Critical risks:**
- [Major] The production import/fixture ban is implied but not yet specified as a concrete scanner. The persona needs a compile-time or CI check proving no production path imports `internal/bootstrap/packs/core`, references `AssetDir: "packs/core"`, or reads deleted bootstrap Core files after the move. "Import guards" should be named as exact tests with denied patterns and allowed test-only exceptions.
- [Minor] The fixture strategy says tests use empty or minimal inline `fs.FS` fixtures with asserted `Stat`, `WalkDir`, and `ReadFile` behavior, but it should also ban copied production Core snapshots by digest or path, not just by top-level directory names. A partial copy of prompts or overlays could still drift without matching the listed directories.
- [Minor] The audit list covers `GC_BOOTSTRAP` dependencies broadly, but the plan should state that `GC_BOOTSTRAP=skip` is unavailable or inert in production command paths, not only that it cannot skip required-system-pack checks.

**Missing evidence:**
- A named scanner test for production imports of `internal/bootstrap/packs/core` and references to `packs/core`.
- A fixture freshness/containment test proving bootstrap fixtures cannot copy production Core assets except through explicitly minimal inline fixtures.
- A negative test proving `GC_BOOTSTRAP=skip` cannot bypass `internal/systempacks` in a production-like command path.

**Required changes:**
- Add an explicit bootstrap-core import/path scanner with denied patterns, scan roots, and allowed test-only exceptions.
- Expand fixture guard tests to catch partial production Core snapshots, not only full directories such as `formulas/`, `orders/`, and `overlay/`.
- Define production behavior for `GC_BOOTSTRAP=skip`: ignored, refused, or test-only gated, while required-system-pack checks still run.
- Add a regression test for hooks/prompt tests that previously read `core.PackFS` from bootstrap and now must read `internal/packs/core` or a minimal fixture.

**Questions:**
- What exact CI test fails if a production file imports `internal/bootstrap/packs/core` after source removal?
- Are `AssetDir: "packs/core"` and copied fixture paths denied everywhere, or allowed in narrowly named legacy tests until deletion?
- Does `GC_BOOTSTRAP=skip` remain user-visible, or is it narrowed to test-only execution?
