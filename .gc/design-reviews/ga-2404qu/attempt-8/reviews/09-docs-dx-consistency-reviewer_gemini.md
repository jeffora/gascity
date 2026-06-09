# Felix Moreau — DeepSeek V4 Flash (Docs & DX Consistency Review, Iteration 8, Independent)

**Verdict:** approve-with-risks

**Scope:** Documentation consistency, operator-facing terminology alignment, tutorial and troubleshooting integrity, maintenance-word disambiguation, runtime state continuity, and verification contracts. Reviewed against the Iteration 8 design (`.gc/design-review-inputs/core-gastown-pack-migration/design.md` updated at `2026-06-07T06:45:00Z`), grounded in the live `docs/` tree, `docs/docs.json`, `cmd/gc` strings, `internal/events`, and `examples/`.

---

## Executive Summary

The Iteration 8 design represents an outstanding, mature iteration of the Core/Gastown pack split migration. It resolves previous critical blockers by integrating an **executable wording matrix**, a **preflight doctor safety contract**, and clear sequencing rules that ensure documentation is only finalized after code behavior is pinned. 

The primary residual risk lies in **coverage scope and terminology disambiguation**. While the mechanisms for verification are robust, the hand-named list of documents to update misses key pages where retired terminology actively lives. Furthermore, the introduction of "Core generic maintenance" as a concept risks perpetuating the exact operator confusion we are trying to resolve (i.e., conflating the Core pack with the retired Maintenance pack).

To guarantee a clean transition, I approve this design with risks, provided the following gaps are addressed during the implementation phase.

---

## Top Strengths

1. **Wording Matrix Bound via Executable Contract:** 
   Lines 865–887 bind the doctor `FixHint`, CLI help, config schema generated docs, tutorials, troubleshooting, examples, pack comments, and docs navigation to a single executable wording matrix with strict golden tests. This is a massive victory for DX consistency and prevents terminology drift across API boundaries.
2. **Strict Sequencing of Documentation Edits:**
   Lines 242–243 ensure that docs are finalized *only after* the runtime state, `dog` pool behavior, order naming, and public-source semantics are completely locked. This prevents half-migrated "ghost docs" where the documentation describes behavior that has not yet landed in code.
3. **Rigorous Doctor Fix Safety Contract:**
   Lines 1277–1305 outline a highly defensive atomic mutation contract. It ensures that healthy cities remain byte-identical after `gc doctor --fix`, and that custom local/operator forks or custom Maintenance-like imports are flagged as manual/diagnostic instead of being destructively auto-removed.

---

## Critical Risks & Gaps

### 1. [Major] The Hand-Named Update List Leaves Gaps (Whole-Tree Gate Required)
The "Update docs and generated references" section (L1378–1397) explicitly names only `system-packs.md`, `shareable-packs.md`, `troubleshooting.md`, and Tutorial 01. However, retired-pack terminology actively exists on several other pages in the live tree that are omitted from the named list:
*   **`docs/tutorials/07-orders.md:92-94`** lists the exact migrated/retired orders (e.g., `mol-dog-jsonl`, `mol-dog-reaper`, `prune-branches`) as built-in orders that "ship with the tutorial template." Post-migration, `prune-branches` is Gastown-owned, and the others are Core-owned, making this unified "ship-with-template" list incorrect and confusing.
*   **`docs/getting-started/coming-from-gastown.md:545`** contains a literal permalink to `examples/gastown/packs/gastown/pack.toml` in the main repo. This directory will be deleted by the source-deletion slice, resulting in a dead link in the main operator migration guide.
*   **`docs/guides/migrating-to-pack-vnext.md:134`** uses legacy `includes = ["packs/gastown"]` and L534 uses `source = "./assets/imports/maintenance"`, both of which represent obsolete paths that will fail post-migration.

The design relies on a manual/scripted `rg-inventory` (L1358–1363) to find these hits, but the golden testing is only asserted for named pages. 
*   **Mitigation:** The docs-lint must be a **whole-tree zero-unclassified-hit gate** covering all files under `docs/` and `examples/`. Any unclassified occurrence of `maintenance`, `system/packs`, `runtime/packs`, or `gastown` must fail CI.

