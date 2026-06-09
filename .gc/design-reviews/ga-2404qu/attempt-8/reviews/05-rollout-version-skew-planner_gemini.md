# Tomoko Hayashi — DeepSeek V4 Flash (Rollout Version Skew Review, Iteration 8, Independent)

**Verdict:** approve-with-risks

**Lane:** Two-repo rollout sequencing, public pack pin integrity, intermediate state safety, rollback granularity.

Reviewed against the iteration 8 design document (containing the machine-readable `behavior-manifest.generated.yaml` specification, `Pre-Resolution Doctor and Legacy Import Recovery` phase, the explicit Duplicate-Definition Gate, and the 7-slice rollout staging) and grounded in the `cmd/gc/`, `internal/config/`, and `internal/builtinpacks/` packages.

---

## Executive Summary

The iteration 8 design is exceptionally robust, and represents a masterclass in two-repo migration orchestration. The design successfully addresses the major architectural and testing blind spots identified in earlier review rounds:

1. **Synthetic Shadowing Defeated**: By requiring the public-pin adoption slice (Slice 2) to explicitly retire and bypass the public synthetic Gastown alias, the design ensures that `PublicGastownPackSource` must resolve through a real remote repository check and exact commit verification. This closes the testing loophole where the local binary's embedded `examples/gastown/packs/gastown` bytes would mask real remote-fetch and pin-integrity verification.
2. **Duplicate-Definition Gate**: The addition of a strict duplicate active definitions gate during config loading prevents the critical risk of an intermediate-state city running overlapping/conflicting formulas, orders, or scripts from both legacy local paths and the new remote pack.
3. **Pre-Resolution Doctor and Legacy Import Recovery**: Performing a pre-load TOML parsing pass restricted purely to imports/sources guarantees that the CLI can boot, diagnose, and repair stale configurations *before* a full, failing config evaluation.

I am granting this design a verdict of **approve-with-risks**. The rollout timeline and rollback invariants are structurally complete. However, the implementation team must address several remaining risk areas around air-gapped environment fallbacks, stale local-state shadowing, and transactional atomicity during doctor-driven directory migrations.

---

## Lane Question Analysis

### 1. What does a fresh `gc init --template gastown` produce in the window between the `gascity-packs` landing and the Gas City `PublicGastownPackVersion` pin update, and is that state deployable?

**It produces a stable, fully deployable, pre-split city locked to the legacy remote pin.**

During this intermediate window (Slice 1 complete, but Slice 2 pending):
1. The `gc` binary’s `PublicGastownPackVersion` still holds the older, pre-migration Git commit SHA.
2. The binary continues to include `"maintenance"` in `requiredBuiltinPackNames` and embeds the legacy Maintenance pack assets.
3. When `gc init` is run, it writes the old pinned SHA to the city's lock file.

Because the old public commit is immutable and reachable on GitHub, and the running binary still auto-materializes and injects the embedded Maintenance pack, this intermediate state loads cleanly, behaves exactly like the pre-migration system, and is fully deployable. This ensures that operators are never exposed to broken "flag-day" configurations.

### 2. Is `PublicGastownPackVersion` pinned to immutable content with materialization-time verification rather than a mutable branch or tag?

**Yes, via a cryptographically immutable Git commit SHA and standard remote-cache integrity validation.**

The design hardcodes `PublicGastownPackVersion` to a 40-character Git commit SHA in `internal/config/public_packs.go`. At materialization time, the public Gastown pack is resolved as a remote import (`config.RepoCacheKey`), bypassing the bundled synthetic pack layout. 

This ensures that:
- It is downloaded and checked out via standard `git clone`/`git checkout` protocols, which enforce Git's cryptographically secure object hashing.
- Stale synthetic aliases are retired (`IsSource(".../gastown") == false`), preventing synthetic-cache shadowing.
- The design correctly establishes that public Gastown is treated as an ordinary remote pack, meaning its content is verified against the pinned Git commit rather than matching the binary's embedded bytes.

### 3. Can Gas City registry changes be reverted after operators fetched the new public pack without leaving cities with neither Maintenance nor Gastown behavior?

**Yes. The design maintains robust backward compatibility and rollback continuity.**

If an operator downgrades the binary (reverting registry changes) after having updated their public pack imports:
1. **Old Binary Inclusion**: The older binary's `requiredBuiltinPackNames` still includes `"maintenance"`. Upon startup, the older binary will auto-materialize and include the legacy Maintenance pack as an implicit config layer, even if the new doctor had previously removed explicit Maintenance imports from `pack.toml`.
2. **New Pack Backward Compatibility**: The new public Gastown pack in the `gascity-packs` repo is explicitly required to remain backward-compatible with older binaries and the presence of Maintenance, ensuring no crashes or load failures due to duplicate symbols or missing loader hooks.

