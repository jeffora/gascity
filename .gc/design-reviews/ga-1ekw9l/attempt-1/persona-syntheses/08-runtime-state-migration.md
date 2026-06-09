# Yelena Markovic

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Post-marker old-binary divergence detection is not mechanically specified. The design says retained legacy state is ignored unless marker or digest checks show conflict, but it does not define per-artifact legacy-source baselines, destination fingerprints, high-water marks, completed steps, or conflict policy for JSONL archives, refs/remotes, push cursors, spawn-storm ledgers, order skip/tracking state, escalation fields, and formula environment state.
- [Blocker] The state migration table is too abstract for deterministic implementation. Claude and Codex both require exact old and new paths, keys, refs, remotes, fields, readers, writers, merge rules, refusal rules, marker fields, proof commands, and expiry behavior instead of placeholders such as "legacy archive directories", "old throttle ledger", or "stable order keys".
- [Blocker] Archive copy, publish, and old-writer protection do not yet prove no data loss. The migration needs a quiesced copy protocol with live-controller/writer checks, process-unique staging, manifest validation, destination-absent first-copy rules, source and destination fingerprints, an immediate pre-marker re-snapshot, and atomic same-filesystem publish or a fail-closed fallback.
- [Blocker] Order skip/tracking aliases are not integrated with the current exact-label dispatch model. Codex verified that dispatch gating, tracking beads, and order history look up exact `order-run:<scoped>` labels. Renamed orders can double-fire unless alias lookup covers dispatch gating, last-run/cursor lookup, stale sweeps, and history until old entries are re-keyed or aliases are safely retired.
- [Major] Push-cursor and git remote reconciliation are ambiguous. "Newest verified cursor wins, conflicts block" is not a complete policy, and split legacy/Core archive repositories can race the same remote or duplicate/drop pending pushes unless cursor conflict semantics, ref conflict rules, push idempotence, and rollback behavior are explicit.
- [Major] Spawn-storm ledger migration lacks a dedup key and post-marker monitoring rule. A read-union can double-count copied events, while "Core-only after marker" can miss throttle events written by an old binary to the retained legacy ledger.
- [Major] Current-system grounding is incomplete for migrated runtime state. The plan must cite the actual Maintenance-pack orders and owning packages/scripts for JSONL export state, archive repos, spawn-storm counts, order tracking labels, push state, escalation markers, and provider/runtime state before decomposition.
- [Major] Downgrade and re-upgrade continuity is underspecified. The plan must state whether Core-owned runtime paths are readable by old binaries, whether downgrade after Core-only writes is manual-only, how old binaries are prevented from silently resuming stale retained state, and how post-marker legacy writes are merged or blocked during re-upgrade.
- [Major] Distributed runtime liveness is not covered. DeepSeek flags that local process-table and local lock checks do not detect a controller running in another Kubernetes pod or node; K8s runtime deployments need provider-aware controller detection, lease checks, or explicit unsupported-mode refusal.
- [Major] Divergence checks risk startup/reload performance problems if they hash large retained archives unconditionally. Marker checks need a tiered strategy such as path identity, size, and mtime first, with cryptographic digest fallback only when cheap metadata changes.
- [Minor] Live-controller refusal must handle old-binary controllers and indeterminate liveness, not only cooperating new controllers.

**Disagreements:**
- There is no verdict disagreement: Claude, Codex, and DeepSeek all block this lane.
- Claude emphasizes marker fingerprints and no-data-loss proof; Codex emphasizes exact path/key ledgers and the order-dispatch label model; DeepSeek emphasizes distributed liveness, hash performance, and remote split-brain. My assessment: these are complementary parts of the same blocker, because runtime-state migration cannot be safely decomposed until the concrete ledger, marker, lock/liveness, and reconciliation contracts are all explicit.
- DeepSeek suggests neutralizing legacy git remotes to prevent split-brain pushes. Claude and Codex do not mandate that exact mechanism, but they agree the design must prove no duplicate or dropped pushes. The implementation can choose a different mechanism only if the push-cursor and remote policy covers rollback and split repositories.
- Claude frames process-unique staging and atomic publish as major supporting requirements; Codex's review makes the broader copy/old-writer race blocking. My assessment: treat the race as blocking and the staging/publish details as required changes that make the blocker testable.
- DeepSeek proposes mtime/size before SHA-256 for performance. The exact cheap-check strategy can vary, but the design must avoid full archive hashing on every startup/reload unless a mismatch warrants it.

