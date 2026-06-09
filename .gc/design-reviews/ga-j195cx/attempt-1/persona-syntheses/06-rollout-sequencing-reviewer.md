# Lena Driscoll

**Persona verdict:** block

**Sources:** Claude, Codex

**Consensus findings:**
- [Blocker] Phase 3 can ship first-party pack floor declarations without the old-reader and exact-floor proof the design requires. Codex flags that the phase table's required command is a legacy contract inventory, not an old-reader pack-load or minimum-floor gate. Claude independently flags that no release gate proves a first-party `[pack] requires_gc` floor is satisfied by the binary that ships it. This can turn an apparently dormant floor edit into a mixed-version break or a self-rejecting built-in pack.
- [Major] First-party metadata additions are not consistently gated before source changes land. Claude calls out that Phase 2's first-party dual declarations are not explicitly blocked on the `dual-declared-graph-v2` compatibility corpus passing on the latest two published Gas City minor releases, current `main`, and the `bd` probe. Codex makes the same rollout-ordering point for Phase 3 pack floors: the merge gate must prove older supported readers and active probes tolerate the new metadata before first-party packs carry it.
- [Major] Phase 4 rollback controls are directionally described but not operationally pinned. Both reviewers ask for concrete per-producer controls for `gc sling`, API sling, orders, convergence, molecule execution, and fanout. The plan needs to name how each migrated producer is disabled before protected writes, what continues during disabled mode, and which zero-write fixture proves rollback behavior.
- [Major] Rollback and release-clock consequences for Phase 8 are under-emphasized. Claude notes that a Phase 8 rollback after requires-only source publication resets the alias-removal window, effectively delaying removal by at least two completed minor releases plus 60 days from a later dual-republishing release. That cost is currently buried outside the rollback notes that a release captain would scan.
- [Minor] Phase 4 and Phase 6 documentation ownership is still prose rather than a checked manifest. Claude identifies overlap between diagnostic-surface docs owned by Phase 4 sub-phases and broader stale-doc cleanup in Phase 6. Rollback of a Phase 4 surface must not leave Phase 6 docs advertising a disabled or not-yet-shipped surface.
- [Minor] Parallel Phase 4 producer sub-phases rely on non-overlapping durable write boundaries without a named assertion. Claude points to the caller manifest as the right source, but the rollout plan does not name a CI test that prevents two parallel sub-phases from owning the same `durable_boundaries[]` entry.

**Disagreements:**
- Claude's overall verdict is `approve-with-risks`; Codex's verdict is `block`. I assess the persona verdict as `block` because Phase 3's required local command is narrower than the prerequisite it is supposed to prove, and that gap can ship a release artifact that supported older binaries reject or misread.
- Claude spreads the compatibility concern across Phase 2 dual declarations, Phase 3 floor values, and later Phase 8/7b enforcement. Codex focuses the blocking concern on Phase 3's executable gate. These are compatible assessments: each first-party metadata introduction needs a gate before the corresponding source edit lands.
- Claude requests additional rollout polish around Phase 8 rollback cost, Phase 4/6 docs partitioning, parallel-boundary CI, proposal supersession, and mixed-version reporting. Codex does not object to those areas, but treats the Phase 3 gate and Phase 4 rollback controls as the minimum blockers.
- Codex frames Phase 4 rollback as missing named controls and command-level evidence. Claude extends that to in-flight roots created while a producer was enabled and then disabled. The required fix should cover both disabled-mode behavior and readability/cleanup of roots already written.

**Missing evidence:**
- No Kimi 2.6 artifact was present for this persona.
- No Phase 3 artifact combines first-party floor edits, old-reader pack-load corpus results, exact minimum-floor comparator output, release-floor JSON validation, and placeholder-floor rejection.
- No gate compares `minimum_gc_for_requires_only` or first-party `[pack] requires_gc` values to the version of the release publishing those floors.
- No explicit Phase 2 ship gate ties the `dual-declared-graph-v2` compatibility fixture to first-party dual-declaration merges before `[requires]` appears in bundled packs, examples, tutorials, or fixtures.
- No per-producer Phase 4 rollback artifact names the control, owner, disabled-mode behavior, command/config surface, zero-write assertion, and behavior for roots created before rollback.
- No rollback-replay fixture proves that roots created by a migrated Phase 4 producer remain readable and cleanable after that producer surface is disabled.
- No manifest assigns each doc/help/schema/OpenAPI/dashboard/example surface to exactly one Phase 4 sub-phase or Phase 6 cleanup owner.
- No CI assertion proves concurrently active Phase 4 sub-phases have disjoint `durable_boundaries[]` ownership.
- No checked consequence proves `engdocs/proposals/formula-migration.md` is either aligned with the canonical design or marked obsolete before Phase 2 ships.
- No mixed-version fleet evidence explains how operators aggregate active-root reports before all controllers include the shared workflow-root predicate.

**Required changes:**
- Replace or extend the Phase 3 required local command with an explicit first-party floor-declaration gate, such as `gc formula validate --all-packs --first-party-floor-declaration-gate --json`. It must fail closed on placeholder floors, missing old-reader binaries, stale corpus output, unsupported-reader failures, and any floor greater than the publishing release version.
- State that Phase 3 cannot be approved from formula alias inventory alone. The release checklist must cite the old-reader pack-load fixture artifact required for first-party `[pack] requires_gc` floors.
- Add a Phase 2 ship gate requiring `gc formula validate --compat-corpus internal/formula/testdata/compat_corpus --json` to prove `dual-declared-graph-v2` is accepted by every supported reader before any first-party `[requires]` table is merged.
- Add a Phase 4 rollback playbook table for 4a-4g naming the disabling control, owner, disabled-mode behavior, producers that continue dual-stamping, roots that remain readable through the shared predicate, and the zero-write or rollback-replay fixture for each sub-phase.
- Add `TestNoOverlappingProducerWriteBoundariesInParallelSubPhases` or an equivalent CI assertion over the caller manifest, and name it in the Phase 4 lockdown table.
- Add an operator-facing Phase 8 rollback-cost callout to the rollback notes, including the alias-window clock reset and the minimum delay of two completed minor releases plus 60 calendar days from the next dual-republishing release.
- Convert the Phase 4/Phase 6 documentation partition into a manifest where every doc/help/schema/OpenAPI/dashboard/example artifact has exactly one phase owner, and rollback of a Phase 4 sub-phase reverts its owned docs with the surface.
- Gate Phase 2 on `engdocs/proposals/formula-migration.md` being either updated to match the canonical design or marked obsolete and superseded by `engdocs/design/formula-compiler-requirements.md`.
