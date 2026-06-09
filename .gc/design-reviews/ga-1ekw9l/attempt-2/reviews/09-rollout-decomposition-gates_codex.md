# Iris Kowalski

**Verdict:** approve-with-risks

**Top strengths:**
- The plan is honest that full Gas City implementation decomposition is blocked until AC6, AC7, and AC14-AC17 support artifacts exist and pass. It explicitly limits early beads to external prerequisites, proof generators, non-mutating inventories, packcompat harness work, and generated docs/schema proof.
- The AC17 table gives each acceptance criterion a proof artifact or command, gate placement, and first dependent slice. That is the right shape for preventing hidden prerequisite drift during decomposition.
- The rollout slices now name cross-repo sequencing, merge gates, rollback paths, and one-way boundaries instead of batching public pin changes, Core extraction, doctor mutation, Maintenance folding, cache cleanup, and source deletion into one landing.

**Critical risks:**
- [Major] The acceptance-proof matrix is still described as a design gate, not as an executable gate with a concrete validator command. The plan says `support/acceptance-proof-matrix.yaml` validates AC1-AC16 evidence availability before bead creation, but it does not name the command, package, or CI/pre-decomposition hook that fails when an AC row points at a missing artifact, missing external commit, or unavailable release gate. Without that executable gate, later task creation can silently outrun the prerequisites.
- [Major] Slice 2 is too broad for a first Gas City public-pin adoption slice. It updates `PublicGastownPackVersion`, adds packcompat in current-loader mode, rejects retired synthetic cache hits, rewires `examples/gastown`, and also requires unsafe legacy import rewrites to be disabled or routed through the mutation coordinator. That mixes pin/cache behavior, examples rewrite, and doctor mutation gating. It is decomposable, but as written one failed sub-gate can block or partially land unrelated work.
- [Major] The cross-repo prerequisite handoff is clear conceptually but not operationally pinned enough for task decomposition. Slice 1a-1c depend on `gascity-packs` ownership rows, behavior-preservation manifests, pin ledgers, packcompat transcripts, and activation commits, but the plan does not say how Gas City beads will represent those external prerequisites, what exact artifact paths/commit fields unblock the Gas City slices, or who records the immutable evidence into the Gas City support artifacts.
- [Minor] The plan says every intermediate code slice runs broad gates, and high-risk loader/doctor/runtime-state slices run process and integration shards. Public-pin, cache, docs/schema, and cross-repo packcompat slices are also high risk; the slice table should mark their exact local, deterministic public-pack CI, release-gate, and rollback verification commands per slice so `tasks.md` cannot default to the fast unit baseline.

**Missing evidence:**
- The exact validator command for `support/acceptance-proof-matrix.yaml`, including failure behavior for missing local files, missing external commits, unavailable release-gate transcripts, and AC rows with manual-only evidence.
- A narrower decomposition of Slice 2 into independently mergeable pin-coherence/cache, example import rewrite, and doctor/import-rewrite safety work.
- The artifact handoff contract for public `gascity-packs` work: required file paths, required commit fields, digest fields, transcript fields, and the Gas City support artifact that records each prerequisite as satisfied.
- Per-slice gate commands for deterministic public-pack CI and release-gate evidence, not just broad category names.

**Required changes:**
- Add the executable acceptance-proof validator command and state that prerequisite-producing decomposition may create only tasks whose outputs are accepted by that validator or explicitly marked as external-prerequisite work.
- Split Slice 2 into at least two slices: one for pin/cache proof and synthetic-alias rejection, and one for `examples/gastown` rewrite plus import-state/doctor safety wiring. Keep behavior-changing import rewrites behind the doctorfix gate.
- Define the cross-repo prerequisite handoff row format in the Gas City support artifacts: external repo, commit, subpath, pack digest, behavior-manifest digest, packcompat transcript path, approving PR, and first Gas City slice allowed to consume it.
- Expand the slice-to-gate table so each code slice names its exact required local command(s), deterministic public-pack command(s), release-gate evidence, and rollback verification.

**Questions:**
- Will prerequisite-producing beads for `gascity-packs` live in the Gas City tracker, the public pack repo tracker, or both, and what evidence closes the Gas City-side gate?
- Should the acceptance-proof validator block all implementation-bead creation, or only behavior-changing/deletion/mutation slices while allowing proof-generator tasks?
