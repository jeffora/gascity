# Petra Novak - DeepSeek V4 Flash

**Verdict:** approve-with-risks

**Lane:** Builtinpacks registry, embed-path migration, materialization safety, Maintenance retirement, and downstream-reference closure.

I have evaluated the updated `plans/core-gastown-pack-migration/requirements.md` (now at `status: questions`, updated to 2026-06-09T01:20:00Z) against the `gc.mayor.requirements.v1` schema. The restructure is highly conformant, successfully segregating implementation detail from product requirements, adding the necessary W6H, Example Mapping, and a comprehensive set of Acceptance Criteria (AC1 to AC14).

By transitioning to `status: questions`, the document correctly acknowledges that key planning artifacts (the AC6 asset migration ledger and AC7 behavior-preservation manifest) are required before implementation approval. This review focuses on the remaining build, embed, and materialization hazards that must be explicitly resolved in design/implementation or pinned within the product contracts.

---

## Live Code Base Inventory (Current Bad-State)
To ensure absolute alignment with the physical codebase, I re-verified the following build-seam locations:
1. **Required Pack Hardcoding:** `cmd/gc/embed_builtin_packs.go` at line 237 declares `required := []string{"core", "maintenance"}`, which forces Maintenance to be auto-included in every city.
2. **Builtin Pack Registry & Imports:** `internal/builtinpacks/registry.go` compile-imports `examples/gastown/packs/maintenance` at line 19, lists it in the `All()` slice at line 56, and aliases `"gastown", "maintenance"` at line 128 under `publicSubpathForPack`. Core's embed source is pinned to `internal/bootstrap/packs/core` at line 53.
3. **Legacy Doctor Messaging:** `cmd/gc/import_state_doctor_check.go` asserts `"should be removed; maintenance/core is supplied implicitly"` at line 194, with the legacy recognizer checking `examples/gastown/packs/` or `.gc/system/packs/` at lines 227–240.

---

## Consensus Strengths
- **Rigorous Schema Adherence (AC1):** The requirements now strictly adhere to `gc.mayor.requirements.v1` section order and meta-structures. It removes file-by-file assignments from the requirements document, avoiding premature implementation commitments.
- **Standalone Maintenance Retirement (AC5):** Explicitly states that Maintenance is no longer bundled, auto-included, materialized as active system pack, or chosen via lockfile resolution.
- **Validated Asset Ledger (AC6):** The inclusion of a mandatory, pre-implementation migration ledger with current path, target, action, split boundary, and rationale ensures traceability and prevents phantom or orphaned rows.
- **Configurable Maintenance Executor (AC9):** Decouples the Core maintenance execution from Go-side role assumptions. While `dog` can be defined as default pack configuration, the SDK is self-sufficient without it, satisfying Zero Framework Cognition (ZFC) principles.

---

## Critical Risks and Gaps (DeepSeek V4 Flash Focus)

While the updated requirements are significantly stronger, they contain critical build and materialization blind spots that other reviewers might accept too quickly:

### 1. [Major] Directory-Level Materialization Safety & Concurrency
**Hazard:** If this migration changes the materialization layouts under `.gc/system/packs/`, the extraction logic must be bulletproof. A naive implementation using direct `mkdir` and progressive copying is highly vulnerable to concurrent reader corruption (where another `gc` process reads a partially written directory) or failures mid-materialization.
**Omission:** The updated requirements are completely silent on directory promotion safety. 
**Required Pin:** The implementation must require **directory-level atomic materialization**:
- Extract pack assets into a unique temporary sibling directory (e.g., `.gc/system/packs/.tmp-core-xyz`).
- On successful extraction, perform an atomic directory rename (using `os.Rename` semantics) into the target path.
- Clean up any stale/partial temporary staging directories on failure or startup.

### 2. [Major] Prohibition of Symlinks, Hardlinks, and Pointer Files
**Hazard:** Materialization and embedding are highly host-dependent if the underlying file system utilizes symlinks or hardlinks. If a materialized pack references symlinks, behaviors can break across different sandboxes, container runtimes, or OS configurations (especially on Windows hosts or restricted k8s sandboxes).
**Omission:** No safety invariant exists to govern file types in the materialized bundle.
**Required Pin:** Embedded and materialized pack layouts **must not contain symlinks, hardlinks, or pointer-file stand-ins**. Add a static check/validation rule enforcing that all materialized assets are regular files or directories.

