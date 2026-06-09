# Owen Gallagher

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash (provided as `06-pack-boundary-containment_gemini.md`)

**Consensus findings:**
- [Blocker] The zero-duplicate-active gate depends on an undefined behavior identity model. Claude, Codex, and DeepSeek all require a stable cross-source identity for formulas, orders, prompts, prompt fragments, scripts, hooks, overlays, commands, doctors, skills, and generated references. The plan cannot prove duplicate containment if moved, renamed, split, or metadata-owned behavior can avoid collision under ordinary precedence rules.
- [Blocker] `internal/packsource` is named as the right authority, but the plan does not yet make it the producer of active roots before config/runtime staging. Claude focuses on the missing `systempacks` to `packsource` ordering before `config.Load*`; Codex focuses on existing active fields such as `FormulaLayers`, `PackDirs`, overlays, commands, doctors, skills, globals, and generated formula views; DeepSeek asks for operation-specific state rules. The shared requirement is that packsource must feed active-root enumeration, not only audit paths after they have become behavior.
- [Blocker] Active retired-source containment is not proven across every behavior surface. Reviewers agree stale `.gc/system/packs/{maintenance,gastown}`, `.gc/runtime/packs/maintenance`, synthetic caches, and public-pack transitions must be tested against all active root kinds, including formulas, orders, prompts, prompt fragments, scripts, hook overlays, commands, doctors, skills, MCP/materialized catalogs, generated references, lock/cache records, docs/reference generation, and `.beads/formulas` symlinks.
- [Major] The raw pack-root enumeration ban is too abstract without a typed active-root API and migration inventory. Claude identifies current discovery sites that take opaque `dir` or `fs.FS` parameters, while Codex requires direct construction from import paths, cache paths, `.gc/system/packs`, and `examples/gastown` to fail unless explicitly allowlisted as non-behavior. A scanner alone will not prevent bypasses unless pack-root handles make raw enumeration visibly wrong.
- [Major] Compatibility, activation, and rollback duplicate policy is incomplete. Claude flags overlap while both in-tree and public sources exist; DeepSeek makes the compatibility-window and old-binary-plus-activation-pin deadlocks explicit; Codex requires the public Gastown/no-Maintenance proof across every active root kind. The plan needs one matrix that defines pin contents, gate timing, duplicate semantics, downgrade limits, and recovery behavior.
- [Major] Stale generated views are not adequately separated from stale pack directories. Codex specifically calls out `.beads/formulas` symlinks, generated command/help references, materialized catalogs, runtime script paths, lock/cache records, and staged formula/order views. These must be diagnostic or cleanup inputs only, never active behavior from retired roots.
- [Major] Doctor and repair flows can deadlock if the normal loader fail-closes before diagnostics can inspect duplicate-active or retired-source state. DeepSeek raises this as a blocker; the concern is consistent with Claude and Codex's requirement that doctor, cache, lock, rollback, and runtime-state migration participate in the same classifier contract.
- [Major] In-flight molecule/session behavior is unresolved. DeepSeek asks whether step definitions, prompt fragments, scripts, and pack metadata are snapshotted into beads or read live from pack directories; Codex similarly asks for coverage of previously staged formula/order/script views. Without this policy, mid-migration slices can break work that was already composed under legacy paths.
- [Major] Custom/fork classification for stale retired directories needs behavior-bearing provenance. Claude and Codex require stale roots to stay ignored unless deliberately classified; DeepSeek identifies false promotion from incidental files. The invariant should be based on registered behavior-bearing content, not directory dirtiness from logs, editor files, or OS metadata.
- [Minor] The implementation-plan artifact should not retain inline review provenance comments. Claude alone raised this, but it is still a cleanup item for schema/document hygiene.

**Disagreements:**
- Claude and Codex return `approve-with-risks`; DeepSeek returns `block`. Assessment: `block` is the correct persona verdict because behavior identity, packsource active-root authority, duplicate timing, and diagnostic recovery are prerequisites for safely decomposing source-moving slices.
- Claude treats the current `packsource` direction as mostly sound but under-specified; Codex says it must be elevated from classifier to active-root graph producer; DeepSeek asks for a state-by-operation matrix. Assessment: these are compatible requirements, not conflicting designs. The plan should define the API, caller mapping, and operation matrix together.
- DeepSeek proposes compatibility-pin duplicate exclusion or a report-only duplicate gate during Slices 2-4, while Claude/Codex emphasize a hard zero-duplicate-active invariant before execution. Assessment: report-only is acceptable only for non-executing diagnostics or if active execution is proven duplicate-free by construction.
- DeepSeek prescribes SHA-256 behavior-file digests for custom/fork detection. Claude and Codex ask for the provenance property without mandating the exact mechanism. Assessment: require behavior-bearing content provenance; SHA-256 over registered behavior files is a strong acceptable implementation.
- Claude uniquely asks whether behavior identity should bind to AC6/AC7 stable row ids. Codex and DeepSeek ask for the schema generically. Assessment: the plan should either use AC6/AC7 rows as the canonical source or explicitly define an equivalent stable manifest-backed identity.

