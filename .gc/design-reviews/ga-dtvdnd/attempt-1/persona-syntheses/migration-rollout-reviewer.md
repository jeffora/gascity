# Camille Okafor

**Persona verdict:** approve-with-risks

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Info] Existing-city migration is now treated as an operator-safe workflow, not silent fallback. All sources recognize `gc doctor --fix --non-interactive`, report-only default behavior, live-session preflight, TOML mutability checks, lock/cache provenance, backup/journal evidence, and post-fix resolved-config proof as the intended repair model.
- [Info] Two-repository rollout and public-pack pinning are materially stronger. Claude, Codex, and DeepSeek all identify AC14/AC15/AC16/AC17 as the release safety chain: validated public Gastown behavior manifest, immutable pins, version-skew matrix, cache provenance, offline fail-closed behavior, rollback/downgrade classifications, and recorded release-gate evidence.
- [Info] The no-silent-fallback posture is consistent for retired Maintenance, in-tree Gastown, stale system packs, synthetic aliases, offline resolution, and diamond/same-name tie breakers.
- [Major] The normal-command behavior for a present, declared retired import before repair is still unspecified. Claude flags this as the key upgrade-path product gap: an unrepaired city with a retired `city.toml`/`pack.toml` import could fail closed with a stable diagnostic, silently drop Gastown behavior, or fail later with an arbitrary downstream error unless the runtime contract is explicit.
- [Major] Old-binary write protection is reactive in the implementation discussion. DeepSeek warns that detecting old-binary writes after a migration marker can still permit stale binaries to write incompatible task-store or runtime state first. The rollout needs a proactive fail-closed guard such as a DB/schema version token or renamed active database at the repair commit point.
- [Major] Task-store migration is under-specified. DeepSeek reports that requirements mention inactive-bead policy, but the implementation migration table omits beads/tasks. Unresolved beads pointing at retired Maintenance commands, roles, or paths must be translated, archived, or blocked with diagnostics rather than left as latent legacy work.
- [Major] Stale legacy directories are a shadow-runtime hazard. Codex and Claude require stale paths to be ignored/reported; DeepSeek correctly notes that leaving `.gc/system/packs/maintenance` or `.gc/runtime/packs/maintenance` in place can still allow custom scripts, cron jobs, or local extensions to execute retired code by path.
- [Minor] Compatibility-pin windows are classified but not given a concrete closure trigger or post-activation inertness check. The temporary guarantee should be enforceable, not aspirational.
- [Minor] The legacy `internal/bootstrap/packs/core` compatibility-shim language is asymmetric with the stronger in-tree Gastown/Maintenance non-resolvability checks. Any shim must be bounded, non-runtime, and have a sunset or explicit owner.
- [Minor] Bundled-Core-vs-stale-locked-Core behavior across binary upgrades is not explicit. The likely desired policy is that the new bundled Core payload wins and stale locked/materialized Core is pruned or reported, never loaded as an alternate Core.
- [Minor] Offline repair sequencing is still unclear: `gc doctor --fix --non-interactive` may need to fetch and verify the public pack, or require a pre-seeded cache. Operators need a deterministic outcome.
- [Minor] Operator rollback is journal-backed but not yet clearly automated. DeepSeek recommends a single-command restore path rather than relying only on release-note/manual recovery.

**Disagreements:**
- Codex sees no unresolved rollout product question before requirements approval. Claude says the pre-repair retired-import runtime contract is still a product decision. My assessment: Claude's concern is decisive for existing-city rollout because it determines what a normal operator command does before repair.
- DeepSeek says compatibility-pin windows are scoped and temporary; Claude says the closure trigger is not named. My assessment: keep DeepSeek's intended model, but require an explicit expiry trigger and inertness proof.
- Codex treats TOML preservation, journal semantics, and live-state refusal as design gates already named by requirements. DeepSeek pushes for proactive DB fail-closed behavior and automated rollback. My assessment: the requirements are strong, but old-binary write prevention and rollback mechanics are necessary rollout evidence, not optional polish.
- Claude focuses on requirements approval, while DeepSeek evaluates the implementation plan and flags cross-document omissions. My assessment: implementation-plan omissions matter here when the requirements already promise inactive-bead policy, stale-path safety, and old-binary reconciliation.

**Missing evidence:**
- The runtime-resolution contract for a present, declared retired optional import encountered by a normal behavior-changing command before `gc doctor --fix` has run.
- A version-skew-matrix closure trigger for compatibility-pin windows and a post-activation check proving in-tree or compatibility sources are inert.
- A bundled-Core-vs-stale-locked-Core precedence rule for binary upgrades, plus a runtime non-resolvability check for legacy Core roots.
- Offline repair sequencing: whether repair fetches and verifies the pinned public Gastown pack or requires a seeded cache first.
- A concrete task-store/bead migration policy for inactive or unresolved beads that reference retired Maintenance paths, commands, roles, or pack-owned behavior.
- A proactive old-binary write guard that prevents stale binaries from opening and mutating migrated state.
- A stale-directory isolation policy that preserves operator edits without leaving legacy paths executable at their old locations.
- Automated rollback/restore behavior using the repair journal and backups.
- Downstream evidence: upgrade matrix tests, repair goldens, version-skew matrix, public pin ledger, offline/cache tests, rollback guidance, and recorded release-gate results.

**Required changes:**
- State the normal-runtime contract for present, declared retired imports in an unrepaired city. Recommended outcome: fail closed before behavior-changing operations with a stable condition code, exact source attribution, and a pointer to `gc doctor --fix`; explicitly distinguish this from a never-declared absent optional pack.
- Tighten AC15 so the version-skew matrix defines the compatibility-pin window closure/expiry trigger and proves compatibility or in-tree paths are inert after activation.
- Add symmetric runtime non-resolvability checks for the legacy Core root and bound any compatibility shim as non-runtime diagnostic/fixture behavior with a sunset or owner.
- State bundled-Core-vs-stale-locked-Core precedence at binary-upgrade time; the bundled Core payload should not be shadowed by stale lock/cache/materialized state.
- Define offline repair sequencing for air-gapped operators, including whether `gc doctor --fix` fetches+verifies the public pack or requires pre-seeded cache input.
- Add explicit task-store/bead migration steps for inactive unresolved beads that reference retired behavior: translate through the asset ledger, archive, or mark blocked with diagnostics.
- Add a proactive old-binary write guard at the migration commit point, such as a state DB version change or database path transition that old binaries fail to open.
- Rename or otherwise isolate stale legacy pack directories during repair rather than retaining them at executable legacy paths.
- Specify an automated rollback/restore command or an equivalent non-interactive recovery path that replays the repair journal and restores staged backups atomically.