### 3. [Minor] Go Bundling Seam Explicitness
**Hazard:** While AC5 and AC6 implicitly cover "retired-Maintenance... consumers," they leave the actual Go build seams to inference. A leftover compile-time import (e.g., `registry.go:19`) will break compilation when directories are removed. 
**Omission:** The Go-level files (`cmd/gc/embed_builtin_packs.go` and `internal/builtinpacks/registry.go`) are not explicitly enumerated as target closure areas.
**Required Pin:** The design document or `source-consumer-closure.yaml` must explicitly list the `All()` registry entry, required pack literals, and per-pack `embed.go` build directives as compile-time closure targets.

### 4. [Minor] Doctor Message Flip vs. Legacy Recognizer Retention
**Hazard:** There is a subtle correctness point regarding existing-city upgrades. The message `"supplied implicitly"` (at `import_state_doctor_check.go:194`) is now incorrect and must change. However, if the legacy recognizer (`examples/gastown/packs/maintenance` or `.gc/system/packs/maintenance`) is completely deleted, existing cities with these legacy paths in their imports will no longer be classified as legacy/retired, rendering diagnostics and `--fix` repair loops silent and ineffective.
**Omission:** AC10 and AC11 require upgrade diagnostics but fail to specify this structural preservation.
**Required Pin:** The legacy Maintenance/Gastown recognizers **must be retained** to classify legacy city states, while the associated warning message must be updated to reflect the retired/removed status of the Maintenance pack.

### 5. [Minor] Settling the Core Relocation Policy
**Hazard:** The Problem Statement lists `internal/bootstrap/packs/core` as a legacy source-tree location. However, the requirements do not clarify whether Core's embed source actually relocates (e.g., to `internal/packs/core`) or stays put while using a compatibility shim.
**Omission:** This leaves the end-state location of the Core pack's compile-time source ambiguous.
**Required Pin:** Specify whether Core's embed source moves to a new canonical directory inside the repository or remains in bootstrap, and define how stale references to the old location are handled.

---

## Missing Evidence
1. **Materialization Invariants:** Documentation/design asserting the absence of symlinks/hardlinks in embedded or materialized structures.
2. **Atomic Directory Transition Strategy:** Code patterns or specifications illustrating staging-and-rename mechanics during pack extraction.
3. **Legacy Diagnostic Matrix:** Clear before/after expected outputs for `gc doctor` and `gc import-state` when scanning a city containing legacy `packs/maintenance` imports.
4. **Behavior Rehoming Mapping:** Explicit destination mappings for former Maintenance behaviors (`dolt-target.sh`, `_bd_trace.sh`, dog molecules, `jsonl-export.sh`, and `reaper.sh`) to guarantee that non-Gastown cities do not experience silent degradation.

---

## Required Changes
1. **Atomic Materialization Rule:** Require that all pack materialization uses unique sibling staging, atomic directory renaming, and failure cleanup to prevent concurrent corruption.
2. **File Safety Invariant:** Formally prohibit symlinks, hardlinks, and pointer-files in the embedded and materialized pack layouts, validated via static tests.
3. **Registry and Embed Go Target Listing:** Explicitly target `internal/builtinpacks/registry.go` (import, `All()`, `publicSubpathForPack` alias) and `cmd/gc/embed_builtin_packs.go` (`requiredBuiltinPackNames` literal) as compile-time closure targets.
4. **Recognizer Retention Rule:** Explicitly require that legacy recognizers remain active and tested, while the diagnostic message flips to retired/removed rather than "supplied implicitly."
5. **Core Source Location Resolution:** Ground the Core relocation policy by specifying whether Core's compile-time source directory shifts or remains, resolving the ambiguity of the legacy label.

---

## Verdict Calibration / Rationale
I choose **approve-with-risks**. The schema compliance, transition to `status: questions`, and the conceptual decoupling of required Core from external Gastown are sound and ready to progress. However, the build, embedding, and materialization layers are where silent runtime corruption and build failures occur. These risks must be handled in the design phase before code execution begins.
