# Yuki Hayashi — Rollout and Version Skew Perspective Independent Review (Iteration 20 / Attempt 20)

**Verdict:** approve

**Scope:** Rollout sequencing, version skew, public pack pin integrity, intermediate state safety, and rollback granularity.

This review evaluates the Iteration 20 / Attempt 20 draft of `design.md` against `requirements.md` and the existing codebase behavior.

---

## Executive Summary

The Iteration 20 design document exhibits a highly disciplined, safety-first approach to rolling out the Core and Gastown pack split. By defining a **7-Slice Rollout Plan** (§3372–3457) and establishing a robust **Release Compatibility Matrix** (§3465–3472), the design successfully avoids "flag-day" release risks, provides explicit roll-back guarantees, and ensures that intermediate states remain test-green and deployable.

Specifically, the design addresses key version skew and rollout challenges:
1. **Decoupled Repo Sequencing:** `gascity-packs` lands the public Gastown behavior first (§3377), ensuring the public commit is available before Gas City updates its internal pin (§3389, §3496).
2. **Non-Destructive Cleanup:** Legacy directories (`.gc/system/packs/maintenance` or `gastown`) are ignored rather than deleted on startup (§2975), preserving operator edits and facilitating seamless binary rollbacks.
3. **Robust Cache Keying:** `RepoCacheKey` for public Gastown uses the normalized source and exact version pin to avoid collision with historical synthetic aliases (§2947).

The design is exceptionally strong and is **approved** from the perspective of **Rollout and Version Skew Planning**.

---

## Top Strengths

- **No Flag-Day Dependency (§3375, §3496):** The sequencing guarantees that `gascity-packs` is updated and verified before Gas City attempts to pin it, preventing references to unavailable public commits.
- **Table-Driven Release Compatibility Matrix (§3465–3472):** Spelling out expected behaviors across four binary/pack skew combinations (and rollback) provides clear guidance for developers and ensures backward compatibility.
- **Preservation of Ignored Legacy Paths (§2975):** Avoiding aggressive startup deletion of stale system pack directories prevents accidental data loss and ensures that a downgraded binary immediately regains access to its required assets.
- **Fail-Closed Verification on Materialization (§137, §2954):** Disallowing silent fallback to obsolete embedded templates ensures that any installation or reachability issue is reported loudly and cleanly, maintaining the integrity of the public pin.

---

## Nuanced Risks & Recommendations

To guarantee a completely seamless and risk-free rollout, the following operational recommendations are provided:

### 1. Mixed-Binary Version Skew in Active Cities
- **The Risk:** In a multi-agent environment, if a city's controller is upgraded to the new binary while active agent sessions (running in background tmux panes or sub-processes) are still running the old binary, those background sessions might execute stale commands or run legacy, unsafe `gc doctor --fix` routines, potentially leading to metadata corruption or state conflicts.
- **Recommendation:** Mandate in the release/migration notes that operators must perform a clean shutdown of all active city sessions (e.g., via `gc stop`) before upgrading the `gc` binary. Add a preflight or startup check in the new controller that warns if active processes from a different binary version are detected in the city.

### 2. Network Reachability and Offline Test Seeding
- **The Risk:** Fresh `gc init --template gastown` resolves the public pack from the network source or remote cache (§2954). If the test harness running `TestPinnedPublicGastownBehavior` (§3402) in Slice 2 does not strictly pre-populate and isolate the local cache, tests may attempt unexpected network calls, causing intermittent failures in air-gapped CI environments.
- **Recommendation:** Ensure that the test harness for `TestPinnedPublicGastownBehavior` strictly runs in a hermetic local-fixture mode where the ordinary remote cache is pre-seeded with a local mock of the pinned Gastown commit, verifying that `gc init` succeeds completely offline when the cache is populated.

### 3. Cryptographic Integrity Check for Public Pins
- **The Risk:** While pinning to a git commit hash (`PublicGastownPackVersion`) is cryptographically secure against simple tag manipulation, a git history rewrite or force-push on the remote repo could theoretically change or remove the commit.
- **Recommendation:** In the long term, consider calculating and asserting a content-hash or integrity checksum (digest) of the materialized public pack upon download, ensuring absolute, tamper-proof alignment between the pinned definition and the materialized files.

---

## Evaluation of the Three Key Questions

### 1. What does a fresh `gc init --template gastown` produce between the `gascity-packs` landing and the Gas City `PublicGastownPackVersion` pin update, and is that state deployable?
- **Planner Finding:** **Deployable and Stable.** During this intermediate window, the operator's local `gc` binary remains unchanged and continues to use the existing in-tree Gastown template or synthetic cache alias. Since the in-tree assets are not deleted until the final Slice 7 (§3449), the operator is completely unaffected by the remote land. Once the Gas City binary is updated (Slice 2), it transitions to resolving the public pack from the remote pin, which is verified and green before release.

### 2. Is `PublicGastownPackVersion` pinned to immutable content with materialization-time verification rather than a mutable branch or tag?
- **Planner Finding:** **Yes.** The design mandates pinning to an "immutable compatibility/activation commit" (§37, §3390, §3424) rather than a mutable branch or tag. Verification occurs at materialization-time via `RepoCacheKey` and ordinary remote-pack cache paths keyed by repository source and immutable version (§2943, §2947).

### 3. Can Gas City registry changes be reverted after operators fetched the new public pack without leaving cities with neither Maintenance nor Gastown behavior?
- **Planner Finding:** **Yes.** This is guaranteed by the design's non-destructive startup policy (§2975) which ignores rather than deletes legacy directories, combined with the fact that the new public Gastown pack remains backward-compatible with older host Core binaries (§3468). If a rollback occurs, the old binary immediately recovers in-tree assets or reads the ignored-but-preserved local directories, ensuring continuous behavior.

---

## Recommendations for Finalization

1. **Mandate "Clean Stop" Upgrade Procedure:** Include a clear requirement in the migration guide that all active city sessions must be stopped before upgrading the binary.
2. **Seeded Offline Cache Tests:** Add a dedicated test assertion in `TestPinnedPublicGastownBehavior` that verifies `gc init` fails cleanly when offline with an empty cache, but succeeds completely offline when the cache is pre-seeded.