### 2. [Minor] Terminology Leak: "Core Generic Maintenance"
The import-state doctor check message is updated to say: *"maintenance is retired; Core supplies generic maintenance and Gastown supplies Gastown-specific behavior"* (L1260–1261). 
Additionally, the design refers to `dog` as the "Core maintenance agent" (L1388–1389). 
This introduces "Core generic maintenance" as an active behavior concept. An operator reading this is highly likely to conclude that the Core pack is simply the new "Maintenance pack" under a different name, reviving the retired mental model.
*   **Mitigation:** The wording matrix must explicitly constrain operator-facing strings to refer to Core operations as **"Core utility infrastructure"**, **"housekeeping operations"**, or **"system-level utility orchestration"** rather than "generic maintenance".

### 3. [Minor] Creation-vs-Update Contradiction on `system-packs.md`
The design is contradictory regarding `docs/reference/system-packs.md`. It lists the file as an existing doc to update (L1380), but also says the source-deletion/docs slice "creates and nav-registers" it (L868). This file does not exist in the live `docs/reference/` tree. 
*   **Mitigation:** Clarify that this is a **creation** step, outline its full canonical structure, and ensure the docs-navigation test requires its registration in `docs/docs.json`.

### 4. [Minor] Worker Pool vs. Agent Ambiguity
The design proposes updating `docs/tutorials/01-cities-and-rigs.md:189` to describe the `dog` pool as "Core's configurable maintenance agent" (L1388–1389). However, `dog` is a *worker pool* (a scaled group with `min/max` settings), not a single agent session.
*   **Mitigation:** Ensure the tutorial prose refers to `dog` as a **"background utility worker pool"** or **"system utility pool"** to match the actual CLI output formats.

---

## Missing Evidence

1.  **Dolt Store Maintenance Safeguard:** There is no explicit assurance in the events section that `gc.store.maintenance.*` events and `[maintenance.dolt]` config schemas are added to a strict allowlist in the docs lint to prevent accidental renames that would break wire-format compatibility.
2.  **Troubleshooting Guidance for Legacy Paths:** While the design specifies replacing `.gc/runtime/packs/maintenance` in `troubleshooting.md`, it does not clarify if the guide should instruct operators on how to manually clean up or inspect their legacy state during the migration, given that the doctor check ignores but preserves these directories on disk.

---

## Required Changes

1.  **Whole-Tree Docs Lint:** Convert the docs-lint into a whole-tree gate. Every match of `maintenance`, `packs/maintenance`, `packs/gastown`, `.gc/system/packs/`, or `\bdog\b` under `docs/` or `examples/` must either be classified under an allowed category or trigger a build failure.
2.  **Explicitly Name Additional Migration Guides:** Include `docs/getting-started/coming-from-gastown.md`, `docs/guides/migrating-to-pack-vnext.md`, and Tutorial 07 in the explicit list of files requiring edits and golden-test verification.
3.  **Refine Import-State Doctor Message:** Avoid using the phrase "Core supplies generic maintenance" in operator-facing CLI strings. Rewrite to:
    > *"the standalone Maintenance pack is retired; Core supplies required system-level utility infrastructure and Gastown supplies Gastown-specific behavior."*
4.  **Resolve System-Packs Verb:** Commit to the **creation** of `docs/reference/system-packs.md` and verify its navigation registration.
5.  **Dolt Event Isolation:** Add `gc.store.maintenance.done` and `gc.store.maintenance.failed` to the docs-lint allowlist.

---

## Questions

*   **Offline Migration Safety:** If an operator runs `gc doctor --fix` offline, and their city relies on public Gastown which is *not* in their local remote-pack cache, will the doctor block the fix gracefully and output a clear error explaining the network dependency?
*   **Compatibility Aliases:** If a user has existing skips targeting `maintenance.gate-sweep`, will Core register compatibility aliases for `core.gate-sweep` so their deployments do not suddenly stall?
