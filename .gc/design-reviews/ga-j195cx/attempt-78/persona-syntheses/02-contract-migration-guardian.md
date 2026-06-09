# Elias Vega

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] The design preserves legacy `contract = "graph.v2"` during the alias window and makes alias removal evidence-gated, but the terminal compatibility contract is not fully pinned. Claude flags missing post-removal parser behavior and diagnostics for stray `contract = "graph.v2"`; Codex flags missing source-spelling compatibility for currently accepted trimmed or case-insensitive legacy contract values. Both gaps can make external or stale formulas fail with inconsistent or unhelpful diagnostics.
- [Major] First-party migration enforcement is asymmetric. The design blocks dropping the legacy alias too early, but does not clearly fail PRs that introduce new first-party graph formulas as legacy-only or requires-only during the alias window, so the release counter can regress after review.
- [Major] The compatibility matrix does not separately fixture-lock `contract = "graph.v2"` with an empty `[requires]` table, omitted `[requires]`, and explicit `formula_compiler = ">=1"`. Claude notes these cases intentionally have different semantics and diagnostics; without side-by-side rows, implementers can conflate inheritance with conflict.
- [Major] External pinned-pack coverage is strong for exact legacy aliases, but incomplete for noncanonical legacy source spellings accepted by older binaries. Codex specifically requires exact, whitespace, case-variant, invalid-string, and dual-declaration rows in the legacy report, old-reader/probe corpus, and external fixtures.
- [Minor] Operator-visible migration progress is not concrete enough. Claude finds that warning reach is delegated to generic projection rules without a named dashboard, release-report row, or command surface, and the LRU warning-suppression counter is not included in the alias-removal gate report.
- [Minor] The interaction between `formula.compiler_requirement_missing` and `formula.version_misuse` needs projection-level clarity. Claude notes that warning suppression or confusing remediation could imply that `version = 2` is a valid bypass unless the pairing and fixture expectations are explicit.

**Disagreements:**
- There is no direct contradiction between Claude and Codex. Both verdicts are `approve-with-risks` and both accept the overall compatibility direction.
- Claude found several additional migration-contract risks that Codex did not mention: post-removal diagnostics, first-party CI symmetry, empty `[requires]` matrix rows, operator surfaces, warning projection, external-support lifecycle examples, waiver thresholds, and LRU reporting. Assessment: these are credible design risks and should be retained because they affect whether the migration contract is executable.
- Codex uniquely emphasizes current implementation compatibility for trimmed and case-insensitive `contract` values. Assessment: include it as a required change because the persona is guarding migration compatibility, and existing accepted source spellings are part of the de facto contract unless deliberately broken with evidence.

**Missing evidence:**
- Kimi 2.6 review was absent for this persona in the current attempt.
- The first-party graph formula inventory is referenced but not seeded with an expected initial set, so reviewers cannot verify the alias-removal counters are computable before implementation.
- Placeholder compatibility artifacts such as `formula-compiler-compatibility.yaml` and `formula-compiler-min-floor.json` do not yet specify which release branch, tag, or binary digest seeds concrete versions.
- No worked example shows `formula-compiler-external-support.md` moving an external consumer from `active` to `expired`, or the evidence that makes the alias-removal gate flip.
- No explicit corpus row covers currently accepted noncanonical legacy `contract` source spellings or how external packs containing them are classified.
- No aggregate signal is defined for excessive active-root waivers during Phase 8.
- The LRU eviction metric for deprecation-warning suppression is not part of the release report or alias-removal gate evidence.

**Required changes:**
- Specify the post-alias-removal parser behavior, diagnostic code, remediation text, and fixture coverage for stray `contract = "graph.v2"`. Prefer keeping the value parseable long enough to emit a migration-specific diagnostic rather than falling through to an unknown-field or missing-requirement error.
- Add a source-level legacy contract normalization policy and executable fixtures. Either preserve current trim/case-insensitive behavior during the alias window while emitting `formula.contract_deprecated` with canonical remediation, or intentionally narrow to exact `graph.v2` with a compatibility-report diagnostic and evidence that no supported external SHA-pinned pack relies on the broader behavior.
- Extend `--legacy-contract-report`, the old-reader/probe corpus, and external pinned-pack fixtures to cover exact `graph.v2`, surrounding whitespace, case variants, invalid strings, and agreeing or conflicting dual declarations.
- Add a symmetric first-party CI guard that fails any PR introducing a new first-party graph formula under the enumerated first-party paths unless it carries both `contract = "graph.v2"` and `[requires] formula_compiler = ">=2"` during the alias window.
- Add rendered compatibility-matrix and seed-fixture rows for `contract = "graph.v2"` plus empty `[requires]`, omitted `[requires]`, and explicit `formula_compiler = ">=1"`, including deterministic diagnostic order and counts.
- Add a named operator migration-progress surface, either a dashboard/release-health row or a documented `gc formula validate --legacy-contract-report --json` command, and include LRU warning-suppression counters in the alias-removal gate report.
- Clarify projection and `OnceKey` behavior for `formula.compiler_requirement_missing` plus `formula.version_misuse` so CLI/API surfaces never suppress the warning while showing the fatal, and lock that in projection-parity tests.
