# Faisal Khoury

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Info] The diagnostic direction is strong. Claude and Codex agree that missing-Core, duplicate-Core, retired-path, cache-integrity, version-skew, and public-pack conditions must use stable condition codes, exact source or nested import-chain attribution, stdout/stderr separation, and a pack-independent registry that can render in bootstrap-only mode with no packs resolved.
- [Info] The repair surface is appropriately bounded. All sources recognize `gc doctor --fix --non-interactive` as the only mutating surface, report-only by default, idempotent, atomic, resumable, post-verified, non-destructive to unrelated TOML, and guarded by durable preflight, journal/backup, live-state evidence, or refusal semantics.
- [Info] Operator messaging correctly distinguishes required Core, optional external Gastown, and retired Maintenance across doctor, import-state, help/docs, and public-pack wording.
- [Major] The implementation must prove the condition-code registry works before normal pack resolution. Codex and DeepSeek both flag the bootstrap catch: if condition rendering or source attribution depends on resolved packs, the most important diagnostics fail exactly when Core is missing or corrupt.
- [Major] Recursive import-chain tracing is not yet designed tightly enough. DeepSeek notes that flattened config loading can lose the chain needed to report `city.toml -> pack-a -> pack-b -> retired path`; AC11 requires that exact source chain when available.
- [Major] The AC2 dev/test Core-less escape hatch is a safety risk unless explicitly implemented and bounded. DeepSeek flags the implementation-plan gap: if ignored, partial-config tests fail; if loose, production can bypass required Core validation.
- [Major] Inactive bead/task-store migration belongs in diagnostics and repair. DeepSeek reports that AC10/AC11 promise inactive-bead policy, but the implementation plan omits task beads; unresolved beads pointing at retired Maintenance paths or roles must be translated, archived, or blocked with diagnostics.
- [Major] Old-binary write protection is currently reactive. Detecting stale writes after the fact can allow old binaries to mutate migrated task-store or runtime state before the new binary reports version skew.
- [Major] Stale legacy directories must not remain executable at their old paths. Ignoring them in active discovery is not enough if custom scripts or external automation can still run `.gc/system/packs/maintenance` or `.gc/runtime/packs/maintenance` by hardcoded path.
- [Major] Air-gapped cache seeding needs a validated operator path. DeepSeek flags that offline/cache semantics are incomplete without a non-interactive way to seed a public Gastown cache while preserving digest, manifest, and provenance validation.
- [Minor] The missing-Core `--fix` mutation target is under-defined. Claude separates safe host-system re-materialization from unsafe edits to user/root/rig/transitive config, but the requirements do not yet make that distinction explicit.
- [Minor] Multi-condition aggregation and partial-fix outcomes need explicit exit-code and machine-readable reporting. Upgrade cases commonly contain missing Core, retired paths, and version skew at once.
- [Minor] Cross-command consistency is required in prose but needs an equality check proving `doctor` and `import-state` emit the same condition code, severity, and source attribution for the same broken city.

**Disagreements:**
- Codex says no product unknown remains before requirements approval. Claude says missing-Core repair semantics, aggregation, and cross-command equality still need sharpening. My assessment: these are narrow but real diagnostics contract gaps, so the lane should approve only with risks.
- DeepSeek escalates implementation-plan omissions such as task-store migration, stale-directory isolation, old-binary DB guards, and cache seeding to Major. Claude focuses on the requirements artifact and rates its carryovers Minor. My assessment: these become Major for this lane once the requirements promise safe repair and exact diagnostics but the implementation evidence omits the state that can still break repair.
- DeepSeek suggests a specific `GC_TEST_ESCAPE_HATCH=1` path for the dev/test escape hatch. My assessment: the exact mechanism can differ, but the requirements/design must prove the hatch is test-only and unreachable from production runtime behavior.

**Missing evidence:**
- The concrete mutation target for missing-Core `--fix`: re-materialize a corrupt/absent host system payload, refuse/report config omissions, or another explicit split.
- A recursive dependency tracer or diagnostic model that preserves nested import chains through config loading.
- A production-boundary proof for the AC2 dev/test Core-less escape hatch.
- A single-run multi-condition reporting contract and partial-fix output contract, including exit codes and machine-readable remaining actions.
- Cross-command goldens proving doctor and import-state emit identical condition code, severity, and source attribution for the same broken city.
- The actual `migration-diagnostics.schema.json` or equivalent binding diagnostic schema and goldens.
- Bead/task-store migration policy for inactive unresolved work that references retired paths, commands, roles, or pack-owned behavior.
- A proactive old-binary write guard, such as a schema version token or database path transition.
- Stale-directory repair behavior that preserves operator edits without leaving retired paths executable.
- A validated air-gapped cache-seeding command or equivalent non-interactive workflow with digest and provenance checks.

**Required changes:**
- Split missing-Core repair semantics: absent/corrupt host-materialized Core system payload may be re-materialized idempotently from the embedded release payload; user/root/rig/transitive config that omits or shadows Core must be report-only or refused, never auto-edited.
- Require doctor and import-state to report all detected conditions in one non-interactive run and require `--fix` to distinguish fully repaired, partially repaired with enumerated manual steps, and nothing fixable through the exit-code matrix.
- Add a cross-command consistency check to AC3/AC11: the same broken city must produce the same condition code, severity, and source attribution from doctor and import-state.
- Design recursive import-chain tracing so diagnostics can name nested retired imports after config loading without relying on flattened state alone.
- Define and test the AC2 dev/test escape hatch as strictly test-only and unreachable from production `gc` runtime, controller, session, dispatch, formula expansion, doctor repair, and city-state mutation.
- Add task-store/bead migration diagnostics and repair policy for inactive unresolved beads that reference retired Maintenance paths or roles.
- Add proactive old-binary write prevention at the migration commit point.
- Rename or otherwise isolate stale legacy pack directories during repair instead of retaining them at executable legacy paths.
- Provide a validated air-gapped cache-seeding path that verifies archive digest, behavior manifest, and provenance before promotion into `.gc/cache`.
