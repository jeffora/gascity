# Leah Okafor

**Persona verdict:** block

**Sources:** Claude, Codex, DeepSeek V4 Flash

**Consensus findings:**
- [Blocker] Byte-preserving TOML mutation is asserted but not specified as an implementation contract. All sources flag that the current BurntSushi/full-encode paths cannot preserve comments, unknown fields, table order, array order, or formatting. The design must name a CST/span-patching mechanism with an outside-bytes verifier, or explicitly ship TOML mutation as refuse-only until that mechanism exists.
- [Blocker] The multi-file transaction model is not decomposition-safe. A WAL and staged publish model is the right direction, but the plan does not define the single durable commit operation, pre-commit versus post-commit states, journal/staged-file fsync ordering, or replay semantics. POSIX rename is per-file, not an atomic cross-file transaction.
- [Blocker] Rollback is physically impossible as currently described. The recovery record stores digests, staged paths, publish order, commit point, completed steps, and rollback instructions, but not original bytes. The design must either add original-byte backups for true rollback or remove rollback language and commit to roll-forward-only convergence.
- [Blocker] The advisory-lock and live-controller exclusion story is incomplete. The plan assumes a shared city mutation lock and controller exclusion, but current protected writers do not all take that lock, the controller lock is separate, and a process-table live-state check leaves a TOCTOU window. The controller, doctor, import/update, pack install/update, concurrent doctor runs, and behavior-changing read/write paths need one shared synchronization contract.
- [Blocker] Required-pack read-path repair contradicts the doctor coordinator's "only writer" invariant. The plan says normal loading regenerates, prunes, and quarantines installed pack files, while the doctorfix section says installed pack directories are written only through the coordinator. That read-path repair can race doctor publishing unless it is removed, made report-only, or routed through the same lock/coordinator boundary.
- [Major] Old-binary and concurrent-writer divergence detection is too narrow. Runtime-state markers detect some old-binary post-marker writes, but the coordinator also mutates manifests, lockfiles, installed pack directories, cache proof files, and import rewrites. Final digest validation and refusal/blocking behavior must cover every touched surface, including drift after an earlier publish step.
- [Major] Provenance refusal is not explicit enough. Automatic fixes must refuse custom, forked, operator-edited, digest-mismatched, or otherwise unproven imports and generated paths rather than relying on path shape or broad provenance language.
- [Major] The existing mutation surface closure is not yet enforceable. The plan needs to reconcile `internal/doctorfix` with current direct `Check.Fix(ctx)` paths, import rewrites, `internal/configedit`, and load-path writes, then back that closure with an inventory and CI scanner or focused tests.
- [Minor] Non-interactive refusal behavior is unspecified. `gc doctor --fix --non-interactive` should define whether refusal returns a non-zero exit, what guidance is emitted, and whether refusal is a hard failure or no-op.

**Disagreements:**
- There is no verdict disagreement in the current inputs: Claude, Codex, and DeepSeek V4 Flash all return `block`.
- TOML preservation framing differs. Claude stresses the current toolchain makes "preserve or refuse" likely inert on real cities; Codex asks for a structured `ScopedTomlPatch` intent and outside-bytes proof; DeepSeek asks either for a named format-preserving editor or an explicit refuse-only slice. My assessment is that all three describe the same blocker: the plan must move TOML preservation from a test assertion into the `doctorfix` API and implementation model.
- Controller exclusion is framed at different strictness levels. DeepSeek wants the active controller to hold the same city advisory lock continuously; Codex wants all-surface old-writer drift detection and a commit protocol; Claude asks for proof that controllers, old binaries, and existing mutators contend on the same lock. My assessment is that continuous controller lock ownership is one acceptable implementation, but the required design outcome is shared mutual exclusion plus drift detection for non-cooperating binaries.
- Recovery strategy can be resolved two ways. Claude and DeepSeek emphasize that rollback requires original bytes; Codex focuses on commit marker mechanics and all-surface recovery. My assessment is that either rollback-with-backups or roll-forward-only can pass, but the current hybrid wording cannot.
- DeepSeek uniquely elevates custom/fork provenance refusal as a standalone blocker. Claude and Codex cover provenance through digest/provenance revalidation and operator-edited path refusal. This synthesis keeps explicit provenance refusal as required because it is a direct guard against destructive auto-fix.

**Missing evidence:**
- The concrete TOML editing substrate, scoped patch representation, and byte-preservation verifier, including how this replaces, wraps, or avoids `internal/configedit` and whole-file `toml.NewEncoder` rewrites.
- A complete inventory of current `CanFix() == true` checks, direct `Fix(ctx)` implementations, import rewrites, pack install/update mutations, lockfile writes, installed pack directory writes, runtime-state writes, load-path repairs, and test helpers, with a disposition for each.
- An enforcement scanner or focused CI test proving protected writes cannot bypass `internal/doctorfix`.
- The shared city advisory lock path/API, acquisition order, owner set, lifecycle, timeout/refusal behavior, stale-lock behavior, and release point.
- Whether controllers hold the shared mutation lock continuously, acquire it only around behavior-changing paths, or use another generation/epoch protocol that prevents doctor/controller interleaving.
- The recovery transaction state machine: durable journal path, staged artifact path, fsync order, commit marker operation, pre-commit and post-commit meanings, final validation, stale-record cleanup, and idempotent replay after a second crash.
- Whether original bytes are backed up for rollback, or whether recovery is explicitly roll-forward-only from staged artifacts and deterministic regeneration.
- All-surface concurrent-writer detection for manifests, lockfiles, cache proofs, installed pack directories, runtime-state markers, and import rewrites.
- Explicit refusal rules for custom, forked, operator-edited, unproven, or digest-mismatched imports and generated files.
- Non-interactive refusal exit status and operator-facing guidance.

**Required changes:**
- Name the TOML preservation strategy in Proposed Implementation. Define the scoped TOML intent payload, outside-bytes verification, and refusal behavior; forbid whole-file TOML reserialization for operator-authored files unless verification proves only intended bytes changed.
- Define the `doctorfix` transaction protocol for multi-file fixes: lock acquisition, staging, journal and staged-file fsync, single commit marker, publish order, final validation, cleanup, and second-crash replay.
- Choose the recovery model. Add original-byte backups for rollback, or make recovery explicitly roll-forward-only and remove "rolls back" claims.
- Define one shared city mutation lock or equivalent synchronization protocol across doctor, controller lifecycle, import/update, pack install/update, load-path repair, concurrent doctor runs, and behavior-changing runtime operations. Include timeout/refusal behavior and the exact release point.
- Resolve the required-pack loader repair contradiction by making read-path repair report-only, moving it through `doctorfix`, or routing it through the same lock with a narrowed only-writer invariant.
- Extend old-binary/concurrent-writer detection and post-commit validation to every surface touched by doctorfix, not only runtime-state markers. A drift after any publish step must leave a diagnostic that blocks behavior-changing runtime use until reconciliation or deterministic re-upgrade.
- Add explicit provenance refusal rules for custom, forked, operator-edited, unproven, and digest-mismatched imports or generated files.
- Add an inventory and enforcement test that closes current direct `Fix(ctx)` and protected-write bypasses, and clarify the relationship between `internal/doctorfix`, `internal/configedit`, and current import/pack mutation helpers.
- Specify `gc doctor --fix --non-interactive` refusal semantics, including exit status and guidance.
