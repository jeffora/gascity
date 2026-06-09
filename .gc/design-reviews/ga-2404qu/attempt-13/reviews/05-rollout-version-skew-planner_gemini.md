# Yuki Hayashi — Gemini (Rollout & Version Skew Reviewer, Attempt 13, Independent DeepSeek V4 Flash Style)

**Verdict:** approve-with-risks

**Lane:** Two-repo rollout sequencing, public pack pin integrity, intermediate state safety, and rollback granularity.

Reviewed against the revised design document in Attempt 13 (`.gc/design-reviews/ga-2404qu/attempt-13/design-before.md` updated 2026-06-07T14:05Z) and grounded in the live codebase in `cmd/gc/`, `internal/config/`, and `internal/packs/`.

---

## Detailed Responses to Lane-Specific Questions

### Q1: What does a fresh `gc init --template gastown` produce between the `gascity-packs` landing and the Gas City `PublicGastownPackVersion` pin update, and is that state deployable?

**It produces a stable, pre-split city pinned to the old, immutable public Gastown commit, which is fully operational and deployable.**

During this intermediate window (after the public pack lands on the remote branch in `gascity-packs` but before Gas City's pin is updated):
1. The older `gc` binary's `PublicGastownPackVersion` still holds the legacy commit SHA.
2. The older binary continues to materialize and inject the embedded Maintenance pack.
3. Running `gc init` writes the old SHA to the city's lock file.

Since the old remote commit is immutable and remains reachable, and the binary continues to supply the embedded Maintenance pack, this intermediate state loads cleanly and behaves exactly like the pre-migration system. This guarantees that operators are never exposed to broken intermediate configurations.

### Q2: Is `PublicGastownPackVersion` pinned to immutable content with materialization-time verification rather than a mutable branch or tag?

**Yes, via a cryptographically immutable Git commit SHA, but materialization-time verification contains a major security/integrity gap.**

The design strictly pins the version to a 40-character Git commit SHA (§2626, §2654). At materialization time, the public Gastown pack is resolved as a remote import from the public repository, bypassing the bundled synthetic pack layout. 
- It is downloaded and checked out via standard Git protocols, enforcing Git's cryptographically secure object hashing.
- Stale synthetic aliases are retired (§2670), preventing synthetic-cache shadowing.

**The Integrity Gap:**
While required host system packs (like Core) are validated using strict Pre-Resolution file-set manifests and digests (§112-126, §600-615), the public Gastown pack lacks a corresponding cryptographic file-set digest in `internal/config`. Relying solely on Git commit SHAs leaves the local cache `.gc/cache` vulnerable to silent file corruption or local tampering. If a file in `.gc/cache` is altered, the loader will parse it without warning as long as the directory name matches the commit SHA.

### Q3: Can Gas City registry changes be reverted after operators fetched the new public pack without leaving cities with neither Maintenance nor Gastown behavior?

**Yes, the binary downgrade path is structurally supported, but the current runtime-state migration design introduces a catastrophic rollback trap.**

If an operator downgrades the binary (reverting registry changes) after having updated their public pack imports:
1. **Old Binary Inclusion**: The older binary's `requiredBuiltinPackNames` still includes `"maintenance"`. Upon startup, the older binary will auto-materialize and include the legacy Maintenance pack as an implicit config layer, even if the new doctor had previously removed explicit Maintenance imports from `pack.toml`.
2. **New Pack Backward Compatibility**: The new public Gastown pack in the `gascity-packs` repo is explicitly required to remain backward-compatible with older binaries and the presence of Maintenance, ensuring no crashes due to duplicate symbols.

**The Rollback Trap:**
When `gc doctor --fix` runs under the new binary, it migrates Maintenance runtime state (`jsonl-export-state.json` and `spawn-storm-counts.json`) to Core, leaving the legacy Maintenance path empty or declaring rollback "advisory" (§491-513). If the operator downgrades, the older binary expects state files under `.gc/runtime/packs/maintenance/`. Finding them gone/moved, the old binary will re-initialize all JSONL export cursors and spawn-storm counts, triggering **infinite spawn storms and duplicate external message exports**.

---

## Critical Risks & Gaps

### 1. [Blocker] Slice 5's Circular Rollout Dependency and Undeployable Intermediate State
- **Line Reference**: `design-before.md:L2653-2659`
- **The Risk**: Slice 5 rollout instructions state: *"update internal/config/PublicGastownPackVersion from the compatibility commit to the activation commit ... then run the no-Maintenance production-loader packcompat gate before removing Maintenance from active required packs."*
- **The Impact**: Updating the pin to the activation commit while Maintenance is still in the required pack list means both will load. By definition, the activation pin re-activates assets that the bundled Maintenance pack still supplies (e.g. `prune-branches`). This will trigger a duplicate active symbol collision, which is fatal under our design's own rules (§706-716). Operators or CI building/testing at this intermediate commit will get immediately rejected.
- **Recommended Action**: Re-write Slice 5 rollout text so that the activation pin update and the removal of Maintenance from `requiredBuiltinPackNames` are implemented in a single, atomic commit, with the no-Maintenance gate run on the candidate combined tree before merging.

### 2. [Major] One-Way Destructive Runtime-State Migration (Rollback Trap)
- **Line Reference**: `design-before.md:L491-513`
- **The Risk**: Moving or renaming `jsonl-export-state.json` and `spawn-storm-counts.json` from `maintenance/` to `core/` means downgrading to an older binary (which only reads `maintenance/`) will cause silent cursor resets and massive task re-exports.
- **The Impact**: Declaring state rollback "advisory" (§509) directly violates our zero-manual-repair mandate. If a critical regression is discovered in the new binary, operators are locked in a split-brain state or forced to manually edit JSONL cursor state.
- **Recommended Action**: Change the doctor migration to **copy** (not move) legacy Maintenance state to Core, or keep both updated during the transition.

### 3. [Major] Non-Hermetic Test and Init Environments in Offline/Air-gapped Scenarios
- **Line Reference**: `design-before.md:L2625-2636` (Slice 2)
- **The Risk**: Bypassing synthetic aliases means `gc init --template gastown` always resolves through the remote git repo.
- **The Impact**: In offline, air-gapped, or sandboxed CI environments, tests that try to fetch from `github.com` will fail immediately. Requiring the internet for standard test suites or initialization is a severe violation of build hermeticity.
- **Recommended Action**: Establish a local, vendored mock/fixture for the public Gastown commit within the Gas City test suite (e.g., in `test/packcompat/testdata/`), ensuring that all compatibility and init tests can pass completely offline without relying on real remote Git fetches.

### 4. [Major] Operator Confusion from Silent Ignore of Stale Directories
- **Line Reference**: `design-before.md:L2707-2710`
- **The Risk**: Stale `.gc/system/packs/maintenance` or `.gc/system/packs/gastown` directories are preserved in place to protect operator edits, relying on the loader to ignore them.
- **The Impact**: Operators may edit files in these directories, assuming they are still active, but their changes will be silently ignored. This creates major operational confusion and debugging overhead.
- **Recommended Action**: During `gc doctor --fix`, stale system pack directories must be **renamed** to a retired namespace (e.g., `.gc/system/packs/maintenance.retired.bak`). This completely isolates them from active search paths while preserving every byte of the operator's custom edits.

### 5. [Minor] Lack of Cryptographic Digest Verification for Public Pack Cache
- **Line Reference**: `design-before.md:L2626` (Slice 2)
- **The Risk**: Local cache corruption is not detected because `PublicGastownPackVersion` is checked via Git commit only, without content-manifest or file-set digest verification.
- **The Impact**: If a file in `.gc/cache` is corrupted or tampered with, the loader will parse it without warning, bypassing the security guarantees established for required host system packs.
- **Recommended Action**: Add materialization-time digest verification for public Gastown content, or explicitly name the existing manifest/file-set digest path that provides equivalent protection.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Attempt 13 Design | Status |
| :--- | :--- | :--- |
| **Gas City points at an unavailable or untested public Gastown commit** | **Excellent.** Locked via the `public-gastown-pins.yaml` authoritative ledger, which requires ordinary remote-cache checkout proof and old/new binary compatibility evidence before the pin can be updated in `internal/config/`. | **Pass** |
| **Migration requires a flag-day release across repositories** | **Weakness identified.** The design allows skipping the compatibility pin in favor of a paired cross-repo release if collisions cannot be avoided (§2657-2659). This introduces a classic flag-day risk. | **Risk (Needs Action)** |
| **Rollback path depends on undocumented manual operator repair** | **Fail.** The rollback contract for state migration is designated as "advisory" (§509) and left to manual recovery, which is a major rollback trap. | **Block** |

---

## Required Changes

To advance this design to a green status, the following concrete gates must be incorporated into the implementation slices:

1. **Slice 2 (Public Pin Gate) - Hermetic Offline Fallback**: Add an integration test asserting that the compatibility pin can be loaded and verified in a fully air-gapped environment using a local pre-populated test cache/fixture.
2. **Slice 4 (Core Doctor Gate) - Manifest Backups**: Assert that `gc doctor --fix` automatically creates pre-migration backups in `.gc/backups/` and that an interrupted or failed fix leaves user files completely byte-identical.
3. **Slice 5 (Maintenance Folding Gate) - Non-Destructive State Migration**: Prove via integration tests that rolling back from a migrated state to the old binary does not cause JSONL export cursor resets or spawn-storm counter resets by ensuring the legacy state is preserved during the transition.
4. **Slice 5 (Maintenance Folding Gate) - Stale Directory Isolation**: Verify that stale system-pack directories are renamed to the `.bak` namespace and are completely unreachable by active glob searches.
5. **Slice 6 (Registry/Cache Gate) - Cryptographic Verification**: Verify that remote pack materialization asserts the downloaded file-set SHA-256 digest against a hardcoded `PublicGastownPackDigest` in the binary config.

---

## Questions

- Are the activation pin bump and Maintenance removal one commit? If not, what is an operator on the intermediate commit supposed to experience?
- For old binaries, does the import resolver prefer the synthetic public alias or the ordinary remote cache when both could satisfy `gascity-packs.git//gastown` at a sha pin?
- After a binary downgrade post-activation, is the supported recovery doctor-driven lock downgrade or manual re-pin per release notes — and which slice's release notes own it?
