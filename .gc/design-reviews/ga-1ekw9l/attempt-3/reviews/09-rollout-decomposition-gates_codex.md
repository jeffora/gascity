# Iris Kowalski - Codex

**Verdict:** approve-with-risks

**Top strengths:**
- The rollout is split into independently understandable slices: public Gastown prerequisite, compatibility pin, Core extraction, `internal/systempacks`, activation/Maintenance fold, registry/cache cleanup, and final stale-source/docs cleanup (`design-before.md:453`-`design-before.md:502`).
- The plan separates compatibility and activation pins, which prevents source deletion and Maintenance removal from being batched with the first public-pack compatibility proof (`design-before.md:461`-`design-before.md:489`).
- High-risk slices name broader gates beyond fast unit tests, including packcompat, packlint, sharded process/integration targets, `make test-fast-parallel`, and `go vet ./...` (`design-before.md:414`-`design-before.md:451`).

**Critical risks:**
- [Major] The plan is decomposition-ready only if the prerequisite artifacts become explicit first-class tasks and dependencies. Open Questions says external commits, generated manifests, pin ledger, ownership rows, and packcompat transcripts are prerequisites rather than design questions, but task decomposition must encode those as blocking beads before any source deletion, Maintenance removal, or activation-pin consumption. Otherwise the blockers are hidden in prose.
- [Major] Several slices are still large enough to hide multiple merge boundaries. Slice 4 combines `internal/systempacks`, strict validation, typed participation, production loader scanner, partial-read allowlist, Core doctor, pre-resolution import recovery, version-skew diagnostics, and mutation coordinator. Those should decompose into smaller vertical tasks with their own acceptance gates and rollback notes.
- [Minor] Rollback boundaries are named, but one-way upgrade constraints need to be explicit per task once mutation coordinator state or runtime-state migration markers can commit. Tasks after those points should say whether rollback is code-only, recovery-rerun, or manual/downgrade-limited.

**Missing evidence:**
- A task dependency model showing public Gastown branch/commit, behavior manifest, pin ledger, and packcompat transcript beads blocking Gas City destructive changes.
- Per-slice acceptance gates mapped to exact commands and artifact paths.
- A decomposition rule that prevents bundling source deletion, activation pin update, doctor mutation, docs updates, and cache alias retirement in one bead.
- A one-way/rollback classification for tasks that commit mutation coordinator state or runtime-state migration markers.

**Required changes:**
- In `tasks.md`, create explicit prerequisite tasks for public Gastown compatibility/activation commits, generated manifests, pin ledger, ownership rows, and packcompat transcripts before any Gas City source-deletion tasks.
- Split Slice 4 into narrower tasks: loader boundary, participation records, scanner/allowlist, diagnostic/read-only reload gating, mutation coordinator, and pre-resolution recovery.
- Add per-task gates that include the relevant exact tests from the Testing section and mark which tasks require process/integration shards.
- Add rollback classification to each task: reversible by git revert, reversible after coordinator recovery, or one-way/manual recovery boundary.

**Questions:**
- Will the first decomposition artifact create beads in both Gas City and `gascity-packs`, or will the external public-pack prerequisite be tracked as an external blocking artifact?
- Which task owns the initial generated behavior manifest schema and sample rows, and which later tasks are blocked by it?
- At what exact slice does rollback stop being a simple git revert because durable city state may have changed?
