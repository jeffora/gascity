# Elias Vega

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] Alias removal is evidence-based in intent, but the executable enforcement path is not fully specified. The design says removal requires zero first-party `legacy_only` and zero first-party `dual_declared` formulas, yet the named legacy report exits successfully when only `legacy_only` is zero; it needs a distinct gate or CI job that blocks on both counts and on unreadable inventory.
- [Major] First-party versus external pack classification is load-bearing but implicit. Legacy reports, dual-declaration conversion gates, requires-only rollout, and external-support expiration all depend on this split, so release decisions still contain discretion unless the predicate is normative and fixture-locked.
- [Major] External-support expiration lacks checkable notice and opt-out evidence. Both reviews identify that `formula-compiler-external-support.md` needs required fields for published migration notice, publication date or release tag, and maintainer opt-out evidence before any `active` row can become `expired`.
- [Major] Warning reach and accepted-alias durability are asserted more strongly than they are fixture-proven. The design requires accepted legacy `contract = "graph.v2"` use to surface across CLI, API, dashboard, dispatch, convergence, fanout, retry, and release validation, but it does not name the golden fixture or restart-durability test that would keep those projections synchronized.
- [Minor] The release artifact stack needs conservative seed files before rollout. The compatibility, min-floor, external-support, and stale-guidance artifacts are part of the safety model, but the current repository snapshot does not contain their initial blocking defaults.
- [Minor] Diagnostic precedence has a gap for legacy `contract = "graph.v2"` combined with unsupported future `[requires]` such as `formula_compiler = ">=3"` or unknown future axes. The intended precedence is inferable, but it is not in the matrix or golden fixtures.
- [Minor] Branch/tag-pinned external packs are ambiguous. One review reads the design as covering tag, branch, local, and transitive external packs through the support artifact; the other reads branch/tag references as non-reproducible unless a lockfile records the resolved revision. The migration guidance must state which interpretation controls.
- [Minor] The two-minor-release and sixty-day removal floor needs a schema-backed release checklist row. The policy names tags, dates, supported readers, inventory paths, and compatibility-corpus results, but not the executable artifact shape.

**Disagreements:**
- There is no verdict disagreement: Claude and Codex both returned `approve-with-risks`.
- Claude emphasizes the missing first-party/external classification predicate, while Codex emphasizes the missing alias-removal command or CI contract. Assessment: these are complementary; a correct removal gate needs both the classification rule and the executable command that applies it.
- Codex describes externally pinned SHA/tag/branch/local/transitive packs as covered until the external-support artifact expires them. Claude flags branch/tag pins as mutable unless lockfile-pinned. Assessment: this ambiguity is itself a contract risk; the design must explicitly define whether tag/branch refs count as pinned for alias-removal purposes and document the lockfile escape hatch if they do not.
- Claude asks for broad accepted-alias projection and restart-durability fixtures. Codex focuses on seeding release artifacts and stale-guidance config before exposing diagnostics. Assessment: both are required at different layers: fixtures prove runtime/reporting behavior, while release artifacts keep rollout and removal fail-closed.

**Missing evidence:**
- Gemini review was absent for this persona in the current attempt.
- A normative first-party/external classification rule and fixture set.
- A named alias-removal gate command or CI job with exit semantics for `legacy_only > 0`, `dual_declared > 0`, unsupported future requirements, and I/O or inventory-read failures.
- Seed contents and schemas for `docs/release/formula-compiler-compatibility.yaml`, `formula-compiler-min-floor.json`, `formula-compiler-external-support.md`, and `formula-compiler-stale-guidance.yaml`.
- Required external-support fields for `notice_url`, `notice_published_at`, `notice_release_tag`, and maintainer opt-out evidence.
- A release-checklist schema that records completed minor release tags, dates, supported reader set, legacy inventory path, compatibility-corpus result, and binary version used to generate saved reports.
- An `accepted-legacy-contract` golden fixture covering CLI, API, Huma response, generated TypeScript, dashboard state, accepted-artifact warnings, release report rows, and restart recovery of suppressed-warning evidence.
- A combined-defect precedence fixture for `contract = "graph.v2"` plus future unsupported `[requires]` values.
- A consumer-facing statement of the post-removal diagnostic code and wording for legacy `contract = "graph.v2"`.

**Required changes:**
- Add a normative first-party/external classification section with a deterministic predicate and fixture-locked tests, then cross-reference it from legacy reports, dual-declaration ownership, requires-only conversion gates, and external-support expiration.
- Define one explicit alias-removal enforcement path, such as `gc formula validate --all-packs --legacy-contract-report --alias-removal --json`, whose nonzero exit blocks removal when first-party `legacy_only > 0`, first-party `dual_declared > 0`, unsupported future requirements are present, or the report cannot read required inputs. Cite that exact path in the removal criteria, rollout gates, CI, and release checklist.
- Seed the release artifacts and stale-guidance config with conservative blocking defaults before any first-party formula conversion or user-visible `[requires]` diagnostic rollout.
- Extend `formula-compiler-external-support.md` so expired rows require migration-notice URL, publication date, release tag, linked inventory evidence, and defined maintainer opt-out evidence. Make CI reject missing or stale expiration evidence.
- State whether branch and tag references count as pinned external packs for alias-removal purposes; if they do not, document the lockfile-pinning escape hatch in external author guidance and migration notices.
- Add the accepted-legacy-contract and future-requirement precedence fixtures named above, including restart-durability coverage for suppressed accepted-alias warnings.
- Add a schema-backed release checklist row for the two-minor-release and sixty-day floor, including tags, dates, reader set, inventory path, compatibility-corpus result, and report-producing binary version.
- Update or supersede stale formula authoring docs in the same PR stack that exposes `[requires]` diagnostics, especially `docs/reference/formula.md` and `engdocs/architecture/formulas.md`.
