# Yuki Hayashi — Gemini (Rollout Version Skew Planner, Attempt 12, Independent DeepSeek V4 Flash Style)

**Verdict:** approve-with-risks

**Lane:** Two-repo rollout sequencing, public pack pin integrity, intermediate state safety, rollback granularity.

Reviewed against the revised design document in Attempt 12 (`.gc/design-reviews/ga-2404qu/attempt-12/design-before.md` updated 2026-06-07T14:05Z) and grounded in the live codebase in `cmd/gc/`, `internal/config/`, and `internal/builtinpacks/`.

---

## Executive Summary

As Yuki Hayashi, the **Rollout Version Skew Planner**, I am updating my verdict to **Verdict: approve-with-risks** for Attempt 12.

The revised design document represents an extremely high-fidelity and realistic rollout plan. By permanently retiring the fiction of the loader's "inactive compatibility assets" (§1763) and establishing the **Attempt 11 Review Resolution Contracts** (specifically the explicit `public-gastown-pins.yaml` ledger and the `Concrete Core Maintenance-Worker Binding`), the architecture successfully ensures that role de-coupling is solid and rollback granularity is well-defined.

However, while the structural rollout plan is exceptionally mature, a rigorous systems-level analysis of intermediate states and backward-compatibility paths reveals several critical risks and logical contradictions. Specifically, the **non-destructive copy strategy** during doctor state migration introduces a split-brain synchronization risk upon downgrade, and the paired cross-repo fallback mechanism undermines our core zero-flag-day invariant. These risks must be resolved to guarantee seamless transitions for operators in the field.

---

## Detailed Responses to Lane-Specific Questions

### Q1: What does a fresh `gc init --template gastown` produce between the `gascity-packs` landing and the Gas City `PublicGastownPackVersion` pin update, and is that state deployable?

**It produces a stable, pre-split city pinned to the legacy remote commit of public Gastown, which is fully operational and deployable.**

During this intermediate window (after the public pack lands but before Gas City updates the pin):
1. The `gc` binary's `PublicGastownPackVersion` still holds the old, immutable Git commit SHA.
2. The binary continues to include `"maintenance"` in `requiredBuiltinPackNames` and embeds the legacy Maintenance pack.
3. When `gc init` runs, it writes the old SHA to the city's lock file.

Since the old remote commit is immutable and remains reachable, and the binary continues to materialize and inject the embedded Maintenance pack, this intermediate state loads cleanly and behaves exactly like the pre-migration system. This guarantees that operators are never exposed to broken intermediate configurations.

### Q2: Is `PublicGastownPackVersion` pinned to immutable content with materialization-time verification rather than a mutable branch or tag?

**Yes, via a cryptographically immutable Git commit SHA and standard remote-cache integrity validation.**

The design strictly pins the version to a 40-character Git commit SHA. At materialization time, the public Gastown pack is resolved as a remote import (`config.RepoCacheKey`) from the public repository, bypassing the bundled synthetic pack layout. 
- It is downloaded and checked out via standard Git protocols, enforcing Git's cryptographically secure object hashing.
- Stale synthetic aliases are retired (`IsSource(".../gastown") == false`), preventing synthetic-cache shadowing.
- The design correctly establishes that public Gastown is treated as an ordinary remote pack, verified against the pinned Git commit rather than matching the binary's embedded bytes.

### Q3: Can Gas City registry changes be reverted after operators fetched the new public pack without leaving cities with neither Maintenance nor Gastown behavior?

**Yes. The design maintains robust backward compatibility and rollback continuity.**

If an operator downgrades the binary (reverting registry changes) after having updated their public pack imports:
1. **Old Binary Inclusion**: The older binary's `requiredBuiltinPackNames` still includes `"maintenance"`. Upon startup, the older binary will auto-materialize and include the legacy Maintenance pack as an implicit config layer, even if the new doctor had previously removed explicit Maintenance imports from `pack.toml`.
2. **New Pack Backward Compatibility**: The new public Gastown pack in the `gascity-packs` repo is explicitly required to remain backward-compatible with older binaries and the presence of Maintenance, ensuring no crashes or load failures due to duplicate symbols or missing loader hooks.

Therefore, the city is never left in a behavioral vacuum; the older binary gracefully provides the fallback Maintenance layer, ensuring zero downtime.

---

## Detailed Risks & Gaps (DeepSeek V4 Flash Style)

