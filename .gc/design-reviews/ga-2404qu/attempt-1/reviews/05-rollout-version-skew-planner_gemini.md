# Yuki Hayashi — DeepSeek V4 Flash Independent Review (Rollout & Version Skew, Iteration 18)

**Verdict:** approve-with-risks

**Lane:** Two-repo rollout sequencing, public pack pin integrity, intermediate state safety, and rollback granularity.

Reviewed against the revised design document in Iteration 18 (`.gc/design-reviews/ga-2404qu/attempt-1/design-after.md`) and grounded in the live codebase in `cmd/gc/`, `internal/config/`, and `internal/builtinpacks/`.

---

## Design Evolution & Top Strengths

The design's evolution from the initial Attempt 1 drafts up to Iteration 18 is exemplary. It demonstrates a profound understanding of version-skew mechanics and operational safety in distributed, multi-agent systems. Key advancements include:

- **Non-Destructive Copy-on-Absence State Migration (§549, §553)**: In direct response to early critiques, the design moves away from destructive state-file renames or unilateral moves. By adopting a "copy when Core destination is absent" policy for the JSONL archive/export state and spawn-storm ledgers, the legacy directories are preserved as high-fidelity rollback targets.
- **Complete Public Synthetic Alias Retirement (§1992, §2645)**: Completely disabling the old embedded synthetic alias short-circuits for public Gastown and Maintenance commits forces clean, remote-first resolution. This removes the risk of "illusionary" compatibility test runs that actually load old embedded bytes.
- **Authoritative Ledger Integration (`public-gastown-pins.yaml`)**: Adding a three-phase ledger (`current_baseline`, `compatibility`, `activation`) that records immutable digests, behavior manifests, and cache keys provides a robust cryptographic foundation for validation.

---

## Detailed Responses to Lane-Specific Questions

### Q1: What does a fresh `gc init --template gastown` produce between the `gascity-packs` landing and the Gas City `PublicGastownPackVersion` pin update, and is that state deployable?

**It produces a stable, pre-split city pinned to the older, immutable public Gastown commit, which is fully operational and deployable.**

