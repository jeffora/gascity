# Lena Driscoll

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex

**Consensus findings:**
- [Major] Phase 4 rollout needs a sharper implementation contract before durable producers migrate. Claude flags Phase 4a as a large bundle whose atomicity is unclear; Codex generalizes the same risk across Phase 4b-4g, where rollback is described as an outcome but not tied to named switches, wrapper boundaries, revert units, or zero-write tests. The rollout is directionally sound, but contributors still lack an executable rule for how each producer can land incrementally and roll back before durable writes.
- [Major] Phase 3 first-party pack-floor changes need stronger compatibility and rollback evidence. Both reviews find ambiguity around updating `[pack] requires_gc` while preserving old-reader support through dual-declared formulas. The design should explicitly say Phase 3 may land pack-by-pack because enforcement is deferred to Phase 7, prove that supported old binaries read the exact Phase 3 source shape, and clarify that dual declarations, pack floors, and source restoration are independently controllable during rollback.
- [Major] Packman provenance is a hard sequencing dependency that is not first-class in the phase table. Codex identifies the missing PR home, owner, saved artifacts, and rollback unit for the schema/provenance contract needed by resolver/import enforcement and alias-removal gates. Claude's Phase 7 cache/lockfile rollback concern depends on the same boundary: the rollout must define how resolver state recorded during enforcement is evicted, re-evaluated, or operationally cleared.
- [Major] The documented migration proposal supersession is not pinned tightly enough to a phase. Claude notes that `engdocs/proposals/formula-migration.md` is identified as superseded but not explicitly listed in the Phase 2 docs bundle that introduces user-visible `[requires]` diagnostics. Leaving the old proposal current would preserve guidance the design intends to retire.
- [Minor] Phase 6 cleanup needs a crisp boundary from predecessor documentation. Claude and Codex both endorse the diagnostic visibility matrix as the right sequencing control, but the rollout should state plainly that Phase 6 covers only cleanup with no diagnostic producer waiting on it; required docs, help, schema, OpenAPI, dashboard, examples, generated clients, and stale-guidance checks land in the producer's own phase.
- [Minor] External-pack migration progress needs an observable signal before alias removal. The alias-removal gate has strong first-party criteria, but the external support row can move from active to expired without a concrete progress measure tied to pinned external revisions, lockfile evidence, or a maintained inventory.

**Disagreements:**
- There is no verdict disagreement: both Claude and Codex return `approve-with-risks`. I keep that persona verdict because the risks are material but all are framed as sequencing, evidence, or rollback-contract gaps rather than proof that the rollout cannot work.
- Claude emphasizes release sequencing details that Codex does not mention: Phase 4a atomicity, Phase 3 incremental batching, Phase 6 cleanup wording, `formula-migration.md` supersession, external-pack progress, release cadence, migration tooling, in-flight molecule rollback, and predecessor artifact rollback. Codex does not contradict these; I retain the Phase 4a, Phase 3, docs, Phase 7, and external-progress items as required or consensus-adjacent because they directly affect rollout safety.
- Codex uniquely elevates packman schema/provenance and old-reader proof for `[pack] requires_gc` as major risks. Claude's Phase 7 resolver-state and Phase 3 batching findings make the same dependency visible from an operations angle, so I treat Codex's packman and probe-corpus changes as required rather than optional.
- Claude asks whether Phase 5, Phase 6, and Phase 7 can overlap with Phase 4 sub-phases. Codex does not raise this. I assess it as a missing evidence/planning question rather than a required change unless the design wants parallel landing to be officially supported.

**Missing evidence:**
- No Kimi 2.6 artifact was present for this persona.
- No explicit Phase 4a landing rule says whether the bundle must be one PR or can be split while the no-new-raw-consumer guard remains report-only until the final PR.
- No saved old-reader/probe artifact proves a supported old Gas City binary can read a first-party pack containing `[pack] requires_gc` plus dual-declared graph formulas.
- No first-class packman schema/provenance readiness artifact is sequenced before resolver/import enforcement.
- No per-producer rollback-control table names the disabling mechanism, owner, and zero-write proof command for Phase 4b-4g durable producers.
- No Phase 7 rollback rule describes cached resolver rejections or lockfile state recorded while enforcement was active.
- No release-cadence reference defines how to plan against "two completed minor releases" plus 60 calendar days.
- No external-pack migration tool or measurable external migration-progress signal is specified.
- No explicit rule explains how already-migrated or artifact-stamped active roots behave when a Phase 4 producer sub-phase is rolled back.

**Required changes:**
- Add a Phase 4a atomicity statement: either the whole bundle lands as one PR, or it may be split only while the blocking no-new-raw-consumer guard remains report-mode until the final bundle PR.
- Add a Phase 4 rollback-control table for 4b-4g that names the concrete switch, wrapper boundary, or revert unit; the rollback owner; and a zero-write proof command for each migrated producer surface.
- Add a Phase 3 compatibility and rollback statement that `[pack] requires_gc` floor declarations can land incrementally pack-by-pack because enforcement waits until Phase 7, and that pack floors, dual declarations, and source restoration can be rolled back independently.
- Extend the old-reader/probe corpus to cover the exact Phase 3 first-party source shape: pack metadata with `requires_gc`, dual formula declarations, current built-in graph syntax, and the supported old Gas City versions or SHAs.
- Promote packman schema/provenance readiness to an explicit prerequisite phase or split Phase 7 into separate "packman provenance contract" and "pack-floor enforcement" phases, each with PR home, owner, required command, saved artifacts, and rollback unit.
- Pin `engdocs/proposals/formula-migration.md` supersession to the Phase 2 docs bundle or another named phase PR that lands before any user-visible `[requires]` diagnostics.
- Specify Phase 7 rollback behavior for resolver cache and lockfile state recorded during enforcement.
- Add one sentence at the start of Phase 6 stating that Phase 6 has no producer that exposes a new diagnostic surface; all predecessor docs, schemas, generated clients, dashboard fixtures, examples, and stale-guidance checks land in the producer's own phase.
- Add a measurable external-pack migration signal to the alias-removal gate, such as inventory rows tied to pinned external revisions or lockfile reports that show external packs have moved to `requires_only` before support status becomes `expired`.