Therefore, the city is never left in a behavioral vacuum; the older binary gracefully provides the fallback Maintenance layer, ensuring zero downtime.

---

## Detailed Risks & Gaps

### [Major] Air-Gapped and Low-Connectivity Local Development Failures
The design requires `PublicGastownPackSource` to resolve through the ordinary remote repository path during the slice-2 adoption phase. 
* **The Risk**: Operators running in air-gapped production environments or local developers working with limited network access will experience immediate failures during `gc init` or doctor verification because the resolver must contact GitHub to fetch the exact commit of `gascity-packs/gastown`.
* **Recommendation**: Introduce a configuration or environment override (e.g. `GC_OFFLINE_PACK_MIRRORS`) allowing operators to point the remote-pack resolver to a local Git mirror or directory path. If this is unavailable, the doctor must produce highly actionable diagnostics indicating that a network connection is required to complete the initial remote materialization.

### [Major] Orphaned/Stale Cache Drift and Ignored Directory Shadowing
The design specifies that:
> *Stale `.gc/system/packs/maintenance` or `.gc/system/packs/gastown` directories are ignored by config loading after the migration.*
* **The Risk**: If an operator had previously modified or added custom, untracked scripts or active hook files in these folders, they will remain on disk. If the config loader silently ignores these directories, the operator's custom behavior will silently stop executing without warning. Furthermore, if relative imports are used by some custom formulas, they might still target these ignored directories, leading to bizarre version skew and debugging nightmares.
* **Recommendation**: The doctor must scan `.gc/system/packs/` and loudly warn if ignored legacy directories exist, reporting any custom or modified files (by comparing file hashes against the known synthetic baseline) to the operator instead of silently bypassing them.

### [Major] TOML Round-trip Lossiness on Operator Manifests
The design-before specifies:
> *TOML edits are scoped and preserving. If comments, unknown tables, array order, or formatting cannot be preserved, `gc doctor --fix` refuses and gives manual steps.*
* **The Risk**: If the pre-resolution doctor uses Go's standard `toml.NewEncoder` to write back modified manifests, the encoder will discard comments, reorder tables, and strip unrecognized custom fields. Since almost all real-world cities contain hand-crafted comments or operator-specific metadata, a naive decode-encode loop will cause the doctor to either corrupt the files or constantly refuse to fix, forcing manual operator repair for virtually 100% of production upgrades.
* **Recommendation**: The implementation of the pre-resolution doctor must use a concrete syntax tree (CST) parser (such as `github.com/pelletier/go-toml/v2`'s AST or targeted regex line-based search and replace) to surgically edit the `imports` array while preserving formatting, comments, and unknown keys exactly.

### [Major] Failure-Atomic Transaction State Sequencing during doctor fixes
The pre-resolution doctor performs both manifest TOML rewrites and runtime-state file migrations (moving JSONL logs, archive repositories, and count files from `/packs/maintenance` to `/packs/core`).
* **The Risk**: If a partial failure occurs mid-execution (for example, the TOML file write succeeds but moving the JSONL archive directory fails due to a disk permission or file lock), the city will be left in a corrupted split-state. The manifest will look fully migrated, but the runtime state will remain stranded in the legacy location, causing a silent loss of consecutive push counts, cursors, or storm ledgers.
* **Recommendation**: All mutating steps must be staged. Move runtime state files only after the preflight is fully green, and write the TOML manifest changes *last* using temp-file-plus-rename (`os.Rename`). If any step fails, the doctor must rollback the written TOML files and report the precise recovery step.

---

## Required Changes & Action Items

To ensure a seamless, production-ready rollout of the split:
1. **Implement CST-based surgically scoped TOML edits**: Refuse naive decode-encode round-trips; edit import strings preserving comments and unknown fields perfectly.
2. **Use absolute literal paths for legacy directory detection**: Avoid suffix-based string matching for `.gc/system/packs/maintenance`.
3. **Stage runtime state moves before writing TOML manifests**: Ensure that the TOML manifest update is the final, atomic, renaming operation.
4. **Sequence install validation as a prerequisite for import removal**: Never delete the `"maintenance"` import if Core materialization or Gastown fetch is failing.
5. **Add air-gapped/offline diagnostics and mirror overrides**: Ensure the remote resolver is configurable to load from local Git mirrors or directories to support offline workflows.

---

## Verdict: Approved with Risks

The iteration 8 design is highly mature, exceptionally well-structured, and comprehensively mitigates the primary rollout and version-skew risks. Resolving the TOML preservation and failure-atomic sequencing during implementation will guarantee a perfect, robust transition for Gas City operators.
