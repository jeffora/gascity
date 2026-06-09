# Iris Kowalski - Codex

**Verdict:** block

**Top strengths:**
- The rollout now separates public Gastown ownership, compatibility proof, activation proof, Gas City compatibility-pin adoption, Core extraction, loader hardening, doctor/runtime migration, Maintenance fold, registry/cache cleanup, and stale-source deletion.
- The external public-pack prerequisite is no longer hand-waved: it names `public-gastown-pins.yaml`, `behavior-preservation.yaml`, `ownership.yaml`, and packcompat transcripts as blocking evidence.
- The testing section goes beyond fast unit tests and calls out packcompat, packlint, production-loader checks, sharded process tests, integration shards, and `go vet ./...`.

**Critical risks:**
- [Major] The rollout slices are still not directly decomposable because gates are mostly global. Each slice needs its own file/module list, prerequisite inputs, exact merge gates, and rollback or one-way boundary. Today Slice 4a, 4b, 4c, 5b, 6, and 7 each bundle several implementation streams, while the proof commands live in a separate Testing section. A task planner could create beads from the slice prose but still miss which command or external artifact must pass before that specific slice merges.
- [Major] External prerequisite honesty is documented, but not enforced as a decomposition rule. The plan says missing public `gascity-packs` commits are not open questions, yet the local plan directory currently contains only `requirements.md` and `implementation-plan.md`; there is no named prerequisite ledger or validator artifact for task creation to consume. Without an explicit rule that deletion, activation-pin, alias-retirement, and Maintenance-removal beads are blocked until exact commits/artifact digests are recorded, decomposition can silently turn "wait for external proof" into ordinary implementation work.
- [Minor] The activation and cleanup tail still combines reversible and one-way changes. Slice 5b mixes removing Maintenance from required packs, moving Core-owned Maintenance assets, consuming public Gastown assets, and stale-path handling; Slice 6 mixes embedded-registry removal, synthetic-alias retirement, cache-key enforcement, and stale-cache rejection; Slice 7 mixes source deletion and final docs. These should be separate decomposition gates so rollback boundaries stay honest.

**Missing evidence:**
- A per-slice gate matrix with columns for changed files/directories, cross-repo prerequisite artifacts, proof commands, generated artifacts, rollback procedure, and whether the slice is reversible after merge.
- A concrete prerequisite ledger path in the Gas City repo, or an instruction that task creation must create prerequisite beads and stop before creating dependent source-deletion or activation beads.
- The exact slice that introduces validators for `public-gastown-pins.yaml`, `behavior-preservation.yaml`, `ownership.yaml`, and packcompat transcript freshness.
- Per-slice production-loader proof. The plan says broad high-risk slices run process/integration shards, but it does not tie each loader/doctor/runtime slice to the specific command proving it avoided copied fixtures and direct `config.Load` bypasses.

**Required changes:**
- Add a rollout/decomposition gate table that maps each slice to concrete affected files, exact prerequisite artifacts, exact test commands, generated freshness checks, and rollback or one-way behavior. Keep the prose narrative, but make the table the source a task planner can decompose from.
- Add an explicit decomposition rule: any bead that deletes in-tree Gastown/Maintenance sources, removes Maintenance from required host packs, consumes the activation pin, or retires synthetic aliases must depend on recorded immutable public-pack commits and digest-checked artifacts. If those artifacts do not exist, task creation should produce prerequisite/blocker beads instead of implementation beads.
- Split the tail-end slices so reversible compatibility work, activation-pin consumption, required-pack removal, embedded-registry cleanup, synthetic-alias retirement, cache-key migration, stale-source deletion, and final docs each have separate gates and rollback expectations.
- Name the first Gas City slice that lands prerequisite validators and require downstream slices to run them before merge.
- Attach production-loader proof commands to the exact slices that need them, not only to the global Testing section.

**Questions:**
- Should the first task plan create external `gascity-packs` prerequisite beads, or should Gas City decomposition stop until those artifacts already exist?
- Where should the local prerequisite ledger live if the external artifacts are not yet available: under `plans/core-gastown-pack-migration/support/`, checked-in implementation paths, or only bead metadata?
