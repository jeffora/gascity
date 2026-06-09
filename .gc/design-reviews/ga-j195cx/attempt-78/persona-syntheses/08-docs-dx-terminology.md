# Avery Brooks

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Info] Both reviewers agree the design now has a strong terminology contract. It separates formula compiler capability, compiler implementation, deprecated `contract`, physical schema, pack revision, formula `version`, formula `[requires]`, and pack `requires_gc` clearly enough for docs and diagnostics to converge.
- [Info] Both reviewers agree the docs rollout is intended to be executable, with `make formula-docs-check`, doctest fixtures, stale-guidance scanning, first-party inventory, generated help/schema/API/TS refreshes, and release artifacts acting as phase gates before user-visible diagnostics.
- [Major] The proposal and architecture docs need to be aligned before requirement diagnostics become visible. Claude found current proposal guidance that still describes `GC_NATIVE_FORMULA=false` as a production-runtime rollback toggle, while the design now treats that path as validation-only probing.
- [Major] The docs gate should catch silence as well as stale statements. Claude found that architecture/reference docs can omit `[requires]`, `formula_compiler`, `formula_v2`, or `requires_gc` while still passing matchers that only look for explicit stale tokens.
- [Minor] The `requires` naming surface remains easy to misuse unless the reference rewrite includes a side-by-side "common confusion" callout for formula `[requires]`, pack `[pack].requires_gc`, and lockfile/resolver revision concepts.
- [Minor] JSON formula loader status remains a docs/DX fork. If JSON formula loading remains enabled, JSON examples and parity fixtures must accompany the TOML reference rewrite; if it is retired first, the design should name the phase and remove JSON rows from the docs/check matrix.

**Disagreements:**
- Codex returned `approve` with no required changes, judging the design sufficient because it blocks user-visible diagnostics until the documentation, examples, generated artifacts, and reports land. Claude returned `approve-with-risks` because several docs currently teach or omit behavior in ways that would mislead operators during the rollout. Assessment: use Claude's stricter verdict for this persona because the lane is specifically responsible for docs/DX terminology failure modes.
- Codex treated missing implementation artifacts as acceptable design-phase evidence because the rollout makes them future gates. Claude wanted more precise gating, especially for proposal supersession, positive-content checks, and JSON loader disposition. Assessment: the design can pass, but only with required rollout changes that make those gates explicit.
- Codex considered the prior open questions closed: exact accepted requirement values, warning transport, suppression rules, and legacy `version` metadata behavior. Claude agreed those questions are largely closed but still asked how the reference docs should teach warning-only `version = 2` cases, dual-declared alias-window examples, and future `[requires]` axes. Assessment: these are documentation acceptance details rather than blockers to the compiler model.

**Missing evidence:**
- The actual `docs/reference/testdata/formula-requirements-doctest.yaml` fixture schema is not shown, so reviewers cannot validate how valid, invalid, and template examples will be classified.
- The `docs/release/formula-compiler-docs-check.json` and stale-guidance report schemas are not detailed enough to prove expected pre-migration findings will be separated from false positives.
- There is no concrete proof that architecture/proposal docs must mention `[requires]` or `formula_compiler` once `formula_v2` guidance appears.
- The design does not pin whether JSON formula examples will ship with Phase 2 or whether JSON formula loading will be retired before the feature lands.
- Future-axis placement under `[requires]` is implied by the `state_store` example but not explicitly frozen for docs authors.

**Required changes:**
- Move `engdocs/proposals/formula-migration.md` supersession into the Phase 2 docs/example/generated-help bundle, or require a same-PR supersession banner before the first user-visible requirement diagnostic ships.
- Extend `docs/release/formula-compiler-stale-guidance.yaml` with a positive-content check: architecture/proposal docs that discuss `[daemon] formula_v2` or `formula_v2 = true` must also reference `[requires]` or `formula_compiler`.
- Add a "Common confusion" subsection or callout to the reference rewrite that compares formula `[requires]` as compiler capability, `[pack].requires_gc` as binary floor, and lockfile/resolver revisions as artifact identity.
- Pin the JSON formula loader path in the rollout: either include JSON `requires` examples and parity fixtures in Phase 2, or schedule JSON formula retirement and remove JSON rows from the active matrix.
