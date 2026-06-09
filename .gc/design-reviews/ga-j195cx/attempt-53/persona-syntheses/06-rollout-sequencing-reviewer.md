# Lena Driscoll

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] The rollout relies on first-party graph formulas staying compatible through the alias window, but it does not name the phase that writes `[pack] requires_gc` for built-in and example packs before resolver/import enforcement lands. Both reviews treat the minimum binary floor and first-party source state as gating evidence; Claude concludes the missing pack-floor write can either make Phase 6 reject nothing or break existing consumers.
- [Blocker] Superseding the older formula-migration proposal is not sequenced early enough. Both reviews identify the stale `GC_NATIVE_FORMULA=false` rollback guidance as unresolved; Claude treats the live contradiction as blocking because operators could follow the old runtime rollback path during Phases 2-3.
- [Major] Phase 3 caller migration needs explicit ordering and rollback controls. Codex asks every Phase 3 caller task to name its rollback control, owner, and proof test; Claude adds that 3a must precede the other migrations, API projection must follow sling/orders, dashboard must follow API, and convergence must not drift from sling identity rules.
- [Major] Phase 4 is framed as a sequential phase but is really a cross-phase invariant: first-party graph formulas must remain dual-declared while any supported validation probe still reads only `contract`. Leaving it numbered as a phase obscures what PR scope should land and when the invariant may end.
- [Major] Alias-window exit criteria are measurable but not schedulable. The design requires two completed minor releases and 60 calendar days, but it does not state release cadence, current version anchor, fast-track rules, or when the clock starts.
- [Major] Phase 7 rollback semantics are underspecified. The design does not say whether reverting first-party requires-only conversion restarts the two-minor-release clock, nor does it require per-pack commit or PR granularity so a single first-party pack regression can be reverted narrowly.
- [Minor] Fixture coverage has a known gap. Claude identifies `cmd/gc/testdata/formulas/ralph-demo.toml` and `cmd/gc/testdata/formulas/ralph-retry-demo.toml` as legacy-path fixtures not covered by the cited stale-guidance globs.
- [Minor] Active legacy workflow roots are not accounted for at Phase 7. The convergence predicate preserves old `gc.formula_contract` metadata, but the rollout does not say whether active roots must drain, be counted by a report, or be explicitly waived.

**Disagreements:**
- Claude's verdict is `block`; Codex's verdict is `approve`. Assessment: the persona verdict is `block` because the unresolved items are specifically about rollout sequencing, rollback, and compatibility gates, which are this persona's review mandate.
- Codex views the rollout as deliberately and sufficiently staged, with no design change required before approval. Claude finds the staging incomplete because pack-floor writes, proposal supersedure, sub-phase ordering, clock rules, and Phase 7 rollback semantics are not assigned to enforceable phases. Assessment: prefer Claude's stricter read; the design's intended gates need to be made operational before implementation planning.
- Codex treats missing first-party inventory, minimum floor values, compatibility YAML, and external support inventory as deferred release artifacts. Claude treats some of those artifacts as missing phase prerequisites that can permanently block or invert later phases. Assessment: the artifacts may remain generated later, but their owning phase and blocking transition must be explicit now.
- Codex identifies Phase 3 rollback controls as a minor decomposition concern. Claude finds concrete ordering and durable-write skew risks if API, sling, orders, and convergence migrate out of sequence. Assessment: require the sequencing rules in the design so decomposition cannot schedule incompatible caller migrations.

**Missing evidence:**
- No Gemini review was present for this persona.
- No phase assigns concrete values for `first_gc_with_requires`, `first_gc_with_canonical_root_metadata`, or `first_gc_without_bd_runtime_fallback` in `formula-compiler-min-floor.json`.
- No phase explicitly assigns the release captain before the phases that depend on release-captain approval and floor transitions.
- No phase states when the two-minor-release clock starts: after canonical metadata writers, after docs/examples ship, or after both.
- No first-party pack inventory is included for graph-formula packs that need `[pack] requires_gc`, dual declaration, or later requires-only conversion.
- No rule states whether external packs without `[pack] requires_gc` but with legacy `contract = "graph.v2"` are accepted or rejected by Phase 6 resolver enforcement.
- No active-root drain criterion, report, or waiver is named for existing roots that carry only legacy formula-contract metadata.
- No per-sub-phase ordering matrix, rollback switch, owner, or proof test is specified for Phase 3 caller migrations.
- No cadence or current-version anchor makes "two minor releases" predictable before the migration is already complete.

**Required changes:**
- Add an initial rollout step that supersedes `engdocs/proposals/formula-migration.md` before Phase 1 code lands, including removal or replacement of the `GC_NATIVE_FORMULA=false` runtime-rollback guidance.
- Add a named phase before resolver/import enforcement that writes `[pack] requires_gc` for every first-party pack containing graph formulas, with the floor tied to `formula-compiler-min-floor.json`.
- Restate Phase 4 as an invariant that holds across Phases 2-7 rather than as a discrete sequential phase.
- Add explicit ordering constraints for Phase 3: 3a precedes 3b-3g; sling and orders precede API projection; dashboard follows API; convergence/fanout/molecule migrations name their allowed parallelism and identity compatibility rules.
- For each Phase 3 caller migration, require a rollback control, owner, release artifact, and proof test.
- Pick and document the Phase 7 rollback clock rule: either a rollback restarts the two-minor-release clock from the next dual-declared release, or it pauses without resetting.
- Require Phase 7 first-party source conversion to land in per-pack or otherwise narrowly revertible units.
- Pin the floor-file write to a specific phase and state when `first_party_requires_only_allowed` may become true.
- Add a release-cadence anchor or fast-track rule for the alias-window exit criteria.
- Add `cmd/gc/testdata/formulas/**/*.toml`, including the Ralph demo fixtures, to the fixture inventory or stale-guidance check.
- Add an active-root drain criterion, report, or explicit waiver for Phase 7.
