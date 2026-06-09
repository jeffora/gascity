# Yuki Hayashi — DeepSeek V4 Flash (Rollout Version Skew Review, Iteration 7, Independent)

**Verdict:** approve-with-risks

**Lane:** Two-repo rollout sequencing, public pack pin integrity, intermediate state safety, rollback granularity.

Reviewed against the iteration 7 design-before document (updated 2026-06-07T02:23, containing the machine-readable `behavior-manifest.generated.yaml` specification, `Pre-Resolution Doctor and Legacy Import Recovery` phase, and 7-slice rollout staging) and grounded in the `cmd/gc/`, `internal/config/`, and `internal/builtinpacks/` packages.

---

## Executive Summary

The iteration 7 design is exceptionally robust and represents an outstanding engineering blueprint for the Core and Gastown split. By moving from high-level prose to a machine-readable, machine-gated **Executable Source-Discovery Manifest**, the design ensures behavior-level preservation is mathematically verifiable. 

Furthermore, the addition of the **Pre-Resolution Doctor and Legacy Import Recovery** phase resolves a critical architectural gap identified in previous review waves. By performing a light, pre-load TOML parsing pass restricted purely to imports and sources, the binary can boot, diagnose, and auto-fix stale configuration *before* any full config evaluation or runtime loading takes place.

I am granting this design a verdict of **approve-with-risks**. The staged rollout and rollback safety are structurally sound. However, there are a few subtle, major operational risk vectors regarding TOML parser round-trip lossiness, suffix-based directory matching fallacies, and transaction-state sequencing during partial failures that the implementation team must address.

---

## Lane Question Analysis

### 1. What does a fresh `gc init --template gastown` produce between the `gascity-packs` landing and the Gas City `PublicGastownPackVersion` pin update, and is that state deployable?

**It produces a stable, pre-split city locked to the legacy remote pin, which is fully operational and deployable.**

During this intermediate window (Slice 1 complete, but Slice 2 pending):
1. The `gc` binary's `PublicGastownPackVersion` still holds the old pinned Git commit SHA.
2. The binary still includes `"maintenance"` in `requiredBuiltinPackNames` and embeds the Maintenance pack assets.
3. When `gc init` runs, it writes the old SHA to the city's lock file.

Since the old public commit remains immutable and reachable on GitHub, and the binary still auto-materializes/injects the embedded Maintenance pack, this intermediate state loads cleanly, behaves exactly like the pre-migration system, and is fully deployable. This guarantees that operators are never exposed to half-migrated or broken "flag-day" configurations.

### 2. Is `PublicGastownPackVersion` pinned to immutable content with materialization-time verification rather than a mutable branch or tag?

**Yes, via a cryptographically immutable Git commit SHA and standard remote-cache integrity validation.**

The design hardcodes `PublicGastownPackVersion` to a 40-character Git commit SHA. At materialization time, the public Gastown pack is resolved as a remote import (`config.RepoCacheKey`), bypassing the bundled synthetic pack layout. 

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

### [Major] TOML Round-trip Lossiness rendering Scoped Edits unusable on Real Cities

The design-before specifies:
> *TOML edits are scoped and preserving. If comments, unknown tables, array order, or formatting cannot be preserved, `gc doctor --fix` refuses and gives manual steps.*

* **The Risk**: If the pre-resolution doctor uses Go's standard `toml.NewEncoder` to write back modified manifests, the encoder will discard comments, reorder tables, and strip unrecognized custom fields. Since almost all real-world cities contain hand-crafted comments or operator-specific metadata, a naive decode-encode loop will cause the doctor to either corrupt the files or constantly refuse to fix, forcing manual operator repair for virtually 100% of production upgrades.
* **Recommendation**: The implementation of the pre-resolution doctor must use a concrete syntax tree (CST) parser (such as `github.com/pelletier/go-toml/v2`'s AST or targeted regex line-based search and replace) to surgically edit the `imports` array while preserving formatting, comments, and unknown keys exactly.

### [Major] Path Suffix Sincerity Fallacy in Ignored-Directory Logic

The design states:
> *Stale `.gc/system/packs/maintenance` or `.gc/system/packs/gastown` directories are ignored by config loading after the migration.*

* **The Risk**: If the config loader or doctor check uses suffix-based path matching (e.g. `strings.HasSuffix(path, "packs/maintenance")`), it risks falsely matching operator-configured custom folders that happen to share a similar suffix structure (such as `/custom-packs/maintenance`). This could lead to custom packs being silently ignored or misclassified as legacy system folders.
* **Recommendation**: Ensure all directory checks resolve paths to absolute paths and perform literal comparisons against the exact city-scoped system paths: `filepath.Join(cityPath, ".gc/system/packs/maintenance")` and `filepath.Join(cityPath, ".gc/system/packs/gastown")`.

### [Major] Absence of Failure-Atomic Rollback Coordination for Multi-File and State Mutations

The pre-resolution doctor performs both manifest TOML rewrites and runtime-state file migrations (moving JSONL logs, archive repositories, and count files from `/packs/maintenance` to `/packs/core`).
* **The Risk**: If a partial failure occurs mid-execution (for example, the TOML file write succeeds but moving the JSONL archive directory fails due to a disk permission or file lock), the city will be left in a corrupted split-state. The manifest will look fully migrated, but the runtime state will remain stranded in the legacy location, causing a silent loss of consecutive push counts, cursors, or storm ledgers.
* **Recommendation**: All mutating steps must be staged. Move runtime state files only after the preflight is fully green, and write the TOML manifest changes *last* using temp-file-plus-rename (`os.Rename`). If any step fails, the doctor must rollback the written TOML files and report the precise recovery step.

### [Major] Dependency Sequencing Gap in Pre-Resolution Doctor

The design states:
> *Generated or known example Maintenance imports are removed only after Core is materialized and public Gastown lock/install validation succeeds.*

* **The Risk**: If `importStateDoctorCheck` removes the Maintenance import line before ensuring that host Core has successfully materialized or that the public Gastown pack can be fetched (especially on air-gapped systems), the city will immediately fail loading and become unbootable.
* **Recommendation**: Explicitly sequence the doctor's fixes. The doctor must prove Core materialization and public Gastown install preflight *succeed* before deleting the `"maintenance"` import or rewriting any `"gastown"` local path in `pack.toml`/`city.toml`.

---

## Required Changes & Action Items

To ensure a seamless, production-ready rollout of the split:
1. **Implement CST-based surgically scoped TOML edits**: Refuse naive decode-encode round-trips; edit import strings preserving comments and unknown fields perfectly.
2. **Use absolute literal paths for legacy directory detection**: Avoid suffix-based string matching for `.gc/system/packs/maintenance`.
3. **Stage runtime state moves before writing TOML manifests**: Ensure that the TOML manifest update is the final, atomic, renaming operation.
4. **Sequence install validation as a prerequisite for import removal**: Never delete the `"maintenance"` import if Core materialization or Gastown fetch is failing.

---

## Verdict: Approved with Risks

The iteration 7 design is highly mature, exceptionally well-structured, and comprehensively mitigates the primary rollout and version-skew risks. Resolving the TOML preservation and failure-atomic sequencing during implementation will guarantee a perfect, robust transition for Gas City operators.
