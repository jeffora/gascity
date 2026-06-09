# Hiroshi Tanabe

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] The implementation plan conflicts with the requirements artifact on Core's canonical source. Codex flags that requirements still name `internal/bootstrap/packs/core` as the sole canonical Core authority and AC2 target, while the implementation plan moves Core to `internal/packs/core` and deletes/removes the bootstrap Core asset source. This must be reconciled before approval or implementers can satisfy stale ACs by preserving the old path.
- [Blocker] `GC_BOOTSTRAP=skip` containment is asserted but not structurally proven. All reviewers require a concrete mechanism and test proving skip cannot bypass required system-pack materialization, file-set integrity, participation validation, retired-source classification, or doctor cleanup. The plan must also decide whether to remove, narrow, or rename skip and what happens to the current doctor unset/restore dance and broad test defaults.
- [Blocker] Deleting `internal/bootstrap/packs/core` must be coupled with embed removal and `EmptyFS` initialization. Claude and DeepSeek both note that `//go:embed packs/**` will fail if the package deletion leaves no matching files. Slice 3 must make deletion, directive removal or placeholder handling, and `bootstrapAssets`/`embeddedBootstrapPacks` defaulting an indivisible build-preserving change.
- [Blocker] The rollout creates a collision-check gap unless replacement system-pack collision validation lands with Core extraction. DeepSeek flags that deleting embedded Core in Slice 3 can make the old collision guard toothless before Slice 4a introduces the new loader validation. The plan needs fail-closed collision protection active during the extraction slice.
- [Major] The hidden-dependency inventory is not anchored to a checked closure artifact. Reviewers found compiler-visible imports plus string-only production/test paths, `Subpath` values, security fixtures, prompt asset reads, docs/examples, generated paths, `GC_BOOTSTRAP` readers, and the embed directive. The plan must name the inventory artifact, row schema, owner/replacement/classification fields, and validator that runs before old package deletion.
- [Major] The proof must combine compile failure and scanner gates. Package deletion catches only direct Go imports; it does not catch `Subpath` strings, constructed paths, docs, generated files, or tests that can silently point at deleted paths. The scanner must explicitly cover the known string-only sites, especially security tests that would otherwise pass vacuously.
- [Major] Test-isolation guidance conflates mechanism tests with content-fidelity tests. Empty/minimal `fs.FS` fixtures are right for bootstrap mechanism tests, but prompt tests, hook tests, and synthetic-repo security/tamper tests need the relocated real Core assets through `internal/packs/core` or the system-pack wrapper.
- [Major] Production `EmptyFS` must not import `testing` or `testing/fstest`. DeepSeek flags that production code must implement a small custom `fs.FS` rather than using test-only helpers. Test code may use `fstest`, but production bootstrap code must not.
- [Minor] `EmptyFS.Open(".")` must return an `fs.ReadDirFile`-compatible empty directory so `fs.WalkDir`, `Stat`, and doctor traversal behavior are well-defined.
- [Minor] Fixture copy guards should be generated from actual relocated Core content, not only a hand-curated denylist of current top-level directories, so new Core directories cannot be copied into tests unnoticed.

**Disagreements:**
- Claude returns `approve-with-risks`, while Codex and DeepSeek block. My assessment: the persona verdict is `block` because the Codex requirements conflict and DeepSeek embed/collision gaps are structural approval blockers, not just implementation nits.
- Claude treats production runtime as neutral because current `BootstrapPacks` is empty in production. Codex and DeepSeek focus on decomposition safety and rollout correctness. My assessment: production no-op status reduces immediate runtime risk but does not eliminate the need for explicit gates before deleting or moving the asset source.
- DeepSeek proposes a placeholder file as one possible embed fix. Claude prefers coupling deletion with removal of `//go:embed packs/**` and `EmptyFS` defaults. My assessment: either mechanism can work, but the plan must choose one and make it atomic with package deletion.
- DeepSeek frames production `fstest` use as a blocker; Claude only says `fstest.MapFS{}` is natural for tests. My assessment: clarify the boundary: production `EmptyFS` is custom and test-only fixtures may use `testing/fstest`.
- Claude and Codex emphasize scanner/inventory coverage; DeepSeek emphasizes collision-gate rollout timing. Both are required: closure proves no stale readers remain, while replacement collision gates preserve fail-closed behavior during extraction.

**Missing evidence:**
- Updated requirements or an explicit prerequisite showing that AC2 no longer treats `internal/bootstrap/packs/core` as the post-migration Core authority.
- A named hidden-dependency inventory artifact with concrete rows for direct imports, `Subpath` strings, string/path test sites, prompt asset reads, security fixtures, generated/docs references, `GC_BOOTSTRAP` readers, and the `//go:embed packs/**` directive.
- A scanner allowlist and validator proving no production or test path silently references the deleted bootstrap Core path except intentional migration/history text.
- A structural `GC_BOOTSTRAP=skip` containment design and positive production command/loader test with skip set and missing or corrupt Core.
- The exact fate of `ensureBootstrapForDoctor` environment unset/restore behavior and the broad `cmd/gc/main_test.go` skip default.
- A Slice 3 sequencing rule that preserves buildability when `internal/bootstrap/packs/core` is deleted.
- Replacement system-pack collision checks active in the extraction slice or an explicit bounded mitigation.
- A split list of mechanism-only bootstrap tests versus content-fidelity tests that must read relocated real Core assets.
- A production `bootstrap.EmptyFS` contract proving `Open(".")`, `ReadDir`, `Stat`, `WalkDir`, and `ReadFile` behavior without `testing` imports.

**Required changes:**
- Reconcile the requirements and implementation plan so both agree on the post-migration Core source authority; if `internal/packs/core` is the end state, update AC2 and related examples before approval or make that update an explicit prerequisite.
- Specify the structural containment mechanism for `GC_BOOTSTRAP=skip`: remove it from production, make it test-only by construction, or add a production negative test proving required system-pack materialization and validation still run under skip.
- Decide and document the fate of `ensureBootstrapForDoctor`'s environment unset/restore dance and the suite-wide test skip default.
- Make Slice 3 atomic for build preservation: delete or move `internal/bootstrap/packs/core`, remove or satisfy `//go:embed packs/**`, default `bootstrapAssets`/`embeddedBootstrapPacks`/`BootstrapPacks` to a non-nil `EmptyFS`, and repoint `registry.go` `Subpath` strings in the same slice.
- Introduce replacement system-pack collision checks in parallel with Core extraction, or document a bounded fail-closed mitigation that prevents a toothless collision window.
- Name the hidden bootstrap dependency inventory artifact, define its schema, seed it with the known file/path rows, and require its validator before old package deletion, old-path scanner enforcement, or fixture rewrites merge.
- Reframe the Q1 proof as compile-time import deletion plus string/path scanner coverage over `internal/`, `cmd/gc`, docs, examples, fixtures, and generated files.
- Split test guidance: mechanism-only tests use empty/minimal inline FS fixtures, while content-fidelity tests read relocated Core assets through `internal/packs/core` or `internal/systempacks`.
- Implement production `bootstrap.EmptyFS` as a custom lightweight `fs.FS` with no `testing` or `testing/fstest` imports, and require tests for `Open(".")`, `ReadDir`, `Stat`, `WalkDir`, and `ReadFile`.
- Derive fixture copy-ban inputs from the actual relocated Core tree, not only from a fixed list of currently known directories.
