# Leah Okafor - Codex

**Verdict:** block

**Top strengths:**
- The design now puts doctor mutation behind an explicit `internal/doctorfix` coordinator with intent staging, publish, recovery, refusal, and controller-running refusal semantics.
- The recovery model correctly rejects the false premise that cross-file per-file renames are atomic; it introduces durable recovery state, step tracking, post-commit validation, and deterministic rerun or rollback behavior.
- Runtime-state migration is assigned to `gc doctor --fix` rather than controller startup, API handlers, or reload paths, which is the right boundary for avoiding live-controller mutation races.

**Critical risks:**
- [Major] The existing `Check.Fix(ctx)` surface is not made mechanically closed. The design says existing checks either return structured intents or become report-only, but it does not require a closure artifact or static gate over current direct fix implementations such as zombie/orphan session stopping, broken worktree deletion, instructions symlink creation, Dolt remote removal, Codex hook installation, and import-state rewrites. Without that gate, new safe import/runtime fixes can coexist with old direct mutations and the coordinator will not actually be the only mutation path.
- [Major] Byte-preserving TOML repair is specified as a desired test outcome, not as a staging contract. Current edit helpers load typed structs and write serialized `city.toml` or `pack.toml`, which is exactly the path that can drop comments, unknown tables, unknown fields, ordering, or formatting. The plan needs to say how a `FixIntent` represents scoped TOML edits, how `Stage` proves the diff touches only intended bytes or semantic locations, and when full-file reserialization is forbidden or refused.
- [Major] Required-pack repair and quarantine still have a boundary tension with `internal/systempacks`. The plan says `LoadRuntimeCity` owns materialization/validation and the Data section says required-pack repair regenerates, prunes, and quarantines files before behavior reads them, while the doctor section says the coordinator is the only path that writes installed pack directories. The document needs to decide whether runtime loading is validate-only and doctor-owned repair performs writes, or whether `systempacks` calls the same locked coordinator.

**Missing evidence:**
- A current-fix-surface table naming every in-tree `CanFix() == true` check, its side effects, and its target disposition: `doctorfix` intent, explicitly allowed non-durable operational action, or report-only.
- A static or generated test that fails when production doctor checks keep direct mutating `Fix(ctx)` implementations outside the approved disposition table.
- A concrete TOML preservation mechanism, such as source-range patches or an edit verifier that compares old and staged bytes and refuses any unrelated byte changes.
- The exact city advisory lock identity and how it relates to the controller lock and pack install/update locks.
- The exact recovery-state path and marker schema for doctor fix intents, not just the fields it should contain.

**Required changes:**
- Add an "Existing Doctor Fix Surface Closure" subsection or support artifact that inventories all current auto-fix checks and states which ones migrate to `FixIntent`, which become report-only, and which non-durable operational fixes remain direct with an explicit reason. Tie this to a scanner or focused test so direct mutation bypasses cannot regress.
- Define the new doctor runner contract: checks should return/report fix intents, the runner invokes `internal/doctorfix`, and `Check.Fix(ctx)` is either removed, deprecated behind an allowlist, or constrained to non-mutating/report-only compatibility.
- Specify the TOML edit representation and byte-preservation verification required before publish. The design should explicitly forbid doctor fixes from using typed full-config reserialization unless the verifier proves only intended bytes changed; otherwise the fix must refuse with operator guidance.
- Resolve the `systempacks` versus `doctorfix` write boundary for required-pack repair/quarantine. If runtime load writes required-pack directories, it must use the same lock, staging, recovery, and controller-running rules; otherwise normal load should fail closed and point operators to `gc doctor --fix`.
- Name the lock path/API and recovery-state path/schema that implementers should use so all pack install, import rewrite, required-pack repair, and runtime-state migration tasks share one synchronization contract.

**Questions:**
- Are session/process cleanup, broken worktree deletion, hook installation, and symlink creation considered part of the same doctor mutation-safety migration, or are they intentionally allowed as operational fixes outside `internal/doctorfix`?
- Should `LoadRuntimeCity` ever mutate `.gc/system/packs/*`, or should it only validate/report while `gc doctor --fix` owns repair and quarantine?
