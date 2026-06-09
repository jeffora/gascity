# Yuki Hayashi — DeepSeek V4 Flash (Rollout Version Skew Review, Iteration 4, Independent)

**Verdict:** approve-with-risks

**Lane:** Two-repo rollout sequencing, public pack pin integrity, intermediate state safety, rollback granularity.

Reviewed against the iteration 4 design-before document (updated 2026-06-07T00:30, containing the newly specified `Required Core Identity And Loader Contract`, `Doctor And Import-State Safety`, and the multi-slice rollout staging) and grounded in the `cmd/gc/`, `internal/config/`, and `internal/builtinpacks/` packages.

---

## Executive Summary

The iteration 4 design document presents an exceptionally thorough, highly structured rollout and version-skew strategy. By transitioning from a monolithic transition plan to an explicit **7-slice staged rollout**, the design systematically addresses the race conditions, un-syncable repositories, and intermediate-state vulnerabilities identified in previous review iterations. 

Crucially, the addition of the **"Maintenance Runtime And Duplicate-Order Contract"** in Slice 5 guarantees behavior-level atomicity during the Maintenance fold, eliminating the threat of duplicate active orders. Furthermore, the commitment to the `TestPinnedPublicGastownBehavior` packcompat gate in Slice 2 ensures that new pins are verified against real remote-pack installs before behavior ownership is transferred.

I am giving this design a verdict of **approve-with-risks**. While the rollout staging and intermediate-state safety are structurally robust, there remain a few major operational and diagnostic gaps regarding version skew detection in the doctor check and import resolution behavior when stale directories are present.

---

## Top Strengths of the Iteration 4 Design

1. **Test-Gated 7-Slice Rollout Partitioning**
   Breaking the migration into seven discrete, independently deployable, and test-gated slices represents the gold standard for two-repo rollout. By updating the public pin in Slice 2 and proving its compatibility *prior* to extracting Core (Slice 3) or deleting old assets (Slice 7), the design eliminates the "Gas City points at an unavailable public commit" risk.

2. **Atomic Maintenance folding & Retirement Table**
   The detailed retirement table and state-path classification (for JSONL archives, cursors, and storm ledgers) resolve the intermediate-state vacuum where Maintenance is retired but Core hasn't fully assumed its duties. Enforcing that "no intermediate slice may expose duplicate active order definitions" is a major architectural win.

3. **Robust Rollback and Mixed-Version Compatibility**
   The release compatibility matrix provides a complete, logical state-space coverage (old binary × old pack, old binary × new pack, new binary × old pack, new binary × new pack, and rollback). Explicitly declaring that a doctor-mutated manifest must remain readable by old binaries ensures that rollbacks are smooth and non-destructive.

---

## Lane Question Analysis

### 1. What does a fresh `gc init --template gastown` produce between the `gascity-packs` landing and the Gas City `PublicGastownPackVersion` pin update, and is that state deployable?

**It produces a stable, pre-migration city locked to the old pinned public commit, which is fully deployable.** 
During this intermediate window (between Slice 1's public repository landing and Slice 2's pin update), the `gc` binary still contains the old `PublicGastownPackVersion` SHA. Any fresh init will write that old SHA to the city's lockfile. 
Since the old public commit is immutable and still exists on GitHub, and the binary's behavior hasn't changed yet, this state loads cleanly, behaves exactly like the pre-migration system, and is completely deployable. This protects operators from landing on an untested or broken "halfway" state.

### 2. Is `PublicGastownPackVersion` pinned to immutable content with materialization-time verification rather than a mutable branch or tag?

**Yes, via Git commit SHA immutability and standard cache validation.**
The design-before specifies that `PublicGastownPackVersion` is a hardcoded Git commit SHA constant. By avoiding branch names or tags, source-level immutability is guaranteed. 
At materialization time, since public packs reside under standard remote-pack cache paths (rather than the bundled synthetic cache), the SDK relies on Git's internal cryptographically secure hashing during `git checkout`. This guarantees that the cached content matches the pinned commit exactly. The design correctly transitions public Gastown from the synthetic aliases to standard remote cache keys (`config.RepoCacheKey`) to avoid hash/alias collisions with stale bundled states.

### 3. Can Gas City registry changes be reverted after operators fetched the new public pack without leaving cities with neither Maintenance nor Gastown behavior?

