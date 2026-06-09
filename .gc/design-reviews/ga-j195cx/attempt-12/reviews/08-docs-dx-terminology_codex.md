# Avery Brooks - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The design now makes `requires.formula_compiler = ">=2"` the user-facing TOML surface and explicitly says the formula declares a requirement while the active binary chooses the compiler implementation.
- The requirement grammar, diagnostic codebook, host-capability mapping, and `formula.version_deprecated` projection are much more concrete than a rename-only migration.
- The docs section correctly identifies that reference docs, config docs, generated help, tutorials, examples, test fixtures, and first-party packs must move together instead of leaving documentation as a post-implementation cleanup.

**Critical risks:**
- [Major] The rollout still allows the docs phase to trail the code phase in practice. The design says docs and examples must update in the same release phase that makes diagnostics user-visible, but the rollout sequence puts docs after parser/caller support ships. Because `formula.contract_deprecated` and `formula.version_deprecated` become API/CLI/dashboard-visible diagnostics in the caller migration, the design needs an explicit gate: no user-visible diagnostic projection or release cut until `docs/reference/formula.md`, generated config/CLI docs, public tutorials, examples, and first-party formula snippets teach `[requires]` as canonical.
- [Major] The `requires` terminology collision is still under-specified for users. The glossary says pack `requires` / `requires_gc` are distinct from formula-level `[requires]`, but pack docs already use `requires_gc`, legacy `pack.requires`, import version constraints, and dependency language. The design needs a small comparison table with TOML location, shape, owner, and resolution semantics for formula `[requires]`, `[pack].requires_gc`, legacy `[pack].requires`, and `[imports.*].version`; otherwise authors will not know which `requires` surface to edit.
- [Major] The docs inventory is still a family-level checklist, not an executable file-level migration plan. It names categories such as tutorials/examples and generated command help, but it does not enumerate the concrete stale surfaces that currently teach `version`, `contract`, `graph.v2`, schema, or formula-v2 terminology. Without a path list and source-of-truth owner for generated docs, stale examples can survive the migration.
- [Major] The generated-docs path is not precise enough. `docs/reference/config.md` is auto-generated and currently describes `formula_v2` as enabling "graph.v2 formula compilation"; updating the rendered Markdown is not the durable action. The design should name the Go struct/comment/schema-generation source that must change, plus the regeneration command and quality gate, so the public docs do not drift on the next schema regeneration.
- [Minor] The copy-paste TOML examples are not yet the final docs UX. The proposed canonical examples omit a top-level `description`, while the current reference doc teaches `version = 1` as a common top-level key. New examples should show a complete modern formula with `description`, optional `[requires]`, and no `version`; legacy or dual-declared examples should be isolated in migration notes with the exact warning behavior.

**Missing evidence:**
- A concrete public-docs gate that blocks release or diagnostic exposure until stale docs and examples are updated.
- A file-by-file inventory for `docs/reference/formula.md`, `docs/reference/config.md` source generation, `docs/reference/cli.md` source generation, tutorials, examples, first-party workflow formulas, fixtures, and PackV2 docs that mention related `requires`, `version`, `schema`, `contract`, or `graph.v2` terms.
- The source-of-truth path for generated config docs and generated CLI/help docs, including the command reviewers should run to regenerate and verify them.
- A scoped stale-guidance CI rule: corpus, forbidden patterns, allowed migration-note exceptions, fixture/testdata exceptions, and failure mode.
- A rendered final canonical formula example and a separate rendered legacy/dual-declared migration example.

**Required changes:**
- Add a release gate to the rollout: the first phase that can expose `formula.contract_deprecated`, `formula.version_deprecated`, or `formula.compiler_requirement_unsatisfied` to CLI/API/dashboard users must also land the canonical reference docs, generated help/schema updates, public examples, and tutorial updates.
- Replace the family-level docs table with a file-level checklist. At minimum include `docs/reference/formula.md`, the generated source for `docs/reference/config.md`, the generated source for `docs/reference/cli.md`, `engdocs/architecture/formulas.md`, `engdocs/proposals/formula-migration.md`, formula tutorials, public examples under `examples/`, first-party workflow formulas, and compatibility fixtures.
- Add a "Which requirement surface do I use?" subsection that distinguishes formula `[requires]`, pack `requires_gc`, legacy pack `requires`, and import `version` constraints by TOML path, shape, who evaluates it, and what failure/remediation users see.
- Specify the stale-guidance CI check before relying on it as an alias-window criterion: scanned paths, forbidden canonical-use patterns, allowed migration-note markers, test fixture allowances, and whether it runs in `make check`, docs CI, or both.
- Update the canonical TOML examples in the design so the user-facing "new formula" example includes `description`, omits `version`, and uses `[requires]` only when needed. Put `contract = "graph.v2"` only in a clearly labeled migration/compatibility example.

**Questions:**
- Should deprecation diagnostics be feature-gated until the docs/examples gate passes, or should the implementation phases land behind an unreleased branch/flag until the documentation bundle is ready?
- Where is the canonical source for `docs/reference/config.md` and `docs/reference/cli.md` text, and should the design require regenerating those docs in the same PR as the behavior change?
- Is legacy `pack.requires` still an active authoring surface, or should formula `[requires]` documentation explicitly label it as historical/migration-only wherever both appear?
