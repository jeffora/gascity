# Ritu Raman - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design gives bootstrap a clear production/test seam: production `bootstrapAssets` becomes a private non-nil empty filesystem, while bootstrap tests inject inline `_test.go` `fstest.MapFS` fixtures.
- The no-Core escape hatch is tightly scoped. Normal command/runtime tests must use `internal/systempacks`, config-package tests that need no-Core behavior use lower-level config helpers, and `GC_BOOTSTRAP=skip` cannot skip Core materialization, provider host packs, retired-source classification, collision checks, cache/lock validation, or typed participation.
- Hidden dependencies are treated as a source-symbol problem, not just a path-string cleanup. The guard explicitly covers `BootstrapPacks`, `PackNames`, `bootstrapSkillDirs()`, old bootstrap Core imports, generated references, hook overlays, prompt tests, packlint fixtures, docs, and copied fixture paths.

**Critical risks:**
- [Minor] The design says tests inject bootstrap fixtures through a "test-only helper" but does not name the helper, package boundary, or production guard that prevents non-test code from using it. The intent is clear in lines 2315-2328 and 2609-2621, but the implementation contract should make the seam mechanically enforceable.
- [Minor] Fixture drift is mostly addressed by forbidding copied production Core and moving real Core fidelity tests to `internal/packs/core`/`internal/systempacks`, but the design should explicitly require a negative test that a synthetic `Entry.Name = "core"` fixture cannot be mistaken for shipped Core content or satisfy runtime required-Core participation.
- [Minor] The hidden-dependency guard is well scoped, but the final checklist should name the actual test or generated artifact that owns it. Otherwise the guard could become several local `rg` assertions with uneven coverage.

**Missing evidence:**
- The exact test-only fixture injection API and a production build/test guard proving it cannot be called from non-test code.
- A focused test where a bootstrap fixture named `core` remains synthetic and cannot satisfy `internal/systempacks` required-Core loading.
- The named hidden-dependency scanner test that fails on `internal/bootstrap/packs/core`, `AssetDir: "packs/core"`, copied production Core fixtures, `bootstrapSkillDirs()`, hook overlay reads from bootstrap, and `GC_BOOTSTRAP` production behavior branches.

**Required changes:**
- Name the bootstrap fixture injection helper and make its test-only status explicit in the design.
- Add a required negative fixture proving bootstrap's synthetic Core-shaped fixture is rejected by runtime required-Core participation and is valid only inside bootstrap unit tests.
- Tie the hidden-dependency guard to a concrete test or generated artifact in the slice gates so it stays fresh as references move.

**Questions:**
- Will `BootstrapPacks` remain an exported mutable var after this migration, or should production mutation disappear behind a `_test.go` helper?
- Should `GC_BOOTSTRAP=skip` be accepted only in tests, or can production commands still accept it solely to emit a deprecation diagnostic?