**Yes.** If registry changes are reverted (or the operator downgrades the binary to an older release):
1. **Old Binary Implicit loading**: The old binary's `requiredBuiltinPackNames` still includes `"maintenance"`, meaning the old binary will auto-materialize and include Maintenance assets as an implicit config layer, even if the new doctor removed explicit imports from `pack.toml`.
2. **New Pack Backward Compatibility**: The new public Gastown pack is explicitly required to remain backward-compatible with older binaries and the presence of Maintenance, ensuring no load failures due to missing loader-level capabilities.
Therefore, reverting registry changes or rolling back the binary preserves a functional behavior set, preventing the city from entering a dead zone with zero maintenance/gastown coverage.

---

## Critical Risks & Gaps

### [Major] Actionable Version-Skew Diagnostics lack concrete implementation specifications
The compatibility matrix promises that for the `new binary | old pack` skew:
> *doctor reports an actionable version-skew diagnostic without mutating custom content.*

However, under the "Doctor and Import-State Safety" or "Core Presence Doctor" specifications, there is no detailed check outlined to detect remote public Gastown imports pinned to commits older than `PublicGastownPackVersion`. 
* **The Risk:** Without an explicit check that parses the public pack's pinned SHA and compares it against the binary's `PublicGastownPackVersion` constant, the new binary might load an outdated public Gastown pack, encounter missing files or broken references (such as missing `../maintenance`), and fail closed with a generic load error rather than an "actionable version-skew diagnostic."
* **Recommendation:** Explicitly specify a doctor check that compares the locked commit SHA of `github.com/gastownhall/gascity-packs/gastown` with the current binary's `PublicGastownPackVersion`. If the lock is older, report a warning with a clear `FixHint` explaining that the lock is on an older pre-split version and recommending `gc doctor --fix` to update the pin.

### [Major] Import Resolution failure during Version Skew when Stale Directories are Ignored
The design states:
> *Stale `.gc/system/packs/maintenance` and `.gc/system/packs/gastown` directories are ignored, diagnosed, and preserved, not deleted.*

Under the `new binary | old pack` skew, the old pack's `pack.toml` still contains `imports = ["../maintenance"]`. 
* **The Risk:** Since the new binary ignores `.gc/system/packs/maintenance`, when the config loader attempts to resolve the `../maintenance` relative import from the old Gastown pack, the resolution will fail. Depending on whether import failures are hard or soft (load error vs diagnostic warning), this will either break config loading entirely or silently drop essential formulas.
* **Recommendation:** The design must clarify the expected behavior when a relative import to `../maintenance` is encountered by the loader in a legacy pack. I recommend that the loader translates relative imports to retired system packs into a soft diagnostic warning rather than a fatal loading error, allowing the new binary to at least boot and run doctor to repair the city.

### [Minor] Air-Gapped Offline materialization gap for Public Gastown
By removing `gastown` from the embedded set in Slice 6 and retiring the public synthetic aliases, a fresh `gc init --template gastown` will always require network access to fetch the remote pack.
* **The Risk:** Operators on air-gapped systems or offline developers will experience silent failures or un-actionable clone errors.
* **Recommendation:** The design mentions an "offline behavior decision." The implementation should ensure that if the clone fails due to being offline, a highly visible diagnostic is printed: *"Public Gastown pack could not be cloned. Fresh Gastown initialization is not supported in offline/air-gapped environments without a pre-populated repository cache."*

---

## Required Changes & Action Items

Before declaring the migration complete, the implementation team must verify:
1. **Add explicit doctor check for remote version skew**: Implement a diagnostic that warns when `Imports["gastown"].version` is older than `PublicGastownPackVersion`.
2. **Implement relative import tolerance**: Ensure that the config loader does not fail closed when a pack imports a missing/retired relative path like `../maintenance`, but instead logs a diagnostic and continues.
3. **Prove rollback idempotency in CI**: Add an integration test that takes a city initialized under the new binary, downgrades it to the pre-migration binary, and asserts that the city remains operational.

---

## Verdict: Approved with Risks

The Stage 4 design is highly mature, robustly staged, and thoroughly mitigates the biggest rollout risks of the two-repo split. Address the doctor diagnostic gap for remote version skew and relative import tolerance during implementation to guarantee a seamless transition for production operators.
