# Yelena Markovic - Codex

**Verdict:** block

**Top strengths:**
- The plan correctly makes runtime-state migration doctor-owned, not a controller-startup side effect. Startup, API handlers, and reload paths may diagnose or refuse, but only `gc doctor --fix` mutates state.
- The mutation coordinator shape is appropriate: shared city advisory lock, live-controller refusal from runtime observation, staged writes, durable recovery state, digest rechecks before rename, and a single commit point.
- The design names the important state classes: JSONL archive state, refs/remotes, pending archive push state, escalation fields, spawn-storm ledgers, explicit formula environment state, and order skip/tracking aliases.

**Critical risks:**
- [Blocker] The migration table is not concrete enough for decomposition or data-safety review. It uses placeholders such as `legacy archive directories`, `Core archive directory`, `pending push state`, and `old throttle ledger` instead of exact current paths, exact new paths, and exact keys. The plan needs path-level rows for the JSONL state file, archive repo directory, archive refs/remotes, push cursor fields, spawn-storm ledger, escalation fields, formula env state, and order skip/tracking keys before an implementer can avoid missing or duplicating state.
- [Blocker] Old-binary post-marker write detection is asserted but not specified. The plan says retained legacy state is ignored unless marker or digest checks show conflict, but it does not define which legacy files are rechecked, how appended JSONL records are detected, how changed Git refs/remotes are compared, how push cursor divergence is reconciled, or what makes a deterministic re-upgrade safe.
- [Major] The interruption and duplicate-writer test matrix is too high-level. `failure injection after each staged publish step` and `detects old-binary post-marker writes` are useful labels, but the design needs explicit fixtures for half-copied archive repos, in-flight archive push state, duplicate `doctor --fix` processes, old binary writes after marker, rollback to older binary, and re-upgrade.

**Missing evidence:**
- No `doctor-fix-inventory.yaml`, runtime-state fixture table, migration diagnostics schema, or version-skew matrix exists in `plans/core-gastown-pack-migration/support/` yet.
- The current implementation plan does not identify the source files or commands that own each legacy runtime-state path, so the inventory cannot yet prove it covers every writer.
- No marker schema is present. The plan names marker fields, but not required field types, digest scope, compatibility versioning, or conflict condition codes.

**Required changes:**
- Replace the generic runtime-state table with exact rows: old path/key, new Core-owned path/key, writer command or script, marker fields, copy/reconcile rule, conflict condition code, rollback behavior, and focused test for each migrated state class.
- Define the post-marker divergence algorithm per artifact. Include JSONL append watermarks or digests, Git refs/remotes comparison, push cursor precedence, spawn-storm read-union cutoff, order skip/tracking alias expiry, and formula env conflict handling.
- Add an explicit round-trip fixture matrix for interrupted archive copy, in-flight push state, duplicate doctor writers, old-binary writes after marker, downgrade to older binary, and deterministic re-upgrade.
- Tie runtime-state diagnostics to `migration-diagnostics.schema.json` so manual reconciliation and version-skew refusal are operator-visible and testable.

**Questions:**
- Which current files and scripts are the authoritative writers for JSONL archive state, spawn-storm ledgers, order skip/tracking state, and escalation fields?
- After a downgrade, is the old binary allowed to keep writing legacy runtime state, or is downgrade intended to be a read-only/manual-recovery state until re-upgrade?