**Missing evidence:**
- The concrete `internal/packsource` API, source-state enum, typed active-root kinds, and owner package.
- A mapping from existing runtime/config fields and discovery outputs to packsource root kinds and classifier states.
- The exact timing of retired-source classification and zero-duplicate-active checks relative to `internal/systempacks`, `config.Load*`, formula/order discovery, prompt lookup, script resolution, hook enumeration, and behavior execution.
- A stable behavior-id schema for formulas, orders, prompts, prompt fragments, scripts, hooks, overlays, commands, doctors, skills, MCP/materialized catalogs, and generated references.
- The source of behavior ids: AC6 stable asset rows, AC7 manifest rows, pack manifests, generated ledgers, formula/order metadata, or a defined combination.
- A state-by-operation matrix for active discovery, config load, runtime staging, install, cache read/promotion, lock refresh/write, doctor diagnostics, doctor mutation, rollback recovery, docs/reference generation, and public-source normalization.
- Compatibility-pin versus activation-pin content rules while embedded Maintenance remains required.
- The rollout row for old binary plus activation pin, including whether activation is a one-way downgrade boundary and what manual recovery path exists.
- Non-executing diagnostic/preflight load semantics for duplicate-active or retired-source failures.
- The policy for already composed molecules, sessions, staged script paths, and pending orders when source paths are retired mid-run.
- The custom/fork provenance rule for stale retired directories, including treatment of non-behavioral temporary files.
- Surface-specific negative tests proving stale retired roots and synthetic caches cannot influence active behavior.
- Whether required file-set validation consults packsource or is deliberately outside the classifier boundary.

**Required changes:**
- Define `internal/packsource` as the source of truth for active root enumeration, not only classification. Its API should return typed active roots with source state, source identity, pack identity, binding, digest/provenance, and allowed operation.
- Require config expansion and runtime staging to consume packsource typed roots when populating formula layers, pack dirs, overlays, commands, doctors, skills, globals, hooks, scripts, generated views, cache records, and lock records.
- Specify that retired-source classification and duplicate candidate construction run before normal config/runtime behavior can absorb retired roots; if any diagnostic path bypasses the execution gate, name it explicitly.
- Define stable behavior ids for every behavior kind and make the zero-duplicate-active gate compare those ids before ordinary precedence or last-wins resolution can choose a winner.
- Bind behavior ids to AC6/AC7 manifest evidence or specify an equivalent manifest-backed schema that survives moves, renames, splits, compatibility pins, activation pins, stale generated roots, and old/new binary views.
- Add a packsource adoption inventory and state-by-operation matrix covering config load, runtime load, install, cache read/promotion, lock refresh/write, doctor report/fix, rollback, docs/reference generation, formula/order discovery, prompt/template discovery, hook overlay discovery, script resolution, command/doctor/skill discovery, public-source normalization, and generated-reference linting.
- Resolve compatibility-window duplication by specifying compatibility-pin contents, activation-pin contents, duplicate policy, gate timing, and whether each window is duplicate-free by construction, isolated/report-only, or hard-fail.
- Add the old-binary plus activation-pin rollout row and make activation-pin adoption a documented downgrade boundary unless a supported old-binary recovery path is defined.
- Define a non-executing diagnostic/doctor preflight load mode, or equivalent repair path, that can inspect duplicate-active and retired-source failures without activating behavior.
- Add stale generated-view containment to rollout and tests: `.beads/formulas` symlinks, generated references, materialized catalogs, lock entries, cache entries, and script targets pointing at retired roots must be diagnostic or cleanup targets only.
- Make raw pack-root enumeration unrepresentable through typed root handles, then define scanner/lint rules and allowlist ownership/expiry for any remaining raw filesystem traversal over pack roots.
- Define custom/fork detection from registered behavior-bearing content, such as SHA-256 digests against pristine behavior files, while ignoring incidental non-behavioral files.
- Specify in-flight molecule/session handling for Slices 2-7, including whether definitions and scripts are snapshotted into beads at composition time and what operators see for retired-source work.
- Remove inline review provenance comments from `implementation-plan.md` and keep review annotations in workflow artifacts.
