# Avery Brooks

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] The rollout plan contradicts the documentation contract. Both reviews agree that diagnostics such as `formula.contract_deprecated`, `formula.compiler_requirement_unsatisfied`, and `formula.version_deprecated` must not become user-visible before the canonical formula reference, generated config/CLI help, tutorials, examples, and first-party snippets teach `[requires]` as the replacement surface.
- [Major] The docs migration is still a category-level checklist instead of an executable file-level plan. The design names families such as reference docs, tutorials, examples, generated command help, and first-party packs, but does not enumerate the concrete files, source-of-truth generators, regeneration commands, ownership, or verification gates needed to keep stale `version`, `contract`, `graph.v2`, schema, and formula-v2 language out of the release.
- [Major] The formula-level `[requires]` terminology collision is under-specified for users. The glossary acknowledges the distinction, but user-facing docs need a comparison of formula `[requires]`, `[pack].requires_gc`, legacy pack `requires`, and import `version` constraints by TOML location, shape, owner, evaluator, and remediation path.
- [Major] Copy-paste examples do not yet show the final authoring UX. The design says first-party formulas should be dual-declared during the alias window, but it lacks a canonical dual-declared TOML example and a complete modern formula example with `description`, optional `[requires]`, no `version`, and `contract = "graph.v2"` isolated to migration guidance.
- [Major] Generated documentation is not tied to durable source changes. `docs/reference/config.md` and `docs/reference/cli.md` can only stay corrected if the design names the Go/comment/schema-generation sources, regeneration commands, and quality gates that update rendered Markdown and generated help together.
- [Minor] Several known stale surfaces need explicit treatment: `docs/reference/formula.md` currently teaches `version = 1`, architecture docs still describe the older formula model, config/CLI docs use `graph.v2` wording, and in-repo example formulas remain legacy-only.
- [Minor] The stale-guidance CI rule is not implementable as written. The design needs scanned paths, forbidden canonical-use patterns, allowed migration-note and fixture exceptions, and the gate where the check runs.
- [Minor] Long-running diagnostic suppression still lacks enough detail for operator visibility. Claude flagged that the OnceKey bounded LRU should specify a minimum bound and re-emission policy, or choose a simpler daemon warning policy during the alias window.

**Disagreements:**
- There is no verdict disagreement: both Claude and Codex returned `approve-with-risks`.
- Claude emphasizes the direct contradiction between Phase 3 diagnostics and Phase 5 docs, including specific current files and first-party formulas. Codex frames this as a missing release gate. Assessment: both are the same release-blocking risk; the design needs both a gate and a concrete file inventory.
- Claude asks for explicit updates to `engdocs/architecture/formulas.md` and named example formula files. Codex asks for generated-doc source ownership and regeneration commands. Assessment: the required change should include both hand-authored docs/examples and generated documentation sources.
- Claude focuses on `pack.requires_gc` as the most visible overload with formula `[requires]`; Codex broadens the comparison to legacy pack `requires` and import `version` constraints. Assessment: use the broader comparison table so the user-facing docs resolve all adjacent terminology at once.
- Claude raises OnceKey LRU suppression policy as missing evidence; Codex does not. Assessment: keep it as a minor required clarification because warning visibility is part of this persona's docs/DX lane, but it is not the primary blocker.

**Missing evidence:**
- No Gemini review artifact was present, so this synthesis uses Claude and Codex only.
- No explicit release gate preventing user-visible deprecation diagnostics from shipping before the documentation and example bundle.
- No file-by-file inventory for `docs/reference/formula.md`, generated config docs, generated CLI/help docs, architecture docs, tutorials, examples, first-party workflow formulas, fixtures, and PackV2 docs that mention related terminology.
- No source-of-truth path, regeneration command, or quality gate for generated `docs/reference/config.md` and `docs/reference/cli.md` wording.
- No rendered final canonical formula example and no separate rendered legacy/dual-declared migration example with expected warning behavior.
- No "Which requirement surface do I use?" comparison for formula `[requires]`, pack `requires_gc`, legacy pack `requires`, and import `version` constraints.
- No implementable stale-guidance CI spec with corpus, forbidden patterns, exceptions, and pass/fail behavior.
- No explicit OnceKey LRU bound or re-emission policy for long-running daemons during the alias window.

**Required changes:**
- Add a rollout gate: the first phase that exposes `formula.contract_deprecated`, `formula.version_deprecated`, or `formula.compiler_requirement_unsatisfied` through CLI, API, dashboard, events, or generated help must also land the canonical reference docs, generated docs/help changes, tutorials, examples, and first-party formula updates.
- Replace the family-level docs table with a file-level migration checklist. Include at minimum `docs/reference/formula.md`, generated sources for `docs/reference/config.md`, generated sources for `docs/reference/cli.md`, `engdocs/architecture/formulas.md`, formula migration docs, PackV2 docs that mention `requires`, tutorials, public examples under `examples/`, first-party workflow formulas, testdata, and compatibility fixtures.
- Add a user-facing "Which requirement surface do I use?" section comparing formula `[requires]`, `[pack].requires_gc`, legacy pack `requires`, and `[imports.*].version` constraints by TOML path, shape, owner, evaluator, failure mode, and remediation wording.
- Add canonical TOML examples: a complete modern formula with `description`, no `version`, and `[requires]` only when needed; plus a clearly labeled migration/compatibility example showing dual declaration with `[requires] formula_compiler = ">=2"` and `contract = "graph.v2"`, including when authors may drop `contract`.
- Specify generated-doc ownership: name the Go structs/comments/templates that drive config and CLI reference text, the regeneration commands, and the quality gate reviewers must run.
- Add a per-file plan for known stale surfaces, including `docs/reference/formula.md` removing `version = 1` from the canonical minimal example, config/CLI text replacing `graph.v2` user-facing wording, `engdocs/architecture/formulas.md` being updated or superseded, and the five listed first-party example formulas moving to dual declaration.
- Define stale-guidance CI precisely: scanned corpus, forbidden canonical-use patterns for `contract`, `version`, `graph.v2`, and formula-v2 terminology, allowed migration-note/test fixture exceptions, opt-out markers if any, and whether the check runs in `make check`, docs CI, or both.
- Specify the OnceKey LRU minimum bound and re-emission policy for daemon diagnostics during the alias window, or choose a policy where warnings always remain visible for daemons until alias removal.