**Missing evidence:**
- A runtime migration marker schema with per-artifact legacy path identity, source fingerprint at commit, destination fingerprint, high-water mark, completed step, and conflict policy.
- A concrete old-to-Core path/key/ref/remote ledger for JSONL export state, JSONL archive repo, archive refs/remotes, spawn-storm ledger, order `order-run:<scoped>` labels, order-tracking labels, formula environment state, escalation fields, and provider state whose path changes.
- Current-system citations for each migrated state row, including Maintenance-pack orders/scripts and owning Go packages.
- A two-snapshot migration algorithm and failure-injection tests for old or non-cooperative writers between staging and marker commit.
- A deterministic push-cursor, ref, and remote reconciliation policy proving no duplicate or dropped pushes across interrupted copy, pending push state, rollback, and re-upgrade.
- Alias-table semantics for dispatch gating, cooldown/last-run lookup, event cursor lookup, stale/orphan sweeps, and `gc order history`.
- Spawn-storm dedup keys and post-marker monitoring of legacy throttle ledgers.
- Provider-aware liveness detection for Kubernetes or an explicit refusal/unsupported-mode policy.
- A downgrade/version-skew matrix covering rollback before migration, after marker before Core-only writes, and after Core-only writes.
- Round-trip tests for interrupted archive copy, half-copied `.git`, duplicate writers, old-binary controller liveness, in-flight push state, rollback, and deterministic re-upgrade or manual reconciliation.

**Required changes:**
- Define the runtime migration marker schema per artifact, including legacy path identity, source fingerprint at marker commit, destination fingerprint, artifact-specific high-water mark, completed step, cheap divergence metadata, and conflict policy.
- Replace placeholder migration-table rows with exact canonical legacy and Core-owned paths, keys, refs, remotes, fields, owners, reader/writer call sites, merge rules, refusal rules, marker fields, proof commands, and expiry/alias-retirement conditions.
- Specify the migration as a quiesced two-snapshot protocol: acquire the city advisory lock, verify no live controllers or affected writer processes, snapshot legacy state, stage the Core copy with a manifest, re-snapshot sources immediately before marker commit, and restart or block if sources changed.
- Make archive publish fail closed with process-unique staging, manifest validation, `.git`/refs/remotes/HEAD validation, destination-absent first-copy semantics, source/destination digest checks, and atomic same-filesystem directory rename or a documented fallback that cannot expose partial state.
- Define push-cursor, ref, and remote reconciliation so rollback or split legacy/Core repositories cannot duplicate, drop, or race archive pushes.
- Add an order-identity alias table and require alias-aware dispatch gating, last-run/cursor lookup, stale/orphan sweep, and `gc order history`; re-key live entries or block alias retirement until migration completes.
- Define spawn-storm ledger deduplication and monitor old throttle-ledger writes as part of post-marker divergence detection and re-upgrade.
- Add provider-aware live-controller detection for Kubernetes runtimes or fail closed with explicit operator guidance when distributed liveness cannot be proven.
- Add a tiered post-marker divergence check that avoids unconditional full hashing of large retained archives on startup/reload.
- Add a downgrade/version-skew policy that either mirrors Core runtime writes during a compatibility window or explicitly makes downgrade after Core-only writes unsupported without manual export/reconciliation.
- Add an integration round trip covering interrupted archive copy, half-copied `.git`, in-flight push state, duplicate old writers, old-binary rollback, version-skew blocking, and deterministic re-upgrade or manual reconciliation.