### 1. [Major] Rollback State Split-Brain and Progress Loss on Downgrade

- **The Risk**: The design specifies that during doctor migration, JSONL state/archives and spawn-storm ledgers are **copied** to Core when the Core destination is absent, and the legacy Maintenance state remains in place as ignored legacy state (§1884).
- **The Impact**: If an operator downgrades the binary after running on the new version, the older binary only reads and writes the legacy path (`.gc/runtime/packs/maintenance/`). Any execution state, skip markers, or spawn-storm counts written by the new binary to the Core path (`.gc/runtime/packs/core/`) are completely invisible to the old binary. This results in a **split-brain state** where the downgraded city suffers from state-synchronization lag, duplicate command execution, or silent progress regression.
- **Recommended Action**: The doctor or migration documentation must warn operators of this state drift. Alternatively, the downgraded binary's doctor should provide an explicit recovery check that can merge state from `.gc/runtime/packs/core/` back to the legacy location if drift is detected.

### 2. [Major] The Paired Cross-Repo Release Trap (Flag-Day Fallacy)

- **The Risk**: In `Activation Pin Has No Inactive Loader Fiction` (§1785), the design states: *"If the compatibility public pack cannot omit colliding active assets, the project must skip the compatibility pin and land a paired cross-repo activation/removal boundary instead."*
- **The Impact**: Skipping the compatibility pin and forcing a paired cross-repo release is by definition a **flag-day release**—violating the core design goal of preventing flag-days across repositories. If an operator updates their binary but is blocked by network issues from fetching the pinned remote pack, or updates the pack on an older binary, they will experience immediate load failures or duplicate symbol collisions.
- **Recommended Action**: Formally forbid skipping the compatibility pin. If collisions exist, they must be resolved by refactoring the public pack assets (e.g., using conditional config or version-spaced subpaths) to ensure that the public pack remains compatible with both old and new binaries during the transition.

### 3. [Major] Degradation of Offline City Initialization and Complex Operator Cache Pre-population

- **The Risk**: Under the new cache policy, the public Gastown pack must resolve through ordinary remote repository paths, and synthetic cache aliases are retired (§1904).
- **The Impact**: A fresh `gc init --template gastown` in an offline or air-gapped environment will fail unless the cache is pre-populated. Air-gapped deployments are blocked from standard initialization unless operators perform complex, undocumented manual pre-population of the remote cache.
- **Recommended Action**: Support a local cache import or fixture injection during `gc init` for air-gapped environments, or document the exact cache directory structure and pre-population protocol in `docs/reference/system-packs.md`.

### 4. [Minor] Cache Invalidation Storm during Rapid Pin Reversions

- **The Risk**: The registry, packman, and config loader all share the public cache key (`Source + Commit + Subpath`).
- **The Impact**: If operators rapidly toggle between commits during a rollback or retry, the registry may repeatedly trigger remote git checkouts and garbage collection, resulting in lock contention or disk I/O thrashing.
- **Recommended Action**: Use standard git reference caches to speed up toggling and prevent complete directory deletion during checkout.

---

## Evaluation against Lane Anti-patterns

| Anti-pattern / Risk | Mitigation in Attempt 12 Design | Status |
| :--- | :--- | :--- |
| **Gas City points at an unavailable or untested public Gastown commit** | **Excellent.** Locked via the `public-gastown-pins.yaml` authoritative ledger, which requires ordinary remote-cache checkout proof and old/new binary compatibility evidence before the pin can be updated in `internal/config/`. | **Pass** |
| **Migration requires a flag-day release across repositories** | **Weakness identified.** The design allows skipping the compatibility pin in favor of a paired cross-repo release if collisions cannot be avoided (§1785). This introduces a classic flag-day risk. | **Risk (Needs Action)** |
| **Rollback path depends on undocumented manual operator repair** | **Partial Risk.** While old binary compatibility is preserved, the split-brain state drift on runtime state after downgrade is undocumented and can cause silent loss of execution progress. | **Risk (Needs Action)** |

---

## Final Verdict: Approved with Risks

The Attempt 12 design is an outstanding, highly disciplined blueprint that comprehensively mitigates the most critical rollout and version-skew risks. By addressing the **downgrade state split-brain** and enforcing the **immutability of the compatibility pin** (preventing the paired cross-repo flag-day trap), the team will ensure a flawless, operator-safe transition.

Proceed with implementation under these guidelines.