During this intermediate window (after the public pack lands on the remote branch in `gascity-packs` but before Gas City's pin is updated):
1. The older `gc` binary's `PublicGastownPackVersion` still holds the legacy commit SHA.
2. The older binary continues to materialize and inject the embedded Maintenance pack because `"maintenance"` remains in `requiredBuiltinPackNames`.
3. Running `gc init` writes the old SHA to the city's lock file.

Since the old remote commit is immutable and remains reachable, and the binary continues to supply the embedded Maintenance pack, this intermediate state loads cleanly and behaves exactly like the pre-migration system. This guarantees that operators are never exposed to broken intermediate configurations.

---

### Q2: Is `PublicGastownPackVersion` pinned to an immutable content hash, or can a mutable ref drift to a different commit between materialization and verification?

**Yes, the pin is defined strictly as an immutable Git commit SHA (e.g., `sha:d3617d13...`).**

At materialization time, the loader (`internal/packman/cache.go:197–213`) strictly compares `git rev-parse HEAD` against the expected pin using `gitutil.SameCommit` and rejects dirty worktrees. This prevents any mutable branch or tag ref drift.

**The Verification Gap:**
While required host system packs (like Core) are validated using strict Pre-Resolution file-set manifests and digests, the remote public Gastown pack lacks a corresponding cryptographic file-set digest in the local binary configuration. Relying solely on Git's content hash leaves the local cache `.gc/cache` vulnerable to silent file corruption or local tampering. To close this gap, the Attempt 14 ledger `public-gastown-pins.yaml` has been introduced to track and validate explicit file-set and behavior digests.

---

### Q3: Can Gas City registry changes be reverted after operators fetched the new public pack without leaving cities with neither Maintenance nor Gastown behavior?

**Yes, the binary downgrade path is structurally supported, but the runtime-state transition introduces a subtle split-brain risk (see Critical Risks).**

If an operator downgrades the binary (reverting registry changes) after having updated their public pack imports:
1. **Old Binary Inclusion**: The older binary's `requiredBuiltinPackNames` still includes `"maintenance"`. Upon startup, the older binary will auto-materialize and include the legacy Maintenance pack as an implicit config layer, even if the new doctor had previously removed explicit Maintenance imports from `pack.toml`.
2. **State Conservation**: Because state migration is non-destructive and copies rather than moves the JSONL state and spawn-storm counts, the legacy directory `.gc/runtime/packs/maintenance/` remains populated. The downgraded old binary will find its historical state and resume normal operations without crashing.

---

## Critical Risks & Gaps

### 1. [Blocker] Critical Go-Side Enforcement Gap for `RequiresGC` Version Constraints

- **The Risk**: Requirement **AC15** explicitly mandates that *"Old binary plus activation pin is unsupported and must fail closed with downgrade guidance."* To support this, the public Gastown pack's `pack.toml` declares `requires_gc = ">=0.14.0"`. Static analysis of `internal/config` confirms that while `RequiresGC` is correctly parsed from `pack.toml` into `config.Pack`, **there is absolutely no validation, comparison, or enforcement logic implemented in the Go codebase**. 
- **The Hazard**: Because there is no active Go-side enforcer checking this parsed constraint, an older binary running in the wild will blindly load an activation-pinned Gastown pack.
- **The Impact**: When the older binary loads the new activation pack, it will also force-include the legacy embedded Maintenance pack (as `"maintenance"` remains in the old binary's `requiredBuiltinPackNames`). This will result in dual active definitions of identical roles, tasks, mail, and molecules (e.g., duplicate `prune-branches` handlers). Rather than failing closed safely with downgrade guidance, the system will execute with dangerous, silent symbol collisions, risking catastrophic double-spend or state corruption.
- **Required Action**: Implement a robust semantic version validation check (e.g., using `github.com/Masterminds/semver`) on the parsed `RequiresGC` string inside `internal/config/` during pack resolution. If the active binary's version is less than the parsed requirement, the loader must immediately abort execution with clear, user-facing downgrade guidance.

### 2. [Blocker] Split-Brain Rollback Gap in Doctor State Migration

- **The Risk**: The state migration table (§549, §553) specifies that `gc doctor --fix` and the first Core run copy state files (like `jsonl-export-state.json` and `spawn-storm-counts.json`) from legacy `maintenance/` to `core/` *"when Core destination is absent"*. 
- **The Scenario**: 
  1. An operator upgrades the binary and runs `gc doctor --fix`. The state is copied to `core/`, and completion is recorded.
  2. The operator discovers an issue and downgrades the binary to the legacy version.
  3. The downgraded binary runs and updates the state files in `maintenance/` (advancing export cursors, logging new spawn storms).
  4. The operator resolves the issue, upgrades back to the new binary, and runs `gc doctor --fix` again.
  5. Because the Core state files already exist from step 1, the "when Core destination is absent" guard is **not met**. The doctor skips copying the updated legacy state.
- **The Impact**: The new binary will load the stale Core state from step 1, discarding all cursor progress, spawn-storm counts, and message-delivery offsets accumulated during the downgrade period. This will lead to silent duplicate message exports, infinite loop re-runs, and critical data inconsistency.
- **Required Action**: Do not rely on the simple absence of the Core destination. The doctor must perform a timestamp-based or monotonic version-counter check (or append a migration epoch header) to detect if the legacy state has advanced past the Core state and safely re-synchronize them during downgrade-upgrade transitions.

### 3. [Major] Squash-Merge Orphaned-Commit Unreachability Risk

- **The Risk**: Slice 1 lands the public pack changes "on a branch" in `gascity-packs`. If `PublicGastownPackVersion` in Gas City is pinned to the branch's tip commit, and that branch is subsequently merged into `main` using a "Squash and Merge" or "Rebase and Merge" policy, the branch's tip commit SHA is discarded. The branch is then deleted.
- **The Impact**: The branch commit becomes orphaned (unreferenced by any branch or tag) and will eventually be garbage-collected by the Git host. When an operator runs a fresh `gc init --template gastown` weeks later, the command will fail with a fatal checkout error because the pinned commit SHA is no longer reachable on the remote repository.
- **Required Action**: The release-engineering pipeline must enforce that any commit pinned in `PublicGastownPackVersion` (including historical compatibility and activation pins) is reachable from a protected, permanent tag or the default branch of `gascity-packs`.

### 4. [Major] Asymmetric Verification Gaps (Zero-Duplicate is Gated, Zero-Gap is Not)

- **The Risk**: During the intermediate rollout slices (Slices 2 to 4), a city runs a hybrid configuration of host Core, bundled Maintenance, and the public Gastown compatibility pin. The design enforces a strict "zero duplicate active definitions" gate to prevent symbol collisions (§706–716). However, there is no symmetric "zero gap" or completeness gate. The compatibility pin deliberately omits colliding assets to avoid duplicate errors. If any of these omitted assets are not fully supplied by the bundled Maintenance pack or the host Core pack (due to a coordination drift or mismatched assumptions), the intermediate city will silently lose that behavior without raising an error.
- **The Impact**: Intermediate cities during slices 2–4 may suffer from silent capability regressions (e.g., missing reapers, unswept orphans, or dropped mail notifications) while passing all CI gates because no test asserts that the union of `compatibility-pin active assets ∪ bundled Maintenance ∪ host Core` matches 100% of the target state's behavior manifest rows.
- **Required Action**: Add a symmetric "no-gap" completeness gate to the Slice 2-4 test suites. The compatibility pin validation test (`TestPinnedPublicGastownBehavior`) must assert that the union of active assets across all three layers matches 100% of the target state's behavior manifest rows.

### 5. [Minor] Codependency and Atomicity of Slice 5a and 5b

- **The Risk**: Slice 5a (pinning the public activation commit) and Slice 5b (removing Maintenance from required host packs) are deeply codependent. Merging Slice 5a into the master branch without Slice 5b causes immediate compile-time or loader failures because the activation-pinned pack expects a role-neutral host environment with no implicit Maintenance injection.
- **The Impact**: Breaking the master branch build or introducing transiently broken intermediate commits that violate basic loader invariants.
- **Required Action**: Formally mandate that Slices 5a and 5b must be bundled and merged as a single, atomic PR/change boundary.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Iteration 18 Design | Status |
| :--- | :--- | :--- |
| **Gas City points at an unavailable or untested public Gastown commit** | **Excellent.** Resolved via the authoritative `public-gastown-pins.yaml` ledger, remote-cache checkout proof, and explicit content digest evidence before the pin can be updated in `internal/config/`. | **Pass** |
| **Migration requires a flag-day release across repositories** | **Excellent.** Eliminated by the two-pin rollout model (compatibility pin followed by activation pin) which ensures every intermediate rollout slice is fully deployable. | **Pass** |
| **Rollback path depends on undocumented manual operator repair** | **Pass with Caveat.** The transition is now non-destructive (copy-on-absence), but the split-brain rollback scenario remains a blocker risk. | **Risk (Needs Action)** |

---

## Actionable Requirements & Proposed Adjustments

1. **RequiresGC Go-Side Enforcer (§3205)**: Implement robust semantic version validation in `internal/config` that compares the active binary's version against the parsed `RequiresGC` string, failing closed with clear downgrade guidance if the requirement is not met.
2. **State Migration Epoch/Sync Guard (§549, §553)**: Replace the "when Core destination is absent" condition with an explicit epoch or timestamp-based comparison to prevent split-brain state loss on downgrade-then-upgrade transitions.
3. **Durable Ref CI Enforcement (§3202)**: Add a Git-reachability check to Gas City's CI that verifies every SHA pinned in `PublicGastownPackVersion` exists and is reachable from a protected ref or default branch in `gascity-packs`.
4. **Symmetric "No-Gap" Completeness Gate (§2633)**: Add an automated validation test verifying that the union of active assets in Slices 2–4 matches 100% of the target behavior manifest rows.
5. **Atomic Slices 5a/5b Merger**: Enforce that the promotion to the activation pin (5a) and the removal of Maintenance from the required host packs (5b) land as a single atomic commit/PR.
